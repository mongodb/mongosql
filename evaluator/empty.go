package evaluator

type Empty struct {
}

func (_ *Empty) Open(ctx *ExecutionCtx) error {
	return nil
}

func (_ *Empty) Next(row *Row) bool {
	return false
}

func (_ *Empty) OpFields() (columns []*Column) {
	return []*Column{}
}

func (_ *Empty) Close() error {
	return nil
}

func (_ *Empty) Err() error {
	return nil
}
