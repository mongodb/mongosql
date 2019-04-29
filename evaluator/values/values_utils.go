package values

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/mathutil"
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/schema"
)

const (
	maxDateParts     = 8
	twoDigitPartYear = 70
)

var (
	// slashDelimitedDDMMYYYYDateFormat matches DD/MM/YYYY date formats.
	// It is used for evaluating dates in mongosql conversion mode. See
	// ParseDateTimeMongo below for more details.
	slashDelimitedDDMMYYYYDateFormat = regexp.MustCompile(`^(\d{1,2}/\d{1,2}/\d{1,4})`)

	// validTimeFormat matches valid time parts at the end of a timestamp.
	// It is used for evaluating dates in mongosql conversion mode. See
	// ParseDateTimeMongo below for more details.
	validTimeFormat = regexp.MustCompile(`((\s|T)\d{1,2}[.:]\d{1,2}([.:]\d{1,2}(\.\d{0,7})?)?Z?)$`)
)

// CompareTo compares a SQLValue to another SQLValue. It returns -1 if left
// compares less than right; 1, if left compares greater than right; and 0
// if left compares equal to right.
func CompareTo(left, right SQLValue, collation *collation.Collation) (int, error) {
	if left.Kind() != right.Kind() {
		err := fmt.Errorf(
			"left SQLValue and right SQLValue are not of same kind (%x and %x, respectively)",
			left.Kind(), right.Kind(),
		)
		panic(err)
	}
	valueKind := left.Kind()

	if right.IsNull() {
		if left.IsNull() {
			return 0, nil
		}
		i, err := CompareTo(right, left, collation)
		return -i, err
	}

	if left.IsNull() {
		return -1, nil
	}

	if left.EvalType() == right.EvalType() {
		switch left.(type) {
		case SQLDate, SQLDecimal128, SQLFloat, SQLInt64, SQLUint64, SQLTimestamp:
			return mathutil.CompareDecimal128(Decimal(left), Decimal(right))
		case SQLVarchar, SQLObjectID:
			return collation.CompareString(String(left), String(right)), nil
		}
	}

	switch lVal := left.(type) {
	case SQLVarchar:
		switch right.(type) {
		case SQLDate, SQLTimestamp:
			// MySQL throws an error if you try to compare varchar =,<,> date/timestamp.
			// It works the other way around, however (i.e. date/timestamp =,<,> varchar).
			return -1, fmt.Errorf("Illegal mix of collations %T and %T", left, right)
		default:
			return mathutil.CompareDecimal128(Decimal(left), Decimal(right))
		}
	case SQLDate:
		switch rVal := right.(type) {
		case SQLVarchar:
			t, _, ok := ParseDateTime(right.String())
			if !ok {
				t, _, _ = ParseDateTime("0001-01-01")
			}
			return mathutil.CompareFloats(Float64(left), Float64(NewSQLDate(valueKind, t)))
		case SQLTimestamp:
			if Timestamp(rVal).Before(Timestamp(lVal)) {
				return 1, nil
			} else if Timestamp(rVal).After(Timestamp(lVal)) {
				return -1, nil
			}
			return 0, nil
		default:
			return mathutil.CompareDecimal128(Decimal(left), Decimal(right))
		}
	case SQLTimestamp:
		switch rVal := right.(type) {
		case SQLVarchar:
			t, _, ok := ParseDateTime(right.String())
			if !ok {
				t, _, _ = ParseDateTime("0001-01-01 00:00:00")
			}
			return mathutil.CompareFloats(Float64(left), Float64(NewSQLTimestamp(valueKind, t)))
		case SQLDate:
			if Timestamp(rVal).Before(Timestamp(lVal)) {
				return 1, nil
			} else if Timestamp(rVal).After(Timestamp(lVal)) {
				return -1, nil
			}
			return 0, nil
		default:
			return mathutil.CompareDecimal128(Decimal(left), Decimal(right))
		}
	default:
		switch right.(type) {
		default:
			return mathutil.CompareDecimal128(Decimal(left), Decimal(right))
		}
	}
}

