package util

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func CheckResultSetSchema(t *testing.T, expected bson.D, found bsoncore.Document) {
	var foundResultSetSchema bson.D
	val := bson.RawValue{
		Type:  bsontype.EmbeddedDocument,
		Value: found,
	}
	err := val.Unmarshal(&foundResultSetSchema)
	if err != nil {
		t.Fatalf("failed to unmarshal bson '%s'", err)
	}

	if !reflect.DeepEqual(expected, foundResultSetSchema) {
		t.Fatalf("expected resultset schema to be equal, but they weren't:\n%s\nand\n%s", expected, foundResultSetSchema)
	}
}

func CheckSelectListOrder(t *testing.T, expected bson.A, found bsoncore.Array) {
	var foundSelectListOrder bson.A
	val := bson.RawValue{
		Type:  bsontype.Array,
		Value: found,
	}
	err := val.Unmarshal(&foundSelectListOrder)
	if err != nil {
		t.Fatalf("failed to unmarshal bson '%s'", err)
	}

	if !reflect.DeepEqual(expected, foundSelectListOrder) {
		t.Fatalf("expected select list order to be equal, but they weren't:\n%s\nand\n%s", expected, foundSelectListOrder)
	}
}

func GenerateTestSchema() (bsoncore.Document, error) {
	schema := bson.M{
		"bsonType": "object",
		"properties": bson.M{
			"a": bson.M{"bsonType": "double"},
		},
	}
	bytes, err := bson.Marshal(&schema)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func GenerateDefaultCollectionSchema() (bsoncore.Document, error) {
	schema := bson.D{
		{"bsonType", "object"},
		{"additionalProperties", true},
	}

	return bson.Marshal(&schema)
}
