package evaluator

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
)

// NullDate is an internal representation for the MySQL
// date 0000-00-00, which cannot be represented as all 0's.
var NullDate = time.Date(-1, 1, 0, 0, 0, 0, 0, schema.DefaultLocale)

// ConvertTo takes a SQLValue v and an evalType, that determines what
// type to convert the passed SQLValue.
func ConvertTo(v SQLValue, evalType EvalType) SQLValue {
	switch evalType {
	case EvalArrNumeric:
		return v.SQLFloat()
	case EvalBoolean:
		return v.SQLBool()
	case EvalDecimal128:
		return v.SQLDecimal128()
	case EvalDouble:
		return v.SQLFloat()
	case EvalInt32:
		return v.SQLInt()
	case EvalInt64:
		return v.SQLInt()
	case EvalNone:
		return v
	case EvalNull:
		return SQLNull
	case EvalObjectID:
		return v.SQLVarchar()
	case EvalString:
		return v.SQLVarchar()
	case EvalDatetime:
		return v.SQLTimestamp()
	case EvalUUID:
		return v.SQLVarchar()
	// Types not corresponding to MongoDB types.
	case EvalDate:
		return v.SQLDate()
	case EvalUint64:
		return v.SQLUint()
	default:
		panic(fmt.Sprintf("EvalType %x should never be seen as a conversion target", evalType))
	}
}

// The following are functions for conversion to go types.
//
// These all return 0 values for failed
// conversion (e.g., NULL).

// Bool converts a SQLValue to a bool.
func Bool(v SQLValue) bool {
	converted, ok := v.SQLBool().(SQLBool)
	if !ok {
		return false
	}
	return converted != 0.0
}

// Decimal converts a SQLValue to a decimal128.
func Decimal(v SQLValue) decimal.Decimal {
	converted, ok := v.SQLDecimal128().(SQLDecimal128)
	if !ok {
		return decimal.Zero
	}
	return decimal.Decimal(converted)
}

// Float64 converts a SQLValue to a float64.
func Float64(v SQLValue) float64 {
	converted, ok := v.SQLFloat().(SQLFloat)
	if !ok {
		return 0.0
	}
	return float64(converted)
}

// Int64 converts a SQLValue to an int64.
func Int64(v SQLValue) int64 {
	converted, ok := v.SQLInt().(SQLInt64)
	if !ok {
		return 0
	}
	return int64(converted)
}

// Uint64 converts a SQLValue to a uint64.
func Uint64(v SQLValue) uint64 {
	converted, ok := v.SQLUint().(SQLUint64)
	if !ok {
		return 0
	}
	return uint64(converted)
}

// The following are functions for Bool Conversion.

// ConvertTo converts the SQLBool receiver, s, to the specified EvalType.
func (s SQLBool) ConvertTo(evalType EvalType) SQLValue {
	return ConvertTo(s, evalType)
}

// SQLBool converts the SQLBool receiver, s, to a SQLBool.
func (s SQLBool) SQLBool() SQLValue {
	return s
}

// SQLDate converts the SQLBool receiver, s, to a SQLDate.
func (s SQLBool) SQLDate() SQLValue {
	if s == SQLTrue {
		return SQLNull
	}
	return SQLDate{Time: NullDate}
}

// SQLDecimal128 converts the SQLBool receiver, s, to a SQLDecimal128.
func (s SQLBool) SQLDecimal128() SQLValue {
	if s == SQLTrue {
		return SQLDecimal128(decimal.New(1, 0))
	}
	return SQLDecimal128(decimal.Zero)
}

// SQLFloat converts the SQLBool receiver, s, to a SQLFloat.
func (s SQLBool) SQLFloat() SQLValue {
	if s == SQLTrue {
		return SQLFloat(1.0)
	}
	return SQLFloat(0.0)
}

