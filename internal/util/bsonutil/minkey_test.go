package bsonutil

import (
	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/json"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestMinKeyValue(t *testing.T) {

	Convey("When converting JSON with MinKey values", t, func() {

		Convey("works for MinKey literal", func() {
			key := "key"
			jsonMap := map[string]interface{}{
				key: json.MinKey{},
			}

			err := ConvertJSONDocumentToBSON(jsonMap)
			So(err, ShouldBeNil)
			So(jsonMap[key], ShouldResemble, bson.MinKey)
		})

		Convey(`works for MinKey document ('{ "$minKey": 1 }')`, func() {
			key := "key"
			jsonMap := map[string]interface{}{
				key: map[string]interface{}{
					"$minKey": 1,
				},
			}

			err := ConvertJSONDocumentToBSON(jsonMap)
			So(err, ShouldBeNil)
			So(jsonMap[key], ShouldResemble, bson.MinKey)
		})
	})
}
