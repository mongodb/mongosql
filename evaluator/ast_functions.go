package evaluator

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"time"
)

//
// SQLAggFunctionExpr represents an aggregate function. These aggregate
// functions are avg, sum, count, max, and min.
//
type SQLAggFunctionExpr struct {
	Name     string
	Distinct bool
	Exprs    []SQLExpr
}

func (f *SQLAggFunctionExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	var distinctMap map[interface{}]bool = nil
	if f.Distinct {
		distinctMap = make(map[interface{}]bool)
	}

	switch f.Name {
	case "avg":
		return f.avgFunc(ctx, distinctMap)
	case "sum":
		return f.sumFunc(ctx, distinctMap)
	case "count":
		return f.countFunc(ctx, distinctMap)
	case "max":
		return f.maxFunc(ctx)
	case "min":
		return f.minFunc(ctx)
	default:
		return nil, fmt.Errorf("aggregate function '%v' is not supported", f.Name)
	}
}

func (f *SQLAggFunctionExpr) String() string {
	var distinct string
	if f.Distinct {
		distinct = "distinct "
	}
	return fmt.Sprintf("%s(%s%v)", f.Name, distinct, f.Exprs[0])
}

func (f *SQLAggFunctionExpr) avgFunc(ctx *EvalCtx, distinctMap map[interface{}]bool) (SQLValue, error) {
	var sum SQLNumeric = SQLInt(0)
	count := 0
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: Rows{row}}
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(evalCtx)
			if err != nil {
				return nil, err
			}
			if distinctMap != nil {
				if distinctMap[eval] {
					// already in our distinct map, so we skip this row
					continue
				} else {
					distinctMap[eval] = true
				}
			}
			count += 1

			n, err := convertToSQLNumeric(eval, ctx)
			if err == nil {
				sum = sum.Add(n)
			}
		}
	}

	return SQLFloat(sum.Float64() / float64(count)), nil
}

func (f *SQLAggFunctionExpr) countFunc(ctx *EvalCtx, distinctMap map[interface{}]bool) (SQLValue, error) {
	var count int64
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: Rows{row}}
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(evalCtx)
			if err != nil {
				return nil, err
			}
			if distinctMap != nil {
				if distinctMap[eval] {
					// already in our distinct map, so we skip this row
					continue
				} else {
					distinctMap[eval] = true
				}
			}

			if eval != nil && eval != SQLNull {
				count += 1
			}
		}
	}
	return SQLInt(count), nil
}

func (f *SQLAggFunctionExpr) maxFunc(ctx *EvalCtx) (SQLValue, error) {
	var max SQLValue
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: Rows{row}}
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(evalCtx)
			if err != nil {
				return nil, err
			}
			if max == nil {
				max = eval
				continue
			}
			compared, err := max.CompareTo(eval)
			if err != nil {
				return nil, err
			}
			if compared < 0 {
				max = eval
			}
		}
	}
	return max, nil
}

func (f *SQLAggFunctionExpr) minFunc(ctx *EvalCtx) (SQLValue, error) {
	var min SQLValue
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: Rows{row}}
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(evalCtx)
			if err != nil {
				return nil, err
			}
			if min == nil {
				min = eval
				continue
			}
			compared, err := min.CompareTo(eval)
			if err != nil {
				return nil, err
			}
			if compared > 0 {
				min = eval
			}
		}
	}
	return min, nil
}

func (f *SQLAggFunctionExpr) sumFunc(ctx *EvalCtx, distinctMap map[interface{}]bool) (SQLValue, error) {
	var sum SQLNumeric = SQLInt(0)
	for _, row := range ctx.Rows {
		evalCtx := &EvalCtx{Rows: Rows{row}}
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(evalCtx)
			if err != nil {
				return nil, err
			}

			if distinctMap != nil {
				if distinctMap[eval] {
					// already in our distinct map, so we skip this row
					continue
				} else {
					distinctMap[eval] = true
				}
			}

			n, err := convertToSQLNumeric(eval, ctx)
			if err == nil {
				sum = sum.Add(n)
			}
		}
	}

	return sum, nil
}

//
// SQLScalarFunctionExpr represents a scalar function.
//
type SQLScalarFunctionExpr struct {
	Name  string
	Exprs []SQLExpr
}

