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
		var sum SQLNumeric
		sum = SQLInt(0)
		for _, arg := range args {
			c, err := arg.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
			switch v := c.(type) {
			case SQLNumeric:
				sum = sum.Add(v)
			case SQLNullValue:
				// treat null as 0
				continue
			default:
				return nil, fmt.Errorf("illegal argument: not numeric")
			}
		}
		return sum, nil
	}),
	"-": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) (SQLValue, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("- function needs at least 2 args")
		}
		var diff *SQLNumeric
		for _, arg := range args {
			c, err := arg.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
			switch v := c.(type) {
			case SQLNumeric:
				if diff == nil {
					diff = &v
				} else {
					d := (*diff).Sub(v)
					diff = &d
				}
			case SQLNullValue:
				// treat null as 0
				continue
			default:
				return nil, fmt.Errorf("illegal argument: not numeric")
			}
		}
		return *diff, nil
	}),
	"*": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) (SQLValue, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("* function needs at least 2 args")
		}
		var product SQLNumeric = SQLFloat(1)
		for _, arg := range args {
			c, err := arg.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
			switch v := c.(type) {
			case SQLNumeric:
				product = product.Product(v)
			case SQLNullValue:
				// treat null as 0
				continue
			default:
				return nil, fmt.Errorf("illegal argument: not numeric")
			}
		}
		return SQLNumeric(product), nil
	}),
	"/": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) (SQLValue, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("/ function must take 2 args")
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
			return nil, fmt.Errorf("both args to / must be numeric")
		}

		if right.Float64() == 0 {
			return nil, fmt.Errorf("divide by zero")
		}
		return SQLFloat(left.Float64() / right.Float64()), nil
	}),
}
