package evaluator

// An empty source for when we find that 0 rows are going to be returned, we don't
// need to hit MongoDB to get back nothing.
type Empty struct {
}

func (_ *Empty) Open(ctx *ExecutionCtx) error {
	return nil
}

func (_ *Empty) Next(row *Row) bool {
	return false
}

func (_ *Empty) OpFields() []*Column {
	return []*Column{}
}

func (_ *Empty) Close() error {
	return nil
}

func (_ *Empty) Err() error {
	return nil
}
