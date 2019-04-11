package optimizer_test

import (
	"context"
	"testing"

	"github.com/10gen/mongoast/internal/parsertest"
	"github.com/10gen/mongoast/optimizer"
	"github.com/10gen/mongoast/parser"

	"github.com/google/go-cmp/cmp"
)

func TestPartialRedundancyElimination(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"initial redundancy test",
			`[
				{"$project":
					{"a":
					    {"$sub":
						[
							{"$add": [
								{"$add": ["$c", "$d"]},
								{"$add": ["$c", "$d"]},
								{"$add": ["$c", "$d"]}
								]
							},
							{"$add": [
								{"$add": ["$c", "$d"]},
								{"$add": ["$c", "$d"]},
								{"$add": ["$c", "$d"]}
								]
							}
						]
						}
					}
				}
			 ]`,
			`[
			  {"$project":
			  	{"a":
					{"$let":
						{"vars": {"mongoast__deduplicated__expr__0": {"$add": ["$c","$d"]}},
						 "in":
						 	{"$let":
								{"vars": {"mongoast__deduplicated__expr__1":
										{"$add": ["$$mongoast__deduplicated__expr__0",
													"$$mongoast__deduplicated__expr__0",
													"$$mongoast__deduplicated__expr__0"]}},
									"in": {"$sub": ["$$mongoast__deduplicated__expr__1","$$mongoast__deduplicated__expr__1"]}
								}
							}
						}
					}
				}
			 }
			 ]`,
		}, {
			"in match",
			`[
				{"$match":
					{"$expr":
						{"$sub": [
							{"$add": ["$a", "$b"]},
							{"$add": ["$a", "$b"]}
							]
						}
					}
				}
			]`,
			`[
				{"$match":
					{"$expr":
						{"$let":
							{"vars":
								{"mongoast__deduplicated__expr__0": {"$add": ["$a","$b"]}},
							"in": {"$sub": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}
							}
						}
					}
				}
			]`,
		}, {
			"do not move lets above $maps",
			`[
				{"$project":
					{"out": {"$map": {
						"input": "$a.b",
						"as": "this",
						"in": {"$add": [
							{"$ifNull": ["$$this", null]},
							{"$ifNull": ["$$this", null]}
						]}
					}}
				}}
			]`,
			`[
				{"$project":
					{"out": {"$map": {
						"input": "$a.b",
						"as": {"$literal": "this"},
						"in": {"$let": {
								"vars": {"mongoast__deduplicated__expr__0": {"$ifNull": ["$$this",{"$literal": null}]}},
								"in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}
									}
								}
							}
						}
					}
				}
			]`,
		}, {
			"do not move lets above $reduce",
			`[
				{"$project":
					{"out": {"$reduce": {
						"input": "$a.b",
						"initialValue": {"$add": [
											{"$add": [1, 2]},
											{"$add": [1, 2]}
										]},
						"in": {"$add": [
							{"$ifNull": ["$$this", "$$value"]},
							{"$ifNull": ["$$this", "$$value"]}
						]}
					}}
				}}
			]`,
			`[
				{"$project":
					{"out": {"$reduce": {
						"input": "$a.b",
						"initialValue": {"$let": {
							"vars": {"mongoast__deduplicated__expr__0":
								{"$add": [{"$literal": {"$numberInt":"1"}},{"$literal": {"$numberInt":"2"}}]}},
							"in": {"$add": [
								"$$mongoast__deduplicated__expr__0",
								"$$mongoast__deduplicated__expr__0"]}}},
						"in": {"$let": {
							"vars": {"mongoast__deduplicated__expr__0":
								{"$ifNull": ["$$this","$$value"]}},
							"in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}
										}
								}
							}
						}
					}
				}
			]`,
		}, {
			"redundant reduce expressions should be let bound",
			`[
				{"$project":
					{"out": {"$add": [
								{"$reduce": {
									"input": [1,2,3],
									"initialValue": 0,
									"in": {"$add": ["$$this", "$$value"]}
								}},
								{"$reduce": {
									"input": [1,2,3],
									"initialValue": 0,
									"in": {"$add": ["$$this", "$$value"]}
								}}
							]
							}
					}
				}
			]`,
			`[
				{"$project":
					{"out": {"$let": {
						"vars": {"mongoast__deduplicated__expr__0":
							{"$reduce": {"input": [1,2,3],
							 "initialValue": 0,
							 "in": {"$add": ["$$this","$$value"]}}}
						 },
						"in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}}
						}
					}
				}
			]`,
		}, {
			"do not move lets above $reduce",
			`[
				{"$project":
					{"out": {"$filter": {
						"input": "$a.b",
						"as": "bar",
						"cond": {"$add": [
											{"$add": ["$$bar", 2]},
											{"$add": ["$$bar", 2]}
										]}
						}
					}}
				}
			]`,
			`[
				{"$project":
					{"out": {"$filter": {
						"input": "$a.b",
						"as": {"$literal": "bar"},
						"cond": {"$let":
							{"vars": {"mongoast__deduplicated__expr__0":
										{"$add": ["$$bar",{"$literal": {"$numberInt":"2"}}]
									}},
							"in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}
							}
						}}
					}}
				}
			]`,
		}, {
			"do not move lets above existing $lets",
			`[
				{"$project":
					{"out": {"$let": {
						"vars": {"foo": "$a"},
						"in": {"$add": [
											{"$add": ["$$foo", 2]},
											{"$add": ["$$foo", 2]}
										]}
						}
					}}
				}
			]`,
			`[
				{"$project":
					{"out": {"$let": {
						"vars": {"foo": "$a"},
						"in": {"$let":
							{"vars": {"mongoast__deduplicated__expr__0": {"$add": ["$$foo",{"$literal": {"$numberInt":"2"}}]}},
							"in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}
							}}
						}
					}}
				}
			]`,
		}, {
			"do not pull out redundancies at the top level of $and",
			`[
				{"$project":
					{"out": {"$and": [
								{"$add": [
									{"$add": [1,2,3]},
									{"$add": [1,2,3]}
									]
								},
								{"$add": [
									{"$add": [1,2,3]},
									{"$add": [1,2,3]}
									]
								}
								]
							}
					}
				}
			]`,
			`[
				{"$project":
					{"out": {"$and": [
						{"$let": {"vars": {"mongoast__deduplicated__expr__0":
									{"$add": [{"$literal": {"$numberInt":"1"}},{"$literal": {"$numberInt":"2"}},{"$literal": {"$numberInt":"3"}}]}},
						 		  "in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}}},
						{"$let": {"vars": {"mongoast__deduplicated__expr__0":
									{"$add": [{"$literal": {"$numberInt":"1"}},{"$literal": {"$numberInt":"2"}},{"$literal": {"$numberInt":"3"}}]}},
								  "in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}}}
								  ]
						}
					}
				}
			]`,
		}, {
			"do not pull out redundancies at the top level of $allElementsTrue",
			`[
				{"$project":
					{"out": {"$allElementsTrue": [
								{"$add": [
									{"$add": [1,2,3]},
									{"$add": [1,2,3]}
									]
								},
								{"$add": [
									{"$add": [1,2,3]},
									{"$add": [1,2,3]}
									]
								}
								]
							}
					}
				}
			]`,
			`[
				{"$project":
					{"out": {"$allElementsTrue": [
						{"$let": {"vars": {"mongoast__deduplicated__expr__0":
									{"$add": [{"$literal": {"$numberInt":"1"}},{"$literal": {"$numberInt":"2"}},{"$literal": {"$numberInt":"3"}}]}},
						 		  "in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}}},
						{"$let": {"vars": {"mongoast__deduplicated__expr__0":
									{"$add": [{"$literal": {"$numberInt":"1"}},{"$literal": {"$numberInt":"2"}},{"$literal": {"$numberInt":"3"}}]}},
								  "in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}}}
								  ]
						}
					}
				}
			]`,
		}, {
			"do not pull out redundancies at the top level of $cond",
			`[
				{"$project":
					{"out": {"$cond": [
								"$foo",
								{"$add": [
									{"$add": [1,2,3]},
									{"$add": [1,2,3]}
									]
								},
								{"$add": [
									{"$add": [1,2,3]},
									{"$add": [1,2,3]}
									]
								}
								]
							}
					}
				}
			]`,
			`[
				{"$project":
					{"out": {"$cond": [
						"$foo",
						{"$let": {"vars": {"mongoast__deduplicated__expr__0":
									{"$add": [{"$literal": {"$numberInt":"1"}},{"$literal": {"$numberInt":"2"}},{"$literal": {"$numberInt":"3"}}]}},
						 		  "in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}}},
						{"$let": {"vars": {"mongoast__deduplicated__expr__0":
									{"$add": [{"$literal": {"$numberInt":"1"}},{"$literal": {"$numberInt":"2"}},{"$literal": {"$numberInt":"3"}}]}},
								  "in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}}}
								  ]
						}
					}
				}
			]`,
		}, {
			"do not pull out redundancies above a $sum",
			`[
				{"$group":
					{
						"_id": 0,
						"out": {"$sum":
							{"$add": [
								{"$add": [1,2,3]},
								{"$add": [1,2,3]}
							]}
						}
					}
				}
			]`,
			`[
				{"$group":
					{
						"_id": 0,
						"out": {"$sum":
							{"$let": {"vars": {"mongoast__deduplicated__expr__0":
											{"$add": [1, 2, 3]}},
									  "in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]
								  }
							  }
						  }
					  }
				  }
			  }
			]`,
		},
		{
			"do not pull out redundancies above the branches field of a $switch or above the case and then fields of each branch",
			`[
				{"$project": {"out": {
					"$switch": {
						"branches": [
							{
								"case": {
									"$eq": [
										{"$add": [
											{"$add": [1,2,3]},
											{"$add": [1,2,3]}
										]},
										0
									]
								},
								"then": {
									"$add": [
										{"$add": [1,2,3]},
										{"$add": [1,2,3]}
									]
								}
							},
							{
								"case": {
									"$eq": [
										{"$add": [
											{"$add": [1,2,3]},
											{"$add": [1,2,3]}
										]},
										0
									]
								},
								"then": {
									"$add": [
										{"$add": [1,2,3]},
										{"$add": [1,2,3]}
									]
								}
							}
						],
						"default": {
							"$add": [
								{"$add": [1,2,3]},
								{"$add": [1,2,3]}
							]
						}
					}
				}}}
			]`,
			`[
				{"$project": {"out": {
					"$switch": {
						"branches": [
							{
								"case": {
									"$let": {
										"vars": {
											"mongoast__deduplicated__expr__0": {"$add": [1,2,3]}
										},
										"in": {
											"$eq": [
												{"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]},
												0
											]
										}
									}
								},
								"then": {
									"$let": {
										"vars": {
											"mongoast__deduplicated__expr__0": {"$add": [1,2,3]}
										},
										"in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}
									}
								}
							},
							{
								"case": {
									"$let": {
										"vars": {
											"mongoast__deduplicated__expr__0": {"$add": [1,2,3]}
										},
										"in": {
											"$eq": [
												{"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]},
												0
											]
										}
									}
								},
								"then": {
									"$let": {
										"vars": {
											"mongoast__deduplicated__expr__0": {"$add": [1,2,3]}
										},
										"in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}
									}
								}
							}
						],
						"default": {
							"$let": {
								"vars": {
									"mongoast__deduplicated__expr__0": {"$add": [1,2,3]}
								},
								"in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}
							}
						}
					}
				}}}
			]`,
		},
		{
			"do not pull out redundancies above the inputs or defaults fields of a $zip",
			`[
				{"$project": {"out": {
					"$zip": {
						"inputs": [
							{"$add": [
								{"$add": [1,2,3]},
								{"$add": [1,2,3]}
							]},
							{"$add": [
								{"$add": [1,2,3]},
								{"$add": [1,2,3]}
							]}
						],
						"useLongestLength": true,
						"defaults": [
							{"$add": [
								{"$add": [1,2,3]},
								{"$add": [1,2,3]}
							]},
							{"$add": [
								{"$add": [1,2,3]},
								{"$add": [1,2,3]}
							]}
						]
					}
				}}}
			]`,
			`[
				{"$project": {"out": {
					"$zip": {
						"inputs": [
							{"$let": {
								"vars": {
									"mongoast__deduplicated__expr__0": {"$add": [1,2,3]}
								},
								"in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}
							}},
							{"$let": {
								"vars": {
									"mongoast__deduplicated__expr__0": {"$add": [1,2,3]}
								},
								"in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}
							}}
						],
						"useLongestLength": true,
						"defaults": [
							{"$let": {
								"vars": {
									"mongoast__deduplicated__expr__0": {"$add": [1,2,3]}
								},
								"in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}
							}},
							{"$let": {
								"vars": {
									"mongoast__deduplicated__expr__0": {"$add": [1,2,3]}
								},
								"in": {"$add": ["$$mongoast__deduplicated__expr__0","$$mongoast__deduplicated__expr__0"]}
							}}
						]
					}
				}}}
			]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in := parsertest.ParsePipeline(tc.input)
			expected := parsertest.ParsePipeline(tc.expected)
			actual := optimizer.RunPasses(context.Background(), in, optimizer.PartialRedundancyElimination)

			expectedStr := parser.DeparsePipeline(expected).String()
			actualStr := parser.DeparsePipeline(actual).String()

			if !cmp.Equal(expectedStr, actualStr) {
				t.Fatalf("\nexpected:\n %s\ngot:\n %s", expectedStr, actualStr)
			}
		})
	}
}
