package evaluator

type SubqueryStage struct {
	// tableName holds the name of this table as seen in outer contexts
	tableName string

	// source holds the source for this select statement
	source PlanStage
}

type SubqueryIter struct {
	plan   *SubqueryStage
	source Iter
}

func (sq *SubqueryStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := sq.source.Open(ctx)
	if err != nil {
		return nil, err
	}
	return &SubqueryIter{sq, sourceIter}, nil
}

func (sq *SubqueryIter) Next(row *Row) bool {

	hasNext := sq.source.Next(row)

	if !hasNext {
		return false
	}

	var tableRows TableRows

	for _, tableRow := range row.Data {

		var values Values

		for _, value := range tableRow.Values {
			value.Name = value.View
			values = append(values, value)
		}

		tableRow.Values = values
		tableRow.Table = sq.plan.tableName

		tableRows = append(tableRows, tableRow)
	}

	row.Data = tableRows

	return true
}

func (sq *SubqueryStage) OpFields() (columns []*Column) {
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

func (sq *SubqueryIter) Close() error {
	return sq.source.Close()
}

func (sq *SubqueryIter) Err() error {
	return sq.source.Err()
}
