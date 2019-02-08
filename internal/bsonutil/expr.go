package bsonutil

import (
	"math"

	"github.com/10gen/mongo-go-driver/bson"
)

//
// Expression Translation Wrappers
//
const (
	OpAbs            = "$abs"
	OpAdd            = "$add"
	OpAnd            = "$and"
	OpAnyElementTrue = "$anyElementTrue"
	OpArrElemAt      = "$arrayElemAt"
	OpAvg            = "$avg"
	OpCeil           = "$ceil"
	OpConcat         = "$concat"
	OpCond           = "$cond"
	OpConvert        = "$convert"
	OpDateFromParts  = "$dateFromParts"
	OpDateFromString = "$dateFromString"
	OpDateToString   = "$dateToString"
	OpDayOfMonth     = "$dayOfMonth"
	OpDayOfWeek      = "$dayOfWeek"
	OpDayOfYear      = "$dayOfYear"
	OpDivide         = "$divide"
	OpEq             = "$eq"
	OpExists         = "$exists"
	OpFilter         = "$filter"
	OpFloor          = "$floor"
	OpGt             = "$gt"
	OpGte            = "$gte"
	OpHour           = "$hour"
	OpIfNull         = "$ifNull"
	OpIn             = "$in"
	OpIndexOfCP      = "$indexOfCP"
	OpLt             = "$lt"
	OpLte            = "$lte"
	OpLet            = "$let"
	OpLiteral        = "$literal"
	OpNaturalLog     = "$ln"
	OpLog            = "$log"
	OpLTrim          = "$ltrim"
	OpMap            = "$map"
	OpMax            = "$max"
	OpMin            = "$min"
	OpMinute         = "$minute"
	OpMillisecond    = "$millisecond"
	OpMod            = "$mod"
	OpMonth          = "$month"
	OpMultiply       = "$multiply"
	OpNeq            = "$ne"
	OpNotIn          = "$nin"
	OpNor            = "$nor"
	OpNot            = "$not"
	OpOr             = "$or"
	OpPow            = "$pow"
	OpRange          = "$range"
	OpReduce         = "$reduce"
	OpRegex          = "$regex"
	OpReverseArray   = "$reverseArray"
	OpRTrim          = "$rtrim"
	OpSecond         = "$second"
	OpSize           = "$size"
	OpSlice          = "$slice"
	OpSplit          = "$split"
	OpSqrt           = "$sqrt"
	OpStdDevPop      = "$stdDevPop"
	OpStdDevSamp     = "$stdDevSamp"
	OpStrLenBytes    = "$strLenBytes"
	OpStrlenCP       = "$strLenCP"
	OpSubstr         = "$substrCP"
	OpSubtract       = "$subtract"
	OpSum            = "$sum"
	OpSwitch         = "$switch"
	OpToLower        = "$toLower"
	OpToUpper        = "$toUpper"
	OpTrim           = "$trim"
	OpTrunc          = "$trunc"
	OpType           = "$type"
	OpWeek           = "$week"
	OpYear           = "$year"
	OpZip            = "$zip"
)

var (
	MgoNullLiteral         = WrapInLiteral(nil)
	DateComponentSeparator = NewArray("!", "\"", "#", WrapInLiteral("$"), "%", "&", "'",
		"(", ")", "*", "+", ",", "-", ".", "/", ":", ";", "<", "=", ">", "?", "@", "[", "\\", "]",
		"^", "_", "`", "{", "|", "}", "~")
)

const (
	// millisecondsPerDay is the number of milliseconds in a day.
	millisecondsPerDay = 8.64e+7
)

// factorial is an array giving the factorial of 0 <= n <= 15.
var factorial = []float64{
	1.0,
	1.0,
	2.0,
	6.0,
	24.0,
	120.0,
	720.0,
	5040.0,
	40320.0,
	362880.0,
	3628800.0,
	39916800.0,
	479001600.0,
	6227020800.0,
	87178291200.0,
	1307674368000.0,
}

// GetLiteral returns the value of an inner $literal if
// one is present, and nil otherwise.
func GetLiteral(v interface{}) (interface{}, bool) {
	if bsonMap, ok := v.(bson.M); ok {
		if bsonVal, ok := bsonMap[OpLiteral]; ok {
			return bsonVal, true
		}
	}
	return nil, false
}

