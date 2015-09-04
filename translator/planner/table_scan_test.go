package planner

import (
	"github.com/erh/mongo-sql-temp/config"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func TestTableScanOperator(t *testing.T) {

	Convey("With a simple test configuration...", t, func() {

		Convey("fetching data from a table scan should return correct results", func() {

			cfg, err := config.ParseConfigData(testCfg)
			So(err, ShouldBeNil)

			session, err := mgo.Dial(cfg.Url)
			So(err, ShouldBeNil)

			rows := []interface{}{
				bson.D{
					bson.DocElem{Name: "_id", Value: 5},
					bson.DocElem{Name: "a", Value: 6},
					bson.DocElem{Name: "b", Value: 7},
				},
				bson.D{
					bson.DocElem{Name: "_id", Value: 15},
					bson.DocElem{Name: "a", Value: 16},
					bson.DocElem{Name: "b", Value: 17},
				},
			}

			dbName := "test"
			colName := "customer"

			collection := session.DB(dbName).C(colName)
			collection.DropCollection()

			for _, row := range rows {
				So(collection.Insert(row), ShouldBeNil)
			}

			ctx := &ExecutionCtx{
				Config: cfg,
				Db:     dbName,
			}

			operator := TableScan{
				collection: colName,
			}

			So(operator.Open(ctx), ShouldBeNil)

			row := &Row{}

			i := 0

			for operator.Next(row) {
				So(len(row.Data), ShouldEqual, 1)
				So(row.Data[0].Table, ShouldEqual, colName)
				So(row.Data[0].Values, ShouldResemble, rows[i])
				i++
			}

			So(operator.Err(), ShouldBeNil)
			So(operator.Close(), ShouldBeNil)
			session.Close()
		})
	})
}
