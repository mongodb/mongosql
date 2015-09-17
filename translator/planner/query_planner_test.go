package planner

import (
	"github.com/erh/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	// Table expressions to use as arguments for the
	// planFromExpr function
	planFromExprTExprs = [][]sqlparser.TableExpr{

		// no table expression
		[]sqlparser.TableExpr{},

		// one table expression
		[]sqlparser.TableExpr{
			&sqlparser.AliasedTableExpr{
				Expr: &sqlparser.TableName{
					Name: []byte(tableOneName),
				},
			},
		},

		// ...
		[]sqlparser.TableExpr{
			&sqlparser.AliasedTableExpr{
				Expr: &sqlparser.TableName{
					Name: []byte(tableOneName),
				},
			},
			&sqlparser.AliasedTableExpr{
				Expr: &sqlparser.TableName{
					Name: []byte(tableTwoName),
				},
			},
		},

		// ...
		[]sqlparser.TableExpr{
			&sqlparser.AliasedTableExpr{
				Expr: &sqlparser.TableName{
					Name: []byte(tableOneName),
				},
			},
			&sqlparser.AliasedTableExpr{
				Expr: &sqlparser.TableName{
					Name: []byte(tableTwoName),
				},
			},
			&sqlparser.AliasedTableExpr{
				Expr: &sqlparser.TableName{
					Name: []byte(tableThreeName),
				},
			},
		},
	}

	ctx = &ExecutionCtx{}
)

func TestPlanFromExpr(t *testing.T) {

	Convey("With a given table expr...", t, func() {

		Convey("planning the from expression with no table should return an error", func() {

			opr, err := planFromExpr(ctx, planFromExprTExprs[0], nil)
			So(err, ShouldNotBeNil)
			So(opr, ShouldBeNil)

		})

		Convey("planning the from expression with one table should return a table scan operator", func() {

			opr, err := planFromExpr(ctx, planFromExprTExprs[1], nil)
			So(err, ShouldBeNil)
			ts, ok := opr.(*TableScan)
			So(ok, ShouldBeTrue)
			So(ts.tableName, ShouldEqual, tableOneName)

		})

		Convey("planning the from expression with two tables should return a cross join operator", func() {

			opr, err := planFromExpr(ctx, planFromExprTExprs[2], nil)
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

			opr, err := planFromExpr(ctx, planFromExprTExprs[3], nil)
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
