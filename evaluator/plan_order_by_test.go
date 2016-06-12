package evaluator

import (
	"testing"

	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestOrderByOperator(t *testing.T) {
	runTest := func(orderby *OrderByStage, rows []bson.D, expectedRows []Values) {

		ctx := &ExecutionCtx{}

		ts := NewBSONSourceStage(1, tableOneName, nil)

		orderby.source = ts
		iter, err := orderby.Open(ctx)
		So(err, ShouldBeNil)

		row := &Row{}

		i := 0

		for iter.Next(row) {
			So(len(row.Data), ShouldEqual, 1)
			So(row.Data[0], ShouldResemble, expectedRows[i])
			row = &Row{}
			i++
		}

		So(iter.Close(), ShouldBeNil)
		So(iter.Err(), ShouldBeNil)
	}

	Convey("An order by operator...", t, func() {

		data := []bson.D{
			bson.D{{"_id", 1}, {"a", 6}, {"b", 7}},
			bson.D{{"_id", 2}, {"a", 6}, {"b", 8}},
			bson.D{{"_id", 3}, {"a", 7}, {"b", 8}},
			bson.D{{"_id", 4}, {"a", 7}, {"b", 7}},
		}

		Convey("single sort keys should sort according to the direction specified", func() {

			Convey("asc", func() {

				terms := []*orderByTerm{
					{expr: NewSQLColumnExpr(0, tableOneName, "a", schema.SQLInt, schema.MongoInt), ascending: false},
				}

				operator := &OrderByStage{
					terms: terms,
				}

				expected := []Values{
					{{1, "a", "a", SQLInt(6)}, {1, "b", "b", SQLInt(7)}},
					{{1, "a", "a", SQLInt(6)}, {1, "b", "b", SQLInt(8)}},
					{{1, "a", "a", SQLInt(7)}, {1, "b", "b", SQLInt(8)}},
					{{1, "a", "a", SQLInt(7)}, {1, "b", "b", SQLInt(7)}},
				}

				runTest(operator, data, expected)

			})

			Convey("desc", func() {

				terms := []*orderByTerm{
					{expr: NewSQLColumnExpr(0, tableOneName, "a", schema.SQLInt, schema.MongoInt), ascending: false},
				}

				operator := &OrderByStage{
					terms: terms,
				}

				expected := []Values{
					{{1, "a", "a", SQLInt(7)}, {1, "b", "b", SQLInt(8)}},
					{{1, "a", "a", SQLInt(7)}, {1, "b", "b", SQLInt(7)}},
					{{1, "a", "a", SQLInt(6)}, {1, "b", "b", SQLInt(7)}},
					{{1, "a", "a", SQLInt(6)}, {1, "b", "b", SQLInt(8)}},
				}

				runTest(operator, data, expected)

			})

		})

		Convey("multiple sort keys should sort according to the direction specified", func() {

			Convey("asc + asc", func() {
				terms := []*orderByTerm{
					{expr: NewSQLColumnExpr(0, tableOneName, "a", schema.SQLInt, schema.MongoInt), ascending: false},
					{expr: NewSQLColumnExpr(0, tableOneName, "b", schema.SQLInt, schema.MongoInt), ascending: false},
				}

				expected := []Values{
					{{1, "a", "a", SQLInt(6)}, {1, "b", "b", SQLInt(7)}},
					{{1, "a", "a", SQLInt(6)}, {1, "b", "b", SQLInt(8)}},
					{{1, "a", "a", SQLInt(7)}, {1, "b", "b", SQLInt(7)}},
					{{1, "a", "a", SQLInt(7)}, {1, "b", "b", SQLInt(8)}},
				}

				operator := &OrderByStage{
					terms: terms,
				}

				runTest(operator, data, expected)
			})

			Convey("asc + desc", func() {
				terms := []*orderByTerm{
					{expr: NewSQLColumnExpr(0, tableOneName, "a", schema.SQLInt, schema.MongoInt), ascending: false},
					{expr: NewSQLColumnExpr(0, tableOneName, "b", schema.SQLInt, schema.MongoInt), ascending: false},
				}

				operator := &OrderByStage{
					terms: terms,
				}

				expected := []Values{
					{{1, "a", "a", SQLInt(6)}, {1, "b", "b", SQLInt(8)}},
					{{1, "a", "a", SQLInt(6)}, {1, "b", "b", SQLInt(7)}},
					{{1, "a", "a", SQLInt(7)}, {1, "b", "b", SQLInt(8)}},
					{{1, "a", "a", SQLInt(7)}, {1, "b", "b", SQLInt(7)}},
				}

				runTest(operator, data, expected)

			})

			Convey("desc + asc", func() {
				terms := []*orderByTerm{
					{expr: NewSQLColumnExpr(0, tableOneName, "a", schema.SQLInt, schema.MongoInt), ascending: false},
					{expr: NewSQLColumnExpr(0, tableOneName, "b", schema.SQLInt, schema.MongoInt), ascending: false},
				}

				operator := &OrderByStage{
					terms: terms,
				}

				expected := []Values{
					{{1, "a", "a", SQLInt(7)}, {1, "b", "b", SQLInt(7)}},
					{{1, "a", "a", SQLInt(7)}, {1, "b", "b", SQLInt(8)}},
					{{1, "a", "a", SQLInt(6)}, {1, "b", "b", SQLInt(7)}},
					{{1, "a", "a", SQLInt(6)}, {1, "b", "b", SQLInt(8)}},
				}

				runTest(operator, data, expected)

			})

			Convey("desc + desc", func() {
				terms := []*orderByTerm{
					{expr: NewSQLColumnExpr(0, tableOneName, "a", schema.SQLInt, schema.MongoInt), ascending: false},
					{expr: NewSQLColumnExpr(0, tableOneName, "b", schema.SQLInt, schema.MongoInt), ascending: false},
				}

				operator := &OrderByStage{
					terms: terms,
				}

				expected := []Values{
					{{1, "a", "a", SQLInt(7)}, {1, "b", "b", SQLInt(8)}},
					{{1, "a", "a", SQLInt(7)}, {1, "b", "b", SQLInt(7)}},
					{{1, "a", "a", SQLInt(6)}, {1, "b", "b", SQLInt(8)}},
					{{1, "a", "a", SQLInt(6)}, {1, "b", "b", SQLInt(7)}},
				}

				runTest(operator, data, expected)

			})
		})

	})
}
