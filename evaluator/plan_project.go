package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/memory"
	"github.com/10gen/sqlproxy/evaluator/results"
)

// ProjectStage handles taking sourced rows and projecting them into a different shape.
type ProjectStage struct {
	// projectedColumns holds the columns and/or expressions used in
	// the pipeline
	projectedColumns ProjectedColumns

	// source is the operator that provides the data to project
	source PlanStage
}

// Children returns a slice of all the Node children of the Node.
func (ps ProjectStage) Children() []Node {
	out := make([]Node, len(ps.projectedColumns)+1)
	for i := range ps.projectedColumns {
		out[i] = ps.projectedColumns[i].Expr
	}
	out[len(ps.projectedColumns)] = ps.source
	return out
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (ps *ProjectStage) ReplaceChild(i int, n Node) {
	if i == len(ps.projectedColumns) {
		ps.source = n.(PlanStage)
		return
	}
	if 0 <= i && i < len(ps.projectedColumns) {
		ps.projectedColumns[i].Expr = panicIfNotSQLExpr("ProjectStage", n)
		return
	}
	panicWithInvalidIndex("ProjectStage", i, len(ps.projectedColumns))
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
	memoryMonitor    memory.Monitor
	source           results.RowIter
	projectedColumns ProjectedColumns
	// err holds any error that may have occurred during processing
	err error
}

// Open returns an iterator over this PlanStage's returned rows.
func (ps *ProjectStage) Open(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (results.RowIter, error) {
	sourceIter, err := ps.source.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	return &ProjectIter{
		cfg:              cfg,
		st:               st.WithCollation(ps.Collation()),
		memoryMonitor:    cfg.memoryMonitor,
		projectedColumns: ps.projectedColumns,
		source:           sourceIter,
	}, nil
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (pj *ProjectIter) Next(ctx context.Context, r *results.Row) bool {
	if !pj.source.Next(ctx, r) {
		return false
	}

	vs := results.RowValues{}
	st := pj.st.WithRows(r)
	for _, projectedColumn := range pj.projectedColumns {
		v, err := projectedColumn.Expr.Evaluate(ctx, pj.cfg, st)
		if err != nil {
			pj.err = err
			return false
		}

		value := results.NewRowValue(
			projectedColumn.SelectID,
			projectedColumn.Database,
			projectedColumn.Table,
			projectedColumn.Name,
			v)

		vs = append(vs, value)
	}

	pj.err = pj.memoryMonitor.Release(r.Data.Size())
	if pj.err != nil {
		return false
	}
	r.Data = vs
	pj.err = pj.memoryMonitor.Acquire(r.Data.Size())
	return pj.err == nil
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (ps *ProjectStage) Columns() (columns []*results.Column) {
	for _, projectedColumn := range ps.projectedColumns {
		columns = append(columns, projectedColumn.Column.Clone())
	}

	if len(columns) == 0 {
		columns = ps.source.Columns()
	}

	return columns
}

// Collation returns the collation to use for comparisons.
func (ps *ProjectStage) Collation() *collation.Collation {
	return ps.source.Collation()
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

func (ps *ProjectStage) clone() PlanStage {
	return &ProjectStage{
		source:           ps.source.clone(),
		projectedColumns: ps.projectedColumns,
	}
}

// ProjectedColumns returns the projectedColumns.
func (ps *ProjectStage) ProjectedColumns() ProjectedColumns {
	return ps.projectedColumns
}
