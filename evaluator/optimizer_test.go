package evaluator_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/astprint"
	"github.com/10gen/mongoast/optimizer"
	"github.com/10gen/sqlproxy/collation"
	. "github.com/10gen/sqlproxy/evaluator"
	. "github.com/10gen/sqlproxy/evaluator/types"
	. "github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"
)

// Fully pushed-down queries are covered by TestPushdownPlan in
// optimizer_pushdown_test.go. This test covers the remaining cases, testing
// how we push down queries that are only partially pushed down on certain
// MongoDB versions.
func TestOptimizePartialPushdown(t *testing.T) {

	bgCtx := context.Background()

	type test struct {
		name     string
		sql      string
		expected []*ast.Pipeline
		versions []string
	}

	tests := []test{
		{
			name:     "count_star",
			sql:      "select count(*) from foo",
			expected: []*ast.Pipeline{},
		},
		{
			name:     "count_star_with_order",
			sql:      "select count(*) from foo order by 1",
			expected: []*ast.Pipeline{},
		},
		{
			name:     "huge_limit",
			sql:      "select a from foo limit 18446744073709551614",
			expected: []*ast.Pipeline{},
		},
		{
			name: "nopushdown_scalar_function_in_select",
			sql:  "select a+a, nopushdown(a+a) from foo",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewProjectStage(
						// Only one add should be push down.
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a+test_DOT_foo_DOT_a",
							astutil.WrapInOp(bsonutil.OpAdd,
								ast.NewFieldRef("a", nil),
								ast.NewFieldRef("a", nil),
							),
						),
						ast.NewAssignProjectItem("a", ast.NewFieldRef("a", nil)),
						ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
					),
				),
			},
		},
		{
			name: "nopushdown_scalar_function_in_where",
			sql:  "select a+a from foo where nopushdown(b=1)",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a", ast.NewFieldRef("a", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_b", ast.NewFieldRef("b", nil)),
					),
				),
			},
		},
		{
			name: "nopushdown_scalar_function_in_orderby",
			sql:  "select a+a from foo order by nopushdown(a=1)",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
			},
		},
		{
			name: "nopushdown_scalar_function_in_orderby_after_where",
			sql:  "select a+a from foo where a > 3 order by nopushdown(a=1)",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewMatchStage(
						ast.NewBinary(bsonutil.OpGt, ast.NewFieldRef("a", nil), astutil.Int64Constant(3)),
					),
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
			},
		},
		{
			name: "nopushdown_scalar_function_in_groupby",
			sql:  "select a+a from foo group by nopushdown(a)",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
			},
		},
		{
			name: "nopushdown_scalar_function_in_join",
			sql: "select * from (select a+a as a from bar) a " +
				" inner join (select a+a as a, concat(b, b) from bar) b on nopushdown(a.a) = b.a",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_a_DOT_a",
							astutil.WrapInOp(bsonutil.OpAdd,
								ast.NewFieldRef("a", nil),
								ast.NewFieldRef("a", nil),
							),
						),
					),
				),
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_bar_DOT_a+test_DOT_bar_DOT_a",
							astutil.WrapInOp(bsonutil.OpAdd,
								ast.NewFieldRef("a", nil),
								ast.NewFieldRef("a", nil),
							),
						),
						ast.NewAssignProjectItem("b", ast.NewFieldRef("b", nil)),
					),
				),
			},
		},
		{
			name:     "inner_joins_subqueries_nested",
			versions: []string{"3.2", "3.4"},
			sql: "select * from (select foo.a from bar join (select foo.a from foo) foo on" +
				" foo.a=bar.b) x join (select g.a from bar join (select foo.a from foo) g on " +
				"g.a=bar.a) y on x.a=y.a",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a", ast.NewFieldRef("a", nil)),
					),
					ast.NewMatchStage(
						ast.NewBinary(bsonutil.OpNeq,
							ast.NewFieldRef("test_DOT_foo_DOT_a", nil),
							astutil.NullLiteral,
						),
					),
					ast.NewLookupStage("bar", ast.NewFieldRef("test_DOT_foo_DOT_a", nil), "b", "__joined_bar",
						nil, nil,
					),
					ast.NewUnwindStage(
						ast.NewFieldRef("__joined_bar", nil),
						"", false,
					),
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_x_DOT_a", ast.NewFieldRef("test_DOT_foo_DOT_a", nil)),
					),
				),
				// another pipeline
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a", ast.NewFieldRef("a", nil)),
					),
					ast.NewMatchStage(
						ast.NewBinary(bsonutil.OpNeq,
							ast.NewFieldRef("test_DOT_foo_DOT_a", nil),
							astutil.NullLiteral,
						),
					),
					ast.NewLookupStage("bar", ast.NewFieldRef("test_DOT_foo_DOT_a", nil), "a", "__joined_bar",
						nil, nil,
					),
					ast.NewUnwindStage(
						ast.NewFieldRef("__joined_bar", nil),
						"", false,
					),
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_y_DOT_a", ast.NewFieldRef("test_DOT_foo_DOT_a", nil)),
					),
				),
			},
		},
		{
			name:     "left_join_inner_join_subqueries_nested",
			versions: []string{"3.2", "3.4"},
			sql: "select * from foo f left join (select b.b from foo f join (select * from " +
				"bar) b on f.a=b.a)  b on f.a=b.b",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_f_DOT__id", ast.NewFieldRef("_id", nil)),
						ast.NewAssignProjectItem("test_DOT_f_DOT_a", ast.NewFieldRef("a", nil)),
						ast.NewAssignProjectItem("test_DOT_f_DOT_b", ast.NewFieldRef("b", nil)),
						ast.NewAssignProjectItem("test_DOT_f_DOT_c", ast.NewFieldRef("c", nil)),
						ast.NewAssignProjectItem("test_DOT_f_DOT_e", astutil.FieldRefFromFieldName("d.e")),
						ast.NewAssignProjectItem("test_DOT_f_DOT_f", astutil.FieldRefFromFieldName("d.f")),
						ast.NewAssignProjectItem("test_DOT_f_DOT_g", ast.NewFieldRef("g", nil)),
					),
				),
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_bar_DOT_a", ast.NewFieldRef("a", nil)),
						ast.NewAssignProjectItem("test_DOT_bar_DOT_b", ast.NewFieldRef("b", nil)),
					),
					ast.NewMatchStage(
						ast.NewBinary(bsonutil.OpNeq,
							ast.NewFieldRef("test_DOT_bar_DOT_a", nil),
							astutil.NullLiteral,
						),
					),
					ast.NewLookupStage("foo", ast.NewFieldRef("test_DOT_bar_DOT_a", nil), "a", "__joined_f",
						nil, nil,
					),
					ast.NewUnwindStage(
						ast.NewFieldRef("__joined_f", nil),
						"", false,
					),
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_b_DOT_b", ast.NewFieldRef("test_DOT_bar_DOT_b", nil)),
					),
				),
			}},

		{
			name: "join_nested_array_tables_0",
			sql: "select * from foo f join merge m1 on f._id=m1._id join (select * from foo) g" +
				" on g.a=f.a join merge_d_a m2 on m2._id=m1._id and m2._id=g.a",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewUnwindStage(ast.NewFieldRef("d", nil), "d_idx", false),
					ast.NewUnwindStage(
						ast.NewFieldRef("a", ast.NewFieldRef("d", nil)),
						"d.a_idx", false,
					),
					ast.NewMatchStage(
						ast.NewBinary(bsonutil.OpNeq,
							ast.NewFieldRef("_id", nil),
							astutil.NullLiteral,
						),
					),
					ast.NewLookupStage("foo", ast.NewFieldRef("_id", nil), "_id", "__joined_f",
						nil, nil,
					),
					ast.NewUnwindStage(ast.NewFieldRef("__joined_f", nil), "", false),
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_f_DOT__id", astutil.FieldRefFromFieldName("__joined_f._id")),
						ast.NewAssignProjectItem("test_DOT_f_DOT_a", astutil.FieldRefFromFieldName("__joined_f.a")),
						ast.NewAssignProjectItem("test_DOT_f_DOT_b", astutil.FieldRefFromFieldName("__joined_f.b")),
						ast.NewAssignProjectItem("test_DOT_f_DOT_c", astutil.FieldRefFromFieldName("__joined_f.c")),
						ast.NewAssignProjectItem("test_DOT_f_DOT_e", astutil.FieldRefFromFieldName("__joined_f.d.e")),
						ast.NewAssignProjectItem("test_DOT_f_DOT_f", astutil.FieldRefFromFieldName("__joined_f.d.f")),
						ast.NewAssignProjectItem("test_DOT_f_DOT_g", astutil.FieldRefFromFieldName("__joined_f.g")),
						ast.NewAssignProjectItem("test_DOT_m1_DOT__id", ast.NewFieldRef("_id", nil)),
						ast.NewAssignProjectItem("test_DOT_m1_DOT_a", ast.NewFieldRef("a", nil)),
						ast.NewAssignProjectItem("test_DOT_m2_DOT__id", ast.NewFieldRef("_id", nil)),
						ast.NewAssignProjectItem("test_DOT_m2_DOT_d_DOT_a", astutil.FieldRefFromFieldName("d.a")),
						ast.NewAssignProjectItem("test_DOT_m2_DOT_d_DOT_a_idx", astutil.FieldRefFromFieldName("d.a_idx")),
						ast.NewAssignProjectItem("test_DOT_m2_DOT_d_idx", ast.NewFieldRef("d_idx", nil)),
					),
				),
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_g_DOT__id", ast.NewFieldRef("_id", nil)),
						ast.NewAssignProjectItem("test_DOT_g_DOT_a", ast.NewFieldRef("a", nil)),
						ast.NewAssignProjectItem("test_DOT_g_DOT_b", ast.NewFieldRef("b", nil)),
						ast.NewAssignProjectItem("test_DOT_g_DOT_c", ast.NewFieldRef("c", nil)),
						ast.NewAssignProjectItem("test_DOT_g_DOT_e", astutil.FieldRefFromFieldName("d.e")),
						ast.NewAssignProjectItem("test_DOT_g_DOT_f", astutil.FieldRefFromFieldName("d.f")),
						ast.NewAssignProjectItem("test_DOT_g_DOT_g", ast.NewFieldRef("g", nil)),
					),
				),
			},
		},
		{
			name:     "join_subqueries_where_limit",
			versions: []string{"3.2", "3.4"},
			sql: "select f.a from foo f join (select bar.a from bar) b on f.a=b.a join " +
				"(select foo.a from foo where foo.a > 4 limit 1) c on b.a=c.a and f.a=c.a and " +
				"f.b=b.a",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewMatchStage(
						ast.NewBinary(bsonutil.OpGt,
							ast.NewFieldRef("a", nil),
							astutil.Int64Constant(4),
						),
					),
					ast.NewLimitStage(1),
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_c_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_b_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_f_DOT_a", ast.NewFieldRef("a", nil)),
						ast.NewAssignProjectItem("test_DOT_f_DOT_b", ast.NewFieldRef("b", nil)),
					),
				),
			},
		},
		{
			name:     "right_non_equijoin",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo right join bar on foo.a < bar.a",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_bar_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
			},
		},
		{
			name:     "self_join_0",
			versions: []string{"3.2", "3.4"},
			sql:      "select * from merge r left join merge_d_a a on r._id=a._id",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_r_DOT__id", ast.NewFieldRef("_id", nil)),
						ast.NewAssignProjectItem("test_DOT_r_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
				ast.NewPipeline(
					ast.NewUnwindStage(ast.NewFieldRef("d", nil), "d_idx", false),
					ast.NewUnwindStage(
						ast.NewFieldRef("a", ast.NewFieldRef("d", nil)),
						"d.a_idx", false,
					),
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_a_DOT__id", ast.NewFieldRef("_id", nil)),
						ast.NewAssignProjectItem("test_DOT_a_DOT_d_DOT_a", astutil.FieldRefFromFieldName("d.a")),
						ast.NewAssignProjectItem("test_DOT_a_DOT_d_DOT_a_idx", astutil.FieldRefFromFieldName("d.a_idx")),
						ast.NewAssignProjectItem("test_DOT_a_DOT_d_idx", ast.NewFieldRef("d_idx", nil)),
					),
				),
			},
		},
		{
			name:     "self_join_4",
			versions: []string{"3.4"},
			sql: "select b._id, c._id from merge r left join merge_b b on r._id=b._id inner" +
				" join merge_c c on r._id=c._id left join merge_d_a a on r._id=a._id",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewUnwindStage(ast.NewFieldRef("b", nil), "", true),
					ast.NewUnwindStage(ast.NewFieldRef("c", nil), "", false),
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_b_DOT__id", astutil.WrapInCond(
							astutil.NullLiteral,
							ast.NewFieldRef("_id", nil),
							ast.NewBinary(bsonutil.OpLte, ast.NewFieldRef("b", nil), astutil.NullLiteral),
							ast.NewBinary(bsonutil.OpEq, ast.NewFieldRef("b", nil), ast.NewArray()),
						)),
						ast.NewAssignProjectItem("test_DOT_c_DOT__id", ast.NewFieldRef("_id", nil)),
						ast.NewAssignProjectItem("test_DOT_r_DOT__id", ast.NewFieldRef("_id", nil)),
					),
				),
				ast.NewPipeline(
					ast.NewUnwindStage(ast.NewFieldRef("d", nil), "", false),
					ast.NewUnwindStage(ast.NewFieldRef("a", ast.NewFieldRef("d", nil)),
						"", false,
					),
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_a_DOT__id", ast.NewFieldRef("_id", nil)),
					),
				),
			},
		},
		{
			name:     "non_equijoin_0",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo inner join bar on foo.a < bar.a",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_bar_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
			},
		},
		{
			name:     "non_equijoin_2",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo, bar where foo.a < bar.a",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_bar_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
			},
		},
		{
			name:     "non_equijoin_3",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo left join bar on foo.a < bar.a",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_bar_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
			},
		},
		{
			name:     "non_equijoin_4",
			versions: []string{"3.2", "3.4"},
			sql:      "select foo.a from foo right join bar on foo.a < bar.a",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_bar_DOT_a", ast.NewFieldRef("a", nil)),
					),
				),
			},
		},
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

					testSchema := MustLoadSchema(optimizerTestSchema)

					testInfo := GetMongoDBInfo(versionByStr[version], testSchema, mongodb.AllPrivileges)
					testVariables := CreateTestVariables(testInfo)
					testSchemaCatalog := GetCatalog(testSchema, testVariables, testInfo)
					defaultDbName := "test"

					statement, err := parser.Parse(test.sql)
					req.Nil(err, "failed to parse statement")

					rCfg := NewRewriterConfig(log.GlobalLogger(), false)

					rewritten, err := RewriteQuery(rCfg, statement)
					req.Nil(err, "failed to rewrite query")

					aCfg := createAlgebrizerCfg(defaultDbName, testSchemaCatalog)
					plan, err := AlgebrizeQuery(aCfg, rewritten)

					req.Nil(err, "failed to algebrize query")

					eCfg := createExecutionCfg("test_db_name", 0, versionByStr[version], MySQLValueKind)
					oCfg := createOptimizerCfg(collation.Default, eCfg)
					optimizedPlan, err := OptimizePlan(context.Background(), oCfg, plan)
					req.Nil(err, "failed to optimize plan")

					pCfg := createPushdownCfg(versionByStr[version], MySQLValueKind)
					pushedDown, err := PushdownPlan(bgCtx, pCfg, optimizedPlan)

					var actualPlan PlanStage
					if err != nil && !IsNonFatalPushdownError(err) {
						actualPlan = optimizedPlan
					} else {
						actualPlan = pushedDown
					}

					actualNonNormalized := GetNodePipeline(actualPlan)
					actual := make([]*ast.Pipeline, len(actualNonNormalized))
					for i, pipeline := range actualNonNormalized {
						actual[i] = optimizer.NormalizePipeline(pipeline)
					}
					expected := make([]*ast.Pipeline, len(test.expected))
					for i, pipeline := range test.expected {
						expected[i] = optimizer.NormalizePipeline(pipeline)
					}

					req.Equalf(len(expected), len(actual),
						"expected %d pipelines in query plan, found %d\nexpected pipelines: "+
							"%#v\nactual pipelines: %#v\nactual plan:\n%s",
						len(expected), len(actual), expected, actual,
						PrettyPrintPlan(actualPlan))

					diff := ShouldResembleDiffed(actual, expected)
					expectedJSON := ""
					actualJSON := ""
					for i := range expected {
						expectedJSON += strconv.Itoa(i) + ":\n" + astprint.ShellString(expected[i]) + "\n"
					}
					for i := range actual {
						actualJSON += strconv.Itoa(i) + ":\n" + astprint.ShellString(actual[i]) + "\n"
					}
					req.Emptyf(diff, "expected pipeline diff to be empty\nexpected: %#v\nactual:"+
						" %#v\nexpected(json):\n%sactual(json):\n%s\n", expected, actual, expectedJSON, actualJSON)
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
	bgCtx := context.Background()
	testSchema := MustLoadSchema(testSchema4)
	testInfo := getMongoDBInfoWithShardedCollection(nil, testSchema, mongodb.AllPrivileges, "foo")
	testVariables := CreateTestVariables(testInfo)
	testSchemaCatalog := GetCatalog(testSchema, testVariables, testInfo)
	defaultDbName := "test"

	type test struct {
		sql      string
		expected []*ast.Pipeline
	}
	runTests := func(tests []test) {
		for _, test := range tests {
			t.Run(test.sql, func(t *testing.T) {
				req := require.New(t)

				statement, err := parser.Parse(test.sql)
				req.NoError(err)

				rCfg := NewRewriterConfig(log.GlobalLogger(), false)

				rewritten, err := RewriteQuery(rCfg, statement)
				req.NoError(err, "failed to rewrite query")

				aCfg := createAlgebrizerCfg(defaultDbName, testSchemaCatalog)
				plan, err := AlgebrizeQuery(aCfg, rewritten)

				req.NoError(err)

				version := []uint8{3, 4, 0}

				eCfg := createExecutionCfg("test_db", 0, version, MySQLValueKind)
				oCfg := createOptimizerCfg(collation.Default, eCfg)
				optimized, err := OptimizePlan(context.Background(), oCfg, plan)
				req.NoError(err)

				pCfg := createPushdownCfg(version, MySQLValueKind)
				pushedDown, err := PushdownPlan(bgCtx, pCfg, optimized)

				var actualPlan PlanStage
				if err != nil && !IsNonFatalPushdownError(err) {
					actualPlan = optimized
				} else {
					actualPlan = pushedDown
				}

				actualNonNormalized := GetNodePipeline(actualPlan)
				actual := make([]*ast.Pipeline, len(actualNonNormalized))
				for i, pipeline := range actualNonNormalized {
					actual[i] = optimizer.NormalizePipeline(pipeline)
				}
				expected := make([]*ast.Pipeline, len(test.expected))
				for i, pipeline := range test.expected {
					expected[i] = optimizer.NormalizePipeline(pipeline)
				}
				req.Equalf(len(expected), len(actual),
					"expected %d pipelines in query plan, found %d\nexpected pipelines: "+
						"%#v\nactual pipelines: %#v\nactual plan:\n%s",
					len(expected), len(actual), expected, actual,
					PrettyPrintPlan(actualPlan))

				diff := ShouldResembleDiffed(actual, expected)
				expectedJSON := ""
				actualJSON := ""
				for i := range expected {
					expectedJSON += strconv.Itoa(i) + ":\n" + astprint.ShellString(expected[i]) + "\n"
				}
				for i := range actual {
					actualJSON += strconv.Itoa(i) + ":\n" + astprint.ShellString(actual[i]) + "\n"
				}
				req.Emptyf(diff, "expected pipeline diff to be empty\nexpected: %#v\nactual:"+
					" %#v\nexpected(json):\n%sactual(json):\n%s\n", expected, actual, expectedJSON, actualJSON)
			})
		}
	}

	tests := []test{
		// should not push down because the from collection is sharded.
		{
			sql: "select * from bar left join foo on bar.a=foo.a and bar.a=foo.f",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_bar_DOT__id", ast.NewFieldRef("_id", nil)),
						ast.NewAssignProjectItem("test_DOT_bar_DOT_a", ast.NewFieldRef("a", nil)),
						ast.NewAssignProjectItem("test_DOT_bar_DOT_b", ast.NewFieldRef("b", nil)),
					),
				),
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_foo_DOT__id", ast.NewFieldRef("_id", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a", ast.NewFieldRef("a", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_b", ast.NewFieldRef("b", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_c", ast.NewFieldRef("c", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_e", astutil.FieldRefFromFieldName("d.e")),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_f", astutil.FieldRefFromFieldName("d.f")),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_g", ast.NewFieldRef("g", nil)),
					),
				),
			},
		},
		{
			sql: "select * from bar right join foo on bar.a=foo.a and bar.a=foo.f",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewLookupStage("bar", ast.NewFieldRef("a", nil), "a", "__joined_bar", nil, nil),
					ast.NewProjectStage(
						ast.NewAssignProjectItem("__joined_bar",
							astutil.WrapInFilter(
								astutil.WrapInCond(
									ast.NewArray(),
									ast.NewFieldRef("__joined_bar", nil),
									ast.NewBinary(ast.LessThanOrEquals,
										ast.NewFieldRef("a", nil),
										astutil.NullConstant(),
									),
								),
								"this",
								astutil.WrapInNullCheckedCond(
									astutil.NullLiteral,
									ast.NewBinary(bsonutil.OpEq,
										ast.NewFieldRef("a", ast.NewVariableRef("this")),
										astutil.FieldRefFromFieldName("d.f"),
									),
									ast.NewFieldRef("a", ast.NewVariableRef("this")),
									astutil.FieldRefFromFieldName("d.f"),
								),
							)),
						ast.NewIncludeProjectItem(ast.NewFieldRef("_id", nil)),
						ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
						ast.NewIncludeProjectItem(ast.NewFieldRef("b", nil)),
						ast.NewIncludeProjectItem(ast.NewFieldRef("c", nil)),
						ast.NewIncludeProjectItem(astutil.FieldRefFromFieldName("d.e")),
						ast.NewIncludeProjectItem(astutil.FieldRefFromFieldName("d.f")),
						ast.NewIncludeProjectItem(ast.NewFieldRef("g", nil)),
					),
					ast.NewUnwindStage(ast.NewFieldRef("__joined_bar", nil), "", true),
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_bar_DOT__id", astutil.FieldRefFromFieldName("__joined_bar._id")),
						ast.NewAssignProjectItem("test_DOT_bar_DOT_a", astutil.FieldRefFromFieldName("__joined_bar.a")),
						ast.NewAssignProjectItem("test_DOT_bar_DOT_b", astutil.FieldRefFromFieldName("__joined_bar.b")),
						ast.NewAssignProjectItem("test_DOT_foo_DOT__id", ast.NewFieldRef("_id", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a", ast.NewFieldRef("a", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_b", ast.NewFieldRef("b", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_c", ast.NewFieldRef("c", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_e", astutil.FieldRefFromFieldName("d.e")),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_f", astutil.FieldRefFromFieldName("d.f")),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_g", ast.NewFieldRef("g", nil)),
						ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
					),
				),
			},
		},

		// after flipping, the from collection, foo is sharded and it should not push down.
		{
			sql: "select * from foo right join bar on foo.a=bar.a and foo.f=bar.a",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_foo_DOT__id", ast.NewFieldRef("_id", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a", ast.NewFieldRef("a", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_b", ast.NewFieldRef("b", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_c", ast.NewFieldRef("c", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_e", astutil.FieldRefFromFieldName("d.e")),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_f", astutil.FieldRefFromFieldName("d.f")),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_g", ast.NewFieldRef("g", nil)),
					),
				),
				ast.NewPipeline(
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_bar_DOT__id", ast.NewFieldRef("_id", nil)),
						ast.NewAssignProjectItem("test_DOT_bar_DOT_a", ast.NewFieldRef("a", nil)),
						ast.NewAssignProjectItem("test_DOT_bar_DOT_b", ast.NewFieldRef("b", nil)),
					),
				),
			},
		},

		// should flip after not being able to be pushed down the first time due to foo being
		// sharded and then push down.
		{
			sql: "select * from bar inner join foo on bar.a=foo.a and bar.a=foo.f",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewMatchStage(
						ast.NewBinary(ast.NotEquals,
							ast.NewFieldRef("a", nil),
							astutil.NullConstant(),
						),
					),
					ast.NewLookupStage("bar", ast.NewFieldRef("a", nil), "a", "__joined_bar", nil, nil),
					ast.NewUnwindStage(ast.NewFieldRef("__joined_bar", nil), "", false),
					ast.NewAddFieldsStage(
						ast.NewAddFieldsItem("__predicate", ast.NewLet(
							[]*ast.LetVariable{ast.NewLetVariable("predicate",
								astutil.WrapInNullCheckedCond(
									astutil.NullConstant(),
									ast.NewBinary(bsonutil.OpEq,
										astutil.FieldRefFromFieldName("__joined_bar.a"),
										astutil.FieldRefFromFieldName("d.f"),
									),
									astutil.FieldRefFromFieldName("__joined_bar.a"),
									astutil.FieldRefFromFieldName("d.f"),
								),
							)},
							astutil.WrapInBinOp(ast.And,
								ast.NewBinary(bsonutil.OpNeq, ast.NewVariableRef("predicate"), astutil.BooleanConstant(false)),
								ast.NewBinary(bsonutil.OpNeq, ast.NewVariableRef("predicate"), astutil.Int32Constant(0)),
								ast.NewBinary(bsonutil.OpNeq, ast.NewVariableRef("predicate"), astutil.StringConstant("0")),
								ast.NewBinary(bsonutil.OpNeq, ast.NewVariableRef("predicate"), astutil.StringConstant("-0")),
								ast.NewBinary(bsonutil.OpNeq, ast.NewVariableRef("predicate"), astutil.StringConstant("0.0")),
								ast.NewBinary(bsonutil.OpNeq, ast.NewVariableRef("predicate"), astutil.StringConstant("-0.0")),
								ast.NewBinary(bsonutil.OpNeq, ast.NewVariableRef("predicate"), astutil.NullConstant()),
							),
						)),
					),
					ast.NewMatchStage(
						ast.NewBinary(bsonutil.OpEq,
							ast.NewFieldRef("__predicate", nil),
							astutil.BooleanConstant(true),
						),
					),
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_bar_DOT__id", astutil.FieldRefFromFieldName("__joined_bar._id")),
						ast.NewAssignProjectItem("test_DOT_bar_DOT_a", astutil.FieldRefFromFieldName("__joined_bar.a")),
						ast.NewAssignProjectItem("test_DOT_bar_DOT_b", astutil.FieldRefFromFieldName("__joined_bar.b")),
						ast.NewAssignProjectItem("test_DOT_foo_DOT__id", ast.NewFieldRef("_id", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_a", ast.NewFieldRef("a", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_b", ast.NewFieldRef("b", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_c", ast.NewFieldRef("c", nil)),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_e", astutil.FieldRefFromFieldName("d.e")),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_f", astutil.FieldRefFromFieldName("d.f")),
						ast.NewAssignProjectItem("test_DOT_foo_DOT_g", ast.NewFieldRef("g", nil)),
						ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
					),
				),
			},
		},
	}

	runTests(tests)
}

