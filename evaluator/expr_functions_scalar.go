package evaluator

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
)

const (
	shortTimeFormat = "2006-01-02"
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
	// MaxGoDurationHours is the largest value of the maximum time.Duration.Hours()
	MaxGoDurationHours = 2562024.0
	// MillisecondsPerDay is the number of milliseconds in a day.
	MillisecondsPerDay = 8.64e+7
	// SecondsPerDay is the number of seconds in a day.
	SecondsPerDay = 8.64e+4
	// SecondsPerHour is the number of seconds in an hour.
	SecondsPerHour = 3600.0
	// SecondsPerMinute is the number of seconds in an minute.
	SecondsPerMinute = 60.0
)

var toMilliseconds = map[string]float64{
	Week:        MillisecondsPerDay * 7,
	Day:         MillisecondsPerDay,
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
	"nullif":      &nullifFunc{},
	"pi":          &constantFunc{SQLFloat(math.Pi)},
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
	Normalize(*SQLScalarFunctionExpr) SQLExpr
}

// reconcilingScalarFunc is an interface for a Scalar Function
// that can be type reconciled.
type reconcilingScalarFunc interface {
	Reconcile(*SQLScalarFunctionExpr) *SQLScalarFunctionExpr
}

type scalarFunc interface {
	Evaluate([]SQLValue, *EvalCtx) (SQLValue, error)
	Validate(exprCount int) error
	Type([]SQLExpr) schema.SQLType
}

// translatableToAggregationScalarFunc is an interface for a Scalar Function
// that can be translated to MongoDB Aggregation Language.
type translatableToAggregationScalarFunc interface {
	FuncToAggregationLanguage(*PushDownTranslator, []SQLExpr) (interface{}, bool)
}

//
// SQLScalarFunctionExpr represents a scalar function.
//
type SQLScalarFunctionExpr struct {
	Name  string
	Func  scalarFunc
	Exprs []SQLExpr
}

// Evaluate evaluates a SQLScalarFunctionExpr to a SQLValue.
func (f *SQLScalarFunctionExpr) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	err := f.Func.Validate(len(f.Exprs))
	if err != nil {
		return SQLNull, fmt.Errorf("%v '%v'", err.Error(), f.Name)
	}

	values, err := evaluateArgs(f.Exprs, ctx)
	if err != nil {
		return SQLNull, err
	}
	return f.Func.Evaluate(values, ctx)
}

