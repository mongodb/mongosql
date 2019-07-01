package values

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/internal/dateutil"
	"github.com/10gen/sqlproxy/internal/mathutil"
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/schema"

	"github.com/shopspring/decimal"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// BaseSQLBool represents a boolean. BaseSQLBool should be treated as an
// abstract type; it should never be used directly during query evaluation, and
// instead should be embedded in other SQLValue implementers who can override
// BaseSQLBool's default implementations as needed.
type BaseSQLBool struct {
	val  float64
	kind SQLValueKind
}

// iSQLBool must be implemented to satisfy the SQLBool interface.
func (BaseSQLBool) iSQLBool() {}

func newBaseSQLBool(kind SQLValueKind, val bool) BaseSQLBool {
	var num float64
	if val {
		num = 1
	}
	return BaseSQLBool{
		val:  num,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLBool) IsNull() bool {
	return false
}

// CloneWithKind copies the SQLValue with the passed kind.
func (s BaseSQLBool) CloneWithKind(kind SQLValueKind) SQLValue {
	kind.AssertValid()
	return NewSQLBool(kind, s.val != 0.0)
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLBool) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLBool) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	if s.val == 0 {
		return []byte{48}, nil
	}
	return []byte{49}, nil
}

// Size returns the size of this SQLValue in bytes.
func (s BaseSQLBool) Size() uint64 {
	return 1
}

// String returns the string representation of this SQLValue.
// String should return the same value regardless of the SQLValue's kind, and
// should not be overridden by any embedding SQLValue implementers.
func (s BaseSQLBool) String() string {
	return s.varchar()
}

func (s BaseSQLBool) varchar() string {
	return strconv.FormatFloat(Float64(s), 'f', -1, 64)
}

// EvalType returns the EvalType of this SQLValue.
func (BaseSQLBool) EvalType() types.EvalType {
	return types.EvalBoolean
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLBool) Value() interface{} {
	return s.val != 0
}

// BSONValue returns a bsoncore.Value that represents the literal value of this SQLValue.
func (s BaseSQLBool) BSONValue() (bsoncore.Value, error) {
	return bsoncore.Value{
		Type: bsontype.Boolean,
		Data: bsoncore.AppendBoolean(nil, s.val != 0),
	}, nil
}

// SQLBool converts the BaseSQLBool receiver, s, to a SQLBool.
func (s BaseSQLBool) SQLBool() SQLBool {
	return NewSQLBool(s.kind, !(s.val == 0))
}

// SQLDate converts the BaseSQLBool receiver, s, to a SQLDate.
func (s BaseSQLBool) SQLDate() SQLDate {
	if s.val == 0 {
		return NewSQLDate(s.kind, NullDate)
	}
	return NewSQLNull(s.kind)
}

// SQLDecimal128 converts the BaseSQLBool receiver, s, to a SQLDecimal128.
func (s BaseSQLBool) SQLDecimal128() SQLDecimal128 {
	if s.val == 0 {
		return NewSQLDecimal128(s.kind, decimal.Zero)
	}
	return NewSQLDecimal128(s.kind, decimal.New(1, 0))
}

// SQLFloat converts the BaseSQLBool receiver, s, to a SQLFloat.
func (s BaseSQLBool) SQLFloat() SQLFloat {
	if s.val == 0 {
		return NewSQLFloat(s.kind, 0.0)
	}
	return NewSQLFloat(s.kind, 1.0)
}

// SQLInt converts the BaseSQLBool receiver, s, to a SQLInt.
func (s BaseSQLBool) SQLInt() SQLInt64 {
	if s.val == 0 {
		return NewSQLInt64(s.kind, 0)
	}
	return NewSQLInt64(s.kind, 1)
}

// SQLObjectID converts the BaseSQLBool receiver, s, to a SQLObjectID.
func (s BaseSQLBool) SQLObjectID() SQLObjectID {
	return NewSQLNull(s.kind)
}

// SQLTimestamp converts the BaseSQLBool receiver, s, to a SQLTimestamp.
func (s BaseSQLBool) SQLTimestamp() SQLTimestamp {
	if s.val == 0 {
		return NewSQLTimestamp(s.kind, NullDate)
	}
	return NewSQLNull(s.kind)
}

// SQLUint converts the BaseSQLBool receiver, s, to a SQLUint64.
func (s BaseSQLBool) SQLUint() SQLUint64 {
	if s.val == 0 {
		return NewSQLUint64(s.kind, 0)
	}
	return NewSQLUint64(s.kind, 1)
}

// SQLVarchar converts the BaseSQLBool receiver, s, to a SQLVarchar.
func (s BaseSQLBool) SQLVarchar() SQLVarchar {
	if s.val == 0 {
		return NewSQLVarchar(s.kind, "0")
	}
	return NewSQLVarchar(s.kind, "1")
}

// BaseSQLDate represents a date. BaseSQLDate should be treated as an abstract
// type; it should never be used directly during query evaluation, and instead
// should be embedded in other SQLValue implementers who can override
// BaseSQLDate's default implementations as needed.
type BaseSQLDate struct {
	datetime time.Time
	kind     SQLValueKind
}

// iSQLDate must be implemented to satisfy the SQLDate interface.
func (BaseSQLDate) iSQLDate() {}

func newBaseSQLDate(kind SQLValueKind, val time.Time) BaseSQLDate {
	t := time.Date(val.Year(), val.Month(), val.Day(), 0, 0, 0, 0, schema.DefaultLocale)
	return BaseSQLDate{
		datetime: t,
		kind:     kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLDate) IsNull() bool {
	return false
}

// CloneWithKind copies the SQLValue with the passed kind.
func (s BaseSQLDate) CloneWithKind(kind SQLValueKind) SQLValue {
	kind.AssertValid()
	return NewSQLDate(kind, s.datetime)
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLDate) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLDate) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	if s.datetime == NullDate {
		return []byte("0000-00-00"), nil
	}
	return strutil.Slice(s.datetime.Format(schema.DateFormat)), nil
}

