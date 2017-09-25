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
		variables:         variable.NewGlobalContainer(),
		logger:            logger,
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = strings.Join(s.cfg.Net.BindIP, ",")
	}

	s.processName = fmt.Sprintf("mongosqld-%s-%d-%s", hostname, os.Getpid(), randomString(6))

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
	atomic.StoreUint32(s.variables.ThreadsConnected, uint32(activeConnections))
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
