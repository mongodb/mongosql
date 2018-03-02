package server

// DOCUMENTATION:
// https://goo.gl/QxzMR3

import (
	"bytes"
	"compress/zlib"
	"io"
	"sync/atomic"

	"github.com/10gen/sqlproxy/mysqlerrors"
)

const (
	minCompressLength = 50
)

// CompressedReader is a struct for reading compressed data.
type CompressedReader struct {
	buf        []byte
	connReader io.Reader
	c          *conn
	zr         io.ReadCloser
}

// CompressedWriter is a struct for writing compressed data.
type CompressedWriter struct {
	connWriter io.Writer
	c          *conn
	zw         *zlib.Writer
}

// NewCompressedReader is the public constructor for CompressedReader structs.
func NewCompressedReader(connReader io.Reader, c *conn) *CompressedReader {
	return &CompressedReader{
		buf:        make([]byte, 0),
		connReader: connReader,
		c:          c,
	}
}

// NewCompressedWriter is the public constructor for CompressedWriter structs.
func NewCompressedWriter(connWriter io.Writer, c *conn) *CompressedWriter {
	return &CompressedWriter{
		connWriter: connWriter,
		c:          c,
		zw:         zlib.NewWriter(new(bytes.Buffer)),
	}
}

// Read reads data from the CompressedReader.
func (cr *CompressedReader) Read(data []byte) (int, error) {
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

func (cr *CompressedReader) uncompressPacket() error {
	readHeader := make([]byte, 7)
	if _, err := io.ReadFull(cr.connReader, readHeader); err != nil {
		return err
	}

	// compressed header structure
	comprLength := int(uint32(readHeader[0]) | uint32(readHeader[1])<<8 |
		uint32(readHeader[2])<<16)
	uncompressedLength := int(uint32(readHeader[4]) | uint32(readHeader[5])<<8 |
		uint32(readHeader[6])<<16)
	compressionSequence := readHeader[3]

	if compressionSequence != cr.c.compressionSequence {
		return mysqlerrors.Defaultf(mysqlerrors.ErNetPacketsOutOfOrder)
	}

	// packets from the client should not be larger than maxPayloadLength
	// when going over the wire
	if comprLength > maxPayloadLength {
		return mysqlerrors.Defaultf(mysqlerrors.ErNetPacketTooLarge)
	}

	cr.c.compressionSequence++

	bytesBuf := &bytes.Buffer{}
	_, err := io.CopyN(bytesBuf, cr.connReader, int64(comprLength))
	if err != nil {
		return err
	}

	// if payload is uncompressed, its length will be specified as zero, and its
	// true length is contained in comprLength
	if uncompressedLength != 0 {
		// write comprData to a bytes.buffer, then read it using zlib into data
		if cr.zr == nil {
			cr.zr, err = zlib.NewReader(bytesBuf)
		} else {
			err = cr.zr.(zlib.Resetter).Reset(bytesBuf, nil)
		}
		if err != nil {
			return err
		}
		defer func() { _ = cr.zr.Close() }()

		data := make([]byte, uncompressedLength)
		lenRead := 0

		// https://goo.gl/mX3rGF
		for lenRead < uncompressedLength {
			n, err := cr.zr.Read(data[lenRead:])
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
		cr.buf = append(cr.buf, bytesBuf.Bytes()...)
	}

	// record bytes received
	bytesReceived := uint64(comprLength) + 7 // +7 for the compression header
	atomic.AddUint64(&cr.c.bytesReceived, bytesReceived)
	atomic.AddUint64(cr.c.server.variables.BytesReceived, bytesReceived)

	return nil
}

// Write writes data to the CompressedWriter.
func (cw *CompressedWriter) Write(data []byte) (int, error) {
	// when asked to write an empty packet, do nothing
	if len(data) == 0 {
		return 0, nil
	}

	totalBytes := len(data)
	length := len(data) - 4
	writeHeader := make([]byte, 7)

	for length >= maxPayloadLength {
		// cut off a slice of size max payload length
		payload := data[:maxPayloadLength]
		payloadLen := len(payload)
		bytesBuf := &bytes.Buffer{}
		bytesBuf.Write(writeHeader)
		cw.zw.Reset(bytesBuf)
		_, err := cw.zw.Write(payload)
		if err != nil {
			return 0, err
		}

		err = cw.zw.Close()
		if err != nil {
			return 0, err
		}

		// if compression expands the payload, do not compress
		compressedPayload := bytesBuf.Bytes()
		if len(compressedPayload) > maxPayloadLength {
			compressedPayload = append(writeHeader, payload...)
			payloadLen = 0
		}

		err = cw.writeToNetwork(compressedPayload, payloadLen)
		if err != nil {
			return 0, err
		}

		length -= maxPayloadLength
		data = data[maxPayloadLength:]
	}

	payloadLen := len(data)

	// do not attempt compression if packet is too small
	if payloadLen < minCompressLength {
		err := cw.writeToNetwork(append(writeHeader, data...), 0)
		if err != nil {
			return 0, err
		}
		return totalBytes, nil
	}

	bytesBuf := &bytes.Buffer{}
	bytesBuf.Write(writeHeader)
	cw.zw.Reset(bytesBuf)
	_, err := cw.zw.Write(data)
	if err != nil {
		return 0, err
	}

	err = cw.zw.Close()
	if err != nil {
		return 0, err
	}

	compressedPayload := bytesBuf.Bytes()

	if len(compressedPayload) > len(data) {
		compressedPayload = append(writeHeader, data...)
		payloadLen = 0
	}

	// add header and send over the wire
	err = cw.writeToNetwork(compressedPayload, payloadLen)
	if err != nil {
		return 0, err
	}

	return totalBytes, nil
}

func (cw *CompressedWriter) writeToNetwork(data []byte, uncomprLength int) error {
	totalBytesSent := uint64(0)
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
		}
		cw.c.compressionSequence++
		totalBytesSent += uint64(maxPayloadLength + 7)
		comprLength -= maxPayloadLength
		data = data[maxPayloadLength:]
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
	totalBytesSent += uint64(len(data))
	atomic.AddUint64(&cw.c.bytesSent, totalBytesSent)
	atomic.AddUint64(cw.c.server.variables.BytesSent, totalBytesSent)

	return nil
}
