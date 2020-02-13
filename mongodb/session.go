package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/mongodb/internal/mongoutil"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/description"
	"go.mongodb.org/mongo-driver/x/mongo/driver/operation"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"
)

// Session holds information used to create a connection
// to MongoDB. See mongodb/README.md for more details.
type Session struct {
	ClientAddresses []string
	Deployment      driver.Deployment
	TopologyKind    description.TopologyKind
	NumConns        int
	Pool            *SessionConnPool
	ReadPreference  *readpref.ReadPref
	errLock         sync.Mutex

	err        error
	AuthSource string
}

// Close closes the direct server connection
// associated with this session.
func (s *Session) Close() error {
	if s.Pool != nil {
		s.Pool.Close()
	}
	return nil
}

// Connection gets a connection to use.
func (s *Session) Connection(ctx context.Context) (driver.Connection, error) {
	c, err := s.Pool.Get(ctx)
	if err == topology.ErrPoolDisconnected {
		s.setError(err)
	}
	return c, err
}

// GetClientAddresses returns all addresses in the session's connection pool.
func (s *Session) GetClientAddresses() []string {
	return s.ClientAddresses
}

// ConnLen is the number of connections that are part of a session.
func (s *Session) ConnLen() int {
	return s.NumConns
}

// Err returns a session level error that may have occurred.
func (s *Session) Err() error {
	if s.err != nil {
		return s.err
	}

	return s.Pool.Err()
}

// setError synchronizes error assignment to the session.
func (s *Session) setError(err error) {
	s.errLock.Lock()
	s.err = err
	s.errLock.Unlock()
}

// Validate checks that the established Session meets the read preference
// requirements. When a SessionProvider creates a Session, it selects a
// driver.Server using a specified ReadPref. That Server is used to check
// out driver.Connections for the Session's.Pool. That ReadPref is stored
// in the Session itself as the field rp. During the Session's lifetime,
// the Server may change its role (for example, it may go from Secondary
// to Primary). This method validates the selected Server still meets the
// read preference requirements from when the Session was created.
func (s *Session) Validate(ctx context.Context) error {
	err := s.Err()
	if err != nil {
		return err
	}

	// Get a connection to get the server description.
	c, err := s.Connection(ctx)
	if err != nil {
		return err
	}

	// Close the connection once we get the description.
	server := c.Description()
	_ = c.Close()

	// Create a ReadPrefSelector using the read preference
	// from this session's creation.
	selector := description.ReadPrefSelector(s.ReadPreference)
	t := description.Topology{
		Kind: s.TopologyKind,
	}

	// Attempt to select the server. If the selector returns an error
	// or does not select the one server, then this session does not
	// meet the read preference requirements.
	servers, err := selector.SelectServer(t, []description.Server{server})
	if err != nil || len(servers) == 0 {
		s.setError(fmt.Errorf("current session does not satisfy read preference"))
		return s.err
	}

	return nil
}

//
// Session helper methods
//

// Aggregate runs the aggregation pipeline passed in against the
// give database and collection.
func (s *Session) Aggregate(ctx context.Context, database, collection string, pipeline []bson.D) (Cursor, error) {
	pipelineArr, err := bsonutil.DocSliceToCoreArray(pipeline)
	if err != nil {
		return nil, fmt.Errorf("error converting aggregation pipeline to raw array: %v", err)
	}

	c := operation.NewAggregate(pipelineArr).
		Database(database).
		Collection(collection).
		Deployment(s.Deployment).
		AllowDiskUse(true).
		ReadPreference(s.ReadPreference)

	// When the query context has a deadline then we supply the amount of
	// time we have left as the max time to try to execute the query in.
	maxTimeMS := int64(0)
	if ctxDeadlineTime, ctxHasDeadLine := ctx.Deadline(); ctxHasDeadLine {
		maxTimeMS = int64(time.Until(ctxDeadlineTime))
		c = c.MaxTimeMS(maxTimeMS)
	}

	err = c.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("error running aggregation: %v", err)
	}

	cursor, err := c.Result(driver.CursorOptions{
		MaxTimeMS: maxTimeMS,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting aggregation result: %v", err)
	}

	return newBatchCursor(cursor), nil
}

