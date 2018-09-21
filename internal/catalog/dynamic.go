package catalog

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/variable"
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

// DynamicTable is a table that returns its data dynamically.
type DynamicTable struct {
	name      TableName
	columns   []*DynamicColumn
	tableType TableType
	generator func() []*DataRow
}

// Name returns the name for the DynamicTable, t.
func (t *DynamicTable) Name() TableName {
	return t.name
}

// Collation returns the collation for the DynamicTable, t.
func (t *DynamicTable) Collation() *collation.Collation {
	return collation.Default
}

// Column returns the column of the specified name.
func (t *DynamicTable) Column(name string) (Column, error) {
	for _, c := range t.columns {
		if strings.ToLower(name) == strings.ToLower(string(c.name)) {
			return c, nil
		}
	}

	return nil, mysqlerrors.Defaultf(mysqlerrors.ErBadFieldError, name, string(t.Name()))
}

// Columns returns the columns in the DynamicTable, t.
func (t *DynamicTable) Columns() []Column {
	var cols []Column
	for _, c := range t.columns {
		cols = append(cols, c)
	}
	return cols
}

// Comments are comments about the DynamicTable, t.
func (t *DynamicTable) Comments() string {
	return ""
}

// PrimaryKeys returns the primary keys for
// the DynamicTable, t.
func (t *DynamicTable) PrimaryKeys() []Column {
	return nil
}

// ForeignKeys returns the foreign keys for the DynamicTable, t.
func (t *DynamicTable) ForeignKeys() []ForeignKey {
	return nil
}

// Indexes returns the indexes for the DynamicTable, t.
func (t *DynamicTable) Indexes() []Index {
	return nil
}

// Type is the type of the DynamicTable, t.
func (t *DynamicTable) Type() TableType {
	return t.tableType
}

// AddColumn adds a column to the DynamicTable, t.
func (t *DynamicTable) AddColumn(name string, sqlType schema.SQLType) (*DynamicColumn, error) {
	for _, c := range t.columns {
		if strings.ToLower(name) == strings.ToLower(string(c.name)) {
			return nil, mysqlerrors.Defaultf(mysqlerrors.ErDupFieldname, name)
		}
	}

	c := &DynamicColumn{
		name:    ColumnName(name),
		sqlType: sqlType,
	}

	t.columns = append(t.columns, c)

	return c, nil
}

// AddColumns is a helper function for combining multiple calls to AddColumn.
func (t *DynamicTable) AddColumns(args ...string) {
	if len(args)%2 != 0 {
		panic(fmt.Errorf("must provide an even number of arguments"))
	}

	var idx int
	for idx < len(args) {
		name := args[idx]
		sqlType, err := schema.GetSQLType(args[idx+1])
		if err != nil {
			panic(err)
		}

		_, err = t.AddColumn(name, sqlType)
		if err != nil {
			panic(err)
		}

		idx += 2
	}
}

// OpenReader opens a DataReader to enumerate over the
// t's generated rows.
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

// ShouldConvert always returns false, as data in dynamic
// columns is never polymorphic.
func (c *DynamicColumn) ShouldConvert(_ variable.PolymorphicTypeConversionModeType) bool {
	return false
}

// Name returns the name of the column.
func (c *DynamicColumn) Name() ColumnName {
	return c.name
}

// Type returns the type of the column.
func (c *DynamicColumn) Type() schema.SQLType {
	return c.sqlType
}

// Comments returns the comments for the column.
func (c *DynamicColumn) Comments() string {
	return c.comments
}
