package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/variable"
)

// CacheStage simulates a source for queries that have been run and cached.
type CacheStage struct {
	cacheSize uint64
	rows      Rows
	columns   []*Column
	collation *collation.Collation
}

// NewCacheStage returns a new CacheStage.
func NewCacheStage(cacheSize uint64, rows Rows, columns []*Column, collation *collation.Collation) *CacheStage {
	return &CacheStage{cacheSize, rows, columns, collation}
}

func (c *CacheStage) clone() *CacheStage {

	return &CacheStage{
		cacheSize: c.cacheSize,
		rows:      c.rows,
		columns:   c.columns,
		collation: c.collation,
	}
}

// CacheIter returns cached rows.
type CacheIter struct {
	cachedRows Rows
	rowNumber  uint64
	totalRows  uint64
	execCtx    *ExecutionCtx
	err        error
}

// Open returns an iterator that returns results from executing this plan stage
// with the given ExecutionContext.
func (c *CacheStage) Open(ctx *ExecutionCtx) (Iter, error) {
	if c.rows == nil {
		return nil, fmt.Errorf("No query in plan cache")
	}

	if ctx.Context() == nil {
		return nil, fmt.Errorf("No connection context provided in the execution context")
	}
	return &CacheIter{
		cachedRows: c.rows,
		execCtx:    ctx,
		totalRows:  uint64(len(c.rows)),
	}, nil
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (ci *CacheIter) Next(row *Row) bool {

	ctx := ci.execCtx.Context()
	if err := ctx.Err(); err != nil {
		ci.err = err
		return false
	}

	if ci.rowNumber >= ci.totalRows {
		return false
	}
	row.Data = ci.cachedRows[ci.rowNumber].Data
	ci.rowNumber++
	return true
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (c *CacheStage) Columns() (columns []*Column) {
	return c.columns
}

// Collation returns the collation to use for comparisons.
func (c *CacheStage) Collation() *collation.Collation {
	return c.collation
}

// Close closes the iterator, returning any error encountered while doing so.
func (*CacheIter) Close() error {
	return nil
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (ci *CacheIter) Err() error {
	return ci.err
}

// cachePlanStage executes a PlanStage within the evalCtx and returns the cached results.
func cachePlanStage(ps PlanStage, evalCtx *EvalCtx) (*CacheStage, error) {
	var iter Iter
	var err error
	execCtx := evalCtx.ExecutionCtx
	if iter, err = ps.Open(execCtx); err != nil {
		return nil, err
	}
	// maxCacheSizeBytes is the maximimum size a single cached query can be in bytes
	// It is set to be equal to the max plan stage size
	maxCacheSizeBytes := execCtx.ConnectionCtx.Variables().GetUInt64(variable.MongoDBMaxStageSize)

	size := uint64(0)
	row, allRows := &Row{}, Rows{}
	for iter.Next(row) {
		if maxCacheSizeBytes != 0 && size > maxCacheSizeBytes {
			return nil, newPlanStageMemoryError(maxCacheSizeBytes)
		}
		allRows = append(allRows, *row)
		size += row.Data.Size()
		row = &Row{}
	}

	if err = iter.Close(); err != nil {
		return nil, err
	}
	if err = iter.Err(); err != nil {
		return nil, err
	}

	return NewCacheStage(size, allRows, ps.Columns(), ps.Collation()), nil
}
