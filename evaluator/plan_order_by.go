package evaluator

import (
	"fmt"
	"sort"

	"github.com/10gen/sqlproxy/collation"
)

// OrderBy sorts records according to one or more keys.
type OrderByStage struct {
	source          PlanStage
	terms           []*orderByTerm
	requiredColumns []SQLExpr
}

// NewOrderByStage returns a new order by stage.
func NewOrderByStage(source PlanStage, reqCols []SQLExpr, terms ...*orderByTerm) *OrderByStage {
	return &OrderByStage{
		source:          source,
		terms:           terms,
		requiredColumns: reqCols,
	}
}

type OrderByIter struct {
	source Iter

	collation *collation.Collation

	terms []*orderByTerm

	// channel on which to send sorted data
	outChan chan Values

	// sorted indicates if the source operator data has been sorted
	sorted bool

	ctx *ExecutionCtx

	// err holds any error encountered during processing
	err error
}

type orderByTerm struct {
	expr      SQLExpr
	ascending bool
}

func (t *orderByTerm) clone() *orderByTerm {
	return &orderByTerm{
		expr:      t.expr,
		ascending: t.ascending,
	}
}

type orderByRow struct {
	// terms contains the terms used to create the termValues. Mostly, these are still here
	// for context and to provide the direction of the sort.
	terms []*orderByTerm

	// termValues hold the evaluated values of the terms.
	termValues []SQLValue

	// data holds the raw data that was evaluated.
	data Values
}

type orderByRows []orderByRow

func (ob *OrderByStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := ob.source.Open(ctx)
	if err != nil {
		return nil, err
	}

	iter := &OrderByIter{
		source:    sourceIter,
		terms:     ob.terms,
		ctx:       ctx,
		collation: ob.Collation(),
	}

	return iter, nil
}

func (ob *OrderByIter) Next(row *Row) bool {
	if !ob.sorted {
		rows, err := ob.sortRows()
		if err != nil {
			ob.err = err
			return false
		}
		ob.outChan = iterChan(rows)
	}

	data, done := <-ob.outChan
	row.Data = data

	return done
}

func (ob *OrderByIter) sortRows() (orderByRows, error) {
	rows := orderByRows{}

	row := &Row{}

	for ob.source.Next(row) {

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
		rows = append(rows, obRow)
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

	return rows, err
}

func iterChan(rows orderByRows) chan Values {
	ch := make(chan Values)

	go func() {
		for _, row := range rows {
			ch <- row.data
		}
		close(ch)
	}()

	return ch
}

func (ob *OrderByIter) Close() error {
	return ob.source.Close()
}

func (ob *OrderByIter) Err() error {
	if err := ob.source.Err(); err != nil {
		return err
	}

	return ob.err
}

func (ob *OrderByStage) Columns() (columns []*Column) {
	return ob.source.Columns()
}

func (ob *OrderByStage) Collation() *collation.Collation {
	return ob.source.Collation()
}

func (ob *OrderByStage) clone() *OrderByStage {
	return &OrderByStage{
		source:          ob.source,
		terms:           ob.terms,
		requiredColumns: ob.requiredColumns,
	}
}

//
// helper functions to order data
//
func (rows orderByRows) Len() int {
	return len(rows)
}

func (rows orderByRows) Swap(i, j int) {
	rows[i], rows[j] = rows[j], rows[i]
}

func (rows orderByRows) Less(i, j int) bool {

	r1 := rows[i]
	r2 := rows[j]

	for i, t := range r1.terms {

		left := r1.termValues[i]
		right := r2.termValues[i]

		cmp, err := CompareTo(left, right)
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
