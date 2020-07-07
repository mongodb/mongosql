package parser

import (
	"fmt"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// shouldPreserveLiteral is a function telling us if this
// particular bsoncore.Value should preserve a user specified $literal
// outside of $project contexts:
// 1. If this is a string beginning with '$'
// 2. If this is a document or array.
func shouldPreserveLiteral(v bsoncore.Value) bool {
	switch v.Type {
	case bsontype.String:
		str := v.StringValue()
		return len(str) > 0 && str[0] == '$'
	case bsontype.EmbeddedDocument, bsontype.Array:
		// technically, we only need to preserve the literal if there is a key or string somewhere
		// within the Document with a '$' prefix, but this avoids having to recurse inside of the Document.
		// We treat Arrays the same because that is the only other type of value that may contain a
		// Document or string.
		return true
	}
	return false
}

// DeparseExpr turns an expression into a bson.Value suitable for use in a non-match aggregation stage.
func DeparseExpr(e ast.Expr, needsLiteral ...bool) bsoncore.Value {
	v, err := DeparseExprErr(e, needsLiteral...)
	if err != nil {
		panic(err)
	}
	return v
}

func DeparseExprErr(e ast.Expr, needsLiteral ...bool) (bsoncore.Value, error) {
	mustIncludeLiteral := len(needsLiteral) != 0 && needsLiteral[0]
	switch te := e.(type) {
	// Until we get a clear distinction between Match and AggExprs, this needs to
	// exist here or DeparseExpr can crash on an AggExpr.
	case *ast.AggExpr:
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsonutil.AppendValueElement(doc, "$expr", DeparseExpr(te.Expr, mustIncludeLiteral))
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.Array:
		values := make([]bsoncore.Value, len(te.Elements))
		for i, e := range te.Elements {
			values[i] = DeparseExpr(e, mustIncludeLiteral)
		}
		return bsonutil.ArrayFromValues(values...), nil
	case *ast.ArrayIndexRef:
		var parent bsoncore.Value
		var index bsoncore.Value
		if te.Parent != nil {
			parent = DeparseExpr(te.Parent, mustIncludeLiteral)
		} else {
			parent = bsonutil.Null()
		}
		if te.Index != nil {
			index = DeparseExpr(te.Index, mustIncludeLiteral)
		} else {
			index = bsonutil.Null()
		}

		_, arr := bsoncore.AppendArrayStart(nil)
		arr = bsonutil.AppendValueElement(arr, "0", parent)
		arr = bsonutil.AppendValueElement(arr, "1", index)
		arr, _ = bsoncore.AppendArrayEnd(arr, 0)

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendArrayElement(doc, "$arrayElemAt", arr)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.Binary:
		arr := bsonutil.ArrayFromValues(DeparseExpr(te.Left, mustIncludeLiteral), DeparseExpr(te.Right, mustIncludeLiteral))

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendArrayElement(doc, string(te.Op), arr.Data)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.Constant:
		if mustIncludeLiteral || shouldPreserveLiteral(te.Value) {
			_, doc := bsoncore.AppendDocumentStart(nil)
			doc = bsonutil.AppendValueElement(doc, "$literal", te.Value)
			doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
			return bsonutil.Document(doc), nil
		}
		return te.Value, nil
	case *ast.Convert:
		_, subdoc := bsoncore.AppendDocumentStart(nil)
		subdoc = bsonutil.AppendValueElement(subdoc, "input", DeparseExpr(te.Input, mustIncludeLiteral))
		subdoc = bsonutil.AppendValueElement(subdoc, "to", DeparseExpr(te.To, mustIncludeLiteral))
		if te.OnError != nil {
			subdoc = bsonutil.AppendValueElement(subdoc, "onError", DeparseExpr(te.OnError, mustIncludeLiteral))
		}
		if te.OnNull != nil {
			subdoc = bsonutil.AppendValueElement(subdoc, "onNull", DeparseExpr(te.OnNull, mustIncludeLiteral))
		}
		subdoc, _ = bsoncore.AppendDocumentEnd(subdoc, 0)

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendDocumentElement(doc, "$convert", subdoc)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.Document:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, i := range te.Elements {
			doc = bsonutil.AppendValueElement(doc, i.Name, DeparseExpr(i.Expr, mustIncludeLiteral))
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.FieldOrArrayIndexRef:
		return deparseFieldOrArrayIndexRef(te), nil
	case *ast.FieldRef:
		return deparseFieldRef(te), nil
	case *ast.Function:
		// Temporary hack until we add LiteralExpr fix, this is to handle
		// the format argument to $dateToString, which cannot be wrapped in
		// $literal in server 3.6. This is safe because $dateTo/FromString
		// were introduced in 3.6, so we do not need to worry about the
		// 3.2- not-using-$literal-in-project-expr-leaves issue. This
		// will be removed when we go to having LiteralExpr.
		var arg bsoncore.Value
		if te.Name == "$dateToString" || te.Name == "$dateFromString" {
			arg = DeparseExpr(te.Arg, false)
		} else {
			arg = DeparseExpr(te.Arg, mustIncludeLiteral)
		}
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsonutil.AppendValueElement(doc, te.Name, arg)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.Let:
		_, vdoc := bsoncore.AppendDocumentStart(nil)
		for _, variable := range te.Variables {
			vdoc = bsonutil.AppendValueElement(vdoc, variable.Name, DeparseExpr(variable.Expr, mustIncludeLiteral))
		}
		vdoc, _ = bsoncore.AppendDocumentEnd(vdoc, 0)

		_, subdoc := bsoncore.AppendDocumentStart(nil)
		subdoc = bsoncore.AppendDocumentElement(subdoc, "vars", vdoc)
		subdoc = bsonutil.AppendValueElement(subdoc, "in", DeparseExpr(te.Expr, mustIncludeLiteral))
		subdoc, _ = bsoncore.AppendDocumentEnd(subdoc, 0)

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendDocumentElement(doc, "$let", subdoc)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.Conditional:
		_, subdoc := bsoncore.AppendDocumentStart(nil)
		subdoc = bsonutil.AppendValueElement(subdoc, "if", DeparseExpr(te.If, mustIncludeLiteral))
		subdoc = bsonutil.AppendValueElement(subdoc, "then", DeparseExpr(te.Then, mustIncludeLiteral))
		subdoc = bsonutil.AppendValueElement(subdoc, "else", DeparseExpr(te.Else, mustIncludeLiteral))
		subdoc, _ = bsoncore.AppendDocumentEnd(subdoc, 0)

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendDocumentElement(doc, "$cond", subdoc)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.Unary:
		arr := bsonutil.ArrayFromValues(DeparseExpr(te.Expr))
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendArrayElement(doc, string(te.Op), arr.Data)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.Map:
		_, subdoc := bsoncore.AppendDocumentStart(nil)
		subdoc = bsonutil.AppendValueElement(subdoc, "input", DeparseExpr(te.Input, mustIncludeLiteral))
		subdoc = bsonutil.AppendValueElement(subdoc, "as", bsonutil.String(te.As))
		subdoc = bsonutil.AppendValueElement(subdoc, "in", DeparseExpr(te.In, mustIncludeLiteral))
		subdoc, _ = bsoncore.AppendDocumentEnd(subdoc, 0)

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendDocumentElement(doc, "$map", subdoc)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.Filter:
		_, subdoc := bsoncore.AppendDocumentStart(nil)
		subdoc = bsonutil.AppendValueElement(subdoc, "input", DeparseExpr(te.Input, mustIncludeLiteral))
		subdoc = bsonutil.AppendValueElement(subdoc, "as", bsonutil.String(te.As))
		subdoc = bsonutil.AppendValueElement(subdoc, "cond", DeparseExpr(te.Cond, mustIncludeLiteral))
		subdoc, _ = bsoncore.AppendDocumentEnd(subdoc, 0)

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendDocumentElement(doc, "$filter", subdoc)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.Reduce:
		_, subdoc := bsoncore.AppendDocumentStart(nil)
		subdoc = bsonutil.AppendValueElement(subdoc, "input", DeparseExpr(te.Input, mustIncludeLiteral))
		subdoc = bsonutil.AppendValueElement(subdoc, "initialValue", DeparseExpr(te.InitialValue, mustIncludeLiteral))
		subdoc = bsonutil.AppendValueElement(subdoc, "in", DeparseExpr(te.In, mustIncludeLiteral))
		subdoc, _ = bsoncore.AppendDocumentEnd(subdoc, 0)

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendDocumentElement(doc, "$reduce", subdoc)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.Unknown:
		return te.Value, nil
	case *ast.VariableRef:
		return bsoncore.Value{
			Type: bsontype.String,
			Data: bsoncore.AppendString(nil, "$$"+te.Name),
		}, nil
	// This cannot actually exist in an AggExpr, but until we get a clear distinction
	// between Match and AggExprs, this needs to exist here or DeparseExpr can crash.
	case *ast.MatchRegex:
		name, err := deparseMatchFieldName(te.Expr)
		if err != nil {
			return bsoncore.Value{}, err
		}

		_, subdoc := bsoncore.AppendDocumentStart(nil)
		subdoc = bsoncore.AppendStringElement(subdoc, "$regex", te.Pattern)
		subdoc = bsoncore.AppendStringElement(subdoc, "$options", te.Options)
		subdoc, _ = bsoncore.AppendDocumentEnd(subdoc, 0)

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendDocumentElement(doc, name, subdoc)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)

		return bsonutil.Document(doc), nil
	// This cannot actually exist in an AggExpr, but until we get a clear distinction
	// between Match and AggExprs, this needs to exist here or DeparseExpr can crash.
	case *ast.Exists:
		name, err := deparseMatchFieldName(te.Ref)
		if err != nil {
			return bsoncore.Value{}, err
		}

		_, subdoc := bsoncore.AppendDocumentStart(nil)
		subdoc = bsonutil.AppendValueElement(subdoc, "$exists", bsonutil.Boolean(te.Exists))
		subdoc, _ = bsoncore.AppendDocumentEnd(subdoc, 0)
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendDocumentElement(doc, name, subdoc)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)

		return bsonutil.Document(doc), nil
	case *ast.Trunc:
		includePrecision := true
		if c, ok := te.Precision.(*ast.Constant); ok {
			asInt32, isInt32 := bsonutil.AsInt32OK(c.Value)
			asInt64, isInt64 := bsonutil.AsInt64OK(c.Value)

			// do not include constant 0 precision
			includePrecision = (isInt32 && asInt32 != 0) || (isInt64 && asInt64 != 0)
		}

		var arr bsoncore.Value
		if includePrecision {
			arr = bsonutil.ArrayFromValues(
				DeparseExpr(te.Number, mustIncludeLiteral),
				DeparseExpr(te.Precision, mustIncludeLiteral),
			)
		} else {
			arr = bsonutil.ArrayFromValues(DeparseExpr(te.Number, mustIncludeLiteral))
		}

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendArrayElement(doc, "$trunc", arr.Data)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	}

	return bsoncore.Value{}, fmt.Errorf("unsupported expr %T", e)
}

func deparseFieldOrArrayIndexRef(e *ast.FieldOrArrayIndexRef) bsoncore.Value {
	return bsoncore.Value{
		Type: bsontype.String,
		Data: bsoncore.AppendString(nil, "$"+ast.GetDottedFieldName(e)),
	}
}

func deparseFieldRef(e *ast.FieldRef) bsoncore.Value {
	return bsoncore.Value{
		Type: bsontype.String,
		Data: bsoncore.AppendString(nil, "$"+ast.GetDottedFieldName(e)),
	}
}
