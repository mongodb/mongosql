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
	switch tn := n.(type) {
	case *ast.AddFieldsStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, item := range tn.Items {
			doc = bsonutil.AppendValueElement(doc, item.Name, DeparseExpr(item.Expr))
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$addFields", doc)
	case *ast.BucketStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsonutil.AppendValueElement(doc, "groupBy", DeparseExpr(tn.GroupBy))
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
				vdoc = bsonutil.AppendValueElement(vdoc, item.Name, DeparseExpr(item.Expr))
			}
			vdoc, _ = bsoncore.AppendDocumentEnd(vdoc, 0)
			doc = bsoncore.AppendDocumentElement(doc, "output", vdoc)
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$bucket", doc)
	case *ast.BucketAutoStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsonutil.AppendValueElement(doc, "groupBy", DeparseExpr(tn.GroupBy))
		doc = bsoncore.AppendInt64Element(doc, "buckets", tn.Buckets)
		if tn.Output != nil {
			_, vdoc := bsoncore.AppendDocumentStart(nil)
			for _, item := range tn.Output {
				vdoc = bsonutil.AppendValueElement(vdoc, item.Name, DeparseExpr(item.Expr))
			}
			vdoc, _ = bsoncore.AppendDocumentEnd(vdoc, 0)
			doc = bsoncore.AppendDocumentElement(doc, "output", vdoc)
		}
		if tn.Granularity != "" {
			doc = bsoncore.AppendStringElement(doc, "granularity", tn.Granularity)
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$bucketAuto", doc)
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
		return makeDocStage("$collStats", doc)
	case *ast.CountStage:
		return makeStringStage("$count", tn.FieldName)
	case *ast.FacetStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, item := range tn.Items {
			doc = bsonutil.AppendValueElement(doc, item.Name, DeparsePipeline(item.Pipeline))
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$facet", doc)
	case *ast.GroupStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		// We need $literal on the _id part of $group stages because
		// that is an error case for the server: $group does not support inclusion-style
		// expressions.
		// It isn't clear why this is the case, and is probably an error we should remove.
		doc = bsonutil.AppendValueElement(doc, "_id", DeparseExpr(tn.By, true))
		for _, i := range tn.Items {
			doc = bsonutil.AppendValueElement(doc, i.Name, DeparseExpr(i.Expr))
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$group", doc)
	case *ast.LimitStage:
		return makeInt64Stage("$limit", tn.Count)
	case *ast.LookupStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendStringElement(doc, "from", tn.From)
		if tn.LocalField != nil {
			localFieldName := ast.GetDottedFieldName(tn.LocalField)
			doc = bsoncore.AppendStringElement(doc, "localField", localFieldName)
		}
		if tn.ForeignField != "" {
			doc = bsoncore.AppendStringElement(doc, "foreignField", tn.ForeignField)
		}
		if tn.Let != nil {
			doc = bsoncore.AppendDocumentElement(doc, "let", deparseLookupLetItems(tn.Let))
		}
		if tn.Pipeline != nil {
			doc = bsonutil.AppendValueElement(doc, "pipeline", DeparsePipeline(tn.Pipeline))
		}
		doc = bsoncore.AppendStringElement(doc, "as", tn.As)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$lookup", doc)
	case *ast.MatchStage:
		return makeValueStage("$match", DeparseMatchExpr(tn.Expr))
	case *ast.ProjectStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, i := range tn.Items {
			switch ti := i.(type) {
			case *ast.IncludeProjectItem:
				doc = bsoncore.AppendInt32Element(doc, ast.GetDottedFieldName(ti.FieldRef), 1)
			case *ast.ExcludeProjectItem:
				doc = bsoncore.AppendInt32Element(doc, ast.GetDottedFieldName(ti.FieldRef), 0)
			case *ast.AssignProjectItem:
				// The true here ensures us that any constants will be wrapped in $literal, which
				// is necessary (at the top level) for mongo server 3.4+, and needed for all
				// constants at any level in versions 3.2-.
				doc = bsonutil.AppendValueElement(doc, ti.Name, DeparseExpr(ti.Expr, true))
			default:
				panic(fmt.Sprintf("unsupported project item %T", i))
			}
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$project", doc)
	case *ast.RedactStage:
		return makeValueStage("$redact", DeparseExpr(tn.Expr))
	case *ast.ReplaceRootStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsonutil.AppendValueElement(doc, "newRoot", DeparseExpr(tn.NewRoot))
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$replaceRoot", doc)
	case *ast.SampleStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendInt64Element(doc, "size", tn.Count)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$sample", doc)
	case *ast.SkipStage:
		return makeInt64Stage("$skip", tn.Count)
	case *ast.SortStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, i := range tn.Items {
			key := deparseMatchFieldName(i.Expr)
			if i.Descending {
				doc = bsoncore.AppendInt32Element(doc, key, -1)
			} else {
				doc = bsoncore.AppendInt32Element(doc, key, 1)
			}
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$sort", doc)
	case *ast.SortByCountStage:
		return makeValueStage("$sortByCount", DeparseExpr(tn.Expr))
	case *ast.SortedMergeStage:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, i := range tn.Items {
			key := deparseMatchFieldName(i.Expr)
			if i.Descending {
				doc = bsoncore.AppendInt32Element(doc, key, -1)
			} else {
				doc = bsoncore.AppendInt32Element(doc, key, 1)
			}
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$sortedMerge", doc)
	case *ast.Unknown:
		return tn.Value
	case *ast.UnwindStage:
		if tn.IncludeArrayIndex == "" && !tn.PreserveNullAndEmptyArrays {
			return makeValueStage("$unwind", DeparseExpr(tn.Path))
		}

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsonutil.AppendValueElement(doc, "path", DeparseExpr(tn.Path))
		if tn.IncludeArrayIndex != "" {
			doc = bsoncore.AppendStringElement(doc, "includeArrayIndex", tn.IncludeArrayIndex)
		}
		if tn.PreserveNullAndEmptyArrays {
			doc = bsoncore.AppendBooleanElement(doc, "preserveNullAndEmptyArrays", tn.PreserveNullAndEmptyArrays)
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return makeDocStage("$unwind", doc)
	}

	panic(fmt.Sprintf("unsupported node %T", n))
}

func deparseLookupLetItems(items []*ast.LookupLetItem) bsoncore.Document {
	_, doc := bsoncore.AppendDocumentStart(nil)
	for _, item := range items {
		doc = bsonutil.AppendValueElement(doc, item.Name, DeparseExpr(item.Expr))
	}
	doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
	return doc
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
