package eval_test

import (
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/eval"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/google/go-cmp/cmp"
)

func TestSubstituteVariables(t *testing.T) {
	testCases := []struct {
		name     string
		expr     ast.Expr
		expected ast.Expr
	}{
		{
			`$$a`,
			ast.NewVariableRef("a"),
			ast.NewConstant(bsonutil.Int32(1)),
		},
		{
			`$$b`,
			ast.NewVariableRef("b"),
			ast.NewVariableRef("b"),
		},
		{
			`$$x.y`,
			ast.NewFieldRef("y", ast.NewVariableRef("x")),
			ast.NewConstant(bsonutil.Int32(5)),
		},
		{
			`$$x.x`,
			ast.NewFieldRef("x", ast.NewVariableRef("x")),
			ast.NewConstant(bsonutil.Null()),
		},
		{
			`$$y.z`,
			ast.NewFieldRef("z", ast.NewVariableRef("y")),
			ast.NewConstant(bsonutil.Int32(10)),
		},
		{
			`$$p.q`,
			ast.NewFieldRef("q", ast.NewVariableRef("p")),
			ast.NewFieldRef("q", ast.NewVariableRef("p")),
		},
		{
			`$$a.b`,
			ast.NewFieldRef("b", ast.NewVariableRef("a")),
			ast.NewConstant(bsonutil.Null()),
		},
		{
			`$a`,
			ast.NewFieldRef("a", nil),
			ast.NewFieldRef("a", nil),
		},
		{
			`$a.b`,
			ast.NewFieldRef("b", ast.NewFieldRef("a", nil)),
			ast.NewFieldRef("b", ast.NewFieldRef("a", nil)),
		},
		{
			`{"$let": {"vars": {"b": 2}, "in": "$$a"}}`,
			ast.NewLet(
				[]*ast.LetVariable{
					ast.NewLetVariable("b", ast.NewConstant(bsonutil.Int32(2))),
				},
				ast.NewVariableRef("a"),
			),
			ast.NewLet(
				[]*ast.LetVariable{
					ast.NewLetVariable("b", ast.NewConstant(bsonutil.Int32(2))),
				},
				ast.NewConstant(bsonutil.Int32(1)),
			),
		},
		{
			`{"$let": {"vars": {"a": 1}, "in": "$$a"}}`,
			ast.NewLet(
				[]*ast.LetVariable{
					ast.NewLetVariable("a", ast.NewConstant(bsonutil.Int32(1))),
				},
				ast.NewVariableRef("a"),
			),
			ast.NewLet(
				[]*ast.LetVariable{
					ast.NewLetVariable("a", ast.NewConstant(bsonutil.Int32(1))),
				},
				ast.NewVariableRef("a"),
			),
		},
	}

	variables := map[string]ast.Expr{
		"a": ast.NewConstant(bsonutil.Int32(1)),
		"x": ast.NewConstant(bsonutil.DocumentFromElements("y", bsonutil.Int32(5))),
		"y": ast.NewDocument(ast.NewDocumentElement("z", ast.NewConstant(bsonutil.Int32(10)))),
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := eval.SubstituteVariables(tc.expr, variables)
			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("actual did not match expected\n  %s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}
