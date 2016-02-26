package evaluator

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
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

	if !hasNext {
		return false
	}

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

	return true
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
	if err := pj.source.Err(); err != nil {
		return err
	}
	return pj.err
}

func (pj *Project) addSelectExprs() {
	for _, column := range pj.source.OpFields() {
		sExpr := SelectExpression{
			Column:     column,
			RefColumns: []*Column{column},
			Expr:       nil,
		}
		pj.sExprs = append(pj.sExprs, sExpr)
	}
}

func (pj *Project) WithSource(source Operator) *Project {
	return &Project{
		sExprs: pj.sExprs,
		source: source,
	}
}

// Optimizations

func (v *optimizer) visitProject(project *Project) (Operator, error) {
	// Check if we can optimize further, if the child operator has a MongoSource.
	ms, ok := canPushDown(project.source)
	if !ok {
		return project, nil
	}

	fieldsToProject := bson.M{}

	// This will contain the rebuilt mapping registry reflecting fields re-mapped by projection.
	fixedMappingRegistry := mappingRegistry{}

	fixedExpressions := SelectExpressions{}

	// Track whether or not we've successfully mapped every field into the $project of the source.
	// If so, this Project node can be removed from the query plan tree.
	canReplaceProject := true

	for _, exp := range project.sExprs {

		// Convert the column's SQL expression into an expression in mongo query language.
		projectedField, ok := TranslateExpr(exp.Expr, ms.mappingRegistry.lookupFieldName)
		if !ok {
			// Expression can't be translated, so it can't be projected.
			// We skip it and leave this Project node in the query plan so that it still gets
			// evaluated during execution.
			canReplaceProject = false
			fixedExpressions = append(fixedExpressions, exp)

			// There might still be fields referenced in this expression
			// that we still need to project, so collect them and add them to the projection.
			refdCols, err := referencedColumns(exp.Expr)
			if err != nil {
				return nil, err
			}
			for _, refdCol := range refdCols {
				fieldName, ok := ms.mappingRegistry.lookupFieldName(refdCol.Table, refdCol.Name)
				if !ok {
					// TODO log that optimization gave up here.
					return project, nil
				}

				fieldsToProject[dottifyFieldName(fieldName)] = getProjectedFieldName(fieldName)
			}
			continue
		}

		safeFieldName := dottifyFieldName(exp.Expr.String())

		fieldsToProject[safeFieldName] = projectedField

		colType, ok := ms.mappingRegistry.lookupFieldType(exp.Column.Table, exp.Column.Name)
		if ok {
			exp.Column.Type = colType
		}
		fixedMappingRegistry.addColumn(exp.Column)
		fixedMappingRegistry.registerMapping(exp.Column.Table, exp.Column.Name, safeFieldName)

		fixedExpressions = append(fixedExpressions,
			SelectExpression{
				Column: exp.Column,
				Expr:   SQLColumnExpr{exp.Column.Table, exp.Column.Name},
			},
		)

	}

	pipeline := ms.pipeline
	pipeline = append(pipeline, bson.D{{"$project", fieldsToProject}})
	ms = ms.WithPipeline(pipeline).WithMappingRegistry(&fixedMappingRegistry)

	if canReplaceProject {
		return ms, nil
	}

	return project.WithSource(ms), nil
}
