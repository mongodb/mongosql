package evaluator

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestOptimizeOperator(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne
	columnType := schema.ColumnType{schema.SQLInt, schema.MongoInt}

	Convey("Subject: OptimizeOperator", t, func() {

		ctx := &ExecutionCtx{
			PlanCtx: &PlanCtx{
				Schema: cfgOne,
				Db:     dbOne,
			},
		}

		tbl := "foo"

		Convey("Given a recursively optimizable tree", func() {

			ms, err := NewMongoSourceStage(ctx.PlanCtx, dbOne, tbl, "")
			So(err, ShouldBeNil)

			Convey("Should optimize from bottom-up", func() {

				filter := &FilterStage{
					source:  ms,
					matcher: &SQLEqualsExpr{SQLColumnExpr{tbl, "a", columnType}, SQLInt(4)},
				}

				limit := &LimitStage{
					source: filter,
					limit:  42,
				}

				verifyOptimizedPipeline(ctx.PlanCtx, limit,
					[]bson.D{
						bson.D{{"$match", bson.M{"a": int64(4)}}},
						bson.D{{"$limit", limit.limit}}})
			})

			Convey("Should optimize adjacent operators of the same type", func() {

				limit := &LimitStage{
					source: ms,
					limit:  22,
				}

				skip := &LimitStage{
					source: limit,
					offset: 20,
				}

				verifyOptimizedPipeline(ctx.PlanCtx, skip,
					[]bson.D{
						bson.D{{"$limit", limit.limit}},
						bson.D{{"$skip", skip.offset}}})
			})

			Convey("Should optimize multiple operators of the same type split a part", func() {

				skip := &LimitStage{
					source: ms,
					offset: 20,
				}

				filter := &FilterStage{
					source:  skip,
					matcher: &SQLEqualsExpr{SQLColumnExpr{tbl, "a", columnType}, SQLInt(5)},
				}

				limit := &LimitStage{
					source: filter,
					limit:  42,
				}

				verifyOptimizedPipeline(ctx.PlanCtx, limit,
					[]bson.D{
						bson.D{{"$skip", skip.offset}},
						bson.D{{"$match", bson.M{"a": int64(5)}}},
						bson.D{{"$limit", limit.limit}}})
			})
		})

	})
}

func TestFilterPushDown(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne
	columnType := schema.ColumnType{schema.SQLInt, schema.MongoInt}

	Convey("Subject: Filter Optimization", t, func() {

		ctx := &ExecutionCtx{
			PlanCtx: &PlanCtx{
				Schema: cfgOne,
				Db:     dbOne,
			},
		}

		tbl := "foo"

		Convey("Given a push-downable filter", func() {

			ms, err := NewMongoSourceStage(ctx.PlanCtx, dbOne, tbl, "")
			So(err, ShouldBeNil)

			filter := &FilterStage{source: ms}

			Convey("Should optimize when the matcher is fully translatable", func() {

				filter.matcher = &SQLEqualsExpr{SQLColumnExpr{tbl, "a", columnType}, SQLInt(3)}

				verifyOptimizedPipeline(ctx.PlanCtx, filter,
					[]bson.D{bson.D{{"$match", bson.M{"a": int64(3)}}}})
			})

			Convey("Should optimize when the matcher is partially translatable", func() {

				filter.matcher = &SQLAndExpr{
					&SQLEqualsExpr{SQLColumnExpr{tbl, "a", columnType}, SQLInt(2)},
					&SQLEqualsExpr{SQLColumnExpr{tbl, "b", columnType}, SQLColumnExpr{tbl, "c", columnType}}}

				optimized, err := OptimizePlan(ctx.PlanCtx, filter)
				So(err, ShouldBeNil)
				newFilter, ok := optimized.(*FilterStage)
				So(ok, ShouldBeTrue)
				So(newFilter.matcher, ShouldResemble, &SQLEqualsExpr{SQLColumnExpr{tbl, "b", columnType}, SQLColumnExpr{tbl, "c", columnType}})
				ms, ok := newFilter.source.(*MongoSourceStage)
				So(ok, ShouldBeTrue)
				So(ms.pipeline, ShouldResemble, []bson.D{bson.D{{"$match", bson.M{"a": int64(2)}}}})
			})
		})

		Convey("Given an immediately evaluated filter", func() {

			ms, err := NewMongoSourceStage(ctx.PlanCtx, dbOne, tbl, "")
			So(err, ShouldBeNil)

			filter := &FilterStage{
				source: ms,
			}

			Convey("Should return an Empty operator when evaluated to false", func() {
				filter.matcher = SQLBool(false)

				optimized, err := OptimizePlan(ctx.PlanCtx, filter)
				So(err, ShouldBeNil)

				_, ok := optimized.(*EmptyStage)
				So(ok, ShouldBeTrue)
			})

			Convey("Should elimate the Filter from the tree and not alter the pipeline when evaluated to true", func() {
				filter.matcher = SQLBool(true)

				verifyOptimizedPipeline(ctx.PlanCtx, filter, ms.pipeline)
			})
		})

		Convey("Given a non-push-downable filter", func() {

			Convey("Should not optimize the pipeline when the source is not valid", func() {

				empty := &EmptyStage{nil}

				filter := &FilterStage{
					source: empty,
				}

				verifyUnoptimizedPipeline(ctx.PlanCtx, filter)
			})

			Convey("Should not optimize the pipeline when the filter is not push-downable", func() {

				ms, err := NewMongoSourceStage(ctx.PlanCtx, dbOne, tbl, "")
				So(err, ShouldBeNil)

				filter := &FilterStage{
					source:  ms,
					matcher: &SQLEqualsExpr{SQLColumnExpr{tbl, "a", columnType}, SQLColumnExpr{tbl, "b", columnType}},
				}

				verifyUnoptimizedPipeline(ctx.PlanCtx, filter)
			})
		})
	})
}

