package analyzer_test

import (
	"testing"

	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/internal/parsertest"

	"github.com/google/go-cmp/cmp"
)

func TestReferencedFieldRoots_MatchExpr(t *testing.T) {
	testCases := []struct {
		input            string
		expectedNames    []string
		expectedComplete bool
	}{
		{`{ "a": 1 }`, []string{"a"}, true},
		{`{ "a": { "$eee": 7 } }`, []string{"a"}, false},
		{`{ "a.1.2.b": { "$eee": 7 } }`, []string{"a"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			expr := parsertest.ParseMatchExpr(tc.input)

			actual, actualComplete := analyzer.ReferencedFieldRoots(expr)
			if !cmp.Equal(tc.expectedNames, actual) {
				t.Fatalf("predicate splits are not equal\n  %s", cmp.Diff(tc.expectedNames, actual))
			}
			if tc.expectedComplete != actualComplete {
				t.Fatalf("expected complete to be %v, but was %v", tc.expectedComplete, actualComplete)
			}
		})
	}
}

func TestReferencedFieldRoots_Expr(t *testing.T) {
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
			actual, actualComplete := analyzer.ReferencedFieldRoots(expr)
			if !cmp.Equal(tc.expectedNames, actual) {
				t.Fatalf("predicate splits are not equal\n  %s", cmp.Diff(tc.expectedNames, actual))
			}
			if tc.expectedComplete != actualComplete {
				t.Fatalf("expected complete to be %v, but was %v", tc.expectedComplete, actualComplete)
			}
		})
	}
}

func TestReferencedFieldRoots_Stage(t *testing.T) {
	testCases := []struct {
		input            string
		expectedNames    []string
		expectedComplete bool
	}{
		{`{ "$match": { "a": 1 } }`, []string{"a"}, true},
		{`{ "$match": { "a": { "$eee": 7 } } }`, []string{"a"}, false},
		{`{ "$project": { "a.b": 1 }}`, []string{"a"}, true},
		{`{"$match": {
			"$expr":
				{"$not": {
					"$eq": [
						{"$arrayElemAt": [
								{"$map": {"input": "$__subquery_alter_1_[2]","as": "this","in": {"$ifNull": ["$$this.4",null]}}},{"$numberInt":"0"}
						]}, null]
					}
				}
			}
		}`, []string{"__subquery_alter_1_[2]"}, true},

		// TODO: bug in the bson libraries json parser
		//{`{ "$project": { "_id": 0, "a": 1, "b": "$b", "c": { "$gt": [1, "$d"] } } }`, []string{"a", "b", "d"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			expr := parsertest.ParseStage(tc.input)

			actual, actualComplete := analyzer.ReferencedFieldRoots(expr)
			if !cmp.Equal(tc.expectedNames, actual) {
				t.Fatalf("predicate splits are not equal\n  %s", cmp.Diff(tc.expectedNames, actual))
			}
			if tc.expectedComplete != actualComplete {
				t.Fatalf("expected complete to be %v, but was %v", tc.expectedComplete, actualComplete)
			}
		})
	}
}

func TestReferencedFieldRoots_Pipeline(t *testing.T) {
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
		{
			`[
		                 { "$match": { "a": 1 } },
		                 { "$project": { "e": { "$add": ["$b", "$c"] } } },
		                 { "$match": { "e": 1} }
		             ]`,
			[]string{"a", "b", "c"},
			true,
		},
		{
			`[
			             { "$match": { "a": 1 } },
			             { "$addFields": { "e": { "$add": ["$b", "$c"] } } },
			             { "$match": { "e": 1} }
			         ]`,
			[]string{"a", "b", "c"},
			false,
		},
		{
			`[
						 { "$project": { "a": 1, "e": 1}},
		                 { "$match": { "a": 1 } },
		                 { "$match": { "e.y": 1} },
		                 { "$match": { "e.y.z": 1} }
		             ]`,
			[]string{"a", "e"},
			true,
		},
		{
			`[
                 { "$match": { "a": 1 } },
                 { "$match": { "e.y": 1} },
                 { "$match": { "e.y.z": 1} }
             ]`,
			[]string{"a", "e"},
			false,
		},
		{
			`[
			             { "$match": { "a": 1 } },
			             { "$project": { "e":  "$d" } },
			             { "$match": { "e.y": 1} },
			             { "$match": { "e.y.z": 1} }
			         ]`,
			[]string{"a", "d"},
			true,
		},
		{
			`[
			         { "$replaceRoot": { "newRoot": "$a" } },
			         { "$match": { "e.y": 1} },
			         { "$match": { "e.y.z": 1} }
		     ]`,
			[]string{"a"},
			true,
		},
		{
			`[
			         { "$replaceRoot": { "newRoot": {"e" : "$a"} } },
			         { "$match": { "e.y": 1} },
			         { "$match": { "e.y.z": 1} }
		     ]`,
			[]string{"a"},
			true,
		},
		{
			`[
			         { "$replaceRoot": { "newRoot": {"e" : "$a"} } },
			         { "$match": { "e.y": 1} },
			         { "$replaceRoot": { "newRoot": {"e" : "$e"} } },
			         { "$match": { "e.y.z": 1} }
		     ]`,
			[]string{"a"},
			true,
		},
		{
			`[
			         { "$replaceRoot": { "newRoot": {"e" : "$a.1.2.3.b.1"} } },
			         { "$match": { "e.y": 1} },
			         { "$replaceRoot": { "newRoot": {"e" : "$e"} } },
			         { "$match": { "e.y.z": 1} }
		     ]`,
			[]string{"a"},
			true,
		},
		{
			`[
					{ "$lookup":
						{ 
						  "from": "foo",
						  "localField": "a.b.c",
						  "foreignField": "c",
						  "as": "hello"
						}
					}
		     ]`,
			[]string{"a"},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			expr := parsertest.ParsePipeline(tc.input)

			actual, actualComplete := analyzer.ReferencedFieldRoots(expr)
			if !cmp.Equal(tc.expectedNames, actual) {
				t.Fatalf("root name sets are not equal\n  %s", cmp.Diff(tc.expectedNames, actual))
			}
			if tc.expectedComplete != actualComplete {
				t.Fatalf("expected complete to be %v, but was %v", tc.expectedComplete, actualComplete)
			}
		})
	}
}
