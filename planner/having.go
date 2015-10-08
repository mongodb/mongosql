package planner

import (
	"github.com/erh/mongo-sql-temp/evaluator"
)

type Having struct {
	// sExprs holds the columns and/or expressions present in
	// the source operator
	sExprs SelectExpressions
	// source is the operator that provides the data to filter
	source Operator
	// err holds any error encountered during processing
	err error
	// data holds all the paged in data from the source operator
	data []evaluator.Row
	// matcher is used to filter results based on a HAVING clause
	matcher evaluator.Matcher
	// hasNext indicates if this operator has more results
	hasNext bool
}

func (hv *Having) Open(ctx *ExecutionCtx) error {
	return hv.source.Open(ctx)
}

func (hv *Having) fetchData() error {

	hv.data = []evaluator.Row{}

	r := &evaluator.Row{}

	// iterator source to create groupings
	for hv.source.Next(r) {
		hv.data = append(hv.data, *r)
		r = &evaluator.Row{}
	}

	return hv.source.Err()
}

func (hv *Having) evalResult() (*evaluator.Row, error) {

	aggValues := map[string]evaluator.Values{}

	row := &evaluator.Row{}
	evalCtx := &evaluator.EvalCtx{Rows: hv.data}

	for _, sExpr := range hv.sExprs {
		if sExpr.Referenced {
			continue
		}

		expr, err := evaluator.NewSQLValue(sExpr.Expr)
		if err != nil {
			return nil, err
		}

		v, err := expr.Evaluate(evalCtx)
		if err != nil {
			return nil, err
		}

		value := evaluator.Value{
			Name: sExpr.Name,
			View: sExpr.View,
			Data: v,
		}

		aggValues[sExpr.Table] = append(aggValues[sExpr.Table], value)

	}

	for table, values := range aggValues {
		row.Data = append(row.Data, evaluator.TableRow{table, values, nil})
	}

	return row, nil
}

func (hv *Having) Next(row *evaluator.Row) bool {
	hv.hasNext = !hv.hasNext

	if !hv.hasNext {
		return false
	}

	if err := hv.fetchData(); err != nil {
		hv.err = err
		return false
	}

	r, err := hv.evalResult()
	if err != nil {
		hv.err = err
	}

	evalCtx := &evaluator.EvalCtx{Rows: hv.data}

	if hv.matcher.Matches(evalCtx) {
		row.Data = r.Data
	} else {
		return false
	}

	return true
}

func (hv *Having) OpFields() (columns []*Column) {
	for _, sExpr := range hv.sExprs {
		if sExpr.Referenced {
			continue
		}
		column := &Column{
			Name:  sExpr.Name,
			View:  sExpr.View,
			Table: sExpr.Table,
		}
		columns = append(columns, column)
	}
	return columns
}

func (hv *Having) Close() error {
	return hv.source.Close()
}

func (hv *Having) Err() error {
	return hv.err
}
