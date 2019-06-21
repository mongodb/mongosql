package evaluator

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/mathutil"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

// nolint: unparam
func (f baseScalarFunctionExpr) asciiEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	str := vs[0].String()
	if str == "" {
		return values.NewSQLInt64(sqlValueKind, 0), nil
	}

	c := str[0]

	return values.NewSQLInt64(sqlValueKind, int64(c)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) charEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {

	var b []byte
	for _, i := range vs {
		if i.IsNull() {
			continue
		}
		v := values.Int64(i)
		if v >= 256 {
			var temp []byte
			num := v / 255
			v = v % 256
			for num >= 256 {
				temp = append(temp, 0)
				num /= 255
			}
			b = append(b, uint8(num))
			b = append(b, temp...)
		}
		b = append(b, uint8(v))
	}

	return values.NewSQLVarchar(sqlValueKind, string(b)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) characterLengthEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	value := []rune(vs[0].String())

	return values.NewSQLInt64(sqlValueKind, int64(len(value))), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) concatEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (v values.SQLValue, err error) {
	if values.HasNullValue(vs...) {
		v = values.NewSQLNull(sqlValueKind)
		err = nil
		return
	}

	defer func() {
		if r := recover(); r != nil {
			v = nil
			err = fmt.Errorf("%v", r)
		}
	}()

	var b bytes.Buffer
	for _, value := range vs {
		b.WriteString(value.String())
	}

	v = values.NewSQLVarchar(sqlValueKind, b.String())
	err = nil
	return
}

// nolint: unparam
func (f baseScalarFunctionExpr) concatWsEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (v values.SQLValue, err error) {
	if vs[0].IsNull() {
		v = values.NewSQLNull(sqlValueKind)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			v = nil
			err = fmt.Errorf("%v", r)
		}
	}()

	var b bytes.Buffer
	separator := vs[0].String()
	trimValues := vs[1:]
	for i, value := range trimValues {
		if value.IsNull() {
			continue
		}
		b.WriteString(value.String())
		if i != len(trimValues)-1 {
			b.WriteString(separator)
		}
	}

	v = values.NewSQLVarchar(sqlValueKind, b.String())
	return
}

// Diverges from MySQL behavior in its handling of negative values
// Converts bases to positive numbers, and returns a negative value if the input is negative
// MySQL claims that "If from_base is a negative number, N is regarded as a signed number.
// Otherwise, N is treated as unsigned." Manual testing shows that it returns the 2's
// complement version if the number is negative unless the to_base is also negative, in which
// case it returns the number with a negative sign at the front
// nolint: unparam
func (f baseScalarFunctionExpr) convEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	num := vs[0].String()
	originalBase := absInt64(values.Int64(vs[1]))
	newBase := absInt64(values.Int64(vs[2]))
	negative := false

	if baseIsInvalid(originalBase) || baseIsInvalid(newBase) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	if string(num[0]) == "-" {
		num = num[1:]
		negative = true
	}

	if strings.Contains(num, ".") {
		num = num[0:strings.Index(num, ".")]
	}

	base10Version, err := strconv.ParseInt(num, int(originalBase), 64)
	if err != nil {
		return values.NewSQLVarchar(sqlValueKind, "0"), nil
	}
	strVersion := strconv.FormatInt(base10Version, int(newBase))

	if negative && strVersion != "0" {
		strVersion = "-" + strVersion
	}

	return values.NewSQLVarchar(sqlValueKind, strings.ToUpper(strVersion)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) convertEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if vs[0].IsNull() {
		return values.NewSQLNull(sqlValueKind), nil
	}

	typ, ok := evalTypeFromSQLTypeValue(vs[1])
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	return values.ConvertTo(vs[0], typ), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) cotEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	tan := math.Tan(values.Float64(vs[0]))
	if tan == 0 {
		return values.NewSQLNull(sqlValueKind),
			mysqlerrors.Defaultf(mysqlerrors.ErDataOutOfRange,
				"DOUBLE",
				fmt.Sprintf("'cot(%v)'",
					values.Float64(vs[0])))
	}

	return values.NewSQLFloat(sqlValueKind, 1/tan), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) dateAddEvaluate(sqlValueKind values.SQLValueKind, collation *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	_, _, ok := values.ParseDateTime(vs[0].String())
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	// Seconds can be fractional vs, so our calculateInterval function will not work right
	// (it is fine for all other units, as they must be integral).
	if vs[2].String() == Second {
		interval := vs[1].SQLFloat()
		vals := []values.SQLValue{values.NewSQLVarchar(sqlValueKind, Second), interval, vs[0]}
		return f.timestampAddEvaluate(sqlValueKind, collation, vals)
	}

	args, neg := dateArithmeticArgs(vs[2].String(), vs[1])
	unit, interval, err := calculateInterval(vs[2].String(), args, neg)
	if err != nil {
		return values.NewSQLNull(sqlValueKind), nil
	}

	vals := []values.SQLValue{
		values.NewSQLVarchar(sqlValueKind, unit),
		values.NewSQLInt64(sqlValueKind, int64(interval)), vs[0],
	}
	return f.timestampAddEvaluate(sqlValueKind, collation, vals)
}

// nolint: unparam
func (f baseScalarFunctionExpr) dateDiffEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {

	var left, right time.Time
	var ok bool

	parseArgs := func(val values.SQLValue) (time.Time, bool) {
		var date time.Time

		date, _, ok = values.StrToDateTime(val.String(), false)
		if !ok {
			return date, false
		}

		date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, schema.DefaultLocale)
		return date, true
	}

	if left, ok = parseArgs(vs[0]); !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	if right, ok = parseArgs(vs[1]); !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	durationDiff := left.Sub(right)
	hoursDiff := durationDiff.Hours()
	daysDiff := hoursDiff / 24

	diff := values.NewSQLInt64(sqlValueKind, int64(daysDiff))
	return diff, nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) dateFormatEvaluate(sqlValueKind values.SQLValueKind,
	collation *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	date, _, ok := values.ParseDateTime(vs[0].String())
	date = date.In(schema.DefaultLocale)
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	v1, ok := vs[1].(values.SQLVarchar)
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	ret, err := f.formatDate(sqlValueKind, collation, date, v1.String())
	if err != nil {
		return nil, err
	}
	return values.NewSQLVarchar(sqlValueKind, ret), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) dateSubEvaluate(sqlValueKind values.SQLValueKind, collation *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	v := vs[1].String()
	if string(v[0]) != "-" {
		v = "-" + v
	} else {
		v = v[1:]
	}

	vals := []values.SQLValue{vs[0], values.NewSQLVarchar(sqlValueKind, v), vs[2]}
	return f.dateAddEvaluate(sqlValueKind, collation, vals)
}

