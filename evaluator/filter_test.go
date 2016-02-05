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

func TestFilterOperator(t *testing.T) {

	runTest := func(filter *Filter, rows []bson.D, expectedRows []Values) {

		collectionTwo.DropCollection()

		for _, row := range rows {
			So(collectionTwo.Insert(row), ShouldBeNil)
		}

		ctx := &ExecutionCtx{
			Schema:  cfgOne,
			Db:      dbOne,
			Session: session,
		}

		ts, err := NewTableScan(ctx, dbOne, tableTwoName, "")
		So(err, ShouldBeNil)

		filter.source = ts

		So(filter.Open(ctx), ShouldBeNil)

		row := &Row{}

		i := 0

		for filter.Next(row) {
			So(len(row.Data), ShouldEqual, 1)
			So(row.Data[0].Table, ShouldEqual, tableTwoName)
			So(row.Data[0].Values, ShouldResemble, expectedRows[i])
			row = &Row{}
			i++
		}

		So(i, ShouldEqual, len(expectedRows))

		So(filter.Close(), ShouldBeNil)
		So(filter.Err(), ShouldBeNil)
	}

	Convey("With a simple test configuration...", t, func() {

		schema, err := schema.ParseSchemaData(testSchema3)
		So(err, ShouldBeNil)

		rows := []bson.D{
			bson.D{{"a", 6}, {"b", 7}, {"_id", 5}},
			bson.D{{"a", 16}, {"b", 17}, {"_id", 15}},
		}

		Convey("a filter operator should only return rows that match", func() {
			queries := []string{
				fmt.Sprintf("select * from %v where a = 16", tableTwoName),
				fmt.Sprintf("select * from %v where a = 6", tableTwoName),
				fmt.Sprintf("select * from %v where a = 99", tableTwoName),
				fmt.Sprintf("select * from %v where b > 9", tableTwoName),
				fmt.Sprintf("select * from %v where b > 9 or a < 5", tableTwoName),
				fmt.Sprintf("select * from %v where b = 7 or a = 6", tableTwoName),
			}

			r0, err := bsonDToValues(rows[0])
			So(err, ShouldBeNil)
			r1, err := bsonDToValues(rows[1])
			So(err, ShouldBeNil)

			expected := [][]Values{{r1}, {r0}, nil, {r1}, {r1}, {r0}}

			for i, query := range queries {
				matcher, err := getWhereSQLExprFromSQL(schema, query)
				So(err, ShouldBeNil)

				operator := &Filter{
					matcher: matcher,
				}

				runTest(operator, rows, expected[i])
			}
		})
	})
}
