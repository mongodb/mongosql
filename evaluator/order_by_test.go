package evaluator

import (
	"github.com/erh/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func orderByTest(operator Operator, rows []bson.D, expectedRows []Values) {

	collectionOne.DropCollection()

	for _, row := range rows {
		So(collectionOne.Insert(row), ShouldBeNil)
	}

	ctx := &ExecutionCtx{
		Schema:  cfgOne,
		Db:      dbOne,
		Session: session,
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

	So(operator.Close(), ShouldBeNil)
	So(operator.Err(), ShouldBeNil)
}

func TestOrderByOperator(t *testing.T) {

	Convey("An order by operator...", t, func() {

		data := []bson.D{
			bson.D{{"_id", 1}, {"a", 6}, {"b", 7}},
			bson.D{{"_id", 2}, {"a", 6}, {"b", 8}},
			bson.D{{"_id", 3}, {"a", 7}, {"b", 8}},
			bson.D{{"_id", 4}, {"a", 7}, {"b", 7}},
		}

		source := &Select{
			source: &TableScan{
				tableName: tableOneName,
			},
			sExprs: SelectExpressions{
				SelectExpression{
					Column: Column{tableOneName, "a", "a", false},
					Expr:   &sqlparser.ColName{[]byte("a"), []byte(tableOneName)},
				},
				SelectExpression{
					Column: Column{tableOneName, "b", "b", false},
					Expr:   &sqlparser.ColName{[]byte("b"), []byte(tableOneName)},
				},
			},
		}

		Convey("single sort keys should sort according to the direction specified", func() {

			Convey("asc", func() {

				keys := []orderByKey{
					{SQLFieldExpr{tableOneName, "a"}, false, true, nil},
				}

				operator := &OrderBy{
					source: source,
					keys:   keys,
				}

				expected := []Values{
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
				}

				orderByTest(operator, data, expected)

			})

			Convey("desc", func() {

				keys := []orderByKey{
					{SQLFieldExpr{tableOneName, "a"}, false, false, nil},
				}

				operator := &OrderBy{
					source: source,
					keys:   keys,
				}

				expected := []Values{
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
				}

				orderByTest(operator, data, expected)

			})

		})

		Convey("multiple sort keys should sort according to the direction specified", func() {

			Convey("asc + asc", func() {
				keys := []orderByKey{
					{SQLFieldExpr{tableOneName, "a"}, false, true, nil},
					{SQLFieldExpr{tableOneName, "b"}, false, true, nil},
				}

				expected := []Values{
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
				}

				operator := &OrderBy{
					source: source,
					keys:   keys,
				}

				orderByTest(operator, data, expected)

			})

			Convey("asc + desc", func() {
				keys := []orderByKey{
					{SQLFieldExpr{tableOneName, "a"}, false, true, nil},
					{SQLFieldExpr{tableOneName, "b"}, false, false, nil},
				}

				operator := &OrderBy{
					source: source,
					keys:   keys,
				}

				expected := []Values{
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
				}

				orderByTest(operator, data, expected)

			})

			Convey("desc + asc", func() {
				keys := []orderByKey{
					{SQLFieldExpr{tableOneName, "a"}, false, false, nil},
					{SQLFieldExpr{tableOneName, "b"}, false, true, nil},
				}

				operator := &OrderBy{
					source: source,
					keys:   keys,
				}

				expected := []Values{
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
				}

				orderByTest(operator, data, expected)

			})

			Convey("desc + desc", func() {
				keys := []orderByKey{
					{SQLFieldExpr{tableOneName, "a"}, false, false, nil},
					{SQLFieldExpr{tableOneName, "b"}, false, false, nil},
				}

				operator := &OrderBy{
					source: source,
					keys:   keys,
				}

				expected := []Values{
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(7)}, {"b", "b", SQLInt(7)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(8)}},
					{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(7)}},
				}

				orderByTest(operator, data, expected)

			})
		})

	})
}
