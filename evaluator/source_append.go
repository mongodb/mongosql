package evaluator

// SourceAppend adds the current row to the source slice of the
// execution context.
type SourceAppend struct {
	// source holds the source for this select statement
	source Operator

	// ctx is the current execution context
	ctx *ExecutionCtx
}

func (sa *SourceAppend) Open(ctx *ExecutionCtx) error {
	sa.ctx = ctx
	return sa.source.Open(ctx)
}

func (sa *SourceAppend) Next(row *Row) bool {
	hasNext := sa.source.Next(row)

	if !hasNext {
		return false
	}

	if len(sa.ctx.SrcRows) == sa.ctx.Depth {
		sa.ctx.SrcRows = append(sa.ctx.SrcRows, row)
	}

	return true
}

func (sa *SourceAppend) OpFields() (columns []*Column) {
	return sa.source.OpFields()
}

func (sa *SourceAppend) Close() error {
	return sa.source.Close()
}

func (sa *SourceAppend) Err() error {
	return sa.source.Err()
}
