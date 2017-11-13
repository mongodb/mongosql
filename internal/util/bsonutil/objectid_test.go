package bsonutil

import (
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/json"
	. "github.com/smartystreets/goconvey/convey"
)

func TestObjectIdValue(t *testing.T) {

	Convey("When converting JSON with ObjectID values", t, func() {

		Convey("works for ObjectID constructor", func() {
			key := "key"
			jsonMap := map[string]interface{}{
				key: json.ObjectID("0123456789abcdef01234567"),
			}

			err := ConvertJSONDocumentToBSON(jsonMap)
			So(err, ShouldBeNil)
			So(jsonMap[key], ShouldEqual, bson.ObjectIdHex("0123456789abcdef01234567"))
		})

		Convey(`works for ObjectID document ('{ "$oid": "0123456789abcdef01234567" }')`, func() {
			key := "key"
			jsonMap := map[string]interface{}{
				key: map[string]interface{}{
					"$oid": "0123456789abcdef01234567",
				},
			}

			err := ConvertJSONDocumentToBSON(jsonMap)
			So(err, ShouldBeNil)
			So(jsonMap[key], ShouldEqual, bson.ObjectIdHex("0123456789abcdef01234567"))
		})
	})
}
