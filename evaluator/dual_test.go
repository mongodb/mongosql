package evaluator

import (
	"fmt"
	"github.com/erh/mongo-sql-temp/config"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	_ fmt.Stringer = nil
)

func TestDualOperator(t *testing.T) {

	Convey("A dual operator...", t, func() {

		Convey("should only ever return one row with no data", func() {
			cfg, err := config.ParseConfigData(testConfig1)
			So(err, ShouldBeNil)

			operator := &Dual{}

			ctx := &ExecutionCtx{
				Config: cfg,
				Db:     dbName,
			}

			So(operator.Open(ctx), ShouldBeNil)

			row := &Row{}

			i := 0

			for operator.Next(row) {
				So(len(row.Data), ShouldEqual, 0)
				i++
			}

			So(i, ShouldEqual, 1)
			So(operator.Err(), ShouldBeNil)
			So(operator.Close(), ShouldBeNil)

		})
	})
}
