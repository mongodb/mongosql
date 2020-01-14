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
	v, err := DeparseMatchExprErr(e)
	if err != nil {
		panic(err)
	}
	return v
}

func DeparseMatchExprErr(e ast.Expr) (bsoncore.Value, error) {
	value, err := deparseMatchSubexpr(e)
	if err != nil {
		return bsoncore.Value{}, err
	}
	if b, ok := value.BooleanOK(); ok && b {
		return bsonutil.EmptyDocument(), nil
	}
	return value, nil
}

func deparseMatchSubexpr(e ast.Expr) (bsoncore.Value, error) {
	switch te := e.(type) {
	case *ast.AggExpr:
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsonutil.AppendValueElement(doc, "$expr", DeparseExpr(te.Expr))
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.Binary:
		switch te.Op {
		case ast.And, ast.Or:
			values, err := flattenMatchExprBinary(te.Op, te.Left, te.Right)
			if err != nil {
				return bsoncore.Value{}, err
			}
			arr := bsonutil.ArrayFromValues(values...)

			_, doc := bsoncore.AppendDocumentStart(nil)
			doc = bsoncore.AppendArrayElement(doc, string(te.Op), arr.Data)
			doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
			return bsonutil.Document(doc), nil
		case ast.Nor:
			values, err := flattenNor(te.Left, te.Right)
			if err != nil {
				return bsoncore.Value{}, err
			}
			arr := bsonutil.ArrayFromValues(values...)

			_, doc := bsoncore.AppendDocumentStart(nil)
			doc = bsoncore.AppendArrayElement(doc, string(te.Op), arr.Data)
			doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
			return bsonutil.Document(doc), nil
		default:
			name, err := deparseMatchFieldName(te.Left)
			if err != nil {
				return bsoncore.Value{}, err
			}

			rightVal, err := deparseMatchSubexpr(te.Right)
			if err != nil {
				return bsoncore.Value{}, err
			}

			_, subdoc := bsoncore.AppendDocumentStart(nil)
			subdoc = bsonutil.AppendValueElement(subdoc, string(te.Op), rightVal)
			subdoc, _ = bsoncore.AppendDocumentEnd(subdoc, 0)

			_, doc := bsoncore.AppendDocumentStart(nil)
			doc = bsoncore.AppendDocumentElement(doc, name, subdoc)
			doc, _ = bsoncore.AppendDocumentEnd(doc, 0)

			return bsonutil.Document(doc), nil
		}
	case *ast.Unary:
		var subdoc bsoncore.Document
		var left ast.Expr
		switch tsub := te.Expr.(type) {
		case *ast.MatchRegex:
			left = tsub.Expr
			subdoc = deparseMatchRegex(tsub)

		case *ast.Binary:
			left = tsub.Left

			rightVal, err := deparseMatchSubexpr(tsub.Right)
			if err != nil {
				return bsoncore.Value{}, err
			}

			_, subdoc = bsoncore.AppendDocumentStart(nil)
			subdoc = bsonutil.AppendValueElement(subdoc, string(tsub.Op), rightVal)
			subdoc, _ = bsoncore.AppendDocumentEnd(subdoc, 0)

		default:
			v := DeparseNode(tsub)

			return bsoncore.Value{}, fmt.Errorf(
				"Unary operand must be either MatchRegex or Binary, but got %T: %v",
				tsub,
				v.String(),
			)
		}

		_, unarySubdoc := bsoncore.AppendDocumentStart(nil)
		unarySubdoc = bsoncore.AppendDocumentElement(unarySubdoc, string(te.Op), subdoc)
		unarySubdoc, _ = bsoncore.AppendDocumentEnd(unarySubdoc, 0)

		name, err := deparseMatchFieldName(left)
		if err != nil {
			return bsoncore.Value{}, err
		}
		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendDocumentElement(doc, name, unarySubdoc)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.Constant:
		return te.Value, nil
	case *ast.Document:
		_, doc := bsoncore.AppendDocumentStart(nil)
		for _, i := range te.Elements {
			exprVal, err := deparseMatchSubexpr(i.Expr)
			if err != nil {
				return bsoncore.Value{}, err
			}
			doc = bsonutil.AppendValueElement(doc, i.Name, exprVal)
		}
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
		return bsonutil.Document(doc), nil
	case *ast.Function:
		return deparseMatchExprFunction(te)
	case *ast.MatchRegex:
		name, err := deparseMatchFieldName(te.Expr)
		if err != nil {
			return bsoncore.Value{}, err
		}
		subdoc := deparseMatchRegex(te)

		_, doc := bsoncore.AppendDocumentStart(nil)
		doc = bsoncore.AppendDocumentElement(doc, name, subdoc)
		doc, _ = bsoncore.AppendDocumentEnd(doc, 0)

		return bsonutil.Document(doc), nil
	case *ast.Exists:
		name, err := deparseMatchFieldName(te.FieldRef)
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
	case *ast.Unknown:
		return te.Value, nil
	}

	return bsoncore.Value{}, fmt.Errorf("unsupported expr %T", e)
}

