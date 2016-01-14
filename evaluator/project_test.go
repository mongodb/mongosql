package evaluator

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var (
	_ fmt.Stringer = nil
)

func projectTest(operator Operator, rows []bson.D, expectedRows []Values) {

	collectionOne.DropCollection()

	for _, row := range rows {
		So(collectionOne.Insert(row), ShouldBeNil)
	}

	ctx := &ExecutionCtx{
		Schema:  cfgOne,
		Db:      dbOne,
		Session: session,
	}

	So(operator.Open(ctx), ShouldBeNil)

	row := &Row{}

	i := 0

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

func TestProjectOperator(t *testing.T) {

	Convey("A project operator...", t, func() {

		rows := []bson.D{
			bson.D{{"a", 6}, {"b", 9}},
			bson.D{{"a", 3}, {"b", 4}},
		}

		sExprs := SelectExpressions{
			SelectExpression{
				Column: Column{tableOneName, "a", "a"},
				Expr:   SQLFieldExpr{tableOneName, "a"},
			},
			SelectExpression{
				Referenced: true,
				Column:     Column{tableOneName, "b", "b"},
				Expr:       SQLFieldExpr{tableOneName, "b"},
			},
		}

		Convey("should filter out referenced columns in select expressions", func() {

			operator := &Project{
				sExprs: sExprs,
				source: &TableScan{
					tableName: tableOneName,
				},
			}

			expected := []Values{{{"a", "a", SQLInt(6)}}, {{"a", "a", SQLInt(3)}}}

			projectTest(operator, rows, expected)
		})

		Convey("should not filter any results if no column is referenced", func() {
			sExprs[1].Referenced = false

			operator := &Project{
				sExprs: sExprs,
				source: &TableScan{
					tableName: tableOneName,
				},
			}

			expected := []Values{{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(9)}}, {{"a", "a", SQLInt(3)}, {"b", "b", SQLInt(4)}}}

			projectTest(operator, rows, expected)
		})

	})
}
