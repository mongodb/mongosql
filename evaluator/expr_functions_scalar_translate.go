package evaluator

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/10gen/mongoast/ast"
	"go.mongodb.org/mongo-driver/bson/bsontype"

	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/dateutil"
	"github.com/10gen/sqlproxy/internal/mathutil"
	"github.com/10gen/sqlproxy/schema"
)

// FuncToAggregation for TO_DAYS has one issue wrt how TO_DAYS is supposed to perform:
// because our date treatment is backed by using MongoDB's $dateFromString function,
// if a date that doesn't exist (e.g., 0000-00-00 or 0001-02-29) is entered, we return
// an error instead of the NULL expected from MySQL. Unfortunately, checking for valid
// dates is too cost prohibitive. If at some point $dateFromString supports an onError/default
// value, we should switch to using that.
func (f *baseScalarFunctionExpr) toDaysToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

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
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return ast.NewFunction(bsonutil.OpAbs, args[0]), nil
}

func (f *baseScalarFunctionExpr) acosToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	input := ast.NewVariableRef("input")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("input", args[0]),
	}

	var acosExpr ast.Expr
	if t.versionAtLeast(4, 1, 7) {
		acosExpr = ast.NewFunction(bsonutil.OpAcos, input)
	} else {
		acosExpr = astutil.WrapInAcosComputation(input)
	}

	// MySQL returns NULL for values outside of the range [-1,1].
	// asin x + acos x = pi/2
	return ast.NewLet(assignments,
		astutil.WrapInCond(
			astutil.NullLiteral,
			acosExpr,
			ast.NewBinary(bsonutil.OpLt, input, astutil.FloatValue(-1.0)),
			ast.NewBinary(bsonutil.OpGt, input, astutil.FloatValue(1.0)),
		),
	), nil
}

func (f *baseScalarFunctionExpr) asciiToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	if !t.versionAtLeast(4, 0, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(ascii)",
			"cannot push down to MongoDB < 4.0",
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	input := ast.NewVariableRef("input")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("input", args[0]),
	}

	firstChar := ast.NewVariableRef("firstChar")
	firstCharAssignment := []*ast.LetVariable{
		ast.NewLetVariable("firstChar",
			astutil.WrapInNullCheckedCond(astutil.NullLiteral,
				astutil.WrapInOp(bsonutil.OpSubstr,
					input, astutil.ZeroInt32Literal, astutil.OneInt32Literal),
				input)),
	}

	branches := make([]ast.Expr, 185)

	// https://en.wikipedia.org/wiki/UTF-8
	// https://www.smashingmagazine.com/2012/06/all-about-unicode-utf8-character-sets/
	// MySQL follows UTF-8 and returns the first byte of a character's UTF-8 encoding.
	// UTF-8 encodes characters using one to four 8-bit bytes using these formats.
	// 0xxxxxxx
	// 110xxxxx 10xxxxxx
	// 1110xxxx 10xxxxxx 10xxxxxx
	// 11110xxx 10xxxxxx 10xxxxxx 10xxxxxx
	// The first byte indicates how many bytes will be used in the encoding.
	// 0-127 indicates just one byte will be used, and corresponds directly to the 128 ascii characters.
	// 192-223 means 2 bytes will be used, 224-239 means 3 bytes, and 240-247 means 4 bytes.
	// The 2nd-4th bytes can range in value from 128-191.
	// Ex: 226 followed by 190 and then 128 is ⾀, and using MySQL, ascii(⾀) gives 226

	// For the 0-127 ascii characters, they can be encoded in just one byte
	// so the ascii function will return their actual ascii value.
	// This checks string equality for the 0-127 ascii characters.
	for i := 0; i <= 127; i++ {
		int64i := int64(i)
		if i == 0 {
			// special case for both the empty string and the "\0" string
			branches[i] = astutil.WrapInCase(
				ast.NewBinary(bsonutil.OpOr,
					ast.NewBinary(bsonutil.OpEq, firstChar, astutil.EmptyStringLiteral),
					ast.NewBinary(bsonutil.OpEq, firstChar, astutil.StringValue(string(int64i)))),
				astutil.Int64Value(int64i))
		} else {
			branches[i] = astutil.WrapInEqCase(
				firstChar, astutil.StringValue(string(int64i)), astutil.Int64Value(int64i))
		}
	}

	// For the UTF-encoding of all other characters, only their first byte is returned by the ascii function,
	// so to figure out what the first byte is, we only need to check if the character is between
	// the lowest and highest possible encodings for a given first byte.
	for i := 192; i <= 247; i++ {
		int64i := int64(i)
		bytei := byte(i)
		var lowStr, highStr string
		if i <= 223 {
			// 192 - 223 means 2 bytes will be used
			lowStr = string([]byte{bytei, 128})
			highStr = string([]byte{bytei, 191})
		} else if i <= 239 {
			// 224 - 239 means 3 bytes will be used
			lowStr = string([]byte{bytei, 128, 128})
			highStr = string([]byte{bytei, 191, 191})
		} else {
			// 240 - 247 means 4 bytes will be used
			lowStr = string([]byte{bytei, 128, 128, 128})
			highStr = string([]byte{bytei, 191, 191, 191})
		}

		branches[i-64] = astutil.WrapInCase(
			ast.NewBinary(bsonutil.OpAnd,
				ast.NewBinary(bsonutil.OpGte, firstChar, astutil.StringValue(lowStr)),
				ast.NewBinary(bsonutil.OpLte, firstChar, astutil.StringValue(highStr))),
			astutil.Int64Value(int64i))
	}

	// special case for null
	branches[184] = astutil.WrapInEqCase(firstChar, astutil.NullLiteral, astutil.NullLiteral)

	return ast.NewLet(assignments, ast.NewLet(firstCharAssignment,
		ast.NewFunction(bsonutil.OpSwitch, ast.NewDocument(
			ast.NewDocumentElement("branches", ast.NewArray(branches...)),
		)))), nil
}

func (f *baseScalarFunctionExpr) asinToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	input := ast.NewVariableRef("input")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("input", args[0]),
	}

	var asinExpr ast.Expr
	if t.versionAtLeast(4, 1, 7) {
		asinExpr = ast.NewFunction(bsonutil.OpAsin, input)
	} else {
		asinExpr = ast.NewBinary(bsonutil.OpSubtract,
			astutil.PiOverTwoLiteral,
			astutil.WrapInAcosComputation(input),
		)
	}

	// MySQL returns NULL for values outside of the range [-1,1].
	// asin(x) =  pi/2 - cos(x) via the identity:
	// asin(x) + acos(x) = pi/2.
	return ast.NewLet(assignments,
		astutil.WrapInCond(
			astutil.NullLiteral,
			asinExpr,
			ast.NewBinary(bsonutil.OpLt, input, astutil.FloatValue(-1.0)),
			ast.NewBinary(bsonutil.OpGt, input, astutil.FloatValue(1.0)),
		),
	), nil
}

func (f *baseScalarFunctionExpr) atanToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(4, 1, 7) {
		return nil, newPushdownFailure(f.ExprName(), "no pushdown implementation")
	}

	if len(exprs) == 1 || len(exprs) == 2 {
		args, err := t.translateArgs(exprs)
		if err != nil {
			return nil, err
		}
		if len(exprs) == 1 {
			return astutil.WrapInOp(bsonutil.OpAtan, args[0]), nil
		}
		if len(exprs) == 2 {
			return astutil.WrapInOp(bsonutil.OpAtan2, args[0], args[1]), nil
		}
	}

	// not using assertEitherArgCount(args []SQLExpr, expectedCountOne, expectedCountTwo int)
	// since function needs either a return value or a clear invocation of panic.
	panic(fmt.Sprintf("need either 1 or 2 args, found %d", len(exprs)))

}

func (f *baseScalarFunctionExpr) atan2ToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	return f.atanToAggregationLanguage(t, exprs)
}

func (f *baseScalarFunctionExpr) ceilToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return ast.NewFunction(bsonutil.OpCeil, args[0]), nil
}