func TestCrossJoinOptimization(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne

	columnType := schema.ColumnType{schema.SQLInt, schema.MongoInt}

	Convey("Subject: Filter/Join Merging", t, func() {

		ctx := &ExecutionCtx{
			PlanCtx: &PlanCtx{
				Schema: cfgOne,
				Db:     dbOne,
			},
		}

		tblOne := "foo"
		tblTwo := "bar"

		msOne, err := NewMongoSourceStage(ctx.PlanCtx, dbOne, tblOne, tblOne)
		So(err, ShouldBeNil)
		msTwo, err := NewMongoSourceStage(ctx.PlanCtx, dbOne, tblTwo, tblTwo)
		So(err, ShouldBeNil)

		join := &JoinStage{
			left:  msOne,
			right: msTwo,
		}

		for _, k := range []JoinKind{CrossJoin, InnerJoin} {
			Convey(fmt.Sprintf("Given a %q without criteria", k), func() {

				join.kind = k

				Convey("Should optimize a filter with criteria for both tables", func() {
					filter := &FilterStage{
						source:  join,
						matcher: &SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLColumnExpr{tblTwo, "a", columnType}},
					}

					o, err := optimizeCrossJoins(filter)
					So(err, ShouldBeNil)

					So(o, ShouldResemble, &JoinStage{
						left:    msOne,
						right:   msTwo,
						kind:    InnerJoin,
						matcher: &SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLColumnExpr{tblTwo, "a", columnType}},
					})
				})

				Convey("Should optimize a filter with criteria for only the right table", func() {
					filter := &FilterStage{
						source:  join,
						matcher: &SQLEqualsExpr{SQLColumnExpr{tblTwo, "a", columnType}, SQLInt(10)},
					}

					o, err := optimizeCrossJoins(filter)
					So(err, ShouldBeNil)

					So(o, ShouldResemble, &JoinStage{
						left:    msOne,
						right:   msTwo,
						kind:    InnerJoin,
						matcher: &SQLEqualsExpr{SQLColumnExpr{tblTwo, "a", columnType}, SQLInt(10)},
					})
				})

				Convey("Should optimize a filter with partially applicable criteria", func() {
					filter := &FilterStage{
						source: join,
						matcher: &SQLAndExpr{
							&SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLInt(10)},
							&SQLEqualsExpr{SQLColumnExpr{tblTwo, "a", columnType}, SQLInt(11)},
						},
					}

					o, err := optimizeCrossJoins(filter)
					So(err, ShouldBeNil)

					So(o, ShouldResemble, &FilterStage{
						source: &JoinStage{
							left:    msOne,
							right:   msTwo,
							kind:    InnerJoin,
							matcher: &SQLEqualsExpr{SQLColumnExpr{tblTwo, "a", columnType}, SQLInt(11)},
						},
						matcher: &SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLInt(10)},
					})
				})

				Convey("Should not optimize a filter with criteria for only the left table", func() {
					filter := &FilterStage{
						source:  join,
						matcher: &SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLInt(10)},
					}

					o, err := optimizeCrossJoins(filter)
					So(err, ShouldBeNil)

					So(o, ShouldResemble, filter)
				})

				Convey("Given a nested %q without criteria", func() {

					tblThree := "baz"
					msThree, err := NewMongoSourceStage(ctx.PlanCtx, dbOne, tblThree, tblThree)
					So(err, ShouldBeNil)

					nestedJoin := &JoinStage{
						left:  join,
						right: msThree,
						kind:  k,
					}

					Convey("Should optimize a filter with criteria for all three tables", func() {
						filter := &FilterStage{
							source: nestedJoin,
							matcher: &SQLAndExpr{
								&SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLColumnExpr{tblTwo, "a", columnType}},
								&SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLColumnExpr{tblThree, "a", columnType}},
							},
						}

						o, err := optimizeCrossJoins(filter)
						So(err, ShouldBeNil)

						So(o, ShouldResemble, &JoinStage{
							left: &JoinStage{
								left:    msOne,
								right:   msTwo,
								kind:    InnerJoin,
								matcher: &SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLColumnExpr{tblTwo, "a", columnType}},
							},
							right:   msThree,
							kind:    InnerJoin,
							matcher: &SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLColumnExpr{tblThree, "a", columnType}},
						})
					})

					Convey("Should optimize a filter with criteria for only the nested join", func() {
						filter := &FilterStage{
							source:  nestedJoin,
							matcher: &SQLEqualsExpr{SQLColumnExpr{tblTwo, "a", columnType}, SQLColumnExpr{tblThree, "a", columnType}},
						}

						o, err := optimizeCrossJoins(filter)
						So(err, ShouldBeNil)

						So(o, ShouldResemble, &JoinStage{
							left: &JoinStage{
								left:  msOne,
								right: msTwo,
								kind:  k,
							},
							right:   msThree,
							kind:    InnerJoin,
							matcher: &SQLEqualsExpr{SQLColumnExpr{tblTwo, "a", columnType}, SQLColumnExpr{tblThree, "a", columnType}},
						})
					})

					Convey("Should optimize a filter with criteria for only the top-level join", func() {
						filter := &FilterStage{
							source:  nestedJoin,
							matcher: &SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLColumnExpr{tblTwo, "a", columnType}},
						}

						o, err := optimizeCrossJoins(filter)
						So(err, ShouldBeNil)

						So(o, ShouldResemble, &JoinStage{
							left: &JoinStage{
								left:    msOne,
								right:   msTwo,
								kind:    InnerJoin,
								matcher: &SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLColumnExpr{tblTwo, "a", columnType}},
							},
							right: msThree,
							kind:  CrossJoin,
						})
					})

					Convey("Should optimize a filter with partially applicable criteria", func() {
						filter := &FilterStage{
							source: nestedJoin,
							matcher: &SQLAndExpr{
								&SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLInt(10)},
								&SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLColumnExpr{tblTwo, "a", columnType}},
							},
						}

						o, err := optimizeCrossJoins(filter)
						So(err, ShouldBeNil)

						So(o, ShouldResemble, &FilterStage{
							source: &JoinStage{
								left: &JoinStage{
									left:    msOne,
									right:   msTwo,
									kind:    InnerJoin,
									matcher: &SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLColumnExpr{tblTwo, "a", columnType}},
								},
								right: msThree,
								kind:  CrossJoin,
							},
							matcher: &SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLInt(10)},
						})
					})
				})
			})
		}
	})
}

