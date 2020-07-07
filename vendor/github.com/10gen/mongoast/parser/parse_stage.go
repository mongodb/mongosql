package parser

import (
	"fmt"
	"strings"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/10gen/mongoast/internal/bytesize"
	"github.com/10gen/mongoast/internal/decimalutil"

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
	case "$currentOp":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$currentOp stage must have a document as its only argument")
		}
		return parseCurrentOpStage(vdoc)
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
	case "$indexStats":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("the $indexStats stage specification must be an empty object")
		}
		elems, _ := vdoc.Elements()
		if len(elems) != 0 {
			return nil, errors.New("the $indexStats stage specification must be an empty object")
		}
		return ast.NewIndexStatsStage(), nil
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
	case "$out":
		collectionName, ok := e.Value().StringValueOK()
		if ok {
			return ast.NewOutStage(collectionName), nil
		}
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$out stage must have a string or a document as its only argument")
		}
		return parseOutStage(vdoc)
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
	case "$replaceWith":
		expr, err := ParseExpr(e.Value())
		if err != nil {
			return nil, errors.Wrap(err, "$replaceWith stage must have an expression as its only argument")
		}
		return parseReplaceRootStageFromExpr(expr)
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
	case "$sortByExpr":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$sortByExpr stage must have a document as its only argument")
		}
		return parseSortByExprStage(vdoc)
	case "$sortedMerge":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$sortedMerge stage must have a document as its only argument")
		}
		return parseSortedMergeStage(vdoc)
	case "$unset":
		return parseUnset(e.Value())
	case "$unwind":
		return parseUnwind(e.Value())
	case "$unionWith":
		return parseUnionWith(e.Value())

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
		default:
			return nil, errors.Errorf("unrecognized option to $collStats: %s", e.Key())
		}
	}

	return ast.NewCollStatsStage(latencyStats, storageStats, count), nil
}

