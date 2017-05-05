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
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/ssl"
	"github.com/10gen/sqlproxy/util"
	"github.com/10gen/sqlproxy/variable"
)

var (
	errBadConn       = mysqlerrors.Unknownf("connection was bad")
	errMalformPacket = mysqlerrors.Defaultf(mysqlerrors.ER_MALFORMED_PACKET)
)

type conn struct {
	server  *Server
	session *mongodb.Session
	logger  *log.Logger
	startDb string

	// synchronization variables for
	// terminating connection
	closer       *sync.Cond
	closed       int32
	queryRunning int32

	// synchronization variables for
	// terminating a query
	ctx    context.Context
	cancel context.CancelFunc

	conn     net.Conn
	reader   *bufio.Reader
	writer   io.Writer
	sequence uint8

	capability   uint32
	connectionID uint32
	user         string
	currentDB    *catalog.Database
	lastInsertID int64
	affectedRows int64

	authPluginName                string
	authPluginData                []byte
	clientRequestedAuthPluginName string
	clientAuthResponse            []byte

	stmtID uint32
	stmts  map[uint32]*stmt

	// status variables
	bytesReceived uint64
	bytesSent     uint64

	catalog   *catalog.Catalog
	variables *variable.Container
}

func newConn(s *Server, c net.Conn) (*conn, error) {
	ctx, cancel := context.WithCancel(context.Background())

	session, err := s.sessionProvider.Session(ctx)
	if err != nil {
		return nil, err
	}

	newConn := &conn{
		server:        s,
		session:       session,
		ctx:           ctx,
		cancel:        cancel,
		conn:          c,
		reader:        bufio.NewReaderSize(c, 1024),
		writer:        c,
		closer:        sync.NewCond(&sync.Mutex{}),
		closed:        0,
		bytesReceived: uint64(0),
		bytesSent:     uint64(0),
		queryRunning:  0,
		connectionID:  atomic.AddUint32(&s.connCount, 1),
		capability: CLIENT_PROTOCOL_41 |
			CLIENT_CONNECT_WITH_DB |
			CLIENT_LONG_FLAG |
			CLIENT_LONG_PASSWORD |
			CLIENT_SECURE_CONNECTION |
			CLIENT_PLUGIN_AUTH |
			CLIENT_PLUGIN_AUTH_LENENC_CLIENT_DATA,
		stmts:     make(map[uint32]*stmt),
		variables: variable.NewSessionContainer(s.variables),
	}

	if *s.opts.Auth {
		newConn.authPluginName = mongosqlAuthClientAuthPluginName
		newConn.authPluginData = []byte{1, 0} // version 1.0 of the mongosql_auth plugin
	} else {
		buf, err := randomBuf(20)
		if err != nil {
			return nil, fmt.Errorf("unable to generate salt: %v", err)
		}
		newConn.authPluginName = nativePasswordPluginName
		newConn.authPluginData = buf
	}

	newConn.logger = newConn.Logger(log.NetworkComponent)

	if len(*s.opts.SSLPEMKeyFile) > 0 {
		newConn.capability |= CLIENT_SSL
	}
	return newConn, nil
}

func (c *conn) close() {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		return
	}

	c.cancel()

	// wait for any running queries to be interrupted
	c.closer.L.Lock()
	for atomic.LoadInt32(&c.queryRunning) != 0 {
		// this establishes a deadline by which we'll forcefully
		// terminate the client connection to ensure we can
		// cleanly terminate the server when we're blocked on a
		// client read/write.
		util.PanicSafeGo(func() {
			timer := time.NewTimer(5 * time.Second)
			<-timer.C
			timer.Stop()
			atomic.StoreInt32(&c.queryRunning, 0)
			c.closer.Signal()
		}, func(err interface{}) {
			c.logger.Errf(log.Always, "connection close error: %v", err)
		})

		c.closer.Wait()
	}

	c.closer.L.Unlock()
	c.session.Close()
	c.conn.Close()

	util.PanicSafeGo(func() {
		c.server.removeConnection(c)
	}, func(interface{}) {})
}

// Catalog returns the catalog.
func (c *conn) Catalog() *catalog.Catalog {
	return c.catalog
}

// ConnectionId returns the connection's identifier.
func (c *conn) ConnectionId() uint32 {
	return c.connectionID
}

// Context returns the connection's context.
func (c *conn) Context() context.Context {
	return c.ctx
}

// DB returns the current database name.
func (c *conn) DB() string {
	if c.currentDB == nil {
		return ""
	}
	return string(c.currentDB.Name)
}

