package analyzer_test

import (
	"testing"

	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/internal/parsertest"

	"github.com/google/go-cmp/cmp"
)

func TestFieldKiller(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{`{"$addFields":	{ "b": "$b", "c" : { "$add": ["$c", 1] } } }`,
			false},
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
			true},
		{`{"$bucketAuto": {
					"groupBy": "$year",
				 	"buckets": 3,
					"output": {
						"count": { "$sum": 1 },
						"years": { "$push": "$year" }
					}
			}
		}`,
			true},
		{`{"$collStats": {}}`, true},
		{`{"$count": "foo"}`, true},
		{`{"$facet":

			{
				"categorizedByTags": [
						{ "$unwind": "$tags" },
						{ "$sortByCount": "$tags" }
				]
			}
		}`, true},
		{`{"$group" : {
			"_id" : { "month": { "$month": "$date" }, "day": { "$dayOfMonth": "$date" }, "year": { "$year": "$date" } },
			"totalPrice": { "$sum": { "$multiply": [ "$price", "$quantity" ] } },
			"averageQuantity": { "$avg": "$quantity" },
			"count": { "$sum": 1 }
        }}`, true},
		{`{"$lookup" : {"from": "bar", "localField": "foo", "foreignField": "foo", "as": "bazz"}}`,
			false},
		{`{"$project": { "a": 1, "b": "$b", "c" : { "$add": ["$c", 1] } } }`,
			true},
		{`{"$project": { "a.b": 1, "b": "$b", "c" : { "$add": ["$c", 1] } } }`,
			true},
		{`{"$project": { "a.b": 0, "b": 0, "c" : 0 }}`,
			false},
		{`{"$project": { "_id": 0 }}`,
			false},
		{`{"$project": { "a": 1, "_id": 0 }}`,
			true},
		{`{"$project": { "a": "$b", "_id": 0 }}`,
			true},
		{`{"$replaceRoot": {"newRoot": "$a"}}`, true},
		{`{"$replaceRoot": {"newRoot" : { "a": 1, "b": "$b", "c" : { "$add": ["$c", 1] } } } }`,
			true},
		{`{"$unwind": "$x"}`, false},
		{`{"$unwind": {"path": "$x", "includeArrayIndex": "x_idx"}}`, false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			stage := parsertest.ParseStage(tc.input)

			out := analyzer.IsFieldKiller(stage)
			if !cmp.Equal(tc.expected, out) {
				t.Fatalf("expected and actual are not equal\n	%s", cmp.Diff(tc.expected, out))
			}
		})
	}
}
