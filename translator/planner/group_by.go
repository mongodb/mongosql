package planner

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
)

type GroupBy struct {
	// fields indicates the columns of the prior select operator
	fields []SelectColumn
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
}

func (gb *GroupBy) getSelectView(group string) string {
	for _, column := range gb.fields[0].Columns {
		if column.Name == group {
			return column.View
		}
	}
	return ""
}

func (gb *GroupBy) evaluateGroupByKey(keys []*sqlparser.ColName, row *Row) string {

	var gbKey string

	for _, key := range keys {
		// TODO: add ability to evaluate arbitrary key
		value, _ := row.GetField(string(key.Qualifier), string(key.Name))

		// TODO: would be better to use a hash for this
		gbKey += fmt.Sprintf("%#v", value)
	}

	return gbKey

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
		key := gb.evaluateGroupByKey(columns, r)
		gb.finalGrouping[key] = append(gb.finalGrouping[key], *r)
		r = &Row{}
	}

	gb.grouped = true

	return gb.source.Err()
}

func (gb *GroupBy) iterChan() chan Row {
	ch := make(chan Row)
	go func() {
		for _, v := range gb.finalGrouping {
			for _, r := range v {
				ch <- r
			}
		}
		close(ch)
	}()
	return ch
}

func (gb *GroupBy) Next(row *Row) bool {
	if /* len(gb.aggFuncs()) != 0 && */ !gb.grouped {
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

func (gb *GroupBy) Open(ctx *ExecutionCtx) error {
	return gb.source.Open(ctx)
}

func (gb *GroupBy) Close() error {
	return gb.source.Close()
}

func (gb *GroupBy) Err() error {
	return gb.err
}

func (gb *GroupBy) OpFields() []*Column {
	return gb.source.OpFields()
}