// WrapInAcosComputation wraps the argument in an expression
// that computes the arccos of the argument.
func WrapInAcosComputation(expr interface{}) interface{} {
	input := "$$input"
	inputLetAssignment := NewM(
		NewDocElem("input", expr),
	)

	absInput := "$$absInput"
	absInputLetAssignment := NewM(
		NewDocElem("absInput", WrapInOp(OpAbs, input)),
	)

	// The power series for arccos does not converge well, so instead use
	// this function: from the Handbook of Mathematical Functions, by
	// Milton Abramowitz and Irene Stegun: arccos(x)=sqrt(1-x) *
	// (a0+a1∗x+a2∗x2+a3∗x3). This function is only good far away from -1,
	// so we just mirror the function for negative values by subtracting
	// from Pi (the value of acos(-1)). The constants a0-a3 are defined as
	// follows:
	a0 := 1.5707288
	a1 := -0.2121144
	a2 := 0.0742610
	a3 := -0.0187293

	firstTerm := WrapInOp(OpSqrt, WrapInOp(OpSubtract, 1.0, absInput))
	secondTerm := WrapInOp(OpAdd,
		a0,
		WrapInOp(OpMultiply, a1, absInput),
		WrapInOp(OpMultiply, a2, WrapInOp(OpPow, absInput, 2)),
		WrapInOp(OpMultiply, a3, WrapInOp(OpPow, absInput, 3)),
	)

	return WrapInLet(inputLetAssignment,
		WrapInLet(absInputLetAssignment,
			WrapInCond(
				WrapInOp(OpMultiply, firstTerm, secondTerm),
				WrapInOp(OpSubtract,
					math.Pi,
					WrapInOp(OpMultiply,
						firstTerm,
						secondTerm)),
				WrapInOp(OpGte, input, 0),
			),
		),
	)
}

// WrapInBinOp builds an expression that evaluates a two argument operator
// on the two passed argument expressions.
func WrapInBinOp(op string, v1, v2 interface{}) bson.M {
	return NewM(NewDocElem(op, NewArray(
		v1,
		v2,
	)))
}

// WrapInCase returns an expression to use as one of the branches arguments to WrapInSwitch.
// caseExpr must evaluate to a boolean.
func WrapInCase(caseExpr, thenExpr interface{}) bson.M {
	return NewM(NewDocElem("case", caseExpr), NewDocElem("then", thenExpr))
}

// WrapInConcat returns the aggregation expression
// {$concat: [expr1, expr2, ...]}
// https://docs.mongodb.com/manual/reference/operator/aggregation/concat/
func WrapInConcat(exprs []interface{}) bson.M {
	return NewM(NewDocElem(OpConcat, exprs))
}

// WrapInCond returns a document that evalutes to truePart
// if any of conds is true, and falsePart otherwise.
func WrapInCond(truePart, falsePart interface{}, conds ...interface{}) interface{} {
	var condition interface{}

	switch len(conds) {
	case 0:
		return falsePart
	case 1:
		condition = conds[0]
	default:
		condition = WrapInOp(OpOr, conds...)
	}

	return NewM(NewDocElem(OpCond, NewArray(
		condition,
		truePart,
		falsePart,
	)))
}

// WrapInConvert takes input and wraps it in a $convert operation naively, without
// accounting for all special conversions needed to reflect mySQL behavior. DO NOT USE this
// function to convert directly; instead, call evaluator/translateConvert for a correct answer in all cases.
func WrapInConvert(input interface{}, to string, onError, onNull interface{}) bson.M {
	return NewM(
		NewDocElem(OpConvert, NewM(
			NewDocElem("input", input),
			NewDocElem("to", to),
			NewDocElem("onError", onError),
			NewDocElem("onNull", onNull),
		)),
	)

}

