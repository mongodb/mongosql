package evaluator

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
)

var (
	ErrIncorrectVarCount = errors.New("incorrect variable parameter count in the call to native function")
	ErrIncorrectCount    = errors.New("incorrect parameter count in function")
)

const (
	avgAggregateName          = "avg"
	countAggregateName        = "count"
	maxAggregateName          = "max"
	minAggregateName          = "min"
	stdAggregateName          = "std"
	stddevAggregateName       = "stddev"
	stddevPopAggregateName    = "stddev_pop"
	stddevSampleAggregateName = "stddev_samp"
	sumAggregateName          = "sum"
)

//
// SQLAggFunctionExpr represents an aggregate function. These aggregate
// functions are avg, sum, count, max, min, std, stddev, stddev_pop, and stddev_samp.
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
	case avgAggregateName:
		return f.avgFunc(ctx, distinctMap)
	case sumAggregateName:
		return f.sumFunc(ctx, distinctMap)
	case countAggregateName:
		return f.countFunc(ctx, distinctMap)
	case maxAggregateName:
		return f.maxFunc(ctx)
	case minAggregateName:
		return f.minFunc(ctx)
	case stdAggregateName, stddevAggregateName, stddevPopAggregateName:
		return f.stdFunc(ctx, distinctMap, false)
	case stddevSampleAggregateName:
		return f.stdFunc(ctx, distinctMap, true)
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

func (f *SQLAggFunctionExpr) Type() schema.SQLType {
	switch f.Name {
	case avgAggregateName, sumAggregateName, stdAggregateName, stddevAggregateName, stddevPopAggregateName, stddevSampleAggregateName:
		switch f.Exprs[0].Type() {
		case schema.SQLInt, schema.SQLInt64:
			// TODO: this should return a decimal when we have decimal support
			return schema.SQLFloat
		default:
			return schema.SQLFloat
		}
	case countAggregateName:
		return schema.SQLInt
	}

	return f.Exprs[0].Type()
}

func (f *SQLAggFunctionExpr) avgFunc(ctx *EvalCtx, distinctMap map[interface{}]bool) (SQLValue, error) {
	count := 0.0

	sum := decimal.Zero

	isDecimal := false

	floatSum, correction := 0.0, 0.0

	for _, row := range ctx.Rows {
		evalCtx := NewEvalCtx(ctx.ExecutionCtx, row)
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(evalCtx)
			if err != nil {
				return nil, err
			}

			if eval == SQLNull {
				continue
			}

			if distinctMap != nil {
				if distinctMap[eval] {
					// already in our distinct map, so we skip this row
					continue
				} else {
					distinctMap[eval] = true
				}
			}

			count++

			if isDecimal || eval.Type() == schema.SQLDecimal128 {
				isDecimal = true
				sum = sum.Add(eval.Decimal128())
				continue
			}

			floatEval := eval.Float64()

			// handle AVG(X) overflowing float64 range
			if runningSum := floatSum + correction; runningSum > math.MaxFloat64-floatEval {
				isDecimal = true
			}

			// this avoids catastrophic cancellation in
			// summing a series of floats or mixed types.
			floatEval, correction = fast2Sum(floatEval, correction)
			floatSum, floatEval = twoSum(floatSum, floatEval)
			correction += floatEval
		}
	}

	if count == 0 {
		return SQLNull, nil
	}

	floatSum += correction

	if isDecimal {
		sum = sum.Add(decimal.NewFromFloat(floatSum))
		avg := sum.Div(decimal.NewFromFloat(count))
		return SQLDecimal128(avg), nil
	}

	return SQLFloat(floatSum / count), nil
}

func (f *SQLAggFunctionExpr) countFunc(ctx *EvalCtx, distinctMap map[interface{}]bool) (SQLValue, error) {
	count, fCount, dCount := uint64(0), float64(math.MaxUint64), decimal.NewFromFloat(math.MaxFloat64)

	inDecimalRange, decimalOne := false, decimal.NewFromFloat(1.0)

	for _, row := range ctx.Rows {
		evalCtx := NewEvalCtx(ctx.ExecutionCtx, row)
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(evalCtx)
			if err != nil {
				return nil, err
			}

			if distinctMap != nil {
				if distinctMap[eval] {
					continue
				} else {
					distinctMap[eval] = true
				}
			}

			if eval != nil && eval != SQLNull {
				inDecimalRange = fCount == math.MaxFloat64
				if inDecimalRange {
					dCount.Add(decimalOne)
				} else if count >= math.MaxUint64 {
					fCount++
				} else {
					count++
				}
			}
		}
	}

	if inDecimalRange {
		return SQLDecimal128(dCount), nil
	} else if count > math.MaxInt64 {
		return SQLFloat(fCount), nil
	}

	return SQLInt(count), nil
}

