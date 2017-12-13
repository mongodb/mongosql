package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/log"
)

// constantColumnReplacer holds the execution context, which has the data
// used to replace the column expressions.
type constantColumnReplacer struct {
	ctx *ExecutionCtx
}

// replaceColumnWithConstant kicks off the replacement of column expressions.
func replaceColumnWithConstant(n node, ctx *ExecutionCtx) (node, error) {
	v := &constantColumnReplacer{ctx}
	n, err := v.visit(n)
	return n, err
}

func (v *constantColumnReplacer) visit(n node) (node, error) {
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

func (msr *mongoSourceReplacer) visit(n node) (node, error) {
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
