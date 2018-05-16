package evaluator

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
)

// SQLTrue and SQLFalse represented the SQLValues for true and false, respectively.
const (
	SQLTrue  = SQLBool(1)
	SQLFalse = SQLBool(0)
)

// SQLBool represents a boolean.
type SQLBool float64

// Bool returns the boolean value of this SQLValue.
func (sb SQLBool) Bool() bool {
	return sb > 0
}

// Decimal128 returns the decimal value of this SQLValue.
func (sb SQLBool) Decimal128() decimal.Decimal {
	return decimal.NewFromFloat(sb.Float64())
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (sb SQLBool) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sb, nil
}

// Float64 returns the floating-point value of this SQLValue.
func (sb SQLBool) Float64() float64 {
	return float64(sb)
}

// Int64 returns the integer value of this SQLValue.
func (sb SQLBool) Int64() int64 {
	return int64(sb)
}

// MySQLEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (sb SQLBool) MySQLEncode(*collation.Charset, int) ([]byte, error) {
	if sb == SQLTrue {
		return []byte{49}, nil
	}
	return []byte{48}, nil
}

// Size returns the size of this SQLValue in bytes.
func (sb SQLBool) Size() uint64 {
	return 1
}

func (sb SQLBool) String() string {
	return strconv.FormatFloat(sb.Float64(), 'f', -1, 64)
}

// ToAggregationLanguage translates SQLBool into something that can
// be used in an aggregation pipeline. If SQLBool cannot be translated,
// it will return nil and false.
func (sb SQLBool) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	return wrapInLiteral(sb.Bool()), true
}

// Type returns the SQLType of this SQLValue.
func (SQLBool) Type() schema.SQLType {
	return schema.SQLBoolean
}

// Uint64 returns the unsigned integer value of this SQLValue.
func (sb SQLBool) Uint64() uint64 {
	return uint64(sb)
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (sb SQLBool) Value() interface{} {
	return sb > 0
}

// SQLDate represents a date.
type SQLDate struct {
	Time time.Time
}

// Decimal128 returns the decimal value of this SQLValue.
func (sd SQLDate) Decimal128() decimal.Decimal {
	return decimal.NewFromFloat(sd.Float64())
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (sd SQLDate) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sd, nil
}

// Float64 returns the floating-point value of this SQLValue.
func (sd SQLDate) Float64() float64 {
	val, _ := strconv.ParseFloat(sd.Time.Format("20060102"), 64)
	return val
}

// Int64 returns the integer value of this SQLValue.
func (sd SQLDate) Int64() int64 {
	val, _ := strconv.ParseInt(sd.Time.Format("20060102"), 10, 64)
	return val
}

// MySQLEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (sd SQLDate) MySQLEncode(*collation.Charset, int) ([]byte, error) {
	if sd.Time == NullDate {
		return []byte("0000-00-00"), nil
	}
	return util.Slice(sd.Time.Format(schema.DateFormat)), nil
}

// Size returns the size of this SQLValue in bytes.
func (sd SQLDate) Size() uint64 {
	return 8
}

func (sd SQLDate) String() string {
	return sd.Time.Format("2006-01-02")
}

// ToAggregationLanguage translates SQLDate into something that can
// be used in an aggregation pipeline. If SQLDate cannot be translated,
// it will return nil and false.
func (sd SQLDate) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	return wrapInLiteral(sd.Time), true
}

// Type returns the SQLType of this SQLValue.
func (SQLDate) Type() schema.SQLType {
	return schema.SQLDate
}

// Uint64 returns the unsigned integer value of this SQLValue.
func (sd SQLDate) Uint64() uint64 {
	val, _ := strconv.ParseUint(sd.Time.Format("20060102"), 10, 64)
	return val
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (sd SQLDate) Value() interface{} {
	return sd.Time
}

// SQLDecimal128 represents a decimal 128 value.
type SQLDecimal128 decimal.Decimal

// Decimal128 returns the decimal value of this SQLValue.
func (sd SQLDecimal128) Decimal128() decimal.Decimal {
	return decimal.Decimal(sd)
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (sd SQLDecimal128) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return sd, nil
}

// Float64 returns the floating-point value of this SQLValue.
func (sd SQLDecimal128) Float64() float64 {
	// second return value is f represents sd exactly
	f, _ := decimal.Decimal(sd).Float64()
	return f
}

// Int64 returns the integer value of this SQLValue.
func (sd SQLDecimal128) Int64() int64 {
	return decimal.Decimal(sd).Truncate(0).IntPart()
}

// MySQLEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (sd SQLDecimal128) MySQLEncode(*collation.Charset, int) ([]byte, error) {
	return []byte(util.FormatDecimal(decimal.Decimal(sd))), nil
}

// Size returns the size of this SQLValue in bytes.
func (sd SQLDecimal128) Size() uint64 {
	return 16
}

func (sd SQLDecimal128) String() string {
	return decimal.Decimal(sd).String()
}

// Type returns the SQLType of this SQLValue.
func (SQLDecimal128) Type() schema.SQLType {
	return schema.SQLDecimal128
}

