package evaluator

import (
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

// Evaluate evaluates a SQLAggFunctionExpr to a SQLValue.
func (f *SQLAggFunctionExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	var distinctMap map[interface{}]bool
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
	ctx *EvalCtx,
	distinctMap map[interface{}]bool) (SQLValue,
	error) {
	count := 0.0

	sum := decimal.Zero

	isDecimal := false

	floatSum, correction := 0.0, 0.0

	for _, row := range ctx.Rows {
		evalCtx := ctx.WithRows(row)
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

func (f *SQLAggFunctionExpr) countFunc(
	ctx *EvalCtx,
	distinctMap map[interface{}]bool) (SQLValue,
	error) {
	count := uint64(0)
	fCount := float64(math.MaxUint64)
	dCount := decimal.NewFromFloat(math.MaxFloat64)

	inDecimalRange, decimalOne := false, decimal.NewFromFloat(1.0)

	for _, row := range ctx.Rows {
		evalCtx := ctx.WithRows(row)
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

	return SQLInt64(count), nil
}

func (f *SQLAggFunctionExpr) maxFunc(ctx *EvalCtx) (SQLValue, error) {
	var max SQLValue = SQLNull
	for _, row := range ctx.Rows {
		evalCtx := ctx.WithRows(row)
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

			compared, err := CompareTo(max, eval, ctx.Collation)
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
		evalCtx := ctx.WithRows(row)
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

			compared, err := CompareTo(min, eval, ctx.Collation)
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
	ctx *EvalCtx,
	distinctMap map[interface{}]bool) (SQLValue,
	error) {

	floatSum, correction := 0.0, 0.0

	isDecimal := false

	sum := decimal.Zero

	allNull := true

	for _, row := range ctx.Rows {
		evalCtx := ctx.WithRows(row)
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
		return SQLNull, nil
	}

	floatSum += correction

	if isDecimal {
		sum = sum.Add(decimal.NewFromFloat(floatSum))
		return SQLDecimal128(sum), nil
	}

	return SQLFloat(floatSum), nil
}

func (f *SQLAggFunctionExpr) stdFunc(
	ctx *EvalCtx,
	distinctMap map[interface{}]bool,
	isSamp bool) (SQLValue,
	error) {
	var data []SQLValue

	sum := decimal.Zero

	floatSum, correction, count := 0.0, 0.0, 0.0

	isDecimal := false

	for _, row := range ctx.Rows {
		evalCtx := ctx.WithRows(row)
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
		return SQLNull, nil
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
		diff += math.Pow(Float64(val)-avg, 2)
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

// ToAggregationLanguage translates SQLAggFunctionExpr into something that can
// be used in an aggregation pipeline. If SQLAggFunctionExpr cannot be translated,
// it will return nil and false.
func (f *SQLAggFunctionExpr) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	transExpr, ok := t.ToAggregationLanguage(f.Exprs[0])
	if !ok || transExpr == nil {
		return nil, false
	}

	name := f.Name

	// We will disallow several SQL aggregation functions over DateTime types below,
	// but count, min, and max are all safe to pushdown for DateTimes in mongo,
	// thus we do not check if the argument column is DateTime typed here
	switch name {
	case minAggregateName, maxAggregateName:
		return bson.M{"$" + name: transExpr}, true
	case countAggregateName:
		if f.Exprs[0] == SQLVarchar("*") {
			return bson.M{"$size": transExpr}, true
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
		}, true
	}

	// All other aggregate functions are not allowed over DateTime types
	dataType := f.Exprs[0].EvalType()
	if dataType == EvalDatetime || dataType == EvalDate {
		return nil, false
	}

	switch name {
	case stdAggregateName, stddevAggregateName, stddevPopAggregateName:
		return bson.M{"$stdDevPop": transExpr}, true
	case stddevSampleAggregateName:
		return bson.M{"$stdDevSamp": transExpr}, true
	default:
		return bson.M{"$" + name: transExpr}, true
	}

}
