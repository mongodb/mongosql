package evaluator

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/mathutil"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/schema"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/parser"
)

const (
	shortTimeFormat      = "2006-01-02"
	incorrectArgCountMsg = "incorrect number of arguments"
)

const (
	// This corresponds to 9999-12-31, which is the max
	// date MySQL supports internally. Interestingly,
	// it will only allow TO_DAYS('9999-12-31') as the max
	// value because of date parser limitations, but
	// TO_DAYS(FROM_DAYS(3652499)) actually works, returning
	// 3652499 as expected. However, our Evaluate implementation
	// cannot handle a number greater than this, so we cap the pushdown,
	// too.
	maxFromDays = 3652499
)

// These constants are strings used as date/time units in some scalar functions.
const (
	Year              = "year"
	Quarter           = "quarter"
	Month             = "month"
	Week              = "week"
	Day               = "day"
	Hour              = "hour"
	Minute            = "minute"
	Second            = "second"
	Microsecond       = "microsecond"
	YearMonth         = "year_month"
	DayHour           = "day_hour"
	DayMinute         = "day_minute"
	DaySecond         = "day_second"
	DayMicrosecond    = "day_microsecond"
	HourMinute        = "hour_minute"
	HourSecond        = "hour_second"
	HourMicrosecond   = "hour_microsecond"
	MinuteSecond      = "minute_second"
	MinuteMicrosecond = "minute_microsecond"
	SecondMicrosecond = "second_microsecond"
)

const (
	// millisecondsPerDay is the number of milliseconds in a day.
	millisecondsPerDay = 8.64e+7
	// secondsPerDay is the number of seconds in a day.
	secondsPerDay = 8.64e+4
	// secondsPerHour is the number of seconds in an hour.
	secondsPerHour = 3600.0
	// secondsPerMinute is the number of seconds in an minute.
	secondsPerMinute = 60.0
)

var toMilliseconds = map[string]float64{
	Week:        millisecondsPerDay * 7,
	Day:         millisecondsPerDay,
	Hour:        3.6e6,
	Minute:      6e4,
	Second:      1e3,
	Microsecond: 1e-3,
}

var (
	zeroDate, _ = time.ParseInLocation(shortTimeFormat, "0000-00-00", schema.DefaultLocale)
)

var stringToNum = map[string]int{
	"0": 0,
	"1": 1,
	"2": 2,
	"3": 3,
	"4": 4,
	"5": 5,
	"6": 6,
	"7": 7,
	"8": 8,
	"9": 9,
	"A": 10, "a": 10,
	"B": 11, "b": 11,
	"C": 12, "c": 12,
	"D": 13, "d": 13,
	"E": 14, "e": 14,
	"F": 15, "f": 15,
	"G": 16, "g": 16,
	"H": 17, "h": 17,
	"I": 18, "i": 18,
	"J": 19, "j": 19,
	"K": 20, "k": 20,
	"L": 21, "l": 21,
	"M": 22, "m": 22,
	"N": 23, "n": 23,
	"O": 24, "o": 24,
	"P": 25, "p": 25,
	"Q": 26, "q": 26,
	"R": 27, "r": 27,
	"S": 28, "s": 28,
	"T": 29, "t": 29,
	"U": 30, "u": 30,
	"V": 31, "v": 31,
	"W": 32, "w": 32,
	"X": 33, "x": 33,
	"Y": 34, "y": 34,
	"Z": 35, "z": 35,
}

var validNumbers = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "a",
	"B", "b", "C", "c", "D", "d", "E", "e", "F", "f", "G", "g", "H", "h", "I", "i", "J", "j",
	"K", "k", "L", "l", "M", "m", "N", "n", "O", "o", "P", "p", "Q", "q", "R", "r", "S", "s", "T",
	"t", "U", "u", "V", "v", "W", "w", "X", "x", "Y", "y", "Z", "z"}

var numToString = map[int]string{
	0:  "0",
	1:  "1",
	2:  "2",
	3:  "3",
	4:  "4",
	5:  "5",
	6:  "6",
	7:  "7",
	8:  "8",
	9:  "9",
	10: "A",
	11: "B",
	12: "C",
	13: "D",
	14: "E",
	15: "F",
	16: "G",
	17: "H",
	18: "I",
	19: "J",
	20: "K",
	21: "L",
	22: "M",
	23: "N",
	24: "O",
	25: "P",
	26: "Q",
	27: "R",
	28: "S",
	29: "T",
	30: "U",
	31: "V",
	32: "W",
	33: "X",
	34: "Y",
	35: "Z",
}

var scalarFuncMap = map[string]scalarFunc{
	"abs":     &absFunc{singleArgFloatMathFunc(math.Abs)},
	"acos":    &acosFunc{singleArgFloatMathFunc(math.Acos)},
	"adddate": &dateArithmeticFunc{&addDateFunc{}, false},
	"asin":    &asinFunc{singleArgFloatMathFunc(math.Asin)},
	"atan": multiArgFloatMathFunc{
		single: singleArgFloatMathFunc(math.Atan),
		dual:   dualArgFloatMathFunc(math.Atan2),
	},
	"atan2":             dualArgFloatMathFunc(math.Atan2),
	"ascii":             &asciiFunc{},
	"cast":              &convertFunc{},
	"ceil":              &ceilFunc{singleArgFloatMathFunc(math.Ceil)},
	"ceiling":           &ceilFunc{singleArgFloatMathFunc(math.Ceil)},
	"char":              &charFunc{},
	"char_length":       &characterLengthFunc{},
	"character_length":  &characterLengthFunc{},
	"coalesce":          &coalesceFunc{},
	"concat":            &concatFunc{},
	"concat_ws":         &concatWsFunc{},
	"connection_id":     &connectionIDFunc{},
	"convert":           &convertFunc{},
	"conv":              &convFunc{},
	"cos":               &cosFunc{singleArgFloatMathFunc(math.Cos)},
	"cot":               &cotFunc{},
	"curdate":           &currentDateFunc{},
	"current_date":      &currentDateFunc{},
	"current_timestamp": &currentTimestampFunc{},
	"current_user":      &userFunc{},
	"curtime":           &curtimeFunc{},
	"database":          &dbFunc{},
	"date":              &dateFunc{},
	"date_add":          &dateArithmeticFunc{&dateAddFunc{}, false},
	"datediff":          &dateDiffFunc{},
	"date_sub":          &dateArithmeticFunc{&dateSubFunc{}, true},
	"date_format":       &dateFormatFunc{},
	"day":               &dayOfMonthFunc{},
	"dayname":           &dayNameFunc{},
	"dayofmonth":        &dayOfMonthFunc{},
	"dayofweek":         &dayOfWeekFunc{},
	"dayofyear":         &dayOfYearFunc{},
	"degrees": &degreesFunc{singleArgFloatMathFunc(
		func(f float64) float64 {
			return f * 180 / math.Pi
		},
	)},
	"elt":           &eltFunc{},
	"exp":           &expFunc{singleArgFloatMathFunc(math.Exp)},
	"extract":       &extractFunc{},
	"field":         &fieldFunc{},
	"floor":         &floorFunc{singleArgFloatMathFunc(math.Floor)},
	"from_days":     &fromDaysFunc{},
	"from_unixtime": &fromUnixtimeFunc{},
	"greatest":      &greatestFunc{},
	"hour":          &hourFunc{},
	"if":            &ifFunc{},
	"ifnull":        &ifnullFunc{},
	"insert":        &insertFunc{},
	"instr":         &instrFunc{},
	"interval":      &intervalFunc{},
	"isnull":        &isnullFunc{},
	"last_day":      &lastDayFunc{},
	"lcase":         &lcaseFunc{},
	"least":         &leastFunc{},
	"left":          &leftFunc{},
	"length":        &lengthFunc{},
	"ln":            singleArgFloatMathFunc(math.Log),
	"locate":        &locateFunc{},
	// Use 0 for ln and logs where base is passed as first arg
	"log":         &logFunc{0},
	"log2":        &logFunc{2},
	"log10":       &logFunc{10},
	"lower":       &lcaseFunc{},
	"lpad":        &padFunc{true, &lpadFunc{}},
	"ltrim":       &ltrimFunc{},
	"makedate":    &makeDateFunc{},
	"md5":         &md5Func{},
	"microsecond": &microsecondFunc{},
	"mid":         &substringFunc{true},
	"minute":      &minuteFunc{},
	"mod":         &modFunc{dualArgFloatMathFunc(math.Mod)},
	"month":       &monthFunc{},
	"monthname":   &monthNameFunc{},
	"not":         &notFunc{},
	"now":         &currentTimestampFunc{},
	"nopushdown":  &nopushdownFunc{},
	"nullif":      &nullifFunc{},
	"pi":          &constantFunc{math.Pi, EvalDouble},
	"pow":         &powFunc{},
	"power":       &powFunc{},
	"quarter":     &quarterFunc{},
	"rand":        &randFunc{},
	"radians": &radiansFunc{singleArgFloatMathFunc(
		func(f float64) float64 {
			return f * math.Pi / 180
		},
	)},
	"repeat":          &repeatFunc{},
	"replace":         &replaceFunc{},
	"reverse":         &reverseFunc{},
	"right":           &rightFunc{},
	"round":           &roundFunc{},
	"rpad":            &padFunc{false, &rpadFunc{}},
	"rtrim":           &rtrimFunc{},
	"schema":          &dbFunc{},
	"second":          &secondFunc{},
	"session_user":    &userFunc{},
	"sign":            &signFunc{},
	"sin":             &sinFunc{singleArgFloatMathFunc(math.Sin)},
	"sleep":           &sleepFunc{},
	"sqrt":            &sqrtFunc{singleArgFloatMathFunc(math.Sqrt)},
	"space":           &spaceFunc{},
	"str_to_date":     &strToDateFunc{},
	"subdate":         &dateArithmeticFunc{&subDateFunc{}, true},
	"substr":          &substringFunc{},
	"substring":       &substringFunc{},
	"substring_index": &substringIndexFunc{},
	"system_user":     &userFunc{},
	"tan":             &tanFunc{singleArgFloatMathFunc(math.Tan)},
	"timediff":        &timeDiffFunc{},
	"timestamp":       &timestampFunc{},
	"timestampadd":    &timestampAddFunc{},
	"timestampdiff":   &timestampDiffFunc{},
	"time_to_sec":     &timeToSecFunc{},
	"to_days":         &toDaysFunc{},
	"to_seconds":      &toSecondsFunc{},
	"trim":            &trimFunc{},
	"truncate":        &truncateFunc{},
	"ucase":           &ucaseFunc{},
	"unix_timestamp":  &unixTimestampFunc{},
	"upper":           &ucaseFunc{},
	"user":            &userFunc{},
	"utc_date":        &utcDateFunc{},
	"utc_time":        &curtimeFunc{},
	"utc_timestamp":   &utcTimestampFunc{},
	"version":         &versionFunc{},
	"week":            &weekFunc{},
	"weekday":         &weekdayFunc{},
	"year":            &yearFunc{},
	"yearweek":        &yearWeekFunc{},
}

// normalizingScalarFunc is an interface for a Scalar Function that
// can be normalized in some way.
type normalizingScalarFunc interface {
	Normalize(SQLValueKind, *SQLScalarFunctionExpr) SQLExpr
}

// reconcilingScalarFunc is an interface for a Scalar Function
// that can be type reconciled.
type reconcilingScalarFunc interface {
	Reconcile(*SQLScalarFunctionExpr) *SQLScalarFunctionExpr
}

// scalarFunc represents a SQL scalar function.
type scalarFunc interface {
	// Evaluate evaluates the scalar function.
	Evaluate(context.Context, *ExecutionConfig, *ExecutionState, []SQLValue) (SQLValue, error)
	// Validate validates that the number of arguments passed to the scalar function
	// is correct.
	Validate(exprCount int) error
	// EvalType returns the EvalType return type of the scalar function.
	EvalType([]SQLExpr) EvalType
	// FuncName returns the name of the scalar function
	FuncName() string
}

// translatableToAggregationScalarFunc is an interface for a Scalar Function
// that can be translated to MongoDB Aggregation Language.
type translatableToAggregationScalarFunc interface {
	FuncToAggregationLanguage(*PushdownTranslator, []SQLExpr) (interface{}, PushdownFailure)
}

//
// SQLScalarFunctionExpr represents a scalar function.
//
type SQLScalarFunctionExpr struct {
	Name  string
	Func  scalarFunc
	Exprs []SQLExpr
}

var _ translatableToAggregation = (*SQLScalarFunctionExpr)(nil)

// ExprName returns a string representing this SQLExpr's name.
func (f *SQLScalarFunctionExpr) ExprName() string {
	return fmt.Sprintf("SQLScalarFunctionExpr(%s)", f.Func.FuncName())
}

// Evaluate evaluates a SQLScalarFunctionExpr to a SQLValue.
func (f *SQLScalarFunctionExpr) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (SQLValue, error) {
	err := f.Func.Validate(len(f.Exprs))
	if err != nil {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), fmt.Errorf("%v '%v'", err.Error(), f.Name)
	}

	values, err := evaluateArgs(ctx, cfg, st, f.Exprs)
	if err != nil {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType()), err
	}
	return f.Func.Evaluate(ctx, cfg, st, values)
}

// Normalize will attempt to change SQLScalarFunctionExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (f *SQLScalarFunctionExpr) Normalize(kind SQLValueKind) Node {
	if nsf, ok := f.Func.(normalizingScalarFunc); ok {
		return nsf.Normalize(kind, f)
	}

	return f
}

// Reconcile will ensure all types are compatible within a SQLScalarFunction.
// If the types are not compatible, it will wrap it in a SQLConvertExpr.
func (f *SQLScalarFunctionExpr) Reconcile() *SQLScalarFunctionExpr {
	if rsf, ok := f.Func.(reconcilingScalarFunc); ok {
		return rsf.Reconcile(f)
	}

	return f
}

// SkipConstantFolding will check if the SQLScalarFunctionExpr requires an evaluation context.
func (f *SQLScalarFunctionExpr) SkipConstantFolding() bool {
	if r, ok := f.Func.(SkipConstantFolding); ok {
		return r.SkipConstantFolding()
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

// ToAggregationLanguage translates SQLScalarFunctionExpr into something that can
// be used in an aggregation pipeline. If SQLScalarFunctionExpr cannot be translated,
// it will return nil and error.
func (f *SQLScalarFunctionExpr) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	if fun, ok := f.Func.(translatableToAggregationScalarFunc); ok {
		res, err := fun.FuncToAggregationLanguage(t, f.Exprs)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	t.Cfg.lg.Debugf(log.Dev,
		"%q cannot be pushed down as an aggregate expression at this time",
		f.Name)

	return nil, newPushdownFailure(
		fmt.Sprintf("SQLScalarFunctionExpr(%s)", f.Name),
		"no FuncToAggregationLanguage implementation",
	)
}

// ToAggregationPredicate translates this expression to the aggregation language
// to be evaluated as a predicate in a $match stage via $expr.
func (f *SQLScalarFunctionExpr) ToAggregationPredicate(t *PushdownTranslator) (interface{}, PushdownFailure) {
	return f.ToAggregationLanguage(t)
}

// EvalType returns the EvalType associated with the SQLScalarFunctionExpr.
func (f *SQLScalarFunctionExpr) EvalType() EvalType {
	return f.Func.EvalType(f.Exprs)
}

type absFunc struct {
	singleArgFloatMathFunc
}

func (*absFunc) FuncName() string {
	return "abs"
}

var _ translatableToAggregationScalarFunc = (*absFunc)(nil)

func (*absFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(abs)",
			fmt.Sprintf("expected 1 arguments, found %d", len(exprs)),
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.NewM(bsonutil.NewDocElem("$abs", args[0])), nil
}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_acos
type acosFunc struct {
	singleArgFloatMathFunc
}

func (*acosFunc) FuncName() string {
	return "acos"
}

var _ translatableToAggregationScalarFunc = (*acosFunc)(nil)

func (*acosFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(acos)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	input := "$$input"
	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("input", args[0]),
	)

	// MySQL returns NULL for values outside of the range [-1,1].
	// asin x + acos x = pi/2
	return bsonutil.WrapInLet(letAssignment,
		bsonutil.WrapInCond(nil,
			bsonutil.WrapInAcosComputation(input),
			bsonutil.WrapInOp(bsonutil.OpLt, input, -1.0),
			bsonutil.WrapInOp(bsonutil.OpGt, input, 1.0),
		),
	), nil
}

type addDateFunc struct{}

func (*addDateFunc) FuncName() string {
	return "addDate"
}

var _ normalizingScalarFunc = (*addDateFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_adddate
func (*addDateFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	adder := &dateAddFunc{}
	return adder.Evaluate(ctx, cfg, st, values)
}

func (*addDateFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}
	return f
}

func (*addDateFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDatetime
}

func (*addDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type asciiFunc struct{}

func (*asciiFunc) FuncName() string {
	return "ascii"
}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_ascii
func (f *asciiFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, EvalInt64), nil
	}

	str := values[0].String()
	if str == "" {
		return NewSQLInt64(cfg.sqlValueKind, 0), nil
	}

	c := str[0]

	return NewSQLInt64(cfg.sqlValueKind, int64(c)), nil
}

func (*asciiFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*asciiFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_asin
type asinFunc struct {
	singleArgFloatMathFunc
}

func (*asinFunc) FuncName() string {
	return "asin"
}

var _ translatableToAggregationScalarFunc = (*asinFunc)(nil)

func (*asinFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(asin)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	input := "$$input"
	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("input", args[0]),
	)

	// MySQL returns NULL for values outside of the range [-1,1].
	// asin(x) =  pi/2 - cos(x) via the identity:
	// asin(x) + acos(x) = pi/2.
	return bsonutil.WrapInLet(letAssignment,
		bsonutil.WrapInCond(nil,
			bsonutil.WrapInOp(bsonutil.OpSubtract, math.Pi/2.0, bsonutil.WrapInAcosComputation(input)),
			bsonutil.WrapInOp(bsonutil.OpLt, input, -1.0),
			bsonutil.WrapInOp(bsonutil.OpGt, input, 1.0),
		),
	), nil
}

type ceilFunc struct {
	singleArgFloatMathFunc
}

func (*ceilFunc) FuncName() string {
	return "ceil"
}

var _ translatableToAggregationScalarFunc = (*ceilFunc)(nil)

func (*ceilFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(ceil)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCeil, args[0])), nil
}

type charFunc struct{}

func (*charFunc) FuncName() string {
	return "char"
}

func (*charFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {

	var b []byte
	for _, i := range values {
		if i.IsNull() {
			continue
		}
		v := Int64(i)
		if v >= 256 {
			var temp []byte
			num := v / 255
			v = v % 256
			for num >= 256 {
				temp = append(temp, 0)
				num /= 255
			}
			b = append(b, uint8(num))
			b = append(b, temp...)
		}
		b = append(b, uint8(v))
	}

	return NewSQLVarchar(cfg.sqlValueKind, string(b)), nil
}

func (*charFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*charFunc) Validate(exprCount int) error {
	if exprCount == 0 {
		return errIncorrectCount
	}

	return nil
}

type characterLengthFunc struct{}

func (*characterLengthFunc) FuncName() string {
	return "characterLength"
}

var _ reconcilingScalarFunc = (*characterLengthFunc)(nil)
var _ translatableToAggregationScalarFunc = (*characterLengthFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_char_length
func (f *characterLengthFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, EvalInt64), nil
	}

	value := []rune(values[0].String())

	return NewSQLInt64(cfg.sqlValueKind, int64(len(value))), nil
}

func (*characterLengthFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(characterLength)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(characterLength)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$strLenCP", args[0]), nil
}

func (*characterLengthFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalString)
}

func (*characterLengthFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*characterLengthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type coalesceFunc struct{}

func (*coalesceFunc) FuncName() string {
	return "coalesce"
}

var _ reconcilingScalarFunc = (*coalesceFunc)(nil)
var _ translatableToAggregationScalarFunc = (*coalesceFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_coalesce
func (f *coalesceFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	for _, value := range values {
		if !value.IsNull() {
			return value, nil
		}
	}

	return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
}

func (*coalesceFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	var coalesce func([]interface{}) interface{}
	coalesce = func(args []interface{}) interface{} {
		if len(args) == 0 {
			return nil
		}
		replacement := coalesce(args[1:])
		return bsonutil.WrapInIfNull(args[0], replacement)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return coalesce(args), nil
}

func (*coalesceFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return f
}

func (*coalesceFunc) EvalType(exprs []SQLExpr) EvalType {
	sorter := &EvalTypeSorter{VarcharHighPriority: true}
	return preferentialTypeWithSorter(sorter, exprs...)
}

func (*coalesceFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, -1)
}

type concatFunc struct{}

func (*concatFunc) FuncName() string {
	return "concat"
}

var _ normalizingScalarFunc = (*concatFunc)(nil)
var _ reconcilingScalarFunc = (*concatFunc)(nil)
var _ translatableToAggregationScalarFunc = (*concatFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_concat
func (f *concatFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (v SQLValue, err error) {
	if hasNullValue(values...) {
		v = NewSQLNull(cfg.sqlValueKind, EvalString)
		err = nil
		return
	}

	defer func() {
		if r := recover(); r != nil {
			v = nil
			err = fmt.Errorf("%v", r)
		}
	}()

	var b bytes.Buffer
	for _, value := range values {
		b.WriteString(value.String())
	}

	v = NewSQLVarchar(cfg.sqlValueKind, b.String())
	err = nil
	return
}

func (*concatFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) < 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(concat)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpConcat, args)), nil
}

func (*concatFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*concatFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalString)
}

func (*concatFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*concatFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, -1)
}

type concatWsFunc struct{}

func (*concatWsFunc) FuncName() string {
	return "concatWs"
}

var _ normalizingScalarFunc = (*concatWsFunc)(nil)
var _ reconcilingScalarFunc = (*concatWsFunc)(nil)
var _ translatableToAggregationScalarFunc = (*concatWsFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_concat-ws
func (f *concatWsFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (v SQLValue, err error) {
	if values[0].IsNull() {
		v = NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values)))
		return
	}

	defer func() {
		if r := recover(); r != nil {
			v = nil
			err = fmt.Errorf("%v", r)
		}
	}()

	var b bytes.Buffer
	separator := values[0].String()
	trimValues := values[1:]
	for i, value := range trimValues {
		if value.IsNull() {
			continue
		}
		b.WriteString(value.String())
		if i != len(trimValues)-1 {
			b.WriteString(separator)
		}
	}

	v = NewSQLVarchar(cfg.sqlValueKind, b.String())
	return
}

func (*concatWsFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) < 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(concatWs)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	var pushArgs []interface{}

	for _, value := range args[1:] {
		pushArgs = append(pushArgs,
			bsonutil.WrapInNullCheckedCond(bsonutil.WrapInLiteral(""), value, value),
			bsonutil.WrapInNullCheckedCond(bsonutil.WrapInLiteral(""), args[0], value),
		)
	}

	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpConcat, pushArgs[:len(pushArgs)-1])), nil
}

func (*concatWsFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if len(f.Exprs) >= 2 {
		firstVal, ok := f.Exprs[0].(SQLValue)
		if ok && firstVal.IsNull() {
			return NewSQLNull(kind, f.EvalType())
		}
	}

	return f
}

func (*concatWsFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalString)
}

func (*concatWsFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*concatWsFunc) Validate(exprCount int) error {
	if ensureArgCount(exprCount, -1) != nil || exprCount < 2 {
		return errIncorrectCount
	}
	return nil
}

type connectionIDFunc struct{}

func (*connectionIDFunc) FuncName() string {
	return "connectionID"
}

func (*connectionIDFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return NewSQLUint64(cfg.sqlValueKind, cfg.connID), nil
}

func (*connectionIDFunc) SkipConstantFolding() bool {
	return true
}

func (*connectionIDFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*connectionIDFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type constantFunc struct {
	value    interface{}
	evalType EvalType
}

func (*constantFunc) FuncName() string {
	return "constant"
}

func (c *constantFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	sqlVal := GoValueToSQLValue(cfg.sqlValueKind, c.value)
	if sqlVal.EvalType() != c.evalType {
		err := fmt.Errorf(
			"actual EvalType %x did not match declared EvalType %x",
			sqlVal.EvalType(), c.evalType,
		)
		return nil, err
	}
	return sqlVal, nil
}

func (c *constantFunc) EvalType(exprs []SQLExpr) EvalType {
	return c.evalType
}

func (*constantFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type convFunc struct{}

func (*convFunc) FuncName() string {
	return "conv"
}

var _ normalizingScalarFunc = (*convFunc)(nil)
var _ reconcilingScalarFunc = (*convFunc)(nil)
var _ translatableToAggregationScalarFunc = (*convFunc)(nil)

// https://dev.mysql.com/doc/refman/8.0/en/mathematical-functions.html#function_conv
// Diverges from MySQL behavior in its handling of negative values
// Converts bases to positive numbers, and returns a negative value if the input is negative
// MySQL claims that "If from_base is a negative number, N is regarded as a signed number.
// Otherwise, N is treated as unsigned." Manual testing shows that it returns the 2's
// complement version if the number is negative unless the to_base is also negative, in which
// case it returns the number with a negative sign at the front
func (f *convFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	num := values[0].String()
	originalBase := absInt64(Int64(values[1]))
	newBase := absInt64(Int64(values[2]))
	negative := false

	if baseIsInvalid(originalBase) || baseIsInvalid(newBase) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	if string(num[0]) == "-" {
		num = num[1:]
		negative = true
	}

	if strings.Contains(num, ".") {
		num = num[0:strings.Index(num, ".")]
	}

	base10Version, err := strconv.ParseInt(num, int(originalBase), 64)
	if err != nil {
		return NewSQLVarchar(cfg.sqlValueKind, "0"), nil
	}
	strVersion := strconv.FormatInt(base10Version, int(newBase))

	if negative && strVersion != "0" {
		strVersion = "-" + strVersion
	}

	return NewSQLVarchar(cfg.sqlValueKind, strings.ToUpper(strVersion)), nil
}

func (*convFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	if v, ok := f.Exprs[1].(SQLValue); ok {
		if baseIsInvalid(absInt64(Int64(v))) {
			return NewSQLNull(kind, f.EvalType())
		}
	}

	if v, ok := f.Exprs[2].(SQLValue); ok {
		if baseIsInvalid(absInt64(Int64(v))) {
			return NewSQLNull(kind, f.EvalType())
		}
	}

	return f
}

func baseIsInvalid(base int64) bool {
	if base < 2 || base > 36 {
		return true
	}

	return false
}

func (*convFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {

	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(conv)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}
	num := args[0]
	oldBase := args[1]
	newBase := args[2]

	// length is how long (in digits) the input number is
	normalizedVars := bsonutil.NewM(
		bsonutil.NewDocElem("originalBase", bsonutil.WrapInOp(bsonutil.OpAbs, oldBase)),
		bsonutil.NewDocElem("newBase", bsonutil.WrapInOp(bsonutil.OpAbs, newBase)),
		bsonutil.NewDocElem("negative", bsonutil.WrapInOp(bsonutil.OpEq, "-", bsonutil.WrapInOp(bsonutil.OpSubstr, num, 0, 1))),
		bsonutil.NewDocElem("nonNegativeNumber", bsonutil.WrapInCond(
			bsonutil.WrapInOp(bsonutil.OpSubstr, num, 1,
				bsonutil.WrapInOp(bsonutil.OpSubtract, bsonutil.WrapInOp(bsonutil.OpStrlenCP, num), 1)),
			num,
			bsonutil.WrapInOp(bsonutil.OpEq, "-", bsonutil.WrapInOp(bsonutil.OpSubstr, num, 0, 1)))),
	)

	indexOfDecimal := bsonutil.NewM(
		bsonutil.NewDocElem("decimalIndex", bsonutil.WrapInOp(bsonutil.OpIndexOfCP, "$$nonNegativeNumber", ".")),
	)

	eliminateDecimal := bsonutil.NewM(
		bsonutil.NewDocElem("number", bsonutil.WrapInCond("$$nonNegativeNumber",
			bsonutil.WrapInOp(bsonutil.OpSubstr, "$$nonNegativeNumber", 0, "$$decimalIndex"),
			bsonutil.WrapInOp(bsonutil.OpEq, "$$decimalIndex", -1))),
	)

	createLength := bsonutil.NewM(
		bsonutil.NewDocElem("length", bsonutil.WrapInOp(bsonutil.OpStrlenCP, "$$number")),
	)

	// indexArr is an array of numbers from 0 to n-1 when n = length
	createIndexArr := bsonutil.NewM(
		bsonutil.NewDocElem("indexArr", bsonutil.WrapInOp(bsonutil.OpRange, 0, "$$length", 1)),
	)

	// charArr breaks the number entered into an array of characters where each char is a digit
	createCharArr := bsonutil.NewM(
		bsonutil.NewDocElem("charArr", bsonutil.WrapInMap("$$indexArr", "this",
			bsonutil.NewArray("$$this", bsonutil.WrapInOp(bsonutil.OpSubstr, "$$number", "$$this", 1)))),
	)

	// This logic takes in the charArr and outputs a 2D array containing the index and the
	// base10 numerical value of the character.
	// i.e. if charArr = ["3", "A", "2"], numArr = [[0, 3], [1, 10], [2, 2]]
	branches1 := bsonutil.NewMArray()
	for _, k := range validNumbers {
		branches1 = append(branches1,
			bsonutil.WrapInCase(
				bsonutil.WrapInOp(bsonutil.OpEq,
					bsonutil.WrapInOp(bsonutil.OpArrElemAt, "$$this", 1),
					k,
				),
				bsonutil.NewArray(
					bsonutil.WrapInOp(bsonutil.OpArrElemAt, "$$this", 0),
					stringToNum[k],
				),
			),
		)
	}
	createNumArr := bsonutil.NewM(
		bsonutil.NewDocElem("numArr", bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpMap, bsonutil.NewM(
				bsonutil.NewDocElem("input", "$$charArr"),
				bsonutil.NewDocElem("in", bsonutil.WrapInSwitch(bsonutil.NewArray(0, 100), branches1...)),
			)),
		)),
	)

	// invalidArr has False for every digit that is valid, and True for every digit that is invalid
	// In order for the input string to be converted to a new number base every entry in this
	// array must be False.
	createInvalidArr := bsonutil.NewM(
		bsonutil.NewDocElem("invalidArr", bsonutil.WrapInMap(
			"$$numArr",
			"this",
			bsonutil.WrapInOp(bsonutil.OpGte, bsonutil.WrapInOp(bsonutil.OpArrElemAt, "$$this", 1), "$$originalBase"),
		)),
	)

	// Given a charArr = [[1, x1]...[i, xi]...[n, xn]] and a base b,
	// This implements the logic: sum(b^(n-i-1) * xi) with i = 0->n-1
	generateBase10 := bsonutil.NewM(
		bsonutil.NewDocElem("base10", bsonutil.WrapInOp(bsonutil.OpSum,
			bsonutil.WrapInMap("$$numArr", "this",
				bsonutil.WrapInOp(bsonutil.OpMultiply,
					bsonutil.WrapInOp(bsonutil.OpArrElemAt, "$$this", 1),
					bsonutil.WrapInOp(bsonutil.OpPow, "$$originalBase",
						bsonutil.WrapInOp(bsonutil.OpSubtract,
							bsonutil.WrapInOp(bsonutil.OpSubtract, "$$length",
								bsonutil.WrapInOp(bsonutil.OpArrElemAt, "$$this", 0)),
							1)))))),
	)

	// numDigits is the length the number will be in the new number base
	// This is equal to: floor(log_newbase(num)) + 1
	numDigits := bsonutil.NewM(
		bsonutil.NewDocElem("numDigits", bsonutil.WrapInOp(bsonutil.OpAdd,
			bsonutil.WrapInOp(bsonutil.OpFloor,
				bsonutil.WrapInOp(bsonutil.OpLog, "$$base10", "$$newBase")), 1)),
	)

	// powers is an array of the powers of the base that you are translating to
	// if the newBase=16 and the resulting number will have length=4 this array
	// will = [1, 16, 256, 4096]
	powers := bsonutil.NewM(
		bsonutil.NewDocElem("powers", bsonutil.WrapInMap(
			bsonutil.WrapInOp(bsonutil.OpRange, bsonutil.WrapInOp(bsonutil.OpSubtract, "$$numDigits", 1), -1, -1),
			"this",
			bsonutil.WrapInOp(bsonutil.OpPow, "$$newBase", "$$this"))),
	)

	// Turns the base10 number into an array of the newBase digits (in their base10 form)
	// i.e. if base10 = 173 (0xAD), numbersArray = [10, 13]
	// Follows generalized version of: https://www.permadi.com/tutorial/numDecToHex/
	generateNumberArray := bsonutil.WrapInMap("$$powers", "this",
		bsonutil.WrapInOp(bsonutil.OpMod,
			bsonutil.WrapInOp(bsonutil.OpFloor,
				bsonutil.WrapInOp(bsonutil.OpDivide, "$$base10", "$$this")), "$$newBase"))

	branches2 := bsonutil.NewMArray()
	for k := 0; k <= len(numToString); k++ {
		branches2 = append(branches2,
			bsonutil.WrapInCase(bsonutil.WrapInOp(bsonutil.OpEq, "$$this", k), numToString[k]))
	}

	// Converts the number array into an array of their character representations
	// i.e. if numbersArray = [10, 13], then charArray=['A', 'D']
	generateCharArray := bsonutil.WrapInMap(generateNumberArray, "this", bsonutil.WrapInSwitch("0", branches2...))

	// Turns the charArray into a single string (the final answer)
	// i.e. if charArray=['A','D'] answer='AD'
	positiveAnswer := bsonutil.NewM(
		bsonutil.NewDocElem("positiveAnswer", bsonutil.WrapInReduce(
			generateCharArray,
			"",
			bsonutil.WrapInOp(bsonutil.OpConcat, "", "$$value", "$$this"),
		)),
	)

	signAdjusted := bsonutil.WrapInCond(bsonutil.WrapInOp(bsonutil.OpConcat, "-", "$$positiveAnswer"),
		"$$positiveAnswer", "$$negative")

	// Puts the nested lets together, checks to make sure that the base is valid,
	// and checks to make sure the entered number is valid as well
	// (invalid = numbers too big like 3 in binary or non-alphanumeric like /)
	// Invalid characters returns an answer of 0, invalid bases return NULL
	return bsonutil.WrapInCond(nil, bsonutil.WrapInLet(normalizedVars,
		bsonutil.WrapInLet(indexOfDecimal,
			bsonutil.WrapInLet(eliminateDecimal,
				bsonutil.WrapInCond(nil,
					bsonutil.WrapInCond("0",
						bsonutil.WrapInLet(createLength,
							bsonutil.WrapInLet(createIndexArr,
								bsonutil.WrapInLet(createCharArr,
									bsonutil.WrapInLet(createNumArr,
										bsonutil.WrapInLet(createInvalidArr,
											bsonutil.WrapInCond("0",
												bsonutil.WrapInLet(generateBase10,
													bsonutil.WrapInLet(numDigits,
														bsonutil.WrapInLet(powers,
															bsonutil.WrapInLet(positiveAnswer,
																signAdjusted)))),
												bsonutil.WrapInOp(bsonutil.OpAnyElementTrue,
													"$$invalidArr"))))))),
						bsonutil.WrapInOp(bsonutil.OpIn, "$$number", bsonutil.NewArray("0", "-0"))),
					bsonutil.WrapInOp(bsonutil.OpOr,
						bsonutil.WrapInOp(bsonutil.OpOr, bsonutil.WrapInOp(bsonutil.OpLt, "$$originalBase", 2),
							bsonutil.WrapInOp(bsonutil.OpGt, "$$originalBase", 36)),
						bsonutil.WrapInOp(bsonutil.OpOr, bsonutil.WrapInOp(bsonutil.OpLt, "$$newBase", 2),
							bsonutil.WrapInOp(bsonutil.OpGt, "$$newBase", 36)))))),
	), bsonutil.WrapInOp(bsonutil.OpEq, nil, num)), nil
}

