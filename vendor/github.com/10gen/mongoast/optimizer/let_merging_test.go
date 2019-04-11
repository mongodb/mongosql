package optimizer_test

import (
	"context"
	"testing"

	"github.com/10gen/mongoast/internal/parsertest"
	"github.com/10gen/mongoast/optimizer"
	"github.com/10gen/mongoast/parser"

	"github.com/google/go-cmp/cmp"
)

func TestLetMerging(t *testing.T) {
	// Note: this is just testing LetMerging, so the let variables
	// in these test cases may contain values that would be removed
	// by LetMinimization. The type of expression in the bindings
	// does not matter; the important thing is the dependence on
	// outer Lets.
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"do not merge dependent lets",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$let": {
							"vars": {
								"v2": "$$v1"
							},
							"in": "$$v2"
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$let": {
							"vars": {
								"v2": "$$v1"
							},
							"in": "$$v2"
						}
					}
				}}}}
			 ]`,
		},
		{
			"do not merge independent shadowed variable",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$let": {
							"vars": {
								"v1": "$y",
								"v2": "$$v1"
							},
							"in": {"$add": ["$$v1", "$$v2"]}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$let": {
							"vars": {
								"v1": "$y",
								"v2": "$$v1"
							},
							"in": {"$add": ["$$v1", "$$v2"]}
						}
					}
				}}}}
			 ]`,
		},
		{
			"merge independent lets",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$let": {
							"vars": {
								"v2": "$y"
							},
							"in": {"$add": ["$$v1", "$$v2"]}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x",
						"v2": "$y"
					},
					"in": {"$add": ["$$v1", "$$v2"]}
				}}}}
			 ]`,
		},
		{
			"merge multiple nested lets that are not direct descendents of an outer let",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$add": [
							{"$let": {
								"vars": {
									"v2": "$y"
								},
								"in": {"$add":["$$v1", "$$v2"]}
							}},
							{"$let": {
								"vars": {
									"v3": "$z"
								},
								"in": {"$add":["$$v1", "$$v3"]}
							}}
						]
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x",
						"v2": "$y",
						"v3": "$z"
					},
					"in": {"$add": [{"$add":["$$v1", "$$v2"]}, {"$add":["$$v1", "$$v3"]}]}
				}}}}
			 ]`,
		},
		{
			"fully merge multiple levels of independent lets",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$let": {
							"vars": {
								"v2": "$y"
							},
							"in": {
								"$let": {
									"vars": {
										"v3": "$z",
										"v4": "$n"
									},
									"in": {
										"$let": {
											"vars": {
												"v5": "$m"
											},
											"in": {"$add": ["$$v1", "$$v2", "$$v3", "$$v4", "$$v5"]}
										}
									}
								}
							}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x",
						"v2": "$y",
						"v3": "$z",
						"v4": "$n",
						"v5": "$m"
					},
					"in": {"$add": ["$$v1", "$$v2", "$$v3", "$$v4", "$$v5"]}
				}}}}
			 ]`,
		},
		{
			"partially merge multiple levels of semi-dependent lets",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$let": {
							"vars": {
								"v2": "$y"
							},
							"in": {
								"$let": {
									"vars": {
										"v3": "$$v1",
										"v4": "$n"
									},
									"in": {
										"$let": {
											"vars": {
												"v5": "$$v2",
												"v6": "$$v3"
											},
											"in": {"$add": ["$$v1", "$$v2", "$$v3", "$$v4", "$$v5", "$$v6"]}
										}
									}
								}
							}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x",
						"v2": "$y",
						"v4": "$n"
					},
					"in": {
						"$let": {
							"vars": {
								"v3": "$$v1",
								"v5": "$$v2"
							},
							"in": {
								"$let": {
									"vars": {
										"v6": "$$v3"
									},
									"in": {"$add": ["$$v1", "$$v2", "$$v3", "$$v4", "$$v5", "$$v6"]}
								}
							}
						}
					}
				}}}}
			 ]`,
		},
		{
			"do not merge let variables nested in $map that reference the as variable",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$map": {
							"input": [1, 2, 3],
							"as": "var",
							"in": {
								"$let": {
									"vars": {
										"v2": {"$add": ["$$var", 2]},
										"v3": "$y"
									},
									"in": {"$add": ["$$v1", "$$v2", "$$v3"]}
								}
							}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x",
						"v3": "$y"
					},
					"in": {
						"$map": {
							"input": [1, 2, 3],
							"as": "var",
							"in": {
								"$let": {
									"vars": {
										"v2": {"$add": ["$$var", 2]}
									},
									"in": {"$add": ["$$v1", "$$v2", "$$v3"]}
								}
							}
						}
					}
				}}}}
			 ]`,
		},
		{
			"do not merge let variables nested in $filter that reference the as variable",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$filter": {
							"input": [1, 2, 3],
							"as": "var",
							"cond": {
								"$let": {
									"vars": {
										"v2": {"$add": ["$$var", 2]},
										"v3": "$y"
									},
									"in": {"$eq": [{"$add": ["$$v1", "$$v2"]}, "$$v3"]}
								}
							}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x",
						"v3": "$y"
					},
					"in": {
						"$filter": {
							"input": [1, 2, 3],
							"as": "var",
							"cond": {
								"$let": {
									"vars": {
										"v2": {"$add": ["$$var", 2]}
									},
									"in": {"$eq": [{"$add": ["$$v1", "$$v2"]}, "$$v3"]}
								}
							}
						}
					}
				}}}}
			 ]`,
		},
		{
			"do not merge let variables nested in $reduce that reference $$this or $$value",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$reduce": {
							"input": [1, 2, 3],
							"initialValue": 0,
							"in": {
								"$let": {
									"vars": {
										"v2": {"$add": ["$$this", 2]},
										"v3": "$y",
										"v4": {"$add": ["$$value", 2]}
									},
									"in": {"$add": ["$$v1", "$$v2", "$$v3", "$$v4"]}
								}
							}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x",
						"v3": "$y"
					},
					"in": {
						"$reduce": {
							"input": [1, 2, 3],
							"initialValue": 0,
							"in": {
								"$let": {
									"vars": {
										"v2": {"$add": ["$$this", 2]},
										"v4": {"$add": ["$$value", 2]}
									},
									"in": {"$add": ["$$v1", "$$v2", "$$v3", "$$v4"]}
								}
							}
						}
					}
				}}}}
			 ]`,
		},
		{
			"do not merge let bindings named using the as field from a sibling $map where the name shadows an outer variable",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$add": [
							{"$map": {
								"input": [1, 2, 3],
								"as": "v1",
								"in": {"$add": ["$$v1", "$$v1"]}
							}},
							{"$let": {
								"vars": {
									"v1": "$y"
								},
								"in": {"$add": ["$$v1", "$$v1"]}
							}}
						]
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$add": [
							{"$map": {
								"input": [1, 2, 3],
								"as": "v1",
								"in": {"$add": ["$$v1", "$$v1"]}
							}},
							{"$let": {
								"vars": {
									"v1": "$y"
								},
								"in": {"$add": ["$$v1", "$$v1"]}
							}}
						]
					}
				}}}}
			 ]`,
		},
		{
			"do not merge let bindings named using the as field from a sibling $filter where the name shadows an outer variable",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$concatArrays": [
							{"$filter": {
								"input": [1, 2, 3],
								"as": "v1",
								"cond": {"$eq": ["$$v1", "$$v1"]}
							}},
							{"$let": {
								"vars": {
									"v1": "$y"
								},
								"in": ["$$v1", "$$v1"]
							}}
						]
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$concatArrays": [
							{"$filter": {
								"input": [1, 2, 3],
								"as": "v1",
								"cond": {"$eq": ["$$v1", "$$v1"]}
							}},
							{"$let": {
								"vars": {
									"v1": "$y"
								},
								"in": ["$$v1", "$$v1"]
							}}
						]
					}
				}}}}
			 ]`,
		},
		{
			"do not merge let bindings named this and value from let expression that is a sibling of a $reduce where the names shadow outer variables",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"this": "$x",
						"value": "$y"
					},
					"in": {
						"$add": [
							{"$reduce": {
								"input": [1, 2, 3],
								"initialValue": 0,
								"in": {"$add": ["$$this", "$$value"]}
							}},
							{"$let": {
								"vars": {
									"this": "$n",
									"value": "$m"
								},
								"in": {"$add": ["$$this", "$$value"]}
							}}
						]
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"this": "$x",
						"value": "$y"
					},
					"in": {
						"$add": [
							{"$reduce": {
								"input": [1, 2, 3],
								"initialValue": 0,
								"in": {"$add": ["$$this", "$$value"]}
							}},
							{"$let": {
								"vars": {
									"this": "$n",
									"value": "$m"
								},
								"in": {"$add": ["$$this", "$$value"]}
							}}
						]
					}
				}}}}
			 ]`,
		},
		{
			"merge let bindings named using the as field from a sibling $map",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$add": [
							{"$map": {
								"input": [1, 2, 3],
								"as": "v2",
								"in": {"$add": ["$$v1", "$$v2"]}
							}},
							{"$let": {
								"vars": {
									"v2": "$y"
								},
								"in": {"$add": ["$$v1", "$$v2"]}
							}}
						]
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x",
						"v2": "$y"
					},
					"in": {
						"$add": [
							{"$map": {
								"input": [1, 2, 3],
								"as": "v2",
								"in": {"$add": ["$$v1", "$$v2"]}
							}},
							{"$add": ["$$v1", "$$v2"]}
						]
					}
				}}}}
			 ]`,
		},
		{
			"merge let bindings named using the as field from a sibling $filter",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$concatArrays": [
							{"$filter": {
								"input": [1, 2, 3],
								"as": "v2",
								"cond": {"$eq": ["$$v1", "$$v2"]}
							}},
							{"$let": {
								"vars": {
									"v2": "$y"
								},
								"in": ["$$v1", "$$v2"]
							}}
						]
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x",
						"v2": "$y"
					},
					"in": {
						"$concatArrays": [
							{"$filter": {
								"input": [1, 2, 3],
								"as": "v2",
								"cond": {"$eq": ["$$v1", "$$v2"]}
							}},
							["$$v1", "$$v2"]
						]
					}
				}}}}
			 ]`,
		},
		{
			"merge let bindings named this and value from let expression that is a sibling of a $reduce",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$add": [
							{"$reduce": {
								"input": [1, 2, 3],
								"initialValue": 0,
								"in": {"$add": ["$$v1", "$$this", "$$value"]}
							}},
							{"$let": {
								"vars": {
									"this": "$y",
									"value": "$z"
								},
								"in": {"$add": ["$$v1", "$$this", "$$value"]}
							}}
						]
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x",
						"this": "$y",
						"value": "$z"
					},
					"in": {
						"$add": [
							{"$reduce": {
								"input": [1, 2, 3],
								"initialValue": 0,
								"in": {"$add": ["$$v1", "$$this", "$$value"]}
							}},
							{"$add": ["$$v1", "$$this", "$$value"]}
						]
					}
				}}}}
			 ]`,
		},
		{
			"do nothing if there are no let expressions",
			`[
				{"$project": {"a": {"$add": ["$x", "$y"]}}}
			 ]`,
			`[
				{"$project": {"a": {"$add": ["$x", "$y"]}}}
			 ]`,
		},
		{
			// Here, even though "v2" is independent of the outer $let,
			// since it is nested in a lazy expression it should not be
			// merged up. If it were merged up, then $strLenCP would be
			// evaluated unconditionally, and that behavior is different
			// from the input pipeline.
			"do not merge independent let bindings nested in lazy expressions",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$cond": [
							{"$lte": ["$y", null]},
							null,
							{"$let": {
								"vars": {
									"v2": {"$strLenCP": "$y"}
								},
								"in": {"$add": ["$$v1", "$$v2"]}
							}}
						]
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$cond": [
							{"$lte": ["$y", null]},
							null,
							{"$let": {
								"vars": {
									"v2": {"$strLenCP": "$y"}
								},
								"in": {"$add": ["$$v1", "$$v2"]}
							}}
						]
					}
				}}}}
			 ]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in := parsertest.ParsePipeline(tc.input)
			expected := parsertest.ParsePipeline(tc.expected)
			actual := optimizer.RunPasses(context.Background(), in, optimizer.LetMerging)

			expectedStr := parser.DeparsePipeline(expected).String()
			actualStr := parser.DeparsePipeline(actual).String()

			if !cmp.Equal(expectedStr, actualStr) {
				t.Fatalf("\nexpected:\n %s\ngot:\n %s", expectedStr, actualStr)
			}
		})
	}
}
