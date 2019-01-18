package evaluator

import (
	"fmt"
	"math"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/schema"
)

// FuncToAggregation for TO_DAYS has one issue wrt how TO_DAYS is supposed to perform:
// because our date treatment is backed by using MongoDB's $dateFromString function,
// if a date that doesn't exist (e.g., 0000-00-00 or 0001-02-29) is entered, we return
// an error instead of the NULL expected from MySQL. Unfortunately, checking for valid
// dates is too cost prohibitive. If at some point $dateFromString supports an onError/default
// value, we should switch to using that.
func (f *baseScalarFunctionExpr) toDaysToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(toDays)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}
	// Subtract dayOne (0000-01-01) from the argument in MongoDB, then convert ms to days.
	// When using $subtract on two dates in MongoDB, the number of ms between the two
	// dates is returned, and the purpose of the TO_DAYS function is to get the number
	// of days since 0000-01-01:
	// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_to-days
	// Unfortunately, we get a slightly wrong number if we try to multiply by days/ms
	// because MySQL itself is using division (and actually gets the wrong day count itself)
	// NOTE: args[0] must come in as a date creating expression, because we rewrite
	// to_days(x) in the algebrizer to to_days(date(x)).
	dayOne := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDivide, bsonutil.NewArray(
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
			args[0],
			dayOne,
		))),
		millisecondsPerDay,
	)),
	)),
	), nil
}

func (f *baseScalarFunctionExpr) absToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(abs)",
			fmt.Sprintf("expected 1 arguments, found %d", len(exprs)),
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.NewM(bsonutil.NewDocElem("$abs", args[0])), nil
}

func (f *baseScalarFunctionExpr) acosToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(acos)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	input := "$$input"
	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("input", args[0]),
	)

	// MySQL returns NULL for values outside of the range [-1,1].
	// asin x + acos x = pi/2
	return bsonutil.WrapInLet(letAssignment,
		bsonutil.WrapInCond(nil,
			bsonutil.WrapInAcosComputation(input),
			bsonutil.WrapInOp(bsonutil.OpLt, input, -1.0),
			bsonutil.WrapInOp(bsonutil.OpGt, input, 1.0),
		),
	), nil
}

func (f *baseScalarFunctionExpr) asinToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(asin)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	input := "$$input"
	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("input", args[0]),
	)

	// MySQL returns NULL for values outside of the range [-1,1].
	// asin(x) =  pi/2 - cos(x) via the identity:
	// asin(x) + acos(x) = pi/2.
	return bsonutil.WrapInLet(letAssignment,
		bsonutil.WrapInCond(nil,
			bsonutil.WrapInOp(bsonutil.OpSubtract, math.Pi/2.0, bsonutil.WrapInAcosComputation(input)),
			bsonutil.WrapInOp(bsonutil.OpLt, input, -1.0),
			bsonutil.WrapInOp(bsonutil.OpGt, input, 1.0),
		),
	), nil
}

func (f *baseScalarFunctionExpr) ceilToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(ceil)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCeil, args[0])), nil
}

func (f *baseScalarFunctionExpr) characterLengthToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(characterLength)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(characterLength)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$strLenCP", args[0]), nil
}

func (f *baseScalarFunctionExpr) concatToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) < 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(concat)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpConcat, args)), nil
}

func (f *baseScalarFunctionExpr) concatWsToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) < 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(concatWs)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	var pushArgs []interface{}

	for _, value := range args[1:] {
		pushArgs = append(pushArgs,
			bsonutil.WrapInNullCheckedCond(bsonutil.WrapInLiteral(""), value, value),
			bsonutil.WrapInNullCheckedCond(bsonutil.WrapInLiteral(""), args[0], value),
		)
	}

	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpConcat, pushArgs[:len(pushArgs)-1])), nil
}

func (f *baseScalarFunctionExpr) convToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {

	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(conv)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}
	num := args[0]
	oldBase := args[1]
	newBase := args[2]

	// length is how long (in digits) the input number is
	normalizedVars := bsonutil.NewM(
		bsonutil.NewDocElem("originalBase", bsonutil.WrapInOp(bsonutil.OpAbs, oldBase)),
		bsonutil.NewDocElem("newBase", bsonutil.WrapInOp(bsonutil.OpAbs, newBase)),
		bsonutil.NewDocElem("negative", bsonutil.WrapInOp(bsonutil.OpEq, "-", bsonutil.WrapInOp(bsonutil.OpSubstr, num, 0, 1))),
		bsonutil.NewDocElem("nonNegativeNumber", bsonutil.WrapInCond(
			bsonutil.WrapInOp(bsonutil.OpSubstr, num, 1,
				bsonutil.WrapInOp(bsonutil.OpSubtract, bsonutil.WrapInOp(bsonutil.OpStrlenCP, num), 1)),
			num,
			bsonutil.WrapInOp(bsonutil.OpEq, "-", bsonutil.WrapInOp(bsonutil.OpSubstr, num, 0, 1)))),
	)

	indexOfDecimal := bsonutil.NewM(
		bsonutil.NewDocElem("decimalIndex", bsonutil.WrapInOp(bsonutil.OpIndexOfCP, "$$nonNegativeNumber", ".")),
	)

	eliminateDecimal := bsonutil.NewM(
		bsonutil.NewDocElem("number", bsonutil.WrapInCond("$$nonNegativeNumber",
			bsonutil.WrapInOp(bsonutil.OpSubstr, "$$nonNegativeNumber", 0, "$$decimalIndex"),
			bsonutil.WrapInOp(bsonutil.OpEq, "$$decimalIndex", -1))),
	)

	createLength := bsonutil.NewM(
		bsonutil.NewDocElem("length", bsonutil.WrapInOp(bsonutil.OpStrlenCP, "$$number")),
	)

	// indexArr is an array of numbers from 0 to n-1 when n = length
	createIndexArr := bsonutil.NewM(
		bsonutil.NewDocElem("indexArr", bsonutil.WrapInOp(bsonutil.OpRange, 0, "$$length", 1)),
	)

	// charArr breaks the number entered into an array of characters where each char is a digit
	createCharArr := bsonutil.NewM(
		bsonutil.NewDocElem("charArr", bsonutil.WrapInMap("$$indexArr", "this",
			bsonutil.NewArray("$$this", bsonutil.WrapInOp(bsonutil.OpSubstr, "$$number", "$$this", 1)))),
	)

	// This logic takes in the charArr and outputs a 2D array containing the index and the
	// base10 numerical value of the character.
	// i.e. if charArr = ["3", "A", "2"], numArr = [[0, 3], [1, 10], [2, 2]]
	branches1 := bsonutil.NewMArray()
	for _, k := range validNumbers {
		branches1 = append(branches1,
			bsonutil.WrapInCase(
				bsonutil.WrapInOp(bsonutil.OpEq,
					bsonutil.WrapInOp(bsonutil.OpArrElemAt, "$$this", 1),
					k,
				),
				bsonutil.NewArray(
					bsonutil.WrapInOp(bsonutil.OpArrElemAt, "$$this", 0),
					stringToNum[k],
				),
			),
		)
	}
	createNumArr := bsonutil.NewM(
		bsonutil.NewDocElem("numArr", bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpMap, bsonutil.NewM(
				bsonutil.NewDocElem("input", "$$charArr"),
				bsonutil.NewDocElem("in", bsonutil.WrapInSwitch(bsonutil.NewArray(0, 100), branches1...)),
			)),
		)),
	)

	// invalidArr has False for every digit that is valid, and True for every digit that is invalid
	// In order for the input string to be converted to a new number base every entry in this
	// array must be False.
	createInvalidArr := bsonutil.NewM(
		bsonutil.NewDocElem("invalidArr", bsonutil.WrapInMap(
			"$$numArr",
			"this",
			bsonutil.WrapInOp(bsonutil.OpGte, bsonutil.WrapInOp(bsonutil.OpArrElemAt, "$$this", 1), "$$originalBase"),
		)),
	)

	// Given a charArr = [[1, x1]...[i, xi]...[n, xn]] and a base b,
	// This implements the logic: sum(b^(n-i-1) * xi) with i = 0->n-1
	generateBase10 := bsonutil.NewM(
		bsonutil.NewDocElem("base10", bsonutil.WrapInOp(bsonutil.OpSum,
			bsonutil.WrapInMap("$$numArr", "this",
				bsonutil.WrapInOp(bsonutil.OpMultiply,
					bsonutil.WrapInOp(bsonutil.OpArrElemAt, "$$this", 1),
					bsonutil.WrapInOp(bsonutil.OpPow, "$$originalBase",
						bsonutil.WrapInOp(bsonutil.OpSubtract,
							bsonutil.WrapInOp(bsonutil.OpSubtract, "$$length",
								bsonutil.WrapInOp(bsonutil.OpArrElemAt, "$$this", 0)),
							1)))))),
	)

	// numDigits is the length the number will be in the new number base
	// This is equal to: floor(log_newbase(num)) + 1
	numDigits := bsonutil.NewM(
		bsonutil.NewDocElem("numDigits", bsonutil.WrapInOp(bsonutil.OpAdd,
			bsonutil.WrapInOp(bsonutil.OpFloor,
				bsonutil.WrapInOp(bsonutil.OpLog, "$$base10", "$$newBase")), 1)),
	)

	// powers is an array of the powers of the base that you are translating to
	// if the newBase=16 and the resulting number will have length=4 this array
	// will = [1, 16, 256, 4096]
	powers := bsonutil.NewM(
		bsonutil.NewDocElem("powers", bsonutil.WrapInMap(
			bsonutil.WrapInOp(bsonutil.OpRange, bsonutil.WrapInOp(bsonutil.OpSubtract, "$$numDigits", 1), -1, -1),
			"this",
			bsonutil.WrapInOp(bsonutil.OpPow, "$$newBase", "$$this"))),
	)

	// Turns the base10 number into an array of the newBase digits (in their base10 form)
	// i.e. if base10 = 173 (0xAD), numbersArray = [10, 13]
	// Follows generalized version of: https://www.permadi.com/tutorial/numDecToHex/
	generateNumberArray := bsonutil.WrapInMap("$$powers", "this",
		bsonutil.WrapInOp(bsonutil.OpMod,
			bsonutil.WrapInOp(bsonutil.OpFloor,
				bsonutil.WrapInOp(bsonutil.OpDivide, "$$base10", "$$this")), "$$newBase"))

	branches2 := bsonutil.NewMArray()
	for k := 0; k <= len(numToString); k++ {
		branches2 = append(branches2,
			bsonutil.WrapInCase(bsonutil.WrapInOp(bsonutil.OpEq, "$$this", k), numToString[k]))
	}

	// Converts the number array into an array of their character representations
	// i.e. if numbersArray = [10, 13], then charArray=['A', 'D']
	generateCharArray := bsonutil.WrapInMap(generateNumberArray, "this", bsonutil.WrapInSwitch("0", branches2...))

	// Turns the charArray into a single string (the final answer)
	// i.e. if charArray=['A','D'] answer='AD'
	positiveAnswer := bsonutil.NewM(
		bsonutil.NewDocElem("positiveAnswer", bsonutil.WrapInReduce(
			generateCharArray,
			"",
			bsonutil.WrapInOp(bsonutil.OpConcat, "", "$$value", "$$this"),
		)),
	)

	signAdjusted := bsonutil.WrapInCond(bsonutil.WrapInOp(bsonutil.OpConcat, "-", "$$positiveAnswer"),
		"$$positiveAnswer", "$$negative")

	// Puts the nested lets together, checks to make sure that the base is valid,
	// and checks to make sure the entered number is valid as well
	// (invalid = numbers too big like 3 in binary or non-alphanumeric like /)
	// Invalid characters returns an answer of 0, invalid bases return NULL
	return bsonutil.WrapInCond(nil, bsonutil.WrapInLet(normalizedVars,
		bsonutil.WrapInLet(indexOfDecimal,
			bsonutil.WrapInLet(eliminateDecimal,
				bsonutil.WrapInCond(nil,
					bsonutil.WrapInCond("0",
						bsonutil.WrapInLet(createLength,
							bsonutil.WrapInLet(createIndexArr,
								bsonutil.WrapInLet(createCharArr,
									bsonutil.WrapInLet(createNumArr,
										bsonutil.WrapInLet(createInvalidArr,
											bsonutil.WrapInCond("0",
												bsonutil.WrapInLet(generateBase10,
													bsonutil.WrapInLet(numDigits,
														bsonutil.WrapInLet(powers,
															bsonutil.WrapInLet(positiveAnswer,
																signAdjusted)))),
												bsonutil.WrapInOp(bsonutil.OpAnyElementTrue,
													"$$invalidArr"))))))),
						bsonutil.WrapInOp(bsonutil.OpIn, "$$number", bsonutil.NewArray("0", "-0"))),
					bsonutil.WrapInOp(bsonutil.OpOr,
						bsonutil.WrapInOp(bsonutil.OpOr, bsonutil.WrapInOp(bsonutil.OpLt, "$$originalBase", 2),
							bsonutil.WrapInOp(bsonutil.OpGt, "$$originalBase", 36)),
						bsonutil.WrapInOp(bsonutil.OpOr, bsonutil.WrapInOp(bsonutil.OpLt, "$$newBase", 2),
							bsonutil.WrapInOp(bsonutil.OpGt, "$$newBase", 36)))))),
	), bsonutil.WrapInOp(bsonutil.OpEq, nil, num)), nil
}