func (*convFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*convFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

func (*convFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalString, EvalInt64, EvalInt64}
	nExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		nExprs,
	}
}

type convertFunc struct{}

func (*convertFunc) FuncName() string {
	return "convert"
}

var _ translatableToAggregationScalarFunc = (*convertFunc)(nil)

func sqlTypeFromSQLExpr(expr SQLExpr) (EvalType, bool) {
	val, ok := expr.(SQLValue)
	if !ok {
		return EvalNone, false
	}

	var typ EvalType
	switch val.String() {
	case string(parser.SIGNED_BYTES):
		typ = EvalInt64
	case string(parser.UNSIGNED_BYTES):
		typ = EvalUint64
	case string(parser.FLOAT_BYTES):
		typ = EvalDouble
	case string(parser.CHAR_BYTES):
		typ = EvalString
	case string(parser.OBJECT_ID_BYTES):
		typ = EvalObjectID
	case string(parser.DATE_BYTES):
		typ = EvalDate
	case string(parser.DATETIME_BYTES):
		typ = EvalDatetime
	case string(parser.DECIMAL_BYTES):
		typ = EvalDecimal128
	case string(parser.BINARY_BYTES):
		// although we represent binary as a string, conversions
		// to it are always going to be invalid
		return EvalString, false
	case string(parser.TIME_BYTES):
		// this type is not supported yet
		return EvalNone, false
	default:
		panic(fmt.Errorf("invalid value %q", val.String()))
	}

	return typ, true
}

// http://dev.mysql.com/doc/refman/5.7/en/cast-functions.html#function_convert
func (f *convertFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if values[0].IsNull() {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	typ, ok := sqlTypeFromSQLExpr(values[1])
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	return NewSQLConvertExpr(values[0], typ).Evaluate(ctx, cfg, st)
}

func (*convertFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	typ, ok := sqlTypeFromSQLExpr(exprs[1])
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(convert)",
			fmt.Sprintf(
				"cannot push down conversions to %s",
				exprs[1].(SQLValue).String(),
			),
		)
	}

	return NewSQLConvertExpr(exprs[0], typ).ToAggregationLanguage(t)
}

func (*convertFunc) EvalType(exprs []SQLExpr) EvalType {
	typ, ok := sqlTypeFromSQLExpr(exprs[1])
	if !ok {
		return EvalString
	}
	return typ
}

func (*convertFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_cos
type cosFunc struct {
	singleArgFloatMathFunc
}

func (*cosFunc) FuncName() string {
	return "cos"
}

var _ translatableToAggregationScalarFunc = (*cosFunc)(nil)

func (*cosFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(cos)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	input := "$$input"
	inputLetAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("input", bsonutil.WrapInOp(bsonutil.OpAbs, args[0])),
	)

	rem, phase := "$$rem", "$$phase"
	remPhaseAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("rem", bsonutil.WrapInOp(bsonutil.OpMod, input, math.Pi/2)),
		bsonutil.NewDocElem("phase", bsonutil.WrapInOp(bsonutil.OpMod,
			bsonutil.WrapInOp(bsonutil.OpTrunc,
				bsonutil.WrapInOp(bsonutil.OpDivide, input, math.Pi/2),
			),
			4.0)),
	)

	// 3.2 does not support $switch, so just use chained $cond, assuming
	// zeroCase will be most common (since it's the first phase). Because we
	// use the Maclaurin Power Series for sine and cos, we need to adjust
	// our input into a domain that is good for our approximation, that
	// being the first quadrant (phase). For phases outside of the first,
	// we can adjust the functions as:
	//
	// phase | Maclaurin Power Series
	// ------------------------------
	// 0     | cos(rem)
	// 1     | -1 * sin(rem)
	// 2     | -1 * cos(rem)
	// 3     | sin(rem)
	// where the phase is defined as the trunc(input / (pi/2)) % 4
	// and the remainder is input % (pi/2).
	threeCase := bsonutil.WrapInCond(bsonutil.WrapInSinPowerSeries(rem),
		nil,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			3))
	twoCase := bsonutil.WrapInCond(bsonutil.WrapInOp(bsonutil.OpMultiply,
		-1.0,
		bsonutil.WrapInCosPowerSeries(rem)),
		threeCase,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			2))
	oneCase := bsonutil.WrapInCond(bsonutil.WrapInOp(bsonutil.OpMultiply,
		-1.0,
		bsonutil.WrapInSinPowerSeries(rem)),
		twoCase,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			1))
	zeroCase := bsonutil.WrapInCond(bsonutil.WrapInCosPowerSeries(rem),
		oneCase,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			0))

	return bsonutil.WrapInLet(inputLetAssignment,
		bsonutil.WrapInLet(remPhaseAssignment,
			zeroCase),
	), nil
}

type cotFunc struct{}

func (*cotFunc) FuncName() string {
	return "cot"
}

var _ translatableToAggregationScalarFunc = (*cotFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_cot
func (f *cotFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	tan := math.Tan(Float64(values[0]))
	if tan == 0 {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))),
			mysqlerrors.Defaultf(mysqlerrors.ErDataOutOfRange,
				"DOUBLE",
				fmt.Sprintf("'cot(%v)'",
					Float64(values[0])))
	}

	return NewSQLFloat(cfg.sqlValueKind, 1/tan), nil
}

func (*cotFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	sf := &sinFunc{}
	denom, err := sf.FuncToAggregationLanguage(t, []SQLExpr{exprs[0]})
	if err != nil {
		return nil, err
	}

	cf := &cosFunc{}
	num, err := cf.FuncToAggregationLanguage(t, []SQLExpr{exprs[0]})
	if err != nil {
		return nil, err
	}

	// epsilon the smallest value we allow for denom, computed to roughly
	// tie-out with mysqld.
	epsilon := 6.123233995736766e-17
	// Replace abs(denom) < epsilon with epsilon since mysql doesn't seem
	// to return NULL or INF for inf values of tan as one would expect, and
	// we do not want to trigger a $divide by 0.
	return bsonutil.WrapInOp(bsonutil.OpDivide,
		num,
		bsonutil.WrapInCond(epsilon,
			denom,
			bsonutil.WrapInOp(bsonutil.OpLte,
				bsonutil.WrapInOp(bsonutil.OpAbs, denom), epsilon,
			),
		),
	), nil
}

func (*cotFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDouble
}

func (*cotFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type currentDateFunc struct{}

func (*currentDateFunc) FuncName() string {
	return "currentDate"
}

var _ translatableToAggregationScalarFunc = (*currentDateFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_curdate
func (*currentDateFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	now := time.Now().In(schema.DefaultLocale)
	t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, schema.DefaultLocale)
	return NewSQLDate(cfg.sqlValueKind, t), nil

}

func (*currentDateFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	now := time.Now().In(schema.DefaultLocale)
	cd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, schema.DefaultLocale)
	return bsonutil.WrapInLiteral(cd), nil
}

func (*currentDateFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDate
}

func (*currentDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type currentTimestampFunc struct{}

func (*currentTimestampFunc) FuncName() string {
	return "currentTimestamp"
}

var _ translatableToAggregationScalarFunc = (*currentTimestampFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_now
func (*currentTimestampFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	value := time.Now().In(schema.DefaultLocale)
	return NewSQLTimestamp(cfg.sqlValueKind, value), nil
}

func (*currentTimestampFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	now := time.Now().In(schema.DefaultLocale)
	return bsonutil.WrapInLiteral(now), nil
}

func (*currentTimestampFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDatetime
}

func (*currentTimestampFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type curtimeFunc struct{}

func (*curtimeFunc) FuncName() string {
	return "curtime"
}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_curtime
func (*curtimeFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return NewSQLTimestamp(cfg.sqlValueKind, time.Now().In(schema.DefaultLocale)), nil
}

func (*curtimeFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDatetime
}

func (*curtimeFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type dateAddFunc struct{}

func (*dateAddFunc) FuncName() string {
	return "dateAdd"
}

var _ normalizingScalarFunc = (*dateAddFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_date-add
func (f *dateAddFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	_, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	timestampadd := &timestampAddFunc{}
	// Seconds can be fractional values, so our calculateInterval function will not work right
	// (it is fine for all other units, as they must be integral).
	if values[2].String() == Second {
		interval := values[1].SQLFloat()
		return timestampadd.Evaluate(
			ctx, cfg, st,
			[]SQLValue{NewSQLVarchar(cfg.sqlValueKind, Second), interval, values[0]},
		)
	}
	args, neg := dateArithmeticArgs(values[2].String(), values[1])
	unit, interval, err := calculateInterval(values[2].String(), args, neg)
	if err != nil {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	vals := []SQLValue{
		NewSQLVarchar(cfg.sqlValueKind, unit),
		NewSQLInt64(cfg.sqlValueKind, int64(interval)), values[0],
	}
	return timestampadd.Evaluate(ctx, cfg, st, vals)
}

func (*dateAddFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*dateAddFunc) EvalType(exprs []SQLExpr) EvalType {
	if exprs[0].EvalType() == EvalDatetime {
		return EvalDatetime
	}

	if exprs[0].EvalType() == EvalDate {
		if unit, ok := exprs[2].(SQLValue); ok {
			switch unit.String() {
			case Hour, Minute, Second:
				return EvalDatetime
			}
		}
	}

	return EvalString
}

func (*dateAddFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type dateArithmeticFunc struct {
	scalarFunc
	isSub bool
}

func (*dateArithmeticFunc) FuncName() string {
	return "dateArithmetic"
}

var _ normalizingScalarFunc = (*dateArithmeticFunc)(nil)
var _ translatableToAggregationScalarFunc = (*dateArithmeticFunc)(nil)

func (f *dateArithmeticFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateArithmetic)",
			incorrectArgCountMsg,
		)
	}

	var date interface{}
	var ok bool
	var err PushdownFailure
	if _, ok = f.scalarFunc.(*addDateFunc); ok {
		// implementation for ADDDATE(DATE_FORMAT("..."), INTERVAL 0 SECOND)
		var fun *SQLScalarFunctionExpr
		if fun, ok = exprs[0].(*SQLScalarFunctionExpr); ok && fun.Name == "date_format" {
			var dateErr error
			if date, dateErr = t.translateDateFormatAsDate(fun); dateErr != nil {
				date = nil
			}
		}
	}

	if date == nil {
		switch exprs[0].EvalType() {
		case EvalDate, EvalDatetime:
		default:
			return nil, newPushdownFailure(
				"SQLScalarFunctionExpr(dateArithmetic)",
				"cannot push down when first arg is EvalDate or EvalDatetime",
			)
		}

		if date, err = t.ToAggregationLanguage(exprs[0]); err != nil {
			return nil, err
		}
	}

	intervalValue, ok := exprs[1].(SQLValue)
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateArithmetic)",
			"cannot push down without literal interval value",
		)
	}

	if Float64(intervalValue) == 0 {
		return date, nil
	}

	unitValue, ok := exprs[2].(SQLValue)
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateArithmetic)",
			"cannot push down without literal unit value",
		)
	}

	var ms int64
	// Second can be a float rather than an int, so handle Second specially.
	// calculateInterval works for all other units, as they must be integral.
	if unitValue.String() == Second {
		ms = round(Float64(intervalValue) * 1000.0)
	} else {
		unitInterval, neg := dateArithmeticArgs(unitValue.String(), intervalValue)
		unit, interval, err := calculateInterval(unitValue.String(), unitInterval, neg)
		if err != nil {
			return nil, newPushdownFailure(
				"SQLScalarFunctionExpr(dateArithmetic)",
				"failed to calculate interval",
				"error", err.Error(),
			)
		}
		ms, err = unitIntervalToMilliseconds(unit, int64(interval))
		if err != nil {
			return nil, newPushdownFailure(
				"SQLScalarFunctionExpr(dateArithmetic)",
				"failed to convert interval to ms",
				"error", err.Error(),
			)
		}
	}
	if f.isSub {
		ms *= -1
	}

	conds := bsonutil.NewArray()
	if _, ok := bsonutil.GetLiteral(date); !ok {
		conds = append(conds, "$$date")
	}

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("date", date),
	)

	letEvaluation := bsonutil.WrapInNullCheckedCond(
		nil,
		bsonutil.WrapInOp(bsonutil.OpAdd, "$$date", ms),
		conds...,
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil
}

func (*dateArithmeticFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}
	return f
}

type dateDiffFunc struct{}

func (*dateDiffFunc) FuncName() string {
	return "dateDiff"
}

var _ normalizingScalarFunc = (*dateDiffFunc)(nil)
var _ translatableToAggregationScalarFunc = (*dateDiffFunc)(nil)

func (f *dateDiffFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {

	var left, right time.Time
	var ok bool

	parseArgs := func(val SQLValue) (time.Time, bool) {
		var date time.Time

		date, _, ok = strToDateTime(val.String(), false)
		if !ok {
			return date, false
		}

		date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, schema.DefaultLocale)
		return date, true
	}

	if left, ok = parseArgs(values[0]); !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	if right, ok = parseArgs(values[1]); !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	durationDiff := left.Sub(right)
	hoursDiff := durationDiff.Hours()
	daysDiff := hoursDiff / 24

	diff := NewSQLInt64(cfg.sqlValueKind, int64(daysDiff))
	return diff, nil
}

func (*dateDiffFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateDiff)",
			incorrectArgCountMsg,
		)
	}

	var date1, date2 interface{}
	var ok bool
	var err PushdownFailure

	parseArgs := func(expr SQLExpr) (interface{}, PushdownFailure) {
		var value SQLValue
		if value, ok = expr.(SQLValue); ok {
			var date time.Time
			date, _, ok = strToDateTime(value.String(), false)
			if !ok {
				return nil, newPushdownFailure(
					"SQLScalarFunctionExpr(dateDiff)",
					"failed to parse datetime from literal",
				)
			}

			date = time.Date(date.Year(),
				date.Month(),
				date.Day(),
				0,
				0,
				0,
				0,
				schema.DefaultLocale)
			return date, nil
		}
		exprType := expr.EvalType()
		if exprType == EvalDatetime || exprType == EvalDate {
			var date interface{}
			date, err = t.ToAggregationLanguage(expr)
			if err != nil {
				return nil, err
			}
			return date, nil
		}
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateDiff)",
			"argument was not a SQLValue, EvalDate, or EvalDatetime",
		)
	}

	if date1, err = parseArgs(exprs[0]); err != nil {
		return nil, err
	}
	if date2, err = parseArgs(exprs[1]); err != nil {
		return nil, err
	}

	// This division needs to truncate because this is dateDiff not
	// timestampDiff, partial days are dropped.
	days := bsonutil.WrapInOp(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide,
		bsonutil.WrapInOp(bsonutil.OpSubtract, date1, date2), 86400000))
	bound := bsonutil.WrapInCond(106751, -106751, bsonutil.WrapInOp(bsonutil.OpGt, days, 106751))

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("days", days),
	)

	letEvaluation := bsonutil.WrapInNullCheckedCond(
		nil,
		bsonutil.WrapInCond(
			bound,
			"$$days",
			bsonutil.WrapInOp(bsonutil.OpGt, "$$days", 106751),
			bsonutil.WrapInOp(bsonutil.OpLt, "$$days", -106751),
		),
		date1,
		date2,
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil
}

func (*dateDiffFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}
	return f
}

func (*dateDiffFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*dateDiffFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type dateFormatFunc struct{}

func (*dateFormatFunc) FuncName() string {
	return "dateFormat"
}

var _ normalizingScalarFunc = (*dateFormatFunc)(nil)
var _ translatableToAggregationScalarFunc = (*dateFormatFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_date-format
func (f *dateFormatFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	date, _, ok := parseDateTime(values[0].String())
	date = date.In(schema.DefaultLocale)
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	v1, ok := values[1].(SQLVarchar)
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	ret, err := formatDate(ctx, cfg, st, date, v1.String())
	if err != nil {
		return nil, err
	}
	return NewSQLVarchar(cfg.sqlValueKind, ret), nil
}

func (*dateFormatFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateFormat)",
			incorrectArgCountMsg,
		)
	}

	date, err := t.ToAggregationLanguage(exprs[0])
	if err != nil {
		return nil, err
	}

	formatValue, ok := exprs[1].(SQLValue)
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateFormat)",
			"format string was not a literal",
		)
	}

	wrapped, ok := bsonutil.WrapInDateFormat(date, formatValue.String())
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dateFormat)",
			"unable to push down format string",
			"formatString", formatValue.String(),
		)
	}
	return wrapped, nil
}

func (*dateFormatFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*dateFormatFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*dateFormatFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type dateFunc struct{}

func (*dateFunc) FuncName() string {
	return "date"
}

var _ normalizingScalarFunc = (*dateFunc)(nil)
var _ translatableToAggregationScalarFunc = (*dateFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_date
func (f *dateFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	// Too-short numbers are padded differently than too-short strings.
	// strToDateTime (called by parseDateTime) handles padding in the too-short
	// string case. We need to fix the string here, where we can still find out
	// the original input type.
	var str string
	switch values[0].(type) {
	case SQLFloat, SQLDecimal128, SQLInt64:
		noDecimal := strings.Split(values[0].String(), ".")[0]
		intLength := len(noDecimal)
		if intLength > 14 {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
		}
		padLen := 0
		switch intLength {
		case 5, 7, 11, 13:
			padLen = 1
		case 3, 4:
			padLen = 6 - intLength
		case 9, 10:
			padLen = 12 - intLength
		}
		str = strings.Repeat("0", padLen) + noDecimal
	default:
		str = values[0].String()
	}

	t, _, ok := parseDateTime(str)
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	return NewSQLDate(cfg.sqlValueKind, t.Truncate(24*time.Hour)), nil
}

func (*dateFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*dateFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 5, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(date)",
			"cannot push down to MongoDB < 3.5",
		)
	}

	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(date)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	val := "$$val"
	inputLet := bsonutil.NewM(
		bsonutil.NewDocElem("val", args[0]),
	)

	wrapInDateFromString := func(v interface{}) bson.M {
		return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDateFromString, bsonutil.NewM(bsonutil.NewDocElem("dateString", v))))
	}

	// CASE 1: it's already a Mongo date, we just return it.
	isDateType := containsBSONType(val, "date")

	// Strip out the time component in the MongoDB ISODate.
	dateVal := bsonutil.NewM(
		bsonutil.NewDocElem(bsonutil.OpDateFromParts, bsonutil.NewM(
			bsonutil.NewDocElem("year", bsonutil.NewM(bsonutil.NewDocElem("$year", val))),
			bsonutil.NewDocElem("month", bsonutil.NewM(bsonutil.NewDocElem("$month", val))),
			bsonutil.NewDocElem("day", bsonutil.NewM(bsonutil.NewDocElem("$dayOfMonth", val))),
		)),
	)

	dateBranch := bsonutil.WrapInCase(isDateType, dateVal)

	// CASE 2: it's a number.
	isNumber := containsBSONType(val, "int", "decimal", "long", "double")

	// Evaluates to true if val positive and has <= X digits.
	hasUpToXDigits := func(x float64) interface{} {
		return bsonutil.WrapInInRange(val, 0, math.Pow(10, x))
	}

	// This handles converting a number in YYMMDD format to YYYYMMDD.
	// if YY < 70, we assume they meant 20YY. if YY > 70, we assume 19YY.
	getPadding := func(v interface{}) interface{} {
		return bsonutil.WrapInCond(
			20000000,
			19000000,
			bsonutil.WrapInOp(bsonutil.OpLt,
				bsonutil.WrapInOp(bsonutil.OpDivide,
					v, 10000),
				70))
	}

	// We interpret this as being format YYMMDD.
	ifSix := bsonutil.WrapInOp(bsonutil.OpAdd, val, getPadding(val))
	sixBranch := bsonutil.WrapInCase(hasUpToXDigits(6), ifSix)

	// This number is good as is! YYYYMMDD.
	eightBranch := bsonutil.WrapInCase(hasUpToXDigits(8), val)

	// If it's twelve digits, interpret as YYMMDDHHMMSS.
	// first drop the last six digits, then pad like we would a six digit number.
	firstSixDigits := bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide, val, 1000000)))
	ifTwelve := bsonutil.WrapInOp(bsonutil.OpAdd, firstSixDigits, getPadding(firstSixDigits))
	twelveBranch := bsonutil.WrapInCase(hasUpToXDigits(12), ifTwelve)

	// If fourteen, YYYYMMDDHHMMSS. just drop the last six digits.
	ifFourteen := bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide, val, 1000000)))
	fourteenBranch := bsonutil.WrapInCase(hasUpToXDigits(14), ifFourteen)

	// Define "num", the input number normalized to 8 digits, in a "let".
	numberVar := bsonutil.WrapInSwitch(nil, sixBranch, eightBranch, twelveBranch, fourteenBranch)
	numberLetVars := bsonutil.NewM(bsonutil.NewDocElem("num", numberVar))

	dateParts := bsonutil.NewM(
		// YYYYMMDD / 10000 = YYYY.

		bsonutil.NewDocElem("year", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide, "$$num", 10000)))),

		// (YYYYMMDD / 100) % 100 = MM.
		bsonutil.NewDocElem("month", bsonutil.WrapInOp(
			bsonutil.OpMod, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide,
				"$$num", 100)),
			), 100)),

		// YYYYMMDD % 100 = DD.
		bsonutil.NewDocElem("day", bsonutil.WrapInOp(bsonutil.OpMod, "$$num", 100)),
	)

	// Try to avoid aggregation errors by catching obviously invalid dates.
	yearValid := bsonutil.WrapInInRange("$$year", 0, 10000)
	monthValid := bsonutil.WrapInInRange("$$month", 1, 13)
	dayValid := bsonutil.WrapInInRange("$$day", 1, 32)

	makeDateOrNull := bsonutil.WrapInCond(bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDateFromParts, bsonutil.NewM(
		bsonutil.NewDocElem("year", "$$year"),
		bsonutil.NewDocElem("month", "$$month"),
		bsonutil.NewDocElem("day", "$$day"),
	)),
	), nil, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpAnd, bsonutil.NewArray(
		yearValid,
		monthValid,
		dayValid,
	))))

	evaluateNumber := bsonutil.WrapInLet(dateParts, makeDateOrNull)
	handleNumberToDate := bsonutil.WrapInLet(numberLetVars, evaluateNumber)
	numberBranch := bsonutil.WrapInCase(isNumber, handleNumberToDate)

	// CASE 3: it's a string
	isString := containsBSONType(val, "string")

	// First split on T, take first substring, then split that on " ", and take first
	// substring. this gives us just the date part of the string. note that if the
	// string doesn't have T or a space, just returns original string.
	trimmedString := bsonutil.WrapInOp(bsonutil.OpArrElemAt,
		bsonutil.WrapInOp(bsonutil.OpSplit,
			bsonutil.WrapInOp(bsonutil.OpArrElemAt,
				bsonutil.WrapInOp(bsonutil.OpSplit, val, "T"),
				0),
			" "),
		0)

	// Convert the string to an array so we can use map/reduce.
	trimmedAsArray := bsonutil.WrapInStringToArray("$$trimmed")

	// isSeparator evaluates to true if a character is in the defined separator list.
	isSeparator := bsonutil.WrapInOp(bsonutil.OpNeq,
		-1,
		bsonutil.WrapInOp("$indexOfArray",
			bsonutil.DateComponentSeparator,
			"$$c"))

	// Use map to convert all separators in the string to - symbol, and leave numbers as-is.
	separatorsNormalized := bsonutil.WrapInMap(trimmedAsArray,
		"c",
		bsonutil.WrapInCond("-",
			"$$c",
			isSeparator))

	// Use reduce to convert characters back to a single string
	joined := bsonutil.WrapInReduce(separatorsNormalized,
		"",
		bsonutil.WrapInOp(bsonutil.OpConcat,
			"$$value",
			"$$this"))

	// If the third character is a -, or if the string is only 6 digits
	// long and has no slashes, then the string is either format YY-MM-DD
	// or YYMMDD and we need to add the appropriate first two year digits
	// (19xx or 20xx) for MongoDB to understand it
	hasShortYear := bsonutil.WrapInOp(bsonutil.OpOr,
		// Length is only 6, assume YYMMDD.
		bsonutil.WrapInOp(bsonutil.OpEq, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, "$$joined")), 6),
		// Third character is -, assume YY-MM-DD.
		bsonutil.WrapInOp(bsonutil.OpEq,
			"-", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubstr, bsonutil.NewArray(
				"$$joined",
				2,
				1,
			)),
			)))

	// $dateFromString actually pads correctly, but not if "/" is used as
	// the separator (it will assume year is last). If this pushdown is
	// shown to be slow by benchmarks, we should reconsider allowing
	// $dateFromString to handle padding. The change would not be trivial
	// due to how MongoDB cannot handle short dates when there are no
	// separators in the date.
	padYear := bsonutil.WrapInOp(bsonutil.OpConcat,
		bsonutil.WrapInCond(
			"20",
			"19",
			// Check if first two digits < 70 to determine padding.
			bsonutil.WrapInOp(
				bsonutil.OpLt, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubstr, bsonutil.NewArray(
					"$$joined",
					0,
					2,
				))), "70")),
		"$$joined")

	// We have to use nested $lets because in the outer one we define $$trimmed and
	// in the inner one we define $$joined. defining $$joined requires knowing the
	// length of trimmed, so we can't do it all in one step.
	innerIn := bsonutil.WrapInCond(padYear, "$$joined", hasShortYear)
	innerLet := bsonutil.WrapInLet(bsonutil.NewM(bsonutil.NewDocElem("joined", joined)), innerIn)

	// Gracefully handle strings that are too short to possibly be valid by returning null.
	tooShort := bsonutil.WrapInOp(bsonutil.OpLt, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, "$$trimmed")), 6)
	outerIn := bsonutil.WrapInCond(nil, wrapInDateFromString(innerLet), tooShort)
	outerLet := bsonutil.WrapInLet(bsonutil.NewM(bsonutil.NewDocElem("trimmed", trimmedString)), outerIn)

	// Make sure if we get the int 0 we return NULL instead
	// of crashing. MySQL uses '0000-00-00' as an error output for some
	// functions and we encode it as the integer 0 within push down.
	stringBranch := bsonutil.WrapInCase(isString,
		bsonutil.WrapInCond(nil,
			outerLet,
			bsonutil.WrapInOp(bsonutil.OpEq,
				0,
				args[0])))

	out := bsonutil.WrapInLet(
		inputLet,
		bsonutil.WrapInSwitch(
			nil,
			dateBranch,
			numberBranch,
			stringBranch,
		),
	)
	return out, nil
}

