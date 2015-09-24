package types

import (
	"github.com/erh/mongo-sql-temp/config"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// Row holds data from one or more tables.
type Row struct {
	Data []TableRow
}

// TableRow holds column data from a given table.
type TableRow struct {
	Table       string
	Values      bson.D
	TableConfig *config.TableConfig
}

// GetField takes a table returns the given value of the given key
// in the document, or nil if it does not exist.
// The second return value is a boolean indicating if the field was found or not, to allow
// the distinction betwen a null value stored in that field from a missing field.
// The key parameter may be a dot-delimited string to reference a field that is nested
// within a subdocument.
func (row *Row) GetField(table, key string) (interface{}, bool) {
	for _, r := range row.Data {
		if r.Table == table {
			for _, entry := range r.Values {
				// TODO optimize
				if strings.ToLower(key) == strings.ToLower(entry.Name) {
					return entry.Value, true
				}
			}
		}
	}
	return nil, false
}
