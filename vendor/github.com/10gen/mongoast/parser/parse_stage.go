package parser

import (
	"fmt"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// ParseStage parses a stage.
func ParseStage(doc bsoncore.Document) (ast.Stage, error) {
	e, err := doc.IndexErr(0)
	if err != nil {
		return nil, errors.Wrap(err, "failed acquiring stage name")
	}

	switch e.Key() {
	case "$addFields":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$addFields stage must have a document as its only argument")
		}
		return parseAddFieldsStage(vdoc)
	case "$bucket":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$bucket stage must have a document as its only argument")
		}
		return parseBucketStage(vdoc)
	case "$bucketAuto":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$bucketAuto stage must have a document as its only argument")
		}
		return parseBucketAutoStage(vdoc)
	case "$collStats":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$collStats stage must have a document as its only argument")
		}
		return parseCollStatsStage(vdoc)
	case "$count":
		fieldName, ok := e.Value().StringValueOK()
		if !ok {
			return nil, errors.New("$count stage must have a string as its only argument")
		}
		return ast.NewCountStage(fieldName), nil
	case "$facet":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$facet stage must have a document as its only argument")
		}
		return parseFacetStage(vdoc)
	case "$group":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$group stage must have a document as its only argument")
		}
		return parseGroupStage(vdoc)
	case "$limit":
		i64, ok := bsonutil.AsInt64OK(e.Value())
		if !ok {
			return nil, errors.New("$limit stage must have an integer as its only argument")
		}
		if i64 <= 0 {
			return nil, errors.New("argument to $limit stage must be positive")
		}
		return ast.NewLimitStage(i64), nil
	case "$lookup":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$lookup stage must have a document as its only argument")
		}
		return parseLookupStage(vdoc)
	case "$match":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$match stage must have a document as its only argument")
		}
		return parseMatchStage(vdoc)
	case "$project":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$project stage must have a document as its only argument")
		}
		return parseProjectStage(vdoc)
	case "$redact":
		expr, err := ParseExpr(e.Value())
		if err != nil {
			return nil, errors.Wrap(err, "$redact stage must have an expression as its only argument")
		}
		return ast.NewRedactStage(expr), nil
	case "$replaceRoot":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$replaceRoot stage must have a document as its only argument")
		}
		return parseReplaceRootStage(vdoc)
	case "$sample":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$sample stage must have a document as its only argument")
		}
		return parseSampleStage(vdoc)
	case "$skip":
		i64, ok := bsonutil.AsInt64OK(e.Value())
		if !ok {
			return nil, errors.New("$skip stage must have an integer as its only argument")
		}
		if i64 < 0 {
			return nil, errors.New("argument to $skip cannot be negative")
		}
		return ast.NewSkipStage(i64), nil
	case "$sort":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$sort stage must have a document as its only argument")
		}
		return parseSortStage(vdoc)
	case "$sortByCount":
		return parseSortByCountStage(e.Value())
	case "$sortedMerge":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$sortedMerge stage must have a document as its only argument")
		}
		return parseSortMergeStage(vdoc)
	case "$unwind":
		return parseUnwind(e.Value())
	default:
		return ast.NewUnknown(bsoncore.Value{
			Type: bsontype.EmbeddedDocument,
			Data: doc,
		}), nil
	}
}

func parseAddFieldsStage(doc bsoncore.Document) (*ast.AddFieldsStage, error) {
	var items []*ast.AddFieldsItem

	elems, _ := doc.Elements()
	for _, e := range elems {
		expr, err := ParseExpr(e.Value())
		if err != nil {
			return nil, errors.Wrapf(err, "failed parsing %s of $addFields", e.Key())
		}
		items = append(items, ast.NewAddFieldsItem(e.Key(), expr))
	}

	if len(items) == 0 {
		return nil, errors.New("$addFields specification must have at least one field")
	}

	return ast.NewAddFieldsStage(items...), nil
}

