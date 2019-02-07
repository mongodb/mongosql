package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/evaluator/memory"
)

type MemoryIter struct {
	monitor memory.Monitor
	source  RowIter

	err error
}

// NewMemoryIter creates a MemoryIter from a RowIter.
func NewMemoryIter(cfg *ExecutionConfig, src RowIter) *MemoryIter {
	return &MemoryIter{
		monitor: cfg.memoryMonitor,
		source:  src,
	}
}

// Next gets the next row.
func (i *MemoryIter) Next(ctx context.Context, row *Row) bool {
	hasNext := i.source.Next(ctx, row)
	if hasNext {
		i.err = i.monitor.Release(row.Data.Size())
		if i.err != nil {
			return false
		}
	}
	return hasNext
}

// Close closes the iter.
func (i *MemoryIter) Close() error {
	err := i.source.Close()
	if err != nil {
		return err
	}
	return i.monitor.ReleaseGlobal()
}

// Err gets the error for an iter.
func (i *MemoryIter) Err() error {
	if err := i.source.Err(); err != nil {
		return err
	}

	return i.err
}