// Uint64 returns the unsigned integer value of this SQLValue.
func (sd SQLDecimal128) Uint64() uint64 {
	return uint64(decimal.Decimal(sd).Truncate(0).IntPart())
}

// ToAggregationLanguage translates SQLDecimal128 into something that can
// be used in an aggregation pipeline. If SQLDecimal128 cannot be translated,
// it will return nil and false.
func (sd SQLDecimal128) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	d, ok := t.translateDecimal(sd)
	if !ok {
		return nil, false
	}
	return wrapInLiteral(d), true
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (sd SQLDecimal128) Value() interface{} {
	return decimal.Decimal(sd)
}

// SQLFloat represents a float.
type SQLFloat float64

// Decimal128 returns the decimal value of this SQLValue.
func (sf SQLFloat) Decimal128() decimal.Decimal {
	return decimal.NewFromFloat(float64(sf))
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (sf SQLFloat) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return sf, nil
}

// Float64 returns the floating-point value of this SQLValue.
func (sf SQLFloat) Float64() float64 {
	return float64(sf)
}

// Int64 returns the integer value of this SQLValue.
func (sf SQLFloat) Int64() int64 {
	return int64(sf)
}

// MySQLEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (sf SQLFloat) MySQLEncode(*collation.Charset, int) ([]byte, error) {
	return strconv.AppendFloat(nil, float64(sf), 'f', -1, 64), nil
}

// Size returns the size of this SQLValue in bytes.
func (sf SQLFloat) Size() uint64 {
	return 8
}

func (sf SQLFloat) String() string {
	return strconv.FormatFloat(float64(sf), 'f', -1, 64)
}

// ToAggregationLanguage translates SQLFloat into something that can
// be used in an aggregation pipeline. If SQLFloat cannot be translated,
// it will return nil and false.
func (sf SQLFloat) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	return wrapInLiteral(sf.Value()), true
}

// Type returns the SQLType of this SQLValue.
func (SQLFloat) Type() schema.SQLType {
	return schema.SQLFloat
}

// Uint64 returns the unsigned integer value of this SQLValue.
func (sf SQLFloat) Uint64() uint64 {
	return uint64(sf)
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (sf SQLFloat) Value() interface{} {
	return float64(sf)
}

// SQLInt represents a 64-bit integer value.
type SQLInt int64

// Decimal128 returns the decimal value of this SQLValue.
func (si SQLInt) Decimal128() decimal.Decimal {
	d, _ := decimal.NewFromString(si.String())
	return d
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (si SQLInt) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return si, nil
}

// Float64 returns the floating-point value of this SQLValue.
func (si SQLInt) Float64() float64 {
	return float64(si)
}

// Int64 returns the integer value of this SQLValue.
func (si SQLInt) Int64() int64 {
	return int64(si)
}

// MySQLEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (si SQLInt) MySQLEncode(*collation.Charset, int) ([]byte, error) {
	return strconv.AppendInt(nil, int64(si), 10), nil
}

// Size returns the size of this SQLValue in bytes.
func (si SQLInt) Size() uint64 {
	return 8
}

func (si SQLInt) String() string {
	return strconv.FormatInt(si.Int64(), 10)
}

// ToAggregationLanguage translates SQLInt into something that can
// be used in an aggregation pipeline. If SQLInt cannot be translated,
// it will return nil and false.
func (si SQLInt) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	return wrapInLiteral(si.Value()), true
}

// Type returns the SQLType of this SQLValue.
func (SQLInt) Type() schema.SQLType {
	return schema.SQLInt
}

// Uint64 returns the unsigned integer value of this SQLValue.
func (si SQLInt) Uint64() uint64 {
	return uint64(si)
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (si SQLInt) Value() interface{} {
	return int64(si)
}

// SQLNoValue represents no value.
type SQLNoValue struct{}

// SQLNone is a constant SQLNoValue.
var SQLNone = SQLNoValue{}

// Decimal128 returns the decimal value of this SQLValue.
func (SQLNoValue) Decimal128() decimal.Decimal {
	return decimal.Zero
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (sn SQLNoValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sn, nil
}

// Float64 returns the floating-point value of this SQLValue.
func (SQLNoValue) Float64() float64 {
	return float64(0)
}

// Int64 returns the integer value of this SQLValue.
func (SQLNoValue) Int64() int64 {
	return int64(0)
}

// MySQLEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (sn SQLNoValue) MySQLEncode(*collation.Charset, int) ([]byte, error) {
	return nil, mysqlerrors.Unknownf("unsupported wire type %T", sn)
}

// Size returns the size of this SQLValue in bytes.
func (SQLNoValue) Size() uint64 {
	return 0
}

func (SQLNoValue) String() string {
	return string(schema.SQLNone)
}

// Type returns the SQLType of this SQLValue.
func (SQLNoValue) Type() schema.SQLType {
	return schema.SQLNone
}

// Uint64 returns the unsigned integer value of this SQLValue.
func (SQLNoValue) Uint64() uint64 {
	return uint64(0)
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (SQLNoValue) Value() interface{} {
	return struct{}{}
}

// SQLNullValue represents a null.
type SQLNullValue struct{}

// SQLNull is a constant SQLNullValue.
var SQLNull = SQLNullValue{}

// Decimal128 returns the decimal value of this SQLValue.
func (SQLNullValue) Decimal128() decimal.Decimal {
	return decimal.Zero
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (nv SQLNullValue) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return nv, nil
}

// Float64 returns the floating-point value of this SQLValue.
func (SQLNullValue) Float64() float64 {
	return float64(0)
}

// Int64 returns the integer value of this SQLValue.
func (SQLNullValue) Int64() int64 {
	return int64(0)
}

// MySQLEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (SQLNullValue) MySQLEncode(*collation.Charset, int) ([]byte, error) {
	return nil, nil
}

// Size returns the size of this SQLValue in bytes.
func (SQLNullValue) Size() uint64 {
	return 0
}

func (SQLNullValue) String() string {
	return string(schema.SQLNull)
}

// ToAggregationLanguage translates SQLNullValue into something that can
// be used in an aggregation pipeline. If SQLNullValue cannot be translated,
// it will return nil and false.
func (SQLNullValue) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	return mgoNullLiteral, true
}

