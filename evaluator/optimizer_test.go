package evaluator

import (
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func TestLimitPushDown(t *testing.T) {

	Convey("TestLimitPushDown", t, func() {

		ctx := &ExecutionCtx{
			Schema: cfgOne,
			Db:     dbOne,
		}

		Convey("Given a push-downable limit", func() {

			ts := &TableScan{
				tableName: "foo",
				dbName:    "test",
				pipeline:  []bson.D{},
			}

			sa := &SourceAppend{
				source: ts,
			}

			limit := &Limit{
				source: sa,
			}

			Convey("Should optimize with only an offset", func() {

				limit.offset = 22

				verifyOptimizedPipeline(ctx, limit,
					[]bson.D{bson.D{{"$skip", limit.offset}}})

			})

			Convey("Should optimized with only a rowcount", func() {

				limit.rowcount = 20

				verifyOptimizedPipeline(ctx, limit,
					[]bson.D{bson.D{{"$limit", limit.rowcount}}})
			})

			Convey("Should optimized with both an offset and a rowcount", func() {

				limit.offset = 22
				limit.rowcount = 20

				verifyOptimizedPipeline(ctx, limit, []bson.D{
					bson.D{{"$skip", limit.offset}},
					bson.D{{"$limit", limit.rowcount}}})
			})
		})

		Convey("Given a non-push-downable limit", func() {

			empty := &Empty{}

			limit := &Limit{
				source: empty,
			}

			Convey("Should not optimized the pipeline", func() {
				verifyUnoptimizedPipeline(ctx, limit)
			})
		})
	})
}

func verifyUnoptimizedPipeline(ctx *ExecutionCtx, op Operator) {
	optimized, err := OptimizeOperator(ctx, op)
	So(err, ShouldBeNil)
	So(optimized, ShouldEqual, op)
}

func verifyOptimizedPipeline(ctx *ExecutionCtx, op Operator, pipeline []bson.D) {
	optimized, err := OptimizeOperator(ctx, op)
	So(err, ShouldBeNil)

	sa, ok := optimized.(*SourceAppend)
	So(ok, ShouldBeTrue)

	ts, ok := sa.source.(*TableScan)
	So(ok, ShouldBeTrue)

	So(ts.pipeline, ShouldResemble, pipeline)
}
