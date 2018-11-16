package evaluator

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/util/option"
	"github.com/10gen/sqlproxy/schema"

	"github.com/shopspring/decimal"
)

// NullDate is an internal representation for the MySQL
// date 0000-00-00, which cannot be represented as all 0's.
var NullDate = time.Date(-1, 1, 0, 0, 0, 0, 0, schema.DefaultLocale)

// SQLValueConverter defines conversion between
// SQLValue types.
type SQLValueConverter interface {
	// SQLBool() converts the receiver to a SQLBool.
	SQLBool() SQLBool
	// SQLDate() converts the receiver to a SQLDate.
	SQLDate() SQLDate
	// SQLDecimal128() converts the receiver to a SQLDecimal128.
	SQLDecimal128() SQLDecimal128
	// SQLFloat() converts the receiver to a SQLFloat.
	SQLFloat() SQLFloat
	// SQLInt() converts the receiver to a SQLInt.
	SQLInt() SQLInt64
	// SQLObjectID() converts the receiver to a SQLObjectID.
	SQLObjectID() SQLObjectID
	// SQLUint() converts the receiver to a SQLUint64.
	SQLUint() SQLUint64
	// SQLTimestamp() converts the receiver to a SQLTimestamp.
	SQLTimestamp() SQLTimestamp
	// SQLVarchar() converts the receiver to a SQLVarchar.
	SQLVarchar() SQLVarchar
}

// SQLProtocolEncoder is an interface for encoding
// a struct using a SQL wire format.
type SQLProtocolEncoder interface {
	WireProtocolEncode(*collation.Charset, int) ([]byte, error)
}

// SQLValue is a SQLExpr with a value.
type SQLValue interface {
	// Every SQLValue is also a SQLExpr.
	SQLExpr
	// Every SQLValue is also a SQLProtocolEncoder.
	SQLProtocolEncoder
	// Every SQLValue is also a SQLValueConverter.
	SQLValueConverter
	// IsNull returns true if the SQLValue is null, and false otherwise.
	IsNull() bool
	// Value returns an interface{} that represents the literal value of this SQLValue.
	Value() interface{}
	// Kind returns the SQLValueKind for this SQLValue.
	Kind() SQLValueKind
	// Size returns the size of this SQLValue in bytes.
	Size() uint64
}

// SQLValueKind is an enum type representing the set of type conversion semantics
// implemented by a given SQLValue.
type SQLValueKind byte

// These are the possible values for SQLValueKind.
const (
	NoSQLValueKind    SQLValueKind = 0x0
	MongoSQLValueKind SQLValueKind = 0x1
	MySQLValueKind    SQLValueKind = 0x2
)

// AssertValid panics if the SQLValueKind is unknown, and does nothing otherwise.
func (k SQLValueKind) AssertValid() {
	switch k {
	case MySQLValueKind, MongoSQLValueKind:
		// valid
	default:
		panic(fmt.Errorf("invalid SQLValueKind %x", k))
	}
}

// NewSQLNull returns a null SQLValue of the provided SQLValueKind and EvalType.
func NewSQLNull(kind SQLValueKind, typ EvalType) SQLValue {
	switch typ {
	case EvalInt64:
		return nullSQLInt64(kind)
	case EvalUint64:
		return nullSQLUint64(kind)
	case EvalDouble, EvalArrNumeric:
		return nullSQLFloat(kind)
	case EvalString:
		return nullSQLVarchar(kind)
	case EvalDate:
		return nullSQLDate(kind)
	case EvalObjectID:
		return nullSQLObjectID(kind)
	case EvalTimestamp, EvalDatetime:
		return nullSQLTimestamp(kind)
	case EvalBoolean:
		return nullSQLBool(kind)
	case EvalDecimal128:
		return nullSQLDecimal128(kind)
	case EvalNone, EvalTuple, EvalUUID:
		return nullSQLVarchar(kind)
	default:
		panic(fmt.Sprintf("invalid EvalType %x in call to NewSQLNull", typ))
	}
}

// NewSQLNullUntyped returns a new SQLValue of the provided SQLValueKind and with
// the default EvalType (currently EvalVarchar).
func NewSQLNullUntyped(kind SQLValueKind) SQLValue {
	return NewSQLNull(kind, EvalNone)
}

// SQLBool represents a boolean.
type SQLBool interface {
	SQLValue
	iSQLBool()
}

