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

	"github.com/10gen/sqlproxy/internal/mathutil"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

// nolint: unparam
func (f baseScalarFunctionExpr) asciiEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, EvalInt64), nil
	}

	str := values[0].String()
	if str == "" {
		return NewSQLInt64(cfg.sqlValueKind, 0), nil
	}

	c := str[0]

	return NewSQLInt64(cfg.sqlValueKind, int64(c)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) charEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {

	var b []byte
	for _, i := range values {
		if i.IsNull() {
			continue
		}
		v := Int64(i)
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

	return NewSQLVarchar(cfg.sqlValueKind, string(b)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) characterLengthEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, EvalInt64), nil
	}

	value := []rune(values[0].String())

	return NewSQLInt64(cfg.sqlValueKind, int64(len(value))), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) coalesceEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	for _, value := range values {
		if !value.IsNull() {
			return value, nil
		}
	}

	return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) concatEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (v SQLValue, err error) {
	if hasNullValue(values...) {
		v = NewSQLNull(cfg.sqlValueKind, EvalString)
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
	for _, value := range values {
		b.WriteString(value.String())
	}

	v = NewSQLVarchar(cfg.sqlValueKind, b.String())
	err = nil
	return
}

// nolint: unparam
func (f baseScalarFunctionExpr) concatWsEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (v SQLValue, err error) {
	if values[0].IsNull() {
		v = NewSQLNull(cfg.sqlValueKind, f.EvalType())
		return
	}

	defer func() {
		if r := recover(); r != nil {
			v = nil
			err = fmt.Errorf("%v", r)
		}
	}()

	var b bytes.Buffer
	separator := values[0].String()
	trimValues := values[1:]
	for i, value := range trimValues {
		if value.IsNull() {
			continue
		}
		b.WriteString(value.String())
		if i != len(trimValues)-1 {
			b.WriteString(separator)
		}
	}

	v = NewSQLVarchar(cfg.sqlValueKind, b.String())
	return
}

// nolint: unparam
func (f baseScalarFunctionExpr) connectionIDEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return NewSQLUint64(cfg.sqlValueKind, cfg.connID), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) piEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.constantEvaluate(ctx, cfg, st, values, math.Pi, EvalDouble)
}

// nolint: unparam
func (f baseScalarFunctionExpr) constantEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue, constVal interface{}, constEvalType EvalType) (SQLValue, error) {
	sqlVal := GoValueToSQLValue(cfg.sqlValueKind, constVal)
	if sqlVal.EvalType() != constEvalType {
		err := fmt.Errorf(
			"actual EvalType %x did not match declared EvalType %x",
			sqlVal.EvalType(), constEvalType,
		)
		return nil, err
	}
	return sqlVal, nil
}

// Diverges from MySQL behavior in its handling of negative values
// Converts bases to positive numbers, and returns a negative value if the input is negative
// MySQL claims that "If from_base is a negative number, N is regarded as a signed number.
// Otherwise, N is treated as unsigned." Manual testing shows that it returns the 2's
// complement version if the number is negative unless the to_base is also negative, in which
// case it returns the number with a negative sign at the front
// nolint: unparam
func (f baseScalarFunctionExpr) convEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	num := values[0].String()
	originalBase := absInt64(Int64(values[1]))
	newBase := absInt64(Int64(values[2]))
	negative := false

	if baseIsInvalid(originalBase) || baseIsInvalid(newBase) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
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
		return NewSQLVarchar(cfg.sqlValueKind, "0"), nil
	}
	strVersion := strconv.FormatInt(base10Version, int(newBase))

	if negative && strVersion != "0" {
		strVersion = "-" + strVersion
	}

	return NewSQLVarchar(cfg.sqlValueKind, strings.ToUpper(strVersion)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) convertEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if values[0].IsNull() {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	typ, ok := sqlTypeFromSQLExpr(values[1])
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	return NewSQLConvertExpr(values[0], typ).Evaluate(ctx, cfg, st)
}

// nolint: unparam
func (f baseScalarFunctionExpr) cotEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	tan := math.Tan(Float64(values[0]))
	if tan == 0 {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()),
			mysqlerrors.Defaultf(mysqlerrors.ErDataOutOfRange,
				"DOUBLE",
				fmt.Sprintf("'cot(%v)'",
					Float64(values[0])))
	}

	return NewSQLFloat(cfg.sqlValueKind, 1/tan), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) currentDateEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	now := time.Now().In(schema.DefaultLocale)
	t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, schema.DefaultLocale)
	return NewSQLDate(cfg.sqlValueKind, t), nil

}

// nolint: unparam
func (f baseScalarFunctionExpr) currentTimestampEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	value := time.Now().In(schema.DefaultLocale)
	return NewSQLTimestamp(cfg.sqlValueKind, value), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) curtimeEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return NewSQLTimestamp(cfg.sqlValueKind, time.Now().In(schema.DefaultLocale)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) dateAddEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	_, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	// Seconds can be fractional values, so our calculateInterval function will not work right
	// (it is fine for all other units, as they must be integral).
	if values[2].String() == Second {
		interval := values[1].SQLFloat()
		vals := []SQLExpr{NewSQLVarchar(cfg.sqlValueKind, Second), interval, values[0]}
		tsAdd, err := NewSQLScalarFunctionExpr("timestampadd", vals)
		if err != nil {
			return nil, err
		}
		return tsAdd.Evaluate(ctx, cfg, st)
	}

	args, neg := dateArithmeticArgs(values[2].String(), values[1])
	unit, interval, err := calculateInterval(values[2].String(), args, neg)
	if err != nil {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	vals := []SQLExpr{
		NewSQLVarchar(cfg.sqlValueKind, unit),
		NewSQLInt64(cfg.sqlValueKind, int64(interval)), values[0],
	}
	tsAdd, err := NewSQLScalarFunctionExpr("timestampadd", vals)
	if err != nil {
		return nil, err
	}
	return tsAdd.Evaluate(ctx, cfg, st)
}