func (c *conn) dispatch(data []byte) error {
	if len(data) < 1 {
		return mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_COM_ERROR)
	}

	cmd := data[0]
	data = data[1:]

	if cmd != COM_PING && cmd != COM_STATISTICS {
		atomic.AddUint64(&c.server.queryCount, 1)
	}

	switch cmd {
	case COM_QUIT:
		atomic.StoreInt32(&c.queryRunning, 0)
		c.close()
		return nil
	case COM_QUERY:
		s := String(c.variables.CharacterSetClient.Decode(data))
		return c.handleQuery(s)
	case COM_PING:
		return c.writeOK(nil)
	case COM_INIT_DB:
		s := String(c.variables.CharacterSetClient.Decode(data))
		if err := c.useDB(s); err != nil {
			return err
		}
		return c.writeOK(nil)
	case COM_FIELD_LIST:
		return c.handleFieldList(data)
	case COM_STMT_PREPARE:
		s := String(c.variables.CharacterSetClient.Decode(data))
		return c.handleStmtPrepare(s)
	case COM_STMT_EXECUTE:
		return c.handleStmtExecute(data)
	case COM_STMT_CLOSE:
		return c.handleStmtClose(data)
	case COM_STMT_SEND_LONG_DATA:
		return c.handleStmtSendLongData(data)
	case COM_STMT_RESET:
		return c.handleStmtReset(data)
	default:
		return mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_COM_ERROR)
	}
}

func (c *conn) handshake() error {
	c.logger.Logf(log.DebugHigh, "writing initial handshake")

	if err := c.writeInitialHandshake(); err != nil {
		err = mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "send initial handshake error: %v", err)
		c.writeError(err)
		return err
	}

	c.logger.Logf(log.DebugHigh, "reading handshake response")
	if err := c.readHandshakeResponse(); err != nil {
		err = mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "recv handshake response error: %v", err)
		c.writeError(err)
		return err
	}

	var err error
	if *c.server.opts.Auth {
		c.logger.Logf(log.DebugHigh, "configuring client authentication")
		switch c.clientRequestedAuthPluginName {
		case mongosqlAuthClientAuthPluginName:
			err = c.authMongoSQLAuthPlugin()
		default:
			err = c.authClearTextPasswordPlugin()
		}

		if err != nil {
			err = mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "error performing authentication: %v", err)
			c.writeError(err)
			return err
		}

		c.logger.Logf(log.DebugHigh, "successfully authenticated as principal %s", c.user)
	} else if c.user != "" {
		if c.user != "ODBC" && c.capability&CLIENT_ODBC == 0 {
			c.logger.Warnf(log.Info, "ignoring provided credentials for '%v'; authentication is not enabled", c.user)
		}
		c.user = ""
	}

	if err = c.setSystemVariables(); err != nil {
		return err
	}

	c.setStatusVariables()

	c.catalog, err = catalog.Build(c.server.schema, c.variables)
	if err != nil {
		err = mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "error building catalog: %v", err)
		c.writeError(err)
		return err
	}

	if c.startDb != "" {
		if err := c.useDB(c.startDb); err != nil {
			err = mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "error using database %v: %v", c.startDb, err)
			c.writeError(err)
			return err
		}
	}

	if err := c.writeOK(nil); err != nil {
		return mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "write ok: %v", err)
	}

	c.sequence = 0
	return nil
}

func (c *conn) Kill(id uint32, scope evaluator.KillScope) error {
	if c.ConnectionId() == id {
		return mysqlerrors.Defaultf(mysqlerrors.ER_QUERY_INTERRUPTED)
	}

	c.logger.Logf(log.DebugHigh, "kill %v requested for [conn%v]", scope, id)

	if scope == evaluator.KillQuery {
		return c.server.killQuery(id)
	}

	return c.server.killConnection(id)
}

// LastInsertId returns the last insert id.
func (c *conn) LastInsertId() int64 {
	return c.lastInsertID
}

