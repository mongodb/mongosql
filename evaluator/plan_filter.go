package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/values"
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

// Children returns a slice of all the Node children of the Node.
func (fs FilterStage) Children() []Node {
	return []Node{fs.matcher, fs.source}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (fs *FilterStage) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		fs.matcher = panicIfNotSQLExpr("FilterStage", n)
	case 1:
		fs.source = panicIfNotPlanStage("FilterStage", n)
	default:
		panicWithInvalidIndex("FilterStage", i, 1)
	}
}

// FilterIter returns only the rows that match the filter expression.
type FilterIter struct {
	cfg     *ExecutionConfig
	st      *ExecutionState
	matcher SQLExpr
	source  RowIter
	err     error
}

// Open returns an iterator that returns results from executing this plan stage
// with the given ExecutionContext.
func (fs *FilterStage) Open(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (RowIter, error) {
	sourceIter, err := fs.source.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}
	return &FilterIter{
		cfg:     cfg,
		st:      st.WithCollation(fs.Collation()),
		matcher: fs.matcher,
		source:  sourceIter,
		err:     nil,
	}, nil
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (fs *FilterStage) Columns() (columns []*results.Column) {
	return fs.source.Columns()
}

// Collation returns the collation to use for comparisons.
func (fs *FilterStage) Collation() *collation.Collation {
	return fs.source.Collation()
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (fi *FilterIter) Next(ctx context.Context, row *results.Row) bool {
	var hasMatch, hasNext bool
	var result values.SQLValue

	for {

		hasNext = fi.source.Next(ctx, row)

		if !hasNext {
			break
		}

		if fi.matcher == nil {
			break
		}

		st := fi.st.WithRows(row)
		result, fi.err = fi.matcher.Evaluate(ctx, fi.cfg, st)
		if fi.err != nil {
			return false
		}

		hasMatch = values.Bool(result)
		if hasMatch {
			break
		}

		fi.err = fi.cfg.memoryMonitor.Release(row.Data.Size())
		if fi.err != nil {
			return false
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

func (fs *FilterStage) clone() PlanStage {
	return NewFilterStage(fs.source.clone(), fs.matcher)
}
