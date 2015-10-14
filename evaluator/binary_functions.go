package evaluator

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

type SQLBinaryFunction func([]SQLValue, *EvalCtx) (SQLValue, error)

type SQLBinaryExprValue struct {
	arguments []SQLValue
	function  func([]SQLValue, *EvalCtx) (SQLValue, error)
}

func (sqlfv *SQLBinaryExprValue) Transform() (*bson.D, error) {
	return nil, fmt.Errorf("transformation of functional expression not supported")
}

func (sqlfunc *SQLBinaryExprValue) CompareTo(ctx *EvalCtx, v SQLValue) (int, error) {
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

func (sqlfunc *SQLBinaryExprValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sqlfunc.function(sqlfunc.arguments, ctx)
}

func (sqlfunc *SQLBinaryExprValue) MongoValue() interface{} {
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

func SQLNumericBinaryOp(args []SQLValue, ctx *EvalCtx, op string) (SQLValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%v function needs at least 2 args", op)
	}

	eval, err := args[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	left, ok := eval.(SQLNumeric)
	if !ok {
		return &SQLNullValue{}, nil
	}

	for _, arg := range args[1:] {
		eval, err = arg.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		right, ok := (eval).(SQLNumeric)
		if !ok {
			return &SQLNullValue{}, nil
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
