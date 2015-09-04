package algebrizer

import (
	"github.com/erh/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewParseCtx(t *testing.T) {

	Convey("With a simple SQL statement...", t, func() {
		Convey("simple select statements should produce correct parse contexts", func() {

			sql := "select * from foo f"

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt := raw.(*sqlparser.Select)
			ctx, err := NewParseCtx(stmt)
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)

			So(len(ctx.Table), ShouldEqual, 1)
			So(ctx.Table[0].Alias, ShouldResemble, "f")
			So(ctx.Table[0].Collection, ShouldResemble, "foo")

		})
	})
}
