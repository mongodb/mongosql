package planner

import (
	"github.com/erh/mongo-sql-temp/config"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"sort"
	"testing"
)

func TestConfigDataSourceIter(t *testing.T) {

	Convey("using config data source should iterate all columns", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		dataSource := ConfigDataSource{cfg, true}

		query := dataSource.Find(EmptyMatcher{})

		iter := query.Iter()

		fieldNames := []string{}

		var doc bson.D
		for iter.Next(&doc) {
			v, found := getKey("COLUMN_NAME", doc)
			So(found, ShouldBeTrue)
			fieldNames = append(fieldNames, v.(string))
		}

		So(len(fieldNames), ShouldEqual, 6)

		sort.Strings(fieldNames)
		So([]string{"a", "b", "c", "d", "e", "f"}, ShouldResemble, fieldNames)
	})
}

func TestConfigDataSourceIterTables(t *testing.T) {

	Convey("using config data source should iterate tables", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		dataSource := ConfigDataSource{cfg, false}

		query := dataSource.Find(EmptyMatcher{})

		iter := query.Iter()

		names := []string{}

		var doc bson.D
		for iter.Next(&doc) {
			v, found := getKey("TABLE_NAME", doc)
			So(found, ShouldBeTrue)
			names = append(names, v.(string))
		}

		So(len(names), ShouldEqual, 4)

		sort.Strings(names)
		So([]string{"bar", "bar", "foo", "silly"}, ShouldResemble, names)
	})
}
