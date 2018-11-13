package evaluator

import (
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/catalog"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
	"github.com/stretchr/testify/require"
)

// Tests for any private functions in the evaluator package.

// TestGetFastPlanStageTest tests the functionality of getFastPlanStage.
func TestGetFastPlanStageTest(t *testing.T) {
	// noRecursiveUnionDistinct tests that only UnionAlls exist under a UnionDistinct.
	noRecursiveUnionDistinct := func(plan PlanStage) bool {
		var aux func(plan PlanStage, underDistinct bool) bool
		aux = func(plan PlanStage, underDistinct bool) bool {
			if _, ok := plan.(*MongoSourceStage); ok {
				return true
			} else if up, ok := plan.(*UnionStage); ok {
				if up.kind == UnionDistinct {
					if underDistinct {
						return false
					}
					return aux(up.left, true) && aux(up.right, true)
				}
				// UnionAll
				return aux(up.left, underDistinct) && aux(up.right, underDistinct)
			}
			return false
		}
		return aux(plan, false)
	}

	type fastPlanTest struct {
		input, expected PlanStage
	}

	// testGetFastPlan is a test helper function for testing getFastPlanStage. The
	// parameter successful is true when we expect getFastPlanStage to succeed, is32
	// tells us if we are testing on MongoDB version 3.2 or not, a distinction that
	// matters for UnionDistinct unions.
	testGetFastPlan := func(t *testing.T, actual, expected PlanStage, successful, is32 bool) {
		req := require.New(t)
		actual, ok := getFastPlanStage(actual, is32, false)
		if successful {
			// This first part of the test is actually testing that the expected
			// plan is correct before we compare the actual plan to the expected plan.
			req.True(noRecursiveUnionDistinct(expected),
				"there should be no UnionDistinct under the top level in expected plan.")
			req.True(ok, "failed to optimize input.")
			req.NotNil(actual, "getFastPlan returned nil.")
			req.Equal(expected, actual)
		} else {
			req.False(ok)
			req.Nil(actual)
		}
	}

	runTests := func(
		t *testing.T,
		subTestName string,
		tests []fastPlanTest,
		successful,
		is32 bool) {
		t.Run(subTestName, func(t *testing.T) {
			req := require.New(t)
			for _, test := range tests {
				// On successful tests, the expected Plan should have is32 set
				// to the passed value, if the kind is UnionDistinct. This keeps
				// us from having to rewrite the tests for MongoDB version 3.2.
				if successful {
					up, ok := test.expected.(*UnionStage)
					req.True(ok, "expected must be a UnionStage")
					if up.kind == UnionDistinct {
						up.is32 = is32
					}
				}
				testGetFastPlan(t, test.input, test.expected, successful, is32)
			}
		})
	}

	// We will use the same mongoSourceStage for all mongoSourceStages
	mongoSourceStage := &MongoSourceStage{
		selectIDs:  []int{0},
		dbName:     "foo",
		tableNames: []string{"bar"},
		aliasNames: []string{"biz"},
		tableType:  catalog.TableType("view"),
	}

	// First successfully fastPlan optimizable tests on server version > 3.2
	successfulUnionAllTests := []fastPlanTest{
		{
			// Optimizable two-way Union All.
			// input
			&ProjectStage{
				source: &UnionStage{
					left:  mongoSourceStage,
					right: mongoSourceStage,
					kind:  UnionAll,
				},
			},
			// expected
			&UnionStage{
				left:  mongoSourceStage,
				right: mongoSourceStage,
				kind:  UnionAll,
			},
		},
		{
			// Optimizable three-way Union All.
			// input
			&ProjectStage{
				source: &UnionStage{
					left: &ProjectStage{
						source: &UnionStage{
							left:  mongoSourceStage,
							right: mongoSourceStage,
							kind:  UnionAll,
						},
					},
					right: mongoSourceStage,
					kind:  UnionAll,
				}},
			// expected
			&UnionStage{
				left: &UnionStage{
					left:  mongoSourceStage,
					right: mongoSourceStage,
					kind:  UnionAll,
				},
				right: mongoSourceStage,
				kind:  UnionAll,
			},
		},
		{
			// Optimizable four-way Union All.
			// input
			&ProjectStage{
				source: &UnionStage{
					left: &ProjectStage{
						source: &UnionStage{
							left: &ProjectStage{
								source: &UnionStage{
									left:  mongoSourceStage,
									right: mongoSourceStage,
									kind:  UnionAll,
								},
							},
							right: mongoSourceStage,
							kind:  UnionAll,
						},
					},
					right: mongoSourceStage,
					kind:  UnionAll,
				}},
			// expected
			&UnionStage{
				left: &UnionStage{
					left: &UnionStage{
						left:  mongoSourceStage,
						right: mongoSourceStage,
						kind:  UnionAll,
					},
					right: mongoSourceStage,
					kind:  UnionAll,
				},
				right: mongoSourceStage,
				kind:  UnionAll,
			},
		},
	}

	// First run the tests with the MongoDB version 3.4+.
	runTests(t, "MongoDB 3.4+ Successful Union All getFastPlan tests",
		successfulUnionAllTests, true, false)

	// Now rerun the tests with MongoDB version 3.2.
	runTests(t, "MongoDB 3.2 Successful Union All getFastPlan tests",
		successfulUnionAllTests, true, true)

	successfulUnionDistinctTests := []fastPlanTest{
		{
			// Optimizable two-way Union Distinct.
			// input
			&ProjectStage{
				source: &GroupByStage{
					source: &UnionStage{
						left:  mongoSourceStage,
						right: mongoSourceStage,
						kind:  UnionDistinct,
					},
				},
			},
			// expected
			&UnionStage{
				left:  mongoSourceStage,
				right: mongoSourceStage,
				kind:  UnionDistinct,
			},
		},
		{
			// Optimizable three-way Union Distinct.
			// input
			&ProjectStage{
				source: &GroupByStage{
					source: &UnionStage{
						left: &ProjectStage{
							source: &GroupByStage{
								source: &UnionStage{
									left:  mongoSourceStage,
									right: mongoSourceStage,
									kind:  UnionDistinct,
								},
							},
						},
						right: mongoSourceStage,
						kind:  UnionDistinct,
					},
				},
			},
			// expected
			&UnionStage{
				left: &UnionStage{
					left:  mongoSourceStage,
					right: mongoSourceStage,
					kind:  UnionAll,
				},
				right: mongoSourceStage,
				kind:  UnionDistinct,
			},
		},
		{
			// Optimizable four-way Union Distinct.
			// input
			&ProjectStage{
				source: &GroupByStage{
					source: &UnionStage{
						left: &ProjectStage{
							source: &GroupByStage{
								source: &UnionStage{
									left: &ProjectStage{
										source: &GroupByStage{
											source: &UnionStage{
												left:  mongoSourceStage,
												right: mongoSourceStage,
												kind:  UnionDistinct,
											},
										},
									},
									right: mongoSourceStage,
									kind:  UnionDistinct,
								},
							},
						},
						right: mongoSourceStage,
						kind:  UnionDistinct,
					},
				},
			},
			// expected
			&UnionStage{
				left: &UnionStage{
					left: &UnionStage{
						left:  mongoSourceStage,
						right: mongoSourceStage,
						kind:  UnionAll,
					},
					right: mongoSourceStage,
					kind:  UnionAll,
				},
				right: mongoSourceStage,
				kind:  UnionDistinct,
			},
		},
	}

	// Run Union Distinct tests in MongoDB version 3.4+.
	runTests(t, "MongoDB 3.4+ Successful Union Distinct getFastPlan tests",
		successfulUnionDistinctTests, true, false)

	// Run the Union Distinct tests with MongoDB version 3.2.
	runTests(t, "MongoDB 3.2 Successful Union Distinct getFastPlan tests",
		successfulUnionDistinctTests, true, true)

	// Next is not optimizable plans.
	notOptimizableTests := []fastPlanTest{
		{
			// A Union Distinct must have a GroupByStage above it.
			&ProjectStage{
				source: &UnionStage{
					left:  mongoSourceStage,
					right: mongoSourceStage,
					kind:  UnionDistinct,
				},
			},
			nil,
		},
		{
			// Any Union must have a ProjectStage above it.
			&UnionStage{
				left:  mongoSourceStage,
				right: mongoSourceStage,
				kind:  UnionAll,
			},
			nil,
		},
	}

	// Run notOptimizable tests, these should return nil.
	runTests(t, "Not Optimizable getFastPlan tests",
		notOptimizableTests, false, false)
}