func (f *baseScalarFunctionExpr) convertToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	typ, ok := evalTypeFromSQLExpr(exprs[1])
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(convert)",
			fmt.Sprintf(
				"cannot push down conversions to %s",
				exprs[1].(SQLValue).String(),
			),
		)
	}

	return NewSQLConvertExpr(exprs[0], typ).ToAggregationLanguage(t)
}

func (f *baseScalarFunctionExpr) cosToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(cos)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	input := "$$input"
	inputLetAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("input", bsonutil.WrapInOp(bsonutil.OpAbs, args[0])),
	)

	rem, phase := "$$rem", "$$phase"
	remPhaseAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("rem", bsonutil.WrapInOp(bsonutil.OpMod, input, math.Pi/2)),
		bsonutil.NewDocElem("phase", bsonutil.WrapInOp(bsonutil.OpMod,
			bsonutil.WrapInOp(bsonutil.OpTrunc,
				bsonutil.WrapInOp(bsonutil.OpDivide, input, math.Pi/2),
			),
			4.0)),
	)

	// 3.2 does not support $switch, so just use chained $cond, assuming
	// zeroCase will be most common (since it's the first phase). Because we
	// use the Maclaurin Power Series for sine and cos, we need to adjust
	// our input into a domain that is good for our approximation, that
	// being the first quadrant (phase). For phases outside of the first,
	// we can adjust the functions as:
	//
	// phase | Maclaurin Power Series
	// ------------------------------
	// 0     | cos(rem)
	// 1     | -1 * sin(rem)
	// 2     | -1 * cos(rem)
	// 3     | sin(rem)
	// where the phase is defined as the trunc(input / (pi/2)) % 4
	// and the remainder is input % (pi/2).
	threeCase := bsonutil.WrapInCond(bsonutil.WrapInSinPowerSeries(rem),
		nil,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			3))
	twoCase := bsonutil.WrapInCond(bsonutil.WrapInOp(bsonutil.OpMultiply,
		-1.0,
		bsonutil.WrapInCosPowerSeries(rem)),
		threeCase,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			2))
	oneCase := bsonutil.WrapInCond(bsonutil.WrapInOp(bsonutil.OpMultiply,
		-1.0,
		bsonutil.WrapInSinPowerSeries(rem)),
		twoCase,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			1))
	zeroCase := bsonutil.WrapInCond(bsonutil.WrapInCosPowerSeries(rem),
		oneCase,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			0))

	return bsonutil.WrapInLet(inputLetAssignment,
		bsonutil.WrapInLet(remPhaseAssignment,
			zeroCase),
	), nil
}

func (f *baseScalarFunctionExpr) cotToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	denom, err := f.sinToAggregationLanguage(t, []SQLExpr{exprs[0]})
	if err != nil {
		return nil, err
	}

	num, err := f.cosToAggregationLanguage(t, []SQLExpr{exprs[0]})
	if err != nil {
		return nil, err
	}

	// epsilon the smallest value we allow for denom, computed to roughly
	// tie-out with mysqld.
	epsilon := 6.123233995736766e-17
	// Replace abs(denom) < epsilon with epsilon since mysql doesn't seem
	// to return NULL or INF for inf values of tan as one would expect, and
	// we do not want to trigger a $divide by 0.
	return bsonutil.WrapInOp(bsonutil.OpDivide,
		num,
		bsonutil.WrapInCond(epsilon,
			denom,
			bsonutil.WrapInOp(bsonutil.OpLte,
				bsonutil.WrapInOp(bsonutil.OpAbs, denom), epsilon,
			),
		),
	), nil
}

// nolint: unparam
func (f *baseScalarFunctionExpr) currentDateToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	now := time.Now().In(schema.DefaultLocale)
	cd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, schema.DefaultLocale)
	return bsonutil.WrapInLiteral(cd), nil
}

// nolint: unparam
func (f *baseScalarFunctionExpr) currentTimestampToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	now := time.Now().In(schema.DefaultLocale)
	return bsonutil.WrapInLiteral(now), nil
}

func (f *baseScalarFunctionExpr) dateAddToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	return f.dateArithmeticToAggregationLanguage(t, exprs, false)
}

func (f *baseScalarFunctionExpr) dateArithmeticToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr, isSub bool) (interface{}, PushdownFailure) {
	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateArithmetic)",
			incorrectArgCountMsg,
		)
	}

	var date interface{}
	var ok bool
	var err PushdownFailure
	if !isSub {
		// implementation for ADDDATE(DATE_FORMAT("..."), INTERVAL 0 SECOND)
		var fun *dateFormatFunc
		if fun, ok = exprs[0].(*dateFormatFunc); ok {
			var dateErr error
			if date, dateErr = t.translateDateFormatAsDate(fun); dateErr != nil {
				date = nil
			}
		}
	}

	if date == nil {
		switch exprs[0].EvalType() {
		case EvalDate, EvalDatetime:
		default:
			return nil, newPushdownFailure(
				"SQLScalarFunctionExpr(dateArithmetic)",
				"cannot push down when first arg is EvalDate or EvalDatetime",
			)
		}

		if date, err = t.ToAggregationLanguage(exprs[0]); err != nil {
			return nil, err
		}
	}

	intervalValue, ok := exprs[1].(SQLValue)
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateArithmetic)",
			"cannot push down without literal interval value",
		)
	}

	if Float64(intervalValue) == 0 {
		return date, nil
	}

	unitValue, ok := exprs[2].(SQLValue)
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateArithmetic)",
			"cannot push down without literal unit value",
		)
	}

	var ms int64
	// Second can be a float rather than an int, so handle Second specially.
	// calculateInterval works for all other units, as they must be integral.
	if unitValue.String() == Second {
		ms = round(Float64(intervalValue) * 1000.0)
	} else {
		unitInterval, neg := dateArithmeticArgs(unitValue.String(), intervalValue)
		unit, interval, err := calculateInterval(unitValue.String(), unitInterval, neg)
		if err != nil {
			return nil, newPushdownFailure(
				"SQLScalarFunctionExpr(dateArithmetic)",
				"failed to calculate interval",
				"error", err.Error(),
			)
		}
		ms, err = unitIntervalToMilliseconds(unit, int64(interval))
		if err != nil {
			return nil, newPushdownFailure(
				"SQLScalarFunctionExpr(dateArithmetic)",
				"failed to convert interval to ms",
				"error", err.Error(),
			)
		}
	}

	if isSub {
		ms *= -1
	}

	conds := bsonutil.NewArray()
	if _, ok := bsonutil.GetLiteral(date); !ok {
		conds = append(conds, "$$date")
	}

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("date", date),
	)

	letEvaluation := bsonutil.WrapInNullCheckedCond(
		nil,
		bsonutil.WrapInOp(bsonutil.OpAdd, "$$date", ms),
		conds...,
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil
}

func (f *baseScalarFunctionExpr) dateDiffToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateDiff)",
			incorrectArgCountMsg,
		)
	}

	var date1, date2 interface{}
	var ok bool
	var err PushdownFailure

	parseArgs := func(expr SQLExpr) (interface{}, PushdownFailure) {
		var value SQLValue
		if value, ok = expr.(SQLValue); ok {
			var date time.Time
			date, _, ok = strToDateTime(value.String(), false)
			if !ok {
				return nil, newPushdownFailure(
					"SQLScalarFunctionExpr(dateDiff)",
					"failed to parse datetime from literal",
				)
			}

			date = time.Date(date.Year(),
				date.Month(),
				date.Day(),
				0,
				0,
				0,
				0,
				schema.DefaultLocale)
			return date, nil
		}
		exprType := expr.EvalType()
		if exprType == EvalDatetime || exprType == EvalDate {
			var date interface{}
			date, err = t.ToAggregationLanguage(expr)
			if err != nil {
				return nil, err
			}
			return date, nil
		}
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateDiff)",
			"argument was not a SQLValue, EvalDate, or EvalDatetime",
		)
	}

	if date1, err = parseArgs(exprs[0]); err != nil {
		return nil, err
	}
	if date2, err = parseArgs(exprs[1]); err != nil {
		return nil, err
	}

	// This division needs to truncate because this is dateDiff not
	// timestampDiff, partial days are dropped.
	days := bsonutil.WrapInOp(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide,
		bsonutil.WrapInOp(bsonutil.OpSubtract, date1, date2), 86400000))
	bound := bsonutil.WrapInCond(106751, -106751, bsonutil.WrapInOp(bsonutil.OpGt, days, 106751))

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("days", days),
	)

	letEvaluation := bsonutil.WrapInNullCheckedCond(
		nil,
		bsonutil.WrapInCond(
			bound,
			"$$days",
			bsonutil.WrapInOp(bsonutil.OpGt, "$$days", 106751),
			bsonutil.WrapInOp(bsonutil.OpLt, "$$days", -106751),
		),
		date1,
		date2,
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil
}

func (f *baseScalarFunctionExpr) dateFormatToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateFormat)",
			incorrectArgCountMsg,
		)
	}

	date, err := t.ToAggregationLanguage(exprs[0])
	if err != nil {
		return nil, err
	}

	formatValue, ok := exprs[1].(SQLValue)
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateFormat)",
			"format string was not a literal",
		)
	}

	wrapped, ok := bsonutil.WrapInDateFormat(date, formatValue.String())
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateFormat)",
			"unable to push down format string",
			"formatString", formatValue.String(),
		)
	}
	return wrapped, nil
}

func (f *baseScalarFunctionExpr) dateSubToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	return f.dateArithmeticToAggregationLanguage(t, exprs, true)
}

func (f *baseScalarFunctionExpr) dayNameToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dayName)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpArrElemAt, bsonutil.NewArray(
				bsonutil.NewArray(
					time.Sunday.String(),
					time.Monday.String(),
					time.Tuesday.String(),
					time.Wednesday.String(),
					time.Thursday.String(),
					time.Friday.String(),
					time.Saturday.String(),
				),
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
					bsonutil.NewM(bsonutil.NewDocElem("$dayOfWeek", args[0])),
					1,
				))),
			)),
		), args[0],
	), nil
}

func (f *baseScalarFunctionExpr) dayOfMonthToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dayOfMonth)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$dayOfMonth", args[0]), nil
}

func (f *baseScalarFunctionExpr) dayOfWeekToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dayOfWeek)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$dayOfWeek", args[0]), nil
}

func (f *baseScalarFunctionExpr) dayOfYearToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dayOfYear)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$dayOfYear", args[0]), nil
}

func (f *baseScalarFunctionExpr) degreesToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(degrees)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInOp(bsonutil.OpDivide, bsonutil.WrapInOp(bsonutil.OpMultiply, args[0], 180.0), math.Pi), nil
}

func (f *baseScalarFunctionExpr) expToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(exp)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.NewM(bsonutil.NewDocElem("$exp", args[0])), nil
}

func (f *baseScalarFunctionExpr) extractToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(extract)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	bsonMap, ok := args[0].(bson.M)
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(extract)",
			"translateArgs returned something other than bson.M",
		)
	}

	bsonVal, ok := bsonMap["$literal"]
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(extract)",
			"first argument was not translated to a $literal",
		)
	}

	unit, ok := bsonVal.(string)
	if !ok {
		// The unit must absolutely be a string.
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(extract)",
			"first argument was not a string",
		)
	}

	switch unit {
	case "year", "month", "hour", "minute", "second":
		return bsonutil.WrapSingleArgFuncWithNullCheck("$"+unit, args[1]), nil
	case "day":
		return bsonutil.WrapSingleArgFuncWithNullCheck("$dayOfMonth", args[1]), nil
	}
	return nil, newPushdownFailure(
		"SQLScalarFunctionExpr(extract)",
		"unknown unit",
		"unit", unit,
	)
}

