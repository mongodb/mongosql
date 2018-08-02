package evaluator

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	"github.com/shopspring/decimal"
)

const (
	regexCharsToEscape = ".^$*+?()[{\\|"
	maxPrecisionInt    = int64(1 << 53)
	punctuation        = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
)

var (
	// ErrNotFullyPushedDown is the error returned when a query that hits MongoDB isn't fully
	// pushed down and the mongosqld_full_pushdown_exec_mode system variable is set.
	ErrNotFullyPushedDown = errors.New("query not fully pushed down")
)

// MySQLCleanNumericString cleans up a numeric string using MySQL's rules (trim, then
// take everything before the first character that isn't . or a number). Must
// handle -, and should return "0" if no viable number can be found.
func MySQLCleanNumericString(s string) string {
	var out bytes.Buffer
	firstDecimal := true
	s = strings.TrimLeft(s, " \t\v\n\r")
	if len(s) == 0 {
		return "0"
	}
	firstChar := s[0]
	if firstChar == '-' || firstChar == '+' {
		out.WriteRune(rune(firstChar))
		s = s[1:]
	}
	for _, c := range s {
		if c == '.' {
			if firstDecimal {
				out.WriteRune(c)
				firstDecimal = false
				continue
			}
			break
		}
		if c >= '0' && c <= '9' {
			out.WriteRune(c)
			continue
		}
		break
	}
	ret := out.String()
	if len(ret) == 0 || ret == "-" {
		return "0"
	}
	return ret
}

// MySQLCleanScientificNotationString cleans up a numeric string that may
// contain scientific notation.
func MySQLCleanScientificNotationString(s string) string {
	s = strings.TrimLeft(s, " \t\v\n\r")
	splitted := strings.Split(s, "e")
	base, exponent := s, ""
	if len(splitted) > 1 {
		// Any extra parts can be safely dropped, mysql only
		// considers the first e.
		base, exponent = splitted[0], splitted[1]
	}

	cleanBase := MySQLCleanNumericString(base)
	// If the base part is _changed_ by MySQLCleanNumericString
	// that means it had trailing charaters, and the exponent should be
	// ignored. Unfortunately, we will TrimLeft twice to make this
	// check work.
	if cleanBase != base {
		return cleanBase
	}
	exponent = MySQLCleanNumericString(exponent)
	return base + "e" + exponent
}

func compareDecimal128(left, right decimal.Decimal) (int, error) {
	return left.Cmp(right), nil
}

func compareFloats(left, right float64) (int, error) {
	cmp := left - right
	if cmp < 0 {
		return -1, nil
	} else if cmp > 0 {
		return 1, nil
	}
	return 0, nil
}

func compareInts(left, right int) int {
	if left < right {
		return -1
	} else if left > right {
		return 1
	}
	return 0
}

// ConvertSQLValueToPattern returns a regular expression that will match the
// string representation of the provided SQLValue.
func ConvertSQLValueToPattern(value SQLValue, escapeChar rune) string {
	pattern := value.String()
	regex := "^"
	escaped := false
	for _, c := range pattern {
		if !escaped && c == escapeChar {
			escaped = true
			continue
		}

		switch {
		case c == '_':
			if escaped {
				regex += "_"
			} else {
				regex += "."
			}
		case c == '%':
			if escaped {
				regex += "%"
			} else {
				regex += ".*"
			}
		case strings.Contains(regexCharsToEscape, string(c)):
			regex += "\\" + string(c)
		default:
			regex += string(c)
		}

		escaped = false
	}

	regex += "$"

	return regex
}

