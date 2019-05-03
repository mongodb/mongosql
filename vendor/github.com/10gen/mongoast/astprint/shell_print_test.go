package astprint_test

import (
	"regexp"
	"testing"

	"github.com/10gen/mongoast/astprint"
	"github.com/10gen/mongoast/internal/parsertest"

	"github.com/google/go-cmp/cmp"
)

func TestShellPrint(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			// Symbol is a little problematic wrt the shell. It works, but it's not printable,
			// and it doesn't execute, so I've left it in a separate $addFields, to make it easier
			// to test the output via shell pasting.
			"test constants",
			`[
				{"$addFields": {
					"a": 3.14,
					"b": "he\\llo",
					"c": {"a\\a\n": 1.0},
					"d": [1.0, 2.0, 3.0],
					"e": {"$binary": {"base64": "aGVsbG8gd29ybGQ=", "subType": "0"}},
					"f": {"$undefined":true},
					"g": {"$oid": "5cacfb7462400d2d7b114ec2"},
					"h": true,
					"i": {"$date": {"$numberLong": "1554840954963"}},
					"j": null,
					"l": {"$dbPointer": {"$ref": "hello", "$id": {"$oid": "5cacfb7462400d2d7b114ec2"}}},
					"m": {"$code": "x=3"},
					"n": {"$regularExpression": {"pattern": "hello/world\\s+", "options": "i"}},
					"o": {"$code": "print(x)", "$scope": {"x": 4}},
					"p": {"$numberInt": "32"},
					"q": {"$timestamp": {"t": 124, "i": 534}},
					"r": {"$numberLong": "50000000000"},
					"s": {"$numberDecimal": "3.14"},
					"t": {"$minKey": 1},
					"u": {"$maxKey": 1}
				}},
				{"$addFields": {
					"n": {"$symbol": "x"},
					"lit": {"$literal": {"$add": [2,3]}}
				}}
			 ]`,
			`[
			    {"$addFields": {
					"a": 3.14,
					"b": "he\\llo",
					"c": {"a\\a\n": 1.0},
					"d": [1.0,2.0,3.0],
					"e": BinData(0, "aGVsbG8gd29ybGQ="),
					"f": undefined,
					"g": ObjectId("5cacfb7462400d2d7b114ec2"),
					"h": true,
					"i": ISODate("2019-04-09T16:15:54.963"),
					"j": null,
					"l": DBPointer("hello", ObjectId("5cacfb7462400d2d7b114ec2")),
					"m": Code("x=3"),
					"n": /hello\/world\s+/i,
					"o": Code("print(x)", {"x": NumberInt("4")}),
					"p": NumberInt("32"),
					"q": Timestamp(124,534),
					"r": NumberLong("50000000000"),
					"s": NumberDecimal("3.14"),
					"t": MinKey,
					"u": MaxKey
					}
				},
            	{"$addFields": {
					"n": Symbol("x"),
					"lit": {"$literal": {"$add": [NumberInt("2"),NumberInt("3")]}}
					}
				}
			 ]`,
		}, {"test agg expressions",
			`[
				{"$addFields": {
					"a": [1, 2, {"$arrayElemAt": [[1,2,3],0]}],
					"b": {"$add": [1,2,3]},
					"c": {"$subtract": [1,2]},
					"d": {"$or": [1,2]},
					"e": "$a.b.c",
					"f": {"$let": {
						"vars": {"x": 3},
						"in": {"$subtract": ["$$x", 1]}
					}},
					"g": {"$cond": [1,2,3]},
					"h": {"$cond": {"if": 1, "then": 2, "else": 3}}
				}}
			 ]`,
			`[
				{"$addFields": {
					"a": [
						NumberInt("1"),
						NumberInt("2"),
						{"$arrayElemAt": [
											[NumberInt("1"),
											 NumberInt("2"),
											 NumberInt("3")
											],
											NumberInt("0")
										 ]
						}],
					"b": {"$add": [
									NumberInt("1"),
									NumberInt("2"),
									NumberInt("3")
								  ]
							  },
					"c": {"$subtract": [
										NumberInt("1"),
										NumberInt("2")
									   ]
							},
					"d": {"$or": [
							NumberInt("1"),
							NumberInt("2")
							]
						},
					"e": "$a.b.c",
					"f": {"$let":
							{"vars": {"x": NumberInt("3")},
							 "in": {"$subtract": ["$$x",NumberInt("1")]}}},
					"g": {"$cond": {"if": NumberInt("1"),
									"then": NumberInt("2"),
									"else": NumberInt("3")}},
					"h": {"$cond": {"if": NumberInt("1"),
									"then": NumberInt("2"),
									"else": NumberInt("3")}}}}
			]`,
		}, {"test match expressions",
			// This adds not-strictly necessary "$and" and "$eq" operators, but this is still correct.
			// If we do not want this behavior, it should actually be fixed in Deparse, rather than
			// in ShellPrint.
			`[
				{"$match": {
					"a": {"$ne": 1},
					"b": {"$gt": 2},
					"$expr": {"$let": {
						"vars": {"x": 3},
						"in": {"$subtract": ["$$x", 1]}
					}}
				}}
			]`,
			`[
				{"$match": {
					"$and": [
						{"a": {"$ne": NumberInt("1")}},
						{"b": {"$gt": NumberInt("2")}},
						{"$expr": {"$let": {
							"vars": {"x": NumberInt("3")},
							"in": {"$subtract": ["$$x",NumberInt("1")]}
								}
							}
						}
						]
					}
				}
			]`,
		}, {"test project",
			`[
				{"$project": {"a": 1, "b": 1, "c": "hello"}}
			]`,
			// Note that we want $literal around "hello" to support versions of mongodb < 3.4.
			`[
				{"$project": {"a": NumberInt("1"),"b": NumberInt("1"), "c": {"$literal": "hello"}}}
			]`,
		}, {"test bucket",
			`[
				{"$bucket": {
					"groupBy": "$price",
					"boundaries": [ 0, 200, 400 ],
					"default": "Other",
					"output": {
						"count": { "$sum": 1 },
						"titles" : { "$push": "$title" }
						}
					}
				}
			]`,
			`[
				{"$bucket": {
					"groupBy": "$price",
					"boundaries": [NumberInt("0"),NumberInt("200"),NumberInt("400")],
					"default": "Other",
					"output": {
						"count": {"$sum": NumberInt("1")},
						"titles": {"$push": "$title"}
						}
					}
				}
			]`,
		}, {"test bucket no optionals",
			`[
				{"$bucket": {
					"groupBy": "$price",
					"boundaries": [ 0, 200, 400 ]
					}
				}
			]`,
			`[
				{"$bucket": {
					"groupBy": "$price",
					"boundaries": [NumberInt("0"),NumberInt("200"),NumberInt("400")]
					}
				}
			]`,
		}, {"test bucketAuto",
			`[
				{"$bucketAuto": {
					"groupBy": "$price",
					"buckets": 200,
					"output": {
						"count": { "$sum": 1 },
						"titles" : { "$push": "$title" }
						}
					},
					"granularity": "E6"
				}
			]`,
			`[
				{"$bucketAuto": {
					"groupBy": "$price",
					"buckets": NumberLong("200"),
					"output": {
						"count": {"$sum": NumberInt("1")},
						"titles": {"$push": "$title"}
						}
					}
				}
			]`,
		}, {"test bucketAuto no optionals",
			`[
				{"$bucketAuto": {
					"groupBy": "$price",
					"buckets": 200
					}
				}
			]`,
			`[
				{"$bucketAuto": {
					"groupBy": "$price",
					"buckets": NumberLong("200")
					}
				}
			]`,
		}, {"test collStats",
			`[{"$collStats": {
					"latencyStats": {
						"histograms": true
					},
					"storageStats": {},
					"count": {}
					}
				}
			]`,
			`[{"$collStats": {
					"latencyStats": {"histograms": true},
					"storageStats": {},
					"count": {}
				}
			  }
			]`,
		}, {"test count",
			`[{"$count" : "hello"}]`,
			`[{"$count": "hello"}]`,
		}, {"test facet",
			// This adds not-strictly necessary "$eq" operators, but this is still correct.  If we
			// do not want this behavior, it should actually be fixed in Deparse, rather than in
			// ShellPrint.
			`[
				{"$facet": {
					"a1": [{"$match": {"a": 1}}],
					"a2": [{"$match": {"a": 2}}]
					}
				}
			]`,
			`[
			    {"$facet": {
					"a1": [{"$match": {"a": {"$eq": NumberInt("1")}}}],
					"a2": [{"$match": {"a": {"$eq": NumberInt("2")}}}]
					}
				}
			]`,
		}, {"test group",
			`[
				{"$group": {
					"_id": null,
					"a2": {"$sum": 1},
					"a3": {"$push": "$b"}
					}
				}
			]`,
			`[
				{"$group": {
					"_id":{"$literal": null},
					"a2": {"$sum": NumberInt("1")},
					"a3": {"$push": "$b"}
					}
				}
			]`,
		}, {"test limit",
			`[
				{"$limit": 100}
			]`,
			`[
				{"$limit": NumberLong("100")}
			]`,
		}, {"test lookup",
			`[
				{"$lookup": {
         			"from": "inventory",
         			"localField": "item",
         			"foreignField": "sku",
         			"as": "inventory_docs"
       				}
  				}
			]`,
			`[
				{"$lookup": {
					"from": "inventory",
					"localField": "item",
					"foreignField": "sku",
					"as": "inventory_docs"
					}
				}
			]`,
		}, {"test expressive lookup, show that literal is removed in addFields",
			`[
				{"$lookup":
	         		{
	           			"from": "warehouses",
	           			"let": { "order_item": "$item", "order_qty": "$ordered" },
	           			"pipeline": [
	              			{ "$match":
	                 			{ "$expr":
	                    			{ "$and":
	                       				[
	                      				   { "$eq": [ "$stock_item",  "$$order_item" ] },
	                     				   { "$gte": [ "$instock", "$$order_qty" ] }
	                       				]
	                    			}
	                 			}
	              			},
	              			{ "$project": { "stock_item": 0, "_id": 0} },
							{ "$addFields": {"hello": "world"}}
	           			],
	           			"as": "stockdata"
	         		}
	    		}
			]`,
			`[
				{"$lookup": {
					"from": "warehouses",
					"let": {"order_item": "$item","order_qty": "$ordered"},
					"pipeline": [
							{"$match": {"$expr": {"$and": [
								{"$eq": ["$stock_item","$$order_item"]},
								{"$gte": ["$instock","$$order_qty"]}]}}
							},
							{"$project": {"stock_item": NumberInt("0"),
										  "_id": NumberInt("0")}
									  },
							{"$addFields": {"hello": "world"}}
						],
					"as": "stockdata"}}
			]`,
		}, {"test redact",
			`[
			 { "$redact": {
			    "$cond": {
					"if": { "$gt": [ { "$size": { "$setIntersection": [ "$tags", ["userAccess"] ] } }, 0 ] },
					"then": "$$DESCEND",
					"else": "$$PRUNE"
					}
				}
			}
			]`,
			`[
			{"$redact": {
				"$cond": {
					"if": {"$gt": [{"$size": {"$setIntersection": ["$tags",["userAccess"]]}},NumberInt("0")]},
					"then": "$$DESCEND",
					"else": "$$PRUNE"
					}
				}
			}
			]`,
		}, {"test replaceRoot",
			`[
			 { "$replaceRoot": {
			 		"newRoot": {"b": "$c", "d": {"$add": [1, 2, "$e"]}}
				}
			}
			]`,
			`[
			 { "$replaceRoot": {
					"newRoot": {"b": "$c",
								"d": {"$add": [
										NumberInt("1"),
										NumberInt("2"),
										"$e"
										]
									}
							}
				}
			}
			]`,
		}, {"test sample",
			`[
				{"$sample": {"size": 100}}
			]`,
			`[
				{"$sample": {"size": NumberLong("100")}}
			]`,
		}, {"test skip",
			`[
				{"$skip": 100}
			]`,
			`[
				{"$skip": NumberLong("100")}
			]`,
		}, {"test sort",
			`[
				{"$sort": {"a": 1}}
			]`,
			`[
				{"$sort": {"a": NumberInt("1")}}
			]`,
		}, {"test sortByCount",
			`[
				{"$sortByCount": "$a"}
			]`,
			`[
				{"$sortByCount": "$a"}
			]`,
		}, {"test sortedMerge",
			`[
				{"$sortedMerge": {"a": 1, "b": -1}}
			]`,
			`[
				{"$sortedMerge": {
					"a": NumberInt("1"),
					"b": NumberInt("-1")
					}
				}
			]`,
		}, {"test unwind",
			`[
				{"$unwind": {
					"path": "$a.b",
					"includeArrayIndex": "funny",
					"preserveNullAndEmptyArrays": true
					}
				}
			]`,
			`[
				{"$unwind": {
					"path": "$a.b",
					"includeArrayIndex": "funny",
					"preserveNullAndEmptyArrays": true
					}
				}
			]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in := parsertest.ParsePipeline(tc.input)
			actual := astprint.ShellString(in)
			// just match modulo white space, too difficult to format.
			re := regexp.MustCompile(`\s+`)
			actualTest := re.ReplaceAllString(actual, ``)

			expected := re.ReplaceAllString(tc.expected, ``)

			if !cmp.Equal(expected, actualTest) {
				t.Fatalf("\nexpected:\n %s\ngot:\n %s", tc.expected, actual)
			}
		})
	}
}
