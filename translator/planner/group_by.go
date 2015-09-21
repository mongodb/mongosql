package planner

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"gopkg.in/mgo.v2/bson"
)

type GroupBy struct {
	// fields indicates the columns of the prior select operator
	fields SelectColumns
	// source is the operator that provides the data to group
	source Operator
	// exprs holds the expressions to group by
	exprs []sqlparser.Expr
	// grouped indicates if the source operator data has been grouped
	grouped bool
	// err holds any error encountered during processing
	err error
	// holds all groupings
	finalGrouping map[string][]Row
	// channel on which to send final grouping
	outChan chan Row
	// execution context
	ctx ExecutionCtx
}

func (gb *GroupBy) Open(ctx *ExecutionCtx) error {
	return gb.init(ctx)
}

func (gb *GroupBy) init(ctx *ExecutionCtx) error {
	gb.ctx = *ctx

	return gb.source.Open(ctx)
}

func (gb *GroupBy) evaluateGroupByKey(keys []*sqlparser.ColName, row *Row) (string, error) {

	var gbKey string

	for _, key := range keys {
		expr, err := NewExpr(key)
		if err != nil {
			panic(err)
		}

		evalCtx := &EvalCtx{rows: []Row{*row}}
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

	// TODO: support aggregate functions
	var columns []*sqlparser.ColName

	for _, expr := range gb.exprs {
		column, ok := expr.(*sqlparser.ColName)
		if !ok {
			return fmt.Errorf("%T not supported as group type")
		}
		columns = append(columns, column)
	}

	gb.finalGrouping = make(map[string][]Row, 0)

	r := &Row{}

	// iterator source to create groupings
	for gb.source.Next(r) {
		key, err := gb.evaluateGroupByKey(columns, r)
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

	aggValues := map[string]bson.D{}

	row := &Row{}

	gb.ctx.Rows = r

	for _, field := range gb.fields {

		expr, err := NewExpr(field.Expr)
		if err != nil {
			panic(err)
		}

		evalCtx := &EvalCtx{rows: r}

		value, err := expr.Evaluate(evalCtx)
		if err != nil {
			return nil, err
		}

		aggValues[field.Table] = append(aggValues[field.Table], bson.DocElem{field.Name, value})
	}

	for table, values := range aggValues {
		row.Data = append(row.Data, TableRow{table, values, nil})
	}

	return row, nil
}

func (gb *GroupBy) iterChan() chan Row {
	ch := make(chan Row)

	go func() {
		for _, v := range gb.finalGrouping {
			r, err := gb.evalAggRow(v)
			if err != nil {
				gb.err = err
				close(ch)
				return
			}
			ch <- *r
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

	r, done := <-gb.outChan
	row.Data = r.Data
	return done
}

func (gb *GroupBy) Close() error {
	return gb.source.Close()
}

func (gb *GroupBy) Err() error {
	return gb.err
}

func (gb *GroupBy) OpFields() (columns []*Column) {
	for _, field := range gb.fields {
		column := &Column{
			Name:  field.Name,
			View:  field.View,
			Table: field.Table,
		}
		columns = append(columns, column)
	}
	return columns
}
