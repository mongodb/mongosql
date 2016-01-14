package evaluator

import (
	"github.com/deafgoat/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func groupByTest(operator Operator, rows []bson.D, expectedRows [][]Values) {

	collectionOne.DropCollection()

	for _, row := range rows {
		So(collectionOne.Insert(row), ShouldBeNil)
	}

	ctx := &ExecutionCtx{
		Schema:  cfgOne,
		Db:      dbOne,
		Session: session,
	}

	So(operator.Open(ctx), ShouldBeNil)

	row := &Row{}

	i := 0

	for operator.Next(row) {
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

	So(operator.Close(), ShouldBeNil)
	So(operator.Err(), ShouldBeNil)
}

func TestGroupByOperator(t *testing.T) {

	Convey("A group by operator...", t, func() {

		data := []bson.D{
			bson.D{{"_id", 1}, {"a", 6}, {"b", 7}},
			bson.D{{"_id", 2}, {"a", 6}, {"b", 8}},
		}

		source := &Project{
			source: &TableScan{
				tableName: tableOneName,
			},
			sExprs: SelectExpressions{
				SelectExpression{
					Column: Column{tableOneName, "a", "a"},
					Expr:   SQLFieldExpr{tableOneName, "a"},
				},
				SelectExpression{
					Column: Column{tableOneName, "b", "b"},
					Expr:   SQLFieldExpr{tableOneName, "b"},
				},
			},
		}

		Convey("should return the right result when using an aggregation function", func() {

			sExprs := SelectExpressions{
				SelectExpression{
					Column: Column{tableOneName, "a", "a"},
					Expr:   SQLFieldExpr{tableOneName, "a"},
				},
				SelectExpression{
					Column: Column{"", "sum(b)", "sum(b)"},
					Expr: &SQLAggFunctionExpr{
						&sqlparser.FuncExpr{
							Name: []byte("sum"),
							Exprs: sqlparser.SelectExprs{
								&sqlparser.NonStarExpr{
									Expr: &sqlparser.ColName{[]byte("b"), []byte(tableOneName)},
								},
							},
						},
					},
				},
			}

			exprs := []SQLExpr{
				SQLFieldExpr{tableOneName, "a"},
			}

			operator := &GroupBy{
				sExprs:  sExprs,
				source:  source,
				exprs:   exprs,
				matcher: SQLBool(true),
			}

			expected := [][]Values{
				[]Values{
					{{"a", "a", SQLInt(6)}},
					{{"sum(b)", "sum(b)", SQLInt(15)}},
				},
			}

			groupByTest(operator, data, expected)

		})

	})
}
