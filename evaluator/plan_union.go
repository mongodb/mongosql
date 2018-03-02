package evaluator

import (
	"context"
	"fmt"
	"sort"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/schema"
)

// UnionKind is an enum representing the different kinds of unions.
type UnionKind int

// These are the possible values for UnionKind.
const (
	UnionDistinct UnionKind = iota
	UnionAll
)

// UnionStage handles combining two result sets.
type UnionStage struct {
	left, right PlanStage
	kind        UnionKind
}

// NewUnionStage creates a new UnionStage.
func NewUnionStage(kind UnionKind, left, right PlanStage) *UnionStage {
	return &UnionStage{
		kind:  kind,
		left:  left,
		right: right,
	}
}

// UnionIter returns rows from the union of two source iterators.
type UnionIter struct {
	left, right Iter
	ctx         *ExecutionCtx
	columns     []*Column
	onChan      chan Row
	errChan     chan error
	err         error
	cancelIter  context.CancelFunc
}

// Open returns an iterator that returns results from executing this plan stage
// with the given ExecutionContext.
func (union *UnionStage) Open(ctx *ExecutionCtx) (Iter, error) {
	left, err := union.left.Open(ctx)
	if err != nil {
		return nil, err
	}

	right, err := union.right.Open(ctx)
	if err != nil {
		return nil, err
	}

	cancelCtx, cancel := context.WithCancel(ctx.Context())

	iter := &UnionIter{
		left:       left,
		right:      right,
		ctx:        ctx,
		columns:    union.Columns(),
		errChan:    make(chan error, 1),
		cancelIter: cancel,
	}

	leftRows := make(chan *Row)
	rightRows := make(chan *Row)

	util.PanicSafeGo(func() {
		iter.fetchRows(cancelCtx, left, leftRows, iter.errChan)
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("%v", err)
	})

	util.PanicSafeGo(func() {
		iter.fetchRows(cancelCtx, right, rightRows, iter.errChan)
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("%v", err)
	})

	iter.onChan = iter.unify(cancelCtx, leftRows, rightRows)

	return iter, nil
}

func (iter *UnionIter) fetchRows(ctx context.Context, it Iter, ch chan *Row, errChan chan error) {
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

			select {
			case syncChan <- r:
				r = &Row{}
			case <-ctx.Done():
			}
		}

		if err := it.Err(); err != nil {
			errChan <- err
		}

		// This err was previously ignored.
		if err := it.Close(); err != nil {
			panic(err)
		}

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
		case <-ctx.Done():
			errChan <- ctx.Err()
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

// Columns returns the ordered set of columns that are contained in results from this plan.
func (union *UnionStage) Columns() []*Column {
	return mergeColumnsByType(union.left.Columns(), union.right.Columns())
}

// Collation returns the collation to use for comparisons.
func (union *UnionStage) Collation() *collation.Collation {
	return union.left.Collation()
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
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
		// past this stage, all columns must
		// present the same table name.
		for i := 0; i < len(row.Data); i++ {
			row.Data[i].Table = iter.columns[i].Table
		}
	}
	return true
}

// Close closes the iterator, returning any error encountered while doing so.
func (iter *UnionIter) Close() error {
	iter.cancelIter()

	if err := iter.left.Close(); err != nil {
		return err
	}

	return iter.right.Close()
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (iter *UnionIter) Err() error {

	if err := iter.left.Err(); err != nil {
		return err
	}

	if err := iter.right.Err(); err != nil {
		return err
	}

	return iter.err
}

func (iter *UnionIter) unify(ctx context.Context, lChan, rChan chan *Row) chan Row {

	ch := make(chan Row)
	closeChan := make(chan struct{})

	// cleanup
	util.PanicSafeGo(func() {
		<-closeChan
		<-closeChan
		close(closeChan)
		close(ch)
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("%v", err)
	})

	// retrieve rows from left and right stages in parallel
	util.PanicSafeGo(func() {
	chanLoop:
		for l := range lChan {
			select {
			case ch <- *l:
			case <-ctx.Done():
				break chanLoop
			}
		}
		closeChan <- struct{}{}
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("left unify error: %v", err)
	})

	util.PanicSafeGo(func() {
	chanLoop:
		for r := range rChan {
			select {
			case ch <- *r:
			case <-ctx.Done():
				break chanLoop
			}
		}
		closeChan <- struct{}{}
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("right unify error: %v", err)
	})

	return ch
}
