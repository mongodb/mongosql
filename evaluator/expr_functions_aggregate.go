package evaluator

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/10gen/sqlproxy/internal/util/bsonutil"
	"github.com/10gen/sqlproxy/internal/util/option"
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
	groupConcatAggregateName  = "group_concat"
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
// functions are avg, sum, count, group_concat, max, min, std, stddev, stddev_pop, and stddev_samp.
//
type SQLAggFunctionExpr struct {
	Name              string
	Distinct          bool
	Exprs             []SQLExpr
	Separator         option.String
	GroupConcatMaxLen int
}

// NewSQLAggFunctionExpr returns a new SQLAggFunctionExpr constructed from the
// provided values. SQLAggFunctionExprs should always be constructed via this
// function instead of via a struct literal.
func NewSQLAggFunctionExpr(Name string, Distinct bool, Exprs []SQLExpr, Separator option.String, GroupConcatMaxLen int) *SQLAggFunctionExpr {
	return &SQLAggFunctionExpr{
		Name:              Name,
		Distinct:          Distinct,
		Exprs:             Exprs,
		Separator:         Separator,
		GroupConcatMaxLen: GroupConcatMaxLen,
	}
}

var _ translatableToAggregation = (*SQLAggFunctionExpr)(nil)

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLAggFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLAggFunctionExpr(%s)", f.Name)
}

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
	case groupConcatAggregateName:
		return f.groupConcatFunc(ctx, cfg, st, distinctMap)
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
	var distinct, separator string
	if f.Distinct {
		distinct = "distinct "
	}
	if f.Separator.IsSome() {
		separator = ` separator "` + f.Separator.Unwrap() + `"`
	}
	return fmt.Sprintf("%s(%s%v%s)", f.Name, distinct, SQLExprs(f.Exprs).String(), separator)
}

// SQLExprs represents a list of SQLExprs.
type SQLExprs []SQLExpr

