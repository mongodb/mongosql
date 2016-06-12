package evaluator

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
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
			v.pipelines = append(v.pipelines, typedN.pipeline)
		}
	}

	return n, nil
}

func TestOptimizePlan(t *testing.T) {
	testSchema, err := schema.New(testSchema1)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}
	defaultDbName := "test"

	test := func(sql string, expected ...[]bson.D) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			selectStatement := statement.(parser.SelectStatement)
			plan, err := Algebrize(selectStatement, defaultDbName, testSchema)
			So(err, ShouldBeNil)
			actualPlan, err := OptimizePlan(plan)
			So(err, ShouldBeNil)

			// fmt.Printf("\nPLAN: %# v", pretty.Formatter(plan))
			// fmt.Printf("\nOPTIMIZED: %# v", pretty.Formatter(actualPlan))

			pg := &pipelineGatherer{}
			pg.visit(actualPlan)

			actual := pg.pipelines

			So(actual, ShouldResembleDiffed, expected)
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
					test("select foo.a, bar.b from foo inner join bar on foo.a = bar.a",
						[]bson.D{
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

					test("select foo.a, bar.b from foo inner join bar on foo.a = bar.a AND foo.b > 10",
						[]bson.D{
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

					test("select foo.a, bar.b from foo inner join bar on foo.a = bar.a AND bar.b > 10",
						[]bson.D{
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
								"__joined_bar.b": bson.M{"$gt": int64(10)},
							}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"bar_DOT_b": "$__joined_bar.b",
							}}},
						},
					)

					test("select foo.a, bar.b from foo, bar where foo.a = bar.a",
						[]bson.D{
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

					test("select foo.a, bar.b from foo left join bar on foo.a = bar.a AND foo.b > 10",
						[]bson.D{
							{{"$lookup", bson.M{
								"from":         "bar",
								"localField":   "a",
								"foreignField": "a",
								"as":           "__joined_bar",
							}}},
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": true,
							}}},
							{{"$project", bson.M{
								"_id": 1,
								"a":   1,
								"b":   1,
								"c":   1,
								"d.e": 1,
								"d.f": 1,
								"g":   1,
								"__joined_bar": bson.M{"$cond": bson.M{
									"if":   bson.M{"$gt": []interface{}{"$b", bson.M{"$literal": SQLInt(10)}}},
									"then": "$__joined_bar",
									"else": nil,
								}},
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
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": true,
							}}},
							{{"$project", bson.M{
								"_id": 1,
								"a":   1,
								"b":   1,
								"c":   1,
								"d.e": 1,
								"d.f": 1,
								"g":   1,
								"__joined_bar": bson.M{"$cond": bson.M{
									"if":   bson.M{"$gt": []interface{}{"$__joined_bar.b", bson.M{"$literal": SQLInt(10)}}},
									"then": "$__joined_bar",
									"else": nil,
								}},
							}}},
							{{"$project", bson.M{
								"foo_DOT_a": "$a",
								"bar_DOT_b": "$__joined_bar.b",
							}}},
						},
					)
				})

				Convey("no push down", func() {
					test("select foo.a from foo inner join bar on foo.a < bar.a")
					test("select foo.a from foo inner join bar on foo.a < foo.b")
					test("select foo.a from foo, bar where foo.a < bar.a")
					test("select foo.a from foo left join bar on foo.a < bar.a")
					test("select foo.a from foo right join bar on foo.a < bar.a")
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

			Convey("subqueries", func() {
				test("select a, (select foo.b from bar) from foo",
					[]bson.D{
						{{"$project", bson.M{
							"foo_DOT_a": "$a",
							"b":         "$b",
						}}},
					},
				)

				test("select a, (select b from bar) from foo",
					[]bson.D{
						{{"$project", bson.M{
							"foo_DOT_a": "$a",
						}}},
					},
					[]bson.D{
						{{"$project", bson.M{
							"bar_DOT_b": "$b",
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
				},
			)

			test("select a from foo where b < c AND a = 10",
				[]bson.D{
					{{"$match", bson.M{
						"a": int64(10),
					}}},
				},
			)

			Convey("no push down", func() {
				test("select a from foo where b < c")
			})
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
						"_id":       0,
						"foo_DOT_a": "$foo_DOT_a",
						"foo_DOT_b": "$foo_DOT_b",
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
						"_id":       0,
						"foo_DOT_c": "$_id.foo_DOT_c",
						"foo_DOT_a": "$foo_DOT_a",
						"foo_DOT_b": "$foo_DOT_b",
					}}},
					{{"$project", bson.M{
						"foo_DOT_a": "$foo_DOT_a",
						"foo_DOT_b": "$foo_DOT_b",
						"foo_DOT_c": "$foo_DOT_c",
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
						"_id":       0,
						"foo_DOT_a": "$foo_DOT_a",
						"foo_DOT_b": "$foo_DOT_b",
						"foo_DOT_c": "$_id.foo_DOT_c",
					}}},
					{{"$project", bson.M{
						"foo_DOT_a":           "$foo_DOT_a",
						"foo_DOT_b":           "$foo_DOT_b",
						"foo_DOT_c+foo_DOT_a": bson.M{"$add": []interface{}{"$foo_DOT_c", "$foo_DOT_a"}},
					}}},
				},
			)

			test("select sum(a), sum(b) from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"sum(foo_DOT_a)": bson.M{
							"$sum": "$a",
						},
						"sum(foo_DOT_b)": bson.M{
							"$sum": "$b",
						},
					}}},
					{{"$project", bson.M{
						"_id":            0,
						"sum(foo_DOT_a)": "$sum(foo_DOT_a)",
						"sum(foo_DOT_b)": "$sum(foo_DOT_b)",
					}}},
					{{"$project", bson.M{
						"sum(foo_DOT_a)": "$sum(foo_DOT_a)",
						"sum(foo_DOT_b)": "$sum(foo_DOT_b)",
					}}},
				},
			)

			test("select c, sum(a), sum(b) from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"sum(foo_DOT_a)": bson.M{
							"$sum": "$a",
						},
						"sum(foo_DOT_b)": bson.M{
							"$sum": "$b",
						},
					}}},
					{{"$project", bson.M{
						"_id":            0,
						"foo_DOT_c":      "$_id.foo_DOT_c",
						"sum(foo_DOT_a)": "$sum(foo_DOT_a)",
						"sum(foo_DOT_b)": "$sum(foo_DOT_b)",
					}}},
					{{"$project", bson.M{
						"foo_DOT_c":      "$foo_DOT_c",
						"sum(foo_DOT_a)": "$sum(foo_DOT_a)",
						"sum(foo_DOT_b)": "$sum(foo_DOT_b)",
					}}},
				},
			)

			test("select a, sum(b) from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"foo_DOT_a": bson.M{
							"$first": "$a",
						},
						"sum(foo_DOT_b)": bson.M{
							"$sum": "$b",
						},
					}}},
					{{"$project", bson.M{
						"_id":            0,
						"foo_DOT_a":      "$foo_DOT_a",
						"sum(foo_DOT_b)": "$sum(foo_DOT_b)",
					}}},
					{{"$project", bson.M{
						"foo_DOT_a":      "$foo_DOT_a",
						"sum(foo_DOT_b)": "$sum(foo_DOT_b)",
					}}},
				},
			)

			test("select a, sum(distinct b) from foo group by c",
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
						"sum(distinct foo_DOT_b)": bson.M{"$sum": "$distinct foo_DOT_b"},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a":               "$foo_DOT_a",
						"sum(distinct foo_DOT_b)": "$sum(distinct foo_DOT_b)",
					}}},
				},
			)

			test("select a, sum(distinct b), c from foo group by c",
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
						"sum(distinct foo_DOT_b)": bson.M{"$sum": "$distinct foo_DOT_b"},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a":               "$foo_DOT_a",
						"foo_DOT_c":               "$foo_DOT_c",
						"sum(distinct foo_DOT_b)": "$sum(distinct foo_DOT_b)",
					}}},
				},
			)

			test("select a + sum(b) from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"foo_DOT_a": bson.M{
							"$first": "$a",
						},
						"sum(foo_DOT_b)": bson.M{
							"$sum": "$b",
						},
					}}},
					{{"$project", bson.M{
						"_id":            0,
						"foo_DOT_a":      "$foo_DOT_a",
						"sum(foo_DOT_b)": "$sum(foo_DOT_b)",
					}}},
					{{"$project", bson.M{
						"foo_DOT_a+sum(foo_DOT_b)": bson.M{"$add": []interface{}{"$foo_DOT_a", "$sum(foo_DOT_b)"}},
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

			test("select a + c + sum(b) from foo group by c",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"foo_DOT_a": bson.M{
							"$first": "$a",
						},
						"sum(foo_DOT_b)": bson.M{
							"$sum": "$b",
						},
					}}},
					{{"$project", bson.M{
						"_id":            0,
						"foo_DOT_a":      "$foo_DOT_a",
						"foo_DOT_c":      "$_id.foo_DOT_c",
						"sum(foo_DOT_b)": "$sum(foo_DOT_b)",
					}}},
					{{"$project", bson.M{
						"foo_DOT_a+foo_DOT_c+sum(foo_DOT_b)": bson.M{"$add": []interface{}{bson.M{"$add": []interface{}{"$foo_DOT_a", "$foo_DOT_c"}}, "$sum(foo_DOT_b)"}},
					}}},
				},
			)

			test("select a + sum(distinct b) from foo group by c",
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
						"sum(distinct foo_DOT_b)": bson.M{"$sum": "$distinct foo_DOT_b"},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a+sum(distinct foo_DOT_b)": bson.M{"$add": []interface{}{"$foo_DOT_a", "$sum(distinct foo_DOT_b)"}},
					}}},
				},
			)

			test("select c + sum(distinct b) from foo group by c",
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
						"sum(distinct foo_DOT_b)": bson.M{"$sum": "$distinct foo_DOT_b"},
					}}},
					{{"$project", bson.M{
						"foo_DOT_c+sum(distinct foo_DOT_b)": bson.M{"$add": []interface{}{"$foo_DOT_c", "$sum(distinct foo_DOT_b)"}},
					}}},
				},
			)

			test("select sum(distinct a + b) from foo group by c",
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
						"sum(distinct foo_DOT_a+foo_DOT_b)": bson.M{"$sum": "$distinct foo_DOT_a+foo_DOT_b"},
					}}},
					{{"$project", bson.M{
						"sum(distinct foo_DOT_a+foo_DOT_b)": "$sum(distinct foo_DOT_a+foo_DOT_b)",
					}}},
				},
			)

			test("select a + sum(distinct a + b) from foo group by c",
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
						"sum(distinct foo_DOT_a+foo_DOT_b)": bson.M{"$sum": "$distinct foo_DOT_a+foo_DOT_b"},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a+sum(distinct foo_DOT_a+foo_DOT_b)": bson.M{"$add": []interface{}{"$foo_DOT_a", "$sum(distinct foo_DOT_a+foo_DOT_b)"}},
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
						"_id":      0,
						"count(*)": "$count(*)",
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
						"_id":              0,
						"count(foo_DOT_a)": "$count(foo_DOT_a)",
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
			test("select sum(a) from foo group by c having sum(b) = 10",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_c", "$c"},
						},
						"sum(foo_DOT_a)": bson.M{
							"$sum": "$a",
						},
						"sum(foo_DOT_b)": bson.M{
							"$sum": "$b",
						},
					}}},
					{{"$project", bson.M{
						"_id":            0,
						"sum(foo_DOT_a)": "$sum(foo_DOT_a)",
						"sum(foo_DOT_b)": "$sum(foo_DOT_b)",
					}}},
					{{"$match", bson.M{
						"sum(foo_DOT_b)": int64(10),
					}}},
					{{"$project", bson.M{
						"sum(foo_DOT_a)": "$sum(foo_DOT_a)",
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

			test("select a from foo group by a order by sum(b)",
				[]bson.D{
					{{"$group", bson.M{
						"_id": bson.D{
							{"foo_DOT_a", "$a"},
						},
						"sum(foo_DOT_b)": bson.M{
							"$sum": "$b",
						},
					}}},
					{{"$project", bson.M{
						"_id":            0,
						"foo_DOT_a":      "$_id.foo_DOT_a",
						"sum(foo_DOT_b)": "$sum(foo_DOT_b)",
					}}},
					{{"$sort", bson.D{
						{"sum(foo_DOT_b)", 1},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a": "$foo_DOT_a",
					}}},
				},
			)

			Convey("no push down", func() {
				// TODO: parser issue
				// test("select a from foo order by a > b")
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
		})
	})
}
