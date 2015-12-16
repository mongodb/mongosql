package evaluator

import (
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func TestTableScanOperator(t *testing.T) {

	Convey("With a simple test configuration...", t, func() {

		Convey("fetching data from a table scan should return correct results in the right order", func() {

			rows := []bson.D{
				bson.D{
					bson.DocElem{Name: "a", Value: 6},
					bson.DocElem{Name: "b", Value: 7},
					bson.DocElem{Name: "_id", Value: 5},
				},
				bson.D{
					bson.DocElem{Name: "a", Value: 16},
					bson.DocElem{Name: "b", Value: 17},
					bson.DocElem{Name: "_id", Value: 15},
				},
			}

			var expected []Values
			for _, document := range rows {
				values, err := bsonDToValues(document)
				So(err, ShouldBeNil)
				expected = append(expected, values)
			}

			collectionTwo.DropCollection()

			for _, row := range rows {
				So(collectionTwo.Insert(row), ShouldBeNil)
			}

			ctx := &ExecutionCtx{
				Schema:  cfgOne,
				Db:      dbOne,
				Session: session,
			}

			operator := TableScan{
				tableName: tableTwoName,
			}

			So(operator.Open(ctx), ShouldBeNil)

			row := &Row{}

			i := 0

			for operator.Next(row) {
				So(len(row.Data), ShouldEqual, 1)
				So(row.Data[0].Table, ShouldEqual, tableTwoName)
				So(row.Data[0].Values, ShouldResemble, expected[i])
				row = &Row{}
				i++
			}

			So(operator.Close(), ShouldBeNil)
			So(operator.Err(), ShouldBeNil)
		})
	})
}

func TestExtractField(t *testing.T) {
	Convey("With a test bson.D", t, func() {
		testD := bson.D{
			{"a", "string"},
			{"b", []interface{}{"inner", bson.D{{"inner2", 1}}}},
			{"c", bson.D{{"x", 5}}},
			{"d", bson.D{{"z", nil}}},
		}

		Convey("regular fields should be extracted by name", func() {
			val := extractFieldByName("a", testD)
			So(val, ShouldEqual, "string")
		})

		Convey("array fields should be extracted by name", func() {
			val := extractFieldByName("b.1", testD)
			So(val, ShouldResemble, bson.D{{"inner2", 1}})
			val = extractFieldByName("b.1.inner2", testD)
			So(val, ShouldEqual, 1)
			val = extractFieldByName("b.0", testD)
			So(val, ShouldEqual, "inner")
		})

		Convey("subdocument fields should be extracted by name", func() {
			val := extractFieldByName("c", testD)
			So(val, ShouldResemble, bson.D{{"x", 5}})
			val = extractFieldByName("c.x", testD)
			So(val, ShouldEqual, 5)

			Convey("even if they contain null values", func() {
				val := extractFieldByName("d", testD)
				So(val, ShouldResemble, bson.D{{"z", nil}})
				val = extractFieldByName("d.z", testD)
				So(val, ShouldEqual, nil)
				val = extractFieldByName("d.z.nope", testD)
				So(val, ShouldEqual, "")
			})
		})

		Convey(`non-existing fields should return ""`, func() {
			val := extractFieldByName("f", testD)
			So(val, ShouldEqual, "")
			val = extractFieldByName("c.nope", testD)
			So(val, ShouldEqual, "")
			val = extractFieldByName("c.nope.NOPE", testD)
			So(val, ShouldEqual, "")
			val = extractFieldByName("b.1000", testD)
			So(val, ShouldEqual, "")
			val = extractFieldByName("b.1.nada", testD)
			So(val, ShouldEqual, "")
		})

	})

	Convey(`Extraction of a non-document should return ""`, t, func() {
		val := extractFieldByName("meh", []interface{}{"meh"})
		So(val, ShouldEqual, "")
	})

	Convey(`Extraction of a nil document should return ""`, t, func() {
		val := extractFieldByName("a", nil)
		So(val, ShouldEqual, "")
	})
}
