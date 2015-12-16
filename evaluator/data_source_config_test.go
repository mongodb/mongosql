package evaluator

import (
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"sort"
	"testing"
)

func TestSchemaDataSourceIter(t *testing.T) {

	Convey("using config data source should iterate all columns", t, func() {

		execCtx := &ExecutionCtx{
			Schema:  cfgThree,
			Session: session,
			Db:      dbOne,
		}

		dataSource := SchemaDataSource{ctx: execCtx, includeColumns: true}

		query := dataSource.Find()

		iter := query.Iter()

		fieldNames := []string{}

		var doc bson.D
		for iter.Next(&doc) {
			v, found := getKey("COLUMN_NAME", doc)
			So(found, ShouldBeTrue)
			fieldNames = append(fieldNames, v.(string))
		}

		So(len(fieldNames), ShouldEqual, 7)

		names := []string{"_id", "a", "b", "c", "d", "e", "f"}
		sort.Strings(fieldNames)
		So(names, ShouldResemble, fieldNames)
		So(iter.Close(), ShouldBeNil)
	})
}

func TestSchemaDataSourceIterTables(t *testing.T) {

	Convey("using config data source should iterate tables", t, func() {

		execCtx := &ExecutionCtx{
			Schema:  cfgOne,
			Session: session,
			Db:      dbOne,
		}

		dataSource := SchemaDataSource{ctx: execCtx}

		query := dataSource.Find()

		iter := query.Iter()

		names := []string{}

		var doc bson.D
		for iter.Next(&doc) {
			v, found := getKey("TABLE_NAME", doc)
			So(found, ShouldBeTrue)
			names = append(names, v.(string))
		}

		So(len(names), ShouldEqual, 6)

		tableNames := []string{"bar", "bar", "bar", "foo", "foo", "silly"}
		sort.Strings(names)
		So(tableNames, ShouldResemble, names)
		So(iter.Close(), ShouldBeNil)
	})
}
