package evaluator

import (
	"fmt"
	"math"
	"time"

	"github.com/10gen/mongoast/ast"

	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/mathutil"
	"github.com/10gen/sqlproxy/schema"

	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// FuncToAggregation for TO_DAYS has one issue wrt how TO_DAYS is supposed to perform:
// because our date treatment is backed by using MongoDB's $dateFromString function,
// if a date that doesn't exist (e.g., 0000-00-00 or 0001-02-29) is entered, we return
// an error instead of the NULL expected from MySQL. Unfortunately, checking for valid
// dates is too cost prohibitive. If at some point $dateFromString supports an onError/default
// value, we should switch to using that.
func (f *baseScalarFunctionExpr) toDaysToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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
	dayOne := astutil.DateConstant(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC))
	return ast.NewFunction(bsonutil.OpTrunc,
		ast.NewBinary(bsonutil.OpDivide,
			ast.NewBinary(bsonutil.OpSubtract, args[0], dayOne),
			astutil.FloatValue(millisecondsPerDay),
		),
	), nil
}

func (f *baseScalarFunctionExpr) absToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return ast.NewFunction(bsonutil.OpAbs, args[0]), nil
}

func (f *baseScalarFunctionExpr) acosToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	assignments, args := minimizeLetAssignments([]string{"input"}, args)

	// MySQL returns NULL for values outside of the range [-1,1].
	// asin x + acos x = pi/2
	return wrapInLet(assignments,
		astutil.WrapInCond(
			astutil.NullLiteral,
			astutil.WrapInAcosComputation(args[0]),
			ast.NewBinary(bsonutil.OpLt, args[0], astutil.FloatValue(-1.0)),
			ast.NewBinary(bsonutil.OpGt, args[0], astutil.FloatValue(1.0)),
		),
	), nil
}

func (f *baseScalarFunctionExpr) asinToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	assignments, args := minimizeLetAssignments([]string{"input"}, args)

	// MySQL returns NULL for values outside of the range [-1,1].
	// asin(x) =  pi/2 - cos(x) via the identity:
	// asin(x) + acos(x) = pi/2.
	return wrapInLet(assignments,
		astutil.WrapInCond(
			astutil.NullLiteral,
			ast.NewBinary(bsonutil.OpSubtract,
				astutil.PiOverTwoLiteral,
				astutil.WrapInAcosComputation(args[0]),
			),
			ast.NewBinary(bsonutil.OpLt, args[0], astutil.FloatValue(-1.0)),
			ast.NewBinary(bsonutil.OpGt, args[0], astutil.FloatValue(1.0)),
		),
	), nil
}

func (f *baseScalarFunctionExpr) ceilToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return ast.NewFunction(bsonutil.OpCeil, args[0]), nil
}

func (f *baseScalarFunctionExpr) characterLengthToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpStrlenCP, exprs[0], t)
}

func (f *baseScalarFunctionExpr) concatToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return astutil.WrapInConcat(args), nil
}

func (f *baseScalarFunctionExpr) concatWsToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	pushArgs := make([]ast.Expr, 0, len(args)*2)

	appendArgs := func(v, cond ast.Expr) {
		pushArgs = append(pushArgs,
			astutil.WrapInCond(astutil.EmptyStringLiteral, v, cond),
			astutil.WrapInCond(astutil.EmptyStringLiteral, args[0], cond),
		)
	}

	columnsToNullCheck := t.ColumnsToNullCheck()

	for _, arg := range args[1:] {
		switch a := arg.(type) {
		case *ast.FieldRef:
			if astutil.AllParentsAreFieldRefs(a) {
				columnName := astutil.FieldRefString(a)
				columnsToNullCheck[columnName] = struct{}{}
				appendArgs(a, toNullCheckedLetVarRef(columnName))
			} else {
				// for cases where a parent is an ast.VariableRef (which should not be
				// null-checked at the top level).
				appendArgs(a, astutil.WrapInNullCheck(a))
			}
		default:
			appendArgs(arg, astutil.WrapInNullCheck(arg))
		}
	}

	return astutil.WrapInConcat(pushArgs[:len(pushArgs)-1]), nil
}

func (f *baseScalarFunctionExpr) convToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	originalBaseRef := ast.NewVariableRef("originalBase")
	newBaseRef := ast.NewVariableRef("newBase")
	nonNegativeNumberRef := ast.NewVariableRef("nonNegativeNumber")

	// length is how long (in digits) the input number is
	normalizedVars := []*ast.LetVariable{
		ast.NewLetVariable("originalBase", ast.NewFunction(bsonutil.OpAbs, oldBase)),
		ast.NewLetVariable("newBase", ast.NewFunction(bsonutil.OpAbs, newBase)),
		ast.NewLetVariable("negative", ast.NewBinary(bsonutil.OpEq,
			astutil.StringValue("-"),
			astutil.WrapInOp(bsonutil.OpSubstr, num, astutil.ZeroInt32Literal, astutil.OneInt32Literal),
		)),
		ast.NewLetVariable("nonNegativeNumber", astutil.WrapInCond(
			astutil.WrapInOp(bsonutil.OpSubstr, num, astutil.OneInt32Literal,
				astutil.WrapInOp(bsonutil.OpSubtract,
					astutil.WrapInOp(bsonutil.OpStrlenCP, num),
					astutil.OneInt32Literal,
				),
			),
			num,
			astutil.WrapInOp(bsonutil.OpEq,
				astutil.StringValue("-"),
				astutil.WrapInOp(bsonutil.OpSubstr, num, astutil.ZeroInt32Literal, astutil.OneInt32Literal),
			),
		)),
	}

	decimalIndexRef := ast.NewVariableRef("decimalIndex")
	indexOfDecimal := []*ast.LetVariable{
		ast.NewLetVariable("decimalIndex",
			astutil.WrapInOp(bsonutil.OpIndexOfCP, nonNegativeNumberRef, astutil.StringValue("."))),
	}

	numberRef := ast.NewVariableRef("number")
	eliminateDecimal := []*ast.LetVariable{
		ast.NewLetVariable("number", astutil.WrapInCond(nonNegativeNumberRef,
			astutil.WrapInOp(bsonutil.OpSubstr, nonNegativeNumberRef, astutil.ZeroInt32Literal, decimalIndexRef),
			ast.NewBinary(bsonutil.OpEq, decimalIndexRef, astutil.Int32Value(-1)))),
	}

	lengthRef := ast.NewVariableRef("length")
	createLength := []*ast.LetVariable{
		ast.NewLetVariable("length", ast.NewFunction(bsonutil.OpStrlenCP, numberRef)),
	}

	// indexArr is an array of numbers from 0 to n-1 when n = length
	createIndexArr := []*ast.LetVariable{
		ast.NewLetVariable("indexArr", astutil.WrapInRange(astutil.ZeroInt32Literal, lengthRef, astutil.OneInt32Literal)),
	}

	// charArr breaks the number entered into an array of characters where each char is a digit
	createCharArr := []*ast.LetVariable{
		ast.NewLetVariable("charArr", astutil.WrapInMap(
			ast.NewVariableRef("indexArr"),
			"this",
			ast.NewArray(
				astutil.ThisVarRef,
				astutil.WrapInOp(bsonutil.OpSubstr, numberRef, astutil.ThisVarRef, astutil.OneInt32Literal),
			),
		)),
	}

	// This logic takes in the charArr and outputs a 2D array containing the index and the
	// base10 numerical value of the character.
	// i.e. if charArr = ["3", "A", "2"], numArr = [[0, 3], [1, 10], [2, 2]]
	branches1 := make([]ast.Expr, len(validNumbers))
	for i, k := range validNumbers {
		branches1[i] = astutil.WrapInCase(
			ast.NewBinary(bsonutil.OpEq,
				astutil.WrapInOp(bsonutil.OpArrElemAt, astutil.ThisVarRef, astutil.OneInt32Literal),
				astutil.StringValue(k),
			),
			ast.NewArray(
				astutil.WrapInOp(bsonutil.OpArrElemAt, astutil.ThisVarRef, astutil.ZeroInt32Literal),
				astutil.Int32Value(stringToNum[k]),
			),
		)
	}

	numArrRef := ast.NewVariableRef("numArr")
	createNumArr := []*ast.LetVariable{
		ast.NewLetVariable("numArr", ast.NewFunction(bsonutil.OpMap, ast.NewDocument(
			ast.NewDocumentElement("input", ast.NewVariableRef("charArr")),
			ast.NewDocumentElement("in", astutil.WrapInSwitch(
				ast.NewArray(astutil.ZeroInt32Literal, astutil.Int32Value(100)), branches1...)),
		))),
	}

	// invalidArr has False for every digit that is valid, and True for every digit that is invalid
	// In order for the input string to be converted to a new number base every entry in this
	// array must be False.
	createInvalidArr := []*ast.LetVariable{
		ast.NewLetVariable("invalidArr", astutil.WrapInMap(
			numArrRef,
			"this",
			ast.NewBinary(bsonutil.OpGte,
				astutil.WrapInOp(bsonutil.OpArrElemAt, astutil.ThisVarRef, astutil.OneInt32Literal),
				originalBaseRef,
			),
		)),
	}

	base10Ref := ast.NewVariableRef("base10")

	// Given a charArr = [[1, x1]...[i, xi]...[n, xn]] and a base b,
	// This implements the logic: sum(b^(n-i-1) * xi) with i = 0->n-1
	generateBase10 := []*ast.LetVariable{
		ast.NewLetVariable("base10", astutil.WrapInOp(bsonutil.OpSum,
			astutil.WrapInMap(numArrRef, "this",
				ast.NewBinary(bsonutil.OpMultiply,
					ast.NewBinary(bsonutil.OpArrElemAt, astutil.ThisVarRef, astutil.OneInt32Literal),
					ast.NewBinary(bsonutil.OpPow,
						originalBaseRef,
						ast.NewBinary(bsonutil.OpSubtract,
							ast.NewBinary(bsonutil.OpSubtract,
								lengthRef,
								ast.NewBinary(bsonutil.OpArrElemAt, astutil.ThisVarRef, astutil.ZeroInt32Literal),
							),
							astutil.OneInt32Literal,
						),
					),
				),
			),
		)),
	}

	// numDigits is the length the number will be in the new number base
	// This is equal to: floor(log_newbase(num)) + 1
	numDigits := []*ast.LetVariable{
		ast.NewLetVariable("numDigits", astutil.WrapInOp(bsonutil.OpAdd,
			astutil.WrapInOp(bsonutil.OpFloor,
				astutil.WrapInOp(bsonutil.OpLog, base10Ref, newBaseRef)), astutil.OneInt32Literal)),
	}

	// powers is an array of the powers of the base that you are translating to
	// if the newBase=16 and the resulting number will have length=4 this array
	// will = [1, 16, 256, 4096]
	powers := []*ast.LetVariable{
		ast.NewLetVariable("powers", astutil.WrapInMap(
			astutil.WrapInRange(
				ast.NewBinary(bsonutil.OpSubtract, ast.NewVariableRef("numDigits"), astutil.OneInt32Literal),
				astutil.Int32Value(-1), astutil.Int32Value(-1)),
			"this",
			astutil.WrapInOp(bsonutil.OpPow, newBaseRef, astutil.ThisVarRef),
		)),
	}

	// Turns the base10 number into an array of the newBase digits (in their base10 form)
	// i.e. if base10 = 173 (0xAD), numbersArray = [10, 13]
	// Follows generalized version of: https://www.permadi.com/tutorial/numDecToHex/
	generateNumberArray := astutil.WrapInMap(
		ast.NewVariableRef("powers"),
		"this",
		ast.NewBinary(bsonutil.OpMod,
			ast.NewFunction(bsonutil.OpFloor,
				ast.NewBinary(bsonutil.OpDivide, base10Ref, astutil.ThisVarRef),
			),
			newBaseRef,
		),
	)

	branches2 := make([]ast.Expr, len(numToString))
	for k := 0; k < len(numToString); k++ {
		branches2[k] = astutil.WrapInCase(
			astutil.WrapInOp(bsonutil.OpEq, astutil.ThisVarRef, astutil.Int32Value(int32(k))),
			astutil.StringValue(numToString[int32(k)]),
		)
	}

	// Converts the number array into an array of their character representations
	// i.e. if numbersArray = [10, 13], then charArray=['A', 'D']
	generateCharArray := astutil.WrapInMap(
		generateNumberArray,
		"this",
		astutil.WrapInSwitch(astutil.StringValue("0"), branches2...),
	)

	positiveAnswerRef := ast.NewVariableRef("positiveAnswer")

	// Turns the charArray into a single string (the final answer)
	// i.e. if charArray=['A','D'] answer='AD'
	positiveAnswer := []*ast.LetVariable{
		ast.NewLetVariable("positiveAnswer", astutil.WrapInReduce(
			generateCharArray,
			astutil.EmptyStringLiteral,
			astutil.WrapInOp(bsonutil.OpConcat,
				astutil.EmptyStringLiteral,
				astutil.ValueVarRef,
				astutil.ThisVarRef,
			),
		)),
	}

	signAdjusted := astutil.WrapInCond(
		astutil.WrapInOp(bsonutil.OpConcat, astutil.StringValue("-"), positiveAnswerRef),
		positiveAnswerRef,
		ast.NewVariableRef("negative"),
	)

	// Puts the nested lets together, checks to make sure that the base is valid,
	// and checks to make sure the entered number is valid as well
	// (invalid = numbers too big like 3 in binary or non-alphanumeric like /)
	// Invalid characters returns an answer of 0, invalid bases return NULL
	return wrapInNullCheckedCond(t.ColumnsToNullCheck(), astutil.NullLiteral, ast.NewLet(normalizedVars,
		ast.NewLet(indexOfDecimal,
			ast.NewLet(eliminateDecimal,
				astutil.WrapInCond(astutil.NullLiteral,
					astutil.WrapInCond(astutil.StringValue("0"),
						ast.NewLet(createLength,
							ast.NewLet(createIndexArr,
								ast.NewLet(createCharArr,
									ast.NewLet(createNumArr,
										ast.NewLet(createInvalidArr,
											astutil.WrapInCond(astutil.StringValue("0"),
												ast.NewLet(generateBase10,
													ast.NewLet(numDigits,
														ast.NewLet(powers,
															ast.NewLet(positiveAnswer,
																signAdjusted)))),
												astutil.WrapInOp(bsonutil.OpAnyElementTrue,
													ast.NewVariableRef("invalidArr")))))))),
						astutil.WrapInOp(bsonutil.OpIn, numberRef, ast.NewArray(astutil.StringValue("0"), astutil.StringValue("-0")))),
					astutil.WrapInOp(bsonutil.OpOr,
						astutil.WrapInOp(bsonutil.OpOr, astutil.WrapInOp(bsonutil.OpLt, originalBaseRef, astutil.Int32Value(2)),
							astutil.WrapInOp(bsonutil.OpGt, originalBaseRef, astutil.Int32Value(36))),
						astutil.WrapInOp(bsonutil.OpOr, astutil.WrapInOp(bsonutil.OpLt, newBaseRef, astutil.Int32Value(2)),
							astutil.WrapInOp(bsonutil.OpGt, newBaseRef, astutil.Int32Value(36))))))),
	), args[0]), nil
}