// doArithmetic performs the given arithmetic operation using
// leftVal and rightVal as operands.
func doArithmetic(leftVal, rightVal SQLValue, op ArithmeticOperator) (SQLValue, error) {

	preferenceType := preferentialType(leftVal, rightVal)
	useDecimal := preferenceType == EvalDecimal128

	leftType := leftVal.EvalType()
	rightType := rightVal.EvalType()

	hasUnsigned := leftType == EvalUint64 || rightType == EvalUint64

	if hasUnsigned {
		useDecimal = true
		preferenceType = EvalDecimal128
	}

	// check if both operands are timestamp or date since
	// arithmetic between time types result in an integer
	if preferenceType == EvalDate || preferenceType == EvalDatetime {
		preferenceType = EvalInt64
	}

	if preferenceType == EvalBoolean {
		preferenceType = EvalDouble
	}

	leftFloat := Float64(leftVal)
	rightFloat := Float64(rightVal)

	leftDecimal := Decimal(leftVal)
	rightDecimal := Decimal(rightVal)

	// use decimal type if Float64() value loses precision
	useDecimal = useDecimal ||
		Int64(leftVal) > maxPrecisionInt ||
		Int64(rightVal) > maxPrecisionInt

	valueD := decimal.Zero
	valueF := 0.0

	exact := false

	switch op {
	case ADD:
		decimalSum := leftDecimal.Add(rightDecimal)
		floatSum := leftFloat + rightFloat
		_, exact = decimalSum.Float64()
		valueD = decimalSum
		valueF = floatSum
	case DIV:
		decimalResult := leftDecimal.Div(rightDecimal)
		floatResult := leftFloat / rightFloat
		_, exact = decimalResult.Float64()
		if useDecimal || !exact {
			// 4 comes from the div_precision_increment variable which
			// we do not allow to be set.
			scale := leftDecimal.Exponent() - 4
			decimalResult = decimalResult.Round(-scale)
			return SQLDecimal128(decimalResult), nil
		}
		return SQLFloat(floatResult), nil
	case MULT:
		decimalProduct := leftDecimal.Mul(rightDecimal)
		floatProduct := leftFloat * rightFloat
		_, exact = decimalProduct.Float64()
		valueD = decimalProduct
		valueF = floatProduct
	case SUB:
		decimalDiff := leftDecimal.Sub(rightDecimal)
		floatDiff := leftFloat - rightFloat
		_, exact = decimalDiff.Float64()
		valueD = decimalDiff
		valueF = floatDiff
	default:
		return nil, fmt.Errorf("unrecognized arithmetic operator: %v", op)
	}

	if !exact || useDecimal {
		return SQLDecimal128(valueD), nil
	}

	switch preferenceType {
	case EvalInt64, EvalInt32:
		return SQLInt64(valueF), nil
	case EvalDouble:
		return SQLFloat(valueF), nil
	}
	return SQLFloat(valueF), nil
}

// fast2Sum returns the exact unevaluated sum of a and b
// where the first member is the float64 nearest the sum
// (ties to even) and the second member is the remainder
// (assuming |b| <= |a|).
//
// T. J. Dekker. A floating-point technique for extending
// the available precision. Numerische Mathematik,
// 18(3):224–242, 1971.
func fast2Sum(a, b float64) (float64, float64) {
	var s, z, t float64
	s = a + b
	z = s - a
	t = b - z
	return s, t
}

func getSQLTupleExprs(left, right SQLExpr) ([]SQLExpr, []SQLExpr, error) {

	getExprs := func(expr SQLExpr) ([]SQLExpr, error) {
		switch typedE := expr.(type) {
		case *SQLTupleExpr:
			return typedE.Exprs, nil
		case *SQLValues:
			var exprs []SQLExpr
			for _, value := range typedE.Values {
				exprs = append(exprs, value)
			}
			return exprs, nil
		default:
			return nil, fmt.Errorf("invalid SQLTupleExpr type '%T'", expr)
		}
	}

	leftExprs, err := getExprs(left)
	if err != nil {
		return nil, nil, err
	}

	rightExprs, err := getExprs(right)
	if err != nil {
		return nil, nil, err
	}

	return leftExprs, rightExprs, nil
}

