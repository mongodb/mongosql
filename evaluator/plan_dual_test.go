package evaluator_test

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	_ fmt.Stringer = nil
)

func TestDualOperator(t *testing.T) {
	Convey("A dual operator...", t, func() {

		Convey("should only ever return one row with no data", func() {

			operator := &evaluator.DualStage{}

			ctx := &evaluator.ExecutionCtx{}

			iter, err := operator.Open(ctx)
			So(err, ShouldBeNil)

			row := &evaluator.Row{}

			i := 0

			for iter.Next(row) {
				So(len(row.Data), ShouldEqual, 0)
				i++
			}

			So(i, ShouldEqual, 1)

			So(iter.Close(), ShouldBeNil)
			So(iter.Err(), ShouldBeNil)

		})
	})
}
