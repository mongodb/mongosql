package planner

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/translator/algebrizer"
	"github.com/mongodb/mongo-tools/common/bsonutil"
	"github.com/mongodb/mongo-tools/common/json"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func getMatcherFromSQL(sql string) (Matcher, error) {
	// Parse the statement, algebrize it, extract the WHERE clause and build a matcher from it.
	raw, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, err
	}
	if selectStatement, ok := raw.(*sqlparser.Select); ok {
		parseCtx, err := algebrizer.NewParseCtx(selectStatement)
		if err != nil {
			return nil, err
		}

		err = algebrizer.AlgebrizeStatement(selectStatement, parseCtx)
		if err != nil {
			return nil, err
		}

		return BuildMatcher(selectStatement.Where.Expr)
	}
	return nil, fmt.Errorf("statement doesn't look like a 'SELECT'")
}

func TestMatcherBuilder(t *testing.T) {
	Convey("Simple WHERE with explicit table names", t, func() {
		matcher, err := getMatcherFromSQL("select * from foo where foo.firstname = 'eliot'")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &Equals{SQLField{"foo", "firstname"}, SQLString("eliot")})
	})
	Convey("Simple WHERE with implicit table names", t, func() {
		matcher, err := getMatcherFromSQL("select * from foo where firstname = 'eliot'")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &Equals{SQLField{"foo", "firstname"}, SQLString("eliot")})
	})
	Convey("WHERE with complex nested matching clauses", t, func() {
		matcher, err := getMatcherFromSQL("select * from foo where NOT((firstname = 'eliot') AND (x>1 OR y<'blah'))")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &Not{
			&And{
				[]Matcher{
					&Equals{SQLField{"foo", "firstname"}, SQLString("eliot")},
					&Or{
						[]Matcher{
							&GreaterThan{SQLField{"foo", "x"}, SQLNumeric(1.0)},
							&LessThan{SQLField{"foo", "y"}, SQLString("blah")},
						},
					},
				},
			},
		})
	})
	Convey("WHERE with complex nested matching clauses", t, func() {
		matcher, err := getMatcherFromSQL("select * from foo where NOT((firstname = 'eliot') AND (x>1 OR y<'blah'))")
		So(err, ShouldBeNil)
		So(matcher, ShouldResemble, &Not{
			&And{
				[]Matcher{
					&Equals{SQLField{"foo", "firstname"}, SQLString("eliot")},
					&Or{
						[]Matcher{
							&GreaterThan{SQLField{"foo", "x"}, SQLNumeric(1.0)},
							&LessThan{SQLField{"foo", "y"}, SQLString("blah")},
						},
					},
				},
			},
		})
	})
}

func TestBasicMatching(t *testing.T) {
	Convey("With a matcher checking for: field b = 'xyz'", t, func() {
		tree := Equals{SQLString("xyz"), &SQLField{"foo", "b"}}
		Convey("using the matcher on a row whose value matches should return true", func() {
			tree := Equals{SQLString("xyz"), &SQLField{"foo", "b"}}
			matchCtx := &MatchCtx{[]*Row{
				{Data: []TableRow{{"foo", bson.D{{"a", 123}, {"b", "xyz"}, {"c", nil}}}}},
			}}
			So(tree.Matches(matchCtx), ShouldBeTrue)
		})

		Convey("using the matcher on a row whose values do not match should return false", func() {
			matchCtx := &MatchCtx{[]*Row{
				{Data: []TableRow{{"foo", bson.D{{"a", 123}, {"b", "WRONG"}, {"c", nil}}}}},
			}}
			So(tree.Matches(matchCtx), ShouldBeFalse)
		})
	})
}

func TestComparisonMatchers(t *testing.T) {
	type compareTest struct {
		less, greater SQLValue
	}

	tests := []compareTest{
		{SQLNumeric(1000.0), SQLNumeric(5000.0)},
		{SQLString("aaa"), SQLString("bbb")},
		{SQLField{"foo", "x"}, SQLField{"foo", "y"}},
	}

	Convey("Equality matcher should return true/false when numerics are equal/unequal", t, func() {
		var tree Matcher
		matchCtx := &MatchCtx{[]*Row{
			{Data: []TableRow{{"foo", bson.D{{"x", 123}, {"y", 456}, {"c", nil}}}}},
		}}
		for _, data := range tests {
			tree = &Equals{data.less, data.less}
			So(tree.Matches(matchCtx), ShouldBeTrue)

			tree = &Equals{data.less, data.greater}
			So(tree.Matches(matchCtx), ShouldBeFalse)

			tree = &NotEquals{data.less, data.greater}
			So(tree.Matches(matchCtx), ShouldBeTrue)

			tree = &NotEquals{data.less, data.less}
			So(tree.Matches(matchCtx), ShouldBeFalse)

			tree = &Not{&Equals{data.less, data.less}}
			So(tree.Matches(matchCtx), ShouldBeFalse)

			tree = &Not{&Equals{data.less, data.greater}}
			So(tree.Matches(matchCtx), ShouldBeTrue)

			/* GT */
			tree = &GreaterThan{data.less, data.greater}
			So(tree.Matches(matchCtx), ShouldBeFalse)

			tree = &GreaterThan{data.greater, data.less}
			So(tree.Matches(matchCtx), ShouldBeTrue)

			tree = &GreaterThan{data.less, data.less}
			So(tree.Matches(matchCtx), ShouldBeFalse)

			/* GTE */
			tree = &GreaterThanOrEqual{data.less, data.greater}
			So(tree.Matches(matchCtx), ShouldBeFalse)

			tree = &GreaterThanOrEqual{data.greater, data.less}
			So(tree.Matches(matchCtx), ShouldBeTrue)

			tree = &GreaterThanOrEqual{data.less, data.less}
			So(tree.Matches(matchCtx), ShouldBeTrue)

			/* LT */
			tree = &LessThan{data.less, data.greater}
			So(tree.Matches(matchCtx), ShouldBeTrue)

			tree = &LessThan{data.greater, data.less}
			So(tree.Matches(matchCtx), ShouldBeFalse)

			tree = &LessThan{data.less, data.less}
			So(tree.Matches(matchCtx), ShouldBeFalse)

			/* LTE */
			tree = &LessThanOrEqual{data.less, data.greater}
			So(tree.Matches(matchCtx), ShouldBeTrue)

			tree = &LessThanOrEqual{data.greater, data.less}
			So(tree.Matches(matchCtx), ShouldBeFalse)

			tree = &LessThanOrEqual{data.less, data.less}
			So(tree.Matches(matchCtx), ShouldBeTrue)
		}
	})
}

func debugJson(data interface{}) {
	if data == nil {
		fmt.Println("<nil>")
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println("error marshaling json: ", err)
	}

	fmt.Println(string(jsonBytes))
}

func TestMatcherTransformation(t *testing.T) {
	Convey("Equality matcher should return true/false when numerics are equal/unequal", t, func() {
		matcher, err := getMatcherFromSQL("select * from foo where foo.firstname = 'eliot'")
		So(err, ShouldBeNil)
		transformed, err := matcher.Transform()
		So(err, ShouldBeNil)
		debugJson(bsonutil.MarshalD(*transformed))

		matcher, err = getMatcherFromSQL("select * from foo where foo.firstname = 'eliot' and x=5")
		So(err, ShouldBeNil)
		transformed, err = matcher.Transform()
		debugJson(bsonutil.MarshalD(*transformed))
	})
}
