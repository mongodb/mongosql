package values

import (
	"strconv"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
)

// MySQLBool represents a boolean with MySQL type conversion semantics.
type MySQLBool struct {
	BaseSQLBool
}

// MySQLDate represents a date value with MySQL type conversion semantics.
type MySQLDate struct {
	BaseSQLDate
}

// SQLDecimal128 converts the SQLDate receiver, s, to a SQLDecimal128.
func (s MySQLDate) SQLDecimal128() SQLDecimal128 {
	flt := Float64(s)
	dec := decimal.NewFromFloat(flt)
	return NewSQLDecimal128(MySQLValueKind, dec)
}

// SQLFloat converts the SQLDate receiver, s, to a SQLFloat.
func (s MySQLDate) SQLFloat() SQLFloat {
	t := s.datetime
	i := t.Day() + int(t.Month())*1e2 + t.Year()*1e4
	return NewSQLFloat(MySQLValueKind, float64(i))
}

// SQLInt converts the SQLDate receiver, s, to a SQLInt.
func (s MySQLDate) SQLInt() SQLInt64 {
	t := s.datetime
	return NewSQLInt64(MySQLValueKind, int64(t.Day())+
		int64(t.Month())*1e2+
		int64(t.Year())*1e4)
}

// SQLUint converts the SQLDate receiver, s, to a SQLUint64.
func (s MySQLDate) SQLUint() SQLUint64 {
	t := s.datetime
	return NewSQLUint64(MySQLValueKind, uint64(t.Day())+
		uint64(t.Month())*1e2+
		uint64(t.Year())*1e4)
}

// MySQLDecimal128 represents a decimal 128 value with MySQL type conversion semantics.
type MySQLDecimal128 struct {
	BaseSQLDecimal128
}

// SQLTimestamp converts the SQLDecimal128 receiver, s, to a SQLTimestamp.
func (s MySQLDecimal128) SQLTimestamp() SQLTimestamp {
	if s.val.Equals(decimal.Zero) {
		return NewSQLTimestamp(MySQLValueKind, NullDate)
	}

	dateStr, ok := paddedDateString(s)
	if !ok {
		return NewSQLNull(MySQLValueKind)
	}

	t, _, ok := ParseDateTime(dateStr)
	if !ok {
		return NewSQLNull(MySQLValueKind)
	}

	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(MySQLValueKind, t)
}

