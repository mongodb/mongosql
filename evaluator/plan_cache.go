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

type CacheIter struct {
	cachedRows Rows
	rowNumber  uint64
	totalRows  uint64
}

func (c *CacheStage) Open(ctx *ExecutionCtx) (Iter, error) {
	if c.rows == nil {
		return nil, fmt.Errorf("No query in plan cache")
	}
	return &CacheIter{c.rows, 0, uint64(len(c.rows))}, nil
}

func (ci *CacheIter) Next(row *Row) bool {
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

func (_ *CacheIter) Close() error {
	return nil
}

func (_ *CacheIter) Err() error {
	return nil
}
