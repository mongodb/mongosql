package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/collation"
)

// A DualStage simulates a source for queries that don't require fields.
// It only ever returns one row.
type DualStage struct{}

// NewDualStage returns a new DualStage.
func NewDualStage() *DualStage {
	return &DualStage{}
}

// DualIter returns rows from a DualStage.
type DualIter struct {
	called bool
}

// Children returns a slice of all the Node children of the Node.
func (DualStage) Children() []Node {
	return []Node{}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (DualStage) ReplaceChild(i int, n Node) {
	panicWithInvalidIndex("DualStage", i, -1)
}

// Open returns an iterator that returns results from executing this plan stage
// with the given ExecutionContext.
func (d *DualStage) Open(_ context.Context, _ *ExecutionConfig, _ *ExecutionState) (RowIter, error) {
	return &DualIter{}, nil
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (di *DualIter) Next(_ context.Context, _ *Row) bool {
	if !di.called {
		di.called = true
		return true
	}
	return false
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (d *DualStage) Columns() (columns []*Column) {
	return []*Column{}
}

// Collation returns the collation to use for comparisons.
func (d *DualStage) Collation() *collation.Collation {
	return collation.Default
}

// Close closes the iterator, returning any error encountered while doing so.
func (*DualIter) Close() error {
	return nil
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (*DualIter) Err() error {
	return nil
}

func (*DualStage) clone() PlanStage {
	return NewDualStage()
}