// Type returns the SQLType of this SQLValue.
func (SQLNullValue) Type() schema.SQLType {
	return schema.SQLNull
}

// Uint64 returns the unsigned integer value of this SQLValue.
func (SQLNullValue) Uint64() uint64 {
	return uint64(0)
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (SQLNullValue) Value() interface{} {
	return nil
}

// SQLObjectID represents a MongoDB ObjectID value.
type SQLObjectID string

// Decimal128 returns the decimal value of this SQLValue.
func (SQLObjectID) Decimal128() decimal.Decimal {
	return decimal.Zero
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (id SQLObjectID) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return id, nil
}

// Float64 returns the floating-point value of this SQLValue.
func (SQLObjectID) Float64() float64 {
	return float64(0)
}

// Int64 returns the integer value of this SQLValue.
func (SQLObjectID) Int64() int64 {
	return int64(0)
}

// MySQLEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (id SQLObjectID) MySQLEncode(*collation.Charset, int) ([]byte, error) {
	if ret, ok := id.SQLVarchar().(SQLVarchar); ok {
		return []byte(string(ret)), nil
	}
	// Should be unreachable.
	return nil, mysqlerrors.Unknownf("unformatable ObjectID")
}

// Size returns the size of this SQLValue in bytes.
func (id SQLObjectID) Size() uint64 {
	return 12
}

func (id SQLObjectID) String() string {
	return string(id)
}

// Type returns the SQLType of this SQLValue.
func (id SQLObjectID) Type() schema.SQLType {
	return schema.SQLObjectID
}

// Uint64 returns the unsigned integer value of this SQLValue.
func (SQLObjectID) Uint64() uint64 {
	return uint64(0)
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (id SQLObjectID) Value() interface{} {
	return bson.ObjectIdHex(string(id))
}

// SQLTimestamp represents a timestamp value.
type SQLTimestamp struct {
	Time time.Time
}

// Decimal128 returns the decimal value of this SQLValue.
func (st SQLTimestamp) Decimal128() decimal.Decimal {
	return decimal.NewFromFloat(st.Float64())
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (st SQLTimestamp) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return st, nil
}

// Float64 returns the floating-point value of this SQLValue.
func (st SQLTimestamp) Float64() float64 {
	if st.Time.Year() == 0 {
		val, _ := strconv.ParseFloat(st.Time.Format("150405"), 64)
		return val
	}
	val, _ := strconv.ParseFloat(st.Time.Format("20060102150405"), 64)
	return val
}

// Int64 returns the integer value of this SQLValue.
func (st SQLTimestamp) Int64() int64 {
	if st.Time.Year() == 0 {
		val, _ := strconv.ParseInt(st.Time.Format("150405"), 10, 64)
		return val
	}
	val, _ := strconv.ParseInt(st.Time.Format("20060102150405"), 10, 64)
	return val
}

// MySQLEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (st SQLTimestamp) MySQLEncode(*collation.Charset, int) ([]byte, error) {
	if st.Time == NullDate {
		return []byte("0000-00-00 00:00:00"), nil
	}
	if strings.Contains(st.Time.String(), ".") {
		return util.Slice(st.Time.Format(schema.TimestampFormatMicros)), nil
	}
	return util.Slice(st.Time.Format(schema.TimestampFormat)), nil
}

// Size returns the size of this SQLValue in bytes.
func (st SQLTimestamp) Size() uint64 {
	return 8
}

func (st SQLTimestamp) String() string {
	ms := st.Time.Round(time.Second)
	if ms.Equal(st.Time) {
		return st.Time.Format("2006-01-02 15:04:05")
	}
	return st.Time.Format("2006-01-02 15:04:05.000000")
}

// ToAggregationLanguage translates SQLTimestamp into something that can
// be used in an aggregation pipeline. If SQLTimestamp cannot be translated,
// it will return nil and false.
func (st SQLTimestamp) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	return wrapInLiteral(st.Time), true
}