func (s SQLExprs) String() string {
	var prefix string
	var buf bytes.Buffer
	for _, e := range s {
		buf.WriteString(fmt.Sprintf("%s%v", prefix, e))
		prefix = ", "
	}
	return buf.String()
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
	case groupConcatAggregateName:
		return EvalString
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

func addBufferEntry(buf *bytes.Buffer, value string, sep string, firstWrite *bool) {
	if *firstWrite {
		buf.WriteString(fmt.Sprintf("%v", value))
		*firstWrite = false
	} else {
		buf.WriteString(fmt.Sprintf("%s%v", sep, value))
	}
}

func (f *SQLAggFunctionExpr) groupConcatFunc(
	ctx context.Context,
	cfg *ExecutionConfig,
	st *ExecutionState,
	distinctMap map[interface{}]bool) (SQLValue,
	error) {

	var b bytes.Buffer
	separator := f.Separator.Else(",")
	maxResultLen := f.GroupConcatMaxLen

	var resultHasEmpty bool
	firstWrite := true
	for _, row := range st.rows {
		subSt := st.WithRows(row)

		var r bytes.Buffer
		var entryHasEmpty bool
		for _, expr := range f.Exprs {
			eval, err := expr.Evaluate(ctx, cfg, subSt)
			if err != nil {
				return nil, err
			}

			if !eval.IsNull() {
				r.WriteString(fmt.Sprintf("%v", eval))
				if eval.String() == "" {
					entryHasEmpty = true
				}
			} else {
				r.Reset()
				entryHasEmpty = false
				break
			}
		}

		// add non-empty elems to result
		if str := r.String(); str != "" || entryHasEmpty {
			if distinctMap != nil {
				if distinctMap[str] {
					continue
				} else {
					distinctMap[str] = true
					addBufferEntry(&b, str, separator, &firstWrite)
				}
			} else {
				addBufferEntry(&b, str, separator, &firstWrite)
			}

			if entryHasEmpty && str == "" {
				resultHasEmpty = true
			}
		}

		if b.Len() > maxResultLen {
			b.Truncate(maxResultLen)
			break
		}
	}

	// return NULL if result string empty
	if b.String() == "" && !resultHasEmpty {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}

	return NewSQLVarchar(cfg.sqlValueKind, b.String()), nil
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
func (f *SQLAggFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	name := f.Name

	// Here we translate the first expr in our SQLAggFunctionExpr. All SQLAggFunctionExprs, with the
	// exception of group_concat, take at most 1 argument. Though group_concat aggregate functions can
	// take multiple exprs as arguments, eg. group_concat(a, b), group_concat requires two steps to
	// aggregate, and the current function, ToAggregationLanguage, is not called until the second step.
	// At this point, we have already constructed the $group stage as well as a new SQLAggFunctionExpr
	// with f.Exprs representing a SQLColumnExpr with the arguments to group_concat already concatenated
	// together. In the example group_concat(a, b), f.Exprs would be the SQLColumnExpr "group_concat(a, b)",
	// representing an array of concatenated "a"'s and "b"'s.
	transExpr, err := t.ToAggregationLanguage(f.Exprs[0])
	if err != nil {
		return nil, err
	}

	// We will disallow several SQL aggregation functions over DateTime types below,
	// but count, min, and max are all safe to pushdown for DateTimes in mongo,
	// thus we do not check if the argument column is DateTime typed here
	switch name {
	case minAggregateName, maxAggregateName:
		return bsonutil.NewM(bsonutil.NewDocElem("$"+name, transExpr)), nil
	case countAggregateName:
		if f.Exprs[0] == NewSQLVarchar(t.valueKind(), "*") {
			return bsonutil.NewM(bsonutil.NewDocElem("$size", transExpr)), nil
		}
		// The below ensure that nulls, undefined, and missing fields
		// are not part of the count.
		return bsonutil.NewM(
			bsonutil.NewDocElem("$sum", bsonutil.NewM(
				bsonutil.NewDocElem("$map", bsonutil.NewM(
					bsonutil.NewDocElem("input", transExpr),
					bsonutil.NewDocElem("as", "i"),
					bsonutil.NewDocElem("in", bsonutil.NewM(
						bsonutil.NewDocElem(bsonutil.OpCond, bsonutil.NewArray(
							bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
								bsonutil.NewM(
									bsonutil.NewDocElem(bsonutil.OpIfNull, bsonutil.NewArray(
										"$$i",
										nil,
									)),
								),
								nil,
							))),
							0,
							1,
						)),
					)),
				)),
			)),
		), nil
	case groupConcatAggregateName:
		maxlen := f.GroupConcatMaxLen
		separator := f.Separator.Else(",")

		// The first time we add something to the list, we don't include a separator.
		firstConcat := bsonutil.WrapInCond(nil,
			"$$this",
			bsonutil.WrapInOp(bsonutil.OpEq, "$$this", nil),
		)

		// The default behavior for adding a new entry to the list is to precede the
		// entry with a separator. We also check whether the length of the result string
		// has already reached group_concat_max_len, in which case we stop adding entries
		// to the result string.
		defaultConcat := bsonutil.WrapInCond("$$value",
			bsonutil.WrapInCond("$$value",
				bsonutil.WrapInOp(bsonutil.OpConcat, "$$value", separator, "$$this"),
				bsonutil.WrapInOp(bsonutil.OpGte, bsonutil.NewM(bsonutil.NewDocElem("$strLenCP", "$$value")), maxlen),
			),
			bsonutil.WrapInOp(bsonutil.OpEq, "$$this", nil),
		)

		result := bsonutil.WrapInReduce(transExpr,
			nil,
			bsonutil.WrapInCond(firstConcat,
				defaultConcat,
				bsonutil.WrapInOp(bsonutil.OpEq, "$$value", nil),
			),
		)

		// We must check whether the result is nil because $substr will translate a nil
		// argument into an empty string.
		truncateOrNil := bsonutil.WrapInCond(nil,
			bsonutil.WrapInSubstr("$$result", 0, maxlen),
			bsonutil.WrapInOp(bsonutil.OpEq, "$$result", nil),
		)

		return bsonutil.WrapInLet(bsonutil.NewM(bsonutil.NewDocElem("result", result)), truncateOrNil), nil
	}

	// All other aggregate functions are not allowed over DateTime types
	dataType := f.Exprs[0].EvalType()
	if dataType == EvalDatetime || dataType == EvalDate {
		return nil, newPushdownFailure(
			fmt.Sprintf("SQLAggFunctionExpr(%s)", f.Name),
			"not allowed over DateTime types",
		)
	}

	switch name {
	case stdAggregateName, stddevAggregateName, stddevPopAggregateName:
		return bsonutil.NewM(bsonutil.NewDocElem("$stdDevPop", transExpr)), nil
	case stddevSampleAggregateName:
		return bsonutil.NewM(bsonutil.NewDocElem("$stdDevSamp", transExpr)), nil
	default:
		return bsonutil.NewM(bsonutil.NewDocElem("$"+name, transExpr)), nil
	}

}
