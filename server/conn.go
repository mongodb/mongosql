package server

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/10gen/sqlproxy/common"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"github.com/mongodb/mongo-tools/common/log"
	"gopkg.in/mgo.v2"
)

var (
	errBadConn       = mysqlerrors.Unknownf("connection was bad")
	errMalformPacket = mysqlerrors.Defaultf(mysqlerrors.ER_MALFORMED_PACKET)
)

type conn struct {
	sync.Mutex

	server  *Server
	session *mgo.Session
	closed  bool

	conn     net.Conn
	reader   *bufio.Reader
	writer   io.Writer
	sequence uint8

	capability   uint32
	connectionID uint32
	status       uint16
	collation    collationId
	charset      string
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

	variables *sessionVariableContainer
}

func newConn(s *Server, c net.Conn) *conn {
	newConn := &conn{
		server:       s,
		session:      s.eval.Session(),
		conn:         c,
		reader:       bufio.NewReaderSize(c, 1024),
		writer:       c,
		connectionID: atomic.AddUint32(&s.connCount, 1),
		status:       SERVER_STATUS_AUTOCOMMIT,
		collation:    DEFAULT_COLLATION_ID,
		capability: CLIENT_PROTOCOL_41 |
			CLIENT_CONNECT_WITH_DB |
			CLIENT_LONG_FLAG |
			CLIENT_LONG_PASSWORD |
			CLIENT_SECURE_CONNECTION,
		charset:   DEFAULT_CHARSET,
		stmts:     make(map[uint32]*stmt),
		variables: newSessionVariableContainer(s.variables),
	}

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

// LastInsertId returns the last insert id.
func (c *conn) LastInsertId() int64 {
	return c.lastInsertID
}

// RowCount returns the number of rows affected by the last statement.
func (c *conn) RowCount() int64 {
	return c.affectedRows
}

// Session returns the mgo.Session currently opened with MongoDB.
func (c *conn) Session() *mgo.Session {
	return c.session
}

// User returns the current user.
func (c *conn) User() string {
	return fmt.Sprintf("%s@%s", c.user, c.conn.RemoteAddr().String())
}

func (c *conn) GetVariable(name string, kind evaluator.VariableKind) (evaluator.SQLValue, error) {
	switch kind {
	case evaluator.GlobalVariable, evaluator.SessionVariable:
		def, err := getSystemVariableDefinition(name)
		if err != nil {
			return nil, err
		}

		defRead, ok := def.(readableVariableDefinition)
		if !ok {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ER_VAR_CANT_BE_READ, name)
		}

		if kind == evaluator.GlobalVariable {
			if value, ok := c.server.variables.getValue(name); ok {
				return value, nil
			}
		} else {
			if value, ok := c.variables.getSessionVariable(name); ok {
				return value, nil
			}
		}

		return defRead.defaultValue(), nil
	default:
		v, ok := c.variables.getUserVariable(name)
		if !ok {
			return nil, nil
		}

		return v, nil
	}
}

func (c *conn) SetVariable(name string, value evaluator.SQLValue, kind evaluator.VariableKind) error {

	switch kind {
	case evaluator.GlobalVariable, evaluator.SessionVariable:
		def, err := getSystemVariableDefinition(name)
		if err != nil {
			return err
		}

		defWrite, ok := def.(settableVariableDefinition)
		if !ok {
			return mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_SYSTEM_VARIABLE, name)
		}

		scope := globalScope
		if kind == evaluator.SessionVariable {
			scope = sessionScope
		}

		return defWrite.apply(c, scope, value)
	default:
		c.variables.setUserVariable(name, value)
		return nil
	}
}

func (c *conn) handshake() error {
	if err := c.writeInitialHandshake(); err != nil {
		c.writeError(err)
		return mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "send initial handshake error: %v", err)
	}

	if err := c.readHandshakeResponse(); err != nil {
		c.writeError(err)
		return mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "recv handshake response error: %v", err)
	}

	if c.server.opts.Auth {
		if (c.capability & CLIENT_PLUGIN_AUTH) == 0 {
			err := mysqlerrors.Defaultf(mysqlerrors.ER_NOT_SUPPORTED_AUTH_MODE)
			c.writeError(err)
			return err
		}

		if c.clientRequestedAuthPluginName != clearPasswordClientAuthPluginName {
			if err := c.writeAuthRequestSwitchPacket(clearPasswordClientAuthPluginName); err != nil {
				c.writeError(err)
				return mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "send auth request switch error: %v", err)
			}

			if err := c.readAuthSwitchResponsePacket(); err != nil {
				c.writeError(err)
				return mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "recv auth response error: %v", err)
			}
		}

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

