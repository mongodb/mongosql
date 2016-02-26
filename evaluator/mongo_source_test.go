package evaluator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestMongoSourceOperator(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne
	sessionProvider, err := NewSessionProvider(cfgOne)
	if err != nil {
		t.Fatalf("failed to set up session provider to test server: %v", err)
		return
	}

	session := sessionProvider.GetSession()
	collectionTwo := session.DB(dbOne).C(tableTwoName)

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

			operator, err := NewMongoSource(ctx, dbOne, tableTwoName, "")
			So(err, ShouldBeNil)
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
			val, ok := extractFieldByName("a", testD)
			So(val, ShouldEqual, "string")
			So(ok, ShouldBeTrue)
		})

		Convey("array fields should be extracted by name", func() {
			val, ok := extractFieldByName("b.1", testD)
			So(val, ShouldResemble, bson.D{{"inner2", 1}})
			So(ok, ShouldBeTrue)
			val, ok = extractFieldByName("b.1.inner2", testD)
			So(val, ShouldEqual, 1)
			So(ok, ShouldBeTrue)
			val, ok = extractFieldByName("b.0", testD)
			So(val, ShouldEqual, "inner")
			So(ok, ShouldBeTrue)
		})

		Convey("subdocument fields should be extracted by name", func() {
			val, ok := extractFieldByName("c", testD)
			So(val, ShouldResemble, bson.D{{"x", 5}})
			So(ok, ShouldBeTrue)
			val, ok = extractFieldByName("c.x", testD)
			So(val, ShouldEqual, 5)
			So(ok, ShouldBeTrue)

			Convey("even if they contain null values", func() {
				val, ok := extractFieldByName("d", testD)
				So(val, ShouldResemble, bson.D{{"z", nil}})
				So(ok, ShouldBeTrue)
				val, ok = extractFieldByName("d.z", testD)
				So(val, ShouldEqual, nil)
				So(ok, ShouldBeTrue)
				val, ok = extractFieldByName("d.z.nope", testD)
				So(val, ShouldEqual, nil)
				So(ok, ShouldBeFalse)
			})
		})

		Convey(`non-existing fields should return (nil,false)`, func() {
			for _, c := range []string{"f", "c.nope", "c.nope.NOPE", "b.1000", "b.1.nada"} {
				val, ok := extractFieldByName(c, testD)
				So(val, ShouldBeNil)
				So(ok, ShouldBeFalse)
			}
		})

	})

	Convey(`Extraction of a non-document should return (nil, false)`, t, func() {
		val, ok := extractFieldByName("meh", []interface{}{"meh"})
		So(val, ShouldBeNil)
		So(ok, ShouldBeFalse)
	})

	Convey(`Extraction of a nil document should return (nil, false)`, t, func() {
		val, ok := extractFieldByName("a", nil)
		So(val, ShouldEqual, nil)
		So(ok, ShouldBeFalse)
	})
}
