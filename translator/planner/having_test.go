package planner

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/translator/evaluator"
	"github.com/erh/mongo-sql-temp/translator/types"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"testing"
)

var (
	_ fmt.Stringer = nil
)

func havingTest(operator Operator, rows []bson.D, expectedRows [][]types.Values) {

	cfg, err := config.ParseConfigData(testConfigSimple)
	So(err, ShouldBeNil)

	session, err := mgo.Dial(cfg.Url)
	So(err, ShouldBeNil)

	collection := session.DB(dbName).C(tableOneName)
	collection.DropCollection()

	for _, row := range rows {
		So(collection.Insert(row), ShouldBeNil)
	}

	ctx := &ExecutionCtx{
		Config: cfg,
		Db:     dbName,
	}

	So(operator.Open(ctx), ShouldBeNil)

	row := &types.Row{}

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
		row = &types.Row{}
		i++
	}
}

func TestHavingOperator(t *testing.T) {

	Convey("A having operator...", t, func() {

		data := []bson.D{
			bson.D{
				{"_id", evaluator.SQLNumeric(1)},
				{"a", evaluator.SQLNumeric(6)},
				{"b", evaluator.SQLNumeric(7)},
			},
			bson.D{
				{"_id", evaluator.SQLNumeric(2)},
				{"a", evaluator.SQLNumeric(6)},
				{"b", evaluator.SQLNumeric(8)},
			},
		}

		source := &Select{
			source: &TableScan{
				tableName: tableOneName,
			},
			sExprs: SelectExpressions{
				SelectExpression{
					Column: Column{tableOneName, "a", "a"},
					Expr:   &sqlparser.ColName{[]byte("a"), []byte(tableOneName)},
				},
				SelectExpression{
					Column: Column{tableOneName, "b", "b"},
					Expr:   &sqlparser.ColName{[]byte("b"), []byte(tableOneName)},
				},
			},
		}

		sExprs := SelectExpressions{
			SelectExpression{
				Column: Column{tableOneName, "a", "a"},
				Expr:   &sqlparser.ColName{[]byte("a"), []byte(tableOneName)},
			},
			SelectExpression{
				Column: Column{"", "sum(b)", "sum(b)"},
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

			matcher, err := evaluator.BuildMatcher(expr)
			So(err, ShouldBeNil)

			operator := &Having{
				sExprs:  sExprs,
				source:  source,
				matcher: matcher,
			}

			expected := [][]types.Values{
				[]types.Values{
					{{"a", "a", evaluator.SQLNumeric(6)}},
					{{"sum(b)", "sum(b)", evaluator.SQLNumeric(15)}},
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

			matcher, err := evaluator.BuildMatcher(expr)
			So(err, ShouldBeNil)

			operator := &Having{
				sExprs:  sExprs,
				source:  source,
				matcher: matcher,
			}

			expected := [][]types.Values{}

			havingTest(operator, data, expected)
		})

	})
}
