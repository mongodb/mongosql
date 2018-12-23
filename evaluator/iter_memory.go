package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/evaluator/memory"
)

type memoryIter struct {
	monitor memory.Monitor
	source  Iter

	err error
}

func newMemoryIter(cfg *ExecutionConfig, src Iter) *memoryIter {
	return &memoryIter{
		monitor: cfg.memoryMonitor,
		source:  src,
	}
}

func (i *memoryIter) Next(ctx context.Context, row *Row) bool {
	hasNext := i.source.Next(ctx, row)
	if hasNext {
		i.err = i.monitor.Release(row.Data.Size())
		if i.err != nil {
			return false
		}
	}
	return hasNext
}

func (i *memoryIter) Close() error {
	err := i.source.Close()
	if err != nil {
		return err
	}
	return i.monitor.ReleaseGlobal()
}

func (i *memoryIter) Err() error {
	if err := i.source.Err(); err != nil {
		return err
	}

	return i.err
}
