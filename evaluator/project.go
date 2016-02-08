package evaluator

import (
	"fmt"
	"strings"
)

// Project ensures that referenced columns - e.g. those used to
// support ORDER BY and GROUP BY clauses - aren't included in
// the final result set.
type Project struct {
	// sExprs holds the columns and/or expressions used in
	// the pipeline
	sExprs SelectExpressions

	// err holds any error that may have occurred during processing
	err error

	// source is the operator that provides the data to project
	source Operator

	// ctx is the current execution context
	ctx *ExecutionCtx
}

var systemVars = map[string]SQLValue{
	"max_allowed_packet": SQLInt(4194304),
}

func (pj *Project) Open(ctx *ExecutionCtx) error {

	pj.ctx = ctx

	// no select field implies a star expression - so we use
	// the fields from the source operator.
	hasExpr := false

	for _, expr := range pj.sExprs {
		if !expr.Referenced {
			hasExpr = true
			break
		}
	}

	err := pj.source.Open(ctx)

	if !hasExpr {
		pj.addSelectExprs()
	}

	return err
}

func (pj *Project) getValue(se SelectExpression, row *Row) (SQLValue, error) {
	// in the case where we have a bare select column and no expression
	if se.Expr == nil {
		se.Expr = SQLColumnExpr{se.Table, se.Name}
	} else {
		// If the column name is actually referencing a system variable, look it up and return
		// its value if it exists.

		// TODO scope system variables per-connection?
		if strings.HasPrefix(se.Name, "@@") {
			if varValue, hasKey := systemVars[se.Name[2:]]; hasKey {
				return varValue, nil
			}
			return nil, fmt.Errorf("unknown system variable %v", se.Name)
		}
	}

	evalCtx := &EvalCtx{
		Rows:    Rows{*row},
		ExecCtx: pj.ctx,
	}

	return se.Expr.Evaluate(evalCtx)
}

func (pj *Project) Next(r *Row) bool {

	hasNext := pj.source.Next(r)

	data := map[string]Values{}

	for _, expr := range pj.sExprs {

		if expr.Referenced {
			continue
		}

		value := Value{
			Name: expr.Name,
			View: expr.View,
		}

		v, ok := r.GetField(expr.Table, expr.Name)
		if !ok {
			v, err := pj.getValue(expr, r)
			if err != nil {
				pj.err = err
				hasNext = false
			}
			value.Data = v
		} else {
			value.Data = v
		}

		data[expr.Table] = append(data[expr.Table], value)
	}

	r.Data = TableRows{}

	for k, v := range data {
		r.Data = append(r.Data, TableRow{k, v})
	}

	return hasNext
}

func (pj *Project) OpFields() (columns []*Column) {
	for _, expr := range pj.sExprs {
		if expr.Referenced {
			continue
		}
		column := &Column{
			Name:  expr.Name,
			View:  expr.View,
			Table: expr.Table,
		}
		columns = append(columns, column)
	}

	if len(columns) == 0 {
		columns = pj.source.OpFields()
	}

	return columns
}

func (pj *Project) Close() error {
	return pj.source.Close()
}

func (pj *Project) Err() error {
	return pj.source.Err()
}

func (pj *Project) addSelectExprs() {
	for _, column := range pj.source.OpFields() {
		sExpr := SelectExpression{
			Column:     column,
			RefColumns: []*Column{column},
			Expr:       nil,
			Referenced: false,
		}
		pj.sExprs = append(pj.sExprs, sExpr)
	}
}