// nolint: unparam
func (f baseScalarFunctionExpr) dayNameEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	t, _, ok := values.ParseDateTime(vs[0].String())
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	return values.NewSQLVarchar(sqlValueKind, t.Weekday().String()), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) dayOfMonthEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	t, _, ok := values.ParseDateTime(values.String(vs[0]))
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	return values.NewSQLInt64(sqlValueKind, int64(t.Day())), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) dayOfWeekEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	t, _, ok := values.ParseDateTime(values.String(vs[0]))
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	return values.NewSQLInt64(sqlValueKind, int64(t.Weekday())+1), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) dayOfYearEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	t, _, ok := values.ParseDateTime(vs[0].String())
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	return values.NewSQLInt64(sqlValueKind, int64(t.YearDay())), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) dualArgFloatMathFuncEvaluate(sqlValueKind values.SQLValueKind,
	vs []values.SQLValue, fn func(float64, float64) float64) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	result := fn(values.Float64(vs[0]), values.Float64(vs[1]))
	if math.IsNaN(result) {
		return values.NewSQLNull(sqlValueKind), nil
	}
	if math.IsInf(result, 0) {
		return values.NewSQLNull(sqlValueKind), nil
	}
	if result == -0 {
		result = 0
	}
	return values.NewSQLFloat(sqlValueKind, result), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) extractEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	t, _, ok := values.ParseDateTime(vs[1].String())
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	units := [6]int{t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second()}

	var unitStrs [6]string
	// For certain units, we need to concatenate the unit vs as strings
	// before returning the int value as to not lose any number's place
	// value. i.e. SELECT EXTRACT(DayMinute FROM "2006-04-03 06:03:23")
	// should return 30603, not 363.
	for idx, val := range units {
		u := strconv.Itoa(val)
		if len(u) == 1 {
			u = "0" + u
		}
		unitStrs[idx] = u
	}

	switch vs[0].String() {
	case Year:
		return values.NewSQLInt64(sqlValueKind, int64(units[0])), nil
	case Quarter:
		return values.NewSQLInt64(sqlValueKind, int64(math.Ceil(float64(units[1])/3.0))), nil
	case Month:
		return values.NewSQLInt64(sqlValueKind, int64(units[1])), nil
	case Week:
		_, w := t.ISOWeek()
		return values.NewSQLInt64(sqlValueKind, int64(w)), nil
	case Day:
		return values.NewSQLInt64(sqlValueKind, int64(units[2])), nil
	case Hour:
		return values.NewSQLInt64(sqlValueKind, int64(units[3])), nil
	case Minute:
		return values.NewSQLInt64(sqlValueKind, int64(units[4])), nil
	case Second:
		return values.NewSQLInt64(sqlValueKind, int64(units[5])), nil
	case Microsecond:
		return values.NewSQLInt64(sqlValueKind, 0), nil
	case YearMonth:
		ym, _ := strconv.ParseInt(unitStrs[0]+unitStrs[1], 10, 64)
		return values.NewSQLInt64(sqlValueKind, ym), nil
	case DayHour:
		dh, _ := strconv.ParseInt(unitStrs[2]+unitStrs[3], 10, 64)
		return values.NewSQLInt64(sqlValueKind, dh), nil
	case DayMinute:
		dm, _ := strconv.ParseInt(unitStrs[2]+unitStrs[3]+unitStrs[4], 10, 64)
		return values.NewSQLInt64(sqlValueKind, dm), nil
	case DaySecond:
		ds, _ := strconv.ParseInt(unitStrs[2]+unitStrs[3]+unitStrs[4]+unitStrs[5], 10, 64)
		return values.NewSQLInt64(sqlValueKind, ds), nil
	case DayMicrosecond:
		dms, _ := strconv.ParseInt(unitStrs[2]+unitStrs[3]+unitStrs[4]+unitStrs[5]+"000000", 10, 64)
		return values.NewSQLInt64(sqlValueKind, dms), nil
	case HourMinute:
		hm, _ := strconv.ParseInt(unitStrs[3]+unitStrs[4], 10, 64)
		return values.NewSQLInt64(sqlValueKind, hm), nil
	case HourSecond:
		hs, _ := strconv.ParseInt(unitStrs[3]+unitStrs[4]+unitStrs[5], 10, 64)
		return values.NewSQLInt64(sqlValueKind, hs), nil
	case HourMicrosecond:
		hms, _ := strconv.ParseInt(unitStrs[3]+unitStrs[4]+unitStrs[5]+"000000", 10, 64)
		return values.NewSQLInt64(sqlValueKind, hms), nil
	case MinuteSecond:
		ms, _ := strconv.ParseInt(unitStrs[4]+unitStrs[5], 10, 64)
		return values.NewSQLInt64(sqlValueKind, ms), nil
	case MinuteMicrosecond:
		mms, _ := strconv.ParseInt(unitStrs[4]+unitStrs[5]+"000000", 10, 64)
		return values.NewSQLInt64(sqlValueKind, mms), nil
	case SecondMicrosecond:
		sms, _ := strconv.ParseInt(unitStrs[5]+"000000", 10, 64)
		return values.NewSQLInt64(sqlValueKind, sms), nil
	default:
		err := fmt.Errorf("unit type '%v' is not supported", vs[0].String())
		return values.NewSQLNull(sqlValueKind), err
	}
}

// nolint: unparam
func (f baseScalarFunctionExpr) fromDaysEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	parseNumeric := func(s string) string {
		f := func(r rune) bool {
			return !unicode.IsNumber(r) &&
				(r != 46 /* unicode '.' */)
		}
		n := strings.FieldsFunc(s, f)
		if len(n) > 0 {
			return n[0]
		}
		return ""
	}

	v := vs[0].String()
	neg := len(v) > 0 && v[0] == '-'
	if neg {
		v = v[1:]
	}
	value, err := strconv.ParseFloat(parseNumeric(v), 64)
	if err != nil {
		return values.NewSQLNull(sqlValueKind), nil
	}
	if neg {
		value = -value
	}

	if value <= 365.5 || value >= 3652499.5 {
		return values.NewSQLNull(sqlValueKind), nil
	}

	abs, maxGoDurationHours := math.Abs(value-366), int64(106751)
	target := int64(math.Floor(abs + .5))

	// edge cases
	if math.Ceil(abs) == 1 {
		target = 0
	}

	if math.Floor(abs) == 3652133 {
		target = 3652133
	}

	date := zeroDate

	for target > 0 && target > maxGoDurationHours {
		date = date.Add(time.Duration(maxGoDurationHours*24) * time.Hour)
		target -= maxGoDurationHours
	}

	date = date.Add(time.Duration(target*24) * time.Hour).Round(time.Second)

	return values.NewSQLDate(sqlValueKind, date.In(schema.DefaultLocale)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) fromUnixtimeEvaluate(sqlValueKind values.SQLValueKind,
	collation *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	value := mathutil.Round(values.Float64(vs[0]))
	if value < 0 {
		return values.NewSQLNull(sqlValueKind), nil
	}

	date := time.Unix(value, 0).In(schema.DefaultLocale)
	if len(vs) == 1 {
		return values.NewSQLTimestamp(sqlValueKind, date), nil
	}
	ret, err := f.formatDate(sqlValueKind, collation, date, vs[1].String())
	if err != nil {
		return nil, err
	}
	return values.NewSQLVarchar(sqlValueKind, ret), nil
}