// SQLDate converts the SQLDecimal128 receiver, s, to a SQLDate.
func (s MySQLDecimal128) SQLDate() SQLDate {
	if s.val.Equals(decimal.Zero) {
		return NewSQLDate(MySQLValueKind, NullDate)
	}

	dateStr, ok := paddedDateString(s)
	if !ok {
		return NewSQLNull(MySQLValueKind)
	}

	t, _, ok := ParseDateTime(dateStr)
	if !ok {
		return NewSQLNull(MySQLValueKind)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		MySQLValueKind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// MySQLFloat represents a float with MySQL type conversion semantics.
type MySQLFloat struct {
	BaseSQLFloat
}

// SQLDate converts the SQLFloat receiver, s, to a SQLDate.
func (s MySQLFloat) SQLDate() SQLDate {
	if s.val == 0.0 {
		return NewSQLDate(MySQLValueKind, NullDate)
	}

	dateStr, ok := paddedDateString(s)
	if !ok {
		return NewSQLNull(MySQLValueKind)
	}

	t, _, ok := ParseDateTime(dateStr)
	if !ok {
		return NewSQLNull(MySQLValueKind)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		MySQLValueKind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLTimestamp converts the SQLFloat receiver, s, to a SQLTimestamp.
func (s MySQLFloat) SQLTimestamp() SQLTimestamp {
	if s.val == 0.0 {
		return NewSQLTimestamp(MySQLValueKind, NullDate)
	}

	dateStr, ok := paddedDateString(s)
	if !ok {
		return NewSQLNull(MySQLValueKind)
	}

	t, _, ok := ParseDateTime(dateStr)
	if !ok {
		return NewSQLNull(MySQLValueKind)
	}

	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(MySQLValueKind, t)
}

// MySQLInt64 represents a 64-bit integer value with MySQL type conversion semantics.
type MySQLInt64 struct {
	BaseSQLInt64
}

// SQLDate converts the SQLInt receiver, s, to a SQLDate.
func (s MySQLInt64) SQLDate() SQLDate {
	if s.val == 0 {
		return NewSQLDate(MySQLValueKind, NullDate)
	}

	dateStr, ok := paddedDateString(s)
	if !ok {
		return NewSQLNull(MySQLValueKind)
	}

	t, _, ok := ParseDateTime(dateStr)
	if !ok {
		return NewSQLNull(MySQLValueKind)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		MySQLValueKind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLTimestamp converts the SQLInt receiver, s, to a SQLTimestamp.
func (s MySQLInt64) SQLTimestamp() SQLTimestamp {
	if s.val == 0 {
		return NewSQLTimestamp(MySQLValueKind, NullDate)
	}

	dateStr, ok := paddedDateString(s)
	if !ok {
		return NewSQLNull(MySQLValueKind)
	}

	t, _, ok := ParseDateTime(dateStr)
	if !ok {
		return NewSQLNull(MySQLValueKind)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(MySQLValueKind, t)
}

// MySQLObjectID represents an ObjectID value with MySQL type conversion semantics.
type MySQLObjectID struct {
	BaseSQLObjectID
}

// SQLDecimal128 converts a MySQLObjectID to a SQLDecimal128 by converting to SQLTimestamp then
// to SQLDecimal128.
func (s MySQLObjectID) SQLDecimal128() SQLDecimal128 {
	return s.SQLTimestamp().SQLDecimal128()
}

// SQLFloat converts a MySQLObjectID to a SQLFloat by converting to SQLTimestamp then to SQLFloat.
func (s MySQLObjectID) SQLFloat() SQLFloat {
	return s.SQLTimestamp().SQLFloat()
}

// SQLInt converts a MySQLObjectID to a SQLInt by converting to SQLTimestamp then to SQLInt.
func (s MySQLObjectID) SQLInt() SQLInt64 {
	return s.SQLTimestamp().SQLInt()
}

// SQLUint converts a MySQLObjectID to a SQLUint64 by converting to SQLTimestamp then to SQLUint.
func (s MySQLObjectID) SQLUint() SQLUint64 {
	return s.SQLTimestamp().SQLUint()
}

// MySQLTimestamp represents a timestamp value with MySQL type conversion semantics.
type MySQLTimestamp struct {
	BaseSQLTimestamp
}

// SQLDecimal128 converts the SQLTimestamp receiver, s, to a SQLDecimal128.
func (s MySQLTimestamp) SQLDecimal128() SQLDecimal128 {
	flt := Float64(s)
	dec := decimal.NewFromFloat(flt)
	return NewSQLDecimal128(MySQLValueKind, dec)
}

// SQLFloat converts the SQLTimestamp receiver, s, to a SQLFloat.
func (s MySQLTimestamp) SQLFloat() SQLFloat {
	t := s.datetime
	return NewSQLFloat(MySQLValueKind, float64(t.Second())+
		float64(t.Minute())*1e2+
		float64(t.Hour())*1e4+
		float64(t.Day())*1e6+
		float64(t.Month())*1e8+
		float64(t.Year())*1e10+
		float64(t.Nanosecond())/1e9)
}

// SQLInt converts the SQLTimestamp receiver, s, to a SQLInt.
func (s MySQLTimestamp) SQLInt() SQLInt64 {
	t := s.datetime
	return NewSQLInt64(MySQLValueKind, int64(t.Second())+
		int64(t.Minute())*1e2+
		int64(t.Hour())*1e4+
		int64(t.Day())*1e6+
		int64(t.Month())*1e8+
		int64(t.Year())*1e10)
}

// SQLUint converts the SQLTimestamp receiver, s, to a SQLUint64.
func (s MySQLTimestamp) SQLUint() SQLUint64 {
	t := s.datetime
	return NewSQLUint64(MySQLValueKind, uint64(t.Second())+
		uint64(t.Minute())*1e2+
		uint64(t.Hour())*1e4+
		uint64(t.Day())*1e6+
		uint64(t.Month())*1e8+
		uint64(t.Year())*1e10)
}

// SQLVarchar converts the SQLTimestamp receiver, s, to a SQLVarchar.
func (s MySQLTimestamp) SQLVarchar() SQLVarchar {
	return NewSQLVarchar(MySQLValueKind, s.varchar())
}
func (s MySQLTimestamp) varchar() string {
	if s.null {
		return "NULL"
	}
	return s.datetime.Format("2006-01-02 15:04:05.000000")
}

// MySQLUint64 represents an unsigned 64-bit integer with MySQL type conversion semantics.
type MySQLUint64 struct {
	BaseSQLUint64
}

// MySQLVarchar represents a string value with MySQL type conversion semantics.
type MySQLVarchar struct {
	BaseSQLVarchar
}

// SQLDate converts the SQLVarchar receiver, s, to a SQLDate.
func (s MySQLVarchar) SQLDate() SQLDate {
	t, _, ok := ParseDateTime(s.varchar())
	if !ok {
		return NewSQLNull(MySQLValueKind)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		MySQLValueKind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLBool converts a MySQLVarchar to a SQLBool by converting to Int then to Bool.
// The conversion to Int will properly clean things up.
func (s MySQLVarchar) SQLBool() SQLBool {
	return s.SQLInt().SQLBool()
}

// SQLDecimal128 converts the SQLVarchar receiver, s, to a SQLDecimal128.
func (s MySQLVarchar) SQLDecimal128() SQLDecimal128 {
	cleaned := strutil.MySQLCleanScientificNotationString(s.val)
	out, err := decimal.NewFromString(cleaned)
	if err != nil {
		return NewSQLDecimal128(MySQLValueKind, decimal.Zero)
	}
	return NewSQLDecimal128(MySQLValueKind, out)
}

// SQLFloat converts the SQLVarchar receiver, s, to a SQLFloat.
func (s MySQLVarchar) SQLFloat() SQLFloat {
	// First, clean up extraneous characters.
	cleaned := strutil.MySQLCleanNumericString(s.val)
	// Then convert to float.
	out, _ := strconv.ParseFloat(cleaned, 64)
	return NewSQLFloat(MySQLValueKind, out)
}

// SQLInt converts the SQLVarchar receiver, s, to a SQLInt.
func (s MySQLVarchar) SQLInt() SQLInt64 {
	// First, clean up extraneous characters.
	cleaned := strutil.MySQLCleanNumericString(s.val)
	// Then convert to int.
	out, _ := strconv.ParseInt(strings.Split(cleaned, ".")[0], 10, 64)
	return NewSQLInt64(MySQLValueKind, out)
}

// SQLTimestamp converts the SQLVarchar receiver, s, to a SQLTimestamp.
func (s MySQLVarchar) SQLTimestamp() SQLTimestamp {
	t, _, ok := ParseDateTime(s.varchar())
	if !ok {
		return NewSQLNull(MySQLValueKind)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(MySQLValueKind, t)
}

// SQLUint converts the SQLVarchar receiver, s, to a SQLUint64.
func (s MySQLVarchar) SQLUint() SQLUint64 {
	// First, clean up extraneous characters.
	cleaned := strutil.MySQLCleanNumericString(s.val)
	// Then convert to int.
	out, _ := strconv.ParseInt(strings.Split(cleaned, ".")[0], 10, 64)
	return NewSQLUint64(MySQLValueKind, uint64(out))
}

// MySQLNull represents a varchar with MySQL type conversion semantics.
// MySQLNull shares all of its conversion implementations with BaseSQLNull.
type MySQLNull struct {
	BaseSQLNull
}
