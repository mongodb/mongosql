package server

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/10gen/sqlproxy/client/openssl"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/util"
	"github.com/10gen/sqlproxy/variable"

	"gopkg.in/mgo.v2"
	"gopkg.in/tomb.v2"
)

var (
	errBadConn       = mysqlerrors.Unknownf("connection was bad")
	errMalformPacket = mysqlerrors.Defaultf(mysqlerrors.ER_MALFORMED_PACKET)
)

type conn struct {
	sync.Mutex

	server    *Server
	session   *mgo.Session
	closed    bool
	tomb      *tomb.Tomb
	logger    *log.Logger
	startDb   string
	queryChan chan struct{}

	conn     net.Conn
	reader   *bufio.Reader
	writer   io.Writer
	sequence uint8

	capability   uint32
	connectionID uint32
	user         string
	currentDB    *schema.Database
	lastInsertID int64
	affectedRows int64

	authPluginName                string
	authPluginData                []byte
	clientRequestedAuthPluginName string
	clientAuthResponse            []byte

	stmtID uint32
	stmts  map[uint32]*stmt

	variables *variable.Container
}

func newConn(s *Server, c net.Conn) *conn {
	newConn := &conn{
		server:       s,
		session:      s.eval.Session(),
		conn:         c,
		reader:       bufio.NewReaderSize(c, 1024),
		writer:       c,
		queryChan:    make(chan struct{}),
		tomb:         &tomb.Tomb{},
		connectionID: atomic.AddUint32(&s.connCount, 1),
		capability: CLIENT_PROTOCOL_41 |
			CLIENT_CONNECT_WITH_DB |
			CLIENT_LONG_FLAG |
			CLIENT_LONG_PASSWORD |
			CLIENT_SECURE_CONNECTION,
		stmts:     make(map[uint32]*stmt),
		variables: variable.NewSessionContainer(s.variables),
	}

	newConn.logger = newConn.Logger(log.NetworkComponent)
	if s.tlsConfig != nil {
		newConn.capability |= CLIENT_SSL
	}
	if s.opts.Auth {
		newConn.authPluginName = clearPasswordClientAuthPluginName
		newConn.capability |= CLIENT_PLUGIN_AUTH | CLIENT_PLUGIN_AUTH_LENENC_CLIENT_DATA
	} else {
		// We are going to leave this as native password auth because
		// when auth isn't turned on, using clear-text will still require
		// the client to opt-in which, when no password is going to be sent,
		// is just annoying.
		newConn.authPluginName = nativePasswordClientAuthPluginName
		newConn.authPluginData, _ = randomBuf(20)
	}

	return newConn
}

func (c *conn) authenticate() error {

	credential := &mgo.Credential{
		Username: c.user,
		Password: string(c.clientAuthResponse),
	}

	if c.currentDB != nil {
		credential.Source = c.currentDB.Name
	}

	// parse user for extra information other than just the username
	// format is username?mechanism=PLAIN&source=db. This is the same
	// as a query string, so everything should be url encoded.
	idx := strings.Index(credential.Username, "?")
	var err error
	if idx > 0 {
		credential.Username, err = url.QueryUnescape(c.user[:idx])
		if err != nil {
			return err
		}
		values, err := url.ParseQuery(c.user[idx+1:])
		if err != nil {
			return err
		}
		for key, value := range values {
			switch strings.ToLower(key) {
			case "mechanism":
				credential.Mechanism = value[0]
			case "source":
				credential.Source = value[0]
			default:
				return mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "unknown authentication option %q", key)
			}
		}
	}

	if credential.Mechanism == "" {
		credential.Mechanism = "SCRAM-SHA-1"
	}

	c.logger.Logf(log.DebugHigh, "attempting to authenticate principal '%v' on '%v' using '%v'", credential.Username, credential.Source, credential.Mechanism)

	if err = c.session.Login(credential); err != nil {
		return err
	}

	c.logger.Logf(log.DebugHigh, "successfully authenticated as principal '%v' on '%v'", credential.Username, credential.Source)
	return nil
}