// greatest cannot be constant folded because we do not have the collation at optimization time.
// nolint: unparam
func (f baseScalarFunctionExpr) greatestEvaluate(sqlValueKind values.SQLValueKind,
	collation *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	valExprs := []types.EvalTyper{}
	for _, val := range vs {
		valExprs = append(valExprs, val)
	}
	convertTo := preferentialType(valExprs...)

	convertedVals := []values.SQLValue{}
	for _, val := range vs {
		newVal := values.ConvertTo(val, convertTo)
		convertedVals = append(convertedVals, newVal)
	}

	var greatest values.SQLValue
	var greatestIdx int

	c, err := values.CompareTo(convertedVals[0], convertedVals[1], collation)
	if err != nil {
		return values.NewSQLNull(sqlValueKind), err
	}

	if c == -1 {
		greatest, greatestIdx = vs[1], 1
	} else {
		greatest, greatestIdx = vs[0], 0
	}

	for i := 2; i < len(vs); i++ {
		c, err = values.CompareTo(greatest, convertedVals[i], collation)
		if err != nil {
			return values.NewSQLNull(sqlValueKind), err
		}
		if c == -1 {
			greatest, greatestIdx = vs[i], i
		}
	}

	allTimeVals, timestamp := areAllTimeTypes(vs)
	if allTimeVals && timestamp {
		t, _, _ := values.ParseDateTime(vs[greatestIdx].String())
		return values.NewSQLTimestamp(sqlValueKind, t), nil
	} else if convertTo == types.EvalDate || convertTo == types.EvalDatetime {
		return vs[greatestIdx], nil
	}

	return convertedVals[greatestIdx], nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) hourEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if vs[0].IsNull() {
		return values.NewSQLNull(sqlValueKind), nil
	}
	_, hour, ok := parseTime(vs[0].String())
	if !ok {
		// If we managed to parse, but minutes or seconds are >= 60
		// MySQL returns NULL for the hour/minute/second function.
		// Rather than return yet another value, we coop the hour value
		// and return -1, thus we can check for -1 here to return NULL
		// rather than the 0 expected if the string could not be parsed
		// at all.
		if hour == -1 {
			return values.NewSQLNull(sqlValueKind), nil
		}
		return values.NewSQLInt64(sqlValueKind, 0), nil
	}

	return values.NewSQLInt64(sqlValueKind, int64(hour)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) insertEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {

	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	s := vs[0].String()
	pos := int(mathutil.Round(values.Float64(vs[1]))) - 1
	length := int(mathutil.Round(values.Float64(vs[2])))
	newstr := vs[3].String()

	if pos < 0 || pos >= len(s) {
		return vs[0], nil
	}

	if pos+length < 0 || pos+length > len(s) {
		return values.NewSQLVarchar(sqlValueKind, s[:pos]+newstr), nil
	}

	return values.NewSQLVarchar(sqlValueKind, s[:pos]+newstr+s[pos+length:]), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) instrEvaluate(sqlValueKind values.SQLValueKind,
	collation *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.locateEvaluate(sqlValueKind, collation, []values.SQLValue{vs[1], vs[0]})
}

// nolint: unparam
func (f baseScalarFunctionExpr) lastDayEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	// Must be a SQLTimestamp at this point.
	tmp, ok := vs[0].(values.SQLDate)
	if !ok {
		return nil,
			fmt.Errorf("unable to evaluate type %v"+
				" in to_days, this points to an error in the algebrizer",
				vs[0])
	}
	t := values.Timestamp(tmp)
	year, month, _ := t.Date()
	first := time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
	return values.NewSQLDate(sqlValueKind, first.AddDate(0, 1, -1)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) lcaseEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	value := strings.ToLower(vs[0].String())

	return values.NewSQLVarchar(sqlValueKind, value), nil
}

// least cannot be constant folded because of the need to have the collation from the mongo collection.
// nolint: unparam
func (f baseScalarFunctionExpr) leastEvaluate(sqlValueKind values.SQLValueKind,
	collation *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	valExprs := []types.EvalTyper{}
	for _, val := range vs {
		valExprs = append(valExprs, val)
	}
	convertTo := preferentialType(valExprs...)

	convertedVals := []values.SQLValue{}
	for _, val := range vs {
		newVal := values.ConvertTo(val, convertTo)
		convertedVals = append(convertedVals, newVal)
	}

	var least values.SQLValue
	var leastIdx int

	c, err := values.CompareTo(convertedVals[0], convertedVals[1], collation)
	if err != nil {
		return values.NewSQLNull(sqlValueKind), err
	}

	if c == -1 {
		least, leastIdx = convertedVals[0], 0
	} else {
		least, leastIdx = convertedVals[1], 1
	}

	for i := 2; i < len(vs); i++ {
		c, err = values.CompareTo(least, convertedVals[i], collation)
		if err != nil {
			return values.NewSQLNull(sqlValueKind), err
		}
		if c == 1 {
			least, leastIdx = vs[i], i
		}
	}

	allTimeVals, timestamp := areAllTimeTypes(vs)
	if allTimeVals && timestamp {
		t, _, _ := values.ParseDateTime(vs[leastIdx].String())
		return values.NewSQLTimestamp(sqlValueKind, t), nil
	} else if convertTo == types.EvalDate || convertTo == types.EvalDatetime {
		return vs[leastIdx], nil
	}

	return convertedVals[leastIdx], nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) leftEvaluate(sqlValueKind values.SQLValueKind,
	collation *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	vals := []values.SQLValue{vs[0],
		values.NewSQLInt64(sqlValueKind, 1),
		vs[1]}
	return f.substringEvaluate(sqlValueKind, collation, vals)
}

// nolint: unparam
func (f baseScalarFunctionExpr) lengthEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	value := vs[0].String()

	return values.NewSQLInt64(sqlValueKind, int64(len(value))), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) locateEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs[:2]...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	substr := []rune(vs[0].String())
	str := []rune(vs[1].String())
	var result int
	if len(vs) == 3 {

		pos := int(values.Float64(vs[2])+0.5) - 1 // MySQL uses 1 as a basis

		if pos < 0 || len(str) <= pos {
			return values.NewSQLInt64(sqlValueKind, 0), nil
		}
		str = str[pos:]
		result = runesIndex(str, substr)
		if result >= 0 {
			result += pos
		}
	} else {
		result = runesIndex(str, substr)
	}

	return values.NewSQLInt64(sqlValueKind, int64(result+1)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) logEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.logarithmEvaluate(sqlValueKind, vs, 0)
}

// nolint: unparam
func (f baseScalarFunctionExpr) lnEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.logarithmEvaluate(sqlValueKind, vs, 0)
}