func (f *baseScalarFunctionExpr) charToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(char)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	// char is not pushed down for more than 1 arg, because there is currently no way
	// to convert multi-byte characters like ⾀ (which is 226, 190, 128).
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(char)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	// The char function takes the input number and converts it into a byte array
	// by converting the number to base 256.
	// The byte array is then converted into a string array,
	// which is then concatenated together into a string.

	// The following is all special casing for decimal, non-positive or null inputs
	valRef := ast.NewVariableRef("val")
	valLetAssignment := []*ast.LetVariable{
		ast.NewLetVariable("val", args[0]),
	}

	// If the argument is null, change the step input to $range to be -1
	//in order to create an empty byte array and return an empty string at the end.
	step := astutil.WrapInNullCheckedCond(astutil.Int32Value(-1), astutil.OneInt32Literal, args[0])

	// if val < 0, val = val % 256
	// This matches the behavior of in-memory evaluation of char for negative numbers,
	// but not the behavior of MySQL's char function for negative numbers.
	positiveVal := ast.NewLet(valLetAssignment, astutil.WrapInCond(
		astutil.WrapInMod(valRef, astutil.Int32Value(256)), valRef,
		astutil.WrapInBinOp(bsonutil.OpLt, valRef, astutil.ZeroInt32Literal)))

	positiveValRef := ast.NewVariableRef("positiveVal")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("positiveVal", positiveVal),
	}

	// This converts val into a byte array representing val base 256.

	// First, create a placeholder for each digit in val base 256
	// by creating an array from 0 to (int(log_256(val)) + 1).
	// These will be each digit of val base 256 (the ones digit, the 256 digit, the 256^2 digit, etc.)
	// If the value is 0 or null, set the upper bound of the array to be 0
	// because $log and $range don't take zero/null inputs respectively.
	upperLimit := astutil.WrapInNullCheckedCond(astutil.ZeroInt32Literal,
		astutil.WrapInCond(astutil.ZeroInt32Literal,
			astutil.WrapInOp(bsonutil.OpAdd,
				astutil.WrapInIntDiv(ast.NewBinary(bsonutil.OpLog, positiveValRef, astutil.Int32Value(256)),
					astutil.OneInt32Literal), astutil.OneInt32Literal),
			astutil.WrapInBinOp(bsonutil.OpEq, positiveValRef, astutil.ZeroInt32Literal)),
		positiveValRef)

	powers := astutil.WrapInRange(astutil.ZeroInt32Literal, upperLimit, step)

	// Calculate the value of each digit in val base 256
	// by calculating (val / 256 ^ (array_element)) % 256 on each element of the array.
	valRemaindersFunc := astutil.WrapInRemainder(
		astutil.WrapInIntDiv(positiveValRef,
			ast.NewBinary(bsonutil.OpPow, astutil.Int32Value(256), astutil.ThisVarRef)),
		astutil.Int32Value(256))

	valBase256 := astutil.WrapInMap(powers, "this", valRemaindersFunc)

	// converting byte array to string array
	branches := make([]ast.Expr, 129)
	for i := 0; i <= 127; i++ {
		var bytei []byte
		bytei = append(bytei, byte(i))
		branches[i] = astutil.WrapInEqCase(
			astutil.ThisVarRef, astutil.Int32Value(int32(i)), astutil.StringValue(string(bytei)))
	}

	// If the number is between 128 and 255, output the literal '�' character instead of the string version
	// of the number, because the string version of the number is an invalid UTF-8 character.
	// Because of this, this function doesn't roundtrip in all cases:
	// ascii(char(128)) will give 239 instead of 128 (because the first byte of '�' is 239).
	// This does not match the in-memory evaluation of the char function, which correctly
	// returns the string version of the number.
	branches[128] = astutil.WrapInCase(
		ast.NewBinary(bsonutil.OpAnd,
			ast.NewBinary(bsonutil.OpGte, astutil.ThisVarRef, astutil.Int32Value(128)),
			ast.NewBinary(bsonutil.OpLte, astutil.ThisVarRef, astutil.Int32Value(255))),
		astutil.StringValue("�"))

	branchFunc := ast.NewFunction(bsonutil.OpSwitch, ast.NewDocument(
		ast.NewDocumentElement("branches", ast.NewArray(branches...))))
	convertedStringArr := astutil.WrapInMap(valBase256, "this", branchFunc)

	// concatenate string array together from right to left to get correct order of characters
	concatenatedString := ast.NewLet(assignments,
		astutil.WrapInReduce(convertedStringArr, astutil.StringValue(""),
			astutil.WrapInOp(bsonutil.OpConcat, astutil.ThisVarRef, astutil.ValueVarRef)))

	return concatenatedString, nil
}

func (f *baseScalarFunctionExpr) characterLengthToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(characterLength)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	assertExactArgCount(exprs, 1)

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpStrlenCP, exprs[0], t)
}

func (f *baseScalarFunctionExpr) concatToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertAtLeastArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return astutil.WrapInConcat(args), nil
}

func (f *baseScalarFunctionExpr) concatWsToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertAtLeastArgCount(exprs, 2)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	sep := args[0]
	args = args[1:]

	pushArgs := make([]ast.Expr, len(args)*2)

	i := 0
	for _, arg := range args {
		pushArgs[i] = astutil.WrapInNullCheckedCond(astutil.EmptyStringLiteral, arg, arg)
		i++

		pushArgs[i] = astutil.WrapInNullCheckedCond(astutil.EmptyStringLiteral, sep, arg)
		i++
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
	return astutil.WrapInNullCheckedCond(astutil.NullLiteral, ast.NewLet(normalizedVars,
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
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	if t.versionAtLeast(4, 1, 7) {
		return ast.NewFunction(bsonutil.OpCos, args[0]), nil
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

func (f *baseScalarFunctionExpr) dateAddToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	return f.dateArithmeticToAggregationLanguage(t, exprs, false)
}

func (f *baseScalarFunctionExpr) dateArithmeticToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr, isSub bool) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 3)

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
		// For a year, we just add n to the year.
	} else if unitValue.String() == Year {
		if !t.versionAtLeast(4, 0, 0) {
			return nil, newPushdownFailure(
				"SQLScalarFunctionExpr(dateArithmetic)",
				"cannot push down to MongoDB < 4.0 (need overflowing $dateFromParts)",
			)
		}
		assignments := []*ast.LetVariable{
			ast.NewLetVariable(
				"dateParts",
				ast.NewFunction(
					bsonutil.OpDateToParts,
					ast.NewDocument(
						ast.NewDocumentElement("date", date),
					),
				),
			),
		}
		expr := ast.NewFunction(
			bsonutil.OpDateFromParts,
			ast.NewDocument(
				ast.NewDocumentElement(
					"year",
					astutil.WrapInBinOp(
						bsonutil.OpAdd,
						ast.NewVariableRef("dateParts.year"),
						astutil.Int64Value(intervalValue.SQLInt().Value().(int64)),
					),
				),
				ast.NewDocumentElement("month", ast.NewVariableRef("dateParts.month")),
				ast.NewDocumentElement("day", ast.NewVariableRef("dateParts.day")),
				ast.NewDocumentElement("hour", ast.NewVariableRef("dateParts.hour")),
				ast.NewDocumentElement("minute", ast.NewVariableRef("dateParts.minute")),
				ast.NewDocumentElement("second", ast.NewVariableRef("dateParts.second")),
				ast.NewDocumentElement("millisecond", ast.NewVariableRef("dateParts.millisecond")),
			),
		)
		return ast.NewLet(assignments, expr), nil
	} else if unitValue.String() == Month {
		if !t.versionAtLeast(4, 0, 0) {
			return nil, newPushdownFailure(
				"SQLScalarFunctionExpr(date_add)",
				"cannot push down to MongoDB < 4.0 (need overflowing $dateFromParts)",
			)
		}
		assignments := []*ast.LetVariable{
			ast.NewLetVariable(
				"dateParts",
				ast.NewFunction(
					bsonutil.OpDateToParts,
					ast.NewDocument(
						ast.NewDocumentElement("date", date),
					),
				),
			),
		}
		expr := ast.NewFunction(
			bsonutil.OpDateFromParts,
			ast.NewDocument(
				ast.NewDocumentElement("year", ast.NewVariableRef("dateParts.year")),
				ast.NewDocumentElement(
					"month",
					astutil.WrapInBinOp(
						bsonutil.OpAdd,
						ast.NewVariableRef("dateParts.month"),
						astutil.Int64Value(intervalValue.SQLInt().Value().(int64)),
					),
				),
				ast.NewDocumentElement("day", ast.NewVariableRef("dateParts.day")),
				ast.NewDocumentElement("hour", ast.NewVariableRef("dateParts.hour")),
				ast.NewDocumentElement("minute", ast.NewVariableRef("dateParts.minute")),
				ast.NewDocumentElement("second", ast.NewVariableRef("dateParts.second")),
				ast.NewDocumentElement("millisecond", ast.NewVariableRef("dateParts.millisecond")),
			),
		)
		return ast.NewLet(assignments, expr), nil
	} else if unitValue.String() == Week {
		if !t.versionAtLeast(4, 0, 0) {
			return nil, newPushdownFailure(
				"SQLScalarFunctionExpr(date_add)",
				"cannot push down to MongoDB < 4.0 (need overflowing $dateFromParts)",
			)
		}
		assignments := []*ast.LetVariable{
			ast.NewLetVariable(
				"dateParts",
				ast.NewFunction(
					bsonutil.OpDateToParts,
					ast.NewDocument(
						ast.NewDocumentElement("date", date),
					),
				),
			),
		}
		expr := ast.NewFunction(
			bsonutil.OpDateFromParts,
			ast.NewDocument(
				ast.NewDocumentElement("year", ast.NewVariableRef("dateParts.year")),
				ast.NewDocumentElement("month", ast.NewVariableRef("dateParts.month")),
				ast.NewDocumentElement(
					"day",
					astutil.WrapInBinOp(
						bsonutil.OpAdd,
						ast.NewVariableRef("dateParts.day"),
						astutil.Int64Value(intervalValue.SQLInt().Value().(int64)*7),
					),
				),
				ast.NewDocumentElement("hour", ast.NewVariableRef("dateParts.hour")),
				ast.NewDocumentElement("minute", ast.NewVariableRef("dateParts.minute")),
				ast.NewDocumentElement("second", ast.NewVariableRef("dateParts.second")),
				ast.NewDocumentElement("millisecond", ast.NewVariableRef("dateParts.millisecond")),
			),
		)
		return ast.NewLet(assignments, expr), nil
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

	dateRef := ast.NewVariableRef("date")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("date", date),
	}

	evaluation := astutil.WrapInNullCheckedCond(
		astutil.NullLiteral,
		astutil.WrapInOp(bsonutil.OpAdd, dateRef, astutil.Int64Value(ms)),
		dateRef,
	)

	return ast.NewLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) dateDiffToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 2)

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
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("days", days),
	}

	evaluation := astutil.WrapInNullCheckedCond(
		astutil.NullLiteral,
		astutil.WrapInCond(
			bound,
			daysRef,
			ast.NewBinary(bsonutil.OpGt, daysRef, upper),
			ast.NewBinary(bsonutil.OpLt, daysRef, lower),
		),
		date1, date2,
	)

	return ast.NewLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) dateFormatToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 2)

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

	wrapped, ok := t.formatDate(date, formatValue.String())
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateFormat)",
			"unable to push down format string",
			"formatString", formatValue.String(),
		)
	}

	return wrapped, nil
}

