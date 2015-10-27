package evaluator

import (
	"sort"
)

type OrderBy struct {
	// source is the operator that provides the data to group
	source Operator

	// keys holds the expression key(s) to order by. For example, in
	// select a, count(b) from foo group by a order by count(b)
	// keys will hold the expression 'count(b)'
	keys []orderByKey

	// channel on which to send sorted rows
	outChan chan Row

	// sorted indicates if the source operator data has been sorted
	sorted bool

	ctx *ExecutionCtx

	// err holds any error encountered during processing
	err error
}

type orderByKey struct {
	value     SQLValue
	isAggFunc bool
	ascending bool
	evalCtx   *EvalCtx
}

type orderByRow struct {
	keys []orderByKey
	data Row
}

type orderByRows []orderByRow

func (ob *OrderBy) Open(ctx *ExecutionCtx) error {
	return ob.init(ctx)
}

func (ob *OrderBy) init(ctx *ExecutionCtx) error {
	ob.ctx = ctx
	return ob.source.Open(ctx)
}

func (ob *OrderBy) evaluateOrderByKeys(row *Row) []orderByKey {

	keys := make([]orderByKey, 0, len(ob.keys))

	for _, key := range ob.keys {

		key.evalCtx = &EvalCtx{Rows: []Row{*row}}

		// for aggregation functions, we set the context in the
		// preceding GROUP BY operator
		if key.isAggFunc {
			key.evalCtx = &EvalCtx{Rows: ob.ctx.Rows}
		}

		keys = append(keys, key)
	}

	return keys
}

func (ob *OrderBy) sortGroups() (orderByRows, error) {
	rows := orderByRows{}

	row := &Row{}

	for ob.source.Next(row) {
		obRow := orderByRow{ob.evaluateOrderByKeys(row), *row}
		rows = append(rows, obRow)
		row = &Row{}
	}

	sort.Sort(rows)

	ob.sorted = true

	return rows, ob.source.Err()

}

func (ob *OrderBy) iterChan(rows orderByRows) chan Row {
	ch := make(chan Row)

	go func() {
		for _, row := range rows {
			ch <- row.data
		}
		close(ch)
	}()

	return ch
}

func (ob *OrderBy) Next(row *Row) bool {

	if !ob.sorted {
		g, err := ob.sortGroups()
		if err != nil {
			ob.err = err
			return false
		}
		ob.outChan = ob.iterChan(g)
	}

	r, done := <-ob.outChan
	row.Data = r.Data

	return done
}

func (ob *OrderBy) Close() error {
	return ob.source.Close()
}

func (ob *OrderBy) Err() error {
	return ob.err
}

func (ob *OrderBy) OpFields() (columns []*Column) {
	return ob.source.OpFields()
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

	for i := range r1.keys {
		left := r1.keys[i]
		right := r2.keys[i]

		eval, err := left.value.Evaluate(left.evalCtx)
		if err != nil {
			return false
		}

		cmp, err := eval.CompareTo(right.evalCtx, right.value)
		if err != nil {
			return false
		}

		if !left.ascending {
			cmp = -cmp
		}

		if cmp != 0 {
			return cmp == -1
		}

	}

	return false
}
