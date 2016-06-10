package relational_test

import (
	"encoding/json"
	"fmt"
	"github.com/10gen/sqlproxy/mongodrdl/mongo"
	"github.com/10gen/sqlproxy/mongodrdl/relational"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMapping(t *testing.T) {

	Convey("Given a mongo collection and relational database", t, func() {
		collection := mongo.NewCollection("test")
		database := relational.NewDatabase("test'")
		noIndexes := []mgo.Index{}

		Convey("Without any documents", func() {
			Convey("Mapping the collection", func() {
				err := database.Map(collection, noIndexes)
				So(err, ShouldBeNil)

				Convey("Should result in no tables", func() {
					So(len(database.Tables), ShouldEqual, 0)
				})
			})
		})

		Convey("Without any fields", func() {
			err := collection.IncludeSample(bson.D{})
			So(err, ShouldBeNil)

			Convey("Mapping the collection", func() {
				err := database.Map(collection, noIndexes)
				So(err, ShouldBeNil)

				Convey("Should result in no tables", func() {
					So(len(database.Tables), ShouldEqual, 0)
				})
			})
		})

		Convey("With a flat document", func() {
			err := collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", 1},
			})
			So(err, ShouldBeNil)

			Convey("Mapping the collection", func() {
				err = database.Map(collection, noIndexes)
				So(err, ShouldBeNil)

				Convey("Should result in 1 table", func() {
					So(len(database.Tables), ShouldEqual, 1)
					table := database.Tables[0]
					So(table.Name, ShouldEqual, collection.Name)
					So(len(table.Columns), ShouldEqual, 2)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "b")
				})
			})
		})

		Convey("With a document containing nested documents", func() {
			err := collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", bson.D{
					{"c", 2},
					{"d", bson.D{
						{"e", 3},
						{"f", 4},
					}}}},
				{"g", 5},
			})

			So(err, ShouldBeNil)

			Convey("Mapping the collection", func() {
				err = database.Map(collection, noIndexes)
				So(err, ShouldBeNil)

				Convey("Should result in 1 table", func() {
					So(len(database.Tables), ShouldEqual, 1)
					table := database.Tables[0]
					So(table.Name, ShouldEqual, collection.Name)
					So(len(table.Columns), ShouldEqual, 5)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "b.c")
					So(table.Columns[2].Name, ShouldEqual, "b.d.e")
					So(table.Columns[3].Name, ShouldEqual, "b.d.f")
					So(table.Columns[4].Name, ShouldEqual, "g")
					assertPipeline(table.Pipeline)
				})
			})
		})

		Convey("With a document containing an array of scalars", func() {
			err := collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", []interface{}{1, 2, 3}},
			})

			So(err, ShouldBeNil)

			Convey("Mapping the collection", func() {
				err = database.Map(collection, noIndexes)
				So(err, ShouldBeNil)

				Convey("Should result in 2 tables", func() {
					So(len(database.Tables), ShouldEqual, 2)
					table := database.Tables[0]
					So(table.Name, ShouldEqual, collection.Name)
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "a")
					assertPipeline(table.Pipeline)

					table = database.Tables[1]
					So(table.Name, ShouldEqual, collection.Name+"_b")
					So(len(table.Columns), ShouldEqual, 3)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "b")
					So(table.Columns[2].Name, ShouldEqual, "b_idx")
					assertPipeline(table.Pipeline, unwind{"b", 1})
				})
			})
		})

		Convey("With a document containing an array of documents with different fields", func() {
			err := collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", []interface{}{
					bson.D{{"c", 1}, {"d", 2}},
					bson.D{{"c", 1}, {"e", 2}}}},
			})

			So(err, ShouldBeNil)

			Convey("Mapping the collection", func() {
				err = database.Map(collection, noIndexes)
				So(err, ShouldBeNil)

				Convey("Should result in 2 tables", func() {
					So(len(database.Tables), ShouldEqual, 2)
					table := database.Tables[0]
					So(table.Name, ShouldEqual, collection.Name)
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "a")
					assertPipeline(table.Pipeline)

					table = database.Tables[1]
					So(table.Name, ShouldEqual, collection.Name+"_b")
					So(len(table.Columns), ShouldEqual, 5)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "b.c")
					So(table.Columns[2].Name, ShouldEqual, "b.d")
					So(table.Columns[3].Name, ShouldEqual, "b.e")
					So(table.Columns[4].Name, ShouldEqual, "b_idx")
					So(len(table.Pipeline), ShouldEqual, 1)
					assertPipeline(table.Pipeline, unwind{"b", 1})
				})
			})
		})

		Convey("With a document containing a parent field after an array", func() {
			err := collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", bson.D{
					{"c", []interface{}{bson.D{{"d", 1}}}},
					{"e", 1}}},
				{"f", 1},
			})

			So(err, ShouldBeNil)

			Convey("Mapping the collection", func() {
				err = database.Map(collection, noIndexes)
				So(err, ShouldBeNil)

				Convey("Should result in 2 tables", func() {
					So(len(database.Tables), ShouldEqual, 2)
					table := database.Tables[0]
					So(table.Name, ShouldEqual, collection.Name)
					So(len(table.Columns), ShouldEqual, 3)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "b.e")
					So(table.Columns[2].Name, ShouldEqual, "f")
					assertPipeline(table.Pipeline)

					table = database.Tables[1]
					So(table.Name, ShouldEqual, collection.Name+"_b_c")
					So(len(table.Columns), ShouldEqual, 5)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "b.c.d")
					So(table.Columns[2].Name, ShouldEqual, "b.c_idx")
					So(table.Columns[3].Name, ShouldEqual, "b.e")
					So(table.Columns[4].Name, ShouldEqual, "f")
					assertPipeline(table.Pipeline, unwind{"b.c", 1})
				})
			})
		})

		Convey("With a document containing an array containing an array", func() {
			err := collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", []interface{}{
					[]interface{}{1},
					[]interface{}{2, 3}}},
			})

			So(err, ShouldBeNil)

			Convey("Mapping the collection", func() {
				err = database.Map(collection, noIndexes)
				So(err, ShouldBeNil)

				Convey("Should result in 2 tables", func() {
					So(len(database.Tables), ShouldEqual, 2)
					table := database.Tables[0]
					So(table.Name, ShouldEqual, collection.Name)
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "a")
					assertPipeline(table.Pipeline)

					table = database.Tables[1]
					So(table.Name, ShouldEqual, collection.Name+"_b")
					So(len(table.Columns), ShouldEqual, 4)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "b")
					So(table.Columns[2].Name, ShouldEqual, "b_idx")
					So(table.Columns[3].Name, ShouldEqual, "b_idx_1")
					assertPipeline(table.Pipeline, unwind{"b", 2})
				})
			})
		})

		Convey("With a document containing the same named arrays in different paths", func() {
			err := collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", bson.D{
					{"d", []interface{}{1, 2}}}},
				{"c", bson.D{
					{"d", []interface{}{2, 3}}}},
			})

			So(err, ShouldBeNil)

			Convey("Mapping the collection", func() {
				err = database.Map(collection, noIndexes)
				So(err, ShouldBeNil)

				Convey("Should result in 3 tables", func() {
					So(len(database.Tables), ShouldEqual, 3)
					table := database.Tables[0]
					So(table.Name, ShouldEqual, collection.Name)
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "a")
					assertPipeline(table.Pipeline)

					table = database.Tables[1]
					So(table.Name, ShouldEqual, collection.Name+"_b_d")
					So(len(table.Columns), ShouldEqual, 3)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "b.d")
					So(table.Columns[2].Name, ShouldEqual, "b.d_idx")
					assertPipeline(table.Pipeline, unwind{"b.d", 1})

					table = database.Tables[2]
					So(table.Name, ShouldEqual, collection.Name+"_c_d")
					So(len(table.Columns), ShouldEqual, 3)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "c.d")
					So(table.Columns[2].Name, ShouldEqual, "c.d_idx")
					assertPipeline(table.Pipeline, unwind{"c.d", 1})
				})
			})
		})

		Convey("With a document containing an array of documents", func() {
			err := collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", []interface{}{
					bson.D{{"c", 1}, {"d", 2}},
					bson.D{{"c", 1}, {"e", 3}}}},
			})

			So(err, ShouldBeNil)

			Convey("Mapping the collection", func() {
				err = database.Map(collection, noIndexes)
				So(err, ShouldBeNil)

				Convey("Should result in 2 tables", func() {
					So(len(database.Tables), ShouldEqual, 2)
					table := database.Tables[0]
					So(table.Name, ShouldEqual, collection.Name)
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "a")
					assertPipeline(table.Pipeline)

					table = database.Tables[1]
					So(table.Name, ShouldEqual, collection.Name+"_b")
					So(len(table.Columns), ShouldEqual, 5)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "b.c")
					So(table.Columns[2].Name, ShouldEqual, "b.d")
					So(table.Columns[3].Name, ShouldEqual, "b.e")
					So(table.Columns[4].Name, ShouldEqual, "b_idx")
					assertPipeline(table.Pipeline, unwind{"b", 1})
				})
			})
		})

		Convey("With a document containing an array of documents containing an array of scalars", func() {
			err := collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", []interface{}{
					bson.D{{"c", 1}, {"d", []interface{}{2, 3}}},
					bson.D{{"c", 1}, {"e", 3}}}},
			})

			So(err, ShouldBeNil)

			Convey("Mapping the collection", func() {
				err = database.Map(collection, noIndexes)
				So(err, ShouldBeNil)

				Convey("Should result in 3 tables", func() {
					So(len(database.Tables), ShouldEqual, 3)

					table := database.Tables[0]
					So(table.Name, ShouldEqual, collection.Name)
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "a")
					assertPipeline(table.Pipeline)

					table = database.Tables[1]
					So(table.Name, ShouldEqual, collection.Name+"_b")
					So(len(table.Columns), ShouldEqual, 4)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "b.c")
					So(table.Columns[2].Name, ShouldEqual, "b.e")
					So(table.Columns[3].Name, ShouldEqual, "b_idx")
					assertPipeline(table.Pipeline, unwind{"b", 1})

					table = database.Tables[2]
					So(table.Name, ShouldEqual, collection.Name+"_b_d")
					So(len(table.Columns), ShouldEqual, 6)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "b.c")
					So(table.Columns[2].Name, ShouldEqual, "b.d")
					So(table.Columns[3].Name, ShouldEqual, "b.d_idx")
					So(table.Columns[4].Name, ShouldEqual, "b.e")
					So(table.Columns[5].Name, ShouldEqual, "b_idx")
					assertPipeline(table.Pipeline, unwind{"b", 1}, unwind{"b.d", 1})
				})
			})
		})

		Convey("With a document containing an array of documents containing an array of documents", func() {
			err := collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", []interface{}{
					bson.D{{"c", 1}, {"d", []interface{}{bson.D{{"e", 1}}}}},
					bson.D{{"c", 1}, {"e", 3}}}},
			})

			So(err, ShouldBeNil)

			Convey("Mapping the collection", func() {
				err = database.Map(collection, noIndexes)
				So(err, ShouldBeNil)

				Convey("Should result in 3 tables", func() {
					So(len(database.Tables), ShouldEqual, 3)

					table := database.Tables[0]
					So(table.Name, ShouldEqual, collection.Name)
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "a")
					assertPipeline(table.Pipeline)

					table = database.Tables[1]
					So(table.Name, ShouldEqual, collection.Name+"_b")
					So(len(table.Columns), ShouldEqual, 4)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "b.c")
					So(table.Columns[2].Name, ShouldEqual, "b.e")
					So(table.Columns[3].Name, ShouldEqual, "b_idx")
					assertPipeline(table.Pipeline, unwind{"b", 1})

					table = database.Tables[2]
					So(table.Name, ShouldEqual, collection.Name+"_b_d")
					So(len(table.Columns), ShouldEqual, 6)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "b.c")
					So(table.Columns[2].Name, ShouldEqual, "b.d.e")
					So(table.Columns[3].Name, ShouldEqual, "b.d_idx")
					So(table.Columns[4].Name, ShouldEqual, "b.e")
					So(table.Columns[5].Name, ShouldEqual, "b_idx")
					assertPipeline(table.Pipeline, unwind{"b", 1}, unwind{"b.d", 1})
				})
			})
		})

		Convey("With a document containing a heterogenous array of an scalar array and a scalar", func() {
			err := collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", []interface{}{
					[]interface{}{1, 2, 3},
					4}},
			})

			So(err, ShouldBeNil)

			Convey("Mapping the collection", func() {
				err = database.Map(collection, noIndexes)
				So(err, ShouldBeNil)

				Convey("Should result in 2 tables", func() {
					So(len(database.Tables), ShouldEqual, 2)

					table := database.Tables[0]
					So(table.Name, ShouldEqual, collection.Name)
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "a")
					assertPipeline(table.Pipeline)

					table = database.Tables[1]
					So(table.Name, ShouldEqual, collection.Name+"_b")
					So(len(table.Columns), ShouldEqual, 4)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "b")
					So(table.Columns[2].Name, ShouldEqual, "b_idx")
					So(table.Columns[3].Name, ShouldEqual, "b_idx_1")
					assertPipeline(table.Pipeline, unwind{"b", 2})
				})
			})
		})

		Convey("With a document containing a heterogenous array of a document array and a document", func() {
			err := collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", []interface{}{
					[]interface{}{bson.D{{"c", 2}, {"d", 3}}, bson.D{{"e", 4}}},
					bson.D{{"c", 5}, {"f", 6}},
					bson.D{{"c", 7}, {"f", 8}}}},
			})

			So(err, ShouldBeNil)

			Convey("Mapping the collection", func() {
				err = database.Map(collection, noIndexes)
				So(err, ShouldBeNil)

				Convey("Should result in 2 tables", func() {
					So(len(database.Tables), ShouldEqual, 2)

					table := database.Tables[0]
					So(table.Name, ShouldEqual, collection.Name)
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "a")
					assertPipeline(table.Pipeline)

					table = database.Tables[1]
					So(table.Name, ShouldEqual, collection.Name+"_b")
					So(len(table.Columns), ShouldEqual, 7)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[1].Name, ShouldEqual, "b.c")
					So(table.Columns[2].Name, ShouldEqual, "b.d")
					So(table.Columns[3].Name, ShouldEqual, "b.e")
					So(table.Columns[4].Name, ShouldEqual, "b.f")
					So(table.Columns[5].Name, ShouldEqual, "b_idx")
					So(table.Columns[6].Name, ShouldEqual, "b_idx_1")
					assertPipeline(table.Pipeline, unwind{"b", 2})
				})
			})
		})

		Convey("Given a collection that already exists with computed array table name", func() {
			// this name is the same as the array table name in
			// the next collection
			otherCollection := mongo.NewCollection("test_b")
			err := otherCollection.IncludeSample(bson.D{
				{"a", 1},
			})
			err = database.Map(otherCollection, noIndexes)

			err = collection.IncludeSample(bson.D{
				{"a", 1},
				{"b", []interface{}{1, 2, 3}},
			})
			So(err, ShouldBeNil)

			Convey("Mapping should result in an error", func() {
				err := database.Map(collection, noIndexes)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestTypeMapping(t *testing.T) {

	Convey("Given an empty collection.", t, func() {
		collection := mongo.NewCollection("test")
		noIndexes := []mgo.Index{}

		var num1 int32 = 42
		var num2 int64 = 100000000
		var num3 float32 = 1.0
		var num4 float64 = 1000.34001

		typeTests := [][]interface{}{
			{"string", "varchar", "Hello, world"},
			{"int32", "numeric", num1},
			{"int64", "numeric", num2},
			{"float32", "numeric", num3},
			{"float64", "numeric", num4},
			{"bool", "boolean", true},
			{"date", "timestamp", time.Date(2015, 1, 1, 1, 1, 1, 1, time.UTC)},
			{"[]uint8", "varchar", []byte{1, 2, 3, 4, 5}},
		}

		for _, typeTest := range typeTests {
			Convey(fmt.Sprintf("Should map %s to %s", typeTest[0].(string), typeTest[1].(string)), func() {
				collection.IncludeSample(bson.D{{typeTest[0].(string), typeTest[2]}})

				database := relational.NewDatabase("test")
				database.Map(collection, noIndexes)

				table := database.Tables[0]
				for _, c := range table.Columns {
					So(c.MongoType, ShouldEqual, typeTest[0].(string))
					So(c.SqlType, ShouldEqual, typeTest[1].(string))
				}
			})
		}

		Convey("Should use the majority type for a field", func() {
			Convey("When the majority type is a scalar and the minority is a scalar", func() {
				collection.IncludeSample(bson.D{{"a", 1}})
				collection.IncludeSample(bson.D{{"a", 2}})
				collection.IncludeSample(bson.D{{"a", 3}})
				collection.IncludeSample(bson.D{{"a", "string"}})
				collection.IncludeSample(bson.D{{"a", 4}})
				collection.IncludeSample(bson.D{{"a", 2.562}})

				database := relational.NewDatabase("test")
				database.Map(collection, noIndexes)

				table := database.Tables[0]
				So(len(table.Columns), ShouldEqual, 1)
				So(table.Columns[0].Name, ShouldEqual, "a")
				So(table.Columns[0].MongoType, ShouldEqual, "int")
				So(table.Columns[0].SqlType, ShouldEqual, "numeric")
			})

			Convey("When the majority is an array and the minority is incompatible", func() {
				collection.IncludeSample(bson.D{{"a", []interface{}{1, 2, 3}}})
				collection.IncludeSample(bson.D{{"a", []interface{}{1, 2, 3}}})
				collection.IncludeSample(bson.D{{"a", []interface{}{1, 2, 3}}})
				collection.IncludeSample(bson.D{{"a", "funny"}})

				database := relational.NewDatabase("test")
				database.Map(collection, noIndexes)

				table := database.Tables[0]
				So(len(table.Columns), ShouldEqual, 2)
				So(table.Columns[0].Name, ShouldEqual, "a")
				So(table.Columns[0].MongoType, ShouldEqual, "int")
				So(table.Columns[0].SqlType, ShouldEqual, "numeric")
				So(table.Columns[1].Name, ShouldEqual, "a_idx")
				So(table.Columns[1].MongoType, ShouldEqual, "int")
				So(table.Columns[1].SqlType, ShouldEqual, "numeric")
			})

			Convey("When the majority is a scalar and the minority is an array whose item type is incompatible", func() {
				collection.IncludeSample(bson.D{{"a", 1}})
				collection.IncludeSample(bson.D{{"a", 2}})
				collection.IncludeSample(bson.D{{"a", 3}})
				collection.IncludeSample(bson.D{{"a", []interface{}{"string1", "string2", "string3"}}})

				database := relational.NewDatabase("test")
				database.Map(collection, noIndexes)

				table := database.Tables[0]
				So(len(table.Columns), ShouldEqual, 1)
				So(table.Columns[0].Name, ShouldEqual, "a")
				So(table.Columns[0].MongoType, ShouldEqual, "int")
				So(table.Columns[0].SqlType, ShouldEqual, "numeric")
			})

			Convey("When the majority is a scalar and the minority is an array whose item type is compatible", func() {
				collection.IncludeSample(bson.D{{"a", 1}})
				collection.IncludeSample(bson.D{{"a", 2}})
				collection.IncludeSample(bson.D{{"a", 3}})
				collection.IncludeSample(bson.D{{"a", []interface{}{1, 2, 3}}})

				database := relational.NewDatabase("test")
				database.Map(collection, noIndexes)

				table := database.Tables[0]
				So(len(table.Columns), ShouldEqual, 2)
				So(table.Columns[0].Name, ShouldEqual, "a")
				So(table.Columns[0].MongoType, ShouldEqual, "int")
				So(table.Columns[0].SqlType, ShouldEqual, "numeric")
				So(table.Columns[1].Name, ShouldEqual, "a_idx")
				So(table.Columns[1].MongoType, ShouldEqual, "int")
				So(table.Columns[1].SqlType, ShouldEqual, "numeric")
			})
		})

		Convey("Should use the majority type for an array", func() {
			Convey("When the majority type is a scalar and the minority is a scalar", func() {
				collection.IncludeSample(bson.D{{"a", []interface{}{1, 2, 3, "string1"}}})
				collection.IncludeSample(bson.D{{"a", []interface{}{"string2", "string3", 4}}})

				database := relational.NewDatabase("test")
				database.Map(collection, noIndexes)

				table := database.Tables[0]
				So(len(table.Columns), ShouldEqual, 2)
				So(table.Columns[0].Name, ShouldEqual, "a")
				So(table.Columns[0].MongoType, ShouldEqual, "int")
				So(table.Columns[0].SqlType, ShouldEqual, "numeric")
				So(table.Columns[1].Name, ShouldEqual, "a_idx")
				So(table.Columns[1].MongoType, ShouldEqual, "int")
				So(table.Columns[1].SqlType, ShouldEqual, "numeric")
			})

			Convey("When the majority is an array and the minority is incompatible", func() {
				collection.IncludeSample(bson.D{{"a", []interface{}{[]interface{}{1, 1}, []interface{}{2, 2}, "string1"}}})
				collection.IncludeSample(bson.D{{"a", []interface{}{"string2", []interface{}{3, 3}}}})

				database := relational.NewDatabase("test")
				database.Map(collection, noIndexes)

				table := database.Tables[0]
				So(len(table.Columns), ShouldEqual, 3)
				So(table.Columns[0].Name, ShouldEqual, "a")
				So(table.Columns[0].MongoType, ShouldEqual, "int")
				So(table.Columns[0].SqlType, ShouldEqual, "numeric")
				So(table.Columns[1].Name, ShouldEqual, "a_idx")
				So(table.Columns[1].MongoType, ShouldEqual, "int")
				So(table.Columns[1].SqlType, ShouldEqual, "numeric")
				So(table.Columns[2].Name, ShouldEqual, "a_idx_1")
				So(table.Columns[2].MongoType, ShouldEqual, "int")
				So(table.Columns[2].SqlType, ShouldEqual, "numeric")
			})

			Convey("When the majority is a scalar and the minority is an array whose item type is incompatible", func() {
				collection.IncludeSample(bson.D{{"a", []interface{}{[]interface{}{1, 1}, []interface{}{2, 2}, "string1"}}})
				collection.IncludeSample(bson.D{{"a", []interface{}{"string2", []interface{}{3, 3}, "string3", "string4"}}})

				database := relational.NewDatabase("test")
				database.Map(collection, noIndexes)

				table := database.Tables[0]
				So(len(table.Columns), ShouldEqual, 2)
				So(table.Columns[0].Name, ShouldEqual, "a")
				So(table.Columns[0].MongoType, ShouldEqual, "string")
				So(table.Columns[0].SqlType, ShouldEqual, "varchar")
				So(table.Columns[1].Name, ShouldEqual, "a_idx")
				So(table.Columns[1].MongoType, ShouldEqual, "int")
				So(table.Columns[1].SqlType, ShouldEqual, "numeric")
			})

			Convey("When the majority is a scalar and the minority is an array whose item type is compatible", func() {
				collection.IncludeSample(bson.D{{"a", []interface{}{[]interface{}{1, 1}, []interface{}{2, 2}, 1}}})
				collection.IncludeSample(bson.D{{"a", []interface{}{2, []interface{}{3, 3}, 3, 4}}})

				database := relational.NewDatabase("test")
				database.Map(collection, noIndexes)

				table := database.Tables[0]
				So(len(table.Columns), ShouldEqual, 3)
				So(table.Columns[0].Name, ShouldEqual, "a")
				So(table.Columns[0].MongoType, ShouldEqual, "int")
				So(table.Columns[0].SqlType, ShouldEqual, "numeric")
				So(table.Columns[1].Name, ShouldEqual, "a_idx")
				So(table.Columns[1].MongoType, ShouldEqual, "int")
				So(table.Columns[1].SqlType, ShouldEqual, "numeric")
				So(table.Columns[2].Name, ShouldEqual, "a_idx_1")
				So(table.Columns[2].MongoType, ShouldEqual, "int")
				So(table.Columns[2].SqlType, ShouldEqual, "numeric")
			})
		})

		Convey("Should use the minority type for a field", func() {
			Convey("When the minority is an array whose item type is compatible with the majority", func() {
				collection.IncludeSample(bson.D{{"a", []interface{}{1, 2, 3}}})
				collection.IncludeSample(bson.D{{"a", 4}})
				collection.IncludeSample(bson.D{{"a", 5}})
				collection.IncludeSample(bson.D{{"a", 6}})

				database := relational.NewDatabase("test")
				database.Map(collection, noIndexes)

				table := database.Tables[0]
				So(len(table.Columns), ShouldEqual, 2)
				So(table.Columns[0].Name, ShouldEqual, "a")
				So(table.Columns[0].MongoType, ShouldEqual, "int")
				So(table.Columns[0].SqlType, ShouldEqual, "numeric")
				So(table.Columns[1].Name, ShouldEqual, "a_idx")
				So(table.Columns[1].MongoType, ShouldEqual, "int")
				So(table.Columns[1].SqlType, ShouldEqual, "numeric")
			})
		})

		Convey("Should use the minority type for an array", func() {
			Convey("When the minority is an array whose item type is compatible with the majority", func() {
				collection.IncludeSample(bson.D{{"a", []interface{}{[]interface{}{1, 1}, []interface{}{2, 2}, 1}}})
				collection.IncludeSample(bson.D{{"a", []interface{}{2, []interface{}{3, 3}, 3, 4}}})

				database := relational.NewDatabase("test")
				database.Map(collection, noIndexes)

				table := database.Tables[0]
				So(len(table.Columns), ShouldEqual, 3)
				So(table.Columns[0].Name, ShouldEqual, "a")
				So(table.Columns[0].MongoType, ShouldEqual, "int")
				So(table.Columns[0].SqlType, ShouldEqual, "numeric")
				So(table.Columns[1].Name, ShouldEqual, "a_idx")
				So(table.Columns[1].MongoType, ShouldEqual, "int")
				So(table.Columns[1].SqlType, ShouldEqual, "numeric")
				So(table.Columns[2].Name, ShouldEqual, "a_idx_1")
				So(table.Columns[2].MongoType, ShouldEqual, "int")
				So(table.Columns[2].SqlType, ShouldEqual, "numeric")
			})
		})

		Convey("Should use the first type for a field when there is no majority of scalar types", func() {
			collection.IncludeSample(bson.D{{"a", 1}})
			collection.IncludeSample(bson.D{{"a", "string1"}})
			collection.IncludeSample(bson.D{{"a", 2}})
			collection.IncludeSample(bson.D{{"a", "string2"}})
			collection.IncludeSample(bson.D{{"a", "string3"}})
			collection.IncludeSample(bson.D{{"a", 3}})

			database := relational.NewDatabase("test")
			err := database.Map(collection, noIndexes)
			So(err, ShouldBeNil)

			table := database.Tables[0]
			So(len(table.Columns), ShouldEqual, 1)
			So(table.Columns[0].Name, ShouldEqual, "a")
			So(table.Columns[0].MongoType, ShouldEqual, "int") // we use a stable sort and int was first
			So(table.Columns[0].SqlType, ShouldEqual, "numeric")
		})

		Convey("Should use the first type for an array when there is a no majority of scalar types", func() {
			collection.IncludeSample(bson.D{{"a", []interface{}{1, 2, 3, "string1", "string2"}}})
			collection.IncludeSample(bson.D{{"a", []interface{}{"string3"}}})

			database := relational.NewDatabase("test")
			err := database.Map(collection, noIndexes)
			So(err, ShouldBeNil)

			table := database.Tables[0]
			So(len(table.Columns), ShouldEqual, 2)
			So(table.Columns[0].Name, ShouldEqual, "a")
			So(table.Columns[0].MongoType, ShouldEqual, "int") // we use a stable sort and int was first
			So(table.Columns[0].SqlType, ShouldEqual, "numeric")
			So(table.Columns[1].Name, ShouldEqual, "a_idx")
			So(table.Columns[1].MongoType, ShouldEqual, "int")
			So(table.Columns[1].SqlType, ShouldEqual, "numeric")
		})

		Convey("Should not include a field when it has no types", func() {
			collection.IncludeSample(bson.D{{"a", 1}, {"b", nil}})

			database := relational.NewDatabase("test")
			err := database.Map(collection, noIndexes)
			So(err, ShouldBeNil)

			table := database.Tables[0]
			So(len(table.Columns), ShouldEqual, 1)
			So(table.Columns[0].Name, ShouldEqual, "a")
			So(table.Columns[0].MongoType, ShouldEqual, "int")
			So(table.Columns[0].SqlType, ShouldEqual, "numeric")
		})

		Convey("Should not include an array when it has no types", func() {
			collection.IncludeSample(bson.D{{"a", 1}, {"b", []interface{}{nil, nil}}})

			database := relational.NewDatabase("test")
			err := database.Map(collection, noIndexes)
			So(err, ShouldBeNil)

			table := database.Tables[0]
			So(len(table.Columns), ShouldEqual, 1)
			So(table.Columns[0].Name, ShouldEqual, "a")
			So(table.Columns[0].MongoType, ShouldEqual, "int")
			So(table.Columns[0].SqlType, ShouldEqual, "numeric")
		})

		Convey("When indexes are present", func() {
			Convey("Subject: 2d index", func() {

				geoIndexes := []mgo.Index{
					mgo.Index{Key: []string{"$2d:a"}},
					mgo.Index{Key: []string{"$2d:b.c"}},
				}

				Convey("Should not map without any type samples", func() {
					collection.IncludeSample(bson.D{{"a", nil}})

					database := relational.NewDatabase("test")
					err := database.Map(collection, geoIndexes)
					So(err, ShouldBeNil)

					// no columns will remove the entire table
					So(len(database.Tables), ShouldEqual, 0)
				})

				Convey("Should map with an array sample", func() {
					collection.IncludeSample(bson.D{{"a", []interface{}{1, 2}}})

					database := relational.NewDatabase("test")
					err := database.Map(collection, geoIndexes)
					So(err, ShouldBeNil)

					table := database.Tables[0]
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[0].MongoType, ShouldEqual, "geo.2darray")
					So(table.Columns[0].SqlType, ShouldEqual, "numeric[]")
				})

				Convey("Should map in a nested document", func() {
					collection.IncludeSample(bson.D{{"b", bson.D{{"c", []interface{}{1, 2}}}}})

					database := relational.NewDatabase("test")
					err := database.Map(collection, geoIndexes)
					So(err, ShouldBeNil)

					table := database.Tables[0]
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "b.c")
					So(table.Columns[0].MongoType, ShouldEqual, "geo.2darray")
					So(table.Columns[0].SqlType, ShouldEqual, "numeric[]")
				})

				Convey("Should map in a document in an array", func() {
					collection.IncludeSample(bson.D{{"b", []interface{}{bson.D{{"c", []interface{}{1, 2}}}}}})

					database := relational.NewDatabase("test")
					err := database.Map(collection, geoIndexes)
					So(err, ShouldBeNil)

					table := database.Tables[0]
					So(len(table.Columns), ShouldEqual, 2)
					So(table.Columns[0].Name, ShouldEqual, "b.c")
					So(table.Columns[0].MongoType, ShouldEqual, "geo.2darray")
					So(table.Columns[0].SqlType, ShouldEqual, "numeric[]")
					So(table.Columns[1].Name, ShouldEqual, "b_idx")
					So(table.Columns[1].MongoType, ShouldEqual, "int")
					So(table.Columns[1].SqlType, ShouldEqual, "numeric")
				})

				Convey("Should map with a document sample", func() {
					collection.IncludeSample(bson.D{{"a", bson.D{{"x", 1}, {"y", 2}}}})

					database := relational.NewDatabase("test")
					err := database.Map(collection, geoIndexes)
					So(err, ShouldBeNil)

					table := database.Tables[0]
					So(len(table.Columns), ShouldEqual, 2)
					So(table.Columns[0].Name, ShouldEqual, "a.x")
					So(table.Columns[0].MongoType, ShouldEqual, "int")
					So(table.Columns[0].SqlType, ShouldEqual, "numeric")
					So(table.Columns[1].Name, ShouldEqual, "a.y")
					So(table.Columns[1].MongoType, ShouldEqual, "int")
					So(table.Columns[1].SqlType, ShouldEqual, "numeric")
				})

				Convey("Should fallback to majority type with a non-document/non-array sample", func() {
					collection.IncludeSample(bson.D{{"a", 10}})

					database := relational.NewDatabase("test")
					err := database.Map(collection, geoIndexes)
					So(err, ShouldBeNil)

					table := database.Tables[0]
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[0].MongoType, ShouldEqual, "int")
					So(table.Columns[0].SqlType, ShouldEqual, "numeric")
				})
			})

			Convey("Subject: 2d sphere", func() {

				geoIndexes := []mgo.Index{
					mgo.Index{Key: []string{"$2dsphere:a"}},
					mgo.Index{Key: []string{"$2dsphere:b.c"}},
				}

				Convey("Should not map without any type samples", func() {
					collection.IncludeSample(bson.D{{"a", nil}})

					database := relational.NewDatabase("test")
					err := database.Map(collection, geoIndexes)
					So(err, ShouldBeNil)

					// no columns will remove the entire table
					So(len(database.Tables), ShouldEqual, 0)
				})

				Convey("Should map the array form", func() {
					collection.IncludeSample(bson.D{{"a", []interface{}{1, 2}}})
					collection.IncludeSample(bson.D{{"a", []interface{}{1, 2}}})
					collection.IncludeSample(bson.D{{"a", bson.D{{"coordinates", []interface{}{1, 2}}}}})

					database := relational.NewDatabase("test")
					err := database.Map(collection, geoIndexes)
					So(err, ShouldBeNil)

					table := database.Tables[0]
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[0].MongoType, ShouldEqual, "geo.2darray")
					So(table.Columns[0].SqlType, ShouldEqual, "numeric[]")
				})

				Convey("Should map the geoJson form", func() {
					collection.IncludeSample(bson.D{{"a", bson.D{{"coordinates", []interface{}{1, 2}}}}})
					collection.IncludeSample(bson.D{{"a", bson.D{{"coordinates", []interface{}{1, 2}}}}})
					collection.IncludeSample(bson.D{{"a", []interface{}{1, 2}}})

					database := relational.NewDatabase("test")
					err := database.Map(collection, geoIndexes)
					So(err, ShouldBeNil)

					table := database.Tables[0]
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "a.coordinates")
					So(table.Columns[0].MongoType, ShouldEqual, "geo.2darray")
					So(table.Columns[0].SqlType, ShouldEqual, "numeric[]")
				})

				Convey("Should map the geoJson form even with a LineString", func() {
					// technically, this isn't going to be correct for the user. However,
					// handling this possibility is very difficult and as such, we are
					// going to assume Point uniformly.
					collection.IncludeSample(bson.D{{"a", bson.D{{"type", "LineString"}, {"coordinates", []interface{}{[]interface{}{1, 2}, []interface{}{3, 4}}}}}})

					database := relational.NewDatabase("test")
					err := database.Map(collection, geoIndexes)
					So(err, ShouldBeNil)

					table := database.Tables[0]
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "a.coordinates")
					So(table.Columns[0].MongoType, ShouldEqual, "geo.2darray")
					So(table.Columns[0].SqlType, ShouldEqual, "numeric[]")
				})

				Convey("Should fallback to majority type when there isn't an array or geoJson form majority", func() {
					collection.IncludeSample(bson.D{{"a", 10}})

					database := relational.NewDatabase("test")
					err := database.Map(collection, geoIndexes)
					So(err, ShouldBeNil)

					table := database.Tables[0]
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "a")
					So(table.Columns[0].MongoType, ShouldEqual, "int")
					So(table.Columns[0].SqlType, ShouldEqual, "numeric")
				})

				Convey("Should map the array form in a nested document", func() {
					collection.IncludeSample(bson.D{{"b", bson.D{{"c", []interface{}{1, 2}}}}})

					database := relational.NewDatabase("test")
					err := database.Map(collection, geoIndexes)
					So(err, ShouldBeNil)

					table := database.Tables[0]
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "b.c")
					So(table.Columns[0].MongoType, ShouldEqual, "geo.2darray")
					So(table.Columns[0].SqlType, ShouldEqual, "numeric[]")
				})

				Convey("Should map the geoJson form in a nested document", func() {
					collection.IncludeSample(bson.D{{"b", bson.D{{"c", bson.D{{"coordinates", []interface{}{1, 2}}}}}}})

					database := relational.NewDatabase("test")
					err := database.Map(collection, geoIndexes)
					So(err, ShouldBeNil)

					table := database.Tables[0]
					So(len(table.Columns), ShouldEqual, 1)
					So(table.Columns[0].Name, ShouldEqual, "b.c.coordinates")
					So(table.Columns[0].MongoType, ShouldEqual, "geo.2darray")
					So(table.Columns[0].SqlType, ShouldEqual, "numeric[]")
				})

				Convey("Should map the array form in an array", func() {
					collection.IncludeSample(bson.D{{"b", []interface{}{bson.D{{"c", []interface{}{1, 2}}}}}})

					database := relational.NewDatabase("test")
					err := database.Map(collection, geoIndexes)
					So(err, ShouldBeNil)

					table := database.Tables[0]
					So(len(table.Columns), ShouldEqual, 2)
					So(table.Columns[0].Name, ShouldEqual, "b.c")
					So(table.Columns[0].MongoType, ShouldEqual, "geo.2darray")
					So(table.Columns[0].SqlType, ShouldEqual, "numeric[]")
					So(table.Columns[1].Name, ShouldEqual, "b_idx")
					So(table.Columns[1].MongoType, ShouldEqual, "int")
					So(table.Columns[1].SqlType, ShouldEqual, "numeric")
				})

				Convey("Should map the geoJson form in an array", func() {
					collection.IncludeSample(bson.D{{"b", []interface{}{bson.D{{"c", bson.D{{"coordinates", []interface{}{1, 2}}}}}}}})

					database := relational.NewDatabase("test")
					err := database.Map(collection, geoIndexes)
					So(err, ShouldBeNil)

					table := database.Tables[0]
					So(len(table.Columns), ShouldEqual, 2)
					So(table.Columns[0].Name, ShouldEqual, "b.c.coordinates")
					So(table.Columns[0].MongoType, ShouldEqual, "geo.2darray")
					So(table.Columns[0].SqlType, ShouldEqual, "numeric[]")
					So(table.Columns[1].Name, ShouldEqual, "b_idx")
					So(table.Columns[1].MongoType, ShouldEqual, "int")
					So(table.Columns[1].SqlType, ShouldEqual, "numeric")
				})
			})
		})
	})
}

type unwind struct {
	field string
	count int
}

func assertPipeline(pipeline []map[string]interface{}, unwinds ...unwind) {

	bytes, _ := json.Marshal(pipeline)
	actual := string(bytes)
	stages := []string{}
	for _, unwind := range unwinds {
		for i := 0; i < unwind.count; i++ {
			indexName := unwind.field + "_idx"
			if i > 0 {
				indexName += "_" + strconv.Itoa(i)
			}
			stages = append(stages, fmt.Sprintf(`{"$unwind":{"includeArrayIndex":"%s","path":"$%s"}}`, indexName, unwind.field))
		}
	}
	expected := "[" + strings.Join(stages, ",") + "]"

	So(len(pipeline), ShouldEqual, len(stages))
	if len(stages) == 0 {
		return
	}
	So(actual, ShouldEqual, expected)
}
