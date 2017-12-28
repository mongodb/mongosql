package evaluator

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"math"
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
	"github.com/shopspring/decimal"
)

const (
	shortTimeFormat = "2006-01-02"
)

const (
	Year               = "year"
	Quarter            = "quarter"
	Month              = "month"
	Week               = "week"
	Day                = "day"
	Hour               = "hour"
	Minute             = "minute"
	Second             = "second"
	Microsecond        = "microsecond"
	YearMonth          = "year_month"
	DayHour            = "day_hour"
	DayMinute          = "day_minute"
	DaySecond          = "day_second"
	DayMicrosecond     = "day_microsecond"
	HourMinute         = "hour_minute"
	HourSecond         = "hour_second"
	HourMicrosecond    = "hour_microsecond"
	MinuteSecond       = "minute_second"
	MinuteMicrosecond  = "minute_microsecond"
	SecondMicrosecond  = "second_microsecond"
	MillisecondsPerDay = 8.64e+7
	SecondsPerDay      = MillisecondsPerDay / 1e3
)

var (
	zeroDate, _ = time.ParseInLocation(shortTimeFormat, "0000-00-00", schema.DefaultLocale)
)

var scalarFuncMap = map[string]scalarFunc{
	"abs":     &absFunc{singleArgFloatMathFunc(math.Abs)},
	"acos":    singleArgFloatMathFunc(math.Acos),
	"adddate": &dateArithmeticFunc{&addDateFunc{}, false},
	"asin":    singleArgFloatMathFunc(math.Asin),
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
	"cos":               singleArgFloatMathFunc(math.Cos),
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
	"degrees":           singleArgFloatMathFunc(func(f float64) float64 { return f * 180 / math.Pi }),
	"elt":               &eltFunc{},
	"exp":               &expFunc{singleArgFloatMathFunc(math.Exp)},
	"extract":           &extractFunc{},
	"floor":             &floorFunc{singleArgFloatMathFunc(math.Floor)},
	"from_days":         &fromDaysFunc{},
	"greatest":          &greatestFunc{},
	"hour":              &hourFunc{},
	"if":                &ifFunc{},
	"ifnull":            &ifnullFunc{},
	"insert":            &insertFunc{},
	"instr":             &instrFunc{},
	"interval":          &intervalFunc{},
	"isnull":            &isnullFunc{},
	"last_day":          &lastDayFunc{},
	"lcase":             &lcaseFunc{},
	"least":             &leastFunc{},
	"left":              &leftFunc{},
	"length":            &lengthFunc{},
	"ln":                singleArgFloatMathFunc(math.Log),
	"locate":            &locateFunc{},
	// Use 0 for ln and logs where base is passed as first arg
	"log":             &logFunc{0},
	"log2":            &logFunc{2},
	"log10":           &logFunc{10},
	"lower":           &lcaseFunc{},
	"lpad":            &padFunc{true, &lpadFunc{}},
	"ltrim":           &ltrimFunc{},
	"makedate":        &makeDateFunc{},
	"md5":             &md5Func{},
	"microsecond":     &microsecondFunc{},
	"mid":             &substringFunc{true},
	"minute":          &minuteFunc{},
	"mod":             &modFunc{dualArgFloatMathFunc(math.Mod)},
	"month":           &monthFunc{},
	"monthname":       &monthNameFunc{},
	"not":             &notFunc{},
	"now":             &currentTimestampFunc{},
	"nullif":          &nullifFunc{},
	"pi":              &constantFunc{SQLFloat(math.Pi)},
	"pow":             &powFunc{},
	"power":           &powFunc{},
	"quarter":         &quarterFunc{},
	"radians":         singleArgFloatMathFunc(func(f float64) float64 { return f * math.Pi / 180 }),
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
	"sin":             singleArgFloatMathFunc(math.Sin),
	"sleep":           &sleepFunc{},
	"sqrt":            &sqrtFunc{singleArgFloatMathFunc(math.Sqrt)},
	"space":           &spaceFunc{},
	"str_to_date":     &strToDateFunc{},
	"subdate":         &dateArithmeticFunc{&subDateFunc{}, true},
	"substr":          &substringFunc{},
	"substring":       &substringFunc{},
	"substring_index": &substringIndexFunc{},
	"system_user":     &userFunc{},
	"tan":             singleArgFloatMathFunc(math.Tan),
	"timediff":        &timeDiffFunc{},
	"timestamp":       &timestampFunc{},
	"timestampadd":    &timestampAddFunc{},
	"timestampdiff":   &timestampDiffFunc{},
	"time_to_sec":     &timeToSecFunc{},
	"to_days":         &toDaysFunc{},
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
	"weekofyear":      &weekOfYearFunc{},
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
	FuncToAggregationLanguage(*pushDownTranslator, []SQLExpr) (interface{}, bool)
}

//
// SQLScalarFunctionExpr represents a scalar function.
//
type SQLScalarFunctionExpr struct {
	Name  string
	Func  scalarFunc
	Exprs []SQLExpr
}

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

func (f *SQLScalarFunctionExpr) Normalize() node {
	if nsf, ok := f.Func.(normalizingScalarFunc); ok {
		return nsf.Normalize(f)
	}

	return f
}

func (f *SQLScalarFunctionExpr) Reconcile() *SQLScalarFunctionExpr {
	if rsf, ok := f.Func.(reconcilingScalarFunc); ok {
		return rsf.Reconcile(f)
	}

	return f
}

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

func (f *SQLScalarFunctionExpr) ToAggregationLanguage(t *pushDownTranslator) (interface{}, bool) {
	if fun, ok := f.Func.(translatableToAggregationScalarFunc); ok {
		return fun.FuncToAggregationLanguage(t, f.Exprs)
	}
	t.logger.Debugf(log.Dev, "%q cannot be pushed down as an aggregate expression at this time", f.Name)
	return nil, false
}

func (f *SQLScalarFunctionExpr) Type() schema.SQLType {
	return f.Func.Type(f.Exprs)
}

type absFunc struct {
	singleArgFloatMathFunc
}

