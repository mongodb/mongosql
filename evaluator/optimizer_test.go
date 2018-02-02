package evaluator_test

import (
	"fmt"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/kr/pretty"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"
)

const (
	emptyFieldNamePrefix = "__empty"
)

// Fully pushed-down queries are covered by TestPushdownPlan in
// optimizer_pushdown_test.go. This test covers the remaining cases, testing
// how we push down queries that are only partially pushed down on certain
// MongoDB versions.
func TestOptimizePartialPushdown(t *testing.T) {

	type test struct {
		name     string
		sql      string
		expected [][]bson.D
		versions []string
	}

	tests := []test{

		test{
			name:     "huge_limit",
			sql:      "select a from foo limit 18446744073709551614",
			expected: [][]bson.D{},
		},
		test{
			name:     "inner_joins_subqueries_nested",
			versions: []string{"3.2", "3.4"},
			sql:      "select * from (select foo.a from bar join (select foo.a from foo) foo on foo.a=bar.b) x join (select g.a from bar join (select foo.a from foo) g on g.a=bar.a) y on x.a=y.a",
			expected: [][]bson.D{
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
					{{"$match", bson.M{"test_DOT_foo_DOT_a": bson.M{"$ne": nil}}}},
					{{"$lookup", bson.M{
						"from":         "bar",
						"localField":   "test_DOT_foo_DOT_a",
						"foreignField": "b",
						"as":           "__joined_bar",
					}}},
					{{"$unwind", bson.M{
						"preserveNullAndEmptyArrays": false,
						"path": "$__joined_bar",
					}}},
					{{"$project", bson.M{
						"test_DOT_foo_DOT_a": "$test_DOT_foo_DOT_a",
					}}},
					{{"$project", bson.M{
						"test_DOT_x_DOT_a": "$test_DOT_foo_DOT_a",
					}}},
				},
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
					{{"$match", bson.M{"test_DOT_foo_DOT_a": bson.M{"$ne": nil}}}},
					{{"$lookup", bson.M{
						"from":         "bar",
						"localField":   "test_DOT_foo_DOT_a",
						"foreignField": "a",
						"as":           "__joined_bar",
					}}},
					{{"$unwind", bson.M{
						"preserveNullAndEmptyArrays": false,
						"path": "$__joined_bar",
					}}},
					{{"$project", bson.M{
						"test_DOT_g_DOT_a": "$test_DOT_foo_DOT_a",
					}}},
					{{"$project", bson.M{
						"test_DOT_y_DOT_a": "$test_DOT_g_DOT_a",
					}}},
				},
			}},

		test{
			name:     "left_join_inner_join_subqueries_nested",
			versions: []string{"3.2", "3.4"},
			sql:      "select * from foo f left join (select b.b from foo f join (select * from bar) b on f.a=b.a)  b on f.a=b.b",
			expected: [][]bson.D{
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_f_DOT__id": "$_id",
						"test_DOT_f_DOT_a":   "$a",
						"test_DOT_f_DOT_b":   "$b",
						"test_DOT_f_DOT_c":   "$c",
						"test_DOT_f_DOT_e":   "$d.e",
						"test_DOT_f_DOT_f":   "$d.f",
						"test_DOT_f_DOT_g":   "$g",
					}}},
				},
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_bar_DOT__id": "$_id",
						"test_DOT_bar_DOT_a":   "$a",
						"test_DOT_bar_DOT_b":   "$b",
					}}},
					{{"$match", bson.M{"test_DOT_bar_DOT_a": bson.M{"$ne": nil}}}},
					{{"$lookup", bson.M{
						"from":         "foo",
						"localField":   "test_DOT_bar_DOT_a",
						"foreignField": "a",
						"as":           "__joined_f",
					}}},
					{{"$unwind", bson.M{
						"path": "$__joined_f",
						"preserveNullAndEmptyArrays": false,
					}}},
					{{"$project", bson.M{
						"test_DOT_b_DOT_b": "$test_DOT_bar_DOT_b",
					}}},
					{{"$project", bson.M{
						"test_DOT_b_DOT_b": "$test_DOT_b_DOT_b",
					}}},
				},
			}},

		test{
			name: "join_nested_array_tables_0",
			sql:  "select * from foo f join merge m1 on f._id=m1._id join (select * from foo) g on g.a=f.a join merge_d_a m2 on m2._id=m1._id and m2._id=g.a",
			expected: [][]bson.D{
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
						"test_DOT_f_DOT__id":          "$__joined_f._id",
						"test_DOT_f_DOT_a":            "$__joined_f.a",
						"test_DOT_f_DOT_b":            "$__joined_f.b",
						"test_DOT_f_DOT_c":            "$__joined_f.c",
						"test_DOT_f_DOT_e":            "$__joined_f.d.e",
						"test_DOT_f_DOT_f":            "$__joined_f.d.f",
						"test_DOT_f_DOT_g":            "$__joined_f.g",
						"test_DOT_m1_DOT__id":         "$_id",
						"test_DOT_m1_DOT_a":           "$a",
						"test_DOT_m2_DOT__id":         "$_id",
						"test_DOT_m2_DOT_d_DOT_a":     "$d.a",
						"test_DOT_m2_DOT_d_DOT_a_idx": "$d.a_idx",
						"test_DOT_m2_DOT_d_idx":       "$d_idx",
					}}},
				},
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_foo_DOT__id": "$_id",
						"test_DOT_foo_DOT_a":   "$a",
						"test_DOT_foo_DOT_b":   "$b",
						"test_DOT_foo_DOT_c":   "$c",
						"test_DOT_foo_DOT_e":   "$d.e",
						"test_DOT_foo_DOT_f":   "$d.f",
						"test_DOT_foo_DOT_g":   "$g",
					}}},
					{{"$project", bson.M{
						"test_DOT_g_DOT__id": "$test_DOT_foo_DOT__id",
						"test_DOT_g_DOT_a":   "$test_DOT_foo_DOT_a",
						"test_DOT_g_DOT_b":   "$test_DOT_foo_DOT_b",
						"test_DOT_g_DOT_c":   "$test_DOT_foo_DOT_c",
						"test_DOT_g_DOT_e":   "$test_DOT_foo_DOT_e",
						"test_DOT_g_DOT_f":   "$test_DOT_foo_DOT_f",
						"test_DOT_g_DOT_g":   "$test_DOT_foo_DOT_g",
					}}},
				},
			}},

		test{
			name:     "join_subqueries_where_limit",
			versions: []string{"3.2", "3.4"},
			sql:      "select f.a from foo f join (select bar.a from bar) b on f.a=b.a join (select foo.a from foo where foo.a > 4 limit 1) c on b.a=c.a and f.a=c.a and f.b=b.a",
			expected: [][]bson.D{
				[]bson.D{
					{{"$match", bson.M{"a": bson.M{"$gt": int64(4)}}}},
					{{"$limit", int64(1)}},
					{{"$project", bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
					{{"$project", bson.M{
						"test_DOT_c_DOT_a": "$test_DOT_foo_DOT_a",
					}}},
				},
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_bar_DOT_a": "$a",
					}}},
					{{"$project", bson.M{
						"test_DOT_b_DOT_a": "$test_DOT_bar_DOT_a",
					}}},
				},
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_f_DOT_a": "$a",
						"test_DOT_f_DOT_b": "$b",
					}}},
				},
			}},
		test{
			name:     "right_non_equijoin",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo right join bar on foo.a < bar.a",
			expected: [][]bson.D{
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
				},
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_bar_DOT_a": "$a",
					}}},
				},
			},
		},

		test{
			name:     "self_join_0",
			versions: []string{"3.2", "3.4"},
			sql:      "select * from merge r left join merge_d_a a on r._id=a._id",
			expected: [][]bson.D{
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_r_DOT__id": "$_id",
						"test_DOT_r_DOT_a":   "$a",
					}}}},
				[]bson.D{
					{{"$unwind", bson.D{
						{"includeArrayIndex", "d_idx"},
						{"path", "$d"},
					}}},
					{{"$unwind", bson.D{
						{"includeArrayIndex", "d.a_idx"},
						{"path", "$d.a"},
					}}},
					{{"$project", bson.M{
						"test_DOT_a_DOT_d_idx":       "$d_idx",
						"test_DOT_a_DOT__id":         "$_id",
						"test_DOT_a_DOT_d_DOT_a":     "$d.a",
						"test_DOT_a_DOT_d_DOT_a_idx": "$d.a_idx",
					}}},
				}},
		},

		test{
			name:     "self_join_4",
			versions: []string{"3.4"},
			sql:      "select b._id, c._id from merge r left join merge_b b on r._id=b._id inner join merge_c c on r._id=c._id left join merge_d_a a on r._id=a._id",
			expected: [][]bson.D{
				[]bson.D{
					{{"$addFields", bson.M{
						"_id_0": bson.D{{"$cond", []interface{}{
							bson.D{{"$or", []interface{}{
								bson.D{{"$lte", []interface{}{"$b", interface{}(nil)}}},
								bson.D{{"$eq", []interface{}{"$b", []interface{}{}}}}}}}, interface{}(nil), "$_id"}}}}}},
					{{"$unwind", bson.D{{"includeArrayIndex", "b_idx"}, {"path", "$b"}, {"preserveNullAndEmptyArrays", true}}}},
					{{"$unwind", bson.D{{"includeArrayIndex", "c_idx"}, {"path", "$c"}}}},
					{{"$project", bson.M{"test_DOT_b_DOT__id": "$_id_0", "test_DOT_c_DOT__id": "$_id", "test_DOT_r_DOT__id": "$_id"}}}},
				[]bson.D{
					{{"$unwind", bson.D{{"includeArrayIndex", "d_idx"}, {"path", "$d"}}}},
					{{"$unwind", bson.D{{"includeArrayIndex", "d.a_idx"}, {"path", "$d.a"}}}},
					{{"$project", bson.M{"test_DOT_a_DOT__id": "$_id"}}}},
			},
		},

		test{
			name:     "non_equijoin_0",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo inner join bar on foo.a < bar.a",
			expected: [][]bson.D{
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
				},
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_bar_DOT_a": "$a",
					}}},
				},
			}},
		test{
			name:     "non_equijoin_2",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo, bar where foo.a < bar.a",
			expected: [][]bson.D{
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
				},
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_bar_DOT_a": "$a",
					}}},
				},
			}},
		test{
			name:     "non_equijoin_3",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo left join bar on foo.a < bar.a",
			expected: [][]bson.D{
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
				},
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_bar_DOT_a": "$a",
					}}},
				},
			}},
		test{
			name:     "non_equijoin_4",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo right join bar on foo.a < bar.a",
			expected: [][]bson.D{
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
				},
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_bar_DOT_a": "$a",
					}}},
				},
			}},
	}

	versionByStr := map[string][]uint8{
		"3.2": []uint8{3, 2, 0},
		"3.4": []uint8{3, 4, 0},
		"3.6": []uint8{3, 6, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			versions := test.versions
			if len(versions) == 0 {
				versions = []string{"3.2", "3.4", "3.6"}
			}

			for _, version := range versions {
				t.Run(version, func(t *testing.T) {
					req := require.New(t)

					testSchema, err := schema.New(optimizerTestSchema)
					req.Nil(err, "failed to load schema")

					testInfo := evaluator.GetMongoDBInfo(versionByStr[version], testSchema, mongodb.AllPrivileges)
					testVariables := evaluator.CreateTestVariables(testInfo)
					testCatalog := evaluator.GetCatalogFromSchema(testSchema, testVariables)
					defaultDbName := "test"

					statement, err := parser.Parse(test.sql)
					req.Nil(err, "failed to parse statement")

					plan, err := evaluator.AlgebrizeQuery(statement, defaultDbName, testVariables, testCatalog)
					req.Nil(err, "failed to algebrize query")

					actualPlan := evaluator.OptimizePlan(createTestConnectionCtx(testInfo, versionByStr[version]...), plan)
					actual := evaluator.GetNodePipeline(actualPlan)

					req.Equalf(len(test.expected), len(actual),
						"expected %d pipelines in query plan, found %d\nexpected pipelines: %#v\nactual pipelines: %#v\nactual plan:\n%s",
						len(test.expected), len(actual), test.expected, actual, evaluator.PrettyPrintPlan(actualPlan))

					diff := ShouldResembleDiffed(actual, test.expected)
					req.Emptyf(diff, "expected pipeline diff to be empty\nexpected: %#v\nactual: %#v\n", test.expected, actual)

				})
			}
		})
	}

}

