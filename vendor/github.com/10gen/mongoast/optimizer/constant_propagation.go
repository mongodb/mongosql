package optimizer

import (
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/eval"
	"github.com/10gen/mongoast/internal/bsonutil"
)

func ConstantPropagation(pipeline *ast.Pipeline, memoryLimit uint64) *ast.Pipeline {
	return aggConstantPropagation(pipeline, memoryLimit).(*ast.Pipeline)
}

func aggConstantPropagation(root ast.Node, memoryLimit uint64) ast.Node {
	// We will use MinKey to signal to the evaluator that fields should
	// be treated as errors rather than missing values. Anything during
	// actual evaluation passed as the value in EvaluateExpr should
	// be a document. MinKey is the Bottom type of bson, so it
	// is the most semantically pleasing choice.
	minKey := bsonutil.MinKey()
	out, _ := ast.Visit(root, func(v ast.Visitor, n ast.Node) ast.Node {
		switch tn := n.(type) {
		case *ast.MatchStage:
			expr := matchConstantPropagation(tn.Expr, memoryLimit)
			if expr != tn.Expr {
				return ast.NewMatchStage(expr)
			}
			return n
		case ast.Expr:
			if evaled, err := eval.PartialEvaluateExpr(tn, minKey, memoryLimit); err == nil {
				return evaled
			}
		}
		return n.Walk(v)
	})
	return out
}

func matchConstantPropagation(root ast.Expr, memoryLimit uint64) ast.Expr {
	// The only place inside a $match stage where there can be expressions
	// evaluating to constants that can be folded is inside a $expr clause.
	out, _ := ast.Visit(root, func(v ast.Visitor, n ast.Node) ast.Node {
		switch tn := n.(type) {
		case *ast.AggExpr:
			subexpr := aggConstantPropagation(tn.Expr, memoryLimit).(ast.Expr)
			if subexpr != tn.Expr {
				return ast.NewAggExpr(subexpr)
			}
		}
		return n.Walk(v)
	})
	return out.(ast.Expr)
}
