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
	mustIncludeLiteral := len(needsLiteral) != 0 && needsLiteral[0]
	switch te := e.(type) {
	case *ast.Array:
		values := make([]bsoncore.Value, len(te.Elements))
		for i, e := range te.Elements {
			values[i] = DeparseExpr(e, mustIncludeLiteral)
		}
		return bsonutil.ArrayFromValues(values...)
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
		return bsonutil.Document(doc)
	case *ast.Binary:
		arr := bsonutil.ArrayFromValues(DeparseExpr(te.Left, mustIncludeLiteral), DeparseExpr(te.Right, mustIncludeLiteral))

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendArrayElement(doc, string(te.Op), arr.Data)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc)
	case *ast.Constant:
		if mustIncludeLiteral || shouldPreserveLiteral(te.Value) {
			_, doc := bsoncore.AppendDocumentStart(nil)
			doc = bsonutil.AppendValueElement(doc, "$literal", te.Value)
			doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
			return bsonutil.Document(doc)
		}
		return te.Value
	case *ast.Document:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, i := range te.Elements {
			doc = bsonutil.AppendValueElement(doc, i.Name, DeparseExpr(i.Expr, mustIncludeLiteral))
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc)
	case *ast.FieldRef:
		return deparseFieldRef(te)
	case *ast.Function:
		arg := DeparseExpr(te.Arg, mustIncludeLiteral)
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsonutil.AppendValueElement(doc, te.Name, arg)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc)
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
		return bsonutil.Document(doc)
	case *ast.Conditional:
		_, subdoc := bsoncore.AppendDocumentStart(nil)
		subdoc = bsonutil.AppendValueElement(subdoc, "if", DeparseExpr(te.If, mustIncludeLiteral))
		subdoc = bsonutil.AppendValueElement(subdoc, "then", DeparseExpr(te.Then, mustIncludeLiteral))
		subdoc = bsonutil.AppendValueElement(subdoc, "else", DeparseExpr(te.Else, mustIncludeLiteral))
		subdoc, _ = bsoncore.AppendDocumentEnd(subdoc, 0)

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendDocumentElement(doc, "$cond", subdoc)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc)
	case *ast.Unknown:
		return te.Value
	case *ast.VariableRef:
		return bsoncore.Value{
			Type: bsontype.String,
			Data: bsoncore.AppendString(nil, "$$"+te.Name),
		}
	}

	panic(fmt.Sprintf("unsupported expr %T", e))
}

func deparseFieldRef(e *ast.FieldRef) bsoncore.Value {
	return bsoncore.Value{
		Type: bsontype.String,
		Data: bsoncore.AppendString(nil, "$"+ast.GetDottedFieldName(e)),
	}
}
