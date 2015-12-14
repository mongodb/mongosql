package evaluator

import (
	"bytes"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type simpleVisitor struct{}

func (sv *simpleVisitor) Visit(e SQLExpr) (SQLExpr, error) {
	return walk(sv, e)
}

func TestPrintExpressionTree(t *testing.T) {

	Convey("Subject: PrintExpressionTree", t, func() {
		var buffer bytes.Buffer

		tests := [][]interface{}{
			{&SQLAndExpr{SQLBool(true), SQLBool(false)}, "SQLAndExpr:\n    SQLBool(true)\n    SQLBool(false)\n"},
			{&SQLAddExpr{SQLInt(10), SQLInt(20)}, "SQLAddExpr:\n    SQLInt(10)\n    SQLInt(20)\n"},
			{SQLBool(true), "SQLBool(true)\n"},
			// SQLDate
			// SQLDateTime
			{&SQLDivideExpr{SQLInt(10), SQLInt(20)}, "SQLDivideExpr:\n    SQLInt(10)\n    SQLInt(20)\n"},
			{&SQLEqualsExpr{SQLBool(true), SQLBool(false)}, "SQLEqualsExpr:\n    SQLBool(true)\n    SQLBool(false)\n"},
			{&SQLExistsExpr{}, "SQLExistsExpr({subquery})\n"},
			{&SQLFieldExpr{"db", "table"}, "SQLFieldExpr(db.table)\n"},
			{SQLFloat(10.42), "SQLFloat(10.42)\n"},
			{&SQLGreaterThanExpr{SQLBool(true), SQLBool(false)}, "SQLGreaterThanExpr:\n    SQLBool(true)\n    SQLBool(false)\n"},
			{&SQLGreaterThanOrEqualExpr{SQLBool(true), SQLBool(false)}, "SQLGreaterThanOrEqualExpr:\n    SQLBool(true)\n    SQLBool(false)\n"},
			{&SQLInExpr{SQLBool(true), SQLBool(false)}, "SQLInExpr:\n    SQLBool(true)\n    SQLBool(false)\n"},
			{SQLInt(10), "SQLInt(10)\n"},
			{&SQLLessThanExpr{SQLBool(true), SQLBool(false)}, "SQLLessThanExpr:\n    SQLBool(true)\n    SQLBool(false)\n"},
			{&SQLLessThanOrEqualExpr{SQLBool(true), SQLBool(false)}, "SQLLessThanOrEqualExpr:\n    SQLBool(true)\n    SQLBool(false)\n"},
			{&SQLLikeExpr{SQLBool(true), SQLBool(false)}, "SQLLikeExpr:\n    SQLBool(true)\n    SQLBool(false)\n"},
			{&SQLMultiplyExpr{SQLInt(10), SQLInt(20)}, "SQLMultiplyExpr:\n    SQLInt(10)\n    SQLInt(20)\n"},
			{&SQLNotExpr{SQLBool(true)}, "SQLNotExpr:\n    SQLBool(true)\n"},
			{&SQLNotEqualsExpr{SQLBool(true), SQLBool(false)}, "SQLNotEqualsExpr:\n    SQLBool(true)\n    SQLBool(false)\n"},
			{&SQLNullCmpExpr{SQLBool(true)}, "SQLNullCmpExpr:\n    SQLBool(true)\n"},
			{SQLNull, "SQLNull\n"},
			{&SQLOrExpr{SQLBool(true), SQLBool(false)}, "SQLOrExpr:\n    SQLBool(true)\n    SQLBool(false)\n"},
			{SQLString("funny"), "SQLString(funny)\n"},
			{&SQLSubqueryCmpExpr{left: SQLBool(true), value: &SQLSubqueryExpr{}}, "SQLSubqueryCmpExpr:\n    SQLBool(true)\n    SQLSubqueryExpr({subquery})\n"},
			{&SQLSubqueryExpr{}, "SQLSubqueryExpr({subquery})\n"},
			{&SQLSubtractExpr{SQLInt(10), SQLInt(20)}, "SQLSubtractExpr:\n    SQLInt(10)\n    SQLInt(20)\n"},
			// SQLTime
			// SQLTimestamp
			{&SQLUnaryMinusExpr{SQLInt(10)}, "SQLUnaryMinusExpr:\n    SQLInt(10)\n"},
			{&SQLUnaryPlusExpr{SQLInt(10)}, "SQLUnaryPlusExpr:\n    SQLInt(10)\n"},
			{&SQLUnaryTildeExpr{SQLInt(10)}, "SQLUnaryTildeExpr:\n    SQLInt(10)\n"},
			{SQLValues{SQLString("funny"), SQLInt(10)}, "SQLValues:\n    SQLString(funny)\n    SQLInt(10)\n"},
		}

		for _, t := range tests {
			Convey(fmt.Sprintf("Printing %T", t[0]), func() {
				PrintExpressionTree(t[0].(SQLExpr), &buffer)
				So(buffer.String(), ShouldEqual, t[1])
			})
		}
	})

}
