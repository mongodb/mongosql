package bsonutil

import (
	"github.com/10gen/mongo-go-driver/bson"
	"github.com/mongodb/mongo-tools/common/json"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestMaxKeyValue(t *testing.T) {

	Convey("When converting JSON with MaxKey values", t, func() {

		Convey("works for MaxKey literal", func() {
			key := "key"
			jsonMap := map[string]interface{}{
				key: json.MaxKey{},
			}

			err := ConvertJSONDocumentToBSON(jsonMap)
			So(err, ShouldBeNil)
			So(jsonMap[key], ShouldResemble, bson.MaxKey)
		})

		Convey(`works for MaxKey document ('{ "$maxKey": 1 }')`, func() {
			key := "key"
			jsonMap := map[string]interface{}{
				key: map[string]interface{}{
					"$maxKey": 1,
				},
			}

			err := ConvertJSONDocumentToBSON(jsonMap)
			So(err, ShouldBeNil)
			So(jsonMap[key], ShouldResemble, bson.MaxKey)
		})
	})
}