func (*dateFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDate
}

func (*dateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dateSubFunc struct{}

func (*dateSubFunc) FuncName() string {
	return "dateSub"
}

var _ normalizingScalarFunc = (*dateSubFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_date-sub
func (f *dateSubFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	dateadd := &dateAddFunc{}

	v := values[1].String()
	if string(v[0]) != "-" {
		v = "-" + v
	} else {
		v = v[1:]
	}

	return dateadd.Evaluate(
		ctx, cfg, st,
		[]SQLValue{values[0], NewSQLVarchar(cfg.sqlValueKind, v), values[2]},
	)
}

func (*dateSubFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*dateSubFunc) EvalType(exprs []SQLExpr) EvalType {
	return (&dateAddFunc{}).EvalType(exprs)
}

func (*dateSubFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type dayNameFunc struct{}

func (*dayNameFunc) FuncName() string {
	return "dayName"
}

var _ reconcilingScalarFunc = (*dayNameFunc)(nil)
var _ translatableToAggregationScalarFunc = (*dayNameFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayname
func (f *dayNameFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	return NewSQLVarchar(cfg.sqlValueKind, t.Weekday().String()), nil
}

func (*dayNameFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dayName)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpArrElemAt, bsonutil.NewArray(
				bsonutil.NewArray(
					time.Sunday.String(),
					time.Monday.String(),
					time.Tuesday.String(),
					time.Wednesday.String(),
					time.Thursday.String(),
					time.Friday.String(),
					time.Saturday.String(),
				),
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
					bsonutil.NewM(bsonutil.NewDocElem("$dayOfWeek", args[0])),
					1,
				))),
			)),
		), args[0],
	), nil
}

func (*dayNameFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalDate)
}

func (*dayNameFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*dayNameFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dayOfMonthFunc struct{}

func (*dayOfMonthFunc) FuncName() string {
	return "dayOfMonth"
}

var _ reconcilingScalarFunc = (*dayOfMonthFunc)(nil)
var _ translatableToAggregationScalarFunc = (*dayOfMonthFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayofmonth
func (f *dayOfMonthFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.Day())), nil
}

func (*dayOfMonthFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dayOfMonth)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$dayOfMonth", args[0]), nil
}

func (*dayOfMonthFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalDate)
}

func (*dayOfMonthFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*dayOfMonthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dayOfWeekFunc struct{}

func (*dayOfWeekFunc) FuncName() string {
	return "dayOfWeek"
}

var _ reconcilingScalarFunc = (*dayOfWeekFunc)(nil)
var _ translatableToAggregationScalarFunc = (*dayOfWeekFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayofweek
func (f *dayOfWeekFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.Weekday())+1), nil
}

func (*dayOfWeekFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dayOfWeek)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$dayOfWeek", args[0]), nil
}

func (*dayOfWeekFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalDate)
}

func (*dayOfWeekFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*dayOfWeekFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dayOfYearFunc struct{}

func (*dayOfYearFunc) FuncName() string {
	return "dayOfYear"
}

var _ reconcilingScalarFunc = (*dayOfYearFunc)(nil)
var _ translatableToAggregationScalarFunc = (*dayOfYearFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayofyear
func (f *dayOfYearFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.YearDay())), nil
}

func (*dayOfYearFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(dayOfYear)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$dayOfYear", args[0]), nil
}

func (*dayOfYearFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalDate)
}

func (*dayOfYearFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*dayOfYearFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dbFunc struct{}

func (*dbFunc) FuncName() string {
	return "db"
}

func (*dbFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return NewSQLVarchar(cfg.sqlValueKind, cfg.dbName), nil
}

func (*dbFunc) SkipConstantFolding() bool {
	return true
}

func (*dbFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*dbFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

// Documentation:
// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_degrees
// Note: by embedding singleArgFloatMathFunc, degreesFunc inherits Evaluate,
// Reconcile, Type, and Validate implementations.
type degreesFunc struct {
	singleArgFloatMathFunc
}

func (*degreesFunc) FuncName() string {
	return "degrees"
}

var _ translatableToAggregationScalarFunc = (*degreesFunc)(nil)

func (*degreesFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(degrees)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInOp(bsonutil.OpDivide, bsonutil.WrapInOp(bsonutil.OpMultiply, args[0], 180.0), math.Pi), nil
}

type dualArgFloatMathFunc func(float64, float64) float64

var _ reconcilingScalarFunc = (*dualArgFloatMathFunc)(nil)
var _ normalizingScalarFunc = (*dualArgFloatMathFunc)(nil)

func (dualArgFloatMathFunc) FuncName() string {
	return "dualArgFloatMath"
}

func (f dualArgFloatMathFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	result := f(Float64(values[0]), Float64(values[1]))
	if math.IsNaN(result) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}
	if math.IsInf(result, 0) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}
	if result == -0 {
		result = 0
	}
	return NewSQLFloat(cfg.sqlValueKind, result), nil
}

func (dualArgFloatMathFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (dualArgFloatMathFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	switch f.Name {
	case "mod":
		return convertAllArgs(f, EvalDouble)
	default:
		return f
	}
}

func (dualArgFloatMathFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDouble
}

func (dualArgFloatMathFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type eltFunc struct{}

func (*eltFunc) FuncName() string {
	return "elt"
}

var _ normalizingScalarFunc = (*eltFunc)(nil)
var _ translatableToAggregationScalarFunc = (*eltFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_elt
func (f *eltFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {

	if hasNullValue(values[0]) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	idx := Int64(values[0])
	if idx <= 0 || int(idx) >= len(values) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	result := values[idx]
	if result.IsNull() {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	return NewSQLVarchar(cfg.sqlValueKind, result.String()), nil
}

func (*eltFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	elems := args[1:]
	index := "$$index"
	// Note: ELT indexes on 1, while arrayElemAt indexes based on 0, so we need to subtract 1.
	return bsonutil.WrapInLet(bsonutil.NewM(
		bsonutil.NewDocElem("index", args[0]),
	), bsonutil.WrapInCond(nil, bsonutil.NewM(
		bsonutil.NewDocElem(bsonutil.OpArrElemAt, bsonutil.NewArray(
			elems,
			bsonutil.WrapInOp(bsonutil.OpSubtract, index, 1),
		)),
	), bsonutil.WrapInOp(bsonutil.OpLte, index, 0)),
	), nil
}

func (*eltFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs[0]) {
		return NewSQLNull(kind, f.EvalType())
	}

	if v, ok := f.Exprs[0].(SQLValue); ok {
		idx := Int64(v)
		if idx <= 0 || int(idx) > len(f.Exprs) {
			return NewSQLNull(kind, f.EvalType())
		}
	}

	return f
}

func (*eltFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*eltFunc) Validate(exprCount int) error {
	if exprCount <= 1 {
		return errIncorrectCount
	}

	return nil
}

type expFunc struct {
	singleArgFloatMathFunc
}

func (*expFunc) FuncName() string {
	return "exp"
}

var _ translatableToAggregationScalarFunc = (*expFunc)(nil)

func (f *expFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(exp)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.NewM(bsonutil.NewDocElem("$exp", args[0])), nil
}

type extractFunc struct{}

func (*extractFunc) FuncName() string {
	return "extract"
}

var _ reconcilingScalarFunc = (*extractFunc)(nil)
var _ translatableToAggregationScalarFunc = (*extractFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_extract
func (f *extractFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[1].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	units := [6]int{t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second()}

	var unitStrs [6]string
	// For certain units, we need to concatenate the unit values as strings
	// before returning the int value as to not lose any number's place
	// value. i.e. SELECT EXTRACT(DayMinute FROM "2006-04-03 06:03:23")
	// should return 30603, not 363.
	for idx, val := range units {
		u := strconv.Itoa(val)
		if len(u) == 1 {
			u = "0" + u
		}
		unitStrs[idx] = u
	}

	switch values[0].String() {
	case Year:
		return NewSQLInt64(cfg.sqlValueKind, int64(units[0])), nil
	case Quarter:
		return NewSQLInt64(cfg.sqlValueKind, int64(math.Ceil(float64(units[1])/3.0))), nil
	case Month:
		return NewSQLInt64(cfg.sqlValueKind, int64(units[1])), nil
	case Week:
		_, w := t.ISOWeek()
		return NewSQLInt64(cfg.sqlValueKind, int64(w)), nil
	case Day:
		return NewSQLInt64(cfg.sqlValueKind, int64(units[2])), nil
	case Hour:
		return NewSQLInt64(cfg.sqlValueKind, int64(units[3])), nil
	case Minute:
		return NewSQLInt64(cfg.sqlValueKind, int64(units[4])), nil
	case Second:
		return NewSQLInt64(cfg.sqlValueKind, int64(units[5])), nil
	case Microsecond:
		return NewSQLInt64(cfg.sqlValueKind, 0), nil
	case YearMonth:
		ym, _ := strconv.ParseInt(unitStrs[0]+unitStrs[1], 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, ym), nil
	case DayHour:
		dh, _ := strconv.ParseInt(unitStrs[2]+unitStrs[3], 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, dh), nil
	case DayMinute:
		dm, _ := strconv.ParseInt(unitStrs[2]+unitStrs[3]+unitStrs[4], 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, dm), nil
	case DaySecond:
		ds, _ := strconv.ParseInt(unitStrs[2]+unitStrs[3]+unitStrs[4]+unitStrs[5], 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, ds), nil
	case DayMicrosecond:
		dms, _ := strconv.ParseInt(unitStrs[2]+unitStrs[3]+unitStrs[4]+unitStrs[5]+"000000", 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, dms), nil
	case HourMinute:
		hm, _ := strconv.ParseInt(unitStrs[3]+unitStrs[4], 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, hm), nil
	case HourSecond:
		hs, _ := strconv.ParseInt(unitStrs[3]+unitStrs[4]+unitStrs[5], 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, hs), nil
	case HourMicrosecond:
		hms, _ := strconv.ParseInt(unitStrs[3]+unitStrs[4]+unitStrs[5]+"000000", 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, hms), nil
	case MinuteSecond:
		ms, _ := strconv.ParseInt(unitStrs[4]+unitStrs[5], 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, ms), nil
	case MinuteMicrosecond:
		mms, _ := strconv.ParseInt(unitStrs[4]+unitStrs[5]+"000000", 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, mms), nil
	case SecondMicrosecond:
		sms, _ := strconv.ParseInt(unitStrs[5]+"000000", 10, 64)
		return NewSQLInt64(cfg.sqlValueKind, sms), nil
	default:
		err := fmt.Errorf("unit type '%v' is not supported", values[0].String())
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), err
	}
}

func (*extractFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(extract)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	bsonMap, ok := args[0].(bson.M)
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(extract)",
			"translateArgs returned something other than bson.M",
		)
	}

	bsonVal, ok := bsonMap["$literal"]
	if !ok {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(extract)",
			"first argument was not translated to a $literal",
		)
	}

	unit, ok := bsonVal.(string)
	if !ok {
		// The unit must absolutely be a string.
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(extract)",
			"first argument was not a string",
		)
	}

	switch unit {
	case "year", "month", "hour", "minute", "second":
		return bsonutil.WrapSingleArgFuncWithNullCheck("$"+unit, args[1]), nil
	case "day":
		return bsonutil.WrapSingleArgFuncWithNullCheck("$dayOfMonth", args[1]), nil
	}
	return nil, newPushdownFailure(
		"SQLScalarFunctionExpr(extract)",
		"unknown unit",
		"unit", unit,
	)
}

func (*extractFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalNone, EvalDatetime}
	nExprs := convertExprs(f.Exprs, argTypes)
	// Do not use constructor here, we already have a valid f.Func to use
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		nExprs,
	}
}

func (*extractFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*extractFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type fieldFunc struct{}

func (*fieldFunc) FuncName() string {
	return "field"
}

var _ normalizingScalarFunc = (*fieldFunc)(nil)
var _ reconcilingScalarFunc = (*fieldFunc)(nil)
var _ translatableToAggregationScalarFunc = (*fieldFunc)(nil)

func (*fieldFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLInt64(cfg.sqlValueKind, 0), nil
	}

	target := values[0]
	candidates := values[1:]

	for idx, candidate := range candidates {
		if candidate == target {
			return NewSQLInt64(cfg.sqlValueKind, int64(idx+1)), nil
		}
	}
	return NewSQLInt64(cfg.sqlValueKind, 0), nil
}

func (*fieldFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLInt64(kind, 0)
	}

	return f
}

func (*fieldFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	var reconcile bool
	firstType := EvalNone
loop:
	for _, expr := range f.Exprs {
		typ := expr.EvalType()
		switch typ {
		case EvalString, EvalInt64,
			EvalDecimal128, EvalDouble:
			// valid types
		default:
			reconcile = true
			break loop
		}
		if firstType == EvalNone {
			firstType = typ
			continue
		}
		if firstType != typ {
			reconcile = true
			break loop
		}
	}

	if reconcile {
		return convertAllArgs(f, EvalDecimal128)
	}
	return f
}

func (*fieldFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*fieldFunc) Validate(exprCount int) error {
	if exprCount <= 1 {
		return errIncorrectVarCount
	}
	return nil
}

func (*fieldFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {

	if len(exprs) <= 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(field)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	target := args[0]
	candidates := args[1:]

	var cases []interface{}
	var results []interface{}
	for idx, candidate := range candidates {
		caseExpr := bsonutil.WrapInOp(bsonutil.OpEq, target, candidate)
		resultExpr := bsonutil.WrapInLiteral(idx + 1)

		cases = append(cases, caseExpr)
		results = append(results, resultExpr)
	}

	var idxSwitch interface{}

	if t.versionAtLeast(3, 4, 0) {
		var branches []bson.M
		for idx, caseExpr := range cases {
			resultExpr := results[idx]
			branch := bsonutil.NewM(bsonutil.NewDocElem("case", caseExpr), bsonutil.NewDocElem("then", resultExpr))
			branches = append(branches, branch)
		}
		idxSwitch = bsonutil.WrapInSwitch(bsonutil.WrapInLiteral(0), branches...)
	} else {
		var lastTerm interface{} = bsonutil.WrapInLiteral(0)

		numTerms := len(cases)
		for idx := numTerms - 1; idx >= 0; idx-- {
			term := bsonutil.WrapInCond(results[idx], lastTerm, cases[idx])
			lastTerm = term
		}

		idxSwitch = lastTerm
	}

	return bsonutil.WrapInNullCheckedCond(bsonutil.WrapInLiteral(0), idxSwitch, args...), nil
}

type floorFunc struct {
	singleArgFloatMathFunc
}

func (*floorFunc) FuncName() string {
	return "floor"
}

var _ translatableToAggregationScalarFunc = (*floorFunc)(nil)

func (*floorFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(floor)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.NewM(bsonutil.NewDocElem("$floor", args[0])), nil
}

type fromDaysFunc struct{}

func (*fromDaysFunc) FuncName() string {
	return "fromDays"
}

var _ normalizingScalarFunc = (*fromDaysFunc)(nil)
var _ reconcilingScalarFunc = (*fromDaysFunc)(nil)
var _ translatableToAggregationScalarFunc = (*fromDaysFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_from-days
func (f *fromDaysFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	parseNumeric := func(s string) string {
		f := func(r rune) bool {
			return !unicode.IsNumber(r) &&
				(r != 46 /* unicode '.' */)
		}
		n := strings.FieldsFunc(s, f)
		if len(n) > 0 {
			return n[0]
		}
		return ""
	}

	v := values[0].String()
	neg := len(v) > 0 && v[0] == '-'
	if neg {
		v = v[1:]
	}
	value, err := strconv.ParseFloat(parseNumeric(v), 64)
	if err != nil {
		return NewSQLVarchar(cfg.sqlValueKind, "0000-00-00"), nil
	}
	if neg {
		value = -value
	}

	if value <= 365.5 || value >= 3652499.5 {
		// Go's zero time starts January 1, year 1, 00:00:00 UTC
		// and thus can not represent the date "0000-00-00". To
		// handle this, we return a varchar instead
		return NewSQLVarchar(cfg.sqlValueKind, "0000-00-00"), nil
	}

	abs, maxGoDurationHours := math.Abs(value-366), int64(106751)
	target := int64(math.Floor(abs + .5))

	// edge cases
	if math.Ceil(abs) == 1 {
		target = 0
	}

	if math.Floor(abs) == 3652133 {
		target = 3652133
	}

	date := zeroDate

	for target > 0 && target > maxGoDurationHours {
		date = date.Add(time.Duration(maxGoDurationHours*24) * time.Hour)
		target -= maxGoDurationHours
	}

	date = date.Add(time.Duration(target*24) * time.Hour).Round(time.Second)

	return NewSQLDate(cfg.sqlValueKind, date.In(schema.DefaultLocale)), nil
}

func (*fromDaysFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(fromDays)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}
	dayOne := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	body := bsonutil.WrapInOp(bsonutil.OpAdd, dayOne,
		bsonutil.WrapInOp(bsonutil.OpMultiply, bsonutil.WrapInRound(args[0]), millisecondsPerDay))
	arg := "$$arg"

	argLetAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("arg", args[0]),
	)

	// This should return "0000-00-00" if the input is too large (> maxFromDays)
	// or too low (< 366).
	return bsonutil.WrapInLet(argLetAssignment, bsonutil.WrapInNullCheckedCond(nil,
		bsonutil.WrapInCond(0,
			body,
			bsonutil.WrapInOp(bsonutil.OpGt, arg, maxFromDays),
			bsonutil.WrapInOp(bsonutil.OpLt, arg, 366),
		),
		arg,
	),
	), nil
}

func (*fromDaysFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*fromDaysFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalInt64)
}

func (*fromDaysFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDate
}

func (*fromDaysFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type fromUnixtimeFunc struct{}

func (*fromUnixtimeFunc) FuncName() string {
	return "fromUnixtime"
}

var _ normalizingScalarFunc = (*fromUnixtimeFunc)(nil)
var _ reconcilingScalarFunc = (*fromUnixtimeFunc)(nil)
var _ translatableToAggregationScalarFunc = (*fromUnixtimeFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_from-unixtime
func (f *fromUnixtimeFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	value := round(Float64(values[0]))
	if value < 0 {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	date := time.Unix(value, 0).In(schema.DefaultLocale)
	if len(values) == 1 {
		return NewSQLTimestamp(cfg.sqlValueKind, date), nil
	}
	ret, err := formatDate(ctx, cfg, st, date, values[1].String())
	if err != nil {
		return nil, err
	}
	return NewSQLVarchar(cfg.sqlValueKind, ret), nil
}

func (*fromUnixtimeFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) > 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(fromUnixtime)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	arg := "$$arg"
	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("arg", args[0]),
	)

	// Just add the argument to 1970-01-01 00:00:00.0000000.
	dayOne := time.Date(1970, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
	letEvaluation := bsonutil.WrapInOp(bsonutil.OpAdd,
		dayOne,
		bsonutil.WrapInOp(bsonutil.OpMultiply,
			bsonutil.WrapInRound(arg),
			1e3))

	ret := bsonutil.WrapInLet(letAssignment,
		bsonutil.WrapInCond(nil,
			letEvaluation,
			bsonutil.WrapInOp(bsonutil.OpLt, arg, bsonutil.WrapInLiteral(0)),
		),
	)

	if len(exprs) == 1 {
		return ret, nil
	}
	if format, ok := exprs[1].(SQLValue); ok {
		wrapped, ok := bsonutil.WrapInDateFormat(ret, format.String())
		if !ok {
			return nil, newPushdownFailure(
				"SQLScalarFunctionExpr(fromUnixtime)",
				"unable to push down format string",
				"formatString", format.String(),
			)
		}
		return wrapped, nil
	}

	return nil, newPushdownFailure(
		"SQLScalarFunctionExpr(fromUnixtime)",
		"unsupported form",
	)
}

func (*fromUnixtimeFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*fromUnixtimeFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalInt64, EvalString}
	nExprs := convertExprs(f.Exprs, argTypes)
	// Do not use constructor here, we already have a valid f.Func to use
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		nExprs,
	}
}

func (*fromUnixtimeFunc) EvalType(exprs []SQLExpr) EvalType {
	if len(exprs) == 1 {
		return EvalDatetime
	}
	return EvalString
}

func (*fromUnixtimeFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

type greatestFunc struct{}

func (*greatestFunc) FuncName() string {
	return "greatest"
}

var _ normalizingScalarFunc = (*greatestFunc)(nil)
var _ translatableToAggregationScalarFunc = (*greatestFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_greatest
func (f *greatestFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	valExprs := []SQLExpr{}
	for _, val := range values {
		valExprs = append(valExprs, val)
	}
	convertTo := preferentialType(valExprs...)

	convertedVals := []SQLValue{}
	for _, val := range values {
		newVal := ConvertTo(val, convertTo)
		convertedVals = append(convertedVals, newVal)
	}

	var greatest SQLValue
	var greatestIdx int

	c, err := CompareTo(convertedVals[0], convertedVals[1], st.collation)
	if err != nil {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), err
	}

	if c == -1 {
		greatest, greatestIdx = values[1], 1
	} else {
		greatest, greatestIdx = values[0], 0
	}

	for i := 2; i < len(values); i++ {
		c, err = CompareTo(greatest, convertedVals[i], st.collation)
		if err != nil {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), err
		}
		if c == -1 {
			greatest, greatestIdx = values[i], i
		}
	}

	allTimeVals, timestamp := areAllTimeTypes(values)
	if allTimeVals && timestamp {
		t, _, _ := parseDateTime(values[greatestIdx].String())
		return NewSQLTimestamp(cfg.sqlValueKind, t), nil
	} else if convertTo == EvalDate || convertTo == EvalDatetime {
		return values[greatestIdx], nil
	}

	return convertedVals[greatestIdx], nil
}

func (*greatestFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	// we can only push down if the types are similar
	for i := 1; i < len(exprs); i++ {
		if !isSimilar(exprs[0].EvalType(), exprs[i].EvalType()) {
			return nil, newPushdownFailure("SQLScalarFunctionExpr(greatest)", "arguments' types are not similar")
		}
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(bsonutil.NewDocElem("$max", args)), args...,
	), nil
}

func (*greatestFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*greatestFunc) EvalType(exprs []SQLExpr) EvalType {
	return preferentialType(exprs...)
}

func (*greatestFunc) Validate(exprCount int) error {
	if exprCount < 2 {
		return errIncorrectVarCount
	}
	return nil
}

type hourFunc struct{}

func (*hourFunc) FuncName() string {
	return "hour"
}

var _ normalizingScalarFunc = (*hourFunc)(nil)
var _ translatableToAggregationScalarFunc = (*hourFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_hour
func (f *hourFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if values[0].IsNull() {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}
	_, hour, ok := parseTime(values[0].String())
	if !ok {
		// If we managed to parse, but minutes or seconds are >= 60
		// MySQL returns NULL for the hour/minute/second function.
		// Rather than return yet another value, we coop the hour value
		// and return -1, thus we can check for -1 here to return NULL
		// rather than the 0 expected if the string could not be parsed
		// at all.
		if hour == -1 {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
		}
		return NewSQLInt64(cfg.sqlValueKind, 0), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(hour)), nil
}

func (*hourFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(hour)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$hour", args[0]), nil
}

func (*hourFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*hourFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*hourFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type ifFunc struct{}

func (*ifFunc) FuncName() string {
	return "if"
}

var _ reconcilingScalarFunc = (*ifFunc)(nil)
var _ translatableToAggregationScalarFunc = (*ifFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/control-flow-functions.html#function_if
func (f *ifFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if values[0].IsNull() {
		return values[2], nil
	}

	switch typedV := values[0].(type) {
	case SQLBool:
		if Bool(typedV) {
			return values[1], nil
		}
		return values[2], nil
	case SQLDate, SQLTimestamp:
		return values[1], nil
	case SQLInt64, SQLFloat:
		v := Float64(typedV)
		if v == 0 {
			return values[2], nil
		}
		return values[1], nil
	case SQLVarchar:
		if v, _ := strconv.ParseFloat(typedV.String(), 64); v == 0 {
			return values[2], nil
		}
		return values[1], nil
	default:
		err := fmt.Errorf("expression type '%v' is not supported", typedV)
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), err
	}
}

func (*ifFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(if)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	if value, ok := bsonutil.GetLiteral(args[0]); ok {
		if value == nil || value == false || value == 0 {
			return args[2], nil
		} else {
			return args[1], nil
		}
	}

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("expr", args[0]),
	)

	letEvaluation := bsonutil.WrapInCond(
		args[2],
		args[1],
		bsonutil.WrapInNullCheck("$$expr"),
		bsonutil.WrapInOp(bsonutil.OpEq, "$$expr", 0),
		bsonutil.WrapInOp(bsonutil.OpEq, "$$expr", false),
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

func (*ifFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return f
}

func (*ifFunc) EvalType(exprs []SQLExpr) EvalType {
	s := &EvalTypeSorter{VarcharHighPriority: true}
	return preferentialTypeWithSorter(s, exprs[1:]...)
}

func (*ifFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type ifnullFunc struct{}

func (*ifnullFunc) FuncName() string {
	return "ifnull"
}

var _ normalizingScalarFunc = (*ifnullFunc)(nil)
var _ reconcilingScalarFunc = (*ifnullFunc)(nil)
var _ translatableToAggregationScalarFunc = (*ifnullFunc)(nil)

func (*ifnullFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if values[0].IsNull() {
		return values[1], nil
	}
	return values[0], nil
}

func (*ifnullFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(ifnull)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInIfNull(args[0], args[1]), nil
}

func (*ifnullFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	sqlVal, ok := f.Exprs[0].(SQLValue)
	if ok {
		if sqlVal.IsNull() {
			return f.Exprs[1]
		}
		return sqlVal
	}

	return f
}

func (*ifnullFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return f
}

func (*ifnullFunc) EvalType(exprs []SQLExpr) EvalType {
	s := &EvalTypeSorter{VarcharHighPriority: true}
	return preferentialTypeWithSorter(s, exprs...)
}

func (*ifnullFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type isnullFunc struct{}

func (*isnullFunc) FuncName() string {
	return "isnull"
}

var _ translatableToAggregationScalarFunc = (*isnullFunc)(nil)

func (f *isnullFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	s := NewSQLIsExpr(values[0], NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))))
	return s.Evaluate(ctx, cfg, st)
}

func (f *isnullFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	s := NewSQLIsExpr(exprs[0], NewSQLNull(t.valueKind(), f.EvalType(exprs)))
	return s.ToAggregationLanguage(t)
}

func (*isnullFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*isnullFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type insertFunc struct{}

func (*insertFunc) FuncName() string {
	return "insert"
}

var _ normalizingScalarFunc = (*insertFunc)(nil)
var _ reconcilingScalarFunc = (*insertFunc)(nil)
var _ translatableToAggregationScalarFunc = (*insertFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_insert
func (f *insertFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {

	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	s := values[0].String()
	pos := int(round(Float64(values[1]))) - 1
	length := int(round(Float64(values[2])))
	newstr := values[3].String()

	if pos < 0 || pos >= len(s) {
		return values[0], nil
	}

	if pos+length < 0 || pos+length > len(s) {
		return NewSQLVarchar(cfg.sqlValueKind, s[:pos]+newstr), nil
	}

	return NewSQLVarchar(cfg.sqlValueKind, s[:pos]+newstr+s[pos+length:]), nil
}

func (*insertFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(insert)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 4 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(insert)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	str, pos, len, newstr := "$$str", "$$pos", "$$len", "$$newstr"
	inputAssignment := bsonutil.NewM(
		// SQL uses 1 indexing, so makes sure to subtract 1 to
		// account for MongoDB's 0 indexing.
		bsonutil.NewDocElem("str", args[0]),
		bsonutil.NewDocElem("pos", bsonutil.WrapInRound(bsonutil.WrapInOp(bsonutil.OpSubtract, args[1], 1))),
		bsonutil.NewDocElem("len", bsonutil.WrapInRound(args[2])),
		bsonutil.NewDocElem("newstr", args[3]),
	)

	totalLength := "$$totalLength"
	totalLengthAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("totalLength", bsonutil.WrapInOp(bsonutil.OpStrlenCP, str)),
	)

	prefix, suffix := "$$prefix", "$$suffix"
	ixAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("prefix", bsonutil.WrapInOp(bsonutil.OpSubstr, str, 0, pos)),
		bsonutil.NewDocElem("suffix", bsonutil.WrapInOp(bsonutil.OpSubstr, str, bsonutil.WrapInOp(bsonutil.OpAdd, pos, len), totalLength)),
	)

	concatenation := bsonutil.WrapInLet(ixAssignment,
		bsonutil.WrapInOp(bsonutil.OpConcat, prefix, newstr, suffix),
	)

	posCheck := bsonutil.WrapInLet(totalLengthAssignment,
		bsonutil.WrapInCond(str,
			concatenation,
			bsonutil.WrapInOp(bsonutil.OpLte, pos, 0),
			bsonutil.WrapInOp(bsonutil.OpGte, pos, totalLength),
		),
	)

	return bsonutil.WrapInLet(inputAssignment,
		bsonutil.WrapInCond(nil,
			posCheck,
			bsonutil.WrapInOp(bsonutil.OpLte, str, nil),
			bsonutil.WrapInOp(bsonutil.OpLte, pos, nil),
			bsonutil.WrapInOp(bsonutil.OpLte, len, nil),
			bsonutil.WrapInOp(bsonutil.OpLte, newstr, nil),
		),
	), nil
}

func (*insertFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*insertFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalString, EvalInt64, EvalInt64, EvalString}
	convertedExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*insertFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*insertFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 4)
}

type instrFunc struct{}

func (*instrFunc) FuncName() string {
	return "instr"
}

var _ normalizingScalarFunc = (*instrFunc)(nil)
var _ reconcilingScalarFunc = (*instrFunc)(nil)
var _ translatableToAggregationScalarFunc = (*instrFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_instr
func (*instrFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	locate := &locateFunc{}
	return locate.Evaluate(
		ctx, cfg, st,
		[]SQLValue{values[1], values[0]},
	)
}

func (*instrFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(instr)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(instr)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	// Mongo Aggregation Pipeline returns NULL if arg1 is NULLish, like
	// we'd want. arg2 being NULL, however, is an error in the pipeline,
	// thus check arg2 for NULLisness.
	arg2 := "$$arg2"
	return bsonutil.WrapInLet(bsonutil.NewM(
		bsonutil.NewDocElem("arg2", args[1]),
	), bsonutil.WrapInCond(nil,
		bsonutil.WrapInOp(bsonutil.OpAdd,
			bsonutil.WrapInOp(bsonutil.OpIndexOfCP, args[0], arg2),
			1,
		),
		bsonutil.WrapInOp(bsonutil.OpLte, arg2, nil),
	),
	), nil
}

func (*instrFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*instrFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalString)
}

func (*instrFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*instrFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type intervalFunc struct{}

func (*intervalFunc) FuncName() string {
	return "interval"
}

var _ normalizingScalarFunc = (*intervalFunc)(nil)
var _ reconcilingScalarFunc = (*intervalFunc)(nil)
var _ translatableToAggregationScalarFunc = (*intervalFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_interval
func (*intervalFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if values[0].IsNull() {
		return NewSQLInt64(cfg.sqlValueKind, -1), nil
	}

	start, end := 1, len(values)-1
	for start != end {
		mid := (start + end + 1) / 2
		if values[mid].IsNull() || Float64(values[mid]) <= Float64(values[0]) {
			start = mid
		} else {
			end = mid - 1
		}
	}

	if values[start].IsNull() || Float64(values[start]) <= Float64(values[0]) {
		return NewSQLInt64(cfg.sqlValueKind, int64(start)), nil
	}
	return NewSQLInt64(cfg.sqlValueKind, int64(start-1)), nil
}

func (*intervalFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}
	return bsonutil.WrapInCond(
		bsonutil.WrapInLiteral(-1), bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpReduce, bsonutil.NewM(
				bsonutil.NewDocElem("input", args[1:]),
				bsonutil.NewDocElem("initialValue", bsonutil.WrapInLiteral(0)),
				bsonutil.NewDocElem("in", bsonutil.WrapInCond(bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpAdd, bsonutil.NewArray(
					"$$value",
					bsonutil.WrapInLiteral(1),
				))), "$$value", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGte, bsonutil.NewArray(
					args[0],
					"$$this",
				))))),
			)),
		), bsonutil.WrapInNullCheck(args[0]),
	), nil
}