func (f *baseScalarFunctionExpr) floorToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(floor)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.NewM(bsonutil.NewDocElem("$floor", args[0])), nil
}

func (f *baseScalarFunctionExpr) fromDaysToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(fromDays)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}
	dayOne := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	body := bsonutil.WrapInOp(bsonutil.OpAdd, dayOne,
		bsonutil.WrapInOp(bsonutil.OpMultiply, bsonutil.WrapInRound(args[0]), millisecondsPerDay))
	arg := "$$arg"

	argLetAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("arg", args[0]),
	)

	// This should return "0000-00-00" if the input is too large (> maxFromDays)
	// or too low (< 366).
	return bsonutil.WrapInLet(argLetAssignment, bsonutil.WrapInNullCheckedCond(nil,
		bsonutil.WrapInCond(0,
			body,
			bsonutil.WrapInOp(bsonutil.OpGt, arg, maxFromDays),
			bsonutil.WrapInOp(bsonutil.OpLt, arg, 366),
		),
		arg,
	),
	), nil
}

func (f *baseScalarFunctionExpr) fromUnixtimeToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) > 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(fromUnixtime)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	arg := "$$arg"
	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("arg", args[0]),
	)

	// Just add the argument to 1970-01-01 00:00:00.0000000.
	dayOne := time.Date(1970, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
	letEvaluation := bsonutil.WrapInOp(bsonutil.OpAdd,
		dayOne,
		bsonutil.WrapInOp(bsonutil.OpMultiply,
			bsonutil.WrapInRound(arg),
			1e3))

	ret := bsonutil.WrapInLet(letAssignment,
		bsonutil.WrapInCond(nil,
			letEvaluation,
			bsonutil.WrapInOp(bsonutil.OpLt, arg, bsonutil.WrapInLiteral(0)),
		),
	)

	if len(exprs) == 1 {
		return ret, nil
	}
	if format, ok := exprs[1].(SQLValue); ok {
		wrapped, ok := bsonutil.WrapInDateFormat(ret, format.String())
		if !ok {
			return nil, newPushdownFailure(
				"SQLScalarFunctionExpr(fromUnixtime)",
				"unable to push down format string",
				"formatString", format.String(),
			)
		}
		return wrapped, nil
	}

	return nil, newPushdownFailure(
		"SQLScalarFunctionExpr(fromUnixtime)",
		"unsupported form",
	)
}

func (f *baseScalarFunctionExpr) greatestToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	// we can only push down if the types are similar
	for i := 1; i < len(exprs); i++ {
		if !isSimilar(exprs[0].EvalType(), exprs[i].EvalType()) {
			return nil, newPushdownFailure("SQLScalarFunctionExpr(greatest)", "arguments' types are not similar")
		}
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(bsonutil.NewDocElem("$max", args)), args...,
	), nil
}

func (f *baseScalarFunctionExpr) hourToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(hour)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$hour", args[0]), nil
}

func (f *baseScalarFunctionExpr) insertToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(insert)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 4 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(insert)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	str, pos, len, newstr := "$$str", "$$pos", "$$len", "$$newstr"
	inputAssignment := bsonutil.NewM(
		// SQL uses 1 indexing, so makes sure to subtract 1 to
		// account for MongoDB's 0 indexing.
		bsonutil.NewDocElem("str", args[0]),
		bsonutil.NewDocElem("pos", bsonutil.WrapInRound(bsonutil.WrapInOp(bsonutil.OpSubtract, args[1], 1))),
		bsonutil.NewDocElem("len", bsonutil.WrapInRound(args[2])),
		bsonutil.NewDocElem("newstr", args[3]),
	)

	totalLength := "$$totalLength"
	totalLengthAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("totalLength", bsonutil.WrapInOp(bsonutil.OpStrlenCP, str)),
	)

	prefix, suffix := "$$prefix", "$$suffix"
	ixAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("prefix", bsonutil.WrapInOp(bsonutil.OpSubstr, str, 0, pos)),
		bsonutil.NewDocElem("suffix", bsonutil.WrapInOp(bsonutil.OpSubstr, str, bsonutil.WrapInOp(bsonutil.OpAdd, pos, len), totalLength)),
	)

	concatenation := bsonutil.WrapInLet(ixAssignment,
		bsonutil.WrapInOp(bsonutil.OpConcat, prefix, newstr, suffix),
	)

	posCheck := bsonutil.WrapInLet(totalLengthAssignment,
		bsonutil.WrapInCond(str,
			concatenation,
			bsonutil.WrapInOp(bsonutil.OpLte, pos, 0),
			bsonutil.WrapInOp(bsonutil.OpGte, pos, totalLength),
		),
	)

	return bsonutil.WrapInLet(inputAssignment,
		bsonutil.WrapInCond(nil,
			posCheck,
			bsonutil.WrapInOp(bsonutil.OpLte, str, nil),
			bsonutil.WrapInOp(bsonutil.OpLte, pos, nil),
			bsonutil.WrapInOp(bsonutil.OpLte, len, nil),
			bsonutil.WrapInOp(bsonutil.OpLte, newstr, nil),
		),
	), nil
}

func (f *baseScalarFunctionExpr) instrToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(instr)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(instr)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	// Mongo Aggregation Pipeline returns NULL if arg1 is NULLish, like
	// we'd want. arg2 being NULL, however, is an error in the pipeline,
	// thus check arg2 for NULLisness.
	arg2 := "$$arg2"
	return bsonutil.WrapInLet(bsonutil.NewM(
		bsonutil.NewDocElem("arg2", args[1]),
	), bsonutil.WrapInCond(nil,
		bsonutil.WrapInOp(bsonutil.OpAdd,
			bsonutil.WrapInOp(bsonutil.OpIndexOfCP, args[0], arg2),
			1,
		),
		bsonutil.WrapInOp(bsonutil.OpLte, arg2, nil),
	),
	), nil
}

func (f *baseScalarFunctionExpr) lastDayToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(lastDay)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	date := "$$date"
	outerLetAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("date", args[0]),
	)

	letAssigment := bsonutil.NewM(
		bsonutil.NewDocElem("year", bsonutil.WrapInOp(bsonutil.OpYear, date)),
		bsonutil.NewDocElem("month", bsonutil.WrapInOp(bsonutil.OpMonth, date)),
	)

	year, month := "$$year", "$$month"
	var letEvaluation bson.M

	// Underflow and overflow in date computation are supported on MongoDB versions >= 4.0.
	// For example, a month value greater than 12 (overflow) and a day value of zero
	// (underflow) are supported date values.
	if t.versionAtLeast(4, 0, 0) {
		// MongoDB interprets day 0 of a given month as the last day of the previous month.
		letEvaluation = bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpDateFromParts, bsonutil.NewM(
				bsonutil.NewDocElem("year", year),
				bsonutil.NewDocElem("month", bsonutil.WrapInOp(bsonutil.OpAdd, 1, month)),
				bsonutil.NewDocElem("day", 0),
			)),
		)

	} else {

		// For MongoDB versions < 4.0, underflow and overflow in date computation are not
		// supported. For example, a day value of zero or a month value of 13 in a date
		// generates an error. In this case, we create a switch on the month value,
		// extracted from $dateFromParts, to determine the last day of the month.
		letEvaluation = bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpDateFromParts, bsonutil.NewM(
				bsonutil.NewDocElem("year", year),
				bsonutil.NewDocElem("month", month),
				bsonutil.NewDocElem("day",
					// The following MongoDB aggregation language implements this go code,
					// which is designed to set the day of a date to the last day of the month.
					// switch month {
					// case 2:
					// 	if isLeapYear(year) == 0 {
					// 		day = 29
					//	} else {
					//		day = 28
					//	}
					// case 4, 6, 9, 11:
					//	day = 30
					// default:
					//      day = 31
					// }
					bsonutil.WrapInSwitch(31,
						bsonutil.WrapInEqCase(month, 2,
							bsonutil.WrapInCond(29, 28, bsonutil.WrapInIsLeapYear(year)),
						),
						bsonutil.WrapInEqCase(month, 4, 30),
						bsonutil.WrapInEqCase(month, 6, 30),
						bsonutil.WrapInEqCase(month, 9, 30),
						bsonutil.WrapInEqCase(month, 11, 30),
					)),
			)),
		)

	}

	return bsonutil.WrapInLet(outerLetAssignment, bsonutil.WrapInLet(letAssigment, letEvaluation)), nil
}

func (f *baseScalarFunctionExpr) lcaseToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(lcase)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$toLower", args[0]), nil
}

func (f *baseScalarFunctionExpr) leastToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	// we can only push down if the types are similar
	for i := 1; i < len(exprs); i++ {
		if !isSimilar(exprs[0].EvalType(), exprs[i].EvalType()) {
			return nil, newPushdownFailure("SQLScalarFunctionExpr(least)", "arguments' types are not similar")
		}
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(bsonutil.NewDocElem("$min", args)), args...,
	), nil

}

func (f *baseScalarFunctionExpr) leftToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(left)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(left)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	conds := bsonutil.NewArray()
	var subStrLength interface{}

	if stringValue, ok := bsonutil.GetLiteral(args[0]); ok {
		if stringValue == nil {
			return bsonutil.MgoNullLiteral, nil
		}
	} else {
		conds = append(conds, "$$string")
	}

	if lengthValue, ok := bsonutil.GetLiteral(args[1]); ok {
		if lengthValue == nil {
			return bsonutil.MgoNullLiteral, nil
		}

		// when length is negative, just use 0. round length to closest integer
		if i, ok := lengthValue.(int64); ok {
			args[1] = bsonutil.WrapInLiteral(int64(math.Max(0, float64(i))))
			subStrLength = "$$length"
		} else {
			args[1] = bsonutil.WrapInRound(bsonutil.WrapInOp(bsonutil.OpMax, args[1], 0))
			subStrLength = "$$length"
		}
	} else {
		subStrLength = bsonutil.WrapInRound(bsonutil.WrapInOp(bsonutil.OpMax, "$$length", 0))
		conds = append(conds, "$$length")
	}

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("string", args[0]),
		bsonutil.NewDocElem("length", args[1]),
	)

	subStrOp := bsonutil.WrapInOp(bsonutil.OpSubstr, "$$string", 0, subStrLength)

	letEvaluation := bsonutil.WrapInNullCheckedCond(nil, subStrOp, conds...)
	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil
}

func (f *baseScalarFunctionExpr) lengthToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(length)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(length)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$strLenBytes", args[0]), nil
}

func (f *baseScalarFunctionExpr) locateToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(locate)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if !(len(exprs) == 2 || len(exprs) == 3) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(locate)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	var locate interface{}
	substr := args[0]
	str := args[1]

	if len(args) == 2 {
		indexOfCP := bsonutil.NewM(bsonutil.NewDocElem("$indexOfCP", bsonutil.NewArray(
			str,
			substr,
		)))
		locate = bsonutil.WrapInOp(bsonutil.OpAdd, indexOfCP, 1)
	} else if len(args) == 3 {
		// if the pos arg is null, we should return 0, not null
		// this is the same result as when the arg is 0
		pos := bsonutil.WrapInIfNull(args[2], 0)

		// round to the nearest int
		pos = bsonutil.WrapInOp(bsonutil.OpAdd, pos, 0.5)
		pos = bsonutil.WrapInOp(bsonutil.OpTrunc, pos)

		// subtract 1 from the pos arg to reconcile indexing style
		pos = bsonutil.WrapInOp(bsonutil.OpSubtract, pos, 1)

		indexOfCP := bsonutil.NewM(bsonutil.NewDocElem("$indexOfCP", bsonutil.NewArray(
			str,
			substr,
			pos,
		)))
		locate = bsonutil.WrapInOp(bsonutil.OpAdd, indexOfCP, 1)

		// if the pos argument was negative, we should return 0
		locate = bsonutil.WrapInCond(
			0,
			locate,
			bsonutil.WrapInOp(bsonutil.OpLt, pos, 0),
		)
	}

	return bsonutil.WrapInNullCheckedCond(
		nil,
		locate,
		str, substr,
	), nil
}

