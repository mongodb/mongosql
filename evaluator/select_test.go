package evaluator

import (
	"fmt"
	"github.com/10gen/sqlproxy/config"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var (
	_ fmt.Stringer = nil
)

func selectTest(operator Operator, rows []bson.D, expectedRows []Values) {

	cfg, err := config.ParseConfigData(testConfig1)
	So(err, ShouldBeNil)

	session, err := mgo.Dial(cfg.Url)
	So(err, ShouldBeNil)

	collection := session.DB(dbName).C(tableTwoName)
	collection.DropCollection()

	for _, row := range rows {
		So(collection.Insert(row), ShouldBeNil)
	}

	ctx := &ExecutionCtx{
		Config: cfg,
		Db:     dbName,
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

	So(operator.Err(), ShouldBeNil)
	So(operator.Close(), ShouldBeNil)

	collection.DropCollection()

	session.Close()

}

func TestSelectOperator(t *testing.T) {

	Convey("With a simple test configuration...", t, func() {
		rows := []bson.D{
			{{"a", SQLInt(6)}, {"b", SQLInt(7)}, {"_id", SQLInt(5)}},
			{{"a", SQLInt(16)}, {"b", SQLInt(17)}, {"_id", SQLInt(15)}},
		}

		Convey("a select operator from one table with a star field return the right columns requested", func() {

			operator := &Select{
				source: &TableScan{
					tableName: tableTwoName,
				},
			}

			var expected []Values
			for _, document := range rows {
				values, err := bsonDToValues(document)
				So(err, ShouldBeNil)
				expected = append(expected, values)
			}
			selectTest(operator, rows, expected)
		})

		Convey("a select operator from one table with non-star fields return the right columns requested", func() {

			expectedRows := []Values{
				{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
				{{"a", "a", SQLInt(16)}, {"b", "b", SQLInt(17)}},
			}

			columns := []Column{{tableTwoName, "a", "a", false}, {tableTwoName, "b", "b", false}}

			sExprs := SelectExpressions{
				{columns[0], nil, nil, false},
				{columns[1], nil, nil, false},
			}

			operator := &Select{
				sExprs: sExprs,
				source: &TableScan{
					tableName: tableTwoName,
				},
			}

			selectTest(operator, rows, expectedRows)

		})

	})
}