func TestGroupByPushDown(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne
	columnType := schema.ColumnType{schema.SQLInt, schema.MongoInt}

	Convey("Subject: GroupBy Optimization", t, func() {

		ctx := &ExecutionCtx{
			PlanCtx: &PlanCtx{
				Schema: cfgOne,
				Db:     dbOne,
			},
		}

		tbl := "foo"

		Convey("Given a group by clause that can be pushed down", func() {

			ms, err := NewMongoSourceStage(ctx.PlanCtx, dbOne, tbl, "")
			So(err, ShouldBeNil)

			gb := &GroupByStage{
				source: ms,
			}

			Convey("Should optimize 'select a, b from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"a": SQLColumnExpr{tbl, "a", columnType},
					"b": SQLColumnExpr{tbl, "b", columnType},
					"c": SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "b")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"foo_DOT_a": bson.M{
								"$first": "$a",
							},
							"foo_DOT_b": bson.M{
								"$first": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id":       0,
							"foo_DOT_a": "$foo_DOT_a",
							"foo_DOT_b": "$foo_DOT_b",
						}}}})
			})

			Convey("Should optimize 'select a, b, c from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"a": SQLColumnExpr{tbl, "a", columnType},
					"b": SQLColumnExpr{tbl, "b", columnType},
					"c": SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "b", "c")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"foo_DOT_a": bson.M{
								"$first": "$a",
							},
							"foo_DOT_b": bson.M{
								"$first": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id":       0,
							"foo_DOT_a": "$foo_DOT_a",
							"foo_DOT_b": "$foo_DOT_b",
							"foo_DOT_c": "$_id.foo_DOT_c",
						}}}})
			})

			Convey("Should optimize 'select a, b, c as Awesome from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"a":       SQLColumnExpr{tbl, "a", columnType},
					"b":       SQLColumnExpr{tbl, "b", columnType},
					"c":       SQLColumnExpr{tbl, "c", columnType},
					"Awesome": SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "b", "Awesome")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"foo_DOT_a": bson.M{
								"$first": "$a",
							},
							"foo_DOT_b": bson.M{
								"$first": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id":       0,
							"foo_DOT_a": "$foo_DOT_a",
							"foo_DOT_b": "$foo_DOT_b",
							"foo_DOT_c": "$_id.foo_DOT_c",
						}}}})
			})

			Convey("Should optimize 'select a, b, c + a as Awesome from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"a":       SQLColumnExpr{tbl, "a", columnType},
					"b":       SQLColumnExpr{tbl, "b", columnType},
					"c":       SQLColumnExpr{tbl, "c", columnType},
					"Awesome": &SQLAddExpr{SQLColumnExpr{tbl, "c", columnType}, SQLColumnExpr{tbl, "a", columnType}},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "b", "Awesome")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"foo_DOT_a": bson.M{
								"$first": "$a",
							},
							"foo_DOT_b": bson.M{
								"$first": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id":                 0,
							"foo_DOT_a":           "$foo_DOT_a",
							"foo_DOT_b":           "$foo_DOT_b",
							"foo_DOT_c+foo_DOT_a": bson.M{"$add": []interface{}{"$_id.foo_DOT_c", "$foo_DOT_a"}},
						}}}})
			})

			Convey("Should optimize 'select sum(a), sum(b) from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"sum(a)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLColumnExpr{tbl, "a", columnType}}},
					"sum(b)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLColumnExpr{tbl, "b", columnType}}},
					"c":      SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "sum(a)", "sum(b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"sum(foo_DOT_a)": bson.M{
								"$sum": "$a",
							},
							"sum(foo_DOT_b)": bson.M{
								"$sum": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id":            0,
							"sum(foo_DOT_a)": "$sum(foo_DOT_a)",
							"sum(foo_DOT_b)": "$sum(foo_DOT_b)",
						}}}})
			})

			Convey("Should optimize 'select c, sum(a), sum(b) from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"sum(a)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLColumnExpr{tbl, "a", columnType}}},
					"sum(b)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLColumnExpr{tbl, "b", columnType}}},
					"c":      SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "c", "sum(a)", "sum(b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"sum(foo_DOT_a)": bson.M{
								"$sum": "$a",
							},
							"sum(foo_DOT_b)": bson.M{
								"$sum": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id":            0,
							"foo_DOT_c":      "$_id.foo_DOT_c",
							"sum(foo_DOT_a)": "$sum(foo_DOT_a)",
							"sum(foo_DOT_b)": "$sum(foo_DOT_b)",
						}}}})
			})

			Convey("Should optimize 'select a, sum(b) from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"a":      SQLColumnExpr{tbl, "a", columnType},
					"sum(b)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLColumnExpr{tbl, "b", columnType}}},
					"c":      SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "sum(b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"foo_DOT_a": bson.M{
								"$first": "$a",
							},
							"sum(foo_DOT_b)": bson.M{
								"$sum": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id":            0,
							"foo_DOT_a":      "$foo_DOT_a",
							"sum(foo_DOT_b)": "$sum(foo_DOT_b)",
						}}}})
			})

			Convey("Should optimize 'select a, sum(distinct b) from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"a":               SQLColumnExpr{tbl, "a", columnType},
					"sum(distinct b)": &SQLAggFunctionExpr{"sum", true, []SQLExpr{SQLColumnExpr{tbl, "b", columnType}}},
					"c":               SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "sum(distinct b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"foo_DOT_a": bson.M{
								"$first": "$a",
							},
							"distinct foo_DOT_b": bson.M{
								"$addToSet": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id":                     0,
							"foo_DOT_a":               "$foo_DOT_a",
							"sum(distinct foo_DOT_b)": bson.M{"$sum": "$distinct foo_DOT_b"},
						}}}})
			})

			Convey("Should optimize 'select a, sum(distinct b), c from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"a":               SQLColumnExpr{tbl, "a", columnType},
					"sum(distinct b)": &SQLAggFunctionExpr{"sum", true, []SQLExpr{SQLColumnExpr{tbl, "b", columnType}}},
					"c":               SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "sum(distinct b)", "c")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"foo_DOT_a": bson.M{
								"$first": "$a",
							},
							"distinct foo_DOT_b": bson.M{
								"$addToSet": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id":                     0,
							"foo_DOT_c":               "$_id.foo_DOT_c",
							"foo_DOT_a":               "$foo_DOT_a",
							"sum(distinct foo_DOT_b)": bson.M{"$sum": "$distinct foo_DOT_b"},
						}}}})
			})

			Convey("Should optimize 'select a + sum(b) from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"a + sum(b)": &SQLAddExpr{SQLColumnExpr{tbl, "a", columnType}, &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLColumnExpr{tbl, "b", columnType}}}},
					"c":          SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a + sum(b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"foo_DOT_a": bson.M{
								"$first": "$a",
							},
							"sum(foo_DOT_b)": bson.M{
								"$sum": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id": 0,
							"foo_DOT_a+sum(foo_DOT_b)": bson.M{"$add": []interface{}{"$foo_DOT_a", "$sum(foo_DOT_b)"}},
						}}}})
			})

			Convey("Should optimize 'select a + b from foo group by a + b'", func() {

				exprs := map[string]SQLExpr{
					"a + b": &SQLAddExpr{SQLColumnExpr{tbl, "a", columnType}, SQLColumnExpr{tbl, "b", columnType}},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a + b")
				gb.keyExprs = constructSelectExpressions(exprs, "a + b")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{"foo_DOT_a+foo_DOT_b": bson.M{"$add": []interface{}{"$a", "$b"}}},
						}}},
						bson.D{{"$project", bson.M{
							"_id": 0,
							"foo_DOT_a+foo_DOT_b": "$_id.foo_DOT_a+foo_DOT_b",
						}}}})
			})

			Convey("Should optimize 'select a + c + sum(b) from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"a + c + sum(b)": &SQLAddExpr{SQLColumnExpr{tbl, "a", columnType}, &SQLAddExpr{SQLColumnExpr{tbl, "c", columnType}, &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLColumnExpr{tbl, "b", columnType}}}}},
					"c":              SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a + c + sum(b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"foo_DOT_a": bson.M{
								"$first": "$a",
							},
							"sum(foo_DOT_b)": bson.M{
								"$sum": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id": 0,
							"foo_DOT_a+foo_DOT_c+sum(foo_DOT_b)": bson.M{"$add": []interface{}{"$foo_DOT_a", bson.M{"$add": []interface{}{"$_id.foo_DOT_c", "$sum(foo_DOT_b)"}}}},
						}}}})
			})

			Convey("Should optimize 'select a + sum(b) as Awesome from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"Awesome": &SQLAddExpr{SQLColumnExpr{tbl, "a", columnType}, &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLColumnExpr{tbl, "b", columnType}}}},
					"c":       SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "Awesome")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"foo_DOT_a": bson.M{
								"$first": "$a",
							},
							"sum(foo_DOT_b)": bson.M{
								"$sum": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id": 0,
							"foo_DOT_a+sum(foo_DOT_b)": bson.M{"$add": []interface{}{"$foo_DOT_a", "$sum(foo_DOT_b)"}},
						}}}})
			})

			Convey("Should optimize 'select a + sum(distinct b) from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"a + sum(distinct b)": &SQLAddExpr{SQLColumnExpr{tbl, "a", columnType}, &SQLAggFunctionExpr{"sum", true, []SQLExpr{SQLColumnExpr{tbl, "b", columnType}}}},
					"c": SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a + sum(distinct b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"foo_DOT_a": bson.M{
								"$first": "$a",
							},
							"distinct foo_DOT_b": bson.M{
								"$addToSet": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id": 0,
							"foo_DOT_a+sum(distinct foo_DOT_b)": bson.M{"$add": []interface{}{"$foo_DOT_a", bson.M{"$sum": "$distinct foo_DOT_b"}}},
						}}}})
			})

			Convey("Should optimize 'select c + sum(distinct b) from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"c + sum(distinct b)": &SQLAddExpr{SQLColumnExpr{tbl, "c", columnType}, &SQLAggFunctionExpr{"sum", true, []SQLExpr{SQLColumnExpr{tbl, "b", columnType}}}},
					"c": SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "c + sum(distinct b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"distinct foo_DOT_b": bson.M{
								"$addToSet": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id": 0,
							"foo_DOT_c+sum(distinct foo_DOT_b)": bson.M{"$add": []interface{}{"$_id.foo_DOT_c", bson.M{"$sum": "$distinct foo_DOT_b"}}},
						}}}})
			})

			Convey("Should optimize 'select sum(distinct a+b) from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"sum(distinct a+b)": &SQLAggFunctionExpr{"sum", true, []SQLExpr{&SQLAddExpr{SQLColumnExpr{tbl, "a", columnType}, SQLColumnExpr{tbl, "b", columnType}}}},
					"c":                 SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "sum(distinct a+b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"distinct foo_DOT_a+foo_DOT_b": bson.M{
								"$addToSet": bson.M{"$add": []interface{}{"$a", "$b"}},
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id": 0,
							"sum(distinct foo_DOT_a+foo_DOT_b)": bson.M{
								"$sum": "$distinct foo_DOT_a+foo_DOT_b"},
						}}}})
			})

			Convey("Should optimize 'select a+sum(distinct a+b) from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"a+sum(distinct a+b)": &SQLAddExpr{SQLColumnExpr{tbl, "a", columnType}, &SQLAggFunctionExpr{"sum", true, []SQLExpr{&SQLAddExpr{SQLColumnExpr{tbl, "a", columnType}, SQLColumnExpr{tbl, "b", columnType}}}}},
					"c": SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a+sum(distinct a+b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"foo_DOT_a": bson.M{
								"$first": "$a",
							},
							"distinct foo_DOT_a+foo_DOT_b": bson.M{
								"$addToSet": bson.M{"$add": []interface{}{"$a", "$b"}},
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id": 0,
							"foo_DOT_a+sum(distinct foo_DOT_a+foo_DOT_b)": bson.M{"$add": []interface{}{"$foo_DOT_a", bson.M{"$sum": "$distinct foo_DOT_a+foo_DOT_b"}}},
						}}}})
			})

			Convey("Should optimize 'select a, e from foo group by f'", func() {

				exprs := map[string]SQLExpr{
					"a": SQLColumnExpr{tbl, "a", columnType},
					"e": SQLColumnExpr{tbl, "e", columnType},
					"f": SQLColumnExpr{tbl, "f", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "e")
				gb.keyExprs = constructSelectExpressions(exprs, "f")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_f": "$d.f",
							},
							"foo_DOT_a": bson.M{
								"$first": "$a",
							},
							"foo_DOT_e": bson.M{
								"$first": "$d.e",
							}}}},
						bson.D{{"$project", bson.M{
							"_id":       0,
							"foo_DOT_a": "$foo_DOT_a",
							"foo_DOT_e": "$foo_DOT_e",
						}}}})
			})

			Convey("Should optimize 'select a, sum(distinct e) from foo group by f'", func() {

				exprs := map[string]SQLExpr{
					"a":               SQLColumnExpr{tbl, "a", columnType},
					"sum(distinct e)": &SQLAggFunctionExpr{"sum", true, []SQLExpr{SQLColumnExpr{tbl, "e", columnType}}},
					"f":               SQLColumnExpr{tbl, "f", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "sum(distinct e)")
				gb.keyExprs = constructSelectExpressions(exprs, "f")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_f": "$d.f",
							},
							"foo_DOT_a": bson.M{
								"$first": "$a",
							},
							"distinct foo_DOT_e": bson.M{
								"$addToSet": "$d.e",
							}}}},
						bson.D{{"$project", bson.M{
							"_id":                     0,
							"foo_DOT_a":               "$foo_DOT_a",
							"sum(distinct foo_DOT_e)": bson.M{"$sum": "$distinct foo_DOT_e"},
						}}}})
			})

			Convey("Should optimize 'select count(*) from foo'", func() {

				exprs := map[string]SQLExpr{
					"count(*)": &SQLAggFunctionExpr{"count", false, []SQLExpr{SQLVarchar("*")}},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "count(*)")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id":      bson.M{},
							"count(*)": bson.M{"$sum": 1},
						}}},
						bson.D{{"$project", bson.M{
							"_id":      0,
							"count(*)": "$count(*)",
						}}}})
			})

			Convey("Should optimize 'select count(a) from foo'", func() {

				exprs := map[string]SQLExpr{
					"count(a)": &SQLAggFunctionExpr{"count", false, []SQLExpr{SQLColumnExpr{tbl, "a", columnType}}},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "count(a)")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{},
							"count(foo_DOT_a)": bson.M{
								"$sum": bson.M{
									"$cond": []interface{}{
										bson.M{
											"$eq": []interface{}{
												bson.M{
													"$ifNull": []interface{}{
														"$a",
														nil,
													},
												},
												nil,
											},
										},
										0,
										1,
									},
								},
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id":              0,
							"count(foo_DOT_a)": "$count(foo_DOT_a)",
						}}}})
			})

			Convey("Should optimize 'select a, count(distinct b) from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"a":                 SQLColumnExpr{tbl, "a", columnType},
					"count(distinct b)": &SQLAggFunctionExpr{"count", true, []SQLExpr{SQLColumnExpr{tbl, "b", columnType}}},
					"c":                 SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "count(distinct b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx.PlanCtx, gb,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"foo_DOT_a":          bson.M{"$first": "$a"},
							"distinct foo_DOT_b": bson.M{"$addToSet": "$b"},
						}}},
						bson.D{{"$project", bson.M{
							"_id":       0,
							"foo_DOT_a": "$foo_DOT_a",
							"count(distinct foo_DOT_b)": bson.M{
								"$sum": bson.M{
									"$map": bson.M{
										"input": "$distinct foo_DOT_b",
										"as":    "i",
										"in": bson.M{
											"$cond": []interface{}{
												bson.M{"$eq": []interface{}{
													bson.M{"$ifNull": []interface{}{"$$i", nil}},
													nil}},
												0,
												1,
											}}}}}}}}})
			})
		})
	})
}

