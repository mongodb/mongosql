package evaluator

import (
	"fmt"
	"strings"
)

// Project ensures that referenced columns - e.g. those used to
// support ORDER BY and GROUP BY clauses - aren't included in
// the final result set.
type ProjectStage struct {
	// sExprs holds the columns and/or expressions used in
	// the pipeline
	sExprs SelectExpressions

	// source is the operator that provides the data to project
	source PlanStage
}

// NewProjectStage creates a new project stage.
func NewProjectStage(source PlanStage, sExprs ...SelectExpression) *ProjectStage {
	return &ProjectStage{
		source: source,
		sExprs: sExprs,
	}
}

type ProjectIter struct {
	source Iter

	sExprs SelectExpressions

	// err holds any error that may have occurred during processing
	err error

	// ctx is the current execution context
	ctx *ExecutionCtx
}

var systemVars = map[string]SQLValue{
	"max_allowed_packet": SQLInt(4194304),
}

func (pj *ProjectStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := pj.source.Open(ctx)
	if err != nil {
		return nil, err
	}

	return &ProjectIter{
		sExprs: pj.sExprs,
		ctx:    ctx,
		source: sourceIter,
	}, nil

}

func (pj *ProjectIter) getValue(se SelectExpression, row *Row) (SQLValue, error) {

	// If the column name is actually referencing a system variable, look it up and return
	// its value if it exists.

	// TODO scope system variables per-connection?
	if strings.HasPrefix(se.Name, "@@") {
		if varValue, hasKey := systemVars[se.Name[2:]]; hasKey {
			return varValue, nil
		}
		return nil, fmt.Errorf("unknown system variable %v", se.Name)
	}

	evalCtx := &EvalCtx{
		Rows:    Rows{*row},
		ExecCtx: pj.ctx,
	}

	return se.Expr.Evaluate(evalCtx)
}

func (pj *ProjectIter) Next(r *Row) bool {

	hasNext := pj.source.Next(r)

	if !hasNext {
		return false
	}

	if len(pj.sExprs) > 0 {
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
	}

	return true
}

func (pj *ProjectStage) OpFields() (columns []*Column) {
	for _, expr := range pj.sExprs {
		column := &Column{
			Name:      expr.Name,
			View:      expr.View,
			Table:     expr.Table,
			SQLType:   expr.SQLType,
			MongoType: expr.MongoType,
		}
		columns = append(columns, column)
	}

	if len(columns) == 0 {
		columns = pj.source.OpFields()
	}

	return columns
}

func (pj *ProjectIter) Close() error {
	return pj.source.Close()
}

func (pj *ProjectIter) Err() error {
	if err := pj.source.Err(); err != nil {
		return err
	}
	return pj.err
}

func (pj *ProjectStage) clone() *ProjectStage {
	return &ProjectStage{
		source: pj.source,
		sExprs: pj.sExprs,
	}
}
