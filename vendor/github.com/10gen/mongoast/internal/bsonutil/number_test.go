package bsonutil_test

import (
	"math"
	"testing"

	"github.com/10gen/mongoast/internal/bsonutil"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestAsInt32(t *testing.T) {
	testCases := []struct {
		name     string
		value    bsoncore.Value
		expected int32
		success  bool
	}{
		{
			"int32",
			bsoncore.Value{
				Type: bsontype.Int32,
				Data: bsoncore.AppendInt32(nil, 1),
			},
			1,
			true,
		},
		{
			"int64",
			bsoncore.Value{
				Type: bsontype.Int64,
				Data: bsoncore.AppendInt64(nil, 1),
			},
			1,
			true,
		},
		{
			"int64 overflow",
			bsoncore.Value{
				Type: bsontype.Int64,
				Data: bsoncore.AppendInt64(nil, math.MaxInt64-10),
			},
			0,
			false,
		},
		{
			"double",
			bsoncore.Value{
				Type: bsontype.Double,
				Data: bsoncore.AppendDouble(nil, 29.0),
			},
			29,
			true,
		},
		{
			"double overflow",
			bsoncore.Value{
				Type: bsontype.Double,
				Data: bsoncore.AppendDouble(nil, float64(math.MaxInt64-10)),
			},
			0,
			false,
		},
		{
			"double not an integer",
			bsoncore.Value{
				Type: bsontype.Double,
				Data: bsoncore.AppendDouble(nil, 29.1),
			},
			0,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, ok := bsonutil.AsInt32OK(tc.value)

			if ok != tc.success {
				t.Fatalf("expected %v, but got %v", tc.success, ok)
			}

			if tc.expected != actual {
				t.Fatalf("expected %v, but got %v", tc.expected, actual)
			}
		})
	}
}