func TestHavingPushDown(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne
	columnType := schema.ColumnType{schema.SQLInt, schema.MongoInt}

	// This is effectively the same as filter, except it always happens after a group by, so we
	// are really just testing that group by did the right thing related to columns...

	Convey("Subject: Having Optimization", t, func() {

		ctx := &ExecutionCtx{
			PlanCtx: &PlanCtx{
				Schema: cfgOne,
				Db:     dbOne,
			},
		}

		tbl := "foo"

		Convey("Given a group by clause that can be pushed down", func() {

			ms, err := NewMongoSourceStage(ctx.PlanCtx, dbOne, tbl, "")
			So(err, ShouldBeNil)

			gb := &GroupByStage{
				source: ms,
			}

			having := &FilterStage{
				source: gb,
			}

			Convey("Should optimize 'select sum(a) from bar group by c having sum(b) = 10'", func() {

				exprs := map[string]SQLExpr{
					"sum(a)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLColumnExpr{tbl, "a", columnType}}},
					"sum(b)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLColumnExpr{tbl, "b", columnType}}},
					"c":      SQLColumnExpr{tbl, "c", columnType},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "sum(a)", "sum(b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")
				having.matcher = &SQLEqualsExpr{SQLColumnExpr{"", "sum(foo.b)", columnType}, SQLInt(10)}

				verifyOptimizedPipeline(ctx.PlanCtx, having,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_c": "$c",
							},
							"sum(foo_DOT_a)": bson.M{
								"$sum": "$a",
							},
							"sum(foo_DOT_b)": bson.M{
								"$sum": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id":            0,
							"sum(foo_DOT_a)": "$sum(foo_DOT_a)",
							"sum(foo_DOT_b)": "$sum(foo_DOT_b)",
						}}},
						bson.D{{"$match", bson.M{
							"sum(foo_DOT_b)": int64(10),
						}}}})
			})
		})
	})
}

