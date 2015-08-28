package translator

import (
	"github.com/erh/mongo-sql-temp/config"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var testConfigSimple = []byte(
	`
schema :
-
  url: localhost
  db: test
  tables:
  -
     table: bar
     collection: test.simple
`)

func TestSimple(t *testing.T) {

	Convey("With a simple test configuration...", t, func() {

		Convey("connecting through the proxy server should return correct results", func() {

			cfg, err := config.ParseConfigData(testConfigSimple)
			So(err, ShouldBeNil)

			eval, err := NewEvalulator(cfg)
			So(err, ShouldBeNil)

			session := eval.getSession()
			defer session.Close()

			collection := eval.getCollection(session, "test.simple")
			collection.DropCollection()
			So(collection.Insert(bson.M{"_id": 5, "a": 6, "b": 7}), ShouldBeNil)
			So(collection.Insert(bson.M{"_id": 15, "a": 16, "c": 17}), ShouldBeNil)

			names, values, err := eval.EvalSelect("test", "select * from bar", nil)
			So(err, ShouldBeNil)
			So(len(names), ShouldEqual, 4)
			So(len(values), ShouldEqual, 2)

			So(names[0], ShouldEqual, "_id")
			So(names[1], ShouldEqual, "a")
			So(names[2], ShouldEqual, "b")
			So(names[3], ShouldEqual, "c")

			So(values[1][0], ShouldEqual, 15)
			So(values[1][1], ShouldEqual, 16)
			So(values[1][2], ShouldEqual, nil)
			So(values[1][3], ShouldEqual, 17)

			for _, row := range values {
				So(len(names), ShouldEqual, len(row))
			}

			names, values, err = eval.EvalSelect("test", "select * from bar where a = 16", nil)
			So(err, ShouldBeNil)
			So(len(values), ShouldEqual, 1)

			names, values, err = eval.EvalSelect("", "select * from test.bar where a = 16", nil)
			So(err, ShouldBeNil)
			So(len(values), ShouldEqual, 1)

			names, values, err = eval.EvalSelect("xxx", "select * from test.bar where a = 16", nil)
			So(err, ShouldBeNil)
			So(len(values), ShouldEqual, 1)

		})
	})
}
