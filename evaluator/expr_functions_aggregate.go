package evaluator

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/option"
	"github.com/shopspring/decimal"
)

var (
	errIncorrectVarCount = errors.New(
		"incorrect variable parameter count in the call to native function")
	errIncorrectCount = errors.New(
		"incorrect parameter count in function")
)

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

// SQLAggFunctionExpr is an interface to group together SQL Aggregation Functions.
type SQLAggFunctionExpr interface {
	// Inherit SQLExpr.
	SQLExpr
	// Basic implementor function.
	iSQLAggFunctionExpr()
	// Distinct returns whether this function operates on only distinct values,
	// e.g., `sum(distinct a) from foo` vs `sum(a) from foo`.
	Distinct() bool
	// Exprs returns the argument expressions to this function.
	Exprs() []SQLExpr
	// Name returns the name of the function.
	Name() string
}

// basicSQLAggFunctionToString() is a helper to convert SQLAggFunctions to strings.
func basicSQLAggFunctionToString(name string, distinct bool, exprs SQLExprs) string {
	distinctStr := ""
	if distinct {
		distinctStr = "distinct "
	}
	return fmt.Sprintf("%s(%s%v)", name, distinctStr, exprs.String())
}

// floatingPointAggregationFunctionEvalType is used to find the EvalType
// of any floating point returning aggregation function, which will
// be double unless the argument is decimal, and decimal if it
// is decimal.
func floatingPointAggregationFunctionEvalType(e EvalType) EvalType {
	if e == EvalDecimal128 {
		return EvalDecimal128
	}
	return EvalDouble
}

// SQLAvgFunctionExpr computes average.
type SQLAvgFunctionExpr struct {
	distinct bool
	exprs    []SQLExpr
}

func (*SQLAvgFunctionExpr) iSQLAggFunctionExpr() {}

var _ translatableToAggregation = (*SQLAvgFunctionExpr)(nil)

// Distinct returns true if this aggregate function operates only on
// distinct values and false otherwise.
func (f *SQLAvgFunctionExpr) Distinct() bool {
	return f.distinct
}

// EvalType for SQLAvgFunctionExpr is the standard floatingPointAggregationFunction.
func (f *SQLAvgFunctionExpr) EvalType() EvalType {
	return floatingPointAggregationFunctionEvalType(f.exprs[0].EvalType())
}

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLAvgFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLAvgFunctionExpr(%s)", f.Name())
}

// Exprs returns the argument expressions to the function.
func (f *SQLAvgFunctionExpr) Exprs() []SQLExpr {
	return f.exprs
}