func (f *baseScalarFunctionExpr) convertToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	typ, ok := evalTypeFromSQLTypeExpr(exprs[1])
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(convert)",
			fmt.Sprintf(
				"cannot push down conversions to %s",
				exprs[1].(SQLValueExpr).Value.String(),
			),
		)
	}

	return NewSQLConvertExpr(exprs[0], typ).ToAggregationLanguage(t)
}

func (f *baseScalarFunctionExpr) cosToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	input := ast.NewVariableRef("input")
	inputLetAssignment := []*ast.LetVariable{
		ast.NewLetVariable("input", ast.NewFunction(bsonutil.OpAbs, args[0])),
	}

	rem, phase := ast.NewVariableRef("rem"), ast.NewVariableRef("phase")
	remPhaseAssignment := []*ast.LetVariable{
		ast.NewLetVariable("rem", ast.NewBinary(bsonutil.OpMod, input, astutil.PiOverTwoLiteral)),
		ast.NewLetVariable("phase", ast.NewBinary(bsonutil.OpMod,
			ast.NewFunction(bsonutil.OpTrunc,
				astutil.WrapInOp(bsonutil.OpDivide, input, astutil.PiOverTwoLiteral),
			),
			astutil.FloatValue(4.0),
		)),
	}

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
	threeCase := astutil.WrapInCond(
		astutil.WrapInSinPowerSeries(rem),
		astutil.NullLiteral,
		ast.NewBinary(bsonutil.OpEq, phase, astutil.Int32Value(3)),
	)
	twoCase := astutil.WrapInCond(
		ast.NewBinary(bsonutil.OpMultiply, astutil.FloatValue(-1.0), astutil.WrapInCosPowerSeries(rem)),
		threeCase,
		ast.NewBinary(bsonutil.OpEq, phase, astutil.Int32Value(2)),
	)
	oneCase := astutil.WrapInCond(
		astutil.WrapInOp(bsonutil.OpMultiply, astutil.FloatValue(-1.0), astutil.WrapInSinPowerSeries(rem)),
		twoCase,
		astutil.WrapInOp(bsonutil.OpEq, phase, astutil.OneInt32Literal),
	)
	zeroCase := astutil.WrapInCond(
		astutil.WrapInCosPowerSeries(rem),
		oneCase,
		astutil.WrapInOp(bsonutil.OpEq, phase, astutil.ZeroInt32Literal),
	)

	return ast.NewLet(inputLetAssignment,
		ast.NewLet(remPhaseAssignment, zeroCase),
	), nil
}

func (f *baseScalarFunctionExpr) cotToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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
	epsilon := astutil.FloatValue(6.123233995736766e-17)
	// Replace abs(denom) < epsilon with epsilon since mysql doesn't seem
	// to return NULL or INF for inf values of tan as one would expect, and
	// we do not want to trigger a $divide by 0.
	return ast.NewBinary(bsonutil.OpDivide,
		num,
		astutil.WrapInCond(
			epsilon,
			denom,
			ast.NewBinary(bsonutil.OpLte,
				ast.NewFunction(bsonutil.OpAbs, denom),
				epsilon,
			),
		),
	), nil
}

// nolint: unparam
func (f *baseScalarFunctionExpr) currentDateToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	now := time.Now().In(schema.DefaultLocale)
	cd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, schema.DefaultLocale)
	return astutil.DateConstant(cd), nil
}

// nolint: unparam
func (f *baseScalarFunctionExpr) currentTimestampToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	now := time.Now().In(schema.DefaultLocale)
	return astutil.DateConstant(now), nil
}

func (f *baseScalarFunctionExpr) dateAddToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	return f.dateArithmeticToAggregationLanguage(t, exprs, false)
}

func (f *baseScalarFunctionExpr) dateArithmeticToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr, isSub bool) (ast.Expr, PushdownFailure) {
	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateArithmetic)",
			incorrectArgCountMsg,
		)
	}

	var date ast.Expr
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
		case types.EvalDate, types.EvalDatetime:
		default:
			return nil, newPushdownFailure(
				"SQLScalarFunctionExpr(dateArithmetic)",
				"cannot push down when first arg is types.EvalDate or types.EvalDatetime",
			)
		}

		if date, err = t.ToAggregationLanguage(exprs[0]); err != nil {
			return nil, err
		}
	}

	intervalValueExpr, ok := exprs[1].(SQLValueExpr)
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateArithmetic)",
			"cannot push down without literal interval value",
		)
	}
	intervalValue := intervalValueExpr.Value

	if values.Float64(intervalValue) == 0 {
		return date, nil
	}

	unitValueExpr, ok := exprs[2].(SQLValueExpr)
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateArithmetic)",
			"cannot push down without literal unit value",
		)
	}
	unitValue := unitValueExpr.Value

	var ms int64
	// Second can be a float rather than an int, so handle Second specially.
	// calculateInterval works for all other units, as they must be integral.
	if unitValue.String() == Second {
		ms = mathutil.Round(values.Float64(intervalValue) * 1000.0)
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

	assignments, dateArg := minimizeLetAssignments([]string{"date"}, []ast.Expr{date})

	date = dateArg[0]
	evaluation := wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		astutil.WrapInOp(bsonutil.OpAdd, date, astutil.Int64Value(ms)),
		date,
	)

	return wrapInLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) dateDiffToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateDiff)",
			incorrectArgCountMsg,
		)
	}

	var date1, date2 ast.Expr
	var err PushdownFailure

	parseArgs := func(expr SQLExpr) (ast.Expr, PushdownFailure) {
		var value values.SQLValue
		if valueExpr, ok := expr.(SQLValueExpr); ok {
			value = valueExpr.Value
			var date time.Time
			date, _, ok = values.StrToDateTime(value.String(), false)
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
			return astutil.DateConstant(date), nil
		}
		exprType := expr.EvalType()
		if exprType == types.EvalDatetime || exprType == types.EvalDate {
			var date ast.Expr
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
	days := ast.NewFunction(bsonutil.OpTrunc,
		ast.NewBinary(bsonutil.OpDivide,
			ast.NewBinary(bsonutil.OpSubtract, date1, date2),
			astutil.Int32Value(86400000),
		),
	)

	upper, lower := astutil.Int32Value(106751), astutil.Int32Value(-106751)
	bound := astutil.WrapInCond(upper, lower, ast.NewBinary(bsonutil.OpGt, days, upper))

	daysRef := ast.NewVariableRef("days")
	letAssignment := []*ast.LetVariable{
		ast.NewLetVariable("days", days),
	}

	letEvaluation := wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		astutil.WrapInCond(
			bound,
			daysRef,
			ast.NewBinary(bsonutil.OpGt, daysRef, upper),
			ast.NewBinary(bsonutil.OpLt, daysRef, lower),
		),
		date1, date2,
	)

	return ast.NewLet(letAssignment, letEvaluation), nil
}

func (f *baseScalarFunctionExpr) dateFormatToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	formatValueExpr, ok := exprs[1].(SQLValueExpr)
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateFormat)",
			"format string was not a literal",
		)
	}
	formatValue := formatValueExpr.Value

	conds, containsNullLiteral := minimizeNullChecks(t.ColumnsToNullCheck(), date)

	wrapped, ok := astutil.WrapInDateFormat(date, formatValue.String(), conds...)
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateFormat)",
			"unable to push down format string",
			"formatString", formatValue.String(),
		)
	}

	if containsNullLiteral {
		return astutil.NullLiteral, nil
	}

	return wrapped, nil
}

func (f *baseScalarFunctionExpr) dateSubToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	return f.dateArithmeticToAggregationLanguage(t, exprs, true)
}

func (f *baseScalarFunctionExpr) dayNameToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		astutil.WrapInOp(bsonutil.OpArrElemAt,
			ast.NewArray(
				astutil.StringValue(time.Sunday.String()),
				astutil.StringValue(time.Monday.String()),
				astutil.StringValue(time.Tuesday.String()),
				astutil.StringValue(time.Wednesday.String()),
				astutil.StringValue(time.Thursday.String()),
				astutil.StringValue(time.Friday.String()),
				astutil.StringValue(time.Saturday.String()),
			),
			ast.NewBinary(bsonutil.OpSubtract,
				astutil.WrapInOp(bsonutil.OpDayOfWeek, args[0]),
				astutil.OneInt32Literal,
			),
		),
		args...,
	), nil
}

func (f *baseScalarFunctionExpr) dayOfMonthToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dayOfMonth)",
			incorrectArgCountMsg,
		)
	}

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpDayOfMonth, exprs[0], t)
}

func (f *baseScalarFunctionExpr) dayOfWeekToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dayOfWeek)",
			incorrectArgCountMsg,
		)
	}

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpDayOfWeek, exprs[0], t)
}

func (f *baseScalarFunctionExpr) dayOfYearToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dayOfYear)",
			incorrectArgCountMsg,
		)
	}

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpDayOfYear, exprs[0], t)
}

func (f *baseScalarFunctionExpr) degreesToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return ast.NewBinary(bsonutil.OpDivide,
		ast.NewBinary(bsonutil.OpMultiply, args[0], astutil.FloatValue(180.0)),
		astutil.PiLiteral,
	), nil
}

func (f *baseScalarFunctionExpr) expToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return ast.NewFunction(bsonutil.OpExp, args[0]), nil
}

func (f *baseScalarFunctionExpr) extractToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(extract)",
			incorrectArgCountMsg,
		)
	}

	unitArg, err := t.ToAggregationLanguage(exprs[0])
	if err != nil {
		return nil, err
	}

	var unit string
	if s, ok := unitArg.(*ast.Constant); ok {
		if unit, ok = s.Value.StringValueOK(); !ok {
			// The unit must absolutely be a string.
			return nil, newPushdownFailure(
				"SQLScalarFunctionExpr(extract)",
				"first argument was not a string",
			)
		}
	} else {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(extract)",
			"first argument was not a literal",
		)
	}

	switch unit {
	case "year", "month", "hour", "minute", "second":
		return wrapSingleArgFuncWithNullCheck("$"+unit, exprs[1], t)
	case "day":
		return wrapSingleArgFuncWithNullCheck(bsonutil.OpDayOfMonth, exprs[1], t)
	}
	return nil, newPushdownFailure(
		"SQLScalarFunctionExpr(extract)",
		"unknown unit",
		"unit", unit,
	)
}

func (f *baseScalarFunctionExpr) floorToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return ast.NewFunction(bsonutil.OpFloor, args[0]), nil
}

