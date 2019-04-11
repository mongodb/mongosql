package analyzer

import (
	"github.com/10gen/mongoast/ast"
)

// IsVariableDefiner returns true if this expression defines variables.
func IsVariableDefiner(expr ast.Expr) bool {
	switch typedExpr := expr.(type) {
	case *ast.Let:
		return true
	case *ast.Function:
		return IsVariableDefiningFunction(typedExpr)
	}
	return false
}

// IsVariableDefiningFunction returns true if this Function can define variables.
func IsVariableDefiningFunction(fun *ast.Function) bool {
	switch fun.Name {
	case "$filter", "$map", "$reduce":
		return true
	}
	return false
}
