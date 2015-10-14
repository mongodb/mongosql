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
	panic("can't generate mongo value from a function")
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

	lEval, err := args[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	left, leftOk := (lEval).(SQLNumeric)

	rEval, err := args[1].Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	right, rightOk := (rEval).(SQLNumeric)

	if !(leftOk && rightOk) {
		return &SQLNullValue{}, nil
	}

	switch op {
	case "+":
		return left.Add(right), nil
	case "-":
		return left.Sub(right), nil
	case "*":
		return left.Product(right), nil
	case "/":
		if right.Float64() == 0 {
			return &SQLNullValue{}, nil
		}
		return SQLFloat(left.Float64() / right.Float64()), nil
	}
	return nil, fmt.Errorf("unsupported numeric binary operation: '%v'", op)
}
