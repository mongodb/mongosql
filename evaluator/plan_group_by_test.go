package evaluator

import (
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGroupByPlanStage(t *testing.T) {
	ctx := createTestExecutionCtx(nil)

	runTest := func(groupBy *GroupByStage, rows []bson.D, expectedRows []Values) {

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

	Convey("Subject: GroupByStage", t, func() {

		data := []bson.D{
			bson.D{{"_id", 1}, {"a", "a"}, {"b", 7}},
			bson.D{{"_id", 2}, {"a", "A"}, {"b", 8}},
			bson.D{{"_id", 3}, {"a", "b"}, {"b", 9}},
		}

		Convey("should return the right result when using an aggregation function", func() {

			projectedColumns := ProjectedColumns{
				ProjectedColumn{
					Column: &Column{1, tableOneName, "a", schema.SQLVarchar, schema.MongoInt, false},
					Expr:   NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString),
				},
				ProjectedColumn{
					Column: &Column{1, "", "sum(b)", schema.SQLFloat, schema.MongoNone, false},
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

func TestGroupByPlanStage_MemoryLimits(t *testing.T) {
	ctx := createTestExecutionCtx(nil)
	ctx.Variables().MongoDBMaxStageSize = 100

	runTest := func(groupBy *GroupByStage, rows []bson.D) {
		bss := NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		groupBy.source = bss

		iter, err := groupBy.Open(ctx)
		So(err, ShouldBeNil)

		row := &Row{}

		ok := iter.Next(row)
		So(ok, ShouldBeFalse)

		So(iter.Close(), ShouldBeNil)
		So(iter.Err(), ShouldNotBeNil)
	}

	Convey("Subject: GroupByStage Memory Limits", t, func() {

		data := []bson.D{
			bson.D{{"_id", 1}, {"a", "a"}, {"b", 7}},
			bson.D{{"_id", 2}, {"a", "A"}, {"b", 8}},
			bson.D{{"_id", 3}, {"a", "b"}, {"b", 9}},
		}

		projectedColumns := ProjectedColumns{
			ProjectedColumn{
				Column: &Column{1, tableOneName, "a", schema.SQLVarchar, schema.MongoInt, false},
				Expr:   NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString),
			},
			ProjectedColumn{
				Column: &Column{1, "", "sum(b)", schema.SQLFloat, schema.MongoNone, false},
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

		runTest(operator, data)

	})
}
