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

func havingTest(operator Operator, rows []bson.D, expectedRows [][]Values) {

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

func TestHavingOperator(t *testing.T) {

	Convey("A having operator...", t, func() {

		data := []bson.D{
			bson.D{
				{"_id", SQLInt(1)},
				{"a", SQLInt(6)},
				{"b", SQLInt(7)},
			},
			bson.D{
				{"_id", SQLInt(2)},
				{"a", SQLInt(6)},
				{"b", SQLInt(8)},
			},
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

		Convey("should return the right result when having clause is true", func() {

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
				Right: sqlparser.NumVal(strconv.FormatFloat(3.0, 'E', -1, 64)),
			}

			matcher, err := NewSQLExpr(expr)
			So(err, ShouldBeNil)

			operator := &Having{
				sExprs:  sExprs,
				source:  source,
				matcher: matcher,
			}

			expected := [][]Values{
				[]Values{
					{{"a", "a", SQLInt(6)}},
					{{"sum(b)", "sum(b)", SQLInt(15)}},
				},
			}

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

			expected := [][]Values{}

			havingTest(operator, data, expected)
		})

	})
}
