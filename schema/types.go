package schema

import (
	"fmt"
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

// GetMongoType converts a string to a MongoType.
func GetMongoType(in string) (MongoType, error) {
	mongoType, mongoTypeExists := mongoTypes[in]

	if !mongoTypeExists {
		return "", fmt.Errorf(`invalid Mongo type "%v"`, in)
	}

	return mongoType, nil
}

// getSQLTypeFromColumnType gets the SQL type from the column time in an alter specification.
func getSQLTypeFromColumnType(colTyp string) (SQLType, error) {
	switch colTyp {
	case "tinyint", "smallint", "mediumint", "int", "integer", "bigint":
		return SQLInt, nil
	case "decimal", "numeric":
		return SQLDecimal, nil
	case "float", "double":
		return SQLFloat, nil
	case "date":
		return SQLDate, nil
	case "datetime", "timestamp":
		return SQLTimestamp, nil
	case "char", "varchar", "binary", "varbinary", "blob", "text":
		return SQLVarchar, nil
	default:
		// This value should never be used due to the error here. The result is not actually
		// polymorphic. SQLInvalidType is not a key in any type conversion maps, which will
		// cause a panic if this type is accidentally used.
		return SQLInvalidType, fmt.Errorf("no SQLType mapping for column type %q", colTyp)
	}
}

// GetMongoTypeFromSQLType gets the MongoType implied for each SQLType.
func GetMongoTypeFromSQLType(colTyp SQLType) MongoType {
	switch colTyp {
	case SQLBoolean:
		return MongoBool
	case SQLInt, SQLUint:
		return MongoInt64
	case SQLDecimal:
		return MongoDecimal128
	case SQLFloat, SQLNumeric:
		return MongoFloat
	case SQLDate, SQLTimestamp:
		return MongoDate
	case SQLVarchar:
		return MongoString
	case SQLNull:
		return MongoNull
	default:
		panic(fmt.Errorf("no mongoType mapping for sqlType %q", colTyp))
	}
}

// A map of mongo type strings to their MongoType constants.
var mongoTypes = map[string]MongoType{
	"array":                   MongoArray,
	"bool":                    MongoBool,
	"bson.Decimal128":         MongoDecimal128,
	"date":                    MongoDate,
	"embedded document":       MongoDocument,
	"mongo.Filter":            MongoFilter,
	"float64":                 MongoFloat,
	"geo.2darray":             MongoGeo2D,
	"int":                     MongoInt,
	"int64":                   MongoInt64,
	"":                        MongoNone,
	"null":                    MongoNull,
	"number":                  MongoNumber,
	"bson.ObjectId":           MongoObjectID,
	"timestamp":               MongoTimestamp,
	"string":                  MongoString,
	"undefined":               MongoUndefined,
	"bson.UUID":               MongoUUID,
	"bson.UUID_Old":           MongoUUIDOld,
	"bson.UUID_Java_Legacy":   MongoUUIDJava,
	"bson.UUID_CSharp_Legacy": MongoUUIDCSharp,
}

// Constants for MongoType - coupled with the `mongoTypes` map.
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

// SQLType is human readable string representation of sql types.
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
	"string":     SQLVarchar,
	"uint32":     SQLUint,
	"uint64":     SQLUint,
}

// GetSQLType converts a string to a SQLType.
func GetSQLType(in string) (SQLType, error) {
	if sqlType, ok := sqlTypeAliases[in]; ok {
		return sqlType, nil
	}

	sqlType, sqlTypeExists := sqlTypes[in]

	if !sqlTypeExists {
		return "", fmt.Errorf(`invalid SQL type "%v"`, in)
	}

	return sqlType, nil
}

// A map of the sql type strings to their SQLType constants.
var sqlTypes = map[string]SQLType{
	"numeric[]": SQLArrNumeric,
	"boolean":   SQLBoolean,
	"date":      SQLDate,
	"decimal":   SQLDecimal,
	"float":     SQLFloat,
	"int":       SQLInt,
	"":          SQLPolymorphic,
	"null":      SQLNull,
	"numeric":   SQLNumeric,
	"objectid":  SQLObjectID,
	"time":      SQLTime,
	"timestamp": SQLTimestamp,
	"uint":      SQLUint,
	"uuid":      SQLUUID,
	"varchar":   SQLVarchar,
}

// Constants for SQLType - coupled with the `sqlTypes` map.
const (
	SQLArrNumeric  SQLType = "numeric[]"
	SQLBoolean     SQLType = "boolean"
	SQLDate        SQLType = "date"
	SQLDecimal     SQLType = "decimal"
	SQLFloat       SQLType = "float"
	SQLInt         SQLType = "int"
	SQLInvalidType SQLType = "invalid type"
	SQLPolymorphic SQLType = ""
	SQLNull        SQLType = "null"
	SQLNumeric     SQLType = "numeric"
	SQLObjectID    SQLType = "objectid"
	SQLTime        SQLType = "time"
	SQLTimestamp   SQLType = "timestamp"
	SQLUint        SQLType = "uint"
	SQLUUID        SQLType = "uuid"
	SQLVarchar     SQLType = "varchar"
)

// GetJSONSchemaTypeFromColumnType gets a json-schema string
// name for the passed SQLType.
func GetJSONSchemaTypeFromColumnType(s SQLType) (string, error) {
	switch s {
	case SQLBoolean:
		return "bool", nil
	case SQLDate, SQLTimestamp:
		return "date", nil
	case SQLDecimal, SQLNumeric:
		return "decimal", nil
	case SQLFloat:
		return "double", nil
	case SQLInt, SQLUint:
		return "long", nil
	case SQLVarchar:
		return "string", nil
	}
	// SQLNull should not be supported here because it is not a valid column type.
	// If it occurs, we want it to be an error so that we can catch it.
	return "", fmt.Errorf("unsupported type '%s' for jsonSchema validator", s)
}

// GetSQLTypeAndMongoTypeFromJSONSchemaType gets the SQLType and MongoType for the given json-schema
// bsonType string.
func GetSQLTypeAndMongoTypeFromJSONSchemaType(s string) (SQLType, MongoType, error) {
	ty := SQLPolymorphic
	switch s {
	case "bool":
		ty = SQLBoolean
	case "date":
		ty = SQLTimestamp
	case "decimal":
		ty = SQLDecimal
	case "double":
		ty = SQLFloat
	case "long":
		ty = SQLInt
	case "string":
		ty = SQLVarchar
	case "null":
		ty = SQLNull
	}
	if ty != SQLPolymorphic {
		return ty, GetMongoTypeFromSQLType(ty), nil
	}
	return SQLPolymorphic, MongoNone, fmt.Errorf("unsupported bsonType '%s' for writeMode jsonSchema validator property", s)
}
