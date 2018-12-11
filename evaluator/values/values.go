package values

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/internal/option"
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
	// Every SQLValue is also a SQLProtocolEncoder.
	SQLProtocolEncoder
	// Every SQLValue is also a SQLValueConverter.
	SQLValueConverter
	// SQLValues have EvalType() and thus implement EvalTyper
	types.EvalTyper
	// SQLValues can all be converted to strings.
	fmt.Stringer
	// IsNull returns true if the SQLValue is null, and false otherwise.
	IsNull() bool
	// Value returns an interface{} that represents the literal value of this SQLValue.
	Value() interface{}
	// Kind returns the SQLValueKind for this SQLValue.
	Kind() SQLValueKind
	// Size returns the size of this SQLValue in bytes.
	Size() uint64
	// CloneWithKind clones the SQLValue with the specified SQLValueKind.
	CloneWithKind(SQLValueKind) SQLValue
}

// NamedSQLValue attaches a name to a values.SQLValue, useful for quickly initialization row values.
type NamedSQLValue struct {
	Name  string
	Value SQLValue
}

// NewNamedSQLValue creates a new Named SQLValue.
func NewNamedSQLValue(name string, value SQLValue) NamedSQLValue {
	return NamedSQLValue{
		Name:  name,
		Value: value,
	}
}

// SQLValueKind is an enum type representing the set of type conversion semantics
// implemented by a given SQLValue.
type SQLValueKind byte

// These are the possible values for SQLValueKind.
const (
	NoSQLValueKind       SQLValueKind = 0x0
	MongoSQLValueKind    SQLValueKind = 0x1
	MySQLValueKind       SQLValueKind = 0x2
	VariableSQLValueKind SQLValueKind = 0x3
)

// AssertValid panics if the SQLValueKind is unknown, and does nothing otherwise.
func (k SQLValueKind) AssertValid() {
	switch k {
	case MySQLValueKind, MongoSQLValueKind, VariableSQLValueKind:
		// valid
	default:
		panic(fmt.Errorf("AssertValid invalid SQLValueKind %x", k))
	}
}

// SQLBool represents a boolean.
type SQLBool interface {
	SQLValue
	iSQLBool()
}

// NewSQLNull returns a null SQLValue of the provided SQLValueKind.
func NewSQLNull(kind SQLValueKind) SQLNull {
	base := newBaseSQLNull(kind)
	switch kind {
	case MySQLValueKind:
		return MySQLNull{base}
	case MongoSQLValueKind:
		return MongoSQLNull{base}
	case VariableSQLValueKind:
		return base
	default:
		panic(fmt.Errorf("newSQLBool unknown SQLValueKind %x", kind))
	}
}

// NewSQLBool returns a new SQLBool with the provided kind and value.
func NewSQLBool(kind SQLValueKind, val bool) SQLBool {
	base := newBaseSQLBool(kind, val)
	switch kind {
	case MySQLValueKind:
		return MySQLBool{base}
	case MongoSQLValueKind:
		return MongoSQLBool{base}
	case VariableSQLValueKind:
		return base
	default:
		panic(fmt.Errorf("newSQLBool unknown SQLValueKind %x", kind))
	}
}

// SQLDate represents a date.
type SQLDate interface {
	SQLValue
	iSQLDate()
}

// NewSQLDate returns a new SQLDate with the provided kind and value.
func NewSQLDate(kind SQLValueKind, val time.Time) SQLDate {
	base := newBaseSQLDate(kind, val)
	switch kind {
	case MySQLValueKind:
		return MySQLDate{base}
	case MongoSQLValueKind:
		return MongoSQLDate{base}
	case VariableSQLValueKind:
		return base
	default:
		panic(fmt.Errorf("newSQLDate unknown SQLValueKind %x", kind))
	}
}

// SQLDecimal128 represents a decimal.
type SQLDecimal128 interface {
	SQLNumber
	iSQLDecimal128()
}

