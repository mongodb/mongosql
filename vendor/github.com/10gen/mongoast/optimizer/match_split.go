package optimizer

import (
	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/ast"
)

// MatchSplit splits a match into as many matches as possible.
func MatchSplit(pipeline *ast.Pipeline) *ast.Pipeline {
	allStages := make([]ast.Stage, 0, len(pipeline.Stages))

	for _, stage := range pipeline.Stages {
		if matchStage, ok := stage.(*ast.MatchStage); ok {
			allStages = append(allStages, splitMatch(matchStage.Expr)...)
		} else {
			allStages = append(allStages, stage)
		}
	}
	pipeline.Stages = allStages
	return pipeline
}

// splitMatch attempts to take the current match stage expr and split it into
// as many match stages as possible. A match stage can only be split on its
// conjunctions.
func splitMatch(expr ast.Expr) []ast.Stage {
	separateMatchStages := []ast.Stage{}
	// If there are unary matches in the three that yield conjunctions from
	// DeMorganize(), then we can further split split the match on them.
	expr = analyzer.DeMorganize(expr)

	var split = func(v ast.Visitor, n ast.Node) ast.Node {
		switch typedN := n.(type) {
		case *ast.Binary:
			if typedN.Op == ast.And {
				return n.Walk(v)
			}
			separateMatchStages = append(separateMatchStages, ast.NewMatchStage(typedN))
		case *ast.Unary:
			separateMatchStages = append(separateMatchStages, ast.NewMatchStage(typedN))
		case *ast.AggExpr:
			// It is possible that we can split the match if this $expr has multiple fields.
			separateMatchStages = append(separateMatchStages, splitAggExpr(typedN)...)
		case ast.Expr:
			separateMatchStages = append(separateMatchStages, ast.NewMatchStage(typedN))
		}
		return n
	}

	ast.Visit(expr, split)

	return separateMatchStages
}

// splitAggExpr will split a $expr into multiple $match stages. However,
// because the logic is wrapped with $expr, we need to make sure each $match
// produced is also wrapped with $expr accordingly.
func splitAggExpr(aggExpr *ast.AggExpr) []ast.Stage {
	// We will re-use the logic of splitMatch, by splitting it normally, then wrapping the results with $expr.
	aggExprMatchStages := []ast.Stage{}

	for _, stage := range splitMatch(aggExpr.Expr) {
		matchStage := stage.(*ast.MatchStage) // This is guaranteed to not panic. If it does, there is a programmer error.
		aggExprMatchStages = append(aggExprMatchStages, ast.NewMatchStage(ast.NewAggExpr(matchStage.Expr)))
	}

	return aggExprMatchStages
}
