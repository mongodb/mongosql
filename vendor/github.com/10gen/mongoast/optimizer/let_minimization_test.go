package optimizer_test

import (
	"context"
	"testing"

	"github.com/10gen/mongoast/internal/testutil"
	"github.com/10gen/mongoast/optimizer"
	"github.com/10gen/mongoast/parser"

	"github.com/google/go-cmp/cmp"
)

func TestLetMinimization(t *testing.T) {
	testCases := testutil.LoadTestCases("let_minimization.json")

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
			actual := optimizer.RunPasses(context.Background(), in, optimizer.LetMinimization)

			expectedStr := parser.DeparsePipeline(expected).String()
			actualStr := parser.DeparsePipeline(actual).String()

			if !cmp.Equal(expectedStr, actualStr) {
				t.Fatalf("\nexpected:\n %s\ngot:\n %s", expectedStr, actualStr)
			}
		})
	}
}
