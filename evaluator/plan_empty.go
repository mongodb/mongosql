package evaluator

import "github.com/10gen/sqlproxy/collation"

// An empty source for when we find that 0 rows are going to be returned, we don't
// need to hit MongoDB to get back nothing.
type EmptyStage struct {
	columns   []*Column
	collation *collation.Collation
}

// NewEmptyStage creates a new Empty stage.
func NewEmptyStage(columns []*Column, collation *collation.Collation) *EmptyStage {
	return &EmptyStage{columns, collation}
}

type EmptyIter struct{}

func (_ *EmptyStage) Open(ctx *ExecutionCtx) (Iter, error) {
	return &EmptyIter{}, nil
}

func (es *EmptyStage) Columns() []*Column {
	return es.columns
}

func (es *EmptyStage) Collation() *collation.Collation {
	return es.collation
}

func (_ *EmptyIter) Next(row *Row) bool {
	return false
}

func (_ *EmptyIter) Close() error {
	return nil
}

func (_ *EmptyIter) Err() error {
	return nil
}
