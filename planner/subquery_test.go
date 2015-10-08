package planner

import (
	"fmt"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/evaluator"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var (
	_ fmt.Stringer = nil
)

func subqueryTest(operator Operator, rows []bson.D, expectedRows []evaluator.Values) {

	cfg, err := config.ParseConfigData(testConfigSimple)
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

	row := &evaluator.Row{}

	i := 0

	for operator.Next(row) {
		So(len(row.Data), ShouldEqual, 1)
		So(row.Data[0].Table, ShouldEqual, tableOneName)
		So(row.Data[0].Values[0].Data, ShouldResemble, expectedRows[i][0].Data)
		row = &evaluator.Row{}
		i++
	}

	So(operator.Err(), ShouldBeNil)
	So(operator.Close(), ShouldBeNil)

	collection.DropCollection()

	session.Close()

}
func TestSubQueryOperator(t *testing.T) {

	Convey("With a simple test configuration...", t, func() {

		rows := []bson.D{
			bson.D{{"a", 6}, {"b", 7}, {"_id", 5}},
			bson.D{{"a", 16}, {"b", 17}, {"_id", 15}},
		}

		Convey("a subquery operator should properly present the table and row data", func() {

			operator := &Subquery{
				source: &TableScan{
					tableName: tableTwoName,
				},
				tableName: tableOneName,
			}

			expected := []evaluator.Values{
				{{"a", "a", 6}, {"b", "b", 7}, {"_id", "_id", 5}},
				{{"a", "a", 16}, {"b", "b", 17}, {"_id", "_id", 15}},
			}

			subqueryTest(operator, rows, expected)

		})
	})
}
