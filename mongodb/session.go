package mongodb

import (
	"context"
	"fmt"

	"github.com/10gen/mongo-go-driver/conn"
	"github.com/10gen/mongo-go-driver/model"
	"github.com/10gen/mongo-go-driver/msg"
	"github.com/10gen/mongo-go-driver/ops"
)

// Cursor wraps the ops.Cursor interface for mongosqld
// and mongodrdl clients.
type Cursor interface {
	ops.Cursor
}

// Session holds information used to create a connection
// to MongoDB.
type Session struct {
	appName    string
	connection conn.Connection
	ctx        context.Context
	dialer     conn.NetDialer
	server     *ops.SelectedServer
}

// Alive returns true if the session connection is
// open and false otherwise.
func (s *Session) Alive() bool {
	cn, _ := s.server.Server.Connection(s.ctx)
	return cn.Alive()
}

// Close closes the direct server connection
// associated with this sesssion.
func (s *Session) Close() error {
	directServerConnection, ok := s.connection.(*DirectServerConnection)
	if !ok {
		return fmt.Errorf("unexpected session connection type: %T", s.connection)
	}
	return directServerConnection.directConnection.Close()
}

// ConnLen is the number of connections that are part of a session.
func (s *Session) ConnLen() int {
	return 1
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

// SelectedServer returns the selected server
// associated with this session.
func (s *Session) SelectedServer() *ops.SelectedServer {
	return s.server
}

// SetContext updates the context associated with this session.
func (s *Session) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// updateDirectConnection update the direct server connection
// for this session with the connection passed in c.
func (s *Session) updateDirectConnection(c conn.Connection) {
	directServerConnection := &DirectServerConnection{
		directConnection: c,
		serverModel:      s.server.Model(),
	}
	s.server = &ops.SelectedServer{
		Server:   serverImpl{directServerConnection},
		ReadPref: s.server.ReadPref,
	}
	s.connection = directServerConnection
}

// validate checks that the established session meets the server
// version requirements for the BI Connector.
func (s *Session) validate() error {
	version := s.Model().Version
	if !version.AtLeast(3, 2, 0) {
		return fmt.Errorf("server version is %v but version >= 3.2.0 required", version.Desc)
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
		return ops.Aggregate(s.ctx, s.server, ns, pipeline, opts)
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
		return ops.ListCollections(s.ctx, s.server, db, opts)
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
		return ops.ListIndexes(s.ctx, s.server, ns, opts)
	}
}

// Login authenticates the session using the specified authenticator.
func (s *Session) Login(a SessionAuthenticator) error {
	return a.Auth(s.ctx, []conn.Connection{s.connection})
}

// Run executes an arbitrary command against the given database.
func (s *Session) Run(db string, cmd interface{}, result interface{}) error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
		return ops.Run(s.ctx, s.server, db, cmd, result)
	}
}

// serverImpl implements the ops.Server interface.
type serverImpl struct {
	directConnection *DirectServerConnection
}

func (s serverImpl) Connection(ctx context.Context) (conn.Connection, error) {
	return s.directConnection, nil
}

func (s serverImpl) Model() *model.Server {
	return s.directConnection.serverModel
}

// DirectServerConnection maintains a direct connection
// to a MongoDB server by implementing the conn.Connection
// interface over a single socket connection.
type DirectServerConnection struct {
	directConnection conn.Connection
	serverModel      *model.Server
}

// NewDirectServerConnection establishes a direct connection to a MongoDB
// server using the supplied server description.
func NewDirectServerConnection(ctx context.Context, serverModel *model.Server, opts ...conn.Option) (*DirectServerConnection, error) {
	directConnection, err := conn.Dial(ctx, serverModel.Addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("could not dial server at %v: %v", serverModel.Addr, err)
	}
	return &DirectServerConnection{
		directConnection: directConnection,
		serverModel:      serverModel,
	}, nil
}

func (d *DirectServerConnection) Alive() bool {
	return d.directConnection.Alive()
}

func (d *DirectServerConnection) Close() error {
	return nil
}

func (d *DirectServerConnection) Model() *model.Conn {
	return d.directConnection.Model()
}

func (d *DirectServerConnection) Expired() bool {
	return d.directConnection.Expired()
}

func (d *DirectServerConnection) Read(ctx context.Context, responseTo int32) (msg.Response, error) {
	return d.directConnection.Read(ctx, responseTo)
}

func (d *DirectServerConnection) Write(ctx context.Context, reqs ...msg.Request) error {
	return d.directConnection.Write(ctx, reqs...)
}
