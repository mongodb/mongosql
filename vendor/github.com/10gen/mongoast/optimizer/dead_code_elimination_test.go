package optimizer_test

import (
	"context"
	"testing"

	"github.com/10gen/mongoast/internal/parsertest"
	"github.com/10gen/mongoast/optimizer"
	"github.com/10gen/mongoast/parser"

	"github.com/google/go-cmp/cmp"
)

func TestDeadCodeElimination(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"delete unnecessary $project items",
			`[
				{"$project": {"a": "$c", "b": "$d"}},
				{"$project": {"out": "$a"}}
			 ]`,
			`[
				{"$project": {"a": "$c"}},
				{"$project": {"out": "$a"}}
			 ]`,
		},
		{
			"inclusion projects use themselves",
			`[
				{"$project": {"a": "$foo"}},
				{"$project": {"a": 1}}
			]`,
			`[
				{"$project": {"a": "$foo"}},
				{"$project": {"a": 1}}
			]`,
		},
		{
			"convert whole $project stage to exclude all live values",
			`[
				{"$project": {"a": "$c", "b": "$d"}},
				{"$project": {"out": "$z", "out2": "$x"}}
			 ]`,
			`[
				{"$project": {"x": 0, "z": 0}},
				{"$project": {"out": "$z", "out2": "$x"}}
			 ]`,
		},
		{
			"properly handle project with _id exclusion",
			`[
    			{"$project": {"_id": {"$numberInt":"0"},
        					  "a": "$foo",
        					  "b": "$bar"
						  }},
    			{"$project": {"a": "$a",
                  			  "b": "$b"}},
				{"$group": {"_id": {},
                			"out_b": "$b"}}
			]`,
			// Note that the _id exclusion is moved to the end,
			// that's fine.
			`[
    			{"$project": {
        					  "b": "$bar",
							  "_id": {"$numberInt":"0"}
						  }},
    			{"$project": {"b": "$b"}},
				{"$group": {"_id": {},
                			"out_b": "$b"}}
			]`,
		},
		{
			"do not change exclusion $projects",
			`[
				{"$project": {"a": 0, "b": 0}},
				{"$project": {"out": "$z"}}
			 ]`,
			`[
				{"$project": {"a": 0, "b": 0}},
				{"$project": {"out": "$z"}}
			 ]`,
		},
		{
			"remove unnecessary $project stage",
			`[
				{"$collStats": {}},
				{"$limit": {"$numberLong":"1"}},
				{"$project": {"newField": {"$numberInt":"1"}}},
				{"$project": {"4": {"$literal": {"$numberLong":"4"}}}}
			]`,
			`[
				{"$collStats": {}},
				{"$limit": {"$numberLong":"1"}},
				{"$project": {"4": {"$literal": {"$numberLong":"4"}}}}
			]`,
		},
		{
			"delete unnecessary $addFields items",
			`[
				{"$addFields": {"a": {"$add": ["$a", "$b"]}, "b": "$d"}},
				{"$project": {"out": "$a"}}
			 ]`,
			`[
				{"$addFields": {"a": {"$add": ["$a", "$b"]}}},
				{"$project": {"out": "$a"}}
			 ]`,
		},
		{
			"delete more unnecessary $addFields items",
			`[
					{"$addFields": {"a": {"$add": ["$a", "$b"]}, "b": "$d", "c": 1}},
				{"$project": {"out": "$b"}}
			 ]`,
			`[
				{"$addFields": {"b": "$d"}},
				{"$project": {"out": "$b"}}
			 ]`,
		},
		{
			"delete whole unnecessary $addFields stage",
			`[
	   			{"$addFields": {"a": {"$add": ["$a", "$b"]}, "b": "$d", "c": 1}},
				{"$project": {"out": "$z"}}
			 ]`,
			`[
				{"$project": {"out": "$z"}}
			 ]`,
		},
		{
			"do not add refs for stages that will be removed",
			`[
				{"$addFields": {"a": 42}},
				{"$addFields": {"a": {"$add": ["$a", "$b"]}, "b": "$d", "c": 1}},
				{"$project": {"out": "$z"}}
			 ]`,
			`[
				{"$project": {"out": "$z"}}
			 ]`,
		},
		{
			"preserve $uwind and deps",
			`[
				{"$addFields": {"u" :"$z", "c": "$q"}},
				{"$unwind" : "$u"},
				{"$addFields": {"a": {"$add": ["$a", "$b"]}, "b": "$d", "c": 1}},
				{"$project": {"out": "$z"}}
			 ]`,
			`[
				{"$addFields": {"u" :"$z"}},
				{"$unwind" : "$u"},
				{"$project": {"out": "$z"}}
			 ]`,
		},
		{
			"preserve $group and deps",
			`[
				{"$addFields": {"u" :"$z", "c": "$q"}},
				{"$group" : {"_id": "$u", "z": {"$sum": "$u"}, "x": {"$sum": 1}}},
				{"$addFields": {"a": {"$add": ["$a", "$z"]}, "b": "$d", "c": 1}},
				{"$project": {"out": "$z"}},
				{"$match": {"out": 42}}
			 ]`,
			`[
				{"$addFields": {"u" :"$z"}},
				{"$group" : {"_id": "$u", "z": {"$sum": "$u"}}},
				{"$project": {"out": "$z"}},
				{"$match": {"out": 42}}
			 ]`,
		},
		{
			"delete whole unnecessary $lookup stage",
			`[
				{"$lookup": {"from": "foo", "as": "bar", "foreignField": "baz", "localField": "baz"}},
				{"$project": {"out": "$z"}}
			 ]`,
			`[
				{"$project": {"out": "$z"}}
			 ]`,
		},
		{
			"keep $lookup due to implicit $unwind dependency, but remove unnecessary $unwind index",
			`[
				{"$lookup": {"from": "foo", "as": "bar", "foreignField": "baz", "localField": "baz"}},
				{"$unwind": {"path": "$bar", "includeArrayIndex": "bar_idx"}},
				{"$project": {"out": "$z"}}
			 ]`,
			`[
				{"$lookup": {"from": "foo", "as": "bar", "foreignField": "baz", "localField": "baz"}},
				{"$unwind": {"path": "$bar"}},
				{"$project": {"out": "$z"}}
			 ]`,
		},
		{
			"do not remove $sort stage",
			`[
				{"$addFields": {"drop": null}},
				{"$lookup": {"from": "foo", "as": "bar", "foreignField": "baz", "localField": "baz"}},
				{"$sort": {"bar": 1}},
				{"$project": {"out": "$z"}}
			 ]`,
			`[
				{"$lookup": {"from": "foo", "as": "bar", "foreignField": "baz", "localField": "baz"}},
				{"$sort": {"bar": 1}},
				{"$project": {"out": "$z"}}
			 ]`,
		},
		{
			"do not remove $sortByCount stage",
			`[
				{"$addFields": {"drop": null}},
				{"$lookup": {"from": "foo", "as": "bar", "foreignField": "baz", "localField": "baz"}},
				{"$sortByCount": "$bar"},
				{"$project": {"out": "$z"}}
			 ]`,
			`[
				{"$lookup": {"from": "foo", "as": "bar", "foreignField": "baz", "localField": "baz"}},
				{"$sortByCount": "$bar"},
				{"$project": {"out": "$z"}}
			 ]`,
		},
		{
			"remove unnecessary $replaceRoot fields",
			`[
				{"$replaceRoot": {"newRoot": {"b": "$c", "a": "$d", "e": "$f" }}},
				{"$project": {"out": "$a"}}
			 ]`,
			`[
				{"$replaceRoot": {"newRoot": {"a": "$d"}}},
				{"$project": {"out": "$a"}}
			 ]`,
		},
		{
			"cannot remove anything when $replaceRoot newRoot is a field",
			`[
				{"$replaceRoot": {"newRoot": "$b"}},
				{"$project": {"out": "$a"}}
			 ]`,
			`[
				{"$replaceRoot": {"newRoot": "$b"}},
				{"$project": {"out": "$a"}}
			 ]`,
		},
		{
			"$replaceRoot is a field killer",
			`[
				{"$project": {"drop": null, "b": 1}},
				{"$replaceRoot": {"newRoot": "$b"}}
			 ]`,
			`[
				{"$project": {"b": 1}},
				{"$replaceRoot": {"newRoot": "$b"}}
			 ]`,
		},
		{
			"field killers should actually kill live fields (these two c's are different)",
			`[
				{"$project": {"c" : 1, "b": 1}},
				{"$project": {"drop": null, "b": 1}},
				{"$replaceRoot": {"newRoot": "$b"}},
				{"$project": {"c": 1}}
			 ]`,
			`[
				{"$project": {"b": 1}},
				{"$project": {"b": 1}},
				{"$replaceRoot": {"newRoot": "$b"}},
				{"$project": {"c": 1}}
			 ]`,
		},
		{
			"remove unnecessary $bucket output defs",
			`[
				{"$bucket": {"groupBy": "$_id", "boundaries": [0,5,10], "output": {"b": {"$sum": "$c"}, "a": {"$sum": "$d"}, "e": {"$sum": "$f"}}}},
				{"$project": {"out": "$a"}}
			 ]`,
			`[
				{"$bucket": {"groupBy": "$_id", "boundaries": [0,5,10], "output": {"a": {"$sum": "$d"}}}},
				{"$project": {"out": "$a"}}
			 ]`,
		},
		{
			"remove unnecessary $bucketAuto output defs",
			`[
				{"$bucketAuto": {"groupBy": "$_id", "buckets": 5, "output": {"b": {"$sum": "$c"}, "a": {"$sum": "$d"}, "e": {"$sum": "$f"}}}},
				{"$project": {"out": "$a"}}
				 ]`,
			`[
				{"$bucketAuto": {"groupBy": "$_id", "buckets": 5, "output": {"a": {"$sum": "$d"}}}},
				{"$project": {"out": "$a"}}
			 ]`,
		},
		{
			"$bucket is a field killer",
			`[
				{"$addFields": {"drop": null}},
				{"$bucket": {"groupBy": "$_id", "boundaries": [0,5,10], "output": {"b": "$c", "a": "$d", "e": "$f" }}}
				 ]`,
			`[
				{"$bucket": {"groupBy": "$_id", "boundaries": [0,5,10], "output": {"b": "$c", "a": "$d", "e": "$f" }}}
			 ]`,
		},
		{
			"$bucketAuto is a field killer",
			`[
				{"$addFields": {"drop": null}},
				{"$bucketAuto": {"groupBy": "$_id", "buckets": 5, "output": {"b": "$c", "a": "$d", "e": "$f" }}}
			 ]`,
			`[
				{"$bucketAuto": {"groupBy": "$_id", "buckets": 5, "output": {"b": "$c", "a": "$d", "e": "$f" }}}
			 ]`,
		},
		{
			"remove unnecessary $facet output defs",
			`[
				{"$addFields": {"c": 42, "drop": null}},
				{"$facet": {"a": [{"$match": {"c" : 42}}], "b": [] }},
				{"$project": {"out": "$a"}}
			]`,
			`[
				{"$addFields": {"c": 42}},
				{"$facet": {"a": [{"$match": {"c" : 42}}] }},
				{"$project": {"out": "$a"}}
			 ]`,
		},
		{
			"$facet is a field killer",
			`[
				{"$addFields": {"drop": null}},
				{"$facet": {"a": [], "b": [] }}
			 ]`,
			`[
				{"$facet": {"a": [], "b": [] }}
			 ]`,
		},
		{
			"remove unnecessary field defs in $facet sub pipeline",
			`[
				{"$addFields": {"c": 42, "drop": null}},
				{"$facet": {"a": [
									 {"$match": {"c" : 42}},
									 {"$addFields": {"$dropMe2": null}},
									 {"$project": {"c": 1}}
								],
							"b": [] }},
				{"$project": {"out": "$a"}}
			 ]`,
			`[
				{"$addFields": {"c": 42}},
				{"$facet": {"a": [
									 {"$match": {"c" : 42}},
									 {"$project": {"c": 1}}
								] }},
				{"$project": {"out": "$a"}}
			 ]`,
		},
		{
			"do not remove includeArrayIndex on subfield (bug fix)",
			`[
					{"$unwind": {"path": "$loc.b.c","includeArrayIndex": "loc.b.c_idx"}},
					{"$match": {"loc.b.c_idx": {"$eq": {"$numberLong":"1"}}}},
					{"$project": {"test_DOT_test4_loc_b_c_DOT__id": "$_id",
						"test_DOT_test4_loc_b_c_DOT_loc_DOT_b_DOT_c": "$loc.b.c",
						"test_DOT_test4_loc_b_c_DOT_loc_DOT_b_DOT_c_idx": "$loc.b.c_idx",
						"_id": {"$numberInt":"0"}}}
			]`,
			`[
					{"$unwind": {"path": "$loc.b.c","includeArrayIndex": "loc.b.c_idx"}},
					{"$match": {"loc.b.c_idx": {"$eq": {"$numberLong":"1"}}}},
					{"$project": {"test_DOT_test4_loc_b_c_DOT__id": "$_id",
						"test_DOT_test4_loc_b_c_DOT_loc_DOT_b_DOT_c": "$loc.b.c",
						"test_DOT_test4_loc_b_c_DOT_loc_DOT_b_DOT_c_idx": "$loc.b.c_idx",
						"_id": {"$numberInt":"0"}}}
			]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in := parsertest.ParsePipeline(tc.input)
			expected := parsertest.ParsePipeline(tc.expected)
			actual := optimizer.RunPasses(context.Background(), in, optimizer.DeadCodeElimination)

			expectedStr := parser.DeparsePipeline(expected).String()
			actualStr := parser.DeparsePipeline(actual).String()

			if !cmp.Equal(expectedStr, actualStr) {
				t.Fatalf("\nexpected:\n %s\ngot:\n %s", expectedStr, actualStr)
			}
		})
	}
}
