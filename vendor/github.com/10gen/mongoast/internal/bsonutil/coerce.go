package bsonutil

import (
	"fmt"

	"github.com/10gen/mongoast/internal/decimalutil"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// CoerceToBoolean coerces any BSON value to a Boolean.
func CoerceToBoolean(v bsoncore.Value) bool {
	// This function should mirror the behavior of Value::coerceToBool in
	// src/mongo/db/pipeline/value.cpp of the MongoDB server source code.
	switch v.Type {
	case bsontype.CodeWithScope,
		bsontype.MinKey,
		bsontype.DBPointer,
		bsontype.JavaScript,
		bsontype.MaxKey,
		bsontype.String,
		bsontype.EmbeddedDocument,
		bsontype.Array,
		bsontype.Binary,
		bsontype.ObjectID,
		bsontype.DateTime,
		bsontype.Regex,
		bsontype.Symbol,
		bsontype.Timestamp:
		return true
	case bsontype.Null, bsontype.Undefined:
		return false
	case bsontype.Boolean:
		return v.Boolean()
	case bsontype.Int32:
		return v.Int32() != 0
	case bsontype.Int64:
		return v.Int64() != 0
	case bsontype.Double:
		return v.Double() != 0
	case bsontype.Decimal128:
		return !decimalutil.IsZero(decimalutil.FromPrimitive(v.Decimal128()))
	}
	panic(fmt.Errorf("unknown type %s %d", v.Type, v.Type))
}
