package evaluator

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

var (
	_ fmt.Stringer = nil
)

func TestFilterPlanStage(t *testing.T) {
	runTest := func(filter *FilterStage, rows []bson.D, expectedRows []Values) {

		ctx := &ExecutionCtx{}

		bss := &BSONSourceStage{tableTwoName, rows}

		filter.source = bss
		iter, err := filter.Open(ctx)

		So(err, ShouldBeNil)

		row := &Row{}

		i := 0

		for iter.Next(row) {
			So(len(row.Data), ShouldEqual, len(expectedRows[i]))
			So(row.Data, ShouldResemble, expectedRows[i])
			row = &Row{}
			i++
		}

		So(i, ShouldEqual, len(expectedRows))

		So(iter.Close(), ShouldBeNil)
		So(iter.Err(), ShouldBeNil)
	}

	Convey("With a simple test configuration...", t, func() {

		schema, err := schema.New(testSchema3)
		So(err, ShouldBeNil)

		rows := []bson.D{
			bson.D{{"a", 6}, {"b", 7}, {"_id", 5}},
			bson.D{{"a", 16}, {"b", 17}, {"_id", 15}},
		}

		Convey("a filter operator should only return rows that match", func() {
			queries := []string{
				"a = 16",
				"a = 6",
				"a = 99",
				"b > 9",
				"b > 9 or a < 5",
				"b = 7 or a = 6",
			}

			r0, err := bsonDToValues(tableTwoName, rows[0])
			So(err, ShouldBeNil)
			r1, err := bsonDToValues(tableTwoName, rows[1])
			So(err, ShouldBeNil)

			expected := [][]Values{{r1}, {r0}, nil, {r1}, {r1}, {r0}}

			for i, query := range queries {
				matcher, err := getSQLExpr(schema, dbOne, tableTwoName, query)
				So(err, ShouldBeNil)

				operator := &FilterStage{
					matcher: matcher,
				}

				runTest(operator, rows, expected[i])
			}
		})
	})
}
