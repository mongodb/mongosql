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

		// If the returned expr is a ast.Binary and the op is not
		// the same as the key, it means we have found a $and, $nor,
		// or $or with one argument, which was not an original use
		// case for mongoast (it really only makes sense in the case of $nor,
		// as this is equivalent to negation). This fixes that case by
		// ensuring that we still wrap the Binary expr in the proper outter function.
		// e.g. {$nor: [{$and: [A,B]}]} was being replaced with simply
		// {$and: [A,B]}.
		if binOpExpr, ok := expr.(*ast.Binary); ok {
			if binOpExpr.Op != ast.BinaryOp(key) {
				expr = ast.NewFunction(key, ast.NewArray(binOpExpr))
			}
		} else {
			// If we arrive here it is because of an issue where some of the "BinaryOps":
			// specifically "$and", "$or", and "$nor", are actually varargs. Normally this would be
			// fine, because the code nests them: {$and: [a,b,c]} => {$and: [{$and: [a, b]}, c}}.
			// Where this fails is when one argument is passed, however. This should never really be
			// done for $and or $or, because it's sort of meaningless, (though it's valid, so should
			// be supported), but for $nor this becomes a real issue because it is the only good way
			// to negate a regular expression match in the match language ($nor with one argument is
			// equivalent to negation, which does not exist as a logical operator in the match
			// language).
			expr = ast.NewFunction(key, ast.NewArray(expr))
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
	if len(elems) == 2 {
		if elems[0].Key() == "$regex" {
			if elems[1].Key() != "$options" {
				return nil, errors.Wrapf(err, `the only viable argument to $regex is "$options" not "%s"`, elems[1].Key())
			}
			pattern := elems[0].Value()
			options := elems[1].Value()
			return ast.NewMatchRegex(e.Key(), pattern, options), nil
		}
	}
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
		case "$regex":
			return ast.NewMatchRegex(e.Key(), op.Value(), bsonutil.String("")), nil
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
