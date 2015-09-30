package planner

import (
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/translator/evaluator"
	"github.com/erh/mongo-sql-temp/translator/types"
	"gopkg.in/mgo.v2/bson"
)

type Select struct {
	// sExprs holds information on the columns referenced in each
	// select expression
	sExprs SelectExpressions

	// err holds any error that may have occurred during processing
	err error

	// source holds the source for this select statement
	source Operator

	// ctx is the current execution context
	ctx *ExecutionCtx
}

func (s *Select) Open(ctx *ExecutionCtx) error {

	if err := s.source.Open(ctx); err != nil {
		return err
	}

	s.ctx = ctx

	// no select field implies a star expression - so we use
	// the fields from the source operator.
	if len(s.sExprs) == 0 {
		s.setSelectExpr()
	}

	return nil
}

func (s *Select) Next(r *types.Row) bool {
	row := &types.Row{}

	hasNext := s.source.Next(row)

	if !hasNext {

		if err := s.source.Err(); err != nil {
			s.err = err
		} else if err := s.source.Close(); err != nil {
			s.err = err
		}

	}

	// star expression, take headers from source
	if len(s.sExprs) == 0 {
		s.setSelectExpr()
	}

	data := map[string][]bson.DocElem{}

	for _, expr := range s.sExprs {

		t, v, err := s.getValue(expr, row)
		if err != nil {
			s.err = err
			hasNext = false
		}
		data[t] = append(data[t], v)
	}

	for k, v := range data {
		r.Data = append(r.Data, types.TableRow{k, v, nil})
	}

	return hasNext
}

func (s *Select) getValue(sc SelectExpression, row *types.Row) (string, bson.DocElem, error) {
	// in the case where we have a bare select column and no expression
	if sc.Expr == nil {
		sc.Expr = &sqlparser.ColName{
			Name:      []byte(sc.Name),
			Qualifier: []byte(sc.Table),
		}
	}

	expr, err := evaluator.NewSQLValue(sc.Expr)
	if err != nil {
		panic(err)
	}

	evalCtx := &evaluator.EvalCtx{Rows: []types.Row{*row}}
	v, err := expr.Evaluate(evalCtx)

	return sc.Table, bson.DocElem{sc.Name, v}, err
}

func (s *Select) OpFields() (columns []*Column) {
	for _, expr := range s.sExprs {
		column := &Column{
			Name:  expr.Name,
			View:  expr.View,
			Table: expr.Table,
		}
		columns = append(columns, column)
	}

	return columns
}

func (s *Select) Close() error {
	return s.source.Close()
}

func (s *Select) Err() error {
	return s.err
}

func (s *Select) setSelectExpr() {
	for _, column := range s.source.OpFields() {
		sExpr := SelectExpression{*column, []*Column{column}, nil}
		s.sExprs = append(s.sExprs, sExpr)
	}
}
