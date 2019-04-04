package optimizer

import (
	"context"

	"github.com/10gen/mongoast/ast"
)

// Optimization is the type of a pipeline optimization.
type Optimization = func(*ast.Pipeline) *ast.Pipeline

// Optimize handles optimizing a pipeline.
func Optimize(ctx context.Context, pipeline *ast.Pipeline) *ast.Pipeline {
	return RunPasses(ctx, pipeline,
		DeadCodeElimination,
		Reorder,
		DeadCodeElimination,
	)
}

// RunPasses runs optimization passes, in order, on a pipeline (and all of its subpipelines.
func RunPasses(ctx context.Context, pipeline *ast.Pipeline, opts ...Optimization) *ast.Pipeline {
	checkCancel := func() bool {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return true
			default:
			}
		}
		return false
	}

	for _, opt := range opts {
		if checkCancel() {
			return pipeline
		}
		pipeline = opt(pipeline)
	}

	for i := range pipeline.Stages {
		switch typedStage := pipeline.Stages[i].(type) {
		case *ast.LookupStage:
			if typedStage.Pipeline != nil {
				pipeline.Stages[i].(*ast.LookupStage).Pipeline = RunPasses(ctx, typedStage.Pipeline, opts...)
			}
		case *ast.FacetStage:
			for j := range typedStage.Items {
				pipeline.Stages[i].(*ast.FacetStage).Items[j].Pipeline = RunPasses(ctx, typedStage.Items[j].Pipeline, opts...)
			}
		}
	}

	return pipeline
}
