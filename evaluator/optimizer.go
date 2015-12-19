package evaluator

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

func OptimizeOperator(ctx *ExecutionCtx, o Operator) (Operator, error) {
	v := &optimizer{ctx}
	return v.Visit(o)
}

type optimizer struct {
	ctx *ExecutionCtx
}

func (v *optimizer) Visit(o Operator) (Operator, error) {
	switch typedO := o.(type) {
	case *Filter:
		return optimizeFilter(v.ctx, typedO)
	}

	return walkOperatorTree(v, o)
}

func optimizeFilter(ctx *ExecutionCtx, filter *Filter) (Operator, error) {

	// we can only optimize a filter if its source is a SourceAppend
	// whose source is a TableScan
	sa, ok := filter.source.(*SourceAppend)
	if !ok {
		return filter, nil
	}
	ts, ok := sa.source.(*TableScan)
	if !ok {
		return filter, nil
	}

	optimizedExpr, err := OptimizeSQLExpr(filter.matcher)
	if err != nil {
		return nil, err
	}

	pipeline := ts.pipeline

	if value, ok := optimizedExpr.(SQLValue); ok {
		// our optimized expression has left us with just a value,
		// we can see if it matches right now. If so, we eliminate
		// the filter from the tree. Otherwise, we return an
		// operator that yields no rows.
		matches, err := Matches(value, nil)
		if err != nil {
			return nil, err
		}
		if !matches {
			return &Empty{}, nil
		}

		// otherwise, the filter simply gets removed from the tree

	} else {
		// TODO: we should first check if the optimizedExpr is a *SQLAndExpr.
		// if so, we can translate each condition separately, push down
		// the ones that are translatable, and evaluate locally the ones
		// that are not.

		dbName := ts.dbName
		if dbName == "" {
			dbName = ctx.Db
		}

		db, ok := ctx.Schema.Databases[dbName]
		if !ok {
			return nil, fmt.Errorf("Database %q could not be found in the schema.", dbName)
		}

		matchBody, ok := TranslatePredicate(optimizedExpr, db)
		if !ok {
			// we were unable to translate the expression into a
			// MongoDB query, so we'll have to evaluate it locally.
			return filter, nil
		}

		pipeline = append(ts.pipeline, bson.M{"$match": matchBody})
	}

	ts = &TableScan{
		pipeline:    pipeline,
		dbName:      ts.dbName,
		tableName:   ts.tableName,
		matcher:     ts.matcher,
		iter:        ts.iter,
		database:    ts.database,
		session:     ts.session,
		tableSchema: ts.tableSchema,
		ctx:         ts.ctx,
		err:         ts.err,
	}

	sa = &SourceAppend{
		source:      ts,
		ctx:         sa.ctx,
		hasSubquery: sa.hasSubquery,
	}

	// Remove the filter from the tree.
	// Alternatively, we could not remove the filter from the tree
	// and simply mark it as fully pushed down such that a Limit, for
	// instance, knows it is able to skip this Filter and push down the
	// limit as well.
	return sa, nil
}
