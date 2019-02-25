package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/memory"
	"github.com/10gen/sqlproxy/evaluator/results"
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
	memoryMonitor        memory.Monitor
	limit, offset, total uint64

	source RowIter
	err    error
}

// Children returns a slice of all the Node children of the Node.
func (l LimitStage) Children() []Node {
	return []Node{l.source}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (l *LimitStage) ReplaceChild(i int, e Node) {
	switch i {
	case 0:
		l.source = e.(PlanStage)
	default:
		panicWithInvalidIndex("LimitStage", i, 0)
	}
}

// Open returns an iterator that returns results from executing this plan stage
// with the given ExecutionContext.
func (l *LimitStage) Open(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (RowIter, error) {
	sourceIter, err := l.source.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}
	return &LimitIter{
		memoryMonitor: cfg.memoryMonitor,
		limit:         l.limit,
		offset:        l.offset,
		source:        sourceIter,
	}, nil
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (l *LimitIter) Next(ctx context.Context, row *results.Row) bool {
	if l.offset != 0 {
		r := &results.Row{}
		for l.source.Next(ctx, r) {
			l.err = l.memoryMonitor.Release(r.Data.Size())
			if l.err != nil {
				return false
			}
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

	return l.source.Next(ctx, row)
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (l *LimitStage) Columns() (columns []*results.Column) {
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
	if err := l.source.Err(); err != nil {
		return err
	}
	return l.err
}

func (l *LimitStage) clone() PlanStage {
	return NewLimitStage(l.source.clone(), l.offset, l.limit)
}
