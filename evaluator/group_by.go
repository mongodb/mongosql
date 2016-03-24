package evaluator

import (
	"bytes"
	"fmt"
)

// orderedGroup holds all the rows belonging to a given key in the groups
// and an slice of the keys for each group.
type orderedGroup struct {
	groups map[string][]Row
	keys   []string
}

// GroupBy groups records according to one or more fields.
type GroupBy struct {
	// selectExprs holds the SelectExpression that should
	// be present in the result of a grouping. This will
	// include SelectExpressions for aggregates that might
	// not be projected, but are required for further
	// processing, such as when ordering by an aggregate.
	selectExprs SelectExpressions

	// source is the operator that provides the data to group
	source Operator

	// keySelectExprs holds the expression(s) to group by. For example, in
	// select a, count(b) from foo group by a,
	// keyExprs will hold the parsed column name 'a'.
	keyExprs SelectExpressions

	// grouped indicates if the source operator data has been grouped
	grouped bool

	// err holds any error encountered during processing
	err error

	// finalGrouping contains all grouped records and an ordered list of
	// the keys as read from the source operator
	finalGrouping orderedGroup

	// channel on which to send rows derived from the final grouping
	outChan chan AggRowCtx

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

	for _, expr := range gb.keyExprs {
		evalCtx := &EvalCtx{Rows: Rows{*row}}
		value, err := expr.Expr.Evaluate(evalCtx)
		if err != nil {
			return "", err
		}

		// TODO: might be better to use a hash for this
		gbKey += fmt.Sprintf("%#v", value)
	}

	return gbKey, nil
}

func (gb *GroupBy) createGroups() error {

	gb.finalGrouping = orderedGroup{
		groups: make(map[string][]Row, 0),
	}

	r := &Row{}

	// iterator source to create groupings
	for gb.source.Next(r) {

		key, err := gb.evaluateGroupByKey(r)
		if err != nil {
			return err
		}

		if gb.finalGrouping.groups[key] == nil {
			gb.finalGrouping.keys = append(gb.finalGrouping.keys, key)
		}

		gb.finalGrouping.groups[key] = append(gb.finalGrouping.groups[key], *r)

		r = &Row{}
	}

	gb.grouped = true

	return gb.source.Err()
}

func (gb *GroupBy) evalAggRow(r []Row) (*Row, error) {

	aggValues := map[string]Values{}

	row := &Row{}

	for _, sExpr := range gb.selectExprs {

		evalCtx := &EvalCtx{Rows: r}

		v, err := sExpr.Expr.Evaluate(evalCtx)
		if err != nil {
			return nil, err
		}

		value := Value{
			Name: sExpr.Name,
			View: sExpr.View,
			Data: v,
		}
		aggValues[sExpr.Table] = append(aggValues[sExpr.Table], value)
	}

	for table, values := range aggValues {
		row.Data = append(row.Data, TableRow{table, values})
	}

	return row, nil
}

func (gb *GroupBy) iterChan() chan AggRowCtx {
	ch := make(chan AggRowCtx)

	go func() {
		for _, key := range gb.finalGrouping.keys {
			v := gb.finalGrouping.groups[key]
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
	for _, sExpr := range gb.selectExprs {
		column := &Column{
			Name:  sExpr.Name,
			View:  sExpr.View,
			Table: sExpr.Table,
		}
		columns = append(columns, column)
	}
	return columns
}

func (gb *GroupBy) String() string {

	b := bytes.NewBufferString("select exprs ( ")

	for _, expr := range gb.selectExprs {
		b.WriteString(fmt.Sprintf("'%v' ", expr.View))
	}

	b.WriteString(") grouped by ( ")

	for _, expr := range gb.keyExprs {
		b.WriteString(fmt.Sprintf("'%v' ", expr.View))
	}

	b.WriteString(")")

	return b.String()

}
