package evaluator

import (
	"strings"

	"gopkg.in/mgo.v2/bson"
)

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
