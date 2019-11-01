package evaluator_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/stretchr/testify/require"
)

var (
	testSchema                = evaluator.MustLoadSchema(testSchema1)
	testInfo                  = evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)
	testVars                  = evaluator.CreateTestVariables(testInfo)
	testCatalog               = evaluator.GetCatalog(testSchema, testVars, testInfo)
	defaultDbName             = "test"
	algebrizerUnitTestVersion = "5.7.12"
)

func setSystemVariable(vars *variable.Container, name variable.Name, value interface{}) {
	var sqlValue values.SQLValue
	switch typedV := value.(type) {
	case string:
		sqlValue = values.NewSQLVarchar(values.MongoSQLValueKind, typedV)
	case int:
		sqlValue = values.NewSQLInt64(values.MongoSQLValueKind, int64(typedV))
	case int64:
		sqlValue = values.NewSQLInt64(values.MongoSQLValueKind, typedV)
	case uint64:
		sqlValue = values.NewSQLUint64(values.MongoSQLValueKind, typedV)
	default:
		panic(fmt.Sprintf("unsupported type: %T for algebrizer_test system variables", value))
	}
	vars.SetSystemVariable(name, sqlValue)
}

const (
	// foo.bar is the default test namespace for a dual source
	// for this test schema (testSchema1).
	defaultDualDbName    = "foo"
	defaultDualTableName = "bar"
)

func createMongoSource(selectID int, tableName, aliasName string) evaluator.PlanStage {
	db, _ := testCatalog.Database(defaultDbName)
	tbl, _ := db.Table(tableName)
	table, _ := tbl.(catalog.MongoDBTable)
	return evaluator.NewMongoSourceStage(db, table, selectID, aliasName)
}

func createMongoDualSource(selectID int) evaluator.PlanStage {
	db, _ := testCatalog.Database(defaultDualDbName)
	tbl, _ := db.Table(defaultDualTableName)
	table, _ := tbl.(catalog.MongoDBTable)
	return evaluator.NewMongoSourceDualStage(db, table, selectID, defaultDualTableName)
}

