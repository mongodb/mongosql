package evaluator

import (
	"testing"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
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
		tableType:  "view",
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
		tableType:  "view",
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
		tableType:  "view",
		pipeline: bsonutil.NewDArray(
			bsonutil.NewD(bsonutil.NewDocElem("$project", bsonutil.NewD(bsonutil.NewDocElem("foo", 1)))),
		),
	}
	mongoSourceStage2 := &MongoSourceStage{
		selectIDs:  []int{0},
		dbName:     "foo",
		tableNames: []string{"bar"},
		aliasNames: []string{"biz"},
		tableType:  "view",
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
		inputEvalType       types.EvalType
		inputIs34           bool
		expectedFields      []string
		expectedBody        bson.D
		expectedHasEmbedded bool
	}

	runTests := func(tests []test) {
		for _, testCase := range tests {
			fakeColumns := make(results.Columns, len(testCase.inputFields))
			for i := range fakeColumns {
				fakeColumns[i] = &results.Column{
					ColumnType: results.ColumnType{
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
			inputEvalType:       types.EvalInt64,
			inputIs34:           true,
			expectedFields:      nonEmbeddedFields,
			expectedBody:        bsonutil.NewD(),
			expectedHasEmbedded: false},

		{inputFields: noConflictEmbeddedFields,
			inputEvalType:  types.EvalInt64,
			inputIs34:      true,
			expectedFields: expectedNoConflictEmbeddedFields,
			expectedBody: bsonutil.NewD(
				bsonutil.NewDocElem("c_DOT_a", "$c.a"),
				bsonutil.NewDocElem("c_DOT_d", "$c.d"),
			),
			expectedHasEmbedded: true},

		{inputFields: conflictedEmbeddedFields,
			inputEvalType:  types.EvalInt64,
			inputIs34:      true,
			expectedFields: expectedConflictedEmbeddedFields,
			expectedBody: bsonutil.NewD(
				bsonutil.NewDocElem("a_DOT_b0", "$a.b"),
				bsonutil.NewDocElem("a_DOT_c1", "$a.c"),
			),
			expectedHasEmbedded: true},

		//tests for pre-3.4+ which should generate project bodies
		{inputFields: nonEmbeddedFields,
			inputEvalType:       types.EvalInt64,
			inputIs34:           false,
			expectedFields:      nonEmbeddedFields,
			expectedBody:        bsonutil.NewD(),
			expectedHasEmbedded: false},

		{inputFields: noConflictEmbeddedFields32,
			inputEvalType:  types.EvalInt64,
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
			inputEvalType:  types.EvalInt64,
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
			inputEvalType:  types.EvalArrNumeric,
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
			inputEvalType:  types.EvalArrNumeric,
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

func TestEvaluateComparison(t *testing.T) {
	knd := values.MySQLValueKind

	// different lengths => error
	t.Run("different length slices returns an error", func(t *testing.T) {
		_, err := evaluateComparison([]values.SQLValue{}, []values.SQLValue{values.NewSQLInt64(knd, 1)}, sqlOpEQ, knd, nil)
		require.NotNil(t, err, "expected error")
	})

	oneTwoThree := []values.SQLValue{values.NewSQLInt64(knd, 1), values.NewSQLInt64(knd, 2), values.NewSQLInt64(knd, 3)}
	twoTwoThree := []values.SQLValue{values.NewSQLInt64(knd, 2), values.NewSQLInt64(knd, 2), values.NewSQLInt64(knd, 3)}
	oneNullThree := []values.SQLValue{values.NewSQLInt64(knd, 1), values.NewSQLNull(knd), values.NewSQLInt64(knd, 3)}

	type test struct {
		name, op    string
		left, right []values.SQLValue
		expected    values.SQLValue
	}

	tests := []test{
		{"equals true case (eq)", sqlOpEQ, oneTwoThree, oneTwoThree, values.NewSQLBool(knd, true)},
		{"equals false case (lt)", sqlOpEQ, oneTwoThree, twoTwoThree, values.NewSQLBool(knd, false)},
		{"equals false case (gt)", sqlOpEQ, twoTwoThree, oneTwoThree, values.NewSQLBool(knd, false)},
		{"equals null on right case", sqlOpEQ, oneTwoThree, oneNullThree, values.NewSQLNull(knd)},
		{"equals null on left case", sqlOpEQ, oneNullThree, oneTwoThree, values.NewSQLNull(knd)},

		{"greater than false case (eq)", sqlOpGT, oneTwoThree, oneTwoThree, values.NewSQLBool(knd, false)},
		{"greater than false case (lt)", sqlOpGT, oneTwoThree, twoTwoThree, values.NewSQLBool(knd, false)},
		{"greater than true case (gt)", sqlOpGT, twoTwoThree, oneTwoThree, values.NewSQLBool(knd, true)},
		{"greater than null on right case", sqlOpGT, oneTwoThree, oneNullThree, values.NewSQLNull(knd)},
		{"greater than null on left case", sqlOpGT, oneNullThree, oneTwoThree, values.NewSQLNull(knd)},

		{"greater than or equals true case (eq)", sqlOpGTE, oneTwoThree, oneTwoThree, values.NewSQLBool(knd, true)},
		{"greater than or equals false case (lt)", sqlOpGTE, oneTwoThree, twoTwoThree, values.NewSQLBool(knd, false)},
		{"greater than or equals true case (gt)", sqlOpGTE, twoTwoThree, oneTwoThree, values.NewSQLBool(knd, true)},
		{"greater than or equals null on right case", sqlOpGTE, oneTwoThree, oneNullThree, values.NewSQLNull(knd)},
		{"greater than or equals null on left case", sqlOpGTE, oneNullThree, oneTwoThree, values.NewSQLNull(knd)},

		{"less than false case (eq)", sqlOpLT, oneTwoThree, oneTwoThree, values.NewSQLBool(knd, false)},
		{"less than true case (lt)", sqlOpLT, oneTwoThree, twoTwoThree, values.NewSQLBool(knd, true)},
		{"less than false case (gt)", sqlOpLT, twoTwoThree, oneTwoThree, values.NewSQLBool(knd, false)},
		{"less than null on right case", sqlOpLT, oneTwoThree, oneNullThree, values.NewSQLNull(knd)},
		{"less than null on left case", sqlOpLT, oneNullThree, oneTwoThree, values.NewSQLNull(knd)},

		{"less than or equals true case (eq)", sqlOpLTE, oneTwoThree, oneTwoThree, values.NewSQLBool(knd, true)},
		{"less than or equals true case (lt)", sqlOpLTE, oneTwoThree, twoTwoThree, values.NewSQLBool(knd, true)},
		{"less than or equals false case (gt)", sqlOpLTE, twoTwoThree, oneTwoThree, values.NewSQLBool(knd, false)},
		{"less than or equals null on right case", sqlOpLTE, oneTwoThree, oneNullThree, values.NewSQLNull(knd)},
		{"less than or equals null on left case", sqlOpLTE, oneNullThree, oneTwoThree, values.NewSQLNull(knd)},

		{"not equals false case (eq)", sqlOpNEQ, oneTwoThree, oneTwoThree, values.NewSQLBool(knd, false)},
		{"not equals true case (lt)", sqlOpNEQ, oneTwoThree, twoTwoThree, values.NewSQLBool(knd, true)},
		{"not equals true case (gt)", sqlOpNEQ, twoTwoThree, oneTwoThree, values.NewSQLBool(knd, true)},
		{"not equals null on right case", sqlOpNEQ, oneTwoThree, oneNullThree, values.NewSQLNull(knd)},
		{"not equals null on left case", sqlOpNEQ, oneNullThree, oneTwoThree, values.NewSQLNull(knd)},

		{"null-safe equals true case (eq)", sqlOpNSE, oneTwoThree, oneTwoThree, values.NewSQLBool(knd, true)},
		{"null-safe equals false case (lt)", sqlOpNSE, oneTwoThree, twoTwoThree, values.NewSQLBool(knd, false)},
		{"null-safe equals false case (gt)", sqlOpNSE, twoTwoThree, oneTwoThree, values.NewSQLBool(knd, false)},
		{"null-safe equals null on right false case", sqlOpNSE, oneTwoThree, oneNullThree, values.NewSQLBool(knd, false)},
		{"null-safe equals null on left false case", sqlOpNSE, oneNullThree, oneTwoThree, values.NewSQLBool(knd, false)},
		{"null-safe equals null true case", sqlOpNSE, oneNullThree, oneNullThree, values.NewSQLBool(knd, true)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := evaluateComparison(tc.left, tc.right, tc.op, knd, nil)
			require.Nil(t, err, "unexpected error")
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestValidateArgs(t *testing.T) {
	intOne := NewSQLValueExpr(values.NewSQLInt64(values.MongoSQLValueKind, 1))
	strOne := NewSQLValueExpr(values.NewSQLVarchar(values.MongoSQLValueKind, "1"))
	intCol := results.NewColumn(1, "", "", "", "", "", "", types.EvalInt64, schema.MongoInt, false)
	strCol := results.NewColumn(1, "", "", "", "", "", "", types.EvalString, schema.MongoString, false)

	tests := []struct {
		name        string
		expr        SQLExpr
		expectError bool
	}{
		{"(SQLExpr Children) all valid args", NewSQLAddExpr(intOne, intOne), false},
		{"(SQLExpr Children) one invalid arg", NewSQLAddExpr(intOne, strOne), true},
		{"(SQLExpr Children) all invalid args", NewSQLAddExpr(strOne, strOne), true},

		{"(ProjectStage Children) SQLSubqueryExpr", NewSQLSubqueryExpr(true, true, NewDualStage()), false},
		{"(ProjectStage Children) SQLExistExpr", NewSQLSubqueryExpr(true, true, NewDualStage()), false},

		{
			"(ProjectStage Children) SQLSubqueryCmpExpr valid args",
			NewSQLSubqueryCmpExpr(true, true,
				NewProjectStage(NewDualStage(), ProjectedColumn{Column: intCol, Expr: intOne}),
				NewProjectStage(NewDualStage(), ProjectedColumn{Column: intCol, Expr: intOne}),
				sqlOpEQ,
			),
			false,
		},
		{
			"(ProjectStage Children) SQLSubqueryCmpExpr invalid args",
			NewSQLSubqueryCmpExpr(true, true,
				NewProjectStage(NewDualStage(), ProjectedColumn{Column: intCol, Expr: intOne}),
				NewProjectStage(NewDualStage(), ProjectedColumn{Column: strCol, Expr: strOne}),
				sqlOpEQ,
			),
			true,
		},
		{
			"(ProjectStage Children) NewSQLSubqueryAllExpr() valid args",
			NewSQLSubqueryAllExpr(true, true,
				NewProjectStage(NewDualStage(), ProjectedColumn{Column: intCol, Expr: intOne}),
				NewProjectStage(NewDualStage(), ProjectedColumn{Column: intCol, Expr: intOne}),
				sqlOpEQ,
			),
			false,
		},
		{
			"(ProjectStage Children) NewSQLSubqueryAllExpr() invalid args",
			NewSQLSubqueryAllExpr(true, true,
				NewProjectStage(NewDualStage(), ProjectedColumn{Column: intCol, Expr: intOne}),
				NewProjectStage(NewDualStage(), ProjectedColumn{Column: strCol, Expr: strOne}),
				sqlOpEQ,
			),
			true,
		},
		{
			"(ProjectStage Children) NewSQLSubqueryAnyExpr() valid args",
			NewSQLSubqueryAnyExpr(true, true,
				NewProjectStage(NewDualStage(), ProjectedColumn{Column: intCol, Expr: intOne}),
				NewProjectStage(NewDualStage(), ProjectedColumn{Column: intCol, Expr: intOne}),
				sqlOpEQ,
			),
			false,
		},
		{
			"(ProjectStage Children) NewSQLSubqueryAnyExpr() invalid args",
			NewSQLSubqueryAnyExpr(true, true,
				NewProjectStage(NewDualStage(), ProjectedColumn{Column: intCol, Expr: intOne}),
				NewProjectStage(NewDualStage(), ProjectedColumn{Column: strCol, Expr: strOne}),
				sqlOpEQ,
			),
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)
			err := validateArgs(test.expr)
			if test.expectError {
				req.NotNil(err, "validateArgs should return an error for these types")
			} else {
				req.Nil(err, "validateArgs should not return an error for these types")
			}
		})
	}
}

func TestReconcile(t *testing.T) {
	type test struct {
		name          string
		expr          SQLExpr
		expectedTypes []types.EvalType
	}

	runReconcileTests := func(t *testing.T, tests []test) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				req := require.New(t)

				reconciled, err := test.expr.reconcile()
				req.Nil(err, "unexpected error")

				children := nodesToExprs(reconciled.Children())

				req.Equal(len(test.expectedTypes), len(children),
					"number of expected types (%d) does not match number of actual children (%d)",
					len(test.expectedTypes), len(children),
				)

				for i, child := range children {
					req.Equal(test.expectedTypes[i], child.EvalType(), "incorrect EvalType for child %d", i)
				}
			})
		}
	}

	knd := values.MySQLValueKind

	intVal := NewSQLValueExpr(values.NewSQLInt64(knd, 1))
	uintVal := NewSQLValueExpr(values.NewSQLUint64(knd, 1))
	floatVal := NewSQLValueExpr(values.NewSQLFloat(knd, 1))
	decimalVal := NewSQLValueExpr(values.NewSQLDecimal128(knd, decimal.NewFromFloat(1.0)))
	boolVal := NewSQLValueExpr(values.NewSQLBool(knd, true))
	strVal := NewSQLValueExpr(values.NewSQLVarchar(knd, "bar"))
	dateVal := NewSQLValueExpr(values.NewSQLDate(knd, time.Now()))
	datetimeVal := NewSQLValueExpr(values.NewSQLTimestamp(knd, time.Now()))

	boolColVal := NewSQLColumnExpr(0, "", "", "", types.EvalBoolean, schema.MongoBool)

	allVals := []SQLExpr{intVal, uintVal, floatVal, decimalVal, boolVal, strVal, dateVal, datetimeVal}
	allTypes := []types.EvalType{types.EvalInt64, types.EvalUint64, types.EvalDouble, types.EvalDecimal128, types.EvalBoolean, types.EvalString, types.EvalDate, types.EvalDatetime}

	// Type order: polymorphic < bool < string < date < datetime < int < uint < double < decimal

	// tests is a list of test cases for the binary and unary operator
	// SQLExprs and the agg function SQLExprs. They have custom reconcile
	// implementations, so they are all tested individually.
	// This list does not contain tests for the scalar functions since those
	// all have their reconcile implementations generated the same way. They
	// use the convertExprs helper function which is tested separately.
	// This list also does not contain tests for the subquery exprs since
	// those all delegate their reconciliation to the reconcileSubqueryPlans
	// helper function which is tested separately.
	tests := []test{
		// arithmetic expressions: do not convert numeric types; do convert
		// non-numeric types to Decimal128 if either argument is a Datetime,
		// non-numeric types to Int64 if either argument is a Date, and
		// other non-numeric types to Float or the numeric type of the other
		// argument (if the other argument is a number).
		// The conversions are made according to the following mapping:
		// (int, int) 			=> (int, int)
		// (int, uint)			=> (int, uint)
		// (int, float) 		=> (int, float)
		// (int, decimal)		=> (int, decimal)
		// (int, bool)			=> (int, int)
		// (int, str)			=> (int, int)
		// (int, date)			=> (int, int)
		// (int, datetime)		=> (int, decimal)
		// (uint, int) 			=> (uint, int)
		// (uint, uint)			=> (uint, uint)
		// (uint, float) 		=> (uint, float)
		// (uint, decimal)		=> (uint, decimal)
		// (uint, bool)			=> (uint, uint)
		// (uint, str)			=> (uint, uint)
		// (uint, date)			=> (uint, int)
		// (uint, datetime)		=> (uint, decimal)
		// (float, int)			=> (float, int)
		// (float, uint)		=> (float, uint)
		// (float, float)		=> (float, float)
		// (float, decimal) 	=> (float, decimal)
		// (float, bool)		=> (float, float)
		// (float, str)			=> (float, float)
		// (float, date)		=> (float, int)
		// (float, datetime) 	=> (float, decimal)
		// (decimal, int)		=> (decimal, int)
		// (decimal, uint)		=> (decimal, uint)
		// (decimal, float)		=> (decimal, float)
		// (decimal, decimal)	=> (decimal, decimal)
		// (decimal, bool)		=> (decimal, decimal)
		// (decimal, str)		=> (decimal, decimal)
		// (decimal, date)		=> (decimal, int)
		// (decimal, datetime)	=> (decimal, decimal)
		// (bool, int)			=> (int, int)
		// (bool, uint)			=> (uint, uint)
		// (bool, float)		=> (float, float)
		// (bool, decimal)		=> (decimal, decimal)
		// (bool, bool)			=> (float, float)
		// (bool, str)			=> (float, float)
		// (bool, date)			=> (int, int)
		// (bool, datetime)		=> (decimal, decimal)
		// (str, int)			=> (int, int)
		// (str, uint)			=> (uint, uint)
		// (str, float)			=> (float, float)
		// (str, decimal)		=> (decimal, decimal)
		// (str, bool)			=> (float, float)
		// (str, str)			=> (float, float)
		// (str, date)			=> (int, int)
		// (str, datetime)		=> (decimal, decimal)
		// (date, int)			=> (int, int)
		// (date, uint)			=> (int, uint)
		// (date, float)		=> (int, float)
		// (date, decimal)		=> (int, decimal)
		// (date, bool)			=> (int, int)
		// (date, str)			=> (int, int)
		// (date, date)			=> (int, int)
		// (date, datetime)		=> (decimal, decimal)
		// (datetime, int)		=> (decimal, int)
		// (datetime, uint)		=> (decimal, uint)
		// (datetime, float)	=> (decimal, float)
		// (datetime, decimal)	=> (decimal, decimal)
		// (datetime, bool)		=> (decimal, decimal)
		// (datetime, str)		=> (decimal, decimal)
		// (datetime, date)		=> (decimal, decimal)
		// (datetime, datetime)	=> (decimal, decimal)

		// add.
		{"add(int,int)", NewSQLAddExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"add(int,uint)", NewSQLAddExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"add(int,float)", NewSQLAddExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"add(int,decimal)", NewSQLAddExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"add(int,bool)", NewSQLAddExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"add(int,str)", NewSQLAddExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"add(int,date)", NewSQLAddExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"add(int,datetime)", NewSQLAddExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"add(uint,int)", NewSQLAddExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"add(uint,uint)", NewSQLAddExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"add(uint,float)", NewSQLAddExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalDouble}},
		{"add(uint,decimal)", NewSQLAddExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"add(uint,bool)", NewSQLAddExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"add(uint,str)", NewSQLAddExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"add(uint,date)", NewSQLAddExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"add(uint,datetime)", NewSQLAddExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"add(float,int)", NewSQLAddExpr(floatVal, intVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"add(float,uint)", NewSQLAddExpr(floatVal, uintVal), []types.EvalType{types.EvalDouble, types.EvalUint64}},
		{"add(float,float)", NewSQLAddExpr(floatVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"add(float,decimal)", NewSQLAddExpr(floatVal, decimalVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"add(float,bool)", NewSQLAddExpr(floatVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"add(float,str)", NewSQLAddExpr(floatVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"add(float,date)", NewSQLAddExpr(floatVal, dateVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"add(float,datetime)", NewSQLAddExpr(floatVal, datetimeVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"add(decimal,int)", NewSQLAddExpr(decimalVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"add(decimal,uint)", NewSQLAddExpr(decimalVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"add(decimal,float)", NewSQLAddExpr(decimalVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"add(decimal,decimal)", NewSQLAddExpr(decimalVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"add(decimal,bool)", NewSQLAddExpr(decimalVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"add(decimal,str)", NewSQLAddExpr(decimalVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"add(decimal,date)", NewSQLAddExpr(decimalVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"add(decimal,datetime)", NewSQLAddExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"add(bool,int)", NewSQLAddExpr(boolVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"add(bool,uint)", NewSQLAddExpr(boolVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"add(bool,float)", NewSQLAddExpr(boolVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"add(bool,decimal)", NewSQLAddExpr(boolVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"add(bool,bool)", NewSQLAddExpr(boolVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"add(bool,str)", NewSQLAddExpr(boolVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"add(bool,date)", NewSQLAddExpr(boolVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"add(bool,datetime)", NewSQLAddExpr(boolVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"add(str,int)", NewSQLAddExpr(strVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"add(str,uint)", NewSQLAddExpr(strVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"add(str,float)", NewSQLAddExpr(strVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"add(str,decimal)", NewSQLAddExpr(strVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"add(str,bool)", NewSQLAddExpr(strVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"add(str,str)", NewSQLAddExpr(strVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"add(str,date)", NewSQLAddExpr(strVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"add(str,datetime)", NewSQLAddExpr(strVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"add(date,int)", NewSQLAddExpr(dateVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"add(date,uint)", NewSQLAddExpr(dateVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"add(date,float)", NewSQLAddExpr(dateVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"add(date,decimal)", NewSQLAddExpr(dateVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"add(date,bool)", NewSQLAddExpr(dateVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"add(date,str)", NewSQLAddExpr(dateVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"add(date,date)", NewSQLAddExpr(dateVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"add(date,datetime)", NewSQLAddExpr(dateVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"add(datetime,int)", NewSQLAddExpr(datetimeVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"add(datetime,uint)", NewSQLAddExpr(datetimeVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"add(datetime,float)", NewSQLAddExpr(datetimeVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"add(datetime,decimal)", NewSQLAddExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"add(datetime,bool)", NewSQLAddExpr(datetimeVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"add(datetime,str)", NewSQLAddExpr(datetimeVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"add(datetime,date)", NewSQLAddExpr(datetimeVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"add(datetime,datetime)", NewSQLAddExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},

		// divide.
		{"div(int,int)", NewSQLDivideExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"div(int,uint)", NewSQLDivideExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"div(int,float)", NewSQLDivideExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"div(int,decimal)", NewSQLDivideExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"div(int,bool)", NewSQLDivideExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"div(int,str)", NewSQLDivideExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"div(int,date)", NewSQLDivideExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"div(int,datetime)", NewSQLDivideExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"div(uint,int)", NewSQLDivideExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"div(uint,uint)", NewSQLDivideExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"div(uint,float)", NewSQLDivideExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalDouble}},
		{"div(uint,decimal)", NewSQLDivideExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"div(uint,bool)", NewSQLDivideExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"div(uint,str)", NewSQLDivideExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"div(uint,date)", NewSQLDivideExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"div(uint,datetime)", NewSQLDivideExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"div(float,int)", NewSQLDivideExpr(floatVal, intVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"div(float,uint)", NewSQLDivideExpr(floatVal, uintVal), []types.EvalType{types.EvalDouble, types.EvalUint64}},
		{"div(float,float)", NewSQLDivideExpr(floatVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"div(float,decimal)", NewSQLDivideExpr(floatVal, decimalVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"div(float,bool)", NewSQLDivideExpr(floatVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"div(float,str)", NewSQLDivideExpr(floatVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"div(float,date)", NewSQLDivideExpr(floatVal, dateVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"div(float,datetime)", NewSQLDivideExpr(floatVal, datetimeVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"div(decimal,int)", NewSQLDivideExpr(decimalVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"div(decimal,uint)", NewSQLDivideExpr(decimalVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"div(decimal,float)", NewSQLDivideExpr(decimalVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"div(decimal,decimal)", NewSQLDivideExpr(decimalVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"div(decimal,bool)", NewSQLDivideExpr(decimalVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"div(decimal,str)", NewSQLDivideExpr(decimalVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"div(decimal,date)", NewSQLDivideExpr(decimalVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"div(decimal,datetime)", NewSQLDivideExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"div(bool,int)", NewSQLDivideExpr(boolVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"div(bool,uint)", NewSQLDivideExpr(boolVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"div(bool,float)", NewSQLDivideExpr(boolVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"div(bool,decimal)", NewSQLDivideExpr(boolVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"div(bool,bool)", NewSQLDivideExpr(boolVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"div(bool,str)", NewSQLDivideExpr(boolVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"div(bool,date)", NewSQLDivideExpr(boolVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"div(bool,datetime)", NewSQLDivideExpr(boolVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"div(str,int)", NewSQLDivideExpr(strVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"div(str,uint)", NewSQLDivideExpr(strVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"div(str,float)", NewSQLDivideExpr(strVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"div(str,decimal)", NewSQLDivideExpr(strVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"div(str,bool)", NewSQLDivideExpr(strVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"div(str,str)", NewSQLDivideExpr(strVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"div(str,date)", NewSQLDivideExpr(strVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"div(str,datetime)", NewSQLDivideExpr(strVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"div(date,int)", NewSQLDivideExpr(dateVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"div(date,uint)", NewSQLDivideExpr(dateVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"div(date,float)", NewSQLDivideExpr(dateVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"div(date,decimal)", NewSQLDivideExpr(dateVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"div(date,bool)", NewSQLDivideExpr(dateVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"div(date,str)", NewSQLDivideExpr(dateVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"div(date,date)", NewSQLDivideExpr(dateVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"div(date,datetime)", NewSQLDivideExpr(dateVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"div(datetime,int)", NewSQLDivideExpr(datetimeVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"div(datetime,uint)", NewSQLDivideExpr(datetimeVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"div(datetime,float)", NewSQLDivideExpr(datetimeVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"div(datetime,decimal)", NewSQLDivideExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"div(datetime,bool)", NewSQLDivideExpr(datetimeVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"div(datetime,str)", NewSQLDivideExpr(datetimeVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"div(datetime,date)", NewSQLDivideExpr(datetimeVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"div(datetime,datetime)", NewSQLDivideExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},

		// idivide.
		{"idiv(int,int)", NewSQLIDivideExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"idiv(int,uint)", NewSQLIDivideExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"idiv(int,float)", NewSQLIDivideExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"idiv(int,decimal)", NewSQLIDivideExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"idiv(int,bool)", NewSQLIDivideExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"idiv(int,str)", NewSQLIDivideExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"idiv(int,date)", NewSQLIDivideExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"idiv(int,datetime)", NewSQLIDivideExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"idiv(uint,int)", NewSQLIDivideExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"idiv(uint,uint)", NewSQLIDivideExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"idiv(uint,float)", NewSQLIDivideExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalDouble}},
		{"idiv(uint,decimal)", NewSQLIDivideExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"idiv(uint,bool)", NewSQLIDivideExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"idiv(uint,str)", NewSQLIDivideExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"idiv(uint,date)", NewSQLIDivideExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"idiv(uint,datetime)", NewSQLIDivideExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"idiv(float,int)", NewSQLIDivideExpr(floatVal, intVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"idiv(float,uint)", NewSQLIDivideExpr(floatVal, uintVal), []types.EvalType{types.EvalDouble, types.EvalUint64}},
		{"idiv(float,float)", NewSQLIDivideExpr(floatVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"idiv(float,decimal)", NewSQLIDivideExpr(floatVal, decimalVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"idiv(float,bool)", NewSQLIDivideExpr(floatVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"idiv(float,str)", NewSQLIDivideExpr(floatVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"idiv(float,date)", NewSQLIDivideExpr(floatVal, dateVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"idiv(float,datetime)", NewSQLIDivideExpr(floatVal, datetimeVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"idiv(decimal,int)", NewSQLIDivideExpr(decimalVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"idiv(decimal,uint)", NewSQLIDivideExpr(decimalVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"idiv(decimal,float)", NewSQLIDivideExpr(decimalVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"idiv(decimal,decimal)", NewSQLIDivideExpr(decimalVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"idiv(decimal,bool)", NewSQLIDivideExpr(decimalVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"idiv(decimal,str)", NewSQLIDivideExpr(decimalVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"idiv(decimal,date)", NewSQLIDivideExpr(decimalVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"idiv(decimal,datetime)", NewSQLIDivideExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"idiv(bool,int)", NewSQLIDivideExpr(boolVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"idiv(bool,uint)", NewSQLIDivideExpr(boolVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"idiv(bool,float)", NewSQLIDivideExpr(boolVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"idiv(bool,decimal)", NewSQLIDivideExpr(boolVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"idiv(bool,bool)", NewSQLIDivideExpr(boolVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"idiv(bool,str)", NewSQLIDivideExpr(boolVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"idiv(bool,date)", NewSQLIDivideExpr(boolVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"idiv(bool,datetime)", NewSQLIDivideExpr(boolVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"idiv(str,int)", NewSQLIDivideExpr(strVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"idiv(str,uint)", NewSQLIDivideExpr(strVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"idiv(str,float)", NewSQLIDivideExpr(strVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"idiv(str,decimal)", NewSQLIDivideExpr(strVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"idiv(str,bool)", NewSQLIDivideExpr(strVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"idiv(str,str)", NewSQLIDivideExpr(strVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"idiv(str,date)", NewSQLIDivideExpr(strVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"idiv(str,datetime)", NewSQLIDivideExpr(strVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"idiv(date,int)", NewSQLIDivideExpr(dateVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"idiv(date,uint)", NewSQLIDivideExpr(dateVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"idiv(date,float)", NewSQLIDivideExpr(dateVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"idiv(date,decimal)", NewSQLIDivideExpr(dateVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"idiv(date,bool)", NewSQLIDivideExpr(dateVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"idiv(date,str)", NewSQLIDivideExpr(dateVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"idiv(date,date)", NewSQLIDivideExpr(dateVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"idiv(date,datetime)", NewSQLIDivideExpr(dateVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"idiv(datetime,int)", NewSQLIDivideExpr(datetimeVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"idiv(datetime,uint)", NewSQLIDivideExpr(datetimeVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"idiv(datetime,float)", NewSQLIDivideExpr(datetimeVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"idiv(datetime,decimal)", NewSQLIDivideExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"idiv(datetime,bool)", NewSQLIDivideExpr(datetimeVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"idiv(datetime,str)", NewSQLIDivideExpr(datetimeVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"idiv(datetime,date)", NewSQLIDivideExpr(datetimeVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"idiv(datetime,datetime)", NewSQLIDivideExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},

		// mod.
		{"mod(int,int)", NewSQLModExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mod(int,uint)", NewSQLModExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"mod(int,float)", NewSQLModExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"mod(int,decimal)", NewSQLModExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"mod(int,bool)", NewSQLModExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mod(int,str)", NewSQLModExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mod(int,date)", NewSQLModExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mod(int,datetime)", NewSQLModExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"mod(uint,int)", NewSQLModExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"mod(uint,uint)", NewSQLModExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"mod(uint,float)", NewSQLModExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalDouble}},
		{"mod(uint,decimal)", NewSQLModExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"mod(uint,bool)", NewSQLModExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"mod(uint,str)", NewSQLModExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"mod(uint,date)", NewSQLModExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"mod(uint,datetime)", NewSQLModExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"mod(float,int)", NewSQLModExpr(floatVal, intVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"mod(float,uint)", NewSQLModExpr(floatVal, uintVal), []types.EvalType{types.EvalDouble, types.EvalUint64}},
		{"mod(float,float)", NewSQLModExpr(floatVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mod(float,decimal)", NewSQLModExpr(floatVal, decimalVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"mod(float,bool)", NewSQLModExpr(floatVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mod(float,str)", NewSQLModExpr(floatVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mod(float,date)", NewSQLModExpr(floatVal, dateVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"mod(float,datetime)", NewSQLModExpr(floatVal, datetimeVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"mod(decimal,int)", NewSQLModExpr(decimalVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"mod(decimal,uint)", NewSQLModExpr(decimalVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"mod(decimal,float)", NewSQLModExpr(decimalVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"mod(decimal,decimal)", NewSQLModExpr(decimalVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mod(decimal,bool)", NewSQLModExpr(decimalVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mod(decimal,str)", NewSQLModExpr(decimalVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mod(decimal,date)", NewSQLModExpr(decimalVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"mod(decimal,datetime)", NewSQLModExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mod(bool,int)", NewSQLModExpr(boolVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mod(bool,uint)", NewSQLModExpr(boolVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"mod(bool,float)", NewSQLModExpr(boolVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mod(bool,decimal)", NewSQLModExpr(boolVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mod(bool,bool)", NewSQLModExpr(boolVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mod(bool,str)", NewSQLModExpr(boolVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mod(bool,date)", NewSQLModExpr(boolVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mod(bool,datetime)", NewSQLModExpr(boolVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mod(str,int)", NewSQLModExpr(strVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mod(str,uint)", NewSQLModExpr(strVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"mod(str,float)", NewSQLModExpr(strVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mod(str,decimal)", NewSQLModExpr(strVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mod(str,bool)", NewSQLModExpr(strVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mod(str,str)", NewSQLModExpr(strVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mod(str,date)", NewSQLModExpr(strVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mod(str,datetime)", NewSQLModExpr(strVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mod(date,int)", NewSQLModExpr(dateVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mod(date,uint)", NewSQLModExpr(dateVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"mod(date,float)", NewSQLModExpr(dateVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"mod(date,decimal)", NewSQLModExpr(dateVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"mod(date,bool)", NewSQLModExpr(dateVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mod(date,str)", NewSQLModExpr(dateVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mod(date,date)", NewSQLModExpr(dateVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mod(date,datetime)", NewSQLModExpr(dateVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mod(datetime,int)", NewSQLModExpr(datetimeVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"mod(datetime,uint)", NewSQLModExpr(datetimeVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"mod(datetime,float)", NewSQLModExpr(datetimeVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"mod(datetime,decimal)", NewSQLModExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mod(datetime,bool)", NewSQLModExpr(datetimeVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mod(datetime,str)", NewSQLModExpr(datetimeVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mod(datetime,date)", NewSQLModExpr(datetimeVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mod(datetime,datetime)", NewSQLModExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},

		// multiply.
		{"mult(int,int)", NewSQLMultiplyExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mult(int,uint)", NewSQLMultiplyExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"mult(int,float)", NewSQLMultiplyExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"mult(int,decimal)", NewSQLMultiplyExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"mult(int,bool)", NewSQLMultiplyExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mult(int,str)", NewSQLMultiplyExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mult(int,date)", NewSQLMultiplyExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mult(int,datetime)", NewSQLMultiplyExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"mult(uint,int)", NewSQLMultiplyExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"mult(uint,uint)", NewSQLMultiplyExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"mult(uint,float)", NewSQLMultiplyExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalDouble}},
		{"mult(uint,decimal)", NewSQLMultiplyExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"mult(uint,bool)", NewSQLMultiplyExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"mult(uint,str)", NewSQLMultiplyExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"mult(uint,date)", NewSQLMultiplyExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"mult(uint,datetime)", NewSQLMultiplyExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"mult(float,int)", NewSQLMultiplyExpr(floatVal, intVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"mult(float,uint)", NewSQLMultiplyExpr(floatVal, uintVal), []types.EvalType{types.EvalDouble, types.EvalUint64}},
		{"mult(float,float)", NewSQLMultiplyExpr(floatVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mult(float,decimal)", NewSQLMultiplyExpr(floatVal, decimalVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"mult(float,bool)", NewSQLMultiplyExpr(floatVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mult(float,str)", NewSQLMultiplyExpr(floatVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mult(float,date)", NewSQLMultiplyExpr(floatVal, dateVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"mult(float,datetime)", NewSQLMultiplyExpr(floatVal, datetimeVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"mult(decimal,int)", NewSQLMultiplyExpr(decimalVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"mult(decimal,uint)", NewSQLMultiplyExpr(decimalVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"mult(decimal,float)", NewSQLMultiplyExpr(decimalVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"mult(decimal,decimal)", NewSQLMultiplyExpr(decimalVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mult(decimal,bool)", NewSQLMultiplyExpr(decimalVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mult(decimal,str)", NewSQLMultiplyExpr(decimalVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mult(decimal,date)", NewSQLMultiplyExpr(decimalVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"mult(decimal,datetime)", NewSQLMultiplyExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mult(bool,int)", NewSQLMultiplyExpr(boolVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mult(bool,uint)", NewSQLMultiplyExpr(boolVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"mult(bool,float)", NewSQLMultiplyExpr(boolVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mult(bool,decimal)", NewSQLMultiplyExpr(boolVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mult(bool,bool)", NewSQLMultiplyExpr(boolVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mult(bool,str)", NewSQLMultiplyExpr(boolVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mult(bool,date)", NewSQLMultiplyExpr(boolVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mult(bool,datetime)", NewSQLMultiplyExpr(boolVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mult(str,int)", NewSQLMultiplyExpr(strVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mult(str,uint)", NewSQLMultiplyExpr(strVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"mult(str,float)", NewSQLMultiplyExpr(strVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mult(str,decimal)", NewSQLMultiplyExpr(strVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mult(str,bool)", NewSQLMultiplyExpr(strVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mult(str,str)", NewSQLMultiplyExpr(strVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"mult(str,date)", NewSQLMultiplyExpr(strVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mult(str,datetime)", NewSQLMultiplyExpr(strVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mult(date,int)", NewSQLMultiplyExpr(dateVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mult(date,uint)", NewSQLMultiplyExpr(dateVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"mult(date,float)", NewSQLMultiplyExpr(dateVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"mult(date,decimal)", NewSQLMultiplyExpr(dateVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"mult(date,bool)", NewSQLMultiplyExpr(dateVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mult(date,str)", NewSQLMultiplyExpr(dateVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mult(date,date)", NewSQLMultiplyExpr(dateVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"mult(date,datetime)", NewSQLMultiplyExpr(dateVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mult(datetime,int)", NewSQLMultiplyExpr(datetimeVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"mult(datetime,uint)", NewSQLMultiplyExpr(datetimeVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"mult(datetime,float)", NewSQLMultiplyExpr(datetimeVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"mult(datetime,decimal)", NewSQLMultiplyExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mult(datetime,bool)", NewSQLMultiplyExpr(datetimeVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mult(datetime,str)", NewSQLMultiplyExpr(datetimeVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mult(datetime,date)", NewSQLMultiplyExpr(datetimeVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"mult(datetime,datetime)", NewSQLMultiplyExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},

		// subtract.
		{"sub(int,int)", NewSQLSubtractExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sub(int,uint)", NewSQLSubtractExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"sub(int,float)", NewSQLSubtractExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"sub(int,decimal)", NewSQLSubtractExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"sub(int,bool)", NewSQLSubtractExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sub(int,str)", NewSQLSubtractExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sub(int,date)", NewSQLSubtractExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sub(int,datetime)", NewSQLSubtractExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"sub(uint,int)", NewSQLSubtractExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"sub(uint,uint)", NewSQLSubtractExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sub(uint,float)", NewSQLSubtractExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalDouble}},
		{"sub(uint,decimal)", NewSQLSubtractExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"sub(uint,bool)", NewSQLSubtractExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sub(uint,str)", NewSQLSubtractExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sub(uint,date)", NewSQLSubtractExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"sub(uint,datetime)", NewSQLSubtractExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"sub(float,int)", NewSQLSubtractExpr(floatVal, intVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"sub(float,uint)", NewSQLSubtractExpr(floatVal, uintVal), []types.EvalType{types.EvalDouble, types.EvalUint64}},
		{"sub(float,float)", NewSQLSubtractExpr(floatVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sub(float,decimal)", NewSQLSubtractExpr(floatVal, decimalVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"sub(float,bool)", NewSQLSubtractExpr(floatVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sub(float,str)", NewSQLSubtractExpr(floatVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sub(float,date)", NewSQLSubtractExpr(floatVal, dateVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"sub(float,datetime)", NewSQLSubtractExpr(floatVal, datetimeVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"sub(decimal,int)", NewSQLSubtractExpr(decimalVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"sub(decimal,uint)", NewSQLSubtractExpr(decimalVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"sub(decimal,float)", NewSQLSubtractExpr(decimalVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"sub(decimal,decimal)", NewSQLSubtractExpr(decimalVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sub(decimal,bool)", NewSQLSubtractExpr(decimalVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sub(decimal,str)", NewSQLSubtractExpr(decimalVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sub(decimal,date)", NewSQLSubtractExpr(decimalVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"sub(decimal,datetime)", NewSQLSubtractExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sub(bool,int)", NewSQLSubtractExpr(boolVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sub(bool,uint)", NewSQLSubtractExpr(boolVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sub(bool,float)", NewSQLSubtractExpr(boolVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sub(bool,decimal)", NewSQLSubtractExpr(boolVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sub(bool,bool)", NewSQLSubtractExpr(boolVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sub(bool,str)", NewSQLSubtractExpr(boolVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sub(bool,date)", NewSQLSubtractExpr(boolVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sub(bool,datetime)", NewSQLSubtractExpr(boolVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sub(str,int)", NewSQLSubtractExpr(strVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sub(str,uint)", NewSQLSubtractExpr(strVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sub(str,float)", NewSQLSubtractExpr(strVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sub(str,decimal)", NewSQLSubtractExpr(strVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sub(str,bool)", NewSQLSubtractExpr(strVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sub(str,str)", NewSQLSubtractExpr(strVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sub(str,date)", NewSQLSubtractExpr(strVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sub(str,datetime)", NewSQLSubtractExpr(strVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sub(date,int)", NewSQLSubtractExpr(dateVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sub(date,uint)", NewSQLSubtractExpr(dateVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"sub(date,float)", NewSQLSubtractExpr(dateVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"sub(date,decimal)", NewSQLSubtractExpr(dateVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"sub(date,bool)", NewSQLSubtractExpr(dateVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sub(date,str)", NewSQLSubtractExpr(dateVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sub(date,date)", NewSQLSubtractExpr(dateVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sub(date,datetime)", NewSQLSubtractExpr(dateVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sub(datetime,int)", NewSQLSubtractExpr(datetimeVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"sub(datetime,uint)", NewSQLSubtractExpr(datetimeVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"sub(datetime,float)", NewSQLSubtractExpr(datetimeVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"sub(datetime,decimal)", NewSQLSubtractExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sub(datetime,bool)", NewSQLSubtractExpr(datetimeVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sub(datetime,str)", NewSQLSubtractExpr(datetimeVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sub(datetime,date)", NewSQLSubtractExpr(datetimeVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sub(datetime,datetime)", NewSQLSubtractExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},

		// logical expressions: do not convert boolean comparable types
		// (bool, int, uint); do convert other types to bool.
		// The conversions are made according to the following mapping:
		// (int, int) 			=> (int, int)
		// (int, uint)			=> (int, uint)
		// (int, float) 		=> (int, bool)
		// (int, decimal)		=> (int, bool)
		// (int, bool)			=> (int, bool)
		// (int, str)			=> (int, bool)
		// (int, date)			=> (int, bool)
		// (int, datetime)		=> (int, bool)
		// (uint, int) 			=> (uint, int)
		// (uint, uint)			=> (uint, uint)
		// (uint, float) 		=> (uint, bool)
		// (uint, decimal)		=> (uint, bool)
		// (uint, bool)			=> (uint, bool)
		// (uint, str)			=> (uint, bool)
		// (uint, date)			=> (uint, bool)
		// (uint, datetime)		=> (uint, bool)
		// (float, int)			=> (bool, int)
		// (float, uint)		=> (bool, uint)
		// (float, float)		=> (bool, bool)
		// (float, decimal) 	=> (bool, bool)
		// (float, bool)		=> (bool, bool)
		// (float, str)			=> (bool, bool)
		// (float, date)		=> (bool, bool)
		// (float, datetime) 	=> (bool, bool)
		// (decimal, int)		=> (bool, int)
		// (decimal, uint)		=> (bool, uint)
		// (decimal, float)		=> (bool, bool)
		// (decimal, decimal)	=> (bool, bool)
		// (decimal, bool)		=> (bool, bool)
		// (decimal, str)		=> (bool, bool)
		// (decimal, date)		=> (bool, bool)
		// (decimal, datetime)	=> (bool, bool)
		// (bool, int)			=> (bool, int)
		// (bool, uint)			=> (bool, uint)
		// (bool, float)		=> (bool, bool)
		// (bool, decimal)		=> (bool, bool)
		// (bool, bool)			=> (bool, bool)
		// (bool, str)			=> (bool, bool)
		// (bool, date)			=> (bool, bool)
		// (bool, datetime)		=> (bool, bool)
		// (str, int)			=> (bool, int)
		// (str, uint)			=> (bool, uint)
		// (str, float)			=> (bool, bool)
		// (str, decimal)		=> (bool, bool)
		// (str, bool)			=> (bool, bool)
		// (str, str)			=> (bool, bool)
		// (str, date)			=> (bool, bool)
		// (str, datetime)		=> (bool, bool)
		// (date, int)			=> (bool, int)
		// (date, uint)			=> (bool, uint)
		// (date, float)		=> (bool, bool)
		// (date, decimal)		=> (bool, bool)
		// (date, bool)			=> (bool, bool)
		// (date, str)			=> (bool, bool)
		// (date, date)			=> (bool, bool)
		// (date, datetime)		=> (bool, bool)
		// (datetime, int)		=> (bool, int)
		// (datetime, uint)		=> (bool, uint)
		// (datetime, float)	=> (bool, bool)
		// (datetime, decimal)	=> (bool, bool)
		// (datetime, bool)		=> (bool, bool)
		// (datetime, str)		=> (bool, bool)
		// (datetime, date)		=> (bool, bool)
		// (datetime, datetime)	=> (bool, bool)

		// and.
		{"and(int,int)", NewSQLAndExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"and(int,uint)", NewSQLAndExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"and(int,float)", NewSQLAndExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"and(int,decimal)", NewSQLAndExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"and(int,bool)", NewSQLAndExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"and(int,str)", NewSQLAndExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"and(int,date)", NewSQLAndExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"and(int,datetime)", NewSQLAndExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"and(uint,int)", NewSQLAndExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"and(uint,uint)", NewSQLAndExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"and(uint,float)", NewSQLAndExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"and(uint,decimal)", NewSQLAndExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"and(uint,bool)", NewSQLAndExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"and(uint,str)", NewSQLAndExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"and(uint,date)", NewSQLAndExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"and(uint,datetime)", NewSQLAndExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"and(float,int)", NewSQLAndExpr(floatVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"and(float,uint)", NewSQLAndExpr(floatVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"and(float,float)", NewSQLAndExpr(floatVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(float,decimal)", NewSQLAndExpr(floatVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(float,bool)", NewSQLAndExpr(floatVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(float,str)", NewSQLAndExpr(floatVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(float,date)", NewSQLAndExpr(floatVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(float,datetime)", NewSQLAndExpr(floatVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(decimal,int)", NewSQLAndExpr(decimalVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"and(decimal,uint)", NewSQLAndExpr(decimalVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"and(decimal,float)", NewSQLAndExpr(decimalVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(decimal,decimal)", NewSQLAndExpr(decimalVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(decimal,bool)", NewSQLAndExpr(decimalVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(decimal,str)", NewSQLAndExpr(decimalVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(decimal,date)", NewSQLAndExpr(decimalVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(decimal,datetime)", NewSQLAndExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(bool,int)", NewSQLAndExpr(boolVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"and(bool,uint)", NewSQLAndExpr(boolVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"and(bool,float)", NewSQLAndExpr(boolVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(bool,decimal)", NewSQLAndExpr(boolVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(bool,bool)", NewSQLAndExpr(boolVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(bool,str)", NewSQLAndExpr(boolVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(bool,date)", NewSQLAndExpr(boolVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(bool,datetime)", NewSQLAndExpr(boolVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(str,int)", NewSQLAndExpr(strVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"and(str,uint)", NewSQLAndExpr(strVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"and(str,float)", NewSQLAndExpr(strVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(str,decimal)", NewSQLAndExpr(strVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(str,bool)", NewSQLAndExpr(strVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(str,str)", NewSQLAndExpr(strVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(str,date)", NewSQLAndExpr(strVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(str,datetime)", NewSQLAndExpr(strVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(date,int)", NewSQLAndExpr(dateVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"and(date,uint)", NewSQLAndExpr(dateVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"and(date,float)", NewSQLAndExpr(dateVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(date,decimal)", NewSQLAndExpr(dateVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(date,bool)", NewSQLAndExpr(dateVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(date,str)", NewSQLAndExpr(dateVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(date,date)", NewSQLAndExpr(dateVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(date,datetime)", NewSQLAndExpr(dateVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(datetime,int)", NewSQLAndExpr(datetimeVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"and(datetime,uint)", NewSQLAndExpr(datetimeVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"and(datetime,float)", NewSQLAndExpr(datetimeVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(datetime,decimal)", NewSQLAndExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(datetime,bool)", NewSQLAndExpr(datetimeVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(datetime,str)", NewSQLAndExpr(datetimeVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(datetime,date)", NewSQLAndExpr(datetimeVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"and(datetime,datetime)", NewSQLAndExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},

		// or.
		{"or(int,int)", NewSQLOrExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"or(int,uint)", NewSQLOrExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"or(int,float)", NewSQLOrExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"or(int,decimal)", NewSQLOrExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"or(int,bool)", NewSQLOrExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"or(int,str)", NewSQLOrExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"or(int,date)", NewSQLOrExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"or(int,datetime)", NewSQLOrExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"or(uint,int)", NewSQLOrExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"or(uint,uint)", NewSQLOrExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"or(uint,float)", NewSQLOrExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"or(uint,decimal)", NewSQLOrExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"or(uint,bool)", NewSQLOrExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"or(uint,str)", NewSQLOrExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"or(uint,date)", NewSQLOrExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"or(uint,datetime)", NewSQLOrExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"or(float,int)", NewSQLOrExpr(floatVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"or(float,uint)", NewSQLOrExpr(floatVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"or(float,float)", NewSQLOrExpr(floatVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(float,decimal)", NewSQLOrExpr(floatVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(float,bool)", NewSQLOrExpr(floatVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(float,str)", NewSQLOrExpr(floatVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(float,date)", NewSQLOrExpr(floatVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(float,datetime)", NewSQLOrExpr(floatVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(decimal,int)", NewSQLOrExpr(decimalVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"or(decimal,uint)", NewSQLOrExpr(decimalVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"or(decimal,float)", NewSQLOrExpr(decimalVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(decimal,decimal)", NewSQLOrExpr(decimalVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(decimal,bool)", NewSQLOrExpr(decimalVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(decimal,str)", NewSQLOrExpr(decimalVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(decimal,date)", NewSQLOrExpr(decimalVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(decimal,datetime)", NewSQLOrExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(bool,int)", NewSQLOrExpr(boolVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"or(bool,uint)", NewSQLOrExpr(boolVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"or(bool,float)", NewSQLOrExpr(boolVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(bool,decimal)", NewSQLOrExpr(boolVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(bool,bool)", NewSQLOrExpr(boolVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(bool,str)", NewSQLOrExpr(boolVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(bool,date)", NewSQLOrExpr(boolVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(bool,datetime)", NewSQLOrExpr(boolVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(str,int)", NewSQLOrExpr(strVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"or(str,uint)", NewSQLOrExpr(strVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"or(str,float)", NewSQLOrExpr(strVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(str,decimal)", NewSQLOrExpr(strVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(str,bool)", NewSQLOrExpr(strVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(str,str)", NewSQLOrExpr(strVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(str,date)", NewSQLOrExpr(strVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(str,datetime)", NewSQLOrExpr(strVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(date,int)", NewSQLOrExpr(dateVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"or(date,uint)", NewSQLOrExpr(dateVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"or(date,float)", NewSQLOrExpr(dateVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(date,decimal)", NewSQLOrExpr(dateVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(date,bool)", NewSQLOrExpr(dateVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(date,str)", NewSQLOrExpr(dateVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(date,date)", NewSQLOrExpr(dateVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(date,datetime)", NewSQLOrExpr(dateVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(datetime,int)", NewSQLOrExpr(datetimeVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"or(datetime,uint)", NewSQLOrExpr(datetimeVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"or(datetime,float)", NewSQLOrExpr(datetimeVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(datetime,decimal)", NewSQLOrExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(datetime,bool)", NewSQLOrExpr(datetimeVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(datetime,str)", NewSQLOrExpr(datetimeVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(datetime,date)", NewSQLOrExpr(datetimeVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"or(datetime,datetime)", NewSQLOrExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},

		// xor.
		{"xor(NewSQLXorExpr(),int)", NewSQLXorExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"xor(int,uint)", NewSQLXorExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"xor(int,float)", NewSQLXorExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"xor(int,decimal)", NewSQLXorExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"xor(int,bool)", NewSQLXorExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"xor(int,str)", NewSQLXorExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"xor(int,date)", NewSQLXorExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"xor(int,datetime)", NewSQLXorExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"xor(uint,int)", NewSQLXorExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"xor(uint,uint)", NewSQLXorExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"xor(uint,float)", NewSQLXorExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"xor(uint,decimal)", NewSQLXorExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"xor(uint,bool)", NewSQLXorExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"xor(uint,str)", NewSQLXorExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"xor(uint,date)", NewSQLXorExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"xor(uint,datetime)", NewSQLXorExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"xor(float,int)", NewSQLXorExpr(floatVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"xor(float,uint)", NewSQLXorExpr(floatVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"xor(float,float)", NewSQLXorExpr(floatVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(float,decimal)", NewSQLXorExpr(floatVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(float,bool)", NewSQLXorExpr(floatVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(float,str)", NewSQLXorExpr(floatVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(float,date)", NewSQLXorExpr(floatVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(float,datetime)", NewSQLXorExpr(floatVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(decimal,int)", NewSQLXorExpr(decimalVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"xor(decimal,uint)", NewSQLXorExpr(decimalVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"xor(decimal,float)", NewSQLXorExpr(decimalVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(decimal,decimal)", NewSQLXorExpr(decimalVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(decimal,bool)", NewSQLXorExpr(decimalVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(decimal,str)", NewSQLXorExpr(decimalVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(decimal,date)", NewSQLXorExpr(decimalVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(decimal,datetime)", NewSQLXorExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(bool,int)", NewSQLXorExpr(boolVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"xor(bool,uint)", NewSQLXorExpr(boolVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"xor(bool,float)", NewSQLXorExpr(boolVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(bool,decimal)", NewSQLXorExpr(boolVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(bool,bool)", NewSQLXorExpr(boolVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(bool,str)", NewSQLXorExpr(boolVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(bool,date)", NewSQLXorExpr(boolVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(bool,datetime)", NewSQLXorExpr(boolVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(str,int)", NewSQLXorExpr(strVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"xor(str,uint)", NewSQLXorExpr(strVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"xor(str,float)", NewSQLXorExpr(strVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(str,decimal)", NewSQLXorExpr(strVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(str,bool)", NewSQLXorExpr(strVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(str,str)", NewSQLXorExpr(strVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(str,date)", NewSQLXorExpr(strVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(str,datetime)", NewSQLXorExpr(strVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(date,int)", NewSQLXorExpr(dateVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"xor(date,uint)", NewSQLXorExpr(dateVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"xor(date,float)", NewSQLXorExpr(dateVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(date,decimal)", NewSQLXorExpr(dateVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(date,bool)", NewSQLXorExpr(dateVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(date,str)", NewSQLXorExpr(dateVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(date,date)", NewSQLXorExpr(dateVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(date,datetime)", NewSQLXorExpr(dateVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(datetime,int)", NewSQLXorExpr(datetimeVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalInt64}},
		{"xor(datetime,uint)", NewSQLXorExpr(datetimeVal, uintVal), []types.EvalType{types.EvalBoolean, types.EvalUint64}},
		{"xor(datetime,float)", NewSQLXorExpr(datetimeVal, floatVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(datetime,decimal)", NewSQLXorExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(datetime,bool)", NewSQLXorExpr(datetimeVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(datetime,str)", NewSQLXorExpr(datetimeVal, strVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(datetime,date)", NewSQLXorExpr(datetimeVal, dateVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"xor(datetime,datetime)", NewSQLXorExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},

		// comparison expressions: do not convert types if they are similar;
		// do convert types that are not similar to the higher precedence
		// type.
		// The conversions are made according to the following mapping:
		// (int, int) 			=> (int, int)
		// (int, uint)			=> (int, uint)
		// (int, float) 		=> (int, float)
		// (int, decimal)		=> (int, decimal)
		// (int, bool)			=> (int, int)
		// (int, str)			=> (int, int)
		// (int, date)			=> (int, int)
		// (int, datetime)		=> (int, int)
		// (uint, int) 			=> (uint, int)
		// (uint, uint)			=> (uint, uint)
		// (uint, float) 		=> (uint, float)
		// (uint, decimal)		=> (uint, decimal)
		// (uint, bool)			=> (uint, uint)
		// (uint, str)			=> (uint, uint)
		// (uint, date)			=> (uint, uint)
		// (uint, datetime)		=> (uint, uint)
		// (float, int)			=> (float, int)
		// (float, uint)		=> (float, uint)
		// (float, float)		=> (float, float)
		// (float, decimal) 	=> (float, decimal)
		// (float, bool)		=> (float, float)
		// (float, str)			=> (float, float)
		// (float, date)		=> (float, float)
		// (float, datetime) 	=> (float, float)
		// (decimal, int)		=> (decimal, int)
		// (decimal, uint)		=> (decimal, uint)
		// (decimal, float)		=> (decimal, float)
		// (decimal, decimal)	=> (decimal, decimal)
		// (decimal, bool)		=> (decimal, decimal)
		// (decimal, str)		=> (decimal, decimal)
		// (decimal, date)		=> (decimal, decimal)
		// (decimal, datetime)	=> (decimal, decimal)
		// (bool, int)			=> (int, int)
		// (bool, uint)			=> (uint, uint)
		// (bool, float)		=> (float, float)
		// (bool, decimal)		=> (decimal, decimal)
		// (bool, bool)			=> (bool, bool)
		// (bool, str)			=> (str, str)
		// (bool, date)			=> (date, date)
		// (bool, datetime)		=> (datetime, datetime)
		// (str, int)			=> (int, int)
		// (str, uint)			=> (uint, uint)
		// (str, float)			=> (float, float)
		// (str, decimal)		=> (decimal, decimal)
		// (str, bool)			=> (str, str)
		// (str, str)			=> (str, str)
		// (str, date)			=> (date, date)
		// (str, datetime)		=> (datetime, datetime)
		// (date, int)			=> (int, int)
		// (date, uint)			=> (uint, uint)
		// (date, float)		=> (float, float)
		// (date, decimal)		=> (decimal, decimal)
		// (date, bool)			=> (date, date)
		// (date, str)			=> (date, date)
		// (date, date)			=> (date, date)
		// (date, datetime)		=> (datetime, datetime)
		// (datetime, int)		=> (int, int)
		// (datetime, uint)		=> (uint, uint)
		// (datetime, float)	=> (float, float)
		// (datetime, decimal)	=> (decimal, decimal)
		// (datetime, bool)		=> (datetime, datetime)
		// (datetime, str)		=> (datetime, datetime)
		// (datetime, date)		=> (datetime, datetime)
		// (datetime, datetime)	=> (datetime, datetime)

		// equal.
		{"eq(int,int)", NewSQLEqualsExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"eq(int,uint)", NewSQLEqualsExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"eq(int,float)", NewSQLEqualsExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"eq(int,decimal)", NewSQLEqualsExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"eq(int,bool)", NewSQLEqualsExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"eq(int,str)", NewSQLEqualsExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"eq(int,date)", NewSQLEqualsExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"eq(int,datetime)", NewSQLEqualsExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"eq(uint,int)", NewSQLEqualsExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"eq(uint,uint)", NewSQLEqualsExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"eq(uint,float)", NewSQLEqualsExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalDouble}},
		{"eq(uint,decimal)", NewSQLEqualsExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"eq(uint,bool)", NewSQLEqualsExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"eq(uint,str)", NewSQLEqualsExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"eq(uint,date)", NewSQLEqualsExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"eq(uint,datetime)", NewSQLEqualsExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"eq(float,int)", NewSQLEqualsExpr(floatVal, intVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"eq(float,uint)", NewSQLEqualsExpr(floatVal, uintVal), []types.EvalType{types.EvalDouble, types.EvalUint64}},
		{"eq(float,float)", NewSQLEqualsExpr(floatVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"eq(float,decimal)", NewSQLEqualsExpr(floatVal, decimalVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"eq(float,bool)", NewSQLEqualsExpr(floatVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"eq(float,str)", NewSQLEqualsExpr(floatVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"eq(float,date)", NewSQLEqualsExpr(floatVal, dateVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"eq(float,datetime)", NewSQLEqualsExpr(floatVal, datetimeVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"eq(decimal,int)", NewSQLEqualsExpr(decimalVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"eq(decimal,uint)", NewSQLEqualsExpr(decimalVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"eq(decimal,float)", NewSQLEqualsExpr(decimalVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"eq(decimal,decimal)", NewSQLEqualsExpr(decimalVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"eq(decimal,bool)", NewSQLEqualsExpr(decimalVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"eq(decimal,str)", NewSQLEqualsExpr(decimalVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"eq(decimal,date)", NewSQLEqualsExpr(decimalVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"eq(decimal,datetime)", NewSQLEqualsExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"eq(bool,int)", NewSQLEqualsExpr(boolVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"eq(bool,uint)", NewSQLEqualsExpr(boolVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"eq(bool,float)", NewSQLEqualsExpr(boolVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"eq(bool,decimal)", NewSQLEqualsExpr(boolVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"eq(bool,bool)", NewSQLEqualsExpr(boolVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"eq(bool,str)", NewSQLEqualsExpr(boolVal, strVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"eq(bool,date)", NewSQLEqualsExpr(boolVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"eq(bool,datetime)", NewSQLEqualsExpr(boolVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"eq(str,int)", NewSQLEqualsExpr(strVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"eq(str,uint)", NewSQLEqualsExpr(strVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"eq(str,float)", NewSQLEqualsExpr(strVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"eq(str,decimal)", NewSQLEqualsExpr(strVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"eq(str,bool)", NewSQLEqualsExpr(strVal, boolVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"eq(str,str)", NewSQLEqualsExpr(strVal, strVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"eq(str,date)", NewSQLEqualsExpr(strVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"eq(str,datetime)", NewSQLEqualsExpr(strVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"eq(date,int)", NewSQLEqualsExpr(dateVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"eq(date,uint)", NewSQLEqualsExpr(dateVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"eq(date,float)", NewSQLEqualsExpr(dateVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"eq(date,decimal)", NewSQLEqualsExpr(dateVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"eq(date,bool)", NewSQLEqualsExpr(dateVal, boolVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"eq(date,str)", NewSQLEqualsExpr(dateVal, strVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"eq(date,date)", NewSQLEqualsExpr(dateVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"eq(date,datetime)", NewSQLEqualsExpr(dateVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"eq(datetime,int)", NewSQLEqualsExpr(datetimeVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"eq(datetime,uint)", NewSQLEqualsExpr(datetimeVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"eq(datetime,float)", NewSQLEqualsExpr(datetimeVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"eq(datetime,decimal)", NewSQLEqualsExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"eq(datetime,bool)", NewSQLEqualsExpr(datetimeVal, boolVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"eq(datetime,str)", NewSQLEqualsExpr(datetimeVal, strVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"eq(datetime,date)", NewSQLEqualsExpr(datetimeVal, dateVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"eq(datetime,datetime)", NewSQLEqualsExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},

		// equals special case: a boolean column expr and a 1 or 0 (number) result in bool conversions.
		{"eq(bool column,1)", NewSQLEqualsExpr(boolColVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"eq(1,bool column)", NewSQLEqualsExpr(intVal, boolColVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"eq(bool column,0)", NewSQLEqualsExpr(boolColVal, NewSQLValueExpr(values.NewSQLInt64(knd, 0))), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"eq(0,bool column)", NewSQLEqualsExpr(NewSQLValueExpr(values.NewSQLInt64(knd, 0)), boolColVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"eq(bool column,2)", NewSQLEqualsExpr(boolColVal, NewSQLValueExpr(values.NewSQLInt64(knd, 2))), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"eq(2,bool column)", NewSQLEqualsExpr(NewSQLValueExpr(values.NewSQLInt64(knd, 2)), boolColVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},

		// greater than.
		{"gt(int,int)", NewSQLGreaterThanExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gt(int,uint)", NewSQLGreaterThanExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"gt(int,float)", NewSQLGreaterThanExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"gt(int,decimal)", NewSQLGreaterThanExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"gt(int,bool)", NewSQLGreaterThanExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gt(int,str)", NewSQLGreaterThanExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gt(int,date)", NewSQLGreaterThanExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gt(int,datetime)", NewSQLGreaterThanExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gt(uint,int)", NewSQLGreaterThanExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"gt(uint,uint)", NewSQLGreaterThanExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gt(uint,float)", NewSQLGreaterThanExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalDouble}},
		{"gt(uint,decimal)", NewSQLGreaterThanExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"gt(uint,bool)", NewSQLGreaterThanExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gt(uint,str)", NewSQLGreaterThanExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gt(uint,date)", NewSQLGreaterThanExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gt(uint,datetime)", NewSQLGreaterThanExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gt(float,int)", NewSQLGreaterThanExpr(floatVal, intVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"gt(float,uint)", NewSQLGreaterThanExpr(floatVal, uintVal), []types.EvalType{types.EvalDouble, types.EvalUint64}},
		{"gt(float,float)", NewSQLGreaterThanExpr(floatVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gt(float,decimal)", NewSQLGreaterThanExpr(floatVal, decimalVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"gt(float,bool)", NewSQLGreaterThanExpr(floatVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gt(float,str)", NewSQLGreaterThanExpr(floatVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gt(float,date)", NewSQLGreaterThanExpr(floatVal, dateVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gt(float,datetime)", NewSQLGreaterThanExpr(floatVal, datetimeVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gt(decimal,int)", NewSQLGreaterThanExpr(decimalVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"gt(decimal,uint)", NewSQLGreaterThanExpr(decimalVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"gt(decimal,float)", NewSQLGreaterThanExpr(decimalVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"gt(decimal,decimal)", NewSQLGreaterThanExpr(decimalVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gt(decimal,bool)", NewSQLGreaterThanExpr(decimalVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gt(decimal,str)", NewSQLGreaterThanExpr(decimalVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gt(decimal,date)", NewSQLGreaterThanExpr(decimalVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gt(decimal,datetime)", NewSQLGreaterThanExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gt(bool,int)", NewSQLGreaterThanExpr(boolVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gt(bool,uint)", NewSQLGreaterThanExpr(boolVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gt(bool,float)", NewSQLGreaterThanExpr(boolVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gt(bool,decimal)", NewSQLGreaterThanExpr(boolVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gt(bool,bool)", NewSQLGreaterThanExpr(boolVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"gt(bool,str)", NewSQLGreaterThanExpr(boolVal, strVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"gt(bool,date)", NewSQLGreaterThanExpr(boolVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"gt(bool,datetime)", NewSQLGreaterThanExpr(boolVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"gt(str,int)", NewSQLGreaterThanExpr(strVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gt(str,uint)", NewSQLGreaterThanExpr(strVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gt(str,float)", NewSQLGreaterThanExpr(strVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gt(str,decimal)", NewSQLGreaterThanExpr(strVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gt(str,bool)", NewSQLGreaterThanExpr(strVal, boolVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"gt(str,str)", NewSQLGreaterThanExpr(strVal, strVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"gt(str,date)", NewSQLGreaterThanExpr(strVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"gt(str,datetime)", NewSQLGreaterThanExpr(strVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"gt(date,int)", NewSQLGreaterThanExpr(dateVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gt(date,uint)", NewSQLGreaterThanExpr(dateVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gt(date,float)", NewSQLGreaterThanExpr(dateVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gt(date,decimal)", NewSQLGreaterThanExpr(dateVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gt(date,bool)", NewSQLGreaterThanExpr(dateVal, boolVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"gt(date,str)", NewSQLGreaterThanExpr(dateVal, strVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"gt(date,date)", NewSQLGreaterThanExpr(dateVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"gt(date,datetime)", NewSQLGreaterThanExpr(dateVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"gt(datetime,int)", NewSQLGreaterThanExpr(datetimeVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gt(datetime,uint)", NewSQLGreaterThanExpr(datetimeVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gt(datetime,float)", NewSQLGreaterThanExpr(datetimeVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gt(datetime,decimal)", NewSQLGreaterThanExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gt(datetime,bool)", NewSQLGreaterThanExpr(datetimeVal, boolVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"gt(datetime,str)", NewSQLGreaterThanExpr(datetimeVal, strVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"gt(datetime,date)", NewSQLGreaterThanExpr(datetimeVal, dateVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"gt(datetime,datetime)", NewSQLGreaterThanExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},

		// greater than or equal.
		{"gte(int,int)", NewSQLGreaterThanOrEqualExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gte(int,uint)", NewSQLGreaterThanOrEqualExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"gte(int,float)", NewSQLGreaterThanOrEqualExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"gte(int,decimal)", NewSQLGreaterThanOrEqualExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"gte(int,bool)", NewSQLGreaterThanOrEqualExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gte(int,str)", NewSQLGreaterThanOrEqualExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gte(int,date)", NewSQLGreaterThanOrEqualExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gte(int,datetime)", NewSQLGreaterThanOrEqualExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gte(uint,int)", NewSQLGreaterThanOrEqualExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"gte(uint,uint)", NewSQLGreaterThanOrEqualExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gte(uint,float)", NewSQLGreaterThanOrEqualExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalDouble}},
		{"gte(uint,decimal)", NewSQLGreaterThanOrEqualExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"gte(uint,bool)", NewSQLGreaterThanOrEqualExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gte(uint,str)", NewSQLGreaterThanOrEqualExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gte(uint,date)", NewSQLGreaterThanOrEqualExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gte(uint,datetime)", NewSQLGreaterThanOrEqualExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gte(float,int)", NewSQLGreaterThanOrEqualExpr(floatVal, intVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"gte(float,uint)", NewSQLGreaterThanOrEqualExpr(floatVal, uintVal), []types.EvalType{types.EvalDouble, types.EvalUint64}},
		{"gte(float,float)", NewSQLGreaterThanOrEqualExpr(floatVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gte(float,decimal)", NewSQLGreaterThanOrEqualExpr(floatVal, decimalVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"gte(float,bool)", NewSQLGreaterThanOrEqualExpr(floatVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gte(float,str)", NewSQLGreaterThanOrEqualExpr(floatVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gte(float,date)", NewSQLGreaterThanOrEqualExpr(floatVal, dateVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gte(float,datetime)", NewSQLGreaterThanOrEqualExpr(floatVal, datetimeVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gte(decimal,int)", NewSQLGreaterThanOrEqualExpr(decimalVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"gte(decimal,uint)", NewSQLGreaterThanOrEqualExpr(decimalVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"gte(decimal,float)", NewSQLGreaterThanOrEqualExpr(decimalVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"gte(decimal,decimal)", NewSQLGreaterThanOrEqualExpr(decimalVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gte(decimal,bool)", NewSQLGreaterThanOrEqualExpr(decimalVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gte(decimal,str)", NewSQLGreaterThanOrEqualExpr(decimalVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gte(decimal,date)", NewSQLGreaterThanOrEqualExpr(decimalVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gte(decimal,datetime)", NewSQLGreaterThanOrEqualExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gte(bool,int)", NewSQLGreaterThanOrEqualExpr(boolVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gte(bool,uint)", NewSQLGreaterThanOrEqualExpr(boolVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gte(bool,float)", NewSQLGreaterThanOrEqualExpr(boolVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gte(bool,decimal)", NewSQLGreaterThanOrEqualExpr(boolVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gte(bool,bool)", NewSQLGreaterThanOrEqualExpr(boolVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"gte(bool,str)", NewSQLGreaterThanOrEqualExpr(boolVal, strVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"gte(bool,date)", NewSQLGreaterThanOrEqualExpr(boolVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"gte(bool,datetime)", NewSQLGreaterThanOrEqualExpr(boolVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"gte(str,int)", NewSQLGreaterThanOrEqualExpr(strVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gte(str,uint)", NewSQLGreaterThanOrEqualExpr(strVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gte(str,float)", NewSQLGreaterThanOrEqualExpr(strVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gte(str,decimal)", NewSQLGreaterThanOrEqualExpr(strVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gte(str,bool)", NewSQLGreaterThanOrEqualExpr(strVal, boolVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"gte(str,str)", NewSQLGreaterThanOrEqualExpr(strVal, strVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"gte(str,date)", NewSQLGreaterThanOrEqualExpr(strVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"gte(str,datetime)", NewSQLGreaterThanOrEqualExpr(strVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"gte(date,int)", NewSQLGreaterThanOrEqualExpr(dateVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gte(date,uint)", NewSQLGreaterThanOrEqualExpr(dateVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gte(date,float)", NewSQLGreaterThanOrEqualExpr(dateVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gte(date,decimal)", NewSQLGreaterThanOrEqualExpr(dateVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gte(date,bool)", NewSQLGreaterThanOrEqualExpr(dateVal, boolVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"gte(date,str)", NewSQLGreaterThanOrEqualExpr(dateVal, strVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"gte(date,date)", NewSQLGreaterThanOrEqualExpr(dateVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"gte(date,datetime)", NewSQLGreaterThanOrEqualExpr(dateVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"gte(datetime,int)", NewSQLGreaterThanOrEqualExpr(datetimeVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"gte(datetime,uint)", NewSQLGreaterThanOrEqualExpr(datetimeVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"gte(datetime,float)", NewSQLGreaterThanOrEqualExpr(datetimeVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"gte(datetime,decimal)", NewSQLGreaterThanOrEqualExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"gte(datetime,bool)", NewSQLGreaterThanOrEqualExpr(datetimeVal, boolVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"gte(datetime,str)", NewSQLGreaterThanOrEqualExpr(datetimeVal, strVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"gte(datetime,date)", NewSQLGreaterThanOrEqualExpr(datetimeVal, dateVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"gte(datetime,datetime)", NewSQLGreaterThanOrEqualExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},

		// less than.
		{"lt(int,int)", NewSQLLessThanExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lt(int,uint)", NewSQLLessThanExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"lt(int,float)", NewSQLLessThanExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"lt(int,decimal)", NewSQLLessThanExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"lt(int,bool)", NewSQLLessThanExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lt(int,str)", NewSQLLessThanExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lt(int,date)", NewSQLLessThanExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lt(int,datetime)", NewSQLLessThanExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lt(uint,int)", NewSQLLessThanExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"lt(uint,uint)", NewSQLLessThanExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lt(uint,float)", NewSQLLessThanExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalDouble}},
		{"lt(uint,decimal)", NewSQLLessThanExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"lt(uint,bool)", NewSQLLessThanExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lt(uint,str)", NewSQLLessThanExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lt(uint,date)", NewSQLLessThanExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lt(uint,datetime)", NewSQLLessThanExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lt(float,int)", NewSQLLessThanExpr(floatVal, intVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"lt(float,uint)", NewSQLLessThanExpr(floatVal, uintVal), []types.EvalType{types.EvalDouble, types.EvalUint64}},
		{"lt(float,float)", NewSQLLessThanExpr(floatVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lt(float,decimal)", NewSQLLessThanExpr(floatVal, decimalVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"lt(float,bool)", NewSQLLessThanExpr(floatVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lt(float,str)", NewSQLLessThanExpr(floatVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lt(float,date)", NewSQLLessThanExpr(floatVal, dateVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lt(float,datetime)", NewSQLLessThanExpr(floatVal, datetimeVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lt(decimal,int)", NewSQLLessThanExpr(decimalVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"lt(decimal,uint)", NewSQLLessThanExpr(decimalVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"lt(decimal,float)", NewSQLLessThanExpr(decimalVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"lt(decimal,decimal)", NewSQLLessThanExpr(decimalVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lt(decimal,bool)", NewSQLLessThanExpr(decimalVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lt(decimal,str)", NewSQLLessThanExpr(decimalVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lt(decimal,date)", NewSQLLessThanExpr(decimalVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lt(decimal,datetime)", NewSQLLessThanExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lt(bool,int)", NewSQLLessThanExpr(boolVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lt(bool,uint)", NewSQLLessThanExpr(boolVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lt(bool,float)", NewSQLLessThanExpr(boolVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lt(bool,decimal)", NewSQLLessThanExpr(boolVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lt(bool,bool)", NewSQLLessThanExpr(boolVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"lt(bool,str)", NewSQLLessThanExpr(boolVal, strVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"lt(bool,date)", NewSQLLessThanExpr(boolVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"lt(bool,datetime)", NewSQLLessThanExpr(boolVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"lt(str,int)", NewSQLLessThanExpr(strVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lt(str,uint)", NewSQLLessThanExpr(strVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lt(str,float)", NewSQLLessThanExpr(strVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lt(str,decimal)", NewSQLLessThanExpr(strVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lt(str,bool)", NewSQLLessThanExpr(strVal, boolVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"lt(str,str)", NewSQLLessThanExpr(strVal, strVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"lt(str,date)", NewSQLLessThanExpr(strVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"lt(str,datetime)", NewSQLLessThanExpr(strVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"lt(date,int)", NewSQLLessThanExpr(dateVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lt(date,uint)", NewSQLLessThanExpr(dateVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lt(date,float)", NewSQLLessThanExpr(dateVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lt(date,decimal)", NewSQLLessThanExpr(dateVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lt(date,bool)", NewSQLLessThanExpr(dateVal, boolVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"lt(date,str)", NewSQLLessThanExpr(dateVal, strVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"lt(date,date)", NewSQLLessThanExpr(dateVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"lt(date,datetime)", NewSQLLessThanExpr(dateVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"lt(datetime,int)", NewSQLLessThanExpr(datetimeVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lt(datetime,uint)", NewSQLLessThanExpr(datetimeVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lt(datetime,float)", NewSQLLessThanExpr(datetimeVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lt(datetime,decimal)", NewSQLLessThanExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lt(datetime,bool)", NewSQLLessThanExpr(datetimeVal, boolVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"lt(datetime,str)", NewSQLLessThanExpr(datetimeVal, strVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"lt(datetime,date)", NewSQLLessThanExpr(datetimeVal, dateVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"lt(datetime,datetime)", NewSQLLessThanExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},

		// less than or equal.
		{"lte(int,int)", NewSQLLessThanOrEqualExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lte(int,uint)", NewSQLLessThanOrEqualExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"lte(int,float)", NewSQLLessThanOrEqualExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"lte(int,decimal)", NewSQLLessThanOrEqualExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"lte(int,bool)", NewSQLLessThanOrEqualExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lte(int,str)", NewSQLLessThanOrEqualExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lte(int,date)", NewSQLLessThanOrEqualExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lte(int,datetime)", NewSQLLessThanOrEqualExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lte(uint,int)", NewSQLLessThanOrEqualExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"lte(uint,uint)", NewSQLLessThanOrEqualExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lte(uint,float)", NewSQLLessThanOrEqualExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalDouble}},
		{"lte(uint,decimal)", NewSQLLessThanOrEqualExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"lte(uint,bool)", NewSQLLessThanOrEqualExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lte(uint,str)", NewSQLLessThanOrEqualExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lte(uint,date)", NewSQLLessThanOrEqualExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lte(uint,datetime)", NewSQLLessThanOrEqualExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lte(float,int)", NewSQLLessThanOrEqualExpr(floatVal, intVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"lte(float,uint)", NewSQLLessThanOrEqualExpr(floatVal, uintVal), []types.EvalType{types.EvalDouble, types.EvalUint64}},
		{"lte(float,float)", NewSQLLessThanOrEqualExpr(floatVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lte(float,decimal)", NewSQLLessThanOrEqualExpr(floatVal, decimalVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"lte(float,bool)", NewSQLLessThanOrEqualExpr(floatVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lte(float,str)", NewSQLLessThanOrEqualExpr(floatVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lte(float,date)", NewSQLLessThanOrEqualExpr(floatVal, dateVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lte(float,datetime)", NewSQLLessThanOrEqualExpr(floatVal, datetimeVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lte(decimal,int)", NewSQLLessThanOrEqualExpr(decimalVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"lte(decimal,uint)", NewSQLLessThanOrEqualExpr(decimalVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"lte(decimal,float)", NewSQLLessThanOrEqualExpr(decimalVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"lte(decimal,decimal)", NewSQLLessThanOrEqualExpr(decimalVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lte(decimal,bool)", NewSQLLessThanOrEqualExpr(decimalVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lte(decimal,str)", NewSQLLessThanOrEqualExpr(decimalVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lte(decimal,date)", NewSQLLessThanOrEqualExpr(decimalVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lte(decimal,datetime)", NewSQLLessThanOrEqualExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lte(bool,int)", NewSQLLessThanOrEqualExpr(boolVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lte(bool,uint)", NewSQLLessThanOrEqualExpr(boolVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lte(bool,float)", NewSQLLessThanOrEqualExpr(boolVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lte(bool,decimal)", NewSQLLessThanOrEqualExpr(boolVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lte(bool,bool)", NewSQLLessThanOrEqualExpr(boolVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"lte(bool,str)", NewSQLLessThanOrEqualExpr(boolVal, strVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"lte(bool,date)", NewSQLLessThanOrEqualExpr(boolVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"lte(bool,datetime)", NewSQLLessThanOrEqualExpr(boolVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"lte(str,int)", NewSQLLessThanOrEqualExpr(strVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lte(str,uint)", NewSQLLessThanOrEqualExpr(strVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lte(str,float)", NewSQLLessThanOrEqualExpr(strVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lte(str,decimal)", NewSQLLessThanOrEqualExpr(strVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lte(str,bool)", NewSQLLessThanOrEqualExpr(strVal, boolVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"lte(str,str)", NewSQLLessThanOrEqualExpr(strVal, strVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"lte(str,date)", NewSQLLessThanOrEqualExpr(strVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"lte(str,datetime)", NewSQLLessThanOrEqualExpr(strVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"lte(date,int)", NewSQLLessThanOrEqualExpr(dateVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lte(date,uint)", NewSQLLessThanOrEqualExpr(dateVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lte(date,float)", NewSQLLessThanOrEqualExpr(dateVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lte(date,decimal)", NewSQLLessThanOrEqualExpr(dateVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lte(date,bool)", NewSQLLessThanOrEqualExpr(dateVal, boolVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"lte(date,str)", NewSQLLessThanOrEqualExpr(dateVal, strVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"lte(date,date)", NewSQLLessThanOrEqualExpr(dateVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"lte(date,datetime)", NewSQLLessThanOrEqualExpr(dateVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"lte(datetime,int)", NewSQLLessThanOrEqualExpr(datetimeVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"lte(datetime,uint)", NewSQLLessThanOrEqualExpr(datetimeVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"lte(datetime,float)", NewSQLLessThanOrEqualExpr(datetimeVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"lte(datetime,decimal)", NewSQLLessThanOrEqualExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"lte(datetime,bool)", NewSQLLessThanOrEqualExpr(datetimeVal, boolVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"lte(datetime,str)", NewSQLLessThanOrEqualExpr(datetimeVal, strVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"lte(datetime,date)", NewSQLLessThanOrEqualExpr(datetimeVal, dateVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"lte(datetime,datetime)", NewSQLLessThanOrEqualExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},

		// not equal.
		{"neq(int,int)", NewSQLNotEqualsExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"neq(int,uint)", NewSQLNotEqualsExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"neq(int,float)", NewSQLNotEqualsExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"neq(int,decimal)", NewSQLNotEqualsExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"neq(int,bool)", NewSQLNotEqualsExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"neq(int,str)", NewSQLNotEqualsExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"neq(int,date)", NewSQLNotEqualsExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"neq(int,datetime)", NewSQLNotEqualsExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"neq(uint,int)", NewSQLNotEqualsExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"neq(uint,uint)", NewSQLNotEqualsExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"neq(uint,float)", NewSQLNotEqualsExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalDouble}},
		{"neq(uint,decimal)", NewSQLNotEqualsExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"neq(uint,bool)", NewSQLNotEqualsExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"neq(uint,str)", NewSQLNotEqualsExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"neq(uint,date)", NewSQLNotEqualsExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"neq(uint,datetime)", NewSQLNotEqualsExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"neq(float,int)", NewSQLNotEqualsExpr(floatVal, intVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"neq(float,uint)", NewSQLNotEqualsExpr(floatVal, uintVal), []types.EvalType{types.EvalDouble, types.EvalUint64}},
		{"neq(float,float)", NewSQLNotEqualsExpr(floatVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"neq(float,decimal)", NewSQLNotEqualsExpr(floatVal, decimalVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"neq(float,bool)", NewSQLNotEqualsExpr(floatVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"neq(float,str)", NewSQLNotEqualsExpr(floatVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"neq(float,date)", NewSQLNotEqualsExpr(floatVal, dateVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"neq(float,datetime)", NewSQLNotEqualsExpr(floatVal, datetimeVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"neq(decimal,int)", NewSQLNotEqualsExpr(decimalVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"neq(decimal,uint)", NewSQLNotEqualsExpr(decimalVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"neq(decimal,float)", NewSQLNotEqualsExpr(decimalVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"neq(decimal,decimal)", NewSQLNotEqualsExpr(decimalVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"neq(decimal,bool)", NewSQLNotEqualsExpr(decimalVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"neq(decimal,str)", NewSQLNotEqualsExpr(decimalVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"neq(decimal,date)", NewSQLNotEqualsExpr(decimalVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"neq(decimal,datetime)", NewSQLNotEqualsExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"neq(bool,int)", NewSQLNotEqualsExpr(boolVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"neq(bool,uint)", NewSQLNotEqualsExpr(boolVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"neq(bool,float)", NewSQLNotEqualsExpr(boolVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"neq(bool,decimal)", NewSQLNotEqualsExpr(boolVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"neq(bool,bool)", NewSQLNotEqualsExpr(boolVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"neq(bool,str)", NewSQLNotEqualsExpr(boolVal, strVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"neq(bool,date)", NewSQLNotEqualsExpr(boolVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"neq(bool,datetime)", NewSQLNotEqualsExpr(boolVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"neq(str,int)", NewSQLNotEqualsExpr(strVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"neq(str,uint)", NewSQLNotEqualsExpr(strVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"neq(str,float)", NewSQLNotEqualsExpr(strVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"neq(str,decimal)", NewSQLNotEqualsExpr(strVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"neq(str,bool)", NewSQLNotEqualsExpr(strVal, boolVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"neq(str,str)", NewSQLNotEqualsExpr(strVal, strVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"neq(str,date)", NewSQLNotEqualsExpr(strVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"neq(str,datetime)", NewSQLNotEqualsExpr(strVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"neq(date,int)", NewSQLNotEqualsExpr(dateVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"neq(date,uint)", NewSQLNotEqualsExpr(dateVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"neq(date,float)", NewSQLNotEqualsExpr(dateVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"neq(date,decimal)", NewSQLNotEqualsExpr(dateVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"neq(date,bool)", NewSQLNotEqualsExpr(dateVal, boolVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"neq(date,str)", NewSQLNotEqualsExpr(dateVal, strVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"neq(date,date)", NewSQLNotEqualsExpr(dateVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"neq(date,datetime)", NewSQLNotEqualsExpr(dateVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"neq(datetime,int)", NewSQLNotEqualsExpr(datetimeVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"neq(datetime,uint)", NewSQLNotEqualsExpr(datetimeVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"neq(datetime,float)", NewSQLNotEqualsExpr(datetimeVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"neq(datetime,decimal)", NewSQLNotEqualsExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"neq(datetime,bool)", NewSQLNotEqualsExpr(datetimeVal, boolVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"neq(datetime,str)", NewSQLNotEqualsExpr(datetimeVal, strVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"neq(datetime,date)", NewSQLNotEqualsExpr(datetimeVal, dateVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"neq(datetime,datetime)", NewSQLNotEqualsExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},

		// null-safe equal.
		{"nse(int,int)", NewSQLNullSafeEqualsExpr(intVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"nse(int,uint)", NewSQLNullSafeEqualsExpr(intVal, uintVal), []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"nse(int,float)", NewSQLNullSafeEqualsExpr(intVal, floatVal), []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"nse(int,decimal)", NewSQLNullSafeEqualsExpr(intVal, decimalVal), []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"nse(int,bool)", NewSQLNullSafeEqualsExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"nse(int,str)", NewSQLNullSafeEqualsExpr(intVal, strVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"nse(int,date)", NewSQLNullSafeEqualsExpr(intVal, dateVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"nse(int,datetime)", NewSQLNullSafeEqualsExpr(intVal, datetimeVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"nse(uint,int)", NewSQLNullSafeEqualsExpr(uintVal, intVal), []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"nse(uint,uint)", NewSQLNullSafeEqualsExpr(uintVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"nse(uint,float)", NewSQLNullSafeEqualsExpr(uintVal, floatVal), []types.EvalType{types.EvalUint64, types.EvalDouble}},
		{"nse(uint,decimal)", NewSQLNullSafeEqualsExpr(uintVal, decimalVal), []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"nse(uint,bool)", NewSQLNullSafeEqualsExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"nse(uint,str)", NewSQLNullSafeEqualsExpr(uintVal, strVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"nse(uint,date)", NewSQLNullSafeEqualsExpr(uintVal, dateVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"nse(uint,datetime)", NewSQLNullSafeEqualsExpr(uintVal, datetimeVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"nse(float,int)", NewSQLNullSafeEqualsExpr(floatVal, intVal), []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"nse(float,uint)", NewSQLNullSafeEqualsExpr(floatVal, uintVal), []types.EvalType{types.EvalDouble, types.EvalUint64}},
		{"nse(float,float)", NewSQLNullSafeEqualsExpr(floatVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"nse(float,decimal)", NewSQLNullSafeEqualsExpr(floatVal, decimalVal), []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"nse(float,bool)", NewSQLNullSafeEqualsExpr(floatVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"nse(float,str)", NewSQLNullSafeEqualsExpr(floatVal, strVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"nse(float,date)", NewSQLNullSafeEqualsExpr(floatVal, dateVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"nse(float,datetime)", NewSQLNullSafeEqualsExpr(floatVal, datetimeVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"nse(decimal,int)", NewSQLNullSafeEqualsExpr(decimalVal, intVal), []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"nse(decimal,uint)", NewSQLNullSafeEqualsExpr(decimalVal, uintVal), []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"nse(decimal,float)", NewSQLNullSafeEqualsExpr(decimalVal, floatVal), []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"nse(decimal,decimal)", NewSQLNullSafeEqualsExpr(decimalVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"nse(decimal,bool)", NewSQLNullSafeEqualsExpr(decimalVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"nse(decimal,str)", NewSQLNullSafeEqualsExpr(decimalVal, strVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"nse(decimal,date)", NewSQLNullSafeEqualsExpr(decimalVal, dateVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"nse(decimal,datetime)", NewSQLNullSafeEqualsExpr(decimalVal, datetimeVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"nse(bool,int)", NewSQLNullSafeEqualsExpr(boolVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"nse(bool,uint)", NewSQLNullSafeEqualsExpr(boolVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"nse(bool,float)", NewSQLNullSafeEqualsExpr(boolVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"nse(bool,decimal)", NewSQLNullSafeEqualsExpr(boolVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"nse(bool,bool)", NewSQLNullSafeEqualsExpr(boolVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"nse(bool,str)", NewSQLNullSafeEqualsExpr(boolVal, strVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"nse(bool,date)", NewSQLNullSafeEqualsExpr(boolVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"nse(bool,datetime)", NewSQLNullSafeEqualsExpr(boolVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"nse(str,int)", NewSQLNullSafeEqualsExpr(strVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"nse(str,uint)", NewSQLNullSafeEqualsExpr(strVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"nse(str,float)", NewSQLNullSafeEqualsExpr(strVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"nse(str,decimal)", NewSQLNullSafeEqualsExpr(strVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"nse(str,bool)", NewSQLNullSafeEqualsExpr(strVal, boolVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"nse(str,str)", NewSQLNullSafeEqualsExpr(strVal, strVal), []types.EvalType{types.EvalString, types.EvalString}},
		{"nse(str,date)", NewSQLNullSafeEqualsExpr(strVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"nse(str,datetime)", NewSQLNullSafeEqualsExpr(strVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"nse(date,int)", NewSQLNullSafeEqualsExpr(dateVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"nse(date,uint)", NewSQLNullSafeEqualsExpr(dateVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"nse(date,float)", NewSQLNullSafeEqualsExpr(dateVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"nse(date,decimal)", NewSQLNullSafeEqualsExpr(dateVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"nse(date,bool)", NewSQLNullSafeEqualsExpr(dateVal, boolVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"nse(date,str)", NewSQLNullSafeEqualsExpr(dateVal, strVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"nse(date,date)", NewSQLNullSafeEqualsExpr(dateVal, dateVal), []types.EvalType{types.EvalDate, types.EvalDate}},
		{"nse(date,datetime)", NewSQLNullSafeEqualsExpr(dateVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"nse(datetime,int)", NewSQLNullSafeEqualsExpr(datetimeVal, intVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"nse(datetime,uint)", NewSQLNullSafeEqualsExpr(datetimeVal, uintVal), []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"nse(datetime,float)", NewSQLNullSafeEqualsExpr(datetimeVal, floatVal), []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"nse(datetime,decimal)", NewSQLNullSafeEqualsExpr(datetimeVal, decimalVal), []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"nse(datetime,bool)", NewSQLNullSafeEqualsExpr(datetimeVal, boolVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"nse(datetime,str)", NewSQLNullSafeEqualsExpr(datetimeVal, strVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"nse(datetime,date)", NewSQLNullSafeEqualsExpr(datetimeVal, dateVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"nse(datetime,datetime)", NewSQLNullSafeEqualsExpr(datetimeVal, datetimeVal), []types.EvalType{types.EvalDatetime, types.EvalDatetime}},

		// is expression: right must always be boolean; does not convert the
		// left if it is numeric or boolean; does convert the left to boolean
		// otherwise.
		{"is(int,bool)", NewSQLIsExpr(intVal, boolVal), []types.EvalType{types.EvalInt64, types.EvalBoolean}},
		{"is(uint,bool)", NewSQLIsExpr(uintVal, boolVal), []types.EvalType{types.EvalUint64, types.EvalBoolean}},
		{"is(float,bool)", NewSQLIsExpr(floatVal, boolVal), []types.EvalType{types.EvalDouble, types.EvalBoolean}},
		{"is(decimal,bool)", NewSQLIsExpr(decimalVal, boolVal), []types.EvalType{types.EvalDecimal128, types.EvalBoolean}},
		{"is(bool,bool)", NewSQLIsExpr(boolVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"is(str,bool)", NewSQLIsExpr(strVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"is(date,bool)", NewSQLIsExpr(dateVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"is(datetime,bool)", NewSQLIsExpr(datetimeVal, boolVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},

		// not expression: does not convert boolean comparable types (bool,
		// int, uint); does convert other types to bool.
		{"not(int)", NewSQLNotExpr(intVal), []types.EvalType{types.EvalInt64}},
		{"not(uint)", NewSQLNotExpr(uintVal), []types.EvalType{types.EvalUint64}},
		{"not(float)", NewSQLNotExpr(floatVal), []types.EvalType{types.EvalBoolean}},
		{"not(decimal)", NewSQLNotExpr(decimalVal), []types.EvalType{types.EvalBoolean}},
		{"not(bool)", NewSQLNotExpr(boolVal), []types.EvalType{types.EvalBoolean}},
		{"not(str)", NewSQLNotExpr(strVal), []types.EvalType{types.EvalBoolean}},
		{"not(date)", NewSQLNotExpr(dateVal), []types.EvalType{types.EvalBoolean}},
		{"not(datetime)", NewSQLNotExpr(datetimeVal), []types.EvalType{types.EvalBoolean}},

		// unary minus expression: does not convert numeric types; does convert
		// other types to int, except string (which is converted to float)
		{"unary_minus(int)", NewSQLUnaryMinusExpr(intVal), []types.EvalType{types.EvalInt64}},
		{"unary_minus(uint)", NewSQLUnaryMinusExpr(uintVal), []types.EvalType{types.EvalUint64}},
		{"unary_minus(float)", NewSQLUnaryMinusExpr(floatVal), []types.EvalType{types.EvalDouble}},
		{"unary_minus(decimal)", NewSQLUnaryMinusExpr(decimalVal), []types.EvalType{types.EvalDecimal128}},
		{"unary_minus(bool)", NewSQLUnaryMinusExpr(boolVal), []types.EvalType{types.EvalInt64}},
		{"unary_minus(str)", NewSQLUnaryMinusExpr(strVal), []types.EvalType{types.EvalDouble}},
		{"unary_minus(date)", NewSQLUnaryMinusExpr(dateVal), []types.EvalType{types.EvalInt64}},
		{"unary_minus(datetime)", NewSQLUnaryMinusExpr(datetimeVal), []types.EvalType{types.EvalInt64}},

		// tilde expression: does not convert numeric types; does convert
		// other types to int.
		{"tilde(int)", NewSQLTildeExpr(intVal), []types.EvalType{types.EvalInt64}},
		{"tilde(uint)", NewSQLTildeExpr(uintVal), []types.EvalType{types.EvalUint64}},
		{"tilde(float)", NewSQLTildeExpr(floatVal), []types.EvalType{types.EvalDouble}},
		{"tilde(decimal)", NewSQLTildeExpr(decimalVal), []types.EvalType{types.EvalDecimal128}},
		{"tilde(bool)", NewSQLTildeExpr(boolVal), []types.EvalType{types.EvalInt64}},
		{"tilde(str)", NewSQLTildeExpr(strVal), []types.EvalType{types.EvalInt64}},
		{"tilde(date)", NewSQLTildeExpr(dateVal), []types.EvalType{types.EvalInt64}},
		{"tilde(datetime)", NewSQLTildeExpr(datetimeVal), []types.EvalType{types.EvalInt64}},

		// agg functions: all of the aggregation functions' reconcile methods are no-ops, so there are no conversions.
		{"avg", NewSQLAggregationFunctionExpr(parser.AvgAggregateName, false, allVals), allTypes},
		{"count", NewSQLAggregationFunctionExpr(parser.CountAggregateName, false, allVals), allTypes},
		{"groupConcat", NewSQLAggregationFunctionExpr(parser.GroupConcatAggregateName, false, allVals), allTypes},
		{"max", NewSQLAggregationFunctionExpr(parser.MaxAggregateName, false, allVals), allTypes},
		{"min", NewSQLAggregationFunctionExpr(parser.MinAggregateName, false, allVals), allTypes},
		{"sum", NewSQLAggregationFunctionExpr(parser.SumAggregateName, false, allVals), allTypes},
		{"stdDev", NewSQLAggregationFunctionExpr(parser.StdDevAggregateName, false, allVals), allTypes},
		{"stdDevSample", NewSQLAggregationFunctionExpr(parser.StdDevSampleAggregateName, false, allVals), allTypes},
	}

	runReconcileTests(t, tests)
}

func TestReconcileSubqueryPlans(t *testing.T) {
	knd := values.MySQLValueKind

	intVal := NewSQLValueExpr(values.NewSQLInt64(knd, 1))
	uintVal := NewSQLValueExpr(values.NewSQLUint64(knd, 1))
	floatVal := NewSQLValueExpr(values.NewSQLFloat(knd, 1))
	decimalVal := NewSQLValueExpr(values.NewSQLDecimal128(knd, decimal.NewFromFloat(1.0)))
	boolVal := NewSQLValueExpr(values.NewSQLBool(knd, true))
	strVal := NewSQLValueExpr(values.NewSQLVarchar(knd, "bar"))
	dateVal := NewSQLValueExpr(values.NewSQLDate(knd, time.Now()))
	datetimeVal := NewSQLValueExpr(values.NewSQLTimestamp(knd, time.Now()))

	intCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalInt64, schema.MongoInt, false)
	uintCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalUint64, schema.MongoInt, false)
	floatCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalDouble, schema.MongoInt, false)
	decimalCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalDecimal128, schema.MongoInt, false)
	boolCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalBoolean, schema.MongoInt, false)
	strCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalString, schema.MongoInt, false)
	dateCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalDate, schema.MongoInt, false)
	datetimeCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalDatetime, schema.MongoInt, false)

	intPC := ProjectedColumn{intCol, intVal}
	uintPC := ProjectedColumn{uintCol, uintVal}
	floatPC := ProjectedColumn{floatCol, floatVal}
	decimalPC := ProjectedColumn{decimalCol, decimalVal}
	boolPC := ProjectedColumn{boolCol, boolVal}
	strPC := ProjectedColumn{strCol, strVal}
	datePC := ProjectedColumn{dateCol, dateVal}
	datetimePC := ProjectedColumn{datetimeCol, datetimeVal}

	type test struct {
		name                                  string
		leftPlanStage, rightPlanStage         PlanStage
		leftExpectedTypes, rightExpectedTypes []types.EvalType
	}

	compareColumnExprTypes := func(expected []types.EvalType, actual []ProjectedColumn, side string, req *require.Assertions) {
		req.Equal(len(expected), len(actual),
			"number of expected types on %s (%d) does not match number of actual projected columns (%d)",
			side, len(expected), len(actual),
		)

		for i, c := range actual {
			req.Equal(expected[i], c.Expr.EvalType(), "incorrect EvalType for column %d on %s side", i, side)
		}
	}

	runTests := func(t *testing.T, tests []test) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				req := require.New(t)

				reconciledLeft, reconciledRight := reconcileSubqueryPlans(test.leftPlanStage, test.rightPlanStage)

				leftColumns := reconciledLeft.(*ProjectStage).projectedColumns
				rightColumns := reconciledRight.(*ProjectStage).projectedColumns

				compareColumnExprTypes(test.leftExpectedTypes, leftColumns, "left", req)
				compareColumnExprTypes(test.rightExpectedTypes, rightColumns, "right", req)
			})
		}
	}

	tests := []test{
		{"empty", NewProjectStage(NewDualStage()), NewProjectStage(NewDualStage()), []types.EvalType{}, []types.EvalType{}},
		{
			"single(int,int)",
			NewProjectStage(NewDualStage(), intPC), NewProjectStage(NewDualStage(), intPC),
			[]types.EvalType{types.EvalInt64}, []types.EvalType{types.EvalInt64},
		},
		{
			"single(int,uint)",
			NewProjectStage(NewDualStage(), intPC), NewProjectStage(NewDualStage(), uintPC),
			[]types.EvalType{types.EvalInt64}, []types.EvalType{types.EvalUint64},
		},
		{
			"single(int,float)",
			NewProjectStage(NewDualStage(), intPC), NewProjectStage(NewDualStage(), floatPC),
			[]types.EvalType{types.EvalInt64}, []types.EvalType{types.EvalDouble},
		},
		{
			"single(int,decimal)",
			NewProjectStage(NewDualStage(), intPC), NewProjectStage(NewDualStage(), decimalPC),
			[]types.EvalType{types.EvalInt64}, []types.EvalType{types.EvalDecimal128},
		},
		{
			"single(int,bool)",
			NewProjectStage(NewDualStage(), intPC), NewProjectStage(NewDualStage(), boolPC),
			[]types.EvalType{types.EvalInt64}, []types.EvalType{types.EvalInt64},
		},
		{
			"single(int,str)",
			NewProjectStage(NewDualStage(), intPC), NewProjectStage(NewDualStage(), strPC),
			[]types.EvalType{types.EvalInt64}, []types.EvalType{types.EvalInt64},
		},
		{
			"single(int,date)",
			NewProjectStage(NewDualStage(), intPC), NewProjectStage(NewDualStage(), datePC),
			[]types.EvalType{types.EvalInt64}, []types.EvalType{types.EvalInt64},
		},
		{
			"single(int,datetime)",
			NewProjectStage(NewDualStage(), intPC), NewProjectStage(NewDualStage(), datetimePC),
			[]types.EvalType{types.EvalInt64}, []types.EvalType{types.EvalInt64},
		},
		{
			"single(uint,int)",
			NewProjectStage(NewDualStage(), uintPC), NewProjectStage(NewDualStage(), intPC),
			[]types.EvalType{types.EvalUint64}, []types.EvalType{types.EvalInt64},
		},
		{
			"single(uint,uint)",
			NewProjectStage(NewDualStage(), uintPC), NewProjectStage(NewDualStage(), uintPC),
			[]types.EvalType{types.EvalUint64}, []types.EvalType{types.EvalUint64},
		},
		{
			"single(uint,float)",
			NewProjectStage(NewDualStage(), uintPC), NewProjectStage(NewDualStage(), floatPC),
			[]types.EvalType{types.EvalUint64}, []types.EvalType{types.EvalDouble},
		},
		{
			"single(uint,decimal)",
			NewProjectStage(NewDualStage(), uintPC), NewProjectStage(NewDualStage(), decimalPC),
			[]types.EvalType{types.EvalUint64}, []types.EvalType{types.EvalDecimal128},
		},
		{
			"single(uint,bool)",
			NewProjectStage(NewDualStage(), uintPC), NewProjectStage(NewDualStage(), boolPC),
			[]types.EvalType{types.EvalUint64}, []types.EvalType{types.EvalUint64},
		},
		{
			"single(uint,str)",
			NewProjectStage(NewDualStage(), uintPC), NewProjectStage(NewDualStage(), strPC),
			[]types.EvalType{types.EvalUint64}, []types.EvalType{types.EvalUint64},
		},
		{
			"single(uint,date)",
			NewProjectStage(NewDualStage(), uintPC), NewProjectStage(NewDualStage(), datePC),
			[]types.EvalType{types.EvalUint64}, []types.EvalType{types.EvalUint64},
		},
		{
			"single(uint,datetime)",
			NewProjectStage(NewDualStage(), uintPC), NewProjectStage(NewDualStage(), datetimePC),
			[]types.EvalType{types.EvalUint64}, []types.EvalType{types.EvalUint64},
		},
		{
			"single(float,int)",
			NewProjectStage(NewDualStage(), floatPC), NewProjectStage(NewDualStage(), intPC),
			[]types.EvalType{types.EvalDouble}, []types.EvalType{types.EvalInt64},
		},
		{
			"single(float,uint)",
			NewProjectStage(NewDualStage(), floatPC), NewProjectStage(NewDualStage(), uintPC),
			[]types.EvalType{types.EvalDouble}, []types.EvalType{types.EvalUint64},
		},
		{
			"single(float,float)",
			NewProjectStage(NewDualStage(), floatPC), NewProjectStage(NewDualStage(), floatPC),
			[]types.EvalType{types.EvalDouble}, []types.EvalType{types.EvalDouble},
		},
		{
			"single(float,decimal)",
			NewProjectStage(NewDualStage(), floatPC), NewProjectStage(NewDualStage(), decimalPC),
			[]types.EvalType{types.EvalDouble}, []types.EvalType{types.EvalDecimal128},
		},
		{
			"single(float,bool)",
			NewProjectStage(NewDualStage(), floatPC), NewProjectStage(NewDualStage(), boolPC),
			[]types.EvalType{types.EvalDouble}, []types.EvalType{types.EvalDouble},
		},
		{
			"single(float,str)",
			NewProjectStage(NewDualStage(), floatPC), NewProjectStage(NewDualStage(), strPC),
			[]types.EvalType{types.EvalDouble}, []types.EvalType{types.EvalDouble},
		},
		{
			"single(float,date)",
			NewProjectStage(NewDualStage(), floatPC), NewProjectStage(NewDualStage(), datePC),
			[]types.EvalType{types.EvalDouble}, []types.EvalType{types.EvalDouble},
		},
		{
			"single(float,datetime)",
			NewProjectStage(NewDualStage(), floatPC), NewProjectStage(NewDualStage(), datetimePC),
			[]types.EvalType{types.EvalDouble}, []types.EvalType{types.EvalDouble},
		},
		{
			"single(decimal,int)",
			NewProjectStage(NewDualStage(), decimalPC), NewProjectStage(NewDualStage(), intPC),
			[]types.EvalType{types.EvalDecimal128}, []types.EvalType{types.EvalInt64},
		},
		{
			"single(decimal,uint)",
			NewProjectStage(NewDualStage(), decimalPC), NewProjectStage(NewDualStage(), uintPC),
			[]types.EvalType{types.EvalDecimal128}, []types.EvalType{types.EvalUint64},
		},
		{
			"single(decimal,float)",
			NewProjectStage(NewDualStage(), decimalPC), NewProjectStage(NewDualStage(), floatPC),
			[]types.EvalType{types.EvalDecimal128}, []types.EvalType{types.EvalDouble},
		},
		{
			"single(decimal,decimal)",
			NewProjectStage(NewDualStage(), decimalPC), NewProjectStage(NewDualStage(), decimalPC),
			[]types.EvalType{types.EvalDecimal128}, []types.EvalType{types.EvalDecimal128},
		},
		{
			"single(decimal,bool)",
			NewProjectStage(NewDualStage(), decimalPC), NewProjectStage(NewDualStage(), boolPC),
			[]types.EvalType{types.EvalDecimal128}, []types.EvalType{types.EvalDecimal128},
		},
		{
			"single(decimal,str)",
			NewProjectStage(NewDualStage(), decimalPC), NewProjectStage(NewDualStage(), strPC),
			[]types.EvalType{types.EvalDecimal128}, []types.EvalType{types.EvalDecimal128},
		},
		{
			"single(decimal,date)",
			NewProjectStage(NewDualStage(), decimalPC), NewProjectStage(NewDualStage(), datePC),
			[]types.EvalType{types.EvalDecimal128}, []types.EvalType{types.EvalDecimal128},
		},
		{
			"single(decimal,datetime)",
			NewProjectStage(NewDualStage(), decimalPC), NewProjectStage(NewDualStage(), datetimePC),
			[]types.EvalType{types.EvalDecimal128}, []types.EvalType{types.EvalDecimal128},
		},
		{
			"single(bool,int)",
			NewProjectStage(NewDualStage(), boolPC), NewProjectStage(NewDualStage(), intPC),
			[]types.EvalType{types.EvalInt64}, []types.EvalType{types.EvalInt64},
		},
		{
			"single(bool,uint)",
			NewProjectStage(NewDualStage(), boolPC), NewProjectStage(NewDualStage(), uintPC),
			[]types.EvalType{types.EvalUint64}, []types.EvalType{types.EvalUint64},
		},
		{
			"single(bool,float)",
			NewProjectStage(NewDualStage(), boolPC), NewProjectStage(NewDualStage(), floatPC),
			[]types.EvalType{types.EvalDouble}, []types.EvalType{types.EvalDouble},
		},
		{
			"single(bool,decimal)",
			NewProjectStage(NewDualStage(), boolPC), NewProjectStage(NewDualStage(), decimalPC),
			[]types.EvalType{types.EvalDecimal128}, []types.EvalType{types.EvalDecimal128},
		},
		{
			"single(bool,bool)",
			NewProjectStage(NewDualStage(), boolPC), NewProjectStage(NewDualStage(), boolPC),
			[]types.EvalType{types.EvalBoolean}, []types.EvalType{types.EvalBoolean},
		},
		{
			"single(bool,str)",
			NewProjectStage(NewDualStage(), boolPC), NewProjectStage(NewDualStage(), strPC),
			[]types.EvalType{types.EvalString}, []types.EvalType{types.EvalString},
		},
		{
			"single(bool,date)",
			NewProjectStage(NewDualStage(), boolPC), NewProjectStage(NewDualStage(), datePC),
			[]types.EvalType{types.EvalDate}, []types.EvalType{types.EvalDate},
		},
		{
			"single(bool,datetime)",
			NewProjectStage(NewDualStage(), boolPC), NewProjectStage(NewDualStage(), datetimePC),
			[]types.EvalType{types.EvalDatetime}, []types.EvalType{types.EvalDatetime},
		},
		{
			"single(str,int)",
			NewProjectStage(NewDualStage(), strPC), NewProjectStage(NewDualStage(), intPC),
			[]types.EvalType{types.EvalInt64}, []types.EvalType{types.EvalInt64},
		},
		{
			"single(str,uint)",
			NewProjectStage(NewDualStage(), strPC), NewProjectStage(NewDualStage(), uintPC),
			[]types.EvalType{types.EvalUint64}, []types.EvalType{types.EvalUint64},
		},
		{
			"single(str,float)",
			NewProjectStage(NewDualStage(), strPC), NewProjectStage(NewDualStage(), floatPC),
			[]types.EvalType{types.EvalDouble}, []types.EvalType{types.EvalDouble},
		},
		{
			"single(str,decimal)",
			NewProjectStage(NewDualStage(), strPC), NewProjectStage(NewDualStage(), decimalPC),
			[]types.EvalType{types.EvalDecimal128}, []types.EvalType{types.EvalDecimal128},
		},
		{
			"single(str,bool)",
			NewProjectStage(NewDualStage(), strPC), NewProjectStage(NewDualStage(), boolPC),
			[]types.EvalType{types.EvalString}, []types.EvalType{types.EvalString},
		},
		{
			"single(str,str)",
			NewProjectStage(NewDualStage(), strPC), NewProjectStage(NewDualStage(), strPC),
			[]types.EvalType{types.EvalString}, []types.EvalType{types.EvalString},
		},
		{
			"single(str,date)",
			NewProjectStage(NewDualStage(), strPC), NewProjectStage(NewDualStage(), datePC),
			[]types.EvalType{types.EvalDate}, []types.EvalType{types.EvalDate},
		},
		{
			"single(str,datetime)",
			NewProjectStage(NewDualStage(), strPC), NewProjectStage(NewDualStage(), datetimePC),
			[]types.EvalType{types.EvalDatetime}, []types.EvalType{types.EvalDatetime},
		},
		{
			"single(date,int)",
			NewProjectStage(NewDualStage(), datePC), NewProjectStage(NewDualStage(), intPC),
			[]types.EvalType{types.EvalInt64}, []types.EvalType{types.EvalInt64},
		},
		{
			"single(date,uint)",
			NewProjectStage(NewDualStage(), datePC), NewProjectStage(NewDualStage(), uintPC),
			[]types.EvalType{types.EvalUint64}, []types.EvalType{types.EvalUint64},
		},
		{
			"single(date,float)",
			NewProjectStage(NewDualStage(), datePC), NewProjectStage(NewDualStage(), floatPC),
			[]types.EvalType{types.EvalDouble}, []types.EvalType{types.EvalDouble},
		},
		{
			"single(date,decimal)",
			NewProjectStage(NewDualStage(), datePC), NewProjectStage(NewDualStage(), decimalPC),
			[]types.EvalType{types.EvalDecimal128}, []types.EvalType{types.EvalDecimal128},
		},
		{
			"single(date,bool)",
			NewProjectStage(NewDualStage(), datePC), NewProjectStage(NewDualStage(), boolPC),
			[]types.EvalType{types.EvalDate}, []types.EvalType{types.EvalDate},
		},
		{
			"single(date,str)",
			NewProjectStage(NewDualStage(), datePC), NewProjectStage(NewDualStage(), strPC),
			[]types.EvalType{types.EvalDate}, []types.EvalType{types.EvalDate},
		},
		{
			"single(date,date)",
			NewProjectStage(NewDualStage(), datePC), NewProjectStage(NewDualStage(), datePC),
			[]types.EvalType{types.EvalDate}, []types.EvalType{types.EvalDate},
		},
		{
			"single(date,datetime)",
			NewProjectStage(NewDualStage(), datePC), NewProjectStage(NewDualStage(), datetimePC),
			[]types.EvalType{types.EvalDatetime}, []types.EvalType{types.EvalDatetime},
		},
		{
			"single(datetime,int)",
			NewProjectStage(NewDualStage(), datetimePC), NewProjectStage(NewDualStage(), intPC),
			[]types.EvalType{types.EvalInt64}, []types.EvalType{types.EvalInt64},
		},
		{
			"single(datetime,uint)",
			NewProjectStage(NewDualStage(), datetimePC), NewProjectStage(NewDualStage(), uintPC),
			[]types.EvalType{types.EvalUint64}, []types.EvalType{types.EvalUint64},
		},
		{
			"single(datetime,float)",
			NewProjectStage(NewDualStage(), datetimePC), NewProjectStage(NewDualStage(), floatPC),
			[]types.EvalType{types.EvalDouble}, []types.EvalType{types.EvalDouble},
		},
		{
			"single(datetime,decimal)",
			NewProjectStage(NewDualStage(), datetimePC), NewProjectStage(NewDualStage(), decimalPC),
			[]types.EvalType{types.EvalDecimal128}, []types.EvalType{types.EvalDecimal128},
		},
		{
			"single(datetime,bool)",
			NewProjectStage(NewDualStage(), datetimePC), NewProjectStage(NewDualStage(), boolPC),
			[]types.EvalType{types.EvalDatetime}, []types.EvalType{types.EvalDatetime},
		},
		{
			"single(datetime,str)",
			NewProjectStage(NewDualStage(), datetimePC), NewProjectStage(NewDualStage(), strPC),
			[]types.EvalType{types.EvalDatetime}, []types.EvalType{types.EvalDatetime},
		},
		{
			"single(datetime,date)",
			NewProjectStage(NewDualStage(), datetimePC), NewProjectStage(NewDualStage(), datePC),
			[]types.EvalType{types.EvalDatetime}, []types.EvalType{types.EvalDatetime},
		},
		{
			"single(datetime,datetime)",
			NewProjectStage(NewDualStage(), datetimePC), NewProjectStage(NewDualStage(), datetimePC),
			[]types.EvalType{types.EvalDatetime}, []types.EvalType{types.EvalDatetime},
		},
		{
			"multiple all pairs same type",
			NewProjectStage(NewDualStage(), intPC, boolPC, datePC),
			NewProjectStage(NewDualStage(), intPC, boolPC, datePC),
			[]types.EvalType{types.EvalInt64, types.EvalBoolean, types.EvalDate},
			[]types.EvalType{types.EvalInt64, types.EvalBoolean, types.EvalDate},
		},
		{
			"multiple all pairs similar type",
			NewProjectStage(NewDualStage(), intPC, uintPC, floatPC),
			NewProjectStage(NewDualStage(), decimalPC, intPC, uintPC),
			[]types.EvalType{types.EvalInt64, types.EvalUint64, types.EvalDouble},
			[]types.EvalType{types.EvalDecimal128, types.EvalInt64, types.EvalUint64},
		},
		{
			"multiple some pairs similar type",
			NewProjectStage(NewDualStage(), intPC, boolPC, floatPC),
			NewProjectStage(NewDualStage(), decimalPC, intPC, strPC),
			[]types.EvalType{types.EvalInt64, types.EvalInt64, types.EvalDouble},
			[]types.EvalType{types.EvalDecimal128, types.EvalInt64, types.EvalDouble},
		},
		{
			"multiple no pairs similar type",
			NewProjectStage(NewDualStage(), intPC, boolPC, floatPC),
			NewProjectStage(NewDualStage(), strPC, intPC, datePC),
			[]types.EvalType{types.EvalInt64, types.EvalInt64, types.EvalDouble},
			[]types.EvalType{types.EvalInt64, types.EvalInt64, types.EvalDouble},
		},
	}

	runTests(t, tests)
}

func TestConvertExprs(t *testing.T) {
	type test struct {
		name          string
		exprs         []SQLExpr
		targetTypes   []types.EvalType
		expectedTypes []types.EvalType
	}

	knd := values.MySQLValueKind

	intVal := NewSQLValueExpr(values.NewSQLInt64(knd, 1))
	uintVal := NewSQLValueExpr(values.NewSQLUint64(knd, 1))
	floatVal := NewSQLValueExpr(values.NewSQLFloat(knd, 1))
	decimalVal := NewSQLValueExpr(values.NewSQLDecimal128(knd, decimal.NewFromFloat(1.0)))
	boolVal := NewSQLValueExpr(values.NewSQLBool(knd, true))
	strVal := NewSQLValueExpr(values.NewSQLVarchar(knd, "bar"))
	dateVal := NewSQLValueExpr(values.NewSQLDate(knd, time.Now()))
	datetimeVal := NewSQLValueExpr(values.NewSQLTimestamp(knd, time.Now()))

	allTypes := []types.EvalType{types.EvalPolymorphic, types.EvalInt64, types.EvalUint64, types.EvalDouble, types.EvalDecimal128, types.EvalBoolean, types.EvalString, types.EvalDate, types.EvalDatetime}

	makeValSlice := func(val SQLExpr) []SQLExpr {
		s := make([]SQLExpr, len(allTypes))
		for i := range s {
			s[i] = val
		}
		return s
	}

	tests := []test{
		// target EvalPolymorphic => no conversion
		// target similar type => no conversion
		// target other type => convert (unconditionally)
		{
			"convertExprs(int)", makeValSlice(intVal), allTypes,
			[]types.EvalType{types.EvalInt64, types.EvalInt64, types.EvalInt64, types.EvalInt64, types.EvalInt64, types.EvalBoolean, types.EvalString, types.EvalDate, types.EvalDatetime},
		},
		{
			"convertExprs(uint)", makeValSlice(uintVal), allTypes,
			[]types.EvalType{types.EvalUint64, types.EvalUint64, types.EvalUint64, types.EvalUint64, types.EvalUint64, types.EvalBoolean, types.EvalString, types.EvalDate, types.EvalDatetime},
		},
		{
			"convertExprs(float)", makeValSlice(floatVal), allTypes,
			[]types.EvalType{types.EvalDouble, types.EvalDouble, types.EvalDouble, types.EvalDouble, types.EvalDouble, types.EvalBoolean, types.EvalString, types.EvalDate, types.EvalDatetime},
		},
		{
			"convertExprs(decimal)", makeValSlice(decimalVal), allTypes,
			[]types.EvalType{types.EvalDecimal128, types.EvalDecimal128, types.EvalDecimal128, types.EvalDecimal128, types.EvalDecimal128, types.EvalBoolean, types.EvalString, types.EvalDate, types.EvalDatetime},
		},
		{
			"convertExprs(bool)", makeValSlice(boolVal), allTypes,
			[]types.EvalType{types.EvalBoolean, types.EvalInt64, types.EvalUint64, types.EvalDouble, types.EvalDecimal128, types.EvalBoolean, types.EvalString, types.EvalDate, types.EvalDatetime},
		},
		{
			"convertExprs(str)", makeValSlice(strVal), allTypes,
			[]types.EvalType{types.EvalString, types.EvalInt64, types.EvalUint64, types.EvalDouble, types.EvalDecimal128, types.EvalBoolean, types.EvalString, types.EvalDate, types.EvalDatetime},
		},
		{
			"convertExprs(date)", makeValSlice(dateVal), allTypes,
			[]types.EvalType{types.EvalDate, types.EvalInt64, types.EvalUint64, types.EvalDouble, types.EvalDecimal128, types.EvalBoolean, types.EvalString, types.EvalDate, types.EvalDatetime},
		},
		{
			"convertExprs(datetime)", makeValSlice(datetimeVal), allTypes,
			[]types.EvalType{types.EvalDatetime, types.EvalInt64, types.EvalUint64, types.EvalDouble, types.EvalDecimal128, types.EvalBoolean, types.EvalString, types.EvalDate, types.EvalDatetime},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			convertedExprs := convertExprs(test.exprs, test.targetTypes)

			for i, ce := range convertedExprs {
				req.Equal(test.expectedTypes[i], ce.EvalType(), "incorrect EvalType for child %d", i)
			}
		})
	}
}
