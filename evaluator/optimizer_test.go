package evaluator

import (
	"fmt"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/kr/pretty"
	. "github.com/smartystreets/goconvey/convey"
)

type pipelineGatherer struct {
	pipelines [][]bson.D
}

func (v *pipelineGatherer) visit(n node) (node, error) {
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	switch typedN := n.(type) {
	case *MongoSourceStage:
		if len(typedN.pipeline) > 0 {
			pipeline := make([]bson.D, len(typedN.pipeline))
			copy(pipeline, typedN.pipeline)
			v.pipelines = append(v.pipelines, pipeline)
		}
	}

	return n, nil
}

func TestOptimizePlan32(t *testing.T) {
	testSchema, err := schema.New(testSchema4)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}
	testInfo := getMongoDBInfo([]uint8{3, 2}, testSchema, mongodb.AllPrivileges)
	testVariables := createTestVariables(testInfo)
	testCatalog := getCatalogFromSchema(testSchema, testVariables)
	defaultDbName := "test"

	test := func(sql string, expected ...[]bson.D) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			plan, err := AlgebrizeQuery(statement, defaultDbName, testVariables, testCatalog)
			So(err, ShouldBeNil)
			actualPlan := OptimizePlan(createTestConnectionCtx(testInfo), plan)

			pg := &pipelineGatherer{}
			pg.visit(actualPlan)

			actual := pg.pipelines

			v := ShouldResembleDiffed(actual, expected)
			if v != "" {
				fmt.Printf("\n ACTUAL: %# v", pretty.Formatter(actual))
				fmt.Printf("\n EXPECTED: %# v", pretty.Formatter(expected))
			}
			So(actual, ShouldResembleDiffed, expected)
		})
	}

	Convey("Subject: OptimizePlan32", t, func() {
		test("select a from foo where a = 10 AND b < c",
			[]bson.D{
				{{"$match", bson.M{
					"a": int64(10),
				}}},
				{{"$project", bson.M{
					"d.e": 1,
					"_id": 1,
					"__predicate": bson.D{
						{"$let", bson.D{
							{"vars", bson.M{
								"predicate": bson.M{
									"$cond": []interface{}{
										bson.M{
											"$or": []interface{}{
												bson.M{
													"$eq": []interface{}{
														bson.M{
															"$ifNull": []interface{}{
																"$b",
																nil,
															},
														},
														nil,
													},
												},
												bson.M{
													"$eq": []interface{}{
														bson.M{
															"$ifNull": []interface{}{
																"$c",
																nil,
															},
														},
														nil,
													},
												},
											},
										},
										nil,
										bson.M{
											"$lt": []interface{}{
												"$b",
												"$c",
											},
										},
									},
								},
							}},
							{"in", bson.D{
								{"$cond", []interface{}{
									bson.D{{"$or", []interface{}{
										bson.D{{"$eq", []interface{}{"$$predicate", false}}},
										bson.D{{"$eq", []interface{}{"$$predicate", 0}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "-0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "0.0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "-0.0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", nil}}},
									}}},
									false,
									true,
								}},
							}},
						}},
					},
					"a":      1,
					"b":      1,
					"c":      1,
					"d.f":    1,
					"g":      1,
					"filter": 1,
				}}},
				{{"$match", bson.M{
					"__predicate": true,
				}}},
				{{"$project", bson.M{
					"foo_DOT_a": "$a",
				}}},
			},
		)
		test("select a from foo where b < c AND a = 10",
			[]bson.D{
				{{"$match", bson.M{
					"a": int64(10),
				}}},
				{{"$project", bson.M{
					"d.e": 1,
					"_id": 1,
					"__predicate": bson.D{
						{"$let", bson.D{
							{"vars", bson.M{
								"predicate": bson.M{
									"$cond": []interface{}{
										bson.M{
											"$or": []interface{}{
												bson.M{
													"$eq": []interface{}{
														bson.M{
															"$ifNull": []interface{}{
																"$b",
																nil,
															},
														},
														nil,
													},
												},
												bson.M{
													"$eq": []interface{}{
														bson.M{
															"$ifNull": []interface{}{
																"$c",
																nil,
															},
														},
														nil,
													},
												},
											},
										},
										nil,
										bson.M{
											"$lt": []interface{}{
												"$b",
												"$c",
											},
										},
									},
								},
							}},
							{"in", bson.D{
								{"$cond", []interface{}{
									bson.D{{"$or", []interface{}{
										bson.D{{"$eq", []interface{}{"$$predicate", false}}},
										bson.D{{"$eq", []interface{}{"$$predicate", 0}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "-0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "0.0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "-0.0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", nil}}},
									}}},
									false,
									true,
								}},
							}},
						}},
					},
					"a":      1,
					"b":      1,
					"c":      1,
					"d.f":    1,
					"g":      1,
					"filter": 1,
				}}},
				{{"$match", bson.M{
					"__predicate": true,
				}}},
				{{"$project", bson.M{
					"foo_DOT_a": "$a",
				}}},
			},
		)
		test("select a from foo where b < c",
			[]bson.D{
				{{"$project", bson.M{
					"d.e": 1,
					"_id": 1,
					"__predicate": bson.D{
						{"$let", bson.D{
							{"vars", bson.M{
								"predicate": bson.M{
									"$cond": []interface{}{
										bson.M{
											"$or": []interface{}{
												bson.M{
													"$eq": []interface{}{
														bson.M{
															"$ifNull": []interface{}{
																"$b",
																nil,
															},
														},
														nil,
													},
												},
												bson.M{
													"$eq": []interface{}{
														bson.M{
															"$ifNull": []interface{}{
																"$c",
																nil,
															},
														},
														nil,
													},
												},
											},
										},
										nil,
										bson.M{
											"$lt": []interface{}{
												"$b",
												"$c",
											},
										},
									},
								},
							}},
							{"in", bson.D{
								{"$cond", []interface{}{
									bson.D{{"$or", []interface{}{
										bson.D{{"$eq", []interface{}{"$$predicate", false}}},
										bson.D{{"$eq", []interface{}{"$$predicate", 0}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "-0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "0.0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "-0.0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", nil}}},
									}}},
									false,
									true,
								}},
							}},
						}},
					},
					"a":      1,
					"b":      1,
					"c":      1,
					"d.f":    1,
					"g":      1,
					"filter": 1,
				}}},
				{{"$match", bson.M{
					"__predicate": true,
				}}},
				{{"$project", bson.M{
					"foo_DOT_a": "$a",
				}}},
			},
		)
		test("select * from bar a join foo b on a.a=b.a and a.a=b.f",
			[]bson.D{
				{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
				{{"$lookup", bson.M{
					"from":         "foo",
					"localField":   "a",
					"foreignField": "a",
					"as":           "__joined_b",
				}}},
				{{"$unwind", bson.M{
					"path": "$__joined_b",
					"preserveNullAndEmptyArrays": false,
				}}},
				{{"$project", bson.M{
					"__joined_b._id":    1,
					"__joined_b.a":      1,
					"__joined_b.b":      1,
					"__joined_b.c":      1,
					"__joined_b.d.e":    1,
					"__joined_b.d.f":    1,
					"__joined_b.filter": 1,
					"__joined_b.g":      1,
					"_id":               1,
					"a":                 1,
					"b":                 1,
					"__predicate": bson.D{
						{"$let", bson.D{
							{"vars", bson.M{
								"predicate": bson.M{
									"$cond": []interface{}{
										bson.M{
											"$or": []interface{}{
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
												bson.M{
													"$eq": []interface{}{
														bson.M{
															"$ifNull": []interface{}{
																"$__joined_b.d.f",
																nil,
															},
														},
														nil,
													},
												},
											},
										},
										nil,
										bson.M{
											"$eq": []interface{}{
												"$a",
												"$__joined_b.d.f",
											},
										},
									},
								},
							}},
							{"in", bson.D{
								{"$cond", []interface{}{
									bson.D{{"$or", []interface{}{
										bson.D{{"$eq", []interface{}{"$$predicate", false}}},
										bson.D{{"$eq", []interface{}{"$$predicate", 0}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "-0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "0.0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "-0.0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", nil}}},
									}}},
									false,
									true,
								}},
							}},
						}},
					},
				}}},
				{{"$match", bson.M{
					"__predicate": true,
				}}},
			},
		)

	})
}