// CompareToPairwise compares two slices of SQLValue in a pairwise manner. It
// returns -1 if the left compares less than the right; 1, if left compares
// greater than right; and 0 if left compares equal to right.
// The pairwise comparison means that left[0] is compared to right[0], left[1]
// is compared to right[1], and so on, using CompareTo(left[i], right[i]). If
// a pairwise comparison returns 0, the next pair is compared. If it returns
// non-0, this function returns that result. If all pairs are compared and
// found to be equal, this function returns 0.
func CompareToPairwise(left, right []SQLValue, collation *collation.Collation) (int, error) {
	if len(left) != len(right) {
		return -1, fmt.Errorf("different number of values on each side (left: %v, right: %v)", len(left), len(right))
	}

	for i, lv := range left {
		c, err := CompareTo(lv, right[i], collation)
		if err != nil {
			return -1, err
		}

		if c != 0 {
			return c, nil
		}
	}

	return 0, nil
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

// ParseDateTime parsers a DateTime.
func ParseDateTime(s string) (time.Time, int, bool) {
	return StrToDateTime(s, false)
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

func daysInMonth(m time.Month, year int) int {
	// This is equivalent to time.daysIn(m, year).
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// StrToDateTime is a port of mysql's str_to_datetime function.
func StrToDateTime(s string, full bool) (time.Time, int, bool) {

	// skip space at start
	var str int
	for str = 0; str < len(s); str++ {
		if !strutil.IsSpace(s[str]) {
			break
		}
	}

	if str >= len(s) || !strutil.IsDigit(s[str]) {
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
		if !strutil.IsDigit(s[pos]) && s[pos] != 'T' {
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
	for state = 0; state < maxDateParts-1 && str < len(s) && strutil.IsDigit(s[str]); state++ {
		start := str
		tempValue := int(s[str]) - int('0')
		str++

		// gather up all the digits for the current part
		scanUntilDelim := !internalFormat && state != microsecondIdx
		fieldLength--
		for str < len(s) && strutil.IsDigit(s[str]) && (scanUntilDelim || fieldLength > 0) {
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

		for str < len(s) && (strutil.IsPunct(s[str]) || strutil.IsSpace(s[str])) {
			if strutil.IsSpace(s[str]) {
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

// ParseDateTimeMongo parses a DateTime same as ParseDateTime above, except
// it handles special cases where mongo's behavior diverges from mysql.
// 1) - 10/11/12 is parsed as 2010-11-12 by mysql and 2012-10-11 by mongo,
//    - 10/11/2012 is invalid in mysql but is parsed as 2012-10-11 by mongo,
//    - 2012/10/11 is parsed as 2012-10-11 by both mysql and mongo.
//    Given those cases, the regular expression used in the function only
//    checks for the format [M]M/[D]D/[YYY]Y (1 or 2 character month/day,
//    and 1 to 4 character year).
// 2) The only valid separators in mongo are "/" and "-" and they must be
//    consistent (i.e. 10-11-12 or 10/11/12, but not 10-11/12); in mysql,
//    many more separators are tolerated and they can be mixed. This is only
//    enforced if separators are present.
// 3) If the time parts are present (after a space or T separating them from
//    the date), they must be properly formatted as HH:MM:SS[Z] in mongo.
//    In mysql, if a string is converted to a date (not a timestamp), the
//    time parts are ignored even if they are malformed.
func ParseDateTimeMongo(s string) (time.Time, int, bool) {
	s = strings.TrimSpace(s)

	sLen := len(s)

	// Case 1: if the string matches the special case slashed date, we
	// rearrange the pieces so that MM/DD/YYYY becomes YYYY-MM-DD.
	if slashDelimitedDDMMYYYYDateFormat.MatchString(s) {
		var month, day, year string
		pos := 0

		// get the month
		for ; s[pos] != '/'; pos++ {
		}
		month = s[0:pos]
		pos++

		// get the day
		start := pos
		for ; s[pos] != '/'; pos++ {
		}
		day = s[start:pos]
		pos++

		// get the year
		// Cases to consider:
		//   - 10/11/12            => 12
		//   - 10/11/2012          => 2012
		//   - 10/11/12 01:02:03   => 12
		//   - 10/11/12T01:02:03   => 12
		//   - 10/11/2012 01:02:03 => 2012
		//   - 10/11/2012T01:02:03 => 2012
		// The year can be no longer than 4 characters. If we encounter the
		// end of the string, a space, or a 'T' before scanning 4 characters
		// then that terminates the year.
		start = pos
		for i := 0; pos < sLen && s[pos] != ' ' && s[pos] != 'T' && i < 4; i++ {
			pos++
		}
		year = s[start:pos]

		// reassemble as YYYY-MM-DD...
		s = year + "-" + month + "-" + day + s[pos:]
	}

	// Case 2: if there are separators, they must be
	// either '-' or '/' and they must be consistent.
	pos := 0

	// First, move past the year. It can be no longer than 4 characters.
	// If we encounter the end of the string or a non-digit before scanning
	// 4 characters, then that terminates the year.
	for ; pos < sLen && strutil.IsDigit(s[pos]) && pos < 4; pos++ {
	}

	// If we reach the end of the string or there are 5 consecutive digits,
	// we know there are no separators (which may be valid!). We delegate
	// parsing to StrToDateTime.
	if pos == sLen || (pos == 4 && strutil.IsDigit(s[pos])) {
		// If there are no separators, the string must be at
		// least 8 characters long to be valid in mongo.
		if sLen < 8 {
			return time.Time{}, 0, false
		}
		return StrToDateTime(s, false)
	}

	// At this point, we know
	//   1) pos < sLen, and
	//   2) s[pos] is not a digit.
	// This implies that s[pos] must be a separator.
	sep := s[pos]
	if sep != '-' && sep != '/' {
		// Anything other than '-' and '/' are invalid in mongo.
		return time.Time{}, 0, false
	}

	// move past the separator
	pos++

	// Next, move past the month. It can be no longer than 2 characters. Same
	// as above, if we encounter the end of the string or a non-digit before
	// scanning 2 characters, then that terminates the month.
	for i := 0; pos < sLen && strutil.IsDigit(s[pos]) && i < 2; i++ {
		pos++
	}

	// If we reach the end of the string or s[pos] is a separator but not the
	// same one as before, this is immediately known to be invalid.
	if pos == sLen || (!strutil.IsDigit(s[pos]) && s[pos] != sep) {
		return time.Time{}, 0, false
	}

	// move past the separator
	pos++

	// Case 3: If the time parts are present, they must be formatted
	// as HH:MM:SS[Z]. At this point, we know the date is formatted
	// with some separators, so we can correctly enforce that the
	// time is also formatted as expected.

	// Now, move past the day. It can be no longer than 2 characters. Same
	// as above, if we encounter the end of the string or a non-digit before
	// scanning 2 characters, then that terminates the month.
	for i := 0; pos < sLen && strutil.IsDigit(s[pos]) && i < 2; i++ {
		pos++
	}

	// If we did not reach the end of the string after scanning the day part,
	// then the rest of the date string must end with a validly formatted
	// time portion. If it does not, this string is invalid.
	if pos < sLen && !validTimeFormat.MatchString(s) {
		return time.Time{}, 0, false
	}

	// At this point, we have handled the special mongo cases, so
	// now we delegate the actual parsing to StrToDateTime.
	return StrToDateTime(s, false)
}
