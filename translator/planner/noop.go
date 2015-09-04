package planner

type Noop struct {
}

func (*Noop) Open(ctx *ExecutionCtx) error {
	return nil
}

func (*Noop) Next(row *Row) bool {
	return false
}

func (*Noop) Close() error {
	return nil
}

func (*Noop) Err() error {
	return nil
}
