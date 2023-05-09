package mongosql

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestRelaxedSchemaChecking(t *testing.T) {
	schema := bson.D{
		{"bsonType", "object"},
		{"additionalProperties", true},
	}

	bytes, err := bson.Marshal(&schema)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	catalogSchema := map[string]map[string]bsoncore.Document{
		"test": {"grades": bsoncore.Document(bytes)},
	}

	translation, err := Translate(TranslationArgs{
		DB:                  "test",
		SQL:                 "select studentid from grades where score > 80",
		CatalogSchema:       catalogSchema,
		relaxSchemaChecking: true,
	})
	if err != nil {
		t.Fatalf("expected err to be nil, got '%s'", err)
	}

	if translation.TargetDB != "test" {
		t.Fatalf("expected targetDB to be 'test', got '%s'", translation.TargetDB)
	}

	if translation.TargetCollection != "grades" {
		t.Fatalf("expected targetCollection to be 'grades', got '%s'", translation.TargetCollection)
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

	if len(pipeline) != 5 {
		t.Fatalf("expected pipeline to have five stages, but found %d", len(pipeline))
	}
}
