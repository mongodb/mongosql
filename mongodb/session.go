package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/mongo-go-driver/mongo"
	"github.com/10gen/mongo-go-driver/mongo/model"
	"github.com/10gen/mongo-go-driver/mongo/options"
	"github.com/10gen/mongo-go-driver/mongo/private/cluster"
	"github.com/10gen/mongo-go-driver/mongo/private/conn"
	"github.com/10gen/mongo-go-driver/mongo/private/ops"
	"github.com/10gen/mongo-go-driver/mongo/readconcern"
	"github.com/10gen/mongo-go-driver/mongo/readpref"
	"github.com/10gen/mongo-go-driver/mongo/writeconcern"
	"github.com/10gen/sqlproxy/internal/bsonutil"
)

// currentOp represents the result of a currentOp command. The 'Opid' can be an integer on a mongod
// and a string on a mongos. The 'Client' field is only present when talking to a mongod.
type currentOp struct {
	Client string      `bson:"client"`
	Opid   interface{} `bson:"opid"`
}

// Cursor wraps the ops.Cursor interface for mongosqld
// and mongodrdl clients.
type Cursor interface {
	ops.Cursor
}

// Session holds information used to create a connection
// to MongoDB.
type Session struct {
	clientAddresses []string
	cluster         *cluster.Cluster
	count           int
	server          cluster.Server
	pool            *sessionConnPool

	err            error
	authSource     string
	selectedServer *ops.SelectedServer
}

// AuthSource returns the session's authentication source.
func (s *Session) AuthSource() string {
	return s.authSource
}

// Close closes the direct server connection
// associated with this sesssion.
func (s *Session) Close() error {
	s.pool.Close()
	return nil
}

// ClusterKind returns the kind of cluster
// this session is attached to.
func (s *Session) ClusterKind() model.ClusterKind {
	return s.selectedServer.ClusterKind
}

// Connection gets a connection to use.
func (s *Session) Connection(ctx context.Context) (conn.Connection, error) {
	c, err := s.pool.Get(ctx)
	if err == conn.ErrPoolClosed {
		s.err = err
	}
	return c, err
}

// GetClientAddresses returns all addresses in the session's connection pool.
func (s *Session) GetClientAddresses() []string {
	return s.clientAddresses
}

// ConnLen is the number of connections that are part of a session.
func (s *Session) ConnLen() int {
	return s.count
}

// Err returns a session level error that may have occurred.
func (s *Session) Err() error {
	if s.err != nil {
		return s.err
	}

	return s.pool.Err()
}

// Model returns the description of the server
// asssociated with this session.
func (s *Session) Model() *model.Server {
	return s.server.Model()
}

// Validate checks that the established session meets the server
// version requirements for the BI Connector.
func (s *Session) Validate() error {

	err := s.Err()
	if err != nil {
		return err
	}

	mdl := s.Model()

	if mdl.LastError != nil {
		s.err = mdl.LastError
		return mdl.LastError
	}

	selector := readpref.Selector(s.selectedServer.ReadPref)
	result, err := selector(s.cluster.Model(), []*model.Server{mdl})
	if err != nil || len(result) == 0 {
		s.err = fmt.Errorf("current session does not satisfy read preference")
		return s.err
	}

	return s.Err()
}

//
// Session helper methods
//

// Aggregate runs the aggregation pipeline passed in against the
// give database and collection.
func (s *Session) Aggregate(ctx context.Context, database, collection string, pipeline interface{}) (ops.Cursor, error) {
	ns := ops.NewNamespace(database, collection)

	opts := []options.AggregateOption{mongo.AllowDiskUse(true)}
	if ctxDeadlineTime, ctxHasDeadLine := ctx.Deadline(); ctxHasDeadLine {
		// When the query context has a deadline then we supply the amount of
		// time we have left as the max time to try to execute the query in.
		opts = append(opts, mongo.MaxTime(time.Until(ctxDeadlineTime)))
	}
	return ops.Aggregate(ctx, s.selectedServer, ns, readconcern.Local(), pipeline, opts...)
}

// Count runs a count command for a specific database and collection.
func (s *Session) Count(ctx context.Context, database, collection string) (int, error) {
	return ops.Count(
		ctx,
		s.selectedServer,
		ops.NewNamespace(database, collection),
		nil,
		nil,
	)
}

// Delete deletes all documents from the specified namespace matching the
// provided query.
func (s *Session) Delete(ctx context.Context, db, col string, query interface{}) error {
	ns := ops.NewNamespace(db, col)

	result := struct {
		N  int `bson:"n"`
		Ok int `bson:"ok"`
	}{}

	deletes := []bson.D{
		bsonutil.NewD(
			bsonutil.NewDocElem("q", query),
			bsonutil.NewDocElem("limit", 0),
		),
	}

	wc := writeconcern.New(writeconcern.WMajority(-1))
	err := ops.Delete(ctx, s.selectedServer, ns, wc, deletes, &result)
	if err != nil {
		return err
	}

	if result.Ok != 1 {
		return fmt.Errorf("error deleting document")
	}

	return nil
}

