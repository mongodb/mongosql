package evaluator

import (
	"fmt"
	"time"

	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
)

// EvalType are the types used in the evaluator. They are bytes for efficiency
// where type conversions are needed. Where a EvalType corresponds
// directly to a type from the BSON spec, we use the same byte value.
// See the bsonspec for more info: http://bsonspec.org/spec.html.
type EvalType byte

// Constants for EvalType.
const (
	// The first group here corresponds directly to BSON types. They use
	// the same value for efficient decoding of raw BSON documents.
	EvalDouble       EvalType = 0x01
	EvalString       EvalType = 0x02
	EvalDocument     EvalType = 0x03
	EvalArray        EvalType = 0x04
	EvalBinary       EvalType = 0x05
	EvalUndefined    EvalType = 0x06
	EvalObjectID     EvalType = 0x07
	EvalBoolean      EvalType = 0x08
	EvalDatetime     EvalType = 0x09
	EvalNull         EvalType = 0x0A
	EvalRegex        EvalType = 0x0B
	EvalDBPointer    EvalType = 0x0C
	EvalJSCode       EvalType = 0x0D
	EvalSymbol       EvalType = 0x0E
	EvalJSCodeWScope EvalType = 0x0F
	EvalInt32        EvalType = 0x10
	EvalTimestamp    EvalType = 0x11
	EvalInt64        EvalType = 0x12
	EvalDecimal128   EvalType = 0x13
	EvalMinKey       EvalType = 0xFF
	EvalMaxKey       EvalType = 0x7F
	// Types not corresponding directly to BSON types.
	//-------
	EvalNone       EvalType = 0x00
	EvalDate       EvalType = 0xFD
	EvalTime       EvalType = 0xFC
	EvalUint32     EvalType = 0xFB
	EvalUint64     EvalType = 0xFA
	EvalJavaUUID   EvalType = 0xF9
	EvalCSharpUUID EvalType = 0xF8
	EvalTuple      EvalType = 0xF7
	// ArrNumeric is for specially handling legacy 2d arrays
	EvalArrNumeric EvalType = 0xF6
)

var evalNumericTypes = map[EvalType]struct{}{
	EvalDouble:     {},
	EvalInt32:      {},
	EvalInt64:      {},
	EvalDecimal128: {},
	EvalUint32:     {},
	EvalUint64:     {},
	EvalArrNumeric: {},
}

// IsNumeric returns true if this type is numeric, otherwise
// false.
func (e EvalType) IsNumeric() bool {
	_, ok := evalNumericTypes[e]
	return ok
}

var evalDateTypes = map[EvalType]struct{}{
	EvalDate:     {},
	EvalDatetime: {},
	// Not used yet.
	EvalTime: {},
}

// IsDate returns true if this type is numeric, otherwise
// false.
func (e EvalType) IsDate() bool {
	if _, ok := evalDateTypes[e]; ok {
		return true
	}
	return false
}

// ZeroValue returns the zero value for the given EvalType receiver.
func (e EvalType) ZeroValue(kind SQLValueKind) SQLValue {
	switch e {
	case EvalInt32, EvalInt64:
		return NewSQLInt64(kind, 0)
	case EvalUint32, EvalUint64:
		return NewSQLUint64(kind, 0)
	case EvalDouble, EvalArrNumeric:
		return NewSQLFloat(kind, 0)
	case EvalObjectID:
		return NewSQLObjectID(kind, "")
	case EvalString:
		return NewSQLVarchar(kind, "")
	case EvalDate:
		return NewSQLDate(kind, time.Time{})
	case EvalTimestamp, EvalDatetime:
		return NewSQLTimestamp(kind, time.Time{})
	case EvalBoolean:
		return NewSQLBool(kind, false)
	case EvalBinary:
		return NewSQLVarchar(kind, "")
	case EvalDecimal128:
		return NewSQLDecimal128(kind, decimal.Decimal{})
	case EvalNone:
		return NewSQLNullUntyped(kind)
	default:
		panic(fmt.Sprintf("invalid EvalType %x in call to ZeroValue", e))
	}
}

// evalTypeToMongoType returns the schema.MongoType for a byte kind.
var evalTypeToMongoType = map[EvalType]schema.MongoType{
	EvalBoolean:    schema.MongoBool,
	EvalDatetime:   schema.MongoDate,
	EvalDecimal128: schema.MongoDecimal128,
	EvalDouble:     schema.MongoFloat,
	EvalInt32:      schema.MongoInt,
	EvalInt64:      schema.MongoInt64,
	EvalNone:       schema.MongoNone,
	EvalNull:       schema.MongoNull,
	EvalObjectID:   schema.MongoObjectID,
	EvalBinary:     schema.MongoUUID,
	EvalString:     schema.MongoString,
	EvalArrNumeric: schema.MongoArray,

	EvalRegex:        schema.MongoNull,
	EvalDBPointer:    schema.MongoNull,
	EvalJSCode:       schema.MongoString,
	EvalSymbol:       schema.MongoString,
	EvalJSCodeWScope: schema.MongoString,
	EvalMinKey:       schema.MongoNull,
	EvalMaxKey:       schema.MongoNull,

	EvalDate:       schema.MongoDate,
	EvalTime:       schema.MongoDate,
	EvalUint32:     schema.MongoInt,
	EvalUint64:     schema.MongoInt64,
	EvalJavaUUID:   schema.MongoUUIDJava,
	EvalCSharpUUID: schema.MongoUUIDCSharp,
	EvalTuple:      schema.MongoArray,
	// mappings that indicate presence of dirty/unsupported data.
	EvalDocument:  schema.MongoDocument,
	EvalArray:     schema.MongoArray,
	EvalUndefined: schema.MongoUndefined,
	EvalTimestamp: schema.MongoTimestamp,
}

