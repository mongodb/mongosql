package evaluator_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	"github.com/kr/pretty"
	"github.com/shopspring/decimal"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAlgebrizeQuery(t *testing.T) {

	testSchema := evaluator.MustLoadSchema(testSchema1)
	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)
	testVars := evaluator.CreateTestVariables(testInfo)
	testVars.SetSystemVariable(variable.MongoDBMaxVarcharLength, 10)
	testCatalog := evaluator.GetCatalogFromSchema(testSchema, testVars)
	defaultDbName := "test"

	test := func(sql string, expectedPlanFactory func() evaluator.PlanStage) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			actual, err := evaluator.AlgebrizeQuery(statement, defaultDbName, testVars, testCatalog)
			So(err, ShouldBeNil)

			expected := expectedPlanFactory()

			if ShouldResemble(actual, expected) != "" {
				fmt.Printf("\nSQL: %s", sql)
				fmt.Printf("\nExpected: %# v", pretty.Formatter(expected))
				fmt.Printf("\nActual: %# v", pretty.Formatter(actual))
			}

			So(actual, ShouldResemble, expected)
		})
	}

	testVariables := func(sql string, container func() *variable.Container, expectedPlanFactory func() evaluator.PlanStage) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)
			vars := container()
			actual, err := evaluator.AlgebrizeQuery(statement, defaultDbName, vars, testCatalog)
			So(err, ShouldBeNil)

			expected := expectedPlanFactory()

			if ShouldResemble(actual, expected) != "" {
				fmt.Printf("\nSQL: %s", sql)
				fmt.Printf("\nExpected: %# v", pretty.Formatter(expected))
				fmt.Printf("\nActual: %# v", pretty.Formatter(actual))
			}

			So(actual, ShouldResemble, expected)
		})
	}

	testError := func(sql, message string) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			actual, err := evaluator.AlgebrizeQuery(statement, defaultDbName, testVars, testCatalog)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, message)
			So(actual, ShouldBeNil)
		})
	}

	createMongoSource := func(selectID int, tableName, aliasName string) evaluator.PlanStage {
		db, _ := testCatalog.Database(defaultDbName)
		table, _ := db.Table(tableName)
		r := evaluator.NewMongoSourceStage(db, table.(*catalog.MongoTable), selectID, aliasName)
		return r
	}

	Convey("Subject: AlgebrizeQuery", t, func() {
		Convey("Show Statements", func() {
			isDBName := "INFORMATION_SCHEMA"
			informationSchemaDB, _ := testCatalog.Database(isDBName)
			subqueryAliasName := "CHARACTER_SETS"
			Convey("charset", func() {
				tbl, err := informationSchemaDB.Table(subqueryAliasName)
				So(err, ShouldBeNil)
				source := evaluator.NewDynamicSourceStage(informationSchemaDB, tbl.(*catalog.DynamicTable), 2, subqueryAliasName)
				subquery := evaluator.NewSubquerySourceStage(
					evaluator.NewProjectStage(
						source,
						createProjectedColumn(2, source, subqueryAliasName, "CHARACTER_SET_NAME", subqueryAliasName, "Charset"),
						createProjectedColumn(2, source, subqueryAliasName, "DESCRIPTION", subqueryAliasName, "Description"),
						createProjectedColumn(2, source, subqueryAliasName, "DEFAULT_COLLATE_NAME", subqueryAliasName, "Default collation"),
						createProjectedColumn(2, source, subqueryAliasName, "MAXLEN", subqueryAliasName, "Maxlen"),
					),
					2,
					subqueryAliasName,
				)
				test("show charset", func() evaluator.PlanStage {
					return evaluator.NewOrderByStage(
						subquery,
						evaluator.NewOrderByTerm(createSQLColumnExprFromSource(subquery, subqueryAliasName, "Charset"), true),
					)
				})
				test("show charset like 'n'", func() evaluator.PlanStage {
					return evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLLikeExpr(createSQLColumnExprFromSource(subquery, subqueryAliasName, "Charset"), evaluator.SQLVarchar("n"), evaluator.SQLVarchar("\\")),
						),
						evaluator.NewOrderByTerm(createSQLColumnExprFromSource(subquery, subqueryAliasName, "Charset"), true),
					)
				})
				test("show charset where `Charset` = 'n'", func() evaluator.PlanStage {
					return evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLEqualsExpr(createSQLColumnExprFromSource(subquery, subqueryAliasName, "Charset"), evaluator.SQLVarchar("n")),
						),
						evaluator.NewOrderByTerm(createSQLColumnExprFromSource(subquery, subqueryAliasName, "Charset"), true),
					)
				})
			})

			Convey("collation", func() {
				subqueryAliasName = "COLLATIONS"
				tbl, err := informationSchemaDB.Table(subqueryAliasName)
				So(err, ShouldBeNil)
				source := evaluator.NewDynamicSourceStage(informationSchemaDB, tbl.(*catalog.DynamicTable), 2, subqueryAliasName)
				subquery := evaluator.NewSubquerySourceStage(
					evaluator.NewProjectStage(
						source,
						createProjectedColumn(2, source, subqueryAliasName, "COLLATION_NAME", subqueryAliasName, "Collation"),
						createProjectedColumn(2, source, subqueryAliasName, "CHARACTER_SET_NAME", subqueryAliasName, "Charset"),
						createProjectedColumn(2, source, subqueryAliasName, "ID", subqueryAliasName, "Id"),
						createProjectedColumn(2, source, subqueryAliasName, "IS_DEFAULT", subqueryAliasName, "Default"),
						createProjectedColumn(2, source, subqueryAliasName, "IS_COMPILED", subqueryAliasName, "Compiled"),
						createProjectedColumn(2, source, subqueryAliasName, "SORTLEN", subqueryAliasName, "Sortlen"),
					),
					2,
					subqueryAliasName,
				)
				test("show collation", func() evaluator.PlanStage {
					return evaluator.NewOrderByStage(
						subquery,
						evaluator.NewOrderByTerm(createSQLColumnExprFromSource(subquery, subqueryAliasName, "Collation"), true),
					)
				})
				test("show collation like 'n'", func() evaluator.PlanStage {
					return evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLLikeExpr(createSQLColumnExprFromSource(subquery, subqueryAliasName, "Collation"), evaluator.SQLVarchar("n"), evaluator.SQLVarchar("\\")),
						),
						evaluator.NewOrderByTerm(createSQLColumnExprFromSource(subquery, subqueryAliasName, "Collation"), true),
					)
				})
				test("show collation where `Collation` = 'n'", func() evaluator.PlanStage {
					return evaluator.NewOrderByStage(
						evaluator.NewFilterStage(
							subquery,
							evaluator.NewSQLEqualsExpr(createSQLColumnExprFromSource(subquery, subqueryAliasName, "Collation"), evaluator.SQLVarchar("n")),
						),
						evaluator.NewOrderByTerm(createSQLColumnExprFromSource(subquery, subqueryAliasName, "Collation"), true),
					)
				})
			})

			Convey("columns", func() {
				subqueryAliasName = "COLUMNS"
				tbl, err := informationSchemaDB.Table(subqueryAliasName)
				So(err, ShouldBeNil)
				source := evaluator.NewDynamicSourceStage(informationSchemaDB, tbl.(*catalog.DynamicTable), 2, subqueryAliasName)
				Convey("plain", func() {
					subquery := evaluator.NewSubquerySourceStage(
						evaluator.NewProjectStage(
							source,
							createProjectedColumn(2, source, subqueryAliasName, "COLUMN_NAME", subqueryAliasName, "Field"),
							createProjectedColumn(2, source, subqueryAliasName, "COLUMN_TYPE", subqueryAliasName, "Type"),
							createProjectedColumn(2, source, subqueryAliasName, "IS_NULLABLE", subqueryAliasName, "Null"),
							createProjectedColumn(2, source, subqueryAliasName, "COLUMN_KEY", subqueryAliasName, "Key"),
							createProjectedColumn(2, source, subqueryAliasName, "COLUMN_DEFAULT", subqueryAliasName, "Default"),
							createProjectedColumn(2, source, subqueryAliasName, "EXTRA", subqueryAliasName, "Extra"),
							createProjectedColumn(2, source, subqueryAliasName, "TABLE_NAME", subqueryAliasName, "TABLE_NAME"),
							createProjectedColumn(2, source, subqueryAliasName, "TABLE_SCHEMA", subqueryAliasName, "TABLE_SCHEMA"),
							createProjectedColumn(2, source, subqueryAliasName, "ORDINAL_POSITION", subqueryAliasName, "ORDINAL_POSITION"),
						),
						2,
						subqueryAliasName,
					)
					for _, from := range []string{"from foo", "from test.foo", "from foo from test", "in foo in test", "from foo in test", "in foo from test"} {
						test(fmt.Sprintf("show columns %s", from), func() evaluator.PlanStage {
							return evaluator.NewProjectStage(
								evaluator.NewOrderByStage(
									evaluator.NewFilterStage(
										subquery,
										evaluator.NewSQLAndExpr(
											evaluator.NewSQLEqualsExpr(createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_NAME"), evaluator.SQLVarchar("foo")),
											evaluator.NewSQLEqualsExpr(createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_SCHEMA"), evaluator.SQLVarchar(defaultDbName)),
										),
									),
									evaluator.NewOrderByTerm(createSQLColumnExprFromSource(subquery, subqueryAliasName, "ORDINAL_POSITION"), true),
								),
								createProjectedColumn(1, subquery, subqueryAliasName, "Field", subqueryAliasName, "Field"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Type", subqueryAliasName, "Type"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Null", subqueryAliasName, "Null"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Key", subqueryAliasName, "Key"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Default", subqueryAliasName, "Default"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Extra", subqueryAliasName, "Extra"),
							)
						})
						test(fmt.Sprintf("show columns %s like 'n'", from), func() evaluator.PlanStage {
							return evaluator.NewProjectStage(
								evaluator.NewOrderByStage(
									evaluator.NewFilterStage(
										subquery,
										evaluator.NewSQLAndExpr(
											evaluator.NewSQLAndExpr(
												evaluator.NewSQLEqualsExpr(
													createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_NAME"),
													evaluator.SQLVarchar("foo"),
												),
												evaluator.NewSQLEqualsExpr(
													createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_SCHEMA"),
													evaluator.SQLVarchar(defaultDbName),
												),
											),
											evaluator.NewSQLLikeExpr(
												createSQLColumnExprFromSource(subquery, subqueryAliasName, "Field"),
												evaluator.SQLVarchar("n"),
												evaluator.SQLVarchar("\\"),
											),
										),
									),
									evaluator.NewOrderByTerm(
										createSQLColumnExprFromSource(subquery, subqueryAliasName, "ORDINAL_POSITION"),
										true,
									),
								),
								createProjectedColumn(1, subquery, subqueryAliasName, "Field", subqueryAliasName, "Field"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Type", subqueryAliasName, "Type"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Null", subqueryAliasName, "Null"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Key", subqueryAliasName, "Key"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Default", subqueryAliasName, "Default"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Extra", subqueryAliasName, "Extra"),
							)
						})
						test(fmt.Sprintf("show columns %s where `Field` = 'n'", from), func() evaluator.PlanStage {
							return evaluator.NewProjectStage(
								evaluator.NewOrderByStage(
									evaluator.NewFilterStage(
										subquery,
										evaluator.NewSQLAndExpr(
											evaluator.NewSQLAndExpr(
												evaluator.NewSQLEqualsExpr(
													createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_NAME"),
													evaluator.SQLVarchar("foo"),
												),
												evaluator.NewSQLEqualsExpr(
													createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_SCHEMA"),
													evaluator.SQLVarchar(defaultDbName),
												),
											),
											evaluator.NewSQLEqualsExpr(
												createSQLColumnExprFromSource(subquery, subqueryAliasName, "Field"),
												evaluator.SQLVarchar("n"),
											),
										),
									),
									evaluator.NewOrderByTerm(
										createSQLColumnExprFromSource(subquery, subqueryAliasName, "ORDINAL_POSITION"),
										true,
									),
								),
								createProjectedColumn(1, subquery, subqueryAliasName, "Field", subqueryAliasName, "Field"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Type", subqueryAliasName, "Type"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Null", subqueryAliasName, "Null"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Key", subqueryAliasName, "Key"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Default", subqueryAliasName, "Default"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Extra", subqueryAliasName, "Extra"),
							)
						})
					}
				})

				Convey("full", func() {
					subquery := evaluator.NewSubquerySourceStage(
						evaluator.NewProjectStage(
							source,
							createProjectedColumn(2, source, subqueryAliasName, "COLUMN_NAME", subqueryAliasName, "Field"),
							createProjectedColumn(2, source, subqueryAliasName, "COLUMN_TYPE", subqueryAliasName, "Type"),
							createProjectedColumn(2, source, subqueryAliasName, "COLLATION_NAME", subqueryAliasName, "Collation"),
							createProjectedColumn(2, source, subqueryAliasName, "IS_NULLABLE", subqueryAliasName, "Null"),
							createProjectedColumn(2, source, subqueryAliasName, "COLUMN_KEY", subqueryAliasName, "Key"),
							createProjectedColumn(2, source, subqueryAliasName, "COLUMN_DEFAULT", subqueryAliasName, "Default"),
							createProjectedColumn(2, source, subqueryAliasName, "EXTRA", subqueryAliasName, "Extra"),
							createProjectedColumn(2, source, subqueryAliasName, "PRIVILEGES", subqueryAliasName, "Privileges"),
							createProjectedColumn(2, source, subqueryAliasName, "COLUMN_COMMENT", subqueryAliasName, "Comment"),
							createProjectedColumn(2, source, subqueryAliasName, "TABLE_NAME", subqueryAliasName, "TABLE_NAME"),
							createProjectedColumn(2, source, subqueryAliasName, "TABLE_SCHEMA", subqueryAliasName, "TABLE_SCHEMA"),
							createProjectedColumn(2, source, subqueryAliasName, "ORDINAL_POSITION", subqueryAliasName, "ORDINAL_POSITION"),
						),
						2,
						subqueryAliasName,
					)
					for _, from := range []string{"from foo", "from test.foo", "from foo from test", "in foo in test", "from foo in test", "in foo from test"} {
						test(fmt.Sprintf("show full columns %s", from), func() evaluator.PlanStage {
							return evaluator.NewProjectStage(
								evaluator.NewOrderByStage(
									evaluator.NewFilterStage(
										subquery,
										evaluator.NewSQLAndExpr(
											evaluator.NewSQLEqualsExpr(
												createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_NAME"),
												evaluator.SQLVarchar("foo"),
											),
											evaluator.NewSQLEqualsExpr(
												createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_SCHEMA"),
												evaluator.SQLVarchar(defaultDbName),
											),
										),
									),
									evaluator.NewOrderByTerm(
										createSQLColumnExprFromSource(subquery, subqueryAliasName, "ORDINAL_POSITION"),
										true,
									),
								),
								createProjectedColumn(1, subquery, subqueryAliasName, "Field", subqueryAliasName, "Field"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Type", subqueryAliasName, "Type"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Collation", subqueryAliasName, "Collation"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Null", subqueryAliasName, "Null"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Key", subqueryAliasName, "Key"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Default", subqueryAliasName, "Default"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Extra", subqueryAliasName, "Extra"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Privileges", subqueryAliasName, "Privileges"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Comment", subqueryAliasName, "Comment"),
							)
						})
						test(fmt.Sprintf("show full columns %s like 'n'", from), func() evaluator.PlanStage {
							return evaluator.NewProjectStage(
								evaluator.NewOrderByStage(
									evaluator.NewFilterStage(
										subquery,
										evaluator.NewSQLAndExpr(
											evaluator.NewSQLAndExpr(
												evaluator.NewSQLEqualsExpr(
													createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_NAME"),
													evaluator.SQLVarchar("foo"),
												),
												evaluator.NewSQLEqualsExpr(
													createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_SCHEMA"),
													evaluator.SQLVarchar(defaultDbName),
												),
											),
											evaluator.NewSQLLikeExpr(
												createSQLColumnExprFromSource(subquery, subqueryAliasName, "Field"), evaluator.SQLVarchar("n"),
												evaluator.SQLVarchar("\\"),
											),
										),
									),
									evaluator.NewOrderByTerm(
										createSQLColumnExprFromSource(subquery, subqueryAliasName, "ORDINAL_POSITION"),
										true,
									),
								),
								createProjectedColumn(1, subquery, subqueryAliasName, "Field", subqueryAliasName, "Field"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Type", subqueryAliasName, "Type"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Collation", subqueryAliasName, "Collation"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Null", subqueryAliasName, "Null"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Key", subqueryAliasName, "Key"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Default", subqueryAliasName, "Default"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Extra", subqueryAliasName, "Extra"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Privileges", subqueryAliasName, "Privileges"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Comment", subqueryAliasName, "Comment"),
							)
						})
						test(fmt.Sprintf("show full columns %s where `Field` = 'n'", from), func() evaluator.PlanStage {
							return evaluator.NewProjectStage(
								evaluator.NewOrderByStage(
									evaluator.NewFilterStage(
										subquery,
										evaluator.NewSQLAndExpr(
											evaluator.NewSQLAndExpr(
												evaluator.NewSQLEqualsExpr(
													createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_NAME"),
													evaluator.SQLVarchar("foo"),
												),
												evaluator.NewSQLEqualsExpr(
													createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_SCHEMA"),
													evaluator.SQLVarchar(defaultDbName),
												),
											),
											evaluator.NewSQLEqualsExpr(createSQLColumnExprFromSource(subquery, subqueryAliasName, "Field"),
												evaluator.SQLVarchar("n"),
											),
										),
									),
									evaluator.NewOrderByTerm(
										createSQLColumnExprFromSource(subquery, subqueryAliasName, "ORDINAL_POSITION"),
										true,
									),
								),
								createProjectedColumn(1, subquery, subqueryAliasName, "Field", subqueryAliasName, "Field"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Type", subqueryAliasName, "Type"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Collation", subqueryAliasName, "Collation"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Null", subqueryAliasName, "Null"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Key", subqueryAliasName, "Key"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Default", subqueryAliasName, "Default"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Extra", subqueryAliasName, "Extra"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Privileges", subqueryAliasName, "Privileges"),
								createProjectedColumn(1, subquery, subqueryAliasName, "Comment", subqueryAliasName, "Comment"),
							)
						})
					}
				})
			})

			Convey("create database", func() {
				dbName := "test"

				test("show create database "+dbName, func() evaluator.PlanStage {
					return evaluator.NewProjectStage(
						evaluator.NewDualStage(),
						evaluator.ProjectedColumn{
							Column: &evaluator.Column{
								SelectID: 1,
								Name:     "Database",
								SQLType:  schema.SQLVarchar,
							},
							Expr: evaluator.SQLVarchar(dbName),
						},
						evaluator.ProjectedColumn{
							Column: &evaluator.Column{
								SelectID: 1,
								Name:     "Create Database",
								SQLType:  schema.SQLVarchar,
							},
							Expr: evaluator.SQLVarchar(catalog.GenerateCreateDatabase(dbName, "")),
						},
					)
				})

				test("show create database if not exists "+dbName, func() evaluator.PlanStage {
					return evaluator.NewProjectStage(
						evaluator.NewDualStage(),
						evaluator.ProjectedColumn{
							Column: &evaluator.Column{
								SelectID: 1,
								Name:     "Database",
								SQLType:  schema.SQLVarchar,
							},
							Expr: evaluator.SQLVarchar(dbName),
						},
						evaluator.ProjectedColumn{
							Column: &evaluator.Column{
								SelectID: 1,
								Name:     "Create Database",
								SQLType:  schema.SQLVarchar,
							},
							Expr: evaluator.SQLVarchar(catalog.GenerateCreateDatabase(dbName, "IF NOT EXISTS")),
						},
					)
				})
			})

			Convey("create table", func() {
				testDB, _ := testCatalog.Database("test")
				tbl, _ := testDB.Table("foo")

				createTableSQL := catalog.GenerateCreateTable(tbl, 10)
				test("show create table foo", func() evaluator.PlanStage {
					return evaluator.NewProjectStage(
						evaluator.NewDualStage(),
						evaluator.ProjectedColumn{
							Column: &evaluator.Column{
								SelectID: 1,
								Name:     "Table",
								SQLType:  schema.SQLVarchar,
							},
							Expr: evaluator.SQLVarchar(string(tbl.Name())),
						},
						evaluator.ProjectedColumn{
							Column: &evaluator.Column{
								SelectID: 1,
								Name:     "Create Table",
								SQLType:  schema.SQLVarchar,
							},
							Expr: evaluator.SQLVarchar(createTableSQL),
						},
					)
				})
				test("show create table .foo", func() evaluator.PlanStage {
					return evaluator.NewProjectStage(
						evaluator.NewDualStage(),
						evaluator.ProjectedColumn{
							Column: &evaluator.Column{
								SelectID: 1,
								Name:     "Table",
								SQLType:  schema.SQLVarchar,
							},
							Expr: evaluator.SQLVarchar(string(tbl.Name())),
						},
						evaluator.ProjectedColumn{
							Column: &evaluator.Column{
								SelectID: 1,
								Name:     "Create Table",
								SQLType:  schema.SQLVarchar,
							},
							Expr: evaluator.SQLVarchar(createTableSQL),
						},
					)
				})
				test("show create table test.foo", func() evaluator.PlanStage {
					return evaluator.NewProjectStage(
						evaluator.NewDualStage(),
						evaluator.ProjectedColumn{
							Column: &evaluator.Column{
								SelectID: 1,
								Name:     "Table",
								SQLType:  schema.SQLVarchar,
							},
							Expr: evaluator.SQLVarchar(string(tbl.Name())),
						},
						evaluator.ProjectedColumn{
							Column: &evaluator.Column{
								SelectID: 1,
								Name:     "Create Table",
								SQLType:  schema.SQLVarchar,
							},
							Expr: evaluator.SQLVarchar(createTableSQL),
						},
					)
				})
			})
			Convey("databases/schemas", func() {
				subqueryAliasName = "SCHEMATA"
				tbl, err := informationSchemaDB.Table(subqueryAliasName)
				So(err, ShouldBeNil)
				source := evaluator.NewDynamicSourceStage(informationSchemaDB, tbl.(*catalog.DynamicTable), 2, subqueryAliasName)
				subquery := evaluator.NewSubquerySourceStage(
					evaluator.NewProjectStage(
						source,
						createProjectedColumn(2, source, subqueryAliasName, "SCHEMA_NAME", subqueryAliasName, "Database"),
					),
					2,
					subqueryAliasName,
				)
				for _, name := range []string{"databases", "schemas"} {
					test(fmt.Sprintf("show %s", name), func() evaluator.PlanStage {
						return evaluator.NewOrderByStage(
							subquery,
							evaluator.NewOrderByTerm(
								createSQLColumnExprFromSource(subquery, subqueryAliasName, "Database"),
								true,
							),
						)
					})
					test(fmt.Sprintf("show %s like 'n'", name), func() evaluator.PlanStage {
						return evaluator.NewOrderByStage(
							evaluator.NewFilterStage(
								subquery,
								evaluator.NewSQLLikeExpr(
									createSQLColumnExprFromSource(subquery, subqueryAliasName, "Database"),
									evaluator.SQLVarchar("n"),
									evaluator.SQLVarchar("\\"),
								),
							),
							evaluator.NewOrderByTerm(
								createSQLColumnExprFromSource(subquery, subqueryAliasName, "Database"),
								true,
							),
						)
					})
					test(fmt.Sprintf("show %s where `Database` = 'n'", name), func() evaluator.PlanStage {
						return evaluator.NewOrderByStage(
							evaluator.NewFilterStage(
								subquery,
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(subquery, subqueryAliasName, "Database"),
									evaluator.SQLVarchar("n"),
								),
							),
							evaluator.NewOrderByTerm(
								createSQLColumnExprFromSource(subquery, subqueryAliasName, "Database"),
								true,
							),
						)
					})
				}
			})
			Convey("status/variables", func() {
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

						tbl, err := informationSchemaDB.Table(fmt.Sprintf("%s_%s", actualScope, kind))
						So(err, ShouldBeNil)
						actualTableName := string(tbl.Name())
						source := evaluator.NewDynamicSourceStage(informationSchemaDB, tbl.(*catalog.DynamicTable), 2, actualTableName)
						subquery := evaluator.NewSubquerySourceStage(
							evaluator.NewProjectStage(
								source,
								createProjectedColumn(2, source, actualTableName, "VARIABLE_NAME", actualTableName, "Variable_name"),
								createProjectedColumn(2, source, actualTableName, "VARIABLE_VALUE", actualTableName, "Value"),
							),
							2,
							subqueryAliasName,
						)
						showName := strings.TrimSpace(scope + " " + kind)
						test(fmt.Sprintf("show %s", showName), func() evaluator.PlanStage {
							return evaluator.NewOrderByStage(
								subquery,
								evaluator.NewOrderByTerm(
									createSQLColumnExprFromSource(subquery, subqueryAliasName, "Variable_name"),
									true,
								),
							)
						})
						test(fmt.Sprintf("show %s like 'n'", showName), func() evaluator.PlanStage {
							return evaluator.NewOrderByStage(
								evaluator.NewFilterStage(
									subquery,
									evaluator.NewSQLLikeExpr(
										createSQLColumnExprFromSource(subquery, subqueryAliasName, "Variable_name"),
										evaluator.SQLVarchar("n"),
										evaluator.SQLVarchar("\\"),
									),
								),
								evaluator.NewOrderByTerm(
									createSQLColumnExprFromSource(subquery, subqueryAliasName, "Variable_name"),
									true,
								),
							)
						})
						test(fmt.Sprintf("show %s where Variable_name = 'n'", showName), func() evaluator.PlanStage {
							return evaluator.NewOrderByStage(
								evaluator.NewFilterStage(
									subquery,
									evaluator.NewSQLEqualsExpr(
										createSQLColumnExprFromSource(subquery, subqueryAliasName, "Variable_name"),
										evaluator.SQLVarchar("n"),
									),
								),
								evaluator.NewOrderByTerm(
									createSQLColumnExprFromSource(subquery, subqueryAliasName, "Variable_name"),
									true,
								),
							)
						})
					}
				}
			})
			Convey("tables", func() {
				subqueryAliasName = "TABLES"
				tbl, err := informationSchemaDB.Table(subqueryAliasName)
				So(err, ShouldBeNil)
				source := evaluator.NewDynamicSourceStage(informationSchemaDB, tbl.(*catalog.DynamicTable), 2, subqueryAliasName)
				columnName := "Tables_in_" + defaultDbName
				Convey("plain", func() {
					subquery := evaluator.NewSubquerySourceStage(
						evaluator.NewProjectStage(
							source,
							createProjectedColumn(2, source, subqueryAliasName, "TABLE_NAME", subqueryAliasName, columnName),
							createProjectedColumn(2, source, subqueryAliasName, "TABLE_SCHEMA", subqueryAliasName, "TABLE_SCHEMA"),
						),
						2,
						subqueryAliasName,
					)

					for _, from := range []string{"", " from test", " in test"} {
						test(fmt.Sprintf("show tables%s", from), func() evaluator.PlanStage {
							return evaluator.NewProjectStage(
								evaluator.NewOrderByStage(
									evaluator.NewFilterStage(
										subquery,
										evaluator.NewSQLEqualsExpr(
											createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_SCHEMA"),
											evaluator.SQLVarchar("test"),
										),
									),
									evaluator.NewOrderByTerm(
										createSQLColumnExprFromSource(subquery, subqueryAliasName, columnName),
										true,
									),
								),
								createProjectedColumn(1, subquery, subqueryAliasName, columnName, subqueryAliasName, columnName),
							)
						})
						test(fmt.Sprintf("show tables%s like 'n'", from), func() evaluator.PlanStage {
							return evaluator.NewProjectStage(
								evaluator.NewOrderByStage(
									evaluator.NewFilterStage(
										subquery,
										evaluator.NewSQLAndExpr(
											evaluator.NewSQLEqualsExpr(
												createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_SCHEMA"),
												evaluator.SQLVarchar("test"),
											),
											evaluator.NewSQLLikeExpr(
												createSQLColumnExprFromSource(subquery, subqueryAliasName, columnName),
												evaluator.SQLVarchar("n"),
												evaluator.SQLVarchar("\\"),
											),
										),
									),
									evaluator.NewOrderByTerm(
										createSQLColumnExprFromSource(subquery, subqueryAliasName, columnName),
										true,
									),
								),
								createProjectedColumn(1, subquery, subqueryAliasName, columnName, subqueryAliasName, columnName),
							)
						})
						test(fmt.Sprintf("show tables%s where `%s` = 'n'", from, columnName), func() evaluator.PlanStage {
							return evaluator.NewProjectStage(
								evaluator.NewOrderByStage(
									evaluator.NewFilterStage(
										subquery,
										evaluator.NewSQLAndExpr(
											evaluator.NewSQLEqualsExpr(
												createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_SCHEMA"),
												evaluator.SQLVarchar("test"),
											),
											evaluator.NewSQLEqualsExpr(
												createSQLColumnExprFromSource(subquery, subqueryAliasName, columnName),
												evaluator.SQLVarchar("n"),
											),
										),
									),
									evaluator.NewOrderByTerm(
										createSQLColumnExprFromSource(subquery, subqueryAliasName, columnName),
										true,
									),
								),
								createProjectedColumn(1, subquery, subqueryAliasName, columnName, subqueryAliasName, columnName),
							)
						})
					}
				})
				Convey("full", func() {
					subquery := evaluator.NewSubquerySourceStage(
						evaluator.NewProjectStage(
							source,
							createProjectedColumn(2, source, subqueryAliasName, "TABLE_NAME", subqueryAliasName, columnName),
							createProjectedColumn(2, source, subqueryAliasName, "TABLE_TYPE", subqueryAliasName, "Table_type"),
							createProjectedColumn(2, source, subqueryAliasName, "TABLE_SCHEMA", subqueryAliasName, "TABLE_SCHEMA"),
						),
						2,
						subqueryAliasName,
					)

					for _, from := range []string{"", " from test", " in test"} {
						test(fmt.Sprintf("show full tables%s", from), func() evaluator.PlanStage {
							return evaluator.NewProjectStage(
								evaluator.NewOrderByStage(
									evaluator.NewFilterStage(
										subquery,
										evaluator.NewSQLEqualsExpr(
											createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_SCHEMA"),
											evaluator.SQLVarchar("test"),
										),
									),
									evaluator.NewOrderByTerm(
										createSQLColumnExprFromSource(subquery, subqueryAliasName, columnName),
										true,
									),
								),
								createProjectedColumn(1, subquery, subqueryAliasName, columnName, subqueryAliasName, columnName),
								createProjectedColumn(1, subquery, subqueryAliasName, "Table_type", subqueryAliasName, "Table_type"),
							)
						})
						test(fmt.Sprintf("show full tables%s like 'n'", from), func() evaluator.PlanStage {
							return evaluator.NewProjectStage(
								evaluator.NewOrderByStage(
									evaluator.NewFilterStage(
										subquery,
										evaluator.NewSQLAndExpr(
											evaluator.NewSQLEqualsExpr(
												createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_SCHEMA"),
												evaluator.SQLVarchar("test"),
											),
											evaluator.NewSQLLikeExpr(
												createSQLColumnExprFromSource(subquery, subqueryAliasName, columnName),
												evaluator.SQLVarchar("n"),
												evaluator.SQLVarchar("\\"),
											),
										),
									),
									evaluator.NewOrderByTerm(
										createSQLColumnExprFromSource(subquery, subqueryAliasName, columnName),
										true,
									),
								),
								createProjectedColumn(1, subquery, subqueryAliasName, columnName, subqueryAliasName, columnName),
								createProjectedColumn(1, subquery, subqueryAliasName, "Table_type", subqueryAliasName, "Table_type"),
							)
						})
						test(fmt.Sprintf("show full tables%s where `%s` = 'n'", from, columnName), func() evaluator.PlanStage {
							return evaluator.NewProjectStage(
								evaluator.NewOrderByStage(
									evaluator.NewFilterStage(
										subquery,
										evaluator.NewSQLAndExpr(
											evaluator.NewSQLEqualsExpr(
												createSQLColumnExprFromSource(subquery, subqueryAliasName, "TABLE_SCHEMA"),
												evaluator.SQLVarchar("test"),
											),
											evaluator.NewSQLEqualsExpr(
												createSQLColumnExprFromSource(subquery, subqueryAliasName, columnName),
												evaluator.SQLVarchar("n"),
											),
										),
									),
									evaluator.NewOrderByTerm(
										createSQLColumnExprFromSource(subquery, subqueryAliasName, columnName),
										true,
									),
								),
								createProjectedColumn(1, subquery, subqueryAliasName, columnName, subqueryAliasName, columnName),
								createProjectedColumn(1, subquery, subqueryAliasName, "Table_type", subqueryAliasName, "Table_type"),
							)
						})
					}
				})
			})
		})

		Convey("Select Statements", func() {
			Convey("dual queries", func() {
				test("select 2 + 3", func() evaluator.PlanStage {
					return evaluator.NewProjectStage(
						evaluator.NewDualStage(),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "2+3", evaluator.NewSQLAddExpr(evaluator.SQLInt(2), evaluator.SQLInt(3))),
					)
				})

				test("select false", func() evaluator.PlanStage {
					return evaluator.NewProjectStage(
						evaluator.NewDualStage(),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "false", evaluator.SQLFalse),
					)
				})

				test("select true", func() evaluator.PlanStage {
					return evaluator.NewProjectStage(
						evaluator.NewDualStage(),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "true", evaluator.SQLTrue),
					)
				})

				test("select 2 + 3 from dual", func() evaluator.PlanStage {
					return evaluator.NewProjectStage(
						evaluator.NewDualStage(),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "2+3", evaluator.NewSQLAddExpr(evaluator.SQLInt(2), evaluator.SQLInt(3))),
					)
				})
			})

			Convey("from", func() {

				Convey("subqueries", func() {
					test("select a from (select a from foo) f", func() evaluator.PlanStage {
						source := createMongoSource(2, "foo", "foo")
						subquery := evaluator.NewSubquerySourceStage(
							evaluator.NewProjectStage(
								source,
								createProjectedColumn(2, source, "foo", "a", "foo", "a"),
							),
							2,
							"f",
						)
						return evaluator.NewProjectStage(subquery, createProjectedColumn(2, subquery, "f", "a", "f", "a"))
					})

					test("select f.a from (select a from foo) f", func() evaluator.PlanStage {
						source := createMongoSource(2, "foo", "foo")
						subquery := evaluator.NewSubquerySourceStage(evaluator.NewProjectStage(source, createProjectedColumn(2, source, "foo", "a", "foo", "a")), 2, "f")
						return evaluator.NewProjectStage(subquery, createProjectedColumn(2, subquery, "f", "a", "f", "a"))
					})

					test("select f.a from (select test.a from foo test) f", func() evaluator.PlanStage {
						source := createMongoSource(2, "foo", "test")
						subquery := evaluator.NewSubquerySourceStage(evaluator.NewProjectStage(source, createProjectedColumn(2, source, "test", "a", "test", "a")), 2, "f")
						return evaluator.NewProjectStage(subquery, createProjectedColumn(2, subquery, "f", "a", "f", "a"))
					})

					testVariables("select g.a from (select a from foo) g",
						func() *variable.Container {
							vars := &variable.Container{
								MongoDBInfo: testInfo,
							}
							vars.SetSystemVariable(variable.SQLSelectLimit, 5)
							return vars
						},
						func() evaluator.PlanStage {
							source := createMongoSource(2, "foo", "foo")
							subquery := evaluator.NewSubquerySourceStage(evaluator.NewProjectStage(source, createProjectedColumn(2, source, "foo", "a", "foo", "a")), 2, "g")
							return evaluator.NewLimitStage(
								evaluator.NewProjectStage(subquery, createProjectedColumn(2, subquery, "g", "a", "g", "a")),
								0,
								5,
							)
						})
				})

				Convey("joins", func() {
					test("select foo.a, bar.a from foo, bar", func() evaluator.PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := evaluator.NewJoinStage(evaluator.CrossJoin, fooSource, barSource, evaluator.SQLTrue)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "foo", "a", "foo", "a"),
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
						)
					})

					test("select f.a, bar.a from foo f, bar", func() evaluator.PlanStage {
						fooSource := createMongoSource(1, "foo", "f")
						barSource := createMongoSource(1, "bar", "bar")
						join := evaluator.NewJoinStage(evaluator.CrossJoin, fooSource, barSource, evaluator.SQLTrue)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "f", "a", "f", "a"),
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
						)
					})

					test("select f.a, b.a from foo f, bar b", func() evaluator.PlanStage {
						fooSource := createMongoSource(1, "foo", "f")
						barSource := createMongoSource(1, "bar", "b")
						join := evaluator.NewJoinStage(evaluator.CrossJoin, fooSource, barSource, evaluator.SQLTrue)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "f", "a", "f", "a"),
							createProjectedColumn(1, join, "b", "a", "b", "a"),
						)
					})

					test("select foo.a, bar.a from foo inner join bar on foo.b = bar.b", func() evaluator.PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := evaluator.NewJoinStage(evaluator.InnerJoin, fooSource, barSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(fooSource, "foo", "b"),
								createSQLColumnExprFromSource(barSource, "bar", "b"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "foo", "a", "foo", "a"),
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
						)
					})

					test("select foo.a, bar.a from foo join bar on foo.b = bar.b", func() evaluator.PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := evaluator.NewJoinStage(evaluator.InnerJoin, fooSource, barSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(fooSource, "foo", "b"),
								createSQLColumnExprFromSource(barSource, "bar", "b"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "foo", "a", "foo", "a"),
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
						)
					})

					test("select foo.a, bar.a from foo left outer join bar on foo.b = bar.b", func() evaluator.PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := evaluator.NewJoinStage(evaluator.LeftJoin, fooSource, barSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(fooSource, "foo", "b"),
								createSQLColumnExprFromSource(barSource, "bar", "b"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "foo", "a", "foo", "a"),
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
						)
					})

					test("select foo.a, bar.a from foo right outer join bar on foo.b = bar.b", func() evaluator.PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := evaluator.NewJoinStage(parser.AST_RIGHT_JOIN, fooSource, barSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(fooSource, "foo", "b"),
								createSQLColumnExprFromSource(barSource, "bar", "b"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "foo", "a", "foo", "a"),
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
						)
					})
					test("select foo.a, bar.a from foo straight_join bar on foo.b = bar.b", func() evaluator.PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := evaluator.NewJoinStage(evaluator.StraightJoin, fooSource, barSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(fooSource, "foo", "b"),
								createSQLColumnExprFromSource(barSource, "bar", "b"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "foo", "a", "foo", "a"),
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
						)
					})
					test("select foo.a, bar.a from foo join bar on foo.a = bar.a and foo.e = bar.d join baz on baz.b = bar.b", func() evaluator.PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						firstJoin := evaluator.NewJoinStage(evaluator.InnerJoin, fooSource, barSource,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(fooSource, "foo", "a"),
									createSQLColumnExprFromSource(barSource, "bar", "a"),
								),
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(fooSource, "foo", "e"),
									createSQLColumnExprFromSource(barSource, "bar", "d"),
								),
							),
						)
						secondJoin := evaluator.NewJoinStage(evaluator.InnerJoin, firstJoin, bazSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(bazSource, "baz", "b"),
								createSQLColumnExprFromSource(barSource, "bar", "b"),
							),
						)
						return evaluator.NewProjectStage(secondJoin,
							createProjectedColumn(1, secondJoin, "foo", "a", "foo", "a"),
							createProjectedColumn(1, secondJoin, "bar", "a", "bar", "a"),
						)
					})
					test("select bar.a, baz.b from bar join baz using (b)", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(barSource, "bar", "b"),
								createSQLColumnExprFromSource(bazSource, "baz", "b"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
							createProjectedColumn(1, join, "baz", "b", "baz", "b"),
						)
					})
					test("select bar.a, baz.b from bar join baz using (a, b)", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(barSource, "bar", "a"),
									createSQLColumnExprFromSource(bazSource, "baz", "a"),
								),
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(barSource, "bar", "b"),
									createSQLColumnExprFromSource(bazSource, "baz", "b"),
								),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
							createProjectedColumn(1, join, "baz", "b", "baz", "b"),
						)
					})
					test("select bar.a, buzz.d, foo.c from bar join buzz join foo using (a, c)", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						buzzSource := createMongoSource(1, "buzz", "buzz")
						fooSource := createMongoSource(1, "foo", "foo")
						firstJoin := evaluator.NewJoinStage(parser.AST_CROSS_JOIN, barSource, buzzSource, evaluator.SQLBool(1))
						secondJoin := evaluator.NewJoinStage(evaluator.InnerJoin, firstJoin, fooSource,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(barSource, "bar", "a"),
									createSQLColumnExprFromSource(fooSource, "foo", "a"),
								),
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(buzzSource, "buzz", "c"),
									createSQLColumnExprFromSource(fooSource, "foo", "c"),
								),
							),
						)
						return evaluator.NewProjectStage(secondJoin,
							createProjectedColumn(1, secondJoin, "bar", "a", "bar", "a"),
							createProjectedColumn(1, secondJoin, "buzz", "d", "buzz", "d"),
							createProjectedColumn(1, secondJoin, "foo", "c", "foo", "c"),
						)
					})
					test("select bar.a, buzz.d, foo.c from bar join foo using (a) join buzz using (c)", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						fooSource := createMongoSource(1, "foo", "foo")
						buzzSource := createMongoSource(1, "buzz", "buzz")
						firstJoin := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, fooSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(barSource, "bar", "a"),
								createSQLColumnExprFromSource(fooSource, "foo", "a"),
							),
						)
						secondJoin := evaluator.NewJoinStage(evaluator.InnerJoin, firstJoin, buzzSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(fooSource, "foo", "c"),
								createSQLColumnExprFromSource(buzzSource, "buzz", "c"),
							),
						)
						return evaluator.NewProjectStage(secondJoin,
							createProjectedColumn(1, secondJoin, "bar", "a", "bar", "a"),
							createProjectedColumn(1, secondJoin, "buzz", "d", "buzz", "d"),
							createProjectedColumn(1, secondJoin, "foo", "c", "foo", "c"),
						)
					})
					test("select bar.a, baz.b from bar join baz using (a, a, a, a, b)", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(barSource, "bar", "a"),
									createSQLColumnExprFromSource(bazSource, "baz", "a"),
								),
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(barSource, "bar", "b"),
									createSQLColumnExprFromSource(bazSource, "baz", "b"),
								),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
							createProjectedColumn(1, join, "baz", "b", "baz", "b"),
						)
					})
					test("select bar.a, baz.b from bar join baz using (a, b, b, b, b)", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(barSource, "bar", "a"),
									createSQLColumnExprFromSource(bazSource, "baz", "a"),
								),
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(barSource, "bar", "b"),
									createSQLColumnExprFromSource(bazSource, "baz", "b"),
								),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
							createProjectedColumn(1, join, "baz", "b", "baz", "b"),
						)
					})
					test("select bar.a, baz.b from bar cross join baz using (b)", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(barSource, "bar", "b"),
								createSQLColumnExprFromSource(bazSource, "baz", "b"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
							createProjectedColumn(1, join, "baz", "b", "baz", "b"),
						)
					})
					test("select bar.a, baz.b from bar inner join baz", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := evaluator.NewJoinStage(evaluator.CrossJoin, barSource, bazSource, evaluator.SQLBool(1))
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
							createProjectedColumn(1, join, "baz", "b", "baz", "b"),
						)
					})
					test("select bar.a, biz.b from bar join (select baz.b, foo.c from baz join foo on baz.a = foo.a) as biz using (b)", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(2, "baz", "baz")
						fooSource := createMongoSource(2, "foo", "foo")
						subJoin := evaluator.NewJoinStage(evaluator.InnerJoin, bazSource, fooSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(bazSource, "baz", "a"),
								createSQLColumnExprFromSource(fooSource, "foo", "a"),
							),
						)
						bizSource := evaluator.NewSubquerySourceStage(
							evaluator.NewProjectStage(subJoin,
								createProjectedColumn(2, subJoin, "baz", "b", "baz", "b"),
								createProjectedColumn(2, subJoin, "foo", "c", "foo", "c"),
							), 2, "biz")
						join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bizSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(barSource, "bar", "b"),
								createSQLColumnExprFromSource(bizSource, "biz", "b"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
							createProjectedColumn(2, join, "biz", "b", "biz", "b"),
						)
					})
					test("select bar.a, biz.b from (select baz.b, foo.c from baz join foo on baz.a = foo.a) as biz join bar using (b)", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(2, "baz", "baz")
						fooSource := createMongoSource(2, "foo", "foo")
						subJoin := evaluator.NewJoinStage(evaluator.InnerJoin, bazSource, fooSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(bazSource, "baz", "a"),
								createSQLColumnExprFromSource(fooSource, "foo", "a"),
							),
						)
						bizSource := evaluator.NewSubquerySourceStage(
							evaluator.NewProjectStage(subJoin,
								createProjectedColumn(2, subJoin, "baz", "b", "baz", "b"),
								createProjectedColumn(2, subJoin, "foo", "c", "foo", "c"),
							), 2, "biz")
						join := evaluator.NewJoinStage(evaluator.InnerJoin, bizSource, barSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(bizSource, "biz", "b"),
								createSQLColumnExprFromSource(barSource, "bar", "b"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
							createProjectedColumn(2, join, "biz", "b", "biz", "b"),
						)
					})
					test("select fiz.b from (select bar.b from bar) as biz join (select foo.b from foo) as fiz using (b)", func() evaluator.PlanStage {
						barSource := createMongoSource(2, "bar", "bar")
						fooSource := createMongoSource(3, "foo", "foo")
						bizSource := evaluator.NewSubquerySourceStage(
							evaluator.NewProjectStage(
								barSource,
								createProjectedColumn(2, barSource, "bar", "b", "bar", "b"),
							),
							2,
							"biz",
						)
						fizSource := evaluator.NewSubquerySourceStage(
							evaluator.NewProjectStage(
								fooSource,
								createProjectedColumn(3, fooSource, "foo", "b", "foo", "b"),
							),
							3,
							"fiz",
						)
						join := evaluator.NewJoinStage(evaluator.InnerJoin, bizSource, fizSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(bizSource, "biz", "b"),
								createSQLColumnExprFromSource(fizSource, "fiz", "b"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(3, join, "fiz", "b", "fiz", "b"))
					})
					test("select * from bar join baz using (b)", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(barSource, "bar", "b"),
								createSQLColumnExprFromSource(bazSource, "baz", "b"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "b", "bar", "b"),
							createProjectedColumn(1, join, "bar", "_id", "bar", "_id"),
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
							createProjectedColumn(1, join, "bar", "d", "bar", "d"),
							createProjectedColumn(1, join, "baz", "_id", "baz", "_id"),
							createProjectedColumn(1, join, "baz", "a", "baz", "a"),
						)
					})
					test("select * from bar join baz using (_id, b)", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(barSource, "bar", "_id"),
									createSQLColumnExprFromSource(bazSource, "baz", "_id"),
								),
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(barSource, "bar", "b"),
									createSQLColumnExprFromSource(bazSource, "baz", "b"),
								),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "_id", "bar", "_id"),
							createProjectedColumn(1, join, "bar", "b", "bar", "b"),
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
							createProjectedColumn(1, join, "bar", "d", "bar", "d"),
							createProjectedColumn(1, join, "baz", "a", "baz", "a"))
					})
					test("select * from bar right join baz using (b)", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := evaluator.NewJoinStage(evaluator.RightJoin, barSource, bazSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(barSource, "bar", "b"),
								createSQLColumnExprFromSource(bazSource, "baz", "b"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "baz", "b", "baz", "b"),
							createProjectedColumn(1, join, "baz", "_id", "baz", "_id"),
							createProjectedColumn(1, join, "baz", "a", "baz", "a"),
							createProjectedColumn(1, join, "bar", "_id", "bar", "_id"),
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
							createProjectedColumn(1, join, "bar", "d", "bar", "d"),
						)
					})
					test("select bar.*, baz.* from bar join baz using (b)", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(barSource, "bar", "b"),
								createSQLColumnExprFromSource(bazSource, "baz", "b"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "_id", "bar", "_id"),
							createProjectedColumn(1, join, "bar", "a", "bar", "a"),
							createProjectedColumn(1, join, "bar", "b", "bar", "b"),
							createProjectedColumn(1, join, "bar", "d", "bar", "d"),
							createProjectedColumn(1, join, "baz", "_id", "baz", "_id"),
							createProjectedColumn(1, join, "baz", "a", "baz", "a"),
							createProjectedColumn(1, join, "baz", "b", "baz", "b"),
						)
					})
					test("select bar.b, baz.b from bar join baz using (b)", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(barSource, "bar", "b"),
								createSQLColumnExprFromSource(bazSource, "baz", "b"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "b", "bar", "b"),
							createProjectedColumn(1, join, "baz", "b", "baz", "b"))
					})
					test("select * from buzz join (baz join bar using (_id)) using (d)", func() evaluator.PlanStage {
						buzzSource := createMongoSource(1, "buzz", "buzz")
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join1 := evaluator.NewJoinStage(evaluator.InnerJoin, bazSource, barSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(bazSource, "baz", "_id"),
								createSQLColumnExprFromSource(barSource, "bar", "_id"),
							),
						)
						join2 := evaluator.NewJoinStage(evaluator.InnerJoin, buzzSource, join1,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(buzzSource, "buzz", "d"),
								createSQLColumnExprFromSource(barSource, "bar", "d"),
							),
						)
						return evaluator.NewProjectStage(join2,
							createProjectedColumn(1, buzzSource, "buzz", "d", "buzz", "d"),
							createProjectedColumn(1, buzzSource, "buzz", "_id", "buzz", "_id"),
							createProjectedColumn(1, buzzSource, "buzz", "c", "buzz", "c"),
							createProjectedColumn(1, bazSource, "baz", "_id", "baz", "_id"),
							createProjectedColumn(1, bazSource, "baz", "a", "baz", "a"),
							createProjectedColumn(1, bazSource, "baz", "b", "baz", "b"),
							createProjectedColumn(1, barSource, "bar", "a", "bar", "a"),
							createProjectedColumn(1, barSource, "bar", "b", "bar", "b"),
						)
					})
					test("select bar.a from bar natural join baz", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLAndExpr(
									evaluator.NewSQLEqualsExpr(
										createSQLColumnExprFromSource(barSource, "bar", "_id"),
										createSQLColumnExprFromSource(bazSource, "baz", "_id"),
									),
									evaluator.NewSQLEqualsExpr(
										createSQLColumnExprFromSource(barSource, "bar", "a"),
										createSQLColumnExprFromSource(bazSource, "baz", "a"),
									),
								),
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(barSource, "bar", "b"),
									createSQLColumnExprFromSource(bazSource, "baz", "b"),
								),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, barSource, "bar", "a", "bar", "a"))
					})
					test("select buzz.c from buzz join bar natural join baz", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						buzzSource := createMongoSource(1, "buzz", "buzz")
						naturalJoin := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, bazSource,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLAndExpr(
									evaluator.NewSQLEqualsExpr(
										createSQLColumnExprFromSource(barSource, "bar", "_id"),
										createSQLColumnExprFromSource(bazSource, "baz", "_id"),
									),
									evaluator.NewSQLEqualsExpr(
										createSQLColumnExprFromSource(barSource, "bar", "a"),
										createSQLColumnExprFromSource(bazSource, "baz", "a"),
									),
								),
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(barSource, "bar", "b"),
									createSQLColumnExprFromSource(bazSource, "baz", "b"),
								),
							),
						)
						join := evaluator.NewJoinStage(evaluator.CrossJoin, buzzSource, naturalJoin, evaluator.SQLTrue)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, buzzSource, "buzz", "c", "buzz", "c"))
					})
					test("select buzz.c from bar join buzz natural join baz", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						buzzSource := createMongoSource(1, "buzz", "buzz")
						naturalJoin := evaluator.NewJoinStage(evaluator.InnerJoin, buzzSource, bazSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(buzzSource, "buzz", "_id"),
								createSQLColumnExprFromSource(bazSource, "baz", "_id"),
							),
						)
						join := evaluator.NewJoinStage(evaluator.CrossJoin, barSource, naturalJoin, evaluator.SQLTrue)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, buzzSource, "buzz", "c", "buzz", "c"))
					})
					test("select bar.a from bar natural join buzz natural join baz", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						buzzSource := createMongoSource(1, "buzz", "buzz")

						njoin1 := evaluator.NewJoinStage(evaluator.InnerJoin, buzzSource, bazSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(buzzSource, "buzz", "_id"),
								createSQLColumnExprFromSource(bazSource, "baz", "_id"),
							),
						)
						njoin2 := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, njoin1,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLAndExpr(
									evaluator.NewSQLAndExpr(
										evaluator.NewSQLEqualsExpr(
											createSQLColumnExprFromSource(barSource, "bar", "_id"),
											createSQLColumnExprFromSource(buzzSource, "buzz", "_id"),
										),
										evaluator.NewSQLEqualsExpr(
											createSQLColumnExprFromSource(barSource, "bar", "a"),
											createSQLColumnExprFromSource(bazSource, "baz", "a"),
										),
									),
									evaluator.NewSQLEqualsExpr(
										createSQLColumnExprFromSource(barSource, "bar", "b"),
										createSQLColumnExprFromSource(bazSource, "baz", "b"),
									),
								),
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(barSource, "bar", "d"),
									createSQLColumnExprFromSource(buzzSource, "buzz", "d"),
								),
							),
						)
						return evaluator.NewProjectStage(njoin2,
							createProjectedColumn(1, barSource, "bar", "a", "bar", "a"))
					})
					test("select baz.a from (select c from buzz) as buzzc natural join baz", func() evaluator.PlanStage {
						bazSource := createMongoSource(1, "baz", "baz")
						buzzSource := createMongoSource(2, "buzz", "buzz")
						buzzcSource := evaluator.NewSubquerySourceStage(
							evaluator.NewProjectStage(buzzSource, createProjectedColumn(2, buzzSource, "buzz", "c", "buzz", "c")), 2, "buzzc")
						join := evaluator.NewJoinStage(evaluator.CrossJoin, buzzcSource, bazSource, evaluator.SQLTrue)
						return evaluator.NewProjectStage(join, createProjectedColumn(1, bazSource, "baz", "a", "baz", "a"))
					})
					test("select buzz.c from bar join buzz using (_id, d) natural join baz", func() evaluator.PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						buzzSource := createMongoSource(1, "buzz", "buzz")
						usingJoin := evaluator.NewJoinStage(evaluator.InnerJoin, barSource, buzzSource,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(barSource, "bar", "_id"),
									createSQLColumnExprFromSource(buzzSource, "buzz", "_id"),
								),
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(barSource, "bar", "d"),
									createSQLColumnExprFromSource(buzzSource, "buzz", "d"),
								),
							),
						)
						naturalJoin := evaluator.NewJoinStage(evaluator.InnerJoin, usingJoin, bazSource,
							evaluator.NewSQLAndExpr(
								evaluator.NewSQLAndExpr(
									evaluator.NewSQLEqualsExpr(
										createSQLColumnExprFromSource(barSource, "bar", "_id"),
										createSQLColumnExprFromSource(bazSource, "baz", "_id"),
									),
									evaluator.NewSQLEqualsExpr(
										createSQLColumnExprFromSource(barSource, "bar", "a"),
										createSQLColumnExprFromSource(bazSource, "baz", "a"),
									),
								),
								evaluator.NewSQLEqualsExpr(
									createSQLColumnExprFromSource(barSource, "bar", "b"),
									createSQLColumnExprFromSource(bazSource, "baz", "b"),
								),
							),
						)
						return evaluator.NewProjectStage(naturalJoin,
							createProjectedColumn(1, buzzSource, "buzz", "c", "buzz", "c"))
					})
					test("select baz.b from baz natural left join buzz", func() evaluator.PlanStage {
						bazSource := createMongoSource(1, "baz", "baz")
						buzzSource := createMongoSource(1, "buzz", "buzz")
						join := evaluator.NewJoinStage(evaluator.LeftJoin, bazSource, buzzSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(bazSource, "baz", "_id"),
								createSQLColumnExprFromSource(buzzSource, "buzz", "_id"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, bazSource, "baz", "b", "baz", "b"))
					})
					test("select baz.b from baz natural right join buzz", func() evaluator.PlanStage {
						bazSource := createMongoSource(1, "baz", "baz")
						buzzSource := createMongoSource(1, "buzz", "buzz")
						join := evaluator.NewJoinStage(evaluator.RightJoin, bazSource, buzzSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(bazSource, "baz", "_id"),
								createSQLColumnExprFromSource(buzzSource, "buzz", "_id"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, bazSource, "baz", "b", "baz", "b"))
					})
					test("select baz.b from baz natural left outer join buzz", func() evaluator.PlanStage {
						bazSource := createMongoSource(1, "baz", "baz")
						buzzSource := createMongoSource(1, "buzz", "buzz")
						join := evaluator.NewJoinStage(evaluator.LeftJoin, bazSource, buzzSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(bazSource, "baz", "_id"),
								createSQLColumnExprFromSource(buzzSource, "buzz", "_id"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, bazSource, "baz", "b", "baz", "b"))
					})
					test("select baz.b from baz natural right outer join buzz", func() evaluator.PlanStage {
						bazSource := createMongoSource(1, "baz", "baz")
						buzzSource := createMongoSource(1, "buzz", "buzz")
						join := evaluator.NewJoinStage(evaluator.RightJoin, bazSource, buzzSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(bazSource, "baz", "_id"),
								createSQLColumnExprFromSource(buzzSource, "buzz", "_id"),
							),
						)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, bazSource, "baz", "b", "baz", "b"))
					})
					test("select baz.b from foo join baz natural right join buzz", func() evaluator.PlanStage {
						bazSource := createMongoSource(1, "baz", "baz")
						buzzSource := createMongoSource(1, "buzz", "buzz")
						fooSource := createMongoSource(1, "foo", "foo")
						njoin := evaluator.NewJoinStage(evaluator.RightJoin, bazSource, buzzSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(bazSource, "baz", "_id"),
								createSQLColumnExprFromSource(buzzSource, "buzz", "_id"),
							),
						)
						join := evaluator.NewJoinStage(evaluator.CrossJoin, fooSource, njoin, evaluator.SQLTrue)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, bazSource, "baz", "b", "baz", "b"))
					})
					test("select baz.b from baz natural right join buzz join foo", func() evaluator.PlanStage {
						bazSource := createMongoSource(1, "baz", "baz")
						buzzSource := createMongoSource(1, "buzz", "buzz")
						fooSource := createMongoSource(1, "foo", "foo")
						njoin := evaluator.NewJoinStage(evaluator.RightJoin, bazSource, buzzSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(bazSource, "baz", "_id"),
								createSQLColumnExprFromSource(buzzSource, "buzz", "_id"),
							),
						)
						join := evaluator.NewJoinStage(evaluator.CrossJoin, njoin, fooSource, evaluator.SQLTrue)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, bazSource, "baz", "b", "baz", "b"))
					})
					test("select baz.b from foo join baz natural left join buzz", func() evaluator.PlanStage {
						bazSource := createMongoSource(1, "baz", "baz")
						buzzSource := createMongoSource(1, "buzz", "buzz")
						fooSource := createMongoSource(1, "foo", "foo")
						njoin := evaluator.NewJoinStage(evaluator.LeftJoin, bazSource, buzzSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(bazSource, "baz", "_id"),
								createSQLColumnExprFromSource(buzzSource, "buzz", "_id"),
							),
						)
						join := evaluator.NewJoinStage(evaluator.CrossJoin, fooSource, njoin, evaluator.SQLTrue)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, bazSource, "baz", "b", "baz", "b"))
					})
					test("select baz.b from baz natural left join buzz join foo", func() evaluator.PlanStage {
						bazSource := createMongoSource(1, "baz", "baz")
						buzzSource := createMongoSource(1, "buzz", "buzz")
						fooSource := createMongoSource(1, "foo", "foo")
						njoin := evaluator.NewJoinStage(evaluator.LeftJoin, bazSource, buzzSource,
							evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(bazSource, "baz", "_id"),
								createSQLColumnExprFromSource(buzzSource, "buzz", "_id"),
							),
						)
						join := evaluator.NewJoinStage(evaluator.CrossJoin, njoin, fooSource, evaluator.SQLTrue)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, bazSource, "baz", "b", "baz", "b"))
					})
					test("select baz.b from (select c from buzz) as buzzc natural left join baz", func() evaluator.PlanStage {
						bazSource := createMongoSource(1, "baz", "baz")
						buzzSource := createMongoSource(2, "buzz", "buzz")
						buzzcSource := evaluator.NewSubquerySourceStage(
							evaluator.NewProjectStage(buzzSource, createProjectedColumn(2, buzzSource, "buzz", "c", "buzz", "c")), 2, "buzzc")
						join := evaluator.NewJoinStage(evaluator.CrossJoin, buzzcSource, bazSource, evaluator.SQLTrue)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, bazSource, "baz", "b", "baz", "b"))
					})
					test("select baz.b from (select c from buzz) as buzzc natural right join baz", func() evaluator.PlanStage {
						bazSource := createMongoSource(1, "baz", "baz")
						buzzSource := createMongoSource(2, "buzz", "buzz")
						buzzcSource := evaluator.NewSubquerySourceStage(
							evaluator.NewProjectStage(buzzSource, createProjectedColumn(2, buzzSource, "buzz", "c", "buzz", "c")), 2, "buzzc")
						join := evaluator.NewJoinStage(evaluator.CrossJoin, buzzcSource, bazSource, evaluator.SQLTrue)
						return evaluator.NewProjectStage(join,
							createProjectedColumn(1, bazSource, "baz", "b", "baz", "b"))
					})
				})

			})
			Convey("select", func() {
				Convey("star simple queries", func() {
					test("select * from foo", func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return evaluator.NewProjectStage(source, createAllProjectedColumnsFromSource(1, source, "foo")...)
					})

					test("select foo.* from foo", func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return evaluator.NewProjectStage(source, createAllProjectedColumnsFromSource(1, source, "foo")...)
					})

					test("select f.* from foo f", func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "f")
						return evaluator.NewProjectStage(source, createAllProjectedColumnsFromSource(1, source, "f")...)
					})

					test("select a, foo.* from foo", func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "foo")
						columns := append(
							evaluator.ProjectedColumns{createProjectedColumn(1, source, "foo", "a", "foo", "a")},
							createAllProjectedColumnsFromSource(1, source, "foo")...)
						return evaluator.NewProjectStage(source, columns...)
					})

					test("select foo.*, a from foo", func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "foo")
						columns := append(
							createAllProjectedColumnsFromSource(1, source, "foo"),
							createProjectedColumn(1, source, "foo", "a", "foo", "a"))
						return evaluator.NewProjectStage(source, columns...)
					})

					test("select a, f.* from foo f", func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "f")
						columns := append(
							evaluator.ProjectedColumns{createProjectedColumn(1, source, "f", "a", "f", "a")},
							createAllProjectedColumnsFromSource(1, source, "f")...)
						return evaluator.NewProjectStage(source, columns...)
					})

					test("select * from foo, bar", func() evaluator.PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := evaluator.NewJoinStage(evaluator.CrossJoin, fooSource, barSource, evaluator.SQLTrue)
						fooCols := createAllProjectedColumnsFromSource(1, fooSource, "foo")
						barCols := createAllProjectedColumnsFromSource(1, barSource, "bar")
						return evaluator.NewProjectStage(join, append(fooCols, barCols...)...)
					})

					test("select foo.*, bar.* from foo, bar", func() evaluator.PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := evaluator.NewJoinStage(evaluator.CrossJoin, fooSource, barSource, evaluator.SQLTrue)
						fooCols := createAllProjectedColumnsFromSource(1, fooSource, "foo")
						barCols := createAllProjectedColumnsFromSource(1, barSource, "bar")
						return evaluator.NewProjectStage(join, append(fooCols, barCols...)...)
					})
				})

				Convey("non-star simple queries", func() {
					test("select a from foo", func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return evaluator.NewProjectStage(source, createProjectedColumn(1, source, "foo", "a", "foo", "a"))
					})

					test("select a from foo f", func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "f")
						return evaluator.NewProjectStage(source, createProjectedColumn(1, source, "f", "a", "f", "a"))
					})

					test("select f.a from foo f", func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "f")
						return evaluator.NewProjectStage(source, createProjectedColumn(1, source, "f", "a", "f", "a"))
					})

					test("select a as b from foo", func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return evaluator.NewProjectStage(source, createProjectedColumn(1, source, "foo", "a", "foo", "b"))
					})

					test("select a + 2 from foo", func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return evaluator.NewProjectStage(source,
							evaluator.CreateProjectedColumnFromSQLExpr(1, "a+2",
								evaluator.NewSQLAddExpr(
									evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt),
									evaluator.SQLInt(2),
								),
							),
						)
					})

					test("select a + 2 as b from foo", func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return evaluator.NewProjectStage(source,
							evaluator.CreateProjectedColumnFromSQLExpr(1, "b",
								evaluator.NewSQLAddExpr(
									evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt),
									evaluator.SQLInt(2),
								),
							),
						)
					})

					test("select ASCII(a) from foo", func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "foo")
						scalarFunExpr, _ := evaluator.NewSQLScalarFunctionExpr(
							"ascii",
							[]evaluator.SQLExpr{evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt)},
						)

						return evaluator.NewProjectStage(source,
							evaluator.CreateProjectedColumnFromSQLExpr(1, "ascii(a)", scalarFunExpr),
						)
					})

					test("select BENCHMARK(1, a) from foo", func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return evaluator.NewProjectStage(source,
							evaluator.CreateProjectedColumnFromSQLExpr(1, "benchmark(1, a)",
								evaluator.NewSQLBenchmarkExpr(
									evaluator.SQLInt(1),
									evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt),
								),
							),
						)
					})
				})

				Convey("subqueries", func() {
					Convey("non-correlated", func() {
						test("select a, (select a from bar) from foo", func() evaluator.PlanStage {
							fooSource := createMongoSource(1, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							return evaluator.NewProjectStage(fooSource,
								createProjectedColumn(1, fooSource, "foo", "a", "foo", "a"),
								evaluator.CreateProjectedColumnFromSQLExpr(1, "(select a from bar)",
									evaluator.NewSQLSubqueryExpr(
										false,
										false,
										evaluator.NewProjectStage(barSource, createProjectedColumn(2, barSource, "bar", "a", "bar", "a")),
									),
								),
							)
						})

						test("select a, (select a from bar) as b from foo", func() evaluator.PlanStage {
							fooSource := createMongoSource(1, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							return evaluator.NewProjectStage(fooSource,
								createProjectedColumn(1, fooSource, "foo", "a", "foo", "a"),
								evaluator.CreateProjectedColumnFromSQLExpr(1, "b",
									evaluator.NewSQLSubqueryExpr(
										false,
										false,
										evaluator.NewProjectStage(barSource, createProjectedColumn(2, barSource, "bar", "a", "bar", "a")),
									),
								),
							)
						})

						test("select a, (select foo.a from foo, bar) from foo", func() evaluator.PlanStage {
							foo1Source := createMongoSource(1, "foo", "foo")
							foo2Source := createMongoSource(2, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							join := evaluator.NewJoinStage(evaluator.CrossJoin, foo2Source, barSource, evaluator.SQLTrue)
							return evaluator.NewProjectStage(foo1Source,
								createProjectedColumn(1, foo1Source, "foo", "a", "foo", "a"),
								evaluator.CreateProjectedColumnFromSQLExpr(1, "(select foo.a from foo, bar)",
									evaluator.NewSQLSubqueryExpr(
										false,
										false,
										evaluator.NewProjectStage(join, createProjectedColumn(2, join, "foo", "a", "foo", "a")),
									),
								),
							)
						})

						test("select exists(select 1 from bar) from foo", func() evaluator.PlanStage {
							fooSource := createMongoSource(1, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							return evaluator.NewProjectStage(fooSource,
								evaluator.CreateProjectedColumnFromSQLExpr(1, "exists (select 1 from bar)",
									evaluator.NewSQLExistsExpr(
										evaluator.NewSQLSubqueryExpr(
											false,
											false,
											evaluator.NewProjectStage(barSource, evaluator.CreateProjectedColumnFromSQLExpr(2, "1", evaluator.SQLInt(1))),
										),
									),
								),
							)
						})
					})

					Convey("correlated", func() {
						test("select a, (select foo.a from bar) from foo", func() evaluator.PlanStage {
							fooSource := createMongoSource(1, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							return evaluator.NewProjectStage(
								fooSource,
								createProjectedColumn(1, fooSource, "foo", "a", "foo", "a"),
								evaluator.CreateProjectedColumnFromSQLExpr(1, "(select foo.a from bar)",
									evaluator.NewSQLSubqueryExpr(
										true,
										false,
										evaluator.NewProjectStage(barSource, createProjectedColumn(1, fooSource, "foo", "a", "foo", "a")),
									),
								),
							)
						})

						test("select * from (select b2.d, b2.b from bar b1 inner join bar b2 on (b1.a=b2.b) group by 1, 2) t0 HAVING (sum(1) > 0 )", func() evaluator.PlanStage {
							b1Source := createMongoSource(2, "bar", "b1")
							b2Source := createMongoSource(2, "bar", "b2")
							subqueryAliasName := "t0"

							matcher := evaluator.NewSQLEqualsExpr(
								createSQLColumnExprFromSource(b1Source, "b1", "a"),
								createSQLColumnExprFromSource(b2Source, "b2", "b"),
							)

							join := evaluator.NewJoinStage(evaluator.InnerJoin, b1Source, b2Source, matcher)

							innerGroup := evaluator.NewGroupByStage(
								join,
								[]evaluator.SQLExpr{
									createSQLColumnExprFromSource(join, "b2", "d"),
									createSQLColumnExprFromSource(join, "b2", "b"),
								},
								evaluator.ProjectedColumns{
									createProjectedColumn(2, join, "b2", "b", "b2", "b"),
									createProjectedColumn(2, join, "b2", "d", "b2", "d"),
								},
							)

							subquery := evaluator.NewSubquerySourceStage(
								evaluator.NewProjectStage(
									innerGroup,
									createProjectedColumn(2, join, "b2", "d", "b2", "d"),
									createProjectedColumn(2, join, "b2", "b", "b2", "b"),
								),
								2,
								subqueryAliasName,
							)

							outerGroup := evaluator.NewGroupByStage(
								subquery,
								nil,
								evaluator.ProjectedColumns{
									createProjectedColumn(2, subquery, subqueryAliasName, "d", subqueryAliasName, "d"),
									createProjectedColumn(2, subquery, subqueryAliasName, "b", subqueryAliasName, "b"),
									evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(1)", &evaluator.SQLAggFunctionExpr{
										Name:  "sum",
										Exprs: []evaluator.SQLExpr{evaluator.SQLInt(1)},
									}),
								},
							)

							filter := evaluator.NewFilterStage(
								outerGroup,
								evaluator.NewSQLGreaterThanExpr(
									evaluator.NewSQLColumnExpr(1, "", "", "sum(1)", schema.SQLFloat, schema.MongoNone),
									evaluator.SQLInt(0),
								),
							)

							project := evaluator.NewProjectStage(
								filter,
								createProjectedColumn(1, subquery, subqueryAliasName, "d", subqueryAliasName, "d"),
								createProjectedColumn(1, subquery, subqueryAliasName, "b", subqueryAliasName, "b"),
							)

							return project
						})

					})
				})
			})

			Convey("where", func() {
				test("select a from foo where a", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewFilterStage(source, evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt)),
						createProjectedColumn(1, source, "foo", "a", "foo", "a"),
					)
				})

				test("select a from foo where false", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewFilterStage(source, evaluator.SQLFalse),
						createProjectedColumn(1, source, "foo", "a", "foo", "a"),
					)
				})

				test("select a from foo where true", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewFilterStage(source, evaluator.SQLTrue),
						createProjectedColumn(1, source, "foo", "a", "foo", "a"),
					)
				})

				test("select a from foo where g = true", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewFilterStage(source,
							evaluator.NewSQLEqualsExpr(
								evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "g", schema.SQLBoolean, schema.MongoBool),
								evaluator.SQLTrue,
							),
						),
						createProjectedColumn(1, source, "foo", "a", "foo", "a"),
					)
				})

				test("select a from foo where a > 10", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewFilterStage(source,
							evaluator.NewSQLGreaterThanExpr(
								evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt),
								evaluator.SQLInt(10),
							),
						),
						createProjectedColumn(1, source, "foo", "a", "foo", "a"),
					)
				})

				test("select a as b from foo where b > 10", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewFilterStage(source,
							evaluator.NewSQLGreaterThanExpr(
								evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "b", schema.SQLInt, schema.MongoInt),
								evaluator.SQLInt(10),
							),
						),
						createProjectedColumn(1, source, "foo", "a", "foo", "b"),
					)
				})

				Convey("subqueries", func() {
					Convey("correlated", func() {
						test("select a from foo where (b) = (select b from bar where foo.a = bar.a)", func() evaluator.PlanStage {
							fooSource := createMongoSource(1, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							return evaluator.NewProjectStage(
								evaluator.NewFilterStage(
									fooSource,
									evaluator.NewSQLEqualsExpr(
										createSQLColumnExprFromSource(fooSource, "foo", "b"),
										evaluator.NewSQLSubqueryExpr(
											true,
											false,
											evaluator.NewProjectStage(
												evaluator.NewFilterStage(
													barSource,
													evaluator.NewSQLEqualsExpr(
														createSQLColumnExprFromSource(fooSource, "foo", "a"),
														createSQLColumnExprFromSource(barSource, "bar", "a"),
													),
												),
												createProjectedColumn(2, barSource, "bar", "b", "bar", "b"),
											),
										),
									),
								),
								createProjectedColumn(1, fooSource, "foo", "a", "foo", "a"),
							)
						})

						test("select a from foo f where (b) = (select b from bar where exists(select 1 from foo where f.a = a))", func() evaluator.PlanStage {
							fooSource := createMongoSource(1, "foo", "f")
							barSource := createMongoSource(2, "bar", "bar")
							foo3Source := createMongoSource(3, "foo", "foo")
							return evaluator.NewProjectStage(
								evaluator.NewFilterStage(
									fooSource,
									evaluator.NewSQLEqualsExpr(
										createSQLColumnExprFromSource(fooSource, "f", "b"),
										evaluator.NewSQLSubqueryExpr(
											true,
											false,
											evaluator.NewProjectStage(
												evaluator.NewFilterStage(
													barSource,
													evaluator.NewSQLExistsExpr(
														evaluator.NewSQLSubqueryExpr(
															true,
															false,
															evaluator.NewProjectStage(
																evaluator.NewFilterStage(
																	foo3Source,
																	evaluator.NewSQLEqualsExpr(
																		createSQLColumnExprFromSource(fooSource, "f", "a"),
																		createSQLColumnExprFromSource(foo3Source, "foo", "a"),
																	),
																),
																evaluator.CreateProjectedColumnFromSQLExpr(3, "1", evaluator.SQLInt(1)),
															),
														),
													),
												),
												createProjectedColumn(2, barSource, "bar", "b", "bar", "b"),
											),
										),
									),
								),
								createProjectedColumn(1, fooSource, "f", "a", "f", "a"),
							)
						})

						test("select a from foo where (b) = (select b from bar where exists(select 1 from foo where bar.a = a))", func() evaluator.PlanStage {
							fooSource := createMongoSource(1, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							foo3Source := createMongoSource(3, "foo", "foo")
							return evaluator.NewProjectStage(
								evaluator.NewFilterStage(
									fooSource,
									evaluator.NewSQLEqualsExpr(
										createSQLColumnExprFromSource(fooSource, "foo", "b"),
										evaluator.NewSQLSubqueryExpr(
											false,
											false,
											evaluator.NewProjectStage(
												evaluator.NewFilterStage(
													barSource,
													evaluator.NewSQLExistsExpr(
														evaluator.NewSQLSubqueryExpr(
															true,
															false,
															evaluator.NewProjectStage(
																evaluator.NewFilterStage(
																	foo3Source,
																	evaluator.NewSQLEqualsExpr(
																		createSQLColumnExprFromSource(barSource, "bar", "a"),
																		createSQLColumnExprFromSource(foo3Source, "foo", "a"),
																	),
																),
																evaluator.CreateProjectedColumnFromSQLExpr(3, "1", evaluator.SQLInt(1)),
															),
														),
													),
												),
												createProjectedColumn(2, barSource, "bar", "b", "bar", "b"),
											),
										),
									),
								),
								createProjectedColumn(1, fooSource, "foo", "a", "foo", "a"),
							)
						})
					})
				})
			})

			Convey("group by", func() {
				test("select sum(a) from foo", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewGroupByStage(source,
							nil,
							evaluator.ProjectedColumns{
								evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)", &evaluator.SQLAggFunctionExpr{
									Name:  "sum",
									Exprs: []evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
								}),
							},
						),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(a)", evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})

				test("select sum(a) from foo group by b", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewGroupByStage(source,
							[]evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
							evaluator.ProjectedColumns{
								evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)", &evaluator.SQLAggFunctionExpr{
									Name:  "sum",
									Exprs: []evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
								}),
							},
						),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(a)", evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})

				test("select a, sum(a) from foo group by b", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewGroupByStage(source,
							[]evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
							evaluator.ProjectedColumns{
								createProjectedColumn(1, source, "foo", "a", "foo", "a"),
								evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)", &evaluator.SQLAggFunctionExpr{
									Name:  "sum",
									Exprs: []evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
								}),
							},
						),
						createProjectedColumn(1, source, "foo", "a", "foo", "a"),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(a)", evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})

				test("select sum(a) from foo group by b order by sum(a)", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewOrderByStage(
							evaluator.NewGroupByStage(source,
								[]evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
								evaluator.ProjectedColumns{
									evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)", &evaluator.SQLAggFunctionExpr{
										Name:  "sum",
										Exprs: []evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
									}),
								},
							),
							evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone), true),
						),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(a)", evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})

				test("select sum(a) as sum_a from foo group by b order by sum_a", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewOrderByStage(
							evaluator.NewGroupByStage(source,
								[]evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
								evaluator.ProjectedColumns{
									evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)", &evaluator.SQLAggFunctionExpr{
										Name:  "sum",
										Exprs: []evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
									}),
								},
							),
							evaluator.NewOrderByTerm(evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone), true),
						),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "sum_a", evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})

				test("select sum(a) from foo f group by b order by (select c from foo where f.b = b)", func() evaluator.PlanStage {
					foo1Source := createMongoSource(1, "foo", "f")
					foo2Source := createMongoSource(2, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewOrderByStage(
							evaluator.NewGroupByStage(foo1Source,
								[]evaluator.SQLExpr{createSQLColumnExprFromSource(foo1Source, "f", "b")},
								evaluator.ProjectedColumns{
									createProjectedColumn(1, foo1Source, "f", "b", "f", "b"),
									evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.f.a)", &evaluator.SQLAggFunctionExpr{
										Name:  "sum",
										Exprs: []evaluator.SQLExpr{createSQLColumnExprFromSource(foo1Source, "f", "a")},
									}),
								},
							),
							evaluator.NewOrderByTerm(
								evaluator.NewSQLSubqueryExpr(
									true,
									false,
									evaluator.NewProjectStage(
										evaluator.NewFilterStage(
											foo2Source,
											evaluator.NewSQLEqualsExpr(
												createSQLColumnExprFromSource(foo1Source, "f", "b"),
												createSQLColumnExprFromSource(foo2Source, "foo", "b"),
											),
										),
										createProjectedColumn(2, foo2Source, "foo", "c", "foo", "c"),
									),
								),
								true,
							),
						),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(a)", evaluator.NewSQLColumnExpr(1, "", "", "sum(test.f.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})

				test("select (select sum(foo.a) from foo as f) from foo group by b", func() evaluator.PlanStage {
					foo1Source := createMongoSource(1, "foo", "foo")
					foo2Source := createMongoSource(2, "foo", "f")
					return evaluator.NewProjectStage(
						evaluator.NewGroupByStage(foo1Source,
							[]evaluator.SQLExpr{createSQLColumnExprFromSource(foo1Source, "foo", "b")},
							evaluator.ProjectedColumns{
								evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)", &evaluator.SQLAggFunctionExpr{
									Name: "sum",
									Exprs: []evaluator.SQLExpr{
										createSQLColumnExprFromSource(foo1Source, "foo", "a"),
									},
								}),
							},
						),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "(select sum(foo.a) from foo as f)",
							evaluator.NewSQLSubqueryExpr(
								true,
								false,
								evaluator.NewProjectStage(
									foo2Source,
									evaluator.CreateProjectedColumnFromSQLExpr(2, "sum(foo.a)", evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone)),
								),
							),
						),
					)
				})

				test("select (select sum(f.a + foo.a) from foo f) from foo group by b", func() evaluator.PlanStage {
					foo1Source := createMongoSource(1, "foo", "foo")
					foo2Source := createMongoSource(2, "foo", "f")
					return evaluator.NewProjectStage(
						evaluator.NewGroupByStage(foo1Source,
							[]evaluator.SQLExpr{createSQLColumnExprFromSource(foo1Source, "foo", "b")},
							evaluator.ProjectedColumns{
								createProjectedColumn(1, foo1Source, "foo", "a", "foo", "a"),
							},
						),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "(select sum(f.a+foo.a) from foo as f)",
							evaluator.NewSQLSubqueryExpr(
								true,
								false,
								evaluator.NewProjectStage(
									evaluator.NewGroupByStage(
										foo2Source,
										nil,
										evaluator.ProjectedColumns{
											evaluator.CreateProjectedColumnFromSQLExpr(2, "sum(test.f.a+test.foo.a)", &evaluator.SQLAggFunctionExpr{
												Name: "sum",
												Exprs: []evaluator.SQLExpr{
													evaluator.NewSQLAddExpr(
														createSQLColumnExprFromSource(foo2Source, "f", "a"),
														createSQLColumnExprFromSource(foo1Source, "foo", "a"),
													)},
											}),
										},
									),
									evaluator.CreateProjectedColumnFromSQLExpr(2, "sum(f.a+foo.a)", evaluator.NewSQLColumnExpr(2, "", "", "sum(test.f.a+test.foo.a)", schema.SQLFloat, schema.MongoNone)),
								),
							),
						),
					)
				})
			})

			Convey("having", func() {
				test("select a from foo group by b having sum(a) > 10", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewFilterStage(
							evaluator.NewGroupByStage(source,
								[]evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
								evaluator.ProjectedColumns{
									createProjectedColumn(1, source, "foo", "a", "foo", "a"),
									evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)", &evaluator.SQLAggFunctionExpr{
										Name:  "sum",
										Exprs: []evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
									}),
								},
							),
							evaluator.NewSQLGreaterThanExpr(
								evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone),
								evaluator.SQLInt(10),
							),
						),
						createProjectedColumn(1, source, "foo", "a", "foo", "a"),
					)
				})

				Convey("subqueries", func() {
					Convey("non-correlated", func() {
						test("select a from foo having exists(select 1 from bar)", func() evaluator.PlanStage {
							fooSource := createMongoSource(1, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							return evaluator.NewProjectStage(
								evaluator.NewFilterStage(
									fooSource,
									evaluator.NewSQLExistsExpr(
										evaluator.NewSQLSubqueryExpr(
											false,
											false,
											evaluator.NewProjectStage(
												barSource,
												evaluator.CreateProjectedColumnFromSQLExpr(2, "1", evaluator.SQLInt(1)),
											),
										),
									),
								),
								createProjectedColumn(1, fooSource, "foo", "a", "foo", "a"),
							)
						})
					})
				})
			})

			Convey("distinct", func() {
				test("select distinct a from foo", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewGroupByStage(source,
							[]evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
							evaluator.ProjectedColumns{
								createProjectedColumn(1, source, "foo", "a", "foo", "a"),
							},
						),
						createProjectedColumn(1, source, "foo", "a", "foo", "a"),
					)
				})

				test("select distinct sum(a) from foo", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewGroupByStage(
							evaluator.NewGroupByStage(source,
								nil,
								evaluator.ProjectedColumns{
									evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)", &evaluator.SQLAggFunctionExpr{
										Name:  "sum",
										Exprs: []evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
									}),
								},
							),
							[]evaluator.SQLExpr{evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone)},
							evaluator.ProjectedColumns{
								evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)", evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone)),
							},
						),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(a)", evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})

				test("select distinct sum(a) from foo having sum(a) > 20", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewGroupByStage(
							evaluator.NewFilterStage(
								evaluator.NewGroupByStage(source,
									nil,
									evaluator.ProjectedColumns{
										evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)", &evaluator.SQLAggFunctionExpr{
											Name:  "sum",
											Exprs: []evaluator.SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
										}),
									},
								),
								evaluator.NewSQLGreaterThanExpr(
									evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone),
									evaluator.SQLInt(20),
								),
							),
							[]evaluator.SQLExpr{evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone)},
							evaluator.ProjectedColumns{
								evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(test.foo.a)", evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone)),
							},
						),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "sum(a)", evaluator.NewSQLColumnExpr(1, "", "", "sum(test.foo.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})
			})

			Convey("order by", func() {
				test("select a from foo order by a", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewOrderByStage(source,
							evaluator.NewOrderByTerm(
								evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt),
								true,
							),
						),
						createProjectedColumn(1, source, "foo", "a", "foo", "a"),
					)
				})

				test("select a as b from foo order by b", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewOrderByStage(source,
							evaluator.NewOrderByTerm(
								evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt),
								true,
							),
						),
						createProjectedColumn(1, source, "foo", "a", "foo", "b"),
					)
				})

				test("select a from foo order by foo.a", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewOrderByStage(source,
							evaluator.NewOrderByTerm(
								evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt),
								true,
							),
						),
						createProjectedColumn(1, source, "foo", "a", "foo", "a"),
					)
				})

				test("select a as b from foo order by foo.a", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewOrderByStage(source,
							evaluator.NewOrderByTerm(
								evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt),
								true,
							),
						),
						createProjectedColumn(1, source, "foo", "a", "foo", "b"),
					)
				})

				test("select a from foo order by 1", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewOrderByStage(source,
							evaluator.NewOrderByTerm(
								evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt),
								true,
							),
						),
						createProjectedColumn(1, source, "foo", "a", "foo", "a"),
					)
				})

				test("select * from foo order by 2", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewOrderByStage(source,
							evaluator.NewOrderByTerm(
								evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt),
								true,
							),
						),
						createAllProjectedColumnsFromSource(1, source, "foo")...,
					)
				})

				test("select foo.* from foo order by 2", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewOrderByStage(source,
							evaluator.NewOrderByTerm(
								evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt),
								true,
							),
						),
						createAllProjectedColumnsFromSource(1, source, "foo")...,
					)
				})

				test("select foo.*, foo.a from foo order by 2", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					columns := append(createAllProjectedColumnsFromSource(1, source, "foo"), createProjectedColumn(1, source, "foo", "a", "foo", "a"))
					return evaluator.NewProjectStage(
						evaluator.NewOrderByStage(source,
							evaluator.NewOrderByTerm(
								evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt),
								true,
							),
						),
						columns...,
					)
				})

				test("select a from foo order by -1", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewOrderByStage(source,
							evaluator.NewOrderByTerm(
								evaluator.SQLInt(-1),
								true,
							),
						),
						createProjectedColumn(1, source, "foo", "a", "foo", "a"),
					)
				})

				test("select a + b as c from foo order by c - b", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewOrderByStage(source,
							evaluator.NewOrderByTerm(
								evaluator.NewSQLSubtractExpr(
									evaluator.NewSQLAddExpr(
										evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt),
										evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "b", schema.SQLInt, schema.MongoInt),
									),
									evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "b", schema.SQLInt, schema.MongoInt),
								),
								true,
							),
						),
						evaluator.CreateProjectedColumnFromSQLExpr(1, "c",
							evaluator.NewSQLAddExpr(
								evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "a", schema.SQLInt, schema.MongoInt),
								evaluator.NewSQLColumnExpr(1, defaultDbName, "foo", "b", schema.SQLInt, schema.MongoInt),
							),
						),
					)
				})

				Convey("subqueries", func() {
					Convey("non-correlated", func() {
						test("select a from foo order by (select a from bar)", func() evaluator.PlanStage {
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
												createProjectedColumn(2, barSource, "bar", "a", "bar", "a"),
											),
										),
										true,
									),
								),
								createProjectedColumn(1, fooSource, "foo", "a", "foo", "a"),
							)
						})
					})

				})
			})

			Convey("limit", func() {
				test("select a from foo limit 10", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewLimitStage(source, 0, 10),
						createProjectedColumn(1, source, "foo", "a", "foo", "a"),
					)
				})

				test("select a from foo limit 10, 20", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewProjectStage(
						evaluator.NewLimitStage(source, 10, 20),
						createProjectedColumn(1, source, "foo", "a", "foo", "a"),
					)
				})

				test("select a from foo limit 10,0", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewEmptyStage([]*evaluator.Column{
						createProjectedColumn(1, source, "foo", "a", "foo", "a").Column,
					}, collation.Default)
				})

				test("select a from foo limit 0, 0", func() evaluator.PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return evaluator.NewEmptyStage([]*evaluator.Column{
						createProjectedColumn(1, source, "foo", "a", "foo", "a").Column,
					}, collation.Default)
				})

				testVariables("select a from foo",
					func() *variable.Container {
						vars := &variable.Container{
							MongoDBInfo: testInfo,
						}
						vars.SetSystemVariable(variable.SQLSelectLimit, 10)
						return vars
					},
					func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return evaluator.NewLimitStage(
							evaluator.NewProjectStage(
								source,
								createProjectedColumn(1, source, "foo", "a", "foo", "a"),
							),
							0,
							10,
						)
					})

				testVariables("select b from foo",
					func() *variable.Container {
						vars := &variable.Container{
							MongoDBInfo: testInfo,
						}
						vars.SetSystemVariable(variable.SQLSelectLimit, uint64(18446744073709551615))
						return vars
					},
					func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return evaluator.NewProjectStage(
							source,
							createProjectedColumn(1, source, "foo", "b", "foo", "b"),
						)
					})

				testVariables("select b from foo limit 10, 20",
					func() *variable.Container {
						vars := &variable.Container{
							MongoDBInfo: testInfo,
						}
						vars.SetSystemVariable(variable.SQLSelectLimit, 5)
						return vars
					},
					func() evaluator.PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return evaluator.NewProjectStage(
							evaluator.NewLimitStage(source, 10, 20),
							createProjectedColumn(1, source, "foo", "b", "foo", "b"),
						)
					})

			})

			Convey("count", func() {
				test("select count(*) from foo", func() evaluator.PlanStage {
					column := evaluator.NewColumn(1, "", "", "", "count(*)", "", "", schema.SQLInt, schema.MongoNone, false)
					projectedColumn := createProjectedColumnFromColumn(1, column, "", "count(*)")
					source := createMongoSource(1, "foo", "foo")
					countStage := evaluator.NewCountStage(source.(*evaluator.MongoSourceStage), projectedColumn)
					return evaluator.NewProjectStage(countStage, projectedColumn)
				})

				test("select count(*) as c from foo", func() evaluator.PlanStage {
					column := evaluator.NewColumn(1, "", "", "", "c", "", "", schema.SQLInt, schema.MongoNone, false)
					projectedColumn := createProjectedColumnFromColumn(1, column, "", "c")
					source := createMongoSource(1, "foo", "foo")
					countStage := evaluator.NewCountStage(source.(*evaluator.MongoSourceStage), projectedColumn)
					return evaluator.NewProjectStage(countStage, projectedColumn)
				})

				test("select count(*) as c from foo order by a", func() evaluator.PlanStage {
					column := evaluator.NewColumn(1, "", "", "", "c", "", "", schema.SQLInt, schema.MongoNone, false)
					projectedColumn := createProjectedColumnFromColumn(1, column, "", "c")
					source := createMongoSource(1, "foo", "foo")
					countStage := evaluator.NewCountStage(source.(*evaluator.MongoSourceStage), projectedColumn)
					return evaluator.NewProjectStage(countStage, projectedColumn)
				})

				test("select count(*) as c from foo order by 1", func() evaluator.PlanStage {
					column := evaluator.NewColumn(1, "", "", "", "c", "", "", schema.SQLInt, schema.MongoNone, false)
					projectedColumn := createProjectedColumnFromColumn(1, column, "", "c")
					source := createMongoSource(1, "foo", "foo")
					countStage := evaluator.NewCountStage(source.(*evaluator.MongoSourceStage), projectedColumn)
					return evaluator.NewProjectStage(countStage, projectedColumn)
				})

				test("select count(*) from foo as c", func() evaluator.PlanStage {
					column := evaluator.NewColumn(1, "", "", "", "count(*)", "", "", schema.SQLInt, schema.MongoNone, false)
					projectedColumn := createProjectedColumnFromColumn(1, column, "", "count(*)")
					source := createMongoSource(1, "foo", "c")
					countStage := evaluator.NewCountStage(source.(*evaluator.MongoSourceStage), projectedColumn)
					return evaluator.NewProjectStage(countStage, projectedColumn)
				})

			})

			Convey("errors", func() {
				testError("select ABASDD()", "scalar function 'abasdd' is not supported")
				testError("select a", `ERROR 1054 (42S22): Unknown column 'a' in 'field list'`)
				testError("select a from idk", `ERROR 1146 (42S02): Table 'test.idk' doesn't exist`)
				testError("select idk from foo", `ERROR 1054 (42S22): Unknown column 'idk' in 'field list'`)
				testError("select f.a from foo", `ERROR 1054 (42S22): Unknown column 'f.a' in 'field list'`)
				testError("select foo.a from foo f", `ERROR 1054 (42S22): Unknown column 'foo.a' in 'field list'`)
				testError("select a + idk from foo", `ERROR 1054 (42S22): Unknown column 'idk' in 'field list'`)

				testError("select a, * from foo", `ERROR 1149 (42000): Cannot have a '*' in conjunction with any other columns`)
				testError("select *, * from foo", `ERROR 1149 (42000): Cannot have a '*' in conjunction with any other columns`)
				testError("select *, a from foo", `ERROR 1149 (42000): Cannot have a '*' in conjunction with any other columns`)

				testError("select a from foo, bar", `ERROR 1052 (23000): Column 'a' in field list is ambiguous`)
				testError("select foo.a from foo f, bar b", `ERROR 1054 (42S22): Unknown column 'foo.a' in 'field list'`)
				testError("select f.a, * from foo f, bar b", `ERROR 1149 (42000): Cannot have a '*' in conjunction with any other columns`)
				testError("select a from foo f, bar b", `ERROR 1052 (23000): Column 'a' in field list is ambiguous`)
				testError("select a, b as a from foo order by a", `ERROR 1052 (23000): Column 'a' in order clause is ambiguous`)

				testError("select (select a, b from foo) from foo", `ERROR 1241 (21000): Operand should contain 1 column(s)`)
				testError("select * from (select a, b as a from foo) f", `ERROR 1060 (42S21): Duplicate column name 'f.a'`)
				testError("select foo.a from (select a from foo)", `ERROR 1248 (42000): Every derived table must have its own alias`)

				testError("select a from foo limit -10", `ERROR 1149 (42000): Rowcount cannot be negative`)
				testError("select a from foo limit -10, 20", `ERROR 1149 (42000): Offset cannot be negative`)
				testError("select a from foo limit -10, -20", `ERROR 1149 (42000): Offset cannot be negative`)
				testError("select a from foo limit b", `ERROR 1691 (HY000): A variable of a non-integer based type in LIMIT clause`)
				testError("select a from foo limit 'c'", `ERROR 1691 (HY000): A variable of a non-integer based type in LIMIT clause`)

				testError("select a from foo, (select * from (select * from bar where foo.b = b) asdf) wegqweg", `ERROR 1054 (42S22): Unknown column 'foo.b' in 'where clause'`)
				testError("select a from foo where sum(a) = 10", `ERROR 1111 (HY000): Invalid use of group function`)

				testError("select a from foo order by 2", `ERROR 1054 (42S22): Unknown column '2' in 'order clause'`)
				testError("select a from foo order by idk", `ERROR 1054 (42S22): Unknown column 'idk' in 'order clause'`)

				testError("select sum(a) from foo group by sum(a)", `ERROR 1056 (42000): Can't group on 'sum(test.foo.a)'`)
				testError("select sum(a) from foo group by (a + sum(a))", `ERROR 1056 (42000): Can't group on 'sum(test.foo.a)'`)
				testError("select sum(a) from foo group by 1", `ERROR 1056 (42000): Can't group on 'sum(test.foo.a)'`)
				testError("select a+sum(a) from foo group by 1", `ERROR 1056 (42000): Can't group on 'sum(test.foo.a)'`)
				testError("select sum(a) from foo group by 2", `ERROR 1054 (42S22): Unknown column '2' in 'group clause'`)

				testError("select a from foo, foo", `ERROR 1066 (42000): Not unique table/alias: 'foo'`)
				testError("select a from foo as bar, bar", `ERROR 1066 (42000): Not unique table/alias: 'bar'`)
				testError("select a from foo as g, foo as g", `ERROR 1066 (42000): Not unique table/alias: 'g'`)

				testError("select a from foo left outer join bar where a = 10", `ERROR 1064 (42000): A left join requires criteria`)

				testError("select * from bar natural join baz using (id)", "ERROR 1064 (42000): A natural join cannot have join criteria")
				testError("select * from bar natural join baz on bar.id=baz.id", "ERROR 1064 (42000): A natural join cannot have join criteria")
				testError("select * from foo natural left join bar using (id)", "ERROR 1064 (42000): A natural left join cannot have join criteria")
				testError("select * from foo natural right join bar using (id)", "ERROR 1064 (42000): A natural right join cannot have join criteria")

				testError("select bar.d, baz.a from bar join baz using (tomato)", `ERROR 1054 (42S22): Unknown column 'bar.tomato' in 'from clause'`)
				testError("select * from baz join bar using (d)", `ERROR 1054 (42S22): Unknown column 'baz.d' in 'from clause'`)
				testError("select bar.d, baz.a from bar join (select * from baz join foo) using (c)", `ERROR 1248 (42000): Every derived table must have its own alias`)
				testError("select bar.d, biz.a from bar join (select * from baz join foo) as biz using (c)", `ERROR 1060 (42S21): Duplicate column name 'biz._id'`)
				testError("select * from bar join foo join baz using (c)", "ERROR 1054 (42S22): Unknown column 'baz.c' in 'from clause'")
				testError("select * from bar join foo join baz using (_id)", "ERROR 1052 (23000): Column '_id' in from clause is ambiguous")
				testError("select * from baz join bar join foo using (c)", "ERROR 1054 (42S22): Unknown column 'c' in 'from clause'")

				testError("select * from (foo join bar) natural join baz", "ERROR 1052 (23000): Column '_id' in from clause is ambiguous")
				testError("select * from foo join bar using (b) natural join baz", "ERROR 1052 (23000): Column '_id' in from clause is ambiguous")
				testError("select bar.d, biz.a from bar natural join (select * from baz join foo) as biz", `ERROR 1060 (42S21): Duplicate column name 'biz._id'`)
				testError("select * from foo left join bar natural join baz using (id)", "ERROR 1064 (42000): A natural join cannot have join criteria")
			})
		})
	})
}

