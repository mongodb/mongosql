package collation

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/mysqlerrors"
)

// Name is the name of a collation.
type Name string

// ID is the identifier of a collation.
type ID uint8

// Collation defines a collation.
type Collation struct {
	// Name is the name of the collation.
	Name Name
	// ID is the id of the collation.
	ID ID
	// Charset is the charset for the collation.
	Charset *Charset
}

// Get gets a collation by its name.
func Get(name Name) (*Collation, error) {
	id, ok := collations[name]
	if !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_COLLATION, name)
	}

	parts := strings.SplitN(string(name), "_", 2)
	cs, err := GetCharset(CharsetName(parts[0]))
	if err != nil {
		return nil, err
	}

	return &Collation{
		Name:    name,
		ID:      id,
		Charset: cs,
	}, nil
}

// GetByID gets a collation from its id.
func GetByID(id ID) (*Collation, error) {
	name, ok := collationNames[id]
	if !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_COLLATION, fmt.Sprintf("id(%v)", id))
	}

	return Get(name)
}

// Must gets a Collation or panics.
func Must(cs *Collation, err error) *Collation {
	if err != nil {
		panic("internal error: " + err.Error())
	}

	return cs
}
