package schema

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// ColumnType is the type of a column.
// TODO: this should not be in this package.
type ColumnType struct {
	SQLType   SQLType
	MongoType MongoType
}

const (
	DateFormat            = "2006-01-02"
	TimeFormat            = "15:04:05"
	TimestampFormat       = "2006-01-02 15:04:05"
	TimestampFormatMicros = "2006-01-02 15:04:05.000000"
)

var (
	DefaultLocale = time.UTC

	// TimestampCtorFormats holds the various formats
	// for constructing a SQL timestamp.
	TimestampCtorFormats = []string{
		"15:4:5",
		"2006-1-2",
		"2006-1-2 15",
		"2006-1-2 15:4",
		"2006-1-2 15:4:5",
		"2006-1-2 15:4:5.000",
		"2006-1-2 15:4:5.000000",
		"2006:1:2",
		"2006:1:2 15",
		"2006:1:2 15:4",
		"2006:1:2 15:4:5",
		"2006:1:2 15:4:5.000",
		"2006:1:2 15:4:5.000000",
	}
)

// MongoType is the type that exists in MongoDB.
type MongoType string

// Constants for MongoType.
const (
	MongoBool       MongoType = "bool"
	MongoDecimal128           = "bson.Decimal128"
	MongoDate                 = "date"
	MongoFilter               = "mongo.Filter"
	MongoFloat                = "float64"
	MongoGeo2D                = "geo.2darray"
	MongoInt                  = "int"
	MongoInt64                = "int64"
	MongoNone                 = ""
	MongoObjectId             = "bson.ObjectId"
	MongoString               = "string"
	MongoUUID                 = "bson.UUID"
	MongoUUIDOld              = "bson.UUID_Old"
	MongoUUIDJava             = "bson.UUID_Java_Legacy"
	MongoUUIDCSharp           = "bson.UUID_CSharp_Legacy"
)

// SQLType is the type to be used in memory.
type SQLType string

// Constants for SQLType.
const (
	SQLArrNumeric SQLType = "numeric[]"
	SQLBoolean            = "boolean"
	SQLDate               = "date"
	SQLDecimal128         = "decimal128"
	SQLFloat              = "float64"
	SQLInt                = "int"
	SQLInt64              = "int64"
	SQLNone               = ""
	SQLNull               = "null"
	SQLNumeric            = "numeric"
	SQLObjectID           = "objectid"
	SQLTimestamp          = "timestamp"
	SQLTuple              = "sqltuple"
	SQLUint64             = "sqluint64"
	SQLUUID               = "uuid"
	SQLVarchar            = "varchar"
)

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

type SQLTypes []SQLType

func (types SQLTypes) Len() int {
	return len(types)
}

func (types SQLTypes) Swap(i, j int) {
	types[i], types[j] = types[j], types[i]
}

func (types SQLTypes) Less(i, j int) bool {

	t1 := types[i]
	t2 := types[j]

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
