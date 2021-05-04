package desugarer

import "github.com/10gen/mongoast/ast"

// desugarSubqueryExprs desugars any non-nested $subquery expression (including
// $subqueryExists and $subqueryComparison) into a sequence of equivalent,
// existing stages and expression. A "nested" subquery expression is one that
// exists in a sub-pipeline. This desugarer should skip such expressions since we
// do not want to pull those out to the top-level.
func desugarSubqueryExprs(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {
	return pipeline
}
