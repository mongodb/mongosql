package evaluator

import (
	"gopkg.in/mgo.v2/bson"
)

// Filter ensures that only rows matching a given criteria are
// returned.
type Filter struct {
	// err holds any error that may have occurred during processing
	err error

	// source holds the source for this select statement
	source Operator

	// matcher is used to filter results gotten from the source operator
	matcher SQLExpr

	// ctx is the current execution context
	ctx *ExecutionCtx

	// hasSubquery is true if this operator source contains a subquery
	hasSubquery bool
}

func NewFilter(source Operator, matcher SQLExpr, hasSubquery bool) *Filter {
	return &Filter{
		source:      source,
		matcher:     matcher,
		hasSubquery: hasSubquery,
	}
}

func (ft *Filter) Open(ctx *ExecutionCtx) error {
	ft.ctx = ctx
	return ft.source.Open(ctx)
}

func (ft *Filter) Next(row *Row) bool {

	var hasNext bool

	for {

		hasNext = ft.source.Next(row)

		if !hasNext {
			break
		}

		evalCtx := &EvalCtx{Rows{*row}, ft.ctx}

		// add parent row(s) to this subquery's evaluation context
		if len(ft.ctx.SrcRows) != 0 {

			bound := len(ft.ctx.SrcRows) - 1

			for _, r := range ft.ctx.SrcRows[:bound] {
				evalCtx.Rows = append(evalCtx.Rows, *r)
			}

			// avoid duplication since subquery row is most recently
			// appended and "*row"
			if !ft.hasSubquery {
				evalCtx.Rows = append(evalCtx.Rows, *ft.ctx.SrcRows[bound])
			}

		}

		if ft.matcher != nil {
			m, err := Matches(ft.matcher, evalCtx)
			if err != nil {
				ft.err = err
				return false
			}
			if m {
				break
			}
		} else {
			break
		}

	}

	return hasNext
}

func (ft *Filter) OpFields() (columns []*Column) {
	return ft.source.OpFields()
}

func (ft *Filter) Close() error {
	return ft.source.Close()
}

func (ft *Filter) Err() error {
	if err := ft.source.Err(); err != nil {
		return err
	}
	return ft.err
}

///////////////
//Optimization
///////////////

func (v *optimizer) visitFilter(filter *Filter) (Operator, error) {

	ms, ok := canPushDown(filter.source)
	if !ok {
		return filter, nil
	}

	optimizedExpr, err := OptimizeSQLExpr(filter.matcher)
	if err != nil {
		return nil, err
	}

	pipeline := ms.pipeline
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
		matchBody, localMatcher = TranslatePredicate(optimizedExpr, ms.mappingRegistry.lookupFieldName, ms.mappingRegistry.lookupFieldType)

		if matchBody == nil {
			// no pieces of the matcher are able to be pushed down,
			// so there is no change in the operator tree.
			return filter, nil
		}

		pipeline = append(ms.pipeline, bson.D{{"$match", matchBody}})
	}

	// if we end up here, it's because we have messed with the pipeline
	// in the current table scan operator, so we need to reconstruct the
	// operator nodes.
	ms = ms.WithPipeline(pipeline)

	if localMatcher != nil {
		// we ended up here because we have a predicate
		// that can be partially pushed down, so we construct
		// a new filter with only the part remaining that
		// cannot be pushed down.
		return NewFilter(ms, localMatcher, filter.hasSubquery), nil
	}

	// everything was able to be pushed down, so the filter
	// is removed from the plan.
	return ms, nil
}
