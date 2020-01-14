package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// ParseExpr parses an expression.
func ParseExpr(v bsoncore.Value) (ast.Expr, error) {
	switch v.Type {
	case bsontype.Array:
		return parseArrayExpr(v.Array())
	case bsontype.Binary,
		bsontype.Boolean,
		bsontype.CodeWithScope,
		bsontype.DBPointer,
		bsontype.DateTime,
		bsontype.Decimal128,
		bsontype.Double,
		bsontype.Int32,
		bsontype.Int64,
		bsontype.JavaScript,
		bsontype.MinKey,
		bsontype.MaxKey,
		bsontype.Null,
		bsontype.ObjectID,
		bsontype.Regex,
		bsontype.Symbol,
		bsontype.Timestamp,
		bsontype.Undefined:
		return ast.NewConstant(v), nil
	case bsontype.String:
		s := v.StringValue()
		if strings.HasPrefix(s, "$$") {
			return parseVariableRef(s[2:])
		} else if strings.HasPrefix(s, "$") {
			return ParseFieldRef(s[1:])
		}

		return ast.NewConstant(v), nil
	case bsontype.EmbeddedDocument:
		return parseDocumentExpr(v.Document())
	}

	return nil, errors.New("unsupported expr")
}

// ParseExprJSON parses an ast.Expr from a string.
func ParseExprJSON(input string) (ast.Expr, error) {
	v, err := parseJSON(input)
	if err != nil {
		return nil, err
	}

	e, err := ParseExpr(v)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func parseArrayExpr(arr bsoncore.Document) (ast.Expr, error) {
	values, _ := arr.Values()
	elements := make([]ast.Expr, len(values))
	for i, v := range values {
		e, err := ParseExpr(v)
		if err != nil {
			return nil, err
		}

		elements[i] = e
	}

	return ast.NewArray(elements...), nil
}

func parseDocumentExpr(doc bsoncore.Document) (ast.Expr, error) {
	e, err := doc.IndexErr(0)
	if err != nil {
		return ast.NewDocument(), nil
	}

	key := e.Key()
	if strings.HasPrefix(key, "$") {
		return parseFunctionExpr(key, e.Value())
	}

	elems, _ := doc.Elements()
	documentElements := make([]*ast.DocumentElement, len(elems))
	for i, e := range elems {
		expr, err := ParseExpr(e.Value())
		if err != nil {
			return nil, err
		}

		documentElements[i] = ast.NewDocumentElement(e.Key(), expr)
	}

	return ast.NewDocument(documentElements...), nil
}

// ParseFieldRef parses a string into a FieldRef or FieldOrArrayIndexRef.
func ParseFieldRef(s string) (ast.FieldLikeRef, error) {
	parts := strings.Split(s, ".")
	var expr ast.FieldLikeRef = ast.NewFieldRef(parts[0], nil)
	for _, part := range parts[1:] {
		if len(part) == 0 {
			return nil, errors.New("invalid field ref")
		}
		index, err := strconv.Atoi(part)
		if err != nil {
			expr = ast.NewFieldRef(part, expr)
		} else {
			expr = ast.NewFieldOrArrayIndexRef(int32(index), expr)
		}
	}
	return expr, nil
}

func parseVariableRef(s string) (ast.Expr, error) {
	parts := strings.Split(s, ".")
	var expr ast.Expr = ast.NewVariableRef(parts[0])
	for _, part := range parts[1:] {
		expr = ast.NewFieldRef(part, expr)
	}
	return expr, nil
}

func parseFunctionExpr(key string, v bsoncore.Value) (ast.Expr, error) {
	switch key {
	case "$and", "$or", "$add", "$multiply", "$concat":
		arr, ok := v.ArrayOK()
		var values []bsoncore.Value
		if !ok {
			values = []bsoncore.Value{v}
		} else {
			values, _ = arr.Values()
		}
		return parseVarArgsExpr(ast.BinaryOp(key), values)
	case "$not", "$abs", "$ceil", "$floor", "$exp", "$ln", "$log10":
		var exprValue bsoncore.Value
		arr, ok := v.ArrayOK()
		if ok {
			arrValues, err := arr.Values()
			if err != nil {
				return nil, err
			}

			if len(arrValues) != 1 {
				return nil, errors.Errorf("expression %s takes exactly 1 arguments, %d were passed in", key, len(arrValues))
			}
			exprValue = arrValues[0]
		} else {
			exprValue = v
		}
		expr, err := ParseExpr(exprValue)
		if err != nil {
			return nil, err
		}
		return ast.NewUnary(ast.UnaryOp(key), expr), nil
	case "$trunc":
		return parseTruncExpr(v)
	case "$arrayElemAt":
		arr, ok := v.ArrayOK()
		if !ok {
			return nil, errors.Errorf("%s requires an array with 2 elements", key)
		}

		values, _ := arr.Values()
		if len(values) != 2 {
			return nil, errors.Errorf("%s requires an array with 2 elements", key)
		}

		parent, err := ParseExpr(values[0])
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing first element of $arrayElemAt")
		}

		index, err := ParseExpr(values[1])
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing second element of $arrayElemAt")
		}
		return ast.NewArrayIndexRef(index, parent), nil
	case "$cmp", "$eq", "$gt", "$gte", "$lt", "$lte", "$ne", "$divide", "$log", "$mod", "$pow", "$subtract":
		arr, ok := v.ArrayOK()
		if !ok {
			return nil, errors.Errorf("%s requires an array with 2 elements", key)
		}

		values, _ := arr.Values()
		if len(values) != 2 {
			return nil, errors.Errorf("%s requires an array with 2 elements", key)
		}

		return parseBinaryExpr(ast.BinaryOp(key), values)
	case "$let":
		return parseLetExpr(v)
	case "$cond":
		return parseConditionalExpr(v)
	case "$map":
		return parseMapExpr(v)
	case "$filter":
		return parseFilterExpr(v)
	case "$reduce":
		return parseReduceExpr(v)
	case "$literal":
		return ast.NewConstant(v), nil
	case "$mergeObjects":
		var elements []ast.Expr

		arr, ok := v.ArrayOK()
		if !ok {
			e, err := ParseExpr(v)
			if err != nil {
				return nil, err
			}
			elements = []ast.Expr{e}
		} else {
			values, _ := arr.Values()
			elements = make([]ast.Expr, len(values))
			for i, v := range values {
				e, err := ParseExpr(v)
				if err != nil {
					return nil, err
				}

				elements[i] = e
			}
		}

		return ast.NewMergeObjects(elements...), nil
	default:
		arg, err := ParseExpr(v)
		if err != nil {
			return nil, errors.Wrapf(err, "failed parsing first element of %s", key)
		}

		return ast.NewFunction(key, arg), nil
	}
}

