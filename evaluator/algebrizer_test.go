package evaluator

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	"github.com/shopspring/decimal"

	"strings"

	"github.com/kr/pretty"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAlgebrizeQuery(t *testing.T) {

	testSchema, err := schema.New(testSchema1)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}
	testInfo := getMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)
	testVars := createTestVariables(testInfo)
	testVars.MongoDBMaxVarcharLength = 10
	testCatalog := getCatalogFromSchema(testSchema, testVars)
	defaultDbName := "test"

	test := func(sql string, expectedPlanFactory func() PlanStage) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			actual, err := AlgebrizeQuery(statement, defaultDbName, testVars, testCatalog)
			So(err, ShouldBeNil)

			expected := expectedPlanFactory()

			if ShouldResemble(actual, expected) != "" {
				fmt.Printf("\nExpected: %# v", pretty.Formatter(expected))
				fmt.Printf("\nActual: %# v", pretty.Formatter(actual))
			}

			So(actual, ShouldResemble, expected)
		})
	}

	testVariables := func(sql string, vars *variable.Container, expectedPlanFactory func() PlanStage) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			actual, err := AlgebrizeQuery(statement, defaultDbName, vars, testCatalog)
			So(err, ShouldBeNil)

			expected := expectedPlanFactory()

			if ShouldResemble(actual, expected) != "" {
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

			actual, err := AlgebrizeQuery(statement, defaultDbName, testVars, testCatalog)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, message)
			So(actual, ShouldBeNil)
		})
	}

	createMongoSource := func(selectID int, tableName, aliasName string) PlanStage {
		db, _ := testCatalog.Database(defaultDbName)
		table, _ := db.Table(tableName)
		r := NewMongoSourceStage(db, table.(*catalog.MongoTable), selectID, aliasName)
		return r
	}

	Convey("Subject: AlgebrizeQuery", t, func() {
		Convey("Show Statements", func() {
			isDBName := "INFORMATION_SCHEMA"
			informationSchemaDB, _ := testCatalog.Database(isDBName)
			subqueryAliasName := "t"
			Convey("charset", func() {
				tbl, _ := informationSchemaDB.Table("CHARACTER_SETS")
				So(err, ShouldBeNil)
				source := NewDynamicSourceStage(tbl.(*catalog.DynamicTable), 2, "CHARACTER_SETS")
				subquery := NewSubquerySourceStage(
					NewProjectStage(
						source,
						createProjectedColumn(2, source, "CHARACTER_SETS", "CHARACTER_SET_NAME", "", "Charset"),
						createProjectedColumn(2, source, "CHARACTER_SETS", "DESCRIPTION", "", "Description"),
						createProjectedColumn(2, source, "CHARACTER_SETS", "DEFAULT_COLLATE_NAME", "", "Default collation"),
						createProjectedColumn(2, source, "CHARACTER_SETS", "MAXLEN", "", "Maxlen"),
					),
					2,
					subqueryAliasName,
				)
				test("show charset", func() PlanStage {
					return NewOrderByStage(
						subquery,
						&orderByTerm{
							expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "Charset"),
							ascending: true,
						},
					)
				})
				test("show charset like 'n'", func() PlanStage {
					return NewOrderByStage(
						NewFilterStage(
							subquery,
							&SQLLikeExpr{
								left:   createSQLColumnExprFromSource(subquery, subquery.aliasName, "Charset"),
								right:  SQLVarchar("n"),
								escape: SQLVarchar("\\"),
							},
						),
						&orderByTerm{
							expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "Charset"),
							ascending: true,
						},
					)
				})
				test("show charset where `Charset` = 'n'", func() PlanStage {
					return NewOrderByStage(
						NewFilterStage(
							subquery,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "Charset"),
								right: SQLVarchar("n"),
							},
						),
						&orderByTerm{
							expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "Charset"),
							ascending: true,
						},
					)
				})
			})
			Convey("collation", func() {
				tbl, _ := informationSchemaDB.Table("COLLATIONS")
				So(err, ShouldBeNil)
				source := NewDynamicSourceStage(tbl.(*catalog.DynamicTable), 2, "COLLATIONS")
				subquery := NewSubquerySourceStage(
					NewProjectStage(
						source,
						createProjectedColumn(2, source, "COLLATIONS", "COLLATION_NAME", "", "Collation"),
						createProjectedColumn(2, source, "COLLATIONS", "CHARACTER_SET_NAME", "", "Charset"),
						createProjectedColumn(2, source, "COLLATIONS", "ID", "", "Id"),
						createProjectedColumn(2, source, "COLLATIONS", "IS_DEFAULT", "", "Default"),
						createProjectedColumn(2, source, "COLLATIONS", "IS_COMPILED", "", "Compiled"),
						createProjectedColumn(2, source, "COLLATIONS", "SORTLEN", "", "Sortlen"),
					),
					2,
					subqueryAliasName,
				)
				test("show collation", func() PlanStage {
					return NewOrderByStage(
						subquery,
						&orderByTerm{
							expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "Collation"),
							ascending: true,
						},
					)
				})
				test("show collation like 'n'", func() PlanStage {
					return NewOrderByStage(
						NewFilterStage(
							subquery,
							&SQLLikeExpr{
								left:   createSQLColumnExprFromSource(subquery, subquery.aliasName, "Collation"),
								right:  SQLVarchar("n"),
								escape: SQLVarchar("\\"),
							},
						),
						&orderByTerm{
							expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "Collation"),
							ascending: true,
						},
					)
				})
				test("show collation where `Collation` = 'n'", func() PlanStage {
					return NewOrderByStage(
						NewFilterStage(
							subquery,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "Collation"),
								right: SQLVarchar("n"),
							},
						),
						&orderByTerm{
							expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "Collation"),
							ascending: true,
						},
					)
				})
			})
			Convey("columns", func() {
				tbl, _ := informationSchemaDB.Table("COLUMNS")
				So(err, ShouldBeNil)
				source := NewDynamicSourceStage(tbl.(*catalog.DynamicTable), 2, "COLUMNS")
				Convey("plain", func() {
					subquery := NewSubquerySourceStage(
						NewProjectStage(
							source,
							createProjectedColumn(2, source, "COLUMNS", "COLUMN_NAME", "", "Field"),
							createProjectedColumn(2, source, "COLUMNS", "COLUMN_TYPE", "", "Type"),
							createProjectedColumn(2, source, "COLUMNS", "IS_NULLABLE", "", "Null"),
							createProjectedColumn(2, source, "COLUMNS", "COLUMN_KEY", "", "Key"),
							createProjectedColumn(2, source, "COLUMNS", "COLUMN_DEFAULT", "", "Default"),
							createProjectedColumn(2, source, "COLUMNS", "EXTRA", "", "Extra"),
							createProjectedColumn(2, source, "COLUMNS", "TABLE_NAME", "", "TABLE_NAME"),
							createProjectedColumn(2, source, "COLUMNS", "TABLE_SCHEMA", "", "TABLE_SCHEMA"),
							createProjectedColumn(2, source, "COLUMNS", "ORDINAL_POSITION", "", "ORDINAL_POSITION"),
						),
						2,
						subqueryAliasName,
					)
					for _, from := range []string{"from foo", "from test.foo", "from foo from test", "in foo in test", "from foo in test", "in foo from test"} {
						test(fmt.Sprintf("show columns %s", from), func() PlanStage {
							return NewProjectStage(
								NewOrderByStage(
									NewFilterStage(
										subquery,
										&SQLAndExpr{
											left: &SQLEqualsExpr{
												left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_NAME"),
												right: SQLVarchar("foo"),
											},
											right: &SQLEqualsExpr{
												left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_SCHEMA"),
												right: SQLVarchar(defaultDbName),
											},
										},
									),
									&orderByTerm{
										expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "ORDINAL_POSITION"),
										ascending: true,
									},
								),
								createProjectedColumn(1, subquery, subquery.aliasName, "Field", "", "Field"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Type", "", "Type"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Null", "", "Null"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Key", "", "Key"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Default", "", "Default"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Extra", "", "Extra"),
							)
						})
						test(fmt.Sprintf("show columns %s like 'n'", from), func() PlanStage {
							return NewProjectStage(
								NewOrderByStage(
									NewFilterStage(
										subquery,
										&SQLAndExpr{
											left: &SQLAndExpr{
												left: &SQLEqualsExpr{
													left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_NAME"),
													right: SQLVarchar("foo"),
												},
												right: &SQLEqualsExpr{
													left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_SCHEMA"),
													right: SQLVarchar(defaultDbName),
												},
											},
											right: &SQLLikeExpr{
												left:   createSQLColumnExprFromSource(subquery, subquery.aliasName, "Field"),
												right:  SQLVarchar("n"),
												escape: SQLVarchar("\\"),
											},
										},
									),
									&orderByTerm{
										expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "ORDINAL_POSITION"),
										ascending: true,
									},
								),
								createProjectedColumn(1, subquery, subquery.aliasName, "Field", "", "Field"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Type", "", "Type"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Null", "", "Null"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Key", "", "Key"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Default", "", "Default"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Extra", "", "Extra"),
							)
						})
						test(fmt.Sprintf("show columns %s where `Field` = 'n'", from), func() PlanStage {
							return NewProjectStage(
								NewOrderByStage(
									NewFilterStage(
										subquery,
										&SQLAndExpr{
											left: &SQLAndExpr{
												left: &SQLEqualsExpr{
													left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_NAME"),
													right: SQLVarchar("foo"),
												},
												right: &SQLEqualsExpr{
													left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_SCHEMA"),
													right: SQLVarchar(defaultDbName),
												},
											},
											right: &SQLEqualsExpr{
												left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "Field"),
												right: SQLVarchar("n"),
											},
										},
									),
									&orderByTerm{
										expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "ORDINAL_POSITION"),
										ascending: true,
									},
								),
								createProjectedColumn(1, subquery, subquery.aliasName, "Field", "", "Field"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Type", "", "Type"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Null", "", "Null"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Key", "", "Key"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Default", "", "Default"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Extra", "", "Extra"),
							)
						})
					}
				})
				Convey("full", func() {
					subquery := NewSubquerySourceStage(
						NewProjectStage(
							source,
							createProjectedColumn(2, source, "COLUMNS", "COLUMN_NAME", "", "Field"),
							createProjectedColumn(2, source, "COLUMNS", "COLUMN_TYPE", "", "Type"),
							createProjectedColumn(2, source, "COLUMNS", "COLLATION_NAME", "", "Collation"),
							createProjectedColumn(2, source, "COLUMNS", "IS_NULLABLE", "", "Null"),
							createProjectedColumn(2, source, "COLUMNS", "COLUMN_KEY", "", "Key"),
							createProjectedColumn(2, source, "COLUMNS", "COLUMN_DEFAULT", "", "Default"),
							createProjectedColumn(2, source, "COLUMNS", "EXTRA", "", "Extra"),
							createProjectedColumn(2, source, "COLUMNS", "PRIVILEGES", "", "Privileges"),
							createProjectedColumn(2, source, "COLUMNS", "COLUMN_COMMENT", "", "Comment"),
							createProjectedColumn(2, source, "COLUMNS", "TABLE_NAME", "", "TABLE_NAME"),
							createProjectedColumn(2, source, "COLUMNS", "TABLE_SCHEMA", "", "TABLE_SCHEMA"),
							createProjectedColumn(2, source, "COLUMNS", "ORDINAL_POSITION", "", "ORDINAL_POSITION"),
						),
						2,
						subqueryAliasName,
					)
					for _, from := range []string{"from foo", "from test.foo", "from foo from test", "in foo in test", "from foo in test", "in foo from test"} {
						test(fmt.Sprintf("show full columns %s", from), func() PlanStage {
							return NewProjectStage(
								NewOrderByStage(
									NewFilterStage(
										subquery,
										&SQLAndExpr{
											left: &SQLEqualsExpr{
												left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_NAME"),
												right: SQLVarchar("foo"),
											},
											right: &SQLEqualsExpr{
												left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_SCHEMA"),
												right: SQLVarchar(defaultDbName),
											},
										},
									),
									&orderByTerm{
										expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "ORDINAL_POSITION"),
										ascending: true,
									},
								),
								createProjectedColumn(1, subquery, subquery.aliasName, "Field", "", "Field"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Type", "", "Type"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Collation", "", "Collation"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Null", "", "Null"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Key", "", "Key"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Default", "", "Default"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Extra", "", "Extra"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Privileges", "", "Privileges"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Comment", "", "Comment"),
							)
						})
						test(fmt.Sprintf("show full columns %s like 'n'", from), func() PlanStage {
							return NewProjectStage(
								NewOrderByStage(
									NewFilterStage(
										subquery,
										&SQLAndExpr{
											left: &SQLAndExpr{
												left: &SQLEqualsExpr{
													left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_NAME"),
													right: SQLVarchar("foo"),
												},
												right: &SQLEqualsExpr{
													left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_SCHEMA"),
													right: SQLVarchar(defaultDbName),
												},
											},
											right: &SQLLikeExpr{
												left:   createSQLColumnExprFromSource(subquery, subquery.aliasName, "Field"),
												right:  SQLVarchar("n"),
												escape: SQLVarchar("\\"),
											},
										},
									),
									&orderByTerm{
										expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "ORDINAL_POSITION"),
										ascending: true,
									},
								),
								createProjectedColumn(1, subquery, subquery.aliasName, "Field", "", "Field"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Type", "", "Type"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Collation", "", "Collation"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Null", "", "Null"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Key", "", "Key"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Default", "", "Default"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Extra", "", "Extra"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Privileges", "", "Privileges"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Comment", "", "Comment"),
							)
						})
						test(fmt.Sprintf("show full columns %s where `Field` = 'n'", from), func() PlanStage {
							return NewProjectStage(
								NewOrderByStage(
									NewFilterStage(
										subquery,
										&SQLAndExpr{
											left: &SQLAndExpr{
												left: &SQLEqualsExpr{
													left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_NAME"),
													right: SQLVarchar("foo"),
												},
												right: &SQLEqualsExpr{
													left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_SCHEMA"),
													right: SQLVarchar(defaultDbName),
												},
											},
											right: &SQLEqualsExpr{
												left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "Field"),
												right: SQLVarchar("n"),
											},
										},
									),
									&orderByTerm{
										expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "ORDINAL_POSITION"),
										ascending: true,
									},
								),
								createProjectedColumn(1, subquery, subquery.aliasName, "Field", "", "Field"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Type", "", "Type"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Collation", "", "Collation"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Null", "", "Null"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Key", "", "Key"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Default", "", "Default"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Extra", "", "Extra"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Privileges", "", "Privileges"),
								createProjectedColumn(1, subquery, subquery.aliasName, "Comment", "", "Comment"),
							)
						})
					}
				})
			})
			Convey("create table", func() {
				testDB, _ := testCatalog.Database("test")
				tbl, _ := testDB.Table("foo")
				createTableSQL := catalog.GenerateCreateTable(tbl, 10)
				test("show create table foo", func() PlanStage {
					return NewProjectStage(
						NewDualStage(),
						ProjectedColumn{
							Column: &Column{
								SelectID: 1,
								Name:     "Table",
								SQLType:  schema.SQLVarchar,
							},
							Expr: SQLVarchar(string(tbl.Name())),
						},
						ProjectedColumn{
							Column: &Column{
								SelectID: 1,
								Name:     "Create Table",
								SQLType:  schema.SQLVarchar,
							},
							Expr: SQLVarchar(createTableSQL),
						},
					)
				})
				test("show create table .foo", func() PlanStage {
					return NewProjectStage(
						NewDualStage(),
						ProjectedColumn{
							Column: &Column{
								SelectID: 1,
								Name:     "Table",
								SQLType:  schema.SQLVarchar,
							},
							Expr: SQLVarchar(string(tbl.Name())),
						},
						ProjectedColumn{
							Column: &Column{
								SelectID: 1,
								Name:     "Create Table",
								SQLType:  schema.SQLVarchar,
							},
							Expr: SQLVarchar(createTableSQL),
						},
					)
				})
				test("show create table test.foo", func() PlanStage {
					return NewProjectStage(
						NewDualStage(),
						ProjectedColumn{
							Column: &Column{
								SelectID: 1,
								Name:     "Table",
								SQLType:  schema.SQLVarchar,
							},
							Expr: SQLVarchar(string(tbl.Name())),
						},
						ProjectedColumn{
							Column: &Column{
								SelectID: 1,
								Name:     "Create Table",
								SQLType:  schema.SQLVarchar,
							},
							Expr: SQLVarchar(createTableSQL),
						},
					)
				})
			})
			Convey("databases/schemas", func() {
				tbl, _ := informationSchemaDB.Table("SCHEMATA")
				So(err, ShouldBeNil)
				source := NewDynamicSourceStage(tbl.(*catalog.DynamicTable), 2, "SCHEMATA")
				subquery := NewSubquerySourceStage(
					NewProjectStage(
						source,
						createProjectedColumn(2, source, "SCHEMATA", "SCHEMA_NAME", "", "Database"),
					),
					2,
					subqueryAliasName,
				)
				for _, name := range []string{"databases", "schemas"} {
					test(fmt.Sprintf("show %s", name), func() PlanStage {
						return NewOrderByStage(
							subquery,
							&orderByTerm{
								expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "Database"),
								ascending: true,
							},
						)
					})
					test(fmt.Sprintf("show %s like 'n'", name), func() PlanStage {
						return NewOrderByStage(
							NewFilterStage(
								subquery,
								&SQLLikeExpr{
									left:   createSQLColumnExprFromSource(subquery, subquery.aliasName, "Database"),
									right:  SQLVarchar("n"),
									escape: SQLVarchar("\\"),
								},
							),
							&orderByTerm{
								expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "Database"),
								ascending: true,
							},
						)
					})
					test(fmt.Sprintf("show %s where `Database` = 'n'", name), func() PlanStage {
						return NewOrderByStage(
							NewFilterStage(
								subquery,
								&SQLEqualsExpr{
									left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "Database"),
									right: SQLVarchar("n"),
								},
							),
							&orderByTerm{
								expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "Database"),
								ascending: true,
							},
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
						tbl, _ := informationSchemaDB.Table(fmt.Sprintf("%s_%s", actualScope, kind))
						actualTableName := string(tbl.Name())
						So(err, ShouldBeNil)
						source := NewDynamicSourceStage(tbl.(*catalog.DynamicTable), 2, actualTableName)
						subquery := NewSubquerySourceStage(
							NewProjectStage(
								source,
								createProjectedColumn(2, source, actualTableName, "VARIABLE_NAME", "", "Variable_name"),
								createProjectedColumn(2, source, actualTableName, "VARIABLE_VALUE", "", "Value"),
							),
							2,
							subqueryAliasName,
						)
						showName := strings.TrimSpace(scope + " " + kind)
						test(fmt.Sprintf("show %s", showName), func() PlanStage {
							return NewOrderByStage(
								subquery,
								&orderByTerm{
									expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "Variable_name"),
									ascending: true,
								},
							)
						})
						test(fmt.Sprintf("show %s like 'n'", showName), func() PlanStage {
							return NewOrderByStage(
								NewFilterStage(
									subquery,
									&SQLLikeExpr{
										left:   createSQLColumnExprFromSource(subquery, subquery.aliasName, "Variable_name"),
										right:  SQLVarchar("n"),
										escape: SQLVarchar("\\"),
									},
								),
								&orderByTerm{
									expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "Variable_name"),
									ascending: true,
								},
							)
						})
						test(fmt.Sprintf("show %s where Variable_name = 'n'", showName), func() PlanStage {
							return NewOrderByStage(
								NewFilterStage(
									subquery,
									&SQLEqualsExpr{
										left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "Variable_name"),
										right: SQLVarchar("n"),
									},
								),
								&orderByTerm{
									expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, "Variable_name"),
									ascending: true,
								},
							)
						})
					}
				}
			})
			Convey("tables", func() {
				tbl, _ := informationSchemaDB.Table("TABLES")
				So(err, ShouldBeNil)
				source := NewDynamicSourceStage(tbl.(*catalog.DynamicTable), 2, "TABLES")
				columnName := "Tables_in_" + defaultDbName
				Convey("plain", func() {
					subquery := NewSubquerySourceStage(
						NewProjectStage(
							source,
							createProjectedColumn(2, source, "TABLES", "TABLE_NAME", "", columnName),
							createProjectedColumn(2, source, "TABLES", "TABLE_SCHEMA", "", "TABLE_SCHEMA"),
						),
						2,
						subqueryAliasName,
					)

					for _, from := range []string{"", " from test", " in test"} {
						test(fmt.Sprintf("show tables%s", from), func() PlanStage {
							return NewProjectStage(
								NewOrderByStage(
									NewFilterStage(
										subquery,
										&SQLEqualsExpr{
											left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_SCHEMA"),
											right: SQLVarchar("test"),
										},
									),
									&orderByTerm{
										expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, columnName),
										ascending: true,
									},
								),
								createProjectedColumn(1, subquery, subquery.aliasName, columnName, "", columnName),
							)
						})
						test(fmt.Sprintf("show tables%s like 'n'", from), func() PlanStage {
							return NewProjectStage(
								NewOrderByStage(
									NewFilterStage(
										subquery,
										&SQLAndExpr{
											left: &SQLEqualsExpr{
												left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_SCHEMA"),
												right: SQLVarchar("test"),
											},
											right: &SQLLikeExpr{
												left:   createSQLColumnExprFromSource(subquery, subquery.aliasName, columnName),
												right:  SQLVarchar("n"),
												escape: SQLVarchar("\\"),
											},
										},
									),
									&orderByTerm{
										expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, columnName),
										ascending: true,
									},
								),
								createProjectedColumn(1, subquery, subquery.aliasName, columnName, "", columnName),
							)
						})
						test(fmt.Sprintf("show tables%s where `%s` = 'n'", from, columnName), func() PlanStage {
							return NewProjectStage(
								NewOrderByStage(
									NewFilterStage(
										subquery,
										&SQLAndExpr{
											left: &SQLEqualsExpr{
												left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_SCHEMA"),
												right: SQLVarchar("test"),
											},
											right: &SQLEqualsExpr{
												left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, columnName),
												right: SQLVarchar("n"),
											},
										},
									),
									&orderByTerm{
										expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, columnName),
										ascending: true,
									},
								),
								createProjectedColumn(1, subquery, subquery.aliasName, columnName, "", columnName),
							)
						})
					}
				})
				Convey("full", func() {
					subquery := NewSubquerySourceStage(
						NewProjectStage(
							source,
							createProjectedColumn(2, source, "TABLES", "TABLE_NAME", "", columnName),
							createProjectedColumn(2, source, "TABLES", "TABLE_TYPE", "", "Type"),
							createProjectedColumn(2, source, "TABLES", "TABLE_SCHEMA", "", "TABLE_SCHEMA"),
						),
						2,
						subqueryAliasName,
					)

					for _, from := range []string{"", " from test", " in test"} {
						test(fmt.Sprintf("show full tables%s", from), func() PlanStage {
							return NewProjectStage(
								NewOrderByStage(
									NewFilterStage(
										subquery,
										&SQLEqualsExpr{
											left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_SCHEMA"),
											right: SQLVarchar("test"),
										},
									),
									&orderByTerm{
										expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, columnName),
										ascending: true,
									},
								),
								createProjectedColumn(1, subquery, subquery.aliasName, columnName, "", columnName),
								createProjectedColumn(1, subquery, subquery.aliasName, "Type", "", "Type"),
							)
						})
						test(fmt.Sprintf("show full tables%s like 'n'", from), func() PlanStage {
							return NewProjectStage(
								NewOrderByStage(
									NewFilterStage(
										subquery,
										&SQLAndExpr{
											left: &SQLEqualsExpr{
												left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_SCHEMA"),
												right: SQLVarchar("test"),
											},
											right: &SQLLikeExpr{
												left:   createSQLColumnExprFromSource(subquery, subquery.aliasName, columnName),
												right:  SQLVarchar("n"),
												escape: SQLVarchar("\\"),
											},
										},
									),
									&orderByTerm{
										expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, columnName),
										ascending: true,
									},
								),
								createProjectedColumn(1, subquery, subquery.aliasName, columnName, "", columnName),
								createProjectedColumn(1, subquery, subquery.aliasName, "Type", "", "Type"),
							)
						})
						test(fmt.Sprintf("show full tables%s where `%s` = 'n'", from, columnName), func() PlanStage {
							return NewProjectStage(
								NewOrderByStage(
									NewFilterStage(
										subquery,
										&SQLAndExpr{
											left: &SQLEqualsExpr{
												left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, "TABLE_SCHEMA"),
												right: SQLVarchar("test"),
											},
											right: &SQLEqualsExpr{
												left:  createSQLColumnExprFromSource(subquery, subquery.aliasName, columnName),
												right: SQLVarchar("n"),
											},
										},
									),
									&orderByTerm{
										expr:      createSQLColumnExprFromSource(subquery, subquery.aliasName, columnName),
										ascending: true,
									},
								),
								createProjectedColumn(1, subquery, subquery.aliasName, columnName, "", columnName),
								createProjectedColumn(1, subquery, subquery.aliasName, "Type", "", "Type"),
							)
						})
					}
				})
			})
		})

		Convey("Select Statements", func() {
			Convey("dual queries", func() {
				test("select 2 + 3", func() PlanStage {
					return NewProjectStage(
						NewDualStage(),
						createProjectedColumnFromSQLExpr(1, "", "2+3", &SQLAddExpr{
							left:  SQLInt(2),
							right: SQLInt(3),
						}),
					)
				})

				test("select false", func() PlanStage {
					return NewProjectStage(
						NewDualStage(),
						createProjectedColumnFromSQLExpr(1, "", "false", SQLFalse),
					)
				})

				test("select true", func() PlanStage {
					return NewProjectStage(
						NewDualStage(),
						createProjectedColumnFromSQLExpr(1, "", "true", SQLTrue),
					)
				})

				test("select 2 + 3 from dual", func() PlanStage {
					return NewProjectStage(
						NewDualStage(),
						createProjectedColumnFromSQLExpr(1, "", "2+3", &SQLAddExpr{
							left:  SQLInt(2),
							right: SQLInt(3),
						}),
					)
				})
			})

			Convey("from", func() {
				Convey("subqueries", func() {
					test("select a from (select a from foo) f", func() PlanStage {
						source := createMongoSource(2, "foo", "foo")
						subquery := NewSubquerySourceStage(NewProjectStage(source, createProjectedColumn(2, source, "foo", "a", "", "a")), 2, "f")
						return NewProjectStage(subquery, createProjectedColumn(1, subquery, "f", "a", "", "a"))
					})

					test("select f.a from (select a from foo) f", func() PlanStage {
						source := createMongoSource(2, "foo", "foo")
						subquery := NewSubquerySourceStage(NewProjectStage(source, createProjectedColumn(2, source, "foo", "a", "", "a")), 2, "f")
						return NewProjectStage(subquery, createProjectedColumn(1, subquery, "f", "a", "", "a"))
					})

					test("select f.a from (select test.a from foo test) f", func() PlanStage {
						source := createMongoSource(2, "foo", "test")
						subquery := NewSubquerySourceStage(NewProjectStage(source, createProjectedColumn(2, source, "test", "a", "", "a")), 2, "f")
						return NewProjectStage(subquery, createProjectedColumn(1, subquery, "f", "a", "", "a"))
					})

					testVariables("select g.a from (select a from foo) g",
						&variable.Container{
							SQLSelectLimit: 5,
							MongoDBInfo:    testInfo,
						},
						func() PlanStage {
							source := createMongoSource(2, "foo", "foo")
							subquery := NewSubquerySourceStage(NewProjectStage(source, createProjectedColumn(2, source, "foo", "a", "", "a")), 2, "g")
							return NewLimitStage(
								NewProjectStage(subquery, createProjectedColumn(1, subquery, "g", "a", "", "a")),
								0,
								5,
							)
						})
				})

				Convey("joins", func() {
					test("select foo.a, bar.a from foo, bar", func() PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := NewJoinStage(CrossJoin, fooSource, barSource, SQLTrue)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "foo", "a", "", "a"),
							createProjectedColumn(1, join, "bar", "a", "", "a"),
						)
					})

					test("select f.a, bar.a from foo f, bar", func() PlanStage {
						fooSource := createMongoSource(1, "foo", "f")
						barSource := createMongoSource(1, "bar", "bar")
						join := NewJoinStage(CrossJoin, fooSource, barSource, SQLTrue)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "f", "a", "", "a"),
							createProjectedColumn(1, join, "bar", "a", "", "a"),
						)
					})

					test("select f.a, b.a from foo f, bar b", func() PlanStage {
						fooSource := createMongoSource(1, "foo", "f")
						barSource := createMongoSource(1, "bar", "b")
						join := NewJoinStage(CrossJoin, fooSource, barSource, SQLTrue)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "f", "a", "", "a"),
							createProjectedColumn(1, join, "b", "a", "", "a"),
						)
					})

					test("select foo.a, bar.a from foo inner join bar on foo.b = bar.b", func() PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := NewJoinStage(InnerJoin, fooSource, barSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(fooSource, "foo", "b"),
								right: createSQLColumnExprFromSource(barSource, "bar", "b"),
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "foo", "a", "", "a"),
							createProjectedColumn(1, join, "bar", "a", "", "a"),
						)
					})

					test("select foo.a, bar.a from foo join bar on foo.b = bar.b", func() PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := NewJoinStage(InnerJoin, fooSource, barSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(fooSource, "foo", "b"),
								right: createSQLColumnExprFromSource(barSource, "bar", "b"),
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "foo", "a", "", "a"),
							createProjectedColumn(1, join, "bar", "a", "", "a"),
						)
					})

					test("select foo.a, bar.a from foo left outer join bar on foo.b = bar.b", func() PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := NewJoinStage(LeftJoin, fooSource, barSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(fooSource, "foo", "b"),
								right: createSQLColumnExprFromSource(barSource, "bar", "b"),
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "foo", "a", "", "a"),
							createProjectedColumn(1, join, "bar", "a", "", "a"),
						)
					})

					test("select foo.a, bar.a from foo right outer join bar on foo.b = bar.b", func() PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := NewJoinStage(RightJoin, fooSource, barSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(fooSource, "foo", "b"),
								right: createSQLColumnExprFromSource(barSource, "bar", "b"),
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "foo", "a", "", "a"),
							createProjectedColumn(1, join, "bar", "a", "", "a"),
						)
					})
					test("select foo.a, bar.a from foo straight_join bar on foo.b = bar.b", func() PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := NewJoinStage(StraightJoin, fooSource, barSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(fooSource, "foo", "b"),
								right: createSQLColumnExprFromSource(barSource, "bar", "b"),
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "foo", "a", "", "a"),
							createProjectedColumn(1, join, "bar", "a", "", "a"),
						)
					})
					test("select foo.a, bar.a from foo join bar on foo.a = bar.a and foo.e = bar.d join baz on baz.b = bar.b", func() PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						firstJoin := NewJoinStage(InnerJoin, fooSource, barSource,
							&SQLAndExpr{
								left: &SQLEqualsExpr{
									left:  createSQLColumnExprFromSource(fooSource, "foo", "a"),
									right: createSQLColumnExprFromSource(barSource, "bar", "a"),
								},
								right: &SQLEqualsExpr{
									left:  createSQLColumnExprFromSource(fooSource, "foo", "e"),
									right: createSQLColumnExprFromSource(barSource, "bar", "d"),
								},
							},
						)
						secondJoin := NewJoinStage(InnerJoin, firstJoin, bazSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(bazSource, "baz", "b"),
								right: createSQLColumnExprFromSource(barSource, "bar", "b"),
							},
						)
						return NewProjectStage(secondJoin,
							createProjectedColumn(1, secondJoin, "foo", "a", "", "a"),
							createProjectedColumn(1, secondJoin, "bar", "a", "", "a"),
						)
					})
					test("select bar.a, baz.b from bar join baz using (b)", func() PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := NewJoinStage(InnerJoin, barSource, bazSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(barSource, "bar", "b"),
								right: createSQLColumnExprFromSource(bazSource, "baz", "b"),
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "", "a"),
							createProjectedColumn(1, join, "baz", "b", "", "b"),
						)
					})
					test("select bar.a, baz.b from bar join baz using (a, b)", func() PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := NewJoinStage(InnerJoin, barSource, bazSource,
							&SQLAndExpr{
								left: &SQLEqualsExpr{
									left:  createSQLColumnExprFromSource(barSource, "bar", "a"),
									right: createSQLColumnExprFromSource(bazSource, "baz", "a"),
								},
								right: &SQLEqualsExpr{
									left:  createSQLColumnExprFromSource(barSource, "bar", "b"),
									right: createSQLColumnExprFromSource(bazSource, "baz", "b"),
								},
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "", "a"),
							createProjectedColumn(1, join, "baz", "b", "", "b"),
						)
					})
					test("select bar.a, buzz.d, foo.c from bar join buzz join foo using (a, c)", func() PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						buzzSource := createMongoSource(1, "buzz", "buzz")
						fooSource := createMongoSource(1, "foo", "foo")
						firstJoin := NewJoinStage(CrossJoin, barSource, buzzSource, SQLBool(1))
						secondJoin := NewJoinStage(InnerJoin, firstJoin, fooSource,
							&SQLAndExpr{
								left: &SQLEqualsExpr{
									left:  createSQLColumnExprFromSource(barSource, "bar", "a"),
									right: createSQLColumnExprFromSource(fooSource, "foo", "a"),
								},
								right: &SQLEqualsExpr{
									left:  createSQLColumnExprFromSource(buzzSource, "buzz", "c"),
									right: createSQLColumnExprFromSource(fooSource, "foo", "c"),
								},
							},
						)
						return NewProjectStage(secondJoin,
							createProjectedColumn(1, secondJoin, "bar", "a", "", "a"),
							createProjectedColumn(1, secondJoin, "buzz", "d", "", "d"),
							createProjectedColumn(1, secondJoin, "foo", "c", "", "c"),
						)
					})
					test("select bar.a, buzz.d, foo.c from bar join foo using (a) join buzz using (c)", func() PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						fooSource := createMongoSource(1, "foo", "foo")
						buzzSource := createMongoSource(1, "buzz", "buzz")
						firstJoin := NewJoinStage(InnerJoin, barSource, fooSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(barSource, "bar", "a"),
								right: createSQLColumnExprFromSource(fooSource, "foo", "a"),
							},
						)
						secondJoin := NewJoinStage(InnerJoin, firstJoin, buzzSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(fooSource, "foo", "c"),
								right: createSQLColumnExprFromSource(buzzSource, "buzz", "c"),
							},
						)
						return NewProjectStage(secondJoin,
							createProjectedColumn(1, secondJoin, "bar", "a", "", "a"),
							createProjectedColumn(1, secondJoin, "buzz", "d", "", "d"),
							createProjectedColumn(1, secondJoin, "foo", "c", "", "c"),
						)
					})
					test("select bar.a, baz.b from bar join baz using (a, a, a, a, b)", func() PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := NewJoinStage(InnerJoin, barSource, bazSource,
							&SQLAndExpr{
								left: &SQLEqualsExpr{
									left:  createSQLColumnExprFromSource(barSource, "bar", "a"),
									right: createSQLColumnExprFromSource(bazSource, "baz", "a"),
								},
								right: &SQLEqualsExpr{
									left:  createSQLColumnExprFromSource(barSource, "bar", "b"),
									right: createSQLColumnExprFromSource(bazSource, "baz", "b"),
								},
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "", "a"),
							createProjectedColumn(1, join, "baz", "b", "", "b"),
						)
					})
					test("select bar.a, baz.b from bar join baz using (a, b, b, b, b)", func() PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := NewJoinStage(InnerJoin, barSource, bazSource,
							&SQLAndExpr{
								left: &SQLEqualsExpr{
									left:  createSQLColumnExprFromSource(barSource, "bar", "a"),
									right: createSQLColumnExprFromSource(bazSource, "baz", "a"),
								},
								right: &SQLEqualsExpr{
									left:  createSQLColumnExprFromSource(barSource, "bar", "b"),
									right: createSQLColumnExprFromSource(bazSource, "baz", "b"),
								},
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "", "a"),
							createProjectedColumn(1, join, "baz", "b", "", "b"),
						)
					})
					test("select bar.a, baz.b from bar cross join baz using (b)", func() PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := NewJoinStage(InnerJoin, barSource, bazSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(barSource, "bar", "b"),
								right: createSQLColumnExprFromSource(bazSource, "baz", "b"),
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "", "a"),
							createProjectedColumn(1, join, "baz", "b", "", "b"),
						)
					})
					test("select bar.a, baz.b from bar inner join baz", func() PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := NewJoinStage(CrossJoin, barSource, bazSource, SQLBool(1))
						return NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "", "a"),
							createProjectedColumn(1, join, "baz", "b", "", "b"),
						)
					})
					test("select bar.a, biz.b from bar join (select baz.b, foo.c from baz join foo on baz.a = foo.a) as biz using (b)", func() PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(2, "baz", "baz")
						fooSource := createMongoSource(2, "foo", "foo")
						subJoin := NewJoinStage(InnerJoin, bazSource, fooSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(bazSource, "baz", "a"),
								right: createSQLColumnExprFromSource(fooSource, "foo", "a"),
							},
						)
						bizSource := NewSubquerySourceStage(
							NewProjectStage(subJoin,
								createProjectedColumn(2, subJoin, "baz", "b", "", "b"),
								createProjectedColumn(2, subJoin, "foo", "c", "", "c"),
							), 2, "biz")
						join := NewJoinStage(InnerJoin, barSource, bizSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(barSource, "bar", "b"),
								right: createSQLColumnExprFromSource(bizSource, "biz", "b"),
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "", "a"),
							createProjectedColumn(1, join, "biz", "b", "", "b"),
						)
					})
					test("select bar.a, biz.b from (select baz.b, foo.c from baz join foo on baz.a = foo.a) as biz join bar using (b)", func() PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(2, "baz", "baz")
						fooSource := createMongoSource(2, "foo", "foo")
						subJoin := NewJoinStage(InnerJoin, bazSource, fooSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(bazSource, "baz", "a"),
								right: createSQLColumnExprFromSource(fooSource, "foo", "a"),
							},
						)
						bizSource := NewSubquerySourceStage(
							NewProjectStage(subJoin,
								createProjectedColumn(2, subJoin, "baz", "b", "", "b"),
								createProjectedColumn(2, subJoin, "foo", "c", "", "c"),
							), 2, "biz")
						join := NewJoinStage(InnerJoin, bizSource, barSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(bizSource, "biz", "b"),
								right: createSQLColumnExprFromSource(barSource, "bar", "b"),
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "", "a"),
							createProjectedColumn(1, join, "biz", "b", "", "b"),
						)
					})
					test("select fiz.b from (select bar.b from bar) as biz join (select foo.b from foo) as fiz using (b)", func() PlanStage {
						barSource := createMongoSource(2, "bar", "bar")
						fooSource := createMongoSource(3, "foo", "foo")
						bizSource := NewSubquerySourceStage(NewProjectStage(barSource, createProjectedColumn(2, barSource, "bar", "b", "", "b")), 2, "biz")
						fizSource := NewSubquerySourceStage(NewProjectStage(fooSource, createProjectedColumn(3, fooSource, "foo", "b", "", "b")), 3, "fiz")
						join := NewJoinStage(InnerJoin, bizSource, fizSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(bizSource, "biz", "b"),
								right: createSQLColumnExprFromSource(fizSource, "fiz", "b"),
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "fiz", "b", "", "b"))
					})
					test("select * from bar join baz using (b)", func() PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := NewJoinStage(InnerJoin, barSource, bazSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(barSource, "bar", "b"),
								right: createSQLColumnExprFromSource(bazSource, "baz", "b"),
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "b", "", "b"),
							createProjectedColumn(1, join, "bar", "a", "", "a"),
							createProjectedColumn(1, join, "bar", "d", "", "d"),
							createProjectedColumn(1, join, "bar", "_id", "", "_id"),
							createProjectedColumn(1, join, "baz", "a", "", "a"),
							createProjectedColumn(1, join, "baz", "_id", "", "_id"))
					})
					test("select * from bar join baz using (_id, b)", func() PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := NewJoinStage(InnerJoin, barSource, bazSource,
							&SQLAndExpr{
								left: &SQLEqualsExpr{
									left:  createSQLColumnExprFromSource(barSource, "bar", "_id"),
									right: createSQLColumnExprFromSource(bazSource, "baz", "_id"),
								},
								right: &SQLEqualsExpr{
									left:  createSQLColumnExprFromSource(barSource, "bar", "b"),
									right: createSQLColumnExprFromSource(bazSource, "baz", "b"),
								},
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "b", "", "b"),
							createProjectedColumn(1, join, "bar", "_id", "", "_id"),
							createProjectedColumn(1, join, "bar", "a", "", "a"),
							createProjectedColumn(1, join, "bar", "d", "", "d"),
							createProjectedColumn(1, join, "baz", "a", "", "a"))
					})
					test("select * from bar right join baz using (b)", func() PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := NewJoinStage(RightJoin, barSource, bazSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(barSource, "bar", "b"),
								right: createSQLColumnExprFromSource(bazSource, "baz", "b"),
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "baz", "b", "", "b"),
							createProjectedColumn(1, join, "baz", "a", "", "a"),
							createProjectedColumn(1, join, "baz", "_id", "", "_id"),
							createProjectedColumn(1, join, "bar", "a", "", "a"),
							createProjectedColumn(1, join, "bar", "d", "", "d"),
							createProjectedColumn(1, join, "bar", "_id", "", "_id"))
					})
					test("select bar.*, baz.* from bar join baz using (b)", func() PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := NewJoinStage(InnerJoin, barSource, bazSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(barSource, "bar", "b"),
								right: createSQLColumnExprFromSource(bazSource, "baz", "b"),
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "a", "", "a"),
							createProjectedColumn(1, join, "bar", "b", "", "b"),
							createProjectedColumn(1, join, "bar", "d", "", "d"),
							createProjectedColumn(1, join, "bar", "_id", "", "_id"),
							createProjectedColumn(1, join, "baz", "a", "", "a"),
							createProjectedColumn(1, join, "baz", "b", "", "b"),
							createProjectedColumn(1, join, "baz", "_id", "", "_id"))
					})
					test("select bar.b, baz.b from bar join baz using (b)", func() PlanStage {
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join := NewJoinStage(InnerJoin, barSource, bazSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(barSource, "bar", "b"),
								right: createSQLColumnExprFromSource(bazSource, "baz", "b"),
							},
						)
						return NewProjectStage(join,
							createProjectedColumn(1, join, "bar", "b", "", "b"),
							createProjectedColumn(1, join, "baz", "b", "", "b"))
					})
					test("select * from buzz join (baz join bar using (_id)) using (d)", func() PlanStage {
						buzzSource := createMongoSource(1, "buzz", "buzz")
						barSource := createMongoSource(1, "bar", "bar")
						bazSource := createMongoSource(1, "baz", "baz")
						join1 := NewJoinStage(InnerJoin, bazSource, barSource,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(bazSource, "baz", "_id"),
								right: createSQLColumnExprFromSource(barSource, "bar", "_id"),
							},
						)
						join2 := NewJoinStage(InnerJoin, buzzSource, join1,
							&SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(buzzSource, "buzz", "d"),
								right: createSQLColumnExprFromSource(barSource, "bar", "d"),
							},
						)
						return NewProjectStage(join2,
							createProjectedColumn(1, buzzSource, "buzz", "d", "", "d"),
							createProjectedColumn(1, buzzSource, "buzz", "c", "", "c"),
							createProjectedColumn(1, buzzSource, "buzz", "_id", "", "_id"),
							createProjectedColumn(1, bazSource, "baz", "_id", "", "_id"),
							createProjectedColumn(1, bazSource, "baz", "a", "", "a"),
							createProjectedColumn(1, bazSource, "baz", "b", "", "b"),
							createProjectedColumn(1, barSource, "bar", "a", "", "a"),
							createProjectedColumn(1, barSource, "bar", "b", "", "b"),
						)
					})
				})
			})

			Convey("select", func() {
				Convey("star simple queries", func() {
					test("select * from foo", func() PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return NewProjectStage(source, createAllProjectedColumnsFromSource(1, source, "")...)
					})

					test("select foo.* from foo", func() PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return NewProjectStage(source, createAllProjectedColumnsFromSource(1, source, "")...)
					})

					test("select f.* from foo f", func() PlanStage {
						source := createMongoSource(1, "foo", "f")
						return NewProjectStage(source, createAllProjectedColumnsFromSource(1, source, "")...)
					})

					test("select a, foo.* from foo", func() PlanStage {
						source := createMongoSource(1, "foo", "foo")
						columns := append(
							ProjectedColumns{createProjectedColumn(1, source, "foo", "a", "", "a")},
							createAllProjectedColumnsFromSource(1, source, "")...)
						return NewProjectStage(source, columns...)
					})

					test("select foo.*, a from foo", func() PlanStage {
						source := createMongoSource(1, "foo", "foo")
						columns := append(
							createAllProjectedColumnsFromSource(1, source, ""),
							createProjectedColumn(1, source, "foo", "a", "", "a"))
						return NewProjectStage(source, columns...)
					})

					test("select a, f.* from foo f", func() PlanStage {
						source := createMongoSource(1, "foo", "f")
						columns := append(
							ProjectedColumns{createProjectedColumn(1, source, "f", "a", "", "a")},
							createAllProjectedColumnsFromSource(1, source, "")...)
						return NewProjectStage(source, columns...)
					})

					test("select * from foo, bar", func() PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := NewJoinStage(CrossJoin, fooSource, barSource, SQLTrue)
						return NewProjectStage(join, createAllProjectedColumnsFromSource(1, join, "")...)
					})

					test("select foo.*, bar.* from foo, bar", func() PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(1, "bar", "bar")
						join := NewJoinStage(CrossJoin, fooSource, barSource, SQLTrue)
						return NewProjectStage(join, createAllProjectedColumnsFromSource(1, join, "")...)
					})
				})

				Convey("non-star simple queries", func() {
					test("select a from foo", func() PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return NewProjectStage(source, createProjectedColumn(1, source, "foo", "a", "", "a"))
					})

					test("select a from foo f", func() PlanStage {
						source := createMongoSource(1, "foo", "f")
						return NewProjectStage(source, createProjectedColumn(1, source, "f", "a", "", "a"))
					})

					test("select f.a from foo f", func() PlanStage {
						source := createMongoSource(1, "foo", "f")
						return NewProjectStage(source, createProjectedColumn(1, source, "f", "a", "", "a"))
					})

					test("select a as b from foo", func() PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return NewProjectStage(source, createProjectedColumn(1, source, "foo", "a", "", "b"))
					})

					test("select a + 2 from foo", func() PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return NewProjectStage(source,
							createProjectedColumnFromSQLExpr(1, "", "a+2",
								&SQLAddExpr{
									left:  NewSQLColumnExpr(1, "foo", "a", schema.SQLInt, schema.MongoInt),
									right: SQLInt(2),
								},
							),
						)
					})

					test("select a + 2 as b from foo", func() PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return NewProjectStage(source,
							createProjectedColumnFromSQLExpr(1, "", "b",
								&SQLAddExpr{
									left:  NewSQLColumnExpr(1, "foo", "a", schema.SQLInt, schema.MongoInt),
									right: SQLInt(2),
								},
							),
						)
					})

					test("select ASCII(a) from foo", func() PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return NewProjectStage(source,
							createProjectedColumnFromSQLExpr(1, "", "ascii(a)",
								&SQLScalarFunctionExpr{
									Name:  "ascii",
									Exprs: []SQLExpr{NewSQLColumnExpr(1, "foo", "a", schema.SQLInt, schema.MongoInt)},
								},
							),
						)
					})
				})

				Convey("subqueries", func() {

					Convey("non-correlated", func() {
						test("select a, (select a from bar) from foo", func() PlanStage {
							fooSource := createMongoSource(1, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							return NewProjectStage(fooSource,
								createProjectedColumn(1, fooSource, "foo", "a", "", "a"),
								createProjectedColumnFromSQLExpr(1, "", "(select a from bar)",
									&SQLSubqueryExpr{
										plan: NewProjectStage(barSource, createProjectedColumn(2, barSource, "bar", "a", "", "a")),
									},
								),
							)
						})

						test("select a, (select a from bar) as b from foo", func() PlanStage {
							fooSource := createMongoSource(1, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							return NewProjectStage(fooSource,
								createProjectedColumn(1, fooSource, "foo", "a", "", "a"),
								createProjectedColumnFromSQLExpr(1, "", "b",
									&SQLSubqueryExpr{
										plan: NewProjectStage(barSource, createProjectedColumn(2, barSource, "bar", "a", "", "a")),
									},
								),
							)
						})

						test("select a, (select foo.a from foo, bar) from foo", func() PlanStage {
							foo1Source := createMongoSource(1, "foo", "foo")
							foo2Source := createMongoSource(2, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							join := NewJoinStage(CrossJoin, foo2Source, barSource, SQLTrue)
							return NewProjectStage(foo1Source,
								createProjectedColumn(1, foo1Source, "foo", "a", "", "a"),
								createProjectedColumnFromSQLExpr(1, "", "(select foo.a from foo, bar)",
									&SQLSubqueryExpr{
										plan: NewProjectStage(join, createProjectedColumn(2, join, "foo", "a", "", "a")),
									},
								),
							)
						})

						test("select exists(select 1 from bar) from foo", func() PlanStage {
							fooSource := createMongoSource(1, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							return NewProjectStage(fooSource,
								createProjectedColumnFromSQLExpr(1, "", "exists (select 1 from bar)",
									&SQLExistsExpr{
										expr: &SQLSubqueryExpr{
											plan: NewProjectStage(
												barSource,
												createProjectedColumnFromSQLExpr(2, "", "1", SQLInt(1)),
											),
										},
									},
								),
							)
						})
					})

					Convey("correlated", func() {
						test("select a, (select foo.a from bar) from foo", func() PlanStage {
							fooSource := createMongoSource(1, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							return NewProjectStage(
								fooSource,
								createProjectedColumn(1, fooSource, "foo", "a", "", "a"),
								createProjectedColumnFromSQLExpr(1, "", "(select foo.a from bar)",
									&SQLSubqueryExpr{
										plan:       NewProjectStage(barSource, createProjectedColumn(2, fooSource, "foo", "a", "", "a")),
										correlated: true,
									},
								),
							)
						})

						test("select * from (select b2.d, b2.b from bar b1 inner join bar b2 on (b1.a=b2.b) group by 1, 2) t0 HAVING (sum(1) > 0 )", func() PlanStage {
							b1Source := createMongoSource(2, "bar", "b1")
							b2Source := createMongoSource(2, "bar", "b2")

							matcher := &SQLEqualsExpr{
								left:  createSQLColumnExprFromSource(b1Source, "b1", "a"),
								right: createSQLColumnExprFromSource(b2Source, "b2", "b"),
							}

							join := NewJoinStage(InnerJoin, b1Source, b2Source, matcher)

							innerGroup := NewGroupByStage(
								join,
								[]SQLExpr{
									createSQLColumnExprFromSource(join, "b2", "d"),
									createSQLColumnExprFromSource(join, "b2", "b"),
								},
								ProjectedColumns{
									createProjectedColumn(2, join, "b2", "b", "b2", "b"),
									createProjectedColumn(2, join, "b2", "d", "b2", "d"),
								},
							)

							subquery := NewSubquerySourceStage(
								NewProjectStage(
									innerGroup,
									createProjectedColumn(2, join, "b2", "d", "", "d"),
									createProjectedColumn(2, join, "b2", "b", "", "b"),
								),
								2,
								"t0",
							)

							outerGroup := NewGroupByStage(
								subquery,
								nil,
								ProjectedColumns{
									createProjectedColumn(2, subquery, subquery.aliasName, "d", subquery.aliasName, "d"),
									createProjectedColumn(2, subquery, subquery.aliasName, "b", subquery.aliasName, "b"),
									createProjectedColumnFromSQLExpr(1, "", "sum(1)", &SQLAggFunctionExpr{
										Name:  "sum",
										Exprs: []SQLExpr{SQLInt(1)},
									}),
								},
							)

							filter := NewFilterStage(
								outerGroup,
								&SQLGreaterThanExpr{
									left:  NewSQLColumnExpr(1, "", "sum(1)", schema.SQLFloat, schema.MongoNone),
									right: SQLInt(0),
								},
							)

							project := NewProjectStage(
								filter,
								createProjectedColumn(1, subquery, subquery.aliasName, "d", "", "d"),
								createProjectedColumn(1, subquery, subquery.aliasName, "b", "", "b"),
							)

							return project
						})
					})
				})
			})

			Convey("where", func() {
				test("select a from foo where a", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewFilterStage(source, NewSQLColumnExpr(1, "foo", "a", schema.SQLInt, schema.MongoInt)),
						createProjectedColumn(1, source, "foo", "a", "", "a"),
					)
				})

				test("select a from foo where false", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewFilterStage(source, SQLFalse),
						createProjectedColumn(1, source, "foo", "a", "", "a"),
					)
				})

				test("select a from foo where true", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewFilterStage(source, SQLTrue),
						createProjectedColumn(1, source, "foo", "a", "", "a"),
					)
				})

				test("select a from foo where g = true", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewFilterStage(source,
							&SQLEqualsExpr{
								left:  NewSQLColumnExpr(1, "foo", "g", schema.SQLBoolean, schema.MongoBool),
								right: SQLTrue,
							},
						),
						createProjectedColumn(1, source, "foo", "a", "", "a"),
					)
				})

				test("select a from foo where a > 10", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewFilterStage(source,
							&SQLGreaterThanExpr{
								left:  NewSQLColumnExpr(1, "foo", "a", schema.SQLInt, schema.MongoInt),
								right: SQLInt(10),
							},
						),
						createProjectedColumn(1, source, "foo", "a", "", "a"),
					)
				})

				test("select a as b from foo where b > 10", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewFilterStage(source,
							&SQLGreaterThanExpr{
								left:  NewSQLColumnExpr(1, "foo", "b", schema.SQLInt, schema.MongoInt),
								right: SQLInt(10),
							},
						),
						createProjectedColumn(1, source, "foo", "a", "", "b"),
					)
				})

				Convey("subqueries", func() {
					Convey("correlated", func() {
						test("select a from foo where (b) = (select b from bar where foo.a = bar.a)", func() PlanStage {
							fooSource := createMongoSource(1, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							return NewProjectStage(
								NewFilterStage(
									fooSource,
									&SQLEqualsExpr{
										left: createSQLColumnExprFromSource(fooSource, "foo", "b"),
										right: &SQLSubqueryExpr{
											correlated: true,
											plan: NewProjectStage(
												NewFilterStage(
													barSource,
													&SQLEqualsExpr{
														left:  createSQLColumnExprFromSource(fooSource, "foo", "a"),
														right: createSQLColumnExprFromSource(barSource, "bar", "a"),
													},
												),
												createProjectedColumn(2, barSource, "bar", "b", "", "b"),
											),
										},
									},
								),
								createProjectedColumn(1, fooSource, "foo", "a", "", "a"),
							)
						})

						test("select a from foo f where (b) = (select b from bar where exists(select 1 from foo where f.a = a))", func() PlanStage {
							fooSource := createMongoSource(1, "foo", "f")
							barSource := createMongoSource(2, "bar", "bar")
							foo3Source := createMongoSource(3, "foo", "foo")
							return NewProjectStage(
								NewFilterStage(
									fooSource,
									&SQLEqualsExpr{
										left: createSQLColumnExprFromSource(fooSource, "f", "b"),
										right: &SQLSubqueryExpr{
											correlated: true,
											plan: NewProjectStage(
												NewFilterStage(
													barSource,
													&SQLExistsExpr{
														expr: &SQLSubqueryExpr{
															correlated: true,
															plan: NewProjectStage(
																NewFilterStage(
																	foo3Source,
																	&SQLEqualsExpr{
																		left:  createSQLColumnExprFromSource(fooSource, "f", "a"),
																		right: createSQLColumnExprFromSource(foo3Source, "foo", "a"),
																	},
																),
																createProjectedColumnFromSQLExpr(3, "", "1", SQLInt(1)),
															),
														},
													},
												),
												createProjectedColumn(2, barSource, "bar", "b", "", "b"),
											),
										},
									},
								),
								createProjectedColumn(1, fooSource, "f", "a", "", "a"),
							)
						})

						test("select a from foo where (b) = (select b from bar where exists(select 1 from foo where bar.a = a))", func() PlanStage {
							fooSource := createMongoSource(1, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							foo3Source := createMongoSource(3, "foo", "foo")
							return NewProjectStage(
								NewFilterStage(
									fooSource,
									&SQLEqualsExpr{
										left: createSQLColumnExprFromSource(fooSource, "foo", "b"),
										right: &SQLSubqueryExpr{
											correlated: false,
											plan: NewProjectStage(
												NewFilterStage(
													barSource,
													&SQLExistsExpr{
														expr: &SQLSubqueryExpr{
															correlated: true,
															plan: NewProjectStage(
																NewFilterStage(
																	foo3Source,
																	&SQLEqualsExpr{
																		left:  createSQLColumnExprFromSource(barSource, "bar", "a"),
																		right: createSQLColumnExprFromSource(foo3Source, "foo", "a"),
																	},
																),
																createProjectedColumnFromSQLExpr(3, "", "1", SQLInt(1)),
															),
														},
													},
												),
												createProjectedColumn(2, barSource, "bar", "b", "", "b"),
											),
										},
									},
								),
								createProjectedColumn(1, fooSource, "foo", "a", "", "a"),
							)
						})
					})
				})
			})

			Convey("group by", func() {
				test("select sum(a) from foo", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewGroupByStage(source,
							nil,
							ProjectedColumns{
								createProjectedColumnFromSQLExpr(1, "", "sum(foo.a)", &SQLAggFunctionExpr{
									Name:  "sum",
									Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
								}),
							},
						),
						createProjectedColumnFromSQLExpr(1, "", "sum(a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})

				test("select sum(a) from foo group by b", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewGroupByStage(source,
							[]SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
							ProjectedColumns{
								createProjectedColumnFromSQLExpr(1, "", "sum(foo.a)", &SQLAggFunctionExpr{
									Name:  "sum",
									Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
								}),
							},
						),
						createProjectedColumnFromSQLExpr(1, "", "sum(a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})

				test("select a, sum(a) from foo group by b", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewGroupByStage(source,
							[]SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
							ProjectedColumns{
								createProjectedColumn(1, source, "foo", "a", "foo", "a"),
								createProjectedColumnFromSQLExpr(1, "", "sum(foo.a)", &SQLAggFunctionExpr{
									Name:  "sum",
									Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
								}),
							},
						),
						createProjectedColumn(1, source, "foo", "a", "", "a"),
						createProjectedColumnFromSQLExpr(1, "", "sum(a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})

				test("select sum(a) from foo group by b order by sum(a)", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewOrderByStage(
							NewGroupByStage(source,
								[]SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
								ProjectedColumns{
									createProjectedColumnFromSQLExpr(1, "", "sum(foo.a)", &SQLAggFunctionExpr{
										Name:  "sum",
										Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
									}),
								},
							),
							&orderByTerm{
								expr:      NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone),
								ascending: true,
							},
						),
						createProjectedColumnFromSQLExpr(1, "", "sum(a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})

				test("select sum(a) as sum_a from foo group by b order by sum_a", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewOrderByStage(
							NewGroupByStage(source,
								[]SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
								ProjectedColumns{
									createProjectedColumnFromSQLExpr(1, "", "sum(foo.a)", &SQLAggFunctionExpr{
										Name:  "sum",
										Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
									}),
								},
							),
							&orderByTerm{
								expr:      NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone),
								ascending: true,
							},
						),
						createProjectedColumnFromSQLExpr(1, "", "sum_a", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})

				test("select sum(a) from foo f group by b order by (select c from foo where f.b = b)", func() PlanStage {
					foo1Source := createMongoSource(1, "foo", "f")
					foo2Source := createMongoSource(2, "foo", "foo")
					return NewProjectStage(
						NewOrderByStage(
							NewGroupByStage(foo1Source,
								[]SQLExpr{createSQLColumnExprFromSource(foo1Source, "f", "b")},
								ProjectedColumns{
									createProjectedColumn(1, foo1Source, "f", "b", "f", "b"),
									createProjectedColumnFromSQLExpr(1, "", "sum(f.a)", &SQLAggFunctionExpr{
										Name:  "sum",
										Exprs: []SQLExpr{createSQLColumnExprFromSource(foo1Source, "f", "a")},
									}),
								},
							),
							&orderByTerm{
								expr: &SQLSubqueryExpr{
									correlated: true,
									plan: NewProjectStage(
										NewFilterStage(
											foo2Source,
											&SQLEqualsExpr{
												left:  createSQLColumnExprFromSource(foo1Source, "f", "b"),
												right: createSQLColumnExprFromSource(foo2Source, "foo", "b"),
											},
										),
										createProjectedColumn(2, foo2Source, "foo", "c", "", "c"),
									),
								},
								ascending: true,
							},
						),
						createProjectedColumnFromSQLExpr(1, "", "sum(a)", NewSQLColumnExpr(1, "", "sum(f.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})

				test("select (select sum(foo.a) from foo as f) from foo group by b", func() PlanStage {
					foo1Source := createMongoSource(1, "foo", "foo")
					foo2Source := createMongoSource(2, "foo", "f")
					return NewProjectStage(
						NewGroupByStage(foo1Source,
							[]SQLExpr{createSQLColumnExprFromSource(foo1Source, "foo", "b")},
							ProjectedColumns{
								createProjectedColumnFromSQLExpr(1, "", "sum(foo.a)", &SQLAggFunctionExpr{
									Name: "sum",
									Exprs: []SQLExpr{
										createSQLColumnExprFromSource(foo1Source, "foo", "a"),
									},
								}),
							},
						),
						createProjectedColumnFromSQLExpr(1, "", "(select sum(foo.a) from foo as f)",
							&SQLSubqueryExpr{
								correlated: true,
								plan: NewProjectStage(
									foo2Source,
									createProjectedColumnFromSQLExpr(2, "", "sum(foo.a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
								),
							},
						),
					)
				})

				test("select (select sum(f.a + foo.a) from foo f) from foo group by b", func() PlanStage {
					foo1Source := createMongoSource(1, "foo", "foo")
					foo2Source := createMongoSource(2, "foo", "f")
					return NewProjectStage(
						NewGroupByStage(foo1Source,
							[]SQLExpr{createSQLColumnExprFromSource(foo1Source, "foo", "b")},
							ProjectedColumns{
								createProjectedColumn(1, foo1Source, "foo", "a", "foo", "a"),
							},
						),
						createProjectedColumnFromSQLExpr(1, "", "(select sum(f.a+foo.a) from foo as f)",
							&SQLSubqueryExpr{
								correlated: true,
								plan: NewProjectStage(
									NewGroupByStage(
										foo2Source,
										nil,
										ProjectedColumns{
											createProjectedColumnFromSQLExpr(2, "", "sum(f.a+foo.a)", &SQLAggFunctionExpr{
												Name: "sum",
												Exprs: []SQLExpr{&SQLAddExpr{
													left:  createSQLColumnExprFromSource(foo2Source, "f", "a"),
													right: createSQLColumnExprFromSource(foo1Source, "foo", "a"),
												}},
											}),
										},
									),
									createProjectedColumnFromSQLExpr(2, "", "sum(f.a+foo.a)", NewSQLColumnExpr(2, "", "sum(f.a+foo.a)", schema.SQLFloat, schema.MongoNone)),
								),
							},
						),
					)
				})
			})

			Convey("having", func() {
				test("select a from foo group by b having sum(a) > 10", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewFilterStage(
							NewGroupByStage(source,
								[]SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
								ProjectedColumns{
									createProjectedColumn(1, source, "foo", "a", "foo", "a"),
									createProjectedColumnFromSQLExpr(1, "", "sum(foo.a)", &SQLAggFunctionExpr{
										Name:  "sum",
										Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
									}),
								},
							),
							&SQLGreaterThanExpr{
								left:  NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone),
								right: SQLInt(10),
							},
						),
						createProjectedColumn(1, source, "foo", "a", "", "a"),
					)
				})

				Convey("subqueries", func() {
					Convey("non-correlated", func() {
						test("select a from foo having exists(select 1 from bar)", func() PlanStage {
							fooSource := createMongoSource(1, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							return NewProjectStage(
								NewFilterStage(
									fooSource,
									&SQLExistsExpr{
										expr: &SQLSubqueryExpr{
											plan: NewProjectStage(
												barSource,
												createProjectedColumnFromSQLExpr(2, "", "1", SQLInt(1)),
											),
										},
									},
								),
								createProjectedColumn(1, fooSource, "foo", "a", "", "a"),
							)
						})
					})
				})
			})

			Convey("distinct", func() {
				test("select distinct a from foo", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewGroupByStage(source,
							[]SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
							ProjectedColumns{
								createProjectedColumn(1, source, "foo", "a", "foo", "a"),
							},
						),
						createProjectedColumn(1, source, "foo", "a", "", "a"),
					)
				})

				test("select distinct sum(a) from foo", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewGroupByStage(
							NewGroupByStage(source,
								nil,
								ProjectedColumns{
									createProjectedColumnFromSQLExpr(1, "", "sum(foo.a)", &SQLAggFunctionExpr{
										Name:  "sum",
										Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
									}),
								},
							),
							[]SQLExpr{NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)},
							ProjectedColumns{
								createProjectedColumnFromSQLExpr(1, "", "sum(foo.a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
							},
						),
						createProjectedColumnFromSQLExpr(1, "", "sum(a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})

				test("select distinct sum(a) from foo having sum(a) > 20", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewGroupByStage(
							NewFilterStage(
								NewGroupByStage(source,
									nil,
									ProjectedColumns{
										createProjectedColumnFromSQLExpr(1, "", "sum(foo.a)", &SQLAggFunctionExpr{
											Name:  "sum",
											Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
										}),
									},
								),
								&SQLGreaterThanExpr{
									left:  NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone),
									right: SQLInt(20),
								},
							),
							[]SQLExpr{NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)},
							ProjectedColumns{
								createProjectedColumnFromSQLExpr(1, "", "sum(foo.a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
							},
						),
						createProjectedColumnFromSQLExpr(1, "", "sum(a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
					)
				})
			})

			Convey("order by", func() {
				test("select a from foo order by a", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewOrderByStage(source,
							&orderByTerm{
								expr:      NewSQLColumnExpr(1, "foo", "a", schema.SQLInt, schema.MongoInt),
								ascending: true,
							},
						),
						createProjectedColumn(1, source, "foo", "a", "", "a"),
					)
				})

				test("select a as b from foo order by b", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewOrderByStage(source,
							&orderByTerm{
								expr:      NewSQLColumnExpr(1, "foo", "a", schema.SQLInt, schema.MongoInt),
								ascending: true,
							},
						),
						createProjectedColumn(1, source, "foo", "a", "", "b"),
					)
				})

				test("select a from foo order by foo.a", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewOrderByStage(source,
							&orderByTerm{
								expr:      NewSQLColumnExpr(1, "foo", "a", schema.SQLInt, schema.MongoInt),
								ascending: true,
							},
						),
						createProjectedColumn(1, source, "foo", "a", "", "a"),
					)
				})

				test("select a as b from foo order by foo.a", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewOrderByStage(source,
							&orderByTerm{
								expr:      NewSQLColumnExpr(1, "foo", "a", schema.SQLInt, schema.MongoInt),
								ascending: true,
							},
						),
						createProjectedColumn(1, source, "foo", "a", "", "b"),
					)
				})

				test("select a from foo order by 1", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewOrderByStage(source,
							&orderByTerm{
								expr:      NewSQLColumnExpr(1, "foo", "a", schema.SQLInt, schema.MongoInt),
								ascending: true,
							},
						),
						createProjectedColumn(1, source, "foo", "a", "", "a"),
					)
				})

				test("select * from foo order by 2", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewOrderByStage(source,
							&orderByTerm{
								expr:      NewSQLColumnExpr(1, "foo", "b", schema.SQLInt, schema.MongoInt),
								ascending: true,
							},
						),
						createAllProjectedColumnsFromSource(1, source, "")...,
					)
				})

				test("select foo.* from foo order by 2", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewOrderByStage(source,
							&orderByTerm{
								expr:      NewSQLColumnExpr(1, "foo", "b", schema.SQLInt, schema.MongoInt),
								ascending: true,
							},
						),
						createAllProjectedColumnsFromSource(1, source, "")...,
					)
				})

				test("select foo.*, foo.a from foo order by 2", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					columns := append(createAllProjectedColumnsFromSource(1, source, ""), createProjectedColumn(1, source, "foo", "a", "", "a"))
					return NewProjectStage(
						NewOrderByStage(source,
							&orderByTerm{
								expr:      NewSQLColumnExpr(1, "foo", "b", schema.SQLInt, schema.MongoInt),
								ascending: true,
							},
						),
						columns...,
					)
				})

				test("select a from foo order by -1", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewOrderByStage(source,
							&orderByTerm{
								expr:      SQLInt(-1),
								ascending: true,
							},
						),
						createProjectedColumn(1, source, "foo", "a", "", "a"),
					)
				})

				test("select a + b as c from foo order by c - b", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewOrderByStage(source,
							&orderByTerm{
								expr: &SQLSubtractExpr{
									left: &SQLAddExpr{
										left:  NewSQLColumnExpr(1, "foo", "a", schema.SQLInt, schema.MongoInt),
										right: NewSQLColumnExpr(1, "foo", "b", schema.SQLInt, schema.MongoInt),
									},
									right: NewSQLColumnExpr(1, "foo", "b", schema.SQLInt, schema.MongoInt),
								},
								ascending: true,
							},
						),
						createProjectedColumnFromSQLExpr(1, "", "c",
							&SQLAddExpr{
								left:  NewSQLColumnExpr(1, "foo", "a", schema.SQLInt, schema.MongoInt),
								right: NewSQLColumnExpr(1, "foo", "b", schema.SQLInt, schema.MongoInt),
							},
						),
					)
				})

				Convey("subqueries", func() {
					Convey("non-correlated", func() {
						test("select a from foo order by (select a from bar)", func() PlanStage {
							fooSource := createMongoSource(1, "foo", "foo")
							barSource := createMongoSource(2, "bar", "bar")
							return NewProjectStage(
								NewOrderByStage(
									fooSource,
									&orderByTerm{
										expr: &SQLSubqueryExpr{
											plan: NewProjectStage(
												barSource,
												createProjectedColumn(2, barSource, "bar", "a", "", "a"),
											),
										},
										ascending: true,
									},
								),
								createProjectedColumn(1, fooSource, "foo", "a", "", "a"),
							)
						})
					})

				})
			})

			Convey("limit", func() {
				test("select a from foo limit 10", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewLimitStage(source, 0, 10),
						createProjectedColumn(1, source, "foo", "a", "", "a"),
					)
				})

				test("select a from foo limit 10, 20", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewProjectStage(
						NewLimitStage(source, 10, 20),
						createProjectedColumn(1, source, "foo", "a", "", "a"),
					)
				})

				test("select a from foo limit 10,0", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewEmptyStage([]*Column{
						createProjectedColumn(1, source, "foo", "a", "", "a").Column,
					}, collation.Default)
				})

				test("select a from foo limit 0, 0", func() PlanStage {
					source := createMongoSource(1, "foo", "foo")
					return NewEmptyStage([]*Column{
						createProjectedColumn(1, source, "foo", "a", "", "a").Column,
					}, collation.Default)
				})

				testVariables("select a from foo",
					&variable.Container{
						SQLSelectLimit: 10,
						MongoDBInfo:    testInfo,
					},
					func() PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return NewLimitStage(
							NewProjectStage(
								source,
								createProjectedColumn(1, source, "foo", "a", "", "a"),
							),
							0,
							10,
						)
					})

				testVariables("select b from foo",
					&variable.Container{
						SQLSelectLimit: 18446744073709551615,
						MongoDBInfo:    testInfo,
					},
					func() PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return NewProjectStage(
							source,
							createProjectedColumn(1, source, "foo", "b", "", "b"),
						)
					})

				testVariables("select b from foo limit 10, 20",
					&variable.Container{
						SQLSelectLimit: 5,
						MongoDBInfo:    testInfo,
					},
					func() PlanStage {
						source := createMongoSource(1, "foo", "foo")
						return NewProjectStage(
							NewLimitStage(source, 10, 20),
							createProjectedColumn(1, source, "foo", "b", "", "b"),
						)
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

				testError("select sum(a) from foo group by sum(a)", `ERROR 1056 (42000): Can't group on 'sum(foo.a)'`)
				testError("select sum(a) from foo group by (a + sum(a))", `ERROR 1056 (42000): Can't group on 'sum(foo.a)'`)
				testError("select sum(a) from foo group by 1", `ERROR 1056 (42000): Can't group on 'sum(foo.a)'`)
				testError("select a+sum(a) from foo group by 1", `ERROR 1056 (42000): Can't group on 'sum(foo.a)'`)
				testError("select sum(a) from foo group by 2", `ERROR 1054 (42S22): Unknown column '2' in 'group clause'`)

				testError("select a from foo, foo", `ERROR 1066 (42000): Not unique table/alias: 'foo'`)
				testError("select a from foo as bar, bar", `ERROR 1066 (42000): Not unique table/alias: 'bar'`)
				testError("select a from foo as g, foo as g", `ERROR 1066 (42000): Not unique table/alias: 'g'`)

				testError("select a from foo left outer join bar where a = 10", `ERROR 1064 (42000): A left join requires criteria`)

				testError("select bar.d, baz.a from bar join baz using (tomato)", `ERROR 1054 (42S22): Unknown column 'bar.tomato' in 'from clause'`)
				testError("select * from baz join bar using (d)", `ERROR 1054 (42S22): Unknown column 'baz.d' in 'from clause'`)
				testError("select bar.d, baz.a from bar join (select * from baz join foo) using (c)", `ERROR 1248 (42000): Every derived table must have its own alias`)
				testError("select bar.d, biz.a from bar join (select * from baz join foo) as biz using (c)", `ERROR 1060 (42S21): Duplicate column name 'biz.a'`)
				testError("select * from bar join foo join baz using (c)", "ERROR 1054 (42S22): Unknown column 'baz.c' in 'from clause'")
				testError("select * from bar join foo join baz using (_id)", "ERROR 1052 (23000): Column '_id' in from clause is ambiguous")
				testError("select * from baz join bar join foo using (c)", "ERROR 1054 (42S22): Unknown column 'c' in 'from clause'")
			})
		})
	})
}

