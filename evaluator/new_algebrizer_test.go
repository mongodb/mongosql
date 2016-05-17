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

	createSQLColumnExpr := func(tableName, columnName string, sqlType schema.SQLType, mongoType schema.MongoType) SQLColumnExpr {
		return SQLColumnExpr{
			tableName:  tableName,
			columnName: columnName,
			columnType: schema.ColumnType{
				SQLType:   sqlType,
				MongoType: mongoType,
			},
		}
	}

	createSQLColumnExprFromSource := func(source PlanStage, tableName, columnName string) SQLColumnExpr {
		for _, c := range source.OpFields() {
			if c.Table == tableName && c.Name == columnName {
				return createSQLColumnExpr(c.Table, c.Name, c.SQLType, c.MongoType)
			}
		}

		panic("column not found")
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
			Expr: createSQLColumnExpr(column.Table, column.Name, column.SQLType, column.MongoType),
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
				join := NewJoinStage(CrossJoin, fooSource, barSource, nil)
				return NewProjectStage(join, createAllSelectExpressionsFromSource(join, "")...)
			})

			test("select foo.*, bar.* from foo, bar", func() PlanStage {
				fooSource := createMongoSource("foo", "foo")
				barSource := createMongoSource("bar", "bar")
				join := NewJoinStage(CrossJoin, fooSource, barSource, nil)
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

			test("select ASCII(a) from foo", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(source,
					createSelectExpressionFromSQLExpr("", "ascii(a)",
						&SQLScalarFunctionExpr{
							Name: "ascii",
							Exprs: []SQLExpr{SQLColumnExpr{
								tableName:  "foo",
								columnName: "a",
								columnType: schema.ColumnType{
									SQLType:   schema.SQLInt,
									MongoType: schema.MongoInt}},
							},
						},
					),
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

			test("select a from (select a from foo limit 1)", func() PlanStage {
				source := createMongoSource("foo", "foo")
				subquery := &SubqueryStage{
					tableName: "",
					source:    NewProjectStage(NewLimitStage(source, 0, 1), createSelectExpression(source, "foo", "a", "", "a")),
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
				join := NewJoinStage(CrossJoin, fooSource, barSource, nil)
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
							plan:       NewProjectStage(barSource, createSelectExpression(fooSource, "foo", "a", "", "a")),
							correlated: true,
						},
					),
				)
			})
		})

		Convey("joins", func() {
			test("select foo.a, bar.a from foo, bar", func() PlanStage {
				fooSource := createMongoSource("foo", "foo")
				barSource := createMongoSource("bar", "bar")
				join := NewJoinStage(CrossJoin, fooSource, barSource, nil)
				return NewProjectStage(join,
					createSelectExpression(join, "foo", "a", "", "a"),
					createSelectExpression(join, "bar", "a", "", "a"),
				)
			})

			test("select f.a, bar.a from foo f, bar", func() PlanStage {
				fooSource := createMongoSource("foo", "f")
				barSource := createMongoSource("bar", "bar")
				join := NewJoinStage(CrossJoin, fooSource, barSource, nil)
				return NewProjectStage(join,
					createSelectExpression(join, "f", "a", "", "a"),
					createSelectExpression(join, "bar", "a", "", "a"),
				)
			})

			test("select f.a, b.a from foo f, bar b", func() PlanStage {
				fooSource := createMongoSource("foo", "f")
				barSource := createMongoSource("bar", "b")
				join := NewJoinStage(CrossJoin, fooSource, barSource, nil)
				return NewProjectStage(join,
					createSelectExpression(join, "f", "a", "", "a"),
					createSelectExpression(join, "b", "a", "", "a"),
				)
			})

			test("select foo.a, bar.a from foo inner join bar on foo.b = bar.b", func() PlanStage {
				fooSource := createMongoSource("foo", "foo")
				barSource := createMongoSource("bar", "bar")
				join := NewJoinStage(InnerJoin, fooSource, barSource,
					&SQLEqualsExpr{
						left:  createSQLColumnExprFromSource(fooSource, "foo", "b"),
						right: createSQLColumnExprFromSource(barSource, "bar", "b"),
					})
				return NewProjectStage(join,
					createSelectExpression(join, "foo", "a", "", "a"),
					createSelectExpression(join, "bar", "a", "", "a"),
				)
			})

			test("select foo.a, bar.a from foo join bar on foo.b = bar.b", func() PlanStage {
				fooSource := createMongoSource("foo", "foo")
				barSource := createMongoSource("bar", "bar")
				join := NewJoinStage(InnerJoin, fooSource, barSource,
					&SQLEqualsExpr{
						left:  createSQLColumnExprFromSource(fooSource, "foo", "b"),
						right: createSQLColumnExprFromSource(barSource, "bar", "b"),
					})
				return NewProjectStage(join,
					createSelectExpression(join, "foo", "a", "", "a"),
					createSelectExpression(join, "bar", "a", "", "a"),
				)
			})

			test("select foo.a, bar.a from foo left outer join bar on foo.b = bar.b", func() PlanStage {
				fooSource := createMongoSource("foo", "foo")
				barSource := createMongoSource("bar", "bar")
				join := NewJoinStage(LeftJoin, fooSource, barSource,
					&SQLEqualsExpr{
						left:  createSQLColumnExprFromSource(fooSource, "foo", "b"),
						right: createSQLColumnExprFromSource(barSource, "bar", "b"),
					})
				return NewProjectStage(join,
					createSelectExpression(join, "foo", "a", "", "a"),
					createSelectExpression(join, "bar", "a", "", "a"),
				)
			})

			test("select foo.a, bar.a from foo right outer join bar on foo.b = bar.b", func() PlanStage {
				fooSource := createMongoSource("foo", "foo")
				barSource := createMongoSource("bar", "bar")
				join := NewJoinStage(RightJoin, fooSource, barSource,
					&SQLEqualsExpr{
						left:  createSQLColumnExprFromSource(fooSource, "foo", "b"),
						right: createSQLColumnExprFromSource(barSource, "bar", "b"),
					})
				return NewProjectStage(join,
					createSelectExpression(join, "foo", "a", "", "a"),
					createSelectExpression(join, "bar", "a", "", "a"),
				)
			})
		})

		Convey("where", func() {
			test("select a from foo where a", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewFilterStage(source,
						&SQLConvertExpr{
							expr:     createSQLColumnExpr("foo", "a", schema.SQLInt, schema.MongoInt),
							convType: schema.SQLBoolean,
						},
					),
					createSelectExpression(source, "foo", "a", "", "a"),
				)
			})

			test("select a from foo where a > 10", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewFilterStage(source,
						&SQLGreaterThanExpr{
							left:  createSQLColumnExpr("foo", "a", schema.SQLInt, schema.MongoInt),
							right: SQLInt(10),
						},
					),
					createSelectExpression(source, "foo", "a", "", "a"),
				)
			})

			test("select a as b from foo where b > 10", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewFilterStage(source,
						&SQLGreaterThanExpr{
							left:  createSQLColumnExpr("foo", "b", schema.SQLInt, schema.MongoInt),
							right: SQLInt(10),
						},
					),
					createSelectExpression(source, "foo", "a", "", "b"),
				)
			})
		})

		Convey("group by", func() {
			test("select sum(a) from foo", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewGroupByStage(source,
						nil,
						SelectExpressions{
							createSelectExpressionFromSQLExpr("", "sum(foo.a)", &SQLAggFunctionExpr{
								Name:  "sum",
								Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
							}),
						},
					),
					createSelectExpressionFromSQLExpr("", "sum(a)", createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
				)
			})

			test("select sum(a) from foo group by b", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewGroupByStage(source,
						SelectExpressions{
							createSelectExpression(source, "foo", "b", "foo", "b"),
						},
						SelectExpressions{
							createSelectExpressionFromSQLExpr("", "sum(foo.a)", &SQLAggFunctionExpr{
								Name:  "sum",
								Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
							}),
						},
					),
					createSelectExpressionFromSQLExpr("", "sum(a)", createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
				)
			})

			test("select a, sum(a) from foo group by b", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewGroupByStage(source,
						SelectExpressions{
							createSelectExpression(source, "foo", "b", "foo", "b"),
						},
						SelectExpressions{
							createSelectExpression(source, "foo", "a", "foo", "a"),
							createSelectExpressionFromSQLExpr("", "sum(foo.a)", &SQLAggFunctionExpr{
								Name:  "sum",
								Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
							}),
						},
					),
					createSelectExpression(source, "foo", "a", "", "a"),
					createSelectExpressionFromSQLExpr("", "sum(a)", createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
				)
			})

			test("select sum(a) from foo group by b order by sum(a)", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewOrderByStage(
						NewGroupByStage(source,
							SelectExpressions{
								createSelectExpression(source, "foo", "b", "foo", "b"),
							},
							SelectExpressions{
								createSelectExpressionFromSQLExpr("", "sum(foo.a)", &SQLAggFunctionExpr{
									Name:  "sum",
									Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
								}),
							},
						),
						&orderByTerm{
							expr:      createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone),
							ascending: true,
						},
					),
					createSelectExpressionFromSQLExpr("", "sum(a)", createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
				)
			})

			test("select sum(a) as sum_a from foo group by b order by sum_a", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewOrderByStage(
						NewGroupByStage(source,
							SelectExpressions{
								createSelectExpression(source, "foo", "b", "foo", "b"),
							},
							SelectExpressions{
								createSelectExpressionFromSQLExpr("", "sum(foo.a)", &SQLAggFunctionExpr{
									Name:  "sum",
									Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
								}),
							},
						),
						&orderByTerm{
							expr:      createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone),
							ascending: true,
						},
					),
					createSelectExpressionFromSQLExpr("", "sum_a", createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
				)
			})

			// We have an issue with subqueries and aggregates...
			// test("select (select sum(foo.a) from foo f) from foo group by b", func() PlanStage {
			// 	source := createMongoSource("foo", "foo")
			// 	return NewProjectStage(
			// 		NewGroupByStage(source,
			// 			SelectExpressions{
			// 				createSelectExpression(source, "foo", "b", "foo", "b"),
			// 			},
			// 			SelectExpressions{
			// 				createSelectExpressionFromSQLExpr("", "sum(foo.a)", &SQLAggFunctionExpr{
			// 					Name:  "sum",
			// 					Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
			// 				}),
			// 			},
			// 		),
			// 		createSelectExpressionFromSQLExpr("", "(select sum(foo.a) from foo f)",
			// 			&SQLSubqueryExpr{
			// 				// correlated: true,
			// 				plan: NewProjectStage(
			// 					source,
			// 					createSelectExpressionFromSQLExpr("", "sum(foo.a)", createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
			// 				),
			// 			},
			// 		),
			// 	)
			// })
		})

		Convey("having", func() {
			test("select a from foo group by b having sum(a) > 10", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewFilterStage(
						NewGroupByStage(source,
							SelectExpressions{
								createSelectExpression(source, "foo", "b", "foo", "b"),
							},
							SelectExpressions{
								createSelectExpression(source, "foo", "a", "foo", "a"),
								createSelectExpressionFromSQLExpr("", "sum(foo.a)", &SQLAggFunctionExpr{
									Name:  "sum",
									Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
								}),
							},
						),
						&SQLGreaterThanExpr{
							left:  createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone),
							right: SQLInt(10),
						},
					),
					createSelectExpression(source, "foo", "a", "", "a"),
				)
			})
		})

		Convey("distinct", func() {
			test("select distinct a from foo", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewGroupByStage(source,
						SelectExpressions{
							createSelectExpression(source, "foo", "a", "foo", "a"),
						},
						SelectExpressions{
							createSelectExpression(source, "foo", "a", "foo", "a"),
						},
					),
					createSelectExpression(source, "foo", "a", "", "a"),
				)
			})

			test("select distinct sum(a) from foo", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewGroupByStage(
						NewGroupByStage(source,
							nil,
							SelectExpressions{
								createSelectExpressionFromSQLExpr("", "sum(foo.a)", &SQLAggFunctionExpr{
									Name:  "sum",
									Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
								}),
							},
						),
						SelectExpressions{
							createSelectExpressionFromSQLExpr("", "sum(foo.a)", createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
						},
						SelectExpressions{
							createSelectExpressionFromSQLExpr("", "sum(foo.a)", createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
						},
					),
					createSelectExpressionFromSQLExpr("", "sum(a)", createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
				)
			})

			test("select distinct sum(a) from foo having sum(a) > 20", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewGroupByStage(
						NewFilterStage(
							NewGroupByStage(source,
								nil,
								SelectExpressions{
									createSelectExpressionFromSQLExpr("", "sum(foo.a)", &SQLAggFunctionExpr{
										Name:  "sum",
										Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
									}),
								},
							),
							&SQLGreaterThanExpr{
								left:  createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone),
								right: SQLInt(20),
							},
						),
						SelectExpressions{
							createSelectExpressionFromSQLExpr("", "sum(foo.a)", createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
						},
						SelectExpressions{
							createSelectExpressionFromSQLExpr("", "sum(foo.a)", createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
						},
					),
					createSelectExpressionFromSQLExpr("", "sum(a)", createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
				)
			})
		})

		Convey("order by", func() {
			test("select a from foo order by a", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewOrderByStage(source,
						&orderByTerm{
							expr:      createSQLColumnExpr("foo", "a", schema.SQLInt, schema.MongoInt),
							ascending: true,
						},
					),
					createSelectExpression(source, "foo", "a", "", "a"),
				)
			})

			test("select a as b from foo order by b", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewOrderByStage(source,
						&orderByTerm{
							expr:      createSQLColumnExpr("foo", "a", schema.SQLInt, schema.MongoInt),
							ascending: true,
						},
					),
					createSelectExpression(source, "foo", "a", "", "b"),
				)
			})

			test("select a from foo order by foo.a", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewOrderByStage(source,
						&orderByTerm{
							expr:      createSQLColumnExpr("foo", "a", schema.SQLInt, schema.MongoInt),
							ascending: true,
						},
					),
					createSelectExpression(source, "foo", "a", "", "a"),
				)
			})

			test("select a as b from foo order by foo.a", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewOrderByStage(source,
						&orderByTerm{
							expr:      createSQLColumnExpr("foo", "a", schema.SQLInt, schema.MongoInt),
							ascending: true,
						},
					),
					createSelectExpression(source, "foo", "a", "", "b"),
				)
			})

			test("select a from foo order by 1", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewOrderByStage(source,
						&orderByTerm{
							expr:      createSQLColumnExpr("foo", "a", schema.SQLInt, schema.MongoInt),
							ascending: true,
						},
					),
					createSelectExpression(source, "foo", "a", "", "a"),
				)
			})

			test("select * from foo order by 2", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewOrderByStage(source,
						&orderByTerm{
							expr:      createSQLColumnExpr("foo", "b", schema.SQLInt, schema.MongoInt),
							ascending: true,
						},
					),
					createAllSelectExpressionsFromSource(source, "")...,
				)
			})

			test("select foo.* from foo order by 2", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewOrderByStage(source,
						&orderByTerm{
							expr:      createSQLColumnExpr("foo", "b", schema.SQLInt, schema.MongoInt),
							ascending: true,
						},
					),
					createAllSelectExpressionsFromSource(source, "")...,
				)
			})

			test("select foo.*, foo.a from foo order by 2", func() PlanStage {
				source := createMongoSource("foo", "foo")
				columns := append(createAllSelectExpressionsFromSource(source, ""), createSelectExpression(source, "foo", "a", "", "a"))
				return NewProjectStage(
					NewOrderByStage(source,
						&orderByTerm{
							expr:      createSQLColumnExpr("foo", "b", schema.SQLInt, schema.MongoInt),
							ascending: true,
						},
					),
					columns...,
				)
			})

			test("select a from foo order by -1", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewOrderByStage(source,
						&orderByTerm{
							expr:      SQLInt(-1),
							ascending: true,
						},
					),
					createSelectExpression(source, "foo", "a", "", "a"),
				)
			})

			test("select a + b as c from foo order by c - b", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewOrderByStage(source,
						&orderByTerm{
							expr: &SQLSubtractExpr{
								left: &SQLAddExpr{
									left:  createSQLColumnExpr("foo", "a", schema.SQLInt, schema.MongoInt),
									right: createSQLColumnExpr("foo", "b", schema.SQLInt, schema.MongoInt),
								},
								right: createSQLColumnExpr("foo", "b", schema.SQLInt, schema.MongoInt),
							},
							ascending: true,
						},
					),
					createSelectExpressionFromSQLExpr("", "c",
						&SQLAddExpr{
							left:  createSQLColumnExpr("foo", "a", schema.SQLInt, schema.MongoInt),
							right: createSQLColumnExpr("foo", "b", schema.SQLInt, schema.MongoInt),
						},
					),
				)
			})
		})

		Convey("limit", func() {
			test("select a from foo limit 10", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewLimitStage(source, 0, 10),
					createSelectExpression(source, "foo", "a", "", "a"),
				)
			})

			test("select a from foo limit 10, 20", func() PlanStage {
				source := createMongoSource("foo", "foo")
				return NewProjectStage(
					NewLimitStage(source, 10, 20),
					createSelectExpression(source, "foo", "a", "", "a"),
				)
			})
		})

		Convey("errors", func() {
			testError("select a from idk", `table "idk" doesn't exist in db "test"`)
			testError("select idk from foo", `unknown column "idk"`)
			testError("select f.a from foo", `unknown column "a" in table "f"`)
			testError("select foo.a from foo f", `unknown column "a" in table "foo"`)
			testError("select a + idk from foo", `unknown column "idk"`)

			testError("select *, * from foo", `cannot have a global * in the field list in conjunction with any other columns`)
			testError("select a, * from foo", `cannot have a global * in the field list in conjunction with any other columns`)
			testError("select *, a from foo", `cannot have a global * in the field list in conjunction with any other columns`)

			testError("select a from foo, bar", `column "a" in the field list is ambiguous`)
			testError("select foo.a from (select a from foo)", `unknown column "a" in table "foo"`)

			testError("select a from foo limit -10", `limit rowcount cannot be negative`)
			testError("select a from foo limit -10, 20", `limit offset cannot be negative`)
			testError("select a from foo limit -10, -20", `limit offset cannot be negative`)
			testError("select a from foo limit b", `limit rowcount must be an integer`)
			testError("select a from foo limit 'c'", `limit rowcount must be an integer`)

			testError("select a from foo order by 2", `unknown column "2" in order clause`)
			testError("select a from foo order by idk", `unknown column "idk"`)

			testError("select sum(a) from foo group by sum(a)", `can't group on "sum(foo.a)"`)
			testError("select sum(a) from foo group by 1", `can't group on "sum(foo.a)"`)
			testError("select sum(a) from foo group by 2", `unknown column "2" in group clause`)
		})

	})
}
