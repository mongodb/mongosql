package evaluator

import (
	"github.com/erh/mixer/sqlparser"
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
		Config:  cfgOne,
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

		source := &Select{
			source: &TableScan{
				tableName: tableOneName,
			},
			sExprs: SelectExpressions{
				SelectExpression{
					Column: Column{tableOneName, "a", "a", false},
					Expr:   &sqlparser.ColName{[]byte("a"), []byte(tableOneName)},
				},
				SelectExpression{
					Column: Column{tableOneName, "b", "b", false},
					Expr:   &sqlparser.ColName{[]byte("b"), []byte(tableOneName)},
				},
			},
		}

		Convey("should return the right result when using an aggregation function", func() {

			sExprs := SelectExpressions{
				SelectExpression{
					Column: Column{tableOneName, "a", "a", false},
					Expr:   &sqlparser.ColName{[]byte("a"), []byte(tableOneName)},
				},
				SelectExpression{
					Column: Column{"", "sum(b)", "sum(b)", false},
					Expr: &sqlparser.FuncExpr{
						Name: []byte("sum"),
						Exprs: sqlparser.SelectExprs{
							&sqlparser.NonStarExpr{
								Expr: &sqlparser.ColName{[]byte("b"), []byte(tableOneName)},
							},
						},
					},
				},
			}

			exprs := []sqlparser.Expr{
				&sqlparser.ColName{[]byte("a"), []byte(tableOneName)},
			}

			matcher := &NoopMatcher{}

			operator := &GroupBy{
				sExprs:  sExprs,
				source:  source,
				exprs:   exprs,
				matcher: matcher,
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
