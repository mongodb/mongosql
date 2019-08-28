package analyzer

import (
	"github.com/10gen/mongoast/ast"
)

// DeMorganize applies DeMorgan's laws to push any $not nodes down as far as
// possible.
func DeMorganize(expr ast.Expr) ast.Expr {
	if un, ok := expr.(*ast.Unary); ok && un.Op == ast.Not {
		expr = deMorganizeNegation(un)
	}

	if bin, ok := expr.(*ast.Binary); ok && (bin.Op == ast.And || bin.Op == ast.Or) {
		left := DeMorganize(bin.Left)
		right := DeMorganize(bin.Right)
		if left != bin.Left || right != bin.Right {
			expr = ast.NewBinary(bin.Op, left, right)
		}
	}

	return expr
}

func deMorganizeNegation(un *ast.Unary) ast.Expr {
	bin, ok := un.Expr.(*ast.Binary)
	if !ok {
		return un
	}

	var op ast.BinaryOp
	switch bin.Op {
	case ast.And:
		op = ast.Or
	case ast.Or:
		op = ast.And
	default:
		return un
	}
	return ast.NewBinary(
		op,
		ast.NewUnary(ast.Not, bin.Left),
		ast.NewUnary(ast.Not, bin.Right),
	)
}
