package optimizer

import (
	"fmt"

	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/internal/stringutil"

	"github.com/10gen/mongoast/ast"

	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// LetMerging merges nested Let expressions if there is no dependence
// between them. Variables in a nested Let expression that do not depend
// on the Variables of its direct ancestor Let expression will be moved
// into the direct ancestor Let expression. LetMerging will remove any
// Let expression that has 0 bindings.
func LetMerging(pipeline *ast.Pipeline) *ast.Pipeline {
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

		case *ast.Function:
			varsToRemove := make([]string, 0, 2)
			switch typedN.Name {
			case "$reduce":
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
			case "$map", "$filter":
				// In case a Let is nested in a $map or $filter function,
				// temporarily add the "as" field variable to the variable
				// set if it does not already exist.
				asField := getAsField(typedN)
				if !variables.Contains(asField) {
					variables.Add(asField)
					varsToRemove = append(varsToRemove, asField)
				}
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

// getAsField attempts to get the "as" field from a "$map" or "$filter"
// function. This function panics if the ast.Function's Arg is not an
// *ast.Document, if the "as" field is not an *ast.Constant or *ast.Unknown,
// or if the BSON type of the "as" field is not a string. If there is
// no "as" field, "this" is returned. If there is an "as" field and
// all the types are as expected, the string value of the "as" field
// is returned.
func getAsField(mapOrFilter *ast.Function) string {
	argAsDoc, ok := mapOrFilter.Arg.(*ast.Document)
	if !ok {
		panic(fmt.Sprintf("expected *ast.Document argument to %s function, but got %T", mapOrFilter.Name, mapOrFilter.Arg))
	}

	asFieldExpr, ok := argAsDoc.FieldsMap()["as"]
	if !ok {
		// the default for the as field is "this" for $map and $filter
		return "this"
	}

	var asFieldBSONValue bsoncore.Value

	switch typedE := asFieldExpr.(type) {
	case *ast.Constant:
		asFieldBSONValue = typedE.Value
	case *ast.Unknown:
		asFieldBSONValue = typedE.Value
	default:
		panic(fmt.Sprintf("expected *ast.Constant or *ast.Unknown value for \"as\" field, but got %T", typedE))
	}

	asField, ok := asFieldBSONValue.StringValueOK()
	if !ok {
		panic(fmt.Sprintf("expected string value for \"as\" field, but got %v", asFieldBSONValue.Type.String()))
	}

	return asField
}
