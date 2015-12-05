package evaluator

// Filter ensures that only rows matching a given criteria are
// returned.
type Filter struct {
	// err holds any error that may have occurred during processing
	err error

	// source holds the source for this select statement
	source Operator

	// matcher is used to filter results gotten from the source operator
	matcher Matcher

	// ctx is the current execution context
	ctx *ExecutionCtx

	// hasSubquery is true if this operator source contains a subquery
	hasSubquery bool
}

func (ft *Filter) Open(ctx *ExecutionCtx) error {
	ft.ctx = ctx
	return ft.source.Open(ctx)
}

func (ft *Filter) Next(row *Row) bool {

	var hasNext bool

	for {

		hasNext = ft.source.Next(row)

		if !hasNext {
			break
		}

		evalCtx := &EvalCtx{[]Row{*row}, ft.ctx}

		// add parent row(s) to this subquery's evaluation context
		if len(ft.ctx.SrcRows) != 0 {

			bound := len(ft.ctx.SrcRows) - 1

			for _, r := range ft.ctx.SrcRows[:bound] {
				evalCtx.Rows = append(evalCtx.Rows, *r)
			}

			// avoid duplication since subquery row is most recently
			// appended and "*row"
			if !ft.hasSubquery {
				evalCtx.Rows = append(evalCtx.Rows, *ft.ctx.SrcRows[bound])
			}

		}

		if ft.matcher != nil {
			m, err := ft.matcher.Matches(evalCtx)
			if err != nil {
				ft.err = err
				return false
			}
			if m {
				break
			}
		} else {
			break
		}

	}

	return hasNext
}

func (ft *Filter) OpFields() (columns []*Column) {
	return ft.source.OpFields()
}

func (ft *Filter) Close() error {
	return ft.source.Close()
}

func (ft *Filter) Err() error {
	if err := ft.source.Err(); err != nil {
		return err
	}
	return ft.err
}