func deparseMatchRegex(expr *ast.MatchRegex) bsoncore.Document {
	_, doc := bsoncore.AppendDocumentStart(nil)
	doc = bsoncore.AppendStringElement(doc, "$regex", expr.Pattern)
	doc = bsoncore.AppendStringElement(doc, "$options", expr.Options)
	doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
	return doc
}

func flattenMatchExprBinary(op ast.BinaryOp, left, right ast.Expr) ([]bsoncore.Value, error) {
	values := make([]bsoncore.Value, 0)

	for _, e := range []ast.Expr{left, right} {
		if b, isBinary := e.(*ast.Binary); isBinary && b.Op == op {
			subvalues, err := flattenMatchExprBinary(op, b.Left, b.Right)
			if err != nil {
				return nil, err
			}
			values = append(values, subvalues...)
			continue
		}

		v, err := deparseMatchSubexpr(e)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}

	return values, nil
}

func flattenNor(left, right ast.Expr) ([]bsoncore.Value, error) {
	values := make([]bsoncore.Value, 0)
	for _, e := range []ast.Expr{left, right} {
		if b, isBinary := e.(*ast.Binary); isBinary && b.Op == ast.Or {
			subvalues, err := flattenNor(b.Left, b.Right)
			if err != nil {
				return nil, err
			}
			values = append(values, subvalues...)
		} else {
			v, err := deparseMatchSubexpr(e)
			if err != nil {
				return nil, err
			}
			values = append(values, v)
		}
	}

	return values, nil
}

func deparseMatchFieldName(e ast.Expr) (string, error) {
	switch te := e.(type) {
	case *ast.ArrayIndexRef:
		if ic, ok := te.Index.(*ast.Constant); ok {
			// TODO: 4.2 will have this feature, but it currently doesn't exist. Hence,
			// we'll render this exactly like FieldOrArrayIndexRef even though we have
			// parsed it correctly.
			index := int(bsonutil.AsInt32(ic.Value))
			if te.Parent != nil {
				parentName, err := deparseMatchFieldName(te.Parent)
				if err != nil {
					return "", err
				}
				return parentName + "." + strconv.Itoa(index), nil
			}

			return strconv.Itoa(index), nil
		}
	case *ast.FieldOrArrayIndexRef:
		if te.Parent != nil {
			parentName, err := deparseMatchFieldName(te.Parent)
			if err != nil {
				return "", err
			}
			return parentName + "." + strconv.Itoa(int(te.Number)), nil
		}

		return strconv.Itoa(int(te.Number)), nil
	case *ast.FieldRef:
		if te.Parent != nil {
			parentName, err := deparseMatchFieldName(te.Parent)
			if err != nil {
				return "", err
			}
			return parentName + "." + te.Name, nil
		}

		return te.Name, nil
	}

	return "", fmt.Errorf("unsupported expr %T", e)
}

func deparseMatchExprFunction(f *ast.Function) (bsoncore.Value, error) {
	var arg bsoncore.Value
	switch ta := f.Arg.(type) {
	case *ast.Array:
		if len(ta.Elements) == 2 {
			if u, ok := ta.Elements[1].(*ast.Unknown); ok {
				// we are of the form "fieldName": { "$op": <value> }
				fieldName, err := deparseMatchFieldName(ta.Elements[0])
				if err != nil {
					return bsoncore.Value{}, err
				}

				_, opDoc := bsoncore.AppendDocumentStart(nil)
				opDoc = bsonutil.AppendValueElement(opDoc, f.Name, u.Value)
				opDoc, _ = bsoncore.AppendDocumentEnd(opDoc, 0)

				_, doc := bsoncore.AppendDocumentStart(nil)
				doc = bsoncore.AppendDocumentElement(doc, fieldName, opDoc)
				doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
				return bsonutil.Document(doc), nil
			}
		}

		// we are of the form { "function": [<arg>, ...] }
		values := make([]bsoncore.Value, len(ta.Elements))
		for i, e := range ta.Elements {
			v, err := deparseMatchSubexpr(e)
			if err != nil {
				return bsoncore.Value{}, err
			}
			values[i] = v
		}
		arg = bsonutil.ArrayFromValues(values...)
	default:
		var err error
		arg, err = deparseMatchSubexpr(f.Arg)
		if err != nil {
			return bsoncore.Value{}, err
		}
	}

	_, doc := bsoncore.AppendDocumentStart(nil)
	doc = bsonutil.AppendValueElement(doc, f.Name, arg)
	doc, _ = bsoncore.AppendDocumentEnd(doc, 0)
	return bsonutil.Document(doc), nil
}
