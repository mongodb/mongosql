package catalog

import (
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

// NewDynamicTable creates a new DynamicTable.
func NewDynamicTable(name string, tableType TableType, generator func() []*DataRow) *DynamicTable {
	return &DynamicTable{
		name:      TableName(name),
		tableType: tableType,
		generator: generator,
	}
}

// DynamicTable is a table that gets its data dynamically.
type DynamicTable struct {
	name      TableName
	columns   []*DynamicColumn
	tableType TableType
	generator func() []*DataRow
}

// Name gets the name for the Table.
func (t *DynamicTable) Name() TableName {
	return t.name
}

// Collation gets the collation for the Table.
func (t *DynamicTable) Collation() *collation.Collation {
	return collation.Default
}

// Column gets the column of the specified name.
func (t *DynamicTable) Column(name string) (Column, error) {
	for _, c := range t.columns {
		if strings.ToLower(name) == strings.ToLower(string(c.name)) {
			return c, nil
		}
	}

	return nil, mysqlerrors.Defaultf(mysqlerrors.ER_BAD_FIELD_ERROR, name, string(t.Name()))
}

// Columns gets the columns for the Table.
func (t *DynamicTable) Columns() []Column {
	var cols []Column
	for _, c := range t.columns {
		cols = append(cols, c)
	}
	return cols
}

// Comments are comments about the table.
func (t *DynamicTable) Comments() string {
	return ""
}

// Type is the type of the table.
func (t *DynamicTable) Type() TableType {
	return t.tableType
}

// AddColumn adds a column to the DynamicTable.
func (t *DynamicTable) AddColumn(name string, sqlType schema.SQLType) (*DynamicColumn, error) {
	for _, c := range t.columns {
		if strings.ToLower(name) == strings.ToLower(string(c.name)) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ER_DUP_FIELDNAME, name)
		}
	}

	c := &DynamicColumn{
		name:    ColumnName(name),
		sqlType: sqlType,
	}

	t.columns = append(t.columns, c)

	return c, nil
}

// OpenReader opens a DataReader to enumerate over the .
func (t *DynamicTable) OpenReader() (DataReader, error) {
	return &dataRowSliceReader{
		rows: t.generator(),
	}, nil
}

// DynamicColumn is a column for a DynamicTable.
type DynamicColumn struct {
	comments string
	name     ColumnName
	sqlType  schema.SQLType
}

// Name gets the name of the column.
func (c *DynamicColumn) Name() ColumnName {
	return c.name
}

// Type gets the type of the column.
func (c *DynamicColumn) Type() schema.SQLType {
	return c.sqlType
}

// Comments gets the comments for the column.
func (c *DynamicColumn) Comments() string {
	return c.comments
}
