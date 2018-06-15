package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/memory"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/ssl"
	"github.com/10gen/sqlproxy/variable"
)

var (
	errBadConn       = mysqlerrors.Unknownf("connection was bad")
	errMalformPacket = mysqlerrors.Defaultf(mysqlerrors.ErMalformedPacket)
)

type flushWriter interface {
	Flush() error
	Write(p []byte) (int, error)
}

type conn struct {
	server  *Server
	session *mongodb.Session
	logger  log.Logger
	startDB string

	// synchronization variables for
	// terminating connection
	closer       *sync.Cond
	closed       int32
	queryRunning int32

	// storage for thread information
	process *Process

	// synchronization variables for
	// terminating a query
	ctx     context.Context
	cancel  context.CancelFunc
	ctxLock *sync.RWMutex

	conn                net.Conn
	reader              io.Reader
	writer              io.Writer
	sequence            uint8
	compressionSequence uint8
	compressionOn       bool

	memoryMonitor *memory.Monitor

	capability   uint32
	connectionID uint32
	user         string
	source       string
	currentDB    *catalog.Database
	lastInsertID int64
	affectedRows int64

	authPluginName                string
	authPluginData                []byte
	clientRequestedAuthPluginName string
	clientAuthResponse            []byte
	clientConnectAttributes       []clientConnectionAttribute

	stmts map[uint32]*stmt

	// status variables
	bytesReceived uint64
	bytesSent     uint64

	catalog   *catalog.Catalog
	variables *variable.Container
}

type clientConnectionAttribute struct {
	key   string
	value string
}

func newConn(s *Server, c net.Conn) (*conn, error) {
	memoryMonitor, err := s.memoryMonitor.CreateChild(
		"Connection",
		s.cfg.Runtime.Memory.MaxPerConnection)
	if err != nil {
		return nil, err
	}

	session, err := s.sessionProvider.Session(s.lifetimeCtx)

	connID := atomic.AddUint32(s.variables.Connections, 1)

	ctx, cancel := context.WithCancel(s.lifetimeCtx)

	newConn := &conn{
		server:        s,
		session:       session,
		ctx:           ctx,
		cancel:        cancel,
		ctxLock:       &sync.RWMutex{},
		conn:          c,
		reader:        c,
		writer:        c,
		closer:        sync.NewCond(&sync.Mutex{}),
		closed:        0,
		bytesReceived: uint64(0),
		bytesSent:     uint64(0),
		queryRunning:  0,
		connectionID:  connID,
		capability: ClientProtocol41 |
			ClientConnectWithDB |
			ClientLongFlag |
			ClientLongPassword |
			ClientSecureConnection |
			ClientCompress |
			ClientConnectAttrs,
		stmts:         make(map[uint32]*stmt),
		variables:     variable.NewSessionContainer(s.variables),
		process:       NewProcess(connID),
		memoryMonitor: memoryMonitor,
	}

	if err != nil {
		newConn.writeError(mysqlerrors.Defaultf(mysqlerrors.ErConnectToForeignDataSource,
			"MongoDB"))
		return nil, fmt.Errorf("unable to connect to MongoDB: %v", err)
	}

	newConn.variables.AllocatedMemory = memoryMonitor.Allocated

	if s.cfg.Security.Enabled {
		newConn.capability = newConn.capability |
			ClientPluginAuth |
			ClientPluginAuthLenencClientData
		newConn.authPluginName = mongosqlAuthClientAuthPluginName
		newConn.authPluginData = []byte{1, 0} // version 1.0 of the mongosql_auth plugin
	} else {
		buf, err := randomBuf(20)
		if err != nil {
			newConn.writeError(mysqlerrors.Defaultf(mysqlerrors.ErUnknownError))
			return nil, fmt.Errorf("unable to generate salt: %v", err)
		}
		newConn.authPluginName = nativePasswordPluginName
		newConn.authPluginData = buf
	}

	newConn.logger = newConn.Logger(log.NetworkComponent)

	if s.cfg.Net.SSL.Mode != "disabled" {
		newConn.capability |= ClientSSL
	}

	return newConn, nil
}