func (c *conn) close() error {
	if c.closed {
		return nil
	}

	connID := c.ConnectionId()

	c.server.Lock()
	delete(c.server.activeConnections, connID)
	c.server.Unlock()

	c.tomb.Kill(mysqlerrors.Defaultf(mysqlerrors.ER_QUERY_INTERRUPTED))

	c.session.Close()

	c.conn.Close()

	c.closed = true

	// this isn't critical so neglecting to lock
	numConns := len(c.server.activeConnections)
	pluralized := util.Pluralize(numConns, "connection", "connections")
	c.logger.Logf(log.Always, "end connection %v (%v %v now open)", c.conn.RemoteAddr(), numConns, pluralized)

	return nil
}

// ConnectionId returns the connection's identifier.
func (c *conn) ConnectionId() uint32 {
	return c.connectionID
}

// DB returns the current database name.
func (c *conn) DB() string {
	if c.currentDB == nil {
		return ""
	}
	return c.currentDB.Name
}

func (c *conn) dispatch(data []byte) error {
	cmd := data[0]
	data = data[1:]

	switch cmd {
	case COM_QUIT:
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
		c.writeError(err)
		return mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "send initial handshake error: %v", err)
	}

	c.logger.Logf(log.DebugHigh, "reading handshake response")
	if err := c.readHandshakeResponse(); err != nil {
		c.writeError(err)
		return mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "recv handshake response error: %v", err)
	}

	if c.server.opts.Auth {
		c.logger.Logf(log.DebugHigh, "configuring client authentication")
		if (c.capability & CLIENT_PLUGIN_AUTH) == 0 {
			err := mysqlerrors.Defaultf(mysqlerrors.ER_NOT_SUPPORTED_AUTH_MODE)
			c.writeError(err)
			return err
		}

		if c.clientRequestedAuthPluginName != clearPasswordClientAuthPluginName {
			c.logger.Logf(log.DebugHigh, "sending authentication switch request")
			if err := c.writeAuthRequestSwitchPacket(clearPasswordClientAuthPluginName); err != nil {
				c.writeError(err)
				return mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "send auth request switch error: %v", err)
			}

			c.logger.Logf(log.DebugHigh, "reading authentication switch response")
			if err := c.readAuthSwitchResponsePacket(); err != nil {
				c.writeError(err)
				return mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "recv auth response error: %v", err)
			}
		}

		c.logger.Logf(log.DebugHigh, "authenticating client response")
		if err := c.authenticate(); err != nil {
			c.writeError(err)
			return mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "failed authentication with MongoDB: %v", err)
		}
	} else {
		c.user = ""
	}

	if err := c.writeOK(nil); err != nil {
		return mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "write ok error: %v", err)
	}

	c.sequence = 0

	return nil
}

