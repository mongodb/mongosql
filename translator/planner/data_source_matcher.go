package planner

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"gopkg.in/mgo.v2/bson"
)

func getFieldCaseInsensitive(doc *bson.M, field string) interface{} {
	field = strings.ToLower(field)
	for k, v := range *doc {
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
		for op_name, val := range k {
			switch op_name {
			case "$eq":
				return valuesEqual(val, field_value), nil
			case "$regex":
				return regexp.Match(val.(string), []byte(field_value.(string)))
			default:
				return false, fmt.Errorf("unknown op name: %s\n", op_name)
			}
		}
		return true, nil
	default:
		return false, fmt.Errorf("can't handle op type: %T %s", op, op)
	}

}

func Matches(query interface{}, doc *bson.M) (bool, error) {
	if query == nil {
		return true, nil
	}

	switch q := query.(type) {
	case bson.M:
		for field, op := range q {
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
