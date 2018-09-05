package catalog

import (
	"strings"

	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/variable"
	"github.com/10gen/sqlproxy/schema"
)

// NewInMemoryTable creates a new InMemoryTable.
func NewInMemoryTable(name string, columns ...*InMemoryColumn) *InMemoryTable {
	return &InMemoryTable{
		name:    TableName(name),
		columns: columns,
	}
}

// InMemoryTable is an in-memory table.
type InMemoryTable struct {
	name    TableName
	columns []*InMemoryColumn
	Rows    []*DataRow
}

// AddColumn adds a columns to the InMemoryTable, t.
func (t *InMemoryTable) AddColumn(name string, sqlType schema.SQLType) (*InMemoryColumn, error) {
	for _, c := range t.columns {
		if strings.ToLower(name) == strings.ToLower(string(c.name)) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErDupFieldname, name)
		}
	}

	c := &InMemoryColumn{
		name:    ColumnName(name),
		sqlType: sqlType,
	}

	t.columns = append(t.columns, c)

	return c, nil
}

// Collation returns the collation for the InMemoryTable, t.
func (t *InMemoryTable) Collation() *collation.Collation {
	return collation.Default
}

// Column returns the column of the specified name.
func (t *InMemoryTable) Column(name string) (Column, error) {
	for _, c := range t.columns {
		if strings.ToLower(name) == strings.ToLower(string(c.name)) {
			return c, nil
		}
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

// Comments returns comments about the InMemoryTable, t.
func (t *InMemoryTable) Comments() string {
	return ""
}

// ForeignKeys returns the foreign keys for the InMemoryTable, t.
func (t *InMemoryTable) ForeignKeys() []ForeignKey {
	return nil
}

// Indexes returns the indexes for the InMemoryTable, t.
func (t *InMemoryTable) Indexes() []Index {
	return nil
}

// Insert inserts a row into the InMemoryTable, t.
func (t *InMemoryTable) Insert(values ...interface{}) {
	t.Rows = append(t.Rows, &DataRow{Values: values})
}

// Name returns the name for the InMemoryTable, t.
func (t *InMemoryTable) Name() TableName {
	return t.name
}

// PrimaryKeys returns the primary keys for
// the InMemoryTable, t.
func (t *InMemoryTable) PrimaryKeys() []Column {
	return nil
}

// Type returns the type of the InMemoryTable, t.
func (t *InMemoryTable) Type() TableType {
	return BaseTable
}

// InMemoryColumn is an in-memory table column.
type InMemoryColumn struct {
	comments string
	name     ColumnName
	sqlType  schema.SQLType
}

// ShouldConvert always returns false, as data in memory
// columns is never polymorphic.
func (c *InMemoryColumn) ShouldConvert(_ variable.PolymorphicTypeConversionModeType) bool {
	return false
}

// Name returns the name of the column.
func (c *InMemoryColumn) Name() ColumnName {
	return c.name
}

// Type returns the type of the column.
func (c *InMemoryColumn) Type() schema.SQLType {
	return c.sqlType
}

// Comments returns the comments for the column.
func (c *InMemoryColumn) Comments() string {
	return c.comments
}