// NewSQLBool returns a new SQLBool with the provided kind and value.
func NewSQLBool(kind SQLValueKind, val bool) SQLBool {
	return newSQLBool(kind, val, false)
}

func nullSQLBool(kind SQLValueKind) SQLBool {
	return newSQLBool(kind, false, true)
}

func newSQLBool(kind SQLValueKind, val bool, null bool) SQLBool {
	var base BaseSQLBool
	if null {
		base = nullBaseSQLBool(kind)
	} else {
		base = newBaseSQLBool(kind, val)
	}
	switch kind {
	case MySQLValueKind:
		return MySQLBool{base}
	case MongoSQLValueKind:
		return MongoSQLBool{base}
	default:
		panic(fmt.Errorf("unknown SQLValueKind %x", kind))
	}
}

// SQLDate represents a date.
type SQLDate interface {
	SQLValue
	iSQLDate()
}

// NewSQLDate returns a new SQLDate with the provided kind and value.
func NewSQLDate(kind SQLValueKind, val time.Time) SQLDate {
	return newSQLDate(kind, val, false)
}

func nullSQLDate(kind SQLValueKind) SQLDate {
	return newSQLDate(kind, NullDate, true)
}

func newSQLDate(kind SQLValueKind, val time.Time, null bool) SQLDate {
	var base BaseSQLDate
	if null {
		base = nullBaseSQLDate(kind)
	} else {
		base = newBaseSQLDate(kind, val)
	}
	switch kind {
	case MySQLValueKind:
		return MySQLDate{base}
	case MongoSQLValueKind:
		return MongoSQLDate{base}
	default:
		panic(fmt.Errorf("unknown SQLValueKind %x", kind))
	}
}

// SQLDecimal128 represents a decimal.
type SQLDecimal128 interface {
	SQLValue
	iSQLDecimal128()
}

// NewSQLDecimal128 returns a new SQLDecimal128 with the provided kind and value.
func NewSQLDecimal128(kind SQLValueKind, val decimal.Decimal) SQLDecimal128 {
	return newSQLDecimal128(kind, val, false)
}

func nullSQLDecimal128(kind SQLValueKind) SQLDecimal128 {
	return newSQLDecimal128(kind, decimal.Zero, true)
}

func newSQLDecimal128(kind SQLValueKind, val decimal.Decimal, null bool) SQLDecimal128 {
	var base BaseSQLDecimal128
	if null {
		base = nullBaseSQLDecimal128(kind)
	} else {
		base = newBaseSQLDecimal128(kind, val)
	}
	switch kind {
	case MySQLValueKind:
		return MySQLDecimal128{base}
	case MongoSQLValueKind:
		return MongoSQLDecimal128{base}
	default:
		panic(fmt.Errorf("unknown SQLValueKind %x", kind))
	}
}

// SQLFloat represents a float.
type SQLFloat interface {
	SQLValue
	iSQLFloat()
}

// NewSQLFloat returns a new SQLFloat with the provided kind and value.
func NewSQLFloat(kind SQLValueKind, val float64) SQLFloat {
	return newSQLFloat(kind, val, false)
}

func nullSQLFloat(kind SQLValueKind) SQLFloat {
	return newSQLFloat(kind, 0, true)
}

func newSQLFloat(kind SQLValueKind, val float64, null bool) SQLFloat {
	var base BaseSQLFloat
	if null {
		base = nullBaseSQLFloat(kind)
	} else {
		base = newBaseSQLFloat(kind, val)
	}
	switch kind {
	case MySQLValueKind:
		return MySQLFloat{base}
	case MongoSQLValueKind:
		return MongoSQLFloat{base}
	default:
		panic(fmt.Errorf("unknown SQLValueKind %x", kind))
	}
}

// SQLInt32 represents an int32.
type SQLInt32 interface {
	SQLValue
	iSQLInt32()
}

// NewSQLInt32 returns a new SQLInt32 with the provided kind and value.
func NewSQLInt32(kind SQLValueKind, arg int32) SQLInt32 {
	panic("SQLInt32 values not currently supported")
}

// SQLInt64 represents an int64.
type SQLInt64 interface {
	SQLValue
	iSQLInt64()
}

// NewSQLInt64 returns a new SQLInt64 with the provided kind and value.
func NewSQLInt64(kind SQLValueKind, val int64) SQLInt64 {
	return newSQLInt64(kind, val, false)
}

func nullSQLInt64(kind SQLValueKind) SQLInt64 {
	return newSQLInt64(kind, 0, true)
}

