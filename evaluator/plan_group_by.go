package evaluator

import (
	"context"
	"fmt"
	"sync"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/memory"
	"github.com/10gen/sqlproxy/internal/procutil"
)

// orderedGroup holds all the rows belonging to a given key in the groups
// and an slice of the keys for each group.
type orderedGroup struct {
	// groups is a map of group key to group members.
	groups map[string][]*Row
	// keys are the groups.
	keys []string
	// sizes holds the allocated memory each group uses.
	sizes map[string]uint64
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

// Children returns a slice of all the Node children of the Node.
func (gb GroupByStage) Children() []Node {
	out := make([]Node, len(gb.projectedColumns)+1+len(gb.keys))
	for i := range gb.projectedColumns {
		out[i] = gb.projectedColumns[i].Expr
	}
	out[len(gb.projectedColumns)] = gb.source
	for i := range gb.keys {
		out[i+len(gb.projectedColumns)+1] = gb.keys[i]
	}
	return out
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (gb *GroupByStage) ReplaceChild(i int, n Node) {
	if i < 0 || i > len(gb.projectedColumns)+len(gb.keys) {
		panicWithInvalidIndex("GroupByStage", i, len(gb.projectedColumns)+len(gb.keys))
	}
	if i == len(gb.projectedColumns) {
		gb.source = panicIfNotPlanStage("GroupByStage", n)
		return
	}
	if i < len(gb.projectedColumns) {
		gb.projectedColumns[i].Expr = panicIfNotSQLExpr("GroupByStage", n)
		return
	}
	gb.keys[i-1-len(gb.projectedColumns)] = panicIfNotSQLExpr("GroupByStage", n)
}

// NewGroupByStage returns a new GroupByStage.
func NewGroupByStage(source PlanStage,
	keys []SQLExpr,
	projectedColumns ProjectedColumns) *GroupByStage {
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
	cfg          *ExecutionConfig
	st           *ExecutionState
	source       RowIter
	stageMonitor memory.Monitor
	collation    *collation.Collation

	projectedColumns ProjectedColumns
	keys             []SQLExpr

	// grouped indicates if the source operator data has been grouped
	grouped bool

	// err holds any error encountered during processing
	err     error
	errLock sync.RWMutex

	// finalGrouping contains all grouped records and an ordered list of
	// the keys as read from the source operator
	finalGrouping orderedGroup

	keyBuffer *collation.KeyBuffer

	// channel on which to send rows derived from the final grouping
	outChan chan aggRowCtx

	cancelIter context.CancelFunc
}

// Open returns an iterator that returns results from executing this plan stage
// with the given ExecutionContext.
func (gb *GroupByStage) Open(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (RowIter, error) {
	sourceIter, err := gb.source.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}

	stageMonitor, err := cfg.memoryMonitor.CreateChild("GroupByStage", cfg.maxStageSize)
	if err != nil {
		return nil, err
	}

	iter := &GroupByIter{
		cfg:              cfg,
		st:               st.WithCollation(gb.Collation()),
		stageMonitor:     stageMonitor,
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
func (gb *GroupByIter) Next(ctx context.Context, row *Row) bool {
	var err error
	if !gb.grouped {
		if err = gb.createGroups(ctx); err != nil {
			gb.setError(err)
			return false
		}
		cancelCtx, cancel := context.WithCancel(ctx)
		gb.cancelIter = cancel
		gb.outChan = gb.iterChan(cancelCtx)
	}

	rCtx, done := <-gb.outChan
	row.Data = rCtx.Row.Data

	err = gb.stageMonitor.Exclude(row.Data.Size())
	if err != nil {
		gb.setError(err)
	}

	return err == nil && done
}

// Close closes the iterator, returning any error encountered while doing so.
func (gb *GroupByIter) Close() error {
	gb.keyBuffer.Reset()
	gb.cancelIter()

	err := gb.source.Close()
	if err != nil {
		return err
	}

	_, err = gb.stageMonitor.Clear()
	return err
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (gb *GroupByIter) Err() error {
	err := gb.source.Err()
	if err != nil {
		return err
	}
	gb.errLock.RLock()
	err = gb.err
	gb.errLock.RUnlock()
	return err
}

func (gb *GroupByIter) evaluateGroupByKey(ctx context.Context, row *Row) (string, error) {

	var gbKey string

	st := gb.st.WithRows(row)
	for _, key := range gb.keys {
		value, err := key.Evaluate(ctx, gb.cfg, st)
		if err != nil {
			return "", err
		}

		gbKey += gb.collation.KeyFromString(gb.keyBuffer, value.String())
	}

	return gbKey, nil
}

func (gb *GroupByIter) createGroups(ctx context.Context) error {

	gb.finalGrouping = orderedGroup{
		groups: make(map[string][]*Row),
		sizes:  make(map[string]uint64),
	}

	// iterator source to create groupings
	r := &Row{}
	for gb.source.Next(ctx, r) {

		err := gb.stageMonitor.Include(r.Data.Size())
		if err != nil {
			return err
		}

		key, err := gb.evaluateGroupByKey(ctx, r)
		if err != nil {
			return err
		}

		if gb.finalGrouping.groups[key] == nil {
			gb.finalGrouping.keys = append(gb.finalGrouping.keys, key)
			gb.finalGrouping.sizes[key] = 0
		}

		gb.finalGrouping.groups[key] = append(gb.finalGrouping.groups[key], r)
		gb.finalGrouping.sizes[key] += r.Data.Size()

		r = &Row{}
	}

	gb.grouped = true

	return gb.source.Err()
}

func (gb *GroupByIter) evaluateProjectedColumns(ctx context.Context, r []*Row) (*Row, error) {
	row := &Row{}

	st := gb.st.WithRows(r...)
	for _, projectedColumn := range gb.projectedColumns {

		v, err := projectedColumn.Expr.Evaluate(ctx, gb.cfg, st)
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

	procutil.PanicSafeGo(func() {
	keyLoop:
		for _, key := range gb.finalGrouping.keys {
			v := gb.finalGrouping.groups[key]
			r, err := gb.evaluateProjectedColumns(ctx, v)
			if err != nil {
				gb.setError(err)
				close(ch)
				return
			}

			size := gb.finalGrouping.sizes[key]
			if err = gb.stageMonitor.Release(size); err != nil {
				gb.setError(err)
				close(ch)
				return
			}
			if err = gb.stageMonitor.Acquire(r.Data.Size()); err != nil {
				gb.setError(err)
				close(ch)
				return
			}

			select {
			case ch <- aggRowCtx{*r, v}:
			case <-ctx.Done():
				break keyLoop
			}
		}
		close(ch)
	}, func(err interface{}) {
		gb.setError(fmt.Errorf("%v", err))
		close(ch)
	})

	return ch
}

func (gb *GroupByIter) setError(err error) {
	gb.errLock.Lock()
	gb.err = err
	gb.errLock.Unlock()
}

func (gb *GroupByStage) clone() PlanStage {
	return NewGroupByStage(gb.source.clone(), gb.keys, gb.projectedColumns)
}
