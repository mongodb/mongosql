package server

import (
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"

	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mysqlerrors"
)

func (c *conn) authClearTextPasswordPlugin() error {
	isUnixSocket := c.conn.LocalAddr().Network() == "unix"

	if (c.capability&CLIENT_SSL) == 0 && !isUnixSocket {
		// require SSL for using cleartext plugin
		err := mysqlerrors.Newf(mysqlerrors.ER_INSECURE_PLAIN_TEXT,
			"ssl is required when using cleartext authentication")
		c.writeError(err)
		return err
	}

	if c.clientRequestedAuthPluginName != clearPasswordClientAuthPluginName {
		err := c.writeAuthSwitchRequest(clearPasswordClientAuthPluginName, []byte{0})
		if err != nil {
			return err
		}

		if err := c.readAuthSwitchResponse(); err != nil {
			return err
		}
	}

	c.logger.Infof(log.Dev, "authenticating client response with %s", clearPasswordClientAuthPluginName)
	username, mechanism, source, err := c.parseUsername()
	if err != nil {
		return fmt.Errorf("failed parsing username: %v", err)
	}

	authenticator := mongodb.CleartextSessionAuthenticator{
		Source:    source,
		Username:  username,
		Password:  string(c.clientAuthResponse[:len(c.clientAuthResponse)-1]), //\0-terminated
		Mechanism: mechanism,
	}

	return c.session.Login(&authenticator)
}

func (c *conn) authMongoSQLAuthPlugin() error {
	c.logger.Infof(log.Dev, "authenticating client response with %s", mongosqlAuthClientAuthPluginName)
	username, mechanism, source, err := c.parseUsername()
	if err != nil {
		return fmt.Errorf("failed parsing username: %v", err)
	}

	switch mechanism {
	case "GSSAPI":
		return c.authMongoSQLAuthGSSAPI(username)
	default:
		return c.authMongoSQLAuthSASL(username, mechanism, source)
	}
}

func (c *conn) authMongoSQLAuthSASL(username, mechanism, source string) error {

	cb := func(conversations []*mongodb.SaslConversation) error {

		var err error
		var data []byte
		for i := 0; i < len(conversations); i++ {
			data = append(data, uint32ToBytes(uint32(len(conversations[i].Payload)))...)
			data = append(data, conversations[i].Payload...)
		}

		if err = c.writeAuthMoreData(data); err != nil {
			return fmt.Errorf("failed writing auth more data: %v", err)
		}

		if data, err = c.readPacket(); err != nil {
			return fmt.Errorf("failed reading auth more data response: %v", err)
		}

		pos := 0
		for i := 0; i < len(conversations); i++ {
			conversations[i].ClientDone = data[pos] == 1
			pos++
			payloadLen := int(binary.LittleEndian.Uint32(data[pos : pos+4]))
			pos += 4
			if payloadLen < 0 || payloadLen+pos > len(data) {
				return fmt.Errorf("payload length out of range: %v", payloadLen)
			}
			conversations[i].Payload = data[pos : pos+payloadLen]
			pos += payloadLen
		}

		return nil
	}

	var err error

	// first send the mechanism to use, followed by any mechanism specific information
	data := append([]byte(mechanism), 0)

	// all mechanisms utilize this
	data = append(data, uint32ToBytes(uint32(c.session.ConnLen()))...)

	if err = c.writeAuthMoreData(data); err != nil {
		return fmt.Errorf("failed writing auth more data: %v", err)
	}

	if data, err = c.readPacket(); err != nil {
		return fmt.Errorf("failed reading auth more data response: %v", err)
	}

	authenticator := mongodb.SaslSessionAuthenticator{
		Source:    source,
		Username:  username,
		Mechanism: mechanism,
		Callback:  cb,
	}

	pos := 0
	for i := 0; i < c.session.ConnLen(); i++ {
		done := data[pos] == 1
		pos++
		payloadLen := int(binary.LittleEndian.Uint32(data[pos : pos+4]))
		pos += 4
		payload := data[pos : pos+payloadLen]
		pos += payloadLen

		authenticator.AddConversation(payload, done)
	}

	return c.session.Login(&authenticator)
}

