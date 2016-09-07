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
	testSchema, err := schema.New(testSchema4)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}
	defaultDbName := "test"

	test := func(sql string, expected ...[]bson.D) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			selectStatement := statement.(parser.SelectStatement)
			plan, err := AlgebrizeSelect(selectStatement, defaultDbName, testSchema)
			So(err, ShouldBeNil)
			actualPlan, err := OptimizePlan(createTestConnectionCtx(), plan)
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

					test("select foo.a, bar.b from foo inner join bar on foo.a = bar.a AND bar.b > 10",
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

					test("select foo.c, bar.a, baz.b from foo inner join bar on foo.a = bar.a inner join baz on bar.a = baz.a",
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
							{{"$project", bson.M{
								"foo_DOT_c": "$c",
								"bar_DOT_a": "$__joined_bar.a",
								"baz_DOT_b": "$__joined_baz.b",
							}}},
						},
					)

					test("select foo.a, bar.a, baz.a from foo inner join bar on foo.a = bar.a inner join baz on bar.a = baz.a",
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
							{{"$unwind", bson.M{
								"path": "$__joined_bar",
								"preserveNullAndEmptyArrays": true,
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
								"__joined_bar": bson.M{"$cond": bson.M{
									"if": bson.M{"$cond": []interface{}{
										bson.M{"$or": []interface{}{
											bson.M{"$eq": []interface{}{
												bson.M{"$ifNull": []interface{}{"$b", nil}},
												nil,
											}},
											bson.M{"$eq": []interface{}{
												bson.M{"$ifNull": []interface{}{bson.M{"$literal": SQLInt(10)}, nil}},
												nil,
											}}}},
										nil,
										bson.M{"$gt": []interface{}{"$b", bson.M{"$literal": SQLInt(10)}}},
									}},
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
								"_id":    1,
								"a":      1,
								"b":      1,
								"c":      1,
								"d.e":    1,
								"d.f":    1,
								"filter": 1,
								"g":      1,
								"__joined_bar": bson.M{"$cond": bson.M{
									"if": bson.M{"$cond": []interface{}{
										bson.M{"$or": []interface{}{
											bson.M{"$eq": []interface{}{
												bson.M{"$ifNull": []interface{}{"$__joined_bar.b", nil}},
												nil,
											}},
											bson.M{"$eq": []interface{}{
												bson.M{"$ifNull": []interface{}{bson.M{"$literal": SQLInt(10)}, nil}},
												nil,
											}}}},
										nil,
										bson.M{"$gt": []interface{}{"$__joined_bar.b", bson.M{"$literal": SQLInt(10)}}},
									}},
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

				test("select exists(select a from bar) from foo",
					[]bson.D{
						{{"$project", bson.M{
							"bar_DOT_a": "$a",
						}}},
					})
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
					{{"$project", bson.M{
						"foo_DOT_a": "$a",
						"foo_DOT_b": "$b",
						"foo_DOT_c": "$c",
					}}},
				},
			)

			test("select a from foo where b < c AND a = 10",
				[]bson.D{
					{{"$match", bson.M{
						"a": int64(10),
					}}},
					{{"$project", bson.M{
						"foo_DOT_a": "$a",
						"foo_DOT_b": "$b",
						"foo_DOT_c": "$c",
					}}},
				},
			)

			Convey("no push down, project columns", func() {
				test("select a from foo where b < c",
					[]bson.D{
						{{"$project", bson.M{
							"foo_DOT_a": "$a",
							"foo_DOT_b": "$b",
							"foo_DOT_c": "$c",
						}}},
					},
				)
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
						"_id":            0,
						"max(foo_DOT_a)": "$max(foo_DOT_a)",
						"max(foo_DOT_b)": "$max(foo_DOT_b)",
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
						"_id":            0,
						"foo_DOT_c":      "$_id.foo_DOT_c",
						"max(foo_DOT_a)": "$max(foo_DOT_a)",
						"max(foo_DOT_b)": "$max(foo_DOT_b)",
					}}},
					{{"$project", bson.M{
						"foo_DOT_c":      "$foo_DOT_c",
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
						"_id":            0,
						"foo_DOT_a":      "$foo_DOT_a",
						"max(foo_DOT_b)": "$max(foo_DOT_b)",
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
						"_id":            0,
						"foo_DOT_a":      "$foo_DOT_a",
						"max(foo_DOT_b)": "$max(foo_DOT_b)",
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
						"_id":            0,
						"foo_DOT_a":      "$foo_DOT_a",
						"foo_DOT_c":      "$_id.foo_DOT_c",
						"max(foo_DOT_b)": "$max(foo_DOT_b)",
					}}},
					{{"$project", bson.M{
						"foo_DOT_a+foo_DOT_c+max(foo_DOT_b)": bson.M{"$add": []interface{}{bson.M{"$add": []interface{}{"$foo_DOT_a", "$foo_DOT_c"}}, "$max(foo_DOT_b)"}},
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
					{{"$project", bson.M{
						"_id":            0,
						"max(foo_DOT_a)": "$max(foo_DOT_a)",
						"max(foo_DOT_b)": "$max(foo_DOT_b)",
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
					{{"$project", bson.M{
						"_id":            0,
						"foo_DOT_a":      "$_id.foo_DOT_a",
						"max(foo_DOT_b)": "$max(foo_DOT_b)",
					}}},
					{{"$sort", bson.D{
						{"max(foo_DOT_b)", 1},
					}}},
					{{"$project", bson.M{
						"foo_DOT_a": "$foo_DOT_a",
					}}},
				},
			)

			Convey("no push down, project columns", func() {
				test("select a from foo order by a > b",
					[]bson.D{
						{{"$project", bson.M{
							"foo_DOT_a": "$a",
							"foo_DOT_b": "$b",
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
	})
}

func TestOptimizeCommand(t *testing.T) {
	testSchema, err := schema.New(testSchema1)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}
	defaultDbName := "test"

	test := func(sql string, expected ...[]bson.D) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			setStatement := statement.(*parser.Set)
			set, err := AlgebrizeCommand(setStatement, defaultDbName, testSchema)
			So(err, ShouldBeNil)
			actualSet, err := OptimizeCommand(createTestConnectionCtx(), set)
			So(err, ShouldBeNil)

			pg := &pipelineGatherer{}
			pg.visit(actualSet)

			actual := pg.pipelines

			So(actual, ShouldResembleDiffed, expected)
		})
	}

	Convey("Subject: OptimizeSet", t, func() {
		test("set @t1 = (select a from foo limit 1)",
			[]bson.D{
				{{"$limit", int64(1)}},
				{{"$project", bson.M{
					"foo_DOT_a": "$a",
				}}},
			},
		)
	})
}

func TestOptimizeEvaluations(t *testing.T) {

	type test struct {
		sql      string
		expected string
		result   SQLExpr
	}

	runTests := func(tests []test) {
		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be optimized to %q", t.sql, t.expected), func() {
				e, err := getSQLExpr(schema, dbOne, tableTwoName, t.sql)
				So(err, ShouldBeNil)
				result, err := optimizeEvaluations(createTestEvalCtx(), e)
				So(err, ShouldBeNil)
				So(result, ShouldResemble, t.result)
			})
		}
	}

	Convey("Subject: optimizeEvaluations", t, func() {

		tests := []test{
			test{"3 = a", "a = 3", &SQLEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"3 < a", "a > 3", &SQLGreaterThanExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"3 <= a", "a >= 3", &SQLGreaterThanOrEqualExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"3 > a", "a < 3", &SQLLessThanExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"3 >= a", "a <= 3", &SQLLessThanOrEqualExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"3 <> a", "a <> 3", &SQLNotEqualsExpr{NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt), SQLInt(3)}},
			test{"3 + 3 = 6", "true", SQLTrue},
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
			test{"concat(NULL, a)", "null", SQLNull},
			test{"concat(a, NULL)", "null", SQLNull},
			test{"concat('go', 'lang')", "golang", SQLVarchar("golang")},
			test{"concat_ws(NULL, a)", "null", SQLNull},
			test{"convert(NULL, SIGNED)", "null", SQLNull},
			test{"exp(NULL)", "null", SQLNull},
			test{"exp(2)", "7.38905609893065", SQLFloat(7.38905609893065)},
			test{"greatest(a, NULL)", "null", SQLNull},
			test{"greatest(2, 3)", "3", SQLInt(3)},
			test{"ifnull(NULL, a)", "bar.a", NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt)},
			test{"ifnull(10, a)", "10", SQLInt(10)},
			test{"least(a, NULL)", "null", SQLNull},
			test{"least(2, 3)", "3", SQLInt(2)},
			test{"locate('bar', 'foobar', NULL)", "null", SQLNull},
			test{"locate('bar', 'foobar')", "4", SQLInt(4)},
			test{"makedate(2000, NULL)", "null", SQLNull},
			test{"makedate(NULL, 10)", "null", SQLNull},
			test{"mod(10, 2)", "0", SQLFloat(0)},
			test{"mod(NULL, 2)", "null", SQLNull},
			test{"mod(10, NULL)", "null", SQLNull},
			test{"nullif(NULL, a)", "null", SQLNull},
			test{"nullif(a, NULL)", "bar.a", NewSQLColumnExpr(1, "bar", "a", schema.SQLInt, schema.MongoInt)},
			test{"pow(a, NULL)", "null", SQLNull},
			test{"pow(NULL, a)", "null", SQLNull},
			test{"pow(2,2)", "4", SQLFloat(4)},
			test{"round(NULL)", "null", SQLNull},
			test{"round(NULL, 2)", "null", SQLNull},
			test{"round(2, NULL)", "null", SQLNull},
			test{"round(2, 2)", "2", SQLFloat(2)},
			test{"substring(NULL, 2)", "null", SQLNull},
			test{"substring(NULL, 2, 3)", "null", SQLNull},
			test{"substring('foobar', NULL)", "null", SQLNull},
			test{"substring('foobar', NULL, 2)", "null", SQLNull},
			test{"substring('foobar', 2, NULL)", "null", SQLNull},
			test{"substring('foobar', 2, 3)", "oob", SQLVarchar("oob")},
		}

		runTests(tests)
	})
}
