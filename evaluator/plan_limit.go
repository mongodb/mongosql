package evaluator

import (
	"github.com/10gen/sqlproxy/collation"
)

// A LimitStage restricts the number of rows returned by a query.
type LimitStage struct {
	// limit is the maximum number of rows to return
	limit uint64

	// offset keeps a 0-indexed offset of where to start returning rows
	offset uint64

	source PlanStage
}

// NewLimitStage returns a new LimitStage.
func NewLimitStage(source PlanStage, offset uint64, limit uint64) *LimitStage {
	return &LimitStage{
		source: source,
		offset: offset,
		limit:  limit,
	}
}

// A LimitIter returns no more than a given number of rows.
type LimitIter struct {
	limit, offset, total uint64

	source Iter
}

// Open returns an iterator that returns results from executing this plan stage
// with the given ExecutionContext.
func (l *LimitStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := l.source.Open(ctx)
	if err != nil {
		return nil, err
	}
	return &LimitIter{
		limit:  l.limit,
		offset: l.offset,
		source: sourceIter,
	}, nil
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (l *LimitIter) Next(row *Row) bool {

	if l.offset != 0 {
		r := &Row{}
		for l.source.Next(r) {
			l.total++
			if l.total == l.offset {
				break
			}

		}

		if l.total < l.offset {
			return false
		}

		l.offset, l.total = 0, 0
	}

	l.total++

	if l.total > l.limit {
		return false
	}

	return l.source.Next(row)
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (l *LimitStage) Columns() (columns []*Column) {
	return l.source.Columns()
}

// Collation returns the collation to use for comparisons.
func (l *LimitStage) Collation() *collation.Collation {
	return l.source.Collation()
}

// Close closes the iterator, returning any error encountered while doing so.
func (l *LimitIter) Close() error {
	return l.source.Close()
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (l *LimitIter) Err() error {
	return l.source.Err()
}
