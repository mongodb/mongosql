package relational

import (
	"fmt"
	"sort"

	"github.com/10gen/mongo-go-driver/bson"
)

const (
	MongoFilterMongoTypeName = "mongo.Filter"
)

type Database struct {
	Name   string              `yaml:"db"`
	Tables TableSlice          `yaml:"tables"`
	Views  map[string]struct{} `yaml:"-"`
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
	PrimaryKey     ColumnSlice              `yaml:"-"`
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

func (t *Table) copyParent(primaryKeyOnly bool) error {
	if t.Parent == nil {
		return nil
	}

	parentPrimaryKey := make([]*Column, len(t.Parent.PrimaryKey))
	copy(parentPrimaryKey, t.Parent.PrimaryKey)
	t.PrimaryKey = append(parentPrimaryKey, t.PrimaryKey...)
	parentPrimaryKey = nil

	source := t.Parent.Columns
	if primaryKeyOnly {
		source = t.Parent.PrimaryKey
	}

	for _, copy := range source {
		if t.Columns.containsName(copy.Name) {
			return fmt.Errorf("A column with the name %s already exists.", copy.Name)
		}

		t.Columns = append(t.Columns, copy)
	}

	t.Columns.Sort()
	parentPipeline := make([]map[string]interface{}, len(t.Parent.Pipeline), len(t.Parent.Pipeline)+len(t.Pipeline))
	copy(parentPipeline, t.Parent.Pipeline)
	t.Pipeline = append(parentPipeline, t.Pipeline...)
	parentPipeline = nil
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

func newColumn(name string, mongoType string) *Column {
	var sqlType string
	switch mongoType {
	case "number":
		sqlType = "numeric"
	case "int", "int32":
		mongoType = "int"
		sqlType = "int"
	case "int64":
		sqlType = "int64"
	case "float32", "float64":
		mongoType = "float64"
		sqlType = "float64"
	case "bson.Decimal128":
		sqlType = "decimal128"
	case "bool":
		sqlType = "boolean"
	case "date":
		sqlType = "timestamp"
	case "geo.2darray":
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
