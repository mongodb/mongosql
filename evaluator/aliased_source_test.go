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

func aliasedSourceTest(operator Operator, rows []bson.D, expectedRows []Values) {

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
		So(row.Data[0].Table, ShouldEqual, tableOneName)
		So(row.Data[0].Values[0].Data, ShouldResemble, expectedRows[i][0].Data)
		row = &Row{}
		i++
	}

	So(i, ShouldEqual, len(expectedRows))

	So(operator.Err(), ShouldBeNil)
	So(operator.Close(), ShouldBeNil)

	collection.DropCollection()

	session.Close()

}

func TestAliasedSourceOperator(t *testing.T) {

	Convey("With a simple test configuration...", t, func() {

		rows := []bson.D{
			bson.D{{"a", 6}, {"b", 7}, {"_id", 5}},
			bson.D{{"a", 16}, {"b", 17}, {"_id", 15}},
		}

		Convey("an aliased source operator should properly present the table and row data", func() {

			operator := &AliasedSource{
				source: &TableScan{
					tableName: tableTwoName,
				},
				tableName: tableOneName,
			}

			expected := []Values{
				{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}, {"_id", "_id", SQLInt(5)}},
				{{"a", "a", SQLInt(16)}, {"b", "b", SQLInt(17)}, {"_id", "_id", SQLInt(15)}},
			}

			aliasedSourceTest(operator, rows, expected)

		})
	})
}
