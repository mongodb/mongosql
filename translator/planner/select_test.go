package planner

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

func selectTest(operator Operator, rows, expectedRows []interface{}) {

	cfg, err := config.ParseConfigData(testConfigSimple)
	So(err, ShouldBeNil)

	session, err := mgo.Dial(cfg.Url)
	So(err, ShouldBeNil)

	collection := session.DB(dbName).C(tableOneName)
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
		So(row.Data[0].Values, ShouldResemble, expectedRows[i])
		row = &Row{}
		i++
	}

	So(operator.Err(), ShouldBeNil)
	So(operator.Close(), ShouldBeNil)

	collection.DropCollection()

	session.Close()

}

func TestSelectOperator(t *testing.T) {

	Convey("With a simple test configuration...", t, func() {

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

		Convey("a select operator from one table with a star field return the right columns requested", func() {

			operator := &Select{
				source: &TableScan{
					tableName: tableOneName,
				},
			}

			selectTest(operator, rows, rows)

		})

		Convey("a select operator from one table with non-star fields return the right columns requested", func() {

			expectedRows := []interface{}{
				bson.D{
					bson.DocElem{Name: "a", Value: 6},
					bson.DocElem{Name: "b", Value: 7},
				},
				bson.D{
					bson.DocElem{Name: "a", Value: 16},
					bson.DocElem{Name: "b", Value: 17},
				},
			}

			columns := []*Column{
				{tableOneName, "a", "a", nil},
				{tableOneName, "b", "b", nil},
			}

			operator := &Select{
				Columns: columns,
				source: &TableScan{
					tableName: tableOneName,
				},
			}

			selectTest(operator, rows, expectedRows)

		})

	})
}
