package catalog

import (
	"fmt"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

// NewMongoTable creates a new MongoTable.
func NewMongoTable(t *schema.Table, tableType TableType, collation *collation.Collation) *MongoTable {
	var columns []*MongoColumn
	var primaryKeys []Column
	for _, c := range t.Columns {
		mc := &MongoColumn{
			name:      ColumnName(c.SQLName),
			sqlType:   c.SQLType,
			MongoName: c.Name,
			MongoType: c.MongoType,
			comments:  fmt.Sprintf(`{ "name": "%s" }`, c.Name),
		}
		columns = append(columns, mc)
		if t.IsPrimaryKey(mc.MongoName) {
			primaryKeys = append(primaryKeys, mc)
		}
	}

	return &MongoTable{
		name:           TableName(t.Name),
		collation:      collation,
		columns:        columns,
		tableType:      tableType,
		primaryKeys:    primaryKeys,
		CollectionName: t.CollectionName,
		Pipeline:       t.Pipeline,
		comments:       fmt.Sprintf(`{ "collectionName": "%s" }`, t.CollectionName),
	}
}

// MongoTable is a table whose data comes from a MongoDB collection.
type MongoTable struct {
	name           TableName
	collation      *collation.Collation
	columns        []*MongoColumn
	primaryKeys    []Column
	indexes        []Index
	foreignKeys    []ForeignKey
	comments       string
	tableType      TableType
	isSharded      bool
	CollectionName string
	Pipeline       []bson.D
}

// Name is the name of the MongoTable, t.
func (t *MongoTable) Name() TableName {
	return t.name
}

// IsSharded returns true if the MongoTable,
// t is in a sharded collection.
func (t *MongoTable) IsSharded() bool {
	return t.isSharded
}

// Collation gets the collation for the MongoTable, t.
func (t *MongoTable) Collation() *collation.Collation {
	return t.collation
}

// Column gets the column of the specified name.
func (t *MongoTable) Column(name string) (Column, error) {
	for _, c := range t.columns {
		if strings.ToLower(name) == strings.ToLower(string(c.name)) {
			return c, nil
		}
	}

	return nil, mysqlerrors.Defaultf(mysqlerrors.ER_BAD_FIELD_ERROR, name, string(t.Name()))
}

// Columns returns the columns in MongoTable, t.
func (t *MongoTable) Columns() []Column {
	var cols []Column
	for _, c := range t.columns {
		cols = append(cols, c)
	}
	return cols
}

// Comments are the comments for the MongoTable, t.
func (t *MongoTable) Comments() string {
	return t.comments
}

// ForeignKeys returns the foreign keys for the MongoTable, t.
func (t *MongoTable) ForeignKeys() []ForeignKey {
	return t.foreignKeys
}

// Indexes returns the indexes for the MongoTable, t.
func (t *MongoTable) Indexes() []Index {
	return t.indexes
}

// PrimaryKeys returns the primary keys for
// the MongoTable, t.
func (t *MongoTable) PrimaryKeys() []Column {
	return t.primaryKeys
}

// Type returns the type of the MongoTable, t.
func (t *MongoTable) Type() TableType {
	return t.tableType
}

// MongoColumn is an mongo table column.
type MongoColumn struct {
	comments string
	name     ColumnName
	sqlType  schema.SQLType

	MongoName string
	MongoType schema.MongoType
}

// Name returns the name of the column.
func (c *MongoColumn) Name() ColumnName {
	return c.name
}

// Comments returns the comments for the column.
func (c *MongoColumn) Comments() string {
	return c.comments
}

// Type returns the SQLType of the column, c.
func (c *MongoColumn) Type() schema.SQLType {
	return c.sqlType
}
