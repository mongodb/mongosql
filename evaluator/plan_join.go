package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/parser"
)

// JoinStrategy specifies the method a Join
// operator utilizes in performing a join operation.
type JoinStrategy byte

const (
	NestedLoop JoinStrategy = iota
	SortMerge
	Hash
)

// JoinKind specifies the the type of join for
// a given joiner.
type JoinKind string

const (
	InnerJoin    JoinKind = parser.AST_JOIN
	StraightJoin JoinKind = parser.AST_STRAIGHT_JOIN
	LeftJoin     JoinKind = parser.AST_LEFT_JOIN
	RightJoin    JoinKind = parser.AST_RIGHT_JOIN
	CrossJoin    JoinKind = parser.AST_CROSS_JOIN
	NaturalJoin  JoinKind = parser.AST_NATURAL_JOIN
)

type JoinChild byte

const (
	leftJoinChild = iota
	rightJoinChild
)

// NestedLoop implementation of a JOIN.
type NestedLoopJoiner struct {
	matcher      SQLExpr
	leftColumns  []*Column
	rightColumns []*Column
	kind         JoinKind
	errChan      chan error
}

// SortMerge implementation of a JOIN.
type SortMergeJoiner struct {
	matcher SQLExpr
	kind    JoinKind
	errChan chan error
}

// Hash implementation of a JOIN.
type HashJoiner struct {
	matcher SQLExpr
	kind    JoinKind
	errChan chan error
}

// Joiner wraps the basic Join function that is
// used to combine data from two different sources.
type Joiner interface {
	Join(left, right chan *Row, ctx *ExecutionCtx) chan Row
}

// Join implements the operator interface for
// join expressions.
type JoinStage struct {
	left, right     PlanStage
	matcher         SQLExpr
	kind            JoinKind
	strategy        JoinStrategy
	requiredColumns []SQLExpr
}

func NewJoinStage(kind JoinKind, left, right PlanStage, predicate SQLExpr, reqCols []SQLExpr) *JoinStage {
	return &JoinStage{
		kind:            kind,
		left:            left,
		right:           right,
		matcher:         predicate,
		requiredColumns: reqCols,
	}
}

type JoinIter struct {
	left, right Iter
	ctx         *ExecutionCtx
	onChan      chan Row
	errChan     chan error
	err         error
}

func (join *JoinStage) Open(ctx *ExecutionCtx) (Iter, error) {
	left, err := join.left.Open(ctx)
	if err != nil {
		return nil, err
	}

	right, err := join.right.Open(ctx)
	if err != nil {
		return nil, err
	}

	iter := &JoinIter{
		left:    left,
		right:   right,
		ctx:     ctx,
		errChan: make(chan error, 1),
	}

	leftRows := make(chan *Row)
	rightRows := make(chan *Row)

	go iter.fetchRows(left, leftRows, iter.errChan)
	go iter.fetchRows(right, rightRows, iter.errChan)

	joiner := NewJoiner(join.strategy, join.kind, join.matcher, join.left.Columns(), join.right.Columns(), iter.errChan)

	iter.onChan = joiner.Join(leftRows, rightRows, ctx)

	return iter, nil
}

// fetchRows reads Row objects from a given Iter, and publishes them on channel ch, closing it when
// the iterator is exhausted. Errors encountered during iteration are published on errChan.
func (iter *JoinIter) fetchRows(it Iter, ch chan *Row, errChan chan error) {
	r := &Row{}

	syncChan := make(chan *Row)

	go func() {
		for it.Next(r) {
			syncChan <- r
			r = &Row{}
		}
		close(syncChan)
	}()

	for {
		select {
		case row, ok := <-syncChan:
			if !ok {
				close(ch)
				return
			}

			if err := it.Err(); err != nil {
				errChan <- err
				return
			}

			ch <- row
		case <-iter.ctx.Tomb().Dying():
			errChan <- iter.ctx.Tomb().Err()
			return
		}
	}
}

