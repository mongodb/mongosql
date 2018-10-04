package evaluator

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/schema"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/shopspring/decimal"
)

// BaseSQLBool represents a boolean. BaseSQLBool should be treated as an
// abstract type; it should never be used directly during query evaluation, and
// instead should be embedded in other SQLValue implementers who can override
// BaseSQLBool's default implementations as needed.
type BaseSQLBool struct {
	val  float64
	null bool
	kind SQLValueKind
}

var _ translatableToAggregation = (*BaseSQLBool)(nil)

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

func nullBaseSQLBool(kind SQLValueKind) BaseSQLBool {
	return BaseSQLBool{
		null: true,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLBool) IsNull() bool {
	return s.null
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLBool) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (s BaseSQLBool) Evaluate(_ context.Context, _ *ExecutionConfig, _ *ExecutionState) (SQLValue, error) {
	return s.SQLBool(), nil
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLBool) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	if s.null {
		return nil, nil
	}
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
	if s.null {
		return "NULL"
	}
	return strconv.FormatFloat(Float64(s), 'f', -1, 64)
}

// ToAggregationLanguage translates SQLBool into something that can
// be used in an aggregation pipeline. If SQLBool cannot be translated,
// it will return nil and false.
func (s BaseSQLBool) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	if s.null {
		return mgoNullLiteral, nil
	}
	return wrapInLiteral(Bool(s)), nil
}

// EvalType returns the EvalType of this SQLValue.
func (BaseSQLBool) EvalType() EvalType {
	return EvalBoolean
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLBool) Value() interface{} {
	if s.null {
		return false
	}
	return s.val != 0
}

// SQLBool converts the BaseSQLBool receiver, s, to a SQLBool.
func (s BaseSQLBool) SQLBool() SQLBool {
	if s.null {
		return nullSQLBool(s.kind)
	}
	return NewSQLBool(s.kind, !(s.val == 0))
}

// SQLDate converts the BaseSQLBool receiver, s, to a SQLDate.
func (s BaseSQLBool) SQLDate() SQLDate {
	if s.null {
		return nullSQLDate(s.kind)
	}
	if s.val == 0 {
		return NewSQLDate(s.kind, NullDate)
	}
	return nullSQLDate(s.kind)
}

// SQLDecimal128 converts the BaseSQLBool receiver, s, to a SQLDecimal128.
func (s BaseSQLBool) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return nullSQLDecimal128(s.kind)
	}
	if s.val == 0 {
		return NewSQLDecimal128(s.kind, decimal.Zero)
	}
	return NewSQLDecimal128(s.kind, decimal.New(1, 0))
}

// SQLFloat converts the BaseSQLBool receiver, s, to a SQLFloat.
func (s BaseSQLBool) SQLFloat() SQLFloat {
	if s.null {
		return nullSQLFloat(s.kind)
	}
	if s.val == 0 {
		return NewSQLFloat(s.kind, 0.0)
	}
	return NewSQLFloat(s.kind, 1.0)
}

// SQLInt converts the BaseSQLBool receiver, s, to a SQLInt.
func (s BaseSQLBool) SQLInt() SQLInt64 {
	if s.null {
		return nullSQLInt64(s.kind)
	}
	if s.val == 0 {
		return NewSQLInt64(s.kind, 0)
	}
	return NewSQLInt64(s.kind, 1)
}

// SQLObjectID converts the BaseSQLBool receiver, s, to a SQLObjectID.
func (s BaseSQLBool) SQLObjectID() SQLObjectID {
	return nullSQLObjectID(s.kind)
}

// SQLTimestamp converts the BaseSQLBool receiver, s, to a SQLTimestamp.
func (s BaseSQLBool) SQLTimestamp() SQLTimestamp {
	if s.null {
		return nullSQLTimestamp(s.kind)
	}
	if s.val == 0 {
		return NewSQLTimestamp(s.kind, NullDate)
	}
	return nullSQLTimestamp(s.kind)
}

// SQLUint converts the BaseSQLBool receiver, s, to a SQLUint64.
func (s BaseSQLBool) SQLUint() SQLUint64 {
	if s.null {
		return nullSQLUint64(s.kind)
	}
	if s.val == 0 {
		return NewSQLUint64(s.kind, 0)
	}
	return NewSQLUint64(s.kind, 1)
}

// SQLVarchar converts the BaseSQLBool receiver, s, to a SQLVarchar.
func (s BaseSQLBool) SQLVarchar() SQLVarchar {
	if s.null {
		return nullSQLVarchar(s.kind)
	}
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
	null     bool
	kind     SQLValueKind
}

var _ translatableToAggregation = (*BaseSQLDate)(nil)

// iSQLDate must be implemented to satisfy the SQLDate interface.
func (BaseSQLDate) iSQLDate() {}

func newBaseSQLDate(kind SQLValueKind, val time.Time) BaseSQLDate {
	t := time.Date(val.Year(), val.Month(), val.Day(), 0, 0, 0, 0, schema.DefaultLocale)
	return BaseSQLDate{
		datetime: t,
		kind:     kind,
	}
}