// SQLInt converts the SQLBool receiver, s, to a SQLInt.
func (s SQLBool) SQLInt() SQLValue {
	if s == SQLTrue {
		return SQLInt64(1)
	}
	return SQLInt64(0)
}

// SQLTimestamp converts the SQLBool receiver, s, to a SQLTimestamp.
func (s SQLBool) SQLTimestamp() SQLValue {
	if s == SQLTrue {
		return SQLNull
	}
	return SQLTimestamp{Time: NullDate}
}

// SQLUint converts the SQLBool receiver, s, to a SQLUint.
func (s SQLBool) SQLUint() SQLValue {
	if s == SQLTrue {
		return SQLUint64(1)
	}
	return SQLUint64(0)
}

// SQLVarchar converts the SQLBool receiver, s, to a SQLVarchar.
func (s SQLBool) SQLVarchar() SQLValue {
	if s == SQLTrue {
		return SQLVarchar("1")
	}
	return SQLVarchar("0")
}

// The following are functions for Date Conversion.

// ConvertTo converts the SQLDate receiver, s, to the specified EvalType.
func (s SQLDate) ConvertTo(evalType EvalType) SQLValue {
	return ConvertTo(s, evalType)
}

// SQLBool converts the SQLDate receiver, s, to a SQLBool.
func (s SQLDate) SQLBool() SQLValue {
	t := s.Time
	if t == NullDate {
		return SQLFalse
	}
	return SQLTrue
}

// SQLDecimal128 converts the SQLDate receiver, s, to a SQLDecimal128.
func (s SQLDate) SQLDecimal128() SQLValue {
	if ret, ok := s.SQLFloat().(SQLFloat); ok {
		return SQLDecimal128(decimal.NewFromFloat(float64(ret)))
	}
	// unreachable.
	return SQLNull
}

// SQLDate converts the SQLDate receiver, s, to a SQLDate.
func (s SQLDate) SQLDate() SQLValue {
	return s
}

// SQLFloat converts the SQLDate receiver, s, to a SQLFloat.
func (s SQLDate) SQLFloat() SQLValue {
	t := s.Time
	return (SQLFloat(t.Day()) +
		SQLFloat(t.Month())*1e2 +
		SQLFloat(t.Year())*1e4)
}

// SQLInt converts the SQLDate receiver, s, to a SQLInt.
func (s SQLDate) SQLInt() SQLValue {
	t := s.Time
	return (SQLInt64(t.Day()) +
		SQLInt64(t.Month())*1e2 +
		SQLInt64(t.Year())*1e4)
}

// SQLTimestamp converts the SQLDate receiver, s, to a SQLTimestamp.
func (s SQLDate) SQLTimestamp() SQLValue {
	t := s.Time
	return SQLTimestamp{
		Time: t,
	}
}

// SQLUint converts the SQLDate receiver, s, to a SQLUint.
func (s SQLDate) SQLUint() SQLValue {
	t := s.Time
	return (SQLUint64(t.Day()) +
		SQLUint64(t.Month())*1e2 +
		SQLUint64(t.Year())*1e4)
}

// SQLVarchar converts the SQLDate receiver, s, to a SQLVarchar.
func (s SQLDate) SQLVarchar() SQLValue {
	return SQLVarchar(s.String())
}

// The following are functions for Decimal Conversion.

// ConvertTo converts the SQLDecimal128 receiver, s, to the specified EvalType.
func (s SQLDecimal128) ConvertTo(evalType EvalType) SQLValue {
	return ConvertTo(s, evalType)
}

// SQLBool converts the SQLDecimal128 receiver, s, to a SQLBool.
func (s SQLDecimal128) SQLBool() SQLValue {
	if s == SQLDecimal128(decimal.Zero) {
		return SQLFalse
	}
	return SQLTrue
}

// SQLDate converts the SQLDecimal128 receiver, s, to a SQLDate.
func (s SQLDecimal128) SQLDate() SQLValue {
	if s == SQLDecimal128(decimal.Zero) {
		return SQLDate{Time: NullDate}
	}

	t, _, ok := parseDateTime(s.String())
	if !ok {
		return SQLNull
	}
	t = t.In(schema.DefaultLocale)
	return SQLDate{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale)}
}