func parseBucketStage(doc bsoncore.Document) (*ast.BucketStage, error) {
	var err error
	var groupBy ast.Expr
	var boundaries []bsoncore.Value
	var defaultID *bsoncore.Value
	var output []*ast.GroupItem

	elems, _ := doc.Elements()
	for _, e := range elems {
		switch e.Key() {
		case "groupBy":
			groupBy, err = ParseExpr(e.Value())
			if err != nil {
				return nil, err
			}
		case "boundaries":
			arr, ok := e.Value().ArrayOK()
			if !ok {
				return nil, errors.Errorf(
					"$bucket 'boundaries' field must be an array, but found type: %v",
					e.Value().Type,
				)
			}
			boundaries, _ = arr.Values()
		case "default":
			defaultValue := e.Value()
			defaultID = &defaultValue
		case "output":
			output, err = parseBucketOutput("$bucket", e.Value())
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.Errorf("unrecognized option to $bucket: %s", e.Key())
		}
	}

	if groupBy == nil || boundaries == nil {
		return nil, errors.New("$bucket requires 'groupBy' and 'boundaries' to be specified")
	}

	return ast.NewBucketStage(groupBy, boundaries, defaultID, output), nil
}

func parseBucketAutoStage(doc bsoncore.Document) (*ast.BucketAutoStage, error) {
	var err error
	var ok bool
	var groupBy ast.Expr
	var buckets int64
	var output []*ast.GroupItem
	var granularity string

	elems, _ := doc.Elements()
	for _, e := range elems {
		switch e.Key() {
		case "groupBy":
			groupBy, err = ParseExpr(e.Value())
			if err != nil {
				return nil, err
			}
		case "buckets":
			buckets, ok = bsonutil.AsInt64OK(e.Value())
			if !ok {
				return nil, errors.Errorf(
					"$bucketAuto 'buckets' field must be a numeric value, but found type: %v",
					e.Value().Type,
				)
			}
			if buckets <= 0 {
				return nil, errors.Errorf(
					"$bucketAuto 'buckets' field must be greater than 0, but found: %d",
					buckets,
				)
			}
		case "output":
			output, err = parseBucketOutput("$bucketAuto", e.Value())
			if err != nil {
				return nil, err
			}
		case "granularity":
			granularity, ok = e.Value().StringValueOK()
			if !ok {
				return nil, errors.Errorf(
					"$bucketAuto 'granularity' field must be a string, but found type: %v",
					e.Value().Type,
				)
			}
		default:
			return nil, errors.Errorf("unrecognized option to $bucketAuto: %s", e.Key())
		}
	}

	if groupBy == nil || buckets == 0 {
		return nil, errors.New("$bucketAuto requires 'groupBy' and 'buckets' to be specified")
	}

	return ast.NewBucketAutoStage(groupBy, buckets, output, granularity), nil
}

func parseBucketOutput(name string, v bsoncore.Value) ([]*ast.GroupItem, error) {
	doc, ok := v.DocumentOK()
	if !ok {
		return nil, errors.Errorf(
			"%s 'output' field must be an object, but found type: %v",
			name,
			v.Type,
		)
	}
	elems, _ := doc.Elements()
	output := make([]*ast.GroupItem, len(elems))
	for i, e := range elems {
		expr, err := ParseExpr(e.Value())
		if err != nil {
			return nil, err
		}
		output[i] = ast.NewGroupItem(e.Key(), expr)
	}
	return output, nil
}

func parseCollStatsStage(doc bsoncore.Document) (*ast.CollStatsStage, error) {
	var latencyStats *ast.CollStatsLatencyStats
	var storageStats *ast.CollStatsStorageStats
	var count *ast.CollStatsCount

	elems, _ := doc.Elements()
	for _, e := range elems {
		switch e.Key() {
		case "latencyStats":
			vdoc, ok := e.Value().DocumentOK()
			if !ok {
				return nil, errors.Errorf(
					"latencyStats argument must be an object, but got latencyStats: %v of type %v",
					e.Value(),
					e.Value().Type,
				)
			}

			var histograms bool
			histogramsValue, err := vdoc.LookupErr("histograms")
			if err == nil {
				histograms, ok = histogramsValue.BooleanOK()
				if !ok {
					return nil, errors.New("histograms option to latencyStats must be bool")
				}
			}
			latencyStats = ast.NewCollStatsLatencyStats(histograms)
		case "storageStats":
			if e.Value().Type != bsontype.EmbeddedDocument {
				return nil, errors.Errorf(
					"storageStats argument must be an object, but got storageStats: %v of type %v",
					e.Value(),
					e.Value().Type,
				)
			}
			storageStats = ast.NewCollStatsStorageStats()
		case "count":
			if e.Value().Type != bsontype.EmbeddedDocument {
				return nil, errors.Errorf(
					"count argument must be an object, but got count: %v of type %v",
					e.Value(),
					e.Value().Type,
				)
			}
			count = ast.NewCollStatsCount()
		}
	}

	return ast.NewCollStatsStage(latencyStats, storageStats, count), nil
}

