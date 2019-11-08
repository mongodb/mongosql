package evaluator

import (
	"context"
	"fmt"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/memory"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/procutil"
)

// NestedLoopJoiner is an implementation of a join.
type NestedLoopJoiner struct {
	cfg          *ExecutionConfig
	st           *ExecutionState
	stageMonitor memory.Monitor
	matcher      SQLExpr
	leftColumns  []*results.Column
	rightColumns []*results.Column
	kind         JoinKind
	errChan      chan error
}

// Joiner wraps the basic Join function that is
// used to combine data from two different sources.
type Joiner interface {
	Join(ctx context.Context, left, right <-chan *results.Row) <-chan results.RowValues
}

// JoinStage implements the operator interface for join expressions.
type JoinStage struct {
	left, right PlanStage
	matcher     SQLExpr
	kind        JoinKind
}

// Children returns a slice of all the Node children of the Node.
func (join JoinStage) Children() []Node {
	return []Node{join.left, join.right, join.matcher}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (join *JoinStage) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		join.left = panicIfNotPlanStage("JoinStage", n)
	case 1:
		join.right = panicIfNotPlanStage("JoinStage", n)
	case 2:
		join.matcher = panicIfNotSQLExpr("JoinStage", n)
	default:
		panicWithInvalidIndex("JoinStage", i, 2)
	}
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
	cfg          *ExecutionConfig
	st           *ExecutionState
	stageMonitor memory.Monitor
	left, right  results.RowIter
	onChan       <-chan results.RowValues
	errChan      chan error
	err          error
	cancelIter   context.CancelFunc
}