func (*intervalFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	sqlVal, ok := f.Exprs[0].(SQLValue)
	if ok && sqlVal.IsNull() {
		return NewSQLInt64(kind, -1)
	}
	return f
}

func (*intervalFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalDouble)
}

func (*intervalFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*intervalFunc) Validate(exprCount int) error {
	if exprCount < 2 {
		return errIncorrectVarCount
	}
	return nil
}

type lastDayFunc struct{}

func (*lastDayFunc) FuncName() string {
	return "lastDay"
}

var _ normalizingScalarFunc = (*lastDayFunc)(nil)
var _ translatableToAggregationScalarFunc = (*lastDayFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_last-day
func (f *lastDayFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	// Must be a SQLTimestamp at this point. If it is not, the algebrizer
	// has been broken. Check where the algebrizer handles
	// scalar functions.
	tmp, ok := values[0].(SQLDate)
	if !ok {
		return nil,
			fmt.Errorf("unable to evaluate type %v"+
				" in to_days, this points to an error in the algebrizer",
				values[0])
	}
	t := Timestamp(tmp)
	year, month, _ := t.Date()
	first := time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
	return NewSQLDate(cfg.sqlValueKind, first.AddDate(0, 1, -1)), nil
}

func (*lastDayFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*lastDayFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(lastDay)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	date := "$$date"
	outerLetAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("date", args[0]),
	)

	letAssigment := bsonutil.NewM(
		bsonutil.NewDocElem("year", bsonutil.WrapInOp(bsonutil.OpYear, date)),
		bsonutil.NewDocElem("month", bsonutil.WrapInOp(bsonutil.OpMonth, date)),
	)

	year, month := "$$year", "$$month"
	var letEvaluation bson.M

	// Underflow and overflow in date computation are supported on MongoDB versions >= 4.0.
	// For example, a month value greater than 12 (overflow) and a day value of zero
	// (underflow) are supported date values.
	if t.versionAtLeast(4, 0, 0) {
		// MongoDB interprets day 0 of a given month as the last day of the previous month.
		letEvaluation = bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpDateFromParts, bsonutil.NewM(
				bsonutil.NewDocElem("year", year),
				bsonutil.NewDocElem("month", bsonutil.WrapInOp(bsonutil.OpAdd, 1, month)),
				bsonutil.NewDocElem("day", 0),
			)),
		)

	} else {

		// For MongoDB versions < 4.0, underflow and overflow in date computation are not
		// supported. For example, a day value of zero or a month value of 13 in a date
		// generates an error. In this case, we create a switch on the month value,
		// extracted from $dateFromParts, to determine the last day of the month.
		letEvaluation = bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpDateFromParts, bsonutil.NewM(
				bsonutil.NewDocElem("year", year),
				bsonutil.NewDocElem("month", month),
				bsonutil.NewDocElem("day",
					// The following MongoDB aggregation language implements this go code,
					// which is designed to set the day of a date to the last day of the month.
					// switch month {
					// case 2:
					// 	if isLeapYear(year) == 0 {
					// 		day = 29
					//	} else {
					//		day = 28
					//	}
					// case 4, 6, 9, 11:
					//	day = 30
					// default:
					//      day = 31
					// }
					bsonutil.WrapInSwitch(31,
						bsonutil.WrapInEqCase(month, 2,
							bsonutil.WrapInCond(29, 28, bsonutil.WrapInIsLeapYear(year)),
						),
						bsonutil.WrapInEqCase(month, 4, 30),
						bsonutil.WrapInEqCase(month, 6, 30),
						bsonutil.WrapInEqCase(month, 9, 30),
						bsonutil.WrapInEqCase(month, 11, 30),
					)),
			)),
		)

	}

	return bsonutil.WrapInLet(outerLetAssignment, bsonutil.WrapInLet(letAssigment, letEvaluation)), nil
}

func (*lastDayFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDate
}

func (*lastDayFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type lcaseFunc struct{}

func (*lcaseFunc) FuncName() string {
	return "lcase"
}

var _ reconcilingScalarFunc = (*lcaseFunc)(nil)
var _ translatableToAggregationScalarFunc = (*lcaseFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_lcase
func (f *lcaseFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	value := strings.ToLower(values[0].String())

	return NewSQLVarchar(cfg.sqlValueKind, value), nil
}

func (*lcaseFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(lcase)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$toLower", args[0]), nil
}

func (*lcaseFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalString)
}

func (*lcaseFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*lcaseFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type leastFunc struct{}

func (*leastFunc) FuncName() string {
	return "least"
}

var _ normalizingScalarFunc = (*leastFunc)(nil)
var _ translatableToAggregationScalarFunc = (*leastFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_least
func (f *leastFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	valExprs := []SQLExpr{}
	for _, val := range values {
		valExprs = append(valExprs, val)
	}
	convertTo := preferentialType(valExprs...)

	convertedVals := []SQLValue{}
	for _, val := range values {
		newVal := ConvertTo(val, convertTo)
		convertedVals = append(convertedVals, newVal)
	}

	var least SQLValue
	var leastIdx int

	c, err := CompareTo(convertedVals[0], convertedVals[1], st.collation)
	if err != nil {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), err
	}

	if c == -1 {
		least, leastIdx = convertedVals[0], 0
	} else {
		least, leastIdx = convertedVals[1], 1
	}

	for i := 2; i < len(values); i++ {
		c, err = CompareTo(least, convertedVals[i], st.collation)
		if err != nil {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), err
		}
		if c == 1 {
			least, leastIdx = values[i], i
		}
	}

	allTimeVals, timestamp := areAllTimeTypes(values)
	if allTimeVals && timestamp {
		t, _, _ := parseDateTime(values[leastIdx].String())
		return NewSQLTimestamp(cfg.sqlValueKind, t), nil
	} else if convertTo == EvalDate || convertTo == EvalDatetime {
		return values[leastIdx], nil
	}

	return convertedVals[leastIdx], nil
}

func (*leastFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	// we can only push down if the types are similar
	for i := 1; i < len(exprs); i++ {
		if !isSimilar(exprs[0].EvalType(), exprs[i].EvalType()) {
			return nil, newPushdownFailure("SQLScalarFunctionExpr(least)", "arguments' types are not similar")
		}
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(bsonutil.NewDocElem("$min", args)), args...,
	), nil

}

func (*leastFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*leastFunc) EvalType(exprs []SQLExpr) EvalType {
	return preferentialType(exprs...)
}

func (*leastFunc) Validate(exprCount int) error {
	if exprCount < 2 {
		return errIncorrectVarCount
	}
	return nil
}

type leftFunc struct{}

func (*leftFunc) FuncName() string {
	return "left"
}

var _ reconcilingScalarFunc = (*leftFunc)(nil)
var _ translatableToAggregationScalarFunc = (*leftFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_left
func (f *leftFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	substring,
		err := NewSQLScalarFunctionExpr("substring",
		[]SQLExpr{values[0],
			NewSQLInt64(cfg.sqlValueKind, 1),
			values[1]})
	if err != nil {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), err
	}
	return substring.Evaluate(ctx, cfg, st)
}

func (*leftFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(left)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(left)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	conds := bsonutil.NewArray()
	var subStrLength interface{}

	if stringValue, ok := bsonutil.GetLiteral(args[0]); ok {
		if stringValue == nil {
			return bsonutil.MgoNullLiteral, nil
		}
	} else {
		conds = append(conds, "$$string")
	}

	if lengthValue, ok := bsonutil.GetLiteral(args[1]); ok {
		if lengthValue == nil {
			return bsonutil.MgoNullLiteral, nil
		}

		// when length is negative, just use 0. round length to closest integer
		if i, ok := lengthValue.(int64); ok {
			args[1] = bsonutil.WrapInLiteral(int64(math.Max(0, float64(i))))
			subStrLength = "$$length"
		} else {
			args[1] = bsonutil.WrapInRound(bsonutil.WrapInOp(bsonutil.OpMax, args[1], 0))
			subStrLength = "$$length"
		}
	} else {
		subStrLength = bsonutil.WrapInRound(bsonutil.WrapInOp(bsonutil.OpMax, "$$length", 0))
		conds = append(conds, "$$length")
	}

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("string", args[0]),
		bsonutil.NewDocElem("length", args[1]),
	)

	subStrOp := bsonutil.WrapInOp(bsonutil.OpSubstr, "$$string", 0, subStrLength)

	letEvaluation := bsonutil.WrapInNullCheckedCond(nil, subStrOp, conds...)
	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil
}

func (*leftFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalString, EvalInt64}
	nExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		nExprs,
	}
}

func (*leftFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*leftFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type lengthFunc struct{}

func (*lengthFunc) FuncName() string {
	return "length"
}

var _ reconcilingScalarFunc = (*lengthFunc)(nil)
var _ translatableToAggregationScalarFunc = (*lengthFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_length
func (f *lengthFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	value := values[0].String()

	return NewSQLInt64(cfg.sqlValueKind, int64(len(value))), nil
}

func (*lengthFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(length)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(length)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$strLenBytes", args[0]), nil
}

func (*lengthFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalString)
}

func (*lengthFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*lengthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type locateFunc struct{}

func (*locateFunc) FuncName() string {
	return "locate"
}

var _ normalizingScalarFunc = (*locateFunc)(nil)
var _ translatableToAggregationScalarFunc = (*locateFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_locate
func (f *locateFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values[:2]...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	substr := []rune(values[0].String())
	str := []rune(values[1].String())
	var result int
	if len(values) == 3 {

		pos := int(Float64(values[2])+0.5) - 1 // MySQL uses 1 as a basis

		if pos < 0 || len(str) <= pos {
			return NewSQLInt64(cfg.sqlValueKind, 0), nil
		}
		str = str[pos:]
		result = runesIndex(str, substr)
		if result >= 0 {
			result += pos
		}
	} else {
		result = runesIndex(str, substr)
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(result+1)), nil
}

func (*locateFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(locate)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if !(len(exprs) == 2 || len(exprs) == 3) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(locate)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	var locate interface{}
	substr := args[0]
	str := args[1]

	if len(args) == 2 {
		indexOfCP := bsonutil.NewM(bsonutil.NewDocElem("$indexOfCP", bsonutil.NewArray(
			str,
			substr,
		)))
		locate = bsonutil.WrapInOp(bsonutil.OpAdd, indexOfCP, 1)
	} else if len(args) == 3 {
		// if the pos arg is null, we should return 0, not null
		// this is the same result as when the arg is 0
		pos := bsonutil.WrapInIfNull(args[2], 0)

		// round to the nearest int
		pos = bsonutil.WrapInOp(bsonutil.OpAdd, pos, 0.5)
		pos = bsonutil.WrapInOp(bsonutil.OpTrunc, pos)

		// subtract 1 from the pos arg to reconcile indexing style
		pos = bsonutil.WrapInOp(bsonutil.OpSubtract, pos, 1)

		indexOfCP := bsonutil.NewM(bsonutil.NewDocElem("$indexOfCP", bsonutil.NewArray(
			str,
			substr,
			pos,
		)))
		locate = bsonutil.WrapInOp(bsonutil.OpAdd, indexOfCP, 1)

		// if the pos argument was negative, we should return 0
		locate = bsonutil.WrapInCond(
			0,
			locate,
			bsonutil.WrapInOp(bsonutil.OpLt, pos, 0),
		)
	}

	return bsonutil.WrapInNullCheckedCond(
		nil,
		locate,
		str, substr,
	), nil
}

func (*locateFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs[:2]...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*locateFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*locateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2, 3)
}

type logFunc struct {
	Base uint32 // 0 for natural log.
}

func (*logFunc) FuncName() string {
	return "log"
}

var _ normalizingScalarFunc = (*logFunc)(nil)
var _ reconcilingScalarFunc = (*logFunc)(nil)
var _ translatableToAggregationScalarFunc = (*logFunc)(nil)

func (f logFunc) Evaluate(_ context.Context, cfg *ExecutionConfig, _ *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	var result float64
	switch f.Base {
	case 0:
		if len(values) == 2 {
			// arbitrary base
			result = math.Log(Float64(values[1])) / math.Log(Float64(values[0]))
		} else {
			// natural base
			result = math.Log(Float64(values[0]))
		}
	case 2:
		result = math.Log2(Float64(values[0]))
	case 10:
		result = math.Log10(Float64(values[0]))
	}
	if math.IsNaN(result) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}
	if math.IsInf(result, 0) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}
	if result == -0 {
		result = 0
	}
	return NewSQLFloat(cfg.sqlValueKind, result), nil
}

func (f *logFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) < 1 || len(exprs) > 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(log)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	// Use ln func rather than log with go's value for E, to avoid compromising values
	// more than we already do between MongoDB and MySQL by introducing a third value for E
	// (i.e., go's)
	if f.Base == 0 {
		// 1 arg implies natural log
		if len(args) == 1 {
			return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCond, bsonutil.NewArray(
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGt, bsonutil.NewArray(
					args[0],
					0,
				))),
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpNaturalLog, args[0])),
				bsonutil.MgoNullLiteral,
			))), nil
		}
		// Two args is based arg.
		// MySQL specifies base then arg, MongoDB expects arg then base, so we have to flip.
		return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCond, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGt, bsonutil.NewArray(
				args[0],
				0,
			))),
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpLog, bsonutil.NewArray(
				args[1],
				args[0],
			))),
			bsonutil.MgoNullLiteral,
		))), nil
	}
	// This will be base 10 or base 2 based on if log10 or log2 was called.
	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCond, bsonutil.NewArray(
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGt, bsonutil.NewArray(
			args[0],
			0,
		))),
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpLog, bsonutil.NewArray(
			args[0],
			f.Base,
		))),
		bsonutil.MgoNullLiteral,
	))), nil
}

func (logFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}
	return f
}

func (logFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalDouble)
}

func (logFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDouble
}

func (f logFunc) Validate(exprCount int) error {
	if f.Base == 0 {
		return ensureArgCount(exprCount, 1, 2)
	}
	return ensureArgCount(exprCount, 1)
}

type lpadFunc struct{}

func (*lpadFunc) FuncName() string {
	return "lpad"
}

var _ normalizingScalarFunc = (*lpadFunc)(nil)
var _ reconcilingScalarFunc = (*lpadFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_lpad
func (*lpadFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return handlePadding(cfg.sqlValueKind, values, true)
}

func (*lpadFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*lpadFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalString, EvalInt64, EvalString}
	convertedExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*lpadFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*lpadFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type ltrimFunc struct{}

func (*ltrimFunc) FuncName() string {
	return "ltrim"
}

var _ reconcilingScalarFunc = (*ltrimFunc)(nil)
var _ translatableToAggregationScalarFunc = (*ltrimFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_ltrim
func (f *ltrimFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	value := strings.TrimLeft(values[0].String(), " ")

	return NewSQLVarchar(cfg.sqlValueKind, value), nil
}

func (*ltrimFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(ltrim)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(ltrim)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	if t.versionAtLeast(4, 0, 0) {
		return bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpLTrim, bsonutil.NewM(
				bsonutil.NewDocElem("input", args[0]),
				bsonutil.NewDocElem("chars", " "),
			)),
		), nil
	}

	ltrimCond := bsonutil.WrapInCond(
		"",
		bsonutil.WrapInLRTrim(true, args[0]), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
			args[0],
			"",
		))),
	)

	return bsonutil.WrapInNullCheckedCond(
		nil,
		ltrimCond,
		args[0],
	), nil
}

func (*ltrimFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalString)
}

func (*ltrimFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*ltrimFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type makeDateFunc struct{}

func (*makeDateFunc) FuncName() string {
	return "makeDate"
}

var _ normalizingScalarFunc = (*makeDateFunc)(nil)
var _ reconcilingScalarFunc = (*makeDateFunc)(nil)
var _ translatableToAggregationScalarFunc = (*makeDateFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_makedate
func (f *makeDateFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	// Floating arguments should be rounded.
	y := round(Float64(values[0]))
	if y < 0 || y > 9999 {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}
	if y >= 0 && y <= 69 {
		y += 2000
	} else if y >= 70 && y <= 99 {
		y += 1900
	}

	d := round(Float64(values[1]))

	if d <= 0 {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	t := time.Date(int(y), 1, 0, 0, 0, 0, 0, schema.DefaultLocale)
	duration := time.Duration(d*24) * time.Hour

	output := t.Add(duration)
	if output.Year() > 9999 {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	return NewSQLDate(cfg.sqlValueKind, output), nil
}

func (*makeDateFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 5, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(makeDate)",
			"cannot push down to MongoDB < 3.5",
		)
	}

	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(makeDate)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	year, day, paddedYear, output := "$$year", "$$day", "$$paddedYear", "$$output"

	inputLetStatement := bsonutil.NewM(
		bsonutil.NewDocElem("year", bsonutil.WrapInRound(args[0])),
		bsonutil.NewDocElem("day", bsonutil.WrapInRound(args[1])),
	)

	branch1900 := bsonutil.WrapInCond(
		bsonutil.WrapInOp(bsonutil.OpAdd, year, 1900),
		year,
		bsonutil.WrapInOp(bsonutil.OpAnd,
			bsonutil.WrapInOp(bsonutil.OpGte, year, 70),
			bsonutil.WrapInOp(bsonutil.OpLte, year, 99),
		))

	branch2000 := bsonutil.WrapInOp(bsonutil.OpAdd, year, 2000)

	// $$paddedYear holds the year + 2000 for years between 0 and 69, and +
	// 1900 for years between 70 and 99. Otherwise, it is the original
	// year.
	paddedYearLetStatement := bsonutil.NewM(bsonutil.NewDocElem("paddedYear", bsonutil.WrapInCond(branch2000, branch1900,
		bsonutil.WrapInOp(bsonutil.OpAnd,
			bsonutil.WrapInOp(bsonutil.OpGte,
				year,
				0),
			bsonutil.WrapInOp(bsonutil.OpLte,
				year,
				69)),
	)),
	)

	// This implements:
	// date(paddedYear) + (day - 1) * millisecondsPerDay.
	addDaysStatement := bsonutil.WrapInOp(
		bsonutil.OpAdd,
		bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpDateFromParts,
				bsonutil.NewM(bsonutil.NewDocElem("year", paddedYear)))),
		bsonutil.WrapInOp(bsonutil.OpMultiply,
			bsonutil.WrapInOp(bsonutil.OpSubtract, day, 1),
			millisecondsPerDay),
	)

	// If the $$paddedYear is more than 9999 or less than 0, return NULL.
	yearRangeCheck := bsonutil.WrapInCond(
		nil,
		addDaysStatement,
		bsonutil.WrapInOp(bsonutil.OpLt, paddedYear, 0),
		bsonutil.WrapInOp(bsonutil.OpGt, paddedYear, 9999),
	)

	// Day range check, return NULL if day < 1.
	dayRangeCheck := bsonutil.WrapInCond(nil,
		yearRangeCheck,
		bsonutil.WrapInOp(bsonutil.OpLt, day, 1),
	)

	outputLetStatement := bsonutil.NewM(bsonutil.NewDocElem("output", dayRangeCheck))

	// Bind lets, and check that output value year < 9999, otherwise MySQL
	// returns NULL.
	return bsonutil.WrapInLet(inputLetStatement,
		bsonutil.WrapInLet(paddedYearLetStatement,
			bsonutil.WrapInLet(outputLetStatement,
				bsonutil.WrapInCond(nil, output,
					bsonutil.WrapInOp(bsonutil.OpGt,
						bsonutil.WrapInOp(bsonutil.OpYear, output),
						9999))),
		)), nil

}

func (*makeDateFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*makeDateFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalInt64, EvalInt64}
	convertedExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*makeDateFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDate
}

func (*makeDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type md5Func struct{}

func (*md5Func) FuncName() string {
	return "md5"
}

var _ normalizingScalarFunc = (*md5Func)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/encryption-functions.html#function_md5
func (f *md5Func) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	h := md5.New()
	_, err := io.WriteString(h, values[0].String())
	if err != nil {
		return nil, err
	}
	return NewSQLVarchar(cfg.sqlValueKind, fmt.Sprintf("%x", h.Sum(nil))), nil
}

func (*md5Func) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*md5Func) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*md5Func) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type microsecondFunc struct{}

func (*microsecondFunc) FuncName() string {
	return "microsecond"
}

var _ normalizingScalarFunc = (*microsecondFunc)(nil)
var _ translatableToAggregationScalarFunc = (*microsecondFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_microsecond
func (f *microsecondFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {

	arg := values[0]

	if arg.IsNull() {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	str := arg.String()
	if str == "" {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	t, _, ok := parseTime(str)
	if !ok {
		return NewSQLInt64(cfg.sqlValueKind, 0), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.Nanosecond()/1000)), nil
}

func (*microsecondFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(microsecond)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMultiply, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem("$millisecond", args[0])),
			1000,
		)),
		), args[0],
	), nil

}

func (*microsecondFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*microsecondFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*microsecondFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type minuteFunc struct{}

func (*minuteFunc) FuncName() string {
	return "minute"
}

var _ normalizingScalarFunc = (*minuteFunc)(nil)
var _ translatableToAggregationScalarFunc = (*minuteFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_minute
func (f *minuteFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if values[0].IsNull() {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	t, hour, ok := parseTime(values[0].String())
	if !ok {
		// If we managed to parse, but minutes or seconds are >= 60
		// MySQL returns NULL for the hour/minute/second function.
		// Rather than return yet another value, we coop the hour value
		// and return -1, thus we can check for -1 here to return NULL
		// rather than the 0 expected if the string could not be parsed
		// at all.
		if hour == -1 {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
		}
		return NewSQLInt64(cfg.sqlValueKind, 0), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.Minute())), nil
}

func (*minuteFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(minute)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$minute", args[0]), nil
}

func (*minuteFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*minuteFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*minuteFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type modFunc struct {
	fun dualArgFloatMathFunc
}

func (*modFunc) FuncName() string {
	return "mod"
}

var _ translatableToAggregationScalarFunc = (*modFunc)(nil)

func (f *modFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.fun.Evaluate(ctx, cfg, st, values)
}

func (*modFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(mod)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.NewM(bsonutil.NewDocElem("$mod", bsonutil.NewArray(
		args[0],
		args[1],
	))), nil
}

func (f *modFunc) Normalize(kind SQLValueKind, e *SQLScalarFunctionExpr) SQLExpr {
	return f.fun.Normalize(kind, e)
}

func (f *modFunc) EvalType(exprs []SQLExpr) EvalType {
	return f.fun.EvalType(exprs)
}

func (f *modFunc) Validate(exprCount int) error {
	return f.fun.Validate(exprCount)
}

type monthFunc struct{}

func (*monthFunc) FuncName() string {
	return "month"
}

var _ reconcilingScalarFunc = (*monthFunc)(nil)
var _ translatableToAggregationScalarFunc = (*monthFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_month
func (f *monthFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.Month())), nil
}

func (*monthFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(month)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$month", args[0]), nil
}

func (*monthFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalDate)
}

func (*monthFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*monthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type monthNameFunc struct{}

func (*monthNameFunc) FuncName() string {
	return "monthName"
}

var _ reconcilingScalarFunc = (*monthNameFunc)(nil)
var _ translatableToAggregationScalarFunc = (*monthNameFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_monthname
func (f *monthNameFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	return NewSQLVarchar(cfg.sqlValueKind, t.Month().String()), nil
}

func (*monthNameFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(monthName)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpArrElemAt, bsonutil.NewArray(
				bsonutil.NewArray(
					time.January.String(),
					time.February.String(),
					time.March.String(),
					time.April.String(),
					time.May.String(),
					time.June.String(),
					time.July.String(),
					time.August.String(),
					time.September.String(),
					time.October.String(),
					time.November.String(),
					time.December.String(),
				),
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
					bsonutil.NewM(bsonutil.NewDocElem("$month", args[0])),
					1,
				))),
			)),
		), args[0],
	), nil
}

func (*monthNameFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalDate)
}

func (*monthNameFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*monthNameFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type multiArgFloatMathFunc struct {
	single singleArgFloatMathFunc
	dual   dualArgFloatMathFunc
}

func (multiArgFloatMathFunc) FuncName() string {
	return "multiArgFloatMath"
}

var _ normalizingScalarFunc = (*multiArgFloatMathFunc)(nil)

func (f multiArgFloatMathFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	if len(values) == 1 {
		return f.single.Evaluate(ctx, cfg, st, values)
	}
	return f.dual.Evaluate(ctx, cfg, st, values)
}

func (multiArgFloatMathFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (multiArgFloatMathFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDouble
}

func (multiArgFloatMathFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

type notFunc struct{}

func (*notFunc) FuncName() string {
	return "not"
}

func (f *notFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	matcher := &SQLNotExpr{values[0]}
	result, err := matcher.Evaluate(ctx, cfg, st)
	if err != nil {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), err
	}
	if Bool(result) {
		return NewSQLInt64(cfg.sqlValueKind, 1), nil
	}
	return NewSQLInt64(cfg.sqlValueKind, 0), nil
}

func (*notFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*notFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type nopushdownFunc struct{}

func (*nopushdownFunc) FuncName() string {
	return "nopushdown"
}

var _ reconcilingScalarFunc = (*nopushdownFunc)(nil)

func (*nopushdownFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return values[0], nil
}

func (*nopushdownFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	return f
}

func (*nopushdownFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return f
}

func (*nopushdownFunc) EvalType(exprs []SQLExpr) EvalType {
	return exprs[0].EvalType()
}

func (*nopushdownFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type nullifFunc struct{}

func (*nullifFunc) FuncName() string {
	return "nullif"
}

var _ normalizingScalarFunc = (*nullifFunc)(nil)
var _ reconcilingScalarFunc = (*nullifFunc)(nil)
var _ translatableToAggregationScalarFunc = (*nullifFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/control-flow-functions.html
func (f *nullifFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if values[0].IsNull() {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	} else if values[1].IsNull() {
		return values[0], nil
	} else {
		eq, _ := CompareTo(values[0], values[1], st.collation)
		if eq == 0 {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
		}
		return values[0], nil
	}
}

func (*nullifFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(nullif)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	if value, ok := bsonutil.GetLiteral(args[0]); ok {
		if value == nil {
			return bsonutil.MgoNullLiteral, nil
		}
		return bsonutil.WrapInCond(nil, args[0], bsonutil.WrapInOp(bsonutil.OpEq, args...)), nil
	}

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("expr", args[0]),
	)

	letEvaluation := bsonutil.WrapInCond(
		nil,
		"$$expr",
		bsonutil.WrapInNullCheck("$$expr"),
		bsonutil.WrapInOp(bsonutil.OpEq, "$$expr", args[1]),
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil
}

func (*nullifFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	firstVal, ok := f.Exprs[0].(SQLValue)
	if ok && firstVal.IsNull() {
		return NewSQLNull(kind, f.EvalType())
	}

	secondVal, ok := f.Exprs[1].(SQLValue)
	if ok && secondVal.IsNull() {
		return f.Exprs[0]
	}

	return f
}

func (*nullifFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return f
}

func (*nullifFunc) EvalType(exprs []SQLExpr) EvalType {
	return exprs[0].EvalType()
}

func (*nullifFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type padFunc struct {
	isLeftPad bool
	fun       scalarFunc
}

func (*padFunc) FuncName() string {
	return "pad"
}

var _ reconcilingScalarFunc = (*padFunc)(nil)
var _ translatableToAggregationScalarFunc = (*padFunc)(nil)

func (f *padFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return f.fun.Evaluate(ctx, cfg, st, values)
}

func (f *padFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {

	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(pad)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(pad)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	// arguments to lpad
	str := args[0]
	lengthVal := args[1]
	padStr := args[2]

	// round to nearest int.
	length := bsonutil.WrapInRound(lengthVal)

	// variables for $let expression - length of padding needed
	// and length of input padding strings
	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("padLen", bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
				length,
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, str)),
			)))),
		bsonutil.NewDocElem("padStrLen", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, padStr))),
		bsonutil.NewDocElem("length", length),
	)

	// logic for generating padding string:

	// do we even need to add padding? only if the desired output
	// length is > length of input string.
	paddingCond := bsonutil.NewM(
		bsonutil.NewDocElem(bsonutil.OpLt, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, str)),
			"$$length",
		)))

	// number of times we need to repeat the padding string to fill space
	padStrRepeats := bsonutil.NewM(
		bsonutil.NewDocElem(bsonutil.OpCeil, bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpDivide, bsonutil.NewArray(
				"$$padLen",
				"$$padStrLen",
			)))))

	// generate an array with padStrRepeats occurrences of padStr
	padParts := bsonutil.NewM(
		bsonutil.NewDocElem(bsonutil.OpMap, bsonutil.NewM(
			bsonutil.NewDocElem("input", bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpRange, bsonutil.NewArray(
					0,
					padStrRepeats,
				)),
			)),
			bsonutil.NewDocElem("in", padStr),
		)))

	// join occurrences together and trim to the exact length needed
	fullPad := bsonutil.NewM(
		bsonutil.NewDocElem(bsonutil.OpSubstr, bsonutil.NewArray(
			bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpReduce, bsonutil.NewM(
					bsonutil.NewDocElem("input", padParts),
					bsonutil.NewDocElem("initialValue", ""),
					bsonutil.NewDocElem("in", bsonutil.NewM(
						bsonutil.NewDocElem(bsonutil.OpConcat, bsonutil.NewArray(
							"$$value",
							"$$this",
						)))),
				))),
			0,
			"$$padLen",
		)),
	)

	// based on length of input string, we either add the padding
	// or just take appropriate substring of input string
	var concatted bson.M
	if f.isLeftPad {
		concatted = bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpConcat, bsonutil.NewArray(
			fullPad,
			str,
		)))
	} else {
		concatted = bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpConcat, bsonutil.NewArray(
			str,
			fullPad,
		)))
	}

	handleConcat := bsonutil.WrapInCond(
		nil,
		concatted, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
			"$$padStrLen",
			0,
		))))

	// handle everything in the case that input length >=0
	handleNonNegativeLength := bsonutil.WrapInCond(
		handleConcat, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubstr, bsonutil.NewArray(
			str,
			0,
			"$$length",
		))), paddingCond)

	// whether the input length is < 0
	lengthIsNegative := bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpLt, bsonutil.NewArray(
		length,
		0,
	)))

	// if it's < 0, then we just want to return null
	negativeCheck := bsonutil.WrapInCond(nil, handleNonNegativeLength, lengthIsNegative)

	return bsonutil.WrapInNullCheckedCond(
		nil,
		bsonutil.WrapInLet(letAssignment, negativeCheck),
		str, lengthVal, padStr,
	), nil
}

