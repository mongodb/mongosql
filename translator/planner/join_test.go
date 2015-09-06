package planner

import (
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/config"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var testCfg = []byte(
	`
schema :
-
  url: localhost
  db: test
  tables:
  -
     table: customers
     collection: test.customers
  tables:
  -
     table: orders
     collection: test.orders
`)

var (
	dbName = "test"
	colOne = "customers"
	colTwo = "orders"

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

	cfg, err := config.ParseConfigData(testCfg)
	So(err, ShouldBeNil)

	session, err := mgo.Dial(cfg.Url)
	So(err, ShouldBeNil)

	c1 := session.DB(dbName).C(colOne)
	c1.DropCollection()

	for _, customer := range customers {
		So(c1.Insert(customer), ShouldBeNil)
	}

	c2 := session.DB(dbName).C(colTwo)
	c2.DropCollection()

	for _, order := range orders {
		So(c2.Insert(order), ShouldBeNil)
	}

	ts1 := &TableScan{
		tableName: colOne,
	}

	ts2 := &TableScan{
		tableName: colTwo,
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
				Qualifier: []byte("customers"),
			},
			Right: &sqlparser.ColName{
				Name:      []byte("orderid"),
				Qualifier: []byte("orders"),
			},
		}

		cfg, err := config.ParseConfigData(testCfg)
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
				So(row.Data[0].Table, ShouldEqual, colOne)
				So(row.Data[1].Table, ShouldEqual, colTwo)
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
					So(row.Data[0].Table, ShouldEqual, colOne)
					So(row.Data[0].Values.Map()["name"], ShouldEqual, expectedResults[i].Name)
					So(row.Data[0].Values.Map()["amount"], ShouldEqual, expectedResults[i].Amount)
				} else {
					So(len(row.Data), ShouldEqual, 2)
					So(row.Data[0].Table, ShouldEqual, colOne)
					So(row.Data[1].Table, ShouldEqual, colTwo)
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
					So(row.Data[0].Table, ShouldEqual, colTwo)
					So(row.Data[0].Values.Map()["name"], ShouldEqual, expectedResults[i].Name)
					So(row.Data[0].Values.Map()["amount"], ShouldEqual, expectedResults[i].Amount)
				} else {
					So(len(row.Data), ShouldEqual, 2)
					So(row.Data[0].Table, ShouldEqual, colTwo)
					So(row.Data[1].Table, ShouldEqual, colOne)
					So(row.Data[1].Values.Map()["name"], ShouldEqual, expectedResults[i].Name)
					So(row.Data[0].Values.Map()["amount"], ShouldEqual, expectedResults[i].Amount)
				}
				i++

			}

			So(i, ShouldEqual, 5)

			So(operator.Err(), ShouldBeNil)
			So(operator.Close(), ShouldBeNil)

		})
	})
}