func TestAlgebrizeQuery(t *testing.T) {
	setSystemVariable(testVars, variable.MongoDBMaxVarcharLength, 10)
	setSystemVariable(testVars, variable.TypeConversionMode, "mysql")

	type planTest struct {
		sql                 string
		expectedPlanFactory func() evaluator.PlanStage
	}

	runPlanTest := func(t *testing.T, testCase planTest) {
		t.Run(testCase.sql, func(t *testing.T) {
			req := require.New(t)

			statement, err := parser.Parse(testCase.sql)
			req.Nil(err, fmt.Sprintf("failed to parse: %s", testCase.sql))

			rCfg := evaluator.NewRewriterConfig(42, "algebrizer_unit_test_dbname", log.GlobalLogger(),
				false, algebrizerUnitTestVersion, "algebrizer_unit_test_remoteHost", "algebrizer_unit_test_user")

			rewritten, err := evaluator.RewriteStatement(rCfg, statement)
			req.Nil(err, "failed to rewrite query")

			aCfg := evaluator.NewAlgebrizerConfig(log.GlobalLogger(), defaultDbName, testCatalog, false)

			actual, err := evaluator.AlgebrizeQuery(aCfg, rewritten)
			req.Nil(err, "failed to algebrize")

			expected := testCase.expectedPlanFactory()
			req.Equal(expected, actual, "actual does not match expected")
		})
	}

	type variableTest struct {
		sql                 string
		container           func() *variable.Container
		info                func() *mongodb.Info
		expectedPlanFactory func() evaluator.PlanStage
	}

	runVariablesTest := func(t *testing.T, testCase variableTest) {
		t.Run(testCase.sql, func(t *testing.T) {
			req := require.New(t)
			statement, err := parser.Parse(testCase.sql)
			req.Nil(err, fmt.Sprintf("failed to parse: %s", testCase.sql))

			rCfg := evaluator.NewRewriterConfig(42, "algebrizer_unit_test_dbname", log.GlobalLogger(),
				false, algebrizerUnitTestVersion, "algebrizer_unit_test_remoteHost", "algebrizer_unit_test_user")

			rewritten, err := evaluator.RewriteStatement(rCfg, statement)
			req.Nil(err, "failed to rewrite query")

			// Rebuild Catalog with new variables.
			cat := evaluator.GetCatalog(testSchema, testCase.container(), testCase.info())

			aCfg := evaluator.NewAlgebrizerConfig(log.GlobalLogger(), defaultDbName, cat, false)

			actual, err := evaluator.AlgebrizeQuery(aCfg, rewritten)
			req.Nil(err, "failed to algebrize")

			expected := testCase.expectedPlanFactory()

			expected = unsetUnimportantFields(expected)
			actual = unsetUnimportantFields(actual)
			req.Equal(expected, actual, "actual does not match expected")
		})
	}

	type errorTest struct {
		sql     string
		message string
	}

	runErrorTest := func(t *testing.T, testCase errorTest) {
		t.Run(testCase.sql, func(t *testing.T) {
			req := require.New(t)
			statement, err := parser.Parse(testCase.sql)
			req.Nil(err, fmt.Sprintf("failed to parse: %s", testCase.sql))

			rCfg := evaluator.NewRewriterConfig(42, "algebrizer_unit_test_dbname", log.GlobalLogger(),
				false, algebrizerUnitTestVersion, "algebrizer_unit_test_remoteHost", "algebrizer_unit_test_user")

			rewritten, err := evaluator.RewriteStatement(rCfg, statement)
			req.Nil(err, "failed to rewrite query")

			aCfg := evaluator.NewAlgebrizerConfig(log.GlobalLogger(), defaultDbName, testCatalog, false)

			_, err = evaluator.AlgebrizeQuery(aCfg, rewritten)

			req.NotNil(err, "succeeded to algebrize in expected error test")

			req.Equal(err.Error(), testCase.message, "actual does not match expected")
		})
	}

	runTestsAsSubtest := func(subTestName string, tests interface{}) {
		t.Run(subTestName, func(t *testing.T) {
			switch typedTests := tests.(type) {
			case []planTest:
				for _, testCase := range typedTests {
					runPlanTest(t, testCase)
				}
			case []variableTest:
				for _, testCase := range typedTests {
					runVariablesTest(t, testCase)
				}
			case []errorTest:
				for _, testCase := range typedTests {
					runErrorTest(t, testCase)
				}
			}
		})
	}

	// Show Statements
	req := require.New(t)
	// Show Charsets
	isDBName := "information_schema"
	informationSchemaDB, _ := testCatalog.Database(isDBName)
	subqueryAliasName := "CHARACTER_SETS"
	tbl, err := informationSchemaDB.Table(subqueryAliasName)
	req.Nil(err, "failed to load table")
	source := evaluator.NewDynamicSourceStage(
		informationSchemaDB,
		tbl.(*catalog.DynamicTable),
		2,
		subqueryAliasName)
	subquery := evaluator.NewSubquerySourceStage(
		evaluator.NewProjectStage(
			source,
			createProjectedColumn(1,
				source,
				subqueryAliasName,
				"CHARACTER_SET_NAME",
				subqueryAliasName,
				"Charset",
				false),
			createProjectedColumn(3,
				source,
				subqueryAliasName,
				"DESCRIPTION",
				subqueryAliasName,
				"Description",
				false),
			createProjectedColumn(2,
				source,
				subqueryAliasName,
				"DEFAULT_COLLATE_NAME",
				subqueryAliasName,
				"Default collation",
				false),
			createProjectedColumn(4,
				source,
				subqueryAliasName,
				"MAXLEN",
				subqueryAliasName,
				"Maxlen",
				false),
		),
		2,
		isDBName,
		subqueryAliasName,
		false,
	)
	showCharsetPlanTests := []planTest{{
		"show charset",
		func() evaluator.PlanStage {
			return evaluator.NewOrderByStage(
				subquery,
				evaluator.NewOrderByTerm(
					createSQLColumnExprFromSource(
						subquery,
						subqueryAliasName,
						"Charset",
						false), true),
			)
		}}, {

		"show charset like 'n'",
		func() evaluator.PlanStage {
			return evaluator.NewOrderByStage(
				evaluator.NewFilterStage(
					subquery,
					evaluator.NewSQLLikeExpr(
						createSQLColumnExprFromSource(
							subquery,
							subqueryAliasName,
							"Charset",
							false),
						evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n")),
						evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "\\")),
						false,
					),
				),
				evaluator.NewOrderByTerm(
					createSQLColumnExprFromSource(
						subquery,
						subqueryAliasName,
						"Charset",
						false),
					true),
			)
		}}, {

		"show charset where `Charset` = 'n'", func() evaluator.PlanStage {
			return evaluator.NewOrderByStage(
				evaluator.NewFilterStage(
					subquery,
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(
							subquery,
							subqueryAliasName,
							"Charset",
							false),
						evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n"))),
				),
				evaluator.NewOrderByTerm(
					createSQLColumnExprFromSource(
						subquery,
						subqueryAliasName,
						"Charset",
						false),
					true),
			)
		}},
	}
	runTestsAsSubtest("Show Charset", showCharsetPlanTests)

	// Show Collation
	subqueryAliasName = "COLLATIONS"
	tbl, err = informationSchemaDB.Table(subqueryAliasName)
	req.Nil(err, "failed to load table")
	source = evaluator.NewDynamicSourceStage(informationSchemaDB,
		tbl.(*catalog.DynamicTable),
		2,
		subqueryAliasName)
	subquery = evaluator.NewSubquerySourceStage(
		evaluator.NewProjectStage(
			source,
			createProjectedColumn(1,
				source,
				subqueryAliasName,
				"COLLATION_NAME",
				subqueryAliasName,
				"Collation",
				false),

			createProjectedColumn(2,
				source,
				subqueryAliasName,
				"CHARACTER_SET_NAME",
				subqueryAliasName,
				"Charset",
				false),

			createProjectedColumn(3,
				source,
				subqueryAliasName,
				"ID",
				subqueryAliasName,
				"Id",
				false),

			createProjectedColumn(4,
				source,
				subqueryAliasName,
				"IS_DEFAULT",
				subqueryAliasName,
				"Default",
				false),

			createProjectedColumn(5,
				source,
				subqueryAliasName,
				"IS_COMPILED",
				subqueryAliasName,
				"Compiled",
				false),

			createProjectedColumn(6,
				source,
				subqueryAliasName,
				"SORTLEN",
				subqueryAliasName,
				"Sortlen",
				false),
		),
		2,
		isDBName,
		subqueryAliasName,
		false,
	)
	showCollationPlanTests := []planTest{{
		"show collation",
		func() evaluator.PlanStage {
			return evaluator.NewOrderByStage(
				subquery,
				evaluator.NewOrderByTerm(createSQLColumnExprFromSource(subquery,
					subqueryAliasName,
					"Collation",
					false),
					true),
			)
		}}, {

		"show collation like 'n'",
		func() evaluator.PlanStage {
			return evaluator.NewOrderByStage(
				evaluator.NewFilterStage(
					subquery,
					evaluator.NewSQLLikeExpr(createSQLColumnExprFromSource(subquery,
						subqueryAliasName,
						"Collation",
						false),
						evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n")),
						evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "\\")),
						false,
					),
				),

				evaluator.NewOrderByTerm(createSQLColumnExprFromSource(subquery,
					subqueryAliasName,
					"Collation",
					false),
					true),
			)
		}}, {

		"show collation where `Collation` = 'n'",
		func() evaluator.PlanStage {
			return evaluator.NewOrderByStage(
				evaluator.NewFilterStage(
					subquery,
					evaluator.NewSQLComparisonExpr(evaluator.EQ,
						createSQLColumnExprFromSource(subquery,
							subqueryAliasName,
							"Collation",
							false),
						evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n"))),
				),

				evaluator.NewOrderByTerm(createSQLColumnExprFromSource(subquery,
					subqueryAliasName,
					"Collation",
					false),
					true),
			)
		}},
	}
	runTestsAsSubtest("Show Collation", showCollationPlanTests)

	// Show Columns
	subqueryAliasName = "COLUMNS"
	tbl, err = informationSchemaDB.Table(subqueryAliasName)
	req.Nil(err, "failed to load table")
	// Show Columns Plain
	source = evaluator.NewDynamicSourceStage(
		informationSchemaDB,
		tbl.(*catalog.DynamicTable),
		2,
		subqueryAliasName)

	subquery = evaluator.NewSubquerySourceStage(
		evaluator.NewProjectStage(
			source,
			createProjectedColumn(4,
				source,
				subqueryAliasName,
				"COLUMN_NAME",
				subqueryAliasName,
				"Field",
				false),
			createProjectedColumn(16,
				source,
				subqueryAliasName,
				"COLUMN_TYPE",
				subqueryAliasName,
				"Type",
				false),
			createProjectedColumn(7,
				source,
				subqueryAliasName,
				"IS_NULLABLE",
				subqueryAliasName,
				"Null",
				false),
			createProjectedColumn(17,
				source,
				subqueryAliasName,
				"COLUMN_KEY",
				subqueryAliasName,
				"Key",
				false),
			createProjectedColumn(6,
				source,
				subqueryAliasName,
				"COLUMN_DEFAULT",
				subqueryAliasName,
				"Default",
				false),
			createProjectedColumn(18,
				source,
				subqueryAliasName,
				"EXTRA",
				subqueryAliasName,
				"Extra",
				false),
			createProjectedColumn(3,
				source,
				subqueryAliasName,
				"TABLE_NAME",
				subqueryAliasName,
				"TABLE_NAME",
				false),
			createProjectedColumn(2,
				source,
				subqueryAliasName,
				"TABLE_SCHEMA",
				subqueryAliasName,
				"TABLE_SCHEMA",
				false),
			createProjectedColumn(5,
				source,
				subqueryAliasName,
				"ORDINAL_POSITION",
				subqueryAliasName,
				"ORDINAL_POSITION",
				false),
		),
		2,
		isDBName,
		subqueryAliasName,
		false,
	)
	commonProjectedColumns := []evaluator.ProjectedColumn{
		createProjectedColumn(1, subquery,
			subqueryAliasName, "Field", subqueryAliasName, "Field", false),
		createProjectedColumn(1, subquery,
			subqueryAliasName, "Type", subqueryAliasName, "Type", false),
		createProjectedColumn(1, subquery,
			subqueryAliasName, "Null", subqueryAliasName, "Null", false),
		createProjectedColumn(1, subquery,
			subqueryAliasName, "Key", subqueryAliasName, "Key", false),
		createProjectedColumn(1, subquery,
			subqueryAliasName, "Default", subqueryAliasName, "Default", false),
		createProjectedColumn(1, subquery,
			subqueryAliasName, "Extra", subqueryAliasName, "Extra", false),
	}
	genPlainColumnTests := func(fromTemplateArg string) []planTest {
		return []planTest{{
			fmt.Sprintf("show columns %s", fromTemplateArg),
			func() evaluator.PlanStage {
				return evaluator.NewProjectStage(
					evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLComparisonExpr(
									evaluator.EQ,
									createSQLColumnExprFromSource(
										subquery,
										subqueryAliasName,
										"TABLE_NAME",
										false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "foo"))),
								evaluator.NewSQLComparisonExpr(
									evaluator.EQ,
									createSQLColumnExprFromSource(
										subquery,
										subqueryAliasName,
										"TABLE_SCHEMA",
										false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, defaultDbName))),
							),
						),
						evaluator.NewOrderByTerm(
							createSQLColumnExprFromSource(
								subquery,
								subqueryAliasName,
								"ORDINAL_POSITION",
								false),
							true),
					),
					commonProjectedColumns...,
				)
			}}, {

			fmt.Sprintf("show columns %s like 'n'", fromTemplateArg),
			func() evaluator.PlanStage {
				return evaluator.NewProjectStage(
					evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLAndExpr(
									evaluator.NewSQLComparisonExpr(
										evaluator.EQ,
										createSQLColumnExprFromSource(
											subquery,
											subqueryAliasName,
											"TABLE_NAME",
											false),
										evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "foo")),
									),
									evaluator.NewSQLComparisonExpr(
										evaluator.EQ,
										createSQLColumnExprFromSource(
											subquery,
											subqueryAliasName,
											"TABLE_SCHEMA",
											false),
										evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, defaultDbName)),
									),
								),
								evaluator.NewSQLLikeExpr(
									createSQLColumnExprFromSource(
										subquery,
										subqueryAliasName,
										"Field",
										false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n")),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "\\")),
									false,
								),
							),
						),
						evaluator.NewOrderByTerm(
							createSQLColumnExprFromSource(
								subquery,
								subqueryAliasName,
								"ORDINAL_POSITION",
								false),
							true,
						),
					),
					commonProjectedColumns...,
				)
			}}, {
			fmt.Sprintf("show columns %s where `Field` = 'n'", fromTemplateArg),
			func() evaluator.PlanStage {
				return evaluator.NewProjectStage(
					evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLAndExpr(
									evaluator.NewSQLComparisonExpr(
										evaluator.EQ,
										createSQLColumnExprFromSource(
											subquery,
											subqueryAliasName,
											"TABLE_NAME",
											false),
										evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "foo")),
									),
									evaluator.NewSQLComparisonExpr(
										evaluator.EQ,
										createSQLColumnExprFromSource(
											subquery,
											subqueryAliasName,
											"TABLE_SCHEMA",
											false),
										evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, defaultDbName)),
									),
								),
								evaluator.NewSQLComparisonExpr(
									evaluator.EQ,
									createSQLColumnExprFromSource(
										subquery,
										subqueryAliasName,
										"Field",
										false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n")),
								),
							),
						),
						evaluator.NewOrderByTerm(
							createSQLColumnExprFromSource(
								subquery,
								subqueryAliasName,
								"ORDINAL_POSITION",
								false),
							true,
						),
					),
					commonProjectedColumns...,
				)
			},
		},
		}
	}

	for _, from := range []string{"from foo", "from test.foo", "from foo from test",
		"in foo in test", "from foo in test", "in foo from test"} {
		plainColumnTests := genPlainColumnTests(from)
		runTestsAsSubtest(
			fmt.Sprintf(
				"Show Plain Columns 'show columns %s'",
				from),
			plainColumnTests)
	}

	// Show Columns Full
	subquery = evaluator.NewSubquerySourceStage(
		evaluator.NewProjectStage(
			source,
			createProjectedColumn(4,
				source,
				subqueryAliasName,
				"COLUMN_NAME",
				subqueryAliasName,
				"Field",
				false),
			createProjectedColumn(16,
				source,
				subqueryAliasName,
				"COLUMN_TYPE",
				subqueryAliasName,
				"Type",
				false),
			createProjectedColumn(15,
				source,
				subqueryAliasName,
				"COLLATION_NAME",
				subqueryAliasName,
				"Collation",
				false),
			createProjectedColumn(7,
				source,
				subqueryAliasName,
				"IS_NULLABLE",
				subqueryAliasName,
				"Null",
				false),
			createProjectedColumn(17,
				source,
				subqueryAliasName,
				"COLUMN_KEY",
				subqueryAliasName,
				"Key",
				false),
			createProjectedColumn(6,
				source,
				subqueryAliasName,
				"COLUMN_DEFAULT",
				subqueryAliasName,
				"Default",
				false),
			createProjectedColumn(18,
				source,
				subqueryAliasName,
				"EXTRA",
				subqueryAliasName,
				"Extra",
				false),
			createProjectedColumn(19,
				source,
				subqueryAliasName,
				"PRIVILEGES",
				subqueryAliasName,
				"Privileges",
				false),
			createProjectedColumn(20,
				source,
				subqueryAliasName,
				"COLUMN_COMMENT",
				subqueryAliasName,
				"Comment",
				false),
			createProjectedColumn(3,
				source,
				subqueryAliasName,
				"TABLE_NAME",
				subqueryAliasName,
				"TABLE_NAME",
				false),
			createProjectedColumn(2,
				source,
				subqueryAliasName,
				"TABLE_SCHEMA",
				subqueryAliasName,
				"TABLE_SCHEMA",
				false),
			createProjectedColumn(5,
				source,
				subqueryAliasName,
				"ORDINAL_POSITION",
				subqueryAliasName,
				"ORDINAL_POSITION",
				false),
		),
		2,
		isDBName,
		subqueryAliasName,
		false,
	)
	commonProjectedColumns = []evaluator.ProjectedColumn{
		createProjectedColumn(1, subquery,
			subqueryAliasName, "Field", subqueryAliasName, "Field", false),
		createProjectedColumn(1, subquery,
			subqueryAliasName, "Type", subqueryAliasName, "Type", false),
		createProjectedColumn(1, subquery,
			subqueryAliasName, "Collation", subqueryAliasName, "Collation", false),
		createProjectedColumn(1, subquery,
			subqueryAliasName, "Null", subqueryAliasName, "Null", false),
		createProjectedColumn(1, subquery,
			subqueryAliasName, "Key", subqueryAliasName, "Key", false),
		createProjectedColumn(1, subquery,
			subqueryAliasName, "Default", subqueryAliasName, "Default", false),
		createProjectedColumn(1, subquery,
			subqueryAliasName, "Extra", subqueryAliasName, "Extra", false),
		createProjectedColumn(1, subquery,
			subqueryAliasName, "Privileges", subqueryAliasName, "Privileges", false),
		createProjectedColumn(1, subquery,
			subqueryAliasName, "Comment", subqueryAliasName, "Comment", false),
	}
	genFullColumnTests := func(fromTemplateArg string) []planTest {
		return []planTest{{
			fmt.Sprintf("show full columns %s", fromTemplateArg),
			func() evaluator.PlanStage {
				return evaluator.NewProjectStage(
					evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLComparisonExpr(
									evaluator.EQ,
									createSQLColumnExprFromSource(subquery,
										subqueryAliasName,
										"TABLE_NAME",
										false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "foo")),
								),
								evaluator.NewSQLComparisonExpr(
									evaluator.EQ,
									createSQLColumnExprFromSource(subquery,
										subqueryAliasName,
										"TABLE_SCHEMA",
										false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, defaultDbName)),
								),
							),
						),
						evaluator.NewOrderByTerm(
							createSQLColumnExprFromSource(subquery,
								subqueryAliasName,
								"ORDINAL_POSITION",
								false),
							true,
						),
					),
					commonProjectedColumns...,
				)
			}}, {
			fmt.Sprintf("show full columns %s like 'n'", fromTemplateArg),
			func() evaluator.PlanStage {
				return evaluator.NewProjectStage(
					evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLAndExpr(
									evaluator.NewSQLComparisonExpr(
										evaluator.EQ,
										createSQLColumnExprFromSource(subquery,
											subqueryAliasName, "TABLE_NAME", false),
										evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "foo")),
									),
									evaluator.NewSQLComparisonExpr(
										evaluator.EQ,
										createSQLColumnExprFromSource(subquery,
											subqueryAliasName, "TABLE_SCHEMA", false),
										evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, defaultDbName)),
									),
								),
								evaluator.NewSQLLikeExpr(
									createSQLColumnExprFromSource(subquery,
										subqueryAliasName, "Field", false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n")),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "\\")),
									false,
								),
							),
						),
						evaluator.NewOrderByTerm(
							createSQLColumnExprFromSource(subquery,
								subqueryAliasName, "ORDINAL_POSITION", false),
							true,
						),
					),
					commonProjectedColumns...,
				)
			}}, {
			fmt.Sprintf("show full columns %s where `Field` = 'n'", fromTemplateArg),
			func() evaluator.PlanStage {
				return evaluator.NewProjectStage(
					evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLAndExpr(
									evaluator.NewSQLComparisonExpr(
										evaluator.EQ,
										createSQLColumnExprFromSource(subquery,
											subqueryAliasName, "TABLE_NAME", false),
										evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "foo")),
									),
									evaluator.NewSQLComparisonExpr(
										evaluator.EQ,
										createSQLColumnExprFromSource(subquery,
											subqueryAliasName, "TABLE_SCHEMA", false),
										evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, defaultDbName)),
									),
								),
								evaluator.NewSQLComparisonExpr(evaluator.EQ,
									createSQLColumnExprFromSource(subquery,
										subqueryAliasName, "Field", false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n")),
								),
							),
						),
						evaluator.NewOrderByTerm(
							createSQLColumnExprFromSource(subquery,
								subqueryAliasName, "ORDINAL_POSITION", false),
							true,
						),
					),
					commonProjectedColumns...,
				)
			}},
		}
	}

	for _, from := range []string{"from foo", "from test.foo", "from foo from test",
		"in foo in test", "from foo in test", "in foo from test"} {
		plainColumnTests := genFullColumnTests(from)
		runTestsAsSubtest(
			fmt.Sprintf(
				"Show Full Columns 'show columns %s'",
				from),
			plainColumnTests)
	}

	// Show Create Database
	dbName := "test"
	createDatabaseTests := []planTest{{
		"show create database " + dbName, func() evaluator.PlanStage {
			return evaluator.NewProjectStage(
				evaluator.NewDualStage(),
				evaluator.ProjectedColumn{
					Column: &results.Column{
						SelectID:   1,
						Name:       "Database",
						ColumnType: results.NewColumnType(types.EvalString, schema.MongoNone),
						Nullable:   true,
					},
					Expr: evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, dbName)),
				},
				evaluator.ProjectedColumn{
					Column: &results.Column{
						SelectID:   1,
						Name:       "Create Database",
						ColumnType: results.NewColumnType(types.EvalString, schema.MongoNone),
						Nullable:   true,
					},
					Expr: evaluator.NewSQLValueExpr(values.NewSQLVarchar(
						valKind, catalog.GenerateCreateDatabase(dbName, ""),
					)),
				},
			)
		}}, {

		"show create database if not exists " + dbName, func() evaluator.PlanStage {
			return evaluator.NewProjectStage(
				evaluator.NewDualStage(),
				evaluator.ProjectedColumn{
					Column: &results.Column{
						SelectID:   1,
						Name:       "Database",
						ColumnType: results.NewColumnType(types.EvalString, schema.MongoNone),
						Nullable:   true,
					},
					Expr: evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, dbName)),
				},
				evaluator.ProjectedColumn{
					Column: &results.Column{
						SelectID:   1,
						Name:       "Create Database",
						ColumnType: results.NewColumnType(types.EvalString, schema.MongoNone),
						Nullable:   true,
					},
					Expr: evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind,
						catalog.GenerateCreateDatabase(dbName, "IF NOT EXISTS"))),
				},
			)
		}},
	}
	runTestsAsSubtest("Show Create Database", createDatabaseTests)
	testDB, _ := testCatalog.Database("test")
	tbl, _ = testDB.Table("foo")

	createTableSQL := catalog.GenerateCreateTable(tbl, 10)

	// Show Create Table
	showCreateTableTests := []planTest{{
		"show create table foo",
		func() evaluator.PlanStage {
			return evaluator.NewProjectStage(
				evaluator.NewDualStage(),
				evaluator.ProjectedColumn{
					Column: &results.Column{
						SelectID: 1,
						Name:     "Table",
						ColumnType: results.NewColumnType(types.EvalString,
							schema.MongoNone),
						Nullable: true,
					},
					Expr: evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, tbl.Name())),
				},
				evaluator.ProjectedColumn{
					Column: &results.Column{
						SelectID: 1,
						Name:     "Create Table",
						ColumnType: results.NewColumnType(types.EvalString,
							schema.MongoNone),
						Nullable: true,
					},
					Expr: evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, createTableSQL)),
				},
			)
		}}, {
		"show create table .foo",
		func() evaluator.PlanStage {
			return evaluator.NewProjectStage(
				evaluator.NewDualStage(),
				evaluator.ProjectedColumn{
					Column: &results.Column{
						SelectID: 1,
						Name:     "Table",
						ColumnType: results.NewColumnType(types.EvalString,
							schema.MongoNone),
						Nullable: true,
					},
					Expr: evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, tbl.Name())),
				},
				evaluator.ProjectedColumn{
					Column: &results.Column{
						SelectID: 1,
						Name:     "Create Table",
						ColumnType: results.NewColumnType(types.EvalString,
							schema.MongoNone),
						Nullable: true,
					},
					Expr: evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, createTableSQL)),
				},
			)
		}}, {
		"show create table test.foo",
		func() evaluator.PlanStage {
			return evaluator.NewProjectStage(
				evaluator.NewDualStage(),
				evaluator.ProjectedColumn{
					Column: &results.Column{
						SelectID: 1,
						Name:     "Table",
						ColumnType: results.NewColumnType(types.EvalString,
							schema.MongoNone),
						Nullable: true,
					},
					Expr: evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, tbl.Name())),
				},
				evaluator.ProjectedColumn{
					Column: &results.Column{
						SelectID: 1,
						Name:     "Create Table",
						ColumnType: results.NewColumnType(types.EvalString,
							schema.MongoNone),
						Nullable: true,
					},
					Expr: evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, createTableSQL)),
				},
			)
		}},
	}
	runTestsAsSubtest("Show Create Table", showCreateTableTests)

	// Show Databases and Schemas

	subqueryAliasName = "SCHEMATA"
	tbl, err = informationSchemaDB.Table(subqueryAliasName)
	req.Nil(err)
	source = evaluator.NewDynamicSourceStage(informationSchemaDB,
		tbl.(*catalog.DynamicTable), 2, subqueryAliasName)
	subquery = evaluator.NewSubquerySourceStage(
		evaluator.NewProjectStage(
			source,
			createProjectedColumn(2, source,
				subqueryAliasName, "SCHEMA_NAME", subqueryAliasName, "Database", false),
		),
		2,
		isDBName,
		subqueryAliasName,
		false,
	)
	genShowDatabaseOrSchemaTests := func(name string) []planTest {
		return []planTest{{
			fmt.Sprintf("show %s", name),
			func() evaluator.PlanStage {
				return evaluator.NewOrderByStage(
					subquery,
					evaluator.NewOrderByTerm(
						createSQLColumnExprFromSource(
							subquery, subqueryAliasName, "Database", false),
						true,
					),
				)
			}}, {
			fmt.Sprintf("show %s like 'n'", name),
			func() evaluator.PlanStage {
				return evaluator.NewOrderByStage(
					evaluator.NewFilterStage(
						subquery,
						evaluator.NewSQLLikeExpr(
							createSQLColumnExprFromSource(subquery, subqueryAliasName, "Database", false),
							evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n")),
							evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "\\")),
							false,
						),
					),
					evaluator.NewOrderByTerm(
						createSQLColumnExprFromSource(subquery, subqueryAliasName, "Database", false),
						true,
					),
				)
			}}, {
			fmt.Sprintf("show %s where `Database` = 'n'", name),
			func() evaluator.PlanStage {
				return evaluator.NewOrderByStage(
					evaluator.NewFilterStage(
						subquery,
						evaluator.NewSQLComparisonExpr(
							evaluator.EQ,
							createSQLColumnExprFromSource(subquery, subqueryAliasName, "Database", false),
							evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n")),
						),
					),
					evaluator.NewOrderByTerm(
						createSQLColumnExprFromSource(subquery, subqueryAliasName, "Database", false),
						true,
					),
				)
			}},
		}
	}
	for _, name := range []string{"databases", "schemas"} {
		runTestsAsSubtest(fmt.Sprintf("Show %s", name), genShowDatabaseOrSchemaTests(name))
	}

	// Show Status/Variables
	for _, kind := range []string{"status", "variables"} {
		for _, scope := range []string{"", "global", "session"} {
			actualScope := scope
			if actualScope == "" {
				actualScope = "session"
			}

			if actualScope == "global" {
				subqueryAliasName = "GLOBAL_"
			} else {
				subqueryAliasName = "SESSION_"
			}

			if kind == "status" {
				subqueryAliasName += "STATUS"
			} else {
				subqueryAliasName += "VARIABLES"
			}

			varTbl, varErr := informationSchemaDB.Table(fmt.Sprintf("%s_%s", actualScope, kind))
			req.Nil(varErr)
			actualTableName := varTbl.Name()
			varSource := evaluator.NewDynamicSourceStage(informationSchemaDB,
				varTbl.(*catalog.DynamicTable), 2, actualTableName)
			varSubquery := evaluator.NewSubquerySourceStage(
				evaluator.NewProjectStage(
					varSource,
					createProjectedColumn(1, varSource, actualTableName, "VARIABLE_NAME",
						actualTableName, "Variable_name", false),
					createProjectedColumn(2, varSource, actualTableName, "VARIABLE_VALUE",
						actualTableName, "Value", false),
				),
				2,
				isDBName,
				subqueryAliasName,
				false,
			)
			showName := strings.TrimSpace(scope + " " + kind)

			tests := []planTest{{
				fmt.Sprintf("show %s", showName),
				func() evaluator.PlanStage {
					return evaluator.NewOrderByStage(
						varSubquery,
						evaluator.NewOrderByTerm(
							createSQLColumnExprFromSource(varSubquery,
								subqueryAliasName, "Variable_name", false),
							true,
						),
					)
				}}, {
				fmt.Sprintf("show %s like 'n'", showName),
				func() evaluator.PlanStage {
					return evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							varSubquery,
							evaluator.NewSQLLikeExpr(
								createSQLColumnExprFromSource(varSubquery,
									subqueryAliasName, "Variable_name", false),
								evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n")),
								evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "\\")),
								false,
							),
						),
						evaluator.NewOrderByTerm(
							createSQLColumnExprFromSource(varSubquery,
								subqueryAliasName, "Variable_name", false),
							true,
						),
					)
				}}, {
				fmt.Sprintf("show %s where Variable_name = 'n'", showName),
				func() evaluator.PlanStage {
					return evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							varSubquery,
							evaluator.NewSQLComparisonExpr(
								evaluator.EQ,
								createSQLColumnExprFromSource(varSubquery,
									subqueryAliasName, "Variable_name", false),
								evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n")),
							),
						),
						evaluator.NewOrderByTerm(
							createSQLColumnExprFromSource(varSubquery,
								subqueryAliasName, "Variable_name", false),
							true,
						),
					)
				}},
			}
			runTestsAsSubtest(fmt.Sprintf("Show %s %s", kind, scope), tests)
		}
	}

	// More Show Tables
	// Plain
	subqueryAliasName = "TABLES"
	tbl, err = informationSchemaDB.Table(subqueryAliasName)
	req.Nil(err)
	source = evaluator.NewDynamicSourceStage(informationSchemaDB,
		tbl.(*catalog.DynamicTable), 2, subqueryAliasName)
	columnName := "Tables_in_" + defaultDbName
	subquery = evaluator.NewSubquerySourceStage(
		evaluator.NewProjectStage(
			source,
			createProjectedColumn(3, source, subqueryAliasName, "TABLE_NAME",
				subqueryAliasName, columnName, false),
			createProjectedColumn(2, source, subqueryAliasName, "TABLE_SCHEMA",
				subqueryAliasName, "TABLE_SCHEMA", false),
		),
		2,
		isDBName,
		subqueryAliasName,
		false,
	)
	for _, from := range []string{"", " from test", " in test"} {
		tests := []planTest{{
			fmt.Sprintf("show tables%s", from),
			func() evaluator.PlanStage {
				return evaluator.NewProjectStage(
					evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLComparisonExpr(
								evaluator.EQ,
								createSQLColumnExprFromSource(subquery,
									subqueryAliasName, "TABLE_SCHEMA", false),
								evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "test")),
							),
						),
						evaluator.NewOrderByTerm(
							createSQLColumnExprFromSource(subquery,
								subqueryAliasName, columnName, false),
							true,
						),
					),
					createProjectedColumn(1, subquery,
						subqueryAliasName, columnName, subqueryAliasName, columnName, false),
				)
			}}, {
			fmt.Sprintf("show tables%s like 'n'", from),
			func() evaluator.PlanStage {
				return evaluator.NewProjectStage(
					evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLComparisonExpr(
									evaluator.EQ,
									createSQLColumnExprFromSource(subquery,
										subqueryAliasName, "TABLE_SCHEMA", false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "test")),
								),
								evaluator.NewSQLLikeExpr(
									createSQLColumnExprFromSource(subquery,
										subqueryAliasName, columnName, false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n")),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "\\")),
									false,
								),
							),
						),
						evaluator.NewOrderByTerm(
							createSQLColumnExprFromSource(subquery,
								subqueryAliasName, columnName, false),
							true,
						),
					),
					createProjectedColumn(1, subquery, subqueryAliasName, columnName,
						subqueryAliasName, columnName, false),
				)
			}}, {
			fmt.Sprintf("show tables%s where `%s` = 'n'", from, columnName),
			func() evaluator.PlanStage {
				return evaluator.NewProjectStage(
					evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLComparisonExpr(
									evaluator.EQ,
									createSQLColumnExprFromSource(subquery,
										subqueryAliasName, "TABLE_SCHEMA", false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "test")),
								),
								evaluator.NewSQLComparisonExpr(
									evaluator.EQ,
									createSQLColumnExprFromSource(subquery,
										subqueryAliasName, columnName, false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n")),
								),
							),
						),
						evaluator.NewOrderByTerm(
							createSQLColumnExprFromSource(subquery,
								subqueryAliasName, columnName, false),
							true,
						),
					),
					createProjectedColumn(1, subquery, subqueryAliasName,
						columnName, subqueryAliasName, columnName, false),
				)
			}},
		}
		runTestsAsSubtest(fmt.Sprintf("Show Plain Tables %s", from), tests)
	}

	// More Show Tables
	// Full
	subquery = evaluator.NewSubquerySourceStage(
		evaluator.NewProjectStage(
			source,
			createProjectedColumn(3, source,
				subqueryAliasName, "TABLE_NAME", subqueryAliasName, columnName, false),
			createProjectedColumn(4, source,
				subqueryAliasName, "TABLE_TYPE", subqueryAliasName, "Table_type", false),
			createProjectedColumn(2, source,
				subqueryAliasName, "TABLE_SCHEMA", subqueryAliasName, "TABLE_SCHEMA", false),
		),
		2,
		isDBName,
		subqueryAliasName,
		false,
	)

	for _, from := range []string{"", " from test", " in test"} {
		tests := []planTest{{
			fmt.Sprintf("show full tables%s", from),
			func() evaluator.PlanStage {
				return evaluator.NewProjectStage(
					evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLComparisonExpr(evaluator.EQ,
								createSQLColumnExprFromSource(subquery,
									subqueryAliasName, "TABLE_SCHEMA", false),
								evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "test")),
							),
						),
						evaluator.NewOrderByTerm(
							createSQLColumnExprFromSource(subquery,
								subqueryAliasName, columnName, false),
							true,
						),
					),
					createProjectedColumn(1, subquery,
						subqueryAliasName, columnName, subqueryAliasName, columnName, false),
					createProjectedColumn(1, subquery,
						subqueryAliasName, "Table_type", subqueryAliasName, "Table_type", false),
				)
			}}, {
			fmt.Sprintf("show full tables%s like 'n'", from),
			func() evaluator.PlanStage {
				return evaluator.NewProjectStage(
					evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLComparisonExpr(evaluator.EQ,
									createSQLColumnExprFromSource(subquery,
										subqueryAliasName, "TABLE_SCHEMA", false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "test")),
								),
								evaluator.NewSQLLikeExpr(
									createSQLColumnExprFromSource(subquery,
										subqueryAliasName, columnName, false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n")),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "\\")),
									false,
								),
							),
						),
						evaluator.NewOrderByTerm(
							createSQLColumnExprFromSource(subquery,
								subqueryAliasName, columnName, false),
							true,
						),
					),
					createProjectedColumn(1, subquery,
						subqueryAliasName, columnName, subqueryAliasName, columnName, false),
					createProjectedColumn(1, subquery,
						subqueryAliasName, "Table_type", subqueryAliasName, "Table_type", false),
				)
			}}, {
			fmt.Sprintf("show full tables%s where `%s` = 'n'", from, columnName),
			func() evaluator.PlanStage {
				return evaluator.NewProjectStage(
					evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLComparisonExpr(evaluator.EQ,
									createSQLColumnExprFromSource(subquery,
										subqueryAliasName, "TABLE_SCHEMA", false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "test")),
								),
								evaluator.NewSQLComparisonExpr(evaluator.EQ,
									createSQLColumnExprFromSource(subquery,
										subqueryAliasName, columnName, false),
									evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "n")),
								),
							),
						),
						evaluator.NewOrderByTerm(
							createSQLColumnExprFromSource(subquery,
								subqueryAliasName, columnName, false),
							true,
						),
					),
					createProjectedColumn(1, subquery,
						subqueryAliasName, columnName, subqueryAliasName, columnName, false),
					createProjectedColumn(1, subquery,
						subqueryAliasName, "Table_type", subqueryAliasName, "Table_type", false),
				)
			}},
		}
		runTestsAsSubtest(fmt.Sprintf("Show Full Tables %s", from), tests)
	}

	// Select Statements
	// Dual Queries
	selectDualTests := []planTest{{

		"select 2 + 3",
		func() evaluator.PlanStage {
			return evaluator.NewProjectStage(
				evaluator.NewDualStage(),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "2+3",
					evaluator.NewSQLAddExpr(evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 2)),
						evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 3)))),
			)
		}}, {

		"select false",
		func() evaluator.PlanStage {
			return evaluator.NewProjectStage(
				evaluator.NewDualStage(),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "false",
					evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, false))),
			)
		}}, {

		"select true",
		func() evaluator.PlanStage {
			return evaluator.NewProjectStage(
				evaluator.NewDualStage(),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "true",
					evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true))),
			)
		}}, {

		"select 2 + 3 from dual",
		func() evaluator.PlanStage {
			return evaluator.NewProjectStage(
				evaluator.NewDualStage(),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "2+3",
					evaluator.NewSQLAddExpr(evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 2)),
						evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 3)))),
			)
		}}, {

		"select sum(2)",
		func() evaluator.PlanStage {
			source := evaluator.NewDualStage()
			return evaluator.NewProjectStage(
				evaluator.NewGroupByStage(source,
					nil,
					evaluator.ProjectedColumns{
						evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(2)",
							evaluator.NewSQLAggregationFunctionExpr(
								"sum",
								false,
								[]evaluator.SQLExpr{evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 2))},
							),
						),
					},
				),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(2)",
					testSQLColumnExpr(1, "", "", "sum(2)",
						types.EvalDouble, schema.MongoNone, false)),
			)
		}},
	}
	runTestsAsSubtest("Select Dual Queries", selectDualTests)

	// Select union queries

	selectUnionQueries := []planTest{{
		"select * from bar union select * from bar",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			barProject := evaluator.NewProjectStage(barSource,
				createProjectedColumn(1, barSource, "bar", "_id", "bar", "_id", false),
				createProjectedColumn(1, barSource, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, barSource, "bar", "b", "bar", "b", false),
				createProjectedColumn(1, barSource, "bar", "d", "bar", "d", false),
			)
			union := evaluator.NewUnionStage(evaluator.UnionDistinct, barProject, barProject)
			groupBy := evaluator.NewGroupByStage(union,
				[]evaluator.SQLExpr{
					createSQLColumnExprFromSource(union, "bar", "_id", false),
					createSQLColumnExprFromSource(union, "bar", "a", false),
					createSQLColumnExprFromSource(union, "bar", "b", false),
					createSQLColumnExprFromSource(union, "bar", "d", false),
				},
				[]evaluator.ProjectedColumn{
					createProjectedColumn(1, union, "bar", "_id", "bar", "_id", false),
					createProjectedColumn(1, union, "bar", "a", "bar", "a", false),
					createProjectedColumn(1, union, "bar", "b", "bar", "b", false),
					createProjectedColumn(1, union, "bar", "d", "bar", "d", false),
				},
			)
			ret := evaluator.NewProjectStage(groupBy,
				createProjectedColumn(1, groupBy, "bar", "_id", "bar", "_id", false),
				createProjectedColumn(1, groupBy, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, groupBy, "bar", "b", "bar", "b", false),
				createProjectedColumn(1, groupBy, "bar", "d", "bar", "d", false),
			)
			return ret
		}}, {
		"select * from bar union select * from bar union select * from bar",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			barProject := evaluator.NewProjectStage(barSource,
				createProjectedColumn(1, barSource, "bar", "_id", "bar", "_id", false),
				createProjectedColumn(1, barSource, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, barSource, "bar", "b", "bar", "b", false),
				createProjectedColumn(1, barSource, "bar", "d", "bar", "d", false),
			)
			union1 := evaluator.NewUnionStage(evaluator.UnionDistinct, barProject, barProject)
			groupBy1 := evaluator.NewGroupByStage(union1,
				[]evaluator.SQLExpr{
					createSQLColumnExprFromSource(union1, "bar", "_id", false),
					createSQLColumnExprFromSource(union1, "bar", "a", false),
					createSQLColumnExprFromSource(union1, "bar", "b", false),
					createSQLColumnExprFromSource(union1, "bar", "d", false),
				},
				[]evaluator.ProjectedColumn{
					createProjectedColumn(1, union1, "bar", "_id", "bar", "_id", false),
					createProjectedColumn(1, union1, "bar", "a", "bar", "a", false),
					createProjectedColumn(1, union1, "bar", "b", "bar", "b", false),
					createProjectedColumn(1, union1, "bar", "d", "bar", "d", false),
				},
			)
			project1 := evaluator.NewProjectStage(groupBy1,
				createProjectedColumn(1, groupBy1, "bar", "_id", "bar", "_id", false),
				createProjectedColumn(1, groupBy1, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, groupBy1, "bar", "b", "bar", "b", false),
				createProjectedColumn(1, groupBy1, "bar", "d", "bar", "d", false),
			)
			union := evaluator.NewUnionStage(evaluator.UnionDistinct, project1, barProject)
			groupBy := evaluator.NewGroupByStage(union,
				[]evaluator.SQLExpr{
					createSQLColumnExprFromSource(union, "bar", "_id", false),
					createSQLColumnExprFromSource(union, "bar", "a", false),
					createSQLColumnExprFromSource(union, "bar", "b", false),
					createSQLColumnExprFromSource(union, "bar", "d", false),
				},
				[]evaluator.ProjectedColumn{
					createProjectedColumn(1, union, "bar", "_id", "bar", "_id", false),
					createProjectedColumn(1, union, "bar", "a", "bar", "a", false),
					createProjectedColumn(1, union, "bar", "b", "bar", "b", false),
					createProjectedColumn(1, union, "bar", "d", "bar", "d", false),
				},
			)
			ret := evaluator.NewProjectStage(groupBy,
				createProjectedColumn(1, groupBy, "bar", "_id", "bar", "_id", false),
				createProjectedColumn(1, groupBy, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, groupBy, "bar", "b", "bar", "b", false),
				createProjectedColumn(1, groupBy, "bar", "d", "bar", "d", false),
			)
			return ret
		}}, {
		"select * from bar union all select * from bar",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			barProject := evaluator.NewProjectStage(barSource,
				createProjectedColumn(1, barSource, "bar", "_id", "bar", "_id", false),
				createProjectedColumn(1, barSource, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, barSource, "bar", "b", "bar", "b", false),
				createProjectedColumn(1, barSource, "bar", "d", "bar", "d", false),
			)
			union := evaluator.NewUnionStage(evaluator.UnionAll, barProject, barProject)
			ret := evaluator.NewProjectStage(union,
				createProjectedColumn(1, union, "bar", "_id", "bar", "_id", false),
				createProjectedColumn(1, union, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, union, "bar", "b", "bar", "b", false),
				createProjectedColumn(1, union, "bar", "d", "bar", "d", false),
			)
			return ret
		}}, {
		"select * from bar union all select * from bar union all select * from bar",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			barProject := evaluator.NewProjectStage(barSource,
				createProjectedColumn(1, barSource, "bar", "_id", "bar", "_id", false),
				createProjectedColumn(1, barSource, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, barSource, "bar", "b", "bar", "b", false),
				createProjectedColumn(1, barSource, "bar", "d", "bar", "d", false),
			)
			union1 := evaluator.NewUnionStage(evaluator.UnionAll, barProject, barProject)
			project1 := evaluator.NewProjectStage(union1,
				createProjectedColumn(1, union1, "bar", "_id", "bar", "_id", false),
				createProjectedColumn(1, union1, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, union1, "bar", "b", "bar", "b", false),
				createProjectedColumn(1, union1, "bar", "d", "bar", "d", false),
			)
			union := evaluator.NewUnionStage(evaluator.UnionAll, project1, barProject)
			ret := evaluator.NewProjectStage(union,
				createProjectedColumn(1, union, "bar", "_id", "bar", "_id", false),
				createProjectedColumn(1, union, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, union, "bar", "b", "bar", "b", false),
				createProjectedColumn(1, union, "bar", "d", "bar", "d", false),
			)
			return ret
		}},
	}
	runTestsAsSubtest("Select Union Queries", selectUnionQueries)

	cteTest := []planTest{{
		"with cte1 as (select a from foo) select a from cte1",
		func() evaluator.PlanStage {
			source := createMongoSource(2, "foo", "foo")
			subquery := evaluator.NewSubquerySourceStage(
				evaluator.NewProjectStage(
					source,
					createProjectedColumn(2, source, "foo", "a", "foo", "a", false),
				),
				2,
				"test",
				"cte1",
				true,
			)
			return evaluator.NewProjectStage(subquery,
				createProjectedColumn(2, subquery, "cte1", "a", "cte1", "a", false))
		}}, {
		"with cte1 as (select a from foo) select a from cte1 UNION ALL select a from cte1",
		func() evaluator.PlanStage {
			source := createMongoSource(2, "foo", "foo")
			subquery := evaluator.NewSubquerySourceStage(
				evaluator.NewProjectStage(
					source,
					createProjectedColumn(2, source, "foo", "a", "foo", "a", false),
				),
				2,
				"test",
				"cte1",
				true,
			)
			branch := evaluator.NewProjectStage(subquery,
				createProjectedColumn(2, subquery, "cte1", "a", "cte1", "a", false))
			union := evaluator.NewUnionStage(evaluator.UnionAll, branch, branch)
			ret := evaluator.NewProjectStage(union, createProjectedColumn(2, union,
				"cte1", "a", "cte1", "a", false))
			return ret
		}},
	}

	runTestsAsSubtest("CTE fun", cteTest)

	// Select From Subqueries
	selectFromSubqueriesPlanTests := []planTest{{
		"select a from (select a from foo) f",
		func() evaluator.PlanStage {
			source := createMongoSource(2, "foo", "foo")
			subquery := evaluator.NewSubquerySourceStage(
				evaluator.NewProjectStage(
					source,
					createProjectedColumn(2, source, "foo", "a", "foo", "a", false),
				),
				2,
				"test",
				"f",
				false,
			)
			return evaluator.NewProjectStage(subquery,
				createProjectedColumn(2, subquery, "f", "a", "f", "a", false))
		}}, {

		"select f.a from (select a from foo) f",
		func() evaluator.PlanStage {
			source := createMongoSource(2, "foo", "foo")
			subquery := evaluator.NewSubquerySourceStage(evaluator.NewProjectStage(source,
				createProjectedColumn(2, source, "foo", "a", "foo", "a", false)), 2, "test", "f", false)
			return evaluator.NewProjectStage(subquery,
				createProjectedColumn(2, subquery, "f", "a", "f", "a", false))
		}}, {

		"select f.a from (select test.a from foo test) f",
		func() evaluator.PlanStage {
			source := createMongoSource(2, "foo", "test")
			subquery := evaluator.NewSubquerySourceStage(evaluator.NewProjectStage(source,
				createProjectedColumn(2, source, "test", "a", "test", "a", false)), 2, "test", "f", false)
			return evaluator.NewProjectStage(subquery,
				createProjectedColumn(2, subquery, "f", "a", "f", "a", false))
		}},
	}
	runTestsAsSubtest("Select From Subqueries (Plan)", selectFromSubqueriesPlanTests)

	selectFromSubqueriesVariablesTests := []variableTest{
		{
			"select g.a from (select a from foo) g",
			func() *variable.Container {
				vars := &variable.Container{}
				setSystemVariable(vars, variable.MongoDBGitVersion, testInfo.GitVersion)
				setSystemVariable(vars, variable.MongoDBVersion, testInfo.Version)
				setSystemVariable(vars, variable.MongoDBVersionCompatibility, testInfo.CompatibleVersion)
				setSystemVariable(vars, variable.SQLSelectLimit, 5)
				setSystemVariable(vars, variable.TypeConversionMode, "mysql")
				setSystemVariable(vars, variable.PolymorphicTypeConversionMode, "off")
				return vars
			},
			func() *mongodb.Info {
				return testInfo
			},
			func() evaluator.PlanStage {
				source := createMongoSource(2, "foo", "foo")
				subquery := evaluator.NewSubquerySourceStage(
					evaluator.NewProjectStage(source,
						createProjectedColumn(2, source, "foo", "a", "foo", "a", false)),
					2, "test", "g", false)
				return evaluator.NewProjectStage(evaluator.NewLimitStage(subquery, 0, 5),
					createProjectedColumn(2, subquery, "g", "a", "g", "a", false))
			}},
		{
			"select fast.a from (select a from foo) fast",
			func() *variable.Container {
				vars := &variable.Container{}
				setSystemVariable(vars, variable.MongoDBGitVersion, testInfo.GitVersion)
				setSystemVariable(vars, variable.MongoDBVersion, testInfo.Version)
				setSystemVariable(vars, variable.MongoDBVersionCompatibility, testInfo.CompatibleVersion)
				setSystemVariable(vars, variable.SQLSelectLimit, 5)
				setSystemVariable(vars, variable.TypeConversionMode, "mysql")
				setSystemVariable(vars, variable.PolymorphicTypeConversionMode, "fast")
				return vars
			},
			func() *mongodb.Info {
				return testInfo
			},
			func() evaluator.PlanStage {
				source := createMongoSource(2, "foo", "foo")
				subquery := evaluator.NewSubquerySourceStage(
					evaluator.NewProjectStage(source,
						createProjectedColumn(2, source, "foo", "a", "foo", "a", false)),
					2, "test", "fast", false)
				return evaluator.NewProjectStage(evaluator.NewLimitStage(subquery, 0, 5),
					createProjectedColumn(2, subquery, "fast", "a", "fast", "a", false))
			}},
		{
			"select safe.a from (select a from foo) safe",
			func() *variable.Container {
				vars := &variable.Container{}
				setSystemVariable(vars, variable.MongoDBGitVersion, testInfo.GitVersion)
				setSystemVariable(vars, variable.MongoDBVersion, testInfo.Version)
				setSystemVariable(vars, variable.MongoDBVersionCompatibility, testInfo.CompatibleVersion)
				setSystemVariable(vars, variable.SQLSelectLimit, 5)
				setSystemVariable(vars, variable.TypeConversionMode, "mysql")
				setSystemVariable(vars, variable.PolymorphicTypeConversionMode, "safe")
				return vars
			},
			func() *mongodb.Info {
				return testInfo
			},
			func() evaluator.PlanStage {
				source := createMongoSource(2, "foo", "foo")
				projectedCol := createProjectedColumn(2, source, "foo", "a", "foo", "a", false)
				projectedCol.Expr = evaluator.NewSQLConvertExpr(projectedCol.Expr,
					types.EvalInt64)
				projectedCol.MongoType = ""
				projectedCol.OriginalName = ""
				projectedCol.OriginalTable = ""
				projectedCol.Table = ""
				subquery := evaluator.NewSubquerySourceStage(
					evaluator.NewProjectStage(source, projectedCol),
					2, "test", "safe", false)
				outterProjectedCol := createProjectedColumn(1, subquery, "safe", "a", "", "a", false)
				outterProjectedCol.Expr = evaluator.NewSQLConvertExpr(outterProjectedCol.Expr,
					types.EvalInt64)
				return evaluator.NewProjectStage(evaluator.NewLimitStage(subquery, 0, 5),
					outterProjectedCol)
			}},
	}
	runTestsAsSubtest("Select From Subqueries (Variables)", selectFromSubqueriesVariablesTests)

	// Select From Joins
	selectFromJoinsPlanTests := []planTest{{
		"select foo.a, bar.a from foo, bar",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(1, "bar", "bar")
			join := evaluator.NewJoinStage(evaluator.CrossJoin, fooSource,
				barSource, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "foo", "a", "foo", "a", false),
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
			)
		}}, {

		"select f.a, bar.a from foo f, bar",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "f")
			barSource := createMongoSource(1, "bar", "bar")
			join := evaluator.NewJoinStage(evaluator.CrossJoin, fooSource,
				barSource, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "f", "a", "f", "a", false),
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
			)
		}}, {

		"select f.a, b.a from foo f, bar b",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "f")
			barSource := createMongoSource(1, "bar", "b")
			join := evaluator.NewJoinStage(evaluator.CrossJoin, fooSource,
				barSource, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "f", "a", "f", "a", false),
				createProjectedColumn(1, join, "b", "a", "b", "a", false),
			)
		}}, {

		"select foo.a, bar.a from foo inner join bar on foo.b = bar.b",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(1, "bar", "bar")
			join := evaluator.NewJoinStage(evaluator.InnerJoin, fooSource,
				barSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(fooSource, "foo", "b", false),
					createSQLColumnExprFromSource(barSource, "bar", "b", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "foo", "a", "foo", "a", false),
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
			)
		}}, {

		"select foo.a, bar.a from foo join bar on foo.b = bar.b",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(1, "bar", "bar")
			join := evaluator.NewJoinStage(evaluator.InnerJoin, fooSource,
				barSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(fooSource, "foo", "b", false),
					createSQLColumnExprFromSource(barSource, "bar", "b", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "foo", "a", "foo", "a", false),
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
			)
		}}, {

		"select foo.a, bar.a from foo left outer join bar on foo.b = bar.b",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(1, "bar", "bar")
			join := evaluator.NewJoinStage(evaluator.LeftJoin, fooSource,
				barSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(fooSource, "foo", "b", false),
					createSQLColumnExprFromSource(barSource, "bar", "b", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "foo", "a", "foo", "a", false),
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
			)
		}}, {

		"select foo.a, bar.a from foo right outer join bar on foo.b = bar.b",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(1, "bar", "bar")
			join := evaluator.NewJoinStage(parser.AST_RIGHT_JOIN, fooSource,
				barSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(fooSource, "foo", "b", false),
					createSQLColumnExprFromSource(barSource, "bar", "b", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "foo", "a", "foo", "a", false),
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
			)
		}}, {

		"select foo.a, bar.a from foo straight_join bar on foo.b = bar.b",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(1, "bar", "bar")
			join := evaluator.NewJoinStage(evaluator.StraightJoin, fooSource,
				barSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(fooSource, "foo", "b", false),
					createSQLColumnExprFromSource(barSource, "bar", "b", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "foo", "a", "foo", "a", false),
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
			)
		}}, {

		"select foo.a, bar.a from foo join bar on foo.a = bar.a" +
			" and foo.e = bar.d join baz on baz.b = bar.b",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			firstJoin := evaluator.NewJoinStage(evaluator.InnerJoin, fooSource,
				barSource,
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(fooSource, "foo", "a", false),
						createSQLColumnExprFromSource(barSource, "bar", "a", false),
					),
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(fooSource, "foo", "e", false),
						createSQLColumnExprFromSource(barSource, "bar", "d", false),
					),
				),
			)
			secondJoin := evaluator.NewJoinStage(evaluator.InnerJoin, firstJoin, bazSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bazSource, "baz", "b", false),
					createSQLColumnExprFromSource(barSource, "bar", "b", false),
				),
			)
			return evaluator.NewProjectStage(secondJoin,
				createProjectedColumn(1, secondJoin, "foo", "a", "foo", "a", false),
				createProjectedColumn(1, secondJoin, "bar", "a", "bar", "a", false),
			)
		}}, {

		"select bar.a, baz.b from bar join baz using (b)",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			join := evaluator.NewJoinStage(evaluator.InnerJoin,
				barSource, bazSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(barSource, "bar", "b", false),
					createSQLColumnExprFromSource(bazSource, "baz", "b", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, join, "baz", "b", "baz", "b", false),
			)
		}}, {

		"select bar.a, baz.b from bar join baz using (a, b)",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			join := evaluator.NewJoinStage(evaluator.InnerJoin,
				barSource, bazSource,
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "a", false),
						createSQLColumnExprFromSource(bazSource, "baz", "a", false),
					),
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "b", false),
						createSQLColumnExprFromSource(bazSource, "baz", "b", false),
					),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, join, "baz", "b", "baz", "b", false),
			)
		}}, {

		"select bar.a, buzz.d, foo.c from bar join buzz join foo using (a, c)",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			buzzSource := createMongoSource(1, "buzz", "buzz")
			fooSource := createMongoSource(1, "foo", "foo")
			firstJoin := evaluator.NewJoinStage(parser.AST_CROSS_JOIN,
				barSource, buzzSource, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			secondJoin := evaluator.NewJoinStage(evaluator.InnerJoin, firstJoin, fooSource,
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "a", false),
						createSQLColumnExprFromSource(fooSource, "foo", "a", false),
					),
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(buzzSource, "buzz", "c", false),
						createSQLColumnExprFromSource(fooSource, "foo", "c", false),
					),
				),
			)
			return evaluator.NewProjectStage(secondJoin,
				createProjectedColumn(1, secondJoin, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, secondJoin, "buzz", "d", "buzz", "d", false),
				createProjectedColumn(1, secondJoin, "foo", "c", "foo", "c", false),
			)
		}}, {

		"select bar.a, buzz.d, foo.c from bar join foo using (a) join buzz using (c)",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			fooSource := createMongoSource(1, "foo", "foo")
			buzzSource := createMongoSource(1, "buzz", "buzz")
			firstJoin := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, fooSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(barSource, "bar", "a", false),
					createSQLColumnExprFromSource(fooSource, "foo", "a", false),
				),
			)
			secondJoin := evaluator.NewJoinStage(evaluator.InnerJoin, firstJoin, buzzSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(fooSource, "foo", "c", false),
					createSQLColumnExprFromSource(buzzSource, "buzz", "c", false),
				),
			)
			return evaluator.NewProjectStage(secondJoin,
				createProjectedColumn(1, secondJoin, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, secondJoin, "buzz", "d", "buzz", "d", false),
				createProjectedColumn(1, secondJoin, "foo", "c", "foo", "c", false),
			)
		}}, {

		"select bar.a, baz.b from bar join baz using (a, a, a, a, b)",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "a", false),
						createSQLColumnExprFromSource(bazSource, "baz", "a", false),
					),
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "b", false),
						createSQLColumnExprFromSource(bazSource, "baz", "b", false),
					),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, join, "baz", "b", "baz", "b", false),
			)
		}}, {

		"select bar.a, baz.b from bar join baz using (a, b, b, b, b)",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "a", false),
						createSQLColumnExprFromSource(bazSource, "baz", "a", false),
					),
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "b", false),
						createSQLColumnExprFromSource(bazSource, "baz", "b", false),
					),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, join, "baz", "b", "baz", "b", false),
			)
		}}, {

		"select bar.a, baz.b from bar cross join baz using (b)",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			join := evaluator.NewJoinStage(evaluator.InnerJoin,
				barSource, bazSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(barSource, "bar", "b", false),
					createSQLColumnExprFromSource(bazSource, "baz", "b", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, join, "baz", "b", "baz", "b", false),
			)
		}}, {

		"select bar.a, baz.b from bar inner join baz",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			join := evaluator.NewJoinStage(evaluator.CrossJoin,
				barSource, bazSource, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, join, "baz", "b", "baz", "b", false),
			)
		}}, {

		"select bar.a, biz.b from bar join (select baz.b, foo.c from" +
			" baz join foo on baz.a = foo.a) as biz using (b)",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(2, "baz", "baz")
			fooSource := createMongoSource(2, "foo", "foo")
			subJoin := evaluator.NewJoinStage(evaluator.InnerJoin, bazSource, fooSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bazSource, "baz", "a", false),
					createSQLColumnExprFromSource(fooSource, "foo", "a", false),
				),
			)
			bizSource := evaluator.NewSubquerySourceStage(
				evaluator.NewProjectStage(subJoin,
					createProjectedColumn(2, subJoin, "baz", "b", "baz", "b", false),
					createProjectedColumn(2, subJoin, "foo", "c", "foo", "c", false),
				), 2, "test", "biz", false)
			join := evaluator.NewJoinStage(evaluator.InnerJoin,
				barSource, bizSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(barSource, "bar", "b", false),
					createSQLColumnExprFromSource(bizSource, "biz", "b", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
				createProjectedColumn(2, join, "biz", "b", "biz", "b", false),
			)
		}}, {

		"select bar.a, biz.b from (select baz.b, foo.c from " +
			"baz join foo on baz.a = foo.a) as biz join bar using (b)",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(2, "baz", "baz")
			fooSource := createMongoSource(2, "foo", "foo")
			subJoin := evaluator.NewJoinStage(evaluator.InnerJoin, bazSource, fooSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bazSource, "baz", "a", false),
					createSQLColumnExprFromSource(fooSource, "foo", "a", false),
				),
			)
			bizSource := evaluator.NewSubquerySourceStage(
				evaluator.NewProjectStage(subJoin,
					createProjectedColumn(2, subJoin, "baz", "b", "baz", "b", false),
					createProjectedColumn(2, subJoin, "foo", "c", "foo", "c", false),
				), 2, "test", "biz", false)
			join := evaluator.NewJoinStage(evaluator.InnerJoin, bizSource, barSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bizSource, "biz", "b", false),
					createSQLColumnExprFromSource(barSource, "bar", "b", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
				createProjectedColumn(2, join, "biz", "b", "biz", "b", false),
			)
		}}, {

		"select fiz.b from (select bar.b from bar) as biz " +
			"join (select foo.b from foo) as fiz using (b)",
		func() evaluator.PlanStage {
			barSource := createMongoSource(2, "bar", "bar")
			fooSource := createMongoSource(3, "foo", "foo")
			bizSource := evaluator.NewSubquerySourceStage(
				evaluator.NewProjectStage(
					barSource,
					createProjectedColumn(2, barSource, "bar", "b", "bar", "b", false),
				),
				2,
				"test",
				"biz",
				false,
			)
			fizSource := evaluator.NewSubquerySourceStage(
				evaluator.NewProjectStage(
					fooSource,
					createProjectedColumn(3, fooSource, "foo", "b", "foo", "b", false),
				),
				3,
				"test",
				"fiz",
				false,
			)
			join := evaluator.NewJoinStage(evaluator.InnerJoin, bizSource, fizSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bizSource, "biz", "b", false),
					createSQLColumnExprFromSource(fizSource, "fiz", "b", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(3, join, "fiz", "b", "fiz", "b", false))
		}}, {

		"select * from bar join baz using (b)",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(barSource, "bar", "b", false),
					createSQLColumnExprFromSource(bazSource, "baz", "b", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "bar", "b", "bar", "b", false),
				createProjectedColumn(1, join, "bar", "_id", "bar", "_id", false),
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, join, "bar", "d", "bar", "d", false),
				createProjectedColumn(1, join, "baz", "_id", "baz", "_id", false),
				createProjectedColumn(1, join, "baz", "a", "baz", "a", false),
			)
		}}, {

		"select * from bar join baz using (_id, b)",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "_id", false),
						createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
					),
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "b", false),
						createSQLColumnExprFromSource(bazSource, "baz", "b", false),
					),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "bar", "_id", "bar", "_id", false),
				createProjectedColumn(1, join, "bar", "b", "bar", "b", false),
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, join, "bar", "d", "bar", "d", false),
				createProjectedColumn(1, join, "baz", "a", "baz", "a", false))
		}}, {

		"select * from bar right join baz using (b)",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			join := evaluator.NewJoinStage(evaluator.RightJoin, barSource, bazSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(barSource, "bar", "b", false),
					createSQLColumnExprFromSource(bazSource, "baz", "b", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "baz", "b", "baz", "b", false),
				createProjectedColumn(1, join, "baz", "_id", "baz", "_id", false),
				createProjectedColumn(1, join, "baz", "a", "baz", "a", false),
				createProjectedColumn(1, join, "bar", "_id", "bar", "_id", false),
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, join, "bar", "d", "bar", "d", false),
			)
		}}, {

		"select bar.*, baz.* from bar join baz using (b)",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(barSource, "bar", "b", false),
					createSQLColumnExprFromSource(bazSource, "baz", "b", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "bar", "_id", "bar", "_id", false),
				createProjectedColumn(1, join, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, join, "bar", "b", "bar", "b", false),
				createProjectedColumn(1, join, "bar", "d", "bar", "d", false),
				createProjectedColumn(1, join, "baz", "_id", "baz", "_id", false),
				createProjectedColumn(1, join, "baz", "a", "baz", "a", false),
				createProjectedColumn(1, join, "baz", "b", "baz", "b", false),
			)
		}}, {

		"select bar.b, baz.b from bar join baz using (b)",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(barSource, "bar", "b", false),
					createSQLColumnExprFromSource(bazSource, "baz", "b", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, join, "bar", "b", "bar", "b", false),
				createProjectedColumn(1, join, "baz", "b", "baz", "b", false))
		}}, {

		"select * from buzz join (baz join bar using (_id)) using (d)",
		func() evaluator.PlanStage {
			buzzSource := createMongoSource(1, "buzz", "buzz")
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			join1 := evaluator.NewJoinStage(evaluator.InnerJoin, bazSource, barSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
					createSQLColumnExprFromSource(barSource, "bar", "_id", false),
				),
			)
			join2 := evaluator.NewJoinStage(evaluator.InnerJoin, buzzSource, join1,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(buzzSource, "buzz", "d", false),
					createSQLColumnExprFromSource(barSource, "bar", "d", false),
				),
			)
			return evaluator.NewProjectStage(join2,
				createProjectedColumn(1, buzzSource, "buzz", "d", "buzz", "d", false),
				createProjectedColumn(1, buzzSource, "buzz", "_id", "buzz", "_id", false),
				createProjectedColumn(1, buzzSource, "buzz", "c", "buzz", "c", false),
				createProjectedColumn(1, bazSource, "baz", "_id", "baz", "_id", false),
				createProjectedColumn(1, bazSource, "baz", "a", "baz", "a", false),
				createProjectedColumn(1, bazSource, "baz", "b", "baz", "b", false),
				createProjectedColumn(1, barSource, "bar", "a", "bar", "a", false),
				createProjectedColumn(1, barSource, "bar", "b", "bar", "b", false),
			)
		}}, {

		"select bar.a from bar natural join baz",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLAndExpr(
						evaluator.NewSQLComparisonExpr(
							evaluator.EQ,
							createSQLColumnExprFromSource(barSource, "bar", "_id", false),
							createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
						),
						evaluator.NewSQLComparisonExpr(
							evaluator.EQ,
							createSQLColumnExprFromSource(barSource, "bar", "a", false),
							createSQLColumnExprFromSource(bazSource, "baz", "a", false),
						),
					),
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "b", false),
						createSQLColumnExprFromSource(bazSource, "baz", "b", false),
					),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, barSource, "bar", "a", "bar", "a", false))
		}}, {

		"select buzz.c from buzz join bar natural join baz",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(1, "buzz", "buzz")
			naturalJoin := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLAndExpr(
						evaluator.NewSQLComparisonExpr(
							evaluator.EQ,
							createSQLColumnExprFromSource(barSource, "bar", "_id", false),
							createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
						),
						evaluator.NewSQLComparisonExpr(
							evaluator.EQ,
							createSQLColumnExprFromSource(barSource, "bar", "a", false),
							createSQLColumnExprFromSource(bazSource, "baz", "a", false),
						),
					),
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "b", false),
						createSQLColumnExprFromSource(bazSource, "baz", "b", false),
					),
				),
			)
			join := evaluator.NewJoinStage(evaluator.CrossJoin,
				buzzSource, naturalJoin, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, buzzSource, "buzz", "c", "buzz", "c", false))
		}}, {

		"select buzz.c from bar join buzz natural join baz",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(1, "buzz", "buzz")
			naturalJoin := evaluator.NewJoinStage(evaluator.InnerJoin,
				buzzSource, bazSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(buzzSource, "buzz", "_id", false),
					createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
				),
			)
			join := evaluator.NewJoinStage(evaluator.CrossJoin,
				barSource, naturalJoin, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, buzzSource, "buzz", "c", "buzz", "c", false))
		}}, {

		"select bar.a from bar natural join buzz natural join baz",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(1, "buzz", "buzz")

			njoin1 := evaluator.NewJoinStage(evaluator.InnerJoin, buzzSource, bazSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(buzzSource, "buzz", "_id", false),
					createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
				),
			)
			njoin2 := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, njoin1,
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLAndExpr(
						evaluator.NewSQLAndExpr(
							evaluator.NewSQLComparisonExpr(
								evaluator.EQ,
								createSQLColumnExprFromSource(barSource, "bar", "_id", false),
								createSQLColumnExprFromSource(buzzSource, "buzz", "_id", false),
							),
							evaluator.NewSQLComparisonExpr(
								evaluator.EQ,
								createSQLColumnExprFromSource(barSource, "bar", "a", false),
								createSQLColumnExprFromSource(bazSource, "baz", "a", false),
							),
						),
						evaluator.NewSQLComparisonExpr(
							evaluator.EQ,
							createSQLColumnExprFromSource(barSource, "bar", "b", false),
							createSQLColumnExprFromSource(bazSource, "baz", "b", false),
						),
					),
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "d", false),
						createSQLColumnExprFromSource(buzzSource, "buzz", "d", false),
					),
				),
			)
			return evaluator.NewProjectStage(njoin2,
				createProjectedColumn(1, barSource, "bar", "a", "bar", "a", false))
		}}, {

		"select baz.a from (select c from buzz) as buzzc natural join baz",
		func() evaluator.PlanStage {
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(2, "buzz", "buzz")
			buzzcSource := evaluator.NewSubquerySourceStage(
				evaluator.NewProjectStage(buzzSource,
					createProjectedColumn(2, buzzSource, "buzz", "c", "buzz", "c", false)),
				2, "test", "buzzc", false)
			join := evaluator.NewJoinStage(evaluator.CrossJoin,
				buzzcSource, bazSource, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			return evaluator.NewProjectStage(join, createProjectedColumn(1, bazSource, "baz", "a", "baz", "a", false))
		}}, {

		"select buzz.c from bar join buzz using (_id, d) natural join baz",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(1, "buzz", "buzz")
			usingJoin := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, buzzSource,
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "_id", false),
						createSQLColumnExprFromSource(buzzSource, "buzz", "_id", false),
					),
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "d", false),
						createSQLColumnExprFromSource(buzzSource, "buzz", "d", false),
					),
				),
			)
			naturalJoin := evaluator.NewJoinStage(evaluator.InnerJoin, usingJoin, bazSource,
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLAndExpr(
						evaluator.NewSQLComparisonExpr(
							evaluator.EQ,
							createSQLColumnExprFromSource(barSource, "bar", "_id", false),
							createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
						),
						evaluator.NewSQLComparisonExpr(
							evaluator.EQ,
							createSQLColumnExprFromSource(barSource, "bar", "a", false),
							createSQLColumnExprFromSource(bazSource, "baz", "a", false),
						),
					),
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "b", false),
						createSQLColumnExprFromSource(bazSource, "baz", "b", false),
					),
				),
			)
			return evaluator.NewProjectStage(naturalJoin,
				createProjectedColumn(1, buzzSource, "buzz", "c", "buzz", "c", false))
		}}, {

		"select baz.b from baz natural left join buzz",
		func() evaluator.PlanStage {
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(1, "buzz", "buzz")
			join := evaluator.NewJoinStage(evaluator.LeftJoin, bazSource, buzzSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
					createSQLColumnExprFromSource(buzzSource, "buzz", "_id", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, bazSource, "baz", "b", "baz", "b", false))
		}}, {

		"select baz.b from baz natural right join buzz",
		func() evaluator.PlanStage {
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(1, "buzz", "buzz")
			join := evaluator.NewJoinStage(evaluator.RightJoin, bazSource, buzzSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
					createSQLColumnExprFromSource(buzzSource, "buzz", "_id", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, bazSource, "baz", "b", "baz", "b", false))
		}}, {

		"select baz.b from baz natural left outer join buzz",
		func() evaluator.PlanStage {
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(1, "buzz", "buzz")
			join := evaluator.NewJoinStage(evaluator.LeftJoin, bazSource, buzzSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
					createSQLColumnExprFromSource(buzzSource, "buzz", "_id", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, bazSource, "baz", "b", "baz", "b", false))
		}}, {

		"select baz.b from baz natural right outer join buzz",
		func() evaluator.PlanStage {
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(1, "buzz", "buzz")
			join := evaluator.NewJoinStage(evaluator.RightJoin, bazSource, buzzSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
					createSQLColumnExprFromSource(buzzSource, "buzz", "_id", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, bazSource, "baz", "b", "baz", "b", false))
		}}, {

		"select baz.b from foo join baz natural right join buzz",
		func() evaluator.PlanStage {
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(1, "buzz", "buzz")
			fooSource := createMongoSource(1, "foo", "foo")
			njoin := evaluator.NewJoinStage(evaluator.RightJoin, bazSource, buzzSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
					createSQLColumnExprFromSource(buzzSource, "buzz", "_id", false),
				),
			)
			join := evaluator.NewJoinStage(
				evaluator.CrossJoin, fooSource, njoin,
				evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, bazSource, "baz", "b", "baz", "b", false))
		}}, {

		"select baz.b from baz natural right join buzz join foo",
		func() evaluator.PlanStage {
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(1, "buzz", "buzz")
			fooSource := createMongoSource(1, "foo", "foo")
			njoin := evaluator.NewJoinStage(evaluator.RightJoin, bazSource, buzzSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
					createSQLColumnExprFromSource(buzzSource, "buzz", "_id", false),
				),
			)
			join := evaluator.NewJoinStage(
				evaluator.CrossJoin, njoin, fooSource,
				evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, bazSource, "baz", "b", "baz", "b", false))
		}}, {

		"select baz.b from foo join baz natural left join buzz",
		func() evaluator.PlanStage {
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(1, "buzz", "buzz")
			fooSource := createMongoSource(1, "foo", "foo")
			njoin := evaluator.NewJoinStage(evaluator.LeftJoin, bazSource, buzzSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
					createSQLColumnExprFromSource(buzzSource, "buzz", "_id", false),
				),
			)
			join := evaluator.NewJoinStage(
				evaluator.CrossJoin, fooSource, njoin,
				evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, bazSource, "baz", "b", "baz", "b", false))
		}}, {

		"select baz.b from baz natural left join buzz join foo",
		func() evaluator.PlanStage {
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(1, "buzz", "buzz")
			fooSource := createMongoSource(1, "foo", "foo")
			njoin := evaluator.NewJoinStage(evaluator.LeftJoin, bazSource, buzzSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
					createSQLColumnExprFromSource(buzzSource, "buzz", "_id", false),
				),
			)
			join := evaluator.NewJoinStage(evaluator.CrossJoin, njoin,
				fooSource, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, bazSource, "baz", "b", "baz", "b", false))
		}}, {

		"select baz.b from (select c from buzz) as buzzc natural left join baz",
		func() evaluator.PlanStage {
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(2, "buzz", "buzz")
			buzzcSource := evaluator.NewSubquerySourceStage(
				evaluator.NewProjectStage(buzzSource,
					createProjectedColumn(2, buzzSource,
						"buzz", "c", "buzz", "c", false)), 2, "test", "buzzc", false)
			join := evaluator.NewJoinStage(evaluator.CrossJoin, buzzcSource,
				bazSource, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, bazSource, "baz", "b", "baz", "b", false))
		}}, {

		"select baz.b from (select c from buzz) as buzzc natural right join baz",
		func() evaluator.PlanStage {
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(2, "buzz", "buzz")
			buzzcSource := evaluator.NewSubquerySourceStage(
				evaluator.NewProjectStage(buzzSource,
					createProjectedColumn(2, buzzSource,
						"buzz", "c", "buzz", "c", false)), 2, "test", "buzzc", false)
			join := evaluator.NewJoinStage(evaluator.CrossJoin, buzzcSource,
				bazSource, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, bazSource, "baz", "b", "baz", "b", false))
		}},
	}

	runTestsAsSubtest("Select From Joins", selectFromJoinsPlanTests)

	simpleSelectStarTests := []planTest{{
		"select * from foo", func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(source,
				createAllProjectedColumnsFromSource(1, source, "foo")...)
		}}, {

		"select foo.* from foo", func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(source,
				createAllProjectedColumnsFromSource(1, source, "foo")...)
		}}, {

		"select f.* from foo f", func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "f")
			return evaluator.NewProjectStage(source,
				createAllProjectedColumnsFromSource(1, source, "f")...)
		}}, {

		"select a, foo.* from foo", func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			columns := append(
				evaluator.ProjectedColumns{
					createProjectedColumn(1, source, "foo", "a", "foo", "a", false)},
				createAllProjectedColumnsFromSource(1, source, "foo")...)
			return evaluator.NewProjectStage(source, columns...)
		}}, {

		"select foo.*, a from foo", func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			columns := append(
				createAllProjectedColumnsFromSource(1, source, "foo"),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false))
			return evaluator.NewProjectStage(source, columns...)
		}}, {

		"select a, f.* from foo f", func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "f")
			columns := append(
				evaluator.ProjectedColumns{
					createProjectedColumn(1, source, "f", "a", "f", "a", false)},
				createAllProjectedColumnsFromSource(1, source, "f")...)
			return evaluator.NewProjectStage(source, columns...)
		}}, {

		"select * from foo, bar", func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(1, "bar", "bar")
			join := evaluator.NewJoinStage(evaluator.CrossJoin, fooSource,
				barSource, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			fooCols := createAllProjectedColumnsFromSource(1, fooSource, "foo")
			barCols := createAllProjectedColumnsFromSource(1,
				barSource, "bar")
			return evaluator.NewProjectStage(join, append(fooCols, barCols...)...)
		}}, {

		"select foo.*, bar.* from foo, bar", func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(1, "bar", "bar")
			join := evaluator.NewJoinStage(evaluator.CrossJoin, fooSource,
				barSource, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			fooCols := createAllProjectedColumnsFromSource(1, fooSource, "foo")
			barCols := createAllProjectedColumnsFromSource(1,
				barSource, "bar")
			return evaluator.NewProjectStage(join, append(fooCols, barCols...)...)
		}},
	}

	runTestsAsSubtest("Select Simple Star", simpleSelectStarTests)
	simpleSelectNonStarTests := []planTest{{
		"select a from foo",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(source,
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false))
		}}, {

		"select a from foo f",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "f")
			return evaluator.NewProjectStage(source,
				createProjectedColumn(1, source, "f", "a", "f", "a", false))
		}}, {

		"select f.a from foo f",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "f")
			return evaluator.NewProjectStage(source,
				createProjectedColumn(1, source, "f", "a", "f", "a", false))
		}}, {

		"select a as b from foo",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(source,
				createProjectedColumn(1, source, "foo", "a", "foo", "b", false))
		}}, {

		"select a + 2 from foo",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(source,
				evaluator.CreateProjectedColumnFromSQLExpr(1, "a+2",
					evaluator.NewSQLAddExpr(
						testSQLColumnExpr(1,
							defaultDbName, "foo", "a",
							types.EvalInt64, schema.MongoInt, false),
						evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 2)),
					),
				),
			)
		}}, {

		"select a + 2 as b from foo", func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(source,
				evaluator.CreateProjectedColumnFromSQLExpr(1, "b",
					evaluator.NewSQLAddExpr(
						testSQLColumnExpr(1,
							defaultDbName, "foo", "a", types.EvalInt64, schema.MongoInt, false),
						evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 2)),
					),
				),
			)
		}}, {

		"select BENCHMARK(1, a) from foo", func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(source,
				evaluator.CreateProjectedColumnFromSQLExpr(1, "benchmark(1, a)",
					evaluator.NewSQLBenchmarkExpr(
						evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
						testSQLColumnExpr(1,
							defaultDbName, "foo", "a", types.EvalInt64, schema.MongoInt, false),
					),
				),
			)
		}},
	}
	runTestsAsSubtest("Select Simple Non-Star", simpleSelectNonStarTests)

	// Select Non-correlated Subqueries
	nonCorrelatedSubqueriesTests := []planTest{{
		"select a, (select a from bar) from foo",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(2, "bar", "bar")
			return evaluator.NewProjectStage(fooSource,
				createProjectedColumn(1, fooSource, "foo", "a", "foo", "a", false),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "(select a from bar)",
					evaluator.NewSQLSubqueryExpr(
						false,
						false,
						evaluator.NewProjectStage(barSource,
							createProjectedColumn(2, barSource, "bar", "a", "bar", "a", false)),
					),
				),
			)
		}}, {

		"select a, (select a from bar) as b from foo",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(2, "bar", "bar")
			return evaluator.NewProjectStage(fooSource,
				createProjectedColumn(1, fooSource, "foo", "a", "foo", "a", false),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "b",
					evaluator.NewSQLSubqueryExpr(
						false,
						false,
						evaluator.NewProjectStage(barSource,
							createProjectedColumn(2, barSource, "bar", "a", "bar", "a", false)),
					),
				),
			)
		}}, {

		"select a, (select foo.a from foo, bar) from foo",
		func() evaluator.PlanStage {
			foo1Source := createMongoSource(1, "foo", "foo")
			foo2Source := createMongoSource(2, "foo", "foo")
			barSource := createMongoSource(2, "bar", "bar")
			join := evaluator.NewJoinStage(evaluator.CrossJoin,
				foo2Source, barSource, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			return evaluator.NewProjectStage(foo1Source,
				createProjectedColumn(1, foo1Source, "foo", "a", "foo", "a", false),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "(select foo.a from foo, bar)",
					evaluator.NewSQLSubqueryExpr(
						false,
						false,
						evaluator.NewProjectStage(join,
							createProjectedColumn(2, join, "foo", "a", "foo", "a", false)),
					),
				),
			)
		}}, {

		"select exists(select 1 from bar) from foo",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(2, "bar", "bar")
			return evaluator.NewProjectStage(fooSource,
				evaluator.CreateProjectedColumnFromSQLExpr(1, "exists (select 1 from bar)",
					evaluator.NewSQLExistsExpr(
						false,
						evaluator.NewProjectStage(barSource,
							evaluator.CreateProjectedColumnFromSQLExpr(2, "1",
								evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)))),
					),
				),
			)
		}},
	}
	runTestsAsSubtest("Select Non-correlated Subqueries", nonCorrelatedSubqueriesTests)

	// Select Correlated Subqueries
	correlatedSubqueriesTests := []planTest{{
		"select a, (select foo.a from bar) from foo",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(2, "bar", "bar")
			return evaluator.NewProjectStage(
				fooSource,
				createProjectedColumn(1, fooSource, "foo", "a", "foo", "a", false),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "(select foo.a from bar)",
					evaluator.NewSQLSubqueryExpr(
						true,
						false,
						evaluator.NewProjectStage(barSource,
							createProjectedColumn(1, fooSource, "foo", "a", "foo", "a", true)),
					),
				),
			)
		}}, {

		"select * from (select b2.d, b2.b from bar b1 inner join bar b2" +
			" on (b1.a=b2.b) group by 1, 2) t0 HAVING (sum(1) > 0 )",
		func() evaluator.PlanStage {
			b1Source := createMongoSource(2, "bar", "b1")
			b2Source := createMongoSource(2, "bar", "b2")
			subqueryAliasName := "t0"

			matcher := evaluator.NewSQLComparisonExpr(
				evaluator.EQ,
				createSQLColumnExprFromSource(b1Source, "b1", "a", false),
				createSQLColumnExprFromSource(b2Source, "b2", "b", false),
			)

			join := evaluator.NewJoinStage(evaluator.InnerJoin, b1Source, b2Source, matcher)

			innerGroup := evaluator.NewGroupByStage(
				join,
				[]evaluator.SQLExpr{
					createSQLColumnExprFromSource(join, "b2", "d", false),
					createSQLColumnExprFromSource(join, "b2", "b", false),
				},
				evaluator.ProjectedColumns{
					createProjectedColumn(2, join, "b2", "b", "b2", "b", false),
					createProjectedColumn(2, join, "b2", "d", "b2", "d", false),
				},
			)

			subquery := evaluator.NewSubquerySourceStage(
				evaluator.NewProjectStage(
					innerGroup,
					createProjectedColumn(2, join, "b2", "d", "b2", "d", false),
					createProjectedColumn(2, join, "b2", "b", "b2", "b", false),
				),
				2,
				"test",
				subqueryAliasName,
				false,
			)

			outerGroup := evaluator.NewGroupByStage(
				subquery,
				nil,
				evaluator.ProjectedColumns{
					createProjectedColumn(2, subquery,
						subqueryAliasName, "d", subqueryAliasName, "d", false),
					createProjectedColumn(2, subquery,
						subqueryAliasName, "b", subqueryAliasName, "b", false),
					evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(1)",
						evaluator.NewSQLAggregationFunctionExpr(
							"sum",
							false,
							[]evaluator.SQLExpr{evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1))},
						),
					),
				},
			)

			filter := evaluator.NewFilterStage(
				outerGroup,
				evaluator.NewSQLComparisonExpr(
					evaluator.LT,
					evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 0)),
					testSQLColumnExpr(1, "", "", "sum(1)",
						types.EvalDouble, schema.MongoNone, false),
				),
			)

			project := evaluator.NewProjectStage(
				filter,
				createProjectedColumn(1, subquery,
					subqueryAliasName, "d", subqueryAliasName, "d", false),
				createProjectedColumn(1, subquery,
					subqueryAliasName, "b", subqueryAliasName, "b", false),
			)

			return project
		}},
	}
	runTestsAsSubtest("Select Correlated Subqueries", correlatedSubqueriesTests)

	// Simple Where Tests
	simpleWhereTests := []planTest{{
		"select a from foo where a",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewFilterStage(source,
					testSQLColumnExpr(1, defaultDbName, "foo", "a",
						types.EvalInt64, schema.MongoInt, false)),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
			)
		}}, {

		"select a from foo where false",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewFilterStage(source, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, false))),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
			)
		}}, {

		"select a from foo where true",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewFilterStage(source, evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true))),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
			)
		}}, {

		"select a from foo where g = true",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewFilterStage(source,
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						testSQLColumnExpr(1, defaultDbName, "foo", "g",
							types.EvalBoolean, schema.MongoBool, false),
						evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)),
					),
				),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
			)
		}}, {

		"select a from foo where a > 10",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewFilterStage(source,
					evaluator.NewSQLComparisonExpr(
						evaluator.LT,
						evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 10)),
						testSQLColumnExpr(1, defaultDbName, "foo", "a",
							types.EvalInt64, schema.MongoInt, false),
					),
				),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
			)
		}}, {

		"select a as b from foo where b > 10",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewFilterStage(source,
					evaluator.NewSQLComparisonExpr(
						evaluator.LT,
						evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 10)),
						testSQLColumnExpr(1, defaultDbName, "foo", "b",
							types.EvalInt64, schema.MongoInt, false),
					),
				),
				createProjectedColumn(1, source, "foo", "a", "foo", "b", false),
			)
		}},
	}
	runTestsAsSubtest("Select Simple Where", simpleWhereTests)

	// Subqueries with Where Tests
	subqueriesWithWhereTests := []planTest{{
		"select a from foo where (b) = (select b from bar where foo.a = bar.a)",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			dualSource := createMongoDualSource(2)
			barSource := createMongoSource(3, "bar", "bar")
			return evaluator.NewProjectStage(
				evaluator.NewFilterStage(
					fooSource,
					evaluator.NewSQLSubqueryCmpExpr(
						true, true,
						evaluator.NewProjectStage(
							dualSource,
							createProjectedColumn(1, fooSource, "foo", "b", "foo", "b", true),
						),
						evaluator.NewProjectStage(
							evaluator.NewFilterStage(
								barSource,
								evaluator.NewSQLComparisonExpr(
									evaluator.EQ,
									createSQLColumnExprFromSource(
										fooSource, "foo", "a", true),
									createSQLColumnExprFromSource(
										barSource, "bar", "a", false),
								),
							),
							createProjectedColumn(3, barSource, "bar", "b", "bar", "b", false),
						),
						"=",
					),
				),
				createProjectedColumn(1, fooSource, "foo", "a", "foo", "a", false),
			)
		}}, {

		"select a from foo f where (b) = (select b from bar where " +
			"exists(select 1 from foo where f.a = a))",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "f")
			dualSource := createMongoDualSource(2)
			barSource := createMongoSource(3, "bar", "bar")
			foo4Source := createMongoSource(4, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewFilterStage(
					fooSource,
					evaluator.NewSQLSubqueryCmpExpr(
						true, true,
						evaluator.NewProjectStage(
							dualSource,
							createProjectedColumn(1, fooSource, "f", "b", "f", "b", true),
						),
						evaluator.NewProjectStage(
							evaluator.NewFilterStage(
								barSource,
								evaluator.NewSQLExistsExpr(
									true,
									evaluator.NewProjectStage(
										evaluator.NewFilterStage(
											foo4Source,
											evaluator.NewSQLComparisonExpr(
												evaluator.EQ,
												createSQLColumnExprFromSource(
													fooSource, "f", "a", true),
												createSQLColumnExprFromSource(
													foo4Source, "foo", "a", false),
											),
										),
										evaluator.CreateProjectedColumnFromSQLExpr(4,
											"1", evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1))),
									),
								),
							),
							createProjectedColumn(3, barSource, "bar", "b", "bar", "b", false),
						),
						"=",
					),
				),
				createProjectedColumn(1, fooSource, "f", "a", "f", "a", false),
			)
		}}, {

		"select a from foo where (b) = (select b from bar where " +
			"exists(select 1 from foo where bar.a = a))",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			dualSource := createMongoDualSource(2)
			barSource := createMongoSource(3, "bar", "bar")
			foo3Source := createMongoSource(4, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewFilterStage(
					fooSource,
					evaluator.NewSQLSubqueryCmpExpr(
						true, false,
						evaluator.NewProjectStage(
							dualSource,
							createProjectedColumn(1, fooSource, "foo", "b", "foo", "b", true),
						),
						evaluator.NewProjectStage(
							evaluator.NewFilterStage(
								barSource,
								evaluator.NewSQLExistsExpr(
									true,
									evaluator.NewProjectStage(
										evaluator.NewFilterStage(
											foo3Source,
											evaluator.NewSQLComparisonExpr(
												evaluator.EQ,
												createSQLColumnExprFromSource(
													barSource, "bar", "a", true),
												createSQLColumnExprFromSource(
													foo3Source, "foo", "a", false),
											),
										),
										evaluator.CreateProjectedColumnFromSQLExpr(4,
											"1", evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1))),
									),
								),
							),
							createProjectedColumn(3, barSource, "bar", "b", "bar", "b", false),
						),
						"=",
					),
				),
				createProjectedColumn(1, fooSource, "foo", "a", "foo", "a", false),
			)
		}},
	}
	runTestsAsSubtest("Select Subqueries with Where", subqueriesWithWhereTests)

	// Group by Tests
	groupByTests := []planTest{{
		"select sum(a) from foo",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewGroupByStage(source,
					nil,
					evaluator.ProjectedColumns{
						evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)",
							evaluator.NewSQLAggregationFunctionExpr(
								"sum",
								false,
								[]evaluator.SQLExpr{
									createSQLColumnExprFromSource(source, "foo", "a", false)},
							),
						),
					},
				),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(a)",
					testSQLColumnExpr(1, defaultDbName, "", "sum(test.foo.a)",
						types.EvalDouble, schema.MongoNone, false)),
			)
		}}, {

		"select sum(a) from foo group by b",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewGroupByStage(source,
					[]evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "b", false)},
					evaluator.ProjectedColumns{
						evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)",
							evaluator.NewSQLAggregationFunctionExpr(
								"sum",
								false,
								[]evaluator.SQLExpr{
									createSQLColumnExprFromSource(source, "foo", "a", false)},
							)),
					},
				),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(a)",
					testSQLColumnExpr(1, defaultDbName, "", "sum(test.foo.a)",
						types.EvalDouble, schema.MongoNone, false)),
			)
		}}, {

		"select a, sum(a) from foo group by b",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewGroupByStage(source,
					[]evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "b", false)},
					evaluator.ProjectedColumns{
						createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)",
							evaluator.NewSQLAggregationFunctionExpr(
								"sum",
								false,
								[]evaluator.SQLExpr{
									createSQLColumnExprFromSource(source, "foo", "a", false)},
							)),
					},
				),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(a)",
					testSQLColumnExpr(1, defaultDbName, "", "sum(test.foo.a)",
						types.EvalDouble, schema.MongoNone, false)),
			)
		}}, {

		"select sum(a) from foo group by b order by sum(a)",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewOrderByStage(
					evaluator.NewGroupByStage(source,
						[]evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "b", false)},
						evaluator.ProjectedColumns{
							evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)",
								evaluator.NewSQLAggregationFunctionExpr(
									"sum",
									false,
									[]evaluator.SQLExpr{
										createSQLColumnExprFromSource(source, "foo", "a", false)},
								)),
						},
					),
					evaluator.NewOrderByTerm(
						testSQLColumnExpr(1, defaultDbName, "", "sum(test.foo.a)",
							types.EvalDouble, schema.MongoNone, false), true),
				),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(a)",
					testSQLColumnExpr(1, defaultDbName, "", "sum(test.foo.a)",
						types.EvalDouble, schema.MongoNone, false)),
			)
		}}, {

		"select sum(a) as sum_a from foo group by b order by sum_a",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewOrderByStage(
					evaluator.NewGroupByStage(source,
						[]evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "b", false)},
						evaluator.ProjectedColumns{
							evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)",
								evaluator.NewSQLAggregationFunctionExpr(
									"sum",
									false,
									[]evaluator.SQLExpr{
										createSQLColumnExprFromSource(source, "foo", "a", false)},
								)),
						},
					),
					evaluator.NewOrderByTerm(
						testSQLColumnExpr(1, defaultDbName, "", "sum(test.foo.a)",
							types.EvalDouble, schema.MongoNone, false), true),
				),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "sum_a",
					testSQLColumnExpr(1, defaultDbName, "", "sum(test.foo.a)",
						types.EvalDouble, schema.MongoNone, false)),
			)
		}}, {

		"select sum(a) from foo f group by b order by (select c from foo where f.b = b)",
		func() evaluator.PlanStage {
			foo1Source := createMongoSource(1, "foo", "f")
			foo2Source := createMongoSource(2, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewOrderByStage(
					evaluator.NewGroupByStage(foo1Source,
						[]evaluator.SQLExpr{createSQLColumnExprFromSource(foo1Source, "f", "b", false)},
						evaluator.ProjectedColumns{
							createProjectedColumn(1, foo1Source, "f", "b", "f", "b", false),
							evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.f.a)",
								evaluator.NewSQLAggregationFunctionExpr(
									"sum",
									false,
									[]evaluator.SQLExpr{
										createSQLColumnExprFromSource(foo1Source, "f", "a", false)},
								)),
						},
					),
					evaluator.NewOrderByTerm(
						evaluator.NewSQLSubqueryExpr(
							true,
							false,
							evaluator.NewProjectStage(
								evaluator.NewFilterStage(
									foo2Source,
									evaluator.NewSQLComparisonExpr(
										evaluator.EQ,
										createSQLColumnExprFromSource(
											foo1Source, "f", "b", true),
										createSQLColumnExprFromSource(
											foo2Source, "foo", "b", false),
									),
								),
								createProjectedColumn(2, foo2Source, "foo", "c", "foo", "c", false),
							),
						),
						true,
					),
				),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(a)",
					testSQLColumnExpr(1, defaultDbName, "", "sum(test.f.a)",
						types.EvalDouble, schema.MongoNone, false)),
			)
		}}, {

		"select (select sum(foo.a) from foo as f) from foo group by b",
		func() evaluator.PlanStage {
			foo1Source := createMongoSource(1, "foo", "foo")
			foo2Source := createMongoSource(2, "foo", "f")
			return evaluator.NewProjectStage(
				evaluator.NewGroupByStage(foo1Source,
					[]evaluator.SQLExpr{createSQLColumnExprFromSource(foo1Source, "foo", "b", false)},
					evaluator.ProjectedColumns{
						evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)",
							evaluator.NewSQLAggregationFunctionExpr(
								"sum",
								false,
								[]evaluator.SQLExpr{
									createSQLColumnExprFromSource(foo1Source, "foo", "a", false)},
							)),
					},
				),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "(select sum(foo.a) from foo as f)",
					evaluator.NewSQLSubqueryExpr(
						true,
						false,
						evaluator.NewProjectStage(
							foo2Source,
							evaluator.CreateProjectedColumnFromSQLExpr(2, "sum(foo.a)",
								testSQLColumnExpr(1, defaultDbName, "", "sum(test.foo.a)",
									types.EvalDouble, schema.MongoNone, true)),
						),
					),
				),
			)
		}}, {

		"select (select sum(f.a + foo.a) from foo f) from foo group by b",
		func() evaluator.PlanStage {
			foo1Source := createMongoSource(1, "foo", "foo")
			foo2Source := createMongoSource(2, "foo", "f")
			return evaluator.NewProjectStage(
				evaluator.NewGroupByStage(foo1Source,
					[]evaluator.SQLExpr{createSQLColumnExprFromSource(foo1Source, "foo", "b", false)},
					evaluator.ProjectedColumns{
						createProjectedColumn(1, foo1Source, "foo", "a", "foo", "a", false),
					},
				),
				evaluator.CreateProjectedColumnFromSQLExpr(1,
					"(select sum(f.a+foo.a) from foo as f)",
					evaluator.NewSQLSubqueryExpr(
						true,
						false,
						evaluator.NewProjectStage(
							evaluator.NewGroupByStage(
								foo2Source,
								nil,
								evaluator.ProjectedColumns{
									evaluator.CreateProjectedColumnFromSQLExpr(2,
										"sum(test.f.a+test.foo.a)",
										evaluator.NewSQLAggregationFunctionExpr(
											"sum",
											false,
											[]evaluator.SQLExpr{
												evaluator.NewSQLAddExpr(
													createSQLColumnExprFromSource(
														foo2Source, "f", "a", false),
													createSQLColumnExprFromSource(
														foo1Source, "foo", "a", true),
												)},
										)),
								},
							),
							evaluator.CreateProjectedColumnFromSQLExpr(2, "sum(f.a+foo.a)",
								testSQLColumnExpr(2, defaultDbName, "",
									"sum(test.f.a+test.foo.a)",
									types.EvalDouble, schema.MongoNone, false)),
						),
					),
				),
			)
		}},
	}
	runTestsAsSubtest("Select Group By Queries", groupByTests)

	// Having Tests
	havingTests := []planTest{{
		"select a from foo group by b having sum(a) > 10",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewFilterStage(
					evaluator.NewGroupByStage(source,
						[]evaluator.SQLExpr{
							createSQLColumnExprFromSource(source, "foo", "b", false)},
						evaluator.ProjectedColumns{
							createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
							evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)",
								evaluator.NewSQLAggregationFunctionExpr(
									"sum",
									false,
									[]evaluator.SQLExpr{
										createSQLColumnExprFromSource(
											source, "foo", "a", false)},
								)),
						},
					),
					evaluator.NewSQLComparisonExpr(
						evaluator.LT,
						evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 10)),
						testSQLColumnExpr(1, defaultDbName, "", "sum(test.foo.a)",
							types.EvalDouble, schema.MongoNone, false),
					),
				),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
			)
		}}, {

		"select a from foo having exists(select 1 from bar)",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(2, "bar", "bar")
			return evaluator.NewProjectStage(
				evaluator.NewFilterStage(
					fooSource,
					evaluator.NewSQLExistsExpr(
						false,
						evaluator.NewProjectStage(
							barSource,
							evaluator.CreateProjectedColumnFromSQLExpr(2, "1",
								evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1))),
						),
					),
				),
				createProjectedColumn(1, fooSource, "foo", "a", "foo", "a", false),
			)
		}},
	}
	runTestsAsSubtest("Select Having Queries", havingTests)

	// Distinct Tests
	distinctTests := []planTest{{
		"select distinct a from foo",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewGroupByStage(source,
					[]evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "a", false)},
					evaluator.ProjectedColumns{
						createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
					},
				),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
			)
		}}, {

		"select distinct sum(a) from foo",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewGroupByStage(
					evaluator.NewGroupByStage(source,
						nil,
						evaluator.ProjectedColumns{
							evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)",
								evaluator.NewSQLAggregationFunctionExpr(
									"sum",
									false,
									[]evaluator.SQLExpr{
										createSQLColumnExprFromSource(
											source, "foo", "a", false)},
								)),
						},
					),
					[]evaluator.SQLExpr{testSQLColumnExpr(1, defaultDbName, "",
						"sum(test.foo.a)", types.EvalDouble, schema.MongoNone, false)},
					evaluator.ProjectedColumns{
						evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)",
							testSQLColumnExpr(1, defaultDbName, "", "sum(test.foo.a)",
								types.EvalDouble, schema.MongoNone, false)),
					},
				),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(a)",
					testSQLColumnExpr(1, defaultDbName, "", "sum(test.foo.a)",
						types.EvalDouble, schema.MongoNone, false)),
			)
		}}, {

		"select distinct sum(a) from foo having sum(a) > 20",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewGroupByStage(
					evaluator.NewFilterStage(
						evaluator.NewGroupByStage(source,
							nil,
							evaluator.ProjectedColumns{
								evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)",
									evaluator.NewSQLAggregationFunctionExpr(
										"sum",
										false,
										[]evaluator.SQLExpr{
											createSQLColumnExprFromSource(
												source, "foo", "a", false)},
									)),
							},
						),
						evaluator.NewSQLComparisonExpr(
							evaluator.LT,
							evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 20)),
							testSQLColumnExpr(1, defaultDbName, "", "sum(test.foo.a)",
								types.EvalDouble, schema.MongoNone, false),
						),
					),
					[]evaluator.SQLExpr{testSQLColumnExpr(1, defaultDbName, "",
						"sum(test.foo.a)", types.EvalDouble, schema.MongoNone, false)},
					evaluator.ProjectedColumns{
						evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)",
							testSQLColumnExpr(1, defaultDbName, "", "sum(test.foo.a)",
								types.EvalDouble,
								schema.MongoNone, false)),
					},
				),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(a)",
					testSQLColumnExpr(1, defaultDbName, "", "sum(test.foo.a)",
						types.EvalDouble, schema.MongoNone, false)),
			)
		}},
	}
	runTestsAsSubtest("Select Distinct Queries", distinctTests)

	// Select Straight_Join Tests
	selectStraightJoinTests := []planTest{{
		"select straight_join foo.a from foo join bar",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(1, "bar", "bar")
			join := evaluator.NewJoinStage(evaluator.StraightJoin, fooSource, barSource,
				evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			return evaluator.NewProjectStage(join, createProjectedColumn(1, join, "foo",
				"a", "foo", "a", false))
		}}, {

		"select straight_join foo.a from foo cross join bar",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(1, "bar", "bar")
			join := evaluator.NewJoinStage(evaluator.StraightJoin, fooSource, barSource,
				evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			return evaluator.NewProjectStage(join, createProjectedColumn(1, join, "foo",
				"a", "foo", "a", false))
		}}, {

		"select straight_join foo.a from foo, bar",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(1, "bar", "bar")
			join := evaluator.NewJoinStage(evaluator.StraightJoin, fooSource, barSource,
				evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			return evaluator.NewProjectStage(join, createProjectedColumn(1, join, "foo",
				"a", "foo", "a", false))
		}}, {

		"select straight_join bar.a from bar natural join baz",
		func() evaluator.PlanStage {
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			join := evaluator.NewJoinStage(evaluator.StraightJoin, barSource, bazSource,
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLAndExpr(
						evaluator.NewSQLComparisonExpr(
							evaluator.EQ,
							createSQLColumnExprFromSource(barSource, "bar", "_id", false),
							createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
						),
						evaluator.NewSQLComparisonExpr(
							evaluator.EQ,
							createSQLColumnExprFromSource(barSource, "bar", "a", false),
							createSQLColumnExprFromSource(bazSource, "baz", "a", false),
						),
					),
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "b", false),
						createSQLColumnExprFromSource(bazSource, "baz", "b", false),
					),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, barSource, "bar", "a", "bar", "a", false))
		}}, {

		"select straight_join bar.a from foo join buzz cross join bar natural join baz",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barbazSource := createMongoSource(1, "buzz", "buzz")
			barSource := createMongoSource(1, "bar", "bar")
			bazSource := createMongoSource(1, "baz", "baz")
			join1 := evaluator.NewJoinStage(evaluator.StraightJoin, fooSource, barbazSource,
				evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			join2 := evaluator.NewJoinStage(evaluator.StraightJoin, barSource, bazSource,
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLAndExpr(
						evaluator.NewSQLComparisonExpr(
							evaluator.EQ,
							createSQLColumnExprFromSource(barSource, "bar", "_id", false),
							createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
						),
						evaluator.NewSQLComparisonExpr(
							evaluator.EQ,
							createSQLColumnExprFromSource(barSource, "bar", "a", false),
							createSQLColumnExprFromSource(bazSource, "baz", "a", false),
						),
					),
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						createSQLColumnExprFromSource(barSource, "bar", "b", false),
						createSQLColumnExprFromSource(bazSource, "baz", "b", false),
					),
				),
			)
			join3 := evaluator.NewJoinStage(evaluator.StraightJoin, join1, join2,
				evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)))
			return evaluator.NewProjectStage(join3,
				createProjectedColumn(1, barSource, "bar", "a", "bar", "a", false))
		}}, {

		// In the following tests, the joins should NOT be translated to straight_joins
		"select straight_join baz.b from baz natural left join buzz",
		func() evaluator.PlanStage {
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(1, "buzz", "buzz")
			join := evaluator.NewJoinStage(evaluator.LeftJoin, bazSource, buzzSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
					createSQLColumnExprFromSource(buzzSource, "buzz", "_id", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, bazSource, "baz", "b", "baz", "b", false))
		}}, {

		"select straight_join baz.b from baz natural right join buzz",
		func() evaluator.PlanStage {
			bazSource := createMongoSource(1, "baz", "baz")
			buzzSource := createMongoSource(1, "buzz", "buzz")
			join := evaluator.NewJoinStage(evaluator.RightJoin, bazSource, buzzSource,
				evaluator.NewSQLComparisonExpr(
					evaluator.EQ,
					createSQLColumnExprFromSource(bazSource, "baz", "_id", false),
					createSQLColumnExprFromSource(buzzSource, "buzz", "_id", false),
				),
			)
			return evaluator.NewProjectStage(join,
				createProjectedColumn(1, bazSource, "baz", "b", "baz", "b", false))
		}},
	}
	runTestsAsSubtest("Select Straight_Join Queries", selectStraightJoinTests)

	// Order by Tests
	orderByTests := []planTest{{
		"select a from foo order by a",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewOrderByStage(source,
					evaluator.NewOrderByTerm(
						testSQLColumnExpr(1,
							defaultDbName, "foo", "a",
							types.EvalInt64, schema.MongoInt, false),
						true,
					),
				),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
			)
		}}, {

		"select a as b from foo order by b",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewOrderByStage(source,
					evaluator.NewOrderByTerm(
						testSQLColumnExpr(1,
							defaultDbName, "foo", "a",
							types.EvalInt64, schema.MongoInt, false),
						true,
					),
				),
				createProjectedColumn(1, source, "foo", "a", "foo", "b", false),
			)
		}}, {

		"select a from foo order by foo.a",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewOrderByStage(source,
					evaluator.NewOrderByTerm(
						testSQLColumnExpr(1, defaultDbName,
							"foo", "a", types.EvalInt64, schema.MongoInt, false),
						true,
					),
				),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
			)
		}}, {

		"select a as b from foo order by foo.a",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewOrderByStage(source,
					evaluator.NewOrderByTerm(
						testSQLColumnExpr(1, defaultDbName,
							"foo", "a", types.EvalInt64, schema.MongoInt, false),
						true,
					),
				),
				createProjectedColumn(1, source, "foo", "a", "foo", "b", false),
			)
		}}, {

		"select a from foo order by 1",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewOrderByStage(source,
					evaluator.NewOrderByTerm(
						testSQLColumnExpr(1, defaultDbName,
							"foo", "a", types.EvalInt64, schema.MongoInt, false),
						true,
					),
				),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
			)
		}}, {

		"select * from foo order by 2",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewOrderByStage(source,
					evaluator.NewOrderByTerm(
						testSQLColumnExpr(1, defaultDbName,
							"foo", "a", types.EvalInt64, schema.MongoInt, false),
						true,
					),
				),
				createAllProjectedColumnsFromSource(1, source, "foo")...,
			)
		}}, {

		"select foo.* from foo order by 2",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewOrderByStage(source,
					evaluator.NewOrderByTerm(
						testSQLColumnExpr(1, defaultDbName,
							"foo", "a", types.EvalInt64, schema.MongoInt, false),
						true,
					),
				),
				createAllProjectedColumnsFromSource(1, source, "foo")...,
			)
		}}, {

		"select foo.*, foo.a from foo order by 2",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			columns := append(createAllProjectedColumnsFromSource(1, source, "foo"),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false))
			return evaluator.NewProjectStage(
				evaluator.NewOrderByStage(source,
					evaluator.NewOrderByTerm(
						testSQLColumnExpr(1, defaultDbName,
							"foo", "a", types.EvalInt64, schema.MongoInt, false),
						true,
					),
				),
				columns...,
			)
		}}, {

		"select a from foo order by -1",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewOrderByStage(source,
					evaluator.NewOrderByTerm(
						evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, -1)),
						true,
					),
				),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
			)
		}}, {

		"select a + b as c from foo order by c - b",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewOrderByStage(source,
					evaluator.NewOrderByTerm(
						evaluator.NewSQLSubtractExpr(
							evaluator.NewSQLAddExpr(
								testSQLColumnExpr(1, defaultDbName,
									"foo", "a", types.EvalInt64, schema.MongoInt, false),
								testSQLColumnExpr(1, defaultDbName,
									"foo", "b", types.EvalInt64, schema.MongoInt, false),
							),
							testSQLColumnExpr(1, defaultDbName,
								"foo", "b", types.EvalInt64, schema.MongoInt, false),
						),
						true,
					),
				),
				evaluator.CreateProjectedColumnFromSQLExpr(1, "c",
					evaluator.NewSQLAddExpr(
						testSQLColumnExpr(1, defaultDbName,
							"foo", "a", types.EvalInt64, schema.MongoInt, false),
						testSQLColumnExpr(1, defaultDbName,
							"foo", "b", types.EvalInt64, schema.MongoInt, false),
					),
				),
			)
		}},
	}
	runTestsAsSubtest("Select Order by Queries", orderByTests)

	// Order by Subqueries Tests
	orderBySubqueriesTests := []planTest{{
		"select a from foo order by (select a from bar)",
		func() evaluator.PlanStage {
			fooSource := createMongoSource(1, "foo", "foo")
			barSource := createMongoSource(2, "bar", "bar")
			return evaluator.NewProjectStage(
				evaluator.NewOrderByStage(
					fooSource,
					evaluator.NewOrderByTerm(
						evaluator.NewSQLSubqueryExpr(
							false,
							false,
							evaluator.NewProjectStage(
								barSource,
								createProjectedColumn(2, barSource, "bar", "a", "bar", "a", false),
							),
						),
						true,
					),
				),
				createProjectedColumn(1, fooSource, "foo", "a", "foo", "a", false),
			)
		}},
	}
	runTestsAsSubtest("Select Order By Subquery", orderBySubqueriesTests)

	// Limit Tests
	limitTests := []planTest{{
		"select a from foo limit 10",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewLimitStage(source, 0, 10),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
			)
		}}, {

		"select a from foo limit 10, 20",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewLimitStage(source, 10, 20),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
			)
		}}, {

		"select a from foo limit 10,0",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewEmptyStage([]*results.Column{
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false).Column,
			}, collation.Default)
		}}, {

		"select a from foo limit 0, 0",
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewEmptyStage([]*results.Column{
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false).Column,
			}, collation.Default)
		}},
	}
	runTestsAsSubtest("Select Limit (Plan)", limitTests)

	limitVariableTests := []variableTest{{
		"select a from foo",
		func() *variable.Container {
			vars := &variable.Container{}
			setSystemVariable(vars, variable.MongoDBGitVersion, testInfo.GitVersion)
			setSystemVariable(vars, variable.MongoDBVersion, testInfo.Version)
			setSystemVariable(vars, variable.MongoDBVersionCompatibility, testInfo.CompatibleVersion)
			setSystemVariable(vars, variable.SQLSelectLimit, 10)
			setSystemVariable(vars, variable.TypeConversionMode, "mysql")
			setSystemVariable(vars, variable.PolymorphicTypeConversionMode, "off")
			return vars
		},
		func() *mongodb.Info {
			return testInfo
		},
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewLimitStage(source, 0, 10),
				createProjectedColumn(1, source, "foo", "a", "foo", "a", false),
			)
		}}, {

		"select b from foo",
		func() *variable.Container {
			vars := &variable.Container{}
			setSystemVariable(vars, variable.MongoDBGitVersion, testInfo.GitVersion)
			setSystemVariable(vars, variable.MongoDBVersion, testInfo.Version)
			setSystemVariable(vars, variable.MongoDBVersionCompatibility, testInfo.CompatibleVersion)
			setSystemVariable(vars, variable.SQLSelectLimit, uint64(18446744073709551615))
			setSystemVariable(vars, variable.TypeConversionMode, "mysql")
			setSystemVariable(vars, variable.PolymorphicTypeConversionMode, "off")
			return vars
		},
		func() *mongodb.Info {
			return testInfo
		},
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				source,
				createProjectedColumn(1, source, "foo", "b", "foo", "b", false),
			)
		}}, {

		"select b from foo fast",
		func() *variable.Container {
			vars := &variable.Container{}
			setSystemVariable(vars, variable.MongoDBGitVersion, testInfo.GitVersion)
			setSystemVariable(vars, variable.MongoDBVersion, testInfo.Version)
			setSystemVariable(vars, variable.MongoDBVersionCompatibility, testInfo.CompatibleVersion)
			setSystemVariable(vars, variable.SQLSelectLimit, uint64(18446744073709551615))
			setSystemVariable(vars, variable.TypeConversionMode, "mysql")
			setSystemVariable(vars, variable.PolymorphicTypeConversionMode, "fast")
			return vars
		},
		func() *mongodb.Info {
			return testInfo
		},
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "fast")
			return evaluator.NewProjectStage(
				source,
				createProjectedColumn(1, source, "fast", "b", "fast", "b", false),
			)
		}}, {

		"select b from foo safe",
		func() *variable.Container {
			vars := &variable.Container{}
			setSystemVariable(vars, variable.MongoDBGitVersion, testInfo.GitVersion)
			setSystemVariable(vars, variable.MongoDBVersion, testInfo.Version)
			setSystemVariable(vars, variable.MongoDBVersionCompatibility, testInfo.CompatibleVersion)
			setSystemVariable(vars, variable.SQLSelectLimit, uint64(18446744073709551615))
			setSystemVariable(vars, variable.TypeConversionMode, "mysql")
			setSystemVariable(vars, variable.PolymorphicTypeConversionMode, "safe")
			return vars
		},
		func() *mongodb.Info {
			return testInfo
		},
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "safe")
			projectedColumn := createProjectedColumn(1, source, "safe", "b", "safe", "b", false)
			projectedColumn.Expr = evaluator.NewSQLConvertExpr(projectedColumn.Expr,
				types.EvalInt64)
			projectedColumn.MongoType = ""
			projectedColumn.OriginalName = ""
			projectedColumn.OriginalTable = ""
			projectedColumn.Table = ""
			return evaluator.NewProjectStage(
				source,
				projectedColumn,
			)
		}}, {

		"select b from foo limit 10, 20",
		func() *variable.Container {
			vars := &variable.Container{}
			setSystemVariable(vars, variable.MongoDBGitVersion, testInfo.GitVersion)
			setSystemVariable(vars, variable.MongoDBVersion, testInfo.Version)
			setSystemVariable(vars, variable.MongoDBVersionCompatibility, testInfo.CompatibleVersion)
			setSystemVariable(vars, variable.SQLSelectLimit, 5)
			setSystemVariable(vars, variable.TypeConversionMode, "mysql")
			setSystemVariable(vars, variable.PolymorphicTypeConversionMode, "off")
			return vars
		},
		func() *mongodb.Info {
			return testInfo
		},
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "foo")
			return evaluator.NewProjectStage(
				evaluator.NewLimitStage(source, 10, 20),
				createProjectedColumn(1, source, "foo", "b", "foo", "b", false),
			)
		}}, {

		"select b from foo fast limit 10, 20",
		func() *variable.Container {
			vars := &variable.Container{}
			setSystemVariable(vars, variable.MongoDBGitVersion, testInfo.GitVersion)
			setSystemVariable(vars, variable.MongoDBVersion, testInfo.Version)
			setSystemVariable(vars, variable.MongoDBVersionCompatibility, testInfo.CompatibleVersion)
			setSystemVariable(vars, variable.SQLSelectLimit, 5)
			setSystemVariable(vars, variable.TypeConversionMode, "mysql")
			setSystemVariable(vars, variable.PolymorphicTypeConversionMode, "fast")
			return vars
		},
		func() *mongodb.Info {
			return testInfo
		},
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "fast")
			return evaluator.NewProjectStage(
				evaluator.NewLimitStage(source, 10, 20),
				createProjectedColumn(1, source, "fast", "b", "fast", "b", false),
			)
		}}, {

		"select b from foo safe limit 10, 20",
		func() *variable.Container {
			vars := &variable.Container{}
			setSystemVariable(vars, variable.MongoDBGitVersion, testInfo.GitVersion)
			setSystemVariable(vars, variable.MongoDBVersion, testInfo.Version)
			setSystemVariable(vars, variable.MongoDBVersionCompatibility, testInfo.CompatibleVersion)
			setSystemVariable(vars, variable.SQLSelectLimit, 5)
			setSystemVariable(vars, variable.TypeConversionMode, "mysql")
			setSystemVariable(vars, variable.PolymorphicTypeConversionMode, "safe")
			return vars
		},
		func() *mongodb.Info {
			return testInfo
		},
		func() evaluator.PlanStage {
			source := createMongoSource(1, "foo", "safe")
			projectedColumn := createProjectedColumn(1, source, "safe", "b", "safe", "b", false)
			projectedColumn.Expr = evaluator.NewSQLConvertExpr(projectedColumn.Expr,
				types.EvalInt64)
			projectedColumn.MongoType = ""
			projectedColumn.OriginalName = ""
			projectedColumn.OriginalTable = ""
			projectedColumn.Table = ""
			return evaluator.NewProjectStage(
				evaluator.NewLimitStage(source, 10, 20),
				projectedColumn,
			)
		}},
	}
	runTestsAsSubtest("Select Limit (Variables)", limitVariableTests)

	// Count Tests
	countTests := []planTest{{
		"select count(*) from foo",
		func() evaluator.PlanStage {
			column := results.NewColumn(1, "", "", "", "count(*)", "", "",
				types.EvalInt64, schema.MongoNone, false, true)
			projectedColumn := createProjectedColumnFromColumn(1, column, "", "count(*)", false)
			source := createMongoSource(1, "foo", "foo")
			countStage := evaluator.NewCountStage(source.(*evaluator.MongoSourceStage),
				projectedColumn)
			return evaluator.NewProjectStage(countStage, projectedColumn)
		}}, {

		"select count(*) as c from foo",
		func() evaluator.PlanStage {
			column := results.NewColumn(1, "", "", "", "c", "", "",
				types.EvalInt64, schema.MongoNone, false, true)

			projectedColumn := createProjectedColumnFromColumn(1, column, "", "c", false)
			source := createMongoSource(1, "foo", "foo")
			countStage := evaluator.NewCountStage(source.(*evaluator.MongoSourceStage),
				projectedColumn)
			return evaluator.NewProjectStage(countStage,
				projectedColumn)
		}}, {

		"select count(*) as c from foo order by a",
		func() evaluator.PlanStage {
			column := results.NewColumn(1, "", "", "", "c", "", "",
				types.EvalInt64, schema.MongoNone, false, true)

			projectedColumn := createProjectedColumnFromColumn(1, column, "", "c", false)
			source := createMongoSource(1, "foo", "foo")
			countStage := evaluator.NewCountStage(source.(*evaluator.MongoSourceStage),
				projectedColumn)
			return evaluator.NewProjectStage(countStage,
				projectedColumn)
		}}, {

		"select count(*) as c from foo order by 1",
		func() evaluator.PlanStage {
			column := results.NewColumn(1, "", "", "", "c", "", "",
				types.EvalInt64, schema.MongoNone, false, true)

			projectedColumn := createProjectedColumnFromColumn(1, column, "", "c", false)
			source := createMongoSource(1, "foo", "foo")
			countStage := evaluator.NewCountStage(source.(*evaluator.MongoSourceStage),
				projectedColumn)
			return evaluator.NewProjectStage(countStage,
				projectedColumn)
		}}, {

		"select count(*) from foo as c",
		func() evaluator.PlanStage {
			column := results.NewColumn(1, "", "", "", "count(*)", "", "",
				types.EvalInt64, schema.MongoNone, false, true)

			projectedColumn := createProjectedColumnFromColumn(1, column, "", "count(*)", false)
			source := createMongoSource(1, "foo", "c")
			countStage := evaluator.NewCountStage(source.(*evaluator.MongoSourceStage),
				projectedColumn)
			return evaluator.NewProjectStage(countStage,
				projectedColumn)
		}},
	}
	runTestsAsSubtest("Select Count", countTests)

	// Error Tests
	errorTests := []errorTest{
		{"select ABASDD()",
			"scalar function 'abasdd' is not supported"},
		{"select a",
			`ERROR 1054 (42S22): Unknown column 'a' in 'field list'`},
		{"select a from idk",
			`ERROR 1146 (42S02): Table 'test.idk' doesn't exist`},
		{"select idk from foo",
			`ERROR 1054 (42S22): Unknown column 'idk' in 'field list'`},
		{"select f.a from foo",
			`ERROR 1054 (42S22): Unknown column 'f.a' in 'field list'`},
		{"select foo.a from foo f",
			`ERROR 1054 (42S22): Unknown column 'foo.a' in 'field list'`},
		{"select a + idk from foo",
			`ERROR 1054 (42S22): Unknown column 'idk' in 'field list'`},

		{"select a from foo, bar",
			`ERROR 1052 (23000): Column 'a' in field list is ambiguous`},
		{"select foo.a from foo f, bar b",
			`ERROR 1054 (42S22): Unknown column 'foo.a' in 'field list'`},
		{"select a from foo f, bar b",
			`ERROR 1052 (23000): Column 'a' in field list is ambiguous`},
		{"select a, b as a from foo order by a",
			`ERROR 1052 (23000): Column 'a' in order clause is ambiguous`},
		{"select a as b, b from foo order by b",
			`ERROR 1052 (23000): Column 'b' in order clause is ambiguous`},
		{"select a as z, b as z from foo order by z",
			`ERROR 1052 (23000): Column 'z' in order clause is ambiguous`},

		{"select (select a, b from foo) from foo",
			`ERROR 1241 (21000): Operand should contain 1 column(s)`},
		{"select * from (select a, b as a from foo) f",
			`ERROR 1060 (42S21): Duplicate column name 'f.a'`},
		{"select foo.a from (select a from foo)",
			`ERROR 1248 (42000): Every derived table must have its own alias`},

		{"select a from foo limit -10",
			`ERROR 1149 (42000): Rowcount cannot be negative`},
		{"select a from foo limit -10, 20",
			`ERROR 1149 (42000): Offset cannot be negative`},
		{"select a from foo limit -10, -20",
			`ERROR 1149 (42000): Offset cannot be negative`},
		{"select a from foo limit b",
			`ERROR 1691 (HY000): A variable of a non-integer based type in LIMIT clause`},
		{"select a from foo limit 'c'",
			`ERROR 1691 (HY000): A variable of a non-integer based type in LIMIT clause`},

		{"select a from foo, (select * from (select * from bar where foo.b = b) asdf) wegqweg",
			`ERROR 1054 (42S22): Unknown column 'foo.b' in 'where clause'`},
		{"select a from foo where sum(a) = 10",
			`ERROR 1111 (HY000): Invalid use of group function`},

		{"select a from foo order by 0",
			`ERROR 1054 (42S22): Unknown column '0' in 'order clause'`},
		{"select a from foo order by 2",
			`ERROR 1054 (42S22): Unknown column '2' in 'order clause'`},
		{"select a from foo order by idk",
			`ERROR 1054 (42S22): Unknown column 'idk' in 'order clause'`},

		{"select sum(a) from foo group by sum(a)",
			`ERROR 1056 (42000): Can't group on 'sum(test.foo.a)'`},
		{"select sum(a) from foo group by (a + sum(a))",
			`ERROR 1056 (42000): Can't group on 'sum(test.foo.a)'`},
		{"select sum(a) from foo group by 1",
			`ERROR 1056 (42000): Can't group on 'sum(test.foo.a)'`},
		{"select sum(a) as d from foo group by d",
			`ERROR 1056 (42000): Can't group on 'd'`},
		{"select count(*) as d from foo group by d",
			`ERROR 1056 (42000): Can't group on 'd'`},
		{"select a+sum(a) as d from foo group by d",
			`ERROR 1056 (42000): Can't group on 'd'`},
		{"select a+count(*) as d from foo group by d",
			`ERROR 1056 (42000): Can't group on 'd'`},
		{"select a+sum(a) from foo group by 1",
			`ERROR 1056 (42000): Can't group on 'sum(test.foo.a)'`},
		{"select sum(a) from foo group by 0",
			`ERROR 1054 (42S22): Unknown column '0' in 'group clause'`},
		{"select sum(a) from foo group by 2",
			`ERROR 1054 (42S22): Unknown column '2' in 'group clause'`},
		{"select a as z, b as z from foo group by z",
			`ERROR 1052 (23000): Column 'z' in group clause is ambiguous`},

		{"select a from foo, foo",
			`ERROR 1066 (42000): Not unique table/alias: 'foo'`},
		{"select a from foo as bar, bar",
			`ERROR 1066 (42000): Not unique table/alias: 'bar'`},
		{"select a from foo as g, foo as g",
			`ERROR 1066 (42000): Not unique table/alias: 'g'`},

		{"select a from foo left outer join bar where a = 10",
			`ERROR 1064 (42000): A left join requires criteria`},

		{"select * from bar natural join baz using (id)",
			"ERROR 1064 (42000): A natural join cannot have join criteria"},
		{"select * from bar natural join baz on bar.id=baz.id",
			"ERROR 1064 (42000): A natural join cannot have join criteria"},
		{"select * from foo natural left join bar using (id)",
			"ERROR 1064 (42000): A natural left join cannot have join criteria"},
		{"select * from foo natural right join bar using (id)",
			"ERROR 1064 (42000): A natural right join cannot have join criteria"},

		{"select bar.d, baz.a from bar join baz using (tomato)",
			`ERROR 1054 (42S22): Unknown column 'bar.tomato' in 'from clause'`},
		{"select * from baz join bar using (d)",
			`ERROR 1054 (42S22): Unknown column 'baz.d' in 'from clause'`},
		{"select bar.d, baz.a from bar join (select * from baz join foo) using (c)",
			`ERROR 1248 (42000): Every derived table must have its own alias`},
		{"select bar.d, biz.a from bar join (select * from baz join foo) as biz using (c)",
			`ERROR 1060 (42S21): Duplicate column name 'biz._id'`},
		{"select * from bar join foo join baz using (c)",
			"ERROR 1054 (42S22): Unknown column 'baz.c' in 'from clause'"},
		{"select * from bar join foo join baz using (_id)",
			"ERROR 1052 (23000): Column '_id' in from clause is ambiguous"},
		{"select * from baz join bar join foo using (c)",
			"ERROR 1054 (42S22): Unknown column 'c' in 'from clause'"},

		{"select * from (foo join bar) natural join baz",
			"ERROR 1052 (23000): Column '_id' in from clause is ambiguous"},
		{"select * from foo join bar using (b) natural join baz",
			"ERROR 1052 (23000): Column '_id' in from clause is ambiguous"},
		{"select * from bar where _id in (select _id from foo join baz)",
			"ERROR 1052 (23000): Column '_id' in field list is ambiguous"},
		{"select bar.d, biz.a from bar natural join (select * from baz join foo) as biz",
			`ERROR 1060 (42S21): Duplicate column name 'biz._id'`},
		{"select * from foo left join bar natural join baz using (id)",
			"ERROR 1064 (42000): A natural join cannot have join criteria"},

		{"select a like 'h_l##r' escape '##' from foo",
			`ERROR 1210 (HY000): Incorrect arguments to ESCAPE`},
	}
	runTestsAsSubtest("Error Tests", errorTests)
}

