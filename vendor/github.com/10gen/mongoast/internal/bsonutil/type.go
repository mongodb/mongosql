package bsonutil

import (
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// TypeToString converts a BSON type to a string with the type name as used
// in the MongoDB server.
func TypeToString(t bsontype.Type) string {
	// We can't use the String() method defined by the Go driver because it
	// returns different strings from what the MongoDB server uses for many of
	// the types, which would make our error messages inconsistent. These
	// strings must match what's used in src/mongo/bson/bsontypes.cppp of the
	// MongoDB server source code.
	switch t {
	case bsontype.Double:
		return "double"
	case bsontype.String:
		return "string"
	case bsontype.EmbeddedDocument:
		return "object"
	case bsontype.Array:
		return "array"
	case bsontype.Binary:
		return "binData"
	case bsontype.Undefined:
		return "undefined"
	case bsontype.ObjectID:
		return "objectId"
	case bsontype.Boolean:
		return "bool"
	case bsontype.DateTime:
		return "date"
	case bsontype.Null:
		return "null"
	case bsontype.Regex:
		return "regex"
	case bsontype.DBPointer:
		return "dbPointer"
	case bsontype.JavaScript:
		return "javascript"
	case bsontype.Symbol:
		return "symbol"
	case bsontype.CodeWithScope:
		return "javascriptWithScope"
	case bsontype.Int32:
		return "int"
	case bsontype.Timestamp:
		return "timestamp"
	case bsontype.Int64:
		return "long"
	case bsontype.Decimal128:
		return "decimal"
	case bsontype.MinKey:
		return "minKey"
	case bsontype.MaxKey:
		return "maxKey"
	default:
		return "invalid"
	}
}
