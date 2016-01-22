package evaluator

import (
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func TestOptimizeOperator(t *testing.T) {
	Convey("Subject: OptimizeOperator", t, func() {

		ctx := &ExecutionCtx{
			Schema: cfgOne,
			Db:     dbOne,
		}

		Convey("Given a recursively optimizable tree", func() {

			ts := &TableScan{
				tableName: "foo",
				dbName:    "test",
				pipeline:  []bson.D{},
			}

			sa := &SourceAppend{
				source: ts,
			}

			Convey("Should optimize from bottom-up", func() {

				filter := &Filter{
					source:  sa,
					matcher: &SQLEqualsExpr{SQLFieldExpr{"foo", "a"}, SQLString("funny")},
				}

				limit := &Limit{
					source:   filter,
					rowcount: 42,
				}

				verifyOptimizedPipeline(ctx, limit,
					[]bson.D{
						bson.D{{"$match", bson.M{"a": "funny"}}},
						bson.D{{"$limit", limit.rowcount}}})
			})

			Convey("Should optimize adjacent operators of the same type", func() {

				limit := &Limit{
					source:   sa,
					rowcount: 22,
				}

				skip := &Limit{
					source: limit,
					offset: 20,
				}

				verifyOptimizedPipeline(ctx, skip,
					[]bson.D{
						bson.D{{"$limit", limit.rowcount}},
						bson.D{{"$skip", skip.offset}}})
			})

			Convey("Should optimize multiple operators of the same type split a part", func() {

				skip := &Limit{
					source: sa,
					offset: 20,
				}

				filter := &Filter{
					source:  skip,
					matcher: &SQLEqualsExpr{SQLFieldExpr{"foo", "a"}, SQLString("funny")},
				}

				limit := &Limit{
					source:   filter,
					rowcount: 42,
				}

				verifyOptimizedPipeline(ctx, limit,
					[]bson.D{
						bson.D{{"$skip", skip.offset}},
						bson.D{{"$match", bson.M{"a": "funny"}}},
						bson.D{{"$limit", limit.rowcount}}})
			})
		})

	})
}

func TestFilterPushDown(t *testing.T) {

	Convey("Subject: Filter Optimization", t, func() {

		ctx := &ExecutionCtx{
			Schema: cfgOne,
			Db:     dbOne,
		}

		Convey("Given a push-downable filter", func() {

			ts := &TableScan{
				tableName: "foo",
				dbName:    "test",
				pipeline:  []bson.D{},
			}

			sa := &SourceAppend{
				source: ts,
			}

			filter := &Filter{
				source: sa,
			}

			Convey("Should optimize when the matcher is fully translatable", func() {

				filter.matcher = &SQLEqualsExpr{SQLFieldExpr{"foo", "a"}, SQLString("funny")}

				verifyOptimizedPipeline(ctx, filter,
					[]bson.D{bson.D{{"$match", bson.M{"a": "funny"}}}})
			})

			Convey("Should optimize when the matcher is partially translatable", func() {

				filter.matcher = &SQLAndExpr{
					&SQLEqualsExpr{SQLFieldExpr{"foo", "a"}, SQLString("funny")},
					&SQLEqualsExpr{SQLFieldExpr{"foo", "b"}, SQLFieldExpr{"foo", "c"}}}

				optimized, err := OptimizeOperator(ctx, filter)
				So(err, ShouldBeNil)
				newFilter, ok := optimized.(*Filter)
				So(ok, ShouldBeTrue)
				So(newFilter.matcher, ShouldResemble, &SQLEqualsExpr{SQLFieldExpr{"foo", "b"}, SQLFieldExpr{"foo", "c"}})
				sa, ok := newFilter.source.(*SourceAppend)
				So(ok, ShouldBeTrue)
				ts, ok := sa.source.(*TableScan)
				So(ok, ShouldBeTrue)

				So(ts.pipeline, ShouldResemble, []bson.D{bson.D{{"$match", bson.M{"a": "funny"}}}})
			})
		})

		Convey("Given an immediately evaluated filter", func() {

			ts := &TableScan{
				tableName: "foo",
				dbName:    "test",
				pipeline:  []bson.D{},
			}

			sa := &SourceAppend{
				source: ts,
			}

			filter := &Filter{
				source: sa,
			}

			Convey("Should return an Empty operator when evaluated to false", func() {
				filter.matcher = SQLBool(false)

				optimized, err := OptimizeOperator(ctx, filter)
				So(err, ShouldBeNil)

				_, ok := optimized.(*Empty)
				So(ok, ShouldBeTrue)
			})

			Convey("Should elimate the Filter from the tree and not alter the pipeline when evaluated to true", func() {
				filter.matcher = SQLBool(true)

				verifyOptimizedPipeline(ctx, filter, []bson.D{})
			})
		})

		Convey("Given a non-push-downable filter", func() {

			Convey("Should not optimize the pipeline when the source is not valid", func() {

				empty := &Empty{}

				filter := &Filter{
					source: empty,
				}

				verifyUnoptimizedPipeline(ctx, filter)
			})

			Convey("Should not optimize the pipeline when the filter is not push-downable", func() {

				ts := &TableScan{
					tableName: "foo",
					dbName:    "test",
					pipeline:  []bson.D{},
				}

				sa := &SourceAppend{
					source: ts,
				}

				filter := &Filter{
					source:  sa,
					matcher: &SQLEqualsExpr{SQLFieldExpr{"foo", "a"}, SQLFieldExpr{"foo", "b"}},
				}

				verifyUnoptimizedPipeline(ctx, filter)
			})
		})
	})
}

func TestLimitPushDown(t *testing.T) {

	Convey("Subject: Limit Optimization", t, func() {

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
