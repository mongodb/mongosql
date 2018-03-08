package evaluator_test

import (
	"fmt"
	"sort"
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

// normalizeBSON replaces all instances of bson.M with bson.D internally, to make
// diffing easier in tests.
func normalizeBSON(input interface{}) interface{} {
	ret := input
	switch typed := input.(type) {
	case [][]bson.D:
		for i, docList := range typed {
			typed[i] = normalizeBSON(docList).([]bson.D)
		}
	case []bson.D:
		for i, doc := range typed {
			typed[i] = normalizeBSON(doc).(bson.D)
		}
	case []interface{}:
		for i, val := range typed {
			typed[i] = normalizeBSON(val)
		}
	case bson.D:
		for i, elem := range typed {
			typed[i] = normalizeBSON(elem).(bson.DocElem)
		}
		sort.Slice(typed, func(i, j int) bool {
			return typed[i].Name < typed[j].Name
		})
	case bson.M:
		out := make(bson.D, len(typed))
		i := 0
		for key := range typed {
			out[i] = bson.DocElem{Name: key, Value: normalizeBSON(typed[key])}
			i++
		}
		sort.Slice(out, func(i, j int) bool {
			return out[i].Name < out[j].Name
		})
		ret = out
	case bson.DocElem:
		typed.Value = normalizeBSON(typed.Value)
		ret = typed
	}
	return ret
}

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

		{
			name:     "count_star",
			sql:      "select count(*) from foo",
			expected: [][]bson.D{},
		},
		{
			name:     "count_star_with_order",
			sql:      "select count(*) from foo order by 1",
			expected: [][]bson.D{},
		},
		{
			name:     "huge_limit",
			sql:      "select a from foo limit 18446744073709551614",
			expected: [][]bson.D{},
		},
		{
			name:     "inner_joins_subqueries_nested",
			versions: []string{"3.2", "3.4"},
			sql: "select * from (select foo.a from bar join (select foo.a from foo) foo on" +
				" foo.a=bar.b) x join (select g.a from bar join (select foo.a from foo) g on " +
				"g.a=bar.a) y on x.a=y.a",
			expected: [][]bson.D{
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
					{{Name: "$match", Value: bson.M{"test_DOT_foo_DOT_a": bson.M{"$ne": nil}}}},
					{{Name: "$lookup", Value: bson.M{
						"from":         "bar",
						"localField":   "test_DOT_foo_DOT_a",
						"foreignField": "b",
						"as":           "__joined_bar",
					}}},
					{{Name: "$unwind", Value: bson.M{
						"preserveNullAndEmptyArrays": false,
						"path": "$__joined_bar",
					}}},
					{{Name: "$project", Value: bson.M{
						"test_DOT_foo_DOT_a": "$test_DOT_foo_DOT_a",
					}}},
					{{Name: "$project", Value: bson.M{
						"test_DOT_x_DOT_a": "$test_DOT_foo_DOT_a",
					}}},
				},
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
					{{Name: "$match", Value: bson.M{"test_DOT_foo_DOT_a": bson.M{"$ne": nil}}}},
					{{Name: "$lookup", Value: bson.M{
						"from":         "bar",
						"localField":   "test_DOT_foo_DOT_a",
						"foreignField": "a",
						"as":           "__joined_bar",
					}}},
					{{Name: "$unwind", Value: bson.M{
						"preserveNullAndEmptyArrays": false,
						"path": "$__joined_bar",
					}}},
					{{Name: "$project", Value: bson.M{
						"test_DOT_g_DOT_a": "$test_DOT_foo_DOT_a",
					}}},
					{{Name: "$project", Value: bson.M{
						"test_DOT_y_DOT_a": "$test_DOT_g_DOT_a",
					}}},
				},
			}},

		{
			name:     "left_join_inner_join_subqueries_nested",
			versions: []string{"3.2", "3.4"},
			sql: "select * from foo f left join (select b.b from foo f join (select * from " +
				"bar) b on f.a=b.a)  b on f.a=b.b",
			expected: [][]bson.D{
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_f_DOT__id": "$_id",
						"test_DOT_f_DOT_a":   "$a",
						"test_DOT_f_DOT_b":   "$b",
						"test_DOT_f_DOT_c":   "$c",
						"test_DOT_f_DOT_e":   "$d.e",
						"test_DOT_f_DOT_f":   "$d.f",
						"test_DOT_f_DOT_g":   "$g",
					}}},
				},
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_bar_DOT__id": "$_id",
						"test_DOT_bar_DOT_a":   "$a",
						"test_DOT_bar_DOT_b":   "$b",
					}}},
					{{Name: "$match", Value: bson.M{"test_DOT_bar_DOT_a": bson.M{"$ne": nil}}}},
					{{Name: "$lookup", Value: bson.M{
						"from":         "foo",
						"localField":   "test_DOT_bar_DOT_a",
						"foreignField": "a",
						"as":           "__joined_f",
					}}},
					{{Name: "$unwind", Value: bson.M{
						"path": "$__joined_f",
						"preserveNullAndEmptyArrays": false,
					}}},
					{{Name: "$project", Value: bson.M{
						"test_DOT_b_DOT_b": "$test_DOT_bar_DOT_b",
					}}},
					{{Name: "$project", Value: bson.M{
						"test_DOT_b_DOT_b": "$test_DOT_b_DOT_b",
					}}},
				},
			}},

		{
			name: "join_nested_array_tables_0",
			sql: "select * from foo f join merge m1 on f._id=m1._id join (select * from foo) g" +
				" on g.a=f.a join merge_d_a m2 on m2._id=m1._id and m2._id=g.a",
			expected: [][]bson.D{
				{
					{{Name: "$unwind", Value: bson.D{
						{Name: "includeArrayIndex", Value: "d_idx"},
						{Name: "path", Value: "$d"},
					}}},
					{{Name: "$unwind", Value: bson.D{
						{Name: "includeArrayIndex", Value: "d.a_idx"},
						{Name: "path", Value: "$d.a"},
					}}},
					{{Name: "$match", Value: bson.M{"_id": bson.M{"$ne": nil}}}},
					{{Name: "$lookup", Value: bson.M{
						"from":         "foo",
						"localField":   "_id",
						"foreignField": "_id",
						"as":           "__joined_f",
					}}},
					{{Name: "$unwind", Value: bson.M{
						"path": "$__joined_f",
						"preserveNullAndEmptyArrays": false,
					}}},
					{{Name: "$project", Value: bson.M{
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
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_foo_DOT__id": "$_id",
						"test_DOT_foo_DOT_a":   "$a",
						"test_DOT_foo_DOT_b":   "$b",
						"test_DOT_foo_DOT_c":   "$c",
						"test_DOT_foo_DOT_e":   "$d.e",
						"test_DOT_foo_DOT_f":   "$d.f",
						"test_DOT_foo_DOT_g":   "$g",
					}}},
					{{Name: "$project", Value: bson.M{
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

		{
			name:     "join_subqueries_where_limit",
			versions: []string{"3.2", "3.4"},
			sql: "select f.a from foo f join (select bar.a from bar) b on f.a=b.a join " +
				"(select foo.a from foo where foo.a > 4 limit 1) c on b.a=c.a and f.a=c.a and " +
				"f.b=b.a",
			expected: [][]bson.D{
				{
					{{Name: "$match", Value: bson.M{"a": bson.M{"$gt": int64(4)}}}},
					{{Name: "$limit", Value: int64(1)}},
					{{Name: "$project", Value: bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
					{{Name: "$project", Value: bson.M{
						"test_DOT_c_DOT_a": "$test_DOT_foo_DOT_a",
					}}},
				},
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_bar_DOT_a": "$a",
					}}},
					{{Name: "$project", Value: bson.M{
						"test_DOT_b_DOT_a": "$test_DOT_bar_DOT_a",
					}}},
				},
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_f_DOT_a": "$a",
						"test_DOT_f_DOT_b": "$b",
					}}},
				},
			}},
		{
			name:     "right_non_equijoin",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo right join bar on foo.a < bar.a",
			expected: [][]bson.D{
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
				},
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_bar_DOT_a": "$a",
					}}},
				},
			},
		},

		{
			name:     "self_join_0",
			versions: []string{"3.2", "3.4"},
			sql:      "select * from merge r left join merge_d_a a on r._id=a._id",
			expected: [][]bson.D{
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_r_DOT__id": "$_id",
						"test_DOT_r_DOT_a":   "$a",
					}}}},
				{
					{{Name: "$unwind", Value: bson.D{
						{Name: "includeArrayIndex", Value: "d_idx"},
						{Name: "path", Value: "$d"},
					}}},
					{{Name: "$unwind", Value: bson.D{
						{Name: "includeArrayIndex", Value: "d.a_idx"},
						{Name: "path", Value: "$d.a"},
					}}},
					{{Name: "$project", Value: bson.M{
						"test_DOT_a_DOT_d_idx":       "$d_idx",
						"test_DOT_a_DOT__id":         "$_id",
						"test_DOT_a_DOT_d_DOT_a":     "$d.a",
						"test_DOT_a_DOT_d_DOT_a_idx": "$d.a_idx",
					}}},
				}},
		},

		{
			name:     "self_join_4",
			versions: []string{"3.4"},
			sql: "select b._id, c._id from merge r left join merge_b b on r._id=b._id inner" +
				" join merge_c c on r._id=c._id left join merge_d_a a on r._id=a._id",
			expected: [][]bson.D{
				{
					{{Name: "$addFields", Value: bson.M{
						"_id_0": bson.D{{Name: "$cond", Value: []interface{}{
							bson.D{{Name: "$or", Value: []interface{}{
								bson.D{{Name: "$lte",
									Value: []interface{}{"$b", interface{}(nil)}}},
								bson.D{{Name: "$eq",
									Value: []interface{}{"$b", []interface{}{}}}}}}},
							interface{}(nil), "$_id"}}}}}},
					{{Name: "$unwind", Value: bson.D{{Name: "includeArrayIndex", Value: "b_idx"},
						{Name: "path", Value: "$b"}, {Name: "preserveNullAndEmptyArrays",
							Value: true}}}},
					{{Name: "$unwind", Value: bson.D{{Name: "includeArrayIndex", Value: "c_idx"},
						{Name: "path", Value: "$c"}}}},
					{{Name: "$project", Value: bson.M{"test_DOT_b_DOT__id": "$_id_0",
						"test_DOT_c_DOT__id": "$_id", "test_DOT_r_DOT__id": "$_id"}}}},
				{
					{{Name: "$unwind", Value: bson.D{{Name: "includeArrayIndex", Value: "d_idx"},
						{Name: "path", Value: "$d"}}}},
					{{Name: "$unwind", Value: bson.D{{Name: "includeArrayIndex", Value: "d.a_idx"},
						{Name: "path", Value: "$d.a"}}}},
					{{Name: "$project", Value: bson.M{"test_DOT_a_DOT__id": "$_id"}}}},
			},
		},

		{
			name:     "non_equijoin_0",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo inner join bar on foo.a < bar.a",
			expected: [][]bson.D{
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
				},
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_bar_DOT_a": "$a",
					}}},
				},
			}},
		{
			name:     "non_equijoin_2",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo, bar where foo.a < bar.a",
			expected: [][]bson.D{
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
				},
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_bar_DOT_a": "$a",
					}}},
				},
			}},
		{
			name:     "non_equijoin_3",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo left join bar on foo.a < bar.a",
			expected: [][]bson.D{
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
				},
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_bar_DOT_a": "$a",
					}}},
				},
			}},
		{
			name:     "non_equijoin_4",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo right join bar on foo.a < bar.a",
			expected: [][]bson.D{
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_foo_DOT_a": "$a",
					}}},
				},
				{
					{{Name: "$project", Value: bson.M{
						"test_DOT_bar_DOT_a": "$a",
					}}},
				},
			}},
	}

	versionByStr := map[string][]uint8{
		"3.2": {3, 2, 0},
		"3.4": {3, 4, 0},
		"3.6": {3, 6, 0},
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

					testSchema := evaluator.MustLoadSchema(optimizerTestSchema)

					testInfo := evaluator.GetMongoDBInfo(versionByStr[version], testSchema,
						mongodb.AllPrivileges)
					testVariables := evaluator.CreateTestVariables(testInfo)
					testCatalog := evaluator.GetCatalogFromSchema(testSchema, testVariables)
					defaultDbName := "test"

					statement, err := parser.Parse(test.sql)
					req.Nil(err, "failed to parse statement")

					plan, err := evaluator.AlgebrizeQuery(statement, defaultDbName, testVariables,
						testCatalog)
					req.Nil(err, "failed to algebrize query")

					actualPlan := evaluator.OptimizePlan(createTestConnectionCtx(testInfo,
						versionByStr[version]...), plan)
					actual := evaluator.GetNodePipeline(actualPlan)
					actual, expected := normalizeBSON(actual).([][]bson.D),
						normalizeBSON(test.expected).([][]bson.D)

					req.Equalf(len(expected), len(actual),
						"expected %d pipelines in query plan, found %d\nexpected pipelines: "+
							"%#v\nactual pipelines: %#v\nactual plan:\n%s",
						len(expected), len(actual), test.expected, actual,
						evaluator.PrettyPrintPlan(actualPlan))

					diff := ShouldResembleDiffed(actual, expected)
					req.Emptyf(diff, "expected pipeline diff to be empty\nexpected: %#v\nactual:"+
						" %#v\n", expected, actual)

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
	testSchema := evaluator.MustLoadSchema(testSchema4)
	testInfo := getMongoDBInfoWithShardedCollection(nil, testSchema, mongodb.AllPrivileges, "foo")
	testVariables := evaluator.CreateTestVariables(testInfo)
	testCatalog := evaluator.GetCatalogFromSchema(testSchema, testVariables)
	defaultDbName := "test"
	test := func(sql string, expected ...[]bson.D) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			plan, err := evaluator.AlgebrizeQuery(statement, defaultDbName, testVariables,
				testCatalog)
			So(err, ShouldBeNil)
			actualPlan := evaluator.OptimizePlan(createTestConnectionCtx(testInfo), plan)

			actual := evaluator.GetNodePipeline(actualPlan)

			actual, expected = normalizeBSON(actual).([][]bson.D),
				normalizeBSON(expected).([][]bson.D)

			v := ShouldResembleDiffed(actual, expected)
			if v != "" {
				fmt.Printf("\n SQL: %v", sql)
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
				{{Name: "$project", Value: bson.M{
					"test_DOT_bar_DOT_b":   "$b",
					"test_DOT_bar_DOT__id": "$_id",
					"test_DOT_bar_DOT_a":   "$a",
				}}}},
			[]bson.D{
				{{
					Name: "$project", Value: bson.M{
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
				{{Name: "$lookup", Value: bson.M{
					"from":         "bar",
					"localField":   "a",
					"foreignField": "a",
					"as":           "__joined_bar",
				}}},
				{{Name: "$project", Value: bson.M{
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
				{{Name: "$addFields", Value: bson.M{"__joined_bar": bson.M{
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
				{{Name: "$unwind", Value: bson.M{
					"path": "$__joined_bar",
					"preserveNullAndEmptyArrays": true,
				}}},
				{{Name: "$project", Value: bson.M{
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
					"_id":                  0,
				}}},
			})
		// after flipping, the from collection, foo is sharded and it should not push down.
		test("select * from foo right join bar on foo.a=bar.a and foo.f=bar.a",
			[]bson.D{
				{{
					Name: "$project", Value: bson.M{
						"test_DOT_foo_DOT_a":   "$a",
						"test_DOT_foo_DOT_b":   "$b",
						"test_DOT_foo_DOT_c":   "$c",
						"test_DOT_foo_DOT_e":   "$d.e",
						"test_DOT_foo_DOT_g":   "$g",
						"test_DOT_foo_DOT_f":   "$d.f",
						"test_DOT_foo_DOT__id": "$_id",
					}}}},
			[]bson.D{
				{{Name: "$project", Value: bson.M{
					"test_DOT_bar_DOT_b":   "$b",
					"test_DOT_bar_DOT__id": "$_id",
					"test_DOT_bar_DOT_a":   "$a",
				}}}})
		// should flip after not being able to be pushed down the first time due to foo being
		// sharded and then push down.
		test("select * from bar inner join foo on bar.a=foo.a and bar.a=foo.f",
			[]bson.D{
				{{Name: "$match", Value: bson.M{"a": bson.M{"$ne": nil}}}},
				{{Name: "$lookup", Value: bson.M{
					"from":         "bar",
					"localField":   "a",
					"foreignField": "a",
					"as":           "__joined_bar"}}},
				{{Name: "$unwind", Value: bson.M{
					"path": "$__joined_bar",
					"preserveNullAndEmptyArrays": false}}},
				{{Name: "$addFields", Value: bson.M{
					"__predicate": bson.D{
						{Name: "$let", Value: bson.D{
							{Name: "vars", Value: bson.M{
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
							{Name: "in", Value: bson.D{
								{Name: "$cond", Value: []interface{}{
									bson.D{{Name: "$or", Value: []interface{}{
										bson.D{{Name: "$eq", Value: []interface{}{"$$predicate",
											false}}},
										bson.D{{Name: "$eq", Value: []interface{}{"$$predicate",
											0}}},
										bson.D{{Name: "$eq", Value: []interface{}{"$$predicate",
											"0"}}},
										bson.D{{Name: "$eq", Value: []interface{}{"$$predicate",
											"-0"}}},
										bson.D{{Name: "$eq", Value: []interface{}{"$$predicate",
											"0.0"}}},
										bson.D{{Name: "$eq", Value: []interface{}{"$$predicate",
											"-0.0"}}},
										bson.D{{Name: "$eq", Value: []interface{}{"$$predicate",
											nil}}},
									}}},
									false,
									true,
								}},
							}}}}}}}},
				{{Name: "$match", Value: bson.M{"__predicate": true}}},
				{{Name: "$project", Value: bson.M{
					"test_DOT_bar_DOT_a":   "$__joined_bar.a",
					"test_DOT_foo_DOT_c":   "$c",
					"test_DOT_foo_DOT_g":   "$g",
					"test_DOT_bar_DOT_b":   "$__joined_bar.b",
					"test_DOT_bar_DOT__id": "$__joined_bar._id",
					"test_DOT_foo_DOT_a":   "$a",
					"test_DOT_foo_DOT_b":   "$b",
					"test_DOT_foo_DOT_e":   "$d.e",
					"test_DOT_foo_DOT_f":   "$d.f",
					"test_DOT_foo_DOT__id": "$_id",
					"_id": 0,
				}}},
			})
	})
}

func TestOptimizeSubqueryPlan(t *testing.T) {
	testSchema := evaluator.MustLoadSchema(testSchema4)
	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)
	testVariables := evaluator.CreateTestVariables(testInfo)
	testCatalog := evaluator.GetCatalogFromSchema(testSchema, testVariables)
	defaultDbName := "test"

	type test struct {
		sql      string
		expected [][]bson.D
	}

	testOptimize := func(t *testing.T, testCase test) {
		t.Run(testCase.sql, func(t *testing.T) {
			req := require.New(t)
			statement, err := parser.Parse(testCase.sql)
			req.Nil(err, "failed to parse query")

			plan, err := evaluator.AlgebrizeQuery(statement,
				defaultDbName, testVariables, testCatalog)
			req.Nil(err, "failed to algebrize query")
			ctx := createTestConnectionCtx(testInfo)
			optimized, err := evaluator.OptimizeSubqueries(ctx,
				ctx.Logger(""), plan, false)
			req.Nil(err, "failed to optimize subqueries")

			subqueryPlan := evaluator.GetSubqueryPlan(optimized)

			actual := evaluator.GetNodePipeline(subqueryPlan)

			actual, expected := normalizeBSON(actual).([][]bson.D),
				normalizeBSON(testCase.expected).([][]bson.D)

			req.Equal(actual, expected, "actual does not match expected")
		})
	}

	type executeTest struct {
		sql  string
		data []bson.D
	}

	testExecute := func(t *testing.T, testCase executeTest) {
		t.Run(testCase.sql, func(t *testing.T) {
			req := require.New(t)
			statement, err := parser.Parse(testCase.sql)
			req.Nil(err, "failed to parse query")

			plan, err := evaluator.AlgebrizeQuery(statement, defaultDbName,
				testVariables, testCatalog)
			req.Nil(err, "failed to algebrize query")

			//fmt.Printf("\n%+v\n", PrettyPrintPlan(plan))

			sourceReplacer := &evaluator.SourceStageReplacer{Data: testCase.data}
			replaced, err := sourceReplacer.VisitStage(plan)
			req.Nil(err, "soureReplacer failed")
			req.Equal(sourceReplacer.Existing, 0, "sourceReplacer.Existing should be 0")

			//fmt.Printf("\n%+v\n", PrettyPrintPlan(replaced.(PlanStage)))

			ctx := createTestConnectionCtx(testInfo)
			optimized, err := evaluator.OptimizeSubqueries(ctx,
				ctx.Logger(""), replaced, true)
			req.Nil(err, "failed to optimize subqueries")

			sourceReplacer = &evaluator.SourceStageReplacer{}
			sourceReplacer.VisitStage(optimized)
			req.Equal(sourceReplacer.Existing, 1,
				"sourceReplacer.Existing should be 1")
			req.Equal(sourceReplacer.Replaced, 0,
				"sourceReplacer.Replaced should be 0")
		})
	}

	testCache := func(t *testing.T, testCase executeTest) {
		t.Run(testCase.sql, func(t *testing.T) {
			req := require.New(t)
			statement, err := parser.Parse(testCase.sql)
			req.Nil(err, "failed to parse query")

			plan, err := evaluator.AlgebrizeQuery(statement, defaultDbName,
				testVariables, testCatalog)
			req.Nil(err, "failed to algebrize query")

			sourceReplacer := &evaluator.SourceStageReplacer{Data: testCase.data}
			replaced, err := sourceReplacer.VisitStage(plan)
			req.Nil(err, "soureReplacer failed")
			req.Equal(sourceReplacer.Existing, 0, "sourceReplacer.Existing should be 0")

			ctx := createTestConnectionCtx(testInfo)

			optimized, err := evaluator.OptimizeSubqueries(ctx,
				ctx.Logger(""), replaced, true)
			req.Nil(err, "failed to optimize subqueries")

			req.Equal(evaluator.GetCacheStateCount(optimized), 1,
				"GetCacheStateCount(optimized) should be 1")

		})
	}

	runTestsAsSubtest := func(subTestName string, tests []test) {
		t.Run(subTestName, func(t *testing.T) {
			for _, testCase := range tests {
				testOptimize(t, testCase)
			}
		})
	}

	type testFunc func(t *testing.T, testCase executeTest)

	runExecuteTestsAsSubtest := func(subTestName string, tf testFunc, tests []executeTest) {
		t.Run(subTestName, func(t *testing.T) {
			for _, testCase := range tests {
				tf(t, testCase)
			}
		})
	}

	makeDocs := func(docs ...[]bson.D) [][]bson.D {
		return docs
	}

	// Subquery Optimization
	optimizeTests := []test{{

		"select a, (select b from bar) from foo",
		makeDocs([]bson.D{
			{{Name: "$project", Value: bson.M{
				"test_DOT_bar_DOT_b": "$b",
			}}},
		})}, {

		"select exists(select a from bar) from foo",
		makeDocs([]bson.D{
			{{Name: "$project", Value: bson.M{
				"test_DOT_bar_DOT_a": "$a",
			}}},
		})}, {

		"select a from bar where `a` = (select `b` from bar where b=2)",
		makeDocs([]bson.D{
			{{Name: "$match", Value: bson.M{
				"b": int64(2),
			}}},
			{{Name: "$project", Value: bson.M{
				"test_DOT_bar_DOT_b": "$b",
			}}},
		})}, {

		"select a from bar where `a` = (select `b` from bar where b = (select a" +
			" from bar where a=1))",
		makeDocs([]bson.D{
			{{Name: "$project", Value: bson.M{
				"test_DOT_bar_DOT_b": "$b",
			}}},
			{{Name: "$project", Value: bson.M{
				"test_DOT_bar_DOT_b": "$test_DOT_bar_DOT_b",
			}}},
		},
			[]bson.D{
				{{Name: "$match", Value: bson.M{
					"a": int64(1),
				}}},
				{{Name: "$project", Value: bson.M{
					"test_DOT_bar_DOT_a": "$a",
				}}},
			})}, {

		"select a from bar where (`a`, `b`) = (select `c`, `b` from foo where" +
			" b=2)",
		makeDocs([]bson.D{
			{{Name: "$match", Value: bson.M{
				"b": int64(2),
			}}},
			{{Name: "$project", Value: bson.M{
				"test_DOT_foo_DOT_c": "$c",
				"test_DOT_foo_DOT_b": "$b",
			}}},
		})},
	}

	runTestsAsSubtest("Subquery Optimization Tests", optimizeTests)

	// Subquery Execution and Replacement
	replacementTests := []executeTest{{
		"select a, (select b from bar) from foo",
		[]bson.D{
			{{Name: "b", Value: 1}},
			{{Name: "a", Value: 1}},
		}}, {

		"select a from bar where `a` = (select `b` from bar where b=2)",
		[]bson.D{
			{{Name: "b", Value: 2}},
			{{Name: "a", Value: 2}},
		}}, {

		"select a from bar where `a` = (select `b` from bar where b = (select a" +
			" from bar where a=1))",
		[]bson.D{
			{{Name: "a", Value: 1}},
			{{Name: "b", Value: 1}},
			{{Name: "a", Value: 1}},
		}}, {

		"select a from bar where (`a`, `b`) = (select `c`, `b` from foo where b=2)",
		[]bson.D{
			{{Name: "b", Value: 1}, {Name: "c", Value: 1}},
			{{Name: "a", Value: 1}},
		}},
	}
	runExecuteTestsAsSubtest("Subquery Execution and Replacement Tests",
		testExecute, replacementTests)

	// Subquery Execution and Cachine
	cacheTests := []executeTest{{
		"select a from foo where a in (select b from bar)",
		[]bson.D{
			{{Name: "a", Value: 1}},
			{{Name: "b", Value: 1}},
		}}, {

		"select a from foo where a not in (select b from bar)",
		[]bson.D{
			{{Name: "a", Value: 1}},
			{{Name: "b", Value: 1}},
		}}, {

		"select a from foo where a < all (select b from bar)",
		[]bson.D{
			{{Name: "a", Value: 1}},
			{{Name: "b", Value: 1}},
		}}, {

		"select a from foo where a >= some (select b from bar)",
		[]bson.D{
			{{Name: "a", Value: 1}},
			{{Name: "b", Value: 1}},
		}}, {

		"select a from foo where a < any (select b from bar)",
		[]bson.D{
			{{Name: "a", Value: 1}},
			{{Name: "b", Value: 1}},
		}}, {

		"select a from foo where (`a`, `c`) in (select `a`, `b` from bar)",
		[]bson.D{
			{{Name: "a", Value: 1}, {Name: "c", Value: 2}},
			{{Name: "a", Value: 1}, {Name: "b", Value: 2}},
		}}, {

		"select a from foo where (`a`, `c`) not in (select `a`, `b` from bar)",
		[]bson.D{
			{{Name: "a", Value: 1}, {Name: "c", Value: 2}},
			{{Name: "a", Value: 1}, {Name: "b", Value: 3}},
		}},
	}
	runExecuteTestsAsSubtest("Subquery Execution and Cache Tests",
		testCache, cacheTests)
}

func TestOptimizeEvaluations(t *testing.T) {

	type test struct {
		sql      string
		expected string
		result   evaluator.SQLExpr
	}

	testSchema := evaluator.MustLoadSchema(testSchema4)
	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	runTests := func(tests []test) {
		schema := evaluator.MustLoadSchema(testSchema3)
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
			{"3 / '3'", "1", evaluator.SQLFloat(1)},
			{"3 * '3'", "9", evaluator.SQLInt(9)},
			{"3 + '3'", "6", evaluator.SQLInt(6)},
			{"3 - '3'", "0", evaluator.SQLInt(0)},
			{"3 div '3'", "1", evaluator.SQLInt(1)},
			{"3 = '3'", "true", evaluator.SQLTrue},
			{"3 <= '3'", "true", evaluator.SQLTrue},
			{"3 >= '3'", "true", evaluator.SQLTrue},
			{"3 < '3'", "false", evaluator.SQLFalse},
			{"3 > '3'", "false", evaluator.SQLFalse},
			{"3 <=> '3'", "true", evaluator.SQLTrue},
			{"3 = a", "a = 3", evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test",
				"bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			{"3 < a", "a > 3", evaluator.NewSQLGreaterThanExpr(evaluator.NewSQLColumnExpr(1, "test",
				"bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			{"3 <= a", "a >= 3", evaluator.NewSQLGreaterThanOrEqualExpr(evaluator.NewSQLColumnExpr(
				1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			{"3 > a", "a < 3", evaluator.NewSQLLessThanExpr(evaluator.NewSQLColumnExpr(1, "test",
				"bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			{"3 >= a", "a <= 3", evaluator.NewSQLLessThanOrEqualExpr(evaluator.NewSQLColumnExpr(1,
				"test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			{"3 <> a", "a <> 3", evaluator.NewSQLNotEqualsExpr(evaluator.NewSQLColumnExpr(1, "test",
				"bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			{"3 + 3 = 6", "true", evaluator.SQLTrue},
			{"3 <=> 3", "true", evaluator.SQLTrue},
			{"NULL <=> 3", "false", evaluator.SQLFalse},
			{"3 <=> NULL", "false", evaluator.SQLFalse},
			{"NULL <=> NULL", "true", evaluator.SQLTrue},
			{"3 / (3 - 2) = a", "a = 3", evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1,
				"test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLFloat(3))},
			{"3 + 3 = 6 AND 1 >= 1 AND 3 = a", "a = 3", evaluator.NewSQLEqualsExpr(
				evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt),
				evaluator.SQLInt(3))},
			{"3 / (3 - 2) = a AND 4 - 2 = b", "a = 3 AND b = 2",
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
						schema.SQLInt, schema.MongoInt), evaluator.SQLFloat(3)),
					evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "b",
						schema.SQLInt, schema.MongoInt), evaluator.SQLInt(2)))},
			{"3 + 3 = 6 OR a = 3", "true", evaluator.SQLTrue},
			{"3 + 3 = 5 OR a = 3", "a = 3", evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1,
				"test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			{"0 OR NULL", "null", evaluator.SQLNull},
			{"1 OR NULL", "true", evaluator.SQLTrue},
			{"NULL OR NULL", "null", evaluator.SQLNull},
			{"0 AND 6+1 = 6", "false", evaluator.SQLFalse},
			{"3 + 3 = 5 AND a = 3", "false", evaluator.SQLFalse},
			{"0 AND NULL", "false", evaluator.SQLFalse},
			{"1 AND NULL", "null", evaluator.SQLNull},
			{"1 AND 6+0 = 6", "true", evaluator.SQLTrue},
			{"3 + 3 = 6 AND a = 3", "a = 3", evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(
				1, "test", "bar", "a", schema.SQLInt, schema.MongoInt), evaluator.SQLInt(3))},
			{"(3 + 3 = 5) XOR a = 3", "a = 3", evaluator.NewSQLEqualsExpr(
				evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt),
				evaluator.SQLInt(3))},
			{"(3 + 3 = 6) XOR a = 3", "a <> 3", evaluator.NewSQLNotExpr(evaluator.NewSQLEqualsExpr(
				evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt),
				evaluator.SQLInt(3)))},
			{"(13 + 9 > 6) XOR (a = 4)", "a <> 4", evaluator.NewSQLNotExpr(
				evaluator.NewSQLEqualsExpr(evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
					schema.SQLInt, schema.MongoInt), evaluator.SQLInt(4)))},
			{"(8 / 5 = 9) XOR (a = 5)", "a = 5", evaluator.NewSQLEqualsExpr(
				evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt),
				evaluator.SQLInt(5))},
			{"false XOR 23", "true", evaluator.SQLTrue},
			{"true XOR 23", "false", evaluator.SQLFalse},
			{"a = 23 XOR true", "a <> 23", evaluator.NewSQLNotExpr(evaluator.NewSQLEqualsExpr(
				evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt),
				evaluator.SQLInt(23)))},
			{"!3", "0", evaluator.SQLFalse},
			{"!NULL", "null", evaluator.SQLNull},
			{"a = ~1", "a = 18446744073709551614", evaluator.NewSQLEqualsExpr(
				evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt),
				evaluator.SQLUint64(18446744073709551614))},
			{"a = ~2398238912332232323", "a = 16048505161377319292", evaluator.NewSQLEqualsExpr(
				evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt),
				evaluator.SQLUint64(16048505161377319292))},
			{"DAYNAME('2016-1-1')", "Friday", evaluator.SQLVarchar("Friday")},
			{"(8-7)", "1", evaluator.SQLInt(1)},
			{"a LIKE NULL", "null", evaluator.SQLNull},
			{"4 LIKE NULL", "null", evaluator.SQLNull},
			{"a = NULL", "null", evaluator.SQLNull},
			{"a > NULL", "null", evaluator.SQLNull},
			{"a >= NULL", "null", evaluator.SQLNull},
			{"a < NULL", "null", evaluator.SQLNull},
			{"a <= NULL", "null", evaluator.SQLNull},
			{"a != NULL", "null", evaluator.SQLNull},
			{"(1, 3) > (3, 4)", "SQLFalse", evaluator.SQLFalse},
			{"(4, 3) > (3, 4)", "SQLTrue", evaluator.SQLTrue},
			{"(4, 31) > (4, 4)", "SQLTrue", evaluator.SQLTrue},

			{"abs(NULL)", "null", evaluator.SQLNull},
			{"abs(-10)", "10", evaluator.SQLFloat(10)},
			{"ascii(NULL)", "null", evaluator.SQLNull},
			{"ascii('a')", "97", evaluator.SQLInt(97)},
			{"char_length(NULL)", "null", evaluator.SQLNull},
			{"character_length(NULL)", "null", evaluator.SQLNull},
			{"concat(NULL, a)", "null", evaluator.SQLNull},
			{"concat(a, NULL)", "null", evaluator.SQLNull},
			{"concat('go', 'lang')", "golang", evaluator.SQLVarchar("golang")},
			{"concat_ws(NULL, a)", "null", evaluator.SQLNull},
			{"convert(NULL, SIGNED)", "null", evaluator.SQLNull},
			{"elt(NULL, 'a', 'b')", "null", evaluator.SQLNull},
			{"elt(4, 'a', 'b')", "null", evaluator.SQLNull},
			{"exp(NULL)", "null", evaluator.SQLNull},
			{"exp(2)", "7.38905609893065", evaluator.SQLFloat(7.38905609893065)},
			{"greatest(a, NULL)", "null", evaluator.SQLNull},
			{"greatest(2, 3)", "3", evaluator.SQLInt(3)},
			{"ifnull(NULL, a)", "bar.a", evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
				schema.SQLInt, schema.MongoInt)},
			{"ifnull(10, a)", "10", evaluator.SQLInt(10)},
			{"interval(NULL, a)", "-1", evaluator.SQLInt(-1)},
			{"interval(0, 1)", "0", evaluator.SQLInt(0)},
			{"interval(1, 2, 3, 4)", "1", evaluator.SQLInt(0)},
			{"interval(1, 1, 2, 3)", "1", evaluator.SQLInt(1)},
			{"interval(-1, NULL, NULL, -0.5, 3, 4)", "1", evaluator.SQLInt(2)},
			{"interval(-3.4, -4, -3.6, -3.4, -3, 1, 2)", "3", evaluator.SQLInt(3)},
			{"interval(8, -4, 0, 7, 8)", "4", evaluator.SQLInt(4)},
			{"interval(8, -3, 1, 7, 7)", "1", evaluator.SQLInt(4)},
			{"interval(7.7, -3, 1, 7, 7)", "1", evaluator.SQLInt(4)},
			{"least(a, NULL)", "null", evaluator.SQLNull},
			{"least(2, 3)", "2", evaluator.SQLInt(2)},
			{"locate('bar', 'foobar', NULL)", "0", evaluator.SQLInt(0)},
			{"locate('bar', 'foobar')", "4", evaluator.SQLInt(4)},
			{"makedate(2000, NULL)", "null", evaluator.SQLNull},
			{"makedate(NULL, 10)", "null", evaluator.SQLNull},
			{"mid('foobar', NULL, 2)", "null", evaluator.SQLNull},
			{"mod(10, 2)", "0", evaluator.SQLFloat(0)},
			{"mod(NULL, 2)", "null", evaluator.SQLNull},
			{"mod(10, NULL)", "null", evaluator.SQLNull},
			{"nullif(NULL, a)", "null", evaluator.SQLNull},
			{"nullif(a, NULL)", "bar.a", evaluator.NewSQLColumnExpr(1, "test", "bar", "a",
				schema.SQLInt, schema.MongoInt)},
			{"pow(a, NULL)", "null", evaluator.SQLNull},
			{"pow(NULL, a)", "null", evaluator.SQLNull},
			{"pow(2,2)", "4", evaluator.SQLFloat(4)},
			{"round(NULL, 2)", "null", evaluator.SQLNull},
			{"round(2, NULL)", "null", evaluator.SQLNull},
			{"round(2, 2)", "2", evaluator.SQLFloat(2)},
			{"repeat('a', NULL)", "null", evaluator.SQLNull},
			{"repeat(NULL, 3)", "null", evaluator.SQLNull},
			{"substring(NULL, 2)", "null", evaluator.SQLNull},
			{"substring(NULL, 2, 3)", "null", evaluator.SQLNull},
			{"substring('foobar', NULL)", "null", evaluator.SQLNull},
			{"substring('foobar', NULL, 2)", "null", evaluator.SQLNull},
			{"substring('foobar', 2, NULL)", "null", evaluator.SQLNull},
			{"substring('foobar', 2, 3)", "oob", evaluator.SQLVarchar("oob")},
			{"substring_index(NULL, 'o', 0)", "", evaluator.SQLNull},
			{"substring_index('foobar', 'o', 0)", "", evaluator.SQLVarchar("")},
		}

		runTests(tests)

	})
}

func TestOptimizeEvaluationFailures(t *testing.T) {

	type test struct {
		sql string
		err error
	}

	testSchema := evaluator.MustLoadSchema(testSchema4)
	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	runTests := func(tests []test) {
		schema := evaluator.MustLoadSchema(testSchema3)
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
			{"pow(-2,2.2)", mysqlerrors.Defaultf(mysqlerrors.ErDataOutOfRange, "DOUBLE",
				"pow(-2,2.2)")},
			{"pow(0,-2.2)", mysqlerrors.Defaultf(mysqlerrors.ErDataOutOfRange, "DOUBLE",
				"pow(0,-2.2)")},
			{"pow(0,-5)", mysqlerrors.Defaultf(mysqlerrors.ErDataOutOfRange, "DOUBLE",
				"pow(0,-5)")},
		}

		runTests(tests)

	})
}
