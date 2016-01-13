package evaluator

// Dual simulates a source for queries that don't require fields.
// It only ever returns one row.
type Dual struct {
	called bool

	sExprs SelectExpressions
}

func (_ *Dual) Open(ctx *ExecutionCtx) error {
	return nil
}

func (d *Dual) Next(row *Row) bool {
	if !d.called {
		d.called = true
		return true
	}
	return false
}

func (d *Dual) OpFields() (columns []*Column) {
	for _, expr := range d.sExprs {
		column := &Column{
			Name:  expr.Name,
			View:  expr.View,
			Table: expr.Table,
		}
		columns = append(columns, column)
	}

	return columns
}

func (_ *Dual) Close() error {
	return nil
}

func (_ *Dual) Err() error {
	return nil
}
