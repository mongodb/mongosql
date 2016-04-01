package evaluator

// Filter ensures that only rows matching a given criteria are
// returned.
type FilterStage struct {
	matcher     SQLExpr
	hasSubquery bool
	source      PlanStage
}

type FilterIter struct {
	matcher     SQLExpr
	hasSubquery bool
	execCtx     *ExecutionCtx
	source      Iter
	err         error
}

func (fs *FilterStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := fs.source.Open(ctx)
	if err != nil {
		return nil, err
	}
	return &FilterIter{
		matcher:     fs.matcher,
		hasSubquery: fs.hasSubquery,
		execCtx:     ctx,
		source:      sourceIter,
		err:         nil,
	}, nil
}

func (fs *FilterStage) OpFields() (columns []*Column) {
	return fs.source.OpFields()
}

func (fi *FilterIter) Next(row *Row) bool {
	var hasNext bool
	for {
		hasNext = fi.source.Next(row)

		if !hasNext {
			break
		}

		evalCtx := &EvalCtx{Rows{*row}, fi.execCtx}

		// add parent row(s) to this subquery's evaluation context
		if len(fi.execCtx.SrcRows) != 0 {

			bound := len(fi.execCtx.SrcRows) - 1

			for _, r := range fi.execCtx.SrcRows[:bound] {
				evalCtx.Rows = append(evalCtx.Rows, *r)
			}

			// avoid duplication since subquery row is most recently
			// appended and "*row"
			if !fi.hasSubquery {
				evalCtx.Rows = append(evalCtx.Rows, *fi.execCtx.SrcRows[bound])
			}

		}

		if fi.matcher != nil {
			m, err := Matches(fi.matcher, evalCtx)
			if err != nil {
				fi.err = err
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

func (fi *FilterIter) Close() error {
	return fi.source.Close()
}

func (fi *FilterIter) Err() error {
	if fi.err != nil {
		return fi.err
	}
	return fi.source.Err()
}

func (fs *FilterStage) clone() *FilterStage {
	return &FilterStage{
		source:      fs.source,
		matcher:     fs.matcher,
		hasSubquery: fs.hasSubquery,
	}
}
