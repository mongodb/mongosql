package evaluator

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	_ fmt.Stringer = nil
)

func TestDualOperator(t *testing.T) {

	Convey("A dual operator...", t, func() {

		Convey("should only ever return one row with no data", func() {

			operator := &Dual{}

			ctx := &ExecutionCtx{
				Schema: cfgOne,
				Db:     dbOne,
			}

			So(operator.Open(ctx), ShouldBeNil)

			row := &Row{}

			i := 0

			for operator.Next(row) {
				So(len(row.Data), ShouldEqual, 0)
				i++
			}

			So(i, ShouldEqual, 1)

			So(operator.Close(), ShouldBeNil)
			So(operator.Err(), ShouldBeNil)

		})
	})
}
