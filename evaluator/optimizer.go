package evaluator

import (
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/variable"
)

// OptimizeCommand applies optimizations to the command
// plan tree to aid in performance.
func OptimizeCommand(ctx ConnectionCtx, c command) command {
	n := optimize(ctx, c, false)
	return n.(command)
}

// OptimizePlan applies optimizations to the plan tree to
// aid in performance.
func OptimizePlan(ctx ConnectionCtx, p PlanStage) PlanStage {
	n := optimize(ctx, p, false)
	return n.(PlanStage)
}

type optimizerStage struct {
	name string
	f    func(node, *EvalCtx, *log.Logger) (node, error)
}

var optimizerStages = []optimizerStage{
	{"evaluations", optimizeEvaluations},
	{"cross joins", optimizeCrossJoins},
	{"inner join", optimizeInnerJoins},
	{"filtering", optimizeFiltering},
	{"pushdown", optimizePushDown},
}

func optimize(ctx ConnectionCtx, n node, isSubquery bool) node {
	logger := ctx.Logger(log.OptimizerComponent)

	if !isSubquery {
		logger.Infof(log.Dev, "running optimization stage 'subqueries'")
		newN, err := optimizeSubqueries(ctx, logger, n, true)
		if err != nil {
			logger.Warnf(log.Admin, "error running optimization stage 'subqueries': %v", err)
		} else if newN != n {
			n = newN
			logger.Debugf(log.Dev, "optimized plan after 'subqueries': \n%v", prettyPrintNode(n))
		}
	}

	evalCtx := NewEvalCtx(NewExecutionCtx(ctx), ctx.Variables().GetCollation(variable.CollationConnection))

	for _, stage := range optimizerStages {
		logger.Infof(log.Dev, "running optimization stage '%s'", stage.name)
		newN, err := stage.f(n, evalCtx, logger)
		if err != nil {
			logger.Warnf(log.Admin, "error running optimization stage '%s': %v", stage.name, err)
			// don't exit here. Just because we couldn't apply one optimization doesn't mean
			// others aren't valid
		} else if newN != n {
			n = newN
			logger.Debugf(log.Dev, "optimized plan after '%s': \n%v", stage.name, prettyPrintNode(n))
		}
	}

	return n
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

func sharesRootTable(logger *log.Logger, local, foreign *MongoSourceStage) bool {
	baseCollectionName := local.collectionNames[0]

	logger.Debugf(log.Dev, "attempting to use self-join optimization for tables %v and %v",
		local.aliasNames, foreign.aliasNames)

	for _, collectionName := range append(local.collectionNames[1:],
		foreign.collectionNames...) {
		if collectionName != baseCollectionName {
			logger.Debugf(log.Dev, "cannot use self-join optimization, "+
				"pipeline has different root tables: %v and %v",
				baseCollectionName, collectionName)
			return false
		}
	}

	return true
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