func TestAlgebrizeCommand(t *testing.T) {

	type test struct {
		sql                 string
		expectedPlanFactory func() evaluator.Command
	}

	type failureTest struct {
		sql           string
		expectedError string
	}

	runTest := func(t *testing.T, testCase test) {
		t.Run(testCase.sql, func(t *testing.T) {
			req := require.New(t)
			statement, err := parser.Parse(testCase.sql)
			req.Nil(err, "failed to parse")

			rCfg := evaluator.NewRewriterConfig(42, "algebrizer_unit_test_dbname", log.GlobalLogger(),
				false, algebrizerUnitTestVersion, "algebrizer_unit_test_remoteHost", "algebrizer_unit_test_user")

			rewritten, err := evaluator.RewriteStatement(rCfg, statement)
			req.Nil(err, "failed to rewrite query")

			// run tests for algebrizing commands in --writeMode, except for the failure tests.
			aCfg := evaluator.NewAlgebrizerConfig(log.GlobalLogger(), defaultDbName, testCatalog, true)

			actual, err := evaluator.AlgebrizeCommand(aCfg, rewritten)

			req.Nil(err, "failed to algebrize")

			expected := testCase.expectedPlanFactory()
			req.Equal(expected, actual, "actual does not match expected")
		})
	}

	runFailureTest := func(t *testing.T, testCase failureTest, writeMode bool) {
		t.Run(testCase.sql, func(t *testing.T) {
			req := require.New(t)
			statement, err := parser.Parse(testCase.sql)
			req.Nil(err, "failed to parse")

			rCfg := evaluator.NewRewriterConfig(42, "algebrizer_unit_test_dbname", log.GlobalLogger(),
				false, algebrizerUnitTestVersion, "algebrizer_unit_test_remoteHost", "algebrizer_unit_test_user")

			rewritten, err := evaluator.RewriteStatement(rCfg, statement)
			req.Nil(err, "failed to rewrite query")

			aCfg := evaluator.NewAlgebrizerConfig(log.GlobalLogger(), defaultDbName, testCatalog, writeMode)

			_, err = evaluator.AlgebrizeCommand(aCfg, rewritten)

			req.NotNil(err)
			req.Equal(testCase.expectedError, err.Error())
		})
	}

	runTestsAsSubtest := func(subTestName string, tests []test) {
		t.Run(subTestName, func(t *testing.T) {
			for _, testCase := range tests {
				runTest(t, testCase)
			}
		})

	}

	runFailureTestsAsSubtest := func(subTestName string, tests []failureTest, writeMode bool) {
		t.Run(subTestName, func(t *testing.T) {
			for _, testCase := range tests {
				runFailureTest(t, testCase, writeMode)
			}
		})

	}

	// Test Algebrizing Kill Statements.
	killTests := []test{{
		"kill 3",
		func() evaluator.Command {
			return evaluator.NewKillCommand(
				evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 3)),
				evaluator.KillConnection,
			)
		}}, {

		"kill query 3",
		func() evaluator.Command {
			return evaluator.NewKillCommand(evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 3)), evaluator.KillQuery)
		}}, {

		"kill query 5*3",
		func() evaluator.Command {
			return evaluator.NewKillCommand(
				evaluator.NewSQLMultiplyExpr(
					evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 5)),
					evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 3)),
				),
				evaluator.KillQuery,
			)
		}}, {

		"kill connection 5-3",
		func() evaluator.Command {
			return evaluator.NewKillCommand(
				evaluator.NewSQLSubtractExpr(
					evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 5)),
					evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 3)),
				),
				evaluator.KillConnection,
			)
		}},
	}
	runTestsAsSubtest("Kill Statements", killTests)

	// Test Algebrizing Flush Statements.
	flushTests := []test{{
		"flush logs",
		func() evaluator.Command {
			return evaluator.NewFlushCommand(evaluator.FlushLogs)
		}}, {

		"flush sample",
		func() evaluator.Command {
			return evaluator.NewFlushCommand(evaluator.FlushSample)
		}},
	}
	runTestsAsSubtest("Flush Statements", flushTests)

	// Test Algebrizing Create Database.
	createDBTests := []test{
		{
			"create database if not exists foo",
			func() evaluator.Command {
				return evaluator.NewCreateDatabaseCommand(testCatalog, "foo", true)
			},
		},
		{
			"create database foo",
			func() evaluator.Command {
				return evaluator.NewCreateDatabaseCommand(testCatalog, "foo", false)
			},
		},
	}
	runTestsAsSubtest("Create Database", createDBTests)

	// Test Algebrizing Drop Database.
	dropDBTests := []test{
		{
			"drop database if exists foo",
			func() evaluator.Command {
				return evaluator.NewDropDatabaseCommand(testCatalog, "foo", true)
			},
		},
		{
			"drop database foo",
			func() evaluator.Command {
				return evaluator.NewDropDatabaseCommand(testCatalog, "foo", false)
			},
		},
	}
	runTestsAsSubtest("Drop Database", dropDBTests)

	// Test Algebrizing Drop Table.
	dropTableTests := []test{
		{
			"drop table if exists bar.foo",
			func() evaluator.Command {
				return evaluator.NewDropTableCommand(testCatalog, "bar", "foo", true)
			},
		},
		{
			"drop table bar.foo",
			func() evaluator.Command {
				return evaluator.NewDropTableCommand(testCatalog, "bar", "foo", false)
			},
		},
		{
			"drop table if exists foo",
			func() evaluator.Command {
				return evaluator.NewDropTableCommand(testCatalog, "test", "foo", true)
			},
		},
		{
			"drop table foo",
			func() evaluator.Command {
				return evaluator.NewDropTableCommand(testCatalog, "test", "foo", false)
			},
		},
	}
	runTestsAsSubtest("Drop Table", dropTableTests)

	// Test Algebrizing Create Table.
	createTableTests := []test{
		{
			// This tests essentially everything interesting in one complex example.
			"create table if not exists bar.foo(" +
				"a int not null unique, " +
				"b text not null comment 'fooo', " +
				"c tinytext comment 'HELLO!', " +
				"unique index bar(a asc, b desc), " +
				"fulltext index(b, c)) comment = 'WORLD'",
			func() evaluator.Command {
				return evaluator.NewCreateTableCommand(testCatalog, "bar",
					testTable(
						"foo",
						"foo",
						[]bson.D{},
						[]*schema.Column{
							schema.NewColumn("a", schema.SQLInt, "a", schema.MongoInt64, false, option.NoneString()),
							schema.NewColumn("b", schema.SQLVarchar, "b", schema.MongoString, false, option.SomeString("fooo")),
							schema.NewColumn("c", schema.SQLVarchar, "c", schema.MongoString, true, option.SomeString("HELLO!")),
						},
						[]schema.Index{
							schema.NewIndex("bar", true, false,
								[]schema.IndexPart{schema.NewIndexPart("a", 1), schema.NewIndexPart("b", -1)},
							),
							schema.NewIndex("", false, true,
								[]schema.IndexPart{schema.NewIndexPart("b", 1), schema.NewIndexPart("c", 1)},
							),
							schema.NewIndex("a_unique", true, false,
								[]schema.IndexPart{schema.NewIndexPart("a", 1)}),
						},
						option.SomeString("WORLD"),
					),
					true)
			},
		}, {
			// Also test that datetime is rewritten to timestamp, and without `if not exists`
			"create table bar.foo(x int, y datetime)",
			func() evaluator.Command {
				return evaluator.NewCreateTableCommand(testCatalog, "bar",
					testTable(
						"foo",
						"foo",
						[]bson.D{},
						[]*schema.Column{
							schema.NewColumn("x", schema.SQLInt, "x", schema.MongoInt64, true,
								option.NoneString()),
							schema.NewColumn("y", schema.SQLTimestamp, "y", schema.MongoDate, true,
								option.NoneString()),
						},
						[]schema.Index{},
						option.NoneString(),
					),
					false)
			},
		}, {
			// Test without database name, add in a unique and test that bit is rewritten to bool.
			// This also shows that we can declare something as null, optionally.
			"create table foo(x bit null unique)",
			func() evaluator.Command {
				return evaluator.NewCreateTableCommand(testCatalog, "test",
					testTable(
						"foo",
						"foo",
						[]bson.D{},
						[]*schema.Column{
							schema.NewColumn("x", schema.SQLBoolean, "x", schema.MongoBool, true,
								option.NoneString()),
						},
						[]schema.Index{
							schema.NewIndex("x_unique", true, false,
								[]schema.IndexPart{schema.NewIndexPart("x", 1)}),
						},
						option.NoneString(),
					),
					false)
			},
		},
	}
	runTestsAsSubtest("Create Table", createTableTests)

	// Insert tests. New scope is to constrain variable lifetimes.
	{
		req := require.New(t)
		testDB, err := testCatalog.Database("test")
		req.NoError(err)
		fooTable, err := testDB.Table("foo")
		req.NoError(err)
		fooCols := fooTable.Columns()
		nullRow := make(evaluator.SQLExprs, len(fooCols))
		basicPositionMap := make(map[string]int, len(fooCols))
		for i, col := range fooTable.Columns() {
			nullRow[i] = evaluator.NewSQLValueExpr(values.NewSQLNull(values.MySQLValueKind))
			basicPositionMap[col.MongoName] = i
		}
		gcPositionMap := map[string]int{
			"g": 0,
			"c": 1,
		}
		cgPositionMap := map[string]int{
			"c": 0,
			"g": 1,
		}
		testInts := make(evaluator.SQLExprs, 5)
		for i := range testInts {
			testInts[i] = evaluator.NewSQLValueExpr(values.NewSQLInt64(values.MySQLValueKind, int64(i)))
		}
		testTrue := evaluator.NewSQLValueExpr(values.NewSQLBool(values.MySQLValueKind, true))
		testHello := evaluator.NewSQLValueExpr(values.NewSQLVarchar(values.MySQLValueKind, "hello"))
		insertTests := []test{
			{
				"insert into foo values()",
				func() evaluator.Command {
					return evaluator.NewInsertCommand("test", "foo", fooTable.Columns(), basicPositionMap, [][]evaluator.SQLExpr{nullRow})
				},
			},
			{
				"insert into foo values(),()",
				func() evaluator.Command {
					return evaluator.NewInsertCommand("test", "foo", fooTable.Columns(), basicPositionMap,
						[][]evaluator.SQLExpr{nullRow, nullRow})
				},
			},
			{
				"insert into foo values(0,1,2,3,4,true,'hello')",
				func() evaluator.Command {
					return evaluator.NewInsertCommand("test", "foo", fooTable.Columns(), basicPositionMap,
						[][]evaluator.SQLExpr{
							{
								testInts[0],
								testInts[1],
								testInts[2],
								testInts[3],
								testInts[4],
								testTrue,
								testHello,
							},
						})
				},
			},
			{
				"insert into foo values(0,1,2,3,4,true,'hello'), (0,1,2,3,4,true,'hello')",
				func() evaluator.Command {
					return evaluator.NewInsertCommand("test", "foo", fooTable.Columns(), basicPositionMap,
						[][]evaluator.SQLExpr{
							{
								testInts[0],
								testInts[1],
								testInts[2],
								testInts[3],
								testInts[4],
								testTrue,
								testHello,
							},
							{
								testInts[0],
								testInts[1],
								testInts[2],
								testInts[3],
								testInts[4],
								testTrue,
								testHello,
							},
						})
				},
			},
			{
				"insert into foo(g,c) values(true, 3)",
				func() evaluator.Command {
					return evaluator.NewInsertCommand("test", "foo", fooTable.Columns(), gcPositionMap,
						[][]evaluator.SQLExpr{
							{
								testTrue,
								testInts[3],
							},
						})
				},
			},
			{
				"insert into foo(g,c) values(true, 3), (true, 0)",
				func() evaluator.Command {
					return evaluator.NewInsertCommand("test", "foo", fooTable.Columns(), gcPositionMap,
						[][]evaluator.SQLExpr{
							{
								testTrue,
								testInts[3],
							},
							{
								testTrue,
								testInts[0],
							},
						})
				},
			},
			{
				"insert into foo(c,g) values(3, true)",
				func() evaluator.Command {
					return evaluator.NewInsertCommand("test", "foo", fooTable.Columns(), cgPositionMap,
						[][]evaluator.SQLExpr{
							{
								testInts[3],
								testTrue,
							},
						})
				},
			},
			{
				"insert into foo(c,g) values(3, true), (4, true)",
				func() evaluator.Command {
					return evaluator.NewInsertCommand("test", "foo", fooTable.Columns(), cgPositionMap,
						[][]evaluator.SQLExpr{
							{
								testInts[3],
								testTrue,
							},
							{
								testInts[4],
								testTrue,
							},
						})
				},
			},
		}

		runTestsAsSubtest("Insert", insertTests)
	}

	insertFailuresTests := []failureTest{
		{
			sql:           "insert into foo values(42, 23)",
			expectedError: "ERROR 1136 (21S01): Column count doesn't match value count at row 1",
		},
		{
			sql:           "insert into foo(c) values(42, 23)",
			expectedError: "ERROR 1136 (21S01): Column count doesn't match value count at row 1",
		},
		{
			sql:           "insert into foo(c, g) values(42, 23), (41)",
			expectedError: "ERROR 1136 (21S01): Column count doesn't match value count at row 2",
		},
		{
			sql:           "insert into foo(c, g) values(42, 23), (41, true), ()",
			expectedError: "ERROR 1136 (21S01): Column count doesn't match value count at row 3",
		},
	}
	runFailureTestsAsSubtest("Insert Failures", insertFailuresTests, true)

	writesOutOfWriteModeTests := []failureTest{
		{
			sql:           "create table foo(x int)",
			expectedError: "create table requires --writeMode",
		},
		{
			sql:           "create database foo",
			expectedError: "create database requires --writeMode",
		},
		{
			sql:           "drop database foo",
			expectedError: "drop database requires --writeMode",
		},
		{
			sql:           "insert into foo values()",
			expectedError: "insert requires --writeMode",
		},
		// We allow `drop table` out of --writeMode to support Tableau.
	}

	runFailureTestsAsSubtest("DDL and Insert Must Fail Out of --writeMode", writesOutOfWriteModeTests, false)
}

