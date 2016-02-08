package evaluator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestSubqueryOperator(t *testing.T) {

	runTest := func(subquery *Subquery, rows []bson.D, expectedRows []Values) {

		ctx := &ExecutionCtx{
			Schema: cfgOne,
			Db:     dbOne,
		}

		ts, err := NewBSONSource(ctx, tableTwoName, rows)
		So(err, ShouldBeNil)

		subquery.source = ts

		So(subquery.Open(ctx), ShouldBeNil)

		row := &Row{}

		i := 0

		for subquery.Next(row) {
			So(len(row.Data), ShouldEqual, 1)
			So(row.Data[0].Table, ShouldEqual, tableOneName)
			So(row.Data[0].Values[0].Data, ShouldResemble, expectedRows[i][0].Data)
			row = &Row{}
			i++
		}

		So(i, ShouldEqual, len(expectedRows))

		So(subquery.Close(), ShouldBeNil)
		So(subquery.Err(), ShouldBeNil)
	}

	Convey("With a simple test configuration...", t, func() {

		rows := []bson.D{
			bson.D{{"a", 6}, {"b", 7}, {"_id", 5}},
			bson.D{{"a", 16}, {"b", 17}, {"_id", 15}},
		}

		Convey("a subquery source operator should properly present the table and row data", func() {

			expected := []Values{
				{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}, {"_id", "_id", SQLInt(5)}},
				{{"a", "a", SQLInt(16)}, {"b", "b", SQLInt(17)}, {"_id", "_id", SQLInt(15)}},
			}

			subquery := &Subquery{
				tableName: tableOneName,
			}

			runTest(subquery, rows, expected)

		})
	})
}
