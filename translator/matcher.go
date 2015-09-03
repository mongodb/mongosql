package translator

import "fmt"
import "reflect"
import "strings"

import "gopkg.in/mgo.v2/bson"

func getFieldCaseInsensitive(doc *bson.M, field string) interface{} {
	field = strings.ToLower(field)
	for k, v := range(*doc) {
		if strings.ToLower(k) == field {
			return v
		}
	}
	return nil
}

func valuesEqual(rawA interface{}, rawB interface{}) bool {
	return reflect.DeepEqual(rawA, rawB)
}

func fieldMatches(field_value interface{}, op interface{}) (bool, error) {
	switch k := op.(type) {
	case bson.M:
		for op_name, val := range(k) {
			if op_name == "$eq" {
				return valuesEqual(val, field_value), nil
			} else {
				return false, fmt.Errorf("unknown op name: %s\n", op_name)
			}
		}
		return true, nil
	default:
		return false, fmt.Errorf("can't handle op  type: %T", op)
	}
	
}

func Matches(query interface{}, doc *bson.M) (bool, error) {
	if query == nil {
		return true, nil
	}

	switch q := query.(type) {
	case bson.M:
		for field, op := range(q) {
			res, err := fieldMatches(getFieldCaseInsensitive(doc, field), op)
			if err != nil {
				return false, err
			}
			if !res {
				return false, nil
			}
		}
	default:
		return false, fmt.Errorf("can't handle query type: %V", query)
	}
	return true, nil
}
