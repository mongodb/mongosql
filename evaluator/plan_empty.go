package evaluator

// An empty source for when we find that 0 rows are going to be returned, we don't
// need to hit MongoDB to get back nothing.
type EmptyStage struct {
	columns []*Column
}

type EmptyIter struct{}

func (_ *EmptyStage) Open(ctx *ExecutionCtx) (Iter, error) {
	return &EmptyIter{}, nil
}

func (es *EmptyStage) OpFields() []*Column {
	return es.columns
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