// ParseStageJSON parses an ast.Stage from a string.
func ParseStageJSON(input string) (ast.Stage, error) {
	v, err := parseJSON(input)
	if err != nil {
		return nil, err
	}

	doc, ok := v.DocumentOK()
	if !ok {
		return nil, errors.New("stages must be documents")
	}

	s, err := ParseStage(doc)
	if err != nil {
		return nil, err
	}

	return s, err
}

func parseFacetStage(doc bsoncore.Document) (*ast.FacetStage, error) {
	elems, _ := doc.Elements()
	items := make([]*ast.FacetItem, len(elems))
	for i, e := range elems {
		arr, ok := e.Value().ArrayOK()
		if !ok {
			return nil, errors.Errorf(
				"arguments to $facet must be arrays, %s is type %v",
				e.Key(),
				e.Value().Type,
			)
		}

		pipeline, err := ParsePipeline(arr)
		if err != nil {
			return nil, err
		}

		items[i] = ast.NewFacetItem(e.Key(), pipeline)
	}

	if len(items) == 0 {
		return nil, errors.New("the $facet specification must be a non-empty object")
	}

	return ast.NewFacetStage(items...), nil
}

func parseGroupStage(doc bsoncore.Document) (*ast.GroupStage, error) {
	var by ast.Expr
	var items []*ast.GroupItem

	elems, _ := doc.Elements()
	for _, e := range elems {
		expr, err := ParseExpr(e.Value())
		if err != nil {
			return nil, errors.Wrapf(err, "failed parsing %s of $group", e.Key())
		}
		if e.Key() == "_id" {
			by = expr
		} else {
			items = append(items, ast.NewGroupItem(e.Key(), expr))
		}
	}

	if by == nil {
		return nil, errors.New("$group stage must have document with _id field")
	}

	return ast.NewGroupStage(by, items...), nil
}

func parseLookupStage(doc bsoncore.Document) (*ast.LookupStage, error) {
	var from string
	var localField string
	var foreignField string
	var as string
	var let []*ast.LookupLetItem
	var pipeline *ast.Pipeline
	var ok bool
	var err error

	elems, _ := doc.Elements()
	for _, e := range elems {
		switch e.Key() {
		case "from":
			from, ok = e.Value().StringValueOK()
			if !ok {
				return nil, errors.Errorf(
					"$lookup argument 'from: %v' must be a string, is type %v",
					e.Value(),
					e.Value().Type,
				)
			}
		case "localField":
			localField, ok = e.Value().StringValueOK()
			if !ok {
				return nil, errors.Errorf(
					"$lookup argument 'localField: %v' must be a string, is type %v",
					e.Value(),
					e.Value().Type,
				)
			}
		case "foreignField":
			foreignField, ok = e.Value().StringValueOK()
			if !ok {
				return nil, errors.Errorf(
					"$lookup argument 'foreignField: %v' must be a string, is type %v",
					e.Value(),
					e.Value().Type,
				)
			}
		case "as":
			as, ok = e.Value().StringValueOK()
			if !ok {
				return nil, errors.Errorf(
					"$lookup argument 'as: %v' must be a string, is type %v",
					e.Value(),
					e.Value().Type,
				)
			}
		case "let":
			vdoc, ok := e.Value().DocumentOK()
			if !ok {
				return nil, errors.Errorf(
					"$lookup argument 'let: %v' must be an object, is type %v",
					e.Value(),
					e.Value().Type,
				)
			}
			let, err = parseLookupLetItems(vdoc)
			if err != nil {
				return nil, err
			}
		case "pipeline":
			arr, ok := e.Value().ArrayOK()
			if !ok {
				return nil, errors.New("'pipeline' option must be specified as an array")
			}
			pipeline, err = ParsePipeline(arr)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.Errorf("unknown argument to $lookup: %s", e.Key())
		}
	}

	if from == "" {
		return nil, errors.Errorf(
			"missing 'from' option to $lookup stage specification: %v",
			doc,
		)
	}
	if as == "" {
		return nil, errors.New("must specify 'as' field for a $lookup")
	}
	if pipeline == nil && (localField == "" || foreignField == "") {
		return nil, errors.New(
			"$lookup requires either 'pipeline' or both 'localField' and 'foreignField' to be specified",
		)
	}
	if pipeline != nil && (localField != "" || foreignField != "") {
		return nil, errors.New(
			"$lookup with 'pipeline' may not specify 'localField' or 'foreignField'",
		)
	}

	return ast.NewLookupStage(from, localField, foreignField, as, let, pipeline), nil
}

