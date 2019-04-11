package optimizer

import (
	"github.com/10gen/mongoast/ast"
)

// LetMinimization removes any let binding that is used 0 or 1 times or
// is a binding for a Constant or Ref. It removes any Let expression that
// has 0 bindings.
func LetMinimization(pipeline *ast.Pipeline) *ast.Pipeline {
	out, _ := ast.Visit(pipeline, func(v ast.Visitor, n ast.Node) ast.Node {
		n = n.Walk(v)
		switch typedN := n.(type) {
		case *ast.Let:
			return minimizeLetVariables(typedN)
		default:
			return n
		}
	})

	return out.(*ast.Pipeline)
}

// letVariableInfo is used to track LetVariable information while
// minimizing bindings.
type letVariableInfo struct {
	count uint32
	expr  ast.Expr
	keep  bool

	isShadowed      bool
	bindingRequired bool
}

func minimizeLetVariables(let *ast.Let) ast.Expr {
	variables := make(map[string]*letVariableInfo)
	for _, letVar := range let.Variables {
		bindingRequired := requiresLetBinding(letVar.Expr)
		variables[letVar.Name] = &letVariableInfo{
			count:           0,
			expr:            letVar.Expr,
			keep:            false,
			isShadowed:      false,
			bindingRequired: bindingRequired,
		}
	}

	// First, count the variable uses
	_, _ = ast.Visit(let.Expr, func(v ast.Visitor, n ast.Node) ast.Node {
		switch typedN := n.(type) {
		case *ast.Let:
			// For a nested Let expression, we need to count the variable uses
			// in both the nested Variables and in the nested "in" expression.
			// If a nested variable shadows one of the outer variables we should
			// not count uses of that variable while walking the "in" expression;
			// those are references to the shadow, NOT the outer variable.
			// The markShadowsInLetAndContinueWalk function handles this for us.
			return markShadowsInLetAndContinueWalk(v, typedN, variables)

		case *ast.Function:
			// $map, $filter, and $reduce define new variables, and those may
			// shadow some of the existing variables. If a nested variable
			// defined by one of these functions shadows one of the outer
			// variables, we should not count uses of that variable while
			// walking the function; those are references to the shadow, NOT
			// the outer variable.
			// The markShaodwsInFunctionAndContinueWalk function handles this
			// for us.
			return markShadowsInFunctionAndContinueWalk(v, typedN, variables)

		case *ast.VariableRef:
			// only count variable references that are known
			if info, ok := variables[typedN.Name]; ok && info.bindingRequired && !info.isShadowed {
				info.count++

				// only keep variables that require a let binding
				// and are used more than once
				info.keep = info.count > 1
			}
			return n

		default:
			return n.Walk(v)
		}
	})

	replacingFieldRefParent := false

	// Next, replace variables that are eligible for replacement
	newIn, _ := ast.Visit(let.Expr, func(v ast.Visitor, n ast.Node) ast.Node {
		switch typedN := n.(type) {
		case *ast.Let:
			// For a nested Let expression, we should not replace variable
			// references in the nested "in" expression if they are shadows
			// of outer variables.
			return markShadowsInLetAndContinueWalk(v, typedN, variables)

		case *ast.Function:
			// For $map, $filter, and $reduce functions, we should not replace
			// variable references nested in the function if they are shadows
			// of outer variables.
			return markShadowsInFunctionAndContinueWalk(v, typedN, variables)

		case *ast.FieldRef:
			// Special handling for FieldRefs: we should only substitute
			// FieldRef Parents with other FieldRefs or with VariableRefs.
			replacingFieldRefParent = true
			n = n.Walk(v)
			replacingFieldRefParent = false
			return n

		case *ast.VariableRef:
			if info, ok := variables[typedN.Name]; ok && !info.keep && !info.isShadowed {
				// If we are replacing a FieldRef's Parent, we should only
				// substitute this VariableRef if the substitution is itself
				// another VariableRef or a FieldRef.
				if replacingFieldRefParent {
					switch info.expr.(type) {
					case *ast.FieldRef, *ast.VariableRef:
						return info.expr
					}

					// If it is neither of those, we'll need to keep this let-binding.
					info.keep = true
					return n
				}

				return info.expr
			}
		}

		return n.Walk(v)
	})

	// Finally, collect the remaining LetVariables
	newVariables := make([]*ast.LetVariable, 0, len(let.Variables))
	for _, oldVariable := range let.Variables {
		if info := variables[oldVariable.Name]; info.keep {
			newVariables = append(newVariables, ast.NewLetVariable(oldVariable.Name, info.expr))
		}
	}

	if len(newVariables) == 0 {
		return newIn.(ast.Expr)
	}

	return ast.NewLet(newVariables, newIn.(ast.Expr))
}