// NewSQLDecimal128 returns a new SQLDecimal128 with the provided kind and value.
func NewSQLDecimal128(kind SQLValueKind, val decimal.Decimal) SQLDecimal128 {
	base := newBaseSQLDecimal128(kind, val)
	switch kind {
	case MySQLValueKind:
		return MySQLDecimal128{base}
	case MongoSQLValueKind:
		return MongoSQLDecimal128{base}
	case VariableSQLValueKind:
		return base
	default:
		panic(fmt.Errorf("newSQLDecimal128 unknown SQLValueKind %x", kind))
	}
}

// SQLFloat represents a float.
type SQLFloat interface {
	SQLNumber
	iSQLFloat()
}

// NewSQLFloat returns a new SQLFloat with the provided kind and value.
func NewSQLFloat(kind SQLValueKind, val float64) SQLFloat {
	base := newBaseSQLFloat(kind, val)
	switch kind {
	case MySQLValueKind:
		return MySQLFloat{base}
	case MongoSQLValueKind:
		return MongoSQLFloat{base}
	case VariableSQLValueKind:
		return base
	default:
		panic(fmt.Errorf("newSQLFloat unknown SQLValueKind %x", kind))
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
	SQLNumber
	iSQLInt64()
}

// NewSQLInt64 returns a new SQLInt64 with the provided kind and value.
func NewSQLInt64(kind SQLValueKind, val int64) SQLInt64 {
	base := newBaseSQLInt64(kind, val)
	switch kind {
	case MySQLValueKind:
		return MySQLInt64{base}
	case MongoSQLValueKind:
		return MongoSQLInt64{base}
	case VariableSQLValueKind:
		return base
	default:
		panic(fmt.Errorf("newSQLInt64 unknown SQLValueKind %x", kind))
	}
}

// SQLObjectID represents a MongoDB ObjectID.
type SQLObjectID interface {
	SQLValue
	iSQLObjectID()
}

// NewSQLObjectID returns a new SQLObjectID with the provided kind and value.
func NewSQLObjectID(kind SQLValueKind, val string) SQLObjectID {
	base := newBaseSQLObjectID(kind, val)
	switch kind {
	case MySQLValueKind:
		return MySQLObjectID{base}
	case MongoSQLValueKind:
		return MongoSQLObjectID{base}
	case VariableSQLValueKind:
		return base
	default:
		panic(fmt.Errorf("newSQLObjectID unknown SQLValueKind %x", kind))
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
	SQLNumber
	iSQLUint64()
}

// NewSQLUint64 returns a new SQLUint64 with the provided kind and value.
func NewSQLUint64(kind SQLValueKind, val uint64) SQLUint64 {
	base := newBaseSQLUint64(kind, val)
	switch kind {
	case MySQLValueKind:
		return MySQLUint64{base}
	case MongoSQLValueKind:
		return MongoSQLUint64{base}
	case VariableSQLValueKind:
		return base
	default:
		panic(fmt.Errorf("newSQLUint64 unknown SQLValueKind %x", kind))
	}
}

// SQLTimestamp represents a timestamp.
type SQLTimestamp interface {
	SQLValue
	iSQLTimestamp()
}

// NewSQLTimestamp returns a new SQLTimestamp with the provided kind and value.
func NewSQLTimestamp(kind SQLValueKind, val time.Time) SQLTimestamp {
	base := newBaseSQLTimestamp(kind, val)
	switch kind {
	case MySQLValueKind:
		return MySQLTimestamp{base}
	case MongoSQLValueKind:
		return MongoSQLTimestamp{base}
	case VariableSQLValueKind:
		return base
	default:
		panic(fmt.Errorf("newSQLTimestamp unknown SQLValueKind %x", kind))
	}
}

// SQLVarchar represents a string.
type SQLVarchar interface {
	SQLValue
	iSQLVarchar()
}

// NewSQLVarchar returns a new SQLVarchar with the provided kind and value.
func NewSQLVarchar(kind SQLValueKind, val string) SQLVarchar {
	base := newBaseSQLVarchar(kind, val)
	switch kind {
	case MySQLValueKind:
		return MySQLVarchar{base}
	case MongoSQLValueKind:
		return MongoSQLVarchar{base}
	case VariableSQLValueKind:
		return base
	default:
		panic(fmt.Errorf("newSQLVarchar unknown SQLValueKind %x", kind))
	}
}

// NewSQLVarcharFromOpt returns a new SQLVarchar with the provided kind.
// If the option.String is a None, the SQLVarchar will be NULL.
// If it is a Some, the SQLVarchar will have the contained string value.
func NewSQLVarcharFromOpt(kind SQLValueKind, opt option.String) SQLVarchar {
	if opt.IsNone() {
		return NewSQLNull(kind)
	}
	return NewSQLVarchar(kind, opt.Unwrap())
}

// SQLNumber represents an SQLValue that is
// also a number, that is Float, Int, Uint, and Decimal128.
type SQLNumber interface {
	SQLValue
	// nolint: megacheck
	iNumber()
}

func (MySQLDecimal128) iNumber() {}
func (MySQLFloat) iNumber()      {}
func (MySQLInt64) iNumber()      {}
func (MySQLUint64) iNumber()     {}

// IsOne returns true if this value can be converted to the Float64 1.
func IsOne(v SQLValue) bool {
	return v.SQLFloat().Value().(float64) == 1.0
}

// IsZero returns true if this value can be converted to the Float64 0.
func IsZero(v SQLValue) bool {
	return v.SQLFloat().Value().(float64) == 0.0
}

// ConvertTo takes a SQLValue v and an evalType, that determines what
// type to convert the passed SQLValue.
func ConvertTo(v SQLValue, evalType types.EvalType) SQLValue {
	if v.IsNull() {
		return v
	}
	switch evalType {
	case types.EvalArrNumeric:
		return v.SQLFloat()
	case types.EvalBoolean:
		return v.SQLBool()
	case types.EvalDecimal128:
		return v.SQLDecimal128()
	case types.EvalDouble:
		return v.SQLFloat()
	case types.EvalInt32:
		return v.SQLInt()
	case types.EvalInt64:
		return v.SQLInt()
	case types.EvalPolymorphic:
		return v
	case types.EvalNull:
		return NewSQLNull(v.Kind())
	case types.EvalObjectID:
		return v.SQLObjectID()
	case types.EvalString:
		return v.SQLVarchar()
	case types.EvalDatetime:
		return v.SQLTimestamp()
	case types.EvalBinary:
		return v.SQLVarchar()
	// Types not corresponding to MongoDB types.
	case types.EvalDate:
		return v.SQLDate()
	case types.EvalUint64:
		return v.SQLUint()
	default:
		panic(fmt.Sprintf("EvalType %x should never be seen as a conversion target", evalType))
	}
}

// HasNullValue returns true if any of the value in values
// is of type SQLNoValue or SQLNullValue.
func HasNullValue(vs ...SQLValue) bool {
	for _, v := range vs {
		if v.IsNull() {
			return true
		}
	}
	return false
}

// SQLValuesSize returns the combined size of these SQLValues in bytes.
func SQLValuesSize(svs ...[]SQLValue) uint64 {
	s := uint64(0)
	for _, sv := range svs {
		for _, v := range sv {
			s += v.Size()
		}
	}

	return s
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

// IsFalsy returns whether a SQLValue is falsy.
func IsFalsy(value SQLValue) bool {
	return !HasNullValue(value) && !Bool(value)
}

var uuidTypes = map[schema.MongoType]struct{}{
	schema.MongoUUID:       {},
	schema.MongoUUIDCSharp: {},
	schema.MongoUUIDJava:   {},
	schema.MongoUUIDOld:    {},
}

// IsUUID returns true if mongoType is of the UUID subtype.
func IsUUID(mongoType schema.MongoType) bool {
	_, ok := uuidTypes[mongoType]
	return ok
}

// reverseByteArray reverses elements in data, beginning
// at start and ending at start + length.
func reverseByteArray(data []byte, start, length int) {
	for left, right := start, start+length-1; left < right; left, right = left+1, right-1 {
		temp := data[left]
		data[left] = data[right]
		data[right] = temp
	}
}

func uuidEncode(data []byte) string {
	return hex.EncodeToString(data[0:4]) + "-" +
		hex.EncodeToString(data[4:6]) + "-" +
		hex.EncodeToString(data[6:8]) + "-" +
		hex.EncodeToString(data[8:10]) + "-" +
		hex.EncodeToString(data[10:16])
}

// BSONValueToSQLValue deserializes raw BSON into SQLValues directly.
func BSONValueToSQLValue(kind SQLValueKind, evalType, uuidSubtype types.EvalType,
	data []byte) (SQLValue, error) {
	switch evalType {
	case types.EvalBoolean:
		if data[0] == 0x0 {
			return NewSQLBool(kind, false), nil
		}
		return NewSQLBool(kind, true), nil
	case types.EvalDecimal128:
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
	case types.EvalDouble:
		ret := math.Float64frombits((uint64(data[0]) << 0) |
			(uint64(data[1]) << 8) |
			(uint64(data[2]) << 16) |
			(uint64(data[3]) << 24) |
			(uint64(data[4]) << 32) |
			(uint64(data[5]) << 40) |
			(uint64(data[6]) << 48) |
			(uint64(data[7]) << 56))
		return NewSQLFloat(kind, ret), nil
	case types.EvalInt32: //32 bit int
		ret := int32((uint32(data[0]) << 0) |
			(uint32(data[1]) << 8) |
			(uint32(data[2]) << 16) |
			(uint32(data[3]) << 24))
		return NewSQLInt64(kind, int64(ret)), nil
	case types.EvalInt64: // 64 bit int
		ret := int64((uint64(data[0]) << 0) |
			(uint64(data[1]) << 8) |
			(uint64(data[2]) << 16) |
			(uint64(data[3]) << 24) |
			(uint64(data[4]) << 32) |
			(uint64(data[5]) << 40) |
			(uint64(data[6]) << 48) |
			(uint64(data[7]) << 56))
		return NewSQLInt64(kind, ret), nil
	case types.EvalObjectID:
		return NewSQLObjectID(kind, hex.EncodeToString(data)), nil
	case types.EvalString:
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
	case types.EvalDatetime: // Date
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
	case types.EvalBinary:
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
			if uuidSubtype == types.EvalJavaUUID {
				reverseByteArray(data, 0, 8)
				reverseByteArray(data, 8, 8)

			} else if uuidSubtype == types.EvalCSharpUUID {
				reverseByteArray(data, 0, 4)
				reverseByteArray(data, 4, 2)
				reverseByteArray(data, 6, 2)
			}
			return NewSQLVarchar(kind, uuidEncode(data)), nil
		}
		// For another other type of BinData we return a SQLNull
		return NewSQLNull(kind), nil
	case types.EvalNull:
		return NewSQLNull(kind), nil
	default:
		// Rather than return an error, we will just return NULL for
		// other types of BSON. This function now only returns error
		// on malformed Decimal128 values.
		return NewSQLNull(kind), nil
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

// SQLNull represents a NULL value.
type SQLNull interface {
	SQLValue
	iSQLBool()
	iSQLDate()
	iSQLDecimal128()
	iSQLFloat()
	iSQLInt32()
	iSQLInt64()
	iSQLObjectID()
	iSQLUint32()
	iSQLUint64()
	iSQLTimestamp()
	iSQLVarchar()
	iSQLNull()
}
