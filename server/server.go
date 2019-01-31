package server

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/memory"
	"github.com/10gen/sqlproxy/evaluator/metrics"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/internal/sample"
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
)

// New creates a NewServer.
func New(ctx context.Context, cnl context.CancelFunc, sch *schema.Schema, sp *mongodb.SessionProvider, cfg *config.Config) (*Server, error) {

	decimal.DivisionPrecision = 34
	component := fmt.Sprintf("%-10v [initandlisten]", log.NetworkComponent)
	logger := log.NewComponentLogger(component, log.GlobalLogger())

	s := &Server{
		cfg:               cfg,
		cancelCtx:         cnl,
		activeConnections: make(map[uint32]*conn),
		fileBasedSchema:   sch,
		sessionProvider:   sp,
		variables:         variable.NewGlobalContainer(cfg),
		logger:            logger,
		memoryMonitor:     memory.NewMonitor("Server", cfg.Runtime.Memory.MaxPerServer),
		recordsChan:       make(chan []metrics.Record),
	}

	s.variables.AllocatedMemory = s.memoryMonitor.Allocated
	s.variables.SetSystemVariable(variable.SampleSize, s.cfg.Schema.Sample.Size)
	s.variables.SetSystemVariable(
		variable.SampleRefreshIntervalSecs,
		s.cfg.Schema.Sample.RefreshIntervalSecs,
	)

	hostname, err := os.Hostname()
	if err != nil {
		hostname = strings.Join(s.cfg.Net.BindIP, ",")
	}

	s.processName = fmt.Sprintf("mongosqld-%s-%d-%s", hostname, os.Getpid(), randomString(6))

	if err := s.populateListeners(); err != nil {
		return nil, fmt.Errorf("error populating listeners: %v", err)
	}

	s.registerSignalListeners(ctx)

	return s, nil
}

// Server manages connections with clients.
type Server struct {
	cfg *config.Config

	activeConnections   map[uint32]*conn
	activeConnectionsMx sync.RWMutex

	records     []metrics.Record
	recordsMx   sync.Mutex
	recordsChan chan []metrics.Record

	memoryMonitor memory.Monitor

	cancelCtx func()
	closed    int32

	logger log.Logger

	processName string

	fileBasedSchemaMx sync.RWMutex
	fileBasedSchema   *schema.Schema
	sampler           *sample.Sampler

	sessionProvider *mongodb.SessionProvider
	variables       *variable.Container

	// startupInfo holds configuration information
	// for the server
	startupInfo []string

	listeners []net.Listener
}

// Alter applies the provided alterations to the server's current schema.
func (s *Server) Alter(ctx context.Context, alts []*schema.Alteration) (*schema.Schema, error) {
	if !s.variables.GetBool(variable.EnableTableAlterations) {
		return nil, fmt.Errorf("cannot alter schema: alterations not enabled")
	}

	s.fileBasedSchemaMx.Lock()
	if s.fileBasedSchema != nil {
		s.fileBasedSchema.AddAlterations(alts...)
		s.fileBasedSchemaMx.Unlock()
	} else {
		s.fileBasedSchemaMx.Unlock()
		err := s.sampler.Alter(ctx, alts)
		if err != nil {
			return nil, err
		}
	}

	return s.getSchema(ctx), nil
}

// Close stops the server and stops accepting connections.
func (s *Server) Close(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		return
	}

	for _, listener := range s.listeners {
		if listener != nil {
			_ = listener.Close()
		}
	}

	// interrupt any in-progress queries
	s.activeConnectionsMx.RLock()
	s.cancelCtx()

	for _, c := range s.activeConnections {
		c.close(ctx)
	}

	s.activeConnectionsMx.RUnlock()
}

func (s *Server) getSchema(ctx context.Context) *schema.Schema {
	s.fileBasedSchemaMx.RLock()
	if s.fileBasedSchema != nil {
		fileBasedSchema := s.fileBasedSchema
		s.fileBasedSchemaMx.RUnlock()
		return fileBasedSchema
	}
	s.fileBasedSchemaMx.RUnlock()

	return s.sampler.Schema(ctx)
}

// isAdminUser returns true if the passed user is the AdminUser.
func (s *Server) isAdminUser(user, source string) bool {
	return s.cfg.MongoDB.Net.Auth.Username == user &&
		s.cfg.MongoDB.Net.Auth.Source == source
}

// isProcessOwner returns true if the named user owns the process with
// the given id.
func (s *Server) isProcessOwner(user string, id uint32) (bool, error) {
	s.activeConnectionsMx.RLock()
	targetConn, ok := s.activeConnections[id]
	if !ok {
		s.activeConnectionsMx.RUnlock()
		return false, mysqlerrors.Defaultf(mysqlerrors.ErNoSuchThread, id)
	}
	s.activeConnectionsMx.RUnlock()
	return targetConn.user == user, nil
}

