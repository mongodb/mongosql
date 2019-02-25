package results

import (
	"strings"

	"github.com/10gen/sqlproxy/evaluator/values"
)

// Row holds data from one or more tables.
type Row struct {
	Data RowValues
}

// GetField takes a selectID, tableName, and columnName and returns the given
// value of the given key in the row, or nil if it does not exist. The second
// return value is a boolean indicating if the field was found or not, to allow
// the distinction between a null value stored in that field from a missing
// field.
func (row *Row) GetField(selectID int,
	databaseName,
	tableName,
	columnName string) (values.SQLValue,
	bool) {
	for _, r := range row.Data {
		if r.SelectID == selectID && strings.EqualFold(r.Database, databaseName) &&
			strings.EqualFold(r.Table, tableName) && strings.EqualFold(r.Name, columnName) {
			return r.Data, true
		}
	}
	return nil, false
}

// Rows holds a slice of `Row`s.
type Rows []Row

// RowValue holds row values for a SQL column.
type RowValue struct {
	SelectID int
	Database string
	Table    string
	Name     string
	Data     values.SQLValue
}

// NewNamelessRow creates a new row from a series of values.SQLValues, giving them
// monotonically increasing selectIDs and no Database, Table, or Column names.
func NewNamelessRow(vs ...values.SQLValue) Row {
	ret := make(RowValues, len(vs))
	for i, val := range vs {
		ret[i] = NewRowValue(i+1, "", "", "", val)
	}
	return Row{
		Data: ret,
	}
}

// NewNamedRow creates a new row from a series of values.SQLValues, giving them
// monotonically increasing selectIDs
func NewNamedRow(databaseName, tableName string, vs ...values.NamedSQLValue) Row {
	ret := make(RowValues, len(vs))
	for i, val := range vs {
		ret[i] = NewRowValue(i+1, databaseName, tableName, val.Name, val.Value)
	}
	return Row{
		Data: ret,
	}
}

// NewRowValue returns a Value with the provided selectID, database, table,
// name, and data.
func NewRowValue(selectID int, database, table, name string, data values.SQLValue) RowValue {
	return RowValue{selectID, database, table, name, data}
}

// NewRowValueFromColumn generates a value from a provided Column and values.SQLValue.
func NewRowValueFromColumn(column Column, sqlValue values.SQLValue) RowValue {
	return NewRowValue(column.SelectID, column.Database, column.Table, column.Name, sqlValue)
}

// Size returns the size of the Value in bytes.
func (v *RowValue) Size() uint64 {
	s := uint64(8) // SelectID
	s += uint64(len(v.Database)) + uint64(len(v.Table)) + uint64(len(v.Name))
	if v.Data != nil {
		s += v.Data.Size()
	}

	return s
}

// RowValues holds a slice of `RowValue`s.
type RowValues []RowValue

// Map returns a map of the Values' names to their values.SQLValues.
func (v RowValues) Map() map[string]values.SQLValue {
	m := make(map[string]values.SQLValue)
	for _, value := range v {
		m[value.Name] = value.Data
	}
	return m
}

// Size returns the sum of the sizes of all the values in the slice.
func (v RowValues) Size() uint64 {
	s := uint64(0)
	for _, sv := range v {
		s += sv.Size()
	}
	return s
}