func (f *SQLAggFunctionExpr) maxFunc(ctx *EvalCtx) (SQLValue, error) {
	var max SQLValue = SQLNull
	for _, row := range ctx.Rows {
		evalCtx := NewEvalCtx(ctx.ExecutionCtx, row)
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(evalCtx)
			if err != nil {
				return nil, err
			}
			if eval != SQLNull {
				if max == SQLNull {
					max = eval
					continue
				}
			} else {
				continue
			}

			compared, err := CompareTo(max, eval)
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
	var min SQLValue = SQLNull
	for _, row := range ctx.Rows {
		evalCtx := NewEvalCtx(ctx.ExecutionCtx, row)
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(evalCtx)
			if err != nil {
				return nil, err
			}

			if eval != SQLNull {
				if min == SQLNull {
					min = eval
					continue
				}
			} else {
				continue
			}

			compared, err := CompareTo(min, eval)
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

	floatSum, correction := 0.0, 0.0

	isDecimal := false

	sum := decimal.Zero

	allNull := true

	for _, row := range ctx.Rows {
		evalCtx := NewEvalCtx(ctx.ExecutionCtx, row)
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(evalCtx)
			if err != nil {
				return nil, err
			}

			if eval == SQLNull {
				continue
			} else {
				allNull = false
			}

			if distinctMap != nil {
				if distinctMap[eval] {
					continue
				} else {
					distinctMap[eval] = true
				}
			}

			evalType := eval.Type()
			if isDecimal || evalType == schema.SQLDecimal128 {
				isDecimal = true
				sum = sum.Add(eval.Decimal128())
				continue
			}

			floatEval := eval.Float64()

			// handle SUM(X) overflowing float64 range
			if runningSum := floatSum + correction; runningSum > math.MaxFloat64-floatEval {
				isDecimal = true
			}

			// this avoids catastrophic cancellation in
			// summing a series of floats or mixed types.
			floatEval, correction = fast2Sum(floatEval, correction)
			floatSum, floatEval = twoSum(floatSum, floatEval)
			correction += floatEval
		}
	}

	if allNull {
		return SQLNull, nil
	}

	floatSum += correction

	if isDecimal {
		sum = sum.Add(decimal.NewFromFloat(floatSum))
		return SQLDecimal128(sum), nil
	}

	return SQLFloat(floatSum), nil
}

func (f *SQLAggFunctionExpr) stdFunc(ctx *EvalCtx, distinctMap map[interface{}]bool, isSamp bool) (SQLValue, error) {
	var data []SQLValue

	sum := decimal.Zero

	floatSum, correction, count := 0.0, 0.0, 0.0

	isDecimal := false

	for _, row := range ctx.Rows {
		evalCtx := NewEvalCtx(ctx.ExecutionCtx, row)
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(evalCtx)
			if err != nil {
				return nil, err
			}

			if eval == SQLNull {
				continue
			}

			if distinctMap != nil {
				if distinctMap[eval] {
					// already in our distinct map, so we skip this row
					continue
				} else {
					distinctMap[eval] = true
				}
			}

			count++

			data = append(data, eval)

			if isDecimal || eval.Type() == schema.SQLDecimal128 {
				isDecimal = true
				sum = sum.Add(eval.Decimal128())
				continue
			}

			floatEval := eval.Float64()

			// handle STDDEV(X) overflowing float64 range
			if runningSum := floatSum + correction; runningSum > math.MaxFloat64-floatEval {
				isDecimal = true
			}

			// this avoids catastrophic cancellation in
			// summing a series of floats or mixed types.
			floatEval, correction = fast2Sum(floatEval, correction)
			floatSum, floatEval = twoSum(floatSum, floatEval)
			correction += floatEval
		}
	}

	if count == 0 {
		return SQLNull, nil
	}

	floatSum += correction

	if isDecimal {
		diff := decimal.Zero
		sum = sum.Add(decimal.NewFromFloat(floatSum))
		avg := sum.Div(decimal.NewFromFloat(count))

		for _, v := range data {
			val := v.Decimal128().Sub(avg)
			diff = diff.Add(val.Mul(val))
		}

		// Sample standard deviation
		if isSamp {
			if count == 1 {
				return SQLNull, nil
			}
			diff = diff.Div(decimal.NewFromFloat(count - 1))
			f, _ := diff.Float64()
			return SQLDecimal128(decimal.NewFromFloat(math.Sqrt(f))), nil
		}

		// Population standard deviation
		diff = diff.Div(decimal.NewFromFloat(count))
		f, _ := diff.Float64()
		return SQLDecimal128(decimal.NewFromFloat(math.Sqrt(f))), nil
	}

	avg := floatSum / count
	diff := 0.0

	for _, val := range data {
		diff += math.Pow(val.Float64()-avg, 2)
	}

	// Sample standard deviation
	if isSamp {
		if count == 1 {
			return SQLNull, nil
		}
		return SQLFloat(math.Sqrt(diff / (count - 1))), nil
	}

	// Population standard deviation
	return SQLFloat(math.Sqrt(diff / count)), nil
}

//
// SQLScalarFunctionExpr represents a scalar function.
//
type SQLScalarFunctionExpr struct {
	Name  string
	Exprs []SQLExpr
}

type scalarFunc interface {
	Evaluate([]SQLValue, *EvalCtx) (SQLValue, error)
	Validate(exprCount int) error
	Type() schema.SQLType
}