// SQLDecimal128 converts the SQLDecimal128 receiver, s, to a SQLDecimal128.
func (s SQLDecimal128) SQLDecimal128() SQLValue {
	return s
}

// SQLFloat converts the SQLDecimal128 receiver, s, to a SQLFloat.
func (s SQLDecimal128) SQLFloat() SQLValue {
	// Second return value tells us if this is exact, we don't care.
	f, _ := decimal.Decimal(s).Float64()
	return SQLFloat(f)
}

// SQLInt converts the SQLDecimal128 receiver, s, to a SQLInt.
func (s SQLDecimal128) SQLInt() SQLValue {
	// Do not care if this is exact.
	f, _ := decimal.Decimal(s).Float64()
	return SQLInt64(round(f))
}

// SQLTimestamp converts the SQLDecimal128 receiver, s, to a SQLTimestamp.
func (s SQLDecimal128) SQLTimestamp() SQLValue {
	if s == SQLDecimal128(decimal.Zero) {
		return SQLTimestamp{Time: NullDate}
	}

	t, _, ok := parseDateTime(s.String())
	if !ok {
		return SQLNull
	}
	t = t.In(schema.DefaultLocale)
	return SQLTimestamp{Time: t}
}

// SQLUint converts the SQLDecimal128 receiver, s, to a SQLUint.
func (s SQLDecimal128) SQLUint() SQLValue {
	// Do not care if this is exact.
	f, _ := decimal.Decimal(s).Float64()
	return SQLUint64(round(f))
}

// SQLVarchar converts the SQLDecimal128 receiver, s, to a SQLVarchar.
func (s SQLDecimal128) SQLVarchar() SQLValue {
	return SQLVarchar(s.String())
}

// The following are functions for Float Conversion.

// ConvertTo converts the SQLFloat receiver, s, to the specified EvalType.
func (s SQLFloat) ConvertTo(evalType EvalType) SQLValue {
	return ConvertTo(s, evalType)
}

// SQLBool converts the SQLFloat receiver, s, to a SQLBool.
func (s SQLFloat) SQLBool() SQLValue {
	if s == 0 {
		return SQLFalse
	}
	return SQLTrue
}

// SQLDate converts the SQLFloat receiver, s, to a SQLDate.
func (s SQLFloat) SQLDate() SQLValue {
	if s == 0.0 {
		return SQLDate{Time: NullDate}
	}

	t, _, ok := parseDateTime(s.String())
	if !ok {
		return SQLNull
	}
	t = t.In(schema.DefaultLocale)
	return SQLDate{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale)}
}

// SQLDecimal128 converts the SQLFloat receiver, s, to a SQLDecimal128.
func (s SQLFloat) SQLDecimal128() SQLValue {
	return SQLDecimal128(decimal.NewFromFloat(float64(s)))
}

// SQLFloat converts the SQLFloat receiver, s, to a SQLFloat.
func (s SQLFloat) SQLFloat() SQLValue {
	return s
}

// SQLInt converts the SQLFloat receiver, s, to a SQLInt.
func (s SQLFloat) SQLInt() SQLValue {
	return SQLInt64(round(float64(s)))
}

// SQLTimestamp converts the SQLFloat receiver, s, to a SQLTimestamp.
func (s SQLFloat) SQLTimestamp() SQLValue {
	if s == 0.0 {
		return SQLTimestamp{Time: NullDate}
	}

	t, _, ok := parseDateTime(s.String())
	if !ok {
		return SQLNull
	}
	t = t.In(schema.DefaultLocale)
	return SQLTimestamp{Time: t}
}

// SQLUint converts the SQLFloat receiver, s, to a SQLUint.
func (s SQLFloat) SQLUint() SQLValue {
	return SQLUint64(round(float64(s)))
}

