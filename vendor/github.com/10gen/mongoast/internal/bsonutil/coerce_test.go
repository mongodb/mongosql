package bsonutil_test

import (
	"fmt"
	"testing"

	"github.com/10gen/mongoast/internal/bsonutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestCoerceToBoolean(t *testing.T) {
	testCases := []struct {
		value    bsoncore.Value
		expected bool
	}{
		{
			bsonutil.Boolean(false),
			false,
		},
		{
			bsonutil.Boolean(true),
			true,
		},
		{
			bsonutil.Double(0),
			false,
		},
		{
			bsonutil.Double(1),
			true,
		},
		{
			bsonutil.String(""),
			true,
		},
		{
			bsonutil.String("foo"),
			true,
		},
		{
			bsonutil.EmptyDocument(),
			true,
		},
		{
			bsonutil.DocumentFromElements(
				"x", bsonutil.Null(),
			),
			true,
		},
		{
			bsonutil.EmptyArray(),
			true,
		},
		{
			bsonutil.ArrayFromValues(
				bsonutil.Null(),
			),
			true,
		},
		{
			bsonutil.Undefined(),
			false,
		},
		{
			bsonutil.ObjectID(primitive.NilObjectID),
			true,
		},
		{
			bsonutil.DateTime(0),
			true,
		},
		{
			bsonutil.Null(),
			false,
		},
		{
			bsonutil.Int32(0),
			false,
		},
		{
			bsonutil.Int32(1),
			true,
		},
		{
			bsonutil.Int64(0),
			false,
		},
		{
			bsonutil.Int64(1),
			true,
		},
		{
			bsonutil.Decimal128(primitive.NewDecimal128(0, 0)),
			false,
		},
		{
			bsonutil.Decimal128(primitive.NewDecimal128(0, 1)),
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s -> %v", tc.value.Type, tc.expected), func(t *testing.T) {
			actual := bsonutil.CoerceToBoolean(tc.value)
			if tc.expected != actual {
				t.Fatalf("expected %v, but got %v", tc.expected, actual)
			}
		})
	}
}
