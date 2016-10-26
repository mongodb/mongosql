package evaluator

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
)

const (
	shortTimeFormat = "2006-01-02"
)

var (
	zeroDate, _ = time.ParseInLocation(shortTimeFormat, "0000-00-00", schema.DefaultLocale)
)

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

type normalizingScalarFunc interface {
	normalize(*SQLScalarFunctionExpr) SQLExpr
}

func NewIfScalarFunctionExpr(condition, truePart, falsePart SQLExpr) *SQLScalarFunctionExpr {
	return &SQLScalarFunctionExpr{
		Name:  "if",
		Exprs: []SQLExpr{condition, truePart, falsePart},
	}
}

var scalarFuncMap = map[string]scalarFunc{
	"abs":     singleArgFloatMathFunc(math.Abs),
	"acos":    singleArgFloatMathFunc(math.Acos),
	"adddate": &addDateFunc{},
	"asin":    singleArgFloatMathFunc(math.Asin),
	"atan": multiArgFloatMathFunc{
		single: singleArgFloatMathFunc(math.Atan),
		dual:   dualArgFloatMathFunc(math.Atan2),
	},
	"atan2":             dualArgFloatMathFunc(math.Atan2),
	"ascii":             &asciiFunc{},
	"cast":              &castFunc{},
	"ceil":              singleArgFloatMathFunc(math.Ceil),
	"ceiling":           singleArgFloatMathFunc(math.Ceil),
	"char_length":       &characterLengthFunc{},
	"character_length":  &characterLengthFunc{},
	"coalesce":          &coalesceFunc{},
	"concat":            &concatFunc{},
	"concat_ws":         &concatWsFunc{},
	"connection_id":     &connectionIdFunc{},
	"convert":           &convertFunc{},
	"cos":               singleArgFloatMathFunc(math.Cos),
	"cot":               &cotFunc{},
	"curdate":           &currentDateFunc{},
	"current_date":      &currentDateFunc{},
	"current_timestamp": &currentTimestampFunc{},
	"current_user":      &userFunc{},
	"database":          &dbFunc{},
	"date":              &dateFunc{},
	"date_add":          &dateAddFunc{},
	"date_sub":          &dateSubFunc{},
	"day":               &dayOfMonthFunc{},
	"dayname":           &dayNameFunc{},
	"dayofmonth":        &dayOfMonthFunc{},
	"dayofweek":         &dayOfWeekFunc{},
	"dayofyear":         &dayOfYearFunc{},
	"degrees":           singleArgFloatMathFunc(func(f float64) float64 { return f * 180 / math.Pi }),
	"exp":               singleArgFloatMathFunc(math.Exp),
	"extract":           &extractFunc{},
	"floor":             singleArgFloatMathFunc(math.Floor),
	"from_days":         &fromDaysFunc{},
	"greatest":          &greatestFunc{},
	"hour":              &hourFunc{},
	"if":                &ifFunc{},
	"ifnull":            &ifnullFunc{},
	"instr":             &instrFunc{},
	"isnull":            &isnullFunc{},
	"last_day":          &lastDayFunc{},
	"lcase":             &lcaseFunc{},
	"least":             &leastFunc{},
	"left":              &leftFunc{},
	"length":            &lengthFunc{},
	"ln":                singleArgFloatMathFunc(math.Log),
	"locate":            &locateFunc{},
	"log":               singleArgFloatMathFunc(math.Log),
	"log2":              singleArgFloatMathFunc(math.Log2),
	"log10":             singleArgFloatMathFunc(math.Log10),
	"lower":             &lcaseFunc{},
	"ltrim":             &ltrimFunc{},
	"makedate":          &makeDateFunc{},
	"mid":               &midFunc{},
	"minute":            &minuteFunc{},
	"mod":               dualArgFloatMathFunc(math.Mod),
	"month":             &monthFunc{},
	"monthname":         &monthNameFunc{},
	"not":               &notFunc{},
	"now":               &currentTimestampFunc{},
	"nullif":            &nullifFunc{},
	"pi":                &constantFunc{SQLFloat(math.Pi)},
	"pow":               &powFunc{},
	"power":             &powFunc{},
	"quarter":           &quarterFunc{},
	"radians":           singleArgFloatMathFunc(func(f float64) float64 { return f * math.Pi / 180 }),
	"replace":           &replaceFunc{},
	"right":             &rightFunc{},
	"round":             &roundFunc{},
	"rtrim":             &rtrimFunc{},
	"schema":            &dbFunc{},
	"second":            &secondFunc{},
	"session_user":      &userFunc{},
	"sign":              &signFunc{},
	"sqrt":              singleArgFloatMathFunc(math.Sqrt),
	"space":             &spaceFunc{},
	"str_to_date":       &strToDateFunc{},
	"subdate":           &subDateFunc{},
	"substr":            &substringFunc{},
	"substring":         &substringFunc{},
	"system_user":       &userFunc{},
	"tan":               singleArgFloatMathFunc(math.Tan),
	"timediff":          &timeDiffFunc{},
	"timestampadd":      &timestampAddFunc{},
	"timestampdiff":     &timestampDiffFunc{},
	"time_to_sec":       &timeToSecFunc{},
	"trim":              &trimFunc{},
	"truncate":          &truncateFunc{},
	"ucase":             &ucaseFunc{},
	"upper":             &ucaseFunc{},
	"user":              &userFunc{},
	"utc_timestamp":     &utcTimestampFunc{},
	"version":           &versionFunc{},
	"week":              &weekFunc{},
	"weekday":           &weekdayFunc{},
	"weekofyear":        &weekOfYearFunc{},
	"year":              &yearFunc{},
	"yearweek":          &yearWeekFunc{},
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

func (f *SQLScalarFunctionExpr) normalize() node {
	if sf, ok := scalarFuncMap[f.Name]; ok {
		if nsf, ok := sf.(normalizingScalarFunc); ok {
			return nsf.normalize(f)
		}
	}

	return f
}

func (f *SQLScalarFunctionExpr) RequiresEvalCtx() bool {
	if sf, ok := scalarFuncMap[f.Name]; ok {
		if r, ok := sf.(RequiresEvalCtx); ok {
			return r.RequiresEvalCtx()
		}
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

func (f *SQLScalarFunctionExpr) Type() schema.SQLType {
	sf, ok := scalarFuncMap[f.Name]
	if ok {
		return sf.Type()
	}

	return schema.SQLNone
}

type constantFunc struct {
	value SQLValue
}

func (c *constantFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return c.value, nil
}

func (c *constantFunc) Type() schema.SQLType {
	return c.value.Type()
}

func (_ *constantFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
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

func (_ singleArgFloatMathFunc) Type() schema.SQLType {
	return schema.SQLFloat
}

func (_ singleArgFloatMathFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

func (_ singleArgFloatMathFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
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

func (_ dualArgFloatMathFunc) Type() schema.SQLType {
	return schema.SQLFloat
}

func (_ dualArgFloatMathFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

func (_ dualArgFloatMathFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
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

func (_ multiArgFloatMathFunc) Type() schema.SQLType {
	return schema.SQLFloat
}

func (_ multiArgFloatMathFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

func (_ multiArgFloatMathFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

type connectionIdFunc struct{}

func (_ *connectionIdFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLUint32(ctx.ExecutionCtx.ConnectionId()), nil
}

func (_ *connectionIdFunc) RequiresEvalCtx() bool {
	return true
}

func (_ *connectionIdFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *connectionIdFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type dbFunc struct{}

func (_ *dbFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLVarchar(ctx.ExecutionCtx.DB()), nil
}

func (_ *dbFunc) RequiresEvalCtx() bool {
	return true
}

func (_ *dbFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *dbFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type userFunc struct{}

func (_ *userFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLVarchar(ctx.ExecutionCtx.User()), nil
}

func (_ *userFunc) RequiresEvalCtx() bool {
	return true
}

func (_ *userFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *userFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type versionFunc struct{}

func (_ *versionFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLVarchar(ctx.Variables().Version), nil
}

func (_ *versionFunc) RequiresEvalCtx() bool {
	return true
}

func (_ *versionFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *versionFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type addDateFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_adddate
func (_ *addDateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	adder := &dateAddFunc{}
	return adder.Evaluate(values, ctx)
}

func (_ *addDateFunc) Type() schema.SQLType {
	return schema.SQLTimestamp
}

func (_ *addDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type asciiFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_ascii
func (_ *asciiFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

func (_ *asciiFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *asciiFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type castFunc struct{}

func (_ *castFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return NewSQLValue(values[0].Value(), schema.SQLType(values[1].String())), nil
}

func (_ *castFunc) Type() schema.SQLType {
	return schema.SQLNone
}

func (_ *castFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type characterLengthFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_char_length
func (_ *characterLengthFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	value := []rune(values[0].String())

	return SQLInt(len(value)), nil
}

func (_ *characterLengthFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *characterLengthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type coalesceFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_coalesce
func (_ *coalesceFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	for _, value := range values {
		if value != SQLNull {
			return value, nil
		}
	}
	return SQLNull, nil
}

func (_ *coalesceFunc) Type() schema.SQLType {
	return schema.SQLNone
}

func (_ *coalesceFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, -1)
}

type concatFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_concat
func (_ *concatFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (v SQLValue, err error) {
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

	var bytes bytes.Buffer
	for _, value := range values {
		bytes.WriteString(value.String())
	}

	v = SQLVarchar(bytes.String())
	err = nil
	return
}

func (_ *concatFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *concatFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *concatFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, -1)
}

type concatWsFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_concat-ws
func (_ *concatWsFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (v SQLValue, err error) {
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

	var bytes bytes.Buffer
	var separator string = values[0].String()
	var trimValues []SQLValue = values[1:]
	for i, value := range trimValues {
		if _, ok := value.(SQLNullValue); ok {
			continue
		}
		bytes.WriteString(value.String())
		if i != len(trimValues)-1 {
			bytes.WriteString(separator)
		}
	}

	v = SQLVarchar(bytes.String())
	return
}

func (_ *concatWsFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if f.Exprs[0] == SQLNull {
		return SQLNull
	}

	return f
}

func (_ *concatWsFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *concatWsFunc) Validate(exprCount int) error {
	if ensureArgCount(exprCount, -1) != nil || exprCount < 2 {
		return ErrIncorrectCount
	}
	return nil
}

type convertFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/cast-functions.html#function_convert
func (_ *convertFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	_, ok := values[0].(SQLNullValue)
	if ok {
		return SQLNull, nil
	}

	switch values[1].String() {
	case string(parser.INTEGER_BYTES):
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
			if typedV {
				i = 1
			} else {
				i = 0
			}
		default:
			return SQLNull, nil
		}

		return SQLInt(i), nil

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
			if typedV {
				f = float64(1)
			} else {
				f = float64(0)
			}
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
			if typedV {
				s = "1"
			} else {
				s = "0"
			}
		default:
			return SQLNull, nil
		}

		return SQLVarchar(s), nil

	case string(parser.DATE_BYTES):
		_, ok := values[0].(SQLTimestamp)
		var s string
		if ok {
			s = (values[0].String())[:10]
		} else {
			s = values[0].String()
		}
		t, ok := parseDateTime(s)
		if !ok {
			return SQLNull, nil
		}

		return SQLDate{Time: t}, nil

	case string(parser.DATETIME_BYTES):
		t, ok := parseDateTime(values[0].String())
		if !ok {
			return SQLNull, nil
		}

		return SQLTimestamp{Time: t}, nil

	default:
		return SQLNull, nil
	}
}

func (_ *convertFunc) Type() schema.SQLType {
	return schema.SQLNone
}

func (_ *convertFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type cotFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_cot
func (_ *cotFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	tan := math.Tan(values[0].Float64())
	if tan == 0 {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_DATA_OUT_OF_RANGE, "DOUBLE", fmt.Sprintf("'cot(%v)'", values[0].Float64()))
	}

	return SQLFloat(1 / tan), nil
}

func (_ *cotFunc) Type() schema.SQLType {
	return schema.SQLFloat
}

func (_ *cotFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type currentDateFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_curdate
func (_ *currentDateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, schema.DefaultLocale)
	return SQLDate{t}, nil

}

func (_ *currentDateFunc) Type() schema.SQLType {
	return schema.SQLDate
}

func (_ *currentDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type currentTimestampFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_now
func (_ *currentTimestampFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	value := time.Now().UTC()
	return SQLTimestamp{value}, nil
}

func (_ *currentTimestampFunc) Type() schema.SQLType {
	return schema.SQLTimestamp
}

func (_ *currentTimestampFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type dateFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_date
func (_ *dateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if _, ok := values[0].(SQLDate); !ok {
		if _, ok := values[0].(SQLTimestamp); !ok {
			return SQLNull, nil
		}
	}

	t, ok := parseDateTime((values[0].String())[:10])
	if !ok {
		return SQLNull, nil
	}

	return SQLDate{Time: t}, nil
}

func (_ *dateFunc) Type() schema.SQLType {
	return schema.SQLDate
}

func (_ *dateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dateAddFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_date-add
func (_ *dateAddFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	_, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	timestampadd := &timestampAddFunc{}
	args, neg := dateArithmeticArgs(values[1])
	unit, interval, err := calculateInterval(values[2].String(), args, neg)

	if err != nil {
		return SQLNull, nil
	}

	return timestampadd.Evaluate([]SQLValue{SQLVarchar(unit), SQLInt(interval), values[0]}, ctx)

}

func (_ *dateAddFunc) Type() schema.SQLType {
	return schema.SQLNone
}

func (_ *dateAddFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type dateSubFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_date-sub
func (_ *dateSubFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	dateadd := &dateAddFunc{}

	v := values[1].String()
	if string(v[0]) != "-" {
		v = "-" + v
	} else {
		v = v[1:]
	}

	return dateadd.Evaluate([]SQLValue{values[0], SQLVarchar(v), values[2]}, ctx)
}

func (_ *dateSubFunc) Type() schema.SQLType {
	return schema.SQLNone
}

func (_ *dateSubFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type dayNameFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayname
func (_ *dayNameFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLVarchar(t.Weekday().String()), nil
}

func (_ *dayNameFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *dayNameFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dayOfMonthFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayofmonth
func (_ *dayOfMonthFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Day())), nil
}

func (_ *dayOfMonthFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *dayOfMonthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dayOfWeekFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayofweek
func (_ *dayOfWeekFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Weekday()) + 1), nil
}

func (_ *dayOfWeekFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *dayOfWeekFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dayOfYearFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayofyear
func (_ *dayOfYearFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.YearDay())), nil
}

func (_ *dayOfYearFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *dayOfYearFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

const (
	YEAR               = "year"
	QUARTER            = "quarter"
	MONTH              = "month"
	WEEK               = "week"
	DAY                = "day"
	HOUR               = "hour"
	MINUTE             = "minute"
	SECOND             = "second"
	MICROSECOND        = "microsecond"
	YEAR_MONTH         = "year_month"
	DAY_HOUR           = "day_hour"
	DAY_MINUTE         = "day_minute"
	DAY_SECOND         = "day_second"
	DAY_MICROSECOND    = "day_microsecond"
	HOUR_MINUTE        = "hour_minute"
	HOUR_SECOND        = "hour_second"
	HOUR_MICROSECOND   = "hour_microsecond"
	MINUTE_SECOND      = "minute_second"
	MINUTE_MICROSECOND = "minute_microsecond"
	SECOND_MICROSECOND = "second_microsecond"
)

type extractFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_extract
func (_ *extractFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[1].String())
	if !ok {
		return SQLNull, nil
	}

	units := [6]int{t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second()}

	var unitStrs [6]string
	// For certain units, we need to concatenate the unit values as strings before returning the int value
	// as to not lose any number's place value.
	// i.e. SELECT EXTRACT(DAY_MINUTE FROM "2006-04-03 06:03:23") should return 30603, not 363.
	for idx, val := range units {
		u := strconv.Itoa(val)
		if len(u) == 1 {
			u = "0" + u
		}
		unitStrs[idx] = u
	}

	switch values[0].String() {
	case YEAR:
		return SQLInt(units[0]), nil
	case QUARTER:
		return SQLInt(int(math.Ceil(float64(units[1]) / 3.0))), nil
	case MONTH:
		return SQLInt(units[1]), nil
	case WEEK:
		_, w := t.ISOWeek()
		return SQLInt(w), nil
	case DAY:
		return SQLInt(units[2]), nil
	case HOUR:
		return SQLInt(units[3]), nil
	case MINUTE:
		return SQLInt(units[4]), nil
	case SECOND:
		return SQLInt(units[5]), nil
	case MICROSECOND:
		return SQLInt(0), nil
	case YEAR_MONTH:
		ym, _ := strconv.Atoi(unitStrs[0] + unitStrs[1])
		return SQLInt(ym), nil
	case DAY_HOUR:
		dh, _ := strconv.Atoi(unitStrs[2] + unitStrs[3])
		return SQLInt(dh), nil
	case DAY_MINUTE:
		dm, _ := strconv.Atoi(unitStrs[2] + unitStrs[3] + unitStrs[4])
		return SQLInt(dm), nil
	case DAY_SECOND:
		ds, _ := strconv.Atoi(unitStrs[2] + unitStrs[3] + unitStrs[4] + unitStrs[5])
		return SQLInt(ds), nil
	case DAY_MICROSECOND:
		dms, _ := strconv.Atoi(unitStrs[2] + unitStrs[3] + unitStrs[4] + unitStrs[5] + "000000")
		return SQLInt(dms), nil
	case HOUR_MINUTE:
		hm, _ := strconv.Atoi(unitStrs[3] + unitStrs[4])
		return SQLInt(hm), nil
	case HOUR_SECOND:
		hs, _ := strconv.Atoi(unitStrs[3] + unitStrs[4] + unitStrs[5])
		return SQLInt(hs), nil
	case HOUR_MICROSECOND:
		hms, _ := strconv.Atoi(unitStrs[3] + unitStrs[4] + unitStrs[5] + "000000")
		return SQLInt(hms), nil
	case MINUTE_SECOND:
		ms, _ := strconv.Atoi(unitStrs[4] + unitStrs[5])
		return SQLInt(ms), nil
	case MINUTE_MICROSECOND:
		mms, _ := strconv.Atoi(unitStrs[4] + unitStrs[5] + "000000")
		return SQLInt(mms), nil
	case SECOND_MICROSECOND:
		sms, _ := strconv.Atoi(unitStrs[5] + "000000")
		return SQLInt(sms), nil
	default:
		return nil, fmt.Errorf("unit type '%v' is not supported", values[0].String())
	}
}

func (_ *extractFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *extractFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type fromDaysFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_from-days
func (_ *fromDaysFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	value := values[0].Float64()

	if value <= 365.5 || value >= 3652499.5 {
		// Go's zero time starts January 1, year 1, 00:00:00 UTC
		// and thus can not represent the date "0000-00-00". To
		// handle this, we return a varchar instead
		return SQLVarchar("0000-00-00"), nil
	}

	target, maxGoDurationHours := int64(math.Floor(math.Abs(value-366))), int64(106751)

	date := zeroDate

	for target > 0 && target > maxGoDurationHours {
		date = date.Add(time.Duration(maxGoDurationHours*24) * time.Hour)
		target -= maxGoDurationHours
	}

	date = date.Add(time.Duration(target*24) * time.Hour)

	return SQLDate{date.In(schema.DefaultLocale)}, nil
}

func (_ *fromDaysFunc) Type() schema.SQLType {
	return schema.SQLDate
}

func (_ *fromDaysFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type greatestFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_greatest
func (_ *greatestFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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
		t, _ := parseDateTime(values[greatestIdx].String())
		return SQLTimestamp{Time: t}, nil
	} else if convertTo == schema.SQLInt || convertTo == schema.SQLFloat {
		return convertedVals[greatestIdx], nil
	} else {
		return values[greatestIdx], nil
	}
}

func (_ *greatestFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *greatestFunc) Type() schema.SQLType {
	return schema.SQLNone
}

func (_ *greatestFunc) Validate(exprCount int) error {
	if exprCount < 2 {
		return ErrIncorrectVarCount
	}
	return nil
}

type hourFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_hour
func (_ *hourFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Hour())), nil
}

func (_ *hourFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *hourFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type ifFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/control-flow-functions.html#function_if
func (_ *ifFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	switch typedV := values[0].(type) {
	case SQLBool:
		if typedV {
			return values[1], nil
		} else {
			return values[2], nil
		}
	case SQLDate, SQLTimestamp, SQLObjectID:
		return values[1], nil
	case SQLInt, SQLFloat:
		v := typedV.Float64()
		if v == 0 {
			return values[2], nil
		} else {
			return values[1], nil
		}
	case SQLNullValue:
		return values[2], nil
	case SQLVarchar:
		if v, _ := strconv.ParseFloat(typedV.String(), 64); v == 0 {
			return values[2], nil
		} else {
			return values[1], nil
		}
	default:
		return nil, fmt.Errorf("expression type '%v' is not supported", typedV)
	}
	return SQLNull, nil
}

func (_ *ifFunc) Type() schema.SQLType {
	return schema.SQLNone
}

func (_ *ifFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type ifnullFunc struct{}

func (_ *ifnullFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if _, ok := values[0].(SQLNullValue); ok {
		return values[1], nil
	} else {
		return values[0], nil
	}
}

func (_ *ifnullFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if f.Exprs[0] == SQLNull {
		return f.Exprs[1]
	} else if v, ok := f.Exprs[0].(SQLValue); ok {
		return v
	}

	return f
}

func (_ *ifnullFunc) Type() schema.SQLType {
	return schema.SQLNone
}

func (_ *ifnullFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type instrFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_instr
func (_ *instrFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	locate := &locateFunc{}
	return locate.Evaluate([]SQLValue{values[1], values[0]}, ctx)
}

func (_ *instrFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *instrFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type isnullFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_isnull
func (_ *isnullFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	_, ok := values[0].(SQLNullValue)
	matcher := SQLBool(ok)

	result, err := Matches(matcher, ctx)
	if err != nil {
		return nil, err
	}
	if SQLBool(result) == SQLTrue {
		return SQLInt(1), nil
	}
	return SQLInt(0), nil
}

func (_ *isnullFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *isnullFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type lastDayFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_last-day
func (_ *lastDayFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	year, month, _ := t.Date()
	first := time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
	return SQLDate{first.AddDate(0, 1, -1)}, nil
}

func (_ *lastDayFunc) Type() schema.SQLType {
	return schema.SQLDate
}

func (_ *lastDayFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type lcaseFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_lcase
func (_ *lcaseFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	value := strings.ToLower(values[0].String())

	return SQLVarchar(value), nil
}

func (_ *lcaseFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *lcaseFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type leastFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_least
func (_ *leastFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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
		t, _ := parseDateTime(values[leastIdx].String())
		return SQLTimestamp{Time: t}, nil
	} else if convertTo == schema.SQLInt || convertTo == schema.SQLFloat {
		return convertedVals[leastIdx], nil
	} else {
		return values[leastIdx], nil
	}
}

func (_ *leastFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *leastFunc) Type() schema.SQLType {
	return schema.SQLNone
}

func (_ *leastFunc) Validate(exprCount int) error {
	if exprCount < 2 {
		return ErrIncorrectVarCount
	}
	return nil
}

type leftFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_left
func (_ *leftFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	substring := &substringFunc{}
	return substring.Evaluate([]SQLValue{values[0], SQLInt(1), values[1]}, ctx)
}

func (_ *leftFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *leftFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type lengthFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_length
func (_ *lengthFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	value := values[0].String()

	return SQLInt(len(value)), nil
}

func (_ *lengthFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *lengthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type locateFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_locate
func (_ *locateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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
		} else {
			str = str[pos:]
			result = runesIndex(str, substr)
			if result >= 0 {
				result += pos
			}
		}
	} else {
		result = runesIndex(str, substr)
	}

	return SQLInt(result + 1), nil
}

func (_ *locateFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *locateFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *locateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2, 3)
}

type ltrimFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_ltrim
func (_ *ltrimFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	value := strings.TrimLeft(values[0].String(), " ")

	return SQLVarchar(value), nil
}

func (_ *ltrimFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *ltrimFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type makeDateFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_makedate
func (_ *makeDateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

func (_ *makeDateFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *makeDateFunc) Type() schema.SQLType {
	return schema.SQLDate
}

func (_ *makeDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type midFunc struct {
	wrapped substringFunc
}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_mid
func (m *midFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return m.wrapped.Evaluate(values, ctx)
}

func (m *midFunc) Type() schema.SQLType {
	return m.wrapped.Type()
}

func (_ *midFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type minuteFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_minute
func (_ *minuteFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Minute())), nil
}

func (_ *minuteFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *minuteFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type monthFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_month
func (_ *monthFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Month())), nil
}

func (_ *monthFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *monthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type monthNameFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_monthname
func (_ *monthNameFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLVarchar(t.Month().String()), nil
}

func (_ *monthNameFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *monthNameFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type notFunc struct{}

func (_ *notFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	matcher := &SQLNotExpr{values[0]}
	result, err := Matches(matcher, ctx)
	if err != nil {
		return nil, err
	}
	if SQLBool(result) == SQLTrue {
		return SQLInt(1), nil
	}
	return SQLInt(0), nil
}

func (_ *notFunc) Type() schema.SQLType {
	return schema.SQLBoolean
}

func (_ *notFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type nullifFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/control-flow-functions.html
func (_ *nullifFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

func (_ *nullifFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if f.Exprs[0] == SQLNull {
		return SQLNull
	}

	if f.Exprs[1] == SQLNull {
		return f.Exprs[0]
	}

	return f
}

func (_ *nullifFunc) Type() schema.SQLType {
	return schema.SQLNone
}

func (_ *nullifFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type quarterFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_quarter
func (_ *quarterFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
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

func (_ *quarterFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *quarterFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type powFunc struct{}

func (_ *powFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	n := math.Pow(values[0].Float64(), values[1].Float64())
	if math.IsNaN(n) {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ER_DATA_OUT_OF_RANGE, "DOUBLE", fmt.Sprintf("pow(%v,%v)", values[0].Float64(), values[1].Float64()))
	}

	return SQLFloat(math.Pow(values[0].Float64(), values[1].Float64())), nil
}

func (_ *powFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *powFunc) Type() schema.SQLType {
	return schema.SQLFloat
}

func (_ *powFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type replaceFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_replace
func (_ *replaceFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	s := values[0].String()
	old := values[1].String()
	new := values[2].String()

	return SQLVarchar(strings.Replace(s, old, new, -1)), nil
}

func (_ *replaceFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *replaceFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type rightFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_right
func (_ *rightFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	substring := &substringFunc{}
	len := -1 * values[1].Int64()
	return substring.Evaluate([]SQLValue{values[0], SQLInt(len)}, ctx)
}

func (_ *rightFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *rightFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type rtrimFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_rtrim
func (_ *rtrimFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	value := strings.TrimRight(values[0].String(), " ")

	return SQLVarchar(value), nil
}

func (_ *rtrimFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *rtrimFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type roundFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_round
func (_ *roundFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

func (_ *roundFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *roundFunc) Type() schema.SQLType {
	return schema.SQLFloat
}

func (_ *roundFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

type secondFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_second
func (_ *secondFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(int(t.Second())), nil
}

func (_ *secondFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *secondFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type signFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_sign
func (_ *signFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	v := values[0].Float64()
	if v < 0 {
		return SQLInt(-1), nil
	}
	if v > 0 {
		return SQLInt(1), nil
	}
	return SQLInt(0), nil
}

func (_ *signFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *signFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type spaceFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_space
func (_ *spaceFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	n := values[0].Int64()
	if n < 1 {
		return SQLVarchar(""), nil
	}

	return SQLVarchar(strings.Repeat(" ", int(n))), nil
}

func (_ *spaceFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *spaceFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type strToDateFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_str-to-date
func (_ *strToDateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

func (_ *strToDateFunc) Type() schema.SQLType {
	return schema.SQLNone
}

func (_ *strToDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type subDateFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_subdate
func (_ *subDateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	subtractor := &dateSubFunc{}
	return subtractor.Evaluate(values, ctx)
}

func (_ *subDateFunc) Type() schema.SQLType {
	return schema.SQLTimestamp
}

func (_ *subDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2, 3)
}

type substringFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_substring
func (_ *substringFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	str := []rune(values[0].String())
	pos := int(values[1].Float64())

	if pos > len(str) {
		return SQLVarchar(""), nil
	} else if pos < 0 {
		pos = len(str) + pos

		if pos < 0 {
			pos = 0
		}
	} else {
		pos-- // MySQL uses 1 as a basis
	}

	if len(values) == 3 {
		length := int(values[2].Float64())
		if length < 1 {
			return SQLVarchar(""), nil
		}

		if pos+length <= len(str) {
			str = str[pos : pos+length]
		}
	} else {
		if pos < len(str) {
			str = str[pos:]
		}
	}

	return SQLVarchar(string(str)), nil
}

func (_ *substringFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *substringFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *substringFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2, 3)
}

type timeDiffFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_timediff
func (_ *timeDiffFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	expr1, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	expr2, ok := parseDateTime(values[1].String())
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

	fmtFrac := func(buf []byte, v uint64, prec int) (nw int, nv uint64) {
		w := len(buf)
		print := false
		for i := 0; i < prec; i++ {
			digit := v % 10
			print = print || digit != 0
			if print {
				w--
				buf[w] = byte(digit) + '0'
			}
			v /= 10
		}

		for len(buf)-w < 6 {
			w--
			buf[w] = '0'
		}
		w--
		buf[w] = '.'
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

	buf := [16]byte{}
	w := len(buf)

	u := uint64(d)
	neg := d < 0
	if neg {
		u = -u
	}

	if u < uint64(time.Second) {
		var prec int
		switch {
		case u == 0:
			return SQLVarchar("00:00:00.000000"), nil
		default:
			prec = 6
		}
		w, u = fmtFrac(buf[:w], u, prec)
		w = fmtInt(buf[:w], u)
		w = fmtDef(buf[:w], 0)
	} else {
		w, u = fmtFrac(buf[:w], u, 9)

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

func (_ *timeDiffFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

func (_ *timeDiffFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

type timestampAddFunc struct{}

//http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_timestampadd
func (_ *timestampAddFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[2].String())
	if !ok {
		return SQLNull, nil
	}

	v := values[1]

	ts := false
	if len(values[2].String()) > 10 {
		ts = true
	}

	switch values[0].String() {
	case YEAR:
		if ts {
			return SQLTimestamp{Time: t.AddDate(int(v.Int64()), 0, 0)}, nil
		}
		return SQLDate{t.AddDate(int(v.Int64()), 0, 0)}, nil
	case QUARTER:
		y, m, d := t.Date()
		mo := int(((int64(m)+v.Int64()*3)%12 + 12) % 12)
		if mo == 0 {
			mo = 12
		}
		if v.Int64()*3 >= 12 || (v.Int64()*3) <= -12 {
			y += int(v.Int64() * 3 / 12)
		}
		if mo-int(m) < 0 && (v.Int64()*3) > 0 {
			y += 1
		} else if mo-int(m) > 0 && (v.Int64()*3) < 0 {
			y -= 1
		}
		lastDayMonth := 32 - (time.Date(y, time.Month(mo), 32, 0, 0, 0, 0, schema.DefaultLocale)).Day()
		if d > lastDayMonth {
			d = lastDayMonth
		}

		if ts {
			return SQLTimestamp{time.Date(y, time.Month(mo), d, t.Hour(), t.Minute(), t.Second(), 0, schema.DefaultLocale)}, nil
		}
		return SQLDate{time.Date(y, time.Month(mo), d, 0, 0, 0, 0, schema.DefaultLocale)}, nil
	case MONTH:
		y, m, d := t.Date()
		mo := int(((int64(m)+v.Int64())%12 + 12) % 12)
		if mo == 0 {
			mo = 12
		}
		if v.Int64() >= 12 || v.Int64() <= -12 {
			y += int(v.Int64() / 12)
		}
		if mo-int(m) < 0 && v.Int64() > 0 {
			y += 1
		} else if mo-int(m) > 0 && v.Int64() < 0 {
			y -= 1
		}
		lastDayMonth := 32 - (time.Date(y, time.Month(mo), 32, 0, 0, 0, 0, schema.DefaultLocale)).Day()
		if d > lastDayMonth {
			d = lastDayMonth
		}

		if ts {
			return SQLTimestamp{time.Date(y, time.Month(mo), d, t.Hour(), t.Minute(), t.Second(), 0, schema.DefaultLocale)}, nil
		}
		return SQLDate{time.Date(y, time.Month(mo), d, 0, 0, 0, 0, schema.DefaultLocale)}, nil
	case WEEK:
		if ts {
			return SQLTimestamp{t.AddDate(0, 0, int(v.Float64())*7)}, nil
		}
		return SQLDate{t.AddDate(0, 0, int(v.Float64())*7)}, nil
	case DAY:
		if ts {
			return SQLTimestamp{t.AddDate(0, 0, int(v.Float64()))}, nil
		}
		return SQLDate{t.AddDate(0, 0, int(v.Float64()))}, nil
	case HOUR:
		duration, _ := time.ParseDuration(v.String() + "h")
		return SQLTimestamp{t.Add(duration)}, nil
	case MINUTE:
		duration, _ := time.ParseDuration(v.String() + "m")
		return SQLTimestamp{t.Add(duration)}, nil
	case SECOND:
		duration, _ := time.ParseDuration(v.String() + "s")
		return SQLTimestamp{t.Add(duration)}, nil
	case MICROSECOND:
		// Microsecond not supported, so return the original time
		return SQLTimestamp{Time: t}, nil
	default:
		return nil, fmt.Errorf("cannot add '%v' to timestamp", values[0])
	}
	return SQLNull, nil
}

func (t *timestampAddFunc) Type() schema.SQLType {
	return schema.SQLNone
}

func (t *timestampAddFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type timestampDiffFunc struct{}

//http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_timestampdiff
func (_ *timestampDiffFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t1, err := parseDateTime(values[1].String())
	if !err {
		return SQLNull, nil
	}

	t2, err := parseDateTime(values[2].String())
	if !err {
		return SQLNull, nil
	}

	duration := t2.Sub(t1)

	switch values[0].String() {
	case YEAR:
		return SQLInt(math.Floor(float64(numMonths(t1, t2) / 12))), nil
	case QUARTER:
		return SQLInt(math.Floor(float64(numMonths(t1, t2) / 3))), nil
	case MONTH:
		return SQLInt(numMonths(t1, t2)), nil
	case WEEK:
		if t1.After(t2) {
			return SQLInt(math.Ceil((duration.Hours()) / 24 / 7)), nil
		}
		return SQLInt(math.Floor((duration.Hours()) / 24 / 7)), nil
	case DAY:
		if t1.After(t2) {
			return SQLInt(math.Ceil(duration.Hours() / 24)), nil
		}
		return SQLInt(math.Floor(duration.Hours() / 24)), nil
	case HOUR:
		return SQLInt(duration.Hours()), nil
	case MINUTE:
		return SQLInt(duration.Minutes()), nil
	case SECOND:
		return SQLInt(duration.Seconds()), nil
	case MICROSECOND:
		return SQLInt(duration.Nanoseconds() / 1000), nil
	default:
		return nil, fmt.Errorf("cannot add '%v' to timestamp", values[0])
	}
}

func (t *timestampDiffFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (t *timestampDiffFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type timeToSecFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_time-to-sec
func (_ *timeToSecFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	components := strings.Split(values[0].String(), ":")
	result := 0.0
	componentized := true

	if len(components) == 1 {
		cmp, err := strconv.ParseFloat(components[0], 64)
		if err != nil {
			return nil, err
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
			return nil, err
		}

		cmp := math.Trunc(float64(component))

		switch i {
		// more on valid time types at https://dev.mysql.com/doc/refman/5.5/en/time.html
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

func (_ *timeToSecFunc) Type() schema.SQLType {
	return schema.SQLFloat
}

func (_ *timeToSecFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type trimFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_trim
func (_ *trimFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

func (_ *trimFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *trimFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 3)
}

type truncateFunc struct{}

//http://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_truncate
func (_ *truncateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

func (_ *truncateFunc) Type() schema.SQLType {
	return schema.SQLFloat
}

func (_ *truncateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type ucaseFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_ucase
func (_ *ucaseFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	value := strings.ToUpper(values[0].String())

	return SQLVarchar(value), nil
}

func (_ *ucaseFunc) Type() schema.SQLType {
	return schema.SQLVarchar
}

func (_ *ucaseFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type utcTimestampFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_utc-timestamp
func (_ *utcTimestampFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLTimestamp{time.Now().In(time.UTC)}, nil
}

func (_ *utcTimestampFunc) Type() schema.SQLType {
	return schema.SQLTimestamp
}

func (_ *utcTimestampFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type weekFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_week
func (_ *weekFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
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
		weekday -= 1
	}

	yearDay := t.YearDay()
	days = yearDay - day1

	if days < 0 {
		if !smallRange {
			return SQLInt(0), nil
		} else {
			y -= 1
			d = time.Date(y, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
			t = time.Date(y, 12, 31, 0, 0, 0, 0, schema.DefaultLocale)
			day1 = dayOneWeekOne(d, iso, mondayFirst)
			days = t.YearDay() - day1
			return SQLInt(days/7 + 1), nil
		}
	}

	if days < 7 && iso {
		firstDay := (8 - int(d.Weekday())) % 7
		if mondayFirst {
			firstDay += 1
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
			weekday -= 1
		}
		if weekday < 4 {
			if iso || (!iso && weekday == 0) {
				return SQLInt(1), nil
			}
		}
	}

	return SQLInt(days/7 + 1), nil
}

func (_ *weekFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *weekFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

type weekdayFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_weekday
func (_ *weekdayFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	w := int(t.Weekday())
	if w == 0 {
		w = 7
	}
	return SQLInt(w - 1), nil
}

func (_ *weekdayFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *weekdayFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type weekOfYearFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_weekofyear
func (_ *weekOfYearFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	week := &weekFunc{}
	return week.Evaluate([]SQLValue{values[0], SQLInt(3)}, ctx)
}

func (_ *weekOfYearFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *weekOfYearFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type yearFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_year
func (_ *yearFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLInt(t.Year()), nil
}

func (_ *yearFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *yearFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type yearWeekFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_yearweek
func (_ *yearWeekFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
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
		y -= 1
	} else if t.Month() == 12 && (wk == 0 || wk == 1) {
		y += 1
	}

	return SQLInt(y*100 + wk), nil
}

func (_ *yearWeekFunc) Type() schema.SQLType {
	return schema.SQLInt
}

func (_ *yearWeekFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

// Helper functions

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

	switch len(args) {
	case 4:
		if unit != DAY_SECOND && unit != DAY_MICROSECOND {
			return unit, 0, fmt.Errorf("invalid argument length")
		}
		val = args[0]*24*60*60 + args[1]*60*60 + args[2]*60 + args[3]
	case 3:
		switch unit {
		case DAY_MINUTE:
			val = args[0]*24*60 + args[1]*60 + args[2]
		case DAY_SECOND, DAY_MICROSECOND, HOUR_SECOND, HOUR_MICROSECOND:
			val = args[0]*60*60 + args[1]*60 + args[2]
		default:
			return unit, 0, fmt.Errorf("invalid argument length")
		}
	case 2:
		switch unit {
		case YEAR_MONTH:
			val = args[0]*12 + args[1]
		case DAY_HOUR:
			val = args[0]*24 + args[1]
		case DAY_MINUTE, HOUR_MINUTE, DAY_SECOND, DAY_MICROSECOND, HOUR_SECOND, HOUR_MICROSECOND, MINUTE_SECOND, MINUTE_MICROSECOND:
			val = args[0]*60 + args[1]
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
		t, _ := parseDateTime(val.String())
		return SQLDate{Time: t}
	case schema.SQLTimestamp:
		t, _ := parseDateTime(val.String())
		return SQLTimestamp{Time: t}
	case schema.SQLBoolean:
		if val.Float64() == 0 {
			return SQLFalse
		}
		return SQLTrue
	default:
		return SQLInt(0)
	}
	return SQLInt(0)
}

func dayOneWeekOne(d time.Time, iso bool, monStart bool) int {
	day1 := (8 - int(d.Weekday())) % 7
	if monStart {
		day1 += 1
	}
	if day1 == 0 {
		day1 = 7
	}
	if day1 > 4 && iso {
		day1 = 1
	}
	return day1
}

// dateArithmeticArgs parses val and returns an integer slice stripped of any spaces, colons, etc.
// It also returns whether the first character in val is "-", indicating whether the arguments should be negative.
func dateArithmeticArgs(val SQLValue) ([]int, int) {
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
	c, _ := strconv.Atoi(curr)
	args = append(args, c)
	return args, neg
}

func evaluateArgs(exprs []SQLExpr, ctx *EvalCtx) ([]SQLValue, error) {

	values := []SQLValue{}

	for _, expr := range exprs {
		value, err := expr.Evaluate(ctx)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
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

func numMonths(startDate time.Time, endDate time.Time) int {
	y1, m1, d1 := startDate.Date()
	y2, m2, d2 := endDate.Date()
	months := ((y2 - y1) * 12) + (int(m2) - int(m1))
	if endDate.After(startDate) {
		if d2 < d1 {
			months -= 1
		} else if d1 == d2 {
			h1, mn1, s1 := startDate.Clock()
			ns1 := startDate.Nanosecond()
			h2, mn2, s2 := endDate.Clock()
			ns2 := endDate.Nanosecond()
			t1 := time.Date(y1, m1, d1, h1, mn1, s1, ns1, schema.DefaultLocale)
			t2 := time.Date(y1, m1, d1, h2, mn2, s2, ns2, schema.DefaultLocale)
			if t1.After(t2) {
				months -= 1
			}
		}
	} else {
		if d1 < d2 {
			months += 1
		} else if d1 == d2 {
			h1, mn1, s1 := startDate.Clock()
			ns1 := startDate.Nanosecond()
			h2, mn2, s2 := endDate.Clock()
			ns2 := endDate.Nanosecond()
			t1 := time.Date(y1, m1, d1, h1, mn1, s1, ns1, schema.DefaultLocale)
			t2 := time.Date(y1, m1, d1, h2, mn2, s2, ns2, schema.DefaultLocale)
			if t2.After(t1) {
				months += 1
			}
		}
	}
	return months
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

func parseDateTime(value string) (time.Time, bool) {
	for _, f := range schema.TimestampCtorFormats {
		t, err := time.Parse(f, value)
		if err == nil {
			return t, true
		}
	}
	return time.Time{}, false
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
