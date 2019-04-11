package optimizer_test

import (
	"context"
	"testing"

	"github.com/10gen/mongoast/internal/parsertest"
	"github.com/10gen/mongoast/optimizer"
	"github.com/10gen/mongoast/parser"

	"github.com/google/go-cmp/cmp"
)

func TestLetMinimization(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"do not remove single required let binding",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": {"$add": ["$$v1", "$$v1"]}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": {"$add": ["$$v1", "$$v1"]}
				}}}}
			 ]`,
		},
		{
			"do not remove multiple required let bindings",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]},
						"v2": {"$add": ["$y", "$y"]}
					},
					"in": {"$add": ["$$v1", "$$v1", "$$v2", "$$v2"]}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]},
						"v2": {"$add": ["$y", "$y"]}
					},
					"in": {"$add": ["$$v1", "$$v1", "$$v2", "$$v2"]}
				}}}}
			 ]`,
		},
		{
			"remove zero-use let binding",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]},
						"v2": {"$add": ["$y", "$y"]}
					},
					"in": {"$add": ["$$v1", "$$v1"]}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": {"$add": ["$$v1", "$$v1"]}
				}}}}
			 ]`,
		},
		{
			"remove single-use let binding",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]},
						"v2": {"$add": ["$y", "$y"]}
					},
					"in": {"$add": ["$$v1", "$$v1", "$$v2"]}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": {"$add": ["$$v1", "$$v1", {"$add": ["$y", "$y"]}]}
				}}}}
			 ]`,
		},
		{
			"remove constant let binding",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]},
						"v2": 1
					},
					"in": {"$add": ["$$v1", "$$v1", "$$v2", "$$v2"]}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": {"$add": ["$$v1", "$$v1", 1, 1]}
				}}}}
			 ]`,
		},
		{
			"remove column ref let binding",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$y",
						"v2": {"$add": ["$x", "$x"]}
					},
					"in": {"$add": ["$$v1", "$$v1", "$$v2", "$$v2"]}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v2": {"$add": ["$x", "$x"]}
					},
					"in": {"$add": ["$y", "$y", "$$v2", "$$v2"]}
				}}}}
			 ]`,
		},
		{
			"remove variable ref let binding",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": {
						"$let": {
							"vars": {
								"v2": "$$v1",
								"v3": {"$add": ["$y", "$y"]}
							},
							"in": {"$add": ["$$v2", "$$v2", "$$v3", "$$v3"]}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": {
						"$let": {
							"vars": {
								"v3": {"$add": ["$y", "$y"]}
							},
							"in": {"$add": ["$$v1", "$$v1", "$$v3", "$$v3"]}
						}
					}
				}}}}
			 ]`,
		},
		{
			"remove let expression with 0 bindings",
			`[
				{"$project": {"a": {"$let": {
					"vars": {},
					"in": {"$add": ["$x", "$y"]}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$add": ["$x", "$y"]}}}
			 ]`,
		},
		{
			"remove let expression with 0 bindings after removing bindings",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]},
						"v2": {"$add": ["$y", "$y"]}
					},
					"in": {"$add": ["$$v1", "$$v2"]}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$add": [{"$add": ["$x", "$x"]}, {"$add": ["$y", "$y"]}]}}}
			 ]`,
		},
		{
			"remove outer let binding if inner usage is removed",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]},
						"v2": {"$add": ["$y", "$y"]}
					},
					"in": {
						"$let": {
							"vars": {
								"v3": {"$add": ["$$v1", "$$v1"]},
								"v4": {"$add": ["$$v2", "$$v2"]}
							},
							"in": {"$add": ["$$v4", "$$v4"]}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v2": {"$add": ["$y", "$y"]}
					},
					"in": {
						"$let": {
							"vars": {
								"v4": {"$add": ["$$v2", "$$v2"]}
							},
							"in": {"$add": ["$$v4", "$$v4"]}
						}
					}
				}}}}
			 ]`,
		},
		{
			"remove shadowed let bindings appropriately",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": {
						"$let": {
							"vars": {
								"v1": {"$add": ["$$v1", "$$v1"]},
								"v2": {"$add": ["$x", "$$v1"]}
							},
							"in": {"$add": ["$$v1", "$$v2", "$$v2"]}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": {
						"$let": {
							"vars": {
								"v2": {"$add": ["$x", "$$v1"]}
							},
							"in": {"$add": [{"$add": ["$$v1", "$$v1"]}, "$$v2", "$$v2"]}
						}
					}
				}}}}
			 ]`,
		},
		{
			"remove unused outer variable when a used inner variable shadows it",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": {
						"$let": {
							"vars": {
								"v1": {"$add": ["$y", "$y"]}
							},
							"in": {"$add": ["$$v1", "$$v1"]}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$y", "$y"]}
					},
					"in": {"$add": ["$$v1", "$$v1"]}
				}}}}
			 ]`,
		},
		{
			"remove unused outer field ref let binding without substituting it when a used inner variable shadows it",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$let": {
							"vars": {
								"v1": {"$add": ["$y", "$y"]}
							},
							"in": {"$add": ["$$v1", "$$v1"]}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$y", "$y"]}
					},
					"in": {"$add": ["$$v1", "$$v1"]}
				}}}}
			 ]`,
		},
		{
			"remove unused outer field ref and constant let bindings without substituting it when multiple used inner variables shadow it",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": {
						"$let": {
							"vars": {
								"v1": 1
							},
							"in": {
								"$let": {
									"vars": {
										"v1": {"$add": ["$y", "$y"]}
									},
									"in": {"$add": ["$$v1", "$$v1"]}
								}
							}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$y", "$y"]}
					},
					"in": {"$add": ["$$v1", "$$v1"]}
				}}}}
			 ]`,
		},
		{
			"remove unused outer let binding that is shadowed by $map as field",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": {
						"$map": {
							"input": [1, 2, 3],
							"as": "v1",
							"in": {"$add": ["$$v1", "$$v1"]}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$map": {
					"input": [1, 2, 3],
					"as": "v1",
					"in": {"$add": ["$$v1", "$$v1"]}
				}}}}
			 ]`,
		},
		{
			"remove unused outer let binding that is shadowed by $filter as field",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": {
						"$filter": {
							"input": [1, 2, 3],
							"as": "v1",
							"cond": {"$eq": ["$$v1", "$$v1"]}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$filter": {
					"input": [1, 2, 3],
					"as": "v1",
					"cond": {"$eq": ["$$v1", "$$v1"]}
				}}}}
			 ]`,
		},
		{
			"remove unused outer let bindings that are shadowed by this and value in $reduce",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"this": {"$add": ["$x", "$x"]},
						"value": {"$add": ["$y", "$y"]}
					},
					"in": {
						"$reduce": {
							"input": [1, 2, 3],
							"initialValue": 0,
							"in": {"$add": ["$$this", "$$value"]}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$reduce": {
					"input": [1, 2, 3],
					"initialValue": 0,
					"in": {"$add": ["$$this", "$$value"]}
				}}}}
			 ]`,
		},
		{
			"do not remove outer let binding that is used after being shadowed by $map as field",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": [
						{"$map": {
							"input": [1, 2, 3],
							"as": "v1",
							"in": {"$add": ["$$v1", "$$v1"]}
						}},
						{"$add": ["$$v1", "$$v1"]}
					]
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": [
						{"$map": {
							"input": [1, 2, 3],
							"as": "v1",
							"in": {"$add": ["$$v1", "$$v1"]}
						}},
						{"$add": ["$$v1", "$$v1"]}
					]
				}}}}
			 ]`,
		},
		{
			"do not remove outer let binding that is used after being shadowed by $filter as field",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": [
						{"$filter": {
							"input": [1, 2, 3],
							"as": "v1",
							"cond": {"$eq": ["$$v1", "$$v1"]}
						}},
						["$$v1", "$$v1"]
					]
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": [
						{"$filter": {
							"input": [1, 2, 3],
							"as": "v1",
							"cond": {"$eq": ["$$v1", "$$v1"]}
						}},
						["$$v1", "$$v1"]
					]
				}}}}
			 ]`,
		},
		{
			"do not remove outer let bindings that are used after being shadowed by this and value in $reduce",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"this": {"$add": ["$x", "$x"]},
						"value": {"$add": ["$y", "$y"]}
					},
					"in": [
						{"$reduce": {
							"input": [1, 2, 3],
							"initialValue": 0,
							"in": {"$add": ["$$this", "$$value"]}
						}},
						{"$add": ["$$this", "$$value"]},
						{"$add": ["$$this", "$$value"]}
					]
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"this": {"$add": ["$x", "$x"]},
						"value": {"$add": ["$y", "$y"]}
					},
					"in": [
						{"$reduce": {
							"input": [1, 2, 3],
							"initialValue": 0,
							"in": {"$add": ["$$this", "$$value"]}
						}},
						{"$add": ["$$this", "$$value"]},
						{"$add": ["$$this", "$$value"]}
					]
				}}}}
			 ]`,
		},
		{
			"remove but do not substitute constant outer let binding that is shadowed by a nested let and then shadowed by a nested $map and used after the $map while still inside the nested let",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": 1
					},
					"in": {
						"$let": {
							"vars": {
								"v1": {"$add": ["$x", "$x"]}
							},
							"in": [
								{"$map": {
									"input": [1, 2, 3],
									"as": "v1",
									"in": {"$add": ["$$v1", 1]}
								}},
								{"$add": ["$$v1", "$$v1"]}
							]
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": [
						{"$map": {
							"input": [1, 2, 3],
							"as": "v1",
							"in": {"$add": ["$$v1", 1]}
						}},
						{"$add": ["$$v1", "$$v1"]}
					]
				}}}}
			 ]`,
		},
		{
			"remove but do not substitute constant outer let binding that is shadowed by a nested let and then shadowed by a nested $filter and used after the $filter while still inside the nested let",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": 1
					},
					"in": {
						"$let": {
							"vars": {
								"v1": {"$add": ["$x", "$x"]}
							},
							"in": [
								{"$filter": {
									"input": [1, 2, 3],
									"as": "v1",
									"cond": {"$eq": ["$$v1", 1]}
								}},
								{"$add": ["$$v1", "$$v1"]}
							]
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": [
						{"$filter": {
							"input": [1, 2, 3],
							"as": "v1",
							"cond": {"$eq": ["$$v1", 1]}
						}},
						{"$add": ["$$v1", "$$v1"]}
					]
				}}}}
			 ]`,
		},
		{
			"remove but do not substitute constant outer let bindings that are shadowed by a nested let and then shadowed by a nested $reduce and used after the $reduce while still inside the nested let",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"this": 1,
						"value": 2
					},
					"in": {
						"$let": {
							"vars": {
								"this": {"$add": ["$x", "$x"]},
								"value": {"$add": ["$y", "$y"]}
							},
							"in": [
								{"$reduce": {
									"input": [1, 2, 3],
									"initialValue": 0,
									"in": {"$add": ["$$this", "$$value"]}
								}},
								{"$add": ["$$this", "$$value"]},
								{"$add": ["$$this", "$$value"]}
							]
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"this": {"$add": ["$x", "$x"]},
						"value": {"$add": ["$y", "$y"]}
					},
					"in": [
						{"$reduce": {
							"input": [1, 2, 3],
							"initialValue": 0,
							"in": {"$add": ["$$this", "$$value"]}
						}},
						{"$add": ["$$this", "$$value"]},
						{"$add": ["$$this", "$$value"]}
					]
				}}}}
			 ]`,
		},
		{
			"remove but do not substitute constant outer let binding that is shadowed by a nested let and then shadowed by a more deeply nested let and used after the second nested let while still inside the first nested let",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": 1
					},
					"in": {
						"$let": {
							"vars": {
								"v1": {"$add": ["$x", "$x"]}
							},
							"in": [
								{"$let": {
									"vars": {
										"v1": {"$add": ["$y", "$y"]}
									},
									"in": {"$add": ["$$v1", "$$v1"]}
								}},
								{"$add": ["$$v1", "$$v1"]}
							]
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": [
						{"$let": {
							"vars": {
								"v1": {"$add": ["$y", "$y"]}
							},
							"in": {"$add": ["$$v1", "$$v1"]}
						}},
						{"$add": ["$$v1", "$$v1"]}
					]
				}}}}
			 ]`,
		},
		{
			"remove nested let bindings in in",
			// the following nested let variables should be removed:
			// 	- v2 is a field ref
			// 	- v5 is a constant
			// 	- v7 is used once
			// 	- v8 is used 0 times
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": {
						"$let": {
							"vars": {
								"v2": "$y",
								"v3": {
									"$let": {
										"vars": {
											"v4": {
												"$let": {
													"vars": {
														"v6": {
															"$let": {
																"vars": {
																	"v8": {"$add": ["$z", "$z"]},
																	"v9": {"$add": ["$$v1", "$$v1"]}
																},
																"in": {"$add": ["$$v9", "$$v9"]}
															}
														},
														"v7": {"$add": ["$n", "$n"]}
													},
													"in": {"$add": ["$$v6", "$$v6", "$$v7"]}
												}
											},
											"v5": 1
										}, 
										"in": {"$add": ["$$v4", "$$v4", "$$v5", "$$v5"]}
									}
								}
							},
							"in": {"$add": ["$$v2", "$$v2", "$$v3", "$$v3"]}
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"$add": ["$x", "$x"]}
					},
					"in": {
						"$let": {
							"vars": {
								"v3": {
									"$let": {
										"vars": {
											"v4": {
												"$let": {
													"vars": {
														"v6": {
															"$let": {
																"vars": {
																	"v9": {"$add": ["$$v1", "$$v1"]}
																},
																"in": {"$add": ["$$v9", "$$v9"]}
															}
														}
													},
													"in": {"$add": ["$$v6", "$$v6", {"$add": ["$n", "$n"]}]}
												}
											}
										}, 
										"in": {"$add": ["$$v4", "$$v4", 1, 1]}
									}
								}
							},
							"in": {"$add": ["$y", "$y", "$$v3", "$$v3"]}
						}
					}
				}}}}
			 ]`,
		},
		{
			"remove nested let bindings in vars",
			// the following nested let variables should be removed:
			// 	- v2 is a field ref
			// 	- v5 is a constant
			// 	- v7 is used once
			// 	- v8 is used 0 times
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {
							"$let": {
								"vars": {
									"v2": "$x",
									"v3": {
										"$let": {
											"vars": {
												"v4": {
													"$let": {
														"vars": {
															"v6": {
																"$let": {
																	"vars": {
																		"v8": {"$add": ["$y", "$y"]},
																		"v9": {"$add": ["$z", "$z"]}
																	},
																	"in": {"$add": ["$$v9", "$$v9"]}
																}
															},
															"v7": {"$add": ["$n", "$n"]}
														},
														"in": {"$add": ["$$v6", "$$v6", "$$v7"]}
													}
												},
												"v5": 1
											}, 
											"in": {"$add": ["$$v4", "$$v4", "$$v5", "$$v5"]}
										}
									}
								},
								"in": {"$add": ["$$v2", "$$v2", "$$v3", "$$v3"]}
							}
						}
					},
					"in": {"$add": ["$$v1", "$$v1"]}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {
							"$let": {
								"vars": {
									"v3": {
										"$let": {
											"vars": {
												"v4": {
													"$let": {
														"vars": {
															"v6": {
																"$let": {
																	"vars": {
																		"v9": {"$add": ["$z", "$z"]}
																	},
																	"in": {"$add": ["$$v9", "$$v9"]}
																}
															}
														},
														"in": {"$add": ["$$v6", "$$v6", {"$add": ["$n", "$n"]}]}
													}
												}
											}, 
											"in": {"$add": ["$$v4", "$$v4", 1, 1]}
										}
									}
								},
								"in": {"$add": ["$x", "$x", "$$v3", "$$v3"]}
							}
						}
					},
					"in": {"$add": ["$$v1", "$$v1"]}
				}}}}
			 ]`,
		},
		{
			"do not substitute single-use let binding into a FieldRef parent",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"f": 1}
					},
					"in": "$$v1.f"
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": {"f": 1}
					},
					"in": "$$v1.f"
				}}}}
			 ]`,
		},
		{
			"substitute single-use let binding into a FieldRef parent if let binding is a FieldRef",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"v1": "$x"
					},
					"in": "$$v1.f"
				}}}}
			 ]`,
			`[
				{"$project": {"a": "$x.f"}}
			 ]`,
		},
		{
			"substitute single-use let binding into a FieldRef parent if let binding is a VariableRef",
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
							"in": "$$v2.f"
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": "$x.f"}}
			 ]`,
		},
		{
			"fully minimize nested lets with multiple shadowed variables",
			`[
				{"$project": {"a": {"$let": {
					"vars": {
						"x": 1,
						"y": 2
					},
					"in": {
						"$let": {
							"vars": {
								"x": 0,
								"y": "$$x"
							},
							"in": "$$y"
						}
					}
				}}}}
			 ]`,
			`[
				{"$project": {"a": {"$literal": 1}}}
			 ]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in := parsertest.ParsePipeline(tc.input)
			expected := parsertest.ParsePipeline(tc.expected)
			actual := optimizer.RunPasses(context.Background(), in, optimizer.LetMinimization)

			expectedStr := parser.DeparsePipeline(expected).String()
			actualStr := parser.DeparsePipeline(actual).String()

			if !cmp.Equal(expectedStr, actualStr) {
				t.Fatalf("\nexpected:\n %s\ngot:\n %s", expectedStr, actualStr)
			}
		})
	}
}