func newSQLInt64(kind SQLValueKind, val int64, null bool) SQLInt64 {
	var base BaseSQLInt64
	if null {
		base = nullBaseSQLInt64(kind)
	} else {
		base = newBaseSQLInt64(kind, val)
	}
	switch kind {
	case MySQLValueKind:
		return MySQLInt64{base}
	case MongoSQLValueKind:
		return MongoSQLInt64{base}
	default:
		panic(fmt.Errorf("unknown SQLValueKind %x", kind))
	}
}

// SQLObjectID represents a MongoDB ObjectID.
type SQLObjectID interface {
	SQLValue
	iSQLObjectID()
}

// NewSQLObjectID returns a new SQLObjectID with the provided kind and value.
func NewSQLObjectID(kind SQLValueKind, val string) SQLObjectID {
	return newSQLObjectID(kind, val, false)
}

func nullSQLObjectID(kind SQLValueKind) SQLObjectID {
	return newSQLObjectID(kind, "", true)
}

func newSQLObjectID(kind SQLValueKind, val string, null bool) SQLObjectID {
	var base BaseSQLObjectID
	if null {
		base = nullBaseSQLObjectID(kind)
	} else {
		base = newBaseSQLObjectID(kind, val)
	}
	switch kind {
	case MySQLValueKind:
		return MySQLObjectID{base}
	case MongoSQLValueKind:
		return MongoSQLObjectID{base}
	default:
		panic(fmt.Errorf("unknown SQLValueKind %x", kind))
	}
}

// SQLUint32 represents a uint32.
type SQLUint32 interface {
	SQLValue
	iSQLUint32()
}

// NewSQLUint32 returns a new SQLUint32 with the provided kind and value.
func NewSQLUint32(kind SQLValueKind, arg uint32) SQLUint32 {
	panic("SQLUint32 values not currently supported")
}

// SQLUint64 represents a uint64.
type SQLUint64 interface {
	SQLValue
	iSQLUint64()
}

// NewSQLUint64 returns a new SQLUint64 with the provided kind and value.
func NewSQLUint64(kind SQLValueKind, val uint64) SQLUint64 {
	return newSQLUint64(kind, val, false)
}

func nullSQLUint64(kind SQLValueKind) SQLUint64 {
	return newSQLUint64(kind, 0, true)
}

func newSQLUint64(kind SQLValueKind, val uint64, null bool) SQLUint64 {
	var base BaseSQLUint64
	if null {
		base = nullBaseSQLUint64(kind)
	} else {
		base = newBaseSQLUint64(kind, val)
	}
	switch kind {
	case MySQLValueKind:
		return MySQLUint64{base}
	case MongoSQLValueKind:
		return MongoSQLUint64{base}
	default:
		panic(fmt.Errorf("unknown SQLValueKind %x", kind))
	}
}

// SQLTimestamp represents a timestamp.
type SQLTimestamp interface {
	SQLValue
	iSQLTimestamp()
}

// NewSQLTimestamp returns a new SQLTimestamp with the provided kind and value.
func NewSQLTimestamp(kind SQLValueKind, val time.Time) SQLTimestamp {
	return newSQLTimestamp(kind, val, false)
}

func nullSQLTimestamp(kind SQLValueKind) SQLTimestamp {
	return newSQLTimestamp(kind, NullDate, true)
}

func newSQLTimestamp(kind SQLValueKind, val time.Time, null bool) SQLTimestamp {
	var base BaseSQLTimestamp
	if null {
		base = nullBaseSQLTimestamp(kind)
	} else {
		base = newBaseSQLTimestamp(kind, val)
	}
	switch kind {
	case MySQLValueKind:
		return MySQLTimestamp{base}
	case MongoSQLValueKind:
		return MongoSQLTimestamp{base}
	default:
		panic(fmt.Errorf("unknown SQLValueKind %x", kind))
	}
}

// SQLVarchar represents a string.
type SQLVarchar interface {
	SQLValue
	iSQLVarchar()
}

// NewSQLVarchar returns a new SQLVarchar with the provided kind and value.
func NewSQLVarchar(kind SQLValueKind, val string) SQLVarchar {
	return newSQLVarchar(kind, val, false)
}

// NewSQLVarcharFromOpt returns a new SQLVarchar with the provided kind.
// If the option.String is a None, the SQLVarchar will be NULL.
// If it is a Some, the SQLVarchar will have the contained string value.
func NewSQLVarcharFromOpt(kind SQLValueKind, opt option.String) SQLVarchar {
	if opt.IsNone() {
		return nullSQLVarchar(kind)
	}
	return NewSQLVarchar(kind, opt.Unwrap())
}

