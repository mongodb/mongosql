package evaluator

import (
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestCacheOperator(t *testing.T) {
	ctx := &ExecutionCtx{
		CacheRows: make(map[string]interface{}),
	}

	runTest := func(cache *CacheStage, optimize bool, rows []bson.D, expectedRows []Values) {
		ts := NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		var plan PlanStage
		var err error

		cache.source = ts
		plan = cache
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

	Convey("A cache operator should return the same thing as the underlying plan stages", t, func() {

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

		cache := &CacheStage{
			source: project,
			key:    1,
		}

		expected := []Values{
			{{1, "foo", "a", SQLInt(6)}, {1, "foo", "b", SQLInt(9)}},
			{{1, "foo", "a", SQLInt(3)}, {1, "foo", "b", SQLInt(4)}},
		}

		runTest(cache, false, rows, expected)

		Convey("and should produce identical results after optimization", func() {
			runTest(cache, true, rows, expected)
		})

	})
}