func (f *baseScalarFunctionExpr) fromDaysToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	assignments, args := minimizeLetAssignments([]string{"n"}, args)
	n := args[0]

	// This should return "0000-00-00" if the input is too large (> maxFromDays)
	// or too low (< 366).
	dayOne := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	body := ast.NewBinary(bsonutil.OpAdd,
		astutil.DateConstant(dayOne),
		ast.NewBinary(bsonutil.OpMultiply,
			astutil.WrapInRound(n), astutil.FloatValue(millisecondsPerDay),
		),
	)

	evaluation := wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		astutil.WrapInCond(
			astutil.NullLiteral,
			body,
			ast.NewBinary(bsonutil.OpGt, n, astutil.Int32Value(maxFromDays)),
			ast.NewBinary(bsonutil.OpLt, n, astutil.Int32Value(366)),
		),
		args...,
	)

	return wrapInLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) fromUnixtimeToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	unixTimestamp := ast.NewVariableRef("unixTimestamp")
	assignment := []*ast.LetVariable{
		ast.NewLetVariable("unixTimestamp", args[0]),
	}

	// Just add the argument to 1970-01-01 00:00:00.0000000.
	dayOne := time.Date(1970, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
	evaluation := ast.NewBinary(bsonutil.OpAdd,
		astutil.DateConstant(dayOne),
		ast.NewBinary(bsonutil.OpMultiply,
			astutil.WrapInRound(unixTimestamp), astutil.Int32Value(1e3),
		),
	)

	ret := ast.NewLet(assignment,
		astutil.WrapInCond(
			astutil.NullLiteral,
			evaluation,
			ast.NewBinary(bsonutil.OpLt, unixTimestamp, astutil.ZeroInt32Literal),
		),
	)

	if len(exprs) == 1 {
		return ret, nil
	}
	if formatExpr, ok := exprs[1].(SQLValueExpr); ok {
		format := formatExpr.Value
		wrapped, ok := astutil.WrapInDateFormat(ret, format.String())
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

func (f *baseScalarFunctionExpr) greatestToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		astutil.WrapInOp(bsonutil.OpMax, args...),
		args...,
	), nil
}

func (f *baseScalarFunctionExpr) hourToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(hour)",
			incorrectArgCountMsg,
		)
	}

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpHour, exprs[0], t)
}

func (f *baseScalarFunctionExpr) insertToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	// subtract 1 to account for difference between mongo and mysql string indexing
	args[1] = astutil.WrapInRound(ast.NewBinary(bsonutil.OpSubtract, args[1], astutil.OneInt32Literal))

	args[2] = astutil.WrapInRound(args[2])

	assignments, args := minimizeLetAssignments(
		[]string{"str", "pos", "len", "newstr"},
		args,
	)

	str := args[0]
	pos := args[1]
	length := args[2]
	newstr := args[3]

	totalLength := ast.NewVariableRef("totalLength")
	totalLengthAssignment := []*ast.LetVariable{
		ast.NewLetVariable("totalLength", ast.NewFunction(bsonutil.OpStrlenCP, str)),
	}

	prefix, suffix := ast.NewVariableRef("prefix"), ast.NewVariableRef("suffix")
	ixAssignment := []*ast.LetVariable{
		ast.NewLetVariable("prefix", astutil.WrapInOp(bsonutil.OpSubstr, str, astutil.ZeroInt32Literal, pos)),
		ast.NewLetVariable("suffix", astutil.WrapInOp(bsonutil.OpSubstr, str, astutil.WrapInOp(bsonutil.OpAdd, pos, length), totalLength)),
	}

	concatenation := ast.NewLet(ixAssignment,
		astutil.WrapInOp(bsonutil.OpConcat, prefix, newstr, suffix),
	)

	posCheck := ast.NewLet(totalLengthAssignment,
		astutil.WrapInCond(str,
			concatenation,
			ast.NewBinary(bsonutil.OpLt, pos, astutil.ZeroInt32Literal),
			ast.NewBinary(bsonutil.OpGte, pos, totalLength),
		),
	)

	evaluation := wrapInNullCheckedCond(t.ColumnsToNullCheck(), astutil.NullLiteral, posCheck, args...)

	return wrapInLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) instrToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	assignments, substrArg := minimizeLetAssignments([]string{"substr"}, args[1:])
	str := args[0]
	substr := substrArg[0]

	// Mongo Aggregation Pipeline returns NULL if str is NULLish, like
	// we'd want. substr being NULL, however, is an error in the pipeline,
	// thus check substr for NULLisness.
	evaluation := wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		ast.NewBinary(bsonutil.OpAdd,
			astutil.OneInt32Literal,
			astutil.WrapInOp(bsonutil.OpIndexOfCP, str, substr),
		),
		substrArg[0],
	)

	return wrapInLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) lastDayToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	date := ast.NewVariableRef("date")
	outerLetAssignment := []*ast.LetVariable{
		ast.NewLetVariable("date", args[0]),
	}

	letAssigment := []*ast.LetVariable{
		ast.NewLetVariable("year", ast.NewFunction(bsonutil.OpYear, date)),
		ast.NewLetVariable("month", ast.NewFunction(bsonutil.OpMonth, date)),
	}

	year, month := ast.NewVariableRef("year"), ast.NewVariableRef("month")
	var letEvaluation *ast.Document

	// Underflow and overflow in date computation are supported on MongoDB versions >= 4.0.
	// For example, a month value greater than 12 (overflow) and a day value of zero
	// (underflow) are supported date values.
	if t.versionAtLeast(4, 0, 0) {
		// MongoDB interprets day 0 of a given month as the last day of the previous month.
		letEvaluation = ast.NewDocument(
			ast.NewDocumentElement(bsonutil.OpDateFromParts, ast.NewDocument(
				ast.NewDocumentElement("year", year),
				ast.NewDocumentElement("month", ast.NewBinary(bsonutil.OpAdd, astutil.OneInt32Literal, month)),
				ast.NewDocumentElement("day", astutil.ZeroInt32Literal),
			)),
		)

	} else {
		const30 := astutil.Int32Value(30)
		// For MongoDB versions < 4.0, underflow and overflow in date computation are not
		// supported. For example, a day value of zero or a month value of 13 in a date
		// generates an error. In this case, we create a switch on the month value,
		// extracted from $dateFromParts, to determine the last day of the month.
		letEvaluation = ast.NewDocument(
			ast.NewDocumentElement(bsonutil.OpDateFromParts, ast.NewDocument(
				ast.NewDocumentElement("year", year),
				ast.NewDocumentElement("month", month),
				ast.NewDocumentElement("day",
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
					astutil.WrapInSwitch(
						astutil.Int32Value(31),
						astutil.WrapInEqCase(month, astutil.Int32Value(2),
							astutil.WrapInCond(astutil.Int32Value(29), astutil.Int32Value(28), astutil.WrapInIsLeapYear(year)),
						),
						astutil.WrapInEqCase(month, astutil.Int32Value(4), const30),
						astutil.WrapInEqCase(month, astutil.Int32Value(6), const30),
						astutil.WrapInEqCase(month, astutil.Int32Value(9), const30),
						astutil.WrapInEqCase(month, astutil.Int32Value(11), const30),
					)),
			)),
		)
	}

	return ast.NewLet(outerLetAssignment, ast.NewLet(letAssigment, letEvaluation)), nil
}

func (f *baseScalarFunctionExpr) lcaseToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(lcase)",
			incorrectArgCountMsg,
		)
	}

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpToLower, exprs[0], t)
}

func (f *baseScalarFunctionExpr) leastToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		astutil.WrapInOp(bsonutil.OpMin, args...),
		args...,
	), nil
}

func (f *baseScalarFunctionExpr) leftToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	assignments, args := minimizeLetAssignments([]string{"str", "length"}, args)

	str := args[0]
	subStrLength := astutil.WrapInRound(astutil.WrapInOp(bsonutil.OpMax, args[1], astutil.ZeroInt32Literal))
	subStrOp := astutil.WrapInOp(bsonutil.OpSubstr, str, astutil.ZeroInt32Literal, subStrLength)
	evaluation := wrapInNullCheckedCond(t.ColumnsToNullCheck(), astutil.NullLiteral, subStrOp, args...)

	return wrapInLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) lengthToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpStrLenBytes, exprs[0], t)
}

func (f *baseScalarFunctionExpr) locateToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	var locate ast.Expr
	substr := args[0]
	str := args[1]

	if len(args) == 2 {
		indexOfCP := astutil.WrapInOp(bsonutil.OpIndexOfCP, str, substr)
		locate = ast.NewBinary(bsonutil.OpAdd, indexOfCP, astutil.OneInt32Literal)
	} else if len(args) == 3 {
		// if the pos arg is null, we should return 0, not null
		// this is the same result as when the arg is 0
		pos := astutil.WrapInIfNull(args[2], astutil.ZeroInt32Literal)

		// round to the nearest int
		pos = ast.NewBinary(bsonutil.OpAdd, pos, astutil.FloatValue(0.5))
		pos = ast.NewFunction(bsonutil.OpTrunc, pos)

		// subtract 1 from the pos arg to reconcile indexing style
		pos = ast.NewBinary(bsonutil.OpSubtract, pos, astutil.OneInt32Literal)

		indexOfCP := astutil.WrapInOp(bsonutil.OpIndexOfCP, str, substr, pos)
		locate = ast.NewBinary(bsonutil.OpAdd, indexOfCP, astutil.OneInt32Literal)

		// if the pos argument was negative, we should return 0
		locate = astutil.WrapInCond(
			astutil.ZeroInt32Literal,
			locate,
			ast.NewBinary(bsonutil.OpLt, pos, astutil.ZeroInt32Literal),
		)
	}

	return wrapInNullCheckedCond(t.ColumnsToNullCheck(), astutil.NullLiteral, locate, args[0], args[1]), nil
}

func (f *baseScalarFunctionExpr) log10ToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	return f.logarithmToAggregationLanguage(t, exprs, 10)
}

func (f *baseScalarFunctionExpr) log2ToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	return f.logarithmToAggregationLanguage(t, exprs, 2)
}

func (f *baseScalarFunctionExpr) logToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	return f.logarithmToAggregationLanguage(t, exprs, 0)
}

func (f *baseScalarFunctionExpr) logarithmToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr, base uint32) (ast.Expr, PushdownFailure) {
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
			return astutil.WrapInCond(
				ast.NewFunction(bsonutil.OpNaturalLog, args[0]),
				astutil.NullLiteral,
				ast.NewBinary(bsonutil.OpGt, args[0], astutil.ZeroInt32Literal),
			), nil
		}
		// Two args is based arg.
		// MySQL specifies base then arg, MongoDB expects arg then base, so we have to flip.
		return astutil.WrapInCond(
			ast.NewBinary(bsonutil.OpLog, args[1], args[0]),
			astutil.NullLiteral,
			ast.NewBinary(bsonutil.OpGt, args[0], astutil.ZeroInt32Literal),
		), nil
	}
	// This will be base 10 or base 2 based on if log10 or log2 was called.
	return astutil.WrapInCond(
		ast.NewBinary(bsonutil.OpLog, args[0], astutil.Int64Value(int64(base))),
		astutil.NullLiteral,
		ast.NewBinary(bsonutil.OpGt, args[0], astutil.ZeroInt32Literal),
	), nil
}

func (f *baseScalarFunctionExpr) lpadToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	return f.padToAggregationLanguage(t, exprs, true)
}

func (f *baseScalarFunctionExpr) ltrimToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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
		return ast.NewFunction(bsonutil.OpLTrim, ast.NewDocument(
			ast.NewDocumentElement("input", args[0]),
			ast.NewDocumentElement("chars", astutil.StringValue(" ")),
		)), nil
	}

	ltrimCond := astutil.WrapInCond(
		astutil.EmptyStringLiteral,
		astutil.WrapInLRTrim(true, args[0]),
		ast.NewBinary(bsonutil.OpEq, args[0], astutil.EmptyStringLiteral),
	)

	return wrapInNullCheckedCond(t.ColumnsToNullCheck(), astutil.NullLiteral, ltrimCond, args...), nil
}

func (f *baseScalarFunctionExpr) makeDateToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	year, day := ast.NewVariableRef("year"), ast.NewVariableRef("day")
	inputLetStatement := []*ast.LetVariable{
		ast.NewLetVariable("year", astutil.WrapInRound(args[0])),
		ast.NewLetVariable("day", astutil.WrapInRound(args[1])),
	}

	branch1900 := astutil.WrapInCond(
		ast.NewBinary(bsonutil.OpAdd, year, astutil.Int32Value(1900)),
		year,
		ast.NewBinary(bsonutil.OpAnd,
			ast.NewBinary(bsonutil.OpGte, year, astutil.Int32Value(70)),
			ast.NewBinary(bsonutil.OpLte, year, astutil.Int32Value(99)),
		))

	branch2000 := ast.NewBinary(bsonutil.OpAdd, year, astutil.Int32Value(2000))

	paddedYear := ast.NewVariableRef("paddedYear")

	// $$paddedYear holds the year + 2000 for years between 0 and 69, and +
	// 1900 for years between 70 and 99. Otherwise, it is the original
	// year.
	paddedYearLetStatement := []*ast.LetVariable{
		ast.NewLetVariable("paddedYear", astutil.WrapInCond(
			branch2000,
			branch1900,
			ast.NewBinary(bsonutil.OpAnd,
				ast.NewBinary(bsonutil.OpGte, year, astutil.ZeroInt32Literal),
				ast.NewBinary(bsonutil.OpLte, year, astutil.Int32Value(69))),
		)),
	}

	// This implements:
	// date(paddedYear) + (day - 1) * millisecondsPerDay.
	addDaysStatement := ast.NewBinary(bsonutil.OpAdd,
		ast.NewFunction(bsonutil.OpDateFromParts, ast.NewDocument(
			ast.NewDocumentElement("year", paddedYear),
		)),
		ast.NewBinary(bsonutil.OpMultiply,
			ast.NewBinary(bsonutil.OpSubtract, day, astutil.OneInt32Literal),
			astutil.FloatValue(millisecondsPerDay),
		),
	)

	// If the $$paddedYear is more than 9999 or less than 0, return NULL.
	yearRangeCheck := astutil.WrapInCond(
		astutil.NullLiteral,
		addDaysStatement,
		ast.NewBinary(bsonutil.OpLt, paddedYear, astutil.ZeroInt32Literal),
		ast.NewBinary(bsonutil.OpGt, paddedYear, astutil.Int32Value(9999)),
	)

	// Day range check, return NULL if day < 1.
	dayRangeCheck := astutil.WrapInCond(
		astutil.NullLiteral,
		yearRangeCheck,
		ast.NewBinary(bsonutil.OpLt, day, astutil.OneInt32Literal),
	)

	output := ast.NewVariableRef("output")
	outputLetStatement := []*ast.LetVariable{ast.NewLetVariable("output", dayRangeCheck)}

	// Bind lets, and check that output value year < 9999, otherwise MySQL
	// returns NULL.
	return ast.NewLet(inputLetStatement,
		ast.NewLet(paddedYearLetStatement,
			ast.NewLet(outputLetStatement,
				astutil.WrapInCond(astutil.NullLiteral, output,
					astutil.WrapInOp(bsonutil.OpGt,
						astutil.WrapInOp(bsonutil.OpYear, output),
						astutil.Int32Value(9999)))),
		)), nil

}

