package evaluator

import (
	"testing"

	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewAlgebrize(t *testing.T) {

	schema, _ := schema.New(testSchema1)

	test := func(sql, dbName string, expectedPlanFactory func() PlanStage) {
		Convey(sql, func() {
			statement, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			selectStatement := statement.(*sqlparser.Select)
			algebrizerContext := NewAlgebrizerContext(schema)
			actualPlan, err := Algebrize(selectStatement, algebrizerContext)
			So(err, ShouldBeNil)

			expectedPlan := expectedPlanFactory()
			So(actualPlan, ShouldResemble, expectedPlan)
		})
	}

	Convey("Subject: Algebrize", t, func() {
		test("select a from foo", "test", func() PlanStage {
			source, _ := NewMongoSourceStage(schema, "test", "foo", "foo")
			return &ProjectStage{
				source: source,
				sExprs: nil,
			}
		})
	})
}
