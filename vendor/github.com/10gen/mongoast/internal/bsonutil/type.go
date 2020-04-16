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

func StringToTypeOK(s string) (bsontype.Type, bool) {
	switch s {
	case "double":
		return bsontype.Double, true
	case "string":
		return bsontype.String, true
	case "object":
		return bsontype.EmbeddedDocument, true
	case "array":
		return bsontype.Array, true
	case "binData":
		return bsontype.Binary, true
	case "undefined":
		return bsontype.Undefined, true
	case "objectId":
		return bsontype.ObjectID, true
	case "bool":
		return bsontype.Boolean, true
	case "date":
		return bsontype.DateTime, true
	case "null":
		return bsontype.Null, true
	case "regex":
		return bsontype.Regex, true
	case "dbPointer":
		return bsontype.DBPointer, true
	case "javascript":
		return bsontype.JavaScript, true
	case "symbol":
		return bsontype.Symbol, true
	case "javascriptWithScope":
		return bsontype.CodeWithScope, true
	case "int":
		return bsontype.Int32, true
	case "timestamp":
		return bsontype.Timestamp, true
	case "long":
		return bsontype.Int64, true
	case "decimal":
		return bsontype.Decimal128, true
	case "minKey":
		return bsontype.MinKey, true
	case "maxKey":
		return bsontype.MaxKey, true
	default:
		return 0, false
	}
}

func Int64ToTypeOK(i int64) (bsontype.Type, bool) {
	typ := bsontype.Type(i)
	if (typ >= bsontype.Double && typ <= bsontype.Decimal128) || typ == bsontype.MinKey || typ == bsontype.MaxKey {
		return typ, true
	}
	return 0, false
}
