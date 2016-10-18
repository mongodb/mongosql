package catalog

import (
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

// NewInMemoryTable creates a new InMemoryTable.
func NewInMemoryTable(name string, tableType TableType, columns ...*InMemoryColumn) *InMemoryTable {
	return &InMemoryTable{
		name:    TableName(name),
		columns: columns,
	}
}

// InMemoryTable is an in-memory table.
type InMemoryTable struct {
	name      TableName
	columns   []*InMemoryColumn
	tableType TableType
	Rows      []*DataRow
}

// Name gets the name for the Table.
func (t *InMemoryTable) Name() TableName {
	return t.name
}

// Collation gets the collation for the Table.
func (t *InMemoryTable) Collation() *collation.Collation {
	return collation.Default
}

// Column gets the column of the specified name.
func (t *InMemoryTable) Column(name string) (Column, error) {
	for _, c := range t.columns {
		if strings.ToLower(name) == strings.ToLower(string(c.name)) {
			return c, nil
		}
	}

	return nil, mysqlerrors.Defaultf(mysqlerrors.ER_BAD_FIELD_ERROR, name, string(t.Name()))
}

// Columns gets the columns for the Table.
func (t *InMemoryTable) Columns() []Column {
	var cols []Column
	for _, c := range t.columns {
		cols = append(cols, c)
	}
	return cols
}

// Comments are comments about the table.
func (t *InMemoryTable) Comments() string {
	return ""
}

// Type is the type of the table.
func (t *InMemoryTable) Type() TableType {
	return BaseTable
}

// AddColumn adds a column to the DynamicTable.
func (t *InMemoryTable) AddColumn(name string, sqlType schema.SQLType) (*InMemoryColumn, error) {
	for _, c := range t.columns {
		if strings.ToLower(name) == strings.ToLower(string(c.name)) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ER_DUP_FIELDNAME, name)
		}
	}

	c := &InMemoryColumn{
		name:    ColumnName(name),
		sqlType: sqlType,
	}

	t.columns = append(t.columns, c)

	return c, nil
}

// Insert inserts a row into the table.
func (t *InMemoryTable) Insert(values ...interface{}) {
	t.Rows = append(t.Rows, &DataRow{Values: values})
}

// InMemoryColumn is an in-memory table column.
type InMemoryColumn struct {
	comments string
	name     ColumnName
	sqlType  schema.SQLType
}

// Name gets the name of the column.
func (c *InMemoryColumn) Name() ColumnName {
	return c.name
}

// PrimaryKey indicates whether this column is part of the primary key.
func (c *InMemoryColumn) PrimaryKey() bool {
	return false
}

// Type gets the type of the column.
func (c *InMemoryColumn) Type() schema.SQLType {
	return c.sqlType
}

// Comments gets the comments for the column.
func (c *InMemoryColumn) Comments() string {
	return c.comments
}
