package evaluator

// Filter ensures that only rows matching a given criteria are
// returned.
type FilterStage struct {
	matcher         SQLExpr
	source          PlanStage
	requiredColumns []SQLExpr
}

func NewFilterStage(source PlanStage, predicate SQLExpr, reqCols []SQLExpr) *FilterStage {
	return &FilterStage{
		source:          source,
		matcher:         predicate,
		requiredColumns: reqCols,
	}
}

type FilterIter struct {
	matcher SQLExpr
	execCtx *ExecutionCtx
	source  Iter
	err     error
}

func (fs *FilterStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := fs.source.Open(ctx)
	if err != nil {
		return nil, err
	}
	return &FilterIter{
		matcher: fs.matcher,
		execCtx: ctx,
		source:  sourceIter,
		err:     nil,
	}, nil
}

func (fs *FilterStage) Columns() (columns []*Column) {
	return fs.source.Columns()
}

func (fi *FilterIter) Next(row *Row) bool {
	var hasMatch, hasNext bool

	for {

		hasNext = fi.source.Next(row)

		if !hasNext {
			break
		}

		if fi.matcher == nil {
			break
		}

		evalCtx := NewEvalCtx(fi.execCtx, row)

		hasMatch, fi.err = Matches(fi.matcher, evalCtx)
		if fi.err != nil {
			return false
		}

		if hasMatch {
			break
		}

		row.Data = nil
	}

	if fi.matcher != nil && !hasMatch {
		row.Data = nil
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
		source:          fs.source,
		matcher:         fs.matcher,
		requiredColumns: fs.requiredColumns,
	}
}
