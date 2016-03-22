package evaluator

import (
	"fmt"

	"github.com/deafgoat/mixer/sqlparser"
	"gopkg.in/mgo.v2/bson"
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
type Join struct {
	left, right Operator
	matcher     SQLExpr
	err         error
	kind        JoinKind
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

	switch s {
	case NestedLoop:
		return &NestedLoopJoiner{matcher, join.kind, errChan}, nil
	case SortMerge:
		return &SortMergeJoiner{matcher, join.kind, errChan}, nil
	case Hash:
		return &HashJoiner{matcher, join.kind, errChan}, nil
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

///////////////
//Optimization
///////////////

const (
	joinedFieldNamePrefix = "__joined_"
)

func (v *optimizer) visitJoin(join *Join) (Operator, error) {

	// 1. the join type must be usable. MongoDB can only do an inner join and a left outer join
	var localSource, foreignSource Operator
	var joinKind JoinKind

	switch join.kind {
	case InnerJoin:
		localSource = join.left
		foreignSource = join.right
		joinKind = InnerJoin
	case LeftJoin:
		localSource = join.left
		foreignSource = join.right
		joinKind = LeftJoin
	case RightJoin:
		localSource = join.right
		foreignSource = join.left
		joinKind = LeftJoin
	default:
		return join, nil
	}

	// 2. we have to be able to push both down and the foreign TableScan
	// operator must have nothing in its pipeline.
	msLocal, ok := localSource.(*MongoSource)
	if !ok {
		return join, nil
	}

	msForeign, ok := foreignSource.(*MongoSource)
	if !ok {
		return join, nil
	}

	if joinKind == InnerJoin && len(msLocal.pipeline) == 0 && len(msForeign.pipeline) > 0 {
		// flip them
		msLocal, msForeign = msForeign, msLocal
	} else if len(msForeign.pipeline) > 0 {
		return join, nil
	}

	// 3. find the local column and the foreign column
	localColumn, foreignColumn, err := getLocalAndForeignColumns(msLocal.aliasName, msForeign.aliasName, join.matcher)
	if err != nil {
		return join, nil
	}

	// 4. construct the $lookup clause
	pipeline := msLocal.pipeline

	localFieldName, ok := msLocal.mappingRegistry.lookupFieldName(localColumn.tableName, localColumn.columnName)
	if !ok {
		return join, nil
	}
	foreignFieldName, ok := msForeign.mappingRegistry.lookupFieldName(foreignColumn.tableName, foreignColumn.columnName)
	if !ok {
		return join, nil
	}

	asField := joinedFieldNamePrefix + msForeign.collectionName

	lookup := bson.M{
		"from":         msForeign.collectionName,
		"localField":   localFieldName,
		"foreignField": foreignFieldName,
		"as":           asField,
	}

	// 5. construct the $unwind clause
	var unwind bson.M

	switch joinKind { // right join was already flipped
	case InnerJoin:
		unwind = bson.M{
			"path": "$" + asField,
			"preserveNullAndEmptyArrays": false,
		}
	case LeftJoin:
		unwind = bson.M{
			"path": "$" + asField,
			"preserveNullAndEmptyArrays": true,
		}
	}

	pipeline = append(pipeline, bson.D{{"$lookup", lookup}})
	pipeline = append(pipeline, bson.D{{"$unwind", unwind}})

	// 6. change all the mappings from the msForeign mapping registry to be nested under
	// the 'asField' we used above.
	newMappingRegistry := msLocal.mappingRegistry.copy()

	newMappingRegistry.columns = append(newMappingRegistry.columns, msForeign.mappingRegistry.columns...)
	if msForeign.mappingRegistry.fields != nil {
		for tableName, columns := range msForeign.mappingRegistry.fields {
			for columnName, fieldName := range columns {
				newMappingRegistry.registerMapping(tableName, columnName, asField+"."+fieldName)
			}
		}
	}

	ms := msLocal.WithPipeline(pipeline).WithMappingRegistry(newMappingRegistry)
	return ms, nil
}

func getLocalAndForeignColumns(localTableName, foreignTableName string, e SQLExpr) (*SQLColumnExpr, *SQLColumnExpr, error) {

	// TODO: we can probably extract from 'e' the parts that only deal with the left or
	// right sides, but not both. These parts of the predicates need to get pushed down
	// independently such that filters for each side, if necessary, are performed
	// server side.

	optimizedExpr, err := OptimizeSQLExpr(e)
	if err != nil {
		return nil, nil, err
	}

	// anything in optimizedExpr that is not an equi-join makes this impossible to push down
	equalExpr, ok := optimizedExpr.(*SQLEqualsExpr)
	if !ok {
		return nil, nil, fmt.Errorf("join condition cannot be pushed down '%v'", e.String())
	}

	// we must have a field from the left table and and a field from the right table
	column1, ok := equalExpr.left.(SQLColumnExpr)
	if !ok {
		return nil, nil, fmt.Errorf("join condition cannot be pushed down '%v'", equalExpr.String())
	}
	column2, ok := equalExpr.right.(SQLColumnExpr)
	if !ok {
		return nil, nil, fmt.Errorf("join condition cannot be pushed down '%v'", equalExpr.String())
	}

	var localColumn, foreignColumn *SQLColumnExpr
	if column1.tableName == localTableName {
		localColumn = &column1
	} else if column1.tableName == foreignTableName {
		foreignColumn = &column1
	}

	if column2.tableName == localTableName {
		localColumn = &column2
	} else if column2.tableName == foreignTableName {
		foreignColumn = &column2
	}

	if localColumn == nil || foreignColumn == nil {
		return nil, nil, fmt.Errorf("join condition cannot be pushed down '%v'", equalExpr.String())
	}

	return localColumn, foreignColumn, nil
}
