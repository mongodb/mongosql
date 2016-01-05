package evaluator

import (
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
	var localMatcher SQLExpr

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
		var matchBody bson.M
		matchBody, localMatcher = TranslatePredicate(optimizedExpr, ctx.Schema.Databases[ctx.Db])

		if matchBody == nil {
			// no pieces of the matcher are able to be pushed down,
			// so there is no change in the operator tree.
			return filter, nil
		}

		pipeline = append(ts.pipeline, bson.D{{"$match", matchBody}})
	}

	// if we end up here, it's because we have messed with the pipeline
	// in the current table scan operator, so we need to reconstruct the
	// operator nodes.
	ts = &TableScan{
		pipeline:  pipeline,
		dbName:    ts.dbName,
		tableName: ts.tableName,
		matcher:   ts.matcher,
	}

	sa = &SourceAppend{
		source:      ts,
		hasSubquery: sa.hasSubquery,
	}

	if localMatcher != nil {
		// we ended up here because we have a predicate
		// that can be partially pushed down, so we construct
		// a new filter with only the part remaining that
		// cannot be pushed down.
		filter = &Filter{
			source:      sa,
			matcher:     localMatcher,
			hasSubquery: filter.hasSubquery,
		}

		return filter, nil
	}

	// everything was able to be pushed down, so the filter
	// is removed from the plan.
	return sa, nil
}