func TestAlgebrizeCommand(t *testing.T) {

	testSchema := evaluator.MustLoadSchema(testSchema1)
	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)
	testVars := evaluator.CreateTestVariables(testInfo)
	testCatalog := evaluator.GetCatalogFromSchema(testSchema, testVars)
	defaultDbName := "test"

	test := func(sql string, expectedPlanFactory func() evaluator.Command) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			actual, err := evaluator.AlgebrizeCommand(statement, defaultDbName, testVars, testCatalog)
			So(err, ShouldBeNil)

			expected := expectedPlanFactory()

			if ShouldResemble(actual, expected) != "" {
				fmt.Printf("\nExpected: %# v", pretty.Formatter(expected))
				fmt.Printf("\nActual: %# v", pretty.Formatter(actual))
			}

			So(actual, ShouldResemble, expected)
		})
	}

	createMongoSource := func(selectID int, tableName, aliasName string) evaluator.PlanStage {
		db, _ := testCatalog.Database(defaultDbName)
		table, _ := db.Table(tableName)
		r := evaluator.NewMongoSourceStage(db, table.(*catalog.MongoTable), selectID, aliasName)
		return r
	}

	Convey("Subject: Algebrize Kill Statements", t, func() {
		test("kill 3", func() evaluator.Command {
			return evaluator.NewKillCommand(evaluator.SQLInt(3), evaluator.KillConnection)
		})
		test("kill query 3", func() evaluator.Command {
			return evaluator.NewKillCommand(evaluator.SQLInt(3), evaluator.KillQuery)
		})
		test("kill query 5*3", func() evaluator.Command {
			return evaluator.NewKillCommand(
				evaluator.NewSQLMultiplyExpr(
					evaluator.SQLInt(5),
					evaluator.SQLInt(3),
				),
				evaluator.KillQuery,
			)
		})
		test("kill connection 5-3", func() evaluator.Command {
			return evaluator.NewKillCommand(
				evaluator.NewSQLSubtractExpr(
					evaluator.SQLInt(5),
					evaluator.SQLInt(3),
				),
				evaluator.KillConnection,
			)
		})
	})

	Convey("Subject: Algebrize Flush Statements", t, func() {
		test("flush logs", func() evaluator.Command {
			return evaluator.NewFlushCommand(evaluator.FlushLogs)
		})
		test("flush sample", func() evaluator.Command {
			return evaluator.NewFlushCommand(evaluator.FlushSample)
		})
	})

	Convey("Subject: Algebrize Set Statements", t, func() {
		test("set @t1 = 132", func() evaluator.Command {
			return evaluator.NewSetCommand(
				[]*evaluator.SQLAssignmentExpr{
					evaluator.NewSQLAssignmentExpr(
						evaluator.NewSQLVariableExpr(
							"t1",
							variable.UserKind,
							variable.SessionScope,
							schema.SQLNone,
						),
						evaluator.SQLInt(132),
					),
				},
			)
		})

		test("set @@max_allowed_packet = 12", func() evaluator.Command {
			return evaluator.NewSetCommand(
				[]*evaluator.SQLAssignmentExpr{
					evaluator.NewSQLAssignmentExpr(
						evaluator.NewSQLVariableExpr(
							"max_allowed_packet",
							variable.SystemKind,
							variable.SessionScope,
							schema.SQLInt,
						),
						evaluator.SQLInt(12),
					),
				},
			)
		})

		test("set @@global.max_allowed_packet = 12", func() evaluator.Command {
			return evaluator.NewSetCommand(
				[]*evaluator.SQLAssignmentExpr{
					evaluator.NewSQLAssignmentExpr(
						evaluator.NewSQLVariableExpr(
							"max_allowed_packet",
							variable.SystemKind,
							variable.GlobalScope,
							schema.SQLInt,
						),
						evaluator.SQLInt(12),
					),
				},
			)
		})

		test("set @@global.max_allowed_packet = (select a from foo)", func() evaluator.Command {
			fooSource := createMongoSource(1, "foo", "foo")
			return evaluator.NewSetCommand(
				[]*evaluator.SQLAssignmentExpr{
					evaluator.NewSQLAssignmentExpr(
						evaluator.NewSQLVariableExpr(
							"max_allowed_packet",
							variable.SystemKind,
							variable.GlobalScope,
							schema.SQLInt,
						),
						evaluator.NewSQLSubqueryExpr(
							false,
							false,
							evaluator.NewProjectStage(
								fooSource,
								createProjectedColumn(1, fooSource, "foo", "a", "foo", "a"),
							),
						),
					),
				},
			)
		})

		test("set @@max_allowed_packet=12, @interactive_timeout=1111", func() evaluator.Command {
			return evaluator.NewSetCommand(
				[]*evaluator.SQLAssignmentExpr{
					evaluator.NewSQLAssignmentExpr(
						evaluator.NewSQLVariableExpr(
							"max_allowed_packet",
							variable.SystemKind,
							variable.SessionScope,
							schema.SQLInt,
						),
						evaluator.SQLInt(12),
					),
					evaluator.NewSQLAssignmentExpr(
						evaluator.NewSQLVariableExpr(
							"interactive_timeout",
							variable.UserKind,
							variable.SessionScope,
							schema.SQLNone,
						),
						evaluator.SQLInt(1111),
					),
				},
			)
		})
	})
}