// Size returns the size of this SQLValue in bytes.
func (s BaseSQLDate) Size() uint64 {
	return 8
}

// String returns the string representation of this SQLValue.
// String should return the same value regardless of the SQLValue's kind, and
// should not be overridden by any embedding SQLValue implementers.
func (s BaseSQLDate) String() string {
	return s.varchar()
}

func (s BaseSQLDate) varchar() string {
	return s.datetime.Format("2006-01-02")
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLDate) EvalType() types.EvalType {
	return types.EvalDate
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLDate) Value() interface{} {
	return s.datetime
}

// BSONValue returns a bsoncore.Value that represents the literal value of this SQLValue.
func (s BaseSQLDate) BSONValue() (bsoncore.Value, error) {
	return bsoncore.Value{
		Type: bsontype.DateTime,
		Data: bsoncore.AppendDateTime(nil, dateutil.UnixMillis(s.datetime)),
	}, nil
}

// SQLBool converts the BaseSQLDate receiver, s, to a SQLBool.
func (s BaseSQLDate) SQLBool() SQLBool {
	t := s.datetime
	if t == NullDate {
		return NewSQLBool(s.kind, false)
	}
	return NewSQLBool(s.kind, true)
}

// SQLDecimal128 converts the BaseSQLDate receiver, s, to a SQLDecimal128.
func (s BaseSQLDate) SQLDecimal128() SQLDecimal128 {
	flt := Float64(s)
	dec := decimal.NewFromFloat(flt)
	return NewSQLDecimal128(s.kind, dec)
}

// SQLDate converts the BaseSQLDate receiver, s, to a SQLDate.
func (s BaseSQLDate) SQLDate() SQLDate {
	return NewSQLDate(s.kind, s.datetime)
}

// SQLFloat converts the BaseSQLDate receiver, s, to a SQLFloat.
func (s BaseSQLDate) SQLFloat() SQLFloat {
	epochMs := dateutil.UnixMillis(s.datetime)
	return NewSQLFloat(s.kind, float64(epochMs))
}

// SQLInt converts the BaseSQLDate receiver, s, to a SQLInt.
func (s BaseSQLDate) SQLInt() SQLInt64 {
	epochMs := dateutil.UnixMillis(s.datetime)
	return NewSQLInt64(s.kind, epochMs)
}

// SQLObjectID converts the BaseSQLDate receiver, s, to a SQLObjectID.
func (s BaseSQLDate) SQLObjectID() SQLObjectID {
	return NewSQLNull(s.kind)
}

// SQLTimestamp converts the BaseSQLDate receiver, s, to a SQLTimestamp.
func (s BaseSQLDate) SQLTimestamp() SQLTimestamp {
	t := s.datetime
	return NewSQLTimestamp(s.kind, t)
}

// SQLUint converts the BaseSQLDate receiver, s, to a SQLUint64.
func (s BaseSQLDate) SQLUint() SQLUint64 {
	epochMs := dateutil.UnixMillis(s.datetime)
	return NewSQLUint64(s.kind, uint64(epochMs))
}

// SQLVarchar converts the BaseSQLDate receiver, s, to a SQLVarchar.
func (s BaseSQLDate) SQLVarchar() SQLVarchar {
	return NewSQLVarchar(s.kind, s.varchar())
}

// BaseSQLDecimal128 represents a decimal 128 value. BaseSQLDecimal128 should be
// treated as an abstract type; it should never be used directly during query
// evaluation, and instead should be embedded in other SQLValue implementers who
// can override BaseSQLDecimal128's default implementations as needed.
type BaseSQLDecimal128 struct {
	val  decimal.Decimal
	kind SQLValueKind
}

func (BaseSQLDecimal128) iNumber() {}

// iSQLDecimal128 must be implemented to satisfy the SQLDecimal128 interface.
func (BaseSQLDecimal128) iSQLDecimal128() {}

func newBaseSQLDecimal128(kind SQLValueKind, val decimal.Decimal) BaseSQLDecimal128 {
	return BaseSQLDecimal128{
		val:  val,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLDecimal128) IsNull() bool {
	return false
}

// CloneWithKind copies the SQLValue with the passed kind.
func (s BaseSQLDecimal128) CloneWithKind(kind SQLValueKind) SQLValue {
	kind.AssertValid()
	return NewSQLDecimal128(kind, s.val)
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLDecimal128) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLDecimal128) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	return []byte(strutil.FormatDecimal(s.val)), nil
}

// Size returns the size of this SQLValue in bytes.
func (s BaseSQLDecimal128) Size() uint64 {
	return 16
}

// String returns the string representation of this SQLValue.
// String should return the same value regardless of the SQLValue's kind, and
// should not be overridden by any embedding SQLValue implementers.
func (s BaseSQLDecimal128) String() string {
	return s.varchar()
}

func (s BaseSQLDecimal128) varchar() string {
	return s.val.String()
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLDecimal128) EvalType() types.EvalType {
	return types.EvalDecimal128
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLDecimal128) Value() interface{} {
	return s.val
}

// BSONValue returns a bsoncore.Value that represents the literal value of this SQLValue.
func (s BaseSQLDecimal128) BSONValue() (bsoncore.Value, error) {
	parsed, err := primitive.ParseDecimal128(s.String())
	if err != nil {
		return bsoncore.Value{}, fmt.Errorf("failed to Decimal128 from SQLValue string: %v", err)
	}
	return bsoncore.Value{
		Type: bsontype.Decimal128,
		Data: bsoncore.AppendDecimal128(nil, parsed),
	}, nil
}

// SQLBool converts the BaseSQLDecimal128 receiver, s, to a SQLBool.
func (s BaseSQLDecimal128) SQLBool() SQLBool {
	return s.SQLInt().SQLBool()
}

