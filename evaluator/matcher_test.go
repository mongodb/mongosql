package evaluator

import (
	"fmt"
	"github.com/10gen/sqlproxy/config"
	"github.com/erh/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/json"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func getMatcherFromSQL(sql string) (Matcher, error) {
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

		return BuildMatcher(selectStatement.Where.Expr)
	}
	return nil, fmt.Errorf("statement doesn't look like a 'SELECT'")
}

func TestBoolMatcher(t *testing.T) {
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
				matcher := BoolMatcher{t}
				match, err := matcher.Matches(evalCtx)
				So(err, ShouldBeNil)
				So(match, ShouldBeTrue)
			}
			for _, t := range shouldBeFalse {
				matcher := BoolMatcher{t}
				match, err := matcher.Matches(evalCtx)
				So(err, ShouldBeNil)
				So(match, ShouldBeFalse)
			}
		})
	})
}

func TestMatcherBuilder(t *testing.T) {
	Convey("Simple WHERE with explicit table names", t, func() {
		matcher, err := getMatcherFromSQL("select * from bar where bar.a = 'eliot'")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &Equals{SQLField{"bar", "a"}, SQLString("eliot")})
	})
	Convey("Simple WHERE with implicit table names", t, func() {
		matcher, err := getMatcherFromSQL("select * from bar where a = 'eliot'")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &Equals{SQLField{"bar", "a"}, SQLString("eliot")})
	})
	Convey("WHERE with complex nested matching clauses", t, func() {
		matcher, err := getMatcherFromSQL("select * from bar where NOT((a = 'eliot') AND (b>1 OR a<'blah'))")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &Not{
			&And{
				[]Matcher{
					&Equals{SQLField{"bar", "a"}, SQLString("eliot")},
					&Or{
						[]Matcher{
							&GreaterThan{SQLField{"bar", "b"}, SQLInt(1)},
							&LessThan{SQLField{"bar", "a"}, SQLString("blah")},
						},
					},
				},
			},
		})
	})
	Convey("WHERE with complex nested matching clauses", t, func() {
		matcher, err := getMatcherFromSQL("select * from bar where NOT((a = 'eliot') AND (b>13 OR a<'blah'))")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &Not{
			&And{
				[]Matcher{
					&Equals{SQLField{"bar", "a"}, SQLString("eliot")},
					&Or{
						[]Matcher{
							&GreaterThan{SQLField{"bar", "b"}, SQLInt(13)},
							&LessThan{SQLField{"bar", "a"}, SQLString("blah")},
						},
					},
				},
			},
		})
	})
}

func TestBasicMatching(t *testing.T) {
	Convey("With a matcher checking for: field b = 'xyz'", t, func() {
		tree := Equals{SQLString("xyz"), &SQLField{"bar", "b"}}
		Convey("using the matcher on a row whose value matches should return true", func() {
			tree := Equals{SQLString("xyz"), &SQLField{"bar", "b"}}
			evalCtx := &EvalCtx{[]Row{
				{Data: []TableRow{{"bar", Values{{"a", "a", 123}, {"b", "b", "xyz"}, {"c", "c", nil}}, nil}}}}, nil}
			m, err := tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)
		})

		Convey("using the matcher on a row whose values do not match should return false", func() {
			evalCtx := &EvalCtx{[]Row{
				{Data: []TableRow{{"bar", Values{{"a", "a", 123}, {"b", "b", "WRONG"}, {"c", "c", nil}}, nil}}}}, nil}
			m, err := tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)
		})
	})
}

func TestComparisonMatchers(t *testing.T) {
	type compareTest struct {
		less, greater SQLValue
	}

	tests := []compareTest{
		{SQLFloat(1000.0), SQLFloat(5000.0)},
		{SQLString("aaa"), SQLString("bbb")},
		{SQLField{"bar", "a"}, SQLField{"bar", "y"}},
	}

	Convey("Equality matcher should return true/false when numerics are equal/unequal", t, func() {
		var tree Matcher
		evalCtx := &EvalCtx{[]Row{
			{Data: []TableRow{{"bar", Values{{"a", "a", 123}, {"y", "y", 456}, {"c", "c", nil}}, nil}}}}, nil}
		for _, data := range tests {
			tree = &Equals{data.less, data.less}
			m, err := tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			tree = &Equals{data.less, data.greater}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			tree = &NotEquals{data.less, data.greater}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			tree = &NotEquals{data.less, data.less}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			tree = &Not{&Equals{data.less, data.less}}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			tree = &Not{&Equals{data.less, data.greater}}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			/* GT */
			tree = &GreaterThan{data.less, data.greater}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			tree = &GreaterThan{data.greater, data.less}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			tree = &GreaterThan{data.less, data.less}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			/* GTE */
			tree = &GreaterThanOrEqual{data.less, data.greater}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			tree = &GreaterThanOrEqual{data.greater, data.less}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			tree = &GreaterThanOrEqual{data.less, data.less}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			/* LT */
			tree = &LessThan{data.less, data.greater}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			tree = &LessThan{data.greater, data.less}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			tree = &LessThan{data.less, data.less}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			/* LTE */
			tree = &LessThanOrEqual{data.less, data.greater}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeTrue)

			tree = &LessThanOrEqual{data.greater, data.less}
			m, err = tree.Matches(evalCtx)
			So(err, ShouldBeNil)
			So(m, ShouldBeFalse)

			tree = &LessThanOrEqual{data.less, data.less}
			m, err = tree.Matches(evalCtx)
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
