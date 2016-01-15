package evaluator

import (
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func subqueryTest(operator Operator, rows []bson.D, expectedRows []Values) {

	collectionTwo.DropCollection()

	for _, row := range rows {
		So(collectionTwo.Insert(row), ShouldBeNil)
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
		So(row.Data[0].Values[0].Data, ShouldResemble, expectedRows[i][0].Data)
		row = &Row{}
		i++
	}

	So(i, ShouldEqual, len(expectedRows))

	So(operator.Close(), ShouldBeNil)
	So(operator.Err(), ShouldBeNil)
}

func TestSubqueryOperator(t *testing.T) {

	Convey("With a simple test configuration...", t, func() {

		rows := []bson.D{
			bson.D{{"a", 6}, {"b", 7}, {"_id", 5}},
			bson.D{{"a", 16}, {"b", 17}, {"_id", 15}},
		}

		Convey("a subquery source operator should properly present the table and row data", func() {

			operator := &Subquery{
				source: &TableScan{
					tableName: tableTwoName,
				},
				tableName: tableOneName,
			}

			expected := []Values{
				{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}, {"_id", "_id", SQLInt(5)}},
				{{"a", "a", SQLInt(16)}, {"b", "b", SQLInt(17)}, {"_id", "_id", SQLInt(15)}},
			}

			subqueryTest(operator, rows, expected)

		})
	})
}