func (c *conn) close() {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		return
	}

	c.cancel()

	// Kill running queries for this connection and ignore any errors.
	// Always do this because queryRunning can get unset while a db operation is running.
	s := c.session
	_ = s.KillOps(s.GetClientAddresses())

	// this establishes a deadline by which we'll forcefully
	// terminate the client connection to ensure we can
	// cleanly terminate the server when we're blocked on a
	// client read/write.
	util.PanicSafeGo(func() {
		timer := time.NewTimer(1 * time.Second)
		<-timer.C
		timer.Stop()
		atomic.StoreInt32(&c.queryRunning, 0)
		c.closer.Signal()
	}, func(err interface{}) {
		c.logger.Errf(log.Dev, "connection close error: %v", err)
	})

	// wait for any running queries to be interrupted
	c.closer.L.Lock()
	for atomic.LoadInt32(&c.queryRunning) != 0 {
		c.closer.Wait()
	}
	c.closer.L.Unlock()

	_ = c.session.Close()
	_ = c.conn.Close()

	util.PanicSafeGo(func() {
		c.server.removeConnection(c)
	}, func(interface{}) {})

	err := c.memoryMonitor.Release(c.memoryMonitor.Allocated())
	if err != nil {
		c.logger.Errf(log.Dev, "memory release error", err)
	}
}

// Catalog returns the catalog.
func (c *conn) Catalog() *catalog.Catalog {
	return c.catalog
}

func (c *conn) MemoryMonitor() *memory.Monitor {
	return c.memoryMonitor
}

// UpdateCatalog updates the catalog to utilize the new schema.
func (c *conn) UpdateCatalog(s *schema.Schema) error {
	err := c.loadMongoDBInfo(s)
	if err != nil {
		return err
	}

	return c.setCatalogFromSchema(s)
}

func (c *conn) setCatalogFromSchema(s *schema.Schema) error {
	cat, err := catalog.Build(s, c.variables)
	if err != nil {
		return err
	}

	infoSchema, err := cat.Database(catalog.InformationSchemaDatabase)
	if err != nil {
		return err
	}

	// also add the PROCESSLIST table to the catalog
	err = c.UpdateWithProcessListTable(infoSchema)
	if err != nil {
		return err
	}

	c.catalog = cat
	return nil
}

// ConnectionId returns the connection's identifier.
func (c *conn) ConnectionID() uint32 {
	return c.connectionID
}

// Context returns the connection's context.
func (c *conn) Context() context.Context {
	var ctx context.Context
	c.ctxLock.RLock()
	ctx = c.ctx
	c.ctxLock.RUnlock()
	return ctx
}

// DB returns the current database name.
func (c *conn) DB() string {
	if c.currentDB == nil {
		return ""
	}
	return string(c.currentDB.Name)
}

// Server returns access to the server
func (c *conn) Server() evaluator.ServerCtx {
	return c.server
}

func (c *conn) dispatch(data []byte) (err error) {
	if len(data) < 1 {
		return mysqlerrors.Defaultf(mysqlerrors.ErUnknownComError)
	}

	cmd := data[0]
	data = data[1:]

	if cmd != ComPing && cmd != ComStatistics {
		atomic.AddUint64(c.server.variables.Queries, 1)
	}

	switch cmd {
	case ComQuit:
		atomic.StoreInt32(&c.queryRunning, 0)
		c.close()
		return nil
	case ComQuery:
		s := util.String(c.variables.GetCharset(variable.CharacterSetClient).Decode(data))
		c.process.UpdateProcess(CommandQuery, s)
		err = c.handleQuery(s)
		c.process.UpdateProcess(CommandSleep, "")
		return err
	case ComPing:
		return c.writeOK(nil)
	case ComInitDB:
		s := util.String(c.variables.GetCharset(variable.CharacterSetClient).Decode(data))
		if err := c.useDB(s); err != nil {
			return err
		}
		return c.writeOK(nil)
	case ComFieldList:
		return c.handleFieldList(data)
	case ComStmtPrepare:
		s := util.String(c.variables.GetCharset(variable.CharacterSetClient).Decode(data))
		return c.handleStmtPrepare(s)
	case ComStmtExecute:
		return c.handleStmtExecute(data)
	case ComStmtClose:
		return c.handleStmtClose(data)
	case ComStmtSendLongData:
		return c.handleStmtSendLongData(data)
	case ComStmtReset:
		return c.handleStmtReset(data)
	default:
		return mysqlerrors.Defaultf(mysqlerrors.ErUnknownComError)
	}
}