func parseVarArgsExpr(op ast.BinaryOp, values []bsoncore.Value) (ast.Expr, error) {
	// Handle the special empty-values cases.
	if len(values) == 0 {
		switch op {
		case ast.And:
			return ast.NewConstant(bsonutil.Boolean(true)), nil
		case ast.Or:
			return ast.NewConstant(bsonutil.Boolean(false)), nil
		case ast.Add: // Add and Multiply will become the identity.
			return ast.NewConstant(bsonutil.Int32(0)), nil
		case ast.Multiply:
			return ast.NewConstant(bsonutil.Int32(1)), nil
		case ast.Concat:
			return ast.NewConstant(bsonutil.String("")), nil
		default:
			panic(fmt.Sprintf("no support for empty case of logical expr binary operator: %v", op))
		}
	}

	var resultExpr ast.Expr
	for i, v := range values {
		expr, err := ParseExpr(v)
		if err != nil {
			return nil, errors.Wrapf(err, "failed parsing expression %d of operator %s", i, op)
		}
		if resultExpr == nil {
			resultExpr = expr
		} else {
			resultExpr = ast.NewBinary(op, resultExpr, expr)
		}
	}

	return resultExpr, nil
}

func parseBinaryExpr(op ast.BinaryOp, values []bsoncore.Value) (ast.Expr, error) {
	left, err := ParseExpr(values[0])
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse left expr in %s", op)
	}
	right, err := ParseExpr(values[1])
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse right expr in %s", op)
	}

	return ast.NewBinary(op, left, right), nil
}

