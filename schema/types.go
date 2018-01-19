package schema

import (
	"time"

	"github.com/10gen/mongo-go-driver/bson"
)

// ColumnType is the type of a column.
type ColumnType struct {
	SQLType   SQLType
	MongoType MongoType
}

// NewColumnType returns a *ColumnType with the specified SQLType and MongoType
func NewColumnType(sqlType SQLType, mongoType MongoType) *ColumnType {
	return &ColumnType{
		SQLType:   sqlType,
		MongoType: mongoType,
	}
}

// Constants for various date and time format strings
const (
	DateFormat            = "2006-01-02"
	TimeFormat            = "15:04:05"
	TimestampFormat       = "2006-01-02 15:04:05"
	TimestampFormatMicros = "2006-01-02 15:04:05.000000"
)

// DefaultLocale is the default locale to use when formatting dates and times.
var DefaultLocale = time.UTC

// MongoType is the type that exists in MongoDB.
type MongoType string

// Constants for MongoType.
const (
	MongoBool       MongoType = "bool"
	MongoDecimal128 MongoType = "bson.Decimal128"
	MongoDate       MongoType = "date"
	MongoFilter     MongoType = "mongo.Filter"
	MongoFloat      MongoType = "float64"
	MongoGeo2D      MongoType = "geo.2darray"
	MongoInt        MongoType = "int"
	MongoInt64      MongoType = "int64"
	MongoNone       MongoType = ""
	MongoNumber     MongoType = "number"
	MongoObjectID   MongoType = "bson.ObjectId"
	MongoString     MongoType = "string"
	MongoUUID       MongoType = "bson.UUID"
	MongoUUIDOld    MongoType = "bson.UUID_Old"
	MongoUUIDJava   MongoType = "bson.UUID_Java_Legacy"
	MongoUUIDCSharp MongoType = "bson.UUID_CSharp_Legacy"
)

// SQLType is the type to be used in memory.
type SQLType string

// Constants for SQLType.
const (
	SQLArrNumeric SQLType = "numeric[]"
	SQLBoolean    SQLType = "boolean"
	SQLDate       SQLType = "date"
	SQLDecimal128 SQLType = "decimal128"
	SQLFloat      SQLType = "float64"
	SQLInt        SQLType = "int"
	SQLInt64      SQLType = "int64"
	SQLNone       SQLType = ""
	SQLNull       SQLType = "null"
	SQLNumeric    SQLType = "numeric"
	SQLObjectID   SQLType = "objectid"
	SQLTimestamp  SQLType = "timestamp"
	SQLTuple      SQLType = "sqltuple"
	SQLUint64     SQLType = "sqluint64"
	SQLUUID       SQLType = "uuid"
	SQLVarchar    SQLType = "varchar"
)

// BSONSpecType are BSON types. These are the byte type designators described
// by the BSON spec for a given BSON type, unless there is no such type in BSON
// (e.g., Date, and Time). The bsonspec defines the byte values, see:
// http://bsonspec.org/spec.html.
type BSONSpecType byte

// Constants for BSONSpecType.
const (
	BSONDouble     BSONSpecType = 0x01
	BSONString     BSONSpecType = 0x02
	BSONUUID       BSONSpecType = 0x05
	BSONObjectID   BSONSpecType = 0x07
	BSONBoolean    BSONSpecType = 0x08
	BSONTimestamp  BSONSpecType = 0x09
	BSONNull       BSONSpecType = 0x0A
	BSONInt        BSONSpecType = 0x10
	BSONInt64      BSONSpecType = 0x12
	BSONDecimal128 BSONSpecType = 0x13
	// SQLTypes not corresponding to BSON types.
	BSONNone       BSONSpecType = 0xFF
	BSONDate       BSONSpecType = 0xFE
	BSONTime       BSONSpecType = 0xFD
	BSONUint64     BSONSpecType = 0xFC
	BSONJavaUUID   BSONSpecType = 0xFB
	BSONCSharpUUID BSONSpecType = 0xFA
)

