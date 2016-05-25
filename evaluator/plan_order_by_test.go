package evaluator

import (
	"testing"

	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestOrderByOperator(t *testing.T) {
	columnType := schema.ColumnType{schema.SQLInt, schema.MongoInt}

	runTest := func(orderby *OrderByStage, rows []bson.D, expectedRows []Values) {

		ctx := &ExecutionCtx{}

		ts := &BSONSourceStage{tableOneName, nil}

		source := &ProjectStage{
			source: ts,
			sExprs: SelectExpressions{
				SelectExpression{
					Column: &Column{tableOneName, "a", "a", schema.SQLInt, schema.MongoInt},
					Expr:   SQLColumnExpr{tableOneName, "a", columnType},
				},
				SelectExpression{
					Column: &Column{tableOneName, "b", "b", schema.SQLInt, schema.MongoInt},
					Expr:   SQLColumnExpr{tableOneName, "b", columnType},
				},
			},
		}

		orderby.source = source
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
					{expr: SQLColumnExpr{tableOneName, "a", columnType}, ascending: false},
				}

				operator := &OrderByStage{
					terms: terms,
				}

				expected := []Values{
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
				}

				runTest(operator, data, expected)

			})

			Convey("desc", func() {

				terms := []*orderByTerm{
					{expr: SQLColumnExpr{tableOneName, "a", columnType}, ascending: false},
				}

				operator := &OrderByStage{
					terms: terms,
				}

				expected := []Values{
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
				}

				runTest(operator, data, expected)

			})

		})

		Convey("multiple sort keys should sort according to the direction specified", func() {

			Convey("asc + asc", func() {
				terms := []*orderByTerm{
					{expr: SQLColumnExpr{tableOneName, "a", columnType}, ascending: false},
					{expr: SQLColumnExpr{tableOneName, "b", columnType}, ascending: false},
				}

				expected := []Values{
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
				}

				operator := &OrderByStage{
					terms: terms,
				}

				runTest(operator, data, expected)
			})

			Convey("asc + desc", func() {
				terms := []*orderByTerm{
					{expr: SQLColumnExpr{tableOneName, "a", columnType}, ascending: false},
					{expr: SQLColumnExpr{tableOneName, "b", columnType}, ascending: false},
				}

				operator := &OrderByStage{
					terms: terms,
				}

				expected := []Values{
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
				}

				runTest(operator, data, expected)

			})

			Convey("desc + asc", func() {
				terms := []*orderByTerm{
					{expr: SQLColumnExpr{tableOneName, "a", columnType}, ascending: false},
					{expr: SQLColumnExpr{tableOneName, "b", columnType}, ascending: false},
				}

				operator := &OrderByStage{
					terms: terms,
				}

				expected := []Values{
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
				}

				runTest(operator, data, expected)

			})

			Convey("desc + desc", func() {
				terms := []*orderByTerm{
					{expr: SQLColumnExpr{tableOneName, "a", columnType}, ascending: false},
					{expr: SQLColumnExpr{tableOneName, "b", columnType}, ascending: false},
				}

				operator := &OrderByStage{
					terms: terms,
				}

				expected := []Values{
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
				}

				runTest(operator, data, expected)

			})
		})

	})
}
