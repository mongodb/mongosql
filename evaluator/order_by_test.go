package evaluator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestOrderByOperator(t *testing.T) {

	runTest := func(orderby *OrderBy, rows []bson.D, expectedRows []Values) {

		ctx := &ExecutionCtx{
			Schema: cfgOne,
			Db:     dbOne,
		}

		ts, err := NewBSONSource(ctx, tableOneName, nil)
		So(err, ShouldBeNil)

		source := &Project{
			source: ts,
			sExprs: SelectExpressions{
				SelectExpression{
					Column: &Column{tableOneName, "a", "a", "int"},
					Expr:   SQLFieldExpr{tableOneName, "a"},
				},
				SelectExpression{
					Column: &Column{tableOneName, "b", "b", "int"},
					Expr:   SQLFieldExpr{tableOneName, "b"},
				},
			},
		}

		orderby.source = source

		So(orderby.Open(ctx), ShouldBeNil)

		row := &Row{}

		i := 0

		for orderby.Next(row) {
			So(len(row.Data), ShouldEqual, 1)
			So(row.Data[0].Table, ShouldEqual, tableOneName)
			So(row.Data[0].Values, ShouldResemble, expectedRows[i])
			row = &Row{}
			i++
		}

		So(orderby.Close(), ShouldBeNil)
		So(orderby.Err(), ShouldBeNil)
	}

	Convey("An order by operator...", t, func() {

		data := []bson.D{
			bson.D{{"_id", 1}, {"a", 6}, {"b", 7}},
			bson.D{{"_id", 2}, {"a", 6}, {"b", 8}},
			bson.D{{"_id", 3}, {"a", 7}, {"b", 8}},
			bson.D{{"_id", 4}, {"a", 7}, {"b", 7}},
		}

		Convey("single sort keys should sort according to the direction specified", func() {

			Convey("asc", func() {

				keys := []orderByKey{
					{&SelectExpression{
						Expr: SQLFieldExpr{tableOneName, "a"},
					}, false, true, nil}}

				operator := &OrderBy{
					keys: keys,
				}

				expected := []Values{
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
				}

				runTest(operator, data, expected)

			})

			Convey("desc", func() {

				keys := []orderByKey{
					{&SelectExpression{
						Expr: SQLFieldExpr{tableOneName, "a"}}, false, false, nil},
				}

				operator := &OrderBy{
					keys: keys,
				}

				expected := []Values{
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
				}

				runTest(operator, data, expected)

			})

		})

		Convey("multiple sort keys should sort according to the direction specified", func() {

			Convey("asc + asc", func() {
				keys := []orderByKey{
					{&SelectExpression{
						Expr: SQLFieldExpr{tableOneName, "a"}}, false, true, nil},
					{&SelectExpression{
						Expr: SQLFieldExpr{tableOneName, "b"}}, false, true, nil},
				}

				expected := []Values{
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
				}

				operator := &OrderBy{
					keys: keys,
				}

				runTest(operator, data, expected)

			})

			Convey("asc + desc", func() {
				keys := []orderByKey{
					{&SelectExpression{
						Expr: SQLFieldExpr{tableOneName, "a"}}, false, true, nil},
					{&SelectExpression{
						Expr: SQLFieldExpr{tableOneName, "b"}}, false, false, nil},
				}

				operator := &OrderBy{
					keys: keys,
				}

				expected := []Values{
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
				}

				runTest(operator, data, expected)

			})

			Convey("desc + asc", func() {
				keys := []orderByKey{
					{&SelectExpression{
						Expr: SQLFieldExpr{tableOneName, "a"}}, false, false, nil},
					{&SelectExpression{
						Expr: SQLFieldExpr{tableOneName, "b"}}, false, true, nil},
				}

				operator := &OrderBy{
					keys: keys,
				}

				expected := []Values{
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
				}

				runTest(operator, data, expected)

			})

			Convey("desc + desc", func() {
				keys := []orderByKey{
					{&SelectExpression{
						Expr: SQLFieldExpr{tableOneName, "a"}}, false, false, nil},
					{&SelectExpression{
						Expr: SQLFieldExpr{tableOneName, "b"}}, false, false, nil},
				}

				operator := &OrderBy{
					keys: keys,
				}

				expected := []Values{
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
				}

				runTest(operator, data, expected)

			})
		})

	})
}