func (f *baseScalarFunctionExpr) log10ToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	return f.logarithmToAggregationLanguage(t, exprs, 10)
}

func (f *baseScalarFunctionExpr) log2ToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	return f.logarithmToAggregationLanguage(t, exprs, 2)
}

func (f *baseScalarFunctionExpr) logToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	return f.logarithmToAggregationLanguage(t, exprs, 0)
}

func (f *baseScalarFunctionExpr) logarithmToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr, base uint32) (interface{}, PushdownFailure) {
	if len(exprs) < 1 || len(exprs) > 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(log)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	// Use ln func rather than log with go's value for E, to avoid compromising values
	// more than we already do between MongoDB and MySQL by introducing a third value for E
	// (i.e., go's)
	if base == 0 {
		// 1 arg implies natural log
		if len(args) == 1 {
			return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCond, bsonutil.NewArray(
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGt, bsonutil.NewArray(
					args[0],
					0,
				))),
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpNaturalLog, args[0])),
				bsonutil.MgoNullLiteral,
			))), nil
		}
		// Two args is based arg.
		// MySQL specifies base then arg, MongoDB expects arg then base, so we have to flip.
		return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCond, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGt, bsonutil.NewArray(
				args[0],
				0,
			))),
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpLog, bsonutil.NewArray(
				args[1],
				args[0],
			))),
			bsonutil.MgoNullLiteral,
		))), nil
	}
	// This will be base 10 or base 2 based on if log10 or log2 was called.
	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCond, bsonutil.NewArray(
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGt, bsonutil.NewArray(
			args[0],
			0,
		))),
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpLog, bsonutil.NewArray(
			args[0],
			base,
		))),
		bsonutil.MgoNullLiteral,
	))), nil
}

func (f *baseScalarFunctionExpr) lpadToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	return f.padToAggregationLanguage(t, exprs, true)
}

func (f *baseScalarFunctionExpr) ltrimToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(ltrim)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(ltrim)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	if t.versionAtLeast(4, 0, 0) {
		return bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpLTrim, bsonutil.NewM(
				bsonutil.NewDocElem("input", args[0]),
				bsonutil.NewDocElem("chars", " "),
			)),
		), nil
	}

	ltrimCond := bsonutil.WrapInCond(
		"",
		bsonutil.WrapInLRTrim(true, args[0]), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
			args[0],
			"",
		))),
	)

	return bsonutil.WrapInNullCheckedCond(
		nil,
		ltrimCond,
		args[0],
	), nil
}

func (f *baseScalarFunctionExpr) makeDateToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 5, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(makeDate)",
			"cannot push down to MongoDB < 3.5",
		)
	}

	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(makeDate)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	year, day, paddedYear, output := "$$year", "$$day", "$$paddedYear", "$$output"

	inputLetStatement := bsonutil.NewM(
		bsonutil.NewDocElem("year", bsonutil.WrapInRound(args[0])),
		bsonutil.NewDocElem("day", bsonutil.WrapInRound(args[1])),
	)

	branch1900 := bsonutil.WrapInCond(
		bsonutil.WrapInOp(bsonutil.OpAdd, year, 1900),
		year,
		bsonutil.WrapInOp(bsonutil.OpAnd,
			bsonutil.WrapInOp(bsonutil.OpGte, year, 70),
			bsonutil.WrapInOp(bsonutil.OpLte, year, 99),
		))

	branch2000 := bsonutil.WrapInOp(bsonutil.OpAdd, year, 2000)

	// $$paddedYear holds the year + 2000 for years between 0 and 69, and +
	// 1900 for years between 70 and 99. Otherwise, it is the original
	// year.
	paddedYearLetStatement := bsonutil.NewM(bsonutil.NewDocElem("paddedYear", bsonutil.WrapInCond(branch2000, branch1900,
		bsonutil.WrapInOp(bsonutil.OpAnd,
			bsonutil.WrapInOp(bsonutil.OpGte,
				year,
				0),
			bsonutil.WrapInOp(bsonutil.OpLte,
				year,
				69)),
	)),
	)

	// This implements:
	// date(paddedYear) + (day - 1) * millisecondsPerDay.
	addDaysStatement := bsonutil.WrapInOp(
		bsonutil.OpAdd,
		bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpDateFromParts,
				bsonutil.NewM(bsonutil.NewDocElem("year", paddedYear)))),
		bsonutil.WrapInOp(bsonutil.OpMultiply,
			bsonutil.WrapInOp(bsonutil.OpSubtract, day, 1),
			millisecondsPerDay),
	)

	// If the $$paddedYear is more than 9999 or less than 0, return NULL.
	yearRangeCheck := bsonutil.WrapInCond(
		nil,
		addDaysStatement,
		bsonutil.WrapInOp(bsonutil.OpLt, paddedYear, 0),
		bsonutil.WrapInOp(bsonutil.OpGt, paddedYear, 9999),
	)

	// Day range check, return NULL if day < 1.
	dayRangeCheck := bsonutil.WrapInCond(nil,
		yearRangeCheck,
		bsonutil.WrapInOp(bsonutil.OpLt, day, 1),
	)

	outputLetStatement := bsonutil.NewM(bsonutil.NewDocElem("output", dayRangeCheck))

	// Bind lets, and check that output value year < 9999, otherwise MySQL
	// returns NULL.
	return bsonutil.WrapInLet(inputLetStatement,
		bsonutil.WrapInLet(paddedYearLetStatement,
			bsonutil.WrapInLet(outputLetStatement,
				bsonutil.WrapInCond(nil, output,
					bsonutil.WrapInOp(bsonutil.OpGt,
						bsonutil.WrapInOp(bsonutil.OpYear, output),
						9999))),
		)), nil

}

func (f *baseScalarFunctionExpr) microsecondToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(microsecond)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMultiply, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem("$millisecond", args[0])),
			1000,
		)),
		), args[0],
	), nil

}

func (f *baseScalarFunctionExpr) midToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(mid)",
			incorrectArgCountMsg,
		)
	}
	return f.substringToAggregationLanguage(t, exprs)
}

func (f *baseScalarFunctionExpr) minuteToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(minute)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$minute", args[0]), nil
}

func (f *baseScalarFunctionExpr) modToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(mod)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.NewM(bsonutil.NewDocElem("$mod", bsonutil.NewArray(
		args[0],
		args[1],
	))), nil
}

func (f *baseScalarFunctionExpr) monthNameToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(monthName)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpArrElemAt, bsonutil.NewArray(
				bsonutil.NewArray(
					time.January.String(),
					time.February.String(),
					time.March.String(),
					time.April.String(),
					time.May.String(),
					time.June.String(),
					time.July.String(),
					time.August.String(),
					time.September.String(),
					time.October.String(),
					time.November.String(),
					time.December.String(),
				),
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
					bsonutil.NewM(bsonutil.NewDocElem("$month", args[0])),
					1,
				))),
			)),
		), args[0],
	), nil
}

func (f *baseScalarFunctionExpr) monthToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(month)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$month", args[0]), nil
}

func (f *baseScalarFunctionExpr) padToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr, isLeftPad bool) (interface{}, PushdownFailure) {

	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(pad)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(pad)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	// arguments to lpad
	str := args[0]
	lengthVal := args[1]
	padStr := args[2]

	// round to nearest int.
	length := bsonutil.WrapInRound(lengthVal)

	// variables for $let expression - length of padding needed
	// and length of input padding strings
	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("padLen", bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
				length,
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, str)),
			)))),
		bsonutil.NewDocElem("padStrLen", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, padStr))),
		bsonutil.NewDocElem("length", length),
	)

	// logic for generating padding string:

	// do we even need to add padding? only if the desired output
	// length is > length of input string.
	paddingCond := bsonutil.NewM(
		bsonutil.NewDocElem(bsonutil.OpLt, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, str)),
			"$$length",
		)))

	// number of times we need to repeat the padding string to fill space
	padStrRepeats := bsonutil.NewM(
		bsonutil.NewDocElem(bsonutil.OpCeil, bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpDivide, bsonutil.NewArray(
				"$$padLen",
				"$$padStrLen",
			)))))

	// generate an array with padStrRepeats occurrences of padStr
	padParts := bsonutil.NewM(
		bsonutil.NewDocElem(bsonutil.OpMap, bsonutil.NewM(
			bsonutil.NewDocElem("input", bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpRange, bsonutil.NewArray(
					0,
					padStrRepeats,
				)),
			)),
			bsonutil.NewDocElem("in", padStr),
		)))

	// join occurrences together and trim to the exact length needed
	fullPad := bsonutil.NewM(
		bsonutil.NewDocElem(bsonutil.OpSubstr, bsonutil.NewArray(
			bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpReduce, bsonutil.NewM(
					bsonutil.NewDocElem("input", padParts),
					bsonutil.NewDocElem("initialValue", ""),
					bsonutil.NewDocElem("in", bsonutil.NewM(
						bsonutil.NewDocElem(bsonutil.OpConcat, bsonutil.NewArray(
							"$$value",
							"$$this",
						)))),
				))),
			0,
			"$$padLen",
		)),
	)

	// based on length of input string, we either add the padding
	// or just take appropriate substring of input string
	var concatted bson.M
	if isLeftPad {
		concatted = bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpConcat, bsonutil.NewArray(
			fullPad,
			str,
		)))
	} else {
		concatted = bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpConcat, bsonutil.NewArray(
			str,
			fullPad,
		)))
	}

	handleConcat := bsonutil.WrapInCond(
		nil,
		concatted, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
			"$$padStrLen",
			0,
		))))

	// handle everything in the case that input length >=0
	handleNonNegativeLength := bsonutil.WrapInCond(
		handleConcat, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubstr, bsonutil.NewArray(
			str,
			0,
			"$$length",
		))), paddingCond)

	// whether the input length is < 0
	lengthIsNegative := bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpLt, bsonutil.NewArray(
		length,
		0,
	)))

	// if it's < 0, then we just want to return null
	negativeCheck := bsonutil.WrapInCond(nil, handleNonNegativeLength, lengthIsNegative)

	return bsonutil.WrapInNullCheckedCond(
		nil,
		bsonutil.WrapInLet(letAssignment, negativeCheck),
		str, lengthVal, padStr,
	), nil
}

func (f *baseScalarFunctionExpr) powToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(pow)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInOp(bsonutil.OpPow, args[0], args[1]), nil
}

func (f *baseScalarFunctionExpr) quarterToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(quarter)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	conds := bsonutil.NewArray()
	if _, ok := bsonutil.GetLiteral(args[0]); !ok {
		conds = append(conds, "$$date")
	}

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("date", args[0]),
	)

	letEvaluation := bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpArrElemAt, bsonutil.NewArray(
				bsonutil.NewArray(1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4),
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
					bsonutil.NewM(bsonutil.NewDocElem("$month", "$$date")),
					1,
				))),
			))), conds...,
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

func (f *baseScalarFunctionExpr) radiansToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(radians)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInOp(bsonutil.OpDivide, bsonutil.WrapInOp(bsonutil.OpMultiply, args[0], math.Pi), 180.0), nil
}

func (f *baseScalarFunctionExpr) repeatToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(repeat)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(repeat)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	str := args[0]

	// num must be rounded to match mysql
	num := bsonutil.WrapInRound(args[1])

	// create array w/ args[1] values e.g. [0,1,2]
	rangeArr := bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpRange, bsonutil.NewArray(
		0,
		num,
		1,
	)))

	// create array of len arg[1], with each item being arg[0]
	mapArgs := bsonutil.NewM(bsonutil.NewDocElem("input", rangeArr), bsonutil.NewDocElem("in", str))
	mapWithArgs := bsonutil.NewM(bsonutil.NewDocElem("$map", mapArgs))

	// append all values of this array together
	inArg := bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpConcat, bsonutil.NewArray(
		"$$this",
		"$$value",
	)))
	reduceArgs := bsonutil.NewM(bsonutil.NewDocElem("input", mapWithArgs), bsonutil.NewDocElem("initialValue", ""), bsonutil.NewDocElem("in", inArg))

	repeat := bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpReduce, reduceArgs))

	return bsonutil.WrapInNullCheckedCond(nil, repeat, str, num), nil

}

