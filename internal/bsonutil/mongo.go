package bsonutil

import (
	"fmt"

	"github.com/10gen/mongo-go-driver/bson"
)

// URI literals
const (
	MongoDBScheme     = "mongodb://"
	DefaultMongoDBURI = "mongodb://localhost:27017"
)

func bsonDToMap(doc bson.D) map[string]interface{} {
	m := map[string]interface{}{}
	for _, l := range doc {
		switch typedV := l.Value.(type) {
		case bson.D:
			m[l.Name] = bsonDToMap(typedV)
		case bson.M:
			m[l.Name] = bsonMToMap(typedV)
		default:
			m[l.Name] = typedV
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
		default:
			m[k] = typedV
		}
	}
	return m
}

// ConvertBSONToMap recursively converts a bson.D/bson.M to a map[string]interface{}.
func ConvertBSONToMap(doc interface{}) map[string]interface{} {
	switch typedD := doc.(type) {
	case bson.D:
		return bsonDToMap(typedD)
	case bson.M:
		return bsonMToMap(typedD)
	}
	panic(fmt.Sprintf("Unrecognized bson type: %T", doc))
}

// PipelineToMapSlice converts a slice of bson.D
// to a slice of map[string]interface - with
// each element recursively converted.
func PipelineToMapSlice(pipeline []bson.D) []map[string]interface{} {
	m := make([]map[string]interface{}, 0)
	for _, stage := range pipeline {
		m = append(m, ConvertBSONToMap(stage))
	}
	return m
}
