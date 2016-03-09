package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/schema"
)

// SQLAddExpr evaluates to the sum of two expressions.
type SQLAddExpr sqlBinaryNode

func (add *SQLAddExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := convertToSQLNumeric(add.left, ctx)
	if err != nil {
		return nil, err
	}

	rightVal, err := convertToSQLNumeric(add.right, ctx)
	if err != nil {
		return nil, err
	}

	if leftVal == nil || rightVal == nil {
		return SQLNull, nil
	}

	return leftVal.Add(rightVal), nil
}

func (add *SQLAddExpr) String() string {
	return fmt.Sprintf("%v+%v", add.left, add.right)
}

func (add *SQLAddExpr) Type() schema.SQLType {
	return preferentialType(add.left, add.right)
}

// SQLDivideExpr evaluates to the quotient of the left expression divided by the right.
type SQLDivideExpr sqlBinaryNode

func (div *SQLDivideExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := convertToSQLNumeric(div.left, ctx)
	if err != nil {
		return nil, err
	}

	rightVal, err := convertToSQLNumeric(div.right, ctx)
	if err != nil {
		return nil, err
	}

	if leftVal == nil || rightVal == nil {
		return SQLNull, nil
	}

	if rightVal.Float64() == 0 {
		// NOTE: this is per the mysql manual.
		return SQLNull, nil
	}

	return SQLFloat(leftVal.Float64() / rightVal.Float64()), nil
}

func (div *SQLDivideExpr) String() string {
	return fmt.Sprintf("%v/%v", div.left, div.right)
}

func (div *SQLDivideExpr) Type() schema.SQLType {
	return preferentialType(div.left, div.right)
}

// SQLMultiplyExpr evaluates to the product of two expressions
type SQLMultiplyExpr sqlBinaryNode

func (mult *SQLMultiplyExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := convertToSQLNumeric(mult.left, ctx)
	if err != nil {
		return nil, err
	}

	rightVal, err := convertToSQLNumeric(mult.right, ctx)
	if err != nil {
		return nil, err
	}

	if leftVal == nil || rightVal == nil {
		return SQLNull, nil
	}

	return leftVal.Product(rightVal), nil
}

func (mult *SQLMultiplyExpr) String() string {
	return fmt.Sprintf("%v*%v", mult.left, mult.right)
}

func (mult *SQLMultiplyExpr) Type() schema.SQLType {
	return preferentialType(mult.left, mult.right)
}

// SQLSubtractExpr evaluates to the difference of the left expression minus the right expressions.
type SQLSubtractExpr sqlBinaryNode

func (sub *SQLSubtractExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftVal, err := convertToSQLNumeric(sub.left, ctx)
	if err != nil {
		return nil, err
	}

	rightVal, err := convertToSQLNumeric(sub.right, ctx)
	if err != nil {
		return nil, err
	}

	if leftVal == nil || rightVal == nil {
		return SQLNull, nil
	}

	return leftVal.Sub(rightVal), nil
}

func (sub *SQLSubtractExpr) String() string {
	return fmt.Sprintf("%v-%v", sub.left, sub.right)
}

func (sub *SQLSubtractExpr) Type() schema.SQLType {
	return preferentialType(sub.left, sub.right)
}

//
// SQLUnaryMinusExpr evaluates to the negation of the expression.
//
type SQLUnaryMinusExpr sqlUnaryNode

func (um *SQLUnaryMinusExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, ok := um.operand.(SQLNumeric); ok {
		return SQLInt(-(round(val.Float64()))), nil
	}

	return um.operand.Evaluate(ctx)
}

func (um *SQLUnaryMinusExpr) String() string {
	return fmt.Sprintf("-%v", um.operand)
}

func (um *SQLUnaryMinusExpr) Type() schema.SQLType {
	return um.operand.Type()
}

//
// SQLUnaryTildeExpr evaluates to the bitwise complement of the expression.
//
type SQLUnaryTildeExpr sqlUnaryNode

func (td *SQLUnaryTildeExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, ok := td.operand.(SQLNumeric); ok {
		return SQLInt(^round(val.Float64())), nil
	}

	return td.operand.Evaluate(ctx)
}

func (td *SQLUnaryTildeExpr) String() string {
	return fmt.Sprintf("~%v", td.operand)
}

func (td *SQLUnaryTildeExpr) Type() schema.SQLType {
	return td.operand.Type()
}

func convertToSQLNumeric(expr SQLExpr, ctx *EvalCtx) (SQLNumeric, error) {
	eval, err := expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	switch v := eval.(type) {
	case SQLNumeric:
		return v, nil
	case *SQLValues:
		if len(v.Values) != 1 {
			return nil, fmt.Errorf("expected only one SQLValues value - got %v", len(v.Values))
		}
		return convertToSQLNumeric(v.Values[0], ctx)
	default:
		return nil, nil
	}
}
