package collation

import (
	"github.com/10gen/sqlproxy/mysqlerrors"
	"golang.org/x/text/encoding"
)

func init() {
	charsetByName = make(map[CharsetName]*Charset)
	charsetByCollationName = make(map[Name]*Charset)

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

	encoding encoding.Encoding
}

// DefaultCharset is the default Charset.
var DefaultCharset *Charset

// NullCharset represents the fact the the client set the character set to null, but
// ultimately reflects the use of the default collation and charset.
var NullCharset *Charset

// Decode converts the given encoded bytes to UTF-8. It returns the converted
// bytes or nil, err if any error occurred.
func (cs *Charset) Decode(bytes []byte) []byte {
	// we are skipping errors because, according to research,
	// errors should never be raised. In addition, we don't ever
	// want to fail during reading because of a failed encoding.
	// We'll just read gibberish instead.
	b, _ := cs.encoding.NewDecoder().Bytes(bytes)
	return b
}

// Encode converts the given bytes from UTF-8. It returns the converted bytes
// or nil, err if any error occurred.
func (cs *Charset) Encode(bytes []byte) []byte {
	// we are skipping errors because, according to research,
	// errors should never be raised. In addition, we don't ever
	// want to fail during writing because of a failed encoding.
	// We'll just write gibberish instead.
	b, _ := encoding.ReplaceUnsupported(cs.encoding.NewEncoder()).Bytes(bytes)
	return b
}

// GetAllCharsets gets all available character sets.
func GetAllCharsets() []*Charset {
	return charsets
}

// GetCharset gets the character set for the specified name.
func GetCharset(s CharsetName) (*Charset, error) {
	charset, ok := charsetByName[s]
	if !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErUnknownCharacterSet, s)
	}

	return charset, nil
}

// MustCharset gets a Charset or panics.
func MustCharset(cs *Charset, err error) *Charset {
	if err != nil {
		panic("internal error: " + err.Error())
	}

	return cs
}
