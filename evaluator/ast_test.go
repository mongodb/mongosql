package evaluator

import (
	"encoding/json"
	"fmt"
	"github.com/10gen/sqlproxy/config"
	"github.com/erh/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func getSQLExprFromSQL(sql string) (SQLExpr, error) {
	// Parse the statement, algebrize it, extract the WHERE clause and build a matcher from it.
	raw, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, err
	}
	if selectStatement, ok := raw.(*sqlparser.Select); ok {
		cfg, err := config.ParseConfigData(testConfig3)
		So(err, ShouldBeNil)
		parseCtx, err := NewParseCtx(selectStatement, cfg, dbOne)
		if err != nil {
			return nil, err
		}

		parseCtx.Database = dbOne

		err = AlgebrizeStatement(selectStatement, parseCtx)
		if err != nil {
			return nil, err
		}

		return NewSQLExpr(selectStatement.Where.Expr)
	}
	return nil, fmt.Errorf("statement doesn't look like a 'SELECT'")
}

func TestMatchesWithValues(t *testing.T) {
	Convey("When evaluating a single value as a match", t, func() {
		evalCtx := &EvalCtx{[]Row{
			{Data: []TableRow{
				{"bar", Values{
					{"a", "a", 123},
					{"b", "b", "xyz"},
					{"c", "c", nil},
				}, nil}}}}, nil}
		Convey("It should evaluate to true iff it's non-zero or parseable as a non-zero number", func() {
			shouldBeTrue := []SQLValue{
				SQLInt(124),
				SQLFloat(1235),
				SQLString("512"),
			}

			shouldBeFalse := []SQLValue{
				SQLInt(0),
				SQLFloat(0),
				SQLString("000"),
				SQLString("skdjbkjb"),
				SQLString(""),
			}
			for _, t := range shouldBeTrue {
				match, err := Matches(t, evalCtx)
				So(err, ShouldBeNil)
				So(match, ShouldBeTrue)
			}
			for _, t := range shouldBeFalse {
				match, err := Matches(t, evalCtx)
				So(err, ShouldBeNil)
				So(match, ShouldBeFalse)
			}
		})
	})
}

func TestBasicBooleanExpressions(t *testing.T) {
	Convey("With a matcher checking for: field b = 'xyz'", t, func() {
		tree := &SQLEqualsExpr{SQLString("xyz"), &SQLFieldExpr{"bar", "b"}}
		Convey("using the matcher on a row whose value matches should return true", func() {
			evalCtx := &EvalCtx{[]Row{
				{Data: []TableRow{{"bar", Values{{"a", "a", 123}, {"b", "b", "xyz"}, {"c", "c", nil}}, nil}}}}, nil}
			m, err := Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)
		})

		Convey("using the matcher on a row whose values do not match should return false", func() {
			evalCtx := &EvalCtx{[]Row{
				{Data: []TableRow{{"bar", Values{{"a", "a", 123}, {"b", "b", "WRONG"}, {"c", "c", nil}}, nil}}}}, nil}
			m, err := Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)
		})
	})
}

func TestComparisonExpressions(t *testing.T) {
	type compareTest struct {
		less, greater SQLExpr
	}

	tests := []compareTest{
		{SQLFloat(1000.0), SQLFloat(5000.0)},
		{SQLString("aaa"), SQLString("bbb")},
		{SQLFieldExpr{"bar", "a"}, SQLFieldExpr{"bar", "y"}},
	}

	Convey("Equality matcher should return true/false when numerics are equal/unequal", t, func() {
		var tree SQLExpr
		evalCtx := &EvalCtx{[]Row{
			{Data: []TableRow{{"bar", Values{{"a", "a", 123}, {"y", "y", 456}, {"c", "c", nil}}, nil}}}}, nil}
		for _, data := range tests {
			tree = &SQLEqualsExpr{data.less, data.less}
			m, err := Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			tree = &SQLEqualsExpr{data.less, data.greater}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			tree = &SQLNotEqualsExpr{data.less, data.greater}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			tree = &SQLNotEqualsExpr{data.less, data.less}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			tree = &SQLNotExpr{&SQLEqualsExpr{data.less, data.less}}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			tree = &SQLNotExpr{&SQLEqualsExpr{data.less, data.greater}}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			/* GT */
			tree = &SQLGreaterThanExpr{data.less, data.greater}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			tree = &SQLGreaterThanExpr{data.greater, data.less}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			tree = &SQLGreaterThanExpr{data.less, data.less}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			/* GTE */
			tree = &SQLGreaterThanOrEqualExpr{data.less, data.greater}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			tree = &SQLGreaterThanOrEqualExpr{data.greater, data.less}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			tree = &SQLGreaterThanOrEqualExpr{data.less, data.less}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			/* LT */
			tree = &SQLLessThanExpr{data.less, data.greater}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			tree = &SQLLessThanExpr{data.greater, data.less}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			tree = &SQLLessThanExpr{data.less, data.less}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			/* LTE */
			tree = &SQLLessThanOrEqualExpr{data.less, data.greater}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			tree = &SQLLessThanOrEqualExpr{data.greater, data.less}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			tree = &SQLLessThanOrEqualExpr{data.less, data.less}
			m, err = Matches(tree, evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)
		}
	})
}

func debugJson(data interface{}) {
	So(data, ShouldNotBeNil)
	_, err := json.Marshal(data)
	So(err, ShouldBeNil)
}