// Count runs a count command for a specific database and collection.
func (s *Session) Count(ctx context.Context, database, collection string) (int, error) {
	cmd := bsonutil.NewD(
		bsonutil.NewDocElem("count", collection),
	)

	result := struct {
		N  int `bson:"n"`
		Ok int `bson:"ok"`
	}{}

	err := s.Run(ctx, database, cmd, &result)
	if err != nil {
		return 0, fmt.Errorf("error getting count for '%v'.'%v': %v", database, collection, err)
	}

	if result.Ok != 1 {
		return 0, fmt.Errorf("error getting count for '%v'.'%v'", database, collection)
	}

	return result.N, nil
}

// Delete deletes all documents from the specified namespace matching the
// provided query.
func (s *Session) Delete(ctx context.Context, db, col string, query bson.D) error {
	deleteDoc := bsonutil.NewD(
		bsonutil.NewDocElem("q", query),
		bsonutil.NewDocElem("limit", 0),
	)

	deleteDocBytes, err := bson.Marshal(deleteDoc)
	if err != nil {
		return fmt.Errorf("error marshaling delete doc: %v", err)
	}

	wc := writeconcern.New(writeconcern.WMajority())
	c := operation.NewDelete(deleteDocBytes).
		Database(db).
		Collection(col).
		Deployment(s.Deployment).
		WriteConcern(wc)

	err = c.Execute(ctx)
	if err != nil {
		return fmt.Errorf("error deleting document: %v", err)
	}

	return nil
}

// Insert inserts the provided documents into the specified namespace.
func (s *Session) Insert(ctx context.Context, db, col string, docs []interface{}) error {
	docsToInsert := make([]bsoncore.Document, len(docs))

	var err error
	for i, doc := range docs {
		docsToInsert[i], err = bson.Marshal(&doc)
		if err != nil {
			return fmt.Errorf("error marshaling insert doc: %v", err)
		}
	}

	wc := writeconcern.New(writeconcern.WMajority())
	c := operation.NewInsert(docsToInsert...).
		Database(db).
		Collection(col).
		Deployment(s.Deployment).
		WriteConcern(wc)

	err = c.Execute(ctx)
	if err != nil {
		return fmt.Errorf("error inserting document(s): %v", err)
	}

	return nil
}

// Upsert replaces a single document matching the provided query
// in the specified namespace with the provided update document.
func (s *Session) Upsert(ctx context.Context, db, col string, query, update interface{}) error {
	updateDoc := bsonutil.NewD(
		bsonutil.NewDocElem("q", query),
		bsonutil.NewDocElem("u", update),
		bsonutil.NewDocElem("upsert", true),
	)

	updateDocBytes, err := bson.Marshal(&updateDoc)
	if err != nil {
		return fmt.Errorf("error marshaling update doc: %v", err)
	}

	wc := writeconcern.New(writeconcern.WMajority())
	c := operation.NewUpdate(updateDocBytes).
		Database(db).
		Collection(col).
		Deployment(s.Deployment).
		WriteConcern(wc)

	err = c.Execute(ctx)
	if err != nil {
		return fmt.Errorf("error upserting document: %v", err)
	}

	return nil
}

// ListCollections returns a cursor to iterate through the collections
// present on the db database with options opts.
func (s *Session) ListCollections(ctx context.Context, db string, opts driver.CursorOptions) (Cursor, error) {
	cmd := operation.NewListCollections(nil).
		Database(db).
		Deployment(s.Deployment).
		ReadPreference(s.ReadPreference)
	if err := cmd.Execute(ctx); err != nil {
		return nil, fmt.Errorf("error listing collections for '%v': %v", db, err)
	}

	cursor, err := cmd.Result(opts)
	if err != nil {
		return nil, fmt.Errorf("error getting listCollections result: %v", err)
	}

	return newListCollectionsCursor(cursor), nil
}

// ListDatabases returns a cursor to iterate through
// the database names present on a server.
func (s *Session) ListDatabases(ctx context.Context) (*operation.ListDatabasesResult, error) {
	cmd := operation.NewListDatabases(nil).
		Database("admin").
		Deployment(s.Deployment).
		ReadPreference(s.ReadPreference)
	if err := cmd.Execute(ctx); err != nil {
		return nil, fmt.Errorf("error listing databases: %v", err)
	}

	res := cmd.Result()
	return &res, nil
}

