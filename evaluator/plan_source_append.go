package evaluator

type SourceAppendStage struct {
	source PlanStage
}

// SourceAppend adds the current row to the source slice of the
// execution context.
type SourceAppendIter struct {
	// source holds the source for this select statement
	source Iter

	// ctx is the current execution context
	ctx *ExecutionCtx
}

func (sa *SourceAppendStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := sa.source.Open(ctx)
	if err != nil {
		return nil, err
	}
	return &SourceAppendIter{sourceIter, ctx}, nil
}

func (sa *SourceAppendIter) Next(row *Row) bool {
	hasNext := sa.source.Next(row)

	if !hasNext {
		return false
	}

	if len(sa.ctx.SrcRows) == sa.ctx.Depth {
		sa.ctx.SrcRows = append(sa.ctx.SrcRows, row)
	}

	return true
}

func (sa *SourceAppendStage) OpFields() (columns []*Column) {
	return sa.source.OpFields()
}

func (sa *SourceAppendIter) Close() error {
	return sa.source.Close()
}

func (sa *SourceAppendIter) Err() error {
	return sa.source.Err()
}
