package server

import (
	"bytes"
	"fmt"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mongodrdl"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	"github.com/shopspring/decimal"
)

// New creates a NewServer.
func New(schema *schema.Schema, sessionProvider *mongodb.SessionProvider, cfg *config.Config) (*Server, error) {

	decimal.DivisionPrecision = 34

	s := &Server{
		cfg:               cfg,
		terminateChan:     make(chan struct{}),
		activeConnections: make(map[uint32]*conn),
		schema:            schema,
		sessionProvider:   sessionProvider,
		variables:         variable.NewGlobalContainer(),

		// status variable backing storage
		bytesReceived:    uint64(0),
		bytesSent:        uint64(0),
		connCount:        uint32(0),
		queryCount:       uint64(0),
		threadsConnected: uint32(0),
		startTime:        time.Now(),
	}

	if schema != nil {
		atomic.StoreUint32(&s.schemaLoaded, uint32(1))
	}

	s.variables.MongoDBMaxStageSize = cfg.Runtime.Memory.MaxPerStage
	s.variables.MongoDBMaxVarcharLength = cfg.Schema.MaxVarcharLength
	s.variables.MongoDBVersionCompatibility = cfg.MongoDB.VersionCompatibility

	if err := s.populateListeners(); err != nil {
		return nil, err
	}

	return s, nil
}

// Server manages connections with clients.
type Server struct {
	cfg *config.Config

	activeConnections   map[uint32]*conn
	activeConnectionsMx sync.RWMutex

	// synchronization variables for
	// terminating server
	terminateChan chan struct{}
	closed        int32
	terminated    bool

	schema          *schema.Schema
	schemaLoaded    uint32
	sessionProvider *mongodb.SessionProvider
	variables       *variable.Container

	// backing server storage for status variables
	bytesReceived    uint64
	bytesSent        uint64
	connCount        uint32
	queryCount       uint64
	startTime        time.Time
	threadsConnected uint32

	// startupInfo holds configuration information
	// for the server
	startupInfo []string

	listeners []net.Listener
}

func (s *Server) SampleSchema(cfg *config.Config) {
	schemaGenerator := createSchemaGenerator(cfg)

	err := schemaGenerator.Init()
	if err != nil {
		schemaGenerator.Logger.Warnf(log.Always, "[initandlisten] error initializing schema: %v", err)
		panic(fmt.Sprintf("unable to initialize schema generator: %v", err))
	}

	defer schemaGenerator.Provider.Close()

	databases := make(map[string][]string, 0)

	var session *mongodb.Session

getDatabasesLoop:
	for {
	dbConnectLoop:
		for {
			session, err = schemaGenerator.Connect()
			if err != nil {
				schemaGenerator.Logger.Warnf(log.Always, "error connecting to MongoDB: %v", err)
				continue
			}
			defer session.Close()
			break dbConnectLoop
		}

		ctx := session.Context()

		dbIter, err := session.ListDatabases()
		if err != nil {
			schemaGenerator.Logger.Warnf(log.Always, "error listing databases: %v", err)
			panic(fmt.Sprintf("error listing databases: %v", err))
		}

		var dbResult struct {
			Name string `bson:"name"`
		}

		for dbIter.Next(ctx, &dbResult) {
			collectionIter, err := session.ListCollections(dbResult.Name)
			if err != nil {
				schemaGenerator.Logger.Errf(log.Always, "can't get the collection names for '%v': %v", dbResult.Name, err)
				panic(fmt.Sprintf("can't get the collection names for '%v': %v", dbResult.Name, err))
			}

			var collectionResult struct {
				Name string `bson:"name"`
			}

			ctx := session.Context()

			collections := []string{}

			for collectionIter.Next(ctx, &collectionResult) {
				collections = append(collections, collectionResult.Name)
			}

			if err := collectionIter.Close(ctx); err != nil {
				schemaGenerator.Logger.Errf(log.Always, "collection iteration close: %v", err)
				continue
			}

			if err := collectionIter.Err(); err != nil {
				schemaGenerator.Logger.Errf(log.Always, "collection iteration error: %v", err)
				continue
			}

			databases[dbResult.Name] = collections
		}

		if err := dbIter.Close(ctx); err != nil {
			schemaGenerator.Logger.Errf(log.Always, "db iteration close: %v", err)
			continue
		}

		if err := dbIter.Err(); err != nil {
			schemaGenerator.Logger.Errf(log.Always, "db iteration error: %v", err)
			continue
		}

		break getDatabasesLoop
	}

	sampledSchema := &schema.Schema{}

	dbSampleBlacklist := []string{"admin", "local"}

	buf := &bytes.Buffer{}

	namespaces := cfg.Schema.Sample.Namespaces
	if len(namespaces) == 0 {
		namespaces = []string{"*"}
	}

	nsMatcher, err := util.NewMatcher(namespaces)
	if err != nil {
		schemaGenerator.Logger.Errf(log.Always, "invalid specification: %v", err)
		panic(fmt.Sprintf("invalid specification: %v", err))
	}

	for db, collections := range databases {
		if util.SliceContains(dbSampleBlacklist, db) {
			continue
		}

		for _, collection := range collections {
			schemaGenerator.ToolOptions.DrdlNamespace.DB = db
			schemaGenerator.ToolOptions.DrdlNamespace.Collection = collection

			ns := db + "." + collection

			if !nsMatcher.Has(ns) {
				schemaGenerator.Logger.Logf(log.Always, "Skipping namespace %v", ns)
				continue
			}

			_, err := schemaGenerator.GenerateWithWriter(buf)
			if err != nil {
				schemaGenerator.Logger.Errf(log.Always, "error generating schema: %v", err)
				panic(fmt.Sprintf("error generating schema: %v", err))
			}

			// load this namespace into the schema
			err = sampledSchema.Load(buf.Bytes())
			if err != nil {
				schemaGenerator.Logger.Errf(log.Always, "error loading schema file: %v", err)
				panic(fmt.Sprintf("error loading schema: %v", err))
			}

			buf.Reset()
		}
	}
	schemaGenerator.Logger.Logf(log.Always, "Schema sampled from MongoDB")
	s.schema = sampledSchema
	atomic.StoreUint32(&s.schemaLoaded, uint32(1))
}

