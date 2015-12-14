package evaluator

import (
	"fmt"
)

// SQLAddExpr evaluates to the sum of two expressions.
type SQLAddExpr sqlBinaryNode

func (add *SQLAddExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftEvald, err := convertToSQLNumeric(add.left, ctx)
	if err != nil {
		return nil, err
	}
	rightEvald, err := convertToSQLNumeric(add.right, ctx)
	if err != nil {
		return nil, err
	}

	return leftEvald.Add(rightEvald), nil
}

// SQLDivideExpr evaluates to the quotient of the left expression divided by the right.
type SQLDivideExpr sqlBinaryNode

func (div *SQLDivideExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftEvald, err := convertToSQLNumeric(div.left, ctx)
	if err != nil {
		return nil, err
	}
	rightEvald, err := convertToSQLNumeric(div.right, ctx)
	if err != nil {
		return nil, err
	}

	if rightEvald.Float64() == 0 {
		// NOTE: this is per the mysql manual.
		return SQLNull, nil
	}

	return SQLFloat(leftEvald.Float64() / rightEvald.Float64()), nil
}

// SQLMultiplyExpr evaluates to the product of two expressions
type SQLMultiplyExpr sqlBinaryNode

func (mult *SQLMultiplyExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftEvald, err := convertToSQLNumeric(mult.left, ctx)
	if err != nil {
		return nil, err
	}
	rightEvald, err := convertToSQLNumeric(mult.right, ctx)
	if err != nil {
		return nil, err
	}

	return leftEvald.Product(rightEvald), nil
}

// SQLSubtractExpr evaluates to the difference of the left expression minus the right expressions.
type SQLSubtractExpr sqlBinaryNode

func (sub *SQLSubtractExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	leftEvald, err := convertToSQLNumeric(sub.left, ctx)
	if err != nil {
		return nil, err
	}
	rightEvald, err := convertToSQLNumeric(sub.right, ctx)
	if err != nil {
		return nil, err
	}

	return leftEvald.Sub(rightEvald), nil
}

//
// SQLUnaryMinusExpr evaluates to the negation of the expression.
//
type SQLUnaryMinusExpr sqlUnaryNode

func (um *SQLUnaryMinusExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, ok := um.operand.(SQLNumeric); ok {
		return SQLInt(-(round(val.Float64()))), nil
	}

	// NOTE: this doesn't seem right... the negation of a non-number should
	// be an illegal
	return um.operand.Evaluate(ctx)
}

//
// SQLUnaryPlusExpr represents a unary plus expression.
//
type SQLUnaryPlusExpr sqlUnaryNode

func (up *SQLUnaryPlusExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, ok := up.operand.(SQLNumeric); ok {
		// NOTE: where is the documentation that this is correct?
		return SQLInt(round(val.Float64())), nil
	}

	// NOTE: this doesn't seem right to ignore it...
	return up.operand.Evaluate(ctx)
}

//
// SQLUnaryTildeExpr evaluates to the bitwise complement of the expression.
//
type SQLUnaryTildeExpr sqlUnaryNode

func (td *SQLUnaryTildeExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	if val, ok := td.operand.(SQLNumeric); ok {
		return SQLInt(^round(val.Float64())), nil
	}

	// NOTE: this doesn't seem right to ignore it...
	return td.operand.Evaluate(ctx)
}

func convertToSQLNumeric(expr SQLExpr, ctx *EvalCtx) (SQLNumeric, error) {
	eval, err := expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	switch v := eval.(type) {
	case SQLNumeric:
		return v, nil
	case SQLValues:
		if len(v) != 1 {
			return nil, fmt.Errorf("expected only one SQLValues value - got %v", len(v))
		}
		return convertToSQLNumeric(v[0], ctx)
	default:
		return nil, fmt.Errorf("can not convert %T to SQLNumeric", eval)
	}
}