func TestOptimizePlan(t *testing.T) {
	testSchema, err := schema.New(testSchema4)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}

	testInfo := getMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)
	testVariables := createTestVariables(testInfo)
	testCatalog := getCatalogFromSchema(testSchema, testVariables)
	defaultDbName := "test"

	test := func(sql string, expected ...[]bson.D) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			plan, err := AlgebrizeQuery(statement, defaultDbName, testVariables, testCatalog)
			So(err, ShouldBeNil)

			actualPlan := OptimizePlan(createTestConnectionCtx(testInfo), plan)

			pg := &pipelineGatherer{}
			pg.visit(actualPlan)

			actual := pg.pipelines

			v := ShouldResembleDiffed(actual, expected)
			if v != "" {
				fmt.Printf("\n ACTUAL: %# v", pretty.Formatter(actual))
				fmt.Printf("\n EXPECTED: %# v", pretty.Formatter(expected))
			}
			So(actual, ShouldResembleDiffed, expected)
		})
	}

	testNoPushdown := func(sql string) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			plan, err := AlgebrizeQuery(statement, defaultDbName, testVariables, testCatalog)
			So(err, ShouldBeNil)

			actualPlan := OptimizePlan(createTestConnectionCtx(testInfo), plan)

			pg := &pipelineGatherer{}
			pg.visit(actualPlan)

			actual := pg.pipelines

			So(len(actual), ShouldEqual, 0)
		})
	}

	Convey("Subject: OptimizePlan", t, func() {
		Convey("from", func() {
			Convey("subqueries", func() {
				test("select a, b from (select a, b from bar) b",
					[]bson.D{
						{{"$project", bson.M{
							"bar_DOT_a": "$a",
							"bar_DOT_b": "$b",
						}}},
						{{"$project", bson.M{
							"b_DOT_a": "$bar_DOT_a",
							"b_DOT_b": "$bar_DOT_b",
						}}},
					},
				)
			})

			Convey("joins", func() {
				Convey("inner join", func() {
					test("select * from bar a join foo b on a.a=b.a and a.a=b.f",
						[]bson.D{
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_b",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_b",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$addFields", bson.M{
								"__predicate": bson.D{
									{"$let", bson.D{
										{"vars", bson.M{
											"predicate": bson.M{
												"$cond": []interface{}{
													bson.M{
														"$or": []interface{}{
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
															bson.M{
																"$eq": []interface{}{
																	bson.M{
																		"$ifNull": []interface{}{
																			"$__joined_b.d.f",
																			nil,
																		},
																	},
																	nil,
																},
															},
														},
													},
													nil,
													bson.M{
														"$eq": []interface{}{
															"$a",
															"$__joined_b.d.f",
														},
													},
												},
											}},
										},
										{"in", bson.D{
											{"$cond", []interface{}{
												bson.D{{"$or", []interface{}{
													bson.D{{"$eq", []interface{}{"$$predicate", false}}},
													bson.D{{"$eq", []interface{}{"$$predicate", 0}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "-0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "0.0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "-0.0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", nil}}},
												}}},
												false,
												true,
											}},
										}},
									}},
								},
							}}},
							{{"$match", bson.M{
								"__predicate": true,
							}}},
						},
					)
					test("select foo.a, bar.b from foo inner join bar on foo.a = bar.a",
						[]bson.D{
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"bar_DOT_b": "$__joined_bar.b",
							}}},
						},
					)

					test("select foo.a, bar.b from foo inner join bar on foo.a = bar.a where foo.b = 10",
						[]bson.D{
							{{"$match", bson.M{"b": int64(10)}}},
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"bar_DOT_b": "$__joined_bar.b",
							}}},
						},
					)

					test("select foo.a, bar.b from foo inner join bar on foo.a = bar.a where foo.b = 10 AND bar.b = 12",
						[]bson.D{
							{{"$match", bson.M{"b": int64(10)}}},
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$match", bson.M{"__joined_bar.b": int64(12)}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"bar_DOT_b": "$__joined_bar.b",
							}}},
						},
					)

					test("select foo.a, bar.b from foo inner join bar on foo.a = bar.a where foo.b = 10 OR bar.b = 12",
						[]bson.D{
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$match", bson.M{
								"$or": []interface{}{
									bson.M{"b": int64(10)},
									bson.M{"__joined_bar.b": int64(12)},
								},
							}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"bar_DOT_b": "$__joined_bar.b",
							}}},
						},
					)

					test("select foo.a, bar.b from foo inner join bar on foo.a = bar.a where foo.b = 11 AND (foo.b = 10 OR bar.b = 12)",
						[]bson.D{
							{{"$match", bson.M{"b": int64(11)}}},
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$match", bson.M{
								"$or": []interface{}{
									bson.M{"b": int64(10)},
									bson.M{"__joined_bar.b": int64(12)},
								},
							}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"bar_DOT_b": "$__joined_bar.b",
							}}},
						},
					)

					test("select foo.a, bar.b from foo inner join bar on foo.a = bar.a where (foo.b = 11 OR foo.b = 10) AND bar.b = 12",
						[]bson.D{
							{{"$match", bson.M{
								"$or": []interface{}{
									bson.M{"b": int64(11)},
									bson.M{"b": int64(10)},
								},
							}}},
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$match", bson.M{"__joined_bar.b": int64(12)}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"bar_DOT_b": "$__joined_bar.b",
							}}},
						},
					)

					test("select foo.a, bar.b from foo inner join (select bar.a, bar.b from bar where bar.b = 12) bar on foo.a = bar.a where bar.a = 10",
						[]bson.D{
							{{"$match", bson.M{"b": int64(12)}}},
							{{"$project", bson.M{
								"bar_DOT_a": "$a",
								"bar_DOT_b": "$b",
							}}},
							{{"$match", bson.M{"bar_DOT_a": int64(10)}}},
							{{"$match", bson.M{"bar_DOT_a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "bar_DOT_a",
								"foreignField": "a",
								"as":           "__joined_foo",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_foo",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$__joined_foo.a",
								"bar_DOT_b": "$bar_DOT_b",
							}}},
						},
					)

					test("select foo.a, bar.a from foo inner join bar on foo.a = bar.a",
						[]bson.D{
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": false,
							}}},
						},
					)

					test("select foo.a, bar.b from foo inner join bar on foo.a = bar.a AND foo.b > 10",
						[]bson.D{
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$match", bson.M{
								"b": bson.M{"$gt": int64(10)},
							}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"bar_DOT_b": "$__joined_bar.b",
							}}},
						},
					)

					test("select foo.a, bar.b from foo inner join bar on foo.a = bar.a AND foo.b > 10 AND (bar.b < 12 OR bar.b > 10)",
						[]bson.D{
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$match", bson.M{
								"$and": []interface{}{
									bson.M{"b": bson.M{"$gt": int64(10)}},
									bson.M{"$or": []interface{}{
										bson.M{"__joined_bar.b": bson.M{"$lt": int64(12)}},
										bson.M{"__joined_bar.b": bson.M{"$gt": int64(10)}},
									}},
								},
							}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"bar_DOT_b": "$__joined_bar.b",
							}}},
						},
					)

					test("select foo.a, bar.b from foo, bar where foo.a = bar.a",
						[]bson.D{
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"bar_DOT_b": "$__joined_bar.b",
							}}},
						},
					)

					test("select foo.a, b.b from foo join (select a, b from bar) b on foo.a=b.a",
						[]bson.D{
							{{"$project", bson.M{
								"bar_DOT_a": "$a",
								"bar_DOT_b": "$b",
							}}},
							{{"$match", bson.M{"bar_DOT_a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "bar_DOT_a",
								"foreignField": "a",
								"as":           "__joined_foo",
							}}},
							{{"$unwind", bson.M{
								"preserveNullAndEmptyArrays": false,
								"path": "$__joined_foo",
							}}},
							{{"$project", bson.M{
								"b_DOT_b":   "$bar_DOT_b",
								"foo_DOT_a": "$__joined_foo.a",
							}}},
						},
					)

					test("select * from (select foo.a from bar join (select foo.a from foo) foo on foo.a=bar.b) x join (select g.a from bar join (select foo.a from foo) g on g.a=bar.a) y on x.a=y.a",
						[]bson.D{
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
							}}},
							{{"$match", bson.M{"foo_DOT_a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "foo_DOT_a",
								"foreignField": "b",
								"as":           "__joined_bar",
							}}},
							{{"$unwind", bson.M{
								"preserveNullAndEmptyArrays": false,
								"path": "$__joined_bar",
							}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$foo_DOT_a",
							}}},
							{{"$project", bson.M{
								"x_DOT_a": "$foo_DOT_a",
							}}},
						},
						[]bson.D{
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
							}}},
							{{"$match", bson.M{"foo_DOT_a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "foo_DOT_a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$unwind", bson.M{
								"preserveNullAndEmptyArrays": false,
								"path": "$__joined_bar",
							}}},
							{{"$project", bson.M{
								"g_DOT_a": "$foo_DOT_a",
							}}},
							{{"$project", bson.M{
								"y_DOT_a": "$g_DOT_a",
							}}},
						},
					)

					test("select * from foo f left join (select b.b from foo f join (select * from bar) b on f.a=b.a)  b on f.a=b.b",
						[]bson.D{
							{{"$project", bson.M{
								"f_DOT__id": "$_id",
								"f_DOT_a":   "$a",
								"f_DOT_b":   "$b",
								"f_DOT_c":   "$c",
								"f_DOT_e":   "$d.e",
								"f_DOT_f":   "$d.f",
								"f_DOT_g":   "$g",
							}}},
						},
						[]bson.D{
							{{"$project", bson.M{
								"bar_DOT__id": "$_id",
								"bar_DOT_a":   "$a",
								"bar_DOT_b":   "$b",
							}}},
							{{"$match", bson.M{"bar_DOT_a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "bar_DOT_a",
								"foreignField": "a",
								"as":           "__joined_f",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_f",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$project", bson.M{
								"b_DOT_b": "$bar_DOT_b",
							}}},
							{{"$project", bson.M{
								"b_DOT_b": "$b_DOT_b",
							}}},
						},
					)

					test("select * from foo f right join (select b.b from foo f join (select * from bar) b on f.a=b.a)  b on f.a=b.b",
						[]bson.D{
							{{"$project", bson.M{
								"bar_DOT__id": "$_id",
								"bar_DOT_a":   "$a",
								"bar_DOT_b":   "$b",
							}}},
							{{"$match", bson.M{"bar_DOT_a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "bar_DOT_a",
								"foreignField": "a",
								"as":           "__joined_f",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_f",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$project", bson.M{
								"b_DOT_b": "$bar_DOT_b",
							}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "b_DOT_b",
								"foreignField": "a",
								"as":           "__joined_f",
							}}},
							{{"$project", bson.M{
								"b_DOT_b": 1,
								"__joined_f": bson.M{
									"$cond": []interface{}{
										bson.M{"$eq": []interface{}{
											bson.M{"$ifNull": []interface{}{"$b_DOT_b", nil}},
											nil,
										}},
										bson.M{"$literal": []interface{}{}},
										"$__joined_f",
									},
								},
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_f",
								"preserveNullAndEmptyArrays": true,
							}}},
						},
					)

					test("select * from foo f join merge m1 on f._id=m1._id join (select * from foo) g on g.a=f.a join merge_d_a m2 on m2._id=m1._id and m2._id=g.a",
						[]bson.D{
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d_idx"},
								{"path", "$d"},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d.a_idx"},
								{"path", "$d.a"},
							}}},
							{{"$match", bson.M{"_id": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "_id",
								"foreignField": "_id",
								"as":           "__joined_f",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_f",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$project", bson.M{
								"f_DOT__id":          "$__joined_f._id",
								"f_DOT_a":            "$__joined_f.a",
								"f_DOT_b":            "$__joined_f.b",
								"f_DOT_c":            "$__joined_f.c",
								"f_DOT_e":            "$__joined_f.d.e",
								"f_DOT_f":            "$__joined_f.d.f",
								"f_DOT_g":            "$__joined_f.g",
								"m1_DOT__id":         "$_id",
								"m1_DOT_a":           "$a",
								"m2_DOT__id":         "$_id",
								"m2_DOT_d_DOT_a":     "$d.a",
								"m2_DOT_d_DOT_a_idx": "$d.a_idx",
								"m2_DOT_d_idx":       "$d_idx",
							}}},
						},
						[]bson.D{
							{{"$project", bson.M{
								"foo_DOT__id": "$_id",
								"foo_DOT_a":   "$a",
								"foo_DOT_b":   "$b",
								"foo_DOT_c":   "$c",
								"foo_DOT_e":   "$d.e",
								"foo_DOT_f":   "$d.f",
								"foo_DOT_g":   "$g",
							}}},
							{{"$project", bson.M{
								"g_DOT__id": "$foo_DOT__id",
								"g_DOT_a":   "$foo_DOT_a",
								"g_DOT_b":   "$foo_DOT_b",
								"g_DOT_c":   "$foo_DOT_c",
								"g_DOT_e":   "$foo_DOT_e",
								"g_DOT_f":   "$foo_DOT_f",
								"g_DOT_g":   "$foo_DOT_g",
							}}},
						},
					)
					test("select f.a from foo f join (select bar.a from bar) b on f.a=b.a join (select foo.a from foo where foo.a > 4 limit 1) c on b.a=c.a and f.a=c.a and f.b=b.a",
						[]bson.D{
							{{"$match", bson.M{"a": bson.M{"$gt": int64(4)}}}},
							{{"$limit", int64(1)}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
							}}},
							{{"$project", bson.M{
								"c_DOT_a": "$foo_DOT_a",
							}}},
						},
						[]bson.D{
							{{"$project", bson.M{
								"bar_DOT_a": "$a",
							}}},
							{{"$project", bson.M{
								"b_DOT_a": "$bar_DOT_a",
							}}},
						},
						[]bson.D{
							{{"$project", bson.M{
								"f_DOT_a": "$a",
								"f_DOT_b": "$b",
							}}},
						},
					)

					test("select * from foo f join merge m1 on f._id=m1._id join merge_d_a m2 on m1._id=m2._id and f._id=m2._id",
						[]bson.D{
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d_idx"},
								{"path", "$d"},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d.a_idx"},
								{"path", "$d.a"},
							}}},
							{{"$match", bson.M{"_id": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "_id",
								"foreignField": "_id",
								"as":           "__joined_f",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_f",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$addFields", bson.M{
								"__predicate": bson.D{
									{"$let", bson.D{
										{"vars", bson.M{
											"predicate": bson.M{
												"$cond": []interface{}{
													bson.M{
														"$or": []interface{}{
															bson.M{
																"$eq": []interface{}{
																	bson.M{
																		"$ifNull": []interface{}{
																			"$__joined_f._id",
																			nil,
																		},
																	},
																	nil,
																},
															},
															bson.M{
																"$eq": []interface{}{
																	bson.M{
																		"$ifNull": []interface{}{
																			"$_id",
																			nil,
																		},
																	},
																	nil,
																},
															},
														},
													},
													nil,
													bson.M{
														"$eq": []interface{}{
															"$__joined_f._id",
															"$_id",
														},
													},
												},
											},
										}},
										{"in", bson.D{
											{"$cond", []interface{}{
												bson.D{{"$or", []interface{}{
													bson.D{{"$eq", []interface{}{"$$predicate", false}}},
													bson.D{{"$eq", []interface{}{"$$predicate", 0}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "-0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "0.0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "-0.0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", nil}}},
												}}},
												false,
												true,
											}},
										}},
									}},
								},
							}}},
							{{"$match", bson.M{
								"__predicate": true,
							}}},
						},
					)

					test("select foo.c, bar.a, baz.b from foo inner join bar on foo.a = bar.a inner join baz on bar.a = baz.a",
						[]bson.D{
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "baz",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_baz",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_baz",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_foo",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_foo",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$project", bson.M{
								"foo_DOT_c": "$__joined_foo.c",
								"bar_DOT_a": "$a",
								"baz_DOT_b": "$__joined_baz.b",
							}}},
						},
					)

					test("select foo.a, bar.a, baz.a from foo inner join bar on foo.a = bar.a inner join baz on bar.a = baz.a",
						[]bson.D{
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "baz",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_baz",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_baz",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_foo",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_foo",
								"preserveNullAndEmptyArrays": false,
							}}},
						},
					)

					test("select foo.a, bar.a, baz.a from bar inner join baz on baz.a = bar.a inner join foo on baz.a = foo.a and baz.a > foo.c",
						[]bson.D{
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_foo",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_foo",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$addFields", bson.M{
								"__predicate": bson.D{
									{"$let", bson.D{
										{"vars", bson.M{
											"predicate": bson.M{
												"$cond": []interface{}{
													bson.M{
														"$or": []interface{}{
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
															bson.M{
																"$eq": []interface{}{
																	bson.M{
																		"$ifNull": []interface{}{
																			"$__joined_foo.c",
																			nil,
																		},
																	},
																	nil,
																},
															},
														},
													},
													nil,
													bson.M{
														"$gt": []interface{}{
															"$a",
															"$__joined_foo.c",
														},
													},
												},
											},
										}},
										{"in", bson.D{
											{"$cond", []interface{}{
												bson.D{{"$or", []interface{}{
													bson.D{{"$eq", []interface{}{"$$predicate", false}}},
													bson.D{{"$eq", []interface{}{"$$predicate", 0}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "-0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "0.0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "-0.0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", nil}}},
												}}},
												false,
												true,
											}},
										}},
									}},
								},
							}}},
							{{"$match", bson.M{
								"__predicate": true,
							}}},
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": false,
							}}},
						},
					)

					test("select * from foo join (bar join baz on bar.a = baz.a) on foo.a = bar.a",
						[]bson.D{
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$match", bson.M{"__joined_bar.a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "baz",
								"localField":   "__joined_bar.a",
								"foreignField": "a",
								"as":           "__joined_baz",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_baz",
								"preserveNullAndEmptyArrays": false,
							}}},
						},
					)

					test("select * from foo f join merge m1 on f._id=m1._id join merge_d_a m2 on m2._id=f._id and m2._id=m1._id",
						[]bson.D{
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d_idx"},
								{"path", "$d"},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d.a_idx"},
								{"path", "$d.a"},
							}}},
							{{"$match", bson.M{"_id": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "_id",
								"foreignField": "_id",
								"as":           "__joined_f",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_f",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$addFields", bson.M{
								"__predicate": bson.D{
									{"$let", bson.D{
										{"vars", bson.M{
											"predicate": bson.M{
												"$cond": []interface{}{
													bson.M{
														"$or": []interface{}{
															bson.M{
																"$eq": []interface{}{
																	bson.M{
																		"$ifNull": []interface{}{
																			"$__joined_f._id",
																			nil,
																		},
																	},
																	nil,
																},
															},
															bson.M{
																"$eq": []interface{}{
																	bson.M{
																		"$ifNull": []interface{}{
																			"$_id",
																			nil,
																		},
																	},
																	nil,
																},
															},
														},
													},
													nil,
													bson.M{
														"$eq": []interface{}{
															"$__joined_f._id",
															"$_id",
														},
													},
												},
											},
										}},
										{"in", bson.D{
											{"$cond", []interface{}{
												bson.D{{"$or", []interface{}{
													bson.D{{"$eq", []interface{}{"$$predicate", false}}},
													bson.D{{"$eq", []interface{}{"$$predicate", 0}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "-0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "0.0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", "-0.0"}}},
													bson.D{{"$eq", []interface{}{"$$predicate", nil}}},
												}}},
												false,
												true,
											}},
										}},
									}},
								},
							}}},
							{{"$match", bson.M{"__predicate": true}}},
						},
					)

					test("select f1.a, b1.b from foo f1 inner join (select b2.b, b2.a, b2._id from bar b2 join (select * from foo) f2 on f2._id = b2._id) b1 on b1._id = f1._id",
						[]bson.D{
							{{"$project", bson.M{
								"foo_DOT__id": "$_id",
								"foo_DOT_a":   "$a",
								"foo_DOT_b":   "$b",
								"foo_DOT_c":   "$c",
								"foo_DOT_e":   "$d.e",
								"foo_DOT_f":   "$d.f",
								"foo_DOT_g":   "$g",
							}}},
							{{"$match", bson.M{"foo_DOT__id": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "foo_DOT__id",
								"foreignField": "_id",
								"as":           "__joined_b2",
							}}},
							{{"$unwind", bson.M{
								"preserveNullAndEmptyArrays": false,
								"path": "$__joined_b2",
							}}},
							{{"$project", bson.M{
								"b2_DOT__id": "$__joined_b2._id",
								"b2_DOT_a":   "$__joined_b2.a",
								"b2_DOT_b":   "$__joined_b2.b",
							}}},
							{{"$match", bson.M{"b2_DOT__id": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "b2_DOT__id",
								"foreignField": "_id",
								"as":           "__joined_f1",
							}}},
							{{"$unwind", bson.M{
								"preserveNullAndEmptyArrays": false,
								"path": "$__joined_f1",
							}}},
							{{"$project", bson.M{
								"b1_DOT_b": "$b2_DOT_b",
								"f1_DOT_a": "$__joined_f1.a",
							}}},
						},
					)

					test("select foo.a, bar.a, baz.a from foo inner join bar on foo.a = bar.a inner join baz on bar.a = baz.a where foo.a = 10 AND bar.a = 12 AND baz.a = 13",
						[]bson.D{
							{{"$match", bson.M{"a": int64(12)}}},
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "baz",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_baz",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_baz",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$match", bson.M{"__joined_baz.a": int64(13)}}},
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_foo",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_foo",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$match", bson.M{"__joined_foo.a": int64(10)}}},
						},
					)

					test("select foo.a, bar.a, baz.a from foo inner join bar on foo.a = bar.a inner join baz on bar.a = baz.a where (foo.a = 10 OR bar.a = 11) AND bar.a = 12 AND baz.a = 13",
						[]bson.D{
							{{"$match", bson.M{"a": int64(12)}}},
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "baz",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_baz",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_baz",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$match", bson.M{"__joined_baz.a": int64(13)}}},
							{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_foo",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_foo",
								"preserveNullAndEmptyArrays": false,
							}}},
							{{"$match", bson.M{
								"$or": []interface{}{
									bson.M{"__joined_foo.a": int64(10)},
									bson.M{"a": int64(11)},
								},
							}}},
						},
					)
				})

				Convey("flip join", func() {
					test("select * from foo r inner join merge_d_a a on r._id=a._id",
						[]bson.D{
							{{"$unwind", bson.D{{"includeArrayIndex", "d_idx"}, {"path", "$d"}}}},
							{{"$unwind", bson.D{{"includeArrayIndex", "d.a_idx"}, {"path", "$d.a"}}}},
							{{"$match", bson.M{"_id": bson.M{"$ne": nil}}}},
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "_id",
								"foreignField": "_id",
								"as":           "__joined_r",
							}}},
							{{"$unwind", bson.M{"path": "$__joined_r", "preserveNullAndEmptyArrays": false}}},
						},
					)
				})

				Convey("left join", func() {
					test("select foo.a, bar.b from foo left outer join bar on foo.a = bar.a",
						[]bson.D{
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$project", bson.M{
								"_id":    1,
								"a":      1,
								"b":      1,
								"c":      1,
								"d.e":    1,
								"d.f":    1,
								"filter": 1,
								"g":      1,
								"__joined_bar": bson.M{
									"$cond": []interface{}{
										bson.M{"$eq": []interface{}{
											bson.M{"$ifNull": []interface{}{"$a", nil}},
											nil,
										}},
										bson.M{"$literal": []interface{}{}},
										"$__joined_bar",
									},
								},
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": true,
							}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"bar_DOT_b": "$__joined_bar.b",
							}}},
						},
					)

					test("select foo.a, bar.b from foo left outer join bar on foo.a = bar.a where foo.a = 10 AND bar.b = 12",
						[]bson.D{
							{{"$match", bson.M{"a": int64(10)}}},
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$project", bson.M{
								"_id":    1,
								"a":      1,
								"b":      1,
								"c":      1,
								"d.e":    1,
								"d.f":    1,
								"filter": 1,
								"g":      1,
								"__joined_bar": bson.M{
									"$cond": []interface{}{
										bson.M{"$eq": []interface{}{
											bson.M{"$ifNull": []interface{}{"$a", nil}},
											nil,
										}},
										bson.M{"$literal": []interface{}{}},
										"$__joined_bar",
									},
								},
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": true,
							}}},
							{{"$match", bson.M{"__joined_bar.b": int64(12)}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"bar_DOT_b": "$__joined_bar.b",
							}}},
						},
					)

					test("select foo.a, bar.b from foo left join bar on foo.a = bar.a AND foo.b > 10",
						[]bson.D{
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$project", bson.M{
								"_id":    1,
								"a":      1,
								"b":      1,
								"c":      1,
								"d.e":    1,
								"d.f":    1,
								"filter": 1,
								"g":      1,
								"__joined_bar": bson.M{
									"$cond": []interface{}{
										bson.M{"$eq": []interface{}{
											bson.M{"$ifNull": []interface{}{"$a", nil}},
											nil,
										}},
										bson.M{"$literal": []interface{}{}},
										"$__joined_bar",
									},
								},
							}}},
							{{"$project", bson.M{
								"__joined_bar": bson.M{
									"$filter": bson.M{
										"input": "$__joined_bar",
										"as":    "this",
										"cond": bson.M{
											"$cond": []interface{}{
												bson.M{
													"$eq": []interface{}{
														bson.M{"$ifNull": []interface{}{"$b", nil}},
														nil,
													},
												},
												nil,
												bson.M{
													"$gt": []interface{}{
														"$b",
														bson.M{
															"$literal": int64(10),
														},
													},
												},
											},
										},
									},
								},
								"a":      1,
								"b":      1,
								"c":      1,
								"d.e":    1,
								"d.f":    1,
								"_id":    1,
								"filter": 1,
								"g":      1,
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": true,
							}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"bar_DOT_b": "$__joined_bar.b",
							}}},
						},
					)

					test("select foo.a, bar.b from foo left join bar on foo.a = bar.a AND bar.b > 10",
						[]bson.D{
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$project", bson.M{
								"_id":    1,
								"a":      1,
								"b":      1,
								"c":      1,
								"d.e":    1,
								"d.f":    1,
								"filter": 1,
								"g":      1,
								"__joined_bar": bson.M{
									"$cond": []interface{}{
										bson.M{"$eq": []interface{}{
											bson.M{"$ifNull": []interface{}{"$a", nil}},
											nil,
										}},
										bson.M{"$literal": []interface{}{}},
										"$__joined_bar",
									},
								},
							}}},
							{{"$project", bson.M{
								"__joined_bar": bson.M{
									"$filter": bson.M{
										"as": "this",
										"cond": bson.M{
											"$cond": []interface{}{
												bson.M{
													"$eq": []interface{}{
														bson.M{"$ifNull": []interface{}{"$$this.b", nil}},
														nil,
													},
												},
												nil,
												bson.M{
													"$gt": []interface{}{
														"$$this.b",
														bson.M{
															"$literal": int64(10),
														},
													},
												},
											},
										},
										"input": "$__joined_bar",
									},
								},
								"b":      int(1),
								"d.e":    int(1),
								"d.f":    int(1),
								"filter": int(1),
								"a":      int(1),
								"c":      int(1),
								"g":      int(1),
								"_id":    int(1),
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": true,
							}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"bar_DOT_b": "$__joined_bar.b",
							}}},
						},
					)

					test("select foo.c, bar.a, baz.b from foo left join bar on foo.a = bar.a left join baz on bar.a = baz.a",
						[]bson.D{
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$project", bson.M{
								"_id":    1,
								"a":      1,
								"b":      1,
								"c":      1,
								"d.e":    1,
								"d.f":    1,
								"filter": 1,
								"g":      1,
								"__joined_bar": bson.M{
									"$cond": []interface{}{
										bson.M{"$eq": []interface{}{
											bson.M{"$ifNull": []interface{}{"$a", nil}},
											nil,
										}},
										bson.M{"$literal": []interface{}{}},
										"$__joined_bar",
									},
								},
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": true,
							}}},
							{{"$lookup", bson.M{
								"from":         "baz",
								"localField":   "__joined_bar.a",
								"foreignField": "a",
								"as":           "__joined_baz",
							}}},
							{{"$project", bson.M{
								"_id":              1,
								"a":                1,
								"b":                1,
								"c":                1,
								"d.e":              1,
								"d.f":              1,
								"filter":           1,
								"g":                1,
								"__joined_bar._id": 1,
								"__joined_bar.a":   1,
								"__joined_bar.b":   1,
								"__joined_baz": bson.M{
									"$cond": []interface{}{
										bson.M{"$eq": []interface{}{
											bson.M{"$ifNull": []interface{}{"$__joined_bar.a", nil}},
											nil,
										}},
										bson.M{"$literal": []interface{}{}},
										"$__joined_baz",
									},
								},
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_baz",
								"preserveNullAndEmptyArrays": true,
							}}},
							{{"$project", bson.M{
								"foo_DOT_c": "$c",
								"bar_DOT_a": "$__joined_bar.a",
								"baz_DOT_b": "$__joined_baz.b",
							}}},
						},
					)
				})

				Convey("right join", func() {
					test("select foo.a from foo right join bar on foo.a < bar.a",
						[]bson.D{
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
							}}},
						},
						[]bson.D{
							{{"$project", bson.M{
								"bar_DOT_a": "$a",
							}}},
						},
					)

					test("select foo.a, bar.b from foo right outer join bar on foo.a = bar.a",
						[]bson.D{
							{{"$lookup", bson.M{
								"from":         "foo",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_foo",
							}}},
							{{"$project", bson.M{
								"_id": 1,
								"a":   1,
								"b":   1,
								"__joined_foo": bson.M{
									"$cond": []interface{}{
										bson.M{"$eq": []interface{}{
											bson.M{"$ifNull": []interface{}{"$a", nil}},
											nil,
										}},
										bson.M{"$literal": []interface{}{}},
										"$__joined_foo",
									},
								},
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_foo",
								"preserveNullAndEmptyArrays": true,
							}}},
							{{"$project", bson.M{
								"bar_DOT_b": "$b",
								"foo_DOT_a": "$__joined_foo.a",
							}}},
						},
					)
				})

				Convey("merge join", func() {
					test("select * from merge r left join merge_d_a a on r._id=a._id",
						[]bson.D{
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d_idx"},
								{"path", "$d"},
								{"preserveNullAndEmptyArrays", true},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d.a_idx"},
								{"path", "$d.a"},
								{"preserveNullAndEmptyArrays", true},
							}}},
						},
					)
					test("select b._id, c._id from merge r inner join merge_b b on r._id=b._id inner join merge_c c on b._id=c._id",
						[]bson.D{
							{{"$unwind", bson.D{{"includeArrayIndex", "b_idx"}, {"path", "$b"}}}},
							{{"$unwind", bson.D{{"includeArrayIndex", "c_idx"}, {"path", "$c"}}}},
						},
					)
					test("select b._id, c._id from merge r left join merge_b b on r._id=b._id left join merge_c c on b._id=c._id",
						[]bson.D{
							{{"$unwind", bson.D{
								{"includeArrayIndex", "b_idx"},
								{"path", "$b"},
								{"preserveNullAndEmptyArrays", true},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "c_idx"},
								{"path", "$c"},
								{"preserveNullAndEmptyArrays", true},
							}}},
						},
					)
					test("select b._id, c._id from merge r left join merge_b b on r._id=b._id left join merge_c c on r._id=c._id",
						[]bson.D{
							{{"$unwind", bson.D{
								{"includeArrayIndex", "b_idx"},
								{"path", "$b"},
								{"preserveNullAndEmptyArrays", true},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "c_idx"},
								{"path", "$c"},
								{"preserveNullAndEmptyArrays", true},
							}}},
						},
					)
					test("select b._id, c._id from merge r inner join merge_b b on r._id=b._id left join merge_c c on r._id=c._id",
						[]bson.D{
							{{"$unwind", bson.D{{"includeArrayIndex", "b_idx"}, {"path", "$b"}}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "c_idx"},
								{"path", "$c"},
								{"preserveNullAndEmptyArrays", true},
							}}},
						},
					)
					test("select b._id, c._id from merge r left join merge_b b on r._id=b._id inner join merge_c c on r._id=c._id",
						[]bson.D{
							{{"$unwind", bson.D{
								{"includeArrayIndex", "b_idx"},
								{"path", "$b"},
								{"preserveNullAndEmptyArrays", true},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "c_idx"},
								{"path", "$c"},
							}}},
						},
					)
					test("select b._id, c._id from merge r left join merge_b b on r._id=b._id inner join merge_c c on r._id=c._id left join merge_d_a a on r._id=a._id",
						[]bson.D{
							{{"$unwind", bson.D{
								{"includeArrayIndex", "b_idx"},
								{"path", "$b"},
								{"preserveNullAndEmptyArrays", true},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "c_idx"},
								{"path", "$c"},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d_idx"},
								{"path", "$d"},
								{"preserveNullAndEmptyArrays", true},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d.a_idx"},
								{"path", "$d.a"},
								{"preserveNullAndEmptyArrays", true},
							}}},
						},
					)
					test("select b._id, c._id from merge r inner join merge_b b on r._id=b._id inner join merge_c c on r._id=c._id inner join merge_d_a a on r._id=a._id",
						[]bson.D{
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d_idx"},
								{"path", "$d"},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d.a_idx"},
								{"path", "$d.a"},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "c_idx"},
								{"path", "$c"},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "b_idx"},
								{"path", "$b"},
							}}},
						},
					)
					test("select b._id, r._id from merge r inner join merge_d d on r._id=d._id inner join merge_d_a a on r._id=a._id inner join merge_b b on r._id=b._id",
						[]bson.D{
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d_idx"},
								{"path", "$d"},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d.a_idx"},
								{"path", "$d.a"},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "b_idx"},
								{"path", "$b"},
							}}},
						},
					)
					test("select b._id, d._id from merge r inner join merge_b b on r._id=b._id inner join merge_d d on r._id=d._id inner join merge_d_a a on r._id=a._id",
						[]bson.D{
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d_idx"},
								{"path", "$d"},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "d.a_idx"},
								{"path", "$d.a"},
							}}},
							{{"$unwind", bson.D{
								{"includeArrayIndex", "b_idx"},
								{"path", "$b"},
							}}},
						},
					)
				})

				Convey("no push down, project columns", func() {
					test("select foo.a from foo inner join bar on foo.a < bar.a",
						[]bson.D{
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
							}}},
						},
						[]bson.D{
							{{"$project", bson.M{
								"bar_DOT_a": "$a",
							}}},
						},
					)
					test("select foo.a from foo inner join bar on foo.a < foo.b",
						[]bson.D{
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"foo_DOT_b": "$b",
							}}},
						},
					)
					test("select foo.a from foo, bar where foo.a < bar.a",
						[]bson.D{
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
							}}},
						},
						[]bson.D{
							{{"$project", bson.M{
								"bar_DOT_a": "$a",
							}}},
						},
					)
					test("select foo.a from foo left join bar on foo.a < bar.a",
						[]bson.D{
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
							}}},
						},
						[]bson.D{
							{{"$project", bson.M{
								"bar_DOT_a": "$a",
							}}},
						},
					)
					test("select foo.a from foo right join bar on foo.a < bar.a",
						[]bson.D{
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
							}}},
						},
						[]bson.D{
							{{"$project", bson.M{
								"bar_DOT_a": "$a",
							}}},
						},
					)
					test("select foo.a, b.b from foo, (select a, b from bar) b where foo.a = b.a",
						[]bson.D{
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
							}}},
						},
						[]bson.D{
							{{"$project", bson.M{
								"bar_DOT_a": "$a",
								"bar_DOT_b": "$b",
							}}},
							{{"$project", bson.M{
								"b_DOT_a": "$bar_DOT_a",
								"b_DOT_b": "$bar_DOT_b",
							}}},
						},
					)
				})
			})
		})

		Convey("select", func() {
			test("select a, b from foo",
				[]bson.D{
					{{"$project", bson.M{
						"foo_DOT_a": "$a",
						"foo_DOT_b": "$b",
					}}},
				},
			)

			Convey("correlated subqueries", func() {
				test("select a, (select foo.b from bar) from foo",
					[]bson.D{
						{{"$project", bson.M{
							"foo_DOT_a": "$a",
							"b":         "$b",
						}}},
					},
				)
			})
		})

		Convey("where", func() {
			test("select a from foo where a = 10",
				[]bson.D{
					{{"$match", bson.M{
						"a": int64(10),
					}}},
					{{"$project", bson.M{
						"foo_DOT_a": "$a",
					}}},
				},
			)

			test("select a from foo where a = 10 AND b < c",
				[]bson.D{
					{{"$match", bson.M{
						"a": int64(10),
					}}},
					{{"$addFields", bson.M{
						"__predicate": bson.D{
							{"$let", bson.D{
								{"vars", bson.M{
									"predicate": bson.M{
										"$cond": []interface{}{
											bson.M{
												"$or": []interface{}{
													bson.M{
														"$eq": []interface{}{
															bson.M{
																"$ifNull": []interface{}{
																	"$b",
																	nil,
																},
															},
															nil,
														},
													},
													bson.M{
														"$eq": []interface{}{
															bson.M{
																"$ifNull": []interface{}{
																	"$c",
																	nil,
																},
															},
															nil,
														},
													},
												},
											},
											nil,
											bson.M{
												"$lt": []interface{}{
													"$b",
													"$c",
												},
											},
										},
									},
								}},
								{"in", bson.D{
									{"$cond", []interface{}{
										bson.D{{"$or", []interface{}{
											bson.D{{"$eq", []interface{}{"$$predicate", false}}},
											bson.D{{"$eq", []interface{}{"$$predicate", 0}}},
											bson.D{{"$eq", []interface{}{"$$predicate", "0"}}},
											bson.D{{"$eq", []interface{}{"$$predicate", "-0"}}},
											bson.D{{"$eq", []interface{}{"$$predicate", "0.0"}}},
											bson.D{{"$eq", []interface{}{"$$predicate", "-0.0"}}},
											bson.D{{"$eq", []interface{}{"$$predicate", nil}}},
										}}},
										false,
										true,
									}},
								}},
							}},
						},
					}}},
					{{"$match", bson.M{
						"__predicate": true,
					}}},
					{{"$project", bson.M{
						"foo_DOT_a": "$a",
					}}},
				},
			)
			test("select a from foo where b < c AND a = 10",
				[]bson.D{
					{{"$match", bson.M{
						"a": int64(10),
					}}},
					{{"$addFields", bson.M{
						"__predicate": bson.D{
							{"$let", bson.D{
								{"vars", bson.M{
									"predicate": bson.M{
										"$cond": []interface{}{
											bson.M{
												"$or": []interface{}{
													bson.M{
														"$eq": []interface{}{
															bson.M{
																"$ifNull": []interface{}{
																	"$b",
																	nil,
																},
															},
															nil,
														},
													},
													bson.M{
														"$eq": []interface{}{
															bson.M{
																"$ifNull": []interface{}{
																	"$c",
																	nil,
																},
															},
															nil,
														},
													},
												},
											},
											nil,
											bson.M{
												"$lt": []interface{}{
													"$b",
													"$c",
												},
											},
										},
									},
								}},
								{"in", bson.D{
									{"$cond", []interface{}{
										bson.D{{"$or", []interface{}{
											bson.D{{"$eq", []interface{}{"$$predicate", false}}},
											bson.D{{"$eq", []interface{}{"$$predicate", 0}}},
											bson.D{{"$eq", []interface{}{"$$predicate", "0"}}},
											bson.D{{"$eq", []interface{}{"$$predicate", "-0"}}},
											bson.D{{"$eq", []interface{}{"$$predicate", "0.0"}}},
											bson.D{{"$eq", []interface{}{"$$predicate", "-0.0"}}},
											bson.D{{"$eq", []interface{}{"$$predicate", nil}}},
										}}},
										false,
										true,
									}},
								}},
							}},
						},
					}}},
					{{"$match", bson.M{
						"__predicate": true,
					}}},
					{{"$project", bson.M{
						"foo_DOT_a": "$a",
					}}},
				},
			)

			test("select a from foo where b < c",
				[]bson.D{
					{{"$addFields", bson.M{
						"__predicate": bson.D{
							{"$let", bson.D{
								{"vars", bson.M{
									"predicate": bson.M{
										"$cond": []interface{}{
											bson.M{
												"$or": []interface{}{
													bson.M{
														"$eq": []interface{}{
															bson.M{
																"$ifNull": []interface{}{
																	"$b",
																	nil,
																},
															},
															nil,
														},
													},
													bson.M{
														"$eq": []interface{}{
															bson.M{
																"$ifNull": []interface{}{
																	"$c",
																	nil,
																},
															},
															nil,
														},
													},
												},
											},
											nil,
											bson.M{
												"$lt": []interface{}{
													"$b",
													"$c",
												},
											},
										},
									},
								}},
								{"in", bson.D{
									{"$cond", []interface{}{
										bson.D{{"$or", []interface{}{
											bson.D{{"$eq", []interface{}{"$$predicate", false}}},
											bson.D{{"$eq", []interface{}{"$$predicate", 0}}},
											bson.D{{"$eq", []interface{}{"$$predicate", "0"}}},
											bson.D{{"$eq", []interface{}{"$$predicate", "-0"}}},
											bson.D{{"$eq", []interface{}{"$$predicate", "0.0"}}},
											bson.D{{"$eq", []interface{}{"$$predicate", "-0.0"}}},
											bson.D{{"$eq", []interface{}{"$$predicate", nil}}},
										}}},
										false,
										true,
									}},
								}},
							}},
						},
					}}},
					{{"$match", bson.M{
						"__predicate": true,
					}}},
					{{"$project", bson.M{
						"foo_DOT_a": "$a",
					}}},
				},
			)

			test("select `d.a` from merge_d_a where `d.a` = 10",
				[]bson.D{
					{{"$unwind", bson.D{
						{"includeArrayIndex", "d_idx"},
						{"path", "$d"},
					}}},
					{{"$unwind", bson.D{
						{"includeArrayIndex", "d.a_idx"},
						{"path", "$d.a"},
					}}},
					{{"$match", bson.M{
						"d.a": int64(10),
					}}},
					{{"$project", bson.M{
						"merge_d_a_DOT_d_DOT_a": "$d.a",
					}}},
				},
			)

			test("select `d.a` from merge_d_a where `d.a` = 10 OR `d.a` = 12",
				[]bson.D{
					{{"$unwind", bson.D{
						{"includeArrayIndex", "d_idx"},
						{"path", "$d"},
					}}},
					{{"$unwind", bson.D{
						{"includeArrayIndex", "d.a_idx"},
						{"path", "$d.a"},
					}}},
					{{"$match", bson.M{
						"$or": []interface{}{
							bson.M{"d.a": int64(10)},
							bson.M{"d.a": int64(12)},
						},
					}}},
					{{"$project", bson.M{
						"merge_d_a_DOT_d_DOT_a": "$d.a",
					}}},
				},
			)

			test("select c from merge_c where c = 10",
				[]bson.D{
					{{"$match", bson.M{
						"c": int64(10),
					}}},
					{{"$unwind", bson.D{
						{"includeArrayIndex", "c_idx"},
						{"path", "$c"},
					}}},
					{{"$match", bson.M{
						"c": int64(10),
					}}},
					{{"$project", bson.M{
						"merge_c_DOT_c": "$c",
					}}},
				},
			)

			test("select c from merge_c where c > 5 AND c < 10",
				[]bson.D{
					{{"$match", bson.M{
						"$and": []interface{}{
							bson.M{"c": bson.M{"$gt": int64(5)}},
							bson.M{"c": bson.M{"$lt": int64(10)}},
						},
					}}},
					{{"$unwind", bson.D{
						{"includeArrayIndex", "c_idx"},
						{"path", "$c"},
					}}},
					{{"$match", bson.M{
						"$and": []interface{}{
							bson.M{"c": bson.M{"$gt": int64(5)}},
							bson.M{"c": bson.M{"$lt": int64(10)}},
						},
					}}},
					{{"$project", bson.M{
						"merge_c_DOT_c": "$c",
					}}},
				},
			)
		})

		Convey("group by", func() {
			test("select a, b from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"foo_DOT_a": bson.M{
							"$first": "$a",
						},
						"foo_DOT_b": bson.M{
							"$first": "$b",
						},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a": "$foo_DOT_a",
						"foo_DOT_b": "$foo_DOT_b",
					}}},
				},
			)

			test("select a, b, c from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"foo_DOT_a": bson.M{
							"$first": "$a",
						},
						"foo_DOT_b": bson.M{
							"$first": "$b",
						},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a": "$foo_DOT_a",
						"foo_DOT_b": "$foo_DOT_b",
						"foo_DOT_c": "$_id.foo_DOT_c",
					}}},
				},
			)

			test("select a, b, c + a from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"foo_DOT_a": bson.M{
							"$first": "$a",
						},
						"foo_DOT_b": bson.M{
							"$first": "$b",
						},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a":           "$foo_DOT_a",
						"foo_DOT_b":           "$foo_DOT_b",
						"foo_DOT_c+foo_DOT_a": bson.M{"$add": []interface{}{"$_id.foo_DOT_c", "$foo_DOT_a"}},
					}}},
				},
			)

			test("select max(a), max(b) from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"max(foo_DOT_a)": bson.M{
							"$max": "$a",
						},
						"max(foo_DOT_b)": bson.M{
							"$max": "$b",
						},
					}}},
					{{"$project", bson.M{
						"max(foo_DOT_a)": "$max(foo_DOT_a)",
						"max(foo_DOT_b)": "$max(foo_DOT_b)",
					}}},
				},
			)

			test("select c, max(a), max(b) from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"max(foo_DOT_a)": bson.M{
							"$max": "$a",
						},
						"max(foo_DOT_b)": bson.M{
							"$max": "$b",
						},
					}}},
					{{"$project", bson.M{
						"foo_DOT_c":      "$_id.foo_DOT_c",
						"max(foo_DOT_a)": "$max(foo_DOT_a)",
						"max(foo_DOT_b)": "$max(foo_DOT_b)",
					}}},
				},
			)

			test("select a, max(b) from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"foo_DOT_a": bson.M{
							"$first": "$a",
						},
						"max(foo_DOT_b)": bson.M{
							"$max": "$b",
						},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a":      "$foo_DOT_a",
						"max(foo_DOT_b)": "$max(foo_DOT_b)",
					}}},
				},
			)

			test("select a, max(distinct b) from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"foo_DOT_a": bson.M{
							"$first": "$a",
						},
						"distinct foo_DOT_b": bson.M{
							"$addToSet": "$b",
						},
					}}},
					{{"$project", bson.M{
						"_id":                     0,
						"foo_DOT_a":               "$foo_DOT_a",
						"max(distinct foo_DOT_b)": bson.M{"$max": "$distinct foo_DOT_b"},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a":               "$foo_DOT_a",
						"max(distinct foo_DOT_b)": "$max(distinct foo_DOT_b)",
					}}},
				},
			)

			test("select a, max(distinct b), c from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"foo_DOT_a": bson.M{
							"$first": "$a",
						},
						"distinct foo_DOT_b": bson.M{
							"$addToSet": "$b",
						},
					}}},
					{{"$project", bson.M{
						"_id":                     0,
						"foo_DOT_c":               "$_id.foo_DOT_c",
						"foo_DOT_a":               "$foo_DOT_a",
						"max(distinct foo_DOT_b)": bson.M{"$max": "$distinct foo_DOT_b"},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a":               "$foo_DOT_a",
						"foo_DOT_c":               "$foo_DOT_c",
						"max(distinct foo_DOT_b)": "$max(distinct foo_DOT_b)",
					}}},
				},
			)

			test("select a + max(b) from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"foo_DOT_a": bson.M{
							"$first": "$a",
						},
						"max(foo_DOT_b)": bson.M{
							"$max": "$b",
						},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a+max(foo_DOT_b)": bson.M{"$add": []interface{}{"$foo_DOT_a", "$max(foo_DOT_b)"}},
					}}},
				},
			)

			// TODO: algebrizer isn't taking into account grouping keys. I can't figure outer
			// if this is an actual problem, or just a different way of handling it.
			// test("select a + b from foo group by a + b",
			// 	[]bson.D{
			// 		{{"$group", bson.M{
			// 			"_id": bson.D{{"foo_DOT_a+foo_DOT_b", bson.M{"$add": []interface{}{"$a", "$b"}}}},
			// 			"foo_DOT_a": bson.M{
			// 				"$first": "$a",
			// 			},
			// 			"foo_DOT_b": bson.M{
			// 				"$first": "$b",
			// 			},
			// 		}}},
			// 		{{"$project", bson.M{
			// 			"_id":                 0,
			// 			"foo_DOT_a":           "$foo_DOT_a",
			// 			"foo_DOT_b":           "$foo_DOT_b",
			// 			"foo_DOT_a+foo_DOT_b": "$_id.foo_DOT_a+foo_DOT_b",
			// 		}}},
			// 		{{"$project", bson.M{
			// 			"foo_DOT_a+foo_DOT_b": "$_id.foo_DOT_a+foo_DOT_b",
			// 		}}},
			// 	},
			// )

			test("select a + c + max(b) from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"foo_DOT_a": bson.M{
							"$first": "$a",
						},
						"max(foo_DOT_b)": bson.M{
							"$max": "$b",
						},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a+foo_DOT_c+max(foo_DOT_b)": bson.M{"$add": []interface{}{bson.M{"$add": []interface{}{"$foo_DOT_a", "$_id.foo_DOT_c"}}, "$max(foo_DOT_b)"}},
					}}},
				},
			)

			test("select a + max(distinct b) from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"foo_DOT_a": bson.M{
							"$first": "$a",
						},
						"distinct foo_DOT_b": bson.M{
							"$addToSet": "$b",
						},
					}}},
					{{"$project", bson.M{
						"_id":                     0,
						"foo_DOT_a":               "$foo_DOT_a",
						"max(distinct foo_DOT_b)": bson.M{"$max": "$distinct foo_DOT_b"},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a+max(distinct foo_DOT_b)": bson.M{"$add": []interface{}{"$foo_DOT_a", "$max(distinct foo_DOT_b)"}},
					}}},
				},
			)

			test("select c + max(distinct b) from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"distinct foo_DOT_b": bson.M{
							"$addToSet": "$b",
						},
					}}},
					{{"$project", bson.M{
						"_id":                     0,
						"foo_DOT_c":               "$_id.foo_DOT_c",
						"max(distinct foo_DOT_b)": bson.M{"$max": "$distinct foo_DOT_b"},
					}}},
					{{"$project", bson.M{
						"foo_DOT_c+max(distinct foo_DOT_b)": bson.M{"$add": []interface{}{"$foo_DOT_c", "$max(distinct foo_DOT_b)"}},
					}}},
				},
			)

			test("select max(distinct a + b) from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"distinct foo_DOT_a+foo_DOT_b": bson.M{
							"$addToSet": bson.M{"$add": []interface{}{"$a", "$b"}},
						},
					}}},
					{{"$project", bson.M{
						"_id": 0,
						"max(distinct foo_DOT_a+foo_DOT_b)": bson.M{"$max": "$distinct foo_DOT_a+foo_DOT_b"},
					}}},
					{{"$project", bson.M{
						"max(distinct foo_DOT_a+foo_DOT_b)": "$max(distinct foo_DOT_a+foo_DOT_b)",
					}}},
				},
			)

			test("select a + max(distinct a + b) from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"foo_DOT_a": bson.M{
							"$first": "$a",
						},
						"distinct foo_DOT_a+foo_DOT_b": bson.M{
							"$addToSet": bson.M{"$add": []interface{}{"$a", "$b"}},
						},
					}}},
					{{"$project", bson.M{
						"_id":                               0,
						"foo_DOT_a":                         "$foo_DOT_a",
						"max(distinct foo_DOT_a+foo_DOT_b)": bson.M{"$max": "$distinct foo_DOT_a+foo_DOT_b"},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a+max(distinct foo_DOT_a+foo_DOT_b)": bson.M{"$add": []interface{}{"$foo_DOT_a", "$max(distinct foo_DOT_a+foo_DOT_b)"}},
					}}},
				},
			)

			test("select sum(a) from foo",
				[]bson.D{
					{{"$group", bson.M{
						"_id":            bson.D{},
						"sum(foo_DOT_a)": bson.M{"$sum": "$a"},
						"sum(foo_DOT_a)_count": bson.M{
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
					{{"$project", bson.M{
						"_id": 0,
						"sum(foo_DOT_a)": bson.M{
							"$cond": []interface{}{
								bson.M{"$or": []interface{}{
									bson.M{"$eq": []interface{}{bson.M{"$ifNull": []interface{}{"$sum(foo_DOT_a)_count", nil}}, nil}},
									bson.M{"$eq": []interface{}{"$sum(foo_DOT_a)_count", 0}},
									bson.M{"$eq": []interface{}{"$sum(foo_DOT_a)_count", false}},
								}},
								bson.M{"$literal": nil},
								"$sum(foo_DOT_a)",
							},
						},
					}}},
					{{"$project", bson.M{
						"sum(foo_DOT_a)": "$sum(foo_DOT_a)",
					}}},
				},
			)

			test("select count(*) from foo",
				[]bson.D{
					{{"$group", bson.M{
						"_id":      bson.D{},
						"count(*)": bson.M{"$sum": 1},
					}}},
					{{"$project", bson.M{
						"count(*)": "$count(*)",
					}}},
				},
			)

			test("select count(a) from foo",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{},
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
					{{"$project", bson.M{
						"count(foo_DOT_a)": "$count(foo_DOT_a)",
					}}},
				},
			)

			test("select count(distinct b) from foo",
				[]bson.D{
					{{"$group", bson.M{
						"_id":                bson.D{},
						"distinct foo_DOT_b": bson.M{"$addToSet": "$b"},
					}}},
					{{"$project", bson.M{
						"_id": 0,
						"count(distinct foo_DOT_b)": bson.M{
							"$sum": bson.M{
								"$map": bson.M{
									"input": "$distinct foo_DOT_b",
									"as":    "i",
									"in": bson.M{
										"$cond": []interface{}{
											bson.M{"$eq": []interface{}{bson.M{"$ifNull": []interface{}{"$$i", nil}}, nil}},
											0,
											1,
										},
									},
								},
							},
						},
					}}},
					{{"$project", bson.M{
						"count(distinct foo_DOT_b)": "$count(distinct foo_DOT_b)",
					}}},
				},
			)
		})

		Convey("having", func() {
			test("select max(a) from foo group by c having max(b) = 10",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"max(foo_DOT_a)": bson.M{
							"$max": "$a",
						},
						"max(foo_DOT_b)": bson.M{
							"$max": "$b",
						},
					}}},
					{{"$match", bson.M{
						"max(foo_DOT_b)": int64(10),
					}}},
					{{"$project", bson.M{
						"max(foo_DOT_a)": "$max(foo_DOT_a)",
					}}},
				},
			)
		})

		Convey("order by", func() {
			test("select a from foo order by b",
				[]bson.D{
					{{"$sort", bson.D{
						{"b", 1},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a": "$a",
					}}},
				},
			)

			test("select a from foo order by a, b desc",
				[]bson.D{
					{{"$sort", bson.D{
						{"a", 1},
						{"b", -1},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a": "$a",
					}}},
				},
			)

			test("select a from foo group by a order by max(b)",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_a", "$a"},
						},
						"max(foo_DOT_b)": bson.M{
							"$max": "$b",
						},
					}}},
					{{"$sort", bson.D{
						{"max(foo_DOT_b)", 1},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a": "$_id.foo_DOT_a",
					}}},
				},
			)

			Convey("no push down, project columns", func() {
				test("select a from foo order by a > b",
					[]bson.D{
						{{"$addFields", bson.M{
							"foo_DOT_a>foo_DOT_b": bson.M{
								"$cond": []interface{}{
									bson.M{
										"$or": []interface{}{
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
											bson.M{
												"$eq": []interface{}{bson.M{
													"$ifNull": []interface{}{
														"$b",
														nil,
													},
												},
													nil,
												},
											},
										},
									},
									nil,
									bson.M{
										"$gt": []interface{}{
											"$a",
											"$b",
										},
									},
								},
							},
						}}},
						{{"$sort", bson.D{
							{
								Name:  "foo_DOT_a>foo_DOT_b",
								Value: int(1),
							},
						}}},
						{{"$project", bson.M{
							"foo_DOT_a": "$a",
						}}},
					},
				)
			})
		})

		Convey("limit", func() {
			test("select a from foo limit 10",
				[]bson.D{
					{{"$limit", int64(10)}},
					{{"$project", bson.M{
						"foo_DOT_a": "$a",
					}}},
				},
			)

			test("select a from foo limit 10, 20",
				[]bson.D{
					{{"$skip", int64(10)}},
					{{"$limit", int64(20)}},
					{{"$project", bson.M{
						"foo_DOT_a": "$a",
					}}},
				},
			)

			testNoPushdown("select a from foo limit 18446744073709551614")
		})

		Convey("custom mongo filter", func() {
			test(`select a from foo where filter='{"a": {"$gt": 3}}'`,
				[]bson.D{
					{{"$match", bson.M{
						"a": map[string]interface{}{
							"$gt": float64(3),
						},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a": "$a",
					}}},
				},
			)

			test(`select a from foo where filter='{"a": {"$elemMatch": {"$gte": 80, "$lt": 85}}}' or b = 40`,
				[]bson.D{
					{{"$match", bson.M{
						"$or": []interface{}{
							bson.M{
								"a": map[string]interface{}{
									"$elemMatch": map[string]interface{}{
										"$gte": float64(80),
										"$lt":  float64(85),
									}},
							},
							bson.M{
								"b": int64(40),
							},
						}},
					}},
					{{"$project", bson.M{
						"foo_DOT_a": "$a",
					}}},
				},
			)
		})

		Convey("Subject: UniqueFieldNameGeneration", func() {
			test("select trim('') from foo",
				[]bson.D{
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
					}}},
				})
			test("select trim(''), trim(a) from foo",
				[]bson.D{
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
						"a": "$a",
					}}},
				})
			test("select trim(''), trim(a), trim(' ') from foo",
				[]bson.D{
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
						"a": "$a",
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 0): bson.M{
							"$literal": "",
						},
					}}},
				})
			test("select trim(''), trim(' '), trim('  '), trim('   ') from foo",
				[]bson.D{
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 0): bson.M{
							"$literal": "",
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 1): bson.M{
							"$literal": "",
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 2): bson.M{
							"$literal": "",
						},
					}}},
				})
			test("select a, b, trim('   ') from foo",
				[]bson.D{
					{{"$project", bson.M{
						"foo_DOT_a": "$a",
						"foo_DOT_b": "$b",
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
					}}},
				})
			test("select trim(a), trim(''), a, trim(' ') from foo",
				[]bson.D{
					{{"$project", bson.M{
						"a": "$a",
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
						"foo_DOT_a": "$a",
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 0): bson.M{
							"$literal": "",
						},
					}}},
				})
			test("select trim('') from (select trim('') from foo) as subq",
				[]bson.D{
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
					}}},
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
					}}},
				})
			test("select trim('') from (select trim('') from (select trim('') from foo) as subq1) as subq2",
				[]bson.D{
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
					}}},
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
					}}},
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
					}}},
				})
			test("select trim('') from (select trim('') from (select trim('') from (select trim('') from foo) as subq1) as subq2) as subq3",
				[]bson.D{
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
					}}},
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
					}}},
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
					}}},
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
					}}},
				})
			test("select trim(''), trim(' ') from foo inner join (select trim(''), trim(' ') from bar) as t2",
				[]bson.D{
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 0): bson.M{
							"$literal": "",
						},
					}}},
				})
			test("select trim(''), trim(' '), trim('  ') from foo inner join (select trim(''), trim(' '), trim('  ') from bar) as t2",
				[]bson.D{
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 0): bson.M{
							"$literal": "",
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 1): bson.M{
							"$literal": "",
						},
					}}},
				})

			test(fmt.Sprintf("select '%v', trim('') from foo", emptyFieldNamePrefix),
				[]bson.D{
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": emptyFieldNamePrefix,
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 0): bson.M{
							"$literal": "",
						},
					}}},
				})

			test(fmt.Sprintf("select '%v', '%v', trim('') from foo", emptyFieldNamePrefix, fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 0)),
				[]bson.D{
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": emptyFieldNamePrefix,
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 0): bson.M{
							"$literal": fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 0),
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 1): bson.M{
							"$literal": "",
						},
					}}},
				})

			test(fmt.Sprintf("select '%v', '%v', trim('') from foo", emptyFieldNamePrefix, fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 1)),
				[]bson.D{
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": emptyFieldNamePrefix,
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 1): bson.M{
							"$literal": fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 1),
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 0): bson.M{
							"$literal": "",
						},
					}}},
				})

			test(fmt.Sprintf("select '%v', '%v', trim(''), trim(' ') from foo", emptyFieldNamePrefix, fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 1)),
				[]bson.D{
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": emptyFieldNamePrefix,
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 0): bson.M{
							"$literal": "",
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 1): bson.M{
							"$literal": fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 1),
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 2): bson.M{
							"$literal": "",
						},
					}}},
				})
			test(fmt.Sprintf("select '%v', '%v', trim(''), trim(' ') from foo", fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 0), fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 1)),
				[]bson.D{
					{{"$project", bson.M{
						emptyFieldNamePrefix: bson.M{
							"$literal": "",
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 0): bson.M{
							"$literal": fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 0),
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 1): bson.M{
							"$literal": fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 1),
						},
						fmt.Sprintf("%v_%v", emptyFieldNamePrefix, 2): bson.M{
							"$literal": "",
						},
					}}},
				})
		})

	})
}

