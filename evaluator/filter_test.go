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

func filterSourceTest(operator Operator, rows []bson.D, expectedRows []Values) {

	collectionTwo.DropCollection()

	for _, row := range rows {
		So(collectionTwo.Insert(row), ShouldBeNil)
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
		So(row.Data[0].Table, ShouldEqual, tableTwoName)
		So(row.Data[0].Values, ShouldResemble, expectedRows[i])
		row = &Row{}
		i++
	}

	So(i, ShouldEqual, len(expectedRows))

	So(operator.Close(), ShouldBeNil)
	So(operator.Err(), ShouldBeNil)
}

func TestFilterOperator(t *testing.T) {

	Convey("With a simple test configuration...", t, func() {

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
				matcher, err := getSQLExprFromSQL(query)
				So(err, ShouldBeNil)

				operator := &Filter{
					source: &TableScan{
						tableName: tableTwoName,
					},
					matcher: matcher,
				}

				filterSourceTest(operator, rows, expected[i])
			}
		})
	})
}
