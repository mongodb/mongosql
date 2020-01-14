package optimizer

import (
	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/internal/stringutil"

	"github.com/10gen/mongoast/ast"
)

// LetMerging merges nested Let expressions if there is no dependence
// between them. Variables in a nested Let expression that do not depend
// on the Variables of its direct ancestor Let expression will be moved
// into the direct ancestor Let expression. LetMerging will remove any
// Let expression that has 0 bindings.
func LetMerging(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {
	out, _ := ast.Visit(pipeline, func(v ast.Visitor, n ast.Node) ast.Node {
		switch typedN := n.(type) {
		case *ast.Let:
			n = mergeLets(typedN)
		}
		return n.Walk(v)
	})

	return out.(*ast.Pipeline)
}

// mergeLets attempts to merge all directly nested Let expressions into the
// provided Let expression.
func mergeLets(let *ast.Let) ast.Expr {
	// Collect variables in the current Let in a map for easy lookup.
	// newVariables is necessary to maintain deterministic order of variables.
	newVariables := make([]*ast.LetVariable, len(let.Variables))
	variables := stringutil.NewStringSet()
	for i, letVar := range let.Variables {
		newVariables[i] = letVar
		variables.Add(letVar.Name)
	}

	// merge is the visit function that finds directly nested Lets
	// and merges their variables into the current Let's set. A
	// "directly nested Let" is any Let expression that exists in
	// the tree for which there is no other Let expression between
	// it and the root. (The root is the "let" argument to this
	// function).
	var merge func(v ast.Visitor, n ast.Node) ast.Node
	merge = func(v ast.Visitor, n ast.Node) ast.Node {
		// Do not attempt to merge Lets nested in lazy expressions. This
		// is necessary because let bindings nested in lazily evaluated
		// expressions may or may not be executed by mongo. Merging them
		// out of the lazy expression would eagerly evaluate them which
		// is an incorrect behavioral change.
		if expr, isExpr := n.(ast.Expr); isExpr && analyzer.HasLazyArgumentSemantics(expr) {
			return n
		}

		switch typedN := n.(type) {
		case *ast.Let:
			newNestedVariables := make([]*ast.LetVariable, 0, len(typedN.Variables))

			for _, letVar := range typedN.Variables {
				// do not merge shadowed variables
				if variables.Contains(letVar.Name) {
					newNestedVariables = append(newNestedVariables, letVar)
					continue
				}

				isDependent := false
				// Walk the LetVariable to find any variable references to the
				// current variable set. If this LetVariable has a VariableRef
				// to a variable in the current set, it is considered dependent
				// on the current set and cannot be merged.
				_, _ = ast.Visit(letVar, func(v ast.Visitor, n ast.Node) ast.Node {
					_ = n.Walk(v)

					switch typedN := n.(type) {
					case *ast.VariableRef:
						if variables.Contains(typedN.Name) {
							isDependent = true
						}
					}

					return n
				})

				if isDependent {
					newNestedVariables = append(newNestedVariables, letVar)
				} else {
					// merge the independent nested variable
					// into the current set of variables.
					variables.Add(letVar.Name)
					newVariables = append(newVariables, letVar)
				}
			}

			if len(newNestedVariables) == 0 {
				// If the nested Let has no variables remaining, it should be
				// replaced with its "in" expression. That means the current
				// Let inherits the "in" expression and it too must be merged.
				return merge(v, typedN.Expr)
			}

			return ast.NewLet(newNestedVariables, typedN.Expr)

		case *ast.Map:
			varsToRemove := make([]string, 0, 1)

			asField := typedN.As
			if !variables.Contains(asField) {
				variables.Add(asField)
				varsToRemove = append(varsToRemove, asField)
			}

			// walk the function with the relevant new variables in scope
			n = n.Walk(v)

			// remove any newly added variables
			for _, varToRemove := range varsToRemove {
				variables.Remove(varToRemove)
			}

			return n

		case *ast.Filter:
			varsToRemove := make([]string, 0, 1)

			asField := typedN.As
			if !variables.Contains(asField) {
				variables.Add(asField)
				varsToRemove = append(varsToRemove, asField)
			}

			// walk the function with the relevant new variables in scope
			n = n.Walk(v)

			// remove any newly added variables
			for _, varToRemove := range varsToRemove {
				variables.Remove(varToRemove)
			}

			return n

		case *ast.Reduce:
			varsToRemove := make([]string, 0, 2)
			// In case a Let is nested in a $reduce function, temporarily
			// add the $$this and $$value variables to the variable set
			// if they do not already exist.
			if !variables.Contains("this") {
				variables.Add("this")
				varsToRemove = append(varsToRemove, "this")
			}
			if !variables.Contains("value") {
				variables.Add("value")
				varsToRemove = append(varsToRemove, "value")
			}

			// walk the function with the relevant new variables in scope
			n = n.Walk(v)

			// remove any newly added variables
			for _, varToRemove := range varsToRemove {
				variables.Remove(varToRemove)
			}

			return n

		default:
			return n.Walk(v)
		}
	}

	// Find directly nested Let expressions and merge independent variables.
	newExpr, _ := ast.Visit(let.Expr, merge)

	return ast.NewLet(newVariables, newExpr.(ast.Expr))
}