func nullSQLVarchar(kind SQLValueKind) SQLVarchar {
	return newSQLVarchar(kind, "", true)
}

func newSQLVarchar(kind SQLValueKind, val string, null bool) SQLVarchar {
	var base BaseSQLVarchar
	if null {
		base = nullBaseSQLVarchar(kind)
	} else {
		base = newBaseSQLVarchar(kind, val)
	}
	switch kind {
	case MySQLValueKind:
		return MySQLVarchar{base}
	case MongoSQLValueKind:
		return MongoSQLVarchar{base}
	default:
		panic(fmt.Errorf("unknown SQLValueKind %x", kind))
	}
}

// SQLNumber represents an SQLValue that is
// also a number, that is Float, Int, Uint, and Decimal128.
type SQLNumber interface {
	SQLValue
	iNumber()
}

func (MySQLDecimal128) iNumber() {}
func (MySQLFloat) iNumber()      {}
func (MySQLInt64) iNumber()      {}
func (MySQLUint64) iNumber()     {}

// SQLValues represents multiple sql values.
type SQLValues struct {
	Values []SQLValue
}

var _ translatableToAggregation = (*SQLValues)(nil)

// ExprName returns a string representing this SQLExpr's name.
func (*SQLValues) ExprName() string {
	return "SQLValues"
}

// IsNull returns false, because a SQLValues instance can never be null.
func (*SQLValues) IsNull() bool {
	return false
}

// Kind returns the SQLValueKind of this SQLValue.
func (sv *SQLValues) Kind() SQLValueKind {
	var kind SQLValueKind
	for _, val := range sv.Values {
		if kind == NoSQLValueKind {
			kind = val.Kind()
			continue
		}
		if kind != val.Kind() {
			panic("kinds of values in sqlvalues do not match")
		}
	}

	if kind == NoSQLValueKind {
		panic("no values in sqlvalues")
	}

	return kind
}

// Evaluate evaluates a SQLExpr and returns a SQLValue.
// For a SQLValue, this means that Evaluate is the identity function.
func (sv *SQLValues) Evaluate(_ context.Context, _ *ExecutionConfig, _ *ExecutionState) (SQLValue, error) {
	return sv, nil
}

