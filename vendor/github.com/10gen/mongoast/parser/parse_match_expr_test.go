package parser_test

import (
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/10gen/mongoast/parser"
	"github.com/pkg/errors"

	"github.com/google/go-cmp/cmp"
)

func TestParseMatchExpr(t *testing.T) {
	testCases := []struct {
		input    string
		expected ast.Expr
		err      error
	}{
		{
			`{}`,
			ast.NewConstant(bsonutil.True),
			nil,
		},
		{
			`{"a": 1}`,
			ast.NewBinary(ast.Equals,
				ast.NewFieldRef("a", nil),
				ast.NewConstant(bsonutil.Int32(1)),
			),
			nil,
		},
		{
			`{"": 1}`,
			nil,
			errors.New("invalid match expression key"),
		},
		{
			`{"a.b": 1}`,
			ast.NewBinary(ast.Equals,
				ast.NewFieldRef("b", ast.NewFieldRef("a", nil)),
				ast.NewConstant(bsonutil.Int32(1)),
			),
			nil,
		},
		{
			`{"a.3": 1}`,
			ast.NewBinary(ast.Equals,
				ast.NewFieldOrArrayIndexRef(3, ast.NewFieldRef("a", nil)),
				ast.NewConstant(bsonutil.Int32(1)),
			),
			nil,
		},
		{
			`{"a.3.b": 1}`,
			ast.NewBinary(ast.Equals,
				ast.NewFieldRef("b", ast.NewFieldOrArrayIndexRef(3, ast.NewFieldRef("a", nil))),
				ast.NewConstant(bsonutil.Int32(1)),
			),
			nil,
		},
		{
			`{".": 1}`,
			nil,
			errors.New("failed parsing . as a field ref: invalid field ref"),
		},
		{
			`{"a": {"$eq": 1}}`,
			ast.NewBinary(ast.Equals,
				ast.NewFieldRef("a", nil),
				ast.NewConstant(bsonutil.Int32(1)),
			),
			nil,
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
			nil,
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
			nil,
		},
		{
			`{"$and": 1}`,
			nil,
			errors.New("$and should have an array value"),
		},
		{
			`{"$and": [1, 2]}`,
			nil,
			errors.New("$and array elements must be documents"),
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
			nil,
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
			nil,
		},
		{
			`{"$expr": {"$eq": ["$a", 1]}}`,
			ast.NewAggExpr(
				ast.NewBinary(ast.Equals,
					ast.NewFieldRef("a", nil),
					ast.NewConstant(bsonutil.Int32(1)),
				),
			),
			nil,
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
			nil,
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
			nil,
		},
		{
			`{"$eee": 1}`,
			ast.NewFunction(
				"$eee",
				ast.NewUnknown(bsonutil.Int32(1)),
			),
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual, err := parser.ParseMatchExprJSON(tc.input)

			if err != nil && tc.err == nil {
				t.Fatalf("err should be nil, but was %v", err)
			} else if err == nil && tc.err != nil {
				t.Fatalf("err should not be nil, expected %v", tc.err)
			} else if err != nil && tc.err != nil && err.Error() != tc.err.Error() {
				t.Fatalf("expected error %q, but got %q", tc.err.Error(), err.Error())
			}

			if tc.err == nil && !cmp.Equal(tc.expected, actual) {
				t.Fatalf("stages are not equal\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}
