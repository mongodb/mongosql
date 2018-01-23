package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

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

	lifetimeCtx, lifetimeCancel := context.WithCancel(context.Background())
	s := &Server{
		cfg:               cfg,
		lifetimeCtx:       lifetimeCtx,
		lifetimeCancel:    lifetimeCancel,
		activeConnections: make(map[uint32]*conn),
		fileBasedSchema:   schema,
		sessionProvider:   sessionProvider,
		variables:         variable.NewGlobalContainer(cfg),
		logger:            logger,
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = strings.Join(s.cfg.Net.BindIP, ",")
	}

	s.processName = fmt.Sprintf("mongosqld-%s-%d-%s", hostname, os.Getpid(), randomString(6))

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
	lifetimeCtx    context.Context
	lifetimeCancel func()
	closed         int32

	logger *log.Logger

	processName string

	fileBasedSchema *schema.Schema
	sampler         *sample.Sampler

	sessionProvider *mongodb.SessionProvider
	variables       *variable.Container

	// startupInfo holds configuration information
	// for the server
	startupInfo []string

	listeners []net.Listener
}

func (s *Server) getSchema() *schema.Schema {
	if s.fileBasedSchema != nil {
		// schema was loaded from a DRDL file
		return s.fileBasedSchema
	}

	return s.sampler.Schema(s.lifetimeCtx)
}

// Alter applies the provided alterations to the server's current schema.
func (s *Server) Alter(ctx context.Context, alts []*schema.Alteration) (*schema.Schema, error) {
	if s.fileBasedSchema != nil {
		return nil, fmt.Errorf("cannot alter schema: schema was loaded from a file")
	}

	if !s.cfg.SetParameter.EnableTableAlterations {
		return nil, fmt.Errorf("cannot alter schema: alterations not enabled")
	}

	err := s.sampler.Alter(ctx, alts)
	if err != nil {
		return nil, err
	}
	return s.getSchema(), nil
}

// Resample forces a sample refresh.
func (s *Server) Resample(ctx context.Context) (*schema.Schema, error) {
	if s.fileBasedSchema != nil {
		return nil, fmt.Errorf("sampling is disabled; schema was loaded from a file")
	}

	err := s.sampler.Refresh(ctx)
	if err != nil {
		return nil, err
	}

	return s.getSchema(), nil
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
	if s.fileBasedSchema == nil {
		s.sampler = sample.NewSampler(&s.cfg.Schema.Sample, s.processName, s.sessionProvider)
		util.PanicSafeGo(func() {
			s.sampler.Run(s.lifetimeCtx)
		}, func(err interface{}) {
			s.logger.Fatalf(log.Always, "error sampling schema: %v", err)
			s.Close()
		})
	}

	// start new goroutine for each listener
	for _, listener := range s.listeners {
		newListener := listener
		s.logger.Infof(log.Always, "waiting for connections at %v", newListener.Addr())
		util.PanicSafeGo(func() {
			listenAndServe(newListener)
		}, func(err interface{}) {
			s.logger.Errf(log.Always, "listen and serve error: %v", err)
		})
	}

	// wait for all active client connections
	// to return cleanly before terminating
	<-s.lifetimeCtx.Done()
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

	// interrupt any in-progress queries
	s.activeConnectionsMx.RLock()
	if len(s.activeConnections) == 0 {
		s.lifetimeCancel()
	}

	for _, c := range s.activeConnections {
		c.close()
	}

	s.activeConnectionsMx.RUnlock()
}

// StartupInfo gets the startup information for logging.
func (s *Server) StartupInfo() []string {
	return s.startupInfo
}

// StoreStartupInfo stores startup information in order
// to log after a log rotation.
func (s *Server) StoreStartupInfo(startupInfo []string) {
	s.startupInfo = startupInfo
}

func (s *Server) addConnection(c *conn) {
	s.activeConnectionsMx.Lock()
	s.activeConnections[c.ConnectionID()] = c
	activeConnections := len(s.activeConnections)
	s.activeConnectionsMx.Unlock()
	atomic.StoreUint32(s.variables.ThreadsConnected, uint32(activeConnections))
	pluralized := util.Pluralize(activeConnections, "connection", "connections")
	address := c.conn.RemoteAddr().String()
	if address == "" {
		address = c.conn.LocalAddr().String()
	}
	c.process.SetHost(c.getFormattedAddress())
	c.logger.Infof(log.Always, "connection accepted from %v #%v (%v %v now open)", address, c.ConnectionID(), activeConnections, pluralized)
}