// Insert inserts the provided documents into the specified namespace.
func (s *Session) Insert(ctx context.Context, db, col string, docs []interface{}) error {
	ns := ops.NewNamespace(db, col)

	result := struct {
		N  int `bson:"n"`
		Ok int `bson:"ok"`
	}{}

	wc := writeconcern.New(writeconcern.WMajority(-1))
	err := ops.Insert(ctx, s.selectedServer, ns, wc, docs, &result)
	if err != nil {
		return err
	}

	if result.Ok != 1 {
		return fmt.Errorf("error inserting document")
	}

	return nil
}

// Upsert replaces a single document matching the provided query in the
// specified namespace with the provided update document.
func (s *Session) Upsert(ctx context.Context, db, col string, query, update interface{}) error {
	ns := ops.NewNamespace(db, col)

	result := struct {
		N  int `bson:"n"`
		Ok int `bson:"ok"`
	}{}

	updates := []bson.D{
		bsonutil.NewD(
			bsonutil.NewDocElem("q", query),
			bsonutil.NewDocElem("u", update),
			bsonutil.NewDocElem("upsert", true),
		),
	}

	wc := writeconcern.New(writeconcern.WMajority(-1))
	err := ops.Update(ctx, s.selectedServer, ns, wc, updates, &result)
	if err != nil {
		return err
	}

	if result.Ok != 1 {
		return fmt.Errorf("error upserting document")
	}

	return nil
}

type cursorRequest struct {
	BatchSize int32 `bson:"batchSize,omitempty"`
}

// The result of a command that returns a cursor.
type cursorReturningResult struct {
	Cursor firstBatchCursorResult `bson:"cursor"`
}

// The first batch of a cursor.
type firstBatchCursorResult struct {
	// The first batch of the cursor.
	FirstBatch []bson.Raw `bson:"firstBatch"`
	// The namespace used to iterate the cursor.
	NS string `bson:"ns"`
	// The cursor's id.
	ID int64 `bson:"id"`
}

func (cursorResult *firstBatchCursorResult) Namespace() ops.Namespace {
	// Assume server returns a valid namespace string.
	namespace := ops.ParseNamespace(cursorResult.NS)
	return namespace
}

func (cursorResult *firstBatchCursorResult) InitialBatch() []bson.Raw {
	return cursorResult.FirstBatch
}

func (cursorResult *firstBatchCursorResult) CursorID() int64 {
	return cursorResult.ID
}

// ListCollections returns a cursor to iterate through
// the collections present on the db database with options opts.
func (s *Session) ListCollections(ctx context.Context, db string, opts ops.ListCollectionsOptions) (ops.Cursor, error) {
	listCollectionsCommand := struct {
		ListCollections int32          `bson:"listCollections"`
		Filter          interface{}    `bson:"filter,omitempty"`
		MaxTimeMS       int64          `bson:"maxTimeMS,omitempty"`
		Cursor          *cursorRequest `bson:"cursor"`
	}{
		ListCollections: 1,
		Filter:          opts.Filter,
		MaxTimeMS:       int64(opts.MaxTime),
		Cursor: &cursorRequest{
			BatchSize: opts.BatchSize,
		},
	}

	var result cursorReturningResult
	err := s.Run(ctx, db, listCollectionsCommand, &result)
	if err != nil {
		return nil, err
	}

	return ops.NewCursor(&result.Cursor, opts.BatchSize, s)
}

type listDatabasesCursor struct {
	databases []bson.Raw
	current   int
	err       error
}

func (cursor *listDatabasesCursor) Next(_ context.Context, result interface{}) bool {
	if cursor.current < len(cursor.databases) {
		err := bson.Unmarshal(cursor.databases[cursor.current].Data, result)
		if err != nil {
			cursor.err = fmt.Errorf("unable to parse listDatabases result: %v", err)
			return false
		}

		cursor.current++
		return true
	}
	return false
}

// Err returns the error status of the cursor.
func (cursor *listDatabasesCursor) Err() error {
	return cursor.err
}

// Close closes the cursor. Ordinarily this is a no-op as the server
// closes the cursor when it is exhausted. Returns the error status
// of this cursor so that clients do not have to call Err() separately.
func (cursor *listDatabasesCursor) Close(_ context.Context) error {
	return nil
}

