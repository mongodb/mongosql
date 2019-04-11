package testutil

import (
	"strings"

	"github.com/10gen/mongoast/ast"
)

// StringToFieldRef builds a FieldRef from a string, used for testing.
func StringToFieldRef(path string) *ast.FieldRef {
	pathParts := strings.Split(path, ".")
	var cur ast.Expr
	for _, part := range pathParts {
		cur = ast.NewFieldRef(part, cur)
	}
	return cur.(*ast.FieldRef)
}

// StringToFieldRefOrVarRef builds a FieldRef from a string, used for testing.
func StringToFieldRefOrVarRef(path string) *ast.FieldRef {
	pathParts := strings.Split(path, ".")
	var cur ast.Expr
	for i, part := range pathParts {
		if i == 0 {
			if strings.HasPrefix(part, "$$") {
				cur = ast.NewVariableRef(strings.TrimPrefix(part, "$$"))
			} else {
				cur = ast.NewFieldRef(strings.TrimPrefix(part, "$"), nil)
			}
		}
		cur = ast.NewFieldRef(part, cur)
	}
	return cur.(*ast.FieldRef)
}