// Open returns an iterator that returns results from executing this plan stage
// with the given ExecutionContext.
func (join *JoinStage) Open(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (results.RowIter, error) {
	stageMonitor, err := cfg.memoryMonitor.CreateChild("JoinStage", cfg.maxStageSize)
	if err != nil {
		return nil, err
	}

	cancelCtx, cancel := context.WithCancel(ctx)

	iter := &JoinIter{
		cfg:          cfg,
		st:           st,
		stageMonitor: stageMonitor,
		cancelIter:   cancel,
		errChan:      make(chan error, 2),
	}

	leftRows := make(chan *results.Row)
	rightRows := make(chan *results.Row)

	initErrChan := make(chan error, 2)
	initDoneChan := make(chan struct{}, 2)

	procutil.PanicSafeGo(func() {
		iterator, err := join.left.Open(ctx, cfg, st)
		if err != nil {
			iter.errChan <- err
			return
		}
		iter.left = iterator
		initDoneChan <- struct{}{}
		iter.fetchRows(cancelCtx, iter.left, leftRows, iter.errChan)
	}, handleError(initErrChan))

	procutil.PanicSafeGo(func() {
		iterator, err := join.right.Open(ctx, cfg, st)
		if err != nil {
			iter.errChan <- err
			return
		}
		iter.right = iterator
		initDoneChan <- struct{}{}
		iter.fetchRows(cancelCtx, iter.right, rightRows, iter.errChan)
	}, handleError(initErrChan))

	joiner := &NestedLoopJoiner{
		cfg,
		st.WithCollation(join.Collation()),
		stageMonitor,
		join.matcher,
		join.left.Columns(),
		join.right.Columns(),
		join.kind,
		iter.errChan,
	}

	// Wait for initialization.
	for doneCount := 0; doneCount < 2; {
		select {
		case err := <-initErrChan:
			return nil, err
		case <-initDoneChan:
			doneCount++
		}
	}

	iter.onChan = joiner.Join(cancelCtx, leftRows, rightRows)

	return iter, nil
}

// fetchRows reads Row objects from a given Iter, and publishes them on channel
// ch, closing it when the iterator is exhausted. Errors encountered during
// iteration are published on errChan.
func (iter *JoinIter) fetchRows(ctx context.Context,
	it results.RowIter,
	ch chan<- *results.Row,
	errChan chan<- error) {
	r := &results.Row{}

	syncChan := make(chan *results.Row)
	fetchErrChan := make(chan error, 1)

	procutil.PanicSafeGo(func() {
	iterLoop:
		for it.Next(ctx, r) {
			select {
			case syncChan <- r:
				r = &results.Row{}
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
			close(ch)
			return
		case err := <-fetchErrChan:
			errChan <- err
			close(ch)
			return
		}
	}
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (iter *JoinIter) Next(_ context.Context, row *results.Row) bool {
	var ok bool
	select {
	case err := <-iter.errChan:
		iter.err = err
		return false
	case row.Data, ok = <-iter.onChan:
		if !ok {
			return false
		}

		iter.err = iter.stageMonitor.Exclude(row.Data.Size())
		return iter.err == nil
	}
}

// Close closes the iterator, returning any error encountered while doing so.
func (iter *JoinIter) Close() error {
	iter.cancelIter()

	err := iter.left.Close()
	if err != nil {
		// There is no way to combine errors.
		_ = iter.right.Close()
		return err
	}

	err = iter.right.Close()
	if err != nil {
		return err
	}

	_, err = iter.stageMonitor.Clear()
	return err
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (join *JoinStage) Columns() []*results.Column {
	left := join.left.Columns()
	right := join.right.Columns()
	columns := make([]*results.Column, len(left), len(left)+len(right))
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

func (nlp *NestedLoopJoiner) readData(ctx context.Context,
	lChan,
	rChan <-chan *results.Row) ([]*results.Row,
	[]*results.Row,
	error) {

	var left []*results.Row
	var right []*results.Row
	errs := make(chan error, 2)

	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	readChan := func(ch <-chan *results.Row, out *[]*results.Row) {
		for {
			select {
			case r, more := <-ch:
				if !more {
					errs <- nil
					return
				}

				*out = append(*out, r)
				err := nlp.stageMonitor.Include(r.Data.Size())
				if err != nil {
					errs <- err
					cancel()
					return
				}
			case <-cancelCtx.Done():
				errs <- cancelCtx.Err()
				return
			}
		}
	}

	procutil.PanicSafeGo(func() {
		readChan(lChan, &left)
	}, func(err interface{}) {
		cancel()
		errs <- fmt.Errorf("%v", err)
	})

	procutil.PanicSafeGo(func() {
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
	rChan <-chan *results.Row) <-chan results.RowValues {

	getNullValues := func(columns []*results.Column) results.RowValues {
		var nilValues results.RowValues
		for _, c := range columns {
			nilValue := values.NewSQLNull(nlp.cfg.sqlValueKind)
			nilValues = append(nilValues, results.NewRowValue(
				c.SelectID,
				c.Database,
				c.Table,
				c.Name,
				nilValue))
		}
		return nilValues
	}

	left, right, err := nlp.readData(ctx, lChan, rChan)
	if err != nil {
		nlp.errChan <- err
		return nil
	}

	ch := make(chan results.RowValues)

	switch nlp.kind {
	case CrossJoin:
		procutil.PanicSafeGo(func() {
			nlp.crossJoin(ctx, left, right, ch)
		}, func(err interface{}) {
			nlp.errChan <- fmt.Errorf("%v", err)
		})
	case InnerJoin, StraightJoin:
		procutil.PanicSafeGo(func() {
			nlp.innerJoin(ctx, left, right, ch)
		}, func(err interface{}) {
			nlp.errChan <- fmt.Errorf("%v", err)
		})
	case LeftJoin:
		procutil.PanicSafeGo(func() {
			nlp.leftJoin(ctx, left, right, ch, getNullValues(nlp.rightColumns))
		}, func(err interface{}) {
			nlp.errChan <- fmt.Errorf("%v", err)
		})
	case RightJoin:
		procutil.PanicSafeGo(func() {
			nlp.rightJoin(ctx, left, right, ch, getNullValues(nlp.leftColumns))
		}, func(err interface{}) {
			nlp.errChan <- fmt.Errorf("%v", err)
		})
	}

	return ch
}

func (nlp *NestedLoopJoiner) innerJoin(ctx context.Context,
	left,
	right []*results.Row,
	ch chan<- results.RowValues) {

outerLoop:
	for i, l := range left {
		err := nlp.stageMonitor.Release(l.Data.Size())
		if err != nil {
			nlp.errChan <- err
			break outerLoop
		}
		for _, r := range right {
			if i == 0 {
				err = nlp.stageMonitor.Release(r.Data.Size())
				if err != nil {
					nlp.errChan <- err
					break outerLoop
				}
			}

			st := nlp.st.WithRows(l, r)
			result, err := nlp.matcher.Evaluate(ctx, nlp.cfg, st)
			if err != nil {
				nlp.errChan <- err
				break outerLoop
			} else if values.Bool(result) {
				err = nlp.stageMonitor.Acquire(l.Data.Size() + r.Data.Size())
				if err != nil {
					nlp.errChan <- err
					break outerLoop
				}
				values := make(results.RowValues, len(l.Data)+len(r.Data))
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
	right []*results.Row,
	ch chan<- results.RowValues,
	nilRightValues results.RowValues) {

	var hasMatch bool

outerLoop:
	for i, l := range left {
		err := nlp.stageMonitor.Release(l.Data.Size())
		if err != nil {
			nlp.errChan <- err
			break outerLoop
		}

		for _, r := range right {
			if i == 0 {
				err = nlp.stageMonitor.Release(r.Data.Size())
				if err != nil {
					nlp.errChan <- err
					return
				}
			}

			st := nlp.st.WithRows(l, r)
			var result values.SQLValue
			result, err = nlp.matcher.Evaluate(ctx, nlp.cfg, st)
			if err != nil {
				nlp.errChan <- err
				break outerLoop
			}
			if values.Bool(result) {
				err = nlp.stageMonitor.Acquire(l.Data.Size() + r.Data.Size())
				if err != nil {
					nlp.errChan <- err
					break outerLoop
				}
				hasMatch = true
				values := make(results.RowValues, len(l.Data)+len(r.Data))
				copy(values, append(l.Data, r.Data...))
				select {
				case ch <- values:
				case <-ctx.Done():
					break outerLoop
				}
			}
		}

		if !hasMatch {
			err = nlp.stageMonitor.Acquire(l.Data.Size() + nilRightValues.Size())
			if err != nil {
				nlp.errChan <- err
				break outerLoop
			}
			values := make(results.RowValues, len(nilRightValues)+len(l.Data))
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
	right []*results.Row,
	ch chan<- results.RowValues,
	nilLeftValues results.RowValues) {

	var hasMatch bool

outerLoop:
	for i, r := range right {
		err := nlp.stageMonitor.Release(r.Data.Size())
		if err != nil {
			nlp.errChan <- err
			break outerLoop
		}

		for _, l := range left {
			if i == 0 {
				err = nlp.stageMonitor.Release(l.Data.Size())
				if err != nil {
					nlp.errChan <- err
					break outerLoop
				}
			}

			st := nlp.st.WithRows(l, r)
			var result values.SQLValue
			result, err = nlp.matcher.Evaluate(ctx, nlp.cfg, st)
			if err != nil {
				nlp.errChan <- err
				break outerLoop
			} else if values.Bool(result) {
				err = nlp.stageMonitor.Acquire(l.Data.Size() + r.Data.Size())
				if err != nil {
					nlp.errChan <- err
					break outerLoop
				}
				hasMatch = true
				values := make(results.RowValues, len(l.Data)+len(r.Data))
				copy(values, append(l.Data, r.Data...))
				select {
				case ch <- values:
				case <-ctx.Done():
					break outerLoop
				}
			}
		}

		if !hasMatch {
			err = nlp.stageMonitor.Acquire(r.Data.Size() + nilLeftValues.Size())
			if err != nil {
				nlp.errChan <- err
				break outerLoop
			}
			values := make(results.RowValues, len(nilLeftValues)+len(r.Data))
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
	right []*results.Row,
	ch chan<- results.RowValues) {

outerLoop:
	for i, l := range left {
		err := nlp.stageMonitor.Release(l.Data.Size())
		if err != nil {
			nlp.errChan <- err
			break outerLoop
		}
		for _, r := range right {
			if i == 0 {
				err = nlp.stageMonitor.Release(r.Data.Size())
				if err != nil {
					nlp.errChan <- err
					break outerLoop
				}
			}

			err = nlp.stageMonitor.Acquire(l.Data.Size() + r.Data.Size())
			if err != nil {
				nlp.errChan <- err
				break outerLoop
			}
			values := make(results.RowValues, len(l.Data)+len(r.Data))
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

func (join *JoinStage) clone() PlanStage {
	return NewJoinStage(join.kind, join.left.clone(), join.right.clone(), join.matcher)
}
