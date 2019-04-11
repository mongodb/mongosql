package analyzer_test

import (
	"testing"

	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/internal/parsertest"

	"github.com/google/go-cmp/cmp"
)

func TestReferencedVariables(t *testing.T) {
	testCases := []struct {
		input         string
		expectedNames []string
	}{
		{
			`{"$add": ["$$a", "$$b.c.d", {"$arrayElemAt": ["$$foo", 2]}, {"$arrayElemAt": ["$bar", 2]}]}`,
			[]string{"a", "b", "foo"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			expr := parsertest.ParseExpr(tc.input)

			actual := analyzer.ReferencedVariableRoots(expr)
			if !cmp.Equal(tc.expectedNames, actual) {
				t.Fatalf("root name sets are not equal\n  %s", cmp.Diff(tc.expectedNames, actual))
			}
		})
	}
}
