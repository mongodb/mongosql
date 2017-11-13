package evaluator

import "github.com/10gen/sqlproxy/collation"

// An EmptyStage is for when we find that 0 rows are going to be returned: we don't
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

func (*EmptyStage) Open(ctx *ExecutionCtx) (Iter, error) {
	return &EmptyIter{}, nil
}

func (es *EmptyStage) Columns() []*Column {
	return es.columns
}

func (es *EmptyStage) Collation() *collation.Collation {
	return es.collation
}

func (*EmptyIter) Next(row *Row) bool {
	return false
}

func (*EmptyIter) Close() error {
	return nil
}

func (*EmptyIter) Err() error {
	return nil
}
