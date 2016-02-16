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

func (f *SQLScalarFunctionExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {

	// TODO:: can we register aliases in the algebrizer so we don't have
	// to account for multiple names for the same function below?

	switch f.Name {
	// connector functions
	case "connection_id":
		return f.connectionIdFunc(ctx)
	case "database":
		return f.dbFunc(ctx)

		// scalar functions
	case "abs":
		return f.absFunc(ctx)
	case "ascii":
		return f.asciiFunc(ctx)
	case "cast":
		return f.castFunc(ctx)
	case "concat":
		return f.concatFunc(ctx)
	case "current_date":
		return f.currentDateFunc(ctx)
	case "current_timestamp":
		return f.currentTimestampFunc(ctx)
	case "dayname":
		return f.dayNameFunc(ctx)
	case "dayofmonth":
		return f.dayOfMonthFunc(ctx)
	case "dayofweek":
		return f.dayOfWeekFunc(ctx)
	case "dayofyear":
		return f.dayOfYearFunc(ctx)
	case "exp":
		return f.expFunc(ctx)
	case "floor":
		return f.floorFunc(ctx)
	case "hour":
		return f.hourFunc(ctx)
	case "isnull":
		return f.isNullFunc(ctx)
	case "lcase":
		return f.lcaseFunc(ctx)
	case "length":
		return f.lengthFunc(ctx)
	case "locate":
		return f.locateFunc(ctx)
	case "minute":
		return f.minuteFunc(ctx)
	case "month":
		return f.monthFunc(ctx)
	case "monthname":
		return f.monthNameFunc(ctx)
	case "not":
		return f.notFunc(ctx)
	case "quarter":
		return f.quarterFunc(ctx)
	case "second":
		return f.secondFunc(ctx)
	case "pow":
		return f.powFunc(ctx)
	case "ucase":
		return f.ucaseFunc(ctx)
	case "week":
		return f.weekFunc(ctx)
	case "year":
		return f.yearFunc(ctx)

	default:
		return nil, fmt.Errorf("scalar function '%v' is not supported", string(f.Name))
	}
}

func (f *SQLScalarFunctionExpr) String() string {
	return fmt.Sprintf("%s(%v)", f.Name, f.Exprs)
}

func (f *SQLScalarFunctionExpr) connectionIdFunc(ctx *EvalCtx) (SQLValue, error) {
	return SQLUint32(ctx.ExecCtx.ConnectionId()), nil
}

func (f *SQLScalarFunctionExpr) dbFunc(ctx *EvalCtx) (SQLValue, error) {
	return SQLString(ctx.ExecCtx.DB()), nil
}

// http://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_abs
func (f *SQLScalarFunctionExpr) absFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	if _, ok := values[0].(SQLNullValue); ok {
		return values[0], nil
	}

	numeric, ok := values[0].(SQLNumeric)
	if !ok {
		return SQLFloat(0), nil
	}

	result := math.Abs(numeric.Float64())
	return SQLFloat(result), nil
}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_ascii
func (f *SQLScalarFunctionExpr) asciiFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	if _, ok := values[0].(SQLNullValue); ok {
		return values[0], nil
	}

	str := values[0].String()
	if str == "" {
		return SQLInt(0), nil
	}

	c := str[0]

	return SQLInt(c), nil
}

func (f *SQLScalarFunctionExpr) castFunc(ctx *EvalCtx) (SQLValue, error) {
	ensureArgCount(f, 2)

	value, err := f.Exprs[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	return NewSQLValue(value.Value(), f.Exprs[1].String())
}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_concat
func (f *SQLScalarFunctionExpr) concatFunc(ctx *EvalCtx) (SQLValue, error) {
	var bytes bytes.Buffer
	for _, arg := range f.Exprs {

		value, err := arg.Evaluate(ctx)
		if err != nil {
			return nil, err
		}

		if _, ok := value.(SQLNullValue); ok {
			return SQLNull, nil
		}

		bytes.WriteString(value.String())
	}

	return SQLString(bytes.String()), nil
}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_curdate
func (f *SQLScalarFunctionExpr) currentDateFunc(ctx *EvalCtx) (SQLValue, error) {
	err := ensureArgCount(f, 0)
	if err != nil {
		return nil, err
	}

	value := time.Now().UTC()

	return SQLDate{value}, nil
}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_now
func (f *SQLScalarFunctionExpr) currentTimestampFunc(ctx *EvalCtx) (SQLValue, error) {
	err := ensureArgCount(f, 0)
	if err != nil {
		return nil, err
	}

	value := time.Now().UTC()

	return SQLTimestamp{value}, nil
}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayname
func (f *SQLScalarFunctionExpr) dayNameFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLString(t.Weekday().String()), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_dayofmonth
func (f *SQLScalarFunctionExpr) dayOfMonthFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Day())), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_dayofweek
func (f *SQLScalarFunctionExpr) dayOfWeekFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Weekday()) + 1), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_dayofyear
func (f *SQLScalarFunctionExpr) dayOfYearFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.YearDay())), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/mathematical-functions.html#function_exp
func (f *SQLScalarFunctionExpr) expFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	n, ok := values[0].(SQLNumeric)
	if !ok {
		return SQLNull, nil
	}

	r := math.Exp(n.Float64())
	return SQLFloat(r), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/mathematical-functions.html#function_floor
