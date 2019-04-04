package optimizer_test

import (
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/10gen/mongoast/optimizer"

	"github.com/google/go-cmp/cmp"
)

func TestReorder(t *testing.T) {
	testCases := []struct {
		name     string
		input    *ast.Pipeline
		expected *ast.Pipeline
	}{
		{
			"move limit before project",
			ast.NewPipeline(
				ast.NewProjectStage(
					ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
				),
				ast.NewLimitStage(10),
				ast.NewProjectStage(
					ast.NewIncludeProjectItem(ast.NewFieldRef("b", nil)),
				),
			),
			ast.NewPipeline(
				ast.NewLimitStage(10),
				ast.NewProjectStage(
					ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
				),
				ast.NewProjectStage(
					ast.NewIncludeProjectItem(ast.NewFieldRef("b", nil)),
				),
			),
		},
		{
			"move limit before 2 projects",
			ast.NewPipeline(
				ast.NewProjectStage(
					ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
				),
				ast.NewProjectStage(
					ast.NewIncludeProjectItem(ast.NewFieldRef("b", nil)),
				),
				ast.NewLimitStage(10),
			),
			ast.NewPipeline(
				ast.NewLimitStage(10),
				ast.NewProjectStage(
					ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
				),
				ast.NewProjectStage(
					ast.NewIncludeProjectItem(ast.NewFieldRef("b", nil)),
				),
			),
		},
		{
			"move limit and skip before project",
			ast.NewPipeline(
				ast.NewProjectStage(
					ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
				),
				ast.NewLimitStage(10),
				ast.NewSkipStage(20),
			),
			ast.NewPipeline(
				ast.NewLimitStage(10),
				ast.NewSkipStage(20),
				ast.NewProjectStage(
					ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
				),
			),
		},
		{
			"move match before sort",
			ast.NewPipeline(
				ast.NewSortStage(
					ast.NewSortItem(ast.NewFieldRef("a", nil), false),
				),
				ast.NewMatchStage(ast.NewConstant(bsonutil.True)),
			),
			ast.NewPipeline(
				ast.NewMatchStage(ast.NewConstant(bsonutil.True)),
				ast.NewSortStage(
					ast.NewSortItem(ast.NewFieldRef("a", nil), false),
				),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := optimizer.RunPasses(nil, tc.input, optimizer.Reorder)

			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("pipelines are not equal\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}
