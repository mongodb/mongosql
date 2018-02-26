package evaluator

import "github.com/10gen/sqlproxy/collation"

// An EmptyStage is for when we find that 0 rows are going to be returned: we don't
// need to hit MongoDB to get back nothing.
type EmptyStage struct {
	columns   []*Column
	collation *collation.Collation
}

// NewEmptyStage creates a new Empty stage.
func NewEmptyStage(columns []*Column, collation *collation.Collation) *EmptyStage {
	return &EmptyStage{columns, collation}
}

// An EmptyIter returns no rows.
type EmptyIter struct{}

// Open returns an iterator that returns results from executing this plan stage
// with the given ExecutionContext.
func (*EmptyStage) Open(ctx *ExecutionCtx) (Iter, error) {
	return &EmptyIter{}, nil
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (es *EmptyStage) Columns() []*Column {
	return es.columns
}

// Collation returns the collation to use for comparisons.
func (es *EmptyStage) Collation() *collation.Collation {
	return es.collation
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (*EmptyIter) Next(row *Row) bool {
	return false
}

// Close closes the iterator, returning any error encountered while doing so.
func (*EmptyIter) Close() error {
	return nil
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (*EmptyIter) Err() error {
	return nil
}
