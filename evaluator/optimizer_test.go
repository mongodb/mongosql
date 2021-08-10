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
        Name: dt_int
        MongoType: int 
        SqlType: int
     -
        Name: dt_str
        MongoType: string
        SqlType: varchar

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

				rCfg := NewRewriterConfig(42, "test_db", log.GlobalLogger(),
					false, evaluatorUnitTestVersion, "evaluator_unit_test_remoteHost", "evaluator_unit_test_user")

				rewritten, err := RewriteStatement(rCfg, statement)
				req.NoError(err, "failed to rewrite query")

				aCfg := createAlgebrizerCfg(defaultDbName, testSchemaCatalog, testVariables, false)
				plan, err := AlgebrizeQuery(bgCtx, aCfg, rewritten)

				req.NoError(err)

				version := []uint8{3, 6, 0}

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
		// should push down self-joins on sharded collections
		{
			sql: "select a._id, array from foo a join foo_array b on a._id = b._id",
			expected: []*ast.Pipeline{
				ast.NewPipeline(
					ast.NewUnwindStage(ast.NewFieldRef("array", nil), "", false),
					ast.NewProjectStage(
						ast.NewAssignProjectItem("test_DOT_a_DOT__id", ast.NewFieldRef("_id", nil)),
						ast.NewAssignProjectItem("test_DOT_b_DOT_array", ast.NewFieldRef("array", nil)),
						ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
					),
				),
			},
		},

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
						ast.NewBinary(ast.And, ast.NewAggExpr(ast.NewBinary(ast.GreaterThan, astutil.FieldRefFromFieldName("d.f"), astutil.NullLiteral)),
							ast.NewBinary(ast.NotEquals, ast.NewFieldRef("a", nil), astutil.NullLiteral)),
					),
					ast.NewLookupStage("bar", ast.NewFieldRef("a", nil), "a", "__joined_bar", nil, nil),
					ast.NewUnwindStage(ast.NewFieldRef("__joined_bar", nil), "", false),
					ast.NewMatchStage(
						ast.NewAggExpr(ast.NewBinary(ast.And, ast.NewBinary(ast.GreaterThan, astutil.FieldRefFromFieldName("__joined_bar.a"), astutil.NullLiteral),
							ast.NewBinary(ast.Equals, ast.NewFieldRef("__joined_bar.a", nil), astutil.FieldRefFromFieldName("d.f")))),
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
		{"3 * '3'", "9", NewSQLValueExpr(NewSQLFloat(valKind, 9))},
		{"3 + '3'", "6", NewSQLValueExpr(NewSQLFloat(valKind, 6))},
		{"a + 0", "a", testSQLColumnExpr(1, "test", "bar", "a",
			EvalInt64, schema.MongoInt, false)},
		{"a - 0", "a", testSQLColumnExpr(1, "test", "bar", "a",
			EvalInt64, schema.MongoInt, false)},
		{"a * 1", "a", testSQLColumnExpr(1, "test", "bar", "a",
			EvalInt64, schema.MongoInt, false)},
		{"a / 1", "a", testSQLColumnExpr(1, "test", "bar", "a",
			EvalInt64, schema.MongoInt, false)},
		{"a / 0", "NULL", NewSQLValueExpr(NewSQLNull(valKind))},
		{"3 - '3'", "0", NewSQLValueExpr(NewSQLFloat(valKind, 0))},
		{"3 div '3'", "1", NewSQLValueExpr(NewSQLInt64(valKind, 1))},
		{"3 = '3'", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 <= '3'", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 >= '3'", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 < '3'", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"3 > '3'", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"3 <=> '3'", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 >= a", "a <= 3", NewSQLComparisonExpr(LTE,
			testSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false),
			NewSQLValueExpr(NewSQLInt64(valKind, 3)),
		)},
		{"3 > a", "a < 3", NewSQLComparisonExpr(
			LT, testSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false),
			NewSQLValueExpr(NewSQLInt64(valKind, 3)),
		)},
		{"3 >= a", "a <= 3", NewSQLComparisonExpr(LTE,
			testSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false),
			NewSQLValueExpr(NewSQLInt64(valKind, 3)),
		)},
		{"3 + 3 = 6", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 <=> 3", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"NULL <=> 3", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"3 <=> NULL", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"NULL <=> NULL", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 / (3 - 2) = a", "3 = a", NewSQLComparisonExpr(
			EQ,
			NewSQLValueExpr(NewSQLFloat(valKind, 3)),
			testSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false),
		)},
		{"3 + 3 = 6 AND 1 >= 1 AND 3 = a", "a = 3", NewSQLComparisonExpr(
			EQ,
			NewSQLValueExpr(NewSQLInt64(valKind, 3)),
			testSQLColumnExpr(
				1, "test", "bar", "a", EvalInt64,
				schema.MongoInt, false),
		)},
		{"3 + 3 = 6 OR a = 3", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 + 3 = 5 OR a = 3", "a = 3", NewSQLComparisonExpr(
			EQ,
			testSQLColumnExpr(
				1, "test", "bar", "a", EvalInt64,
				schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 3)))},
		{"0 OR NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"1 OR NULL", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"NULL OR NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"0 AND 6+1 = 6", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"3 + 3 = 5 AND a = 3", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"0 AND NULL", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"1 AND NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"1 AND 6+0 = 6", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"3 + 3 = 6 AND a = 3", "1 and a = 3", NewSQLComparisonExpr(
			EQ,
			testSQLColumnExpr(
				1, "test", "bar", "a", EvalInt64,
				schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 3)))},
		{"(3 + 3 = 5) XOR a = 3", "a = 3", NewSQLComparisonExpr(EQ,
			testSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 3)))},
		{"(3 + 3 = 6) XOR a = 3", "a <> 3", NewSQLNotExpr(NewSQLComparisonExpr(EQ,
			testSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 3))))},
		{"(13 + 9 > 6) XOR (a = 4)", "a <> 4", NewSQLNotExpr(
			NewSQLComparisonExpr(EQ, testSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 4))))},
		{"(8 / 5 = 9) XOR (a = 5)", "a = 5", NewSQLComparisonExpr(EQ,
			testSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 5)))},
		{"false XOR 23", "true", NewSQLValueExpr(NewSQLBool(valKind, true))},
		{"true XOR 23", "false", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"a = 23 XOR true", "a <> 23", NewSQLNotExpr(NewSQLComparisonExpr(EQ,
			testSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false), NewSQLValueExpr(NewSQLInt64(valKind, 23))))},
		{"!3", "0", NewSQLValueExpr(NewSQLBool(valKind, false))},
		{"!NULL", "null", NewSQLValueExpr(NewSQLNull(valKind))},
		{"a = ~1", "a = 18446744073709551614", NewSQLComparisonExpr(EQ,
			testSQLColumnExpr(1, "test", "bar", "a",
				EvalInt64, schema.MongoInt, false),
			NewSQLValueExpr(NewSQLUint64(valKind, uint64(18446744073709551614))))},
		{"a = ~2398238912332232323", "a = 16048505161377319292", NewSQLComparisonExpr(EQ,
			testSQLColumnExpr(1, "test", "bar", "a",
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
