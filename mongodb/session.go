package mongodb

import (
	"context"
	"fmt"

	"github.com/10gen/mongo-go-driver/cluster"
	"github.com/10gen/mongo-go-driver/conn"
	"github.com/10gen/mongo-go-driver/model"
	"github.com/10gen/mongo-go-driver/ops"
	"github.com/10gen/mongo-go-driver/readpref"
)

// Cursor wraps the ops.Cursor interface for mongosqld
// and mongodrdl clients.
type Cursor interface {
	ops.Cursor
}

// Session holds information used to create a connection
// to MongoDB.
type Session struct {
	ctx      context.Context
	cluster  *cluster.Cluster
	server   cluster.Server
	provider conn.Provider
	pool     conn.Pool
	count    int

	authSource     string
	selectedServer *ops.SelectedServer
}

// Close closes the direct server connection
// associated with this sesssion.
func (s *Session) Close() error {
	s.pool.Close()
	return nil
}

// Connection gets a connection to use.
func (s *Session) Connection(ctx context.Context) (conn.Connection, error) {
	return s.provider(ctx)
}

// ConnLen is the number of connections that are part of a session.
func (s *Session) ConnLen() int {
	return s.count
}

// Context returns the context associated with this session.
func (s *Session) Context() context.Context {
	return s.ctx
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
	version := s.Model().Version
	if !version.AtLeast(3, 2, 0) {
		return fmt.Errorf("server version is %v but version >= 3.2.0 required", version.Desc)
	}

	selector := readpref.Selector(s.selectedServer.ReadPref)
	result, err := selector(s.cluster.Model(), []*model.Server{s.server.Model()})
	if err != nil || len(result) == 0 {
		return fmt.Errorf("current session does not satisfy read preference")
	}

	return nil
}

//
// Session helper methods
//

// Aggregate runs the aggregation pipeline passed in against the
// give database and collection.
func (s *Session) Aggregate(database, collection string, pipeline interface{}) (ops.Cursor, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		ns := ops.NewNamespace(database, collection)
		opts := ops.AggregationOptions{AllowDiskUse: true}
		return ops.Aggregate(s.ctx, s.selectedServer, ns, pipeline, opts)
	}
}

// ListCollections returns a cursor to iterate through
// the collections present on the db database.
func (s *Session) ListCollections(db string) (ops.Cursor, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		opts := ops.ListCollectionsOptions{}
		return ops.ListCollections(s.ctx, s.selectedServer, db, opts)
	}
}

// ListDatabases returns a cursor to iterate through
// the database names present on a server.
func (s *Session) ListDatabases() (ops.Cursor, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		opts := ops.ListDatabasesOptions{}
		return ops.ListDatabases(s.ctx, s.selectedServer, opts)
	}
}

// ListIndexes returns a cursor to iterate through
// the indexes on the c collection within the db database.
func (s *Session) ListIndexes(db, c string) (ops.Cursor, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	default:
		opts := ops.ListIndexesOptions{}
		ns := ops.NewNamespace(db, c)
		return ops.ListIndexes(s.ctx, s.selectedServer, ns, opts)
	}
}

// Login authenticates the session using the specified authenticator.
func (s *Session) Login(a SessionAuthenticator) error {
	var conns []conn.Connection

	s.authSource = a.source()

	// checkout all the connections
	for i := 0; i < s.count; i++ {
		c, err := s.Connection(s.ctx)
		if err != nil {
			return err
		}
		defer c.Close()

		conns = append(conns, c)
	}

	return a.Auth(s.ctx, conns)
}

// Run executes an arbitrary command against the given database.
func (s *Session) Run(db string, cmd interface{}, result interface{}) error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
		return ops.Run(s.ctx, s.selectedServer, db, cmd, result)
	}
}