func parseTruncExpr(v bsoncore.Value) (ast.Expr, error) {
	var numberValue bsoncore.Value
	var err error

	// default precision value of 0
	var precision ast.Expr = ast.NewConstant(bsonutil.Int32(0))

	arr, ok := v.ArrayOK()
	if ok {
		arrValues, err := arr.Values()
		if err != nil {
			return nil, err
		}

		switch len(arrValues) {
		case 1:
			numberValue = arrValues[0]
		case 2:
			numberValue = arrValues[0]
			precision, err = ParseExpr(arrValues[1])
			if err != nil {
				return nil, errors.Wrap(err, "could not parse precision expr in $trunc")
			}
		default:
			return nil, errors.Errorf("expression $trunc takes at least 1 argument, and at most 2, but %d were passed in", len(arrValues))
		}
	} else {
		numberValue = v
	}

	number, err := ParseExpr(numberValue)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse number expr in $trunc")
	}

	return ast.NewTrunc(number, precision), nil
}

func parseLetExpr(v bsoncore.Value) (ast.Expr, error) {
	doc, ok := v.DocumentOK()
	if !ok {
		return nil, errors.New("$let requires a document")
	}

	var varsValue bsoncore.Value
	var exprValue bsoncore.Value
	varsFound := false
	exprFound := false
	elems, _ := doc.Elements()
	for _, e := range elems {
		switch e.Key() {
		case "vars":
			varsValue = e.Value()
			varsFound = true
		case "in":
			exprValue = e.Value()
			exprFound = true
		default:
			return nil, errors.Errorf("unrecognized parameter to $let: %s", e.Key())
		}
	}

	if !varsFound {
		return nil, errors.New("missing 'vars' parameter to $let")
	}

	if !exprFound {
		return nil, errors.New("missing 'in' parameter to $let")
	}

	varsDoc, ok := varsValue.DocumentOK()
	if !ok {
		return nil, errors.New("invalid parameter: expected an object (vars)")
	}

	varElems, _ := varsDoc.Elements()
	vars := make([]*ast.LetVariable, len(varElems))
	for i, e := range varElems {
		if err := validateVariableName(e.Key()); err != nil {
			return nil, err
		}
		expr, err := ParseExpr(e.Value())
		if err != nil {
			return nil, errors.Wrapf(err, "could not parse expr for variable %s", e.Key())
		}
		vars[i] = ast.NewLetVariable(e.Key(), expr)
	}

	expr, err := ParseExpr(exprValue)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse expr for $let in clause")
	}

	return ast.NewLet(vars, expr), nil
}

func parseConditionalExpr(v bsoncore.Value) (ast.Expr, error) {
	var err error
	var ifClause ast.Expr
	var thenClause ast.Expr
	var elseClause ast.Expr

	if doc, ok := v.DocumentOK(); ok {
		elems, _ := doc.Elements()
		for _, e := range elems {
			switch e.Key() {
			case "if":
				ifClause, err = ParseExpr(e.Value())
				if err != nil {
					return nil, errors.Wrap(err, "could not parse expr for $cond if clause")
				}
			case "then":
				thenClause, err = ParseExpr(e.Value())
				if err != nil {
					return nil, errors.Wrap(err, "could not parse expr for $cond then clause")
				}
			case "else":
				elseClause, err = ParseExpr(e.Value())
				if err != nil {
					return nil, errors.Wrap(err, "could not parse expr for $cond else clause")
				}
			default:
				return nil, errors.Errorf("unrecognized parameter to $cond: %s", e.Key())
			}
		}
	} else if arr, ok := v.ArrayOK(); ok {
		values, _ := arr.Values()
		if len(values) != 3 {
			return nil, errors.New("expression $cond takes exactly 3 arguments")
		}
		ifClause, err = ParseExpr(values[0])
		if err != nil {
			return nil, errors.Wrap(err, "could not parse expr for $cond if clause")
		}
		thenClause, err = ParseExpr(values[1])
		if err != nil {
			return nil, errors.Wrap(err, "could not parse expr for $cond then clause")
		}
		elseClause, err = ParseExpr(values[2])
		if err != nil {
			return nil, errors.Wrap(err, "could not parse expr for $cond else clause")
		}
	} else {
		return nil, errors.New("$cond requires a document or an array")
	}

	if ifClause == nil {
		return nil, errors.New("missing 'if' parameter to $cond")
	}

	if thenClause == nil {
		return nil, errors.New("missing 'then' parameter to $cond")
	}

	if elseClause == nil {
		return nil, errors.New("missing 'else' parameter to $cond")
	}

	return ast.NewConditional(ifClause, thenClause, elseClause), nil
}

