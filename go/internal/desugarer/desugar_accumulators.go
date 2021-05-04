package desugarer

import "github.com/10gen/mongoast/ast"

// desugarAccumulators desugars any $sql-prefixed accumulators in $group stages
// into appropriate, equivalent expressions and/or stages.
func desugarAccumulators(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {
	return pipeline
}
