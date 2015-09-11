package planner

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

type SQLFunction func([]SQLValue, *MatchCtx) SQLValue

type SQLFuncValue struct {
	arguments []SQLValue
	function  func([]SQLValue, *MatchCtx) SQLValue
}

func (sqlfv *SQLFuncValue) Transform() (*bson.D, error) {
	return nil, fmt.Errorf("transformation of functional expression not supported")
}

func (sqlfunc *SQLFuncValue) CompareTo(ctx *MatchCtx, v SQLValue) (int, error) {
	left := sqlfunc.Evaluate(ctx)
	right := v.Evaluate(ctx)
	return left.CompareTo(ctx, right)
}

func (sqlfunc *SQLFuncValue) Evaluate(ctx *MatchCtx) SQLValue {
	return sqlfunc.function(sqlfunc.arguments, ctx)
}

func (sqlfunc *SQLFuncValue) MongoValue() interface{} {
	panic("can't generate mongo value from a function")
}

var funcMap = map[string]SQLFunction{
	"+": SQLFunction(func(args []SQLValue, ctx *MatchCtx) SQLValue {
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
	"-": SQLFunction(func(args []SQLValue, ctx *MatchCtx) SQLValue {
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
	"*": SQLFunction(func(args []SQLValue, ctx *MatchCtx) SQLValue {
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
	"/": SQLFunction(func(args []SQLValue, ctx *MatchCtx) SQLValue {
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
