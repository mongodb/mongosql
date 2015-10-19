package evaluator

import (
	"fmt"
	"github.com/erh/mongo-sql-temp/config"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var (
	_ fmt.Stringer = nil
)

func filterSourceTest(operator Operator, rows []bson.D, expectedRows []Values) {

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

	So(operator.Err(), ShouldBeNil)
	So(operator.Close(), ShouldBeNil)

	collection.DropCollection()

	session.Close()

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

			expected := [][]Values{
				{bsonDToValues(rows[1])},
				{bsonDToValues(rows[0])},
				nil,
				{bsonDToValues(rows[1])},
				{bsonDToValues(rows[1])},
				{bsonDToValues(rows[0])},
			}

			for i, query := range queries {
				matcher, err := getMatcherFromSQL(query)
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