func testTable(tbl, col string,
	pipeline []bson.D, cols []*schema.Column,
	indexes []schema.Index, comment option.String) *schema.Table {
	lg := log.NewComponentLogger(log.SchemaComponent, log.GlobalLogger())
	out, err := schema.NewTable(lg, tbl, col, pipeline, cols, indexes, comment)
	if err != nil {
		panic("this table should not error")
	}
	return out
}

func TestAlgebrizeExpr(t *testing.T) {
	testDB, _ := testCatalog.Database("test")
	table, _ := testDB.Table("foo")
	fooTable, _ := table.(catalog.MongoDBTable)
	source := evaluator.NewMongoSourceStage(testDB, fooTable, 1, "foo")

	type test struct {
		sql      string
		expected evaluator.SQLExpr
		version  []uint8
	}

	runTest := func(t *testing.T, testCase test) {
		t.Run(testCase.sql, func(t *testing.T) {
			req := require.New(t)
			statement, err := parser.Parse("select " + testCase.sql + " from foo")
			req.Nil(err, "failed to parse")

			rCfg := evaluator.NewRewriterConfig(42, "algebrizer_unit_test_dbname", log.GlobalLogger(),
				false, algebrizerUnitTestVersion, "algebrizer_unit_test_remoteHost", "algebrizer_unit_test_user")

			rewritten, err := evaluator.RewriteStatement(rCfg, statement)
			req.Nil(err, "failed to rewrite query")

			testCatalog = evaluator.GetCatalog(
				testSchema, evaluator.CreateTestVariables(
					evaluator.GetMongoDBInfo(testCase.version, testSchema, mongodb.AllPrivileges),
				), testInfo)

			aCfg := evaluator.NewAlgebrizerConfig(log.GlobalLogger(), "test", testCatalog, false)

			actual, err := evaluator.AlgebrizeQuery(aCfg, rewritten)
			actualExpr := actual.(*evaluator.ProjectStage).ProjectedColumns()[0].Expr
			req.Nil(err, "failed to algebrize")
			req.Equal(testCase.expected, actualExpr, "actual does not match expected")
		})
	}

	type errorTest struct {
		sql      string
		expected string
	}

	runErrorTest := func(t *testing.T, testCase errorTest) {
		t.Run(testCase.sql, func(t *testing.T) {
			req := require.New(t)
			statement, err := parser.Parse("select " + testCase.sql + " from foo")
			req.Nil(err, "failed to parse")

			rCfg := evaluator.NewRewriterConfig(42, "algebrizer_unit_test_dbname", log.GlobalLogger(),
				false, algebrizerUnitTestVersion, "algebrizer_unit_test_remoteHost", "algebrizer_unit_test_user")

			rewritten, err := evaluator.RewriteStatement(rCfg, statement)
			req.Nil(err, "failed to rewrite query")

			aCfg := evaluator.NewAlgebrizerConfig(log.GlobalLogger(), "test", testCatalog, false)

			_, err = evaluator.AlgebrizeQuery(aCfg, rewritten)

			req.NotNil(err, "successfully algebrized when it should have failed")

			req.Equal(err.Error(), testCase.expected, "actual does not match expected")
		})
	}

	runTestsAsSubtest := func(subTestName string, tests interface{}) {
		t.Run(subTestName, func(t *testing.T) {
			switch typedTests := tests.(type) {
			case []test:
				for _, testCase := range typedTests {
					runTest(t, testCase)
				}
			case []errorTest:
				for _, testCase := range typedTests {
					runErrorTest(t, testCase)
				}
			}
		})
	}

	createSQLColumnExpr := func(columnName string) evaluator.SQLColumnExpr {
		for _, c := range source.Columns() {
			if c.Name == columnName {
				return testSQLColumnExpr(1,
					c.Database, c.Table, c.Name, c.EvalType, c.MongoType, false)
			}
		}

		panic("column not found")
	}

	// Algebrize Expressions
	expectedDate := time.Date(2006, time.December, 31, 0, 0, 0, 0, time.UTC)
	expectedDate2 := time.Date(2014, 6, 7, 10, 32, 46, 5000, time.UTC)
	expectedDate3 := time.Date(0, 1, 1, 10, 32, 46, 5000, time.UTC)
	d, _ := decimal.NewFromString("100000000000000000000000000000000000")
	testVersion := []uint8{4, 0, 0}

	exprTests := []test{{
		"a = 1 AND b = 2",
		evaluator.NewSQLAndExpr(
			evaluator.NewSQLComparisonExpr(
				evaluator.EQ,
				createSQLColumnExpr("a"),
				evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
			),
			evaluator.NewSQLComparisonExpr(
				evaluator.EQ,
				createSQLColumnExpr("b"),
				evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 2)),
			),
		),
		testVersion,
	}, {
		"a + 1",
		evaluator.NewSQLAddExpr(
			createSQLColumnExpr("a"),
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
		),
		testVersion,
	}, {
		"DATE '2006-12-31'",
		evaluator.NewSQLValueExpr(values.NewSQLDate(valKind, expectedDate)),
		testVersion,
	}, {
		"DATE '06-12-31'",
		evaluator.NewSQLValueExpr(values.NewSQLDate(valKind, expectedDate)),
		testVersion,
	}, {
		"DATE '20061231'",
		evaluator.NewSQLValueExpr(values.NewSQLDate(valKind, expectedDate)),
		testVersion,
	}, {
		"DATE '061231'",
		evaluator.NewSQLValueExpr(values.NewSQLDate(valKind, expectedDate)),
		testVersion,
	}, {
		"a / 1",
		evaluator.NewSQLDivideExpr(
			createSQLColumnExpr("a"),
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
		),
		testVersion,
	}, {
		"a = 1",
		evaluator.NewSQLComparisonExpr(
			evaluator.EQ,
			createSQLColumnExpr("a"),
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
		),
		testVersion,
	}, {
		"g = 0",
		evaluator.NewSQLComparisonExpr(
			evaluator.EQ,
			createSQLColumnExpr("g"),
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 0)),
		),
		testVersion,
	}, {
		"g = 1",
		evaluator.NewSQLComparisonExpr(
			evaluator.EQ,
			createSQLColumnExpr("g"),
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
		),
		testVersion,
	}, {
		"g = 2",
		evaluator.NewSQLComparisonExpr(
			evaluator.EQ,
			createSQLColumnExpr("g"),
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 2)),
		),
		testVersion,
	}, {
		"0 = g",
		evaluator.NewSQLComparisonExpr(
			evaluator.EQ,
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 0)),
			createSQLColumnExpr("g"),
		),
		testVersion,
	}, {
		"1 = g",
		evaluator.NewSQLComparisonExpr(
			evaluator.EQ,
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
			createSQLColumnExpr("g"),
		),
		testVersion,
	}, {
		"2 = g",
		evaluator.NewSQLComparisonExpr(
			evaluator.EQ,
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 2)),
			createSQLColumnExpr("g"),
		),
		testVersion,
	}, {
		"a > 1",
		evaluator.NewSQLComparisonExpr(
			evaluator.LT,
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
			createSQLColumnExpr("a"),
		),
		testVersion,
	}, {
		"a >= 1",
		evaluator.NewSQLComparisonExpr(
			evaluator.LTE,
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
			createSQLColumnExpr("a"),
		),
		testVersion,
	}, {
		"a is true",
		evaluator.NewSQLIsExpr(
			createSQLColumnExpr("a"),
			evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)),
		),
		testVersion,
	}, {
		"a is not true",
		evaluator.NewSQLNotExpr(
			evaluator.NewSQLIsExpr(
				createSQLColumnExpr("a"),
				evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)),
			),
		),
		testVersion,
	}, {
		"a IS NULL",
		evaluator.NewSQLIsExpr(
			createSQLColumnExpr("a"),
			evaluator.NewSQLValueExpr(values.NewSQLNull(valKind)),
		),
		testVersion,
	}, {
		"a IS NOT NULL",
		evaluator.NewSQLNotExpr(
			evaluator.NewSQLIsExpr(
				createSQLColumnExpr("a"),
				evaluator.NewSQLValueExpr(values.NewSQLNull(valKind)),
			),
		),
		testVersion,
	}, {
		"a < 1",
		evaluator.NewSQLComparisonExpr(
			evaluator.LT,
			createSQLColumnExpr("a"),
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
		),
		testVersion,
	}, {
		"a <= 1",
		evaluator.NewSQLComparisonExpr(
			evaluator.LTE,
			createSQLColumnExpr("a"),
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
		),
		testVersion,
	}, {
		"a * 1",
		evaluator.NewSQLMultiplyExpr(
			createSQLColumnExpr("a"),
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
		),
		testVersion,
	}, {
		"NOT a",
		evaluator.NewSQLNotExpr(
			createSQLColumnExpr("a"),
		),
		testVersion,
	}, {
		"a != 1",
		evaluator.NewSQLComparisonExpr(
			evaluator.NEQ,
			createSQLColumnExpr("a"),
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
		),
		testVersion,
	}, {
		"a <=> 1",
		evaluator.NewSQLNullSafeEqualsExpr(
			createSQLColumnExpr("a"),
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
		),
		testVersion,
	}, {
		"NULL",
		evaluator.NewSQLValueExpr(values.NewSQLNull(valKind)),
		testVersion,
	}, {
		"TRUE",
		evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, true)),
		testVersion,
	}, {
		"FALSE",
		evaluator.NewSQLValueExpr(values.NewSQLBool(valKind, false)),
		testVersion,
	}, {
		"20",
		evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 20)),
		testVersion,
	}, {
		"-20",
		evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, -20)),
		testVersion,
	}, {
		"202E-1",
		evaluator.NewSQLValueExpr(values.NewSQLFloat(valKind, 20.2)),
		testVersion,
	}, {
		"-202E-1",
		evaluator.NewSQLValueExpr(values.NewSQLFloat(valKind, -20.2)),
		testVersion,
	}, {
		"20.2",
		evaluator.NewSQLValueExpr(values.NewSQLDecimal128(valKind, decimal.New(202, -1))),
		testVersion,
	}, {
		"-20.2",
		evaluator.NewSQLValueExpr(values.NewSQLDecimal128(valKind, decimal.New(-202, -1))),
		testVersion,
	}, {
		"100000000000000000000000000000000000",
		evaluator.NewSQLValueExpr(values.NewSQLDecimal128(valKind, d)),
		testVersion,
	}, {
		"a = 1 OR b = 2",
		evaluator.NewSQLOrExpr(
			evaluator.NewSQLComparisonExpr(
				evaluator.EQ,
				createSQLColumnExpr("a"),
				evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
			),
			evaluator.NewSQLComparisonExpr(
				evaluator.EQ,
				createSQLColumnExpr("b"),
				evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 2)),
			),
		),
		testVersion,
	}, {
		"(1)",
		evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
		testVersion,
	}, {
		"a BETWEEN 0 AND 20",
		evaluator.NewSQLAndExpr(
			evaluator.NewSQLComparisonExpr(
				evaluator.LTE,
				evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 0)),
				createSQLColumnExpr("a"),
			),
			evaluator.NewSQLComparisonExpr(
				evaluator.LTE,
				createSQLColumnExpr("a"),
				evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 20)),
			),
		),
		testVersion,
	}, {
		"a NOT BETWEEN 0 AND 20",
		evaluator.NewSQLOrExpr(
			evaluator.NewSQLComparisonExpr(
				evaluator.LT,
				createSQLColumnExpr("a"),
				evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 0)),
			),
			evaluator.NewSQLComparisonExpr(
				evaluator.LT,
				evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 20)),
				createSQLColumnExpr("a"),
			),
		),
		testVersion,
	}, {
		"a - 1",
		evaluator.NewSQLSubtractExpr(
			createSQLColumnExpr("a"),
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
		),
		testVersion,
	}, {
		"TIMESTAMP '2014-06-07 10:32:46.000005'",
		evaluator.NewSQLValueExpr(values.NewSQLTimestamp(valKind, expectedDate2)),
		testVersion,
	}, {
		"TIMESTAMP '2014-6-7 10:32:46.000005'",
		evaluator.NewSQLValueExpr(values.NewSQLTimestamp(valKind, expectedDate2)),
		testVersion,
	}, {
		"TIMESTAMP '14-06-07 10:32:46.000005'",
		evaluator.NewSQLValueExpr(values.NewSQLTimestamp(valKind, expectedDate2)),
		testVersion,
	}, {
		"TIMESTAMP '14-6-7 10:32:46.000005'",
		evaluator.NewSQLValueExpr(values.NewSQLTimestamp(valKind, expectedDate2)),
		testVersion,
	}, {
		"TIMESTAMP '2014:06:07 10:32:46.000005'",
		evaluator.NewSQLValueExpr(values.NewSQLTimestamp(valKind, expectedDate2)),
		testVersion,
	}, {
		"TIMESTAMP '14:06:07 10:32:46.000005'",
		evaluator.NewSQLValueExpr(values.NewSQLTimestamp(valKind, expectedDate2)),
		testVersion,
	}, {
		"TIMESTAMP '20140607103246.000005'",
		evaluator.NewSQLValueExpr(values.NewSQLTimestamp(valKind, expectedDate2)),
		testVersion,
	}, {
		"TIMESTAMP '140607103246.000005'",
		evaluator.NewSQLValueExpr(values.NewSQLTimestamp(valKind, expectedDate2)),
		testVersion,
	}, {
		"TIMESTAMP '146.07103246.000005'",
		evaluator.NewSQLValueExpr(values.NewSQLTimestamp(valKind, expectedDate2)),
		testVersion,
	}, {
		"TIMESTAMP '14.06.07.10.32.46.000005'",
		evaluator.NewSQLValueExpr(values.NewSQLTimestamp(valKind, expectedDate2)),
		testVersion,
	}, {
		"(a)",
		createSQLColumnExpr("a"),
		testVersion,
	}, {
		"-a",
		evaluator.NewSQLUnaryMinusExpr(createSQLColumnExpr("a")),
		testVersion,
	}, {
		"-c",
		evaluator.NewSQLUnaryMinusExpr(createSQLColumnExpr("c")),
		testVersion,
	}, {
		"-g",
		evaluator.NewSQLUnaryMinusExpr(createSQLColumnExpr("g")),
		testVersion,
	}, {
		"-_id",
		evaluator.NewSQLUnaryMinusExpr(createSQLColumnExpr("_id")),
		testVersion,
	}, {
		"'a'",
		evaluator.NewSQLValueExpr(values.NewSQLVarchar(valKind, "a")),
		testVersion,
	}, {
		"~a",
		evaluator.NewSQLTildeExpr(createSQLColumnExpr("a")),
		testVersion,
	}, {
		"TIME '10:32:46.000005'",
		evaluator.NewSQLValueExpr(values.NewSQLTimestamp(valKind, expectedDate3)),
		testVersion,
	}, {
		"TIME '103246.000005'",
		evaluator.NewSQLValueExpr(values.NewSQLTimestamp(valKind, expectedDate3)),
		testVersion,
	}, {
		"benchmark(1, a)",
		evaluator.NewSQLBenchmarkExpr(
			evaluator.NewSQLValueExpr(values.NewSQLInt64(valKind, 1)),
			createSQLColumnExpr("a"),
		),
		testVersion,
	},
	}

	runTestsAsSubtest("Algebrize Expressions", exprTests)

	// 3.2.0 Tests
	fl, _ := strconv.ParseFloat("1000000000000000000000000000000000000", 64)
	threeTwoTests := []test{
		{
			"30.2",
			evaluator.NewSQLValueExpr(values.NewSQLFloat(valKind, 30.2)),
			[]uint8{3, 2, 0},
		}, {
			"-30.2",
			evaluator.NewSQLValueExpr(values.NewSQLFloat(valKind, -30.2)),
			[]uint8{3, 2, 0},
		}, {
			"1000000000000000000000000000000000000",
			evaluator.NewSQLValueExpr(values.NewSQLFloat(valKind, fl)),
			[]uint8{3, 2, 0},
		},
	}
	runTestsAsSubtest("Algebrize 3.2 Expressions", threeTwoTests)

	// Variable Tests
	varGlobal := evaluator.NewSQLVariableExpr("sql_auto_is_null",
		variable.SystemKind, variable.GlobalScope,
		values.NewSQLBool(values.VariableSQLValueKind, false))
	varSession := evaluator.NewSQLVariableExpr("sql_auto_is_null",
		variable.SystemKind, variable.SessionScope,
		values.NewSQLBool(values.VariableSQLValueKind, false))
	variableTests := []test{
		{
			"@@global.sql_auto_is_null",
			varGlobal,
			testVersion,
		}, {
			"@@session.sql_auto_is_null",
			varSession,
			testVersion,
		}, {
			"@@local.sql_auto_is_null",
			varSession,
			testVersion,
		}, {
			"@@sql_auto_is_null",
			varSession,
			testVersion,
		}, {
			"@hmmm",
			evaluator.NewSQLVariableExpr("hmmm", variable.UserKind,
				variable.SessionScope, values.NewSQLNull(values.VariableSQLValueKind)),
			testVersion,
		},
	}
	runTestsAsSubtest("Algebrize Variable Expressions", variableTests)

	errorTests := []errorTest{
		{
			"DATE '2014-13-07'",
			"ERROR 1525 (HY000): Incorrect DATE value: '2014-13-07'",
		}, {
			"DATE '2014-12-32'",
			"ERROR 1525 (HY000): Incorrect DATE value: '2014-12-32'",
		}, {
			"DATE '2006-12-31 10:32:46'",
			"ERROR 1525 (HY000): Incorrect DATE value: '2006-12-31 10:32:46'",
		}, {
			"TIME '2014-12-32'",
			"ERROR 1525 (HY000): Incorrect TIME value: '2014-12-32'",
		}, {
			"TIME '2006-12-31 10:32:46.000005'",
			"ERROR 1525 (HY000): Incorrect TIME value: '2006-12-31 10:32:46.000005'",
		}, {
			"TIMESTAMP '2014-06-07'",
			"ERROR 1525 (HY000): Incorrect DATETIME value: '2014-06-07'",
		},
	}
	runTestsAsSubtest("Algebrize Expression Errors", errorTests)
}

