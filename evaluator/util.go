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

func containsAnyInt(ints []int, test []int) bool {
	for _, value := range test {
		if containsInt(ints, value) {
			return true
		}
	}

	return false
}

func containsInt(ints []int, i int) bool {
	for _, value := range ints {
		if value == i {
			return true
		}
	}
	return false
}

func containsStringFunc(strs []string, str string, f func(string, string) bool) bool {
	for _, n := range strs {
		if f(n, str) {
			return true
		}
	}

	return false
}

func containsString(strs []string, str string) bool {
	return containsStringFunc(strs, str, func(s1, s2 string) bool {
		return s1 == s2
	})
}

func containsStringInsensitive(strs []string, str string) bool {
	return containsStringFunc(strs, str, func(s1, s2 string) bool {
		return strings.ToLower(s1) == strings.ToLower(s2)
	})
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

func findKeyInDoc(key string, d interface{}) (interface{}, bool) {

	var doc bson.M
	switch typedD := d.(type) {
	case bson.M:
		doc = typedD
	case *bson.M:
		doc = *typedD
	default:
		return nil, false
	}

	i := strings.Index(key, ".")
	if i > 0 {
		ckey := key[0:i]
		v, ok := doc[ckey]
		if !ok {
			return nil, false
		}

		return findKeyInDoc(key[i+1:], v)
	}

	v, ok := doc[key]
	return v, ok
}

func findArrayInDoc(key string, doc interface{}) ([]interface{}, bool) {
	v, ok := findKeyInDoc(key, doc)
	if !ok {
		return nil, ok
	}

	result, ok := v.([]interface{})
	return result, ok
}

func findDocInDoc(key string, doc interface{}) (bson.M, bool) {
	v, ok := findKeyInDoc(key, doc)
	if !ok {
		return nil, ok
	}

	result, ok := v.(bson.M)
	return result, ok
}

func findStringInDoc(key string, doc interface{}) (string, bool) {
	v, ok := findKeyInDoc(key, doc)
	if !ok {
		return "", ok
	}

	result, ok := v.(string)
	return result, ok
}

func getKey(key string, doc bson.D) (interface{}, bool) {
	index := strings.Index(key, ".")
	if index == -1 {
		for _, entry := range doc {
			if strings.ToLower(key) == strings.ToLower(entry.Name) { // TODO optimize
				return entry.Value, true
			}
		}
		return nil, false
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
