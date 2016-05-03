package evaluator

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	_ fmt.Stringer = nil
)

func TestDualOperator(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne

	Convey("A dual operator...", t, func() {

		Convey("should only ever return one row with no data", func() {

			operator := &DualStage{}

			ctx := &ExecutionCtx{
				PlanCtx: &PlanCtx{
					Schema: cfgOne,
					Db:     dbOne,
				},
			}

			iter, err := operator.Open(ctx)
			So(err, ShouldBeNil)

			row := &Row{}

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