func nullBaseSQLDate(kind SQLValueKind) BaseSQLDate {
	return BaseSQLDate{
		null: true,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLDate) IsNull() bool {
	return s.null
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLDate) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (s BaseSQLDate) Evaluate(_ context.Context, _ *ExecutionConfig, _ *ExecutionState) (SQLValue, error) {
	return s.SQLDate(), nil
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLDate) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	if s.null {
		return nil, nil
	}
	if s.datetime == NullDate {
		return []byte("0000-00-00"), nil
	}
	return util.Slice(s.datetime.Format(schema.DateFormat)), nil
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
	if s.null {
		return "NULL"
	}
	return s.datetime.Format("2006-01-02")
}

// ToAggregationLanguage translates SQLDate into something that can
// be used in an aggregation pipeline. If SQLDate cannot be translated,
// it will return nil and false.
func (s BaseSQLDate) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	if s.null {
		return mgoNullLiteral, nil
	}
	return wrapInLiteral(s.datetime), nil
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLDate) EvalType() EvalType {
	return EvalDate
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLDate) Value() interface{} {
	if s.null {
		return NullDate
	}
	return s.datetime
}

// SQLBool converts the BaseSQLDate receiver, s, to a SQLBool.
func (s BaseSQLDate) SQLBool() SQLBool {
	if s.null {
		return nullSQLBool(s.kind)
	}
	t := s.datetime
	if t == NullDate {
		return NewSQLBool(s.kind, false)
	}
	return NewSQLBool(s.kind, true)
}

// SQLDecimal128 converts the BaseSQLDate receiver, s, to a SQLDecimal128.
func (s BaseSQLDate) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return nullSQLDecimal128(s.kind)
	}
	flt := Float64(s)
	dec := decimal.NewFromFloat(flt)
	return NewSQLDecimal128(s.kind, dec)
}

// SQLDate converts the BaseSQLDate receiver, s, to a SQLDate.
func (s BaseSQLDate) SQLDate() SQLDate {
	if s.null {
		return nullSQLDate(s.kind)
	}
	return NewSQLDate(s.kind, s.datetime)
}

// SQLFloat converts the BaseSQLDate receiver, s, to a SQLFloat.
func (s BaseSQLDate) SQLFloat() SQLFloat {
	if s.null {
		return nullSQLFloat(s.kind)
	}
	epochMs := s.datetime.UnixNano() / 1000000
	return NewSQLFloat(s.kind, float64(epochMs))
}

// SQLInt converts the BaseSQLDate receiver, s, to a SQLInt.
func (s BaseSQLDate) SQLInt() SQLInt64 {
	if s.null {
		return nullSQLInt64(s.kind)
	}
	epochMs := s.datetime.UnixNano() / 1000000
	return NewSQLInt64(s.kind, epochMs)
}

// SQLObjectID converts the BaseSQLDate receiver, s, to a SQLObjectID.
func (s BaseSQLDate) SQLObjectID() SQLObjectID {
	return nullSQLObjectID(s.kind)
}

// SQLTimestamp converts the BaseSQLDate receiver, s, to a SQLTimestamp.
func (s BaseSQLDate) SQLTimestamp() SQLTimestamp {
	if s.null {
		return nullSQLTimestamp(s.kind)
	}
	t := s.datetime
	return NewSQLTimestamp(s.kind, t)
}

// SQLUint converts the BaseSQLDate receiver, s, to a SQLUint64.
func (s BaseSQLDate) SQLUint() SQLUint64 {
	if s.null {
		return nullSQLUint64(s.kind)
	}
	epochMs := s.datetime.UnixNano() / 1000000
	return NewSQLUint64(s.kind, uint64(epochMs))
}

// SQLVarchar converts the BaseSQLDate receiver, s, to a SQLVarchar.
func (s BaseSQLDate) SQLVarchar() SQLVarchar {
	if s.null {
		return nullSQLVarchar(s.kind)
	}
	return NewSQLVarchar(s.kind, s.varchar())
}

// BaseSQLDecimal128 represents a decimal 128 value. BaseSQLDecimal128 should be
// treated as an abstract type; it should never be used directly during query
// evaluation, and instead should be embedded in other SQLValue implementers who
// can override BaseSQLDecimal128's default implementations as needed.
type BaseSQLDecimal128 struct {
	val  decimal.Decimal
	null bool
	kind SQLValueKind
}

var _ translatableToAggregation = (*BaseSQLDecimal128)(nil)

// iSQLDecimal128 must be implemented to satisfy the SQLDecimal128 interface.
func (BaseSQLDecimal128) iSQLDecimal128() {}

func newBaseSQLDecimal128(kind SQLValueKind, val decimal.Decimal) BaseSQLDecimal128 {
	return BaseSQLDecimal128{
		val:  val,
		kind: kind,
	}
}

func nullBaseSQLDecimal128(kind SQLValueKind) BaseSQLDecimal128 {
	return BaseSQLDecimal128{
		null: true,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLDecimal128) IsNull() bool {
	return s.null
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLDecimal128) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (s BaseSQLDecimal128) Evaluate(_ context.Context, _ *ExecutionConfig, _ *ExecutionState) (SQLValue, error) {
	return s.SQLDecimal128(), nil
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLDecimal128) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	if s.null {
		return nil, nil
	}
	return []byte(util.FormatDecimal(s.val)), nil
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
	if s.null {
		return "NULL"
	}
	return s.val.String()
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLDecimal128) EvalType() EvalType {
	return EvalDecimal128
}

