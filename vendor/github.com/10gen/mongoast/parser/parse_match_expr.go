package parser

import (
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// ParseMatchExpr parses a match expression.
func ParseMatchExpr(doc bsoncore.Document) (ast.Expr, error) {
	var expr ast.Expr
	var err error

	elems, _ := doc.Elements()
	for _, e := range elems {
		var right ast.Expr
		right, err = parseMatchExprElement(e)

		if err != nil {
			return nil, err
		}

		if expr == nil {
			expr = right
		} else {
			expr = ast.NewBinary(ast.And, expr, right)
		}
	}

	if expr == nil {
		return ast.NewConstant(bsonutil.True), nil
	}
	return expr, nil
}

// ParseMatchExprJSON parses an ast.Expr from a string.
func ParseMatchExprJSON(input string) (ast.Expr, error) {
	v, err := parseJSON(input)
	if err != nil {
		return nil, err
	}

	doc, ok := v.DocumentOK()
	if !ok {
		return nil, errors.New("match expressions must be documents")
	}

	e, err := ParseMatchExpr(doc)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func parseMatchExprElement(e bsoncore.Element) (ast.Expr, error) {
	key := e.Key()
	if len(key) <= 0 {
		return nil, errors.New("invalid match expression key")
	}

	if key[0] == '$' {
		return parseNonFieldMatchExpr(e)
	}

	return parseFieldMatchExpr(e)
}

func parseNonFieldMatchExpr(e bsoncore.Element) (ast.Expr, error) {
	key := e.Key()
	switch key {
	case "$and", "$nor", "$or":
		arr, ok := e.Value().ArrayOK()
		if !ok {
			return nil, errors.Errorf("%s should have an array value", key)
		}

		values, _ := arr.Values()

		var expr ast.Expr
		for _, v := range values {
			partDoc, ok := v.DocumentOK()
			if !ok {
				return nil, errors.Errorf("%s array elements must be documents", key)
			}

			partExpr, err := ParseMatchExpr(partDoc)
			if err != nil {
				return nil, err
			}

			if expr == nil {
				expr = partExpr
			} else {
				expr = ast.NewBinary(ast.BinaryOp(key), expr, partExpr)
			}
		}

		return expr, nil
	case "$expr":
		expr, err := ParseExpr(e.Value())
		if err != nil {
			return nil, err
		}
		return ast.NewAggExpr(expr), nil
	default:
		return ast.NewFunction(e.Key(), ast.NewUnknown(e.Value())), nil
	}
}

func parseFieldMatchExpr(e bsoncore.Element) (ast.Expr, error) {
	left, err := ParseFieldRef(e.Key())
	if err != nil {
		return nil, errors.Wrapf(err, "failed parsing %s as a field ref", e.Key())
	}

	opDoc, ok := e.Value().DocumentOK()
	if !ok {
		right := ast.NewConstant(e.Value())
		return ast.NewBinary(ast.Equals, left, right), nil
	}

	var result ast.Expr
	elems, _ := opDoc.Elements()
	for _, op := range elems {
		key := op.Key()
		if len(key) <= 0 {
			return nil, errors.New("invalid match expression key")
		}

		if key[0] != '$' {
			right := ast.NewConstant(e.Value())
			return ast.NewBinary(ast.Equals, left, right), nil
		}

		var expr ast.Expr

		switch key {
		case "$eq":
			right := ast.NewConstant(op.Value())
			expr = ast.NewBinary(ast.Equals, left, right)
		case "$gt":
			right := ast.NewConstant(op.Value())
			expr = ast.NewBinary(ast.GreaterThan, left, right)
		case "$gte":
			right := ast.NewConstant(op.Value())
			expr = ast.NewBinary(ast.GreaterThanOrEquals, left, right)
		case "$lt":
			right := ast.NewConstant(op.Value())
			expr = ast.NewBinary(ast.LessThan, left, right)
		case "$lte":
			right := ast.NewConstant(op.Value())
			expr = ast.NewBinary(ast.LessThanOrEquals, left, right)
		case "$ne":
			right := ast.NewConstant(op.Value())
			expr = ast.NewBinary(ast.NotEquals, left, right)
		default:
			expr = ast.NewFunction(
				op.Key(),
				ast.NewArray(left, ast.NewUnknown(op.Value())),
			)
		}

		if result == nil {
			result = expr
		} else {
			result = ast.NewBinary(ast.And, result, expr)
		}
	}

	return result, nil
}
