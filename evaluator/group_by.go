package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
)

// GroupBy groups records according to one or more fields.
type GroupBy struct {
	// sExprs holds the columns and/or expressions present in
	// the source operator
	sExprs SelectExpressions

	// source is the operator that provides the data to group
	source Operator

	// exprs holds the expression(s) to group by. For example, in
	// select a, count(b) from foo group by a
	// exprs will hold the parsed column name 'a'.
	exprs []sqlparser.Expr

	// grouped indicates if the source operator data has been grouped
	grouped bool

	// err holds any error encountered during processing
	err error

	// finalGrouping has a key derived from the group by clause and
	// a value corresponding to all rows that are part of the group
	finalGrouping map[string][]Row

	// channel on which to send rows derived from the final grouping
	outChan chan AggRowCtx

	// matcher is used to filter results based on a HAVING clause
	matcher SQLExpr

	// orderBy holds the expression(s) to order by. For example, in
	// select a, count(b) from foo group by a order by count(b)
	// exprs will hold the expression 'count(b)'
	orderBy sqlparser.OrderBy

	ctx *ExecutionCtx
}

func (gb *GroupBy) Open(ctx *ExecutionCtx) error {
	return gb.init(ctx)
}

func (gb *GroupBy) init(ctx *ExecutionCtx) error {
	gb.ctx = ctx
	return gb.source.Open(ctx)
}

func (gb *GroupBy) evaluateGroupByKey(row *Row) (string, error) {

	var gbKey string

	for _, key := range gb.exprs {
		expr, err := NewSQLExpr(key)
		if err != nil {
			panic(err)
		}

		evalCtx := &EvalCtx{Rows: []Row{*row}}
		value, err := expr.Evaluate(evalCtx)
		if err != nil {
			return "", err
		}

		// TODO: might be better to use a hash for this
		gbKey += fmt.Sprintf("%#v", value)
	}

	return gbKey, nil
}

func (gb *GroupBy) createGroups() error {

	gb.finalGrouping = make(map[string][]Row, 0)

	r := &Row{}

	// iterator source to create groupings
	for gb.source.Next(r) {
		key, err := gb.evaluateGroupByKey(r)
		if err != nil {
			return err
		}

		gb.finalGrouping[key] = append(gb.finalGrouping[key], *r)
		r = &Row{}
	}

	gb.grouped = true

	return gb.source.Err()
}

func (gb *GroupBy) evalAggRow(r []Row) (*Row, error) {

	aggValues := map[string]Values{}

	row := &Row{}

	for _, sExpr := range gb.sExprs {

		expr, err := NewSQLExpr(sExpr.Expr)
		if err != nil {
			return nil, err
		}

		evalCtx := &EvalCtx{Rows: r}

		v, err := expr.Evaluate(evalCtx)
		if err != nil {
			return nil, err
		}

		m, err := Matches(gb.matcher, evalCtx)
		if err != nil {
			return nil, err
		}

		if m {
			value := Value{
				Name: sExpr.Name,
				View: sExpr.View,
				Data: v,
			}
			aggValues[sExpr.Table] = append(aggValues[sExpr.Table], value)
		}
	}

	for table, values := range aggValues {
		row.Data = append(row.Data, TableRow{table, values, nil})
	}

	return row, nil
}

func (gb *GroupBy) iterChan() chan AggRowCtx {
	ch := make(chan AggRowCtx)

	go func() {
		for _, v := range gb.finalGrouping {
			r, err := gb.evalAggRow(v)
			if err != nil {
				gb.err = err
				close(ch)
				return
			}

			// check we have some matching data
			if len(r.Data) != 0 {
				ch <- AggRowCtx{*r, v}
			}
		}
		close(ch)
	}()
	return ch
}

func (gb *GroupBy) Next(row *Row) bool {
	if !gb.grouped {
		if err := gb.createGroups(); err != nil {
			gb.err = err
			return false
		}
		gb.outChan = gb.iterChan()
	}

	rCtx, done := <-gb.outChan
	gb.ctx.GroupRows = rCtx.Ctx
	row.Data = rCtx.Row.Data

	return done
}

func (gb *GroupBy) Close() error {
	return gb.source.Close()
}

func (gb *GroupBy) Err() error {
	if err := gb.source.Err(); err != nil {
		return err
	}
	return gb.err
}

func (gb *GroupBy) OpFields() (columns []*Column) {
	for _, sExpr := range gb.sExprs {
		column := &Column{
			Name:  sExpr.Name,
			View:  sExpr.View,
			Table: sExpr.Table,
		}
		columns = append(columns, column)
	}
	return columns
}
