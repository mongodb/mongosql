package analyzer

import (
	"github.com/10gen/mongoast/ast"
)

// ReferencedVariableRoots returns all the variables referenced in
// the provided expression.
//
// What we mean by roots, is $$foo.bar.car is a valid VariableRef,
// and the root here is $$foo.
func ReferencedVariableRoots(e ast.Expr) []string {
	variables := []string{}
	_, _ = ast.Visit(e, func(v ast.Visitor, n ast.Node) ast.Node {
		switch tn := n.(type) {
		case *ast.VariableRef:
			variables = append(variables, tn.Name)
		case *ast.ArrayIndexRef, *ast.FieldRef, *ast.FieldOrArrayIndexRef:
			rootName, isVariable := GetVariableRefRootNameFromRef(tn.(ast.Expr))
			// If the root is not a variable add it.
			if isVariable {
				variables = append(variables, rootName)
			}
		default:
			_ = n.Walk(v)
		}
		return n
	})

	return variables
}
