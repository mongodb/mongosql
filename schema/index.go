package schema

import (
	"strconv"
	"strings"
)

// Index represents a database index.
type Index struct {
	sqlName   string
	mongoName string
	unique    bool
	fullText  bool
	parts     []IndexPart
}

// Indexes is a type over a slice of Index so that we may define a method over it.
type Indexes []Index

// NewIndex creates an Index.
func NewIndex(name string, unique, fullText bool, parts []IndexPart) Index {
	sqlName := string(normalizeSQLName(name))
	return Index{
		sqlName:   sqlName,
		mongoName: name,
		unique:    unique,
		fullText:  fullText,
		parts:     parts,
	}
}

// createUniqueNames creates unique names for all unnamed indexes in a slice
// of indexes. This is mutating, but returns itself for convenience.
func (is Indexes) createUniqueNames() Indexes {
	names := is.collectNames()
	// If there are no empty names, we don't need to do anything.
	if _, ok := names[""]; !ok {
		return is
	}
	for i := range is {
		if is[i].sqlName == "" {
			// When creating names, just lowercase. This isn't strictly necessary
			// for MongoDB, but it is the only way to ensure there are no collisions
			// on the sql side (the collisions need to be on sql names.
			newName := makeUnique(names, strings.ToLower(is[i].makeName()))
			names[newName] = struct{}{}
			is[i].sqlName = newName
			is[i].mongoName = newName
		}
	}
	return is
}

func (is Indexes) collectNames() map[string]struct{} {
	ret := make(map[string]struct{})
	for _, index := range is {
		ret[index.sqlName] = struct{}{}
	}
	return ret
}

// makeName uses MongoDB's method of creating an index name from the keys and index types.
// e.g., {a: -1, b: 1} becomes a_-1_b_1, and {c: 'text'} becomes c_text. We use the sqlName
// because we want to ensure there are no collisions on the sql side, and sql has strictly
// more collisions since names with differing case still collide.
func (i Index) makeName() string {
	if len(i.parts) < 1 {
		panic("an index with 0 parts should not be possible")
	}
	ret := i.parts[0].sqlName
	if i.fullText {
		ret += "_text"
		for _, part := range i.parts[1:] {
			ret += "_" + part.sqlName + "_text"
		}
		return ret
	}
	ret += "_" + directionToStr(i.parts[0].Direction())
	for _, part := range i.parts[1:] {
		ret += "_" + part.sqlName + "_" + directionToStr(part.Direction())
	}
	return ret
}

func makeUnique(uniqueNames map[string]struct{}, baseName string) string {
	current := baseName
	for i := 0; ; i++ {
		if _, ok := uniqueNames[current]; !ok {
			return current
		}
		current = baseName + "_" + strconv.Itoa(i)
	}
}

// SQLName returns the sqlName.
func (i *Index) SQLName() string {
	return i.sqlName
}

// MongoName returns the mongoName.
func (i *Index) MongoName() string {
	return i.mongoName
}

// Unique returns true if this Index is unique.
func (i *Index) Unique() bool {
	return i.unique
}

// FullText returns true if this Index is FullText.
func (i *Index) FullText() bool {
	return i.fullText
}

// Parts returns the index parts.
func (i *Index) Parts() []IndexPart {
	return i.parts
}

// Direction is an enum for index part directions (ascending vs descending/1 vs -1).
type Direction int

// Possible values for Direction.
const (
	ASC  Direction = 1
	DESC Direction = -1
)

func getDirection(direction int) Direction {
	if direction >= 0 {
		return ASC
	}
	return DESC
}

// Slightly more efficient than Itoa.
func directionToStr(direction Direction) string {
	if direction == ASC {
		return "1"
	}
	return "-1"
}

// IndexPart is a part of an index: a pair of column and direction.
// e.g., for UNIQUE KEY(a ASC, b DESC) we will have the IndexParts
// {name: "a", direction: 1}, {name: "b", direction: -1}
type IndexPart struct {
	sqlName   string
	mongoName string
	direction Direction
}

// NewIndexPart creates a new IndexPart.
func NewIndexPart(name string, direction int) IndexPart {
	sqlName := string(normalizeSQLName(name))
	return IndexPart{
		sqlName:   sqlName,
		mongoName: name,
		direction: getDirection(direction),
	}
}

// MongoName returns the name of this IndexPart.
func (i *IndexPart) MongoName() string {
	return i.mongoName
}

// SQLName returns the name of this IndexPart.
func (i *IndexPart) SQLName() string {
	return i.sqlName
}

// Direction returns the direction of this IndexPart.
func (i *IndexPart) Direction() Direction {
	return i.direction
}

// DeepCopy deep copies.
func (i *Index) DeepCopy() Index {
	parts := make([]IndexPart, len(i.parts))
	for i, part := range i.parts {
		parts[i] = IndexPart{
			sqlName:   part.sqlName,
			mongoName: part.mongoName,
			direction: part.direction,
		}
	}
	ret := Index{
		sqlName:   i.sqlName,
		mongoName: i.mongoName,
		unique:    i.unique,
		fullText:  i.fullText,
		parts:     parts,
	}
	return ret
}