// Normalize will attempt to change SQLScalarFunctionExpr into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (f *SQLScalarFunctionExpr) Normalize() Node {
	if nsf, ok := f.Func.(normalizingScalarFunc); ok {
		return nsf.Normalize(f)
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

// RequiresEvalCtx will check if the SQLScalarFunctionExpr requires an evaluation context.
func (f *SQLScalarFunctionExpr) RequiresEvalCtx() bool {
	if r, ok := f.Func.(RequiresEvalCtx); ok {
		return r.RequiresEvalCtx()
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
// it will return nil and false.
func (f *SQLScalarFunctionExpr) ToAggregationLanguage(
	t *PushDownTranslator) (interface{}, bool) {
	if fun, ok := f.Func.(translatableToAggregationScalarFunc); ok {
		return fun.FuncToAggregationLanguage(t, f.Exprs)
	}
	t.Ctx.Logger().Debugf(log.Dev,
		"%q cannot be pushed down as an aggregate expression at this time",
		f.Name)
	return nil, false
}

// Type returns the SQLType associated with the SQLScalarFunctionExpr.
func (f *SQLScalarFunctionExpr) Type() schema.SQLType {
	return f.Func.Type(f.Exprs)
}

type absFunc struct {
	singleArgFloatMathFunc
}

func (*absFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return bson.M{"$abs": args[0]}, true
}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_acos
type acosFunc struct {
	singleArgFloatMathFunc
}

func (*acosFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	input := "$$input"
	letAssignment := bson.M{
		"input": args[0],
	}

	// MySQL returns NULL for values outside of the range [-1,1].
	// asin x + acos x = pi/2
	return wrapInLet(letAssignment,
		wrapInCond(nil,
			wrapInAcosComputation(input),
			wrapInOp(mgoOperatorLt, input, -1.0),
			wrapInOp(mgoOperatorGt, input, 1.0),
		),
	), true
}

type addDateFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_adddate
func (*addDateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	adder := &dateAddFunc{}
	return adder.Evaluate(values, ctx)
}

func (*addDateFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}
	return f
}

func (*addDateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLTimestamp
}

func (*addDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type asciiFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_ascii
func (*asciiFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	str := values[0].String()
	if str == "" {
		return SQLInt(0), nil
	}

	c := str[0]

	return SQLInt(c), nil
}

func (*asciiFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*asciiFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_asin
type asinFunc struct {
	singleArgFloatMathFunc
}

func (*asinFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	input := "$$input"
	letAssignment := bson.M{
		"input": args[0],
	}

	// MySQL returns NULL for values outside of the range [-1,1].
	// asin(x) =  pi/2 - cos(x) via the identity:
	// asin(x) + acos(x) = pi/2.
	return wrapInLet(letAssignment,
		wrapInCond(nil,
			wrapInOp(mgoOperatorSubtract, math.Pi/2.0, wrapInAcosComputation(input)),
			wrapInOp(mgoOperatorLt, input, -1.0),
			wrapInOp(mgoOperatorGt, input, 1.0),
		),
	), true
}

type ceilFunc struct {
	singleArgFloatMathFunc
}

func (*ceilFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return bson.M{mgoOperatorCeil: args[0]}, true
}

type charFunc struct{}

func (*charFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {

	var b []byte
	for _, i := range values {
		if i == SQLNull {
			continue
		}
		v := i.Int64()
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

	return SQLVarchar(string(b)), nil
}

func (*charFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*charFunc) Validate(exprCount int) error {
	if exprCount == 0 {
		return errIncorrectCount
	}

	return nil
}

type characterLengthFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_char_length
func (*characterLengthFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	value := []rune(values[0].String())

	return SQLInt(len(value)), nil
}

func (*characterLengthFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapSingleArgFuncWithNullCheck("$strLenCP", args[0]), true
}

func (*characterLengthFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNone)
}

func (*characterLengthFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*characterLengthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type coalesceFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_coalesce
func (*coalesceFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	for _, value := range values {
		if value != SQLNull {
			return value, nil
		}
	}
	return SQLNull, nil
}

func (*coalesceFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	var coalesce func([]interface{}) interface{}
	coalesce = func(args []interface{}) interface{} {
		if len(args) == 0 {
			return nil
		}
		replacement := coalesce(args[1:])
		return bson.M{mgoOperatorIfNull: []interface{}{args[0], replacement}}
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return coalesce(args), true
}

func (*coalesceFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return f
}

func (*coalesceFunc) Type(exprs []SQLExpr) schema.SQLType {
	sorter := &schema.SQLTypesSorter{VarcharHighPriority: true}
	return preferentialTypeWithSorter(sorter, exprs...)
}

func (*coalesceFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, -1)
}

type concatFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_concat
func (*concatFunc) Evaluate(
	values []SQLValue,
	ctx *EvalCtx) (v SQLValue, err error) {
	if hasNullValue(values...) {
		v = SQLNull
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

	v = SQLVarchar(b.String())
	err = nil
	return
}

func (*concatFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) < 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return bson.M{mgoOperatorConcat: args}, true
}

func (*concatFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*concatFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNone)
}

func (*concatFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*concatFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, -1)
}

type concatWsFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_concat-ws
func (*concatWsFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (v SQLValue, err error) {
	if _, ok := values[0].(SQLNullValue); ok {
		v = SQLNull
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
		if _, ok := value.(SQLNullValue); ok {
			continue
		}
		b.WriteString(value.String())
		if i != len(trimValues)-1 {
			b.WriteString(separator)
		}
	}

	v = SQLVarchar(b.String())
	return
}

func (*concatWsFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) < 2 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	var pushArgs []interface{}

	for _, value := range args[1:] {
		pushArgs = append(pushArgs,
			bson.M{mgoOperatorCond: []interface{}{
				bson.M{mgoOperatorEq: []interface{}{
					bson.M{mgoOperatorIfNull: []interface{}{value, nil}},
					nil}},
				wrapInLiteral(""), value}},
			bson.M{mgoOperatorCond: []interface{}{
				bson.M{mgoOperatorEq: []interface{}{
					bson.M{mgoOperatorIfNull: []interface{}{value, nil}},
					nil}},
				wrapInLiteral(""), args[0]}})
	}

	return bson.M{mgoOperatorConcat: pushArgs[:len(pushArgs)-1]}, true
}

func (*concatWsFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if len(f.Exprs) >= 2 && f.Exprs[0] == SQLNull {
		return SQLNull
	}

	return f
}

func (*concatWsFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNone)
}

func (*concatWsFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*concatWsFunc) Validate(exprCount int) error {
	if ensureArgCount(exprCount, -1) != nil || exprCount < 2 {
		return errIncorrectCount
	}
	return nil
}

type connectionIDFunc struct{}

func (*connectionIDFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLUint32(ctx.ExecutionCtx.ConnectionID()), nil
}

func (*connectionIDFunc) RequiresEvalCtx() bool {
	return true
}

func (*connectionIDFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*connectionIDFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type constantFunc struct {
	value SQLValue
}

func (c *constantFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return c.value, nil
}

func (c *constantFunc) Type(exprs []SQLExpr) schema.SQLType {
	return c.value.Type()
}

func (*constantFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type convFunc struct{}

// https://dev.mysql.com/doc/refman/8.0/en/mathematical-functions.html#function_conv
// Diverges from MySQL behavior in its handling of negative values
// Converts bases to positive numbers, and returns a negative value if the input is negative
// MySQL claims that "If from_base is a negative number, N is regarded as a signed number.
// Otherwise, N is treated as unsigned." Manual testing shows that it returns the 2's
// complement version if the number is negative unless the to_base is also negative, in which
// case it returns the number with a negative sign at the front
func (*convFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	num := values[0].String()
	originalBase := absInt64(values[1].Int64())
	newBase := absInt64(values[2].Int64())
	negative := false

	if baseIsInvalid(originalBase) || baseIsInvalid(newBase) {
		return SQLNull, nil
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
		return SQLVarchar("0"), nil
	}
	strVersion := strconv.FormatInt(base10Version, int(newBase))

	if negative && strVersion != "0" {
		strVersion = "-" + strVersion
	}

	return SQLVarchar(strings.ToUpper(strVersion)), nil
}

func (*convFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	if v, ok := f.Exprs[1].(SQLValue); ok {
		if baseIsInvalid(absInt64(v.Int64())) {
			return SQLNull
		}
	}

	if v, ok := f.Exprs[2].(SQLValue); ok {
		if baseIsInvalid(absInt64(v.Int64())) {
			return SQLNull
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

func (*convFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{}, bool) {

	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}
	num := args[0]
	oldBase := args[1]
	newBase := args[2]

	// length is how long (in digits) the input number is
	normalizedVars := bson.M{
		"originalBase": wrapInOp(mgoOperatorAbs, oldBase),
		"newBase":      wrapInOp(mgoOperatorAbs, newBase),
		"negative":     wrapInOp(mgoOperatorEq, "-", wrapInOp(mgoOperatorSubstr, num, 0, 1)),
		"nonNegativeNumber": wrapInCond(
			wrapInOp(mgoOperatorSubstr, num, 1,
				wrapInOp(mgoOperatorSubtract, wrapInOp(mgoOperatorStrlenCP, num), 1)),
			num,
			wrapInOp(mgoOperatorEq, "-", wrapInOp(mgoOperatorSubstr, num, 0, 1))),
	}

	indexOfDecimal := bson.M{
		"decimalIndex": wrapInOp(mgoOperatorIndexOfCP, "$$nonNegativeNumber", "."),
	}

	eliminateDecimal := bson.M{
		"number": wrapInCond("$$nonNegativeNumber",
			wrapInOp(mgoOperatorSubstr, "$$nonNegativeNumber", 0, "$$decimalIndex"),
			wrapInOp(mgoOperatorEq, "$$decimalIndex", -1)),
	}

	createLength := bson.M{
		"length": wrapInOp(mgoOperatorStrlenCP, "$$number"),
	}

	// indexArr is an array of numbers from 0 to n-1 when n = length
	createIndexArr := bson.M{
		"indexArr": wrapInOp(mgoOperatorRange, 0, "$$length", 1),
	}

	// charArr breaks the number entered into an array of characters where each char is a digit
	createCharArr := bson.M{
		"charArr": wrapInMap("$$indexArr", "this",
			[]interface{}{"$$this", wrapInOp(mgoOperatorSubstr, "$$number", "$$this", 1)}),
	}

	// This logic takes in the charArr and outputs a 2D array containing the index and the
	// base10 numerical value of the character.
	// i.e. if charArr = ["3", "A", "2"], numArr = [[0, 3], [1, 10], [2, 2]]
	branches1 := make([]bson.M, 0)
	for _, k := range validNumbers {
		branches1 = append(branches1,
			wrapInCase(
				wrapInOp(mgoOperatorEq,
					wrapInOp(mgoOperatorArrElemAt, "$$this", 1),
					k,
				),
				[]interface{}{
					wrapInOp(mgoOperatorArrElemAt, "$$this", 0),
					stringToNum[k],
				},
			),
		)
	}
	createNumArr := bson.M{
		"numArr": bson.M{
			mgoOperatorMap: bson.M{
				"input": "$$charArr",
				"in":    wrapInSwitch([]interface{}{0, 100}, branches1...),
			},
		},
	}

	// invalidArr has False for every digit that is valid, and True for every digit that is invalid
	// In order for the input string to be converted to a new number base every entry in this
	// array must be False.
	createInvalidArr := bson.M{
		"invalidArr": wrapInMap(
			"$$numArr",
			"this",
			wrapInOp(mgoOperatorGte, wrapInOp(mgoOperatorArrElemAt, "$$this", 1), "$$originalBase"),
		),
	}

	// Given a charArr = [[1, x1]...[i, xi]...[n, xn]] and a base b,
	// This implements the logic: sum(b^(n-i-1) * xi) with i = 0->n-1
	generateBase10 := bson.M{
		"base10": wrapInOp(mgoOperatorSum,
			wrapInMap("$$numArr", "this",
				wrapInOp(mgoOperatorMultiply,
					wrapInOp(mgoOperatorArrElemAt, "$$this", 1),
					wrapInOp(mgoOperatorPow, "$$originalBase",
						wrapInOp(mgoOperatorSubtract,
							wrapInOp(mgoOperatorSubtract, "$$length",
								wrapInOp(mgoOperatorArrElemAt, "$$this", 0)),
							1))))),
	}

	// numDigits is the length the number will be in the new number base
	// This is equal to: floor(log_newbase(num)) + 1
	numDigits := bson.M{
		"numDigits": wrapInOp(mgoOperatorAdd,
			wrapInOp(mgoOperatorFloor,
				wrapInOp(mgoOperatorLog, "$$base10", "$$newBase")), 1),
	}

	// powers is an array of the powers of the base that you are translating to
	// if the newBase=16 and the resulting number will have length=4 this array
	// will = [1, 16, 256, 4096]
	powers := bson.M{
		"powers": wrapInMap(
			wrapInOp(mgoOperatorRange, wrapInOp(mgoOperatorSubtract, "$$numDigits", 1), -1, -1),
			"this",
			wrapInOp(mgoOperatorPow, "$$newBase", "$$this")),
	}

	// Turns the base10 number into an array of the newBase digits (in their base10 form)
	// i.e. if base10 = 173 (0xAD), numbersArray = [10, 13]
	// Follows generalized version of: https://www.permadi.com/tutorial/numDecToHex/
	generateNumberArray := wrapInMap("$$powers", "this",
		wrapInOp(mgoOperatorMod,
			wrapInOp(mgoOperatorFloor,
				wrapInOp(mgoOperatorDivide, "$$base10", "$$this")), "$$newBase"))

	branches2 := make([]bson.M, 0)
	for k := 0; k <= len(numToString); k++ {
		branches2 = append(branches2,
			wrapInCase(wrapInOp(mgoOperatorEq, "$$this", k), numToString[k]))
	}

	// Converts the number array into an array of their character representations
	// i.e. if numbersArray = [10, 13], then charArray=['A', 'D']
	generateCharArray := wrapInMap(generateNumberArray, "this", wrapInSwitch("0", branches2...))

	// Turns the charArray into a single string (the final answer)
	// i.e. if charArray=['A','D'] answer='AD'
	positiveAnswer := bson.M{
		"positiveAnswer": wrapInReduce(
			generateCharArray,
			"",
			wrapInOp(mgoOperatorConcat, "", "$$value", "$$this"),
		),
	}

	signAdjusted := wrapInCond(wrapInOp(mgoOperatorConcat, "-", "$$positiveAnswer"),
		"$$positiveAnswer", "$$negative")

	// Puts the nested lets together, checks to make sure that the base is valid,
	// and checks to make sure the entered number is valid as well
	// (invalid = numbers too big like 3 in binary or non-alphanumeric like /)
	// Invalid characters returns an answer of 0, invalid bases return NULL
	return wrapInCond(nil, wrapInLet(normalizedVars,
		wrapInLet(indexOfDecimal,
			wrapInLet(eliminateDecimal,
				wrapInCond(nil,
					wrapInCond("0",
						wrapInLet(createLength,
							wrapInLet(createIndexArr,
								wrapInLet(createCharArr,
									wrapInLet(createNumArr,
										wrapInLet(createInvalidArr,
											wrapInCond("0",
												wrapInLet(generateBase10,
													wrapInLet(numDigits,
														wrapInLet(powers,
															wrapInLet(positiveAnswer,
																signAdjusted)))),
												wrapInOp(mgoOperatorAnyElementTrue,
													"$$invalidArr"))))))),
						wrapInOp(mgoOperatorIn, "$$number", []interface{}{"0", "-0"})),
					wrapInOp(mgoOperatorOr,
						wrapInOp(mgoOperatorOr, wrapInOp(mgoOperatorLt, "$$originalBase", 2),
							wrapInOp(mgoOperatorGt, "$$originalBase", 36)),
						wrapInOp(mgoOperatorOr, wrapInOp(mgoOperatorLt, "$$newBase", 2),
							wrapInOp(mgoOperatorGt, "$$newBase", 36)))))),
	), wrapInOp(mgoOperatorEq, nil, num)), true
}

func (*convFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*convFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

func (*convFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLInt, schema.SQLInt}
	defaults := []SQLValue{SQLVarchar("0"), SQLInt(0), SQLInt(0)}
	nExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		nExprs,
	}
}

type convertFunc struct{}

func sqlTypeFromSQLExpr(expr SQLExpr) (schema.SQLType, bool) {
	val, ok := expr.(SQLValue)
	if !ok {
		return schema.SQLNone, false
	}

	var typ schema.SQLType
	switch val.String() {
	case string(parser.SIGNED_BYTES):
		typ = schema.SQLInt
	case string(parser.UNSIGNED_BYTES):
		typ = schema.SQLUint64
	case string(parser.FLOAT_BYTES):
		typ = schema.SQLFloat
	case string(parser.CHAR_BYTES):
		typ = schema.SQLVarchar
	case string(parser.DATE_BYTES):
		typ = schema.SQLDate
	case string(parser.DATETIME_BYTES):
		typ = schema.SQLTimestamp
	case string(parser.DECIMAL_BYTES):
		typ = schema.SQLDecimal128
	case string(parser.TIME_BYTES):
		typ = schema.SQLTimestamp
	default:
		return schema.SQLNone, false
	}

	return typ, true
}

// http://dev.mysql.com/doc/refman/5.7/en/cast-functions.html#function_convert
func (*convertFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	_, ok := values[0].(SQLNullValue)
	if ok {
		return SQLNull, nil
	}

	typ, ok := sqlTypeFromSQLExpr(values[1])
	if !ok {
		return SQLNull, nil
	}

	return NewSQLConvertExpr(values[0], typ, SQLNone).Evaluate(ctx)
}

func (conv *convertFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {

	typ, ok := sqlTypeFromSQLExpr(exprs[1])
	if !ok {
		return nil, false
	}

	return NewSQLConvertExpr(exprs[0], typ, SQLNone).ToAggregationLanguage(t)
}

func (*convertFunc) Type(exprs []SQLExpr) schema.SQLType {
	typ, _ := sqlTypeFromSQLExpr(exprs[1])
	return typ
}

func (*convertFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_cos
type cosFunc struct {
	singleArgFloatMathFunc
}

func (*cosFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	input := "$$input"
	inputLetAssignment := bson.M{
		"input": wrapInOp(mgoOperatorAbs, args[0]),
	}
	rem, phase := "$$rem", "$$phase"
	remPhaseAssignment := bson.M{
		"rem": wrapInOp(mgoOperatorMod, input, math.Pi/2),
		"phase": wrapInOp(mgoOperatorMod,
			wrapInOp(mgoOperatorTrunc,
				wrapInOp(mgoOperatorDivide, input, math.Pi/2),
			),
			4.0),
	}

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
	threeCase := wrapInCond(wrapInSinPowerSeries(rem),
		nil,
		wrapInOp(mgoOperatorEq,
			phase,
			3))
	twoCase := wrapInCond(wrapInOp(mgoOperatorMultiply,
		-1.0,
		wrapInCosPowerSeries(rem)),
		threeCase,
		wrapInOp(mgoOperatorEq,
			phase,
			2))
	oneCase := wrapInCond(wrapInOp(mgoOperatorMultiply,
		-1.0,
		wrapInSinPowerSeries(rem)),
		twoCase,
		wrapInOp(mgoOperatorEq,
			phase,
			1))
	zeroCase := wrapInCond(wrapInCosPowerSeries(rem),
		oneCase,
		wrapInOp(mgoOperatorEq,
			phase,
			0))

	return wrapInLet(inputLetAssignment,
		wrapInLet(remPhaseAssignment,
			zeroCase),
	), true
}

type cotFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_cot
func (*cotFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	tan := math.Tan(values[0].Float64())
	if tan == 0 {
		return SQLNull,
			mysqlerrors.Defaultf(mysqlerrors.ErDataOutOfRange,
				"DOUBLE",
				fmt.Sprintf("'cot(%v)'",
					values[0].Float64()))
	}

	return SQLFloat(1 / tan), nil
}

func (*cotFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	sf := &sinFunc{}
	denom, pushedDown := sf.FuncToAggregationLanguage(t, []SQLExpr{exprs[0]})
	if !pushedDown {
		return nil, false
	}

	cf := &cosFunc{}
	num, pushedDown := cf.FuncToAggregationLanguage(t, []SQLExpr{exprs[0]})
	if !pushedDown {
		return nil, false
	}

	// epsilon the smallest value we allow for denom, computed to roughly
	// tie-out with mysqld.
	epsilon := 6.123233995736766e-17
	// Replace abs(denom) < epsilon with epsilon since mysql doesn't seem
	// to return NULL or INF for inf values of tan as one would expect, and
	// we do not want to trigger a $divide by 0.
	return wrapInOp(mgoOperatorDivide,
		num,
		wrapInCond(epsilon,
			denom,
			wrapInOp(mgoOperatorLte,
				wrapInOp(mgoOperatorAbs, denom), epsilon,
			),
		),
	), true
}

func (*cotFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLFloat
}

func (*cotFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type currentDateFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_curdate
func (*currentDateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	now := time.Now().In(schema.DefaultLocale)
	t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, schema.DefaultLocale)
	return SQLDate{t}, nil

}

func (*currentDateFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	now := time.Now().In(schema.DefaultLocale)
	cd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, schema.DefaultLocale)
	return wrapInLiteral(cd), true
}

func (*currentDateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLDate
}

func (*currentDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type currentTimestampFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_now
func (*currentTimestampFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	value := time.Now().In(schema.DefaultLocale)
	return SQLTimestamp{value}, nil
}

func (*currentTimestampFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	now := time.Now().In(schema.DefaultLocale)
	return wrapInLiteral(now), true
}

func (*currentTimestampFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLTimestamp
}

func (*currentTimestampFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type curtimeFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_curtime
func (*curtimeFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLTimestamp{time.Now().In(schema.DefaultLocale)}, nil
}

func (*curtimeFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLTimestamp
}

func (*curtimeFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type dateAddFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_date-add
func (*dateAddFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	_, _, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	timestampadd := &timestampAddFunc{}
	args, neg := dateArithmeticArgs(values[2].String(), values[1])
	unit, interval, err := calculateInterval(values[2].String(), args, neg)
	if err != nil {
		return SQLNull, nil
	}

	return timestampadd.Evaluate([]SQLValue{SQLVarchar(unit), SQLInt(interval), values[0]}, ctx)
}

func (*dateAddFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*dateAddFunc) Type(exprs []SQLExpr) schema.SQLType {
	if exprs[0].Type() == schema.SQLTimestamp {
		return schema.SQLTimestamp
	}

	if exprs[0].Type() == schema.SQLDate {
		if unit, ok := exprs[2].(SQLValue); ok {
			switch unit.String() {
			case Hour, Minute, Second:
				return schema.SQLTimestamp
			}
		}
	}

	return schema.SQLVarchar
}

func (*dateAddFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type dateArithmeticFunc struct {
	scalarFunc
	isSub bool
}

func (f *dateArithmeticFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 3 {
		return nil, false
	}

	var date interface{}
	var ok bool
	if _, ok = f.scalarFunc.(*addDateFunc); ok {
		// implementation for ADDDATE(DATE_FORMAT("..."), INTERVAL 0 SECOND)
		var fun *SQLScalarFunctionExpr
		if fun, ok = exprs[0].(*SQLScalarFunctionExpr); ok && fun.Name == "date_format" {
			if date, ok = t.translateDateFormatAsDate(fun); !ok {
				date = nil
			}
		}
	}

	if date == nil {
		switch exprs[0].Type() {
		case schema.SQLDate, schema.SQLTimestamp:
		default:
			return nil, false
		}

		if date, ok = t.ToAggregationLanguage(exprs[0]); !ok {
			return nil, false
		}
	}

	intervalValue, ok := exprs[1].(SQLValue)
	if !ok {
		return nil, false
	}

	if intervalValue.Float64() == 0 {
		return date, true
	}

	unitValue, ok := exprs[2].(SQLValue)
	if !ok {
		return nil, false
	}

	unitInterval, neg := dateArithmeticArgs(unitValue.String(), intervalValue)
	unit, interval, err := calculateInterval(unitValue.String(), unitInterval, neg)
	if err != nil {
		return nil, false
	}

	ms, err := unitIntervalToMilliseconds(unit, int64(interval))
	if err != nil {
		return nil, false
	}

	if f.isSub {
		ms *= -1
	}

	letAssignment := bson.M{
		"date": date,
	}

	letEvaluation := wrapInNullCheckedCond(
		nil,
		wrapInOp(mgoOperatorAdd, "$$date", ms),
		"$$date",
	)
	return wrapInLet(letAssignment, letEvaluation), true

}

func (*dateArithmeticFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}
	return f
}

type dateDiffFunc struct{}

func (*dateDiffFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {

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
		return SQLNull, nil
	}

	if right, ok = parseArgs(values[1]); !ok {
		return SQLNull, nil
	}

	durationDiff := left.Sub(right)
	hoursDiff := durationDiff.Hours()
	daysDiff := hoursDiff / 24

	diff := SQLInt(int(daysDiff))
	return diff, nil
}

func (*dateDiffFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 2 {
		return nil, false
	}

	var date1, date2 interface{}
	var ok bool

	parseArgs := func(expr SQLExpr) (interface{}, bool) {
		var value SQLValue
		if value, ok = expr.(SQLValue); ok {
			var date time.Time
			date, _, ok = strToDateTime(value.String(), false)
			if !ok {
				return nil, false
			}

			date = time.Date(date.Year(),
				date.Month(),
				date.Day(),
				0,
				0,
				0,
				0,
				schema.DefaultLocale)
			return date, true
		}
		exprType := expr.Type()
		if exprType == schema.SQLTimestamp || exprType == schema.SQLDate {
			var date interface{}
			date, ok = t.ToAggregationLanguage(expr)
			if !ok {
				return nil, false
			}
			return date, true
		}
		return nil, false
	}

	if date1, ok = parseArgs(exprs[0]); !ok {
		return nil, false
	}
	if date2, ok = parseArgs(exprs[1]); !ok {
		return nil, false
	}

	// This division needs to truncate because this is dateDiff not
	// timestampDiff, partial days are dropped.
	days := wrapInOp(mgoOperatorTrunc, wrapInOp(mgoOperatorDivide,
		wrapInOp(mgoOperatorSubtract, date1, date2), 86400000))
	bound := wrapInCond(106751, -106751, wrapInOp(mgoOperatorGt, days, 106751))

	letAssignment := bson.M{
		"days": days,
	}

	letEvaluation := wrapInNullCheckedCond(
		nil,
		wrapInCond(
			bound,
			"$$days",
			wrapInOp(mgoOperatorGt, "$$days", 106751),
			wrapInOp(mgoOperatorLt, "$$days", -106751),
		),
		date1,
		date2,
	)
	return wrapInLet(letAssignment, letEvaluation), true

}

func (*dateDiffFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}
	return f
}

func (*dateDiffFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*dateDiffFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type dateFormatFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_date-format
func (*dateFormatFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	date, _, ok := parseDateTime(values[0].String())
	date = date.In(schema.DefaultLocale)
	if !ok {
		return SQLNull, nil
	}

	v1, ok := values[1].(SQLVarchar)
	if !ok {
		return SQLNull, nil
	}

	ret, err := formatDate(date, v1.String(), ctx)
	if err != nil {
		return nil, err
	}
	return SQLVarchar(ret), nil
}

func (*dateFormatFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 2 {
		return nil, false
	}

	date, ok := t.ToAggregationLanguage(exprs[0])
	if !ok {
		return nil, false
	}

	formatValue, ok := exprs[1].(SQLValue)
	if !ok {
		return nil, false
	}

	return wrapInDateFormat(date, formatValue.String())
}

func (*dateFormatFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*dateFormatFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*dateFormatFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type dateFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_date
func (*dateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	// Too-short numbers are padded differently than too-short strings.
	// strToDateTime (called by parseDateTime) handles padding in the too-short
	// string case. We need to fix the string here, where we can still find out
	// the original input type.
	var str string
	switch values[0].(type) {
	case SQLFloat, SQLDecimal128, SQLInt:
		noDecimal := strings.Split(values[0].String(), ".")[0]
		intLength := len(noDecimal)
		if intLength > 14 {
			return SQLNull, nil
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
		return SQLNull, nil
	}

	return SQLDate{Time: t.Truncate(24 * time.Hour)}, nil
}

func (*dateFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (df *dateFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if !t.Ctx.VersionAtLeast(3, 5, 0) {
		return nil, false
	}

	if len(exprs) != 1 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	val := "$$val"
	inputLet := bson.M{
		"val": args[0],
	}

	wrapInDateFromString := func(v interface{}) bson.M {
		return bson.M{mgoOperatorDateFromString: bson.M{"dateString": v}}
	}

	// CASE 1: it's already a Mongo date, we just return it.
	isDateType := containsBSONType(val, "date")
	dateBranch := wrapInCase(isDateType, val)

	// CASE 2: it's a number.
	isNumber := containsBSONType(val, "int", "decimal", "long", "double")

	// Evaluates to true if val positive and has <= X digits.
	hasUpToXDigits := func(x float64) interface{} {
		return wrapInInRange(val, 0, math.Pow(10, x))
	}

	// This handles converting a number in YYMMDD format to YYYYMMDD.
	// if YY < 70, we assume they meant 20YY. if YY > 70, we assume 19YY.
	getPadding := func(v interface{}) interface{} {
		return wrapInCond(
			20000000,
			19000000,
			wrapInOp(mgoOperatorLt,
				wrapInOp(mgoOperatorDivide,
					v, 10000),
				70))
	}

	// We interpret this as being format YYMMDD.
	ifSix := wrapInOp(mgoOperatorAdd, val, getPadding(val))
	sixBranch := wrapInCase(hasUpToXDigits(6), ifSix)

	// This number is good as is! YYYYMMDD.
	eightBranch := wrapInCase(hasUpToXDigits(8), val)

	// If it's twelve digits, interpret as YYMMDDHHMMSS.
	// first drop the last six digits, then pad like we would a six digit number.
	firstSixDigits := bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide, val, 1000000)}
	ifTwelve := wrapInOp(mgoOperatorAdd, firstSixDigits, getPadding(firstSixDigits))
	twelveBranch := wrapInCase(hasUpToXDigits(12), ifTwelve)

	// If fourteen, YYYYMMDDHHMMSS. just drop the last six digits.
	ifFourteen := bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide, val, 1000000)}
	fourteenBranch := wrapInCase(hasUpToXDigits(14), ifFourteen)

	// Define "num", the input number normalized to 8 digits, in a "let".
	numberVar := wrapInSwitch(nil, sixBranch, eightBranch, twelveBranch, fourteenBranch)
	numberLetVars := bson.M{"num": numberVar}

	dateParts := bson.M{
		// YYYYMMDD / 10000 = YYYY.
		"year": bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide, "$$num", 10000)},
		// (YYYYMMDD / 100) % 100 = MM.
		"month": wrapInOp(
			mgoOperatorMod,
			bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide,
				"$$num", 100),
			},
			100),
		// YYYYMMDD % 100 = DD.
		"day": wrapInOp(mgoOperatorMod, "$$num", 100),
	}

	// Try to avoid aggregation errors by catching obviously invalid dates.
	yearValid := wrapInInRange("$$year", 0, 10000)
	monthValid := wrapInInRange("$$month", 1, 13)
	dayValid := wrapInInRange("$$day", 1, 32)

	makeDateOrNull := wrapInCond(
		bson.M{mgoOperatorDateFromParts: bson.M{
			"year":  "$$year",
			"month": "$$month",
			"day":   "$$day",
		}},
		nil,
		bson.M{mgoOperatorAnd: []interface{}{yearValid, monthValid, dayValid}},
	)

	evaluateNumber := wrapInLet(dateParts, makeDateOrNull)
	handleNumberToDate := wrapInLet(numberLetVars, evaluateNumber)
	numberBranch := wrapInCase(isNumber, handleNumberToDate)

	// CASE 3: it's a string
	isString := containsBSONType(val, "string")

	// First split on T, take first substring, then split that on " ", and take first
	// substring. this gives us just the date part of the string. note that if the
	// string doesn't have T or a space, just returns original string.
	trimmedString := wrapInOp(mgoOperatorArrElemAt,
		wrapInOp(mgoOperatorSplit,
			wrapInOp(mgoOperatorArrElemAt,
				wrapInOp(mgoOperatorSplit, val, "T"),
				0),
			" "),
		0)

	// Convert the string to an array so we can use map/reduce.
	trimmedAsArray := wrapInStringToArray("$$trimmed")

	// isSeparator evaluates to true if a character is in the defined separator list.
	isSeparator := wrapInOp(mgoOperatorNeq,
		-1,
		wrapInOp("$indexOfArray",
			dateComponentSeparator,
			"$$c"))

	// Use map to convert all separators in the string to - symbol, and leave numbers as-is.
	separatorsNormalized := wrapInMap(trimmedAsArray,
		"c",
		wrapInCond("-",
			"$$c",
			isSeparator))

	// Use reduce to convert characters back to a single string
	joined := wrapInReduce(separatorsNormalized,
		"",
		wrapInOp(mgoOperatorConcat,
			"$$value",
			"$$this"))

	// If the third character is a -, or if the string is only 6 digits
	// long and has no slashes, then the string is either format YY-MM-DD
	// or YYMMDD and we need to add the appropriate first two year digits
	// (19xx or 20xx) for MongoDB to understand it
	hasShortYear := wrapInOp(mgoOperatorOr,
		// Length is only 6, assume YYMMDD.
		wrapInOp(mgoOperatorEq, bson.M{mgoOperatorStrlenCP: "$$joined"}, 6),
		// Third character is -, assume YY-MM-DD.
		wrapInOp(mgoOperatorEq,
			"-",
			bson.M{mgoOperatorSubstr: []interface{}{"$$joined",
				2,
				1}}))

	// $dateFromString actually pads correctly, but not if "/" is used as
	// the separator (it will assume year is last). If this pushdown is
	// shown to be slow by benchmarks, we should reconsider allowing
	// $dateFromString to handle padding. The change would not be trivial
	// due to how MongoDB cannot handle short dates when there are no
	// separators in the date.
	padYear := wrapInOp(mgoOperatorConcat,
		wrapInCond(
			"20",
			"19",
			// Check if first two digits < 70 to determine padding.
			wrapInOp(
				mgoOperatorLt,
				bson.M{mgoOperatorSubstr: []interface{}{"$$joined", 0, 2}},
				"70")),
		"$$joined")

	// We have to use nested $lets because in the outer one we define $$trimmed and
	// in the inner one we define $$joined. defining $$joined requires knowing the
	// length of trimmed, so we can't do it all in one step.
	innerIn := wrapInCond(padYear, "$$joined", hasShortYear)
	innerLet := wrapInLet(bson.M{"joined": joined}, innerIn)

	// Gracefully handle strings that are too short to possibly be valid by returning null.
	tooShort := wrapInOp(mgoOperatorLt, bson.M{mgoOperatorStrlenCP: "$$trimmed"}, 6)
	outerIn := wrapInCond(nil, wrapInDateFromString(innerLet), tooShort)
	outerLet := wrapInLet(bson.M{"trimmed": trimmedString}, outerIn)

	// Make sure if we get the int 0 we return NULL instead
	// of crashing. MySQL uses '0000-00-00' as an error output for some
	// functions and we encode it as the integer 0 within push down.
	stringBranch := wrapInCase(isString,
		wrapInCond(nil,
			outerLet,
			wrapInOp(mgoOperatorEq,
				0,
				args[0])))

	return wrapInLet(inputLet,
			wrapInSwitch(nil,
				dateBranch,
				numberBranch,
				stringBranch)),
		true

}

func (*dateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLDate
}

func (*dateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dateSubFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_date-sub
func (*dateSubFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	dateadd := &dateAddFunc{}

	v := values[1].String()
	if string(v[0]) != "-" {
		v = "-" + v
	} else {
		v = v[1:]
	}

	return dateadd.Evaluate([]SQLValue{values[0], SQLVarchar(v), values[2]}, ctx)
}

func (*dateSubFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*dateSubFunc) Type(exprs []SQLExpr) schema.SQLType {
	return (&dateAddFunc{}).Type(exprs)
}

func (*dateSubFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type dayNameFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayname
func (*dayNameFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLVarchar(t.Weekday().String()), nil
}

func (*dayNameFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapInNullCheckedCond(
		nil,
		bson.M{mgoOperatorArrElemAt: []interface{}{
			[]interface{}{
				time.Sunday.String(),
				time.Monday.String(),
				time.Tuesday.String(),
				time.Wednesday.String(),
				time.Thursday.String(),
				time.Friday.String(),
				time.Saturday.String(),
			},
			bson.M{mgoOperatorSubtract: []interface{}{
				bson.M{"$dayOfWeek": args[0]},
				1}}}},
		args[0],
	), true
}

func (*dayNameFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (*dayNameFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*dayNameFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dayOfMonthFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayofmonth
func (*dayOfMonthFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(t.Day()), nil
}

func (*dayOfMonthFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapSingleArgFuncWithNullCheck("$dayOfMonth", args[0]), true
}

func (*dayOfMonthFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (*dayOfMonthFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*dayOfMonthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dayOfWeekFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayofweek
func (*dayOfWeekFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Weekday()) + 1), nil
}

func (*dayOfWeekFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapSingleArgFuncWithNullCheck("$dayOfWeek", args[0]), true
}

func (*dayOfWeekFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (*dayOfWeekFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*dayOfWeekFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dayOfYearFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayofyear
func (*dayOfYearFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(t.YearDay()), nil
}

func (*dayOfYearFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapSingleArgFuncWithNullCheck("$dayOfYear", args[0]), true
}

func (*dayOfYearFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (*dayOfYearFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*dayOfYearFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dbFunc struct{}

func (*dbFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLVarchar(ctx.ExecutionCtx.DB()), nil
}

func (*dbFunc) RequiresEvalCtx() bool {
	return true
}

func (*dbFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
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

func (*degreesFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapInOp(mgoOperatorDivide, wrapInOp(mgoOperatorMultiply, args[0], 180.0), math.Pi), true
}

type dualArgFloatMathFunc func(float64, float64) float64

func (f dualArgFloatMathFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	result := f(values[0].Float64(), values[1].Float64())
	if math.IsNaN(result) {
		return SQLNull, nil
	}
	if math.IsInf(result, 0) {
		return SQLNull, nil
	}
	if result == -0 {
		result = 0
	}
	return SQLFloat(result), nil
}

func (dualArgFloatMathFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (dualArgFloatMathFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	switch f.Name {
	case "mod":
		return convertAllArgs(f, schema.SQLFloat, SQLNone)
	default:
		return f
	}
}

func (dualArgFloatMathFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLFloat
}

func (dualArgFloatMathFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type eltFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_elt
func (*eltFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {

	if hasNullValue(values[0]) {
		return SQLNull, nil
	}

	idx := values[0].Int64()
	if idx <= 0 || int(idx) >= len(values) {
		return SQLNull, nil
	}

	result := values[idx]
	if result == SQLNull {
		return SQLNull, nil
	}

	return SQLVarchar(result.String()), nil
}

func (*eltFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	elems := args[1:]
	index := "$$index"
	// Note: ELT indexes on 1, while arrayElemAt indexes based on 0, so we need to subtract 1.
	return wrapInLet(
		bson.M{
			"index": args[0],
		},
		wrapInCond(nil,
			bson.M{
				mgoOperatorArrElemAt: []interface{}{elems, wrapInOp(mgoOperatorSubtract, index, 1)},
			},
			wrapInOp(mgoOperatorLte, index, 0),
		),
	), true
}

func (*eltFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs[0]) {
		return SQLNull
	}

	if v, ok := f.Exprs[0].(SQLValue); ok {
		idx := v.Int64()
		if idx <= 0 || int(idx) > len(f.Exprs) {
			return SQLNull
		}
	}

	return f
}

func (*eltFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
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

func (f *expFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return bson.M{"$exp": args[0]}, true
}

type extractFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_extract
func (*extractFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, _, ok := parseDateTime(values[1].String())
	if !ok {
		return SQLNull, nil
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
		return SQLInt(units[0]), nil
	case Quarter:
		return SQLInt(int(math.Ceil(float64(units[1]) / 3.0))), nil
	case Month:
		return SQLInt(units[1]), nil
	case Week:
		_, w := t.ISOWeek()
		return SQLInt(w), nil
	case Day:
		return SQLInt(units[2]), nil
	case Hour:
		return SQLInt(units[3]), nil
	case Minute:
		return SQLInt(units[4]), nil
	case Second:
		return SQLInt(units[5]), nil
	case Microsecond:
		return SQLInt(0), nil
	case YearMonth:
		ym, _ := strconv.Atoi(unitStrs[0] + unitStrs[1])
		return SQLInt(ym), nil
	case DayHour:
		dh, _ := strconv.Atoi(unitStrs[2] + unitStrs[3])
		return SQLInt(dh), nil
	case DayMinute:
		dm, _ := strconv.Atoi(unitStrs[2] + unitStrs[3] + unitStrs[4])
		return SQLInt(dm), nil
	case DaySecond:
		ds, _ := strconv.Atoi(unitStrs[2] + unitStrs[3] + unitStrs[4] + unitStrs[5])
		return SQLInt(ds), nil
	case DayMicrosecond:
		dms, _ := strconv.Atoi(unitStrs[2] + unitStrs[3] + unitStrs[4] + unitStrs[5] + "000000")
		return SQLInt(dms), nil
	case HourMinute:
		hm, _ := strconv.Atoi(unitStrs[3] + unitStrs[4])
		return SQLInt(hm), nil
	case HourSecond:
		hs, _ := strconv.Atoi(unitStrs[3] + unitStrs[4] + unitStrs[5])
		return SQLInt(hs), nil
	case HourMicrosecond:
		hms, _ := strconv.Atoi(unitStrs[3] + unitStrs[4] + unitStrs[5] + "000000")
		return SQLInt(hms), nil
	case MinuteSecond:
		ms, _ := strconv.Atoi(unitStrs[4] + unitStrs[5])
		return SQLInt(ms), nil
	case MinuteMicrosecond:
		mms, _ := strconv.Atoi(unitStrs[4] + unitStrs[5] + "000000")
		return SQLInt(mms), nil
	case SecondMicrosecond:
		sms, _ := strconv.Atoi(unitStrs[5] + "000000")
		return SQLInt(sms), nil
	default:
		return SQLNull, fmt.Errorf("unit type '%v' is not supported", values[0].String())
	}
}

func (*extractFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if len(exprs) != 2 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	bsonMap, ok := args[0].(bson.M)
	if !ok {
		return nil, false
	}

	bsonVal, ok := bsonMap["$literal"]
	if !ok {
		return nil, false
	}

	unitVal, _ := NewSQLValue(bsonVal, schema.SQLVarchar, schema.SQLNone)

	unit := unitVal.String()

	switch unit {
	case "year", "month", "hour", "minute", "second":
		return wrapSingleArgFuncWithNullCheck("$"+unit, args[1]), true
	case "day":
		return wrapSingleArgFuncWithNullCheck("$dayOfMonth", args[1]), true
	}
	return nil, false
}

func (*extractFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLNone, schema.SQLTimestamp}
	defaults := []SQLValue{SQLNone, SQLNone}
	nExprs := convertExprs(f.Exprs, argTypes, defaults)
	// Do not use constructor here, we already have a valid f.Func to use
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		nExprs,
	}
}

func (*extractFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*extractFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type fieldFunc struct{}

func (*fieldFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLInt(0), nil
	}

	target := values[0]
	candidates := values[1:]

	for idx, candidate := range candidates {
		if candidate.Value() == target.Value() {
			return SQLInt(idx + 1), nil
		}
	}
	return SQLInt(0), nil
}

func (*fieldFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLInt(0)
	}

	return f
}

func (*fieldFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	var reconcile bool
	var firstType schema.SQLType
loop:
	for _, expr := range f.Exprs {
		typ := expr.Type()
		switch typ {
		case schema.SQLVarchar, schema.SQLInt, schema.SQLInt64,
			schema.SQLDecimal128, schema.SQLFloat, schema.SQLNumeric:
			// valid types
		default:
			reconcile = true
			break loop
		}
		if firstType == schema.SQLNone {
			firstType = typ
			continue
		}
		if firstType != typ {
			reconcile = true
			break loop
		}
	}

	if reconcile {
		return convertAllArgs(f, schema.SQLDecimal128, SQLNone)
	}
	return f
}

func (*fieldFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*fieldFunc) Validate(exprCount int) error {
	if exprCount <= 1 {
		return errIncorrectVarCount
	}
	return nil
}

func (*fieldFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {

	if len(exprs) <= 1 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	var anyArgNull []interface{}
	for _, arg := range args {
		isNull := wrapInOp(mgoOperatorEq, wrapInIfNull(arg, nil), nil)
		anyArgNull = append(anyArgNull, isNull)
	}

	target := args[0]
	candidates := args[1:]

	var cases []interface{}
	var results []interface{}
	for idx, candidate := range candidates {
		caseExpr := wrapInOp(mgoOperatorEq, target, candidate)
		resultExpr := wrapInLiteral(idx + 1)

		cases = append(cases, caseExpr)
		results = append(results, resultExpr)
	}

	var idxSwitch interface{}

	if t.Ctx.VersionAtLeast(3, 4, 0) {
		var branches []bson.M
		for idx, caseExpr := range cases {
			resultExpr := results[idx]
			branch := bson.M{"case": caseExpr, "then": resultExpr}
			branches = append(branches, branch)
		}
		idxSwitch = wrapInSwitch(wrapInLiteral(0), branches...)
	} else {
		var lastTerm interface{} = wrapInLiteral(0)

		numTerms := len(cases)
		for idx := numTerms - 1; idx >= 0; idx-- {
			term := wrapInCond(results[idx], lastTerm, cases[idx])
			lastTerm = term
		}

		idxSwitch = lastTerm
	}

	return wrapInCond(wrapInLiteral(0), idxSwitch, anyArgNull...), true
}

type floorFunc struct {
	singleArgFloatMathFunc
}

func (*floorFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return bson.M{"$floor": args[0]}, true
}

type fromDaysFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_from-days
func (*fromDaysFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
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
		return SQLVarchar("0000-00-00"), nil
	}
	if neg {
		value = -value
	}

	if value <= 365.5 || value >= 3652499.5 {
		// Go's zero time starts January 1, year 1, 00:00:00 UTC
		// and thus can not represent the date "0000-00-00". To
		// handle this, we return a varchar instead
		return SQLVarchar("0000-00-00"), nil
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

	return SQLDate{date.In(schema.DefaultLocale)}, nil
}

func (*fromDaysFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}
	dayOne := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	body := wrapInOp(mgoOperatorAdd, dayOne,
		wrapInOp(mgoOperatorMultiply, wrapInRoundValue(args[0]), MillisecondsPerDay))
	arg := "$$arg"
	argLetAssignment := bson.M{
		"arg": args[0],
	}
	// This should return "0000-00-00" if the input is too large (> maxFromDays)
	// or too low (< 366).
	return wrapInLet(argLetAssignment, wrapInCond(nil,
		wrapInCond(0,
			body,
			wrapInOp(mgoOperatorGt, arg, maxFromDays),
			wrapInOp(mgoOperatorLt, arg, 366),
		),
		wrapInNullCheck(arg),
	),
	), true
}

func (*fromDaysFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*fromDaysFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLInt, SQLNone)
}

func (*fromDaysFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLDate
}

func (*fromDaysFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type fromUnixtimeFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_from-unixtime
func (*fromUnixtimeFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	value := round(values[0].Float64())
	if value < 0 {
		return SQLNull, nil
	}

	date := time.Unix(value, 0).In(schema.DefaultLocale)
	if len(values) == 1 {
		return SQLTimestamp{Time: date}, nil
	}
	ret, err := formatDate(date, values[1].String(), ctx)
	if err != nil {
		return nil, err
	}
	return SQLVarchar(ret), nil
}

func (*fromUnixtimeFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if len(exprs) > 2 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	arg := "$$arg"
	letAssignment := bson.M{
		"arg": args[0],
	}

	// Just add the argument to 1970-01-01 00:00:00.0000000.
	dayOne := time.Date(1970, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
	letEvaluation := wrapInOp(mgoOperatorAdd,
		dayOne,
		wrapInOp(mgoOperatorMultiply,
			wrapInRoundValue(arg),
			1e3))

	ret := wrapInLet(letAssignment,
		wrapInCond(nil,
			letEvaluation,
			wrapInOp(mgoOperatorLt, arg, wrapInLiteral(0)),
		),
	)

	if len(exprs) == 1 {
		return ret, true
	}
	if format, ok := exprs[1].(SQLValue); ok {
		return wrapInDateFormat(ret, format.String())
	}
	return nil, false
}

func (*fromUnixtimeFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*fromUnixtimeFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLInt, schema.SQLVarchar}
	defaults := []SQLValue{SQLNone, SQLNone}
	nExprs := convertExprs(f.Exprs, argTypes, defaults)
	// Do not use constructor here, we already have a valid f.Func to use
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		nExprs,
	}
}

func (*fromUnixtimeFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLTimestamp
}

func (*fromUnixtimeFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

type greatestFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_greatest
func (*greatestFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	valExprs := []SQLExpr{}
	for _, val := range values {
		valExprs = append(valExprs, val)
	}
	convertTo := preferentialType(valExprs...)

	convertedVals := []SQLValue{}
	for _, val := range values {
		newVal := convertType(val, convertTo)
		convertedVals = append(convertedVals, newVal)
	}

	var greatest SQLValue
	var greatestIdx int

	c, err := CompareTo(convertedVals[0], convertedVals[1], ctx.Collation)
	if err != nil {
		return SQLNull, err
	}

	if c == -1 {
		greatest, greatestIdx = values[1], 1
	} else {
		greatest, greatestIdx = values[0], 0
	}

	for i := 2; i < len(values); i++ {
		c, err = CompareTo(greatest, convertedVals[i], ctx.Collation)
		if err != nil {
			return SQLNull, err
		}
		if c == -1 {
			greatest, greatestIdx = values[i], i
		}
	}

	allTimeVals, timestamp := areAllTimeTypes(values)
	if allTimeVals && timestamp {
		t, _, _ := parseDateTime(values[greatestIdx].String())
		return SQLTimestamp{Time: t}, nil
	} else if convertTo == schema.SQLDate || convertTo == schema.SQLTimestamp {
		return values[greatestIdx], nil
	}

	return convertedVals[greatestIdx], nil
}

func (*greatestFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	// we can only push down if the types are similar
	for i := 1; i < len(exprs); i++ {
		if !isSimilar(exprs[0].Type(), exprs[i].Type()) {
			return nil, false
		}
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapInNullCheckedCond(
		nil,
		bson.M{"$max": args},
		args...,
	), true
}

func (*greatestFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*greatestFunc) Type(exprs []SQLExpr) schema.SQLType {
	return preferentialType(exprs...)
}

func (*greatestFunc) Validate(exprCount int) error {
	if exprCount < 2 {
		return errIncorrectVarCount
	}
	return nil
}

type hourFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_hour
func (*hourFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if values[0] == SQLNull {
		return SQLNull, nil
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
			return SQLNull, nil
		}
		return SQLInt(0), nil
	}

	return SQLInt(hour), nil
}

func (*hourFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapSingleArgFuncWithNullCheck("$hour", args[0]), true
}

func (*hourFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*hourFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*hourFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type ifFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/control-flow-functions.html#function_if
func (*ifFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	switch typedV := values[0].(type) {
	case SQLBool:
		if typedV.Bool() {
			return values[1], nil
		}
		return values[2], nil
	case SQLDate, SQLTimestamp, SQLObjectID:
		return values[1], nil
	case SQLInt, SQLFloat:
		v := typedV.Float64()
		if v == 0 {
			return values[2], nil
		}
		return values[1], nil
	case SQLNullValue:
		return values[2], nil
	case SQLVarchar:
		if v, _ := strconv.ParseFloat(typedV.String(), 64); v == 0 {
			return values[2], nil
		}
		return values[1], nil
	default:
		return SQLNull, fmt.Errorf("expression type '%v' is not supported", typedV)
	}
}

func (*ifFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if len(exprs) != 3 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"expr": args[0],
	}

	letEvaluation := wrapInCond(
		args[2],
		args[1],
		wrapInNullCheck("$$expr"),
		wrapInOp(mgoOperatorEq, "$$expr", 0),
		wrapInOp(mgoOperatorEq, "$$expr", false),
	)

	return wrapInLet(letAssignment, letEvaluation), true

}

func (*ifFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return f
}

func (*ifFunc) Type(exprs []SQLExpr) schema.SQLType {
	s := &schema.SQLTypesSorter{VarcharHighPriority: true}
	return preferentialTypeWithSorter(s, exprs[1:]...)
}

func (*ifFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type ifnullFunc struct{}

func (*ifnullFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if _, ok := values[0].(SQLNullValue); ok {
		return values[1], nil
	}
	return values[0], nil
}

func (*ifnullFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if len(exprs) != 2 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapInIfNull(args[0], args[1]), true
}

func (*ifnullFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if f.Exprs[0] == SQLNull {
		return f.Exprs[1]
	} else if v, ok := f.Exprs[0].(SQLValue); ok {
		return v
	}

	return f
}

func (*ifnullFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return f
}

func (*ifnullFunc) Type(exprs []SQLExpr) schema.SQLType {
	s := &schema.SQLTypesSorter{VarcharHighPriority: true}
	return preferentialTypeWithSorter(s, exprs...)
}

func (*ifnullFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type isnullFunc struct{}

func (*isnullFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	s := NewSQLIsExpr(values[0], SQLNull)
	return s.Evaluate(ctx)
}

func (*isnullFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	s := NewSQLIsExpr(exprs[0], SQLNull)
	return s.ToAggregationLanguage(t)
}

func (*isnullFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*isnullFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type insertFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_insert
func (*insertFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {

	if hasNullValue(values...) {
		return SQLNull, nil
	}

	s := values[0].String()
	pos := int(round(values[1].Float64())) - 1
	length := int(round(values[2].Float64()))
	newstr := values[3].String()

	if pos < 0 || pos >= len(s) {
		return values[0], nil
	}

	if pos+length < 0 || pos+length > len(s) {
		return SQLVarchar(s[:pos] + newstr), nil
	}

	return SQLVarchar(s[:pos] + newstr + s[pos+length:]), nil
}

func (*insertFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}

	if len(exprs) != 4 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	str, pos, len, newstr := "$$str", "$$pos", "$$len", "$$newstr"
	inputAssignment := bson.M{
		"str": args[0],
		// SQL uses 1 indexing, so makes sure to subtract 1 to
		// account for MongoDB's 0 indexing.
		"pos":    wrapInRoundValue(wrapInOp(mgoOperatorSubtract, args[1], 1)),
		"len":    wrapInRoundValue(args[2]),
		"newstr": args[3],
	}

	totalLength := "$$totalLength"
	totalLengthAssignment := bson.M{
		"totalLength": wrapInOp(mgoOperatorStrlenCP, str),
	}

	prefix, suffix := "$$prefix", "$$suffix"
	ixAssignment := bson.M{
		"prefix": wrapInOp(mgoOperatorSubstr, str, 0, pos),
		"suffix": wrapInOp(mgoOperatorSubstr, str, wrapInOp(mgoOperatorAdd, pos, len), totalLength),
	}

	concatenation := wrapInLet(ixAssignment,
		wrapInOp(mgoOperatorConcat, prefix, newstr, suffix),
	)

	posCheck := wrapInLet(totalLengthAssignment,
		wrapInCond(str,
			concatenation,
			wrapInOp(mgoOperatorLte, pos, 0),
			wrapInOp(mgoOperatorGte, pos, totalLength),
		),
	)

	return wrapInLet(inputAssignment,
		wrapInCond(nil,
			posCheck,
			wrapInOp(mgoOperatorLte, str, nil),
			wrapInOp(mgoOperatorLte, pos, nil),
			wrapInOp(mgoOperatorLte, len, nil),
			wrapInOp(mgoOperatorLte, newstr, nil),
		),
	), true
}

func (*insertFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*insertFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLInt, schema.SQLInt, schema.SQLVarchar}
	defaults := []SQLValue{SQLNone, SQLNone, SQLNone, SQLNone}
	convertedExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*insertFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*insertFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 4)
}

type instrFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_instr
func (*instrFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	locate := &locateFunc{}
	return locate.Evaluate([]SQLValue{values[1], values[0]}, ctx)
}

func (*instrFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}

	if len(exprs) != 2 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	// Mongo Aggregation Pipeline returns NULL if arg1 is NULLish, like
	// we'd want. arg2 being NULL, however, is an error in the pipeline,
	// thus check arg2 for NULLisness.
	arg2 := "$$arg2"
	return wrapInLet(bson.M{
		"arg2": args[1],
	},
		wrapInCond(nil,
			wrapInOp(mgoOperatorAdd,
				wrapInOp(mgoOperatorIndexOfCP, args[0], arg2),
				1,
			),
			wrapInOp(mgoOperatorLte, arg2, nil),
		),
	), true
}

func (*instrFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*instrFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNone)
}

func (*instrFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*instrFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type intervalFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_interval
func (*intervalFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if values[0].Type() == schema.SQLNull {
		return SQLInt(-1), nil
	}

	start, end := 1, len(values)-1
	for start != end {
		mid := (start + end + 1) / 2
		if values[mid].Type() == schema.SQLNull || values[mid].Float64() <= values[0].Float64() {
			start = mid
		} else {
			end = mid - 1
		}
	}

	if values[start].Type() == schema.SQLNull || values[start].Float64() <= values[0].Float64() {
		return SQLInt(start), nil
	}
	return SQLInt(start - 1), nil
}

func (*intervalFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}
	return wrapInCond(
		wrapInLiteral(-1),
		bson.M{
			mgoOperatorReduce: bson.M{
				"input":        args[1:],
				"initialValue": wrapInLiteral(0),
				"in": wrapInCond(
					bson.M{mgoOperatorAdd: []interface{}{"$$value", wrapInLiteral(1)}},
					"$$value",
					bson.M{mgoOperatorGte: []interface{}{args[0], "$$this"}},
				),
			},
		},
		wrapInNullCheck(args[0]),
	), true
}

func (*intervalFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if f.Exprs[0].Type() == schema.SQLNull {
		return SQLInt(-1)
	}
	return f
}

func (*intervalFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLFloat, SQLNone)
}

func (*intervalFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt64
}

func (*intervalFunc) Validate(exprCount int) error {
	if exprCount < 2 {
		return errIncorrectVarCount
	}
	return nil
}

type lastDayFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_last-day
func (*lastDayFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
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
	t := tmp.Time
	year, month, _ := t.Date()
	first := time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
	return SQLDate{first.AddDate(0, 1, -1)}, nil
}

func (*lastDayFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*lastDayFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if len(exprs) != 1 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	date := "$$date"
	letAssignment := bson.M{
		"date": args[0],
	}

	year, month := "$$year", "$$month"
	innerLetAssigment := bson.M{
		"year":  wrapInOp(mgoOperatorYear, date),
		"month": wrapInOp(mgoOperatorMonth, date),
	}

	// This is the template that we will use to construct a date from parts using
	// $dateFromParts.
	template := bson.M{
		"year":  year,
		"month": month,
		"day":
		// The following MongoDB aggregation language implements this go code,
		// which is designed to set the day of a date to the last day of the month.
		// switch m {
		// case 2:
		// 	if isLeapYear(y) == 0 {
		// 		d = 29
		//	} else {
		//		d = 28
		//	}
		// case 4, 6, 9, 11:
		//	d = 30
		// default:
		//      d = 31
		// }
		wrapInSwitch(31,
			wrapInEqCase(month, 2,
				wrapInCond(29, 28, wrapInIsLeapYear(year)),
			),
			wrapInEqCase(month, 4, 30),
			wrapInEqCase(month, 6, 30),
			wrapInEqCase(month, 9, 30),
			wrapInEqCase(month, 11, 30),
		),
	}

	return wrapInLet(letAssignment,
		wrapInLet(innerLetAssigment,
			bson.M{mgoOperatorDateFromParts: template}),
	), true
}

func (*lastDayFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLDate
}

func (*lastDayFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type lcaseFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_lcase
func (*lcaseFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	value := strings.ToLower(values[0].String())

	return SQLVarchar(value), nil
}

func (*lcaseFunc) FuncToAggregationLanguage(
	t *PushDownTranslator,
	exprs []SQLExpr) (interface{},
	bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapSingleArgFuncWithNullCheck("$toLower", args[0]), true
}

func (*lcaseFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNone)
}

func (*lcaseFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*lcaseFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type leastFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_least
func (*leastFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	valExprs := []SQLExpr{}
	for _, val := range values {
		valExprs = append(valExprs, val)
	}
	convertTo := preferentialType(valExprs...)

	convertedVals := []SQLValue{}
	for _, val := range values {
		newVal := convertType(val, convertTo)
		convertedVals = append(convertedVals, newVal)
	}

	var least SQLValue
	var leastIdx int

	c, err := CompareTo(convertedVals[0], convertedVals[1], ctx.Collation)
	if err != nil {
		return SQLNull, err
	}

	if c == -1 {
		least, leastIdx = convertedVals[0], 0
	} else {
		least, leastIdx = convertedVals[1], 1
	}

	for i := 2; i < len(values); i++ {
		c, err = CompareTo(least, convertedVals[i], ctx.Collation)
		if err != nil {
			return SQLNull, err
		}
		if c == 1 {
			least, leastIdx = values[i], i
		}
	}

	allTimeVals, timestamp := areAllTimeTypes(values)
	if allTimeVals && timestamp {
		t, _, _ := parseDateTime(values[leastIdx].String())
		return SQLTimestamp{Time: t}, nil
	} else if convertTo == schema.SQLDate || convertTo == schema.SQLTimestamp {
		return values[leastIdx], nil
	}

	return convertedVals[leastIdx], nil
}

func (*leastFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	// we can only push down if the types are similar
	for i := 1; i < len(exprs); i++ {
		if !isSimilar(exprs[0].Type(), exprs[i].Type()) {
			return nil, false
		}
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapInNullCheckedCond(
		nil,
		bson.M{"$min": args},
		args...,
	), true

}

func (*leastFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*leastFunc) Type(exprs []SQLExpr) schema.SQLType {
	return preferentialType(exprs...)
}

func (*leastFunc) Validate(exprCount int) error {
	if exprCount < 2 {
		return errIncorrectVarCount
	}
	return nil
}

type leftFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_left
func (*leftFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	substring,
		err := NewSQLScalarFunctionExpr("substring",
		[]SQLExpr{values[0],
			SQLInt(1),
			values[1]})
	if err != nil {
		return SQLNull, err
	}
	return substring.Evaluate(ctx)
}

func (*leftFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {

	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}
	if len(exprs) != 2 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"string": args[0],
		"length": args[1],
	}

	// when length is negative, just use 0. round length to closest integer
	subStrLength, _ := wrapInRound([]interface{}{wrapInOp(mgoOperatorMax, "$$length", 0)})

	subStrOp := bson.M{mgoOperatorSubstr: []interface{}{"$$string", 0, subStrLength}}

	letEvaluation := wrapInNullCheckedCond(nil, subStrOp, "$$string", "$$length")
	return wrapInLet(letAssignment, letEvaluation), true

}

func (*leftFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLInt}
	defaults := []SQLValue{SQLNone, SQLNone}
	nExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		nExprs,
	}
}

func (*leftFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*leftFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type lengthFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_length
func (*lengthFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	value := values[0].String()

	return SQLInt(len(value)), nil
}

func (*lengthFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapSingleArgFuncWithNullCheck("$strLenBytes", args[0]), true
}

func (*lengthFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNone)
}

func (*lengthFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*lengthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type locateFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_locate
func (*locateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values[:2]...) {
		return SQLNull, nil
	}

	substr := []rune(values[0].String())
	str := []rune(values[1].String())
	var result int
	if len(values) == 3 {

		pos := int(values[2].Float64()+0.5) - 1 // MySQL uses 1 as a basis

		if pos < 0 || len(str) <= pos {
			return SQLInt(0), nil
		}
		str = str[pos:]
		result = runesIndex(str, substr)
		if result >= 0 {
			result += pos
		}
	} else {
		result = runesIndex(str, substr)
	}

	return SQLInt(result + 1), nil
}

func (*locateFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}
	if !(len(exprs) == 2 || len(exprs) == 3) {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	var locate interface{}
	substr := args[0]
	str := args[1]

	if len(args) == 2 {
		indexOfCP := bson.M{"$indexOfCP": []interface{}{str, substr}}
		locate = wrapInOp(mgoOperatorAdd, indexOfCP, 1)
	} else if len(args) == 3 {
		// if the pos arg is null, we should return 0, not null
		// this is the same result as when the arg is 0
		pos := wrapInIfNull(args[2], 0)

		// round to the nearest int
		pos = wrapInOp(mgoOperatorAdd, pos, 0.5)
		pos = wrapInOp(mgoOperatorTrunc, pos)

		// subtract 1 from the pos arg to reconcile indexing style
		pos = wrapInOp(mgoOperatorSubtract, pos, 1)

		indexOfCP := bson.M{"$indexOfCP": []interface{}{str, substr, pos}}
		locate = wrapInOp(mgoOperatorAdd, indexOfCP, 1)

		// if the pos argument was negative, we should return 0
		locate = wrapInCond(
			0,
			locate,
			wrapInOp(mgoOperatorLt, pos, 0),
		)
	}

	return wrapInNullCheckedCond(
		nil,
		locate,
		str, substr,
	), true
}

func (*locateFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs[:2]...) {
		return SQLNull
	}

	return f
}

func (*locateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*locateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2, 3)
}

type logFunc struct {
	Base uint32 // 0 for natural log.
}

func (f logFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	var result float64
	switch f.Base {
	case 0:
		if len(values) == 2 {
			// arbitrary base
			result = math.Log(values[1].Float64()) / math.Log(values[0].Float64())
		} else {
			// natural base
			result = math.Log(values[0].Float64())
		}
	case 2:
		result = math.Log2(values[0].Float64())
	case 10:
		result = math.Log10(values[0].Float64())
	}
	if math.IsNaN(result) {
		return SQLNull, nil
	}
	if math.IsInf(result, 0) {
		return SQLNull, nil
	}
	if result == -0 {
		result = 0
	}
	return SQLFloat(result), nil
}

func (f *logFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) < 1 || len(exprs) > 2 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	// Use ln func rather than log with go's value for E, to avoid compromising values
	// more than we already do between MongoDB and MySQL by introducing a third value for E
	// (i.e., go's)
	if f.Base == 0 {
		// 1 arg implies natural log
		if len(args) == 1 {
			return bson.M{mgoOperatorCond: []interface{}{
				bson.M{mgoOperatorGt: []interface{}{args[0], 0}},
				bson.M{mgoOperatorNaturalLog: args[0]},
				mgoNullLiteral}}, true
		}
		// Two args is based arg.
		// MySQL specifies base then arg, MongoDB expects arg then base, so we have to flip.
		return bson.M{mgoOperatorCond: []interface{}{
			bson.M{mgoOperatorGt: []interface{}{args[0], 0}},
			bson.M{mgoOperatorLog: []interface{}{args[1], args[0]}},
			mgoNullLiteral}}, true
	}
	// This will be base 10 or base 2 based on if log10 or log2 was called.
	return bson.M{mgoOperatorCond: []interface{}{
		bson.M{mgoOperatorGt: []interface{}{args[0], 0}},
		bson.M{mgoOperatorLog: []interface{}{args[0], f.Base}},
		mgoNullLiteral}}, true
}

func (logFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}
	return f
}

func (logFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLFloat, SQLNone)
}

func (logFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLFloat
}

func (f logFunc) Validate(exprCount int) error {
	if f.Base == 0 {
		return ensureArgCount(exprCount, 1, 2)
	}
	return ensureArgCount(exprCount, 1)
}

type lpadFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_lpad
func (*lpadFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return handlePadding(values, true)
}

func (*lpadFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*lpadFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLInt, schema.SQLVarchar}
	defaults := []SQLValue{SQLNull, SQLNull, SQLNull}
	convertedExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*lpadFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*lpadFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type ltrimFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_ltrim
func (*ltrimFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	value := strings.TrimLeft(values[0].String(), " ")

	return SQLVarchar(value), nil
}

func (*ltrimFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	if t.Ctx.VersionAtLeast(4, 0, 0) {
		return bson.M{
			mgoOperatorLTrim: bson.M{
				"input": args[0],
				"chars": " ",
			},
		}, true
	}

	ltrimCond := wrapInCond(
		"",
		wrapLRTrim(true, args[0]),
		bson.M{mgoOperatorEq: []interface{}{args[0], ""}},
	)

	return wrapInNullCheckedCond(
		nil,
		ltrimCond,
		args[0],
	), true
}

func (*ltrimFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNull)
}

func (*ltrimFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*ltrimFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type makeDateFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_makedate
func (*makeDateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	// Floating arguments should be rounded.
	y := round(values[0].Float64())
	if y < 0 || y > 9999 {
		return SQLNull, nil
	}
	if y >= 0 && y <= 69 {
		y += 2000
	} else if y >= 70 && y <= 99 {
		y += 1900
	}

	d := round(values[1].Float64())

	if d <= 0 {
		return SQLNull, nil
	}

	t := time.Date(int(y), 1, 0, 0, 0, 0, 0, schema.DefaultLocale)
	duration := time.Duration(d*24) * time.Hour

	output := t.Add(duration)
	if output.Year() > 9999 {
		return SQLNull, nil
	}

	return SQLDate{Time: output}, nil
}

func (*makeDateFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 5, 0) {
		return nil, false
	}

	if len(exprs) != 2 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	year, day, paddedYear, output := "$$year", "$$day", "$$paddedYear", "$$output"

	inputLetStatement := bson.M{
		"year": wrapInRoundValue(args[0]),
		"day":  wrapInRoundValue(args[1]),
	}

	branch1900 := wrapInCond(
		wrapInOp(mgoOperatorAdd, year, 1900),
		year,
		wrapInOp(mgoOperatorAnd,
			wrapInOp(mgoOperatorGte, year, 70),
			wrapInOp(mgoOperatorLte, year, 99),
		))

	branch2000 := wrapInOp(mgoOperatorAdd, year, 2000)

	// $$paddedYear holds the year + 2000 for years between 0 and 69, and +
	// 1900 for years between 70 and 99. Otherwise, it is the original
	// year.
	paddedYearLetStatement := bson.M{"paddedYear": wrapInCond(branch2000, branch1900,
		wrapInOp(mgoOperatorAnd,
			wrapInOp(mgoOperatorGte,
				year,
				0),
			wrapInOp(mgoOperatorLte,
				year,
				69)),
	)}

	// This implements:
	// date(paddedYear) + (day - 1) * MillisecondsPerDay.
	addDaysStatement := wrapInOp(mgoOperatorAdd,
		bson.M{mgoOperatorDateFromParts: bson.M{"year": paddedYear}},
		wrapInOp(mgoOperatorMultiply,
			wrapInOp(mgoOperatorSubtract, day, 1),
			MillisecondsPerDay),
	)

	// If the $$paddedYear is more than 9999 or less than 0, return NULL.
	yearRangeCheck := wrapInCond(
		nil,
		addDaysStatement,
		wrapInOp(mgoOperatorLt, paddedYear, 0),
		wrapInOp(mgoOperatorGt, paddedYear, 9999),
	)

	// Day range check, return NULL if day < 1.
	dayRangeCheck := wrapInCond(nil,
		yearRangeCheck,
		wrapInOp(mgoOperatorLt, day, 1),
	)

	outputLetStatement := bson.M{"output": dayRangeCheck}

	// Bind lets, and check that output value year < 9999, otherwise MySQL
	// returns NULL.
	return wrapInLet(inputLetStatement,
		wrapInLet(paddedYearLetStatement,
			wrapInLet(outputLetStatement,
				wrapInCond(nil, output,
					wrapInOp(mgoOperatorGt,
						wrapInOp(mgoOperatorYear, output),
						9999))),
		)), true

}

func (*makeDateFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*makeDateFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLInt, schema.SQLInt}
	defaults := []SQLValue{SQLNone, SQLNone}
	convertedExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*makeDateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLDate
}

func (*makeDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type md5Func struct{}

// https://dev.mysql.com/doc/refman/5.7/en/encryption-functions.html#function_md5
func (*md5Func) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	h := md5.New()
	_, err := io.WriteString(h, values[0].String())
	if err != nil {
		return nil, err
	}
	return SQLVarchar(fmt.Sprintf("%x", h.Sum(nil))), nil
}

func (*md5Func) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*md5Func) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*md5Func) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type microsecondFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_microsecond
func (*microsecondFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {

	arg := values[0]

	if arg == SQLNull {
		return SQLNull, nil
	}

	str := arg.String()
	if str == "" {
		return SQLNull, nil
	}

	t, _, ok := parseTime(str)
	if !ok {
		return SQLInt(0), nil
	}

	return SQLInt(t.Nanosecond() / 1000), nil
}

func (*microsecondFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapInNullCheckedCond(
		nil,
		bson.M{mgoOperatorMultiply: []interface{}{
			bson.M{"$millisecond": args[0]}, 1000,
		}},
		args[0],
	), true

}

func (*microsecondFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*microsecondFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*microsecondFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type minuteFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_minute
func (*minuteFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if values[0] == SQLNull {
		return SQLNull, nil
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
			return SQLNull, nil
		}
		return SQLInt(0), nil
	}

	return SQLInt(t.Minute()), nil
}

func (*minuteFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapSingleArgFuncWithNullCheck("$minute", args[0]), true
}

func (*minuteFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*minuteFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*minuteFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type modFunc struct {
	fun dualArgFloatMathFunc
}

func (f *modFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return f.fun.Evaluate(values, ctx)
}

func (*modFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 2 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return bson.M{"$mod": []interface{}{args[0], args[1]}}, true
}

func (f *modFunc) Normalize(e *SQLScalarFunctionExpr) SQLExpr {
	return f.fun.Normalize(e)
}

func (f *modFunc) Type(exprs []SQLExpr) schema.SQLType {
	return f.fun.Type(exprs)
}

func (f *modFunc) Validate(exprCount int) error {
	return f.fun.Validate(exprCount)
}

type monthFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_month
func (*monthFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Month())), nil
}

func (*monthFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapSingleArgFuncWithNullCheck("$month", args[0]), true
}

func (*monthFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (*monthFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*monthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type monthNameFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_monthname
func (*monthNameFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLVarchar(t.Month().String()), nil
}

func (*monthNameFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapInNullCheckedCond(
		nil,
		bson.M{mgoOperatorArrElemAt: []interface{}{
			[]interface{}{
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
			},
			bson.M{mgoOperatorSubtract: []interface{}{
				bson.M{"$month": args[0]},
				1}}}},
		args[0],
	), true
}

func (*monthNameFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (*monthNameFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*monthNameFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type multiArgFloatMathFunc struct {
	single singleArgFloatMathFunc
	dual   dualArgFloatMathFunc
}

func (f multiArgFloatMathFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	if len(values) == 1 {
		return f.single.Evaluate(values, ctx)
	}
	return f.dual.Evaluate(values, ctx)
}

func (multiArgFloatMathFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (multiArgFloatMathFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLFloat
}

func (multiArgFloatMathFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

type notFunc struct{}

func (*notFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	matcher := &SQLNotExpr{values[0]}
	result, err := Matches(matcher, ctx)
	if err != nil {
		return SQLNull, err
	}
	if NewSQLBool(result) == SQLTrue {
		return SQLInt(1), nil
	}
	return SQLInt(0), nil
}

func (*notFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*notFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type nullifFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/control-flow-functions.html
func (*nullifFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if _, ok := values[0].(SQLNullValue); ok {
		return SQLNull, nil
	} else if _, ok := values[1].(SQLNullValue); ok {
		return values[0], nil
	} else {
		eq, _ := CompareTo(values[0], values[1], ctx.Collation)
		if eq == 0 {
			return SQLNull, nil
		}
		return values[0], nil
	}
}

func (*nullifFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 2 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"expr": args[0],
	}

	letEvaluation := wrapInNullCheckedCond(
		nil,
		wrapInCond(
			nil,
			"$$expr",
			wrapInOp(mgoOperatorEq, "$$expr", args[1]),
		),
		"$$expr",
	)

	return wrapInLet(letAssignment, letEvaluation), true

}

func (*nullifFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if f.Exprs[0] == SQLNull {
		return SQLNull
	}

	if f.Exprs[1] == SQLNull {
		return f.Exprs[0]
	}

	return f
}

func (*nullifFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return f
}

func (*nullifFunc) Type(exprs []SQLExpr) schema.SQLType {
	return exprs[0].Type()
}

func (*nullifFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type padFunc struct {
	isLeftPad bool
	fun       scalarFunc
}

func (f *padFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return f.fun.Evaluate(values, ctx)
}

func (f *padFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {

	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}

	if len(exprs) != 3 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	// arguments to lpad
	str := args[0]
	lengthVal := args[1]
	padStr := args[2]

	// round to nearest int.
	length, ok := wrapInRound([]interface{}{lengthVal})
	if !ok {
		return nil, false
	}

	// variables for $let expression - length of padding needed
	// and length of input padding strings
	letAssignment := bson.M{
		"padLen": bson.M{
			mgoOperatorSubtract: []interface{}{
				length,
				bson.M{mgoOperatorStrlenCP: str}}},
		"padStrLen": bson.M{mgoOperatorStrlenCP: padStr},
		"length":    length,
	}

	// logic for generating padding string:

	// do we even need to add padding? only if the desired output
	// length is > length of input string.
	paddingCond := bson.M{
		mgoOperatorLt: []interface{}{
			bson.M{mgoOperatorStrlenCP: str},
			"$$length"}}

	// number of times we need to repeat the padding string to fill space
	padStrRepeats := bson.M{
		mgoOperatorCeil: bson.M{
			mgoOperatorDivide: []interface{}{"$$padLen", "$$padStrLen"}}}

	// generate an array with padStrRepeats occurrences of padStr
	padParts := bson.M{
		mgoOperatorMap: bson.M{
			"input": bson.M{
				mgoOperatorRange: []interface{}{
					0,
					padStrRepeats}},
			"in": padStr}}
	// join occurrences together and trim to the exact length needed
	fullPad := bson.M{
		mgoOperatorSubstr: []interface{}{
			bson.M{
				mgoOperatorReduce: bson.M{
					"input":        padParts,
					"initialValue": "",
					"in": bson.M{
						mgoOperatorConcat: []interface{}{"$$value", "$$this"}}}},
			0,
			"$$padLen"}}

	// based on length of input string, we either add the padding
	// or just take appropriate substring of input string
	var concatted bson.M
	if f.isLeftPad {
		concatted = bson.M{mgoOperatorConcat: []interface{}{fullPad, str}}
	} else {
		concatted = bson.M{mgoOperatorConcat: []interface{}{str, fullPad}}
	}

	handleConcat := wrapInCond(
		nil,
		concatted,
		bson.M{mgoOperatorEq: []interface{}{"$$padStrLen", 0}})

	// handle everything in the case that input length >=0
	handleNonNegativeLength := wrapInCond(
		handleConcat,
		bson.M{mgoOperatorSubstr: []interface{}{str, 0, "$$length"}},
		paddingCond)

	// whether the input length is < 0
	lengthIsNegative := bson.M{mgoOperatorLt: []interface{}{length, 0}}

	// if it's < 0, then we just want to return null
	negativeCheck := wrapInCond(nil, handleNonNegativeLength, lengthIsNegative)

	return wrapInNullCheckedCond(
			nil,
			wrapInLet(letAssignment, negativeCheck),
			str, lengthVal, padStr),
		true

}

func (f *padFunc) Reconcile(e *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	if ReconcileFunc, ok := f.fun.(reconcilingScalarFunc); ok {
		return ReconcileFunc.Reconcile(e)
	}
	panic("Unreachable, lpad and rpad are bout reconciling")
}

func (f *padFunc) Type(exprs []SQLExpr) schema.SQLType {
	return f.fun.Type(exprs)
}

func (f *padFunc) Validate(exprCount int) error {
	return f.fun.Validate(exprCount)
}

type powFunc struct{}

func (*powFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	v0 := values[0].Float64()
	v1 := values[1].Float64()

	n := math.Pow(v0, v1)
	zeroBaseExpNeg := v0 == 0 && v1 < 0
	if math.IsNaN(n) || zeroBaseExpNeg {
		return SQLNull,
			mysqlerrors.Defaultf(mysqlerrors.ErDataOutOfRange,
				"DOUBLE",
				fmt.Sprintf("pow(%v,%v)",
					values[0].Float64(),
					values[1].Float64()))
	}

	return SQLFloat(math.Pow(v0, v1)), nil
}

func (f *powFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 2 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapInOp(mgoOperatorPow, args[0], args[1]), true
}

func (*powFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*powFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLNumeric, SQLNone)
}

func (*powFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLFloat
}

func (*powFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type quarterFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_quarter
func (*quarterFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
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

	return SQLInt(q), nil
}

func (*quarterFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"date": args[0],
	}

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{mgoOperatorArrElemAt: []interface{}{
			[]interface{}{1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4},
			bson.M{mgoOperatorSubtract: []interface{}{
				bson.M{"$month": "$$date"},
				1}}}},
		"$$date",
	)

	return wrapInLet(letAssignment, letEvaluation), true

}

func (*quarterFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (*quarterFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*quarterFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type randFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_rand
func (*randFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	uniqueID := values[0].Uint64()
	if ctx.RandomExprs == nil {
		ctx.RandomExprs = make(map[uint64]*rand.Rand)
	}
	if r, ok := ctx.RandomExprs[uniqueID]; ok {
		return SQLFloat(r.Float64()), nil
	}
	if len(values) == 2 {
		r := rand.New(rand.NewSource(round(values[1].Float64())))
		ctx.RandomExprs[uniqueID] = r
		return SQLFloat(r.Float64()), nil
	}
	r := rand.New(rand.NewSource(rand.Int63()))
	ctx.RandomExprs[uniqueID] = r
	return SQLFloat(r.Float64()), nil
}

func (*randFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLNumeric, SQLNull)
}

func (*randFunc) RequiresEvalCtx() bool {
	return true
}

func (*randFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLFloat
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

func (*radiansFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapInOp(mgoOperatorDivide, wrapInOp(mgoOperatorMultiply, args[0], math.Pi), 180.0), true
}

type repeatFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_repeat
func (*repeatFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (v SQLValue, err error) {
	if hasNullValue(values...) {
		v = SQLNull
		err = nil
		return
	}

	str := values[0].String()
	if len(str) < 1 {
		v = SQLVarchar("")
		err = nil
		return
	}

	rep := int(roundToDecimalPlaces(0, values[1].Float64()))
	if rep < 1 {
		v = SQLVarchar("")
		err = nil
		return
	}

	var b bytes.Buffer
	for i := 0; i < rep; i++ {
		b.WriteString(str)
	}

	v = SQLVarchar(b.String())
	err = nil
	return
}

func (*repeatFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}

	if len(exprs) != 2 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	str := args[0]

	// num must be rounded to match mysql
	num, ok := wrapInRound(args[1:])
	if !ok {
		return nil, false
	}

	// create array w/ args[1] values e.g. [0,1,2]
	rangeArr := bson.M{mgoOperatorRange: []interface{}{0, num, 1}}

	// create array of len arg[1], with each item being arg[0]
	mapArgs := bson.M{"input": rangeArr, "in": str}
	mapWithArgs := bson.M{"$map": mapArgs}

	// append all values of this array together
	inArg := bson.M{mgoOperatorConcat: []interface{}{"$$this", "$$value"}}
	reduceArgs := bson.M{"input": mapWithArgs, "initialValue": "", "in": inArg}

	repeat := bson.M{mgoOperatorReduce: reduceArgs}

	return wrapInNullCheckedCond(nil, repeat, str, num), true

}

func (*repeatFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*repeatFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLNumeric}
	defaults := []SQLValue{SQLNone, SQLNone}
	convertedExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*repeatFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*repeatFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type replaceFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_replace
func (*replaceFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	s := values[0].String()
	old := values[1].String()
	new := values[2].String()

	return SQLVarchar(strings.Replace(s, old, new, -1)), nil
}

func (*replaceFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}

	if len(exprs) != 3 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	split := "$$split"
	assignment := bson.M{
		"split": wrapInOp(mgoOperatorSplit, args[0], args[1]),
	}

	this, value := "$$this", "$$value"
	body := wrapInReduce(split,
		nil,
		wrapInCond(this,
			wrapInOp(mgoOperatorConcat, value, args[2], this),
			wrapInOp(mgoOperatorEq, value, nil),
		),
	)

	return wrapInLet(assignment, body), true
}

func (*replaceFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*replaceFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar}
	defaults := []SQLValue{SQLNone, SQLNone, SQLNone}
	newExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		&replaceFunc{},
		newExprs,
	}
}

func (*replaceFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*replaceFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type reverseFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_reverse
func (*reverseFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}
	s := values[0].String()
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return SQLVarchar(string(runes)), nil
}

func (*reverseFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}

	if len(exprs) != 1 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}
	return wrapInCond(
			nil,
			wrapInLet(bson.M{"input": args[0]},
				wrapInReduce(
					bson.M{mgoOperatorRange: []interface{}{
						0,
						bson.M{mgoOperatorStrlenCP: "$$input"},
					}},
					"",
					bson.M{mgoOperatorConcat: []interface{}{
						bson.M{"$substrCP": []interface{}{
							"$$input",
							"$$this",
							1,
						}},
						"$$value",
					}}),
			),
			bson.M{mgoOperatorLte: []interface{}{args[0], nil}},
		),
		true
}

func (*reverseFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (rf *reverseFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar}
	defaults := []SQLValue{SQLNone}
	newExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		rf,
		newExprs,
	}
}

func (*reverseFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*reverseFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type rightFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_right
func (*rightFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	str := values[0].String()
	posFloat := values[1].Float64()

	if posFloat > float64(len(str)) {
		return SQLVarchar(str), nil
	}

	startPos := math.Min(0, -1.0*posFloat)

	substring,
		err := NewSQLScalarFunctionExpr("substring",
		[]SQLExpr{values[0],
			SQLFloat(startPos)})
	if err != nil {
		return SQLNull, err
	}

	return substring.Evaluate(ctx)
}

func (*rightFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {

	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}
	if len(exprs) != 2 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"string": args[0],
		"length": args[1],
	}

	// when length is negative, just use 0. round length to closest integer
	subStrLength, _ := wrapInRound([]interface{}{wrapInOp(mgoOperatorMax, "$$length", 0)})

	// start = max(0, strLen - subStrLen)
	start := wrapInOp(mgoOperatorMax,
		0,
		wrapInOp(mgoOperatorSubtract,
			bson.M{mgoOperatorStrlenCP: "$$string"},
			subStrLength))

	subStrOp := bson.M{mgoOperatorSubstr: []interface{}{
		"$$string",
		start,
		subStrLength}}

	letEvaluation := wrapInNullCheckedCond(nil, subStrOp, "$$string", "$$length")
	return wrapInLet(letAssignment, letEvaluation), true

}

func (*rightFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLInt}
	defaults := []SQLValue{SQLNone, SQLNone}
	convertedExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*rightFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*rightFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type roundFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_round
func (*roundFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	base := values[0].Float64()

	var decimal int64
	if len(values) == 2 {
		decimal = values[1].Int64()

		if decimal < 0 {
			return SQLFloat(0), nil
		}
	} else {
		decimal = 0
	}

	rounded := roundToDecimalPlaces(decimal, base)

	return SQLFloat(rounded), nil
}

func (*roundFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !(len(exprs) == 2 || len(exprs) == 1) {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapInRound(args)

}

func (*roundFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*roundFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLFloat, SQLNone)
}

func (*roundFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLFloat
}

func (*roundFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

type rpadFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_rpad
func (*rpadFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return handlePadding(values, false)
}

func (*rpadFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}
	return f
}

func (*rpadFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLInt, schema.SQLVarchar}
	defaults := []SQLValue{SQLNull, SQLNull, SQLNull}
	convertedExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*rpadFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*rpadFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type rtrimFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_rtrim
func (*rtrimFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	value := strings.TrimRight(values[0].String(), " ")

	return SQLVarchar(value), nil
}

func (*rtrimFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	if t.Ctx.VersionAtLeast(4, 0, 0) {
		return bson.M{
			mgoOperatorRTrim: bson.M{
				"input": args[0],
				"chars": " ",
			},
		}, true
	}

	rtrimCond := wrapInCond(
		"",
		wrapLRTrim(false, args[0]),
		bson.M{mgoOperatorEq: []interface{}{args[0], ""}})

	return wrapInNullCheckedCond(
		nil,
		rtrimCond,
		args[0],
	), true
}

func (*rtrimFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNull)
}

func (*rtrimFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*rtrimFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type secondFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_second
func (*secondFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if values[0] == SQLNull {
		return SQLNull, nil
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
			return SQLNull, nil
		}
		return SQLInt(0), nil
	}

	return SQLInt(t.Second()), nil
}

func (*secondFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapSingleArgFuncWithNullCheck("$second", args[0]), true
}

func (*secondFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*secondFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*secondFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type signFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_sign
func (*signFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	v := values[0].Float64()
	// Positive numbers are more common than negative in most data sets.
	if v > 0 {
		return SQLInt(1), nil
	}
	if v < 0 {
		return SQLInt(-1), nil
	}
	return SQLInt(0), nil
}

func (*signFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapInCond(nil,
		wrapInCond(wrapInLiteral(0),
			wrapInCond(wrapInLiteral(1),
				wrapInLiteral(-1),
				bson.M{mgoOperatorGt: []interface{}{args[0], wrapInLiteral(0)}},
			),
			bson.M{mgoOperatorEq: []interface{}{args[0], wrapInLiteral(0)}},
		),
		bson.M{mgoOperatorLte: []interface{}{args[0], nil}},
	), true

}

func (*signFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (sf *signFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLNumeric}
	defaults := []SQLValue{SQLNull}
	newExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		sf,
		newExprs,
	}
}

func (*signFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*signFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_sin
type sinFunc struct {
	singleArgFloatMathFunc
}

func (*sinFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	input, absInput := "$$input", "$$absInput"
	inputLetAssignment := bson.M{
		"input": args[0],
	}
	absInputLetAssignment := bson.M{
		"absInput": wrapInOp(mgoOperatorAbs, input),
	}
	rem, phase := "$$rem", "$$phase"
	remPhaseAssignment := bson.M{
		"rem": wrapInOp(mgoOperatorMod, absInput, math.Pi/2),
		"phase": wrapInOp(mgoOperatorMod,
			wrapInOp(mgoOperatorTrunc,
				wrapInOp(mgoOperatorDivide, absInput, math.Pi/2),
			),
			4.0),
	}

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
	threeCase := wrapInCond(wrapInOp(mgoOperatorMultiply,
		-1.0,
		wrapInCosPowerSeries(rem)),
		nil,
		wrapInOp(mgoOperatorEq,
			phase,
			3))
	twoCase := wrapInCond(wrapInOp(mgoOperatorMultiply,
		-1.0,
		wrapInSinPowerSeries(rem)),
		threeCase,
		wrapInOp(mgoOperatorEq,
			phase,
			2))
	oneCase := wrapInCond(wrapInCosPowerSeries(rem),
		twoCase,
		wrapInOp(mgoOperatorEq,
			phase,
			1))
	zeroCase := wrapInCond(wrapInSinPowerSeries(rem),
		oneCase,
		wrapInOp(mgoOperatorEq,
			phase,
			0))

	// cos(-x) = cos(x), but sin(-x) = -sin(x), so if the original input is negative multiply by -1.
	return wrapInLet(inputLetAssignment,
		wrapInLet(absInputLetAssignment,
			wrapInLet(remPhaseAssignment,
				wrapInCond(zeroCase,
					wrapInOp(mgoOperatorMultiply, -1.0, zeroCase),
					wrapInOp(mgoOperatorGte, input, 0),
				),
			),
		),
	), true
}

type singleArgFloatMathFunc func(float64) float64

func (f singleArgFloatMathFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	result := f(values[0].Float64())
	if math.IsNaN(result) {
		return SQLNull, nil
	}
	if math.IsInf(result, 0) {
		return SQLNull, nil
	}
	if result == -0 {
		result = 0
	}
	return SQLFloat(result), nil
}

func (singleArgFloatMathFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (singleArgFloatMathFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	switch f.Name {
	case "abs", "ceil", "exp", "degrees", "floor", "ln", "log", "log10", "log2", "radians", "sqrt":
		return convertAllArgs(f, schema.SQLFloat, SQLNone)
	default:
		return f
	}
}

func (singleArgFloatMathFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLFloat
}

func (singleArgFloatMathFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type sleepFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/miscellaneous-functions.html#function_sleep
func (*sleepFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {

	err := mysqlerrors.Defaultf(mysqlerrors.ErWrongArguments, "sleep")

	if hasNullValue(values...) {
		return nil, err
	}

	n := values[0].Float64()

	if n < 0 {
		return nil, err
	}

	timer := time.NewTimer(time.Second * time.Duration(n))

	select {
	case <-timer.C:
	case <-ctx.Context().Done():
		timer.Stop()
		return nil, ctx.Context().Err()
	}

	return SQLInt(0), nil

}

func (*sleepFunc) RequiresEvalCtx() bool {
	return true
}

func (*sleepFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*sleepFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type spaceFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_space
func (*spaceFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	f := values[0].Float64()
	n := round(f)
	if n < 1 {
		return SQLVarchar(""), nil
	}

	return SQLVarchar(strings.Repeat(" ", int(n))), nil
}

func (*spaceFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}

	if len(exprs) != 1 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	n := "$$n"
	return wrapInLet(bson.M{"n": wrapInRoundValue(args[0])},
		wrapInCond(nil,
			wrapInReduce(wrapInRange(0, n, 1),
				"",
				wrapInOp(mgoOperatorConcat, "$$value", " "),
			),
			wrapInOp(mgoOperatorLte, n, nil),
		),
	), true
}

func (*spaceFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLInt}
	defaults := []SQLValue{SQLNone}
	convertedExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*spaceFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*spaceFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type sqrtFunc struct {
	singleArgFloatMathFunc
}

func (*sqrtFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapInCond(
		bson.M{"$sqrt": args[0]},
		nil,
		bson.M{mgoOperatorGte: []interface{}{args[0], 0}},
	), true
}

type strToDateFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_str-to-date
func (*strToDateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	str, ok := values[0].(SQLVarchar)
	if !ok {
		return SQLNull, nil
	}

	ft, ok := values[1].(SQLVarchar)
	if !ok {
		return SQLNull, nil
	}

	s := str.String()
	f := ft.String()
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
	for idx, char := range f {
		if !skipToken {
			if char == 37 && idx != len(f)-1 {
				token := "%" + string(f[idx+1])
				skipToken = true
				goToken := fmtTokens[token]
				if goToken != "" {
					format += goToken
				} else {
					return SQLNull, nil
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
		return SQLNull, nil
	}

	if ts {
		return SQLTimestamp{d}, nil
	}

	return SQLDate{d}, nil
}

func (*strToDateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLTimestamp
}

func (*strToDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type subDateFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_subdate
func (*subDateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	subtractor := &dateSubFunc{}
	return subtractor.Evaluate(values, ctx)
}

func (*subDateFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*subDateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLTimestamp
}

func (*subDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2, 3)
}

type substringFunc struct {
	isMid bool
}

func (*substringFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	str := []rune(values[0].String())
	if values[0].String() == "" {
		return SQLVarchar(""), nil
	}

	posFloat := values[1].Float64()
	var pos int
	if posFloat >= 0 {
		pos = int(posFloat + 0.5)
	} else {
		pos = int(posFloat - 0.5)
	}

	if pos > len(str) || pos == 0 {
		return SQLVarchar(""), nil
	} else if pos < 0 {
		pos = len(str) + pos

		if pos < 0 {
			return SQLVarchar(""), nil
		}
	} else {
		pos-- // MySQL uses 1 as a basis
	}

	if len(values) == 3 {
		length := int(values[2].Float64() + 0.5)
		if length < 1 {
			return SQLVarchar(""), nil
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
	return SQLVarchar(string(str)), nil
}

func (f *substringFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}
	if (len(exprs) != 2 && len(exprs) != 3) ||
		(len(exprs) == 2 && f.isMid) {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}
	strVal := args[0]
	indexVal := args[1]

	var lenVal interface{}
	if len(args) == 3 {
		lenVal = args[2]
	} else {
		lenVal = bson.M{"$strLenCP": args[0]}
	}

	indexNegVal := wrapInLet(
		bson.M{
			"indexValNeg": bson.M{
				mgoOperatorTrunc: bson.M{
					mgoOperatorAdd: []interface{}{
						bson.M{mgoOperatorMultiply: []interface{}{indexVal, -1}}, 0.5}}}},
		wrapInCond(
			bson.M{mgoOperatorSubtract: []interface{}{bson.M{mgoOperatorStrlenCP: strVal},
				"$$indexValNeg"}},
			"$$indexValNeg",
			bson.M{mgoOperatorGte: []interface{}{bson.M{mgoOperatorStrlenCP: strVal},
				"$$indexValNeg"}}))

	indexPosVal := bson.M{
		mgoOperatorSubtract: []interface{}{
			bson.M{
				mgoOperatorTrunc: bson.M{
					mgoOperatorAdd: []interface{}{indexVal, 0.5}},
			}, 1}}

	roundOffIndex := wrapInCond(
		bson.M{mgoOperatorTrunc: bson.M{mgoOperatorAdd: []interface{}{indexVal, 0.5}}},
		bson.M{mgoOperatorTrunc: bson.M{mgoOperatorAdd: []interface{}{indexVal, -0.5}}},
		bson.M{mgoOperatorGte: []interface{}{indexVal, 0}})

	indexValBsonM := wrapInLet(
		bson.M{"roundOffIndex": roundOffIndex},
		wrapInCond(
			bson.M{mgoOperatorStrlenCP: strVal},
			wrapInCond(
				indexPosVal,
				indexNegVal,
				bson.M{mgoOperatorGt: []interface{}{"$$roundOffIndex", 0}}),
			bson.M{mgoOperatorEq: []interface{}{"$$roundOffIndex", 0}},
		))

	lenValBsonM := wrapInCond(
		0,
		bson.M{mgoOperatorTrunc: bson.M{mgoOperatorAdd: []interface{}{lenVal, 0.5}}},
		bson.M{mgoOperatorLte: []interface{}{lenVal, 0}},
	)

	return wrapInNullCheckedCond(
		nil,
		bson.M{"$substrCP": []interface{}{strVal, indexValBsonM, lenValBsonM}},
		strVal, indexVal, lenVal), true
}

func (*substringFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*substringFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLInt, schema.SQLInt}
	defaults := []SQLValue{SQLNone, SQLNone, SQLNone}
	convertedExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*substringFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*substringFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2, 3)
}

type substringIndexFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_substring-index
func (*substringIndexFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {

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
		return SQLNull, nil
	}

	r := []rune(values[0].String())
	delim := []rune(values[1].String())

	count := int(round(values[2].Float64()))

	if count == 0 {
		return SQLVarchar(""), nil
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

	return SQLVarchar(string(r)), nil
}

func (*substringIndexFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}

	if len(exprs) != 3 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	delim, split := "$$delim", "$$split"
	inputAssignment := bson.M{
		"delim": args[1],
	}

	splitAssignment := bson.M{
		"split": wrapInOp(mgoOperatorSlice,
			wrapInOp(mgoOperatorSplit, args[0], delim),
			wrapInRoundValue(args[2]),
		),
	}

	this, value := "$$this", "$$value"
	body := wrapInReduce(split,
		nil,
		wrapInCond(this,
			wrapInOp(mgoOperatorConcat, value, delim, this),
			wrapInOp(mgoOperatorEq, value, nil),
		),
	)

	return wrapInLet(inputAssignment,
		wrapInLet(splitAssignment,
			body,
		),
	), true
}

func (*substringIndexFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	if v, ok := f.Exprs[2].(SQLValue); ok {
		if v.Int64() == 0 {
			return SQLVarchar("")
		}
	}

	return f
}

func (sif *substringIndexFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLInt}
	defaults := []SQLValue{SQLNone, SQLNone, SQLNone}
	newExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		sif,
		newExprs,
	}
}

func (*substringIndexFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*substringIndexFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type tanFunc struct {
	singleArgFloatMathFunc
}

func (*tanFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	sf := &sinFunc{}
	num, pushedDown := sf.FuncToAggregationLanguage(t, []SQLExpr{exprs[0]})
	if !pushedDown {
		return nil, false
	}

	cf := &cosFunc{}
	denom, pushedDown := cf.FuncToAggregationLanguage(t, []SQLExpr{exprs[0]})
	if !pushedDown {
		return nil, false
	}

	// epsilon the smallest value we allow for denom, computed to roughly
	// tie-out with mysqld.
	epsilon := 6.123233995736766e-17
	// Replace abs(denom) < epsilon with epsilon since mysql doesn't seem
	// to return NULL or INF for inf values of tan as one would expect, and
	// we do not want to trigger a $divide by 0.
	return wrapInOp(mgoOperatorDivide,
		num,
		wrapInCond(epsilon,
			denom,
			wrapInOp(mgoOperatorLte,
				wrapInOp(mgoOperatorAbs, denom), epsilon,
			),
		),
	), true
}

type timeDiffFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_timediff
func (*timeDiffFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	expr1, _, ok := parseTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	expr2, _, ok := parseTime(values[1].String())
	if !ok {
		return SQLNull, nil
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
		return SQLVarchar("00:00:00.000000"), nil
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

	return SQLVarchar(string(buf[w:])), nil
}

func (*timeDiffFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*timeDiffFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*timeDiffFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type timeToSecFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_time-to-sec
func (*timeToSecFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
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
			return SQLNull, err
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
			return SQLNull, err
		}

		cmp := math.Trunc(component)

		switch i {
		// more on valid time types at https://dev.mysql.com/doc/refman/5.7/en/time.html
		case 0:
			if cmp > 838 || cmp < -838 {
				if !componentized {
					return SQLNull, nil
				}
				cmp = math.Copysign(838.0, cmp)
				components = []string{"", "59", "59"}
			}
		default:
			if cmp > 59 {
				return SQLNull, nil
			}
		}

		signBit = signBit || math.Signbit(cmp)
		result += math.Abs(cmp) * (3600.0 / (math.Pow(60, float64(i))))
	}

	if signBit {
		return SQLFloat(math.Copysign(result, -1)), nil
	}

	return SQLFloat(result), nil
}

func (*timeToSecFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*timeToSecFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLFloat
}

func (*timeToSecFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type timestampAddFunc struct{}

//http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_timestampadd
func (*timestampAddFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}
	v := int(round(values[1].Float64()))
	// values[2] must be a SQLTimestamp or the algebrizer has been broken.
	// Check the handling of scalar functions in the algebrizer.
	tmp, ok := values[2].(SQLTimestamp)
	if !ok {
		return nil,
			fmt.Errorf("unable to evaluate type %v in timestampadd,"+
				" this points to an error in the algebrizer",
				values[2])
	}
	t := tmp.Time

	ts := false
	if len(values[2].String()) > 10 {
		ts = true
	}

	switch values[0].String() {
	case Year:
		return SQLTimestamp{t.AddDate(v, 0, 0)}, nil
	case Quarter:
		y, mp, d := t.Date()
		m := int(mp)
		interval := v * 3
		y += (m + interval - 1) / 12
		m = (m+interval-1)%12 + 1
		switch m {
		case 2:
			if isLeapYear(y) {
				d = util.MinInt(d, 29)
			} else {

				d = util.MinInt(d, 28)
			}
		case 4, 6, 9, 11:
			d = util.MinInt(d, 30)
		}
		if ts {
			return SQLTimestamp{time.Date(y,
					time.Month(m),
					d,
					t.Hour(),
					t.Minute(),
					t.Second(),
					t.Nanosecond(),
					schema.DefaultLocale)},
				nil
		}
		return SQLTimestamp{time.Date(y,
				time.Month(m),
				d,
				0,
				0,
				0,
				0,
				schema.DefaultLocale)},
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
				d = util.MinInt(d, 29)
			} else {

				d = util.MinInt(d, 28)
			}
		case 4, 6, 9, 11:
			d = util.MinInt(d, 30)
		}
		if ts {
			return SQLTimestamp{time.Date(y,
					time.Month(m),
					d,
					t.Hour(),
					t.Minute(),
					t.Second(),
					t.Nanosecond(),
					schema.DefaultLocale)},
				nil
		}
		return SQLTimestamp{time.Date(y,
				time.Month(m),
				d,
				0,
				0,
				0,
				0,
				schema.DefaultLocale)},
			nil
	case Week:
		return SQLTimestamp{t.AddDate(0, 0, v*7)}, nil
	case Day:
		return SQLTimestamp{t.AddDate(0, 0, v)}, nil
	case Hour:
		duration := time.Duration(v) * time.Hour
		return SQLTimestamp{t.Add(duration)}, nil
	case Minute:
		duration := time.Duration(v) * time.Minute
		return SQLTimestamp{t.Add(duration)}, nil
	case Second:
		// Seconds can actually be fractional rather than integer.
		duration := time.Duration(int64(values[1].Float64() * 1e9))
		return SQLTimestamp{t.Add(duration)}, nil
	case Microsecond:
		duration := time.Duration(int64(values[1].Float64())) * time.Microsecond
		return SQLTimestamp{Time: t.Add(duration).Round(time.Millisecond)}, nil
	default:
		return SQLNull, fmt.Errorf("cannot add '%v' to timestamp", values[0])
	}
}

func (*timestampAddFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 5, 0) {
		return nil, false
	}

	if len(exprs) != 3 {
		return nil, false
	}

	unit := exprs[0].String()
	args, ok := t.translateArgs(exprs[1:])
	if !ok {
		return nil, false
	}
	interval := args[0]

	timestampExpr := args[1]
	// This is a very large and costly expression, make sure to bind it in
	// a let (in the switch at the end of the function).
	letAssignment := bson.M{
		"timestampArg": timestampExpr,
	}

	// Use timestampArg to refer to $$timestampArg below, referencing the var defined above.
	timestampArg := "$$timestampArg"

	// handleSimpleCase generates code for cases where we do not need to
	// use $dateFromParts, we just round the interval if the round argument
	// is true, and multiply by the number of milliseconds corresponded to
	// by 'u' then add to the timestamp.
	handleSimpleCase := func(u string, round bool) interface{} {
		if round {
			return wrapInOp(mgoOperatorAdd,
				timestampArg,
				wrapInOp(mgoOperatorMultiply,
					wrapInRoundValue(interval),
					toMilliseconds[u]))
		}
		return wrapInOp(mgoOperatorAdd,
			timestampArg,
			wrapInOp(mgoOperatorMultiply,
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
		dayExpr := wrapInOp(mgoOperatorDayOfMonth, timestampArg)
		// This template is used in a call to $dateFromParts.
		// The Year case modifies part of the template.
		template := bson.M{
			"year":  "$$newYear",
			"month": "$$newMonth",
			"day":
			// The following MongoDB aggregation language implements this go code,
			// the goal of which is to keep days from overflowing when adding
			// Quarters or Months.
			// switch m {
			// case 2:
			// 	if isLeapYear(y) {
			// 		d = util.MinInt(d, 29)
			//	} else {
			//		d = util.MinInt(d, 28)
			//	}
			// case 4, 6, 9, 11:
			//	d = util.MinInt(d, 30)
			// }
			// otherwise d is left unchanged as the day of the input timestamp.
			wrapInSwitch(wrapInOp(mgoOperatorDayOfMonth, timestampArg),
				wrapInEqCase(newMonth, 2,
					wrapInCond(wrapInOp(mgoOperatorMin, dayExpr, 29),
						wrapInOp(mgoOperatorMin, dayExpr, 28),
						wrapInIsLeapYear(newYear)),
				),
				wrapInEqCase(newMonth, 4,
					wrapInOp(mgoOperatorMin, dayExpr, 30)),
				wrapInEqCase(newMonth, 6,
					wrapInOp(mgoOperatorMin, dayExpr, 30)),
				wrapInEqCase(newMonth, 9,
					wrapInOp(mgoOperatorMin, dayExpr, 30)),
				wrapInEqCase(newMonth, 11,
					wrapInOp(mgoOperatorMin, dayExpr, 30)),
			),
			"hour":        wrapInOp(mgoOperatorHour, timestampArg),
			"minute":      wrapInOp(mgoOperatorMinute, timestampArg),
			"second":      wrapInOp(mgoOperatorSecond, timestampArg),
			"millisecond": wrapInOp(mgoOperatorMillisecond, timestampArg),
		}
		var sharedComputationLetAssignment interface{}
		var newYearMonthLetAssignment interface{}
		switch u {
		case Year:
			// For Year intervals, the year, month, and day use
			// different, simpler equations. Keep everything but
			// year, to year we add the rounded interval. There is
			// no SharedComputation part, so we do not wrapInLet.
			// Note that the rest of the template is maintained.
			template["year"] = wrapInOp(mgoOperatorAdd,
				wrapInRoundValue(interval),
				wrapInOp(mgoOperatorYear,
					timestampArg))
			template["month"] = wrapInOp(mgoOperatorMonth,
				timestampArg)
			template["day"] = wrapInOp(mgoOperatorDayOfMonth,
				timestampArg)
			return bson.M{mgoOperatorDateFromParts: template}
		// For Quarter and Month intervals, only the SharedComputation
		// part changes.
		case Quarter:
			// SharedComputation = Month + round(interval) * 3 - 1.
			sharedComputationLetAssignment = bson.M{
				"sharedComputation": wrapInOp(mgoOperatorSubtract,
					wrapInOp(mgoOperatorAdd,
						wrapInOp(mgoOperatorMonth, timestampArg),
						wrapInOp(mgoOperatorMultiply,
							wrapInRoundValue(interval),
							3),
					),
					1),
			}
		case Month:
			// SharedComputation = Month + round(interval) - 1.
			sharedComputationLetAssignment = bson.M{
				"sharedComputation": wrapInOp(mgoOperatorSubtract,
					wrapInOp(mgoOperatorAdd,
						wrapInOp(mgoOperatorMonth, timestampArg),
						wrapInRoundValue(interval),
					),
					1),
			}
		}

		newYearMonthLetAssignment = bson.M{
			// Year = Year + SharedComputation / 12, where / truncates.
			"newYear": wrapInOp(mgoOperatorAdd,
				wrapInOp(mgoOperatorYear, timestampArg),
				wrapInIntDiv(sharedComputation, 12),
			),
			// Month = SharedComputation % 12 + 1.
			"newMonth": wrapInOp(mgoOperatorAdd,
				wrapInOp(mgoOperatorMod,
					sharedComputation,
					12),
				1),
		}

		// Add lets for Quarter and Month.
		return wrapInLet(sharedComputationLetAssignment,
			wrapInLet(newYearMonthLetAssignment,
				bson.M{mgoOperatorDateFromParts: template},
			),
		)
	}

	// wrapInLet to bind $$timestampArg.
	switch unit {
	case Year, Month, Quarter:
		return wrapInLet(letAssignment, handleDateFromPartsCase(unit)), true
	// It is wrong to round for Second, and rounding for Microsecond is
	// just pointless since MongoDB supports only milliseconds, and will
	// automatically round to the nearest millisecond for us.
	case Second, Microsecond:
		return wrapInLet(letAssignment, handleSimpleCase(unit, false)), true
	default:
		return wrapInLet(letAssignment, handleSimpleCase(unit, true)), true
	}
}

func (*timestampAddFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLNone, schema.SQLInt, schema.SQLNone}
	defaults := []SQLValue{SQLNone, SQLNone, SQLNone}
	convertedExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (t *timestampAddFunc) Type(exprs []SQLExpr) schema.SQLType {
	// Checking the length of the argument to return conditional
	// types is not safe with pushdown. Timestamp add will
	// just always return a timestamp. There is no way to fix
	// this wrt Mongo DB's semantics.
	return schema.SQLTimestamp
}

func (t *timestampAddFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type timestampDiffFunc struct{}

//http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_timestampdiff
func (*timestampDiffFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
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
	t1 := tmp1.Time

	tmp2, ok := values[2].(SQLTimestamp)
	if !ok {
		return nil, fmt.Errorf("unable to evaluate type %v in timestampdiff,"+
			" this points to an error in the algebrizer",
			values[2])
	}
	t2 := tmp2.Time

	duration := t2.Sub(t1)

	switch values[0].String() {
	case Year:
		return SQLInt(float64(numMonths(t1, t2) / 12)), nil
	case Quarter:
		return SQLInt(float64(numMonths(t1, t2) / 3)), nil
	case Month:
		return SQLInt(numMonths(t1, t2)), nil
	case Week:
		if t1.After(t2) {
			return SQLInt(math.Ceil((duration.Hours()) / 24 / 7)), nil
		}
		return SQLInt(math.Floor((duration.Hours()) / 24 / 7)), nil
	case Day:
		if t1.After(t2) {
			return SQLInt(math.Ceil(duration.Hours() / 24)), nil
		}
		return SQLInt(math.Floor(duration.Hours() / 24)), nil
	case Hour:
		return SQLInt(duration.Hours()), nil
	case Minute:
		return SQLInt(duration.Minutes()), nil
	case Second:
		return SQLInt(duration.Seconds()), nil
	case Microsecond:
		return SQLInt(duration.Nanoseconds() / 1000), nil
	default:
		return SQLNull, fmt.Errorf("cannot add '%v' to timestamp", values[0])
	}
}

func (t *timestampDiffFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*timestampDiffFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 5, 0) {
		return nil, false
	}

	if len(exprs) != 3 {
		return nil, false
	}

	unit := exprs[0].String()

	args, ok := t.translateArgs(exprs[1:])
	if !ok {
		return nil, false
	}

	timestampExpr1, timestampExpr2 := args[0], args[1]

	// This is a very large and costly expression, make sure to bind it in
	// a let (in the switch at the end of the function).
	letAssignment := bson.M{
		"timestampArg1": timestampExpr1,
		"timestampArg2": timestampExpr2,
	}

	// Use timestampArg{1,2} to refer to $$timestampArg{1,2} below,
	// referencing the var defined above.
	timestampArg1, timestampArg2 := "$$timestampArg1", "$$timestampArg2"

	// handleSimpleCase generates code for cases where we do not need to
	// use and date part access functions (like $dayOfMonth), we just
	// subtract: timestampArg2 - timestampArg1 then divide by the number of
	// milliseconds corresponded to by 'u'.
	handleSimpleCase := func(u string) interface{} {
		return wrapInIntDiv(wrapInOp(mgoOperatorSubtract,
			timestampArg2,
			timestampArg1),
			toMilliseconds[u])
	}

	// handleDatePartsCase handles cases where we need to use
	// date part access functions (like $dayOfMonth).
	handleDatePartsCase := func(u string) interface{} {
		year1, month1 := "$$year1", "$$month1"
		year2, month2 := "$$year2", "$$month2"
		datePartsLetAssignment := bson.M{
			"year1":        wrapInOp(mgoOperatorYear, timestampArg1),
			"month1":       wrapInOp(mgoOperatorMonth, timestampArg1),
			"day1":         wrapInOp(mgoOperatorDayOfMonth, timestampArg1),
			"hour1":        wrapInOp(mgoOperatorHour, timestampArg1),
			"minute1":      wrapInOp(mgoOperatorMinute, timestampArg1),
			"second1":      wrapInOp(mgoOperatorSecond, timestampArg1),
			"millisecond1": wrapInOp(mgoOperatorMillisecond, timestampArg1),
			"year2":        wrapInOp(mgoOperatorYear, timestampArg2),
			"month2":       wrapInOp(mgoOperatorMonth, timestampArg2),
			"day2":         wrapInOp(mgoOperatorDayOfMonth, timestampArg2),
			"hour2":        wrapInOp(mgoOperatorHour, timestampArg2),
			"minute2":      wrapInOp(mgoOperatorMinute, timestampArg2),
			"second2":      wrapInOp(mgoOperatorSecond, timestampArg2),
			"millisecond2": wrapInOp(mgoOperatorMillisecond, timestampArg2),
		}

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
				return wrapInCond(wrapInLiteral(1),
					wrapInLiteral(0),
					wrapInOp(mgoOperatorGt, "$$month"+arg1, "$$month"+arg2),
					wrapInOp(mgoOperatorGt, "$$day"+arg1, "$$day"+arg2),
					wrapInOp(mgoOperatorGt, "$$hour"+arg1, "$$hour"+arg2),
					wrapInOp(mgoOperatorGt, "$$minute"+arg1, "$$minute"+arg2),
					wrapInOp(mgoOperatorGt, "$$second"+arg1, "$$second"+arg2),
					wrapInOp(mgoOperatorGt, "$$millisecond"+arg1, "$$millisecond"+arg2),
				)
			}
			// output = year2 - year1.
			outputLetAssignment = bson.M{
				"output": wrapInOp(mgoOperatorSubtract, year2, year1),
			}
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
				return wrapInCond(wrapInLiteral(1),
					wrapInLiteral(0),
					wrapInOp(mgoOperatorGt, "$$day"+arg1, "$$day"+arg2),
					wrapInOp(mgoOperatorGt, "$$hour"+arg1, "$$hour"+arg2),
					wrapInOp(mgoOperatorGt, "$$minute"+arg1, "$$minute"+arg2),
					wrapInOp(mgoOperatorGt, "$$second"+arg1, "$$second"+arg2),
					wrapInOp(mgoOperatorGt, "$$millisecond"+arg1, "$$millisecond"+arg2),
				)

			}
			// output = (year2 - year1) * 12 + month2 - month1.
			outputLetAssignment = bson.M{
				"output": wrapInOp(mgoOperatorAdd,
					wrapInOp(mgoOperatorMultiply,
						wrapInOp(mgoOperatorSubtract, year2, year1),
						12),
					wrapInOp(mgoOperatorSubtract, month2, month1),
				),
			}
		}

		// Generate epsilons and whether we add or subtract said epsilon, which
		// is decided on whether or not "output" is negative or positive.
		ltBranch := wrapInOp(mgoOperatorAdd, output, generateEpsilon("2", "1"))
		gtBranch := wrapInOp(mgoOperatorSubtract, output, generateEpsilon("1", "2"))
		applyEpsilonExpr := wrapInLet(outputLetAssignment,
			wrapInSwitch(wrapInLiteral(0),
				wrapInCase(wrapInOp(mgoOperatorLt, output, wrapInLiteral(0)), ltBranch),
				wrapInCase(wrapInOp(mgoOperatorGt, output, wrapInLiteral(0)), gtBranch),
			),
		)

		retExpr := wrapInLet(datePartsLetAssignment,
			wrapInLet(outputLetAssignment,
				applyEpsilonExpr,
			),
		)
		// Quarter is just the number of months integer divided by 3.
		if u == Quarter {
			return wrapInIntDiv(retExpr, 3)
		}
		return retExpr
	}

	// wrapInLet to bind $$timestampArg.
	switch unit {
	case Year, Month, Quarter:
		return wrapInLet(letAssignment, handleDatePartsCase(unit)), true
	default:
		return wrapInLet(letAssignment, handleSimpleCase(unit)), true
	}
}

func (t *timestampDiffFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type timestampFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_timestamp
func (*timestampFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	t = t.In(schema.DefaultLocale)

	if len(values) == 1 {
		return SQLTimestamp{t}, nil
	}

	d, ok := parseDuration(values[1])
	if !ok {
		return SQLNull, nil
	}

	t = t.Add(d).Round(time.Microsecond)

	return SQLTimestamp{t}, nil
}

func (tf *timestampFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 5, 0) {
		return nil, false
	}

	if len(exprs) != 1 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	val := "$$val"
	inputLet := bson.M{
		"val": args[0],
	}

	wrapInDateFromString := func(v interface{}) bson.M {
		return bson.M{mgoOperatorDateFromString: bson.M{"dateString": v}}
	}

	// CASE 1: it's already a Mongo date, we just return it
	isDateType := containsBSONType(val, "date")
	dateBranch := wrapInCase(isDateType, val)

	// CASE 2: it's a number.
	isNumber := containsBSONType(val, "int", "decimal", "long", "double")

	// evaluates to true if val positive and has <= X digits.
	hasUpToXDigits := func(x float64) interface{} {
		return wrapInInRange(val, 0, math.Pow(10, x))
	}

	// This handles converting a number in YYMMDDHHMMSS format to YYYYMMDDHHMMSS.
	// if YY < 70, we assume they meant 20YY. if YY > 70, we assume 19YY.
	getPadding := func(v interface{}) interface{} {
		return wrapInCond(
			20000000000000,
			19000000000000,
			wrapInOp(mgoOperatorLt,
				wrapInOp(mgoOperatorDivide,
					v, 10000000000),
				70))
	}

	// Constant for the HHMMSS factor to handle dates that do not have HHMMSS.
	hhmmssFactor := 1000000

	// We interpret this as being format YYMMDD, multiply by hhmmssFactor for HHMMSS then pad.
	ifSix := wrapInOp(mgoOperatorAdd,
		wrapInOp(mgoOperatorMultiply,
			val,
			hhmmssFactor),
		getPadding(wrapInOp(mgoOperatorMultiply,
			val,
			hhmmssFactor)))
	sixBranch := wrapInCase(hasUpToXDigits(6), ifSix)

	// This number is YYYYMMDD, again, multiply by hhmmssFactor.
	eightBranch := wrapInCase(hasUpToXDigits(8), wrapInOp(mgoOperatorMultiply, val, hhmmssFactor))

	// If it's twelve digits, interpret as YYMMDDHHMMSS. Make sure to pad the number.
	ifTwelve := wrapInOp(mgoOperatorAdd, val, getPadding(val))
	twelveBranch := wrapInCase(hasUpToXDigits(12), ifTwelve)

	// if fourteen, YYYYMMDDHHMMSS, we can use as it as is.
	fourteenBranch := wrapInCase(hasUpToXDigits(14), val)

	// define "num", the input number normalized to 14 digits, in a "let"
	numberVar := wrapInSwitch(nil, sixBranch, eightBranch, twelveBranch, fourteenBranch)
	numberLetVars := bson.M{"num": numberVar}

	dateParts := bson.M{
		// YYYYMMDDHHMMSS / 10000000000 = YYYY
		"year": bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide,
			"$$num",
			10000000000)},
		// (YYYYMMDDHHMMSS / 100000000) % 100 = MM
		"month": wrapInOp(mgoOperatorMod,
			bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide,
				"$$num",
				100000000)},
			100),
		// YYYYMMDDHHMMSS / 1000000) % 100 = DD
		"day": wrapInOp(mgoOperatorMod,
			bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide,
				"$$num",
				1000000)},
			100),
		// YYYYMMDDHHMMSS / 10000) % 100 = HH
		"hour": wrapInOp(mgoOperatorMod,
			bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide,
				"$$num",
				10000)},
			100),
		// YYYYMMDDHHMMSS / 100) % 100 = MM
		"minute": wrapInOp(mgoOperatorMod,
			bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide,
				"$$num",
				100)},
			100),
		// YYYYMMDDHHMMSS % 100 = SS
		"second": bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorMod,
			"$$num",
			100)},
		// YYYYMMDDHHMMSS.FFFFF % 1 * 1000 = ms
		"millisecond": bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorMultiply,
			wrapInOp(mgoOperatorMod,
				"$$num",
				1),
			1000)},
	}

	// try to avoid aggregation errors by catching obviously invalid dates
	yearValid := wrapInInRange("$$year", 0, 10000)
	monthValid := wrapInInRange("$$month", 1, 13)
	dayValid := wrapInInRange("$$day", 1, 32)
	// Mongo DB actually supports HH=24 which converts to 0, but MySQL does not (it returns NULL)
	// so we stick to MySQL semantics and cap valid hours at 23.
	// Interestingly, $dateFromString does NOT support HH=24.
	hourValid := wrapInInRange("$$hour", 0, 24)
	minuteValid := wrapInInRange("$$minute", 0, 60)
	secondValid := wrapInInRange("$$second", 0, 60)

	makeDateOrNull := wrapInCond(
		bson.M{mgoOperatorDateFromParts: bson.M{
			"year":        "$$year",
			"month":       "$$month",
			"day":         "$$day",
			"hour":        "$$hour",
			"minute":      "$$minute",
			"second":      "$$second",
			"millisecond": "$$millisecond",
		}},
		nil,
		bson.M{mgoOperatorAnd: []interface{}{yearValid,
			monthValid,
			dayValid,
			hourValid,
			minuteValid,
			secondValid}},
	)

	evaluateNumber := wrapInLet(dateParts, makeDateOrNull)
	handleNumberToDate := wrapInLet(numberLetVars, evaluateNumber)
	numberBranch := wrapInCase(isNumber, handleNumberToDate)

	// CASE 3: it's a string
	isString := containsBSONType(val, "string")

	// First split on T, take first substring, then split that on " ", and
	// take first substring. this gives us just the date part of the
	// string. note that if the string doesn't have T or a space, just
	// returns original string
	trimmedDateString := wrapInOp(mgoOperatorArrElemAt,
		wrapInOp(mgoOperatorSplit,
			wrapInOp(mgoOperatorArrElemAt,
				wrapInOp(mgoOperatorSplit, val, "T"),
				0),
			" "),
		0)

	// Repeat the step above but take the second element to get the time
	// part. Replace with "" if we can not find a second element.
	trimmedTimeString := wrapInIfNull(
		wrapInOp(mgoOperatorArrElemAt,
			wrapInOp(mgoOperatorSplit, val, "T"),
			1),
		wrapInIfNull(
			wrapInOp(mgoOperatorArrElemAt,
				wrapInOp(mgoOperatorSplit, val, " "),
				1),
			""),
	)

	// Convert the date and time strings to arrays so we can use
	// map/reduce.
	trimmedDateAsArray := wrapInStringToArray("$$trimmedDate")
	trimmedTimeAsArray := wrapInStringToArray("$$trimmedTime")

	// isSeparator evaluates to true if a character is in the defined
	// separator list
	isSeparator := wrapInOp(mgoOperatorNeq,
		-1,
		wrapInOp("$indexOfArray",
			dateComponentSeparator,
			"$$c"))

	// Use map to convert all separators in the date string to - symbol,
	// and leave numbers as-is
	dateNormalized := wrapInMap(trimmedDateAsArray,
		"c",
		wrapInCond("-",
			"$$c",
			isSeparator))
	// Use map to convert all separators in the time string to '.' symbol,
	// and leave numbers as-is. We use '.' instead of ':' so that MongoDB
	// correctly handles fractional seconds. 10.11.23.1234 is parsed
	// correctly as 10:11:23.1234, saving us some effort (and runtime).
	timeNormalized := wrapInMap(trimmedTimeAsArray,
		"c",
		wrapInCond(".",
			"$$c",
			isSeparator))

	// Use reduce to convert characters back to a single string for date and time.
	dateJoined := wrapInReduce(dateNormalized,
		"",
		wrapInOp(mgoOperatorConcat,
			"$$value",
			"$$this"))
	timeJoined := wrapInReduce(timeNormalized,
		"",
		wrapInOp(mgoOperatorConcat,
			"$$value",
			"$$this"))

	// if the third character is a -, or if the string is only 6 digits
	// long and has no slashes, then the string is either format YY/MM/DD
	// or YYMMDD and we need to add the appropriate first two year digits
	// (19xx or 20xx) for Mongo to understand it
	hasShortYear := wrapInOp(mgoOperatorOr,
		// length is only 6, assume YYMMDD
		wrapInOp(mgoOperatorEq, bson.M{mgoOperatorStrlenCP: "$$dateJoined"}, 6),
		// third character is -, assume YY-MM-DD
		wrapInOp(mgoOperatorEq,
			"-",
			bson.M{mgoOperatorSubstr: []interface{}{"$$dateJoined",
				2,
				1}}))

	// mgoOperatorDateFromString actually pads correctly, but not if "/" is
	// used as the separator (it will assume year is last). If this
	// pushdown is shown to be slow by benchmarks, we should reconsider
	// allowing mgoOperatorDateFromString to handle padding. The change
	// would not be trivial due to how MongoDB cannot handle short dates
	// when there are no separators in the date.
	padYear := wrapInOp(mgoOperatorConcat,
		wrapInCond(
			"20",
			"19",
			// check if first two digits < 70 to determine padding
			wrapInOp(
				mgoOperatorLt,
				bson.M{mgoOperatorSubstr: []interface{}{"$$dateJoined", 0, 2}},
				"70")),
		"$$dateJoined")

	// we have to use nested $lets because in the outer one we define
	// $$trimmedDate and in the inner one we define $$dateJoined. defining
	// $$dateJoined requires knowing the length of trimmedDate, so we can't
	// do it all in one step.
	innerIn := wrapInCond(padYear, "$$dateJoined", hasShortYear)
	innerLet := wrapInLet(bson.M{"dateJoined": dateJoined}, innerIn)

	// Concat the time back into the date.
	concatedDate := wrapInOp(mgoOperatorConcat,
		innerLet,
		timeJoined)

	// gracefully handle strings that are too short to possibly be valid by returning null
	tooShort := wrapInOp(mgoOperatorLt, bson.M{mgoOperatorStrlenCP: "$$trimmedDate"}, 6)
	outerIn := wrapInCond(nil, wrapInDateFromString(concatedDate), tooShort)
	outerLet := wrapInLet(bson.M{"trimmedDate": trimmedDateString,
		"trimmedTime": trimmedTimeString,
	}, outerIn)

	// Make sure if we get the int 0 we return NULL instead
	// of crashing. MySQL uses '0000-00-00' as an error output for some
	// functions and we encode it as the integer 0 within push down.
	stringBranch := wrapInCase(isString,
		wrapInCond(nil,
			outerLet,
			wrapInOp(mgoOperatorEq,
				0,
				args[0])))

	return wrapInLet(inputLet, wrapInSwitch(nil, dateBranch, numberBranch, stringBranch)), true

}

