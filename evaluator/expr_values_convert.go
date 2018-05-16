package evaluator

import (
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
)

// NullDate is an internal representation for the MySQL
// date 0000-00-00, which cannot be represented as all 0's.
var NullDate = time.Date(-1, 1, 0, 0, 0, 0, 0, schema.DefaultLocale)

// ConvertTo takes a SQLValue v and a byte bsonSpecType, that determines what.
// type to convert the passed SQLValue. The byte is based off the.
// BSON spec byte type for types that are BSON types.
func ConvertTo(v SQLValue, bsonSpecType schema.BSONSpecType) SQLValue {
	switch bsonSpecType {
	case schema.BSONBoolean:
		return v.SQLBool()
	case schema.BSONDecimal128:
		return v.SQLDecimal128()
	case schema.BSONDouble:
		return v.SQLFloat()
	case schema.BSONInt:
		return v.SQLInt()
	case schema.BSONInt64:
		return v.SQLInt()
	case schema.BSONNone:
		return v
	case schema.BSONNull:
		return SQLNull
	case schema.BSONObjectID:
		return v.SQLObjectID()
	case schema.BSONString:
		return v.SQLVarchar()
	case schema.BSONDatetime:
		return v.SQLTimestamp()
	case schema.BSONUUID:
		return v.SQLUUID()
	// Types not corresponding to MongoDB types.
	case schema.BSONDate:
		return v.SQLDate()
	case schema.BSONUint64:
		if maybeInt, ok := v.SQLInt().(SQLInt); ok {
			return SQLUint64(maybeInt)
		}
		return SQLNull
	}
	return SQLNull
}

// Bool Conversion

// ConvertTo converts the SQLBool receiver, s, to the specified BSONSpecType.
func (s SQLBool) ConvertTo(bsonSpecType schema.BSONSpecType) SQLValue {
	return ConvertTo(s, bsonSpecType)
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
		return SQLDecimal128(decimal.Zero)
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
		return SQLInt(1)
	}
	return SQLInt(0)
}

// SQLObjectID converts the SQLBool receiver, s, to a SQLObjectID.
func (s SQLBool) SQLObjectID() SQLValue {
	if s == SQLTrue {
		return SQLNull
	}
	return SQLNull
}

// SQLTimestamp converts the SQLBool receiver, s, to a SQLTimestamp.
func (s SQLBool) SQLTimestamp() SQLValue {
	if s == SQLTrue {
		return SQLNull
	}
	return SQLTimestamp{Time: NullDate}
}

// SQLUUID converts the SQLBool receiver, s, to a SQLUUID.
func (s SQLBool) SQLUUID() SQLValue {
	if s == SQLTrue {
		return SQLNull
	}
	return SQLNull
}

// SQLVarchar converts the SQLBool receiver, s, to a SQLVarchar.
func (s SQLBool) SQLVarchar() SQLValue {
	if s == SQLTrue {
		return SQLVarchar("1")
	}
	return SQLVarchar("0")
}

// Date Conversion

// ConvertTo converts the SQLDate receiver, s, to the specified BSONSpecType.
func (s SQLDate) ConvertTo(bsonSpecType schema.BSONSpecType) SQLValue {
	return ConvertTo(s, bsonSpecType)
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
	return (SQLFloat(t.Day())*1e6 +
		SQLFloat(t.Month())*1e8 +
		SQLFloat(t.Year())*1e10)
}

// SQLInt converts the SQLDate receiver, s, to a SQLInt.
func (s SQLDate) SQLInt() SQLValue {
	t := s.Time
	return (SQLInt(t.Day())*1e6 +
		SQLInt(t.Month())*1e8 +
		SQLInt(t.Year())*1e10)
}

// SQLObjectID converts the SQLDate receiver, s, to a SQLObjectID.
func (s SQLDate) SQLObjectID() SQLValue {
	return SQLNull
}

// SQLTimestamp converts the SQLDate receiver, s, to a SQLTimestamp.
func (s SQLDate) SQLTimestamp() SQLValue {
	t := s.Time
	return SQLTimestamp{
		Time: t,
	}
}

// SQLUUID converts the SQLDate receiver, s, to a SQLUUID.
func (s SQLDate) SQLUUID() SQLValue {
	return SQLNull
}

// SQLVarchar converts the SQLDate receiver, s, to a SQLVarchar.
func (s SQLDate) SQLVarchar() SQLValue {
	return SQLVarchar(s.String())
}