// nolint: unparam
func (f baseScalarFunctionExpr) dateDiffEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {

	var left, right time.Time
	var ok bool

	parseArgs := func(val SQLValue) (time.Time, bool) {
		var date time.Time

		date, _, ok = strToDateTime(val.String(), false)
		if !ok {
			return date, false
		}

		date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, schema.DefaultLocale)
		return date, true
	}

	if left, ok = parseArgs(values[0]); !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	if right, ok = parseArgs(values[1]); !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	durationDiff := left.Sub(right)
	hoursDiff := durationDiff.Hours()
	daysDiff := hoursDiff / 24

	diff := NewSQLInt64(cfg.sqlValueKind, int64(daysDiff))
	return diff, nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) dateFormatEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	date, _, ok := parseDateTime(values[0].String())
	date = date.In(schema.DefaultLocale)
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	v1, ok := values[1].(SQLVarchar)
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	ret, err := formatDate(ctx, cfg, st, date, v1.String())
	if err != nil {
		return nil, err
	}
	return NewSQLVarchar(cfg.sqlValueKind, ret), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) dateEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	// Too-short numbers are padded differently than too-short strings.
	// strToDateTime (called by parseDateTime) handles padding in the too-short
	// string case. We need to fix the string here, where we can still find out
	// the original input type.
	var str string
	switch values[0].(type) {
	case SQLFloat, SQLDecimal128, SQLInt64:
		noDecimal := strings.Split(values[0].String(), ".")[0]
		intLength := len(noDecimal)
		if intLength > 14 {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
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
		str = strings.Repeat("0", padLen) + noDecimal
	default:
		str = values[0].String()
	}

	t, _, ok := parseDateTime(str)
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	return NewSQLDate(cfg.sqlValueKind, t.Truncate(24*time.Hour)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) dateSubEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	v := values[1].String()
	if string(v[0]) != "-" {
		v = "-" + v
	} else {
		v = v[1:]
	}

	vals := []SQLExpr{values[0], NewSQLVarchar(cfg.sqlValueKind, v), values[2]}
	dateadd, err := NewSQLScalarFunctionExpr("date_add", vals)
	if err != nil {
		return nil, err
	}

	return dateadd.Evaluate(ctx, cfg, st)
}

// nolint: unparam
func (f baseScalarFunctionExpr) dayNameEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	return NewSQLVarchar(cfg.sqlValueKind, t.Weekday().String()), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) dayOfMonthEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.Day())), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) dayOfWeekEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.Weekday())+1), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) dayOfYearEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.YearDay())), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) databaseEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return NewSQLVarchar(cfg.sqlValueKind, cfg.dbName), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) dualArgFloatMathFuncEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue, fn func(float64, float64) float64) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	result := fn(Float64(values[0]), Float64(values[1]))
	if math.IsNaN(result) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}
	if math.IsInf(result, 0) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}
	if result == -0 {
		result = 0
	}
	return NewSQLFloat(cfg.sqlValueKind, result), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) eltEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {

	if hasNullValue(values[0]) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	idx := Int64(values[0])
	if idx <= 0 || int(idx) >= len(values) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	result := values[idx]
	if result.IsNull() {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	return NewSQLVarchar(cfg.sqlValueKind, result.String()), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) extractEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[1].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	units := [6]int{t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second()}

	var unitStrs [6]string
	// For certain units, we need to concatenate the unit values as strings
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

	switch values[0].String() {
	case Year:
		return NewSQLInt64(cfg.sqlValueKind, int64(units[0])), nil
	case Quarter:
		return NewSQLInt64(cfg.sqlValueKind, int64(math.Ceil(float64(units[1])/3.0))), nil
	case Month:
		return NewSQLInt64(cfg.sqlValueKind, int64(units[1])), nil
	case Week:
		_, w := t.ISOWeek()
		return NewSQLInt64(cfg.sqlValueKind, int64(w)), nil
	case Day:
		return NewSQLInt64(cfg.sqlValueKind, int64(units[2])), nil
	case Hour:
		return NewSQLInt64(cfg.sqlValueKind, int64(units[3])), nil
	case Minute:
		return NewSQLInt64(cfg.sqlValueKind, int64(units[4])), nil
	case Second:
		return NewSQLInt64(cfg.sqlValueKind, int64(units[5])), nil
	case Microsecond:
		return NewSQLInt64(cfg.sqlValueKind, 0), nil
	case YearMonth:
		ym, _ := strconv.ParseInt(unitStrs[0]+unitStrs[1], 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, ym), nil
	case DayHour:
		dh, _ := strconv.ParseInt(unitStrs[2]+unitStrs[3], 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, dh), nil
	case DayMinute:
		dm, _ := strconv.ParseInt(unitStrs[2]+unitStrs[3]+unitStrs[4], 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, dm), nil
	case DaySecond:
		ds, _ := strconv.ParseInt(unitStrs[2]+unitStrs[3]+unitStrs[4]+unitStrs[5], 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, ds), nil
	case DayMicrosecond:
		dms, _ := strconv.ParseInt(unitStrs[2]+unitStrs[3]+unitStrs[4]+unitStrs[5]+"000000", 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, dms), nil
	case HourMinute:
		hm, _ := strconv.ParseInt(unitStrs[3]+unitStrs[4], 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, hm), nil
	case HourSecond:
		hs, _ := strconv.ParseInt(unitStrs[3]+unitStrs[4]+unitStrs[5], 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, hs), nil
	case HourMicrosecond:
		hms, _ := strconv.ParseInt(unitStrs[3]+unitStrs[4]+unitStrs[5]+"000000", 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, hms), nil
	case MinuteSecond:
		ms, _ := strconv.ParseInt(unitStrs[4]+unitStrs[5], 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, ms), nil
	case MinuteMicrosecond:
		mms, _ := strconv.ParseInt(unitStrs[4]+unitStrs[5]+"000000", 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, mms), nil
	case SecondMicrosecond:
		sms, _ := strconv.ParseInt(unitStrs[5]+"000000", 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, sms), nil
	default:
		err := fmt.Errorf("unit type '%v' is not supported", values[0].String())
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), err
	}
}

