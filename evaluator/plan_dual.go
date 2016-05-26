package evaluator

// Dual simulates a source for queries that don't require fields.
// It only ever returns one row.
type DualStage struct{}

func NewDualStage() *DualStage {
	return &DualStage{}
}

type DualIter struct {
	called bool
}

func (d *DualStage) Open(ctx *ExecutionCtx) (Iter, error) {
	return &DualIter{}, nil
}

func (di *DualIter) Next(row *Row) bool {
	if !di.called {
		di.called = true
		return true
	}
	return false
}

func (d *DualStage) Columns() (columns []*Column) {
	return []*Column{}
}

func (_ *DualIter) Close() error {
	return nil
}

func (_ *DualIter) Err() error {
	return nil
}
