package evaluator

// Project ensures that referenced columns - e.g. those used to
// support ORDER BY and GROUP BY clauses - aren't included in
// the final result set.
type Project struct {
	// sExprs holds the columns and/or expressions used in
	// the pipeline
	sExprs SelectExpressions
	// viewColumns holds the final list of columns we return
	// to the client
	viewColumns []*Column
	// source is the operator that provides the data to project
	source Operator
}

func (pj *Project) Open(ctx *ExecutionCtx) error {

	err := pj.source.Open(ctx)

	for _, sExpr := range pj.sExprs {
		if !sExpr.Referenced {
			viewColumn := sExpr.Column
			pj.viewColumns = append(pj.viewColumns, &viewColumn)
		}
	}

	if len(pj.viewColumns) == 0 {
		pj.viewColumns = pj.source.OpFields()
	}

	return err
}

func (pj *Project) Next(r *Row) bool {

	hasNext := pj.source.Next(r)
	tableValues := map[string]Values{}

	for _, column := range pj.viewColumns {
		field, _ := r.GetField(column.Table, column.Name)
		value := Value{column.Name, column.View, field}
		tableValues[column.Table] = append(tableValues[column.Table], value)
	}

	r.Data = []TableRow{}

	for table, values := range tableValues {
		r.Data = append(r.Data, TableRow{table, values, nil})
	}

	return hasNext
}

func (pj *Project) OpFields() (columns []*Column) {
	return pj.viewColumns
}

func (pj *Project) Close() error {
	return pj.source.Close()
}

func (pj *Project) Err() error {
	return pj.source.Err()
}
