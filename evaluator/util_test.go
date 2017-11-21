package evaluator

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/10gen/mongo-go-driver/bson"
)

func TestUtils(t *testing.T) {
	type test struct {
		bson  []bson.D
		depth uint32
	}

	runTests := func(tests []test) {
		for _, test := range tests {
			Convey(fmt.Sprintf("%q should have depth %d", test.bson, test.depth), t, func() {
				depth := computeDocNestingDepthWithMaxDepth(test.bson, maxDepth)
				So(depth, ShouldEqual, test.depth)
			})
		}
	}

	tests := []test{
		test{
			[]bson.D{
				{{"$match", bson.M{
					"a": int64(10),
				}}},
			},
			3,
		},
		test{
			[]bson.D{
				{{"$match", bson.M{"a": bson.M{"$ne": nil}}}},
				{{"$lookup", bson.M{
					"from":         "foo",
					"localField":   "a",
					"foreignField": "a",
					"as":           "__joined_b",
				}}},
				{{"$unwind", bson.M{
					"path": "$__joined_b",
					"preserveNullAndEmptyArrays": false,
				}}},
				{{"$project", bson.M{
					"__joined_b._id":    1,
					"__joined_b.a":      1,
					"__joined_b.b":      1,
					"__joined_b.c":      1,
					"__joined_b.d.e":    1,
					"__joined_b.d.f":    1,
					"__joined_b.filter": 1,
					"__joined_b.g":      1,
					"_id":               1,
					"a":                 1,
					"b":                 1,
					"__predicate": bson.D{
						{"$let", bson.D{
							{"vars", bson.M{
								"predicate": bson.M{
									"$let": bson.M{
										"vars": bson.M{
											"left":  "$a",
											"right": "$__joined_b.d.f",
										},
										"in": bson.M{
											"$cond": []interface{}{
												bson.M{
													"$or": []interface{}{
														bson.M{
															"$eq": []interface{}{
																bson.M{
																	"$ifNull": []interface{}{
																		"$$left",
																		nil,
																	},
																},
																nil,
															},
														},
														bson.M{
															"$eq": []interface{}{
																bson.M{
																	"$ifNull": []interface{}{
																		"$$right",
																		nil,
																	},
																},
																nil,
															},
														},
													},
												},
												nil,
												bson.M{
													"$eq": []interface{}{
														"$$left",
														"$$right",
													},
												},
											},
										},
									},
								},
							}},
							{"in", bson.D{
								{"$cond", []interface{}{
									bson.D{{"$or", []interface{}{
										bson.D{{"$eq", []interface{}{"$$predicate", false}}},
										bson.D{{"$eq", []interface{}{"$$predicate", 0}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "-0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "0.0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", "-0.0"}}},
										bson.D{{"$eq", []interface{}{"$$predicate", nil}}},
									}}},
									false,
									true,
								}},
							}},
						}},
					},
				}}},
				{{"$match", bson.M{
					"__predicate": true,
				}}},
				{{"$project", bson.M{
					"test_DOT_a_DOT_b":   "$b",
					"test_DOT_a_DOT__id": "$_id",
					"test_DOT_b_DOT_e":   "$__joined_b.d.e",
					"test_DOT_b_DOT_g":   "$__joined_b.g",
					"test_DOT_b_DOT_f":   "$__joined_b.d.f",
					"test_DOT_b_DOT__id": "$__joined_b._id",
					"test_DOT_a_DOT_a":   "$a",
					"test_DOT_b_DOT_a":   "$__joined_b.a",
					"test_DOT_b_DOT_b":   "$__joined_b.b",
					"test_DOT_b_DOT_c":   "$__joined_b.c",
				}}},
			},
			16,
		},
	}

	runTests(tests)
}