func BenchmarkAlgebrizeQuery(b *testing.B) {
	sch := evaluator.MustLoadSchema(testSchema4)
	info := evaluator.GetMongoDBInfo([]uint8{3, 2}, sch, mongodb.AllPrivileges)
	vars := evaluator.CreateTestVariables(info)
	ctlg := evaluator.GetCatalog(sch, vars, info)

	bench := func(name, sql string) {
		statement, err := parser.Parse(sql)
		if err != nil {
			b.Fatal(err)
		}

		rCfg := evaluator.NewRewriterConfig(42, "algebrizer_unit_test_dbname", log.GlobalLogger(),
			false, algebrizerUnitTestVersion, "algebrizer_unit_test_remoteHost", "algebrizer_unit_test_user")
		if err != nil {
			b.Fatal(err)
		}

		rewritten, err := evaluator.RewriteStatement(rCfg, statement)
		if err != nil {
			b.Fatal(err)
		}

		aCfg := evaluator.NewAlgebrizerConfig(log.GlobalLogger(), defaultDbName, ctlg, false)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_, err = evaluator.AlgebrizeQuery(aCfg, rewritten)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	bench("subquery", "select a, b from (select a, b from bar) b")
	bench("join", "select * from bar a join foo b on a.a=b.a and a.a=b.f")
	bench("subquery_join", "select * from (select foo.a from bar join (select foo.a from foo) foo"+
		" on foo.a=bar.b) x join (select g.a from bar join "+
		"(select foo.a from foo) g on g.a=bar.a) y on x.a=y.a")
}