// requiresLetBinding returns false if an ast.Expr
// does not need to be stored as a let variable.
func requiresLetBinding(e ast.Expr) bool {
	switch e.(type) {
	case *ast.Constant, *ast.Unknown, ast.Ref:
		return false
	default:
		return true
	}
}

// markShadowsInLetAndContinueWalk walks the provided Let expression's
// variables and "in" expression using the provided visitor. Before
// walking the "in" expression, this function marks variables in the
// provided map as shadowed if a variable with the same name appears
// in the provided Let's list of variables.
func markShadowsInLetAndContinueWalk(v ast.Visitor, let *ast.Let, variables map[string]*letVariableInfo) ast.Node {
	// shadows is a map of the shadowed variables encountered in this
	// nested Let. It maps from variable name to whether or not that
	// variable was previously shadowed.
	shadows := make(map[string]bool)

	for i, letVar := range let.Variables {
		// walk the nested variable
		let.Variables[i] = letVar.Walk(v).(*ast.LetVariable)

		if info, isShadow := variables[letVar.Name]; isShadow {
			// mark if this variable was or was not shadowed before
			shadows[letVar.Name] = info.isShadowed

			// mark the outer variable as shadowed if necessary
			info.isShadowed = true
		}
	}

	// walk the nested "in" expression with the appropriate "isShadowed" information
	let.Expr = let.Expr.Walk(v).(ast.Expr)

	// mark each shadowed variable as it was before
	for name, wasShadowed := range shadows {
		variables[name].isShadowed = wasShadowed
	}

	return let
}

// markShadowsInFunctionAndContinueWalk walks the provided Function using
// the provided visitor. Before walking the it, this function checks if
// the Function is a variable-defining function ($map, $filter, $reduce)
// and marks variables in the provided map as shadowed if a variable with
// the same name is defined by the function.
func markShadowsInFunctionAndContinueWalk(v ast.Visitor, f *ast.Function, variables map[string]*letVariableInfo) ast.Node {
	var n ast.Node
	switch f.Name {
	case "$reduce":
		thisInfo, thisIsShadowed := variables["this"]
		valueInfo, valueIsShadowed := variables["value"]

		// remember if these variables were or were not shadowed already
		thisWasShadowed := thisInfo != nil && thisInfo.isShadowed
		valueWasShadowed := valueInfo != nil && valueInfo.isShadowed

		// if "this" and/or "value" shadow an outer variable,
		// mark them as shadowed before walking the function
		if thisIsShadowed {
			thisInfo.isShadowed = true
		}
		if valueIsShadowed {
			valueInfo.isShadowed = true
		}

		// walk the function
		n = f.Walk(v)

		// mark them as they were before
		if thisIsShadowed {
			thisInfo.isShadowed = thisWasShadowed
		}
		if valueIsShadowed {
			valueInfo.isShadowed = valueWasShadowed
		}

	case "$map", "$filter":
		asField := getAsField(f)

		info, isShadowed := variables[asField]
		wasShadowed := info != nil && info.isShadowed

		// if the as field shadows an outer variable, mark that
		// variable as shadowed before walking the function
		if isShadowed {
			info.isShadowed = true
		}

		n = f.Walk(v)

		// mark it as it was before
		if isShadowed {
			info.isShadowed = wasShadowed
		}

	default:
		// no other functions define new variables
		n = f.Walk(v)
	}

	return n
}