// ToAggregationLanguage translates SQLDecimal128 into something that can
// be used in an aggregation pipeline. If SQLDecimal128 cannot be translated,
// it will return nil and false.
func (s BaseSQLDecimal128) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	if s.null {
		return mgoNullLiteral, nil
	}
	d, ok := t.translateDecimal(s)
	if !ok {
		return nil, fmt.Errorf("could not translate '%s' as a decimal", s)
	}
	return wrapInLiteral(d), nil
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLDecimal128) Value() interface{} {
	if s.null {
		return decimal.Zero
	}
	return s.val
}

// SQLBool converts the BaseSQLDecimal128 receiver, s, to a SQLBool.
func (s BaseSQLDecimal128) SQLBool() SQLBool {
	return s.SQLInt().SQLBool()
}

// SQLDate converts the BaseSQLDecimal128 receiver, s, to a SQLDate.
func (s BaseSQLDecimal128) SQLDate() SQLDate {
	if s.null {
		return nullSQLDate(s.kind)
	}
	t := time.Unix(0, s.val.IntPart()*1000000)
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the BaseSQLDecimal128 receiver, s, to a SQLDecimal128.
func (s BaseSQLDecimal128) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return nullSQLDecimal128(s.kind)
	}
	return NewSQLDecimal128(s.kind, s.val)
}

// SQLFloat converts the BaseSQLDecimal128 receiver, s, to a SQLFloat.
func (s BaseSQLDecimal128) SQLFloat() SQLFloat {
	if s.null {
		return nullSQLFloat(s.kind)
	}
	// Second return value tells us if this is exact, we don't care.
	f, _ := s.val.Float64()
	return NewSQLFloat(s.kind, f)
}

// SQLInt converts the BaseSQLDecimal128 receiver, s, to a SQLInt.
func (s BaseSQLDecimal128) SQLInt() SQLInt64 {
	if s.null {
		return nullSQLInt64(s.kind)
	}
	// Do not care if this is exact.
	f, _ := s.val.Float64()
	return NewSQLInt64(s.kind, round(f))
}

// SQLObjectID converts the BaseSQLDecimal128 receiver, s, to a SQLObjectID.
func (s BaseSQLDecimal128) SQLObjectID() SQLObjectID {
	return nullSQLObjectID(s.kind)
}

// SQLTimestamp converts the BaseSQLDecimal128 receiver, s, to a SQLTimestamp.
func (s BaseSQLDecimal128) SQLTimestamp() SQLTimestamp {
	if s.null {
		return nullSQLTimestamp(s.kind)
	}
	t := time.Unix(0, s.val.IntPart()*1000000)
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(s.kind, t)
}

// SQLUint converts the BaseSQLDecimal128 receiver, s, to a SQLUint64.
func (s BaseSQLDecimal128) SQLUint() SQLUint64 {
	if s.null {
		return nullSQLUint64(s.kind)
	}
	// Do not care if this is exact.
	f, _ := s.val.Float64()
	return NewSQLUint64(s.kind, uint64(round(f)))
}

// SQLVarchar converts the BaseSQLDecimal128 receiver, s, to a SQLVarchar.
func (s BaseSQLDecimal128) SQLVarchar() SQLVarchar {
	if s.null {
		return nullSQLVarchar(s.kind)
	}
	return NewSQLVarchar(s.kind, s.val.String())
}

// BaseSQLFloat represents a float. BaseSQLFloat should be treated as an
// abstract type; it should never be used directly during query evaluation, and
// instead should be embedded in other SQLValue implementers who can override
// BaseSQLFloat's default implementations as needed.
type BaseSQLFloat struct {
	val  float64
	null bool
	kind SQLValueKind
}

var _ translatableToAggregation = (*BaseSQLFloat)(nil)

// iSQLFloat must be implemented to satisfy the SQLFloat interface.
func (BaseSQLFloat) iSQLFloat() {}

func newBaseSQLFloat(kind SQLValueKind, val float64) BaseSQLFloat {
	return BaseSQLFloat{
		val:  val,
		kind: kind,
	}
}

func nullBaseSQLFloat(kind SQLValueKind) BaseSQLFloat {
	return BaseSQLFloat{
		null: true,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLFloat) IsNull() bool {
	return s.null
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLFloat) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (s BaseSQLFloat) Evaluate(_ context.Context, _ *ExecutionConfig, _ *ExecutionState) (SQLValue, error) {
	return s.SQLFloat(), nil
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLFloat) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	if s.null {
		return nil, nil
	}
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
	if s.null {
		return "NULL"
	}
	return strconv.FormatFloat(s.val, 'f', -1, 64)
}

// ToAggregationLanguage translates SQLFloat into something that can
// be used in an aggregation pipeline. If SQLFloat cannot be translated,
// it will return nil and false.
func (s BaseSQLFloat) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	if s.null {
		return mgoNullLiteral, nil
	}
	return wrapInLiteral(s.Value()), nil
}

// SQLBool converts the BaseSQLFloat receiver, s, to a SQLBool.
func (s BaseSQLFloat) SQLBool() SQLBool {
	return s.SQLInt().SQLBool()
}