func parseMapExpr(v bsoncore.Value) (ast.Expr, error) {
	var err error
	var inputClause ast.Expr
	asClause := "this"
	var inClause ast.Expr

	if doc, ok := v.DocumentOK(); ok {
		elems, _ := doc.Elements()
		for _, e := range elems {
			switch e.Key() {
			case "input":
				inputClause, err = ParseExpr(e.Value())
				if err != nil {
					return nil, errors.Wrap(err, "could not parse expr for $map input clause")
				}
			case "as":
				asClause, ok = e.Value().StringValueOK()
				if !ok {
					return nil, errors.New("$map as clause must be string")
				}
			case "in":
				inClause, err = ParseExpr(e.Value())
				if err != nil {
					return nil, errors.Wrap(err, "could not parse expr for $map in clause")
				}
			default:
				return nil, errors.Errorf("unrecognized parameter to $map: %s", e.Key())
			}
		}
	} else {
		return nil, errors.New("$map requires a document")
	}

	if inputClause == nil {
		return nil, errors.New("missing 'input' parameter to $map")
	}

	if inClause == nil {
		return nil, errors.New("missing 'in' parameter to $map")
	}

	return ast.NewMap(inputClause, asClause, inClause), nil
}

func parseFilterExpr(v bsoncore.Value) (ast.Expr, error) {
	var err error
	var inputClause ast.Expr
	asClause := "this"
	var condClause ast.Expr

	if doc, ok := v.DocumentOK(); ok {
		elems, _ := doc.Elements()
		for _, e := range elems {
			switch e.Key() {
			case "input":
				inputClause, err = ParseExpr(e.Value())
				if err != nil {
					return nil, errors.Wrap(err, "could not parse expr for $filter input clause")
				}
			case "as":
				asClause, ok = e.Value().StringValueOK()
				if !ok {
					return nil, errors.New("$filter as clause must be string")
				}
			case "cond":
				condClause, err = ParseExpr(e.Value())
				if err != nil {
					return nil, errors.Wrap(err, "could not parse expr for $filter cond clause")
				}
			default:
				return nil, errors.Errorf("unrecognized parameter to $filter: %s", e.Key())
			}
		}
	} else {
		return nil, errors.New("$filter requires a document")
	}

	if inputClause == nil {
		return nil, errors.New("missing 'input' parameter to $filter")
	}

	if condClause == nil {
		return nil, errors.New("missing 'cond' parameter to $filter")
	}

	return ast.NewFilter(inputClause, asClause, condClause), nil
}

func parseReduceExpr(v bsoncore.Value) (ast.Expr, error) {
	var err error
	var inputClause ast.Expr
	var initialValue ast.Expr
	var inClause ast.Expr

	if doc, ok := v.DocumentOK(); ok {
		elems, _ := doc.Elements()
		for _, e := range elems {
			switch e.Key() {
			case "input":
				inputClause, err = ParseExpr(e.Value())
				if err != nil {
					return nil, errors.Wrap(err, "could not parse expr for $reduce input clause")
				}
			case "initialValue":
				initialValue, err = ParseExpr(e.Value())
				if err != nil {
					return nil, errors.Wrap(err, "could not parse expr for $reduce initialValue clause")
				}
			case "in":
				inClause, err = ParseExpr(e.Value())
				if err != nil {
					return nil, errors.Wrap(err, "could not parse expr for $reduce in clause")
				}
			default:
				return nil, errors.Errorf("unrecognized parameter to $reduce: %s", e.Key())
			}
		}
	} else {
		return nil, errors.New("$reduce requires a document")
	}

	if inputClause == nil {
		return nil, errors.New("missing 'input' parameter to $reduce")
	}
	if initialValue == nil {
		return nil, errors.New("missing 'initialValue' parameter to $reduce")
	}
	if inClause == nil {
		return nil, errors.New("missing 'in' parameter to $reduce")
	}

	return ast.NewReduce(inputClause, initialValue, inClause), nil
}
