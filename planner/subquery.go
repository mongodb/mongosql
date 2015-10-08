package planner

import (
	"github.com/erh/mongo-sql-temp/evaluator"
)

type Subquery struct {
	// tableName holds the name of this table as seen in outer contexts
	tableName string

	// err holds any error that may have occurred during processing
	err error

	// source holds the source for this select statement
	source Operator

	// ctx is the current execution context
	ctx *ExecutionCtx
}

func (sq *Subquery) Open(ctx *ExecutionCtx) error {
	return sq.source.Open(ctx)
}

func (sq *Subquery) Next(row *evaluator.Row) bool {
	hasNext := sq.source.Next(row)

	var tableRows []evaluator.TableRow

	for _, tableRow := range row.Data {

		var values evaluator.Values

		for _, value := range tableRow.Values {
			value.Name = value.View
			values = append(values, value)
		}

		tableRow.Values = values
		tableRow.Table = sq.tableName

		tableRows = append(tableRows, tableRow)
	}

	row.Data = tableRows

	return hasNext
}

func (sq *Subquery) OpFields() (columns []*Column) {
	for _, expr := range sq.source.OpFields() {
		column := &Column{
			// name is referenced by view in outer context
			Name:  expr.View,
			View:  expr.View,
			Table: sq.tableName,
		}
		columns = append(columns, column)
	}

	return columns
}

func (sq *Subquery) Close() error {
	return sq.source.Close()
}

func (sq *Subquery) Err() error {
	return sq.err
}
