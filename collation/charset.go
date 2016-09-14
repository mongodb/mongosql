package collation

import (
	"github.com/10gen/sqlproxy/mysqlerrors"
	"golang.org/x/text/encoding"
)

func init() {
	charsetByName = make(map[CharsetName]*Charset, 0)
	charsetByCollationName = make(map[Name]*Charset, 0)

	for _, charset := range charsets {
		charsetByCollationName[charset.DefaultCollationName] = charset
		charsetByName[charset.Name] = charset
	}
}

var charsetByName map[CharsetName]*Charset
var charsetByCollationName map[Name]*Charset

// CharsetName is the name of a character set.
type CharsetName string

// Charset is a set of characters.
type Charset struct {
	Name                 CharsetName
	DefaultCollationName Name
	Description          string
	MaxLen               uint8

	decoder *encoding.Decoder
	encoder *encoding.Encoder
}

// NullCharset is the noop charset.
var NullCharset = &Charset{
	Name:                 CharsetName(""),
	DefaultCollationName: "binary",
	Description:          "NULL",
	encoder:              encoding.ReplaceUnsupported(encoding.Nop.NewEncoder()),
	decoder:              encoding.Nop.NewDecoder(),
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

// GetAllCharsets gets all available character sets.
func GetAllCharsets() []*Charset {
	return charsets
}

// GetCharset gets the character set for the specified name.
func GetCharset(s CharsetName) (*Charset, error) {
	// ensure the character set is supported
	e, ok := charsetEncodings[s]
	if !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_CHARACTER_SET, s)
	}

	charset, ok := charsetByName[s]
	if !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_CHARACTER_SET, s)
	}

	charset.decoder = e.NewDecoder()
	charset.encoder = encoding.ReplaceUnsupported(e.NewEncoder())

	return charset, nil
}

// MustCharset gets a Charset or panics.
func MustCharset(cs *Charset, err error) *Charset {
	if err != nil {
		panic("internal error: " + err.Error())
	}

	return cs
}