// SQLDate converts the BaseSQLDecimal128 receiver, s, to a SQLDate.
func (s BaseSQLDecimal128) SQLDate() SQLDate {
	i := s.val.IntPart()
	sec := i / 1000
	nsec := (i % 1000) * 1000000
	t := time.Unix(sec, nsec)
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the BaseSQLDecimal128 receiver, s, to a SQLDecimal128.
func (s BaseSQLDecimal128) SQLDecimal128() SQLDecimal128 {
	return NewSQLDecimal128(s.kind, s.val)
}

// SQLFloat converts the BaseSQLDecimal128 receiver, s, to a SQLFloat.
func (s BaseSQLDecimal128) SQLFloat() SQLFloat {
	// Second return value tells us if this is exact, we don't care.
	f, _ := s.val.Float64()
	return NewSQLFloat(s.kind, f)
}

// SQLInt converts the BaseSQLDecimal128 receiver, s, to a SQLInt.
func (s BaseSQLDecimal128) SQLInt() SQLInt64 {
	// Do not care if this is exact.
	f, _ := s.val.Float64()
	return NewSQLInt64(s.kind, mathutil.Round(f))
}

// SQLObjectID converts the BaseSQLDecimal128 receiver, s, to a SQLObjectID.
func (s BaseSQLDecimal128) SQLObjectID() SQLObjectID {
	return NewSQLNull(s.kind)
}

// SQLTimestamp converts the BaseSQLDecimal128 receiver, s, to a SQLTimestamp.
func (s BaseSQLDecimal128) SQLTimestamp() SQLTimestamp {
	i := s.val.IntPart()
	sec := i / 1000
	nsec := (i % 1000) * 1000000
	t := time.Unix(sec, nsec)
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(s.kind, t)
}

// SQLUint converts the BaseSQLDecimal128 receiver, s, to a SQLUint64.
func (s BaseSQLDecimal128) SQLUint() SQLUint64 {
	// Do not care if this is exact.
	f, _ := s.val.Float64()
	return NewSQLUint64(s.kind, uint64(mathutil.Round(f)))
}

// SQLVarchar converts the BaseSQLDecimal128 receiver, s, to a SQLVarchar.
func (s BaseSQLDecimal128) SQLVarchar() SQLVarchar {
	return NewSQLVarchar(s.kind, s.val.String())
}

// BaseSQLFloat represents a float. BaseSQLFloat should be treated as an
// abstract type; it should never be used directly during query evaluation, and
// instead should be embedded in other SQLValue implementers who can override
// BaseSQLFloat's default implementations as needed.
type BaseSQLFloat struct {
	val  float64
	kind SQLValueKind
}

func (BaseSQLFloat) iNumber() {}

// iSQLFloat must be implemented to satisfy the SQLFloat interface.
func (BaseSQLFloat) iSQLFloat() {}

func newBaseSQLFloat(kind SQLValueKind, val float64) BaseSQLFloat {
	return BaseSQLFloat{
		val:  val,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLFloat) IsNull() bool {
	return false
}

// CloneWithKind copies the SQLValue with the passed kind.
func (s BaseSQLFloat) CloneWithKind(kind SQLValueKind) SQLValue {
	kind.AssertValid()
	return NewSQLFloat(kind, s.val)
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLFloat) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLFloat) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	return strconv.AppendFloat(nil, s.val, 'f', -1, 64), nil
}

// Size returns the size of this SQLValue in bytes.
func (s BaseSQLFloat) Size() uint64 {
	return 8
}

// String returns the string representation of this SQLValue.
// String should return the same value regardless of the SQLValue's kind, and
// should not be overridden by any embedding SQLValue implementers.
func (s BaseSQLFloat) String() string {
	return s.varchar()
}

func (s BaseSQLFloat) varchar() string {
	return strconv.FormatFloat(s.val, 'f', -1, 64)
}

// SQLBool converts the BaseSQLFloat receiver, s, to a SQLBool.
func (s BaseSQLFloat) SQLBool() SQLBool {
	return s.SQLInt().SQLBool()
}

// SQLDate converts the BaseSQLFloat receiver, s, to a SQLDate.
func (s BaseSQLFloat) SQLDate() SQLDate {
	i := int64(s.val)
	sec := i / 1000
	nsec := (i % 1000) * 1000000
	t := time.Unix(sec, nsec)
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the BaseSQLFloat receiver, s, to a SQLDecimal128.
func (s BaseSQLFloat) SQLDecimal128() SQLDecimal128 {
	return NewSQLDecimal128(s.kind, decimal.NewFromFloat(s.val))
}

// SQLFloat converts the BaseSQLFloat receiver, s, to a SQLFloat.
func (s BaseSQLFloat) SQLFloat() SQLFloat {
	return NewSQLFloat(s.kind, s.val)
}

// SQLInt converts the BaseSQLFloat receiver, s, to a SQLInt.
func (s BaseSQLFloat) SQLInt() SQLInt64 {
	return NewSQLInt64(s.kind, mathutil.Round(s.val))
}

// SQLObjectID converts the BaseSQLFloat receiver, s, to a SQLObjectID.
func (s BaseSQLFloat) SQLObjectID() SQLObjectID {
	return NewSQLNull(s.kind)
}

// SQLTimestamp converts the BaseSQLFloat receiver, s, to a SQLTimestamp.
func (s BaseSQLFloat) SQLTimestamp() SQLTimestamp {
	i := int64(s.val)
	sec := i / 1000
	nsec := (i % 1000) * 1000000
	t := time.Unix(sec, nsec)
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(s.kind, t)

}

// SQLUint converts the BaseSQLFloat receiver, s, to a SQLUint64.
func (s BaseSQLFloat) SQLUint() SQLUint64 {
	return NewSQLUint64(s.kind, uint64(mathutil.Round(s.val)))
}

// SQLVarchar converts the BaseSQLFloat receiver, s, to a SQLVarchar.
func (s BaseSQLFloat) SQLVarchar() SQLVarchar {
	return NewSQLVarchar(s.kind, s.varchar())
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLFloat) EvalType() types.EvalType {
	return types.EvalDouble
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLFloat) Value() interface{} {
	return s.val
}

// BSONValue returns a bsoncore.Value that represents the literal value of this SQLValue.
func (s BaseSQLFloat) BSONValue() (bsoncore.Value, error) {
	return bsoncore.Value{
		Type: bsontype.Double,
		Data: bsoncore.AppendDouble(nil, s.val),
	}, nil
}

// BaseSQLInt64 represents a 64-bit integer value. BaseSQLInt64 should be
// treated as an abstract type; it should never be used directly during query
// evaluation, and instead should be embedded in other SQLValue implementers who
// can override BaseSQLInt64's default implementations as needed.
type BaseSQLInt64 struct {
	val  int64
	kind SQLValueKind
}

func (BaseSQLInt64) iNumber() {}

// iSQLInt64 must be implemented to satisfy the SQLInt64 interface.
func (BaseSQLInt64) iSQLInt64() {}

func newBaseSQLInt64(kind SQLValueKind, val int64) BaseSQLInt64 {
	return BaseSQLInt64{
		val:  val,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLInt64) IsNull() bool {
	return false
}

// CloneWithKind copies the SQLValue with the passed kind.
func (s BaseSQLInt64) CloneWithKind(kind SQLValueKind) SQLValue {
	kind.AssertValid()
	return NewSQLInt64(kind, s.val)
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLInt64) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLInt64) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	return strconv.AppendInt(nil, s.val, 10), nil
}

