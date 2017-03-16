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
			name:      ColumnName(c.SqlName),
			sqlType:   c.SqlType,
			MongoName: c.Name,
			MongoType: c.MongoType,
			comments:  fmt.Sprintf(`{ "name": "%s" }`, c.Name),
		}
		columns = append(columns, mc)
		if isPrimaryKey(t, mc.MongoName) {
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

// MongoTable is a table whose data comes from elsewhere.
type MongoTable struct {
	name        TableName
	collation   *collation.Collation
	columns     []*MongoColumn
	primaryKeys []Column
	comments    string
	tableType   TableType

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

// PrimaryKeys returns the primary keys for
// the table.
func (t *MongoTable) PrimaryKeys() []Column {
	return t.primaryKeys
}

// Type is the type of the table.
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

func isPrimaryKey(t *schema.Table, mongoName string) bool {
	if mongoName == "_id" {
		return true
	}

	for _, d := range t.Pipeline {
		unwindVal, ok := d.Map()["$unwind"]
		if !ok {
			return false
		}

		unwind, ok := unwindVal.(bson.D)
		if !ok {
			return false
		}

		arrayIndexNameVal, ok := unwind.Map()["includeArrayIndex"]
		if !ok {
			continue
		}

		arrayIndexName, ok := arrayIndexNameVal.(string)
		if !ok {
			continue
		}

		if mongoName == arrayIndexName {
			return true
		}
	}

	return false
}