// nolint: unparam
func (f baseScalarFunctionExpr) fieldEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLInt64(cfg.sqlValueKind, 0), nil
	}

	target := values[0]
	candidates := values[1:]

	for idx, candidate := range candidates {
		if candidate == target {
			return NewSQLInt64(cfg.sqlValueKind, int64(idx+1)), nil
		}
	}
	return NewSQLInt64(cfg.sqlValueKind, 0), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) fromDaysEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
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

	v := values[0].String()
	neg := len(v) > 0 && v[0] == '-'
	if neg {
		v = v[1:]
	}
	value, err := strconv.ParseFloat(parseNumeric(v), 64)
	if err != nil {
		return NewSQLVarchar(cfg.sqlValueKind, "0000-00-00"), nil
	}
	if neg {
		value = -value
	}

	if value <= 365.5 || value >= 3652499.5 {
		// Go's zero time starts January 1, year 1, 00:00:00 UTC
		// and thus can not represent the date "0000-00-00". To
		// handle this, we return a varchar instead
		return NewSQLVarchar(cfg.sqlValueKind, "0000-00-00"), nil
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

	return NewSQLDate(cfg.sqlValueKind, date.In(schema.DefaultLocale)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) fromUnixtimeEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	value := round(Float64(values[0]))
	if value < 0 {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	date := time.Unix(value, 0).In(schema.DefaultLocale)
	if len(values) == 1 {
		return NewSQLTimestamp(cfg.sqlValueKind, date), nil
	}
	ret, err := formatDate(ctx, cfg, st, date, values[1].String())
	if err != nil {
		return nil, err
	}
	return NewSQLVarchar(cfg.sqlValueKind, ret), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) greatestEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	valExprs := []SQLExpr{}
	for _, val := range values {
		valExprs = append(valExprs, val)
	}
	convertTo := preferentialType(valExprs...)

	convertedVals := []SQLValue{}
	for _, val := range values {
		newVal := ConvertTo(val, convertTo)
		convertedVals = append(convertedVals, newVal)
	}

	var greatest SQLValue
	var greatestIdx int

	c, err := CompareTo(convertedVals[0], convertedVals[1], st.collation)
	if err != nil {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), err
	}

	if c == -1 {
		greatest, greatestIdx = values[1], 1
	} else {
		greatest, greatestIdx = values[0], 0
	}

	for i := 2; i < len(values); i++ {
		c, err = CompareTo(greatest, convertedVals[i], st.collation)
		if err != nil {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType()), err
		}
		if c == -1 {
			greatest, greatestIdx = values[i], i
		}
	}

	allTimeVals, timestamp := areAllTimeTypes(values)
	if allTimeVals && timestamp {
		t, _, _ := parseDateTime(values[greatestIdx].String())
		return NewSQLTimestamp(cfg.sqlValueKind, t), nil
	} else if convertTo == EvalDate || convertTo == EvalDatetime {
		return values[greatestIdx], nil
	}

	return convertedVals[greatestIdx], nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) hourEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if values[0].IsNull() {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}
	_, hour, ok := parseTime(values[0].String())
	if !ok {
		// If we managed to parse, but minutes or seconds are >= 60
		// MySQL returns NULL for the hour/minute/second function.
		// Rather than return yet another value, we coop the hour value
		// and return -1, thus we can check for -1 here to return NULL
		// rather than the 0 expected if the string could not be parsed
		// at all.
		if hour == -1 {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
		}
		return NewSQLInt64(cfg.sqlValueKind, 0), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(hour)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) insertEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {

	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	s := values[0].String()
	pos := int(round(Float64(values[1]))) - 1
	length := int(round(Float64(values[2])))
	newstr := values[3].String()

	if pos < 0 || pos >= len(s) {
		return values[0], nil
	}

	if pos+length < 0 || pos+length > len(s) {
		return NewSQLVarchar(cfg.sqlValueKind, s[:pos]+newstr), nil
	}

	return NewSQLVarchar(cfg.sqlValueKind, s[:pos]+newstr+s[pos+length:]), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) instrEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	locate, err := NewSQLScalarFunctionExpr("locate", []SQLExpr{values[1], values[0]})
	if err != nil {
		return nil, err
	}
	return locate.Evaluate(ctx, cfg, st)
}