func TestJoinPushDown(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne
	columnType := schema.ColumnType{schema.SQLInt, schema.MongoInt}

	Convey("Subject: Join Optimization", t, func() {

		ctx := &ExecutionCtx{
			PlanCtx: &PlanCtx{
				Schema: cfgOne,
				Db:     dbOne,
			},
		}

		tblOne := "foo"
		tblTwo := "bar"

		msOne, err := NewMongoSourceStage(ctx.PlanCtx, dbOne, tblOne, "")
		So(err, ShouldBeNil)

		msTwo, err := NewMongoSourceStage(ctx.PlanCtx, dbOne, tblTwo, "")
		So(err, ShouldBeNil)

		Convey("Given a push-downable inner join", func() {

			join := &JoinStage{
				left:  msOne,
				right: msTwo,
			}

			Convey("Should optimize an inner join with an equality comparison", func() {
				join.matcher = &SQLEqualsExpr{SQLColumnExpr{tblOne, "c", columnType}, SQLColumnExpr{tblTwo, "b", columnType}}

				verifyOptimizedPipeline(ctx.PlanCtx, join,
					[]bson.D{
						bson.D{{"$lookup", bson.M{
							"from":         tblTwo,
							"localField":   "c",
							"foreignField": "b",
							"as":           joinedFieldNamePrefix + tblTwo,
						}}},
						bson.D{{"$unwind", bson.M{
							"path": "$" + joinedFieldNamePrefix + tblTwo,
							"preserveNullAndEmptyArrays": false,
						}}}})
			})

			Convey("Should optimize an inner join with a flipped equality comparison", func() {
				join.matcher = &SQLEqualsExpr{SQLColumnExpr{tblTwo, "b", columnType}, SQLColumnExpr{tblOne, "c", columnType}}

				verifyOptimizedPipeline(ctx.PlanCtx, join,
					[]bson.D{
						bson.D{{"$lookup", bson.M{
							"from":         tblTwo,
							"localField":   "c",
							"foreignField": "b",
							"as":           joinedFieldNamePrefix + tblTwo,
						}}},
						bson.D{{"$unwind", bson.M{
							"path": "$" + joinedFieldNamePrefix + tblTwo,
							"preserveNullAndEmptyArrays": false,
						}}}})
			})

			Convey("Should optimize an inner join with an equality comparison and other criteria", func() {
				join.matcher = &SQLAndExpr{
					&SQLEqualsExpr{SQLColumnExpr{tblOne, "c", columnType}, SQLColumnExpr{tblTwo, "b", columnType}},
					&SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLInt(10)}}

				verifyOptimizedPipeline(ctx.PlanCtx, join,
					[]bson.D{
						bson.D{{"$lookup", bson.M{
							"from":         tblTwo,
							"localField":   "c",
							"foreignField": "b",
							"as":           joinedFieldNamePrefix + tblTwo,
						}}},
						bson.D{{"$unwind", bson.M{
							"path": "$" + joinedFieldNamePrefix + tblTwo,
							"preserveNullAndEmptyArrays": false,
						}}},
						bson.D{{"$match", bson.M{"a": int64(10)}}}})
			})

			Convey("Should optimize an inner join with criteria on the local table and an equality comparison", func() {
				join.matcher = &SQLAndExpr{
					&SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLInt(10)},
					&SQLEqualsExpr{SQLColumnExpr{tblOne, "c", columnType}, SQLColumnExpr{tblTwo, "b", columnType}}}

				verifyOptimizedPipeline(ctx.PlanCtx, join,
					[]bson.D{
						bson.D{{"$lookup", bson.M{
							"from":         tblTwo,
							"localField":   "c",
							"foreignField": "b",
							"as":           joinedFieldNamePrefix + tblTwo,
						}}},
						bson.D{{"$unwind", bson.M{
							"path": "$" + joinedFieldNamePrefix + tblTwo,
							"preserveNullAndEmptyArrays": false,
						}}},
						bson.D{{"$match", bson.M{"a": int64(10)}}}})
			})

			Convey("Should optimize an inner join with criteria on the foreign table and an equality comparison", func() {
				join.matcher = &SQLAndExpr{
					&SQLEqualsExpr{SQLColumnExpr{tblTwo, "a", columnType}, SQLInt(10)},
					&SQLEqualsExpr{SQLColumnExpr{tblOne, "c", columnType}, SQLColumnExpr{tblTwo, "b", columnType}}}

				verifyOptimizedPipeline(ctx.PlanCtx, join,
					[]bson.D{
						bson.D{{"$lookup", bson.M{
							"from":         tblTwo,
							"localField":   "c",
							"foreignField": "b",
							"as":           joinedFieldNamePrefix + tblTwo,
						}}},
						bson.D{{"$unwind", bson.M{
							"path": "$" + joinedFieldNamePrefix + tblTwo,
							"preserveNullAndEmptyArrays": false,
						}}},
						bson.D{{"$match", bson.M{"__joined_bar.a": int64(10)}}}})
			})
		})

		Convey("Given a push-downable left join", func() {

			join := &JoinStage{
				kind:  LeftJoin,
				left:  msOne,
				right: msTwo,
			}

			Convey("Should optimize a left join with an equality comparison", func() {
				join.matcher = &SQLEqualsExpr{SQLColumnExpr{tblOne, "c", columnType}, SQLColumnExpr{tblTwo, "b", columnType}}

				verifyOptimizedPipeline(ctx.PlanCtx, join,
					[]bson.D{
						bson.D{{"$lookup", bson.M{
							"from":         tblTwo,
							"localField":   "c",
							"foreignField": "b",
							"as":           joinedFieldNamePrefix + tblTwo,
						}}},
						bson.D{{"$unwind", bson.M{
							"path": "$" + joinedFieldNamePrefix + tblTwo,
							"preserveNullAndEmptyArrays": true,
						}}}})
			})

			Convey("Should optimize a left join with a flipped equality comparison", func() {
				join.matcher = &SQLEqualsExpr{SQLColumnExpr{tblTwo, "b", columnType}, SQLColumnExpr{tblOne, "c", columnType}}

				verifyOptimizedPipeline(ctx.PlanCtx, join,
					[]bson.D{
						bson.D{{"$lookup", bson.M{
							"from":         tblTwo,
							"localField":   "c",
							"foreignField": "b",
							"as":           joinedFieldNamePrefix + tblTwo,
						}}},
						bson.D{{"$unwind", bson.M{
							"path": "$" + joinedFieldNamePrefix + tblTwo,
							"preserveNullAndEmptyArrays": true,
						}}}})
			})

			Convey("Should optimize a left join with an equality comparison and other criteria", func() {
				join.matcher = &SQLAndExpr{
					&SQLEqualsExpr{SQLColumnExpr{tblOne, "c", columnType}, SQLColumnExpr{tblTwo, "b", columnType}},
					&SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLColumnExpr{tblTwo, "a", columnType}}}

				join.kind = InnerJoin
				join.matcher = &SQLEqualsExpr{SQLColumnExpr{tblOne, "c", columnType}, SQLColumnExpr{tblTwo, "b", columnType}}

				verifyOptimizedPipeline(ctx.PlanCtx, join,
					[]bson.D{
						bson.D{{"$lookup", bson.M{
							"from":         tblTwo,
							"localField":   "c",
							"foreignField": "b",
							"as":           joinedFieldNamePrefix + tblTwo,
						}}},
						bson.D{{"$unwind", bson.M{
							"path": "$" + joinedFieldNamePrefix + tblTwo,
							"preserveNullAndEmptyArrays": true,
						}}},
						bson.D{{"$project", bson.M{
							"a":   1,
							"b":   1,
							"c":   1,
							"d.e": 1,
							"d.f": 1,
							"_id": 1,
							"__joined_bar": bson.M{
								"$cond": bson.M{
									"if":   bson.M{"$eq": []interface{}{"$a", "$__joined_bar.a"}},
									"then": "$__joined_bar",
									"else": nil,
								}}}}}})
			})

			Convey("Should optimize a left join with criteria and an equality comparison", func() {
				join.matcher = &SQLAndExpr{
					&SQLEqualsExpr{SQLColumnExpr{tblOne, "a", columnType}, SQLInt(10)},
					&SQLEqualsExpr{SQLColumnExpr{tblOne, "c", columnType}, SQLColumnExpr{tblTwo, "b", columnType}}}

				msTwo.pipeline = append(msTwo.pipeline, bson.D{{"$test", 1}})

				join.kind = RightJoin
				join.matcher = &SQLEqualsExpr{SQLColumnExpr{tblOne, "c", columnType}, SQLColumnExpr{tblTwo, "b", columnType}}

				verifyOptimizedPipeline(ctx.PlanCtx, join,
					[]bson.D{
						bson.D{{"$lookup", bson.M{
							"from":         tblTwo,
							"localField":   "c",
							"foreignField": "b",
							"as":           joinedFieldNamePrefix + tblTwo,
						}}},
						bson.D{{"$unwind", bson.M{
							"path": "$" + joinedFieldNamePrefix + tblTwo,
							"preserveNullAndEmptyArrays": true,
						}}},
						bson.D{{"$project", bson.M{
							"a":   1,
							"b":   1,
							"c":   1,
							"d.e": 1,
							"d.f": 1,
							"_id": 1,
							"__joined_bar": bson.M{
								"$cond": bson.M{
									"if":   bson.M{"$eq": []interface{}{"$a", bson.M{"$literal": SQLInt(10)}}},
									"then": "$__joined_bar",
									"else": nil}}}}}})
			})
		})

		Convey("Given a non-push-downable join", func() {

			join := &JoinStage{
				left:  msOne,
				right: msTwo,
			}

			Convey("Should not optimize the pipeline when the left source is not a TableScan", func() {
				join.left = &EmptyStage{nil}
				verifyUnoptimizedPipeline(ctx.PlanCtx, join)
			})

			Convey("Should not optimize the pipeline when the right source is not a TableScan", func() {
				join.right = &EmptyStage{nil}
				verifyUnoptimizedPipeline(ctx.PlanCtx, join)
			})

			Convey("Should not optimize a right join", func() {
				join.kind = RightJoin
				join.matcher = &SQLEqualsExpr{SQLColumnExpr{tblOne, "c", columnType}, SQLColumnExpr{tblTwo, "b", columnType}}

				verifyUnoptimizedPipeline(ctx.PlanCtx, join)
			})

			for _, kind := range []JoinKind{InnerJoin, LeftJoin} {
				Convey(fmt.Sprintf("Should not optimize a %v when the on clause does not contain an equality comparison", kind), func() {
					join.kind = kind
					join.matcher = &SQLGreaterThanExpr{SQLColumnExpr{tblOne, "c", columnType}, SQLColumnExpr{tblTwo, "b", columnType}}
					verifyUnoptimizedPipeline(ctx.PlanCtx, join)
				})

				Convey(fmt.Sprintf("Should not optimize a %v when the on clause does not contain fields from both sides", kind), func() {
					join.kind = kind
					join.matcher = &SQLEqualsExpr{SQLColumnExpr{tblOne, "c", columnType}, SQLColumnExpr{tblOne, "b", columnType}}
					verifyUnoptimizedPipeline(ctx.PlanCtx, join)
				})

				Convey(fmt.Sprintf("Should not optimize a %v when the right side already has a pipeline", kind), func() {
					msTwo.pipeline = append(msTwo.pipeline, bson.D{{"$test", 1}})

					join.kind = kind
					join.matcher = &SQLEqualsExpr{SQLColumnExpr{tblOne, "c", columnType}, SQLColumnExpr{tblTwo, "b", columnType}}
					verifyUnoptimizedPipeline(ctx.PlanCtx, join)
				})
			}
		})
	})
}

