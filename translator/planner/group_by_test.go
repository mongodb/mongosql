package planner

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/translator/types"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var (
	_ fmt.Stringer = nil
)

func groupByTest(operator Operator, rows []interface{}, expectedRows [][]bson.D) {

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

	Convey("With a simple test configuration...", t, func() {

		data := []interface{}{
			bson.D{
				bson.DocElem{Name: "_id", Value: 1},
				bson.DocElem{Name: "a", Value: 6},
				bson.DocElem{Name: "b", Value: 7},
			},
			bson.D{
				bson.DocElem{Name: "_id", Value: 2},
				bson.DocElem{Name: "a", Value: 6},
				bson.DocElem{Name: "b", Value: 8},
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

		Convey("a group by operator using an aggregation function should return the right aggregate result", func() {

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

			operator := &GroupBy{
				sExprs: sExprs,
				source: source,
				exprs:  exprs,
			}

			expected := [][]bson.D{[]bson.D{bson.D{{"a", 6}}, bson.D{{"sum(b)", int64(15)}}}}

			groupByTest(operator, data, expected)

		})

	})
}
