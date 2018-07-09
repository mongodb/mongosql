package evaluator

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/schema"
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
func (s BaseSQLBool) Evaluate(ctx *EvalCtx) (SQLValue, error) {
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

func (s BaseSQLBool) String() string {
	if s.null {
		return "NULL"
	}
	return strconv.FormatFloat(Float64(s), 'f', -1, 64)
}

// ToAggregationLanguage translates SQLBool into something that can
// be used in an aggregation pipeline. If SQLBool cannot be translated,
// it will return nil and false.
func (s BaseSQLBool) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	if s.null {
		return mgoNullLiteral, true
	}
	return wrapInLiteral(Bool(s)), true
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

// SQLBool converts the SQLBool receiver, s, to a SQLBool.
func (s BaseSQLBool) SQLBool() SQLBool {
	if s.null {
		return nullSQLBool(s.kind)
	}
	return NewSQLBool(s.kind, !(s.val == 0))
}

// SQLDate converts the SQLBool receiver, s, to a SQLDate.
func (s BaseSQLBool) SQLDate() SQLDate {
	if s.null {
		return NewSQLNull(s.kind, EvalDate).(SQLDate)
	}
	if s.val == 0 {
		return NewSQLDate(s.kind, NullDate)
	}
	return NewSQLNull(s.kind, EvalDate).(SQLDate)
}

// SQLDecimal128 converts the SQLBool receiver, s, to a SQLDecimal128.
func (s BaseSQLBool) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return NewSQLNull(s.kind, EvalDecimal128).(SQLDecimal128)
	}
	if s.val == 0 {
		return NewSQLDecimal128(s.kind, decimal.Zero)
	}
	return NewSQLDecimal128(s.kind, decimal.New(1, 0))
}

// SQLFloat converts the SQLBool receiver, s, to a SQLFloat.
func (s BaseSQLBool) SQLFloat() SQLFloat {
	if s.null {
		return NewSQLNull(s.kind, EvalDouble).(SQLFloat)
	}
	if s.val == 0 {
		return NewSQLFloat(s.kind, 0.0)
	}
	return NewSQLFloat(s.kind, 1.0)
}

// SQLInt converts the SQLBool receiver, s, to a SQLInt.
func (s BaseSQLBool) SQLInt() SQLInt64 {
	if s.null {
		return NewSQLNull(s.kind, EvalInt64).(SQLInt64)
	}
	if s.val == 0 {
		return NewSQLInt64(s.kind, 0)
	}
	return NewSQLInt64(s.kind, 1)
}

// SQLTimestamp converts the SQLBool receiver, s, to a SQLTimestamp.
func (s BaseSQLBool) SQLTimestamp() SQLTimestamp {
	if s.null {
		return NewSQLNull(s.kind, EvalTimestamp).(SQLTimestamp)
	}
	if s.val == 0 {
		return NewSQLTimestamp(s.kind, NullDate)
	}
	return NewSQLNull(s.kind, EvalTimestamp).(SQLTimestamp)
}

// SQLUint converts the SQLBool receiver, s, to a SQLUint.
func (s BaseSQLBool) SQLUint() SQLUint64 {
	if s.null {
		return NewSQLNull(s.kind, EvalUint64).(SQLUint64)
	}
	if s.val == 0 {
		return NewSQLUint64(s.kind, 0)
	}
	return NewSQLUint64(s.kind, 1)
}

