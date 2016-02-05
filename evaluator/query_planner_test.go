package evaluator

import (
	"github.com/deafgoat/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPlanFromExpr(t *testing.T) {

	ctx := &ExecutionCtx{
		Schema: cfgOne,
		Db:     dbOne,
	}

	Convey("With a given table expr...", t, func() {

		Convey("planning the from expression with no table should return an error", func() {

			tables := []sqlparser.TableExpr{}

			opr, err := planFromExpr(ctx, tables, nil)
			So(err, ShouldNotBeNil)
			So(opr, ShouldBeNil)

		})

		Convey("planning the from expression with one table should return a table scan operator", func() {

			tables := []sqlparser.TableExpr{
				&sqlparser.AliasedTableExpr{
					Expr: &sqlparser.TableName{
						Name:      []byte(tableOneName),
						Qualifier: []byte(dbOne),
					},
				},
			}

			opr, err := planFromExpr(ctx, tables, nil)
			So(err, ShouldBeNil)
			ts, ok := opr.(*TableScan)
			So(ok, ShouldBeTrue)
			So(ts.tableName, ShouldEqual, tableOneName)

		})

		Convey("planning the from expression with two tables should return a cross join operator", func() {

			tables := []sqlparser.TableExpr{
				&sqlparser.AliasedTableExpr{
					Expr: &sqlparser.TableName{
						Name:      []byte(tableOneName),
						Qualifier: []byte(dbOne),
					},
				},
				&sqlparser.AliasedTableExpr{
					Expr: &sqlparser.TableName{
						Name:      []byte(tableTwoName),
						Qualifier: []byte(dbOne),
					},
				},
			}

			opr, err := planFromExpr(ctx, tables, nil)
			So(err, ShouldBeNil)
			join, ok := opr.(*Join)
			So(ok, ShouldBeTrue)
			So(join.kind, ShouldEqual, CrossJoin)

			left, ok := join.left.(*TableScan)
			So(ok, ShouldBeTrue)
			So(left.tableName, ShouldEqual, tableOneName)

			right, ok := join.right.(*TableScan)
			So(ok, ShouldBeTrue)
			So(right.tableName, ShouldEqual, tableTwoName)

		})

		Convey("planning the from expression with more than two tables should return a left-leaning cross join operator", func() {

			tables := []sqlparser.TableExpr{
				&sqlparser.AliasedTableExpr{
					Expr: &sqlparser.TableName{
						Name:      []byte(tableOneName),
						Qualifier: []byte(dbOne),
					},
				},
				&sqlparser.AliasedTableExpr{
					Expr: &sqlparser.TableName{
						Name:      []byte(tableTwoName),
						Qualifier: []byte(dbOne),
					},
				},
				&sqlparser.AliasedTableExpr{
					Expr: &sqlparser.TableName{
						Name:      []byte(tableThreeName),
						Qualifier: []byte(dbOne),
					},
				},
			}

			opr, err := planFromExpr(ctx, tables, nil)
			So(err, ShouldBeNil)

			join, ok := opr.(*Join)
			So(ok, ShouldBeTrue)
			So(join.kind, ShouldEqual, CrossJoin)

			ts, ok := join.right.(*TableScan)
			So(ok, ShouldBeTrue)
			So(ts.tableName, ShouldEqual, tableThreeName)

			join, ok = join.left.(*Join)
			So(ok, ShouldBeTrue)

			left, ok := join.left.(*TableScan)
			So(ok, ShouldBeTrue)
			So(left.tableName, ShouldEqual, tableOneName)

			right, ok := join.right.(*TableScan)
			So(ok, ShouldBeTrue)
			So(right.tableName, ShouldEqual, tableTwoName)

		})
	})
}
