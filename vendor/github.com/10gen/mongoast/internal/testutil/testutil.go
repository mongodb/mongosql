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