// SQLVarchar converts the SQLFloat receiver, s, to a SQLVarchar.
func (s SQLFloat) SQLVarchar() SQLValue {
	return SQLVarchar(s.String())
}

// The following are functions for Int Conversion.

// ConvertTo converts the SQLInt receiver, s, to the specified EvalType.
func (s SQLInt64) ConvertTo(evalType EvalType) SQLValue {
	return ConvertTo(s, evalType)
}

// SQLBool converts the SQLInt receiver, s, to a SQLBool.
func (s SQLInt64) SQLBool() SQLValue {
	if s == 0 {
		return SQLFalse
	}
	return SQLTrue
}

// SQLDate converts the SQLInt receiver, s, to a SQLDate.
func (s SQLInt64) SQLDate() SQLValue {
	if s == 0 {
		return SQLDate{Time: NullDate}
	}

	t, _, ok := parseDateTime(s.String())
	if !ok {
		return SQLNull
	}
	t = t.In(schema.DefaultLocale)
	return SQLDate{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale)}
}

// SQLDecimal128 converts the SQLInt receiver, s, to a SQLDecimal128.
func (s SQLInt64) SQLDecimal128() SQLValue {
	return SQLDecimal128(decimal.New(int64(s), 0))
}

// SQLFloat converts the SQLInt receiver, s, to a SQLFloat.
func (s SQLInt64) SQLFloat() SQLValue {
	return SQLFloat(s)
}

// SQLInt converts the SQLInt receiver, s, to a SQLInt.
func (s SQLInt64) SQLInt() SQLValue {
	return s
}

// SQLTimestamp converts the SQLInt receiver, s, to a SQLTimestamp.
func (s SQLInt64) SQLTimestamp() SQLValue {
	if s == 0 {
		return SQLTimestamp{Time: NullDate}
	}

	t, _, ok := parseDateTime(s.String())
	if !ok {
		return SQLNull
	}
	t = t.In(schema.DefaultLocale)
	return SQLTimestamp{Time: t}
}

// SQLUint converts the SQLUint receiver, s, to a SQLUint.
func (s SQLInt64) SQLUint() SQLValue {
	return SQLUint64(s)
}

// SQLVarchar converts the SQLInt receiver, s, to a SQLVarchar.
func (s SQLInt64) SQLVarchar() SQLValue {
	return SQLVarchar(s.String())
}

// The following are functions for SQLNoValue Conversion.

// ConvertTo converts the SQLNoValue receiver to the specified EvalType.
func (SQLNoValue) ConvertTo(evalType EvalType) SQLValue {
	return SQLNull
}

// SQLBool converts the SQLNoValue receiver to a SQLBool.
func (SQLNoValue) SQLBool() SQLValue {
	return SQLFalse
}

// SQLDate converts the SQLNoValue receiver to a SQLDate.
func (SQLNoValue) SQLDate() SQLValue {
	return SQLDate{Time: NullDate}
}

// SQLDecimal128 converts the SQLNoValue receiver to a SQLDecimal128.
func (SQLNoValue) SQLDecimal128() SQLValue {
	return SQLDecimal128(decimal.Zero)
}

// SQLFloat converts the SQLNoValue receiver to a SQLFloat.
func (SQLNoValue) SQLFloat() SQLValue {
	return SQLFloat(0.0)
}

// SQLInt converts the SQLNoValue receiver to a SQLInt.
func (SQLNoValue) SQLInt() SQLValue {
	return SQLInt64(0)
}

// SQLTimestamp converts the SQLNoValue receiver to a SQLTimestamp.
func (SQLNoValue) SQLTimestamp() SQLValue {
	return SQLTimestamp{Time: NullDate}
}

// SQLUint converts the SQLNoValue receiver to a SQLUint.
func (SQLNoValue) SQLUint() SQLValue {
	return SQLUint64(0)
}

