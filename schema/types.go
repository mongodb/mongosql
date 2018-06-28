package schema

import (
	"time"
)

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
	MongoArray      MongoType = "array"
	MongoBool       MongoType = "bool"
	MongoDecimal128 MongoType = "bson.Decimal128"
	MongoDate       MongoType = "date"
	MongoDocument   MongoType = "embedded document"
	MongoFilter     MongoType = "mongo.Filter"
	MongoFloat      MongoType = "float64"
	MongoGeo2D      MongoType = "geo.2darray"
	MongoInt        MongoType = "int"
	MongoInt64      MongoType = "int64"
	MongoNone       MongoType = ""
	MongoNull       MongoType = "null"
	MongoNumber     MongoType = "number"
	MongoObjectID   MongoType = "bson.ObjectId"
	MongoTimestamp  MongoType = "timestamp"
	MongoString     MongoType = "string"
	MongoUndefined  MongoType = "undefined"
	MongoUUID       MongoType = "bson.UUID"
	MongoUUIDOld    MongoType = "bson.UUID_Old"
	MongoUUIDJava   MongoType = "bson.UUID_Java_Legacy"
	MongoUUIDCSharp MongoType = "bson.UUID_CSharp_Legacy"
)

// SQLType is human readible string representation of sql types.
type SQLType string

// sqlTypeAliases maps types in drdl files
// to the actual types we use internally,
// where they differ.
var sqlTypeAliases = map[string]SQLType{
	"decimal128": SQLDecimal,
	"float64":    SQLFloat,
	"int32":      SQLInt,
	"int64":      SQLInt,
	"sqluint":    SQLUint,
	"sqluint32":  SQLUint,
	"sqluint64":  SQLUint,
	"sqltuple":   SQLTuple,
	"string":     SQLVarchar,
	"uint32":     SQLUint,
	"uint64":     SQLUint,
}

// GetSQLType converts a string to a SQLType.
func GetSQLType(in string) SQLType {
	if out, ok := sqlTypeAliases[in]; ok {
		return out
	}
	return SQLType(in)
}

// Constants for SQLType.
const (
	SQLArrNumeric SQLType = "numeric[]"
	SQLBoolean    SQLType = "boolean"
	SQLDate       SQLType = "date"
	SQLDecimal    SQLType = "decimal"
	SQLFloat      SQLType = "float"
	SQLInt        SQLType = "int"
	SQLNone       SQLType = ""
	SQLNull       SQLType = "null"
	SQLNumeric    SQLType = "numeric"
	SQLObjectID   SQLType = "objectid"
	SQLTime       SQLType = "time"
	SQLTimestamp  SQLType = "timestamp"
	SQLTuple      SQLType = "tuple"
	SQLUint       SQLType = "uint"
	SQLUUID       SQLType = "uuid"
	SQLVarchar    SQLType = "varchar"
)
