package normalizer

import (
	"github.com/10gen/mongoast/ast"
)

// Normalize modifies an AST so that all field references occur on the
// left-hand side of binary operators.
func Normalize(node ast.Node) ast.Node {
	return normalizeVisitor{}.Visit(node)
}

type normalizeVisitor struct{}

func (v normalizeVisitor) Visit(node ast.Node) ast.Node {
	expr, ok := node.(*ast.Binary)
	if ok {
		switch expr.Op {
		case ast.LessThan, ast.LessThanOrEquals, ast.GreaterThan, ast.GreaterThanOrEquals, ast.Equals, ast.NotEquals:
			_, leftFieldRef := expr.Left.(*ast.FieldRef)
			_, rightFieldRef := expr.Right.(*ast.FieldRef)
			if rightFieldRef && !leftFieldRef {
				return ast.NewBinary(expr.Op.Flip(), expr.Right, expr.Left)
			}
		}
	}
	return node.Walk(v)
}
