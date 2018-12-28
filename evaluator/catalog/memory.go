package catalog

import (
	"strings"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
)

// NewInMemoryTable creates a new InMemoryTable.
func NewInMemoryTable(name string, columns ...*InMemoryColumn) *InMemoryTable {
	var columnMap = make(map[string]*InMemoryColumn)
	for _, col := range columns {
		columnMap[strings.ToLower(string(col.Name()))] = col
	}
	return &InMemoryTable{
		name:      TableName(name),
		columns:   columns,
		columnMap: columnMap,
	}
}

// InMemoryTable is an in-memory table.
type InMemoryTable struct {
	name      TableName
	columns   []*InMemoryColumn
	columnMap map[string]*InMemoryColumn
	Rows      []*DataRow
}

// AddColumn adds a columns to the InMemoryTable, t.
func (t *InMemoryTable) AddColumn(name string, sqlType SQLType) (*InMemoryColumn, error) {
	lowerName := strings.ToLower(name)
	if _, ok := t.columnMap[lowerName]; ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErDupFieldname, name)
	}

	c := &InMemoryColumn{
		name:    ColumnName(name),
		sqlType: sqlType,
	}

	t.columns = append(t.columns, c)
	t.columnMap[lowerName] = c

	return c, nil
}

// Collation returns the collation for the InMemoryTable.
func (_ *InMemoryTable) Collation() *collation.Collation {
	return collation.Default
}

// Column returns the column of the specified name.
func (t *InMemoryTable) Column(name string) (Column, error) {
	if c, ok := t.columnMap[strings.ToLower(name)]; ok {
		return c, nil
	}
	return nil, mysqlerrors.Defaultf(mysqlerrors.ErBadFieldError, name, string(t.Name()))
}

// Columns returns the columns for the InMemoryTable, t.
func (t *InMemoryTable) Columns() []Column {
	var cols []Column
	for _, c := range t.columns {
		cols = append(cols, c)
	}
	return cols
}

// Comments returns comments about the InMemoryTable.
func (_ *InMemoryTable) Comments() string {
	return ""
}

// ForeignKeys returns the foreign keys for the InMemoryTable.
func (_ *InMemoryTable) ForeignKeys() []ForeignKey {
	return nil
}

// Indexes returns the indexes for the InMemoryTable.
func (_ *InMemoryTable) Indexes() []Index {
	return nil
}

// IsMongoTable return true if this is a table from MongoDB.
func (t *InMemoryTable) IsMongoTable() bool {
	return false
}

// Insert inserts a row into the InMemoryTable, t.
func (t *InMemoryTable) Insert(values ...interface{}) {
	t.Rows = append(t.Rows, &DataRow{Values: values})
}

// IsSharded returns false for any InMemoryTable.
func (_ *InMemoryTable) IsSharded() bool {
	return false
}

// Name returns the name for the InMemoryTable, t.
func (t *InMemoryTable) Name() TableName {
	return t.name
}

// Pipeline returns nil for any InMemoryTable.
func (_ *InMemoryTable) Pipeline() []bson.D {
	return nil
}

// PrimaryKeys returns the primary keys for
// the InMemoryTable.
func (_ *InMemoryTable) PrimaryKeys() []Column {
	return nil
}

// Type returns the type of the InMemoryTable.
func (_ *InMemoryTable) Type() TableType {
	return BaseTable
}

// InMemoryColumn is an in-memory table column.
type InMemoryColumn struct {
	comments string
	name     ColumnName
	sqlType  SQLType
}

// ShouldConvert always returns false, as data in memory
// columns is never polymorphic.
func (c *InMemoryColumn) ShouldConvert(_ string) bool {
	return false
}

// Name returns the name of the column.
func (c *InMemoryColumn) Name() ColumnName {
	return c.name
}

// Type returns the type of the column.
func (c *InMemoryColumn) Type() SQLType {
	return c.sqlType
}

// Comments returns the comments for the column.
func (c *InMemoryColumn) Comments() string {
	return c.comments
}