func (f *baseScalarFunctionExpr) microsecondToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		ast.NewBinary(bsonutil.OpMultiply,
			astutil.Int32Value(1000),
			astutil.WrapInOp(bsonutil.OpMillisecond, args[0]),
		),
		args...,
	), nil

}

func (f *baseScalarFunctionExpr) midToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(mid)",
			incorrectArgCountMsg,
		)
	}
	return f.substringToAggregationLanguage(t, exprs)
}

func (f *baseScalarFunctionExpr) minuteToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(minute)",
			incorrectArgCountMsg,
		)
	}

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpMinute, exprs[0], t)
}

func (f *baseScalarFunctionExpr) modToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return ast.NewBinary(bsonutil.OpMod, args[0], args[1]), nil
}

func (f *baseScalarFunctionExpr) monthNameToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		astutil.WrapInOp(bsonutil.OpArrElemAt,
			ast.NewArray(
				astutil.StringValue(time.January.String()),
				astutil.StringValue(time.February.String()),
				astutil.StringValue(time.March.String()),
				astutil.StringValue(time.April.String()),
				astutil.StringValue(time.May.String()),
				astutil.StringValue(time.June.String()),
				astutil.StringValue(time.July.String()),
				astutil.StringValue(time.August.String()),
				astutil.StringValue(time.September.String()),
				astutil.StringValue(time.October.String()),
				astutil.StringValue(time.November.String()),
				astutil.StringValue(time.December.String()),
			),
			ast.NewBinary(bsonutil.OpSubtract,
				ast.NewFunction(bsonutil.OpMonth, args[0]),
				astutil.OneInt32Literal,
			),
		),
		args...,
	), nil
}

func (f *baseScalarFunctionExpr) monthToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(month)",
			incorrectArgCountMsg,
		)
	}

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpMonth, exprs[0], t)
}

func (f *baseScalarFunctionExpr) padToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr, isLeftPad bool) (ast.Expr, PushdownFailure) {
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

	args[1] = astutil.WrapInRound(args[1])

	assignments, args := minimizeLetAssignments([]string{"str", "length", "padStr"}, args)

	str := args[0]
	length := args[1]
	padStr := args[2]

	padLenRef, padStrLenRef := ast.NewVariableRef("padLen"), ast.NewVariableRef("padStrLen")
	subAssignments := []*ast.LetVariable{
		ast.NewLetVariable("padStrLen", ast.NewFunction(bsonutil.OpStrlenCP, padStr)),
		ast.NewLetVariable("padLen", ast.NewBinary(bsonutil.OpSubtract,
			length,
			ast.NewFunction(bsonutil.OpStrlenCP, str),
		)),
	}

	// logic for generating padding string:

	// do we even need to add padding? only if the desired output
	// length is > length of input string.
	paddingCond := ast.NewBinary(bsonutil.OpLt, ast.NewFunction(bsonutil.OpStrlenCP, str), length)

	// number of times we need to repeat the padding string to fill space
	padStrRepeats := ast.NewFunction(bsonutil.OpCeil, ast.NewBinary(bsonutil.OpDivide, padLenRef, padStrLenRef))

	// generate an array with padStrRepeats occurrences of padStr
	padParts := ast.NewDocument(
		ast.NewDocumentElement(bsonutil.OpMap, ast.NewDocument(
			ast.NewDocumentElement("input", astutil.WrapInOp(bsonutil.OpRange, astutil.ZeroInt32Literal, padStrRepeats)),
			ast.NewDocumentElement("in", padStr),
		)),
	)

	// join occurrences together and trim to the exact length needed
	fullPad := astutil.WrapInOp(bsonutil.OpSubstr,
		astutil.WrapInReduce(
			padParts,
			astutil.EmptyStringLiteral,
			astutil.WrapInOp(bsonutil.OpConcat, astutil.ValueVarRef, astutil.ThisVarRef),
		),
		astutil.ZeroInt32Literal,
		padLenRef,
	)

	// based on length of input string, we either add the padding
	// or just take appropriate substring of input string
	var concatted *ast.Function
	if isLeftPad {
		concatted = astutil.WrapInOp(bsonutil.OpConcat, fullPad, str)
	} else {
		concatted = astutil.WrapInOp(bsonutil.OpConcat, str, fullPad)
	}

	handleConcat := astutil.WrapInCond(
		astutil.NullLiteral, concatted,
		ast.NewBinary(bsonutil.OpEq, padStrLenRef, astutil.ZeroInt32Literal),
	)

	// handle everything in the case that input length >=0
	handleNonNegativeLength := astutil.WrapInCond(
		handleConcat,
		astutil.WrapInOp(bsonutil.OpSubstr, str, astutil.ZeroInt32Literal, length),
		paddingCond,
	)

	// if length < 0, we just return null
	negativeCheck := astutil.WrapInCond(
		astutil.NullLiteral,
		handleNonNegativeLength,
		ast.NewBinary(bsonutil.OpLt, length, astutil.ZeroInt32Literal),
	)

	evaluation := wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		ast.NewLet(subAssignments, negativeCheck),
		args...,
	)

	return wrapInLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) powToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return ast.NewBinary(bsonutil.OpPow, args[0], args[1]), nil
}

func (f *baseScalarFunctionExpr) quarterToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	assignments, args := minimizeLetAssignments([]string{"date"}, args)

	date := args[0]

	one, two, three, four := astutil.OneInt32Literal, astutil.Int32Value(2), astutil.Int32Value(3), astutil.Int32Value(4)

	evaluation := wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		astutil.WrapInOp(bsonutil.OpArrElemAt,
			ast.NewArray(one, one, one, two, two, two, three, three, three, four, four, four),
			ast.NewBinary(bsonutil.OpSubtract,
				ast.NewFunction(bsonutil.OpMonth, date),
				astutil.OneInt32Literal,
			),
		),
		args...,
	)

	return wrapInLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) radiansToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return ast.NewBinary(bsonutil.OpDivide,
		ast.NewBinary(bsonutil.OpMultiply, args[0], astutil.PiLiteral),
		astutil.FloatValue(180.0),
	), nil
}

func (f *baseScalarFunctionExpr) repeatToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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
	num := astutil.WrapInRound(args[1])

	// create array w/ args[1] values e.g. [0,1,2]
	rangeArr := astutil.WrapInRange(astutil.ZeroInt32Literal, num, astutil.OneInt32Literal)

	// create array of len arg[1], with each item being arg[0]
	m := ast.NewFunction(bsonutil.OpMap, ast.NewDocument(
		ast.NewDocumentElement("input", rangeArr),
		ast.NewDocumentElement("in", str),
	))

	repeat := astutil.WrapInReduce(m, astutil.EmptyStringLiteral,
		// append all values of this array together
		astutil.WrapInOp(bsonutil.OpConcat, astutil.ThisVarRef, astutil.ValueVarRef),
	)

	return wrapInNullCheckedCond(t.ColumnsToNullCheck(), astutil.NullLiteral, repeat, args...), nil
}

func (f *baseScalarFunctionExpr) replaceToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	assignment := []*ast.LetVariable{
		ast.NewLetVariable("split", astutil.WrapInOp(bsonutil.OpSplit, args[0], args[1])),
	}

	body := astutil.WrapInReduce(
		ast.NewVariableRef("split"),
		astutil.NullLiteral,
		astutil.WrapInCond(
			astutil.ThisVarRef,
			astutil.WrapInOp(bsonutil.OpConcat, astutil.ValueVarRef, args[2], astutil.ThisVarRef),
			ast.NewBinary(bsonutil.OpEq, astutil.ValueVarRef, astutil.NullLiteral),
		),
	)

	return ast.NewLet(assignment, body), nil
}

func (f *baseScalarFunctionExpr) reverseToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	assignments, args := minimizeLetAssignments([]string{"str"}, args)

	str := args[0]
	evaluation := wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		astutil.WrapInReduce(
			astutil.WrapInOp(bsonutil.OpRange, astutil.ZeroInt32Literal, astutil.WrapInOp(bsonutil.OpStrlenCP, str)),
			astutil.EmptyStringLiteral,
			astutil.WrapInOp(bsonutil.OpConcat,
				astutil.WrapInOp(bsonutil.OpSubstr,
					str, astutil.ThisVarRef, astutil.OneInt32Literal,
				),
				astutil.ValueVarRef,
			),
		),
		args...,
	)

	return wrapInLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) rightToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	assignments, args := minimizeLetAssignments([]string{"str", "length"}, args)

	str := args[0]
	subStrLength := astutil.WrapInRound(astutil.WrapInOp(bsonutil.OpMax, astutil.ZeroInt32Literal, args[1]))
	strLength := ast.NewFunction(bsonutil.OpStrlenCP, str)

	// start = max(0, strLen - subStrLen)
	start := astutil.WrapInOp(bsonutil.OpMax, astutil.ZeroInt32Literal, ast.NewBinary(bsonutil.OpSubtract, strLength, subStrLength))

	subStrOp := astutil.WrapInOp(bsonutil.OpSubstr, str, start, subStrLength)

	evaluation := wrapInNullCheckedCond(t.ColumnsToNullCheck(), astutil.NullLiteral, subStrOp, args...)

	return wrapInLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) roundToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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
		return astutil.WrapInRound(args[0]), nil
	case 2:
		if arg1Expr, ok := exprs[1].(SQLValueExpr); ok {
			arg1 := arg1Expr.Value
			return astutil.WrapInRoundWithPrecision(args[0], values.Float64(arg1)), nil
		}
		fallthrough
	default:
		return nil, newPushdownFailure("SQLScalarFunctionExpr(round)", "unsupported form")
	}
}

func (f *baseScalarFunctionExpr) rpadToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	return f.padToAggregationLanguage(t, exprs, false)
}

func (f *baseScalarFunctionExpr) rtrimToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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
		return ast.NewFunction(bsonutil.OpRTrim, ast.NewDocument(
			ast.NewDocumentElement("input", args[0]),
			ast.NewDocumentElement("chars", astutil.StringValue(" ")),
		)), nil
	}

	rtrimCond := astutil.WrapInCond(
		astutil.EmptyStringLiteral,
		astutil.WrapInLRTrim(false, args[0]),
		astutil.WrapInOp(bsonutil.OpEq, args[0], astutil.EmptyStringLiteral),
	)

	return wrapInNullCheckedCond(t.ColumnsToNullCheck(), astutil.NullLiteral, rtrimCond, args...), nil
}

func (f *baseScalarFunctionExpr) secondToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(second)",
			incorrectArgCountMsg,
		)
	}

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpSecond, exprs[0], t)
}

func (f *baseScalarFunctionExpr) signToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		astutil.WrapInCond(
			astutil.ZeroInt32Literal,
			astutil.WrapInCond(
				astutil.OneInt32Literal,
				astutil.Int32Value(-1),
				ast.NewBinary(bsonutil.OpGt, args[0], astutil.ZeroInt32Literal),
			),
			ast.NewBinary(bsonutil.OpEq, args[0], astutil.ZeroInt32Literal),
		),
		args...,
	), nil
}