func (c *conn) Kill(id uint32, scope evaluator.KillScope) error {
	if c.ConnectionId() == id {
		return mysqlerrors.Defaultf(mysqlerrors.ER_QUERY_INTERRUPTED)
	}

	c.server.Lock()
	k, ok := c.server.activeConnections[id]
	if !ok {
		c.server.Unlock()
		return mysqlerrors.Defaultf(mysqlerrors.ER_NO_SUCH_THREAD, id)
	}

	c.server.Unlock()

	c.logger.Logf(log.DebugHigh, "kill %v requested for [conn%v]", scope, k.ConnectionId())

	switch scope {
	case evaluator.KillQuery:
		k.tomb.Kill(mysqlerrors.Defaultf(mysqlerrors.ER_QUERY_INTERRUPTED))
		return nil
	default:
		return k.close()
	}
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

func (c *conn) readAuthSwitchResponsePacket() error {
	data, err := c.readPacket()
	if err != nil {
		return err
	}

	c.clientAuthResponse = data[:len(data)-1]

	return nil
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

	readHeader()

	if (c.capability & CLIENT_SSL) != 0 {
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
	} else if c.server.opts.Auth {
		// We are here because we asked the client to use SSL and they refused. Therefore, we'll
		// terminate the connection.
		err := mysqlerrors.Newf(mysqlerrors.ER_INSECURE_PLAIN_TEXT, "ssl is required when using authentication")
		c.logger.Errf(log.Always, err.Error())
		c.writeError(err)
		return err
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
		clientPluginNameBytes := data[pos : pos+bytes.IndexByte(data[pos:], 0)]
		pos += len(clientPluginNameBytes) + 1

		c.clientRequestedAuthPluginName = String(c.variables.CharacterSetClient.Decode(clientPluginNameBytes))
		if err != nil {
			return err
		}
	}

	if (c.capability & CLIENT_CONNECT_ATTRS) != 0 {
		// ignore this
	}

	return nil
}

func (c *conn) readPacket() ([]byte, error) {
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

	if length < maxPayloadLen {
		return data, nil
	}

	buf, err := c.readPacket()
	if err != nil {
		return nil, errBadConn
	}

	return append(data, buf...), nil
}

// RowCount returns the number of rows affected by the last statement.
func (c *conn) RowCount() int64 {
	return c.affectedRows
}

func (c *conn) run() {
	defer func() {
		r := recover()
		if err, ok := r.(error); ok {
			const size = 4096
			buf := make([]byte, size)
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
		go func() {
			data, err := c.readPacket()
			packetReadChan <- packetRead{data, err}
			c.queryChan = make(chan struct{})
		}()

		waitTimeout := time.Duration(c.variables.WaitTimeoutSecs) * time.Second

		select {
		case <-time.After(waitTimeout):
			c.logger.Logf(log.Always, "client wait time out after %v", waitTimeout.String())
			close(c.queryChan)
			return
		case pkt = <-packetReadChan:
			if pkt.err != nil {
				close(c.queryChan)
				return
			}
		}

		if err := c.dispatch(pkt.data); err != nil {
			c.logger.Logf(log.Always, "dispatch error: %v", err)
			if err != errBadConn {
				c.writeError(err)
			}
		}

		if c.closed {
			close(c.queryChan)
			return
		}

		close(c.queryChan)
		c.sequence = 0
	}
}

// Session returns a new mgo.Session connected to MongoDB.
func (c *conn) Session() (session *mgo.Session) {
	return c.session.Copy()
}

func (c *conn) status() uint16 {
	if c.variables.AutoCommit {
		return SERVER_STATUS_AUTOCOMMIT
	}

	return 0
}

func (c *conn) Tomb() *tomb.Tomb {
	return c.tomb
}

func (c *conn) useDB(db string) error {
	if !c.variables.MongoDBInfo.IsAnyAllowedDatabase(mongodb.DatabaseName(db)) {
		return mysqlerrors.Defaultf(mysqlerrors.ER_BAD_DB_ERROR, db)
	}

	s := c.server.databases[strings.ToLower(db)]
	if s == nil {
		return mysqlerrors.Defaultf(mysqlerrors.ER_BAD_DB_ERROR, db)
	}

	c.currentDB = s
	c.variables.CollationDatabase = c.variables.CollationServer
	return nil
}

// User returns the current user.
func (c *conn) User() string {
	return fmt.Sprintf("%s@%s", c.user, c.conn.RemoteAddr().String())
}

func (c *conn) useTLS() error {
	tlsc, err := openssl.Server(c.conn, c.server.tlsConfig)
	if err != nil {
		return err
	}

	if err := tlsc.Handshake(); err != nil {
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

func (c *conn) writeAuthRequestSwitchPacket(pluginName string) error {

	data := make([]byte, 4, 128)

	// status [1]
	data = append(data, 0xFE)

	// plugin name string[NUL]
	data = append(data, []byte(pluginName)...)
	data = append(data, 0)

	// auth plugin data string[EOF]
	// nothing to add

	return c.writePacket(data)
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

	col, err := collation.Get(c.variables.CharacterSetResults.DefaultCollationName)
	if err != nil {
		return err
	}

	//charset
	data = append(data, uint8(col.ID))

	//status
	status := c.status()
	data = append(data, byte(status), byte(status>>8))

	//below 13 byte may not be used
	//capability flag upper 2 bytes
	data = append(data, byte(c.capability>>16), byte(c.capability>>24))

	if (c.capability & CLIENT_PLUGIN_AUTH) != 0 {
		data = append(data, byte(len(c.authPluginData)))
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

	for length >= maxPayloadLen {

		data[0] = 0xff
		data[1] = 0xff
		data[2] = 0xff

		data[3] = c.sequence

		if n, err := c.writer.Write(data[:4+maxPayloadLen]); err != nil {
			return errBadConn
		} else if n != (4 + maxPayloadLen) {
			return errBadConn
		} else {
			c.sequence++
			length -= maxPayloadLen
			data = data[maxPayloadLen:]
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
	} else {
		c.sequence++
		return nil
	}
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
