package evaluator

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/config"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var (
	_ fmt.Stringer = nil
)

func limitTest(operator Operator, rows []bson.D, expectedRows []Values) {

	cfg, err := config.ParseConfigData(testConfig1)
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

	So(i, ShouldEqual, len(expectedRows))

	So(operator.Err(), ShouldBeNil)
	So(operator.Close(), ShouldBeNil)

	collection.DropCollection()

	session.Close()

}

func TestLimitOperator(t *testing.T) {

	Convey("A limit operator...", t, func() {

		rows := []bson.D{
			bson.D{{"a", 1}}, bson.D{{"a", 2}}, bson.D{{"a", 3}},
			bson.D{{"a", 4}}, bson.D{{"a", 5}}, bson.D{{"a", 6}}, bson.D{{"a", 7}},
		}

		operator := &Limit{
			source: &Project{
				sExprs: SelectExpressions{
					SelectExpression{
						Column: Column{tableOneName, "a", "a"},
						Expr:   &sqlparser.ColName{[]byte("a"), []byte(tableOneName)},
					},
				},
				source: &TableScan{
					tableName: tableOneName,
				},
			},
		}

		Convey("should accordingly handle limits less than the total number of records", func() {

			operator.rowcount = 2

			expected := []Values{{{"a", "a", 1}}, {{"a", "a", 2}}}

			limitTest(operator, rows, expected)
		})

		Convey("should limit results with offsets accordingly", func() {

			operator.rowcount = 2
			operator.offset = 4

			expected := []Values{{{"a", "a", 5}}, {{"a", "a", 6}}}

			limitTest(operator, rows, expected)
		})

		Convey("should accordingly handle limits with offsets greater than the number of records", func() {
			operator.rowcount = 2
			operator.offset = 40

			expected := []Values{}

			limitTest(operator, rows, expected)
		})

		Convey("should accordingly handle limits and offsets greater than the number of records", func() {

			operator.rowcount = 40
			operator.offset = 40
			expected := []Values{}

			limitTest(operator, rows, expected)
		})

		Convey("should accordingly handle limits that are greater than the number of records", func() {

			operator.rowcount = 40

			expected := []Values{{{"a", "a", 1}},
				{{"a", "a", 2}}, {{"a", "a", 3}}, {{"a", "a", 4}},
				{{"a", "a", 5}}, {{"a", "a", 6}}, {{"a", "a", 7}},
			}

			limitTest(operator, rows, expected)
		})

		Convey("should accordingly handle limit and offset 1 at start", func() {
			operator.rowcount = 1
			operator.offset = 1

			expected := []Values{{{"a", "a", 2}}}

			limitTest(operator, rows, expected)

		})

		Convey("should accordingly handle limit 1 at end", func() {

			operator.rowcount = 1
			operator.offset = 6

			expected := []Values{{{"a", "a", 7}}}

			limitTest(operator, rows, expected)
		})

		Convey("should accordingly handle limit/offset 0", func() {

			operator.rowcount = 0
			operator.offset = 0

			expected := []Values{}

			limitTest(operator, rows, expected)
		})

		Convey("should accordingly handle lone limit", func() {

			operator.rowcount = 3
			operator.offset = 0

			expected := []Values{{{"a", "a", 1}}, {{"a", "a", 2}}, {{"a", "a", 3}}}

			limitTest(operator, rows, expected)
		})

		Convey("should accordingly handle lone offset", func() {

			operator.rowcount = 0
			operator.offset = 4

			expected := []Values{}

			limitTest(operator, rows, expected)
		})
	})
}