func (c *conn) handshake() error {
	c.logger.Infof(log.Dev, "writing initial handshake")

	if err := c.writeInitialHandshake(); err != nil {
		err = mysqlerrors.Newf(mysqlerrors.ErHandshakeError, "send initial handshake error: %v",
			err)
		c.writeError(err)
		return err
	}

	c.logger.Infof(log.Dev, "reading handshake response")
	if err := c.readHandshakeResponse(); err != nil {
		err = mysqlerrors.Newf(mysqlerrors.ErHandshakeError, "recv handshake response error: %v",
			err)
		c.writeError(err)
		return err
	}

	currentSchema := c.server.getSchema()
	if currentSchema == nil {
		err := mysqlerrors.Newf(mysqlerrors.ErHandshakeError, "MongoDB schema not yet available")
		c.writeError(err)
		return err
	}

	var err error
	if c.server.cfg.Security.Enabled {
		c.logger.Infof(log.Dev, "configuring client authentication for principal %s", c.user)
		switch c.clientRequestedAuthPluginName {
		case mongosqlAuthClientAuthPluginName:
			err = c.authMongoSQLAuthPlugin()
		default:
			err = c.authClearTextPasswordPlugin()
		}

		if err != nil {
			c.writeError(mysqlerrors.Newf(mysqlerrors.ErAccessDeniedError, "Access denied for "+
				"user '%s'", c.user))
			return err
		}

		c.logger.Infof(log.Dev, "successfully authenticated as principal %s on %s",
			c.user, c.source)
	} else {
		if c.user != "" {
			if c.user != "ODBC" && c.capability&ClientODBC == 0 {
				c.logger.Warnf(log.Dev,
					"ignoring provided credentials for '%v'; authentication is "+
						"not enabled", c.user)
			}
			c.user = ""
		}
		// If the server is running without security, but the client requests a
		// auth mechanism other than nativePasswordPlugin, we need to switch to
		// nativePasswordPlugin. Unfortunately, when security is disabled,
		// however, the client will not tell us what plugin they are using. So
		// the only safe thing to do is unconditionally request a switch to
		// mysql native password. If we do not do this, clients using our ODBC
		// plugin (which requests using mongosql-auth by default) will be
		// unable to connect to local BIC running without security.
		// See
		// https://dev.mysql.com/doc/internals/en/connection-phase-packets.html
		// about how connection phase packets are structured.
		// See
		// https://bit.ly/2qRd1A5
		// For more information on when client plugins are explicitly
		// requested.
		nullTerminatedAuthPluginData := append([]byte{}, c.authPluginData...)
		nullTerminatedAuthPluginData = append(nullTerminatedAuthPluginData, 0)
		err = c.writeAuthSwitchRequest(nativePasswordPluginName,
			nullTerminatedAuthPluginData)
		if err != nil {
			err = mysqlerrors.Newf(mysqlerrors.ErHandshakeError,
				err.Error())
			return err
		}
		if err = c.readAuthSwitchResponse(); err != nil {
			err = mysqlerrors.Newf(mysqlerrors.ErHandshakeError,
				err.Error())
			return err
		}
	}
	c.process.SetUser(c.user)

	if err = c.loadMongoDBInfo(currentSchema); err != nil {
		err = mysqlerrors.Newf(mysqlerrors.ErHandshakeError, err.Error())
		c.writeError(err)
		return err
	}

	c.logger.Infof(log.Admin, "connected to MongoDB %v, git version: %v",
		c.variables.MongoDBInfo.Version, c.variables.MongoDBInfo.GitVersion)

	if c.variables.MongoDBInfo.CompatibleVersion != "" {
		c.logger.Infof(log.Admin, "MongoDB version compatibility is %v",
			c.variables.MongoDBInfo.CompatibleVersion)
	}

	if !c.variables.MongoDBInfo.VersionAtLeast(3, 2) {
		err = mysqlerrors.Newf(mysqlerrors.ErHandshakeError,
			"MongoDB version is %v but version >= 3.2 required",
			c.variables.MongoDBInfo.Version)
		c.writeError(err)
		return err
	}
	c.setStatusVariables()

	err = c.setCatalogFromSchema(currentSchema)
	if err != nil {
		err = mysqlerrors.Newf(mysqlerrors.ErHandshakeError, "error building catalog: %v", err)
		c.writeError(err)
		return err
	}

	if c.startDB != "" {
		if err = c.useDB(c.startDB); err != nil {
			err = mysqlerrors.Newf(mysqlerrors.ErHandshakeError, "error using database %v: %v",
				c.startDB, err)
			c.writeError(err)
			return err
		}
	}

	if err = c.writeOK(nil); err != nil {
		return mysqlerrors.Newf(mysqlerrors.ErHandshakeError, "write ok: %v", err)
	}

	c.sequence = 0
	c.compressionSequence = 0

	// set up compression reader and writer
	if c.capability&ClientCompress != 0 {
		c.reader = NewCompressedReader(c.reader, c)
		c.writer = NewCompressedWriter(c.writer, c)
		c.compressionOn = true
	}

	// using same buffer size as mysql
	// definition: https://goo.gl/qgMv3x
	// usage: https://github.com/mysql/mysql-server/blob/5.7/sql/net_serv.cc#L434
	c.writer = bufio.NewWriterSize(c.writer, maxPayloadLength+4)

	return nil
}

