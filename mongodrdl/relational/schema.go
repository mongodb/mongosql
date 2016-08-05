package relational

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"sort"
	"strings"
)

const (
	MongoFilterMongoTypeName = "mongo.Filter"
)

type Database struct {
	Name   string     `yaml:"db"`
	Tables TableSlice `yaml:"tables"`
}

func NewDatabase(name string) *Database {
	return &Database{
		Name: name,
	}
}

func (d *Database) AddTable(name string, collectionName string) (*Table, error) {
	for _, t := range d.Tables {
		if t.Name == name {
			return nil, fmt.Errorf("A table with the name %s already exists.", name)
		}
	}

	t := newTable(name, collectionName)
	d.Tables = append(d.Tables, t)
	d.Tables.Sort()
	return t, nil
}

// +++++++++++++++++++++
type Table struct {
	Name           string                   `yaml:"table"`
	CollectionName string                   `yaml:"collection"`
	Pipeline       []map[string]interface{} `yaml:"pipeline"`
	Columns        ColumnSlice              `yaml:"columns"`
	Parent         *Table                   `yaml:"-"`
}

func newTable(name string, collectionName string) *Table {
	return &Table{
		Name:           name,
		CollectionName: collectionName,
	}
}

func (t *Table) AddColumn(name string, mongoType string) (*Column, error) {
	if t.Columns.containsName(name) {
		return nil, fmt.Errorf("A column with the name %s already exists.", name)
	}

	c := newColumn(name, mongoType)
	t.Columns = append(t.Columns, c)
	t.Columns.Sort()
	return c, nil
}

func (t *Table) addUnwind(path string, indexFieldName string) {
	unwind := bson.M{
		"$unwind": bson.M{
			"path":              "$" + path,
			"includeArrayIndex": indexFieldName,
		},
	}

	t.Pipeline = append(t.Pipeline, unwind)
}

func (t *Table) copyParent() error {
	if t.Parent == nil {
		return nil
	}

	for _, copy := range t.Parent.Columns {
		if t.Columns.containsName(copy.Name) {
			return fmt.Errorf("A column with the name %s already exists.", copy.Name)
		}

		t.Columns = append(t.Columns, copy)
	}

	t.Columns.Sort()
	t.Pipeline = append(t.Parent.Pipeline, t.Pipeline...)
	return nil
}

func (t *Table) rootName() string {
	if t.Parent != nil {
		return t.Parent.rootName()
	}

	return t.Name
}

// +++++++++++++++++++++
type TableSlice []*Table

func (slice TableSlice) Len() int           { return len(slice) }
func (slice TableSlice) Less(i, j int) bool { return slice[i].Name < slice[j].Name }
func (slice TableSlice) Swap(i, j int)      { slice[i], slice[j] = slice[j], slice[i] }
func (slice TableSlice) Sort()              { sort.Sort(slice) }

// +++++++++++++++++++++
type Column struct {
	Name      string
	MongoType string
	SqlName   string
	SqlType   string
}

func isNumeric(mongoType string) bool {
	return strings.HasPrefix(mongoType, "int") ||
		strings.HasPrefix(mongoType, "float") ||
		mongoType == "bson.Decimal128"
}

func newColumn(name string, mongoType string) *Column {
	var sqlType string
	switch {
	case isNumeric(mongoType):
		sqlType = "numeric"
	case mongoType == "bool":
		sqlType = "boolean"
	case mongoType == "date":
		sqlType = "timestamp"
	case mongoType == "geo.2darray":
		sqlType = "numeric[]"
	default:
		sqlType = "varchar"
	}

	c := &Column{
		Name:      name,
		MongoType: mongoType,
		SqlName:   name,
		SqlType:   sqlType,
	}

	return c
}

// +++++++++++++++++++++
type ColumnSlice []*Column

func (slice ColumnSlice) Len() int { return len(slice) }
func (slice ColumnSlice) Less(i, j int) bool {
	// synthetic should always sort last
	if slice[i].MongoType == MongoFilterMongoTypeName && slice[j].MongoType != MongoFilterMongoTypeName {
		return false
	} else if slice[j].MongoType == MongoFilterMongoTypeName && slice[i].MongoType != MongoFilterMongoTypeName {
		return true
	}
	return slice[i].Name < slice[j].Name
}
func (slice ColumnSlice) Swap(i, j int) { slice[i], slice[j] = slice[j], slice[i] }
func (slice ColumnSlice) Sort()         { sort.Sort(slice) }
func (slice ColumnSlice) containsName(name string) bool {
	for _, c := range slice {
		if c.Name == name {
			return true
		}
	}
	return false
}
