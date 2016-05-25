package evaluator

import (
	"testing"

	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

var (
	customers = []bson.D{
		bson.D{
			bson.DocElem{Name: "name", Value: "personA"},
			bson.DocElem{Name: "orderid", Value: 1},
			bson.DocElem{Name: "_id", Value: 1},
		},
		bson.D{
			bson.DocElem{Name: "name", Value: "personB"},
			bson.DocElem{Name: "orderid", Value: 2},
			bson.DocElem{Name: "_id", Value: 2},
		},
		bson.D{
			bson.DocElem{Name: "name", Value: "personC"},
			bson.DocElem{Name: "orderid", Value: 3},
			bson.DocElem{Name: "_id", Value: 3},
		},
		bson.D{
			bson.DocElem{Name: "name", Value: "personD"},
			bson.DocElem{Name: "orderid", Value: 4},
			bson.DocElem{Name: "_id", Value: 4},
		},
	}

	orders = []bson.D{
		bson.D{
			bson.DocElem{Name: "orderid", Value: 1},
			bson.DocElem{Name: "amount", Value: 1000},
			bson.DocElem{Name: "_id", Value: 1},
		},
		bson.D{
			bson.DocElem{Name: "orderid", Value: 1},
			bson.DocElem{Name: "amount", Value: 450},
			bson.DocElem{Name: "_id", Value: 2},
		},
		bson.D{
			bson.DocElem{Name: "orderid", Value: 2},
			bson.DocElem{Name: "amount", Value: 1300},
			bson.DocElem{Name: "_id", Value: 3},
		},
		bson.D{
			bson.DocElem{Name: "orderid", Value: 4},
			bson.DocElem{Name: "amount", Value: 390},
			bson.DocElem{Name: "_id", Value: 4},
		},
		bson.D{
			bson.DocElem{Name: "orderid", Value: 5},
			bson.DocElem{Name: "amount", Value: 760},
			bson.DocElem{Name: "_id", Value: 5},
		},
	}
)

func setupJoinOperator(on SQLExpr, kind JoinKind) PlanStage {

	ms1 := &BSONSourceStage{tableOneName, customers}
	ms2 := &BSONSourceStage{tableTwoName, orders}

	return &JoinStage{
		left:    ms1,
		right:   ms2,
		matcher: on,
		kind:    kind,
	}

}

func TestJoinPlanStage(t *testing.T) {
	Convey("With a simple test configuration...", t, func() {

		criteria := &SQLEqualsExpr{
			left: &SQLColumnExpr{
				tableName:  tableOneName,
				columnName: "orderid",
				columnType: schema.ColumnType{
					SQLType:   schema.SQLInt,
					MongoType: schema.MongoInt,
				},
			},
			right: &SQLColumnExpr{
				tableName:  tableTwoName,
				columnName: "orderid",
				columnType: schema.ColumnType{
					SQLType:   schema.SQLInt,
					MongoType: schema.MongoInt,
				},
			},
		}

		ctx := &ExecutionCtx{}

		row := &Row{}

		i := 0

		Convey("an inner join should return correct results", func() {

			operator := setupJoinOperator(criteria, sqlparser.AST_JOIN)

			iter, err := operator.Open(ctx)
			So(err, ShouldBeNil)

			expectedResults := []struct {
				Name   interface{}
				Amount interface{}
			}{
				{"personA", 1000},
				{"personA", 450},
				{"personB", 1300},
				{"personD", 390},
			}

			for iter.Next(row) {
				So(len(row.Data), ShouldEqual, 6)
				So(row.Data[0].Table, ShouldEqual, tableOneName)
				So(row.Data[4].Table, ShouldEqual, tableTwoName)
				So(row.Data[0].Data, ShouldEqual, expectedResults[i].Name)
				So(row.Data[4].Data, ShouldEqual, expectedResults[i].Amount)
				i++
			}

			So(i, ShouldEqual, 4)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)

		})

		Convey("an left join should return correct results", func() {

			operator := setupJoinOperator(criteria, sqlparser.AST_LEFT_JOIN)

			iter, err := operator.Open(ctx)
			So(err, ShouldBeNil)

			expectedResults := []struct {
				Name   interface{}
				Amount interface{}
			}{
				{"personA", 1000},
				{"personA", 450},
				{"personB", 1300},
				{"personC", nil},
				{"personD", 390},
			}

			for iter.Next(row) {
				// left entry with no corresponding right entry
				if i == 3 {
					So(len(row.Data), ShouldEqual, 3)
					So(row.Data[0].Table, ShouldEqual, tableOneName)
					So(row.Data[0].Data, ShouldEqual, expectedResults[i].Name)
				} else {
					So(len(row.Data), ShouldEqual, 6)
					So(row.Data[0].Table, ShouldEqual, tableOneName)
					So(row.Data[4].Table, ShouldEqual, tableTwoName)
					So(row.Data[0].Data, ShouldEqual, expectedResults[i].Name)
					So(row.Data[4].Data, ShouldEqual, expectedResults[i].Amount)
				}
				i++

			}

			So(i, ShouldEqual, 5)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)

		})

		Convey("an right join should return correct results", func() {

			operator := setupJoinOperator(criteria, sqlparser.AST_RIGHT_JOIN)

			iter, err := operator.Open(ctx)
			So(err, ShouldBeNil)

			expectedResults := []struct {
				Name   interface{}
				Amount interface{}
			}{
				{"personA", 1000},
				{"personA", 450},
				{"personB", 1300},
				{"personD", 390},
				{nil, 760},
			}

			for iter.Next(row) {
				// right entry with no corresponding left entry
				if i == 4 {
					So(len(row.Data), ShouldEqual, 3)
					So(row.Data[4].Table, ShouldEqual, tableTwoName)
					So(row.Data[4].Data, ShouldEqual, expectedResults[i].Amount)
				} else {
					So(len(row.Data), ShouldEqual, 6)
					So(row.Data[0].Table, ShouldEqual, tableOneName)
					So(row.Data[4].Table, ShouldEqual, tableTwoName)
					So(row.Data[0].Data, ShouldEqual, expectedResults[i].Name)
					So(row.Data[4].Data, ShouldEqual, expectedResults[i].Amount)
				}
				i++
			}

			So(i, ShouldEqual, 5)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)

		})

		Convey("a cross join should return correct results", func() {

			operator := setupJoinOperator(criteria, sqlparser.AST_CROSS_JOIN)

			iter, err := operator.Open(ctx)
			So(err, ShouldBeNil)

			expectedNames := []string{"personA", "personB", "personC", "personD", "personE"}
			expectedAmounts := []int{1000, 450, 1300, 390, 760}

			for iter.Next(row) {
				So(len(row.Data), ShouldEqual, 6)
				So(row.Data[0].Table, ShouldEqual, tableOneName)
				So(row.Data[4].Table, ShouldEqual, tableTwoName)
				So(row.Data[0].Data, ShouldEqual, expectedNames[i/5])
				So(row.Data[4].Data, ShouldEqual, expectedAmounts[i%5])
				i++
			}

			So(i, ShouldEqual, 20)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)

		})
	})
}