// SQLDate converts the BaseSQLFloat receiver, s, to a SQLDate.
func (s BaseSQLFloat) SQLDate() SQLDate {
	if s.null {
		return nullSQLDate(s.kind)
	}
	t := time.Unix(0, int64(s.val)*1000000)
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the BaseSQLFloat receiver, s, to a SQLDecimal128.
func (s BaseSQLFloat) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return nullSQLDecimal128(s.kind)
	}
	return NewSQLDecimal128(s.kind, decimal.NewFromFloat(s.val))
}

// SQLFloat converts the BaseSQLFloat receiver, s, to a SQLFloat.
func (s BaseSQLFloat) SQLFloat() SQLFloat {
	if s.null {
		return nullSQLFloat(s.kind)
	}
	return NewSQLFloat(s.kind, s.val)
}

// SQLInt converts the BaseSQLFloat receiver, s, to a SQLInt.
func (s BaseSQLFloat) SQLInt() SQLInt64 {
	if s.null {
		return nullSQLInt64(s.kind)
	}
	return NewSQLInt64(s.kind, round(s.val))
}

// SQLObjectID converts the BaseSQLFloat receiver, s, to a SQLObjectID.
func (s BaseSQLFloat) SQLObjectID() SQLObjectID {
	return nullSQLObjectID(s.kind)
}

// SQLTimestamp converts the BaseSQLFloat receiver, s, to a SQLTimestamp.
func (s BaseSQLFloat) SQLTimestamp() SQLTimestamp {
	if s.null {
		return nullSQLTimestamp(s.kind)
	}
	t := time.Unix(0, int64(s.val)*1000000)
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(s.kind, t)

}

// SQLUint converts the BaseSQLFloat receiver, s, to a SQLUint64.
func (s BaseSQLFloat) SQLUint() SQLUint64 {
	if s.null {
		return nullSQLUint64(s.kind)
	}
	return NewSQLUint64(s.kind, uint64(round(s.val)))
}

// SQLVarchar converts the BaseSQLFloat receiver, s, to a SQLVarchar.
func (s BaseSQLFloat) SQLVarchar() SQLVarchar {
	if s.null {
		return nullSQLVarchar(s.kind)
	}
	return NewSQLVarchar(s.kind, s.varchar())
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLFloat) EvalType() EvalType {
	return EvalDouble
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLFloat) Value() interface{} {
	if s.null {
		return float64(0)
	}
	return s.val
}

// BaseSQLInt64 represents a 64-bit integer value. BaseSQLInt64 should be
// treated as an abstract type; it should never be used directly during query
// evaluation, and instead should be embedded in other SQLValue implementers who
// can override BaseSQLInt64's default implementations as needed.
type BaseSQLInt64 struct {
	val  int64
	null bool
	kind SQLValueKind
}

var _ translatableToAggregation = (*BaseSQLInt64)(nil)

// iSQLInt64 must be implemented to satisfy the SQLInt64 interface.
func (BaseSQLInt64) iSQLInt64() {}

func newBaseSQLInt64(kind SQLValueKind, val int64) BaseSQLInt64 {
	return BaseSQLInt64{
		val:  val,
		kind: kind,
	}
}

func nullBaseSQLInt64(kind SQLValueKind) BaseSQLInt64 {
	return BaseSQLInt64{
		null: true,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLInt64) IsNull() bool {
	return s.null
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLInt64) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (s BaseSQLInt64) Evaluate(_ context.Context, _ *ExecutionConfig, _ *ExecutionState) (SQLValue, error) {
	return s.SQLInt(), nil
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLInt64) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	if s.null {
		return nil, nil
	}
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
	if s.null {
		return "NULL"
	}
	return strconv.FormatInt(Int64(s), 10)
}

// ToAggregationLanguage translates SQLInt into something that can
// be used in an aggregation pipeline. If SQLInt cannot be translated,
// it will return nil and false.
func (s BaseSQLInt64) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	if s.null {
		return mgoNullLiteral, nil
	}
	return wrapInLiteral(s.Value()), nil
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLInt64) EvalType() EvalType {
	return EvalInt64
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLInt64) Value() interface{} {
	if s.null {
		return int64(0)
	}
	return s.val
}

// SQLBool converts the BaseSQLInt64 receiver, s, to a SQLBool.
func (s BaseSQLInt64) SQLBool() SQLBool {
	if s.null {
		return nullSQLBool(s.kind)
	}
	if s.val == 0 {
		return NewSQLBool(s.kind, false)
	}
	return NewSQLBool(s.kind, true)
}

// SQLDate converts the BaseSQLInt64 receiver, s, to a SQLDate.
func (s BaseSQLInt64) SQLDate() SQLDate {
	if s.null {
		return nullSQLDate(s.kind)
	}
	t := time.Unix(0, s.val*1000000)
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the BaseSQLInt64 receiver, s, to a SQLDecimal128.
func (s BaseSQLInt64) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return nullSQLDecimal128(s.kind)
	}
	return NewSQLDecimal128(s.kind, decimal.New(s.val, 0))
}

// SQLFloat converts the BaseSQLInt64 receiver, s, to a SQLFloat.
func (s BaseSQLInt64) SQLFloat() SQLFloat {
	if s.null {
		return nullSQLFloat(s.kind)
	}
	return NewSQLFloat(s.kind, float64(s.val))
}

