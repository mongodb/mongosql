package bsonutil

import (
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// Concat concatenates two strings.
func Concat(a, b bsoncore.Value) (bsoncore.Value, error) {
	switch a.Type {
	case bsontype.Null, bsontype.Undefined:
		return Null(), nil
	case bsontype.String:
		switch b.Type {
		case bsontype.Null, bsontype.Undefined:
			return Null(), nil
		case bsontype.String:
			return String(a.StringValue() + b.StringValue()), nil
		default:
			return Null(), errors.Errorf(
				"$concat only supports strings, not %s",
				TypeToString(b.Type),
			)
		}
	default:
		return Null(), errors.Errorf(
			"$concat only supports strings, not %s",
			TypeToString(a.Type),
		)
	}
}