func TestAlgebrizeCommand(t *testing.T) {

	testSchema, err := schema.New(testSchema1)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}
	testInfo := getMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)
	testVars := createTestVariables(testInfo)
	testCatalog := getCatalogFromSchema(testSchema, testVars)
	defaultDbName := "test"

	test := func(sql string, expectedPlanFactory func() command) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			actual, err := AlgebrizeCommand(statement, defaultDbName, testVars, testCatalog)
			So(err, ShouldBeNil)

			expected := expectedPlanFactory()

			if ShouldResemble(actual, expected) != "" {
				fmt.Printf("\nExpected: %# v", pretty.Formatter(expected))
				fmt.Printf("\nActual: %# v", pretty.Formatter(actual))
			}

			So(actual, ShouldResemble, expected)
		})
	}

	createMongoSource := func(selectID int, tableName, aliasName string) PlanStage {
		db, _ := testCatalog.Database(defaultDbName)
		table, _ := db.Table(tableName)
		r := NewMongoSourceStage(db, table.(*catalog.MongoTable), selectID, aliasName)
		return r
	}

	Convey("Subject: Algebrize Kill Statements", t, func() {
		test("kill 3", func() command {
			return NewKillCommand(SQLInt(3), KillConnection)
		})
		test("kill query 3", func() command {
			return NewKillCommand(SQLInt(3), KillQuery)
		})
		test("kill query 5*3", func() command {
			return NewKillCommand(
				&SQLMultiplyExpr{
					SQLInt(5),
					SQLInt(3),
				}, KillQuery,
			)
		})
		test("kill connection 5-3", func() command {
			return NewKillCommand(
				&SQLSubtractExpr{
					SQLInt(5),
					SQLInt(3),
				}, KillConnection,
			)
		})
	})

	Convey("Subject: Algebrize Flush Statements", t, func() {
		test("flush logs", func() command {
			return NewFlushCommand()
		})
	})

	Convey("Subject: Algebrize Set Statements", t, func() {
		test("set @t1 = 132", func() command {
			return NewSetCommand(
				[]*SQLAssignmentExpr{
					&SQLAssignmentExpr{
						variable: &SQLVariableExpr{
							Name:  "t1",
							Kind:  variable.UserKind,
							Scope: variable.SessionScope,
						},
						expr: SQLInt(132),
					},
				},
			)
		})

		test("set @@max_allowed_packet = 12", func() command {
			return NewSetCommand(
				[]*SQLAssignmentExpr{
					&SQLAssignmentExpr{
						variable: NewSQLVariableExpr(
							"max_allowed_packet",
							variable.SystemKind,
							variable.SessionScope,
							schema.SQLInt,
						),
						expr: SQLInt(12),
					},
				},
			)
		})

		test("set @@global.max_allowed_packet = 12", func() command {
			return NewSetCommand(
				[]*SQLAssignmentExpr{
					&SQLAssignmentExpr{
						variable: NewSQLVariableExpr(
							"max_allowed_packet",
							variable.SystemKind,
							variable.GlobalScope,
							schema.SQLInt,
						),
						expr: SQLInt(12),
					},
				},
			)
		})

		test("set @@global.max_allowed_packet = (select a from foo)", func() command {
			fooSource := createMongoSource(1, "foo", "foo")
			return NewSetCommand(
				[]*SQLAssignmentExpr{
					&SQLAssignmentExpr{
						variable: NewSQLVariableExpr(
							"max_allowed_packet",
							variable.SystemKind,
							variable.GlobalScope,
							schema.SQLInt,
						),
						expr: &SQLSubqueryExpr{
							correlated: false,
							plan: NewProjectStage(
								fooSource,
								createProjectedColumn(1, fooSource, "foo", "a", "", "a")),
						},
					},
				},
			)
		})

		test("set @@max_allowed_packet=12, @interactive_timeout=1111", func() command {
			return NewSetCommand(
				[]*SQLAssignmentExpr{
					&SQLAssignmentExpr{
						variable: NewSQLVariableExpr(
							"max_allowed_packet",
							variable.SystemKind,
							variable.SessionScope,
							schema.SQLInt,
						),
						expr: SQLInt(12),
					},
					&SQLAssignmentExpr{
						variable: &SQLVariableExpr{
							Name:  "interactive_timeout",
							Kind:  variable.UserKind,
							Scope: variable.SessionScope,
						},
						expr: SQLInt(1111),
					},
				},
			)
		})
	})
}