// nolint: unparam
func (f baseScalarFunctionExpr) log2Evaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.logarithmEvaluate(sqlValueKind, vs, 2)
}

// nolint: unparam
func (f baseScalarFunctionExpr) log10Evaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.logarithmEvaluate(sqlValueKind, vs, 10)
}

// nolint: unparam
func (f baseScalarFunctionExpr) logarithmEvaluate(sqlValueKind values.SQLValueKind, vs []values.SQLValue, base uint32) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	var result float64
	switch base {
	case 0:
		if len(vs) == 2 {
			// arbitrary base
			result = math.Log(values.Float64(vs[1])) / math.Log(values.Float64(vs[0]))
		} else {
			// natural base
			result = math.Log(values.Float64(vs[0]))
		}
	case 2:
		result = math.Log2(values.Float64(vs[0]))
	case 10:
		result = math.Log10(values.Float64(vs[0]))
	}
	if math.IsNaN(result) {
		return values.NewSQLNull(sqlValueKind), nil
	}
	if math.IsInf(result, 0) {
		return values.NewSQLNull(sqlValueKind), nil
	}
	if result == -0 {
		result = 0
	}
	return values.NewSQLFloat(sqlValueKind, result), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) lpadEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return handlePadding(sqlValueKind, vs, true)
}

// nolint: unparam
func (f baseScalarFunctionExpr) ltrimEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	value := strings.TrimLeft(vs[0].String(), " ")

	return values.NewSQLVarchar(sqlValueKind, value), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) makeDateEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	// Floating arguments should be mathutil.Rounded.
	y := mathutil.Round(values.Float64(vs[0]))
	if y < 0 || y > 9999 {
		return values.NewSQLNull(sqlValueKind), nil
	}
	if y >= 0 && y <= 69 {
		y += 2000
	} else if y >= 70 && y <= 99 {
		y += 1900
	}

	d := mathutil.Round(values.Float64(vs[1]))

	if d <= 0 {
		return values.NewSQLNull(sqlValueKind), nil
	}

	t := time.Date(int(y), 1, 0, 0, 0, 0, 0, schema.DefaultLocale)
	duration := time.Duration(d*24) * time.Hour

	output := t.Add(duration)
	if output.Year() > 9999 {
		return values.NewSQLNull(sqlValueKind), nil
	}

	return values.NewSQLDate(sqlValueKind, output), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) md5Evaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	h := md5.New()
	_, err := io.WriteString(h, vs[0].String())
	if err != nil {
		return nil, err
	}
	return values.NewSQLVarchar(sqlValueKind, fmt.Sprintf("%x", h.Sum(nil))), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) microsecondEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {

	arg := vs[0]

	if arg.IsNull() {
		return values.NewSQLNull(sqlValueKind), nil
	}

	str := arg.String()
	if str == "" {
		return values.NewSQLNull(sqlValueKind), nil
	}

	t, _, ok := parseTime(str)
	if !ok {
		return values.NewSQLInt64(sqlValueKind, 0), nil
	}

	return values.NewSQLInt64(sqlValueKind, int64(t.Nanosecond()/1000)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) minuteEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if vs[0].IsNull() {
		return values.NewSQLNull(sqlValueKind), nil
	}

	t, hour, ok := parseTime(vs[0].String())
	if !ok {
		// If we managed to parse, but minutes or seconds are >= 60
		// MySQL returns NULL for the hour/minute/second function.
		// Rather than return yet another value, we coop the hour value
		// and return -1, thus we can check for -1 here to return NULL
		// rather than the 0 expected if the string could not be parsed
		// at all.
		if hour == -1 {
			return values.NewSQLNull(sqlValueKind), nil
		}
		return values.NewSQLInt64(sqlValueKind, 0), nil
	}

	return values.NewSQLInt64(sqlValueKind, int64(t.Minute())), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) modEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.dualArgFloatMathFuncEvaluate(sqlValueKind, vs, math.Mod)
}

// nolint: unparam
func (f baseScalarFunctionExpr) monthEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	t, _, ok := values.ParseDateTime(vs[0].String())
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	return values.NewSQLInt64(sqlValueKind, int64(t.Month())), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) monthNameEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	t, _, ok := values.ParseDateTime(vs[0].String())
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	return values.NewSQLVarchar(sqlValueKind, t.Month().String()), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) nopushdownEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return vs[0], nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) powEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	v0 := values.Float64(vs[0])
	v1 := values.Float64(vs[1])

	n := math.Pow(v0, v1)
	zeroBaseExpNeg := v0 == 0 && v1 < 0
	if math.IsNaN(n) || zeroBaseExpNeg {
		return values.NewSQLNull(sqlValueKind),
			mysqlerrors.Defaultf(mysqlerrors.ErDataOutOfRange,
				"DOUBLE",
				fmt.Sprintf("pow(%v,%v)",
					values.Float64(vs[0]),
					values.Float64(vs[1])))
	}

	return values.NewSQLFloat(sqlValueKind, math.Pow(v0, v1)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) quarterEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	t, _, ok := values.ParseDateTime(vs[0].String())
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	q := 0
	switch t.Month() {
	case 1, 2, 3:
		q = 1
	case 4, 5, 6:
		q = 2
	case 7, 8, 9:
		q = 3
	case 10, 11, 12:
		q = 4
	}

	return values.NewSQLInt64(sqlValueKind, int64(q)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) randEvaluateWithFullEvaluationState(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, vs []values.SQLValue) (values.SQLValue, error) {
	uniqueID := values.Uint64(vs[0])

	if len(vs) == 2 {
		seed := mathutil.Round(values.Float64(vs[1]))
		r := st.RandomWithSeed(uniqueID, seed)
		return values.NewSQLFloat(cfg.sqlValueKind, r.Float64()), nil
	}

	r := st.Random(uniqueID)
	return values.NewSQLFloat(cfg.sqlValueKind, r.Float64()), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) repeatEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (v values.SQLValue, err error) {
	if values.HasNullValue(vs...) {
		v = values.NewSQLNull(sqlValueKind)
		err = nil
		return
	}

	str := vs[0].String()
	if len(str) < 1 {
		v = values.NewSQLVarchar(sqlValueKind, "")
		err = nil
		return
	}

	rep := int(mathutil.RoundToDecimalPlaces(0, values.Float64(vs[1])))
	if rep < 1 {
		v = values.NewSQLVarchar(sqlValueKind, "")
		err = nil
		return
	}

	var b bytes.Buffer
	for i := 0; i < rep; i++ {
		b.WriteString(str)
	}

	v = values.NewSQLVarchar(sqlValueKind, b.String())
	err = nil
	return
}