func (f *baseScalarFunctionExpr) replaceToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(replace)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(replace)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	split := "$$split"
	assignment := bsonutil.NewM(
		bsonutil.NewDocElem("split", bsonutil.WrapInOp(bsonutil.OpSplit, args[0], args[1])),
	)

	this, value := "$$this", "$$value"
	body := bsonutil.WrapInReduce(split,
		nil,
		bsonutil.WrapInCond(this,
			bsonutil.WrapInOp(bsonutil.OpConcat, value, args[2], this),
			bsonutil.WrapInOp(bsonutil.OpEq, value, nil),
		),
	)

	return bsonutil.WrapInLet(assignment, body), nil
}

func (f *baseScalarFunctionExpr) reverseToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(reverse)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(reverse)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInCond(
		nil,
		bsonutil.WrapInLet(bsonutil.NewM(bsonutil.NewDocElem("input", args[0])), bsonutil.WrapInReduce(bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpRange, bsonutil.NewArray(
				0,
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, "$$input")),
			)),
		), "", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpConcat, bsonutil.NewArray(
			bsonutil.NewM(
				bsonutil.NewDocElem("$substrCP", bsonutil.NewArray(
					"$$input",
					"$$this",
					1,
				)),
			),
			"$$value",
		)),
		)),
		), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpLte, bsonutil.NewArray(
			args[0],
			nil,
		))),
	), nil
}

func (f *baseScalarFunctionExpr) rightToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {

	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(right)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(right)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	conds := bsonutil.NewArray()
	var strLength, subStrLength interface{}

	if stringValue, ok := bsonutil.GetLiteral(args[0]); ok {
		// string is literal
		if stringValue == nil {
			return bsonutil.MgoNullLiteral, nil
		}

		if s, ok := stringValue.(string); ok {
			strLength = bsonutil.WrapInLiteral(len(s))
		} else {
			strLength = bsonutil.WrapInOp(bsonutil.OpStrlenCP, "$$string")
		}
	} else {
		// string is not a literal
		strLength = bsonutil.WrapInOp(bsonutil.OpStrlenCP, "$$string")
		conds = append(conds, "$$string")
	}

	if lengthValue, ok := bsonutil.GetLiteral(args[1]); ok {
		if lengthValue == nil {
			return bsonutil.MgoNullLiteral, nil
		}

		// when length is negative, just use 0. round length to closest integer
		if i, ok := lengthValue.(int64); ok {
			args[1] = bsonutil.WrapInLiteral(int64(math.Max(0, float64(i))))
			subStrLength = "$$length"
		} else {
			args[1] = bsonutil.WrapInRound(bsonutil.WrapInOp(bsonutil.OpMax, args[1], 0))
			subStrLength = "$$length"
		}
	} else {
		subStrLength = bsonutil.WrapInRound(bsonutil.WrapInOp(bsonutil.OpMax, "$$length", 0))
		conds = append(conds, "$$length")
	}

	// start = max(0, strLen - subStrLen)
	start := bsonutil.WrapInOp(bsonutil.OpMax, 0, bsonutil.WrapInOp(bsonutil.OpSubtract, strLength, subStrLength))

	subStrOp := bsonutil.WrapInOp(bsonutil.OpSubstr, "$$string", start, subStrLength)

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("string", args[0]),
		bsonutil.NewDocElem("length", args[1]),
	)

	letEvaluation := bsonutil.WrapInNullCheckedCond(nil, subStrOp, conds...)
	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil
}

func (f *baseScalarFunctionExpr) roundToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) < 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(round)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs[0:1])
	if err != nil {
		return nil, err
	}
	switch len(exprs) {
	case 1:
		return bsonutil.WrapInRound(args[0]), nil
	case 2:
		if arg1, ok := exprs[1].(SQLValue); ok {
			return bsonutil.WrapInRoundWithPrecision(args[0], Float64(arg1)), nil
		}
		fallthrough
	default:
		return nil, newPushdownFailure("SQLScalarFunctionExpr(round)", "unsupported form")
	}
}

func (f *baseScalarFunctionExpr) rpadToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	return f.padToAggregationLanguage(t, exprs, false)
}

func (f *baseScalarFunctionExpr) rtrimToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(rtrim)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(rtrim)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	if t.versionAtLeast(4, 0, 0) {
		return bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpRTrim, bsonutil.NewM(
				bsonutil.NewDocElem("input", args[0]),
				bsonutil.NewDocElem("chars", " "),
			)),
		), nil
	}

	rtrimCond := bsonutil.WrapInCond(
		"",
		bsonutil.WrapInLRTrim(false, args[0]), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
			args[0],
			"",
		))))

	return bsonutil.WrapInNullCheckedCond(
		nil,
		rtrimCond,
		args[0],
	), nil
}

func (f *baseScalarFunctionExpr) secondToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(second)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$second", args[0]), nil
}

func (f *baseScalarFunctionExpr) signToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(sign)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInCond(nil,
		bsonutil.WrapInCond(bsonutil.WrapInLiteral(0),
			bsonutil.WrapInCond(bsonutil.WrapInLiteral(1),
				bsonutil.WrapInLiteral(-1), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGt, bsonutil.NewArray(
					args[0],
					bsonutil.WrapInLiteral(0),
				))),
			), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
				args[0],
				bsonutil.WrapInLiteral(0),
			))),
		), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpLte, bsonutil.NewArray(
			args[0],
			nil,
		))),
	), nil
}

func (f *baseScalarFunctionExpr) sinToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(sin)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	input, absInput := "$$input", "$$absInput"
	inputLetAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("input", args[0]),
	)

	absInputLetAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("absInput", bsonutil.WrapInOp(bsonutil.OpAbs, input)),
	)

	rem, phase := "$$rem", "$$phase"
	remPhaseAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("rem", bsonutil.WrapInOp(bsonutil.OpMod, absInput, math.Pi/2)),
		bsonutil.NewDocElem("phase", bsonutil.WrapInOp(bsonutil.OpMod,
			bsonutil.WrapInOp(bsonutil.OpTrunc,
				bsonutil.WrapInOp(bsonutil.OpDivide, absInput, math.Pi/2),
			),
			4.0)),
	)

	// 3.2 does not support $switch, so just use chained $cond, assuming
	// zeroCase will be most common (since it's the first phase) Because we
	// use the Maclaurin Power Series for sin and cos, we need to adjust
	// our input into a domain that is good for our approximation, that
	// being the first quadrant (phase). For phases outside of the first,
	// we can adjust the functions as:
	//
	// phase | Maclaurin Power Series
	// ------------------------------
	// 0     | sin(rem)
	// 1     | cos(rem)
	// 2     | -1 * sin(rem)
	// 3     | -1 * cos(rem)
	// where the phase is defined as the trunc(input / (pi/2)) % 4
	// and the remainder is input % (pi/2).

	threeCase := bsonutil.WrapInCond(bsonutil.WrapInOp(bsonutil.OpMultiply,
		-1.0,
		bsonutil.WrapInCosPowerSeries(rem)),
		nil,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			3))
	twoCase := bsonutil.WrapInCond(bsonutil.WrapInOp(bsonutil.OpMultiply,
		-1.0,
		bsonutil.WrapInSinPowerSeries(rem)),
		threeCase,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			2))
	oneCase := bsonutil.WrapInCond(bsonutil.WrapInCosPowerSeries(rem),
		twoCase,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			1))
	zeroCase := bsonutil.WrapInCond(bsonutil.WrapInSinPowerSeries(rem),
		oneCase,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			0))

	// cos(-x) = cos(x), but sin(-x) = -sin(x), so if the original input is negative multiply by -1.
	return bsonutil.WrapInLet(inputLetAssignment,
		bsonutil.WrapInLet(absInputLetAssignment,
			bsonutil.WrapInLet(remPhaseAssignment,
				bsonutil.WrapInCond(zeroCase,
					bsonutil.WrapInOp(bsonutil.OpMultiply, -1.0, zeroCase),
					bsonutil.WrapInOp(bsonutil.OpGte, input, 0),
				),
			),
		),
	), nil
}

func (f *baseScalarFunctionExpr) spaceToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(space)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(space)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	n := "$$n"
	return bsonutil.WrapInLet(bsonutil.NewM(bsonutil.NewDocElem("n", bsonutil.WrapInRound(args[0]))),
		bsonutil.WrapInCond(nil,
			bsonutil.WrapInReduce(bsonutil.WrapInRange(0, n, 1),
				"",
				bsonutil.WrapInOp(bsonutil.OpConcat, "$$value", " "),
			),
			bsonutil.WrapInOp(bsonutil.OpLte, n, nil),
		),
	), nil
}

func (f *baseScalarFunctionExpr) sqrtToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(sqrt)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInCond(bsonutil.NewM(bsonutil.NewDocElem("$sqrt", args[0])), nil, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGte, bsonutil.NewArray(
		args[0],
		0,
	)))), nil
}

func (f *baseScalarFunctionExpr) substringIndexToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(substringIndex)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(substringIndex)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	delim, split := "$$delim", "$$split"
	inputAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("delim", args[1]),
	)

	splitAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("split", bsonutil.WrapInOp(bsonutil.OpSlice,
			bsonutil.WrapInOp(bsonutil.OpSplit, args[0], delim),
			bsonutil.WrapInRound(args[2]),
		)),
	)

	this, value := "$$this", "$$value"
	body := bsonutil.WrapInReduce(split,
		nil,
		bsonutil.WrapInCond(this,
			bsonutil.WrapInOp(bsonutil.OpConcat, value, delim, this),
			bsonutil.WrapInOp(bsonutil.OpEq, value, nil),
		),
	)

	return bsonutil.WrapInLet(inputAssignment,
		bsonutil.WrapInLet(splitAssignment,
			body,
		),
	), nil
}

func (f *baseScalarFunctionExpr) substringToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(substring)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if len(exprs) != 2 && len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(substring)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}
	strVal := args[0]
	indexVal := args[1]

	var lenVal interface{}
	if len(args) == 3 {
		lenVal = args[2]
	} else {
		lenVal = bsonutil.NewM(bsonutil.NewDocElem("$strLenCP", args[0]))
	}

	indexNegVal := bsonutil.WrapInLet(bsonutil.NewM(
		bsonutil.NewDocElem("indexValNeg", bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpAdd, bsonutil.NewArray(
					bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMultiply, bsonutil.NewArray(
						indexVal,
						-1,
					))),
					0.5,
				))))))), bsonutil.WrapInCond(bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, strVal)),
		"$$indexValNeg",
	))), "$$indexValNeg", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGte, bsonutil.NewArray(
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, strVal)),
		"$$indexValNeg",
	)))))

	indexPosVal := bsonutil.NewM(
		bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
			bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.NewM(
					bsonutil.NewDocElem(bsonutil.OpAdd, bsonutil.NewArray(
						indexVal,
						0.5,
					)))),
			),
			1,
		)))

	roundOffIndex := bsonutil.WrapInCond(
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpAdd, bsonutil.NewArray(
			indexVal,
			0.5))))),
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpAdd, bsonutil.NewArray(
			indexVal,
			-0.5))))),
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGte, bsonutil.NewArray(
			indexVal,
			0,
		))))

	indexValBSONM := bsonutil.WrapInLet(
		bsonutil.NewM(bsonutil.NewDocElem("roundOffIndex", roundOffIndex)),
		bsonutil.WrapInCond(
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, strVal)),
			bsonutil.WrapInCond(
				indexPosVal,
				indexNegVal, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGt,
					bsonutil.NewArray(
						"$$roundOffIndex",
						0,
					)))), bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
					"$$roundOffIndex",
					0,
				))),
		))

	lenValBSONM := bsonutil.WrapInCond(
		0,
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpAdd, bsonutil.NewArray(
			lenVal,
			0.5,
		))))), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpLte, bsonutil.NewArray(
			lenVal,
			0,
		))),
	)

	return bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(bsonutil.NewDocElem("$substrCP", bsonutil.NewArray(
			strVal,
			indexValBSONM,
			lenValBSONM,
		))), strVal, indexVal, lenVal,
	), nil
}

func (f *baseScalarFunctionExpr) tanToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	num, err := f.sinToAggregationLanguage(t, []SQLExpr{exprs[0]})
	if err != nil {
		return nil, err
	}

	denom, err := f.cosToAggregationLanguage(t, []SQLExpr{exprs[0]})
	if err != nil {
		return nil, err
	}

	// epsilon the smallest value we allow for denom, computed to roughly
	// tie-out with mysqld.
	epsilon := 6.123233995736766e-17
	// Replace abs(denom) < epsilon with epsilon since mysql doesn't seem
	// to return NULL or INF for inf values of tan as one would expect, and
	// we do not want to trigger a $divide by 0.
	return bsonutil.WrapInOp(bsonutil.OpDivide,
		num,
		bsonutil.WrapInCond(epsilon,
			denom,
			bsonutil.WrapInOp(bsonutil.OpLte,
				bsonutil.WrapInOp(bsonutil.OpAbs, denom), epsilon,
			),
		),
	), nil
}

