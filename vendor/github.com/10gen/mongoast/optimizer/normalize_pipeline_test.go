package optimizer_test

import (
	"testing"

	"github.com/10gen/mongoast/internal/testutil"
	"github.com/10gen/mongoast/optimizer"
	"github.com/10gen/mongoast/parser"

	"github.com/google/go-cmp/cmp"
)

func TestNormalizePipeline(t *testing.T) {
	testCases := testutil.LoadTestCases("normalize_pipeline.json")

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
			actual := optimizer.NormalizePipeline(in)

			expectedStr := parser.DeparsePipeline(expected).String()
			actualStr := parser.DeparsePipeline(actual).String()

			if !cmp.Equal(expectedStr, actualStr) {
				t.Fatalf("\nexpected:\n %s\ngot:\n %s", expectedStr, actualStr)
			}
		})
	}
}
