package server

// DOCUMENTATION:
// https://web.archive.org/web/20170404163346/https://dev.mysql.com/doc/internals/en/compression.html

import (
	"bytes"
	"compress/zlib"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"io"
	"sync/atomic"
)

const (
	minCompressLength = 50
)

type compressedReader struct {
	buf        []byte
	connReader io.Reader
	c          *conn
}

type compressedWriter struct {
	connWriter io.Writer
	c          *conn
}

func NewCompressedReader(connReader io.Reader, c *conn) *compressedReader {
	return &compressedReader{
		buf:        make([]byte, 0),
		connReader: connReader,
		c:          c,
	}
}

func NewCompressedWriter(connWriter io.Writer, c *conn) *compressedWriter {
	return &compressedWriter{
		connWriter: connWriter,
		c:          c,
	}
}

func (cr *compressedReader) Read(data []byte) (int, error) {
	// while buffer is not large enough, read in more packets
	for len(cr.buf) < len(data) {
		err := cr.uncompressPacket()
		if err != nil {
			return 0, err
		}
	}

	copy(data, cr.buf[:len(data)])

	cr.buf = cr.buf[len(data):]
	return len(data), nil
}

func (cr *compressedReader) uncompressPacket() error {
	header := []byte{0, 0, 0, 0, 0, 0, 0}

	if _, err := io.ReadFull(cr.connReader, header); err != nil {
		return err
	}

	// compressed header structure
	comprLength := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)
	uncompressedLength := int(uint32(header[4]) | uint32(header[5])<<8 | uint32(header[6])<<16)
	compressionSequence := uint8(header[3])

	if compressionSequence != cr.c.compressionSequence {
		return mysqlerrors.Defaultf(mysqlerrors.ER_NET_PACKETS_OUT_OF_ORDER)
	}

	// packets from the client should not be larger than maxPayloadLength
	// when going over the wire
	if comprLength > maxPayloadLength {
		return mysqlerrors.Defaultf(mysqlerrors.ER_NET_PACKET_TOO_LARGE)
	}

	cr.c.compressionSequence++

	b := &bytes.Buffer{}
	_, err := io.CopyN(b, cr.connReader, int64(comprLength))

	if err != nil {
		return err
	}

	// if payload is uncompressed, its length will be specified as zero, and its
	// true length is contained in comprLength
	if uncompressedLength != 0 {
		// write comprData to a bytes.buffer, then read it using zlib into data

		r, err := zlib.NewReader(b)

		if r != nil {
			defer r.Close()
		}

		if err != nil {
			return err
		}

		data := make([]byte, uncompressedLength)
		lenRead := 0

		// http://grokbase.com/t/gg/golang-nuts/146y9ppn6b/go-nuts-stream-compression-with-compress-flate
		for lenRead < uncompressedLength {

			tmp := data[lenRead:]

			n, err := r.Read(tmp)
			lenRead += n

			if err == io.EOF {
				if lenRead < uncompressedLength {
					return io.ErrUnexpectedEOF
				}
				break
			}

			if err != nil {
				return err
			}
		}

		cr.buf = append(cr.buf, data...)

	} else {
		cr.buf = append(cr.buf, b.Bytes()...)
	}

	// record bytes received
	bytesReceived := uint64(comprLength) + 7 // +7 for the compression header

	atomic.AddUint64(&cr.c.bytesReceived, bytesReceived)
	atomic.AddUint64(&cr.c.server.bytesReceived, bytesReceived)

	return nil
}

func (cw *compressedWriter) Write(data []byte) (int, error) {

	// when asked to write an empty packet, do nothing
	if len(data) == 0 {
		return 0, nil
	}
	totalBytes := len(data)

	length := len(data) - 4

	for length >= maxPayloadLength {
		// cut off a slice of size max payload length
		dataSmall := data[:maxPayloadLength]
		lenSmall := len(dataSmall)

		// create a zlib writer
		// and write uncompressed packet to it
		var b bytes.Buffer
		writer := zlib.NewWriter(&b)

		_, err := writer.Write(dataSmall)
		writer.Close()

		if err != nil {
			return 0, err
		}

		err = cw.writeComprPacketToNetwork(b.Bytes(), lenSmall)
		if err != nil {
			return 0, err
		}

		length -= maxPayloadLength
		data = data[maxPayloadLength:]
	}

	lenSmall := len(data)

	// do not compress if packet is too small
	if lenSmall < minCompressLength {
		err := cw.writeComprPacketToNetwork(data, 0)
		if err != nil {
			return 0, err
		}

		return totalBytes, nil
	}

	var b bytes.Buffer
	// create a zlib writer
	// and write uncompressed packet to it
	writer := zlib.NewWriter(&b)

	_, err := writer.Write(data)
	writer.Close()

	if err != nil {
		return 0, err
	}

	// add header and send over the wire
	err = cw.writeComprPacketToNetwork(b.Bytes(), lenSmall)
	if err != nil {
		return 0, err
	}

	return totalBytes, nil
}

func (cw *compressedWriter) writeComprPacketToNetwork(data []byte, uncomprLength int) error {
	data = append([]byte{0, 0, 0, 0, 0, 0, 0}, data...)

	comprLength := len(data) - 7 // sans compressed header

	for comprLength >= maxPayloadLength {
		// compression header
		data[0] = byte(0xff & maxPayloadLength)
		data[1] = byte(0xff & (maxPayloadLength >> 8))
		data[2] = byte(0xff & (maxPayloadLength >> 16))

		data[3] = cw.c.compressionSequence

		data[4] = byte(0xff & uncomprLength)
		data[5] = byte(0xff & (uncomprLength >> 8))
		data[6] = byte(0xff & (uncomprLength >> 16))

		if _, err := cw.connWriter.Write(data[:7+maxPayloadLength]); err != nil {
			return err
		} else {
			cw.c.compressionSequence++

			nBytesSent := uint64(maxPayloadLength + 7) //max packet size
			atomic.AddUint64(&cw.c.bytesSent, nBytesSent)
			atomic.AddUint64(&cw.c.server.bytesSent, nBytesSent)

			comprLength -= maxPayloadLength
			data = data[maxPayloadLength:]
		}
	}

	// compression
	data[0] = byte(0xff & comprLength)
	data[1] = byte(0xff & (comprLength >> 8))
	data[2] = byte(0xff & (comprLength >> 16))

	data[3] = cw.c.compressionSequence

	data[4] = byte(0xff & uncomprLength)
	data[5] = byte(0xff & (uncomprLength >> 8))
	data[6] = byte(0xff & (uncomprLength >> 16))

	if _, err := cw.connWriter.Write(data); err != nil {
		return err
	}

	cw.c.compressionSequence++

	nBytesSent := uint64(len(data))
	atomic.AddUint64(&cw.c.bytesSent, nBytesSent)
	atomic.AddUint64(&cw.c.server.bytesSent, nBytesSent)

	return nil
}