var optimizerTestSchema = []byte(`
schema:
-
  db: test
  tables:
  -
     table: datetest
     collection: datetest
     columns:
     -
        Name: dt
        MongoType: date
        SqlName: dt
        SqlType: date

  -
     table: foo
     collection: foo
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: int
        SqlType: int
     -
        Name: c
        MongoType: int
        SqlType: int
     -
        Name: d.e
        MongoType: int
        SqlName: e
        SqlType: int
     -
        Name: d.f
        MongoType: int
        SqlName: f
        SqlType: int
     -
        Name: g
        MongoType: bool
        SqlName: g
        SqlType: boolean
     -
        Name: _id
        MongoType: bson.ObjectId
        SqlType: varchar
     -
        Name: filter
        MongoType: mongo.Filter
        SqlName: filter
        SqlType: varchar
  -
     table: bar
     collection: bar
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: int
        SqlType: int
     -
        Name: _id
        MongoType: bson.ObjectId
        SqlType: varchar
  -
     table: baz
     collection: baz
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: int
        SqlType: int
     -
        Name: _id
        MongoType: bson.ObjectId
        SqlType: varchar
  -
    table: merge
    collection: merge
    pipeline: []
    columns:
    - Name: _id
      MongoType: bson.ObjectId
      SqlName: _id
      SqlType: varchar
    - Name: a
      MongoType: float64
      SqlName: a
      SqlType: float64
  - table: merge_b
    collection: merge
    pipeline:
    - $unwind:
        includeArrayIndex: b_idx
        path: $b
    columns:
    - Name: _id
      MongoType: bson.ObjectId
      SqlName: _id
      SqlType: varchar
    - Name: b
      MongoType: float64
      SqlName: b
      SqlType: float64
    - Name: b_idx
      MongoType: int
      SqlName: b_idx
      SqlType: int
  - table: merge_c
    collection: merge
    pipeline:
    - $unwind:
        includeArrayIndex: c_idx
        path: $c
    columns:
    - Name: _id
      MongoType: bson.ObjectId
      SqlName: _id
      SqlType: varchar
    - Name: c
      MongoType: float64
      SqlName: c
      SqlType: float64
    - Name: c_idx
      MongoType: int
      SqlName: c_idx
      SqlType: int
  - table: merge_d
    collection: merge
    pipeline:
    - $unwind:
        includeArrayIndex: d_idx
        path: $d
    columns:
    - Name: _id
      MongoType: bson.ObjectId
      SqlName: _id
      SqlType: varchar
    - Name: d_idx
      MongoType: int
      SqlName: d_idx
      SqlType: int
  - table: merge_d_a
    collection: merge
    pipeline:
    - $unwind:
        includeArrayIndex: d_idx
        path: $d
    - $unwind:
        includeArrayIndex: d.a_idx
        path: $d.a
    columns:
    - Name: _id
      MongoType: bson.ObjectId
      SqlName: _id
      SqlType: varchar
    - Name: d.a
      MongoType: float64
      SqlName: d.a
      SqlType: float64
    - Name: d.a_idx
      MongoType: int
      SqlName: d.a_idx
      SqlType: int
    - Name: d_idx
      MongoType: int
      SqlName: d_idx
      SqlType: int
`)

