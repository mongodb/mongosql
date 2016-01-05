package evaluator

import (
	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func getQuery(raw sqlparser.Statement) string {
	buf := sqlparser.NewTrackedBuffer(nil)
	raw.Format(buf)
	return buf.String()
}

func TestAlgebrizeTableExpr(t *testing.T) {

	Convey("When algerizing a select expression...", t, func() {

		Convey("aliased table names should be correctly parsed", func() {

			sql := "select * from foo f"
			algebrizedSQL := "select f.a, f.x, f.first, f.last, f.age, f.b from foo as f"

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)
			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)
			cfg, err := schema.ParseSchemaData(testSchema2)
			So(err, ShouldBeNil)
			ctx, err := NewParseCtx(stmt, cfg, dbOne)
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)
			So(AlgebrizeStatement(stmt, ctx), ShouldBeNil)

			// check that statement is properly algebrized
			So(getQuery(raw), ShouldEqual, algebrizedSQL)
		})

		Convey("a query with an implicit WHERE clause reference should produce the correct algebrized nodes", func() {

			sql := "select a from foo f where x > 3"

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)
			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)
			cfg, err := schema.ParseSchemaData(testSchema2)
			So(err, ShouldBeNil)
			ctx, err := NewParseCtx(stmt, cfg, dbOne)
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)
			So(AlgebrizeStatement(stmt, ctx), ShouldBeNil)

		})
		Convey("subqueries should produce the correct algebrized nodes", func() {

			sql := `select f.first, f.last, (select f.age from foo f where f.age > 34) from foo f where f.first = 'eliot' and f.last = 'horowitz'`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)
			cfg, err := schema.ParseSchemaData(testSchema2)
			So(err, ShouldBeNil)
			ctx, err := NewParseCtx(stmt, cfg, dbOne)
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)

			So(AlgebrizeStatement(stmt, ctx), ShouldBeNil)
		})

		Convey("join statements should produce the correct algebrized nodes", func() {

			sql := `select o.orderid, o.customername, o.orderdate from orders as o join customers as c on o.customerid = c.customerid where c.customerid > 4 and c.customerid < 9`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)
			cfg, err := schema.ParseSchemaData(testSchema2)
			So(err, ShouldBeNil)
			ctx, err := NewParseCtx(stmt, cfg, dbOne)
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)

			So(AlgebrizeStatement(stmt, ctx), ShouldBeNil)
		})

		Convey("a subquery that references outer aliased nodes should properly algebrized", func() {

			sql := `select f3.z, (select f1.a from baz f2 where f3.a = f2.b), f1.b from foo f1, bar f3`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)
			cfg, err := schema.ParseSchemaData(testSchema2)
			So(err, ShouldBeNil)
			ctx, err := NewParseCtx(stmt, cfg, dbOne)
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)

			So(AlgebrizeStatement(stmt, ctx), ShouldBeNil)
		})

		Convey("a derived table should require an alias", func() {

			sql := `select o.orderid from (select * from orders) join customers on o.customerid = customers.customerid`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)

			cfg, err := schema.ParseSchemaData(testSchema2)
			So(err, ShouldBeNil)

			_, err = NewParseCtx(stmt, cfg, dbOne)
			So(err, ShouldNotBeNil)
		})

		Convey("a derived table with an alias different from referenced select expression should fail", func() {

			sql := `select o.orderid from (select * from orders) as f join customers on o.customerid = customers.customerid`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)

			cfg, err := schema.ParseSchemaData(testSchema2)
			So(err, ShouldBeNil)

			ctx, err := NewParseCtx(stmt, cfg, dbOne)
			So(err, ShouldBeNil)

			So(AlgebrizeStatement(stmt, ctx), ShouldNotBeNil)
		})

		Convey("a derived table with an alias same from referenced select expression should pass", func() {

			sql := `select f.orderid from (select * from orders) as f join customers on f.customerid = customers.customerid`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)

			cfg, err := schema.ParseSchemaData(testSchema2)
			So(err, ShouldBeNil)

			ctx, err := NewParseCtx(stmt, cfg, dbOne)
			So(err, ShouldBeNil)

			So(AlgebrizeStatement(stmt, ctx), ShouldBeNil)
		})

		Convey("an aliased table can not use the original name", func() {

			sql := `select orders.customerid from orders o`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)

			cfg, err := schema.ParseSchemaData(testSchema2)
			So(err, ShouldBeNil)

			ctx, err := NewParseCtx(stmt, cfg, dbOne)
			So(err, ShouldBeNil)

			So(AlgebrizeStatement(stmt, ctx), ShouldNotBeNil)
		})

		Convey("a non-existent unqualified column reference should fail", func() {

			sql := `select DNE from orders o`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)

			cfg, err := schema.ParseSchemaData(testSchema2)
			So(err, ShouldBeNil)

			ctx, err := NewParseCtx(stmt, cfg, dbOne)
			So(err, ShouldBeNil)

			So(AlgebrizeStatement(stmt, ctx), ShouldNotBeNil)
		})

		Convey("a non-existent qualified column reference should fail", func() {

			sql := `select o.DNE from orders o`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)

			cfg, err := schema.ParseSchemaData(testSchema2)
			So(err, ShouldBeNil)

			ctx, err := NewParseCtx(stmt, cfg, dbOne)
			So(err, ShouldBeNil)

			So(AlgebrizeStatement(stmt, ctx), ShouldNotBeNil)
		})

		Convey("a non-existent qualified star expression reference should fail", func() {

			sql := `select dasd.* from (select * from foo) asd`

			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)

			cfg, err := schema.ParseSchemaData(testSchema2)
			So(err, ShouldBeNil)

			ctx, err := NewParseCtx(stmt, cfg, dbOne)
			So(err, ShouldBeNil)

			So(AlgebrizeStatement(stmt, ctx), ShouldNotBeNil)
		})

		Convey("nested subquery sources should rely on children sources to determine valid columns", func() {

			sql := "select sum_a_ok AS `sum_a_ok` FROM (  SELECT SUM(`bar`.`c`) AS `sum_a_ok`,  (COUNT(1) > 0) AS `havclause`,  1 AS `_Tableau_const_expr` FROM (select a, (select b from foo where b = 4) as x from bar) as bar GROUP BY 3) `t0` where  havclause"
			raw, err := sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok := raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)

			cfg, err := schema.ParseSchemaData(testSchema2)
			So(err, ShouldBeNil)

			_, err = NewParseCtx(stmt, cfg, dbOne)
			So(err, ShouldNotBeNil)

			sql = `select a,b from (select a from (select a,b from bar) y) x`
			raw, err = sqlparser.Parse(sql)
			So(err, ShouldBeNil)

			stmt, ok = raw.(*sqlparser.Select)
			So(ok, ShouldBeTrue)

			cfg, err = schema.ParseSchemaData(testSchema2)
			So(err, ShouldBeNil)

			_, err = NewParseCtx(stmt, cfg, dbOne)
			So(err, ShouldNotBeNil)

		})

	})
}