func (f *baseScalarFunctionExpr) dateSubToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	return f.dateArithmeticToAggregationLanguage(t, exprs, true)
}

func (f *baseScalarFunctionExpr) dayNameToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return astutil.WrapInNullCheckedCond(
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
	assertExactArgCount(exprs, 1)

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpDayOfMonth, exprs[0], t)
}

func (f *baseScalarFunctionExpr) dayOfWeekToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpDayOfWeek, exprs[0], t)
}

func (f *baseScalarFunctionExpr) dayOfYearToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpDayOfYear, exprs[0], t)
}

func (f *baseScalarFunctionExpr) degreesToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	if t.versionAtLeast(4, 1, 7) {
		return ast.NewFunction(bsonutil.OpRadiansToDegrees, args[0]), nil
	}

	return ast.NewBinary(bsonutil.OpDivide,
		ast.NewBinary(bsonutil.OpMultiply, args[0], astutil.FloatValue(180.0)),
		astutil.PiLiteral,
	), nil
}

func (f *baseScalarFunctionExpr) expToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return ast.NewFunction(bsonutil.OpExp, args[0]), nil
}

func (f *baseScalarFunctionExpr) extractToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 2)

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
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return ast.NewFunction(bsonutil.OpFloor, args[0]), nil
}

// formatDate wraps an Aggregation Expression that evaluates to a date
// in a date_format expression that will use '$dateFromString' to format
// a date to a string.
func (t *PushdownTranslator) formatDate(date ast.Expr, mysqlFormat string) (ast.Expr, bool) {
	formats := []string{}
	// postFormats are formats that mongodb cannot handle directly, they
	// will be weaved between format strings that mongodb can handle.  Consider
	// "%Y %d %b %h %e", the pseudocode we need to generate has to be:
	// concat(
	//    dateToString(date, "%Y %d "),
	//    b_code(date),
	//    dateToString(date, " %h "),
	//    e_code(date),
	//)
	// Because dateToString understands %Y, %d, and %h,
	// but does not understand %b and %e.
	postFormats := []rune{}
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
				case 'b':
					if !t.versionAtLeast(3, 6, 0) {
						return nil, false
					}
					formats = append(formats, format)
					postFormats = append(postFormats, 'b')
					format = ""
				case 'e':
					if !t.versionAtLeast(3, 6, 0) {
						return nil, false
					}
					formats = append(formats, format)
					postFormats = append(postFormats, 'e')
					format = ""
				case 'l':
					if !t.versionAtLeast(4, 0, 0) {
						return nil, false
					}
					formats = append(formats, format)
					postFormats = append(postFormats, 'l')
					format = ""
				case 'p':
					if !t.versionAtLeast(3, 6, 0) {
						return nil, false
					}
					formats = append(formats, format)
					postFormats = append(postFormats, 'p')
					format = ""
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
	formats = append(formats, format)

	if val, ok := date.(*ast.Constant); ok && val.Value.Type == bsontype.Null {
		return nil, true
	}

	createFormatChunk := func(date ast.Expr, format string) ast.Expr {
		if format == "" {
			return astutil.EmptyStringLiteral
		}
		return astutil.WrapInNullCheckedCond(
			astutil.NullLiteral,
			astutil.WrapInDateToString(date, format),
			date)
	}

	if len(formats) == 1 {
		return createFormatChunk(date, format), true
	}

	// We need to do a let here because our mongoast deduplication pass isn't correctly pulling out
	// the $month call.
	month := ast.NewVariableRef("month")
	evaluation := astutil.WrapInSwitch(
		astutil.EmptyStringLiteral,
		astutil.WrapInEqCase(month, astutil.Int32Value(1), astutil.StringValue("Jan")),
		astutil.WrapInEqCase(month, astutil.Int32Value(2), astutil.StringValue("Feb")),
		astutil.WrapInEqCase(month, astutil.Int32Value(3), astutil.StringValue("Mar")),
		astutil.WrapInEqCase(month, astutil.Int32Value(4), astutil.StringValue("Apr")),
		astutil.WrapInEqCase(month, astutil.Int32Value(5), astutil.StringValue("May")),
		astutil.WrapInEqCase(month, astutil.Int32Value(6), astutil.StringValue("Jun")),
		astutil.WrapInEqCase(month, astutil.Int32Value(7), astutil.StringValue("Jul")),
		astutil.WrapInEqCase(month, astutil.Int32Value(8), astutil.StringValue("Aug")),
		astutil.WrapInEqCase(month, astutil.Int32Value(9), astutil.StringValue("Sep")),
		astutil.WrapInEqCase(month, astutil.Int32Value(10), astutil.StringValue("Oct")),
		astutil.WrapInEqCase(month, astutil.Int32Value(11), astutil.StringValue("Nov")),
		astutil.WrapInEqCase(month, astutil.Int32Value(12), astutil.StringValue("Dec")),
	)

	getExpr := func(formatTy rune) ast.Expr {
		switch formatTy {
		// b is month name as three characters
		case 'b':
			inputLetAssignment := []*ast.LetVariable{ast.NewLetVariable("month", astutil.WrapInMonth(date))}
			return ast.NewLet(inputLetAssignment, evaluation)
		// e is non-zero padded day of month, so we just get day of month and convert to string.
		case 'e':
			return astutil.WrapInConvert(
				astutil.WrapInDayOfMonth(date),
				"string",
				astutil.NullLiteral,
				astutil.NullLiteral,
			)
		// l is 12 hour hour, non-zero padded.
		// hour == 0 : 12
		// hour == 12 : 12
		// else hour % 12
		case 'l':
			return astutil.WrapInConvert(
				astutil.WrapInSwitch(
					astutil.WrapInBinOp(bsonutil.OpMod,
						astutil.WrapInHour(date),
						astutil.Int32Value(12),
					),
					astutil.WrapInEqCase(astutil.WrapInHour(date), astutil.Int32Value(0), astutil.Int32Value(12)),
					astutil.WrapInEqCase(astutil.WrapInHour(date), astutil.Int32Value(12), astutil.Int32Value(12)),
				),
				"string",
				astutil.NullLiteral,
				astutil.NullLiteral,
			)
		// p is 'AM' or 'PM'
		case 'p':
			return astutil.WrapInCond(
				astutil.StringValue("AM"),
				astutil.StringValue("PM"),
				astutil.WrapInBinOp(
					bsonutil.OpLt,
					astutil.WrapInHour(date),
					astutil.Int32Value(12),
				),
			)
		}

		return nil
	}

	args := []ast.Expr{}
	for i, format := range formats[0 : len(formats)-1] {
		arg := createFormatChunk(date, format)
		args = append(args, arg)
		args = append(args, getExpr(postFormats[i]))
	}
	arg := createFormatChunk(date, formats[len(formats)-1])
	args = append(args, arg)
	return astutil.WrapInNullCheckedCond(
		astutil.NullLiteral,
		astutil.WrapInConcat(args),
	), true
}

