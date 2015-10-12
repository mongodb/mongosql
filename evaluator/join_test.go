package evaluator

import (
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/config"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var (
	customers = []interface{}{
		bson.D{
			bson.DocElem{Name: "name", Value: "personA"},
			bson.DocElem{Name: "orderid", Value: 1},
		},
		bson.D{
			bson.DocElem{Name: "name", Value: "personB"},
			bson.DocElem{Name: "orderid", Value: 2},
		},
		bson.D{
			bson.DocElem{Name: "name", Value: "personC"},
			bson.DocElem{Name: "orderid", Value: 3},
		},
		bson.D{
			bson.DocElem{Name: "name", Value: "personD"},
			bson.DocElem{Name: "orderid", Value: 4},
		},
	}

	orders = []interface{}{
		bson.D{
			bson.DocElem{Name: "orderid", Value: 1},
			bson.DocElem{Name: "amount", Value: 1000},
		},
		bson.D{
			bson.DocElem{Name: "orderid", Value: 1},
			bson.DocElem{Name: "amount", Value: 450},
		},
		bson.D{
			bson.DocElem{Name: "orderid", Value: 2},
			bson.DocElem{Name: "amount", Value: 1300},
		},
		bson.D{
			bson.DocElem{Name: "orderid", Value: 4},
			bson.DocElem{Name: "amount", Value: 390},
		},
		bson.D{
			bson.DocElem{Name: "orderid", Value: 5},
			bson.DocElem{Name: "amount", Value: 760},
		},
	}
)

func setupJoinOperator(criteria sqlparser.BoolExpr, joinType string) Operator {

	cfg, err := config.ParseConfigData(testConfig1)
	So(err, ShouldBeNil)

	session, err := mgo.Dial(cfg.Url)
	So(err, ShouldBeNil)

	c1 := session.DB(dbName).C(tableOneName)
	c1.DropCollection()

	for _, customer := range customers {
		So(c1.Insert(customer), ShouldBeNil)
	}

	c2 := session.DB(dbName).C(tableTwoName)
	c2.DropCollection()

	for _, order := range orders {
		So(c2.Insert(order), ShouldBeNil)
	}

	ts1 := &TableScan{
		tableName: tableOneName,
	}

	ts2 := &TableScan{
		tableName: tableTwoName,
	}

	session.Close()

	return &Join{
		left:  ts1,
		right: ts2,
		on:    criteria,
		kind:  joinType,
	}

}

func TestJoinOperator(t *testing.T) {

	Convey("With a simple test configuration...", t, func() {

		criteria := &sqlparser.ComparisonExpr{
			Operator: sqlparser.AST_EQ,
			Left: &sqlparser.ColName{
				Name:      []byte("orderid"),
				Qualifier: []byte(tableOneName),
			},
			Right: &sqlparser.ColName{
				Name:      []byte("orderid"),
				Qualifier: []byte(tableTwoName),
			},
		}

		cfg, err := config.ParseConfigData(testConfig1)
		So(err, ShouldBeNil)

		ctx := &ExecutionCtx{
			Config: cfg,
			Db:     dbName,
		}

		row := &Row{}

		i := 0

		Convey("an inner join should return correct results", func() {

			operator := setupJoinOperator(criteria, sqlparser.AST_JOIN)

			So(operator.Open(ctx), ShouldBeNil)

			expectedResults := []struct {
				Name   string
				Amount int
			}{
				{"personA", 1000},
				{"personA", 450},
				{"personB", 1300},
				{"personD", 390},
			}

			for operator.Next(row) {
				So(len(row.Data), ShouldEqual, 2)
				So(row.Data[0].Table, ShouldEqual, tableOneName)
				So(row.Data[1].Table, ShouldEqual, tableTwoName)
				So(row.Data[0].Values.Map()["name"], ShouldEqual, expectedResults[i].Name)
				So(row.Data[1].Values.Map()["amount"], ShouldEqual, expectedResults[i].Amount)
				i++
			}

			So(i, ShouldEqual, 4)

			So(operator.Err(), ShouldBeNil)
			So(operator.Close(), ShouldBeNil)

		})

		Convey("an left join should return correct results", func() {

			operator := setupJoinOperator(criteria, sqlparser.AST_LEFT_JOIN)

			So(operator.Open(ctx), ShouldBeNil)

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

			for operator.Next(row) {
				// left entry with no corresponding right entry
				if i == 3 {
					So(len(row.Data), ShouldEqual, 1)
					So(row.Data[0].Table, ShouldEqual, tableOneName)
					So(row.Data[0].Values.Map()["name"], ShouldEqual, expectedResults[i].Name)
					So(row.Data[0].Values.Map()["amount"], ShouldEqual, expectedResults[i].Amount)
				} else {
					So(len(row.Data), ShouldEqual, 2)
					So(row.Data[0].Table, ShouldEqual, tableOneName)
					So(row.Data[1].Table, ShouldEqual, tableTwoName)
					So(row.Data[0].Values.Map()["name"], ShouldEqual, expectedResults[i].Name)
					So(row.Data[1].Values.Map()["amount"], ShouldEqual, expectedResults[i].Amount)
				}
				i++

			}

			So(i, ShouldEqual, 5)

			So(operator.Err(), ShouldBeNil)
			So(operator.Close(), ShouldBeNil)

		})

		Convey("an right join should return correct results", func() {

			operator := setupJoinOperator(criteria, sqlparser.AST_RIGHT_JOIN)

			So(operator.Open(ctx), ShouldBeNil)

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

			i := 0

			for operator.Next(row) {
				// right entry with no corresponding left entry
				if i == 4 {
					So(len(row.Data), ShouldEqual, 1)
					So(row.Data[0].Table, ShouldEqual, tableTwoName)
					So(row.Data[0].Values.Map()["name"], ShouldEqual, expectedResults[i].Name)
					So(row.Data[0].Values.Map()["amount"], ShouldEqual, expectedResults[i].Amount)
				} else {
					So(len(row.Data), ShouldEqual, 2)
					So(row.Data[0].Table, ShouldEqual, tableTwoName)
					So(row.Data[1].Table, ShouldEqual, tableOneName)
					So(row.Data[1].Values.Map()["name"], ShouldEqual, expectedResults[i].Name)
					So(row.Data[0].Values.Map()["amount"], ShouldEqual, expectedResults[i].Amount)
				}
				i++

			}

			So(i, ShouldEqual, 5)

			So(operator.Err(), ShouldBeNil)
			So(operator.Close(), ShouldBeNil)

		})

		Convey("a cross join should return correct results", func() {

			operator := setupJoinOperator(criteria, sqlparser.AST_CROSS_JOIN)

			So(operator.Open(ctx), ShouldBeNil)

			expectedNames := []string{"personA", "personB", "personC", "personD", "personE"}
			expectedAmounts := []int{1000, 450, 1300, 390, 760}

			i := 0

			for operator.Next(row) {
				So(len(row.Data), ShouldEqual, 2)
				So(row.Data[0].Table, ShouldEqual, tableOneName)
				So(row.Data[1].Table, ShouldEqual, tableTwoName)
				So(row.Data[0].Values.Map()["name"], ShouldEqual, expectedNames[i/5])
				So(row.Data[1].Values.Map()["amount"], ShouldEqual, expectedAmounts[i%5])
				i++
			}

			So(i, ShouldEqual, 20)

			So(operator.Err(), ShouldBeNil)
			So(operator.Close(), ShouldBeNil)

		})
	})
}
