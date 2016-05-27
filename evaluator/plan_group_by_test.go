package evaluator

import (
	"testing"

	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestGroupByPlanStage(t *testing.T) {
	runTest := func(groupBy *GroupByStage, rows []bson.D, expectedRows []Values) {
		ctx := &ExecutionCtx{}

		bss := &BSONSourceStage{tableOneName, rows}

		groupBy.source = bss

		iter, err := groupBy.Open(ctx)
		So(err, ShouldBeNil)

		row := &Row{}

		i := 0

		for iter.Next(row) {
			So(len(row.Data), ShouldEqual, len(expectedRows[i]))
			So(row.Data, ShouldResemble, expectedRows[i])
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
			bson.D{{"_id", 3}, {"a", 7}, {"b", 9}},
		}

		Convey("should return the right result when using an aggregation function", func() {

			columnType := schema.ColumnType{schema.SQLInt, schema.MongoInt}
			projectedColumns := ProjectedColumns{
				ProjectedColumn{
					Column: &Column{tableOneName, "a", columnType.SQLType, columnType.MongoType},
					Expr:   SQLColumnExpr{tableOneName, "a", columnType},
				},
				ProjectedColumn{
					Column: &Column{"", "sum(b)", schema.SQLFloat, schema.MongoNone},
					Expr: &SQLAggFunctionExpr{
						Name: "sum",
						Exprs: []SQLExpr{
							SQLColumnExpr{tableOneName, "b", columnType},
						},
					},
				},
			}

			keys := []SQLExpr{SQLColumnExpr{tableOneName, "a", columnType}}

			operator := &GroupByStage{
				projectedColumns: projectedColumns,
				keys:             keys,
			}

			expected := []Values{
				{{tableOneName, "a", SQLInt(6)}, {"", "sum(b)", SQLFloat(15)}},
				{{tableOneName, "a", SQLInt(7)}, {"", "sum(b)", SQLFloat(9)}},
			}

			runTest(operator, data, expected)
		})

	})
}