func TestAlgebrizeExpr(t *testing.T) {
	testSchema := evaluator.MustLoadSchema(testSchema1)
	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)
	testVars := evaluator.CreateTestVariables(testInfo)
	testCatalog := evaluator.GetCatalogFromSchema(testSchema, testVars)
	testDB, _ := testCatalog.Database("test")
	fooTable, _ := testDB.Table("foo")
	source := evaluator.NewMongoSourceStage(testDB, fooTable.(*catalog.MongoTable), 1, "foo")

	test := func(sql string, expected evaluator.SQLExpr) {
		Convey(sql, func() {
			statement, err := parser.Parse("select " + sql + " from foo")
			So(err, ShouldBeNil)

			actualPlan, err := evaluator.AlgebrizeQuery(statement, "test", testVars, testCatalog)
			So(err, ShouldBeNil)
			actual := evaluator.GetProjectProjectedColumnExpr(actualPlan)
			So(actual, ShouldResemble, expected)
		})
	}

	testError := func(sql, message string) {
		Convey(sql, func() {
			statement, err := parser.Parse("select " + sql + " from foo")
			So(err, ShouldBeNil)

			actual, err := evaluator.AlgebrizeQuery(statement, "test", testVars, testCatalog)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, message)
			So(actual, ShouldBeNil)
		})
	}

	createSQLColumnExpr := func(columnName string) evaluator.SQLColumnExpr {
		for _, c := range source.Columns() {
			if c.Name == columnName {
				return evaluator.NewSQLColumnExpr(1, c.Database, c.Table, c.Name, c.SQLType, c.MongoType)
			}
		}

		panic("column not found")
	}

	Convey("Subject: Algebrize Expressions", t, func() {
		Convey("And", func() {
			test("a = 1 AND b = 2", evaluator.NewSQLAndExpr(
				evaluator.NewSQLEqualsExpr(
					createSQLColumnExpr("a"),
					evaluator.SQLInt(1),
				),
				evaluator.NewSQLEqualsExpr(
					createSQLColumnExpr("b"),
					evaluator.SQLInt(2),
				),
			))
		})
		Convey("Add", func() {
			test("a + 1", evaluator.NewSQLAddExpr(
				createSQLColumnExpr("a"),
				evaluator.SQLInt(1),
			))
		})

		SkipConvey("Case", func() {
		})

		Convey("Date", func() {
			expected := time.Date(2006, time.December, 31, 0, 0, 0, 0, time.UTC)
			test("DATE '2006-12-31'", evaluator.SQLDate{Time: expected})
			test("DATE '06-12-31'", evaluator.SQLDate{Time: expected})
			test("DATE '20061231'", evaluator.SQLDate{Time: expected})
			test("DATE '061231'", evaluator.SQLDate{Time: expected})
			testError("DATE '2014-13-07'", "ERROR 1525 (HY000): Incorrect DATE value: '2014-13-07'")
			testError("DATE '2014-12-32'", "ERROR 1525 (HY000): Incorrect DATE value: '2014-12-32'")
			testError("DATE '2006-12-31 10:32:46'", "ERROR 1525 (HY000): Incorrect DATE value: '2006-12-31 10:32:46'")
		})

		Convey("Divide", func() {
			test("a / 1", evaluator.NewSQLDivideExpr(
				createSQLColumnExpr("a"),
				evaluator.SQLInt(1),
			))
		})

		Convey("Equals", func() {
			test("a = 1", evaluator.NewSQLEqualsExpr(
				createSQLColumnExpr("a"),
				evaluator.SQLInt(1),
			))
			test("g = 0", evaluator.NewSQLEqualsExpr(
				createSQLColumnExpr("g"),
				evaluator.NewSQLConvertExpr(
					evaluator.SQLInt(0),
					schema.SQLBoolean,
					evaluator.SQLNone,
				),
			))
			test("g = 1", evaluator.NewSQLEqualsExpr(
				createSQLColumnExpr("g"),
				evaluator.NewSQLConvertExpr(
					evaluator.SQLInt(1),
					schema.SQLBoolean,
					evaluator.SQLNone,
				),
			))
			test("g = 2", evaluator.NewSQLEqualsExpr(
				evaluator.NewSQLConvertExpr(
					createSQLColumnExpr("g"),
					schema.SQLInt,
					evaluator.SQLNone,
				),
				evaluator.SQLInt(2),
			))
			test("0 = g", evaluator.NewSQLEqualsExpr(
				createSQLColumnExpr("g"),
				evaluator.NewSQLConvertExpr(
					evaluator.SQLInt(0),
					schema.SQLBoolean,
					evaluator.SQLNone,
				),
			))
			test("1 = g", evaluator.NewSQLEqualsExpr(
				createSQLColumnExpr("g"),
				evaluator.NewSQLConvertExpr(
					evaluator.SQLInt(1),
					schema.SQLBoolean,
					evaluator.SQLNone,
				),
			))
			test("2 = g", evaluator.NewSQLEqualsExpr(
				evaluator.SQLInt(2),
				evaluator.NewSQLConvertExpr(
					createSQLColumnExpr("g"),
					schema.SQLInt,
					evaluator.SQLNone,
				),
			))
		})

		SkipConvey("Exists", func() {
		})

		Convey("Greater Than", func() {
			test("a > 1", evaluator.NewSQLGreaterThanExpr(
				createSQLColumnExpr("a"),
				evaluator.SQLInt(1),
			))
		})

		Convey("Greater Than Or Equal", func() {
			test("a >= 1", evaluator.NewSQLGreaterThanOrEqualExpr(
				createSQLColumnExpr("a"),
				evaluator.SQLInt(1),
			))
		})

		SkipConvey("In", func() {
		})

		Convey("Is", func() {
			test("a is true", evaluator.NewSQLIsExpr(
				createSQLColumnExpr("a"),
				evaluator.SQLTrue,
			))
		})

		Convey("Is Not", func() {
			test("a is not true", evaluator.NewSQLNotExpr(
				evaluator.NewSQLIsExpr(
					createSQLColumnExpr("a"),
					evaluator.SQLTrue,
				),
			))
		})

		Convey("Is Null", func() {
			test("a IS NULL", evaluator.NewSQLIsExpr(
				createSQLColumnExpr("a"),
				evaluator.SQLNull,
			))
		})

		Convey("Is Not Null", func() {
			test("a IS NOT NULL", evaluator.NewSQLNotExpr(
				evaluator.NewSQLIsExpr(
					createSQLColumnExpr("a"),
					evaluator.SQLNull,
				),
			))
		})

		Convey("Less Than", func() {
			test("a < 1", evaluator.NewSQLLessThanExpr(
				createSQLColumnExpr("a"),
				evaluator.SQLInt(1),
			))
		})

		Convey("Less Than Or Equal", func() {
			test("a <= 1", evaluator.NewSQLLessThanOrEqualExpr(
				createSQLColumnExpr("a"),
				evaluator.SQLInt(1),
			))
		})

		SkipConvey("Like", func() {
		})

		Convey("Multiple", func() {
			test("a * 1", evaluator.NewSQLMultiplyExpr(
				createSQLColumnExpr("a"),
				evaluator.SQLInt(1),
			))
		})

		Convey("Not", func() {
			test("NOT a", evaluator.NewSQLNotExpr(
				createSQLColumnExpr("a"),
			))
		})

		Convey("NotEquals", func() {
			test("a != 1", evaluator.NewSQLNotEqualsExpr(
				createSQLColumnExpr("a"),
				evaluator.SQLInt(1),
			))
		})

		Convey("NullSafeEquals", func() {
			test("a <=> 1", evaluator.NewSQLNullSafeEqualsExpr(
				createSQLColumnExpr("a"),
				evaluator.SQLInt(1),
			))
		})

		SkipConvey("Not In", func() {
		})

		Convey("Null", func() {
			test("NULL", evaluator.SQLNull)
		})

		Convey("True", func() {
			test("TRUE", evaluator.SQLTrue)
		})

		Convey("False", func() {
			test("FALSE", evaluator.SQLFalse)
		})

		Convey("Number", func() {
			test("20", evaluator.SQLInt(20))
			test("-20", evaluator.SQLInt(-20))
			test("202E-1", evaluator.SQLFloat(20.2))
			test("-202E-1", evaluator.SQLFloat(-20.2))
			test("20.2", evaluator.SQLDecimal128(decimal.New(202, -1)))
			test("-20.2", evaluator.SQLDecimal128(decimal.New(-202, -1)))
			d, _ := decimal.NewFromString("100000000000000000000000000000000000")
			test("100000000000000000000000000000000000", evaluator.SQLDecimal128(d))

			oldVersionArray := testInfo.VersionArray
			testInfo.VersionArray = []uint8{3, 2, 0}
			test("30.2", evaluator.SQLFloat(30.2))
			test("-30.2", evaluator.SQLFloat(-30.2))
			f, _ := strconv.ParseFloat("1000000000000000000000000000000000000", 64)
			test("1000000000000000000000000000000000000", evaluator.SQLFloat(f))
			testInfo.VersionArray = oldVersionArray
		})

		Convey("Or", func() {
			test("a = 1 OR b = 2", evaluator.NewSQLOrExpr(
				evaluator.NewSQLEqualsExpr(
					createSQLColumnExpr("a"),
					evaluator.SQLInt(1),
				),
				evaluator.NewSQLEqualsExpr(
					createSQLColumnExpr("b"),
					evaluator.SQLInt(2),
				),
			))
		})

		Convey("Paren Boolean", func() {
			test("(1)", evaluator.SQLInt(1))
		})

		Convey("Range", func() {
			test("a BETWEEN 0 AND 20", evaluator.NewSQLAndExpr(
				evaluator.NewSQLGreaterThanOrEqualExpr(
					createSQLColumnExpr("a"),
					evaluator.SQLInt(0),
				),
				evaluator.NewSQLLessThanOrEqualExpr(
					createSQLColumnExpr("a"),
					evaluator.SQLInt(20),
				),
			))

			test("a NOT BETWEEN 0 AND 20", evaluator.NewSQLNotExpr(
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLGreaterThanOrEqualExpr(
						createSQLColumnExpr("a"),
						evaluator.SQLInt(0),
					),
					evaluator.NewSQLLessThanOrEqualExpr(
						createSQLColumnExpr("a"),
						evaluator.SQLInt(20),
					),
				),
			))
		})

		Convey("Scalar Function", func() {
			f, _ := evaluator.NewSQLScalarFunctionExpr(
				"ascii",
				[]evaluator.SQLExpr{createSQLColumnExpr("a")},
			)
			test("ascii(a)", f)

			test("benchmark(1, a)", evaluator.NewSQLBenchmarkExpr(
				evaluator.SQLInt(1),
				createSQLColumnExpr("a"),
			))
		})

		SkipConvey("Subquery", func() {
		})

		Convey("Subtract", func() {
			test("a - 1", evaluator.NewSQLSubtractExpr(
				createSQLColumnExpr("a"),
				evaluator.SQLInt(1),
			))
		})

		Convey("Time", func() {
			expected := time.Date(0, 1, 1, 10, 32, 46, 5000, time.UTC)
			test("TIME '10:32:46.000005'", evaluator.SQLTimestamp{Time: expected})
			test("TIME '103246.000005'", evaluator.SQLTimestamp{Time: expected})

			testError("TIME '2014-12-32'", "ERROR 1525 (HY000): Incorrect TIME value: '2014-12-32'")
			testError("TIME '2006-12-31 10:32:46.000005'", "ERROR 1525 (HY000): Incorrect TIME value: '2006-12-31 10:32:46.000005'")
		})

		Convey("Timestamp", func() {
			expected := time.Date(2014, 6, 7, 10, 32, 46, 5000, time.UTC)
			test("TIMESTAMP '2014-06-07 10:32:46.000005'", evaluator.SQLTimestamp{Time: expected})
			test("TIMESTAMP '2014-6-7 10:32:46.000005'", evaluator.SQLTimestamp{Time: expected})
			test("TIMESTAMP '14-06-07 10:32:46.000005'", evaluator.SQLTimestamp{Time: expected})
			test("TIMESTAMP '14-6-7 10:32:46.000005'", evaluator.SQLTimestamp{Time: expected})
			test("TIMESTAMP '2014:06:07 10:32:46.000005'", evaluator.SQLTimestamp{Time: expected})
			test("TIMESTAMP '14:06:07 10:32:46.000005'", evaluator.SQLTimestamp{Time: expected})
			test("TIMESTAMP '20140607103246.000005'", evaluator.SQLTimestamp{Time: expected})
			test("TIMESTAMP '140607103246.000005'", evaluator.SQLTimestamp{Time: expected})
			test("TIMESTAMP '146.07103246.000005'", evaluator.SQLTimestamp{Time: expected})
			test("TIMESTAMP '14.06.07.10.32.46.000005'", evaluator.SQLTimestamp{Time: expected})

			testError("TIMESTAMP '2014-06-07'", "ERROR 1525 (HY000): Incorrect DATETIME value: '2014-06-07'")
		})

		Convey("Tuple", func() {
			test("(a)", createSQLColumnExpr("a"))
		})

		Convey("Unary Minus", func() {
			test("-a", evaluator.NewSQLUnaryMinusExpr(createSQLColumnExpr("a")))
			test("-c", evaluator.NewSQLUnaryMinusExpr(createSQLColumnExpr("c")))
			test("-g", evaluator.NewSQLUnaryMinusExpr(evaluator.NewSQLConvertExpr(createSQLColumnExpr("g"), schema.SQLInt, evaluator.SQLNone)))
			test("-_id", evaluator.NewSQLUnaryMinusExpr(evaluator.NewSQLConvertExpr(createSQLColumnExpr("_id"), schema.SQLInt, evaluator.SQLNone)))
		})

		Convey("Unary Tilde", func() {
			test("~a", evaluator.NewSQLUnaryTildeExpr(createSQLColumnExpr("a")))
		})

		Convey("Varchar", func() {
			test("'a'", evaluator.SQLVarchar("a"))
		})

		Convey("Variable", func() {
			varGlobal := evaluator.NewSQLVariableExpr("sql_auto_is_null", variable.SystemKind, variable.GlobalScope, schema.SQLBoolean)
			varSession := evaluator.NewSQLVariableExpr("sql_auto_is_null", variable.SystemKind, variable.SessionScope, schema.SQLBoolean)

			test("@@global.sql_auto_is_null", varGlobal)
			test("@@session.sql_auto_is_null", varSession)
			test("@@local.sql_auto_is_null", varSession)
			test("@@sql_auto_is_null", varSession)
			test("@hmmm", evaluator.NewSQLVariableExpr("hmmm", variable.UserKind, variable.SessionScope, schema.SQLNone))
		})

	})
}

