package evaluator

import (
	"testing"
	"time"

	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

var (
	testCtx = &EvalCtx{
		Rows: Rows{
			Row{Values{{tableOneName, "a", 1}, {tableOneName, "b", 1}}},
			Row{Values{{tableOneName, "a", 2}, {tableOneName, "b", 2}}},
			Row{Values{{tableOneName, "a", 3}, {tableOneName, "b", 3}}},
			Row{Values{{tableOneName, "a", 4}, {tableOneName, "b", 4}}},
		},
	}
)

func TestNewSQLValue(t *testing.T) {

	Convey("When creating a SQLValue with no column type specified calling NewSQLValue on a", t, func() {

		Convey("SQLValue should return the same object passed in", func() {
			v := SQLTrue
			newV, err := NewSQLValue(v, schema.SQLBoolean, schema.MongoBool)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, v)
		})

		Convey("nil value should return SQLNull", func() {
			v, err := NewSQLValue(nil, schema.SQLNull, schema.MongoBool)
			So(err, ShouldBeNil)
			So(v, ShouldResemble, SQLNull)
		})

		Convey("bson object id should return its string value", func() {
			v := bson.ObjectId("56a10dd56ce28a89a8ed6edb")
			newV, err := NewSQLValue(v, schema.SQLVarchar, schema.MongoObjectId)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v.Hex())
		})

		Convey("string objects should return the string value", func() {
			v := "56a10dd56ce28a89a8ed6edb"
			newV, err := NewSQLValue(v, schema.SQLVarchar, schema.MongoString)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v)
		})

		Convey("int objects should return the int value", func() {
			v1 := int(6)
			newV, err := NewSQLValue(v1, schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v1)

			v2 := int32(6)
			newV, err = NewSQLValue(v2, schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v2)

			v3 := uint32(6)
			newV, err = NewSQLValue(v3, schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v3)
		})

		Convey("float objects should return the float value", func() {
			v := float64(6.3)
			newV, err := NewSQLValue(v, schema.SQLFloat, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v)
		})

		Convey("time objects should return the appropriate value", func() {
			v := time.Date(2014, time.December, 31, 0, 0, 0, 0, schema.DefaultLocale)
			newV, err := NewSQLValue(v, schema.SQLDate, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlDate, ok := newV.(SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, SQLDate{v})

			v = time.Date(2014, time.December, 31, 10, 0, 0, 0, schema.DefaultLocale)
			newV, err = NewSQLValue(v, schema.SQLTimestamp, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlTimestamp, ok := newV.(SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTimestamp, ShouldResemble, SQLTimestamp{v})
		})
	})

	Convey("When creating a SQLValue with a column type specified calling NewSQLValue on a", t, func() {

		Convey("a SQLVarchar/SQLVarchar column type should attempt to coerce to the SQLVarchar type", func() {

			t := schema.SQLVarchar

			newV, err := NewSQLValue(t, schema.SQLVarchar, schema.MongoString)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLVarchar(t))

			newV, err = NewSQLValue(6, schema.SQLVarchar, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLVarchar("6"))

			newV, err = NewSQLValue(6.6, schema.SQLVarchar, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLVarchar("6.6"))

			newV, err = NewSQLValue(int64(6), schema.SQLVarchar, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLVarchar("6"))

			_id := bson.ObjectId("56a10dd56ce28a89a8ed6edb")
			newV, err = NewSQLValue(_id, schema.SQLVarchar, schema.MongoObjectId)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLObjectID(_id.Hex()))

		})

		Convey("a SQLInt column type should attempt to coerce to the SQLInt type", func() {

			_, err := NewSQLValue(true, schema.SQLInt, schema.MongoBool)
			So(err, ShouldNotBeNil)

			_, err = NewSQLValue("6", schema.SQLInt, schema.MongoString)
			So(err, ShouldNotBeNil)

			newV, err := NewSQLValue(int(6), schema.SQLInt, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

			newV, err = NewSQLValue(int32(6), schema.SQLInt, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

			newV, err = NewSQLValue(int64(6), schema.SQLInt, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

			newV, err = NewSQLValue(float64(6.6), schema.SQLInt, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLInt(6))

		})

		Convey("a SQLFloat column type should attempt to coerce to the SQLFloat type", func() {

			_, err := NewSQLValue(true, schema.SQLFloat, schema.MongoBool)
			So(err, ShouldNotBeNil)

			_, err = NewSQLValue("6.6", schema.SQLFloat, schema.MongoString)
			So(err, ShouldNotBeNil)

			newV, err := NewSQLValue(int(6), schema.SQLFloat, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6))

			newV, err = NewSQLValue(int32(6), schema.SQLFloat, schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6))

			newV, err = NewSQLValue(int64(6), schema.SQLFloat, schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6))

			newV, err = NewSQLValue(float64(6.6), schema.SQLFloat, schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, SQLFloat(6.6))

		})

		Convey("a SQLDate column type should attempt to coerce to the SQLDate type", func() {

			// Time type
			v1 := time.Date(2014, time.May, 11, 0, 0, 0, 0, schema.DefaultLocale)
			v2 := time.Date(2014, time.May, 11, 10, 32, 12, 0, schema.DefaultLocale)

			newV, err := NewSQLValue(v1, schema.SQLDate, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlDate, ok := newV.(SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, SQLDate{v1})

			newV, err = NewSQLValue(v2, schema.SQLDate, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlDate, ok = newV.(SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, SQLDate{v1})

			// String type
			dates := []string{"2014-05-11", "2014-05-11 15:04:05", "2014-05-11 15:04:05.233"}

			for _, d := range dates {

				newV, err := NewSQLValue(d, schema.SQLDate, schema.MongoNone)
				So(err, ShouldBeNil)

				sqlDate, ok := newV.(SQLDate)
				So(ok, ShouldBeTrue)
				So(sqlDate, ShouldResemble, SQLDate{v1})

			}

			// invalid dates and those outside valid range
			// should return the default date
			dates = []string{"2014-12-44-44", "999-1-1", "10000-1-1"}

			for _, d := range dates {
				_, err = NewSQLValue(d, schema.SQLDate, schema.MongoNone)
				So(err, ShouldNotBeNil)
			}
		})

		Convey("a SQLTimestamp column type should attempt to coerce to the SQLTimestamp type", func() {

			// Time type
			v1 := time.Date(2014, time.May, 11, 15, 4, 5, 0, schema.DefaultLocale)

			newV, err := NewSQLValue(v1, schema.SQLTimestamp, schema.MongoNone)
			So(err, ShouldBeNil)

			sqlTs, ok := newV.(SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTs, ShouldResemble, SQLTimestamp{v1})

			// String type
			newV, err = NewSQLValue("2014-05-11 15:04:05.000", schema.SQLTimestamp, schema.MongoNone)
			So(err, ShouldBeNil)

			sqlTs, ok = newV.(SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTs, ShouldResemble, SQLTimestamp{v1})

			// invalid dates should return the default date
			dates := []string{"2044-12-40", "1966-15-1", "43223-3223"}

			for _, d := range dates {
				_, err = NewSQLValue(d, schema.SQLTimestamp, schema.MongoNone)
				So(err, ShouldNotBeNil)
			}
		})
	})
}
