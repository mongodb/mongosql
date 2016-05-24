package evaluator

// Project ensures that referenced columns - e.g. those used to
// support ORDER BY and GROUP BY clauses - aren't included in
// the final result set.
type ProjectStage struct {
	// sExprs holds the columns and/or expressions used in
	// the pipeline
	sExprs SelectExpressions

	// source is the operator that provides the data to project
	source PlanStage
}

// NewProjectStage creates a new project stage.
func NewProjectStage(source PlanStage, sExprs ...SelectExpression) *ProjectStage {
	return &ProjectStage{
		source: source,
		sExprs: sExprs,
	}
}

type ProjectIter struct {
	source Iter

	sExprs SelectExpressions

	// err holds any error that may have occurred during processing
	err error

	// ctx is the current execution context
	ctx *ExecutionCtx
}

var systemVars = map[string]SQLValue{
	"max_allowed_packet": SQLInt(4194304),
}

func (pj *ProjectStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := pj.source.Open(ctx)
	if err != nil {
		return nil, err
	}

	return &ProjectIter{
		sExprs: pj.sExprs,
		ctx:    ctx,
		source: sourceIter,
	}, nil

}

func (pj *ProjectIter) Next(r *Row) bool {

	hasNext := pj.source.Next(r)

	if !hasNext {
		return false
	}

	evalCtx := &EvalCtx{
		Rows:    Rows{*r},
		ExecCtx: pj.ctx,
	}

	values := Values{}
	for _, projectedColumn := range pj.sExprs {
		v, err := projectedColumn.Expr.Evaluate(evalCtx)
		if err != nil {
			pj.err = err
			hasNext = false
		}
		value := Value{
			Table: projectedColumn.Table,
			Name:  projectedColumn.Name,
			Data:  v,
		}

		values = append(values, value)

	}
	r.Data = values

	return true
}

func (pj *ProjectStage) OpFields() (columns []*Column) {
	for _, expr := range pj.sExprs {
		column := &Column{
			Name:      expr.Name,
			View:      expr.View,
			Table:     expr.Table,
			SQLType:   expr.SQLType,
			MongoType: expr.MongoType,
		}
		columns = append(columns, column)
	}

	if len(columns) == 0 {
		columns = pj.source.OpFields()
	}

	return columns
}

func (pj *ProjectIter) Close() error {
	return pj.source.Close()
}

func (pj *ProjectIter) Err() error {
	if err := pj.source.Err(); err != nil {
		return err
	}
	return pj.err
}

func (pj *ProjectStage) clone() *ProjectStage {
	return &ProjectStage{
		source: pj.source,
		sExprs: pj.sExprs,
	}
}