// WrapInCosPowerSeries wraps the argument in an expression that computes the
// cos Maclaurin power series of the argument, expr.
// http://mathworld.wolfram.com/MaclaurinSeries.html
func WrapInCosPowerSeries(expr interface{}) bson.M {
	input := "$$input"
	inputLetAssignment := NewM(
		NewDocElem("input", expr),
	)

	return WrapInLet(inputLetAssignment,
		WrapInOp(OpAdd,
			1,
			WrapInPowerSeriesTerm(input, 2),
			WrapInPowerSeriesTerm(input, 4),
			WrapInPowerSeriesTerm(input, 6),
			WrapInPowerSeriesTerm(input, 8),
			WrapInPowerSeriesTerm(input, 10),
			WrapInPowerSeriesTerm(input, 12),
			WrapInPowerSeriesTerm(input, 14),
		),
	)
}

// WrapInDateFormat wraps an Aggregation Expression that evaluates to a date
// in a date_format expression that will use '$dateFromString' to format
// a date to a string.
// The conditions on which to return nil may be provided; otherwise the date will
// be wrapped in a null-check.
func WrapInDateFormat(date interface{}, mysqlFormat string, conds ...interface{}) (interface{}, bool) {
	var format string
	for i := 0; i < len(mysqlFormat); i++ {
		if mysqlFormat[i] == '%' {
			if i != len(mysqlFormat)-1 {
				switch mysqlFormat[i+1] {
				case '%':
					format += "%%"
				case 'd':
					format += "%d"
				case 'f':
					format += "%L000"
				case 'H', 'k':
					format += "%H"
				case 'i':
					format += "%M"
				case 'j':
					format += "%j"
				case 'm':
					format += "%m"
				case 's', 'S':
					format += "%S"
				case 'T':
					format += "%H:%M:%S"
				case 'U':
					format += "%U"
				case 'Y':
					format += "%Y"
				default:
					return nil, false
				}
				i++
			} else {
				// MongoDB fails when the last character is a % sign in the format string.
				return nil, false
			}
		} else {
			format += string(mysqlFormat[i])
		}
	}

	if val, ok := GetLiteral(date); ok && val == nil {
		return nil, true
	} else if len(conds) == 0 {
		conds = append(conds, WrapInNullCheck(date))
	}

	return WrapInCond(
		nil,
		NewM(NewDocElem(OpDateToString, NewM(
			NewDocElem("format", format),
			NewDocElem("date", date),
		))),
		conds...,
	), true
}

// WrapInDateFromParts returns a date given the year, month and day passed in.
func WrapInDateFromParts(year, month, dayOfMonth interface{}) bson.M {
	return NewM(
		NewDocElem(OpDateFromParts, NewM(
			NewDocElem("year", NewM(NewDocElem(OpYear, year))),
			NewDocElem("month", NewM(NewDocElem(OpMonth, month))),
			NewDocElem("day", NewM(NewDocElem(OpDayOfMonth, dayOfMonth))),
		)),
	)

}

// WrapInDateToString converts date to a string according to the specified format.
func WrapInDateToString(date interface{}, format string) bson.M {
	return NewM(
		NewDocElem(OpDateToString, NewM(
			NewDocElem("date", date),
			NewDocElem("format", format),
		)),
	)

}

// WrapInEqCase returns a document that is a case arm that checks equality between expr1 and expr2.
func WrapInEqCase(expr1, expr2, thenExpr interface{}) bson.M {
	caseExpr := WrapInOp(OpEq, expr1, expr2)
	return NewM(NewDocElem("case", caseExpr), NewDocElem("then", thenExpr))
}

// WrapInIfNull returns v if it isn't nil, otherwise, it returns ifNull.
func WrapInIfNull(v, ifNull interface{}) interface{} {
	if value, ok := GetLiteral(v); ok {
		if value == nil {
			return ifNull
		}
		return v
	}
	return NewM(NewDocElem(OpIfNull, NewArray(
		v,
		ifNull,
	)))
}

// WrapInInRange returns an expression that evaluates to true if val is in range [min, max).
// val must evaluate to a number.
func WrapInInRange(val interface{}, min, max float64) interface{} {
	return WrapInOp(OpAnd,
		WrapInOp(OpGte, val, min),
		WrapInOp(OpLt, val, max))
}

// WrapInIntDiv performs an integer division (truncated division).
func WrapInIntDiv(numerator, denominator interface{}) interface{} {
	return WrapInOp(OpTrunc,
		WrapInOp(OpDivide, numerator, denominator))
}