func (f *baseScalarFunctionExpr) fromDaysToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	nRef := ast.NewVariableRef("n")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("n", args[0]),
	}

	// This should return "0000-00-00" if the input is too large (> maxFromDays)
	// or too low (< 366).
	dayOne := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	body := ast.NewBinary(bsonutil.OpAdd,
		astutil.DateConstant(dayOne),
		ast.NewBinary(bsonutil.OpMultiply,
			nRef, astutil.FloatValue(millisecondsPerDay),
		),
	)

	evaluation := astutil.WrapInNullCheckedCond(
		astutil.NullLiteral,
		astutil.WrapInCond(
			astutil.NullLiteral,
			body,
			ast.NewBinary(bsonutil.OpGt, nRef, astutil.Int32Value(maxFromDays)),
			ast.NewBinary(bsonutil.OpLt, nRef, astutil.Int32Value(366)),
		),
		nRef,
	)

	return ast.NewLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) fromUnixtimeToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertAtMostArgCount(exprs, 2)

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
			unixTimestamp, astutil.Int32Value(1e3),
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
		wrapped, ok := t.formatDate(ret, format.String())
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

	return astutil.WrapInNullCheckedCond(
		astutil.NullLiteral,
		astutil.WrapInOp(bsonutil.OpMax, args...),
		args...,
	), nil
}

func (f *baseScalarFunctionExpr) hourToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpHour, exprs[0], t)
}

func (f *baseScalarFunctionExpr) insertToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(insert)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	assertExactArgCount(exprs, 4)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	// subtract 1 to account for difference between mongo and mysql string indexing
	args[1] = ast.NewBinary(bsonutil.OpSubtract, args[1], astutil.OneInt32Literal)

	strRef, posRef := ast.NewVariableRef("str"), ast.NewVariableRef("pos")
	lengthRef, newstrRef := ast.NewVariableRef("len"), ast.NewVariableRef("newstr")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("str", args[0]),
		ast.NewLetVariable("pos", args[1]),
		ast.NewLetVariable("len", args[2]),
		ast.NewLetVariable("newstr", args[3]),
	}

	totalLength := ast.NewVariableRef("totalLength")
	totalLengthAssignment := []*ast.LetVariable{
		ast.NewLetVariable("totalLength", ast.NewFunction(bsonutil.OpStrlenCP, strRef)),
	}

	prefix, suffix := ast.NewVariableRef("prefix"), ast.NewVariableRef("suffix")
	ixAssignment := []*ast.LetVariable{
		ast.NewLetVariable("prefix", astutil.WrapInOp(bsonutil.OpSubstr, strRef, astutil.ZeroInt32Literal, posRef)),
		ast.NewLetVariable("suffix", astutil.WrapInOp(bsonutil.OpSubstr, strRef, astutil.WrapInOp(bsonutil.OpAdd, posRef, lengthRef), totalLength)),
	}

	concatenation := ast.NewLet(ixAssignment,
		astutil.WrapInOp(bsonutil.OpConcat, prefix, newstrRef, suffix),
	)

	posCheck := ast.NewLet(totalLengthAssignment,
		astutil.WrapInCond(strRef,
			concatenation,
			ast.NewBinary(bsonutil.OpLt, posRef, astutil.ZeroInt32Literal),
			ast.NewBinary(bsonutil.OpGte, posRef, totalLength),
		),
	)

	evaluation := astutil.WrapInNullCheckedCond(astutil.NullLiteral, posCheck, strRef, posRef, lengthRef, newstrRef)

	return ast.NewLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) instrToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(instr)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	assertExactArgCount(exprs, 2)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	str := args[0]
	substrRef := ast.NewVariableRef("substr")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("substr", args[1]),
	}

	// Mongo Aggregation Pipeline returns NULL if str is NULLish, like
	// we'd want. substr being NULL, however, is an error in the pipeline,
	// thus check substr for NULLisness.
	evaluation := astutil.WrapInNullCheckedCond(
		astutil.NullLiteral,
		ast.NewBinary(bsonutil.OpAdd,
			astutil.OneInt32Literal,
			astutil.WrapInOp(bsonutil.OpIndexOfCP, str, substrRef),
		),
		substrRef,
	)

	return ast.NewLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) lastDayToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

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
	assertExactArgCount(exprs, 1)

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

	return astutil.WrapInNullCheckedCond(
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
	assertExactArgCount(exprs, 2)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	strRef, lengthRef := ast.NewVariableRef("str"), ast.NewVariableRef("len")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("str", args[0]),
		ast.NewLetVariable("len", args[1]),
	}

	subStrLength := astutil.WrapInOp(bsonutil.OpMax, lengthRef, astutil.ZeroInt32Literal)
	subStrOp := astutil.WrapInOp(bsonutil.OpSubstr, strRef, astutil.ZeroInt32Literal, subStrLength)
	evaluation := astutil.WrapInNullCheckedCond(astutil.NullLiteral, subStrOp, strRef, lengthRef)

	return ast.NewLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) lengthToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(length)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	assertExactArgCount(exprs, 1)

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpStrLenBytes, exprs[0], t)
}

func (f *baseScalarFunctionExpr) locateToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(locate)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	assertEitherArgCount(exprs, 2, 3)

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

	return astutil.WrapInNullCheckedCond(astutil.NullLiteral, locate, args[0], args[1]), nil
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
	assertEitherArgCount(exprs, 1, 2)

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

func (f *baseScalarFunctionExpr) lnToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	return f.logarithmToAggregationLanguage(t, exprs, 0)
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
	assertExactArgCount(exprs, 1)

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

	return astutil.WrapInNullCheckedCond(astutil.NullLiteral, ltrimCond, args...), nil
}

func (f *baseScalarFunctionExpr) makeDateToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(3, 5, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(makeDate)",
			"cannot push down to MongoDB < 3.5",
		)
	}

	assertExactArgCount(exprs, 2)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	year, day := ast.NewVariableRef("year"), ast.NewVariableRef("day")
	inputLetStatement := []*ast.LetVariable{
		ast.NewLetVariable("year", args[0]),
		ast.NewLetVariable("day", args[1]),
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
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return astutil.WrapInNullCheckedCond(
		astutil.NullLiteral,
		ast.NewBinary(bsonutil.OpMultiply,
			astutil.Int32Value(1000),
			astutil.WrapInOp(bsonutil.OpMillisecond, args[0]),
		),
		args...,
	), nil

}

func (f *baseScalarFunctionExpr) midToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 3)

	return f.substringToAggregationLanguage(t, exprs)
}

func (f *baseScalarFunctionExpr) minuteToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpMinute, exprs[0], t)
}

func (f *baseScalarFunctionExpr) modToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 2)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return ast.NewBinary(bsonutil.OpMod, args[0], args[1]), nil
}

func (f *baseScalarFunctionExpr) monthNameToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return astutil.WrapInNullCheckedCond(
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
	assertExactArgCount(exprs, 1)

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpMonth, exprs[0], t)
}

