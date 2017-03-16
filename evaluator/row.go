package evaluator

import (
	"reflect"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"
)

// Row holds data from one or more tables.
type Row struct {
	Data Values
}

type Rows []Row

// Value holds row values for a SQL column.
type Value struct {
	SelectID int
	Table    string
	Name     string
	Data     SQLValue
}

type Values []Value

var bsonDType = reflect.TypeOf(bson.D{})

// GetField takes a selectID, tableName, and columnName and returns the given value of the given key
// in the row, or nil if it does not exist.
// The second return value is a boolean indicating if the field was found or not, to allow
// the distinction betwen a null value stored in that field from a missing field.
func (row *Row) GetField(selectID int, tableName, columnName string) (SQLValue, bool) {
	for _, r := range row.Data {
		if r.SelectID == selectID && strings.EqualFold(r.Table, tableName) && strings.EqualFold(r.Name, columnName) {
			return r.Data, true
		}
	}
	return nil, false
}

// GetValues gets the values of the columns.
func (row *Row) GetValues() []SQLValue {
	values := make([]SQLValue, 0)

	for _, v := range row.Data {
		values = append(values, v.Data)
	}

	return values
}

func (values Values) Map() map[string]SQLValue {
	m := make(map[string]SQLValue, 0)
	for _, value := range values {
		m[value.Name] = value.Data
	}
	return m
}
