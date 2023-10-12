package mongosql

import (
	"reflect"
	"testing"

	"github.com/10gen/mongosql-rs/go/mongosql/internal/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestExcludeNamespaces(t *testing.T) {
	schema := bson.M{
		"bsonType": "object",
		"properties": bson.M{
			"a": bson.M{"bsonType": "double"},
		},
		"additionalProperties": false,
	}

	bytes, err := bson.Marshal(&schema)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	catalogSchema := map[string]map[string]bsoncore.Document{
		"bar": {"foo": bytes},
	}

	translation, err := Translate(TranslationArgs{
		DB:                "bar",
		SQL:               "select * from foo",
		CatalogSchema:     catalogSchema,
		ExcludeNamespaces: true,
	})
	if err != nil {
		t.Fatalf("expected err to be nil, got '%s'", err)
	}

	if translation.TargetDB != "bar" {
		t.Fatalf("expected targetDB to be 'bar', got '%s'", translation.TargetDB)
	}

	if translation.TargetCollection != "foo" {
		t.Fatalf("expected targetCollection to be 'foo', got '%s'", translation.TargetCollection)
	}

	var pipeline []bson.D
	val := bson.RawValue{
		Type:  bsontype.Array,
		Value: translation.Pipeline,
	}
	err = val.Unmarshal(&pipeline)
	if err != nil {
		t.Fatalf("expected pipeline to unmarshal into []bson.D, but failed: %s", err)
	}

	if len(pipeline) != 2 {
		t.Fatalf("expected pipeline to have two stages, but found %d", len(pipeline))
	}

	expectedPipeline := bson.A{
		bson.D{
			{"$project", bson.D{
				{"foo", "$$ROOT"},
				{"_id", int32(0)},
			}}},
		bson.D{{"$replaceWith", "$foo"}},
	}

	for i := range pipeline {
		if !reflect.DeepEqual(expectedPipeline[i], pipeline[i]) {
			t.Fatalf("expected stages to be equal, but they weren't:\n%s\nand\n%s", expectedPipeline[i], pipeline[i])
		}
	}

	expectedResultSetSchema := bson.D{
		{"bsonType", "object"},
		{"properties", bson.D{
			{"a", bson.D{
				{"bsonType", "double"},
			}},
		}},
		{"required", bson.A{"a"}},
		{"additionalProperties", false},
	}

	util.CheckResultSetSchema(t, expectedResultSetSchema, translation.ResultSetSchema)
}
