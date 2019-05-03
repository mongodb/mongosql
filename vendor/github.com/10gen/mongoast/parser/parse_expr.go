package parser

import (
	"strconv"
	"strings"

	"github.com/10gen/mongoast/ast"

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

func ParseFieldRef(s string) (ast.Expr, error) {
	parts := strings.Split(s, ".")
	var expr ast.Expr = ast.NewFieldRef(parts[0], nil)
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
	case "$and", "$or":
		arr, ok := v.ArrayOK()
		if !ok {
			return nil, errors.Errorf("%s requires an array", key)
		}

		values, _ := arr.Values()

		return parseLogicalExpr(ast.BinaryOp(key), values)
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
	case "$cmp", "$eq", "$gt", "$gte", "$lt", "$lte", "$ne":
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
	case "$literal":
		return ast.NewConstant(v), nil
	default:
		arg, err := ParseExpr(v)
		if err != nil {
			return nil, errors.Wrapf(err, "failed parsing first element of %s", key)
		}

		return ast.NewFunction(key, arg), nil
	}
}

func parseLogicalExpr(op ast.BinaryOp, values []bsoncore.Value) (ast.Expr, error) {
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
