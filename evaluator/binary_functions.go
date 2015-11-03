package evaluator

import (
	"fmt"
)

type SQLBinaryFunction func([]SQLValue, *EvalCtx) (SQLValue, error)

type SQLBinaryValue struct {
	arguments []SQLValue
	function  func([]SQLValue, *EvalCtx) (SQLValue, error)
}

func (sqlfunc *SQLBinaryValue) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
	left, err := sqlfunc.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	right, err := v.Evaluate(ctx)
	if err != nil {
		return 0, err
	}
	return left.CompareTo(ctx, right)
}

func (sqlfunc *SQLBinaryValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sqlfunc.function(sqlfunc.arguments, ctx)
}

func (sqlfunc *SQLBinaryValue) MongoValue() interface{} {
	return nil
}

var binaryFuncMap = map[string]SQLBinaryFunction{
	"+": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) (SQLValue, error) {
		return SQLNumericBinaryOp(args, ctx, "+")
	}),
	"-": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) (SQLValue, error) {
		return SQLNumericBinaryOp(args, ctx, "-")
	}),
	"*": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) (SQLValue, error) {
		return SQLNumericBinaryOp(args, ctx, "*")
	}),
	"/": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) (SQLValue, error) {
		return SQLNumericBinaryOp(args, ctx, "/")
	}),
}

func convertToSQLNumeric(v SQLValue, ctx *EvalCtx) (SQLNumeric, error) {
	v, err := v.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	eval, ok := v.(SQLNumeric)
	if !ok {
		v, ok := v.(SQLValues)
		if !ok {
			return nil, fmt.Errorf("could not convert binary operand to SQLValues")
		}
		if len(v.Values) != 1 {
			return nil, fmt.Errorf("expected only one other composite operand - got %v", len(v.Values))
		}
		eval, ok = v.Values[0].(SQLNumeric)
		if !ok {
			return nil, fmt.Errorf("could not convert binary operand to SQLNumeric")
		}
	}
	return eval, nil
}

func SQLNumericBinaryOp(args []SQLValue, ctx *EvalCtx, op string) (SQLValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%v function needs at least 2 args", op)
	}

	left, err := convertToSQLNumeric(args[0], ctx)
	if err != nil {
		return nil, err
	}

	for _, arg := range args[1:] {
		right, err := convertToSQLNumeric(arg, ctx)
		if err != nil {
			return nil, err
		}
		switch op {
		case "+":
			left = left.Add(right)
		case "-":
			left = left.Sub(right)
		case "*":
			left = left.Product(right)
		case "/":
			if right.Float64() == 0 {
				return &SQLNullValue{}, nil
			}
			left = SQLFloat(left.Float64() / right.Float64())
		default:
			return nil, fmt.Errorf("unsupported numeric binary operation: '%v'", op)
		}
	}
	return left, nil
}
