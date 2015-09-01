package translator

import (
	"github.com/erh/mongo-sql-temp/config"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func TestConfigDataSourceIter(t *testing.T) {

	Convey("With a simple test configuration...", t, func() {

		Convey("using config data source should work", func() {

			cfg, err := config.ParseConfigData(testConfigSimple)
			So(err, ShouldBeNil)

			dataSource := ConfigDataSource{cfg}

			query := dataSource.Find(bson.M{})

			iter := query.Iter()

			var doc bson.M

			So(iter.Next(&doc), ShouldBeTrue)
			So(iter.Next(&doc), ShouldBeTrue)
			So(iter.Next(&doc), ShouldBeTrue)
			So(iter.Next(&doc), ShouldBeTrue)
			So(iter.Next(&doc), ShouldBeTrue)
			So(iter.Next(&doc), ShouldBeTrue)
			So(doc["TABLE_NAME"], ShouldEqual, "silly")
			So(doc["COLUMN_NAME"], ShouldEqual, "f")
			So(iter.Next(&doc), ShouldBeFalse)
		})
	})
}
