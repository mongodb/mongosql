package evaluator_test

import (
	"fmt"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	_      fmt.Stringer = nil
	limit  uint64
	offset uint64
)

func TestLimitPlanStage(t *testing.T) {
	runTest := func(limit uint64, offset uint64, rows []bson.D, expectedRows []evaluator.Values) {
		ctx := &evaluator.ExecutionCtx{}

		ts := evaluator.NewBSONSourceStage(1, tableOneName, collation.Default, rows)

		l := evaluator.NewLimitStage(ts, offset, limit)

		iter, err := l.Open(ctx)
		So(err, ShouldBeNil)

		row := &evaluator.Row{}

		i := 0

		for iter.Next(row) {
			So(len(row.Data), ShouldEqual, len(expectedRows[i]))
			So(row.Data, ShouldResemble, expectedRows[i])
			row = &evaluator.Row{}
			i++
		}

		So(i, ShouldEqual, len(expectedRows))

		So(iter.Close(), ShouldBeNil)
		So(iter.Err(), ShouldBeNil)
	}

	Convey("A limit operator...", t, func() {

		rows := []bson.D{
			{{Name: "a", Value: 1}},
			{{Name: "a", Value: 2}},
			{{Name: "a", Value: 3}},
			{{Name: "a", Value: 4}},
			{{Name: "a", Value: 5}},
			{{Name: "a", Value: 6}},
			{{Name: "a", Value: 7}},
		}

		Convey("should return only 'limit' records if the limit is less than the total number of"+
			" records", func() {

			limit = 2
			offset = 0

			expected := []evaluator.Values{
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(1)}},
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(2)}},
			}

			runTest(limit, offset, rows, expected)
		})

		Convey("should return the right slice of the records with an offset leaving less records"+
			" than the limit covers", func() {

			limit = 2
			offset = 4

			expected := []evaluator.Values{
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(5)}},
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(6)}},
			}

			runTest(limit, offset, rows, expected)
		})

		Convey("should return no records if the offset is greater than the number of records",
			func() {
				limit = 2
				offset = 40

				expected := []evaluator.Values{}

				runTest(limit, offset, rows, expected)
			})

		Convey("should return no records if the limit and offset are both greater than the "+
			"number of records", func() {

			limit = 40
			offset = 40
			expected := []evaluator.Values{}

			runTest(limit, offset, rows, expected)
		})

		Convey("should only return the number of records if the limit is greater than the number "+
			"of records", func() {

			limit = 40
			offset = 0

			expected := []evaluator.Values{
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(1)}},
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(2)}},
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(3)}},
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(4)}},
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(5)}},
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(6)}},
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(7)}},
			}

			runTest(limit, offset, rows, expected)
		})

		Convey("should return the one record if the limit is 1 with an offset of 1", func() {
			limit = 1
			offset = 1

			expected := []evaluator.Values{{{SelectID: 1, Database: evaluator.BSONSourceDB,
				Table: tableOneName, Name: "a",
				Data: evaluator.SQLInt(2)}}}

			runTest(limit, offset, rows, expected)

		})

		Convey("should return the one record if all but the last record is skipped", func() {

			limit = 1
			offset = 6

			expected := []evaluator.Values{{{SelectID: 1, Database: evaluator.BSONSourceDB,
				Table: tableOneName, Name: "a",
				Data: evaluator.SQLInt(7)}}}

			runTest(limit, offset, rows, expected)
		})

		Convey("should return no records with a limit and offset of 0", func() {

			limit = 0
			offset = 0

			expected := []evaluator.Values{}

			runTest(limit, offset, rows, expected)
		})

		Convey("should return only 'limit' records if there is no offset", func() {

			limit = 3
			offset = 0

			expected := []evaluator.Values{
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(1)}},
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(2)}},
				{{SelectID: 1, Database: evaluator.BSONSourceDB, Table: tableOneName, Name: "a",
					Data: evaluator.SQLInt(3)}}}

			runTest(limit, offset, rows, expected)
		})

		Convey("should return no records if the limit is 0", func() {

			limit = 0
			offset = 4

			expected := []evaluator.Values{}

			runTest(limit, offset, rows, expected)
		})
	})
}
