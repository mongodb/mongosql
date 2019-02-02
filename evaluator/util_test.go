package evaluator_test

import (
	"fmt"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	. "github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/stretchr/testify/require"
)

func TestComputeDocNestingDepth(t *testing.T) {
	type test struct {
		bson  []bson.D
		depth uint32
	}

	runTests := func(tests []test) {
		for idx, test := range tests {
			name := fmt.Sprintf("%d", idx)
			t.Run(name, func(t *testing.T) {
				depth := ComputeDocNestingDepthWithMaxDepth(test.bson, MaxDepth)
				require.Equal(t, test.depth, depth)
			})
		}
	}

	tests := []test{
		{bsonutil.NewDArray(
			bsonutil.NewD(bsonutil.NewDocElem("$match", bsonutil.NewM(
				bsonutil.NewDocElem("a", int64(10)),
			)),
			),
		), 3},
		{bsonutil.NewDArray(
			bsonutil.NewD(bsonutil.NewDocElem("$match", bsonutil.NewM(bsonutil.NewDocElem("a", bsonutil.NewM(bsonutil.NewDocElem("$ne", nil)))))),
			bsonutil.NewD(bsonutil.NewDocElem("$lookup", bsonutil.NewM(
				bsonutil.NewDocElem("from", "foo"),
				bsonutil.NewDocElem("localField", "a"),
				bsonutil.NewDocElem("foreignField", "a"),
				bsonutil.NewDocElem("as", "__joined_b"),
			)),
			),
			bsonutil.NewD(bsonutil.NewDocElem("$unwind", bsonutil.NewM(
				bsonutil.NewDocElem("path", "$__joined_b"),
				bsonutil.NewDocElem("preserveNullAndEmptyArrays", false),
			)),
			),
			bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
				bsonutil.NewDocElem("__joined_b._id", 1),
				bsonutil.NewDocElem("__joined_b.a", 1),
				bsonutil.NewDocElem("__joined_b.b", 1),
				bsonutil.NewDocElem("__joined_b.c", 1),
				bsonutil.NewDocElem("__joined_b.d.e", 1),
				bsonutil.NewDocElem("__joined_b.d.f", 1),
				bsonutil.NewDocElem("__joined_b.filter", 1),
				bsonutil.NewDocElem("__joined_b.g", 1),
				bsonutil.NewDocElem("_id", 1),
				bsonutil.NewDocElem("a", 1),
				bsonutil.NewDocElem("b", 1),
				bsonutil.NewDocElem("__predicate", bsonutil.NewD(
					bsonutil.NewDocElem("$let", bsonutil.NewD(
						bsonutil.NewDocElem("vars", bsonutil.NewM(
							bsonutil.NewDocElem("predicate", bsonutil.NewM(
								bsonutil.NewDocElem("$let", bsonutil.NewM(
									bsonutil.NewDocElem("vars", bsonutil.NewM(
										bsonutil.NewDocElem("left", "$a"),
										bsonutil.NewDocElem("right", "$__joined_b.d.f"),
									)),
									bsonutil.NewDocElem("in", bsonutil.NewM(
										bsonutil.NewDocElem("$cond", bsonutil.NewArray(
											bsonutil.NewM(
												bsonutil.NewDocElem("$or", bsonutil.NewArray(
													bsonutil.NewM(
														bsonutil.NewDocElem("$eq", bsonutil.NewArray(
															bsonutil.NewM(
																bsonutil.NewDocElem("$ifNull", bsonutil.NewArray(
																	"$$left",
																	nil,
																)),
															),
															nil,
														)),
													),
													bsonutil.NewM(
														bsonutil.NewDocElem("$eq", bsonutil.NewArray(
															bsonutil.NewM(
																bsonutil.NewDocElem("$ifNull", bsonutil.NewArray(
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
											bsonutil.NewM(
												bsonutil.NewDocElem("$eq", bsonutil.NewArray(
													"$$left",
													"$$right",
												)),
											),
										)),
									)),
								)),
							)),
						)),
						bsonutil.NewDocElem("in", bsonutil.NewD(
							bsonutil.NewDocElem("$cond", bsonutil.NewArray(
								bsonutil.NewD(bsonutil.NewDocElem("$or", bsonutil.NewArray(
									bsonutil.NewD(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
										"$$predicate",
										false,
									)),
									),
									bsonutil.NewD(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
										"$$predicate",
										0,
									)),
									),
									bsonutil.NewD(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
										"$$predicate",
										"0",
									)),
									),
									bsonutil.NewD(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
										"$$predicate",
										"-0",
									)),
									),
									bsonutil.NewD(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
										"$$predicate",
										"0.0",
									)),
									),
									bsonutil.NewD(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
										"$$predicate",
										"-0.0",
									)),
									),
									bsonutil.NewD(bsonutil.NewDocElem("$eq", bsonutil.NewArray(
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
			bsonutil.NewD(bsonutil.NewDocElem("$match", bsonutil.NewM(
				bsonutil.NewDocElem("__predicate", true),
			)),
			),
			bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewM(
				bsonutil.NewDocElem("test_DOT_a_DOT_b", "$b"),
				bsonutil.NewDocElem("test_DOT_a_DOT__id", "$_id"),
				bsonutil.NewDocElem("test_DOT_b_DOT_e", "$__joined_b.d.e"),
				bsonutil.NewDocElem("test_DOT_b_DOT_g", "$__joined_b.g"),
				bsonutil.NewDocElem("test_DOT_b_DOT_f", "$__joined_b.d.f"),
				bsonutil.NewDocElem("test_DOT_b_DOT__id", "$__joined_b._id"),
				bsonutil.NewDocElem("test_DOT_a_DOT_a", "$a"),
				bsonutil.NewDocElem("test_DOT_b_DOT_a", "$__joined_b.a"),
				bsonutil.NewDocElem("test_DOT_b_DOT_b", "$__joined_b.b"),
				bsonutil.NewDocElem("test_DOT_b_DOT_c", "$__joined_b.c"),
			)),
			),
		), 16},
	}

	runTests(tests)
}
