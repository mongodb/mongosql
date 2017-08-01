package server

import (
	"bytes"
	"fmt"
	"testing"
)

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
		t.Fatal(fmt.Sprintf("c.bytesReceived updated incorrectly, expected %d and saw %d", expBytesRecieved, c.bytesReceived))
	}

	if c.server.bytesReceived != expBytesRecieved {
		t.Fatal(fmt.Sprintf("c.server.bytesReceived updated incorrectly, expected %d and saw %d", expBytesRecieved, c.server.bytesReceived))
	}

	if int(c.sequence) != (headers / 4) {
		t.Fatal(fmt.Sprintf("c.sequence updated incorrectly, expected % and saw %d", (headers / 4), c.compressionSequence))
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
		t.Fatal(fmt.Sprintf("c.bytesSent updated incorrectly, expected %d and saw %d", expBytesSent, c.bytesSent))
	}

	if c.server.bytesSent != expBytesSent {
		t.Fatal(fmt.Sprintf("c.server.bytesSent updated incorrectly, expected %d and saw %d", expBytesSent, c.server.bytesSent))
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

			if bytes.Compare(data, test.data) != 0 {
				t.Fatal("ReadPacket or WritePacket Failed")
			}
		})
	}
}
