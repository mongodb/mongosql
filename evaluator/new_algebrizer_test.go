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

	createSelectExpressionFromColumn := func(column *Column, projectedTableName, projectedColumnName string) SelectExpression {
		return SelectExpression{
			Column: &Column{
				Table:     projectedTableName,
				Name:      projectedColumnName,
				View:      projectedColumnName, // ???
				SQLType:   column.SQLType,
				MongoType: column.MongoType,
			},
			Expr: SQLColumnExpr{
				tableName:  column.Table,
				columnName: column.Name,
				columnType: schema.ColumnType{
					SQLType:   column.SQLType,
					MongoType: column.MongoType,
				},
			},
		}
	}

	createSelectExpression := func(source PlanStage, sourceTableName, sourceColumnName, projectedTableName, projectedColumnName string) SelectExpression {
		for _, c := range source.OpFields() {
			if c.Table == sourceTableName && c.Name == sourceColumnName {
				return createSelectExpressionFromColumn(c, projectedTableName, projectedColumnName)
			}
		}

		panic(fmt.Sprintf("no column found with the name %q", sourceColumnName))
	}

	createAllSelectExpressionsFromSource := func(source PlanStage, projectedTableName string) SelectExpressions {
		results := SelectExpressions{}
		for _, c := range source.OpFields() {
			results = append(results, createSelectExpressionFromColumn(c, projectedTableName, c.Name))
		}

		return results
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

	createMongoSource := func(tableName, aliasName string) PlanStage {
		r, _ := NewMongoSourceStage(testSchema, defaultDbName, tableName, aliasName)
		return r
	}

	Convey("Subject: Algebrize", t, func() {
		Convey("star simple queries", func() {
			test("select * from foo", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(source, createAllSelectExpressionsFromSource(source, "")...)
			})

			test("select foo.* from foo", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(source, createAllSelectExpressionsFromSource(source, "")...)
			})

			test("select f.* from foo f", func() PlanStage {
				source := createMongoSource("foo", "f")
				return NewProjectStage(source, createAllSelectExpressionsFromSource(source, "")...)
			})

			test("select a, foo.* from foo", func() PlanStage {
				source := createMongoSource("foo", "foo")
				columns := append(
					SelectExpressions{createSelectExpression(source, "foo", "a", "", "a")},
					createAllSelectExpressionsFromSource(source, "")...)
				return NewProjectStage(source, columns...)
			})

			test("select foo.*, a from foo", func() PlanStage {
				source := createMongoSource("foo", "foo")
				columns := append(
					createAllSelectExpressionsFromSource(source, ""),
					createSelectExpression(source, "foo", "a", "", "a"))
				return NewProjectStage(source, columns...)
			})

			test("select a, f.* from foo f", func() PlanStage {
				source := createMongoSource("foo", "f")
				columns := append(
					SelectExpressions{createSelectExpression(source, "f", "a", "", "a")},
					createAllSelectExpressionsFromSource(source, "")...)
				return NewProjectStage(source, columns...)
			})

			test("select * from foo, bar", func() PlanStage {
				fooSource := createMongoSource("foo", "foo")
				barSource := createMongoSource("bar", "bar")
				join := NewJoinStage(CrossJoin, fooSource, barSource)
				return NewProjectStage(join, createAllSelectExpressionsFromSource(join, "")...)
			})

			test("select foo.*, bar.* from foo, bar", func() PlanStage {
				fooSource := createMongoSource("foo", "foo")
				barSource := createMongoSource("bar", "bar")
				join := NewJoinStage(CrossJoin, fooSource, barSource)
				return NewProjectStage(join, createAllSelectExpressionsFromSource(join, "")...)
			})
		})

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

		Convey("subqueries in select", func() {
			test("select a, (select a from bar) from foo", func() PlanStage {
				fooSource := createMongoSource("foo", "foo")
				barSource := createMongoSource("bar", "bar")
				return NewProjectStage(fooSource,
					createSelectExpression(fooSource, "foo", "a", "", "a"),
					createSelectExpressionFromSQLExpr("", "(select a from bar)",
						&SQLSubqueryExpr{
							plan: NewProjectStage(barSource, createSelectExpression(barSource, "bar", "a", "", "a")),
						},
					),
				)
			})

			test("select a, (select a from bar) as b from foo", func() PlanStage {
				fooSource := createMongoSource("foo", "foo")
				barSource := createMongoSource("bar", "bar")
				return NewProjectStage(fooSource,
					createSelectExpression(fooSource, "foo", "a", "", "a"),
					createSelectExpressionFromSQLExpr("", "b",
						&SQLSubqueryExpr{
							plan: NewProjectStage(barSource, createSelectExpression(barSource, "bar", "a", "", "a")),
						},
					),
				)
			})

			test("select a, (select foo.a from foo, bar) from foo", func() PlanStage {
				fooSource := createMongoSource("foo", "foo")
				barSource := createMongoSource("bar", "bar")
				join := NewJoinStage(CrossJoin, fooSource, barSource)
				return NewProjectStage(fooSource,
					createSelectExpression(fooSource, "foo", "a", "", "a"),
					createSelectExpressionFromSQLExpr("", "(select foo.a from foo, bar)",
						&SQLSubqueryExpr{
							plan: NewProjectStage(join, createSelectExpression(join, "foo", "a", "", "a")),
						},
					),
				)
			})

			test("select a, (select foo.a from bar) from foo", func() PlanStage {
				fooSource := createMongoSource("foo", "foo")
				barSource := createMongoSource("bar", "bar")
				return NewProjectStage(fooSource,
					createSelectExpression(fooSource, "foo", "a", "", "a"),
					createSelectExpressionFromSQLExpr("", "(select foo.a from bar)",
						&SQLSubqueryExpr{
							plan: NewProjectStage(barSource, createSelectExpression(fooSource, "foo", "a", "", "a")),
						},
					),
				)
			})
		})

		Convey("errors", func() {
			testError("select a from idk", `table "idk" doesn't exist in db "test"`)
			testError("select idk from foo", `unknown column "idk"`)
			testError("select f.a from foo", `unknown column "a" in table "f"`)
			testError("select foo.a from foo f", `unknown column "a" in table "foo"`)
			testError("select a + idk from foo", `unknown column "idk"`)

			testError("select *, * from foo", `cannot have a global * in the field list conjunction with any other columns`)
			testError("select a, * from foo", `cannot have a global * in the field list conjunction with any other columns`)
			testError("select *, a from foo", `cannot have a global * in the field list conjunction with any other columns`)

			testError("select a from foo, bar", `column "a" in the field list is ambiguous`)
			testError("select foo.a from (select a from foo)", `unknown column "a" in table "foo"`)
		})

	})
}
