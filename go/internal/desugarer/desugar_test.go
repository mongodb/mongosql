package desugarer

import (
	"context"
	"testing"

	"github.com/10gen/mongoast/optimizer"
	"github.com/10gen/mongoast/parser"
	"github.com/google/go-cmp/cmp"
)

func TestDesugar(t *testing.T) {

	tests := []struct {
		name      string
		file      string
		desugarer optimizer.Optimization
	}{
		{
			name:      "desugarJoins",
			file:      "desugar_joins.json",
			desugarer: desugarJoins,
		},
		{
			name:      "desugarAccumulators",
			file:      "desugar_accumulators.json",
			desugarer: desugarAccumulators,
		},
		{
			name:      "desugarSubqueryExprs",
			file:      "desugar_subquery_expressions.json",
			desugarer: desugarSubqueryExprs,
		},
		{
			name:      "desugarUnsupportedOperators",
			file:      "desugar_unsupported_operators.json",
			desugarer: desugarUnsupportedOperators,
		},
		{
			name:      "desugarNullSemantics",
			file:      "desugar_sql_null_semantics.json",
			desugarer: desugarSQLNullSemantics,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testCases := LoadTestCases(test.file)

			for _, tc := range testCases {
				t.Run(tc.Name, func(t *testing.T) {
					if tc.Skip != nil {
						t.Skip(*tc.Skip)
					}
					in, err := parser.ParsePipeline(tc.Input)
					if err != nil {
						t.Fatalf("Failed to parse input pipeline: %v", err)
					}
					expected, err := parser.ParsePipeline(tc.Expected)
					if err != nil {
						t.Fatalf("Failed to parse expected pipeline: %v", err)
					}
					actual := optimizer.RunPasses(context.Background(), in, 0, test.desugarer)

					expectedStr := parser.DeparsePipeline(expected).String()
					actualStr := parser.DeparsePipeline(actual).String()

					if !cmp.Equal(expectedStr, actualStr) {
						t.Fatalf("\nexpected:\n %s\ngot:\n %s", expectedStr, actualStr)
					}
				})
			}
		})
	}
}
