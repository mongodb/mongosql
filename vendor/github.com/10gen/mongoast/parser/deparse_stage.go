package parser

import (
	"fmt"
	"strconv"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"

	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// DeparseStage turns a stage into a bson.Value.
func DeparseStage(n ast.Stage) bsoncore.Value {
	v, err := DeparseStageErr(n)
	if err != nil {
		panic(err)
	}
	return v
}

func DeparseStageErr(n ast.Stage) (bsoncore.Value, error) {
	switch tn := n.(type) {
	case *ast.AddFieldsStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, item := range tn.Items {
			v, err := DeparseExprErr(item.Expr)
			if err != nil {
				return bsoncore.Value{}, err
			}
			doc = bsonutil.AppendValueElement(doc, item.Name, v)
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$addFields", doc), nil
	case *ast.BucketStage:
		v, err := DeparseExprErr(tn.GroupBy)
		if err != nil {
			return bsoncore.Value{}, err
		}
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsonutil.AppendValueElement(doc, "groupBy", v)
		_, arr := bsoncore.AppendArrayStart(nil)
		for i, v := range tn.Boundaries {
			arr = bsonutil.AppendValueElement(arr, strconv.Itoa(i), v)
		}
		arr, _ = bsoncore.AppendArrayEnd(arr, 0)
		doc = bsoncore.AppendArrayElement(doc, "boundaries", arr)
		if tn.Default != nil {
			doc = bsonutil.AppendValueElement(doc, "default", *tn.Default)
		}
		if tn.Output != nil {
			_, vdoc := bsoncore.AppendDocumentStart(nil)
			for _, item := range tn.Output {
				v, err := DeparseExprErr(item.Expr)
				if err != nil {
					return bsoncore.Value{}, err
				}
				vdoc = bsonutil.AppendValueElement(vdoc, item.Name, v)
			}
			vdoc, _ = bsoncore.AppendDocumentEnd(vdoc, 0)
			doc = bsoncore.AppendDocumentElement(doc, "output", vdoc)
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$bucket", doc), nil
	case *ast.BucketAutoStage:
		v, err := DeparseExprErr(tn.GroupBy)
		if err != nil {
			return bsoncore.Value{}, err
		}
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsonutil.AppendValueElement(doc, "groupBy", v)
		doc = bsoncore.AppendInt64Element(doc, "buckets", tn.Buckets)
		if tn.Output != nil {
			_, vdoc := bsoncore.AppendDocumentStart(nil)
			for _, item := range tn.Output {
				v, err := DeparseExprErr(item.Expr)
				if err != nil {
					return bsoncore.Value{}, err
				}
				vdoc = bsonutil.AppendValueElement(vdoc, item.Name, v)
			}
			vdoc, _ = bsoncore.AppendDocumentEnd(vdoc, 0)
			doc = bsoncore.AppendDocumentElement(doc, "output", vdoc)
		}
		if tn.Granularity != "" {
			doc = bsoncore.AppendStringElement(doc, "granularity", tn.Granularity)
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$bucketAuto", doc), nil
	case *ast.CollStatsStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		if tn.LatencyStats != nil {
			_, vdoc := bsoncore.AppendDocumentStart(nil)
			vdoc = bsoncore.AppendBooleanElement(vdoc, "histograms", tn.LatencyStats.Histograms)
			vdoc, _ = bsoncore.AppendDocumentEnd(vdoc, 0)
			doc = bsoncore.AppendDocumentElement(doc, "latencyStats", vdoc)
		}
		if tn.StorageStats != nil {
			_, vdoc := bsoncore.AppendDocumentStart(nil)
			vdoc, _ = bsoncore.AppendDocumentEnd(vdoc, 0)
			doc = bsoncore.AppendDocumentElement(doc, "storageStats", vdoc)
		}
		if tn.Count != nil {
			_, vdoc := bsoncore.AppendDocumentStart(nil)
			vdoc, _ = bsoncore.AppendDocumentEnd(vdoc, 0)
			doc = bsoncore.AppendDocumentElement(doc, "count", vdoc)
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$collStats", doc), nil
	case *ast.CountStage:
		return makeStringStage("$count", tn.FieldName), nil
	case *ast.CurrentOpStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		if tn.AllUsers {
			doc = bsoncore.AppendBooleanElement(doc, "allUsers", true)
		}
		if tn.IdleConnections {
			doc = bsoncore.AppendBooleanElement(doc, "idleConnections", true)
		}
		if tn.IdleCursors {
			doc = bsoncore.AppendBooleanElement(doc, "idleCursors", true)
		}
		if tn.IdleSessions {
			doc = bsoncore.AppendBooleanElement(doc, "idleSessions", true)
		}
		if tn.LocalOps {
			doc = bsoncore.AppendBooleanElement(doc, "localOps", true)
		}
		if tn.Debug {
			doc = bsoncore.AppendBooleanElement(doc, "debug", true)
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$currentOp", doc), nil
	case *ast.FacetStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, item := range tn.Items {
			v, err := DeparsePipelineErr(item.Pipeline)
			if err != nil {
				return bsoncore.Value{}, err
			}
			doc = bsonutil.AppendValueElement(doc, item.Name, v)
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$facet", doc), nil
	case *ast.GroupStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		// We need $literal on the _id part of $group stages because
		// that is an error case for the server: $group does not support inclusion-style
		// expressions.
		// It isn't clear why this is the case, and is probably an error we should remove.
		v, err := DeparseExprErr(tn.By, true)
		if err != nil {
			return bsoncore.Value{}, err
		}
		doc = bsonutil.AppendValueElement(doc, "_id", v)
		for _, i := range tn.Items {
			v, err := DeparseExprErr(i.Expr)
			if err != nil {
				return bsoncore.Value{}, err
			}
			doc = bsonutil.AppendValueElement(doc, i.Name, v)
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$group", doc), nil
	case *ast.IndexStatsStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$indexStats", doc), nil
	case *ast.LimitStage:
		return makeInt64Stage("$limit", tn.Count), nil
	case *ast.LookupStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		if tn.FromDB != "" {
			doc = bsoncore.AppendDocumentElement(doc, "from", deparseLookupFromDocument(tn.FromDB, tn.From))
		} else {
			doc = bsoncore.AppendStringElement(doc, "from", tn.From)
		}
		if tn.LocalField != nil {
			localFieldName := ast.GetDottedFieldName(tn.LocalField)
			doc = bsoncore.AppendStringElement(doc, "localField", localFieldName)
		}
		if tn.ForeignField != "" {
			doc = bsoncore.AppendStringElement(doc, "foreignField", tn.ForeignField)
		}
		if tn.Let != nil {
			doc = bsoncore.AppendDocumentElement(doc, "let", DeparseLookupLetItems(tn.Let))
		}
		if tn.Pipeline != nil {
			v, err := DeparsePipelineErr(tn.Pipeline)
			if err != nil {
				return bsoncore.Value{}, err
			}
			doc = bsonutil.AppendValueElement(doc, "pipeline", v)
		}
		doc = bsoncore.AppendStringElement(doc, "as", tn.As)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$lookup", doc), nil
	case *ast.MatchStage:
		v, err := DeparseMatchExprErr(tn.Expr)
		if err != nil {
			return bsoncore.Value{}, err
		}
		return makeValueStage("$match", v), nil
	case *ast.OutStage:
		if tn.Atlas != nil {
			_, vdoc := bsoncore.AppendDocumentStart(nil)
			if tn.Atlas.ProjectID != "" {
				vdoc = bsoncore.AppendStringElement(vdoc, "projectId", tn.Atlas.ProjectID)
			}
			vdoc = bsoncore.AppendStringElement(vdoc, "clusterName", tn.Atlas.ClusterName)
			vdoc = bsoncore.AppendStringElement(vdoc, "db", tn.Atlas.DatabaseName)
			vdoc = bsoncore.AppendStringElement(vdoc, "coll", tn.Atlas.CollectionName)
			vdoc, _ = bsoncore.AppendDocumentEnd(vdoc, 0)
			_, doc := bsoncore.AppendDocumentStart(nil)
			doc = bsoncore.AppendDocumentElement(doc, "atlas", vdoc)
			doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
			return makeDocStage("$out", doc), nil
		} else if tn.S3 != nil {
			_, vdoc := bsoncore.AppendDocumentStart(nil)
			vdoc = bsonutil.AppendValueElement(vdoc, "bucket", DeparseExpr(tn.S3.Bucket))
			vdoc = bsonutil.AppendValueElement(vdoc, "filename", DeparseExpr(tn.S3.Filename))
			if tn.S3.Region != "" {
				vdoc = bsoncore.AppendStringElement(vdoc, "region", tn.S3.Region)
			}
			if tn.S3.Format != "" || tn.S3.MaxFileSizeBytes != 0 {
				_, fdoc := bsoncore.AppendDocumentStart(nil)
				if tn.S3.Format != "" {
					fdoc = bsoncore.AppendStringElement(fdoc, "name", tn.S3.Format)
				}
				if tn.S3.MaxFileSizeBytes != 0 {
					fdoc = bsoncore.AppendInt64Element(fdoc, "maxFileSize", tn.S3.MaxFileSizeBytes)
				}
				fdoc, _ = bsoncore.AppendDocumentEnd(fdoc, 0)
				vdoc = bsoncore.AppendDocumentElement(vdoc, "format", fdoc)
			}
			vdoc, _ = bsoncore.AppendDocumentEnd(vdoc, 0)
			_, doc := bsoncore.AppendDocumentStart(nil)
			doc = bsoncore.AppendDocumentElement(doc, "s3", vdoc)
			doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
			return makeDocStage("$out", doc), nil
		} else if tn.S3URL != "" {
			_, doc := bsoncore.AppendDocumentStart(nil)
			doc = bsoncore.AppendStringElement(doc, "s3", tn.S3URL)
			doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
			return makeDocStage("$out", doc), nil
		}
		return makeStringStage("$out", tn.CollectionName), nil
	case *ast.ProjectStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, i := range tn.Items {
			switch ti := i.(type) {
			case *ast.IncludeProjectItem:
				doc = bsoncore.AppendInt32Element(doc, ast.GetDottedFieldName(ti.Ref), 1)
			case *ast.ExcludeProjectItem:
				doc = bsoncore.AppendInt32Element(doc, ast.GetDottedFieldName(ti.Ref), 0)
			case *ast.AssignProjectItem:
				// The true here ensures us that any constants will be wrapped in $literal, which
				// is necessary (at the top level) for mongo server 3.4+, and needed for all
				// constants at any level in versions 3.2-.
				v, err := DeparseExprErr(ti.Expr, true)
				if err != nil {
					return bsoncore.Value{}, err
				}
				doc = bsonutil.AppendValueElement(doc, ti.Name, v)
			default:
				return bsoncore.Value{}, fmt.Errorf("unsupported project item %T", i)
			}
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$project", doc), nil
	case *ast.RedactStage:
		v, err := DeparseExprErr(tn.Expr)
		if err != nil {
			return bsoncore.Value{}, err
		}
		return makeValueStage("$redact", v), nil
	case *ast.ReplaceRootStage:
		v, err := DeparseExprErr(tn.NewRoot)
		if err != nil {
			return bsoncore.Value{}, err
		}
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsonutil.AppendValueElement(doc, "newRoot", v)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$replaceRoot", doc), nil
	case *ast.SampleStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendInt64Element(doc, "size", tn.Count)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$sample", doc), nil
	case *ast.SkipStage:
		return makeInt64Stage("$skip", tn.Count), nil
	case *ast.SortStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, i := range tn.Items {
			key, err := deparseMatchFieldName(i.Expr)
			if err != nil {
				return bsoncore.Value{}, err
			}
			if i.Descending {
				doc = bsoncore.AppendInt32Element(doc, key, -1)
			} else {
				doc = bsoncore.AppendInt32Element(doc, key, 1)
			}
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$sort", doc), nil
	case *ast.SortByCountStage:
		v, err := DeparseExprErr(tn.Expr)
		if err != nil {
			return bsoncore.Value{}, err
		}
		return makeValueStage("$sortByCount", v), nil
	case *ast.SortByExprStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, i := range tn.Items {
			key, err := DeparseExprErr(i.Expr)
			if err != nil {
				return bsoncore.Value{}, err
			}
			if i.Descending {
				doc = bsoncore.AppendInt32Element(doc, key.String(), -1)
			} else {
				doc = bsoncore.AppendInt32Element(doc, key.String(), 1)
			}
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$sortByExpr", doc), nil
	case *ast.SortedMergeStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, i := range tn.Items {
			key, err := DeparseExprErr(i.Expr)
			if err != nil {
				return bsoncore.Value{}, err
			}
			if i.Descending {
				doc = bsoncore.AppendInt32Element(doc, key.String(), -1)
			} else {
				doc = bsoncore.AppendInt32Element(doc, key.String(), 1)
			}
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$sortedMerge", doc), nil
	case *ast.Unknown:
		return tn.Value, nil
	case *ast.UnwindStage:
		v, err := DeparseExprErr(tn.Path)
		if err != nil {
			return bsoncore.Value{}, err
		}

		if tn.IncludeArrayIndex == "" && !tn.PreserveNullAndEmptyArrays {
			return makeValueStage("$unwind", v), nil
		}

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsonutil.AppendValueElement(doc, "path", v)
		if tn.IncludeArrayIndex != "" {
			doc = bsoncore.AppendStringElement(doc, "includeArrayIndex", tn.IncludeArrayIndex)
		}
		if tn.PreserveNullAndEmptyArrays {
			doc = bsoncore.AppendBooleanElement(doc, "preserveNullAndEmptyArrays", tn.PreserveNullAndEmptyArrays)
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$unwind", doc), nil
	case *ast.UnionWithStage:
		if tn.Pipeline != nil {
			_, doc := bsoncore.AppendDocumentStart(nil)
			doc = bsoncore.AppendStringElement(doc, "coll", tn.Coll)

			v, err := DeparsePipelineErr(tn.Pipeline)
			if err != nil {
				return bsoncore.Value{}, err
			}
			doc = bsonutil.AppendValueElement(doc, "pipeline", v)
			doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
			return makeDocStage("$unionWith", doc), nil
		} else {
			return makeStringStage("$unionWith", tn.Coll), nil
		}
	}

	return bsoncore.Value{}, fmt.Errorf("unsupported node %T", n)
}

// This function must only be called if db is non-empty.
func deparseLookupFromDocument(db, coll string) bsoncore.Document {
	_, doc := bsoncore.AppendDocumentStart(nil)
	doc = bsoncore.AppendStringElement(doc, "db", db)
	doc = bsoncore.AppendStringElement(doc, "coll", coll)
	doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
	return doc
}

func DeparseLookupLetItems(items []*ast.LookupLetItem) bsoncore.Document {
	doc, err := DeparseLookupLetItemsErr(items)
	if err != nil {
		panic(err)
	}
	return doc
}

func DeparseLookupLetItemsErr(items []*ast.LookupLetItem) (bsoncore.Document, error) {
	_, doc := bsoncore.AppendDocumentStart(nil)
	for _, item := range items {
		v, err := DeparseExprErr(item.Expr)
		if err != nil {
			return nil, err
		}
		doc = bsonutil.AppendValueElement(doc, item.Name, v)
	}
	doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
	return doc, nil
}

func makeDocStage(name string, subdoc bsoncore.Document) bsoncore.Value {
	_, doc := bsoncore.AppendDocumentStart(nil)
	doc = bsoncore.AppendDocumentElement(doc, name, subdoc)
	doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
	return bsonutil.Document(doc)
}

func makeInt64Stage(name string, i64 int64) bsoncore.Value {
	_, doc := bsoncore.AppendDocumentStart(nil)
	doc = bsoncore.AppendInt64Element(doc, name, i64)
	doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
	return bsonutil.Document(doc)
}

func makeStringStage(name string, s string) bsoncore.Value {
	_, doc := bsoncore.AppendDocumentStart(nil)
	doc = bsoncore.AppendStringElement(doc, name, s)
	doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
	return bsonutil.Document(doc)
}

func makeValueStage(name string, v bsoncore.Value) bsoncore.Value {
	_, doc := bsoncore.AppendDocumentStart(nil)
	doc = bsonutil.AppendValueElement(doc, name, v)
	doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
	return bsonutil.Document(doc)
}
