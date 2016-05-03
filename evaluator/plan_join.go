package evaluator

import (
	"fmt"
	"sync"

	"github.com/deafgoat/mixer/sqlparser"
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
	InnerJoin    JoinKind = sqlparser.AST_JOIN
	StraightJoin JoinKind = sqlparser.AST_STRAIGHT_JOIN
	LeftJoin     JoinKind = sqlparser.AST_LEFT_JOIN
	RightJoin    JoinKind = sqlparser.AST_RIGHT_JOIN
	CrossJoin    JoinKind = sqlparser.AST_CROSS_JOIN
	NaturalJoin  JoinKind = sqlparser.AST_NATURAL_JOIN
)

// NestedLoop implementation of a JOIN.
type NestedLoopJoiner struct {
	matcher SQLExpr
	kind    JoinKind
	errChan chan error
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
	left, right PlanStage
	matcher     SQLExpr
	kind        JoinKind
	strategy    JoinStrategy
}

type JoinIter struct {
	strategy    JoinStrategy
	kind        JoinKind
	matcher     SQLExpr
	joiner      Joiner
	left, right Iter
	execCtx     *ExecutionCtx

	leftRows  chan *Row
	rightRows chan *Row
	onChan    chan Row
	errChan   chan error
	init      sync.Once
	err       error
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

	return &JoinIter{
		kind:     join.kind,
		strategy: join.strategy,
		matcher:  join.matcher,
		left:     left,
		right:    right,
		execCtx:  ctx,
	}, nil
}

// fetchRows reads Row objects from a given Iter, and publishes them on channel ch, closing it when
// the iterator is exhausted. Errors encountered during iteration are published on errChan.
func fetchRows(it Iter, ch chan *Row, errChan chan error) {
	r := &Row{}

	for it.Next(r) {
		ch <- r
		r = &Row{}
	}
	close(ch)

	if err := it.Err(); err != nil {
		errChan <- err
	}
}

func (join *JoinIter) Next(row *Row) bool {
	join.init.Do(func() {
		join.errChan = make(chan error, 1)
		join.joiner = NewJoiner(join.strategy, join.kind, join.matcher, join.errChan)
		join.leftRows = make(chan *Row)
		join.rightRows = make(chan *Row)
		go fetchRows(join.left, join.leftRows, join.errChan)
		go fetchRows(join.right, join.rightRows, join.errChan)
		join.onChan = join.joiner.Join(join.leftRows, join.rightRows, join.execCtx)
	})
	select {
	case err := <-join.errChan:
		join.err = err
		return false
	case data, ok := <-join.onChan:
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

func (join *JoinStage) OpFields() []*Column {
	left := join.left.OpFields()
	right := join.right.OpFields()
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
		left:     join.left,
		right:    join.right,
		matcher:  join.matcher,
		kind:     join.kind,
		strategy: join.strategy,
	}
}

// NewJoiner returns a new Joiner implementation for the given
// strategy. The implementation uses the supplied matcher in
// evaluating the join criteria and performs joins according
// to the joinType
func NewJoiner(s JoinStrategy, kind JoinKind, matcher SQLExpr, errChan chan error) Joiner {

	switch s {
	case NestedLoop:
		return &NestedLoopJoiner{matcher, kind, errChan}
	case SortMerge:
		return &SortMergeJoiner{matcher, kind, errChan}
	case Hash:
		return &HashJoiner{matcher, kind, errChan}
	default:
		panic(fmt.Sprintf("unknown join strategy: %v", s))
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

	ch := make(chan Row)

	switch nlp.kind {

	case InnerJoin:
		go nlp.innerJoin(lChan, rChan, ch, ctx)
	case LeftJoin:
		go nlp.sideJoin(lChan, rChan, ch, ctx)
	case RightJoin:
		go nlp.sideJoin(rChan, lChan, ch, ctx)
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
			evalCtx := &EvalCtx{Rows{*l, *r}, ctx}
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

func (nlp *NestedLoopJoiner) sideJoin(lChan, rChan chan *Row, ch chan Row, ctx *ExecutionCtx) {
	left := readFromChan(lChan)
	right := readFromChan(rChan)

	var hasMatch bool

	for _, l := range left {
		for _, r := range right {
			evalCtx := &EvalCtx{Rows{*l, *r}, ctx}
			m, err := Matches(nlp.matcher, evalCtx)
			if err != nil {
				nlp.errChan <- err
			} else if m {
				hasMatch = true
				ch <- Row{Data: append(l.Data, r.Data...)}
			}
		}

		if !hasMatch {
			ch <- Row{Data: l.Data}
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
