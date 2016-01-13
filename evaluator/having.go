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

	// matcher is used to filter results based on a HAVING clause
	matcher SQLExpr

	// hasNext indicates if this operator has more results
	hasNext bool
}

func (hv *Having) Open(ctx *ExecutionCtx) error {
	return hv.source.Open(ctx)
}

func (hv *Having) Next(r *Row) bool {

	hv.hasNext = !hv.hasNext

	if !hv.hasNext {
		return false
	}
	rows := Rows{}

	for hv.source.Next(r) {
		rows = append(rows, *r)
		r = &Row{}
	}

	if err := hv.source.Err(); err != nil {
		hv.err = err
		return false
	}

	evalCtx := &EvalCtx{Rows: rows}

	m, err := Matches(hv.matcher, evalCtx)
	if err != nil {
		hv.err = err
		return false
	}

	return m
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