// Normalize will attempt to change SQLValues into a more recognizeable form that
// may be more amenable to MongoDB's query language.
func (sv *SQLValues) Normalize(_ SQLValueKind) Node {
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
// it will return nil and error.
func (sv *SQLValues) ToAggregationLanguage(t *PushdownTranslator) (interface{}, PushdownFailure) {
	var transExprs []interface{}

	for _, expr := range sv.Values {
		transExpr, err := t.ToAggregationLanguage(expr)
		if err != nil {
			return nil, err
		}
		transExprs = append(transExprs, transExpr)
	}

	return transExprs, nil
}

// EvalType returns the EvalType of this SQLValue.
func (sv *SQLValues) EvalType() EvalType {
	if len(sv.Values) == 1 {
		return sv.Values[0].EvalType()
	} else if len(sv.Values) == 0 {
		return EvalNone
	}

	return EvalTuple
}

// Value returns an interface{} that represents the literal value of this SQLValue.
func (sv *SQLValues) Value() interface{} {
	values := []interface{}{}
	for _, v := range sv.Values {
		values = append(values, v.Value())
	}
	return values
}

// WireProtocolEncode returns a byte slice that contains a wire-protocol
// representation of this SQLValue.
func (sv *SQLValues) WireProtocolEncode(*collation.Charset, int) ([]byte, error) {
	return nil, mysqlerrors.Unknownf("unsupported type %T for wire protocol", sv)
}

// ConvertTo converts the *SQLValues receiver, s, to the specified EvalType.
func (sv *SQLValues) ConvertTo(evalType EvalType) SQLValue {
	return ConvertTo(sv.Values[0], evalType)
}

// SQLBool converts the *SQLValues receiver, s, to a SQLBool.
func (sv *SQLValues) SQLBool() SQLBool {
	return sv.Values[0].SQLBool()
}

// SQLDate converts the *SQLValues receiver, s, to a SQLDate.
func (sv *SQLValues) SQLDate() SQLDate {
	return sv.Values[0].SQLDate()
}

// SQLDecimal128 converts the *SQLValues receiver, s, to a SQLDecimal128.
func (sv *SQLValues) SQLDecimal128() SQLDecimal128 {
	return sv.Values[0].SQLDecimal128()
}

// SQLFloat converts the *SQLValues receiver, s, to a SQLFloat.
func (sv *SQLValues) SQLFloat() SQLFloat {
	return sv.Values[0].SQLFloat()
}

// SQLInt converts the *SQLValues receiver, s, to a SQLInt.
func (sv *SQLValues) SQLInt() SQLInt64 {
	return sv.Values[0].SQLInt()
}

// SQLObjectID converts the *SQLValues receiver, s, to a SQLObjectID.
func (sv *SQLValues) SQLObjectID() SQLObjectID {
	return sv.Values[0].SQLObjectID()
}

// SQLTimestamp converts the *SQLValues receiver, s, to a SQLTimestamp.
func (sv *SQLValues) SQLTimestamp() SQLTimestamp {
	return sv.Values[0].SQLTimestamp()
}

// SQLUint converts the *SQLValues receiver, s, to a SQLUint64.
func (sv *SQLValues) SQLUint() SQLUint64 {
	return sv.Values[0].SQLUint()
}

// SQLVarchar converts the *SQLValues receiver, s, to a SQLVarchar.
func (sv *SQLValues) SQLVarchar() SQLVarchar {
	values := make([]string, len(sv.Values))
	for i, n := range sv.Values {
		if n, ok := n.SQLVarchar().(SQLVarchar); ok {
			values[i] = String(n)
		}
		values[i] = "NULL"
	}
	return NewSQLVarchar(sv.Kind(), strings.Join(values, ", "))
}

// CompareTo compares two SQLValues. It returns -1 if
// left compares less than right; 1, if left compares
// greater than right; and 0 if left compares equal to
// right.
func CompareTo(left, right SQLValue, collation *collation.Collation) (int, error) {
	if left.Kind() != right.Kind() {
		err := fmt.Errorf(
			"left and right SQLValues are not of same kind (%x and %x, respectively)",
			left.Kind(), right.Kind(),
		)
		panic(err)
	}
	valueKind := left.Kind()

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

	if right.IsNull() {
		if left.IsNull() {
			return 0, nil
		}
		i, err := CompareTo(right, left, collation)
		return -i, err
	}

	if left.IsNull() {
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
	}

	if left.EvalType() == right.EvalType() {
		switch leftVal := left.(type) {
		case SQLDate, SQLDecimal128, SQLFloat, SQLInt64, SQLUint64, SQLTimestamp:
			return compareDecimal128(Decimal(left), Decimal(right))
		case SQLVarchar:
			rightVal, _ := right.(SQLVarchar)
			return collation.CompareString(String(leftVal), String(rightVal)), nil
		}
	}

	switch lVal := left.(type) {
	case SQLVarchar:
		if right.IsNull() {
			return 1, nil
		}
		switch right.(type) {
		case SQLDate, SQLTimestamp:
			// MySQL throws an error if you try to compare varchar =,<,> date/timestamp.
			// It works the other way around, however (i.e. date/timestamp =,<,> varchar).
			return -1, fmt.Errorf("Illegal mix of collations %T and %T", left, right)
		default:
			return compareDecimal128(Decimal(left), Decimal(right))
		}
	case SQLDate:
		if right.IsNull() {
			return 1, nil
		}
		switch rVal := right.(type) {
		case SQLVarchar:
			t, _, ok := parseDateTime(right.String())
			if !ok {
				t, _, _ = parseDateTime("0001-01-01")
			}
			return compareFloats(Float64(left), Float64(NewSQLDate(valueKind, t)))
		case SQLTimestamp:
			if Timestamp(rVal).Before(Timestamp(lVal)) {
				return 1, nil
			} else if Timestamp(rVal).After(Timestamp(lVal)) {
				return -1, nil
			}
			return 0, nil
		default:
			return compareDecimal128(Decimal(left), Decimal(right))
		}
	case SQLTimestamp:
		if right.IsNull() {
			return 1, nil
		}
		switch rVal := right.(type) {
		case SQLVarchar:
			t, _, ok := parseDateTime(right.String())
			if !ok {
				t, _, _ = parseDateTime("0001-01-01 00:00:00")
			}
			return compareFloats(Float64(left), Float64(NewSQLTimestamp(valueKind, t)))
		case SQLDate:
			if Timestamp(rVal).Before(Timestamp(lVal)) {
				return 1, nil
			} else if Timestamp(rVal).After(Timestamp(lVal)) {
				return -1, nil
			}
			return 0, nil
		default:
			return compareDecimal128(Decimal(left), Decimal(right))
		}
	default:
		if right.IsNull() {
			return 1, nil
		}
		switch right.(type) {
		default:
			return compareDecimal128(Decimal(left), Decimal(right))
		}
	}
}

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
		return NewSQLNull(v.Kind(), evalType)
	case EvalObjectID:
		return v.SQLObjectID()
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

// Bool converts a SQLValue to a bool.
func Bool(v SQLValue) bool {
	return v.SQLBool().Value().(bool)
}

// Decimal converts a SQLValue to a decimal128.
func Decimal(v SQLValue) decimal.Decimal {
	return v.SQLDecimal128().Value().(decimal.Decimal)
}

// Float64 converts a SQLValue to a float64.
func Float64(v SQLValue) float64 {
	return v.SQLFloat().Value().(float64)
}

// Int64 converts a SQLValue to an int64.
func Int64(v SQLValue) int64 {
	return v.SQLInt().Value().(int64)
}

// Uint64 converts a SQLValue to a uint64.
func Uint64(v SQLValue) uint64 {
	return v.SQLUint().Value().(uint64)
}

// Timestamp converts a SQLValue to a time.Time.
func Timestamp(v SQLValue) time.Time {
	return v.SQLTimestamp().Value().(time.Time)
}

// String converts a SQLValue to a string.
func String(v SQLValue) string {
	return v.SQLVarchar().Value().(string)
}

// BSONValueToSQLValue deserializes raw BSON into SQLValues directly.
func BSONValueToSQLValue(kind SQLValueKind, evalType, uuidSubtype EvalType,
	data []byte) (SQLValue, error) {
	switch evalType {
	case EvalBoolean:
		if data[0] == 0x0 {
			return NewSQLBool(kind, false), nil
		}
		return NewSQLBool(kind, true), nil
	case EvalDecimal128:
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
		return NewSQLDecimal128(kind, gd), nil
	case EvalDouble:
		ret := math.Float64frombits((uint64(data[0]) << 0) |
			(uint64(data[1]) << 8) |
			(uint64(data[2]) << 16) |
			(uint64(data[3]) << 24) |
			(uint64(data[4]) << 32) |
			(uint64(data[5]) << 40) |
			(uint64(data[6]) << 48) |
			(uint64(data[7]) << 56))
		return NewSQLFloat(kind, ret), nil
	case EvalInt32: //32 bit int
		ret := int32((uint32(data[0]) << 0) |
			(uint32(data[1]) << 8) |
			(uint32(data[2]) << 16) |
			(uint32(data[3]) << 24))
		return NewSQLInt64(kind, int64(ret)), nil
	case EvalInt64: // 64 bit int
		ret := int64((uint64(data[0]) << 0) |
			(uint64(data[1]) << 8) |
			(uint64(data[2]) << 16) |
			(uint64(data[3]) << 24) |
			(uint64(data[4]) << 32) |
			(uint64(data[5]) << 40) |
			(uint64(data[6]) << 48) |
			(uint64(data[7]) << 56))
		return NewSQLInt64(kind, ret), nil
	case EvalObjectID:
		return NewSQLObjectID(kind, hex.EncodeToString(data)), nil
	case EvalString:
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
		return NewSQLVarchar(kind, string(data)), nil
	case EvalDatetime: // Date
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
		return NewSQLTimestamp(kind, t), nil
	case EvalUUID:
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
			return NewSQLVarchar(kind, uuidEncode(data)), nil
		} else if subType == 0x03 {
			if uuidSubtype == EvalJavaUUID {
				reverseByteArray(data, 0, 8)
				reverseByteArray(data, 8, 8)

			} else if uuidSubtype == EvalCSharpUUID {
				reverseByteArray(data, 0, 4)
				reverseByteArray(data, 4, 2)
				reverseByteArray(data, 6, 2)
			}
			return NewSQLVarchar(kind, uuidEncode(data)), nil
		}
		return nil,
			fmt.Errorf("unexpected UUID subtype: %#02x", subType)
	case EvalNull:
		return NewSQLNullUntyped(kind), nil
	default:
		return nil, fmt.Errorf("unexpected bson type: found '%s'", EvalTypeToMongoType(evalType))
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
		//  says all of these values are out of range.
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

// divmod is a helper function for SQLDecimal128 values that performs
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
