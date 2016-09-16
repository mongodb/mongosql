package evaluator

import (
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestGroupByPlanStage(t *testing.T) {

	runTest := func(groupBy *GroupByStage, rows []bson.D, expectedRows []Values) {
		ctx := &ExecutionCtx{}

		bss := NewBSONSourceStage(1, tableOneName, collation.Must(collation.Get("utf8_general_ci")), rows)

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
			bson.D{{"_id", 1}, {"a", "a"}, {"b", 7}},
			bson.D{{"_id", 2}, {"a", "A"}, {"b", 8}},
			bson.D{{"_id", 3}, {"a", "b"}, {"b", 9}},
		}

		Convey("should return the right result when using an aggregation function", func() {

			projectedColumns := ProjectedColumns{
				ProjectedColumn{
					Column: &Column{1, tableOneName, "a", schema.SQLVarchar, schema.MongoInt},
					Expr:   NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString),
				},
				ProjectedColumn{
					Column: &Column{1, "", "sum(b)", schema.SQLFloat, schema.MongoNone},
					Expr: &SQLAggFunctionExpr{
						Name: "sum",
						Exprs: []SQLExpr{
							NewSQLColumnExpr(1, tableOneName, "b", schema.SQLInt, schema.MongoInt),
						},
					},
				},
			}

			keys := []SQLExpr{NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString)}

			operator := &GroupByStage{
				projectedColumns: projectedColumns,
				keys:             keys,
			}

			expected := []Values{
				{{1, tableOneName, "a", SQLVarchar("a")}, {1, "", "sum(b)", SQLFloat(15)}},
				{{1, tableOneName, "a", SQLVarchar("b")}, {1, "", "sum(b)", SQLFloat(9)}},
			}

			runTest(operator, data, expected)
		})

	})
}
