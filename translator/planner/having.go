package planner

import (
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/translator/evaluator"
	"github.com/erh/mongo-sql-temp/translator/types"
)

type Having struct {
	// source is the operator that provides the data to filter
	source Operator
	// err holds any error encountered during processing
	err error
	// matcher is used to evaluate each row in determining whether
	// it passes the 'HAVING' filter
	matcher evaluator.Matcher
	// expr is the boolean expression to match
	expr sqlparser.BoolExpr
}

func (hv *Having) Open(ctx *ExecutionCtx) (err error) {
	hv.matcher, err = evaluator.BuildMatcher(hv.expr)
	if err != nil {
		return err
	}

	return hv.source.Open(ctx)
}

func (hv *Having) Next(row *types.Row) bool {
	r := &types.Row{}

	var hasNext bool

	for {
		hasNext = hv.source.Next(r)

		if !hasNext {
			if err := hv.source.Err(); err != nil {
				hv.err = err
			} else if err := hv.source.Close(); err != nil {
				hv.err = err
			}
			return false
		}

		evalCtx := &evaluator.EvalCtx{[]types.Row{*r}}

		if hv.matcher.Matches(evalCtx) {
			row.Data = r.Data
			break
		}
	}

	return hasNext
}

func (hv *Having) Close() error {
	return hv.source.Close()
}

func (hv *Having) Err() error {
	return hv.err
}

func (hv *Having) OpFields() (columns []*Column) {
	return hv.source.OpFields()
}