// hasNullValue returns true if any of the value in values
// is of type SQLNoValue or SQLNullValue.
func hasNullValue(values ...SQLValue) bool {
	for _, value := range values {
		switch v := value.(type) {
		case SQLNoValue, SQLNullValue:
			return true
		case *SQLValues:
			if hasNullValue(v.Values...) {
				return true
			}
		}
	}
	return false
}

// hasNullExpr returns true if any of the expr in exprs
// is of type SQLNoValue or SQLNullValue.
func hasNullExpr(exprs ...SQLExpr) bool {
	for _, e := range exprs {
		switch typedE := e.(type) {
		case SQLNoValue, SQLNullValue:
			return true
		case *SQLTupleExpr:
			return hasNullExpr(typedE.Exprs...)
		case *SQLValues:
			return hasNullValue(typedE.Values...)
		}
	}

	return false
}

// IsFalsy returns whether a SQLValue is falsy.
func IsFalsy(value SQLValue) bool {
	return !hasNullValue(value) && !Bool(value)
}

// IsFullyPushedDown returns an error if this PlanStage is not fully optimized and pushed down.
// Otherwise, it returns nil.
func IsFullyPushedDown(plan PlanStage) error {
	// Even if the top-level stage isn't a *MongoSourceStage PlanStage
	// it may still be fully pushed down. e.g. a RowGeneratorStage
	// plan is still fully pushed down.
	if _, ok := plan.(*MongoSourceStage); !ok {
		ok, err := containsMongoSource(plan)
		if err != nil {
			return err
		}

		isFullyOptimized := func(ps PlanStage) bool {
			// case 1. ↳ Project(...) -> ↳ RowGeneratorStage(...) -> ↳ MongoSource
			if pr, ok1 := ps.(*ProjectStage); ok1 {
				if rg, ok2 := pr.source.(*RowGeneratorStage); ok2 {
					if _, ok3 := rg.source.(*MongoSourceStage); ok3 {
						return true
					}
				}
			}
			// For subqueries, we may execute and cache the result for use in an
			// outer query. While the subquery might take a fair bit of time to
			// execute, it is still considered as fully pushed down and thus not
			// whitelisted here.
			return false
		}

		if ok && !isFullyOptimized(plan) {
			return ErrNotFullyPushedDown
		}
		// If we get here, we got a PlanStage that contains a LeafNode (which we
		// do not push down) e.g. a DynamicSourceStage PlanStage.
	}

	return nil
}

// NormalizeUUID takes a UUID's kind and bytes and converts
// the bytes to the standard UUID representation.
func NormalizeUUID(kind schema.MongoType, bytes []byte) error {
	if len(bytes) != 16 {
		return fmt.Errorf("expected UUID bytes to be 16, not %d", len(bytes))
	}

	switch kind {
	case schema.MongoUUID, schema.MongoUUIDOld:
		return nil
	case schema.MongoUUIDCSharp:
		reverseByteArray(bytes, 0, 4)
		reverseByteArray(bytes, 4, 2)
		reverseByteArray(bytes, 6, 2)
	case schema.MongoUUIDJava:
		reverseByteArray(bytes, 0, 8)
		reverseByteArray(bytes, 8, 8)
	default:
		return fmt.Errorf("unrecognized UUID type: %v", kind)
	}
	return nil
}

// reverseByteArray reverses elements in data, beginning
// at start and ending at start + length.
func reverseByteArray(data []byte, start, length int) {
	for left, right := start, start+length-1; left < right; left, right = left+1, right-1 {
		temp := data[left]
		data[left] = data[right]
		data[right] = temp
	}
}

// round founds a float64 to an int64 using MySQL rounding conventions (round
// ties away from 0). This is the simplest implementation of round I have found.
// https://github.com/golang/go/issues/4594#issuecomment-66073312.
func round(x float64) int64 {
	if x < 0 {
		return int64(math.Ceil(x - 0.5))
	}
	return int64(math.Floor(x + 0.5))
}

