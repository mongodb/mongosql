package evaluator

import (
	"context"
	"fmt"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/variable"
)

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

// A GroupByStage groups records according to one or more fields.
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
	keys []SQLExpr
}

// NewGroupByStage returns a new GroupByStage.
func NewGroupByStage(source PlanStage, keys []SQLExpr, projectedColumns ProjectedColumns) *GroupByStage {
	return &GroupByStage{
		source:           source,
		keys:             keys,
		projectedColumns: projectedColumns,
	}
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (gb *GroupByStage) Columns() (columns []*Column) {
	for _, projectedColumn := range gb.projectedColumns {
		columns = append(columns, projectedColumn.Column.clone())
	}
	return columns
}

// Collation returns the collation to use for comparisons.
func (gb *GroupByStage) Collation() *collation.Collation {
	return gb.source.Collation()
}

// GroupByIter returns grouped rows.
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

	keyBuffer *collation.KeyBuffer

	// channel on which to send rows derived from the final grouping
	outChan chan aggRowCtx

	ctx *ExecutionCtx

	cancelIter context.CancelFunc
}

// Open returns an iterator that returns results from executing this plan stage
// with the given ExecutionContext.
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
		keyBuffer:        &collation.KeyBuffer{},
		cancelIter:       func() {},
	}

	return iter, nil
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (gb *GroupByIter) Next(row *Row) bool {
	if !gb.grouped {
		if err := gb.createGroups(); err != nil {
			gb.err = err
			return false
		}
		ctx, cancel := context.WithCancel(gb.ctx.Context())
		gb.cancelIter = cancel
		gb.outChan = gb.iterChan(ctx)
	}

	rCtx, done := <-gb.outChan
	row.Data = rCtx.Row.Data

	return done
}

// Close closes the iterator, returning any error encountered while doing so.
func (gb *GroupByIter) Close() error {
	gb.keyBuffer.Reset()
	gb.cancelIter()
	return gb.source.Close()
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
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

		gbKey += gb.collation.KeyFromString(gb.keyBuffer, value.String())
	}

	return gbKey, nil
}

func (gb *GroupByIter) createGroups() error {

	gb.finalGrouping = orderedGroup{
		groups: make(map[string][]*Row, 0),
	}

	maxSize := gb.ctx.Variables().GetUInt64(variable.MongoDBMaxStageSize)
	size := uint64(0)

	// iterator source to create groupings
	r := &Row{}
	for gb.source.Next(r) {

		size += r.Data.Size()
		if maxSize != 0 && size > maxSize {
			return fmt.Errorf("aborted group by: maximum size per stage exceeded: limit is %d bytes", maxSize)
		}

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

		value := NewValue(
			projectedColumn.SelectID,
			projectedColumn.Database,
			projectedColumn.Table,
			projectedColumn.Name,
			v)

		row.Data = append(row.Data, value)
	}

	return row, nil
}

func (gb *GroupByIter) iterChan(ctx context.Context) chan aggRowCtx {
	ch := make(chan aggRowCtx)

	util.PanicSafeGo(func() {
	keyLoop:
		for _, key := range gb.finalGrouping.keys {
			v := gb.finalGrouping.groups[key]
			r, err := gb.evaluateProjectedColumns(v)
			if err != nil {
				gb.err = err
				close(ch)
				return
			}

			// check we have some matching data
			select {
			case ch <- aggRowCtx{*r, v}:
			case <-ctx.Done():
				break keyLoop
			}
		}
		close(ch)
	}, func(err interface{}) {
		gb.err = fmt.Errorf("%v", err)
	})

	return ch
}

func (gb *GroupByStage) clone() *GroupByStage {
	return &GroupByStage{
		source:           gb.source,
		keys:             gb.keys,
		projectedColumns: gb.projectedColumns,
	}
}
