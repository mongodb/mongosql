package bsonutil

import (
	"go.mongodb.org/mongo-driver/bson"
)

// D is the type of bson documents.
type D = bson.D

// NewDocElem returns a bson.E with Key key and Value value.
func NewDocElem(key string, value interface{}) bson.E {
	return bson.E{Key: key, Value: value}
}

// NewM returns a new bson.M made from key value pairs kvs.
func NewM(kvs ...bson.E) bson.M {
	m := make(bson.M, len(kvs))

	for _, kv := range kvs {
		m[kv.Key] = kv.Value
	}

	return m
}

// NewMArray returns an array of bson.M.
func NewMArray(bsonMs ...bson.M) []bson.M {
	//return empty slice instead of nil slice
	if len(bsonMs) == 0 {
		return []bson.M{}
	}

	return bsonMs
}

// NewD returns a new bson.D made from key value pairs kvs.
func NewD(kvs ...bson.E) bson.D {
	//return empty bson.D instead of nil bson.D
	if len(kvs) == 0 {
		return bson.D{}
	}

	return kvs
}

// NewDArray returns an array of bson.D.
func NewDArray(bsonDs ...bson.D) []bson.D {
	//return empty slice instead of nil slice
	if len(bsonDs) == 0 {
		return []bson.D{}
	}

	return bsonDs
}

// NewArray returns an array of interface{}.
func NewArray(vals ...interface{}) bson.A {
	//return empty slice instead of nil slice
	if len(vals) == 0 {
		return bson.A{}
	}

	return vals
}