func (f *baseScalarFunctionExpr) padToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr, isLeftPad bool) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(pad)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	assertExactArgCount(exprs, 3)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	strRef, lenRef, padStrRef := ast.NewVariableRef("str"), ast.NewVariableRef("len"), ast.NewVariableRef("padStr")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("str", args[0]),
		ast.NewLetVariable("len", args[1]),
		ast.NewLetVariable("padStr", args[2]),
	}

	padLenRef, padStrLenRef := ast.NewVariableRef("padLen"), ast.NewVariableRef("padStrLen")
	subAssignments := []*ast.LetVariable{
		ast.NewLetVariable("padStrLen", ast.NewFunction(bsonutil.OpStrlenCP, padStrRef)),
		ast.NewLetVariable("padLen", ast.NewBinary(bsonutil.OpSubtract,
			lenRef,
			ast.NewFunction(bsonutil.OpStrlenCP, strRef),
		)),
	}

	// logic for generating padding string:

	// do we even need to add padding? only if the desired output
	// length is > length of input string.
	paddingCond := ast.NewBinary(bsonutil.OpLt, ast.NewFunction(bsonutil.OpStrlenCP, strRef), lenRef)

	// number of times we need to repeat the padding string to fill space
	padStrRepeats := ast.NewFunction(bsonutil.OpCeil, ast.NewBinary(bsonutil.OpDivide, padLenRef, padStrLenRef))

	// generate an array with padStrRepeats occurrences of padStr
	padParts := ast.NewDocument(
		ast.NewDocumentElement(bsonutil.OpMap, ast.NewDocument(
			ast.NewDocumentElement("input", astutil.WrapInOp(bsonutil.OpRange, astutil.ZeroInt32Literal, padStrRepeats)),
			ast.NewDocumentElement("in", padStrRef),
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
		concatted = astutil.WrapInOp(bsonutil.OpConcat, fullPad, strRef)
	} else {
		concatted = astutil.WrapInOp(bsonutil.OpConcat, strRef, fullPad)
	}

	handleConcat := astutil.WrapInCond(
		astutil.NullLiteral, concatted,
		ast.NewBinary(bsonutil.OpEq, padStrLenRef, astutil.ZeroInt32Literal),
	)

	// handle everything in the case that input length >=0
	handleNonNegativeLength := astutil.WrapInCond(
		handleConcat,
		astutil.WrapInOp(bsonutil.OpSubstr, strRef, astutil.ZeroInt32Literal, lenRef),
		paddingCond,
	)

	// if length < 0, we just return null
	negativeCheck := astutil.WrapInCond(
		astutil.NullLiteral,
		handleNonNegativeLength,
		ast.NewBinary(bsonutil.OpLt, lenRef, astutil.ZeroInt32Literal),
	)

	evaluation := astutil.WrapInNullCheckedCond(
		astutil.NullLiteral,
		ast.NewLet(subAssignments, negativeCheck),
		strRef, lenRef, padStrRef,
	)

	return ast.NewLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) powToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 2)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return ast.NewBinary(bsonutil.OpPow, args[0], args[1]), nil
}

func (f *baseScalarFunctionExpr) quarterToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	dateRef := ast.NewVariableRef("date")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("date", args[0]),
	}

	one, two, three, four := astutil.OneInt32Literal, astutil.Int32Value(2), astutil.Int32Value(3), astutil.Int32Value(4)

	evaluation := astutil.WrapInNullCheckedCond(
		astutil.NullLiteral,
		astutil.WrapInOp(bsonutil.OpArrElemAt,
			ast.NewArray(one, one, one, two, two, two, three, three, three, four, four, four),
			ast.NewBinary(bsonutil.OpSubtract,
				ast.NewFunction(bsonutil.OpMonth, dateRef),
				astutil.OneInt32Literal,
			),
		),
		dateRef,
	)

	return ast.NewLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) radiansToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	if t.versionAtLeast(4, 1, 7) {
		return ast.NewFunction(bsonutil.OpDegreesToRadians, args[0]), nil
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

	assertExactArgCount(exprs, 2)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	str := args[0]

	// create array w/ args[1] values e.g. [0,1,2]
	rangeArr := astutil.WrapInRange(astutil.ZeroInt32Literal, args[1], astutil.OneInt32Literal)

	// create array of len arg[1], with each item being arg[0]
	m := ast.NewFunction(bsonutil.OpMap, ast.NewDocument(
		ast.NewDocumentElement("input", rangeArr),
		ast.NewDocumentElement("in", str),
	))

	repeat := astutil.WrapInReduce(m, astutil.EmptyStringLiteral,
		// append all values of this array together
		astutil.WrapInOp(bsonutil.OpConcat, astutil.ThisVarRef, astutil.ValueVarRef),
	)

	return astutil.WrapInNullCheckedCond(astutil.NullLiteral, repeat, args...), nil
}

func (f *baseScalarFunctionExpr) replaceToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(replace)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	assertExactArgCount(exprs, 3)

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

	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	strRef := ast.NewVariableRef("str")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("str", args[0]),
	}

	evaluation := astutil.WrapInNullCheckedCond(
		astutil.NullLiteral,
		astutil.WrapInReduce(
			astutil.WrapInOp(bsonutil.OpRange, astutil.ZeroInt32Literal, astutil.WrapInOp(bsonutil.OpStrlenCP, strRef)),
			astutil.EmptyStringLiteral,
			astutil.WrapInOp(bsonutil.OpConcat,
				astutil.WrapInOp(bsonutil.OpSubstr,
					strRef, astutil.ThisVarRef, astutil.OneInt32Literal,
				),
				astutil.ValueVarRef,
			),
		),
		strRef,
	)

	return ast.NewLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) rightToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(right)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	assertExactArgCount(exprs, 2)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	strRef, lenRef := ast.NewVariableRef("str"), ast.NewVariableRef("len")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("str", args[0]),
		ast.NewLetVariable("len", args[1]),
	}

	subStrLength := astutil.WrapInOp(bsonutil.OpMax, astutil.ZeroInt32Literal, lenRef)
	strLength := ast.NewFunction(bsonutil.OpStrlenCP, strRef)

	// start = max(0, strLen - subStrLen)
	start := astutil.WrapInOp(bsonutil.OpMax, astutil.ZeroInt32Literal, ast.NewBinary(bsonutil.OpSubtract, strLength, subStrLength))

	subStrOp := astutil.WrapInOp(bsonutil.OpSubstr, strRef, start, subStrLength)

	evaluation := astutil.WrapInNullCheckedCond(astutil.NullLiteral, subStrOp, strRef, lenRef)

	return ast.NewLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) roundToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertAtLeastArgCount(exprs, 1)

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
	assertExactArgCount(exprs, 1)

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

	return astutil.WrapInNullCheckedCond(astutil.NullLiteral, rtrimCond, args...), nil
}

func (f *baseScalarFunctionExpr) secondToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpSecond, exprs[0], t)
}