func createSchemaGenerator(cfg *config.Config) *mongodrdl.SchemaGenerator {
	component := "MONGODRDL"
	schemaGenerator := &mongodrdl.SchemaGenerator{
		ToolOptions: &options.DrdlOptions{
			DrdlAuth: &options.DrdlAuth{
				Username:  cfg.MongoDB.Net.Auth.Username,
				Password:  cfg.MongoDB.Net.Auth.Password,
				Source:    cfg.MongoDB.Net.Auth.Source,
				Mechanism: cfg.MongoDB.Net.Auth.Mechanism,
			},
			DrdlConnection: &options.DrdlConnection{
				Host: cfg.MongoDB.Net.URI,
			},
			DrdlLog: &options.DrdlLog{
				VLevel: cfg.SystemLog.Verbosity,
			},
			DrdlSSL: &options.DrdlSSL{
				UseSSL:              cfg.MongoDB.Net.SSL.Enabled,
				SSLAllowInvalidCert: cfg.MongoDB.Net.SSL.AllowInvalidCertificates,
				SSLAllowInvalidHost: cfg.MongoDB.Net.SSL.AllowInvalidHostnames,
				SSLPEMKeyFile:       cfg.MongoDB.Net.SSL.PEMKeyFile,
				SSLPEMKeyPassword:   cfg.MongoDB.Net.SSL.PEMKeyPassword,
				SSLCAFile:           cfg.MongoDB.Net.SSL.CAFile,
				SSLCRLFile:          cfg.MongoDB.Net.SSL.CRLFile,
				SSLFipsMode:         cfg.MongoDB.Net.SSL.FIPSMode,
			},
			DrdlNamespace: &options.DrdlNamespace{},
		},
		OutputOptions: &options.DrdlOutput{
			UUIDSubtype3Encoding: cfg.Schema.Sample.UUIDSubtype3Encoding,
		},
		SampleOptions: &options.DrdlSample{
			Size: cfg.Schema.Sample.Size,
		},
		Logger: log.NewComponentLogger(
			fmt.Sprintf("%-10v [schemaDiscovery]", component),
			log.GlobalLogger(),
		),
	}

	return schemaGenerator
}

// Run starts the server and begins accepting connections.
func (s *Server) Run() {

	logger := log.NewComponentLogger(log.NetworkComponent, log.GlobalLogger())

	listenAndServe := func(listener net.Listener) {
		for atomic.LoadInt32(&s.closed) == 0 {
			conn, err := listener.Accept()
			if err != nil {
				if atomic.LoadInt32(&s.closed) == 0 {
					logger.Errf(log.Always, "[initandlisten] unable to accept connection: %v", err)
				}
				continue
			}

			util.PanicSafeGo(func() {
				s.serveConnection(conn)
			}, func(err interface{}) {
				logger.Errf(log.Always, "[initandlisten] unable to serve new connection: %v", err)
			})
		}
	}

	// start new goroutine for each listener
	for _, listener := range s.listeners {
		newListener := listener
		logger.Logf(log.Always, "[initandlisten] waiting for connections at %v", newListener.Addr())
		util.PanicSafeGo(func() {
			listenAndServe(newListener)
		}, func(err interface{}) {
			logger.Errf(log.Always, "[initandlisten] listen and serve error: %v", err)
		})
	}

	// wait for all active client connections
	// to return cleanly before terminating
	<-s.terminateChan
}

