package bsonutil_test

import (
	"testing"

	. "github.com/10gen/sqlproxy/internal/util/bsonutil"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/stretchr/testify/require"
)

func TestDeepCopyPipeline(t *testing.T) {
	type test struct {
		bson []bson.D
	}

	req := require.New(t)
	runTests := func(tests []test) {
		for _, test := range tests {
			copy := DeepCopyPipeline(test.bson)
			req.Equal(copy, test.bson)
			req.False(&copy == &test.bson)
		}
	}

	tests := []test{
		{
			[]bson.D{
				{{Name: "$match", Value: bson.M{
					"a": int64(10),
				}}},
			},
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
					"path":                       "$__joined_b",
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
										bson.D{{Name: "$eq",
											Value: []interface{}{"$$predicate",
												false}}},
										bson.D{{Name: "$eq",
											Value: []interface{}{"$$predicate",
												0}}},
										bson.D{{Name: "$eq",
											Value: []interface{}{"$$predicate",
												"0"}}},
										bson.D{{Name: "$eq",
											Value: []interface{}{"$$predicate",
												"-0"}}},
										bson.D{{Name: "$eq",
											Value: []interface{}{"$$predicate",
												"0.0"}}},
										bson.D{{Name: "$eq",
											Value: []interface{}{"$$predicate",
												"-0.0"}}},
										bson.D{{Name: "$eq",
											Value: []interface{}{"$$predicate",
												nil}}},
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
		},
	}

	runTests(tests)
}
