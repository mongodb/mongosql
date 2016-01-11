package evaluator

import (
	"fmt"
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

// JoinType specifies the the type of join for
// a given joiner.
type JoinType string

const (
	InnerJoin    JoinType = sqlparser.AST_JOIN
	StraightJoin          = sqlparser.AST_STRAIGHT_JOIN
	LeftJoin              = sqlparser.AST_LEFT_JOIN
	RightJoin             = sqlparser.AST_RIGHT_JOIN
	CrossJoin             = sqlparser.AST_CROSS_JOIN
	NaturalJoin           = sqlparser.AST_NATURAL_JOIN
)

// NestedLoop implementation of a JOIN.
type NestedLoopJoiner struct {
	matcher  SQLExpr
	joinType JoinType
	errChan  chan error
}

// SortMerge implementation of a JOIN.
type SortMergeJoiner struct {
	matcher  SQLExpr
	joinType JoinType
	errChan  chan error
}

// Hash implementation of a JOIN.
type HashJoiner struct {
	matcher  SQLExpr
	joinType JoinType
	errChan  chan error
}

// Joiner wraps the basic Join function that is
// used to combine data from two different sources.
type Joiner interface {
	Join(left, right chan *Row, ctx *ExecutionCtx) chan Row
}

// Join implements the operator interface for
// join expressions.
type Join struct {
	left, right Operator
	matcher     SQLExpr
	on          sqlparser.BoolExpr
	err         error
	kind        string
	strategy    JoinStrategy
	leftRows    chan *Row
	rightRows   chan *Row
	onChan      chan Row
	errChan     chan error
}

func (join *Join) Open(ctx *ExecutionCtx) error {
	return join.init(ctx)
}

func (join *Join) fetchRows(opr Operator, ch chan *Row) {

	r := &Row{}

	for opr.Next(r) {
		ch <- r
		r = &Row{}
	}
	close(ch)

	if err := opr.Err(); err != nil {
		join.errChan <- err
	}
}

func (join *Join) init(ctx *ExecutionCtx) (err error) {

	// default join mechanism is nested loop
	err = join.left.Open(ctx)
	if err != nil {
		return err
	}

	err = join.right.Open(ctx)
	if err != nil {
		return err
	}

	join.leftRows = make(chan *Row)
	go join.fetchRows(join.left, join.leftRows)

	join.rightRows = make(chan *Row)
	go join.fetchRows(join.right, join.rightRows)

	join.errChan = make(chan error, 1)

	join.matcher, err = NewSQLExpr(join.on)
	if err != nil {
		return err
	}

	joiner, err := NewJoiner(join)
	if err != nil {
		return err
	}

	join.onChan = joiner.Join(join.leftRows, join.rightRows, ctx)

	return nil

}

func (join *Join) Next(row *Row) bool {
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

func (join *Join) Close() error {

	if err := join.left.Close(); err != nil {
		return err
	}

	if err := join.right.Close(); err != nil {
		return err
	}

	return nil
}

func (join *Join) OpFields() []*Column {
	left := join.left.OpFields()
	right := join.right.OpFields()
	return append(left, right...)
}

func (join *Join) Err() error {

	if err := join.left.Err(); err != nil {
		return err
	}

	if err := join.right.Err(); err != nil {
		return err
	}

	return join.err
}

// NewJoiner returns a new Joiner implementation for the given
// strategy. The implementation uses the supplied matcher in
// evaluating the join criteria and performs joins according
// to the joinType
func NewJoiner(join *Join) (Joiner, error) {

	s := join.strategy
	matcher := join.matcher
	errChan := join.errChan
	joinType := getJoinKind(join.kind)

	switch s {
	case NestedLoop:
		return &NestedLoopJoiner{matcher, joinType, errChan}, nil
	case SortMerge:
		return &SortMergeJoiner{matcher, joinType, errChan}, nil
	case Hash:
		return &HashJoiner{matcher, joinType, errChan}, nil
	default:
		return nil, fmt.Errorf("unknown join strategy")
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

// getJoinKind returns the join type for the given string.
func getJoinKind(s string) JoinType {
	switch s {
	case sqlparser.AST_JOIN:
		return InnerJoin
	case sqlparser.AST_STRAIGHT_JOIN:
		return StraightJoin
	case sqlparser.AST_LEFT_JOIN:
		return LeftJoin
	case sqlparser.AST_RIGHT_JOIN:
		return RightJoin
	case sqlparser.AST_CROSS_JOIN:
		return CrossJoin
	case sqlparser.AST_NATURAL_JOIN:
		return NaturalJoin
	default:
		return ""
	}
}

// NestedLoopJoiner implementation.
func (nlp *NestedLoopJoiner) Join(lChan, rChan chan *Row, ctx *ExecutionCtx) chan Row {

	ch := make(chan Row)

	switch nlp.joinType {

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
			evalCtx := &EvalCtx{[]Row{*l, *r}, ctx}
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
			evalCtx := &EvalCtx{[]Row{*l, *r}, ctx}
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
