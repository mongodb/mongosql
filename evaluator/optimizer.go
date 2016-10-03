package evaluator

import (
	"os"

	"github.com/10gen/sqlproxy/log"
)

// OptimizeCommand applies optimizations to the command
// plan tree to aid in performance.
func OptimizeCommand(ctx ConnectionCtx, c command) (command, error) {
	n, err := optimize(ctx, c)
	if err != nil {
		return nil, err
	}
	return n.(command), nil
}

// OptimizePlan applies optimizations to the plan tree to
// aid in performance.
func OptimizePlan(ctx ConnectionCtx, p PlanStage) (PlanStage, error) {
	n, err := optimize(ctx, p)
	if err != nil {
		return nil, err
	}

	return n.(PlanStage), nil
}

func optimize(ctx ConnectionCtx, n node) (node, error) {

	logger := ctx.Logger(log.OptimizerComponent)

	// we use the collation connection during optimization because the nodes
	// that get evaluated during this phase are literal values.
	evalCtx := NewEvalCtx(NewExecutionCtx(ctx), ctx.Variables().CollationConnection)
	newN, err := optimizeEvaluations(evalCtx, n)
	if err != nil {
		logger.Warnf(log.DebugHigh, err.Error())
		return n, nil
	}
	n = newN

	if os.Getenv(NoPushDown) != "" {
		logger.Warnf(log.DebugHigh, "pushdown turned off")
		return n, nil
	}

	newN, err = optimizeCrossJoins(n)
	if err != nil {
		logger.Warnf(log.DebugHigh, err.Error())
		return n, nil
	}
	n = newN

	newN, err = optimizePushDown(ctx, logger, n)
	if err != nil {
		logger.Warnf(log.DebugHigh, err.Error())
		return n, nil
	}
	n = newN

	return n, nil
}

func combineExpressions(exprs []SQLExpr) SQLExpr {
	var combined SQLExpr
	if len(exprs) > 0 {
		combined = exprs[0]
		for _, expr := range exprs[1:] {
			combined = &SQLAndExpr{combined, expr}
		}
	}
	return combined
}

func splitExpression(e SQLExpr) []SQLExpr {
	andE, ok := e.(*SQLAndExpr)
	if !ok {
		return []SQLExpr{e}
	}

	left := splitExpression(andE.left)
	right := splitExpression(andE.right)
	return append(left, right...)
}

func splitExpressionIntoParts(e SQLExpr) (expressionParts, error) {
	// this splits hierarchical SQLAndExprs into a flattened list.
	exprs := splitExpression(e)
	result := []expressionPart{}
	for _, expr := range exprs {
		tableNames, err := referencedTables(expr)
		if err != nil {
			return nil, err
		}
		result = append(result, expressionPart{expr, tableNames})
	}
	return result, nil
}

type expressionParts []expressionPart

func (parts expressionParts) combine() SQLExpr {
	var combined SQLExpr
	if len(parts) > 0 {
		combined = parts[0].expr
		for _, part := range parts[1:] {
			combined = &SQLAndExpr{combined, part.expr}
		}
	}
	return combined
}

type expressionPart struct {
	expr       SQLExpr
	tableNames []string
}

func referencedTables(e SQLExpr) ([]string, error) {
	finder := &sqlExprReferencedTableCollector{}
	_, err := finder.visit(e)
	if err != nil {
		return nil, err
	}

	return finder.tableNames, nil
}

type sqlExprReferencedTableCollector struct {
	tableNames []string
}

func (v *sqlExprReferencedTableCollector) visit(n node) (node, error) {
	switch typedN := n.(type) {
	case SQLColumnExpr:
		v.tableNames = append(v.tableNames, typedN.tableName)
	}
	return walk(v, n)
}
