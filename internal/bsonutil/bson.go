package bsonutil

import (
	oldbson "github.com/10gen/mongo-go-driver/bson"
)

// NewDocElem returns a bson.DocElem with Key key and Value value.
func NewDocElem(key string, value interface{}) oldbson.DocElem {
	return oldbson.DocElem{Name: key, Value: value}
}

// NewM returns a new bson.M made from key value pairs kvs.
func NewM(kvs ...oldbson.DocElem) oldbson.M {
	m := make(map[string]interface{}, len(kvs))

	for _, kv := range kvs {
		m[kv.Name] = kv.Value
	}

	return m
}

// NewMArray returns an array of bson.M.
func NewMArray(bsonMs ...oldbson.M) []oldbson.M {
	//return empty slice instead of nil slice
	if len(bsonMs) == 0 {
		return []oldbson.M{}
	}

	return bsonMs
}

// NewD returns a new bson.D made from key value pairs kvs.
func NewD(kvs ...oldbson.DocElem) oldbson.D {
	//return empty slice instead of nil slice
	if len(kvs) == 0 {
		return oldbson.D{}
	}

	return kvs
}

// NewDArray returns an array of bson.D.
func NewDArray(bsonDs ...oldbson.D) []oldbson.D {
	//return empty slice instead of nil slice
	if len(bsonDs) == 0 {
		return []oldbson.D{}
	}

	return bsonDs
}

// NewArray returns an array of interface{}.
func NewArray(vals ...interface{}) []interface{} {
	//return empty slice instead of nil slice
	if len(vals) == 0 {
		return []interface{}{}
	}
	// NewArray is just []interface{}
	return vals
}
