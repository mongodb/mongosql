package server

import (
	"net"
	"runtime"
	"sync"

	"github.com/10gen/sqlproxy"
	"github.com/10gen/sqlproxy/client/openssl"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
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
	sync.Mutex

	eval              *sqlproxy.Evaluator
	opts              options.SqldOptions
	databases         map[string]*schema.Database
	activeConnections map[uint32]*conn
	variables         *variable.Container
	tlsConfig         *cfg.Ctx

	connCount uint32

	running bool

	listeners []net.Listener
}

// New creates a NewServer.
func New(schema *schema.Schema, eval *sqlproxy.Evaluator, opts options.SqldOptions) (*Server, error) {

	decimal.DivisionPrecision = 34

	s := &Server{
		eval:              eval,
		opts:              opts,
		running:           false,
		activeConnections: make(map[uint32]*conn),
		databases:         schema.Databases,
		variables:         variable.NewGlobalContainer(),
	}

	var err error

	err = s.populateListeners()
	if err != nil {
		return nil, err
	}

	if len(opts.SSLPEMKeyFile) > 0 {
		s.tlsConfig, err = openssl.SetupSqldCtx(opts, true)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Run starts the server and begins accepting connections.
func (s *Server) Run() {
	s.running = true
	logger := log.NewComponentLogger(log.NetworkComponent, log.GlobalLogger())

	// start new go routine for each additional listener
	for _, listener := range s.listeners[1:] {
		logger.Logf(log.Always, "[initandlisten] waiting for connections at %v", listener.Addr())

		go func() {
			for s.running {
				conn, err := listener.Accept()
				if err != nil {
					if s.running {
						logger.Logf(log.Always, "[initandlisten] %v", err)
					}
					continue
				}
				go s.onConn(conn)
			}

		}()
	}

	logger.Logf(log.Always, "[initandlisten] waiting for connections at %v", s.listeners[0].Addr())
	for s.running {
		conn, err := s.listeners[0].Accept()
		if err != nil {
			if s.running {
				logger.Logf(log.Always, "[initandlisten] %v", err)
			}
			continue
		}

		go s.onConn(conn)
	}

	// wait for all active client connections to return
	// cleanly before terminating
	for _, conn := range s.activeConnections {
		conn.Lock()
		<-conn.queryChan
		conn.Unlock()
		conn.close()
	}
}

// Close stops the server and stops accepting connections.
func (s *Server) Close() {
	s.running = false
	for _, listener := range s.listeners {
		if listener != nil {
			listener.Close()
		}
	}

	// interrrupt any in-progress queries
	s.Lock()
	for _, conn := range s.activeConnections {
		conn.tomb.Kill(mysqlerrors.Defaultf(mysqlerrors.ER_QUERY_INTERRUPTED))
	}
	s.Unlock()
}

func (s *Server) onConn(c net.Conn) {
	conn := newConn(s, c)

	// this isn't critical so neglecting to lock active connections
	numConns := len(s.activeConnections)
	pluralized := util.Pluralize(numConns, "connection", "connections")
	conn.logger.Logf(log.Info, "connection #%v accepted from %v (%v %v now open)", conn.connectionID, c.RemoteAddr(), numConns+1, pluralized)

	defer func() {
		if err := recover(); err != nil {
			const size = 4096
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			conn.logger.Logf(log.Always, "panic with %v: %v\n%s", c.RemoteAddr(), err, buf)
		}

		conn.close()
	}()

	err := conn.handshake()
	if err != nil {
		conn.logger.Errf(log.Always, "handshake error: %v", err)
		c.Close()
		return
	}

	schema := s.eval.Schema()
	conn.variables.MongoDBInfo, err = mongodb.LoadInfo(conn.Session(), &schema, s.opts.Auth)
	if err != nil {
		conn.logger.Errf(log.Always, "error retrieving information from MongoDB: %v", err)
		c.Close()
		return
	}

	if conn.startDb != "" {
		if err := conn.useDB(conn.startDb); err != nil {
			conn.logger.Errf(log.Always, "error connecting to db %v: %v", conn.startDb, err)
			c.Close()
			return
		}
	}

	s.Lock()
	s.activeConnections[conn.ConnectionId()] = conn
	s.Unlock()

	conn.run()
}
