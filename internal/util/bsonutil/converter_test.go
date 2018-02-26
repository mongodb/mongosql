package bsonutil

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/json"
	. "github.com/smartystreets/goconvey/convey"
)

func TestObjectIdBSONToJSON(t *testing.T) {

	Convey("Converting a BSON ObjectId", t, func() {
		Convey("that is valid to JSON should produce a json.ObjectId", func() {
			bsonObjID := bson.NewObjectId()
			jsonObjID := json.ObjectID(bsonObjID.Hex())

			_jObjID, err := ConvertBSONValueToJSON(bsonObjID)
			So(err, ShouldBeNil)
			jObjID, ok := _jObjID.(json.ObjectID)
			So(ok, ShouldBeTrue)

			So(jObjID, ShouldNotEqual, bsonObjID)
			So(jObjID, ShouldEqual, jsonObjID)
		})
	})
}

func TestArraysBSONToJSON(t *testing.T) {
	Convey("Converting BSON arrays to JSON arrays", t, func() {
		Convey("should work for empty arrays", func() {
			jArr, err := ConvertBSONValueToJSON([]interface{}{})
			So(err, ShouldBeNil)

			So(jArr, ShouldResemble, []interface{}{})
		})

		Convey("should work for one-level deep arrays", func() {
			objID := bson.NewObjectId()
			bsonArr := []interface{}{objID, 28, 0.999, "plain"}
			_jArr, err := ConvertBSONValueToJSON(bsonArr)
			So(err, ShouldBeNil)
			jArr, ok := _jArr.([]interface{})
			So(ok, ShouldBeTrue)

			So(len(jArr), ShouldEqual, 4)
			So(jArr[0], ShouldEqual, json.ObjectID(objID.Hex()))
			So(jArr[1], ShouldEqual, 28)
			So(jArr[2], ShouldEqual, 0.999)
			So(jArr[3], ShouldEqual, "plain")
		})

		Convey("should work for arrays with embedded objects", func() {
			bsonObj := []interface{}{
				80,
				bson.M{
					"a": int64(20),
					"b": bson.M{
						"c": bson.RegEx{Pattern: "hi", Options: "i"},
					},
				},
			}

			_XjObj, err := ConvertBSONValueToJSON(bsonObj)
			So(err, ShouldBeNil)
			_jObj, ok := _XjObj.([]interface{})
			So(ok, ShouldBeTrue)
			jObj, ok := _jObj[1].(bson.M)
			So(ok, ShouldBeTrue)
			So(len(jObj), ShouldEqual, 2)
			So(jObj["a"], ShouldEqual, json.NumberLong(20))
			jjObj, ok := jObj["b"].(bson.M)
			So(ok, ShouldBeTrue)

			So(jjObj["c"], ShouldResemble, json.RegExp{Pattern: "hi", Options: "i"})
			So(jjObj["c"], ShouldNotResemble, json.RegExp{Pattern: "i", Options: "hi"})
		})

	})
}

func TestDateBSONToJSON(t *testing.T) {
	timeNow := time.Now()
	secs := int64(timeNow.Unix())
	nanosecs := timeNow.Nanosecond()
	millis := int64(nanosecs / 1e6)

	timeNowSecs := time.Unix(secs, int64(0))
	timeNowMillis := time.Unix(secs, int64(millis*1e6))

	Convey("Converting BSON time.Time 's dates to JSON", t, func() {
		// json.Date is stored as an int64 representing the number of milliseconds since the epoch
		Convey(fmt.Sprintf("should work with second granularity: %v", timeNowSecs), func() {
			_jObj, err := ConvertBSONValueToJSON(timeNowSecs)
			So(err, ShouldBeNil)
			jObj, ok := _jObj.(json.Date)
			So(ok, ShouldBeTrue)

			So(int64(jObj), ShouldEqual, secs*1e3)
		})

		Convey(fmt.Sprintf("should work with millisecond granularity: %v", timeNowMillis), func() {
			_jObj, err := ConvertBSONValueToJSON(timeNowMillis)
			So(err, ShouldBeNil)
			jObj, ok := _jObj.(json.Date)
			So(ok, ShouldBeTrue)

			So(int64(jObj), ShouldEqual, secs*1e3+millis)
		})

		Convey(fmt.Sprintf("should work with nanosecond granularity: %v", timeNow), func() {
			_jObj, err := ConvertBSONValueToJSON(timeNow)
			So(err, ShouldBeNil)
			jObj, ok := _jObj.(json.Date)
			So(ok, ShouldBeTrue)

			// we lose nanosecond precision
			So(int64(jObj), ShouldEqual, secs*1e3+millis)
		})

	})
}