func shouldFlip(n sqlBinaryNode) bool {
	if _, ok := n.left.(SQLValue); ok {
		if _, ok := n.right.(SQLValue); !ok {
			return true
		}
	}

	return false
}

// twoSum returns the exact unevaluated sum of a and b,
// where the first member is the double nearest the sum
// (ties to even) and the second member is the remainder.
//
// O. Møller. Quasi double-precision in floating-point
// addition. BIT, 5:37–50, 1965.
//
// D. Knuth. The Art of Computer Programming, vol 2.
// Addison-Wesley, Reading, MA, 3rd ed, 1998.
func twoSum(a, b float64) (float64, float64) {
	var s, aPrime, bPrime, deltaA, deltaB, t float64
	s = a + b
	aPrime = s - b
	bPrime = s - aPrime
	deltaA = a - aPrime
	deltaB = b - bPrime
	t = deltaA + deltaB
	return s, t
}

func getSQLInExprs(right SQLExpr) []SQLExpr {
	var exprs []SQLExpr

	// The right child could be a non-SQLValues SQLValue
	// if the tuple can be evaluated and/or simplified. For
	// example in these sorts of cases: (1), (8-7), (date "2005-03-22").
	// The right child could be of type *SQLValues when each of the
	// expressions in the tuple are evaluated to a SQLValue.
	// Finally, it could be of type *SQLTupleExpr when
	// OptimizeExpr yielded no change.
	sqlValue, isSQLValue := right.(SQLValue)
	sqlValues, isSQLValues := right.(*SQLValues)
	sqlTupleExpr, isSQLTupleExpr := right.(*SQLTupleExpr)

	if isSQLValues {
		for _, value := range sqlValues.Values {
			exprs = append(exprs, value.(SQLExpr))
		}
	} else if isSQLValue {
		exprs = []SQLExpr{sqlValue.(SQLExpr)}
	} else if isSQLTupleExpr {
		exprs = sqlTupleExpr.Exprs
	}

	return exprs
}

func translateTupleExpr(leftExpr, rightExpr SQLExpr, op string) (SQLExpr, error) {
	left, right, err := getSQLTupleExprs(leftExpr, rightExpr)
	if err != nil {
		return nil, err
	}

	if len(left) != len(right) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, len(left))
	}

	var constructTupleExpr func(string, []SQLExpr, []SQLExpr, bool) (SQLExpr, error)
	constructTupleExpr = func(op string, left, right []SQLExpr, isEqual bool) (SQLExpr, error) {
		if len(left) == 1 {
			return comparisonExpr(left[0], right[0], op)
		}
		rightChild, err := constructTupleExpr(op, left[1:], right[1:], isEqual)
		if !isEqual {
			return &SQLOrExpr{&SQLNotEqualsExpr{left[0], right[0]}, rightChild}, err
		}
		return &SQLAndExpr{&SQLEqualsExpr{left[0], right[0]}, rightChild}, err
	}

	var translationFunc func(int) (SQLExpr, error)
	translationFunc = func(i int) (SQLExpr, error) {
		if len(left[i:]) == 0 {
			return SQLFalse, nil
		}
		var leftChild SQLExpr
		var err error

		if i == 0 {
			cmpOp := op
			if op == sqlOpLTE {
				cmpOp = sqlOpLT
			} else if op == sqlOpGTE {
				cmpOp = sqlOpGT
			}
			leftChild, err = comparisonExpr(left[0], right[0], cmpOp)
		} else {
			leftChild, err = constructTupleExpr(op, left[:i+1], right[:i+1], true)
		}

		if err != nil {
			return nil, err
		}

		rightChild, err := translationFunc(i + 1)

		return &SQLOrExpr{leftChild, rightChild}, err
	}

	switch op {
	case sqlOpEQ:
		return constructTupleExpr(op, left, right, true)
	case sqlOpNEQ:
		return constructTupleExpr(op, left, right, false)
	default:
		return translationFunc(0)
	}
}