func (f *baseScalarFunctionExpr) signToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return astutil.WrapInNullCheckedCond(
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
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	if t.versionAtLeast(4, 1, 7) {
		return ast.NewFunction(bsonutil.OpSin, args[0]), nil
	}

	inputRef := ast.NewVariableRef("input")
	inputLetAssignment := []*ast.LetVariable{
		ast.NewLetVariable("input", args[0]),
	}

	absInput := ast.NewVariableRef("absInput")
	absInputLetAssignment := []*ast.LetVariable{
		ast.NewLetVariable("absInput", ast.NewFunction(bsonutil.OpAbs, inputRef)),
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
	return ast.NewLet(inputLetAssignment,
		ast.NewLet(absInputLetAssignment,
			ast.NewLet(remPhaseAssignment,
				astutil.WrapInCond(zeroCase,
					ast.NewBinary(bsonutil.OpMultiply, astutil.FloatValue(-1.0), zeroCase),
					ast.NewBinary(bsonutil.OpGte, inputRef, astutil.ZeroInt32Literal),
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

	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	nRef := ast.NewVariableRef("n")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("n", args[0]),
	}

	evaluation := astutil.WrapInNullCheckedCond(
		astutil.NullLiteral,
		astutil.WrapInReduce(
			astutil.WrapInRange(astutil.ZeroInt32Literal, nRef, astutil.OneInt32Literal),
			astutil.EmptyStringLiteral,
			astutil.WrapInOp(bsonutil.OpConcat, astutil.ValueVarRef, astutil.StringValue(" ")),
		),
		nRef,
	)

	return ast.NewLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) sqrtToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

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

func (f *baseScalarFunctionExpr) strToDateToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(4, 0, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(strToDate)",
			"cannot push down to MongoDB < 4.0",
		)
	}

	assertExactArgCount(exprs, 2)

	str, err := t.ToAggregationLanguage(exprs[0])
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

	formatValueString := formatValueExpr.Value.String()

	// if the input date string is null, return null
	if val, ok := str.(*ast.Constant); ok && val.Value.Type == bsontype.Null {
		return nil, nil
	}

	// pushdown uses $dateFromString, so it is unable to support the
	// MySQL date format operators for %a, %b, %c, %e, %M, %W, %y
	// which are supported in the in-memory evaluation
	var format string
	for i := 0; i < len(formatValueString); i++ {
		if formatValueString[i] == '%' {
			if i != len(formatValueString)-1 {
				switch formatValueString[i+1] {
				case 'd':
					format += "%d"
				case 'H':
					format += "%H"
				case 'i':
					format += "%M"
				case 'm':
					format += "%m"
				case 's', 'S':
					format += "%S"
				case 'T':
					format += "%H:%M:%S"
				case 'Y':
					format += "%Y"
				default:
					return nil, newPushdownFailure(
						"SQLScalarFunctionExpr(strToDate)",
						"unable to push down format string",
						"formatString", formatValueString,
					)
				}
				i++
			} else {
				// MongoDB fails when the last character is a % sign in the format string.
				return nil, newPushdownFailure(
					"SQLScalarFunctionExpr(strToDate)",
					"unable to push down format string",
					"formatString", formatValueString,
				)
			}
		} else {
			format += string(formatValueString[i])
		}
	}

	// $dateFromString requires that the format string and corresponding date string
	// contain %Y, %m, and %d and the corresponding year, month, and date values.
	// This manually adds default year, month, and date values if the inputted
	// strings don't contains them in order to match the in-memory evaluation default values.
	var arr []ast.Expr
	arr = append(arr, str)

	if !strings.Contains(format, "%Y") {
		format += " %Y"
		arr = append(arr, astutil.StringValue(" 0000"))
	}

	if !strings.Contains(format, "%m") {
		format += " %m"
		arr = append(arr, astutil.StringValue(" 01"))
	}

	if !strings.Contains(format, "%d") {
		format += " %d"
		arr = append(arr, astutil.StringValue(" 01"))
	}

	newString := astutil.WrapInConcat(arr)

	return astutil.WrapInNullCheckedCond(
		astutil.NullLiteral,
		ast.NewFunction(bsonutil.OpDateFromString, ast.NewDocument(
			ast.NewDocumentElement("dateString", newString),
			ast.NewDocumentElement("format", astutil.StringValue(format)),
			ast.NewDocumentElement("onError", astutil.NullLiteral),
		)),
		newString,
	), nil
}

func (f *baseScalarFunctionExpr) substringIndexToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(substringIndex)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	assertExactArgCount(exprs, 3)

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
			args[2],
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
	assertEitherArgCount(exprs, 2, 3)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	strRef := ast.NewVariableRef("str")
	strAssignment := []*ast.LetVariable{
		ast.NewLetVariable("str", args[0]),
	}

	var length, strLen ast.Expr

	// store the string's length since it is reused in multiple places.
	strLenAssignment := make([]*ast.LetVariable, 0, 1)
	switch sa := args[0].(type) {
	case *ast.Constant:
		strVal, isStr := sa.Value.StringValueOK()
		if isStr {
			strLen = astutil.Int64Value(int64(len(strVal)))
		}
	default:
		strLen = ast.NewVariableRef("strLen")
		strLenAssignment = append(strLenAssignment,
			ast.NewLetVariable("strLen", ast.NewFunction(bsonutil.OpStrlenCP, strRef)),
		)
	}

	// get the "pos" and "length" arguments
	subAssignments := make([]*ast.LetVariable, 1, 2)
	subCondArgs := make([]ast.Expr, 1, 2)

	roundedPosRef, roundedNegPosRef := ast.NewVariableRef("roundedPos"), ast.NewVariableRef("roundedNegPos")

	// the position argument needs to be calculated
	calculatedPos := ast.NewLet(
		[]*ast.LetVariable{
			ast.NewLetVariable("roundedPos", args[1]),
			ast.NewLetVariable("roundedNegPos",
				ast.NewBinary(bsonutil.OpMultiply, args[1], astutil.Int32Value(-1)),
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

	posRef := ast.NewVariableRef("pos")

	// if it is not literal or a column, the null-check can (and should) be on the binding
	switch args[1].(type) {
	case *ast.Constant, *ast.FieldRef:
	default:
		args[1] = posRef
	}

	subAssignments[0] = ast.NewLetVariable("pos", calculatedPos)
	subCondArgs[0] = args[1]

	if len(args) == 2 {
		// if length is not provided, use the str length.
		length = strLen
	} else {
		lengthRef := ast.NewVariableRef("length")
		subAssignments = append(subAssignments, ast.NewLetVariable("length", args[2]))
		length = astutil.WrapInOp(bsonutil.OpMax, astutil.ZeroInt32Literal, lengthRef)
		subCondArgs = append(subCondArgs, lengthRef)
	}

	return ast.NewLet(strAssignment,
		astutil.WrapInNullCheckedCond(
			astutil.NullLiteral,
			ast.NewLet(strLenAssignment,
				ast.NewLet(subAssignments,
					astutil.WrapInNullCheckedCond(
						astutil.NullLiteral,
						astutil.WrapInOp(bsonutil.OpSubstr, strRef, posRef, length),
						subCondArgs...,
					),
				),
			),
			strRef,
		),
	), nil
}

func (f *baseScalarFunctionExpr) tanToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	if t.versionAtLeast(4, 1, 7) {
		args, err := t.translateArgs(exprs)
		if err != nil {
			return nil, err
		}
		return ast.NewFunction(bsonutil.OpTan, args[0]), nil
	}

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

	assertExactArgCount(exprs, 3)

	unit := exprs[0].String()
	args, err := t.translateArgs(exprs[1:])
	if err != nil {
		return nil, err
	}
	interval := args[0]

	// This is a very large and costly expression, make sure to bind it in
	// a let (in the switch at the end of the function).
	timestampArgRef := ast.NewVariableRef("timestampArg")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("timestampArg", args[1]),
	}

	// handleSimpleCase generates code for cases where we do not need to
	// use $dateFromParts, we just round the interval if the round argument
	// is true, and multiply by the number of milliseconds corresponded to
	// by 'u' then add to the timestamp.
	handleSimpleCase := func(u string, round bool) *ast.Binary {
		if round {
			return ast.NewBinary(bsonutil.OpAdd,
				timestampArgRef,
				ast.NewBinary(bsonutil.OpMultiply,
					astutil.WrapInRound(interval),
					astutil.FloatValue(toMilliseconds[u]),
				),
			)
		}
		return ast.NewBinary(bsonutil.OpAdd,
			timestampArgRef,
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
		dayExpr := ast.NewFunction(bsonutil.OpDayOfMonth, timestampArgRef)

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
					ast.NewFunction(bsonutil.OpDayOfMonth, timestampArgRef),
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
			ast.NewDocumentElement("hour", ast.NewFunction(bsonutil.OpHour, timestampArgRef)),
			ast.NewDocumentElement("minute", ast.NewFunction(bsonutil.OpMinute, timestampArgRef)),
			ast.NewDocumentElement("second", ast.NewFunction(bsonutil.OpSecond, timestampArgRef)),
			ast.NewDocumentElement("millisecond", ast.NewFunction(bsonutil.OpMillisecond, timestampArgRef)),
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
					ast.NewFunction(bsonutil.OpYear, timestampArgRef),
				)),
				ast.NewDocumentElement("month", ast.NewFunction(bsonutil.OpMonth, timestampArgRef)),
				ast.NewDocumentElement("day", ast.NewFunction(bsonutil.OpDayOfMonth, timestampArgRef)),
				ast.NewDocumentElement("hour", ast.NewFunction(bsonutil.OpHour, timestampArgRef)),
				ast.NewDocumentElement("minute", ast.NewFunction(bsonutil.OpMinute, timestampArgRef)),
				ast.NewDocumentElement("second", ast.NewFunction(bsonutil.OpSecond, timestampArgRef)),
				ast.NewDocumentElement("millisecond", ast.NewFunction(bsonutil.OpMillisecond, timestampArgRef)),
			)
			return ast.NewFunction(bsonutil.OpDateFromParts, template)

		// For Quarter and Month intervals, only the SharedComputation
		// part changes.
		case Quarter:
			// SharedComputation = Month + round(interval) * 3 - 1.
			sharedComputationLetAssignment = []*ast.LetVariable{
				ast.NewLetVariable("sharedComputation", ast.NewBinary(bsonutil.OpSubtract,
					ast.NewBinary(bsonutil.OpAdd,
						ast.NewFunction(bsonutil.OpMonth, timestampArgRef),
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
						ast.NewFunction(bsonutil.OpMonth, timestampArgRef),
						astutil.WrapInRound(interval),
					),
					astutil.OneInt32Literal,
				)),
			}
		}

		newYearMonthLetAssignment = []*ast.LetVariable{
			// Year = Year + floor(SharedComputation / 12).
			ast.NewLetVariable("newYear", ast.NewBinary(bsonutil.OpAdd,
				ast.NewFunction(bsonutil.OpYear, timestampArgRef),
				astutil.WrapInOp(bsonutil.OpFloor,
					ast.NewBinary(bsonutil.OpDivide, sharedComputationRef, astutil.Int32Value(12))),
			)),

			// Month = ((SharedComputation % 12) + 12) % 12 + 1 in order to get month between 1-12.
			// want to calculate (sharedComputation mod 12), but Go's % operator does remainder
			// a mod b = ((a % b) + b) % b
			ast.NewLetVariable("newMonth", ast.NewBinary(bsonutil.OpAdd,
				astutil.OneInt32Literal,
				ast.NewBinary(bsonutil.OpMod,
					ast.NewBinary(bsonutil.OpAdd,
						ast.NewBinary(bsonutil.OpMod, sharedComputationRef, astutil.Int32Value(12)),
						astutil.Int32Value(12),
					),
					astutil.Int32Value(12)),
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
		return ast.NewLet(assignments, handleDateFromPartsCase(unit)), nil
	// It is wrong to round for Second, and rounding for Microsecond is
	// just pointless since MongoDB supports only milliseconds, and will
	// automatically round to the nearest millisecond for us.
	case Second, Microsecond:
		return ast.NewLet(assignments, handleSimpleCase(unit, false)), nil
	default:
		return ast.NewLet(assignments, handleSimpleCase(unit, true)), nil
	}
}

func (f *baseScalarFunctionExpr) timestampDiffToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(3, 5, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(timestampDiff)",
			"cannot push down to MongoDB < 3.5",
		)
	}

	assertExactArgCount(exprs, 3)

	unit := exprs[0].String()

	args, err := t.translateArgs(exprs[1:])
	if err != nil {
		return nil, err
	}

	// This is a very large and costly expression, make sure to bind it in
	// a let (in the switch at the end of the function).
	timestampArg1Ref, timestampArg2Ref := ast.NewVariableRef("timestampArg1"), ast.NewVariableRef("timestampArg2")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("timestampArg1", args[0]),
		ast.NewLetVariable("timestampArg2", args[1]),
	}

	// handleSimpleCase generates code for cases where we do not need to
	// use and date part access functions (like $dayOfMonth), we just
	// subtract: timestampArg2 - timestampArg1 then divide by the number of
	// milliseconds corresponded to by 'u'.
	handleSimpleCase := func(u string) ast.Expr {
		return astutil.WrapInIntDiv(
			ast.NewBinary(bsonutil.OpSubtract, timestampArg2Ref, timestampArg1Ref),
			astutil.FloatValue(toMilliseconds[u]),
		)
	}

	// handleDatePartsCase handles cases where we need to use
	// date part access functions (like $dayOfMonth).
	handleDatePartsCase := func(u string) ast.Expr {
		year1, year2 := ast.NewVariableRef("year1"), ast.NewVariableRef("year2")
		month1, month2 := ast.NewVariableRef("month1"), ast.NewVariableRef("month2")
		datePartsLetAssignment := []*ast.LetVariable{
			ast.NewLetVariable("year1", ast.NewFunction(bsonutil.OpYear, timestampArg1Ref)),
			ast.NewLetVariable("month1", ast.NewFunction(bsonutil.OpMonth, timestampArg1Ref)),
			ast.NewLetVariable("day1", ast.NewFunction(bsonutil.OpDayOfMonth, timestampArg1Ref)),
			ast.NewLetVariable("hour1", ast.NewFunction(bsonutil.OpHour, timestampArg1Ref)),
			ast.NewLetVariable("minute1", ast.NewFunction(bsonutil.OpMinute, timestampArg1Ref)),
			ast.NewLetVariable("second1", ast.NewFunction(bsonutil.OpSecond, timestampArg1Ref)),
			ast.NewLetVariable("millisecond1", ast.NewFunction(bsonutil.OpMillisecond, timestampArg1Ref)),
			ast.NewLetVariable("year2", ast.NewFunction(bsonutil.OpYear, timestampArg2Ref)),
			ast.NewLetVariable("month2", ast.NewFunction(bsonutil.OpMonth, timestampArg2Ref)),
			ast.NewLetVariable("day2", ast.NewFunction(bsonutil.OpDayOfMonth, timestampArg2Ref)),
			ast.NewLetVariable("hour2", ast.NewFunction(bsonutil.OpHour, timestampArg2Ref)),
			ast.NewLetVariable("minute2", ast.NewFunction(bsonutil.OpMinute, timestampArg2Ref)),
			ast.NewLetVariable("second2", ast.NewFunction(bsonutil.OpSecond, timestampArg2Ref)),
			ast.NewLetVariable("millisecond2", ast.NewFunction(bsonutil.OpMillisecond, timestampArg2Ref)),
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

	// ast.NewLet to bind $$timestampArg1 and 2.
	switch unit {
	case Year, Month, Quarter:
		return ast.NewLet(assignments, handleDatePartsCase(unit)), nil
	default:
		return ast.NewLet(assignments, handleSimpleCase(unit)), nil
	}
}

func (f *baseScalarFunctionExpr) toSecondsToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

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

// We'll just accept a date as a time. Tableau uses TIME_TO_SEC.
func (f *baseScalarFunctionExpr) timeToSecToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	if !t.versionAtLeast(3, 6, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(time_to_sec)",
			"cannot push down to MongoDB < 3.6",
		)
	}
	assertExactArgCount(exprs, 1)

	// We will only pushdown for things statically known to be datetime/date because
	// we do not have a time type, and doing reconciliation will cause an issue with
	// time strings. Since the goal is just to make pushdown work in BI tools attached to
	// ADL, this will be fine. If someone needs to call time_to_sec on another type they
	// will need to manually convert to a datetime in a way that is suitable to their
	// uses.
	if exprs[0].EvalType() != types.EvalDatetime && exprs[0].EvalType() != types.EvalDate {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(time_to_sec)",
			"can only push down for values known to be datetime statically",
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}
	assignments := []*ast.LetVariable{
		ast.NewLetVariable(
			"dateParts",
			ast.NewFunction(
				bsonutil.OpDateToParts,
				ast.NewDocument(
					ast.NewDocumentElement("date", args[0]),
				),
			),
		),
	}
	expr := ast.NewFunction(
		bsonutil.OpAdd,
		ast.NewArray(
			ast.NewFunction(bsonutil.OpMultiply, ast.NewArray(ast.NewVariableRef("dateParts.hour"), astutil.Int64Value(60*60))),
			ast.NewFunction(bsonutil.OpMultiply, ast.NewArray(ast.NewVariableRef("dateParts.minute"), astutil.Int64Value(60))),
			ast.NewVariableRef("dateParts.second"),
		),
	)
	return ast.NewLet(assignments, expr), nil
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

	return astutil.WrapInNullCheckedCond(astutil.NullLiteral, trim, args...), nil
}

