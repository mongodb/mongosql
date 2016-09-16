package collation

import (
	"fmt"
	"strconv"

	"github.com/10gen/sqlproxy/mysqlerrors"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"gopkg.in/mgo.v2"
)

func init() {
	collationByID = make(map[ID]*Collation, 0)
	collationByName = make(map[Name]*Collation, 0)

	for _, collation := range collations {
		collationByName[collation.Name] = collation
		collationByID[collation.ID] = collation
	}

	Default = Must(Get("utf8_bin"))
	DefaultCharset = MustCharset(GetCharset("utf8"))
}

var collationByID map[ID]*Collation
var collationByName map[Name]*Collation

// Default is the default Collation.
var Default *Collation

// DefaultCharset is the default Charset.
var DefaultCharset *Charset

// Name is the name of a collation.
type Name string

// ID is the identifier of a collation.
type ID uint8

// Collation defines a collation.
type Collation struct {
	language         language.Tag
	ignoreCase       bool
	ignoreDiacritics bool
	ignoreWidth      bool
	collator         *collate.Collator

	// SortLen is related to the amount of memory
	// required to sort strings expressed in
	// collation's character set.
	SortLen uint8
	// Name is the name of the collation.
	Name Name
	// ID is the id of the collation.
	ID ID
	// Default indicates if this collation is the default for its character set.
	Default bool
	// CharsetName indicates the default character set for this collation.
	CharsetName CharsetName
}

func (c *Collation) ensureCollator() {
	if c.collator != nil {
		return
	}

	opts := []collate.Option{collate.OptionsFromTag(c.language)}
	if c.ignoreCase {
		opts = append(opts, collate.IgnoreCase)
	}
	if c.ignoreDiacritics {
		opts = append(opts, collate.IgnoreDiacritics)
	}
	if c.ignoreWidth {
		opts = append(opts, collate.IgnoreWidth)
	}

	c.collator = collate.New(c.language, opts...)
}

// CompareString returns an integer comparing the two strings. The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
func (c *Collation) CompareString(a, b string) int {
	c.ensureCollator()
	return c.collator.CompareString(a, b)
}

// FromMongoDB creates a Collation from an mgo.Collation.
func FromMongoDB(mc *mgo.Collation) (*Collation, error) {
	t, err := language.Parse(mc.Locale)
	if err != nil {
		return nil, err
	}

	t, err = t.SetTypeForKey("ka", mc.Alternate)
	if err != nil {
		return nil, err
	}

	t, err = t.SetTypeForKey("kb", strconv.FormatBool(mc.Backwards))
	if err != nil {
		return nil, err
	}

	t, err = t.SetTypeForKey("kc", strconv.FormatBool(mc.CaseLevel))
	if err != nil {
		return nil, err
	}

	t, err = t.SetTypeForKey("kf", mc.CaseFirst)
	if err != nil {
		return nil, err
	}

	t, err = t.SetTypeForKey("kn", strconv.FormatBool(mc.NumericOrdering))
	if err != nil {
		return nil, err
	}

	t, err = t.SetTypeForKey("kr", strconv.FormatBool(mc.Backwards))
	if err != nil {
		return nil, err
	}

	if mc.Strength > 0 {
		var value string
		switch mc.Strength {
		case 1:
			value = "level1"
		case 2:
			value = "level2"
		case 0, 3:
			value = "level3"
		case 4:
			value = "level4"
		case 5:
			value = "identic"
		}
		t, err = t.SetTypeForKey("ks", value)
		if err != nil {
			return nil, err
		}
	}

	return &Collation{
		language:         t,
		collator:         collate.New(t),
		ignoreCase:       mc.Strength < 3 && mc.CaseLevel,
		ignoreDiacritics: mc.Strength == 1,
		ignoreWidth:      mc.Strength == 3 && mc.CaseLevel,
		CharsetName:      CharsetName("utf8"),
		SortLen:          8,
	}, nil
}

// ToMongoDB creates a mgo.Collation from a Collation.
func ToMongoDB(c *Collation) *mgo.Collation {

	boolFromType := func(key string) bool {
		switch c.language.TypeForKey(key) {
		case "true":
			return true
		case "false":
			return false
		}

		return false
	}

	mc := &mgo.Collation{
		Locale: c.language.Parent().String(),
	}

	mc.Alternate = c.language.TypeForKey("ka")
	mc.Backwards = boolFromType("kb")
	mc.CaseLevel = boolFromType("kc")
	mc.CaseFirst = c.language.TypeForKey("kf")
	mc.NumericOrdering = boolFromType("kn")
	mc.Backwards = boolFromType("kr")

	switch c.language.TypeForKey("ks") {
	case "level1":
		mc.Strength = 1
	case "level2":
		mc.Strength = 2
	case "level4":
		mc.Strength = 4
	case "identic":
		mc.Strength = 5
	default:
		mc.Strength = 3
	}

	return mc
}

// GetAll gets all available collations.
func GetAll() []*Collation {
	return collations
}

// Get gets a collation by its name.
func Get(name Name) (*Collation, error) {
	collation, ok := collationByName[name]
	if !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_COLLATION, fmt.Sprintf("name(%v)", name))
	}

	return collation, nil
}

// GetByID gets a collation by its id.
func GetByID(id ID) (*Collation, error) {
	collation, ok := collationByID[id]
	if !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_UNKNOWN_COLLATION, fmt.Sprintf("id(%v)", id))
	}

	return collation, nil
}

// Must gets a Collation or panics.
func Must(cs *Collation, err error) *Collation {
	if err != nil {
		panic("internal error: " + err.Error())
	}

	return cs
}