func (f *baseScalarFunctionExpr) sinToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	inputLetAssignment, args := minimizeLetAssignments([]string{"input"}, args)
	input := args[0]

	absInput := ast.NewVariableRef("absInput")
	absInputLetAssignment := []*ast.LetVariable{
		ast.NewLetVariable("absInput", ast.NewFunction(bsonutil.OpAbs, input)),
	}

	rem, phase := ast.NewVariableRef("rem"), ast.NewVariableRef("phase")
	remPhaseAssignment := []*ast.LetVariable{
		ast.NewLetVariable("rem", ast.NewBinary(bsonutil.OpMod, absInput, astutil.PiOverTwoLiteral)),
		ast.NewLetVariable("phase", ast.NewBinary(bsonutil.OpMod,
			ast.NewFunction(bsonutil.OpTrunc,
				ast.NewBinary(bsonutil.OpDivide, absInput, astutil.PiOverTwoLiteral),
			),
			astutil.FloatValue(4.0)),
		),
	}

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

	threeCase := astutil.WrapInCond(
		astutil.WrapInOp(bsonutil.OpMultiply, astutil.FloatValue(-1.0), astutil.WrapInCosPowerSeries(rem)),
		astutil.NullLiteral,
		astutil.WrapInOp(bsonutil.OpEq, phase, astutil.Int32Value(3)),
	)
	twoCase := astutil.WrapInCond(
		astutil.WrapInOp(bsonutil.OpMultiply, astutil.FloatValue(-1.0), astutil.WrapInSinPowerSeries(rem)),
		threeCase,
		astutil.WrapInOp(bsonutil.OpEq, phase, astutil.Int32Value(2)),
	)
	oneCase := astutil.WrapInCond(
		astutil.WrapInCosPowerSeries(rem),
		twoCase,
		astutil.WrapInOp(bsonutil.OpEq, phase, astutil.OneInt32Literal),
	)
	zeroCase := astutil.WrapInCond(
		astutil.WrapInSinPowerSeries(rem),
		oneCase,
		astutil.WrapInOp(bsonutil.OpEq, phase, astutil.ZeroInt32Literal),
	)

	// cos(-x) = cos(x), but sin(-x) = -sin(x), so if the original input is negative multiply by -1.
	return wrapInLet(inputLetAssignment,
		ast.NewLet(absInputLetAssignment,
			ast.NewLet(remPhaseAssignment,
				astutil.WrapInCond(zeroCase,
					ast.NewBinary(bsonutil.OpMultiply, astutil.FloatValue(-1.0), zeroCase),
					ast.NewBinary(bsonutil.OpGte, input, astutil.ZeroInt32Literal),
				),
			),
		),
	), nil
}

func (f *baseScalarFunctionExpr) spaceToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	assignments, args := minimizeLetAssignments([]string{"n"}, args)

	n := astutil.WrapInRound(args[0])
	evaluation := wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		astutil.WrapInReduce(
			astutil.WrapInRange(astutil.ZeroInt32Literal, n, astutil.OneInt32Literal),
			astutil.EmptyStringLiteral,
			astutil.WrapInOp(bsonutil.OpConcat, astutil.ValueVarRef, astutil.StringValue(" ")),
		),
		args...,
	)

	return wrapInLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) sqrtToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	return astutil.WrapInCond(
		ast.NewFunction(bsonutil.OpSqrt, args[0]),
		astutil.NullLiteral,
		ast.NewBinary(bsonutil.OpGte, args[0], astutil.ZeroInt32Literal),
	), nil
}

func (f *baseScalarFunctionExpr) substringIndexToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	delim, split := ast.NewVariableRef("delim"), ast.NewVariableRef("split")
	inputAssignment := []*ast.LetVariable{
		ast.NewLetVariable("delim", args[1]),
	}

	splitAssignment := []*ast.LetVariable{
		ast.NewLetVariable("split", astutil.WrapInOp(bsonutil.OpSlice,
			astutil.WrapInOp(bsonutil.OpSplit, args[0], delim),
			astutil.WrapInRound(args[2]),
		)),
	}

	body := astutil.WrapInReduce(split,
		astutil.NullLiteral,
		astutil.WrapInCond(astutil.ThisVarRef,
			astutil.WrapInOp(bsonutil.OpConcat, astutil.ValueVarRef, delim, astutil.ThisVarRef),
			ast.NewBinary(bsonutil.OpEq, astutil.ValueVarRef, astutil.NullLiteral),
		),
	)

	return ast.NewLet(inputAssignment,
		ast.NewLet(splitAssignment, body),
	), nil
}

func (f *baseScalarFunctionExpr) substringToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	strAssignment, strArg := minimizeLetAssignments([]string{"str"}, args[0:1])
	strCond, strIsNull := minimizeNullChecks(t.ColumnsToNullCheck(), strArg[0])
	if strIsNull {
		return astutil.NullLiteral, nil
	}

	str := strArg[0]
	pos := ast.NewVariableRef("pos")
	var length, strLen ast.Expr

	// store the string's length since it is reused in multiple places.
	strLenAssignment := make([]*ast.LetVariable, 0, 1)
	switch sa := strArg[0].(type) {
	case *ast.Constant:
		strVal, isStr := sa.Value.StringValueOK()
		if isStr {
			strLen = astutil.Int64Value(int64(len(strVal)))
		}
	default:
		strLen = ast.NewVariableRef("strLen")
		strLenAssignment = append(strLenAssignment,
			ast.NewLetVariable("strLen", ast.NewFunction(bsonutil.OpStrlenCP, str)),
		)
	}

	// get the "pos" and "length" arguments
	subAssignments := make([]*ast.LetVariable, 1, 2)
	subCondArgs := make([]ast.Expr, 1, 2)

	roundedPosRef, roundedNegPosRef := ast.NewVariableRef("roundedPos"), ast.NewVariableRef("roundedNegPos")

	// the position argument needs to be
	calculatedPos := ast.NewLet(
		[]*ast.LetVariable{
			ast.NewLetVariable("roundedPos", astutil.WrapInRound(args[1])),
			ast.NewLetVariable("roundedNegPos", astutil.WrapInRound(
				ast.NewBinary(bsonutil.OpMultiply, args[1], astutil.Int32Value(-1))),
			),
		},
		astutil.WrapInCond(
			// if the position is 0, use the string's length.
			strLen,
			astutil.WrapInCond(
				// if the position is positive, subtract 1 to account for SQL's 1-indexing.
				ast.NewBinary(bsonutil.OpSubtract, roundedPosRef, astutil.OneInt32Literal),
				// if the position is negative,
				astutil.WrapInCond(
					// subtract it from the end of the string if it is smaller than the length;
					ast.NewBinary(bsonutil.OpSubtract, strLen, roundedNegPosRef),
					// or use it directly if it is too large.
					roundedNegPosRef,
					ast.NewBinary(bsonutil.OpGte, strLen, roundedNegPosRef),
				),
				ast.NewBinary(bsonutil.OpGt, roundedPosRef, astutil.ZeroInt32Literal),
			),
			ast.NewBinary(bsonutil.OpEq, roundedPosRef, astutil.ZeroInt32Literal),
		),
	)

	// if it is not literal or a column, the null-check can (and should) be on the binding
	switch args[1].(type) {
	case *ast.Constant, *ast.FieldRef:
	default:
		args[1] = pos
	}

	subAssignments[0] = ast.NewLetVariable("pos", calculatedPos)
	subCondArgs[0] = args[1]

	if len(args) == 2 {
		// if length is not provided, use the str length.
		length = strLen
	} else {
		lengthAssignment, lengthArg := minimizeLetAssignments([]string{"length"}, args[2:])
		subAssignments = append(subAssignments, lengthAssignment...)
		length = astutil.WrapInRound(astutil.WrapInOp(bsonutil.OpMax, astutil.ZeroInt32Literal, lengthArg[0]))
		subCondArgs = append(subCondArgs, lengthArg[0])
	}

	subConds, containsNullLiteral := minimizeNullChecks(t.ColumnsToNullCheck(), subCondArgs...)
	if containsNullLiteral {
		return astutil.NullLiteral, nil
	}

	return wrapInLet(strAssignment,
		astutil.WrapInCond(
			astutil.NullLiteral,
			wrapInLet(strLenAssignment,
				wrapInLet(subAssignments,
					astutil.WrapInCond(
						astutil.NullLiteral,
						astutil.WrapInOp(bsonutil.OpSubstr, str, pos, length),
						subConds...,
					),
				),
			),
			strCond...,
		),
	), nil
}

func (f *baseScalarFunctionExpr) tanToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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
	epsilon := astutil.FloatValue(6.123233995736766e-17)
	// Replace abs(denom) < epsilon with epsilon since mysql doesn't seem
	// to return NULL or INF for inf values of tan as one would expect, and
	// we do not want to trigger a $divide by 0.
	return ast.NewBinary(bsonutil.OpDivide,
		num,
		astutil.WrapInCond(
			epsilon,
			denom,
			ast.NewBinary(bsonutil.OpLte,
				ast.NewFunction(bsonutil.OpAbs, denom),
				epsilon,
			),
		),
	), nil
}

func (f *baseScalarFunctionExpr) timestampAddToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	// This is a very large and costly expression, make sure to bind it in
	// a let (in the switch at the end of the function).
	letAssignment, tsArg := minimizeLetAssignments([]string{"timestampArg"}, args[1:])
	timestampArg := tsArg[0]

	// handleSimpleCase generates code for cases where we do not need to
	// use $dateFromParts, we just round the interval if the round argument
	// is true, and multiply by the number of milliseconds corresponded to
	// by 'u' then add to the timestamp.
	handleSimpleCase := func(u string, round bool) *ast.Binary {
		if round {
			return ast.NewBinary(bsonutil.OpAdd,
				timestampArg,
				ast.NewBinary(bsonutil.OpMultiply,
					astutil.WrapInRound(interval),
					astutil.FloatValue(toMilliseconds[u]),
				),
			)
		}
		return ast.NewBinary(bsonutil.OpAdd,
			timestampArg,
			astutil.WrapInOp(bsonutil.OpMultiply,
				interval,
				astutil.FloatValue(toMilliseconds[u]),
			),
		)
	}

	// handleDateFromPartsCase handles cases where we need to use
	// $dateFromParts because we want to add a Year, a Month, or 3 Months
	// (a Quarter) to the specific date part.
	handleDateFromPartsCase := func(u string) ast.Expr {
		// Start with the equations for Quarter/Month, since they are
		// the same. They use a shared computation part
		// (sharedComputation) that changes based on if this is a
		// Quarter or Month.
		sharedComputationRef := ast.NewVariableRef("sharedComputation")
		newYearRef, newMonthRef := ast.NewVariableRef("newYear"), ast.NewVariableRef("newMonth")
		dayExpr := ast.NewFunction(bsonutil.OpDayOfMonth, timestampArg)

		thirty := astutil.Int32Value(30)

		// This template is used in a call to $dateFromParts.
		// The Year case modifies part of the template.
		template := ast.NewDocument(
			ast.NewDocumentElement("year", newYearRef),
			ast.NewDocumentElement("month", newMonthRef),
			ast.NewDocumentElement("day",
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
				astutil.WrapInSwitch(
					ast.NewFunction(bsonutil.OpDayOfMonth, timestampArg),
					astutil.WrapInEqCase(newMonthRef, astutil.Int32Value(2),
						astutil.WrapInCond(
							astutil.WrapInOp(bsonutil.OpMin, dayExpr, astutil.Int32Value(29)),
							astutil.WrapInOp(bsonutil.OpMin, dayExpr, astutil.Int32Value(28)),
							astutil.WrapInIsLeapYear(newYearRef),
						),
					),
					astutil.WrapInEqCase(newMonthRef, astutil.Int32Value(4),
						astutil.WrapInOp(bsonutil.OpMin, dayExpr, thirty)),
					astutil.WrapInEqCase(newMonthRef, astutil.Int32Value(6),
						astutil.WrapInOp(bsonutil.OpMin, dayExpr, thirty)),
					astutil.WrapInEqCase(newMonthRef, astutil.Int32Value(9),
						astutil.WrapInOp(bsonutil.OpMin, dayExpr, thirty)),
					astutil.WrapInEqCase(newMonthRef, astutil.Int32Value(11),
						astutil.WrapInOp(bsonutil.OpMin, dayExpr, thirty)),
				)),
			ast.NewDocumentElement("hour", ast.NewFunction(bsonutil.OpHour, timestampArg)),
			ast.NewDocumentElement("minute", ast.NewFunction(bsonutil.OpMinute, timestampArg)),
			ast.NewDocumentElement("second", ast.NewFunction(bsonutil.OpSecond, timestampArg)),
			ast.NewDocumentElement("millisecond", ast.NewFunction(bsonutil.OpMillisecond, timestampArg)),
		)

		var sharedComputationLetAssignment []*ast.LetVariable
		var newYearMonthLetAssignment []*ast.LetVariable
		switch u {
		case Year:
			// For Year intervals, the year, month, and day use
			// different, simpler equations. Keep everything but
			// year, to year we add the rounded interval. There is
			// no SharedComputation part, so we do not ast.NewLet.
			// Note that the rest of the template is maintained.
			template = ast.NewDocument(
				ast.NewDocumentElement("year", ast.NewBinary(bsonutil.OpAdd,
					astutil.WrapInRound(interval),
					ast.NewFunction(bsonutil.OpYear, timestampArg),
				)),
				ast.NewDocumentElement("month", ast.NewFunction(bsonutil.OpMonth, timestampArg)),
				ast.NewDocumentElement("day", ast.NewFunction(bsonutil.OpDayOfMonth, timestampArg)),
				ast.NewDocumentElement("hour", ast.NewFunction(bsonutil.OpHour, timestampArg)),
				ast.NewDocumentElement("minute", ast.NewFunction(bsonutil.OpMinute, timestampArg)),
				ast.NewDocumentElement("second", ast.NewFunction(bsonutil.OpSecond, timestampArg)),
				ast.NewDocumentElement("millisecond", ast.NewFunction(bsonutil.OpMillisecond, timestampArg)),
			)
			return ast.NewFunction(bsonutil.OpDateFromParts, template)

		// For Quarter and Month intervals, only the SharedComputation
		// part changes.
		case Quarter:
			// SharedComputation = Month + round(interval) * 3 - 1.
			sharedComputationLetAssignment = []*ast.LetVariable{
				ast.NewLetVariable("sharedComputation", ast.NewBinary(bsonutil.OpSubtract,
					ast.NewBinary(bsonutil.OpAdd,
						ast.NewFunction(bsonutil.OpMonth, timestampArg),
						ast.NewBinary(bsonutil.OpMultiply,
							astutil.WrapInRound(interval),
							astutil.Int32Value(3),
						),
					),
					astutil.OneInt32Literal,
				)),
			}

		case Month:
			// SharedComputation = Month + round(interval) - 1.
			sharedComputationLetAssignment = []*ast.LetVariable{
				ast.NewLetVariable("sharedComputation", ast.NewBinary(bsonutil.OpSubtract,
					ast.NewBinary(bsonutil.OpAdd,
						ast.NewFunction(bsonutil.OpMonth, timestampArg),
						astutil.WrapInRound(interval),
					),
					astutil.OneInt32Literal,
				)),
			}
		}

		newYearMonthLetAssignment = []*ast.LetVariable{
			// Year = Year + SharedComputation / 12, where / truncates.
			ast.NewLetVariable("newYear", ast.NewBinary(bsonutil.OpAdd,
				ast.NewFunction(bsonutil.OpYear, timestampArg),
				astutil.WrapInIntDiv(sharedComputationRef, astutil.Int32Value(12)),
			)),

			// Month = SharedComputation % 12 + 1.
			ast.NewLetVariable("newMonth", ast.NewBinary(bsonutil.OpAdd,
				astutil.OneInt32Literal,
				ast.NewBinary(bsonutil.OpMod, sharedComputationRef, astutil.Int32Value(12)),
			)),
		}

		// Add lets for Quarter and Month.
		return ast.NewLet(sharedComputationLetAssignment,
			ast.NewLet(newYearMonthLetAssignment, ast.NewFunction(bsonutil.OpDateFromParts, template)),
		)
	}

	// ast.NewLet to bind $$timestampArg.
	switch unit {
	case Year, Month, Quarter:
		return wrapInLet(letAssignment, handleDateFromPartsCase(unit)), nil
	// It is wrong to round for Second, and rounding for Microsecond is
	// just pointless since MongoDB supports only milliseconds, and will
	// automatically round to the nearest millisecond for us.
	case Second, Microsecond:
		return wrapInLet(letAssignment, handleSimpleCase(unit, false)), nil
	default:
		return wrapInLet(letAssignment, handleSimpleCase(unit, true)), nil
	}
}