func (f *padFunc) Reconcile(e *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	if ReconcileFunc, ok := f.fun.(reconcilingScalarFunc); ok {
		return ReconcileFunc.Reconcile(e)
	}
	panic("Unreachable, lpad and rpad are bout reconciling")
}

func (f *padFunc) EvalType(exprs []SQLExpr) EvalType {
	return f.fun.EvalType(exprs)
}

func (f *padFunc) Validate(exprCount int) error {
	return f.fun.Validate(exprCount)
}

type powFunc struct{}

func (*powFunc) FuncName() string {
	return "pow"
}

var _ normalizingScalarFunc = (*powFunc)(nil)
var _ reconcilingScalarFunc = (*powFunc)(nil)
var _ translatableToAggregationScalarFunc = (*powFunc)(nil)

func (f *powFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	v0 := Float64(values[0])
	v1 := Float64(values[1])

	n := math.Pow(v0, v1)
	zeroBaseExpNeg := v0 == 0 && v1 < 0
	if math.IsNaN(n) || zeroBaseExpNeg {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))),
			mysqlerrors.Defaultf(mysqlerrors.ErDataOutOfRange,
				"DOUBLE",
				fmt.Sprintf("pow(%v,%v)",
					Float64(values[0]),
					Float64(values[1])))
	}

	return NewSQLFloat(cfg.sqlValueKind, math.Pow(v0, v1)), nil
}

func (f *powFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(pow)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInOp(bsonutil.OpPow, args[0], args[1]), nil
}

func (*powFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*powFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalDouble)
}

func (*powFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDouble
}

func (*powFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type quarterFunc struct{}

func (*quarterFunc) FuncName() string {
	return "quarter"
}

var _ reconcilingScalarFunc = (*quarterFunc)(nil)
var _ translatableToAggregationScalarFunc = (*quarterFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_quarter
func (f *quarterFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
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

	return NewSQLInt64(cfg.sqlValueKind, int64(q)), nil
}

func (*quarterFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(quarter)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	conds := bsonutil.NewArray()
	if _, ok := bsonutil.GetLiteral(args[0]); !ok {
		conds = append(conds, "$$date")
	}

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("date", args[0]),
	)

	letEvaluation := bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpArrElemAt, bsonutil.NewArray(
				bsonutil.NewArray(1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4),
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
					bsonutil.NewM(bsonutil.NewDocElem("$month", "$$date")),
					1,
				))),
			))), conds...,
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

func (*quarterFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalDate)
}

func (*quarterFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*quarterFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type randFunc struct{}

func (*randFunc) FuncName() string {
	return "rand"
}

var _ reconcilingScalarFunc = (*randFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_rand
func (*randFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	uniqueID := Uint64(values[0])

	if len(values) == 2 {
		seed := round(Float64(values[1]))
		r := st.RandomWithSeed(uniqueID, seed)
		return NewSQLFloat(cfg.sqlValueKind, r.Float64()), nil
	}

	r := st.Random(uniqueID)
	return NewSQLFloat(cfg.sqlValueKind, r.Float64()), nil
}

func (*randFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalDouble)
}

func (*randFunc) SkipConstantFolding() bool {
	return true
}

func (*randFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDouble
}

func (*randFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

// Documentation:
// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_radians
// Note: by embedding singleArgFloatMathFunc, radiansFunc inherits Evaluate,
// Reconcile, Type, and Validate implementations.
type radiansFunc struct {
	singleArgFloatMathFunc
}

func (*radiansFunc) FuncName() string {
	return "radians"
}

var _ translatableToAggregationScalarFunc = (*radiansFunc)(nil)

func (*radiansFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(radians)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInOp(bsonutil.OpDivide, bsonutil.WrapInOp(bsonutil.OpMultiply, args[0], math.Pi), 180.0), nil
}

type repeatFunc struct{}

func (*repeatFunc) FuncName() string {
	return "repeat"
}

var _ normalizingScalarFunc = (*repeatFunc)(nil)
var _ reconcilingScalarFunc = (*repeatFunc)(nil)
var _ translatableToAggregationScalarFunc = (*repeatFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_repeat
func (f *repeatFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (v SQLValue, err error) {
	if hasNullValue(values...) {
		v = NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values)))
		err = nil
		return
	}

	str := values[0].String()
	if len(str) < 1 {
		v = NewSQLVarchar(cfg.sqlValueKind, "")
		err = nil
		return
	}

	rep := int(roundToDecimalPlaces(0, Float64(values[1])))
	if rep < 1 {
		v = NewSQLVarchar(cfg.sqlValueKind, "")
		err = nil
		return
	}

	var b bytes.Buffer
	for i := 0; i < rep; i++ {
		b.WriteString(str)
	}

	v = NewSQLVarchar(cfg.sqlValueKind, b.String())
	err = nil
	return
}

func (*repeatFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(repeat)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(repeat)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	str := args[0]

	// num must be rounded to match mysql
	num := bsonutil.WrapInRound(args[1])

	// create array w/ args[1] values e.g. [0,1,2]
	rangeArr := bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpRange, bsonutil.NewArray(
		0,
		num,
		1,
	)))

	// create array of len arg[1], with each item being arg[0]
	mapArgs := bsonutil.NewM(bsonutil.NewDocElem("input", rangeArr), bsonutil.NewDocElem("in", str))
	mapWithArgs := bsonutil.NewM(bsonutil.NewDocElem("$map", mapArgs))

	// append all values of this array together
	inArg := bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpConcat, bsonutil.NewArray(
		"$$this",
		"$$value",
	)))
	reduceArgs := bsonutil.NewM(bsonutil.NewDocElem("input", mapWithArgs), bsonutil.NewDocElem("initialValue", ""), bsonutil.NewDocElem("in", inArg))

	repeat := bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpReduce, reduceArgs))

	return bsonutil.WrapInNullCheckedCond(nil, repeat, str, num), nil

}

func (*repeatFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*repeatFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalString, EvalDouble}
	convertedExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*repeatFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*repeatFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type replaceFunc struct{}

func (*replaceFunc) FuncName() string {
	return "replace"
}

var _ normalizingScalarFunc = (*replaceFunc)(nil)
var _ reconcilingScalarFunc = (*replaceFunc)(nil)
var _ translatableToAggregationScalarFunc = (*replaceFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_replace
func (f *replaceFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	s := values[0].String()
	old := values[1].String()
	new := values[2].String()

	return NewSQLVarchar(cfg.sqlValueKind, strings.Replace(s, old, new, -1)), nil
}

func (*replaceFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(replace)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(replace)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	split := "$$split"
	assignment := bsonutil.NewM(
		bsonutil.NewDocElem("split", bsonutil.WrapInOp(bsonutil.OpSplit, args[0], args[1])),
	)

	this, value := "$$this", "$$value"
	body := bsonutil.WrapInReduce(split,
		nil,
		bsonutil.WrapInCond(this,
			bsonutil.WrapInOp(bsonutil.OpConcat, value, args[2], this),
			bsonutil.WrapInOp(bsonutil.OpEq, value, nil),
		),
	)

	return bsonutil.WrapInLet(assignment, body), nil
}

func (*replaceFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*replaceFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalString, EvalString, EvalString}
	newExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		&replaceFunc{},
		newExprs,
	}
}

func (*replaceFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*replaceFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type reverseFunc struct{}

func (*reverseFunc) FuncName() string {
	return "reverse"
}

var _ normalizingScalarFunc = (*reverseFunc)(nil)
var _ reconcilingScalarFunc = (*reverseFunc)(nil)
var _ translatableToAggregationScalarFunc = (*reverseFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_reverse
func (rf *reverseFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, rf.EvalType(valsAsExprs(values))), nil
	}
	s := values[0].String()
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return NewSQLVarchar(cfg.sqlValueKind, string(runes)), nil
}

func (*reverseFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(reverse)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(reverse)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInCond(
		nil,
		bsonutil.WrapInLet(bsonutil.NewM(bsonutil.NewDocElem("input", args[0])), bsonutil.WrapInReduce(bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpRange, bsonutil.NewArray(
				0,
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, "$$input")),
			)),
		), "", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpConcat, bsonutil.NewArray(
			bsonutil.NewM(
				bsonutil.NewDocElem("$substrCP", bsonutil.NewArray(
					"$$input",
					"$$this",
					1,
				)),
			),
			"$$value",
		)),
		)),
		), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpLte, bsonutil.NewArray(
			args[0],
			nil,
		))),
	), nil
}

func (*reverseFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (rf *reverseFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalString}
	newExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		rf,
		newExprs,
	}
}

func (*reverseFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*reverseFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type rightFunc struct{}

func (*rightFunc) FuncName() string {
	return "right"
}

var _ reconcilingScalarFunc = (*rightFunc)(nil)
var _ translatableToAggregationScalarFunc = (*rightFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_right
func (f *rightFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	str := values[0].String()
	posFloat := Float64(values[1])

	if posFloat > float64(len(str)) {
		return NewSQLVarchar(cfg.sqlValueKind, str), nil
	}

	startPos := math.Min(0, -1.0*posFloat)

	substring,
		err := NewSQLScalarFunctionExpr("substring",
		[]SQLExpr{values[0],
			NewSQLFloat(cfg.sqlValueKind, startPos)})
	if err != nil {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), err
	}

	return substring.Evaluate(ctx, cfg, st)
}

func (*rightFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {

	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(right)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(right)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	conds := bsonutil.NewArray()
	var strLength, subStrLength interface{}

	if stringValue, ok := bsonutil.GetLiteral(args[0]); ok {
		// string is literal
		if stringValue == nil {
			return bsonutil.MgoNullLiteral, nil
		}

		if s, ok := stringValue.(string); ok {
			strLength = bsonutil.WrapInLiteral(len(s))
		} else {
			strLength = bsonutil.WrapInOp(bsonutil.OpStrlenCP, "$$string")
		}
	} else {
		// string is not a literal
		strLength = bsonutil.WrapInOp(bsonutil.OpStrlenCP, "$$string")
		conds = append(conds, "$$string")
	}

	if lengthValue, ok := bsonutil.GetLiteral(args[1]); ok {
		if lengthValue == nil {
			return bsonutil.MgoNullLiteral, nil
		}

		// when length is negative, just use 0. round length to closest integer
		if i, ok := lengthValue.(int64); ok {
			args[1] = bsonutil.WrapInLiteral(int64(math.Max(0, float64(i))))
			subStrLength = "$$length"
		} else {
			args[1] = bsonutil.WrapInRound(bsonutil.WrapInOp(bsonutil.OpMax, args[1], 0))
			subStrLength = "$$length"
		}
	} else {
		subStrLength = bsonutil.WrapInRound(bsonutil.WrapInOp(bsonutil.OpMax, "$$length", 0))
		conds = append(conds, "$$length")
	}

	// start = max(0, strLen - subStrLen)
	start := bsonutil.WrapInOp(bsonutil.OpMax, 0, bsonutil.WrapInOp(bsonutil.OpSubtract, strLength, subStrLength))

	subStrOp := bsonutil.WrapInOp(bsonutil.OpSubstr, "$$string", start, subStrLength)

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("string", args[0]),
		bsonutil.NewDocElem("length", args[1]),
	)

	letEvaluation := bsonutil.WrapInNullCheckedCond(nil, subStrOp, conds...)
	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil
}

func (*rightFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalString, EvalInt64}
	convertedExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*rightFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*rightFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type roundFunc struct{}

func (*roundFunc) FuncName() string {
	return "round"
}

var _ normalizingScalarFunc = (*roundFunc)(nil)
var _ reconcilingScalarFunc = (*roundFunc)(nil)
var _ translatableToAggregationScalarFunc = (*roundFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_round
func (f *roundFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	base := Float64(values[0])

	var decimal int64
	if len(values) == 2 {
		decimal = Int64(values[1])

		if decimal < 0 {
			return NewSQLFloat(cfg.sqlValueKind, 0), nil
		}
	} else {
		decimal = 0
	}

	rounded := roundToDecimalPlaces(decimal, base)

	return NewSQLFloat(cfg.sqlValueKind, rounded), nil
}

func (*roundFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) < 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(round)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs[0:1])
	if err != nil {
		return nil, err
	}
	switch len(exprs) {
	case 1:
		return bsonutil.WrapInRound(args[0]), nil
	case 2:
		if arg1, ok := exprs[1].(SQLValue); ok {
			return bsonutil.WrapInRoundWithPrecision(args[0], Float64(arg1)), nil
		}
		fallthrough
	default:
		return nil, newPushdownFailure("SQLScalarFunctionExpr(round)", "unsupported form")
	}
}

func (*roundFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*roundFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalDouble)
}

func (*roundFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDouble
}

func (*roundFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

type rpadFunc struct{}

func (*rpadFunc) FuncName() string {
	return "rpad"
}

var _ normalizingScalarFunc = (*rpadFunc)(nil)
var _ reconcilingScalarFunc = (*rpadFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_rpad
func (*rpadFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return handlePadding(cfg.sqlValueKind, values, false)
}

func (*rpadFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}
	return f
}

func (*rpadFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalString, EvalInt64, EvalString}
	convertedExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*rpadFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*rpadFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type rtrimFunc struct{}

func (*rtrimFunc) FuncName() string {
	return "rtrim"
}

var _ reconcilingScalarFunc = (*rtrimFunc)(nil)
var _ translatableToAggregationScalarFunc = (*rtrimFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_rtrim
func (f *rtrimFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	value := strings.TrimRight(values[0].String(), " ")

	return NewSQLVarchar(cfg.sqlValueKind, value), nil
}

func (*rtrimFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(rtrim)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(rtrim)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	if t.versionAtLeast(4, 0, 0) {
		return bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpRTrim, bsonutil.NewM(
				bsonutil.NewDocElem("input", args[0]),
				bsonutil.NewDocElem("chars", " "),
			)),
		), nil
	}

	rtrimCond := bsonutil.WrapInCond(
		"",
		bsonutil.WrapInLRTrim(false, args[0]), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
			args[0],
			"",
		))))

	return bsonutil.WrapInNullCheckedCond(
		nil,
		rtrimCond,
		args[0],
	), nil
}

func (*rtrimFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalString)
}

func (*rtrimFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*rtrimFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type secondFunc struct{}

func (*secondFunc) FuncName() string {
	return "second"
}

var _ normalizingScalarFunc = (*secondFunc)(nil)
var _ translatableToAggregationScalarFunc = (*secondFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_second
func (f *secondFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if values[0].IsNull() {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	t, hour, ok := parseTime(values[0].String())
	if !ok {
		// If we managed to parse, but minutes or seconds are >= 60
		// MySQL returns NULL for the hour/minute/second function.
		// Rather than return yet another value, we coop the hour value
		// and return -1, thus we can check for -1 here to return NULL
		// rather than the 0 expected if the string could not be parsed
		// at all.
		if hour == -1 {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
		}
		return NewSQLInt64(cfg.sqlValueKind, 0), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.Second())), nil
}

func (*secondFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(second)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$second", args[0]), nil
}

func (*secondFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*secondFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*secondFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type signFunc struct{}

func (*signFunc) FuncName() string {
	return "sign"
}

var _ normalizingScalarFunc = (*signFunc)(nil)
var _ reconcilingScalarFunc = (*signFunc)(nil)
var _ translatableToAggregationScalarFunc = (*signFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_sign
func (sf *signFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, sf.EvalType(valsAsExprs(values))), nil
	}

	v := Float64(values[0])
	// Positive numbers are more common than negative in most data sets.
	if v > 0 {
		return NewSQLInt64(cfg.sqlValueKind, 1), nil
	}
	if v < 0 {
		return NewSQLInt64(cfg.sqlValueKind, -1), nil
	}
	return NewSQLInt64(cfg.sqlValueKind, 0), nil
}

func (*signFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(sign)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInCond(nil,
		bsonutil.WrapInCond(bsonutil.WrapInLiteral(0),
			bsonutil.WrapInCond(bsonutil.WrapInLiteral(1),
				bsonutil.WrapInLiteral(-1), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGt, bsonutil.NewArray(
					args[0],
					bsonutil.WrapInLiteral(0),
				))),
			), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
				args[0],
				bsonutil.WrapInLiteral(0),
			))),
		), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpLte, bsonutil.NewArray(
			args[0],
			nil,
		))),
	), nil
}

func (*signFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (sf *signFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalDouble}
	newExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		sf,
		newExprs,
	}
}

func (*signFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*signFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_sin
type sinFunc struct {
	singleArgFloatMathFunc
}

func (*sinFunc) FuncName() string {
	return "sin"
}

var _ translatableToAggregationScalarFunc = (*sinFunc)(nil)

func (*sinFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(sin)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	input, absInput := "$$input", "$$absInput"
	inputLetAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("input", args[0]),
	)

	absInputLetAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("absInput", bsonutil.WrapInOp(bsonutil.OpAbs, input)),
	)

	rem, phase := "$$rem", "$$phase"
	remPhaseAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("rem", bsonutil.WrapInOp(bsonutil.OpMod, absInput, math.Pi/2)),
		bsonutil.NewDocElem("phase", bsonutil.WrapInOp(bsonutil.OpMod,
			bsonutil.WrapInOp(bsonutil.OpTrunc,
				bsonutil.WrapInOp(bsonutil.OpDivide, absInput, math.Pi/2),
			),
			4.0)),
	)

	// 3.2 does not support $switch, so just use chained $cond, assuming
	// zeroCase will be most common (since it's the first phase) Because we
	// use the Maclaurin Power Series for sin and cos, we need to adjust
	// our input into a domain that is good for our approximation, that
	// being the first quadrant (phase). For phases outside of the first,
	// we can adjust the functions as:
	//
	// phase | Maclaurin Power Series
	// ------------------------------
	// 0     | sin(rem)
	// 1     | cos(rem)
	// 2     | -1 * sin(rem)
	// 3     | -1 * cos(rem)
	// where the phase is defined as the trunc(input / (pi/2)) % 4
	// and the remainder is input % (pi/2).

	threeCase := bsonutil.WrapInCond(bsonutil.WrapInOp(bsonutil.OpMultiply,
		-1.0,
		bsonutil.WrapInCosPowerSeries(rem)),
		nil,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			3))
	twoCase := bsonutil.WrapInCond(bsonutil.WrapInOp(bsonutil.OpMultiply,
		-1.0,
		bsonutil.WrapInSinPowerSeries(rem)),
		threeCase,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			2))
	oneCase := bsonutil.WrapInCond(bsonutil.WrapInCosPowerSeries(rem),
		twoCase,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			1))
	zeroCase := bsonutil.WrapInCond(bsonutil.WrapInSinPowerSeries(rem),
		oneCase,
		bsonutil.WrapInOp(bsonutil.OpEq,
			phase,
			0))

	// cos(-x) = cos(x), but sin(-x) = -sin(x), so if the original input is negative multiply by -1.
	return bsonutil.WrapInLet(inputLetAssignment,
		bsonutil.WrapInLet(absInputLetAssignment,
			bsonutil.WrapInLet(remPhaseAssignment,
				bsonutil.WrapInCond(zeroCase,
					bsonutil.WrapInOp(bsonutil.OpMultiply, -1.0, zeroCase),
					bsonutil.WrapInOp(bsonutil.OpGte, input, 0),
				),
			),
		),
	), nil
}

type singleArgFloatMathFunc func(float64) float64

var _ normalizingScalarFunc = (*singleArgFloatMathFunc)(nil)
var _ reconcilingScalarFunc = (*singleArgFloatMathFunc)(nil)

func (singleArgFloatMathFunc) FuncName() string {
	return "singleArgFloatMath"
}

func (f singleArgFloatMathFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	result := f(Float64(values[0]))
	if math.IsNaN(result) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}
	if math.IsInf(result, 0) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}
	if result == -0 {
		result = 0
	}
	return NewSQLFloat(cfg.sqlValueKind, result), nil
}

func (singleArgFloatMathFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (singleArgFloatMathFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	switch f.Name {
	case "abs", "ceil", "exp", "degrees", "floor", "ln", "log", "log10", "log2", "radians", "sqrt":
		return convertAllArgs(f, EvalDouble)
	default:
		return f
	}
}

func (singleArgFloatMathFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDouble
}

func (singleArgFloatMathFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type sleepFunc struct{}

func (*sleepFunc) FuncName() string {
	return "sleep"
}

// https://dev.mysql.com/doc/refman/5.7/en/miscellaneous-functions.html#function_sleep
func (*sleepFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {

	err := mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, "sleep")

	if hasNullValue(values...) {
		return nil, err
	}

	n := Float64(values[0])

	if n < 0 {
		return nil, err
	}

	timer := time.NewTimer(time.Second * time.Duration(n))

	select {
	case <-timer.C:
	case <-ctx.Done():
		timer.Stop()
		return nil, ctx.Err()
	}

	return NewSQLInt64(cfg.sqlValueKind, 0), nil
}

func (*sleepFunc) SkipConstantFolding() bool {
	return true
}

func (*sleepFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*sleepFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type spaceFunc struct{}

func (*spaceFunc) FuncName() string {
	return "space"
}

var _ reconcilingScalarFunc = (*spaceFunc)(nil)
var _ translatableToAggregationScalarFunc = (*spaceFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_space
func (f *spaceFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	flt := Float64(values[0])
	n := round(flt)
	if n < 1 {
		return NewSQLVarchar(cfg.sqlValueKind, ""), nil
	}

	return NewSQLVarchar(cfg.sqlValueKind, strings.Repeat(" ", int(n))), nil
}

func (*spaceFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(space)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(space)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	n := "$$n"
	return bsonutil.WrapInLet(bsonutil.NewM(bsonutil.NewDocElem("n", bsonutil.WrapInRound(args[0]))),
		bsonutil.WrapInCond(nil,
			bsonutil.WrapInReduce(bsonutil.WrapInRange(0, n, 1),
				"",
				bsonutil.WrapInOp(bsonutil.OpConcat, "$$value", " "),
			),
			bsonutil.WrapInOp(bsonutil.OpLte, n, nil),
		),
	), nil
}

func (*spaceFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalInt64}
	convertedExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*spaceFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*spaceFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type sqrtFunc struct {
	singleArgFloatMathFunc
}

func (*sqrtFunc) FuncName() string {
	return "sqrt"
}

var _ translatableToAggregationScalarFunc = (*sqrtFunc)(nil)

func (*sqrtFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(sqrt)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapInCond(bsonutil.NewM(bsonutil.NewDocElem("$sqrt", args[0])), nil, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGte, bsonutil.NewArray(
		args[0],
		0,
	)))), nil
}

type strToDateFunc struct{}

func (*strToDateFunc) FuncName() string {
	return "strToDate"
}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_str-to-date
func (f *strToDateFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	str, ok := values[0].(SQLVarchar)
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	ft, ok := values[1].(SQLVarchar)
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	s := str.String()
	ftStr := ft.String()
	fmtTokens := map[string]string{
		"%a": "Mon",
		"%b": "Jan",
		"%c": "1",
		"%d": "02",
		"%e": "2",
		"%H": "15",
		"%i": "04",
		"%k": "13",
		"%M": "January",
		"%m": "01",
		"%S": "05",
		"%s": "05",
		"%T": "15:04:05",
		"%W": "Monday",
		"%w": "Mon",
		"%Y": "2006",
		"%y": "06",
	}

	format := ""
	skipToken := false
	ts := false
	for idx, char := range ftStr {
		if !skipToken {
			if char == 37 && idx != len(ftStr)-1 {
				token := "%" + string(ftStr[idx+1])
				skipToken = true
				goToken := fmtTokens[token]
				if goToken != "" {
					format += goToken
				} else {
					return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
				}
				if token == "%H" || token == "%i" || token == "%k" || token == "%p" ||
					token == "%S" || token == "%s" || token == "%T" {
					ts = true
				}
			} else {
				format += string(char)
			}
		} else {
			skipToken = false
		}
	}

	d, err := time.Parse(format, s)
	if err != nil {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	if ts {
		return NewSQLTimestamp(cfg.sqlValueKind, d), nil
	}

	return NewSQLDate(cfg.sqlValueKind, d), nil
}

func (*strToDateFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDatetime
}

func (*strToDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type subDateFunc struct{}

func (*subDateFunc) FuncName() string {
	return "subDate"
}

var _ normalizingScalarFunc = (*subDateFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_subdate
func (*subDateFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	subtractor := &dateSubFunc{}
	return subtractor.Evaluate(ctx, cfg, st, values)
}

func (*subDateFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*subDateFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDatetime
}

func (*subDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2, 3)
}

type substringFunc struct {
	isMid bool
}

func (*substringFunc) FuncName() string {
	return "substring"
}

var _ normalizingScalarFunc = (*substringFunc)(nil)
var _ reconcilingScalarFunc = (*substringFunc)(nil)
var _ translatableToAggregationScalarFunc = (*substringFunc)(nil)

func (f *substringFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	str := []rune(values[0].String())
	if values[0].String() == "" {
		return NewSQLVarchar(cfg.sqlValueKind, ""), nil
	}

	posFloat := Float64(values[1])
	var pos int
	if posFloat >= 0 {
		pos = int(posFloat + 0.5)
	} else {
		pos = int(posFloat - 0.5)
	}

	if pos > len(str) || pos == 0 {
		return NewSQLVarchar(cfg.sqlValueKind, ""), nil
	} else if pos < 0 {
		pos = len(str) + pos

		if pos < 0 {
			return NewSQLVarchar(cfg.sqlValueKind, ""), nil
		}
	} else {
		pos-- // MySQL uses 1 as a basis
	}

	if len(values) == 3 {
		length := int(Float64(values[2]) + 0.5)
		if length < 1 {
			return NewSQLVarchar(cfg.sqlValueKind, ""), nil
		}
		if pos <= len(str) {
			str = str[pos:]
		}
		if length <= len(str) {
			str = str[:length]
		}
	} else {
		if pos < len(str) {
			str = str[pos:]
		}
	}
	return NewSQLVarchar(cfg.sqlValueKind, string(str)), nil
}

func (f *substringFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(substring)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if (len(exprs) != 2 && len(exprs) != 3) ||
		(len(exprs) == 2 && f.isMid) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(substring)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}
	strVal := args[0]
	indexVal := args[1]

	var lenVal interface{}
	if len(args) == 3 {
		lenVal = args[2]
	} else {
		lenVal = bsonutil.NewM(bsonutil.NewDocElem("$strLenCP", args[0]))
	}

	indexNegVal := bsonutil.WrapInLet(bsonutil.NewM(
		bsonutil.NewDocElem("indexValNeg", bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpAdd, bsonutil.NewArray(
					bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMultiply, bsonutil.NewArray(
						indexVal,
						-1,
					))),
					0.5,
				))))))), bsonutil.WrapInCond(bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, strVal)),
		"$$indexValNeg",
	))), "$$indexValNeg", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGte, bsonutil.NewArray(
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, strVal)),
		"$$indexValNeg",
	)))))

	indexPosVal := bsonutil.NewM(
		bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
			bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.NewM(
					bsonutil.NewDocElem(bsonutil.OpAdd, bsonutil.NewArray(
						indexVal,
						0.5,
					)))),
			),
			1,
		)))

	roundOffIndex := bsonutil.WrapInCond(
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpAdd, bsonutil.NewArray(
			indexVal,
			0.5))))),
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpAdd, bsonutil.NewArray(
			indexVal,
			-0.5))))),
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGte, bsonutil.NewArray(
			indexVal,
			0,
		))))

	indexValBSONM := bsonutil.WrapInLet(
		bsonutil.NewM(bsonutil.NewDocElem("roundOffIndex", roundOffIndex)),
		bsonutil.WrapInCond(
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, strVal)),
			bsonutil.WrapInCond(
				indexPosVal,
				indexNegVal, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGt,
					bsonutil.NewArray(
						"$$roundOffIndex",
						0,
					)))), bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
					"$$roundOffIndex",
					0,
				))),
		))

	lenValBSONM := bsonutil.WrapInCond(
		0,
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpAdd, bsonutil.NewArray(
			lenVal,
			0.5,
		))))), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpLte, bsonutil.NewArray(
			lenVal,
			0,
		))),
	)

	return bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(bsonutil.NewDocElem("$substrCP", bsonutil.NewArray(
			strVal,
			indexValBSONM,
			lenValBSONM,
		))), strVal, indexVal, lenVal,
	), nil
}

