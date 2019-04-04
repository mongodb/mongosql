package analyzer_test

import (
	"testing"

	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/internal/parsertest"

	"github.com/google/go-cmp/cmp"
)

func TestDefinedFields(t *testing.T) {
	testCases := []struct {
		input         string
		expectedNames []string
	}{
		{`{"$addFields":	{ "b": "$b", "c" : { "$add": ["$c", 1] } } }`,
			[]string{"b", "c"}},
		{`{"$bucket": {
				"groupBy": "$price",
				"boundaries": [ 0, 200, 400 ],
				"default": "Other"
			}
		}`,
			[]string{"_id", "count"}},
		{`{"$bucket": {
				"groupBy": "$price",
				"boundaries": [ 0, 200, 400 ],
				"default": "Other",
				"output": {
				"count": { "$sum": 1 },
				"titles" : { "$push": "$title" }
				}
			}
		}`,
			[]string{"_id", "count", "titles"}},
		{`{"$bucketAuto": {
					"groupBy": "$year",
				 	"buckets": 3,
					"output": {
						"count": { "$sum": 1 },
						"years": { "$push": "$year" }
					}
			}
		}`,
			[]string{"_id", "count", "years"}},
		{`{"$bucketAuto": {
					"groupBy": "$year",
				 	"buckets": 3
			}
		}`,
			[]string{"_id", "count"}},
		{`{"$collStats": {}}`, []string{"ns", "shard", "host", "localTime", "latencyStats", "storageStats", "count"}},
		{`{"$count": "foo"}`, []string{"foo"}},
		{`{"$facet": 

			{
				"categorizedByTags": [
						{ "$unwind": "$tags" },
						{ "$sortByCount": "$tags" }
				]
			}
		}`, []string{"categorizedByTags"}},
		{`{"$group" : {
			"_id" : { "month": { "$month": "$date" }, "day": { "$dayOfMonth": "$date" }, "year": { "$year": "$date" } },
			"totalPrice": { "$sum": { "$multiply": [ "$price", "$quantity" ] } },
			"averageQuantity": { "$avg": "$quantity" },
			"count": { "$sum": 1 }
        }}`, []string{"_id", "totalPrice", "averageQuantity", "count"}},
		{`{"$lookup" : {"from": "bar", "localField": "foo", "foreignField": "foo", "as": "bazz"}}`,
			[]string{"bazz"}},
		{`{"$project":	{ "a": 1, "b": "$b", "c" : { "$add": ["$c", 1] } } }`,
			[]string{"a", "b", "c"}},
		// Note that even though we are defining a.b, we want the algorithm to return a.
		// We will always consider the top most part of a field definition. The upshot of this
		// is that if we have a project that defined a, followed by one that defines a.b, we will
		// still consider that second definition of a.b to be necessary for any references to a,
		// this is important to correctness.
		{`{"$project":	{ "a.b": 1, "b": "$b", "c" : { "$add": ["$c", 1] } } }`,
			[]string{"a", "b", "c"}},
		{`{"$replaceRoot": {"newRoot": "$a"}}`, []string{}},
		{`{"$replaceRoot":	{"newRoot" : { "a": 1, "b": "$b", "c" : { "$add": ["$c", 1] } } } }`,
			[]string{"a", "b", "c"}},
		{`{"$unwind": "$x"}`, []string{"x"}},
		{`{"$unwind": {"path": "$x", "includeArrayIndex": "x_idx"}}`, []string{"x", "x_idx"}},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			stage := parsertest.ParseStage(tc.input)

			definedNames := analyzer.DefinedFields(stage)
			if !cmp.Equal(tc.expectedNames, definedNames) {
				t.Fatalf("expected and actual are not equal\n	%s", cmp.Diff(tc.expectedNames, definedNames))
			}
		})
	}
}
