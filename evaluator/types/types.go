package types

import (
	"fmt"

	"github.com/10gen/sqlproxy/schema"
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
	// EvalPolymorphic can represent any underlying type,
	// it is used in situations where types either cannot
	// be determined statically, or where we just do not care
	// what the type is.
	// An example of non-statically determinable type is a case
	// expression:
	//   select case
	//          when x=1 then 3.14
	//          when x=2 then "hello"
	//        end
	//        from foo
	// In fact, the type here can vary from row to row.
	// As an example of a place where we just do not care about the type,
	// consider the arguments to the convert function:
	//   select convert(x, INT) from foo
	// We actually do not need to worry about the type of x at query planning
	// time, any type is viable in this expression. This is conversely to something
	// like
	//   select x + 3 from foo
	// where the type of x needs to be numeric or converted to numeric before
	// the addition occurs.
	EvalPolymorphic EvalType = 0x00
	// MongoDB does not have a Date type, only a Datetime type,
	// we maintain this so that we can correctly drop the time
	// portions for display purposes.
	EvalDate EvalType = 0xFD
	// MongoDB also does not support a Time type, only a Datetime type.
	EvalTime EvalType = 0xFC
	// MongoDB does not have unsigned types, we keep these next two types
	// so that we can correctly display the signed integers as unsigned
	// (e.g., -1 becomes max uint32 or max uint64 as appropriate.)
	EvalUint32 EvalType = 0xFB
	EvalUint64 EvalType = 0xFA
	// The next two types are used for the two uuid subtybe 3 encodings.
	EvalJavaUUID   EvalType = 0xF9
	EvalCSharpUUID EvalType = 0xF8
	// ArrNumeric is for specially handling legacy 2d arrays.
	EvalArrNumeric EvalType = 0xF6
	// EvalNumber does not map to an actual data type value. Therefore,
	// no concrete values can actually have type EvalNumber. We simply use this
	// internally as an expected argument type or a return type for certain
	// scalar functions that can accept any numeric argument or could return
	// any numeric type.
	EvalNumber EvalType = 0xF7
)

// EvalTyper defines an interface that has an EvalType method.
type EvalTyper interface {
	// EvalType returns the EvalType resulting from evaluating the expression, or
	// the type of a SQLValue. (for instance, SQLEqualsExpr.EvalType() returns EvalBoolean).
	EvalType() EvalType
}

