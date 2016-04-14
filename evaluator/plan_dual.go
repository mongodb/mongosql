package evaluator

// Dual simulates a source for queries that don't require fields.
// It only ever returns one row.
type DualStage struct {
	sExprs SelectExpressions
}

type DualIter struct {
	sExprs SelectExpressions
	called bool
}

func (d *DualStage) Open(ctx *ExecutionCtx) (Iter, error) {
	return &DualIter{sExprs: d.sExprs}, nil
}

func (di *DualIter) Next(row *Row) bool {
	if !di.called {
		di.called = true
		return true
	}
	return false
}

func (d *DualStage) OpFields() (columns []*Column) {
	for _, expr := range d.sExprs {
		column := &Column{
			Name:      expr.Name,
			View:      expr.View,
			Table:     expr.Table,
			SQLType:   expr.SQLType,
			MongoType: expr.MongoType,
		}
		columns = append(columns, column)
	}

	return columns
}

func (_ *DualIter) Close() error {
	return nil
}

func (_ *DualIter) Err() error {
	return nil
}
