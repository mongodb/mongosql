package parser

import (
	"fmt"
	"strconv"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"

	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// DeparseMatchExpr turns an ast.Expr into a bson.Value suitable for using in a $match stage.
func DeparseMatchExpr(e ast.Expr) bsoncore.Value {
	value := deparseMatchSubexpr(e)
	if b, ok := value.BooleanOK(); ok && b {
		return bsonutil.EmptyDocument()
	}
	return value
}

func deparseMatchSubexpr(e ast.Expr) bsoncore.Value {
	switch te := e.(type) {
	case *ast.AggExpr:
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsonutil.AppendValueElement(doc, "$expr", DeparseExpr(te.Expr))
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc)
	case *ast.Binary:
		switch te.Op {
		case ast.And, ast.Nor, ast.Or:
			arr := bsonutil.ArrayFromValues(flattenMatchExprBinary(te.Op, te.Left, te.Right)...)

			_, doc := bsoncore.AppendDocumentStart(nil)
			doc = bsoncore.AppendArrayElement(doc, string(te.Op), arr.Data)
			doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
			return bsonutil.Document(doc)
		default:
			name := deparseMatchFieldName(te.Left)

			_, subdoc := bsoncore.AppendDocumentStart(nil)
			subdoc = bsonutil.AppendValueElement(subdoc, string(te.Op), deparseMatchSubexpr(te.Right))
			subdoc, _ = bsoncore.AppendDocumentEnd(subdoc, 0)

			_, doc := bsoncore.AppendDocumentStart(nil)
			doc = bsoncore.AppendDocumentElement(doc, name, subdoc)
			doc, _ = bsoncore.AppendDocumentEnd(doc, 0)

			return bsonutil.Document(doc)
		}
	case *ast.Constant:
		return te.Value
	case *ast.Document:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, i := range te.Elements {
			doc = bsonutil.AppendValueElement(doc, i.Name, deparseMatchSubexpr(i.Expr))
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc)
	case *ast.Function:
		return deparseMatchExprFunction(te)
	case *ast.MatchRegex:
		name := te.Field

		_, subdoc := bsoncore.AppendDocumentStart(nil)
		subdoc = bsonutil.AppendValueElement(subdoc, "$regex", te.Pattern)
		subdoc = bsonutil.AppendValueElement(subdoc, "$options", te.Options)
		subdoc, _ = bsoncore.AppendDocumentEnd(subdoc, 0)

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendDocumentElement(doc, name, subdoc)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)

		return bsonutil.Document(doc)
	case *ast.Unknown:
		return te.Value
	}

	panic(fmt.Sprintf("unsupported expr %T", e))
}

func flattenMatchExprBinary(op ast.BinaryOp, left, right ast.Expr) []bsoncore.Value {
	values := make([]bsoncore.Value, 0)

	for _, e := range []ast.Expr{left, right} {
		if b, isBinary := e.(*ast.Binary); isBinary && b.Op == op {
			values = append(values, flattenMatchExprBinary(op, b.Left, b.Right)...)
			continue
		}

		values = append(values, deparseMatchSubexpr(e))
	}

	return values
}

func deparseMatchFieldName(e ast.Expr) string {
	switch te := e.(type) {
	case *ast.ArrayIndexRef:
		if ic, ok := te.Index.(*ast.Constant); ok {
			// TODO: 4.2 will have this feature, but it currently doesn't exist. Hence,
			// we'll render this exactly like FieldOrArrayIndexRef even though we have
			// parsed it correctly.
			index := int(bsonutil.AsInt32(ic.Value))
			if te.Parent != nil {
				return deparseMatchFieldName(te.Parent) + "." + strconv.Itoa(index)
			}

			return strconv.Itoa(index)
		}
	case *ast.FieldOrArrayIndexRef:
		if te.Parent != nil {
			return deparseMatchFieldName(te.Parent) + "." + strconv.Itoa(int(te.Number))
		}

		return strconv.Itoa(int(te.Number))
	case *ast.FieldRef:
		if te.Parent != nil {
			return deparseMatchFieldName(te.Parent) + "." + te.Name
		}

		return te.Name
	}

	panic(fmt.Sprintf("unsupported expr %T", e))
}

func deparseMatchExprFunction(f *ast.Function) bsoncore.Value {
	var arg bsoncore.Value
	switch ta := f.Arg.(type) {
	case *ast.Array:
		if len(ta.Elements) == 2 {
			if u, ok := ta.Elements[1].(*ast.Unknown); ok {
				// we are of the form "fieldName": { "$op": <value> }
				fieldName := deparseMatchFieldName(ta.Elements[0])

				_, opDoc := bsoncore.AppendDocumentStart(nil)
				opDoc = bsonutil.AppendValueElement(opDoc, f.Name, u.Value)
				opDoc, _ = bsoncore.AppendDocumentEnd(opDoc, 0)

				_, doc := bsoncore.AppendDocumentStart(nil)
				doc = bsoncore.AppendDocumentElement(doc, fieldName, opDoc)
				doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
				return bsonutil.Document(doc)
			}
		}

		// we are of the form { "function": [<arg>, ...] }
		values := make([]bsoncore.Value, len(ta.Elements))
		for i, e := range ta.Elements {
			values[i] = deparseMatchSubexpr(e)
		}
		arg = bsonutil.ArrayFromValues(values...)
	default:
		arg = deparseMatchSubexpr(f.Arg)
	}

	_, doc := bsoncore.AppendDocumentStart(nil)
	doc = bsonutil.AppendValueElement(doc, f.Name, arg)
	doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
	return bsonutil.Document(doc)
}