type fnDef struct {
	argCounts []int
	impl      func([]SQLValue, *EvalCtx) (SQLValue, error)
}

func newFnDef(impl func([]SQLValue, *EvalCtx) (SQLValue, error), argCounts ...int) fnDef {
	return fnDef{argCounts, impl}
}

var fnMap = map[string]fnDef{
	"abs":               newFnDef(absFunc, 1),
	"ascii":             newFnDef(asciiFunc, 1),
	"cast":              newFnDef(castFunc, 2),
	"concat":            newFnDef(concatFunc),
	"connection_id":     newFnDef(connectionIdFunc),
	"current_date":      newFnDef(currentDateFunc),
	"current_timestamp": newFnDef(currentTimestampFunc),
	"database":          newFnDef(dbFunc),
	"dayname":           newFnDef(dayNameFunc, 1),
	"dayofmonth":        newFnDef(dayOfMonthFunc, 1),
	"dayofweek":         newFnDef(dayOfWeekFunc, 1),
	"dayofyear":         newFnDef(dayOfYearFunc, 1),
	"exp":               newFnDef(expFunc, 1),
	"floor":             newFnDef(floorFunc, 1),
	"hour":              newFnDef(hourFunc, 1),
	"lcase":             newFnDef(lcaseFunc, 1),
	"length":            newFnDef(lengthFunc, 1),
	"locate":            newFnDef(locateFunc, 2, 3),
	"log10":             newFnDef(log10Func, 1),
	"ltrim":             newFnDef(ltrimFunc, 1),
	"minute":            newFnDef(minuteFunc, 1),
	"mod":               newFnDef(modFunc, 2),
	"month":             newFnDef(monthFunc, 1),
	"monthname":         newFnDef(monthNameFunc, 1),
	"pow":               newFnDef(powFunc, 2),
	"power":             newFnDef(powFunc, 2),
	"quarter":           newFnDef(quarterFunc, 1),
	"sqrt":              newFnDef(sqrtFunc, 1),
	"rtrim":             newFnDef(rtrimFunc, 1),
	"second":            newFnDef(secondFunc, 1),
	"substring":         newFnDef(substringFunc, 2, 3),
	"ucase":             newFnDef(ucaseFunc, 1),
	"week":              newFnDef(weekFunc, 1),
	"year":              newFnDef(yearFunc, 1),
}

func (f *SQLScalarFunctionExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	fd, ok := fnMap[f.Name]
	if ok {
		values, err := evaluateArgsWithCount(f, ctx, fd.argCounts...)
		if err != nil {
			return nil, err
		}

		return fd.impl(values, ctx)
	}

	// any functions not registered globally go here. There shouldn't be many here,
	// only those that need some special processing.
	switch f.Name {
	case "isnull":
		return f.isNullFunc(ctx)
	case "not":
		return f.notFunc(ctx)
	default:
		return nil, fmt.Errorf("scalar function '%v' is not supported", string(f.Name))
	}
}

func (f *SQLScalarFunctionExpr) String() string {
	return fmt.Sprintf("%s(%v)", f.Name, f.Exprs)
}

func (f *SQLScalarFunctionExpr) isNullFunc(ctx *EvalCtx) (SQLValue, error) {
	matcher := &SQLNullCmpExpr{f.Exprs[0]}
	result, err := Matches(matcher, ctx)
	if err != nil {
		return nil, err
	}
	return SQLBool(result), nil
}

func (f *SQLScalarFunctionExpr) notFunc(ctx *EvalCtx) (SQLValue, error) {
	matcher := &SQLNotExpr{f.Exprs[0]}
	result, err := Matches(matcher, ctx)
	if err != nil {
		return nil, err
	}
	return SQLBool(result), nil
}

func connectionIdFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLUint32(ctx.ExecCtx.ConnectionId()), nil
}

func dbFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLString(ctx.ExecCtx.DB()), nil
}

// http://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_abs
func absFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_ascii
func asciiFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

func castFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return NewSQLValue(values[0].Value(), values[1].String())
}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_concat
func concatFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	var bytes bytes.Buffer
	for _, value := range values {
		bytes.WriteString(value.String())
	}

	return SQLString(bytes.String()), nil
}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_curdate
func currentDateFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	value := time.Now().UTC()
	return SQLDate{value}, nil
}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_now
func currentTimestampFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	value := time.Now().UTC()
	return SQLTimestamp{value}, nil
}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayname
func dayNameFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLString(t.Weekday().String()), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_dayofmonth
func dayOfMonthFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Day())), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_dayofweek
func dayOfWeekFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Weekday()) + 1), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_dayofyear
func dayOfYearFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.YearDay())), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/mathematical-functions.html#function_exp
func expFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	n, ok := values[0].(SQLNumeric)
	if !ok {
		return SQLNull, nil
	}

	r := math.Exp(n.Float64())
	return SQLFloat(r), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/mathematical-functions.html#function_floor
func floorFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	n, ok := values[0].(SQLNumeric)
	if !ok {
		return SQLNull, nil
	}

	r := math.Floor(n.Float64())
	return SQLFloat(r), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_hour
func hourFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Hour())), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/string-functions.html#function_lcase
func lcaseFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	value := strings.ToLower(values[0].String())

	return SQLString(value), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/string-functions.html#function_length
func lengthFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	value := values[0].String()

	return SQLInt(len(value)), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/string-functions.html#function_locate
func locateFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	substr := values[0].String()
	str := values[1].String()
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
			str = string([]rune(str)[pos:])
			result = strings.Index(str, substr)
			if result > 0 {
				result += pos
			}
		}
	} else {
		result = strings.Index(str, substr)
	}

	return SQLInt(result + 1), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/mathematical-functions.html#function_log10
func log10Func(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

// https://dev.mysql.com/doc/refman/5.5/en/string-functions.html#function_ltrim
func ltrimFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	value := strings.TrimLeft(values[0].String(), " ")

	return SQLString(value), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_minute
func minuteFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Minute())), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/mathematical-functions.html#function_mod
func modFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_month
func monthFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Month())), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_monthname
func monthNameFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLString(t.Month().String()), nil
}

func powFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_quarter
func quarterFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

// https://dev.mysql.com/doc/refman/5.5/en/string-functions.html#function_rtrim
func rtrimFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	value := strings.TrimRight(values[0].String(), " ")

	return SQLString(value), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_second
func secondFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Second())), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/mathematical-functions.html#function_sqrt
func sqrtFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

// https://dev.mysql.com/doc/refman/5.5/en/string-functions.html#function_substring
func substringFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

// https://dev.mysql.com/doc/refman/5.5/en/string-functions.html#function_ucase
func ucaseFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if anyNull(values) {
		return SQLNull, nil
	}

	value := strings.ToUpper(values[0].String())

	return SQLString(value), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_week
func weekFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	// TODO: this one takes a mode as an optional second argument...
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	_, w := t.ISOWeek()

	return SQLInt(w), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_year
func yearFunc(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(t.Year()), nil
}

func anyNull(values []SQLValue) bool {
	for _, v := range values {
		if _, ok := v.(SQLNullValue); ok {
			return true
		}
	}

	return false
}

func ensureArgCount(f *SQLScalarFunctionExpr, counts ...int) error {
	if len(counts) == 0 {
		return nil
	}

	found := false
	actual := len(f.Exprs)
	for _, i := range counts {
		if actual == i {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("the '%v' function expects [%v] argument(s) but received %v", f.Name, counts, len(f.Exprs))
	}

	return nil
}

func evaluateArgs(f *SQLScalarFunctionExpr, ctx *EvalCtx) ([]SQLValue, error) {
	values := []SQLValue{}
	for _, arg := range f.Exprs {
		value, err := arg.Evaluate(ctx)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

func evaluateArgsWithCount(f *SQLScalarFunctionExpr, ctx *EvalCtx, counts ...int) ([]SQLValue, error) {
	err := ensureArgCount(f, counts...)
	if err != nil {
		return nil, err
	}

	return evaluateArgs(f, ctx)
}

const (
	dateTimeFormat = "2006-1-2 15:4:5"
	dateFormat     = "2006-1-2"
	timeFormat     = "15:4:5"
)

func parseDateTime(value string) (time.Time, bool) {
	t, err := time.Parse(dateTimeFormat, value)
	if err == nil {
		return t, true
	}

	t, err = time.Parse(dateFormat, value)
	if err == nil {
		return t, true
	}

	t, err = time.Parse(timeFormat, value)
	return t, err == nil
}