// SQLVarchar converts the SQLNoValue receiver to a SQLVarchar.
func (SQLNoValue) SQLVarchar() SQLValue {
	return SQLVarchar(schema.SQLNone)
}

// The following are functions for SQLNull Conversion.

// ConvertTo converts the SQLNullValue receiver to the specified EvalType.
func (SQLNullValue) ConvertTo(evalType EvalType) SQLValue {
	return SQLNull
}

// SQLBool converts the SQLNullValue receiver to a SQLBool.
func (SQLNullValue) SQLBool() SQLValue {
	return SQLNull
}

// SQLDate converts the SQLNullValue receiver to a SQLDate.
func (SQLNullValue) SQLDate() SQLValue {
	return SQLNull
}

// SQLDecimal128 converts the SQLNullValue receiver to a SQLDecimal128.
func (SQLNullValue) SQLDecimal128() SQLValue {
	return SQLNull
}

// SQLFloat converts the SQLNullValue receiver to a SQLFloat.
func (SQLNullValue) SQLFloat() SQLValue {
	return SQLNull
}

// SQLInt converts the SQLNullValue receiver to a SQLInt.
func (SQLNullValue) SQLInt() SQLValue {
	return SQLNull
}

// SQLTimestamp converts the SQLNullValue receiver to a SQLTimestamp.
func (SQLNullValue) SQLTimestamp() SQLValue {
	return SQLNull
}

// SQLUint converts the SQLNullValue receiver to a SQLUint.
func (SQLNullValue) SQLUint() SQLValue {
	return SQLNull
}

// SQLVarchar converts the SQLNullValue receiver to a SQLVarchar.
func (SQLNullValue) SQLVarchar() SQLValue {
	return SQLNull
}

// The following are functions for Timestamp Conversion.

// ConvertTo converts the SQLTimestamp receiver, s, to the specified EvalType.
func (s SQLTimestamp) ConvertTo(evalType EvalType) SQLValue {
	return ConvertTo(s, evalType)
}

// SQLBool converts the SQLTimestamp receiver, s, to a SQLBool.
func (s SQLTimestamp) SQLBool() SQLValue {
	t := s.Time
	if t == NullDate {
		return SQLFalse
	}
	return SQLTrue
}

// SQLDate converts the SQLTimestamp receiver, s, to a SQLDate.
func (s SQLTimestamp) SQLDate() SQLValue {
	t := s.Time
	return SQLDate{
		Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	}
}

// SQLDecimal128 converts the SQLTimestamp receiver, s, to a SQLDecimal128.
func (s SQLTimestamp) SQLDecimal128() SQLValue {
	if ret, ok := s.SQLFloat().(SQLFloat); ok {
		return SQLDecimal128(decimal.NewFromFloat(float64(ret)))
	}
	// unreachable.
	return SQLNull
}

// SQLFloat converts the SQLTimestamp receiver, s, to a SQLFloat.
func (s SQLTimestamp) SQLFloat() SQLValue {
	t := s.Time
	return (SQLFloat(t.Second()) +
		SQLFloat(t.Minute())*1e2 +
		SQLFloat(t.Hour())*1e4 +
		SQLFloat(t.Day())*1e6 +
		SQLFloat(t.Month())*1e8 +
		SQLFloat(t.Year())*1e10 +
		SQLFloat(t.Nanosecond())/1e9)
}

// SQLInt converts the SQLTimestamp receiver, s, to a SQLInt.
func (s SQLTimestamp) SQLInt() SQLValue {
	t := s.Time
	return (SQLInt64(t.Second()) +
		SQLInt64(t.Minute())*1e2 +
		SQLInt64(t.Hour())*1e4 +
		SQLInt64(t.Day())*1e6 +
		SQLInt64(t.Month())*1e8 +
		SQLInt64(t.Year())*1e10)
}

// SQLTimestamp converts the SQLTimestamp receiver, s, to a SQLTimestamp.
func (s SQLTimestamp) SQLTimestamp() SQLValue {
	return s
}