// LastInsertId returns the last insert id.
func (c *conn) LastInsertId() int64 {
	return c.lastInsertID
}

// VersionAtLeast is a function comparing the MongoDB
// server version for which we are translating so that
// we know which aggregation language features are present.
func (c *conn) VersionAtLeast(version ...uint8) bool {
	return c.Variables().MongoDBInfo.VersionAtLeast(version...)
}

// Logger returns a logger sufficient
// for reporting errors in translation.
func (c *conn) Logger(componentStr ...string) log.Logger {
	globalLogger := log.GlobalLogger()
	var component string
	if len(componentStr) == 0 || len(componentStr) > 0 && componentStr[0] == "" {
		component = fmt.Sprintf("%-10v [conn%v]", globalLogger.GetComponent(), c.connectionID)
	} else {
		component = fmt.Sprintf("%-10v [conn%v]", componentStr[0], c.connectionID)
	}
	return log.NewComponentLogger(component, globalLogger)
}

func (c *conn) readHandshakeResponse() error {

	data, err := c.readPacket()
	if err != nil {
		return err
	}

	var pos int

	readHeader := func() {
		pos = 0

		c.capability &= binary.LittleEndian.Uint32(data[:4])

		pos += 4

		// in the case of SSL, some clients won't send anything else until SSL is negotiated.
		if len(data) > 4 {
			// skip max packet size
			pos += 4

			// charset
			var col *collation.Collation
			col, err = collation.GetByID(collation.ID(data[pos]))
			pos++

			if err == nil {
				names := []variable.Name{
					variable.CharacterSetClient,
					variable.CharacterSetConnection,
					variable.CharacterSetResults,
				}

				for _, name := range names {
					err = c.variables.Set(name, variable.SessionScope, variable.SystemKind,
						string(col.CharsetName))
					if err != nil {
						break
					}
				}
			}
			if err != nil {
				c.logger.Warnf(log.Dev, "failed to set collation: %v", err)
			}

			// skip reserved 23[00]
			pos += 23
		}
	}

	readHeader()

	clientSSL := c.capability&ClientSSL != 0
	switch c.server.cfg.Net.SSL.Mode {
	case "disabled":
		// return an error if client is using SSL
		if clientSSL {
			// we shouldn't ever actually reach this, because the
			// connection should fail during capability negotiation
			return fmt.Errorf("SSL handshake received but server is started without SSL")
		}
	case "allowSSL":
		// negotiate SSL if the client is using it
		// otherwise, proceed without ssl
	case "requireSSL":
		// return an error if client not using SSL
		if !clientSSL {
			return fmt.Errorf("This server is configured to only allow SSL connections")
		}
	}

	if clientSSL {
		c.logger.Infof(log.Dev, "negotiating ssl")
		if err = c.useTLS(); err != nil {
			err = mysqlerrors.Newf(mysqlerrors.ErHandshakeError, "ssl configuration error: %v", err)
			c.writeError(err)
			return err
		}
		c.logger.Infof(log.Dev, "ssl connection established")

		data, err = c.readPacket()
		if err != nil {
			err = mysqlerrors.Newf(mysqlerrors.ErHandshakeError, "continuation after successful "+
				"ssl negotiation failed: %v", err)
			c.writeError(err)
			return err
		}

		// We need to read the handshake response header again because, now that we have TLS, the
		// client resends its handshake response packet.
		readHeader()
	}

	// user name string[NUL]
	userBytes := data[pos : pos+bytes.IndexByte(data[pos:], 0)]
	pos += len(userBytes) + 1
	c.user = util.String(c.variables.GetCharset(variable.CharacterSetClient).Decode(userBytes))

	// auth response string[NUL]
	if (c.capability & ClientPluginAuthLenencClientData) != 0 {
		authLen, _, count := lengthEncodedInt(data[pos:])
		pos += count
		c.clientAuthResponse = data[pos : pos+int(authLen)]
		pos += int(authLen)
	} else if (c.capability & ClientSecureConnection) != 0 {
		authLen := int(data[pos])
		pos++
		c.clientAuthResponse = data[pos : pos+authLen]
		pos += authLen
	} else {
		c.clientAuthResponse = data[pos : pos+bytes.IndexByte(data[pos:], 0)]
		pos += len(c.clientAuthResponse) + 1
	}

	if (c.capability & ClientInteractive) != 0 {
		c.variables.SetSystemVariable(variable.WaitTimeoutSecs,
			c.variables.GetInt64(variable.InteractiveTimeoutSecs))
	}

	if pos == len(data) {
		return nil
	}

	if (c.capability & ClientConnectWithDB) != 0 {
		dbBytes := data[pos : pos+bytes.IndexByte(data[pos:], 0)]
		pos += len(dbBytes) + 1

		db := util.String(c.variables.GetCharset(variable.CharacterSetClient).Decode(dbBytes))
		c.startDB = db
	}

	if pos == len(data) {
		return nil
	}

	if (c.capability & ClientPluginAuth) != 0 {
		// The Java Driver has a bug where it sends an extra nul byte when
		// there is no initial db. So, if we don't have ClientConnectWithDB
		// and the current byte is nul, we'll skip it.
		// REF: https://bugs.mysql.com/bug.php?id=79612
		if c.capability&ClientConnectWithDB == 0 && data[pos] == 0 {
			pos++
			if pos == len(data) {
				return nil
			}
		}

		clientPluginNameBytes := data[pos : pos+bytes.IndexByte(data[pos:], 0)]
		pos += len(clientPluginNameBytes) + 1

		// this is always a utf8 string
		c.clientRequestedAuthPluginName = util.String(clientPluginNameBytes)
	}

	// MySQL and the Java SQL driver (and possibly other clients) only set
	// ClientConnectAttrs when authentication is used.
	if (c.capability & ClientConnectAttrs) != 0 {

		l, _, count := lengthEncodedInt(data[pos:])

		attrsLen := int(l)

		pos += count
		attrPos := 0

		attrs := make([]clientConnectionAttribute, 0)
		logString := ""

		for attrPos < attrsLen {
			keyBytes, _, keyLength, err := lengthEncodedString(data[pos+attrPos:])
			if err != nil {
				c.logger.Infof(log.Admin, "error parsing connection attribute key at index %d: %v",
					len(attrs), err)
				return fmt.Errorf("invalid connection attribute at index %v: %v", len(attrs), err)
			}
			attrPos += keyLength

			key := util.String(c.variables.GetCharset(variable.CharacterSetClient).Decode(keyBytes))

			valBytes, _, valLength, err := lengthEncodedString(data[pos+attrPos:])
			if err != nil {
				c.logger.Infof(log.Admin, "error parsing connection attribute value (key %s): %v",
					key, err)
				return fmt.Errorf("invalid connection attribute for key %s: %v", key, err)
			}
			attrPos += valLength

			val := util.String(c.variables.GetCharset(variable.CharacterSetClient).Decode(valBytes))

			attrs = append(attrs, clientConnectionAttribute{key, val})
			logString += fmt.Sprintf("%s:%s, ", key, val)
		}

		pos += attrPos
		c.clientConnectAttributes = attrs
		if len(attrs) > 0 {
			c.logger.Infof(log.Admin, "client provided connection attributes %s",
				logString[:len(logString)-2])
		}

	}
	return nil
}

