package evaluator

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

var (
	_ fmt.Stringer = nil
)

func TestProjectOperator(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne

	runTest := func(project *Project, rows []bson.D, expectedRows []Values) {

		ctx := &ExecutionCtx{
			Schema: cfgOne,
			Db:     dbOne,
		}

		ts, err := NewBSONSource(ctx, tableOneName, rows)
		So(err, ShouldBeNil)

		project.source = ts

		So(project.Open(ctx), ShouldBeNil)

		row := &Row{}

		i := 0

		for project.Next(row) {
			So(len(row.Data), ShouldEqual, 1)
			So(row.Data[0].Table, ShouldEqual, tableOneName)
			So(row.Data[0].Values, ShouldResemble, expectedRows[i])
			row = &Row{}
			i++
		}

		So(i, ShouldEqual, len(expectedRows))

		So(project.Close(), ShouldBeNil)
		So(project.Err(), ShouldBeNil)
	}

	Convey("A project operator...", t, func() {

		rows := []bson.D{
			bson.D{{"a", 6}, {"b", 9}},
			bson.D{{"a", 3}, {"b", 4}},
		}

		sExprs := SelectExpressions{
			SelectExpression{
				Column: &Column{tableOneName, "a", "a", "int"},
				Expr:   SQLColumnExpr{tableOneName, "a"},
			},
			SelectExpression{
				Referenced: true,
				Column:     &Column{tableOneName, "b", "b", "int"},
				Expr:       SQLColumnExpr{tableOneName, "b"},
			},
		}

		Convey("should filter out referenced columns in select expressions", func() {

			project := &Project{
				sExprs: sExprs,
			}

			expected := []Values{{{"a", "a", SQLInt(6)}}, {{"a", "a", SQLInt(3)}}}

			runTest(project, rows, expected)
		})

		Convey("should not filter any results if no column is referenced", func() {
			sExprs[1].Referenced = false

			project := &Project{
				sExprs: sExprs,
			}

			expected := []Values{{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(9)}}, {{"a", "a", SQLInt(3)}, {"b", "b", SQLInt(4)}}}

			runTest(project, rows, expected)
		})

	})
}
