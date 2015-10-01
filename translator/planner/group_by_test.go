package planner

import (
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/translator/evaluator"
	"github.com/erh/mongo-sql-temp/translator/types"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func groupByTest(operator Operator, rows []bson.D, expectedRows [][]types.Values) {

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
					Column: Column{tableOneName, "a", "a"},
					Expr:   &sqlparser.ColName{[]byte("a"), []byte(tableOneName)},
				},
				SelectExpression{
					Column: Column{tableOneName, "b", "b"},
					Expr:   &sqlparser.ColName{[]byte("b"), []byte(tableOneName)},
				},
			},
		}

		Convey("should return the right result when using an aggregation function", func() {

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

			exprs := []sqlparser.Expr{
				&sqlparser.ColName{[]byte("a"), []byte(tableOneName)},
			}

			matcher := &evaluator.NoopMatch{}

			operator := &GroupBy{
				sExprs:  sExprs,
				source:  source,
				exprs:   exprs,
				matcher: matcher,
			}

			expected := [][]types.Values{
				[]types.Values{
					{{"a", "a", evaluator.SQLInt(6)}},
					{{"sum(b)", "sum(b)", evaluator.SQLInt(15)}},
				},
			}

			groupByTest(operator, data, expected)

		})

	})
}
