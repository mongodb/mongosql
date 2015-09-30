package planner

import (
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/translator/types"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func TestTableScanOperator(t *testing.T) {

	Convey("With a simple test configuration...", t, func() {

		Convey("fetching data from a table scan should return correct results in the right order", func() {

			cfg, err := config.ParseConfigData(testConfigSimple)
			So(err, ShouldBeNil)

			session, err := mgo.Dial(cfg.Url)
			So(err, ShouldBeNil)

			rows := []bson.D{
				bson.D{
					bson.DocElem{Name: "a", Value: 6},
					bson.DocElem{Name: "b", Value: 7},
					bson.DocElem{Name: "_id", Value: 5},
				},
				bson.D{
					bson.DocElem{Name: "a", Value: 16},
					bson.DocElem{Name: "b", Value: 17},
					bson.DocElem{Name: "_id", Value: 15},
				},
			}

			var expected []types.Values
			for _, document := range rows {
				expected = append(expected, bsonDToValues(document))
			}

			collection := session.DB(dbName).C(tableTwoName)
			collection.DropCollection()

			for _, row := range rows {
				So(collection.Insert(row), ShouldBeNil)
			}

			ctx := &ExecutionCtx{
				Config: cfg,
				Db:     dbName,
			}

			operator := TableScan{
				tableName: tableTwoName,
			}

			So(operator.Open(ctx), ShouldBeNil)

			row := &types.Row{}

			i := 0

			for operator.Next(row) {
				So(len(row.Data), ShouldEqual, 1)
				So(row.Data[0].Table, ShouldEqual, tableTwoName)
				So(row.Data[0].Values, ShouldResemble, expected[i])
				row = &types.Row{}
				i++
			}

			So(operator.Err(), ShouldBeNil)
			So(operator.Close(), ShouldBeNil)
			session.Close()
		})
	})
}