// Evaluate does in memory evaluation for SQLAvgFunctionExpr.
func (f *SQLAvgFunctionExpr) Evaluate(ctx context.Context,
	cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var distinctMap map[interface{}]bool
	if f.distinct {
		distinctMap = make(map[interface{}]bool)
	}
	count := 0.0
	sum := decimal.Zero
	isDecimal := false
	floatSum, correction := 0.0, 0.0
	for _, row := range st.rows {
		subSt := st.WithRows(row)
		for _, expr := range f.exprs {
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

			// handle Avg(X) overflowing float64 range
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

// Name returns name.
func (*SQLAvgFunctionExpr) Name() string {
	return "avg"
}

// String converts to string.
func (f *SQLAvgFunctionExpr) String() string {
	return basicSQLAggFunctionToString(f.Name(), f.distinct, f.exprs)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate in a $match stage via $expr.
func (f *SQLAvgFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// ToAggregationLanguage generates aggregation language for the aggregation function.
func (f *SQLAvgFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	// We cannot average date types.
	dataType := f.exprs[0].EvalType()
	if dataType == EvalDatetime || dataType == EvalDate {
		return nil, newPushdownFailure(f.Name(), fmt.Sprintf("cannot pushdown avg for date types"))
	}
	transExpr, err := t.ToAggregationLanguage(f.exprs[0])
	if err != nil || transExpr == nil {
		return nil, err
	}
	return bsonutil.NewD(bsonutil.NewDocElem("$avg", transExpr)), nil
}

// SQLCountFunctionExpr counts.
type SQLCountFunctionExpr struct {
	distinct bool
	exprs    []SQLExpr
}

func (*SQLCountFunctionExpr) iSQLAggFunctionExpr() {}

var _ translatableToAggregation = (*SQLCountFunctionExpr)(nil)

// Distinct returns true if this aggregate function operates only on
// distinct values and false otherwise.
func (f *SQLCountFunctionExpr) Distinct() bool {
	return f.distinct
}

// EvalType for SQLCountFunctionExpr is always EvalInt64.
func (*SQLCountFunctionExpr) EvalType() EvalType {
	return EvalInt64
}

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLCountFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLCountFunctionExpr(%s)", f.Name())
}

// Exprs returns the argument expressions to the function.
func (f *SQLCountFunctionExpr) Exprs() []SQLExpr {
	return f.exprs
}

// Evaluate does in memory evaluation for SQLCountFunctionExpr
func (f *SQLCountFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var distinctMap map[interface{}]bool
	if f.distinct {
		distinctMap = make(map[interface{}]bool)
	}
	count := uint64(0)
	fCount := float64(math.MaxUint64)
	dCount := decimal.NewFromFloat(math.MaxFloat64)
	inDecimalRange, decimalOne := false, decimal.NewFromFloat(1.0)

	for _, row := range st.rows {
		subSt := st.WithRows(row)
		for _, expr := range f.exprs {
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

// Name returns name.
func (*SQLCountFunctionExpr) Name() string {
	return "count"
}

// String converts to string.
func (f *SQLCountFunctionExpr) String() string {
	return basicSQLAggFunctionToString(f.Name(), f.distinct, f.exprs)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate in a $match stage via $expr.
func (f *SQLCountFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// ToAggregationLanguage generates aggregation language for the aggregation function.
func (f *SQLCountFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	transExpr, err := t.ToAggregationLanguage(f.exprs[0])
	if err != nil || transExpr == nil {
		return nil, err
	}
	if f.exprs[0] == NewSQLVarchar(t.valueKind(), "") {
		return bsonutil.NewD(bsonutil.NewDocElem("$size", transExpr)), nil
	}

	// The below ensure that nulls, undefined, and missing fields
	// are not part of the count.
	return bsonutil.WrapInOp(bsonutil.OpSum,
		bsonutil.WrapInMap(
			transExpr,
			"i",
			bsonutil.WrapInCond(
				0,
				1,
				bsonutil.WrapInOp(bsonutil.OpLte, "$$i", nil),
			),
		),
	), nil
}

// SQLGroupConcatFunctionExpr is the GROUP_CONCAT function in mysql.
type SQLGroupConcatFunctionExpr struct {
	distinct          bool
	exprs             []SQLExpr
	Separator         option.String
	GroupConcatMaxLen int
}

func (*SQLGroupConcatFunctionExpr) iSQLAggFunctionExpr() {}

var _ translatableToAggregation = (*SQLGroupConcatFunctionExpr)(nil)

// Distinct returns true if this aggregate function operates only on
// distinct values and false otherwise.
func (f *SQLGroupConcatFunctionExpr) Distinct() bool {
	return f.distinct
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate in a $match stage via $expr.
func (f *SQLGroupConcatFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// ToAggregationLanguage generates aggregation language for the aggregation function.
func (f *SQLGroupConcatFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	transExpr, err := t.ToAggregationLanguage(f.exprs[0])
	if err != nil || transExpr == nil {
		return nil, err
	}

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
			bsonutil.WrapInOp(bsonutil.OpGte,
				bsonutil.NewD(bsonutil.NewDocElem("$strLenCP", "$$value")),
				maxlen),
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

	return bsonutil.WrapInLet(bsonutil.NewD(
		bsonutil.NewDocElem("result", result)),
		truncateOrNil), nil
}

// EvalType for SQLGroupConcatFunctionExpr always returns EvalString.
func (f *SQLGroupConcatFunctionExpr) EvalType() EvalType {
	return EvalString
}

func addBufferEntry(buf *bytes.Buffer, value string, sep string, firstWrite *bool) {
	if *firstWrite {
		buf.WriteString(fmt.Sprintf("%v", value))
		*firstWrite = false
	} else {
		buf.WriteString(fmt.Sprintf("%s%v", sep, value))
	}
}

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLGroupConcatFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLGroupConcatFunctionExpr(%s)", f.Name())
}

// Exprs returns the argument expressions to the function.
func (f *SQLGroupConcatFunctionExpr) Exprs() []SQLExpr {
	return f.exprs
}

// Evaluate does in memory computation for SQLGroupConcatFunctionExpr.
func (f *SQLGroupConcatFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var distinctMap map[interface{}]bool
	if f.distinct {
		distinctMap = make(map[interface{}]bool)
	}
	var b bytes.Buffer
	separator := f.Separator.Else(",")
	maxResultLen := f.GroupConcatMaxLen

	var resultHasEmpty bool
	firstWrite := true
	for _, row := range st.rows {
		subSt := st.WithRows(row)

		var r bytes.Buffer
		var entryHasEmpty bool
		for _, expr := range f.exprs {
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

// Name returns name.
func (*SQLGroupConcatFunctionExpr) Name() string {
	return "group_concat"
}

// String converts to a string.
func (f *SQLGroupConcatFunctionExpr) String() string {
	var distinct, separator string
	if f.distinct {
		distinct = "distinct "
	}
	if f.Separator.IsSome() {
		separator = ` separator "` + f.Separator.Unwrap() + `"`
	}
	return fmt.Sprintf("%s(%s%v%s)", f.Name(), distinct, SQLExprs(f.exprs).String(), separator)
}

// SQLMaxFunctionExpr is a function that finds the maximal element.
type SQLMaxFunctionExpr struct {
	distinct bool
	exprs    []SQLExpr
}

func (*SQLMaxFunctionExpr) iSQLAggFunctionExpr() {}

var _ translatableToAggregation = (*SQLMaxFunctionExpr)(nil)

// Distinct returns true if this aggregate function operates only on
// distinct values and false otherwise.
func (f *SQLMaxFunctionExpr) Distinct() bool {
	return f.distinct
}

// EvalType for SQLMaxFunctionExpr returns the type of e.
func (f *SQLMaxFunctionExpr) EvalType() EvalType {
	return f.exprs[0].EvalType()
}

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLMaxFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLMaxFunctionExpr(%s)", f.Name())
}

// Exprs returns the argument expressions to the function.
func (f *SQLMaxFunctionExpr) Exprs() []SQLExpr {
	return f.exprs
}

// Evaluate for SQLMaxFunctionExpr does in memory computation for max.
func (f *SQLMaxFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	max := NewSQLNull(cfg.sqlValueKind, f.EvalType())
	for _, row := range st.rows {
		subSt := st.WithRows(row)
		for _, expr := range f.exprs {
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

// Name returns name.
func (*SQLMaxFunctionExpr) Name() string {
	return "max"
}

// String converts to string.
func (f *SQLMaxFunctionExpr) String() string {
	return basicSQLAggFunctionToString(f.Name(), f.distinct, f.exprs)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate in a $match stage via $expr.
func (f *SQLMaxFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// ToAggregationLanguage generates aggregation language for the aggregation function.
func (f *SQLMaxFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	transExpr, err := t.ToAggregationLanguage(f.exprs[0])
	if err != nil || transExpr == nil {
		return nil, err
	}

	return bsonutil.NewD(bsonutil.NewDocElem("$max", transExpr)), nil
}

// SQLMinFunctionExpr is a function that finds the minimal element.
type SQLMinFunctionExpr struct {
	distinct bool
	exprs    []SQLExpr
}

func (*SQLMinFunctionExpr) iSQLAggFunctionExpr() {}

var _ translatableToAggregation = (*SQLMinFunctionExpr)(nil)

// Distinct returns true if this aggregate function operates only on
// distinct values and false otherwise.
func (f *SQLMinFunctionExpr) Distinct() bool {
	return f.distinct
}

// EvalType for SQLMinFunctionExpr returns the type of e.
func (f *SQLMinFunctionExpr) EvalType() EvalType {
	return f.exprs[0].EvalType()
}

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLMinFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLMinFunctionExpr(%s)", f.Name())
}

// Exprs returns the argument expressions to the function.
func (f *SQLMinFunctionExpr) Exprs() []SQLExpr {
	return f.exprs
}

// Evaluate for SQLMinFunctionExpr computes the minimal element in memory.
func (f *SQLMinFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	min := NewSQLNull(cfg.sqlValueKind, f.EvalType())
	for _, row := range st.rows {
		subSt := st.WithRows(row)
		for _, expr := range f.exprs {
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

// Name returns name.
func (*SQLMinFunctionExpr) Name() string {
	return "min"
}

// String converts to string.
func (f *SQLMinFunctionExpr) String() string {
	return basicSQLAggFunctionToString(f.Name(), f.distinct, f.exprs)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate in a $match stage via $expr.
func (f *SQLMinFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// ToAggregationLanguage generates aggregation language for the aggregation function.
func (f *SQLMinFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	transExpr, err := t.ToAggregationLanguage(f.exprs[0])
	if err != nil || transExpr == nil {
		return nil, err
	}

	return bsonutil.NewD(bsonutil.NewDocElem("$min", transExpr)), nil
}

// SQLSumFunctionExpr computes the summation of elements.
type SQLSumFunctionExpr struct {
	distinct bool
	exprs    []SQLExpr
}

func (*SQLSumFunctionExpr) iSQLAggFunctionExpr() {}

var _ translatableToAggregation = (*SQLSumFunctionExpr)(nil)

// Distinct returns true if this aggregate function operates only on
// distinct values and false otherwise.
func (f *SQLSumFunctionExpr) Distinct() bool {
	return f.distinct
}

// EvalType for SQLSumFunctionExpr is a standard floating point aggregation.
func (f *SQLSumFunctionExpr) EvalType() EvalType {
	return floatingPointAggregationFunctionEvalType(f.exprs[0].EvalType())
}

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLSumFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLSumFunctionExpr(%s)", f.Name())
}

// Exprs returns the argument expressions to the function.
func (f *SQLSumFunctionExpr) Exprs() []SQLExpr {
	return f.exprs
}

// Evaluate for SQLSumFunctionExpr computes summations in memory.
func (f *SQLSumFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var distinctMap map[interface{}]bool
	if f.distinct {
		distinctMap = make(map[interface{}]bool)
	}
	floatSum, correction := 0.0, 0.0
	isDecimal := false
	sum := decimal.Zero
	allNull := true
	for _, row := range st.rows {
		subSt := st.WithRows(row)
		for _, expr := range f.exprs {
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

// Name returns name.
func (*SQLSumFunctionExpr) Name() string {
	return "sum"
}

// String converts to string.
func (f *SQLSumFunctionExpr) String() string {
	return basicSQLAggFunctionToString(f.Name(), f.distinct, f.exprs)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate in a $match stage via $expr.
func (f *SQLSumFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// ToAggregationLanguage generates aggregation language for the aggregation function.
func (f *SQLSumFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	// We cannot sum date types.
	dataType := f.exprs[0].EvalType()
	if dataType == EvalDatetime || dataType == EvalDate {
		return nil, newPushdownFailure(f.Name(), fmt.Sprintf("cannot pushdown sum for date types"))
	}
	transExpr, err := t.ToAggregationLanguage(f.exprs[0])
	if err != nil || transExpr == nil {
		return nil, err
	}
	return bsonutil.NewD(bsonutil.NewDocElem("$sum", transExpr)), nil
}

// SQLStdDevFunctionExpr computes a normal standard distribution for a population.
type SQLStdDevFunctionExpr struct {
	// StdDev has multiple names and we want to recover the one actually used
	// for display purposes.
	name     string
	distinct bool
	exprs    []SQLExpr
}

func (*SQLStdDevFunctionExpr) iSQLAggFunctionExpr() {}

var _ translatableToAggregation = (*SQLStdDevFunctionExpr)(nil)

// Distinct returns true if this aggregate function operates only on
// distinct values and false otherwise.
func (f *SQLStdDevFunctionExpr) Distinct() bool {
	return f.distinct
}

// EvalType returns the type of the value this aggregate expression evaluates to.
func (f *SQLStdDevFunctionExpr) EvalType() EvalType {
	return floatingPointAggregationFunctionEvalType(f.exprs[0].EvalType())
}

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLStdDevFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLStdDevFunctionExpr(%s)", f.Name())
}

// Exprs returns the argument expressions to the function.
func (f *SQLStdDevFunctionExpr) Exprs() []SQLExpr {
	return f.exprs
}

// Evaluate for SQLStdDevFunctionExpr computes the standard deviation of a population
// in memory.
func (f *SQLStdDevFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var distinctMap map[interface{}]bool
	if f.distinct {
		distinctMap = make(map[interface{}]bool)
	}
	var data []SQLValue
	sum := decimal.Zero
	floatSum, correction, count := 0.0, 0.0, 0.0
	isDecimal := false
	for _, row := range st.rows {
		subSt := st.WithRows(row)
		for _, expr := range f.exprs {
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

		diff = diff.Div(decimal.NewFromFloat(count))
		f, _ := diff.Float64()
		return NewSQLDecimal128(cfg.sqlValueKind, decimal.NewFromFloat(math.Sqrt(f))), nil
	}

	avg := floatSum / count
	diff := 0.0

	for _, val := range data {
		diff += math.Pow(Float64(val)-avg, 2)
	}

	return NewSQLFloat(cfg.sqlValueKind, math.Sqrt(diff/count)), nil
}

// Name returns name.
func (f *SQLStdDevFunctionExpr) Name() string {
	return f.name
}

// String converts to string.
func (f *SQLStdDevFunctionExpr) String() string {
	return basicSQLAggFunctionToString(f.name, f.distinct, f.exprs)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate in a $match stage via $expr.
func (f *SQLStdDevFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// ToAggregationLanguage generates aggregation language for the aggregation function.
func (f *SQLStdDevFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	// We cannot stddev date types.
	dataType := f.exprs[0].EvalType()
	if dataType == EvalDatetime || dataType == EvalDate {
		return nil, newPushdownFailure(f.Name(), fmt.Sprintf("cannot pushdown std for date types"))
	}
	transExpr, err := t.ToAggregationLanguage(f.exprs[0])
	if err != nil || transExpr == nil {
		return nil, err
	}
	return bsonutil.NewD(bsonutil.NewDocElem("$stdDevPop", transExpr)), nil
}

// SQLStdDevSampleFunctionExpr computes standard deviation of a sample.
type SQLStdDevSampleFunctionExpr struct {
	distinct bool
	exprs    []SQLExpr
}

func (*SQLStdDevSampleFunctionExpr) iSQLAggFunctionExpr() {}

var _ translatableToAggregation = (*SQLStdDevSampleFunctionExpr)(nil)

// Distinct returns true if this aggregate function operates only on
// distinct values and false otherwise.
func (f *SQLStdDevSampleFunctionExpr) Distinct() bool {
	return f.distinct
}

// EvalType returns the type of the value this aggregate expression evaluates to.
func (f *SQLStdDevSampleFunctionExpr) EvalType() EvalType {
	return floatingPointAggregationFunctionEvalType(f.exprs[0].EvalType())
}

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLStdDevSampleFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLStdDevSampleFunctionExpr(%s)", f.Name())
}

// Exprs returns the argument expressions to the function.
func (f *SQLStdDevSampleFunctionExpr) Exprs() []SQLExpr {
	return f.exprs
}

// Evaluate for SQLStdDevSampleFunctionExpr computes standard deviation for
// a sample in memory.
func (f *SQLStdDevSampleFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	var data []SQLValue
	var distinctMap map[interface{}]bool
	if f.distinct {
		distinctMap = make(map[interface{}]bool)
	}
	sum := decimal.Zero
	floatSum, correction, count := 0.0, 0.0, 0.0
	isDecimal := false
	for _, row := range st.rows {
		subSt := st.WithRows(row)
		for _, expr := range f.exprs {
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

		if count == 1 {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
		}
		diff = diff.Div(decimal.NewFromFloat(count - 1))
		f, _ := diff.Float64()
		return NewSQLDecimal128(cfg.sqlValueKind, decimal.NewFromFloat(math.Sqrt(f))), nil
	}

	avg := floatSum / count
	diff := 0.0

	for _, val := range data {
		diff += math.Pow(Float64(val)-avg, 2)
	}

	if count == 1 {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), nil
	}
	return NewSQLFloat(cfg.sqlValueKind, math.Sqrt(diff/(count-1))), nil
}

// Name returns name.
func (*SQLStdDevSampleFunctionExpr) Name() string {
	return "stddev_samp"
}

// String converts to string.
func (f *SQLStdDevSampleFunctionExpr) String() string {
	return basicSQLAggFunctionToString(f.Name(), f.distinct, f.exprs)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate in a $match stage via $expr.
func (f *SQLStdDevSampleFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// ToAggregationLanguage generates aggregation language for the aggregation function.
func (f *SQLStdDevSampleFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	// We cannot stddev date types.
	dataType := f.exprs[0].EvalType()
	if dataType == EvalDatetime || dataType == EvalDate {
		return nil, newPushdownFailure(f.Name(), fmt.Sprintf("cannot pushdown stddev_samp for date types"))
	}
	transExpr, err := t.ToAggregationLanguage(f.exprs[0])
	if err != nil || transExpr == nil {
		return nil, err
	}
	return bsonutil.NewD(bsonutil.NewDocElem("$stdDevSamp", transExpr)), nil
}
