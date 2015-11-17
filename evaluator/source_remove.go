package evaluator

// SourceRemove removes the current row from the source slice of the
// execution context.
type SourceRemove struct {
	// source holds the source for this select statement
	source Operator

	// ctx is the current execution context
	ctx *ExecutionCtx

	// hasSubquery is true if this operator source contains a subquery
	hasSubquery bool
}

func (sr *SourceRemove) Open(ctx *ExecutionCtx) error {
	sr.ctx = ctx
	return sr.source.Open(ctx)
}

func (sr *SourceRemove) Next(row *Row) bool {

	hasNext := sr.source.Next(row)

	if !hasNext {
		return false
	}

	bound := len(sr.ctx.SrcRows) - 1

	if sr.hasSubquery && bound == sr.ctx.Depth {
		sr.ctx.SrcRows = sr.ctx.SrcRows[:bound]
	}

	return true
}

func (sr *SourceRemove) OpFields() (columns []*Column) {
	return sr.source.OpFields()
}

func (sr *SourceRemove) Close() error {
	return sr.source.Close()
}

func (sr *SourceRemove) Err() error {
	return sr.source.Err()
}