// WrapInIsLeapYear creates an expression that returns true if the argument is
// a leap year, and false otherwise. This function assume val is an integer
// year.
func WrapInIsLeapYear(val interface{}) bson.M {
	v := "$$val"
	letAssignment := NewM(
		NewDocElem("val", val),
	)

	// This computes the expression:
	// (v % 4 == 0) && (v % 100 != 0) || (v % 400 == 0).
	return WrapInLet(letAssignment,
		WrapInOp(OpOr,
			WrapInOp(OpAnd,
				WrapInOp(OpEq,
					WrapInOp(OpMod, v, 4),
					0),
				WrapInOp(OpNeq,
					WrapInOp(OpMod, v, 100),
					0),
			),
			WrapInOp(OpEq,
				WrapInOp(OpMod, v, 400),
				0),
		),
	)
}

// WrapInLet returns a document with v as vars, and i as in.
func WrapInLet(v, i interface{}) bson.M {
	return NewM(
		NewDocElem(OpLet,
			NewM(NewDocElem("vars", v),
				NewDocElem("in", i))))
}

// WrapInLiteral returns a document with v passed to $literal.
func WrapInLiteral(v interface{}) bson.M {
	return NewM(NewDocElem(OpLiteral, v))
}

// WrapInLiteralIfNeeded returns a document with v passed to $literal if it is not already wrapped
// in $literal.
func WrapInLiteralIfNeeded(v interface{}) bson.M {
	if val, ok := v.(bson.M); ok {
		if _, ok := val[OpLiteral]; ok && len(val) == 1 {
			return val
		}
	}
	return WrapInLiteral(v)
}

// WrapInLRTrim returns a trimmed version of args.
func WrapInLRTrim(isLTrimType bool, args interface{}) interface{} {
	var (
		splitArray = NewM(NewDocElem(OpSplit, NewArray(
			args,
			" ",
		)))
		substrIndex  interface{}
		substrLength interface{}
	)

	if !isLTrimType {
		splitArray = NewM(NewDocElem(OpReverseArray, splitArray))
	}

	mapInput := WrapInLet(NewM(NewDocElem("splitArray", splitArray)),
		NewM(
			NewDocElem(OpZip, NewM(
				NewDocElem("inputs", NewArray(
					"$$splitArray",
					NewM(
						NewDocElem(OpRange, NewArray(
							0,
							NewM(NewDocElem(OpSize, "$$splitArray")),
						))),
				))))))

	mapIn := WrapInCond(NewM(
		NewDocElem(OpStrlenCP, args)),
		NewM(NewDocElem(OpArrElemAt, NewArray(
			"$$zipArray",
			1,
		))), NewM(NewDocElem(OpEq, NewArray(
			NewM(NewDocElem(OpArrElemAt, NewArray(
				"$$zipArray",
				0,
			))),
			"",
		))))

	min := NewM(NewDocElem(OpMin, WrapInMap(mapInput, "zipArray", mapIn)))

	if isLTrimType {
		substrIndex = min
		substrLength = NewM(NewDocElem(OpStrlenCP, args))
	} else {
		substrIndex = 0
		substrLength = NewM(NewDocElem(OpSubtract, NewArray(
			NewM(NewDocElem(OpStrlenCP, args)),
			min,
		)))
	}

	return NewM(
		NewDocElem(OpSubstr, NewArray(
			args,
			substrIndex,
			substrLength,
		)),
	)

}

// WrapInMap returns the aggregation expression {$map: {input: input, as: as, in: in }}.
// https://docs.mongodb.com/manual/reference/operator/aggregation/map/
func WrapInMap(input, as, in interface{}) bson.M {
	return NewM(
		NewDocElem(OpMap,
			NewM(
				NewDocElem("input", input),
				NewDocElem("as", as),
				NewDocElem("in", in))))
}

// WrapInNullCheck returns true if v is null, false otherwise.
func WrapInNullCheck(v interface{}) interface{} {
	if _, ok := GetLiteral(v); ok {
		return v
	}

	return WrapInOp(OpLte, v, nil)
}

