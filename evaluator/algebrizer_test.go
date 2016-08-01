package evaluator

import (
	"fmt"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/kr/pretty"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAlgebrizeSelect(t *testing.T) {

	testSchema, err := schema.New(testSchema1)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}
	defaultDbName := "test"

	test := func(sql string, expectedPlanFactory func() PlanStage) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			selectStatement := statement.(parser.SelectStatement)
			actual, err := AlgebrizeSelect(selectStatement, defaultDbName, testSchema)
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

			selectStatement := statement.(parser.SelectStatement)
			actual, err := AlgebrizeSelect(selectStatement, defaultDbName, testSchema)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, message)
			So(actual, ShouldBeNil)
		})
	}

	createMongoSource := func(selectID int, tableName, aliasName string) PlanStage {
		r, _ := NewMongoSourceStage(selectID, testSchema, defaultDbName, tableName, aliasName)
		return r
	}

	createReqCols := func(source []PlanStage, cols []string, selectIDs map[string]int) []SQLExpr {
		var reqCols []SQLExpr
		for _, s := range source {
			for _, c := range s.Columns() {
				if containsString(cols, c.Table+"."+c.Name) {
					col := createProjectedColumnFromColumn(selectIDs[c.Table+"."+c.Name], c, c.Table, c.Name)
					reqCols = append(reqCols, col.Expr)
				}
			}
		}
		return reqCols
	}

	createReqColsStar := func(source PlanStage, tableName string, selectID int) []SQLExpr {
		var reqCols []SQLExpr
		allCols := createAllProjectedColumnsFromSource(selectID, source, tableName)
		for _, c := range allCols {
			reqCols = append(reqCols, c.Expr)
		}
		return reqCols
	}

	Convey("Subject: Algebrize Select Statements", t, func() {
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
					source := createMongoSource(2, "foo", "f")
					subquery := NewProjectStage(source, createProjectedColumnSubquery(2, source, "f", "a", "a"))
					return NewProjectStage(subquery, createProjectedColumn(1, subquery, "f", "a", "", "a"))
				})

				test("select f.a from (select a from foo) f", func() PlanStage {
					source := createMongoSource(2, "foo", "f")
					subquery := NewProjectStage(source, createProjectedColumnSubquery(2, source, "f", "a", "a"))
					return NewProjectStage(subquery, createProjectedColumn(1, subquery, "f", "a", "", "a"))
				})
			})

			Convey("joins", func() {
				test("select foo.a, bar.a from foo, bar", func() PlanStage {
					fooSource := createMongoSource(1, "foo", "foo")
					barSource := createMongoSource(1, "bar", "bar")
					reqCols := createReqCols([]PlanStage{fooSource, barSource}, []string{"foo.a", "bar.a"}, map[string]int{"foo.a": 1, "bar.a": 1})
					join := NewJoinStage(CrossJoin, fooSource, barSource, SQLTrue, reqCols)
					return NewProjectStage(join,
						createProjectedColumn(1, join, "foo", "a", "", "a"),
						createProjectedColumn(1, join, "bar", "a", "", "a"),
					)
				})

				test("select f.a, bar.a from foo f, bar", func() PlanStage {
					fooSource := createMongoSource(1, "foo", "f")
					barSource := createMongoSource(1, "bar", "bar")
					reqCols := createReqCols([]PlanStage{fooSource, barSource}, []string{"f.a", "bar.a"}, map[string]int{"f.a": 1, "bar.a": 1})
					join := NewJoinStage(CrossJoin, fooSource, barSource, SQLTrue, reqCols)
					return NewProjectStage(join,
						createProjectedColumn(1, join, "f", "a", "", "a"),
						createProjectedColumn(1, join, "bar", "a", "", "a"),
					)
				})

				test("select f.a, b.a from foo f, bar b", func() PlanStage {
					fooSource := createMongoSource(1, "foo", "f")
					barSource := createMongoSource(1, "bar", "b")
					reqCols := createReqCols([]PlanStage{fooSource, barSource}, []string{"f.a", "b.a"}, map[string]int{"f.a": 1, "b.a": 1})
					join := NewJoinStage(CrossJoin, fooSource, barSource, SQLTrue, reqCols)
					return NewProjectStage(join,
						createProjectedColumn(1, join, "f", "a", "", "a"),
						createProjectedColumn(1, join, "b", "a", "", "a"),
					)
				})

				test("select foo.a, bar.a from foo inner join bar on foo.b = bar.b", func() PlanStage {
					fooSource := createMongoSource(1, "foo", "foo")
					barSource := createMongoSource(1, "bar", "bar")
					reqColsJoin := createReqCols([]PlanStage{fooSource, barSource}, []string{"foo.b", "bar.b"}, map[string]int{"foo.b": 1, "bar.b": 1})
					reqColsSelect := createReqCols([]PlanStage{fooSource, barSource}, []string{"foo.a", "bar.a"}, map[string]int{"foo.a": 1, "bar.a": 1})
					join := NewJoinStage(InnerJoin, fooSource, barSource,
						&SQLEqualsExpr{
							left:  createSQLColumnExprFromSource(fooSource, "foo", "b"),
							right: createSQLColumnExprFromSource(barSource, "bar", "b"),
						}, append(reqColsJoin, reqColsSelect...))
					return NewProjectStage(join,
						createProjectedColumn(1, join, "foo", "a", "", "a"),
						createProjectedColumn(1, join, "bar", "a", "", "a"),
					)
				})

				test("select foo.a, bar.a from foo join bar on foo.b = bar.b", func() PlanStage {
					fooSource := createMongoSource(1, "foo", "foo")
					barSource := createMongoSource(1, "bar", "bar")
					reqColsJoin := createReqCols([]PlanStage{fooSource, barSource}, []string{"foo.b", "bar.b"}, map[string]int{"foo.b": 1, "bar.b": 1})
					reqColsSelect := createReqCols([]PlanStage{fooSource, barSource}, []string{"foo.a", "bar.a"}, map[string]int{"foo.a": 1, "bar.a": 1})
					join := NewJoinStage(InnerJoin, fooSource, barSource,
						&SQLEqualsExpr{
							left:  createSQLColumnExprFromSource(fooSource, "foo", "b"),
							right: createSQLColumnExprFromSource(barSource, "bar", "b"),
						}, append(reqColsJoin, reqColsSelect...))
					return NewProjectStage(join,
						createProjectedColumn(1, join, "foo", "a", "", "a"),
						createProjectedColumn(1, join, "bar", "a", "", "a"),
					)
				})

				test("select foo.a, bar.a from foo left outer join bar on foo.b = bar.b", func() PlanStage {
					fooSource := createMongoSource(1, "foo", "foo")
					barSource := createMongoSource(1, "bar", "bar")
					reqColsJoin := createReqCols([]PlanStage{fooSource, barSource}, []string{"foo.b", "bar.b"}, map[string]int{"foo.b": 1, "bar.b": 1})
					reqColsSelect := createReqCols([]PlanStage{fooSource, barSource}, []string{"foo.a", "bar.a"}, map[string]int{"foo.a": 1, "bar.a": 1})
					join := NewJoinStage(LeftJoin, fooSource, barSource,
						&SQLEqualsExpr{
							left:  createSQLColumnExprFromSource(fooSource, "foo", "b"),
							right: createSQLColumnExprFromSource(barSource, "bar", "b"),
						}, append(reqColsJoin, reqColsSelect...))
					return NewProjectStage(join,
						createProjectedColumn(1, join, "foo", "a", "", "a"),
						createProjectedColumn(1, join, "bar", "a", "", "a"),
					)
				})

				test("select foo.a, bar.a from foo right outer join bar on foo.b = bar.b", func() PlanStage {
					fooSource := createMongoSource(1, "foo", "foo")
					barSource := createMongoSource(1, "bar", "bar")
					reqColsJoin := createReqCols([]PlanStage{fooSource, barSource}, []string{"foo.b", "bar.b"}, map[string]int{"foo.b": 1, "bar.b": 1})
					reqColsSelect := createReqCols([]PlanStage{fooSource, barSource}, []string{"foo.a", "bar.a"}, map[string]int{"foo.a": 1, "bar.a": 1})
					join := NewJoinStage(RightJoin, fooSource, barSource,
						&SQLEqualsExpr{
							left:  createSQLColumnExprFromSource(fooSource, "foo", "b"),
							right: createSQLColumnExprFromSource(barSource, "bar", "b"),
						}, append(reqColsJoin, reqColsSelect...))
					return NewProjectStage(join,
						createProjectedColumn(1, join, "foo", "a", "", "a"),
						createProjectedColumn(1, join, "bar", "a", "", "a"),
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
					fooCols := createReqColsStar(fooSource, "foo", 1)
					barCols := createReqColsStar(barSource, "bar", 1)
					join := NewJoinStage(CrossJoin, fooSource, barSource, SQLTrue, append(fooCols, barCols...))
					return NewProjectStage(join, createAllProjectedColumnsFromSource(1, join, "")...)
				})

				test("select foo.*, bar.* from foo, bar", func() PlanStage {
					fooSource := createMongoSource(1, "foo", "foo")
					barSource := createMongoSource(1, "bar", "bar")
					fooCols := createReqColsStar(fooSource, "foo", 1)
					barCols := createReqColsStar(barSource, "bar", 1)
					join := NewJoinStage(CrossJoin, fooSource, barSource, SQLTrue, append(fooCols, barCols...))
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
									plan: NewCacheStage(2,
										NewProjectStage(barSource, createProjectedColumn(2, barSource, "bar", "a", "", "a")),
									),
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
									plan: NewCacheStage(2,
										NewProjectStage(barSource, createProjectedColumn(2, barSource, "bar", "a", "", "a")),
									),
								},
							),
						)
					})

					test("select a, (select foo.a from foo, bar) from foo", func() PlanStage {
						foo1Source := createMongoSource(1, "foo", "foo")
						foo2Source := createMongoSource(2, "foo", "foo")
						barSource := createMongoSource(2, "bar", "bar")
						reqCols := createReqCols([]PlanStage{foo2Source}, []string{"foo.a"}, map[string]int{"foo.a": 2})
						join := NewJoinStage(CrossJoin, foo2Source, barSource, SQLTrue, reqCols)
						return NewProjectStage(foo1Source,
							createProjectedColumn(1, foo1Source, "foo", "a", "", "a"),
							createProjectedColumnFromSQLExpr(1, "", "(select foo.a from foo, bar)",
								&SQLSubqueryExpr{
									plan: NewCacheStage(2,
										NewProjectStage(join, createProjectedColumn(2, join, "foo", "a", "", "a")),
									),
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
										plan: NewCacheStage(2,
											NewProjectStage(
												barSource,
												createProjectedColumnFromSQLExpr(2, "", "1", SQLInt(1)),
											),
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
				})
			})

		})

		Convey("where", func() {
			test("select a from foo where a", func() PlanStage {
				source := createMongoSource(1, "foo", "foo")
				reqCols := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				return NewProjectStage(
					NewFilterStage(source, NewSQLColumnExpr(1, "foo", "a", schema.SQLInt, schema.MongoInt), reqCols),
					createProjectedColumn(1, source, "foo", "a", "", "a"),
				)
			})

			test("select a from foo where false", func() PlanStage {
				source := createMongoSource(1, "foo", "foo")
				reqCols := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				return NewProjectStage(
					NewFilterStage(source, SQLFalse, reqCols),
					createProjectedColumn(1, source, "foo", "a", "", "a"),
				)
			})

			test("select a from foo where true", func() PlanStage {
				source := createMongoSource(1, "foo", "foo")
				reqCols := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				return NewProjectStage(
					NewFilterStage(source, SQLTrue, reqCols),
					createProjectedColumn(1, source, "foo", "a", "", "a"),
				)
			})

			test("select a from foo where g = true", func() PlanStage {
				source := createMongoSource(1, "foo", "foo")
				reqColsWhere := createReqCols([]PlanStage{source}, []string{"foo.g"}, map[string]int{"foo.g": 1})
				reqColsSelect := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				return NewProjectStage(
					NewFilterStage(source,
						&SQLEqualsExpr{
							left:  NewSQLColumnExpr(1, "foo", "g", schema.SQLBoolean, schema.MongoBool),
							right: SQLTrue,
						},
						append(reqColsWhere, reqColsSelect...),
					),
					createProjectedColumn(1, source, "foo", "a", "", "a"),
				)
			})

			test("select a from foo where a > 10", func() PlanStage {
				source := createMongoSource(1, "foo", "foo")
				reqCols := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				return NewProjectStage(
					NewFilterStage(source,
						&SQLGreaterThanExpr{
							left:  NewSQLColumnExpr(1, "foo", "a", schema.SQLInt, schema.MongoInt),
							right: SQLInt(10),
						},
						reqCols,
					),
					createProjectedColumn(1, source, "foo", "a", "", "a"),
				)
			})

			test("select a as b from foo where b > 10", func() PlanStage {
				source := createMongoSource(1, "foo", "foo")
				reqColsWhere := createReqCols([]PlanStage{source}, []string{"foo.b"}, map[string]int{"foo.b": 1})
				reqColsSelect := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				return NewProjectStage(
					NewFilterStage(source,
						&SQLGreaterThanExpr{
							left:  NewSQLColumnExpr(1, "foo", "b", schema.SQLInt, schema.MongoInt),
							right: SQLInt(10),
						},
						append(reqColsWhere, reqColsSelect...),
					),
					createProjectedColumn(1, source, "foo", "a", "", "b"),
				)
			})

			Convey("subqueries", func() {
				Convey("correlated", func() {
					test("select a from foo where (b) = (select b from bar where foo.a = bar.a)", func() PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(2, "bar", "bar")
						reqColsSelect := createReqCols([]PlanStage{fooSource}, []string{"foo.a"}, map[string]int{"foo.a": 1})
						reqColsSet := createReqCols([]PlanStage{fooSource}, []string{"foo.b"}, map[string]int{"foo.b": 1})
						reqColsSub := createReqCols([]PlanStage{fooSource, barSource}, []string{"foo.a", "bar.a", "bar.b"}, map[string]int{"foo.a": 1, "bar.a": 2, "bar.b": 2})
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
												reqColsSub,
											),
											createProjectedColumn(2, barSource, "bar", "b", "", "b"),
										),
									},
								},
								append(reqColsSet, reqColsSelect...),
							),
							createProjectedColumn(1, fooSource, "foo", "a", "", "a"),
						)
					})

					test("select a from foo f where (b) = (select b from bar where exists(select 1 from foo where f.a = a))", func() PlanStage {
						fooSource := createMongoSource(1, "foo", "f")
						barSource := createMongoSource(2, "bar", "bar")
						foo3Source := createMongoSource(3, "foo", "foo")
						reqColsSelect := createReqCols([]PlanStage{fooSource}, []string{"f.a"}, map[string]int{"f.a": 1})
						reqColsSet := createReqCols([]PlanStage{fooSource}, []string{"f.b"}, map[string]int{"f.b": 1})
						reqColsSub := createReqCols([]PlanStage{barSource}, []string{"bar.b"}, map[string]int{"bar.b": 2})
						reqColsSub2 := createReqCols([]PlanStage{fooSource, foo3Source}, []string{"foo.a", "f.a"}, map[string]int{"foo.a": 3, "f.a": 1})
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
																reqColsSub2,
															),
															createProjectedColumnFromSQLExpr(3, "", "1", SQLInt(1)),
														),
													},
												},
												append(reqColsSelect, reqColsSub...),
											),
											createProjectedColumn(2, barSource, "bar", "b", "", "b"),
										),
									},
								},
								append(reqColsSet, reqColsSelect...),
							),
							createProjectedColumn(1, fooSource, "f", "a", "", "a"),
						)
					})

					test("select a from foo where (b) = (select b from bar where exists(select 1 from foo where bar.a = a))", func() PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(2, "bar", "bar")
						foo3Source := createMongoSource(3, "foo", "foo")
						reqColsSelect := createReqCols([]PlanStage{fooSource}, []string{"foo.a"}, map[string]int{"foo.a": 1})
						reqColsSet := createReqCols([]PlanStage{fooSource}, []string{"foo.b"}, map[string]int{"foo.b": 1})
						reqColsSub := createReqCols([]PlanStage{barSource}, []string{"bar.a", "bar.b"}, map[string]int{"bar.a": 2, "bar.b": 2})
						reqColsSub2 := createReqCols([]PlanStage{foo3Source}, []string{"foo.a"}, map[string]int{"foo.a": 3})
						reqColsSub3 := createReqCols([]PlanStage{barSource}, []string{"bar.a"}, map[string]int{"bar.a": 2})
						return NewProjectStage(
							NewFilterStage(
								fooSource,
								&SQLEqualsExpr{
									left: createSQLColumnExprFromSource(fooSource, "foo", "b"),
									right: &SQLSubqueryExpr{
										correlated: false,
										plan: NewCacheStage(2,
											NewProjectStage(
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
																	append(reqColsSub3, reqColsSub2...),
																),
																createProjectedColumnFromSQLExpr(3, "", "1", SQLInt(1)),
															),
														},
													},
													reqColsSub,
												),
												createProjectedColumn(2, barSource, "bar", "b", "", "b"),
											),
										),
									},
								},
								append(reqColsSet, reqColsSelect...),
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
				reqColsAgg := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				reqColsSum := NewSQLColumnExpr(1, "", "sum(foo.a)", "float64", "")
				return NewProjectStage(
					NewGroupByStage(source,
						nil,
						ProjectedColumns{
							createProjectedColumnFromSQLExpr(1, "", "sum(foo.a)", &SQLAggFunctionExpr{
								Name:  "sum",
								Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
							}),
						},
						append([]SQLExpr{reqColsSum}, reqColsAgg...),
					),
					createProjectedColumnFromSQLExpr(1, "", "sum(a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
				)
			})

			test("select sum(a) from foo group by b", func() PlanStage {
				source := createMongoSource(1, "foo", "foo")
				reqColsGroupBy := createReqCols([]PlanStage{source}, []string{"foo.b"}, map[string]int{"foo.b": 1})
				reqColsAgg := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				reqColsSum := NewSQLColumnExpr(1, "", "sum(foo.a)", "float64", "")
				return NewProjectStage(
					NewGroupByStage(source,
						[]SQLExpr{createSQLColumnExprFromSource(source, "foo", "b")},
						ProjectedColumns{
							createProjectedColumnFromSQLExpr(1, "", "sum(foo.a)", &SQLAggFunctionExpr{
								Name:  "sum",
								Exprs: []SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
							}),
						},
						append(append([]SQLExpr{reqColsSum}, reqColsGroupBy...), reqColsAgg...),
					),
					createProjectedColumnFromSQLExpr(1, "", "sum(a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
				)
			})

			test("select a, sum(a) from foo group by b", func() PlanStage {
				source := createMongoSource(1, "foo", "foo")
				reqColsGroupBy := createReqCols([]PlanStage{source}, []string{"foo.b"}, map[string]int{"foo.b": 1})
				reqColsSelect := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				reqColsSum := NewSQLColumnExpr(1, "", "sum(foo.a)", "float64", "")
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
						append(append(reqColsSelect, reqColsSum), reqColsGroupBy...),
					),
					createProjectedColumn(1, source, "foo", "a", "", "a"),
					createProjectedColumnFromSQLExpr(1, "", "sum(a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
				)
			})

			test("select sum(a) from foo group by b order by sum(a)", func() PlanStage {
				source := createMongoSource(1, "foo", "foo")
				reqColsGroupBy := createReqCols([]PlanStage{source}, []string{"foo.b"}, map[string]int{"foo.b": 1})
				reqColsAgg := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				reqColsSum := NewSQLColumnExpr(1, "", "sum(foo.a)", "float64", "")
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
							append(append([]SQLExpr{reqColsSum}, reqColsGroupBy...), reqColsAgg...),
						),
						[]SQLExpr{reqColsSum},
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
				reqCols := createReqCols([]PlanStage{source}, []string{"foo.b"}, map[string]int{"foo.b": 1})
				reqColsAgg := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				reqColsSum := NewSQLColumnExpr(1, "", "sum(foo.a)", "float64", "")
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
							append(append([]SQLExpr{reqColsSum}, reqCols...), reqColsAgg...),
						),
						[]SQLExpr{reqColsSum},
						&orderByTerm{
							expr:      NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone),
							ascending: true,
						},
					),
					createProjectedColumnFromSQLExpr(1, "", "sum_a", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
				)
			})

			test("select sum(a) from foo f group by b order by (select c from foo where f._id = _id)", func() PlanStage {
				foo1Source := createMongoSource(1, "foo", "f")
				foo2Source := createMongoSource(2, "foo", "foo")
				reqCols := createReqCols([]PlanStage{foo1Source}, []string{"f.b", "f._id"}, map[string]int{"f.a": 1, "f.b": 1})
				reqColsAgg := createReqCols([]PlanStage{foo1Source}, []string{"f.a"}, map[string]int{"f.a": 1})
				reqColsSum := NewSQLColumnExpr(1, "", "sum(f.a)", "float64", "")
				reqColsSub := createReqCols([]PlanStage{foo1Source}, []string{"f._id"}, map[string]int{"f._id": 1})
				reqColsSub2 := createReqCols([]PlanStage{foo2Source}, []string{"foo.c"}, map[string]int{"foo.c": 2})
				reqColsWhere := createReqCols([]PlanStage{foo2Source}, []string{"foo._id"}, map[string]int{"foo._id": 2})
				return NewProjectStage(
					NewOrderByStage(
						NewGroupByStage(foo1Source,
							[]SQLExpr{createSQLColumnExprFromSource(foo1Source, "f", "b")},
							ProjectedColumns{
								createProjectedColumn(1, foo1Source, "f", "_id", "f", "_id"),
								createProjectedColumnFromSQLExpr(1, "", "sum(f.a)", &SQLAggFunctionExpr{
									Name:  "sum",
									Exprs: []SQLExpr{createSQLColumnExprFromSource(foo1Source, "f", "a")},
								}),
							},
							append(append([]SQLExpr{reqColsSum}, reqCols...), reqColsAgg...),
						),
						append([]SQLExpr{reqColsSum}, reqColsSub...),
						&orderByTerm{
							expr: &SQLSubqueryExpr{
								correlated: true,
								plan: NewProjectStage(
									NewFilterStage(
										foo2Source,
										&SQLEqualsExpr{
											left:  createSQLColumnExprFromSource(foo1Source, "f", "_id"),
											right: createSQLColumnExprFromSource(foo2Source, "foo", "_id"),
										},
										append(append(reqColsSub, reqColsWhere...), reqColsSub2...),
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
				reqCols := createReqCols([]PlanStage{foo1Source, foo2Source}, []string{"foo.b"}, map[string]int{"foo.b": 1})
				reqColsAgg := createReqCols([]PlanStage{foo1Source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				reqColsSum := NewSQLColumnExpr(1, "", "sum(foo.a)", "float64", "")
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
						append(append([]SQLExpr{reqColsSum}, reqCols...), reqColsAgg...),
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
				reqCols := createReqCols([]PlanStage{foo1Source, foo2Source}, []string{"foo.b", "foo.a"}, map[string]int{"foo.b": 1, "foo.a": 1})
				reqColsAgg1 := createReqCols([]PlanStage{foo2Source}, []string{"f.a"}, map[string]int{"f.a": 2})
				reqColsAgg2 := createReqCols([]PlanStage{foo1Source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				reqColsSum := NewSQLColumnExpr(2, "", "sum(f.a+foo.a)", "float64", "")
				return NewProjectStage(
					NewGroupByStage(foo1Source,
						[]SQLExpr{createSQLColumnExprFromSource(foo1Source, "foo", "b")},
						ProjectedColumns{
							createProjectedColumn(1, foo1Source, "foo", "a", "foo", "a"),
						},
						reqCols,
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
									append(append([]SQLExpr{reqColsSum}, reqColsAgg1...), reqColsAgg2...),
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
				reqColsSelect := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				reqColsGroupBy := createReqCols([]PlanStage{source}, []string{"foo.b"}, map[string]int{"foo.b": 1})
				reqColsSum := NewSQLColumnExpr(1, "", "sum(foo.a)", "float64", "")
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
							append(append(reqColsSelect, reqColsGroupBy...), reqColsSum),
						),
						&SQLGreaterThanExpr{
							left:  NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone),
							right: SQLInt(10),
						},
						append(reqColsSelect, reqColsSum),
					),
					createProjectedColumn(1, source, "foo", "a", "", "a"),
				)
			})

			Convey("subqueries", func() {
				Convey("non-correlated", func() {
					test("select a from foo having exists(select 1 from bar)", func() PlanStage {
						fooSource := createMongoSource(1, "foo", "foo")
						barSource := createMongoSource(2, "bar", "bar")
						reqCols := createReqCols([]PlanStage{fooSource}, []string{"foo.a"}, map[string]int{"foo.a": 1})
						return NewProjectStage(
							NewFilterStage(
								fooSource,
								&SQLExistsExpr{
									expr: &SQLSubqueryExpr{
										plan: NewCacheStage(2,
											NewProjectStage(
												barSource,
												createProjectedColumnFromSQLExpr(2, "", "1", SQLInt(1)),
											),
										),
									},
								},
								reqCols,
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
				reqCols := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				return NewProjectStage(
					NewGroupByStage(source,
						[]SQLExpr{createSQLColumnExprFromSource(source, "foo", "a")},
						ProjectedColumns{
							createProjectedColumn(1, source, "foo", "a", "foo", "a"),
						},
						reqCols,
					),
					createProjectedColumn(1, source, "foo", "a", "", "a"),
				)
			})

			test("select distinct sum(a) from foo", func() PlanStage {
				source := createMongoSource(1, "foo", "foo")
				reqColsAgg := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				reqColsSum := NewSQLColumnExpr(1, "", "sum(foo.a)", "float64", "")
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
							append([]SQLExpr{reqColsSum}, reqColsAgg...),
						),
						[]SQLExpr{NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)},
						ProjectedColumns{
							createProjectedColumnFromSQLExpr(1, "", "sum(foo.a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
						},
						[]SQLExpr{reqColsSum},
					),
					createProjectedColumnFromSQLExpr(1, "", "sum(a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
				)
			})

			test("select distinct sum(a) from foo having sum(a) > 20", func() PlanStage {
				source := createMongoSource(1, "foo", "foo")
				reqColsAgg := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				reqColsSum := NewSQLColumnExpr(1, "", "sum(foo.a)", "float64", "")
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
								append([]SQLExpr{reqColsSum}, reqColsAgg...),
							),
							&SQLGreaterThanExpr{
								left:  NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone),
								right: SQLInt(20),
							},
							[]SQLExpr{reqColsSum},
						),
						[]SQLExpr{NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)},
						ProjectedColumns{
							createProjectedColumnFromSQLExpr(1, "", "sum(foo.a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
						},
						[]SQLExpr{reqColsSum},
					),
					createProjectedColumnFromSQLExpr(1, "", "sum(a)", NewSQLColumnExpr(1, "", "sum(foo.a)", schema.SQLFloat, schema.MongoNone)),
				)
			})
		})

		Convey("order by", func() {
			test("select a from foo order by a", func() PlanStage {
				source := createMongoSource(1, "foo", "foo")
				reqCols := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				return NewProjectStage(
					NewOrderByStage(source,
						reqCols,
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
				reqCols := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				return NewProjectStage(
					NewOrderByStage(source,
						reqCols,
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
				reqCols := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				return NewProjectStage(
					NewOrderByStage(source,
						reqCols,
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
				reqCols := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				return NewProjectStage(
					NewOrderByStage(source,
						reqCols,
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
				reqCols := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				return NewProjectStage(
					NewOrderByStage(source,
						reqCols,
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
				reqCols := createReqColsStar(source, "foo", 1)
				return NewProjectStage(
					NewOrderByStage(source,
						reqCols,
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
				reqCols := createReqColsStar(source, "foo", 1)
				return NewProjectStage(
					NewOrderByStage(source,
						reqCols,
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
				reqCols := createReqColsStar(source, "foo", 1)
				columns := append(createAllProjectedColumnsFromSource(1, source, ""), createProjectedColumn(1, source, "foo", "a", "", "a"))
				return NewProjectStage(
					NewOrderByStage(source,
						reqCols,
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
				reqCols := createReqCols([]PlanStage{source}, []string{"foo.a"}, map[string]int{"foo.a": 1})
				return NewProjectStage(
					NewOrderByStage(source,
						reqCols,
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
				reqCols := createReqCols([]PlanStage{source}, []string{"foo.a", "foo.b"}, map[string]int{"foo.a": 1, "foo.b": 1})
				return NewProjectStage(
					NewOrderByStage(source,
						reqCols,
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
						reqCols := createReqCols([]PlanStage{fooSource}, []string{"foo.a", "bar.a"}, map[string]int{"foo.a": 1, "bar.a": 2})
						return NewProjectStage(
							NewOrderByStage(
								fooSource,
								reqCols,
								&orderByTerm{
									expr: &SQLSubqueryExpr{
										plan: NewCacheStage(2,
											NewProjectStage(
												barSource,
												createProjectedColumn(2, barSource, "bar", "a", "", "a"),
											),
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
				})
			})

			test("select a from foo limit 0, 0", func() PlanStage {
				source := createMongoSource(1, "foo", "foo")
				return NewEmptyStage([]*Column{
					createProjectedColumn(1, source, "foo", "a", "", "a").Column,
				})
			})
		})

		Convey("errors", func() {
			testError("select a", `ERROR 1054 (42S22): Unknown column 'a' in 'field list'`)

			testError("select a from idk", `ERROR 1051 (42S02): Unknown table 'test.idk'`)
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
		})
	})
}

func TestAlgebrizeSet(t *testing.T) {

	testSchema, err := schema.New(testSchema1)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}
	defaultDbName := "test"

	test := func(sql string, expectedPlanFactory func() *SetExecutor) {
		Convey(sql, func() {
			statement, err := parser.Parse(sql)
			So(err, ShouldBeNil)

			setStatement := statement.(*parser.Set)
			actual, err := AlgebrizeSet(setStatement, defaultDbName, testSchema)
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
		r, _ := NewMongoSourceStage(selectID, testSchema, defaultDbName, tableName, aliasName)
		return r
	}

	Convey("Subject: Algebrize Set Statements", t, func() {
		test("set @t1 = 12", func() *SetExecutor {
			return NewSetExecutor(
				[]*SQLAssignmentExpr{
					&SQLAssignmentExpr{
						variable: &SQLVariableExpr{
							Name: "t1",
							Kind: UserVariable,
						},
						expr: SQLInt(12),
					},
				},
			)
		})

		test("set @@t1 = 12", func() *SetExecutor {
			return NewSetExecutor(
				[]*SQLAssignmentExpr{
					&SQLAssignmentExpr{
						variable: &SQLVariableExpr{
							Name: "t1",
							Kind: SessionVariable,
						},
						expr: SQLInt(12),
					},
				},
			)
		})

		test("set @@global.t1 = 12", func() *SetExecutor {
			return NewSetExecutor(
				[]*SQLAssignmentExpr{
					&SQLAssignmentExpr{
						variable: &SQLVariableExpr{
							Name: "t1",
							Kind: GlobalVariable,
						},
						expr: SQLInt(12),
					},
				},
			)
		})

		test("set @@global.t1 = (select a from foo)", func() *SetExecutor {
			fooSource := createMongoSource(1, "foo", "foo")
			return NewSetExecutor(
				[]*SQLAssignmentExpr{
					&SQLAssignmentExpr{
						variable: &SQLVariableExpr{
							Name: "t1",
							Kind: GlobalVariable,
						},
						expr: &SQLSubqueryExpr{
							correlated: false,
							plan: NewCacheStage(1,
								NewProjectStage(
									fooSource,
									createProjectedColumn(1, fooSource, "foo", "a", "", "a")),
							),
						},
					},
				},
			)
		})

		test("set @@t1=12, @t2=11", func() *SetExecutor {
			return NewSetExecutor(
				[]*SQLAssignmentExpr{
					&SQLAssignmentExpr{
						variable: &SQLVariableExpr{
							Name: "t1",
							Kind: SessionVariable,
						},
						expr: SQLInt(12),
					},
					&SQLAssignmentExpr{
						variable: &SQLVariableExpr{
							Name: "t2",
							Kind: UserVariable,
						},
						expr: SQLInt(11),
					},
				},
			)
		})
	})
}

func TestAlgebrizeExpr(t *testing.T) {
	testSchema, _ := schema.New(testSchema1)
	source, _ := NewMongoSourceStage(1, testSchema, "test", "foo", "foo")

	test := func(sql string, expected SQLExpr) {
		Convey(sql, func() {
			statement, err := parser.Parse("select " + sql + " from foo")
			So(err, ShouldBeNil)

			selectStatement := statement.(*parser.Select)
			actualPlan, err := AlgebrizeSelect(selectStatement, "test", testSchema)
			So(err, ShouldBeNil)

			actual := (actualPlan.(*ProjectStage)).projectedColumns[0].Expr

			So(actual, ShouldResemble, expected)
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

		SkipConvey("Ctor", func() {
			test("TIMESTAMP '2014-06-07 00:00:00.000'", SQLDate{time.Now()})
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

		Convey("Is", func() {
			test("a is true", &SQLIsExpr{
				left:  createSQLColumnExpr("a"),
				right: SQLBool(true),
			})
		})

		Convey("Is Not", func() {
			test("a is not true", &SQLNotExpr{&SQLIsExpr{
				left:  createSQLColumnExpr("a"),
				right: SQLBool(true),
			}})
		})

		Convey("Is Null", func() {
			test("a IS NULL", &SQLIsExpr{
				left:  createSQLColumnExpr("a"),
				right: SQLNull,
			})
		})

		Convey("Is Not Null", func() {
			test("a IS NOT NULL", &SQLNotExpr{&SQLIsExpr{
				left:  createSQLColumnExpr("a"),
				right: SQLNull,
			}})
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

		Convey("True", func() {
			test("TRUE", SQLTrue)
		})

		Convey("False", func() {
			test("FALSE", SQLFalse)
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

		Convey("Variable", func() {
			test("@@global.test_variable", &SQLVariableExpr{Name: "test_variable", Kind: GlobalVariable})
			test("@@session.test_variable", &SQLVariableExpr{Name: "test_variable", Kind: SessionVariable})
			test("@@local.test_variable", &SQLVariableExpr{Name: "test_variable", Kind: SessionVariable})
			test("@@test_variable", &SQLVariableExpr{Name: "test_variable", Kind: SessionVariable})
			test("@hmmm", &SQLVariableExpr{Name: "hmmm", Kind: UserVariable})

		})
	})
}