// nolint: unparam
func (f baseScalarFunctionExpr) replaceEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	s := vs[0].String()
	old := vs[1].String()
	new := vs[2].String()

	return values.NewSQLVarchar(sqlValueKind, strings.Replace(s, old, new, -1)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) reverseEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}
	s := vs[0].String()
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return values.NewSQLVarchar(sqlValueKind, string(runes)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) rightEvaluate(sqlValueKind values.SQLValueKind, collation *collation.Collation,
	vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	str := vs[0].String()
	posFloat := values.Float64(vs[1])

	if posFloat > float64(len(str)) {
		return values.NewSQLVarchar(sqlValueKind, str), nil
	}

	startPos := math.Min(0, -1.0*posFloat)

	return f.substringEvaluate(sqlValueKind, collation,
		[]values.SQLValue{vs[0], values.NewSQLFloat(sqlValueKind, startPos)})
}

// nolint: unparam
func (f baseScalarFunctionExpr) roundEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	base := values.Float64(vs[0])

	var decimal int64
	if len(vs) == 2 {
		decimal = values.Int64(vs[1])

		if decimal < 0 {
			return values.NewSQLFloat(sqlValueKind, 0), nil
		}
	} else {
		decimal = 0
	}

	rounded := mathutil.RoundToDecimalPlaces(decimal, base)

	return values.NewSQLFloat(sqlValueKind, rounded), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) rpadEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return handlePadding(sqlValueKind, vs, false)
}

// nolint: unparam
func (f baseScalarFunctionExpr) rtrimEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	value := strings.TrimRight(vs[0].String(), " ")

	return values.NewSQLVarchar(sqlValueKind, value), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) secondEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if vs[0].IsNull() {
		return values.NewSQLNull(sqlValueKind), nil
	}

	t, hour, ok := parseTime(vs[0].String())
	if !ok {
		// If we managed to parse, but minutes or seconds are >= 60
		// MySQL returns NULL for the hour/minute/second function.
		// Rather than return yet another value, we coop the hour value
		// and return -1, thus we can check for -1 here to return NULL
		// rather than the 0 expected if the string could not be parsed
		// at all.
		if hour == -1 {
			return values.NewSQLNull(sqlValueKind), nil
		}
		return values.NewSQLInt64(sqlValueKind, 0), nil
	}

	return values.NewSQLInt64(sqlValueKind, int64(t.Second())), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) signEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	v := values.Float64(vs[0])
	// Positive numbers are more common than negative in most data sets.
	if v > 0 {
		return values.NewSQLInt64(sqlValueKind, 1), nil
	}
	if v < 0 {
		return values.NewSQLInt64(sqlValueKind, -1), nil
	}
	return values.NewSQLInt64(sqlValueKind, 0), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) singleArgFloatMathFuncEvaluate(sqlValueKind values.SQLValueKind,
	vs []values.SQLValue, fn func(float64) float64) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	result := fn(values.Float64(vs[0]))
	if math.IsNaN(result) {
		return values.NewSQLNull(sqlValueKind), nil
	}
	if math.IsInf(result, 0) {
		return values.NewSQLNull(sqlValueKind), nil
	}
	if result == -0 {
		result = 0
	}
	return values.NewSQLFloat(sqlValueKind, result), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) sleepEvaluateWithFullEvaluationState(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, vs []values.SQLValue) (values.SQLValue, error) {

	err := mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, "sleep")

	if values.HasNullValue(vs...) {
		return nil, err
	}

	n := values.Float64(vs[0])

	if n < 0 {
		return nil, err
	}

	timer := time.NewTimer(time.Second * time.Duration(n))

	select {
	case <-timer.C:
	case <-ctx.Done():
		timer.Stop()
		return nil, ctx.Err()
	}

	return values.NewSQLInt64(cfg.sqlValueKind, 0), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) spaceEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	flt := values.Float64(vs[0])
	n := mathutil.Round(flt)
	if n < 1 {
		return values.NewSQLVarchar(sqlValueKind, ""), nil
	}

	return values.NewSQLVarchar(sqlValueKind, strings.Repeat(" ", int(n))), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) strToDateEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	str, ok := vs[0].(values.SQLVarchar)
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	ft, ok := vs[1].(values.SQLVarchar)
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	s := str.String()
	ftStr := ft.String()
	fmtTokens := map[string]string{
		"%a": "Mon",
		"%b": "Jan",
		"%c": "1",
		"%d": "02",
		"%e": "2",
		"%H": "15",
		"%i": "04",
		"%k": "13",
		"%M": "January",
		"%m": "01",
		"%S": "05",
		"%s": "05",
		"%T": "15:04:05",
		"%W": "Monday",
		"%w": "Mon",
		"%Y": "2006",
		"%y": "06",
	}

	format := ""
	skipToken := false
	ts := false
	for idx, char := range ftStr {
		if !skipToken {
			if char == 37 && idx != len(ftStr)-1 {
				token := "%" + string(ftStr[idx+1])
				skipToken = true
				goToken := fmtTokens[token]
				if goToken != "" {
					format += goToken
				} else {
					return values.NewSQLNull(sqlValueKind), nil
				}
				if token == "%H" || token == "%i" || token == "%k" || token == "%p" ||
					token == "%S" || token == "%s" || token == "%T" {
					ts = true
				}
			} else {
				format += string(char)
			}
		} else {
			skipToken = false
		}
	}

	d, err := time.Parse(format, s)
	if err != nil {
		return values.NewSQLNull(sqlValueKind), nil
	}

	if ts {
		return values.NewSQLTimestamp(sqlValueKind, d), nil
	}

	return values.NewSQLDate(sqlValueKind, d), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) midEvaluate(sqlValueKind values.SQLValueKind, collation *collation.Collation,
	vs []values.SQLValue) (values.SQLValue, error) {
	return f.substringEvaluate(sqlValueKind, collation, vs)
}

