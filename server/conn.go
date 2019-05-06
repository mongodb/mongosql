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

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/memory"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mongodb/ssl"
	"github.com/10gen/sqlproxy/schema"
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

	cancelConnCtx  context.CancelFunc
	cancelQueryCtx context.CancelFunc

	conn                net.Conn
	reader              io.Reader
	writer              io.Writer
	sequence            uint8
	compressionSequence uint8
	compressionOn       bool

	memoryMonitor memory.Monitor

	capability   uint32
	connectionID uint32
	user         string
	source       string
	affectedRows int64

	authPluginName                string
	authPluginData                []byte
	clientRequestedAuthPluginName string
	clientAuthResponse            []byte
	clientConnectAttributes       []clientConnectionAttribute
	constrainedDelegation         bool

	stmts map[uint32]*stmt

	// status variables
	bytesReceived uint64
	bytesSent     uint64

	catalog     catalog.Catalog
	currentDB   catalog.Database
	variables   *variable.Container
	mongoDBInfo *mongodb.Info
}

type clientConnectionAttribute struct {
	key   string
	value string
}

func newConn(ctx context.Context, cancelConnCtx context.CancelFunc, s *Server, c net.Conn) (*conn, error) {
	memoryMonitor, err := s.memoryMonitor.CreateChild(
		"Connection",
		s.cfg.Runtime.Memory.MaxPerConnection)
	if err != nil {
		return nil, err
	}

	session, err := s.sessionProvider.Session(ctx)

	connID := atomic.AddUint32(s.variables.Connections, 1)

	newConn := &conn{
		server:         s,
		session:        session,
		cancelConnCtx:  cancelConnCtx,
		cancelQueryCtx: nil,
		conn:           c,
		reader:         c,
		writer:         c,
		closer:         sync.NewCond(&sync.Mutex{}),
		closed:         0,
		bytesReceived:  uint64(0),
		bytesSent:      uint64(0),
		queryRunning:   0,
		connectionID:   connID,
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
		newConn.writeError(mysqlerrors.Defaultf(mysqlerrors.ErConnectToForeignDataSource, "MongoDB"))
		return nil, fmt.Errorf("unable to connect to MongoDB: %v", err)
	}

	newConn.variables.AllocatedMemory = memoryMonitor.Allocated

	if s.cfg.Security.Enabled {
		newConn.capability = newConn.capability |
			ClientPluginAuth |
			ClientPluginAuthLenencClientData
		newConn.authPluginName = mongosqlAuthClientAuthPluginName
		newConn.authPluginData = []byte{1, 0} // version 1.0 of the mongosql_auth plugin
		newConn.constrainedDelegation = s.cfg.Security.GSSAPI.ConstrainedDelegation
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

func (c *conn) canListProcess(processUser string) bool {
	return c.server.isAdminUser(c.user, c.source) || processUser == c.user ||
		c.mongoDBInfo.IsAllowedCluster(mongodb.InprogPrivilege)
}

func (c *conn) close(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		return
	}

	c.cancelConnCtx()

	// Kill running queries for this connection and ignore any errors.
	// Always do this because queryRunning can get unset while a db operation is running.
	s := c.session
	_ = s.KillOps(ctx, s.GetClientAddresses())

	// this establishes a deadline by which we'll forcefully
	// terminate the client connection to ensure we can
	// cleanly terminate the server when we're blocked on a
	// client read/write.
	procutil.PanicSafeGo(func() {
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

	procutil.PanicSafeGo(func() {
		c.server.removeConnection(c)
	}, func(interface{}) {})

	err := c.memoryMonitor.Release(c.memoryMonitor.Allocated())
	if err != nil {
		c.logger.Errf(log.Dev, "memory release error", err)
	}
}

// updateCatalog updates the catalog to utilize the new schema.
func (c *conn) updateCatalog(ctx context.Context, s *schema.Schema) error {
	err := c.loadMongoDBInfo(ctx, s)
	if err != nil {
		return err
	}

	return c.setCatalogFromSchema(s)
}

func (c *conn) setCatalogFromSchema(s *schema.Schema) error {
	cat, err := catalog.Build(s, c.variables, c.mongoDBInfo)
	if err != nil {
		return err
	}

	infoSchema, err := cat.Database(catalog.InformationSchemaDatabase)
	if err != nil {
		return err
	}

	// also add the PROCESSLIST table to the catalog
	err = c.updateWithProcessListTable(infoSchema)
	if err != nil {
		return err
	}

	c.catalog = cat
	return nil
}

// DB returns the current database name.
func (c *conn) DB() string {
	if c.currentDB == nil {
		return ""
	}
	return string(c.currentDB.Name())
}

// dispatch runs the command supplied to the connection.
func (c *conn) dispatch(ctx context.Context, data []byte) (err error) {
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
		c.close(ctx)
		return nil
	case ComQuery:
		s := strutil.String(c.variables.GetCharset(variable.CharacterSetClient).Decode(data))
		c.process.UpdateProcess(CommandQuery, s)
		err = c.handleQuery(ctx, s)
		c.process.UpdateProcess(CommandSleep, "")
		return err
	case ComPing:
		return c.writeOK(nil)
	case ComInitDB:
		s := strutil.String(c.variables.GetCharset(variable.CharacterSetClient).Decode(data))
		if err := c.useDB(s); err != nil {
			return err
		}
		return c.writeOK(nil)
	case ComFieldList:
		return c.handleFieldList(data)
	case ComStmtPrepare:
		s := strutil.String(c.variables.GetCharset(variable.CharacterSetClient).Decode(data))
		return c.handleStmtPrepare(s)
	case ComStmtExecute:
		return c.handleStmtExecute(ctx, data)
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

func (c *conn) handshake(ctx context.Context) error {
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

	currentSchema := c.server.schemaManager.Schema(ctx)
	if currentSchema == nil {
		//lastErr := c.server.schemaManager.GetLastErr()
		msg := "MongoDB schema not yet available"
		/* TODO
		if lastErr != nil {
			msg = fmt.Sprintf("%s; %v", msg, lastErr)
		}
		*/
		err := mysqlerrors.Newf(mysqlerrors.ErHandshakeError, msg)
		c.writeError(err)
		return err
	}

	var err error
	if c.server.cfg.Security.Enabled {
		c.logger.Infof(log.Dev, "configuring client authentication for principal %s", c.user)
		switch c.clientRequestedAuthPluginName {
		case mongosqlAuthClientAuthPluginName:
			err = c.authMongoSQLAuthPlugin(ctx)
		default:
			err = c.authClearTextPasswordPlugin(ctx)
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

	if err = c.loadMongoDBInfo(ctx, currentSchema); err != nil {
		err = mysqlerrors.Newf(mysqlerrors.ErHandshakeError, err.Error())
		c.writeError(err)
		return err
	}

	c.logger.Infof(log.Admin, "connected to MongoDB %v, git version: %v",
		c.mongoDBInfo.Version, c.mongoDBInfo.GitVersion)

	if c.mongoDBInfo.CompatibleVersion != "" {
		c.logger.Infof(log.Admin, "MongoDB version compatibility is %v",
			c.mongoDBInfo.CompatibleVersion)
	}

	if !c.mongoDBInfo.VersionAtLeast(3, 2) {
		err = mysqlerrors.Newf(mysqlerrors.ErHandshakeError,
			"MongoDB version is %v but version >= 3.2 required",
			c.mongoDBInfo.Version)
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

// VersionAtLeast is a function comparing the MongoDB
// server version for which we are translating so that
// we know which aggregation language features are present.
func (c *conn) VersionAtLeast(version ...uint8) bool {
	return c.mongoDBInfo.VersionAtLeast(version...)
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

	// readBytes and the following closuers return nil if the index goes out of bounds.
	readBytes := func(num int) []byte {
		if pos+num > len(data) {
			return nil
		}
		if num < 0 {
			return nil
		}

		result := data[pos : pos+num]
		pos += num
		return result
	}

	readLengthEncodedInt := func() *int {

		length := readBytes(1)
		if length == nil {
			return nil
		}

		// Read the first byte in the sequence to determine the length of bytes to read.
		switch length[0] {

		// If it is 251, the value is NULL.
		case 0xfb:
			return nil

		// If it is 252, the value is in the following 2 bytes.
		case 0xfc:
			numBytes := readBytes(2)
			if numBytes == nil {
				return nil
			}
			result := int(uint64(numBytes[0]) | uint64(numBytes[1])<<8)
			return &result

		// If it is 253, the value is in the following 3 bytes.
		case 0xfd:
			numBytes := readBytes(3)
			if numBytes == nil {
				return nil
			}
			result := int(uint64(numBytes[0]) | uint64(numBytes[1])<<8 | uint64(numBytes[2])<<16)
			return &result

		// If it is 254, the value is in the following 8 bytes.
		case 0xfe:
			numBytes := readBytes(8)
			if numBytes == nil {
				return nil
			}
			result := int(uint64(numBytes[0]) | uint64(numBytes[1])<<8 | uint64(numBytes[2])<<16 |
				uint64(numBytes[3])<<24 | uint64(numBytes[4])<<32 | uint64(numBytes[5])<<40 |
				uint64(numBytes[6])<<48 | uint64(numBytes[7])<<56)
			return &result
		}

		// If it is any value between 0 and 250, the value is the length itself.
		result := int(length[0])
		return &result
	}

	readUntilNul := func() []byte {
		if pos >= len(data) {
			return nil
		}
		result := readBytes(bytes.IndexByte(data[pos:], 0x0))
		pos++
		return result
	}

	readHeader := func() error {
		pos = 0
		b := readBytes(4)
		if b == nil {
			return fmt.Errorf("capability flags bytes expected but incomplete")
		}
		c.capability &= binary.LittleEndian.Uint32(b)

		// Fail if the client handshake is not of the version we accept.
		if c.capability&clientProtocol41 == 0 {
			return fmt.Errorf("server requires CLIENT_PROTOCOL_41")
		}

		// In the case of SSL, some clients won't send anything else until SSL is negotiated.
		if len(data) > 4 {
			// Skip max packet size.
			pos += 4

			// charset
			var col *collation.Collation
			b = readBytes(1)
			if b == nil {
				return fmt.Errorf("character set bytes expected but incomplete")
			}
			col, err = collation.GetByID(collation.ID(b[0]))

			if err == nil {
				names := []variable.Name{
					variable.CharacterSetClient,
					variable.CharacterSetConnection,
					variable.CharacterSetResults,
				}

				for _, name := range names {
					err = c.variables.Set(name, variable.SessionScope, variable.SystemKind,
						values.NewSQLVarchar(values.VariableSQLValueKind, string(col.CharsetName)))
					if err != nil {
						break
					}
				}
			} else {
				c.logger.Warnf(log.Dev, "failed to set collation: %v", err)
			}

			// Skip reserved 23[00].
			pos += 23
		}
		return nil
	}

	err = readHeader()
	if err != nil {
		return err
	}

	clientSSL := c.capability&ClientSSL != 0
	switch c.server.cfg.Net.SSL.Mode {
	case "disabled":
		// Return an error if client is using SSL.
		if clientSSL {
			// We shouldn't ever actually reach this, because the
			// connection should fail during capability negotiation.
			return fmt.Errorf("SSL handshake received but server is started without SSL")
		}
	case "allowSSL":
		// Negotiate SSL if the client is using it.
		// Otherwise, proceed without ssl.
	case "requireSSL":
		// Return an error if client not using SSL.
		if !clientSSL {
			return fmt.Errorf("this server is configured to only allow SSL connections")
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
		err = readHeader()
		if err != nil {
			return err
		}
	}

	// user name string[NUL]
	b := readUntilNul()
	if b == nil {
		return fmt.Errorf("user name b expected but incomplete")
	}
	c.user = strutil.String(c.variables.GetCharset(variable.CharacterSetClient).Decode(b))

	// auth response string[NUL]
	if (c.capability & ClientPluginAuthLenencClientData) != 0 {
		num := readLengthEncodedInt()
		if num == nil {
			return fmt.Errorf("authentication length bytes expected but incomplete")
		}
		b = readBytes(*num)
		if b == nil {
			return fmt.Errorf("authentication bytes expected but incomplete")
		}
		c.clientAuthResponse = b
	} else if (c.capability & ClientSecureConnection) != 0 {
		b = readBytes(1)
		if b == nil {
			return fmt.Errorf("authentication length bytes expected but incomplete")
		}
		b = readBytes(int(b[0]))
		if b == nil {
			return fmt.Errorf("authentication bytes expected but incomplete")
		}
		c.clientAuthResponse = b
	} else {
		b = readUntilNul()
		if b == nil {
			return fmt.Errorf("authentication bytes expected but incomplete")
		}
		c.clientAuthResponse = b
	}

	if (c.capability & ClientInteractive) != 0 {
		c.variables.SetSystemVariable(variable.WaitTimeoutSecs,
			values.NewSQLInt64(values.VariableSQLValueKind, c.variables.GetInt64(variable.InteractiveTimeoutSecs)))
	}

	if pos == len(data) {
		return nil
	}

	if (c.capability & ClientConnectWithDB) != 0 {
		b = readUntilNul()
		if b == nil {
			return fmt.Errorf("database bytes expected but incomplete")
		}

		db := strutil.String(c.variables.GetCharset(variable.CharacterSetClient).Decode(b))
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

		b = readUntilNul()
		if b == nil {
			return fmt.Errorf("client Plugin bytes expected but incomplete")
		}

		// This is always a utf8 string.
		c.clientRequestedAuthPluginName = strutil.String(b)
	}

	// MySQL and the Java SQL driver (and possibly other clients) only set
	// ClientConnectAttrs when authentication is used.
	if (c.capability & ClientConnectAttrs) != 0 {

		num := readLengthEncodedInt()
		if num == nil {
			return fmt.Errorf("attribute length bytes expected but incomplete")
		}

		attrsLen := *num + pos

		attrs := make([]clientConnectionAttribute, 0)
		logString := ""

		for pos < attrsLen {
			num = readLengthEncodedInt()
			if num == nil {
				return fmt.Errorf("attribute key length bytes expected but incomplete")
			}

			b = readBytes(*num)
			if b == nil {
				c.logger.Infof(log.Admin, "error parsing connection attribute key at index %d: EOF",
					len(attrs))
				return fmt.Errorf("invalid connection attribute at index %v: EOF", len(attrs))
			}

			key := strutil.String(c.variables.GetCharset(variable.CharacterSetClient).Decode(b))

			num = readLengthEncodedInt()
			if num == nil {
				return fmt.Errorf("attribute value length bytes expected but incomplete")
			}

			b = readBytes(*num)
			if b == nil {
				c.logger.Infof(log.Admin, "error parsing connection attribute value (key %s): EOF",
					key)
				return fmt.Errorf("invalid connection attribute for key %s: EOF", key)
			}

			val := strutil.String(c.variables.GetCharset(variable.CharacterSetClient).Decode(b))

			attrs = append(attrs, clientConnectionAttribute{key, val})
			logString += fmt.Sprintf("%s:%s, ", key, val)
		}

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

func (c *conn) run(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			c.logger.Errf(log.Dev, "error serving connection: %v\n%s\n", err, buf)
		}
		c.close(ctx)
	}()

	type packetRead struct {
		data []byte
		err  error
	}

	packetReadChan := make(chan packetRead)
	var pkt packetRead

	for {
		procutil.PanicSafeGo(func() {
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

		// Create the context for the query.
		queryCtx, cancelQueryCtx := context.WithCancel(ctx)
		c.cancelQueryCtx = cancelQueryCtx

		if err := c.dispatch(queryCtx, pkt.data); err != nil {
			select {
			case <-queryCtx.Done():
				// This only happens if the query was interrupted.
				err = mysqlerrors.Defaultf(mysqlerrors.ErQueryInterrupted)
			default:
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

// setStatusVariables sets status variables for this client connection.
func (c *conn) setStatusVariables() {
	sessionVariables, globalVariables := c.variables, c.server.variables

	sessionVariables.BytesReceived = &c.bytesReceived
	sessionVariables.BytesSent = &c.bytesSent
	sessionVariables.Connections = globalVariables.Connections
	sessionVariables.Queries = globalVariables.Queries
	sessionVariables.ThreadsConnected = globalVariables.ThreadsConnected
	sessionVariables.StartTime = globalVariables.StartTime

	topology := "standalone"
	if c.mongoDBInfo.IsMongos() {
		topology = "mongos"
	}
	sessionVariables.SetSystemVariable(variable.MongoDBGitVersion,
		values.NewSQLVarchar(values.VariableSQLValueKind, c.mongoDBInfo.GitVersion))
	sessionVariables.SetSystemVariable(variable.MongoDBTopology,
		values.NewSQLVarchar(values.VariableSQLValueKind, topology))
	sessionVariables.SetSystemVariable(variable.MongoDBVersion,
		values.NewSQLVarchar(values.VariableSQLValueKind, c.mongoDBInfo.Version))
}

// loadMongoDBInfo sets system variables that store information about MongoDB.
func (c *conn) loadMongoDBInfo(ctx context.Context, currentSchema *schema.Schema) (err error) {
	c.mongoDBInfo, err = mongodb.LoadInfo(
		ctx,
		c.logger,
		c.server.sessionProvider,
		c.session,
		currentSchema,
		c.server.cfg,
	)
	if err != nil {
		return fmt.Errorf("error retrieving information from MongoDB: %v", err)
	}

	err = c.mongoDBInfo.SetCompatibleVersion(c.server.variables.GetString(variable.MongoDBVersionCompatibility))
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

func (c *conn) useDB(dbName string) error {
	db, err := c.catalog.Database(dbName)
	if err != nil {
		return err
	}
	c.currentDB = db
	serverCollationName := string(c.variables.GetCollation(variable.CollationServer).Name)
	c.variables.SetSystemVariable(variable.CollationDatabase,
		values.NewSQLVarchar(values.VariableSQLValueKind, serverCollationName))
	c.process.SetDB(string(db.Name()))
	return nil
}

// RemoteHost returns the hostname for the current connection.
func (c *conn) remoteHost() string {
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

// updateWithProcessListTable adds the PROCESSLIST table to the catalog after the latter has
// been created. This function resides here due to dependency constraints.
func (c *conn) updateWithProcessListTable(d catalog.Database) error {
	intv := func(name string, i int64) values.NamedSQLValue {
		return values.NewNamedSQLValue(name, values.NewSQLInt64(values.MongoSQLValueKind, i))
	}
	strv := func(name, s string) values.NamedSQLValue {
		return values.NewNamedSQLValue(name, values.NewSQLVarchar(values.MongoSQLValueKind, s))
	}
	pl := "PROCESSLIST"
	id := "ID"
	user := "USER"
	host := "HOST"
	db := "DB"
	command := "COMMAND"
	time := "TIME"
	state := "STATE"
	info := "INFO"
	t := catalog.NewDynamicTable(pl, catalog.SystemView, func() results.Rows {
		var rows results.Rows

		s := c.server

		// Grab a snapshot of the active processes.
		s.activeConnectionsMx.RLock()
		processList := make([]*Process, len(s.activeConnections))
		i := 0
		for _, currConn := range s.activeConnections {
			processList[i] = currConn.process
			i++
		}
		s.activeConnectionsMx.RUnlock()

		for _, p := range processList {
			// If this is the current users process we can show it. If it is
			// not, we need to check that either security is disabled, the
			// user has the `inprog` privilege, or the user is the admin user.
			p.lock.RLock()
			if !c.server.cfg.Security.Enabled || c.canListProcess(p.user) {
				rows = append(rows, results.NewNamedRow(
					"information_schema",
					pl,
					intv(id, int64(p.id)),
					strv(user, p.user),
					strv(host, p.host),
					strv(db, p.db),
					strv(command, p.command),
					intv(time, int64(p.ComputeUptime())),
					strv(state, p.state),
					strv(info, p.info)))
			}
			p.lock.RUnlock()
		}
		return rows
	})

	t.AddColumns(pl,
		catalog.NewDynamicColumnDeclaration(id, types.EvalInt64),
		catalog.NewDynamicColumnDeclaration(user, types.EvalString),
		catalog.NewDynamicColumnDeclaration(host, types.EvalString),
		catalog.NewDynamicColumnDeclaration(db, types.EvalString),
		catalog.NewDynamicColumnDeclaration(command, types.EvalString),
		catalog.NewDynamicColumnDeclaration(time, types.EvalInt64),
		catalog.NewDynamicColumnDeclaration(state, types.EvalString),
		catalog.NewDynamicColumnDeclaration(info, types.EvalString),
	)

	return d.AddTable(t)
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

// nolint: unparam
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