// nolint: unparam
func (f baseScalarFunctionExpr) intervalEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if values[0].IsNull() {
		return NewSQLInt64(cfg.sqlValueKind, -1), nil
	}

	start, end := 1, len(values)-1
	for start != end {
		mid := (start + end + 1) / 2
		if values[mid].IsNull() || Float64(values[mid]) <= Float64(values[0]) {
			start = mid
		} else {
			end = mid - 1
		}
	}

	if values[start].IsNull() || Float64(values[start]) <= Float64(values[0]) {
		return NewSQLInt64(cfg.sqlValueKind, int64(start)), nil
	}
	return NewSQLInt64(cfg.sqlValueKind, int64(start-1)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) lastDayEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	// Must be a SQLTimestamp at this point. If it is not, the algebrizer
	// has been broken. Check where the algebrizer handles
	// scalar functions.
	tmp, ok := values[0].(SQLDate)
	if !ok {
		return nil,
			fmt.Errorf("unable to evaluate type %v"+
				" in to_days, this points to an error in the algebrizer",
				values[0])
	}
	t := Timestamp(tmp)
	year, month, _ := t.Date()
	first := time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
	return NewSQLDate(cfg.sqlValueKind, first.AddDate(0, 1, -1)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) lcaseEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	value := strings.ToLower(values[0].String())

	return NewSQLVarchar(cfg.sqlValueKind, value), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) leastEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	valExprs := []SQLExpr{}
	for _, val := range values {
		valExprs = append(valExprs, val)
	}
	convertTo := preferentialType(valExprs...)

	convertedVals := []SQLValue{}
	for _, val := range values {
		newVal := ConvertTo(val, convertTo)
		convertedVals = append(convertedVals, newVal)
	}

	var least SQLValue
	var leastIdx int

	c, err := CompareTo(convertedVals[0], convertedVals[1], st.collation)
	if err != nil {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), err
	}

	if c == -1 {
		least, leastIdx = convertedVals[0], 0
	} else {
		least, leastIdx = convertedVals[1], 1
	}

	for i := 2; i < len(values); i++ {
		c, err = CompareTo(least, convertedVals[i], st.collation)
		if err != nil {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType()), err
		}
		if c == 1 {
			least, leastIdx = values[i], i
		}
	}

	allTimeVals, timestamp := areAllTimeTypes(values)
	if allTimeVals && timestamp {
		t, _, _ := parseDateTime(values[leastIdx].String())
		return NewSQLTimestamp(cfg.sqlValueKind, t), nil
	} else if convertTo == EvalDate || convertTo == EvalDatetime {
		return values[leastIdx], nil
	}

	return convertedVals[leastIdx], nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) leftEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	substring,
		err := NewSQLScalarFunctionExpr("substring",
		[]SQLExpr{values[0],
			NewSQLInt64(cfg.sqlValueKind, 1),
			values[1]})
	if err != nil {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), err
	}
	return substring.Evaluate(ctx, cfg, st)
}

// nolint: unparam
func (f baseScalarFunctionExpr) lengthEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	value := values[0].String()

	return NewSQLInt64(cfg.sqlValueKind, int64(len(value))), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) locateEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values[:2]...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	substr := []rune(values[0].String())
	str := []rune(values[1].String())
	var result int
	if len(values) == 3 {

		pos := int(Float64(values[2])+0.5) - 1 // MySQL uses 1 as a basis

		if pos < 0 || len(str) <= pos {
			return NewSQLInt64(cfg.sqlValueKind, 0), nil
		}
		str = str[pos:]
		result = runesIndex(str, substr)
		if result >= 0 {
			result += pos
		}
	} else {
		result = runesIndex(str, substr)
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(result+1)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) logEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.logarithmEvaluate(ctx, cfg, st, values, 0)
}

// nolint: unparam
func (f baseScalarFunctionExpr) lnEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.logarithmEvaluate(ctx, cfg, st, values, 0)
}

// nolint: unparam
func (f baseScalarFunctionExpr) log2Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.logarithmEvaluate(ctx, cfg, st, values, 2)
}

// nolint: unparam
func (f baseScalarFunctionExpr) log10Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.logarithmEvaluate(ctx, cfg, st, values, 10)
}

// nolint: unparam
func (f baseScalarFunctionExpr) logarithmEvaluate(_ context.Context, cfg *ExecutionConfig, _ *ExecutionState, values []SQLValue, base uint32) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	var result float64
	switch base {
	case 0:
		if len(values) == 2 {
			// arbitrary base
			result = math.Log(Float64(values[1])) / math.Log(Float64(values[0]))
		} else {
			// natural base
			result = math.Log(Float64(values[0]))
		}
	case 2:
		result = math.Log2(Float64(values[0]))
	case 10:
		result = math.Log10(Float64(values[0]))
	}
	if math.IsNaN(result) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}
	if math.IsInf(result, 0) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}
	if result == -0 {
		result = 0
	}
	return NewSQLFloat(cfg.sqlValueKind, result), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) lpadEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return handlePadding(cfg.sqlValueKind, values, true)
}

