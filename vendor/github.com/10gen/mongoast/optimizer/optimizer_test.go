package optimizer_test

import (
	"context"
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/testutil"
	"github.com/10gen/mongoast/optimizer"
	"github.com/10gen/mongoast/parser"

	"github.com/google/go-cmp/cmp"
)

func TestSubpipelineOptimization(t *testing.T) {
	testCases := []struct {
		name     string
		input    *ast.Pipeline
		expected *ast.Pipeline
	}{
		{
			"move limit before project in lookup subpipeline",
			ast.NewPipeline(
				ast.NewLookupStage("foo", "bar", "baz", "buz", []*ast.LookupLetItem{},
					ast.NewPipeline(
						ast.NewProjectStage(
							ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
						),
						ast.NewLimitStage(10),
						ast.NewProjectStage(
							ast.NewIncludeProjectItem(ast.NewFieldRef("b", nil)),
						),
					))),
			ast.NewPipeline(
				ast.NewLookupStage("foo", "bar", "baz", "buz", []*ast.LookupLetItem{},
					ast.NewPipeline(
						ast.NewLimitStage(10),
						ast.NewProjectStage(
							ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
						),
						ast.NewProjectStage(
							ast.NewIncludeProjectItem(ast.NewFieldRef("b", nil)),
						),
					))),
		},
		{
			"move limit before project in two facet subpipelines",
			ast.NewPipeline(
				ast.NewFacetStage(
					ast.NewFacetItem("foo",
						ast.NewPipeline(
							ast.NewProjectStage(
								ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
							),
							ast.NewLimitStage(10),
							ast.NewProjectStage(
								ast.NewIncludeProjectItem(ast.NewFieldRef("b", nil)),
							),
						)),
					ast.NewFacetItem("bar",
						ast.NewPipeline(
							ast.NewProjectStage(
								ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
							),
							ast.NewLimitStage(10),
							ast.NewProjectStage(
								ast.NewIncludeProjectItem(ast.NewFieldRef("b", nil)),
							),
						)),
				)),
			ast.NewPipeline(
				ast.NewFacetStage(
					ast.NewFacetItem("foo",
						ast.NewPipeline(
							ast.NewLimitStage(10),
							ast.NewProjectStage(
								ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
							),
							ast.NewProjectStage(
								ast.NewIncludeProjectItem(ast.NewFieldRef("b", nil)),
							),
						)),
					ast.NewFacetItem("bar",
						ast.NewPipeline(
							ast.NewLimitStage(10),
							ast.NewProjectStage(
								ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
							),
							ast.NewProjectStage(
								ast.NewIncludeProjectItem(ast.NewFieldRef("b", nil)),
							),
						)),
				)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := optimizer.RunPasses(context.Background(), tc.input, optimizer.Reorder)

			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("pipelines are not equal\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}

func TestPassInclusion(t *testing.T) {
	testCases := testutil.LoadTestCases("pass_inclusion.json")

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
			actual := optimizer.Optimize(context.Background(), in)

			expectedStr := parser.DeparsePipeline(expected).String()
			actualStr := parser.DeparsePipeline(actual).String()

			if !cmp.Equal(expectedStr, actualStr) {
				t.Fatalf("\nexpected:\n %s\ngot:\n %s", expectedStr, actualStr)
			}
		})
	}
}