// nolint: unparam
func (f baseScalarFunctionExpr) substringEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	str := []rune(vs[0].String())
	if vs[0].String() == "" {
		return values.NewSQLVarchar(sqlValueKind, ""), nil
	}

	posFloat := values.Float64(vs[1])
	var pos int
	if posFloat >= 0 {
		pos = int(posFloat + 0.5)
	} else {
		pos = int(posFloat - 0.5)
	}

	if pos > len(str) || pos == 0 {
		return values.NewSQLVarchar(sqlValueKind, ""), nil
	} else if pos < 0 {
		pos = len(str) + pos

		if pos < 0 {
			return values.NewSQLVarchar(sqlValueKind, ""), nil
		}
	} else {
		pos-- // MySQL uses 1 as a basis
	}

	if len(vs) == 3 {
		length := int(values.Float64(vs[2]) + 0.5)
		if length < 1 {
			return values.NewSQLVarchar(sqlValueKind, ""), nil
		}
		if pos <= len(str) {
			str = str[pos:]
		}
		if length <= len(str) {
			str = str[:length]
		}
	} else {
		if pos < len(str) {
			str = str[pos:]
		}
	}
	return values.NewSQLVarchar(sqlValueKind, string(str)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) substringIndexEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {

	matches := func(r []rune, pos int, delim []rune) bool {
		for i, j := pos, 0; i < len(r) && j < len(delim); i, j = i+1, j+1 {
			if r[i] != delim[j] {
				return false
			}
		}
		return true
	}

	reverse := func(r []rune) {
		for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
	}

	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	r := []rune(vs[0].String())
	delim := []rune(vs[1].String())

	count := int(mathutil.Round(values.Float64(vs[2])))

	if count == 0 {
		return values.NewSQLVarchar(sqlValueKind, ""), nil
	}

	reversed := count < 0
	if reversed {
		reverse(r)
		reverse(delim)

		count = -count
	}

	matchCount := 0
	i := 0
	for {
		if matches(r, i, delim) {
			matchCount++
			i += len(delim)
		} else {
			i++
		}

		if i >= len(r) || matchCount >= count {
			break
		}
	}

	if matchCount >= count {
		r = r[:i-len(delim)]
	}

	if reversed {
		reverse(r)
	}

	return values.NewSQLVarchar(sqlValueKind, string(r)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) timeDiffEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	expr1, _, ok := parseTime(vs[0].String())
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	expr2, _, ok := parseTime(vs[1].String())
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	d := expr1.Sub(expr2)

	fmtDef := func(buf []byte, v int) int {
		def := []byte{':', '0', '0', ':', '0', '0'}
		w := len(buf)
		for _, b := range def[v:] {
			w--
			buf[w] = b
		}
		return w
	}

	// Precision is exactly 6 digits (microsecond range)
	fmtFrac := func(buf []byte, v uint64) (nw int, nv uint64) {

		w, print := len(buf), false
		end := len(buf) - 1

		for i := 0; i < 6; i++ {

			digit := v % 10

			print = print || digit != 0

			if print {
				buf[end-i] = byte(digit) + '0'
				w--
			}
			v /= 10
		}

		if w != len(buf) {

			// If we have printed anything, pad the rest of the range with zeroes
			for len(buf)-w < 6 {
				buf[end] = '0'
				end--
				w--
			}
			w--
			buf[w] = '.'
		}

		return w, v
	}

	fmtInt := func(buf []byte, v uint64) int {
		w := len(buf)
		if v == 0 {
			w--
			buf[w] = '0'
		} else {
			for v > 0 {
				w--
				buf[w] = byte(v%10) + '0'
				v /= 10
			}
		}
		for len(buf)-w < 2 {
			w--
			buf[w] = '0'
		}
		return w
	}

	u := uint64(d)

	if u == 0 {
		return values.NewSQLVarchar(sqlValueKind, "00:00:00.000000"), nil
	}

	buf := [30]byte{}
	w := len(buf)

	neg := d < 0
	if neg {
		u = -u
	}

	// Shave off nanosecond precision (we need up to microseconds)
	u /= 1000

	// Handle fractional portion (< 1 second)
	w, u = fmtFrac(buf[:w], u)

	if u < uint64(time.Microsecond) {

		w = fmtInt(buf[:w], u)
		w = fmtDef(buf[:w], 0)

	} else {

		// u is now integer seconds
		w = fmtInt(buf[:w], u%60)
		u /= 60

		w--
		buf[w] = ':'

		// u is now integer minutes
		if u > 0 {
			w = fmtInt(buf[:w], u%60)
			u /= 60

			w--
			buf[w] = ':'

			// u is now integer hours
			if u > 0 {
				w = fmtInt(buf[:w], u)
			} else {
				w = fmtDef(buf[:w], 4)
			}
		} else {
			w = fmtDef(buf[:w], 1)
		}
	}

	if neg {
		w--
		buf[w] = '-'
	}

	return values.NewSQLVarchar(sqlValueKind, string(buf[w:])), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) timeToSecEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	dateParts := strings.Split(vs[0].String(), " ")
	hasDatePart := len(dateParts) == 2
	components := strings.Split(vs[0].String(), ":")
	if hasDatePart {
		components = strings.Split(dateParts[1], ":")
	}

	result, componentized := 0.0, true

	if len(components) == 1 {
		cmp, err := strconv.ParseFloat(components[0], 64)
		if err != nil {
			return values.NewSQLNull(sqlValueKind), err
		}

		component := strconv.FormatFloat(math.Trunc(cmp), 'f', -1, 64)

		l := len(component)
		components, componentized = []string{"0", "0", "0"}, false

		// MySQL interprets abbreviated vs without colons using the
		// assumption that the two rightmost digits represent seconds.
		switch l {
		case 1, 2:
			components[2] = component
		case 3, 4:
			components[1],
				components[2] = component[:l-2],
				component[l-2:l]
		case 5:
			components[0],
				components[1],
				components[2] = component[:l-4],
				component[l-4:l-2],
				component[l-2:l]
		default:
			components[0],
				components[1],
				components[2] = component[:l-4],
				component[l-4:l-2],
				component[l-2:l]
		}
	}

	signBit := false

	for i := 0; i < 3 && i < len(components); i++ {
		component, err := strconv.ParseFloat(components[i], 64)
		if err != nil {
			return values.NewSQLNull(sqlValueKind), err
		}

		cmp := math.Trunc(component)

		switch i {
		// more on valid time types at https://dev.mysql.com/doc/refman/5.7/en/time.html
		case 0:
			if cmp > 838 || cmp < -838 {
				if !componentized {
					return values.NewSQLNull(sqlValueKind), nil
				}
				cmp = math.Copysign(838.0, cmp)
				components = []string{"", "59", "59"}
			}
		default:
			if cmp > 59 {
				return values.NewSQLNull(sqlValueKind), nil
			}
		}

		signBit = signBit || math.Signbit(cmp)
		result += math.Abs(cmp) * (3600.0 / (math.Pow(60, float64(i))))
	}

	if signBit {
		return values.NewSQLFloat(sqlValueKind, math.Copysign(result, -1)), nil
	}

	return values.NewSQLFloat(sqlValueKind, result), nil
}

// isLeapYear returns true if the passed int year is a leap year.
func isLeapYear(y int) bool {
	return (y%4 == 0) && (y%100 != 0) || (y%400 == 0)
}