func (*substringFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*substringFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalString, EvalInt64, EvalInt64}
	convertedExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*substringFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*substringFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2, 3)
}

type substringIndexFunc struct{}

func (*substringIndexFunc) FuncName() string {
	return "substringIndex"
}

var _ normalizingScalarFunc = (*substringIndexFunc)(nil)
var _ reconcilingScalarFunc = (*substringIndexFunc)(nil)
var _ translatableToAggregationScalarFunc = (*substringIndexFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_substring-index
func (sif *substringIndexFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {

	matches := func(r []rune, pos int, delim []rune) bool {
		for i, j := pos, 0; i < len(r) && j < len(delim); i, j = i+1, j+1 {
			if r[i] != delim[j] {
				return false
			}
		}
		return true
	}

	reverse := func(r []rune) {
		for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
	}

	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, sif.EvalType(valsAsExprs(values))), nil
	}

	r := []rune(values[0].String())
	delim := []rune(values[1].String())

	count := int(round(Float64(values[2])))

	if count == 0 {
		return NewSQLVarchar(cfg.sqlValueKind, ""), nil
	}

	reversed := count < 0
	if reversed {
		reverse(r)
		reverse(delim)

		count = -count
	}

	matchCount := 0
	i := 0
	for {
		if matches(r, i, delim) {
			matchCount++
			i += len(delim)
		} else {
			i++
		}

		if i >= len(r) || matchCount >= count {
			break
		}
	}

	if matchCount >= count {
		r = r[:i-len(delim)]
	}

	if reversed {
		reverse(r)
	}

	return NewSQLVarchar(cfg.sqlValueKind, string(r)), nil
}

func (*substringIndexFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(substringIndex)",
			"cannot push down to MongoDB < 3.4",
		)
	}

	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(substringIndex)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	delim, split := "$$delim", "$$split"
	inputAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("delim", args[1]),
	)

	splitAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("split", bsonutil.WrapInOp(bsonutil.OpSlice,
			bsonutil.WrapInOp(bsonutil.OpSplit, args[0], delim),
			bsonutil.WrapInRound(args[2]),
		)),
	)

	this, value := "$$this", "$$value"
	body := bsonutil.WrapInReduce(split,
		nil,
		bsonutil.WrapInCond(this,
			bsonutil.WrapInOp(bsonutil.OpConcat, value, delim, this),
			bsonutil.WrapInOp(bsonutil.OpEq, value, nil),
		),
	)

	return bsonutil.WrapInLet(inputAssignment,
		bsonutil.WrapInLet(splitAssignment,
			body,
		),
	), nil
}

func (*substringIndexFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	if v, ok := f.Exprs[2].(SQLValue); ok {
		if Int64(v) == 0 {
			return NewSQLVarchar(kind, "")
		}
	}

	return f
}

func (sif *substringIndexFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalString, EvalString, EvalInt64}
	newExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		sif,
		newExprs,
	}
}

func (*substringIndexFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*substringIndexFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type tanFunc struct {
	singleArgFloatMathFunc
}

func (*tanFunc) FuncName() string {
	return "tan"
}

var _ translatableToAggregationScalarFunc = (*tanFunc)(nil)

func (*tanFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	sf := &sinFunc{}
	num, err := sf.FuncToAggregationLanguage(t, []SQLExpr{exprs[0]})
	if err != nil {
		return nil, err
	}

	cf := &cosFunc{}
	denom, err := cf.FuncToAggregationLanguage(t, []SQLExpr{exprs[0]})
	if err != nil {
		return nil, err
	}

	// epsilon the smallest value we allow for denom, computed to roughly
	// tie-out with mysqld.
	epsilon := 6.123233995736766e-17
	// Replace abs(denom) < epsilon with epsilon since mysql doesn't seem
	// to return NULL or INF for inf values of tan as one would expect, and
	// we do not want to trigger a $divide by 0.
	return bsonutil.WrapInOp(bsonutil.OpDivide,
		num,
		bsonutil.WrapInCond(epsilon,
			denom,
			bsonutil.WrapInOp(bsonutil.OpLte,
				bsonutil.WrapInOp(bsonutil.OpAbs, denom), epsilon,
			),
		),
	), nil
}

type timeDiffFunc struct{}

func (*timeDiffFunc) FuncName() string {
	return "timeDiff"
}

var _ normalizingScalarFunc = (*timeDiffFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_timediff
func (f *timeDiffFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	expr1, _, ok := parseTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	expr2, _, ok := parseTime(values[1].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	d := expr1.Sub(expr2)

	fmtDef := func(buf []byte, v int) int {
		def := []byte{':', '0', '0', ':', '0', '0'}
		w := len(buf)
		for _, b := range def[v:] {
			w--
			buf[w] = b
		}
		return w
	}

	// Precision is exactly 6 digits (microsecond range)
	fmtFrac := func(buf []byte, v uint64) (nw int, nv uint64) {

		w, print := len(buf), false
		end := len(buf) - 1

		for i := 0; i < 6; i++ {

			digit := v % 10

			print = print || digit != 0

			if print {
				buf[end-i] = byte(digit) + '0'
				w--
			}
			v /= 10
		}

		if w != len(buf) {

			// If we have printed anything, pad the rest of the range with zeroes
			for len(buf)-w < 6 {
				buf[end] = '0'
				end--
				w--
			}
			w--
			buf[w] = '.'
		}

		return w, v
	}

	fmtInt := func(buf []byte, v uint64) int {
		w := len(buf)
		if v == 0 {
			w--
			buf[w] = '0'
		} else {
			for v > 0 {
				w--
				buf[w] = byte(v%10) + '0'
				v /= 10
			}
		}
		for len(buf)-w < 2 {
			w--
			buf[w] = '0'
		}
		return w
	}

	u := uint64(d)

	if u == 0 {
		return NewSQLVarchar(cfg.sqlValueKind, "00:00:00.000000"), nil
	}

	buf := [30]byte{}
	w := len(buf)

	neg := d < 0
	if neg {
		u = -u
	}

	// Shave off nanosecond precision (we need up to microseconds)
	u /= 1000

	// Handle fractional portion (< 1 second)
	w, u = fmtFrac(buf[:w], u)

	if u < uint64(time.Microsecond) {

		w = fmtInt(buf[:w], u)
		w = fmtDef(buf[:w], 0)

	} else {

		// u is now integer seconds
		w = fmtInt(buf[:w], u%60)
		u /= 60

		w--
		buf[w] = ':'

		// u is now integer minutes
		if u > 0 {
			w = fmtInt(buf[:w], u%60)
			u /= 60

			w--
			buf[w] = ':'

			// u is now integer hours
			if u > 0 {
				w = fmtInt(buf[:w], u)
			} else {
				w = fmtDef(buf[:w], 4)
			}
		} else {
			w = fmtDef(buf[:w], 1)
		}
	}

	if neg {
		w--
		buf[w] = '-'
	}

	return NewSQLVarchar(cfg.sqlValueKind, string(buf[w:])), nil
}

func (*timeDiffFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*timeDiffFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*timeDiffFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type timeToSecFunc struct{}

func (*timeToSecFunc) FuncName() string {
	return "timeToSec"
}

var _ normalizingScalarFunc = (*timeToSecFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_time-to-sec
func (f *timeToSecFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	dateParts := strings.Split(values[0].String(), " ")
	hasDatePart := len(dateParts) == 2
	components := strings.Split(values[0].String(), ":")
	if hasDatePart {
		components = strings.Split(dateParts[1], ":")
	}

	result, componentized := 0.0, true

	if len(components) == 1 {
		cmp, err := strconv.ParseFloat(components[0], 64)
		if err != nil {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), err
		}

		component := strconv.FormatFloat(math.Trunc(cmp), 'f', -1, 64)

		l := len(component)
		components, componentized = []string{"0", "0", "0"}, false

		// MySQL interprets abbreviated values without colons using the
		// assumption that the two rightmost digits represent seconds.
		switch l {
		case 1, 2:
			components[2] = component
		case 3, 4:
			components[1],
				components[2] = component[:l-2],
				component[l-2:l]
		case 5:
			components[0],
				components[1],
				components[2] = component[:l-4],
				component[l-4:l-2],
				component[l-2:l]
		default:
			components[0],
				components[1],
				components[2] = component[:l-4],
				component[l-4:l-2],
				component[l-2:l]
		}
	}

	signBit := false

	for i := 0; i < 3 && i < len(components); i++ {
		component, err := strconv.ParseFloat(components[i], 64)
		if err != nil {
			return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), err
		}

		cmp := math.Trunc(component)

		switch i {
		// more on valid time types at https://dev.mysql.com/doc/refman/5.7/en/time.html
		case 0:
			if cmp > 838 || cmp < -838 {
				if !componentized {
					return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
				}
				cmp = math.Copysign(838.0, cmp)
				components = []string{"", "59", "59"}
			}
		default:
			if cmp > 59 {
				return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
			}
		}

		signBit = signBit || math.Signbit(cmp)
		result += math.Abs(cmp) * (3600.0 / (math.Pow(60, float64(i))))
	}

	if signBit {
		return NewSQLFloat(cfg.sqlValueKind, math.Copysign(result, -1)), nil
	}

	return NewSQLFloat(cfg.sqlValueKind, result), nil
}

func (*timeToSecFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*timeToSecFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDouble
}

func (*timeToSecFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type timestampAddFunc struct{}

func (*timestampAddFunc) FuncName() string {
	return "timestampAdd"
}

var _ reconcilingScalarFunc = (*timestampAddFunc)(nil)
var _ translatableToAggregationScalarFunc = (*timestampAddFunc)(nil)

//http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_timestampadd
func (f *timestampAddFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	v := int(round(Float64(values[1])))
	// values[2] must be a SQLTimestamp or the algebrizer has been broken.
	// Check the handling of scalar functions in the algebrizer.
	tmp, ok := values[2].(SQLTimestamp)
	if !ok {
		return nil,
			fmt.Errorf("unable to evaluate type %v in timestampadd,"+
				" this points to an error in the algebrizer",
				values[2])
	}
	t := Timestamp(tmp)

	ts := false
	if len(values[2].String()) > 10 {
		ts = true
	}

	switch values[0].String() {
	case Year:
		return NewSQLTimestamp(cfg.sqlValueKind, t.AddDate(v, 0, 0)), nil
	case Quarter:
		y, mp, d := t.Date()
		m := int(mp)
		interval := v * 3
		y += (m + interval - 1) / 12
		m = (m+interval-1)%12 + 1
		switch m {
		case 2:
			if isLeapYear(y) {
				d = mathutil.MinInt(d, 29)
			} else {

				d = mathutil.MinInt(d, 28)
			}
		case 4, 6, 9, 11:
			d = mathutil.MinInt(d, 30)
		}
		if ts {
			return NewSQLTimestamp(cfg.sqlValueKind, time.Date(y,
					time.Month(m),
					d,
					t.Hour(),
					t.Minute(),
					t.Second(),
					t.Nanosecond(),
					schema.DefaultLocale)),
				nil
		}
		return NewSQLTimestamp(cfg.sqlValueKind, time.Date(y,
				time.Month(m),
				d,
				0,
				0,
				0,
				0,
				schema.DefaultLocale)),
			nil
	case Month:
		y, mp, d := t.Date()
		m := int(mp)
		interval := v
		y += (m + interval - 1) / 12
		m = (m+interval-1)%12 + 1
		switch m {
		case 2:
			if isLeapYear(y) {
				d = mathutil.MinInt(d, 29)
			} else {

				d = mathutil.MinInt(d, 28)
			}
		case 4, 6, 9, 11:
			d = mathutil.MinInt(d, 30)
		}
		if ts {
			return NewSQLTimestamp(cfg.sqlValueKind, time.Date(y,
					time.Month(m),
					d,
					t.Hour(),
					t.Minute(),
					t.Second(),
					t.Nanosecond(),
					schema.DefaultLocale)),
				nil
		}
		return NewSQLTimestamp(cfg.sqlValueKind, time.Date(y,
				time.Month(m),
				d,
				0,
				0,
				0,
				0,
				schema.DefaultLocale)),
			nil
	case Week:
		return NewSQLTimestamp(cfg.sqlValueKind, t.AddDate(0, 0, v*7)), nil
	case Day:
		return NewSQLTimestamp(cfg.sqlValueKind, t.AddDate(0, 0, v)), nil
	case Hour:
		duration := time.Duration(v) * time.Hour
		return NewSQLTimestamp(cfg.sqlValueKind, t.Add(duration)), nil
	case Minute:
		duration := time.Duration(v) * time.Minute
		return NewSQLTimestamp(cfg.sqlValueKind, t.Add(duration)), nil
	case Second:
		// Seconds can actually be fractional rather than integer.
		duration := time.Duration(int64(Float64(values[1]) * 1e9))
		return NewSQLTimestamp(cfg.sqlValueKind, t.Add(duration)), nil
	case Microsecond:
		duration := time.Duration(int64(Float64(values[1]))) * time.Microsecond
		return NewSQLTimestamp(cfg.sqlValueKind, t.Add(duration).Round(time.Millisecond)), nil
	default:
		err := fmt.Errorf("cannot add '%v' to timestamp", values[0])
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), err
	}
}

func (*timestampAddFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 5, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(timestampAdd)",
			"cannot push down to MongoDB < 3.5",
		)
	}

	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(timestampAdd)",
			incorrectArgCountMsg,
		)
	}

	unit := exprs[0].String()
	args, err := t.translateArgs(exprs[1:])
	if err != nil {
		return nil, err
	}
	interval := args[0]

	timestampExpr := args[1]
	// This is a very large and costly expression, make sure to bind it in
	// a let (in the switch at the end of the function).
	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("timestampArg", timestampExpr),
	)

	// Use timestampArg to refer to $$timestampArg below, referencing the var defined above.
	timestampArg := "$$timestampArg"

	// handleSimpleCase generates code for cases where we do not need to
	// use $dateFromParts, we just round the interval if the round argument
	// is true, and multiply by the number of milliseconds corresponded to
	// by 'u' then add to the timestamp.
	handleSimpleCase := func(u string, round bool) interface{} {
		if round {
			return bsonutil.WrapInOp(bsonutil.OpAdd,
				timestampArg,
				bsonutil.WrapInOp(bsonutil.OpMultiply,
					bsonutil.WrapInRound(interval),
					toMilliseconds[u]))
		}
		return bsonutil.WrapInOp(bsonutil.OpAdd,
			timestampArg,
			bsonutil.WrapInOp(bsonutil.OpMultiply,
				interval,
				toMilliseconds[u]))
	}

	// handleDateFromPartsCase handles cases where we need to use
	// $dateFromParts because we want to add a Year, a Month, or 3 Months
	// (a Quarter) to the specific date part.
	handleDateFromPartsCase := func(u string) interface{} {
		// Start with the equations for Quarter/Month, since they are
		// the same. They use a shared computation part
		// (sharedComputation) that changes based on if this is a
		// Quarter or Month.
		sharedComputation := "$$sharedComputation"
		newYear, newMonth := "$$newYear", "$$newMonth"
		dayExpr := bsonutil.WrapInOp(bsonutil.OpDayOfMonth, timestampArg)
		// This template is used in a call to $dateFromParts.
		// The Year case modifies part of the template.
		template := bsonutil.NewM(
			bsonutil.NewDocElem("year", "$$newYear"),
			bsonutil.NewDocElem("month", "$$newMonth"),
			bsonutil.NewDocElem("day",
				// The following MongoDB aggregation language implements this go code,
				// the goal of which is to keep days from overflowing when adding
				// Quarters or Months.
				// switch m {
				// case 2:
				// 	if isLeapYear(y) {
				// 		d = mathutil.MinInt(d, 29)
				//	} else {
				//		d = mathutil.MinInt(d, 28)
				//	}
				// case 4, 6, 9, 11:
				//	d = mathutil.MinInt(d, 30)
				// }
				// otherwise d is left unchanged as the day of the input timestamp.
				bsonutil.WrapInSwitch(bsonutil.WrapInOp(bsonutil.OpDayOfMonth, timestampArg),
					bsonutil.WrapInEqCase(newMonth, 2,
						bsonutil.WrapInCond(bsonutil.WrapInOp(bsonutil.OpMin, dayExpr, 29),
							bsonutil.WrapInOp(bsonutil.OpMin, dayExpr, 28),
							bsonutil.WrapInIsLeapYear(newYear)),
					),
					bsonutil.WrapInEqCase(newMonth, 4,
						bsonutil.WrapInOp(bsonutil.OpMin, dayExpr, 30)),
					bsonutil.WrapInEqCase(newMonth, 6,
						bsonutil.WrapInOp(bsonutil.OpMin, dayExpr, 30)),
					bsonutil.WrapInEqCase(newMonth, 9,
						bsonutil.WrapInOp(bsonutil.OpMin, dayExpr, 30)),
					bsonutil.WrapInEqCase(newMonth, 11,
						bsonutil.WrapInOp(bsonutil.OpMin, dayExpr, 30)),
				)),
			bsonutil.NewDocElem("hour", bsonutil.WrapInOp(bsonutil.OpHour, timestampArg)),
			bsonutil.NewDocElem("minute", bsonutil.WrapInOp(bsonutil.OpMinute, timestampArg)),
			bsonutil.NewDocElem("second", bsonutil.WrapInOp(bsonutil.OpSecond, timestampArg)),
			bsonutil.NewDocElem("millisecond", bsonutil.WrapInOp(bsonutil.OpMillisecond, timestampArg)),
		)

		var sharedComputationLetAssignment interface{}
		var newYearMonthLetAssignment interface{}
		switch u {
		case Year:
			// For Year intervals, the year, month, and day use
			// different, simpler equations. Keep everything but
			// year, to year we add the rounded interval. There is
			// no SharedComputation part, so we do not bsonutil.WrapInLet.
			// Note that the rest of the template is maintained.
			template["year"] = bsonutil.WrapInOp(bsonutil.OpAdd,
				bsonutil.WrapInRound(interval),
				bsonutil.WrapInOp(bsonutil.OpYear,
					timestampArg))
			template["month"] = bsonutil.WrapInOp(bsonutil.OpMonth,
				timestampArg)
			template["day"] = bsonutil.WrapInOp(bsonutil.OpDayOfMonth,
				timestampArg)
			return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDateFromParts, template))
		// For Quarter and Month intervals, only the SharedComputation
		// part changes.
		case Quarter:
			// SharedComputation = Month + round(interval) * 3 - 1.
			sharedComputationLetAssignment = bsonutil.NewM(
				bsonutil.NewDocElem("sharedComputation", bsonutil.WrapInOp(bsonutil.OpSubtract,
					bsonutil.WrapInOp(bsonutil.OpAdd,
						bsonutil.WrapInOp(bsonutil.OpMonth, timestampArg),
						bsonutil.WrapInOp(bsonutil.OpMultiply,
							bsonutil.WrapInRound(interval),
							3),
					),
					1)),
			)

		case Month:
			// SharedComputation = Month + round(interval) - 1.
			sharedComputationLetAssignment = bsonutil.NewM(
				bsonutil.NewDocElem("sharedComputation", bsonutil.WrapInOp(bsonutil.OpSubtract,
					bsonutil.WrapInOp(bsonutil.OpAdd,
						bsonutil.WrapInOp(bsonutil.OpMonth, timestampArg),
						bsonutil.WrapInRound(interval),
					),
					1)),
			)

		}

		newYearMonthLetAssignment = bsonutil.NewM(
			// Year = Year + SharedComputation / 12, where / truncates.

			bsonutil.NewDocElem("newYear", bsonutil.WrapInOp(bsonutil.OpAdd,
				bsonutil.WrapInOp(bsonutil.OpYear, timestampArg),
				bsonutil.WrapInIntDiv(sharedComputation, 12),
			)),

			// Month = SharedComputation % 12 + 1.
			bsonutil.NewDocElem("newMonth", bsonutil.WrapInOp(bsonutil.OpAdd,
				bsonutil.WrapInOp(bsonutil.OpMod,
					sharedComputation,
					12),
				1)),
		)

		// Add lets for Quarter and Month.
		return bsonutil.WrapInLet(sharedComputationLetAssignment,
			bsonutil.WrapInLet(newYearMonthLetAssignment, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDateFromParts, template))),
		)
	}

	// bsonutil.WrapInLet to bind $$timestampArg.
	switch unit {
	case Year, Month, Quarter:
		return bsonutil.WrapInLet(letAssignment, handleDateFromPartsCase(unit)), nil
	// It is wrong to round for Second, and rounding for Microsecond is
	// just pointless since MongoDB supports only milliseconds, and will
	// automatically round to the nearest millisecond for us.
	case Second, Microsecond:
		return bsonutil.WrapInLet(letAssignment, handleSimpleCase(unit, false)), nil
	default:
		return bsonutil.WrapInLet(letAssignment, handleSimpleCase(unit, true)), nil
	}
}

func (*timestampAddFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalNone, EvalDouble, EvalNone}
	convertedExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*timestampAddFunc) EvalType(exprs []SQLExpr) EvalType {
	// Checking the length of the argument to return conditional
	// types is not safe with pushdown. Timestamp add will
	// just always return a timestamp. There is no way to fix
	// this with respect to MongoDB's time semantics.
	return EvalDatetime
}

func (*timestampAddFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type timestampDiffFunc struct{}

func (*timestampDiffFunc) FuncName() string {
	return "timestampDiff"
}

var _ translatableToAggregationScalarFunc = (*timestampDiffFunc)(nil)

//http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_timestampdiff
func (f *timestampDiffFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	// These must be SQLTimestamps at this point. If they are not, something has
	// broken in the algebrizer. Check the handling of scalar functions in the
	// algebrizer.
	tmp1, ok := values[1].(SQLTimestamp)
	if !ok {
		return nil,
			fmt.Errorf("unable to evaluate type %v in timestampdiff,"+
				" this points to an error in the algebrizer",
				values[1])
	}
	t1 := Timestamp(tmp1)

	tmp2, ok := values[2].(SQLTimestamp)
	if !ok {
		return nil, fmt.Errorf("unable to evaluate type %v in timestampdiff,"+
			" this points to an error in the algebrizer",
			values[2])
	}
	t2 := Timestamp(tmp2)

	duration := t2.Sub(t1)

	switch values[0].String() {
	case Year:
		return NewSQLInt64(cfg.sqlValueKind, int64(numMonths(t1, t2)/12)), nil
	case Quarter:
		return NewSQLInt64(cfg.sqlValueKind, int64(numMonths(t1, t2)/3)), nil
	case Month:
		return NewSQLInt64(cfg.sqlValueKind, int64(numMonths(t1, t2))), nil
	case Week:
		if t1.After(t2) {
			return NewSQLInt64(cfg.sqlValueKind, int64(math.Ceil((duration.Hours())/24/7))), nil
		}
		return NewSQLInt64(cfg.sqlValueKind, int64(math.Floor((duration.Hours())/24/7))), nil
	case Day:
		if t1.After(t2) {
			return NewSQLInt64(cfg.sqlValueKind, int64(math.Ceil(duration.Hours()/24))), nil
		}
		return NewSQLInt64(cfg.sqlValueKind, int64(math.Floor(duration.Hours()/24))), nil
	case Hour:
		return NewSQLInt64(cfg.sqlValueKind, int64(duration.Hours())), nil
	case Minute:
		return NewSQLInt64(cfg.sqlValueKind, int64(duration.Minutes())), nil
	case Second:
		return NewSQLInt64(cfg.sqlValueKind, int64(duration.Seconds())), nil
	case Microsecond:
		return NewSQLInt64(cfg.sqlValueKind, duration.Nanoseconds()/1000), nil
	default:
		err := fmt.Errorf("cannot add '%v' to timestamp", values[0])
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), err
	}
}

func (*timestampDiffFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*timestampDiffFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 5, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(timestampDiff)",
			"cannot push down to MongoDB < 3.5",
		)
	}

	if len(exprs) != 3 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(timestampDiff)",
			incorrectArgCountMsg,
		)
	}

	unit := exprs[0].String()

	args, err := t.translateArgs(exprs[1:])
	if err != nil {
		return nil, err
	}

	timestampExpr1, timestampExpr2 := args[0], args[1]

	// This is a very large and costly expression, make sure to bind it in
	// a let (in the switch at the end of the function).
	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("timestampArg1", timestampExpr1),
		bsonutil.NewDocElem("timestampArg2", timestampExpr2),
	)

	// Use timestampArg{1,2} to refer to $$timestampArg{1,2} below,
	// referencing the var defined above.
	timestampArg1, timestampArg2 := "$$timestampArg1", "$$timestampArg2"

	// handleSimpleCase generates code for cases where we do not need to
	// use and date part access functions (like $dayOfMonth), we just
	// subtract: timestampArg2 - timestampArg1 then divide by the number of
	// milliseconds corresponded to by 'u'.
	handleSimpleCase := func(u string) interface{} {
		return bsonutil.WrapInIntDiv(bsonutil.WrapInOp(bsonutil.OpSubtract,
			timestampArg2,
			timestampArg1),
			toMilliseconds[u])
	}

	// handleDatePartsCase handles cases where we need to use
	// date part access functions (like $dayOfMonth).
	handleDatePartsCase := func(u string) interface{} {
		year1, month1 := "$$year1", "$$month1"
		year2, month2 := "$$year2", "$$month2"
		datePartsLetAssignment := bsonutil.NewM(
			bsonutil.NewDocElem("year1", bsonutil.WrapInOp(bsonutil.OpYear, timestampArg1)),
			bsonutil.NewDocElem("month1", bsonutil.WrapInOp(bsonutil.OpMonth, timestampArg1)),
			bsonutil.NewDocElem("day1", bsonutil.WrapInOp(bsonutil.OpDayOfMonth, timestampArg1)),
			bsonutil.NewDocElem("hour1", bsonutil.WrapInOp(bsonutil.OpHour, timestampArg1)),
			bsonutil.NewDocElem("minute1", bsonutil.WrapInOp(bsonutil.OpMinute, timestampArg1)),
			bsonutil.NewDocElem("second1", bsonutil.WrapInOp(bsonutil.OpSecond, timestampArg1)),
			bsonutil.NewDocElem("millisecond1", bsonutil.WrapInOp(bsonutil.OpMillisecond, timestampArg1)),
			bsonutil.NewDocElem("year2", bsonutil.WrapInOp(bsonutil.OpYear, timestampArg2)),
			bsonutil.NewDocElem("month2", bsonutil.WrapInOp(bsonutil.OpMonth, timestampArg2)),
			bsonutil.NewDocElem("day2", bsonutil.WrapInOp(bsonutil.OpDayOfMonth, timestampArg2)),
			bsonutil.NewDocElem("hour2", bsonutil.WrapInOp(bsonutil.OpHour, timestampArg2)),
			bsonutil.NewDocElem("minute2", bsonutil.WrapInOp(bsonutil.OpMinute, timestampArg2)),
			bsonutil.NewDocElem("second2", bsonutil.WrapInOp(bsonutil.OpSecond, timestampArg2)),
			bsonutil.NewDocElem("millisecond2", bsonutil.WrapInOp(bsonutil.OpMillisecond, timestampArg2)),
		)

		var outputLetAssignment interface{}
		var generateEpsilon func(arg1, arg2 string) interface{}
		output := "$$output"
		if u == Year {
			// For years, the output will be year2 - year1, but we
			// need to adjust that by the yearEpsilon, which is 0
			// or 1, depending on the remainder of the date object.
			// For instance if we have 2016-01-29 - 2015-01-30, the
			// answer is actually 0, because 30 > 29, giving us a
			// yearEpsilon of 1, and 1 - 1 = 0. If output is
			// positive we subtract the epsilon, if output is
			// negative, we add the epsilon, meaning we always go
			// toward 0.
			generateEpsilon = func(arg1, arg2 string) interface{} {
				return bsonutil.WrapInCond(bsonutil.WrapInLiteral(1),
					bsonutil.WrapInLiteral(0),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$month"+arg1, "$$month"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$day"+arg1, "$$day"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$hour"+arg1, "$$hour"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$minute"+arg1, "$$minute"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$second"+arg1, "$$second"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$millisecond"+arg1, "$$millisecond"+arg2),
				)
			}
			// output = year2 - year1.
			outputLetAssignment = bsonutil.NewM(
				bsonutil.NewDocElem("output", bsonutil.WrapInOp(bsonutil.OpSubtract, year2, year1)),
			)

		} else {
			// For months/quarters, the output will be (year2 -
			// year1) * 12 + month2 - month1, but we need to adjust
			// that by the monthEpsilon, which is 0 or 1, depending
			// on the remainder of the date object. For instance
			// if we have 2016-01-29 - 2015-01-30, the answer is
			// actually 11, because 30 > 29, giving us a
			// monthEpsilon of 1, and 12 - 1 = 11. If the output
			// is positive we subtract the epsilon, if output is
			// negative, we add the epsilon, meaning we always go
			// toward 0.
			generateEpsilon = func(arg1, arg2 string) interface{} {
				return bsonutil.WrapInCond(bsonutil.WrapInLiteral(1),
					bsonutil.WrapInLiteral(0),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$day"+arg1, "$$day"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$hour"+arg1, "$$hour"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$minute"+arg1, "$$minute"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$second"+arg1, "$$second"+arg2),
					bsonutil.WrapInOp(bsonutil.OpGt, "$$millisecond"+arg1, "$$millisecond"+arg2),
				)

			}
			// output = (year2 - year1) * 12 + month2 - month1.
			outputLetAssignment = bsonutil.NewM(
				bsonutil.NewDocElem("output", bsonutil.WrapInOp(bsonutil.OpAdd,
					bsonutil.WrapInOp(bsonutil.OpMultiply,
						bsonutil.WrapInOp(bsonutil.OpSubtract, year2, year1),
						12),
					bsonutil.WrapInOp(bsonutil.OpSubtract, month2, month1),
				)),
			)

		}

		// Generate epsilons and whether we add or subtract said epsilon, which
		// is decided on whether or not "output" is negative or positive.
		ltBranch := bsonutil.WrapInOp(bsonutil.OpAdd, output, generateEpsilon("2", "1"))
		gtBranch := bsonutil.WrapInOp(bsonutil.OpSubtract, output, generateEpsilon("1", "2"))
		applyEpsilonExpr := bsonutil.WrapInLet(outputLetAssignment,
			bsonutil.WrapInSwitch(bsonutil.WrapInLiteral(0),
				bsonutil.WrapInCase(bsonutil.WrapInOp(bsonutil.OpLt, output, bsonutil.WrapInLiteral(0)), ltBranch),
				bsonutil.WrapInCase(bsonutil.WrapInOp(bsonutil.OpGt, output, bsonutil.WrapInLiteral(0)), gtBranch),
			),
		)

		retExpr := bsonutil.WrapInLet(datePartsLetAssignment,
			bsonutil.WrapInLet(outputLetAssignment,
				applyEpsilonExpr,
			),
		)
		// Quarter is just the number of months integer divided by 3.
		if u == Quarter {
			return bsonutil.WrapInIntDiv(retExpr, 3)
		}
		return retExpr
	}

	// bsonutil.WrapInLet to bind $$timestampArg.
	switch unit {
	case Year, Month, Quarter:
		return bsonutil.WrapInLet(letAssignment, handleDatePartsCase(unit)), nil
	default:
		return bsonutil.WrapInLet(letAssignment, handleSimpleCase(unit)), nil
	}
}

