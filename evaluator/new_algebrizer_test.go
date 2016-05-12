package evaluator

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	"github.com/kr/pretty"
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

			fmt.Printf("\nExpected: %# v", pretty.Formatter(expectedPlan))
			fmt.Printf("\nActual: %# v", pretty.Formatter(actualPlan))

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

	createSelectExpression := func(tableName, columnName string, expr SQLExpr) SelectExpression {
		column := &Column{
			Table:     tableName,
			Name:      columnName,
			View:      columnName, // ???
			SQLType:   expr.Type(),
			MongoType: schema.MongoNone,
		}

		return SelectExpression{Column: column, Expr: expr}
	}

	Convey("Subject: Algebrize", t, func() {

		Convey("simple queries", func() {
			test("select a from foo", func() PlanStage {
				source, _ := NewMongoSourceStage(testSchema, defaultDbName, "foo", "foo")
				return &ProjectStage{
					source: source,
					sExprs: SelectExpressions{
						createSelectExpression("", "a",
							SQLColumnExpr{
								tableName:  "foo",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoInt}}),
					},
				}
			})

			test("select a from foo f", func() PlanStage {
				source, _ := NewMongoSourceStage(testSchema, defaultDbName, "foo", "f")
				return &ProjectStage{
					source: source,
					sExprs: SelectExpressions{
						createSelectExpression("", "a",
							SQLColumnExpr{
								tableName:  "f",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoInt}}),
					},
				}
			})

			test("select f.a from foo f", func() PlanStage {
				source, _ := NewMongoSourceStage(testSchema, defaultDbName, "foo", "f")
				return &ProjectStage{
					source: source,
					sExprs: SelectExpressions{
						createSelectExpression("", "a",
							SQLColumnExpr{
								tableName:  "f",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoInt}}),
					},
				}
			})

			test("select a as b from foo", func() PlanStage {
				source, _ := NewMongoSourceStage(testSchema, defaultDbName, "foo", "foo")
				return &ProjectStage{
					source: source,
					sExprs: SelectExpressions{
						createSelectExpression("", "b",
							SQLColumnExpr{
								tableName:  "foo",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoInt}}),
					},
				}
			})

			test("select a + 2 from foo", func() PlanStage {
				source, _ := NewMongoSourceStage(testSchema, defaultDbName, "foo", "foo")
				return &ProjectStage{
					source: source,
					sExprs: SelectExpressions{
						createSelectExpression("", "a+2",
							&SQLAddExpr{
								left: SQLColumnExpr{
									tableName:  "foo",
									columnName: "a",
									columnType: schema.ColumnType{
										SQLType:   schema.SQLInt,
										MongoType: schema.MongoInt}},
								right: SQLInt(2)}),
					},
				}
			})

			test("select a + 2 as b from foo", func() PlanStage {
				source, _ := NewMongoSourceStage(testSchema, defaultDbName, "foo", "foo")
				return &ProjectStage{
					source: source,
					sExprs: SelectExpressions{
						createSelectExpression("", "b",
							&SQLAddExpr{
								left: SQLColumnExpr{
									tableName:  "foo",
									columnName: "a",
									columnType: schema.ColumnType{
										SQLType:   schema.SQLInt,
										MongoType: schema.MongoInt}},
								right: SQLInt(2)}),
					},
				}
			})
		})

		Convey("joins", func() {
			test("select foo.a, bar.a from foo, bar", func() PlanStage {
				fooSource, _ := NewMongoSourceStage(testSchema, defaultDbName, "foo", "foo")
				barSource, _ := NewMongoSourceStage(testSchema, defaultDbName, "bar", "bar")
				return &ProjectStage{
					source: &JoinStage{
						left:  fooSource,
						right: barSource,
						kind:  CrossJoin,
					},
					sExprs: SelectExpressions{
						createSelectExpression("", "a",
							SQLColumnExpr{
								tableName:  "foo",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoInt}}),
						createSelectExpression("", "a",
							SQLColumnExpr{
								tableName:  "bar",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoInt}}),
					},
				}
			})

			test("select f.a, bar.a from foo f, bar", func() PlanStage {
				fooSource, _ := NewMongoSourceStage(testSchema, defaultDbName, "foo", "f")
				barSource, _ := NewMongoSourceStage(testSchema, defaultDbName, "bar", "bar")
				return &ProjectStage{
					source: &JoinStage{
						left:  fooSource,
						right: barSource,
						kind:  CrossJoin,
					},
					sExprs: SelectExpressions{
						createSelectExpression("", "a",
							SQLColumnExpr{
								tableName:  "f",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoInt}}),
						createSelectExpression("", "a",
							SQLColumnExpr{
								tableName:  "bar",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoInt}}),
					},
				}
			})

			test("select f.a, b.a from foo f, bar b", func() PlanStage {
				fooSource, _ := NewMongoSourceStage(testSchema, defaultDbName, "foo", "f")
				barSource, _ := NewMongoSourceStage(testSchema, defaultDbName, "bar", "b")
				return &ProjectStage{
					source: &JoinStage{
						left:  fooSource,
						right: barSource,
						kind:  CrossJoin,
					},
					sExprs: SelectExpressions{
						createSelectExpression("", "a",
							SQLColumnExpr{
								tableName:  "f",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoInt}}),
						createSelectExpression("", "a",
							SQLColumnExpr{
								tableName:  "b",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoInt}}),
					},
				}
			})
		})

		Convey("subqueries as sources", func() {
			test("select a from (select a from foo)", func() PlanStage {
				source, _ := NewMongoSourceStage(testSchema, defaultDbName, "foo", "foo")
				return &ProjectStage{
					source: &SubqueryStage{
						tableName: "",
						source: &ProjectStage{
							source: source,
							sExprs: SelectExpressions{
								createSelectExpression("", "a",
									SQLColumnExpr{
										tableName:  "foo",
										columnName: "a",
										columnType: schema.ColumnType{
											SQLType:   schema.SQLInt,
											MongoType: schema.MongoInt}}),
							},
						},
					},
					sExprs: SelectExpressions{
						createSelectExpression("", "a",
							SQLColumnExpr{
								tableName:  "",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoNone}}),
					},
				}
			})

			test("select a from (select a from foo) f", func() PlanStage {
				source, _ := NewMongoSourceStage(testSchema, defaultDbName, "foo", "foo")
				return &ProjectStage{
					source: &SubqueryStage{
						tableName: "f",
						source: &ProjectStage{
							source: source,
							sExprs: SelectExpressions{
								createSelectExpression("f", "a",
									SQLColumnExpr{
										tableName:  "foo",
										columnName: "a",
										columnType: schema.ColumnType{
											SQLType:   schema.SQLInt,
											MongoType: schema.MongoInt}}),
							},
						},
					},
					sExprs: SelectExpressions{
						createSelectExpression("", "a",
							SQLColumnExpr{
								tableName:  "f",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoNone}}),
					},
				}
			})

			test("select f.a from (select a from foo) f", func() PlanStage {
				source, _ := NewMongoSourceStage(testSchema, defaultDbName, "foo", "foo")
				return &ProjectStage{
					source: &SubqueryStage{
						tableName: "f",
						source: &ProjectStage{
							source: source,
							sExprs: SelectExpressions{
								createSelectExpression("f", "a",
									SQLColumnExpr{
										tableName:  "foo",
										columnName: "a",
										columnType: schema.ColumnType{
											SQLType:   schema.SQLInt,
											MongoType: schema.MongoInt}}),
							},
						},
					},
					sExprs: SelectExpressions{
						createSelectExpression("", "a",
							SQLColumnExpr{
								tableName:  "f",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoNone}}),
					},
				}
			})
		})

		Convey("errors", func() {
			testError("select a from idk", `table "idk" doesn't exist in db "test"`)
			testError("select idk from foo", `unknown column "idk"`)
			testError("select f.a from foo", `unknown column "a" in table "f"`)
			testError("select foo.a from foo f", `unknown column "a" in table "foo"`)
			testError("select a + idk from foo", `unknown column "idk"`)

			testError("select a from foo, bar", `column "a" in the field list is ambiguous`)
		})

	})
}
