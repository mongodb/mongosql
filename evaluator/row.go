package evaluator

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/mongodb/mongo-tools/common/bsonutil"
	"gopkg.in/mgo.v2/bson"
)

const (
	Dot = "_DOT_"
)

// Row holds data from one or more tables.
type Row struct {
	Data Values
}

type Rows []Row

// Value holds row value for an SQL column.
type Value struct {
	Table string
	Name  string
	Data  interface{}
}

type Values []Value

var bsonDType = reflect.TypeOf(bson.D{})

// GetField takes a table returns the given value of the given key
// in the document, or nil if it does not exist.
// The second return value is a boolean indicating if the field was found or not, to allow
// the distinction betwen a null value stored in that field from a missing field.
func (row *Row) GetField(table, name string) (interface{}, bool) {
	for _, r := range row.Data {
		if strings.EqualFold(r.Table, table) && strings.EqualFold(r.Name, name) {
			return r.Data, true
		}
	}
	return nil, false
}

// GetValues gets the values of the columns.
func (row *Row) GetValues() []interface{} {
	values := make([]interface{}, 0)

	for _, v := range row.Data {
		values = append(values, v.Data)
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
func extractFieldByName(fieldName string, document interface{}) (interface{}, bool) {
	dotParts := strings.Split(fieldName, ".")
	var subdoc interface{} = document

	for _, path := range dotParts {
		docValue := reflect.ValueOf(subdoc)
		if !docValue.IsValid() {
			return nil, false
		}
		docType := docValue.Type()
		docKind := docType.Kind()
		if docKind == reflect.Map {
			subdocVal := docValue.MapIndex(reflect.ValueOf(path))
			if subdocVal.Kind() == reflect.Invalid {
				return nil, false
			}
			subdoc = subdocVal.Interface()
		} else if docKind == reflect.Slice {
			if docType == bsonDType {
				// dive into a D as a document
				asD := subdoc.(bson.D)
				var err error
				subdoc, err = bsonutil.FindValueByKey(path, &asD)
				if err != nil {
					return nil, false
				}
			} else {
				//  check that the path can be converted to int
				arrayIndex, err := strconv.Atoi(path)
				if err != nil {
					return nil, false
				}
				// bounds check for slice
				if arrayIndex < 0 || arrayIndex >= docValue.Len() {
					return nil, false
				}
				subdocVal := docValue.Index(arrayIndex)
				if subdocVal.Kind() == reflect.Invalid {
					return nil, false
				}
				subdoc = subdocVal.Interface()
			}
		} else {
			// trying to index into a non-compound type - just return blank.
			return nil, false
		}
	}
	return subdoc, true
}
