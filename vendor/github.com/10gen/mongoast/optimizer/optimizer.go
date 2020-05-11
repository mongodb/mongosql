package optimizer

import (
	"context"

	"github.com/10gen/mongoast/ast"
)

// Optimization is the type of a pipeline optimization.
type Optimization = func(pipeline *ast.Pipeline, memoryLimit uint64) *ast.Pipeline

// Optimize handles optimizing a pipeline.
func Optimize(ctx context.Context, pipeline *ast.Pipeline, memoryLimit uint64) *ast.Pipeline {
	return RunPasses(ctx, pipeline, memoryLimit,
		GroupKeyExtraction,
		DeadCodeElimination,
		ProjectionStageCompressionUp,
		ProjectionStageCompressionDown,
		DeadCodeElimination,
		MatchSplit,
		Reorder,
		MatchCoalescing,
		SortCoalescing,
		ConstantPropagation,
		LetMinimization,
		LetMerging,
		PartialRedundancyElimination,
		LetMinimization,
		LetMerging,
	)
}

// RunPasses runs optimization passes, in order, on a pipeline (and all of its subpipelines.
func RunPasses(ctx context.Context, pipeline *ast.Pipeline, memoryLimit uint64, opts ...Optimization) *ast.Pipeline {
	checkCancel := func() bool {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	}

	for _, opt := range opts {
		if checkCancel() {
			return pipeline
		}
		pipeline = opt(NormalizePipeline(pipeline), memoryLimit)
	}
	pipeline = NormalizePipeline(pipeline)

	for i := range pipeline.Stages {
		switch typedStage := pipeline.Stages[i].(type) {
		case *ast.LookupStage:
			if typedStage.Pipeline != nil {
				pipeline.Stages[i].(*ast.LookupStage).Pipeline = RunPasses(
					ctx,
					typedStage.Pipeline,
					memoryLimit,
					opts...,
				)
			}
		case *ast.FacetStage:
			for j := range typedStage.Items {
				pipeline.Stages[i].(*ast.FacetStage).Items[j].Pipeline = RunPasses(
					ctx,
					typedStage.Items[j].Pipeline,
					memoryLimit,
					opts...,
				)
			}
		case *ast.UnionWithStage:
			if typedStage.Pipeline != nil {
				pipeline.Stages[i].(*ast.UnionWithStage).Pipeline = RunPasses(
					ctx,
					typedStage.Pipeline,
					memoryLimit,
					opts...,
				)
			}
		}
	}

	return pipeline
}
