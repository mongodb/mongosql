package evaluator

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewAlgebrize(t *testing.T) {

	testSchema, _ := schema.New(testSchema1)
	defaultDbName := "test"

	test := func(sql string, expectedPlanFactory func() PlanStage) {
		Convey(sql, func() {
			statement, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			selectStatement := statement.(*sqlparser.Select)
			actualPlan, err := Algebrize(selectStatement, defaultDbName, testSchema)
			So(err, ShouldBeNil)

			expectedPlan := expectedPlanFactory()

			//fmt.Printf("\nExpected: %# v", pretty.Formatter(expectedPlan))
			//fmt.Printf("\nActual: %# v", pretty.Formatter(actualPlan))

			So(actualPlan, ShouldResemble, expectedPlan)
		})
	}

	testError := func(sql, message string) {
		Convey(sql, func() {
			statement, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			selectStatement := statement.(*sqlparser.Select)
			actualPlan, err := Algebrize(selectStatement, defaultDbName, testSchema)
			So(actualPlan, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, fmt.Errorf(message))
		})
	}

	createSelectExpressionFromSQLExpr := func(tableName, columnName string, expr SQLExpr) SelectExpression {
		column := &Column{
			Table:     tableName,
			Name:      columnName,
			View:      columnName, // ???
			SQLType:   expr.Type(),
			MongoType: schema.MongoNone,
		}

		return SelectExpression{Column: column, Expr: expr}
	}

	createSelectExpression := func(source PlanStage, sourceTableName, sourceColumnName, projectedTableName, projectedColumnName string) SelectExpression {
		for _, c := range source.OpFields() {
			if c.Table == sourceTableName && c.Name == sourceColumnName {
				return SelectExpression{
					Column: &Column{
						Table:     projectedTableName,
						Name:      projectedColumnName,
						View:      projectedColumnName, // ???
						SQLType:   c.SQLType,
						MongoType: schema.MongoNone,
					},
					Expr: SQLColumnExpr{
						tableName:  c.Table,
						columnName: c.Name,
						columnType: schema.ColumnType{
							SQLType:   c.SQLType,
							MongoType: c.MongoType,
						},
					},
				}
			}
		}

		panic(fmt.Sprintf("no column found with the name %q", sourceColumnName))
	}

	createMongoSource := func(tableName, aliasName string) PlanStage {
		r, _ := NewMongoSourceStage(testSchema, defaultDbName, tableName, aliasName)
		return r
	}

	Convey("Subject: Algebrize", t, func() {
		Convey("non-star simple queries", func() {
			test("select a from foo", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(source, createSelectExpression(source, "foo", "a", "", "a"))
			})

			test("select a from foo f", func() PlanStage {
				source := createMongoSource("foo", "f")
				return NewProjectStage(source, createSelectExpression(source, "f", "a", "", "a"))
			})

			test("select f.a from foo f", func() PlanStage {
				source := createMongoSource("foo", "f")
				return NewProjectStage(source, createSelectExpression(source, "f", "a", "", "a"))
			})

			test("select a as b from foo", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(source, createSelectExpression(source, "foo", "a", "", "b"))
			})

			test("select a + 2 from foo", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(source,
					createSelectExpressionFromSQLExpr("", "a+2",
						&SQLAddExpr{
							left: SQLColumnExpr{
								tableName:  "foo",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoInt}},
							right: SQLInt(2),
						},
					),
				)
			})

			test("select a + 2 as b from foo", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(source,
					createSelectExpressionFromSQLExpr("", "b",
						&SQLAddExpr{
							left: SQLColumnExpr{
								tableName:  "foo",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoInt}},
							right: SQLInt(2),
						},
					),
				)
			})
		})

		Convey("joins", func() {
			test("select foo.a, bar.a from foo, bar", func() PlanStage {
				fooSource := createMongoSource("foo", "foo")
				barSource := createMongoSource("bar", "bar")
				join := NewJoinStage(CrossJoin, fooSource, barSource)
				return NewProjectStage(join,
					createSelectExpression(join, "foo", "a", "", "a"),
					createSelectExpression(join, "bar", "a", "", "a"),
				)
			})

			test("select f.a, bar.a from foo f, bar", func() PlanStage {
				fooSource := createMongoSource("foo", "f")
				barSource := createMongoSource("bar", "bar")
				join := NewJoinStage(CrossJoin, fooSource, barSource)
				return NewProjectStage(join,
					createSelectExpression(join, "f", "a", "", "a"),
					createSelectExpression(join, "bar", "a", "", "a"),
				)
			})

			test("select f.a, b.a from foo f, bar b", func() PlanStage {
				fooSource := createMongoSource("foo", "f")
				barSource := createMongoSource("bar", "b")
				join := NewJoinStage(CrossJoin, fooSource, barSource)
				return NewProjectStage(join,
					createSelectExpression(join, "f", "a", "", "a"),
					createSelectExpression(join, "b", "a", "", "a"),
				)
			})
		})

		Convey("subqueries as sources", func() {
			test("select a from (select a from foo)", func() PlanStage {
				source := createMongoSource("foo", "foo")
				subquery := &SubqueryStage{
					tableName: "",
					source:    NewProjectStage(source, createSelectExpression(source, "foo", "a", "", "a")),
				}
				return NewProjectStage(subquery, createSelectExpression(subquery, "", "a", "", "a"))
			})

			test("select a from (select a from foo) f", func() PlanStage {
				source := createMongoSource("foo", "foo")
				subquery := &SubqueryStage{
					tableName: "f",
					source:    NewProjectStage(source, createSelectExpression(source, "foo", "a", "f", "a")),
				}
				return NewProjectStage(subquery, createSelectExpression(subquery, "f", "a", "", "a"))
			})

			test("select f.a from (select a from foo) f", func() PlanStage {
				source := createMongoSource("foo", "foo")
				subquery := &SubqueryStage{
					tableName: "f",
					source:    NewProjectStage(source, createSelectExpression(source, "foo", "a", "f", "a")),
				}
				return NewProjectStage(subquery, createSelectExpression(subquery, "f", "a", "", "a"))
			})
		})

		Convey("errors", func() {
			testError("select a from idk", `table "idk" doesn't exist in db "test"`)
			testError("select idk from foo", `unknown column "idk"`)
			testError("select f.a from foo", `unknown column "a" in table "f"`)
			testError("select foo.a from foo f", `unknown column "a" in table "foo"`)
			testError("select a + idk from foo", `unknown column "idk"`)

			testError("select a from foo, bar", `column "a" in the field list is ambiguous`)
			testError("select foo.a from (select a from foo)", `unknown column "a" in table "foo"`)
		})

	})
}
