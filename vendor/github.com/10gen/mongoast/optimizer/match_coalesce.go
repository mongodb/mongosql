package optimizer

import (
	"github.com/10gen/mongoast/ast"
)

// MatchCoalescing coalesces multiple adjacent matches into one.
func MatchCoalescing(pipeline *ast.Pipeline) *ast.Pipeline {
	allStages := make([]ast.Stage, 0, len(pipeline.Stages))

	var currentMatchStage *ast.MatchStage
	for _, stage := range pipeline.Stages {
		if matchStage, ok := stage.(*ast.MatchStage); ok {
			currentMatchStage = collapse(currentMatchStage, matchStage)
		} else {
			if currentMatchStage != nil {
				allStages = append(allStages, currentMatchStage)
			}
			allStages = append(allStages, stage)
			currentMatchStage = nil
		}
	}

	if currentMatchStage != nil {
		allStages = append(allStages, currentMatchStage)
	}

	pipeline.Stages = allStages
	return pipeline
}

func collapse(a, b *ast.MatchStage) *ast.MatchStage {
	if a == nil {
		return b
	}

	if b == nil {
		return a
	}

	var expr ast.Expr

	// Check if a and b are both $expr's.
	aAggExpr, aok := a.Expr.(*ast.AggExpr)
	bAggExpr, bok := b.Expr.(*ast.AggExpr)
	if aok && bok {
		expr = ast.NewAggExpr(ast.NewBinary(ast.And, aAggExpr.Expr, bAggExpr.Expr))
	} else {
		expr = ast.NewBinary(ast.And, a.Expr, b.Expr)
	}

	return ast.NewMatchStage(expr)
}