// SQLVarchar converts the SQLBool receiver, s, to a SQLVarchar.
func (s BaseSQLBool) SQLVarchar() SQLVarchar {
	if s.null {
		return NewSQLNull(s.kind, EvalString).(SQLVarchar)
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
func (s BaseSQLDate) Evaluate(ctx *EvalCtx) (SQLValue, error) {
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

func (s BaseSQLDate) String() string {
	if s.null {
		return "NULL"
	}
	return s.datetime.Format("2006-01-02")
}

// ToAggregationLanguage translates SQLDate into something that can
// be used in an aggregation pipeline. If SQLDate cannot be translated,
// it will return nil and false.
func (s BaseSQLDate) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	if s.null {
		return mgoNullLiteral, true
	}
	return wrapInLiteral(s.datetime), true
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

// SQLBool converts the SQLDate receiver, s, to a SQLBool.
func (s BaseSQLDate) SQLBool() SQLBool {
	if s.null {
		return NewSQLNull(s.kind, EvalBoolean).(SQLBool)
	}
	t := s.datetime
	if t == NullDate {
		return NewSQLBool(s.kind, false)
	}
	return NewSQLBool(s.kind, true)
}

// SQLDecimal128 converts the SQLDate receiver, s, to a SQLDecimal128.
func (s BaseSQLDate) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return NewSQLNull(s.kind, EvalDecimal128).(SQLDecimal128)
	}
	flt := Float64(s)
	dec := decimal.NewFromFloat(flt)
	return NewSQLDecimal128(s.kind, dec)
}

// SQLDate converts the SQLDate receiver, s, to a SQLDate.
func (s BaseSQLDate) SQLDate() SQLDate {
	if s.null {
		return NewSQLNull(s.kind, EvalDate).(SQLDate)
	}
	return NewSQLDate(s.kind, s.datetime)
}

// SQLFloat converts the SQLDate receiver, s, to a SQLFloat.
func (s BaseSQLDate) SQLFloat() SQLFloat {
	if s.null {
		return NewSQLNull(s.kind, EvalDouble).(SQLFloat)
	}
	epochMs := s.datetime.UnixNano() / 1000000
	return NewSQLFloat(s.kind, float64(epochMs))
}

// SQLInt converts the SQLDate receiver, s, to a SQLInt.
func (s BaseSQLDate) SQLInt() SQLInt64 {
	if s.null {
		return NewSQLNull(s.kind, EvalInt64).(SQLInt64)
	}
	epochMs := s.datetime.UnixNano() / 1000000
	return NewSQLInt64(s.kind, epochMs)
}

// SQLTimestamp converts the SQLDate receiver, s, to a SQLTimestamp.
func (s BaseSQLDate) SQLTimestamp() SQLTimestamp {
	if s.null {
		return NewSQLNull(s.kind, EvalTimestamp).(SQLTimestamp)
	}
	t := s.datetime
	return NewSQLTimestamp(s.kind, t)
}

// SQLUint converts the SQLDate receiver, s, to a SQLUint.
func (s BaseSQLDate) SQLUint() SQLUint64 {
	if s.null {
		return NewSQLNull(s.kind, EvalUint64).(SQLUint64)
	}
	epochMs := s.datetime.UnixNano() / 1000000
	return NewSQLUint64(s.kind, uint64(epochMs))
}

// SQLVarchar converts the SQLDate receiver, s, to a SQLVarchar.
func (s BaseSQLDate) SQLVarchar() SQLVarchar {
	if s.null {
		return NewSQLNull(s.kind, EvalString).(SQLVarchar)
	}
	return NewSQLVarchar(s.kind, s.String())
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
func (s BaseSQLDecimal128) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return s.SQLDecimal128(), nil
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLDecimal128) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	if s.null {
		return nil, nil
	}
	return []byte(util.FormatDecimal(decimal.Decimal(s.val))), nil
}

// Size returns the size of this SQLValue in bytes.
func (s BaseSQLDecimal128) Size() uint64 {
	return 16
}

func (s BaseSQLDecimal128) String() string {
	if s.null {
		return "NULL"
	}
	return decimal.Decimal(s.val).String()
}

// EvalType returns the SQLType of this SQLValue.
func (BaseSQLDecimal128) EvalType() EvalType {
	return EvalDecimal128
}

// ToAggregationLanguage translates SQLDecimal128 into something that can
// be used in an aggregation pipeline. If SQLDecimal128 cannot be translated,
// it will return nil and false.
func (s BaseSQLDecimal128) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	if s.null {
		return mgoNullLiteral, true
	}
	d, ok := t.translateDecimal(s)
	if !ok {
		return nil, false
	}
	return wrapInLiteral(d), true
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (s BaseSQLDecimal128) Value() interface{} {
	if s.null {
		return decimal.Zero
	}
	return decimal.Decimal(s.val)
}

// SQLBool converts the SQLDecimal128 receiver, s, to a SQLBool.
func (s BaseSQLDecimal128) SQLBool() SQLBool {
	if s.null {
		return NewSQLNull(s.kind, EvalBoolean).(SQLBool)
	}
	if s.val.Equals(decimal.Zero) {
		return NewSQLBool(s.kind, false)
	}
	return NewSQLBool(s.kind, true)
}

