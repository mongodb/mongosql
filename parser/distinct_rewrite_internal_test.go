package parser

import (
	"testing"

	"github.com/10gen/sqlproxy/internal/util/option"
)

func TestGetInnerAndOuterAlias(t *testing.T) {
	tcases := []struct {
		desc          string
		expr          *NonStarExpr
		expectedInner string
		expectedOuter string
	}{
		{
			desc: "already named",
			expr: &NonStarExpr{
				Expr: &ColName{
					Database:  option.NoneString(),
					Qualifier: option.NoneString(),
					Name:      "foo",
				},
				As: option.SomeString("bar"),
			},
			expectedInner: "bar",
			expectedOuter: "bar",
		},
		{
			desc: "already named with database and column name",
			expr: &NonStarExpr{
				Expr: &ColName{
					Database:  option.SomeString("a"),
					Qualifier: option.SomeString("b"),
					Name:      "foo",
				},
				As: option.SomeString("bar"),
			},
			expectedInner: "bar",
			expectedOuter: "bar",
		},
		{
			desc: "not named ColName Expr",
			expr: &NonStarExpr{
				Expr: &ColName{
					Database:  option.NoneString(),
					Qualifier: option.NoneString(),
					Name:      "foo",
				},
				As: option.NoneString(),
			},
			expectedInner: "$___mongosqld_as_0",
			expectedOuter: "foo",
		},
		{
			desc: "not named ColName Expr with database and column name",
			expr: &NonStarExpr{
				Expr: &ColName{
					Database:  option.SomeString("a"),
					Qualifier: option.SomeString("a"),
					Name:      "foo",
				},
				As: option.NoneString(),
			},
			expectedInner: "$___mongosqld_as_0",
			expectedOuter: "a.a.foo",
		},
		{
			desc: "not named ColName Expr",
			expr: &NonStarExpr{
				Expr: &AndExpr{
					Left: &ColName{
						Database:  option.NoneString(),
						Qualifier: option.NoneString(),
						Name:      "foo",
					},
					Right: &ColName{
						Database:  option.NoneString(),
						Qualifier: option.NoneString(),
						Name:      "bar",
					},
				},
				As: option.NoneString(),
			},
			expectedInner: "$___mongosqld_as_0",
			expectedOuter: "foo and bar",
		},
	}

	for _, tcase := range tcases {
		d := NewDistinctRewriter()
		innerAs, outerAs := d.getInnerAndOuterAlias(tcase.expr)
		if innerAs != tcase.expectedInner {
			t.Errorf("for test case '%s'\n  inner alias output: '%s'\n  "+
				"does not match expected inner alias: '%s'",
				tcase.desc, innerAs, tcase.expectedInner)
		}
		if outerAs != tcase.expectedOuter {
			t.Errorf("for test case '%s'\n  outer alias output: '%s'\n  "+
				"does not match expected outer alias: '%s'",
				tcase.desc, outerAs, tcase.expectedOuter)
		}
	}
}