const maxDateParts = 8
const maxHour = 838
const maxMinute = 59
const maxSecond = 59
const twoDigitPartYear = 70
const timeSeparator = ':'

// strToDateTime is a port of mysql's str_to_datetime function.
func strToDateTime(s string, full bool) (time.Time, int, bool) {

	// skip space at start
	var str int
	for str = 0; str < len(s); str++ {
		if !isSpace(s[str]) {
			break
		}
	}

	if str >= len(s) || !isDigit(s[str]) {
		return time.Time{}, 0, false
	}

	const (
		yearIdx        = 0
		monthIdx       = 1
		dayIdx         = 2
		hourIdx        = 3
		minuteIdx      = 4
		secondIdx      = 5
		microsecondIdx = 6
	)

	date := make([]int, maxDateParts)
	dateLengths := make([]int, maxDateParts)
	yearLength := 2
	fieldLength := 2
	internalFormat := false

	// calc number of digits in first part.
	var pos int
	for pos = str; pos < len(s); pos++ {
		if !isDigit(s[pos]) && s[pos] != 'T' {
			break
		}
	}

	dateLengths[yearIdx] = 0
	numDigits := pos - str
	if numDigits == len(s) || s[pos] == '.' {
		// found date in internal format (only numbers like YYYYMMDD)
		if numDigits == 4 || numDigits == 8 || numDigits >= 14 {
			yearLength = 4
			fieldLength = 4
		}
		internalFormat = true
	} else {
		fieldLength = 4
	}

	var state int
	notZeroDate := false
	for state = 0; state < maxDateParts-1 && str < len(s) && isDigit(s[str]); state++ {
		start := str
		tempValue := int(s[str]) - int('0')
		str++

		// gather up all the digits for the current part
		scanUntilDelim := !internalFormat && state != microsecondIdx
		fieldLength--
		for str < len(s) && isDigit(s[str]) && (scanUntilDelim || fieldLength > 0) {
			tempValue = tempValue*10 + (int(s[str]) - int('0'))
			str++
			fieldLength--
		}

		dateLengths[state] = str - start
		if tempValue > 999999 {
			return time.Time{}, 0, false
		}

		date[state] = tempValue
		if tempValue > 0 {
			notZeroDate = true
		}

		// all fields except for year and fractional seconds are of length 2.
		fieldLength = 2

		if str == len(s) {
			state++
			break
		}

		// Allow a 'T' after day to allow CCYYMMDDT type of fields
		if state == dayIdx && s[str] == 'T' {
			str++
			continue
		}

		if state == secondIdx {
			if s[str] == '.' { //followed by a period
				str++
				fieldLength = 6
			}
			continue
		}

		for str < len(s) && (isPunct(s[str]) || isSpace(s[str])) {
			if isSpace(s[str]) {
				if state != dayIdx { // only allow space between date and time
					return time.Time{}, 0, false
				}
			}
			str++
		}

		if state == microsecondIdx {
			state++
		}
	}

	numFields := state
	for state < maxDateParts {
		dateLengths[state] = 0
		date[state] = 0
		state++
	}

	if !internalFormat {
		yearLength = dateLengths[yearIdx]

		if yearLength == 0 {
			return time.Time{}, 0, false
		}
	}

	fractionalLength := dateLengths[microsecondIdx]
	if fractionalLength < 6 {
		date[microsecondIdx] *= int(math.Pow10(6 - fractionalLength))
	}

	if yearLength == 2 && notZeroDate {
		if date[yearIdx] < twoDigitPartYear {
			date[yearIdx] += 2000
		} else {
			date[yearIdx] += 1900
		}
	}

	// If we managed to parse, but minutes or seconds are >= 60
	// MySQL returns NULL for the hour/minute/second function.
	// Rather than return yet another value, we coopt the hour
	// value and return -1, since it will not be needed.
	if numFields < 3 ||
		(numFields <= 3 && full) ||
		(date[yearIdx] == 0 && date[monthIdx] == 0 && date[dayIdx] == 0) ||
		date[yearIdx] > 9999 ||
		date[monthIdx] > 12 ||
		date[dayIdx] > daysInMonth(time.Month(date[monthIdx]), date[yearIdx]) ||
		date[hourIdx] > 23 || date[monthIdx] > 59 || date[secondIdx] > 59 {
		return time.Time{}, -1, false
	}

	return time.Date(date[yearIdx],
			time.Month(date[monthIdx]),
			date[dayIdx],
			date[hourIdx],
			date[minuteIdx],
			date[secondIdx],
			date[microsecondIdx]*1000,
			schema.DefaultLocale),
		date[hourIdx],
		true
}