func TestLimitPushDown(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne

	Convey("Subject: Limit Optimization", t, func() {

		ctx := &ExecutionCtx{
			PlanCtx: &PlanCtx{
				Schema: cfgOne,
				Db:     dbOne,
			},
		}

		tbl := "foo"

		Convey("Given a push-downable limit", func() {

			ms, err := NewMongoSourceStage(ctx.PlanCtx, dbOne, tbl, "")
			So(err, ShouldBeNil)

			limit := &LimitStage{
				source: ms,
			}

			Convey("Should optimize with only an offset", func() {

				limit.offset = 22

				verifyOptimizedPipeline(ctx.PlanCtx, limit,
					[]bson.D{bson.D{{"$skip", limit.offset}}})

			})

			Convey("Should optimized with only a rowcount", func() {

				limit.limit = 20

				verifyOptimizedPipeline(ctx.PlanCtx, limit,
					[]bson.D{bson.D{{"$limit", limit.limit}}})
			})

			Convey("Should optimized with both an offset and a rowcount", func() {

				limit.offset = 22
				limit.limit = 20

				verifyOptimizedPipeline(ctx.PlanCtx, limit, []bson.D{
					bson.D{{"$skip", limit.offset}},
					bson.D{{"$limit", limit.limit}}})
			})
		})

		Convey("Given a non-push-downable limit", func() {

			empty := &EmptyStage{nil}

			limit := &LimitStage{
				source: empty,
			}

			Convey("Should not optimized the pipeline", func() {
				verifyUnoptimizedPipeline(ctx.PlanCtx, limit)
			})
		})
	})
}

