package evaluator

// An empty source for when we find that 0 rows are going to be returned, we don't
// need to hit MongoDB to get back nothing.
type EmptyStage struct{}
type EmptyIter struct{}

func (_ *EmptyStage) Open(ctx *ExecutionCtx) (Iter, error) {
	return &EmptyIter{}, nil
}
func (_ *EmptyStage) OpFields() []*Column {
	return []*Column{}
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