func (c *conn) readPacket() ([]byte, error) {
	header := []byte{0, 0, 0, 0}

	if _, err := io.ReadFull(c.reader, header); err != nil {
		return nil, err
	}

	length := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)

	// packets from the client should not be larger than the maxPayload length
	// when going over the wire
	if length > maxPayloadLength {
		// this check occurs in uncompressPacket for compressed packets, so
		// check only needed for vanilla packets
		if !c.compressionOn {
			c.writeError(mysqlerrors.Defaultf(mysqlerrors.ErNetPacketTooLarge))
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErNetPacketTooLarge)
		}
	}

	sequence := header[3]
	if sequence != c.sequence {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErNetPacketsOutOfOrder)
	}

	c.sequence++

	data := make([]byte, length)
	if _, err := io.ReadFull(c.reader, data); err != nil {
		c.logger.Errf(log.Dev, "read full error: %v", err)
		return nil, errBadConn
	}

	// in uncompressed world, record bytes received
	if !c.compressionOn {
		bytesReceived := uint64(length) + 4 // for the header
		atomic.AddUint64(&c.bytesReceived, bytesReceived)
		atomic.AddUint64(c.server.variables.BytesReceived, bytesReceived)
	}
	return data, nil
}

// RowCount returns the number of rows affected by the last statement.
func (c *conn) RowCount() int64 {
	return c.affectedRows
}

