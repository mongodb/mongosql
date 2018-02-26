package evaluator_test

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/evaluator"
)

func TestComputeDocNestingDepth(t *testing.T) {
	type test struct {
		bson  []bson.D
		depth uint32
	}

	runTests := func(tests []test) {
		for _, test := range tests {
			Convey(fmt.Sprintf("%q should have depth %d", test.bson, test.depth), t, func() {
				depth := evaluator.ComputeDocNestingDepthWithMaxDepth(test.bson, evaluator.MaxDepth)
				So(depth, ShouldEqual, test.depth)
			})
		}
	}

	tests := []test{
		{
			[]bson.D{
				{{Name: "$match", Value: bson.M{
					"a": int64(10),
				}}},
			},
			3,
		},
		{
			[]bson.D{
				{{Name: "$match", Value: bson.M{"a": bson.M{"$ne": nil}}}},
				{{Name: "$lookup", Value: bson.M{
					"from":         "foo",
					"localField":   "a",
					"foreignField": "a",
					"as":           "__joined_b",
				}}},
				{{Name: "$unwind", Value: bson.M{
					"path": "$__joined_b",
					"preserveNullAndEmptyArrays": false,
				}}},
				{{Name: "$project", Value: bson.M{
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
						{Name: "$let", Value: bson.D{
							{Name: "vars", Value: bson.M{
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
							{Name: "in", Value: bson.D{
								{Name: "$cond", Value: []interface{}{
									bson.D{{Name: "$or", Value: []interface{}{
										bson.D{{Name: "$eq", Value: []interface{}{"$$predicate", false}}},
										bson.D{{Name: "$eq", Value: []interface{}{"$$predicate", 0}}},
										bson.D{{Name: "$eq", Value: []interface{}{"$$predicate", "0"}}},
										bson.D{{Name: "$eq", Value: []interface{}{"$$predicate", "-0"}}},
										bson.D{{Name: "$eq", Value: []interface{}{"$$predicate", "0.0"}}},
										bson.D{{Name: "$eq", Value: []interface{}{"$$predicate", "-0.0"}}},
										bson.D{{Name: "$eq", Value: []interface{}{"$$predicate", nil}}},
									}}},
									false,
									true,
								}},
							}},
						}},
					},
				}}},
				{{Name: "$match", Value: bson.M{
					"__predicate": true,
				}}},
				{{Name: "$project", Value: bson.M{
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

func TestCleanNumericString(t *testing.T) {
	type test struct {
		input, output string
	}

	runTests := func(tests []test) {
		for _, test := range tests {
			Convey(fmt.Sprintf("%q should clean to %q", test.input, test.output), t, func() {
				output := evaluator.CleanNumericString(test.input)
				So(output, ShouldEqual, test.output)
			})
		}
	}
	tests := []test{
		{"     -12345.1234.34xwwyzz   :", "-12345.1234"},
		{"    - 12345.1234.34xwwyzz   :", "0"},
		{"1234", "1234"},
		{"  1234  ", "1234"},
		{"   -3.14159265xyz", "-3.14159265"},
		{" Hello World  ", "0"},
		{"1.2.3.4", "1.2"},
	}

	runTests(tests)
}