func (f *baseScalarFunctionExpr) truncateToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	// Since the Mongo agg language currently does not have a $trunc operator for versions < 4.1.9, we push down the mySQL truncate() function by using
	// the following algorithm in the MongoDB agg language. For versions >= 4.1.9, we simply use $trunc.
	// 1. If both parameters are greater than or equal to zero, then the second parameter is used as an exponent with base 10.
	// Then, the first parameter is multiplied by this value. For example if we have 3.2456 and 2, then we get 3.2456 * 10^2 = 324.56.
	// Then, we take the floor (nearest integer less than the given value) of the calculated value. Here, it would be floor(324.56) = 324.
	// Finally, we divide by 10 to the power of the second parameter. Here, this would be 324/10^2 = 3.24.
	// 2. If the first parameter is greater than zero and the second parameter is less than zero, the procedure changes a bit. We take 10
	// raised to the absolute value of the second parameter. Then we divide the first parameter by this value. For example, if we have 324.56 and -2,
	// we would get 324.56/10^2 = 3.2456. Then, we take the floor of this value, which is 3, and finally multiply that by 10 to the power of the second parameter,
	// which gives us 300 (negative numbers truncate places to the left of the decimal point).
	// If the first parameter is less than zero, then we simply use the ceiling rather than the floor at the appropriate part of the algorithm.
	// If the second parameter is a float, it is rounded to the nearest integer and then the above process is repeated with that integer.

	assertExactArgCount(exprs, 2)

	dValueExpr, isScalar := exprs[1].(SQLValueExpr)
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	numToTruncate := args[0]
	truncatePlaces := args[1]

	if t.versionAtLeast(4, 1, 9) {
		return astutil.WrapInOp(bsonutil.OpTrunc, numToTruncate, truncatePlaces), nil
	}

	var d float64
	var pow ast.Expr

	if isScalar {
		d = values.Float64(dValueExpr.Value)
		pow = astutil.FloatValue(math.Pow(10, math.Abs(math.Round(d))))
	} else {
		pow = astutil.WrapInOp(bsonutil.OpPow, astutil.Int32Value(10), astutil.WrapInOp(bsonutil.OpAbs, truncatePlaces))
	}

	multiplyThenDivide := ast.NewBinary(bsonutil.OpDivide,
		astutil.WrapInCond(
			astutil.WrapInOp(bsonutil.OpFloor, ast.NewBinary(bsonutil.OpMultiply, numToTruncate, pow)),
			astutil.WrapInOp(bsonutil.OpCeil, ast.NewBinary(bsonutil.OpMultiply, numToTruncate, pow)),
			ast.NewBinary(bsonutil.OpGte, numToTruncate, astutil.ZeroInt32Literal),
		),
		pow,
	)

	divideThenMultiply := ast.NewBinary(bsonutil.OpMultiply,
		astutil.WrapInCond(
			astutil.WrapInOp(bsonutil.OpFloor, ast.NewBinary(bsonutil.OpDivide, numToTruncate, pow)),
			astutil.WrapInOp(bsonutil.OpCeil, ast.NewBinary(bsonutil.OpDivide, numToTruncate, pow)),
			ast.NewBinary(bsonutil.OpGte, numToTruncate, astutil.ZeroInt32Literal),
		),
		pow,
	)

	if isScalar {
		if d >= 0 {
			return multiplyThenDivide, nil
		}
		return divideThenMultiply, nil
	}

	return astutil.WrapInCond(
		multiplyThenDivide,
		divideThenMultiply,
		ast.NewBinary(bsonutil.OpGte, truncatePlaces, astutil.ZeroInt32Literal),
	), nil

}

