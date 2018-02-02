package evaluator

import "github.com/10gen/sqlproxy/internal/memory"

type memoryIter struct {
	ctx    *ExecutionCtx
	plan   PlanStage
	source Iter

	err error
}

func (i *memoryIter) Next(row *Row) bool {
	hasNext := i.source.Next(row)
	if hasNext {
		i.err = i.ctx.MemoryMonitor().Release(row.Data.Size())
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

	cacheFreer := cacheStageMemoryFreer{memMonitor: i.ctx.MemoryMonitor()}
	_, err = cacheFreer.visit(i.plan)
	return err
}

func (i *memoryIter) Err() error {
	if err := i.source.Err(); err != nil {
		return err
	}

	return i.err
}

type cacheStageMemoryFreer struct {
	memMonitor *memory.Monitor
}

func (v *cacheStageMemoryFreer) visit(n Node) (Node, error) {

	switch typedN := n.(type) {
	case *CacheStage:
		size := typedN.cacheSize
		typedN.cacheSize = 0
		typedN.rows = Rows{}
		return n, v.memMonitor.Release(size)
	}

	return walk(v, n)
}