// Size returns the size of this SQLValue in bytes.
func (s BaseSQLInt64) Size() uint64 {
	return 8
}

// String returns the string representation of this SQLValue.
// String should return the same value regardless of the SQLValue's kind, and
// should not be overridden by any embedding SQLValue implementers.
func (s BaseSQLInt64) String() string {
	return s.varchar()
}

func (s BaseSQLInt64) varchar() string {
	return strconv.FormatInt(Int64(s), 10)
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLInt64) EvalType() types.EvalType {
	return types.EvalInt64
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLInt64) Value() interface{} {
	return s.val
}

// BSONValue returns a bsoncore.Value that represents the literal value of this SQLValue.
func (s BaseSQLInt64) BSONValue() (bsoncore.Value, error) {
	return bsoncore.Value{
		Type: bsontype.Int64,
		Data: bsoncore.AppendInt64(nil, s.val),
	}, nil
}

// SQLBool converts the BaseSQLInt64 receiver, s, to a SQLBool.
func (s BaseSQLInt64) SQLBool() SQLBool {
	if s.val == 0 {
		return NewSQLBool(s.kind, false)
	}
	return NewSQLBool(s.kind, true)
}

// SQLDate converts the BaseSQLInt64 receiver, s, to a SQLDate.
func (s BaseSQLInt64) SQLDate() SQLDate {
	sec := s.val / 1000
	nsec := (s.val % 1000) * 1000000
	t := time.Unix(sec, nsec)
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the BaseSQLInt64 receiver, s, to a SQLDecimal128.
func (s BaseSQLInt64) SQLDecimal128() SQLDecimal128 {
	return NewSQLDecimal128(s.kind, decimal.New(s.val, 0))
}

// SQLFloat converts the BaseSQLInt64 receiver, s, to a SQLFloat.
func (s BaseSQLInt64) SQLFloat() SQLFloat {
	return NewSQLFloat(s.kind, float64(s.val))
}

// SQLInt converts the BaseSQLInt64 receiver, s, to a SQLInt.
func (s BaseSQLInt64) SQLInt() SQLInt64 {
	return NewSQLInt64(s.kind, s.val)
}

// SQLObjectID converts the BaseSQLInt64 receiver, s, to a SQLObjectID.
func (s BaseSQLInt64) SQLObjectID() SQLObjectID {
	return NewSQLNull(s.kind)
}

// SQLTimestamp converts the BaseSQLInt64 receiver, s, to a SQLTimestamp.
func (s BaseSQLInt64) SQLTimestamp() SQLTimestamp {
	sec := s.val / 1000
	nsec := (s.val % 1000) * 1000000
	t := time.Unix(sec, nsec)
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(s.kind, t)
}

// SQLUint converts the BaseSQLUint64 receiver, s, to a SQLUint64.
func (s BaseSQLInt64) SQLUint() SQLUint64 {
	return NewSQLUint64(s.kind, uint64(s.val))
}

// SQLVarchar converts the BaseSQLInt64 receiver, s, to a SQLVarchar.
func (s BaseSQLInt64) SQLVarchar() SQLVarchar {
	return NewSQLVarchar(s.kind, s.varchar())
}

// BaseSQLObjectID represents a MongoDB ObjectId using its string value.
// BaseSQLObjectID should be treated as an abstract type; it should never
// be used directly during query evaluation, and instead should be embedded
// in other SQLValue implementers who can override BaseSQLObjectID's default
// implementations as needed.
type BaseSQLObjectID struct {
	val  string
	kind SQLValueKind
}

// iSQLObjectID must be implemented to satisfy the SQLObjectID interface.
func (BaseSQLObjectID) iSQLObjectID() {}

func newBaseSQLObjectID(kind SQLValueKind, val string) BaseSQLObjectID {
	return BaseSQLObjectID{
		val:  val,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLObjectID) IsNull() bool {
	return false
}

// CloneWithKind copies the SQLValue with the passed kind.
func (s BaseSQLObjectID) CloneWithKind(kind SQLValueKind) SQLValue {
	kind.AssertValid()
	return NewSQLObjectID(kind, s.val)
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLObjectID) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLObjectID) WireProtocolEncode(charSet *collation.Charset,
	mongoDBVarcharLength int) ([]byte, error) {
	b := []byte(s.val)
	if string(charSet.Name) == "utf8" {
		return b, nil
	}
	ret := charSet.Encode(b)
	// ObjectIds are serialized as Varchars and must respect the maximum
	// varchar length.
	if mongoDBVarcharLength != 0 && len(ret) > mongoDBVarcharLength {
		ret = []byte(string(ret)[:mongoDBVarcharLength])
	}

	return ret, nil
}

// Size returns the size of this SQLValue in bytes.
func (s BaseSQLObjectID) Size() uint64 {
	return uint64(len(s.val))
}

// String returns the string representation of this SQLValue.
// String should return the same value regardless of the SQLValue's kind, and
// should not be overridden by any embedding SQLValue implementers.
func (s BaseSQLObjectID) String() string {
	return s.varchar()
}

func (s BaseSQLObjectID) varchar() string {
	return s.val
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLObjectID) EvalType() types.EvalType {
	return types.EvalObjectID
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLObjectID) Value() interface{} {
	return bson.ObjectIdHex(s.val)
}

// BSONValue returns a bsoncore.Value that represents the literal value of this SQLValue.
func (s BaseSQLObjectID) BSONValue() (bsoncore.Value, error) {
	oid, err := primitive.ObjectIDFromHex(s.val)
	if err != nil {
		return bsoncore.Value{}, fmt.Errorf("failed to parse ObjectID from SQLValue: %v", err)
	}
	return bsoncore.Value{
		Type: bsontype.ObjectID,
		Data: bsoncore.AppendObjectID(nil, oid),
	}, nil
}

// SQLBool converts a BaseSQLObjectID to a SQLBool.
func (s BaseSQLObjectID) SQLBool() SQLBool {
	return NewSQLBool(s.kind, false)
}

// SQLDate converts the BaseSQLObjectID receiver, s, to a SQLDate.
func (s BaseSQLObjectID) SQLDate() SQLDate {
	return s.SQLTimestamp().SQLDate()
}

// SQLDecimal128 converts a BaseSQLObjectID to a SQLDecimal128 by converting to SQLTimestamp then
// to SQLDecimal128.
func (s BaseSQLObjectID) SQLDecimal128() SQLDecimal128 {
	return s.SQLTimestamp().SQLDecimal128()
}

// SQLFloat converts a BaseSQLObjectID to a SQLFloat by converting to SQLTimestamp then to SQLFloat.
func (s BaseSQLObjectID) SQLFloat() SQLFloat {
	return s.SQLTimestamp().SQLFloat()
}

// SQLInt converts a BaseSQLObjectID to a SQLInt by converting to SQLTimestamp then to SQLInt.
func (s BaseSQLObjectID) SQLInt() SQLInt64 {
	return s.SQLTimestamp().SQLInt()
}

// SQLObjectID converts the BaseSQLObjectID receiver, s, to a SQLObjectID.
func (s BaseSQLObjectID) SQLObjectID() SQLObjectID {
	return NewSQLObjectID(s.kind, s.val)
}

// SQLTimestamp converts the BaseSQLObjectID receiver, s, to a SQLTimestamp.
func (s BaseSQLObjectID) SQLTimestamp() SQLTimestamp {
	// If it's a valid ObjectId, attempt to get the timestamp from it.
	if len(s.varchar()) == 24 {
		_, err := hex.DecodeString(s.varchar())
		if err == nil {
			t := bson.ObjectIdHex(s.varchar()).Time().In(schema.DefaultLocale)
			return NewSQLTimestamp(s.kind, t)
		}
	}
	return NewSQLTimestamp(s.kind, NullDate)
}

// SQLUint converts a BaseSQLObjectID to a SQLUint64 by converting to SQLTimestamp then to SQLUint.
func (s BaseSQLObjectID) SQLUint() SQLUint64 {
	return s.SQLTimestamp().SQLUint()
}

// SQLVarchar converts the BaseSQLObjectID receiver, s, to a SQLVarchar.
func (s BaseSQLObjectID) SQLVarchar() SQLVarchar {
	return NewSQLVarchar(s.kind, s.val)
}

// BaseSQLTimestamp represents a timestamp value. BaseSQLTimestamp should be
// treated as an abstract type; it should never be used directly during query
// evaluation, and instead should be embedded in other SQLValue implementers who
// can override BaseSQLTimestamp's default implementations as needed.
type BaseSQLTimestamp struct {
	datetime time.Time
	null     bool
	kind     SQLValueKind
}

// iSQLTimestamp must be implemented to satisfy the SQLTimestamp interface.
func (BaseSQLTimestamp) iSQLTimestamp() {}

func newBaseSQLTimestamp(kind SQLValueKind, val time.Time) BaseSQLTimestamp {
	return BaseSQLTimestamp{
		datetime: val,
		kind:     kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLTimestamp) IsNull() bool {
	return false
}

// CloneWithKind copies the SQLValue with the passed kind.
func (s BaseSQLTimestamp) CloneWithKind(kind SQLValueKind) SQLValue {
	kind.AssertValid()
	return NewSQLTimestamp(kind, s.datetime)
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLTimestamp) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLTimestamp) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	if s.null {
		return nil, nil
	}
	if s.datetime == NullDate {
		return []byte("0000-00-00 00:00:00"), nil
	}
	if strings.Contains(s.datetime.String(), ".") {
		return strutil.Slice(s.datetime.Format(schema.TimestampFormatMicros)), nil
	}
	return strutil.Slice(s.datetime.Format(schema.TimestampFormat)), nil
}

