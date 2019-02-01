package evaluator

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"

	"github.com/shopspring/decimal"
)

const (
	regexCharsToEscape = ".^$*+?()[{\\|"
	maxPrecisionInt    = int64(1 << 53)
	punctuation        = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
)

var (
	// ErrNotFullyPushedDown is the error returned when a query that hits MongoDB isn't fully
	// pushed down and the full_pushdown_exec_mode system variable is set.
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
	pattern := String(value)
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

func getPlanStats(plan PlanStage, pCfg *PushdownConfig) (*PlanStats, error) {
	pushedDown := IsFullyPushedDown(plan) == nil

	explain, err := explainQuery(plan, pCfg)
	if err != nil {
		return nil, err
	}

	stats := &PlanStats{
		FullyPushedDown: pushedDown,
		Explain:         explain,
	}
	return stats, nil
}

// hasNullValue returns true if any of the value in values
// is of type SQLNoValue or SQLNullValue.
func hasNullValue(values ...SQLValue) bool {
	for _, v := range values {
		if v.IsNull() {
			return true
		}
	}
	return false
}

// hasNullExpr returns true if any of the expr in exprs
// is of type SQLNoValue or SQLNullValue.
func hasNullExpr(exprs ...SQLExpr) bool {
	for _, e := range exprs {
		switch typedE := e.(type) {
		case SQLValue:
			if typedE.IsNull() {
				return true
			}
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

// getMongoDBVersion is a utility function that gets the MongoDB version to use for new
// configurations based on the mongodb_version_compatibility or mongodb_version variable
// in the provided container.
func getMongoDBVersion(vars catalog.VariableContainer) []uint8 {
	compatibilityVersion := vars.GetString(variable.MongoDBVersionCompatibility)
	if len(compatibilityVersion) == 0 {
		compatibilityVersion = vars.GetString(variable.MongoDBVersion)
	}
	version, err := procutil.VersionToSlice(compatibilityVersion)
	if err != nil {
		panic(fmt.Sprintf("invalid version provided: %v", compatibilityVersion))
	}
	return version
}

func getMySQLVersion(vars catalog.VariableContainer) string {
	return vars.GetString(variable.Version)
}

// GoValueToSQLValue is only needed for dynamic sources and reading variables
// and a few places in testing. As the name suggests, it converts a go value
// to a SQLValue.
func GoValueToSQLValue(kind SQLValueKind, v interface{}) SQLValue {
	switch vTyped := v.(type) {
	case nil:
		return NewPolymorphicSQLNull(kind)
	case bool:
		return NewSQLBool(kind, vTyped)
	case int:
		return NewSQLInt64(kind, int64(vTyped))
	case int64:
		return NewSQLInt64(kind, vTyped)
	case float64:
		return NewSQLFloat(kind, vTyped)
	case uint16:
		return NewSQLUint64(kind, uint64(vTyped))
	case uint32:
		return NewSQLUint64(kind, uint64(vTyped))
	case uint64:
		return NewSQLUint64(kind, vTyped)
	case string:
		return NewSQLVarchar(kind, vTyped)
	case variable.Name:
		return NewSQLVarchar(kind, string(vTyped))
	default:
		panic(fmt.Sprintf(
			"unexpected go type %T from dynamic source or system variable in GoValueToSQLValue",
			vTyped))
	}
}

func paddedDateString(val SQLValue) (string, bool) {
	switch val.(type) {
	case SQLFloat, SQLDecimal128, SQLInt64:
		noDecimal := strings.Split(val.String(), ".")[0]

		intLength := len(noDecimal)
		if intLength > 14 {
			return "", false
		}

		padLen := 0
		switch intLength {
		case 5, 7, 11, 13:
			padLen = 1
		case 3, 4:
			padLen = 6 - intLength
		case 9, 10:
			padLen = 12 - intLength
		}

		str := strings.Repeat("0", padLen) + noDecimal
		return str, true
	}

	panic(fmt.Errorf("paddedDateString cannot be called with argument of type %T", val))
}

// panicIfNotPlanStage returns a PlanStage from a Node n, or panics if the Node is not a PlanStage.
func panicIfNotPlanStage(s string, n Node) PlanStage {
	ret, ok := n.(PlanStage)
	if ok {
		return ret
	}
	panic(fmt.Sprintf("attempted to convert Node %v to PlanStage in ReplaceChild for %s", n, s))
}

// panicIfNotSQLExpr returns a SQLExpr from a Node n, or panics if the Node is not a SQLExpr.
func panicIfNotSQLExpr(s string, n Node) SQLExpr {
	ret, ok := n.(SQLExpr)
	if ok {
		return ret
	}
	panic(fmt.Sprintf("attempted to convert Node %v to SQLExpr in ReplaceChild for %s", n, s))
}

// panicIfNotProjectStage returns a *ProjectStage from a PlanStage p, or panics if the PlanStage is not a *ProjectStage.
func panicIfNotProjectStage(side string, p PlanStage) *ProjectStage {
	ret, ok := p.(*ProjectStage)
	if ok {
		return ret
	}
	panic(fmt.Sprintf("expected ProjectStage for %s PlanStage, got %T", side, p))
}

// panicWithInvalidIndex formats the panics for ReplaceChild methods.
func panicWithInvalidIndex(s string, index, max int) {
	if max < 0 {
		panic(fmt.Sprintf("%v requested ReplaceChild of index %d, but has no children", s, index))
	}
	panic(fmt.Sprintf("%v requested ReplaceChild of index %d (has max index of %d)", s, index, max))
}

// nolint: unparam
func parseDateTime(s string) (time.Time, int, bool) {
	return strToDateTime(s, false)
}

func parseTime(s string) (time.Time, int, bool) {

	timeParts := strings.Split(s, ".")
	// Truncate extra decimals, e.g.: "26:11:59.23.24.25"
	// should be treated as "26:11:59.23".
	if len(timeParts) > 1 {
		s = strings.Join(timeParts[0:2], ".")
	}
	noFractions := timeParts[0]
	if len(noFractions) >= 12 {

		// Probably a datetime.
		dt, hour, ok := strToDateTime(s, true)
		if ok {
			return dt, hour, true
		}
	}

	// The result will be 0 if parsing failed, so we don't care about the result.
	dur, hour, ok := strToTime(s)

	return time.Date(0, 1, 1, 0, 0, 0, 0, schema.DefaultLocale).Add(dur), hour, ok
}

func parseDuration(v SQLValue) (time.Duration, bool) {
	buf := []byte(v.String())

	var h, m, s, f int

	hours, mins, secs, frac := []byte{}, []byte{}, []byte{}, []byte{}

	emitFrac := func(buf []byte) int {
		i := bytes.IndexByte(buf, '.')
		if i != -1 && len(frac) == 0 {
			x := 0
			for x < len(buf)-i-1 {
				idx := i + x + 1
				if buf[idx] == ':' || buf[idx] == '.' {
					break
				}
				x++
			}
			frac = buf[i+1 : i+x+1]
		}
		return i
	}

	// nolint: unparam
	emitToken := func(buf []byte, v byte) int {
		w, l := 0, len(buf)-1
		for w < l {
			if buf[w] == v {
				break
			}
			w++
		}
		return w
	}

	fmtNumeric := func(buf []byte) []byte {
		x, l, w := -1, len(buf), len(buf)
		i := bytes.IndexByte(buf, '.')
		tmp := make([]byte, w+2)
		if i != -1 {
			w = i + 2
			copy(tmp[w:], buf[i:])
		} else {
			i, w = l, l+2
		}

		for w > 0 && i > 0 {
			w, x, i = w-1, x+1, i-1
			if x%2 == 0 && x > 0 {
				tmp[w], w = ':', w-1
			}
			tmp[w] = buf[i]
			if x == 4 {
				break
			}
		}

		return append(buf[:i], tmp[w:]...)
	}

	if bytes.IndexByte(buf, ':') == -1 {
		buf = fmtNumeric(buf)
	}

	h = emitToken(buf, ':')

	if h != 0 {
		hours, buf, f = buf[0:h], buf[h+1:], emitFrac(buf[0:h+1])
		if f != -1 {
			secs, hours = hours[:f], hours[:0]
		} else {
			m = emitToken(buf, ':')
			if m != 0 {
				mins, buf, f = buf[0:m], buf[m+1:], emitFrac(buf[0:m+1])
				if f != -1 {
					mins = mins[:f]
				} else {
					s = emitToken(buf, ':')
					if s != 0 {
						secs, f = buf[0:s], emitFrac(buf[0:s+1])
						if f != -1 {
							secs = secs[:f]
						} else {
							secs = buf
						}
					}
				}
			}
		}
	}

	if len(mins) > 0 {
		if m, err := strconv.Atoi(string(mins)); err != nil || m > 60 {
			return 0, false
		}
	}

	if len(secs) > 0 {
		if s, err := strconv.ParseFloat(string(secs), 64); err != nil || s > 60 {
			return 0, false
		}
	}

	str := ""

	if len(hours) != 0 {
		str = fmt.Sprintf("%vh", string(hours))
	}

	if len(mins) != 0 {
		str = fmt.Sprintf("%v%vm", str, string(mins))
	}

	switch len(secs) {
	case 0:
		if len(frac) != 0 {
			str = fmt.Sprintf("%v0.%vs", str, string(frac))
		}
	default:
		if len(frac) != 0 {
			str = fmt.Sprintf("%v%v.%vs", str, string(secs), string(frac))
		} else {
			str = fmt.Sprintf("%v%vs", str, string(secs))
		}
	}

	dur, err := time.ParseDuration(str)
	if err != nil {
		return 0, false
	}

	return dur, true
}

// roundToDecimalPlaces rounds base to d number of decimal places.
func roundToDecimalPlaces(d int64, base float64) float64 {
	var rounded float64
	pow := math.Pow(10, float64(d))
	digit := pow * base
	_, div := math.Modf(digit)
	if base > 0 {
		if div >= 0.5 {
			rounded = math.Ceil(digit) / pow
		} else {
			rounded = math.Floor(digit) / pow
		}
	} else {
		if math.Abs(div) >= 0.5 {
			rounded = math.Floor(digit) / pow
		} else {
			rounded = math.Ceil(digit) / pow
		}
	}
	return rounded
}

func runesIndex(r, sep []rune) int {
	for i := 0; i <= len(r)-len(sep); i++ {
		found := true
		for j := 0; j < len(sep); j++ {
			if r[i+j] != sep[j] {
				found = false
				break
			}
		}

		if found {
			return i
		}
	}

	return -1
}

func baseIsInvalid(base int64) bool {
	if base < 2 || base > 36 {
		return true
	}

	return false
}

func evalTypeFromSQLExpr(expr SQLExpr) (EvalType, bool) {
	val, ok := expr.(SQLValue)
	if !ok {
		panic("argument to evalTypeFromSQLExpr must be a SQLValue representing a viable convert target type")
	}

	var typ EvalType
	switch val.String() {
	case string(parser.SIGNED_BYTES):
		typ = EvalInt64
	case string(parser.UNSIGNED_BYTES):
		typ = EvalUint64
	case string(parser.FLOAT_BYTES):
		typ = EvalDouble
	case string(parser.CHAR_BYTES):
		typ = EvalString
	case string(parser.OBJECT_ID_BYTES):
		typ = EvalObjectID
	case string(parser.DATE_BYTES):
		typ = EvalDate
	case string(parser.DATETIME_BYTES):
		typ = EvalDatetime
	case string(parser.DECIMAL_BYTES):
		typ = EvalDecimal128
	case string(parser.BINARY_BYTES):
		// Although we represent binary as a string, conversions
		// to it are always going to be invalid.
		return EvalString, false
	case string(parser.TIME_BYTES):
		// We do not support the TIME type yet. Just use EvalDatetime
		// for now.
		return EvalDatetime, false
	default:
		panic(fmt.Errorf("invalid value %q", val.String()))
	}

	return typ, true
}

// formatDate takes a time.Time object and outputs a string formatted using
// MySQL's format string specification.
func (f *baseScalarFunctionExpr) formatDate(sqlValueKind SQLValueKind,
	collation *collation.Collation, date time.Time, format string) (string, error) {
	formatRunes := []rune(format)

	noPad := func(s string) (string, error) {
		str := date.Format(s)
		if len(str) == 2 && str[0] == '0' {
			str = str[1:]
		}
		return str, nil
	}

	suffixFmt := func(i int) (string, error) {
		formatted := date.Format(strconv.Itoa(i))
		i, err := strconv.Atoi(formatted)
		if err != nil {
			return "", err
		}
		suffix := "th"
		switch i % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		}
		return formatted + suffix, nil
	}

	weekFmt := func(i int64) (string, error) {
		args := []SQLValue{NewSQLDate(sqlValueKind, date), NewSQLInt64(sqlValueKind, i)}
		eval, err := f.weekEvaluate(sqlValueKind, collation, args)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%02v", eval.String()), nil
	}

	yearFmt := func(i int64) (string, error) {
		args := []SQLValue{NewSQLDate(sqlValueKind, date), NewSQLInt64(sqlValueKind, i)}
		eval, err := f.yearWeekEvaluate(sqlValueKind, collation, args)
		if err != nil {
			return "", err
		}
		return eval.String()[:4], nil
	}

	zeroPad := func(s string) (string, error) {
		return fmt.Sprintf("%02v", date.Format(s)), nil
	}

	fmtTokens := map[rune]string{
		'a': "Mon",
		'b': "Jan",
		'c': "1",
		'e': "2",
		'i': "04",
		'l': "3",
		'M': "January",
		'm': "01",
		'p': "PM",
		'r': "03:04:05 PM",
		'S': "05",
		's': "05",
		'T': "15:04:05",
		'W': "Monday",
		'Y': "2006",
		'y': "06",
	}

	formatters := map[rune]func() (string, error){
		'D': func() (string, error) { return suffixFmt(2) },
		'd': func() (string, error) { return zeroPad("2") },
		'f': func() (string, error) { return date.Format(".000000")[1:], nil },
		'H': func() (string, error) { return zeroPad("15") },
		'h': func() (string, error) { return zeroPad("3") },
		'I': func() (string, error) { return zeroPad("3") },
		'j': func() (string, error) { return fmt.Sprintf("%03v", date.YearDay()), nil },
		'k': func() (string, error) { return noPad("15") },
		'U': func() (string, error) { return weekFmt(0) },
		'u': func() (string, error) { return weekFmt(1) },
		'V': func() (string, error) { return weekFmt(2) },
		'v': func() (string, error) { return weekFmt(3) },
		'w': func() (string, error) { return strconv.Itoa(int(date.Weekday())), nil },
		'X': func() (string, error) { return yearFmt(0) },
		'x': func() (string, error) { return yearFmt(1) },
		'%': func() (string, error) { return "%", nil },
	}

	for k, v := range fmtTokens {
		localV := v
		formatters[k] = func() (string, error) {
			return date.Format(localV), nil
		}
	}

	var result string
	for i := 0; i < len(formatRunes); i++ {
		if formatRunes[i] == '%' && i != len(formatRunes)-1 {
			if formatter, ok := formatters[formatRunes[i+1]]; ok {
				s, err := formatter()
				if err != nil {
					return "", err
				}
				result += s
				i++
			} else {
				result += string(formatRunes[i])
			}
		} else {
			result += string(formatRunes[i])
		}
	}

	return result, nil
}

// areAllTimeTypes checks if all SQLValues are either type SQLTimestamp or
// SQLDate and there is at least one SQLTimestamp type. This is necessary
// because if the former is true, MySQL will always return a SQLTimestamp type
// in the greatest and least functions. i.e. SELECT GREATEST(DATE
// "2006-05-11", TIMESTAMP "2005-04-12", DATE "2004-06-04") returns TIMESTAMP
// "2006-05-11 00:00:00"
func areAllTimeTypes(values []SQLValue) (bool, bool) {
	allTimeTypes := true
	timestamp := false
	for _, v := range values {
		if _, ok := v.(SQLTimestamp); !ok {
			if _, ok := v.(SQLDate); !ok {
				allTimeTypes = false
			}
		} else {
			timestamp = true
		}
	}
	return allTimeTypes, timestamp
}

// handlePadding is used by the lpad and rpad functions. creates the
// specified padding string and pads the original string. padding
// goes on the left side if isLeftPad = true, on the right side otherwise.
func handlePadding(kind SQLValueKind, values []SQLValue, isLeftPad bool) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(kind, EvalString), nil
	}

	var length int
	// length should be converted to float before we get to here
	if floatLength := Float64(values[1]); floatLength < float64(0) {
		length = int(floatLength - 0.5)
	} else {
		length = int(floatLength + 0.5)
	}

	str := []rune(values[0].String())
	padStr := []rune(values[2].String())
	padLen := length - len(str)

	// either:
	// 1) padding string is empty and the input string is not long enough to not need padding
	// 2) output length is negative and therefore impossible
	if (len(padStr) == 0 && len(str) < length) || length < 0 {
		return NewSQLNull(kind, EvalString), nil
	}

	// the string is already long enough
	if len(str) >= length {
		return NewSQLVarchar(kind, string(str[:length])), nil
	}

	// repeat padding as many times as needed to fill room
	numRepeats := math.Ceil(float64(padLen) / float64(len(padStr)))

	padding := []rune(strings.Repeat(string(padStr), int(numRepeats)))

	// in case room % len(padstr) != 0, chop off end
	padding = padding[:padLen]

	finalPad := string(padding)
	finalStr := string(str)

	if isLeftPad {
		return NewSQLVarchar(kind, finalPad+finalStr), nil
	}

	return NewSQLVarchar(kind, finalStr+finalPad), nil
}

func numMonths(startDate time.Time, endDate time.Time) int {
	y1, m1, d1 := startDate.Date()
	y2, m2, d2 := endDate.Date()
	months := ((y2 - y1) * 12) + (int(m2) - int(m1))
	if endDate.After(startDate) {
		if d2 < d1 {
			months--
		} else if d1 == d2 {
			h1, mn1, s1 := startDate.Clock()
			ns1 := startDate.Nanosecond()
			h2, mn2, s2 := endDate.Clock()
			ns2 := endDate.Nanosecond()
			t1 := time.Date(y1, m1, d1, h1, mn1, s1, ns1, schema.DefaultLocale)
			t2 := time.Date(y1, m1, d1, h2, mn2, s2, ns2, schema.DefaultLocale)
			if t1.After(t2) {
				months--
			}
		}
	} else {
		if d1 < d2 {
			months++
		} else if d1 == d2 {
			h1, mn1, s1 := startDate.Clock()
			ns1 := startDate.Nanosecond()
			h2, mn2, s2 := endDate.Clock()
			ns2 := endDate.Nanosecond()
			t1 := time.Date(y1, m1, d1, h1, mn1, s1, ns1, schema.DefaultLocale)
			t2 := time.Date(y1, m1, d1, h2, mn2, s2, ns2, schema.DefaultLocale)
			if t2.After(t1) {
				months++
			}
		}
	}
	return months
}

// daysFromYearZeroCalculation calculates the number of days
// between the given year and date 0 - "0000-00-00".
func daysFromYearZeroCalculation(date time.Time) (float64, error) {
	year := date.Year()
	if year > len(yearZeroDayDifferenceSlice)-1 {
		return 0, fmt.Errorf("invalid year in date: %v", year)
	}

	// Zero out any time parts of the date.
	date = time.Date(year, date.Month(), date.Day(), 0, 0, 0, 0, schema.DefaultLocale)
	dateYearStart := time.Date(year, time.January, 1, 0, 0, 0, 0, schema.DefaultLocale)

	dayDifference := yearZeroDayDifferenceSlice[year]
	// Now add the remaining days not accounted for by the year difference.
	// The difference between "0000-01-01" and "0000-00-00" is 1 day.
	if year == 0 && date.Equal(dateYearStart) {
		dayDifference += 1
	} else {
		dayDifference += math.Trunc(date.Sub(dateYearStart).Hours() / 24.0)
	}
	return dayDifference, nil
}

// weekCalculation calculates the week for a given date and mode in memory.
// It is used by both the WEEK and YEARWEEK mysql scalar functions.
// Returns -1 on error. Callers should check for -1 and return proper
// default value (likely SQLNull).
func weekCalculation(date time.Time, mode int) int {

	// zeroCheck replaces results of week 0 with the week for (year-1)-12-31 for modes that
	// are 1-53 only. That means that in 1-53 modes, certain dates at the beginning of the year
	// map to week 52 or 53 of the previous year.
	zeroCheck := func(date time.Time, output, mode int) int {
		if output == 0 {
			return weekCalculation(time.Date(date.Year()-1,
				12,
				31,
				0,
				0,
				0,
				0,
				schema.DefaultLocale),
				mode)
		}
		return output
	}

	// fiftyThreeCheck is used to handle cases where the last week of a
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
	// unlike MongoDB aggregation pipeline, which numbers days 1-7,
	// go time.Time numbers days 0-6, with 0 being Sunday.
	fiftyThreeCheck := func(date time.Time, output int, janOneDaysOfWeek ...int) int {
		if output == 53 {
			day := date.Day()
			nextJanOne := time.Date(date.Year()+1, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
			nextJanOneDayOfWeek := int(nextJanOne.Weekday())
			switch nextJanOneDayOfWeek {
			case janOneDaysOfWeek[0]:
				if day >= 29 {
					output = 1
				}
			case janOneDaysOfWeek[1]:
				if day >= 30 {
					output = 1
				}
			case janOneDaysOfWeek[2]:
				if day >= 31 {
					output = 1
				}
			}
		}
		return output
	}

	// computeDaySubtract computes the main week calculation shared by everything.
	// The calculation is:
	// trunc((date - dayOne) / (7 * millisecondsPerDay) + 1).
	computeDaySubtract := func(date, dayOne time.Time) int {
		return int(float64(date.Sub(dayOne))/
			(7.0*float64(millisecondsPerDay)*float64(time.Millisecond)) +
			1.0)
	}

	// computeDayInYear sets up dayOne for modes where the first week is defined
	// by having Sunday (1) or Monday (2) in the year, and computes the subtraction.
	// these modes are 0, 2, 5, 7.
	computeDayInYear := func(date time.Time, startDay, dayOfWeek int) int {
		// These are more simple than the 4 days mode. The diff from JanOne
		// can be defined using (7 - x + startDay) % 7.
		// This differs slightly from pushdown because MongoDB uses 1-7 for Sunday-Saturday
		// while go uses 0-6.
		dayOne := time.Date(date.Year(), 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
		diff := (7 - dayOfWeek + startDay) % 7
		dayOne = dayOne.Add(time.Duration(diff * int(time.Hour) * 24))
		return computeDaySubtract(date, dayOne)
	}

	// compute4DaysInYear sets up dayOne for modes where the first
	// week is defined by having 4 days in the year and computes the subtraction,
	// these are modes 1, 3, 4, and 6.
	compute4DaysInYear := func(date time.Time, startDay, dayOfWeek int) int {
		// This description is used for Monday as first day of the
		// week. See below for an explanation of the Sunday first day
		// case. Calculate the first day of the first week of this
		// year based on the dayOfWeek of YYYY-01-01 of this year, note
		// that it may be from the previous year. The Day Diff column
		// is the
		// amount of days to Add or Subtract from YYYY-01-01:
		// Day Of the Week Jan 1   |   Day Diff
		// ---------------------------------------------
		//                     0   |   + 1
		//                     1   |   + 0
		//                     2   |   - 1
		//                     3   |   - 2
		//                     4   |   - 3
		//                     5   |   + 3
		//                     6   |   + 2
		// For Sunday, we can see that 0 should be + 0, and the rest follow as expected.
		// Thus we can just add startDay since it is 0 for Sunday and 1 for Monday.
		dayOne := time.Date(date.Year(), 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
		diff := -dayOfWeek + startDay
		if diff < -3 {
			diff += 7
		}
		dayOne = dayOne.Add(time.Duration(diff * int(time.Hour) * 24))
		return computeDaySubtract(date, dayOne)
	}

	jan1 := time.Date(date.Year(), 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
	jan1DayInWeek := int(jan1.Weekday())
	switch mode {
	// First day of week: Sunday, with a Sunday in this year.
	case 0, 2:
		output := computeDayInYear(date, 0, jan1DayInWeek)
		if mode == 2 {
			output = zeroCheck(date, output, 0)
		}
		return output
	// First day of week: Monday, with 4 days in this year.
	case 1, 3:
		output := compute4DaysInYear(date, 1, jan1DayInWeek)
		if mode == 3 {
			output = zeroCheck(date, output, 1)
			output = fiftyThreeCheck(date, output, 4, 3, 2)
		}
		return output
	// First day of week: Sunday, with 4 days in this year.
	case 4, 6:
		output := compute4DaysInYear(date, 0, jan1DayInWeek)
		if mode == 6 {
			output = zeroCheck(date, output, 4)
			output = fiftyThreeCheck(date, output, 3, 2, 1)
		}
		return output
	// First day of week: Monday, with a Monday in this year.
	case 5, 7:
		output := computeDayInYear(date, 1, jan1DayInWeek)
		if mode == 7 {
			output = zeroCheck(date, output, 5)
		}
		return output
	}
	return -1
}

// toRadians converts the provided value (in degrees) to radians.
func toRadians(f float64) float64 {
	return f * math.Pi / 180
}

// toDegrees converts the provided value (in radians) to degrees.
func toDegrees(f float64) float64 {
	return f * 180 / math.Pi
}

// dateArithmeticArgs parses val and returns an integer slice stripped of any
// spaces, colons, etc. It also returns whether the first character in val is
// "-", indicating whether the arguments should be negative.
func dateArithmeticArgs(unit string, val SQLValue) ([]int, int) {
	var args []int
	neg := 1
	prev := -1
	curr := ""
	for idx, char := range val.String() {
		if idx == 0 && char == 45 {
			neg = -1
		}
		if char >= 48 && char <= 57 {
			if prev >= 48 && char <= 57 {
				curr += string(char)
			} else {
				curr = string(char)
			}
			prev = int(char)
		} else if prev != -1 {
			c, _ := strconv.Atoi(curr)
			args = append(args, c)
			curr = ""
			prev = int(char)
		}
	}
	if unit != Microsecond && strings.HasSuffix(unit, Microsecond) {
		curr = curr + strings.Repeat("0", 6-len(curr))
	}
	c, _ := strconv.Atoi(curr)
	args = append(args, c)
	return args, neg
}

// calculateInterval converts each of the values in args to unit, and returns
// the sum of these multiplied by neg.
func calculateInterval(unit string, args []int, neg int) (string, int, error) {
	var val int
	var u string
	sp := strings.SplitAfter(unit, "_")
	if len(sp) > 1 {
		u = sp[1]
	} else {
		u = unit
	}

	const day int = 24
	const hour int = 60
	const minute int = 60
	const second int = 1000000

	switch len(args) {
	case 5:
		switch unit {
		case DayMicrosecond:
			val = args[0]*day*hour*minute*second +
				args[1]*hour*minute*second +
				args[2]*minute*second +
				args[3]*second +
				args[4]
		default:
			return unit, 0, fmt.Errorf("invalid argument length")
		}
	case 4:
		switch unit {
		case DayMicrosecond, HourMicrosecond:
			val = args[0]*hour*minute*second +
				args[1]*minute*second +
				args[2]*second +
				args[3]
		case DaySecond:
			val = args[0]*day*hour*minute +
				args[1]*hour*minute +
				args[2]*minute +
				args[3]
		default:
			return unit, 0, fmt.Errorf("invalid argument length")
		}
	case 3:
		switch unit {
		case DayMicrosecond, HourMicrosecond, MinuteMicrosecond:
			val = args[0]*minute*second + args[1]*second + args[2]
		case DaySecond, HourSecond:
			val = args[0]*hour*minute + args[1]*minute + args[2]
		case DayMinute:
			val = args[0]*day*hour + args[1]*hour + args[2]
		default:
			return unit, 0, fmt.Errorf("invalid argument length")
		}
	case 2:
		switch unit {
		case DayMicrosecond, HourMicrosecond, MinuteMicrosecond, SecondMicrosecond:
			val = args[0]*second + args[1]
		case DaySecond, HourSecond, MinuteSecond:
			val = args[0]*minute + args[1]
		case DayMinute, HourMinute:
			val = args[0]*hour + args[1]
		case DayHour:
			val = args[0]*day + args[1]
		case YearMonth:
			val = args[0]*12 + args[1]
		default:
			return unit, 0, fmt.Errorf("invalid argument length")
		}
	case 1:
		val = args[0]
	default:
		return unit, 0, fmt.Errorf("invalid argument length")
	}

	return u, val * neg, nil
}

func shouldFlip(n sqlBinaryNode) bool {
	if _, ok := n.left.(SQLValue); ok {
		if _, ok := n.right.(SQLValue); !ok {
			return true
		}
	}

	return false
}

func unitIntervalToMilliseconds(unit string, interval int64) (int64, error) {
	switch unit {
	case Day:
		return interval * 24 * 60 * 60 * 1000, nil
	case Hour:
		return interval * 60 * 60 * 1000, nil
	case Minute:
		return interval * 60 * 1000, nil
	case Second:
		return interval * 1000, nil
	case Microsecond:
		return interval / 1000, nil
	default:
		return 0, fmt.Errorf("cannot compute milliseconds for the unit %v", unit)
	}
}

func nodesToExprs(nodes []Node) []SQLExpr {
	var ok bool
	ret := make([]SQLExpr, len(nodes))
	for i := range nodes {
		ret[i], ok = nodes[i].(SQLExpr)
		if !ok {
			panic(fmt.Sprintf("non-SQLExpr %v: %T found in nodesToExprs", nodes[i], nodes[i]))
		}
	}
	return ret
}

func isBooleanColumnAndNumber(left, right SQLExpr) bool {
	if _, ok := left.(SQLColumnExpr); !ok {
		return false
	}

	if left.EvalType() != EvalBoolean {
		return false
	}

	if _, ok := right.(SQLNumber); !ok {
		return false
	}

	if _, ok := right.(SQLBool); ok {
		return false
	}

	return true
}
