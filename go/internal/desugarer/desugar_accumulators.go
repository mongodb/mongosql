package desugarer

import (
	"strings"

	"github.com/10gen/mongoast/ast"
)

var sqlAccumulatorToMQL = map[string]string{
	"$sqlAvg":          "$avg",
	"$sqlLast":         "$last",
	"$sqlMergeObjects": "$mergeObjects",
	"$sqlStdDevPop":    "$stdDevPop",
	"$sqlStdDevSamp":   "$stdDevSamp",
	"$sqlSum":          "$sum",
}

// desugarAccumulators desugars any $sql-prefixed accumulators in $group stages
// into appropriate, equivalent expressions and/or stages.
func desugarAccumulators(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {

	// $sql-prefixed accumulators may be distinct or non-distinct accumulations.
	// Non-distinct accumulators are replaced in the $group stage by appropriate
	// desugared MQL accumulator functions.
	// Distinct accumulators are replaced in the $group stage by a $addToSet
	// accumulator, and the $group stage is followed by a $project that performs
	// the actual target accumulation operation.
	// Since a pipeline may or may not contain distinct accumulations, the most
	// we know is that the new pipeline will be at least as long as the input
	// pipeline, so we start with that capacity.
	newStages := make([]ast.Stage, 0, len(pipeline.Stages))

	for _, stage := range pipeline.Stages {
		switch s := stage.(type) {
		case *ast.GroupStage:

			// All group items are collected as project items. needsProject is
			// set to true if we encounter at least one distinct accumulator.
			projectItems := make([]ast.ProjectItem, len(s.Items))
			needsProject := false

			for i, item := range s.Items {
				accFunc := item.Expr.(*ast.Function)

				// There is nothing to desugar for non-$sql-prefixed accumulators.
				if !strings.HasPrefix(accFunc.Name, "$sql") {
					projectItems[i] = ast.NewIncludeProjectItem(ast.NewFieldRef(item.Name, nil))
					continue
				}

				accRef := ast.NewFieldRef(item.Name, nil)
				accArg := accFunc.Arg.(*ast.Document).FieldsMap()

				if accArg["distinct"].(*ast.Constant).Value.Boolean() {
					groupExpr, projectExpr := desugarDistinctAccumulator(
						accFunc.Name, accArg["var"], accRef)

					item.Expr = groupExpr
					projectItems[i] = ast.NewAssignProjectItem(item.Name, projectExpr)
					needsProject = true
				} else {
					item.Expr = desugarNonDistinctAccumulator(accFunc.Name, accArg["var"])
					projectItems[i] = ast.NewIncludeProjectItem(accRef)
				}
			}

			newStages = append(newStages, s)
			if needsProject {
				newStages = append(newStages, ast.NewProjectStage(projectItems...))
			}

		default:
			newStages = append(newStages, s)
		}
	}

	return ast.NewPipeline(newStages...)
}

func desugarDistinctAccumulator(accumulator string, expr ast.Expr, accRef *ast.FieldRef) (*ast.Function, ast.Expr) {
	var projectExpr ast.Expr
	if accumulator == "$sqlCount" {
		projectExpr = ast.NewReduce(
			accRef,
			zeroLiteral,
			wrapInOp("$add",
				ast.NewVariableRef("value"),
				sqlCountCond(ast.NewVariableRef("this")),
			),
		)
	} else {
		projectExpr = ast.NewFunction(sqlAccumulatorToMQL[accumulator], accRef)
	}

	groupExpr := ast.NewFunction("$addToSet", expr)
	return groupExpr, projectExpr
}

func desugarNonDistinctAccumulator(accumulator string, expr ast.Expr) *ast.Function {
	if accumulator == "$sqlCount" {
		return ast.NewFunction("$sum", sqlCountCond(expr))
	}

	return ast.NewFunction(sqlAccumulatorToMQL[accumulator], expr)
}

func sqlCountCond(expr ast.Expr) *ast.Conditional {
	return ast.NewConditional(
		wrapInOp("$in",
			ast.NewFunction("$type", expr),
			ast.NewArray(
				missingStringLiteral,
				nullStringLiteral,
			),
		),
		zeroLiteral,
		oneLiteral,
	)
}