// refreshContext creates a new context for this connection.
func (c *conn) refreshContext() {
	// need to lock context when writing because other threads might be trying to read
	// the connection's context at the same time.
	c.ctxLock.Lock()
	c.ctx, c.cancel = context.WithCancel(c.server.lifetimeCtx)
	c.ctxLock.Unlock()
}

func (c *conn) run() {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			c.logger.Errf(log.Dev, "error serving connection: %v\n%s\n", err, buf)
		}
		c.close()
	}()

	type packetRead struct {
		data []byte
		err  error
	}

	packetReadChan := make(chan packetRead)
	var pkt packetRead

	for {
		util.PanicSafeGo(func() {
			data, err := c.readPacket()
			packetReadChan <- packetRead{data, err}
			atomic.StoreInt32(&c.queryRunning, 1)
		}, func(err interface{}) {
			c.logger.Errf(log.Dev, "packet read error: %v", err)
		})

		waitTimeout := time.Duration(c.variables.GetInt64(variable.WaitTimeoutSecs)) * time.Second
		timer := time.NewTimer(waitTimeout)
		var timeoutTime time.Time

		select {
		case timeoutTime = <-timer.C:
			c.logger.Warnf(log.Admin, "client wait time out after %v", waitTimeout.String())
		case pkt = <-packetReadChan:
			if pkt.err != nil && atomic.LoadInt32(&c.closed) != 1 {
				c.logger.Errf(log.Always, "client read error: %v", pkt.err)
			}
		}

		timer.Stop()

		if !timeoutTime.IsZero() || pkt.err != nil {
			atomic.StoreInt32(&c.queryRunning, 0)
			c.closer.Signal()
			return
		}

		if err := c.dispatch(pkt.data); err != nil {
			// Only cancel a context when a query interruption is requested.
			if err == context.Canceled {
				err = mysqlerrors.Defaultf(mysqlerrors.ErQueryInterrupted)
			}
			c.logger.Errf(log.Admin, "dispatch error: %v", err)
			if err != errBadConn {
				c.writeError(err)
			}
		}

		atomic.StoreInt32(&c.queryRunning, 0)
		c.closer.Signal()
		// reset both packet sequences
		c.sequence = 0
		c.compressionSequence = 0
	}
}

// Session returns a new mongodb.Session connected to MongoDB.
func (c *conn) Session() (session *mongodb.Session) {
	return c.session
}

