package evaluator

import (
	"testing"

	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestGroupByStage(t *testing.T) {
	columnType := schema.ColumnType{schema.SQLInt, schema.MongoInt}

	runTest := func(groupBy *GroupByStage, rows []bson.D, expectedRows [][]Values) {
		ctx := &ExecutionCtx{}

		bss := &BSONSourceStage{tableOneName, rows}

		groupBy.source = &ProjectStage{
			source: bss,
			sExprs: SelectExpressions{
				SelectExpression{
					Column: &Column{tableOneName, "a", "a", columnType.SQLType, columnType.MongoType},
					Expr:   SQLColumnExpr{tableOneName, "a", columnType},
				},
				SelectExpression{
					Column: &Column{tableOneName, "b", "b", columnType.SQLType, columnType.MongoType},
					Expr:   SQLColumnExpr{tableOneName, "b", columnType},
				},
			},
		}

		iter, err := groupBy.Open(ctx)
		So(err, ShouldBeNil)

		row := &Row{}

		i := 0

		for iter.Next(row) {
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

		So(iter.Close(), ShouldBeNil)
		So(iter.Err(), ShouldBeNil)
	}

	Convey("A group by operator...", t, func() {

		data := []bson.D{
			bson.D{{"_id", 1}, {"a", 6}, {"b", 7}},
			bson.D{{"_id", 2}, {"a", 6}, {"b", 8}},
		}

		Convey("should return the right result when using an aggregation function", func() {

			sExprs := SelectExpressions{
				SelectExpression{
					Column: &Column{tableOneName, "a", "a", columnType.SQLType, columnType.MongoType},
					Expr:   SQLColumnExpr{tableOneName, "a", columnType},
				},
				SelectExpression{
					Column: &Column{"", "sum(b)", "sum(b)", schema.SQLFloat, schema.MongoNone},
					Expr: &SQLAggFunctionExpr{
						Name: "sum",
						Exprs: []SQLExpr{
							SQLColumnExpr{tableOneName, "b", columnType},
						},
					},
				},
			}

			exprs := SelectExpressions{
				SelectExpression{
					Column: &Column{tableOneName, "a", "a", columnType.SQLType, columnType.MongoType},
					Expr:   SQLColumnExpr{tableOneName, "a", columnType},
				},
			}

			operator := &GroupByStage{
				selectExprs: sExprs,
				keyExprs:    exprs,
			}

			expected := [][]Values{
				[]Values{
					{{"a", "a", SQLInt(6)}},
					{{"sum(b)", "sum(b)", SQLFloat(15)}},
				},
			}

			runTest(operator, data, expected)

		})

	})
}