func (*absFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return bson.M{"$abs": args[0]}, true
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

type ceilFunc struct {
	singleArgFloatMathFunc
}

func (*ceilFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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
		return ErrIncorrectCount
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

func (*characterLengthFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.versionAtLeast(3, 4, 0) {
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

func (*coalesceFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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
func (*concatFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (v SQLValue, err error) {
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

func (*concatFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (*concatWsFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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
		return ErrIncorrectCount
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

type convertFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/cast-functions.html#function_convert
func (*convertFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	_, ok := values[0].(SQLNullValue)
	if ok {
		return SQLNull, nil
	}

	switch values[1].String() {
	case string(parser.SIGNED_BYTES):
		var i int64
		switch typedV := values[0].(type) {
		case SQLDate:
			i, _ = strconv.ParseInt(strings.Replace(typedV.String(), "-", "", -1), 10, 64)
		case SQLTimestamp:
			stripped := strings.Replace(strings.Replace(strings.Replace(typedV.String(), "-", "", -1), ":", "", -1), " ", "", -1)
			i, _ = strconv.ParseInt(stripped, 10, 64)
		case SQLFloat:
			i = int64(roundToDecimalPlaces(0, typedV.Float64()))
		case SQLInt:
			i = int64(typedV)
		case SQLVarchar:
			f, _ := strconv.ParseFloat(typedV.String(), 64)
			i = int64(f)
		case SQLBool:
			if typedV.Bool() {
				i = 1
			} else {
				i = 0
			}
		case SQLDecimal128:
			i = decimal.Decimal(typedV).Round(0).IntPart()
		default:
			return SQLNull, nil
		}

		return SQLInt(i), nil

	case string(parser.UNSIGNED_BYTES):
		var u uint64
		switch typedV := values[0].(type) {
		case SQLDate:
			i, _ := strconv.ParseInt(strings.Replace(typedV.String(), "-", "", -1), 10, 64)
			u = uint64(i)
		case SQLTimestamp:
			stripped := strings.Replace(strings.Replace(strings.Replace(typedV.String(), "-", "", -1), ":", "", -1), " ", "", -1)
			i, _ := strconv.ParseInt(stripped, 10, 64)
			u = uint64(i)
		case SQLFloat:
			u = uint64(roundToDecimalPlaces(0, typedV.Float64()))
		case SQLInt:
			u = uint64(typedV)
		case SQLVarchar:
			f, _ := strconv.ParseFloat(typedV.String(), 64)
			u = uint64(f)
		case SQLBool:
			if typedV.Bool() {
				u = 1
			} else {
				u = 0
			}
		case SQLDecimal128:
			u = uint64(decimal.Decimal(typedV).Round(0).IntPart())
		default:
			return SQLNull, nil
		}

		return SQLUint64(u), nil

	case string(parser.FLOAT_BYTES):
		var f float64
		switch typedV := values[0].(type) {
		case SQLDate:
			f, _ = strconv.ParseFloat(strings.Replace(typedV.String(), "-", "", -1), 64)
		case SQLTimestamp:
			stripped := strings.Replace(strings.Replace(strings.Replace(typedV.String(), "-", "", -1), ":", "", -1), " ", "", -1)
			f, _ = strconv.ParseFloat(stripped, 64)
		case SQLFloat:
			f = float64(typedV)
		case SQLInt:
			f = float64(typedV)
		case SQLVarchar:
			f, _ = strconv.ParseFloat(typedV.String(), 64)
		case SQLBool:
			if typedV.Bool() {
				f = float64(1)
			} else {
				f = float64(0)
			}
		case SQLDecimal128:
			f = typedV.Float64()
		default:
			return SQLNull, nil
		}

		return SQLFloat(f), nil

	case string(parser.CHAR_BYTES):
		var s string
		switch typedV := values[0].(type) {
		case SQLDate, SQLTimestamp:
			s = typedV.String()
		case SQLFloat:
			s = strconv.FormatFloat(typedV.Float64(), 'f', -1, 64)
		case SQLInt:
			s = strconv.FormatInt(int64(typedV), 10)
		case SQLVarchar:
			s = typedV.String()
		case SQLBool:
			if typedV.Bool() {
				s = "1"
			} else {
				s = "0"
			}
		case SQLDecimal128:
			s = util.FormatDecimal(decimal.Decimal(typedV))
		default:
			return SQLNull, nil
		}

		return SQLVarchar(s), nil

	case string(parser.DATE_BYTES):
		var t time.Time
		switch typedV := values[0].(type) {
		case SQLDate:
			t = typedV.Time
		case SQLTimestamp:
			t = typedV.Time
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		default:
			var ok bool
			t, _, ok = parseDateTime(typedV.String())
			if !ok {
				switch tv := values[0].(type) {
				case SQLBool, SQLUint32, SQLUint64, SQLInt, SQLFloat, SQLDecimal128:
					if tv.Int64() != 0 {
						return SQLNull, nil
					}
					t = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
				default:
					return SQLNull, nil
				}

				t = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
			} else {
				t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
			}
		}

		return SQLDate{Time: t}, nil

	case string(parser.DATETIME_BYTES):
		var t time.Time
		switch typedV := values[0].(type) {
		case SQLDate:
			t = typedV.Time
		case SQLTimestamp:
			t = typedV.Time
		default:
			var ok bool
			t, _, ok = parseDateTime(typedV.String())
			if !ok {
				switch tv := values[0].(type) {
				case SQLBool, SQLUint32, SQLUint64, SQLInt, SQLFloat, SQLDecimal128:
					if tv.Int64() != 0 {
						return SQLNull, nil
					}
					t = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
				default:
					return SQLNull, nil
				}
			}
		}

		return SQLTimestamp{Time: t}, nil

	case string(parser.DECIMAL_BYTES):
		var d decimal.Decimal
		switch typedV := values[0].(type) {
		case SQLDate, SQLTimestamp:
			i := typedV.Int64()
			d = decimal.New(i, 0)
		default:
			var err error
			d, err = decimal.NewFromString(typedV.String())
			if err != nil {
				return SQLNull, nil
			}
		}

		return SQLDecimal128(d), nil

	case string(parser.TIME_BYTES):
		var t time.Time
		switch typedV := values[0].(type) {
		case SQLDate:
			t = typedV.Time
			t = time.Date(0, 1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
		case SQLTimestamp:
			t = typedV.Time
			t = time.Date(0, 1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
		default:
			d, _, ok := strToTime(typedV.String())
			if !ok {
				if _, ok := typedV.(SQLArithmetic); !ok || typedV.Int64() != 0 {
					return SQLNull, nil
				}

				t = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
			} else {
				t = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Add(d)
			}
		}

		return SQLTimestamp{t}, nil

	default:
		return SQLNull, nil
	}
}

func (*convertFunc) Type(exprs []SQLExpr) schema.SQLType {
	if v, ok := exprs[1].(SQLValue); ok {
		switch v.String() {
		case string(parser.SIGNED_BYTES):
			return schema.SQLInt
		case string(parser.UNSIGNED_BYTES):
			return schema.SQLUint64
		case string(parser.FLOAT_BYTES):
			return schema.SQLFloat
		case string(parser.CHAR_BYTES):
			return schema.SQLVarchar
		case string(parser.DATE_BYTES):
			return schema.SQLDate
		case string(parser.DATETIME_BYTES):
			return schema.SQLTimestamp
		case string(parser.DECIMAL_BYTES):
			return schema.SQLDecimal128
		}
	}
	return schema.SQLNone
}

func (*convertFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type cotFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_cot
func (*cotFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	tan := math.Tan(values[0].Float64())
	if tan == 0 {
		return SQLNull, mysqlerrors.Defaultf(mysqlerrors.ER_DATA_OUT_OF_RANGE, "DOUBLE", fmt.Sprintf("'cot(%v)'", values[0].Float64()))
	}

	return SQLFloat(1 / tan), nil
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

func (*currentDateFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (*currentTimestampFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (f *dateArithmeticFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 3 {
		return nil, false
	}

	var date interface{}
	var ok bool
	if _, ok = f.scalarFunc.(*addDateFunc); ok {
		// implementation for ADDDATE(DATE_FORMAT("..."), INTERVAL 0 SECOND)
		if fun, ok := exprs[0].(*SQLScalarFunctionExpr); ok && fun.Name == "date_format" {
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

func (*dateDiffFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 2 {
		return nil, false
	}

	var date1, date2 interface{}
	var ok bool

	parseArgs := func(expr SQLExpr) (interface{}, bool) {
		if value, ok := expr.(SQLValue); ok {

			date, _, ok := strToDateTime(value.String(), false)
			if !ok {
				return nil, false
			}

			date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, schema.DefaultLocale)
			return date, true
		}
		exprType := expr.Type()
		if exprType == schema.SQLTimestamp || exprType == schema.SQLDate {
			date, ok := t.ToAggregationLanguage(expr)
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

	days := wrapInOp(mgoOperatorDivide, wrapInOp(mgoOperatorSubtract, date1, date2), 86400000)
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
	if !ok {
		return SQLNull, nil
	}

	v1, ok := values[1].(SQLVarchar)
	if !ok {
		return SQLNull, nil
	}

	date = date.In(schema.DefaultLocale)
	format := []rune(v1.String())

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
	for i := 0; i < len(format); i++ {
		if format[i] == '%' && i != len(format)-1 {
			if formatter, ok := formatters[format[i+1]]; ok {
				s, err := formatter()
				if err != nil {
					return SQLNull, err
				}
				result += s
				i++
			} else {
				result += string(format[i])
			}
		} else {
			result += string(format[i])
		}
	}

	return SQLVarchar(result), nil
}

func (*dateFormatFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

	mysqlFormat := formatValue.String()
	var format string
	for i := 0; i < len(mysqlFormat); i++ {
		if mysqlFormat[i] == '%' {
			if i != len(mysqlFormat)-1 {
				switch mysqlFormat[i+1] {
				case '%':
					format += "%%"
				case 'd':
					format += "%d"
				case 'f':
					format += "%L000"
				case 'H', 'k':
					format += "%H"
				case 'i':
					format += "%M"
				case 'j':
					format += "%j"
				case 'm':
					format += "%m"
				case 's', 'S':
					format += "%S"
				case 'T':
					format += "%H:%M:%S"
				case 'U':
					format += "%U"
				case 'Y':
					format += "%Y"
				default:
					return nil, false
				}
				i++
			} else {
				// MongoDB fails when the last character is a % sign in the format string.
				return nil, false
			}
		} else {
			format += string(mysqlFormat[i])
		}
	}

	return wrapInNullCheckedCond(
		nil,
		bson.M{"$dateToString": bson.M{
			"format": format,
			"date":   date,
		}},
		date,
	), true
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
	// too-short numbers are padded differently than too-short strings.
	// strToDateTime (called by parseDateTime) handles padding in the too-short
	// string case. we need to fix the string here, where we can still find out
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

func (*dateFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.versionAtLeast(3, 5, 0) {
		return nil, false
	}

	if len(exprs) != 1 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	val := args[0]

	wrapInDateFromString := func(v interface{}) bson.M {
		return bson.M{"$dateFromString": bson.M{"dateString": v}}
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

	// this handles converting a number in YYMMDD format to YYYYMMDD.
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

	// we interpret this as being format YYMMDD
	ifSix := wrapInOp(mgoOperatorAdd, val, getPadding(val))
	sixBranch := wrapInCase(hasUpToXDigits(6), ifSix)

	// this number is good as is! YYYYMMDD
	eightBranch := wrapInCase(hasUpToXDigits(8), val)

	// if it's twelve digits, interpret as YYMMDDHHMMSS.
	// first drop the last six digits, then pad like we would a six digit number.
	firstSixDigits := bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide, val, 1000000)}
	ifTwelve := wrapInOp(mgoOperatorAdd, firstSixDigits, getPadding(firstSixDigits))
	twelveBranch := wrapInCase(hasUpToXDigits(12), ifTwelve)

	// if fourteen, YYYYMMDDHHMMSS. just drop the last six digits.
	ifFourteen := bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide, val, 1000000)}
	fourteenBranch := wrapInCase(hasUpToXDigits(14), ifFourteen)

	// define "num", the input number normalized to 8 digits, in a "let"
	numberVar := wrapInSwitch(nil, sixBranch, eightBranch, twelveBranch, fourteenBranch)
	numberLetVars := bson.M{"num": numberVar}

	dateParts := bson.M{
		// YYYYMMDD / 10000 = YYYY
		"year": bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide, "$$num", 10000)},
		// (YYYYMMDD / 100) % 100 = MM
		"month": wrapInOp(mgoOperatorMod, bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide, "$$num", 100)}, 100),
		// YYYYMMDD % 100 = DD
		"day": wrapInOp(mgoOperatorMod, "$$num", 100),
	}

	// try to avoid aggregation errors by catching obviously invalid dates
	yearValid := wrapInInRange("$$year", 0, 10000)
	monthValid := wrapInInRange("$$month", 1, 13)
	dayValid := wrapInInRange("$$day", 1, 32)

	makeDateOrNull := wrapInCond(
		bson.M{"$dateFromParts": bson.M{
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

	// first split on T, take first substring, then split that on " ", and take first
	// substring. this gives us just the date part of the string. note that if the
	// string doesn't have T or a space, just returns original string
	trimmedString := wrapInOp(mgoOperatorArrElemAt,
		wrapInOp(mgoOperatorSplit,
			wrapInOp(mgoOperatorArrElemAt,
				wrapInOp(mgoOperatorSplit, val, "T"),
				0),
			" "),
		0)

	// convert the string to an array so we can use map/reduce
	trimmedAsArray := wrapInStringToArray("$$trimmed")

	// isSeparator evaluates to true if a character is in the defined separator list
	isSeparator := wrapInOp(mgoOperatorNeq, -1, wrapInOp("$indexOfArray", dateComponentSeparator, "$$c"))

	// use map to convert all separators in the string to - symbol, and leave numbers as-is
	separatorsNormalized := wrapInMap(trimmedAsArray, "c", wrapInCond("-", "$$c", isSeparator))

	// use reduce to convert characters back to a single string
	joined := wrapInReduce(separatorsNormalized, "", wrapInOp(mgoOperatorConcat, "$$value", "$$this"))

	// if the third character is a -, or if the string is only 6 digits long and has no slashes,
	// then the string is either format YY-MM-DD or YYMMDD and we need to add the appropriate first
	// two year digits (19xx or 20xx) for Mongo to understand it
	hasShortYear := wrapInOp(mgoOperatorOr,
		// length is only 6, assume YYMMDD
		wrapInOp(mgoOperatorEq, bson.M{mgoOperatorStrlenCP: "$$joined"}, 6),
		// third character is -, assume YY-MM-DD
		wrapInOp(mgoOperatorEq, "-", bson.M{mgoOperatorSubstr: []interface{}{"$$joined", 2, 1}}))

	// $dateFromString actually pads correctly, but not if "/" is used as the separator (it will assume year is last).
	// If this pushdown is shown to be slow by benchmarks, we should reconsider allowing $dateFromString to handle padding.
	// The change would not be trivial due to how MongoDB cannot handle short dates when there are no separators in the date.
	padYear := wrapInOp(mgoOperatorConcat,
		wrapInCond(
			"20",
			"19",
			// check if first two digits < 70 to determine padding
			wrapInOp(
				mgoOperatorLt,
				bson.M{mgoOperatorSubstr: []interface{}{"$$joined", 0, 2}},
				"70")),
		"$$joined")

	// we have to use nested $lets because in the outer one we define $$trimmed and
	// in the inner one we define $$joined. defining $$joined requires knowing the
	// length of trimmed, so we can't do it all in one step.
	innerIn := wrapInCond(padYear, "$$joined", hasShortYear)
	innerLet := wrapInLet(bson.M{"joined": joined}, innerIn)

	// gracefully handle strings that are too short to possibly be valid by returning null
	tooShort := wrapInOp(mgoOperatorLt, bson.M{mgoOperatorStrlenCP: "$$trimmed"}, 6)
	outerIn := wrapInCond(nil, wrapInDateFromString(innerLet), tooShort)
	outerLet := wrapInLet(bson.M{"trimmed": trimmedString}, outerIn)

	stringBranch := wrapInCase(isString, outerLet)

	return wrapInSwitch(nil, dateBranch, numberBranch, stringBranch), true

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

func (*dayNameFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

	return SQLInt(int(t.Day())), nil
}

func (*dayOfMonthFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (*dayOfWeekFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

	return SQLInt(int(t.YearDay())), nil
}

func (*dayOfYearFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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
		return ErrIncorrectCount
	}

	return nil
}

type expFunc struct {
	singleArgFloatMathFunc
}

func (f *expFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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
	// For certain units, we need to concatenate the unit values as strings before returning the int value
	// as to not lose any number's place value.
	// i.e. SELECT EXTRACT(DayMinute FROM "2006-04-03 06:03:23") should return 30603, not 363.
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

func (*extractFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

type floorFunc struct {
	singleArgFloatMathFunc
}

func (*floorFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

	if value <= 365.5 || value >= 3652499.5 || value <= 0 {
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

func (*fromDaysFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*fromDaysFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLDate
}

func (*fromDaysFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
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

func (*greatestFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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
		return ErrIncorrectVarCount
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

func (*hourFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (*ifFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (*ifnullFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

type insertFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_insert
func (*insertFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {

	if hasNullValue(values...) {
		return SQLNull, nil
	}

	s := values[0].String()
	pos := int(values[1].Int64()) - 1
	length := int(values[2].Int64())
	newstr := values[3].String()

	if pos < 0 || pos > len(s) {
		return values[0], nil
	}

	if pos+length < 0 || pos+length > len(s) {
		return SQLVarchar(s[:pos] + newstr), nil
	}

	return SQLVarchar(s[:pos] + newstr + s[pos+length:]), nil
}

func (*insertFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
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

func (*intervalFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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
		return ErrIncorrectVarCount
	}
	return nil
}

type isnullFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_isnull
func (*isnullFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	_, ok := values[0].(SQLNullValue)
	matcher := NewSQLBool(ok)

	result, err := Matches(matcher, ctx)
	if err != nil {
		return SQLNull, err
	}

	if NewSQLBool(result) == SQLTrue {
		return SQLInt(1), nil
	}

	return SQLInt(0), nil
}

func (*isnullFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	return wrapInNullCheckedCond(1, 0, args[0]), true
}

func (*isnullFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return f
}

func (*isnullFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*isnullFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type lastDayFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_last-day
func (*lastDayFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

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

func (*lcaseFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (*leastFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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
		return ErrIncorrectVarCount
	}
	return nil
}

type leftFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_left
func (*leftFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	substring, err := NewSQLScalarFunctionExpr("substring", []SQLExpr{values[0], SQLInt(1), values[1]})
	if err != nil {
		return SQLNull, err
	}
	return substring.Evaluate(ctx)
}

func (*leftFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {

	if !t.versionAtLeast(3, 4, 0) {
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

func (*lengthFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.versionAtLeast(3, 4, 0) {
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
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	substr := []rune(values[0].String())
	str := []rune(values[1].String())
	result := 0
	if len(values) == 3 {

		pos := int(values[2].Float64()) - 1 // MySQL uses 1 as a basis

		if len(str) <= pos {
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

func (*locateFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, false
	}
	if !(len(exprs) == 2 || len(exprs) == 3) {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	var indexOfCPArgs []interface{}
	if len(args) == 2 {
		indexOfCPArgs = []interface{}{args[1], args[0]}
	} else {
		indexOfCPArgs = []interface{}{args[1], args[0], wrapInOp(mgoOperatorSubtract, args[2], 1)}
	}

	return wrapInNullCheckedCond(
		nil,
		wrapInOp(mgoOperatorAdd, bson.M{"$indexOfCP": indexOfCPArgs}, 1),
		args[1], args[0],
	), true
}

func (*locateFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
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

func (f *logFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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
				bson.M{"$ln": args[0]},
				mgoNullLiteral}}, true
		}
		// Two args is based arg.
		// MySQL specifies base then arg, MongoDB expects arg then base, so we have to flip.
		return bson.M{mgoOperatorCond: []interface{}{
			bson.M{mgoOperatorGt: []interface{}{args[0], 0}},
			bson.M{"$log": []interface{}{args[1], args[0]}},
			mgoNullLiteral}}, true
	}
	// This will be base 10 or base 2 based on if log10 or log2 was called.
	return bson.M{mgoOperatorCond: []interface{}{
		bson.M{mgoOperatorGt: []interface{}{args[0], 0}},
		bson.M{"$log": []interface{}{args[0], f.Base}},
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

func (*ltrimFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, false
	}
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}
	ltrimCond := wrapInCond(
		"",
		wrapLRTrim(true, args[0]),
		bson.M{mgoOperatorEq: []interface{}{args[0], ""}})

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

	y := values[0].Int64()
	if y < 0 || y > 9999 {
		return SQLNull, nil
	}
	if y >= 0 && y <= 69 {
		y += 2000
	} else if y >= 70 && y <= 99 {
		y += 1900
	}

	d := values[1].Int64()
	if d <= 0 {
		return SQLNull, nil
	}

	t := time.Date(int(y), 1, 0, 0, 0, 0, 0, schema.DefaultLocale)
	duration := time.Duration(d*24) * time.Hour

	return SQLDate{Time: t.Add(duration)}, nil
}

func (*makeDateFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
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
	io.WriteString(h, values[0].String())
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

	return SQLInt(int(t.Nanosecond() / 1000)), nil
}

func (*microsecondFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

	return SQLInt(int(t.Minute())), nil
}

func (*minuteFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (*modFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (*monthFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (*monthNameFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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
	return schema.SQLBoolean
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

func (*nullifFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (f *padFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {

	if !t.versionAtLeast(3, 4, 0) {
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
	vars := bson.M{
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
			bson.M{
				mgoOperatorLet: bson.M{
					"vars": vars,
					"in":   negativeCheck}},
			str),
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
		return SQLNull, mysqlerrors.Defaultf(mysqlerrors.ER_DATA_OUT_OF_RANGE, "DOUBLE", fmt.Sprintf("pow(%v,%v)", values[0].Float64(), values[1].Float64()))
	}

	return SQLFloat(math.Pow(v0, v1)), nil
}

func (*powFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (*powFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLFloat, SQLNone)
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

func (*quarterFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (*repeatFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.versionAtLeast(3, 4, 0) {
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

func (*reverseFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.versionAtLeast(3, 4, 0) {
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

	substring, err := NewSQLScalarFunctionExpr("substring", []SQLExpr{values[0], SQLFloat(startPos)})
	if err != nil {
		return SQLNull, err
	}

	return substring.Evaluate(ctx)
}

func (*rightFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {

	if !t.versionAtLeast(3, 4, 0) {
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

func (*roundFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (*rtrimFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, false
	}
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
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

	return SQLInt(int(t.Second())), nil
}

func (*secondFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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
	// Positive numbers are more common than negative in most data sets
	if v > 0 {
		return SQLInt(1), nil
	}
	if v < 0 {
		return SQLInt(-1), nil
	}
	return SQLInt(0), nil
}

func (*signFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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
	case "abs", "ceil", "exp", "floor", "ln", "log", "log10", "log2", "sqrt":
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

	err := mysqlerrors.Defaultf(mysqlerrors.ER_WRONG_ARGUMENTS, "sleep")

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

	n := values[0].Int64()
	if n < 1 {
		return SQLVarchar(""), nil
	}

	return SQLVarchar(strings.Repeat(" ", int(n))), nil
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

func (*sqrtFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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
		"%a": "Mon", "%b": "Jan", "%c": "1", "%d": "02", "%e": "2", "%H": "15", "%i": "04", "%k": "13", "%M": "January",
		"%m": "01", "%S": "05", "%s": "05", "%T": "15:04:05", "%W": "Monday", "%w": "Mon", "%Y": "2006", "%y": "06",
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
	pos := 0
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

func (f *substringFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.versionAtLeast(3, 4, 0) {
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
			bson.M{mgoOperatorSubtract: []interface{}{bson.M{mgoOperatorStrlenCP: strVal}, "$$indexValNeg"}},
			"$$indexValNeg",
			bson.M{mgoOperatorGte: []interface{}{bson.M{mgoOperatorStrlenCP: strVal}, "$$indexValNeg"}}))

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
	count := int(values[2].Int64())

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

func (*substringIndexFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (*substringIndexFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
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

		component := strconv.FormatFloat(math.Trunc(float64(cmp)), 'f', -1, 64)

		l := len(component)
		components, componentized = []string{"0", "0", "0"}, false

		// MySQL interprets abbreviated values without colons using the
		// assumption that the two rightmost digits represent seconds.
		switch l {
		case 1, 2:
			components[2] = component
		case 3, 4:
			components[1], components[2] = component[:l-2], component[l-2:l]
		case 5:
			components[0], components[1], components[2] = component[:l-4], component[l-4:l-2], component[l-2:l]
		default:
			components[0], components[1], components[2] = component[:l-4], component[l-4:l-2], component[l-2:l]
		}
	}

	signBit := false

	for i := 0; i < 3 && i < len(components); i++ {
		component, err := strconv.ParseFloat(components[i], 64)
		if err != nil {
			return SQLNull, err
		}

		cmp := math.Trunc(float64(component))

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
	t, _, ok := parseDateTime(values[2].String())
	if !ok {
		return SQLNull, nil
	}

	v := values[1]

	ts := false
	if len(values[2].String()) > 10 {
		ts = true
	}

	switch values[0].String() {
	case Year:
		if ts {
			return SQLTimestamp{Time: t.AddDate(int(v.Int64()), 0, 0)}, nil
		}
		return SQLDate{t.AddDate(int(v.Int64()), 0, 0)}, nil
	case Quarter:
		y, m, d := t.Date()
		mo := int(((int64(m)+v.Int64()*3)%12 + 12) % 12)
		if mo == 0 {
			mo = 12
		}
		if v.Int64()*3 >= 12 || (v.Int64()*3) <= -12 {
			y += int(v.Int64() * 3 / 12)
		}
		if mo-int(m) < 0 && (v.Int64()*3) > 0 {
			y++
		} else if mo-int(m) > 0 && (v.Int64()*3) < 0 {
			y--
		}
		lastDayMonth := 32 - (time.Date(y, time.Month(mo), 32, 0, 0, 0, 0, schema.DefaultLocale)).Day()
		if d > lastDayMonth {
			d = lastDayMonth
		}

		if ts {
			return SQLTimestamp{time.Date(y, time.Month(mo), d, t.Hour(), t.Minute(), t.Second(), 0, schema.DefaultLocale)}, nil
		}
		return SQLDate{time.Date(y, time.Month(mo), d, 0, 0, 0, 0, schema.DefaultLocale)}, nil
	case Month:
		y, m, d := t.Date()
		mo := int(((int64(m)+v.Int64())%12 + 12) % 12)
		if mo == 0 {
			mo = 12
		}
		if v.Int64() >= 12 || v.Int64() <= -12 {
			y += int(v.Int64() / 12)
		}
		if mo-int(m) < 0 && v.Int64() > 0 {
			y++
		} else if mo-int(m) > 0 && v.Int64() < 0 {
			y--
		}
		lastDayMonth := 32 - (time.Date(y, time.Month(mo), 32, 0, 0, 0, 0, schema.DefaultLocale)).Day()
		if d > lastDayMonth {
			d = lastDayMonth
		}

		if ts {
			return SQLTimestamp{time.Date(y, time.Month(mo), d, t.Hour(), t.Minute(), t.Second(), 0, schema.DefaultLocale)}, nil
		}
		return SQLDate{time.Date(y, time.Month(mo), d, 0, 0, 0, 0, schema.DefaultLocale)}, nil
	case Week:
		if ts {
			return SQLTimestamp{t.AddDate(0, 0, int(v.Float64())*7)}, nil
		}
		return SQLDate{t.AddDate(0, 0, int(v.Float64())*7)}, nil
	case Day:
		if ts {
			return SQLTimestamp{t.AddDate(0, 0, int(v.Float64()))}, nil
		}
		return SQLDate{t.AddDate(0, 0, int(v.Float64()))}, nil
	case Hour:
		duration, _ := time.ParseDuration(v.String() + "h")
		return SQLTimestamp{t.Add(duration)}, nil
	case Minute:
		duration, _ := time.ParseDuration(v.String() + "m")
		return SQLTimestamp{t.Add(duration)}, nil
	case Second:
		duration, _ := time.ParseDuration(v.String() + "s")
		return SQLTimestamp{t.Add(duration)}, nil
	case Microsecond:
		duration, _ := time.ParseDuration(v.String() + "us")
		return SQLTimestamp{Time: t.Add(duration)}, nil
	default:
		return SQLNull, fmt.Errorf("cannot add '%v' to timestamp", values[0])
	}
}

func (t *timestampAddFunc) Type(exprs []SQLExpr) schema.SQLType {
	if v, ok := exprs[2].(SQLValue); ok {
		if len(v.String()) > 10 {
			return schema.SQLTimestamp
		}
	}
	return schema.SQLDate
}

func (t *timestampAddFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type timestampDiffFunc struct{}

//http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_timestampdiff
func (*timestampDiffFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t1, _, err := parseDateTime(values[1].String())
	if !err {
		return SQLNull, nil
	}

	t2, _, err := parseDateTime(values[2].String())
	if !err {
		return SQLNull, nil
	}

	duration := t2.Sub(t1)

	switch values[0].String() {
	case Year:
		return SQLInt(math.Floor(float64(numMonths(t1, t2) / 12))), nil
	case Quarter:
		return SQLInt(math.Floor(float64(numMonths(t1, t2) / 3))), nil
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

func (*timestampFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.versionAtLeast(3, 5, 0) {
		return nil, false
	}

	if len(exprs) != 1 {
		return nil, false
	}

	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
	}

	val := args[0]

	wrapInDateFromString := func(v interface{}) bson.M {
		return bson.M{"$dateFromString": bson.M{"dateString": v}}
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
	ifSix := wrapInOp(mgoOperatorAdd, wrapInOp(mgoOperatorMultiply, val, hhmmssFactor), getPadding(wrapInOp(mgoOperatorMultiply, val, hhmmssFactor)))
	sixBranch := wrapInCase(hasUpToXDigits(6), ifSix)

	// This number is YYYYMMDD, again, multiply by hhmmssFactor.
	eightBranch := wrapInCase(hasUpToXDigits(8), wrapInOp(mgoOperatorMultiply, val, hhmmssFactor))

	// If it's twelve digits, interpret as YYMMDDHHMMSS.  Make sure to pad the number.
	ifTwelve := wrapInOp(mgoOperatorAdd, val, getPadding(val))
	twelveBranch := wrapInCase(hasUpToXDigits(12), ifTwelve)

	// if fourteen, YYYYMMDDHHMMSS, we can use as it as is.
	fourteenBranch := wrapInCase(hasUpToXDigits(14), val)

	// define "num", the input number normalized to 14 digits, in a "let"
	numberVar := wrapInSwitch(nil, sixBranch, eightBranch, twelveBranch, fourteenBranch)
	numberLetVars := bson.M{"num": numberVar}

	dateParts := bson.M{
		// YYYYMMDDHHMMSS / 10000000000 = YYYY
		"year": bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide, "$$num", 10000000000)},
		// (YYYYMMDDHHMMSS / 100000000) % 100 = MM
		"month": wrapInOp(mgoOperatorMod, bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide, "$$num", 100000000)}, 100),
		// YYYYMMDDHHMMSS / 1000000) % 100 = DD
		"day": wrapInOp(mgoOperatorMod, bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide, "$$num", 1000000)}, 100),
		// YYYYMMDDHHMMSS / 10000) % 100 = HH
		"hour": wrapInOp(mgoOperatorMod, bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide, "$$num", 10000)}, 100),
		// YYYYMMDDHHMMSS / 100) % 100 = MM
		"minute": wrapInOp(mgoOperatorMod, bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorDivide, "$$num", 100)}, 100),
		// YYYYMMDDHHMMSS % 100 = SS
		"second": bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorMod, "$$num", 100)},
		// YYYYMMDDHHMMSS.FFFFF % 1 * 1000 = ms
		"millisecond": bson.M{mgoOperatorTrunc: wrapInOp(mgoOperatorMultiply, wrapInOp(mgoOperatorMod, "$$num", 1), 1000)},
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
		bson.M{"$dateFromParts": bson.M{
			"year":        "$$year",
			"month":       "$$month",
			"day":         "$$day",
			"hour":        "$$hour",
			"minute":      "$$minute",
			"second":      "$$second",
			"millisecond": "$$millisecond",
		}},
		nil,
		bson.M{mgoOperatorAnd: []interface{}{yearValid, monthValid, dayValid, hourValid, minuteValid, secondValid}},
	)

	evaluateNumber := wrapInLet(dateParts, makeDateOrNull)
	handleNumberToDate := wrapInLet(numberLetVars, evaluateNumber)
	numberBranch := wrapInCase(isNumber, handleNumberToDate)

	// CASE 3: it's a string
	isString := containsBSONType(val, "string")

	// First split on T, take first substring, then split that on " ", and take first
	// substring. this gives us just the date part of the string. note that if the
	// string doesn't have T or a space, just returns original string
	trimmedDateString := wrapInOp(mgoOperatorArrElemAt,
		wrapInOp(mgoOperatorSplit,
			wrapInOp(mgoOperatorArrElemAt,
				wrapInOp(mgoOperatorSplit, val, "T"),
				0),
			" "),
		0)

	// Repeat the step above but take the second element to get the time part.  Replace
	// with "" if we can not find a second element.
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

	// Convert the date and time strings to arrays so we can use map/reduce.
	trimmedDateAsArray := wrapInStringToArray("$$trimmedDate")
	trimmedTimeAsArray := wrapInStringToArray("$$trimmedTime")

	// isSeparator evaluates to true if a character is in the defined separator list
	isSeparator := wrapInOp(mgoOperatorNeq, -1, wrapInOp("$indexOfArray", dateComponentSeparator, "$$c"))

	// Use map to convert all separators in the date string to - symbol, and leave numbers as-is
	dateNormalized := wrapInMap(trimmedDateAsArray, "c", wrapInCond("-", "$$c", isSeparator))
	// Use map to convert all separators in the time string to '.' symbol, and leave numbers as-is.
	// We use '.' instead of ':' so that mongo correctly handles fractional seconds. 10.11.23.1234
	// is parsed correctly as 10:11:23.1234, saving us some effort (and runtime).
	timeNormalized := wrapInMap(trimmedTimeAsArray, "c", wrapInCond(".", "$$c", isSeparator))

	// Use reduce to convert characters back to a single string for date and time.
	dateJoined := wrapInReduce(dateNormalized, "", wrapInOp(mgoOperatorConcat, "$$value", "$$this"))
	timeJoined := wrapInReduce(timeNormalized, "", wrapInOp(mgoOperatorConcat, "$$value", "$$this"))

	// if the third character is a -, or if the string is only 6 digits long and has no slashes,
	// then the string is either format YY/MM/DD or YYMMDD and we need to add the appropriate first
	// two year digits (19xx or 20xx) for Mongo to understand it
	hasShortYear := wrapInOp(mgoOperatorOr,
		// length is only 6, assume YYMMDD
		wrapInOp(mgoOperatorEq, bson.M{mgoOperatorStrlenCP: "$$dateJoined"}, 6),
		// third character is -, assume YY-MM-DD
		wrapInOp(mgoOperatorEq, "-", bson.M{mgoOperatorSubstr: []interface{}{"$$dateJoined", 2, 1}}))

	// $dateFromString actually pads correctly, but not if "/" is used as the separator (it will assume year is last).
	// If this pushdown is shown to be slow by benchmarks, we should reconsider allowing $dateFromString to handle padding.
	// The change would not be trivial due to how MongoDB cannot handle short dates when there are no separators in the date.
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

	// we have to use nested $lets because in the outer one we define $$trimmedDate and
	// in the inner one we define $$dateJoined. defining $$dateJoined requires knowing the
	// length of trimmedDate, so we can't do it all in one step.
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

	stringBranch := wrapInCase(isString, outerLet)

	return wrapInSwitch(nil, dateBranch, numberBranch, stringBranch), true

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
// number of days since 0000-00-00.  There are a few interesting issues here:
// 1. 0000-00-00 is not a valid date, so TO_DAYS('0000-00-00') is supposed to return NULL
//    and TO_DAYS('0000-01-01') is supposed to be 1 rather than the 0 we return.
// 2. However, due to a bug in MySQL treating year 0 as a non-leap year, our results
//    are correct for any date after 0000-02-29 (which MySQL thinks isn't a day).
//    year zero should be a leap year: https://en.wikipedia.org/wiki/Year_zero,
//    and both MongoDB and the go time library treat it as such.
//    If, at some point, MySQL should correct their calendar, we could switch to adding
//    1 to our result to be inline with them.
type toDaysFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_to-days
func (*toDaysFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	df := &dateFunc{}

	// Reuse the date function to handle any padding and conversion from ints and strings.
	maybeSQLDate, err := df.Evaluate(values, ctx)
	if err != nil {
		return SQLNull, nil
	}

	switch typedD := maybeSQLDate.(type) {
	case SQLNullValue:
		return SQLNull, nil
	case SQLDate:
		date := typedD.Time
		start, _ := time.ParseInLocation(shortTimeFormat, "0000-01-01", schema.DefaultLocale)
		// maxGoDurationHours is the largest integer value of the maximum time.Duration.Hours()
		target, maxGoDurationHours := 1.0, int64(2562024)
		date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, schema.DefaultLocale)
		for date.Sub(start).Hours() > 24 {
			for date.Sub(start).Hours() > float64(maxGoDurationHours) {
				date = date.Add(time.Duration(-maxGoDurationHours) * time.Hour)
				// 106571 is 2562024/24, so the number of days per maximum duration
				target += float64(106751)
			}
			date, target = date.AddDate(0, 0, -1), target+1
		}

		return SQLInt(target), nil
	}
	// Should be unreachable because date() should only return a valid date or NULL.
	return SQLNull, nil
}

func (*toDaysFunc) Normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

// FuncToAggregation for TO_DAYS has one issue wrt how TO_DAYS is supposed to perform:
// because our date treatment is backed by using MongoDB's $dateFromString function,
// if a date that doesn't exist (e.g., 0000-00-00 or 0001-02-29) is entered, we return
// an error instead of the NULL expected from MySQL.  Unfortunately, checking for valid
// dates is too cost prohibitive.  If at some point $dateFromString supports an onError/default
// value, we should switch to using that.
func (*toDaysFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if len(exprs) != 1 {
		return nil, false
	}
	// Call the Date function FuncToAggregationLanguage on our argument, this will
	// give use an aggregation language expression that will convert the argument
	// to a proper mongo date.
	df := &dateFunc{}
	argConvertedToDate, ok := df.FuncToAggregationLanguage(t, exprs)
	if !ok {
		return nil, false
	}
	// Subtract dayOne (0000-01-01) from the argument in mongo, then convert ms to days.
	// When using $subtract on two dates in MongoDB, the number of ms between the two
	// dates is returned, and the purpose of the TO_DAYS function is to get the number
	// of days since 0000-01-01:
	// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_to-days
	// Unfortunately, we get a slightly wrong number if we try to multiply by days/ms
	// becuase MySQL itself is using division (and actually gets the wrong day count itself)
	dayOne := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	return bson.M{mgoOperatorTrunc: bson.M{mgoOperatorDivide: []interface{}{
		bson.M{mgoOperatorSubtract: []interface{}{argConvertedToDate, dayOne}},
		MillisecondsPerDay,
	}}}, true
}

func (*toDaysFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt64
}

func (*toDaysFunc) Validate(exprCount int) error {
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

func (*trimFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
	if !t.versionAtLeast(3, 4, 0) {
		return nil, false
	}
	if len(exprs) != 1 {
		return nil, false
	}
	args, ok := t.translateArgs(exprs)
	if !ok {
		return nil, false
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

func (*truncateFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (*ucaseFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLFloat(0.0), nil
	}

	// Our times are parsed as if in UTC. However, we need to
	// parse it in the actual location the server's running
	// in - to account for any timezone difference.
	y, m, d := t.Date()
	ts := time.Date(y, m, d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), now.Location())
	return SQLUint64(ts.Unix()), nil
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
	return SQLVarchar(ctx.ExecutionCtx.User()), nil
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

func (*utcDateFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (*utcTimestampFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	y := t.Year()
	d := time.Date(y, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
	iso := false
	mondayFirst := false
	smallRange := false
	weekday := int(d.Weekday())
	var days int
	var day1 int
	if len(values) == 2 {
		v, _ := values[1].(SQLInt)
		switch v {
		case 1:
			mondayFirst = true
			iso = true
			day1 = dayOneWeekOne(d, iso, mondayFirst)
		case 2:
			smallRange = true
			day1 = dayOneWeekOne(d, iso, mondayFirst)
		case 3:
			smallRange = true
			mondayFirst = true
			iso = true
			day1 = dayOneWeekOne(d, iso, mondayFirst)
		case 4:
			iso = true
			day1 = dayOneWeekOne(d, iso, mondayFirst)
		case 5:
			mondayFirst = true
			day1 = dayOneWeekOne(d, iso, mondayFirst)
		case 6:
			smallRange = true
			iso = true
			day1 = dayOneWeekOne(d, iso, mondayFirst)
		case 7:
			mondayFirst = true
			smallRange = true
			day1 = dayOneWeekOne(d, iso, mondayFirst)
		default:
			day1 = dayOneWeekOne(d, iso, mondayFirst)
		}
	} else {
		day1 = dayOneWeekOne(d, iso, mondayFirst)
	}

	if mondayFirst {
		if weekday == 0 {
			weekday = 7
		}
		weekday--
	}

	yearDay := t.YearDay()
	days = yearDay - day1

	if days < 0 {
		if !smallRange {
			return SQLInt(0), nil
		}
		y--
		d = time.Date(y, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
		t = time.Date(y, 12, 31, 0, 0, 0, 0, schema.DefaultLocale)
		day1 = dayOneWeekOne(d, iso, mondayFirst)
		days = t.YearDay() - day1
		return SQLInt(days/7 + 1), nil
	}

	if days < 7 && iso {
		firstDay := (8 - int(d.Weekday())) % 7
		if mondayFirst {
			firstDay++
		}
		if day1 == 0 {
			firstDay = 7
		}

		if yearDay >= firstDay {
			if firstDay == day1 {
				return SQLInt(1), nil
			}
			return SQLInt(2), nil
		}
	}

	if smallRange && days >= 52*7 {
		weekday = int(t.AddDate(1, 0, 0).Weekday())
		if mondayFirst {
			if weekday == 0 {
				weekday = 7
			}
			weekday--
		}
		if weekday < 4 {
			if iso || (!iso && weekday == 0) {
				return SQLInt(1), nil
			}
		}
	}

	return SQLInt(days/7 + 1), nil
}

func (*weekFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

		arg1Val, _ := NewSQLValue(bsonVal, schema.SQLInt, schema.SQLNone)
		mode = arg1Val.Int64()
	}

	if mode == 0 {
		return wrapSingleArgFuncWithNullCheck("$week", args[0]), true
	}
	return nil, false
}

func (*weekFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLDate, schema.SQLInt}
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

type weekOfYearFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_weekofyear
func (*weekOfYearFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	week := &weekFunc{}
	return week.Evaluate([]SQLValue{values[0], SQLInt(3)}, ctx)
}

func (*weekOfYearFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*weekOfYearFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
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

func (*weekdayFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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

func (*weekdayFunc) Reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
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

func (*yearFunc) FuncToAggregationLanguage(t *pushDownTranslator, exprs []SQLExpr) (interface{}, bool) {
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
	t, _, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	var v uint
	if len(values) == 2 {
		v = uint(values[1].(SQLInt))
		if v == 0 {
			v = 2
		} else if v == 1 {
			v = 3
		}
	} else {
		v = 2
	}

	weekFunc := &weekFunc{}
	w, _ := weekFunc.Evaluate([]SQLValue{values[0], SQLInt(v)}, ctx)

	week, ok := w.(SQLInt)
	if !ok {
		return SQLNull, nil
	}

	y := t.Year()
	wk := int(week)
	if t.Month() == 1 && (wk == 52 || wk == 53) {
		y--
	} else if t.Month() == 12 && (wk == 0 || wk == 1) {
		y++
	}

	return SQLInt(y*100 + wk), nil
}

func (*yearWeekFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (*yearWeekFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

func (t *pushDownTranslator) translateArgs(exprs []SQLExpr) ([]interface{}, bool) {
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

func NewSQLScalarFunctionExpr(name string, exprs []SQLExpr) (*SQLScalarFunctionExpr, error) {
	fun, ok := scalarFuncMap[name]
	if !ok {
		return nil, fmt.Errorf("scalar function '%v' is not supported", name)
	}

	sf := &SQLScalarFunctionExpr{name, fun, exprs}

	return sf.Reconcile(), nil
}

func convertAllArgs(f *SQLScalarFunctionExpr, convType schema.SQLType, defaultValue SQLValue) *SQLScalarFunctionExpr {
	nExprs := convertAllExprs(f.Exprs, convType, defaultValue)
	return &SQLScalarFunctionExpr{
		f.Name,
		f.Func,
		nExprs,
	}
}

func NewIfScalarFunctionExpr(condition, truePart, falsePart SQLExpr) *SQLScalarFunctionExpr {
	return &SQLScalarFunctionExpr{
		Name:  "if",
		Func:  &ifFunc{},
		Exprs: []SQLExpr{condition, truePart, falsePart},
	}
}

// areAllTimeTypes checks if all SQLValues are either type SQLTimestamp or SQLDate
// and there is at least one SQLTimestamp type. This is necessary because if the former is true,
// MySQL will always return a SQLTimestamp type in the greatest and least functions.
// i.e. SELECT GREATEST(DATE "2006-05-11", TIMESTAMP "2005-04-12", DATE "2004-06-04")
// returns TIMESTAMP "2006-05-11 00:00:00"
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

// calculateInterval converts each of the values in args to unit, and returns the sum of these multiplied by neg.
func calculateInterval(unit string, args []int, neg int) (string, int, error) {
	var val int
	var u string
	sp := strings.SplitAfter(unit, "_")
	if len(sp) > 1 {
		u = string(sp[1])
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
			val = args[0]*day*hour*minute*second + args[1]*hour*minute*second + args[2]*minute*second + args[3]*second + args[4]
		default:
			return unit, 0, fmt.Errorf("invalid argument length")
		}
	case 4:
		switch unit {
		case DayMicrosecond, HourMicrosecond:
			val = args[0]*hour*minute*second + args[1]*minute*second + args[2]*second + args[3]
		case DaySecond:
			val = args[0]*day*hour*minute + args[1]*hour*minute + args[2]*minute + args[3]
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

// dateArithmeticArgs parses val and returns an integer slice stripped of any spaces, colons, etc.
// It also returns whether the first character in val is "-", indicating whether the arguments should be negative.
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

func dayOneWeekOne(d time.Time, iso bool, monStart bool) int {
	day1 := (8 - int(d.Weekday())) % 7
	if monStart {
		day1++
	}
	if day1 == 0 {
		day1 = 7
	}
	if day1 > 4 && iso {
		day1 = 1
	}
	return day1
}

func ensureArgCount(exprCount int, counts ...int) error {
	// for scalar functions that accept a variable number of arguments
	if len(counts) == 1 && counts[0] == -1 {
		if exprCount == 0 {
			return ErrIncorrectVarCount
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
		return ErrIncorrectCount
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

	h, m, s, i := 0, 0, 0, 0
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
		hours, buf, i = buf[0:h], buf[h+1:], emitFrac(buf[0:h+1])
		if i != -1 {
			secs, hours = hours[:i], hours[:0]
		} else {
			m = emitToken(buf, ':')
			if m != 0 {
				mins, buf, i = buf[0:m], buf[m+1:], emitFrac(buf[0:m+1])
				if i != -1 {
					mins = mins[:i]
				} else {
					s = emitToken(buf, ':')
					if s != 0 {
						secs, i = buf[0:s], emitFrac(buf[0:s+1])
						if i != -1 {
							secs = secs[:i]
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
