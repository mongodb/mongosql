package desugarer

import "github.com/10gen/mongoast/ast"

// desugarUnsupportedOperators desugars any new operators ($sqlBetween,
// $divide, $sqlSlice, $like, $assert, $nullIf, $coalesce) into appropriate,
// equivalent aggregation operators.
func desugarUnsupportedOperators(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {
	return pipeline
}
