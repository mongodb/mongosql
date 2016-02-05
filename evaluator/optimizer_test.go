package evaluator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestOptimizeOperator(t *testing.T) {
	Convey("Subject: OptimizeOperator", t, func() {

		ctx := &ExecutionCtx{
			Schema: cfgOne,
			Db:     dbOne,
		}

		tbl := "foo"

		Convey("Given a recursively optimizable tree", func() {

			ts, err := NewTableScan(ctx, dbOne, tbl, "")
			So(err, ShouldBeNil)

			sa := &SourceAppend{
				source: ts,
			}

			Convey("Should optimize from bottom-up", func() {

				filter := &Filter{
					source:  sa,
					matcher: &SQLEqualsExpr{SQLFieldExpr{tbl, "a"}, SQLString("funny")},
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
					matcher: &SQLEqualsExpr{SQLFieldExpr{tbl, "a"}, SQLString("funny")},
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

		tbl := "foo"

		Convey("Given a push-downable filter", func() {

			ts, err := NewTableScan(ctx, dbOne, tbl, "")
			So(err, ShouldBeNil)

			sa := &SourceAppend{
				source: ts,
			}

			filter := &Filter{
				source: sa,
			}

			Convey("Should optimize when the matcher is fully translatable", func() {

				filter.matcher = &SQLEqualsExpr{SQLFieldExpr{tbl, "a"}, SQLString("funny")}

				verifyOptimizedPipeline(ctx, filter,
					[]bson.D{bson.D{{"$match", bson.M{"a": "funny"}}}})
			})

			Convey("Should optimize when the matcher is partially translatable", func() {

				filter.matcher = &SQLAndExpr{
					&SQLEqualsExpr{SQLFieldExpr{tbl, "a"}, SQLString("funny")},
					&SQLEqualsExpr{SQLFieldExpr{tbl, "b"}, SQLFieldExpr{tbl, "c"}}}

				optimized, err := OptimizeOperator(ctx, filter)
				So(err, ShouldBeNil)
				newFilter, ok := optimized.(*Filter)
				So(ok, ShouldBeTrue)
				So(newFilter.matcher, ShouldResemble, &SQLEqualsExpr{SQLFieldExpr{tbl, "b"}, SQLFieldExpr{tbl, "c"}})
				sa, ok := newFilter.source.(*SourceAppend)
				So(ok, ShouldBeTrue)
				ts, ok := sa.source.(*TableScan)
				So(ok, ShouldBeTrue)

				So(ts.pipeline, ShouldResemble, []bson.D{bson.D{{"$match", bson.M{"a": "funny"}}}})
			})
		})

		Convey("Given an immediately evaluated filter", func() {

			ts, err := NewTableScan(ctx, dbOne, tbl, "")
			So(err, ShouldBeNil)

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

				verifyOptimizedPipeline(ctx, filter, ts.pipeline)
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

				ts, err := NewTableScan(ctx, dbOne, tbl, "")
				So(err, ShouldBeNil)

				sa := &SourceAppend{
					source: ts,
				}

				filter := &Filter{
					source:  sa,
					matcher: &SQLEqualsExpr{SQLFieldExpr{tbl, "a"}, SQLFieldExpr{tbl, "b"}},
				}

				verifyUnoptimizedPipeline(ctx, filter)
			})
		})
	})
}

func TestGroupByPushDown(t *testing.T) {

	Convey("Subject: GroupBy Optimization", t, func() {

		ctx := &ExecutionCtx{
			Db:     dbOne,
			Schema: cfgOne,
		}

		tbl := "foo"

		Convey("Given a group by clause that can be pushed down", func() {

			ts, err := NewTableScan(ctx, dbOne, tbl, "")
			So(err, ShouldBeNil)

			sa := &SourceAppend{
				source: ts,
			}

			gb := &GroupBy{
				source: sa,
			}

			Convey("Should optimize 'select a, b from foo group by c'", func() {

				exprs := map[string]SQLExpr{
					"a": SQLFieldExpr{tbl, "a"},
					"b": SQLFieldExpr{tbl, "b"},
					"c": SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "b")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"a": SQLFieldExpr{tbl, "a"},
					"b": SQLFieldExpr{tbl, "b"},
					"c": SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "b", "c")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"a":       SQLFieldExpr{tbl, "a"},
					"b":       SQLFieldExpr{tbl, "b"},
					"c":       SQLFieldExpr{tbl, "c"},
					"Awesome": SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "b", "Awesome")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"a":       SQLFieldExpr{tbl, "a"},
					"b":       SQLFieldExpr{tbl, "b"},
					"c":       SQLFieldExpr{tbl, "c"},
					"Awesome": &SQLAddExpr{SQLFieldExpr{tbl, "c"}, SQLFieldExpr{tbl, "a"}},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "b", "Awesome")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"sum(a)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLFieldExpr{tbl, "a"}}},
					"sum(b)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLFieldExpr{tbl, "b"}}},
					"c":      SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "sum(a)", "sum(b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"sum(a)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLFieldExpr{tbl, "a"}}},
					"sum(b)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLFieldExpr{tbl, "b"}}},
					"c":      SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "c", "sum(a)", "sum(b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"a":      SQLFieldExpr{tbl, "a"},
					"sum(b)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLFieldExpr{tbl, "b"}}},
					"c":      SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "sum(b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"a":               SQLFieldExpr{tbl, "a"},
					"sum(distinct b)": &SQLAggFunctionExpr{"sum", true, []SQLExpr{SQLFieldExpr{tbl, "b"}}},
					"c":               SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "sum(distinct b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"a":               SQLFieldExpr{tbl, "a"},
					"sum(distinct b)": &SQLAggFunctionExpr{"sum", true, []SQLExpr{SQLFieldExpr{tbl, "b"}}},
					"c":               SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "sum(distinct b)", "c")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"a + sum(b)": &SQLAddExpr{SQLFieldExpr{tbl, "a"}, &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLFieldExpr{tbl, "b"}}}},
					"c":          SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a + sum(b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"a + b": &SQLAddExpr{SQLFieldExpr{tbl, "a"}, SQLFieldExpr{tbl, "b"}},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a + b")
				gb.keyExprs = constructSelectExpressions(exprs, "a + b")

				verifyOptimizedPipeline(ctx, gb,
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
					"a + c + sum(b)": &SQLAddExpr{SQLFieldExpr{tbl, "a"}, &SQLAddExpr{SQLFieldExpr{tbl, "c"}, &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLFieldExpr{tbl, "b"}}}}},
					"c":              SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a + c + sum(b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"Awesome": &SQLAddExpr{SQLFieldExpr{tbl, "a"}, &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLFieldExpr{tbl, "b"}}}},
					"c":       SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "Awesome")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"a + sum(distinct b)": &SQLAddExpr{SQLFieldExpr{tbl, "a"}, &SQLAggFunctionExpr{"sum", true, []SQLExpr{SQLFieldExpr{tbl, "b"}}}},
					"c": SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a + sum(distinct b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"c + sum(distinct b)": &SQLAddExpr{SQLFieldExpr{tbl, "c"}, &SQLAggFunctionExpr{"sum", true, []SQLExpr{SQLFieldExpr{tbl, "b"}}}},
					"c": SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "c + sum(distinct b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"sum(distinct a+b)": &SQLAggFunctionExpr{"sum", true, []SQLExpr{&SQLAddExpr{SQLFieldExpr{tbl, "a"}, SQLFieldExpr{tbl, "b"}}}},
					"c":                 SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "sum(distinct a+b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"a+sum(distinct a+b)": &SQLAddExpr{SQLFieldExpr{tbl, "a"}, &SQLAggFunctionExpr{"sum", true, []SQLExpr{&SQLAddExpr{SQLFieldExpr{tbl, "a"}, SQLFieldExpr{tbl, "b"}}}}},
					"c": SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a+sum(distinct a+b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
					"a": SQLFieldExpr{tbl, "a"},
					"e": SQLFieldExpr{tbl, "e"},
					"f": SQLFieldExpr{tbl, "f"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "e")
				gb.keyExprs = constructSelectExpressions(exprs, "f")

				verifyOptimizedPipeline(ctx, gb,
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
					"a":               SQLFieldExpr{tbl, "a"},
					"sum(distinct e)": &SQLAggFunctionExpr{"sum", true, []SQLExpr{SQLFieldExpr{tbl, "e"}}},
					"f":               SQLFieldExpr{tbl, "f"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "sum(distinct e)")
				gb.keyExprs = constructSelectExpressions(exprs, "f")

				verifyOptimizedPipeline(ctx, gb,
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
					"count(*)": &SQLAggFunctionExpr{"count", false, []SQLExpr{SQLString("*")}},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "count(*)")

				verifyOptimizedPipeline(ctx, gb,
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
					"count(a)": &SQLAggFunctionExpr{"count", false, []SQLExpr{SQLFieldExpr{tbl, "a"}}},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "count(a)")

				verifyOptimizedPipeline(ctx, gb,
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
					"a":                 SQLFieldExpr{tbl, "a"},
					"count(distinct b)": &SQLAggFunctionExpr{"count", true, []SQLExpr{SQLFieldExpr{tbl, "b"}}},
					"c":                 SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "a", "count(distinct b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")

				verifyOptimizedPipeline(ctx, gb,
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
	// This is effectively the same as filter, except it always happens after a group by, so we
	// are really just testing that group by did the right thing related to columns...

	Convey("Subject: Having Optimization", t, func() {

		ctx := &ExecutionCtx{
			Db:     dbOne,
			Schema: cfgOne,
		}

		tbl := "foo"

		Convey("Given a group by clause that can be pushed down", func() {

			ts, err := NewTableScan(ctx, dbOne, tbl, "")
			So(err, ShouldBeNil)

			sa := &SourceAppend{
				source: ts,
			}

			gb := &GroupBy{
				source: sa,
			}

			having := &Filter{
				source: gb,
			}

			Convey("Should optimize 'select sum(a) from bar group by c having sum(b) = 10'", func() {

				exprs := map[string]SQLExpr{
					"sum(a)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLFieldExpr{tbl, "a"}}},
					"sum(b)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLFieldExpr{tbl, "b"}}},
					"c":      SQLFieldExpr{tbl, "c"},
				}

				gb.selectExprs = constructSelectExpressions(exprs, "sum(a)", "sum(b)")
				gb.keyExprs = constructSelectExpressions(exprs, "c")
				having.matcher = &SQLEqualsExpr{SQLFieldExpr{"", "sum(foo.b)"}, SQLInt(10)}

				verifyOptimizedPipeline(ctx, having,
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

func TestLimitPushDown(t *testing.T) {

	Convey("Subject: Limit Optimization", t, func() {

		ctx := &ExecutionCtx{
			Schema: cfgOne,
			Db:     dbOne,
		}

		tbl := "foo"

		Convey("Given a push-downable limit", func() {

			ts, err := NewTableScan(ctx, dbOne, tbl, "")
			So(err, ShouldBeNil)

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

func TestOrderByPushDown(t *testing.T) {

	Convey("Subject: Order By Optimization", t, func() {

		ctx := &ExecutionCtx{
			Schema: cfgOne,
			Db:     dbOne,
		}

		tbl := "foo"

		Convey("Given a push-downable order by", func() {

			ts, err := NewTableScan(ctx, dbOne, tbl, "")
			So(err, ShouldBeNil)

			sa := &SourceAppend{
				source: ts,
			}

			orderBy := &OrderBy{
				source: sa,
			}

			Convey("Should optimize order by with simple column references 'select a from foo order by a, b DESC, e'", func() {

				exprs := map[string]SQLExpr{
					"a": SQLFieldExpr{tbl, "a"},
					"b": SQLFieldExpr{tbl, "b"},
					"e": SQLFieldExpr{tbl, "e"},
				}
				orderBy.keys = constructOrderByKeys(exprs, "a", "b", "e")

				verifyOptimizedPipeline(ctx, orderBy,
					[]bson.D{bson.D{{"$sort", bson.M{
						"a":   1,
						"b":   -1,
						"d.e": 1,
					}}}})
			})

			Convey("Should optimize order by with aggregation expressions that have already been pushed down 'select a from foo group by a order by sum(b)'", func() {

				exprs := map[string]SQLExpr{
					"a":      SQLFieldExpr{tbl, "a"},
					"sum(b)": &SQLAggFunctionExpr{"sum", false, []SQLExpr{SQLFieldExpr{tbl, "b"}}},
				}

				groupBy := &GroupBy{
					source:      sa,
					keyExprs:    constructSelectExpressions(exprs, "a"),
					selectExprs: constructSelectExpressions(exprs, "a"),
				}

				orderBy.source = groupBy
				orderBy.keys = constructOrderByKeys(exprs, "sum(b)")

				verifyOptimizedPipeline(ctx, orderBy,
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
						bson.D{{"$sort", bson.M{
							"sum(foo_DOT_b)": 1,
						}}}})
			})
		})

		Convey("Given a non-push-downable order by", func() {

			empty := &Empty{}

			orderBy := &OrderBy{
				source: empty,
			}

			Convey("Should not optimized the pipeline", func() {

				exprs := map[string]SQLExpr{
					"a": SQLFieldExpr{tbl, "a"},
					"b": SQLFieldExpr{tbl, "b"},
					"c": SQLFieldExpr{tbl, "c"},
				}
				orderBy.keys = constructOrderByKeys(exprs, "a", "b", "c")

				verifyUnoptimizedPipeline(ctx, orderBy)
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
