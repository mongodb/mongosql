package evaluator

import (
	"testing"

	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEmptyOperator(t *testing.T) {

	Convey("When using the empty operator", t, func() {

		e := EmptyStage{
			[]*Column{
				&Column{
					Table:     "foo",
					Name:      "a",
					SQLType:   schema.SQLInt,
					MongoType: schema.MongoInt,
				},
			},
		}

		Convey("Open should return nil error", func() {
			iter, err := e.Open(nil)
			So(err, ShouldBeNil)

			Convey("Next should return false", func() {
				So(iter.Next(nil), ShouldBeFalse)
			})

			Convey("Columns should return the table fields", func() {
				res := e.Columns()
				So(len(res), ShouldEqual, 1)
				So(res[0].Table, ShouldEqual, "foo")
				So(res[0].Name, ShouldEqual, "a")
				So(res[0].SQLType, ShouldEqual, schema.SQLInt)
				So(res[0].MongoType, ShouldEqual, schema.MongoInt)
			})

			Convey("Close should return nil", func() {
				res := iter.Close()
				So(res, ShouldBeNil)
			})

			Convey("Err should return nil", func() {
				res := iter.Err()
				So(res, ShouldBeNil)
			})
		})
	})

}
