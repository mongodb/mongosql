package mongo_test

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/schema/mapping"
	"github.com/10gen/sqlproxy/internal/schema/mongo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAddIndexes(t *testing.T) {
	Convey("Given a new collection schema", t, func() {
		schema := mongo.NewCollectionSchema()

		Convey("with an included sample", func() {
			err := schema.IncludeSample(bsonutil.NewD(
				bsonutil.NewDocElem("a", int32(1)),
				bsonutil.NewDocElem("loc", bsonutil.NewArray(
					"a",
					"b",
					"c",
				)),
				bsonutil.NewDocElem("b", bsonutil.NewD(
					bsonutil.NewDocElem("geo", true),
				)),
			))
			So(err, ShouldBeNil)

			Convey("Adding some indexes", func() {

				indexes := bsonutil.NewDArray(
					bsonutil.NewD(bsonutil.NewDocElem("key", bsonutil.NewD(bsonutil.NewDocElem("loc", "2d")))),
					bsonutil.NewD(bsonutil.NewDocElem("key", bsonutil.NewD(bsonutil.NewDocElem("b.geo", "2dsphere")))),
				)

				schema.AddIndexes(indexes)

				Convey("Should add indexes to the appropriate schematas", func() {
					aIdxs := schema.Properties["a"].Indexes
					locIdxs := schema.Properties["loc"].Indexes
					bIdxs := schema.Properties["b"].Indexes
					g := "geo"
					fmt.Println(schema.Properties["b"])
					bGeoIdxs := mapping.PolymorphicMajorityCountHeuristic(
						schema.Properties["b"])[0].Properties[g].Indexes

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