// Kill attempts to kill all queries running on the connection with id.
func (s *Server) Kill(ctx context.Context, requestingConnID, targetConnID uint32, scope evaluator.KillScope) error {
	if requestingConnID == targetConnID {
		return mysqlerrors.Defaultf(mysqlerrors.ErQueryInterrupted)
	}

	s.logger.Debugf(log.Admin, "kill %v requested for [conn%v]", scope, targetConnID)

	if scope == evaluator.KillQuery {
		return s.killQuery(ctx, targetConnID, requestingConnID)
	}

	return s.killConnection(ctx, targetConnID)
}

// killConnection kills a connection.
func (s *Server) killConnection(ctx context.Context, targetConnID uint32) error {
	s.activeConnectionsMx.RLock()
	targetConn, ok := s.activeConnections[targetConnID]
	if !ok {
		s.activeConnectionsMx.RUnlock()
		return mysqlerrors.Defaultf(mysqlerrors.ErNoSuchThread, targetConnID)
	}
	s.activeConnectionsMx.RUnlock()
	targetConn.close(ctx)
	return nil
}

// killQuery kills a query.
func (s *Server) killQuery(ctx context.Context, targetConnID uint32, requestingConnID uint32) error {
	s.activeConnectionsMx.RLock()
	targetConn, ok := s.activeConnections[targetConnID]
	if !ok {
		s.activeConnectionsMx.RUnlock()
		return mysqlerrors.Defaultf(mysqlerrors.ErNoSuchThread, targetConnID)
	}

	requestingConn, ok := s.activeConnections[requestingConnID]
	if !ok {
		s.activeConnectionsMx.RUnlock()
		return mysqlerrors.Defaultf(mysqlerrors.ErNoSuchThread, requestingConnID)
	}
	s.activeConnectionsMx.RUnlock()

	// If KillOps fails in killing any operation for the client addresses, we still cancel the
	// target connection's context. This is because the alternative of having the target query
	// running without cancelling the context will prevent the user on the target connection from
	// issuing subsequent queries until the current query is completed. It is preferable to allow
	// subsequent queries to be issued, with the possiblity of no connections being available in the
	// target connection's session to execute them, than to not allow any queries to be accepted.
	clientAddresses := targetConn.session.GetClientAddresses()

	// Cancel the query's context before doing KillOps for testing purposes to prevent receiving a
	// QueryPlanKilled error from MongoDB.
	targetConn.cancelQueryCtx()
	return requestingConn.session.KillOps(ctx, clientAddresses)
}

// Resample forces a sample refresh.
func (s *Server) Resample(ctx context.Context) (*schema.Schema, error) {
	s.fileBasedSchemaMx.RLock()
	if s.fileBasedSchema != nil {
		s.fileBasedSchemaMx.RUnlock()
		return nil, fmt.Errorf("sampling is disabled; schema was loaded from a file")
	}
	s.fileBasedSchemaMx.RUnlock()

	err := s.sampler.Refresh(ctx)
	if err != nil {
		return nil, err
	}

	return s.getSchema(ctx), nil
}

// loadMongoDBInfo loads MongoDB-specific information for the server.
func (s *Server) loadMongoDBInfo(ctx context.Context) error {
	adminSession, err := s.sessionProvider.AuthenticatedAdminSession(ctx)
	if err != nil {
		return fmt.Errorf("failed to create admin session for loading server cluster information: %v", err)
	}

	i := &mongodb.Info{}
	if err = i.LoadVersionInfo(ctx, adminSession); err != nil {
		return fmt.Errorf("failed to load server version information: %v", err)
	}

	if err = i.LoadMongosInfo(ctx, adminSession); err != nil {
		return fmt.Errorf("failed to load server topology information: %v", err)

	}

	topology := "standalone"
	if i.IsMongos() {
		topology = "mongos"
	}

	s.variables.SetSystemVariable(variable.MongoDBGitVersion, i.GitVersion)
	s.variables.SetSystemVariable(variable.MongoDBTopology, topology)
	s.variables.SetSystemVariable(variable.MongoDBVersion, i.Version)
	return nil
}

