package evaluator

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestEmptyOperator(t *testing.T) {

	Convey("When using the empty operator", t, func() {

		e := EmptyStage{}

		Convey("Open should return nil error", func() {
			iter, err := e.Open(nil)
			So(err, ShouldBeNil)

			Convey("Next should return false", func() {
				So(iter.Next(nil), ShouldBeFalse)
			})

			Convey("OpFields should return an empty array", func() {
				res := e.OpFields()
				So(res, ShouldBeEmpty)
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
