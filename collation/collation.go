package collation

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/mysqlerrors"
)

func init() {
	collationByID = make(map[ID]*Collation, 0)
	collationByName = make(map[Name]*Collation, 0)

	for _, collation := range collations {
		collationByName[collation.Name] = collation
		collationByID[collation.ID] = collation
	}
}

var collationByID map[ID]*Collation
var collationByName map[Name]*Collation

// Name is the name of a collation.
type Name string

// ID is the identifier of a collation.
type ID uint8

// Collation defines a collation.
type Collation struct {
	// SortLen is related to the amount of memory
	// required to sort strings expressed in
	// collation's character set.
	SortLen uint8
	// Name is the name of the collation.
	Name Name
	// ID is the id of the collation.
	ID ID
	// Default indicates if this collation is
	// the default for its characters.
	Default bool
	// Charset is the charset for the collation.
	Charset *Charset
}

// GetAll gets all available collations.
func GetAll() []*Collation {
	allCollations := []*Collation{}

	for _, collation := range collations {
		c, err := Get(collation.Name)
		if err != nil {
			// certain character sets aren't supported
			// thus collations that rely on such sets
			// are omitted
			continue
		}
		allCollations = append(allCollations, c)
	}

	return allCollations
}

// Get gets a collation by its name.
func Get(name Name) (*Collation, error) {
	collation, ok := collationByName[name]
	if !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_COLLATION, fmt.Sprintf("name(%v)", name))
	}

	if collation.Charset != nil {
		return collation, nil
	}

	parts := strings.SplitN(string(collation.Name), "_", 2)
	cs, err := GetCharset(CharsetName(parts[0]))
	if err != nil {
		return nil, err
	}

	collation.Charset = cs
	return collation, nil
}

// GetByID gets a collation by its id.
func GetByID(id ID) (*Collation, error) {
	collation, ok := collationByID[id]
	if !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_COLLATION, fmt.Sprintf("id(%v)", id))
	}

	if collation.Charset != nil {
		return collation, nil
	}

	parts := strings.SplitN(string(collation.Name), "_", 2)
	cs, err := GetCharset(CharsetName(parts[0]))
	if err != nil {
		return nil, err
	}

	collation.Charset = cs
	return collation, nil
}

// Must gets a Collation or panics.
func Must(cs *Collation, err error) *Collation {
	if err != nil {
		panic("internal error: " + err.Error())
	}

	return cs
}
