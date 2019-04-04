package parser_test

import (
	"testing"

	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/10gen/mongoast/internal/parsertest"
	"github.com/10gen/mongoast/parser"

	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestDeparseStage(t *testing.T) {
	testCases := []struct {
		input    string
		expected bsoncore.Value
	}{
		{
			`{"$addFields": {"a": "$foo", "b": 1}}`,
			bsonutil.DocumentFromElements(
				"$addFields", bsonutil.DocumentFromElements(
					"a", bsonutil.String("$foo"),
					"b", bsonutil.DocumentFromElements(
						"$literal", bsonutil.Int32(1),
					),
				),
			),
		},
		{
			`{"$bucket": {"groupBy": "$a", "boundaries": [0, 10, 20]}}`,
			bsonutil.DocumentFromElements(
				"$bucket", bsonutil.DocumentFromElements(
					"groupBy", bsonutil.String("$a"),
					"boundaries", bsonutil.ArrayFromValues(
						bsonutil.Int32(0),
						bsonutil.Int32(10),
						bsonutil.Int32(20),
					),
				),
			),
		},
		{
			`{"$bucket": {"groupBy": "$a", "boundaries": [0, 10, 20], "default": -1, "output": {"count": {"$sum": 1}, "total": {"$sum": "$a"}}}}`,
			bsonutil.DocumentFromElements(
				"$bucket", bsonutil.DocumentFromElements(
					"groupBy", bsonutil.String("$a"),
					"boundaries", bsonutil.ArrayFromValues(
						bsonutil.Int32(0),
						bsonutil.Int32(10),
						bsonutil.Int32(20),
					),
					"default", bsonutil.Int32(-1),
					"output", bsonutil.DocumentFromElements(
						"count", bsonutil.DocumentFromElements(
							"$sum", bsonutil.DocumentFromElements(
								"$literal", bsonutil.Int32(1),
							),
						),
						"total", bsonutil.DocumentFromElements(
							"$sum", bsonutil.String("$a"),
						),
					),
				),
			),
		},
		{
			`{"$bucketAuto": {"groupBy": "$a", "buckets": 2}}`,
			bsonutil.DocumentFromElements(
				"$bucketAuto", bsonutil.DocumentFromElements(
					"groupBy", bsonutil.String("$a"),
					"buckets", bsonutil.Int64(2),
				),
			),
		},
		{
			`{"$bucketAuto": {"groupBy": "$a", "buckets": 2, "output": {"count": {"$sum": 1}, "total": {"$sum": "$a"}}, "granularity": "R5"}}`,
			bsonutil.DocumentFromElements(
				"$bucketAuto", bsonutil.DocumentFromElements(
					"groupBy", bsonutil.String("$a"),
					"buckets", bsonutil.Int64(2),
					"output", bsonutil.DocumentFromElements(
						"count", bsonutil.DocumentFromElements(
							"$sum", bsonutil.DocumentFromElements(
								"$literal", bsonutil.Int32(1),
							),
						),
						"total", bsonutil.DocumentFromElements(
							"$sum", bsonutil.String("$a"),
						),
					),
					"granularity", bsonutil.String("R5"),
				),
			),
		},
		{
			`{"$collStats": {}}`,
			bsonutil.DocumentFromElements(
				"$collStats", bsonutil.EmptyDocument(),
			),
		},
		{
			`{"$collStats": {"latencyStats": {}}}`,
			bsonutil.DocumentFromElements(
				"$collStats", bsonutil.DocumentFromElements(
					"latencyStats", bsonutil.DocumentFromElements(
						"histograms", bsonutil.False,
					),
				),
			),
		},
		{
			`{"$collStats": {"latencyStats": {"histograms": true}}}`,
			bsonutil.DocumentFromElements(
				"$collStats", bsonutil.DocumentFromElements(
					"latencyStats", bsonutil.DocumentFromElements(
						"histograms", bsonutil.True,
					),
				),
			),
		},
		{
			`{"$collStats": {"storageStats": {}}}`,
			bsonutil.DocumentFromElements(
				"$collStats", bsonutil.DocumentFromElements(
					"storageStats", bsonutil.EmptyDocument(),
				),
			),
		},
		{
			`{"$collStats": {"count": {}}}`,
			bsonutil.DocumentFromElements(
				"$collStats", bsonutil.DocumentFromElements(
					"count", bsonutil.EmptyDocument(),
				),
			),
		},
		{
			`{"$collStats": {"latencyStats": {}, "storageStats": {}, "count": {}}}`,
			bsonutil.DocumentFromElements(
				"$collStats", bsonutil.DocumentFromElements(
					"latencyStats", bsonutil.DocumentFromElements(
						"histograms", bsonutil.False,
					),
					"storageStats", bsonutil.EmptyDocument(),
					"count", bsonutil.EmptyDocument(),
				),
			),
		},
		{
			`{"$count": "a"}`,
			bsonutil.DocumentFromElements(
				"$count", bsonutil.String("a"),
			),
		},
		{
			`{"$facet": {"a1": [{"$match": {"a": 1}}], "a2": [{"$match": {"a": 2}}]}}`,
			bsonutil.DocumentFromElements(
				"$facet", bsonutil.DocumentFromElements(
					"a1", bsonutil.ArrayFromValues(
						bsonutil.DocumentFromElements(
							"$match", bsonutil.DocumentFromElements(
								"a", bsonutil.DocumentFromElements(
									"$eq", bsonutil.Int32(1),
								),
							),
						),
					),
					"a2", bsonutil.ArrayFromValues(
						bsonutil.DocumentFromElements(
							"$match", bsonutil.DocumentFromElements(
								"a", bsonutil.DocumentFromElements(
									"$eq", bsonutil.Int32(2),
								),
							),
						),
					),
				),
			),
		},
		{
			`{"$group": {"_id": 1, "a": {"$sum": "$a"}}}`,
			bsonutil.DocumentFromElements(
				"$group", bsonutil.DocumentFromElements(
					"_id", bsonutil.DocumentFromElements(
						"$literal", bsonutil.Int32(1),
					),
					"a", bsonutil.DocumentFromElements(
						"$sum", bsonutil.String("$a"),
					),
				),
			),
		},
		{
			`{"$limit": 10}`,
			bsonutil.DocumentFromElements(
				"$limit", bsonutil.Int64(10),
			),
		},
		{
			`{"$lookup": { "from": "foo", "localField": "a", "foreignField": "b", "as": "bar" } }`,
			bsonutil.DocumentFromElements(
				"$lookup", bsonutil.DocumentFromElements(
					"from", bsonutil.String("foo"),
					"localField", bsonutil.String("a"),
					"foreignField", bsonutil.String("b"),
					"as", bsonutil.String("bar"),
				),
			),
		},
		{
			`{"$lookup": { "from": "foo", "pipeline": [{ "$match": { "a": 1 } }], "as": "bar" } }`,
			bsonutil.DocumentFromElements(
				"$lookup", bsonutil.DocumentFromElements(
					"from", bsonutil.String("foo"),
					"pipeline", bsonutil.ArrayFromValues(
						bsonutil.DocumentFromElements(
							"$match", bsonutil.DocumentFromElements(
								"a", bsonutil.DocumentFromElements(
									"$eq", bsonutil.Int32(1),
								),
							),
						),
					),
					"as", bsonutil.String("bar"),
				),
			),
		},
		{
			`{"$lookup": { "from": "foo", "let": {"x": "$a"}, "pipeline": [{"$match": {"x": 1 }}], "as": "bar"}}`,
			bsonutil.DocumentFromElements(
				"$lookup", bsonutil.DocumentFromElements(
					"from", bsonutil.String("foo"),
					"let", bsonutil.DocumentFromElements(
						"x", bsonutil.String("$a"),
					),
					"pipeline", bsonutil.ArrayFromValues(
						bsonutil.DocumentFromElements(
							"$match", bsonutil.DocumentFromElements(
								"x", bsonutil.DocumentFromElements(
									"$eq", bsonutil.Int32(1),
								),
							),
						),
					),
					"as", bsonutil.String("bar"),
				),
			),
		},
		{
			`{"$match": { "a": { "$gt" : 3 } } }`,
			bsonutil.DocumentFromElements(
				"$match", bsonutil.DocumentFromElements(
					"a", bsonutil.DocumentFromElements(
						"$gt", bsonutil.Int32(3),
					),
				),
			),
		},
		{
			`{"$match": { "a": { "$eee" : 3 } } }`,
			bsonutil.DocumentFromElements(
				"$match", bsonutil.DocumentFromElements(
					"a", bsonutil.DocumentFromElements(
						"$eee", bsonutil.Int32(3),
					),
				),
			),
		},
		{
			`{"$match": { "a": { "$eee" : [1,2] } } }`,
			bsonutil.DocumentFromElements(
				"$match", bsonutil.DocumentFromElements(
					"a", bsonutil.DocumentFromElements(
						"$eee", bsonutil.ArrayFromValues(
							bsonutil.Int32(1),
							bsonutil.Int32(2),
						),
					),
				),
			),
		},
		{
			`{"$project": { "a": 1, "b": 1 } }`,
			bsonutil.DocumentFromElements(
				"$project", bsonutil.DocumentFromElements(
					"a", bsonutil.Int32(1),
					"b", bsonutil.Int32(1),
				),
			),
		},
		{
			`{"$project": { "a": "$a" } }`,
			bsonutil.DocumentFromElements(
				"$project", bsonutil.DocumentFromElements(
					"a", bsonutil.String("$a"),
				),
			),
		},
		{
			`{"$project": { "a": "$a.b.c" } }`,
			bsonutil.DocumentFromElements(
				"$project", bsonutil.DocumentFromElements(
					"a", bsonutil.String("$a.b.c"),
				),
			),
		},
		{
			`{"$project": { "a": { "$eee": [1, 2] } } }`,
			bsonutil.DocumentFromElements(
				"$project", bsonutil.DocumentFromElements(
					"a", bsonutil.DocumentFromElements(
						"$eee", bsonutil.ArrayFromValues(
							bsonutil.DocumentFromElements(
								"$literal", bsonutil.Int32(1),
							),
							bsonutil.DocumentFromElements(
								"$literal", bsonutil.Int32(2),
							),
						),
					),
				),
			),
		},
		{
			`{"$project": { "a.b": 0 } }`,
			bsonutil.DocumentFromElements(
				"$project", bsonutil.DocumentFromElements(
					"a.b", bsonutil.Int32(0),
				),
			),
		},
		{
			`{"$project": { "a.b": 1 } }`,
			bsonutil.DocumentFromElements(
				"$project", bsonutil.DocumentFromElements(
					"a.b", bsonutil.Int32(1),
				),
			),
		},
		{
			`{"$project": { "a": { "b": 0 } } }`,
			bsonutil.DocumentFromElements(
				"$project", bsonutil.DocumentFromElements(
					"a.b", bsonutil.Int32(0),
				),
			),
		},
		{
			`{"$project": { "a": { "b": 1 } } }`,
			bsonutil.DocumentFromElements(
				"$project", bsonutil.DocumentFromElements(
					"a.b", bsonutil.Int32(1),
				),
			),
		},
		{
			`{"$redact": { "$cond": { "if": { "$eq": ["$a", 5] }, "then": 1, "else": 0 } } }`,
			bsonutil.DocumentFromElements(
				"$redact", bsonutil.DocumentFromElements(
					"$cond", bsonutil.DocumentFromElements(
						"if", bsonutil.DocumentFromElements(
							"$eq", bsonutil.ArrayFromValues(
								bsonutil.String("$a"),
								bsonutil.DocumentFromElements(
									"$literal", bsonutil.Int32(5),
								),
							),
						),
						"then", bsonutil.DocumentFromElements(
							"$literal", bsonutil.Int32(1),
						),
						"else", bsonutil.DocumentFromElements(
							"$literal", bsonutil.Int32(0),
						),
					),
				),
			),
		},
		{
			`{"$replaceRoot": { "newRoot": { "a": "$foo" } } }`,
			bsonutil.DocumentFromElements(
				"$replaceRoot", bsonutil.DocumentFromElements(
					"newRoot", bsonutil.DocumentFromElements(
						"a", bsonutil.String("$foo"),
					),
				),
			),
		},
		{
			`{"$skip": 10}`,
			bsonutil.DocumentFromElements(
				"$skip", bsonutil.Int64(10),
			),
		},
		{
			`{"$sort": { "a": 1, "b": -1 } }`,
			bsonutil.DocumentFromElements(
				"$sort", bsonutil.DocumentFromElements(
					"a", bsonutil.Int32(1),
					"b", bsonutil.Int32(-1),
				),
			),
		},
		{
			`{"$sort": { "a.2": 1, "b": -1 } }`,
			bsonutil.DocumentFromElements(
				"$sort", bsonutil.DocumentFromElements(
					"a.2", bsonutil.Int32(1),
					"b", bsonutil.Int32(-1),
				),
			),
		},
		{
			`{"$sortByCount": "$a"}`,
			bsonutil.DocumentFromElements(
				"$sortByCount", bsonutil.String("$a"),
			),
		},
		{
			`{"$sortedMerge": { "a": 1, "b": -1 } }`,
			bsonutil.DocumentFromElements(
				"$sortedMerge", bsonutil.DocumentFromElements(
					"a", bsonutil.Int32(1),
					"b", bsonutil.Int32(-1),
				),
			),
		},
		{
			`{"$eee": { "a": 1, "b": -1 } }`,
			bsonutil.DocumentFromElements(
				"$eee", bsonutil.DocumentFromElements(
					"a", bsonutil.Int32(1),
					"b", bsonutil.Int32(-1),
				),
			),
		},
		{
			`{"$unwind": "$a.b"}`,
			bsonutil.DocumentFromElements(
				"$unwind", bsonutil.String("$a.b"),
			),
		},
		{
			`{"$unwind": { "path": "$a.b"} }`,
			bsonutil.DocumentFromElements(
				"$unwind", bsonutil.String("$a.b"),
			),
		},
		{
			`{"$unwind": { "path": "$a.b", "preserveNullAndEmptyArrays": true} }`,
			bsonutil.DocumentFromElements(
				"$unwind", bsonutil.DocumentFromElements(
					"path", bsonutil.String("$a.b"),
					"preserveNullAndEmptyArrays", bsonutil.True,
				),
			),
		},
		{
			`{"$unwind": { "path": "$a.b", "includeArrayIndex": "funny" } }`,
			bsonutil.DocumentFromElements(
				"$unwind", bsonutil.DocumentFromElements(
					"path", bsonutil.String("$a.b"),
					"includeArrayIndex", bsonutil.String("funny"),
				),
			),
		},
		{
			`{"$unwind": { "path": "$a.b", "includeArrayIndex": "funny", "preserveNullAndEmptyArrays": true} }`,
			bsonutil.DocumentFromElements(
				"$unwind", bsonutil.DocumentFromElements(
					"path", bsonutil.String("$a.b"),
					"includeArrayIndex", bsonutil.String("funny"),
					"preserveNullAndEmptyArrays", bsonutil.True,
				),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			stage := parsertest.ParseStage(tc.input)

			actual := parser.DeparseStage(stage)

			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("bson is not equal\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}
