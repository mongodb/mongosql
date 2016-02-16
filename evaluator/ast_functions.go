package evaluator

import (
	"bytes"
	"fmt"
	"math"
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
	case "isnull":
		return f.isNullFunc(ctx)
	case "not":
		return f.notFunc(ctx)
	case "pow":
		return f.powFunc(ctx)

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
	err := ensureArgCount(f, 1)
	if err != nil {
		return nil, err
	}

	value, err := f.Exprs[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if _, ok := value.(SQLNullValue); ok {
		return value, nil
	}

	numeric, ok := value.(SQLNumeric)
	if !ok {
		return SQLFloat(0), nil
	}

	result := math.Abs(numeric.Float64())
	return SQLFloat(result), nil
}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_ascii
func (f *SQLScalarFunctionExpr) asciiFunc(ctx *EvalCtx) (SQLValue, error) {
	err := ensureArgCount(f, 1)
	if err != nil {
		return nil, err
	}

	value, err := f.Exprs[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if _, ok := value.(SQLNullValue); ok {
		return value, nil
	}

	str := value.String()
	if str == "" {
		return SQLInt(0), nil
	}

	c := str[0]

	return SQLInt(c), nil
}

func (f *SQLScalarFunctionExpr) castFunc(ctx *EvalCtx) (SQLValue, error) {
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
	value := time.Now().UTC()

	return SQLDate{value}, nil
}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_now
func (f *SQLScalarFunctionExpr) currentTimestampFunc(ctx *EvalCtx) (SQLValue, error) {
	value := time.Now().UTC()

	return SQLTimestamp{value}, nil
}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayname
func (f *SQLScalarFunctionExpr) dayNameFunc(ctx *EvalCtx) (SQLValue, error) {
	err := ensureArgCount(f, 1)
	if err != nil {
		return nil, err
	}

	value, err := f.Exprs[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	t, err := time.Parse("2006-1-1", value.String())
	if err != nil {
		return SQLNull, nil
	}

	return SQLString(t.Weekday().String()), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_dayofmonth
func (f *SQLScalarFunctionExpr) dayOfMonthFunc(ctx *EvalCtx) (SQLValue, error) {
	err := ensureArgCount(f, 1)
	if err != nil {
		return nil, err
	}

	value, err := f.Exprs[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	t, err := time.Parse("2006-1-1", value.String())
	if err != nil {
		return SQLNull, nil
	}

	return SQLInt(int(t.Day())), nil
}

func (f *SQLScalarFunctionExpr) dayOfWeekFunc(ctx *EvalCtx) (SQLValue, error) {
	err := ensureArgCount(f, 1)
	if err != nil {
		return nil, err
	}

	value, err := f.Exprs[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	t, err := time.Parse("2006-1-1", value.String())
	if err != nil {
		return SQLNull, nil
	}

	return SQLInt(int(t.Weekday()) + 1), nil
}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_dayofyear
func (f *SQLScalarFunctionExpr) dayOfYearFunc(ctx *EvalCtx) (SQLValue, error) {
	err := ensureArgCount(f, 1)
	if err != nil {
		return nil, err
	}

	value, err := f.Exprs[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	t, err := time.Parse("2006-1-1", value.String())
	if err != nil {
		return SQLNull, nil
	}

	return SQLInt(int(t.YearDay())), nil
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

func (f *SQLScalarFunctionExpr) powFunc(ctx *EvalCtx) (SQLValue, error) {
	base, err := f.Exprs[0].Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	exponent, err := f.Exprs[1].Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if bNum, ok := base.(SQLNumeric); ok {
		if eNum, ok := exponent.(SQLNumeric); ok {
			return SQLFloat(math.Pow(bNum.Float64(), eNum.Float64())), nil
		}
		return nil, fmt.Errorf("exponent must be a number, but got %t", exponent)
	}
	return nil, fmt.Errorf("base must be a number, but got %T", base)
}

func ensureArgCount(f *SQLScalarFunctionExpr, count int) error {
	if len(f.Exprs) != count {
		return fmt.Errorf("the '%v' function expects %v argument but received %v", f.Name, count, len(f.Exprs))
	}

	return nil
}