func (f *baseScalarFunctionExpr) timestampDiffToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	// This is a very large and costly expression, make sure to bind it in
	// a let (in the switch at the end of the function).
	assignments, timestampArgs := minimizeLetAssignments([]string{"timestampArg1", "timestampArg2"}, args)
	timestampArg1, timestampArg2 := timestampArgs[0], timestampArgs[1]

	// handleSimpleCase generates code for cases where we do not need to
	// use and date part access functions (like $dayOfMonth), we just
	// subtract: timestampArg2 - timestampArg1 then divide by the number of
	// milliseconds corresponded to by 'u'.
	handleSimpleCase := func(u string) ast.Expr {
		return astutil.WrapInIntDiv(
			ast.NewBinary(bsonutil.OpSubtract, timestampArg2, timestampArg1),
			astutil.FloatValue(toMilliseconds[u]),
		)
	}

	// handleDatePartsCase handles cases where we need to use
	// date part access functions (like $dayOfMonth).
	handleDatePartsCase := func(u string) ast.Expr {
		year1, year2 := ast.NewVariableRef("year1"), ast.NewVariableRef("year2")
		month1, month2 := ast.NewVariableRef("month1"), ast.NewVariableRef("month2")
		datePartsLetAssignment := []*ast.LetVariable{
			ast.NewLetVariable("year1", ast.NewFunction(bsonutil.OpYear, timestampArg1)),
			ast.NewLetVariable("month1", ast.NewFunction(bsonutil.OpMonth, timestampArg1)),
			ast.NewLetVariable("day1", ast.NewFunction(bsonutil.OpDayOfMonth, timestampArg1)),
			ast.NewLetVariable("hour1", ast.NewFunction(bsonutil.OpHour, timestampArg1)),
			ast.NewLetVariable("minute1", ast.NewFunction(bsonutil.OpMinute, timestampArg1)),
			ast.NewLetVariable("second1", ast.NewFunction(bsonutil.OpSecond, timestampArg1)),
			ast.NewLetVariable("millisecond1", ast.NewFunction(bsonutil.OpMillisecond, timestampArg1)),
			ast.NewLetVariable("year2", ast.NewFunction(bsonutil.OpYear, timestampArg2)),
			ast.NewLetVariable("month2", ast.NewFunction(bsonutil.OpMonth, timestampArg2)),
			ast.NewLetVariable("day2", ast.NewFunction(bsonutil.OpDayOfMonth, timestampArg2)),
			ast.NewLetVariable("hour2", ast.NewFunction(bsonutil.OpHour, timestampArg2)),
			ast.NewLetVariable("minute2", ast.NewFunction(bsonutil.OpMinute, timestampArg2)),
			ast.NewLetVariable("second2", ast.NewFunction(bsonutil.OpSecond, timestampArg2)),
			ast.NewLetVariable("millisecond2", ast.NewFunction(bsonutil.OpMillisecond, timestampArg2)),
		}

		var outputLetAssignment []*ast.LetVariable
		var generateEpsilon func(string, string) ast.Expr
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
			generateEpsilon = func(arg1, arg2 string) ast.Expr {
				return astutil.WrapInCond(
					astutil.OneInt32Literal,
					astutil.ZeroInt32Literal,
					astutil.WrapInOp(bsonutil.OpGt, ast.NewVariableRef("month"+arg1), ast.NewVariableRef("month"+arg2)),
					astutil.WrapInOp(bsonutil.OpGt, ast.NewVariableRef("day"+arg1), ast.NewVariableRef("day"+arg2)),
					astutil.WrapInOp(bsonutil.OpGt, ast.NewVariableRef("hour"+arg1), ast.NewVariableRef("hour"+arg2)),
					astutil.WrapInOp(bsonutil.OpGt, ast.NewVariableRef("minute"+arg1), ast.NewVariableRef("minute"+arg2)),
					astutil.WrapInOp(bsonutil.OpGt, ast.NewVariableRef("second"+arg1), ast.NewVariableRef("second"+arg2)),
					astutil.WrapInOp(bsonutil.OpGt, ast.NewVariableRef("millisecond"+arg1), ast.NewVariableRef("millisecond"+arg2)),
				)
			}
			// output = year2 - year1.
			outputLetAssignment = []*ast.LetVariable{
				ast.NewLetVariable("output", ast.NewBinary(bsonutil.OpSubtract, year2, year1)),
			}

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
			generateEpsilon = func(arg1, arg2 string) ast.Expr {
				return astutil.WrapInCond(
					astutil.OneInt32Literal,
					astutil.ZeroInt32Literal,
					astutil.WrapInOp(bsonutil.OpGt, ast.NewVariableRef("day"+arg1), ast.NewVariableRef("day"+arg2)),
					astutil.WrapInOp(bsonutil.OpGt, ast.NewVariableRef("hour"+arg1), ast.NewVariableRef("hour"+arg2)),
					astutil.WrapInOp(bsonutil.OpGt, ast.NewVariableRef("minute"+arg1), ast.NewVariableRef("minute"+arg2)),
					astutil.WrapInOp(bsonutil.OpGt, ast.NewVariableRef("second"+arg1), ast.NewVariableRef("second"+arg2)),
					astutil.WrapInOp(bsonutil.OpGt, ast.NewVariableRef("millisecond"+arg1), ast.NewVariableRef("millisecond"+arg2)),
				)

			}
			// output = (year2 - year1) * 12 + month2 - month1.
			outputLetAssignment = []*ast.LetVariable{
				ast.NewLetVariable("output", ast.NewBinary(bsonutil.OpAdd,
					astutil.WrapInOp(bsonutil.OpSubtract, month2, month1),
					ast.NewBinary(bsonutil.OpMultiply,
						astutil.Int32Value(12),
						ast.NewBinary(bsonutil.OpSubtract, year2, year1),
					),
				)),
			}
		}

		outputRef := ast.NewVariableRef("output")

		// Generate epsilons and whether we add or subtract said epsilon, which
		// is decided on whether or not "output" is negative or positive.
		ltBranch := ast.NewBinary(bsonutil.OpAdd, outputRef, generateEpsilon("2", "1"))
		gtBranch := ast.NewBinary(bsonutil.OpSubtract, outputRef, generateEpsilon("1", "2"))
		applyEpsilonExpr := ast.NewLet(outputLetAssignment,
			astutil.WrapInSwitch(astutil.ZeroInt32Literal,
				astutil.WrapInCase(ast.NewBinary(bsonutil.OpLt, outputRef, astutil.ZeroInt32Literal), ltBranch),
				astutil.WrapInCase(ast.NewBinary(bsonutil.OpGt, outputRef, astutil.ZeroInt32Literal), gtBranch),
			),
		)

		retExpr := ast.NewLet(datePartsLetAssignment,
			ast.NewLet(outputLetAssignment, applyEpsilonExpr),
		)
		// Quarter is just the number of months integer divided by 3.
		if u == Quarter {
			return astutil.WrapInIntDiv(retExpr, astutil.Int32Value(3))
		}
		return retExpr
	}

	// wrapInLet to bind $$timestampArg1 and 2.
	switch unit {
	case Year, Month, Quarter:
		return wrapInLet(assignments, handleDatePartsCase(unit)), nil
	default:
		return wrapInLet(assignments, handleSimpleCase(unit)), nil
	}
}