// SQLUint converts the SQLTimestamp receiver, s, to a SQLUint.
func (s SQLTimestamp) SQLUint() SQLValue {
	t := s.Time
	return (SQLUint64(t.Second()) +
		SQLUint64(t.Minute())*1e2 +
		SQLUint64(t.Hour())*1e4 +
		SQLUint64(t.Day())*1e6 +
		SQLUint64(t.Month())*1e8 +
		SQLUint64(t.Year())*1e10)
}

// SQLVarchar converts the SQLTimestamp receiver, s, to a SQLVarchar.
func (s SQLTimestamp) SQLVarchar() SQLValue {
	return SQLVarchar(s.String())
}

// The following are functions for Uint Conversion.

// ConvertTo converts the SQLUint receiver, s, to the specified EvalType.
func (s SQLUint64) ConvertTo(evalType EvalType) SQLValue {
	return ConvertTo(s, evalType)
}

// SQLBool converts the SQLUint receiver, s, to a SQLBool.
func (s SQLUint64) SQLBool() SQLValue {
	if s == 0 {
		return SQLFalse
	}
	return SQLTrue
}

// SQLDate converts the SQLUint receiver, s, to a SQLDate.
func (s SQLUint64) SQLDate() SQLValue {
	if s == 0 {
		return SQLDate{Time: NullDate}
	}

	t, _, ok := parseDateTime(s.String())
	if !ok {
		return SQLNull
	}
	t = t.In(schema.DefaultLocale)
	return SQLDate{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale)}
}

// SQLDecimal128 converts the SQLUint receiver, s, to a SQLDecimal128.
func (s SQLUint64) SQLDecimal128() SQLValue {
	return SQLDecimal128(decimal.New(int64(s), 0))
}

// SQLFloat converts the SQLUint receiver, s, to a SQLFloat.
func (s SQLUint64) SQLFloat() SQLValue {
	return SQLFloat(s)
}

// SQLInt converts the SQLUint receiver, s, to a SQLInt.
func (s SQLUint64) SQLInt() SQLValue {
	return s
}

// SQLTimestamp converts the SQLUint receiver, s, to a SQLTimestamp.
func (s SQLUint64) SQLTimestamp() SQLValue {
	if s == 0 {
		return SQLTimestamp{Time: NullDate}
	}

	t, _, ok := parseDateTime(s.String())
	if !ok {
		return SQLNull
	}
	t = t.In(schema.DefaultLocale)
	return SQLDate{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale)}
}

// SQLUint converts the SQLUint receiver, s, to a SQLUint.
func (s SQLUint64) SQLUint() SQLValue {
	return s
}

// SQLVarchar converts the SQLUint receiver, s, to a SQLVarchar.
func (s SQLUint64) SQLVarchar() SQLValue {
	return SQLVarchar(s.String())
}

// *SQLValues Conversion

// ConvertTo converts the *SQLValues receiver, s, to the specified EvalType.
func (s *SQLValues) ConvertTo(evalType EvalType) SQLValue {
	return ConvertTo(s.Values[0], evalType)
}

// SQLBool converts the *SQLValues receiver, s, to a SQLBool.
func (s *SQLValues) SQLBool() SQLValue {
	return s.Values[0].SQLBool()
}

// SQLDate converts the *SQLValues receiver, s, to a SQLDate.
func (s *SQLValues) SQLDate() SQLValue {
	return s.Values[0].SQLDate()
}

// SQLDecimal128 converts the *SQLValues receiver, s, to a SQLDecimal128.
func (s *SQLValues) SQLDecimal128() SQLValue {
	return s.Values[0].SQLDecimal128()
}

// SQLFloat converts the *SQLValues receiver, s, to a SQLFloat.
func (s *SQLValues) SQLFloat() SQLValue {
	return s.Values[0].SQLFloat()
}