func TestNoSharedPipelines(t *testing.T) {
	sql := "select _id from merge_b limit 2"

	testSchema := evaluator.MustLoadSchema(testSchema4)
	testInfo := evaluator.GetMongoDBInfo([]uint8{3, 2}, testSchema, mongodb.AllPrivileges)
	testVariables := evaluator.CreateTestVariables(testInfo)
	testCatalog := evaluator.GetCatalogFromSchema(testSchema, testVariables)
	defaultDbName := "test"

	Convey("Subject: NoSharedPipelines", t, func() {
		statement, err := parser.Parse(sql)
		So(err, ShouldBeNil)

		plan, err := evaluator.AlgebrizeQuery(statement, defaultDbName, testVariables, testCatalog)
		So(err, ShouldBeNil)

		expectedPipelines := [][]bson.D{
			{{
				bson.DocElem{
					Name: "$unwind",
					Value: bson.D{
						bson.DocElem{
							Name:  "includeArrayIndex",
							Value: "b_idx",
						},
						bson.DocElem{
							Name:  "path",
							Value: "$b",
						},
					},
				},
			}},
		}

		actualPipelines := evaluator.GetNodePipeline(plan)
		So(actualPipelines, ShouldResemble, expectedPipelines)

		db, err := testCatalog.Database("test")
		So(err, ShouldBeNil)
		table, err := db.Table("merge_b")
		So(err, ShouldBeNil)
		mTab, ok := table.(*catalog.MongoTable)
		So(ok, ShouldBeTrue)
		mTab.Pipeline[0] = bson.D{}

		actualPipelines = evaluator.GetNodePipeline(plan)
		So(actualPipelines, ShouldResemble, expectedPipelines)
	})
}

func BenchmarkAlgbrizeQuery(b *testing.B) {

	testSchema := evaluator.MustLoadSchema(testSchema4)
	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)
	testVariables := evaluator.CreateTestVariables(testInfo)
	testCatalog := evaluator.GetCatalogFromSchema(testSchema, testVariables)
	defaultDbName := "test"
	bench := func(name, sql string) {
		statement, err := parser.Parse(sql)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_, err := evaluator.AlgebrizeQuery(statement, defaultDbName, testVariables, testCatalog)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	bench("subquery", "select a, b from (select a, b from bar) b")
	bench("join", "select * from bar a join foo b on a.a=b.a and a.a=b.f")
	bench("subquery_join", "select * from (select foo.a from bar join (select foo.a from foo) foo on foo.a=bar.b) x join (select g.a from bar join (select foo.a from foo) g on g.a=bar.a) y on x.a=y.a")
}
