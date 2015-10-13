package evaluator

type Subquery struct {
	// tableName holds the name of this table as seen in outer contexts
	tableName string

	// err holds any error that may have occurred during processing
	err error

	// source holds the source for this select statement
	source Operator

	// matcher is used to filter results returned by this operator
	matcher Matcher

	// ctx is the current execution context
	ctx *ExecutionCtx
}

func (sq *Subquery) Open(ctx *ExecutionCtx) error {
	sq.ctx = ctx
	return sq.source.Open(ctx)
}

func (sq *Subquery) Next(row *Row) bool {

	var hasNext bool

	for {

		hasNext = sq.source.Next(row)

		var tableRows []TableRow

		for _, tableRow := range row.Data {

			var values Values

			for _, value := range tableRow.Values {
				value.Name = value.View
				values = append(values, value)
			}

			tableRow.Values = values
			tableRow.Table = sq.tableName

			tableRows = append(tableRows, tableRow)
		}

		row.Data = tableRows

		evalCtx := &EvalCtx{[]Row{*row}, sq.ctx}

		if sq.matcher != nil {
			m, err := sq.matcher.Matches(evalCtx)
			if err != nil {
				sq.err = err
				return false
			}
			if m {
				break
			}
		} else {
			break
		}

		if !hasNext {
			break
		}

	}

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