type subqueryFinder struct {
	count         int
	firstSubquery *SQLSubqueryExpr
}

func (v *subqueryFinder) visit(n node) (node, error) {
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	switch typedN := n.(type) {
	case *SQLSubqueryExpr:
		v.count += 1
		v.firstSubquery = typedN
	}

	return n, nil
}

type sourceStageReplacer struct {
	data            []bson.D
	existing        int
	replaced        int
	lastSourceStage *BSONSourceStage
}

func (v *sourceStageReplacer) visit(n node) (node, error) {
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	switch typedN := n.(type) {
	case *BSONSourceStage:
		v.existing += 1
		if v.lastSourceStage == nil {
			v.lastSourceStage = typedN
		}
	case *MongoSourceStage:
		bs := NewBSONSourceStage(typedN.selectIDs[0], typedN.tableNames[0], typedN.collation, v.data[0:1])
		v.data = v.data[1:]
		v.replaced += 1
		n = bs
	}

	return n, nil
}

func TestOptimizeSubqueryPlan(t *testing.T) {
	testSchema, err := schema.New(testSchema4)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}
	testInfo := getMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)
	testVariables := createTestVariables(testInfo)
	testCatalog := getCatalogFromSchema(testSchema, testVariables)
	defaultDbName := "test"

	testOptimize := func(sql string, expected ...[]bson.D) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			plan, err := AlgebrizeQuery(statement, defaultDbName, testVariables, testCatalog)
			So(err, ShouldBeNil)
			ctx := createTestConnectionCtx(testInfo)
			optimized, err := optimizeSubqueries(ctx, ctx.Logger(""), plan, false)
			So(err, ShouldBeNil)

			finder := &subqueryFinder{}
			finder.visit(optimized)

			subqueryPlan := finder.firstSubquery.plan

			pg := &pipelineGatherer{}
			pg.visit(subqueryPlan)

			actual := pg.pipelines

			So(actual, ShouldResembleDiffed, expected)
		})
	}

	testExecute := func(sql string, data []bson.D) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			plan, err := AlgebrizeQuery(statement, defaultDbName, testVariables, testCatalog)
			So(err, ShouldBeNil)

			//fmt.Printf("\n%+v\n", PrettyPrintPlan(plan))

			sourceReplacer := &sourceStageReplacer{data: data}
			replaced, err := sourceReplacer.visit(plan)
			So(err, ShouldBeNil)
			So(sourceReplacer.existing, ShouldEqual, 0)

			//fmt.Printf("\n%+v\n", PrettyPrintPlan(replaced.(PlanStage)))

			ctx := createTestConnectionCtx(testInfo)
			optimized, err := optimizeSubqueries(ctx, ctx.Logger(""), replaced, true)
			So(err, ShouldBeNil)

			sourceReplacer = &sourceStageReplacer{}
			sourceReplacer.visit(optimized)
			So(sourceReplacer.existing, ShouldEqual, 1)
			So(sourceReplacer.replaced, ShouldEqual, 0)
		})
	}

	Convey("Subject: OptimizeSubqueryPlan", t, func() {
		Convey("subquery optimization", func() {
			testOptimize("select a, (select b from bar) from foo",
				[]bson.D{
					{{"$project", bson.M{
						"bar_DOT_b": "$b",
					}}},
				})
			testOptimize("select exists(select a from bar) from foo",
				[]bson.D{
					{{"$project", bson.M{
						"bar_DOT_a": "$a",
					}}},
				})
			testOptimize("select a from bar where `a` = (select `b` from bar where b=2)",
				[]bson.D{
					{{"$match", bson.M{
						"b": int64(2),
					}}},
					{{"$project", bson.M{
						"bar_DOT_b": "$b",
					}}},
				})
			testOptimize("select a from bar where `a` = (select `b` from bar where b = (select a from bar where a=1))",
				[]bson.D{
					bson.D{{"$project", bson.M{
						"bar_DOT_b": "$b",
					}}},
					bson.D{{"$project", bson.M{
						"bar_DOT_b": "$bar_DOT_b",
					}}},
				},
				[]bson.D{
					bson.D{{"$match", bson.M{
						"a": int64(1),
					}}},
					bson.D{{"$project", bson.M{
						"bar_DOT_a": "$a",
					}}},
				})
			testOptimize("select a from bar where (`a`, `b`) = (select `c`, `b` from foo where b=2)",
				[]bson.D{
					{{"$match", bson.M{
						"b": int64(2),
					}}},
					{{"$project", bson.M{
						"foo_DOT_c": "$c",
						"foo_DOT_b": "$b",
					}}},
				})
		})
		Convey("subquery execution and replacement", func() {
			testExecute("select a, (select b from bar) from foo",
				[]bson.D{
					{{"b", 1}},
					{{"a", 1}},
				})
			testExecute("select a from bar where `a` = (select `b` from bar where b=2)",
				[]bson.D{
					{{"b", 2}},
					{{"a", 2}},
				})
			testExecute("select a from bar where `a` = (select `b` from bar where b = (select a from bar where a=1))",
				[]bson.D{
					{{"a", 1}},
					{{"b", 1}},
					{{"a", 1}},
				})
			testExecute("select a from bar where (`a`, `b`) = (select `c`, `b` from foo where b=2)",
				[]bson.D{
					{{"b", 1}, {"c", 1}},
					{{"a", 1}},
				})
		})
	})
}

