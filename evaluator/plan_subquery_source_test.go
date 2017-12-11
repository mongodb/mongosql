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

func TestSubquerySourceStage(t *testing.T) {
	ctx := &evaluator.ExecutionCtx{}

	testSchema, err := schema.New(testSchema4)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}

	testInfo := getMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	runTest := func(selectID int, aliasName string, optimize bool, rows []bson.D, expectedRows []evaluator.Values) {
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
			bson.D{{"a", 6}, {"b", 9}},
			bson.D{{"a", 3}, {"b", 4}},
		}

		selectID := 42
		aliasName := "funny"

		expected := []evaluator.Values{
			{{42, evaluator.BSONSourceDB, "funny", "a", evaluator.SQLInt(6)}, {42, evaluator.BSONSourceDB, "funny", "b", evaluator.SQLInt(9)}},
			{{42, evaluator.BSONSourceDB, "funny", "a", evaluator.SQLInt(3)}, {42, evaluator.BSONSourceDB, "funny", "b", evaluator.SQLInt(4)}},
		}

		runTest(selectID, aliasName, false, rows, expected)

		Convey("and should produce identical results after optimization", func() {
			runTest(selectID, aliasName, true, rows, expected)
		})

	})
}