// WrapInNullCheckedCond returns a document that evaluates to truePart
// if any of the null checked conds is true, and falsePart otherwise.
func WrapInNullCheckedCond(truePart, falsePart interface{}, conds ...interface{}) interface{} {
	var condition interface{}
	newConds := NewArray()
	for _, cond := range conds {
		if value, ok := GetLiteral(cond); !ok {
			newConds = append(newConds, WrapInNullCheck(cond))
		} else if value == nil {
			return truePart
		}
	}
	switch len(newConds) {
	case 0:
		return falsePart
	case 1:
		condition = newConds[0]
	default:
		condition = NewM(NewDocElem(OpOr, newConds))
	}

	return NewM(NewDocElem(OpCond, NewArray(
		condition,
		truePart,
		falsePart,
	)))
}

// WrapInOp returns a document which passes all arguments to the op.
func WrapInOp(op string, args ...interface{}) interface{} {
	return NewM(NewDocElem(op, args))
}

// WrapInPowerSeriesTerm takes an input and a power and produces the power
// series term for that integer as a MongoDB aggregration expression that is
// defined as input^power/ factorial(power).
func WrapInPowerSeriesTerm(input interface{}, power uint32) interface{} {
	ret := WrapInOp(OpDivide, WrapInOp(OpPow, input, power), factorial[power])
	pmod4 := power % 4
	// powers that are equal to 3 or 2 modulo 4 are negative in the Cos and
	// Sine series.
	if pmod4 == 3 || pmod4 == 2 {
		return WrapInOp(OpMultiply, -1.0, ret)
	}
	return ret
}

// WrapInRange returns the aggregation expression {$range: [start, stop, step]}.
// https://docs.mongodb.com/manual/reference/operator/aggregation/range/
func WrapInRange(start, stop, step interface{}) interface{} {
	if step != nil {
		return NewM(NewDocElem(OpRange, NewArray(
			start,
			stop,
			step,
		)))
	}
	return WrapInOp(OpRange, start, stop)
}

// WrapInReduce returns the aggregation expression
// {$reduce: {input: input, initialValue: initialValue, in: in }}.
// https://docs.mongodb.com/manual/reference/operator/aggregation/range/
func WrapInReduce(input, initialValue, in interface{}) bson.M {
	return NewM(
		NewDocElem(OpReduce,
			NewM(
				NewDocElem("input", input),
				NewDocElem("initialValue", initialValue),
				NewDocElem("in", in))))
}

// WrapInRound generates an expression to round a floating point number
// the way MySQL does. This is the simplest implementation of round I have found:
// https://github.com/golang/go/issues/4594#issuecomment-66073312.
func WrapInRound(val interface{}) interface{} {
	// The MongoDB aggregation language generated by this function for
	// non-literal values implements the following algorithm presented in go
	// code:
	// if x < 0 {
	//      return math.Ceil(x-.5)
	// }
	// return math.Floor(x+.5)
	// For literal values, the rounding is done in-memory, in Go.

	switch typedVal := val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return typedVal
	case float32:
		return int(math.Round(float64(typedVal)))
	case float64:
		return int(math.Round(typedVal))
	default:
		condExpr := WrapInOp(OpLt, val, 0.0)
		lt0 := WrapInOp(OpCeil, WrapInOp(OpSubtract, val, 0.5))
		gte0 := WrapInOp(OpFloor, WrapInOp(OpAdd, val, 0.5))
		return WrapInCond(lt0, gte0, condExpr)
	}
}

// WrapInRoundWithPrecision returns arg rounded to placeVal places.
func WrapInRoundWithPrecision(arg interface{}, placeVal float64) bson.M {
	decimal := math.Pow(float64(10), placeVal)
	if decimal < 1 {
		return WrapInLiteral(0)
	}

	letAssignment := NewM(
		NewDocElem("decimal", decimal),
	)

	letEvaluation := NewM(
		NewDocElem(OpDivide, NewArray(
			NewM(
				NewDocElem(OpCond, NewArray(
					NewM(
						NewDocElem(OpGte, NewArray(
							arg,
							0,
						))),
					NewM(
						NewDocElem(OpFloor, NewM(
							NewDocElem(OpAdd, NewArray(
								NewM(
									NewDocElem(OpMultiply, NewArray(
										arg,
										"$$decimal",
									)),
								),
								0.5,
							)),
						)),
					),
					NewM(
						NewDocElem(OpCeil, NewM(
							NewDocElem(OpSubtract, NewArray(
								NewM(
									NewDocElem(OpMultiply, NewArray(
										arg,
										"$$decimal",
									)),
								),
								0.5,
							)),
						)),
					),
				)),
			),
			"$$decimal",
		)),
	)

	return WrapInLet(letAssignment, letEvaluation)
}

