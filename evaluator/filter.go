package evaluator

type Filter struct {
	// err holds any error that may have occurred during processing
	err error

	// source holds the source for this select statement
	source Operator

	// matcher is used to filter results gotten from the source operator
	matcher Matcher

	// ctx is the current execution context
	ctx *ExecutionCtx
}

func (ft *Filter) Open(ctx *ExecutionCtx) error {
	ft.ctx = ctx
	return ft.source.Open(ctx)
}

func (ft *Filter) Next(row *Row) bool {

	var hasNext bool

	for {

		hasNext = ft.source.Next(row)

		evalCtx := &EvalCtx{[]Row{*row}, ft.ctx}

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

		if !hasNext {
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
	return ft.err
}
