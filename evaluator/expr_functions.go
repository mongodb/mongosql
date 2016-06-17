package evaluator

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/10gen/sqlproxy/schema"
)

var (
	ErrIncorrectVarCount = errors.New("incorrect variable parameter count in the call to native function")
	ErrIncorrectCount    = errors.New("incorrect parameter count in function")
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
	case "std", "stddev", "stddev_pop":
		return f.stdFunc(ctx, distinctMap, false)
	case "stddev_samp":
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
	case "avg", "sum", "std", "stddev", "stddev_pop", "stddev_samp":
		switch f.Exprs[0].Type() {
		case schema.SQLInt, schema.SQLInt64:
			// TODO: this should return a decimal when we have decimal support
			return schema.SQLFloat
		default:
			return schema.SQLFloat
		}
	case "count":
		return schema.SQLInt
	}

	return f.Exprs[0].Type()
}

func (f *SQLAggFunctionExpr) avgFunc(ctx *EvalCtx, distinctMap map[interface{}]bool) (SQLValue, error) {
	var sum SQLNumeric = SQLFloat(0)
	count := 0
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

			n, err := convertToSQLNumeric(eval, ctx)
			if err == nil && n != nil {
				sum = sum.Add(n)
			}
		}
	}

	if count == 0 {
		return SQLNull, nil
	}

	return SQLFloat(sum.Float64() / float64(count)), nil
}

func (f *SQLAggFunctionExpr) countFunc(ctx *EvalCtx, distinctMap map[interface{}]bool) (SQLValue, error) {
	var count int64
	for _, row := range ctx.Rows {
		evalCtx := NewEvalCtx(ctx.ExecutionCtx, row)
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
				count++
			}
		}
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

	var sum SQLNumeric = SQLFloat(0)
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
					// already in our distinct map, so we skip this row
					continue
				} else {
					distinctMap[eval] = true
				}
			}

			n, err := convertToSQLNumeric(eval, ctx)
			if err == nil && n != nil {
				sum = sum.Add(n)
			}
		}
	}

	if allNull {
		return SQLNull, nil
	}

	return SQLFloat(sum.Float64()), nil
}

func (f *SQLAggFunctionExpr) stdFunc(ctx *EvalCtx, distinctMap map[interface{}]bool, isSamp bool) (SQLValue, error) {
	var sum SQLNumeric = SQLFloat(0)
	var data []SQLNumeric
	var diff float64 = 0.0
	var avg float64
	count := 0
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

			n, err := convertToSQLNumeric(eval, ctx)
			if err == nil && n != nil {
				sum = sum.Add(n)
				data = append(data, n)
			}
		}
	}

	if count == 0 {
		return SQLNull, nil
	}

	avg = sum.Float64() / float64(count)

	for _, val := range data {
		diff += math.Pow(val.Float64()-avg, 2)
	}

	// Sample standard deviation
	if isSamp && count == 1 {
		return SQLNull, nil
	} else if isSamp {
		return SQLFloat(math.Sqrt(diff / float64(count-1))), nil
	}
	// Population standard deviation
	return SQLFloat(math.Sqrt(diff / float64(count))), nil
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

var scalarFuncMap = map[string]scalarFunc{
	"abs":               &absFunc{},
	"ascii":             &asciiFunc{},
	"cast":              &castFunc{},
	"char_length":       &lengthFunc{},
	"coalesce":          &coalesceFunc{},
	"concat":            &concatFunc{},
	"concat_ws":         &concatWsFunc{},
	"connection_id":     &connectionIdFunc{},
	"current_date":      &currentDateFunc{},
	"current_timestamp": &currentTimestampFunc{},
	"current_user":      &userFunc{},
	"database":          &dbFunc{},
	"day":               &dayOfMonthFunc{},
	"dayname":           &dayNameFunc{},
	"dayofmonth":        &dayOfMonthFunc{},
	"dayofweek":         &dayOfWeekFunc{},
	"dayofyear":         &dayOfYearFunc{},
	"exp":               &expFunc{},
	"floor":             &floorFunc{},
	"hour":              &hourFunc{},
	"if":                &ifFunc{},
	"ifnull":            &ifnullFunc{},
	"instr":             &instrFunc{},
	"isnull":            &isnullFunc{},
	"lcase":             &lcaseFunc{},
	"left":              &leftFunc{},
	"length":            &lengthFunc{},
	"ln":                &naturalLogFunc{},
	"locate":            &locateFunc{},
	"log":               &naturalLogFunc{},
	"log2":              &log2Func{},
	"log10":             &log10Func{},
	"lower":             &lcaseFunc{},
	"ltrim":             &ltrimFunc{},
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
	"substr":            &substringFunc{},
	"substring":         &substringFunc{},
	"system_user":       &userFunc{},
	"ucase":             &ucaseFunc{},
	"upper":             &ucaseFunc{},
	"user":              &userFunc{},
	"version":           &versionFunc{},
	"week":              &weekFunc{},
	"year":              &yearFunc{},
}

func (f *SQLScalarFunctionExpr) RequiresEvalCtx() bool {
	if sf, ok := scalarFuncMap[f.Name]; ok {
		if r, ok := sf.(RequiresEvalCtx); ok {
			return r.RequiresEvalCtx()
		}
	}

	return false
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
