package translator

import (
	"github.com/siddontang/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func getQuery(raw sqlparser.Statement) string {
	buf := sqlparser.NewTrackedBuffer(nil)
	raw.Format(buf)
	return buf.String()
}

func TestAlgebrizeTableExpr(t *testing.T) {

	Convey("With a simple SQL statement...", t, func() {

		Convey("algebrizing table names should produce the correct algebrized nodes", func() {

			sql := "select * from foo f"

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)
			stmt := raw.(*sqlparser.Select)
			ctx, err := NewParseCtx(stmt)
			So(err, ShouldBeNil)
			So(algebrizeStatement(stmt, ctx), ShouldBeNil)
		})

		Convey("algebrizing column names should produce the correct algebrized nodes", func() {

			sql := "select firstname as x, lastname as y from foo f where x = eliot"
			algebrizedSQL := "select firstname as x, lastname as y from foo as f where firstname = eliot"

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt := raw.(*sqlparser.Select)
			ctx, err := NewParseCtx(stmt)
			So(err, ShouldBeNil)

			So(algebrizeStatement(stmt, ctx), ShouldBeNil)

			// check that statement is properly algebrized
			So(getQuery(raw), ShouldEqual, algebrizedSQL)
		})

		Convey("algebrizing subqueries should produce the correct algebrized nodes", func() {

			sql := `select f.first as c, f.last, (select age as o from foo where o > 34) from foo f where c = eliot and f.last = horowitz`
			algebrizedSQL := `select f.first as c, f.last, (select age as o from foo where age > 34) from foo as f where first = eliot and last = horowitz`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt := raw.(*sqlparser.Select)
			ctx, err := NewParseCtx(stmt)
			So(err, ShouldBeNil)

			So(algebrizeStatement(stmt, ctx), ShouldBeNil)

			// check that statement is properly algebrized
			So(getQuery(raw), ShouldEqual, algebrizedSQL)
		})

		Convey("algebrizing join statements should produce the correct algebrized nodes", func() {

			sql := `select o.orderid, o.customername, o.orderdate from orders as o join customers as c on orders.customerid = customers.customerid where c.customerid > 4 and c.customerid < 9`
			algebrizedSQL := `select o.orderid, o.customername, o.orderdate from orders as o join customers as c on orders.customerid = customers.customerid where customerid > 4 and customerid < 9`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt := raw.(*sqlparser.Select)
			ctx, err := NewParseCtx(stmt)
			So(err, ShouldBeNil)

			So(algebrizeStatement(stmt, ctx), ShouldBeNil)

			// check that statement is properly algebrized
			So(getQuery(raw), ShouldEqual, algebrizedSQL)
		})

	})
}