// SQLTypeToBSONType returns the byte kind for a SQLType.
var SQLTypeToBSONType = map[SQLType]BSONSpecType{
	SQLArrNumeric: BSONDouble,
	SQLBoolean:    BSONBoolean,
	SQLDate:       BSONDate,
	SQLDecimal128: BSONDecimal128,
	SQLFloat:      BSONDouble,
	SQLInt:        BSONInt,
	SQLInt64:      BSONInt64,
	SQLNone:       BSONNone,
	SQLNull:       BSONNull,
	SQLObjectID:   BSONObjectID,
	SQLTimestamp:  BSONTimestamp,
	SQLUint64:     BSONUint64,
	SQLUUID:       BSONUUID,
	SQLVarchar:    BSONString,
	SQLNumeric:    BSONDouble,
}

const (
	zeroFloat  = float64(0)
	zeroInt    = int64(0)
	zeroString = ""
)

var (
	zeroDecimal128, _ = bson.ParseDecimal128("0")
	zeroBSON          = bson.ObjectId("")
	zeroTime          = time.Time{}
)

// ZeroValue returns the zero value for sqlType.
func (sqlType SQLType) ZeroValue() interface{} {
	switch sqlType {
	case SQLNumeric, SQLInt, SQLInt64:
		return zeroInt
	case SQLFloat, SQLArrNumeric:
		return zeroFloat
	case SQLVarchar:
		return zeroString
	case SQLTimestamp, SQLDate:
		return zeroTime
	case SQLBoolean:
		return false
	case SQLNone, SQLNull:
		return nil
	case SQLObjectID:
		return zeroBSON
	case SQLDecimal128:
		return zeroDecimal128
	}
	return ""
}

// SQLTypesSorter is a type used for sorting SQLTypes according to some
// configurable sorting rules.
type SQLTypesSorter struct {
	Types               []SQLType
	VarcharHighPriority bool
}

func (s *SQLTypesSorter) Len() int {
	return len(s.Types)
}

func (s *SQLTypesSorter) Swap(i, j int) {
	s.Types[i], s.Types[j] = s.Types[j], s.Types[i]
}

func (s *SQLTypesSorter) Less(i, j int) bool {

	t1 := s.Types[i]
	t2 := s.Types[j]

	if s.VarcharHighPriority {
		if t1 == SQLVarchar {
			return false
		} else if t2 == SQLVarchar {
			return true
		}
	}

	switch t1 {
	case SQLNone, SQLNull:
		return true
	case SQLVarchar:
		switch t2 {
		case SQLDecimal128, SQLInt, SQLInt64, SQLUint64, SQLFloat, SQLNumeric, SQLTimestamp, SQLDate, SQLBoolean:
			return true
		default:
			return false
		}
	case SQLBoolean:
		switch t2 {
		case SQLDecimal128, SQLInt, SQLInt64, SQLUint64, SQLFloat, SQLNumeric, SQLTimestamp, SQLDate:
			return true
		default:
			return false
		}
	case SQLInt, SQLInt64:
		switch t2 {
		case SQLDecimal128, SQLFloat, SQLNumeric, SQLUint64:
			return true
		default:
			return false
		}
	case SQLTimestamp:
		switch t2 {
		case SQLDecimal128, SQLInt, SQLInt64, SQLUint64, SQLFloat, SQLNumeric:
			return true
		default:
			return false
		}
	case SQLUint64:
		switch t2 {
		case SQLDecimal128, SQLFloat:
			return true
		default:
			return false
		}
	case SQLDate:
		switch t2 {
		case SQLDecimal128, SQLInt, SQLInt64, SQLUint64, SQLFloat, SQLNumeric, SQLTimestamp:
			return true
		default:
			return false
		}
	case SQLFloat, SQLNumeric:
		switch t2 {
		case SQLDecimal128:
			return true
		default:
			return false
		}
	}
	return false
}
