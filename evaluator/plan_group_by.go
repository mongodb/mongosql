package evaluator

import "github.com/10gen/sqlproxy/collation"

// orderedGroup holds all the rows belonging to a given key in the groups
// and an slice of the keys for each group.
type orderedGroup struct {
	groups map[string][]*Row
	keys   []string
}

type aggRowCtx struct {
	Row Row
	Ctx []*Row
}

type GroupByStage struct {
	// projectedColumns holds the ProjectedColumn that should
	// be present in the result of a grouping. This will
	// include ProjectedColumns for aggregates that might
	// not be projected, but are required for further
	// processing, such as when ordering by an aggregate.
	projectedColumns ProjectedColumns

	// source is the operator that provides the data to group
	source PlanStage

	// keys holds the expression(s) to group by. For example, in
	// select a, count(b) from foo group by a,
	// keys will hold the parsed column name 'a'.
	keys            []SQLExpr
	requiredColumns []SQLExpr
}

func NewGroupByStage(source PlanStage, keys []SQLExpr, projectedColumns ProjectedColumns, reqCols []SQLExpr) *GroupByStage {
	return &GroupByStage{
		source:           source,
		keys:             keys,
		projectedColumns: projectedColumns,
		requiredColumns:  reqCols,
	}
}

func (gb *GroupByStage) Columns() (columns []*Column) {
	for _, expr := range gb.projectedColumns {
		column := &Column{
			Name:      expr.Name,
			Table:     expr.Table,
			SQLType:   expr.SQLType,
			MongoType: expr.MongoType,
		}
		columns = append(columns, column)
	}
	return columns
}

func (gb *GroupByStage) Collation() *collation.Collation {
	return gb.source.Collation()
}

// GroupBy groups records according to one or more fields.
type GroupByIter struct {
	source    Iter
	collation *collation.Collation

	projectedColumns ProjectedColumns
	keys             []SQLExpr

	// grouped indicates if the source operator data has been grouped
	grouped bool

	// err holds any error encountered during processing
	err error

	// finalGrouping contains all grouped records and an ordered list of
	// the keys as read from the source operator
	finalGrouping orderedGroup

	// channel on which to send rows derived from the final grouping
	outChan chan aggRowCtx

	ctx *ExecutionCtx
}

func (gb *GroupByStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := gb.source.Open(ctx)
	if err != nil {
		return nil, err
	}

	iter := &GroupByIter{
		ctx:              ctx,
		source:           sourceIter,
		projectedColumns: gb.projectedColumns,
		keys:             gb.keys,
		collation:        gb.Collation(),
	}

	return iter, nil
}

func (gb *GroupByIter) Next(row *Row) bool {
	if !gb.grouped {
		if err := gb.createGroups(); err != nil {
			gb.err = err
			return false
		}
		gb.outChan = gb.iterChan()
	}

	rCtx, done := <-gb.outChan
	row.Data = rCtx.Row.Data

	return done
}

func (gb *GroupByIter) Close() error {
	return gb.source.Close()
}

func (gb *GroupByIter) Err() error {
	if err := gb.source.Err(); err != nil {
		return err
	}
	return gb.err
}

func (gb *GroupByIter) evaluateGroupByKey(row *Row) (string, error) {

	var gbKey string

	evalCtx := NewEvalCtx(gb.ctx, gb.collation, row)
	for _, key := range gb.keys {

		value, err := key.Evaluate(evalCtx)
		if err != nil {
			return "", err
		}

		gbKey += value.String()
	}

	return gbKey, nil
}

func (gb *GroupByIter) createGroups() error {

	gb.finalGrouping = orderedGroup{
		groups: make(map[string][]*Row, 0),
	}

	r := &Row{}

	// iterator source to create groupings
	for gb.source.Next(r) {

		key, err := gb.evaluateGroupByKey(r)
		if err != nil {
			return err
		}

		if gb.finalGrouping.groups[key] == nil {
			gb.finalGrouping.keys = append(gb.finalGrouping.keys, key)
		}

		gb.finalGrouping.groups[key] = append(gb.finalGrouping.groups[key], r)

		r = &Row{}
	}

	gb.grouped = true

	return gb.source.Err()
}

func (gb *GroupByIter) evaluateProjectedColumns(r []*Row) (*Row, error) {

	row := &Row{}
	evalCtx := NewEvalCtx(gb.ctx, gb.collation, r...)

	for _, projectedColumn := range gb.projectedColumns {

		v, err := projectedColumn.Expr.Evaluate(evalCtx)
		if err != nil {
			return nil, err
		}

		value := Value{
			SelectID: projectedColumn.SelectID,
			Table:    projectedColumn.Table,
			Name:     projectedColumn.Name,
			Data:     v,
		}

		row.Data = append(row.Data, value)
	}

	return row, nil
}

func (gb *GroupByIter) iterChan() chan aggRowCtx {
	ch := make(chan aggRowCtx)

	go func() {
		for _, key := range gb.finalGrouping.keys {
			v := gb.finalGrouping.groups[key]

			r, err := gb.evaluateProjectedColumns(v)
			if err != nil {
				gb.err = err
				close(ch)
				return
			}

			// check we have some matching data
			if len(r.Data) != 0 {
				ch <- aggRowCtx{*r, v}
			}
		}
		close(ch)
	}()
	return ch
}

func (gb *GroupByStage) clone() *GroupByStage {
	return &GroupByStage{
		source:           gb.source,
		keys:             gb.keys,
		projectedColumns: gb.projectedColumns,
		requiredColumns:  gb.requiredColumns,
	}
}
