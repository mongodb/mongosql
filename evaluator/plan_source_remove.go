package evaluator

// SourceRemove removes the current row from the source slice of the
// execution context.
type SourceRemoveStage struct {
	// source holds the source for this select statement
	source PlanStage
}

type SourceRemoveIter struct {
	source Iter
	// ctx is the current execution context
	ctx *ExecutionCtx
}

func (sr *SourceRemoveStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := sr.source.Open(ctx)
	if err != nil {
		return nil, err
	}
	return &SourceRemoveIter{sourceIter, ctx}, nil
}

func (sr *SourceRemoveIter) Next(row *Row) bool {

	hasNext := sr.source.Next(row)

	if !hasNext {
		return false
	}

	bound := len(sr.ctx.SrcRows) - 1

	if bound == sr.ctx.Depth {
		sr.ctx.SrcRows = sr.ctx.SrcRows[:bound]
	}

	return true
}

func (sr *SourceRemoveStage) OpFields() (columns []*Column) {
	return sr.source.OpFields()
}

func (sr *SourceRemoveIter) Close() error {
	return sr.source.Close()
}

func (sr *SourceRemoveIter) Err() error {
	return sr.source.Err()
}