func daysInMonth(m time.Month, year int) int {
	// This is equivalent to time.daysIn(m, year).
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// strToTime is a port of mysql's str_to_time function.
// We also return the hour as an int because MySQL return hour values
// up to 838, and the time.Duration stores hours modulo 24.
func strToTime(s string) (time.Duration, int, bool) {
	parts := make([]int, 5)
	const (
		dayIdx         = 0
		hourIdx        = 1
		minuteIdx      = 2
		secondIdx      = 3
		microsecondIdx = 4
	)

	negative := false
	var state int
	str := 0
	for ; str < len(s); str++ {
		if !isSpace(s[str]) {
			break
		}
	}

	if str < len(s) && str == '-' {
		negative = true
	}

	if str == len(s) {
		return time.Duration(0), 0, false
	}

	value := 0
	for ; str < len(s) && isDigit(s[str]); str++ {
		value = value*10 + int((s[str] - '0'))
	}

	endOfDays := str

	for ; str < len(s); str++ {
		if !isSpace(s[str]) {
			break
		}
	}

	foundDays := false
	foundHours := false
	if str+1 < len(s) && str != endOfDays && isDigit(s[str]) {
		parts[dayIdx] = value
		state = hourIdx
		foundDays = true
	} else if str+1 < len(s) && s[str] == timeSeparator && isDigit(s[str+1]) {
		parts[hourIdx] = value
		state = minuteIdx
		foundHours = true
		str++
	} else {
		parts[hourIdx] = value / 10000
		parts[minuteIdx] = value / 100 % 100
		parts[secondIdx] = value % 100
		state = secondIdx
	}

	if state != secondIdx {
		for {
			for value = 0; str < len(s) && isDigit(s[str]); str++ {
				value = value*10 + int(s[str]-'0')
			}

			parts[state] = value
			state++
			if state == microsecondIdx ||
				(len(s)-str) < 2 ||
				s[str] != timeSeparator ||
				!isDigit(s[str+1]) {
				break
			}
			str++
		}

		if state != secondIdx {
			if !foundDays && !foundHours {
				parts[microsecondIdx] = parts[minuteIdx]
				parts[secondIdx] = parts[hourIdx]
				parts[minuteIdx] = parts[dayIdx]
			}
		}
	}

	if str+1 < len(s) && s[str] == '.' && isDigit(s[str+1]) {
		str++
		value = 0
		fieldLength := 0
		for ; str < len(s) && isDigit(s[str]) && fieldLength < 6; str++ {
			value = value*10 + int(s[str]-'0')
			fieldLength++
		}

		parts[microsecondIdx] = value * int(math.Pow10(6-fieldLength))
	}

	// garbage at the end...
	if str != len(s) {
		return time.Duration(0), 0, false
	}

	// If we managed to parse, but minutes or seconds are >= 60
	// MySQL returns NULL for the hour/minute/second function.
	// Rather than return yet another value, we coop the hour
	// value and return -1, since it will not be needed.
	if parts[minuteIdx] >= 60 || parts[secondIdx] >= 60 {
		return time.Duration(0), -1, false
	}

	hour := parts[dayIdx]*24 + parts[hourIdx]
	result := time.Duration(hour)*time.Hour +
		time.Duration(parts[minuteIdx])*time.Minute +
		time.Duration(parts[secondIdx])*time.Second +
		time.Duration(parts[microsecondIdx])*time.Microsecond
	if negative {
		result = -result
	}

	returnHour := hour
	if hour > maxHour {
		returnHour = maxHour
	}

	if hour <= maxHour && (hour != maxHour || parts[minuteIdx] != maxMinute ||
		parts[secondIdx] != maxSecond || parts[microsecondIdx] > 0) {
		return result, returnHour, true
	}

	// out of range... usually would add a warning
	return time.Duration(maxHour)*time.Hour +
		time.Duration(maxMinute)*time.Minute +
		time.Duration(maxSecond)*time.Second, returnHour, true
}

func isDigit(c byte) bool {
	switch c {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	default:
		return false
	}
}

var idTypes = map[schema.MongoType]struct{}{
	schema.MongoObjectID:   {},
	schema.MongoUUID:       {},
	schema.MongoUUIDCSharp: {},
	schema.MongoUUIDJava:   {},
	schema.MongoUUIDOld:    {},
}

var uuidTypes = map[schema.MongoType]struct{}{
	schema.MongoUUID:       {},
	schema.MongoUUIDCSharp: {},
	schema.MongoUUIDJava:   {},
	schema.MongoUUIDOld:    {},
}

// IsUUID returns true if mongoType is of the UUID subtype.
func IsUUID(mongoType schema.MongoType) bool {
	_, ok := uuidTypes[mongoType]
	return ok
}

// isIDType returns true if mongoType is a UUID or an ObjectID.
func isIDType(mongoType schema.MongoType) bool {
	_, ok := idTypes[mongoType]
	return ok
}

func isLeapYear(y int) bool {
	return (y%4 == 0) && (y%100 != 0) || (y%400 == 0)
}

func isPunct(c byte) bool {
	return strings.IndexByte(punctuation, c) != -1
}

func isSpace(c byte) bool {
	switch c {
	case ' ':
		return true
	default:
		return false
	}
}

// databaseFromPlanStage returns the database name from columns returned from the planStage.
// It returns the empty string if the columns come from more than one database or the dual database.
func databaseFromPlanStage(plan PlanStage) string {
	dbName := ""
	for _, column := range plan.Columns() {
		if dbName == "" {
			dbName = column.Database
		} else if column.Database != dbName {
			dbName = ""
			break
		}
	}
	return dbName
}

func uuidEncode(data []byte) string {
	return hex.EncodeToString(data[0:4]) + "-" +
		hex.EncodeToString(data[4:6]) + "-" +
		hex.EncodeToString(data[6:8]) + "-" +
		hex.EncodeToString(data[8:10]) + "-" +
		hex.EncodeToString(data[10:16])
}

// GoValueToSQLValue is only needed for dynamic sources and reading variables
// and a few places in testing. As the name suggests, it converts a go value
// to a SQLValue.
func GoValueToSQLValue(v interface{}) SQLValue {
	switch vTyped := v.(type) {
	case nil:
		return SQLNull
	case bool:
		if vTyped {
			return SQLBool(1.0)
		}
		return SQLBool(0.0)
	case int:
		return SQLInt64(vTyped)
	case int64:
		return SQLInt64(vTyped)
	case float64:
		return SQLFloat(vTyped)
	case uint16:
		return SQLUint64(vTyped)
	case uint32:
		return SQLUint64(vTyped)
	case uint64:
		return SQLUint64(vTyped)
	case string:
		return SQLVarchar(vTyped)
	case variable.Name:
		return SQLVarchar(vTyped)
	default:
		panic(fmt.Sprintf(
			"unexpected go type %T from dynamic source or system variable in GoValueToSQLValue",
			vTyped))
	}
}