func TestPushdownSharding(t *testing.T) {
	testSchema, err := schema.New(testSchema4)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}
	testInfo := getMongoDBInfoWithShardedCollection(nil, testSchema, mongodb.AllPrivileges, "foo")
	testVariables := evaluator.CreateTestVariables(testInfo)
	testCatalog := evaluator.GetCatalogFromSchema(testSchema, testVariables)
	defaultDbName := "test"
	test := func(sql string, expected ...[]bson.D) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			plan, err := evaluator.AlgebrizeQuery(statement, defaultDbName, testVariables, testCatalog)
			So(err, ShouldBeNil)
			actualPlan := evaluator.OptimizePlan(createTestConnectionCtx(testInfo), plan)

			actual := evaluator.GetNodePipeline(actualPlan)

			v := ShouldResembleDiffed(actual, expected)
			if v != "" {
				fmt.Printf("\n ACTUAL: %#v", pretty.Formatter(actual))
				fmt.Printf("\n EXPECTED: %#v", pretty.Formatter(expected))
			}
			So(actual, ShouldResembleDiffed, expected)
		})
	}

	Convey("Join behaviour against sharded collections", t, func() {
		// should not push down because the from collection is sharded.
		test("select * from bar left join foo on bar.a=foo.a and bar.a=foo.f",
			[]bson.D{
				{{"$project", bson.M{
					"test_DOT_bar_DOT_b":   "$b",
					"test_DOT_bar_DOT__id": "$_id",
					"test_DOT_bar_DOT_a":   "$a",
				}}}},
			[]bson.D{
				{{
					"$project", bson.M{
						"test_DOT_foo_DOT_a":   "$a",
						"test_DOT_foo_DOT_b":   "$b",
						"test_DOT_foo_DOT_c":   "$c",
						"test_DOT_foo_DOT_e":   "$d.e",
						"test_DOT_foo_DOT_g":   "$g",
						"test_DOT_foo_DOT_f":   "$d.f",
						"test_DOT_foo_DOT__id": "$_id",
					}}}},
		)
		// should push down because the from collection is not sharded after flipping.
		test("select * from bar right join foo on bar.a=foo.a and bar.a=foo.f",
			[]bson.D{
				{{"$lookup", bson.M{
					"from":         "bar",
					"localField":   "a",
					"foreignField": "a",
					"as":           "__joined_bar",
				}}},
				{{"$project", bson.M{
					"c":      1,
					"d.f":    1,
					"g":      1,
					"_id":    1,
					"filter": 1,
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
					"a":   1,
					"b":   1,
					"d.e": 1,
				}}},
				{{"$addFields", bson.M{"__joined_bar": bson.M{
					"$filter": bson.M{
						"cond": bson.M{
							"$let": bson.M{
								"vars": bson.M{
									"left": "$$this.a", "right": "$d.f"},
								"in": bson.M{
									"$cond": []interface{}{bson.M{
										"$or": []interface{}{bson.M{
											"$eq": []interface{}{bson.M{
												"$ifNull": []interface{}{
													"$$left", nil,
												}}, nil,
											},
										}, bson.M{
											"$eq": []interface{}{bson.M{
												"$ifNull": []interface{}{"$$right", nil}},
												nil}}}},
										nil,
										bson.M{
											"$eq": []interface{}{"$$left", "$$right"}}}}}},
						"input": "$__joined_bar", "as": "this"}}}}},
				{{"$unwind", bson.M{
					"path": "$__joined_bar",
					"preserveNullAndEmptyArrays": true,
				}}},
				{{"$project", bson.M{
					"test_DOT_bar_DOT_b":   "$__joined_bar.b",
					"test_DOT_foo_DOT_f":   "$d.f",
					"test_DOT_foo_DOT_c":   "$c",
					"test_DOT_foo_DOT_e":   "$d.e",
					"test_DOT_foo_DOT_g":   "$g",
					"test_DOT_foo_DOT__id": "$_id",
					"test_DOT_bar_DOT_a":   "$__joined_bar.a",
					"test_DOT_bar_DOT__id": "$__joined_bar._id",
					"test_DOT_foo_DOT_a":   "$a",
					"test_DOT_foo_DOT_b":   "$b",
				}}},
			})
		// after flipping, the from collection, foo is sharded and it should not push down.
		test("select * from foo right join bar on foo.a=bar.a and foo.f=bar.a",
			[]bson.D{
				{{
					"$project", bson.M{
						"test_DOT_foo_DOT_a":   "$a",
						"test_DOT_foo_DOT_b":   "$b",
						"test_DOT_foo_DOT_c":   "$c",
						"test_DOT_foo_DOT_e":   "$d.e",
						"test_DOT_foo_DOT_g":   "$g",
						"test_DOT_foo_DOT_f":   "$d.f",
						"test_DOT_foo_DOT__id": "$_id",
					}}}},
			[]bson.D{
				{{"$project", bson.M{
					"test_DOT_bar_DOT_b":   "$b",
					"test_DOT_bar_DOT__id": "$_id",
					"test_DOT_bar_DOT_a":   "$a",
				}}}})
		// should flip after not being able to be pushed down the first time due to foo being sharded and then
		// push down.
		test("select * from bar inner join foo on bar.a=foo.a and bar.a=foo.f",
			[]bson.D{
				{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
				{{"$lookup", bson.M{
					"from":         "bar",
					"localField":   "a",
					"foreignField": "a",
					"as":           "__joined_bar"}}},
				{{"$unwind", bson.M{
					"path": "$__joined_bar",
					"preserveNullAndEmptyArrays": false}}},
				{{"$addFields", bson.M{
					"__predicate": bson.D{
						{"$let", bson.D{
							{"vars", bson.M{
								"predicate": bson.M{
									"$let": bson.M{
										"vars": bson.M{
											"right": "$d.f",
											"left":  "$__joined_bar.a",
										},
										"in": bson.M{
											"$cond": []interface{}{
												bson.M{
													"$or": []interface{}{
														bson.M{
															"$eq": []interface{}{
																bson.M{
																	"$ifNull": []interface{}{
																		"$$left",
																		nil,
																	},
																},
																nil,
															},
														},
														bson.M{
															"$eq": []interface{}{bson.M{
																"$ifNull": []interface{}{
																	"$$right",
																	nil,
																},
															},
																nil,
															},
														},
													},
												},
												nil,
												bson.M{"$eq": []interface{}{
													"$$left",
													"$$right",
												},
												},
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
							}}}}}}}},
				{{"$match", bson.M{"__predicate": true}}},
				{{"$project", bson.M{
					"test_DOT_bar_DOT_a":   "$__joined_bar.a",
					"test_DOT_foo_DOT_c":   "$c",
					"test_DOT_foo_DOT_g":   "$g",
					"test_DOT_bar_DOT_b":   "$__joined_bar.b",
					"test_DOT_bar_DOT__id": "$__joined_bar._id",
					"test_DOT_foo_DOT_a":   "$a",
					"test_DOT_foo_DOT_b":   "$b",
					"test_DOT_foo_DOT_e":   "$d.e",
					"test_DOT_foo_DOT_f":   "$d.f",
					"test_DOT_foo_DOT__id": "$_id"}}},
			})
	})
}

func TestOptimizeSubqueryPlan(t *testing.T) {
	testSchema, err := schema.New(testSchema4)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}
	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)
	testVariables := evaluator.CreateTestVariables(testInfo)
	testCatalog := evaluator.GetCatalogFromSchema(testSchema, testVariables)
	defaultDbName := "test"

	testOptimize := func(sql string, expected ...[]bson.D) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			plan, err := evaluator.AlgebrizeQuery(statement, defaultDbName, testVariables, testCatalog)
			So(err, ShouldBeNil)
			ctx := createTestConnectionCtx(testInfo)
			optimized, err := evaluator.OptimizeSubqueries(ctx, ctx.Logger(""), plan, false)
			So(err, ShouldBeNil)

			subqueryPlan := evaluator.GetSubqueryPlan(optimized)

			actual := evaluator.GetNodePipeline(subqueryPlan)

			So(actual, ShouldResembleDiffed, expected)
		})
	}

	testExecute := func(sql string, data []bson.D) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			plan, err := evaluator.AlgebrizeQuery(statement, defaultDbName, testVariables, testCatalog)
			So(err, ShouldBeNil)

			//fmt.Printf("\n%+v\n", PrettyPrintPlan(plan))

			sourceReplacer := &evaluator.SourceStageReplacer{Data: data}
			replaced, err := sourceReplacer.VisitStage(plan)
			So(err, ShouldBeNil)
			So(sourceReplacer.Existing, ShouldEqual, 0)

			//fmt.Printf("\n%+v\n", PrettyPrintPlan(replaced.(PlanStage)))

			ctx := createTestConnectionCtx(testInfo)
			optimized, err := evaluator.OptimizeSubqueries(ctx, ctx.Logger(""), replaced, true)
			So(err, ShouldBeNil)

			sourceReplacer = &evaluator.SourceStageReplacer{}
			sourceReplacer.VisitStage(optimized)
			So(sourceReplacer.Existing, ShouldEqual, 1)
			So(sourceReplacer.Replaced, ShouldEqual, 0)
		})
	}

	testCache := func(sql string, data []bson.D) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			plan, err := evaluator.AlgebrizeQuery(statement, defaultDbName, testVariables, testCatalog)
			So(err, ShouldBeNil)

			sourceReplacer := &evaluator.SourceStageReplacer{Data: data}
			replaced, err := sourceReplacer.VisitStage(plan)
			So(err, ShouldBeNil)
			So(sourceReplacer.Existing, ShouldEqual, 0)

			ctx := createTestConnectionCtx(testInfo)

			optimized, err := evaluator.OptimizeSubqueries(ctx, ctx.Logger(""), replaced, true)
			So(err, ShouldBeNil)

			So(evaluator.GetCacheStateCount(optimized), ShouldEqual, 1)

		})
	}

	Convey("Subject: OptimizeSubqueryPlan", t, func() {
		Convey("subquery optimization", func() {
			testOptimize("select a, (select b from bar) from foo",
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_bar_DOT_b": "$b",
					}}},
				})
			testOptimize("select exists(select a from bar) from foo",
				[]bson.D{
					{{"$project", bson.M{
						"test_DOT_bar_DOT_a": "$a",
					}}},
				})
			testOptimize("select a from bar where `a` = (select `b` from bar where b=2)",
				[]bson.D{
					{{"$match", bson.M{
						"b": int64(2),
					}}},
					{{"$project", bson.M{
						"test_DOT_bar_DOT_b": "$b",
					}}},
				})
			testOptimize("select a from bar where `a` = (select `b` from bar where b = (select a from bar where a=1))",
				[]bson.D{
					bson.D{{"$project", bson.M{
						"test_DOT_bar_DOT_b": "$b",
					}}},
					bson.D{{"$project", bson.M{
						"test_DOT_bar_DOT_b": "$test_DOT_bar_DOT_b",
					}}},
				},
				[]bson.D{
					bson.D{{"$match", bson.M{
						"a": int64(1),
					}}},
					bson.D{{"$project", bson.M{
						"test_DOT_bar_DOT_a": "$a",
					}}},
				})
			testOptimize("select a from bar where (`a`, `b`) = (select `c`, `b` from foo where b=2)",
				[]bson.D{
					{{"$match", bson.M{
						"b": int64(2),
					}}},
					{{"$project", bson.M{
						"test_DOT_foo_DOT_c": "$c",
						"test_DOT_foo_DOT_b": "$b",
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
		Convey("subquery execution and caching", func() {
			testCache("select a from foo where a in (select b from bar)",
				[]bson.D{
					{{"a", 1}},
					{{"b", 1}},
				})
			testCache("select a from foo where a not in (select b from bar)",
				[]bson.D{
					{{"a", 1}},
					{{"b", 1}},
				})
			testCache("select a from foo where a < all (select b from bar)",
				[]bson.D{
					{{"a", 1}},
					{{"b", 1}},
				})
			testCache("select a from foo where a >= some (select b from bar)",
				[]bson.D{
					{{"a", 1}},
					{{"b", 1}},
				})
			testCache("select a from foo where a < any (select b from bar)",
				[]bson.D{
					{{"a", 1}},
					{{"b", 1}},
				})

			testCache("select a from foo where (`a`, `c`) in (select `a`, `b` from bar)",
				[]bson.D{
					{{"a", 1}, {"c", 2}},
					{{"a", 1}, {"b", 2}},
				})
			testCache("select a from foo where (`a`, `c`) not in (select `a`, `b` from bar)",
				[]bson.D{
					{{"a", 1}, {"c", 2}},
					{{"a", 1}, {"b", 3}},
				})
		})
	})
}

func TestOptimizeEvaluations(t *testing.T) {

	type test struct {
		sql      string
		expected string
		result   evaluator.SQLExpr
	}

	testSchema, err := schema.New(testSchema4)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}

	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	runTests := func(tests []test) {
		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should be optimized to %q", t.sql, t.expected), func() {
				e, err := evaluator.GetSQLExpr(schema, dbOne, tableTwoName, t.sql)
				So(err, ShouldBeNil)

				ctx := createTestEvalCtx(testInfo)
				result, err := evaluator.OptimizeEvaluations(e, ctx, ctx.Logger(""))
				So(err, ShouldBeNil)
				So(result, ShouldResemble, t.result)
			})
		}
	}

	Convey("Subject: OptimizeEvaluations", t, func() {

		tests := []test{
			test{"3 / '3'", "1", evaluator.SQLFloat(1)},
			test{"3 * '3'", "9", evaluator.SQLInt(9)},
			test{"3 + '3'", "6", evaluator.SQLInt(6)},
			test{"3 - '3'", "0", evaluator.SQLInt(0)},
			test{"3 div '3'", "1", evaluator.SQLInt(1)},
			test{"3 = '3'", "true", evaluator.SQLTrue},
			test{"3 <= '3'", "true", evaluator.SQLTrue},
			test{"3 >= '3'", "true", evaluator.SQLTrue},
			test{"3 < '3'", "false", evaluator.SQLFalse},
			test{"3 > '3'", "false", evaluator.SQLFalse},
			test{"3 <=> '3'", "true", evaluator.SQLTrue},
			test{"3 = a", "a = 3", evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			test{"3 < a", "a > 3", evaluator.NewSQLGreaterThanExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			test{"3 <= a", "a >= 3", evaluator.NewSQLGreaterThanOrEqualExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			test{"3 > a", "a < 3", evaluator.NewSQLLessThanExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			test{"3 >= a", "a <= 3", evaluator.NewSQLLessThanOrEqualExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			test{"3 <> a", "a <> 3", evaluator.NewSQLNotEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			test{"3 + 3 = 6", "true", evaluator.SQLTrue},
			test{"3 <=> 3", "true", evaluator.SQLTrue},
			test{"NULL <=> 3", "false", evaluator.SQLFalse},
			test{"3 <=> NULL", "false", evaluator.SQLFalse},
			test{"NULL <=> NULL", "true", evaluator.SQLTrue},
			test{"3 / (3 - 2) = a", "a = 3", evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLFloat(3))},
			test{"3 + 3 = 6 AND 1 >= 1 AND 3 = a", "a = 3", evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			test{"3 / (3 - 2) = a AND 4 - 2 = b", "a = 3 AND b = 2",
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLFloat(3)),
					evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "b", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(2)))},
			test{"3 + 3 = 6 OR a = 3", "true", evaluator.SQLTrue},
			test{"3 + 3 = 5 OR a = 3", "a = 3", evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			test{"0 OR NULL", "null", evaluator.SQLNull},
			test{"1 OR NULL", "true", evaluator.SQLTrue},
			test{"NULL OR NULL", "null", evaluator.SQLNull},
			test{"0 AND 6+1 = 6", "false", evaluator.SQLFalse},
			test{"3 + 3 = 5 AND a = 3", "false", evaluator.SQLFalse},
			test{"0 AND NULL", "false", evaluator.SQLFalse},
			test{"1 AND NULL", "null", evaluator.SQLNull},
			test{"1 AND 6+0 = 6", "true", evaluator.SQLTrue},
			test{"3 + 3 = 6 AND a = 3", "a = 3", evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			test{"(3 + 3 = 5) XOR a = 3", "a = 3", evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			test{"(3 + 3 = 6) XOR a = 3", "a <> 3", evaluator.NewSQLNotExpr(evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3)))},
			test{"(13 + 9 > 6) XOR (a = 4)", "a <> 4", evaluator.NewSQLNotExpr(evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(4)))},
			test{"(8 / 5 = 9) XOR (a = 5)", "a = 5", evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(5))},
			test{"false XOR 23", "true", evaluator.SQLTrue},
			test{"true XOR 23", "false", evaluator.SQLFalse},
			test{"a = 23 XOR true", "a <> 23", evaluator.NewSQLNotExpr(evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(23)))},
			test{"!3", "0", evaluator.SQLFalse},
			test{"!NULL", "null", evaluator.SQLNull},
			test{"a = ~1", "a = 18446744073709551614", evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLUint64(18446744073709551614))},
			test{"a = ~2398238912332232323", "a = 16048505161377319292", evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLUint64(16048505161377319292))},
			test{"DAYNAME('2016-1-1')", "Friday", evaluator.SQLVarchar("Friday")},
			test{"(8-7)", "1", evaluator.SQLInt(1)},
			test{"a LIKE NULL", "null", evaluator.SQLNull},
			test{"4 LIKE NULL", "null", evaluator.SQLNull},
			test{"a = NULL", "null", evaluator.SQLNull},
			test{"a > NULL", "null", evaluator.SQLNull},
			test{"a >= NULL", "null", evaluator.SQLNull},
			test{"a < NULL", "null", evaluator.SQLNull},
			test{"a <= NULL", "null", evaluator.SQLNull},
			test{"a != NULL", "null", evaluator.SQLNull},
			test{"(1, 3) > (3, 4)", "SQLFalse", evaluator.SQLFalse},
			test{"(4, 3) > (3, 4)", "SQLTrue", evaluator.SQLTrue},
			test{"(4, 31) > (4, 4)", "SQLTrue", evaluator.SQLTrue},

			test{"abs(NULL)", "null", evaluator.SQLNull},
			test{"abs(-10)", "10", evaluator.SQLFloat(10)},
			test{"ascii(NULL)", "null", evaluator.SQLNull},
			test{"ascii('a')", "97", evaluator.SQLInt(97)},
			test{"char_length(NULL)", "null", evaluator.SQLNull},
			test{"character_length(NULL)", "null", evaluator.SQLNull},
			test{"concat(NULL, a)", "null", evaluator.SQLNull},
			test{"concat(a, NULL)", "null", evaluator.SQLNull},
			test{"concat('go', 'lang')", "golang", evaluator.SQLVarchar("golang")},
			test{"concat_ws(NULL, a)", "null", evaluator.SQLNull},
			test{"convert(NULL, SIGNED)", "null", evaluator.SQLNull},
			test{"elt(NULL, 'a', 'b')", "null", evaluator.SQLNull},
			test{"elt(4, 'a', 'b')", "null", evaluator.SQLNull},
			test{"exp(NULL)", "null", evaluator.SQLNull},
			test{"exp(2)", "7.38905609893065", evaluator.SQLFloat(7.38905609893065)},
			test{"greatest(a, NULL)", "null", evaluator.SQLNull},
			test{"greatest(2, 3)", "3", evaluator.SQLInt(3)},
			test{"ifnull(NULL, a)", "bar.a", evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt)},
			test{"ifnull(10, a)", "10", evaluator.SQLInt(10)},
			test{"interval(NULL, a)", "-1", evaluator.SQLInt(-1)},
			test{"interval(0, 1)", "0", evaluator.SQLInt(0)},
			test{"interval(1, 2, 3, 4)", "1", evaluator.SQLInt(0)},
			test{"interval(1, 1, 2, 3)", "1", evaluator.SQLInt(1)},
			test{"interval(-1, NULL, NULL, -0.5, 3, 4)", "1", evaluator.SQLInt(2)},
			test{"interval(-3.4, -4, -3.6, -3.4, -3, 1, 2)", "3", evaluator.SQLInt(3)},
			test{"interval(8, -4, 0, 7, 8)", "4", evaluator.SQLInt(4)},
			test{"interval(8, -3, 1, 7, 7)", "1", evaluator.SQLInt(4)},
			test{"interval(7.7, -3, 1, 7, 7)", "1", evaluator.SQLInt(4)},
			test{"least(a, NULL)", "null", evaluator.SQLNull},
			test{"least(2, 3)", "2", evaluator.SQLInt(2)},
			test{"locate('bar', 'foobar', NULL)", "0", evaluator.SQLInt(0)},
			test{"locate('bar', 'foobar')", "4", evaluator.SQLInt(4)},
			test{"makedate(2000, NULL)", "null", evaluator.SQLNull},
			test{"makedate(NULL, 10)", "null", evaluator.SQLNull},
			test{"mid('foobar', NULL, 2)", "null", evaluator.SQLNull},
			test{"mod(10, 2)", "0", evaluator.SQLFloat(0)},
			test{"mod(NULL, 2)", "null", evaluator.SQLNull},
			test{"mod(10, NULL)", "null", evaluator.SQLNull},
			test{"nullif(NULL, a)", "null", evaluator.SQLNull},
			test{"nullif(a, NULL)", "bar.a", evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt)},
			test{"pow(a, NULL)", "null", evaluator.SQLNull},
			test{"pow(NULL, a)", "null", evaluator.SQLNull},
			test{"pow(2,2)", "4", evaluator.SQLFloat(4)},
			test{"round(NULL, 2)", "null", evaluator.SQLNull},
			test{"round(2, NULL)", "null", evaluator.SQLNull},
			test{"round(2, 2)", "2", evaluator.SQLFloat(2)},
			test{"repeat('a', NULL)", "null", evaluator.SQLNull},
			test{"repeat(NULL, 3)", "null", evaluator.SQLNull},
			test{"substring(NULL, 2)", "null", evaluator.SQLNull},
			test{"substring(NULL, 2, 3)", "null", evaluator.SQLNull},
			test{"substring('foobar', NULL)", "null", evaluator.SQLNull},
			test{"substring('foobar', NULL, 2)", "null", evaluator.SQLNull},
			test{"substring('foobar', 2, NULL)", "null", evaluator.SQLNull},
			test{"substring('foobar', 2, 3)", "oob", evaluator.SQLVarchar("oob")},
			test{"substring_index(NULL, 'o', 0)", "", evaluator.SQLNull},
			test{"substring_index('foobar', 'o', 0)", "", evaluator.SQLVarchar("")},
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

	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	runTests := func(tests []test) {
		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)
		for _, t := range tests {
			Convey(fmt.Sprintf("%q should fail with error %q", t.sql, t.err), func() {
				e, err := evaluator.GetSQLExpr(schema, dbOne, tableTwoName, t.sql)
				So(err, ShouldBeNil)

				ctx := createTestEvalCtx(testInfo)
				_, err = evaluator.OptimizeEvaluations(e, ctx, ctx.Logger(""))
				So(err, ShouldResemble, t.err)
			})
		}
	}

	Convey("Subject: OptimizeEvaluations failures", t, func() {

		tests := []test{
			test{"pow(-2,2.2)", mysqlerrors.Defaultf(mysqlerrors.ER_DATA_OUT_OF_RANGE, "DOUBLE", "pow(-2,2.2)")},
			test{"pow(0,-2.2)", mysqlerrors.Defaultf(mysqlerrors.ER_DATA_OUT_OF_RANGE, "DOUBLE", "pow(0,-2.2)")},
			test{"pow(0,-5)", mysqlerrors.Defaultf(mysqlerrors.ER_DATA_OUT_OF_RANGE, "DOUBLE", "pow(0,-5)")},
		}

		runTests(tests)

	})
}
