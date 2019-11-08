package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/results"
)

// PlanStage represents a single a node in the Plan tree.
type PlanStage interface {
	Node

	// Open returns an iterator that returns results from executing this plan stage with the given
	// ExecutionContext.
	Open(context.Context, *ExecutionConfig, *ExecutionState) (results.RowIter, error)

	// Columns returns the ordered set of columns that are contained in results from this plan.
	Columns() []*results.Column

	// Collation returns the collation to use for comparisons.
	Collation() *collation.Collation

	clone() PlanStage
}

// FastPlanStage is a PlanStage that has a FastOpen method.
type FastPlanStage interface {
	PlanStage

	// FastOpen returns a FastIter that streams bson.RawD documents.
	FastOpen(context.Context, *ExecutionConfig, *ExecutionState) (results.DocIter, error)
}
