package evaluator

import "github.com/10gen/sqlproxy/collation"

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

func (d *DualStage) Collation() *collation.Collation {
	return collation.Default
}

func (_ *DualIter) Close() error {
	return nil
}

func (_ *DualIter) Err() error {
	return nil
}