// ListIndexes returns a cursor to iterate through the
// indexes on the c collection within the db database.
func (s *Session) ListIndexes(ctx context.Context, db, col string) (Cursor, error) {
	cmd := operation.NewListIndexes().
		Database(db).
		Collection(col).
		Deployment(s.Deployment)
	if err := cmd.Execute(ctx); err != nil {
		return nil, fmt.Errorf("error listing indexes for '%v'.'%v': %v", db, col, err)
	}

	cursor, err := cmd.Result(driver.CursorOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting listIndexes result: %v", err)
	}

	return newBatchCursor(cursor), nil
}

// currentOp represents the result of a currentOp command. The 'Opid' can be an integer on a mongod
// and a string on a mongos.
//
// To kill an operation, we typically only need the 'opid': { killOp: <opid> }.
// However, on server versions >= 4.0, if the operation is a 'getmore', we need
// to use 'killCursors' instead of 'killOp'. For that, we need:
//   - the database name ('command.$db')
//   - the collection name ('command.collection')
//   - the cursor ID ('command.getMore')
//
// The 'command' field may not be a document, so we use interface{} as the type in the struct.
type currentOp struct {
	Opid    interface{} `bson:"opid"`
	Op      string      `bson:"op"`
	Command interface{} `bson:"command"`
}

type killCursorsArgs struct {
	cursorID   int64
	db         string
	collection string
}

// isGetMore returns true if currentOp.Op is "getmore" and the expected
// 'command' field is present. It also returns the cursor ID, database,
// and collection if this operation is a getMore.
func (co currentOp) isGetMore() (killCursorsArgs, bool) {
	if co.Op != "getmore" {
		return killCursorsArgs{0, "", ""}, false
	}

	if co.Command != nil {
		if doc, ok := co.Command.(bson.D); ok {
			// if 'command' is present and is a document, we can search
			// for the required fields.
			var getMore int64
			var collection, db string
			for _, e := range doc {
				switch e.Key {
				case "getMore":
					getMore = e.Value.(int64)
				case "collection":
					collection = e.Value.(string)
				case "$db":
					db = e.Value.(string)
				}
			}

			if getMore != 0 && collection != "" && db != "" {
				return killCursorsArgs{getMore, db, collection}, true
			}
		}
	}

	panic("invalid currentOp response: getMore operation missing 'command' field")
}

// KillOps kills all operations running on a list of client addresses.
func (s *Session) KillOps(ctx context.Context, clientAddresses []string) error {
	if len(clientAddresses) == 0 {
		return nil
	}

	currentOpsToKill, err := s.listCurrentOpsForClients(ctx, clientAddresses)
	if err != nil {
		return err
	}

	version, err := s.Version()
	if err != nil {
		return err
	}

	useKillCursors := procutil.VersionAtLeast(version.VersionArray, []uint8{'4', '0', '0'})

	for _, op := range currentOpsToKill {
		if useKillCursors {
			if args, isGetMore := op.isGetMore(); isGetMore {
				err = s.killCursors(ctx, args.cursorID, args.db, args.collection)
				if err != nil {
					return err
				}
				continue
			}
		}
		err = s.killOp(ctx, op.Opid)
		if err != nil {
			return err
		}
	}
	return nil
}

// listCurrentOpsForClients returns all operations belonging to the provided
// session's user from a list of client addresses.
func (s *Session) listCurrentOpsForClients(ctx context.Context, clientAddresses []string) ([]currentOp, error) {
	clientAddressArray := make(bson.A, len(clientAddresses))
	for i, clientAddress := range clientAddresses {
		clientAddressArray[i] = clientAddress
	}

	currentOpsCommand := bsonutil.NewD(
		bsonutil.NewDocElem("currentOp", int32(1)),
		bsonutil.NewDocElem("$and", bsonutil.NewArray(
			bsonutil.NewD(
				// The two conditions in the $or condition handle whether we are talking to a mongod or a
				// mongos. A mongos reports its client addresses in a different format than a mongod.
				bsonutil.NewDocElem("$or", bsonutil.NewArray(
					bsonutil.NewD(
						bsonutil.NewDocElem("client", bsonutil.NewD(bsonutil.NewDocElem("$in", clientAddressArray))),
					),
					bsonutil.NewD(
						bsonutil.NewDocElem("command.$client.mongos.client", bsonutil.NewD(bsonutil.NewDocElem("$in", clientAddressArray))),
					),
				)),
			),
			// ignore the 'currentOp' command itself.
			bsonutil.NewD(
				bsonutil.NewDocElem("command.currentOp", bsonutil.NewD(
					bsonutil.NewDocElem("$ne", 1),
				)),
			),
		)),
	)

	// If auth source is empty, this indicates we're running in unauthenticated mode. We should
	// not use the $ownOps parameter in this case since operations don't have any associated
	// MongoDB users.
	if s.AuthSource != "" {
		currentOpsCommand = append(currentOpsCommand, bsonutil.NewDocElem("$ownOps", true))
	}

	currentOpResponse := struct {
		InProg []currentOp `bson:"inprog"`
	}{}

	err := s.Run(ctx, "admin", currentOpsCommand, &currentOpResponse)
	if err != nil {
		return nil, fmt.Errorf("error running currentOp: %v", err)
	}

	return currentOpResponse.InProg, nil
}

