package evaluator

import (
	"testing"

	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestProjectOperator(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne
	ctx := &ExecutionCtx{Schema: cfgOne, Db: dbOne}
	columnType := schema.ColumnType{schema.SQLInt, schema.MongoInt}

	runTest := func(project *Project, optimize bool, rows []bson.D, expectedRows []Values) {
		ts, err := NewBSONSource(ctx, tableOneName, rows)
		So(err, ShouldBeNil)
		var operator Operator
		operator = project.WithSource(ts)
		if optimize {
			operator, err = OptimizeOperator(ctx, operator)
			So(err, ShouldBeNil)
		}

		So(operator.Open(ctx), ShouldBeNil)

		i := 0
		row := &Row{}

		for operator.Next(row) {
			So(len(row.Data), ShouldEqual, 1)
			So(row.Data[0].Table, ShouldEqual, tableOneName)
			So(row.Data[0].Values, ShouldResemble, expectedRows[i])
			row = &Row{}
			i++
		}

		So(i, ShouldEqual, len(expectedRows))

		So(operator.Close(), ShouldBeNil)
		So(operator.Err(), ShouldBeNil)
	}

	Convey("A project operator...", t, func() {

		rows := []bson.D{
			bson.D{{"a", 6}, {"b", 9}},
			bson.D{{"a", 3}, {"b", 4}},
		}

		sExprs := SelectExpressions{
			SelectExpression{
				Column: &Column{tableOneName, "a", "a", schema.SQLInt, schema.MongoInt},
				Expr:   SQLColumnExpr{tableOneName, "a", columnType},
			},
			SelectExpression{
				Referenced: true,
				Column:     &Column{tableOneName, "b", "b", schema.SQLInt, schema.MongoInt},
				Expr:       SQLColumnExpr{tableOneName, "b", columnType},
			},
		}

		Convey("should filter out referenced columns in select expressions", func() {

			project := &Project{
				sExprs: sExprs,
			}

			expected := []Values{{{"a", "a", SQLInt(6)}}, {{"a", "a", SQLInt(3)}}}

			runTest(project, false, rows, expected)
			Convey("and should produce identical results after optimization", func() {
				runTest(project, true, rows, expected)
			})
		})

		Convey("should not filter any results if no column is referenced", func() {
			sExprs[1].Referenced = false

			project := &Project{
				sExprs: sExprs,
			}

			expected := []Values{{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(9)}}, {{"a", "a", SQLInt(3)}, {"b", "b", SQLInt(4)}}}

			runTest(project, false, rows, expected)

			Convey("and should produce identical results after optimization", func() {
				runTest(project, true, rows, expected)
			})
		})

	})
}