// nolint: unparam
func (f baseScalarFunctionExpr) ltrimEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	value := strings.TrimLeft(values[0].String(), " ")

	return NewSQLVarchar(cfg.sqlValueKind, value), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) makeDateEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	// Floating arguments should be rounded.
	y := round(Float64(values[0]))
	if y < 0 || y > 9999 {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}
	if y >= 0 && y <= 69 {
		y += 2000
	} else if y >= 70 && y <= 99 {
		y += 1900
	}

	d := round(Float64(values[1]))

	if d <= 0 {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	t := time.Date(int(y), 1, 0, 0, 0, 0, 0, schema.DefaultLocale)
	duration := time.Duration(d*24) * time.Hour

	output := t.Add(duration)
	if output.Year() > 9999 {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	return NewSQLDate(cfg.sqlValueKind, output), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) md5Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	h := md5.New()
	_, err := io.WriteString(h, values[0].String())
	if err != nil {
		return nil, err
	}
	return NewSQLVarchar(cfg.sqlValueKind, fmt.Sprintf("%x", h.Sum(nil))), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) microsecondEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {

	arg := values[0]

	if arg.IsNull() {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	str := arg.String()
	if str == "" {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	t, _, ok := parseTime(str)
	if !ok {
		return NewSQLInt64(cfg.sqlValueKind, 0), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.Nanosecond()/1000)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) minuteEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if values[0].IsNull() {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	t, hour, ok := parseTime(values[0].String())
	if !ok {
		// If we managed to parse, but minutes or seconds are >= 60
		// MySQL returns NULL for the hour/minute/second function.
		// Rather than return yet another value, we coop the hour value
		// and return -1, thus we can check for -1 here to return NULL
		// rather than the 0 expected if the string could not be parsed
		// at all.
		if hour == -1 {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
		}
		return NewSQLInt64(cfg.sqlValueKind, 0), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.Minute())), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) modEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.dualArgFloatMathFuncEvaluate(ctx, cfg, st, values, math.Mod)
}

// nolint: unparam
func (f baseScalarFunctionExpr) monthEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.Month())), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) monthNameEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	return NewSQLVarchar(cfg.sqlValueKind, t.Month().String()), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) nopushdownEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return values[0], nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) powEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	v0 := Float64(values[0])
	v1 := Float64(values[1])

	n := math.Pow(v0, v1)
	zeroBaseExpNeg := v0 == 0 && v1 < 0
	if math.IsNaN(n) || zeroBaseExpNeg {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()),
			mysqlerrors.Defaultf(mysqlerrors.ErDataOutOfRange,
				"DOUBLE",
				fmt.Sprintf("pow(%v,%v)",
					Float64(values[0]),
					Float64(values[1])))
	}

	return NewSQLFloat(cfg.sqlValueKind, math.Pow(v0, v1)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) quarterEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
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

	return NewSQLInt64(cfg.sqlValueKind, int64(q)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) randEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	uniqueID := Uint64(values[0])

	if len(values) == 2 {
		seed := round(Float64(values[1]))
		r := st.RandomWithSeed(uniqueID, seed)
		return NewSQLFloat(cfg.sqlValueKind, r.Float64()), nil
	}

	r := st.Random(uniqueID)
	return NewSQLFloat(cfg.sqlValueKind, r.Float64()), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) repeatEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (v SQLValue, err error) {
	if hasNullValue(values...) {
		v = NewSQLNull(cfg.sqlValueKind, f.EvalType())
		err = nil
		return
	}

	str := values[0].String()
	if len(str) < 1 {
		v = NewSQLVarchar(cfg.sqlValueKind, "")
		err = nil
		return
	}

	rep := int(roundToDecimalPlaces(0, Float64(values[1])))
	if rep < 1 {
		v = NewSQLVarchar(cfg.sqlValueKind, "")
		err = nil
		return
	}

	var b bytes.Buffer
	for i := 0; i < rep; i++ {
		b.WriteString(str)
	}

	v = NewSQLVarchar(cfg.sqlValueKind, b.String())
	err = nil
	return
}

// nolint: unparam
func (f baseScalarFunctionExpr) replaceEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	s := values[0].String()
	old := values[1].String()
	new := values[2].String()

	return NewSQLVarchar(cfg.sqlValueKind, strings.Replace(s, old, new, -1)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) reverseEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}
	s := values[0].String()
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return NewSQLVarchar(cfg.sqlValueKind, string(runes)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) rightEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	str := values[0].String()
	posFloat := Float64(values[1])

	if posFloat > float64(len(str)) {
		return NewSQLVarchar(cfg.sqlValueKind, str), nil
	}

	startPos := math.Min(0, -1.0*posFloat)

	substring,
		err := NewSQLScalarFunctionExpr("substring",
		[]SQLExpr{values[0],
			NewSQLFloat(cfg.sqlValueKind, startPos)})
	if err != nil {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), err
	}

	return substring.Evaluate(ctx, cfg, st)
}

// nolint: unparam
func (f baseScalarFunctionExpr) roundEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	base := Float64(values[0])

	var decimal int64
	if len(values) == 2 {
		decimal = Int64(values[1])

		if decimal < 0 {
			return NewSQLFloat(cfg.sqlValueKind, 0), nil
		}
	} else {
		decimal = 0
	}

	rounded := roundToDecimalPlaces(decimal, base)

	return NewSQLFloat(cfg.sqlValueKind, rounded), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) rpadEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return handlePadding(cfg.sqlValueKind, values, false)
}