// SQLInt converts the BaseSQLInt64 receiver, s, to a SQLInt.
func (s BaseSQLInt64) SQLInt() SQLInt64 {
	if s.null {
		return nullSQLInt64(s.kind)
	}
	return NewSQLInt64(s.kind, s.val)
}

// SQLObjectID converts the BaseSQLInt64 receiver, s, to a SQLObjectID.
func (s BaseSQLInt64) SQLObjectID() SQLObjectID {
	return nullSQLObjectID(s.kind)
}

// SQLTimestamp converts the BaseSQLInt64 receiver, s, to a SQLTimestamp.
func (s BaseSQLInt64) SQLTimestamp() SQLTimestamp {
	if s.null {
		return NewSQLNull(s.kind, EvalTimestamp).(SQLTimestamp)
	}
	t := time.Unix(0, s.val*1000000)
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(s.kind, t)
}

// SQLUint converts the BaseSQLUint64 receiver, s, to a SQLUint64.
func (s BaseSQLInt64) SQLUint() SQLUint64 {
	if s.null {
		return nullSQLUint64(s.kind)
	}
	return NewSQLUint64(s.kind, uint64(s.val))
}

// SQLVarchar converts the BaseSQLInt64 receiver, s, to a SQLVarchar.
func (s BaseSQLInt64) SQLVarchar() SQLVarchar {
	if s.null {
		return nullSQLVarchar(s.kind)
	}
	return NewSQLVarchar(s.kind, s.varchar())
}

// BaseSQLObjectID represents a MongoDB ObjectId using its string value.
// BaseSQLObjectID should be treated as an abstract type; it should never
// be used directly during query evaluation, and instead should be embedded
// in other SQLValue implementers who can override BaseSQLObjectID's default
// implementations as needed.
type BaseSQLObjectID struct {
	val  string
	null bool
	kind SQLValueKind
}

var _ translatableToAggregation = (*BaseSQLObjectID)(nil)

// iSQLObjectID must be implemented to satisfy the SQLObjectID interface.
func (BaseSQLObjectID) iSQLObjectID() {}

func newBaseSQLObjectID(kind SQLValueKind, val string) BaseSQLObjectID {
	return BaseSQLObjectID{
		val:  val,
		kind: kind,
	}
}

func nullBaseSQLObjectID(kind SQLValueKind) BaseSQLObjectID {
	return BaseSQLObjectID{
		null: true,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLObjectID) IsNull() bool {
	return s.null
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLObjectID) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (s BaseSQLObjectID) Evaluate(_ context.Context, _ *ExecutionConfig, _ *ExecutionState) (SQLValue, error) {
	return s.SQLObjectID(), nil
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLObjectID) WireProtocolEncode(charSet *collation.Charset,
	mongoDBVarcharLength int) ([]byte, error) {
	if s.null {
		return nil, nil
	}
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
	if s.null {
		return "NULL"
	}
	return s.val
}

// ToAggregationLanguage translates SQLObjectID into something that can
// be used in an aggregation pipeline. If SQLObjectID cannot be translated,
// it will return nil and false.
func (s BaseSQLObjectID) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	if s.null {
		return mgoNullLiteral, nil
	}
	return wrapInLiteral(bson.ObjectIdHex(s.val)), nil
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLObjectID) EvalType() EvalType {
	return EvalObjectID
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLObjectID) Value() interface{} {
	if s.null {
		return ""
	}
	return bson.ObjectIdHex(s.val)
}

// SQLBool converts a BaseSQLObjectID to a SQLBool.
func (s BaseSQLObjectID) SQLBool() SQLBool {
	return NewSQLBool(s.kind, false)
}

// SQLDate converts the BaseSQLObjectID receiver, s, to a SQLDate.
func (s BaseSQLObjectID) SQLDate() SQLDate {
	if s.IsNull() {
		return nullSQLDate(s.kind)
	}

	return s.SQLTimestamp().SQLDate()
}

// SQLDecimal128 converts a BaseSQLObjectID to a SQLDecimal128 by converting to SQLTimestamp then
// to SQLDecimal128.
func (s BaseSQLObjectID) SQLDecimal128() SQLDecimal128 {
	if s.IsNull() {
		return nullSQLDecimal128(s.kind)
	}
	return s.SQLTimestamp().SQLDecimal128()
}

// SQLFloat converts a BaseSQLObjectID to a SQLFloat by converting to SQLTimestamp then to SQLFloat.
func (s BaseSQLObjectID) SQLFloat() SQLFloat {
	if s.IsNull() {
		return nullSQLFloat(s.kind)
	}
	return s.SQLTimestamp().SQLFloat()
}

// SQLInt converts a BaseSQLObjectID to a SQLInt by converting to SQLTimestamp then to SQLInt.
func (s BaseSQLObjectID) SQLInt() SQLInt64 {
	if s.IsNull() {
		return nullSQLInt64(s.kind)
	}
	return s.SQLTimestamp().SQLInt()
}

// SQLObjectID converts the BaseSQLObjectID receiver, s, to a SQLObjectID.
func (s BaseSQLObjectID) SQLObjectID() SQLObjectID {
	if s.null {
		return nullSQLObjectID(s.kind)
	}
	return NewSQLObjectID(s.kind, s.val)
}

