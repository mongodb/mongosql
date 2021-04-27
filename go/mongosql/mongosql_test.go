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
	db, collection, pipelineBytes := mongosql.Translate("bar", "select *")

	if db != "bar" {
		t.Fatalf("expected db to be 'bar', got '%s'", db)
	}

	if collection != "foo" {
		t.Fatalf("expected collection to be 'foo', got '%s'", collection)
	}

	var pipeline []bson.D
	val := bson.RawValue{
		Type:  bsontype.Array,
		Value: pipelineBytes,
	}
	err := val.Unmarshal(&pipeline)
	if err != nil {
		t.Fatalf("expected pipeline to unmarshal into []bson.D, but failed: %s", err)
	}

	if len(pipeline) != 1 {
		t.Fatalf("expected pipeline to have one stage, but found %d", len(pipeline))
	}

	expectedStage := bson.D{
		{"$project", bson.D{
			{"_id", int32(0)},
			{"a", "$a"},
		}},
	}
	if !reflect.DeepEqual(expectedStage, pipeline[0]) {
		t.Fatalf("expected stages to be equal, but they weren't:\n%s\nand\n%s", expectedStage, pipeline[0])
	}
}