func TestOptimizeEvaluations(t *testing.T) {

	type test struct {
		sql      string
		expected string
		result   SQLExpr
	}

	testSchema, err := schema.New(testSchema4)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}

	testInfo := getMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	runTests := func(tests []test) {
		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be optimized to %q", t.sql, t.expected), func() {
				e, err := getSQLExpr(schema, dbOne, tableTwoName, t.sql)
				So(err, ShouldBeNil)

				ctx := createTestEvalCtx(testInfo)
				result, err := optimizeEvaluations(e, ctx, ctx.Logger(""))
				So(err, ShouldBeNil)
				So(result, ShouldResemble, t.result)
			})
		}
	}

	Convey("Subject: optimizeEvaluations", t, func() {

		tests := []test{
			test{"3 / '3'", "1", SQLFloat(1)},
			test{"3 * '3'", "9", SQLInt(9)},
			test{"3 + '3'", "6", SQLInt(6)},
			test{"3 - '3'", "0", SQLInt(0)},
			test{"3 div '3'", "1", SQLInt(1)},
			test{"3 = '3'", "true", SQLTrue},
			test{"3 <= '3'", "true", SQLTrue},
			test{"3 >= '3'", "true", SQLTrue},
			test{"3 < '3'", "false", SQLFalse},
			test{"3 > '3'", "false", SQLFalse},
			test{"3 <=> '3'", "true", SQLTrue},
			test{"3 = a", "a = 3", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"3 < a", "a > 3", &SQLGreaterThanExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"3 <= a", "a >= 3", &SQLGreaterThanOrEqualExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"3 > a", "a < 3", &SQLLessThanExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"3 >= a", "a <= 3", &SQLLessThanOrEqualExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"3 <> a", "a <> 3", &SQLNotEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"3 + 3 = 6", "true", SQLTrue},
			test{"3 <=> 3", "true", SQLTrue},
			test{"NULL <=> 3", "false", SQLFalse},
			test{"3 <=> NULL", "false", SQLFalse},
			test{"NULL <=> NULL", "true", SQLTrue},
			test{"3 / (3 - 2) = a", "a = 3", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLFloat(3)}},
			test{"3 + 3 = 6 AND 1 >= 1 AND 3 = a", "a = 3", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"3 / (3 - 2) = a AND 4 - 2 = b", "a = 3 AND b = 2",
				&SQLAndExpr{
					&SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLFloat(3)},
					&SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "b", schema.SQLInt, schema.MongoInt), SQLInt(2)}}},
			test{"3 + 3 = 6 OR a = 3", "true", SQLTrue},
			test{"3 + 3 = 5 OR a = 3", "a = 3", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"0 OR NULL", "null", SQLNull},
			test{"1 OR NULL", "true", SQLTrue},
			test{"NULL OR NULL", "null", SQLNull},
			test{"0 AND 6+1 = 6", "false", SQLFalse},
			test{"3 + 3 = 5 AND a = 3", "false", SQLFalse},
			test{"0 AND NULL", "false", SQLFalse},
			test{"1 AND NULL", "null", SQLNull},
			test{"1 AND 6+0 = 6", "true", SQLTrue},
			test{"3 + 3 = 6 AND a = 3", "a = 3", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"(3 + 3 = 5) XOR a = 3", "a = 3", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"(3 + 3 = 6) XOR a = 3", "a <> 3", &SQLNotExpr{operand: &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}}},
			test{"(13 + 9 > 6) XOR (a = 4)", "a <> 4", &SQLNotExpr{operand: &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(4)}}},
			test{"(8 / 5 = 9) XOR (a = 5)", "a = 5", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(5)}},
			test{"false XOR 23", "true", SQLTrue},
			test{"true XOR 23", "false", SQLFalse},
			test{"a = 23 XOR true", "a <> 23", &SQLNotExpr{operand: &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(23)}}},
			test{"!3", "0", SQLFalse},
			test{"!NULL", "null", SQLNull},
			test{"a = ~1", "a = 18446744073709551614", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLUint64(18446744073709551614)}},
			test{"a = ~2398238912332232323", "a = 16048505161377319292", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLUint64(16048505161377319292)}},
			test{"DAYNAME('2016-1-1')", "Friday", SQLVarchar("Friday")},
			test{"(8-7)", "1", SQLInt(1)},
			test{"a LIKE NULL", "null", SQLNull},
			test{"4 LIKE NULL", "null", SQLNull},
			test{"a = NULL", "null", SQLNull},
			test{"a > NULL", "null", SQLNull},
			test{"a >= NULL", "null", SQLNull},
			test{"a < NULL", "null", SQLNull},
			test{"a <= NULL", "null", SQLNull},
			test{"a != NULL", "null", SQLNull},
			test{"(1, 3) > (3, 4)", "SQLFalse", SQLFalse},
			test{"(4, 3) > (3, 4)", "SQLTrue", SQLTrue},
			test{"(4, 31) > (4, 4)", "SQLTrue", SQLTrue},

			test{"abs(NULL)", "null", SQLNull},
			test{"abs(-10)", "10", SQLFloat(10)},
			test{"ascii(NULL)", "null", SQLNull},
			test{"ascii('a')", "97", SQLInt(97)},
			test{"char_length(NULL)", "null", SQLNull},
			test{"character_length(NULL)", "null", SQLNull},
			test{"concat(NULL, a)", "null", SQLNull},
			test{"concat(a, NULL)", "null", SQLNull},
			test{"concat('go', 'lang')", "golang", SQLVarchar("golang")},
			test{"concat_ws(NULL, a)", "null", SQLNull},
			test{"convert(NULL, SIGNED)", "null", SQLNull},
			test{"elt(NULL, 'a', 'b')", "null", SQLNull},
			test{"elt(4, 'a', 'b')", "null", SQLNull},
			test{"exp(NULL)", "null", SQLNull},
			test{"exp(2)", "7.38905609893065", SQLFloat(7.38905609893065)},
			test{"greatest(a, NULL)", "null", SQLNull},
			test{"greatest(2, 3)", "3", SQLInt(3)},
			test{"ifnull(NULL, a)", "bar.a", NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt)},
			test{"ifnull(10, a)", "10", SQLInt(10)},
			test{"interval(NULL, a)", "-1", SQLInt(-1)},
			test{"interval(0, 1)", "0", SQLInt(0)},
			test{"interval(1, 2, 3, 4)", "1", SQLInt(0)},
			test{"interval(1, 1, 2, 3)", "1", SQLInt(1)},
			test{"interval(-1, NULL, NULL, -0.5, 3, 4)", "1", SQLInt(2)},
			test{"interval(-3.4, -4, -3.6, -3.4, -3, 1, 2)", "3", SQLInt(3)},
			test{"interval(8, -4, 0, 7, 8)", "4", SQLInt(4)},
			test{"interval(8, -3, 1, 7, 7)", "1", SQLInt(4)},
			test{"interval(7.7, -3, 1, 7, 7)", "1", SQLInt(4)},
			test{"least(a, NULL)", "null", SQLNull},
			test{"least(2, 3)", "2", SQLInt(2)},
			test{"locate('bar', 'foobar', NULL)", "null", SQLNull},
			test{"locate('bar', 'foobar')", "4", SQLInt(4)},
			test{"makedate(2000, NULL)", "null", SQLNull},
			test{"makedate(NULL, 10)", "null", SQLNull},
			test{"mid('foobar', NULL, 2)", "null", SQLNull},
			test{"mod(10, 2)", "0", SQLFloat(0)},
			test{"mod(NULL, 2)", "null", SQLNull},
			test{"mod(10, NULL)", "null", SQLNull},
			test{"nullif(NULL, a)", "null", SQLNull},
			test{"nullif(a, NULL)", "bar.a", NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt)},
			test{"pow(a, NULL)", "null", SQLNull},
			test{"pow(NULL, a)", "null", SQLNull},
			test{"pow(2,2)", "4", SQLFloat(4)},
			test{"round(NULL, 2)", "null", SQLNull},
			test{"round(2, NULL)", "null", SQLNull},
			test{"round(2, 2)", "2", SQLFloat(2)},
			test{"repeat('a', NULL)", "null", SQLNull},
			test{"repeat(NULL, 3)", "null", SQLNull},
			test{"substring(NULL, 2)", "null", SQLNull},
			test{"substring(NULL, 2, 3)", "null", SQLNull},
			test{"substring('foobar', NULL)", "null", SQLNull},
			test{"substring('foobar', NULL, 2)", "null", SQLNull},
			test{"substring('foobar', 2, NULL)", "null", SQLNull},
			test{"substring('foobar', 2, 3)", "oob", SQLVarchar("oob")},
			test{"substring_index(NULL, 'o', 0)", "", SQLNull},
			test{"substring_index('foobar', 'o', 0)", "", SQLVarchar("")},
		}

		runTests(tests)

	})
}

