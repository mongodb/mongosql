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
			bson.D{{"a", 6}, {"b", 9}},
			bson.D{{"a", 3}, {"b", 4}},
		}

		projectedColumns := evaluator.ProjectedColumns{
			evaluator.ProjectedColumn{
				Column: &evaluator.Column{1, "", "", evaluator.BSONSourceDB, "a", "a", "", schema.SQLInt, schema.MongoInt, false},
				Expr:   evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "a", schema.SQLInt, schema.MongoInt),
			},
			evaluator.ProjectedColumn{
				Column: &evaluator.Column{1, "", "", evaluator.BSONSourceDB, "b", "b", "", schema.SQLInt, schema.MongoInt, false},
				Expr:   evaluator.NewSQLColumnExpr(1, evaluator.BSONSourceDB, tableOneName, "b", schema.SQLInt, schema.MongoInt),
			},
		}

		expected := []evaluator.Values{
			{{1, evaluator.BSONSourceDB, "", "a", evaluator.SQLInt(6)}, {1, evaluator.BSONSourceDB, "", "b", evaluator.SQLInt(9)}},
			{{1, evaluator.BSONSourceDB, "", "a", evaluator.SQLInt(3)}, {1, evaluator.BSONSourceDB, "", "b", evaluator.SQLInt(4)}},
		}

		runTest(projectedColumns, false, rows, expected)

		Convey("and should produce identical results after optimization", func() {
			runTest(projectedColumns, true, rows, expected)
		})

	})
}