// nolint: unparam
func (f baseScalarFunctionExpr) timestampAddEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	v := int(mathutil.Round(values.Float64(vs[1])))
	// vs[2] must be a SQLTimestamp or the algebrizer has been broken.
	// Check the handling of scalar functions in the algebrizer.
	tmp, ok := vs[2].(values.SQLTimestamp)
	if !ok {
		return nil,
			fmt.Errorf("unable to evaluate type %v in timestampadd,"+
				" this points to an error in the algebrizer",
				vs[2])
	}
	t := values.Timestamp(tmp)

	ts := false
	if len(vs[2].String()) > 10 {
		ts = true
	}

	switch vs[0].String() {
	case Year:
		return values.NewSQLTimestamp(sqlValueKind, t.AddDate(v, 0, 0)), nil
	case Quarter:
		y, mp, d := t.Date()
		m := int(mp)
		interval := v * 3
		y += int(math.Floor(float64(m+interval-1) / float64(12)))
		// want to calculate ((m + interval - 1) mod 12), but Go's % operator does remainder
		// a mod b = ((a % b) + b) % b
		m = (((m+interval-1)%12)+12)%12 + 1
		switch m {
		case 2:
			if isLeapYear(y) {
				d = mathutil.MinInt(d, 29)
			} else {

				d = mathutil.MinInt(d, 28)
			}
		case 4, 6, 9, 11:
			d = mathutil.MinInt(d, 30)
		}
		if ts {
			return values.NewSQLTimestamp(sqlValueKind, time.Date(y,
					time.Month(m),
					d,
					t.Hour(),
					t.Minute(),
					t.Second(),
					t.Nanosecond(),
					schema.DefaultLocale)),
				nil
		}
		return values.NewSQLTimestamp(sqlValueKind, time.Date(y,
				time.Month(m),
				d,
				0,
				0,
				0,
				0,
				schema.DefaultLocale)),
			nil
	case Month:
		y, mp, d := t.Date()
		m := int(mp)
		interval := v
		y += int(math.Floor(float64(m+interval-1) / float64(12)))
		// want to calculate ((m + interval - 1) mod 12), but Go's % operator does remainder
		// a mod b = ((a % b) + b) % b
		m = (((m+interval-1)%12)+12)%12 + 1
		switch m {
		case 2:
			if isLeapYear(y) {
				d = mathutil.MinInt(d, 29)
			} else {

				d = mathutil.MinInt(d, 28)
			}
		case 4, 6, 9, 11:
			d = mathutil.MinInt(d, 30)
		}
		if ts {
			return values.NewSQLTimestamp(sqlValueKind, time.Date(y,
					time.Month(m),
					d,
					t.Hour(),
					t.Minute(),
					t.Second(),
					t.Nanosecond(),
					schema.DefaultLocale)),
				nil
		}
		return values.NewSQLTimestamp(sqlValueKind, time.Date(y,
				time.Month(m),
				d,
				0,
				0,
				0,
				0,
				schema.DefaultLocale)),
			nil
	case Week:
		return values.NewSQLTimestamp(sqlValueKind, t.AddDate(0, 0, v*7)), nil
	case Day:
		return values.NewSQLTimestamp(sqlValueKind, t.AddDate(0, 0, v)), nil
	case Hour:
		duration := time.Duration(v) * time.Hour
		return values.NewSQLTimestamp(sqlValueKind, t.Add(duration)), nil
	case Minute:
		duration := time.Duration(v) * time.Minute
		return values.NewSQLTimestamp(sqlValueKind, t.Add(duration)), nil
	case Second:
		// Seconds can actually be fractional rather than integer.
		duration := time.Duration(int64(values.Float64(vs[1]) * 1e9))
		return values.NewSQLTimestamp(sqlValueKind, t.Add(duration)), nil
	case Microsecond:
		duration := time.Duration(int64(values.Float64(vs[1]))) * time.Microsecond
		return values.NewSQLTimestamp(sqlValueKind, t.Add(duration).Round(time.Millisecond)), nil
	default:
		err := fmt.Errorf("cannot add '%v' to timestamp", vs[0])
		return values.NewSQLNull(sqlValueKind), err
	}
}

// nolint: unparam
func (f baseScalarFunctionExpr) timestampDiffEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	// These must be SQLTimestamps at this point. If they are not, something has
	// broken in the algebrizer. Check the handling of scalar functions in the
	// algebrizer.
	tmp1, ok := vs[1].(values.SQLTimestamp)
	if !ok {
		return nil,
			fmt.Errorf("unable to evaluate type %v in timestampdiff,"+
				" this points to an error in the algebrizer",
				vs[1])
	}
	t1 := values.Timestamp(tmp1)

	tmp2, ok := vs[2].(values.SQLTimestamp)
	if !ok {
		return nil, fmt.Errorf("unable to evaluate type %v in timestampdiff,"+
			" this points to an error in the algebrizer",
			vs[2])
	}
	t2 := values.Timestamp(tmp2)

	duration := t2.Sub(t1)

	switch vs[0].String() {
	case Year:
		return values.NewSQLInt64(sqlValueKind, int64(numMonths(t1, t2)/12)), nil
	case Quarter:
		return values.NewSQLInt64(sqlValueKind, int64(numMonths(t1, t2)/3)), nil
	case Month:
		return values.NewSQLInt64(sqlValueKind, int64(numMonths(t1, t2))), nil
	case Week:
		if t1.After(t2) {
			return values.NewSQLInt64(sqlValueKind, int64(math.Ceil((duration.Hours())/24/7))), nil
		}
		return values.NewSQLInt64(sqlValueKind, int64(math.Floor((duration.Hours())/24/7))), nil
	case Day:
		if t1.After(t2) {
			return values.NewSQLInt64(sqlValueKind, int64(math.Ceil(duration.Hours()/24))), nil
		}
		return values.NewSQLInt64(sqlValueKind, int64(math.Floor(duration.Hours()/24))), nil
	case Hour:
		return values.NewSQLInt64(sqlValueKind, int64(duration.Hours())), nil
	case Minute:
		return values.NewSQLInt64(sqlValueKind, int64(duration.Minutes())), nil
	case Second:
		return values.NewSQLInt64(sqlValueKind, int64(duration.Seconds())), nil
	case Microsecond:
		return values.NewSQLInt64(sqlValueKind, duration.Nanoseconds()/1000), nil
	default:
		err := fmt.Errorf("cannot add '%v' to timestamp", vs[0])
		return values.NewSQLNull(sqlValueKind), err
	}
}

