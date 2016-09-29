package evaluator

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"github.com/mongodb/mongo-tools/common/bsonutil"
	"gopkg.in/mgo.v2/bson"
)

const (
	Dot        = "_DOT_"
	sqlOpEQ    = "="
	sqlOpNEQ   = "!="
	sqlOpNSE   = "<=>"
	sqlOpGT    = ">"
	sqlOpGTE   = ">="
	sqlOpLT    = "<"
	sqlOpLTE   = "<="
	sqlOpLike  = "like"
	sqlOpIn    = "in"
	sqlOpNotIn = "not in"
)

// comparisonExpr returns a SQLExpr formed using op comparison operator.
func comparisonExpr(left, right SQLExpr, op string) (SQLExpr, error) {
	switch op {
	case sqlOpEQ:
		// BI-235: Allow users to pass custom $match stage
		if isMongoFilterExpr(left) {
			filter, err := parseMongoFilter(left, right)
			if err != nil {
				return nil, err
			}
			return &MongoFilterExpr{left.(SQLColumnExpr), right, filter}, nil
		}
		return &SQLEqualsExpr{left, right}, nil
	case sqlOpLT:
		return &SQLLessThanExpr{left, right}, nil
	case sqlOpGT:
		return &SQLGreaterThanExpr{left, right}, nil
	case sqlOpLTE:
		return &SQLLessThanOrEqualExpr{left, right}, nil
	case sqlOpGTE:
		return &SQLGreaterThanOrEqualExpr{left, right}, nil
	case sqlOpNEQ:
		return &SQLNotEqualsExpr{left, right}, nil
	case sqlOpLike:
		return &SQLLikeExpr{left, right}, nil
	case sqlOpNSE:
		return &SQLNullSafeEqualsExpr{left, right}, nil
	case sqlOpIn:
		if eval, ok := right.(*SQLSubqueryExpr); ok {
			return &SQLSubqueryCmpExpr{subqueryIn, left, eval, ""}, nil
		}
		return &SQLInExpr{left, right}, nil
	case sqlOpNotIn:
		if eval, ok := right.(*SQLSubqueryExpr); ok {
			return &SQLSubqueryCmpExpr{subqueryNotIn, left, eval, ""}, nil
		}
		return &SQLNotExpr{&SQLInExpr{left, right}}, nil
	default:
		return nil, mysqlerrors.Newf(mysqlerrors.ER_NOT_SUPPORTED_YET, "No support for comparison operator '%v'", op)
	}
}

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

func isMongoFilterExpr(expr SQLExpr) bool {
	if colExpr, ok := expr.(SQLColumnExpr); ok {
		if colExpr.columnType.MongoType == schema.MongoFilter {
			return true
		}
	}
	return false
}

func parseMongoFilter(left, right SQLExpr) (bson.M, error) {
	if _, ok := right.(SQLVarchar); !ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_CANT_USE_OPTION_HERE, left.String())
	}

	filter := bson.M{}
	err := json.Unmarshal([]byte(right.String()), &filter)
	if err != nil {
		return nil, mysqlerrors.Newf(mysqlerrors.ER_PARSE_ERROR, "parse mongo filter error: %s", err)
	}

	return filter, nil
}