func TestProjectPushdown(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne
	columnType := schema.ColumnType{schema.SQLInt, schema.MongoInt}

	Convey("Subject: Project Optimization", t, func() {

		ctx := &ExecutionCtx{
			PlanCtx: &PlanCtx{
				Schema: cfgOne,
				Db:     dbOne,
			},
		}

		tbl := "foo"
		ms, err := NewMongoSourceStage(ctx.PlanCtx, dbOne, tbl, "")
		So(err, ShouldBeNil)
		Convey("given a push-downable project", func() {
			exprs := map[string]SQLExpr{
				"a":      SQLColumnExpr{tbl, "a", columnType},
				"sum(b)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLColumnExpr{tbl, "b", columnType}}},
				"c":      SQLColumnExpr{tbl, "c", columnType},
			}

			proj := &ProjectStage{
				sExprs: constructSelectExpressions(exprs, "a", "sum(b)"),
				source: ms,
			}

			Convey("The pipeline should contain a $project stage with correct field mappings", func() {
				verifyOptimizedPipeline(ctx.PlanCtx, proj,
					[]bson.D{
						{{
							"$project", bson.M{
								"foo_DOT_a":      "$a",
								"sum(foo_DOT_b)": bson.M{"$sum": "$b"},
							},
						}},
					},
				)
			})
		})

		Convey("given a partially-push-downable project", func() {
			exprs := map[string]SQLExpr{
				"a": SQLColumnExpr{tbl, "a", columnType},
				"ascii(substring('xxx',b))": &SQLScalarFunctionExpr{
					"ascii", []SQLExpr{
						&SQLScalarFunctionExpr{
							"substring", []SQLExpr{
								SQLVarchar("xxx"),
								SQLColumnExpr{tbl, "b", columnType},
							},
						},
					},
				},
				"sum(c)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLColumnExpr{tbl, "c", columnType}}},
			}

			proj := &ProjectStage{
				sExprs: constructSelectExpressions(exprs, "a", "ascii(substring('xxx',b))", "sum(c)"),
				source: ms,
			}
			Convey("the project node should not be removed from query plan tree", func() {
				optimized, err := OptimizePlan(ctx.PlanCtx, proj)
				So(err, ShouldBeNil)
				So(optimized, ShouldHaveSameTypeAs, (*ProjectStage)(nil))

				optimizedProject := optimized.(*ProjectStage)
				So(optimizedProject.source, ShouldHaveSameTypeAs, (*MongoSourceStage)(nil))

				Convey("the pipeline should contain all fields required to compute projection", func() {
					mongoSource := optimizedProject.source.(*MongoSourceStage)
					So(mongoSource.pipeline, ShouldResemble,
						[]bson.D{
							{{
								"$project", bson.M{
									"foo_DOT_a":      "$a",                 // pushed-down
									"b":              "$b",                 // NOT pushed down, but needed for ascii(substring(...))
									"sum(foo_DOT_c)": bson.M{"$sum": "$c"}, // pushed down
								},
							}},
						},
					)
				})

			})

		})
	})
}