func parseCurrentOpStage(doc bsoncore.Document) (*ast.CurrentOpStage, error) {
	var allUsers bool
	var idleConnections bool
	var idleCursors bool
	var idleSessions bool
	var localOps bool
	var debug bool
	var ok bool

	elems, _ := doc.Elements()
	for _, e := range elems {
		switch e.Key() {
		case "allUsers":
			allUsers, ok = e.Value().BooleanOK()
			if !ok {
				return nil, errors.Errorf(
					"the 'allUsers' parameter of the $currentOp stage must be a boolean value but found: %s",
					bsonutil.TypeToString(e.Value().Type),
				)
			}
		case "debug":
			debug, ok = e.Value().BooleanOK()
			if !ok {
				return nil, errors.Errorf(
					"the 'debug' parameter of the $currentOp stage must be a boolean value but found: %s",
					bsonutil.TypeToString(e.Value().Type),
				)
			}
		case "idleConnections":
			idleConnections, ok = e.Value().BooleanOK()
			if !ok {
				return nil, errors.Errorf(
					"the 'idleConnections' parameter of the $currentOp stage must be a boolean value but found: %s",
					bsonutil.TypeToString(e.Value().Type),
				)
			}
		case "idleCursors":
			idleCursors, ok = e.Value().BooleanOK()
			if !ok {
				return nil, errors.Errorf(
					"the 'idleCursors' parameter of the $currentOp stage must be a boolean value but found: %s",
					bsonutil.TypeToString(e.Value().Type),
				)
			}
		case "idleSessions":
			idleSessions, ok = e.Value().BooleanOK()
			if !ok {
				return nil, errors.Errorf(
					"the 'idleSessions' parameter of the $currentOp stage must be a boolean value but found: %s",
					bsonutil.TypeToString(e.Value().Type),
				)
			}
		case "localOps":
			localOps, ok = e.Value().BooleanOK()
			if !ok {
				return nil, errors.Errorf(
					"the 'localOps' parameter of the $currentOp stage must be a boolean value but found: %v",
					bsonutil.TypeToString(e.Value().Type),
				)
			}
		default:
			return nil, errors.Errorf(
				"unrecognized option '%s' in $currentOp stage",
				e.Key(),
			)
		}
	}

	return ast.NewCurrentOpStage(
		allUsers,
		idleConnections,
		idleCursors,
		idleSessions,
		localOps,
		debug,
	), nil
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
		if e.Key() == "" {
			return nil, errors.New("$facet field names may not be empty strings")
		}
		if strings.Contains(e.Key(), ".") {
			return nil, errors.New("$facet field names may not contain '.'")
		}

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
		if len(pipeline.Stages) == 0 {
			return nil, errors.New("sub-pipeline in $facet stage cannot be empty")
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
	var fromDB string
	var fromColl string
	var localField *ast.FieldRef
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
			switch e.Value().Type {
			case bsontype.String:
				fromColl = e.Value().StringValue()
			case bsontype.EmbeddedDocument:
				fromDB, fromColl, err = parseLookupFromDoc(e.Value().Document())
				if err != nil {
					return nil, err
				}
			default:
				return nil, errors.Errorf(
					"$lookup argument 'from: %v' must be a string or document, is type %v",
					e.Value(),
					e.Value().Type,
				)
			}
		case "localField":
			localFieldName, ok := e.Value().StringValueOK()
			if !ok {
				return nil, errors.Errorf(
					"$lookup argument 'localField: %v' must be a string, is type %v",
					e.Value(),
					e.Value().Type,
				)
			}
			if localFieldName != "" {
				localFieldExpr, err := ParseFieldRef(localFieldName)
				if err != nil {
					return nil, err
				}
				localField = localFieldExpr.(*ast.FieldRef)
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
			let, err = ParseLookupLetItems(vdoc)
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

	if fromColl == "" {
		return nil, errors.Errorf(
			"missing 'from' option to $lookup stage specification: %v",
			doc,
		)
	}
	if as == "" {
		return nil, errors.New("must specify 'as' field for a $lookup")
	}
	if pipeline == nil && (localField == nil || foreignField == "") {
		return nil, errors.New(
			"$lookup requires either 'pipeline' or both 'localField' and 'foreignField' to be specified",
		)
	}
	if pipeline != nil && (localField != nil || foreignField != "") {
		return nil, errors.New(
			"$lookup with 'pipeline' may not specify 'localField' or 'foreignField'",
		)
	}

	return ast.NewLookupStageWithDB(fromDB, fromColl, localField, foreignField, as, let, pipeline), nil
}

func parseLookupFromDoc(doc bsoncore.Document) (string, string, error) {
	var fromDB string
	var fromColl string
	var ok bool
	elems, _ := doc.Elements()
	for _, e := range elems {
		switch e.Key() {
		case "db":
			fromDB, ok = e.Value().StringValueOK()
			if !ok {
				return "", "", errors.Errorf(
					"$lookup argument 'from.db : %v' must be a string, is type %v",
					e.Value(),
					e.Value().Type,
				)
			}
		case "coll":
			fromColl, ok = e.Value().StringValueOK()
			if !ok {
				return "", "", errors.Errorf(
					"$lookup argument 'from.coll : %v' must be a string, is type %v",
					e.Value(),
					e.Value().Type,
				)
			}
		default:
			return "", "", errors.Errorf("invalid field in $lookup 'from' document: %s", e.Key())
		}
	}
	if fromColl == "" {
		return "", "", errors.Errorf("$lookup 'from' document must have a 'coll' field")
	}
	return fromDB, fromColl, nil
}

func ParseLookupLetItems(doc bsoncore.Document) ([]*ast.LookupLetItem, error) {
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

func parseOutStage(doc bsoncore.Document) (ast.Stage, error) {
	elems, _ := doc.Elements()
	if len(elems) != 1 {
		return nil, errors.New("$out stage must be a document with a single element")
	}

	e := elems[0]
	switch e.Key() {
	case "atlas":
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$out option 'atlas' must be a document")
		}
		return parseOutToAtlasStage(vdoc)
	case "s3":
		url, ok := e.Value().StringValueOK()
		if ok {
			return ast.NewOutToS3URLStage(url), nil
		}
		vdoc, ok := e.Value().DocumentOK()
		if !ok {
			return nil, errors.New("$out option 's3' must be a string or a document")
		}
		return parseOutToS3Stage(vdoc)
	default:
		return nil, errors.Errorf(
			"unrecognized option to $out stage: %s, valid options are 'atlas' and 's3'",
			e.Key(),
		)
	}
}

func parseOutToAtlasStage(doc bsoncore.Document) (*ast.OutStage, error) {
	var ok bool
	var projectID string
	var clusterName string
	var databaseName string
	var collectionName string

	elems, _ := doc.Elements()
	for _, e := range elems {
		switch e.Key() {
		case "projectId":
			projectID, ok = e.Value().StringValueOK()
			if !ok {
				return nil, errors.New("$out option 'atlas.projectId' must be a string")
			}
		case "clusterName":
			clusterName, ok = e.Value().StringValueOK()
			if !ok {
				return nil, errors.New("$out option 'atlas.clusterName' must be a string")
			}
		case "db":
			databaseName, ok = e.Value().StringValueOK()
			if !ok {
				return nil, errors.New("$out option 'atlas.db' must be a string")
			}
		case "coll":
			collectionName, ok = e.Value().StringValueOK()
			if !ok {
				return nil, errors.New("$out option 'atlas.coll' must be a string")
			}
		default:
			return nil, errors.Errorf("unrecognized option to $out stage: atlas.%s", e.Key())
		}
	}

	if clusterName == "" {
		return nil, errors.New("$out option 'atlas.clusterName' must be specified")
	}
	if databaseName == "" {
		return nil, errors.New("$out option 'atlas.db' must be specified")
	}
	if collectionName == "" {
		return nil, errors.New("$out option 'atlas.coll' must be specified")
	}

	return ast.NewOutToAtlasStage(projectID, clusterName, databaseName, collectionName), nil
}

func parseOutToS3Stage(doc bsoncore.Document) (*ast.OutStage, error) {
	var err error
	var ok bool
	var bucketName ast.Expr
	var prefixName ast.Expr
	var formatName string
	var regionName string
	var maxFileSize int64

	elems, _ := doc.Elements()
	for _, e := range elems {
		switch e.Key() {
		case "bucket":
			bucketName, err = ParseExpr(e.Value())
			if err != nil {
				return nil, errors.Wrap(err, "failed parsing $out option 's3.bucket'")
			}
		case "filename":
			prefixName, err = ParseExpr(e.Value())
			if err != nil {
				return nil, errors.Wrap(err, "failed parsing $out option 's3.filename'")
			}
		case "region":
			regionName, ok = e.Value().StringValueOK()
			if !ok {
				return nil, errors.New("$out option 's3.region' must be a string")
			}
		case "format":
			formatDoc, ok := e.Value().DocumentOK()
			if !ok {
				return nil, errors.New("$out option 's3.format' must be a document")
			}
			formatElems, _ := formatDoc.Elements()
			for _, fe := range formatElems {
				switch fe.Key() {
				case "name":
					formatName, ok = fe.Value().StringValueOK()
					if !ok {
						return nil, errors.New("$out option 's3.format.name' must be a string")
					}
				case "maxFileSize":
					maxFileSizeStr, ok := fe.Value().StringValueOK()
					if ok {
						maxFileSizeParsed, err := bytesize.Parse(maxFileSizeStr)
						if err != nil {
							return nil, errors.Wrap(err, "failed parsing $out option 's3.format.maxFileSize'")
						}
						maxFileSize = int64(maxFileSizeParsed)
					} else {
						maxFileSize, ok = bsonutil.AsInt64OK(fe.Value())
						if !ok {
							return nil, errors.New("$out option 's3.format.maxFileSize' must be a string or an integer")
						}
						if maxFileSize < 0 {
							return nil, errors.New("$out option 's3.format.maxFileSize' must not be negative")
						}
					}
				default:
					return nil, errors.Errorf("unrecognized option to $out stage: s3.format.%s", fe.Key())
				}
			}
		default:
			return nil, errors.Errorf("unrecognized option to $out stage: s3.%s", e.Key())
		}
	}

	if bucketName == nil {
		return nil, errors.New("$out option 's3.bucket' must be specified")
	}
	if prefixName == nil {
		return nil, errors.New("$out option 's3.filename' must be specified")
	}

	return ast.NewOutToS3Stage(bucketName, prefixName, regionName, formatName, maxFileSize), nil
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
	if projectStage.IsInclusion() {
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
			ref, err := ParseFieldRef(fullKey)
			if err != nil {
				return nil, errors.Wrapf(err, "failed parsing project field ref %s", fullKey)
			}

			exclude := false
			switch value.Type {
			case bsontype.Boolean:
				exclude = !value.Boolean()
			case bsontype.Decimal128:
				exclude = decimalutil.IsZero(decimalutil.FromPrimitive(value.Decimal128()))
			case bsontype.Double:
				exclude = value.Double() == 0.0
			case bsontype.Int32:
				exclude = value.Int32() == 0
			case bsontype.Int64:
				exclude = value.Int64() == 0
			}

			if exclude {
				items = append(items, ast.NewExcludeProjectItem(ref))
			} else {
				items = append(items, ast.NewIncludeProjectItem(ref))
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
				"unrecognized option to $replaceRoot stage: %s, only valid option is 'newRoot'",
				e.Key(),
			)
		}
	}

	if newRoot == nil {
		return nil, errors.New("no newRoot specified for the $replaceRoot stage")
	}

	return parseReplaceRootStageFromExpr(newRoot)
}

func parseReplaceRootStageFromExpr(newRoot ast.Expr) (*ast.ReplaceRootStage, error) {
	switch n := newRoot.(type) {
	case *ast.FieldRef,
		*ast.VariableRef,
		*ast.ArrayIndexRef,
		*ast.FieldOrArrayIndexRef,
		*ast.Document:
	case *ast.Function:
		switch n.Name {
		case "$mergeObjects":
		default:
			return nil, errors.New("'newRoot' expression must evaluate to an object")
		}
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
	items, err := parseSortItems(doc, true)
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

func parseSortByExprStage(doc bsoncore.Document) (*ast.SortByExprStage, error) {
	items, err := parseSortItems(doc, false)
	if err != nil {
		return nil, errors.Wrap(err, "failed parsing $sortByExpr items")
	}

	if len(items) == 0 {
		return nil, errors.New("$sortByExpr stage must have at least one sort key")
	}

	return ast.NewSortByExprStage(items...), nil
}

func parseSortedMergeStage(doc bsoncore.Document) (*ast.SortedMergeStage, error) {
	items, err := parseSortItems(doc, false)
	if err != nil {
		return nil, errors.Wrap(err, "failed parsing sort items")
	}

	return ast.NewSortedMergeStage(items...), nil
}

func parseSortItems(doc bsoncore.Document, requireFieldRefs bool) ([]*ast.SortItem, error) {
	elems, _ := doc.Elements()
	items := make([]*ast.SortItem, len(elems))
	for i, e := range elems {
		var expr ast.Expr
		var err error
		if requireFieldRefs {
			expr, err = ParseFieldRef(e.Key())
			if err != nil {
				return nil, errors.Wrapf(err, "failed parsing sort field ref %s", e.Key())
			}
		} else {
			expr, err = ParseExprJSON(e.Key())
			if err != nil {
				return nil, errors.Wrapf(err, "failed parsing sort expression %s", e.Key())
			}
		}

		descending, err := parseSortItemDescendingValue(e.Value())
		if err != nil {
			return nil, err
		}

		items[i] = ast.NewSortItem(expr, descending)
	}

	return items, nil
}

func parseSortItemDescendingValue(value bsoncore.Value) (bool, error) {
	compareDecimal := func(x decimalutil.Decimal128, y int32) int {
		return decimalutil.Compare(x, decimalutil.FromInt32(y))
	}

	switch value.Type {
	case bsontype.Decimal128:
		d128 := decimalutil.FromPrimitive(value.Decimal128())
		if compareDecimal(d128, -1) <= 0 && compareDecimal(d128, -2) > 0 {
			return true, nil
		} else if compareDecimal(d128, 1) < 0 || compareDecimal(d128, 2) >= 0 {
			return false, errors.New("sort key ordering must be 1 (for ascending) or -1 (for descending)")
		}
	case bsontype.Double:
		f64 := value.Double()
		if f64 <= -1.0 && f64 > -2.0 {
			return true, nil
		} else if f64 < 1.0 || f64 >= 2.0 {
			return false, errors.New("sort key ordering must be 1 (for ascending) or -1 (for descending)")
		}
	case bsontype.Int32:
		i32 := value.Int32()
		if i32 == -1 {
			return true, nil
		} else if i32 != 1 {
			return false, errors.New("sort key ordering must be 1 (for ascending) or -1 (for descending)")
		}
	case bsontype.Int64:
		i64 := value.Int64()
		if i64 == -1 {
			return true, nil
		} else if i64 != 1 {
			return false, errors.New("sort key ordering must be 1 (for ascending) or -1 (for descending)")
		}
	default:
		return false, fmt.Errorf("sort key ordering must be specified using a number")
	}

	return false, nil
}

func parseUnset(v bsoncore.Value) (*ast.ProjectStage, error) {
	switch v.Type {
	case bsontype.String:
		fieldName := v.StringValue()
		fieldRef, err := ParseFieldRef(fieldName)
		if err != nil {
			return nil, errors.Wrapf(err, "failed parsing project field ref %s", fieldName)
		}
		return ast.NewProjectStage(
			ast.NewExcludeProjectItem(fieldRef.(*ast.FieldRef)),
		), nil
	case bsontype.Array:
		vals, _ := v.Array().Values()
		if len(vals) == 0 {
			return nil, errors.New("$unset specification must be a string or an array with at least one field")
		}
		items := make([]ast.ProjectItem, len(vals))
		for i, v := range vals {
			fieldName, ok := v.StringValueOK()
			if !ok {
				return nil, errors.New("$unset specification must be a string or an array containing only string values")
			}
			fieldRef, err := ParseFieldRef(fieldName)
			if err != nil {
				return nil, errors.Wrapf(err, "failed parsing project field ref %s", fieldName)
			}
			items[i] = ast.NewExcludeProjectItem(fieldRef.(*ast.FieldRef))
		}
		return ast.NewProjectStage(items...), nil
	default:
		return nil, errors.New("$unset specification must be a string or an array")
	}
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
			default:
				return nil, errors.Errorf("unrecognized option to $unwind stage: %s", e.Key())
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

	isValidPathRef := true
	_, _ = ast.Visit(path, func(v ast.Visitor, n ast.Node) ast.Node {
		switch n.(type) {
		// Only these reference types are valid for the $unwind field path
		case *ast.ArrayIndexRef, *ast.FieldRef, *ast.FieldOrArrayIndexRef:
			_ = n.Walk(v)
		default:
			isValidPathRef = false
		}
		return n
	})

	if !isValidPathRef {
		return nil, errors.New("unwind field path must be a field reference")
	}

	return ast.NewUnwindStage(path.(ast.Ref), includeArrayIndex, preserveNullAndEmptyArrays), nil
}

func parseUnionWith(v bsoncore.Value) (*ast.UnionWithStage, error) {
	switch v.Type {
	case bsontype.String:
		coll := v.StringValue()
		return ast.NewUnionWithStage(coll, nil), nil
	case bsontype.EmbeddedDocument:
		vdoc := v.Document()
		elems, _ := vdoc.Elements()
		if len(elems) < 1 {
			return nil, errors.New("the $unionWith document must have at least one argument")
		}

		var coll string
		var err error
		pipeline := ast.NewPipeline()

		for _, e := range elems {
			switch e.Key() {
			case "coll":
				switch e.Value().Type {
				case bsontype.String:
					coll = elems[0].Value().StringValue()
				default:
					return nil, errors.Errorf(
						"$unionWith argument 'coll: %v' must be a string, it is type %v",
						e.Value(),
						e.Value().Type,
					)
				}
			case "pipeline":
				arr, ok := elems[1].Value().ArrayOK()
				if !ok {
					return nil, errors.New("'pipeline' option must be specified as an array")
				}
				pipeline, err = ParsePipeline(arr)
				if err != nil {
					return nil, err
				}
				for _, stage := range pipeline.Stages {
					if stage.StageName() == "$out" || stage.StageName() == "$merge" {
						return nil, errors.New("the pipeline inside $unionWith cannot include $out or $merge")
					}
				}
			default:
				return nil, errors.New("unknown argument to $unionWith")
			}
		}
		if coll == "" {
			return nil, errors.Errorf(
				"missing 'coll' option to $unionWith stage specification: %v",
				v.Document(),
			)
		}
		return ast.NewUnionWithStage(coll, pipeline), nil
	default:
		return nil, errors.New("invalid $unionWith argument(s)")
	}
}