func (*timestampFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*timestampFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLTimestamp
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

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_to-days
func (*toDaysFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	// This must be a SQLDate at this point. If it is not, the algebrizer has been broken.
	// Check the handling of scalar functions in the algebrizer.
	tmp, ok := values[0].(SQLDate)
	if !ok {
		return nil, fmt.Errorf("unable to evaluate input value %v in to_days", values[0])
	}
	date := tmp.Time

	// First compute the days from YearOne.
	target := daysFromYearOneCalculation(date)

	return SQLInt(target), nil
}

func (*toDaysFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

// FuncToAggregation for TO_DAYS has one issue wrt how TO_DAYS is supposed to perform:
// because our date treatment is backed by using MongoDB's mgoOperatorDateFromString function,
// if a date that doesn't exist (e.g., 0000-00-00 or 0001-02-29) is entered, we return
// an error instead of the NULL expected from MySQL. Unfortunately, checking for valid
// dates is too cost prohibitive. If at some point $dateFromString supports an onError/default
// value, we should switch to using that.
func (*toDaysFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
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
	return bson.M{mgoOperatorTrunc: bson.M{mgoOperatorDivide: []interface{}{
		bson.M{mgoOperatorSubtract: []interface{}{args[0], dayOne}},
		MillisecondsPerDay,
	}}}, true
}

func (*toDaysFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt64
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

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_to-seconds
func (*toSecondsFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	// This must be a SQLTimestamp at this point. If it is not, the algebrizer has been broken.
	// Check the handling of scalar functions in the algebrizer.
	tmp, ok := values[0].(SQLTimestamp)
	if !ok {
		return nil, fmt.Errorf("unable to evaluate input value %v in to_seconds", values[0])
	}
	date := tmp.Time

	// First compute the days from YearOne and convert to seconds.
	target := daysFromYearOneCalculation(date) * SecondsPerDay

	// Now add remainder hours, minutes, and seconds.
	target += float64(date.Hour())*SecondsPerHour +
		float64(date.Minute())*SecondsPerMinute + float64(date.Second())

	// target is now seconds since dayOne.
	return SQLInt(target), nil
}

func (*toSecondsFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}
	// Subtract dayOne (0000-01-01) from the argument in mongo, then
	// convertms to seconds. When using $subtract on two dates in
	// MongoDB, the number of ms between the two dates is returned, and
	// the purpose of the TO_SECONDS function is to get the number of
	// seconds since 0000-01-01:
	dayOne := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	return wrapInOp(mgoOperatorMultiply,
		wrapInOp(mgoOperatorSubtract, args[0], dayOne),
		1e-3,
	), true
}

func (*toSecondsFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt64
}

func (*toSecondsFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type trimFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_trim
func (*trimFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
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

	return SQLVarchar(value), nil
}

func (*trimFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.Ctx.VersionAtLeast(3, 4, 0) {
		return nil, false
	}
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	if t.Ctx.VersionAtLeast(4, 0, 0) {
		return bson.M{
			mgoOperatorTrim: bson.M{
				"input": args[0],
				"chars": " ",
			},
		}, true
	}

	rtrimCond := wrapInCond(
		"",
		wrapLRTrim(false, args[0]),
		bson.M{mgoOperatorEq: []interface{}{args[0], ""}})

	ltrimCond := wrapInCond(
		"",
		wrapLRTrim(true, "$$rtrim"),
		bson.M{mgoOperatorEq: []interface{}{"$$rtrim", ""}})

	trimCond := wrapInLet(bson.M{"rtrim": rtrimCond}, ltrimCond)

	trim := wrapInCond(
		"",
		trimCond,
		bson.M{mgoOperatorEq: []interface{}{args[0], ""}})

	return wrapInNullCheckedCond(
		nil,
		trim,
		args[0],
	), true
}

func (*trimFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNull)
}

func (*trimFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*trimFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 3)
}

type truncateFunc struct{}

//http://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_truncate
func (*truncateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	var truncated float64
	x := values[0].Float64()
	d := values[1].Float64()

	if d >= 0 {
		pow := math.Pow(10, d)
		i, _ := math.Modf(x * pow)
		truncated = i / pow
	} else {
		pow := math.Pow(10, math.Abs(d))
		i, _ := math.Modf(x / pow)
		truncated = i * pow
	}

	return SQLFloat(truncated), nil
}

func (*truncateFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 2 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	bsonMap, ok := args[1].(bson.M)
	if !ok {
		return nil, false
	}

	bsonVal, ok := bsonMap["$literal"]
	if !ok {
		return nil, false
	}

	dVal, _ := NewSQLValue(bsonVal, schema.SQLFloat, schema.SQLNone)

	d := dVal.Float64()
	if d >= 0 {
		pow := math.Pow(10, d)
		return bson.M{mgoOperatorDivide: []interface{}{
			bson.M{mgoOperatorCond: []interface{}{
				bson.M{mgoOperatorGte: []interface{}{args[0], 0}},
				bson.M{mgoOperatorFloor: bson.M{mgoOperatorMultiply: []interface{}{
					args[0], pow}}},
				bson.M{mgoOperatorCeil: bson.M{mgoOperatorMultiply: []interface{}{
					args[0], pow}}}}},
			pow}}, true
	}

	pow := math.Pow(10, math.Abs(d))
	return bson.M{mgoOperatorMultiply: []interface{}{
		bson.M{mgoOperatorCond: []interface{}{
			bson.M{mgoOperatorGte: []interface{}{args[0], 0}},
			bson.M{mgoOperatorFloor: bson.M{mgoOperatorDivide: []interface{}{
				args[0], pow}}},
			bson.M{mgoOperatorCeil: bson.M{mgoOperatorDivide: []interface{}{
				args[0], pow}}}}},
		pow}}, true
}

func (*truncateFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*truncateFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLFloat, schema.SQLNone}
	defaults := []SQLValue{SQLNone, SQLNone}
	convertedExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*truncateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLFloat
}

