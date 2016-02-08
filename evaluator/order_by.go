package evaluator

import (
	"fmt"
	"sort"

	"gopkg.in/mgo.v2/bson"
)

// OrderBy sorts records according to one or more keys.
type OrderBy struct {
	// source is the operator that provides the data to order
	source Operator

	// keys holds the SQLExpr(s) to order by. For example, in
	// select a, count(b) from foo group by a order by count(b)
	// keys will hold the SQLValue for 'count(b)'. For multiple
	// order by criteria, they are stored in the same order within
	// the keys slice.
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
	expr      *SelectExpression
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

func (ob *OrderBy) sortRows() (orderByRows, error) {
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
		rows, err := ob.sortRows()
		if err != nil {
			ob.err = err
			return false
		}
		ob.outChan = ob.iterChan(rows)
	}

	r, done := <-ob.outChan
	row.Data = r.Data

	return done
}

func (ob *OrderBy) Close() error {
	return ob.source.Close()
}

func (ob *OrderBy) Err() error {
	if err := ob.source.Err(); err != nil {
		return err
	}

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

func orderValue(key orderByKey) (SQLValue, error) {

	expr := key.expr

	if key.isAggFunc {

		row := key.evalCtx.Rows[0]

		value, ok := row.GetField(expr.Table, expr.Name)
		if ok {
			return NewSQLValue(value, "")
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

		cmp, err := leftVal.CompareTo(rightVal)
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

///////////////
//Optimization
///////////////

func (v *optimizer) visitOrderBy(orderBy *OrderBy) (Operator, error) {

	sa, ts, ok := canPushDown(orderBy.source)
	if !ok {
		return orderBy, nil
	}

	sort := bson.D{}

	for _, key := range orderBy.keys {

		var tableName, columnName string

		switch typedE := key.expr.Expr.(type) {
		case SQLColumnExpr:
			tableName, columnName = typedE.tableName, typedE.columnName
		default:
			// MongoDB only allows sorting by a field, so pushing down a
			// non-field requires it to be pre-calculated by a previous
			// push down. If it has been pre-calculated, then it will
			// exist in the mapping registry. Otherwise, it won't, and
			// the order by cannot be pushed down.
			columnName = typedE.String()
		}

		fieldName, ok := ts.mappingRegistry.lookupFieldName(tableName, columnName)
		if !ok {
			return orderBy, nil
		}
		direction := 1
		if !key.ascending {
			direction = -1
		}
		sort = append(sort, bson.DocElem{fieldName, direction})
	}

	pipeline := ts.pipeline
	pipeline = append(pipeline, bson.D{{"$sort", sort}})

	ts = ts.WithPipeline(pipeline)
	sa = sa.WithSource(ts)

	return sa, nil
}
