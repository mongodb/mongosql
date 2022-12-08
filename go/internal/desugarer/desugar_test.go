package desugarer

import (
	"context"
	"testing"

	"github.com/10gen/mongoast/optimizer"
	"github.com/10gen/mongoast/parser"
	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestDesugar(t *testing.T) {

	t.Run("all", func(t *testing.T) {
		testCases := LoadTestCases("desugar_all.json")

		for _, tc := range testCases {
			t.Run(tc.Name, func(t *testing.T) {
				if tc.Skip != nil {
					t.Skip(*tc.Skip)
				}

				// we will use 'currentDBName' as the current Database.
				actual, err := Desugar(tc.Input, "currentDBName")
				if err != nil {
					t.Fatalf("Failed to desugar pipeline: %v", err)
				}

				expected, err := parser.ParsePipeline(tc.Expected)
				if err != nil {
					t.Fatalf("Failed to parse expected pipeline: %v", err)
				}

				expectedStr := parser.DeparsePipeline(expected).String()
				actualStr := bsoncore.Array(actual).String()

				if !cmp.Equal(expectedStr, actualStr) {
					t.Fatalf("\nexpected:\n %s\ngot:\n %s", expectedStr, actualStr)
				}
			})
		}
	})

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
			name:      "desugarMatchNullSemantics",
			file:      "desugar_sql_match_null_semantics.json",
			desugarer: desugarMatch,
		},
		{
			name:      "desugarNullSemantics",
			file:      "desugar_sql_null_semantics.json",
			desugarer: desugarSQLNullSemantics,
		},
		{
			name: "simplifyLookups",
			file: "simplify_lookups.json",
			// we will use 'currentDBName' as the current Database.
			desugarer: getSimplifyLookups("currentDBName"),
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