func parseLookupLetItems(doc bsoncore.Document) ([]*ast.LookupLetItem, error) {
	elems, _ := doc.Elements()
	items := make([]*ast.LookupLetItem, len(elems))
	for i, e := range elems {
		if err := validateVariableName(e.Key()); err != nil {
			return nil, err
		}
		expr, err := ParseExpr(e.Value())
		if err != nil {
			return nil, err
		}
		items[i] = ast.NewLookupLetItem(e.Key(), expr)
	}
	return items, nil
}

func parseMatchStage(doc bsoncore.Document) (*ast.MatchStage, error) {
	expr, err := ParseMatchExpr(doc)
	if err != nil {
		return nil, errors.Wrap(err, "failed parsing $match stage")
	}

	return ast.NewMatchStage(expr), nil
}

func parseProjectStage(doc bsoncore.Document) (*ast.ProjectStage, error) {
	items, err := parseProjectStageItems(doc, "")
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, errors.New("$project must have at least one field")
	}

	projectStage := ast.NewProjectStage(items...)
	if len(projectStage.NonExcludeItems()) > 0 {
		excludeItems := projectStage.ExcludeItems()
		if len(excludeItems) > 1 {
			return nil, fmt.Errorf("cannot exclude fields other than '_id' in an inclusion projection")
		}
		if _, ok := excludeItems["_id"]; !ok && len(excludeItems) != 0 {
			return nil, fmt.Errorf("cannot exclude fields other than '_id' in an inclusion projection")
		}
	}
	return projectStage, nil
}

func parseProjectStageItems(doc bsoncore.Document, prefix string) ([]ast.ProjectItem, error) {
	var items []ast.ProjectItem

	elems, _ := doc.Elements()
	for _, e := range elems {
		fullKey := prefix + e.Key()
		value := e.Value()
		if value.IsNumber() || value.Type == bsontype.Boolean {
			expr, err := parseFieldRef(fullKey)
			if err != nil {
				return nil, errors.Wrapf(err, "failed parsing project field ref %s", fullKey)
			}

			exclude := false
			switch value.Type {
			case bsontype.Boolean:
				exclude = !value.Boolean()
			case bsontype.Double:
				exclude = value.Double() == 0.0
			case bsontype.Int32:
				exclude = value.Int32() == 0
			case bsontype.Int64:
				exclude = value.Int64() == 0
			}

			if exclude {
				items = append(items, ast.NewExcludeProjectItem(expr.(*ast.FieldRef)))
			} else {
				items = append(items, ast.NewIncludeProjectItem(expr.(*ast.FieldRef)))
			}
		} else {
			if isNestedDoc(value) {
				subItems, err := parseProjectStageItems(value.Document(), fullKey+".")
				if err != nil {
					return nil, err
				}
				items = append(items, subItems...)
			} else {
				expr, err := ParseExpr(value)
				if err != nil {
					return nil, errors.Wrapf(err, "failed parsing expression of key %q", fullKey)
				}
				items = append(items, ast.NewAssignProjectItem(fullKey, expr))
			}
		}
	}

	return items, nil
}

func isNestedDoc(value bsoncore.Value) bool {
	doc, ok := value.DocumentOK()
	if !ok {
		return false
	}

	elems, err := doc.Elements()
	if err != nil || len(elems) == 0 {
		return false
	}

	return elems[0].Key()[0] != '$'
}

func parseReplaceRootStage(doc bsoncore.Document) (*ast.ReplaceRootStage, error) {
	var err error
	var newRoot ast.Expr

	elems, _ := doc.Elements()
	for _, e := range elems {
		switch e.Key() {
		case "newRoot":
			newRoot, err = ParseExpr(e.Value())
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.Errorf(
				"unrecognized option to $replaceRoot stage: '%s', only valid option is 'newRoot'",
				e.Key(),
			)
		}
	}

	if newRoot == nil {
		return nil, errors.New("no newRoot specified for the $replaceRoot stage")
	}
	switch newRoot.(type) {
	case *ast.FieldRef, *ast.Document:
	default:
		return nil, errors.New("'newRoot' expression must evaluate to an object")
	}

	return ast.NewReplaceRootStage(newRoot), nil
}

