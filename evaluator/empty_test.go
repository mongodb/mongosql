package evaluator

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestEmptyOperator(t *testing.T) {

	Convey("When using the empty operator", t, func() {

		e := Empty{}

		Convey("Open should return nil", func() {
			res := e.Open(nil)
			So(res, ShouldBeNil)
		})

		Convey("Next should return false", func() {
			res := e.Next(nil)
			So(res, ShouldBeFalse)
		})

		Convey("OpFields should return an empty array", func() {
			res := e.OpFields()
			So(res, ShouldBeEmpty)
		})

		Convey("Close should return nil", func() {
			res := e.Close()
			So(res, ShouldBeNil)
		})

		Convey("Err should return nil", func() {
			res := e.Err()
			So(res, ShouldBeNil)
		})

	})

}