// Size returns the size of this SQLValue in bytes.
func (s BaseSQLTimestamp) Size() uint64 {
	return 8
}

// String returns the string representation of this SQLValue.
// String should return the same value regardless of the SQLValue's kind, and
// should not be overridden by any embedding SQLValue implementers.
func (s BaseSQLTimestamp) String() string {
	if s.null {
		return "NULL"
	}
	return s.datetime.Format("2006-01-02 15:04:05.000000")
}

func (s BaseSQLTimestamp) varchar() string {
	if s.null {
		return "NULL"
	}
	return s.datetime.Format("2006-01-02T15:04:05.000Z")
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLTimestamp) EvalType() types.EvalType {
	return types.EvalDatetime
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLTimestamp) Value() interface{} {
	if s.null {
		return NullDate
	}
	return s.datetime
}

// BSONValue returns a bsoncore.Value that represents the literal value of this SQLValue.
func (s BaseSQLTimestamp) BSONValue() (bsoncore.Value, error) {
	return bsoncore.Value{
		Type: bsontype.DateTime,
		Data: bsoncore.AppendDateTime(nil, dateutil.UnixMillis(s.datetime)),
	}, nil
}

// SQLBool converts the BaseSQLTimestamp receiver, s, to a SQLBool.
func (s BaseSQLTimestamp) SQLBool() SQLBool {
	t := s.datetime
	if t == NullDate {
		return NewSQLBool(s.kind, false)
	}
	return NewSQLBool(s.kind, true)
}

