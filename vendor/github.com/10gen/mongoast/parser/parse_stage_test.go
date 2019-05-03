package parser_test

import (
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/10gen/mongoast/internal/parsertest"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"

	"github.com/google/go-cmp/cmp"
)

func TestParseStage(t *testing.T) {
	testCases := []struct {
		input    string
		expected ast.Stage
		err      error
	}{
		{
			`{"$count": "a"}`,
			ast.NewCountStage("a"),
			nil,
		},
		{
			`{"$count": 1}`,
			nil,
			errors.New("$count stage must have a string as its only argument"),
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
			nil,
		},
		{
			`{"$group": 1}`,
			nil,
			errors.New("$group stage must have a document as its only argument"),
		},
		{
			`{"$group": {"count": { "$sum": "$a" }}}`,
			nil,
			errors.New("$group stage must have document with _id field"),
		},
		{
			`{"$limit": 10}`,
			ast.NewLimitStage(10),
			nil,
		},
		{
			`{"$limit": 10.0}`,
			ast.NewLimitStage(10),
			nil,
		},
		{
			`{"$limit": "x"}`,
			nil,
			errors.New("$limit stage must have an integer as its only argument"),
		},
		{
			`{"$limit": 0}`,
			nil,
			errors.New("argument to $limit stage must be positive"),
		},
		{
			`{"$limit": -1}`,
			nil,
			errors.New("argument to $limit stage must be positive"),
		},
		{
			`{"$sample": {"size": 10}}`,
			ast.NewSampleStage(10),
			nil,
		},
		{
			`{"$sample": {"size": 10.0}}`,
			ast.NewSampleStage(10),
			nil,
		},
		{
			`{"$sample": 1}`,
			nil,
			errors.New("$sample stage must have a document as its only argument"),
		},
		{
			`{"$sample": {}}`,
			nil,
			errors.New("$sample stage must specify a size"),
		},
		{
			`{"$sample": {"size": "x"}}`,
			nil,
			errors.New("size argument to $sample must be a number"),
		},
		{
			`{"$sample": {"size": -1}}`,
			nil,
			errors.New("size argument to $sample must not be negative"),
		},
		{
			`{"$sample": {"size": 1, "foo": 1}}`,
			nil,
			errors.New("unrecognized option to $sample: foo"),
		},
		{
			`{"$match": {"a": 1}}`,
			ast.NewMatchStage(
				ast.NewBinary(ast.Equals,
					ast.NewFieldRef("a", nil),
					ast.NewConstant(bsonutil.Int32(1))),
			),
			nil,
		},
		{
			`{"$match": {}}`,
			ast.NewMatchStage(
				ast.NewConstant(bsonutil.Boolean(true)),
			),
			nil,
		},
		{
			`{"$match": 1}`,
			nil,
			errors.New("$match stage must have a document as its only argument"),
		},
		{
			`{"$project": {"a": "$a", "b": 1, "c": 1, "d": -1, "e": 1.2, "_id": false}}`,
			ast.NewProjectStage(
				ast.NewAssignProjectItem("a", ast.NewFieldRef("a", nil)),
				ast.NewIncludeProjectItem(ast.NewFieldRef("b", nil)),
				ast.NewIncludeProjectItem(ast.NewFieldRef("c", nil)),
				ast.NewIncludeProjectItem(ast.NewFieldRef("d", nil)),
				ast.NewIncludeProjectItem(ast.NewFieldRef("e", nil)),
				ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
			),
			nil,
		},
		{
			`{"$project": {"a": {"$numberLong": "1"}, "_id": {"$numberLong": "0"}}}`,
			ast.NewProjectStage(
				ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
				ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
			),
			nil,
		},
		{
			`{"$project": {"a": {"$numberDecimal": "1"}, "_id": {"$numberDecimal": "0"}}}`,
			ast.NewProjectStage(
				ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
				ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
			),
			nil,
		},
		{
			`{"$project": {"foo": {"a": 1, "b": 1}, "_id": 0}}`,
			ast.NewProjectStage(
				ast.NewIncludeProjectItem(ast.NewFieldRef("a", ast.NewFieldRef("foo", nil))),
				ast.NewIncludeProjectItem(ast.NewFieldRef("b", ast.NewFieldRef("foo", nil))),
				ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
			),
			nil,
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
			nil,
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
			nil,
		},
		{
			`{"$project": 1}`,
			nil,
			errors.New("$project stage must have a document as its only argument"),
		},
		{
			`{"$project": {}}`,
			nil,
			errors.New("$project must have at least one field"),
		},
		{
			`{"$project": {"a": 1, "b": 0}}`,
			nil,
			errors.New("cannot exclude fields other than '_id' in an inclusion projection"),
		},
		{
			`{"$project": {"a": 1, "_id": 0, "c": 0}}`,
			nil,
			errors.New("cannot exclude fields other than '_id' in an inclusion projection"),
		},
		{
			`{"$project": {"a": 0, "_id": 0, "c": {"$add": [1, 3]}}}`,
			nil,
			errors.New("cannot exclude fields other than '_id' in an inclusion projection"),
		},
		{
			`{"$project": {"a": 1, "_id": 0, "c": {"$add": [1, 3]}}}`,
			ast.NewProjectStage(
				ast.NewIncludeProjectItem(ast.NewFieldRef("a", nil)),
				ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
				ast.NewAssignProjectItem(
					"c", ast.NewFunction(
						"$add", ast.NewArray(
							ast.NewConstant(bsonutil.Int32(1)),
							ast.NewConstant(bsonutil.Int32(3)),
						),
					),
				),
			),
			nil,
		},
		{
			`{"$skip": 10}`,
			ast.NewSkipStage(10),
			nil,
		},
		{
			`{"$skip": "x"}`,
			nil,
			errors.New("$skip stage must have an integer as its only argument"),
		},
		{
			`{"$skip": -1}`,
			nil,
			errors.New("argument to $skip cannot be negative"),
		},
		{
			`{"$sort": {"a": -1, "b": 1, "c": -1.3, "d": 1.9}}`,
			ast.NewSortStage(
				ast.NewSortItem(ast.NewFieldRef("a", nil), true),
				ast.NewSortItem(ast.NewFieldRef("b", nil), false),
				ast.NewSortItem(ast.NewFieldRef("c", nil), true),
				ast.NewSortItem(ast.NewFieldRef("d", nil), false),
			),
			nil,
		},
		{
			`{"$sort": {"a": {"$numberDecimal": "-1"}, "b": {"$numberDecimal": "1"}}}`,
			ast.NewSortStage(
				ast.NewSortItem(ast.NewFieldRef("a", nil), true),
				ast.NewSortItem(ast.NewFieldRef("b", nil), false),
			),
			nil,
		},
		{
			`{"$sort": {"a": {"$numberDecimal": "-1.3"}, "b": {"$numberDecimal": "1.9"}}}`,
			ast.NewSortStage(
				ast.NewSortItem(ast.NewFieldRef("a", nil), true),
				ast.NewSortItem(ast.NewFieldRef("b", nil), false),
			),
			nil,
		},
		{
			`{"$sort": {"a.2": -1, "b": 1}}`,
			ast.NewSortStage(
				ast.NewSortItem(ast.NewFieldOrArrayIndexRef(2, ast.NewFieldRef("a", nil)), true),
				ast.NewSortItem(ast.NewFieldRef("b", nil), false),
			),
			nil,
		},
		{
			`{"$sort": 1}`,
			nil,
			errors.New("$sort stage must have a document as its only argument"),
		},
		{
			`{"$sort": {"a": -2.1}}`,
			nil,
			errors.New("failed parsing sort items: $sort key ordering must be 1 (for ascending) or -1 (for descending)"),
		},
		{
			`{"$sort": {"a": 0.1}}`,
			nil,
			errors.New("failed parsing sort items: $sort key ordering must be 1 (for ascending) or -1 (for descending)"),
		},
		{
			`{"$sort": {"a": 2.1}}`,
			nil,
			errors.New("failed parsing sort items: $sort key ordering must be 1 (for ascending) or -1 (for descending)"),
		},
		{
			`{"$sort": {"a": -2}}`,
			nil,
			errors.New("failed parsing sort items: $sort key ordering must be 1 (for ascending) or -1 (for descending)"),
		},
		{
			`{"$sort": {"a": 0}}`,
			nil,
			errors.New("failed parsing sort items: $sort key ordering must be 1 (for ascending) or -1 (for descending)"),
		},
		{
			`{"$sort": {"a": 2}}`,
			nil,
			errors.New("failed parsing sort items: $sort key ordering must be 1 (for ascending) or -1 (for descending)"),
		},
		{
			`{"$sort": {"a": {"$numberLong": "-2"}}}`,
			nil,
			errors.New("failed parsing sort items: $sort key ordering must be 1 (for ascending) or -1 (for descending)"),
		},
		{
			`{"$sort": {"a": {"$numberLong": "0"}}}`,
			nil,
			errors.New("failed parsing sort items: $sort key ordering must be 1 (for ascending) or -1 (for descending)"),
		},
		{
			`{"$sort": {"a": {"$numberLong": "2"}}}`,
			nil,
			errors.New("failed parsing sort items: $sort key ordering must be 1 (for ascending) or -1 (for descending)"),
		},
		{
			`{"$sort": {"a": {"$numberDecimal": "-2"}}}`,
			nil,
			errors.New("failed parsing sort items: $sort key ordering must be 1 (for ascending) or -1 (for descending)"),
		},
		{
			`{"$sort": {"a": {"$numberDecimal": "0"}}}`,
			nil,
			errors.New("failed parsing sort items: $sort key ordering must be 1 (for ascending) or -1 (for descending)"),
		},
		{
			`{"$sort": {"a": {"$numberDecimal": "2"}}}`,
			nil,
			errors.New("failed parsing sort items: $sort key ordering must be 1 (for ascending) or -1 (for descending)"),
		},
		{
			`{"$sort": {"a": "x"}}`,
			nil,
			errors.New("failed parsing sort items: $sort key ordering must be specified using a number"),
		},
		{
			`{"$sort": {}}`,
			nil,
			errors.New("$sort stage must have at least one sort key"),
		},
		{
			`{"$hmmm": 10}`,
			ast.NewUnknown(bsonutil.DocumentFromElements(
				"$hmmm", bsonutil.Int32(10),
			)),
			nil,
		},
		{
			`{"$unwind": "$a"}`,
			ast.NewUnwindStage(
				ast.NewFieldRef("a", nil),
				"",
				false,
			),
			nil,
		},
		{
			`{"$unwind": "$a.b"}`,
			ast.NewUnwindStage(
				ast.NewFieldRef("b", ast.NewFieldRef("a", nil)),
				"",
				false,
			),
			nil,
		},
		{
			`{"$unwind": { "path": "$a.b", "includeArrayIndex": "c", "preserveNullAndEmptyArrays": true}}`,
			ast.NewUnwindStage(
				ast.NewFieldRef("b", ast.NewFieldRef("a", nil)),
				"c",
				true,
			),
			nil,
		},
		{
			`{"$unwind": { "path": "$a.b", "includeArrayIndex": true}}`,
			nil,
			errors.New("includeArrayIndex must be a string"),
		},
		{
			`{"$unwind": { "path": "$a.b", "preserveNullAndEmptyArrays": "x"}}`,
			nil,
			errors.New("preserveNullAndEmptyArrays must be a boolean"),
		},
		{
			`{"$unwind": "$$x"}`,
			nil,
			errors.New("unwind field path must be a field reference"),
		},
		{
			`{"$unwind": 1}`,
			nil,
			errors.New("expected either a string or an object as specification for $unwind stage, got 32-bit integer"),
		},
		{
			`{"$unwind": {}}`,
			nil,
			errors.New("no path specified to $unwind stage"),
		},
		{
			`{"$lookup": { "from": "foo", "localField": "a", "foreignField": "b", "as": "bar"}}`,
			ast.NewLookupStage("foo", "a", "b", "bar", nil, nil),
			nil,
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
			nil,
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
			nil,
		},
		{
			`{"$lookup": { "from": "foo", "let": {}, "pipeline": [{"$match": {"x": 1}}], "as": "bar"}}`,
			ast.NewLookupStage(
				"foo",
				"",
				"",
				"bar",
				[]*ast.LookupLetItem{},
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
			nil,
		},
		{
			`{"$lookup": { "from": "foo", "pipeline": [], "as": "bar"}}`,
			ast.NewLookupStage(
				"foo",
				"",
				"",
				"bar",
				nil,
				ast.NewPipeline(
					[]ast.Stage{}...,
				),
			),
			nil,
		},
		{
			`{"$lookup": 1}`,
			nil,
			errors.New("$lookup stage must have a document as its only argument"),
		},
		{
			`{"$lookup": { "from": 1, "pipeline": [], "as": "bar" }}`,
			nil,
			errors.New("$lookup argument 'from: {\"$numberInt\":\"1\"}' must be a string, is type 32-bit integer"),
		},
		{
			`{"$lookup": { "from": "foo", "localField": 1, "foreignField": "b", "as": "bar" }}`,
			nil,
			errors.New("$lookup argument 'localField: {\"$numberInt\":\"1\"}' must be a string, is type 32-bit integer"),
		},
		{
			`{"$lookup": { "from": "foo", "localField": "a", "foreignField": 1, "as": "bar" }}`,
			nil,
			errors.New("$lookup argument 'foreignField: {\"$numberInt\":\"1\"}' must be a string, is type 32-bit integer"),
		},
		{
			`{"$lookup": { "from": "foo", "pipeline": [], "as": 1 }}`,
			nil,
			errors.New("$lookup argument 'as: {\"$numberInt\":\"1\"}' must be a string, is type 32-bit integer"),
		},
		{
			`{"$lookup": { "from": "foo", "let": "x", "pipeline": [], "as": "bar" }}`,
			nil,
			errors.New("$lookup argument 'let: \"x\"' must be an object, is type string"),
		},
		{
			`{"$lookup": { "from": "foo", "pipeline": 1, "as": "bar" }}`,
			nil,
			errors.New("'pipeline' option must be specified as an array"),
		},
		{
			`{"$lookup": { "from": "foo", "pipeline": [], "as": "bar", "foo": 1 }}`,
			nil,
			errors.New("unknown argument to $lookup: foo"),
		},
		{
			`{"$lookup": { "pipeline": [], "as": "bar" }}`,
			nil,
			errors.New("missing 'from' option to $lookup stage specification: {\"pipeline\": [],\"as\": \"bar\"}"),
		},
		{
			`{"$lookup": { "from": "foo", "pipeline": [] }}`,
			nil,
			errors.New("must specify 'as' field for a $lookup"),
		},
		{
			`{"$lookup": { "from": "foo", "as": "bar" }}`,
			nil,
			errors.New("$lookup requires either 'pipeline' or both 'localField' and 'foreignField' to be specified"),
		},
		{
			`{"$lookup": { "from": "foo", "localField": "a", "as": "bar" }}`,
			nil,
			errors.New("$lookup requires either 'pipeline' or both 'localField' and 'foreignField' to be specified"),
		},
		{
			`{"$lookup": { "from": "foo", "foreignField": "b", "as": "bar" }}`,
			nil,
			errors.New("$lookup requires either 'pipeline' or both 'localField' and 'foreignField' to be specified"),
		},
		{
			`{"$lookup": { "from": "foo", "pipeline": [], "localField": "a", "as": "bar" }}`,
			nil,
			errors.New("$lookup with 'pipeline' may not specify 'localField' or 'foreignField'"),
		},
		{
			`{"$lookup": { "from": "foo", "pipeline": [], "foreignField": "b", "as": "bar" }}`,
			nil,
			errors.New("$lookup with 'pipeline' may not specify 'localField' or 'foreignField'"),
		},
		{
			`{"$collStats": {}}`,
			ast.NewCollStatsStage(nil, nil, nil),
			nil,
		},
		{
			`{"$collStats": { "latencyStats": {}}}`,
			ast.NewCollStatsStage(
				ast.NewCollStatsLatencyStats(false),
				nil,
				nil,
			),
			nil,
		},
		{
			`{"$collStats": { "latencyStats": { "histograms": false }}}`,
			ast.NewCollStatsStage(
				ast.NewCollStatsLatencyStats(false),
				nil,
				nil,
			),
			nil,
		},
		{
			`{"$collStats": { "latencyStats": { "histograms": true }}}`,
			ast.NewCollStatsStage(
				ast.NewCollStatsLatencyStats(true),
				nil,
				nil,
			),
			nil,
		},
		{
			`{"$collStats": { "storageStats": {}}}`,
			ast.NewCollStatsStage(
				nil,
				ast.NewCollStatsStorageStats(),
				nil,
			),
			nil,
		},
		{
			`{"$collStats": { "count": {}}}`,
			ast.NewCollStatsStage(
				nil,
				nil,
				ast.NewCollStatsCount(),
			),
			nil,
		},
		{
			`{"$collStats": { "latencyStats": {}, "storageStats": {}, "count": {}}}`,
			ast.NewCollStatsStage(
				ast.NewCollStatsLatencyStats(false),
				ast.NewCollStatsStorageStats(),
				ast.NewCollStatsCount(),
			),
			nil,
		},
		{
			`{"$collStats": 1}`,
			nil,
			errors.New("$collStats stage must have a document as its only argument"),
		},
		{
			`{"$collStats": { "latencyStats": "x" } }`,
			nil,
			errors.New("latencyStats argument must be an object, but got latencyStats: \"x\" of type string"),
		},
		{
			`{"$collStats": { "latencyStats": { "histograms": {} } } }`,
			nil,
			errors.New("histograms option to latencyStats must be bool"),
		},
		{
			`{"$collStats": { "storageStats": "x" } }`,
			nil,
			errors.New("storageStats argument must be an object, but got storageStats: \"x\" of type string"),
		},
		{
			`{"$collStats": { "count": "x" } }`,
			nil,
			errors.New("count argument must be an object, but got count: \"x\" of type string"),
		},
		{
			`{"$addFields": { "a": "$foo", "b": 1}}`,
			ast.NewAddFieldsStage(
				ast.NewAddFieldsItem("a", ast.NewFieldRef("foo", nil)),
				ast.NewAddFieldsItem("b", ast.NewConstant(bsonutil.Int32(1))),
			),
			nil,
		},
		{
			`{"$addFields": 1}`,
			nil,
			errors.New("$addFields stage must have a document as its only argument"),
		},
		{
			`{"$addFields": {}}`,
			nil,
			errors.New("$addFields specification must have at least one field"),
		},
		{
			`{"$replaceRoot": { "newRoot": { "a": "$foo" }}}`,
			ast.NewReplaceRootStage(
				ast.NewDocument(
					ast.NewDocumentElement("a", ast.NewFieldRef("foo", nil)),
				),
			),
			nil,
		},
		{
			`{"$replaceRoot": 1}`,
			nil,
			errors.New("$replaceRoot stage must have a document as its only argument"),
		},
		{
			`{"$replaceRoot": { "foo": 1 }}`,
			nil,
			errors.New("unrecognized option to $replaceRoot stage: foo, only valid option is 'newRoot'"),
		},
		{
			`{"$replaceRoot": { "newRoot": 0 }}`,
			nil,
			errors.New("'newRoot' expression must evaluate to an object"),
		},
		{
			`{"$replaceRoot": {}}`,
			nil,
			errors.New("no newRoot specified for the $replaceRoot stage"),
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
			nil,
		},
		{
			`{"$facet": 1}`,
			nil,
			errors.New("$facet stage must have a document as its only argument"),
		},
		{
			`{"$facet": { "foo": { "bar": 1 } } }`,
			nil,
			errors.New("arguments to $facet must be arrays, foo is type embedded document"),
		},
		{
			`{"$facet": {}}`,
			nil,
			errors.New("the $facet specification must be a non-empty object"),
		},
		{
			`{"$sortByCount": "$a"}`,
			ast.NewSortByCountStage(
				ast.NewFieldRef("a", nil),
			),
			nil,
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
			nil,
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
			nil,
		},
		{
			`{"$bucket": 1 }`,
			nil,
			errors.New("$bucket stage must have a document as its only argument"),
		},
		{
			`{"$bucket": { "groupBy": "$a", "boundaries": 1 } }`,
			nil,
			errors.New("$bucket 'boundaries' field must be an array, but found type: 32-bit integer"),
		},
		{
			`{"$bucket": { "groupBy": "$a", "boundaries": [0, 10, 20], "foo": 1 } }`,
			nil,
			errors.New("unrecognized option to $bucket: foo"),
		},
		{
			`{"$bucket": { "groupBy": "$a" } }`,
			nil,
			errors.New("$bucket requires 'groupBy' and 'boundaries' to be specified"),
		},
		{
			`{"$bucket": { "boundaries": [0, 10, 20] } }`,
			nil,
			errors.New("$bucket requires 'groupBy' and 'boundaries' to be specified"),
		},
		{
			`{"$bucket": { "groupBy": "$a", "boundaries": [0, 10, 20], "output": 1 } }`,
			nil,
			errors.New("$bucket 'output' field must be an object, but found type: 32-bit integer"),
		},
		{
			`{"$bucketAuto": { "groupBy": "$a", "buckets": 2 } }`,
			ast.NewBucketAutoStage(
				ast.NewFieldRef("a", nil),
				2,
				nil,
				"",
			),
			nil,
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
			nil,
		},
		{
			`{"$bucketAuto": 1}`,
			nil,
			errors.New("$bucketAuto stage must have a document as its only argument"),
		},
		{
			`{"$bucketAuto": { "groupBy": "$a", "buckets": "x" } }`,
			nil,
			errors.New("$bucketAuto 'buckets' field must be a numeric value, but found type: string"),
		},
		{
			`{"$bucketAuto": { "groupBy": "$a", "buckets": 0 } }`,
			nil,
			errors.New("$bucketAuto 'buckets' field must be greater than 0, but found: 0"),
		},
		{
			`{"$bucketAuto": { "groupBy": "$a", "buckets": -1 } }`,
			nil,
			errors.New("$bucketAuto 'buckets' field must be greater than 0, but found: -1"),
		},
		{
			`{"$bucketAuto": { "groupBy": "$a", "buckets": 2, "granularity": 1 } }`,
			nil,
			errors.New("$bucketAuto 'granularity' field must be a string, but found type: 32-bit integer"),
		},
		{
			`{"$bucketAuto": { "groupBy": "$a", "buckets": 2, "foo": 1 } }`,
			nil,
			errors.New("unrecognized option to $bucketAuto: foo"),
		},
		{
			`{"$bucketAuto": { "groupBy": "$a" } }`,
			nil,
			errors.New("$bucketAuto requires 'groupBy' and 'buckets' to be specified"),
		},
		{
			`{"$bucketAuto": { "buckets": 2 } }`,
			nil,
			errors.New("$bucketAuto requires 'groupBy' and 'buckets' to be specified"),
		},
		{
			`{"$bucketAuto": { "groupBy": "$a", "buckets": 2, "output": 1 } }`,
			nil,
			errors.New("$bucketAuto 'output' field must be an object, but found type: 32-bit integer"),
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
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual, err := parsertest.ParseStageErr(tc.input)

			if err != nil && tc.err == nil {
				t.Fatalf("err should be nil, but was %v", err)
			} else if err == nil && tc.err != nil {
				t.Fatalf("err should not be nil, expected %v", tc.err)
			} else if err != nil && tc.err != nil && err.Error() != tc.err.Error() {
				t.Fatalf("expected error %q, but got %q", tc.err.Error(), err.Error())
			}

			if tc.err == nil && !cmp.Equal(tc.expected, actual) {
				t.Fatalf("stages are not equal\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}
