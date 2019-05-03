package optimizer_test

import (
	"context"
	"testing"

	"github.com/10gen/mongoast/internal/testutil"
	"github.com/10gen/mongoast/optimizer"
	"github.com/10gen/mongoast/parser"

	"github.com/google/go-cmp/cmp"
)

func TestLetMerging(t *testing.T) {
	// Note: this is just testing LetMerging, so the let variables
	// in these test cases may contain values that would be removed
	// by LetMinimization. The type of expression in the bindings
	// does not matter; the important thing is the dependence on
	// outer Lets.
	testCases := testutil.LoadTestCases("let_merging.json")

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			in, err := parser.ParsePipeline(tc.Input)
			if err != nil {
				t.Fatalf("Failed to parse input pipeline: %v", err)
			}
			expected, err := parser.ParsePipeline(tc.Expected)
			if err != nil {
				t.Fatalf("Failed to parse expected pipeline: %v", err)
			}
			actual := optimizer.RunPasses(context.Background(), in, optimizer.LetMerging)

			expectedStr := parser.DeparsePipeline(expected).String()
			actualStr := parser.DeparsePipeline(actual).String()

			if !cmp.Equal(expectedStr, actualStr) {
				t.Fatalf("\nexpected:\n %s\ngot:\n %s", expectedStr, actualStr)
			}
		})
	}
}
