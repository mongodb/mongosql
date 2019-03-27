package evaluator_test

import (
	"fmt"
	"testing"

	"github.com/10gen/mongoast/ast"

	. "github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/stretchr/testify/require"
)

func TestComputeDocNestingDepth(t *testing.T) {
	type test struct {
		expr  ast.Expr
		depth uint32
	}

	runTests := func(tests []test) {
		for idx, test := range tests {
			name := fmt.Sprintf("%d", idx)
			t.Run(name, func(t *testing.T) {
				depth := ComputeDocNestingDepthWithMaxDepth(test.expr, MaxDepth)
				require.Equal(t, test.depth, depth)
			})
		}
	}

	tests := []test{
		{
			ast.NewArray( // 0
				ast.NewDocument( // 1
					ast.NewDocumentElement("$match", ast.NewDocument( // 2
						ast.NewDocumentElement("a",
							astutil.Int64Value(10), // 3
						),
					)),
				),
			),
			3,
		},
		{
			ast.NewDocument( // 0
				ast.NewDocumentElement("$match", ast.NewDocument( // 1
					ast.NewDocumentElement("a", ast.NewAggExpr( // 2
						ast.NewFunction("$add", // 3
							ast.NewArray( // 4
								ast.NewFieldRef("b", nil), // 5
								ast.NewFieldRef("c", nil), // 5
							),
						),
					)),
				)),
			),
			5,
		},
		{
			ast.NewArray( // 0
				ast.NewDocument( // 1
					ast.NewDocumentElement("$project", ast.NewDocument( // 2
						ast.NewDocumentElement("x", ast.NewBinary("$and", // 3
							ast.NewFieldRef("a", nil), // 5
							ast.NewFieldRef("b", nil),
						)),
					)),
				),
			),
			5,
		},
		{
			ast.NewArray( // 0
				ast.NewDocument( // 1
					ast.NewDocumentElement("$match", ast.NewDocument( // 2
						ast.NewDocumentElement("a", ast.NewFunction("$ne", // 3
							astutil.NullLiteral, // 4
						)),
					)),
				),
				ast.NewDocument( // 1
					ast.NewDocumentElement("$lookup", ast.NewDocument( // 2
						ast.NewDocumentElement("from", astutil.StringValue("foo")),       // 3
						ast.NewDocumentElement("localField", astutil.StringValue("a")),   // 3
						ast.NewDocumentElement("foreignField", astutil.StringValue("a")), // 3
						ast.NewDocumentElement("as", astutil.StringValue("__joined_b")),  // 3
					)),
				),
				ast.NewDocument( // 1
					ast.NewDocumentElement("$unwind", ast.NewDocument( // 2
						ast.NewDocumentElement("path", astutil.StringValue("$__joined_b")),                // 3
						ast.NewDocumentElement("preserveNullAndEmptyArrays", astutil.BooleanValue(false)), // 3
					)),
				),
				ast.NewDocument( // 1
					ast.NewDocumentElement("$project", ast.NewDocument( // 2
						ast.NewDocumentElement("__joined_b._id", astutil.Int32Value(1)),    // 3
						ast.NewDocumentElement("__joined_b.a", astutil.Int32Value(1)),      // 3
						ast.NewDocumentElement("__joined_b.b", astutil.Int32Value(1)),      // 3
						ast.NewDocumentElement("__joined_b.c", astutil.Int32Value(1)),      // 3
						ast.NewDocumentElement("__joined_b.d.e", astutil.Int32Value(1)),    // 3
						ast.NewDocumentElement("__joined_b.d.f", astutil.Int32Value(1)),    // 3
						ast.NewDocumentElement("__joined_b.filter", astutil.Int32Value(1)), // 3
						ast.NewDocumentElement("__joined_b.g", astutil.Int32Value(1)),      // 3
						ast.NewDocumentElement("_id", astutil.Int32Value(1)),               // 3
						ast.NewDocumentElement("a", astutil.Int32Value(1)),                 // 3
						ast.NewDocumentElement("b", astutil.Int32Value(1)),                 // 3
						ast.NewDocumentElement("__predicate", ast.NewLet( // 3
							[]*ast.LetVariable{
								ast.NewLetVariable("predicate", ast.NewLet( // 6 (vars are +3)
									[]*ast.LetVariable{
										ast.NewLetVariable("left", ast.NewFieldRef("a", nil)),               // 9 (vars are +3)
										ast.NewLetVariable("right", ast.NewFieldRef("__joined_b.d.f", nil)), // 9 (vars are +3)
									},
									ast.NewFunction("$cond", // 8 (in is +2)
										ast.NewArray( // 9
											ast.NewBinary("$or", // 10
												ast.NewBinary("$lte", // 12 (Binary args are +2)
													ast.NewVariableRef("left"), // 14 (Binary args are +2)
													astutil.NullLiteral,        // 14 (Binary args are +2)
												),
												ast.NewBinary("$lte", // 12 (Binary args are +2)
													ast.NewVariableRef("right"), // 14 (Binary args are +2)
													astutil.NullLiteral,         // 14 (Binary args are +2)
												),
											),
											astutil.NullLiteral, // 10
											ast.NewBinary("$eq", // 10
												ast.NewVariableRef("left"),  // 12 (Binary args are +2)
												ast.NewVariableRef("right"), // 12 (Binary args are +2)
											),
										)),
								)),
							},
							ast.NewFunction("$cond", // 5 (in is +2)
								ast.NewArray( // 6
									ast.NewFunction("$or", // 7
										ast.NewArray( // 8
											ast.NewBinary("$eq", // 9
												ast.NewVariableRef("predicate"), // 11
												astutil.BooleanValue(false),     // 11
											),
											ast.NewBinary("$eq", ast.NewVariableRef("predicate"), astutil.Int32Value(0)),
											ast.NewBinary("$eq", ast.NewVariableRef("predicate"), astutil.StringValue("0")),
											ast.NewBinary("$eq", ast.NewVariableRef("predicate"), astutil.StringValue("-0")),
											ast.NewBinary("$eq", ast.NewVariableRef("predicate"), astutil.StringValue("0.0")),
											ast.NewBinary("$eq", ast.NewVariableRef("predicate"), astutil.StringValue("-0.0")),
											ast.NewBinary("$eq", ast.NewVariableRef("predicate"), astutil.NullLiteral),
										),
									),
									astutil.BooleanValue(false), // 7
									astutil.BooleanValue(true),  // 7
								),
							),
						)),
					)),
				),
				ast.NewDocument( // 1
					ast.NewDocumentElement("$match", ast.NewDocument( // 2
						ast.NewDocumentElement("__predicate", astutil.BooleanValue(true)), // 3
					)),
				),
				ast.NewDocument( // 1
					ast.NewDocumentElement("$project", ast.NewDocument( // 2
						ast.NewDocumentElement("test_DOT_a_DOT_b", astutil.StringValue("$b")),                // 3
						ast.NewDocumentElement("test_DOT_a_DOT__id", astutil.StringValue("$_id")),            // 3
						ast.NewDocumentElement("test_DOT_b_DOT_e", astutil.StringValue("$__joined_b.d.e")),   // 3
						ast.NewDocumentElement("test_DOT_b_DOT_g", astutil.StringValue("$__joined_b.g")),     // 3
						ast.NewDocumentElement("test_DOT_b_DOT_f", astutil.StringValue("$__joined_b.d.f")),   // 3
						ast.NewDocumentElement("test_DOT_b_DOT__id", astutil.StringValue("$__joined_b._id")), // 3
						ast.NewDocumentElement("test_DOT_a_DOT_a", astutil.StringValue("$a")),                // 3
						ast.NewDocumentElement("test_DOT_b_DOT_a", astutil.StringValue("$__joined_b.a")),     // 3
						ast.NewDocumentElement("test_DOT_b_DOT_b", astutil.StringValue("$__joined_b.b")),     // 3
						ast.NewDocumentElement("test_DOT_b_DOT_c", astutil.StringValue("$__joined_b.c")),     // 3
					)),
				),
			),
			14,
		},
	}

	runTests(tests)
}