// SQLDate converts the SQLDecimal128 receiver, s, to a SQLDate.
func (s BaseSQLDecimal128) SQLDate() SQLDate {
	if s.null {
		return NewSQLNull(s.kind, EvalDate).(SQLDate)
	}
	if s.val.Equals(decimal.Zero) {
		return NewSQLDate(s.kind, NullDate)
	}

	t, _, ok := parseDateTime(s.String())
	if !ok {
		return NewSQLNull(s.kind, EvalDate).(SQLDate)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the SQLDecimal128 receiver, s, to a SQLDecimal128.
func (s BaseSQLDecimal128) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return NewSQLNull(s.kind, EvalDecimal128).(SQLDecimal128)
	}
	return NewSQLDecimal128(s.kind, s.val)
}

// SQLFloat converts the SQLDecimal128 receiver, s, to a SQLFloat.
func (s BaseSQLDecimal128) SQLFloat() SQLFloat {
	if s.null {
		return NewSQLNull(s.kind, EvalDouble).(SQLFloat)
	}
	// Second return value tells us if this is exact, we don't care.
	f, _ := decimal.Decimal(s.val).Float64()
	return NewSQLFloat(s.kind, f)
}

// SQLInt converts the SQLDecimal128 receiver, s, to a SQLInt.
func (s BaseSQLDecimal128) SQLInt() SQLInt64 {
	if s.null {
		return NewSQLNull(s.kind, EvalInt64).(SQLInt64)
	}
	// Do not care if this is exact.
	f, _ := decimal.Decimal(s.val).Float64()
	return NewSQLInt64(s.kind, round(f))
}

// SQLTimestamp converts the SQLDecimal128 receiver, s, to a SQLTimestamp.
func (s BaseSQLDecimal128) SQLTimestamp() SQLTimestamp {
	if s.null {
		return NewSQLNull(s.kind, EvalTimestamp).(SQLTimestamp)
	}
	if s.val.Equals(decimal.Zero) {
		return NewSQLTimestamp(s.kind, NullDate)
	}

	t, _, ok := parseDateTime(s.String())
	if !ok {
		return NewSQLNull(s.kind, EvalTimestamp).(SQLTimestamp)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(s.kind, t)
}

// SQLUint converts the SQLDecimal128 receiver, s, to a SQLUint.
func (s BaseSQLDecimal128) SQLUint() SQLUint64 {
	if s.null {
		return NewSQLNull(s.kind, EvalUint64).(SQLUint64)
	}
	// Do not care if this is exact.
	f, _ := decimal.Decimal(s.val).Float64()
	return NewSQLUint64(s.kind, uint64(round(f)))
}

// SQLVarchar converts the SQLDecimal128 receiver, s, to a SQLVarchar.
func (s BaseSQLDecimal128) SQLVarchar() SQLVarchar {
	if s.null {
		return NewSQLNull(s.kind, EvalString).(SQLVarchar)
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
func (s BaseSQLFloat) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return s.SQLFloat(), nil
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLFloat) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	if s.null {
		return nil, nil
	}
	return strconv.AppendFloat(nil, float64(s.val), 'f', -1, 64), nil
}

// Size returns the size of this SQLValue in bytes.
func (s BaseSQLFloat) Size() uint64 {
	return 8
}

func (s BaseSQLFloat) String() string {
	if s.null {
		return "NULL"
	}
	return strconv.FormatFloat(float64(s.val), 'f', -1, 64)
}

// ToAggregationLanguage translates SQLFloat into something that can
// be used in an aggregation pipeline. If SQLFloat cannot be translated,
// it will return nil and false.
func (s BaseSQLFloat) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	if s.null {
		return mgoNullLiteral, true
	}
	return wrapInLiteral(s.Value()), true
}

// SQLBool converts the SQLFloat receiver, s, to a SQLBool.
func (s BaseSQLFloat) SQLBool() SQLBool {
	if s.null {
		return NewSQLNull(s.kind, EvalBoolean).(SQLBool)
	}
	if s.val == 0 {
		return NewSQLBool(s.kind, false)
	}
	return NewSQLBool(s.kind, true)
}

