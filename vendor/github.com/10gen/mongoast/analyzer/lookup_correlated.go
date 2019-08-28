package analyzer

import (
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/stringutil"
)

// IsLookupCorrelated returns whether or not the lookup is correlated.
func IsLookupCorrelated(n *ast.LookupStage) bool {
	if n.LocalField != nil {
		return true
	}
	if n.Let == nil {
		return false
	}

	variables := stringutil.NewStringSet()
	for _, item := range n.Let {
		if ContainsFieldRef(item.Expr) {
			variables.Add(item.Name)
		}
	}
	return UsesVariables(n.Pipeline, variables)
}

type usesVariablesVisitor struct {
	variables     *stringutil.StringSet
	usesVariables bool
}

func (v *usesVariablesVisitor) Visit(node ast.Node) ast.Node {
	switch tn := node.(type) {
	case *ast.Let:
		remaining := stringutil.NewStringSet()
		remaining.AddSet(v.variables)
		for _, letVariable := range tn.Variables {
			variables := stringutil.NewStringSet()
			variables.Add(letVariable.Name)
			if UsesVariables(tn.Expr, variables) {
				v.Visit(letVariable)
			}
			remaining.Remove(letVariable.Name)
		}
		v.usesVariables = v.usesVariables || UsesVariables(tn.Expr, remaining)
		return node
	case *ast.Map:
		v.Visit(tn.Input)
		remaining := stringutil.NewStringSet()
		remaining.AddSet(v.variables)
		if tn.As != "" {
			remaining.Remove(tn.As)
		} else {
			remaining.Remove("this")
		}
		v.usesVariables = v.usesVariables || UsesVariables(tn.In, remaining)
		return node
	case *ast.Filter:
		// Similarly to $map, even if the "as" variable isn't used in the condition,
		// the input array can still affect the output (because it affects the size of the
		// output array). We cannot tell if it will affect the size of the output array without
		// knowing if the condition is true or not, so we will just say that the variable is used
		// if it appears in the "input" expression.
		v.Visit(tn.Input)
		remaining := stringutil.NewStringSet()
		remaining.AddSet(v.variables)
		if tn.As != "" {
			remaining.Remove(tn.As)
		} else {
			remaining.Remove("this")
		}
		v.usesVariables = v.usesVariables || UsesVariables(tn.Cond, remaining)
		return node
	case *ast.Reduce:
		variables := stringutil.NewStringSet()
		variables.Add("this")
		// The input array is only used if $$this is used in the "in" expression
		if UsesVariables(tn.In, variables) {
			v.Visit(tn.Input)
		}
		variables = stringutil.NewStringSet()
		variables.Add("value")
		// The initialValue is only used if $$value is used in the "in" expression
		if UsesVariables(tn.In, variables) {
			v.Visit(tn.InitialValue)
		}
		// Reduce can only override the variables "$$this" and "$$value"
		remaining := stringutil.NewStringSet()
		remaining.AddSet(v.variables)
		remaining.Remove("this")
		remaining.Remove("value")
		v.usesVariables = v.usesVariables || UsesVariables(tn.In, remaining)
		return node
	case *ast.VariableRef:
		v.usesVariables = v.usesVariables || v.variables.Contains(tn.Name)
		return node
	}
	return node.Walk(v)
}

// UsesVariables returns whether any of the variables are used in the pipeline.
func UsesVariables(node ast.Node, variables *stringutil.StringSet) bool {
	visitor := usesVariablesVisitor{variables, false}
	(&visitor).Visit(node)
	return visitor.usesVariables
}

type containsFieldRefVisitor struct {
	hasFieldRef bool
}

func (v *containsFieldRefVisitor) Visit(node ast.Node) ast.Node {
	switch node.(type) {
	case *ast.FieldRef:
		v.hasFieldRef = true
		return node
	}
	return node.Walk(v)
}

// ContainsFieldRef returns whether the expression has a field reference in it.
func ContainsFieldRef(node ast.Node) bool {
	visitor := containsFieldRefVisitor{false}
	(&visitor).Visit(node)
	return visitor.hasFieldRef
}
