package evaluator

import (
	"github.com/10gen/sqlproxy/config"
	"strings"
)

// Row holds data from one or more tables.
type Row struct {
	Data []TableRow
}

// TableRow holds column data from a given table.
type TableRow struct {
	Table       string
	Values      Values
	TableConfig *config.TableConfig
}

// Value holds row value for an SQL column.
type Value struct {
	Name string
	View string
	Data interface{}
}

type Values []Value

// GetField takes a table returns the given value of the given key
// in the document, or nil if it does not exist.
// The second return value is a boolean indicating if the field was found or not, to allow
// the distinction betwen a null value stored in that field from a missing field.
// The key parameter may be a dot-delimited string to reference a field that is nested
// within a subdocument.
func (row *Row) GetField(table, name string) (interface{}, bool) {
	for _, r := range row.Data {
		if r.Table == table {
			for _, entry := range r.Values {
				// TODO optimize
				if strings.ToLower(name) == strings.ToLower(entry.Name) {
					return entry.Data, true
				}
			}
		}
	}
	return nil, false
}

// GetValues gets the values of the columns - referenced by name - in
// the row.
func (row *Row) GetValues(columns []*Column) []interface{} {
	values := make([]interface{}, 0)

	for _, column := range columns {
		value, _ := row.GetField(column.Table, column.Name)
		values = append(values, value)
	}

	return values
}

func (values Values) Map() map[string]interface{} {
	m := make(map[string]interface{}, 0)
	for _, value := range values {
		m[value.Name] = value.Data
	}
	return m
}
