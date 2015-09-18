package planner

import (
	"github.com/erh/mixer/sqlparser"
	"gopkg.in/mgo.v2/bson"
)

type Select struct {
	// selectColumns holds information on the columns referenced in each
	// select expression
	selectColumns SelectColumns

	// err holds any error that may have occurred during processing
	err error

	// source holds the source for this select statement
	source Operator

	// ctx is the current execution context
	ctx *ExecutionCtx
}

// SelectColumn holds all columns referenced in select expressions.
type SelectColumn struct {
	Column
	Columns []*Column
	Expr    sqlparser.Expr
}

func (s *Select) Open(ctx *ExecutionCtx) error {

	if err := s.source.Open(ctx); err != nil {
		return err
	}

	s.ctx = ctx

	// no select expression imply a star expression - in which case
	// we grab all the fields
	if len(s.selectColumns) == 0 {
		s.setSelectExpr()
	}

	return nil
}

func (s *Select) setSelectExpr() {
	for _, column := range s.source.OpFields() {
		sExpr := SelectColumn{*column, []*Column{column}, nil}
		s.selectColumns = append(s.selectColumns, sExpr)
	}
}

func (s *Select) Next(r *Row) bool {

	row := &Row{}

	hasNext := s.source.Next(row)

	if !hasNext {

		if err := s.source.Err(); err != nil {
			s.err = err
		} else if err := s.source.Close(); err != nil {
			s.err = err
		}

	}

	// star expression, take headers from source
	if len(s.selectColumns) == 0 {
		s.setSelectExpr()
	}

	data := map[string][]bson.DocElem{}

	for _, expr := range s.selectColumns {

		t, v, err := s.getValue(expr, row)
		if err != nil {
			s.err = err
			hasNext = false
		}
		data[t] = append(data[t], v)
	}

	for k, v := range data {
		r.Data = append(r.Data, TableRow{k, v, nil})
	}

	return hasNext
}

func (s *Select) getValue(sc SelectColumn, row *Row) (string, bson.DocElem, error) {
	// in case we have a bare select column with no expression
	if sc.Expr == nil {
		sc.Expr = &sqlparser.ColName{
			Name:      []byte(sc.Name),
			Qualifier: []byte(sc.Table),
		}
	}

	expr, err := NewExpr(sc.Expr)
	if err != nil {
		panic(err)
	}

	s.ctx.Row = *row
	v, err := expr.Evaluate(s.ctx)

	return sc.Table, bson.DocElem{sc.View, v}, err
}

func (s *Select) Close() error {
	return s.source.Close()
}

func (s *Select) OpFields() (columns []*Column) {

	for _, expr := range s.selectColumns {
		column := &Column{
			Name:  expr.View,
			View:  expr.View,
			Table: expr.Table,
		}
		columns = append(columns, column)
	}

	return columns
}

func (s *Select) Err() error {
	return s.err
}

type SelectColumns []SelectColumn

func (sc SelectColumns) GetColumns() []*Column {

	columns := make([]*Column, 0)

	for _, selectColumn := range sc {
		columns = append(columns, selectColumn.Columns...)
	}

	return columns
}
