package planner

import (
	"fmt"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/translator/evaluator"
	"github.com/erh/mongo-sql-temp/translator/types"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var (
	_ fmt.Stringer = nil
)

func selectTest(operator Operator, rows []bson.D, expectedRows []types.Values) {

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

	row := &types.Row{}

	i := 0

	for operator.Next(row) {
		So(len(row.Data), ShouldEqual, 1)
		So(row.Data[0].Table, ShouldEqual, tableTwoName)
		So(row.Data[0].Values, ShouldResemble, expectedRows[i])
		row = &types.Row{}
		i++
	}

	So(operator.Err(), ShouldBeNil)
	So(operator.Close(), ShouldBeNil)

	collection.DropCollection()

	session.Close()

}

func TestSelectOperator(t *testing.T) {

	Convey("With a simple test configuration...", t, func() {
		rows := []bson.D{
			{{"a", evaluator.SQLInt(6)}, {"b", evaluator.SQLInt(7)}, {"_id", evaluator.SQLInt(5)}},
			{{"a", evaluator.SQLInt(16)}, {"b", evaluator.SQLInt(17)}, {"_id", evaluator.SQLInt(15)}},
		}

		Convey("a select operator from one table with a star field return the right columns requested", func() {

			operator := &Select{
				source: &TableScan{
					tableName: tableTwoName,
				},
			}

			var expected []types.Values
			for _, document := range rows {
				expected = append(expected, bsonDToValues(document))
			}
			selectTest(operator, rows, expected)
		})

		Convey("a select operator from one table with non-star fields return the right columns requested", func() {

			expectedRows := []types.Values{
				{
					{"a", "a", evaluator.SQLInt(6)},
					{"b", "b", evaluator.SQLInt(7)},
				},
				{
					{"a", "a", evaluator.SQLInt(16)},
					{"b", "b", evaluator.SQLInt(17)},
				},
			}

			columns := []Column{
				{tableTwoName, "a", "a"},
				{tableTwoName, "b", "b"},
			}

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
