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
}

type groupByResult map[int][]*Row

func (gb *GroupBy) Open(ctx *ExecutionCtx) error {
	return gb.source.Open(ctx)
}

func (gb *GroupBy) getSelectView(group string) string {
	for _, column := range gb.fields[0].Columns {
		if column.Name == group {
			return column.View
		}
	}
	return ""
}

func (gb *GroupBy) group() error {
	r := &Row{}

	// TODO: support aggregate functions
	var groups []string

	for _, expr := range gb.exprs {
		column, ok := expr.(*sqlparser.ColName)
		if ok {
			name := string(column.Name)
			group := gb.getSelectView(name)
			groups = append(groups, group)
		} else {
			return fmt.Errorf("%T not supported as group type")
		}
	}

	// iterator source to create groupings
	for gb.source.Next(r) {

	}

	return gb.source.Err()
}

func (gb *GroupBy) Next(row *Row) bool {
	if !gb.grouped {
		if err := gb.group(); err != nil {
			gb.err = err
			return false
		}
	}

	// return gb.next(row)
	return false
}

func (gb *GroupBy) Close() error {
	return gb.source.Close()
}

func (gb *GroupBy) OpFields() []*Column {
	return gb.source.OpFields()
}

func (gb *GroupBy) Err() error {
	return gb.err
}