func TestOptimizeEvaluationFailures(t *testing.T) {

	type test struct {
		sql string
		err error
	}

	testSchema, err := schema.New(testSchema4)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}

	testInfo := getMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	runTests := func(tests []test) {
		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should fail with error %q", t.sql, t.err), func() {
				e, err := getSQLExpr(schema, dbOne, tableTwoName, t.sql)
				So(err, ShouldBeNil)

				ctx := createTestEvalCtx(testInfo)
				_, err = optimizeEvaluations(e, ctx, ctx.Logger(""))
				So(err, ShouldResemble, t.err)
			})
		}
	}

	Convey("Subject: optimizeEvaluations failures", t, func() {

		tests := []test{
			test{"pow(-2,2.2)", mysqlerrors.Defaultf(mysqlerrors.ER_DATA_OUT_OF_RANGE, "DOUBLE", "pow(-2,2.2)")},
			test{"pow(0,-2.2)", mysqlerrors.Defaultf(mysqlerrors.ER_DATA_OUT_OF_RANGE, "DOUBLE", "pow(0,-2.2)")},
			test{"pow(0,-5)", mysqlerrors.Defaultf(mysqlerrors.ER_DATA_OUT_OF_RANGE, "DOUBLE", "pow(0,-5)")},
		}

		runTests(tests)

	})
}
