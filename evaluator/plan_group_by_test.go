package evaluator_test

import (
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGroupByPlanStage(t *testing.T) {
	ctx := createTestExecutionCtx(nil)

	runTest := func(projectedColumns evaluator.ProjectedColumns, keys []evaluator.SQLExpr,
		rows []bson.D, expectedRows []evaluator.Values) {

		bss := evaluator.NewBSONSourceStage(1, tableOneName,
			collation.Must(collation.Get("utf8_general_ci")), rows)

		groupBy := evaluator.NewGroupByStage(bss, keys, projectedColumns)

		iter, err := groupBy.Open(ctx)
		So(err, ShouldBeNil)

		row := &evaluator.Row{}

		i := 0

		for iter.Next(row) {
			So(len(row.Data), ShouldEqual, len(expectedRows[i]))
			So(row.Data, ShouldResemble, expectedRows[i])
			row = &evaluator.Row{}
			i++
		}

		So(iter.Close(), ShouldBeNil)
		So(iter.Err(), ShouldBeNil)
	}

	Convey("Subject: GroupByStage", t, func() {

		data := []bson.D{
			{{Name: "_id", Value: 1}, {Name: "a", Value: "a"}, {Name: "b", Value: 7}},
			{{Name: "_id", Value: 2}, {Name: "a", Value: "A"}, {Name: "b", Value: 8}},
			{{Name: "_id", Value: 3}, {Name: "a", Value: "b"}, {Name: "b", Value: 9}},
		}

		Convey("should return the right result when using an aggregation function", func() {

			projectedColumns := evaluator.ProjectedColumns{
				evaluator.ProjectedColumn{
					Column: &evaluator.Column{SelectID: 1, Table: tableOneName,
						OriginalTable: tableOneName, Database: evaluator.BSONSourceDB, Name: "a",
						OriginalName: "a", MappingRegistryName: "", SQLType: schema.SQLVarchar,
						MongoType: schema.MongoInt, PrimaryKey: false},
					Expr: evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "a",
						schema.SQLVarchar, schema.MongoString),
				},
				evaluator.ProjectedColumn{
					Column: &evaluator.Column{SelectID: 1, Table: "", OriginalTable: "",
						Database: evaluator.BSONSourceDB, Name: "sum(b)", OriginalName: "sum(b)",
						MappingRegistryName: "", SQLType: schema.SQLFloat,
						MongoType: schema.MongoNone, PrimaryKey: false},
					Expr: &evaluator.SQLAggFunctionExpr{
						Name: "sum",
						Exprs: []evaluator.SQLExpr{
							evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "b",
								schema.SQLInt, schema.MongoInt),
						},
					},
				},
			}

			keys := []evaluator.SQLExpr{evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB,
				tableOneName, "a", schema.SQLVarchar, schema.MongoString)}

			expected := []evaluator.Values{
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLVarchar("a")}, {SelectID: 1,
					Database: evaluator.BSONSourceDB, Table: "", Name: "sum(b)",
					Data: evaluator.SQLFloat(15)}},
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLVarchar("b")}, {SelectID: 1,
					Database: evaluator.BSONSourceDB, Table: "", Name: "sum(b)",
					Data: evaluator.SQLFloat(9)}},
			}

			runTest(projectedColumns, keys, data, expected)
		})

	})
}

func TestGroupByPlanStage_MemoryLimits(t *testing.T) {
	ctx := createTestExecutionCtx(nil)
	ctx.Variables().SetSystemVariable(variable.MongoDBMaxStageSize, 100)

	runTest := func(projectedColumns evaluator.ProjectedColumns, keys []evaluator.SQLExpr,
		rows []bson.D) {
		bss := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		groupBy := evaluator.NewGroupByStage(bss, keys, projectedColumns)

		iter, err := groupBy.Open(ctx)
		So(err, ShouldBeNil)

		row := &evaluator.Row{}

		ok := iter.Next(row)
		So(ok, ShouldBeFalse)

		So(iter.Close(), ShouldBeNil)
		So(iter.Err(), ShouldNotBeNil)
	}

	Convey("Subject: GroupByStage Memory Limits", t, func() {

		data := []bson.D{
			{{Name: "_id", Value: 1}, {Name: "a", Value: "a"}, {Name: "b", Value: 7}},
			{{Name: "_id", Value: 2}, {Name: "a", Value: "A"}, {Name: "b", Value: 8}},
			{{Name: "_id", Value: 3}, {Name: "a", Value: "b"}, {Name: "b", Value: 9}},
		}

		projectedColumns := evaluator.ProjectedColumns{
			evaluator.ProjectedColumn{
				Column: &evaluator.Column{SelectID: 1, Table: tableOneName,
					OriginalTable: tableOneName, Database: evaluator.BSONSourceDB, Name: "a",
					OriginalName: "a", MappingRegistryName: "", SQLType: schema.SQLVarchar,
					MongoType: schema.MongoInt, PrimaryKey: false},
				Expr: evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "a",
					schema.SQLVarchar, schema.MongoString),
			},
			evaluator.ProjectedColumn{
				Column: &evaluator.Column{SelectID: 1, Table: "", OriginalTable: "",
					Database: evaluator.BSONSourceDB, Name: "sum(b)", OriginalName: "sum(b)",
					MappingRegistryName: "", SQLType: schema.SQLFloat, MongoType: schema.MongoNone,
					PrimaryKey: false},
				Expr: &evaluator.SQLAggFunctionExpr{
					Name: "sum",
					Exprs: []evaluator.SQLExpr{
						evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "b",
							schema.SQLInt, schema.MongoInt),
					},
				},
			},
		}

		keys := []evaluator.SQLExpr{evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB,
			tableOneName, "a", schema.SQLVarchar, schema.MongoString)}

		runTest(projectedColumns, keys, data)

	})
}
