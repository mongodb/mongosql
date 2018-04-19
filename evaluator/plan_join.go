package evaluator

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/variable"
)

// JoinStrategy is an enum that specifies the method a Join
// operator utilizes in performing a join operation.
type JoinStrategy byte

// These are the possible values for JoinStrategy.
const (
	NestedLoop JoinStrategy = iota
	SortMerge
	Hash
)

// NestedLoopJoiner is an implementation of a join.
type NestedLoopJoiner struct {
	ctx          *ExecutionCtx
	matcher      SQLExpr
	leftColumns  []*Column
	rightColumns []*Column
	kind         JoinKind
	collation    *collation.Collation
	errChan      chan error
}

// Joiner wraps the basic Join function that is
// used to combine data from two different sources.
type Joiner interface {
	Join(ctx context.Context, left, right <-chan *Row, execCtx *ExecutionCtx) <-chan Values
}

// JoinStage implements the operator interface for join expressions.
type JoinStage struct {
	left, right PlanStage
	matcher     SQLExpr
	kind        JoinKind
	strategy    JoinStrategy
}

// NewJoinStage returns a new JoinStage.
func NewJoinStage(kind JoinKind, left, right PlanStage, predicate SQLExpr) *JoinStage {
	return &JoinStage{
		kind:    kind,
		left:    left,
		right:   right,
		matcher: predicate,
	}
}

// JoinIter returns rows from a joined table.
type JoinIter struct {
	left, right Iter
	ctx         *ExecutionCtx
	onChan      <-chan Values
	errChan     chan error
	err         error
	cancelIter  context.CancelFunc
}

// Open returns an iterator that returns results from executing this plan stage
// with the given ExecutionContext.
func (join *JoinStage) Open(ctx *ExecutionCtx) (Iter, error) {
	cancelCtx, cancel := context.WithCancel(ctx.Context())

	iter := &JoinIter{
		ctx:        ctx,
		cancelIter: cancel,
		errChan:    make(chan error, 2),
	}

	leftRows := make(chan *Row)
	rightRows := make(chan *Row)

	util.PanicSafeGo(func() {
		iterator, err := join.left.Open(ctx)
		if err != nil {
			iter.errChan <- err
			return
		}
		iter.left = iterator
		iter.fetchRows(cancelCtx, iter.left, leftRows, iter.errChan)
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("%v", err)
	})

	util.PanicSafeGo(func() {
		iterator, err := join.right.Open(ctx)
		if err != nil {
			iter.errChan <- err
			return
		}
		iter.right = iterator
		iter.fetchRows(cancelCtx, iter.right, rightRows, iter.errChan)
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("%v", err)
	})

	joiner := NewJoiner(
		ctx,
		join.strategy,
		join.kind,
		join.Collation(),
		join.matcher,
		join.left.Columns(),
		join.right.Columns(),
		iter.errChan,
	)

	iter.onChan = joiner.Join(cancelCtx, leftRows, rightRows, ctx)

	return iter, nil
}

