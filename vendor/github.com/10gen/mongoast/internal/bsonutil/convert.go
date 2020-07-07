package bsonutil

import (
	"fmt"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// IsImplementedConvertToType returns false if the specified type is a valid
// but currently unimplemented type for the $convert to parameter.
// Note: unsupported types are inherently implemented since they result in
// errors at evaluation time.
func IsImplementedConvertToType(to bsontype.Type) bool {
	switch to {
	case bsontype.Double,
		bsontype.ObjectID,
		bsontype.Boolean,
		bsontype.DateTime,
		bsontype.Int32,
		bsontype.Int64,
		bsontype.Decimal128:
		return false
	default:
		return true
	}
}

// ToString is a shorthand for Convert where the target type is a string
// and no default behavior is specified in case of an error or null value.
func ToString(v bsoncore.Value) (bsoncore.Value, error) {
	return Convert(v, bsontype.String)
}

// Convert takes the input BSON value and converts it to a BSON value of the
// specified type.
func Convert(input bsoncore.Value, to bsontype.Type) (bsoncore.Value, error) {
	if input.Type == bsontype.Null {
		return Null(), nil
	}

	var converted bsoncore.Value
	var ok bool
	switch to {
	case bsontype.String:
		converted, ok = toString(input)
	}

	if !ok {
		return bsoncore.Value{}, fmt.Errorf("unsupported conversion from %s to %s with no onError value", TypeToString(input.Type), TypeToString(to))
	}

	return converted, nil
}

func toString(v bsoncore.Value) (bsoncore.Value, bool) {
	switch v.Type {
	case bsontype.Boolean:
		return String(strconv.FormatBool(v.Boolean())), true
	case bsontype.Double:
		return String(strconv.FormatFloat(v.Double(), 'f', -1, 64)), true
	case bsontype.Decimal128:
		return String(v.Decimal128().String()), true
	case bsontype.Int32:
		return String(strconv.FormatInt(int64(v.Int32()), 10)), true
	case bsontype.Int64:
		return String(strconv.FormatInt(v.Int64(), 10)), true
	case bsontype.ObjectID:
		return String(v.ObjectID().Hex()), true
	case bsontype.String:
		return v, true
	case bsontype.DateTime:
		return String(ISODateString(v)), true
	default:
		return bsoncore.Value{}, false
	}
}
