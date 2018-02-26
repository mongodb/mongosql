package evaluator_test

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/10gen/mongo-go-driver/bson"
)

func TestProjectOperator(t *testing.T) {
	ctx := &evaluator.ExecutionCtx{}

	testSchema, err := schema.New(testSchema4, &lgr)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}

	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	runTest := func(projectedColumns evaluator.ProjectedColumns, optimize bool, rows []bson.D, expectedRows []evaluator.Values) {
		ts := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		var plan evaluator.PlanStage
		var err error

		project := evaluator.NewProjectStage(ts, projectedColumns...)

		plan = project
		if optimize {
			plan = evaluator.OptimizePlan(createTestConnectionCtx(testInfo), plan)
		}

		iter, err := plan.Open(ctx)
		So(err, ShouldBeNil)

		i := 0
		row := &evaluator.Row{}

		for iter.Next(row) {
			So(len(row.Data), ShouldEqual, len(expectedRows[i]))
			So(row.Data, ShouldResemble, expectedRows[i])
			row = &evaluator.Row{}
			i++
		}

		So(i, ShouldEqual, len(expectedRows))

		So(iter.Close(), ShouldBeNil)
		So(iter.Err(), ShouldBeNil)
	}

	Convey("A project operator should produce the correct results", t, func() {

		rows := []bson.D{
			{{Name: "a", Value: 6}, {Name: "b", Value: 9}},
			{{Name: "a", Value: 3}, {Name: "b", Value: 4}},
		}

		projectedColumns := evaluator.ProjectedColumns{
			evaluator.ProjectedColumn{
				Column: &evaluator.Column{SelectID: 1, Table: "", OriginalTable: "", Database: evaluator.BSONSourceDB, Name: "a", OriginalName: "a", MappingRegistryName: "", SQLType: schema.SQLInt, MongoType: schema.MongoInt, PrimaryKey: false},
				Expr:   evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "a", schema.SQLInt, schema.MongoInt),
			},
			evaluator.ProjectedColumn{
				Column: &evaluator.Column{SelectID: 1, Table: "", OriginalTable: "", Database: evaluator.BSONSourceDB, Name: "b", OriginalName: "b", MappingRegistryName: "", SQLType: schema.SQLInt, MongoType: schema.MongoInt, PrimaryKey: false},
				Expr:   evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "b", schema.SQLInt, schema.MongoInt),
			},
		}

		expected := []evaluator.Values{
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: "", Name: "a", Data: evaluator.SQLInt(6)}, {SelectID: 1, Database: evaluator.BSONSourceDB, Table: "", Name: "b", Data: evaluator.SQLInt(9)}},
			{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: "", Name: "a", Data: evaluator.SQLInt(3)}, {SelectID: 1, Database: evaluator.BSONSourceDB, Table: "", Name: "b", Data: evaluator.SQLInt(4)}},
		}

		runTest(projectedColumns, false, rows, expected)

		Convey("and should produce identical results after optimization", func() {
			runTest(projectedColumns, true, rows, expected)
		})

	})
}