type normalizingScalarFunc interface {
	normalize(*SQLScalarFunctionExpr) SQLExpr
}

func NewIfScalarFunctionExpr(condition, truePart, falsePart SQLExpr) *SQLScalarFunctionExpr {
	return &SQLScalarFunctionExpr{
		Name:  "if",
		Exprs: []SQLExpr{condition, truePart, falsePart},
	}
}

var scalarFuncMap = map[string]scalarFunc{
	"abs":               &absFunc{},
	"ascii":             &asciiFunc{},
	"cast":              &castFunc{},
	"char_length":       &lengthFunc{},
	"coalesce":          &coalesceFunc{},
	"concat":            &concatFunc{},
	"concat_ws":         &concatWsFunc{},
	"connection_id":     &connectionIdFunc{},
	"convert":           &convertFunc{},
	"current_date":      &currentDateFunc{},
	"current_timestamp": &currentTimestampFunc{},
	"current_user":      &userFunc{},
	"database":          &dbFunc{},
	"date":              &dateFunc{},
	"date_add":          &dateAddFunc{},
	"date_sub":          &dateSubFunc{},
	"day":               &dayOfMonthFunc{},
	"dayname":           &dayNameFunc{},
	"dayofmonth":        &dayOfMonthFunc{},
	"dayofweek":         &dayOfWeekFunc{},
	"dayofyear":         &dayOfYearFunc{},
	"exp":               &expFunc{},
	"extract":           &extractFunc{},
	"floor":             &floorFunc{},
	"greatest":          &greatestFunc{},
	"hour":              &hourFunc{},
	"if":                &ifFunc{},
	"ifnull":            &ifnullFunc{},
	"instr":             &instrFunc{},
	"isnull":            &isnullFunc{},
	"lcase":             &lcaseFunc{},
	"least":             &leastFunc{},
	"left":              &leftFunc{},
	"length":            &lengthFunc{},
	"ln":                &naturalLogFunc{},
	"locate":            &locateFunc{},
	"log":               &naturalLogFunc{},
	"log2":              &log2Func{},
	"log10":             &log10Func{},
	"lower":             &lcaseFunc{},
	"ltrim":             &ltrimFunc{},
	"makedate":          &makeDateFunc{},
	"minute":            &minuteFunc{},
	"mod":               &modFunc{},
	"month":             &monthFunc{},
	"monthname":         &monthNameFunc{},
	"not":               &notFunc{},
	"now":               &currentTimestampFunc{},
	"nullif":            &nullifFunc{},
	"pow":               &powFunc{},
	"power":             &powFunc{},
	"quarter":           &quarterFunc{},
	"right":             &rightFunc{},
	"round":             &roundFunc{},
	"rtrim":             &rtrimFunc{},
	"schema":            &dbFunc{},
	"second":            &secondFunc{},
	"session_user":      &userFunc{},
	"sqrt":              &sqrtFunc{},
	"str_to_date":       &strToDateFunc{},
	"substr":            &substringFunc{},
	"substring":         &substringFunc{},
	"system_user":       &userFunc{},
	"timestampadd":      &timestampAddFunc{},
	"timestampdiff":     &timestampDiffFunc{},
	"ucase":             &ucaseFunc{},
	"upper":             &ucaseFunc{},
	"user":              &userFunc{},
	"version":           &versionFunc{},
	"week":              &weekFunc{},
	"weekday":           &weekdayFunc{},
	"weekofyear":        &weekOfYearFunc{},
	"year":              &yearFunc{},
	"yearweek":          &yearWeekFunc{},
}

func (f *SQLScalarFunctionExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	sf, ok := scalarFuncMap[f.Name]
	if ok {
		err := sf.Validate(len(f.Exprs))
		if err != nil {
			return nil, fmt.Errorf("%v '%v'", err.Error(), f.Name)
		}

		values, err := evaluateArgs(f.Exprs, ctx)
		if err != nil {
			return nil, err
		}

		return sf.Evaluate(values, ctx)
	}

	return nil, fmt.Errorf("scalar function '%v' is not supported", string(f.Name))
}

func (f *SQLScalarFunctionExpr) normalize() node {
	if sf, ok := scalarFuncMap[f.Name]; ok {
		if nsf, ok := sf.(normalizingScalarFunc); ok {
			return nsf.normalize(f)
		}
	}

	return f
}

func (f *SQLScalarFunctionExpr) RequiresEvalCtx() bool {
	if sf, ok := scalarFuncMap[f.Name]; ok {
		if r, ok := sf.(RequiresEvalCtx); ok {
			return r.RequiresEvalCtx()
		}
	}

	return false
}

func (f *SQLScalarFunctionExpr) String() string {
	var exprs []string
	for _, expr := range f.Exprs {
		exprs = append(exprs, expr.String())
	}
	return fmt.Sprintf("%s(%v)", f.Name, strings.Join(exprs, ","))
}

func (f *SQLScalarFunctionExpr) Type() schema.SQLType {
	sf, ok := scalarFuncMap[f.Name]
	if ok {
		return sf.Type()
	}

	return schema.SQLNone
}