// Close stops the server and stops accepting connections.
func (s *Server) Close() {
	if !atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		return
	}

	for _, listener := range s.listeners {
		if listener != nil {
			listener.Close()
		}
	}

	// interrrupt any in-progress queries
	s.activeConnectionsMx.RLock()
	if len(s.activeConnections) == 0 && !s.terminated {
		close(s.terminateChan)
	}

	for _, c := range s.activeConnections {
		c.close()
	}
	s.activeConnectionsMx.RUnlock()
}

// GetStartupInfo returns startup information for the server.
func (s *Server) GetStartupInfo() []string {
	return s.startupInfo
}

// StoreStartupInfo stores startup information in order
// to log after a log rotation.
func (s *Server) StoreStartupInfo(startupInfo []string) {
	s.startupInfo = startupInfo
}

func (s *Server) addConnection(c *conn) {
	s.activeConnectionsMx.Lock()
	s.activeConnections[c.ConnectionId()] = c
	activeConnections := len(s.activeConnections)
	s.activeConnectionsMx.Unlock()
	atomic.StoreUint32(&s.threadsConnected, uint32(activeConnections))
	pluralized := util.Pluralize(activeConnections, "connection", "connections")
	address := c.conn.RemoteAddr().String()
	if address == "" {
		address = c.conn.LocalAddr().String()
	}
	c.logger.Logf(log.Always, "connection accepted from %v #%v (%v %v now open)", address, c.ConnectionId(), activeConnections, pluralized)
}

func (s *Server) killConnection(connID uint32) error {
	s.activeConnectionsMx.RLock()
	c, ok := s.activeConnections[connID]
	if !ok {
		s.activeConnectionsMx.RUnlock()
		return mysqlerrors.Defaultf(mysqlerrors.ER_NO_SUCH_THREAD, connID)
	}
	c.close()
	s.activeConnectionsMx.RUnlock()
	return nil
}

func (s *Server) killQuery(connID uint32) error {
	s.activeConnectionsMx.RLock()
	c, ok := s.activeConnections[connID]
	if !ok {
		s.activeConnectionsMx.RUnlock()
		return mysqlerrors.Defaultf(mysqlerrors.ER_NO_SUCH_THREAD, connID)
	}

	c.cancel()
	s.activeConnectionsMx.RUnlock()
	return nil
}

func (s *Server) removeConnection(c *conn) {
	s.activeConnectionsMx.Lock()
	delete(s.activeConnections, c.ConnectionId())
	if atomic.LoadInt32(&s.closed) == 1 && len(s.activeConnections) == 0 {
		s.terminated = true
		close(s.terminateChan)
	}
	activeConnections := len(s.activeConnections)
	s.activeConnectionsMx.Unlock()
	atomic.StoreUint32(&s.threadsConnected, uint32(activeConnections))
	pluralized := util.Pluralize(activeConnections, "connection", "connections")
	address := c.conn.RemoteAddr().String()
	if address == "" {
		address = c.conn.LocalAddr().String()
	}
	c.logger.Logf(log.Always, "end connection %v (%v %v now open)", address, activeConnections, pluralized)
}

func (s *Server) serveConnection(c net.Conn) {
	conn, err := newConn(s, c)
	if err != nil {
		logger := log.GlobalLogger()
		address := c.RemoteAddr().String()
		if address == "" {
			address = c.LocalAddr().String()
		}
		logger.Logf(log.Info, "[initandlisten] connection accepted from %v, but unable to connect to MongoDB: %v", address, err)
		c.Close()
		return
	}

	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			conn.logger.Errf(log.Info, "%v, %s", err, buf)
		}
		conn.close()
	}()

	s.addConnection(conn)

	atomic.StoreInt32(&conn.queryRunning, 1)
	if err := conn.handshake(); err != nil {
		conn.logger.Errf(log.Always, "handshake error: %v", err)
		atomic.StoreInt32(&conn.queryRunning, 0)
		return
	}
	atomic.StoreInt32(&conn.queryRunning, 0)

	conn.run()
}