// TestEnsureFastPlanProjectInvariant tests the
// functionality of ensureFastPlanProjectInvariant.
func TestEnsureFastPlanProjectInvariant(t *testing.T) {
	req := require.New(t)
	mongoSourceStage := &MongoSourceStage{
		selectIDs:  []int{0},
		dbName:     "foo",
		tableNames: []string{"bar"},
		aliasNames: []string{"biz"},
		tableType:  catalog.TableType("view"),
		pipeline: bsonutil.NewDArray(
			bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewD(bsonutil.NewDocElem("foo", 1)))),
		),
	}

	// Check to make sure id:0 is added to the final $project of
	// a mongoSourceStage.
	fastPlan, ok := getFastPlanStage(mongoSourceStage, false, false)
	req.NotNil(fastPlan)
	req.True(ok)
	ensureFastPlanProjectInvariant(fastPlan)
	ms, ok := fastPlan.(*MongoSourceStage)
	req.True(ok)
	req.Equal(0, ms.pipeline[0][0].Value.(bson.D).Map()["_id"],
		"_id:0 must be added to project")

	mongoSourceStage1 := &MongoSourceStage{
		selectIDs:  []int{0},
		dbName:     "foo",
		tableNames: []string{"bar"},
		aliasNames: []string{"biz"},
		tableType:  catalog.TableType("view"),
		pipeline: bsonutil.NewDArray(
			bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewD(bsonutil.NewDocElem("foo", 1)))),
		),
	}
	mongoSourceStage2 := &MongoSourceStage{
		selectIDs:  []int{0},
		dbName:     "foo",
		tableNames: []string{"bar"},
		aliasNames: []string{"biz"},
		tableType:  catalog.TableType("view"),
		pipeline: bsonutil.NewDArray(
			bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewD(bsonutil.NewDocElem("bar", 2)))),
		),
	}
	optimizableUnion := &ProjectStage{
		source: &UnionStage{
			left:  mongoSourceStage1,
			right: mongoSourceStage2,
			kind:  UnionAll,
		},
	}

	// Check to make sure id:0 is added to the final $project of
	// a both mongoSourceStages in a Union.
	fastPlan, ok = getFastPlanStage(optimizableUnion, false, false)
	req.NotNil(fastPlan)
	req.True(ok)
	ensureFastPlanProjectInvariant(fastPlan)
	us, ok := fastPlan.(*UnionStage)
	req.True(ok)
	ms1, ok := us.left.(*MongoSourceStage)
	req.True(ok)
	req.Equal(0, ms1.pipeline[0][0].Value.(bson.D).Map()["_id"],
		"_id:0 must be added to left stage project")
	ms2, ok := us.right.(*MongoSourceStage)
	req.True(ok)
	req.Equal(0, ms2.pipeline[0][0].Value.(bson.D).Map()["_id"],
		"_id:0 must be added to right stage project")
}