// Type returns the SQLType of this SQLValue.
func (SQLTimestamp) Type() schema.SQLType {
	return schema.SQLTimestamp
}

// Uint64 returns the unsigned integer value of this SQLValue.
func (st SQLTimestamp) Uint64() uint64 {
	if st.Time.Year() == 0 {
		val, _ := strconv.ParseUint(st.Time.Format("150405"), 10, 64)
		return val
	}
	val, _ := strconv.ParseUint(st.Time.Format("20060102150405"), 10, 64)
	return val
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (st SQLTimestamp) Value() interface{} {
	return st.Time
}

// SQLUUID represents a MongoDB UUID value.
type SQLUUID struct {
	kind  schema.MongoType
	bytes []byte
}

// Decimal128 returns the decimal value of this SQLValue.
func (SQLUUID) Decimal128() decimal.Decimal {
	return decimal.Zero
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (uuid SQLUUID) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return uuid, nil
}

// Float64 returns the floating-point value of this SQLValue.
func (SQLUUID) Float64() float64 {
	return float64(0)
}

// Int64 returns the integer value of this SQLValue.
func (SQLUUID) Int64() int64 {
	return int64(0)
}

// Size returns the size of this SQLValue in bytes.
func (uuid SQLUUID) Size() uint64 {
	return 16
}

func (uuid SQLUUID) String() string {
	if uuid, ok := uuid.SQLVarchar().(SQLVarchar); ok {
		return string(uuid)
	}
	// unreachable
	return ""
}

// Type returns the SQLType of this SQLValue.
func (SQLUUID) Type() schema.SQLType {
	return schema.SQLUUID
}

// Uint64 returns the unsigned integer value of this SQLValue.
func (SQLUUID) Uint64() uint64 {
	return uint64(0)
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (uuid SQLUUID) Value() interface{} {
	return uuid.bytes
}

// ToAggregationLanguage translates SQLUUID into something that can
// be used in an aggregation pipeline. If SQLUUID cannot be translated,
// it will return nil and false.
func (uuid SQLUUID) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	value := bson.Binary{Kind: 0x03, Data: uuid.bytes}
	if uuid.kind == schema.MongoUUID {
		value.Kind = 0x04
	}
	return value, true
}

// MySQLEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (uuid SQLUUID) MySQLEncode(*collation.Charset, int) ([]byte, error) {
	return []byte(uuid.String()), nil
}

// SQLUint32 represents an unsigned 32-bit integer.
type SQLUint32 uint32

// Decimal128 returns the decimal value of this SQLValue.
func (su SQLUint32) Decimal128() decimal.Decimal {
	d, _ := decimal.NewFromString(su.String())
	return d
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (su SQLUint32) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return su, nil
}

// Float64 returns the floating-point value of this SQLValue.
func (su SQLUint32) Float64() float64 {
	return float64(su)
}

// Int64 returns the integer value of this SQLValue.
func (su SQLUint32) Int64() int64 {
	return int64(su)
}

// Size returns the size of this SQLValue in bytes.
func (su SQLUint32) Size() uint64 {
	return 4
}

func (su SQLUint32) String() string {
	return strconv.FormatInt(su.Int64(), 10)
}

// Type returns the SQLType of this SQLValue.
func (su SQLUint32) Type() schema.SQLType {
	return schema.SQLInt
}

// Uint64 returns the unsigned integer value of this SQLValue.
func (su SQLUint32) Uint64() uint64 {
	return uint64(su)
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (su SQLUint32) Value() interface{} {
	return uint32(su)
}

// ToAggregationLanguage translates SQLUint32 into something that can
// be used in an aggregation pipeline. If SQLUint32 cannot be translated,
// it will return nil and false.
func (su SQLUint32) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	return wrapInLiteral(su.Value()), true
}

// MySQLEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (su SQLUint32) MySQLEncode(*collation.Charset, int) ([]byte, error) {
	return strconv.AppendInt(nil, int64(su), 10), nil
}

// SQLUint64 represents an unsigned 64-bit integer.
type SQLUint64 uint64

// Decimal128 returns the decimal value of this SQLValue.
func (su SQLUint64) Decimal128() decimal.Decimal {
	d, _ := decimal.NewFromString(su.String())
	return d
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (su SQLUint64) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return su, nil
}

// Float64 returns the floating-point value of this SQLValue.
func (su SQLUint64) Float64() float64 {
	return float64(su)
}

// Int64 returns the integer value of this SQLValue.
func (su SQLUint64) Int64() int64 {
	return int64(su)
}

// Size returns the size of this SQLValue in bytes.
func (su SQLUint64) Size() uint64 {
	return 8
}

func (su SQLUint64) String() string {
	return strconv.FormatUint(uint64(su), 10)
}

// Type returns the SQLType of this SQLValue.
func (su SQLUint64) Type() schema.SQLType {
	return schema.SQLUint64
}