// setStatusVariables sets status variables for this client connection.
func (c *conn) setStatusVariables() {
	sessionVariables, globalVariables := c.variables, c.server.variables

	sessionVariables.BytesReceived = &c.bytesReceived
	sessionVariables.BytesSent = &c.bytesSent
	sessionVariables.Connections = globalVariables.Connections
	sessionVariables.Queries = globalVariables.Queries
	sessionVariables.ThreadsConnected = globalVariables.ThreadsConnected
	sessionVariables.StartTime = globalVariables.StartTime
}

// loadMongoDBInfo sets system variables that store information about MongoDB
func (c *conn) loadMongoDBInfo(currentSchema *schema.Schema) (err error) {
	c.variables.MongoDBInfo, err = mongodb.LoadInfo(
		c.logger,
		c.server.sessionProvider,
		c.session,
		currentSchema,
		c.server.cfg,
	)
	if err != nil {
		return fmt.Errorf("error retrieving information from MongoDB: %v", err)
	}

	err = c.variables.MongoDBInfo.SetCompatibleVersion(
		c.server.variables.GetString(variable.MongoDBVersionCompatibility))
	if err != nil {
		return fmt.Errorf("error setting compatibility version: %v", err)
	}

	return nil
}

func (c *conn) status() uint16 {
	if c.variables.GetBool(variable.Autocommit) {
		return ServerSatusAutocommit
	}

	return 0
}

func (c *conn) useDB(db string) error {
	d, err := c.catalog.Database(db)
	if err != nil {
		return err
	}

	c.currentDB = d
	c.variables.SetSystemVariable(variable.CollationDatabase,
		string(c.variables.GetCollation(variable.CollationServer).Name))
	c.process.SetDB(string(d.Name))
	return nil
}

// RemoteHost returns the hostname for the current connection.
func (c *conn) RemoteHost() string {
	host, _, err := net.SplitHostPort(c.conn.RemoteAddr().String())
	if err != nil {
		host = c.conn.RemoteAddr().String()
		// For socket connections which have neither host nor port
		if host == "" {
			host = "localhost"
		}
	}
	return host
}

// User returns the current user's name.
func (c *conn) User() string {
	return c.user
}

// AuthenticationDatabase returns the current user's source database name.
func (c *conn) AuthenticationDatabase() string {
	return c.source
}

func (c *conn) getFormattedAddress() string {
	hasPort := true
	addr := c.conn.RemoteAddr().String()
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
		hasPort = false
	}
	if host == "127.0.0.1" || host == "" {
		host = "localhost"
	}
	// has port
	if hasPort {
		return host + ":" + port
	}

	return host
}

func (c *conn) useTLS() error {
	tlsc, err := ssl.Handshake(c.conn, c.server.cfg)
	if err != nil {
		return err
	}

	c.conn = tlsc
	c.reader = c.conn
	c.writer = c.conn

	return nil
}

func (c *conn) Variables() *variable.Container {
	return c.variables
}

func (c *conn) writeEOF(status uint16) error {
	data := make([]byte, 4, 9)

	data = append(data, EOFHeader)
	if c.capability&ClientProtocol41 > 0 {
		data = append(data, 0, 0)
		data = append(data, byte(status), byte(status>>8))
	}
	return c.writePacketandFlush(data)
}

func (c *conn) writeError(e error) {
	var m *mysqlerrors.MySQLError
	var ok bool
	if m, ok = e.(*mysqlerrors.MySQLError); !ok {
		m = mysqlerrors.Unknownf(e.Error())
	}

	data := make([]byte, 4, 16+len(m.Message))

	data = append(data, ErrHeader)
	data = append(data, byte(m.Code), byte(m.Code>>8))

	if c.capability&ClientProtocol41 > 0 && c.bytesSent != 0 {
		data = append(data, '#')
		data = append(data, m.State...)
	}

	data = append(data, m.Message...)
	_ = c.writePacketandFlush(data)
}