func (s *Server) killConnection(targetConnID uint32, requestingConnID uint32) error {
	s.activeConnectionsMx.RLock()
	targetConn, ok := s.activeConnections[targetConnID]
	if !ok {
		s.activeConnectionsMx.RUnlock()
		return mysqlerrors.Defaultf(mysqlerrors.ER_NO_SUCH_THREAD, targetConnID)
	}

	requestingConn, ok := s.activeConnections[requestingConnID]
	if !ok {
		s.activeConnectionsMx.RUnlock()
		return mysqlerrors.Defaultf(mysqlerrors.ER_NO_SUCH_THREAD, requestingConnID)
	}

	if requestingConn.user != targetConn.user ||
		requestingConn.Session().AuthSource() != targetConn.Session().AuthSource() {
		s.activeConnectionsMx.RUnlock()
		return mysqlerrors.Defaultf(mysqlerrors.ER_KILL_DENIED_ERROR, targetConnID)
	}
	s.activeConnectionsMx.RUnlock()

	targetConn.close()
	return nil
}

func (s *Server) killQuery(targetConnID uint32, requestingConnID uint32) error {
	s.activeConnectionsMx.RLock()
	targetConn, ok := s.activeConnections[targetConnID]
	if !ok {
		s.activeConnectionsMx.RUnlock()
		return mysqlerrors.Defaultf(mysqlerrors.ER_NO_SUCH_THREAD, targetConnID)
	}

	requestingConn, ok := s.activeConnections[requestingConnID]
	if !ok {
		s.activeConnectionsMx.RUnlock()
		return mysqlerrors.Defaultf(mysqlerrors.ER_NO_SUCH_THREAD, requestingConnID)
	}

	if requestingConn.user != targetConn.user ||
		requestingConn.Session().AuthSource() != targetConn.Session().AuthSource() {
		s.activeConnectionsMx.RUnlock()
		return mysqlerrors.Defaultf(mysqlerrors.ER_KILL_DENIED_ERROR, targetConnID)
	}

	s.activeConnectionsMx.RUnlock()

	// If KillOps fails in killing any operation for the client addresses, we
	// still cancel the target connection's context. This is because the alternative of having the target query
	// running without cancelling the context will prevent the user on the target connection from issuing subsequent
	// queries until the current query is completed. It is preferable to allow subsequent queries to be issued,
	// with the possiblity of no connections being available in the target connection's
	// session to execute them, than to not allow any queries to be accepted.
	clientAddresses := targetConn.Session().GetClientAddresses()

	// Cancel the connection before doing KillOps for testing purposes to prevent receiving a QueryPlanKilled error from MongoDB.
	targetConn.cancel()
	return requestingConn.Session().KillOps(requestingConn.logger, clientAddresses)
}

func (s *Server) removeConnection(c *conn) {
	s.activeConnectionsMx.Lock()
	delete(s.activeConnections, c.ConnectionID())
	if atomic.LoadInt32(&s.closed) == 1 && len(s.activeConnections) == 0 {
		s.lifetimeCancel()
	}
	activeConnections := len(s.activeConnections)
	s.activeConnectionsMx.Unlock()
	atomic.StoreUint32(s.variables.ThreadsConnected, uint32(activeConnections))
	pluralized := util.Pluralize(activeConnections, "connection", "connections")
	address := c.conn.RemoteAddr().String()
	if address == "" {
		address = c.conn.LocalAddr().String()
	}
	c.logger.Infof(log.Always, "end connection %v (%v %v now open)", address, activeConnections, pluralized)
}

func (s *Server) serveConnection(c net.Conn) {
	conn, err := newConn(s, c)
	if err != nil {
		address := c.RemoteAddr().String()
		if address == "" {
			address = c.LocalAddr().String()
		}

		s.logger.Errf(log.Always, "connection accepted "+
			"from %v, but could not initialize: %v", address, err)
		c.Close()
		return
	}

	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			conn.logger.Errf(log.Dev, "error serving connection: %v, %s", err, buf)
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
