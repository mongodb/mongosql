package bsonutil

import (
	"go.mongodb.org/mongo-driver/bson"
)

// URI literals
const (
	MongoDBScheme     = "mongodb://"
	DefaultMongoDBURI = "mongodb://localhost:27017"
)

func bsonDToMap(doc bson.D) map[string]interface{} {
	m := map[string]interface{}{}
	for _, e := range doc {
		switch typedV := e.Value.(type) {
		case bson.D:
			m[e.Key] = bsonDToMap(typedV)
		case bson.M:
			m[e.Key] = bsonMToMap(typedV)
		case bson.A:
			m[e.Key] = bsonAToMaps(typedV)
		default:
			m[e.Key] = typedV
		}
	}
	return m
}

func bsonMToMap(doc bson.M) map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range doc {
		switch typedV := v.(type) {
		case bson.D:
			m[k] = bsonDToMap(typedV)
		case bson.M:
			m[k] = bsonMToMap(typedV)
		case bson.A:
			m[k] = bsonAToMaps(typedV)
		default:
			m[k] = typedV
		}
	}
	return m
}

func bsonAToMaps(a bson.A) []interface{} {
	s := make([]interface{}, len(a))
	for i, e := range a {
		switch typedE := e.(type) {
		case bson.D:
			s[i] = bsonDToMap(typedE)
		case bson.M:
			s[i] = bsonMToMap(typedE)
		case bson.A:
			s[i] = bsonAToMaps(typedE)
		default:
			s[i] = typedE
		}
	}

	return s
}

// PipelineToMapSlice converts a slice of bson.D
// to a slice of map[string]interface - with
// each element recursively converted.
func PipelineToMapSlice(pipeline []bson.D) []map[string]interface{} {
	m := make([]map[string]interface{}, 0)
	for _, stage := range pipeline {
		m = append(m, bsonDToMap(stage))
	}
	return m
}
