package parser_test

import (
	"testing"

	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/10gen/mongoast/internal/parsertest"
	"github.com/10gen/mongoast/parser"

	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestDeparseExpr(t *testing.T) {
	testCases := []struct {
		input    string
		expected bsoncore.Value
	}{
		// Variables
		{
			`"$$a"`,
			bsonutil.String("$$a"),
		},
		{
			`"$$a.b"`,
			bsonutil.String("$$a.b"),
		},
		{
			`"$$a.b.c"`,
			bsonutil.String("$$a.b.c"),
		},
		// Fields
		{
			`"$a"`,
			bsonutil.String("$a"),
		},
		{
			`"$a.b"`,
			bsonutil.String("$a.b"),
		},
		{
			`{"$arrayElemAt": ["$a", 2]}`,
			bsonutil.DocumentFromElements(
				"$arrayElemAt", bsonutil.ArrayFromValues(
					bsonutil.String("$a"),
					bsonutil.DocumentFromElements(
						"$literal", bsonutil.Int32(2),
					),
				),
			),
		},
		{
			`{"$arrayElemAt": ["$a", "$b"]}`,
			bsonutil.DocumentFromElements(
				"$arrayElemAt", bsonutil.ArrayFromValues(
					bsonutil.String("$a"),
					bsonutil.String("$b"),
				),
			),
		},
		// Document
		{
			`{"a": 1 }`,
			bsonutil.DocumentFromElements(
				"a", bsonutil.DocumentFromElements(
					"$literal",
					bsonutil.Int32(1),
				),
			),
		},
		// Logical
		{
			`{"$and": [1, 2]}`,
			bsonutil.DocumentFromElements(
				"$and", bsonutil.ArrayFromValues(
					bsonutil.DocumentFromElements(
						"$literal", bsonutil.Int32(1),
					),
					bsonutil.DocumentFromElements(
						"$literal", bsonutil.Int32(2),
					),
				),
			),
		},
		{
			`{"$or": [1, 2]}`,
			bsonutil.DocumentFromElements(
				"$or", bsonutil.ArrayFromValues(
					bsonutil.DocumentFromElements(
						"$literal", bsonutil.Int32(1),
					),
					bsonutil.DocumentFromElements(
						"$literal", bsonutil.Int32(2),
					),
				),
			),
		},
		// Comparison
		{
			`{"$eq": ["$a", 1]}`,
			bsonutil.DocumentFromElements(
				"$eq", bsonutil.ArrayFromValues(
					bsonutil.String("$a"),
					bsonutil.DocumentFromElements(
						"$literal", bsonutil.Int32(1),
					),
				),
			),
		},
		{
			`{"$gt": ["$a", 1]}`,
			bsonutil.DocumentFromElements(
				"$gt", bsonutil.ArrayFromValues(
					bsonutil.String("$a"),
					bsonutil.DocumentFromElements(
						"$literal", bsonutil.Int32(1),
					),
				),
			),
		},
		{
			`{"$gte": ["$a", 1]}`,
			bsonutil.DocumentFromElements(
				"$gte", bsonutil.ArrayFromValues(
					bsonutil.String("$a"),
					bsonutil.DocumentFromElements(
						"$literal", bsonutil.Int32(1),
					),
				),
			),
		},
		{
			`{"$lt": ["$a", 1]}`,
			bsonutil.DocumentFromElements(
				"$lt", bsonutil.ArrayFromValues(
					bsonutil.String("$a"),
					bsonutil.DocumentFromElements(
						"$literal", bsonutil.Int32(1),
					),
				),
			),
		},
		{
			`{"$lte": ["$a", 1]}`,
			bsonutil.DocumentFromElements(
				"$lte", bsonutil.ArrayFromValues(
					bsonutil.String("$a"),
					bsonutil.DocumentFromElements(
						"$literal", bsonutil.Int32(1),
					),
				),
			),
		},
		{
			`{"$ne": ["$a", 1]}`,
			bsonutil.DocumentFromElements(
				"$ne", bsonutil.ArrayFromValues(
					bsonutil.String("$a"),
					bsonutil.DocumentFromElements(
						"$literal", bsonutil.Int32(1),
					),
				),
			),
		},
		// Known Functions
		{
			`{"$sum": "$a"}`,
			bsonutil.DocumentFromElements(
				"$sum", bsonutil.String("$a"),
			),
		},
		// Let
		{
			`{"$let": {"vars": {"a": 1, "b": "$x"}, "in": {"$sum": ["$$a", "$$b"]}}}`,
			bsonutil.DocumentFromElements(
				"$let", bsonutil.DocumentFromElements(
					"vars", bsonutil.DocumentFromElements(
						"a", bsonutil.DocumentFromElements(
							"$literal", bsonutil.Int32(1),
						),
						"b", bsonutil.String("$x"),
					),
					"in", bsonutil.DocumentFromElements(
						"$sum", bsonutil.ArrayFromValues(
							bsonutil.String("$$a"),
							bsonutil.String("$$b"),
						),
					),
				),
			),
		},
		// Conditional
		{
			`{"$cond": {"if": {"$eq": ["$a", 5]}, "then": 1, "else": 0}}`,
			bsonutil.DocumentFromElements(
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
		},
		// Unknown
		{
			`{"a": { "$eee": 1}}`,
			bsonutil.DocumentFromElements(
				"a", bsonutil.DocumentFromElements(
					"$eee", bsonutil.DocumentFromElements(
						"$literal", bsonutil.Int32(1),
					),
				),
			),
		},
		{
			`{"$eee": [{"a": 1}, {"b": 2}]}`,
			bsonutil.DocumentFromElements(
				"$eee", bsonutil.ArrayFromValues(
					bsonutil.DocumentFromElements(
						"a", bsonutil.DocumentFromElements(
							"$literal", bsonutil.Int32(1),
						),
					),
					bsonutil.DocumentFromElements(
						"b", bsonutil.DocumentFromElements(
							"$literal", bsonutil.Int32(2),
						),
					),
				),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			expr := parsertest.ParseExpr(tc.input)

			actual := parser.DeparseExpr(expr)

			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("bson is not equal\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}
