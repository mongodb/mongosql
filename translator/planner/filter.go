package planner

import (
	"github.com/erh/mongo-sql-temp/translator/evaluator"
	"github.com/erh/mongo-sql-temp/translator/types"
)

type Filter struct {
	source  Operator
	matcher evaluator.Matcher
}

func (f *Filter) Open(ctx *ExecutionCtx) error {
	return f.source.Open(ctx)
}

func (f *Filter) OpFields() []*Column {
	return f.source.OpFields()
}

func (f *Filter) Close() error {
	return f.source.Close()
}

func (f *Filter) Next(row *types.Row) bool {
	for f.source.Next(row) {
		// TODO don't allocate this repeatedly inside the loop to avoid GC?
		ctx := &evaluator.EvalCtx{[]types.Row{*row}}
		if f.matcher.Matches(ctx) {
			return true
		} else {
			// the row from source does not match - keep iterating
			continue
		}
	}
	return false
}

func (f *Filter) Err() error {
	return nil
}
