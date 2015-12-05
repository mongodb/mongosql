package evaluator

// Having serves as a filter for HAVING expressions -
// involving aggregate functions.
type Having struct {
	// sExprs holds the columns and/or expressions present in
	// the source operator
	sExprs SelectExpressions

	// source is the operator that provides the data to filter
	source Operator

	// err holds any error encountered during processing
	err error

	// data holds all the paged in data from the source operator
	data []Row

	// matcher is used to filter results based on a HAVING clause
	matcher Matcher

	// hasNext indicates if this operator has more results
	hasNext bool
}

func (hv *Having) Open(ctx *ExecutionCtx) error {
	return hv.source.Open(ctx)
}

func (hv *Having) fetchData() error {

	hv.data = []Row{}

	r := &Row{}

	// iterator source to create groupings
	for hv.source.Next(r) {
		hv.data = append(hv.data, *r)
		r = &Row{}
	}

	return hv.source.Err()
}

func (hv *Having) evalResult() (*Row, error) {

	aggValues := map[string]Values{}

	row := &Row{}
	evalCtx := &EvalCtx{Rows: hv.data}

	for _, sExpr := range hv.sExprs {
		if sExpr.Referenced {
			continue
		}

		expr, err := NewSQLValue(sExpr.Expr)
		if err != nil {
			return nil, err
		}

		v, err := expr.Evaluate(evalCtx)
		if err != nil {
			return nil, err
		}

		value := Value{
			Name: sExpr.Name,
			View: sExpr.View,
			Data: v,
		}

		aggValues[sExpr.Table] = append(aggValues[sExpr.Table], value)

	}

	for table, values := range aggValues {
		row.Data = append(row.Data, TableRow{table, values, nil})
	}

	return row, nil
}

func (hv *Having) Next(row *Row) bool {
	hv.hasNext = !hv.hasNext

	if !hv.hasNext {
		return false
	}

	if err := hv.fetchData(); err != nil {
		hv.err = err
		return false
	}

	r, err := hv.evalResult()
	if err != nil {
		hv.err = err
		return false
	}

	evalCtx := &EvalCtx{Rows: hv.data}

	m, err := hv.matcher.Matches(evalCtx)
	if err != nil {
		hv.err = err
		return false
	}

	if m {
		row.Data = r.Data
	} else {
		return false
	}

	return true
}

func (hv *Having) OpFields() (columns []*Column) {
	for _, sExpr := range hv.sExprs {
		if sExpr.Referenced {
			continue
		}
		column := &Column{
			Name:  sExpr.Name,
			View:  sExpr.View,
			Table: sExpr.Table,
		}
		columns = append(columns, column)
	}
	return columns
}

func (hv *Having) Close() error {
	return hv.source.Close()
}

func (hv *Having) Err() error {
	if err := hv.source.Err(); err != nil {
		return err
	}
	return hv.err
}
