package evaluator

import (
	"testing"

	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestProjectOperator(t *testing.T) {
	ctx := &ExecutionCtx{}

	runTest := func(project *ProjectStage, optimize bool, rows []bson.D, expectedRows []Values) {
		ts := NewBSONSourceStage(1, tableOneName, rows)

		project = project.clone()
		var plan PlanStage
		var err error

		project = project.clone()
		project.source = ts
		plan = project
		if optimize {
			plan, err = OptimizePlan(plan)
			So(err, ShouldBeNil)
		}

		iter, err := plan.Open(ctx)
		So(err, ShouldBeNil)

		i := 0
		row := &Row{}

		for iter.Next(row) {
			So(len(row.Data), ShouldEqual, len(expectedRows[i]))
			So(row.Data, ShouldResemble, expectedRows[i])
			row = &Row{}
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

		projectedColumns := ProjectedColumns{
			ProjectedColumn{
				Column: &Column{1, "", "a", schema.SQLInt, schema.MongoInt},
				Expr:   NewSQLColumnExpr(1, tableOneName, "a", schema.SQLInt, schema.MongoInt),
			},
			ProjectedColumn{
				Column: &Column{1, "", "b", schema.SQLInt, schema.MongoInt},
				Expr:   NewSQLColumnExpr(1, tableOneName, "b", schema.SQLInt, schema.MongoInt),
			},
		}

		project := &ProjectStage{
			projectedColumns: projectedColumns,
		}

		expected := []Values{
			{{1, "", "a", SQLInt(6)}, {1, "", "b", SQLInt(9)}},
			{{1, "", "a", SQLInt(3)}, {1, "", "b", SQLInt(4)}},
		}

		runTest(project, false, rows, expected)

		Convey("and should produce identical results after optimization", func() {
			runTest(project, true, rows, expected)
		})

	})
}
