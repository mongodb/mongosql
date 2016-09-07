package collation

import (
	"github.com/10gen/sqlproxy/mysqlerrors"
	"golang.org/x/text/encoding"
)

// CharsetName is the name of a character set.
type CharsetName string

// Charset is a set of characters.
type Charset struct {
	Name               CharsetName
	DefaultCollationID ID

	decoder *encoding.Decoder
	encoder *encoding.Encoder
}

// Decode converts the given encoded bytes to UTF-8. It returns the converted bytes or nil, err if any error occurred.
func (cs *Charset) Decode(bytes []byte) []byte {
	// we are skipping errors because, according to research, errors should never be raised. In addition,
	// we don't ever want to fail during reading because of a failed encoding. We'll just read gibberish instead.
	b, _ := cs.decoder.Bytes(bytes)
	return b
}

// Encode converts the given bytes from UTF-8. It returns the converted bytes or nil, err if any error occurred.
func (cs *Charset) Encode(bytes []byte) []byte {
	// we are skipping errors because, according to research, errors should never be raised. In addition,
	// we don't ever want to fail during writing because of a failed encoding. We'll just write gibberish instead.
	b, _ := cs.encoder.Bytes(bytes)
	return b
}

// GetCharset gets the character set for the specified name.
func GetCharset(s CharsetName) (*Charset, error) {

	var e encoding.Encoding
	var collationID ID
	if s == "" {
		e = encoding.Nop
		collationID = 0
	} else {
		var ok bool
		e, ok = charsetEncodings[s]
		if !ok {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_CHARACTER_SET, s)
		}

		collationID, ok = charsets[s]
		if !ok {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_CHARACTER_SET, s)
		}
	}

	return &Charset{
		Name:               CharsetName(s),
		DefaultCollationID: collationID,
		decoder:            e.NewDecoder(),
		encoder:            encoding.ReplaceUnsupported(e.NewEncoder()),
	}, nil
}

// MustCharset gets a Charset or panics.
func MustCharset(cs *Charset, err error) *Charset {
	if err != nil {
		panic("internal error: " + err.Error())
	}

	return cs
}
