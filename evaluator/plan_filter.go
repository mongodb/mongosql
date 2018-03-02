package evaluator

import (
	"github.com/10gen/sqlproxy/collation"
)

// A FilterStage ensures that only rows matching a given criteria are
// returned.
type FilterStage struct {
	matcher SQLExpr
	source  PlanStage
}

// NewFilterStage returns a new FilterStage.
func NewFilterStage(source PlanStage, predicate SQLExpr) *FilterStage {
	return &FilterStage{
		source:  source,
		matcher: predicate,
	}
}

// FilterIter returns only the rows that match the filter expression.
type FilterIter struct {
	matcher   SQLExpr
	execCtx   *ExecutionCtx
	source    Iter
	collation *collation.Collation
	err       error
}

// Open returns an iterator that returns results from executing this plan stage
// with the given ExecutionContext.
func (fs *FilterStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := fs.source.Open(ctx)
	if err != nil {
		return nil, err
	}
	return &FilterIter{
		matcher:   fs.matcher,
		execCtx:   ctx,
		source:    sourceIter,
		err:       nil,
		collation: fs.Collation(),
	}, nil
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (fs *FilterStage) Columns() (columns []*Column) {
	return fs.source.Columns()
}

// Collation returns the collation to use for comparisons.
func (fs *FilterStage) Collation() *collation.Collation {
	return fs.source.Collation()
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (fi *FilterIter) Next(row *Row) bool {
	var hasMatch, hasNext bool

	for {

		hasNext = fi.source.Next(row)

		if !hasNext {
			break
		}

		if fi.matcher == nil {
			break
		}

		evalCtx := NewEvalCtx(fi.execCtx, fi.collation, row)

		hasMatch, fi.err = Matches(fi.matcher, evalCtx)
		if fi.err != nil {
			return false
		}

		if hasMatch {
			break
		}

		row.Data = nil
	}

	if fi.matcher != nil && !hasMatch {
		row.Data = nil
	}

	return hasNext
}

// Close closes the iterator, returning any error encountered while doing so.
func (fi *FilterIter) Close() error {
	return fi.source.Close()
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (fi *FilterIter) Err() error {
	if fi.err != nil {
		return fi.err
	}
	return fi.source.Err()
}
