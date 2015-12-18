package evaluator

import (
	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewParseCtx(t *testing.T) {

	Convey("With a simple SQL statement...", t, func() {
		Convey("simple select statements should produce correct parse contexts", func() {

			sql := "select * from foo f"

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)
			cfg, err := schema.ParseSchemaData(testSchema2)
			So(err, ShouldBeNil)
			ctx, err := NewParseCtx(stmt, cfg, dbOne)
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)

			So(len(ctx.Tables), ShouldEqual, 1)
			So(ctx.Tables[0].Alias, ShouldResemble, "f")
			So(ctx.Tables[0].Name, ShouldResemble, "foo")

		})
	})
}
