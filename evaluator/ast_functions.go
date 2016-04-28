package evaluator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/schema"
)

var (
	ErrIncorrectVarCount = errors.New("incorrect variable parameter count in the call to native function")
	ErrIncorrectCount    = errors.New("incorrect parameter count in function")
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

func (f *SQLAggFunctionExpr) Type() schema.SQLType {
	switch f.Name {
	case "count":
		return schema.SQLInt
	default:
		return f.Exprs[0].Type()
	}
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
			if err == nil && n != nil {
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
			if err == nil && n != nil {
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

type scalarFunc interface {
	Evaluate([]SQLValue, *EvalCtx) (SQLValue, error)
	Validate(exprCount int) error
	Type() schema.SQLType
}

var scalarFuncMap = map[string]scalarFunc{
	"abs":               &absFunc{},
	"ascii":             &asciiFunc{},
	"cast":              &castFunc{},
	"coalesce":          &coalesceFunc{},
	"concat":            &concatFunc{},
	"connection_id":     &connectionIdFunc{},
	"current_date":      &currentDateFunc{},
	"current_timestamp": &currentTimestampFunc{},
	"database":          &dbFunc{},
	"dayname":           &dayNameFunc{},
	"day":               &dayOfMonthFunc{},
	"dayofmonth":        &dayOfMonthFunc{},
	"dayofweek":         &dayOfWeekFunc{},
	"dayofyear":         &dayOfYearFunc{},
	"exp":               &expFunc{},
	"floor":             &floorFunc{},
	"hour":              &hourFunc{},
	"isnull":            &isnullFunc{},
	"lcase":             &lcaseFunc{},
	"length":            &lengthFunc{},
	"locate":            &locateFunc{},
	"log10":             &log10Func{},
	"ltrim":             &ltrimFunc{},
	"minute":            &minuteFunc{},
	"mod":               &modFunc{},
	"month":             &monthFunc{},
	"monthname":         &monthNameFunc{},
	"not":               &notFunc{},
	"now":               &currentTimestampFunc{},
	"pow":               &powFunc{},
	"power":             &powFunc{},
	"quarter":           &quarterFunc{},
	"sqrt":              &sqrtFunc{},
	"rtrim":             &rtrimFunc{},
	"second":            &secondFunc{},
	"substring":         &substringFunc{},
	"ucase":             &ucaseFunc{},
	"week":              &weekFunc{},
	"year":              &yearFunc{},
}

func (f *SQLScalarFunctionExpr) RequiresEvalCtx() bool {
	switch f.Name {
	case "connection_id", "database":
		return true
	default:
		return false
	}
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
