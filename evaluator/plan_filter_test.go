package evaluator_test

import (
	"fmt"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	_ fmt.Stringer = nil
)

func TestFilterPlanStage(t *testing.T) {
	runTest := func(matcher evaluator.SQLExpr, rows []bson.D, expectedRows []evaluator.Values) {

		ctx := &evaluator.ExecutionCtx{}

		bss := evaluator.NewBSONSourceStage(1, tableTwoName, collation.Default, rows)
		filter := evaluator.NewFilterStage(bss, matcher)

		iter, err := filter.Open(ctx)

		So(err, ShouldBeNil)

		row := &evaluator.Row{}

		i := 0

		for iter.Next(row) {
			So(len(row.Data), ShouldEqual, len(expectedRows[i]))
			So(row.Data, ShouldResemble, expectedRows[i])
			row = &evaluator.Row{}
			i++
		}

		So(i, ShouldEqual, len(expectedRows))

		So(iter.Close(), ShouldBeNil)
		So(iter.Err(), ShouldBeNil)
	}

	Convey("With a simple test configuration...", t, func() {

		schema := evaluator.MustLoadSchema(testSchema3)

		rows := []bson.D{
			{{Name: "a", Value: 6}, {Name: "b", Value: 7}, {Name: "_id", Value: 5}},
			{{Name: "a", Value: 16}, {Name: "b", Value: 17}, {Name: "_id", Value: 15}},
		}

		Convey("a filter operator should only return rows that match", func() {
			queries := []string{
				"a = 16",
				"a = 6",
				"a = 99",
				"b > 9",
				"b > 9 or a < 5",
				"b = 7 or a = 6",
			}

			r0, err := bsonDToValues(1, evaluator.BSONSourceDB, tableTwoName, rows[0])
			So(err, ShouldBeNil)
			r1, err := bsonDToValues(1, evaluator.BSONSourceDB, tableTwoName, rows[1])
			So(err, ShouldBeNil)

			expected := [][]evaluator.Values{{r1}, {r0}, nil, {r1}, {r1}, {r0}}

			for i, query := range queries {
				matcher,
					err := evaluator.GetSQLExpr(schema,
					evaluator.BSONSourceDB,
					tableTwoName,
					query)
				So(err, ShouldBeNil)

				runTest(matcher, rows, expected[i])
			}
		})
	})
}
