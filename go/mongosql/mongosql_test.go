package mongosql_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/10gen/mongosql-rs/go/mongosql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

func TestVersion(t *testing.T) {
	v := mongosql.Version()

	parts := strings.SplitN(v, "-", 2)

	release := parts[0]
	releaseParts := strings.Split(release, ".")
	if len(releaseParts) != 3 {
		t.Fatalf("expected version %q to have three release parts", v)
	}

	if len(parts) == 2 {
		preRelease := parts[1]
		if len(preRelease) < 1 {
			t.Fatalf("expected version %q to have non-empty pre-release", v)
		}
	}
}

func TestTranslate(t *testing.T) {
	translation, err := mongosql.Translate(mongosql.TranslationArgs{
		DB:  "bar",
		SQL: "select * from foo",
	})
	if err != nil {
		t.Fatalf("expected err to be nil, got '%s'", err)
	}

	if translation.TargetDB != "bar" {
		t.Fatalf("expected targetDB to be 'bar', got '%s'", translation.TargetDB)
	}

	if *translation.TargetCollection != "foo" {
		t.Fatalf("expected targetCollection to be 'foo', got '%s'", *translation.TargetCollection)
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

	expectedStage0 := bson.D{
		{"$project", bson.D{
			{"_id", int32(0)},
			{"foo", "$$ROOT"},
		}},
	}
	if !reflect.DeepEqual(expectedStage0, pipeline[0]) {
		t.Fatalf("expected stages to be equal, but they weren't:\n%s\nand\n%s", expectedStage0, pipeline[0])
	}
	expectedStage1 := bson.D{
		{"$project", bson.D{
			{"_id", int32(0)},
			{"foo", "$foo"},
		}},
	}
	if !reflect.DeepEqual(expectedStage1, pipeline[1]) {
		t.Fatalf("expected stages to be equal, but they weren't:\n%s\nand\n%s", expectedStage1, pipeline[1])
	}
}

func TestTranslateError(t *testing.T) {
	_, err := mongosql.Translate(mongosql.TranslationArgs{
		DB:  "bar",
		SQL: "notavalidquery",
	})

	if err == nil {
		t.Fatalf("expected error to be non-nil, but it was nil")
	}

	if !strings.Contains(err.Error(), "parse error: Unrecognized token `notavalidquery`") {
		t.Fatalf("error message did not contain expected text: %q", err.Error())
	}
}
