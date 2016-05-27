package evaluator

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewAlgebrizeStatements(t *testing.T) {

	testSchema, _ := schema.New(testSchema1)
	defaultDbName := "test"

	test := func(sql string, expectedPlanFactory func() PlanStage) {
		Convey(sql, func() {
			statement, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			selectStatement := statement.(sqlparser.SelectStatement)
			actual, err := Algebrize(selectStatement, defaultDbName, testSchema)
			So(err, ShouldBeNil)

			expected := expectedPlanFactory()

			//fmt.Printf("\nExpected: %# v", pretty.Formatter(expected))
			//fmt.Printf("\nActual: %# v", pretty.Formatter(actual))

			So(actual, ShouldResemble, expected)
		})
	}

	testError := func(sql, message string) {
		Convey(sql, func() {
			statement, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			selectStatement := statement.(sqlparser.SelectStatement)
			actual, err := Algebrize(selectStatement, defaultDbName, testSchema)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, fmt.Errorf(message))
			So(actual, ShouldBeNil)
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
		for _, c := range source.Columns() {
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
				SQLType:   column.SQLType,
				MongoType: column.MongoType,
			},
			Expr: createSQLColumnExpr(column.Table, column.Name, column.SQLType, column.MongoType),
		}
	}

	createSelectExpression := func(source PlanStage, sourceTableName, sourceColumnName, projectedTableName, projectedColumnName string) SelectExpression {
		for _, c := range source.Columns() {
			if c.Table == sourceTableName && c.Name == sourceColumnName {
				return createSelectExpressionFromColumn(c, projectedTableName, projectedColumnName)
			}
		}

		panic(fmt.Sprintf("no column found with the name %q", sourceColumnName))
	}

	createAllSelectExpressionsFromSource := func(source PlanStage, projectedTableName string) SelectExpressions {
		results := SelectExpressions{}
		for _, c := range source.Columns() {
			results = append(results, createSelectExpressionFromColumn(c, projectedTableName, c.Name))
		}

		return results
	}

	createSelectExpressionFromSQLExpr := func(tableName, columnName string, expr SQLExpr) SelectExpression {
		column := &Column{
			Table:   tableName,
			Name:    columnName,
			SQLType: expr.Type(),
		}

		if sqlColExpr, ok := expr.(SQLColumnExpr); ok {
			column.MongoType = sqlColExpr.columnType.MongoType
		}

		return SelectExpression{Column: column, Expr: expr}
	}

	createMongoSource := func(tableName, aliasName string) PlanStage {
		r, _ := NewMongoSourceStage(testSchema, defaultDbName, tableName, aliasName)
		return r
	}

	Convey("Subject: Algebrize Statements", t, func() {
		Convey("dual queries", func() {
			test("select 2 + 3", func() PlanStage {
				return NewProjectStage(
					NewDualStage(),
					createSelectExpressionFromSQLExpr("", "2+3", &SQLAddExpr{
						left:  SQLInt(2),
						right: SQLInt(3),
					}),
				)
			})
		})

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
			test("select a from (select a from foo) f", func() PlanStage {
				source := createMongoSource("foo", "foo")
				subquery := NewProjectStage(source, createSelectExpression(source, "foo", "a", "f", "a"))
				return NewProjectStage(subquery, createSelectExpression(subquery, "f", "a", "", "a"))
			})

			test("select f.a from (select a from foo) f", func() PlanStage {
				source := createMongoSource("foo", "foo")
				subquery := NewProjectStage(source, createSelectExpression(source, "foo", "a", "f", "a"))
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
					NewFilterStage(source, createSQLColumnExpr("foo", "a", schema.SQLInt, schema.MongoInt)),
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
						[]SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
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
						[]SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
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
							[]SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
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
							[]SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
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
							[]SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
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
						[]SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
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
						[]SQLExpr{createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)},
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
						[]SQLExpr{createSQLColumnExpr("", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)},
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
			testError("select a", `unknown column "a"`)

			testError("select a from idk", `table "idk" doesn't exist in db "test"`)
			testError("select idk from foo", `unknown column "idk"`)
			testError("select f.a from foo", `unknown column "a" in table "f"`)
			testError("select foo.a from foo f", `unknown column "a" in table "foo"`)
			testError("select a + idk from foo", `unknown column "idk"`)

			testError("select *, * from foo", `cannot have a global * in the field list in conjunction with any other columns`)
			testError("select a, * from foo", `cannot have a global * in the field list in conjunction with any other columns`)
			testError("select *, a from foo", `cannot have a global * in the field list in conjunction with any other columns`)

			testError("select a from foo, bar", `column "a" in the field list is ambiguous`)
			testError("select foo.a from (select a from foo)", `every derived table must have it's own alias`)
			testError("select foo.a from foo f, bar b", `unknown column "a" in table "foo"`)
			testError("select f.a, * from foo f, bar b", `cannot have a global * in the field list in conjunction with any other columns`)
			testError("select a from foo f, bar b", `column "a" in the field list is ambiguous`)

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

func TestNewAlgebrizeExpr(t *testing.T) {
	testSchema, _ := schema.New(testSchema1)
	source, _ := NewMongoSourceStage(testSchema, "test", "foo", "foo")

	test := func(sql string, expected SQLExpr) {
		Convey(sql, func() {
			statement, err := sqlparser.Parse("select " + sql + " from foo")
			So(err, ShouldBeNil)

			selectStatement := statement.(*sqlparser.Select)
			actualPlan, err := Algebrize(selectStatement, "test", testSchema)
			So(err, ShouldBeNil)

			actual := (actualPlan.(*ProjectStage)).projectedColumns[0].Expr

			So(actual, ShouldResemble, expected)
		})
	}

	createSQLColumnExpr := func(columnName string) SQLColumnExpr {
		for _, c := range source.Columns() {
			if c.Name == columnName {
				return SQLColumnExpr{
					tableName:  c.Table,
					columnName: c.Name,
					columnType: schema.ColumnType{
						SQLType:   c.SQLType,
						MongoType: c.MongoType,
					},
				}
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

		Convey("Is Null", func() {
			test("a IS NULL", &SQLNullCmpExpr{
				createSQLColumnExpr("a"),
			})
		})

		Convey("Is Not Null", func() {
			test("a IS NOT NULL", &SQLNotExpr{
				&SQLNullCmpExpr{
					createSQLColumnExpr("a"),
				},
			})
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

		SkipConvey("Not In", func() {
		})

		Convey("Null", func() {
			test("NULL", SQLNull)
		})

		Convey("Number", func() {
			test("20", SQLInt(20))
			test("-20", SQLInt(-20))
			test("20.2", SQLFloat(20.2))
			test("-20.2", SQLFloat(-20.2))
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
			// TODO: this seems weird... not sure what is up here
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

		Convey("Tuple", func() {
			test("(a, 1)", &SQLTupleExpr{
				Exprs: []SQLExpr{createSQLColumnExpr("a"), SQLInt(1)},
			})

			test("(a)", createSQLColumnExpr("a"))
		})

		Convey("Unary Minus", func() {
			test("-a", &SQLUnaryMinusExpr{createSQLColumnExpr("a")})
		})

		Convey("Unary Tilde", func() {
			test("~a", &SQLUnaryTildeExpr{createSQLColumnExpr("a")})
		})

		Convey("Varchar", func() {
			test("'a'", SQLVarchar("a"))
		})
	})
}