// Uint64 returns the unsigned integer value of this SQLValue.
func (su SQLUint64) Uint64() uint64 {
	return uint64(su)
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (su SQLUint64) Value() interface{} {
	return uint64(su)
}

// ToAggregationLanguage translates SQLUint64 into something that can
// be used in an aggregation pipeline. If SQLUint64 cannot be translated,
// it will return nil and false.
func (su SQLUint64) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	val, ok := t.getValue(su)
	if !ok {
		return nil, false
	}

	ui := val.(uint64)
	if ui > math.MaxInt64 {
		return nil, false
	}
	return wrapInLiteral(val), true
}

// MySQLEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (su SQLUint64) MySQLEncode(*collation.Charset, int) ([]byte, error) {
	return strconv.AppendInt(nil, int64(su), 10), nil
}

// SQLValues represents multiple sql values.
type SQLValues struct {
	Values []SQLValue
}

// Decimal128 returns the decimal value of this SQLValue.
func (sv *SQLValues) Decimal128() decimal.Decimal {
	d, _ := decimal.NewFromString(sv.String())
	return d
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (sv *SQLValues) Evaluate(ctx *EvalCtx) (SQLValue, error) {
	return sv, nil
}

// Float64 returns the floating-point value of this SQLValue.
func (sv *SQLValues) Float64() float64 {
	return sv.Values[0].Float64()
}

// Int64 returns the integer value of this SQLValue.
func (sv *SQLValues) Int64() int64 {
	return int64(sv.Values[0].Float64())
}

// Normalize will attempt to change SQLValues into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (sv *SQLValues) Normalize() Node {
	if len(sv.Values) == 1 {
		return sv.Values[0]
	}

	return sv
}

// Size returns the size of this SQLValue in bytes.
func (sv *SQLValues) Size() uint64 {
	s := uint64(0)
	for _, v := range sv.Values {
		s += v.Size()
	}

	return s
}

func (sv *SQLValues) String() string {
	var values []string
	for _, n := range sv.Values {
		values = append(values, n.String())
	}
	return strings.Join(values, ", ")
}

// ToAggregationLanguage translates SQLValues into something that can
// be used in an aggregation pipeline. If SQLValues cannot be translated,
// it will return nil and false.
func (sv *SQLValues) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	var transExprs []interface{}

	for _, expr := range sv.Values {
		transExpr, ok := t.ToAggregationLanguage(expr)
		if !ok {
			return nil, false
		}
		transExprs = append(transExprs, transExpr)
	}

	return transExprs, true
}

// Type returns the SQLType of this SQLValue.
func (sv *SQLValues) Type() schema.SQLType {
	if len(sv.Values) == 1 {
		return sv.Values[0].Type()
	} else if len(sv.Values) == 0 {
		return schema.SQLNone
	}

	return schema.SQLTuple
}

// Uint64 returns the unsigned integer value of this SQLValue.
func (sv *SQLValues) Uint64() uint64 {
	return sv.Values[0].Uint64()
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (sv *SQLValues) Value() interface{} {
	values := []interface{}{}
	for _, v := range sv.Values {
		values = append(values, v.Value())
	}
	return values
}

// MySQLEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (sv *SQLValues) MySQLEncode(*collation.Charset, int) ([]byte, error) {
	return nil, mysqlerrors.Unknownf("unsupported type %T for wire protocol", sv)
}

// SQLVarchar represents a string value.
type SQLVarchar string

// Decimal128 returns the decimal value of this SQLValue.
func (sv SQLVarchar) Decimal128() decimal.Decimal {
	d, err := decimal.NewFromString(sv.String())
	if err != nil {
		return decimal.Zero
	}
	return d
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (sv SQLVarchar) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return sv, nil
}

// Float64 returns the floating-point value of this SQLValue.
func (sv SQLVarchar) Float64() float64 {
	val, _ := strconv.ParseFloat(string(sv), 64)
	return val
}

// Int64 returns the integer value of this SQLValue.
func (sv SQLVarchar) Int64() int64 {
	val, _ := strconv.ParseInt(string(sv), 10, 64)
	return val
}

// MySQLEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (sv SQLVarchar) MySQLEncode(charSet *collation.Charset,
	mongoDBVarcharLength int) ([]byte,
	error) {
	b := []byte(sv)
	if string(charSet.Name) == "utf8" {
		return b, nil
	}
	ret := charSet.Encode(b)
	// Varchars are counted by characters, not b. Use runes to
	// account for multi-byte characters. Since we know the number
	// of characters can't be more than the number of b, we can
	// skip the character length check if the byte length is satisfactory.
	if mongoDBVarcharLength != 0 && len(ret) > mongoDBVarcharLength {
		runes := []rune(string(ret))
		if len(runes) > mongoDBVarcharLength {
			runes = runes[:mongoDBVarcharLength]
			ret = []byte(string(runes))
		}
	}
	return ret, nil
}

// Size returns the size of this SQLValue in bytes.
func (sv SQLVarchar) Size() uint64 {
	return uint64(len(sv))
}

func (sv SQLVarchar) String() string {
	return string(sv)
}

// ToAggregationLanguage translates SQLVarchar into something that can
// be used in an aggregation pipeline. If SQLVarchar cannot be translated,
// it will return nil and false.
func (sv SQLVarchar) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	return wrapInLiteral(sv.Value()), true
}

