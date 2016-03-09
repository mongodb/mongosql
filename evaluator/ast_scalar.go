package evaluator

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/schema"
)

type connectionIdFunc struct{}

func (_ *connectionIdFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLUint32(ctx.ExecCtx.ConnectionId()), nil
}

func (_ *connectionIdFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *connectionIdFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type dbFunc struct{}

func (_ *dbFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLString(ctx.ExecCtx.DB()), nil
}

func (_ *dbFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *dbFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type absFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_abs
func (_ *absFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	numeric, ok := values[0].(SQLNumeric)
	if !ok {
		return SQLFloat(0), nil
	}

	result := math.Abs(numeric.Float64())
	return SQLFloat(result), nil
}

func (_ *absFunc) Type() schema.SQLType {
	return schema.SQLFloat
}

func (_ *absFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type asciiFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_ascii
func (_ *asciiFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	str := values[0].String()
	if str == "" {
		return SQLInt(0), nil
	}

	c := str[0]

	return SQLInt(c), nil
}

func (_ *asciiFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *asciiFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type castFunc struct{}

func (_ *castFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return NewSQLValue(values[0].Value(), schema.SQLType(values[1].String()), schema.MongoNone)
}

func (_ *castFunc) Type() schema.SQLType {
	return schema.SQLNone
}

func (_ *castFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type coalesceFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_coalesce
func (_ *coalesceFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	for _, value := range values {
		if value != SQLNull {
			return value, nil
		}
	}
	return SQLNull, nil
}

func (_ *coalesceFunc) Type() schema.SQLType {
	return schema.SQLNone
}

func (_ *coalesceFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, -1)
}

type concatFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_concat
func (_ *concatFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (v SQLValue, err error) {
	if anyNull(values) {
		v = SQLNull
		err = nil
		return
	}

	defer func() {
		if r := recover(); r != nil {
			v = nil
			err = fmt.Errorf("%v", r)
		}
	}()

	var bytes bytes.Buffer
	for _, value := range values {
		bytes.WriteString(value.String())
	}

	v = SQLString(bytes.String())
	err = nil
	return
}

func (_ *concatFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *concatFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, -1)
}

type currentDateFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_curdate
func (_ *currentDateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	value := time.Now().In(schema.DefaultLocale)
	return SQLDate{value}, nil
}

func (_ *currentDateFunc) Type() schema.SQLType {
	return schema.SQLDate
}

func (_ *currentDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type currentTimestampFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_now
func (_ *currentTimestampFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	value := time.Now().UTC()
	return SQLTimestamp{value}, nil
}

func (_ *currentTimestampFunc) Type() schema.SQLType {
	return schema.SQLTimestamp
}

func (_ *currentTimestampFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type dayNameFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayname
func (_ *dayNameFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLString(t.Weekday().String()), nil
}

func (_ *dayNameFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *dayNameFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dayOfMonthFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayofmonth
func (_ *dayOfMonthFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Day())), nil
}

func (_ *dayOfMonthFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *dayOfMonthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dayOfWeekFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayofweek
func (_ *dayOfWeekFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Weekday()) + 1), nil
}

func (_ *dayOfWeekFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *dayOfWeekFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dayOfYearFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayofyear
func (_ *dayOfYearFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.YearDay())), nil
}

func (_ *dayOfYearFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *dayOfYearFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type expFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_exp
func (_ *expFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	n, ok := values[0].(SQLNumeric)
	if !ok {
		return SQLNull, nil
	}

	r := math.Exp(n.Float64())
	return SQLFloat(r), nil
}

func (_ *expFunc) Type() schema.SQLType {
	return schema.SQLFloat
}

func (_ *expFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type floorFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_floor
func (_ *floorFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	n, ok := values[0].(SQLNumeric)
	if !ok {
		return SQLNull, nil
	}

	r := math.Floor(n.Float64())
	return SQLFloat(r), nil
}

func (_ *floorFunc) Type() schema.SQLType {
	return schema.SQLFloat
}

func (_ *floorFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type hourFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_hour
func (_ *hourFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Hour())), nil
}

func (_ *hourFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *hourFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type isnullFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_isnull
func (_ *isnullFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	matcher := &SQLNullCmpExpr{values[0]}
	result, err := Matches(matcher, ctx)
	if err != nil {
		return nil, err
	}
	if SQLBool(result) == SQLTrue {
		return SQLInt(1), nil
	}
	return SQLInt(0), nil
}

func (_ *isnullFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *isnullFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type lcaseFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_lcase
func (_ *lcaseFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	value := strings.ToLower(values[0].String())

	return SQLString(value), nil
}

func (_ *lcaseFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *lcaseFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type lengthFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_length
func (_ *lengthFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	value := values[0].String()

	return SQLInt(len(value)), nil
}

func (_ *lengthFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *lengthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type locateFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_locate
func (_ *locateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	substr := []rune(values[0].String())
	str := []rune(values[1].String())
	result := 0
	if len(values) == 3 {
		posValue, ok := values[2].(SQLNumeric)
		if !ok {
			return SQLNull, nil
		}

		pos := int(posValue.Float64()) - 1 // MySQL uses 1 as a basis

		if len(str) <= pos {
			result = 0
		} else {
			str = str[pos:]
			result = runesIndex(str, substr)
			if result > 0 {
				result += pos
			}
		}
	} else {
		result = runesIndex(str, substr)
	}

	return SQLInt(result + 1), nil
}

func (_ *locateFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *locateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2, 3)
}

type log10Func struct{}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_log10
func (_ *log10Func) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	nValue, ok := values[0].(SQLNumeric)
	if !ok {
		return SQLNull, nil
	}

	n := nValue.Float64()

	if n <= 0 {
		return SQLFloat(0), nil
	}

	r := math.Log10(n)
	return SQLFloat(r), nil
}

func (_ *log10Func) Type() schema.SQLType {
	return schema.SQLFloat
}

func (_ *log10Func) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type ltrimFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_ltrim
func (_ *ltrimFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	value := strings.TrimLeft(values[0].String(), " ")

	return SQLString(value), nil
}

func (_ *ltrimFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *ltrimFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type minuteFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_minute
func (_ *minuteFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Minute())), nil
}

func (_ *minuteFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *minuteFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type modFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_mod
func (_ *modFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	nValue, ok := values[0].(SQLNumeric)
	if !ok {
		return SQLNull, nil
	}

	mValue, ok := values[1].(SQLNumeric)
	if !ok {
		return SQLNull, nil
	}
	n := nValue.Float64()
	m := mValue.Float64()

	if m == 0 {
		return SQLNull, nil
	}

	r := math.Mod(n, m)
	return SQLFloat(r), nil
}

func (_ *modFunc) Type() schema.SQLType {
	return schema.SQLFloat
}

func (_ *modFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type monthFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_month
func (_ *monthFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Month())), nil
}

func (_ *monthFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *monthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type monthNameFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_monthname
func (_ *monthNameFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLString(t.Month().String()), nil
}

func (_ *monthNameFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *monthNameFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type notFunc struct{}

func (_ *notFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	matcher := &SQLNotExpr{values[0]}
	result, err := Matches(matcher, ctx)
	if err != nil {
		return nil, err
	}
	if SQLBool(result) == SQLTrue {
		return SQLInt(1), nil
	}
	return SQLInt(0), nil
}

func (_ *notFunc) Type() schema.SQLType {
	return schema.SQLBoolean
}

func (_ *notFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type powFunc struct{}

func (_ *powFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	if bNum, ok := values[0].(SQLNumeric); ok {
		if eNum, ok := values[1].(SQLNumeric); ok {
			return SQLFloat(math.Pow(bNum.Float64(), eNum.Float64())), nil
		}
		return nil, fmt.Errorf("exponent must be a number, but got %t", values[1])
	}
	return nil, fmt.Errorf("base must be a number, but got %T", values[0])
}

func (_ *powFunc) Type() schema.SQLType {
	return schema.SQLFloat
}

func (_ *powFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type quarterFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_quarter
func (_ *quarterFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
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

	return SQLInt(q), nil
}

func (_ *quarterFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *quarterFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type rtrimFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_rtrim
func (_ *rtrimFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	value := strings.TrimRight(values[0].String(), " ")

	return SQLString(value), nil
}

func (_ *rtrimFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *rtrimFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type secondFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_second
func (_ *secondFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Second())), nil
}

func (_ *secondFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *secondFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type sqrtFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_sqrt
func (_ *sqrtFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	nValue, ok := values[0].(SQLNumeric)
	if !ok {
		return SQLNull, nil
	}

	n := nValue.Float64()
	if n < 0 {
		return SQLNull, nil
	}

	r := math.Sqrt(n)
	return SQLFloat(r), nil
}

func (_ *sqrtFunc) Type() schema.SQLType {
	return schema.SQLFloat
}

func (_ *sqrtFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type substringFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_substring
func (_ *substringFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	str := []rune(values[0].String())
	posValue, ok := values[1].(SQLNumeric)
	if !ok {
		return SQLNull, nil
	}

	pos := int(posValue.Float64())

	if pos > len(str) {
		return SQLString(""), nil
	} else if pos < 0 {
		pos = len(str) + pos

		if pos < 0 {
			pos = 0
		}
	} else {
		pos-- // MySQL uses 1 as a basis
	}

	if len(values) == 3 {
		lenValue, ok := values[2].(SQLNumeric)
		if !ok {
			return SQLNull, nil
		}

		length := int(lenValue.Float64())
		if length < 1 {
			return SQLString(""), nil
		}

		str = str[pos : pos+length]
	} else {
		str = str[pos:]
	}

	return SQLString(string(str)), nil
}

func (_ *substringFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *substringFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2, 3)
}

type ucaseFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_ucase
func (_ *ucaseFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	value := strings.ToUpper(values[0].String())

	return SQLString(value), nil
}

func (_ *ucaseFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *ucaseFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type weekFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_week
func (_ *weekFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	// TODO: this one takes a mode as an optional second argument...
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	_, w := t.ISOWeek()

	return SQLInt(w), nil
}

func (_ *weekFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *weekFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type yearFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_year
func (_ *yearFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(t.Year()), nil
}

func (_ *yearFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *yearFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

// Helper functions
func anyNull(values []SQLValue) bool {
	for _, v := range values {
		if _, ok := v.(SQLNullValue); ok {
			return true
		}
	}

	return false
}

func evaluateArgs(exprs []SQLExpr, ctx *EvalCtx) ([]SQLValue, error) {

	values := []SQLValue{}

	for _, expr := range exprs {
		value, err := expr.Evaluate(ctx)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

func ensureArgCount(exprCount int, counts ...int) error {
	// for scalar functions that accept a variable number of arguments
	if len(counts) == 1 && counts[0] == -1 {
		if exprCount == 0 {
			return ErrIncorrectVarCount
		}
		return nil
	}

	found := false
	actual := exprCount
	for _, i := range counts {
		if actual == i {
			found = true
			break
		}
	}

	if !found {
		return ErrIncorrectCount
	}

	return nil
}

func parseDateTime(value string) (time.Time, bool) {
	for _, f := range schema.TimestampCtorFormats {
		t, err := time.Parse(f, value)
		if err == nil {
			return t, true
		}
	}
	return time.Time{}, false
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