// nolint: unparam
func (f baseScalarFunctionExpr) rtrimEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	value := strings.TrimRight(values[0].String(), " ")

	return NewSQLVarchar(cfg.sqlValueKind, value), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) secondEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if values[0].IsNull() {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	t, hour, ok := parseTime(values[0].String())
	if !ok {
		// If we managed to parse, but minutes or seconds are >= 60
		// MySQL returns NULL for the hour/minute/second function.
		// Rather than return yet another value, we coop the hour value
		// and return -1, thus we can check for -1 here to return NULL
		// rather than the 0 expected if the string could not be parsed
		// at all.
		if hour == -1 {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
		}
		return NewSQLInt64(cfg.sqlValueKind, 0), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.Second())), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) signEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	v := Float64(values[0])
	// Positive numbers are more common than negative in most data sets.
	if v > 0 {
		return NewSQLInt64(cfg.sqlValueKind, 1), nil
	}
	if v < 0 {
		return NewSQLInt64(cfg.sqlValueKind, -1), nil
	}
	return NewSQLInt64(cfg.sqlValueKind, 0), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) singleArgFloatMathFuncEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue, fn func(float64) float64) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	result := fn(Float64(values[0]))
	if math.IsNaN(result) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}
	if math.IsInf(result, 0) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}
	if result == -0 {
		result = 0
	}
	return NewSQLFloat(cfg.sqlValueKind, result), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) sleepEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {

	err := mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, "sleep")

	if hasNullValue(values...) {
		return nil, err
	}

	n := Float64(values[0])

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

	return NewSQLInt64(cfg.sqlValueKind, 0), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) spaceEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	flt := Float64(values[0])
	n := round(flt)
	if n < 1 {
		return NewSQLVarchar(cfg.sqlValueKind, ""), nil
	}

	return NewSQLVarchar(cfg.sqlValueKind, strings.Repeat(" ", int(n))), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) strToDateEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	str, ok := values[0].(SQLVarchar)
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	ft, ok := values[1].(SQLVarchar)
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
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
					return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
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
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	if ts {
		return NewSQLTimestamp(cfg.sqlValueKind, d), nil
	}

	return NewSQLDate(cfg.sqlValueKind, d), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) midEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.substringEvaluate(ctx, cfg, st, values)
}

// nolint: unparam
func (f baseScalarFunctionExpr) substringEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	str := []rune(values[0].String())
	if values[0].String() == "" {
		return NewSQLVarchar(cfg.sqlValueKind, ""), nil
	}

	posFloat := Float64(values[1])
	var pos int
	if posFloat >= 0 {
		pos = int(posFloat + 0.5)
	} else {
		pos = int(posFloat - 0.5)
	}

	if pos > len(str) || pos == 0 {
		return NewSQLVarchar(cfg.sqlValueKind, ""), nil
	} else if pos < 0 {
		pos = len(str) + pos

		if pos < 0 {
			return NewSQLVarchar(cfg.sqlValueKind, ""), nil
		}
	} else {
		pos-- // MySQL uses 1 as a basis
	}

	if len(values) == 3 {
		length := int(Float64(values[2]) + 0.5)
		if length < 1 {
			return NewSQLVarchar(cfg.sqlValueKind, ""), nil
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
	return NewSQLVarchar(cfg.sqlValueKind, string(str)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) substringIndexEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {

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

	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	r := []rune(values[0].String())
	delim := []rune(values[1].String())

	count := int(round(Float64(values[2])))

	if count == 0 {
		return NewSQLVarchar(cfg.sqlValueKind, ""), nil
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

	return NewSQLVarchar(cfg.sqlValueKind, string(r)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) timeDiffEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	expr1, _, ok := parseTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	expr2, _, ok := parseTime(values[1].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
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
		return NewSQLVarchar(cfg.sqlValueKind, "00:00:00.000000"), nil
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

	return NewSQLVarchar(cfg.sqlValueKind, string(buf[w:])), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) timeToSecEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	dateParts := strings.Split(values[0].String(), " ")
	hasDatePart := len(dateParts) == 2
	components := strings.Split(values[0].String(), ":")
	if hasDatePart {
		components = strings.Split(dateParts[1], ":")
	}

	result, componentized := 0.0, true

	if len(components) == 1 {
		cmp, err := strconv.ParseFloat(components[0], 64)
		if err != nil {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType()), err
		}

		component := strconv.FormatFloat(math.Trunc(cmp), 'f', -1, 64)

		l := len(component)
		components, componentized = []string{"0", "0", "0"}, false

		// MySQL interprets abbreviated values without colons using the
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
			return NewSQLNull(cfg.sqlValueKind, f.EvalType()), err
		}

		cmp := math.Trunc(component)

		switch i {
		// more on valid time types at https://dev.mysql.com/doc/refman/5.7/en/time.html
		case 0:
			if cmp > 838 || cmp < -838 {
				if !componentized {
					return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
				}
				cmp = math.Copysign(838.0, cmp)
				components = []string{"", "59", "59"}
			}
		default:
			if cmp > 59 {
				return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
			}
		}

		signBit = signBit || math.Signbit(cmp)
		result += math.Abs(cmp) * (3600.0 / (math.Pow(60, float64(i))))
	}

	if signBit {
		return NewSQLFloat(cfg.sqlValueKind, math.Copysign(result, -1)), nil
	}

	return NewSQLFloat(cfg.sqlValueKind, result), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) timestampAddEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	v := int(round(Float64(values[1])))
	// values[2] must be a SQLTimestamp or the algebrizer has been broken.
	// Check the handling of scalar functions in the algebrizer.
	tmp, ok := values[2].(SQLTimestamp)
	if !ok {
		return nil,
			fmt.Errorf("unable to evaluate type %v in timestampadd,"+
				" this points to an error in the algebrizer",
				values[2])
	}
	t := Timestamp(tmp)

	ts := false
	if len(values[2].String()) > 10 {
		ts = true
	}

	switch values[0].String() {
	case Year:
		return NewSQLTimestamp(cfg.sqlValueKind, t.AddDate(v, 0, 0)), nil
	case Quarter:
		y, mp, d := t.Date()
		m := int(mp)
		interval := v * 3
		y += (m + interval - 1) / 12
		m = (m+interval-1)%12 + 1
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
			return NewSQLTimestamp(cfg.sqlValueKind, time.Date(y,
					time.Month(m),
					d,
					t.Hour(),
					t.Minute(),
					t.Second(),
					t.Nanosecond(),
					schema.DefaultLocale)),
				nil
		}
		return NewSQLTimestamp(cfg.sqlValueKind, time.Date(y,
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
		y += (m + interval - 1) / 12
		m = (m+interval-1)%12 + 1
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
			return NewSQLTimestamp(cfg.sqlValueKind, time.Date(y,
					time.Month(m),
					d,
					t.Hour(),
					t.Minute(),
					t.Second(),
					t.Nanosecond(),
					schema.DefaultLocale)),
				nil
		}
		return NewSQLTimestamp(cfg.sqlValueKind, time.Date(y,
				time.Month(m),
				d,
				0,
				0,
				0,
				0,
				schema.DefaultLocale)),
			nil
	case Week:
		return NewSQLTimestamp(cfg.sqlValueKind, t.AddDate(0, 0, v*7)), nil
	case Day:
		return NewSQLTimestamp(cfg.sqlValueKind, t.AddDate(0, 0, v)), nil
	case Hour:
		duration := time.Duration(v) * time.Hour
		return NewSQLTimestamp(cfg.sqlValueKind, t.Add(duration)), nil
	case Minute:
		duration := time.Duration(v) * time.Minute
		return NewSQLTimestamp(cfg.sqlValueKind, t.Add(duration)), nil
	case Second:
		// Seconds can actually be fractional rather than integer.
		duration := time.Duration(int64(Float64(values[1]) * 1e9))
		return NewSQLTimestamp(cfg.sqlValueKind, t.Add(duration)), nil
	case Microsecond:
		duration := time.Duration(int64(Float64(values[1]))) * time.Microsecond
		return NewSQLTimestamp(cfg.sqlValueKind, t.Add(duration).Round(time.Millisecond)), nil
	default:
		err := fmt.Errorf("cannot add '%v' to timestamp", values[0])
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), err
	}
}