// ListDatabases returns a cursor to iterate through
// the database names present on a server.
func (s *Session) ListDatabases(ctx context.Context) (ops.Cursor, error) {
	listDataBasesOptions := ops.ListDatabasesOptions{}

	var result struct {
		Databases []bson.Raw `bson:"databases"`
	}
	listDatabasesCommand := struct {
		ListDatabases int32 `bson:"listDatabases"`
		MaxTimeMS     int64 `bson:"maxTimeMS,omitempty"`
	}{
		ListDatabases: 1,
		MaxTimeMS:     int64(listDataBasesOptions.MaxTime / time.Millisecond),
	}

	err := s.Run(ctx, "admin", listDatabasesCommand, &result)
	if err != nil {
		return nil, err
	}

	return &listDatabasesCursor{
		databases: result.Databases,
		current:   0,
	}, nil
}

// ListIndexes returns a cursor to iterate through
// the indexes on the c collection within the db database.
func (s *Session) ListIndexes(ctx context.Context, db, c string) (ops.Cursor, error) {
	opts := ops.ListIndexesOptions{}
	ns := ops.NewNamespace(db, c)

	listIndexesCommand := struct {
		ListIndexes string `bson:"listIndexes"`
	}{
		ListIndexes: ns.Collection,
	}

	var result cursorReturningResult
	err := s.Run(ctx, db, listIndexesCommand, &result)
	switch err {
	case nil:
		return ops.NewCursor(&result.Cursor, opts.BatchSize, s)
	default:
		if conn.IsNsNotFound(err) {
			return ops.NewExhaustedCursor()
		}
		return nil, err
	}
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

	for _, op := range currentOpsToKill {
		err := s.killOp(ctx, op.Opid)
		if err != nil {
			return err
		}
	}
	return nil
}

// listCurrentOpsForClients returns all operations belonging to the provided
// session's user from a list of client addresses.
func (s *Session) listCurrentOpsForClients(ctx context.Context, clientAddresses []string) ([]currentOp, error) {
	// The two conditions in the $or condition handle whether we are talking to a mongod or a
	// mongos. A mongos reports its client addreses in a different format than a mongod.
	currentOpsCommand := struct {
		CurrentOp int32    `bson:"currentOp"`
		OwnOps    int32    `bson:"$ownOps,omitempty"`
		Or        []bson.M `bson:"$or,omitempty"`
	}{
		CurrentOp: 1,
		Or: bsonutil.NewMArray(
			bsonutil.NewM(bsonutil.NewDocElem("client", bsonutil.NewM(bsonutil.NewDocElem("$in", clientAddresses)))),
			bsonutil.NewM(bsonutil.NewDocElem("command.$client.mongos.client", bsonutil.NewM(bsonutil.NewDocElem("$in", clientAddresses)))),
		),
	}

	// If auth source is empty, this indicates we're running in unauthenticated mode. We should
	// not use the $ownOps parameter in this case since operations don't have any associated
	// MongoDB users.
	if s.AuthSource() != "" {
		currentOpsCommand.OwnOps = 1
	}

	var currentOpResponse struct {
		InProg []currentOp `bson:"inprog"`
	}

	err := s.Run(ctx, "admin", currentOpsCommand, &currentOpResponse)
	if err != nil {
		return nil, err
	}
	return currentOpResponse.InProg, nil
}

// killOp kills an operation on the server with the input opID.
func (s *Session) killOp(ctx context.Context, opID interface{}) error {
	killOpCommand := struct {
		KillOp int         `bson:"killOp"`
		Op     interface{} `bson:"op"`
	}{
		KillOp: 1,
		Op:     opID,
	}
	return s.Run(ctx, "admin", killOpCommand, &struct{}{})
}

// Login authenticates the session using the specified authenticator.
func (s *Session) Login(ctx context.Context, a SessionAuthenticator) error {
	var conns []conn.Connection

	s.authSource = a.source()

	// checkout all the connections
	for i := 0; i < s.count; i++ {
		c, err := s.Connection(ctx)
		if err != nil {
			s.err = err
			return s.err
		}
		defer func() {
			_ = c.Close()
		}()

		conns = append(conns, c)
	}

	s.err = a.Auth(ctx, conns)
	return s.err
}

// Version returns the server version for this
// session.
func (s *Session) Version() ([]uint8, error) {
	if len(s.server.Model().Version.Parts) == 0 {
		// Because the driver does not directly provide the server version, check
		// out a connection from the pool to get its version information.
		c, err := s.Connection(context.Background())
		if err != nil {
			return nil, err
		}
		version := c.Model().Server.Version.Parts
		err = c.Close()
		return version, err
	}
	return s.server.Model().Version.Parts, nil
}

// Run executes an arbitrary command against the given database.
func (s *Session) Run(ctx context.Context, db string, cmd interface{}, result interface{}) error {
	return ops.Run(ctx, s.selectedServer, db, cmd, result)
}
