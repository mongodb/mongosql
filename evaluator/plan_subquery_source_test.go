package evaluator_test

import (
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/mongodb"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/10gen/mongo-go-driver/bson"
)

func TestSubquerySourceStage(t *testing.T) {
	ctx := &evaluator.ExecutionCtx{}

	testSchema := evaluator.MustLoadSchema(testSchema4)
	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	runTest := func(selectID int, aliasName string, optimize bool, rows []bson.D,
		expectedRows []evaluator.Values) {
		ts := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		var plan evaluator.PlanStage
		var err error

		s := evaluator.NewSubquerySourceStage(ts, selectID, aliasName)
		plan = s
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

	Convey("A subquery source operator should produce the correct results", t, func() {

		rows := []bson.D{
			{{Name: "a", Value: 6}, {Name: "b", Value: 9}},
			{{Name: "a", Value: 3}, {Name: "b", Value: 4}},
		}

		selectID := 42
		aliasName := "funny"

		expected := []evaluator.Values{
			{{SelectID: 42, Database: evaluator.BSONSourceDB, Table: "funny", Name: "a",
				Data: evaluator.SQLInt(6)}, {SelectID: 42, Database: evaluator.BSONSourceDB,
				Table: "funny", Name: "b", Data: evaluator.SQLInt(9)}},
			{{SelectID: 42, Database: evaluator.BSONSourceDB, Table: "funny", Name: "a",
				Data: evaluator.SQLInt(3)}, {SelectID: 42, Database: evaluator.BSONSourceDB,
				Table: "funny", Name: "b", Data: evaluator.SQLInt(4)}},
		}

		runTest(selectID, aliasName, false, rows, expected)

		Convey("and should produce identical results after optimization", func() {
			runTest(selectID, aliasName, true, rows, expected)
		})

	})
}
