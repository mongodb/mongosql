package analyzer_test

import (
	"testing"

	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/internal/parsertest"

	"github.com/google/go-cmp/cmp"
)

func TestReferencedFieldNames_MatchExpr(t *testing.T) {
	testCases := []struct {
		input            string
		expectedNames    []string
		expectedComplete bool
	}{
		{`{ "a": 1 }`, []string{"a"}, true},
		{`{ "a": { "$eee": 7 } }`, []string{"a"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			expr := parsertest.ParseMatchExpr(tc.input)

			actual, actualComplete := analyzer.ReferencedFieldNames(expr)
			if !cmp.Equal(tc.expectedNames, actual) {
				t.Fatalf("predicate splits are not equal\n  %s", cmp.Diff(tc.expectedNames, actual))
			}
			if tc.expectedComplete != actualComplete {
				t.Fatalf("expected complete to be %v, but was %v", tc.expectedComplete, actualComplete)
			}
		})
	}
}

func TestReferencedFieldNames_Expr(t *testing.T) {
	testCases := []struct {
		input            string
		expectedNames    []string
		expectedComplete bool
	}{
		{`1`, nil, true},
		{`"$a"`, []string{"a"}, true},

		// TODO: bug in the bson libraries json parser
		// {`{ "$gt": ["$a", "$b"] }`, []string{"a", "b"}, true},
		// {`{ "$and": ["$a", "$b", "$a"] }`, []string{"a", "b"}, true},
		// {`{ "$and": ["$a", "$b", { "$gt": ["$a", "$c"] }] }`, []string{"a", "b", "c"}, true},
		// {`{ "$and": ["$a", { "$eee": ["$b", "$c"] }] }`, []string{"a"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			expr := parsertest.ParseExpr(tc.input)

			actual, actualComplete := analyzer.ReferencedFieldNames(expr)
			if !cmp.Equal(tc.expectedNames, actual) {
				t.Fatalf("predicate splits are not equal\n  %s", cmp.Diff(tc.expectedNames, actual))
			}
			if tc.expectedComplete != actualComplete {
				t.Fatalf("expected complete to be %v, but was %v", tc.expectedComplete, actualComplete)
			}
		})
	}
}

func TestReferencedFieldNames_Stage(t *testing.T) {
	testCases := []struct {
		input            string
		expectedNames    []string
		expectedComplete bool
	}{
		{`{ "$match": { "a": 1 } }`, []string{"a"}, true},
		{`{ "$match": { "a": { "$eee": 7 } } }`, []string{"a"}, false},

		// TODO: bug in the bson libraries json parser
		//{`{ "$project": { "_id": 0, "a": 1, "b": "$b", "c": { "$gt": [1, "$d"] } } }`, []string{"a", "b", "d"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			expr := parsertest.ParseStage(tc.input)

			actual, actualComplete := analyzer.ReferencedFieldNames(expr)
			if !cmp.Equal(tc.expectedNames, actual) {
				t.Fatalf("predicate splits are not equal\n  %s", cmp.Diff(tc.expectedNames, actual))
			}
			if tc.expectedComplete != actualComplete {
				t.Fatalf("expected complete to be %v, but was %v", tc.expectedComplete, actualComplete)
			}
		})
	}
}

func TestReferencedFieldNames_Pipeline(t *testing.T) {
	testCases := []struct {
		input            string
		expectedNames    []string
		expectedComplete bool
	}{
		{
			`[{ "$match": { "a": 1 } }]`,
			[]string{"a"},
			false,
		},
		{
			`[
				{ "$match": { "a": 1 } },
				{ "$project": { "e": "$b" } },
				{ "$match": { "e": 1} }
			]`,
			[]string{"a", "b"},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			expr := parsertest.ParsePipeline(tc.input)

			actual, actualComplete := analyzer.ReferencedFieldNames(expr)
			if !cmp.Equal(tc.expectedNames, actual) {
				t.Fatalf("predicate splits are not equal\n  %s", cmp.Diff(tc.expectedNames, actual))
			}
			if tc.expectedComplete != actualComplete {
				t.Fatalf("expected complete to be %v, but was %v", tc.expectedComplete, actualComplete)
			}
		})
	}
}
