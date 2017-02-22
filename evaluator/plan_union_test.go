package evaluator

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

const (
	SQL1  = SQLInt(1)
	SQL2  = SQLInt(2)
	SQL3  = SQLInt(3)
	SQL4  = SQLInt(4)
	SQL5  = SQLInt(5)
	SQL6  = SQLInt(6)
	SQL7  = SQLInt(7)
	SQL8  = SQLInt(8)
	SQL9  = SQLInt(9)
	SQL10 = SQLInt(10)
	SQL11 = SQLInt(11)
	SQL12 = SQLInt(12)
	SQL13 = SQLInt(13)
	SQL14 = SQLInt(14)
	SQL15 = SQLInt(15)
	SQL16 = SQLInt(16)
)

var (
	basicTable1 = []bson.D{
		bson.D{
			bson.DocElem{Name: "a", Value: 1},
			bson.DocElem{Name: "b", Value: 2},
		},
		bson.D{
			bson.DocElem{Name: "a", Value: 3},
			bson.DocElem{Name: "b", Value: 4},
		},
		bson.D{
			bson.DocElem{Name: "a", Value: 5},
			bson.DocElem{Name: "b", Value: 6},
		},
		bson.D{
			bson.DocElem{Name: "a", Value: 7},
			bson.DocElem{Name: "b", Value: 8},
		},
	}

	basicTable2 = []bson.D{
		bson.D{
			bson.DocElem{Name: "c", Value: 9},
			bson.DocElem{Name: "d", Value: 10},
		},
		bson.D{
			bson.DocElem{Name: "c", Value: 11},
			bson.DocElem{Name: "d", Value: 12},
		},
		bson.D{
			bson.DocElem{Name: "c", Value: 13},
			bson.DocElem{Name: "d", Value: 14},
		},
		bson.D{
			bson.DocElem{Name: "c", Value: 15},
			bson.DocElem{Name: "d", Value: 16},
		},
	}
)

type result map[string]interface{}

func containsRow(results []result, row *Row) ([]result, bool) {
	toRemove := -1

	contains := false
	for i, result := range results {
		matches := true
		for _, value := range row.Data {
			result_val, ok := result[value.Name]
			So(ok, ShouldBeTrue)
			if result_val != value.Data {
				matches = false
				break
			}
		}

		if matches {
			toRemove = i
			contains = true
			break
		}
	}

	// necessary for multiset equality
	if toRemove >= 0 {
		results = append(results[:toRemove], results[toRemove+1:]...)
	}

	return results, contains
}

func TestUnionPlanStage(t *testing.T) {

	testSchema, err := schema.New(testSchema4)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}

	testInfo := getMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)

	ctx := createTestExecutionCtx(testInfo)

	test := func(testName string, expectedColumns []string, expectedResults []result, planStageFactory func() PlanStage) {
		Convey(testName, func() {
			row := &Row{}

			unionStage := planStageFactory()
			iter, err := unionStage.Open(ctx)
			So(err, ShouldBeNil)

			columns := unionStage.Columns()
			So(len(expectedColumns), ShouldEqual, len(columns))
			for i, col := range columns {
				So(col.Name, ShouldEqual, expectedColumns[i])
			}

			for iter.Next(row) {
				trimmed, contains := containsRow(expectedResults, row)
				expectedResults = trimmed
				So(contains, ShouldBeTrue)
			}
			So(expectedResults, ShouldBeEmpty)

			err = iter.Err()
			So(err, ShouldBeNil)

			err = iter.Close()
			So(err, ShouldBeNil)
		})
	}

	Convey("Union Plan Stage", t, func() {
		test("a b before c d",
			[]string{"a", "b"},
			[]result{{"a": SQL1, "b": SQL2}, {"a": SQL3, "b": SQL4},
				{"a": SQL5, "b": SQL6}, {"a": SQL7, "b": SQL8},
				{"a": SQL9, "b": SQL10}, {"a": SQL11, "b": SQL12},
				{"a": SQL13, "b": SQL14}, {"a": SQL15, "b": SQL16}},
			func() PlanStage {
				return NewUnionStage(UnionDistinct,
					NewBSONSourceStage(1, "foo", collation.Default, basicTable1),
					NewBSONSourceStage(2, "bar", collation.Default, basicTable2),
				)
			})

		test("c d before a b",
			[]string{"c", "d"},
			[]result{{"c": SQL1, "d": SQL2}, {"c": SQL3, "d": SQL4},
				{"c": SQL5, "d": SQL6}, {"c": SQL7, "d": SQL8},
				{"c": SQL9, "d": SQL10}, {"c": SQL11, "d": SQL12},
				{"c": SQL13, "d": SQL14}, {"c": SQL15, "d": SQL16}},
			func() PlanStage {
				return NewUnionStage(UnionDistinct,
					NewBSONSourceStage(1, "foo", collation.Default, basicTable2),
					NewBSONSourceStage(2, "bar", collation.Default, basicTable1),
				)
			})
	})
}
