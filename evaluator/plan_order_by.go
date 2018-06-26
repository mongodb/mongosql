package evaluator

import (
	"context"
	"fmt"
	"sort"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/memory"
	"github.com/10gen/sqlproxy/internal/util"
)

// An OrderByStage sorts records according to one or more keys.
type OrderByStage struct {
	source PlanStage
	terms  []*OrderByTerm
}

// NewOrderByStage returns a new order by stage.
func NewOrderByStage(source PlanStage, terms ...*OrderByTerm) *OrderByStage {
	return &OrderByStage{
		source: source,
		terms:  terms,
	}
}

// OrderByIter returns ordered rows.
type OrderByIter struct {
	source Iter

	stageMonitor *memory.Monitor

	collation *collation.Collation

	terms []*OrderByTerm

	// channel on which to send sorted data
	outChan chan Values

	// sorted indicates if the source operator data has been sorted
	sorted bool

	ctx *ExecutionCtx

	// err holds any error encountered during processing
	err error

	errChan chan error

	cancelIter context.CancelFunc
}

// OrderByTerm represents an expression by which rows should be ordered
// and the order in which those rows should be sorted.
type OrderByTerm struct {
	expr      SQLExpr
	ascending bool
}

// NewOrderByTerm returns a new OrderByTerm Struct.
func NewOrderByTerm(expr SQLExpr, ascending bool) *OrderByTerm {
	return &OrderByTerm{
		expr:      expr,
		ascending: ascending,
	}
}

type orderByRow struct {
	// terms contains the terms used to create the termValues. Mostly, these are still here
	// for context and to provide the direction of the sort.
	terms []*OrderByTerm

	// termValues hold the evaluated values of the terms.
	termValues []SQLValue

	// data holds the raw data that was evaluated.
	data Values
}

type orderByRows struct {
	rows      []orderByRow
	collation *collation.Collation
}

// Open returns an iterator that returns results from executing this plan stage
// with the given ExecutionContext.
func (ob *OrderByStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := ob.source.Open(ctx)
	if err != nil {
		return nil, err
	}

	stageMonitor, err := newStageMemoryMonitor(ctx, "OrderByStage")
	if err != nil {
		return nil, err
	}

	iter := &OrderByIter{
		source:       sourceIter,
		stageMonitor: stageMonitor,
		terms:        ob.terms,
		ctx:          ctx,
		collation:    ob.Collation(),
		errChan:      make(chan error),
		cancelIter:   func() {},
	}

	return iter, nil
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (ob *OrderByIter) Next(row *Row) bool {
	if !ob.sorted {
		rows, err := ob.sortRows()
		if err != nil {
			ob.err = err
			return false
		}
		ctx, cancel := context.WithCancel(ob.ctx.Context())
		ob.cancelIter = cancel
		ob.startIterChan(ctx, rows)
	}

	select {
	case data, done := <-ob.outChan:
		row.Data = data
		ob.err = ob.stageMonitor.Exclude(row.Data.Size())
		return ob.err == nil && done
	case <-ob.ctx.Context().Done():
		ob.err = ob.ctx.Context().Err()
		return false
	case err := <-ob.errChan:
		ob.err = err
		return false
	}

}

func (ob *OrderByIter) sortRows() ([]orderByRow, error) {
	rows := orderByRows{
		collation: ob.collation,
	}

	row := &Row{}
	for ob.source.Next(row) {
		err := ob.stageMonitor.Include(row.Data.Size())
		if err != nil {
			return nil, err
		}

		ctx := NewEvalCtx(ob.ctx, ob.collation, row)
		var values []SQLValue
		for _, t := range ob.terms {
			v, err := t.expr.Evaluate(ctx)
			if err != nil {
				return nil, err
			}

			values = append(values, v)
		}

		obRow := orderByRow{ob.terms, values, row.Data}
		rows.rows = append(rows.rows, obRow)
		row = &Row{}
	}

	err := ob.source.Err()

	defer func() {
		if err == nil {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	sort.Stable(rows)

	ob.sorted = true

	return rows.rows, err
}

// Close closes the iterator, returning any error encountered while doing so.
func (ob *OrderByIter) Close() error {
	ob.cancelIter()
	err := ob.source.Close()
	if err != nil {
		return err
	}
	_, err = ob.stageMonitor.Clear()
	return err
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (ob *OrderByIter) Err() error {
	if err := ob.source.Err(); err != nil {
		return err
	}

	return ob.err
}

func (ob *OrderByIter) startIterChan(ctx context.Context, rows []orderByRow) {
	ob.outChan = make(chan Values)
	util.PanicSafeGo(func() {
	rowLoop:
		for _, row := range rows {
			select {
			case ob.outChan <- row.data:
			case <-ctx.Done():
				break rowLoop
			}
		}
		close(ob.outChan)
	}, func(err interface{}) {
		ob.errChan <- fmt.Errorf("%v", err)
	})
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (ob *OrderByStage) Columns() (columns []*Column) {
	return ob.source.Columns()
}

// Collation returns the collation to use for comparisons.
func (ob *OrderByStage) Collation() *collation.Collation {
	return ob.source.Collation()
}

//
// helper functions to order data
//
func (rows orderByRows) Len() int {
	return len(rows.rows)
}

func (rows orderByRows) Swap(i, j int) {
	rows.rows[i], rows.rows[j] = rows.rows[j], rows.rows[i]
}

func (rows orderByRows) Less(i, j int) bool {

	r1 := rows.rows[i]
	r2 := rows.rows[j]

	for i, t := range r1.terms {

		left := r1.termValues[i]
		right := r2.termValues[i]

		cmp, err := CompareTo(left, right, rows.collation)

		if err != nil {
			panic(err)
		}

		if !t.ascending {
			cmp = -cmp
		}

		if cmp != 0 {
			return cmp == -1
		}

	}

	return false
}

func (ob *OrderByStage) clone() PlanStage {
	newTerms := make([]*OrderByTerm, len(ob.terms))
	for i, term := range ob.terms {
		newTerms[i] = term.clone()
	}
	return NewOrderByStage(ob.source.clone(), newTerms...)
}

func (obt *OrderByTerm) clone() *OrderByTerm {
	return NewOrderByTerm(obt.expr, obt.ascending)
}
