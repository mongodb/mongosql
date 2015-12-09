package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"strings"
)

// Select returns results for the columns referenced in a
// select query.
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
	hasExpr := false

	for _, expr := range s.sExprs {
		if !expr.Referenced {
			hasExpr = true
		}
	}

	if !hasExpr {
		s.setSelectExpr()
	}

	return nil
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
		return false
	}

	// star expression, take headers from source
	if len(s.sExprs) == 0 {
		s.setSelectExpr()
	}

	data := map[string][]Value{}

	for _, expr := range s.sExprs {

		v, err := s.getValue(expr, row)
		if err != nil {
			s.err = err
			hasNext = false
		}

		value := Value{
			Name: expr.Name,
			View: expr.View,
			Data: v,
		}

		data[expr.Table] = append(data[expr.Table], value)
	}

	for k, v := range data {
		r.Data = append(r.Data, TableRow{k, v, nil})
	}

	return hasNext
}

var systemVars = map[string]SQLValue{
	"max_allowed_packet": SQLInt(4194304),
}

func (s *Select) getValue(sc SelectExpression, row *Row) (SQLValue, error) {
	// in the case where we have a bare select column and no expression
	if sc.Expr == nil {
		sc.Expr = &sqlparser.ColName{
			Name:      []byte(sc.Name),
			Qualifier: []byte(sc.Table),
		}
	} else {
		// If the column name is actually referencing a system variable, look it up and return
		// its value if it exists.

		// TODO scope system variables per-connection?
		if strings.HasPrefix(sc.Name, "@@") {
			if varValue, hasKey := systemVars[sc.Name[2:]]; hasKey {
				return varValue, nil
			}
			return nil, fmt.Errorf("unknown system variable %v", sc.Name)
		}
	}

	expr, err := NewSQLExpr(sc.Expr)
	if err != nil {
		return nil, err
	}

	evalCtx := &EvalCtx{
		Rows:    []Row{*row},
		ExecCtx: s.ctx,
	}

	return expr.Evaluate(evalCtx)
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
	if err := s.source.Err(); err != nil {
		return err
	}

	return s.err
}

func (s *Select) setSelectExpr() {
	for _, column := range s.source.OpFields() {
		sExpr := SelectExpression{*column, []*Column{column}, nil, false}
		s.sExprs = append(s.sExprs, sExpr)
	}
}
