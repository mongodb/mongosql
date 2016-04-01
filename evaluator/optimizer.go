package evaluator

import (
	"os"

	"github.com/mongodb/mongo-tools/common/log"
)

// OptimizeOperator applies optimizations to the operator tree to
// aid in performance.
func OptimizeOperator(ctx *ExecutionCtx, o Operator) (Operator, error) {
	if os.Getenv(NoOptimize) != "" {
		return o, nil
	}

	newO, err := optimizeOperatorSQLExprs(o)
	if err != nil {
		return o, nil
	}
	o = newO

	log.Logf(log.DebugHigh, "SQL Expr Optimization query plan: \n%v\n", PrettyPrintPlan(o))

	newO, err = optimizeCrossJoins(o)
	if err != nil {
		return o, nil
	}
	o = newO

	log.Logf(log.DebugHigh, "Cross Join Optimization query plan: \n%v\n", PrettyPrintPlan(o))

	newO, err = optimizePushDown(ctx, o)
	if err != nil {
		return o, nil
	}
	o = newO

	log.Logf(log.DebugHigh, "Optimized query plan: \n%v\n", PrettyPrintPlan(o))
	return o, nil
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
	_, err := finder.Visit(e)
	if err != nil {
		return nil, err
	}

	return finder.tableNames, nil
}

type sqlExprReferencedTableCollector struct {
	tableNames []string
}

func (v *sqlExprReferencedTableCollector) Visit(e SQLExpr) (SQLExpr, error) {
	switch typedE := e.(type) {
	case SQLColumnExpr:
		v.tableNames = append(v.tableNames, typedE.tableName)
	}
	return walk(v, e)
}

func containsString(strs []string, str string) bool {
	for _, n := range strs {
		if n == str {
			return true
		}
	}

	return false
}
