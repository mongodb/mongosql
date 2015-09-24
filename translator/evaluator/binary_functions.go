package evaluator

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

type SQLBinaryFunction func([]SQLValue, *EvalCtx) SQLValue

var binaryFuncMap = map[string]SQLBinaryFunction{
	"+": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) SQLValue {
		sum := float64(0)
		for _, arg := range args {
			c := arg.Evaluate(ctx)
			switch v := c.(type) {
			case SQLNumeric:
				sum = sum + float64(v)
			case SQLNullValue:
				// treat null as 0
				continue
			default:
				panic("illegal argument: not numeric")
			}
		}
		return SQLNumeric(sum)
	}),
	"-": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) SQLValue {
		if len(args) < 2 {
			panic("- function needs at least 2 args")
		}
		var diff *float64
		for _, arg := range args {
			c := arg.Evaluate(ctx)
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
				panic("illegal argument: not numeric")
			}
		}
		return SQLNumeric(*diff)
	}),
	"*": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) SQLValue {
		if len(args) < 2 {
			panic("* function needs at least 2 args")
		}
		product := float64(1)
		for _, arg := range args {
			c := arg.Evaluate(ctx)
			switch v := c.(type) {
			case SQLNumeric:
				product = product * float64(v)
			case SQLNullValue:
				// treat null as 0
				continue
			default:
				panic("illegal argument: not numeric")
			}
		}
		return SQLNumeric(product)
	}),
	"/": SQLBinaryFunction(func(args []SQLValue, ctx *EvalCtx) SQLValue {
		if len(args) != 2 {
			panic("/ function must take 2 args")
		}
		left, leftOk := (args[0].Evaluate(ctx)).(SQLNumeric)
		right, rightOk := (args[1].Evaluate(ctx)).(SQLNumeric)
		if !(leftOk && rightOk) {
			panic("both args to / must be numeric")
		}
		if float64(right) == 0 {
			panic("divide by zero")
		}
		return SQLNumeric(float64(left) / float64(right))
	}),
}
