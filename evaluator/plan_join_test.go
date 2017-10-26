package evaluator

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/10gen/mongo-go-driver/bson"
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

func setupJoinOperator(on SQLExpr, kind joinKind) PlanStage {

	ms1 := NewBSONSourceStage(1, tableOneName, collation.Default, customers)
	ms2 := NewBSONSourceStage(1, tableTwoName, collation.Default, orders)

	return &JoinStage{
		left:    ms1,
		right:   ms2,
		matcher: on,
		kind:    kind,
	}

}

func TestJoinPlanStage(t *testing.T) {

	testSchema, err := schema.New(testSchema4)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}

	testInfo := getMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	Convey("Subject: JoinStage", t, func() {

		criteria := &SQLEqualsExpr{
			left: &SQLColumnExpr{
				selectID:   1,
				tableName:  tableOneName,
				columnName: "orderid",
				columnType: schema.ColumnType{
					SQLType:   schema.SQLInt,
					MongoType: schema.MongoInt,
				},
			},
			right: &SQLColumnExpr{
				selectID:   1,
				tableName:  tableTwoName,
				columnName: "orderid",
				columnType: schema.ColumnType{
					SQLType:   schema.SQLInt,
					MongoType: schema.MongoInt,
				},
			},
		}

		ctx := createTestExecutionCtx(testInfo)

		row := &Row{}

		i := 0

		Convey("an inner join should return correct results", func() {

			operator := setupJoinOperator(criteria, innerJoin)

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

		Convey("a left join should return correct results", func() {

			operator := setupJoinOperator(criteria, leftJoin)

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
				So(len(row.Data), ShouldEqual, 6)
				So(row.Data[0].Table, ShouldEqual, tableOneName)
				So(row.Data[4].Table, ShouldEqual, tableTwoName)
				So(row.Data[0].Data, ShouldEqual, expectedResults[i].Name)
				So(row.Data[4].Data, ShouldEqual, expectedResults[i].Amount)
				i++
			}

			So(i, ShouldEqual, 5)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)

		})

		Convey("a right join should return correct results", func() {

			operator := setupJoinOperator(criteria, rightJoin)

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
				So(len(row.Data), ShouldEqual, 6)
				So(row.Data[0].Table, ShouldEqual, tableOneName)
				So(row.Data[4].Table, ShouldEqual, tableTwoName)
				So(row.Data[0].Data, ShouldEqual, expectedResults[i].Name)
				So(row.Data[4].Data, ShouldEqual, expectedResults[i].Amount)
				i++
			}

			So(i, ShouldEqual, 5)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)

		})

		Convey("a cross join should return correct results", func() {

			operator := setupJoinOperator(criteria, crossJoin)

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

		Convey("a straight join should return correct results", func() {

			operator := setupJoinOperator(criteria, straightJoin)

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
	})
}

func TestJoinPlanStage_MemoryLimits(t *testing.T) {

	testSchema, err := schema.New(testSchema4)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}

	testInfo := getMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	Convey("Subject: JoinStage Memory Limits", t, func() {

		criteria := &SQLEqualsExpr{
			left: &SQLColumnExpr{
				selectID:   1,
				tableName:  tableOneName,
				columnName: "orderid",
				columnType: schema.ColumnType{
					SQLType:   schema.SQLInt,
					MongoType: schema.MongoInt,
				},
			},
			right: &SQLColumnExpr{
				selectID:   1,
				tableName:  tableTwoName,
				columnName: "orderid",
				columnType: schema.ColumnType{
					SQLType:   schema.SQLInt,
					MongoType: schema.MongoInt,
				},
			},
		}

		ctx := createTestExecutionCtx(testInfo)
		ctx.Variables().SetSystemVariable(variable.MongoDBMaxStageSize, 500)

		row := &Row{}

		Convey("inner join", func() {

			operator := setupJoinOperator(criteria, innerJoin)

			iter, err := operator.Open(ctx)
			So(err, ShouldBeNil)

			ok := iter.Next(row)
			So(ok, ShouldBeFalse)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldNotBeNil)
		})

		Convey("left join", func() {

			operator := setupJoinOperator(criteria, leftJoin)

			iter, err := operator.Open(ctx)
			So(err, ShouldBeNil)

			ok := iter.Next(row)
			So(ok, ShouldBeFalse)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldNotBeNil)
		})

		Convey("right join", func() {

			operator := setupJoinOperator(criteria, rightJoin)

			iter, err := operator.Open(ctx)
			So(err, ShouldBeNil)

			ok := iter.Next(row)
			So(ok, ShouldBeFalse)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldNotBeNil)
		})

		Convey("cross join", func() {

			operator := setupJoinOperator(nil, rightJoin)

			iter, err := operator.Open(ctx)
			So(err, ShouldBeNil)

			ok := iter.Next(row)
			So(ok, ShouldBeFalse)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldNotBeNil)
		})
	})
}
