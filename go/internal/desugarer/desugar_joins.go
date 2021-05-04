package desugarer

import "github.com/10gen/mongoast/ast"

// desugarJoins desugars any $join stages in the pipeline into a sequence
// of equivalent, existing MQL stages.
func desugarJoins(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {
	return pipeline
}
