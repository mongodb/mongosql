package eval_test

import (
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/eval"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestEvaluateExpr(t *testing.T) {
	testCases := []struct {
		name     string
		expr     ast.Expr
		value    bsoncore.Value
		expected bsoncore.Value
		err      error
	}{
		// Field Access
		{
			`a`,
			ast.NewFieldRef("a", nil),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
			),
			bsonutil.Int32(1),
			nil,
		},
		{
			`a.b`,
			ast.NewFieldRef("b", ast.NewFieldRef("a", nil)),
			bsonutil.DocumentFromElements(
				"a", bsonutil.DocumentFromElements(
					"b", bsonutil.Int32(1),
				),
			),
			bsonutil.Int32(1),
			nil,
		},
		{
			`a.b.c`,
			ast.NewFieldRef("c", ast.NewFieldRef("b", ast.NewFieldRef("a", nil))),
			bsonutil.DocumentFromElements(
				"a", bsonutil.DocumentFromElements(
					"b", bsonutil.DocumentFromElements(
						"c", bsonutil.Int32(2),
					),
				),
			),
			bsonutil.Int32(2),
			nil,
		},
		{
			`a.2 (match syntax) as a field`,
			ast.NewFieldOrArrayIndexRef(2, ast.NewFieldRef("a", nil)),
			bsonutil.DocumentFromElements(
				"a", bsonutil.DocumentFromElements(
					"2", bsonutil.Int32(1),
				),
			),
			bsonutil.Int32(1),
			nil,
		},
		{
			`a.2.c (match syntax) as a field`,
			ast.NewFieldRef("c", ast.NewFieldOrArrayIndexRef(2, ast.NewFieldRef("a", nil))),
			bsonutil.DocumentFromElements(
				"a", bsonutil.DocumentFromElements(
					"2", bsonutil.DocumentFromElements(
						"c", bsonutil.Int32(2),
					),
				),
			),
			bsonutil.Int32(2),
			nil,
		},
		{
			`a.2 (match syntax) as an array`,
			ast.NewFieldOrArrayIndexRef(2, ast.NewFieldRef("a", nil)),
			bsonutil.DocumentFromElements(
				"a", bsonutil.ArrayFromValues(
					bsonutil.Int32(3),
					bsonutil.Int32(2),
					bsonutil.Int32(1),
				),
			),
			bsonutil.Int32(1),
			nil,
		},
		{
			`a.2.c (match syntax) as an array`,
			ast.NewFieldRef("c", ast.NewFieldOrArrayIndexRef(2, ast.NewFieldRef("a", nil))),
			bsonutil.DocumentFromElements(
				"a", bsonutil.ArrayFromValues(
					bsonutil.Int32(3),
					bsonutil.Int32(2),
					bsonutil.DocumentFromElements(
						"c", bsonutil.Int32(1),
					),
				),
			),
			bsonutil.Int32(1),
			nil,
		},
		{
			`a.$[2]`,
			ast.NewArrayIndexRef(
				ast.NewConstant(bsonutil.Int32(2)),
				ast.NewFieldRef("a", nil),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.ArrayFromValues(
					bsonutil.Int32(3),
					bsonutil.Int32(2),
					bsonutil.Int32(1),
				),
			),
			bsonutil.Int32(1),
			nil,
		},
		{
			`a.$[2].c`,
			ast.NewFieldRef(
				"c", ast.NewArrayIndexRef(
					ast.NewConstant(bsonutil.Int32(2)),
					ast.NewFieldRef("a", nil),
				),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.ArrayFromValues(
					bsonutil.Int32(3),
					bsonutil.Int32(2),
					bsonutil.DocumentFromElements(
						"c", bsonutil.Int32(1),
					),
				),
			),
			bsonutil.Int32(1),
			nil,
		},
		{
			`a[b]`,
			ast.NewArrayIndexRef(
				ast.NewFieldRef("b", nil),
				ast.NewFieldRef("a", nil),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.ArrayFromValues(
					bsonutil.Int32(1),
					bsonutil.Int32(2),
					bsonutil.Int32(3),
				),
				"b", bsonutil.Int32(1),
			),
			bsonutil.Int32(2),
			nil,
		},
		{
			`[1, 2, 4, 8][a]`,
			ast.NewArrayIndexRef(
				ast.NewFieldRef("a", nil),
				ast.NewArray(
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(2)),
					ast.NewConstant(bsonutil.Int32(4)),
					ast.NewConstant(bsonutil.Int32(8)),
				),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(2),
			),
			bsonutil.Int32(4),
			nil,
		},
		{
			`a.b as an integer`,
			ast.NewFieldRef("b", ast.NewFieldRef("a", nil)),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
			),
			bsonutil.Null(),
			bsoncore.ErrElementNotFound,
		},
		{
			`a.2 (match syntax) as an integer`,
			ast.NewFieldOrArrayIndexRef(2, ast.NewFieldRef("a", nil)),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
			),
			bsonutil.Null(),
			bsoncore.ErrElementNotFound,
		},
		{
			`a.$[2] as an integer`,
			ast.NewArrayIndexRef(
				ast.NewConstant(bsonutil.Int32(2)),
				ast.NewFieldRef("a", nil),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
			),
			bsonutil.Null(),
			bsoncore.ErrElementNotFound,
		},
		{
			`a.$[b] with string index`,
			ast.NewArrayIndexRef(
				ast.NewFieldRef("b", nil),
				ast.NewFieldRef("a", nil),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.ArrayFromValues(
					bsonutil.Int32(1),
					bsonutil.Int32(2),
					bsonutil.Int32(3),
				),
				"b", bsonutil.String("foo"),
			),
			bsonutil.Null(),
			errors.New("array index must be an integer"),
		},
		{
			`a.$[b] with index out of range`,
			ast.NewArrayIndexRef(
				ast.NewFieldRef("b", nil),
				ast.NewFieldRef("a", nil),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.ArrayFromValues(
					bsonutil.Int32(1),
					bsonutil.Int32(2),
					bsonutil.Int32(3),
				),
				"b", bsonutil.Int32(5),
			),
			bsonutil.Null(),
			errors.New("out of bounds"),
		},
		{
			`[1, 2, 4, 8][a] with index out range`,
			ast.NewArrayIndexRef(
				ast.NewFieldRef("a", nil),
				ast.NewArray(
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(2)),
					ast.NewConstant(bsonutil.Int32(4)),
					ast.NewConstant(bsonutil.Int32(8)),
				),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(5),
			),
			bsonutil.Null(),
			errors.New("array index out of range"),
		},
		{
			`[1, 2, 4, 8][a] with negative index`,
			ast.NewArrayIndexRef(
				ast.NewFieldRef("a", nil),
				ast.NewArray(
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(2)),
					ast.NewConstant(bsonutil.Int32(4)),
					ast.NewConstant(bsonutil.Int32(8)),
				),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(-1),
			),
			bsonutil.Null(),
			errors.New("array index out of range"),
		},
		// Logical
		{
			`true && true`,
			ast.NewBinary(ast.And,
				ast.NewConstant(bsonutil.Boolean(true)),
				ast.NewConstant(bsonutil.Boolean(true)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		{
			`true && false`,
			ast.NewBinary(ast.And,
				ast.NewConstant(bsonutil.Boolean(true)),
				ast.NewConstant(bsonutil.Boolean(false)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(false),
			nil,
		},
		{
			`false && true`,
			ast.NewBinary(ast.And,
				ast.NewConstant(bsonutil.Boolean(false)),
				ast.NewConstant(bsonutil.Boolean(true)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(false),
			nil,
		},
		{
			`false && false`,
			ast.NewBinary(ast.And,
				ast.NewConstant(bsonutil.Boolean(false)),
				ast.NewConstant(bsonutil.Boolean(false)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(false),
			nil,
		},
		{
			`true || true`,
			ast.NewBinary(ast.Or,
				ast.NewConstant(bsonutil.Boolean(true)),
				ast.NewConstant(bsonutil.Boolean(true)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		{
			`true || false`,
			ast.NewBinary(ast.Or,
				ast.NewConstant(bsonutil.Boolean(true)),
				ast.NewConstant(bsonutil.Boolean(false)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		{
			`false || true`,
			ast.NewBinary(ast.Or,
				ast.NewConstant(bsonutil.Boolean(false)),
				ast.NewConstant(bsonutil.Boolean(true)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		{
			`false || false`,
			ast.NewBinary(ast.Or,
				ast.NewConstant(bsonutil.Boolean(false)),
				ast.NewConstant(bsonutil.Boolean(false)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(false),
			nil,
		},
		{
			`1 && 2`,
			ast.NewBinary(ast.And,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(2)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		{
			`0 && 1`,
			ast.NewBinary(ast.And,
				ast.NewConstant(bsonutil.Int32(0)),
				ast.NewConstant(bsonutil.Int32(1)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(false),
			nil,
		},
		{
			`1 || 2`,
			ast.NewBinary(ast.Or,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(2)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		{
			`0 || 1`,
			ast.NewBinary(ast.Or,
				ast.NewConstant(bsonutil.Int32(0)),
				ast.NewConstant(bsonutil.Int32(1)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		{
			`0 || 0`,
			ast.NewBinary(ast.Or,
				ast.NewConstant(bsonutil.Int32(0)),
				ast.NewConstant(bsonutil.Int32(0)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(false),
			nil,
		},
		// Comparison
		{
			`10 == 10`,
			ast.NewBinary(ast.Equals,
				ast.NewConstant(bsonutil.Int32(10)),
				ast.NewConstant(bsonutil.Int32(10)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		{
			`1 == 10`,
			ast.NewBinary(ast.Equals,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(10)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(false),
			nil,
		},
		{
			`10 > 10`,
			ast.NewBinary(ast.GreaterThan,
				ast.NewConstant(bsonutil.Int32(10)),
				ast.NewConstant(bsonutil.Int32(10)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(false),
			nil,
		},
		{
			`1 > 10`,
			ast.NewBinary(ast.GreaterThan,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(10)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(false),
			nil,
		},
		{
			`10 > 1`,
			ast.NewBinary(ast.GreaterThan,
				ast.NewConstant(bsonutil.Int32(10)),
				ast.NewConstant(bsonutil.Int32(1)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		{
			`10 >= 10`,
			ast.NewBinary(ast.GreaterThanOrEquals,
				ast.NewConstant(bsonutil.Int32(10)),
				ast.NewConstant(bsonutil.Int32(10)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		{
			`1 >= 10`,
			ast.NewBinary(ast.GreaterThanOrEquals,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(10)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(false),
			nil,
		},
		{
			`10 >= 1`,
			ast.NewBinary(ast.GreaterThanOrEquals,
				ast.NewConstant(bsonutil.Int32(10)),
				ast.NewConstant(bsonutil.Int32(1)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		{
			`10 < 10`,
			ast.NewBinary(ast.LessThan,
				ast.NewConstant(bsonutil.Int32(10)),
				ast.NewConstant(bsonutil.Int32(10)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(false),
			nil,
		},
		{
			`1 < 10`,
			ast.NewBinary(ast.LessThan,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(10)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		{
			`10 < 1`,
			ast.NewBinary(ast.LessThan,
				ast.NewConstant(bsonutil.Int32(10)),
				ast.NewConstant(bsonutil.Int32(1)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(false),
			nil,
		},
		{
			`10 <= 10`,
			ast.NewBinary(ast.LessThanOrEquals,
				ast.NewConstant(bsonutil.Int32(10)),
				ast.NewConstant(bsonutil.Int32(10)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		{
			`1 <= 10`,
			ast.NewBinary(ast.LessThanOrEquals,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(10)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		{
			`10 <= 1`,
			ast.NewBinary(ast.LessThanOrEquals,
				ast.NewConstant(bsonutil.Int32(10)),
				ast.NewConstant(bsonutil.Int32(1)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(false),
			nil,
		},
		{
			`10 != 10`,
			ast.NewBinary(ast.NotEquals,
				ast.NewConstant(bsonutil.Int32(10)),
				ast.NewConstant(bsonutil.Int32(10)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(false),
			nil,
		},
		{
			`1 != 10`,
			ast.NewBinary(ast.NotEquals,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(10)),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		{
			`{ $cmp: [1, 10] }`,
			ast.NewBinary(ast.Compare,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(10)),
			),
			bsonutil.Null(),
			bsonutil.Int32(-1),
			nil,
		},
		{
			`{ $cmp: [10, 1] }`,
			ast.NewBinary(ast.Compare,
				ast.NewConstant(bsonutil.Int32(10)),
				ast.NewConstant(bsonutil.Int32(1)),
			),
			bsonutil.Null(),
			bsonutil.Int32(1),
			nil,
		},
		{
			`{ $cmp: [1, 1] }`,
			ast.NewBinary(ast.Compare,
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(1)),
			),
			bsonutil.Null(),
			bsonutil.Int32(0),
			nil,
		},
		// Document Creation
		{
			`{}`,
			ast.NewDocument(),
			bsonutil.Null(),
			bsonutil.EmptyDocument(),
			nil,
		},
		{
			`{a: b}`,
			ast.NewDocument(
				ast.NewDocumentElement(
					"a", ast.NewFieldRef("b", nil),
				),
			),
			bsonutil.DocumentFromElements(
				"b", bsonutil.Int32(1),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
			),
			nil,
		},
		// Let
		{
			`{ $let: { vars: { a: 1, b : 2 }, in: { a < b } }`,
			ast.NewLet(
				[]*ast.LetVariable{
					ast.NewLetVariable("a", ast.NewConstant(bsonutil.Int32(1))),
					ast.NewLetVariable("b", ast.NewConstant(bsonutil.Int32(2))),
				},
				ast.NewBinary(
					ast.LessThan,
					ast.NewVariableRef("a"),
					ast.NewVariableRef("b"),
				),
			),
			bsonutil.Null(),
			bsonutil.Boolean(true),
			nil,
		},
		// Conditional
		{
			`{ $cond: { if: 1 < 2, then: 3, else: 4 } }`,
			ast.NewConditional(
				ast.NewBinary(
					ast.LessThan,
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(2)),
				),
				ast.NewConstant(bsonutil.Int32(3)),
				ast.NewConstant(bsonutil.Int32(4)),
			),
			bsonutil.Null(),
			bsonutil.Int32(3),
			nil,
		},
		{
			`{ $cond: { if: 1 > 2, then: 3, else: 4 } }`,
			ast.NewConditional(
				ast.NewBinary(
					ast.GreaterThan,
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(2)),
				),
				ast.NewConstant(bsonutil.Int32(3)),
				ast.NewConstant(bsonutil.Int32(4)),
			),
			bsonutil.Null(),
			bsonutil.Int32(4),
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := eval.EvaluateExpr(tc.expr, tc.value)
			var wrapped error
			if tc.err == bsoncore.ErrElementNotFound {
				wrapped = bsoncore.ErrElementNotFound
			} else {
				wrapped = errors.Wrap(tc.err, "failed evaluating expression")
			}
			if err != nil && wrapped == nil {
				t.Fatalf("err should be nil, but got %v", err)
			} else if err == nil && wrapped != nil {
				t.Fatalf("err should not be nil, expected %v", tc.err)
			} else if err != nil && wrapped != nil && err.Error() != wrapped.Error() {
				t.Fatalf("expected error %q, but got %q", wrapped.Error(), err.Error())
			}

			if err == nil {
				if tc.expected.Type != actual.Type {
					t.Fatalf("results types are not the same\n  %s", cmp.Diff(tc.expected.Type, actual.Type))
				}

				if !cmp.Equal(tc.expected, actual) {
					t.Fatalf("results are not equal\n  %s", cmp.Diff(tc.expected, actual))
				}
			}
		})
	}
}

func TestPartialEvaluateExpr(t *testing.T) {
	testCases := []struct {
		name     string
		expr     ast.Expr
		value    bsoncore.Value
		expected ast.Expr
	}{
		{
			`a + b`,
			ast.NewFunction(
				"$add",
				ast.NewArray(
					ast.NewFieldRef("a", nil),
					ast.NewFieldRef("b", nil),
				),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
				"b", bsonutil.Int32(2),
			),
			ast.NewFunction(
				"$add",
				ast.NewArray(
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(2)),
				),
			),
		},
		{
			`2 * (a + b)`,
			ast.NewFunction(
				"$multiply",
				ast.NewArray(
					ast.NewConstant(bsonutil.Int32(2)),
					ast.NewFunction(
						"$add",
						ast.NewArray(
							ast.NewFieldRef("a", nil),
							ast.NewFieldRef("b", nil),
						),
					),
				),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
				"b", bsonutil.Int32(2),
			),
			ast.NewFunction(
				"$multiply",
				ast.NewArray(
					ast.NewConstant(bsonutil.Int32(2)),
					ast.NewFunction(
						"$add",
						ast.NewArray(
							ast.NewConstant(bsonutil.Int32(1)),
							ast.NewConstant(bsonutil.Int32(2)),
						),
					),
				),
			),
		},
		{
			`a + b < 5`,
			ast.NewBinary(
				ast.LessThan,
				ast.NewFunction(
					"$add",
					ast.NewArray(
						ast.NewFieldRef("a", nil),
						ast.NewFieldRef("b", nil),
					),
				),
				ast.NewConstant(bsonutil.Int32(5)),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
				"b", bsonutil.Int32(2),
			),
			ast.NewBinary(
				ast.LessThan,
				ast.NewFunction(
					"$add",
					ast.NewArray(
						ast.NewConstant(bsonutil.Int32(1)),
						ast.NewConstant(bsonutil.Int32(2)),
					),
				),
				ast.NewConstant(bsonutil.Int32(5)),
			),
		},
		{
			`{ x: a + b }`,
			ast.NewDocument(
				ast.NewDocumentElement(
					"x", ast.NewFunction(
						"$add",
						ast.NewArray(
							ast.NewFieldRef("a", nil),
							ast.NewFieldRef("b", nil),
						),
					),
				),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
				"b", bsonutil.Int32(2),
			),
			ast.NewDocument(
				ast.NewDocumentElement(
					"x", ast.NewFunction(
						"$add",
						ast.NewArray(
							ast.NewConstant(bsonutil.Int32(1)),
							ast.NewConstant(bsonutil.Int32(2)),
						),
					),
				),
			),
		},
		{
			`[a + b]`,
			ast.NewArray(
				ast.NewFunction(
					"$add",
					ast.NewArray(
						ast.NewFieldRef("a", nil),
						ast.NewFieldRef("b", nil),
					),
				),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
				"b", bsonutil.Int32(2),
			),
			ast.NewArray(
				ast.NewFunction(
					"$add",
					ast.NewArray(
						ast.NewConstant(bsonutil.Int32(1)),
						ast.NewConstant(bsonutil.Int32(2)),
					),
				),
			),
		},
		{
			`x.a + x.b`,
			ast.NewFunction(
				"$add",
				ast.NewArray(
					ast.NewFieldRef("a", ast.NewFieldRef("x", nil)),
					ast.NewFieldRef("b", ast.NewFieldRef("x", nil)),
				),
			),
			bsonutil.DocumentFromElements(
				"x", bsonutil.DocumentFromElements(
					"a", bsonutil.Int32(1),
					"b", bsonutil.Int32(2),
				),
			),
			ast.NewFunction(
				"$add",
				ast.NewArray(
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(2)),
				),
			),
		},
		{
			`x.0 + x.1`,
			ast.NewFunction(
				"$add",
				ast.NewArray(
					ast.NewFieldOrArrayIndexRef(0, ast.NewFieldRef("x", nil)),
					ast.NewFieldOrArrayIndexRef(1, ast.NewFieldRef("x", nil)),
				),
			),
			bsonutil.DocumentFromElements(
				"x", bsonutil.ArrayFromValues(
					bsonutil.Int32(1),
					bsonutil.Int32(2),
				),
			),
			ast.NewFunction(
				"$add",
				ast.NewArray(
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(2)),
				),
			),
		},
		{
			`{ $let: { vars: { x: $a + $b }, in: $$x } }`,
			ast.NewLet(
				[]*ast.LetVariable{
					ast.NewLetVariable(
						"x", ast.NewFunction(
							"$add", ast.NewArray(
								ast.NewFieldRef("a", nil),
								ast.NewFieldRef("b", nil),
							),
						),
					),
				},
				ast.NewVariableRef("x"),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
				"b", bsonutil.Int32(2),
			),
			ast.NewFunction(
				"$add", ast.NewArray(
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(2)),
				),
			),
		},
		{
			`{ $let: { vars: { x: 1 }, in: { $$x + $a } }`,
			ast.NewLet(
				[]*ast.LetVariable{
					ast.NewLetVariable(
						"x", ast.NewConstant(bsonutil.Int32(1)),
					),
				},
				ast.NewFunction(
					"$add", ast.NewArray(
						ast.NewVariableRef("x"),
						ast.NewFieldRef("a", nil),
					),
				),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
			),
			ast.NewFunction(
				"$add", ast.NewArray(
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(1)),
				),
			),
		},
		{
			`{ $cond: { if: $a + $b == 3, then: $a, else: $b } }`,
			ast.NewConditional(
				ast.NewBinary(
					ast.Equals,
					ast.NewFunction(
						"$add", ast.NewArray(
							ast.NewFieldRef("a", nil),
							ast.NewFieldRef("b", nil),
						),
					),
					ast.NewConstant(bsonutil.Int32(3)),
				),
				ast.NewFieldRef("a", nil),
				ast.NewFieldRef("b", nil),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
				"b", bsonutil.Int32(2),
			),
			ast.NewConditional(
				ast.NewBinary(
					ast.Equals,
					ast.NewFunction(
						"$add", ast.NewArray(
							ast.NewConstant(bsonutil.Int32(1)),
							ast.NewConstant(bsonutil.Int32(2)),
						),
					),
					ast.NewConstant(bsonutil.Int32(3)),
				),
				ast.NewConstant(bsonutil.Int32(1)),
				ast.NewConstant(bsonutil.Int32(2)),
			),
		},
		{
			`{ $cond: { if: $a == 1, then: $a + $b, else: 0 } }`,
			ast.NewConditional(
				ast.NewBinary(
					ast.Equals,
					ast.NewFieldRef("a", nil),
					ast.NewConstant(bsonutil.Int32(1)),
				),
				ast.NewFunction(
					"$add", ast.NewArray(
						ast.NewFieldRef("a", nil),
						ast.NewFieldRef("b", nil),
					),
				),
				ast.NewConstant(bsonutil.Int32(0)),
			),
			bsonutil.DocumentFromElements(
				"a", bsonutil.Int32(1),
				"b", bsonutil.Int32(2),
			),
			ast.NewFunction(
				"$add", ast.NewArray(
					ast.NewConstant(bsonutil.Int32(1)),
					ast.NewConstant(bsonutil.Int32(2)),
				),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := eval.PartialEvaluateExpr(tc.expr, tc.value)
			if err != nil {
				t.Fatalf("expected no error, but got %v", err)
			}

			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("results are not equal\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}
