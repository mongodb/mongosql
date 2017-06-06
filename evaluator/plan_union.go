package evaluator

import (
	"fmt"
	"sort"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/schema"
)

type UnionKind int

const (
	UnionDistinct UnionKind = iota
	UnionAll
)

type UnionStage struct {
	left, right PlanStage
	kind        UnionKind
}

func NewUnionStage(kind UnionKind, left, right PlanStage) *UnionStage {
	return &UnionStage{
		kind:  kind,
		left:  left,
		right: right,
	}
}

type UnionIter struct {
	left, right Iter
	ctx         *ExecutionCtx
	columns     []*Column
	onChan      chan Row
	errChan     chan error
	err         error
}

func (union *UnionStage) Open(ctx *ExecutionCtx) (Iter, error) {
	left, err := union.left.Open(ctx)
	if err != nil {
		return nil, err
	}

	right, err := union.right.Open(ctx)
	if err != nil {
		return nil, err
	}

	iter := &UnionIter{
		left:    left,
		right:   right,
		ctx:     ctx,
		columns: union.Columns(),

		errChan: make(chan error, 1),
	}

	leftRows := make(chan *Row)
	rightRows := make(chan *Row)

	util.PanicSafeGo(func() {
		iter.fetchRows(left, leftRows, iter.errChan)
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("%v", err)
	})

	util.PanicSafeGo(func() {
		iter.fetchRows(right, rightRows, iter.errChan)
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("%v", err)
	})

	iter.onChan = iter.unify(leftRows, rightRows)

	return iter, nil
}

func (iter *UnionIter) fetchRows(it Iter, ch chan *Row, errChan chan error) {
	r := &Row{}

	syncChan := make(chan *Row)
	fetchErrChan := make(chan error, 1)

	util.PanicSafeGo(func() {
		for it.Next(r) {
			// Need to match row info with parent
			for i, col := range iter.columns {
				r.Data[i].Name = col.Name
				r.Data[i].Data, _ = NewSQLValue(r.Data[i].Data, col.SQLType, schema.SQLNone)
			}

			syncChan <- r
			r = &Row{}
		}

		if err := it.Err(); err != nil {
			errChan <- err
		}

		it.Close()
		close(syncChan)
	}, func(err interface{}) {
		fetchErrChan <- fmt.Errorf("union fetch error: %v", err)
	})

	for {
		select {
		case row, ok := <-syncChan:
			if !ok {
				close(ch)
				return
			}

			ch <- row
		case <-iter.ctx.Context().Done():
			errChan <- iter.ctx.Context().Err()
			return
		case err := <-fetchErrChan:
			errChan <- err
			return
		}
	}
}

func mergeColumnsByType(lcols, rcols []*Column) []*Column {
	outCols := make([]*Column, len(lcols))

	sorter := &schema.SQLTypesSorter{}
	for i, lcol := range lcols {
		rcol := rcols[i]
		sorter.Types = []schema.SQLType{lcol.SQLType, rcol.SQLType}
		sort.Sort(sorter)

		outCol := lcol.clone()
		outCol.SQLType = sorter.Types[1] // Use "gte" type
		outCols[i] = outCol
	}

	return lcols
}

func (union *UnionStage) Columns() []*Column {
	return mergeColumnsByType(union.left.Columns(), union.right.Columns())
}

func (union *UnionStage) Collation() *collation.Collation {
	return union.left.Collation()
}

func (iter *UnionIter) Next(row *Row) bool {
	select {
	case err := <-iter.errChan:
		iter.err = err
		return false
	case data, ok := <-iter.onChan:
		row.Data = data.Data
		if !ok {
			return false
		}
	}
	return true
}

func (iter *UnionIter) Close() error {

	if err := iter.left.Close(); err != nil {
		return err
	}

	if err := iter.right.Close(); err != nil {
		return err
	}

	return nil
}

func (iter *UnionIter) Err() error {

	if err := iter.left.Err(); err != nil {
		return err
	}

	if err := iter.right.Err(); err != nil {
		return err
	}

	return iter.err
}

func (iter *UnionIter) unify(lChan, rChan chan *Row) chan Row {

	ch := make(chan Row)
	closeChan := make(chan struct{})

	// cleanup
	util.PanicSafeGo(func() {
		_, _ = <-closeChan, <-closeChan
		close(closeChan)
		close(ch)
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("%v", err)
	})

	// retrieve rows from left and right stages in parallel
	util.PanicSafeGo(func() {
		for l := range lChan {
			ch <- *l
		}
		closeChan <- struct{}{}
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("left unify error: %v", err)
	})

	util.PanicSafeGo(func() {
		for r := range rChan {
			ch <- *r
		}
		closeChan <- struct{}{}
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("right unify error: %v", err)
	})

	return ch
}
