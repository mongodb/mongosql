package translator

import (
	"github.com/erh/mongo-sql-temp/config"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"sort"
	"testing"
)

func TestConfigDataSourceIter(t *testing.T) {

	Convey("using config data source should work", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)
		
		dataSource := ConfigDataSource{cfg, true}
		
		query := dataSource.Find(bson.M{})
		
		iter := query.Iter()
		
		var doc bson.M

		fieldNames := []string{}

		for iter.Next(&doc) {
			fieldNames = append(fieldNames, doc["COLUMN_NAME"].(string) )
		}
		
		So(len(fieldNames), ShouldEqual, 6)

		sort.Strings(fieldNames)
		So([]string{"a", "b", "c", "d", "e", "f"}, ShouldResemble, fieldNames)
	})
}

func TestConfigDataSourceSelect(t *testing.T) {

	Convey("using config data source should work", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)

		eval, err := NewEvalulator(cfg)
		So(err, ShouldBeNil)

		_, values, err := eval.EvalSelect("information_schema", "select * from columns", nil)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 6)

		_, values, err = eval.EvalSelect("information_schema", "select * from columns where COLUMN_NAME = 'f'", nil)
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 1)

	})
}

func TestConfigDataSourceIterTables(t *testing.T) {

	Convey("using config data source should work", t, func() {

		cfg, err := config.ParseConfigData(testConfigSimple)
		So(err, ShouldBeNil)
		
		dataSource := ConfigDataSource{cfg, false}
		
		query := dataSource.Find(bson.M{})
		
		iter := query.Iter()
		
		var doc bson.M

		names := []string{}

		for iter.Next(&doc) {
			names = append(names, doc["TABLE_NAME"].(string) )
		}
		
		So(len(names), ShouldEqual, 3)

		sort.Strings(names)
		So([]string{"bar", "bar", "silly"}, ShouldResemble, names)
	})
}
