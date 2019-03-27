package analyzer_test

import (
	"testing"

	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/10gen/mongoast/internal/parsertest"

	"github.com/google/go-cmp/cmp"
)

func TestSplitPredicates(t *testing.T) {
	testCases := []struct {
		input    string
		expected []ast.Expr
	}{
		{
			`1`,
			[]ast.Expr{ast.NewConstant(bsonutil.Int32(1))},
		},
		{
			`{ "$and": [1, 2] }`,
			[]ast.Expr{
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(2)),
			},
		},
		{
			`{ "$or": [1, 2] }`,
			[]ast.Expr{
				ast.NewBinary(ast.Or,
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(2)),
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			expr := parsertest.ParseExpr(tc.input)

			actual := analyzer.SplitPredicate(expr)
			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("predicate splits are not equal\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}
