package parser_test

import (
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/10gen/mongoast/internal/parsertest"

	"github.com/google/go-cmp/cmp"
)

func TestParseExpr(t *testing.T) {
	testCases := []struct {
		input    string
		expected ast.Expr
	}{
		// Constants
		{
			`null`,
			ast.NewConstant(bsonutil.Null()),
		},
		{
			`1`,
			ast.NewConstant(bsonutil.Int32(1)),
		},
		{
			`"a"`,
			ast.NewConstant(bsonutil.String("a")),
		},
		{
			`{"$numberDecimal": "1"}`,
			ast.NewConstant(bsonutil.Decimal128FromInt64(1)),
		},
		{
			`{"a": 1}`,
			ast.NewDocument(
				ast.NewDocumentElement("a", ast.NewConstant(bsonutil.Int32(1))),
			),
		},
		{
			`[1,2]`,
			ast.NewArray(
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(2)),
			),
		},
		// Variables
		{
			`"$$a"`,
			ast.NewVariableRef("a"),
		},
		{
			`"$$a.b"`,
			ast.NewFieldRef("b", ast.NewVariableRef("a")),
		},
		{
			`"$$a.b.c"`,
			ast.NewFieldRef("c", ast.NewFieldRef("b", ast.NewVariableRef("a"))),
		},
		// FieldRef
		{
			`"$a"`,
			ast.NewFieldRef("a", nil),
		},
		{
			`"$a.b"`,
			ast.NewFieldRef("b", ast.NewFieldRef("a", nil)),
		},
		{
			`"$a.b.c"`,
			ast.NewFieldRef("c", ast.NewFieldRef("b", ast.NewFieldRef("a", nil))),
		},
		// Logical
		{
			`{ "$and": [1] }`,
			ast.NewConstant(bsonutil.Int32(1)),
		},
		{
			`{ "$and": [1, 0] }`,
			ast.NewBinary(
				ast.And,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
		},
		{
			`{ "$and": [1, 0, 2] }`,
			ast.NewBinary(
				ast.And,
				ast.NewBinary(
					ast.And,
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(0)),
				),
				ast.NewConstant(bsonutil.Int32(2)),
			),
		},
		{
			`{ "$or": [1, 0, 2] }`,
			ast.NewBinary(
				ast.Or,
				ast.NewBinary(
					ast.Or,
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(0)),
				),
				ast.NewConstant(bsonutil.Int32(2)),
			),
		},
		// Comparisons
		{
			`{ "$eq": [1, 0] }`,
			ast.NewBinary(
				ast.Equals,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
		},
		{
			`{ "$gt": [1, 0] }`,
			ast.NewBinary(
				ast.GreaterThan,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
		},
		{
			`{ "$gte": [1, 0] }`,
			ast.NewBinary(
				ast.GreaterThanOrEquals,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
		},
		{
			`{ "$lt": [1, 0] }`,
			ast.NewBinary(
				ast.LessThan,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
		},
		{
			`{ "$lte": [1, 0] }`,
			ast.NewBinary(
				ast.LessThanOrEquals,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
		},
		{
			`{ "$ne": [1, 0] }`,
			ast.NewBinary(
				ast.NotEquals,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
		},
		{
			`{ "$ne": [1, {}] }`,
			ast.NewBinary(
				ast.NotEquals,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewDocument(),
			),
		},
		{
			`{ "$or": [1] }`,
			ast.NewConstant(bsonutil.Int32(1)),
		},
		{
			`{ "$or": [1, 0] }`,
			ast.NewBinary(
				ast.Or,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
		},
		// Function
		{
			`{ "$arrayElemAt": ["$a", 2] }`,
			ast.NewArrayIndexRef(
				ast.NewConstant(bsonutil.Int32(2)),
				ast.NewFieldRef("a", nil),
			),
		},
		{
			`{ "$arrayElemAt": ["$a", "$b"] }`,
			ast.NewArrayIndexRef(
				ast.NewFieldRef("b", nil),
				ast.NewFieldRef("a", nil),
			),
		},
		{
			`{ "$sum": "$a" }`,
			ast.NewFunction("$sum", ast.NewFieldRef("a", nil)),
		},
		{
			`{ "$sum": ["$a"] }`,
			ast.NewFunction("$sum", ast.NewArray(ast.NewFieldRef("a", nil))),
		},
		{
			`{ "$ltrim": { "input": "$a", "chars": "abc" } }`,
			ast.NewFunction(
				"$ltrim",
				ast.NewDocument(
					ast.NewDocumentElement("input", ast.NewFieldRef("a", nil)),
					ast.NewDocumentElement("chars", ast.NewConstant(bsonutil.String("abc"))),
				),
			),
		},
		// Let
		{
			`{ "$let": { "vars": { "a": 1, "b": "$x" }, "in": { "$sum": ["$$a", "$$b"] } } }`,
			ast.NewLet(
				[]*ast.LetVariable{
					ast.NewLetVariable("a", ast.NewConstant(bsonutil.Int32(1))),
					ast.NewLetVariable("b", ast.NewFieldRef("x", nil)),
				},
				ast.NewFunction(
					"$sum",
					ast.NewArray(
						ast.NewVariableRef("a"),
						ast.NewVariableRef("b"),
					),
				),
			),
		},
		// Conditional
		{
			`{ "$cond": { "if": { "$eq": ["$a", 5] }, "then": 1, "else": 0 } }`,
			ast.NewConditional(
				ast.NewBinary(
					ast.Equals,
					ast.NewFieldRef("a", nil),
					ast.NewConstant(bsonutil.Int32(5)),
				),
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
		},
		{
			`{ "$cond": [{ "$eq": ["$a", 5] }, 1, 0] }`,
			ast.NewConditional(
				ast.NewBinary(
					ast.Equals,
					ast.NewFieldRef("a", nil),
					ast.NewConstant(bsonutil.Int32(5)),
				),
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual := parsertest.ParseExpr(tc.input)

			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("pipelines are not equal\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}