// SQLDate converts the SQLFloat receiver, s, to a SQLDate.
func (s BaseSQLFloat) SQLDate() SQLDate {
	if s.null {
		return NewSQLNull(s.kind, EvalDate).(SQLDate)
	}
	t := time.Unix(0, int64(s.val)*1000000)
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the SQLFloat receiver, s, to a SQLDecimal128.
func (s BaseSQLFloat) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return NewSQLNull(s.kind, EvalDecimal128).(SQLDecimal128)
	}
	return NewSQLDecimal128(s.kind, decimal.NewFromFloat(float64(s.val)))
}

// SQLFloat converts the SQLFloat receiver, s, to a SQLFloat.
func (s BaseSQLFloat) SQLFloat() SQLFloat {
	if s.null {
		return NewSQLNull(s.kind, EvalDouble).(SQLFloat)
	}
	return NewSQLFloat(s.kind, s.val)
}

// SQLInt converts the SQLFloat receiver, s, to a SQLInt.
func (s BaseSQLFloat) SQLInt() SQLInt64 {
	if s.null {
		return NewSQLNull(s.kind, EvalInt64).(SQLInt64)
	}
	return NewSQLInt64(s.kind, round(float64(s.val)))
}

// SQLTimestamp converts the SQLFloat receiver, s, to a SQLTimestamp.
func (s BaseSQLFloat) SQLTimestamp() SQLTimestamp {
	if s.null {
		return NewSQLNull(s.kind, EvalTimestamp).(SQLTimestamp)
	}
	t := time.Unix(0, int64(s.val)*1000000)
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(s.kind, t)

}

// SQLUint converts the SQLFloat receiver, s, to a SQLUint.
func (s BaseSQLFloat) SQLUint() SQLUint64 {
	if s.null {
		return NewSQLNull(s.kind, EvalUint64).(SQLUint64)
	}
	return NewSQLUint64(s.kind, uint64(round(float64(s.val))))
}

// SQLVarchar converts the SQLFloat receiver, s, to a SQLVarchar.
func (s BaseSQLFloat) SQLVarchar() SQLVarchar {
	if s.null {
		return NewSQLNull(s.kind, EvalString).(SQLVarchar)
	}
	return NewSQLVarchar(s.kind, s.String())
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
	return float64(s.val)
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
func (s BaseSQLInt64) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return s.SQLInt(), nil
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLInt64) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	if s.null {
		return nil, nil
	}
	return strconv.AppendInt(nil, int64(s.val), 10), nil
}

// Size returns the size of this SQLValue in bytes.
func (s BaseSQLInt64) Size() uint64 {
	return 8
}

func (s BaseSQLInt64) String() string {
	if s.null {
		return "NULL"
	}
	return strconv.FormatInt(Int64(s), 10)
}

// ToAggregationLanguage translates SQLInt into something that can
// be used in an aggregation pipeline. If SQLInt cannot be translated,
// it will return nil and false.
func (s BaseSQLInt64) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	if s.null {
		return mgoNullLiteral, true
	}
	return wrapInLiteral(s.Value()), true
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
	return int64(s.val)
}

// SQLBool converts the SQLInt receiver, s, to a SQLBool.
func (s BaseSQLInt64) SQLBool() SQLBool {
	if s.null {
		return NewSQLNull(s.kind, EvalBoolean).(SQLBool)
	}
	if s.val == 0 {
		return NewSQLBool(s.kind, false)
	}
	return NewSQLBool(s.kind, true)
}

// SQLDate converts the SQLInt receiver, s, to a SQLDate.
func (s BaseSQLInt64) SQLDate() SQLDate {
	if s.null {
		return NewSQLNull(s.kind, EvalDate).(SQLDate)
	}
	t := time.Unix(0, s.val*1000000)
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the SQLInt receiver, s, to a SQLDecimal128.
func (s BaseSQLInt64) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return NewSQLNull(s.kind, EvalDecimal128).(SQLDecimal128)
	}
	return NewSQLDecimal128(s.kind, decimal.New(int64(s.val), 0))
}

// SQLFloat converts the SQLInt receiver, s, to a SQLFloat.
func (s BaseSQLInt64) SQLFloat() SQLFloat {
	if s.null {
		return NewSQLNull(s.kind, EvalDouble).(SQLFloat)
	}
	return NewSQLFloat(s.kind, float64(s.val))
}