// Decimal Conversion

// ConvertTo converts the SQLDecimal128 receiver, s, to the specified BSONSpecType.
func (s SQLDecimal128) ConvertTo(bsonSpecType schema.BSONSpecType) SQLValue {
	return ConvertTo(s, bsonSpecType)
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
	return SQLInt(round(f))
}

// SQLObjectID converts the SQLDecimal128 receiver, s, to a SQLObjectID.
func (s SQLDecimal128) SQLObjectID() SQLValue {
	return SQLNull
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

// SQLUUID converts the SQLDecimal128 receiver, s, to a SQLUUID.
func (s SQLDecimal128) SQLUUID() SQLValue {
	return SQLNull
}

// SQLVarchar converts the SQLDecimal128 receiver, s, to a SQLVarchar.
func (s SQLDecimal128) SQLVarchar() SQLValue {
	return SQLVarchar(s.String())
}

// Float Conversion

// ConvertTo converts the SQLFloat receiver, s, to the specified BSONSpecType.
func (s SQLFloat) ConvertTo(bsonSpecType schema.BSONSpecType) SQLValue {
	return ConvertTo(s, bsonSpecType)
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
	return SQLInt(round(float64(s)))
}

// SQLObjectID converts the SQLFloat receiver, s, to a SQLObjectID.
func (s SQLFloat) SQLObjectID() SQLValue {
	return SQLNull
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

// SQLUUID converts the SQLFloat receiver, s, to a SQLUUID.
func (s SQLFloat) SQLUUID() SQLValue {
	return SQLNull
}

// SQLVarchar converts the SQLFloat receiver, s, to a SQLVarchar.
func (s SQLFloat) SQLVarchar() SQLValue {
	return SQLVarchar(s.String())
}

// Int Conversion

// ConvertTo converts the SQLInt receiver, s, to the specified BSONSpecType.
func (s SQLInt) ConvertTo(bsonSpecType schema.BSONSpecType) SQLValue {
	return ConvertTo(s, bsonSpecType)
}

// SQLBool converts the SQLInt receiver, s, to a SQLBool.
func (s SQLInt) SQLBool() SQLValue {
	if s == 0 {
		return SQLFalse
	}
	return SQLTrue
}

// SQLDate converts the SQLInt receiver, s, to a SQLDate.
func (s SQLInt) SQLDate() SQLValue {
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
func (s SQLInt) SQLDecimal128() SQLValue {
	return SQLDecimal128(decimal.New(int64(s), 0))
}

// SQLFloat converts the SQLInt receiver, s, to a SQLFloat.
func (s SQLInt) SQLFloat() SQLValue {
	return SQLFloat(s)
}

// SQLInt converts the SQLInt receiver, s, to a SQLInt.
func (s SQLInt) SQLInt() SQLValue {
	return s
}

// SQLObjectID converts the SQLInt receiver, s, to a SQLObjectID.
func (s SQLInt) SQLObjectID() SQLValue {
	return SQLNull
}

// SQLTimestamp converts the SQLInt receiver, s, to a SQLTimestamp.
func (s SQLInt) SQLTimestamp() SQLValue {
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

// SQLUUID converts the SQLInt receiver, s, to a SQLUUID.
func (s SQLInt) SQLUUID() SQLValue {
	return SQLNull
}

// SQLVarchar converts the SQLInt receiver, s, to a SQLVarchar.
func (s SQLInt) SQLVarchar() SQLValue {
	return SQLVarchar(s.String())
}

// SQLNoValue Conversion

// ConvertTo converts the SQLNoValue receiver to the specified BSONSpecType.
func (SQLNoValue) ConvertTo(bsonSpecType schema.BSONSpecType) SQLValue {
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
	return SQLInt(0)
}

// SQLObjectID converts the SQLNoValue receiver to a SQLObjectID.
func (SQLNoValue) SQLObjectID() SQLValue {
	return SQLNull
}

// SQLTimestamp converts the SQLNoValue receiver to a SQLTimestamp.
func (SQLNoValue) SQLTimestamp() SQLValue {
	return SQLTimestamp{Time: NullDate}
}

// SQLUUID converts the SQLNoValue receiver to a SQLUUID.
func (SQLNoValue) SQLUUID() SQLValue {
	return SQLNull
}

// SQLVarchar converts the SQLNoValue receiver to a SQLVarchar.
func (SQLNoValue) SQLVarchar() SQLValue {
	return SQLVarchar(schema.SQLNone)
}

// SQLNull Conversion

// ConvertTo converts the SQLNullValue receiver to the specified BSONSpecType.
func (SQLNullValue) ConvertTo(bsonSpecType schema.BSONSpecType) SQLValue {
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

// SQLObjectID converts the SQLNullValue receiver to a SQLObjectID.
func (SQLNullValue) SQLObjectID() SQLValue {
	return SQLNull
}

// SQLTimestamp converts the SQLNullValue receiver to a SQLTimestamp.
func (SQLNullValue) SQLTimestamp() SQLValue {
	return SQLNull
}

// SQLUUID converts the SQLNullValue receiver to a SQLUUID.
func (SQLNullValue) SQLUUID() SQLValue {
	return SQLNull
}

// SQLVarchar converts the SQLNullValue receiver to a SQLVarchar.
func (SQLNullValue) SQLVarchar() SQLValue {
	return SQLNull
}

// ObjectID Conversion

// ConvertTo converts the SQLObjectID receiver, s, to the specified BSONSpecType.
func (s SQLObjectID) ConvertTo(bsonSpecType schema.BSONSpecType) SQLValue {
	return ConvertTo(s, bsonSpecType)
}

// SQLBool converts a SQLObjectID to a SQLBool.
func (SQLObjectID) SQLBool() SQLValue {
	return SQLTrue
}

// SQLDate converts a SQLObjectID to a SQLDate.
func (SQLObjectID) SQLDate() SQLValue {
	return SQLNull
}

// SQLDecimal128 converts a SQLObjectID to a SQLDecimal128.
func (SQLObjectID) SQLDecimal128() SQLValue {
	return SQLDecimal128(decimal.Zero)
}

// SQLFloat converts a SQLObjectID to a SQLFloat.
func (SQLObjectID) SQLFloat() SQLValue {
	return SQLFloat(0.0)
}

// SQLInt converts a SQLObjectID to a SQLInt.
func (SQLObjectID) SQLInt() SQLValue {
	return SQLInt(0)
}

// SQLObjectID converts the SQLObjectID receiver, s, to a SQLObjectID.
func (s SQLObjectID) SQLObjectID() SQLValue {
	return s
}

// SQLTimestamp converts a SQLObjectID to a SQLTimestamp.
func (SQLObjectID) SQLTimestamp() SQLValue {
	return SQLNull
}

// SQLUUID converts a SQLObjectID to a SQLUUID.
func (SQLObjectID) SQLUUID() SQLValue {
	return SQLNull
}

// SQLVarchar converts the SQLObjectID receiver, s, to a SQLVarchar.
func (s SQLObjectID) SQLVarchar() SQLValue {
	return SQLVarchar(string(s))
}

// Timestamp Conversion

// ConvertTo converts the SQLTimestamp receiver, s, to the specified BSONSpecType.
func (s SQLTimestamp) ConvertTo(bsonSpecType schema.BSONSpecType) SQLValue {
	return ConvertTo(s, bsonSpecType)
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
	return (SQLInt(t.Second()) +
		SQLInt(t.Minute())*1e2 +
		SQLInt(t.Hour())*1e4 +
		SQLInt(t.Day())*1e6 +
		SQLInt(t.Month())*1e8 +
		SQLInt(t.Year())*1e10)
}

// SQLObjectID converts the SQLTimestamp receiver, s, to a SQLObjectID.
func (s SQLTimestamp) SQLObjectID() SQLValue {
	return SQLNull
}

// SQLTimestamp converts the SQLTimestamp receiver, s, to a SQLTimestamp.
func (s SQLTimestamp) SQLTimestamp() SQLValue {
	return s
}

// SQLUUID converts the SQLTimestamp receiver, s, to a SQLUUID.
func (s SQLTimestamp) SQLUUID() SQLValue {
	return SQLNull
}

// SQLVarchar converts the SQLTimestamp receiver, s, to a SQLVarchar.
func (s SQLTimestamp) SQLVarchar() SQLValue {
	return SQLVarchar(s.String())
}

// UInt32 Conversion

// ConvertTo converts the SQLUint32 receiver, s, to the specified BSONSpecType.
func (s SQLUint32) ConvertTo(bsonSpecType schema.BSONSpecType) SQLValue {
	return ConvertTo(s, bsonSpecType)
}

// SQLBool converts the SQLUint32 receiver, s, to a SQLBool.
func (s SQLUint32) SQLBool() SQLValue {
	if s == 0 {
		return SQLFalse
	}
	return SQLTrue
}

// SQLDate converts the SQLUint32 receiver, s, to a SQLDate.
func (s SQLUint32) SQLDate() SQLValue {
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

// SQLDecimal128 converts the SQLUint32 receiver, s, to a SQLDecimal128.
func (s SQLUint32) SQLDecimal128() SQLValue {
	return SQLDecimal128(decimal.New(int64(s), 0))
}

// SQLFloat converts the SQLUint32 receiver, s, to a SQLFloat.
func (s SQLUint32) SQLFloat() SQLValue {
	return SQLFloat(s)
}

// SQLInt converts the SQLUint32 receiver, s, to a SQLInt.
func (s SQLUint32) SQLInt() SQLValue {
	return s
}

// SQLObjectID converts the SQLUint32 receiver, s, to a SQLObjectID.
func (s SQLUint32) SQLObjectID() SQLValue {
	return SQLNull
}

// SQLTimestamp converts the SQLUint32 receiver, s, to a SQLTimestamp.
func (s SQLUint32) SQLTimestamp() SQLValue {
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

// SQLUUID converts the SQLUint32 receiver, s, to a SQLUUID.
func (s SQLUint32) SQLUUID() SQLValue {
	return SQLNull
}

// SQLVarchar converts the SQLUint32 receiver, s, to a SQLVarchar.
func (s SQLUint32) SQLVarchar() SQLValue {
	return SQLVarchar(s.String())
}

// UInt64 Conversion

// ConvertTo converts the SQLUint64 receiver, s, to the specified BSONSpecType.
func (s SQLUint64) ConvertTo(bsonSpecType schema.BSONSpecType) SQLValue {
	return ConvertTo(s, bsonSpecType)
}

// SQLBool converts the SQLUint64 receiver, s, to a SQLBool.
func (s SQLUint64) SQLBool() SQLValue {
	if s == 0 {
		return SQLFalse
	}
	return SQLTrue
}

// SQLDate converts the SQLUint64 receiver, s, to a SQLDate.
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

// SQLDecimal128 converts the SQLUint64 receiver, s, to a SQLDecimal128.
func (s SQLUint64) SQLDecimal128() SQLValue {
	return SQLDecimal128(decimal.New(int64(s), 0))
}

// SQLFloat converts the SQLUint64 receiver, s, to a SQLFloat.
func (s SQLUint64) SQLFloat() SQLValue {
	return SQLFloat(s)
}

// SQLInt converts the SQLUint64 receiver, s, to a SQLInt.
func (s SQLUint64) SQLInt() SQLValue {
	return s
}

// SQLObjectID converts the SQLUint64 receiver, s, to a SQLObjectID.
func (s SQLUint64) SQLObjectID() SQLValue {
	return SQLNull
}

// SQLTimestamp converts the SQLUint64 receiver, s, to a SQLTimestamp.
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

// SQLUUID converts the SQLUint64 receiver, s, to a SQLUUID.
func (s SQLUint64) SQLUUID() SQLValue {
	return SQLNull
}

// SQLVarchar converts the SQLUint64 receiver, s, to a SQLVarchar.
func (s SQLUint64) SQLVarchar() SQLValue {
	return SQLVarchar(s.String())
}

// UUID Conversion

// ConvertTo converts the SQLUUID receiver, s, to the specified BSONSpecType.
func (s SQLUUID) ConvertTo(bsonSpecType schema.BSONSpecType) SQLValue {
	return ConvertTo(s, bsonSpecType)
}

// SQLBool converts the SQLUUID receiver to a SQLBool.
func (SQLUUID) SQLBool() SQLValue {
	return SQLTrue
}

// SQLDate converts the SQLUUID receiver to a SQLDate.
func (SQLUUID) SQLDate() SQLValue {
	return SQLNull
}

// SQLDecimal128 converts the SQLUUID receiver to a SQLDecimal128.
func (SQLUUID) SQLDecimal128() SQLValue {
	return SQLDecimal128(decimal.Zero)
}

// SQLFloat converts the SQLUUID receiver to a SQLFloat.
func (SQLUUID) SQLFloat() SQLValue {
	return SQLFloat(0.0)
}

// SQLInt converts the SQLUUID receiver to a SQLInt.
func (SQLUUID) SQLInt() SQLValue {
	return SQLInt(0)
}

// SQLObjectID converts the SQLUUID receiver to a SQLObjectID.
func (SQLUUID) SQLObjectID() SQLValue {
	return SQLNull
}

// SQLTimestamp converts the SQLUUID receiver to a SQLTimestamp.
func (SQLUUID) SQLTimestamp() SQLValue {
	return SQLNull
}

// SQLUUID converts the SQLUUID receiver, s, to a SQLUUID.
func (s SQLUUID) SQLUUID() SQLValue {
	return s
}

// SQLVarchar converts the SQLUUID receiver, s, to a SQLVarchar.
func (s SQLUUID) SQLVarchar() SQLValue {
	str := hex.EncodeToString(s.bytes)
	ret := str[0:8] +
		"-" + str[8:12] +
		"-" + str[12:16] +
		"-" + str[16:20] +
		"-" + str[20:]
	return SQLVarchar(ret)
}

// *SQLValues Conversion

// ConvertTo converts the *SQLValues receiver, s, to the specified BSONSpecType.
func (s *SQLValues) ConvertTo(bsonSpecType schema.BSONSpecType) SQLValue {
	return ConvertTo(s.Values[0], bsonSpecType)
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

// SQLObjectID converts the *SQLValues receiver, s, to a SQLObjectID.
func (s *SQLValues) SQLObjectID() SQLValue {
	return s.Values[0].SQLObjectID()
}

// SQLTimestamp converts the *SQLValues receiver, s, to a SQLTimestamp.
func (s *SQLValues) SQLTimestamp() SQLValue {
	return s.Values[0].SQLTimestamp()
}

// SQLUUID converts the *SQLValues receiver, s, to a SQLUUID.
func (s *SQLValues) SQLUUID() SQLValue {
	return s.Values[0].SQLUUID()
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

// Varchar Conversion

// ConvertTo converts the SQLVarchar receiver, s, to the specified BSONSpecType.
func (s SQLVarchar) ConvertTo(bsonSpecType schema.BSONSpecType) SQLValue {
	return ConvertTo(s, bsonSpecType)
}

// SQLBool converts the SQLVarchar receiver, s, to a SQLBool.
func (s SQLVarchar) SQLBool() SQLValue {
	return s.SQLInt().SQLBool()
}

// SQLDate converts the SQLVarchar receiver, s, to a SQLDate.
func (s SQLVarchar) SQLDate() SQLValue {
	t, _, ok := parseDateTime(s.String())
	if !ok {
		return SQLNull
	}
	t = t.In(schema.DefaultLocale)
	return SQLDate{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale)}
}

// SQLDecimal128 converts the SQLVarchar receiver, s, to a SQLDecimal128.
func (s SQLVarchar) SQLDecimal128() SQLValue {
	cleaned := CleanNumericString(string(s))
	out, err := decimal.NewFromString(cleaned)
	if err != nil {
		return SQLDecimal128(decimal.Zero)
	}
	return SQLDecimal128(out)
}

// SQLFloat converts the SQLVarchar receiver, s, to a SQLFloat.
func (s SQLVarchar) SQLFloat() SQLValue {
	cleaned := CleanNumericString(string(s))
	out, _ := strconv.ParseFloat(cleaned, 64)
	return SQLFloat(out)
}

// SQLInt converts the SQLVarchar receiver, s, to a SQLInt.
func (s SQLVarchar) SQLInt() SQLValue {
	// First, clean up extraneous characters.
	cleaned := CleanNumericString(string(s))
	// Then convert to int.
	out, _ := strconv.ParseInt(strings.Split(cleaned, ".")[0], 10, 64)
	return SQLInt(out)
}

// SQLObjectID converts the SQLVarchar receiver, s, to a SQLObjectID.
func (s SQLVarchar) SQLObjectID() SQLValue {
	return SQLNull
}

// SQLTimestamp converts the SQLVarchar receiver, s, to a SQLTimestamp.
func (s SQLVarchar) SQLTimestamp() SQLValue {
	t, _, ok := parseDateTime(s.String())
	if !ok {
		return SQLNull
	}
	t = t.In(schema.DefaultLocale)
	return SQLTimestamp{Time: t}
}

// SQLUUID converts the SQLVarchar receiver, s, to a SQLUUID.
func (s SQLVarchar) SQLUUID() SQLValue {
	return SQLNull
}

// SQLVarchar converts the SQLVarchar receiver, s, to a SQLVarchar.
func (s SQLVarchar) SQLVarchar() SQLValue {
	return s
}