func TestOptimizeEvaluations(t *testing.T) {

	type test struct {
		sql      string
		expected string
		result   SQLExpr
	}

	runTests := func(tests []test) {
		schema := MustLoadSchema(testSchema3)
		for _, tst := range tests {
			tName := fmt.Sprintf("%q should be optimized to %q", tst.sql, tst.expected)
			t.Run(tName, func(t *testing.T) {
				req := require.New(t)

				e, err := GetSQLExpr(schema, dbOne, tableTwoName, tst.sql, false, nil)
				req.NoError(err)

				eCfg := createTestExecutionCfg(MySQLValueKind)
				oCfg := createOptimizerCfg(collation.Default, eCfg)
				result, err := OptimizeEvaluations(oCfg, e)
				req.NoError(err)

				expectedVal, ok := tst.result.(SQLValue)
				if ok && expectedVal.IsNull() {
					actualVal, ok := result.(SQLValue)
					req.True(ok)
					req.True(actualVal.IsNull())
				} else {
					req.Zero(convey.ShouldResemble(result, tst.result))
				}
			})
		}
	}

	tests := []test{
		{"3 / '3'", "1", NewSQLValueExpr(NewSQLFloat(valKind, 1))},
		{"3 * '3'", "9", NewSQLValueExpr(NewSQLInt64(valKind, 9))},
		{"3 + '3'", "6", NewSQLValueExpr(NewSQLInt64(valKind, 6))},
		{"a + 0", "a", NewSQLColumnExpr(1, "test", "bar", "a",
			EvalInt64, schema.MongoInt, false)},
		{"a - 0", "a", NewSQLColumnExpr(1, "test", "bar", "a",
			EvalInt64, schema.MongoInt, false)},
		{"a * 1", "a", NewSQLColumnExpr(1, "test", "bar", "a",
			EvalInt64, schema.MongoInt, false)},
		{"a / 1", "a", NewSQLColumnExpr(1, "test", "bar", "a",
			EvalInt64, schema.MongoInt, false)},
		{"a / 0", "NULL", NewSQLValueExpr(NewSQLNull(valKind))},
		{"3 - '3'", "0", NewSQLValueExpr(NewSQLInt64(valKind, 0))},
		{"3 div '3'", "1", NewSQLValueExpr(NewSQLInt64(valKind, 1))},
		{"3 = '3'", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 <= '3'", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 >= '3'", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 < '3'", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"3 > '3'", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"3 <=> '3'", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 = a", "a = 3", NewSQLEqualsExpr(
			NewSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false),
			NewSQLValueExpr(NewSQLInt64(valKind, 3)),
		)},
		{"3 < a", "a > 3", NewSQLGreaterThanExpr(
			NewSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false),
			NewSQLValueExpr(NewSQLInt64(valKind, 3)),
		)},
		{"3 <= a", "a >= 3", NewSQLGreaterThanOrEqualExpr(
			NewSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false),
			NewSQLValueExpr(NewSQLInt64(valKind, 3)))},
		{"3 > a", "a < 3", NewSQLLessThanExpr(
			NewSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false),
			NewSQLValueExpr(NewSQLInt64(valKind, 3)),
		)},
		{"3 >= a", "a <= 3", NewSQLLessThanOrEqualExpr(
			NewSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false),
			NewSQLValueExpr(NewSQLInt64(valKind, 3)),
		)},
		{"3 <> a", "a <> 3", NewSQLNotEqualsExpr(
			NewSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false),
			NewSQLValueExpr(NewSQLInt64(valKind, 3)),
		)},
		{"3 + 3 = 6", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 <=> 3", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"NULL <=> 3", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"3 <=> NULL", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"NULL <=> NULL", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 / (3 - 2) = a", "a = 3", NewSQLEqualsExpr(
			NewSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false),
			NewSQLValueExpr(NewSQLFloat(valKind, 3)),
		)},
		{"3 + 3 = 6 AND 1 >= 1 AND 3 = a", "1 AND a = 3",
			NewSQLAndExpr(NewSQLValueExpr(NewSQLBool(valKind, true)),
				NewSQLEqualsExpr(NewSQLColumnExpr(
					1, "test", "bar", "a", EvalInt64,
					schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 3))))},
		{"3 / (3 - 2) = a AND 4 - 2 = b", "a = 3 AND b = 2",
			NewSQLAndExpr(
				NewSQLEqualsExpr(NewSQLColumnExpr(1, "test", "bar", "a",
					EvalInt64, schema.MongoInt, false), NewSQLValueExpr(NewSQLFloat(valKind, 3))),
				NewSQLEqualsExpr(NewSQLColumnExpr(1, "test", "bar", "b",
					EvalInt64, schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 2))))},
		{"3 + 3 = 6 OR a = 3", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 + 3 = 5 OR a = 3", "0 OR a = 3",
			NewSQLOrExpr(NewSQLValueExpr(NewSQLBool(valKind, false)),
				NewSQLEqualsExpr(NewSQLColumnExpr(
					1, "test", "bar", "a", EvalInt64,
					schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 3))))},
		{"0 OR NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"1 OR NULL", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"NULL OR NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"0 AND 6+1 = 6", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"3 + 3 = 5 AND a = 3", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"0 AND NULL", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"1 AND NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"1 AND 6+0 = 6", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 + 3 = 6 AND a = 3", "1 and a = 3",
			NewSQLAndExpr(NewSQLValueExpr(NewSQLBool(valKind, true)),
				NewSQLEqualsExpr(NewSQLColumnExpr(
					1, "test", "bar", "a", EvalInt64,
					schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 3))))},
		{"(3 + 3 = 5) XOR a = 3", "a = 3", NewSQLEqualsExpr(
			NewSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 3)))},
		{"(3 + 3 = 6) XOR a = 3", "a <> 3", NewSQLNotExpr(NewSQLEqualsExpr(
			NewSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 3))))},
		{"(13 + 9 > 6) XOR (a = 4)", "a <> 4", NewSQLNotExpr(
			NewSQLEqualsExpr(NewSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 4))))},
		{"(8 / 5 = 9) XOR (a = 5)", "a = 5", NewSQLEqualsExpr(
			NewSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 5)))},
		{"false XOR 23", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"true XOR 23", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"a = 23 XOR true", "a <> 23", NewSQLNotExpr(NewSQLEqualsExpr(
			NewSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 23))))},
		{"!3", "0", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"!NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"a = ~1", "a = 18446744073709551614", NewSQLEqualsExpr(
			NewSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false),
			NewSQLValueExpr(NewSQLUint64(valKind, uint64(18446744073709551614))))},
		{"a = ~2398238912332232323", "a = 16048505161377319292", NewSQLEqualsExpr(
			NewSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false),
			NewSQLValueExpr(NewSQLUint64(valKind, uint64(16048505161377319292))))},
		{"DAYNAME('2016-1-1')", "Friday", NewSQLValueExpr(NewSQLVarchar(valKind, "Friday"))},
		{"(8-7)", "1", NewSQLValueExpr(NewSQLInt64(valKind, 1))},
		{"a LIKE NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"4 LIKE NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"a = NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"a > NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"a >= NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"a < NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"a <= NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"a != NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"(1, 3) > (3, 4)", "SQLFalse", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"(4, 3) > (3, 4)", "SQLTrue", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"(4, 31) > (4, 4)", "SQLTrue", NewSQLValueExpr(NewSQLBool(valKind, true))},

		{"abs(NULL)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"abs(-10)", "10", NewSQLValueExpr(NewSQLFloat(valKind, 10))},
		{"ascii(NULL)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"ascii('a')", "97", NewSQLValueExpr(NewSQLInt64(valKind, 97))},
		{"char_length(NULL)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"character_length(NULL)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"concat(NULL, a)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"concat(a, NULL)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"concat('go', 'lang')", "golang", NewSQLValueExpr(NewSQLVarchar(valKind, "golang"))},
		{"concat_ws(NULL, a)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"convert(NULL, SIGNED)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"elt(NULL, 'a', 'b')", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"elt(4, 'a', 'b')", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"exp(NULL)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"exp(2)", "7.38905609893065", NewSQLValueExpr(NewSQLFloat(valKind, 7.38905609893065))},
		{"greatest(a, NULL)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"greatest(2, 3)", "3", NewSQLValueExpr(NewSQLInt64(valKind, 3))},
		{"ifnull(NULL, 10)", "10", NewSQLValueExpr(NewSQLInt64(valKind, 10))},
		{"ifnull(10, 1)", "10", NewSQLValueExpr(NewSQLInt64(valKind, 10))},
		{"interval(NULL, a)", "-1", NewSQLValueExpr(NewSQLInt64(valKind, -1))},
		{"interval(0, 1)", "0", NewSQLValueExpr(NewSQLInt64(valKind, 0))},
		{"interval(1, 2, 3, 4)", "1", NewSQLValueExpr(NewSQLInt64(valKind, 0))},
		{"interval(1, 1, 2, 3)", "1", NewSQLValueExpr(NewSQLInt64(valKind, 1))},
		{"interval(-1, NULL, NULL, -0.5, 3, 4)", "1", NewSQLValueExpr(NewSQLInt64(valKind, 2))},
		{"interval(-3.4, -4, -3.6, -3.4, -3, 1, 2)", "3", NewSQLValueExpr(NewSQLInt64(valKind, 3))},
		{"interval(8, -4, 0, 7, 8)", "4", NewSQLValueExpr(NewSQLInt64(valKind, 4))},
		{"interval(8, -3, 1, 7, 7)", "1", NewSQLValueExpr(NewSQLInt64(valKind, 4))},
		{"interval(7.7, -3, 1, 7, 7)", "1", NewSQLValueExpr(NewSQLInt64(valKind, 4))},
		{"least(a, NULL)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"least(2, 3)", "2", NewSQLValueExpr(NewSQLInt64(valKind, 2))},
		{"locate('bar', 'foobar', NULL)", "0", NewSQLValueExpr(NewSQLInt64(valKind, 0))},
		{"locate('bar', 'foobar')", "4", NewSQLValueExpr(NewSQLInt64(valKind, 4))},
		{"makedate(2000, NULL)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"makedate(NULL, 10)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"mid('foobar', NULL, 2)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"mod(10, 2)", "0", NewSQLValueExpr(NewSQLFloat(valKind, 0))},
		{"mod(NULL, 2)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"mod(10, NULL)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"nullif(1, 1)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"nullif(1, null)", "1", NewSQLValueExpr(NewSQLInt64(valKind, 1))},
		{"pow(a, NULL)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"pow(NULL, a)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"pow(2,2)", "4", NewSQLValueExpr(NewSQLFloat(valKind, 4))},
		{"round(NULL, 2)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"round(2, NULL)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"round(2, 2)", "2", NewSQLValueExpr(NewSQLFloat(valKind, 2))},
		{"repeat('a', NULL)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"repeat(NULL, 3)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"substring(NULL, 2)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"substring(NULL, 2, 3)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"substring('foobar', NULL)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"substring('foobar', NULL, 2)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"substring('foobar', 2, NULL)", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"substring('foobar', 2, 3)", "oob", NewSQLValueExpr(NewSQLVarchar(valKind, "oob"))},
		{"substring_index(NULL, 'o', 0)", "", NewSQLValueExpr(NewSQLNull(valKind))},
		{"substring_index('foobar', 'o', 0)", "", NewSQLValueExpr(NewSQLVarchar(valKind, ""))},
	}

	runTests(tests)
}