// SQLInt converts the SQLInt receiver, s, to a SQLInt.
func (s BaseSQLInt64) SQLInt() SQLInt64 {
	if s.null {
		return NewSQLNull(s.kind, EvalInt64).(SQLInt64)
	}
	return NewSQLInt64(s.kind, s.val)
}

// SQLTimestamp converts the SQLInt receiver, s, to a SQLTimestamp.
func (s BaseSQLInt64) SQLTimestamp() SQLTimestamp {
	if s.null {
		return NewSQLNull(s.kind, EvalTimestamp).(SQLTimestamp)
	}
	t := time.Unix(0, s.val*1000000)
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(s.kind, t)
}

// SQLUint converts the SQLUint receiver, s, to a SQLUint.
func (s BaseSQLInt64) SQLUint() SQLUint64 {
	if s.null {
		return NewSQLNull(s.kind, EvalUint64).(SQLUint64)
	}
	return NewSQLUint64(s.kind, uint64(s.val))
}

// SQLVarchar converts the SQLInt receiver, s, to a SQLVarchar.
func (s BaseSQLInt64) SQLVarchar() SQLVarchar {
	if s.null {
		return NewSQLNull(s.kind, EvalString).(SQLVarchar)
	}
	return NewSQLVarchar(s.kind, s.String())
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
func (s BaseSQLTimestamp) Evaluate(ctx *EvalCtx) (SQLValue, error) {
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

func (s BaseSQLTimestamp) String() string {
	if s.null {
		return "NULL"
	}
	return s.datetime.Format("2006-01-02T15:04:05.000Z")
}

// ToAggregationLanguage translates SQLTimestamp into something that can
// be used in an aggregation pipeline. If SQLTimestamp cannot be translated,
// it will return nil and false.
func (s BaseSQLTimestamp) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	if s.null {
		return mgoNullLiteral, true
	}
	return wrapInLiteral(s.datetime), true
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

// SQLBool converts the SQLTimestamp receiver, s, to a SQLBool.
func (s BaseSQLTimestamp) SQLBool() SQLBool {
	if s.null {
		return NewSQLNull(s.kind, EvalBoolean).(SQLBool)
	}
	t := s.datetime
	if t == NullDate {
		return NewSQLBool(s.kind, false)
	}
	return NewSQLBool(s.kind, true)
}

// SQLDate converts the SQLTimestamp receiver, s, to a SQLDate.
func (s BaseSQLTimestamp) SQLDate() SQLDate {
	if s.null {
		return NewSQLNull(s.kind, EvalDate).(SQLDate)
	}
	t := s.datetime
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the SQLTimestamp receiver, s, to a SQLDecimal128.
func (s BaseSQLTimestamp) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return NewSQLNull(s.kind, EvalDecimal128).(SQLDecimal128)
	}
	flt := Float64(s)
	dec := decimal.NewFromFloat(flt)
	return NewSQLDecimal128(s.kind, dec)
}

// SQLFloat converts the SQLTimestamp receiver, s, to a SQLFloat.
func (s BaseSQLTimestamp) SQLFloat() SQLFloat {
	if s.null {
		return NewSQLNull(s.kind, EvalDouble).(SQLFloat)
	}
	epochMs := s.datetime.UnixNano() / 1000000
	return NewSQLFloat(s.kind, float64(epochMs))
}

// SQLInt converts the SQLTimestamp receiver, s, to a SQLInt.
func (s BaseSQLTimestamp) SQLInt() SQLInt64 {
	if s.null {
		return NewSQLNull(s.kind, EvalInt64).(SQLInt64)
	}
	epochMs := s.datetime.UnixNano() / 1000000
	return NewSQLInt64(s.kind, epochMs)
}

// SQLTimestamp converts the SQLTimestamp receiver, s, to a SQLTimestamp.
func (s BaseSQLTimestamp) SQLTimestamp() SQLTimestamp {
	if s.null {
		return NewSQLNull(s.kind, EvalTimestamp).(SQLTimestamp)
	}
	return NewSQLTimestamp(s.kind, s.datetime)
}

// SQLUint converts the SQLTimestamp receiver, s, to a SQLUint.
func (s BaseSQLTimestamp) SQLUint() SQLUint64 {
	if s.null {
		return NewSQLNull(s.kind, EvalUint64).(SQLUint64)
	}
	epochMs := s.datetime.UnixNano() / 1000000
	return NewSQLUint64(s.kind, uint64(epochMs))
}

// SQLVarchar converts the SQLTimestamp receiver, s, to a SQLVarchar.
func (s BaseSQLTimestamp) SQLVarchar() SQLVarchar {
	if s.null {
		return NewSQLNull(s.kind, EvalString).(SQLVarchar)
	}
	return NewSQLVarchar(s.kind, s.String())
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
func (s BaseSQLUint64) Evaluate(_ *EvalCtx) (SQLValue, error) {
	return s.SQLUint(), nil
}

// Size returns the size of this SQLValue in bytes.
func (s BaseSQLUint64) Size() uint64 {
	return 8
}

func (s BaseSQLUint64) String() string {
	if s.null {
		return "NULL"
	}
	return strconv.FormatUint(uint64(s.val), 10)
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
	return uint64(s.val)
}

// ToAggregationLanguage translates SQLUint into something that can
// be used in an aggregation pipeline. If SQLUint cannot be translated,
// it will return nil and false.
func (s BaseSQLUint64) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	if s.null {
		return mgoNullLiteral, true
	}
	val, ok := t.getValue(s)
	if !ok {
		return nil, false
	}

	ui := val.(uint64)
	if ui > math.MaxInt64 {
		return nil, false
	}
	return wrapInLiteral(val), true
}

// WireProtocolEncode returns a byte slice that contains MySQL's wire-protocol
// representation of this SQLValue.
func (s BaseSQLUint64) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	if s.null {
		return nil, nil
	}
	return strconv.AppendUint(nil, uint64(s.val), 10), nil
}

