package translator

import (
	"github.com/siddontang/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestAlgebrizeTableExpr(t *testing.T) {

	Convey("With a simple SQL statement...", t, func() {

		Convey("algebrizing the parsed statements should produce the correct nodes", func() {

			sql := "select * from foo f"

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)
			stmt := raw.(*sqlparser.Select)
			algebrized, err := algebrizeTableExpr(stmt.From[0])
			So(err, ShouldBeNil)
			So(len(algebrized.Nodes), ShouldEqual, 1)
			So(algebrized.Nodes[0].depth, ShouldEqual, 0)
			So(algebrized.Nodes[0].nAlias, ShouldEqual, "f")
			So(algebrized.Nodes[0].nName, ShouldEqual, "foo")
			So(algebrized.Nodes[0].nType, ShouldEqual, Table)
		})
	})
}