// nolint: unparam
func (f baseScalarFunctionExpr) timestampDiffEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	// These must be SQLTimestamps at this point. If they are not, something has
	// broken in the algebrizer. Check the handling of scalar functions in the
	// algebrizer.
	tmp1, ok := values[1].(SQLTimestamp)
	if !ok {
		return nil,
			fmt.Errorf("unable to evaluate type %v in timestampdiff,"+
				" this points to an error in the algebrizer",
				values[1])
	}
	t1 := Timestamp(tmp1)

	tmp2, ok := values[2].(SQLTimestamp)
	if !ok {
		return nil, fmt.Errorf("unable to evaluate type %v in timestampdiff,"+
			" this points to an error in the algebrizer",
			values[2])
	}
	t2 := Timestamp(tmp2)

	duration := t2.Sub(t1)

	switch values[0].String() {
	case Year:
		return NewSQLInt64(cfg.sqlValueKind, int64(numMonths(t1, t2)/12)), nil
	case Quarter:
		return NewSQLInt64(cfg.sqlValueKind, int64(numMonths(t1, t2)/3)), nil
	case Month:
		return NewSQLInt64(cfg.sqlValueKind, int64(numMonths(t1, t2))), nil
	case Week:
		if t1.After(t2) {
			return NewSQLInt64(cfg.sqlValueKind, int64(math.Ceil((duration.Hours())/24/7))), nil
		}
		return NewSQLInt64(cfg.sqlValueKind, int64(math.Floor((duration.Hours())/24/7))), nil
	case Day:
		if t1.After(t2) {
			return NewSQLInt64(cfg.sqlValueKind, int64(math.Ceil(duration.Hours()/24))), nil
		}
		return NewSQLInt64(cfg.sqlValueKind, int64(math.Floor(duration.Hours()/24))), nil
	case Hour:
		return NewSQLInt64(cfg.sqlValueKind, int64(duration.Hours())), nil
	case Minute:
		return NewSQLInt64(cfg.sqlValueKind, int64(duration.Minutes())), nil
	case Second:
		return NewSQLInt64(cfg.sqlValueKind, int64(duration.Seconds())), nil
	case Microsecond:
		return NewSQLInt64(cfg.sqlValueKind, duration.Nanoseconds()/1000), nil
	default:
		err := fmt.Errorf("cannot add '%v' to timestamp", values[0])
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), err
	}
}