// SQLBool converts the SQLUint receiver, s, to a SQLBool.
func (s BaseSQLUint64) SQLBool() SQLBool {
	if s.null {
		return NewSQLNull(s.kind, EvalBoolean).(SQLBool)
	}
	if s.val == 0 {
		return NewSQLBool(s.kind, false)
	}
	return NewSQLBool(s.kind, true)
}

// SQLDate converts the SQLUint receiver, s, to a SQLDate.
func (s BaseSQLUint64) SQLDate() SQLDate {
	if s.null {
		return NewSQLNull(s.kind, EvalDate).(SQLDate)
	}
	if s.val == 0 {
		return NewSQLDate(s.kind, NullDate)
	}

	t, _, ok := parseDateTime(s.String())
	if !ok {
		return NewSQLNull(s.kind, EvalDate).(SQLDate)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the SQLUint receiver, s, to a SQLDecimal128.
func (s BaseSQLUint64) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return NewSQLNull(s.kind, EvalDecimal128).(SQLDecimal128)
	}
	return NewSQLDecimal128(s.kind, decimal.New(int64(s.val), 0))
}

// SQLFloat converts the SQLUint receiver, s, to a SQLFloat.
func (s BaseSQLUint64) SQLFloat() SQLFloat {
	if s.null {
		return NewSQLNull(s.kind, EvalDouble).(SQLFloat)
	}
	return NewSQLFloat(s.kind, float64(s.val))
}

// SQLInt converts the SQLUint receiver, s, to a SQLInt.
func (s BaseSQLUint64) SQLInt() SQLInt64 {
	if s.null {
		return NewSQLNull(s.kind, EvalInt64).(SQLInt64)
	}
	return NewSQLInt64(s.kind, int64(s.val))
}