// SQLTimestamp converts the BaseSQLObjectID receiver, s, to a SQLTimestamp.
func (s BaseSQLObjectID) SQLTimestamp() SQLTimestamp {
	if s.IsNull() {
		return nullSQLTimestamp(s.kind)
	}

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
	if s.IsNull() {
		return nullSQLUint64(s.kind)
	}
	return s.SQLTimestamp().SQLUint()
}

// SQLVarchar converts the BaseSQLObjectID receiver, s, to a SQLVarchar.
func (s BaseSQLObjectID) SQLVarchar() SQLVarchar {
	if s.null {
		return nullSQLVarchar(s.kind)
	}
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

var _ translatableToAggregation = (*BaseSQLTimestamp)(nil)

// iSQLTimestamp must be implemented to satisfy the SQLTimestamp interface.
func (BaseSQLTimestamp) iSQLTimestamp() {}

func newBaseSQLTimestamp(kind SQLValueKind, val time.Time) BaseSQLTimestamp {
	return BaseSQLTimestamp{
		datetime: val,
		kind:     kind,
	}
}

func nullBaseSQLTimestamp(kind SQLValueKind) BaseSQLTimestamp {
	return BaseSQLTimestamp{
		null: true,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLTimestamp) IsNull() bool {
	return s.null
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLTimestamp) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (s BaseSQLTimestamp) Evaluate(_ context.Context, _ *ExecutionConfig, _ *ExecutionState) (SQLValue, error) {
	return s.SQLTimestamp(), nil
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
		return util.Slice(s.datetime.Format(schema.TimestampFormatMicros)), nil
	}
	return util.Slice(s.datetime.Format(schema.TimestampFormat)), nil
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

// ToAggregationLanguage translates SQLTimestamp into something that can
// be used in an aggregation pipeline. If SQLTimestamp cannot be translated,
// it will return nil and false.
func (s BaseSQLTimestamp) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	if s.null {
		return mgoNullLiteral, nil
	}
	return wrapInLiteral(s.datetime), nil
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLTimestamp) EvalType() EvalType {
	return EvalDatetime
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLTimestamp) Value() interface{} {
	if s.null {
		return NullDate
	}
	return s.datetime
}

// SQLBool converts the BaseSQLTimestamp receiver, s, to a SQLBool.
func (s BaseSQLTimestamp) SQLBool() SQLBool {
	if s.null {
		return nullSQLBool(s.kind)
	}
	t := s.datetime
	if t == NullDate {
		return NewSQLBool(s.kind, false)
	}
	return NewSQLBool(s.kind, true)
}

// SQLDate converts the BaseSQLTimestamp receiver, s, to a SQLDate.
func (s BaseSQLTimestamp) SQLDate() SQLDate {
	if s.null {
		return nullSQLDate(s.kind)
	}
	t := s.datetime
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the BaseSQLTimestamp receiver, s, to a SQLDecimal128.
func (s BaseSQLTimestamp) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return nullSQLDecimal128(s.kind)
	}
	flt := Float64(s)
	dec := decimal.NewFromFloat(flt)
	return NewSQLDecimal128(s.kind, dec)
}

// SQLFloat converts the BaseSQLTimestamp receiver, s, to a SQLFloat.
func (s BaseSQLTimestamp) SQLFloat() SQLFloat {
	if s.null {
		return nullSQLFloat(s.kind)
	}
	epochMs := s.datetime.UnixNano() / 1000000
	return NewSQLFloat(s.kind, float64(epochMs))
}

// SQLInt converts the BaseSQLTimestamp receiver, s, to a SQLInt.
func (s BaseSQLTimestamp) SQLInt() SQLInt64 {
	if s.null {
		return nullSQLInt64(s.kind)
	}
	epochMs := s.datetime.UnixNano() / 1000000
	return NewSQLInt64(s.kind, epochMs)
}

// SQLObjectID converts the BaseSQLTimestamp receiver, s, to a SQLObjectID.
func (s BaseSQLTimestamp) SQLObjectID() SQLObjectID {
	return nullSQLObjectID(s.kind)
}

// SQLTimestamp converts the BaseSQLTimestamp receiver, s, to a SQLTimestamp.
func (s BaseSQLTimestamp) SQLTimestamp() SQLTimestamp {
	if s.null {
		return nullSQLTimestamp(s.kind)
	}
	return NewSQLTimestamp(s.kind, s.datetime)
}

// SQLUint converts the BaseSQLTimestamp receiver, s, to a SQLUint64.
func (s BaseSQLTimestamp) SQLUint() SQLUint64 {
	if s.null {
		return nullSQLUint64(s.kind)
	}
	epochMs := s.datetime.UnixNano() / 1000000
	return NewSQLUint64(s.kind, uint64(epochMs))
}

// SQLVarchar converts the BaseSQLTimestamp receiver, s, to a SQLVarchar.
func (s BaseSQLTimestamp) SQLVarchar() SQLVarchar {
	if s.null {
		return nullSQLVarchar(s.kind)
	}
	return NewSQLVarchar(s.kind, s.varchar())
}

