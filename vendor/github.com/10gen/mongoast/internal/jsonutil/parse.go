package jsonutil

import (
	"strings"

	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// ParseJSON parses a JSON input string into a bsoncore.Value.
func ParseJSON(input string) bsoncore.Value {
	vr, err := bsonrw.NewExtJSONValueReader(strings.NewReader(input), false)
	if err != nil {
		panic(err)
	}
	c := bsonrw.NewCopier()

	t, bytes, err := c.CopyValueToBytes(vr)

	if err != nil {
		panic(err)
	}

	return bsoncore.Value{
		Type: t,
		Data: bytes,
	}
}
