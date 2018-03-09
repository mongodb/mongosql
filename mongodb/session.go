package mongodb

import (
	"context"
	"fmt"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/mongo-go-driver/mongo"
	"github.com/10gen/mongo-go-driver/mongo/model"
	"github.com/10gen/mongo-go-driver/mongo/private/cluster"
	"github.com/10gen/mongo-go-driver/mongo/private/conn"
	"github.com/10gen/mongo-go-driver/mongo/private/ops"
	"github.com/10gen/mongo-go-driver/mongo/readconcern"
	"github.com/10gen/mongo-go-driver/mongo/readpref"
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
	ctx             context.Context
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

// Context returns the context associated with this session.
func (s *Session) Context() context.Context {
	return s.ctx
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

// SetContext updates the context associated with this session.
func (s *Session) SetContext(ctx context.Context) {
	s.ctx = ctx
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
func (s *Session) Aggregate(database, collection string, pipeline interface{}) (ops.Cursor, error) {
	select {
	case <-s.Context().Done():
		return nil, s.Context().Err()
	default:
		ns := ops.NewNamespace(database, collection)
		opts := mongo.AllowDiskUse(true)
		return ops.Aggregate(s.Context(), s.selectedServer, ns, readconcern.Local(), pipeline, opts)
	}
}

// Count runs a count command for a specific database and collection.
func (s *Session) Count(database, collection string) (int, error) {
	select {
	case <-s.Context().Done():
		return 0, s.Context().Err()
	default:
		return ops.Count(
			s.Context(),
			s.selectedServer,
			ops.NewNamespace(database, collection),
			nil,
			nil,
		)
	}
}

// ListCollections returns a cursor to iterate through
// the collections present on the db database with options opts.
func (s *Session) ListCollections(db string, opts ops.ListCollectionsOptions) (ops.Cursor, error) {
	select {
	case <-s.Context().Done():
		return nil, s.Context().Err()
	default:
		return ops.ListCollections(s.Context(), s.selectedServer, db, opts)
	}
}

// ListDatabases returns a cursor to iterate through
// the database names present on a server.
func (s *Session) ListDatabases() (ops.Cursor, error) {
	select {
	case <-s.Context().Done():
		return nil, s.Context().Err()
	default:
		opts := ops.ListDatabasesOptions{}
		return ops.ListDatabases(s.Context(), s.selectedServer, opts)
	}
}

// ListIndexes returns a cursor to iterate through
// the indexes on the c collection within the db database.
func (s *Session) ListIndexes(db, c string) (ops.Cursor, error) {
	select {
	case <-s.Context().Done():
		return nil, s.Context().Err()
	default:
		opts := ops.ListIndexesOptions{}
		ns := ops.NewNamespace(db, c)
		return ops.ListIndexes(s.Context(), s.selectedServer, ns, opts)
	}
}

// KillOps kills all operations running on a list of client addresses.
func (s *Session) KillOps(clientAddresses []string) error {
	select {
	case <-s.Context().Done():
		return s.Context().Err()
	default:

		if len(clientAddresses) == 0 {
			return nil
		}

		currentOpsToKill, err := s.listCurrentOpsForClients(clientAddresses)
		if err != nil {
			return err
		}

		for _, op := range currentOpsToKill {
			err := s.killOp(op.Opid)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// listCurrentOpsForClients returns all operations belonging to the provided
// session's user from a list of client addresses.
func (s *Session) listCurrentOpsForClients(clientAddresses []string) ([]currentOp, error) {
	select {
	case <-s.Context().Done():
		return nil, s.Context().Err()
	default:

		// The two conditions in the $or condition handle whether we are talking to a mongod or a
		// mongos. A mongos reports its client addreses in a different format than a mongod.
		currentOpsCommand := struct {
			CurrentOp int32    `bson:"currentOp"`
			OwnOps    int32    `bson:"$ownOps,omitempty"`
			Or        []bson.M `bson:"$or,omitempty"`
		}{
			CurrentOp: 1,
			Or: []bson.M{
				{"client": bson.M{"$in": clientAddresses}},
				{"command.$client.mongos.client": bson.M{"$in": clientAddresses}},
			},
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

		err := s.Run("admin", currentOpsCommand, &currentOpResponse)
		if err != nil {
			return nil, err
		}
		return currentOpResponse.InProg, nil
	}
}

// killOp kills an operation on the server with the input opID.
func (s *Session) killOp(opID interface{}) error {
	select {
	case <-s.Context().Done():
		return s.Context().Err()
	default:

		killOpCommand := struct {
			KillOp int         `bson:"killOp"`
			Op     interface{} `bson:"op"`
		}{
			KillOp: 1,
			Op:     opID,
		}
		return s.Run("admin", killOpCommand, &struct{}{})
	}
}

// Login authenticates the session using the specified authenticator.
func (s *Session) Login(a SessionAuthenticator) error {
	var conns []conn.Connection

	s.authSource = a.source()

	// checkout all the connections
	for i := 0; i < s.count; i++ {
		c, err := s.Connection(s.Context())
		if err != nil {
			s.err = err
			return s.err
		}
		defer func() {
			_ = c.Close()
		}()

		conns = append(conns, c)
	}

	s.err = a.Auth(s.Context(), conns)
	return s.err
}

// Run executes an arbitrary command against the given database.
func (s *Session) Run(db string, cmd interface{}, result interface{}) error {
	select {
	case <-s.Context().Done():
		return s.Context().Err()
	default:
		return ops.Run(s.Context(), s.selectedServer, db, cmd, result)
	}
}
