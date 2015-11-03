package evaluator

// Dual simulates a source for queries that don't require fields.
// It only ever returns one row.
type Dual struct {
	called bool
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

func (_ *Dual) OpFields() (columns []*Column) {
	return nil
}

func (_ *Dual) Close() error {
	return nil
}

func (_ *Dual) Err() error {
	return nil
}