var evalNumericTypes = map[EvalType]struct{}{
	EvalDouble:     {},
	EvalInt32:      {},
	EvalInt64:      {},
	EvalDecimal128: {},
	EvalUint32:     {},
	EvalUint64:     {},
	EvalArrNumeric: {},
	EvalNumber:     {},
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

// evalTypeToMongoType returns the schema.MongoType for a byte kind.
var evalTypeToMongoType = map[EvalType]schema.MongoType{
	EvalBoolean:     schema.MongoBool,
	EvalDatetime:    schema.MongoDate,
	EvalDecimal128:  schema.MongoDecimal128,
	EvalDouble:      schema.MongoFloat,
	EvalInt32:       schema.MongoInt,
	EvalInt64:       schema.MongoInt64,
	EvalPolymorphic: schema.MongoNone,
	EvalNull:        schema.MongoNull,
	EvalObjectID:    schema.MongoObjectID,
	EvalBinary:      schema.MongoUUID,
	EvalString:      schema.MongoString,
	EvalArrNumeric:  schema.MongoArray,

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
	schema.SQLArrNumeric:  EvalArrNumeric,
	schema.SQLBoolean:     EvalBoolean,
	schema.SQLDate:        EvalDate,
	schema.SQLDecimal:     EvalDecimal128,
	schema.SQLFloat:       EvalDouble,
	schema.SQLInt:         EvalInt64,
	schema.SQLPolymorphic: EvalPolymorphic,
	schema.SQLNull:        EvalNull,
	schema.SQLObjectID:    EvalObjectID,
	schema.SQLTimestamp:   EvalDatetime,
	schema.SQLUint:        EvalUint64,
	schema.SQLUUID:        EvalBinary,
	schema.SQLVarchar:     EvalString,
	schema.SQLNumeric:     EvalDouble,
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
	EvalArrNumeric:  schema.SQLArrNumeric,
	EvalBoolean:     schema.SQLBoolean,
	EvalDate:        schema.SQLDate,
	EvalTime:        schema.SQLTime,
	EvalDecimal128:  schema.SQLDecimal,
	EvalDouble:      schema.SQLFloat,
	EvalInt32:       schema.SQLInt,
	EvalUint64:      schema.SQLUint,
	EvalInt64:       schema.SQLInt,
	EvalPolymorphic: schema.SQLPolymorphic,
	EvalNull:        schema.SQLNull,
	EvalObjectID:    schema.SQLObjectID,
	EvalDatetime:    schema.SQLTimestamp,
	EvalBinary:      schema.SQLUUID,
	EvalJavaUUID:    schema.SQLUUID,
	EvalCSharpUUID:  schema.SQLUUID,
	EvalString:      schema.SQLVarchar,
}

// EvalTypeToSQLType returns the schema.SQLType for a EvalType.
func EvalTypeToSQLType(e EvalType) schema.SQLType {
	if ret, ok := evalTypeToSQLType[e]; ok {
		return ret
	}
	panic(fmt.Sprintf("unknown EvalType in EvalTypeToSQLType: %x", e))
}

// evalTypeToString returns the name for an EvalType.
var evalTypeToString = map[EvalType]string{
	EvalDouble:       "double",
	EvalString:       "string",
	EvalDocument:     "document",
	EvalArray:        "array",
	EvalBinary:       "binary",
	EvalUndefined:    "undefined",
	EvalObjectID:     "objectID",
	EvalBoolean:      "boolean",
	EvalDatetime:     "datetime",
	EvalNull:         "null",
	EvalRegex:        "regex",
	EvalDBPointer:    "dbPointer",
	EvalJSCode:       "jsCode",
	EvalSymbol:       "symbol",
	EvalJSCodeWScope: "jsCodeWScope",
	EvalInt32:        "int32",
	EvalTimestamp:    "timestamp",
	EvalInt64:        "int64",
	EvalDecimal128:   "decimal128",
	EvalMinKey:       "minKey",
	EvalMaxKey:       "maxKey",
	EvalPolymorphic:  "polymorphic",
	EvalDate:         "date",
	EvalTime:         "time",
	EvalUint32:       "uint32",
	EvalUint64:       "uint64",
	EvalJavaUUID:     "javaUUID",
	EvalCSharpUUID:   "cSharpUUID",
	EvalArrNumeric:   "arrNumeric",
	EvalNumber:       "number",
}

// EvalTypeToString returns the string name of this EvalType.
func EvalTypeToString(e EvalType) string {
	name, ok := evalTypeToString[e]
	if ok {
		return name
	}
	panic(fmt.Sprintf("unknown EvalType in EvalTypeToString: 0x%x", e))
}

// stringToEvalType returns the EvalType for a name.
var stringToEvalType = map[string]EvalType{
	"double":       EvalDouble,
	"string":       EvalString,
	"document":     EvalDocument,
	"array":        EvalArray,
	"binary":       EvalBinary,
	"undefined":    EvalUndefined,
	"objectID":     EvalObjectID,
	"boolean":      EvalBoolean,
	"datetime":     EvalDatetime,
	"null":         EvalNull,
	"regex":        EvalRegex,
	"dbPointer":    EvalDBPointer,
	"jsCode":       EvalJSCode,
	"symbol":       EvalSymbol,
	"jsCodeWScope": EvalJSCodeWScope,
	"int32":        EvalInt32,
	"timestamp":    EvalTimestamp,
	"int64":        EvalInt64,
	"decimal128":   EvalDecimal128,
	"minKey":       EvalMinKey,
	"maxKey":       EvalMaxKey,
	"polymorphic":  EvalPolymorphic,
	"date":         EvalDate,
	"time":         EvalTime,
	"uint32":       EvalUint32,
	"uint64":       EvalUint64,
	"javaUUID":     EvalJavaUUID,
	"cSharpUUID":   EvalCSharpUUID,
	"arrNumeric":   EvalArrNumeric,
	"number":       EvalNumber,
}

// EvalTypeFromString returns the EvalType for a name.
func EvalTypeFromString(s string) EvalType {
	e, ok := stringToEvalType[s]
	if ok {
		return e
	}
	panic(fmt.Sprintf("unknown EvalType name in EvalTypeFromtString: %q", s))
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
	// Polymorphic is less than all types.
	case EvalPolymorphic:
		return true
	case EvalObjectID:
		switch t2 {
		case EvalDecimal128, EvalInt32,
			EvalInt64, EvalUint64, EvalDouble,
			EvalDate, EvalDatetime, EvalString,
			EvalBoolean:
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
	case EvalString:
		switch t2 {
		case EvalDecimal128, EvalInt32,
			EvalInt64, EvalUint64, EvalDouble,
			EvalDate, EvalDatetime:
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
	case EvalDatetime:
		switch t2 {
		case EvalDecimal128, EvalInt32,
			EvalInt64, EvalUint64, EvalDouble:
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
	case EvalUint64:
		switch t2 {
		case EvalDecimal128, EvalDouble:
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
	// Decimal128 is greater than all types.
	case EvalDecimal128:
		return false
	default:
		panic(fmt.Sprintf("Should not have a SQLValue of type %v", t1))
	}
}