// nolint: unparam
func (f baseScalarFunctionExpr) timestampEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	t = t.In(schema.DefaultLocale)

	if len(values) == 1 {
		return NewSQLTimestamp(cfg.sqlValueKind, t), nil
	}

	d, ok := parseDuration(values[1])
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	t = t.Add(d).Round(time.Microsecond)

	return NewSQLTimestamp(cfg.sqlValueKind, t), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) toDaysEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	// This must be a SQLDate at this point. If it is not, the algebrizer has been broken.
	// Check the handling of scalar functions in the algebrizer.
	tmp, ok := values[0].(SQLDate)
	if !ok {
		return nil, fmt.Errorf("unable to evaluate input value %v in to_days", values[0])
	}
	date := Timestamp(tmp)

	// First compute the days from YearOne.
	target, err := daysFromYearZeroCalculation(date)
	if err != nil {
		return nil, err
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(target)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) toSecondsEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	// This must be a SQLTimestamp at this point. If it is not, the algebrizer has been broken.
	// Check the handling of scalar functions in the algebrizer.
	tmp, ok := values[0].(SQLTimestamp)
	if !ok {
		return nil, fmt.Errorf("unable to evaluate input value %v in to_seconds", values[0])
	}
	date := Timestamp(tmp)

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
	return NewSQLInt64(cfg.sqlValueKind, int64(target)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) trimEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	value := values[0].String()
	end := "both"
	toTrim := " "
	if len(values) == 3 {
		end = values[1].String()
		toTrim = values[2].String()
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

	return NewSQLVarchar(cfg.sqlValueKind, value), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) truncateEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	var truncated float64
	x := Float64(values[0])
	d := Float64(values[1])

	if d >= 0 {
		pow := math.Pow(10, d)
		i, _ := math.Modf(x * pow)
		truncated = i / pow
	} else {
		pow := math.Pow(10, math.Abs(d))
		i, _ := math.Modf(x / pow)
		truncated = i * pow
	}

	return NewSQLFloat(cfg.sqlValueKind, truncated), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) ucaseEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	value := strings.ToUpper(values[0].String())

	return NewSQLVarchar(cfg.sqlValueKind, value), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) unixTimestampEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	now := time.Now()

	if len(values) == 0 {
		return NewSQLUint64(cfg.sqlValueKind, uint64(now.Unix())), nil
	}

	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	t, _, ok := parseDateTime(values[0].String())
	if !ok || t.Before(epoch) {
		return NewSQLFloat(cfg.sqlValueKind, 0.0), nil
	}

	// Our times are parsed as if in UTC. However, we need to
	// parse it in the actual location the server's running
	// in - to account for any timezone difference.
	y, m, d := t.Date()
	ts := time.Date(y, m, d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), now.Location())
	return NewSQLUint64(cfg.sqlValueKind, uint64(ts.Unix())), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) userEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	str := fmt.Sprintf("%s@%s", cfg.user, cfg.remoteHost)
	return NewSQLVarchar(cfg.sqlValueKind, str), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) utcDateEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	now := time.Now().In(time.UTC)
	t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return NewSQLDate(cfg.sqlValueKind, t), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) utcTimestampEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return NewSQLTimestamp(cfg.sqlValueKind, time.Now().In(time.UTC)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) versionEvaluate(_ context.Context, cfg *ExecutionConfig, _ *ExecutionState, _ []SQLValue) (SQLValue, error) {
	return NewSQLVarchar(cfg.sqlValueKind, cfg.mySQLVersion), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) weekEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}
	check, ok := values[0].(SQLDate)
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	dateArg := Timestamp(check)
	// Mode should always be less than MAX_INT.
	mode := int(Int64(values[1]))

	ret := weekCalculation(dateArg, mode)
	if ret == -1 {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}
	return NewSQLInt64(cfg.sqlValueKind, int64(ret)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) weekdayEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	w := int(t.Weekday())
	if w == 0 {
		w = 7
	}
	return NewSQLInt64(cfg.sqlValueKind, int64(w-1)), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) yearEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.Year())), nil
}

// nolint: unparam
func (f baseScalarFunctionExpr) yearWeekEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}
	check, ok := values[0].(SQLDate)
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	date := Timestamp(check)
	year := date.Year()
	// Mode should always be less than MAX_INT.
	mode := int(Int64(values[1]))

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
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
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
	return NewSQLInt64(cfg.sqlValueKind, int64(year*100+week)), nil
}

// atan, atan2

// nolint: unparam
func (f baseScalarFunctionExpr) absEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(ctx, cfg, st, values, math.Abs)
}

// nolint: unparam
func (f baseScalarFunctionExpr) acosEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(ctx, cfg, st, values, math.Acos)
}

// nolint: unparam
func (f baseScalarFunctionExpr) asinEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(ctx, cfg, st, values, math.Asin)
}

// nolint: unparam
func (f baseScalarFunctionExpr) ceilEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(ctx, cfg, st, values, math.Ceil)
}

// nolint: unparam
func (f baseScalarFunctionExpr) cosEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(ctx, cfg, st, values, math.Cos)
}

// nolint: unparam
func (f baseScalarFunctionExpr) expEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(ctx, cfg, st, values, math.Exp)
}

// nolint: unparam
func (f baseScalarFunctionExpr) floorEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(ctx, cfg, st, values, math.Floor)
}

// nolint: unparam
func (f baseScalarFunctionExpr) radiansEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(ctx, cfg, st, values, toRadians)
}

// nolint: unparam
func (f baseScalarFunctionExpr) degreesEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(ctx, cfg, st, values, toDegrees)
}

// nolint: unparam
func (f baseScalarFunctionExpr) sinEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(ctx, cfg, st, values, math.Sin)
}

// nolint: unparam
func (f baseScalarFunctionExpr) sqrtEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(ctx, cfg, st, values, math.Sqrt)
}

// nolint: unparam
func (f baseScalarFunctionExpr) tanEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(ctx, cfg, st, values, math.Tan)
}

// nolint: unparam
func (f baseScalarFunctionExpr) atanSingleArgEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.singleArgFloatMathFuncEvaluate(ctx, cfg, st, values, math.Atan)
}

// nolint: unparam
func (f baseScalarFunctionExpr) atanDualArgEvaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.dualArgFloatMathFuncEvaluate(ctx, cfg, st, values, math.Atan2)
}

// nolint: unparam
func (f baseScalarFunctionExpr) atan2Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.dualArgFloatMathFuncEvaluate(ctx, cfg, st, values, math.Atan2)
}