// SQLInt converts the *SQLValues receiver, s, to a SQLInt.
func (s *SQLValues) SQLInt() SQLValue {
	return s.Values[0].SQLInt()
}

// SQLTimestamp converts the *SQLValues receiver, s, to a SQLTimestamp.
func (s *SQLValues) SQLTimestamp() SQLValue {
	return s.Values[0].SQLTimestamp()
}

// SQLUint converts the *SQLValues receiver, s, to a SQLUint.
func (s *SQLValues) SQLUint() SQLValue {
	return s.Values[0].SQLUint()
}

// SQLVarchar converts the *SQLValues receiver, s, to a SQLVarchar.
func (s *SQLValues) SQLVarchar() SQLValue {
	values := make([]string, len(s.Values))
	for i, n := range s.Values {
		if n, ok := n.SQLVarchar().(SQLVarchar); ok {
			values[i] = string(n)
		}
		values[i] = "NULL"
	}
	return SQLVarchar(strings.Join(values, ", "))
}

// The following are functions for Varchar Conversion.

// ConvertTo converts the SQLVarchar receiver, s, to the specified EvalType.
func (s SQLVarchar) ConvertTo(evalType EvalType) SQLValue {
	return ConvertTo(s, evalType)
}

// SQLBool converts the SQLVarchar receiver, s, to a SQLBool.
func (s SQLVarchar) SQLBool() SQLValue {
	// Note that we convert to Bool by converting to Int then to Bool,
	// these are the specified semantics of mysql.
	return s.SQLInt().SQLBool()
}

// SQLDate converts the SQLVarchar receiver, s, to a SQLDate.
func (s SQLVarchar) SQLDate() SQLValue {
	t, _, ok := parseDateTime(s.String())
	if !ok {
		return SQLDate{Time: NullDate}
	}
	t = t.In(schema.DefaultLocale)
	return SQLDate{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale)}
}

// SQLDecimal128 converts the SQLVarchar receiver, s, to a SQLDecimal128.
func (s SQLVarchar) SQLDecimal128() SQLValue {
	cleaned := MySQLCleanScientificNotationString(string(s))
	out, err := decimal.NewFromString(cleaned)
	if err != nil {
		return SQLDecimal128(decimal.Zero)
	}
	return SQLDecimal128(out)
}

// SQLFloat converts the SQLVarchar receiver, s, to a SQLFloat.
func (s SQLVarchar) SQLFloat() SQLValue {
	cleaned := MySQLCleanNumericString(string(s))
	out, _ := strconv.ParseFloat(cleaned, 64)
	return SQLFloat(out)
}

// SQLInt converts the SQLVarchar receiver, s, to a SQLInt.
func (s SQLVarchar) SQLInt() SQLValue {
	// First, clean up extraneous characters.
	cleaned := MySQLCleanNumericString(string(s))
	// Then convert to int.
	out, _ := strconv.ParseInt(strings.Split(cleaned, ".")[0], 10, 64)
	return SQLInt64(out)
}

// SQLTimestamp converts the SQLVarchar receiver, s, to a SQLTimestamp.
func (s SQLVarchar) SQLTimestamp() SQLValue {
	t, _, ok := parseDateTime(s.String())
	if !ok {
		return SQLTimestamp{Time: NullDate}
	}
	t = t.In(schema.DefaultLocale)
	return SQLTimestamp{Time: t}
}

// SQLUint converts the SQLVarchar receiver, s, to a SQLUint.
func (s SQLVarchar) SQLUint() SQLValue {
	// First, clean up extraneous characters.
	cleaned := MySQLCleanNumericString(string(s))
	// Then convert to int.
	out, _ := strconv.ParseInt(strings.Split(cleaned, ".")[0], 10, 64)
	return SQLUint64(out)
}

// SQLVarchar converts the SQLVarchar receiver, s, to a SQLVarchar.
func (s SQLVarchar) SQLVarchar() SQLValue {
	return s
}
