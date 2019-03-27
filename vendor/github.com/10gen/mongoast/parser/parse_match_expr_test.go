package parser_test

import (
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/10gen/mongoast/internal/parsertest"

	"github.com/google/go-cmp/cmp"
)

func TestParseMatchExpr(t *testing.T) {
	testCases := []struct {
		input    string
		expected ast.Expr
	}{
		{
			`{}`,
			ast.NewConstant(bsonutil.True),
		},
		{
			`{"a": 1}`,
			ast.NewBinary(ast.Equals,
				ast.NewFieldRef("a", nil),
				ast.NewConstant(bsonutil.Int32(1)),
			),
		},
		{
			`{"a.b": 1}`,
			ast.NewBinary(ast.Equals,
				ast.NewFieldRef("b", ast.NewFieldRef("a", nil)),
				ast.NewConstant(bsonutil.Int32(1)),
			),
		},
		{
			`{"a.3": 1}`,
			ast.NewBinary(ast.Equals,
				ast.NewFieldOrArrayIndexRef(3, ast.NewFieldRef("a", nil)),
				ast.NewConstant(bsonutil.Int32(1)),
			),
		},
		{
			`{"a.3.b": 1}`,
			ast.NewBinary(ast.Equals,
				ast.NewFieldRef("b", ast.NewFieldOrArrayIndexRef(3, ast.NewFieldRef("a", nil))),
				ast.NewConstant(bsonutil.Int32(1)),
			),
		},
		{
			`{"a.$[3].b": 1}`,
			ast.NewBinary(ast.Equals,
				ast.NewFieldRef(
					"b", ast.NewArrayIndexRef(
						ast.NewConstant(bsonutil.Int32(3)),
						ast.NewFieldRef("a", nil),
					),
				),
				ast.NewConstant(bsonutil.Int32(1)),
			),
		},
		{
			`{"a": {"$eq": 1}}`,
			ast.NewBinary(ast.Equals,
				ast.NewFieldRef("a", nil),
				ast.NewConstant(bsonutil.Int32(1)),
			),
		},
		{
			`{"a": {"$eq": 1, "$gt": 3}}`,
			ast.NewBinary(ast.And,
				ast.NewBinary(ast.Equals,
					ast.NewFieldRef("a", nil),
					ast.NewConstant(bsonutil.Int32(1)),
				),
				ast.NewBinary(ast.GreaterThan,
					ast.NewFieldRef("a", nil),
					ast.NewConstant(bsonutil.Int32(3)),
				),
			),
		},
		{
			`{"$and": [{"a": {"$eq": 1}}, {"a": {"$gt": 3}}]}`,
			ast.NewBinary(ast.And,
				ast.NewBinary(ast.Equals,
					ast.NewFieldRef("a", nil),
					ast.NewConstant(bsonutil.Int32(1)),
				),
				ast.NewBinary(ast.GreaterThan,
					ast.NewFieldRef("a", nil),
					ast.NewConstant(bsonutil.Int32(3)),
				),
			),
		},
		{
			`{"$or": [{"a": 1}, {"b": 2}, {"c": 3}]}`,
			ast.NewBinary(ast.Or,
				ast.NewBinary(ast.Or,
					ast.NewBinary(ast.Equals,
						ast.NewFieldRef("a", nil),
						ast.NewConstant(bsonutil.Int32(1)),
					),
					ast.NewBinary(ast.Equals,
						ast.NewFieldRef("b", nil),
						ast.NewConstant(bsonutil.Int32(2)),
					),
				),
				ast.NewBinary(ast.Equals,
					ast.NewFieldRef("c", nil),
					ast.NewConstant(bsonutil.Int32(3)),
				),
			),
		},
		{
			`{"$nor": [{"a": 1}, {"b": 2}, {"c": 3}]}`,
			ast.NewBinary(ast.Nor,
				ast.NewBinary(ast.Nor,
					ast.NewBinary(ast.Equals,
						ast.NewFieldRef("a", nil),
						ast.NewConstant(bsonutil.Int32(1)),
					),
					ast.NewBinary(ast.Equals,
						ast.NewFieldRef("b", nil),
						ast.NewConstant(bsonutil.Int32(2)),
					),
				),
				ast.NewBinary(ast.Equals,
					ast.NewFieldRef("c", nil),
					ast.NewConstant(bsonutil.Int32(3)),
				),
			),
		},
		{
			`{"$expr": {"$eq": ["$a", 1]}}`,
			ast.NewAggExpr(
				ast.NewBinary(ast.Equals,
					ast.NewFieldRef("a", nil),
					ast.NewConstant(bsonutil.Int32(1)),
				),
			),
		},
		{
			`{"a": {"$in": [1, 2, 3]}}`,
			ast.NewFunction(
				"$in",
				ast.NewArray(
					ast.NewFieldRef("a", nil),
					ast.NewUnknown(bsonutil.ArrayFromValues(
						bsonutil.Int32(1),
						bsonutil.Int32(2),
						bsonutil.Int32(3),
					)),
				),
			),
		},
		{
			`{"a": {"$eee": 1, "$gt": 3}}`,
			ast.NewBinary(ast.And,
				ast.NewFunction(
					"$eee",
					ast.NewArray(
						ast.NewFieldRef("a", nil),
						ast.NewUnknown(bsonutil.Int32(1)),
					),
				),
				ast.NewBinary(ast.GreaterThan,
					ast.NewFieldRef("a", nil),
					ast.NewConstant(bsonutil.Int32(3)),
				),
			),
		},
		{
			`{"$eee": 1}`,
			ast.NewFunction(
				"$eee",
				ast.NewUnknown(bsonutil.Int32(1)),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual := parsertest.ParseMatchExpr(tc.input)

			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("stages are not equal\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}
