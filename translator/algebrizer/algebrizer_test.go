package algebrizer

import (
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/config"
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

		Convey("algebrizing column names should produce the correct algebrized nodes", func() {

			sql := `select f.first from foo f where f.first = 'eliot'`
			algebrizedSQL := `select foo.first from foo as f where foo.first = 'eliot'`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)

			cfg, err := config.ParseConfigData(testConfigSimple)
			So(err, ShouldBeNil)
			ctx, err := NewParseCtx(stmt, cfg, dbName)
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)

			So(AlgebrizeStatement(stmt, ctx), ShouldBeNil)

			// check that statement is properly algebrized
			So(getQuery(raw), ShouldEqual, algebrizedSQL)
		})

		Convey("algebrizing table names should produce the correct algebrized nodes", func() {

			sql := "select * from foo f"
			algebrizedSQL := "select * from foo as f"

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)
			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)
			cfg, err := config.ParseConfigData(testConfigSimple)
			So(err, ShouldBeNil)
			ctx, err := NewParseCtx(stmt, cfg, dbName)
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)
			So(AlgebrizeStatement(stmt, ctx), ShouldBeNil)

			// check that statement is properly algebrized
			So(getQuery(raw), ShouldEqual, algebrizedSQL)
		})

		Convey("algebrizing a query with implicit table reference with where should produce the correct algebrized nodes", func() {

			sql := "select a from foo f where x > 3"
			algebrizedSQL := "select foo.a from foo as f where foo.x > 3"

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)
			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)
			cfg, err := config.ParseConfigData(testConfigSimple)
			So(err, ShouldBeNil)
			ctx, err := NewParseCtx(stmt, cfg, dbName)
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)
			So(AlgebrizeStatement(stmt, ctx), ShouldBeNil)

			// check that statement is properly algebrized
			So(getQuery(raw), ShouldEqual, algebrizedSQL)
		})
		Convey("algebrizing subqueries should produce the correct algebrized nodes", func() {

			sql := `select f.first, f.last, (select f.age from foo f where f.age > 34) from foo f where f.first = 'eliot' and f.last = 'horowitz'`
			algebrizedSQL := `select foo.first, foo.last, (select foo.age from foo as f where foo.age > 34) from foo as f where foo.first = 'eliot' and foo.last = 'horowitz'`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)
			cfg, err := config.ParseConfigData(testConfigSimple)
			So(err, ShouldBeNil)
			ctx, err := NewParseCtx(stmt, cfg, dbName)
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)

			So(AlgebrizeStatement(stmt, ctx), ShouldBeNil)

			// check that statement is properly algebrized
			So(getQuery(raw), ShouldEqual, algebrizedSQL)
		})

		Convey("algebrizing join statements should produce the correct algebrized nodes", func() {

			sql := `select o.orderid, o.customername, o.orderdate from orders as o join customers as c on o.customerid = c.customerid where c.customerid > 4 and c.customerid < 9`
			algebrizedSQL := `select orders.orderid, orders.customername, orders.orderdate from orders as o join customers as c on orders.customerid = customers.customerid where customers.customerid > 4 and customers.customerid < 9`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)
			cfg, err := config.ParseConfigData(testConfigSimple)
			So(err, ShouldBeNil)
			ctx, err := NewParseCtx(stmt, cfg, dbName)
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)

			So(AlgebrizeStatement(stmt, ctx), ShouldBeNil)

			// check that statement is properly algebrized
			So(getQuery(raw), ShouldEqual, algebrizedSQL)
		})

		Convey("algebrizing a subquery that references outer aliased nodes should properly algebrized", func() {

			sql := `select f3.z, (select f1.a from baz f2 where f3.a = f2.b ), f1.b from foo f1, bar f3`
			algebrizedSQL := `select bar.z, (select foo.a from baz as f2 where bar.a = baz.b), foo.b from foo as f1, bar as f3`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)
			cfg, err := config.ParseConfigData(testConfigSimple)
			So(err, ShouldBeNil)
			ctx, err := NewParseCtx(stmt, cfg, dbName)
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)

			So(AlgebrizeStatement(stmt, ctx), ShouldBeNil)

			// check that statement is properly algebrized
			So(getQuery(raw), ShouldEqual, algebrizedSQL)
		})

		Convey("algebrizing a derived table should require an alias", func() {

			sql := `select o.orderid from (select * from orders) join customers on o.customerid = customers.customerid`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)

			cfg, err := config.ParseConfigData(testConfigSimple)
			So(err, ShouldBeNil)

			_, err = NewParseCtx(stmt, cfg, dbName)
			So(err, ShouldNotBeNil)
		})

		Convey("algebrizing a derived table with an alias different from referenced select expression should fail", func() {

			sql := `select o.orderid from (select * from orders) as f join customers on o.customerid = customers.customerid`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)

			cfg, err := config.ParseConfigData(testConfigSimple)
			So(err, ShouldBeNil)

			ctx, err := NewParseCtx(stmt, cfg, dbName)
			So(err, ShouldBeNil)

			So(AlgebrizeStatement(stmt, ctx), ShouldNotBeNil)
		})

		Convey("algebrizing a derived table with an alias same from referenced select expression should pass", func() {

			sql := `select f.orderid from (select * from orders) as f join customers on f.customerid = customers.customerid`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)

			cfg, err := config.ParseConfigData(testConfigSimple)
			So(err, ShouldBeNil)

			ctx, err := NewParseCtx(stmt, cfg, dbName)
			So(err, ShouldBeNil)

			So(AlgebrizeStatement(stmt, ctx), ShouldBeNil)
		})
		// TODO: negative test cases
	})
}
