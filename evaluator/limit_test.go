package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var (
	_ fmt.Stringer = nil
)

func limitTest(operator Operator, rows []bson.D, expectedRows []Values) {

	collectionOne.DropCollection()

	for _, row := range rows {
		So(collectionOne.Insert(row), ShouldBeNil)
	}

	ctx := &ExecutionCtx{
		Config:  cfgOne,
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

func TestLimitOperator(t *testing.T) {

	Convey("A limit operator...", t, func() {

		rows := []bson.D{
			bson.D{{"a", 1}}, bson.D{{"a", 2}}, bson.D{{"a", 3}},
			bson.D{{"a", 4}}, bson.D{{"a", 5}}, bson.D{{"a", 6}}, bson.D{{"a", 7}},
		}

		operator := &Limit{
			source: &Project{
				sExprs: SelectExpressions{
					SelectExpression{
						Column: Column{tableOneName, "a", "a", false},
						Expr:   &sqlparser.ColName{[]byte("a"), []byte(tableOneName)},
					},
				},
				source: &TableScan{
					tableName: tableOneName,
				},
			},
		}

		Convey("should return only 'limit' records if the limit is less than the total number of records", func() {

			operator.rowcount = 2

			expected := []Values{{{"a", "a", SQLInt(1)}}, {{"a", "a", SQLInt(2)}}}

			limitTest(operator, rows, expected)
		})

		Convey("should return the right slice of the records with an offset leaving less records than the limit covers", func() {

			operator.rowcount = 2
			operator.offset = 4

			expected := []Values{{{"a", "a", SQLInt(5)}}, {{"a", "a", SQLInt(6)}}}

			limitTest(operator, rows, expected)
		})

		Convey("should return no records if the offset is greater than the number of records", func() {
			operator.rowcount = 2
			operator.offset = 40

			expected := []Values{}

			limitTest(operator, rows, expected)
		})

		Convey("should return no records if the limit and offset are both greater than the number of records", func() {

			operator.rowcount = 40
			operator.offset = 40
			expected := []Values{}

			limitTest(operator, rows, expected)
		})

		Convey("should only return the number of records if the limit is greater than the number of records", func() {

			operator.rowcount = 40

			expected := []Values{{{"a", "a", SQLInt(1)}},
				{{"a", "a", SQLInt(2)}}, {{"a", "a", SQLInt(3)}}, {{"a", "a", SQLInt(4)}},
				{{"a", "a", SQLInt(5)}}, {{"a", "a", SQLInt(6)}}, {{"a", "a", SQLInt(7)}},
			}

			limitTest(operator, rows, expected)
		})

		Convey("should return the one record if the limit is 1 with an offset of 1", func() {
			operator.rowcount = 1
			operator.offset = 1

			expected := []Values{{{"a", "a", SQLInt(2)}}}

			limitTest(operator, rows, expected)

		})

		Convey("should return the one record if all but the last record is skipped", func() {

			operator.rowcount = 1
			operator.offset = 6

			expected := []Values{{{"a", "a", SQLInt(7)}}}

			limitTest(operator, rows, expected)
		})

		Convey("should return no records with a limit and offset of 0", func() {

			operator.rowcount = 0
			operator.offset = 0

			expected := []Values{}

			limitTest(operator, rows, expected)
		})

		Convey("should return only 'limit' records if there is no offset", func() {

			operator.rowcount = 3
			operator.offset = 0

			expected := []Values{{{"a", "a", SQLInt(1)}}, {{"a", "a", SQLInt(2)}}, {{"a", "a", SQLInt(3)}}}

			limitTest(operator, rows, expected)
		})

		Convey("should return no records if the limit is 0", func() {

			operator.rowcount = 0
			operator.offset = 4

			expected := []Values{}

			limitTest(operator, rows, expected)
		})
	})
}