func (f *SQLScalarFunctionExpr) floorFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	n, ok := values[0].(SQLNumeric)
	if !ok {
		return SQLNull, nil
	}

	r := math.Floor(n.Float64())
	return SQLFloat(r), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_hour
func (f *SQLScalarFunctionExpr) hourFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Hour())), nil
}

func (f *SQLScalarFunctionExpr) isNullFunc(ctx *EvalCtx) (SQLValue, error) {
	matcher := &SQLNullCmpExpr{f.Exprs[0]}
	result, err := Matches(matcher, ctx)
	if err != nil {
		return nil, err
	}
	return SQLBool(result), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/string-functions.html#function_lcase
func (f *SQLScalarFunctionExpr) lcaseFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	if _, ok := values[0].(SQLNullValue); ok {
		return values[0], nil
	}

	value := strings.ToLower(values[0].String())

	return SQLString(value), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/string-functions.html#function_length
func (f *SQLScalarFunctionExpr) lengthFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	if _, ok := values[0].(SQLNullValue); ok {
		return values[0], nil
	}

	value := values[0].String()

	return SQLInt(len(value)), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/string-functions.html#function_locate
func (f *SQLScalarFunctionExpr) locateFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 2, 3)
	if err != nil {
		return nil, err
	}

	if _, ok := values[0].(SQLNullValue); ok {
		return values[0], nil
	}
	if _, ok := values[1].(SQLNullValue); ok {
		return values[1], nil
	}

	substr := values[0].String()
	str := values[1].String()
	result := 0
	if len(values) == 3 {
		posValue, ok := values[2].(SQLNumeric)
		if !ok {
			return SQLNull, nil
		}

		pos := int(math.Floor(posValue.Float64())) - 1

		if len(str) <= pos {
			result = 0
		} else {
			str = str[pos:]
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

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_minute
func (f *SQLScalarFunctionExpr) minuteFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Minute())), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_month
func (f *SQLScalarFunctionExpr) monthFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Month())), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_monthname
func (f *SQLScalarFunctionExpr) monthNameFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLString(t.Month().String()), nil
}

func (f *SQLScalarFunctionExpr) notFunc(ctx *EvalCtx) (SQLValue, error) {
	matcher := &SQLNotExpr{f.Exprs[0]}
	result, err := Matches(matcher, ctx)
	if err != nil {
		return nil, err
	}
	return SQLBool(result), nil
}

func (f *SQLScalarFunctionExpr) powFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 2)
	if err != nil {
		return nil, err
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
func (f *SQLScalarFunctionExpr) quarterFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

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

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_second
func (f *SQLScalarFunctionExpr) secondFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Second())), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/string-functions.html#function_ucase
func (f *SQLScalarFunctionExpr) ucaseFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	if _, ok := values[0].(SQLNullValue); ok {
		return values[0], nil
	}

	value := strings.ToUpper(values[0].String())

	return SQLString(value), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_week
func (f *SQLScalarFunctionExpr) weekFunc(ctx *EvalCtx) (SQLValue, error) {

	// TODO: this one takes a mode as an optional second argument...

	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	_, w := t.ISOWeek()

	return SQLInt(w), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_year
func (f *SQLScalarFunctionExpr) yearFunc(ctx *EvalCtx) (SQLValue, error) {
	values, err := evaluateArgsWithCount(f, ctx, 1)
	if err != nil {
		return nil, err
	}

	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(t.Year()), nil
}

func ensureArgCount(f *SQLScalarFunctionExpr, counts ...int) error {
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
