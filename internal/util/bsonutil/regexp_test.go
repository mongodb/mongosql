package bsonutil

import (
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/json"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRegExpValue(t *testing.T) {

	Convey("When converting JSON with RegExp values", t, func() {

		Convey("works for RegExp constructor", func() {
			key := "key"
			jsonMap := map[string]interface{}{
				key: json.RegExp{Pattern: "foo", Options: "i"},
			}

			err := ConvertJSONDocumentToBSON(jsonMap)
			So(err, ShouldBeNil)
			So(jsonMap[key], ShouldResemble, bson.RegEx{Pattern: "foo", Options: "i"})
		})

		Convey(`works for RegExp document ('{ "$regex": "foo", "$options": "i" }')`, func() {
			key := "key"
			jsonMap := map[string]interface{}{
				key: map[string]interface{}{
					"$regex":   "foo",
					"$options": "i",
				},
			}

			err := ConvertJSONDocumentToBSON(jsonMap)
			So(err, ShouldBeNil)
			So(jsonMap[key], ShouldResemble, bson.RegEx{Pattern: "foo", Options: "i"})
		})

		Convey(`can use multiple options ('{ "$regex": "bar", "$options": "gims" }')`, func() {
			key := "key"
			jsonMap := map[string]interface{}{
				key: map[string]interface{}{
					"$regex":   "bar",
					"$options": "gims",
				},
			}

			err := ConvertJSONDocumentToBSON(jsonMap)
			So(err, ShouldBeNil)
			So(jsonMap[key], ShouldResemble, bson.RegEx{Pattern: "bar", Options: "gims"})
		})

		Convey(`fails for an invalid option ('{ "$regex": "baz", "$options": "y" }')`, func() {
			key := "key"
			jsonMap := map[string]interface{}{
				key: map[string]interface{}{
					"$regex":   "baz",
					"$options": "y",
				},
			}

			err := ConvertJSONDocumentToBSON(jsonMap)
			So(err, ShouldNotBeNil)
		})
	})
}