func (c *conn) writeInitialHandshake() error {
	data := make([]byte, 4, 128)

	// min version 10
	data = append(data, minProtocolVersion)

	// server version[00]
	version := c.server.variables.GetString(variable.Version)
	versionComment := c.server.variables.GetString(variable.VersionComment)

	data = append(data, fmt.Sprintf("%s %s", version, versionComment)...)
	data = append(data, 0)

	// connection id
	data = append(data,
		byte(c.connectionID),
		byte(c.connectionID>>8),
		byte(c.connectionID>>16),
		byte(c.connectionID>>24))

	// auth-plugin-data-part-1
	count := 0
	for count < 8 && count < len(c.authPluginData) {
		data = append(data, c.authPluginData[count])
		count++
	}
	// must have at least 8 bytes
	for count < 8 {
		data = append(data, 0)
		count++
	}

	// filler [00]
	data = append(data, 0)

	// capability flag lower 2 bytes
	data = append(data, byte(c.capability), byte(c.capability>>8))

	// charset
	data = append(data, byte(collation.Default.ID))

	// status
	status := c.status()
	data = append(data, byte(status), byte(status>>8))

	// below 13 byte may not be used
	// capability flag upper 2 bytes
	data = append(data, byte(c.capability>>16), byte(c.capability>>24))

	if (c.capability & ClientPluginAuth) != 0 {
		max := len(c.authPluginData)
		if max > 0 && max < 21 {
			max = 21
		}
		data = append(data, byte(max))
	} else {
		data = append(data, 0)
	}

	// reserved 10 [00]
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)

	if (c.capability & ClientSecureConnection) != 0 {
		count := 0
		if len(c.authPluginData) > 8 {
			data = append(data, c.authPluginData[8:]...)
			count = len(c.authPluginData) - 8
		}

		for count < 13 {
			data = append(data, 0)
			count++
		}
	}

	if (c.capability & ClientPluginAuth) != 0 {
		// auth-plugin-name string[NUL]
		data = append(data, []byte(c.authPluginName)...)
		data = append(data, 0)
	}

	return c.writePacketandFlush(data)
}

func (c *conn) writePacket(data []byte) error {
	// data slice is passed in with four empty bytes for the header
	length := len(data) - 4
	totalBytesSent := uint64(0)

	for length >= maxPayloadLength {

		data[0] = byte(0xff & maxPayloadLength)
		data[1] = byte(0xff & (maxPayloadLength >> 8))
		data[2] = byte(0xff & (maxPayloadLength >> 16))

		data[3] = c.sequence

		_, err := c.writer.Write(data[:4+maxPayloadLength])

		if err != nil {
			c.logger.Errf(log.Dev, "write maxPayloadLength error: %v", err)
			return errBadConn
		}
		c.sequence++
		totalBytesSent += uint64(maxPayloadLength + 4)

		length -= maxPayloadLength
		data = data[maxPayloadLength:]
	}

	// header
	data[0] = byte(0xff & length)
	data[1] = byte(0xff & (length >> 8))
	data[2] = byte(0xff & (length >> 16))

	data[3] = c.sequence

	if _, err := c.writer.Write(data); err != nil {
		c.logger.Errf(log.Dev, "write packet error: %v", err)
		return errBadConn
	}

	c.sequence++

	// in uncompressed world, record bytes sent
	if !c.compressionOn {
		totalBytesSent += uint64(len(data))

		atomic.AddUint64(&c.bytesSent, totalBytesSent)
		atomic.AddUint64(c.server.variables.BytesSent, totalBytesSent)
	}

	return nil
}

// flush writes all bytes in the buffer to the underlying writer.
// Before buffering is set up, it does nothing.
func (c *conn) flush() error {
	// if buffering exists, flush; else, do nothing
	bufWriter, ok := c.writer.(flushWriter)
	if ok {
		return bufWriter.Flush()
	}
	return nil // flush does nothing when no buffering is used
}

// writePacketandFlush adds data to the buffer and then calls flush.
// Before buffering is set up, this will simply write 'data'.
func (c *conn) writePacketandFlush(data []byte) error {
	err := c.writePacket(data)
	if err != nil {
		return err
	}

	return c.flush()
}

func (c *conn) writeOK(r *Result) error {
	if r == nil {
		r = &Result{Status: c.status()}
	}
	data := make([]byte, 4, 32)

	data = append(data, OkHeader)

	data = append(data, putLengthEncodedInt(r.AffectedRows)...)
	data = append(data, putLengthEncodedInt(r.InsertID)...)

	if c.capability&ClientProtocol41 > 0 {
		data = append(data, byte(r.Status), byte(r.Status>>8))
		data = append(data, 0, 0)
	}

	return c.writePacketandFlush(data)
}