// ConnectionId returns the connection's identifier.
func (c *conn) Logger(componentStr string) *log.Logger {
	component, globalLogger := "", log.GlobalLogger()
	if componentStr != "" {
		component = fmt.Sprintf("%-10v [conn%v]", componentStr, c.connectionID)
	} else {
		component = fmt.Sprintf("%-10v [conn%v]", globalLogger.GetComponent(), c.connectionID)
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

		//capability
		c.capability = binary.LittleEndian.Uint32(data[:4])
		pos += 4

		// in the case of SSL, some clients won't send anything else until SSL is negotiated.
		if len(data) > 4 {
			//skip max packet size
			pos += 4

			//charset
			col, err := collation.GetByID(collation.ID(data[pos]))
			pos++

			if err == nil {
				names := []variable.Name{
					variable.CharacterSetClient,
					variable.CharacterSetConnection,
					variable.CharacterSetResults,
				}

				for _, name := range names {
					err = c.variables.Set(name, variable.SessionScope, variable.SystemKind, string(col.CharsetName))
					if err != nil {
						break
					}
				}
			}
			if err != nil {
				c.logger.Warnf(log.Info, "failed to set collation: %v", err)
			}

			//skip reserved 23[00]
			pos += 23
		}
	}

	readHeader()

	if (c.capability & CLIENT_SSL) != 0 {
		c.logger.Logf(log.DebugLow, "negotiating ssl")
		if err := c.useTLS(); err != nil {
			err = mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "ssl configuration error: %v", err)
			c.writeError(err)
			return err
		}

		data, err = c.readPacket()
		if err != nil {
			err = mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "continuation after successfull ssl negotiation failed: %v", err)
			c.writeError(err)
			return err
		}

		// We need to read the handshake response header again because, now that we have TLS, the
		// client resends its handshake response packet.
		readHeader()
	}

	//user name string[NUL]
	userBytes := data[pos : pos+bytes.IndexByte(data[pos:], 0)]
	pos += len(userBytes) + 1
	c.user = String(c.variables.CharacterSetClient.Decode(userBytes))

	//auth response string[NUL]
	if (c.capability & CLIENT_PLUGIN_AUTH_LENENC_CLIENT_DATA) != 0 {
		authLen, _, count := lengthEncodedInt(data[pos:])
		pos += count
		c.clientAuthResponse = data[pos : pos+int(authLen)]
		pos += int(authLen)
	} else if (c.capability & CLIENT_SECURE_CONNECTION) != 0 {
		authLen := int(data[pos])
		pos++
		c.clientAuthResponse = data[pos : pos+authLen]
		pos += authLen
	} else {
		c.clientAuthResponse = data[pos : pos+bytes.IndexByte(data[pos:], 0)]
		pos += len(c.clientAuthResponse) + 1
	}

	if (c.capability & CLIENT_INTERACTIVE) != 0 {
		c.variables.WaitTimeoutSecs = c.variables.InteractiveTimeoutSecs
	}

	if pos == len(data) {
		return nil
	}

	if (c.capability & CLIENT_CONNECT_WITH_DB) != 0 {
		dbBytes := data[pos : pos+bytes.IndexByte(data[pos:], 0)]
		pos += len(dbBytes) + 1

		db := String(c.variables.CharacterSetClient.Decode(dbBytes))
		c.startDb = db
	}

	if pos == len(data) {
		return nil
	}

	if (c.capability & CLIENT_PLUGIN_AUTH) != 0 {
		// The Java Driver has a bug where it sends an extra nul byte when
		// there is no initial db. So, if we don't have CLIENT_CONNECT_WITH_DB
		// and the current byte is nul, we'll skip it.
		// REF: https://bugs.mysql.com/bug.php?id=79612
		if c.capability&CLIENT_CONNECT_WITH_DB == 0 && data[pos] == 0 {
			pos++
			if pos == len(data) {
				return nil
			}
		}

		clientPluginNameBytes := data[pos : pos+bytes.IndexByte(data[pos:], 0)]
		pos += len(clientPluginNameBytes) + 1

		// this is always a utf8 string
		c.clientRequestedAuthPluginName = String(clientPluginNameBytes)
	}

	if (c.capability & CLIENT_CONNECT_ATTRS) != 0 {
		// ignore this
	}

	return nil
}

func (c *conn) readPacket() ([]byte, error) {
	data, err := c.readPacketHelper()
	if err != nil {
		return nil, err
	}

	headerLength, payloadLength := uint64(4), len(data)
	for payloadLength > maxPayloadLength {
		payloadLength -= maxPayloadLength
		headerLength += 4
	}

	nBytesReceived := uint64(len(data)) + headerLength
	atomic.AddUint64(&c.bytesReceived, nBytesReceived)
	atomic.AddUint64(&c.server.bytesReceived, nBytesReceived)
	return data, nil
}

