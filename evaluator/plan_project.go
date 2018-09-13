package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/memory"
)

// ProjectStage handles taking sourced rows and projecting them into a different shape.
type ProjectStage struct {
	// projectedColumns holds the columns and/or expressions used in
	// the pipeline
	projectedColumns ProjectedColumns

	// source is the operator that provides the data to project
	source PlanStage
}

// NewProjectStage creates a new project stage.
func NewProjectStage(source PlanStage, projectedColumns ...ProjectedColumn) *ProjectStage {
	return &ProjectStage{
		source:           source,
		projectedColumns: projectedColumns,
	}
}

// ProjectIter returns rows with specific columns projected.
type ProjectIter struct {
	cfg              *ExecutionConfig
	st               *ExecutionState
	memoryMonitor    *memory.Monitor
	source           Iter
	projectedColumns ProjectedColumns
	// err holds any error that may have occurred during processing
	err error
}

// Open returns an iterator over this PlanStage's returned rows.
func (pj *ProjectStage) Open(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (Iter, error) {
	sourceIter, err := pj.source.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	return &ProjectIter{
		cfg:              cfg,
		st:               st.WithCollation(pj.Collation()),
		memoryMonitor:    cfg.memoryMonitor,
		projectedColumns: pj.projectedColumns,
		source:           sourceIter,
	}, nil
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (pj *ProjectIter) Next(ctx context.Context, r *Row) bool {
	if !pj.source.Next(ctx, r) {
		return false
	}

	values := Values{}
	st := pj.st.WithRows(r)
	for _, projectedColumn := range pj.projectedColumns {
		v, err := projectedColumn.Expr.Evaluate(ctx, pj.cfg, st)
		if err != nil {
			pj.err = err
			return false
		}

		value := NewValue(
			projectedColumn.SelectID,
			projectedColumn.Database,
			projectedColumn.Table,
			projectedColumn.Name,
			v)

		values = append(values, value)
	}

	pj.err = pj.memoryMonitor.Release(r.Data.Size())
	if pj.err != nil {
		return false
	}
	r.Data = values
	pj.err = pj.memoryMonitor.Acquire(r.Data.Size())
	return pj.err == nil
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (pj *ProjectStage) Columns() (columns []*Column) {
	for _, projectedColumn := range pj.projectedColumns {
		columns = append(columns, projectedColumn.Column.clone())
	}

	if len(columns) == 0 {
		columns = pj.source.Columns()
	}

	return columns
}

// Collation returns the collation to use for comparisons.
func (pj *ProjectStage) Collation() *collation.Collation {
	return pj.source.Collation()
}

// Close closes the iterator, returning any error encountered while doing so.
func (pj *ProjectIter) Close() error {
	return pj.source.Close()
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (pj *ProjectIter) Err() error {
	if err := pj.source.Err(); err != nil {
		return err
	}
	return pj.err
}

func (pj *ProjectStage) clone() PlanStage {
	return &ProjectStage{
		source:           pj.source.clone(),
		projectedColumns: pj.projectedColumns,
	}
}

// ProjectedColumns returns the projectedColumns.
func (pj *ProjectStage) ProjectedColumns() ProjectedColumns {
	return pj.projectedColumns
}