func (*timestampDiffFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type timestampFunc struct{}

func (*timestampFunc) FuncName() string {
	return "timestamp"
}

var _ normalizingScalarFunc = (*timestampFunc)(nil)
var _ translatableToAggregationScalarFunc = (*timestampFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_timestamp
func (f *timestampFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	t = t.In(schema.DefaultLocale)

	if len(values) == 1 {
		return NewSQLTimestamp(cfg.sqlValueKind, t), nil
	}

	d, ok := parseDuration(values[1])
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	t = t.Add(d).Round(time.Microsecond)

	return NewSQLTimestamp(cfg.sqlValueKind, t), nil
}

func (*timestampFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 5, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(timestamp)",
			"cannot push down to MongoDB < 3.5",
		)
	}

	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(timestamp)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	val := "$$val"
	inputLet := bsonutil.NewM(
		bsonutil.NewDocElem("val", args[0]),
	)

	wrapInDateFromString := func(v interface{}) bson.M {
		return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDateFromString, bsonutil.NewM(bsonutil.NewDocElem("dateString", v))))
	}

	// CASE 1: it's already a Mongo date, we just return it
	isDateType := containsBSONType(val, "date")
	dateBranch := bsonutil.WrapInCase(isDateType, val)

	// CASE 2: it's a number.
	isNumber := containsBSONType(val, "int", "decimal", "long", "double")

	// evaluates to true if val positive and has <= X digits.
	hasUpToXDigits := func(x float64) interface{} {
		return bsonutil.WrapInInRange(val, 0, math.Pow(10, x))
	}

	// This handles converting a number in YYMMDDHHMMSS format to YYYYMMDDHHMMSS.
	// if YY < 70, we assume they meant 20YY. if YY > 70, we assume 19YY.
	getPadding := func(v interface{}) interface{} {
		return bsonutil.WrapInCond(
			20000000000000,
			19000000000000,
			bsonutil.WrapInOp(bsonutil.OpLt,
				bsonutil.WrapInOp(bsonutil.OpDivide,
					v, 10000000000),
				70))
	}

	// Constant for the HHMMSS factor to handle dates that do not have HHMMSS.
	hhmmssFactor := 1000000

	// We interpret this as being format YYMMDD, multiply by hhmmssFactor for HHMMSS then pad.
	ifSix := bsonutil.WrapInOp(bsonutil.OpAdd,
		bsonutil.WrapInOp(bsonutil.OpMultiply,
			val,
			hhmmssFactor),
		getPadding(bsonutil.WrapInOp(bsonutil.OpMultiply,
			val,
			hhmmssFactor)))
	sixBranch := bsonutil.WrapInCase(hasUpToXDigits(6), ifSix)

	// This number is YYYYMMDD, again, multiply by hhmmssFactor.
	eightBranch := bsonutil.WrapInCase(hasUpToXDigits(8), bsonutil.WrapInOp(bsonutil.OpMultiply, val, hhmmssFactor))

	// If it's twelve digits, interpret as YYMMDDHHMMSS. Make sure to pad the number.
	ifTwelve := bsonutil.WrapInOp(bsonutil.OpAdd, val, getPadding(val))
	twelveBranch := bsonutil.WrapInCase(hasUpToXDigits(12), ifTwelve)

	// if fourteen, YYYYMMDDHHMMSS, we can use as it as is.
	fourteenBranch := bsonutil.WrapInCase(hasUpToXDigits(14), val)

	// define "num", the input number normalized to 14 digits, in a "let"
	numberVar := bsonutil.WrapInSwitch(nil, sixBranch, eightBranch, twelveBranch, fourteenBranch)
	numberLetVars := bsonutil.NewM(bsonutil.NewDocElem("num", numberVar))

	dateParts := bsonutil.NewM(
		// YYYYMMDDHHMMSS / 10000000000 = YYYY

		bsonutil.NewDocElem("year", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide,
			"$$num",
			10000000000)),
		)),
		// (YYYYMMDDHHMMSS / 100000000) % 100 = MM

		bsonutil.NewDocElem("month", bsonutil.WrapInOp(bsonutil.OpMod, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide,
			"$$num",
			100000000)),
		), 100)),

		// YYYYMMDDHHMMSS / 1000000) % 100 = DD
		bsonutil.NewDocElem("day", bsonutil.WrapInOp(bsonutil.OpMod, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide,
			"$$num",
			1000000)),
		), 100)),

		// YYYYMMDDHHMMSS / 10000) % 100 = HH
		bsonutil.NewDocElem("hour", bsonutil.WrapInOp(bsonutil.OpMod, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide,
			"$$num",
			10000)),
		), 100)),

		// YYYYMMDDHHMMSS / 100) % 100 = MM
		bsonutil.NewDocElem("minute", bsonutil.WrapInOp(bsonutil.OpMod, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpDivide,
			"$$num",
			100)),
		), 100)),

		// YYYYMMDDHHMMSS % 100 = SS
		bsonutil.NewDocElem("second", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpMod,
			"$$num",
			100)),
		)),
		// YYYYMMDDHHMMSS.FFFFF % 1 * 1000 = ms

		bsonutil.NewDocElem("millisecond", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.WrapInOp(bsonutil.OpMultiply,
			bsonutil.WrapInOp(bsonutil.OpMod,
				"$$num",
				1),
			1000)),
		)),
	)

	// try to avoid aggregation errors by catching obviously invalid dates
	yearValid := bsonutil.WrapInInRange("$$year", 0, 10000)
	monthValid := bsonutil.WrapInInRange("$$month", 1, 13)
	dayValid := bsonutil.WrapInInRange("$$day", 1, 32)
	// Mongo DB actually supports HH=24 which converts to 0, but MySQL does not (it returns NULL)
	// so we stick to MySQL semantics and cap valid hours at 23.
	// Interestingly, $dateFromString does NOT support HH=24.
	hourValid := bsonutil.WrapInInRange("$$hour", 0, 24)
	minuteValid := bsonutil.WrapInInRange("$$minute", 0, 60)
	secondValid := bsonutil.WrapInInRange("$$second", 0, 60)

	makeDateOrNull := bsonutil.WrapInCond(bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDateFromParts, bsonutil.NewM(
		bsonutil.NewDocElem("year", "$$year"),
		bsonutil.NewDocElem("month", "$$month"),
		bsonutil.NewDocElem("day", "$$day"),
		bsonutil.NewDocElem("hour", "$$hour"),
		bsonutil.NewDocElem("minute", "$$minute"),
		bsonutil.NewDocElem("second", "$$second"),
		bsonutil.NewDocElem("millisecond", "$$millisecond"),
	)),
	), nil, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpAnd, bsonutil.NewArray(
		yearValid,
		monthValid,
		dayValid,
		hourValid,
		minuteValid,
		secondValid,
	)),
	))

	evaluateNumber := bsonutil.WrapInLet(dateParts, makeDateOrNull)
	handleNumberToDate := bsonutil.WrapInLet(numberLetVars, evaluateNumber)
	numberBranch := bsonutil.WrapInCase(isNumber, handleNumberToDate)

	// CASE 3: it's a string
	isString := containsBSONType(val, "string")

	// First split on T, take first substring, then split that on " ", and
	// take first substring. this gives us just the date part of the
	// string. note that if the string doesn't have T or a space, just
	// returns original string
	trimmedDateString := bsonutil.WrapInOp(bsonutil.OpArrElemAt,
		bsonutil.WrapInOp(bsonutil.OpSplit,
			bsonutil.WrapInOp(bsonutil.OpArrElemAt,
				bsonutil.WrapInOp(bsonutil.OpSplit, val, "T"),
				0),
			" "),
		0)

	// Repeat the step above but take the second element to get the time
	// part. Replace with "" if we can not find a second element.
	trimmedTimeString := bsonutil.WrapInIfNull(
		bsonutil.WrapInOp(bsonutil.OpArrElemAt,
			bsonutil.WrapInOp(bsonutil.OpSplit, val, "T"),
			1),
		bsonutil.WrapInIfNull(
			bsonutil.WrapInOp(bsonutil.OpArrElemAt,
				bsonutil.WrapInOp(bsonutil.OpSplit, val, " "),
				1),
			""),
	)

	// Convert the date and time strings to arrays so we can use
	// map/reduce.
	trimmedDateAsArray := bsonutil.WrapInStringToArray("$$trimmedDate")
	trimmedTimeAsArray := bsonutil.WrapInStringToArray("$$trimmedTime")

	// isSeparator evaluates to true if a character is in the defined
	// separator list
	isSeparator := bsonutil.WrapInOp(bsonutil.OpNeq,
		-1,
		bsonutil.WrapInOp("$indexOfArray",
			bsonutil.DateComponentSeparator,
			"$$c"))

	// Use map to convert all separators in the date string to - symbol,
	// and leave numbers as-is
	dateNormalized := bsonutil.WrapInMap(trimmedDateAsArray,
		"c",
		bsonutil.WrapInCond("-",
			"$$c",
			isSeparator))
	// Use map to convert all separators in the time string to '.' symbol,
	// and leave numbers as-is. We use '.' instead of ':' so that MongoDB
	// correctly handles fractional seconds. 10.11.23.1234 is parsed
	// correctly as 10:11:23.1234, saving us some effort (and runtime).
	timeNormalized := bsonutil.WrapInMap(trimmedTimeAsArray,
		"c",
		bsonutil.WrapInCond(".",
			"$$c",
			isSeparator))

	// Use reduce to convert characters back to a single string for date and time.
	dateJoined := bsonutil.WrapInReduce(dateNormalized,
		"",
		bsonutil.WrapInOp(bsonutil.OpConcat,
			"$$value",
			"$$this"))
	timeJoined := bsonutil.WrapInReduce(timeNormalized,
		"",
		bsonutil.WrapInOp(bsonutil.OpConcat,
			"$$value",
			"$$this"))

	// if the third character is a -, or if the string is only 6 digits
	// long and has no slashes, then the string is either format YY/MM/DD
	// or YYMMDD and we need to add the appropriate first two year digits
	// (19xx or 20xx) for Mongo to understand it
	hasShortYear := bsonutil.WrapInOp(bsonutil.OpOr,
		// length is only 6, assume YYMMDD
		bsonutil.WrapInOp(bsonutil.OpEq, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, "$$dateJoined")), 6),
		// third character is -, assume YY-MM-DD
		bsonutil.WrapInOp(bsonutil.OpEq,
			"-", bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubstr, bsonutil.NewArray(
				"$$dateJoined",
				2,
				1,
			)),
			)))

	// "$dateFromString" actually pads correctly, but not if "/" is
	// used as the separator (it will assume year is last). If this
	// pushdown is shown to be slow by benchmarks, we should reconsider
	// allowing "$dateFromString" to handle padding. The change
	// would not be trivial due to how MongoDB cannot handle short dates
	// when there are no separators in the date.
	padYear := bsonutil.WrapInOp(bsonutil.OpConcat,
		bsonutil.WrapInCond(
			"20",
			"19",
			// check if first two digits < 70 to determine padding
			bsonutil.WrapInOp(
				bsonutil.OpLt, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubstr, bsonutil.NewArray(
					"$$dateJoined",
					0,
					2,
				))), "70")),
		"$$dateJoined")

	// we have to use nested $lets because in the outer one we define
	// $$trimmedDate and in the inner one we define $$dateJoined. defining
	// $$dateJoined requires knowing the length of trimmedDate, so we can't
	// do it all in one step.
	innerIn := bsonutil.WrapInCond(padYear, "$$dateJoined", hasShortYear)
	innerLet := bsonutil.WrapInLet(bsonutil.NewM(bsonutil.NewDocElem("dateJoined", dateJoined)), innerIn)

	// Concat the time back into the date.
	concatedDate := bsonutil.WrapInOp(bsonutil.OpConcat,
		innerLet,
		timeJoined)

	// gracefully handle strings that are too short to possibly be valid by returning null
	tooShort := bsonutil.WrapInOp(bsonutil.OpLt, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpStrlenCP, "$$trimmedDate")), 6)
	outerIn := bsonutil.WrapInCond(nil, wrapInDateFromString(concatedDate), tooShort)
	outerLet := bsonutil.WrapInLet(bsonutil.NewM(bsonutil.NewDocElem("trimmedDate", trimmedDateString),
		bsonutil.NewDocElem("trimmedTime", trimmedTimeString),
	), outerIn)

	// Make sure if we get the int 0 we return NULL instead
	// of crashing. MySQL uses '0000-00-00' as an error output for some
	// functions and we encode it as the integer 0 within push down.
	stringBranch := bsonutil.WrapInCase(isString,
		bsonutil.WrapInCond(nil,
			outerLet,
			bsonutil.WrapInOp(bsonutil.OpEq,
				0,
				args[0])))

	return bsonutil.WrapInLet(inputLet, bsonutil.WrapInSwitch(nil, dateBranch, numberBranch, stringBranch)), nil

}

func (*timestampFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*timestampFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDatetime
}

func (*timestampFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

// toDaysFunc is an implementation of the mysql function TO_DAYS, which returns
// number of days since 0000-00-00. There are a few interesting issues here:
// 1. 0000-00-00 is not a valid date, so TO_DAYS('0000-00-00') is supposed to return NULL
//    and TO_DAYS('0000-01-01') is supposed to be 1 rather than the 0 we return.
// 2. However, due to a bug in MySQL treating year 0 as a non-leap year, our results
//    are correct for any date after 0000-02-29 (which MySQL thinks isn't a day).
//    year zero should be a leap year:
//    https://en.wikipedia.org/wiki/Year_zero,
//    Both MongoDB and the go time library treat year 0 as a leap year, as
//    well. If, at some point, MySQL should correct their calendar, we could
//    switch to adding to our result to tie-out with them.
type toDaysFunc struct{}

func (*toDaysFunc) FuncName() string {
	return "toDays"
}

var _ normalizingScalarFunc = (*toDaysFunc)(nil)
var _ translatableToAggregationScalarFunc = (*toDaysFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_to-days
func (f *toDaysFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	// This must be a SQLDate at this point. If it is not, the algebrizer has been broken.
	// Check the handling of scalar functions in the algebrizer.
	tmp, ok := values[0].(SQLDate)
	if !ok {
		return nil, fmt.Errorf("unable to evaluate input value %v in to_days", values[0])
	}
	date := Timestamp(tmp)

	// First compute the days from YearOne.
	target, err := daysFromYearZeroCalculation(date)
	if err != nil {
		return nil, err
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(target)), nil
}

func (*toDaysFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

// FuncToAggregation for TO_DAYS has one issue wrt how TO_DAYS is supposed to perform:
// because our date treatment is backed by using MongoDB's $dateFromString function,
// if a date that doesn't exist (e.g., 0000-00-00 or 0001-02-29) is entered, we return
// an error instead of the NULL expected from MySQL. Unfortunately, checking for valid
// dates is too cost prohibitive. If at some point $dateFromString supports an onError/default
// value, we should switch to using that.
func (*toDaysFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(toDays)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}
	// Subtract dayOne (0000-01-01) from the argument in MongoDB, then convert ms to days.
	// When using $subtract on two dates in MongoDB, the number of ms between the two
	// dates is returned, and the purpose of the TO_DAYS function is to get the number
	// of days since 0000-01-01:
	// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_to-days
	// Unfortunately, we get a slightly wrong number if we try to multiply by days/ms
	// because MySQL itself is using division (and actually gets the wrong day count itself)
	// NOTE: args[0] must come in as a date creating expression, because we rewrite
	// to_days(x) in the algebrizer to to_days(date(x)).
	dayOne := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDivide, bsonutil.NewArray(
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
			args[0],
			dayOne,
		))),
		millisecondsPerDay,
	)),
	)),
	), nil
}

func (*toDaysFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*toDaysFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

// toSecondsFunc is an implementation of the mysql function TO_SECONDS, which
// returns number of seconds since 0000-00-00. There are a few interesting
// issues here:
// 1. 0000-00-00 is not a valid date, so TO_SECONDS('0000-00-00') is supposed to
// return NULL
//    and TO_SECONDS('0000-01-01') is supposed to be 86400 (seconds in 1 day)
//    rather than the 0 we return.
// 2. However, due to a bug in MySQL treating year 0 as a non-leap year, our results
//    are correct for any date after 0000-02-29 (which MySQL thinks isn't a
//    day). year zero should be a leap year:
//    https://en.wikipedia.org/wiki/Year_zero.
//    Both MongoDB and the go time library treat year 0 as a leap year, as
//    well. If, at some point, MySQL should correct their calendar, we could
//    switch to adding to our result to tie-out with them.
type toSecondsFunc struct{}

func (*toSecondsFunc) FuncName() string {
	return "toSeconds"
}

var _ translatableToAggregationScalarFunc = (*toSecondsFunc)(nil)

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_to-seconds
func (f *toSecondsFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	// This must be a SQLTimestamp at this point. If it is not, the algebrizer has been broken.
	// Check the handling of scalar functions in the algebrizer.
	tmp, ok := values[0].(SQLTimestamp)
	if !ok {
		return nil, fmt.Errorf("unable to evaluate input value %v in to_seconds", values[0])
	}
	date := Timestamp(tmp)

	// First compute the days from YearOne and convert to seconds.
	target, err := daysFromYearZeroCalculation(date)
	if err != nil {
		return nil, err
	}
	target *= secondsPerDay

	// Now add remainder hours, minutes, and seconds.
	target += float64(date.Hour())*secondsPerHour +
		float64(date.Minute())*secondsPerMinute + float64(date.Second())

	// target is now seconds since dayOne.
	return NewSQLInt64(cfg.sqlValueKind, int64(target)), nil
}

func (*toSecondsFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(toSeconds)",
			incorrectArgCountMsg,
		)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}
	// Subtract dayOne (0000-01-01) from the argument in mongo, then
	// convertms to seconds. When using $subtract on two dates in
	// MongoDB, the number of ms between the two dates is returned, and
	// the purpose of the TO_SECONDS function is to get the number of
	// seconds since 0000-01-01:
	dayOne := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	return bsonutil.WrapInOp(bsonutil.OpMultiply,
		bsonutil.WrapInOp(bsonutil.OpSubtract, args[0], dayOne),
		1e-3,
	), nil
}

func (*toSecondsFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*toSecondsFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type trimFunc struct{}

func (*trimFunc) FuncName() string {
	return "trim"
}

var _ reconcilingScalarFunc = (*trimFunc)(nil)
var _ translatableToAggregationScalarFunc = (*trimFunc)(nil)

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_trim
func (f *trimFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	value := values[0].String()
	end := "both"
	toTrim := " "
	if len(values) == 3 {
		end = values[1].String()
		toTrim = values[2].String()
	}

	save := ""
	for save != value {
		save = value
		switch end {
		case "both":
			value = strings.TrimPrefix(value, toTrim)
			value = strings.TrimSuffix(value, toTrim)
		case "leading":
			value = strings.TrimPrefix(value, toTrim)
		case "trailing":
			value = strings.TrimSuffix(value, toTrim)
		}
	}

	return NewSQLVarchar(cfg.sqlValueKind, value), nil
}

func (*trimFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(trim)",
			"cannot push down to MongoDB < 3.4",
		)
	}
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(trim)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	if t.versionAtLeast(4, 0, 0) {
		return bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpTrim, bsonutil.NewM(
				bsonutil.NewDocElem("input", args[0]),
				bsonutil.NewDocElem("chars", " "),
			)),
		), nil
	}

	rtrimCond := bsonutil.WrapInCond(
		"",
		bsonutil.WrapInLRTrim(false, args[0]), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
			args[0],
			"",
		))))

	ltrimCond := bsonutil.WrapInCond(
		"",
		bsonutil.WrapInLRTrim(true, "$$rtrim"), bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
			"$$rtrim",
			"",
		))))

	trimCond := bsonutil.WrapInLet(bsonutil.NewM(bsonutil.NewDocElem("rtrim", rtrimCond)), ltrimCond)

	trim := bsonutil.WrapInCond(
		"",
		trimCond, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpEq, bsonutil.NewArray(
			args[0],
			"",
		))))

	return bsonutil.WrapInNullCheckedCond(
		nil,
		trim,
		args[0],
	), nil
}

func (*trimFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalString)
}

func (*trimFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*trimFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 3)
}

type truncateFunc struct{}

func (*truncateFunc) FuncName() string {
	return "truncate"
}

var _ normalizingScalarFunc = (*truncateFunc)(nil)
var _ reconcilingScalarFunc = (*truncateFunc)(nil)
var _ translatableToAggregationScalarFunc = (*truncateFunc)(nil)

//http://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_truncate
func (f *truncateFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	var truncated float64
	x := Float64(values[0])
	d := Float64(values[1])

	if d >= 0 {
		pow := math.Pow(10, d)
		i, _ := math.Modf(x * pow)
		truncated = i / pow
	} else {
		pow := math.Pow(10, math.Abs(d))
		i, _ := math.Modf(x / pow)
		truncated = i * pow
	}

	return NewSQLFloat(cfg.sqlValueKind, truncated), nil
}

func (*truncateFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(truncate)",
			incorrectArgCountMsg,
		)
	}
	dValue, ok := exprs[1].(SQLValue)
	if !ok {
		return nil, newPushdownFailure("SQLScalarFunctionExpr(truncate)", "second arg is not a literal")
	}

	d := Float64(dValue)

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	if d >= 0 {
		pow := math.Pow(10, d)
		return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpDivide, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCond, bsonutil.NewArray(
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGte, bsonutil.NewArray(
					args[0],
					0,
				))),
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpFloor, bsonutil.NewM(
					bsonutil.NewDocElem(bsonutil.OpMultiply, bsonutil.NewArray(
						args[0],
						pow,
					))))),
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCeil, bsonutil.NewM(
					bsonutil.NewDocElem(bsonutil.OpMultiply, bsonutil.NewArray(
						args[0],
						pow,
					))))),
			))),
			pow,
		))), nil
	}

	pow := math.Pow(10, math.Abs(d))
	return bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMultiply, bsonutil.NewArray(
		bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCond, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpGte, bsonutil.NewArray(
				args[0],
				0,
			))),
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpFloor, bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpDivide, bsonutil.NewArray(
					args[0],
					pow,
				))))),
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpCeil, bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpDivide, bsonutil.NewArray(
					args[0],
					pow,
				))))),
		))),
		pow,
	))), nil
}

func (*truncateFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*truncateFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalDouble, EvalNone}
	convertedExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*truncateFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDouble
}

func (*truncateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type ucaseFunc struct{}

func (*ucaseFunc) FuncName() string {
	return "ucase"
}

var _ reconcilingScalarFunc = (*ucaseFunc)(nil)
var _ translatableToAggregationScalarFunc = (*ucaseFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_ucase
func (f *ucaseFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	value := strings.ToUpper(values[0].String())

	return NewSQLVarchar(cfg.sqlValueKind, value), nil
}

func (*ucaseFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(ucase)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$toUpper", args[0]), nil
}

func (*ucaseFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalString)
}

func (*ucaseFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*ucaseFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type unixTimestampFunc struct{}

func (*unixTimestampFunc) FuncName() string {
	return "unixTimestamp"
}

var _ normalizingScalarFunc = (*unixTimestampFunc)(nil)
var _ translatableToAggregationScalarFunc = (*unixTimestampFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_unix-timestamp
func (f *unixTimestampFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	now := time.Now()

	if len(values) == 0 {
		return NewSQLUint64(cfg.sqlValueKind, uint64(now.Unix())), nil
	}

	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	t, _, ok := parseDateTime(values[0].String())
	if !ok || t.Before(epoch) {
		return NewSQLFloat(cfg.sqlValueKind, 0.0), nil
	}

	// Our times are parsed as if in UTC. However, we need to
	// parse it in the actual location the server's running
	// in - to account for any timezone difference.
	y, m, d := t.Date()
	ts := time.Date(y, m, d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), now.Location())
	return NewSQLUint64(cfg.sqlValueKind, uint64(ts.Unix())), nil
}

func (*unixTimestampFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	now := time.Now()

	if len(exprs) != 1 {
		return bsonutil.WrapInLiteral(now.Unix()), nil
	}

	arg, err := (&timestampFunc{}).FuncToAggregationLanguage(t, exprs)
	if err != nil {
		return nil, err
	}

	// Subtract epoch (1970-01-01) from the argument in MongoDB, then
	// convert ms to seconds. When using $subtract on two dates in
	// MongoDB, the number of milliseconds between the two
	// timestamps is returned.
	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	_, tzCompensation := now.Zone()

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("diff", bsonutil.NewM(
			bsonutil.NewDocElem(bsonutil.OpTrunc, bsonutil.NewM(
				bsonutil.NewDocElem(bsonutil.OpDivide, bsonutil.NewArray(
					bsonutil.NewM(
						bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
							bsonutil.NewM(
								bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
									arg,
									epoch,
								)),
							),
							tzCompensation*1000,
						)),
					),
					1000,
				)),
			)),
		)),
	)

	letEvaluation := bsonutil.WrapInCond("$$diff", 0.0, bsonutil.WrapInOp(bsonutil.OpGt, "$$diff", 0))
	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil
}

func (*unixTimestampFunc) Normalize(kind SQLValueKind, f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return NewSQLNull(kind, f.EvalType())
	}

	return f
}

func (*unixTimestampFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalUint64
}

func (*unixTimestampFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0, 1)
}

type userFunc struct{}

func (*userFunc) FuncName() string {
	return "user"
}

func (*userFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	str := fmt.Sprintf("%s@%s", cfg.user, cfg.remoteHost)
	return NewSQLVarchar(cfg.sqlValueKind, str), nil
}

func (*userFunc) SkipConstantFolding() bool {
	return true
}

func (*userFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*userFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type utcDateFunc struct{}

func (*utcDateFunc) FuncName() string {
	return "utcDate"
}

var _ translatableToAggregationScalarFunc = (*utcDateFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_utc-date
func (*utcDateFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	now := time.Now().In(time.UTC)
	t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return NewSQLDate(cfg.sqlValueKind, t), nil
}

func (*utcDateFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	now := time.Now().In(time.UTC)
	cUTCd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return bsonutil.WrapInLiteral(cUTCd), nil
}

func (*utcDateFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDate
}

func (*utcDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type utcTimestampFunc struct{}

func (*utcTimestampFunc) FuncName() string {
	return "utcTimestamp"
}

var _ translatableToAggregationScalarFunc = (*utcTimestampFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_utc-timestamp
func (*utcTimestampFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	return NewSQLTimestamp(cfg.sqlValueKind, time.Now().In(time.UTC)), nil
}

func (*utcTimestampFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	return bsonutil.WrapInLiteral(time.Now().In(time.UTC)), nil
}

func (*utcTimestampFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalDatetime
}

func (*utcTimestampFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type versionFunc struct{}

func (*versionFunc) FuncName() string {
	return "version"
}

func (*versionFunc) Evaluate(_ context.Context, cfg *ExecutionConfig, _ *ExecutionState, _ []SQLValue) (SQLValue, error) {
	return NewSQLVarchar(cfg.sqlValueKind, cfg.mySQLVersion), nil
}

func (*versionFunc) SkipConstantFolding() bool {
	return true
}

func (*versionFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalString
}

func (*versionFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type weekFunc struct{}

func (*weekFunc) FuncName() string {
	return "week"
}

var _ reconcilingScalarFunc = (*weekFunc)(nil)
var _ translatableToAggregationScalarFunc = (*weekFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_week
func (f *weekFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}
	check, ok := values[0].(SQLDate)
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	dateArg := Timestamp(check)
	// Mode should always be less than MAX_INT.
	mode := int(Int64(values[1]))

	ret := weekCalculation(dateArg, mode)
	if ret == -1 {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}
	return NewSQLInt64(cfg.sqlValueKind, int64(ret)), nil
}

func (*weekFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) < 1 || len(exprs) > 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(week)",
			incorrectArgCountMsg,
		)
	}
	mode := int64(0)
	if len(exprs) == 2 {
		modeValue, ok := exprs[1].(SQLValue)
		if !ok {
			return nil, newPushdownFailure("SQLScalarFunctionExpr(week)", "mode is not a literal")
		}
		mode = Int64(modeValue)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}
	return bsonutil.WrapInWeekCalculation(args[0], mode), nil
}

func (*weekFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []EvalType{EvalNone, EvalInt64}
	convertedExprs := convertExprs(f.Exprs, argTypes)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*weekFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*weekFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

type weekdayFunc struct{}

func (*weekdayFunc) FuncName() string {
	return "weekday"
}

var _ translatableToAggregationScalarFunc = (*weekdayFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_weekday
func (f *weekdayFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	w := int(t.Weekday())
	if w == 0 {
		w = 7
	}
	return NewSQLInt64(cfg.sqlValueKind, int64(w-1)), nil
}

func (*weekdayFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(weekday)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	letAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("date", args[0]),
	)

	conds := bsonutil.NewArray()
	if _, ok := bsonutil.GetLiteral(args[0]); !ok {
		conds = append(conds, "$$date")
	}

	letEvaluation := bsonutil.WrapInNullCheckedCond(
		nil, bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMod, bsonutil.NewArray(
			bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpAdd, bsonutil.NewArray(
				bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpMod, bsonutil.NewArray(
					bsonutil.NewM(bsonutil.NewDocElem(bsonutil.OpSubtract, bsonutil.NewArray(
						bsonutil.NewM(bsonutil.NewDocElem("$dayOfWeek", "$$date")),
						2,
					)),
					),
					7,
				)),
				),
				7,
			)),
			),
			7,
		)),
		), conds...,
	)

	return bsonutil.WrapInLet(letAssignment, letEvaluation), nil

}