// BaseSQLUint64 represents an unsigned 64-bit integer. BaseSQLUint64 should be
// treated as an abstract type; it should never be used directly during query
// evaluation, and instead should be embedded in other SQLValue implementers who
// can override BaseSQLUint64's default implementations as needed.
type BaseSQLUint64 struct {
	val  uint64
	null bool
	kind SQLValueKind
}

var _ translatableToAggregation = (*BaseSQLUint64)(nil)

// iSQLUint64 must be implemented to satisfy the SQLUint64 interface.
func (BaseSQLUint64) iSQLUint64() {}

func newBaseSQLUint64(kind SQLValueKind, val uint64) BaseSQLUint64 {
	return BaseSQLUint64{
		val:  val,
		kind: kind,
	}
}

func nullBaseSQLUint64(kind SQLValueKind) BaseSQLUint64 {
	return BaseSQLUint64{
		null: true,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLUint64) IsNull() bool {
	return s.null
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLUint64) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (s BaseSQLUint64) Evaluate(_ context.Context, _ *ExecutionConfig, _ *ExecutionState) (SQLValue, error) {
	return s.SQLUint(), nil
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
	if s.null {
		return "NULL"
	}
	return strconv.FormatUint(s.val, 10)
}

// EvalType returns the SQLType of this SQLValue.
func (s BaseSQLUint64) EvalType() EvalType {
	return EvalUint64
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLUint64) Value() interface{} {
	if s.null {
		return uint64(0)
	}
	return s.val
}

// ToAggregationLanguage translates SQLUint into something that can
// be used in an aggregation pipeline. If SQLUint cannot be translated,
// it will return nil and false.
func (s BaseSQLUint64) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	if s.null {
		return mgoNullLiteral, nil
	}
	val, ok := t.getValue(s)
	if !ok {
		return nil, fmt.Errorf("could not getValue of '%s'", s)
	}

	ui := val.(uint64)
	if ui > math.MaxInt64 {
		return nil, fmt.Errorf("value was greater than max signed integer: %d", ui)
	}
	return wrapInLiteral(val), nil
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLUint64) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	if s.null {
		return nil, nil
	}
	return strconv.AppendUint(nil, s.val, 10), nil
}

// SQLBool converts the BaseSQLUint64 receiver, s, to a SQLBool.
func (s BaseSQLUint64) SQLBool() SQLBool {
	if s.null {
		return nullSQLBool(s.kind)
	}
	if s.val == 0 {
		return NewSQLBool(s.kind, false)
	}
	return NewSQLBool(s.kind, true)
}

// SQLDate converts the BaseSQLUint64 receiver, s, to a SQLDate.
func (s BaseSQLUint64) SQLDate() SQLDate {
	if s.null {
		return nullSQLDate(s.kind)
	}
	if s.val == 0 {
		return NewSQLDate(s.kind, NullDate)
	}

	t, _, ok := parseDateTime(s.varchar())
	if !ok {
		return nullSQLDate(s.kind)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the BaseSQLUint64 receiver, s, to a SQLDecimal128.
func (s BaseSQLUint64) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return nullSQLDecimal128(s.kind)
	}
	return NewSQLDecimal128(s.kind, decimal.New(int64(s.val), 0))
}

// SQLFloat converts the BaseSQLUint64 receiver, s, to a SQLFloat.
func (s BaseSQLUint64) SQLFloat() SQLFloat {
	if s.null {
		return nullSQLFloat(s.kind)
	}
	return NewSQLFloat(s.kind, float64(s.val))
}

// SQLInt converts the BaseSQLUint64 receiver, s, to a SQLInt.
func (s BaseSQLUint64) SQLInt() SQLInt64 {
	if s.null {
		return nullSQLInt64(s.kind)
	}
	return NewSQLInt64(s.kind, int64(s.val))
}

// SQLObjectID converts the BaseSQLUint64 receiver, s, to a SQLObjectID.
func (s BaseSQLUint64) SQLObjectID() SQLObjectID {
	return nullSQLObjectID(s.kind)
}

