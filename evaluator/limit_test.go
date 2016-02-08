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

func TestLimitOperator(t *testing.T) {

	runTest := func(limit *Limit, rows []bson.D, expectedRows []Values) {

		ctx := &ExecutionCtx{
			Schema: cfgOne,
			Db:     dbOne,
		}

		ts, err := NewBSONSource(ctx, tableOneName, rows)
		So(err, ShouldBeNil)

		limit.source = &Project{
			sExprs: SelectExpressions{
				SelectExpression{
					Column: &Column{tableOneName, "a", "a", "int"},
					Expr:   SQLFieldExpr{tableOneName, "a"},
				},
			},
			source: ts,
		}

		So(limit.Open(ctx), ShouldBeNil)

		row := &Row{}

		i := 0

		for limit.Next(row) {
			So(len(row.Data), ShouldEqual, 1)
			So(row.Data[0].Table, ShouldEqual, tableOneName)
			So(row.Data[0].Values, ShouldResemble, expectedRows[i])
			row = &Row{}
			i++
		}

		So(i, ShouldEqual, len(expectedRows))

		So(limit.Close(), ShouldBeNil)
		So(limit.Err(), ShouldBeNil)
	}

	Convey("A limit operator...", t, func() {

		rows := []bson.D{
			bson.D{{"a", 1}}, bson.D{{"a", 2}}, bson.D{{"a", 3}},
			bson.D{{"a", 4}}, bson.D{{"a", 5}}, bson.D{{"a", 6}}, bson.D{{"a", 7}},
		}

		operator := &Limit{}

		Convey("should return only 'limit' records if the limit is less than the total number of records", func() {

			operator.rowcount = 2

			expected := []Values{{{"a", "a", SQLInt(1)}}, {{"a", "a", SQLInt(2)}}}

			runTest(operator, rows, expected)
		})

		Convey("should return the right slice of the records with an offset leaving less records than the limit covers", func() {

			operator.rowcount = 2
			operator.offset = 4

			expected := []Values{{{"a", "a", SQLInt(5)}}, {{"a", "a", SQLInt(6)}}}

			runTest(operator, rows, expected)
		})

		Convey("should return no records if the offset is greater than the number of records", func() {
			operator.rowcount = 2
			operator.offset = 40

			expected := []Values{}

			runTest(operator, rows, expected)
		})

		Convey("should return no records if the limit and offset are both greater than the number of records", func() {

			operator.rowcount = 40
			operator.offset = 40
			expected := []Values{}

			runTest(operator, rows, expected)
		})

		Convey("should only return the number of records if the limit is greater than the number of records", func() {

			operator.rowcount = 40

			expected := []Values{{{"a", "a", SQLInt(1)}},
				{{"a", "a", SQLInt(2)}}, {{"a", "a", SQLInt(3)}}, {{"a", "a", SQLInt(4)}},
				{{"a", "a", SQLInt(5)}}, {{"a", "a", SQLInt(6)}}, {{"a", "a", SQLInt(7)}},
			}

			runTest(operator, rows, expected)
		})

		Convey("should return the one record if the limit is 1 with an offset of 1", func() {
			operator.rowcount = 1
			operator.offset = 1

			expected := []Values{{{"a", "a", SQLInt(2)}}}

			runTest(operator, rows, expected)

		})

		Convey("should return the one record if all but the last record is skipped", func() {

			operator.rowcount = 1
			operator.offset = 6

			expected := []Values{{{"a", "a", SQLInt(7)}}}

			runTest(operator, rows, expected)
		})

		Convey("should return no records with a limit and offset of 0", func() {

			operator.rowcount = 0
			operator.offset = 0

			expected := []Values{}

			runTest(operator, rows, expected)
		})

		Convey("should return only 'limit' records if there is no offset", func() {

			operator.rowcount = 3
			operator.offset = 0

			expected := []Values{{{"a", "a", SQLInt(1)}}, {{"a", "a", SQLInt(2)}}, {{"a", "a", SQLInt(3)}}}

			runTest(operator, rows, expected)
		})

		Convey("should return no records if the limit is 0", func() {

			operator.rowcount = 0
			operator.offset = 4

			expected := []Values{}

			runTest(operator, rows, expected)
		})
	})
}