func (f *baseScalarFunctionExpr) ucaseToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpToUpper, exprs[0], t)
}

func (f *baseScalarFunctionExpr) unixTimestampToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	now := time.Now()

	if len(exprs) != 1 {
		return astutil.Int64Value(now.Unix()), nil
	}

	if !t.versionAtLeast(3, 6, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(unix_timestamp)",
			"cannot push down to MongoDB < 3.6",
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	// Subtract epoch (1970-01-01) from the argument in MongoDB, then
	// convert ms to seconds. When using $subtract on two dates in
	// MongoDB, the number of milliseconds between the two
	// timestamps is returned.
	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	tz, tzErr := dateutil.GetIANATimezoneName(now)
	if tzErr != nil {
		return nil, newPushdownFailure(f.ExprName(), "failed to get timezone name", tzErr.Error())
	}

	diffRef := ast.NewVariableRef("diff")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("diff", ast.NewFunction(bsonutil.OpTrunc,
			ast.NewBinary(bsonutil.OpDivide,
				ast.NewBinary(bsonutil.OpSubtract,
					ast.NewFunction(bsonutil.OpDateFromString, ast.NewDocument(
						ast.NewDocumentElement("dateString",
							ast.NewFunction(bsonutil.OpDateToString, ast.NewDocument(
								ast.NewDocumentElement("date", args[0]),
								ast.NewDocumentElement("format", astutil.StringConstant("%Y-%m-%dT%H:%M:%S.%L")),
							)),
						),
						ast.NewDocumentElement("timezone", astutil.StringConstant(tz)),
					)),
					astutil.DateConstant(epoch),
				),
				astutil.Int32Value(1000),
			),
		)),
	}

	evaluation := astutil.WrapInCond(
		diffRef,
		astutil.FloatValue(0.0),
		astutil.WrapInOp(bsonutil.OpGt, diffRef, astutil.ZeroInt32Literal),
	)

	return ast.NewLet(assignments, evaluation), nil
}

func (f *baseScalarFunctionExpr) weekToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertEitherArgCount(exprs, 1, 2)

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
	assertExactArgCount(exprs, 1)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	dateRef := ast.NewVariableRef("date")
	assignments := []*ast.LetVariable{
		ast.NewLetVariable("date", args[0]),
	}

	seven := astutil.Int32Value(7)

	evaluation := astutil.WrapInNullCheckedCond(
		astutil.NullLiteral,
		ast.NewBinary(bsonutil.OpMod,
			ast.NewBinary(bsonutil.OpAdd,
				ast.NewBinary(bsonutil.OpMod,
					ast.NewBinary(bsonutil.OpSubtract,
						ast.NewFunction(bsonutil.OpDayOfWeek, dateRef),
						astutil.Int32Value(2),
					),
					seven,
				),
				seven,
			),
			seven,
		),
		dateRef,
	)

	return ast.NewLet(assignments, evaluation), nil

}

func (f *baseScalarFunctionExpr) yearToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertExactArgCount(exprs, 1)

	return wrapSingleArgFuncWithNullCheck(bsonutil.OpYear, exprs[0], t)
}

func (f *baseScalarFunctionExpr) yearWeekToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (ast.Expr, PushdownFailure) {
	assertEitherArgCount(exprs, 1, 2)

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

	dateRef := ast.NewVariableRef("date")
	inputAssignment := []*ast.LetVariable{
		ast.NewLetVariable("date", args[0]),
	}

	month, year := ast.NewVariableRef("month"), ast.NewVariableRef("year")
	monthAssignment := []*ast.LetVariable{
		ast.NewLetVariable("month", ast.NewFunction(bsonutil.OpMonth, dateRef)),
		ast.NewLetVariable("year", ast.NewFunction(bsonutil.OpYear, dateRef)),
	}

	var weekCalc ast.Expr

	// Unlike WEEK, YEARWEEK always uses the 1-53 modes. Thus
	// we always call week with the 1-53 of a 0-53, 1-53 pair.
	switch mode {

	// First day of week: Sunday, with a Sunday in this year.
	case 0, 2:
		weekCalc = astutil.WrapInWeekCalculation(dateRef, 2)
	// First day of weekCalc: Monday, with 4 days in this year.
	case 1, 3:
		weekCalc = astutil.WrapInWeekCalculation(dateRef, 3)
	// First day of weekCalc: Sunday, with 4 days in this year.
	case 4, 6:
		weekCalc = astutil.WrapInWeekCalculation(dateRef, 6)
	// First day of weekCalc: Monday, with a Monday in this year.
	case 5, 7:
		weekCalc = astutil.WrapInWeekCalculation(dateRef, 7)
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

	return ast.NewLet(inputAssignment,
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

// wrapSingleArgFuncWithNullCheck returns a null checked version of the
// argued function (operator and argument).
func wrapSingleArgFuncWithNullCheck(op string, expr SQLExpr, t *PushdownTranslator) (ast.Expr, PushdownFailure) {
	arg, err := t.ToAggregationLanguage(expr)
	if err != nil {
		return nil, err
	}

	return astutil.WrapSingleArgFuncWithNullCheck(op, arg), nil
}

// helper functions for customizing panic messages
func assertExactArgCount(args []SQLExpr, expectedCount int) {
	if len(args) != expectedCount {
		panic(fmt.Sprintf("need exactly %d args, found %d", expectedCount, len(args)))
	}
}

func assertAtLeastArgCount(args []SQLExpr, expectedMinCount int) {
	if len(args) < expectedMinCount {
		panic(fmt.Sprintf("need at least %d args, found %d", expectedMinCount, len(args)))
	}
}

func assertAtMostArgCount(args []SQLExpr, expectedMaxCount int) {
	if len(args) > expectedMaxCount {
		panic(fmt.Sprintf("need at most %d args, found %d", expectedMaxCount, len(args)))
	}
}

func assertEitherArgCount(args []SQLExpr, expectedCountOne, expectedCountTwo int) {
	if !(len(args) == expectedCountOne || len(args) == expectedCountTwo) {
		panic(fmt.Sprintf("need either %d or %d args, found %d", expectedCountOne, expectedCountTwo, len(args)))
	}
}
