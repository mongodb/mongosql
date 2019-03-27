package parser_test

import (
	"testing"

	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/10gen/mongoast/internal/parsertest"
	"github.com/10gen/mongoast/parser"

	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestDeparseMatchExpr(t *testing.T) {
	testCases := []struct {
		input    string
		expected bsoncore.Value
	}{
		// Empty
		{
			`{}`,
			bsonutil.EmptyDocument(),
		},
		// Fields
		{
			`{"a": 1}`,
			bsonutil.DocumentFromElements(
				"a", bsonutil.DocumentFromElements(
					"$eq", bsonutil.Int32(1),
				),
			),
		},
		{
			`{"a.b": 1}`,
			bsonutil.DocumentFromElements(
				"a.b", bsonutil.DocumentFromElements(
					"$eq", bsonutil.Int32(1),
				),
			),
		},
		{
			`{"a.2": 1}`,
			bsonutil.DocumentFromElements(
				"a.2", bsonutil.DocumentFromElements(
					"$eq", bsonutil.Int32(1),
				),
			),
		},
		{
			`{"a.$[2]": 1}`,
			bsonutil.DocumentFromElements(
				"a.2", bsonutil.DocumentFromElements(
					"$eq", bsonutil.Int32(1),
				),
			),
		},
		// Document
		{
			`{"a": { "b": 1 } }`,
			bsonutil.DocumentFromElements(
				"a", bsonutil.DocumentFromElements(
					"$eq", bsonutil.DocumentFromElements(
						"b", bsonutil.Int32(1),
					),
				),
			),
		},
		// Logical
		{
			`{"a": 1, "b": 2}`,
			bsonutil.DocumentFromElements(
				"$and", bsonutil.ArrayFromValues(
					bsonutil.DocumentFromElements(
						"a", bsonutil.DocumentFromElements(
							"$eq", bsonutil.Int32(1),
						),
					),
					bsonutil.DocumentFromElements(
						"b", bsonutil.DocumentFromElements(
							"$eq", bsonutil.Int32(2),
						),
					),
				),
			),
		},
		{
			`{"$and": [{"a": 1}, {"b": 2}]}`,
			bsonutil.DocumentFromElements(
				"$and", bsonutil.ArrayFromValues(
					bsonutil.DocumentFromElements(
						"a", bsonutil.DocumentFromElements(
							"$eq", bsonutil.Int32(1),
						),
					),
					bsonutil.DocumentFromElements(
						"b", bsonutil.DocumentFromElements(
							"$eq", bsonutil.Int32(2),
						),
					),
				),
			),
		},
		{
			`{"$or": [{"a": 1}, {"b": 2}]}`,
			bsonutil.DocumentFromElements(
				"$or", bsonutil.ArrayFromValues(
					bsonutil.DocumentFromElements(
						"a", bsonutil.DocumentFromElements(
							"$eq", bsonutil.Int32(1),
						),
					),
					bsonutil.DocumentFromElements(
						"b", bsonutil.DocumentFromElements(
							"$eq", bsonutil.Int32(2),
						),
					),
				),
			),
		},
		{
			`{"$and": [{"a": 1}, {"b": 2}, {"c": 3}]}`,
			bsonutil.DocumentFromElements(
				"$and", bsonutil.ArrayFromValues(
					bsonutil.DocumentFromElements(
						"a", bsonutil.DocumentFromElements(
							"$eq", bsonutil.Int32(1),
						),
					),
					bsonutil.DocumentFromElements(
						"b", bsonutil.DocumentFromElements(
							"$eq", bsonutil.Int32(2),
						),
					),
					bsonutil.DocumentFromElements(
						"c", bsonutil.DocumentFromElements(
							"$eq", bsonutil.Int32(3),
						),
					),
				),
			),
		},
		{
			`{"$or": [{"a": 1}, {"b": 2}, {"c": 3}]}`,
			bsonutil.DocumentFromElements(
				"$or", bsonutil.ArrayFromValues(
					bsonutil.DocumentFromElements(
						"a", bsonutil.DocumentFromElements(
							"$eq", bsonutil.Int32(1),
						),
					),
					bsonutil.DocumentFromElements(
						"b", bsonutil.DocumentFromElements(
							"$eq", bsonutil.Int32(2),
						),
					),
					bsonutil.DocumentFromElements(
						"c", bsonutil.DocumentFromElements(
							"$eq", bsonutil.Int32(3),
						),
					),
				),
			),
		},
		{
			`{"$nor": [{"a": 1}, {"b": 2}, {"c": 3}]}`,
			bsonutil.DocumentFromElements(
				"$nor", bsonutil.ArrayFromValues(
					bsonutil.DocumentFromElements(
						"a", bsonutil.DocumentFromElements(
							"$eq", bsonutil.Int32(1),
						),
					),
					bsonutil.DocumentFromElements(
						"b", bsonutil.DocumentFromElements(
							"$eq", bsonutil.Int32(2),
						),
					),
					bsonutil.DocumentFromElements(
						"c", bsonutil.DocumentFromElements(
							"$eq", bsonutil.Int32(3),
						),
					),
				),
			),
		},
		// Comparison
		{
			`{"a": { "$eq": 1}}`,
			bsonutil.DocumentFromElements(
				"a", bsonutil.DocumentFromElements(
					"$eq", bsonutil.Int32(1),
				),
			),
		},
		{
			`{"a": { "$gt": 1}}`,
			bsonutil.DocumentFromElements(
				"a", bsonutil.DocumentFromElements(
					"$gt", bsonutil.Int32(1),
				),
			),
		},
		{
			`{"a": { "$gte": 1}}`,
			bsonutil.DocumentFromElements(
				"a", bsonutil.DocumentFromElements(
					"$gte", bsonutil.Int32(1),
				),
			),
		},
		{
			`{"a": { "$lt": 1}}`,
			bsonutil.DocumentFromElements(
				"a", bsonutil.DocumentFromElements(
					"$lt", bsonutil.Int32(1),
				),
			),
		},
		{
			`{"a": { "$lte": 1}}`,
			bsonutil.DocumentFromElements(
				"a", bsonutil.DocumentFromElements(
					"$lte", bsonutil.Int32(1),
				),
			),
		},
		{
			`{"a": { "$ne": 1}}`,
			bsonutil.DocumentFromElements(
				"a", bsonutil.DocumentFromElements(
					"$ne", bsonutil.Int32(1),
				),
			),
		},
		// Aggregation Expression
		{
			`{"$expr": { "$eq": ["$a", 1]}}`,
			bsonutil.DocumentFromElements(
				"$expr", bsonutil.DocumentFromElements(
					"$eq", bsonutil.ArrayFromValues(
						bsonutil.String("$a"),
						bsonutil.DocumentFromElements(
							"$literal", bsonutil.Int32(1),
						),
					),
				),
			),
		},
		// Unknown
		{
			`{"a": { "$eee": 1}}`,
			bsonutil.DocumentFromElements(
				"a", bsonutil.DocumentFromElements(
					"$eee", bsonutil.Int32(1),
				),
			),
		},
		{
			`{"$eee": [{"a": 1}, {"b": 2}]}`,
			bsonutil.DocumentFromElements(
				"$eee", bsonutil.ArrayFromValues(
					bsonutil.DocumentFromElements(
						"a", bsonutil.Int32(1),
					),
					bsonutil.DocumentFromElements(
						"b", bsonutil.Int32(2),
					),
				),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			expr := parsertest.ParseMatchExpr(tc.input)

			actual := parser.DeparseMatchExpr(expr)

			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("bson is not equal\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}
