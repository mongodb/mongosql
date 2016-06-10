package mongo_test

import (
	_ "fmt"
	"github.com/10gen/sqlproxy/mongodrdl/mongo"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func TestSampling(t *testing.T) {

	Convey("Given an empty collection", t, func() {
		collection := mongo.NewCollection("test")

		Convey("Including a flat document", func() {
			err := collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", 1},
			})
			So(err, ShouldBeNil)

			Convey("Should result in 2 fields", func() {
				So(len(collection.Fields), ShouldEqual, 2)
				So(collection.Fields["a"].Count, ShouldEqual, 1)
				So(collection.Fields["b"].Count, ShouldEqual, 1)
			})

			Convey("And including an additional flat document with the same fields", func() {
				err := collection.IncludeSample(bson.D{
					{"a", 1},
					{"b", 1},
				})
				So(err, ShouldBeNil)

				Convey("Should increase the 2 fields counts", func() {
					So(len(collection.Fields), ShouldEqual, 2)
					So(collection.Fields["a"].Count, ShouldEqual, 2)
					So(collection.Fields["b"].Count, ShouldEqual, 2)
				})
			})

			Convey("And including an additional flat document with different fields", func() {
				err = collection.IncludeSample(bson.D{
					{"c", 2},
					{"d", 2},
				})
				So(err, ShouldBeNil)

				Convey("Should result in 4 fields", func() {
					So(len(collection.Fields), ShouldEqual, 4)
					So(collection.Fields["a"].Count, ShouldEqual, 1)
					So(collection.Fields["b"].Count, ShouldEqual, 1)
					So(collection.Fields["c"].Count, ShouldEqual, 1)
					So(collection.Fields["d"].Count, ShouldEqual, 1)
				})
			})

			Convey("And including an additional flat document with the same fields but different types", func() {
				err = collection.IncludeSample(bson.D{
					{"a", "string"},
					{"b", 3.2},
				})
				So(err, ShouldBeNil)
				err = collection.IncludeSample(bson.D{
					{"a", []interface{}{"string", 1}},
					{"b", bson.D{{"c", 1}}},
				})
				So(err, ShouldBeNil)

				Convey("Should result in 2 fields", func() {
					So(len(collection.Fields), ShouldEqual, 2)
					So(collection.Fields["a"].Count, ShouldEqual, 3)
					So(collection.Fields["b"].Count, ShouldEqual, 3)
				})

				Convey("And should result in each field having multiple types", func() {
					So(len(collection.Fields["a"].Types), ShouldEqual, 3)
					So(collection.Fields["a"].Types[0].Name(), ShouldEqual, "int")
					So(collection.Fields["a"].Types[1].Name(), ShouldEqual, "string")
					So(collection.Fields["a"].Types[2].Name(), ShouldEqual, "array")
					So(len(collection.Fields["b"].Types), ShouldEqual, 3)
					So(collection.Fields["b"].Types[0].Name(), ShouldEqual, "int")
					So(collection.Fields["b"].Types[1].Name(), ShouldEqual, "float64")
					So(collection.Fields["b"].Types[2].Name(), ShouldEqual, "document")
				})
			})
		})

		Convey("Including a document with a nested document", func() {
			err := collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", bson.D{
					{"c", 1},
					{"d", 1},
				}},
			})
			So(err, ShouldBeNil)

			Convey("Should result in 2 fields", func() {
				So(len(collection.Fields), ShouldEqual, 2)
				So(collection.Fields["a"].Count, ShouldEqual, 1)
				So(collection.Fields["b"].Count, ShouldEqual, 1)

				Convey("And the nested document should have 2 fields", func() {
					doc, ok := collection.Fields["b"].Types[0].(*mongo.Document)
					So(ok, ShouldBeTrue)
					So(doc.Fields["c"].Count, ShouldEqual, 1)
					So(doc.Fields["d"].Count, ShouldEqual, 1)
				})
			})

			Convey("And including an additional document with the same structure", func() {
				err = collection.IncludeSample(bson.D{
					{"a", 1},
					{"b", bson.D{
						{"c", "string"},
						{"d", 1},
					}},
				})
				So(err, ShouldBeNil)

				Convey("Should result in 2 fields", func() {
					So(len(collection.Fields), ShouldEqual, 2)
					So(collection.Fields["a"].Count, ShouldEqual, 2)
					So(collection.Fields["b"].Count, ShouldEqual, 2)

					Convey("And the nested document should have 2 fields", func() {
						doc, ok := collection.Fields["b"].Types[0].(*mongo.Document)
						So(ok, ShouldBeTrue)
						So(doc.Fields["c"].Count, ShouldEqual, 2)
						So(doc.Fields["d"].Count, ShouldEqual, 2)
					})
				})
			})

			Convey("And including an additional document with a different structure", func() {
				err = collection.IncludeSample(bson.D{
					{"c", 1},
					{"b", bson.D{
						{"c", "string"},
						{"e", 1},
					}},
				})
				So(err, ShouldBeNil)

				Convey("Should result in 3 fields", func() {
					So(len(collection.Fields), ShouldEqual, 3)
					So(collection.Fields["a"].Count, ShouldEqual, 1)
					So(collection.Fields["b"].Count, ShouldEqual, 2)
					So(collection.Fields["c"].Count, ShouldEqual, 1)

					Convey("And the nested document should have 3 fields", func() {
						doc, ok := collection.Fields["b"].Types[0].(*mongo.Document)
						So(ok, ShouldBeTrue)
						So(len(doc.Fields), ShouldEqual, 3)
						So(doc.Fields["c"].Count, ShouldEqual, 2)
						So(doc.Fields["d"].Count, ShouldEqual, 1)
						So(doc.Fields["e"].Count, ShouldEqual, 1)
					})
				})
			})
		})

		Convey("Including a document with a homogenous array", func() {
			err := collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", []interface{}{1, 2, 3}},
			})
			So(err, ShouldBeNil)

			Convey("Should result in 2 fields", func() {
				So(len(collection.Fields), ShouldEqual, 2)
				So(collection.Fields["a"].Count, ShouldEqual, 1)
				So(collection.Fields["b"].Count, ShouldEqual, 1)

				Convey("And the array should have 1 type with the correct count", func() {
					array, ok := collection.Fields["b"].Types[0].(*mongo.Array)
					So(ok, ShouldBeTrue)
					So(array.Count(), ShouldEqual, 1)
					So(len(array.Types), ShouldEqual, 1)

					scalar, ok := array.Types[0].(*mongo.Scalar)
					So(ok, ShouldBeTrue)
					So(scalar.Count(), ShouldEqual, 3)
				})
			})

			Convey("And including an additional document with the same structure", func() {
				err = collection.IncludeSample(bson.D{
					{"a", 1},
					{"b", []interface{}{1, 2, 3}},
				})
				So(err, ShouldBeNil)

				Convey("Should result in 2 fields", func() {
					So(len(collection.Fields), ShouldEqual, 2)
					So(collection.Fields["a"].Count, ShouldEqual, 2)
					So(collection.Fields["b"].Count, ShouldEqual, 2)

					Convey("And the array should have 1 type", func() {
						array, ok := collection.Fields["b"].Types[0].(*mongo.Array)
						So(ok, ShouldBeTrue)
						So(array.Count(), ShouldEqual, 2)
						So(len(array.Types), ShouldEqual, 1)

						scalar, ok := array.Types[0].(*mongo.Scalar)
						So(ok, ShouldBeTrue)
						So(scalar.Count(), ShouldEqual, 6)
					})
				})
			})

			Convey("And including an additional document with a different structure", func() {
				err = collection.IncludeSample(bson.D{
					{"c", 1},
					{"b", []interface{}{"string", "string"}},
				})
				So(err, ShouldBeNil)

				Convey("Should result in 3 fields", func() {
					So(len(collection.Fields), ShouldEqual, 3)
					So(collection.Fields["a"].Count, ShouldEqual, 1)
					So(collection.Fields["b"].Count, ShouldEqual, 2)
					So(collection.Fields["c"].Count, ShouldEqual, 1)

					Convey("And the array should have 1 type", func() {
						array, ok := collection.Fields["b"].Types[0].(*mongo.Array)
						So(ok, ShouldBeTrue)
						So(array.Count(), ShouldEqual, 2)
						So(len(array.Types), ShouldEqual, 2)

						scalar, ok := array.Types[0].(*mongo.Scalar)
						So(ok, ShouldBeTrue)
						So(scalar.Name(), ShouldEqual, "int")
						So(scalar.Count(), ShouldEqual, 3)

						scalar, ok = array.Types[1].(*mongo.Scalar)
						So(ok, ShouldBeTrue)
						So(scalar.Name(), ShouldEqual, "string")
						So(scalar.Count(), ShouldEqual, 2)
					})
				})
			})
		})
	})
}
