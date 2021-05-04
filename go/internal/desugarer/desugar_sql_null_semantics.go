package desugarer

import "github.com/10gen/mongoast/ast"

// desugarSQLNullSemantics desugars any $sql-prefixed operators (other than
// $sqlSlice and $sqlBetween) into their corresponding aggregation operator
// wrapped in operations that null-check the arguments.
func desugarSQLNullSemantics(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {
	return pipeline
}