func (f *baseScalarFunctionExpr) timestampToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	inputLet, args := minimizeLetAssignments([]string{"val"}, args)
	val := args[0]

	wrapInDateFromString := func(v ast.Expr) *ast.Function {
		return ast.NewFunction(bsonutil.OpDateFromString,
			ast.NewDocument(ast.NewDocumentElement("dateString", v)))
	}

	// CASE 1: it's already a Mongo date, we just return it
	isDateType := containsBSONType(val, "date")
	dateBranch := astutil.WrapInCase(isDateType, val)

	// CASE 2: it's a number.
	isNumber := containsBSONType(val, "int", "decimal", "long", "double")

	// evaluates to true if val positive and has <= X digits.
	hasUpToXDigits := func(x float64) ast.Expr {
		return astutil.WrapInInRange(val, 0, math.Pow(10, x))
	}

	// This handles converting a number in YYMMDDHHMMSS format to YYYYMMDDHHMMSS.
	// if YY < 70, we assume they meant 20YY. if YY > 70, we assume 19YY.
	getPadding := func(v ast.Expr) ast.Expr {
		return astutil.WrapInCond(
			astutil.Int64Value(20000000000000),
			astutil.Int64Value(19000000000000),
			ast.NewBinary(bsonutil.OpLt,
				ast.NewBinary(bsonutil.OpDivide, v, astutil.Int64Value(10000000000)),
				astutil.Int32Value(70)),
		)
	}

	// Constant for the HHMMSS factor to handle dates that do not have HHMMSS.
	hhmmssFactor := astutil.Int32Value(1000000)

	// We interpret this as being format YYMMDD, multiply by hhmmssFactor for HHMMSS then pad.
	ifSix := ast.NewBinary(bsonutil.OpAdd,
		ast.NewBinary(bsonutil.OpMultiply, val, hhmmssFactor),
		getPadding(ast.NewBinary(bsonutil.OpMultiply, val, hhmmssFactor)),
	)
	sixBranch := astutil.WrapInCase(hasUpToXDigits(6), ifSix)

	// This number is YYYYMMDD, again, multiply by hhmmssFactor.
	eightBranch := astutil.WrapInCase(hasUpToXDigits(8), ast.NewBinary(bsonutil.OpMultiply, val, hhmmssFactor))

	// If it's twelve digits, interpret as YYMMDDHHMMSS. Make sure to pad the number.
	ifTwelve := astutil.WrapInOp(bsonutil.OpAdd, val, getPadding(val))
	twelveBranch := astutil.WrapInCase(hasUpToXDigits(12), ifTwelve)

	// if fourteen, YYYYMMDDHHMMSS, we can use as it as is.
	fourteenBranch := astutil.WrapInCase(hasUpToXDigits(14), val)

	// define "num", the input number normalized to 14 digits, in a "let"
	numberVar := astutil.WrapInSwitch(astutil.NullLiteral, sixBranch, eightBranch, twelveBranch, fourteenBranch)
	numberLetVars := []*ast.LetVariable{ast.NewLetVariable("num", numberVar)}
	numRef := ast.NewVariableRef("num")

	oneHundred := astutil.Int32Value(100)

	dateParts := []*ast.LetVariable{
		// YYYYMMDDHHMMSS / 10000000000 = YYYY
		ast.NewLetVariable("year", ast.NewFunction(bsonutil.OpTrunc,
			ast.NewBinary(bsonutil.OpDivide, numRef, astutil.Int64Value(10000000000)),
		)),

		// (YYYYMMDDHHMMSS / 100000000) % 100 = MM
		ast.NewLetVariable("month", ast.NewBinary(bsonutil.OpMod,
			ast.NewFunction(bsonutil.OpTrunc,
				ast.NewBinary(bsonutil.OpDivide, numRef, astutil.Int32Value(100000000)),
			),
			oneHundred,
		)),

		// YYYYMMDDHHMMSS / 1000000) % 100 = DD
		ast.NewLetVariable("day", astutil.WrapInOp(bsonutil.OpMod,
			ast.NewFunction(bsonutil.OpTrunc,
				ast.NewBinary(bsonutil.OpDivide, numRef, astutil.Int32Value(1000000)),
			),
			oneHundred,
		)),

		// YYYYMMDDHHMMSS / 10000) % 100 = HH
		ast.NewLetVariable("hour", astutil.WrapInOp(bsonutil.OpMod,
			ast.NewFunction(bsonutil.OpTrunc,
				ast.NewBinary(bsonutil.OpDivide, numRef, astutil.Int32Value(10000)),
			),
			oneHundred,
		)),

		// YYYYMMDDHHMMSS / 100) % 100 = MM
		ast.NewLetVariable("minute", astutil.WrapInOp(bsonutil.OpMod,
			ast.NewFunction(bsonutil.OpTrunc,
				ast.NewBinary(bsonutil.OpDivide, numRef, oneHundred),
			),
			oneHundred,
		)),

		// YYYYMMDDHHMMSS % 100 = SS
		ast.NewLetVariable("second", ast.NewFunction(bsonutil.OpTrunc,
			astutil.WrapInOp(bsonutil.OpMod, numRef, oneHundred),
		)),

		// YYYYMMDDHHMMSS.FFFFF % 1 * 1000 = ms
		ast.NewLetVariable("millisecond", ast.NewFunction(bsonutil.OpTrunc,
			ast.NewBinary(bsonutil.OpMultiply,
				ast.NewBinary(bsonutil.OpMod, numRef, astutil.OneInt32Literal),
				astutil.Int32Value(1000),
			),
		)),
	}

	yearRef := ast.NewVariableRef("year")
	monthRef := ast.NewVariableRef("month")
	dayRef := ast.NewVariableRef("day")
	hourRef := ast.NewVariableRef("hour")
	minuteRef := ast.NewVariableRef("minute")
	secondRef := ast.NewVariableRef("second")

	// try to avoid aggregation errors by catching obviously invalid dates
	yearValid := astutil.WrapInInRange(yearRef, 0, 10000)
	monthValid := astutil.WrapInInRange(monthRef, 1, 13)
	dayValid := astutil.WrapInInRange(dayRef, 1, 32)
	// Mongo DB actually supports HH=24 which converts to 0, but MySQL does not (it returns NULL)
	// so we stick to MySQL semantics and cap valid hours at 23.
	// Interestingly, $dateFromString does NOT support HH=24.
	hourValid := astutil.WrapInInRange(hourRef, 0, 24)
	minuteValid := astutil.WrapInInRange(minuteRef, 0, 60)
	secondValid := astutil.WrapInInRange(secondRef, 0, 60)

	makeDateOrNull := astutil.WrapInCond(
		ast.NewFunction(bsonutil.OpDateFromParts, ast.NewDocument(
			ast.NewDocumentElement("year", yearRef),
			ast.NewDocumentElement("month", monthRef),
			ast.NewDocumentElement("day", dayRef),
			ast.NewDocumentElement("hour", hourRef),
			ast.NewDocumentElement("minute", minuteRef),
			ast.NewDocumentElement("second", secondRef),
			ast.NewDocumentElement("millisecond", ast.NewVariableRef("millisecond")),
		)),
		astutil.NullLiteral,
		astutil.WrapInOp(bsonutil.OpAnd, yearValid, monthValid, dayValid, hourValid, minuteValid, secondValid),
	)

	evaluateNumber := ast.NewLet(dateParts, makeDateOrNull)
	handleNumberToDate := ast.NewLet(numberLetVars, evaluateNumber)
	numberBranch := astutil.WrapInCase(isNumber, handleNumberToDate)

	// CASE 3: it's a string
	isString := containsBSONType(val, "string")

	// First split on T, take first substring, then split that on " ", and
	// take first substring. this gives us just the date part of the
	// string. note that if the string doesn't have T or a space, just
	// returns original string
	trimmedDateString := astutil.WrapInOp(bsonutil.OpArrElemAt,
		astutil.WrapInOp(bsonutil.OpSplit,
			astutil.WrapInOp(bsonutil.OpArrElemAt,
				astutil.WrapInOp(bsonutil.OpSplit, val, astutil.StringValue("T")),
				astutil.ZeroInt32Literal),
			astutil.StringValue(" ")),
		astutil.ZeroInt32Literal)

	// Repeat the step above but take the second element to get the time
	// part. Replace with "" if we can not find a second element.
	trimmedTimeString := astutil.WrapInIfNull(
		astutil.WrapInOp(bsonutil.OpArrElemAt,
			astutil.WrapInOp(bsonutil.OpSplit, val, astutil.StringValue("T")),
			astutil.OneInt32Literal,
		),
		astutil.WrapInIfNull(
			astutil.WrapInOp(bsonutil.OpArrElemAt,
				astutil.WrapInOp(bsonutil.OpSplit, val, astutil.StringValue(" ")),
				astutil.OneInt32Literal,
			),
			astutil.EmptyStringLiteral,
		),
	)

	// Convert the date and time strings to arrays so we can use
	// map/reduce.
	trimmedDateAsArray := astutil.WrapInStringToArray(ast.NewVariableRef("trimmedDate"))
	trimmedTimeAsArray := astutil.WrapInStringToArray(ast.NewVariableRef("trimmedTime"))

	cRef := ast.NewVariableRef("c")

	// isSeparator evaluates to true if a character is in the defined
	// separator list
	isSeparator := ast.NewBinary(bsonutil.OpNeq,
		astutil.Int32Value(-1),
		astutil.WrapInOp("$indexOfArray", astutil.DateComponentSeparator, cRef),
	)

	// Use map to convert all separators in the date string to - symbol,
	// and leave numbers as-is
	dateNormalized := astutil.WrapInMap(
		trimmedDateAsArray,
		"c",
		astutil.WrapInCond(astutil.StringValue("-"), cRef, isSeparator),
	)
	// Use map to convert all separators in the time string to '.' symbol,
	// and leave numbers as-is. We use '.' instead of ':' so that MongoDB
	// correctly handles fractional seconds. 10.11.23.1234 is parsed
	// correctly as 10:11:23.1234, saving us some effort (and runtime).
	timeNormalized := astutil.WrapInMap(
		trimmedTimeAsArray,
		"c",
		astutil.WrapInCond(astutil.StringValue("."), cRef, isSeparator),
	)

	// Use reduce to convert characters back to a single string for date and time.
	dateJoined := astutil.WrapInReduce(
		dateNormalized,
		astutil.EmptyStringLiteral,
		astutil.WrapInOp(bsonutil.OpConcat, astutil.ValueVarRef, astutil.ThisVarRef),
	)
	timeJoined := astutil.WrapInReduce(
		timeNormalized,
		astutil.EmptyStringLiteral,
		astutil.WrapInOp(bsonutil.OpConcat, astutil.ValueVarRef, astutil.ThisVarRef),
	)

	dateJoinedRef := ast.NewVariableRef("dateJoined")

	// if the third character is a -, or if the string is only 6 digits
	// long and has no slashes, then the string is either format YY/MM/DD
	// or YYMMDD and we need to add the appropriate first two year digits
	// (19xx or 20xx) for Mongo to understand it
	hasShortYear := ast.NewBinary(bsonutil.OpOr,
		// length is only 6, assume YYMMDD
		ast.NewBinary(bsonutil.OpEq,
			ast.NewFunction(bsonutil.OpStrlenCP, dateJoinedRef),
			astutil.Int32Value(6),
		),
		// third character is -, assume YY-MM-DD
		astutil.WrapInOp(bsonutil.OpEq,
			astutil.StringValue("-"),
			astutil.WrapInOp(bsonutil.OpSubstr, dateJoinedRef, astutil.Int32Value(2), astutil.OneInt32Literal),
		),
	)

	// "$dateFromString" actually pads correctly, but not if "/" is
	// used as the separator (it will assume year is last). If this
	// pushdown is shown to be slow by benchmarks, we should reconsider
	// allowing "$dateFromString" to handle padding. The change
	// would not be trivial due to how MongoDB cannot handle short dates
	// when there are no separators in the date.
	padYear := astutil.WrapInOp(bsonutil.OpConcat,
		astutil.WrapInCond(
			astutil.StringValue("20"),
			astutil.StringValue("19"),
			// check if first two digits < 70 to determine padding
			ast.NewBinary(bsonutil.OpLt,
				astutil.WrapInOp(bsonutil.OpSubstr, dateJoinedRef, astutil.ZeroInt32Literal, astutil.Int32Value(2)),
				astutil.StringValue("70"),
			),
		),
		dateJoinedRef,
	)

	// We have to use nested $lets because in the outer one we define
	// $$trimmedDate and in the inner one we define $$dateJoined. Defining
	// $$dateJoined requires knowing the length of trimmedDate, so we can't
	// do it all in one step.
	innerIn := astutil.WrapInCond(padYear, dateJoinedRef, hasShortYear)
	innerLet := ast.NewLet([]*ast.LetVariable{ast.NewLetVariable("dateJoined", dateJoined)}, innerIn)

	// Concat the time back into the date.
	concatedDate := astutil.WrapInOp(bsonutil.OpConcat, innerLet, timeJoined)

	// gracefully handle strings that are too short to possibly be valid by returning null
	tooShort := ast.NewBinary(bsonutil.OpLt,
		ast.NewFunction(bsonutil.OpStrlenCP, ast.NewVariableRef("trimmedDate")),
		astutil.Int32Value(6),
	)
	outerIn := astutil.WrapInCond(astutil.NullLiteral, wrapInDateFromString(concatedDate), tooShort)
	outerLet := ast.NewLet([]*ast.LetVariable{
		ast.NewLetVariable("trimmedDate", trimmedDateString),
		ast.NewLetVariable("trimmedTime", trimmedTimeString),
	}, outerIn)

	// Make sure if we get the int 0 we return NULL instead
	// of crashing. MySQL uses '0000-00-00' as an error output for some
	// functions and we encode it as the integer 0 within push down.
	stringBranch := astutil.WrapInCase(isString,
		astutil.WrapInCond(
			astutil.NullLiteral,
			outerLet,
			ast.NewBinary(bsonutil.OpEq, astutil.ZeroInt32Literal, val),
		),
	)

	return wrapInLet(inputLet, astutil.WrapInSwitch(astutil.NullLiteral, dateBranch, numberBranch, stringBranch)), nil
}

func (f *baseScalarFunctionExpr) toSecondsToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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
	return ast.NewBinary(bsonutil.OpMultiply,
		astutil.FloatValue(1e-3),
		ast.NewBinary(bsonutil.OpSubtract, args[0], astutil.DateConstant(dayOne)),
	), nil
}
func (f *baseScalarFunctionExpr) trimToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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
		return ast.NewFunction(bsonutil.OpTrim, ast.NewDocument(
			ast.NewDocumentElement("input", args[0]),
			ast.NewDocumentElement("chars", astutil.StringValue(" ")),
		)), nil
	}

	rtrimCond := astutil.WrapInCond(
		astutil.EmptyStringLiteral,
		astutil.WrapInLRTrim(false, args[0]),
		ast.NewBinary(bsonutil.OpEq, args[0], astutil.EmptyStringLiteral),
	)

	rtrimRef := ast.NewVariableRef("rtrim")

	ltrimCond := astutil.WrapInCond(
		astutil.EmptyStringLiteral,
		astutil.WrapInLRTrim(true, rtrimRef),
		ast.NewBinary(bsonutil.OpEq, rtrimRef, astutil.EmptyStringLiteral),
	)

	trimCond := ast.NewLet([]*ast.LetVariable{ast.NewLetVariable("rtrim", rtrimCond)}, ltrimCond)

	trim := astutil.WrapInCond(
		astutil.EmptyStringLiteral,
		trimCond,
		ast.NewBinary(bsonutil.OpEq, args[0], astutil.EmptyStringLiteral),
	)

	return wrapInNullCheckedCond(t.ColumnsToNullCheck(), astutil.NullLiteral, trim, args...), nil
}

func (f *baseScalarFunctionExpr) truncateToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(truncate)",
			incorrectArgCountMsg,
		)
	}
	dValueExpr, ok := exprs[1].(SQLValueExpr)
	if !ok {
		return nil, newPushdownFailure("SQLScalarFunctionExpr(truncate)", "second arg is not a literal")
	}

	d := values.Float64(dValueExpr.Value)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	if d >= 0 {
		pow := astutil.FloatValue(math.Pow(10, d))
		return ast.NewBinary(bsonutil.OpDivide,
			astutil.WrapInCond(
				ast.NewFunction(bsonutil.OpFloor, ast.NewBinary(bsonutil.OpMultiply, args[0], pow)),
				ast.NewFunction(bsonutil.OpCeil, ast.NewBinary(bsonutil.OpMultiply, args[0], pow)),
				ast.NewBinary(bsonutil.OpGte, args[0], astutil.ZeroInt32Literal),
			),
			pow,
		), nil
	}

	pow := astutil.FloatValue(math.Pow(10, math.Abs(d)))
	return ast.NewBinary(bsonutil.OpMultiply,
		pow,
		astutil.WrapInCond(
			ast.NewFunction(bsonutil.OpFloor, ast.NewBinary(bsonutil.OpDivide, args[0], pow)),
			ast.NewFunction(bsonutil.OpCeil, ast.NewBinary(bsonutil.OpDivide, args[0], pow)),
			ast.NewBinary(bsonutil.OpGte, args[0], astutil.ZeroInt32Literal),
		),
	), nil
}

