package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/collation"
)

// CacheStage simulates a source for queries that have been run and cached.
type CacheStage struct {
	cacheSize uint64
	rows      Rows
	columns   []*Column
	collation *collation.Collation
}

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

type CacheIter struct {
	cachedRows Rows
	rowNumber  uint64
	totalRows  uint64
	execCtx    *ExecutionCtx
	err        error
}

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

func (c *CacheStage) Columns() (columns []*Column) {
	return c.columns
}

func (c *CacheStage) Collation() *collation.Collation {
	return c.collation
}

func (*CacheIter) Close() error {
	return nil
}

func (ci *CacheIter) Err() error {
	return it.err
}