// SQLDate converts the BaseSQLTimestamp receiver, s, to a SQLDate.
func (s BaseSQLTimestamp) SQLDate() SQLDate {
	t := s.datetime
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the BaseSQLTimestamp receiver, s, to a SQLDecimal128.
func (s BaseSQLTimestamp) SQLDecimal128() SQLDecimal128 {
	flt := Float64(s)
	dec := decimal.NewFromFloat(flt)
	return NewSQLDecimal128(s.kind, dec)
}

// SQLFloat converts the BaseSQLTimestamp receiver, s, to a SQLFloat.
func (s BaseSQLTimestamp) SQLFloat() SQLFloat {
	epochMs := dateutil.UnixMillis(s.datetime)
	return NewSQLFloat(s.kind, float64(epochMs))
}

// SQLInt converts the BaseSQLTimestamp receiver, s, to a SQLInt.
func (s BaseSQLTimestamp) SQLInt() SQLInt64 {
	epochMs := dateutil.UnixMillis(s.datetime)
	return NewSQLInt64(s.kind, epochMs)
}

// SQLObjectID converts the BaseSQLTimestamp receiver, s, to a SQLObjectID.
func (s BaseSQLTimestamp) SQLObjectID() SQLObjectID {
	return NewSQLNull(s.kind)
}

// SQLTimestamp converts the BaseSQLTimestamp receiver, s, to a SQLTimestamp.
func (s BaseSQLTimestamp) SQLTimestamp() SQLTimestamp {
	return NewSQLTimestamp(s.kind, s.datetime)
}

// SQLUint converts the BaseSQLTimestamp receiver, s, to a SQLUint64.
func (s BaseSQLTimestamp) SQLUint() SQLUint64 {
	epochMs := dateutil.UnixMillis(s.datetime)
	return NewSQLUint64(s.kind, uint64(epochMs))
}

// SQLVarchar converts the BaseSQLTimestamp receiver, s, to a SQLVarchar.
func (s BaseSQLTimestamp) SQLVarchar() SQLVarchar {
	return NewSQLVarchar(s.kind, s.varchar())
}

// BaseSQLUint64 represents an unsigned 64-bit integer. BaseSQLUint64 should be
// treated as an abstract type; it should never be used directly during query
// evaluation, and instead should be embedded in other SQLValue implementers who
// can override BaseSQLUint64's default implementations as needed.
type BaseSQLUint64 struct {
	val  uint64
	kind SQLValueKind
}

func (BaseSQLUint64) iNumber() {}

// iSQLUint64 must be implemented to satisfy the SQLUint64 interface.
func (BaseSQLUint64) iSQLUint64() {}

func newBaseSQLUint64(kind SQLValueKind, val uint64) BaseSQLUint64 {
	return BaseSQLUint64{
		val:  val,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLUint64) IsNull() bool {
	return false
}

// CloneWithKind copies the SQLValue with the passed kind.
func (s BaseSQLUint64) CloneWithKind(kind SQLValueKind) SQLValue {
	kind.AssertValid()
	return NewSQLUint64(kind, s.val)
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLUint64) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// Size returns the size of this SQLValue in bytes.
func (s BaseSQLUint64) Size() uint64 {
	return 8
}

// String returns the string representation of this SQLValue.
// String should return the same value regardless of the SQLValue's kind, and
// should not be overridden by any embedding SQLValue implementers.
func (s BaseSQLUint64) String() string {
	return s.varchar()
}

func (s BaseSQLUint64) varchar() string {
	return strconv.FormatUint(s.val, 10)
}

// EvalType returns the SQLType of this SQLValue.
func (s BaseSQLUint64) EvalType() types.EvalType {
	return types.EvalUint64
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLUint64) Value() interface{} {
	return s.val
}

// BSONValue returns a bsoncore.Value that represents the literal value of this SQLValue.
func (s BaseSQLUint64) BSONValue() (bsoncore.Value, error) {
	if s.val > math.MaxInt64 {
		return bsoncore.Value{}, fmt.Errorf("uint64 greater than MaxInt64")
	}
	return bsoncore.Value{
		Type: bsontype.Int64,
		Data: bsoncore.AppendInt64(nil, int64(s.val)),
	}, nil
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLUint64) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	return strconv.AppendUint(nil, s.val, 10), nil
}

// SQLBool converts the BaseSQLUint64 receiver, s, to a SQLBool.
func (s BaseSQLUint64) SQLBool() SQLBool {
	if s.val == 0 {
		return NewSQLBool(s.kind, false)
	}
	return NewSQLBool(s.kind, true)
}

// SQLDate converts the BaseSQLUint64 receiver, s, to a SQLDate.
func (s BaseSQLUint64) SQLDate() SQLDate {
	if s.val == 0 {
		return NewSQLDate(s.kind, NullDate)
	}

	t, _, ok := ParseDateTime(s.varchar())
	if !ok {
		return NewSQLNull(s.kind)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the BaseSQLUint64 receiver, s, to a SQLDecimal128.
func (s BaseSQLUint64) SQLDecimal128() SQLDecimal128 {
	var d decimal.Decimal
	var err error

	if s.val > math.MaxInt64 {
		d, err = decimal.NewFromString(fmt.Sprintf("%v", s.val))
		if err != nil {
			return nil
		}
	} else {
		d = decimal.New(int64(s.val), 0)
	}

	return NewSQLDecimal128(s.kind, d)
}

// SQLFloat converts the BaseSQLUint64 receiver, s, to a SQLFloat.
func (s BaseSQLUint64) SQLFloat() SQLFloat {
	return NewSQLFloat(s.kind, float64(s.val))
}

// SQLInt converts the BaseSQLUint64 receiver, s, to a SQLInt.
func (s BaseSQLUint64) SQLInt() SQLInt64 {
	return NewSQLInt64(s.kind, int64(s.val))
}

// SQLObjectID converts the BaseSQLUint64 receiver, s, to a SQLObjectID.
func (s BaseSQLUint64) SQLObjectID() SQLObjectID {
	return NewSQLNull(s.kind)
}

// SQLTimestamp converts the BaseSQLUint64 receiver, s, to a SQLTimestamp.
func (s BaseSQLUint64) SQLTimestamp() SQLTimestamp {
	if s.val == 0 {
		return NewSQLTimestamp(s.kind, NullDate)
	}

	t, _, ok := ParseDateTime(s.varchar())
	if !ok {
		return NewSQLNull(s.kind)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLUint converts the BaseSQLUint64 receiver, s, to a SQLUint64.
func (s BaseSQLUint64) SQLUint() SQLUint64 {
	return NewSQLUint64(s.kind, s.val)
}

// SQLVarchar converts the BaseSQLUint64 receiver, s, to a SQLVarchar.
func (s BaseSQLUint64) SQLVarchar() SQLVarchar {
	return NewSQLVarchar(s.kind, s.varchar())
}

// BaseSQLVarchar represents a string value. BaseSQLVarchar should be treated as
// an abstract type; it should never be used directly during query evaluation,
// and instead should be embedded in other SQLValue implementers who can
// override BaseSQLVarchar's default implementations as needed.
type BaseSQLVarchar struct {
	val  string
	kind SQLValueKind
}

// iSQLVarchar must be implemented to satisfy the SQLVarchar interface.
func (BaseSQLVarchar) iSQLVarchar() {}

func newBaseSQLVarchar(kind SQLValueKind, val string) BaseSQLVarchar {
	return BaseSQLVarchar{
		val:  val,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLVarchar) IsNull() bool {
	return false
}

// CloneWithKind copies the SQLValue with the passed kind.
func (s BaseSQLVarchar) CloneWithKind(kind SQLValueKind) SQLValue {
	kind.AssertValid()
	return NewSQLVarchar(kind, s.val)
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLVarchar) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLVarchar) WireProtocolEncode(charSet *collation.Charset,
	mongoDBVarcharLength int) ([]byte,
	error) {
	b := []byte(s.val)
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
func (s BaseSQLVarchar) Size() uint64 {
	return uint64(len(s.val))
}

// String returns the string representation of this SQLValue.
// String should return the same value regardless of the SQLValue's kind, and
// should not be overridden by any embedding SQLValue implementers.
func (s BaseSQLVarchar) String() string {
	return s.varchar()
}

func (s BaseSQLVarchar) varchar() string {
	return s.val
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLVarchar) EvalType() types.EvalType {
	return types.EvalString
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLVarchar) Value() interface{} {
	return s.val
}

// BSONValue returns a bsoncore.Value that represents the literal value of this SQLValue.
func (s BaseSQLVarchar) BSONValue() (bsoncore.Value, error) {
	return bsoncore.Value{
		Type: bsontype.String,
		Data: bsoncore.AppendString(nil, s.val),
	}, nil
}

// SQLBool converts the BaseSQLVarchar receiver, s, to a SQLBool.
func (s BaseSQLVarchar) SQLBool() SQLBool {
	// Note that we convert to Bool by converting to Int then to Bool,
	// these are the specified semantics of mysql.
	return s.SQLInt().SQLBool()
}

// SQLDate converts the BaseSQLVarchar receiver, s, to a SQLDate.
func (s BaseSQLVarchar) SQLDate() SQLDate {
	t, _, ok := ParseDateTimeMongo(strings.TrimSpace(s.val))
	if !ok {
		return NewSQLNull(s.kind)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the BaseSQLVarchar receiver, s, to a SQLDecimal128.
func (s BaseSQLVarchar) SQLDecimal128() SQLDecimal128 {
	out, err := decimal.NewFromString(s.val)
	if err != nil {
		return NewSQLDecimal128(s.kind, decimal.Zero)
	}
	return NewSQLDecimal128(s.kind, out)
}

// SQLFloat converts the BaseSQLVarchar receiver, s, to a SQLFloat.
func (s BaseSQLVarchar) SQLFloat() SQLFloat {
	out, _ := strconv.ParseFloat(s.val, 64)
	return NewSQLFloat(s.kind, out)
}

// SQLInt converts the BaseSQLVarchar receiver, s, to a SQLInt.
func (s BaseSQLVarchar) SQLInt() SQLInt64 {
	out, _ := strconv.ParseInt(s.val, 10, 64)
	return NewSQLInt64(s.kind, out)
}

// SQLObjectID converts the BaseSQLVarchar receiver, s, to a SQLObjectID.
func (s BaseSQLVarchar) SQLObjectID() SQLObjectID {
	// Return null if this is not a valid ObjectID.
	if len(s.val) == 24 {
		_, err := hex.DecodeString(s.varchar())
		if err == nil {
			return NewSQLObjectID(s.kind, s.val)
		}
	}

	return NewSQLNull(s.kind)
}

// SQLTimestamp converts the BaseSQLVarchar receiver, s, to a SQLTimestamp.
func (s BaseSQLVarchar) SQLTimestamp() SQLTimestamp {
	t, _, ok := ParseDateTimeMongo(strings.TrimSpace(s.val))
	if !ok {
		return NewSQLNull(s.kind)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(s.kind, t)
}

// SQLUint converts the BaseSQLVarchar receiver, s, to a SQLUint64.
func (s BaseSQLVarchar) SQLUint() SQLUint64 {
	out, _ := strconv.ParseInt(s.val, 10, 64)
	return NewSQLUint64(s.kind, uint64(out))
}

// SQLVarchar converts the BaseSQLVarchar receiver, s, to a SQLVarchar.
func (s BaseSQLVarchar) SQLVarchar() SQLVarchar {
	return NewSQLVarchar(s.kind, s.val)
}

// BaseSQLNull represents a NULL value. The interesting thing about
// SQLNull is that it implements the value type interfaces for all value
// types.
type BaseSQLNull struct {
	kind SQLValueKind
}

func newBaseSQLNull(kind SQLValueKind) BaseSQLNull {
	return BaseSQLNull{
		kind: kind,
	}
}

// SQLNull must be implemented to satisfy the SQLBool interface.
func (BaseSQLNull) iSQLBool() {}

// SQLNull must be implemented to satisfy the SQLDate interface.
func (BaseSQLNull) iSQLDate() {}

// SQLNull must be implemented to satisfy the SQLDecimal128 interface.
func (BaseSQLNull) iSQLDecimal128() {}

// SQLNull must be implemented to satisfy the SQLFloat interface.
func (BaseSQLNull) iSQLFloat() {}

// SQLNull must be implemented to satisfy the SQLInt32 interface.
func (BaseSQLNull) iSQLInt32() {}

// SQLNull must be implemented to satisfy the SQLInt64 interface.
func (BaseSQLNull) iSQLInt64() {}

// SQLNull must be implemented to satisfy the SQLObjectID interface.
func (BaseSQLNull) iSQLObjectID() {}

// SQLNull must be implemented to satisfy the SQLUint32 interface.
func (BaseSQLNull) iSQLUint32() {}

// SQLNull must be implemented to satisfy the SQLUint64 interface.
func (BaseSQLNull) iSQLUint64() {}

// SQLNull must be implemented to satisfy the SQLTimestamp interface.
func (BaseSQLNull) iSQLTimestamp() {}

// SQLNull must be implemented to satisfy the SQLVarchar interface.
func (BaseSQLNull) iSQLVarchar() {}

// SQLNull must be implemented to satisfy the SQLNull interface.
func (BaseSQLNull) iSQLNull() {}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLNull) IsNull() bool {
	return true
}

// CloneWithKind copies the SQLValue with the passed kind.
func (s BaseSQLNull) CloneWithKind(kind SQLValueKind) SQLValue {
	kind.AssertValid()
	return NewSQLNull(kind)
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLNull) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLNull) WireProtocolEncode(charSet *collation.Charset,
	mongoDBNullLength int) ([]byte,
	error) {
	return nil, nil
}

// Size returns the size of this SQLValue in bytes.
func (s BaseSQLNull) Size() uint64 {
	return 0
}

// String returns the string representation of this SQLValue.
// String should return the same value regardless of the SQLValue's kind, and
// should not be overridden by any embedding SQLValue implementers.
func (s BaseSQLNull) String() string {
	return s.varchar()
}

func (s BaseSQLNull) varchar() string {
	return "NULL"
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLNull) EvalType() types.EvalType {
	return types.EvalNull
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLNull) Value() interface{} {
	return false
}

// BSONValue returns a bsoncore.Value that represents the literal value of this SQLValue.
func (s BaseSQLNull) BSONValue() (bsoncore.Value, error) {
	return bsoncore.Value{
		Type: bsontype.Null,
	}, nil
}

// SQLBool converts the BaseSQLNull receiver, s, to a SQLBool.
func (s BaseSQLNull) SQLBool() SQLBool {
	return NewSQLBool(s.kind, false)
}

// SQLDate converts the BaseSQLNull receiver, s, to a SQLDate.
func (s BaseSQLNull) SQLDate() SQLDate {
	return NewSQLDate(s.kind, NullDate)
}

// SQLDecimal128 converts the BaseSQLNull receiver, s, to a SQLDecimal128.
func (s BaseSQLNull) SQLDecimal128() SQLDecimal128 {
	return NewSQLDecimal128(s.kind, decimal.NewFromFloat(0.0))
}

// SQLFloat converts the BaseSQLNull receiver, s, to a SQLFloat.
func (s BaseSQLNull) SQLFloat() SQLFloat {
	return NewSQLFloat(s.kind, 0.0)
}

// SQLInt converts the BaseSQLNull receiver, s, to a SQLInt.
func (s BaseSQLNull) SQLInt() SQLInt64 {
	return NewSQLInt64(s.kind, 0)
}

// SQLObjectID converts the BaseSQLNull receiver, s, to a SQLObjectID.
func (s BaseSQLNull) SQLObjectID() SQLObjectID {
	return NewSQLObjectID(s.kind, "000000000000000000000000")
}

// SQLTimestamp converts the BaseSQLNull receiver, s, to a SQLTimestamp.
func (s BaseSQLNull) SQLTimestamp() SQLTimestamp {
	return NewSQLTimestamp(s.kind, NullDate)
}

// SQLUint converts the BaseSQLNull receiver, s, to a SQLUint64.
func (s BaseSQLNull) SQLUint() SQLUint64 {
	return NewSQLUint64(s.kind, 0)
}

// SQLVarchar converts the BaseSQLNull receiver, s, to a SQLVarchar.
func (s BaseSQLNull) SQLVarchar() SQLVarchar {
	return NewSQLVarchar(s.kind, "")
}
