package evaluator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestGroupByOperator(t *testing.T) {

	runTest := func(groupBy *GroupBy, rows []bson.D, expectedRows [][]Values) {

		ctx := &ExecutionCtx{
			Schema: cfgOne,
			Db:     dbOne,
		}

		ts, err := NewBSONSource(ctx, tableOneName, rows)
		So(err, ShouldBeNil)

		groupBy.source = &Project{
			source: ts,
			sExprs: SelectExpressions{
				SelectExpression{
					Column: &Column{tableOneName, "a", "a", "int"},
					Expr:   SQLColumnExpr{tableOneName, "a"},
				},
				SelectExpression{
					Column: &Column{tableOneName, "b", "b", "int"},
					Expr:   SQLColumnExpr{tableOneName, "b"},
				},
			},
		}

		So(groupBy.Open(ctx), ShouldBeNil)

		row := &Row{}

		i := 0

		for groupBy.Next(row) {
			So(len(row.Data), ShouldEqual, 2)
			aggregateTable := 1
			if row.Data[0].Table == "" {
				aggregateTable = 0
			}

			So(row.Data[aggregateTable].Table, ShouldEqual, "")
			So(row.Data[1-aggregateTable].Table, ShouldEqual, tableOneName)
			So(row.Data[1-aggregateTable].Values, ShouldResemble, expectedRows[i][0])
			So(row.Data[aggregateTable].Values, ShouldResemble, expectedRows[i][1])
			row = &Row{}
			i++
		}

		So(groupBy.Close(), ShouldBeNil)
		So(groupBy.Err(), ShouldBeNil)
	}

	Convey("A group by operator...", t, func() {

		data := []bson.D{
			bson.D{{"_id", 1}, {"a", 6}, {"b", 7}},
			bson.D{{"_id", 2}, {"a", 6}, {"b", 8}},
		}

		Convey("should return the right result when using an aggregation function", func() {

			sExprs := SelectExpressions{
				SelectExpression{
					Column: &Column{tableOneName, "a", "a", "int"},
					Expr:   SQLColumnExpr{tableOneName, "a"},
				},
				SelectExpression{
					Column: &Column{"", "sum(b)", "sum(b)", "int"},
					Expr: &SQLAggFunctionExpr{
						Name: "sum",
						Exprs: []SQLExpr{
							SQLColumnExpr{tableOneName, "b"},
						},
					},
				},
			}

			exprs := SelectExpressions{
				SelectExpression{
					Column: &Column{tableOneName, "a", "a", "int"},
					Expr:   SQLColumnExpr{tableOneName, "a"},
				},
			}

			operator := &GroupBy{
				selectExprs: sExprs,
				keyExprs:    exprs,
			}

			expected := [][]Values{
				[]Values{
					{{"a", "a", SQLInt(6)}},
					{{"sum(b)", "sum(b)", SQLInt(15)}},
				},
			}

			runTest(operator, data, expected)

		})

	})
}