func (*truncateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type ucaseFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_ucase
func (*ucaseFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	value := strings.ToUpper(values[0].String())

	return SQLVarchar(value), nil
}

func (*ucaseFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapSingleArgFuncWithNullCheck("$toUpper", args[0]), true
}

func (*ucaseFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNone)
}

func (*ucaseFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*ucaseFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type unixTimestampFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_unix-timestamp
func (*unixTimestampFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	now := time.Now()

	if len(values) == 0 {
		return SQLUint64(now.Unix()), nil
	}

	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	t, _, ok := parseDateTime(values[0].String())
	if !ok || t.Before(epoch) {
		return SQLFloat(0.0), nil
	}

	// Our times are parsed as if in UTC. However, we need to
	// parse it in the actual location the server's running
	// in - to account for any timezone difference.
	y, m, d := t.Date()
	ts := time.Date(y, m, d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), now.Location())
	return SQLUint64(ts.Unix()), nil
}

func (*unixTimestampFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	now := time.Now()

	if len(exprs) != 1 {
		return wrapInLiteral(now.Unix()), true
	}

	arg, ok := (&timestampFunc{}).FuncToAggregationLanguage(t, exprs)
	if !ok {
		return nil, false
	}

	// Subtract epoch (1970-01-01) from the argument in MongoDB, then
	// convert ms to seconds. When using $subtract on two dates in
	// MongoDB, the number of milliseconds between the two
	// timestamps is returned.
	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	_, tzCompensation := now.Zone()

	letAssignment := bson.M{
		"diff": bson.M{
			mgoOperatorTrunc: bson.M{
				mgoOperatorDivide: []interface{}{
					bson.M{
						mgoOperatorSubtract: []interface{}{
							bson.M{
								mgoOperatorSubtract: []interface{}{arg, epoch},
							},
							tzCompensation * 1000,
						},
					},
					1000,
				},
			},
		},
	}

	letEvaluation := wrapInCond("$$diff", 0.0, wrapInOp(mgoOperatorGt, "$$diff", 0))
	return wrapInLet(letAssignment, letEvaluation), true
}