// fetchRows reads Row objects from a given Iter, and publishes them on channel
// ch, closing it when the iterator is exhausted. Errors encountered during
// iteration are published on errChan.
func (iter *JoinIter) fetchRows(ctx context.Context,
	it Iter,
	ch chan<- *Row,
	errChan chan<- error) {
	r := &Row{}

	syncChan := make(chan *Row)
	fetchErrChan := make(chan error, 1)

	util.PanicSafeGo(func() {
	iterLoop:
		for it.Next(r) {
			select {
			case syncChan <- r:
				r = &Row{}
			case <-ctx.Done():
				break iterLoop
			}
		}

		err := it.Close()
		// This err was previously ignored.
		if err != nil {
			panic(err)
		}
		close(syncChan)
	}, func(err interface{}) {
		fetchErrChan <- fmt.Errorf("join fetch error: %v", err)
	})

	for {
		select {
		case row, ok := <-syncChan:
			if !ok {
				if err := it.Err(); err != nil {
					errChan <- err
				}
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

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (iter *JoinIter) Next(row *Row) bool {
	var ok bool
	select {
	case err := <-iter.errChan:
		iter.err = err
		return false
	case row.Data, ok = <-iter.onChan:
		if !ok {
			return false
		}
	}

	return true
}

// Close closes the iterator, returning any error encountered while doing so.
func (iter *JoinIter) Close() error {
	iter.cancelIter()

	if err := iter.left.Close(); err != nil {
		// There is no way to combine errors.
		_ = iter.right.Close()
		return err
	}

	return iter.right.Close()
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (join *JoinStage) Columns() []*Column {
	left := join.left.Columns()
	right := join.right.Columns()
	columns := make([]*Column, len(left), len(left)+len(right))
	copy(columns, left)
	columns = append(columns, right...)
	return columns
}

// Collation returns the collation to use for comparisons.
func (join *JoinStage) Collation() *collation.Collation {
	return join.left.Collation()
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (iter *JoinIter) Err() error {

	if err := iter.left.Err(); err != nil {
		return err
	}

	if err := iter.right.Err(); err != nil {
		return err
	}

	return iter.err
}

// NewJoiner returns a new Joiner implementation for the given
// strategy. The implementation uses the supplied matcher in
// evaluating the join criteria and performs joins according
// to the joinType
func NewJoiner(ctx *ExecutionCtx,
	s JoinStrategy,
	kind JoinKind,
	collation *collation.Collation,
	matcher SQLExpr,
	leftColumns,
	rightColumns []*Column,
	errChan chan error) Joiner {

	switch s {
	case NestedLoop:
		return &NestedLoopJoiner{ctx, matcher, leftColumns, rightColumns, kind, collation, errChan}
	default:
		panic(fmt.Sprintf("unsupported join strategy: %v", s))
	}
}

func (nlp *NestedLoopJoiner) readData(ctx context.Context,
	lChan,
	rChan <-chan *Row) ([]*Row,
	[]*Row,
	error) {

	maxSize := nlp.ctx.Variables().GetUInt64(variable.MongoDBMaxStageSize)
	size := uint64(0)

	var left []*Row
	var right []*Row
	errs := make(chan error, 2)

	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	readChan := func(ch <-chan *Row, out *[]*Row) {
		for {
			select {
			case r, more := <-ch:
				if !more {
					errs <- nil
					return
				}

				*out = append(*out, r)
				newSize := atomic.AddUint64(&size, r.Data.Size())
				if maxSize != 0 && newSize > maxSize {
					errs <- newPlanStageMemoryError(maxSize)
					cancel()
					return
				}
			case <-cancelCtx.Done():
				errs <- cancelCtx.Err()
				return
			}
		}
	}

	util.PanicSafeGo(func() {
		readChan(lChan, &left)
	}, func(err interface{}) {
		cancel()
		errs <- fmt.Errorf("%v", err)
	})

	util.PanicSafeGo(func() {
		readChan(rChan, &right)
	}, func(err interface{}) {
		cancel()
		errs <- fmt.Errorf("%v", err)
	})

	// We need to block 2x, once for each side of the join. Either an error or nil
	// will get returned from each side. If we have an error, cancel the other side
	// and return the error. Otherwise, move on and wait for the other side to finish,
	// sending back it's error (or lack thereof) with the result.
	err := <-errs
	if err != nil {
		cancel()
		return nil, nil, err
	}

	err = <-errs

	return left, right, err
}

// Join is the join implementation for a NestedLoopJoiner.
func (nlp *NestedLoopJoiner) Join(ctx context.Context,
	lChan,
	rChan <-chan *Row,
	execCtx *ExecutionCtx) <-chan Values {

	getNilValues := func(columns []*Column) Values {
		var nilValues Values
		for _, c := range columns {
			nilValues = append(nilValues, NewValue(
				c.SelectID,
				c.Database,
				c.Table,
				c.Name,
				nil))
		}
		return nilValues
	}

	left, right, err := nlp.readData(ctx, lChan, rChan)
	if err != nil {
		nlp.errChan <- err
		return nil
	}

	ch := make(chan Values)

	switch nlp.kind {
	case CrossJoin:
		util.PanicSafeGo(func() {
			nlp.crossJoin(ctx, left, right, ch, execCtx)
		}, func(err interface{}) {
			nlp.errChan <- fmt.Errorf("%v", err)
		})
	case InnerJoin, StraightJoin:
		util.PanicSafeGo(func() {
			nlp.innerJoin(ctx, left, right, ch, execCtx)
		}, func(err interface{}) {
			nlp.errChan <- fmt.Errorf("%v", err)
		})
	case LeftJoin:
		util.PanicSafeGo(func() {
			nlp.leftJoin(ctx, left, right, ch, execCtx, getNilValues(nlp.rightColumns))
		}, func(err interface{}) {
			nlp.errChan <- fmt.Errorf("%v", err)
		})
	case RightJoin:
		util.PanicSafeGo(func() {
			nlp.rightJoin(ctx, left, right, ch, execCtx, getNilValues(nlp.leftColumns))
		}, func(err interface{}) {
			nlp.errChan <- fmt.Errorf("%v", err)
		})
	}

	return ch
}

func (nlp *NestedLoopJoiner) innerJoin(ctx context.Context,
	left,
	right []*Row,
	ch chan<- Values,
	execCtx *ExecutionCtx) {

outerLoop:
	for _, l := range left {
		for _, r := range right {
			evalCtx := NewEvalCtx(execCtx, nlp.collation, l, r)
			m, err := Matches(nlp.matcher, evalCtx)
			if err != nil {
				nlp.errChan <- err
				close(ch)
				return
			} else if m {
				values := make(Values, len(l.Data)+len(r.Data))
				copy(values, append(l.Data, r.Data...))
				select {
				case ch <- values:
				case <-ctx.Done():
					break outerLoop
				}
			}
		}
	}

	close(ch)
}

func (nlp *NestedLoopJoiner) leftJoin(ctx context.Context,
	left,
	right []*Row,
	ch chan<- Values,
	execCtx *ExecutionCtx,
	nilRightValues Values) {

	var hasMatch bool

outerLoop:
	for _, l := range left {
		for _, r := range right {
			evalCtx := NewEvalCtx(execCtx, nlp.collation, l, r)
			m, err := Matches(nlp.matcher, evalCtx)
			if err != nil {
				nlp.errChan <- err
				close(ch)
				return
			} else if m {
				hasMatch = true
				values := make(Values, len(l.Data)+len(r.Data))
				copy(values, append(l.Data, r.Data...))
				select {
				case ch <- values:
				case <-ctx.Done():
					break outerLoop
				}
			}
		}

		if !hasMatch {
			values := make(Values, len(nilRightValues)+len(l.Data))
			copy(values, append(l.Data, nilRightValues...))
			select {
			case ch <- values:
			case <-ctx.Done():
				break outerLoop
			}
		}

		hasMatch = false
	}

	close(ch)
}

func (nlp *NestedLoopJoiner) rightJoin(ctx context.Context,
	left,
	right []*Row,
	ch chan<- Values,
	execCtx *ExecutionCtx,
	nilLeftValues Values) {

	var hasMatch bool

outerLoop:
	for _, r := range right {
		for _, l := range left {
			evalCtx := NewEvalCtx(execCtx, nlp.collation, l, r)
			m, err := Matches(nlp.matcher, evalCtx)
			if err != nil {
				nlp.errChan <- err
				close(ch)
				return
			} else if m {
				hasMatch = true
				values := make(Values, len(l.Data)+len(r.Data))
				copy(values, append(l.Data, r.Data...))
				select {
				case ch <- values:
				case <-ctx.Done():
					break outerLoop
				}
			}
		}

		if !hasMatch {
			values := make(Values, len(nilLeftValues)+len(r.Data))
			copy(values, append(nilLeftValues, r.Data...))
			select {
			case ch <- values:
			case <-ctx.Done():
				break outerLoop
			}
		}

		hasMatch = false
	}

	close(ch)
}

func (nlp *NestedLoopJoiner) crossJoin(ctx context.Context,
	left,
	right []*Row,
	ch chan<- Values,
	_ *ExecutionCtx) {

outerLoop:
	for _, l := range left {
		for _, r := range right {
			values := make(Values, len(l.Data)+len(r.Data))
			copy(values, append(l.Data, r.Data...))
			select {
			case ch <- values:
			case <-ctx.Done():
				break outerLoop
			}
		}
	}

	close(ch)
}
