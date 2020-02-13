package server

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/mongodb/provider"
)

// TestCannotGetSession is a regression test for a panic when serving
// a new connection. This panic occurs when an error (such as an
// unreachable mongod) occurs during connection setup, and a client
// connection write error occurs while trying to report that error
// back to a client.
func TestCannotGetSession(t *testing.T) {
	req := require.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.Default()
	cfg.MongoDB.Net.URI = "mongodb://invalid:12345"

	sp, err := provider.NewSqldSessionProvider(cfg)
	req.NoError(err)
	defer sp.Close()

	srv, err := New(ctx, cancel, nil, sp, cfg)
	req.NoError(err)

	conn := newMockNetConn("host:port")
	conn.writeErr = fmt.Errorf("failed to write to mock net conn")

	// Prior to the bugfix, this function would have panicked. Now, it
	// should exit successfully.
	srv.serveConnection(ctx, conn)
}

func readHelper(t *testing.T, packets []byte, expLen int) []byte {

	s := newMockServer()
	c := newMockConn(s)

	reader := bytes.NewReader(packets)
	c.reader = reader

	lenRead := 0
	headers := 0
	result := make([]byte, 0)

	for lenRead != expLen {

		data, err := c.readPacket()

		if err != nil {
			t.Fatal(err.Error())
		}

		if len(data) == 0 {
			break
		}

		lenRead += len(data)
		headers += 4
		result = append(result, data...)

	}

	expBytesRecieved := uint64(lenRead + headers)

	if c.bytesReceived != expBytesRecieved {
		t.Fatal(fmt.Sprintf("c.bytesReceived updated incorrectly, expected %d and saw %d",
			expBytesRecieved, c.bytesReceived))
	}

	if *c.server.variables.BytesReceived != expBytesRecieved {
		t.Fatal(fmt.Sprintf("c.server.bytesReceived updated incorrectly, expected %d and saw %d",
			expBytesRecieved, c.server.variables.BytesReceived))
	}

	if int(c.sequence) != (headers / 4) {
		t.Fatal(fmt.Sprintf("c.sequence updated incorrectly, expected %d and saw %d", (headers / 4),
			c.compressionSequence))
	}

	return result
}

// compressHelper compresses dataPacket and checks state variables
func writeHelper(t *testing.T, data []byte) []byte {

	s := newMockServer()
	c := newMockConn(s)

	var b bytes.Buffer
	writer := &b
	c.writer = writer

	err := c.writePacket(data)

	if err != nil {
		t.Fatal(err.Error())
	}

	expBytesSent := uint64(len(b.Bytes()))

	if c.bytesSent != expBytesSent {
		t.Fatal(fmt.Sprintf("c.bytesSent updated incorrectly, expected %d and saw %d", expBytesSent,
			c.bytesSent))
	}

	if *c.server.variables.BytesSent != expBytesSent {
		t.Fatal(fmt.Sprintf("c.server.bytesSent updated incorrectly, expected %d and saw %d",
			expBytesSent, c.server.variables.BytesSent))
	}

	return b.Bytes()
}

func TestReaderAndWriter(t *testing.T) {

	tests := []struct {
		data []byte
		desc string
	}{
		{data: make([]byte, 33554430),
			desc: "33554430 bytes",
		},
		{data: make([]byte, 16777215),
			desc: "16777215 bytes",
		},
		{data: make([]byte, 20000000),
			desc: "20 million bytes",
		},
		{data: make([]byte, 40000000),
			desc: "40 million bytes",
		},
		{data: []byte("a"),
			desc: "a"},
		{data: []byte{0},
			desc: "0 byte"},
		{data: []byte("hello world"),
			desc: "hello world"},
		{data: make([]byte, 100),
			desc: "100 bytes"},
		{data: make([]byte, 32768),
			desc: "32768 bytes"},
		{data: make([]byte, 330000),
			desc: "33000 bytes"},
		{data: make([]byte, 10),
			desc: "10 bytes",
		},
		{data: make([]byte, 111111111),
			desc: "111111111 bytes",
		},
	}

	for _, test := range tests {
		s := fmt.Sprintf("Test writePacket with %s", test.desc)
		t.Run(s, func(t *testing.T) {

			dataWithHeader := append(make([]byte, 4), test.data...)
			packets := writeHelper(t, dataWithHeader)

			data := readHelper(t, packets, len(test.data))

			if !bytes.Equal(data, test.data) {
				t.Fatal("ReadPacket or WritePacket Failed")
			}
		})
	}
}

func newMockNetConn(addr string) *mockNetConn {
	return &mockNetConn{
		dAddr:    &mockAddr{addr},
		writeErr: nil,
	}
}

type mockAddr struct {
	addrString string
}

func (d *mockAddr) String() string {
	return d.addrString
}

type mockNetConn struct {
	dAddr    *mockAddr
	writeErr error
}

func (d *mockNetConn) RemoteAddr() net.Addr {
	return d.dAddr
}

func (d *mockNetConn) Write(b []byte) (n int, err error) {
	return -1, d.writeErr
}

func TestUserStringFunction(t *testing.T) {
	c := &conn{}
	c.user = "user"

	nconn := newMockNetConn("host:port")
	c.conn = nconn
	if c.user != "user" {
		t.Fatal("User func should return exactly the user")
	}

	if c.remoteHost() != "host" {
		t.Fatal("RemoteHost func should return exactly the host")
	}

	nconn = newMockNetConn("")
	c.conn = nconn
	if c.remoteHost() != "localhost" {
		t.Fatal("RemoteHost func should return localhost if no host is provided")
	}

}

// These functions don't do anything and are here just to satisfy net.Conn and net.Addr interfaces
func (*mockAddr) Network() string                       { return "" }
func (*mockNetConn) Read(b []byte) (n int, err error)   { return -1, nil }
func (*mockNetConn) Close() error                       { return nil }
func (*mockNetConn) LocalAddr() net.Addr                { return nil }
func (*mockNetConn) SetDeadline(t time.Time) error      { return nil }
func (*mockNetConn) SetReadDeadline(t time.Time) error  { return nil }
func (*mockNetConn) SetWriteDeadline(t time.Time) error { return nil }
