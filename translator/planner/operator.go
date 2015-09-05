package planner

import (
	"github.com/erh/mongo-sql-temp/config"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// ExecutionCtx holds data that is used by each operator.
type ExecutionCtx struct {
	Config *config.Config
	Db     string
}

// Operator defines a set of functions that are implemented by each
// node in the query tree.
type Operator interface {
	Open(*ExecutionCtx) error
	Next(*Row) bool
	Close() error
	Err() error
}

// Row holds data from one or more tables.
type Row struct {
	Data []TableRow
}

// TableRow holds column data from a given table.
type TableRow struct {
	Table  string
	Values bson.D
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
			return getKey(key, r.Values)
		}
	}
	return nil, false
}

func getKey(key string, doc bson.D) (interface{}, bool) {
	index := strings.Index(key, ".")
	if index == -1 {
		value, hasValue := doc.Map()[key]
		return value, hasValue
	}
	left := key[0:index]
	docMap := doc.Map()
	value, hasValue := docMap[left]
	if value == nil {
		return value, hasValue
	}
	subDoc, ok := docMap[left].(bson.D)
	if !ok {
		return nil, false
	}
	return getKey(key[index+1:], subDoc)
}