func TestOrderByPushDown(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne
	columnType := schema.ColumnType{schema.SQLInt, schema.MongoInt}

	Convey("Subject: Order By Optimization", t, func() {

		ctx := &ExecutionCtx{
			PlanCtx: &PlanCtx{
				Schema: cfgOne,
				Db:     dbOne,
			},
		}

		tbl := "foo"

		Convey("Given a push-downable order by", func() {

			ms, err := NewMongoSourceStage(ctx.PlanCtx, dbOne, tbl, "")
			So(err, ShouldBeNil)

			orderBy := &OrderByStage{
				source: ms,
			}

			Convey("Should optimize order by with simple column references 'select a from foo order by a, b DESC, e'", func() {

				exprs := map[string]SQLExpr{
					"a": SQLColumnExpr{tbl, "a", columnType},
					"b": SQLColumnExpr{tbl, "b", columnType},
					"e": SQLColumnExpr{tbl, "e", columnType},
				}
				orderBy.keys = constructOrderByKeys(exprs, "a", "b", "e")

				verifyOptimizedPipeline(ctx.PlanCtx, orderBy,
					[]bson.D{bson.D{{"$sort", bson.D{
						{"a", 1},
						{"b", -1},
						{"d.e", 1},
					}}}})
			})

			Convey("Should optimize order by with aggregation expressions that have already been pushed down 'select a from foo group by a order by sum(b)'", func() {

				exprs := map[string]SQLExpr{
					"a":      SQLColumnExpr{tbl, "a", columnType},
					"sum(b)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLColumnExpr{tbl, "b", columnType}}},
				}

				groupBy := &GroupByStage{
					source:      ms,
					keyExprs:    constructSelectExpressions(exprs, "a"),
					selectExprs: constructSelectExpressions(exprs, "a", "sum(b)"),
				}

				orderBy.source = groupBy
				orderBy.keys = constructOrderByKeys(exprs, "sum(b)")

				verifyOptimizedPipeline(ctx.PlanCtx, orderBy,
					[]bson.D{
						bson.D{{"$group", bson.M{
							"_id": bson.M{
								"foo_DOT_a": "$a",
							},
							"sum(foo_DOT_b)": bson.M{
								"$sum": "$b",
							},
						}}},
						bson.D{{"$project", bson.M{
							"_id":            0,
							"foo_DOT_a":      "$_id.foo_DOT_a",
							"sum(foo_DOT_b)": "$sum(foo_DOT_b)",
						}}},
						bson.D{{"$sort", bson.D{
							{"sum(foo_DOT_b)", 1},
						}}}})
			})
		})

		Convey("Given a non-push-downable order by", func() {

			empty := &EmptyStage{nil}

			orderBy := &OrderByStage{
				source: empty,
			}

			Convey("Should not optimized the pipeline", func() {

				exprs := map[string]SQLExpr{
					"a": SQLColumnExpr{tbl, "a", columnType},
					"b": SQLColumnExpr{tbl, "b", columnType},
					"c": SQLColumnExpr{tbl, "c", columnType},
				}
				orderBy.keys = constructOrderByKeys(exprs, "a", "b", "c")

				verifyUnoptimizedPipeline(ctx.PlanCtx, orderBy)
			})
		})
	})
}

func verifyUnoptimizedPipeline(ctx *PlanCtx, op PlanStage) {
	optimized, err := OptimizePlan(ctx, op)
	So(err, ShouldBeNil)
	So(optimized, ShouldEqual, op)
}

func verifyOptimizedPipeline(ctx *PlanCtx, op PlanStage, pipeline []bson.D) {

	optimized, err := OptimizePlan(ctx, op)
	So(err, ShouldBeNil)

	ms, ok := optimized.(*MongoSourceStage)
	So(ok, ShouldBeTrue)

	So(ms.pipeline, ShouldResemble, pipeline)
}
