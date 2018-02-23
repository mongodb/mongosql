package evaluator

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/log"
)

// constantColumnReplacer holds the execution context, which has the data
// used to replace the column expressions.
type constantColumnReplacer struct {
	ctx *ExecutionCtx
}

// replaceColumnWithConstant kicks off the replacement of column expressions.
func replaceColumnWithConstant(n Node, ctx *ExecutionCtx) (Node, error) {
	v := &constantColumnReplacer{ctx}
	n, err := v.visit(n)
	return n, err
}

func (v *constantColumnReplacer) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case SQLColumnExpr:
		for _, row := range v.ctx.SrcRows {
			if val, ok := row.GetField(typedN.selectID, typedN.databaseName, typedN.tableName, typedN.columnName); ok {
				return val, nil
			}
		}
	}
	return walk(v, n)
}

// joinOnExpression holds the SQLColumnExpr value
// used within a join stage matcher.
type joinOnExpression struct {
	exprCollector *sqlColExprCollector
}

func (v *joinOnExpression) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case *JoinStage:
		v.exprCollector.visit(typedN.matcher)
		return typedN, nil
	default:
		return walk(v, n)
	}
}

// sqlColExprCounter is a used to hold and count all
// SQLColumnExpr values found during a Node visit.
type sqlColExprCounter struct {
	counts map[string]int
	exprs  []SQLColumnExpr
}

// newSQLColumnExprCounter returns a new sqlColExprCounter.
func newSQLColumnExprCounter() *sqlColExprCounter {
	return &sqlColExprCounter{
		counts: make(map[string]int),
	}
}

func (c *sqlColExprCounter) add(e SQLColumnExpr) {
	s := e.String()
	if _, ok := c.counts[s]; ok {
		c.counts[s]++
	} else {
		c.counts[s] = 1
		c.exprs = append(c.exprs, e)
	}
}

func (c *sqlColExprCounter) copyExprs() []SQLColumnExpr {
	exprs := make([]SQLColumnExpr, len(c.exprs))
	copy(exprs, c.exprs)
	return exprs
}

func (c *sqlColExprCounter) remove(e SQLColumnExpr) {
	s := e.String()
	for i, expr := range c.exprs {
		if strings.EqualFold(s, expr.String()) {
			c.counts[s]--
			if c.counts[s] == 0 {
				delete(c.counts, s)
				c.exprs = append(c.exprs[:i], c.exprs[i+1:]...)
			}
			return
		}
	}
}

// sqlColExprCollector is used to track which SQLColumnExpr values
// have been visited and the select context in which they were found.
type sqlColExprCollector struct {
	selectIDs         []int
	referencedColumns *sqlColExprCounter
	removeMode        bool
}

// newSQLColExprCollector returns a new sqlColExprCollector.
func newSQLColExprCollector(selectIDs []int) *sqlColExprCollector {
	return &sqlColExprCollector{
		selectIDs:         selectIDs,
		referencedColumns: newSQLColumnExprCounter(),
	}
}

// Remove visits and removes the SQLColumnExpr
// values within e from the expression collector.
func (c *sqlColExprCollector) Remove(e SQLExpr) {
	c.removeMode = true
	c.Add(e)
	c.removeMode = false
}

// Remove visits and removes the SQLColumnExpr values
// within each element of the slice, e, from the
// expression collector.
func (c *sqlColExprCollector) RemoveAll(e []SQLExpr) {
	c.removeMode = true
	c.AddAll(e)
	c.removeMode = false
}

// Remove visits and adds the SQLColumnExpr values
// within each element of the slice, e, to the
// expression collector.
func (c *sqlColExprCollector) AddAll(exprs []SQLExpr) {
	for _, e := range exprs {
		c.Add(e)
	}
}

// Remove visits and adds the SQLColumnExpr values
// within e to the expression collector.
func (c *sqlColExprCollector) Add(e SQLExpr) {
	c.visit(e)
}

func (c *sqlColExprCollector) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case SQLColumnExpr:
		if containsInt(c.selectIDs, typedN.selectID) {
			if c.removeMode {
				c.referencedColumns.remove(typedN)
			} else {
				c.referencedColumns.add(typedN)
			}
		}
		return typedN, nil
	case *SQLSubqueryExpr:
		if typedN.correlated {
			return walk(c, n)
		}
		return n, nil
	default:
		return walk(c, n)
	}
}

// getJoinOnExpressions gets all column expressions referenced in
// a join 'on' clause in the given plan stage.
func (c *sqlColExprCollector) getJoinOnExpressions(ps PlanStage) {
	v := &joinOnExpression{c}
	v.visit(ps)
}

type mongoSourceReplacer struct {
	cacheMap map[string]*CacheStage
	ctx      *EvalCtx
}

// replaceMongoSourceStages finds MongoSource stages in the query plan, executes them, and replaces them with CacheStages.
func replaceMongoSourceStages(e SQLExpr, ctx *EvalCtx) (SQLExpr, error) {
	logger := ctx.Logger(log.OptimizerComponent)

	r := &mongoSourceReplacer{cacheMap: make(map[string]*CacheStage), ctx: ctx}

	logger.Infof(log.Dev, "caching MongoSource stages for benchmarking")

	expr, err := r.visit(e)
	if err != nil {
		return nil, err
	}

	sqlExpr, ok := expr.(SQLExpr)
	if !ok {
		return nil, fmt.Errorf("replaced plan was not a SQLExpr")
	}
	if sqlExpr != e {
		logger.Infof(log.Dev, "plan after cache replacement:\n%v", sqlExpr)
	}
	return sqlExpr, nil
}

func (msr *mongoSourceReplacer) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case *MongoSourceStage:

		key := fmt.Sprintf("%s.%s", typedN.dbName, typedN.tableNames)

		// If a MongoSourceStage is in the cache, reuse it.
		if cache, ok := msr.cacheMap[key]; ok {
			return cache.clone(), nil
		}

		newCache, err := cachePlanStage(typedN, msr.ctx)
		if err != nil {
			return nil, err
		}
		msr.cacheMap[key] = newCache
		return newCache, nil
	}
	return walk(msr, n)
}