func (*unixTimestampFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*unixTimestampFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLUint64
}

func (*unixTimestampFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0, 1)
}

type userFunc struct{}

func (*userFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLVarchar(ctx.ExecutionCtx.User() + "@" + ctx.ExecutionCtx.RemoteHost()), nil
}

func (*userFunc) RequiresEvalCtx() bool {
	return true
}

func (*userFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*userFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type utcDateFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_utc-date
func (*utcDateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	now := time.Now().In(time.UTC)
	t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return SQLDate{t}, nil
}

func (*utcDateFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	now := time.Now().In(time.UTC)
	cUTCd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return wrapInLiteral(cUTCd), true
}

func (*utcDateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLDate
}

func (*utcDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type utcTimestampFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_utc-timestamp
func (*utcTimestampFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLTimestamp{time.Now().In(time.UTC)}, nil
}

func (*utcTimestampFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	return wrapInLiteral(time.Now().In(time.UTC)), true
}

func (*utcTimestampFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLTimestamp
}

func (*utcTimestampFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type versionFunc struct{}

func (*versionFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLVarchar(ctx.Variables().GetString(variable.Version)), nil
}

func (*versionFunc) RequiresEvalCtx() bool {
	return true
}

func (*versionFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*versionFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type weekFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_week
func (*weekFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}
	check, ok := values[0].(SQLDate)
	if !ok {
		return SQLNull, nil
	}

	dateArg := check.Time
	// Mode should always be less than MAX_INT.
	mode := int(values[1].Int64())

	ret := weekCalculation(dateArg, mode)
	if ret == -1 {
		return SQLNull, nil
	}
	return SQLInt(ret), nil
}

func (wf *weekFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) < 1 || len(exprs) > 2 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	mode := int64(0)
	if len(args) == 2 {
		bsonMap, ok := args[1].(bson.M)
		if !ok {
			return nil, false
		}

		bsonVal, ok := bsonMap["$literal"]
		if !ok {
			return nil, false
		}

		argOneVal, _ := NewSQLValue(bsonVal, schema.SQLInt, schema.SQLNone)
		mode = argOneVal.Int64()
	}

	return wrapInWeekCalculation(args[0], mode), true
}

func (*weekFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLNone, schema.SQLInt}
	defaults := []SQLValue{SQLNull, SQLNone}
	convertedExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		convertedExprs,
	}
}

