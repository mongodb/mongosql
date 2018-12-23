package bsonutil_test

import (
	"testing"

	. "github.com/10gen/sqlproxy/internal/bsonutil"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/stretchr/testify/require"
)

func TestDeepCopyDSlice(t *testing.T) {
	type test struct {
		bson []bson.D
	}

	req := require.New(t)
	runTests := func(tests []test) {
		for _, test := range tests {
			copy := DeepCopyDSlice(test.bson)
			req.Equal(copy, test.bson)
			req.False(&copy == &test.bson)
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
