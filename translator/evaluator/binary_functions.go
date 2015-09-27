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
		sum := float64(0)
		for _, arg := range args {
			c, err := arg.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
			switch v := c.(type) {
			case SQLNumeric:
				sum = sum + float64(v)
			case SQLNullValue:
				// treat null as 0
				continue
			default:
				return nil, fmt.Errorf("illegal argument: not numeric")
			}
		}
		return SQLNumeric(sum), nil
	}),
	"-": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) (SQLValue, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("- function needs at least 2 args")
		}
		var diff *float64
		for _, arg := range args {
			c, err := arg.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
			switch v := c.(type) {
			case SQLNumeric:
				if diff == nil {
					diff = new(float64)
					*diff = float64(v)
				} else {
					*diff = *diff - float64(v)
				}
			case SQLNullValue:
				// treat null as 0
				continue
			default:
				return nil, fmt.Errorf("illegal argument: not numeric")
			}
		}
		return SQLNumeric(*diff), nil
	}),
	"*": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) (SQLValue, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("* function needs at least 2 args")
		}
		product := float64(1)
		for _, arg := range args {
			c, err := arg.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
			switch v := c.(type) {
			case SQLNumeric:
				product = product * float64(v)
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

		if float64(right) == 0 {
			return nil, fmt.Errorf("divide by zero")
		}
		return SQLNumeric(float64(left) / float64(right)), nil
	}),
}
