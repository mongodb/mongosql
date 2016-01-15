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

func (add *SQLAddExpr) String() string {
	return fmt.Sprintf("%v+%v", add.left, add.right)
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

func (div *SQLDivideExpr) String() string {
	return fmt.Sprintf("%v/%v", div.left, div.right)
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

func (mult *SQLMultiplyExpr) String() string {
	return fmt.Sprintf("%v*%v", mult.left, mult.right)
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

func (sub *SQLSubtractExpr) String() string {
	return fmt.Sprintf("%v-%v", sub.left, sub.right)
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
		return nil, fmt.Errorf("can not convert %T to SQLNumeric", eval)
	}
}