func parseSampleStage(doc bsoncore.Document) (*ast.SampleStage, error) {
	var size *int64

	elems, _ := doc.Elements()
	for _, e := range elems {
		switch e.Key() {
		case "size":
			value, ok := bsonutil.AsInt64OK(e.Value())
			if !ok {
				return nil, errors.New("size argument to $sample must be a number")
			}
			if value < 0 {
				return nil, errors.New("size argument to $sample must not be negative")
			}
			size = &value
		default:
			return nil, errors.Errorf("unrecognized option to $sample: %s", e.Key())
		}
	}

	if size == nil {
		return nil, errors.New("$sample stage must specify a size")
	}

	return ast.NewSampleStage(*size), nil
}

func parseSortStage(doc bsoncore.Document) (*ast.SortStage, error) {
	items, err := parseSortItems(doc)
	if err != nil {
		return nil, errors.Wrap(err, "failed parsing sort items")
	}

	if len(items) == 0 {
		return nil, errors.New("$sort stage must have at least one sort key")
	}

	return ast.NewSortStage(items...), nil
}

func parseSortByCountStage(v bsoncore.Value) (*ast.SortByCountStage, error) {
	expr, err := ParseExpr(v)
	if err != nil {
		return nil, errors.Wrap(err, "failed parsing $sortByCount stage")
	}

	return ast.NewSortByCountStage(expr), nil
}

func parseSortMergeStage(doc bsoncore.Document) (*ast.SortedMergeStage, error) {
	items, err := parseSortItems(doc)
	if err != nil {
		return nil, errors.Wrap(err, "failed parsing sort items")
	}

	return ast.NewSortedMergeStage(items...), nil
}

func parseSortItems(doc bsoncore.Document) ([]*ast.SortItem, error) {

	elems, _ := doc.Elements()
	items := make([]*ast.SortItem, len(elems))
	for i, e := range elems {
		expr, err := parseMatchFieldRef(e.Key())
		if err != nil {
			return nil, errors.Wrapf(err, "failed parsing sort field ref %s", e.Key())
		}

		descending := false
		value := e.Value()
		switch value.Type {
		case bsontype.Double:
			f64 := value.Double()
			if f64 <= -1.0 && f64 > -2.0 {
				descending = true
			} else if f64 < 1.0 || f64 >= 2.0 {
				return nil, errors.New("$sort key ordering must be 1 (for ascending) or -1 (for descending)")
			}
		case bsontype.Int32:
			i32 := value.Int32()
			if i32 == -1 {
				descending = true
			} else if i32 != 1 {
				return nil, errors.New("$sort key ordering must be 1 (for ascending) or -1 (for descending)")
			}
		case bsontype.Int64:
			i64 := value.Int64()
			if i64 == -1 {
				descending = true
			} else if i64 != 1 {
				return nil, errors.New("$sort key ordering must be 1 (for ascending) or -1 (for descending)")
			}
		default:
			return nil, fmt.Errorf("$sort key ordering must be specified using a number")
		}

		items[i] = ast.NewSortItem(expr, descending)
	}

	return items, nil
}

func parseUnwind(v bsoncore.Value) (*ast.UnwindStage, error) {
	var err error
	var ok bool
	var path ast.Expr
	var includeArrayIndex string
	preserveNullAndEmptyArrays := false

	switch v.Type {
	case bsontype.String:
		path, err = ParseExpr(v)
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing unwind field path")
		}

	case bsontype.EmbeddedDocument:
		doc := v.Document()
		elems, _ := doc.Elements()
		for _, e := range elems {
			switch e.Key() {
			case "path":
				path, err = ParseExpr(e.Value())
				if err != nil {
					return nil, errors.Wrap(err, "failed parsing unwind field path")
				}
			case "includeArrayIndex":
				includeArrayIndex, ok = e.Value().StringValueOK()
				if !ok {
					return nil, errors.New("includeArrayIndex must be a string")
				}
			case "preserveNullAndEmptyArrays":
				preserveNullAndEmptyArrays, ok = e.Value().BooleanOK()
				if !ok {
					return nil, errors.New("preserveNullAndEmptyArrays must be a boolean")
				}
			}
		}

	default:
		return nil, errors.Errorf(
			"expected either a string or an object as specification for $unwind stage, got %v",
			v.Type,
		)
	}

	if path == nil {
		return nil, errors.New("no path specified to $unwind stage")
	}

	fieldRef, ok := path.(*ast.FieldRef)
	if !ok {
		return nil, errors.New("unwind field path must be a field reference")
	}

	return ast.NewUnwindStage(fieldRef, includeArrayIndex, preserveNullAndEmptyArrays), nil
}
