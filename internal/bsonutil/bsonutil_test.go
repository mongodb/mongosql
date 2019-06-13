package bsonutil_test

import (
	"math"
	"testing"

	. "github.com/10gen/sqlproxy/internal/bsonutil"

	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDeepCopyDSlice(t *testing.T) {
	type test struct {
		bson []bson.D
	}

	req := require.New(t)
	runTests := func(tests []test) {
		for _, test := range tests {
			copied := DeepCopyDSlice(test.bson)
			req.Equal(copied, test.bson)
			req.False(&copied == &test.bson)
		}
	}

	tests := []test{
		{NewDArray(
			NewD(NewDocElem("$match", NewM(
				NewDocElem("a", int64(10)),
			)),
			),
		)},
		{NewDArray(
			NewD(NewDocElem("$match", NewM(NewDocElem("a", NewM(NewDocElem("$ne", nil)))))),
			NewD(NewDocElem("$lookup", NewM(
				NewDocElem("from", "foo"),
				NewDocElem("localField", "a"),
				NewDocElem("foreignField", "a"),
				NewDocElem("as", "__joined_b"),
			)),
			),
			NewD(NewDocElem("$unwind", NewM(
				NewDocElem("path", "$__joined_b"),
				NewDocElem("preserveNullAndEmptyArrays", false),
			)),
			),
			NewD(NewDocElem("$project", NewM(
				NewDocElem("__joined_b._id", 1),
				NewDocElem("__joined_b.a", 1),
				NewDocElem("__joined_b.b", 1),
				NewDocElem("__joined_b.c", 1),
				NewDocElem("__joined_b.d.e", 1),
				NewDocElem("__joined_b.d.f", 1),
				NewDocElem("__joined_b.filter", 1),
				NewDocElem("__joined_b.g", 1),
				NewDocElem("_id", 1),
				NewDocElem("a", 1),
				NewDocElem("b", 1),
				NewDocElem("__predicate", NewD(
					NewDocElem("$let", NewD(
						NewDocElem("vars", NewM(
							NewDocElem("predicate", NewM(
								NewDocElem("$let", NewM(
									NewDocElem("vars", NewM(
										NewDocElem("left", "$a"),
										NewDocElem("right", "$__joined_b.d.f"),
									)),
									NewDocElem("in", NewM(
										NewDocElem("$cond", NewArray(
											NewM(
												NewDocElem("$or", NewArray(
													NewM(
														NewDocElem("$eq", NewArray(
															NewM(
																NewDocElem("$ifNull", NewArray(
																	"$$left",
																	nil,
																)),
															),
															nil,
														)),
													),
													NewM(
														NewDocElem("$eq", NewArray(
															NewM(
																NewDocElem("$ifNull", NewArray(
																	"$$right",
																	nil,
																)),
															),
															nil,
														)),
													),
												)),
											),
											nil,
											NewM(
												NewDocElem("$eq", NewArray(
													"$$left",
													"$$right",
												)),
											),
										)),
									)),
								)),
							)),
						)),
						NewDocElem("in", NewD(
							NewDocElem("$cond", NewArray(
								NewD(NewDocElem("$or", NewArray(
									NewD(NewDocElem("$eq", NewArray(
										"$$predicate",
										false,
									)),
									),
									NewD(NewDocElem("$eq", NewArray(
										"$$predicate",
										0,
									)),
									),
									NewD(NewDocElem("$eq", NewArray(
										"$$predicate",
										"0",
									)),
									),
									NewD(NewDocElem("$eq", NewArray(
										"$$predicate",
										"-0",
									)),
									),
									NewD(NewDocElem("$eq", NewArray(
										"$$predicate",
										"0.0",
									)),
									),
									NewD(NewDocElem("$eq", NewArray(
										"$$predicate",
										"-0.0",
									)),
									),
									NewD(NewDocElem("$eq", NewArray(
										"$$predicate",
										nil,
									)),
									),
								)),
								),
								false,
								true,
							)),
						)),
					)),
				)),
			)),
			),
			NewD(NewDocElem("$match", NewM(
				NewDocElem("__predicate", true),
			)),
			),
			NewD(NewDocElem("$project", NewM(
				NewDocElem("test_DOT_a_DOT_b", "$b"),
				NewDocElem("test_DOT_a_DOT__id", "$_id"),
				NewDocElem("test_DOT_b_DOT_e", "$__joined_b.d.e"),
				NewDocElem("test_DOT_b_DOT_g", "$__joined_b.g"),
				NewDocElem("test_DOT_b_DOT_f", "$__joined_b.d.f"),
				NewDocElem("test_DOT_b_DOT__id", "$__joined_b._id"),
				NewDocElem("test_DOT_a_DOT_a", "$a"),
				NewDocElem("test_DOT_b_DOT_a", "$__joined_b.a"),
				NewDocElem("test_DOT_b_DOT_b", "$__joined_b.b"),
				NewDocElem("test_DOT_b_DOT_c", "$__joined_b.c"),
			)),
			),
		)},
	}

	runTests(tests)
}

func TestNormalizeBSON(t *testing.T) {
	type test struct {
		name                   string
		pipeline               []bson.D
		expectedWithExtJSON    []bson.D
		expectedWithoutExtJSON []bson.D
	}

	testObjectID, _ := primitive.ObjectIDFromHex("57e193d7a9cc81b4027498b5")
	testDecimal, _ := primitive.ParseDecimal128("1234")

	tests := []test{
		{
			"nothing to normalize",
			NewDArray(
				NewD(
					NewDocElem("a", 1),
					NewDocElem("b", NewD(
						NewDocElem("c", "d"),
					)),
				),
			),
			NewDArray(
				NewD(
					NewDocElem("a", 1),
					NewDocElem("b", NewD(
						NewDocElem("c", "d"),
					)),
				),
			),
			NewDArray(
				NewD(
					NewDocElem("a", 1),
					NewDocElem("b", NewD(
						NewDocElem("c", "d"),
					)),
				),
			),
		},
		{
			"non-nested arrays",
			NewDArray(
				NewD(
					NewDocElem("a", []interface{}{1, 2, 3}),
					NewDocElem("b", NewD(
						NewDocElem("c", []interface{}{"x", "y", "z"}),
					)),
				),
				NewD(
					NewDocElem("d", 1),
					NewDocElem("e", NewArray(10, 11, 12)),
				),
			),
			NewDArray(
				NewD(
					NewDocElem("a", bson.A{1, 2, 3}),
					NewDocElem("b", NewD(
						NewDocElem("c", bson.A{"x", "y", "z"}),
					)),
				),
				NewD(
					NewDocElem("d", 1),
					NewDocElem("e", NewArray(10, 11, 12)),
				),
			),
			NewDArray(
				NewD(
					NewDocElem("a", bson.A{1, 2, 3}),
					NewDocElem("b", NewD(
						NewDocElem("c", bson.A{"x", "y", "z"}),
					)),
				),
				NewD(
					NewDocElem("d", 1),
					NewDocElem("e", NewArray(10, 11, 12)),
				),
			),
		},
		{
			"nested arrays",
			NewDArray(
				NewD(
					NewDocElem("a", 1),
					NewDocElem("b", []interface{}{
						2,
						"c",
						[]interface{}{3, 4, "z"},
						NewD(NewDocElem("n", []interface{}{"a", "b", "c"}))},
					),
				),
				NewD(
					NewDocElem("c", NewD(
						NewDocElem("d", []interface{}{1, 2, 3}),
					)),
				),
				NewD(
					NewDocElem("e", NewArray(
						1,
						[]interface{}{
							[]interface{}{"i", "j", "k"},
							NewD(
								NewDocElem("f", []interface{}{}),
							),
						},
						NewArray(4, 5, 6),
					)),
				),
			),
			NewDArray(
				NewD(
					NewDocElem("a", 1),
					NewDocElem("b", bson.A{
						2,
						"c",
						bson.A{3, 4, "z"},
						NewD(NewDocElem("n", bson.A{"a", "b", "c"}))},
					),
				),
				NewD(
					NewDocElem("c", NewD(
						NewDocElem("d", bson.A{1, 2, 3}),
					)),
				),
				NewD(
					NewDocElem("e", NewArray(
						1,
						bson.A{
							bson.A{"i", "j", "k"},
							NewD(
								NewDocElem("f", bson.A{}),
							),
						},
						NewArray(4, 5, 6),
					)),
				),
			),
			NewDArray(
				NewD(
					NewDocElem("a", 1),
					NewDocElem("b", bson.A{
						2,
						"c",
						bson.A{3, 4, "z"},
						NewD(NewDocElem("n", bson.A{"a", "b", "c"}))},
					),
				),
				NewD(
					NewDocElem("c", NewD(
						NewDocElem("d", bson.A{1, 2, 3}),
					)),
				),
				NewD(
					NewDocElem("e", NewArray(
						1,
						bson.A{
							bson.A{"i", "j", "k"},
							NewD(
								NewDocElem("f", bson.A{}),
							),
						},
						NewArray(4, 5, 6),
					)),
				),
			),
		},
		{
			"primitive types to normalize",
			NewDArray(
				NewD(NewDocElem("_id", NewD(
					NewDocElem("$oid", "57e193d7a9cc81b4027498b5"),
				))),
				NewD(NewDocElem("Symbol", NewD(
					NewDocElem("$symbol", "symbol"),
				))),
				NewD(NewDocElem("String", "string")),
				NewD(NewDocElem("Int32", NewD(
					NewDocElem("$numberInt", "42"),
				))),
				NewD(NewDocElem("Int64", NewD(
					NewDocElem("$numberLong", "42"),
				))),
				NewD(NewDocElem("Double", NewD(
					NewDocElem("$numberDouble", "42.42"),
				))),
				NewD(NewDocElem("SpecialFloat", NewD(
					NewDocElem("$numberDouble", "Infinity"),
				))),
				NewD(NewDocElem("Decimal", NewD(
					NewDocElem("$numberDecimal", "1234"),
				))),
				NewD(NewDocElem("Binary", NewD(
					NewDocElem("$binary", NewD(
						NewDocElem("base64", "o0w498Or7cijeBSpkquNtg=="),
						NewDocElem("subType", "03"),
					)),
				))),
				NewD(NewDocElem("BinaryUserDefined", NewD(
					NewDocElem("$binary", NewD(
						NewDocElem("base64", "AQIDBAU="),
						NewDocElem("subType", "80"),
					)),
				))),
				NewD(NewDocElem("Code", NewD(
					NewDocElem("$code", "function() {}"),
				))),
				NewD(NewDocElem("CodeWithScope", NewD(
					NewDocElem("$code", "function() {}"),
					NewDocElem("$scope", NewD()),
				))),
				NewD(NewDocElem("Subdocument", NewD(
					NewDocElem("foo", "bar"),
				))),
				NewD(NewDocElem("Array", []interface{}{
					NewD(NewDocElem("$numberInt", "1")),
					NewD(NewDocElem("$numberInt", "2")),
					NewD(NewDocElem("$numberInt", "3")),
					NewD(NewDocElem("$numberInt", "4")),
					NewD(NewDocElem("$numberInt", "5")),
				})),
				NewD(NewDocElem("Timestamp", NewD(
					NewDocElem("$timestamp", NewD(
						NewDocElem("t", 42),
						NewDocElem("i", 1),
					)),
				))),
				NewD(NewDocElem("RegularExpression", NewD(
					NewDocElem("$regularExpression", NewD(
						NewDocElem("pattern", "foo*"),
						NewDocElem("options", "ix"),
					)),
				))),
				NewD(NewDocElem("DatetimeEpoch", NewD(
					NewDocElem("$date", NewD(
						NewDocElem("$numberLong", "0"),
					)),
				))),
				NewD(NewDocElem("DatetimePositive", NewD(
					NewDocElem("$date", NewD(
						NewDocElem("$numberLong", "9223372036854775807"),
					)),
				))),
				NewD(NewDocElem("DatetimeNegative", NewD(
					NewDocElem("$date", NewD(
						NewDocElem("$numberLong", "-9223372036854775808"),
					)),
				))),
				NewD(NewDocElem("DatetimeString", NewD(
					NewDocElem("$date", "1970-01-01T00:00:00Z"),
				))),
				NewD(NewDocElem("True", true)),
				NewD(NewDocElem("False", false)),
				NewD(NewDocElem("DBPointer", NewD(
					NewDocElem("$dbPointer", NewD(
						NewDocElem("$ref", "db.collection"),
						NewDocElem("$id", NewD(
							NewDocElem("$oid", "57e193d7a9cc81b4027498b5"),
						)),
					)),
				))),
				NewD(NewDocElem("Minkey", NewD(
					NewDocElem("$minKey", 1),
				))),
				NewD(NewDocElem("Maxkey", NewD(
					NewDocElem("$maxKey", 1),
				))),
				NewD(NewDocElem("Undefined", NewD(
					NewDocElem("$undefined", true),
				))),
			),
			NewDArray(
				NewD(NewDocElem("_id", testObjectID)),
				NewD(NewDocElem("Symbol", primitive.Symbol("symbol"))),
				NewD(NewDocElem("String", "string")),
				NewD(NewDocElem("Int32", int32(42))),
				NewD(NewDocElem("Int64", int64(42))),
				NewD(NewDocElem("Double", 42.42)),
				NewD(NewDocElem("SpecialFloat", math.Inf(1))),
				NewD(NewDocElem("Decimal", testDecimal)),
				NewD(NewDocElem("Binary", primitive.Binary{
					Subtype: 0x03,
					Data:    []byte{0xa3, 0x4c, 0x38, 0xf7, 0xc3, 0xab, 0xed, 0xc8, 0xa3, 0x78, 0x14, 0xa9, 0x92, 0xab, 0x8d, 0xb6},
				})),
				NewD(NewDocElem("BinaryUserDefined", primitive.Binary{
					Subtype: 0x80,
					Data:    []byte{0x1, 0x2, 0x3, 0x4, 0x5},
				})),
				NewD(NewDocElem("Code", primitive.JavaScript("function() {}"))),
				NewD(NewDocElem("CodeWithScope", primitive.CodeWithScope{
					Code:  primitive.JavaScript("function() {}"),
					Scope: NewD(),
				})),
				NewD(NewDocElem("Subdocument", NewD(
					NewDocElem("foo", "bar"),
				))),
				NewD(NewDocElem("Array", bson.A{
					int32(1),
					int32(2),
					int32(3),
					int32(4),
					int32(5),
				})),
				NewD(NewDocElem("Timestamp", primitive.Timestamp{
					T: 42,
					I: 1,
				})),
				NewD(NewDocElem("RegularExpression", primitive.Regex{
					Pattern: "foo*",
					Options: "ix",
				})),
				NewD(NewDocElem("DatetimeEpoch", primitive.DateTime(0))),
				NewD(NewDocElem("DatetimePositive", primitive.DateTime(9223372036854775807))),
				NewD(NewDocElem("DatetimeNegative", primitive.DateTime(-9223372036854775808))),
				NewD(NewDocElem("DatetimeString", primitive.DateTime(0))),
				NewD(NewDocElem("True", true)),
				NewD(NewDocElem("False", false)),
				NewD(NewDocElem("DBPointer", primitive.DBPointer{
					DB:      "db.collection",
					Pointer: testObjectID,
				})),
				NewD(NewDocElem("Minkey", primitive.MinKey{})),
				NewD(NewDocElem("Maxkey", primitive.MaxKey{})),
				NewD(NewDocElem("Undefined", primitive.Undefined{})),
			),
			NewDArray(
				NewD(NewDocElem("_id", NewD(
					NewDocElem("$oid", "57e193d7a9cc81b4027498b5"),
				))),
				NewD(NewDocElem("Symbol", NewD(
					NewDocElem("$symbol", "symbol"),
				))),
				NewD(NewDocElem("String", "string")),
				NewD(NewDocElem("Int32", NewD(
					NewDocElem("$numberInt", "42"),
				))),
				NewD(NewDocElem("Int64", NewD(
					NewDocElem("$numberLong", "42"),
				))),
				NewD(NewDocElem("Double", NewD(
					NewDocElem("$numberDouble", "42.42"),
				))),
				NewD(NewDocElem("SpecialFloat", NewD(
					NewDocElem("$numberDouble", "Infinity"),
				))),
				NewD(NewDocElem("Decimal", NewD(
					NewDocElem("$numberDecimal", "1234"),
				))),
				NewD(NewDocElem("Binary", NewD(
					NewDocElem("$binary", NewD(
						NewDocElem("base64", "o0w498Or7cijeBSpkquNtg=="),
						NewDocElem("subType", "03"),
					)),
				))),
				NewD(NewDocElem("BinaryUserDefined", NewD(
					NewDocElem("$binary", NewD(
						NewDocElem("base64", "AQIDBAU="),
						NewDocElem("subType", "80"),
					)),
				))),
				NewD(NewDocElem("Code", NewD(
					NewDocElem("$code", "function() {}"),
				))),
				NewD(NewDocElem("CodeWithScope", NewD(
					NewDocElem("$code", "function() {}"),
					NewDocElem("$scope", NewD()),
				))),
				NewD(NewDocElem("Subdocument", NewD(
					NewDocElem("foo", "bar"),
				))),
				NewD(NewDocElem("Array", NewArray(
					NewD(NewDocElem("$numberInt", "1")),
					NewD(NewDocElem("$numberInt", "2")),
					NewD(NewDocElem("$numberInt", "3")),
					NewD(NewDocElem("$numberInt", "4")),
					NewD(NewDocElem("$numberInt", "5")),
				))),
				NewD(NewDocElem("Timestamp", NewD(
					NewDocElem("$timestamp", NewD(
						NewDocElem("t", 42),
						NewDocElem("i", 1),
					)),
				))),
				NewD(NewDocElem("RegularExpression", NewD(
					NewDocElem("$regularExpression", NewD(
						NewDocElem("pattern", "foo*"),
						NewDocElem("options", "ix"),
					)),
				))),
				NewD(NewDocElem("DatetimeEpoch", NewD(
					NewDocElem("$date", NewD(
						NewDocElem("$numberLong", "0"),
					)),
				))),
				NewD(NewDocElem("DatetimePositive", NewD(
					NewDocElem("$date", NewD(
						NewDocElem("$numberLong", "9223372036854775807"),
					)),
				))),
				NewD(NewDocElem("DatetimeNegative", NewD(
					NewDocElem("$date", NewD(
						NewDocElem("$numberLong", "-9223372036854775808"),
					)),
				))),
				NewD(NewDocElem("DatetimeString", NewD(
					NewDocElem("$date", "1970-01-01T00:00:00Z"),
				))),
				NewD(NewDocElem("True", true)),
				NewD(NewDocElem("False", false)),
				NewD(NewDocElem("DBPointer", NewD(
					NewDocElem("$dbPointer", NewD(
						NewDocElem("$ref", "db.collection"),
						NewDocElem("$id", NewD(
							NewDocElem("$oid", "57e193d7a9cc81b4027498b5"),
						)),
					)),
				))),
				NewD(NewDocElem("Minkey", NewD(
					NewDocElem("$minKey", 1),
				))),
				NewD(NewDocElem("Maxkey", NewD(
					NewDocElem("$maxKey", 1),
				))),
				NewD(NewDocElem("Undefined", NewD(
					NewDocElem("$undefined", true),
				))),
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			// test with convertExtJSON=true
			actual, err := NormalizeBSON(test.pipeline, true)
			req.Nil(err)
			req.Equal(test.expectedWithExtJSON, actual)

			// test with convertExtJSON=false (i.e. arrays only)
			actual, err = NormalizeBSON(test.pipeline, false)
			req.Nil(err)
			req.Equal(test.expectedWithoutExtJSON, actual)
		})
	}
}