func (*weekdayFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*weekdayFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type yearFunc struct{}

func (*yearFunc) FuncName() string {
	return "year"
}

var _ reconcilingScalarFunc = (*yearFunc)(nil)
var _ translatableToAggregationScalarFunc = (*yearFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_year
func (f *yearFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	return NewSQLInt64(cfg.sqlValueKind, int64(t.Year())), nil
}

func (*yearFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) != 1 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(year)",
			incorrectArgCountMsg,
		)
	}
	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	return bsonutil.WrapSingleArgFuncWithNullCheck("$year", args[0]), nil
}

func (*yearFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, EvalDate)
}

func (*yearFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*yearFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type yearWeekFunc struct{}

func (*yearWeekFunc) FuncName() string {
	return "yearWeek"
}

var _ translatableToAggregationScalarFunc = (*yearWeekFunc)(nil)

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_yearweek
func (f *yearWeekFunc) Evaluate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, values []SQLValue) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}
	check, ok := values[0].(SQLDate)
	if !ok {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	date := Timestamp(check)
	year := date.Year()
	// Mode should always be less than MAX_INT.
	mode := int(Int64(values[1]))

	var week int

	// Unlike WEEK, YEARWEEK always uses the 1-53 modes. Thus
	// we always call week with the 1-53 of a 0-53, 1-53 pair.
	switch mode {

	// First day of week: Sunday, with a Sunday in this year.
	case 0, 2:
		week = weekCalculation(date, 2)
	// First day of week: Monday, with 4 days in this year.
	case 1, 3:
		week = weekCalculation(date, 3)
	// First day of week: Sunday, with 4 days in this year.
	case 4, 6:
		week = weekCalculation(date, 6)
	// First day of week: Monday, with a Monday in this year.
	case 5, 7:
		week = weekCalculation(date, 7)
	}

	if week == -1 {
		return NewSQLNull(cfg.sqlValueKind, f.EvalType(valsAsExprs(values))), nil
	}

	switch week {
	case 1:
		if date.Month() == 12 {
			year++
		}
	case 52, 53:
		if date.Month() == 1 {
			year--
		}

	}
	return NewSQLInt64(cfg.sqlValueKind, int64(year*100+week)), nil
}

func (*yearWeekFunc) FuncToAggregationLanguage(t *PushdownTranslator, exprs []SQLExpr) (interface{}, PushdownFailure) {
	if len(exprs) < 1 || len(exprs) > 2 {
		return nil, newPushdownFailure(
			"SQLScalarFunctionExpr(yearWeek)",
			incorrectArgCountMsg,
		)
	}
	mode := int64(0)
	if len(exprs) == 2 {
		modeValue, ok := exprs[1].(SQLValue)
		if !ok {
			return nil, newPushdownFailure("SQLScalarFunctionExpr(yearWeek)", "mode is not a literal")
		}
		mode = Int64(modeValue)
	}

	args, err := t.translateArgs(exprs)
	if err != nil {
		return nil, err
	}

	date, month, year, week := "$$date", "$$month", "$$year", "$$week"
	inputAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("date", args[0]),
	)

	monthAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("month", bsonutil.WrapInOp(bsonutil.OpMonth, date)),
		bsonutil.NewDocElem("year", bsonutil.WrapInOp(bsonutil.OpYear, date)),
	)

	var weekCalc interface{}

	// Unlike WEEK, YEARWEEK always uses the 1-53 modes. Thus
	// we always call week with the 1-53 of a 0-53, 1-53 pair.
	switch mode {

	// First day of week: Sunday, with a Sunday in this year.
	case 0, 2:
		weekCalc = bsonutil.WrapInWeekCalculation(date, 2)
	// First day of weekCalc: Monday, with 4 days in this year.
	case 1, 3:
		weekCalc = bsonutil.WrapInWeekCalculation(date, 3)
	// First day of weekCalc: Sunday, with 4 days in this year.
	case 4, 6:
		weekCalc = bsonutil.WrapInWeekCalculation(date, 6)
	// First day of weekCalc: Monday, with a Monday in this year.
	case 5, 7:
		weekCalc = bsonutil.WrapInWeekCalculation(date, 7)
	}

	weekAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("week", weekCalc),
	)

	newYear := "$$newYear"
	newYearAssignment := bsonutil.NewM(
		bsonutil.NewDocElem("newYear", bsonutil.WrapInSwitch(year,
			bsonutil.WrapInEqCase(week, 1, bsonutil.WrapInCond(
				bsonutil.WrapInOp(bsonutil.OpAdd, year, 1), year,
				bsonutil.WrapInOp(bsonutil.OpEq, month, 12),
			),
			),
			bsonutil.WrapInEqCase(week, 52, bsonutil.WrapInCond(
				bsonutil.WrapInOp(bsonutil.OpSubtract, year, 1), year,
				bsonutil.WrapInOp(bsonutil.OpEq, month, 1),
			),
			),
			bsonutil.WrapInEqCase(week, 53, bsonutil.WrapInCond(
				bsonutil.WrapInOp(bsonutil.OpSubtract, year, 1), year,
				bsonutil.WrapInOp(bsonutil.OpEq, month, 1),
			),
			),
		)),
	)

	return bsonutil.WrapInLet(inputAssignment,
		bsonutil.WrapInLet(monthAssignment,
			bsonutil.WrapInLet(weekAssignment,
				bsonutil.WrapInLet(newYearAssignment,
					bsonutil.WrapInOp(bsonutil.OpAdd,
						bsonutil.WrapInOp(bsonutil.OpMultiply, newYear, 100),
						week,
					),
				),
			),
		),
	), nil

}

func (*yearWeekFunc) EvalType(exprs []SQLExpr) EvalType {
	return EvalInt64
}

func (*yearWeekFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

func (t *PushdownTranslator) translateArgs(exprs []SQLExpr) ([]interface{}, PushdownFailure) {
	args := []interface{}{}
	for _, e := range exprs {
		r, err := t.ToAggregationLanguage(e)
		if err != nil {
			return nil, err
		}
		args = append(args, r)
	}
	return args, nil
}

// NewSQLScalarFunctionExpr returns a new SQLScalarFunctionExpr with the
// provided name and arguments.
func NewSQLScalarFunctionExpr(name string, exprs []SQLExpr) (*SQLScalarFunctionExpr, error) {
	fun, ok := scalarFuncMap[name]
	if !ok {
		return nil, fmt.Errorf("scalar function '%v' is not supported", name)
	}

	sf := &SQLScalarFunctionExpr{name, fun, exprs}

	return sf.Reconcile(), nil
}

func convertAllArgs(f *SQLScalarFunctionExpr, convType EvalType) *SQLScalarFunctionExpr {
	nExprs := convertAllExprs(f.Exprs, convType)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		nExprs,
	}
}

// NewIfScalarFunctionExpr returns a new "if" SQLScalarFunctionExpr with the
// provided components of the conditional.
func NewIfScalarFunctionExpr(condition, truePart, falsePart SQLExpr) *SQLScalarFunctionExpr {
	return &SQLScalarFunctionExpr{
		Name:  "if",
		Func:  &ifFunc{},
		Exprs: []SQLExpr{condition, truePart, falsePart},
	}
}

// areAllTimeTypes checks if all SQLValues are either type SQLTimestamp or
// SQLDate and there is at least one SQLTimestamp type. This is necessary
// because if the former is true, MySQL will always return a SQLTimestamp type
// in the greatest and least functions. i.e. SELECT GREATEST(DATE
// "2006-05-11", TIMESTAMP "2005-04-12", DATE "2004-06-04") returns TIMESTAMP
// "2006-05-11 00:00:00"
func areAllTimeTypes(values []SQLValue) (bool, bool) {
	allTimeTypes := true
	timestamp := false
	for _, v := range values {
		if _, ok := v.(SQLTimestamp); !ok {
			if _, ok := v.(SQLDate); !ok {
				allTimeTypes = false
			}
		} else {
			timestamp = true
		}
	}
	return allTimeTypes, timestamp
}

// calculateInterval converts each of the values in args to unit, and returns
// the sum of these multiplied by neg.
func calculateInterval(unit string, args []int, neg int) (string, int, error) {
	var val int
	var u string
	sp := strings.SplitAfter(unit, "_")
	if len(sp) > 1 {
		u = sp[1]
	} else {
		u = unit
	}

	const day int = 24
	const hour int = 60
	const minute int = 60
	const second int = 1000000

	switch len(args) {
	case 5:
		switch unit {
		case DayMicrosecond:
			val = args[0]*day*hour*minute*second +
				args[1]*hour*minute*second +
				args[2]*minute*second +
				args[3]*second +
				args[4]
		default:
			return unit, 0, fmt.Errorf("invalid argument length")
		}
	case 4:
		switch unit {
		case DayMicrosecond, HourMicrosecond:
			val = args[0]*hour*minute*second +
				args[1]*minute*second +
				args[2]*second +
				args[3]
		case DaySecond:
			val = args[0]*day*hour*minute +
				args[1]*hour*minute +
				args[2]*minute +
				args[3]
		default:
			return unit, 0, fmt.Errorf("invalid argument length")
		}
	case 3:
		switch unit {
		case DayMicrosecond, HourMicrosecond, MinuteMicrosecond:
			val = args[0]*minute*second + args[1]*second + args[2]
		case DaySecond, HourSecond:
			val = args[0]*hour*minute + args[1]*minute + args[2]
		case DayMinute:
			val = args[0]*day*hour + args[1]*hour + args[2]
		default:
			return unit, 0, fmt.Errorf("invalid argument length")
		}
	case 2:
		switch unit {
		case DayMicrosecond, HourMicrosecond, MinuteMicrosecond, SecondMicrosecond:
			val = args[0]*second + args[1]
		case DaySecond, HourSecond, MinuteSecond:
			val = args[0]*minute + args[1]
		case DayMinute, HourMinute:
			val = args[0]*hour + args[1]
		case DayHour:
			val = args[0]*day + args[1]
		case YearMonth:
			val = args[0]*12 + args[1]
		default:
			return unit, 0, fmt.Errorf("invalid argument length")
		}
	case 1:
		val = args[0]
	default:
		return unit, 0, fmt.Errorf("invalid argument length")
	}

	return u, val * neg, nil
}

// dateArithmeticArgs parses val and returns an integer slice stripped of any
// spaces, colons, etc. It also returns whether the first character in val is
// "-", indicating whether the arguments should be negative.
func dateArithmeticArgs(unit string, val SQLValue) ([]int, int) {
	var args []int
	neg := 1
	prev := -1
	curr := ""
	for idx, char := range val.String() {
		if idx == 0 && char == 45 {
			neg = -1
		}
		if char >= 48 && char <= 57 {
			if prev >= 48 && char <= 57 {
				curr += string(char)
			} else {
				curr = string(char)
			}
			prev = int(char)
		} else if prev != -1 {
			c, _ := strconv.Atoi(curr)
			args = append(args, c)
			curr = ""
			prev = int(char)
		}
	}
	if unit != Microsecond && strings.HasSuffix(unit, Microsecond) {
		curr = curr + strings.Repeat("0", 6-len(curr))
	}
	c, _ := strconv.Atoi(curr)
	args = append(args, c)
	return args, neg
}

func ensureArgCount(exprCount int, counts ...int) error {
	// for scalar functions that accept a variable number of arguments
	if len(counts) == 1 && counts[0] == -1 {
		if exprCount == 0 {
			return errIncorrectVarCount
		}
		return nil
	}

	found := false
	actual := exprCount
	for _, i := range counts {
		if actual == i {
			found = true
			break
		}
	}

	if !found {
		return errIncorrectCount
	}

	return nil
}

func evaluateArgs(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, exprs []SQLExpr) ([]SQLValue, error) {

	values := []SQLValue{}

	for _, expr := range exprs {
		value, err := expr.Evaluate(ctx, cfg, st)
		if err != nil {
			return []SQLValue{}, err
		}

		values = append(values, value)
	}

	return values, nil
}

func unitIntervalToMilliseconds(unit string, interval int64) (int64, error) {
	switch unit {
	case Day:
		return interval * 24 * 60 * 60 * 1000, nil
	case Hour:
		return interval * 60 * 60 * 1000, nil
	case Minute:
		return interval * 60 * 1000, nil
	case Second:
		return interval * 1000, nil
	case Microsecond:
		return interval / 1000, nil
	default:
		return 0, fmt.Errorf("cannot compute milliseconds for the unit %v", unit)
	}
}

func numMonths(startDate time.Time, endDate time.Time) int {
	y1, m1, d1 := startDate.Date()
	y2, m2, d2 := endDate.Date()
	months := ((y2 - y1) * 12) + (int(m2) - int(m1))
	if endDate.After(startDate) {
		if d2 < d1 {
			months--
		} else if d1 == d2 {
			h1, mn1, s1 := startDate.Clock()
			ns1 := startDate.Nanosecond()
			h2, mn2, s2 := endDate.Clock()
			ns2 := endDate.Nanosecond()
			t1 := time.Date(y1, m1, d1, h1, mn1, s1, ns1, schema.DefaultLocale)
			t2 := time.Date(y1, m1, d1, h2, mn2, s2, ns2, schema.DefaultLocale)
			if t1.After(t2) {
				months--
			}
		}
	} else {
		if d1 < d2 {
			months++
		} else if d1 == d2 {
			h1, mn1, s1 := startDate.Clock()
			ns1 := startDate.Nanosecond()
			h2, mn2, s2 := endDate.Clock()
			ns2 := endDate.Nanosecond()
			t1 := time.Date(y1, m1, d1, h1, mn1, s1, ns1, schema.DefaultLocale)
			t2 := time.Date(y1, m1, d1, h2, mn2, s2, ns2, schema.DefaultLocale)
			if t2.After(t1) {
				months++
			}
		}
	}
	return months
}

// nolint: unparam
func parseDateTime(s string) (time.Time, int, bool) {
	return strToDateTime(s, false)
}

func parseTime(s string) (time.Time, int, bool) {

	timeParts := strings.Split(s, ".")
	// Truncate extra decimals, e.g.: "26:11:59.23.24.25"
	// should be treated as "26:11:59.23".
	if len(timeParts) > 1 {
		s = strings.Join(timeParts[0:2], ".")
	}
	noFractions := timeParts[0]
	if len(noFractions) >= 12 {

		// Probably a datetime.
		dt, hour, ok := strToDateTime(s, true)
		if ok {
			return dt, hour, true
		}
	}

	// The result will be 0 if parsing failed, so we don't care about the result.
	dur, hour, ok := strToTime(s)

	return time.Date(0, 1, 1, 0, 0, 0, 0, schema.DefaultLocale).Add(dur), hour, ok
}

func parseDuration(v SQLValue) (time.Duration, bool) {
	buf := []byte(v.String())

	var h, m, s, f int

	hours, mins, secs, frac := []byte{}, []byte{}, []byte{}, []byte{}

	emitFrac := func(buf []byte) int {
		i := bytes.IndexByte(buf, '.')
		if i != -1 && len(frac) == 0 {
			x := 0
			for x < len(buf)-i-1 {
				idx := i + x + 1
				if buf[idx] == ':' || buf[idx] == '.' {
					break
				}
				x++
			}
			frac = buf[i+1 : i+x+1]
		}
		return i
	}

	// nolint: unparam
	emitToken := func(buf []byte, v byte) int {
		w, l := 0, len(buf)-1
		for w < l {
			if buf[w] == v {
				break
			}
			w++
		}
		return w
	}

	fmtNumeric := func(buf []byte) []byte {
		x, l, w := -1, len(buf), len(buf)
		i := bytes.IndexByte(buf, '.')
		tmp := make([]byte, w+2)
		if i != -1 {
			w = i + 2
			copy(tmp[w:], buf[i:])
		} else {
			i, w = l, l+2
		}

		for w > 0 && i > 0 {
			w, x, i = w-1, x+1, i-1
			if x%2 == 0 && x > 0 {
				tmp[w], w = ':', w-1
			}
			tmp[w] = buf[i]
			if x == 4 {
				break
			}
		}

		return append(buf[:i], tmp[w:]...)
	}

	if bytes.IndexByte(buf, ':') == -1 {
		buf = fmtNumeric(buf)
	}

	h = emitToken(buf, ':')

	if h != 0 {
		hours, buf, f = buf[0:h], buf[h+1:], emitFrac(buf[0:h+1])
		if f != -1 {
			secs, hours = hours[:f], hours[:0]
		} else {
			m = emitToken(buf, ':')
			if m != 0 {
				mins, buf, f = buf[0:m], buf[m+1:], emitFrac(buf[0:m+1])
				if f != -1 {
					mins = mins[:f]
				} else {
					s = emitToken(buf, ':')
					if s != 0 {
						secs, f = buf[0:s], emitFrac(buf[0:s+1])
						if f != -1 {
							secs = secs[:f]
						} else {
							secs = buf
						}
					}
				}
			}
		}
	}

	if len(mins) > 0 {
		if m, err := strconv.Atoi(string(mins)); err != nil || m > 60 {
			return 0, false
		}
	}

	if len(secs) > 0 {
		if s, err := strconv.ParseFloat(string(secs), 64); err != nil || s > 60 {
			return 0, false
		}
	}

	str := ""

	if len(hours) != 0 {
		str = fmt.Sprintf("%vh", string(hours))
	}

	if len(mins) != 0 {
		str = fmt.Sprintf("%v%vm", str, string(mins))
	}

	switch len(secs) {
	case 0:
		if len(frac) != 0 {
			str = fmt.Sprintf("%v0.%vs", str, string(frac))
		}
	default:
		if len(frac) != 0 {
			str = fmt.Sprintf("%v%v.%vs", str, string(secs), string(frac))
		} else {
			str = fmt.Sprintf("%v%vs", str, string(secs))
		}
	}

	dur, err := time.ParseDuration(str)
	if err != nil {
		return 0, false
	}

	return dur, true
}

// roundToDecimalPlaces rounds base to d number of decimal places.
func roundToDecimalPlaces(d int64, base float64) float64 {
	var rounded float64
	pow := math.Pow(10, float64(d))
	digit := pow * base
	_, div := math.Modf(digit)
	if base > 0 {
		if div >= 0.5 {
			rounded = math.Ceil(digit) / pow
		} else {
			rounded = math.Floor(digit) / pow
		}
	} else {
		if math.Abs(div) >= 0.5 {
			rounded = math.Floor(digit) / pow
		} else {
			rounded = math.Ceil(digit) / pow
		}
	}
	return rounded
}

func runesIndex(r, sep []rune) int {
	for i := 0; i <= len(r)-len(sep); i++ {
		found := true
		for j := 0; j < len(sep); j++ {
			if r[i+j] != sep[j] {
				found = false
				break
			}
		}

		if found {
			return i
		}
	}

	return -1
}

// handlePadding is used by the lpad and rpad functions. creates the
// specified padding string and pads the original string. padding
// goes on the left side if isLeftPad = true, on the right side otherwise.
func handlePadding(kind SQLValueKind, values []SQLValue, isLeftPad bool) (SQLValue, error) {
	if hasNullValue(values...) {
		return NewSQLNull(kind, EvalString), nil
	}

	var length int
	// length should be converted to float before we get to here
	if floatLength := Float64(values[1]); floatLength < float64(0) {
		length = int(floatLength - 0.5)
	} else {
		length = int(floatLength + 0.5)
	}

	str := []rune(values[0].String())
	padStr := []rune(values[2].String())
	padLen := length - len(str)

	// either:
	// 1) padding string is empty and the input string is not long enough to not need padding
	// 2) output length is negative and therefore impossible
	if (len(padStr) == 0 && len(str) < length) || length < 0 {
		return NewSQLNull(kind, EvalString), nil
	}

	// the string is already long enough
	if len(str) >= length {
		return NewSQLVarchar(kind, string(str[:length])), nil
	}

	// repeat padding as many times as needed to fill room
	numRepeats := math.Ceil(float64(padLen) / float64(len(padStr)))

	padding := []rune(strings.Repeat(string(padStr), int(numRepeats)))

	// in case room % len(padstr) != 0, chop off end
	padding = padding[:padLen]

	finalPad := string(padding)
	finalStr := string(str)

	if isLeftPad {
		return NewSQLVarchar(kind, finalPad+finalStr), nil
	}

	return NewSQLVarchar(kind, finalStr+finalPad), nil
}

// daysFromYearZeroCalculation calculates the number of days
// between the given year and date 0 - "0000-00-00".
func daysFromYearZeroCalculation(date time.Time) (float64, error) {
	year := date.Year()
	if year > len(yearZeroDayDifferenceSlice)-1 {
		return 0, fmt.Errorf("invalid year in date: %v", year)
	}

	// Zero out any time parts of the date.
	date = time.Date(year, date.Month(), date.Day(), 0, 0, 0, 0, schema.DefaultLocale)
	dateYearStart := time.Date(year, time.January, 1, 0, 0, 0, 0, schema.DefaultLocale)

	dayDifference := yearZeroDayDifferenceSlice[year]
	// Now add the remaining days not accounted for by the year difference.
	// The difference between "0000-01-01" and "0000-00-00" is 1 day.
	if year == 0 && date.Equal(dateYearStart) {
		dayDifference += 1
	} else {
		dayDifference += math.Trunc(date.Sub(dateYearStart).Hours() / 24.0)
	}
	return dayDifference, nil
}

// formatDate takes a time.Time object and outputs a string formatted using
// MySQL's format string specification.
func formatDate(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState, date time.Time, format string) (string, error) {
	formatRunes := []rune(format)

	noPad := func(s string) (string, error) {
		str := date.Format(s)
		if len(str) == 2 && str[0] == '0' {
			str = str[1:]
		}
		return str, nil
	}

	suffixFmt := func(i int) (string, error) {
		formatted := date.Format(strconv.Itoa(i))
		i, err := strconv.Atoi(formatted)
		if err != nil {
			return "", err
		}
		suffix := "th"
		switch i % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		}
		return formatted + suffix, nil
	}

	weekFmt := func(i int64) (string, error) {
		wf := &weekFunc{}
		args := []SQLValue{NewSQLDate(cfg.sqlValueKind, date), NewSQLInt64(cfg.sqlValueKind, i)}
		eval, err := wf.Evaluate(ctx, cfg, st, args)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%02v", eval.String()), nil
	}

	yearFmt := func(i int64) (string, error) {
		yw := &yearWeekFunc{}
		args := []SQLValue{NewSQLDate(cfg.sqlValueKind, date), NewSQLInt64(cfg.sqlValueKind, i)}
		eval, err := yw.Evaluate(ctx, cfg, st, args)
		if err != nil {
			return "", err
		}
		return eval.String()[:4], nil
	}

	zeroPad := func(s string) (string, error) {
		return fmt.Sprintf("%02v", date.Format(s)), nil
	}

	fmtTokens := map[rune]string{
		'a': "Mon",
		'b': "Jan",
		'c': "1",
		'e': "2",
		'i': "04",
		'l': "3",
		'M': "January",
		'm': "01",
		'p': "PM",
		'r': "03:04:05 PM",
		'S': "05",
		's': "05",
		'T': "15:04:05",
		'W': "Monday",
		'Y': "2006",
		'y': "06",
	}

	formatters := map[rune]func() (string, error){
		'D': func() (string, error) { return suffixFmt(2) },
		'd': func() (string, error) { return zeroPad("2") },
		'f': func() (string, error) { return date.Format(".000000")[1:], nil },
		'H': func() (string, error) { return zeroPad("15") },
		'h': func() (string, error) { return zeroPad("3") },
		'I': func() (string, error) { return zeroPad("3") },
		'j': func() (string, error) { return fmt.Sprintf("%03v", date.YearDay()), nil },
		'k': func() (string, error) { return noPad("15") },
		'U': func() (string, error) { return weekFmt(0) },
		'u': func() (string, error) { return weekFmt(1) },
		'V': func() (string, error) { return weekFmt(2) },
		'v': func() (string, error) { return weekFmt(3) },
		'w': func() (string, error) { return strconv.Itoa(int(date.Weekday())), nil },
		'X': func() (string, error) { return yearFmt(0) },
		'x': func() (string, error) { return yearFmt(1) },
		'%': func() (string, error) { return "%", nil },
	}

	for k, v := range fmtTokens {
		localV := v
		formatters[k] = func() (string, error) {
			return date.Format(localV), nil
		}
	}

	var result string
	for i := 0; i < len(formatRunes); i++ {
		if formatRunes[i] == '%' && i != len(formatRunes)-1 {
			if formatter, ok := formatters[formatRunes[i+1]]; ok {
				s, err := formatter()
				if err != nil {
					return "", err
				}
				result += s
				i++
			} else {
				result += string(formatRunes[i])
			}
		} else {
			result += string(formatRunes[i])
		}
	}

	return result, nil
}

// weekCalculation calculates the week for a given date and mode in memory.
// It is used by both the WEEK and YEARWEEK mysql scalar functions.
// Returns -1 on error. Callers should check for -1 and return proper
// default value (likely SQLNull).
func weekCalculation(date time.Time, mode int) int {

	// zeroCheck replaces results of week 0 with the week for (year-1)-12-31 for modes that
	// are 1-53 only. That means that in 1-53 modes, certain dates at the beginning of the year
	// map to week 52 or 53 of the previous year.
	zeroCheck := func(date time.Time, output, mode int) int {
		if output == 0 {
			return weekCalculation(time.Date(date.Year()-1,
				12,
				31,
				0,
				0,
				0,
				0,
				schema.DefaultLocale),
				mode)
		}
		return output
	}

	// fiftyThreeCheck is used to handle cases where the last week of a
	// year may actually map as the first week of the next year. This is
	// only possible in the cases where the first week is defined by having
	// 4 days in the year, and where 0 weeks are not allowed, so that is
	// modes 3 and 6. In these modes it is possible that 12-31, 12-30, and even
	// 12-29 map to week 1 of the next year. This is similar in design to
	// zeroCheck, except that it is only needed in the modes with 4 days
	// used to decide the first week of the month. We only need to check
	// the day if our computeDaySubtract results in week 53, giving us
	// faster common cases. janOneDaysOfWeek are the days of the week
	// for the next Jan-1 that result in one of the last three days
	// of the year potentially mapping to the next year. Note that
	// unlike MongoDB aggregation pipeline, which numbers days 1-7,
	// go time.Time numbers days 0-6, with 0 being Sunday.
	fiftyThreeCheck := func(date time.Time, output int, janOneDaysOfWeek ...int) int {
		if output == 53 {
			day := date.Day()
			nextJanOne := time.Date(date.Year()+1, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
			nextJanOneDayOfWeek := int(nextJanOne.Weekday())
			switch nextJanOneDayOfWeek {
			case janOneDaysOfWeek[0]:
				if day >= 29 {
					output = 1
				}
			case janOneDaysOfWeek[1]:
				if day >= 30 {
					output = 1
				}
			case janOneDaysOfWeek[2]:
				if day >= 31 {
					output = 1
				}
			}
		}
		return output
	}

	// computeDaySubtract computes the main week calculation shared by everything.
	// The calculation is:
	// trunc((date - dayOne) / (7 * millisecondsPerDay) + 1).
	computeDaySubtract := func(date, dayOne time.Time) int {
		return int(float64(date.Sub(dayOne))/
			(7.0*float64(millisecondsPerDay)*float64(time.Millisecond)) +
			1.0)
	}

	// computeDayInYear sets up dayOne for modes where the first week is defined
	// by having Sunday (1) or Monday (2) in the year, and computes the subtraction.
	// these modes are 0, 2, 5, 7.
	computeDayInYear := func(date time.Time, startDay, dayOfWeek int) int {
		// These are more simple than the 4 days mode. The diff from JanOne
		// can be defined using (7 - x + startDay) % 7.
		// This differs slightly from pushdown because MongoDB uses 1-7 for Sunday-Saturday
		// while go uses 0-6.
		dayOne := time.Date(date.Year(), 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
		diff := (7 - dayOfWeek + startDay) % 7
		dayOne = dayOne.Add(time.Duration(diff * int(time.Hour) * 24))
		return computeDaySubtract(date, dayOne)
	}

	// compute4DaysInYear sets up dayOne for modes where the first
	// week is defined by having 4 days in the year and computes the subtraction,
	// these are modes 1, 3, 4, and 6.
	compute4DaysInYear := func(date time.Time, startDay, dayOfWeek int) int {
		// This description is used for Monday as first day of the
		// week. See below for an explanation of the Sunday first day
		// case. Calculate the first day of the first week of this
		// year based on the dayOfWeek of YYYY-01-01 of this year, note
		// that it may be from the previous year. The Day Diff column
		// is the
		// amount of days to Add or Subtract from YYYY-01-01:
		// Day Of the Week Jan 1   |   Day Diff
		// ---------------------------------------------
		//                     0   |   + 1
		//                     1   |   + 0
		//                     2   |   - 1
		//                     3   |   - 2
		//                     4   |   - 3
		//                     5   |   + 3
		//                     6   |   + 2
		// For Sunday, we can see that 0 should be + 0, and the rest follow as expected.
		// Thus we can just add startDay since it is 0 for Sunday and 1 for Monday.
		dayOne := time.Date(date.Year(), 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
		diff := -dayOfWeek + startDay
		if diff < -3 {
			diff += 7
		}
		dayOne = dayOne.Add(time.Duration(diff * int(time.Hour) * 24))
		return computeDaySubtract(date, dayOne)
	}

	jan1 := time.Date(date.Year(), 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
	jan1DayInWeek := int(jan1.Weekday())
	switch mode {
	// First day of week: Sunday, with a Sunday in this year.
	case 0, 2:
		output := computeDayInYear(date, 0, jan1DayInWeek)
		if mode == 2 {
			output = zeroCheck(date, output, 0)
		}
		return output
	// First day of week: Monday, with 4 days in this year.
	case 1, 3:
		output := compute4DaysInYear(date, 1, jan1DayInWeek)
		if mode == 3 {
			output = zeroCheck(date, output, 1)
			output = fiftyThreeCheck(date, output, 4, 3, 2)
		}
		return output
	// First day of week: Sunday, with 4 days in this year.
	case 4, 6:
		output := compute4DaysInYear(date, 0, jan1DayInWeek)
		if mode == 6 {
			output = zeroCheck(date, output, 4)
			output = fiftyThreeCheck(date, output, 3, 2, 1)
		}
		return output
	// First day of week: Monday, with a Monday in this year.
	case 5, 7:
		output := computeDayInYear(date, 1, jan1DayInWeek)
		if mode == 7 {
			output = zeroCheck(date, output, 5)
		}
		return output
	}
	return -1
}
