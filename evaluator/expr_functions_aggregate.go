package evaluator

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/shopspring/decimal"
)

var (
	errIncorrectVarCount = errors.New(
		"incorrect variable parameter count in the call to native function")
	errIncorrectCount = errors.New(
		"incorrect parameter count in function")
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

var _ translatableToAggregation = (*SQLAggFunctionExpr)(nil)

// Evaluate evaluates a SQLAggFunctionExpr to a SQLValue.
func (f *SQLAggFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var distinctMap map[interface{}]bool
	if f.Distinct {
		distinctMap = make(map[interface{}]bool)
	}

	switch f.Name {
	case avgAggregateName:
		return f.avgFunc(ctx, cfg, st, distinctMap)
	case sumAggregateName:
		return f.sumFunc(ctx, cfg, st, distinctMap)
	case countAggregateName:
		return f.countFunc(ctx, cfg, st, distinctMap)
	case maxAggregateName:
		return f.maxFunc(ctx, cfg, st)
	case minAggregateName:
		return f.minFunc(ctx, cfg, st)
	case stdAggregateName, stddevAggregateName, stddevPopAggregateName:
		return f.stdFunc(ctx, cfg, st, distinctMap, false)
	case stddevSampleAggregateName:
		return f.stdFunc(ctx, cfg, st, distinctMap, true)
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

// EvalType returns the EvalType associated with SQLAggFunctionExpr.
func (f *SQLAggFunctionExpr) EvalType() EvalType {
	switch f.Name {
	case avgAggregateName,
		sumAggregateName,
		stdAggregateName,
		stddevAggregateName,
		stddevPopAggregateName,
		stddevSampleAggregateName:
		switch f.Exprs[0].EvalType() {
		case EvalInt64, EvalInt32:
			return EvalDouble
		default:
			return EvalDecimal128
		}
	case countAggregateName:
		return EvalInt64
	}

	return f.Exprs[0].EvalType()
}

func (f *SQLAggFunctionExpr) avgFunc(
	ctx context.Context,
	cfg *ExecutionConfig,
	st *ExecutionState,
	distinctMap map[interface{}]bool) (SQLValue,
	error) {

	count := 0.0

	sum := decimal.Zero

	isDecimal := false

	floatSum, correction := 0.0, 0.0

	for _, row := range st.rows {
		subSt := st.WithRows(row)
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(ctx, cfg, subSt)
			if err != nil {
				return nil, err
			}

			if eval.IsNull() {
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

			if isDecimal || eval.EvalType() == EvalDecimal128 {
				isDecimal = true
				sum = sum.Add(Decimal(eval))
				continue
			}

			floatEval := Float64(eval)

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
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	floatSum += correction

	if isDecimal {
		sum = sum.Add(decimal.NewFromFloat(floatSum))
		avg := sum.Div(decimal.NewFromFloat(count))
		return NewSQLDecimal128(cfg.sqlValueKind, avg), nil
	}

	return NewSQLFloat(cfg.sqlValueKind, floatSum/count), nil
}

func (f *SQLAggFunctionExpr) countFunc(
	ctx context.Context,
	cfg *ExecutionConfig,
	st *ExecutionState,
	distinctMap map[interface{}]bool) (SQLValue,
	error) {

	count := uint64(0)
	fCount := float64(math.MaxUint64)
	dCount := decimal.NewFromFloat(math.MaxFloat64)

	inDecimalRange, decimalOne := false, decimal.NewFromFloat(1.0)

	for _, row := range st.rows {
		subSt := st.WithRows(row)
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(ctx, cfg, subSt)
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

			if eval != nil && !eval.IsNull() {
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
		return NewSQLDecimal128(cfg.sqlValueKind, dCount), nil
	} else if count > math.MaxInt64 {
		return NewSQLFloat(cfg.sqlValueKind, fCount), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(count)), nil
}

func (f *SQLAggFunctionExpr) maxFunc(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	max := NewSQLNull(cfg.sqlValueKind, f.EvalType())
	for _, row := range st.rows {
		subSt := st.WithRows(row)
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(ctx, cfg, subSt)
			if err != nil {
				return nil, err
			}
			if !eval.IsNull() {
				if max.IsNull() {
					max = eval
					continue
				}
			} else {
				continue
			}

			compared, err := CompareTo(max, eval, subSt.collation)
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

func (f *SQLAggFunctionExpr) minFunc(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	min := NewSQLNull(cfg.sqlValueKind, f.EvalType())
	for _, row := range st.rows {
		subSt := st.WithRows(row)
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(ctx, cfg, subSt)
			if err != nil {
				return nil, err
			}

			if !eval.IsNull() {
				if min.IsNull() {
					min = eval
					continue
				}
			} else {
				continue
			}

			compared, err := CompareTo(min, eval, subSt.collation)
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

func (f *SQLAggFunctionExpr) sumFunc(
	ctx context.Context,
	cfg *ExecutionConfig,
	st *ExecutionState,
	distinctMap map[interface{}]bool) (SQLValue,
	error) {

	floatSum, correction := 0.0, 0.0

	isDecimal := false

	sum := decimal.Zero

	allNull := true

	for _, row := range st.rows {
		subSt := st.WithRows(row)
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(ctx, cfg, subSt)
			if err != nil {
				return nil, err
			}

			if eval.IsNull() {
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

			evalType := eval.EvalType()
			if isDecimal || evalType == EvalDecimal128 {
				isDecimal = true
				sum = sum.Add(Decimal(eval))
				continue
			}

			floatEval := Float64(eval)

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
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	floatSum += correction

	if isDecimal {
		sum = sum.Add(decimal.NewFromFloat(floatSum))
		return NewSQLDecimal128(cfg.sqlValueKind, sum), nil
	}

	return NewSQLFloat(cfg.sqlValueKind, floatSum), nil
}

func (f *SQLAggFunctionExpr) stdFunc(
	ctx context.Context,
	cfg *ExecutionConfig,
	st *ExecutionState,
	distinctMap map[interface{}]bool,
	isSamp bool) (SQLValue,
	error) {

	var data []SQLValue

	sum := decimal.Zero

	floatSum, correction, count := 0.0, 0.0, 0.0

	isDecimal := false

	for _, row := range st.rows {
		subSt := st.WithRows(row)
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(ctx, cfg, subSt)
			if err != nil {
				return nil, err
			}

			if eval.IsNull() {
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

			if isDecimal || eval.EvalType() == EvalDecimal128 {
				isDecimal = true
				sum = sum.Add(Decimal(eval))
				continue
			}

			floatEval := Float64(eval)

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
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	floatSum += correction

	if isDecimal {
		diff := decimal.Zero
		sum = sum.Add(decimal.NewFromFloat(floatSum))
		avg := sum.Div(decimal.NewFromFloat(count))

		for _, v := range data {
			val := Decimal(v).Sub(avg)
			diff = diff.Add(val.Mul(val))
		}

		// Sample standard deviation
		if isSamp {
			if count == 1 {
				return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
			}
			diff = diff.Div(decimal.NewFromFloat(count - 1))
			f, _ := diff.Float64()
			return NewSQLDecimal128(cfg.sqlValueKind, decimal.NewFromFloat(math.Sqrt(f))), nil
		}

		// Population standard deviation
		diff = diff.Div(decimal.NewFromFloat(count))
		f, _ := diff.Float64()
		return NewSQLDecimal128(cfg.sqlValueKind, decimal.NewFromFloat(math.Sqrt(f))), nil
	}

	avg := floatSum / count
	diff := 0.0

	for _, val := range data {
		diff += math.Pow(Float64(val)-avg, 2)
	}

	// Sample standard deviation
	if isSamp {
		if count == 1 {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
		}
		return NewSQLFloat(cfg.sqlValueKind, math.Sqrt(diff/(count-1))), nil
	}

	// Population standard deviation
	return NewSQLFloat(cfg.sqlValueKind, math.Sqrt(diff/count)), nil
}

// ToAggregationLanguage translates SQLAggFunctionExpr into something that can
// be used in an aggregation pipeline. If SQLAggFunctionExpr cannot be translated,
// it will return nil and an error.
func (f *SQLAggFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	transExpr, err := t.ToAggregationLanguage(f.Exprs[0])
	if err != nil || transExpr == nil {
		return nil, fmt.Errorf("failed to push down aggregate function %s", f.Name)
	}

	name := f.Name

	// We will disallow several SQL aggregation functions over DateTime types below,
	// but count, min, and max are all safe to pushdown for DateTimes in mongo,
	// thus we do not check if the argument column is DateTime typed here
	switch name {
	case minAggregateName, maxAggregateName:
		return bson.M{"$" + name: transExpr}, nil
	case countAggregateName:
		if f.Exprs[0] == NewSQLVarchar(t.valueKind(), "*") {
			return bson.M{"$size": transExpr}, nil
		}
		// The below ensure that nulls, undefined, and missing fields
		// are not part of the count.
		return bson.M{
			"$sum": bson.M{
				"$map": bson.M{
					"input": transExpr,
					"as":    "i",
					"in": bson.M{
						mgoOperatorCond: []interface{}{
							bson.M{mgoOperatorEq: []interface{}{
								bson.M{mgoOperatorIfNull: []interface{}{
									"$$i",
									nil}},
								nil}},
							0,
							1,
						},
					},
				},
			},
		}, nil
	}

	// All other aggregate functions are not allowed over DateTime types
	dataType := f.Exprs[0].EvalType()
	if dataType == EvalDatetime || dataType == EvalDate {
		return nil, fmt.Errorf("%v is not allowed over DateTime types", name)
	}

	switch name {
	case stdAggregateName, stddevAggregateName, stddevPopAggregateName:
		return bson.M{"$stdDevPop": transExpr}, nil
	case stddevSampleAggregateName:
		return bson.M{"$stdDevSamp": transExpr}, nil
	default:
		return bson.M{"$" + name: transExpr}, nil
	}

}