// WrapInSinPowerSeries wraps the argument in an expression that computes the
// sin Maclaurin power series of the argument, expr.
// http://mathworld.wolfram.com/MaclaurinSeries.html
func WrapInSinPowerSeries(expr interface{}) bson.M {
	input := "$$input"
	inputLetAssignment := NewM(
		NewDocElem("input", expr),
	)

	return WrapInLet(inputLetAssignment,
		WrapInOp(OpAdd,
			input,
			WrapInPowerSeriesTerm(input, 3),
			WrapInPowerSeriesTerm(input, 5),
			WrapInPowerSeriesTerm(input, 7),
			WrapInPowerSeriesTerm(input, 9),
			WrapInPowerSeriesTerm(input, 11),
			WrapInPowerSeriesTerm(input, 13),
			WrapInPowerSeriesTerm(input, 15),
		),
	)
}

// WrapInStringToArray converts an expression v (which must evaluate to a string)
// to an array e.g. "hello" -> ["h", "e", "l", "l", "o"] and returns the array.
func WrapInStringToArray(v interface{}) bson.M {
	input := NewM(NewDocElem(OpRange, NewArray(
		0,
		NewM(NewDocElem(OpStrlenCP, v)),
	)))
	in := NewM(NewDocElem(OpSubstr, NewArray(
		v,
		"$$this",
		1,
	)))
	return NewM(
		NewDocElem(OpMap,
			NewM(
				NewDocElem("input", input),
				NewDocElem("in", in))))
}

// WrapInSubstr returns the aggregation expression
// {$substr: [string: string, start: start, length: length]}
// https://docs.mongodb.com/manual/reference/operator/aggregation/substr/
// nolint: unparam
func WrapInSubstr(str string, start int, length int) bson.M {
	return NewM(NewDocElem(OpSubstr, NewArray(
		str,
		start,
		length,
	)))
}

// WrapInSwitch returns the aggregation expression
// {$switch: branches: branches, default: defaultExpr }
// https://docs.mongodb.com/manual/reference/operator/aggregation/switch/
func WrapInSwitch(defaultExpr interface{}, branches ...bson.M) bson.M {
	return NewM(
		NewDocElem(OpSwitch,
			NewM(
				NewDocElem("branches", branches),
				NewDocElem("default", defaultExpr))))
}

// WrapInType wraps the passed expression in an expression
// that returns the type of the expression.
func WrapInType(v interface{}) bson.M {
	return NewM(NewDocElem(OpType, v))
}