func (c *conn) readPacketHelper() ([]byte, error) {
	header := []byte{0, 0, 0, 0}

	if _, err := io.ReadFull(c.reader, header); err != nil {
		return nil, errBadConn
	}

	length := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)

	sequence := uint8(header[3])
	if sequence != c.sequence {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_NET_PACKETS_OUT_OF_ORDER)
	}

	c.sequence++

	data := make([]byte, length)
	if _, err := io.ReadFull(c.reader, data); err != nil {
		return nil, errBadConn
	}

	if length < maxPayloadLength {
		return data, nil
	}

	buf, err := c.readPacketHelper()
	if err != nil {
		return nil, errBadConn
	}

	data = append(data, buf...)
	return data, nil
}

// RowCount returns the number of rows affected by the last statement.
func (c *conn) RowCount() int64 {
	return c.affectedRows
}

// refreshContext creates a new context for this connection.
func (c *conn) refreshContext() {
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.session.SetContext(c.ctx)
}

func (c *conn) run() {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			c.logger.Errf(log.Info, "%v, %s", err, buf)
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
			c.logger.Errf(log.Always, "packet read error: %v", err)
		})

		waitTimeout := time.Duration(c.variables.WaitTimeoutSecs) * time.Second
		timer := time.NewTimer(waitTimeout)
		var timeoutTime time.Time

		select {
		case timeoutTime = <-timer.C:
			c.logger.Logf(log.Always, "client wait time out after %v", waitTimeout.String())
		case pkt = <-packetReadChan:
			if pkt.err != nil && atomic.LoadInt32(&c.closed) != 1 {
				c.logger.Logf(log.Always, "client read error: %v", pkt.err)
			}
		}

		timer.Stop()

		if !timeoutTime.IsZero() || pkt.err != nil {
			atomic.StoreInt32(&c.queryRunning, 0)
			c.closer.Signal()
			return
		}

		if err := c.dispatch(pkt.data); err != nil {
			// we only cancel a context when a query interruption
			// is requested
			if err == context.Canceled {
				err = mysqlerrors.Defaultf(mysqlerrors.ER_QUERY_INTERRUPTED)
			}

			c.logger.Errf(log.Always, "dispatch error: %v", err)
			if err != errBadConn {
				c.writeError(err)
			}
		}

		atomic.StoreInt32(&c.queryRunning, 0)
		c.closer.Signal()
		c.sequence = 0
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
	globalVariables.BytesReceived = &c.server.bytesReceived

	sessionVariables.BytesSent = &c.bytesSent
	globalVariables.BytesSent = &c.server.bytesSent

	sessionVariables.Connections = &c.server.connCount
	globalVariables.Connections = &c.server.connCount

	sessionVariables.Queries = &c.server.queryCount
	globalVariables.Queries = &c.server.queryCount

	sessionVariables.ThreadsConnected = &c.server.threadsConnected
	globalVariables.ThreadsConnected = &c.server.threadsConnected

	sessionVariables.StartTime = &c.server.startTime
	globalVariables.StartTime = &c.server.startTime
}

// setSystemVariables sets system variables for this client connection.
func (c *conn) setSystemVariables() (err error) {
	c.variables.MongoDBInfo, err = mongodb.LoadInfo(c.logger, c.session,
		c.server.schema, *c.server.opts.Auth)
	if err != nil {
		err = mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR,
			"error retrieving information from MongoDB: %v", err)
		c.writeError(err)
		return err
	}

	err = c.variables.MongoDBInfo.SetCompatibleVersion(
		c.server.variables.MongoDBVersionCompatibility)
	if err != nil {
		err = mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR,
			"error setting compatibility version: %v", err)
		c.writeError(err)
		return err
	}

	c.logger.Logf(log.Info, "connected to MongoDB %v, git version: %v",
		c.variables.MongoDBInfo.Version, c.variables.MongoDBInfo.GitVersion)

	if c.variables.MongoDBInfo.CompatibleVersion != "" {
		c.logger.Logf(log.Info, "MongoDB version compatibility is %v",
			c.variables.MongoDBInfo.CompatibleVersion)
	}

	if !c.variables.MongoDBInfo.VersionAtLeast(3, 2) {
		err = mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR,
			"MongoDB version is %v but version >= 3.2 required",
			c.variables.MongoDBInfo.Version)
		c.writeError(err)
		return err
	}

	return nil
}

func (c *conn) status() uint16 {
	if c.variables.AutoCommit {
		return SERVER_STATUS_AUTOCOMMIT
	}

	return 0
}

func (c *conn) useDB(db string) error {
	d, err := c.catalog.Database(db)
	if err != nil {
		return err
	}

	c.currentDB = d
	c.variables.CollationDatabase = c.variables.CollationServer
	return nil
}

// User returns the current user.
func (c *conn) User() string {
	return fmt.Sprintf("%s@%s", c.user, c.conn.RemoteAddr().String())
}