func (c *conn) writeInitialHandshake() error {
	data := make([]byte, 4, 128)

	//min version 10
	data = append(data, minProtocolVersion)

	//server version[00]
	data = append(data, common.VersionStr...)
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

	//charset, utf-8 default
	data = append(data, uint8(c.collation))

	//status
	data = append(data, byte(c.status), byte(c.status>>8))

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

		//charset, skip, if you want to use another charset, use set names
		//c.collation = CollationId(data[pos])
		pos++

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
		log.Log(log.Always, err.Error())
		c.writeError(err)
		return err
	}

	//user name string[NUL]
	c.user = string(data[pos : pos+bytes.IndexByte(data[pos:], 0)])
	pos += len(c.user) + 1

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

	if pos == len(data) {
		return nil
	}

	if (c.capability & CLIENT_CONNECT_WITH_DB) != 0 {
		db := string(data[pos : pos+bytes.IndexByte(data[pos:], 0)])
		pos += len(db) + 1

		if err := c.useDB(db); err != nil {
			return err
		}
	}

	if pos == len(data) {
		return nil
	}

	if (c.capability & CLIENT_PLUGIN_AUTH) != 0 {
		c.clientRequestedAuthPluginName = string(data[pos : pos+bytes.IndexByte(data[pos:], 0)])
		pos += len(c.clientRequestedAuthPluginName) + 1
	}

	if (c.capability & CLIENT_CONNECT_ATTRS) != 0 {
		// ignore this
	}

	return nil
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

func (c *conn) readAuthSwitchResponsePacket() error {
	data, err := c.readPacket()
	if err != nil {
		return err
	}

	c.clientAuthResponse = data[:len(data)-1]

	return nil
}

func (c *conn) useTLS() error {

	tlsc := tls.Server(c.conn, c.server.tlsConfig)
	if err := tlsc.Handshake(); err != nil {
		return err
	}
	c.conn = tlsc
	c.reader = bufio.NewReaderSize(c.conn, 1024)
	c.writer = c.conn

	return nil
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

	return c.session.Login(credential)
}

func (c *conn) run() {
	defer func() {
		r := recover()
		if err, ok := r.(error); ok {
			const size = 4096
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]

			log.Logf(log.Always, "%v, %s", err, buf)
		}

		c.close()
	}()

	for {
		data, err := c.readPacket()

		if err != nil {
			return
		}

		if err := c.dispatch(data); err != nil {
			log.Logf(log.Always, "[conn%v] dispatch error: %v", c.connectionID, err)
			if err != errBadConn {
				c.writeError(err)
			}
		}

		if c.closed {
			return
		}

		c.sequence = 0
	}
}

func (c *conn) dispatch(data []byte) error {
	cmd := data[0]
	data = data[1:]

	switch cmd {
	case COM_QUIT:
		c.close()
		return nil
	case COM_QUERY:
		return c.handleQuery(String(data))
	case COM_PING:
		return c.writeOK(nil)
	case COM_INIT_DB:
		if err := c.useDB(String(data)); err != nil {
			return err
		}
		return c.writeOK(nil)
	case COM_FIELD_LIST:
		return c.handleFieldList(data)
	case COM_STMT_PREPARE:
		return c.handleStmtPrepare(String(data))
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

func (c *conn) close() error {
	if c.closed {
		return nil
	}

	c.session.Close()

	c.conn.Close()

	c.closed = true

	log.Logf(log.Info, "[conn%v] end connection %v", c.connectionID, c.conn.RemoteAddr())

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

func (c *conn) useDB(db string) error {
	s := c.server.databases[db]
	if s == nil {
		return mysqlerrors.Defaultf(mysqlerrors.ER_BAD_DB_ERROR, db)
	}

	c.currentDB = s
	return nil
}

func (c *conn) writeOK(r *Result) error {
	if r == nil {
		r = &Result{Status: c.status}
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

func (c *conn) writeEOF(status uint16) error {
	data := make([]byte, 4, 9)

	data = append(data, EOF_HEADER)
	if c.capability&CLIENT_PROTOCOL_41 > 0 {
		data = append(data, 0, 0)
		data = append(data, byte(status), byte(status>>8))
	}

	return c.writePacket(data)
}
