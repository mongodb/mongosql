package optimizer

import (
	"context"

	"github.com/10gen/mongoast/ast"
)

// Optimize handles optimizing a pipeline.
func Optimize(ctx context.Context, pipeline *ast.Pipeline) *ast.Pipeline {
	pipeline = Reorder(pipeline)

	return pipeline
}