func (f *baseScalarFunctionExpr) ucaseToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(ucase)",
			incorrectArgCountMsg,
		)
	}

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpToUpper, exprs[0], t)
}

func (f *baseScalarFunctionExpr) unixTimestampToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	now := time.Now()

	if len(exprs) != 1 {
		return astutil.Int64Value(now.Unix()), nil
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

	diffRef := ast.NewVariableRef("diff")
	letAssignment := []*ast.LetVariable{
		ast.NewLetVariable("diff", ast.NewFunction(bsonutil.OpTrunc,
			ast.NewBinary(bsonutil.OpDivide,
				ast.NewBinary(bsonutil.OpSubtract,
					ast.NewBinary(bsonutil.OpSubtract, arg, astutil.DateConstant(epoch)),
					astutil.Int64Value(int64(tzCompensation*1000)),
				),
				astutil.Int32Value(1000),
			),
		)),
	}

	letEvaluation := astutil.WrapInCond(
		diffRef,
		astutil.FloatValue(0.0),
		astutil.WrapInOp(bsonutil.OpGt, diffRef, astutil.ZeroInt32Literal),
	)

	return ast.NewLet(letAssignment, letEvaluation), nil
}

// nolint: unparam
func (f *baseScalarFunctionExpr) utcDateToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	now := time.Now().In(time.UTC)
	cUTCd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return astutil.DateConstant(cUTCd), nil
}

// nolint: unparam
func (f *baseScalarFunctionExpr) utcTimestampToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	return astutil.DateConstant(time.Now().In(time.UTC)), nil
}

func (f *baseScalarFunctionExpr) weekToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) < 1 || len(exprs) > 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(week)",
			incorrectArgCountMsg,
		)
	}
	mode := int64(0)
	if len(exprs) == 2 {
		modeValueExpr, ok := exprs[1].(SQLValueExpr)
		if !ok {
			return nil, newPushdownFailure("SQLScalarFunctionExpr(week)", "mode is not a literal")
		}
		mode = values.Int64(modeValueExpr.Value)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}
	return astutil.WrapInWeekCalculation(args[0], mode), nil
}

func (f *baseScalarFunctionExpr) weekdayToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
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

	assignments, args := minimizeLetAssignments([]string{"date"}, args)

	seven := astutil.Int32Value(7)

	evaluation := wrapInNullCheckedCond(
		t.ColumnsToNullCheck(),
		astutil.NullLiteral,
		ast.NewBinary(bsonutil.OpMod,
			ast.NewBinary(bsonutil.OpAdd,
				ast.NewBinary(bsonutil.OpMod,
					ast.NewBinary(bsonutil.OpSubtract,
						ast.NewFunction(bsonutil.OpDayOfWeek, args[0]),
						astutil.Int32Value(2),
					),
					seven,
				),
				seven,
			),
			seven,
		),
		args...,
	)

	return wrapInLet(assignments, evaluation), nil

}

func (f *baseScalarFunctionExpr) yearToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(year)",
			incorrectArgCountMsg,
		)
	}

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpYear, exprs[0], t)
}

func (f *baseScalarFunctionExpr) yearWeekToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if len(exprs) < 1 || len(exprs) > 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(yearWeek)",
			incorrectArgCountMsg,
		)
	}
	mode := int64(0)
	if len(exprs) == 2 {
		modeValueExpr, ok := exprs[1].(SQLValueExpr)
		if !ok {
			return nil, newPushdownFailure("SQLScalarFunctionExpr(yearWeek)", "mode is not a literal")
		}
		mode = values.Int64(modeValueExpr.Value)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	inputAssignment, args := minimizeLetAssignments([]string{"date"}, args[0:1])
	date := args[0]

	month, year := ast.NewVariableRef("month"), ast.NewVariableRef("year")
	monthAssignment := []*ast.LetVariable{
		ast.NewLetVariable("month", ast.NewFunction(bsonutil.OpMonth, date)),
		ast.NewLetVariable("year", ast.NewFunction(bsonutil.OpYear, date)),
	}

	var weekCalc ast.Expr

	// Unlike WEEK, YEARWEEK always uses the 1-53 modes. Thus
	// we always call week with the 1-53 of a 0-53, 1-53 pair.
	switch mode {

	// First day of week: Sunday, with a Sunday in this year.
	case 0, 2:
		weekCalc = astutil.WrapInWeekCalculation(date, 2)
	// First day of weekCalc: Monday, with 4 days in this year.
	case 1, 3:
		weekCalc = astutil.WrapInWeekCalculation(date, 3)
	// First day of weekCalc: Sunday, with 4 days in this year.
	case 4, 6:
		weekCalc = astutil.WrapInWeekCalculation(date, 6)
	// First day of weekCalc: Monday, with a Monday in this year.
	case 5, 7:
		weekCalc = astutil.WrapInWeekCalculation(date, 7)
	}

	week := ast.NewVariableRef("week")
	weekAssignment := []*ast.LetVariable{
		ast.NewLetVariable("week", weekCalc),
	}

	newYearAssignment := []*ast.LetVariable{
		ast.NewLetVariable("newYear", astutil.WrapInSwitch(
			year,
			astutil.WrapInEqCase(week, astutil.OneInt32Literal,
				astutil.WrapInCond(
					ast.NewBinary(bsonutil.OpAdd, year, astutil.OneInt32Literal),
					year,
					ast.NewBinary(bsonutil.OpEq, month, astutil.Int32Value(12)),
				),
			),
			astutil.WrapInEqCase(week, astutil.Int32Value(52),
				astutil.WrapInCond(
					ast.NewBinary(bsonutil.OpSubtract, year, astutil.OneInt32Literal),
					year,
					ast.NewBinary(bsonutil.OpEq, month, astutil.OneInt32Literal),
				),
			),
			astutil.WrapInEqCase(week, astutil.Int32Value(53),
				astutil.WrapInCond(
					ast.NewBinary(bsonutil.OpSubtract, year, astutil.OneInt32Literal),
					year,
					ast.NewBinary(bsonutil.OpEq, month, astutil.OneInt32Literal),
				),
			),
		)),
	}

	return wrapInLet(inputAssignment,
		ast.NewLet(monthAssignment,
			ast.NewLet(weekAssignment,
				ast.NewLet(newYearAssignment,
					ast.NewBinary(bsonutil.OpAdd,
						ast.NewBinary(bsonutil.OpMultiply,
							ast.NewVariableRef("newYear"),
							astutil.Int32Value(100),
						),
						week,
					),
				),
			),
		),
	), nil
}

func (t *PushdownTranslator) translateArgs(exprs []SQLExpr) ([]ast.Expr, PushdownFailure) {
	args := make([]ast.Expr, len(exprs))
	for i, e := range exprs {
		r, err := t.ToAggregationLanguage(e)
		if err != nil {
			return nil, err
		}
		args[i] = r
	}
	return args, nil
}

// minimizeLetAssignments iterates through a slice of arguments to be assigned in
// a $let and "minimizes" them according to the following rules:
//   - literals and columns are not assigned because their values can be used directly.
//   - other values are assigned with their corresponding varName and their
//   argument value is replaced with the assignment reference.
//
// For example,
//   minimizeLetAssignments(
//     []string{"left", "right"},
//     []ast.Expr{*ast.FieldRef{Name: "a"}, *ast.Document{Elements: ...}}
//   )
// returns the following:
//   (
//     []bson.DocElem{bson.DocElem{"right", ...}},
//     [][]ast.Expr{*ast.FieldRef{Name: "a"}, *ast.VariableRef{Name: "right"}}
//   )
// In this example, "$a" (a column) is not assigned to a variable since "$a" can be
// used directly wherever that variable would have been used. The same is true for
// literals.
//
// minimizeLetAssignments returns the slice of args because, as shown above, the
// arguments can be manipulated.
func minimizeLetAssignments(varNames []string, args []ast.Expr) ([]*ast.LetVariable, []ast.Expr) {
	assignments := make([]*ast.LetVariable, 0, len(args))
	newArgs := make([]ast.Expr, len(args))
	for i, arg := range args {
		switch arg.(type) {
		case *ast.Constant, *ast.FieldRef:
			// do nothing (this is the actual "minimization").
			//   - literals and columns are not assigned because they can
			//     be used directly.
			newArgs[i] = arg

		default:
			// store anything that is not a literal or column in a let assignment.
			assignments = append(assignments, ast.NewLetVariable(varNames[i], arg))

			// update the argument to be the variable reference.
			newArgs[i] = ast.NewVariableRef(varNames[i])
		}
	}

	return assignments, newArgs
}

// minimizeNullChecks iterates through a slice of arguments to null-check and
// "minimizes" them according to the following rules:
//   - literals are not included in the returned slice of conditions.
//       - if a null literal is encountered, this function returns an
//       empty slice and true.
//   - columns are not null-checked in-line, but instead are added to
//   the map of columnsToNullCheck and a reference to the null-check
//   is used.
//   - other values are wrapped in a null-check.
//
// For example,
//   minimizeNullChecks(
//     ...,
//     []ast.Expr{*ast.FieldRef{Name: "a"}, *ast.Document{Elements: ...}}
//   )
// returns the following:
//   ([]ast.Expr{
//   	*ast.VariableRef{Name: "a_is_null"},
//   	ast.Binary{Op: "lte", left: ..., right: *ast.Constant(null)}
//   }, false)
//
// minimizeNullChecks returns a slice of conditions (for use for WrapInCond)
// and a bool indicating whether or not a literal null value was found.
func minimizeNullChecks(columnsToNullCheck map[string]struct{}, argsToNullCheck ...ast.Expr) ([]ast.Expr, bool) {
	minimizedConds := make([]ast.Expr, 0, len(argsToNullCheck))
	for _, arg := range argsToNullCheck {
		switch a := arg.(type) {
		case *ast.Constant:
			// literals do not need to be included in the slice of
			// conditions. Their values can be inspected here in Go:
			if a.Value.Type == bsontype.Null {
				// if an argument to null-check is a null literal, return
				// true to indicate the provided conditions contained a
				// literal null value.
				return []ast.Expr{}, true
			}

		case *ast.FieldRef:
			// columns that need to be null-checked are added to the map of
			// columnsToNullCheck and the null-checked variable ref is
			// included in the slice of conditions.
			if astutil.AllParentsAreFieldRefs(a) {
				columnName := astutil.FieldRefString(a)
				columnsToNullCheck[columnName] = struct{}{}
				minimizedConds = append(minimizedConds, toNullCheckedLetVarRef(columnName))
			} else {
				// for cases where a parent is an ast.VariableRef (which should not be
				// null-checked at the top level).
				minimizedConds = append(minimizedConds, astutil.WrapInNullCheck(a))
			}

		default:
			// anything that is not a literal or column needs to be wrapped in
			// a null-check and that is included in the slice of conditions.
			minimizedConds = append(minimizedConds, astutil.WrapInNullCheck(arg))
		}
	}

	return minimizedConds, false
}

// wrapSingleArgFuncWithNullCheck returns a null checked version of the
// argued function (operator and argument).
// If the expr is a column, this function will add it to the PushdownTranslator
// and will use the null-checked variable binding as the condition.
func wrapSingleArgFuncWithNullCheck(op string, expr SQLExpr, t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	arg, err := t.ToAggregationLanguage(expr)
	if err != nil {
		return nil, err
	}

	switch a := arg.(type) {
	case *ast.FieldRef:
		var nullCheck ast.Expr
		if astutil.AllParentsAreFieldRefs(a) {
			columnName := astutil.FieldRefString(a)
			t.ColumnsToNullCheck()[columnName] = struct{}{}
			nullCheck = toNullCheckedLetVarRef(columnName)
		} else {
			// for cases where a parent is an ast.VariableRef (which should not be
			// null-checked at the top level).
			nullCheck = astutil.WrapInNullCheck(a)
		}

		return astutil.WrapInCond(
			astutil.NullLiteral,
			ast.NewFunction(op, arg),
			nullCheck,
		), nil
	}

	return astutil.WrapSingleArgFuncWithNullCheck(op, arg), nil
}

func wrapInNullCheckedCond(
	columnsToNullCheck map[string]struct{},
	truePart, falsePart ast.Expr,
	argsToNullCheck ...ast.Expr,
) ast.Expr {
	conds, containsNullLiteral := minimizeNullChecks(columnsToNullCheck, argsToNullCheck...)
	if containsNullLiteral {
		return truePart
	}

	return astutil.WrapInCond(truePart, falsePart, conds...)
}

// wrapInLet returns a $let if there are assignments, and the evaluation otherwise
func wrapInLet(assignments []*ast.LetVariable, evaluation ast.Expr) ast.Expr {
	if len(assignments) == 0 {
		return evaluation
	}

	return ast.NewLet(assignments, evaluation)
}
