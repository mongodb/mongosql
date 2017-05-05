package server

import (
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/util"
	"github.com/10gen/sqlproxy/variable"
	"github.com/shopspring/decimal"
)

// New creates a NewServer.
func New(schema *schema.Schema, sessionProvider *mongodb.SessionProvider, opts options.SqldOptions) (*Server, error) {

	decimal.DivisionPrecision = 34

	s := &Server{
		opts:              opts,
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

	s.variables.MongoDBVersionCompatibility = *opts.MongoVersionCompatibility

	if err := s.populateListeners(); err != nil {
		return nil, err
	}

	return s, nil
}

// Server manages connections with clients.
type Server struct {
	opts options.SqldOptions

	activeConnections   map[uint32]*conn
	activeConnectionsMx sync.RWMutex

	// synchronization variables for
	// terminating server
	terminateChan chan struct{}
	closed        int32
	terminated    bool

	schema          *schema.Schema
	sessionProvider *mongodb.SessionProvider
	variables       *variable.Container

	// backing server storage for status variables
	bytesReceived    uint64
	bytesSent        uint64
	connCount        uint32
	queryCount       uint64
	startTime        time.Time
	threadsConnected uint32

	listeners []net.Listener
}

// Run starts the server and begins accepting connections.
func (s *Server) Run() {

	logger := log.NewComponentLogger(log.NetworkComponent, log.GlobalLogger())

	listenAndServe := func(listener net.Listener) {
		for atomic.LoadInt32(&s.closed) == 0 {
			conn, err := listener.Accept()
			if err != nil {
				if atomic.LoadInt32(&s.closed) == 0 {
					logger.Errf(log.Always, "[initandlisten] uanble to accept connection: %v", err)
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

func (s *Server) addConnection(c *conn) {
	s.activeConnectionsMx.Lock()
	s.activeConnections[c.ConnectionId()] = c
	activeConnections := len(s.activeConnections)
	s.activeConnectionsMx.Unlock()
	atomic.StoreUint32(&s.threadsConnected, uint32(activeConnections))
	pluralized := util.Pluralize(activeConnections, "connection", "connections")
	c.logger.Logf(log.Always, "connection accepted from %v #%v (%v %v now open)", c.conn.RemoteAddr(), c.ConnectionId(), activeConnections, pluralized)
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
	c.logger.Logf(log.Always, "end connection %v (%v %v now open)", c.conn.RemoteAddr(), activeConnections, pluralized)
}

func (s *Server) serveConnection(c net.Conn) {
	conn, err := newConn(s, c)
	if err != nil {
		logger := log.GlobalLogger()
		logger.Logf(log.Info, "[initandlisten] connection accepted from %v, but unable to connect to MongoDB: %v", c.RemoteAddr(), err)
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
