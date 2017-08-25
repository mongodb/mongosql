package server

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/sample"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	"github.com/shopspring/decimal"
)

// New creates a NewServer.
func New(schema *schema.Schema, sessionProvider *mongodb.SessionProvider, cfg *config.Config) (*Server, error) {

	decimal.DivisionPrecision = 34
	component := fmt.Sprintf("%-10v [initandlisten]", log.NetworkComponent)
	logger := log.NewComponentLogger(component, log.GlobalLogger())

	s := &Server{
		cfg:               cfg,
		terminateChan:     make(chan struct{}),
		activeConnections: make(map[uint32]*conn),
		schema:            schema,
		sessionProvider:   sessionProvider,
		variables:         variable.NewGlobalContainer(),
		logger:            logger,

		// status variable backing storage
		bytesReceived:    uint64(0),
		bytesSent:        uint64(0),
		connCount:        uint32(0),
		queryCount:       uint64(0),
		threadsConnected: uint32(0),
		startTime:        time.Now(),
	}

	if schema != nil {
		atomic.StoreUint32(&s.isSchemaLoaded, uint32(1))
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

	logger *log.Logger
	schema *schema.Schema

	// this is an atomic variable that is used
	// to indicate when an asynchronous load of
	// the schema has been completed
	isSchemaLoaded  uint32
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

func (s *Server) SampleSchema(ctx context.Context,
	opts *config.SchemaSampleOptions) {

	lgr := log.NewComponentLogger(
		fmt.Sprintf("%-10v [schemaDiscovery]", "SAMPLE"),
		log.GlobalLogger(),
	)

	waitToReconnect := func() { time.Sleep(5 * time.Second) }

	for {
		session, err := s.sessionProvider.Session(ctx)
		if err != nil {
			s.logger.Errf(log.Always, "error connecting to MongoDB: %v", err)
			waitToReconnect()
			continue
		}
		defer session.Close()

		var sampleRecord *sample.Record

		if opts.Mode == "read" {
			//
			// TODO (BI-1122): For read-only mongosqld, look up versions/schemas
			// collection. If unable, sample without inserting.
			//
			s.schema, err = sample.ReadSchema(opts, session, lgr)
			if err == nil {
				break
			} else if err == sample.ErrNotFound {
				s.schema, _, err = sample.SampleSchema(opts, session, lgr)
				if err == nil {
					break
				}
			}
		} else {
			//
			//	TODO (BI-1122): For read/write mongosqld sample and
			//	then attempt to insert if lock can be
			//	acquired. Otherwise, move on.
			//
			s.schema, sampleRecord, err = sample.SampleSchema(opts, session, lgr)
			if err == nil {
				// TODO (BI-1171): attempt to acquire database lock
				// if ErrLockAcquireFailed is returned, move on
				err = sample.AcquireLock(session, sampleRecord.Database)
				if err != nil && err == sample.ErrLockAcquireFailed {
					break
				}

				// TODO (BI-1171): also need to refresh lock 'ticket' regularly
				err = sample.InsertSampleRecord(sampleRecord, session, lgr)
				if err == nil {
					break
				}
			}
		}

		lgr.Errf(log.Always, err.Error())
		waitToReconnect()
	}

	atomic.StoreUint32(&s.isSchemaLoaded, uint32(1))
}

// Run starts the server and begins accepting connections.
func (s *Server) Run() {
	listenAndServe := func(listener net.Listener) {
		for atomic.LoadInt32(&s.closed) == 0 {
			conn, err := listener.Accept()
			if err != nil {
				if atomic.LoadInt32(&s.closed) == 0 {
					s.logger.Errf(log.Always, "unable to accept connection: %v", err)
				}
				continue
			}

			util.PanicSafeGo(func() {
				s.serveConnection(conn)
			}, func(err interface{}) {
				s.logger.Errf(log.Always, "unable to serve new connection: %v", err)
			})
		}
	}

	// asynchronously load the schema from MongoDB by sampling if needed
	if s.schema == nil {
		ctx, cancel := context.WithCancel(context.Background())
		util.PanicSafeGo(func() {
			s.SampleSchema(ctx, &s.cfg.Schema.Sample)
		}, func(err interface{}) {
			cancel()
			s.logger.Errf(log.Always, "error sampling schema: %v", err)
			s.Close()
		})
	}

	// start new goroutine for each listener
	for _, listener := range s.listeners {
		newListener := listener
		s.logger.Logf(log.Always, "waiting for connections at %v", newListener.Addr())
		util.PanicSafeGo(func() {
			listenAndServe(newListener)
		}, func(err interface{}) {
			s.logger.Errf(log.Always, "listen and serve error: %v", err)
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
		address := c.RemoteAddr().String()
		if address == "" {
			address = c.LocalAddr().String()
		}

		s.logger.Logf(log.Info, "connection accepted "+
			"from %v, but could not initialize: %v", address, err)
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
