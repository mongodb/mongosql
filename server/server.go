package server

import (
	"net"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/client/openssl"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/util"
	"github.com/10gen/sqlproxy/variable"
	"github.com/shopspring/decimal"
	cfg "github.com/spacemonkeygo/openssl"
)

// Server manages connections with clients.
type Server struct {
	eval                *sqlproxy.Evaluator
	opts                options.SqldOptions
	connCount           uint32
	activeConnections   map[uint32]*conn
	activeConnectionsMx sync.RWMutex

	// synchronization variables for
	// terminating server
	terminateChan chan struct{}
	closed        int32
	terminated    bool

	variables *variable.Container
	tlsConfig *cfg.Ctx

	listeners []net.Listener
}

// New creates a NewServer.
func New(schema *schema.Schema, eval *sqlproxy.Evaluator, opts options.SqldOptions) (*Server, error) {

	decimal.DivisionPrecision = 34

	s := &Server{
		eval:              eval,
		opts:              opts,
		closed:            0,
		terminated:        false,
		terminateChan:     make(chan struct{}),
		activeConnections: make(map[uint32]*conn),
		variables:         variable.NewGlobalContainer(),
	}

	s.variables.MongoDBStrictDecimalParsingOn32 = opts.StrictDecimalParsingOn32

	var err error

	if len(opts.SSLPEMKeyFile) > 0 {
		s.tlsConfig, err = openssl.SetupSqldCtx(opts, true)
		if err != nil {
			return nil, err
		}
	}

	err = s.populateListeners()
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Run starts the server and begins accepting connections.
func (s *Server) Run() {

	logger := log.NewComponentLogger(log.NetworkComponent, log.GlobalLogger())

	listenAndServe := func(listener net.Listener) {

		for atomic.LoadInt32(&s.closed) == 0 {
			conn, err := listener.Accept()
			if err != nil {
				if atomic.LoadInt32(&s.closed) == 0 {
					logger.Logf(log.Always, "[initandlisten] %v", err)
				}
				continue
			}
			go s.serveConnection(conn)
		}
	}

	// start new goroutine for each listener
	for _, listener := range s.listeners {
		logger.Logf(log.Always, "[initandlisten] waiting for connections at %v", listener.Addr())
		go listenAndServe(listener)
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
	pluralized := util.Pluralize(activeConnections, "connection", "connections")
	c.logger.Logf(log.Info, "connection #%v accepted from %v (%v %v now open)", c.ConnectionId(), c.conn.RemoteAddr(), activeConnections, pluralized)
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
	c.tomb.Kill(mysqlerrors.Defaultf(mysqlerrors.ER_QUERY_INTERRUPTED))
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
	pluralized := util.Pluralize(activeConnections, "connection", "connections")
	c.logger.Logf(log.Always, "end connection %v (%v %v now open)", c.conn.RemoteAddr(), activeConnections, pluralized)
}

func (s *Server) serveConnection(c net.Conn) {
	conn := newConn(s, c)

	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			conn.logger.Errf(log.Info, "%v, %s", err, buf)
		}
		c.Close()
		conn.close()
	}()

	s.addConnection(conn)

	if err := conn.handshake(); err != nil {
		conn.logger.Errf(log.Always, "handshake error: %v", err)
		return
	}

	conn.run()
}