// nolint: unparam
func (f baseScalarFunctionExpr) timestampEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	t, _, ok := values.ParseDateTime(vs[0].String())
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	t = t.In(schema.DefaultLocale)

	if len(vs) == 1 {
		panic(fmt.Errorf("timestampEvaluate should only be called with 2 arguments"))
	}

	d, ok := parseDuration(vs[1])
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	t = t.Add(d).Round(time.Microsecond)

	return values.NewSQLTimestamp(sqlValueKind, t), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) toDaysEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	// This must be a SQLDate at this point. If it is not, the algebrizer has been broken.
	// Check the handling of scalar functions in the algebrizer.
	tmp, ok := vs[0].(values.SQLDate)
	if !ok {
		return nil, fmt.Errorf("unable to evaluate input value %v in to_days", vs[0])
	}
	date := values.Timestamp(tmp)

	// First compute the days from YearOne.
	target, err := daysFromYearZeroCalculation(date)
	if err != nil {
		return nil, err
	}

	return values.NewSQLInt64(sqlValueKind, int64(target)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) toSecondsEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	// This must be a SQLTimestamp at this point. If it is not, the algebrizer has been broken.
	// Check the handling of scalar functions in the algebrizer.
	tmp, ok := vs[0].(values.SQLTimestamp)
	if !ok {
		return nil, fmt.Errorf("unable to evaluate input value %v in to_seconds", vs[0])
	}
	date := values.Timestamp(tmp)

	// First compute the days from YearOne and convert to seconds.
	target, err := daysFromYearZeroCalculation(date)
	if err != nil {
		return nil, err
	}
	target *= secondsPerDay

	// Now add remainder hours, minutes, and seconds.
	target += float64(date.Hour())*secondsPerHour +
		float64(date.Minute())*secondsPerMinute + float64(date.Second())

	// target is now seconds since dayOne.
	return values.NewSQLInt64(sqlValueKind, int64(target)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) trimEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	value := vs[0].String()
	end := "both"
	toTrim := " "
	if len(vs) == 3 {
		end = vs[1].String()
		toTrim = vs[2].String()
	}

	save := ""
	for save != value {
		save = value
		switch end {
		case "both":
			value = strings.TrimPrefix(value, toTrim)
			value = strings.TrimSuffix(value, toTrim)
		case "leading":
			value = strings.TrimPrefix(value, toTrim)
		case "trailing":
			value = strings.TrimSuffix(value, toTrim)
		}
	}

	return values.NewSQLVarchar(sqlValueKind, value), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) truncateEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	var truncated float64
	x := values.Float64(vs[0])
	d := values.Float64(vs[1])

	if d >= 0 {
		pow := math.Pow(10, d)
		i, _ := math.Modf(x * pow)
		truncated = i / pow
	} else {
		pow := math.Pow(10, math.Abs(d))
		i, _ := math.Modf(x / pow)
		truncated = i * pow
	}

	return values.NewSQLFloat(sqlValueKind, truncated), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) ucaseEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	value := strings.ToUpper(vs[0].String())

	return values.NewSQLVarchar(sqlValueKind, value), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) unixTimestampEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}

	now := time.Now()

	if len(vs) == 0 {
		return values.NewSQLUint64(sqlValueKind, uint64(now.Unix())), nil
	}

	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	t, _, ok := values.ParseDateTime(vs[0].String())
	if !ok || t.Before(epoch) {
		return values.NewSQLFloat(sqlValueKind, 0.0), nil
	}

	// Our times are parsed as if in UTC. However, we need to
	// parse it in the actual location the server's running
	// in - to account for any timezone difference.
	y, m, d := t.Date()
	ts := time.Date(y, m, d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), now.Location())
	return values.NewSQLUint64(sqlValueKind, uint64(ts.Unix())), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) weekEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}
	check, ok := vs[0].(values.SQLDate)
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	dateArg := values.Timestamp(check)
	// Mode should always be less than MAX_INT.
	mode := int(values.Int64(vs[1]))

	ret := weekCalculation(dateArg, mode)
	if ret == -1 {
		return values.NewSQLNull(sqlValueKind), nil
	}
	return values.NewSQLInt64(sqlValueKind, int64(ret)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) weekdayEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	t, _, ok := values.ParseDateTime(vs[0].String())
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	w := int(t.Weekday())
	if w == 0 {
		w = 7
	}
	return values.NewSQLInt64(sqlValueKind, int64(w-1)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) yearEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	t, _, ok := values.ParseDateTime(vs[0].String())
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	return values.NewSQLInt64(sqlValueKind, int64(t.Year())), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) yearWeekEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	if values.HasNullValue(vs...) {
		return values.NewSQLNull(sqlValueKind), nil
	}
	check, ok := vs[0].(values.SQLDate)
	if !ok {
		return values.NewSQLNull(sqlValueKind), nil
	}

	date := values.Timestamp(check)
	year := date.Year()
	// Mode should always be less than MAX_INT.
	mode := int(values.Int64(vs[1]))

	var week int

	// Unlike WEEK, YEARWEEK always uses the 1-53 modes. Thus
	// we always call week with the 1-53 of a 0-53, 1-53 pair.
	switch mode {

	// First day of week: Sunday, with a Sunday in this year.
	case 0, 2:
		week = weekCalculation(date, 2)
	// First day of week: Monday, with 4 days in this year.
	case 1, 3:
		week = weekCalculation(date, 3)
	// First day of week: Sunday, with 4 days in this year.
	case 4, 6:
		week = weekCalculation(date, 6)
	// First day of week: Monday, with a Monday in this year.
	case 5, 7:
		week = weekCalculation(date, 7)
	}

	if week == -1 {
		return values.NewSQLNull(sqlValueKind), nil
	}

	switch week {
	case 1:
		if date.Month() == 12 {
			year++
		}
	case 52, 53:
		if date.Month() == 1 {
			year--
		}

	}
	return values.NewSQLInt64(sqlValueKind, int64(year*100+week)), nil
}

// atan, atan2

// nolint: unparam
func (f baseScalarFunctionExpr) absEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(sqlValueKind, vs, math.Abs)
}

// nolint: unparam
func (f baseScalarFunctionExpr) acosEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(sqlValueKind, vs, math.Acos)
}

// nolint: unparam
func (f baseScalarFunctionExpr) asinEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(sqlValueKind, vs, math.Asin)
}

// nolint: unparam
func (f baseScalarFunctionExpr) ceilEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(sqlValueKind, vs, math.Ceil)
}

// nolint: unparam
func (f baseScalarFunctionExpr) cosEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(sqlValueKind, vs, math.Cos)
}

// nolint: unparam
func (f baseScalarFunctionExpr) expEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(sqlValueKind, vs, math.Exp)
}

// nolint: unparam
func (f baseScalarFunctionExpr) floorEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(sqlValueKind, vs, math.Floor)
}

// nolint: unparam
func (f baseScalarFunctionExpr) radiansEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(sqlValueKind, vs, toRadians)
}

// nolint: unparam
func (f baseScalarFunctionExpr) degreesEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(sqlValueKind, vs, toDegrees)
}

// nolint: unparam
func (f baseScalarFunctionExpr) sinEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(sqlValueKind, vs, math.Sin)
}

// nolint: unparam
func (f baseScalarFunctionExpr) sqrtEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(sqlValueKind, vs, math.Sqrt)
}

// nolint: unparam
func (f baseScalarFunctionExpr) tanEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(sqlValueKind, vs, math.Tan)
}

// nolint: unparam
func (f baseScalarFunctionExpr) atanSingleArgEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(sqlValueKind, vs, math.Atan)
}

// nolint: unparam
func (f baseScalarFunctionExpr) atanDualArgEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, vs []values.SQLValue) (values.SQLValue, error) {
	return f.dualArgFloatMathFuncEvaluate(sqlValueKind, vs, math.Atan2)
}

// nolint: unparam
func (f baseScalarFunctionExpr) atan2SingleArgEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, values []values.SQLValue) (values.SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(sqlValueKind, values, math.Atan)
}

// nolint: unparam
func (f baseScalarFunctionExpr) atan2DualArgEvaluate(sqlValueKind values.SQLValueKind, _ *collation.Collation, values []values.SQLValue) (values.SQLValue, error) {
	return f.dualArgFloatMathFuncEvaluate(sqlValueKind, values, math.Atan2)
}