func TestMaxKeyBSONToJSON(t *testing.T) {
	Convey("Converting a BSON Maxkey to JSON", t, func() {
		Convey("should produce a json.MaxKey", func() {
			_jObj, err := ConvertBSONValueToJSON(bson.MaxKey)
			So(err, ShouldBeNil)
			jObj, ok := _jObj.(json.MaxKey)
			So(ok, ShouldBeTrue)

			So(jObj, ShouldResemble, json.MaxKey{})
		})
	})
}

func TestMinKeyBSONToJSON(t *testing.T) {
	Convey("Converting a BSON Maxkey to JSON", t, func() {
		Convey("should produce a json.MinKey", func() {
			_jObj, err := ConvertBSONValueToJSON(bson.MinKey)
			So(err, ShouldBeNil)
			jObj, ok := _jObj.(json.MinKey)
			So(ok, ShouldBeTrue)

			So(jObj, ShouldResemble, json.MinKey{})
		})
	})
}

func Test64BitIntBSONToJSON(t *testing.T) {
	Convey("Converting a BSON int64 to JSON", t, func() {
		Convey("should produce a json.NumberLong", func() {
			_jObj, err := ConvertBSONValueToJSON(int32(243))
			So(err, ShouldBeNil)
			jObj, ok := _jObj.(json.NumberInt)
			So(ok, ShouldBeTrue)

			So(jObj, ShouldEqual, json.NumberInt(243))
		})
	})

}

func Test32BitIntBSONToJSON(t *testing.T) {
	Convey("Converting a BSON int32 integer to JSON", t, func() {
		Convey("should produce a json.NumberInt", func() {
			_jObj, err := ConvertBSONValueToJSON(int64(888234334343))
			So(err, ShouldBeNil)
			jObj, ok := _jObj.(json.NumberLong)
			So(ok, ShouldBeTrue)

			So(jObj, ShouldEqual, json.NumberLong(888234334343))
		})
	})

}

func TestRegExBSONToJSON(t *testing.T) {
	Convey("Converting a BSON Regular Expression (= /decision/gi) to JSON", t, func() {
		Convey("should produce a json.RegExp", func() {
			_jObj, err := ConvertBSONValueToJSON(bson.RegEx{Pattern: "decision", Options: "gi"})
			So(err, ShouldBeNil)
			jObj, ok := _jObj.(json.RegExp)
			So(ok, ShouldBeTrue)
			So(jObj, ShouldResemble, json.RegExp{Pattern: "decision", Options: "gi"})
		})
	})

}

func TestUndefinedValueBSONToJSON(t *testing.T) {
	Convey("Converting a BSON Undefined type to JSON", t, func() {
		Convey("should produce a json.Undefined", func() {
			_jObj, err := ConvertBSONValueToJSON(bson.Undefined)
			So(err, ShouldBeNil)
			jObj, ok := _jObj.(json.Undefined)
			So(ok, ShouldBeTrue)

			So(jObj, ShouldResemble, json.Undefined{})
		})
	})
}

