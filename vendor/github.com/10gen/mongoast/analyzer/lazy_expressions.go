package analyzer

import (
	"github.com/10gen/mongoast/ast"
)

// HasLazyArgumentSemantics returns true if it is not safe to eagerly
// evaluate this expression's arguments.
func HasLazyArgumentSemantics(e ast.Expr) bool {
	switch typedExpr := e.(type) {
	case *ast.Binary:
		switch typedExpr.Op {
		case ast.And, ast.Or, ast.Nor:
			return true
		}
	case *ast.Conditional:
		return true
	case *ast.Function:
		switch typedExpr.Name {
		case "$allElementsTrue", // stops on first False
			"$and",            // stops on first False
			"$anyElementTrue", // stops on first True
			"$cond",           // conditional
			"$convert",        // onError and onNull are only evaluated lazily
			"$dateFromString", // onError and onNull are only evaluated lazily
			"$ifNull",         // conditional
			"$or",             // stops on first True
			"$switch",         // conditional
			"$zip":            // defaults evaluated on demand
			return true
		}
	}
	return false
}