func TestAlgebrizeExpr(t *testing.T) {
	testSchema, _ := schema.New(testSchema1)
	testInfo := getMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)
	testVars := createTestVariables(testInfo)
	testCatalog := getCatalogFromSchema(testSchema, testVars)
	testDB, _ := testCatalog.Database("test")
	fooTable, _ := testDB.Table("foo")
	source := NewMongoSourceStage(testDB, fooTable.(*catalog.MongoTable), 1, "foo")

	test := func(sql string, expected SQLExpr) {
		Convey(sql, func() {
			statement, err := parser.Parse("select " + sql + " from foo")
			So(err, ShouldBeNil)

			actualPlan, err := AlgebrizeQuery(statement, "test", testVars, testCatalog)
			So(err, ShouldBeNil)
			actual := (actualPlan.(*ProjectStage)).projectedColumns[0].Expr
			So(actual, ShouldResemble, expected)
		})
	}

	testError := func(sql, message string) {
		Convey(sql, func() {
			statement, err := parser.Parse("select " + sql + " from foo")
			So(err, ShouldBeNil)

			actual, err := AlgebrizeQuery(statement, "test", testVars, testCatalog)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, message)
			So(actual, ShouldBeNil)
		})
	}

	createSQLColumnExpr := func(columnName string) SQLColumnExpr {
		for _, c := range source.Columns() {
			if c.Name == columnName {
				return NewSQLColumnExpr(1, c.Table, c.Name, c.SQLType, c.MongoType)
			}
		}

		panic("column not found")
	}

	Convey("Subject: Algebrize Expressions", t, func() {
		Convey("And", func() {
			test("a = 1 AND b = 2", &SQLAndExpr{
				left: &SQLEqualsExpr{
					left:  createSQLColumnExpr("a"),
					right: SQLInt(1),
				},
				right: &SQLEqualsExpr{
					left:  createSQLColumnExpr("b"),
					right: SQLInt(2),
				},
			})
		})
		Convey("Add", func() {
			test("a + 1", &SQLAddExpr{
				left:  createSQLColumnExpr("a"),
				right: SQLInt(1),
			})
		})

		SkipConvey("Case", func() {
		})

		Convey("Date", func() {
			expected := time.Date(2006, time.December, 31, 0, 0, 0, 0, time.UTC)
			test("DATE '2006-12-31'", SQLDate{expected})
			test("DATE '06-12-31'", SQLDate{expected})
			test("DATE '20061231'", SQLDate{expected})
			test("DATE '061231'", SQLDate{expected})

			testError("DATE '2014-13-07'", "ERROR 1525 (HY000): Incorrect DATE value: '2014-13-07'")
			testError("DATE '2014-12-32'", "ERROR 1525 (HY000): Incorrect DATE value: '2014-12-32'")
			testError("DATE '2006-12-31 10:32:46'", "ERROR 1525 (HY000): Incorrect DATE value: '2006-12-31 10:32:46'")
		})

		Convey("Divide", func() {
			test("a / 1", &SQLDivideExpr{
				left:  createSQLColumnExpr("a"),
				right: SQLInt(1),
			})
		})

		Convey("Equals", func() {
			test("a = 1", &SQLEqualsExpr{
				left:  createSQLColumnExpr("a"),
				right: SQLInt(1),
			})
			test("g = 0", &SQLEqualsExpr{
				left:  createSQLColumnExpr("g"),
				right: &SQLConvertExpr{SQLInt(0), schema.SQLBoolean, SQLNone},
			})
			test("g = 1", &SQLEqualsExpr{
				left:  createSQLColumnExpr("g"),
				right: &SQLConvertExpr{SQLInt(1), schema.SQLBoolean, SQLNone},
			})
			test("g = 2", &SQLEqualsExpr{
				left:  &SQLConvertExpr{createSQLColumnExpr("g"), schema.SQLInt, SQLNone},
				right: SQLInt(2),
			})
			test("0 = g", &SQLEqualsExpr{
				left:  createSQLColumnExpr("g"),
				right: &SQLConvertExpr{SQLInt(0), schema.SQLBoolean, SQLNone},
			})
			test("1 = g", &SQLEqualsExpr{
				left:  createSQLColumnExpr("g"),
				right: &SQLConvertExpr{SQLInt(1), schema.SQLBoolean, SQLNone},
			})
			test("2 = g", &SQLEqualsExpr{
				left:  SQLInt(2),
				right: &SQLConvertExpr{createSQLColumnExpr("g"), schema.SQLInt, SQLNone},
			})
		})

		SkipConvey("Exists", func() {
		})

		Convey("Greater Than", func() {
			test("a > 1", &SQLGreaterThanExpr{
				left:  createSQLColumnExpr("a"),
				right: SQLInt(1),
			})
		})

		Convey("Greater Than Or Equal", func() {
			test("a >= 1", &SQLGreaterThanOrEqualExpr{
				left:  createSQLColumnExpr("a"),
				right: SQLInt(1),
			})
		})

		SkipConvey("In", func() {
		})

		Convey("Is", func() {
			test("a is true", NewSQLIsExpr(
				createSQLColumnExpr("a"),
				SQLTrue,
			))
		})

		Convey("Is Not", func() {
			test("a is not true", &SQLNotExpr{NewSQLIsExpr(
				createSQLColumnExpr("a"),
				SQLTrue,
			)})
		})

		Convey("Is Null", func() {
			test("a IS NULL", NewSQLIsExpr(
				createSQLColumnExpr("a"),
				SQLNull,
			))
		})

		Convey("Is Not Null", func() {
			test("a IS NOT NULL", &SQLNotExpr{NewSQLIsExpr(
				createSQLColumnExpr("a"),
				SQLNull,
			)})
		})

		Convey("Less Than", func() {
			test("a < 1", &SQLLessThanExpr{
				left:  createSQLColumnExpr("a"),
				right: SQLInt(1),
			})
		})

		Convey("Less Than Or Equal", func() {
			test("a <= 1", &SQLLessThanOrEqualExpr{
				left:  createSQLColumnExpr("a"),
				right: SQLInt(1),
			})
		})

		SkipConvey("Like", func() {
		})

		Convey("Multiple", func() {
			test("a * 1", &SQLMultiplyExpr{
				left:  createSQLColumnExpr("a"),
				right: SQLInt(1),
			})
		})

		Convey("Not", func() {
			test("NOT a", &SQLNotExpr{
				createSQLColumnExpr("a"),
			})
		})

		Convey("NotEquals", func() {
			test("a != 1", &SQLNotEqualsExpr{
				left:  createSQLColumnExpr("a"),
				right: SQLInt(1),
			})
		})

		Convey("NullSafeEquals", func() {
			test("a <=> 1", &SQLNullSafeEqualsExpr{
				left:  createSQLColumnExpr("a"),
				right: SQLInt(1),
			})
		})

		SkipConvey("Not In", func() {
		})

		Convey("Null", func() {
			test("NULL", SQLNull)
		})

		Convey("True", func() {
			test("TRUE", SQLTrue)
		})

		Convey("False", func() {
			test("FALSE", SQLFalse)
		})

		Convey("Number", func() {
			test("20", SQLInt(20))
			test("-20", SQLInt(-20))
			test("202E-1", SQLFloat(20.2))
			test("-202E-1", SQLFloat(-20.2))
			test("20.2", SQLDecimal128(decimal.New(202, -1)))
			test("-20.2", SQLDecimal128(decimal.New(-202, -1)))
			d, _ := decimal.NewFromString("100000000000000000000000000000000000")
			test("100000000000000000000000000000000000", SQLDecimal128(d))

			oldVersionArray := testInfo.VersionArray
			testInfo.VersionArray = []uint8{3, 2, 0}
			test("30.2", SQLFloat(30.2))
			test("-30.2", SQLFloat(-30.2))
			f, _ := strconv.ParseFloat("1000000000000000000000000000000000000", 64)
			test("1000000000000000000000000000000000000", SQLFloat(f))
			testInfo.VersionArray = oldVersionArray
		})

		Convey("Or", func() {
			test("a = 1 OR b = 2", &SQLOrExpr{
				left: &SQLEqualsExpr{
					left:  createSQLColumnExpr("a"),
					right: SQLInt(1),
				},
				right: &SQLEqualsExpr{
					left:  createSQLColumnExpr("b"),
					right: SQLInt(2),
				},
			})
		})

		Convey("Paren Boolean", func() {
			test("(1)", SQLInt(1))
		})

		Convey("Range", func() {
			test("a BETWEEN 0 AND 20", &SQLAndExpr{
				left: &SQLGreaterThanOrEqualExpr{
					left:  createSQLColumnExpr("a"),
					right: SQLInt(0),
				},
				right: &SQLLessThanOrEqualExpr{
					left:  createSQLColumnExpr("a"),
					right: SQLInt(20),
				},
			})

			test("a NOT BETWEEN 0 AND 20", &SQLNotExpr{
				&SQLAndExpr{
					left: &SQLGreaterThanOrEqualExpr{
						left:  createSQLColumnExpr("a"),
						right: SQLInt(0),
					},
					right: &SQLLessThanOrEqualExpr{
						left:  createSQLColumnExpr("a"),
						right: SQLInt(20),
					},
				},
			})
		})

		Convey("Scalar Function", func() {
			test("ascii(a)", &SQLScalarFunctionExpr{
				Name:  "ascii",
				Exprs: []SQLExpr{createSQLColumnExpr("a")},
			})
		})

		SkipConvey("Subquery", func() {
		})

		Convey("Subtract", func() {
			test("a - 1", &SQLSubtractExpr{
				left:  createSQLColumnExpr("a"),
				right: SQLInt(1),
			})
		})

		Convey("Time", func() {
			expected := time.Date(0, 1, 1, 10, 32, 46, 5000, time.UTC)
			test("TIME '10:32:46.000005'", SQLTimestamp{expected})
			test("TIME '103246.000005'", SQLTimestamp{expected})

			testError("TIME '2014-12-32'", "ERROR 1525 (HY000): Incorrect TIME value: '2014-12-32'")
			testError("TIME '2006-12-31 10:32:46.000005'", "ERROR 1525 (HY000): Incorrect TIME value: '2006-12-31 10:32:46.000005'")
		})

		Convey("Timestamp", func() {
			expected := time.Date(2014, 6, 7, 10, 32, 46, 5000, time.UTC)
			test("TIMESTAMP '2014-06-07 10:32:46.000005'", SQLTimestamp{expected})
			test("TIMESTAMP '2014-6-7 10:32:46.000005'", SQLTimestamp{expected})
			test("TIMESTAMP '14-06-07 10:32:46.000005'", SQLTimestamp{expected})
			test("TIMESTAMP '14-6-7 10:32:46.000005'", SQLTimestamp{expected})
			test("TIMESTAMP '2014:06:07 10:32:46.000005'", SQLTimestamp{expected})
			test("TIMESTAMP '14:06:07 10:32:46.000005'", SQLTimestamp{expected})
			test("TIMESTAMP '20140607103246.000005'", SQLTimestamp{expected})
			test("TIMESTAMP '140607103246.000005'", SQLTimestamp{expected})
			test("TIMESTAMP '146.07103246.000005'", SQLTimestamp{expected})
			test("TIMESTAMP '14.06.07.10.32.46.000005'", SQLTimestamp{expected})

			testError("TIMESTAMP '2014-06-07'", "ERROR 1525 (HY000): Incorrect DATETIME value: '2014-06-07'")
		})

		Convey("Tuple", func() {
			test("(a)", createSQLColumnExpr("a"))
		})

		Convey("Unary Minus", func() {
			test("-a", &SQLUnaryMinusExpr{createSQLColumnExpr("a")})
			test("-c", &SQLUnaryMinusExpr{createSQLColumnExpr("c")})
			test("-g", &SQLUnaryMinusExpr{&SQLConvertExpr{createSQLColumnExpr("g"), schema.SQLInt, SQLNone}})
			test("-_id", &SQLUnaryMinusExpr{&SQLConvertExpr{createSQLColumnExpr("_id"), schema.SQLInt, SQLNone}})
		})

		Convey("Unary Tilde", func() {
			test("~a", &SQLUnaryTildeExpr{createSQLColumnExpr("a")})
		})

		Convey("Varchar", func() {
			test("'a'", SQLVarchar("a"))
		})

		Convey("Variable", func() {
			varGlobal := NewSQLVariableExpr("sql_auto_is_null", variable.SystemKind, variable.GlobalScope, schema.SQLBoolean)
			varSession := NewSQLVariableExpr("sql_auto_is_null", variable.SystemKind, variable.SessionScope, schema.SQLBoolean)

			test("@@global.sql_auto_is_null", varGlobal)
			test("@@session.sql_auto_is_null", varSession)
			test("@@local.sql_auto_is_null", varSession)
			test("@@sql_auto_is_null", varSession)
			test("@hmmm", &SQLVariableExpr{Name: "hmmm", Kind: variable.UserKind, Scope: variable.SessionScope})
		})
	})
}

func TestNoSharedPipelines(t *testing.T) {
	sql := "select _id from merge_b limit 2"

	testSchema, err := schema.New(testSchema4)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}
	testInfo := getMongoDBInfo([]uint8{3, 2}, testSchema, mongodb.AllPrivileges)
	testVariables := createTestVariables(testInfo)
	testCatalog := getCatalogFromSchema(testSchema, testVariables)
	defaultDbName := "test"

	Convey("Subject: NoSharedPipelines", t, func() {
		statement, err := parser.Parse(sql)
		So(err, ShouldBeNil)

		plan, err := AlgebrizeQuery(statement, defaultDbName, testVariables, testCatalog)
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

		pg := &pipelineGatherer{}
		pg.visit(plan)
		So(pg.pipelines, ShouldResemble, expectedPipelines)

		db, err := testCatalog.Database("test")
		So(err, ShouldBeNil)
		table, err := db.Table("merge_b")
		So(err, ShouldBeNil)
		mTab, ok := table.(*catalog.MongoTable)
		So(ok, ShouldBeTrue)
		mTab.Pipeline[0] = bson.D{}

		pg = &pipelineGatherer{}
		pg.visit(plan)
		So(pg.pipelines, ShouldResemble, expectedPipelines)
	})
}
