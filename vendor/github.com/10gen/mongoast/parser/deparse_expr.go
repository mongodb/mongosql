package parser

import (
	"fmt"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// DeparseExpr turns an expression into a bson.Value suitable for use in a non-match aggregation stage.
func DeparseExpr(e ast.Expr) bsoncore.Value {
	switch te := e.(type) {
	case *ast.Array:
		values := make([]bsoncore.Value, len(te.Elements))
		for i, e := range te.Elements {
			values[i] = DeparseExpr(e)
		}
		return bsonutil.ArrayFromValues(values...)

	case *ast.ArrayIndexRef:
		var parent bsoncore.Value
		var index bsoncore.Value
		if te.Parent != nil {
			parent = DeparseExpr(te.Parent)
		} else {
			parent = bsonutil.Null()
		}
		if te.Index != nil {
			index = DeparseExpr(te.Index)
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
		arr := bsonutil.ArrayFromValues(DeparseExpr(te.Left), DeparseExpr(te.Right))

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendArrayElement(doc, string(te.Op), arr.Data)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc)
	case *ast.Constant:
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsonutil.AppendValueElement(doc, "$literal", te.Value)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc)
	case *ast.Document:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, i := range te.Elements {
			doc = bsonutil.AppendValueElement(doc, i.Name, DeparseExpr(i.Expr))
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc)
	case *ast.FieldRef:
		return deparseFieldRef(te)
	case *ast.Function:
		arg := DeparseExpr(te.Arg)
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsonutil.AppendValueElement(doc, te.Name, arg)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc)
	case *ast.Let:
		_, vdoc := bsoncore.AppendDocumentStart(nil)
		for _, variable := range te.Variables {
			vdoc = bsonutil.AppendValueElement(vdoc, variable.Name, DeparseExpr(variable.Expr))
		}
		vdoc, _ = bsoncore.AppendDocumentEnd(vdoc, 0)

		_, subdoc := bsoncore.AppendDocumentStart(nil)
		subdoc = bsoncore.AppendDocumentElement(subdoc, "vars", vdoc)
		subdoc = bsonutil.AppendValueElement(subdoc, "in", DeparseExpr(te.Expr))
		subdoc, _ = bsoncore.AppendDocumentEnd(subdoc, 0)

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendDocumentElement(doc, "$let", subdoc)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc)
	case *ast.Conditional:
		_, subdoc := bsoncore.AppendDocumentStart(nil)
		subdoc = bsonutil.AppendValueElement(subdoc, "if", DeparseExpr(te.If))
		subdoc = bsonutil.AppendValueElement(subdoc, "then", DeparseExpr(te.Then))
		subdoc = bsonutil.AppendValueElement(subdoc, "else", DeparseExpr(te.Else))
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
