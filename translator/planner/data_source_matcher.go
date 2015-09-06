package planner

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"gopkg.in/mgo.v2/bson"
)

func getFieldCaseInsensitive(doc *bson.D, field string) interface{} {
	field = strings.ToLower(field)
	for _, entry := range *doc {
		if strings.ToLower(entry.Name) == field {
			return entry.Value
		}
	}
	return nil
}

func valuesEqual(rawA interface{}, rawB interface{}) bool {
	return reflect.DeepEqual(rawA, rawB)
}

func fieldMatches_op(field_value interface{}, op_name string, val interface{}) (bool, error) {
	switch op_name {
	case "$eq":
		return valuesEqual(val, field_value), nil
	case "$regex":
		return regexp.Match(val.(string), []byte(field_value.(string)))
	default:
		return false, fmt.Errorf("unknown op name: %s\n", op_name)
	}
}


func fieldMatches(field_value interface{}, op interface{}) (bool, error) {
	switch k := op.(type) {
	case bson.M:
		for op_name, val := range k {
			single, err := fieldMatches_op(field_value, op_name, val)
			if err != nil || !single {
				return single, err
			}
		}
		return true, nil
	case bson.D:
		for _, entry := range k {
			single, err := fieldMatches_op(field_value, entry.Name, entry.Value)
			if err != nil || !single {
				return single, err
			}
		}
		return true, nil

	default:
		return false, fmt.Errorf("can't handle op type: %T %s", op, op)
	}

}

func Matches(query interface{}, doc *bson.D) (bool, error) {
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
	case *bson.D:
		for _, entry := range *q {
			res, err := fieldMatches(getFieldCaseInsensitive(doc, entry.Name), entry.Value)
			if err != nil {
				return false, err
			}
			if !res {
				return false, nil
			}
		}
	case bson.D:
		for _, entry := range q {
			res, err := fieldMatches(getFieldCaseInsensitive(doc, entry.Name), entry.Value)
			if err != nil {
				return false, err
			}
			if !res {
				return false, nil
			}
		}

	default:
		return false, fmt.Errorf("can't handle query type: %T %V", query, query)
	}
	return true, nil
}