// WrapInWeekCalculation calculates the week of a given date based on the
// passed argument, expr, which is some MongoDB Aggregation Pipeline
// expression, and the mode, which is an integer.
func WrapInWeekCalculation(expr interface{}, mode int64) interface{} {
	date, year := "$$date", "$$year"
	getJan1 := func() interface{} {
		return NewM(
			NewDocElem(OpDateFromParts, NewM(
				NewDocElem("year", year),
				NewDocElem("month", 1),
				NewDocElem("day", 1),
			)),
		)

	}

	getNextJan1 := func() interface{} {
		return NewM(
			NewDocElem(OpDateFromParts, NewM(
				NewDocElem("year", WrapInOp(OpAdd, year, 1)),
				NewDocElem("month", 1),
				NewDocElem("day", 1),
			)),
		)

	}

	// generateDaySubtract generates the main week calculation shared
	// by all modes except 0, 2 (since those can use MongoDB's week function).
	// The calculation is:
	// trunc((date - dayOne) / (7 * millisecondsPerDay) + 1).
	generateDaySubtract := func(dayOne interface{}) interface{} {
		return WrapInOp(OpTrunc,
			WrapInOp(OpAdd,
				WrapInOp(OpDivide,
					WrapInOp(OpSubtract, date, dayOne),
					7*millisecondsPerDay),
				1),
		)
	}

	// generate4DaysBody generates the body for modes where the first
	// week is defined by having 4 days in the year, these are modes
	// 1, 3, 4, and 6.
	generate4DaysBody := func(diffConstant int) interface{} {
		// This description is used for Monday as first day of the
		// week. See below for an explanation of the Sunday first day
		// case. Calculate the first day of the first week of this
		// year based on the dayOfWeek of YYYY-01-01 of this year, note
		// that it may be from the previous year. The Day Diff column
		// is the
		// amount of days to Add or Subtract from YYYY-01-01:
		// Day Of Week for Jan 1   |   Day Diff
		// ---------------------------------------------
		//                     1   |   + 1
		//                     2   |   + 0
		//                     3   |   - 1
		//                     4   |   - 2
		//                     5   |   - 3
		//                     6   |   + 3
		//                     7   |   + 2
		// This can be simplified to:
		// diff = -x + 2
		// if diff > -4 {
		//      return diff
		// }
		// return diff + 7
		// for the Sunday version of this, we need to use -x + 1
		// instead of -x + 2, and that is the only difference, that is
		// the point of "diffConstant", it will be either 1 or 2.
		jan1 := getJan1()
		jan1DayOfWeek := "$$jan1DayOfWeek"
		dayOfWeekLetAssignment := NewM(
			NewDocElem("jan1DayOfWeek", WrapInOp(OpDayOfWeek, jan1)),
		)

		dayOne := WrapInOp(OpAdd, jan1,
			WrapInOp(OpMultiply,
				WrapInLet(NewM(NewDocElem("diff", WrapInOp(OpAdd,
					WrapInOp(OpMultiply, jan1DayOfWeek, -1),
					diffConstant)),
				), WrapInCond("$$diff",
					WrapInOp(OpAdd, "$$diff", 7),
					WrapInOp(OpGt, "$$diff", -4),
				),
				),
				millisecondsPerDay,
			),
		)
		return WrapInLet(dayOfWeekLetAssignment, generateDaySubtract(dayOne))
	}

	// generateMondayBody generates the body for modes where the first
	// week is defined by having a Monday, these are modes
	// 5 and 7.
	generateMondayBody := func() interface{} {
		// These are more simple than the 4 days mode. The diff from Jan1
		// can be defined using (7 - x + 2) % 7.
		jan1 := getJan1()
		jan1DayOfWeek := "$$jan1DayOfWeek"
		dayOfWeekLetAssignment := NewM(
			NewDocElem("jan1DayOfWeek", WrapInOp(OpDayOfWeek, jan1)),
		)

		dayOne := WrapInOp(OpAdd, jan1,
			WrapInOp(OpMultiply,
				WrapInOp(OpMod,
					WrapInOp(OpAdd,
						WrapInOp(OpSubtract,
							7,
							jan1DayOfWeek,
						),
						2),
					7),
				millisecondsPerDay,
			),
		)
		return WrapInLet(dayOfWeekLetAssignment, generateDaySubtract(dayOne))
	}

	// WrapInZeroCheck - half of all modes allow weeks numbers
	// between 0-53, and the other half allow 1-53. To compute the week
	// for modes allowing 1-53, we compute the week for the associated 0-53
	// mode, and if it results in week 0, we return week('(year-1)-12-31'),
	// which will be either 52 or 53 as the 1-53 modes consider such a date
	// as being in the previous year. This means that
	// WrapInWeekCalculation must be recursive, which is why it is
	// separated from the FuncToAggregation for weekFunc. Note that the
	// recursive step is, at most, depth 1, because only used in 1-53
	// modes, but recursively calls with a 0-53 mode.
	WrapInZeroCheck := func(body interface{}, m int64) interface{} {
		lastDayLastYear := NewM(
			NewDocElem(OpDateFromParts, NewM(
				NewDocElem("year", WrapInOp(OpSubtract, year, 1)),
				NewDocElem("month", 12),
				NewDocElem("day", 31),
			)),
		)

		output := "$$output"
		letAssignment := NewM(
			NewDocElem("output", body),
		)

		return WrapInLet(letAssignment,
			WrapInCond(output,
				WrapInWeekCalculation(lastDayLastYear, m),
				WrapInOp(OpNeq, output, 0),
			),
		)
	}

	// wrapInFiftyThreeCheck is used to handle cases where the last week of a
	// year may actually map as the first week of the next year. This is
	// only possible in the cases where the first week is defined by having
	// 4 days in the year, and where 0 weeks are not allowed, so that is
	// modes 3 and 6. In these modes it is possible that 12-31, 12-30, and even
	// 12-29 map to week 1 of the next year. This is similar in design to
	// zeroCheck, except that it is only needed in the modes with 4 days
	// used to decide the first week of the month. We only need to check
	// the day if our computeDaySubtract results in week 53, giving us
	// faster common cases. janOneDaysOfWeek are the days of the week
	// for the next Jan-1 that result in one of the last three days
	// of the year potentially mapping to the next year. Note that
	// MongoDB aggregation pipeline numbers days 1-7, with 1 being Sunday.
	wrapInFiftyThreeCheck := func(body interface{}, janOneDaysOfWeek ...int) interface{} {
		output, day := "$$output", "$$day"
		outputLetAssignment := NewM(
			NewDocElem("output", body),
		)

		nextJan1 := getNextJan1()
		// Day Of Week for Jan 1  |  First Day In December Mapping to Next Year
		// --------------------------------------------------------------------
		// janOneDaysOfWeek[0]    |  29
		// janOneDaysOfWeek[1]    |  30
		// janOneDaysOfWeek[2]    |  31
		nextJan1DayOfWeek := WrapInOp(OpDayOfWeek, nextJan1)
		return WrapInLet(outputLetAssignment,
			WrapInCond(
				WrapInLet(NewM(
					NewDocElem("day", WrapInOp(OpDayOfMonth, date)),
				), WrapInSwitch(
					53,
					WrapInEqCase(nextJan1DayOfWeek,
						janOneDaysOfWeek[0],
						WrapInCond(1,
							53,
							WrapInOp(OpGte,
								day,
								29))),
					WrapInEqCase(nextJan1DayOfWeek,
						janOneDaysOfWeek[1],
						WrapInCond(1,
							53,
							WrapInOp(OpGte,
								day,
								30))),
					WrapInEqCase(nextJan1DayOfWeek,
						janOneDaysOfWeek[2],
						WrapInCond(1,
							53,
							WrapInOp(OpGte,
								day,
								31))),
				),
				),
				output,
				WrapInOp(OpEq, output, 53),
			),
		)
	}

	var body interface{}
	switch mode {
	// First day of week: Sunday, with a Sunday in this year.
	// This is what MongoDB's $week function does, so we use it.
	case 0, 2:
		body = WrapSingleArgFuncWithNullCheck(OpWeek, date)
		if mode == 2 {
			body = WrapInZeroCheck(body, 0)
		}
	// First day of week: Monday, with 4 days in this year.
	case 1, 3:
		body = generate4DaysBody(2)
		if mode == 3 {
			body = WrapInZeroCheck(body, 1)
			body = wrapInFiftyThreeCheck(body, 5, 4, 3)
		}
	// First day of week: Sunday, with 4 days in this year.
	case 4, 6:
		body = generate4DaysBody(1)
		if mode == 6 {
			body = WrapInZeroCheck(body, 4)
			body = wrapInFiftyThreeCheck(body, 4, 3, 2)
		}
	// First day of week: Monday, with a Monday in this year.
	case 5, 7:
		body = generateMondayBody()
		if mode == 7 {
			body = WrapInZeroCheck(body, 5)
		}
	}

	// Bind expressions that would be expensive to recompute.
	return WrapInLet(NewM(
		NewDocElem("date", expr),
	), WrapInLet(NewM(
		NewDocElem("year", WrapInOp(OpYear, date)),
	), body))
}

// WrapSingleArgFuncWithNullCheck returns a null checked version
// of the arg passed to name.
func WrapSingleArgFuncWithNullCheck(name string, arg interface{}) interface{} {
	return WrapInNullCheckedCond(nil, NewM(NewDocElem(name, arg)), arg)
}