func (iter *JoinIter) Next(row *Row) bool {
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

func (join *JoinIter) Close() error {

	if err := join.left.Close(); err != nil {
		return err
	}

	if err := join.right.Close(); err != nil {
		return err
	}

	return nil
}

func (join *JoinStage) Columns() []*Column {
	left := join.left.Columns()
	right := join.right.Columns()
	return append(left, right...)
}

func (join *JoinIter) Err() error {

	if err := join.left.Err(); err != nil {
		return err
	}

	if err := join.right.Err(); err != nil {
		return err
	}

	return join.err
}

func (join *JoinStage) clone() *JoinStage {
	return &JoinStage{
		left:            join.left,
		right:           join.right,
		matcher:         join.matcher,
		kind:            join.kind,
		strategy:        join.strategy,
		requiredColumns: join.requiredColumns,
	}
}

// NewJoiner returns a new Joiner implementation for the given
// strategy. The implementation uses the supplied matcher in
// evaluating the join criteria and performs joins according
// to the joinType
func NewJoiner(s JoinStrategy, kind JoinKind, matcher SQLExpr, leftColumns, rightColumns []*Column, errChan chan error) Joiner {

	switch s {
	case NestedLoop:
		return &NestedLoopJoiner{matcher, leftColumns, rightColumns, kind, errChan}
	case SortMerge:
		return &SortMergeJoiner{matcher, kind, errChan}
	case Hash:
		return &HashJoiner{matcher, kind, errChan}
	default:
		panic(fmt.Sprintf("unsupported join strategy: %v", s))
	}
}

// readFromChan reads data from the ch channel and
// returns all the data read as a slice of Rows.
func readFromChan(ch chan *Row) []*Row {
	r := []*Row{}

	for data := range ch {
		r = append(r, data)
	}

	return r
}

// NestedLoopJoiner implementation.
func (nlp *NestedLoopJoiner) Join(lChan, rChan chan *Row, ctx *ExecutionCtx) chan Row {

	getNilValues := func(columns []*Column) Values {
		var nilValues Values
		for _, c := range columns {
			nilValues = append(nilValues, Value{
				SelectID: c.SelectID,
				Table:    c.Table,
				Name:     c.Name,
			})
		}
		return nilValues
	}

	ch := make(chan Row)

	switch nlp.kind {
	case InnerJoin:
		go nlp.innerJoin(lChan, rChan, ch, ctx)
	case LeftJoin:
		go nlp.leftJoin(lChan, rChan, ch, ctx, getNilValues(nlp.rightColumns))
	case RightJoin:
		go nlp.rightJoin(lChan, rChan, ch, ctx, getNilValues(nlp.leftColumns))
	case StraightJoin:
	case CrossJoin:
		go nlp.crossJoin(lChan, rChan, ch, ctx)
	case NaturalJoin:
	}

	return ch
}

func (nlp *NestedLoopJoiner) innerJoin(lChan, rChan chan *Row, ch chan Row, ctx *ExecutionCtx) {
	left := readFromChan(lChan)
	right := readFromChan(rChan)

	for _, l := range left {
		for _, r := range right {
			evalCtx := NewEvalCtx(ctx, l, r)
			m, err := Matches(nlp.matcher, evalCtx)
			if err != nil {
				nlp.errChan <- err
			} else if m {
				ch <- Row{Data: append(l.Data, r.Data...)}
			}
		}
	}

	close(ch)
}

func (nlp *NestedLoopJoiner) leftJoin(lChan, rChan chan *Row, ch chan Row, ctx *ExecutionCtx, nilRightValues Values) {
	left := readFromChan(lChan)
	right := readFromChan(rChan)

	var hasMatch bool

	for _, l := range left {
		for _, r := range right {
			evalCtx := NewEvalCtx(ctx, l, r)
			m, err := Matches(nlp.matcher, evalCtx)
			if err != nil {
				nlp.errChan <- err
			} else if m {
				hasMatch = true
				ch <- Row{Data: append(l.Data, r.Data...)}
			}
		}

		if !hasMatch {
			ch <- Row{Data: append(l.Data, nilRightValues...)}
		}

		hasMatch = false
	}

	close(ch)
}

func (nlp *NestedLoopJoiner) rightJoin(lChan, rChan chan *Row, ch chan Row, ctx *ExecutionCtx, nilLeftValues Values) {
	left := readFromChan(lChan)
	right := readFromChan(rChan)

	var hasMatch bool

	for _, r := range right {
		for _, l := range left {
			evalCtx := NewEvalCtx(ctx, l, r)
			m, err := Matches(nlp.matcher, evalCtx)
			if err != nil {
				nlp.errChan <- err
			} else if m {
				hasMatch = true
				ch <- Row{Data: append(l.Data, r.Data...)}
			}
		}

		if !hasMatch {
			ch <- Row{Data: append(nilLeftValues, r.Data...)}
		}

		hasMatch = false
	}

	close(ch)
}

func (nlp *NestedLoopJoiner) crossJoin(lChan, rChan chan *Row, ch chan Row, _ *ExecutionCtx) {
	left := readFromChan(lChan)
	right := readFromChan(rChan)

	for _, l := range left {
		for _, r := range right {
			ch <- Row{Data: append(l.Data, r.Data...)}
		}
	}

	close(ch)
}

// SortMergeJoiner implementation.
func (smj *SortMergeJoiner) Join(lChan, rChan chan *Row, ctx *ExecutionCtx) chan Row {
	return nil
}

// HashJoiner implementation.
func (hj *HashJoiner) Join(lChan, rChan chan *Row, ctx *ExecutionCtx) chan Row {
	return nil
}