// EvalTypeToMongoType returns the schema.MongoType for an EvalType.
func EvalTypeToMongoType(e EvalType) schema.MongoType {
	if ret, ok := evalTypeToMongoType[e]; ok {
		return ret
	}
	panic(fmt.Sprintf("unknown EvalType in EvalTypeToMongoType: 0x%x", e))
}

// sqlTypeToEvalType returns the EvalType kind for a schema.SQLType.
var sqlTypeToEvalType = map[schema.SQLType]EvalType{
	schema.SQLArrNumeric: EvalArrNumeric,
	schema.SQLBoolean:    EvalBoolean,
	schema.SQLDate:       EvalDate,
	schema.SQLDecimal:    EvalDecimal128,
	schema.SQLFloat:      EvalDouble,
	schema.SQLInt:        EvalInt64,
	schema.SQLNone:       EvalNone,
	schema.SQLNull:       EvalNull,
	schema.SQLObjectID:   EvalObjectID,
	schema.SQLTimestamp:  EvalDatetime,
	schema.SQLUint:       EvalUint64,
	schema.SQLUUID:       EvalBinary,
	schema.SQLVarchar:    EvalString,
	schema.SQLNumeric:    EvalDouble,
}

// SQLTypeToEvalType returns the EvalType kind for a schema.SQLType.
func SQLTypeToEvalType(s schema.SQLType) EvalType {
	if ret, ok := sqlTypeToEvalType[s]; ok {
		return ret
	}
	panic(fmt.Sprintf("unknown schema.SQLType in SQLTypeToEvalType: %s", s))
}

// evalTypeToSQLType returns the schema.SQLType for a EvalType.
var evalTypeToSQLType = map[EvalType]schema.SQLType{
	EvalArrNumeric: schema.SQLArrNumeric,
	EvalBoolean:    schema.SQLBoolean,
	EvalDate:       schema.SQLDate,
	EvalTime:       schema.SQLTime,
	EvalDecimal128: schema.SQLDecimal,
	EvalDouble:     schema.SQLFloat,
	EvalInt32:      schema.SQLInt,
	EvalUint64:     schema.SQLUint,
	EvalInt64:      schema.SQLInt,
	EvalNone:       schema.SQLNone,
	EvalNull:       schema.SQLNull,
	EvalObjectID:   schema.SQLObjectID,
	EvalDatetime:   schema.SQLTimestamp,
	EvalBinary:     schema.SQLUUID,
	EvalJavaUUID:   schema.SQLUUID,
	EvalCSharpUUID: schema.SQLUUID,
	EvalString:     schema.SQLVarchar,
	EvalTuple:      schema.SQLTuple,
}

// EvalTypeToSQLType returns the schema.SQLType for a EvalType.
func EvalTypeToSQLType(e EvalType) schema.SQLType {
	if ret, ok := evalTypeToSQLType[e]; ok {
		return ret
	}
	panic(fmt.Sprintf("unknown EvalType in EvalTypeToSQLType: %x", e))
}

// EvalTypeSorter is a type used for sorting EvalTypes according to some
// configurable sorting rules.
type EvalTypeSorter struct {
	Types               []EvalType
	VarcharHighPriority bool
}

func (s *EvalTypeSorter) Len() int {
	return len(s.Types)
}

func (s *EvalTypeSorter) Swap(i, j int) {
	s.Types[i], s.Types[j] = s.Types[j], s.Types[i]
}

func (s *EvalTypeSorter) Less(i, j int) bool {

	t1 := s.Types[i]
	t2 := s.Types[j]

	if s.VarcharHighPriority {
		if t1 == EvalString {
			return false
		} else if t2 == EvalString {
			return true
		}
	}

	switch t1 {
	case EvalNone, EvalNull:
		return true
	case EvalString:
		switch t2 {
		case EvalDecimal128, EvalInt32,
			EvalInt64, EvalUint64, EvalDouble,
			EvalDate, EvalDatetime:
			return true
		default:
			return false
		}
	case EvalBoolean:
		switch t2 {
		case EvalDecimal128, EvalInt32,
			EvalInt64, EvalUint64, EvalDouble,
			EvalDate, EvalDatetime, EvalString:
			return true
		default:
			return false
		}
	case EvalInt32, EvalInt64:
		switch t2 {
		case EvalDecimal128, EvalDouble, EvalUint64:
			return true
		default:
			return false
		}
	case EvalDatetime:
		switch t2 {
		case EvalDecimal128, EvalInt32,
			EvalInt64, EvalUint64, EvalDouble:
			return true
		default:
			return false
		}
	case EvalUint64:
		switch t2 {
		case EvalDecimal128, EvalDouble:
			return true
		default:
			return false
		}
	case EvalDate:
		switch t2 {
		case EvalDecimal128, EvalInt32, EvalInt64,
			EvalUint64, EvalDouble, EvalDatetime:
			return true
		default:
			return false
		}
	case EvalDouble:
		switch t2 {
		case EvalDecimal128:
			return true
		default:
			return false
		}
	}
	return false
}

// isBooleanComparable returns true if this EvalType can
// be used directly as a boolean without conversion.
func isBooleanComparable(evalType EvalType) bool {
	return evalType == EvalInt64 || evalType == EvalBoolean ||
		evalType == EvalInt32 || evalType == EvalUint64 ||
		evalType == EvalUint32
}