// Run starts the server and begins accepting connections.
func (s *Server) Run(ctx context.Context) {
	// seed the global random number generator for calls to RAND with no seed argument.
	rand.Seed(time.Now().UnixNano())
	var once sync.Once

	listenAndServe := func(listener net.Listener) {
		once.Do(func() {
			if err := s.loadMongoDBInfo(ctx); err != nil {
				s.logger.Errf(log.Always, "unable to load MongoDB information: %v", err)
			}
		})

		for atomic.LoadInt32(&s.closed) == 0 {
			conn, err := listener.Accept()
			if err != nil {
				if atomic.LoadInt32(&s.closed) == 0 {
					s.logger.Errf(log.Always, "unable to accept connection: %v", err)
				}
				continue
			}

			procutil.PanicSafeGo(func() {
				s.serveConnection(ctx, conn)
			}, func(err interface{}) {
				s.logger.Errf(log.Always, "unable to serve new connection: %v", err)
			})
		}
	}

	// Asynchronously load the schema from MongoDB by sampling if needed.
	if s.fileBasedSchema == nil {
		s.sampler = sample.NewSampler(
			&s.cfg.Schema.Sample,
			s.processName,
			s.sessionProvider,
			s.variables,
		)
		procutil.PanicSafeGo(func() {
			s.sampler.Run(ctx)
		}, func(err interface{}) {
			s.logger.Fatalf(log.Always, "error sampling schema: %v", err)
			s.Close(ctx)
		})
	}

	// Asynchronously send metrics records to the current metrics backend.
	procutil.PanicSafeGo(func() {
		s.runTracker(ctx)
	}, func(err interface{}) {
		s.logger.Errf(log.Admin, "unexpected error tracking metrics: %v", err)
	})

	// start new goroutine for each listener
	for _, listener := range s.listeners {
		newListener := listener
		s.logger.Infof(log.Always, "waiting for connections at %v", newListener.Addr())
		procutil.PanicSafeGo(func() {
			listenAndServe(newListener)
		}, func(err interface{}) {
			s.logger.Errf(log.Always, "listen and serve error: %v", err)
		})
	}

	// Wait for all of the active client connections to return cleanly before terminating.
	<-ctx.Done()
}

// RotateLogs rotates the log file.
func (s *Server) RotateLogs() error {
	s.logger.Infof(log.Always, "log rotation initiated")
	log.Flush()
	archive, err := log.Rotate()
	if err != nil {
		return err
	}
	if archive == "" {
		s.logger.Infof(log.Always, "rotated logs using 'reopen' strategy")
	} else {
		s.logger.Infof(log.Always, "rotated logs; old log file at %s", archive)
		for _, info := range s.startupInfo {
			s.logger.Infof(log.Always, info)
		}
	}
	return nil
}

// StoreStartupInfo stores startup information in order
// to log after a log rotation.
func (s *Server) StoreStartupInfo(startupInfo []string) {
	s.startupInfo = startupInfo
}

func (s *Server) addConnection(c *conn) int {
	s.activeConnectionsMx.Lock()
	s.activeConnections[c.connectionID] = c
	activeConnections := len(s.activeConnections)
	s.activeConnectionsMx.Unlock()
	atomic.StoreUint32(s.variables.ThreadsConnected, uint32(activeConnections))
	pluralized := strutil.Pluralize(activeConnections, "connection", "connections")
	address := c.conn.RemoteAddr().String()
	if address == "" {
		address = c.conn.LocalAddr().String()
	}
	c.process.SetHost(c.getFormattedAddress())
	c.logger.Infof(log.Always, "connection accepted from %v #%v (%v %v now open)", address,
		c.connectionID, activeConnections, pluralized)

	// current number of active connections
	return activeConnections
}

func (s *Server) removeConnection(c *conn) {
	s.activeConnectionsMx.Lock()
	delete(s.activeConnections, c.connectionID)
	if atomic.LoadInt32(&s.closed) == 1 && len(s.activeConnections) == 0 {
		s.cancelCtx()
	}
	activeConnections := len(s.activeConnections)
	s.activeConnectionsMx.Unlock()
	atomic.StoreUint32(s.variables.ThreadsConnected, uint32(activeConnections))
	pluralized := strutil.Pluralize(activeConnections, "connection", "connections")
	address := c.conn.RemoteAddr().String()
	if address == "" {
		address = c.conn.LocalAddr().String()
	}
	c.logger.Infof(log.Always, "end connection %v (%v %v now open)", address, activeConnections,
		pluralized)
}

func (s *Server) serveConnection(ctx context.Context, c net.Conn) {
	connCtx, cancelConnCtx := context.WithCancel(ctx)
	conn, err := newConn(connCtx, cancelConnCtx, s, c)
	if err != nil {
		address := c.RemoteAddr().String()
		if address == "" {
			address = c.LocalAddr().String()
		}

		s.logger.Errf(log.Always, "connection accepted "+
			"from %v, but could not initialize: %v", address, err)
		_ = c.Close()
		return
	}

	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			conn.logger.Errf(log.Dev, "error serving connection: %v, %s", err, buf)
		}
		conn.close(connCtx)
	}()

	activeConnections := int64(s.addConnection(conn))
	maxConnections := s.variables.GetInt64(variable.MaxConnections)

	if activeConnections > maxConnections && maxConnections > 0 {
		conn.logger.Errf(
			log.Always, mysqlerrors.Defaultf(mysqlerrors.ErConCountError).Message)
		conn.close(connCtx)
		return
	}

	atomic.StoreInt32(&conn.queryRunning, 1)
	if err := conn.handshake(ctx); err != nil {
		conn.logger.Errf(log.Always, "handshake error: %v", err)
		atomic.StoreInt32(&conn.queryRunning, 0)
		return
	}
	atomic.StoreInt32(&conn.queryRunning, 0)

	conn.run(connCtx)
}