func TestTimestampBSONToJSON(t *testing.T) {
	Convey("Converting a BSON Timestamp to JSON", t, func() {
		Convey("should produce a json.Timestamp", func() {
			// {t:803434343, i:9} == bson.MongoTimestamp(803434343*2**32 + 9)
			_jObj, err := ConvertBSONValueToJSON(bson.MongoTimestamp(uint64(803434343<<32) | uint64(9)))
			So(err, ShouldBeNil)
			jObj, ok := _jObj.(json.Timestamp)
			So(ok, ShouldBeTrue)

			So(jObj, ShouldResemble, json.Timestamp{Seconds: 803434343, Increment: 9})
			So(jObj, ShouldNotResemble, json.Timestamp{Seconds: 803434343, Increment: 8})
		})
	})
}

func TestBinaryBSONToJSON(t *testing.T) {
	Convey("Converting BSON Binary data to JSON", t, func() {
		Convey("should produce a json.BinData", func() {
			_jObj, err := ConvertBSONValueToJSON(bson.Binary{Kind: '\x01', Data: []byte("\x05\x20\x02\xae\xf7")})
			So(err, ShouldBeNil)
			jObj, ok := _jObj.(json.BinData)
			So(ok, ShouldBeTrue)

			base64data1 := base64.StdEncoding.EncodeToString([]byte("\x05\x20\x02\xae\xf7"))
			base64data2 := base64.StdEncoding.EncodeToString([]byte("\x05\x20\x02\xaf\xf7"))
			So(jObj, ShouldResemble, json.BinData{Type: '\x01', Base64: base64data1})
			So(jObj, ShouldNotResemble, json.BinData{Type: '\x01', Base64: base64data2})
		})
	})
}

func TestGenericBytesBSONToJSON(t *testing.T) {
	Convey("Converting Go bytes to JSON", t, func() {
		Convey("should produce a json.BinData with Type=0x00 (Generic)", func() {
			_jObj, err := ConvertBSONValueToJSON([]byte("this is something that's cool"))
			So(err, ShouldBeNil)
			jObj, ok := _jObj.(json.BinData)
			So(ok, ShouldBeTrue)

			base64data := base64.StdEncoding.EncodeToString([]byte("this is something that's cool"))
			So(jObj, ShouldResemble, json.BinData{Type: 0x00, Base64: base64data})
			So(jObj, ShouldNotResemble, json.BinData{Type: 0x01, Base64: base64data})
		})
	})
}

func TestUnknownBSONTypeToJSON(t *testing.T) {
	Convey("Converting an unknown BSON type to JSON", t, func() {
		Convey("should produce an error", func() {
			_, err := ConvertBSONValueToJSON(func() {})
			So(err, ShouldNotBeNil)
		})
	})
}

func TestJSCodeBSONToJSON(t *testing.T) {
	Convey("Converting BSON Javascript code to JSON", t, func() {
		Convey("should produce a json.Javascript", func() {
			Convey("without scope if the scope for the BSON Javascript code is nil", func() {
				_jObj, err := ConvertBSONValueToJSON(bson.JavaScript{Code: "function() { return null; }", Scope: nil})
				So(err, ShouldBeNil)
				jObj, ok := _jObj.(json.JavaScript)
				So(ok, ShouldBeTrue)

				So(jObj, ShouldResemble, json.JavaScript{Code: "function() { return null; }", Scope: nil})
			})

			Convey("with scope if the scope for the BSON Javascript code is non-nil", func() {
				_jObj, err := ConvertBSONValueToJSON(bson.JavaScript{Code: "function() { return x; }", Scope: bson.M{"x": 2}})
				So(err, ShouldBeNil)
				jObj, ok := _jObj.(json.JavaScript)
				So(ok, ShouldBeTrue)
				So(jObj.Scope.(bson.M)["x"], ShouldEqual, 2)
				So(jObj.Code, ShouldEqual, "function() { return x; }")
			})
		})
	})
}
