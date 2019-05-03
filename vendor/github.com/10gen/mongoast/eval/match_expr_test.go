package eval_test

import (
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/eval"
	"github.com/10gen/mongoast/internal/bsonutil"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestEvaluateMatchExpr(t *testing.T) {
	testCases := []struct {
		name     string
		doc      bsoncore.Value
		expr     ast.Expr
		expected bool
	}{
		{
			"string match",
			bsonutil.DocumentFromElements(
				"name", bsonutil.String("foo"),
			),
			ast.NewBinary(
				ast.Equals,
				ast.NewFieldRef("name", nil),
				ast.NewConstant(bsonutil.String("foo")),
			),
			true,
		},
		{
			"string no-match",
			bsonutil.DocumentFromElements(
				"name", bsonutil.String("foo"),
			),
			ast.NewBinary(
				ast.Equals,
				ast.NewFieldRef("name", nil),
				ast.NewConstant(bsonutil.String("bar")),
			),
			false,
		},
		{
			"long match",
			bsonutil.DocumentFromElements(
				"num", bsonutil.Int64(0),
			),
			ast.NewBinary(
				ast.Equals,
				ast.NewFieldRef("num", nil),
				ast.NewConstant(bsonutil.Int64(0)),
			),
			true,
		},
		{
			"long non-match",
			bsonutil.DocumentFromElements(
				"num", bsonutil.Int64(0),
			),
			ast.NewBinary(
				ast.Equals,
				ast.NewFieldRef("num", nil),
				ast.NewConstant(bsonutil.Int64(1)),
			),
			false,
		},
		{
			"date match",
			bsonutil.DocumentFromElements(
				"date", bsonutil.DateTime(0),
			),
			ast.NewBinary(
				ast.Equals,
				ast.NewFieldRef("date", nil),
				ast.NewConstant(bsonutil.DateTime(0)),
			),
			true,
		},
		{
			"date non-match",
			bsonutil.DocumentFromElements(
				"date", bsonutil.DateTime(0),
			),
			ast.NewBinary(
				ast.Equals,
				ast.NewFieldRef("date", nil),
				ast.NewConstant(bsonutil.DateTime(1)),
			),
			false,
		},
		{
			"missing field match",
			bsonutil.EmptyDocument(),
			ast.NewBinary(
				ast.Equals,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Null()),
			),
			true,
		},
		{
			"missing field non-match",
			bsonutil.EmptyDocument(),
			ast.NewBinary(
				ast.Equals,
				ast.NewFieldRef("x", nil),
				ast.NewConstant(bsonutil.Int32(0)),
			),
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := eval.EvaluateMatchExpr(tc.expr, tc.doc)
			if err != nil {
				t.Fatalf("expected no error, but got %v", err)
			}

			if tc.expected != actual {
				t.Fatalf("expected %v, but got %v", tc.expected, actual)
			}
		})
	}
}
