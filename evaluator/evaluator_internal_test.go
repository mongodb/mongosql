package evaluator

import (
	"testing"
	"time"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/astutil"
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
		pipeline: ast.NewPipeline(
			ast.NewProjectStage(ast.NewIncludeProjectItem(ast.NewFieldRef("foo", nil))),
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
	req.True(projectStageExcludesID(ms.pipeline.Stages[0].(*ast.ProjectStage)),
		"_id:0 must be added to project")

	mongoSourceStage1 := &MongoSourceStage{
		selectIDs:  []int{0},
		dbName:     "foo",
		tableNames: []string{"bar"},
		aliasNames: []string{"biz"},
		tableType:  "view",
		pipeline: ast.NewPipeline(
			ast.NewProjectStage(ast.NewIncludeProjectItem(ast.NewFieldRef("foo", nil))),
		),
	}
	mongoSourceStage2 := &MongoSourceStage{
		selectIDs:  []int{0},
		dbName:     "foo",
		tableNames: []string{"bar"},
		aliasNames: []string{"biz"},
		tableType:  "view",
		pipeline: ast.NewPipeline(
			ast.NewProjectStage(ast.NewAssignProjectItem("bar", astutil.Int32Constant(2))),
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
	req.True(projectStageExcludesID(ms1.pipeline.Stages[0].(*ast.ProjectStage)),
		"_id:0 must be added to left stage project")
	ms2, ok := us.right.(*MongoSourceStage)
	req.True(ok)
	req.True(projectStageExcludesID(ms2.pipeline.Stages[0].(*ast.ProjectStage)),
		"_id:0 must be added to right stage project")
}

func projectStageExcludesID(project *ast.ProjectStage) bool {
	for field := range project.ExcludeItems() {
		if field == "_id" {
			return true
		}
	}

	return false
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
		expectedBody        []*ast.AddFieldsItem
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
			expectedBody:        []*ast.AddFieldsItem{},
			expectedHasEmbedded: false},

		{inputFields: noConflictEmbeddedFields,
			inputEvalType:  types.EvalInt64,
			inputIs34:      true,
			expectedFields: expectedNoConflictEmbeddedFields,
			expectedBody: []*ast.AddFieldsItem{
				ast.NewAddFieldsItem("c_DOT_a", astutil.FieldRefFromFieldName("c.a")),
				ast.NewAddFieldsItem("c_DOT_d", astutil.FieldRefFromFieldName("c.d")),
			},
			expectedHasEmbedded: true},

		{inputFields: conflictedEmbeddedFields,
			inputEvalType:  types.EvalInt64,
			inputIs34:      true,
			expectedFields: expectedConflictedEmbeddedFields,
			expectedBody: []*ast.AddFieldsItem{
				ast.NewAddFieldsItem("a_DOT_b0", astutil.FieldRefFromFieldName("a.b")),
				ast.NewAddFieldsItem("a_DOT_c1", astutil.FieldRefFromFieldName("a.c")),
			},
			expectedHasEmbedded: true},

		//tests for pre-3.4+ which should generate project bodies
		{inputFields: nonEmbeddedFields,
			inputEvalType:       types.EvalInt64,
			inputIs34:           false,
			expectedFields:      nonEmbeddedFields,
			expectedBody:        []*ast.AddFieldsItem{},
			expectedHasEmbedded: false},

		{inputFields: noConflictEmbeddedFields32,
			inputEvalType:  types.EvalInt64,
			inputIs34:      false,
			expectedFields: expectedNoConflictEmbeddedFields,
			expectedBody: []*ast.AddFieldsItem{
				ast.NewAddFieldsItem("a", astutil.TrueLiteral),
				ast.NewAddFieldsItem("b", astutil.TrueLiteral),
				ast.NewAddFieldsItem("c_DOT_a", astutil.FieldRefFromFieldName("c.a")),
				ast.NewAddFieldsItem("c_DOT_d", astutil.FieldRefFromFieldName("c.d")),
			},
			expectedHasEmbedded: true},

		{inputFields: conflictedEmbeddedFields32,
			inputEvalType:  types.EvalInt64,
			inputIs34:      false,
			expectedFields: expectedConflictedEmbeddedFields,
			expectedBody: []*ast.AddFieldsItem{
				ast.NewAddFieldsItem("a_DOT_b", astutil.TrueLiteral),
				ast.NewAddFieldsItem("a_DOT_c", astutil.TrueLiteral),
				ast.NewAddFieldsItem("a_DOT_c0", astutil.TrueLiteral),
				ast.NewAddFieldsItem("a_DOT_b0", astutil.FieldRefFromFieldName("a.b")),
				ast.NewAddFieldsItem("a_DOT_c1", astutil.FieldRefFromFieldName("a.c")),
				ast.NewAddFieldsItem("b", astutil.TrueLiteral),
			},
			expectedHasEmbedded: true},

		// tests for 3.4+ which should generate addFields bodies,
		// with EvalArrNumeric type for some of the fields.
		{inputFields: noConflictEmbeddedFieldsArr,
			inputEvalType:  types.EvalArrNumeric,
			inputIs34:      true,
			expectedFields: expectedNoConflictEmbeddedFieldsArr,
			expectedBody: []*ast.AddFieldsItem{
				ast.NewAddFieldsItem("c_DOT_a_DOT_1", astutil.WrapInOp(bsonutil.OpArrElemAt,
					astutil.FieldRefFromFieldName("c.a"),
					astutil.Int64Value(1),
				)),
				ast.NewAddFieldsItem("c_DOT_d", astutil.FieldRefFromFieldName("c.d")),
			},
			expectedHasEmbedded: true},

		{inputFields: conflictedEmbeddedFieldsArr,
			inputEvalType:  types.EvalArrNumeric,
			inputIs34:      true,
			expectedFields: expectedConflictedEmbeddedFieldsArr,
			expectedBody: []*ast.AddFieldsItem{
				ast.NewAddFieldsItem("a_DOT_c_DOT_1", astutil.WrapInOp(bsonutil.OpArrElemAt,
					ast.NewFieldRef("a_DOT_c", nil),
					astutil.Int64Value(1),
				)),
				ast.NewAddFieldsItem("a_DOT_b0", astutil.FieldRefFromFieldName("a.b")),
				ast.NewAddFieldsItem("a_DOT_c", astutil.FieldRefFromFieldName("a.c")),
			},
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
	intCol := results.NewColumn(1, "", "", "", "", "", "", types.EvalInt64, schema.MongoInt, false, true)
	strCol := results.NewColumn(1, "", "", "", "", "", "", types.EvalString, schema.MongoString, false, true)

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
	escapeVal := NewSQLValueExpr(values.NewSQLVarchar(knd, "\\")) // This is being used in the SQL like expression reconcile tests as the escape character.
	dateVal := NewSQLValueExpr(values.NewSQLDate(knd, time.Now()))
	datetimeVal := NewSQLValueExpr(values.NewSQLTimestamp(knd, time.Now()))
	boolColVal := NewSQLColumnExpr(0, "", "", "", types.EvalBoolean, schema.MongoBool, false, true)

	allVals := []SQLExpr{intVal, uintVal, floatVal, decimalVal, boolVal, strVal, dateVal, datetimeVal}
	allTypes := []types.EvalType{types.EvalInt64, types.EvalUint64, types.EvalDouble, types.EvalDecimal128, types.EvalBoolean, types.EvalString, types.EvalDate, types.EvalDatetime}

	makeTypeSlice := func(typ types.EvalType) []types.EvalType {
		s := make([]types.EvalType, len(allVals))
		for i := range s {
			s[i] = typ
		}
		return s
	}
	// Type order: polymorphic < objectid < bool < string < date < datetime < int < uint < double < decimal

	// tests is a list of test cases for any SQLExprs with custom reconcile
	// implementations. SQLExprs whose reconcile functions delegate to the
	// reconcileArithmetic, reconcileComparison, reconcileSubqueryPlans, or
	// convertExprs functions are tested separately.

	tests := []test{

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
		// equals special case: a boolean column expr and a 1 or 0 (number) result in bool conversions.
		{"eq(bool column,1)", NewSQLComparisonExpr(EQ, boolColVal, intVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"eq(1,bool column)", NewSQLComparisonExpr(EQ, intVal, boolColVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"eq(bool column,0)", NewSQLComparisonExpr(EQ, boolColVal, NewSQLValueExpr(values.NewSQLInt64(knd, 0))), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"eq(0,bool column)", NewSQLComparisonExpr(EQ, NewSQLValueExpr(values.NewSQLInt64(knd, 0)), boolColVal), []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"eq(bool column,2)", NewSQLComparisonExpr(EQ, boolColVal, NewSQLValueExpr(values.NewSQLInt64(knd, 2))), []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"eq(2,bool column)", NewSQLComparisonExpr(EQ, NewSQLValueExpr(values.NewSQLInt64(knd, 2)), boolColVal), []types.EvalType{types.EvalInt64, types.EvalInt64}},

		// like expression
		{"like(str, int)", NewSQLLikeExpr(strVal, intVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(str, uint)", NewSQLLikeExpr(strVal, uintVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(str, float)", NewSQLLikeExpr(strVal, floatVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(str, decimal)", NewSQLLikeExpr(strVal, decimalVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(str, bool)", NewSQLLikeExpr(strVal, boolVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(str, str)", NewSQLLikeExpr(strVal, strVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(str, date)", NewSQLLikeExpr(strVal, dateVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(str, datetime)", NewSQLLikeExpr(strVal, datetimeVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(int, int)", NewSQLLikeExpr(intVal, intVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(int, uint)", NewSQLLikeExpr(intVal, uintVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(int, float)", NewSQLLikeExpr(intVal, floatVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(int, decimal)", NewSQLLikeExpr(intVal, decimalVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(int, bool)", NewSQLLikeExpr(intVal, boolVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(int, str)", NewSQLLikeExpr(intVal, strVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(int, date)", NewSQLLikeExpr(intVal, dateVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(int, datetime)", NewSQLLikeExpr(intVal, datetimeVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(uint, int)", NewSQLLikeExpr(uintVal, intVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(uint, uint)", NewSQLLikeExpr(uintVal, uintVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(uint, float)", NewSQLLikeExpr(uintVal, floatVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(uint, decimal)", NewSQLLikeExpr(uintVal, decimalVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(uint, bool)", NewSQLLikeExpr(uintVal, boolVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(uint, str)", NewSQLLikeExpr(uintVal, strVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(uint, date)", NewSQLLikeExpr(uintVal, dateVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(uint, datetime)", NewSQLLikeExpr(uintVal, intVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(float, int)", NewSQLLikeExpr(floatVal, intVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(float, uint)", NewSQLLikeExpr(floatVal, uintVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(float, float)", NewSQLLikeExpr(floatVal, floatVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(float, decimal)", NewSQLLikeExpr(floatVal, decimalVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(float, bool)", NewSQLLikeExpr(floatVal, boolVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(float, str)", NewSQLLikeExpr(floatVal, strVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(float, date)", NewSQLLikeExpr(floatVal, dateVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(float, datetime)", NewSQLLikeExpr(floatVal, datetimeVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(decimal, int)", NewSQLLikeExpr(decimalVal, intVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(decimal, uint)", NewSQLLikeExpr(decimalVal, uintVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(decimal, float)", NewSQLLikeExpr(decimalVal, floatVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(decimal, decimal)", NewSQLLikeExpr(decimalVal, decimalVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(decimal, bool)", NewSQLLikeExpr(decimalVal, boolVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(decimal, str)", NewSQLLikeExpr(decimalVal, strVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(decimal, date)", NewSQLLikeExpr(decimalVal, dateVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(decimal, datetime)", NewSQLLikeExpr(decimalVal, datetimeVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(bool, int)", NewSQLLikeExpr(boolVal, intVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(bool, uint)", NewSQLLikeExpr(boolVal, uintVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(bool, float)", NewSQLLikeExpr(boolVal, floatVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(bool, decimal)", NewSQLLikeExpr(boolVal, decimalVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(bool, bool)", NewSQLLikeExpr(boolVal, boolVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(bool, str)", NewSQLLikeExpr(boolVal, strVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(bool, date)", NewSQLLikeExpr(boolVal, dateVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(bool, datetime)", NewSQLLikeExpr(boolVal, datetimeVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(date, int)", NewSQLLikeExpr(dateVal, intVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(date, uint)", NewSQLLikeExpr(dateVal, uintVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(date, float)", NewSQLLikeExpr(dateVal, floatVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(date, decimal)", NewSQLLikeExpr(dateVal, decimalVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(date, bool)", NewSQLLikeExpr(dateVal, boolVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(date, str)", NewSQLLikeExpr(dateVal, strVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(date, date)", NewSQLLikeExpr(dateVal, dateVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(date, datetime)", NewSQLLikeExpr(dateVal, datetimeVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(datetime, int)", NewSQLLikeExpr(datetimeVal, intVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(datetime, uint)", NewSQLLikeExpr(datetimeVal, uintVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(datetime, float)", NewSQLLikeExpr(datetimeVal, floatVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(datetime, decimal)", NewSQLLikeExpr(datetimeVal, decimalVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(datetime, bool)", NewSQLLikeExpr(datetimeVal, boolVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(datetime, str)", NewSQLLikeExpr(datetimeVal, strVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(datetime, date)", NewSQLLikeExpr(datetimeVal, dateVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},
		{"like(datetime, datetime)", NewSQLLikeExpr(datetimeVal, datetimeVal, escapeVal, true), []types.EvalType{types.EvalString, types.EvalString, types.EvalString}},

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

		// agg functions: most of the aggregation functions' reconcile methods are no-ops, so there are no conversions.
		{"avg", NewSQLAggregationFunctionExpr(parser.AvgAggregateName, false, allVals), allTypes},
		{"count", NewSQLAggregationFunctionExpr(parser.CountAggregateName, false, allVals), allTypes},
		{"max", NewSQLAggregationFunctionExpr(parser.MaxAggregateName, false, allVals), allTypes},
		{"min", NewSQLAggregationFunctionExpr(parser.MinAggregateName, false, allVals), allTypes},
		{"sum", NewSQLAggregationFunctionExpr(parser.SumAggregateName, false, allVals), allTypes},
		{"stdDev", NewSQLAggregationFunctionExpr(parser.StdDevAggregateName, false, allVals), allTypes},
		{"stdDevSample", NewSQLAggregationFunctionExpr(parser.StdDevSampleAggregateName, false, allVals), allTypes},
		// group_concat operates on strings, so it converts all other types to strings.
		{"groupConcat", NewSQLAggregationFunctionExpr(parser.GroupConcatAggregateName, false, allVals),
			makeTypeSlice(types.EvalString)},
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
	dateVal := NewSQLValueExpr(values.NewSQLDate(knd, time.Now()))
	datetimeVal := NewSQLValueExpr(values.NewSQLTimestamp(knd, time.Now()))
	objectIDVal := NewSQLValueExpr(values.NewSQLObjectID(knd, "000000000000000000000000"))

	strVal := NewSQLValueExpr(values.NewSQLVarchar(knd, "bar"))
	strColExpr := NewSQLColumnExpr(0, "db", "test", "s", types.EvalString, schema.MongoString, false, false)

	intCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalInt64, schema.MongoInt, false, true)
	uintCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalUint64, schema.MongoInt, false, true)
	floatCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalDouble, schema.MongoInt, false, true)
	decimalCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalDecimal128, schema.MongoInt, false, true)
	boolCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalBoolean, schema.MongoInt, false, true)
	strCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalString, schema.MongoInt, false, true)
	dateCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalDate, schema.MongoInt, false, true)
	datetimeCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalDatetime, schema.MongoInt, false, true)
	objectIDCol := results.NewColumn(0, "", "", "", "", "", "", types.EvalObjectID, schema.MongoInt, false, true)

	intPC := ProjectedColumn{intCol, intVal}
	uintPC := ProjectedColumn{uintCol, uintVal}
	floatPC := ProjectedColumn{floatCol, floatVal}
	decimalPC := ProjectedColumn{decimalCol, decimalVal}
	boolPC := ProjectedColumn{boolCol, boolVal}
	datePC := ProjectedColumn{dateCol, dateVal}
	datetimePC := ProjectedColumn{datetimeCol, datetimeVal}
	objectIDPC := ProjectedColumn{objectIDCol, objectIDVal}

	strColPC := ProjectedColumn{strCol, strColExpr}
	strValPC := ProjectedColumn{strCol, strVal}

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
			NewProjectStage(NewDualStage(), intPC), NewProjectStage(NewDualStage(), strValPC),
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
			"single(int,objectid)",
			NewProjectStage(NewDualStage(), intPC), NewProjectStage(NewDualStage(), objectIDPC),
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
			NewProjectStage(NewDualStage(), uintPC), NewProjectStage(NewDualStage(), strValPC),
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
			"single(uint,objectid)",
			NewProjectStage(NewDualStage(), uintPC), NewProjectStage(NewDualStage(), objectIDPC),
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
			NewProjectStage(NewDualStage(), floatPC), NewProjectStage(NewDualStage(), strValPC),
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
			"single(float,objectid)",
			NewProjectStage(NewDualStage(), floatPC), NewProjectStage(NewDualStage(), objectIDPC),
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
			NewProjectStage(NewDualStage(), decimalPC), NewProjectStage(NewDualStage(), strValPC),
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
			"single(decimal,objectid)",
			NewProjectStage(NewDualStage(), decimalPC), NewProjectStage(NewDualStage(), objectIDPC),
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
			NewProjectStage(NewDualStage(), boolPC), NewProjectStage(NewDualStage(), strValPC),
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
			"single(bool,objectid)",
			NewProjectStage(NewDualStage(), boolPC), NewProjectStage(NewDualStage(), objectIDPC),
			[]types.EvalType{types.EvalBoolean}, []types.EvalType{types.EvalBoolean},
		},
		{
			"single(str,int)",
			NewProjectStage(NewDualStage(), strValPC), NewProjectStage(NewDualStage(), intPC),
			[]types.EvalType{types.EvalInt64}, []types.EvalType{types.EvalInt64},
		},
		{
			"single(str,uint)",
			NewProjectStage(NewDualStage(), strValPC), NewProjectStage(NewDualStage(), uintPC),
			[]types.EvalType{types.EvalUint64}, []types.EvalType{types.EvalUint64},
		},
		{
			"single(str,float)",
			NewProjectStage(NewDualStage(), strValPC), NewProjectStage(NewDualStage(), floatPC),
			[]types.EvalType{types.EvalDouble}, []types.EvalType{types.EvalDouble},
		},
		{
			"single(str,decimal)",
			NewProjectStage(NewDualStage(), strValPC), NewProjectStage(NewDualStage(), decimalPC),
			[]types.EvalType{types.EvalDecimal128}, []types.EvalType{types.EvalDecimal128},
		},
		{
			"single(str,bool)",
			NewProjectStage(NewDualStage(), strValPC), NewProjectStage(NewDualStage(), boolPC),
			[]types.EvalType{types.EvalString}, []types.EvalType{types.EvalString},
		},
		{
			"single(str,str)",
			NewProjectStage(NewDualStage(), strValPC), NewProjectStage(NewDualStage(), strValPC),
			[]types.EvalType{types.EvalString}, []types.EvalType{types.EvalString},
		},
		{
			"single(str,date)",
			NewProjectStage(NewDualStage(), strValPC), NewProjectStage(NewDualStage(), datePC),
			[]types.EvalType{types.EvalDate}, []types.EvalType{types.EvalDate},
		},
		{
			"single(str,datetime)",
			NewProjectStage(NewDualStage(), strValPC), NewProjectStage(NewDualStage(), datetimePC),
			[]types.EvalType{types.EvalDatetime}, []types.EvalType{types.EvalDatetime},
		},
		// str + objectid resolve to a objectids when the str is a literal (SQLValueExpr).
		{
			"single(strVal,objectid)",
			NewProjectStage(NewDualStage(), strValPC), NewProjectStage(NewDualStage(), objectIDPC),
			[]types.EvalType{types.EvalObjectID}, []types.EvalType{types.EvalObjectID},
		},
		// otherwise, resolves to a string.
		{
			"single(strCol,objectid)",
			NewProjectStage(NewDualStage(), strColPC), NewProjectStage(NewDualStage(), objectIDPC),
			[]types.EvalType{types.EvalString}, []types.EvalType{types.EvalString},
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
			NewProjectStage(NewDualStage(), datePC), NewProjectStage(NewDualStage(), strValPC),
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
			"single(date,objectid)",
			NewProjectStage(NewDualStage(), datePC), NewProjectStage(NewDualStage(), objectIDPC),
			[]types.EvalType{types.EvalDate}, []types.EvalType{types.EvalDate},
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
			NewProjectStage(NewDualStage(), datetimePC), NewProjectStage(NewDualStage(), strValPC),
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
			"single(datetime,objectid)",
			NewProjectStage(NewDualStage(), datetimePC), NewProjectStage(NewDualStage(), objectIDPC),
			[]types.EvalType{types.EvalDatetime}, []types.EvalType{types.EvalDatetime},
		},
		{
			"single(objectid,int)",
			NewProjectStage(NewDualStage(), objectIDPC), NewProjectStage(NewDualStage(), intPC),
			[]types.EvalType{types.EvalInt64}, []types.EvalType{types.EvalInt64},
		},
		{
			"single(objectid,uint)",
			NewProjectStage(NewDualStage(), objectIDPC), NewProjectStage(NewDualStage(), uintPC),
			[]types.EvalType{types.EvalUint64}, []types.EvalType{types.EvalUint64},
		},
		{
			"single(objectid,float)",
			NewProjectStage(NewDualStage(), objectIDPC), NewProjectStage(NewDualStage(), floatPC),
			[]types.EvalType{types.EvalDouble}, []types.EvalType{types.EvalDouble},
		},
		{
			"single(objectid,decimal)",
			NewProjectStage(NewDualStage(), objectIDPC), NewProjectStage(NewDualStage(), decimalPC),
			[]types.EvalType{types.EvalDecimal128}, []types.EvalType{types.EvalDecimal128},
		},
		{
			"single(objectid,bool)",
			NewProjectStage(NewDualStage(), objectIDPC), NewProjectStage(NewDualStage(), boolPC),
			[]types.EvalType{types.EvalBoolean}, []types.EvalType{types.EvalBoolean},
		},
		// str + objectid resolve to a objectids when the str is a literal (SQLValueExpr).
		{
			"single(objectid,strVal)",
			NewProjectStage(NewDualStage(), objectIDPC), NewProjectStage(NewDualStage(), strValPC),
			[]types.EvalType{types.EvalObjectID}, []types.EvalType{types.EvalObjectID},
		},
		// otherwise, resolves to a string.
		{
			"single(objectid,strCol)",
			NewProjectStage(NewDualStage(), objectIDPC), NewProjectStage(NewDualStage(), strColPC),
			[]types.EvalType{types.EvalString}, []types.EvalType{types.EvalString},
		},
		{
			"single(objectid,date)",
			NewProjectStage(NewDualStage(), objectIDPC), NewProjectStage(NewDualStage(), datePC),
			[]types.EvalType{types.EvalDate}, []types.EvalType{types.EvalDate},
		},
		{
			"single(objectid,datetime)",
			NewProjectStage(NewDualStage(), objectIDPC), NewProjectStage(NewDualStage(), datetimePC),
			[]types.EvalType{types.EvalDatetime}, []types.EvalType{types.EvalDatetime},
		},
		{
			"single(objectid,objectid)",
			NewProjectStage(NewDualStage(), objectIDPC), NewProjectStage(NewDualStage(), objectIDPC),
			[]types.EvalType{types.EvalObjectID}, []types.EvalType{types.EvalObjectID},
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
			NewProjectStage(NewDualStage(), decimalPC, intPC, strValPC),
			[]types.EvalType{types.EvalInt64, types.EvalInt64, types.EvalDouble},
			[]types.EvalType{types.EvalDecimal128, types.EvalInt64, types.EvalDouble},
		},
		{
			"multiple no pairs similar type",
			NewProjectStage(NewDualStage(), intPC, boolPC, floatPC),
			NewProjectStage(NewDualStage(), strValPC, intPC, datePC),
			[]types.EvalType{types.EvalInt64, types.EvalInt64, types.EvalDouble},
			[]types.EvalType{types.EvalInt64, types.EvalInt64, types.EvalDouble},
		},
	}

	runTests(t, tests)
}

func TestReconcileArithmetic(t *testing.T) {
	// Arguments to arithmetic expressions are converted using the following
	// rules: do not convert numeric types; do convert non-numeric types to
	// Decimal128 if either argument is a Datetime, non-numeric types to Int64
	// if either argument is a Date, and other non-numeric types to Float.

	type test struct {
		name          string
		input         []SQLExpr
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
	objectIDVal := NewSQLValueExpr(values.NewSQLObjectID(knd, "000000000000000000000000"))

	tests := []test{
		{"int", []SQLExpr{intVal}, []types.EvalType{types.EvalInt64}},
		{"uint", []SQLExpr{uintVal}, []types.EvalType{types.EvalUint64}},
		{"float", []SQLExpr{floatVal}, []types.EvalType{types.EvalDouble}},
		{"decimal", []SQLExpr{decimalVal}, []types.EvalType{types.EvalDecimal128}},
		{"bool", []SQLExpr{boolVal}, []types.EvalType{types.EvalInt64}},
		{"str", []SQLExpr{strVal}, []types.EvalType{types.EvalDouble}},
		{"date", []SQLExpr{dateVal}, []types.EvalType{types.EvalInt64}},
		{"datetime", []SQLExpr{datetimeVal}, []types.EvalType{types.EvalDecimal128}},
		{"objectid", []SQLExpr{objectIDVal}, []types.EvalType{types.EvalDouble}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			children := reconcileArithmetic(test.input)
			childType := children[0].EvalType()

			req.Equal(test.expectedTypes[0], childType, "incorrect EvalType: got %v, expected %v", types.EvalTypeToSQLType(childType), types.EvalTypeToSQLType(test.expectedTypes[0]))
		})
	}
}

func TestReconcileComparison(t *testing.T) {
	// Arguments to comparison expressions are converted using the
	// following rules: do not convert types if they are similar; do
	// convert types that are not similar to the higher precedence type
	// unless the operator is associative, in which case all types
	// should be converted to float.

	type test struct {
		name          string
		binaryNode    sqlBinaryNode
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
	objectIDVal := NewSQLValueExpr(values.NewSQLObjectID(knd, "000000000000000000000000"))

	strCol := NewSQLColumnExpr(0, "db", "test", "s", types.EvalString, schema.MongoString, false, false)

	tests := []test{
		{"sqlBinaryNode(int,int)", sqlBinaryNode{intVal, intVal}, []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sqlBinaryNode(int,uint)", sqlBinaryNode{intVal, uintVal}, []types.EvalType{types.EvalInt64, types.EvalUint64}},
		{"sqlBinaryNode(int,float)", sqlBinaryNode{intVal, floatVal}, []types.EvalType{types.EvalInt64, types.EvalDouble}},
		{"sqlBinaryNode(int,decimal)", sqlBinaryNode{intVal, decimalVal}, []types.EvalType{types.EvalInt64, types.EvalDecimal128}},
		{"sqlBinaryNode(int,bool)", sqlBinaryNode{intVal, boolVal}, []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sqlBinaryNode(int,str)", sqlBinaryNode{intVal, strVal}, []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sqlBinaryNode(int,date)", sqlBinaryNode{intVal, dateVal}, []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sqlBinaryNode(int,datetime)", sqlBinaryNode{intVal, datetimeVal}, []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sqlBinaryNode(int,objectid)", sqlBinaryNode{intVal, objectIDVal}, []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sqlBinaryNode(uint,int)", sqlBinaryNode{uintVal, intVal}, []types.EvalType{types.EvalUint64, types.EvalInt64}},
		{"sqlBinaryNode(uint,uint)", sqlBinaryNode{uintVal, uintVal}, []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sqlBinaryNode(uint,float)", sqlBinaryNode{uintVal, floatVal}, []types.EvalType{types.EvalUint64, types.EvalDouble}},
		{"sqlBinaryNode(uint,decimal)", sqlBinaryNode{uintVal, decimalVal}, []types.EvalType{types.EvalUint64, types.EvalDecimal128}},
		{"sqlBinaryNode(uint,bool)", sqlBinaryNode{uintVal, boolVal}, []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sqlBinaryNode(uint,str)", sqlBinaryNode{uintVal, strVal}, []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sqlBinaryNode(uint,date)", sqlBinaryNode{uintVal, dateVal}, []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sqlBinaryNode(uint,datetime)", sqlBinaryNode{uintVal, datetimeVal}, []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sqlBinaryNode(uint,objectid)", sqlBinaryNode{uintVal, objectIDVal}, []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sqlBinaryNode(float,int)", sqlBinaryNode{floatVal, intVal}, []types.EvalType{types.EvalDouble, types.EvalInt64}},
		{"sqlBinaryNode(float,uint)", sqlBinaryNode{floatVal, uintVal}, []types.EvalType{types.EvalDouble, types.EvalUint64}},
		{"sqlBinaryNode(float,float)", sqlBinaryNode{floatVal, floatVal}, []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sqlBinaryNode(float,decimal)", sqlBinaryNode{floatVal, decimalVal}, []types.EvalType{types.EvalDouble, types.EvalDecimal128}},
		{"sqlBinaryNode(float,bool)", sqlBinaryNode{floatVal, boolVal}, []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sqlBinaryNode(float,str)", sqlBinaryNode{floatVal, strVal}, []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sqlBinaryNode(float,date)", sqlBinaryNode{floatVal, dateVal}, []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sqlBinaryNode(float,datetime)", sqlBinaryNode{floatVal, datetimeVal}, []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sqlBinaryNode(float,objectid)", sqlBinaryNode{floatVal, objectIDVal}, []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sqlBinaryNode(decimal,int)", sqlBinaryNode{decimalVal, intVal}, []types.EvalType{types.EvalDecimal128, types.EvalInt64}},
		{"sqlBinaryNode(decimal,uint)", sqlBinaryNode{decimalVal, uintVal}, []types.EvalType{types.EvalDecimal128, types.EvalUint64}},
		{"sqlBinaryNode(decimal,float)", sqlBinaryNode{decimalVal, floatVal}, []types.EvalType{types.EvalDecimal128, types.EvalDouble}},
		{"sqlBinaryNode(decimal,decimal)", sqlBinaryNode{decimalVal, decimalVal}, []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sqlBinaryNode(decimal,bool)", sqlBinaryNode{decimalVal, boolVal}, []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sqlBinaryNode(decimal,str)", sqlBinaryNode{decimalVal, strVal}, []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sqlBinaryNode(decimal,date)", sqlBinaryNode{decimalVal, dateVal}, []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sqlBinaryNode(decimal,datetime)", sqlBinaryNode{decimalVal, datetimeVal}, []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sqlBinaryNode(decimal,objectid)", sqlBinaryNode{decimalVal, objectIDVal}, []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sqlBinaryNode(bool,int)", sqlBinaryNode{boolVal, intVal}, []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sqlBinaryNode(bool,uint)", sqlBinaryNode{boolVal, uintVal}, []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sqlBinaryNode(bool,float)", sqlBinaryNode{boolVal, floatVal}, []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sqlBinaryNode(bool,decimal)", sqlBinaryNode{boolVal, decimalVal}, []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sqlBinaryNode(bool,bool)", sqlBinaryNode{boolVal, boolVal}, []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"sqlBinaryNode(bool,str)", sqlBinaryNode{boolVal, strVal}, []types.EvalType{types.EvalString, types.EvalString}},
		{"sqlBinaryNode(bool,date)", sqlBinaryNode{boolVal, dateVal}, []types.EvalType{types.EvalDate, types.EvalDate}},
		{"sqlBinaryNode(bool,datetime)", sqlBinaryNode{boolVal, datetimeVal}, []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"sqlBinaryNode(bool,objectid)", sqlBinaryNode{boolVal, objectIDVal}, []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		{"sqlBinaryNode(str,int)", sqlBinaryNode{strVal, intVal}, []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sqlBinaryNode(str,uint)", sqlBinaryNode{strVal, uintVal}, []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sqlBinaryNode(str,float)", sqlBinaryNode{strVal, floatVal}, []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sqlBinaryNode(str,decimal)", sqlBinaryNode{strVal, decimalVal}, []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sqlBinaryNode(str,bool)", sqlBinaryNode{strVal, boolVal}, []types.EvalType{types.EvalString, types.EvalString}},
		{"sqlBinaryNode(str,str)", sqlBinaryNode{strVal, strVal}, []types.EvalType{types.EvalString, types.EvalString}},
		{"sqlBinaryNode(str,date)", sqlBinaryNode{strVal, dateVal}, []types.EvalType{types.EvalDate, types.EvalDate}},
		{"sqlBinaryNode(str,datetime)", sqlBinaryNode{strVal, datetimeVal}, []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		// str + objectid resolve to a objectids when the str is a literal (SQLValueExpr).
		{"sqlBinaryNode(strVal,objectid)", sqlBinaryNode{strVal, objectIDVal}, []types.EvalType{types.EvalObjectID, types.EvalObjectID}},
		// otherwise, resolves to a string.
		{"sqlBinaryNode(strCol,objectid)", sqlBinaryNode{strCol, objectIDVal}, []types.EvalType{types.EvalString, types.EvalString}},
		{"sqlBinaryNode(date,int)", sqlBinaryNode{dateVal, intVal}, []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sqlBinaryNode(date,uint)", sqlBinaryNode{dateVal, uintVal}, []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sqlBinaryNode(date,float)", sqlBinaryNode{dateVal, floatVal}, []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sqlBinaryNode(date,decimal)", sqlBinaryNode{dateVal, decimalVal}, []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sqlBinaryNode(date,bool)", sqlBinaryNode{dateVal, boolVal}, []types.EvalType{types.EvalDate, types.EvalDate}},
		{"sqlBinaryNode(date,str)", sqlBinaryNode{dateVal, strVal}, []types.EvalType{types.EvalDate, types.EvalDate}},
		{"sqlBinaryNode(date,date)", sqlBinaryNode{dateVal, dateVal}, []types.EvalType{types.EvalDate, types.EvalDate}},
		{"sqlBinaryNode(date,datetime)", sqlBinaryNode{dateVal, datetimeVal}, []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"sqlBinaryNode(date,objectid)", sqlBinaryNode{dateVal, objectIDVal}, []types.EvalType{types.EvalDate, types.EvalDate}},
		{"sqlBinaryNode(datetime,int)", sqlBinaryNode{datetimeVal, intVal}, []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sqlBinaryNode(datetime,uint)", sqlBinaryNode{datetimeVal, uintVal}, []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sqlBinaryNode(datetime,float)", sqlBinaryNode{datetimeVal, floatVal}, []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sqlBinaryNode(datetime,decimal)", sqlBinaryNode{datetimeVal, decimalVal}, []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sqlBinaryNode(datetime,bool)", sqlBinaryNode{datetimeVal, boolVal}, []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"sqlBinaryNode(datetime,str)", sqlBinaryNode{datetimeVal, strVal}, []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"sqlBinaryNode(datetime,date)", sqlBinaryNode{datetimeVal, dateVal}, []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"sqlBinaryNode(datetime,datetime)", sqlBinaryNode{datetimeVal, datetimeVal}, []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"sqlBinaryNode(datetime,objectid)", sqlBinaryNode{datetimeVal, objectIDVal}, []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"sqlBinaryNode(objectid,int)", sqlBinaryNode{objectIDVal, intVal}, []types.EvalType{types.EvalInt64, types.EvalInt64}},
		{"sqlBinaryNode(objectid,uint)", sqlBinaryNode{objectIDVal, uintVal}, []types.EvalType{types.EvalUint64, types.EvalUint64}},
		{"sqlBinaryNode(objectid,float)", sqlBinaryNode{objectIDVal, floatVal}, []types.EvalType{types.EvalDouble, types.EvalDouble}},
		{"sqlBinaryNode(objectid,decimal)", sqlBinaryNode{objectIDVal, decimalVal}, []types.EvalType{types.EvalDecimal128, types.EvalDecimal128}},
		{"sqlBinaryNode(objectid,bool)", sqlBinaryNode{objectIDVal, boolVal}, []types.EvalType{types.EvalBoolean, types.EvalBoolean}},
		// objectid + str resolve to a objectids when the str is a literal (SQLValueExpr).
		{"sqlBinaryNode(objectid,strVal)", sqlBinaryNode{objectIDVal, strVal}, []types.EvalType{types.EvalObjectID, types.EvalObjectID}},
		// otherwise, resolves to a string.
		{"sqlBinaryNode(objectid,strCol)", sqlBinaryNode{objectIDVal, strCol}, []types.EvalType{types.EvalString, types.EvalString}},
		{"sqlBinaryNode(objectid,date)", sqlBinaryNode{objectIDVal, dateVal}, []types.EvalType{types.EvalDate, types.EvalDate}},
		{"sqlBinaryNode(objectid,datetime)", sqlBinaryNode{objectIDVal, datetimeVal}, []types.EvalType{types.EvalDatetime, types.EvalDatetime}},
		{"sqlBinaryNode(objectid,objectid)", sqlBinaryNode{objectIDVal, objectIDVal}, []types.EvalType{types.EvalObjectID, types.EvalObjectID}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			newNode := test.binaryNode.reconcileComparison()
			leftType := newNode.left.EvalType()
			rightType := newNode.right.EvalType()

			req.Equal(test.expectedTypes[0], leftType, "incorrect EvalType for left: got %v, expected %v", types.EvalTypeToSQLType(leftType), types.EvalTypeToSQLType(test.expectedTypes[0]))
			req.Equal(test.expectedTypes[1], rightType, "incorrect EvalType for right: got %v, expected %v", types.EvalTypeToSQLType(rightType), types.EvalTypeToSQLType(test.expectedTypes[1]))
		})
	}
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
			[]types.EvalType{types.EvalInt64, types.EvalInt64, types.EvalUint64, types.EvalDouble, types.EvalDecimal128, types.EvalBoolean, types.EvalString, types.EvalDate, types.EvalDatetime},
		},
		{
			"convertExprs(uint)", makeValSlice(uintVal), allTypes,
			[]types.EvalType{types.EvalUint64, types.EvalInt64, types.EvalUint64, types.EvalDouble, types.EvalDecimal128, types.EvalBoolean, types.EvalString, types.EvalDate, types.EvalDatetime},
		},
		{
			"convertExprs(float)", makeValSlice(floatVal), allTypes,
			[]types.EvalType{types.EvalDouble, types.EvalInt64, types.EvalUint64, types.EvalDouble, types.EvalDecimal128, types.EvalBoolean, types.EvalString, types.EvalDate, types.EvalDatetime},
		},
		{
			"convertExprs(decimal)", makeValSlice(decimalVal), allTypes,
			[]types.EvalType{types.EvalDecimal128, types.EvalInt64, types.EvalUint64, types.EvalDouble, types.EvalDecimal128, types.EvalBoolean, types.EvalString, types.EvalDate, types.EvalDatetime},
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

func TestStrToDateEvalType(t *testing.T) {
	type test struct {
		name         string
		testExpr     []SQLExpr
		expectedType types.EvalType
	}

	knd := values.MySQLValueKind
	tests := []test{
		{"sql_str_to_date_type_0", []SQLExpr{NewSQLValueExpr(values.NewSQLVarchar(knd, "2000-01-01")), NewSQLValueExpr(values.NewSQLVarchar(knd, "%Y-%m-%d"))}, types.EvalDate},
		{"sql_str_to_date_type_1", []SQLExpr{NewSQLValueExpr(values.NewSQLVarchar(knd, "2000-01-01")), NewSQLValueExpr(values.NewSQLVarchar(knd, "hello%iworld"))}, types.EvalDatetime},
		{"sql_str_to_date_type_2", []SQLExpr{NewSQLValueExpr(values.NewSQLVarchar(knd, "2000-01-01")), NewSQLValueExpr(values.NewSQLVarchar(knd, "hello%%Hworld"))}, types.EvalDatetime},
		{"sql_str_to_date_type_3", []SQLExpr{NewSQLValueExpr(values.NewSQLVarchar(knd, "2000-01-01")), NewSQLValueExpr(values.NewSQLVarchar(knd, "hi%%wd"))}, types.EvalDate},
		{"sql_str_to_date_type_4", []SQLExpr{NewSQLValueExpr(values.NewSQLVarchar(knd, "2000-01-01")), NewSQLValueExpr(values.NewSQLVarchar(knd, "%S%Y"))}, types.EvalDatetime},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			outputType := strToDateEvalType(test.testExpr)
			req.Equal(outputType, test.expectedType)
		})
	}
}
