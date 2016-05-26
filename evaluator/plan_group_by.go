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

// AggRowCtx holds evaluated data as well as the relevant context used to evaluate the data
// used for passing data - used to process aggregation functions - between iterators.
type aggRowCtx struct {
	// Row contains the evaluated data for each record.
	Row Row
	// Ctx contains the rows used in evaluating any aggregation
	// function used in the GROUP BY expression.
	Ctx []Row
}

type GroupByStage struct {
	// selectExprs holds the SelectExpression that should
	// be present in the result of a grouping. This will
	// include SelectExpressions for aggregates that might
	// not be projected, but are required for further
	// processing, such as when ordering by an aggregate.
	selectExprs SelectExpressions

	// source is the operator that provides the data to group
	source PlanStage

	// keySelectExprs holds the expression(s) to group by. For example, in
	// select a, count(b) from foo group by a,
	// keyExprs will hold the parsed column name 'a'.
	keyExprs SelectExpressions
}

func NewGroupByStage(source PlanStage, keys SelectExpressions, aggregates SelectExpressions) *GroupByStage {
	return &GroupByStage{
		source:      source,
		keyExprs:    keys,
		selectExprs: aggregates,
	}
}

func (gb *GroupByStage) Columns() (columns []*Column) {
	for _, expr := range gb.selectExprs {
		column := &Column{
			Name:      expr.Name,
			Table:     expr.Table,
			SQLType:   expr.SQLType,
			MongoType: expr.MongoType,
		}
		columns = append(columns, column)
	}
	return columns
}

// GroupBy groups records according to one or more fields.
type GroupByIter struct {
	source Iter

	selectExprs SelectExpressions
	keyExprs    SelectExpressions

	// grouped indicates if the source operator data has been grouped
	grouped bool

	// err holds any error encountered during processing
	err error

	// finalGrouping contains all grouped records and an ordered list of
	// the keys as read from the source operator
	finalGrouping orderedGroup

	// channel on which to send rows derived from the final grouping
	outChan chan aggRowCtx

	ctx *ExecutionCtx
}

func (gb *GroupByStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := gb.source.Open(ctx)
	if err != nil {
		return nil, err
	}

	iter := &GroupByIter{
		ctx:         ctx,
		source:      sourceIter,
		selectExprs: gb.selectExprs,
		keyExprs:    gb.keyExprs,
	}

	return iter, nil
}

func evaluateGroupByKey(row *Row, keyExprs SelectExpressions) (string, error) {

	var gbKey string

	for _, expr := range keyExprs {

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

func (gb *GroupByIter) createGroups() error {

	gb.finalGrouping = orderedGroup{
		groups: make(map[string][]Row, 0),
	}

	r := &Row{}

	// iterator source to create groupings
	for gb.source.Next(r) {

		key, err := evaluateGroupByKey(r, gb.keyExprs)
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

func (gb *GroupByIter) evalAggRow(r []Row) (*Row, error) {

	row := &Row{}
	evalCtx := &EvalCtx{Rows: r}

	for _, sExpr := range gb.selectExprs {

		v, err := sExpr.Expr.Evaluate(evalCtx)
		if err != nil {
			return nil, err
		}

		value := Value{
			Table: sExpr.Table,
			Name:  sExpr.Name,
			Data:  v,
		}

		row.Data = append(row.Data, value)
	}

	return row, nil
}

func (gb *GroupByIter) iterChan() chan aggRowCtx {
	ch := make(chan aggRowCtx)

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
				ch <- aggRowCtx{*r, v}
			}
		}
		close(ch)
	}()
	return ch
}

func (gb *GroupByIter) Next(row *Row) bool {
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

func (gb *GroupByIter) Close() error {
	return gb.source.Close()
}

func (gb *GroupByIter) Err() error {
	if err := gb.source.Err(); err != nil {
		return err
	}
	return gb.err
}

func (gb *GroupByIter) String() string {

	b := bytes.NewBufferString("select exprs ( ")

	for _, expr := range gb.selectExprs {
		b.WriteString(fmt.Sprintf("'%v' ", expr.Name))
	}

	b.WriteString(") grouped by ( ")

	for _, expr := range gb.keyExprs {
		b.WriteString(fmt.Sprintf("'%v' ", expr.Name))
	}

	b.WriteString(")")

	return b.String()

}

func (gb *GroupByStage) clone() *GroupByStage {
	return &GroupByStage{
		source:      gb.source,
		keyExprs:    gb.keyExprs,
		selectExprs: gb.selectExprs,
	}
}