// killCursors kills a cursor on the server for the input collection
// with the input cursor ID.
func (s *Session) killCursors(ctx context.Context, cursorID int64, db, col string) error {
	killCursorsCommand := bsonutil.NewD(
		bsonutil.NewDocElem("killCursors", col),
		bsonutil.NewDocElem("cursors", bsonutil.NewArray(cursorID)),
	)

	killCursorsReponse := struct {
		CursorsKilled []int64
	}{}

	err := s.Run(ctx, db, killCursorsCommand, &killCursorsReponse)
	if err != nil {
		return err
	}

	if len(killCursorsReponse.CursorsKilled) == 0 || killCursorsReponse.CursorsKilled[0] != cursorID {
		return fmt.Errorf("failed to kill cursor for '%v'.'%v', please try again", db, col)
	}

	return nil
}

// killOp kills an operation on the server with the input opID.
func (s *Session) killOp(ctx context.Context, opID interface{}) error {
	killOpCommand := bsonutil.NewD(
		bsonutil.NewDocElem("killOp", 1),
		bsonutil.NewDocElem("op", opID),
	)

	return s.Run(ctx, "admin", killOpCommand, &struct{}{})
}

// Login authenticates the session using the specified authenticator.
func (s *Session) Login(ctx context.Context, a SessionAuthenticator) error {
	s.AuthSource = a.source()

	// checkout all the connections
	conns := make([]driver.Connection, s.NumConns)
	for i := 0; i < s.NumConns; i++ {
		c, err := s.Connection(ctx)
		if err != nil {
			s.setError(err)
			return s.err
		}
		defer func() {
			_ = c.Close()
		}()

		conns[i] = c
	}

	err := a.Auth(ctx, conns)
	s.setError(err)
	return s.err
}

// VersionInfo contains server version info.
type VersionInfo struct {
	Version      string  `bson:"version"`
	GitVersion   string  `bson:"gitVersion"`
	VersionArray []uint8 `bson:"-"`
}

// Version returns the server version for this session, as well as the git version.
func (s *Session) Version() (*VersionInfo, error) {
	info := VersionInfo{}

	cmd := bsonutil.NewD(
		bsonutil.NewDocElem("buildInfo", 1),
	)

	err := s.Run(context.Background(), "admin", cmd, &info)
	if err != nil {
		return nil, err
	}

	info.VersionArray, err = procutil.VersionToSlice(info.Version)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// DropCollection drops a database with majority write concern.
func (s *Session) DropCollection(ctx context.Context, db, col string) error {
	cmd := operation.NewDropCollection().
		Database(db).
		Collection(col).
		Deployment(s.Deployment).
		WriteConcern(writeconcern.New(writeconcern.WMajority()))

	return cmd.Execute(ctx)
}

// DropDatabase drops a database with majority write concern.
func (s *Session) DropDatabase(ctx context.Context, db string) error {
	cmd := operation.NewDropDatabase().
		Database(db).
		Deployment(s.Deployment).
		WriteConcern(writeconcern.New(writeconcern.WMajority()))

	return cmd.Execute(ctx)
}

// Run executes an arbitrary command against the given database.
func (s *Session) Run(ctx context.Context, db string, cmd bson.D, result interface{}) error {
	return mongoutil.ExecuteWithDeployment(ctx, db, s.Deployment, s.ReadPreference, cmd, result)
}
