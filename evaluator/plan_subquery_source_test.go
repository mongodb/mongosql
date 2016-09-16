package evaluator

import (
	"testing"

	"github.com/10gen/sqlproxy/collation"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestSubquerySourceStage(t *testing.T) {
	ctx := &ExecutionCtx{}

	runTest := func(s *SubquerySourceStage, optimize bool, rows []bson.D, expectedRows []Values) {
		ts := NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		var plan PlanStage
		var err error

		s = s.clone()
		s.source = ts
		plan = s
		if optimize {
			plan, err = OptimizePlan(createTestConnectionCtx(), plan)
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

	Convey("A subquery source operator should produce the correct results", t, func() {

		rows := []bson.D{
			bson.D{{"a", 6}, {"b", 9}},
			bson.D{{"a", 3}, {"b", 4}},
		}

		project := &SubquerySourceStage{
			selectID:  42,
			aliasName: "funny",
		}

		expected := []Values{
			{{42, "funny", "a", SQLInt(6)}, {42, "funny", "b", SQLInt(9)}},
			{{42, "funny", "a", SQLInt(3)}, {42, "funny", "b", SQLInt(4)}},
		}

		runTest(project, false, rows, expected)

		Convey("and should produce identical results after optimization", func() {
			runTest(project, true, rows, expected)
		})

	})
}