func TestBuildProjectBodyForMongoSource(t *testing.T) {
	req := require.New(t)
	mkFieldSet := func(in []string) map[string]struct{} {
		ret := make(map[string]struct{}, len(in))
		for _, field := range in {
			ret[field] = struct{}{}
		}
		return ret
	}

	type test struct {
		inputFields         []string
		inputEvalType       EvalType
		inputIs34           bool
		expectedFields      []string
		expectedBody        bson.D
		expectedHasEmbedded bool
	}

	runTests := func(tests []test) {
		for _, testCase := range tests {
			fakeColumns := make(Columns, len(testCase.inputFields))
			for i := range fakeColumns {
				fakeColumns[i] = &Column{
					ColumnType: ColumnType{
						EvalType: testCase.inputEvalType,
					},
				}
			}
			projectBody, fields, hasEmbedded := buildProjectBodyForMongoSource(
				testCase.inputFields, mkFieldSet(testCase.inputFields), fakeColumns,
				testCase.inputIs34)
			req.Equal(testCase.expectedFields, fields)
			req.Equal(testCase.expectedHasEmbedded, hasEmbedded)
			if hasEmbedded {
				req.Equal(testCase.expectedBody, projectBody)
			}
		}
	}

	nonEmbeddedFields := []string{"a", "b", "c", "d"}
	noConflictEmbeddedFields := []string{"a", "b", "c.a", "c.d"}
	expectedNoConflictEmbeddedFields := []string{"a", "b", "c_DOT_a", "c_DOT_d"}
	conflictedEmbeddedFields := []string{"a_DOT_b", "a_DOT_c", "a_DOT_c0", "a.b", "a.c", "b"}
	expectedConflictedEmbeddedFields := []string{"a_DOT_b", "a_DOT_c", "a_DOT_c0",
		"a_DOT_b0", "a_DOT_c1", "b"}

	// buildProjectBodryForMongoSource overwrites its input, so we need to redeclare these
	// two inputs.
	noConflictEmbeddedFields32 := []string{"a", "b", "c.a", "c.d"}
	conflictedEmbeddedFields32 := []string{"a_DOT_b", "a_DOT_c", "a_DOT_c0", "a.b", "a.c", "b"}

	noConflictEmbeddedFieldsArr := []string{"a", "b", "c.a.1", "c.d"}
	expectedNoConflictEmbeddedFieldsArr := []string{"a", "b", "c_DOT_a_DOT_1", "c_DOT_d"}
	conflictedEmbeddedFieldsArr := []string{"a_DOT_b", "a_DOT_c.1", "a_DOT_c0",
		"a.b", "a.c", "b"}
	expectedConflictedEmbeddedFieldsArr := []string{"a_DOT_b", "a_DOT_c_DOT_1", "a_DOT_c0",
		"a_DOT_b0", "a_DOT_c", "b"}

	tests := []test{
		// tests for 3.4+ which should generate addFields bodies
		{inputFields: nonEmbeddedFields,
			inputEvalType:       EvalInt64,
			inputIs34:           true,
			expectedFields:      nonEmbeddedFields,
			expectedBody:        bsonutil.NewD(),
			expectedHasEmbedded: false},

		{inputFields: noConflictEmbeddedFields,
			inputEvalType:  EvalInt64,
			inputIs34:      true,
			expectedFields: expectedNoConflictEmbeddedFields,
			expectedBody: bsonutil.NewD(
				bsonutil.NewDocElem("c_DOT_a", "$c.a"),
				bsonutil.NewDocElem("c_DOT_d", "$c.d"),
			),
			expectedHasEmbedded: true},

		{inputFields: conflictedEmbeddedFields,
			inputEvalType:  EvalInt64,
			inputIs34:      true,
			expectedFields: expectedConflictedEmbeddedFields,
			expectedBody: bsonutil.NewD(
				bsonutil.NewDocElem("a_DOT_b0", "$a.b"),
				bsonutil.NewDocElem("a_DOT_c1", "$a.c"),
			),
			expectedHasEmbedded: true},

		//tests for pre-3.4+ which should generate project bodies
		{inputFields: nonEmbeddedFields,
			inputEvalType:       EvalInt64,
			inputIs34:           false,
			expectedFields:      nonEmbeddedFields,
			expectedBody:        bsonutil.NewD(),
			expectedHasEmbedded: false},

		{inputFields: noConflictEmbeddedFields32,
			inputEvalType:  EvalInt64,
			inputIs34:      false,
			expectedFields: expectedNoConflictEmbeddedFields,
			expectedBody: bsonutil.NewD(
				bsonutil.NewDocElem("a", true),
				bsonutil.NewDocElem("b", true),
				bsonutil.NewDocElem("c_DOT_a", "$c.a"),
				bsonutil.NewDocElem("c_DOT_d", "$c.d"),
			),
			expectedHasEmbedded: true},

		{inputFields: conflictedEmbeddedFields32,
			inputEvalType:  EvalInt64,
			inputIs34:      false,
			expectedFields: expectedConflictedEmbeddedFields,
			expectedBody: bsonutil.NewD(
				bsonutil.NewDocElem("a_DOT_b", true),
				bsonutil.NewDocElem("a_DOT_c", true),
				bsonutil.NewDocElem("a_DOT_c0", true),
				bsonutil.NewDocElem("a_DOT_b0", "$a.b"),
				bsonutil.NewDocElem("a_DOT_c1", "$a.c"),
				bsonutil.NewDocElem("b", true),
			),
			expectedHasEmbedded: true},

		// tests for 3.4+ which should generate addFields bodies,
		// with EvalArrNumeric type for some of the fields.
		{inputFields: noConflictEmbeddedFieldsArr,
			inputEvalType:  EvalArrNumeric,
			inputIs34:      true,
			expectedFields: expectedNoConflictEmbeddedFieldsArr,
			expectedBody: bsonutil.NewD(
				bsonutil.NewDocElem("c_DOT_a_DOT_1", bsonutil.NewM(bsonutil.NewDocElem("$arrayElemAt", bsonutil.NewArray(
					"$c.a",
					1,
				)))),
				bsonutil.NewDocElem("c_DOT_d", "$c.d"),
			),
			expectedHasEmbedded: true},

		{inputFields: conflictedEmbeddedFieldsArr,
			inputEvalType:  EvalArrNumeric,
			inputIs34:      true,
			expectedFields: expectedConflictedEmbeddedFieldsArr,
			expectedBody: bsonutil.NewD(
				bsonutil.NewDocElem("a_DOT_c_DOT_1", bsonutil.NewM(bsonutil.NewDocElem("$arrayElemAt", bsonutil.NewArray(
					"$a_DOT_c",
					1,
				)))),
				bsonutil.NewDocElem("a_DOT_b0", "$a.b"),
				bsonutil.NewDocElem("a_DOT_c", "$a.c"),
			),
			expectedHasEmbedded: true},
	}

	runTests(tests)
}
