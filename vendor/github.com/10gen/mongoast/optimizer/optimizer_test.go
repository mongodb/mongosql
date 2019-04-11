package optimizer_test

import (
	"context"
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/parsertest"
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
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"ensure optimizer runs DCE",
			`[
				{"$project": {"a": "$c", "b": "$d"}},
				{"$project": {"out": "$a"}}
			 ]`,
			`[
				{"$project": {"a": "$c"}},
				{"$project": {"out": "$a"}}
			 ]`,
		},
		{
			"ensure optimizer runs Reorder Pass",
			`[
				{"$sort": {"a": 1}},
				{"$match": {"a": 1}}
			 ]`,
			`[
				{"$match": {"a": 1}},
				{"$sort": {"a": 1}}
			 ]`,
		},
		{
			"ensure optimizer runs PRE",
			`[
				{"$project":
					{"a":
					    {"$sub":
						[
							{"$add": [
								{"$add": ["$c", "$d"]},
								{"$add": ["$c", "$d"]},
								{"$add": ["$c", "$d"]}
								]
							},
							{"$add": [
								{"$add": ["$c", "$d"]},
								{"$add": ["$c", "$d"]},
								{"$add": ["$c", "$d"]}
								]
							}
						]
						}
					}
				}
			 ]`,
			`[
			  {"$project":
			  	{"a":
					{"$let":
						{"vars": {"mongoast__deduplicated__expr__0": {"$add": ["$c","$d"]}},
						 "in":
						 	{"$let":
								{"vars": {"mongoast__deduplicated__expr__1":
										{"$add": ["$$mongoast__deduplicated__expr__0",
													"$$mongoast__deduplicated__expr__0",
													"$$mongoast__deduplicated__expr__0"]}},
									"in": {"$sub": ["$$mongoast__deduplicated__expr__1","$$mongoast__deduplicated__expr__1"]}
								}
							}
						}
					}
				}
			 }
			 ]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in := parsertest.ParsePipeline(tc.input)
			expected := parsertest.ParsePipeline(tc.expected)
			actual := optimizer.Optimize(context.Background(), in)

			expectedStr := parser.DeparsePipeline(expected).String()
			actualStr := parser.DeparsePipeline(actual).String()

			if !cmp.Equal(expectedStr, actualStr) {
				t.Fatalf("\nexpected:\n %s\ngot:\n %s", expectedStr, actualStr)
			}
		})
	}
}
