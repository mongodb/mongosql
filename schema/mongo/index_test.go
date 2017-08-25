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
					a_idxs := schema.Properties["a"].Indexes
					loc_idxs := schema.Properties["loc"].Indexes
					b_idxs := schema.Properties["b"].Indexes
					b_geo_idxs := schema.Properties["b"].DominantSchema().Properties["geo"].Indexes

					So(a_idxs, ShouldHaveLength, 0)
					So(b_idxs, ShouldHaveLength, 0)
					So(loc_idxs, ShouldHaveLength, 1)
					So(b_geo_idxs, ShouldHaveLength, 1)
					So(loc_idxs[0], ShouldEqual, Index2D)
					So(b_geo_idxs[0], ShouldEqual, Index2DSphere)
				})
			})
		})
	})
}
