package catalog

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"gopkg.in/mgo.v2/bson"
)

// NewMongoTable creates a new MongoTable.
func NewMongoTable(t *schema.Table, collation *collation.Collation) *MongoTable {
	var columns []*MongoColumn
	for _, c := range t.RawColumns {
		columns = append(columns, &MongoColumn{
			name:      ColumnName(c.SqlName),
			sqlType:   c.SqlType,
			MongoName: c.Name,
			MongoType: c.MongoType,
			comments:  fmt.Sprintf(`{ "name": "%s" }`, c.Name),
		})
	}

	return &MongoTable{
		name:           TableName(t.Name),
		collation:      collation,
		columns:        columns,
		CollectionName: t.CollectionName,
		Pipeline:       t.Pipeline,
		comments:       fmt.Sprintf(`{ "collectionName": "%s" }`, t.CollectionName),
	}
}

// MongoTable is a table whose data comes from elsewhere.
type MongoTable struct {
	name      TableName
	collation *collation.Collation
	columns   []*MongoColumn
	comments  string

	CollectionName string
	Pipeline       []bson.D
}

// Name is the name of the Table.
func (t *MongoTable) Name() TableName {
	return t.name
}

// Collation gets the collation for the table.
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

// Columns are the columns for the Table.
func (t *MongoTable) Columns() []Column {
	var cols []Column
	for _, c := range t.columns {
		cols = append(cols, c)
	}
	return cols
}

// Comments are the comments for the Table.
func (t *MongoTable) Comments() string {
	return t.comments
}

// Type is the type of the table.
func (t *MongoTable) Type() TableType {
	return View
}

// MongoColumn is an mongo table column.
type MongoColumn struct {
	comments string
	name     ColumnName
	sqlType  schema.SQLType

	MongoName string
	MongoType schema.MongoType
}

// Name gets the name of the column.
func (c *MongoColumn) Name() ColumnName {
	return c.name
}

// Type gets the type of the column.
func (c *MongoColumn) Type() schema.SQLType {
	return c.sqlType
}

// Comments gets the comments for the column.
func (c *MongoColumn) Comments() string {
	return c.comments
}
