package mongo_test

import (
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/schema/mongo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAddIndexes(t *testing.T) {
	Convey("Given a new collection schema", t, func() {
		schema := mongo.NewCollectionSchema()

		Convey("with an included sample", func() {
			err := schema.IncludeSample(bson.D{
				{Name: "a", Value: int32(1)},
				{Name: "loc", Value: []interface{}{"a", "b", "c"}},
				{Name: "b", Value: bson.D{
					{Name: "geo", Value: true},
				}},
			})
			So(err, ShouldBeNil)

			Convey("Adding some indexes", func() {

				indexes := []bson.D{
					{bson.DocElem{Name: "key", Value: bson.D{
						bson.DocElem{Name: "loc", Value: "2d"},
					}}},
					{bson.DocElem{Name: "key", Value: bson.D{
						bson.DocElem{Name: "b.geo", Value: "2dsphere"},
					}}},
				}

				schema.AddIndexes(indexes)

				Convey("Should add indexes to the appropriate schematas", func() {
					aIdxs := schema.Properties["a"].Indexes
					locIdxs := schema.Properties["loc"].Indexes
					bIdxs := schema.Properties["b"].Indexes
					bGeoIdxs := schema.Properties["b"].DominantSchema().Properties["geo"].Indexes

					So(aIdxs, ShouldHaveLength, 0)
					So(bIdxs, ShouldHaveLength, 0)
					So(locIdxs, ShouldHaveLength, 1)
					So(bGeoIdxs, ShouldHaveLength, 1)
					So(locIdxs[0], ShouldEqual, mongo.Index2D)
					So(bGeoIdxs[0], ShouldEqual, mongo.Index2DSphere)
				})
			})
		})
	})
}
