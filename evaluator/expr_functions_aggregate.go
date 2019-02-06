package evaluator

import (
	"bytes"
	"context"
	"fmt"
	"math"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/option"
	"github.com/shopspring/decimal"
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

type baseAggFunctionExpr struct {
	distinct bool
	exprs    []SQLExpr
}

func (baseAggFunctionExpr) iSQLAggFunctionExpr() {}

func (b baseAggFunctionExpr) Distinct() bool {
	return b.distinct
}

func (b baseAggFunctionExpr) Exprs() []SQLExpr {
	return b.exprs
}

// Children returns a slice of all the Node children of the Node.
func (b baseAggFunctionExpr) Children() []Node {
	out := make([]Node, len(b.exprs))
	for i := range b.exprs {
		out[i] = b.exprs[i]
	}
	return out
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (b baseAggFunctionExpr) ReplaceChild(i int, n Node) {
	if i < 0 || i >= len(b.exprs) {
		panicWithInvalidIndex("baseAggFunctionExpr", i, len(b.exprs)-1)
	}
	b.exprs[i] = panicIfNotSQLExpr("baseAggFunctionExpr", n)
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
func floatingPointAggregationFunctionEvalType(e types.EvalType) types.EvalType {
	if e == types.EvalDecimal128 {
		return types.EvalDecimal128
	}
	return types.EvalDouble
}

// nullCheckAndWrapInAggOp returns a document for the provided aggregation function
// with the operand wrapped in a null-check $let binding (if one is needed). It also clears
// the PushdownTranslator's column references.
//
// In general, columns that need to be null-checked anywhere in a pipeline are null-checked
// in an outer $let binding. The $group stage has the following prototype form:
//   { $group: { _id: <expr>, <field1>: { <accumulator>: <expr> }, ... } }
// The <exprs> may be wrapped in $let, but the outer <accumulators> cannot.
// This function is used for this situation: the operand of the accumulator (<expr>) is
// wrapped in the $let instead of the entire accumulator.
//
// For example, nullCheckAndWrapInAggOp("$avg", { $cond: [$$aIsNull, 0, $a] }, t) produces:
//   {
//     $avg: {
//       $let: {
//         vars: { aIsNull: { $lte: [$a, null] } },
//         in: { $cond: [$$aIsNull, 0, $a] }
//       }
//   }
// as opposed to the INVALID translation:
//  {
//    $let: {
//      vars: { aIsNull: { $lte: [$a, null] } },
//      in: { $avg: { $cond: [$$aIsNull, 0, $a] } }
//    }
//  }
func (t *PushdownTranslator) nullCheckAndWrapInAggOp(aggOp string, operand interface{}) bson.D {
	nullCheckedOperand := t.withNullCheckedColumnsScope(operand)
	t.ClearColumnsToNullCheck()
	return bsonutil.NewD(bsonutil.NewDocElem(aggOp, nullCheckedOperand))
}

// SQLAvgFunctionExpr computes average.
type SQLAvgFunctionExpr struct {
	baseAggFunctionExpr
}

// NewSQLAvgFunctionExpr is a constructor for SQLAvgFunctionExpr.
func NewSQLAvgFunctionExpr(distinct bool, exprs []SQLExpr) *SQLAvgFunctionExpr {
	return &SQLAvgFunctionExpr{baseAggFunctionExpr{
		distinct: distinct,
		exprs:    exprs,
	}}
}

// EvalType for SQLAvgFunctionExpr is the standard floatingPointAggregationFunction.
func (f *SQLAvgFunctionExpr) EvalType() types.EvalType {
	return floatingPointAggregationFunctionEvalType(f.exprs[0].EvalType())
}

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLAvgFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLAvgFunctionExpr(%s)", f.Name())
}

// Evaluate does in memory evaluation for SQLAvgFunctionExpr.
func (f *SQLAvgFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(f)
	if err != nil {
		return nil, err
	}

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

			if isDecimal || eval.EvalType() == types.EvalDecimal128 {
				isDecimal = true
				sum = sum.Add(values.Decimal(eval))
				continue
			}

			floatEval := values.Float64(eval)

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
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	floatSum += correction

	if isDecimal {
		sum = sum.Add(decimal.NewFromFloat(floatSum))
		avg := sum.Div(decimal.NewFromFloat(count))
		return values.NewSQLDecimal128(cfg.sqlValueKind, avg), nil
	}

	return values.NewSQLFloat(cfg.sqlValueKind, floatSum/count), nil
}

// Name returns name.
func (*SQLAvgFunctionExpr) Name() string {
	return "avg"
}

// nolint: unparam
func (f *SQLAvgFunctionExpr) reconcile() (SQLExpr, error) {
	return f, nil
}

// String converts to string.
func (f *SQLAvgFunctionExpr) String() string {
	return basicSQLAggFunctionToString(f.Name(), f.distinct, f.exprs)
}

// FoldConstants simplifies *SQLAvgFunctionExpr based on statically known constants.
func (f *SQLAvgFunctionExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(f); err != nil {
		return nil, err
	}

	if hasNullExpr(f.exprs[0]) {
		return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
	}
	return f, nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (f *SQLAvgFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// ToAggregationLanguage generates aggregation language for the aggregation function.
func (f *SQLAvgFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	// We cannot average date types.
	dataType := f.exprs[0].EvalType()
	if dataType == types.EvalDatetime || dataType == types.EvalDate {
		return nil, newPushdownFailure(f.Name(), fmt.Sprintf("cannot pushdown avg for date types"))
	}
	transExpr, err := t.ToAggregationLanguage(f.exprs[0])
	if err != nil || transExpr == nil {
		return nil, err
	}
	return t.nullCheckAndWrapInAggOp(bsonutil.OpAvg, transExpr), nil
}

// SQLCountFunctionExpr counts.
type SQLCountFunctionExpr struct {
	baseAggFunctionExpr
}

// NewSQLCountFunctionExpr is a constructor for SQLCountFunctionExpr.
func NewSQLCountFunctionExpr(distinct bool, exprs []SQLExpr) *SQLCountFunctionExpr {
	return &SQLCountFunctionExpr{baseAggFunctionExpr{
		distinct: distinct,
		exprs:    exprs,
	}}
}

// EvalType for SQLCountFunctionExpr is always types.EvalInt64.
func (*SQLCountFunctionExpr) EvalType() types.EvalType {
	return types.EvalInt64
}

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLCountFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLCountFunctionExpr(%s)", f.Name())
}

// Evaluate does in memory evaluation for SQLCountFunctionExpr
func (f *SQLCountFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(f)
	if err != nil {
		return nil, err
	}

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
		return values.NewSQLDecimal128(cfg.sqlValueKind, dCount), nil
	} else if count > math.MaxInt64 {
		return values.NewSQLFloat(cfg.sqlValueKind, fCount), nil
	}

	return values.NewSQLInt64(cfg.sqlValueKind, int64(count)), nil
}

// Name returns name.
func (*SQLCountFunctionExpr) Name() string {
	return "count"
}

// nolint: unparam
func (f *SQLCountFunctionExpr) reconcile() (SQLExpr, error) {
	return f, nil
}

// String converts to string.
func (f *SQLCountFunctionExpr) String() string {
	return basicSQLAggFunctionToString(f.Name(), f.distinct, f.exprs)
}

// FoldConstants simplifies *SQLCountFunctionExpr based on statically known constants.
func (f *SQLCountFunctionExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(f); err != nil {
		return nil, err
	}

	// Unlike the other aggregation functions, we do not want to return null
	// if the argument is null, count(NULL) returns 0 bizarrely.
	if hasNullExpr(f.exprs[0]) {
		return NewSQLValueExpr(values.NewSQLInt64(cfg.sqlValueKind, 0)), nil
	}
	return f, nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (f *SQLCountFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// ToAggregationLanguage generates aggregation language for the aggregation function.
func (f *SQLCountFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	transExpr, err := t.ToAggregationLanguage(f.exprs[0])
	if err != nil || transExpr == nil {
		return nil, err
	}
	if f.exprs[0] == NewSQLValueExpr(values.NewSQLVarchar(t.valueKind(), "")) {
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
	baseAggFunctionExpr
	Separator         option.String
	GroupConcatMaxLen int
}

// NewSQLGroupConcatFunctionExpr is a constructor for SQLGroupConcatFunctionExpr.
func NewSQLGroupConcatFunctionExpr(distinct bool, exprs []SQLExpr) *SQLGroupConcatFunctionExpr {
	return &SQLGroupConcatFunctionExpr{
		baseAggFunctionExpr: baseAggFunctionExpr{
			distinct: distinct,
			exprs:    exprs,
		},
		Separator:         option.NoneString(),
		GroupConcatMaxLen: 0,
	}
}

// FoldConstants simplifies *SQLGroupConcatFunctionExpr based on statically known constants.
func (f *SQLGroupConcatFunctionExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(f); err != nil {
		return nil, err
	}

	if hasNullExpr(f.exprs[0]) {
		return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
	}
	return f, nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
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
func (f *SQLGroupConcatFunctionExpr) EvalType() types.EvalType {
	return types.EvalString
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

// Evaluate does in memory computation for SQLGroupConcatFunctionExpr.
func (f *SQLGroupConcatFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(f)
	if err != nil {
		return nil, err
	}

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
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	return values.NewSQLVarchar(cfg.sqlValueKind, b.String()), nil
}

// Name returns name.
func (*SQLGroupConcatFunctionExpr) Name() string {
	return "group_concat"
}

// nolint: unparam
func (f *SQLGroupConcatFunctionExpr) reconcile() (SQLExpr, error) {
	return f, nil
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
	baseAggFunctionExpr
}

// NewSQLMaxFunctionExpr is a constructor for SQLMaxFunctionExpr.
func NewSQLMaxFunctionExpr(distinct bool, exprs []SQLExpr) *SQLMaxFunctionExpr {
	return &SQLMaxFunctionExpr{baseAggFunctionExpr{
		distinct: distinct,
		exprs:    exprs,
	}}
}

// EvalType for SQLMaxFunctionExpr returns the type of e.
func (f *SQLMaxFunctionExpr) EvalType() types.EvalType {
	return f.exprs[0].EvalType()
}

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLMaxFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLMaxFunctionExpr(%s)", f.Name())
}

// Evaluate for SQLMaxFunctionExpr does in memory computation for max.
func (f *SQLMaxFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(f)
	if err != nil {
		return nil, err
	}

	max := values.SQLValue(values.NewSQLNull(cfg.sqlValueKind))
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

			compared, err := values.CompareTo(max, eval, subSt.collation)
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

// nolint: unparam
func (f *SQLMaxFunctionExpr) reconcile() (SQLExpr, error) {
	return f, nil
}

// String converts to string.
func (f *SQLMaxFunctionExpr) String() string {
	return basicSQLAggFunctionToString(f.Name(), f.distinct, f.exprs)
}

// FoldConstants simplifies *SQLMaxFunctionExpr based on statically known constants.
func (f *SQLMaxFunctionExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(f); err != nil {
		return nil, err
	}

	if hasNullExpr(f.exprs[0]) {
		return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
	}
	return f, nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (f *SQLMaxFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// ToAggregationLanguage generates aggregation language for the aggregation function.
func (f *SQLMaxFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	transExpr, err := t.ToAggregationLanguage(f.exprs[0])
	if err != nil || transExpr == nil {
		return nil, err
	}

	return t.nullCheckAndWrapInAggOp(bsonutil.OpMax, transExpr), nil
}

// SQLMinFunctionExpr is a function that finds the minimal element.
type SQLMinFunctionExpr struct {
	baseAggFunctionExpr
}

// NewSQLMinFunctionExpr is a constructor for SQLMinFunctionExpr.
func NewSQLMinFunctionExpr(distinct bool, exprs []SQLExpr) *SQLMinFunctionExpr {
	return &SQLMinFunctionExpr{baseAggFunctionExpr{
		distinct: distinct,
		exprs:    exprs,
	}}
}

// EvalType for SQLMinFunctionExpr returns the type of e.
func (f *SQLMinFunctionExpr) EvalType() types.EvalType {
	return f.exprs[0].EvalType()
}

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLMinFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLMinFunctionExpr(%s)", f.Name())
}

// Evaluate for SQLMinFunctionExpr computes the minimal element in memory.
func (f *SQLMinFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(f)
	if err != nil {
		return nil, err
	}

	min := values.SQLValue(values.NewSQLNull(cfg.sqlValueKind))
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

			compared, err := values.CompareTo(min, eval, subSt.collation)
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

// nolint: unparam
func (f *SQLMinFunctionExpr) reconcile() (SQLExpr, error) {
	return f, nil
}

// String converts to string.
func (f *SQLMinFunctionExpr) String() string {
	return basicSQLAggFunctionToString(f.Name(), f.distinct, f.exprs)
}

// FoldConstants simplifies *SQLMinFunctionExpr based on statically known constants.
func (f *SQLMinFunctionExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(f); err != nil {
		return nil, err
	}

	if hasNullExpr(f.exprs[0]) {
		return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
	}
	return f, nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (f *SQLMinFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// ToAggregationLanguage generates aggregation language for the aggregation function.
func (f *SQLMinFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	transExpr, err := t.ToAggregationLanguage(f.exprs[0])
	if err != nil || transExpr == nil {
		return nil, err
	}

	return t.nullCheckAndWrapInAggOp(bsonutil.OpMin, transExpr), nil
}

// SQLSumFunctionExpr computes the summation of elements.
type SQLSumFunctionExpr struct {
	baseAggFunctionExpr
}

// NewSQLSumFunctionExpr is a constructor for SQLSumFunctionExpr.
func NewSQLSumFunctionExpr(distinct bool, exprs []SQLExpr) *SQLSumFunctionExpr {
	return &SQLSumFunctionExpr{baseAggFunctionExpr{
		distinct: distinct,
		exprs:    exprs,
	}}
}

// EvalType for SQLSumFunctionExpr is a standard floating point aggregation.
func (f *SQLSumFunctionExpr) EvalType() types.EvalType {
	return floatingPointAggregationFunctionEvalType(f.exprs[0].EvalType())
}

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLSumFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLSumFunctionExpr(%s)", f.Name())
}

// Evaluate for SQLSumFunctionExpr computes summations in memory.
func (f *SQLSumFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(f)
	if err != nil {
		return nil, err
	}

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
			if isDecimal || evalType == types.EvalDecimal128 {
				isDecimal = true
				sum = sum.Add(values.Decimal(eval))
				continue
			}

			floatEval := values.Float64(eval)

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
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	floatSum += correction

	if isDecimal {
		sum = sum.Add(decimal.NewFromFloat(floatSum))
		return values.NewSQLDecimal128(cfg.sqlValueKind, sum), nil
	}

	return values.NewSQLFloat(cfg.sqlValueKind, floatSum), nil
}

// Name returns name.
func (*SQLSumFunctionExpr) Name() string {
	return "sum"
}

// nolint: unparam
func (f *SQLSumFunctionExpr) reconcile() (SQLExpr, error) {
	return f, nil
}

// String converts to string.
func (f *SQLSumFunctionExpr) String() string {
	return basicSQLAggFunctionToString(f.Name(), f.distinct, f.exprs)
}

// FoldConstants simplifies *SQLSumFunctionExpr based on statically known constants.
func (f *SQLSumFunctionExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(f); err != nil {
		return nil, err
	}

	if hasNullExpr(f.exprs[0]) {
		return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
	}
	return f, nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (f *SQLSumFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// ToAggregationLanguage generates aggregation language for the aggregation function.
func (f *SQLSumFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	// We cannot sum date types.
	dataType := f.exprs[0].EvalType()
	if dataType == types.EvalDatetime || dataType == types.EvalDate {
		return nil, newPushdownFailure(f.Name(), fmt.Sprintf("cannot pushdown sum for date types"))
	}
	transExpr, err := t.ToAggregationLanguage(f.exprs[0])
	if err != nil || transExpr == nil {
		return nil, err
	}
	return t.nullCheckAndWrapInAggOp(bsonutil.OpSum, transExpr), nil
}

// SQLStdDevFunctionExpr computes a normal standard distribution for a population.
type SQLStdDevFunctionExpr struct {
	baseAggFunctionExpr
	// StdDev has multiple names and we want to recover the one actually used
	// for display purposes.
	name string
}

// NewSQLStdDevFunctionExpr is a constructor for SQLStdDevFunctionExpr.
func NewSQLStdDevFunctionExpr(name string, distinct bool, exprs []SQLExpr) *SQLStdDevFunctionExpr {
	return &SQLStdDevFunctionExpr{
		baseAggFunctionExpr: baseAggFunctionExpr{
			distinct: distinct,
			exprs:    exprs,
		}, name: name,
	}
}

// EvalType returns the type of the value this aggregate expression evaluates to.
func (f *SQLStdDevFunctionExpr) EvalType() types.EvalType {
	return floatingPointAggregationFunctionEvalType(f.exprs[0].EvalType())
}

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLStdDevFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLStdDevFunctionExpr(%s)", f.Name())
}

// Evaluate for SQLStdDevFunctionExpr computes the standard deviation of a population
// in memory.
func (f *SQLStdDevFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(f)
	if err != nil {
		return nil, err
	}

	var distinctMap map[interface{}]bool
	if f.distinct {
		distinctMap = make(map[interface{}]bool)
	}
	var data []values.SQLValue
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

			if isDecimal || eval.EvalType() == types.EvalDecimal128 {
				isDecimal = true
				sum = sum.Add(values.Decimal(eval))
				continue
			}

			floatEval := values.Float64(eval)

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
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	floatSum += correction

	if isDecimal {
		diff := decimal.Zero
		sum = sum.Add(decimal.NewFromFloat(floatSum))
		avg := sum.Div(decimal.NewFromFloat(count))

		for _, v := range data {
			val := values.Decimal(v).Sub(avg)
			diff = diff.Add(val.Mul(val))
		}

		diff = diff.Div(decimal.NewFromFloat(count))
		f, _ := diff.Float64()
		return values.NewSQLDecimal128(cfg.sqlValueKind, decimal.NewFromFloat(math.Sqrt(f))), nil
	}

	avg := floatSum / count
	diff := 0.0

	for _, val := range data {
		diff += math.Pow(values.Float64(val)-avg, 2)
	}

	return values.NewSQLFloat(cfg.sqlValueKind, math.Sqrt(diff/count)), nil
}

// Name returns name.
func (f *SQLStdDevFunctionExpr) Name() string {
	return f.name
}

// nolint: unparam
func (f *SQLStdDevFunctionExpr) reconcile() (SQLExpr, error) {
	return f, nil
}

// String converts to string.
func (f *SQLStdDevFunctionExpr) String() string {
	return basicSQLAggFunctionToString(f.name, f.distinct, f.exprs)
}

// FoldConstants simplifies *SQLStdDevFunctionExpr based on statically known constants.
func (f *SQLStdDevFunctionExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(f); err != nil {
		return nil, err
	}

	if hasNullExpr(f.exprs[0]) {
		return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
	}
	return f, nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (f *SQLStdDevFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// ToAggregationLanguage generates aggregation language for the aggregation function.
func (f *SQLStdDevFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	// We cannot stddev date types.
	dataType := f.exprs[0].EvalType()
	if dataType == types.EvalDatetime || dataType == types.EvalDate {
		return nil, newPushdownFailure(f.Name(), fmt.Sprintf("cannot pushdown std for date types"))
	}
	transExpr, err := t.ToAggregationLanguage(f.exprs[0])
	if err != nil || transExpr == nil {
		return nil, err
	}
	return t.nullCheckAndWrapInAggOp(bsonutil.OpStdDevPop, transExpr), nil
}

// SQLStdDevSampleFunctionExpr computes standard deviation of a sample.
type SQLStdDevSampleFunctionExpr struct {
	baseAggFunctionExpr
}

// NewSQLStdDevSampleFunctionExpr is a constructor for SQLStdDevSampleFunctionExpr.
func NewSQLStdDevSampleFunctionExpr(distinct bool, exprs []SQLExpr) *SQLStdDevSampleFunctionExpr {
	return &SQLStdDevSampleFunctionExpr{baseAggFunctionExpr{
		distinct: distinct,
		exprs:    exprs,
	}}
}

// EvalType returns the type of the value this aggregate expression evaluates to.
func (f *SQLStdDevSampleFunctionExpr) EvalType() types.EvalType {
	return floatingPointAggregationFunctionEvalType(f.exprs[0].EvalType())
}

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLStdDevSampleFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLStdDevSampleFunctionExpr(%s)", f.Name())
}

// Evaluate for SQLStdDevSampleFunctionExpr computes standard deviation for
// a sample in memory.
func (f *SQLStdDevSampleFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (values.SQLValue, error) {
	err := validateArgs(f)
	if err != nil {
		return nil, err
	}

	var data []values.SQLValue
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

			if isDecimal || eval.EvalType() == types.EvalDecimal128 {
				isDecimal = true
				sum = sum.Add(values.Decimal(eval))
				continue
			}

			floatEval := values.Float64(eval)

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
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}

	floatSum += correction

	if isDecimal {
		diff := decimal.Zero
		sum = sum.Add(decimal.NewFromFloat(floatSum))
		avg := sum.Div(decimal.NewFromFloat(count))

		for _, v := range data {
			val := values.Decimal(v).Sub(avg)
			diff = diff.Add(val.Mul(val))
		}

		if count == 1 {
			return values.NewSQLNull(cfg.sqlValueKind), nil
		}
		diff = diff.Div(decimal.NewFromFloat(count - 1))
		f, _ := diff.Float64()
		return values.NewSQLDecimal128(cfg.sqlValueKind, decimal.NewFromFloat(math.Sqrt(f))), nil
	}

	avg := floatSum / count
	diff := 0.0

	for _, val := range data {
		diff += math.Pow(values.Float64(val)-avg, 2)
	}

	if count == 1 {
		return values.NewSQLNull(cfg.sqlValueKind), nil
	}
	return values.NewSQLFloat(cfg.sqlValueKind, math.Sqrt(diff/(count-1))), nil
}

// Name returns name.
func (*SQLStdDevSampleFunctionExpr) Name() string {
	return "stddev_samp"
}

// nolint: unparam
func (f *SQLStdDevSampleFunctionExpr) reconcile() (SQLExpr, error) {
	return f, nil
}

// String converts to string.
func (f *SQLStdDevSampleFunctionExpr) String() string {
	return basicSQLAggFunctionToString(f.Name(), f.distinct, f.exprs)
}

// FoldConstants simplifies *SQLStdDevSampleFunctionExpr based on statically known constants.
func (f *SQLStdDevSampleFunctionExpr) FoldConstants(cfg *OptimizerConfig) (SQLExpr, error) {
	if err := validateArgs(f); err != nil {
		return nil, err
	}

	if hasNullExpr(f.exprs[0]) {
		return NewSQLValueExpr(values.NewSQLNull(cfg.sqlValueKind)), nil
	}
	return f, nil
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate directly in a $match stage via $expr.
func (f *SQLStdDevSampleFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// ToAggregationLanguage generates aggregation language for the aggregation function.
func (f *SQLStdDevSampleFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	// We cannot stddev date types.
	dataType := f.exprs[0].EvalType()
	if dataType == types.EvalDatetime || dataType == types.EvalDate {
		return nil, newPushdownFailure(f.Name(), fmt.Sprintf("cannot pushdown stddev_samp for date types"))
	}
	transExpr, err := t.ToAggregationLanguage(f.exprs[0])
	if err != nil || transExpr == nil {
		return nil, err
	}
	return t.nullCheckAndWrapInAggOp(bsonutil.OpStdDevSamp, transExpr), nil
}