func (f *baseScalarFunctionExpr) timestampAddToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 5, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(timestampAdd)",
			"cannot push down to MongoDB < 3.5",
		)
	}

	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(timestampAdd)",
			incorrectArgCountMsg,
		)
	}

	unit := exprs[0].String()
	args, err := t.translateArgs(exprs[1:])
	if err != nil {
		return nil, err
	}
	interval := args[0]

	timestampExpr := args[1]
	// This is a very large and costly expression, make sure to bind it in
	// a let (in the switch at the end of the function).
	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("timestampArg", timestampExpr),
	)

	// Use timestampArg to refer to $$timestampArg below, referencing the var defined above.
	timestampArg := "$$timestampArg"

	// handleSimpleCase generates code for cases where we do not need to
	// use $dateFromParts, we just round the interval if the round argument
	// is true, and multiply by the number of milliseconds corresponded to
	// by 'u' then add to the timestamp.
	handleSimpleCase := func(u string, round bool) interface{} {
		if round {
			return bsonutil.WrapInOp(bsonutil.OpAdd,
				timestampArg,
				bsonutil.WrapInOp(bsonutil.OpMultiply,
					bsonutil.WrapInRound(interval),
					toMilliseconds[u]))
		}
		return bsonutil.WrapInOp(bsonutil.OpAdd,
			timestampArg,
			bsonutil.WrapInOp(bsonutil.OpMultiply,
				interval,
				toMilliseconds[u]))
	}

	// handleDateFromPartsCase handles cases where we need to use
	// $dateFromParts because we want to add a Year, a Month, or 3 Months
	// (a Quarter) to the specific date part.
	handleDateFromPartsCase := func(u string) interface{} {
		// Start with the equations for Quarter/Month, since they are
		// the same. They use a shared computation part
		// (sharedComputation) that changes based on if this is a
		// Quarter or Month.
		sharedComputation := "$$sharedComputation"
		newYear, newMonth := "$$newYear", "$$newMonth"
		dayExpr := bsonutil.WrapInOp(bsonutil.OpDayOfMonth, timestampArg)
		// This template is used in a call to $dateFromParts.
		// The Year case modifies part of the template.
		template := bsonutil.NewM(
			bsonutil.NewDocElem("year", "$$newYear"),
			bsonutil.NewDocElem("month", "$$newMonth"),
			bsonutil.NewDocElem("day",
				// The following MongoDB aggregation language implements this go code,
				// the goal of which is to keep days from overflowing when adding
				// Quarters or Months.
				// switch m {
				// case 2:
				// 	if isLeapYear(y) {
				// 		d = util.MinInt(d, 29)
				//	} else {
				//		d = util.MinInt(d, 28)
				//	}
				// case 4, 6, 9, 11:
				//	d = util.MinInt(d, 30)
				// }
				// otherwise d is left unchanged as the day of the input timestamp.
				bsonutil.WrapInSwitch(bsonutil.WrapInOp(bsonutil.OpDayOfMonth, timestampArg),
					bsonutil.WrapInEqCase(newMonth, 2,
						bsonutil.WrapInCond(bsonutil.WrapInOp(bsonutil.OpMin, dayExpr, 29),
							bsonutil.WrapInOp(bsonutil.OpMin, dayExpr, 28),
							bsonutil.WrapInIsLeapYear(newYear)),
					),
					bsonutil.WrapInEqCase(newMonth, 4,
						bsonutil.WrapInOp(bsonutil.OpMin, dayExpr, 30)),
					bsonutil.WrapInEqCase(newMonth, 6,
						bsonutil.WrapInOp(bsonutil.OpMin, dayExpr, 30)),
					bsonutil.WrapInEqCase(newMonth, 9,
						bsonutil.WrapInOp(bsonutil.OpMin, dayExpr, 30)),
					bsonutil.WrapInEqCase(newMonth, 11,
						bsonutil.WrapInOp(bsonutil.OpMin, dayExpr, 30)),
				)),
			bsonutil.NewDocElem("hour", bsonutil.WrapInOp(bsonutil.OpHour, timestampArg)),
			bsonutil.NewDocElem("minute", bsonutil.WrapInOp(bsonutil.OpMinute, timestampArg)),
			bsonutil.NewDocElem("second", bsonutil.WrapInOp(bsonutil.OpSecond, timestampArg)),
			bsonutil.NewDocElem("millisecond", bsonutil.WrapInOp(bsonutil.OpMillisecond, timestampArg)),
		)

		var sharedComputationLetAssignment interface{}
		var newYearMonthLetAssignment interface{}
		switch u {
		case Year:
			// For Year intervals, the year, month, and day use
			// different, simpler equations. Keep everything but
			// year, to year we add the rounded interval. There is
			// no SharedComputation part, so we do not bsonutil.WrapInLet.
			// Note that the rest of the template is maintained.
			template["year"] = bsonutil.WrapInOp(bsonutil.OpAdd,
				bsonutil.WrapInRound(interval),
				bsonutil.WrapInOp(bsonutil.OpYear,
					timestampArg))
			template["month"] = bsonutil.WrapInOp(bsonutil.OpMonth,
				timestampArg)
			template["day"] = bsonutil.WrapInOp(bsonutil.OpDayOfMonth,
				timestampArg)
			return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDateFromParts, template))
		// For Quarter and Month intervals, only the SharedComputation
		// part changes.
		case Quarter:
			// SharedComputation = Month + round(interval) * 3 - 1.
			sharedComputationLetAssignment = bsonutil.NewM(
				bsonutil.NewDocElem("sharedComputation", bsonutil.WrapInOp(bsonutil.OpSubtract,
					bsonutil.WrapInOp(bsonutil.OpAdd,
						bsonutil.WrapInOp(bsonutil.OpMonth, timestampArg),
						bsonutil.WrapInOp(bsonutil.OpMultiply,
							bsonutil.WrapInRound(interval),
							3),
					),
					1)),
			)

		case Month:
			// SharedComputation = Month + round(interval) - 1.
			sharedComputationLetAssignment = bsonutil.NewM(
				bsonutil.NewDocElem("sharedComputation", bsonutil.WrapInOp(bsonutil.OpSubtract,
					bsonutil.WrapInOp(bsonutil.OpAdd,
						bsonutil.WrapInOp(bsonutil.OpMonth, timestampArg),
						bsonutil.WrapInRound(interval),
					),
					1)),
			)

		}

		newYearMonthLetAssignment = bsonutil.NewM(
			// Year = Year + SharedComputation / 12, where / truncates.

			bsonutil.NewDocElem("newYear", bsonutil.WrapInOp(bsonutil.OpAdd,
				bsonutil.WrapInOp(bsonutil.OpYear, timestampArg),
				bsonutil.WrapInIntDiv(sharedComputation, 12),
			)),

			// Month = SharedComputation % 12 + 1.
			bsonutil.NewDocElem("newMonth", bsonutil.WrapInOp(bsonutil.OpAdd,
				bsonutil.WrapInOp(bsonutil.OpMod,
					sharedComputation,
					12),
				1)),
		)

		// Add lets for Quarter and Month.
		return bsonutil.WrapInLet(sharedComputationLetAssignment,
			bsonutil.WrapInLet(newYearMonthLetAssignment, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDateFromParts, template))),
		)
	}

	// bsonutil.WrapInLet to bind $$timestampArg.
	switch unit {
	case Year, Month, Quarter:
		return bsonutil.WrapInLet(letAssignment, handleDateFromPartsCase(unit)), nil
	// It is wrong to round for Second, and rounding for Microsecond is
	// just pointless since MongoDB supports only milliseconds, and will
	// automatically round to the nearest millisecond for us.
	case Second, Microsecond:
		return bsonutil.WrapInLet(letAssignment, handleSimpleCase(unit, false)), nil
	default:
		return bsonutil.WrapInLet(letAssignment, handleSimpleCase(unit, true)), nil
	}
}

func (f *baseScalarFunctionExpr) timestampDiffToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 5, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(timestampDiff)",
			"cannot push down to MongoDB < 3.5",
		)
	}

	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(timestampDiff)",
			incorrectArgCountMsg,
		)
	}

	unit := exprs[0].String()

	args, err := t.translateArgs(exprs[1:])
	if err != nil {
		return nil, err
	}

	timestampExpr1, timestampExpr2 := args[0], args[1]

	// This is a very large and costly expression, make sure to bind it in
	// a let (in the switch at the end of the function).
	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("timestampArg1", timestampExpr1),
		bsonutil.NewDocElem("timestampArg2", timestampExpr2),
	)

	// Use timestampArg{1,2} to refer to $$timestampArg{1,2} below,
	// referencing the var defined above.
	timestampArg1, timestampArg2 := "$$timestampArg1", "$$timestampArg2"

	// handleSimpleCase generates code for cases where we do not need to
	// use and date part access functions (like $dayOfMonth), we just
	// subtract: timestampArg2 - timestampArg1 then divide by the number of
	// milliseconds corresponded to by 'u'.
	handleSimpleCase := func(u string) interface{} {
		return bsonutil.WrapInIntDiv(bsonutil.WrapInOp(bsonutil.OpSubtract,
			timestampArg2,
			timestampArg1),
			toMilliseconds[u])
	}

	// handleDatePartsCase handles cases where we need to use
	// date part access functions (like $dayOfMonth).
	handleDatePartsCase := func(u string) interface{} {
		year1, month1 := "$$year1", "$$month1"
		year2, month2 := "$$year2", "$$month2"
		datePartsLetAssignment := bsonutil.NewM(
			bsonutil.NewDocElem("year1", bsonutil.WrapInOp(bsonutil.OpYear, timestampArg1)),
			bsonutil.NewDocElem("month1", bsonutil.WrapInOp(bsonutil.OpMonth, timestampArg1)),
			bsonutil.NewDocElem("day1", bsonutil.WrapInOp(bsonutil.OpDayOfMonth, timestampArg1)),
			bsonutil.NewDocElem("hour1", bsonutil.WrapInOp(bsonutil.OpHour, timestampArg1)),
			bsonutil.NewDocElem("minute1", bsonutil.WrapInOp(bsonutil.OpMinute, timestampArg1)),
			bsonutil.NewDocElem("second1", bsonutil.WrapInOp(bsonutil.OpSecond, timestampArg1)),
			bsonutil.NewDocElem("millisecond1", bsonutil.WrapInOp(bsonutil.OpMillisecond, timestampArg1)),
			bsonutil.NewDocElem("year2", bsonutil.WrapInOp(bsonutil.OpYear, timestampArg2)),
			bsonutil.NewDocElem("month2", bsonutil.WrapInOp(bsonutil.OpMonth, timestampArg2)),
			bsonutil.NewDocElem("day2", bsonutil.WrapInOp(bsonutil.OpDayOfMonth, timestampArg2)),
			bsonutil.NewDocElem("hour2", bsonutil.WrapInOp(bsonutil.OpHour, timestampArg2)),
			bsonutil.NewDocElem("minute2", bsonutil.WrapInOp(bsonutil.OpMinute, timestampArg2)),
			bsonutil.NewDocElem("second2", bsonutil.WrapInOp(bsonutil.OpSecond, timestampArg2)),
			bsonutil.NewDocElem("millisecond2", bsonutil.WrapInOp(bsonutil.OpMillisecond, timestampArg2)),
		)

		var outputLetAssignment interface{}
		var generateEpsilon func(arg1, arg2 string) interface{}
		output := "$$output"
		if u == Year {
			// For years, the output will be year2 - year1, but we
			// need to adjust that by the yearEpsilon, which is 0
			// or 1, depending on the remainder of the date object.
			// For instance if we have 2016-01-29 - 2015-01-30, the
			// answer is actually 0, because 30 > 29, giving us a
			// yearEpsilon of 1, and 1 - 1 = 0. If output is
			// positive we subtract the epsilon, if output is
			// negative, we add the epsilon, meaning we always go
			// toward 0.
			generateEpsilon = func(arg1, arg2 string) interface{} {
				return bsonutil.WrapInCond(bsonutil.WrapInLiteral(1),
					bsonutil.WrapInLiteral(0),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$month"+arg1, "$$month"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$day"+arg1, "$$day"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$hour"+arg1, "$$hour"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$minute"+arg1, "$$minute"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$second"+arg1, "$$second"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$millisecond"+arg1, "$$millisecond"+arg2),
				)
			}
			// output = year2 - year1.
			outputLetAssignment = bsonutil.NewM(
				bsonutil.NewDocElem("output", bsonutil.WrapInOp(bsonutil.OpSubtract, year2, year1)),
			)

		} else {
			// For months/quarters, the output will be (year2 -
			// year1) * 12 + month2 - month1, but we need to adjust
			// that by the monthEpsilon, which is 0 or 1, depending
			// on the remainder of the date object. For instance
			// if we have 2016-01-29 - 2015-01-30, the answer is
			// actually 11, because 30 > 29, giving us a
			// monthEpsilon of 1, and 12 - 1 = 11. If the output
			// is positive we subtract the epsilon, if output is
			// negative, we add the epsilon, meaning we always go
			// toward 0.
			generateEpsilon = func(arg1, arg2 string) interface{} {
				return bsonutil.WrapInCond(bsonutil.WrapInLiteral(1),
					bsonutil.WrapInLiteral(0),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$day"+arg1, "$$day"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$hour"+arg1, "$$hour"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$minute"+arg1, "$$minute"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$second"+arg1, "$$second"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$millisecond"+arg1, "$$millisecond"+arg2),
				)

			}
			// output = (year2 - year1) * 12 + month2 - month1.
			outputLetAssignment = bsonutil.NewM(
				bsonutil.NewDocElem("output", bsonutil.WrapInOp(bsonutil.OpAdd,
					bsonutil.WrapInOp(bsonutil.OpMultiply,
						bsonutil.WrapInOp(bsonutil.OpSubtract, year2, year1),
						12),
					bsonutil.WrapInOp(bsonutil.OpSubtract, month2, month1),
				)),
			)

		}

		// Generate epsilons and whether we add or subtract said epsilon, which
		// is decided on whether or not "output" is negative or positive.
		ltBranch := bsonutil.WrapInOp(bsonutil.OpAdd, output, generateEpsilon("2", "1"))
		gtBranch := bsonutil.WrapInOp(bsonutil.OpSubtract, output, generateEpsilon("1", "2"))
		applyEpsilonExpr := bsonutil.WrapInLet(outputLetAssignment,
			bsonutil.WrapInSwitch(bsonutil.WrapInLiteral(0),
				bsonutil.WrapInCase(bsonutil.WrapInOp(bsonutil.OpLt, output, bsonutil.WrapInLiteral(0)), ltBranch),
				bsonutil.WrapInCase(bsonutil.WrapInOp(bsonutil.OpGt, output, bsonutil.WrapInLiteral(0)), gtBranch),
			),
		)

		retExpr := bsonutil.WrapInLet(datePartsLetAssignment,
			bsonutil.WrapInLet(outputLetAssignment,
				applyEpsilonExpr,
			),
		)
		// Quarter is just the number of months integer divided by 3.
		if u == Quarter {
			return bsonutil.WrapInIntDiv(retExpr, 3)
		}
		return retExpr
	}

	// bsonutil.WrapInLet to bind $$timestampArg.
	switch unit {
	case Year, Month, Quarter:
		return bsonutil.WrapInLet(letAssignment, handleDatePartsCase(unit)), nil
	default:
		return bsonutil.WrapInLet(letAssignment, handleSimpleCase(unit)), nil
	}
}