// SQLTimestamp converts the SQLUint receiver, s, to a SQLTimestamp.
func (s BaseSQLUint64) SQLTimestamp() SQLTimestamp {
	if s.null {
		return NewSQLNull(s.kind, EvalTimestamp).(SQLTimestamp)
	}
	if s.val == 0 {
		return NewSQLTimestamp(s.kind, NullDate)
	}

	t, _, ok := parseDateTime(s.String())
	if !ok {
		return NewSQLNull(s.kind, EvalTimestamp).(SQLTimestamp)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLUint converts the SQLUint receiver, s, to a SQLUint.
func (s BaseSQLUint64) SQLUint() SQLUint64 {
	if s.null {
		return NewSQLNull(s.kind, EvalUint64).(SQLUint64)
	}
	return NewSQLUint64(s.kind, s.val)
}

// SQLVarchar converts the SQLUint receiver, s, to a SQLVarchar.
func (s BaseSQLUint64) SQLVarchar() SQLVarchar {
	if s.null {
		return NewSQLNull(s.kind, EvalString).(SQLVarchar)
	}
	return NewSQLVarchar(s.kind, s.String())
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
func (s BaseSQLVarchar) Evaluate(_ *EvalCtx) (SQLValue, error) {
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
	// skip the character length check if the byte length is satisactory.
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

func (s BaseSQLVarchar) String() string {
	if s.null {
		return "NULL"
	}
	return string(s.val)
}

// ToAggregationLanguage translates SQLVarchar into something that can
// be used in an aggregation pipeline. If SQLVarchar cannot be translated,
// it will return nil and false.
func (s BaseSQLVarchar) ToAggregationLanguage(t *PushDownTranslator) (interface{}, bool) {
	if s.null {
		return mgoNullLiteral, true
	}
	return wrapInLiteral(s.Value()), true
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
	return string(s.val)
}

// SQLBool converts the SQLVarchar receiver, s, to a SQLBool.
func (s BaseSQLVarchar) SQLBool() SQLBool {
	if s.null {
		return NewSQLNull(s.kind, EvalBoolean).(SQLBool)
	}
	// Note that we convert to Bool by converting to Int then to Bool,
	// these are the specified semantics of mysql.
	return s.SQLInt().SQLBool()
}

// SQLDate converts the SQLVarchar receiver, s, to a SQLDate.
func (s BaseSQLVarchar) SQLDate() SQLDate {
	if s.null {
		return NewSQLNull(s.kind, EvalDate).(SQLDate)
	}
	t, _, ok := parseDateTime(strings.TrimSpace(s.String()))
	if !ok {
		return NewSQLNull(s.kind, EvalDate).(SQLDate)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLDate(
		s.kind,
		time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, schema.DefaultLocale),
	)
}

// SQLDecimal128 converts the SQLVarchar receiver, s, to a SQLDecimal128.
func (s BaseSQLVarchar) SQLDecimal128() SQLDecimal128 {
	if s.null {
		return NewSQLNull(s.kind, EvalDecimal128).(SQLDecimal128)
	}
	out, err := decimal.NewFromString(s.val)
	if err != nil {
		return NewSQLDecimal128(s.kind, decimal.Zero)
	}
	return NewSQLDecimal128(s.kind, out)
}

// SQLFloat converts the SQLVarchar receiver, s, to a SQLFloat.
func (s BaseSQLVarchar) SQLFloat() SQLFloat {
	if s.null {
		return NewSQLNull(s.kind, EvalDouble).(SQLFloat)
	}
	out, _ := strconv.ParseFloat(s.val, 64)
	return NewSQLFloat(s.kind, out)
}

// SQLInt converts the SQLVarchar receiver, s, to a SQLInt.
func (s BaseSQLVarchar) SQLInt() SQLInt64 {
	if s.null {
		return NewSQLNull(s.kind, EvalInt64).(SQLInt64)
	}
	out, _ := strconv.ParseInt(s.val, 10, 64)
	return NewSQLInt64(s.kind, out)
}

// SQLTimestamp converts the SQLVarchar receiver, s, to a SQLTimestamp.
func (s BaseSQLVarchar) SQLTimestamp() SQLTimestamp {
	if s.null {
		return NewSQLNull(s.kind, EvalTimestamp).(SQLTimestamp)
	}
	t, _, ok := parseDateTime(strings.TrimSpace(s.String()))
	if !ok {
		return NewSQLNull(s.kind, EvalTimestamp).(SQLTimestamp)
	}
	t = t.In(schema.DefaultLocale)
	return NewSQLTimestamp(s.kind, t)
}

// SQLUint converts the SQLVarchar receiver, s, to a SQLUint.
func (s BaseSQLVarchar) SQLUint() SQLUint64 {
	if s.null {
		return NewSQLNull(s.kind, EvalUint64).(SQLUint64)
	}
	out, _ := strconv.ParseInt(s.val, 10, 64)
	return NewSQLUint64(s.kind, uint64(out))
}

// SQLVarchar converts the SQLVarchar receiver, s, to a SQLVarchar.
func (s BaseSQLVarchar) SQLVarchar() SQLVarchar {
	if s.null {
		return NewSQLNull(s.kind, EvalString).(SQLVarchar)
	}
	return NewSQLVarchar(s.kind, s.val)
}
