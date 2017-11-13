package mongo

import (
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAddIndexes(t *testing.T) {
	Convey("Given a new collection schema", t, func() {
		schema := NewCollectionSchema()

		Convey("with an included sample", func() {
			err := schema.IncludeSample(bson.D{
				{"a", int32(1)},
				{"loc", []interface{}{"a", "b", "c"}},
				{"b", bson.D{
					{"geo", true},
				}},
			})
			So(err, ShouldBeNil)

			Convey("Adding some indexes", func() {
				idxs := map[string]IndexType{
					"loc":   Index2D,
					"b.geo": Index2DSphere,
				}
				schema.addIndexes(idxs, "")

				Convey("Should add indexes to the appropriate schematas", func() {
					aIdxs := schema.Properties["a"].Indexes
					locIdxs := schema.Properties["loc"].Indexes
					bIdxs := schema.Properties["b"].Indexes
					bGeoIdxs := schema.Properties["b"].DominantSchema().Properties["geo"].Indexes

					So(aIdxs, ShouldHaveLength, 0)
					So(bIdxs, ShouldHaveLength, 0)
					So(locIdxs, ShouldHaveLength, 1)
					So(bGeoIdxs, ShouldHaveLength, 1)
					So(locIdxs[0], ShouldEqual, Index2D)
					So(bGeoIdxs[0], ShouldEqual, Index2DSphere)
				})
			})
		})
	})
}