// SQLTimestamp converts the BaseSQLUint64 receiver, s, to a SQLTimestamp.
func (s BaseSQLUint64) SQLTimestamp() SQLTimestamp {
	if s.null {
		return nullSQLTimestamp(s.kind)
	}
	if s.val == 0 {
		return NewSQLTimestamp(s.kind, NullDate)
	}

	t, _, ok := parseDateTime(s.varchar())
	if !ok {
		return nullSQLTimestamp(s.kind)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLUint converts the BaseSQLUint64 receiver, s, to a SQLUint64.
func (s BaseSQLUint64) SQLUint() SQLUint64 {
	if s.null {
		return nullSQLUint64(s.kind)
	}
	return NewSQLUint64(s.kind, s.val)
}

// SQLVarchar converts the BaseSQLUint64 receiver, s, to a SQLVarchar.
func (s BaseSQLUint64) SQLVarchar() SQLVarchar {
	if s.null {
		return nullSQLVarchar(s.kind)
	}
	return NewSQLVarchar(s.kind, s.varchar())
}

// BaseSQLVarchar represents a string value. BaseSQLVarchar should be treated as
// an abstract type; it should never be used directly during query evaluation,
// and instead should be embedded in other SQLValue implementers who can
// override BaseSQLVarchar's default implementations as needed.
type BaseSQLVarchar struct {
	val  string
	null bool
	kind SQLValueKind
}

var _ translatableToAggregation = (*BaseSQLVarchar)(nil)

// iSQLVarchar must be implemented to satisfy the SQLVarchar interface.
func (BaseSQLVarchar) iSQLVarchar() {}

func newBaseSQLVarchar(kind SQLValueKind, val string) BaseSQLVarchar {
	return BaseSQLVarchar{
		val:  val,
		kind: kind,
	}
}

func nullBaseSQLVarchar(kind SQLValueKind) BaseSQLVarchar {
	return BaseSQLVarchar{
		null: true,
		kind: kind,
	}
}

// IsNull returns true if the SQLValue is null, and false otherwise.
func (s BaseSQLVarchar) IsNull() bool {
	return s.null
}

// Kind returns the SQLValueKind for this SQLValue.
func (s BaseSQLVarchar) Kind() SQLValueKind {
	s.kind.AssertValid()
	return s.kind
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (s BaseSQLVarchar) Evaluate(_ context.Context, _ *ExecutionConfig, _ *ExecutionState) (SQLValue, error) {
	return s.SQLVarchar(), nil
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLVarchar) WireProtocolEncode(charSet *collation.Charset,
	mongoDBVarcharLength int) ([]byte,
	error) {
	if s.null {
		return nil, nil
	}
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
	if s.null {
		return "NULL"
	}
	return s.val
}

// ToAggregationLanguage translates SQLVarchar into something that can
// be used in an aggregation pipeline. If SQLVarchar cannot be translated,
// it will return nil and false.
func (s BaseSQLVarchar) ToAggregationLanguage(t *PushdownTranslator) (interface{}, error) {
	if s.null {
		return mgoNullLiteral, nil
	}
	return wrapInLiteral(s.Value()), nil
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLVarchar) EvalType() EvalType {
	return EvalString
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLVarchar) Value() interface{} {
	if s.null {
		return ""
	}
	return s.val
}

// SQLBool converts the BaseSQLVarchar receiver, s, to a SQLBool.
func (s BaseSQLVarchar) SQLBool() SQLBool {
	if s.null {
		return nullSQLBool(s.kind)
	}
	// Note that we convert to Bool by converting to Int then to Bool,
	// these are the specified semantics of mysql.
	return s.SQLInt().SQLBool()
}

// SQLDate converts the BaseSQLVarchar receiver, s, to a SQLDate.
func (s BaseSQLVarchar) SQLDate() SQLDate {
	if s.null {
		return nullSQLDate(s.kind)
	}
	t, _, ok := parseDateTime(strings.TrimSpace(s.val))
	if !ok {
		return nullSQLDate(s.kind)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the BaseSQLVarchar receiver, s, to a SQLDecimal128.
func (s BaseSQLVarchar) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return nullSQLDecimal128(s.kind)
	}
	out, err := decimal.NewFromString(s.val)
	if err != nil {
		return NewSQLDecimal128(s.kind, decimal.Zero)
	}
	return NewSQLDecimal128(s.kind, out)
}

// SQLFloat converts the BaseSQLVarchar receiver, s, to a SQLFloat.
func (s BaseSQLVarchar) SQLFloat() SQLFloat {
	if s.null {
		return nullSQLFloat(s.kind)
	}
	out, _ := strconv.ParseFloat(s.val, 64)
	return NewSQLFloat(s.kind, out)
}

// SQLInt converts the BaseSQLVarchar receiver, s, to a SQLInt.
func (s BaseSQLVarchar) SQLInt() SQLInt64 {
	if s.null {
		return nullSQLInt64(s.kind)
	}
	out, _ := strconv.ParseInt(s.val, 10, 64)
	return NewSQLInt64(s.kind, out)
}

// SQLObjectID converts the BaseSQLVarchar receiver, s, to a SQLObjectID.
func (s BaseSQLVarchar) SQLObjectID() SQLObjectID {
	if s.null {
		return nullSQLObjectID(s.kind)
	}

	// Return null if this is not a valid ObjectID.
	if len(s.val) == 24 {
		_, err := hex.DecodeString(s.varchar())
		if err == nil {
			return NewSQLObjectID(s.kind, s.val)
		}
	}

	return nullSQLObjectID(s.kind)
}

// SQLTimestamp converts the BaseSQLVarchar receiver, s, to a SQLTimestamp.
func (s BaseSQLVarchar) SQLTimestamp() SQLTimestamp {
	if s.null {
		return nullSQLTimestamp(s.kind)
	}
	t, _, ok := parseDateTime(strings.TrimSpace(s.val))
	if !ok {
		return nullSQLTimestamp(s.kind)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(s.kind, t)
}

// SQLUint converts the BaseSQLVarchar receiver, s, to a SQLUint64.
func (s BaseSQLVarchar) SQLUint() SQLUint64 {
	if s.null {
		return nullSQLUint64(s.kind)
	}
	out, _ := strconv.ParseInt(s.val, 10, 64)
	return NewSQLUint64(s.kind, uint64(out))
}

// SQLVarchar converts the BaseSQLVarchar receiver, s, to a SQLVarchar.
func (s BaseSQLVarchar) SQLVarchar() SQLVarchar {
	if s.null {
		return nullSQLVarchar(s.kind)
	}
	return NewSQLVarchar(s.kind, s.val)
}