// Type returns the SQLType of this SQLValue.
func (SQLVarchar) Type() schema.SQLType {
	return schema.SQLVarchar
}

// Uint64 returns the unsigned integer value of this SQLValue.
func (sv SQLVarchar) Uint64() uint64 {
	val, _ := strconv.ParseUint(string(sv), 10, 64)
	return val
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (sv SQLVarchar) Value() interface{} {
	return string(sv)
}

// NewSQLBool returns a SQLBool representation of the provided boolean literal.
func NewSQLBool(b bool) SQLBool {
	if b {
		return SQLTrue
	}
	return SQLFalse
}

// CompareTo compares two SQLValues. It returns -1 if
// left compares less than right; 1, if left compares
// greater than right; and 0 if left compares equal to
// right.
func CompareTo(left, right SQLValue, collation *collation.Collation) (int, error) {
	switch leftVal := left.(type) {
	case *SQLValues:
		err := fmt.Errorf("operand should contain %v columns", len(leftVal.Values))

		rightVal, ok := right.(*SQLValues)
		if !ok {
			// This allows for comparisons such as:
			// `select a, b from foo where (a) < 3`
			if len(leftVal.Values) != 1 {
				return -1, err
			}
			rightVal = &SQLValues{[]SQLValue{right}}
		} else if len(leftVal.Values) != len(rightVal.Values) {
			return -1, err
		}

		for i := 0; i < len(leftVal.Values); i++ {
			_, noLeft := leftVal.Values[i].(SQLNoValue)
			_, noRight := rightVal.Values[i].(SQLNoValue)

			if noLeft && !noRight {
				return -1, nil
			}

			if !noLeft && noRight {
				return 1, nil
			}

			if noLeft && noRight {
				return 0, nil
			}

			c, err := CompareTo(leftVal.Values[i], rightVal.Values[i], collation)
			if err != nil {
				return c, err
			}

			if c != 0 {
				return c, nil
			}
		}
		return 0, nil
	default:
		switch right.(type) {
		case *SQLValues:
			i, err := CompareTo(right, left, collation)
			if err != nil {
				return i, err
			}
			return -i, nil
		}
	}

	if left.Type() == right.Type() {
		switch leftVal := left.(type) {
		case SQLDate, SQLDecimal128, SQLFloat, SQLInt, SQLUint32, SQLUint64, SQLTimestamp:
			return compareDecimal128(left.Decimal128(), right.Decimal128())
		case SQLVarchar:
			rightVal, _ := right.(SQLVarchar)
			return collation.CompareString(string(leftVal), string(rightVal)), nil
		case SQLObjectID:
			rightVal, _ := right.(SQLObjectID)
			if !bson.IsObjectIdHex(leftVal.String()) {
				return -1, fmt.Errorf("%v is not a valid ObjectID", leftVal.String())
			}
			if !bson.IsObjectIdHex(rightVal.String()) {
				return -1, fmt.Errorf("%v is not a valid ObjectID", rightVal.String())
			}

			s1 := []byte(leftVal.String())
			s2 := []byte(rightVal.String())
			return compareBytes(s1, s2)
		case SQLNullValue:
			return 0, nil
		case SQLUUID:
			rightVal, ok := right.(SQLUUID)
			if !ok {
				return -1, fmt.Errorf("%v is not a valid UUID", right.String())
			}
			return compareBytes(leftVal.bytes, rightVal.bytes)
		}
	}

	// Different types
	switch lVal := left.(type) {
	case SQLNullValue:
		switch right.(type) {
		case *SQLValues:
			i, err := CompareTo(right, left, collation)
			if err != nil {
				return i, err
			}
			return -i, nil
		default:
			return -1, nil
		}
	case SQLVarchar:
		switch right.(type) {
		case SQLDate, SQLTimestamp:
			// MySQL throws an error if you try to compare varchar =,<,> date/timestamp.
			// It works the other way around, however (i.e. date/timestamp =,<,> varchar).
			return -1, fmt.Errorf("Illegal mix of collations %T and %T", left, right)
		case SQLNullValue:
			return 1, nil
		default:
			return compareDecimal128(left.Decimal128(), right.Decimal128())
		}
	case SQLDate:
		switch rVal := right.(type) {
		case SQLVarchar:
			t, _, ok := parseDateTime(right.String())
			if !ok {
				t, _, _ = parseDateTime("0001-01-01")
			}
			return compareFloats(left.Float64(), SQLDate{Time: t}.Float64())
		case SQLTimestamp:
			if rVal.Time.Before(lVal.Time) {
				return 1, nil
			} else if rVal.Time.After(lVal.Time) {
				return -1, nil
			}
			return 0, nil
		case SQLNullValue:
			return 1, nil
		default:
			return compareDecimal128(left.Decimal128(), right.Decimal128())
		}
	case SQLTimestamp:
		switch rVal := right.(type) {
		case SQLVarchar:
			t, _, ok := parseDateTime(right.String())
			if !ok {
				t, _, _ = parseDateTime("0001-01-01 00:00:00")
			}
			return compareFloats(left.Float64(), SQLTimestamp{Time: t}.Float64())
		case SQLNullValue:
			return 1, nil
		case SQLDate:
			if rVal.Time.Before(lVal.Time) {
				return 1, nil
			} else if rVal.Time.After(lVal.Time) {
				return -1, nil
			}
			return 0, nil
		default:
			return compareDecimal128(left.Decimal128(), right.Decimal128())
		}
	case SQLUUID:
		switch right.(type) {
		case SQLVarchar:
			uuid, _ := GetBinaryFromExpr(schema.MongoUUID, right)
			return compareBytes(lVal.bytes, uuid.Data)
		default:
			return compareDecimal128(left.Decimal128(), right.Decimal128())
		}

	default:
		switch right.(type) {
		case SQLNullValue:
			return 1, nil
		default:
			return compareDecimal128(left.Decimal128(), right.Decimal128())
		}
	}
}

// BSONValueToSQLValue deserializes raw BSON into SQLTypes directly.
func BSONValueToSQLValue(valueType, bsonSpecType, uuidSubtype schema.BSONSpecType,
	data []byte, fieldName string) (SQLValue, error) {
	switch bsonSpecType {
	case schema.BSONBoolean:
		if data[0] == 0x0 {
			return SQLFalse, nil
		}
		return SQLTrue, nil
	case schema.BSONDecimal128:
		h := (uint64(data[0]) << 0) |
			(uint64(data[1]) << 8) |
			(uint64(data[2]) << 16) |
			(uint64(data[3]) << 24) |
			(uint64(data[4]) << 32) |
			(uint64(data[5]) << 40) |
			(uint64(data[6]) << 48) |
			(uint64(data[7]) << 56)
		l := (uint64(data[8]) << 0) |
			(uint64(data[9]) << 8) |
			(uint64(data[10]) << 16) |
			(uint64(data[11]) << 24) |
			(uint64(data[12]) << 32) |
			(uint64(data[13]) << 40) |
			(uint64(data[14]) << 48) |
			(uint64(data[15]) << 56)
		bd := NewBSONDecimal128(l, h)
		gd, err := decimal.NewFromString(bd.String())
		if err != nil {
			return nil, err
		}
		return SQLDecimal128(gd), nil
	case schema.BSONDouble:
		ret := math.Float64frombits((uint64(data[0]) << 0) |
			(uint64(data[1]) << 8) |
			(uint64(data[2]) << 16) |
			(uint64(data[3]) << 24) |
			(uint64(data[4]) << 32) |
			(uint64(data[5]) << 40) |
			(uint64(data[6]) << 48) |
			(uint64(data[7]) << 56))
		return SQLFloat(ret), nil
	case schema.BSONInt: //32 bit int
		ret := int32((uint32(data[0]) << 0) |
			(uint32(data[1]) << 8) |
			(uint32(data[2]) << 16) |
			(uint32(data[3]) << 24))
		return SQLInt(ret), nil
	case schema.BSONInt64: // 64 bit int
		ret := int64((uint64(data[0]) << 0) |
			(uint64(data[1]) << 8) |
			(uint64(data[2]) << 16) |
			(uint64(data[3]) << 24) |
			(uint64(data[4]) << 32) |
			(uint64(data[5]) << 40) |
			(uint64(data[6]) << 48) |
			(uint64(data[7]) << 56))
		return SQLInt(ret), nil
	case schema.BSONObjectID:
		return SQLObjectID(hex.EncodeToString(data)), nil
	case schema.BSONNull:
		return SQLNull, nil
	case schema.BSONString:
		l := ((uint32(data[0]) << 0) |
			(uint32(data[1]) << 8) |
			(uint32(data[2]) << 16) |
			(uint32(data[3]) << 24)) - 1
		if data[len(data)-1] != '\x00' {
			return nil, fmt.Errorf("corrupted string field: not 0x00 terminated")
		}
		if len(data) != int(l)+5 {
			return nil, fmt.Errorf("corrupted string field: length mismatch")
		}
		data = data[4 : len(data)-1]
		if !utf8.Valid(data) {
			return nil, fmt.Errorf("corrupted string field: not valid unicode")
		}
		return SQLVarchar(data), nil
	case schema.BSONDatetime: // Date
		i := int64((uint64(data[0]) << 0) |
			(uint64(data[1]) << 8) |
			(uint64(data[2]) << 16) |
			(uint64(data[3]) << 24) |
			(uint64(data[4]) << 32) |
			(uint64(data[5]) << 40) |
			(uint64(data[6]) << 48) |
			(uint64(data[7]) << 56))
		var t time.Time
		if i == -62135596800000 {
			t = time.Time{}.In(schema.DefaultLocale)
		} else {
			t = time.Unix(i/1e3, i%1e3*1e6).In(schema.DefaultLocale)
		}
		return SQLTimestamp{Time: t}, nil
	case schema.BSONUUID:
		l := ((uint32(data[0]) << 0) |
			(uint32(data[1]) << 8) |
			(uint32(data[2]) << 16) |
			(uint32(data[3]) << 24))
		subType := data[4]
		data = data[5:]
		if len(data) != int(l) {
			return nil, fmt.Errorf("corrupted binary field")
		}
		if subType == 0x04 {
			return SQLUUID{kind: schema.MongoUUID, bytes: data}, nil
		} else if subType == 0x03 {
			if uuidSubtype == schema.BSONJavaUUID {
				reverseByteArray(data, 0, 8)
				reverseByteArray(data, 8, 8)

			} else if uuidSubtype == schema.BSONCSharpUUID {
				reverseByteArray(data, 0, 4)
				reverseByteArray(data, 4, 2)
				reverseByteArray(data, 6, 2)
			}
			return SQLUUID{kind: schema.MongoUUID, bytes: data}, nil
		}
		return nil,
			fmt.Errorf("UUID types 0x3 and 0x4 are the only supported binary subtybes, not %#02x",
				subType)
	default:
		readableBsonType := string(bsonSpecType)
		readableValueType := string(valueType)
		if val, ok := schema.BSONTypeToMongoType[bsonSpecType]; ok {
			readableBsonType = string(val)
		}
		if val, ok := schema.BSONTypeToMongoType[valueType]; ok {
			readableValueType = string(val)
		}
		return nil, fmt.Errorf("unexpected bson type: found %s but expected %s for field %s",
			readableBsonType, readableValueType, fieldName)
	}
}

// BSONDecimal128 holds decimal128 BSON values.
type BSONDecimal128 struct {
	h, l uint64
}

// NewBSONDecimal128 is a constructor for BSONDecimal128.
func NewBSONDecimal128(h, l uint64) BSONDecimal128 {
	return BSONDecimal128{h, l}
}

// String() formats a BSONDecimal128 as a string.
func (d BSONDecimal128) String() string {
	var pos int     // positive sign
	var e int       // exponent
	var h, l uint64 // significand high/low

	if d.h>>63&1 == 0 {
		pos = 1
	}

	switch d.h >> 58 & (1<<5 - 1) {
	case 0x1F:
		return "NaN"
	case 0x1E:
		return "-Inf"[pos:]
	}

	l = d.l
	if d.h>>61&3 == 3 {
		// Bits: 1*sign 2*ignored 14*exponent 111*significand.
		// Implicit 0b100 prefix in significand.
		e = int(d.h>>47&(1<<14-1)) - 6176
		//h = 4<<47 | d.h&(1<<47-1)
		// Spec says all of these values are out of range.
		h, l = 0, 0
	} else {
		// Bits: 1*sign 14*exponent 113*significand
		e = int(d.h>>49&(1<<14-1)) - 6176
		h = d.h & (1<<49 - 1)
	}

	// Would be handled by the logic below, but that's trivial and common.
	if h == 0 && l == 0 && e == 0 {
		return "-0"[pos:]
	}

	var repr [48]byte // Loop 5 times over 9 digits plus dot, negative sign, and leading zero.
	var last = len(repr)
	var i = len(repr)
	var dot = len(repr) + e
	var rem uint32
Loop:
	for d9 := 0; d9 < 5; d9++ {
		h, l, rem = divmod(h, l, 1e9)
		for d1 := 0; d1 < 9; d1++ {
			// Handle "-0.0", "0.00123400", "-1.00E-6", "1.050E+3", etc.
			if i < len(repr) &&
				(dot == i || (l == 0 &&
					h == 0 &&
					rem > 0 &&
					rem < 10 &&
					(dot < i-6 || e > 0))) {
				e += len(repr) - i
				i--
				repr[i] = '.'
				last = i - 1
				dot = len(repr) // Unmark.
			}
			c := '0' + byte(rem%10)
			rem /= 10
			i--
			repr[i] = c
			// Handle "0E+3", "1E+3", etc.
			if l == 0 && h == 0 && rem == 0 && i == len(repr)-1 && (dot < i-5 || e > 0) {
				last = i
				break Loop
			}
			if c != '0' {
				last = i
			}
			// Break early. Works without it, but why.
			if dot > i && l == 0 && h == 0 && rem == 0 {
				break Loop
			}
		}
	}
	repr[last-1] = '-'
	last--

	if e > 0 {
		return string(repr[last+pos:]) + "E+" + strconv.Itoa(e)
	}
	if e < 0 {
		return string(repr[last+pos:]) + "E" + strconv.Itoa(e)
	}
	return string(repr[last+pos:])
}

// divmod is a helper function for BSONDecimal128s that preforms
// both division and remainder efficiently.
// nolint: unparam
func divmod(h, l uint64, div uint32) (qh, ql uint64, rem uint32) {
	div64 := uint64(div)
	a := h >> 32
	aq := a / div64
	ar := a % div64
	b := ar<<32 + h&(1<<32-1)
	bq := b / div64
	br := b % div64
	c := br<<32 + l>>32
	cq := c / div64
	cr := c % div64
	d := cr<<32 + l&(1<<32-1)
	dq := d / div64
	dr := d % div64
	return (aq<<32 | bq), (cq<<32 | dq), uint32(dr)
}