func (c *conn) useTLS() error {
	tlsc, err := ssl.Handshake(c.conn, c.server.opts)
	if err != nil {
		return err
	}

	c.conn = tlsc
	c.reader = bufio.NewReaderSize(c.conn, 1024)
	c.writer = c.conn

	return nil
}

func (c *conn) Variables() *variable.Container {
	return c.variables
}

func (c *conn) writeEOF(status uint16) error {
	data := make([]byte, 4, 9)

	data = append(data, EOF_HEADER)
	if c.capability&CLIENT_PROTOCOL_41 > 0 {
		data = append(data, 0, 0)
		data = append(data, byte(status), byte(status>>8))
	}

	return c.writePacket(data)
}

func (c *conn) writeError(e error) error {
	var m *mysqlerrors.MySqlError
	var ok bool
	if m, ok = e.(*mysqlerrors.MySqlError); !ok {
		m = mysqlerrors.Unknownf(e.Error())
	}

	data := make([]byte, 4, 16+len(m.Message))

	data = append(data, ERR_HEADER)
	data = append(data, byte(m.Code), byte(m.Code>>8))

	if c.capability&CLIENT_PROTOCOL_41 > 0 {
		data = append(data, '#')
		data = append(data, m.State...)
	}

	data = append(data, m.Message...)

	return c.writePacket(data)
}

func (c *conn) writeInitialHandshake() error {
	data := make([]byte, 4, 128)

	//min version 10
	data = append(data, minProtocolVersion)

	//server version[00]
	version := c.server.variables.Version
	versionComment := c.server.variables.VersionComment

	data = append(data, fmt.Sprintf("%s %s", version, versionComment)...)
	data = append(data, 0)

	//connection id
	data = append(data, byte(c.connectionID), byte(c.connectionID>>8), byte(c.connectionID>>16), byte(c.connectionID>>24))

	//auth-plugin-data-part-1
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

	//filler [00]
	data = append(data, 0)

	//capability flag lower 2 bytes
	data = append(data, byte(c.capability), byte(c.capability>>8))

	//charset
	data = append(data, byte(collation.Default.ID))

	//status
	status := c.status()
	data = append(data, byte(status), byte(status>>8))

	//below 13 byte may not be used
	//capability flag upper 2 bytes
	data = append(data, byte(c.capability>>16), byte(c.capability>>24))

	if (c.capability & CLIENT_PLUGIN_AUTH) != 0 {
		max := len(c.authPluginData)
		if max > 0 && max < 21 {
			max = 21
		}
		data = append(data, byte(max))
	} else {
		data = append(data, 0)
	}

	//reserved 10 [00]
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)

	if (c.capability & CLIENT_SECURE_CONNECTION) != 0 {
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

	if (c.capability & CLIENT_PLUGIN_AUTH) != 0 {
		// auth-plugin-name string[NUL]
		data = append(data, []byte(c.authPluginName)...)
		data = append(data, 0)
	}

	return c.writePacket(data)
}

func (c *conn) writePacket(data []byte) error {
	length := len(data) - 4

	for length >= maxPayloadLength {

		data[0] = 0xff
		data[1] = 0xff
		data[2] = 0xff

		data[3] = c.sequence

		if n, err := c.writer.Write(data[:4+maxPayloadLength]); err != nil {
			return errBadConn
		} else if n != (4 + maxPayloadLength) {
			return errBadConn
		} else {
			c.sequence++
			length -= maxPayloadLength
			data = data[maxPayloadLength:]
		}
	}

	data[0] = byte(length)
	data[1] = byte(length >> 8)
	data[2] = byte(length >> 16)
	data[3] = c.sequence

	if n, err := c.writer.Write(data); err != nil {
		return errBadConn
	} else if n != len(data) {
		return errBadConn
	}

	c.sequence++
	nBytesSent := uint64(len(data))
	atomic.AddUint64(&c.bytesSent, nBytesSent)
	atomic.AddUint64(&c.server.bytesSent, nBytesSent)
	return nil
}

func (c *conn) writeOK(r *Result) error {
	if r == nil {
		r = &Result{Status: c.status()}
	}
	data := make([]byte, 4, 32)

	data = append(data, OK_HEADER)

	data = append(data, putLengthEncodedInt(r.AffectedRows)...)
	data = append(data, putLengthEncodedInt(r.InsertId)...)

	if c.capability&CLIENT_PROTOCOL_41 > 0 {
		data = append(data, byte(r.Status), byte(r.Status>>8))
		data = append(data, 0, 0)
	}

	return c.writePacket(data)
}
