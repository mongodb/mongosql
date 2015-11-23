package evaluator

type AliasedSource struct {
	// tableName holds the name of this table as seen in outer contexts
	tableName string

	// err holds any error that may have occurred during processing
	err error

	// source holds the source for this select statement
	source Operator

	// ctx is the current execution context
	ctx *ExecutionCtx
}

func (as *AliasedSource) Open(ctx *ExecutionCtx) error {
	as.ctx = ctx
	return as.source.Open(ctx)
}

func (as *AliasedSource) Next(row *Row) bool {

	hasNext := as.source.Next(row)

	if !hasNext {
		return false
	}

	var tableRows []TableRow

	for _, tableRow := range row.Data {

		var values Values

		for _, value := range tableRow.Values {
			value.Name = value.View
			values = append(values, value)
		}

		tableRow.Values = values
		tableRow.Table = as.tableName

		tableRows = append(tableRows, tableRow)
	}

	row.Data = tableRows

	return true
}

func (as *AliasedSource) OpFields() (columns []*Column) {
	for _, expr := range as.source.OpFields() {
		column := &Column{
			// name is referenced by view in outer context
			Name:  expr.View,
			View:  expr.View,
			Table: as.tableName,
		}
		columns = append(columns, column)
	}

	return columns
}

func (as *AliasedSource) Close() error {
	return as.source.Close()
}

func (as *AliasedSource) Err() error {
	return as.err
}