func (c *conn) authMongoSQLAuthGSSAPI(username string) error {
	cb := func(payload []byte) ([]byte, error) {
		var err error
		var data []byte
		data = append(data, uint32ToBytes(uint32(len(payload)))...)
		data = append(data, payload...)

		if err = c.writeAuthMoreData(data); err != nil {
			return nil, fmt.Errorf("failed writing auth more data: %v", err)
		}

		if data, err = c.readPacket(); err != nil {
			return nil, fmt.Errorf("failed reading auth more data response: %v", err)
		}

		if len(data) < 5 {
			return nil, fmt.Errorf("payload too short")
		}

		// ignoring the first byte as it's the status. In GSSAPI, the GSSAPI
		// protocol carries completion information.
		payloadLen := int(binary.LittleEndian.Uint32(data[1:5]))
		if len(data) < payloadLen+5 {
			return nil, fmt.Errorf("payload too short")
		}
		payload = data[5 : 5+payloadLen]

		return payload, nil
	}

	// first send the mechanism to use, followed by any mechanism specific information
	data := append([]byte("GSSAPI"), 0)
	data = append(data, 1, 0, 0, 0) // number of conversations

	var err error
	if err = c.writeAuthMoreData(data); err != nil {
		return fmt.Errorf("failed writing auth more data: %v", err)
	}

	if data, err = c.readPacket(); err != nil {
		return fmt.Errorf("failed reading auth more data response: %v", err)
	}

	if len(data) < 5 {
		return fmt.Errorf("payload too short")
	}
	payloadLen := int(binary.LittleEndian.Uint32(data[1:5]))
	if len(data) < payloadLen+5 {
		return fmt.Errorf("payload too short")
	}
	payload := data[5 : 5+payloadLen]

	hostname := c.server.cfg.Security.GSSAPI.Hostname
	if hostname == "" {
		hostname = c.server.cfg.Net.BindIP[0]
	}

	authenticator := mongodb.GssapiSessionAuthenticator{
		InitialPayload: payload,
		Callback:       cb,

		HostServiceName: c.server.cfg.Security.GSSAPI.ServiceName,
		HostAddr:        hostname,

		RemoteServiceName: c.server.cfg.MongoDB.Net.Auth.GSSAPIServiceName,
	}

	return c.session.Login(&authenticator)
}

func (c *conn) parseUsername() (username string, mechanism string, source string, err error) {
	username = c.user
	mechanism = c.server.cfg.Security.DefaultMechanism
	source = c.server.cfg.Security.DefaultSource

	// parse user for extra information other than just the username
	// format is username?mechanism=PLAIN&source=db. This is the same
	// as a query string, so everything should be url encoded.
	idx := strings.Index(username, "?")
	if idx > 0 {
		username, err = url.QueryUnescape(c.user[:idx])
		if err != nil {
			return
		}
		var values url.Values
		values, err = url.ParseQuery(c.user[idx+1:])
		if err != nil {
			return
		}
		for key, value := range values {
			switch strings.ToLower(key) {
			case "mechanism":
				mechanism = value[0]
			case "source":
				source = value[0]
			case "servicename":
			case "servicerealm":
			default:
				err = mysqlerrors.Newf(mysqlerrors.ER_HANDSHAKE_ERROR, "unknown authentication option %q", key)
				return
			}
		}
	}

	switch mechanism {
	case "PLAIN", "GSSAPI":
		source = "$external"
	}

	return
}

func (c *conn) readAuthSwitchResponse() error {
	c.logger.Infof(log.Dev, "reading auth switch response")
	data, err := c.readPacket()
	if err != nil {
		return fmt.Errorf("auth switch response error: %v", err)
	}

	if len(data) == 0 {
		c.clientAuthResponse = make([]byte, 0)
	} else {
		c.clientAuthResponse = data
	}

	return nil
}

func (c *conn) writeAuthMoreData(pluginData []byte) error {
	c.logger.Infof(log.Dev, "sending auth more data")

	data := make([]byte, 4, len(pluginData)+5)

	// status [1]
	data = append(data, 0x01)

	// plugin data
	data = append(data, pluginData...)

	if err := c.writePacketandFlush(data); err != nil {
		return fmt.Errorf("auth more data error: %v", err)
	}

	return nil
}

func (c *conn) writeAuthSwitchRequest(plugin string, pluginData []byte) error {
	c.logger.Infof(log.Dev, "sending auth switch request for %s", plugin)
	data := make([]byte, 4, 128)

	// status [1]
	data = append(data, 0xFE)

	// plugin name string[NUL]
	// this is a utf8 string
	data = append(data, []byte(plugin)...)
	data = append(data, 0)

	// auth plugin data string[EOF]
	data = append(data, pluginData...)

	if err := c.writePacketandFlush(data); err != nil {
		return fmt.Errorf("auth switch request error: %v", err)
	}

	return nil
}
