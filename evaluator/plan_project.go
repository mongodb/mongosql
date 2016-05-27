package evaluator

// ProjectStage handles taking sourced rows and projecting them into a different shape.
type ProjectStage struct {
	// projectedColumns holds the columns and/or expressions used in
	// the pipeline
	projectedColumns ProjectedColumns

	// source is the operator that provides the data to project
	source PlanStage
}

// NewProjectStage creates a new project stage.
func NewProjectStage(source PlanStage, projectedColumns ...ProjectedColumn) *ProjectStage {
	return &ProjectStage{
		source:           source,
		projectedColumns: projectedColumns,
	}
}

type ProjectIter struct {
	source Iter

	projectedColumns ProjectedColumns

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
		projectedColumns: pj.projectedColumns,
		ctx:              ctx,
		source:           sourceIter,
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
	for _, projectedColumn := range pj.projectedColumns {
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

func (pj *ProjectStage) Columns() (columns []*Column) {
	for _, projectedColumn := range pj.projectedColumns {
		column := &Column{
			Name:      projectedColumn.Name,
			Table:     projectedColumn.Table,
			SQLType:   projectedColumn.SQLType,
			MongoType: projectedColumn.MongoType,
		}
		columns = append(columns, column)
	}

	if len(columns) == 0 {
		columns = pj.source.Columns()
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
		source:           pj.source,
		projectedColumns: pj.projectedColumns,
	}
}
