package evaluator

import (
	"fmt"
	"sort"

	"github.com/10gen/sqlproxy/schema"
)

// OrderBy sorts records according to one or more keys.
type OrderByStage struct {
	// source is the operator that provides the data to order
	source PlanStage

	// keys holds the SQLExpr(s) to order by. For example, in
	// select a, count(b) from foo group by a order by count(b)
	// keys will hold the SQLValue for 'count(b)'. For multiple
	// order by criteria, they are stored in the same order within
	// the keys slice.
	keys []orderByKey
}

type OrderByIter struct {
	keys []orderByKey

	source Iter

	// channel on which to send sorted rows
	outChan chan Row

	// sorted indicates if the source operator data has been sorted
	sorted bool

	ctx *ExecutionCtx

	// err holds any error encountered during processing
	err error
}

type orderByKey struct {
	expr      *SelectExpression
	isAggFunc bool
	ascending bool
	evalCtx   *EvalCtx
}

func (k orderByKey) clone() orderByKey {
	return orderByKey{
		expr:      k.expr,
		isAggFunc: k.isAggFunc,
		ascending: k.ascending,
	}
}

type orderByRow struct {
	keys []orderByKey
	data Row
}

type orderByRows []orderByRow

func (ob *OrderByStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := ob.source.Open(ctx)
	if err != nil {
		return nil, err
	}
	return &OrderByIter{source: sourceIter, keys: ob.keys, ctx: ctx}, nil
}

func (ob *OrderByIter) evaluateOrderByKeys(row *Row) []orderByKey {

	keys := make([]orderByKey, 0, len(ob.keys))

	for _, key := range ob.keys {
		key.evalCtx = &EvalCtx{Rows: Rows{*row}}

		// for aggregation functions, we set the context in the
		// preceding GROUP BY operator
		if key.isAggFunc && len(ob.ctx.GroupRows) != 0 {
			key.evalCtx = &EvalCtx{Rows: ob.ctx.GroupRows}
		}

		keys = append(keys, key)
	}

	return keys
}

func (ob *OrderByIter) sortRows() (orderByRows, error) {
	rows := orderByRows{}

	row := &Row{}

	for ob.source.Next(row) {
		obRow := orderByRow{ob.evaluateOrderByKeys(row), *row}
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

	sort.Sort(rows)

	ob.sorted = true

	return rows, err

}

func iterChan(rows orderByRows) chan Row {
	ch := make(chan Row)

	go func() {
		for _, row := range rows {
			ch <- row.data
		}
		close(ch)
	}()

	return ch
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

	r, done := <-ob.outChan
	row.Data = r.Data

	return done
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

func (ob *OrderByStage) OpFields() (columns []*Column) {
	return ob.source.OpFields()
}

func (ob *OrderByStage) clone() *OrderByStage {
	return &OrderByStage{
		source: ob.source,
		keys:   ob.keys,
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

func orderValue(key orderByKey) (SQLValue, error) {

	expr := key.expr

	if key.isAggFunc {

		row := key.evalCtx.Rows[0]

		value, ok := row.GetField(expr.Table, expr.Name)
		if ok {
			return NewSQLValue(value, schema.SQLNone, schema.MongoNone)
		}
	}

	return expr.Expr.Evaluate(key.evalCtx)
}

func (rows orderByRows) Less(i, j int) bool {

	r1 := rows[i]
	r2 := rows[j]

	for i := range r1.keys {

		left := r1.keys[i]
		right := r2.keys[i]

		leftVal, err := orderValue(left)
		if err != nil {
			panic(err)
		}

		rightVal, err := orderValue(right)
		if err != nil {
			panic(err)
		}

		cmp, err := CompareTo(leftVal, rightVal)
		if err != nil {
			panic(err)
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