func (f *baseScalarFunctionExpr) timestampToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 5, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(timestamp)",
			"cannot push down to MongoDB < 3.5",
		)
	}

	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(timestamp)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	val := "$$val"
	inputLet := bsonutil.NewM(
		bsonutil.NewDocElem("val", args[0]),
	)

	wrapInDateFromString := func(v interface{}) bson.M {
		return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDateFromString, bsonutil.NewM(bsonutil.NewDocElem("dateString", v))))
	}

	// CASE 1: it's already a Mongo date, we just return it
	isDateType := containsBSONType(val, "date")
	dateBranch := bsonutil.WrapInCase(isDateType, val)

	// CASE 2: it's a number.
	isNumber := containsBSONType(val, "int", "decimal", "long", "double")

	// evaluates to true if val positive and has <= X digits.
	hasUpToXDigits := func(x float64) interface{} {
		return bsonutil.WrapInInRange(val, 0, math.Pow(10, x))
	}

	// This handles converting a number in YYMMDDHHMMSS format to YYYYMMDDHHMMSS.
	// if YY < 70, we assume they meant 20YY. if YY > 70, we assume 19YY.
	getPadding := func(v interface{}) interface{} {
		return bsonutil.WrapInCond(
			20000000000000,
			19000000000000,
			bsonutil.WrapInOp(bsonutil.OpLt,
				bsonutil.WrapInOp(bsonutil.OpDivide,
					v, 10000000000),
				70))
	}

	// Constant for the HHMMSS factor to handle dates that do not have HHMMSS.
	hhmmssFactor := 1000000

	// We interpret this as being format YYMMDD, multiply by hhmmssFactor for HHMMSS then pad.
	ifSix := bsonutil.WrapInOp(bsonutil.OpAdd,
		bsonutil.WrapInOp(bsonutil.OpMultiply,
			val,
			hhmmssFactor),
		getPadding(bsonutil.WrapInOp(bsonutil.OpMultiply,
			val,
			hhmmssFactor)))
	sixBranch := bsonutil.WrapInCase(hasUpToXDigits(6), ifSix)

	// This number is YYYYMMDD, again, multiply by hhmmssFactor.
	eightBranch := bsonutil.WrapInCase(hasUpToXDigits(8), bsonutil.WrapInOp(bsonutil.OpMultiply, val, hhmmssFactor))

	// If it's twelve digits, interpret as YYMMDDHHMMSS. Make sure to pad the number.
	ifTwelve := bsonutil.WrapInOp(bsonutil.OpAdd, val, getPadding(val))
	twelveBranch := bsonutil.WrapInCase(hasUpToXDigits(12), ifTwelve)

	// if fourteen, YYYYMMDDHHMMSS, we can use as it as is.
	fourteenBranch := bsonutil.WrapInCase(hasUpToXDigits(14), val)

	// define "num", the input number normalized to 14 digits, in a "let"
	numberVar := bsonutil.WrapInSwitch(nil, sixBranch, eightBranch, twelveBranch, fourteenBranch)
	numberLetVars := bsonutil.NewM(bsonutil.NewDocElem("num", numberVar))

	dateParts := bsonutil.NewM(
		// YYYYMMDDHHMMSS / 10000000000 = YYYY

		bsonutil.NewDocElem("year", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide,
			"$$num",
			10000000000)),
		)),
		// (YYYYMMDDHHMMSS / 100000000) % 100 = MM

		bsonutil.NewDocElem("month", bsonutil.WrapInOp(bsonutil.OpMod, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide,
			"$$num",
			100000000)),
		), 100)),

		// YYYYMMDDHHMMSS / 1000000) % 100 = DD
		bsonutil.NewDocElem("day", bsonutil.WrapInOp(bsonutil.OpMod, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide,
			"$$num",
			1000000)),
		), 100)),

		// YYYYMMDDHHMMSS / 10000) % 100 = HH
		bsonutil.NewDocElem("hour", bsonutil.WrapInOp(bsonutil.OpMod, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide,
			"$$num",
			10000)),
		), 100)),

		// YYYYMMDDHHMMSS / 100) % 100 = MM
		bsonutil.NewDocElem("minute", bsonutil.WrapInOp(bsonutil.OpMod, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide,
			"$$num",
			100)),
		), 100)),

		// YYYYMMDDHHMMSS % 100 = SS
		bsonutil.NewDocElem("second", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpMod,
			"$$num",
			100)),
		)),
		// YYYYMMDDHHMMSS.FFFFF % 1 * 1000 = ms

		bsonutil.NewDocElem("millisecond", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpMultiply,
			bsonutil.WrapInOp(bsonutil.OpMod,
				"$$num",
				1),
			1000)),
		)),
	)

	// try to avoid aggregation errors by catching obviously invalid dates
	yearValid := bsonutil.WrapInInRange("$$year", 0, 10000)
	monthValid := bsonutil.WrapInInRange("$$month", 1, 13)
	dayValid := bsonutil.WrapInInRange("$$day", 1, 32)
	// Mongo DB actually supports HH=24 which converts to 0, but MySQL does not (it returns NULL)
	// so we stick to MySQL semantics and cap valid hours at 23.
	// Interestingly, $dateFromString does NOT support HH=24.
	hourValid := bsonutil.WrapInInRange("$$hour", 0, 24)
	minuteValid := bsonutil.WrapInInRange("$$minute", 0, 60)
	secondValid := bsonutil.WrapInInRange("$$second", 0, 60)

	makeDateOrNull := bsonutil.WrapInCond(bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDateFromParts, bsonutil.NewM(
		bsonutil.NewDocElem("year", "$$year"),
		bsonutil.NewDocElem("month", "$$month"),
		bsonutil.NewDocElem("day", "$$day"),
		bsonutil.NewDocElem("hour", "$$hour"),
		bsonutil.NewDocElem("minute", "$$minute"),
		bsonutil.NewDocElem("second", "$$second"),
		bsonutil.NewDocElem("millisecond", "$$millisecond"),
	)),
	), nil, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpAnd, bsonutil.NewArray(
		yearValid,
		monthValid,
		dayValid,
		hourValid,
		minuteValid,
		secondValid,
	)),
	))

	evaluateNumber := bsonutil.WrapInLet(dateParts, makeDateOrNull)
	handleNumberToDate := bsonutil.WrapInLet(numberLetVars, evaluateNumber)
	numberBranch := bsonutil.WrapInCase(isNumber, handleNumberToDate)

	// CASE 3: it's a string
	isString := containsBSONType(val, "string")

	// First split on T, take first substring, then split that on " ", and
	// take first substring. this gives us just the date part of the
	// string. note that if the string doesn't have T or a space, just
	// returns original string
	trimmedDateString := bsonutil.WrapInOp(bsonutil.OpArrElemAt,
		bsonutil.WrapInOp(bsonutil.OpSplit,
			bsonutil.WrapInOp(bsonutil.OpArrElemAt,
				bsonutil.WrapInOp(bsonutil.OpSplit, val, "T"),
				0),
			" "),
		0)

	// Repeat the step above but take the second element to get the time
	// part. Replace with "" if we can not find a second element.
	trimmedTimeString := bsonutil.WrapInIfNull(
		bsonutil.WrapInOp(bsonutil.OpArrElemAt,
			bsonutil.WrapInOp(bsonutil.OpSplit, val, "T"),
			1),
		bsonutil.WrapInIfNull(
			bsonutil.WrapInOp(bsonutil.OpArrElemAt,
				bsonutil.WrapInOp(bsonutil.OpSplit, val, " "),
				1),
			""),
	)

	// Convert the date and time strings to arrays so we can use
	// map/reduce.
	trimmedDateAsArray := bsonutil.WrapInStringToArray("$$trimmedDate")
	trimmedTimeAsArray := bsonutil.WrapInStringToArray("$$trimmedTime")

	// isSeparator evaluates to true if a character is in the defined
	// separator list
	isSeparator := bsonutil.WrapInOp(bsonutil.OpNeq,
		-1,
		bsonutil.WrapInOp("$indexOfArray",
			bsonutil.DateComponentSeparator,
			"$$c"))

	// Use map to convert all separators in the date string to - symbol,
	// and leave numbers as-is
	dateNormalized := bsonutil.WrapInMap(trimmedDateAsArray,
		"c",
		bsonutil.WrapInCond("-",
			"$$c",
			isSeparator))
	// Use map to convert all separators in the time string to '.' symbol,
	// and leave numbers as-is. We use '.' instead of ':' so that MongoDB
	// correctly handles fractional seconds. 10.11.23.1234 is parsed
	// correctly as 10:11:23.1234, saving us some effort (and runtime).
	timeNormalized := bsonutil.WrapInMap(trimmedTimeAsArray,
		"c",
		bsonutil.WrapInCond(".",
			"$$c",
			isSeparator))

	// Use reduce to convert characters back to a single string for date and time.
	dateJoined := bsonutil.WrapInReduce(dateNormalized,
		"",
		bsonutil.WrapInOp(bsonutil.OpConcat,
			"$$value",
			"$$this"))
	timeJoined := bsonutil.WrapInReduce(timeNormalized,
		"",
		bsonutil.WrapInOp(bsonutil.OpConcat,
			"$$value",
			"$$this"))

	// if the third character is a -, or if the string is only 6 digits
	// long and has no slashes, then the string is either format YY/MM/DD
	// or YYMMDD and we need to add the appropriate first two year digits
	// (19xx or 20xx) for Mongo to understand it
	hasShortYear := bsonutil.WrapInOp(bsonutil.OpOr,
		// length is only 6, assume YYMMDD
		bsonutil.WrapInOp(bsonutil.OpEq, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, "$$dateJoined")), 6),
		// third character is -, assume YY-MM-DD
		bsonutil.WrapInOp(bsonutil.OpEq,
			"-", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubstr, bsonutil.NewArray(
				"$$dateJoined",
				2,
				1,
			)),
			)))

	// "$dateFromString" actually pads correctly, but not if "/" is
	// used as the separator (it will assume year is last). If this
	// pushdown is shown to be slow by benchmarks, we should reconsider
	// allowing "$dateFromString" to handle padding. The change
	// would not be trivial due to how MongoDB cannot handle short dates
	// when there are no separators in the date.
	padYear := bsonutil.WrapInOp(bsonutil.OpConcat,
		bsonutil.WrapInCond(
			"20",
			"19",
			// check if first two digits < 70 to determine padding
			bsonutil.WrapInOp(
				bsonutil.OpLt, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubstr, bsonutil.NewArray(
					"$$dateJoined",
					0,
					2,
				))), "70")),
		"$$dateJoined")

	// we have to use nested $lets because in the outer one we define
	// $$trimmedDate and in the inner one we define $$dateJoined. defining
	// $$dateJoined requires knowing the length of trimmedDate, so we can't
	// do it all in one step.
	innerIn := bsonutil.WrapInCond(padYear, "$$dateJoined", hasShortYear)
	innerLet := bsonutil.WrapInLet(bsonutil.NewM(bsonutil.NewDocElem("dateJoined", dateJoined)), innerIn)

	// Concat the time back into the date.
	concatedDate := bsonutil.WrapInOp(bsonutil.OpConcat,
		innerLet,
		timeJoined)

	// gracefully handle strings that are too short to possibly be valid by returning null
	tooShort := bsonutil.WrapInOp(bsonutil.OpLt, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, "$$trimmedDate")), 6)
	outerIn := bsonutil.WrapInCond(nil, wrapInDateFromString(concatedDate), tooShort)
	outerLet := bsonutil.WrapInLet(bsonutil.NewM(bsonutil.NewDocElem("trimmedDate", trimmedDateString),
		bsonutil.NewDocElem("trimmedTime", trimmedTimeString),
	), outerIn)

	// Make sure if we get the int 0 we return NULL instead
	// of crashing. MySQL uses '0000-00-00' as an error output for some
	// functions and we encode it as the integer 0 within push down.
	stringBranch := bsonutil.WrapInCase(isString,
		bsonutil.WrapInCond(nil,
			outerLet,
			bsonutil.WrapInOp(bsonutil.OpEq,
				0,
				args[0])))

	return bsonutil.WrapInLet(inputLet, bsonutil.WrapInSwitch(nil, dateBranch, numberBranch, stringBranch)), nil

}

