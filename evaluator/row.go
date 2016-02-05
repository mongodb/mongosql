package evaluator

import (
	"github.com/mongodb/mongo-tools/common/bsonutil"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"strconv"
	"strings"
)

const (
	Dot = "_DOT_"
)

// Row holds data from one or more tables.
type Row struct {
	Data TableRows
}

type Rows []Row

// TableRow holds column data from a given table.
type TableRow struct {
	Table  string
	Values Values
}

type TableRows []TableRow

// Value holds row value for an SQL column.
type Value struct {
	Name string
	View string
	Data interface{}
}

type Values []Value

var bsonDType = reflect.TypeOf(bson.D{})

// GetField takes a table returns the given value of the given key
// in the document, or nil if it does not exist.
// The second return value is a boolean indicating if the field was found or not, to allow
// the distinction betwen a null value stored in that field from a missing field.
// The key parameter may be a dot-delimited string to reference a field that is nested
// within a subdocument.
func (row *Row) GetField(table, name string) (interface{}, bool) {
	for _, r := range row.Data {
		// TODO: need to remove the need for tables
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

// dottifyFieldName translates any dots in a field name into the Dot constant
func dottifyFieldName(fieldName string) string {
	return strings.Replace(fieldName, ".", Dot, -1)
}

// deDottifyFieldName translates any Dot constant in a field name into a '.'
func deDottifyFieldName(fieldName string) string {
	return strings.Replace(fieldName, Dot, ".", -1)
}

// extractFieldByName takes a field name and document, and returns a value representing
// the value of that field in the document in a format that can be printed as a string.
// It will also handle dot-delimited field names for nested arrays or documents.
func extractFieldByName(fieldName string, document interface{}) interface{} {
	dotParts := strings.Split(fieldName, ".")
	var subdoc interface{} = document

	for _, path := range dotParts {
		docValue := reflect.ValueOf(subdoc)
		if !docValue.IsValid() {
			return ""
		}
		docType := docValue.Type()
		docKind := docType.Kind()
		if docKind == reflect.Map {
			subdocVal := docValue.MapIndex(reflect.ValueOf(path))
			if subdocVal.Kind() == reflect.Invalid {
				return ""
			}
			subdoc = subdocVal.Interface()
		} else if docKind == reflect.Slice {
			if docType == bsonDType {
				// dive into a D as a document
				asD := subdoc.(bson.D)
				var err error
				subdoc, err = bsonutil.FindValueByKey(path, &asD)
				if err != nil {
					return ""
				}
			} else {
				//  check that the path can be converted to int
				arrayIndex, err := strconv.Atoi(path)
				if err != nil {
					return ""
				}
				// bounds check for slice
				if arrayIndex < 0 || arrayIndex >= docValue.Len() {
					return ""
				}
				subdocVal := docValue.Index(arrayIndex)
				if subdocVal.Kind() == reflect.Invalid {
					return ""
				}
				subdoc = subdocVal.Interface()
			}
		} else {
			// trying to index into a non-compound type - just return blank.
			return ""
		}
	}
	return subdoc
}
