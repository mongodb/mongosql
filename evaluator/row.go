package evaluator

import (
	"strings"
)

// Row holds data from one or more tables.
type Row struct {
	Data Values
}

// GetField takes a selectID, tableName, and columnName and returns the given
// value of the given key in the row, or nil if it does not exist.  The second
// return value is a boolean indicating if the field was found or not, to allow
// the distinction between a null value stored in that field from a missing
// field.
func (row *Row) GetField(selectID int,
	databaseName,
	tableName,
	columnName string) (SQLValue,
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

// Value holds row values for a SQL column.
type Value struct {
	SelectID int
	Database string
	Table    string
	Name     string
	Data     SQLValue
}

// NewValue returns a Value with the provided selectID, database, table,
// name, and data.
func NewValue(selectID int, database, table, name string, data SQLValue) Value {
	return Value{selectID, database, table, name, data}
}

// NewValueFromColumn generates a value from a provided Column and SQLValue.
func NewValueFromColumn(column Column, sqlValue SQLValue) Value {
	return NewValue(column.SelectID, column.Database, column.Table, column.Name, sqlValue)
}

// Size returns the size of the Value in bytes.
func (v *Value) Size() uint64 {
	s := uint64(8) // SelectID
	s += uint64(len(v.Table)) + uint64(len(v.Name))
	if v.Data != nil {
		s += v.Data.Size()
	}

	return s
}

// Values holds a slice of `Value`s.
type Values []Value

// Map returns a map of the Values' names to their SQLValues.
func (v Values) Map() map[string]SQLValue {
	m := make(map[string]SQLValue)
	for _, value := range v {
		m[value.Name] = value.Data
	}
	return m
}

// Size returns the sum of the sizes of all the values in the slice.
func (v Values) Size() uint64 {
	s := uint64(0)
	for _, sv := range v {
		s += sv.Size()
	}
	return s
}
