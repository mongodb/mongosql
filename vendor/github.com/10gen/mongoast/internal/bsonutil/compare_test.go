package bsonutil_test

import (
	"fmt"
	"testing"

	"github.com/10gen/mongoast/internal/bsonutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestCompare(t *testing.T) {
	testCases := []struct {
		a        bsoncore.Value
		b        bsoncore.Value
		expected int
	}{
		{
			bsonutil.ArrayFromValues(bsonutil.Int32(1)),
			bsonutil.ArrayFromValues(bsonutil.Int32(0)),
			1,
		},
		{
			bsonutil.ArrayFromValues(
				bsonutil.Int32(0),
				bsonutil.Int32(1),
			),
			bsonutil.ArrayFromValues(
				bsonutil.Int32(2),
				bsonutil.Int32(1),
				bsonutil.Int32(0),
			),
			-1,
		},
		{
			bsonutil.ArrayFromValues(
				bsonutil.Int32(0),
				bsonutil.Int32(1),
			),
			bsonutil.ArrayFromValues(
				bsonutil.Int32(0),
				bsonutil.Int32(1),
				bsonutil.Int32(2),
			),
			-1,
		},
		{
			bsonutil.Binary(0, []byte{0xFF, 0x00}),
			bsonutil.Binary(1, []byte{0x00, 0xFF}),
			1,
		},
		{
			bsonutil.Binary(1, []byte{0x00, 0xFF}),
			bsonutil.Binary(0, []byte{0x00, 0xFF}),
			1,
		},
		{
			bsonutil.Binary(1, []byte{0x00}),
			bsonutil.Binary(0, []byte{0x00, 0x00}),
			-1,
		},
		{
			bsonutil.Boolean(false),
			bsonutil.Boolean(true),
			-1,
		},
		{
			bsonutil.CodeWithScope(
				"function a() {}",
				bsonutil.MakeDoc("x", 1),
			),
			bsonutil.CodeWithScope(
				"function b() {}",
				bsonutil.MakeDoc("x", 0),
			),
			-1,
		},
		{
			bsonutil.CodeWithScope(
				"function() {}",
				bsonutil.MakeDoc("y", 1),
			),
			bsonutil.CodeWithScope(
				"function() {}",
				bsonutil.MakeDoc("x", 1),
			),
			1,
		},
		{
			bsonutil.DateTime(0),
			bsonutil.DateTime(0),
			0,
		},
		{
			bsonutil.DateTime(0),
			bsonutil.DateTime(1),
			-1,
		},
		{
			bsonutil.DateTime(1),
			bsonutil.DateTime(0),
			1,
		},
		{
			bsonutil.DBPointer("b", primitive.NilObjectID),
			bsonutil.DBPointer("ab", primitive.NilObjectID),
			-1,
		},
		{
			bsonutil.DBPointer("xyz", primitive.NilObjectID),
			bsonutil.DBPointer("abc", primitive.NilObjectID),
			1,
		},
		{
			bsonutil.DBPointer("x", bsonutil.ObjectIDFromHex("012345678901234567890123").ObjectID()),
			bsonutil.DBPointer("y", bsonutil.ObjectIDFromHex("123456789012345678901234").ObjectID()),
			-1,
		},
		{
			bsonutil.Decimal128(parseDecimal128("1")),
			bsonutil.Decimal128(parseDecimal128("2")),
			-1,
		},
		{
			bsonutil.Decimal128(parseDecimal128("1")),
			bsonutil.Double(2),
			-1,
		},
		{
			bsonutil.Decimal128(parseDecimal128("1")),
			bsonutil.Int32(2),
			-1,
		},
		{
			bsonutil.Decimal128(parseDecimal128("1")),
			bsonutil.Int64(2),
			-1,
		},
		{
			bsonutil.DocumentFromElements("a", bsonutil.Int32(1)),
			bsonutil.DocumentFromElements("a", bsonutil.Int32(1)),
			0,
		},
		{
			bsonutil.DocumentFromElements("a", bsonutil.Int32(1)),
			bsonutil.DocumentFromElements("a", bsonutil.Int32(2)),
			-1,
		},
		{
			bsonutil.DocumentFromElements("b", bsonutil.Int32(1)),
			bsonutil.DocumentFromElements("a", bsonutil.Int32(2)),
			1,
		},
		{
			bsonutil.DocumentFromElements(
				"b", bsonutil.Int32(1),
				"a", bsonutil.Int32(2),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(2),
				"b", bsonutil.Int32(1),
			),
			1,
		},
		{
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
				"b", bsonutil.Int32(2),
				"c", bsonutil.Int32(3),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
				"b", bsonutil.Int32(2),
			),
			1,
		},
		{
			bsonutil.Double(2.5),
			bsonutil.Double(1.5),
			1,
		},
		{
			bsonutil.Double(1),
			bsonutil.Decimal128(parseDecimal128("2")),
			-1,
		},
		{
			bsonutil.Int32(10),
			bsonutil.Int32(10),
			0,
		},
		{
			bsonutil.Int32(1),
			bsonutil.Decimal128(parseDecimal128("2")),
			-1,
		},
		{
			bsonutil.Int32(10),
			bsonutil.Int64(11),
			-1,
		},
		{
			bsonutil.Int64(1),
			bsonutil.Decimal128(parseDecimal128("2")),
			-1,
		},
		{
			bsonutil.JavaScript("function a() {}"),
			bsonutil.JavaScript("function b() {}"),
			-1,
		},
		{
			bsonutil.MaxKey(),
			bsonutil.MaxKey(),
			0,
		},
		{
			bsonutil.MinKey(),
			bsonutil.MinKey(),
			0,
		},
		{
			bsonutil.Null(),
			bsonutil.Null(),
			0,
		},
		{
			bsonutil.ObjectIDFromHex("123456789012345678901234"),
			bsonutil.ObjectIDFromHex("012345678901234567890123"),
			1,
		},
		{
			bsonutil.Regex("abc", "y"),
			bsonutil.Regex("def", "x"),
			-1,
		},
		{
			bsonutil.Regex("abc", "y"),
			bsonutil.Regex("abc", "x"),
			1,
		},
		{
			bsonutil.Regex("abc", ""),
			bsonutil.Regex("abc", "x"),
			-1,
		},
		{
			bsonutil.String("a"),
			bsonutil.String("b"),
			-1,
		},
		{
			bsonutil.Symbol("y"),
			bsonutil.Symbol("x"),
			1,
		},
		{
			bsonutil.Timestamp(1, 2),
			bsonutil.Timestamp(2, 1),
			-1,
		},
		{
			bsonutil.Timestamp(1, 2),
			bsonutil.Timestamp(1, 1),
			1,
		},
		{
			bsonutil.Undefined(),
			bsonutil.Undefined(),
			0,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s with %s -> %v", tc.a.Type, tc.b.Type, tc.expected), func(t *testing.T) {
			actual, err := bsonutil.Compare(tc.a, tc.b)

			if err != nil {
				t.Fatalf("expected no error, but got %v", err)
			}

			if tc.expected != actual {
				t.Fatalf("expected %v, but got %v", tc.expected, actual)
			}
		})
	}
}

func parseDecimal128(s string) primitive.Decimal128 {
	d128, err := primitive.ParseDecimal128(s)
	if err != nil {
		panic(err)
	}
	return d128
}
