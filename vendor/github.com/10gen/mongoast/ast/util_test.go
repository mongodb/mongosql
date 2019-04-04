package ast_test

import (
	"fmt"
	"testing"

	"github.com/10gen/mongoast/ast"
)

func TestGetDottedFieldName(t *testing.T) {
	testCases := []struct {
		name     string
		input    ast.Expr
		expected string
	}{
		{
			"All field refs",
			ast.NewFieldRef("c", ast.NewFieldRef("b", ast.NewFieldRef("a", nil))),
			"a.b.c",
		},
		{
			"Array refs",
			ast.NewFieldOrArrayIndexRef(2, ast.NewFieldOrArrayIndexRef(1, ast.NewFieldRef("a", nil))),
			"a.1.2",
		},
		{
			"Array refs mixed",
			ast.NewFieldRef("b", ast.NewFieldOrArrayIndexRef(1, ast.NewFieldRef("a", nil))),
			"a.1.b",
		},
		{
			"Field Or Array Index refs with VariableRef",
			ast.NewFieldOrArrayIndexRef(2, ast.NewFieldOrArrayIndexRef(1, ast.NewVariableRef("a"))),
			"$a.1.2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := ast.GetDottedFieldName(tc.input)
			if tc.expected != input {
				t.Fatal(fmt.Sprintf("expected %s, got %s", tc.expected, input))
			}
		})
	}
}