func (*weekFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*weekFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

type weekdayFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_weekday
func (*weekdayFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	w := int(t.Weekday())
	if w == 0 {
		w = 7
	}
	return SQLInt(w - 1), nil
}

func (*weekdayFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	letAssignment := bson.M{
		"date": args[0],
	}

	letEvaluation := wrapInNullCheckedCond(
		nil,
		bson.M{mgoOperatorMod: []interface{}{
			bson.M{mgoOperatorAdd: []interface{}{
				bson.M{mgoOperatorMod: []interface{}{
					bson.M{mgoOperatorSubtract: []interface{}{
						bson.M{"$dayOfWeek": "$$date"}, 2,
					}}, 7,
				}}, 7,
			}}, 7,
		}},
		"$$date",
	)

	return wrapInLet(letAssignment, letEvaluation), true

}

func (*weekdayFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*weekdayFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type yearFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_year
func (*yearFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(t.Year()), nil
}

func (*yearFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapSingleArgFuncWithNullCheck("$year", args[0]), true
}

func (*yearFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (*yearFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*yearFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type yearWeekFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_yearweek
func (*yearWeekFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}
	check, ok := values[0].(SQLDate)
	if !ok {
		return SQLNull, nil
	}

	date := check.Time
	year := date.Year()
	// Mode should always be less than MAX_INT.
	mode := int(values[1].Int64())

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
		return SQLNull, nil
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
	return SQLInt(year*100 + week), nil
}

func (wf *yearWeekFunc) FuncToAggregationLanguage(
	t *PushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) < 1 || len(exprs) > 2 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	mode := int64(0)
	if len(args) == 2 {
		bsonMap, ok := args[1].(bson.M)
		if !ok {
			return nil, false
		}

		bsonVal, ok := bsonMap["$literal"]
		if !ok {
			return nil, false
		}

		argOneVal, _ := NewSQLValue(bsonVal, schema.SQLInt, schema.SQLNone)
		mode = argOneVal.Int64()
	}

	date, month, year, week := "$$date", "$$month", "$$year", "$$week"
	inputAssignment := bson.M{
		"date": args[0],
	}
	monthAssignment := bson.M{
		"month": wrapInOp(mgoOperatorMonth, date),
		"year":  wrapInOp(mgoOperatorYear, date),
	}

	var weekCalc interface{}

	// Unlike WEEK, YEARWEEK always uses the 1-53 modes. Thus
	// we always call week with the 1-53 of a 0-53, 1-53 pair.
	switch mode {

	// First day of week: Sunday, with a Sunday in this year.
	case 0, 2:
		weekCalc = wrapInWeekCalculation(date, 2)
	// First day of weekCalc: Monday, with 4 days in this year.
	case 1, 3:
		weekCalc = wrapInWeekCalculation(date, 3)
	// First day of weekCalc: Sunday, with 4 days in this year.
	case 4, 6:
		weekCalc = wrapInWeekCalculation(date, 6)
	// First day of weekCalc: Monday, with a Monday in this year.
	case 5, 7:
		weekCalc = wrapInWeekCalculation(date, 7)
	}

	weekAssignment := bson.M{
		"week": weekCalc,
	}

	newYear := "$$newYear"
	newYearAssignment := bson.M{
		"newYear": wrapInSwitch(year,
			wrapInEqCase(week, 1, wrapInCond(
				wrapInOp(mgoOperatorAdd, year, 1), year,
				wrapInOp(mgoOperatorEq, month, 12),
			),
			),
			wrapInEqCase(week, 52, wrapInCond(
				wrapInOp(mgoOperatorSubtract, year, 1), year,
				wrapInOp(mgoOperatorEq, month, 1),
			),
			),
			wrapInEqCase(week, 53, wrapInCond(
				wrapInOp(mgoOperatorSubtract, year, 1), year,
				wrapInOp(mgoOperatorEq, month, 1),
			),
			),
		),
	}

	return wrapInLet(inputAssignment,
		wrapInLet(monthAssignment,
			wrapInLet(weekAssignment,
				wrapInLet(newYearAssignment,
					wrapInOp(mgoOperatorAdd,
						wrapInOp(mgoOperatorMultiply, newYear, 100),
						week,
					),
				),
			),
		),
	), true

}

func (*yearWeekFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*yearWeekFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

func (
	t *PushDownTranslator) translateArgs(exprs []SQLExpr) ([]interface{}, bool) {
	args := []interface{}{}
	for _, e := range exprs {
		r, ok := t.ToAggregationLanguage(e)
		if !ok {
			return nil, false
		}
		args = append(args, r)
	}
	return args, true
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

func convertAllArgs(f *SQLScalarFunctionExpr,
	convType schema.SQLType,
	defaultValue SQLValue) *SQLScalarFunctionExpr {
	nExprs := convertAllExprs(f.Exprs, convType, defaultValue)
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
func calculateInterval(unit string,
	args []int,
	neg int) (string,
	int,
	error) {
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

// convertType converts val to the SQL type indicated by t.
func convertType(val SQLValue, t schema.SQLType) SQLValue {
	switch t {
	case schema.SQLInt:
		return SQLInt(val.Int64())
	case schema.SQLFloat:
		return SQLFloat(val.Float64())
	case schema.SQLVarchar:
		return SQLVarchar(val.String())
	case schema.SQLDate:
		t, _, _ := parseDateTime(val.String())
		return SQLDate{Time: t}
	case schema.SQLTimestamp:
		t, _, _ := parseDateTime(val.String())
		return SQLTimestamp{Time: t}
	case schema.SQLBoolean:
		if val.Float64() == 0 {
			return SQLFalse
		}
		return SQLTrue
	case schema.SQLDecimal128:
		return SQLDecimal128(val.Decimal128())
	}
	return SQLInt(0)
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

func evaluateArgs(exprs []SQLExpr, ctx *EvalCtx) ([]SQLValue, error) {

	values := []SQLValue{}

	for _, expr := range exprs {
		value, err := expr.Evaluate(ctx)
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
func handlePadding(values []SQLValue, isLeftPad bool) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	var length int
	// length should be converted to float before we get to here
	if floatLength := values[1].Float64(); floatLength < float64(0) {
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
		return SQLNull, nil
	}

	// the string is already long enough
	if len(str) >= length {
		return SQLVarchar(str[:length]), nil
	}

	// repeat padding as many times as needed to fill room
	numRepeats := math.Ceil(float64(padLen) / float64(len(padStr)))

	padding := []rune(strings.Repeat(string(padStr), int(numRepeats)))

	// in case room % len(padstr) != 0, chop off end
	padding = padding[:padLen]

	finalPad := string(padding)
	finalStr := string(str)

	if isLeftPad {
		return SQLVarchar(finalPad + finalStr), nil
	}

	return SQLVarchar(finalStr + finalPad), nil
}

// daysFromYearOneCalculation calculates the number of days since year one in a given unit
// (days or seconds). The argument inSeconds is set to true for second output.
func daysFromYearOneCalculation(date time.Time) float64 {
	// 0 - out any time parts of the date.
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, schema.DefaultLocale)
	targetInc := MaxGoDurationHours / 24.0
	target := 1.0
	start := time.Date(0, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
	for date.Sub(start).Hours() > 24 {
		for date.Sub(start).Hours() > MaxGoDurationHours {
			date = date.Add(time.Duration(-MaxGoDurationHours) * time.Hour)
			target += targetInc
		}
		// Subtract a day from date, add a day's worth of seconds to target
		date, target = date.AddDate(0, 0, -1), target+1.0
	}
	return target
}

// formatDate takes a time.Time object and outputs a string formatted using
// MySQL's format string specification.
func formatDate(date time.Time, format string, ctx *EvalCtx) (string, error) {
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

	weekFmt := func(i int) (string, error) {
		wf := &weekFunc{}
		args := []SQLValue{SQLDate{date}, SQLInt(i)}
		eval, err := wf.Evaluate(args, ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%02v", eval.String()), nil
	}

	yearFmt := func(i int) (string, error) {
		yw := &yearWeekFunc{}
		args := []SQLValue{SQLDate{date}, SQLInt(i)}
		eval, err := yw.Evaluate(args, ctx)
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
	// trunc((date - dayOne) / (7 * MillisecondsPerDay) + 1).
	computeDaySubtract := func(date, dayOne time.Time) int {
		return int(float64(date.Sub(dayOne))/
			(7.0*float64(MillisecondsPerDay)*float64(time.Millisecond)) +
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
