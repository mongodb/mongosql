package evaluator

import (
	"sort"

	"github.com/10gen/sqlproxy/collation"
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

	go iter.fetchRows(left, leftRows, iter.errChan)
	go iter.fetchRows(right, rightRows, iter.errChan)

	iter.onChan = unify(leftRows, rightRows)

	return iter, nil
}

func (iter *UnionIter) fetchRows(it Iter, ch chan *Row, errChan chan error) {
	r := &Row{}

	syncChan := make(chan *Row)

	go func() {
		for it.Next(r) {
			// Need to match row info with parent
			for i, col := range iter.columns {
				r.Data[i].Name = col.Name
				r.Data[i].Data = NewSQLValue(r.Data[i].Data, col.SQLType)
			}

			syncChan <- r
			r = &Row{}
		}

		if err := it.Err(); err != nil {
			errChan <- err
		}

		it.Close()
		close(syncChan)
	}()

	for {
		select {
		case row, ok := <-syncChan:
			if !ok {
				close(ch)
				return
			}

			ch <- row
		case <-iter.ctx.Tomb().Dying():
			errChan <- iter.ctx.Tomb().Err()
			return
		}
	}
}

func mergeColumnsByType(lcols, rcols []*Column) []*Column {
	outCols := make([]*Column, len(lcols))

	for i, lcol := range lcols {
		rcol := rcols[i]
		sqlTypes := schema.SQLTypes{lcol.SQLType, rcol.SQLType}
		sort.Sort(sqlTypes)

		outCol := lcol.clone()
		outCol.SQLType = sqlTypes[1] // Use "gte" type
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

func unify(lChan, rChan chan *Row) chan Row {

	ch := make(chan Row)
	closeChan := make(chan struct{})

	// cleanup
	go func() {
		_, _ = <-closeChan, <-closeChan
		close(closeChan)
		close(ch)
	}()

	// retrieve rows from left and right stages in parallel
	go func() {
		for l := range lChan {
			ch <- *l
		}
		closeChan <- struct{}{}
	}()

	go func() {
		for r := range rChan {
			ch <- *r
		}
		closeChan <- struct{}{}
	}()

	return ch
}
