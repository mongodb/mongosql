package evaluator

import (
	"fmt"
	"github.com/deafgoat/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"testing"
)

var (
	_ fmt.Stringer = nil
)

func havingTest(operator Operator, rows []bson.D, expectedRows Values) {

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

	for operator.Next(row) {
		So(len(row.Data), ShouldEqual, 2)

		aggregateTable := 1
		if row.Data[0].Table == "" {
			aggregateTable = 0
		}

		So(row.Data[aggregateTable].Table, ShouldEqual, "")
		So(row.Data[1-aggregateTable].Table, ShouldEqual, tableOneName)
		So(row.Data[1-aggregateTable].Values, ShouldResemble, expectedRows)
		So(row.Data[aggregateTable].Values, ShouldResemble, expectedRows)
		row = &Row{}
	}

	So(operator.Close(), ShouldBeNil)
	So(operator.Err(), ShouldBeNil)
}

func TestHavingOperator(t *testing.T) {

	Convey("A having operator...", t, func() {

		data := []bson.D{
			{{"_id", SQLInt(1)}, {"a", SQLInt(6)}, {"b", SQLInt(7)}},
			{{"_id", SQLInt(2)}, {"a", SQLInt(6)}, {"b", SQLInt(8)}},
		}

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

		source := &Project{
			source: &TableScan{
				tableName: tableOneName,
			},
			sExprs: sExprs,
		}

		Convey("should return the right result when having clause is true", func() {

			matchExpr := &sqlparser.ComparisonExpr{
				Operator: sqlparser.AST_GT,
				Left: &sqlparser.FuncExpr{
					Name: []byte("sum"),
					Exprs: sqlparser.SelectExprs{
						&sqlparser.NonStarExpr{
							Expr: &sqlparser.ColName{[]byte("b"), []byte(tableOneName)},
						},
					},
				},
				Right: sqlparser.NumVal(strconv.FormatFloat(3.0, 'E', -1, 64)),
			}

			matcher, err := NewSQLExpr(matchExpr)
			So(err, ShouldBeNil)

			operator := &Having{
				sExprs:  sExprs,
				source:  source,
				matcher: matcher,
			}

			expected := Values{{"a", "a", SQLInt(6)}, {"sum(b)", "sum(b)", SQLInt(15)}}

			havingTest(operator, data, expected)
		})

		Convey("should return no result when having clause is false", func() {

			expr := &sqlparser.ComparisonExpr{
				Operator: sqlparser.AST_GT,
				Left: &sqlparser.FuncExpr{
					Name: []byte("sum"),
					Exprs: sqlparser.SelectExprs{
						&sqlparser.NonStarExpr{
							Expr: &sqlparser.ColName{[]byte("b"), []byte(tableOneName)},
						},
					},
				},
				Right: sqlparser.NumVal(strconv.FormatFloat(999.0, 'E', -1, 64)),
			}

			matcher, err := NewSQLExpr(expr)
			So(err, ShouldBeNil)

			operator := &Having{
				sExprs:  sExprs,
				source:  source,
				matcher: matcher,
			}

			expected := Values{}

			havingTest(operator, data, expected)
		})

	})
}