func (f *baseScalarFunctionExpr) toSecondsToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(toSeconds)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}
	// Subtract dayOne (0000-01-01) from the argument in mongo, then
	// convertms to seconds. When using $subtract on two dates in
	// MongoDB, the number of ms between the two dates is returned, and
	// the purpose of the TO_SECONDS function is to get the number of
	// seconds since 0000-01-01:
	dayOne := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	return bsonutil.WrapInOp(bsonutil.OpMultiply,
		bsonutil.WrapInOp(bsonutil.OpSubtract, args[0], dayOne),
		1e-3,
	), nil
}
func (f *baseScalarFunctionExpr) trimToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(trim)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(trim)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	if t.versionAtLeast(4, 0, 0) {
		return bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpTrim, bsonutil.NewM(
				bsonutil.NewDocElem("input", args[0]),
				bsonutil.NewDocElem("chars", " "),
			)),
		), nil
	}

	rtrimCond := bsonutil.WrapInCond(
		"",
		bsonutil.WrapInLRTrim(false, args[0]), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
			args[0],
			"",
		))))

	ltrimCond := bsonutil.WrapInCond(
		"",
		bsonutil.WrapInLRTrim(true, "$$rtrim"), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
			"$$rtrim",
			"",
		))))

	trimCond := bsonutil.WrapInLet(bsonutil.NewM(bsonutil.NewDocElem("rtrim", rtrimCond)), ltrimCond)

	trim := bsonutil.WrapInCond(
		"",
		trimCond, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
			args[0],
			"",
		))))

	return bsonutil.WrapInNullCheckedCond(
		nil,
		trim,
		args[0],
	), nil
}

func (f *baseScalarFunctionExpr) truncateToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(truncate)",
			incorrectArgCountMsg,
		)
	}
	dValue, ok := exprs[1].(SQLValue)
	if !ok {
		return nil, newPushdownFailure("SQLScalarFunctionExpr(truncate)", "second arg is not a literal")
	}

	d := Float64(dValue)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	if d >= 0 {
		pow := math.Pow(10, d)
		return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDivide, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCond, bsonutil.NewArray(
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGte, bsonutil.NewArray(
					args[0],
					0,
				))),
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpFloor, bsonutil.NewM(
					bsonutil.NewDocElem(bsonutil.OpMultiply, bsonutil.NewArray(
						args[0],
						pow,
					))))),
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCeil, bsonutil.NewM(
					bsonutil.NewDocElem(bsonutil.OpMultiply, bsonutil.NewArray(
						args[0],
						pow,
					))))),
			))),
			pow,
		))), nil
	}

	pow := math.Pow(10, math.Abs(d))
	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMultiply, bsonutil.NewArray(
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCond, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGte, bsonutil.NewArray(
				args[0],
				0,
			))),
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpFloor, bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpDivide, bsonutil.NewArray(
					args[0],
					pow,
				))))),
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCeil, bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpDivide, bsonutil.NewArray(
					args[0],
					pow,
				))))),
		))),
		pow,
	))), nil
}

func (f *baseScalarFunctionExpr) ucaseToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(ucase)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$toUpper", args[0]), nil
}

func (f *baseScalarFunctionExpr) unixTimestampToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	now := time.Now()

	if len(exprs) != 1 {
		return bsonutil.WrapInLiteral(now.Unix()), nil
	}

	arg, err := f.timestampToAggregationLanguage(t, exprs)
	if err != nil {
		return nil, err
	}

	// Subtract epoch (1970-01-01) from the argument in MongoDB, then
	// convert ms to seconds. When using $subtract on two dates in
	// MongoDB, the number of milliseconds between the two
	// timestamps is returned.
	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	_, tzCompensation := now.Zone()

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("diff", bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpDivide, bsonutil.NewArray(
					bsonutil.NewM(
						bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
							bsonutil.NewM(
								bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
									arg,
									epoch,
								)),
							),
							tzCompensation*1000,
						)),
					),
					1000,
				)),
			)),
		)),
	)

	letEvaluation := bsonutil.WrapInCond("$$diff", 0.0, bsonutil.WrapInOp(bsonutil.OpGt, "$$diff", 0))
	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil
}

// nolint: unparam
func (f *baseScalarFunctionExpr) utcDateToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	now := time.Now().In(time.UTC)
	cUTCd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return bsonutil.WrapInLiteral(cUTCd), nil
}

// nolint: unparam
func (f *baseScalarFunctionExpr) utcTimestampToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	return bsonutil.WrapInLiteral(time.Now().In(time.UTC)), nil
}

func (f *baseScalarFunctionExpr) weekToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) < 1 || len(exprs) > 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(week)",
			incorrectArgCountMsg,
		)
	}
	mode := int64(0)
	if len(exprs) == 2 {
		modeValue, ok := exprs[1].(SQLValue)
		if !ok {
			return nil, newPushdownFailure("SQLScalarFunctionExpr(week)", "mode is not a literal")
		}
		mode = Int64(modeValue)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}
	return bsonutil.WrapInWeekCalculation(args[0], mode), nil
}

func (f *baseScalarFunctionExpr) weekdayToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(weekday)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("date", args[0]),
	)

	conds := bsonutil.NewArray()
	if _, ok := bsonutil.GetLiteral(args[0]); !ok {
		conds = append(conds, "$$date")
	}

	letEvaluation := bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMod, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpAdd, bsonutil.NewArray(
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMod, bsonutil.NewArray(
					bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
						bsonutil.NewM(bsonutil.NewDocElem("$dayOfWeek", "$$date")),
						2,
					)),
					),
					7,
				)),
				),
				7,
			)),
			),
			7,
		)),
		), conds...,
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

func (f *baseScalarFunctionExpr) yearToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(year)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$year", args[0]), nil
}

func (f *baseScalarFunctionExpr) yearWeekToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) < 1 || len(exprs) > 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(yearWeek)",
			incorrectArgCountMsg,
		)
	}
	mode := int64(0)
	if len(exprs) == 2 {
		modeValue, ok := exprs[1].(SQLValue)
		if !ok {
			return nil, newPushdownFailure("SQLScalarFunctionExpr(yearWeek)", "mode is not a literal")
		}
		mode = Int64(modeValue)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	date, month, year, week := "$$date", "$$month", "$$year", "$$week"
	inputAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("date", args[0]),
	)

	monthAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("month", bsonutil.WrapInOp(bsonutil.OpMonth, date)),
		bsonutil.NewDocElem("year", bsonutil.WrapInOp(bsonutil.OpYear, date)),
	)

	var weekCalc interface{}

	// Unlike WEEK, YEARWEEK always uses the 1-53 modes. Thus
	// we always call week with the 1-53 of a 0-53, 1-53 pair.
	switch mode {

	// First day of week: Sunday, with a Sunday in this year.
	case 0, 2:
		weekCalc = bsonutil.WrapInWeekCalculation(date, 2)
	// First day of weekCalc: Monday, with 4 days in this year.
	case 1, 3:
		weekCalc = bsonutil.WrapInWeekCalculation(date, 3)
	// First day of weekCalc: Sunday, with 4 days in this year.
	case 4, 6:
		weekCalc = bsonutil.WrapInWeekCalculation(date, 6)
	// First day of weekCalc: Monday, with a Monday in this year.
	case 5, 7:
		weekCalc = bsonutil.WrapInWeekCalculation(date, 7)
	}

	weekAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("week", weekCalc),
	)

	newYear := "$$newYear"
	newYearAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("newYear", bsonutil.WrapInSwitch(year,
			bsonutil.WrapInEqCase(week, 1, bsonutil.WrapInCond(
				bsonutil.WrapInOp(bsonutil.OpAdd, year, 1), year,
				bsonutil.WrapInOp(bsonutil.OpEq, month, 12),
			),
			),
			bsonutil.WrapInEqCase(week, 52, bsonutil.WrapInCond(
				bsonutil.WrapInOp(bsonutil.OpSubtract, year, 1), year,
				bsonutil.WrapInOp(bsonutil.OpEq, month, 1),
			),
			),
			bsonutil.WrapInEqCase(week, 53, bsonutil.WrapInCond(
				bsonutil.WrapInOp(bsonutil.OpSubtract, year, 1), year,
				bsonutil.WrapInOp(bsonutil.OpEq, month, 1),
			),
			),
		)),
	)

	return bsonutil.WrapInLet(inputAssignment,
		bsonutil.WrapInLet(monthAssignment,
			bsonutil.WrapInLet(weekAssignment,
				bsonutil.WrapInLet(newYearAssignment,
					bsonutil.WrapInOp(bsonutil.OpAdd,
						bsonutil.WrapInOp(bsonutil.OpMultiply, newYear, 100),
						week,
					),
				),
			),
		),
	), nil

}

func (t *PushdownTranslator) translateArgs(exprs []SQLExpr) ([]interface{}, PushdownFailure) {
	args := []interface{}{}
	for _, e := range exprs {
		r, err := t.ToAggregationLanguage(e)
		if err != nil {
			return nil, err
		}
		args = append(args, r)
	}
	return args, nil
}
