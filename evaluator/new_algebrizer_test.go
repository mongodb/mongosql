package evaluator

import (
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
	})
}
