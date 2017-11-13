package evaluator

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/10gen/sqlproxy/variable"

	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
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

func NewSQLScalarFunctionExpr(name string, exprs []SQLExpr) (*SQLScalarFunctionExpr, error) {
	_, ok := scalarFuncMap[name]
	if !ok {
		return nil, fmt.Errorf("scalar function '%v' is not supported", name)
	}

	sf := &SQLScalarFunctionExpr{name, exprs}

	return sf.reconcile(), nil
}

type scalarFunc interface {
	Evaluate([]SQLValue, *EvalCtx) (SQLValue, error)
	Validate(exprCount int) error
	Type([]SQLExpr) schema.SQLType
}

type reconcilingScalarFunc interface {
	reconcile(*SQLScalarFunctionExpr) *SQLScalarFunctionExpr
}

type normalizingScalarFunc interface {
	normalize(*SQLScalarFunctionExpr) SQLExpr
}

func convertAllArgs(f *SQLScalarFunctionExpr, convType schema.SQLType, defaultValue SQLValue) *SQLScalarFunctionExpr {
	newExprs := convertAllExprs(f.Exprs, convType, defaultValue)
	return &SQLScalarFunctionExpr{
		f.Name,
		newExprs,
	}
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
	"cast":              &convertFunc{},
	"ceil":              singleArgFloatMathFunc(math.Ceil),
	"ceiling":           singleArgFloatMathFunc(math.Ceil),
	"char":              &charFunc{},
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
	"curtime":           &curtimeFunc{},
	"database":          &dbFunc{},
	"date":              &dateFunc{},
	"date_add":          &dateAddFunc{},
	"datediff":          &dateDiffFunc{},
	"date_sub":          &dateSubFunc{},
	"date_format":       &dateFormatFunc{},
	"day":               &dayOfMonthFunc{},
	"dayname":           &dayNameFunc{},
	"dayofmonth":        &dayOfMonthFunc{},
	"dayofweek":         &dayOfWeekFunc{},
	"dayofyear":         &dayOfYearFunc{},
	"degrees":           singleArgFloatMathFunc(func(f float64) float64 { return f * 180 / math.Pi }),
	"elt":               &eltFunc{},
	"exp":               singleArgFloatMathFunc(math.Exp),
	"extract":           &extractFunc{},
	"floor":             singleArgFloatMathFunc(math.Floor),
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
	"log":               singleArgFloatMathFunc(math.Log),
	"log2":              singleArgFloatMathFunc(math.Log2),
	"log10":             singleArgFloatMathFunc(math.Log10),
	"lower":             &lcaseFunc{},
	"lpad":              &lpadFunc{},
	"ltrim":             &ltrimFunc{},
	"makedate":          &makeDateFunc{},
	"microsecond":       &microsecondFunc{},
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
	"repeat":            &repeatFunc{},
	"replace":           &replaceFunc{},
	"right":             &rightFunc{},
	"round":             &roundFunc{},
	"rpad":              &rpadFunc{},
	"rtrim":             &rtrimFunc{},
	"schema":            &dbFunc{},
	"second":            &secondFunc{},
	"session_user":      &userFunc{},
	"sign":              &signFunc{},
	"sin":               singleArgFloatMathFunc(math.Sin),
	"sleep":             &sleepFunc{},
	"sqrt":              singleArgFloatMathFunc(math.Sqrt),
	"space":             &spaceFunc{},
	"str_to_date":       &strToDateFunc{},
	"subdate":           &subDateFunc{},
	"substr":            &substringFunc{},
	"substring":         &substringFunc{},
	"substring_index":   &substringIndexFunc{},
	"system_user":       &userFunc{},
	"tan":               singleArgFloatMathFunc(math.Tan),
	"timediff":          &timeDiffFunc{},
	"timestamp":         &timestampFunc{},
	"timestampadd":      &timestampAddFunc{},
	"timestampdiff":     &timestampDiffFunc{},
	"time_to_sec":       &timeToSecFunc{},
	"to_days":           &toDaysFunc{},
	"trim":              &trimFunc{},
	"truncate":          &truncateFunc{},
	"ucase":             &ucaseFunc{},
	"upper":             &ucaseFunc{},
	"user":              &userFunc{},
	"utc_date":          &utcDateFunc{},
	"utc_time":          &curtimeFunc{},
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
			return SQLNull, fmt.Errorf("%v '%v'", err.Error(), f.Name)
		}

		values, err := evaluateArgs(f.Exprs, ctx)
		if err != nil {
			return SQLNull, err
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

func (f *SQLScalarFunctionExpr) reconcile() *SQLScalarFunctionExpr {
	if sf, ok := scalarFuncMap[f.Name]; ok {
		if rsf, ok := sf.(reconcilingScalarFunc); ok {
			return rsf.reconcile(f)
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
		return sf.Type(f.Exprs)
	}

	return schema.SQLNone
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

func (_ singleArgFloatMathFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ singleArgFloatMathFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	switch f.Name {
	case "abs", "ceil", "exp", "floor", "ln", "log", "log10", "log2", "sqrt":
		return convertAllArgs(f, schema.SQLFloat, SQLNone)
	default:
		return f
	}
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

func (_ dualArgFloatMathFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ dualArgFloatMathFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	switch f.Name {
	case "mod":
		return convertAllArgs(f, schema.SQLFloat, SQLNone)
	default:
		return f
	}
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

func (_ multiArgFloatMathFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *connectionIdFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *dbFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *userFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *userFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type versionFunc struct{}

func (_ *versionFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLVarchar(ctx.Variables().GetString(variable.Version)), nil
}

func (_ *versionFunc) RequiresEvalCtx() bool {
	return true
}

func (_ *versionFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *addDateFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}
	return f
}

func (_ *addDateFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *asciiFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (_ *asciiFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type charFunc struct{}

func (_ *charFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {

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

func (_ *charFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *charFunc) Validate(exprCount int) error {
	if exprCount == 0 {
		return ErrIncorrectCount
	}

	return nil
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

func (_ *characterLengthFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (_ *characterLengthFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

func (_ *characterLengthFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNone)
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

func (_ *coalesceFunc) Type(exprs []SQLExpr) schema.SQLType {
	sorter := &schema.SQLTypesSorter{VarcharHighPriority: true}
	return preferentialTypeWithSorter(sorter, exprs...)
}

func (_ *coalesceFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, -1)
}

func (_ *coalesceFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return f
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

	var b bytes.Buffer
	for _, value := range values {
		b.WriteString(value.String())
	}

	v = SQLVarchar(b.String())
	err = nil
	return
}

func (_ *concatFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *concatFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *concatFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, -1)
}

func (_ *concatFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNone)
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

	var b bytes.Buffer
	var separator string = values[0].String()
	var trimValues []SQLValue = values[1:]
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

func (_ *concatWsFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if len(f.Exprs) >= 2 && f.Exprs[0] == SQLNull {
		return SQLNull
	}

	return f
}

func (_ *concatWsFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNone)
}

func (_ *concatWsFunc) Type(exprs []SQLExpr) schema.SQLType {
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
			t, ok = parseDateTime(typedV.String())
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
			t, ok = parseDateTime(typedV.String())
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
			d, ok := strToTime(typedV.String())
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

func (_ *convertFunc) Type(exprs []SQLExpr) schema.SQLType {
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
		return SQLNull, mysqlerrors.Defaultf(mysqlerrors.ER_DATA_OUT_OF_RANGE, "DOUBLE", fmt.Sprintf("'cot(%v)'", values[0].Float64()))
	}

	return SQLFloat(1 / tan), nil
}

func (_ *cotFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLFloat
}

func (_ *cotFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type currentDateFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_curdate
func (_ *currentDateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	now := time.Now().In(schema.DefaultLocale)
	t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, schema.DefaultLocale)
	return SQLDate{t}, nil

}

func (_ *currentDateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLDate
}

func (_ *currentDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type currentTimestampFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_now
func (_ *currentTimestampFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	value := time.Now().Round(time.Second).In(schema.DefaultLocale)
	return SQLTimestamp{value}, nil
}

func (_ *currentTimestampFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLTimestamp
}

func (_ *currentTimestampFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type curtimeFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_curtime
func (_ *curtimeFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLTimestamp{time.Now().In(schema.DefaultLocale)}, nil
}

func (_ *curtimeFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLTimestamp
}

func (_ *curtimeFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type dateFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_date
func (_ *dateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	t, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	return SQLDate{Time: t.Truncate(24 * time.Hour)}, nil
}

func (_ *dateFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *dateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLDate
}

func (_ *dateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type dateAddFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_date-add
func (_ *dateAddFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	_, ok := parseDateTime(values[0].String())
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

func (_ *dateAddFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *dateAddFunc) Type(exprs []SQLExpr) schema.SQLType {
	if exprs[0].Type() == schema.SQLTimestamp {
		return schema.SQLTimestamp
	}

	if exprs[0].Type() == schema.SQLDate {
		if unit, ok := exprs[2].(SQLValue); ok {
			switch unit.String() {
			case HOUR, MINUTE, SECOND:
				return schema.SQLTimestamp
			}
		}
	}

	return schema.SQLVarchar
}

func (_ *dateAddFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type dateDiffFunc struct{}

func (_ *dateDiffFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {

	var left, right time.Time
	var ok bool

	parseArgs := func(val SQLValue) (time.Time, bool) {
		var date time.Time

		date, ok = strToDateTime(val.String(), false)
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

func (_ *dateDiffFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}
	return f
}

func (_ *dateDiffFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (_ *dateDiffFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type dateSubFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_date-sub
func (_ *dateSubFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

func (_ *dateSubFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *dateSubFunc) Type(exprs []SQLExpr) schema.SQLType {
	return (&dateAddFunc{}).Type(exprs)
}

func (_ *dateSubFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type dateFormatFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_date-format
func (_ *dateFormatFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	date, ok := parseDateTime(values[0].String())
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

func (_ *dateFormatFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *dateFormatFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *dateFormatFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
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

func (_ *dayNameFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (_ *dayNameFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *dayOfMonthFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (_ *dayOfMonthFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *dayOfWeekFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (_ *dayOfWeekFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *dayOfYearFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (_ *dayOfYearFunc) Type(exprs []SQLExpr) schema.SQLType {
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

type eltFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_elt
func (_ *eltFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {

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

func (_ *eltFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *eltFunc) Validate(exprCount int) error {
	if exprCount <= 1 {
		return ErrIncorrectCount
	}

	return nil
}

func (_ *eltFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
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
		return SQLNull, fmt.Errorf("unit type '%v' is not supported", values[0].String())
	}
}

func (_ *extractFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLNone, schema.SQLTimestamp}
	defaults := []SQLValue{SQLNone, SQLNone}
	newExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		newExprs,
	}
}

func (_ *extractFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *fromDaysFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *fromDaysFunc) Type(exprs []SQLExpr) schema.SQLType {
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
	} else if convertTo == schema.SQLDate || convertTo == schema.SQLTimestamp {
		return values[greatestIdx], nil
	}

	return convertedVals[greatestIdx], nil
}

func (_ *greatestFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *greatestFunc) Type(exprs []SQLExpr) schema.SQLType {
	return preferentialType(exprs...)
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
	if values[0] == SQLNull {
		return SQLNull, nil
	}
	t, ok := parseTime(values[0].String())
	if !ok {
		return SQLInt(0), nil
	}

	return SQLInt(int(t.Hour())), nil
}

func (_ *hourFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *hourFunc) Type(exprs []SQLExpr) schema.SQLType {
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
		if typedV.Bool() {
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
		return SQLNull, fmt.Errorf("expression type '%v' is not supported", typedV)
	}
	return SQLNull, nil
}

func (_ *ifFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return f
}

func (_ *ifFunc) Type(exprs []SQLExpr) schema.SQLType {
	s := &schema.SQLTypesSorter{VarcharHighPriority: true}
	return preferentialTypeWithSorter(s, exprs[1:]...)
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

func (_ *ifnullFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return f
}

func (_ *ifnullFunc) Type(exprs []SQLExpr) schema.SQLType {
	s := &schema.SQLTypesSorter{VarcharHighPriority: true}
	return preferentialTypeWithSorter(s, exprs...)
}

func (_ *ifnullFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type insertFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_insert
func (_ *insertFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {

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

func (_ *insertFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *insertFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 4)
}

func (_ *insertFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

type instrFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_instr
func (_ *instrFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	locate := &locateFunc{}
	return locate.Evaluate([]SQLValue{values[1], values[0]}, ctx)
}

func (_ *instrFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (_ *instrFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type intervalFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_interval
func (_ *intervalFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

func (_ *intervalFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if f.Exprs[0].Type() == schema.SQLNull {
		return SQLInt(-1)
	}
	return f
}

func (_ *intervalFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLFloat, SQLNone)
}

func (_ *intervalFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt64
}

func (_ *intervalFunc) Validate(exprCount int) error {
	if exprCount < 2 {
		return ErrIncorrectVarCount
	}
	return nil
}

type isnullFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html#function_isnull
func (_ *isnullFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

func (_ *isnullFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return f
}

func (_ *isnullFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *lastDayFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *lastDayFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *lcaseFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNone)
}

func (_ *lcaseFunc) Type(exprs []SQLExpr) schema.SQLType {
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
	} else if convertTo == schema.SQLDate || convertTo == schema.SQLTimestamp {
		return values[leastIdx], nil
	}

	return convertedVals[leastIdx], nil
}

func (_ *leastFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *leastFunc) Type(exprs []SQLExpr) schema.SQLType {
	return preferentialType(exprs...)
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
	substring, err := NewSQLScalarFunctionExpr("substring", []SQLExpr{values[0], SQLInt(1), values[1]})
	if err != nil {
		return SQLNull, err
	}
	return substring.Evaluate(ctx)
}

func (_ *leftFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLInt}
	defaults := []SQLValue{SQLNone, SQLNone}
	newExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		newExprs,
	}
}

func (_ *leftFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *lengthFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNone)
}

func (_ *lengthFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *locateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (_ *locateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2, 3)
}

type lpadFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_lpad
func (_ *lpadFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return handlePadding(values, true)
}

func (_ *lpadFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *lpadFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

func (_ *lpadFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLInt, schema.SQLVarchar}
	defaults := []SQLValue{SQLNull, SQLNull, SQLNull}
	newExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		newExprs,
	}
}

func (_ *lpadFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
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

func (_ *ltrimFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *ltrimFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

func (_ *ltrimFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNull)
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

func (_ *makeDateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLDate
}

func (_ *makeDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type microsecondFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_microsecond
func (_ *microsecondFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {

	arg := values[0]

	if arg == SQLNull {
		return SQLNull, nil
	}

	str := arg.String()
	if str == "" {
		return SQLNull, nil
	}

	t, ok := parseTime(str)
	if !ok {
		return SQLInt(0), nil
	}

	return SQLInt(int(t.Nanosecond() / 1000)), nil
}

func (_ *microsecondFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *microsecondFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (_ *microsecondFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type midFunc struct {
	wrapped substringFunc
}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_mid
func (m *midFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return m.wrapped.Evaluate(values, ctx)
}

func (_ *midFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLNone, schema.SQLInt}
	defaults := []SQLValue{SQLNone, SQLNone, SQLNone}
	newExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		newExprs,
	}
}

func (m *midFunc) Type(exprs []SQLExpr) schema.SQLType {
	return m.wrapped.Type(exprs)
}

func (_ *midFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type minuteFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_minute
func (_ *minuteFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if values[0] == SQLNull {
		return SQLNull, nil
	}

	t, ok := parseTime(values[0].String())
	if !ok {
		return SQLInt(0), nil
	}

	return SQLInt(int(t.Minute())), nil
}

func (_ *minuteFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *minuteFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *monthFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (_ *monthFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *monthNameFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (_ *monthNameFunc) Type(exprs []SQLExpr) schema.SQLType {
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
		return SQLNull, err
	}
	if NewSQLBool(result) == SQLTrue {
		return SQLInt(1), nil
	}
	return SQLInt(0), nil
}

func (_ *notFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *nullifFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return f
}

func (_ *nullifFunc) Type(exprs []SQLExpr) schema.SQLType {
	return exprs[0].Type()
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

func (_ *quarterFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (_ *quarterFunc) Type(exprs []SQLExpr) schema.SQLType {
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

	v0 := values[0].Float64()
	v1 := values[1].Float64()

	n := math.Pow(v0, v1)
	zeroBaseExpNeg := v0 == 0 && v1 < 0
	if math.IsNaN(n) || zeroBaseExpNeg {
		return SQLNull, mysqlerrors.Defaultf(mysqlerrors.ER_DATA_OUT_OF_RANGE, "DOUBLE", fmt.Sprintf("pow(%v,%v)", values[0].Float64(), values[1].Float64()))
	}

	return SQLFloat(math.Pow(v0, v1)), nil
}

func (_ *powFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *powFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLFloat, SQLNone)
}

func (_ *powFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLFloat
}

func (_ *powFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type repeatFunc struct{}

// http://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_repeat
func (_ *repeatFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (v SQLValue, err error) {
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

func (_ *repeatFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *repeatFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *repeatFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

func (_ *repeatFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLNumeric}
	defaults := []SQLValue{SQLNone, SQLNone}
	newExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		newExprs,
	}
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

func (_ *replaceFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *replaceFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

type rightFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_right
func (_ *rightFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
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

func (_ *rightFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLInt}
	defaults := []SQLValue{SQLNone, SQLNone}
	newExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		newExprs,
	}
}

func (_ *rightFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *rtrimFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *rtrimFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

func (_ *rtrimFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNull)
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

func (_ *roundFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLFloat, SQLNone)
}

func (_ *roundFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLFloat
}

func (_ *roundFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

type rpadFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_rpad
func (_ *rpadFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return handlePadding(values, false)
}

func (_ *rpadFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *rpadFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

func (_ *rpadFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLInt, schema.SQLVarchar}
	defaults := []SQLValue{SQLNull, SQLNull, SQLNull}
	newExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		newExprs,
	}
}

func (_ *rpadFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}
	return f
}

type secondFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_second
func (_ *secondFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if values[0] == SQLNull {
		return SQLNull, nil
	}

	t, ok := parseTime(values[0].String())
	if !ok {
		return SQLInt(0), nil
	}

	return SQLInt(int(t.Second())), nil
}

func (_ *secondFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *secondFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *signFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (_ *signFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type sleepFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/miscellaneous-functions.html#function_sleep
func (_ *sleepFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {

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

func (_ *sleepFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (_ *sleepFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

func (_ *sleepFunc) RequiresEvalCtx() bool {
	return true
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

func (_ *spaceFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *strToDateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLTimestamp
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

func (_ *subDateFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *subDateFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *substringFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *substringFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLVarchar, schema.SQLInt, schema.SQLInt}
	defaults := []SQLValue{SQLNone, SQLNone, SQLNone}
	newExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		newExprs,
	}
}

func (_ *substringFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *substringFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2, 3)
}

type substringIndexFunc struct{}

// https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_substring-index
func (_ *substringIndexFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {

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

func (_ *substringIndexFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *substringIndexFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 3)
}

func (_ *substringIndexFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
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

type timeDiffFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_timediff
func (_ *timeDiffFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	expr1, ok := parseTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	expr2, ok := parseTime(values[1].String())
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

func (_ *timeDiffFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *timeDiffFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *timeDiffFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

type timestampFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_timestamp
func (_ *timestampFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	t, ok := parseDateTime(values[0].String())
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

func (_ *timestampFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *timestampFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLTimestamp
}

func (_ *timestampFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
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
		duration, _ := time.ParseDuration(v.String() + "us")
		return SQLTimestamp{Time: t.Add(duration)}, nil
	default:
		return SQLNull, fmt.Errorf("cannot add '%v' to timestamp", values[0])
	}
	return SQLNull, nil
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
		return SQLNull, fmt.Errorf("cannot add '%v' to timestamp", values[0])
	}
}

func (t *timestampDiffFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *timeToSecFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *timeToSecFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *trimFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *trimFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 3)
}

func (_ *trimFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNull)
}

type toDaysFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_to-days
func (_ *toDaysFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	if hasNullValue(values...) {
		return SQLNull, nil
	}

	date, ok := parseDateTime(values[0].String())
	if !ok {
		return SQLNull, nil
	}

	start, _ := time.ParseInLocation(shortTimeFormat, "0000-01-01", schema.DefaultLocale)
	target, maxGoDurationHours := 1.0, int64(2562024)
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, schema.DefaultLocale)
	for date.Sub(start).Hours() > 24 {
		for date.Sub(start).Hours() > float64(maxGoDurationHours) {
			date = date.Add(time.Duration(-maxGoDurationHours) * time.Hour)
			target += float64(106751)
		}
		date, target = date.AddDate(0, 0, -1), target+1
	}

	return SQLFloat(target), nil
}

func (_ *toDaysFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
}

func (_ *toDaysFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLFloat
}

func (_ *toDaysFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
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

func (_ *truncateFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLFloat, schema.SQLNone}
	defaults := []SQLValue{SQLNone, SQLNone}
	newExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		newExprs,
	}
}

func (_ *truncateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLFloat
}

func (_ *truncateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 2)
}

func (_ *truncateFunc) normalize(f *SQLScalarFunctionExpr) SQLExpr {
	if hasNullExpr(f.Exprs...) {
		return SQLNull
	}

	return f
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

func (_ *ucaseFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLVarchar, SQLNone)
}

func (_ *ucaseFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLVarchar
}

func (_ *ucaseFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1)
}

type utcDateFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_utc-date
func (_ *utcDateFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	now := time.Now().In(time.UTC)
	t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return SQLDate{t}, nil
}

func (_ *utcDateFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLDate
}

func (_ *utcDateFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 0)
}

type utcTimestampFunc struct{}

// https://dev.mysql.com/doc/refman/5.5/en/date-and-time-functions.html#function_utc-timestamp
func (_ *utcTimestampFunc) Evaluate(values []SQLValue, ctx *EvalCtx) (SQLValue, error) {
	return SQLTimestamp{time.Now().UTC().Round(time.Second).In(time.UTC)}, nil
}

func (_ *utcTimestampFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *weekFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	argTypes := []schema.SQLType{schema.SQLDate, schema.SQLInt}
	defaults := []SQLValue{SQLNull, SQLNone}
	newExprs := convertExprs(f.Exprs, argTypes, defaults)
	return &SQLScalarFunctionExpr{
		f.Name,
		newExprs,
	}
}

func (_ *weekFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *weekdayFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (_ *weekdayFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *weekOfYearFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *yearFunc) reconcile(f *SQLScalarFunctionExpr) *SQLScalarFunctionExpr {
	return convertAllArgs(f, schema.SQLDate, SQLNull)
}

func (_ *yearFunc) Type(exprs []SQLExpr) schema.SQLType {
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

func (_ *yearWeekFunc) Type(exprs []SQLExpr) schema.SQLType {
	return schema.SQLInt
}

func (_ *yearWeekFunc) Validate(exprCount int) error {
	return ensureArgCount(exprCount, 1, 2)
}

// Helper functions

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
		case DAY_MICROSECOND:
			val = args[0]*day*hour*minute*second + args[1]*hour*minute*second + args[2]*minute*second + args[3]*second + args[4]
		default:
			return unit, 0, fmt.Errorf("invalid argument length")
		}
	case 4:
		switch unit {
		case DAY_MICROSECOND, HOUR_MICROSECOND:
			val = args[0]*hour*minute*second + args[1]*minute*second + args[2]*second + args[3]
		case DAY_SECOND:
			val = args[0]*day*hour*minute + args[1]*hour*minute + args[2]*minute + args[3]
		default:
			return unit, 0, fmt.Errorf("invalid argument length")
		}
	case 3:
		switch unit {
		case DAY_MICROSECOND, HOUR_MICROSECOND, MINUTE_MICROSECOND:
			val = args[0]*minute*second + args[1]*second + args[2]
		case DAY_SECOND, HOUR_SECOND:
			val = args[0]*hour*minute + args[1]*minute + args[2]
		case DAY_MINUTE:
			val = args[0]*day*hour + args[1]*hour + args[2]
		default:
			return unit, 0, fmt.Errorf("invalid argument length")
		}
	case 2:
		switch unit {
		case DAY_MICROSECOND, HOUR_MICROSECOND, MINUTE_MICROSECOND, SECOND_MICROSECOND:
			val = args[0]*second + args[1]
		case DAY_SECOND, HOUR_SECOND, MINUTE_SECOND:
			val = args[0]*minute + args[1]
		case DAY_MINUTE, HOUR_MINUTE:
			val = args[0]*hour + args[1]
		case DAY_HOUR:
			val = args[0]*day + args[1]
		case YEAR_MONTH:
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
	if unit != MICROSECOND && strings.HasSuffix(unit, MICROSECOND) {
		curr = curr + strings.Repeat("0", 6-len(curr))
	}
	c, _ := strconv.Atoi(curr)
	args = append(args, c)
	return args, neg
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
	case DAY:
		return interval * 24 * 60 * 60 * 1000, nil
	case HOUR:
		return interval * 60 * 60 * 1000, nil
	case MINUTE:
		return interval * 60 * 1000, nil
	case SECOND:
		return interval * 1000, nil
	case MICROSECOND:
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

func parseDateTime(s string) (time.Time, bool) {
	return strToDateTime(s, false)
}

func parseTime(s string) (time.Time, bool) {
	if len(s) >= 12 {
		// probably a datetime
		dt, ok := strToDateTime(s, true)
		if ok {
			return dt, true
		}
	}

	// the result will be 0 if parsing failed, so we don't care about the result.
	dur, ok := strToTime(s)

	return time.Date(0, 1, 1, 0, 0, 0, 0, schema.DefaultLocale).Add(dur), ok
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
