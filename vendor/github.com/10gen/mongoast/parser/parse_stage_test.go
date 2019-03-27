package parser_test

import (
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/10gen/mongoast/internal/parsertest"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"

	"github.com/google/go-cmp/cmp"
)

func TestParseStage(t *testing.T) {
	testCases := []struct {
		input    string
		expected ast.Stage
	}{
		{
			`{"$count": "a"}`,
			ast.NewCountStage("a"),
		},
		{
			`{"$group": {"_id": 1, "count": { "$sum": "$a" }}}`,
			ast.NewGroupStage(
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewGroupItem(
					"count",
					ast.NewFunction(
						"$sum",
						ast.NewFieldRef("a", nil),
					),
				),
			),
		},
		{
			`{"$limit": 10}`,
			ast.NewLimitStage(10),
		},
		{
			`{"$limit": 10.0}`,
			ast.NewLimitStage(10),
		},
		{
			`{"$sample": {"size": 10}}`,
			ast.NewSampleStage(10),
		},
		{
			`{"$sample": {"size": 10.0}}`,
			ast.NewSampleStage(10),
		},
		{
			`{"$match": {"a": 1}}`,
			ast.NewMatchStage(
				ast.NewBinary(ast.Equals,
					ast.NewFieldRef("a", nil),
					ast.NewConstant(bsonutil.Int32(1))),
			),
		},
		{
			`{"$project": {"a": "$a", "b": 1, "c": 0, "d": -1, "e": 1.2, "f": false}}`,
			ast.NewProjectStage(
				ast.NewAssignProjectItem("a", ast.NewFieldRef("a", nil)),
				ast.NewIncludeProjectItem(ast.NewFieldRef("b", nil)),
				ast.NewExcludeProjectItem(ast.NewFieldRef("c", nil)),
				ast.NewIncludeProjectItem(ast.NewFieldRef("d", nil)),
				ast.NewIncludeProjectItem(ast.NewFieldRef("e", nil)),
				ast.NewExcludeProjectItem(ast.NewFieldRef("f", nil)),
			),
		},
		{
			`{"$project": {"foo": {"a": 1, "b": 0}}}`,
			ast.NewProjectStage(
				ast.NewIncludeProjectItem(ast.NewFieldRef("a", ast.NewFieldRef("foo", nil))),
				ast.NewExcludeProjectItem(ast.NewFieldRef("b", ast.NewFieldRef("foo", nil))),
			),
		},
		{
			`{"$project": {"x": { "$gt": [{ "$numberDecimal": "1" }, { "$numberDecimal": "2" }] }}}`,
			ast.NewProjectStage(
				ast.NewAssignProjectItem("x", ast.NewBinary(
					ast.GreaterThan,
					ast.NewConstant(bsonutil.Decimal128FromInt64(1)),
					ast.NewConstant(bsonutil.Decimal128FromInt64(2)),
				)),
			),
		},
		{
			`{"$project": {"foo": {"d": {"$add": ["$foo.a", "$foo.b"]}}}}`,
			ast.NewProjectStage(
				ast.NewAssignProjectItem(
					"foo.d", ast.NewFunction(
						"$add", ast.NewArray(
							ast.NewFieldRef("a", ast.NewFieldRef("foo", nil)),
							ast.NewFieldRef("b", ast.NewFieldRef("foo", nil)),
						),
					),
				),
			),
		},
		{
			`{"$skip": 10}`,
			ast.NewSkipStage(10),
		},
		{
			`{"$sort": {"a": -1, "b": 1, "c": -1.3, "d": 1.9}}`,
			ast.NewSortStage(
				ast.NewSortItem(ast.NewFieldRef("a", nil), true),
				ast.NewSortItem(ast.NewFieldRef("b", nil), false),
				ast.NewSortItem(ast.NewFieldRef("c", nil), true),
				ast.NewSortItem(ast.NewFieldRef("d", nil), false),
			),
		},
		{
			`{"$sort": {"a.2": -1, "b": 1}}`,
			ast.NewSortStage(
				ast.NewSortItem(ast.NewFieldOrArrayIndexRef(2, ast.NewFieldRef("a", nil)), true),
				ast.NewSortItem(ast.NewFieldRef("b", nil), false),
			),
		},
		{
			`{"$hmmm": 10}`,
			ast.NewUnknown(bsonutil.DocumentFromElements(
				"$hmmm", bsonutil.Int32(10),
			)),
		},
		{
			`{"$unwind": "$a"}`,
			ast.NewUnwindStage(
				ast.NewFieldRef("a", nil),
				"",
				false,
			),
		},
		{
			`{"$unwind": "$a.b"}`,
			ast.NewUnwindStage(
				ast.NewFieldRef("b", ast.NewFieldRef("a", nil)),
				"",
				false,
			),
		},
		{
			`{"$unwind": { "path": "$a.b", "includeArrayIndex": "c", "preserveNullAndEmptyArrays": true}}`,
			ast.NewUnwindStage(
				ast.NewFieldRef("b", ast.NewFieldRef("a", nil)),
				"c",
				true,
			),
		},
		{
			`{"$lookup": { "from": "foo", "localField": "a", "foreignField": "b", "as": "bar"}}`,
			ast.NewLookupStage("foo", "a", "b", "bar", nil, nil),
		},
		{
			`{"$lookup": { "from": "foo", "pipeline": [{"$match": {"a": 1}}], "as": "bar"}}`,
			ast.NewLookupStage(
				"foo",
				"",
				"",
				"bar",
				nil,
				ast.NewPipeline(
					ast.NewMatchStage(
						ast.NewBinary(
							ast.Equals,
							ast.NewFieldRef("a", nil),
							ast.NewConstant(bsonutil.Int32(1)),
						),
					),
				),
			),
		},
		{
			`{"$lookup": { "from": "foo", "let": {"x": "$a"}, "pipeline": [{"$match": {"x": 1}}], "as": "bar"}}`,
			ast.NewLookupStage(
				"foo",
				"",
				"",
				"bar",
				[]*ast.LookupLetItem{
					ast.NewLookupLetItem(
						"x",
						ast.NewFieldRef("a", nil),
					),
				},
				ast.NewPipeline(
					ast.NewMatchStage(
						ast.NewBinary(
							ast.Equals,
							ast.NewFieldRef("x", nil),
							ast.NewConstant(bsonutil.Int32(1)),
						),
					),
				),
			),
		},
		{
			`{"$collStats": {}}`,
			ast.NewCollStatsStage(nil, nil, nil),
		},
		{
			`{"$collStats": { "latencyStats": {}}}`,
			ast.NewCollStatsStage(
				ast.NewCollStatsLatencyStats(false),
				nil,
				nil,
			),
		},
		{
			`{"$collStats": { "latencyStats": { "histograms": false }}}`,
			ast.NewCollStatsStage(
				ast.NewCollStatsLatencyStats(false),
				nil,
				nil,
			),
		},
		{
			`{"$collStats": { "latencyStats": { "histograms": true }}}`,
			ast.NewCollStatsStage(
				ast.NewCollStatsLatencyStats(true),
				nil,
				nil,
			),
		},
		{
			`{"$collStats": { "storageStats": {}}}`,
			ast.NewCollStatsStage(
				nil,
				ast.NewCollStatsStorageStats(),
				nil,
			),
		},
		{
			`{"$collStats": { "count": {}}}`,
			ast.NewCollStatsStage(
				nil,
				nil,
				ast.NewCollStatsCount(),
			),
		},
		{
			`{"$collStats": { "latencyStats": {}, "storageStats": {}, "count": {}}}`,
			ast.NewCollStatsStage(
				ast.NewCollStatsLatencyStats(false),
				ast.NewCollStatsStorageStats(),
				ast.NewCollStatsCount(),
			),
		},
		{
			`{"$addFields": { "a": "$foo", "b": 1}}`,
			ast.NewAddFieldsStage(
				ast.NewAddFieldsItem("a", ast.NewFieldRef("foo", nil)),
				ast.NewAddFieldsItem("b", ast.NewConstant(bsonutil.Int32(1))),
			),
		},
		{
			`{"$replaceRoot": { "newRoot": { "a": "$foo" }}}`,
			ast.NewReplaceRootStage(
				ast.NewDocument(
					ast.NewDocumentElement("a", ast.NewFieldRef("foo", nil)),
				),
			),
		},
		{
			`{"$facet": { "a1": [{ "$match": { "a": 1 } }], "a2": [{ "$match": { "a": 2 } }] } }`,
			ast.NewFacetStage(
				ast.NewFacetItem(
					"a1", ast.NewPipeline(
						ast.NewMatchStage(
							ast.NewBinary(
								ast.Equals,
								ast.NewFieldRef("a", nil),
								ast.NewConstant(bsonutil.Int32(1)),
							),
						),
					),
				),
				ast.NewFacetItem(
					"a2", ast.NewPipeline(
						ast.NewMatchStage(
							ast.NewBinary(
								ast.Equals,
								ast.NewFieldRef("a", nil),
								ast.NewConstant(bsonutil.Int32(2)),
							),
						),
					),
				),
			),
		},
		{
			`{"$sortByCount": "$a"}`,
			ast.NewSortByCountStage(
				ast.NewFieldRef("a", nil),
			),
		},
		{
			`{"$bucket": { "groupBy": "$a", "boundaries": [0, 10, 20] } }`,
			ast.NewBucketStage(
				ast.NewFieldRef("a", nil),
				[]bsoncore.Value{
					bsonutil.Int32(0),
					bsonutil.Int32(10),
					bsonutil.Int32(20),
				},
				nil,
				nil,
			),
		},
		{
			`{"$bucket": { "groupBy": "$a", "boundaries": [0, 10, 20], "default": -1, "output": { "total": { "$sum": "$a" } } } }`,
			ast.NewBucketStage(
				ast.NewFieldRef("a", nil),
				[]bsoncore.Value{
					bsonutil.Int32(0),
					bsonutil.Int32(10),
					bsonutil.Int32(20),
				},
				bsonutil.ValuePtr(bsonutil.Int32(-1)),
				[]*ast.GroupItem{
					ast.NewGroupItem(
						"total", ast.NewFunction(
							"$sum", ast.NewFieldRef("a", nil),
						),
					),
				},
			),
		},
		{
			`{"$bucketAuto": { "groupBy": "$a", "buckets": 2 } }`,
			ast.NewBucketAutoStage(
				ast.NewFieldRef("a", nil),
				2,
				nil,
				"",
			),
		},
		{
			`{"$bucketAuto": { "groupBy": "$a", "buckets": 2, "output": { "total": { "$sum": "$a" } }, "granularity": "R5" } }`,
			ast.NewBucketAutoStage(
				ast.NewFieldRef("a", nil),
				2,
				[]*ast.GroupItem{
					ast.NewGroupItem(
						"total", ast.NewFunction(
							"$sum", ast.NewFieldRef("a", nil),
						),
					),
				},
				"R5",
			),
		},
		{
			`{"$redact": { "$cond": { "if": { "$eq": ["$a", 5] }, "then": "$$PRUNE", "else": "$$DESCEND" } } }`,
			ast.NewRedactStage(
				ast.NewConditional(
					ast.NewBinary(
						ast.Equals,
						ast.NewFieldRef("a", nil),
						ast.NewConstant(bsonutil.Int32(5)),
					),
					ast.NewVariableRef("PRUNE"),
					ast.NewVariableRef("DESCEND"),
				),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual := parsertest.ParseStage(tc.input)

			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("stages are not equal\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}

func TestInvalidParseStage(t *testing.T) {
	invalidTestCases := []string{
		`{"$project": "{}"}`,
	}

	for _, input := range invalidTestCases {
		t.Run(input, func(t *testing.T) {
			_, err := parsertest.ParseStageErr(input)
			if err == nil {
				t.Fatalf("parsing stage should have resulted in an error\n  %s", input)
			}
		})
	}
}
