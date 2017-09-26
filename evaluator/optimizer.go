package evaluator

import (
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
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

	evalCtx := NewEvalCtx(NewExecutionCtx(ctx), ctx.Variables().CollationConnection)

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

func canMergeTables(logger *log.Logger, local, foreign *MongoSourceStage, matcher SQLExpr) bool {
	return sharesRootTable(logger, local, foreign) &&
		meetsMergePKCriteria(logger, local, foreign, matcher)
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

func meetsMergePKCriteria(logger *log.Logger, local, foreign *MongoSourceStage, matcher SQLExpr) bool {
	// don't perform optimization on MongoDB views as
	// renames might have occured on fields.
	if local.isView() {
		logger.Debugf(log.Dev, "cannot merge join stage, local "+
			"table is MongoDB view")
		return false
	}

	if foreign.isView() {
		logger.Debugf(log.Dev, "cannot merge join stage, foreign "+
			"table is MongoDB view")
		return false
	}

	exprs := splitExpression(matcher)

	numPK := func(columns []*Column, table string) int {
		n, keys := 0, make(map[string]struct{})
		for _, c := range columns {
			if _, counted := keys[c.Name]; !counted && c.PrimaryKey &&
				c.Table == table {
				n, keys[c.Name] = n+1, struct{}{}
			}
		}
		return n
	}

	// When we attempt to merge two different arrays (tables) of
	// the same underlying MongoDB collection, we only require
	// one primary key equality match.
	//
	// For example, the join clause "test_array1 JOIN test_array1"
	// requires as many primary keys as are specified in the DRDL
	// file for "test_array1". However, if we attempted to merge
	// "test_array1 JOIN test_array2" instead, we'd only need an
	// equality on the base table's primary key field.
	//
	isSameArrayJoin := local.tableNames[0] == foreign.tableNames[0]
	numRequiredPKConjunctions := 1

	if isSameArrayJoin && len(local.tableNames) == 1 {
		numLocalPK := numPK(local.mappingRegistry.columns, local.aliasNames[0])
		numForeignPK := numPK(foreign.mappingRegistry.columns, foreign.aliasNames[0])
		numRequiredPKConjunctions = util.MinInt(numLocalPK, numForeignPK)
	}

	if numRequiredPKConjunctions == 0 {
		logger.Debugf(log.Dev, "cannot merge join stage, table "+
			"has no primary key")
		return false
	}

	numPKConjunctions := 0

	logger.Debugf(log.Dev, "join merge: examining match criteria...")

	registries := []*mappingRegistry{
		local.mappingRegistry,
		foreign.mappingRegistry,
	}

	seenPrimaryKeys := make(map[string]struct{})

	for _, expr := range exprs {
		if equalExpr, ok := expr.(*SQLEqualsExpr); ok {
			column1, _ := equalExpr.left.(SQLColumnExpr)
			column2, _ := equalExpr.right.(SQLColumnExpr)

			invalidLeftColumn := !containsString(local.aliasNames,
				column1.tableName) &&
				!containsString(foreign.aliasNames, column1.tableName)
			invalidRightColumn := !containsString(local.aliasNames,
				column2.tableName) &&
				!containsString(foreign.aliasNames, column2.tableName)

			if invalidLeftColumn || invalidRightColumn {
				logger.Debugf(log.Dev, "join merge: found unexpected "+
					"table references, moving on...")
				continue
			}

			if column1.selectID != column2.selectID {
				logger.Debugf(log.Dev, "join merge: found unmatched "+
					"select identifiers (%v and %v), moving on...",
					column1.selectID, column2.selectID)
				continue
			}

			columnOneName, c1RegistryIdx, ok := lookupSQLColumn(
				column1.tableName, column1.columnName, registries)
			if !ok {
				panic("Unable to find field mapping for merge column1. " +
					"This should never happen.")
			}

			columnTwoName, c2RegistryIdx, ok := lookupSQLColumn(
				column2.tableName, column2.columnName, registries)
			if !ok {
				panic("Unable to find field mapping for merge column2. " +
					"This should never happen.")
			}

			c1IsPK := registries[c1RegistryIdx].isPrimaryKey(column1.columnName)
			c2IsPK := registries[c2RegistryIdx].isPrimaryKey(column2.columnName)

			if !c1IsPK || !c2IsPK {
				logger.Debugf(log.Dev, "join merge: criteria contains "+
					"non-primary key (%v and %v), moving on...",
					column1.String(), column2.String())
				continue
			}

			if columnOneName != columnTwoName {
				logger.Debugf(log.Dev, "join merge: criteria contains "+
					"unmatched primary keys (%v and %v), moving on...",
					columnOneName, columnTwoName)
				continue
			}

			if _, ok := seenPrimaryKeys[columnOneName]; ok {
				logger.Debugf(log.Dev, "join merge: ignoring duplicate "+
					"primary key criteria '%v' and moving on...",
					column1.String())
				continue
			}

			seenPrimaryKeys[columnOneName] = struct{}{}

			numPKConjunctions++
		}
	}

	if numPKConjunctions < numRequiredPKConjunctions {
		logger.Debugf(log.Dev, "join merge: criteria conjunction "+
			"contains %v unique primary key equality %v (need %v)",
			numPKConjunctions, util.Pluralize(numPKConjunctions, "pair",
				"pairs"), numRequiredPKConjunctions)
		return false
	}

	return true
}

func sharesRootTable(logger *log.Logger, local, foreign *MongoSourceStage) bool {
	baseCollectionName := local.collectionNames[0]

	logger.Debugf(log.Dev, "attempting to merge tables %v and %v",
		local.aliasNames, foreign.aliasNames)

	for _, collectionName := range append(local.collectionNames[1:],
		foreign.collectionNames...) {
		if collectionName != baseCollectionName {
			logger.Debugf(log.Dev, "cannot merge join stage, "+
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
