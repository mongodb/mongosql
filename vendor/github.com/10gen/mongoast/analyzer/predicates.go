package analyzer

import (
	"github.com/10gen/mongoast/ast"
)

// SplitPredicate splits the expression into expressions that can be evaluated individually.
func SplitPredicate(expr ast.Expr) []ast.Expr {
	var result []ast.Expr
	switch exprT := expr.(type) {
	case *ast.Binary:
		if exprT.Op == ast.And {
			result = append(result, SplitPredicate(exprT.Left)...)
			result = append(result, SplitPredicate(exprT.Right)...)
		} else {
			result = append(result, exprT)
		}
	default:
		result = append(result, exprT)
	}

	return result
}
