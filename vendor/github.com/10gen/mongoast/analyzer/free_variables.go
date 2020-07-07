package analyzer

import (
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/stringutil"
)

// globalSystemVariables contains the MongoDB system variables that should not be considered free
// variables, since they are always bound.
var globalSystemVariables = stringutil.NewStringSet(
	"CLUSTER_TIME",
	"CURRENT",
	"DESCEND",
	"KEEP",
	"NOW",
	"PRUNE",
	"REMOVE",
	"ROOT",
)

// StageHasFreeVars returns true if the stage has any free variables
// (i.e. the stage is correlated) and false if it does not.
func StageHasFreeVars(stage ast.Stage) bool {
	return HasFreeVars(stage, globalSystemVariables)
}

// HasFreeVars returns true if the node has any free variables and false
// if it does not. boundVariables are the in-scope local variables at a
// given point while visiting the node.
func HasFreeVars(node ast.Node, boundVariables *stringutil.StringSet) bool {
	visitor := freeVariablesVisitor{boundVariables, false}
	(&visitor).Visit(node)
	return visitor.usesFreeVariables
}

type freeVariablesVisitor struct {
	boundVariables    *stringutil.StringSet
	usesFreeVariables bool
}

func (v *freeVariablesVisitor) Visit(node ast.Node) ast.Node {
	// If we already found a free var, just return.
	if v.usesFreeVariables {
		return node
	}
	switch tn := node.(type) {
	case *ast.Let:
		localBoundVariables := stringutil.NewStringSet()
		localBoundVariables.AddSet(v.boundVariables)
		for _, letVariable := range tn.Variables {
			// Report if any of the let binding expressions have free variables. We do not want to
			// bind any of the let variables for this check, so we use the top level boundVariables.
			v.usesFreeVariables = v.usesFreeVariables || HasFreeVars(letVariable, v.boundVariables)
			// Add each variable name to our local boundVariables.
			localBoundVariables.Add(letVariable.Name)
		}
		// Now check the let expression (the in part) with the localBoundVariables.
		v.usesFreeVariables = v.usesFreeVariables || HasFreeVars(tn.Expr, localBoundVariables)
		return node
	case *ast.Map:
		v.usesFreeVariables = v.usesFreeVariables || HasFreeVars(tn.Input, v.boundVariables)
		localBoundVariables := stringutil.NewStringSet()
		localBoundVariables.AddSet(v.boundVariables)
		if tn.As != "" {
			localBoundVariables.Add(tn.As)
		} else {
			localBoundVariables.Add("this")
		}
		v.usesFreeVariables = v.usesFreeVariables || HasFreeVars(tn.In, localBoundVariables)
		return node
	case *ast.Filter:
		v.usesFreeVariables = v.usesFreeVariables || HasFreeVars(tn.Input, v.boundVariables)
		localBoundVariables := stringutil.NewStringSet()
		localBoundVariables.AddSet(v.boundVariables)
		if tn.As != "" {
			localBoundVariables.Add(tn.As)
		} else {
			localBoundVariables.Add("this")
		}
		v.usesFreeVariables = v.usesFreeVariables || HasFreeVars(tn.Cond, localBoundVariables)
		return node
	case *ast.Reduce:
		v.usesFreeVariables = v.usesFreeVariables || HasFreeVars(tn.Input, v.boundVariables)
		localBoundVariables := stringutil.NewStringSet()
		localBoundVariables.AddSet(v.boundVariables)
		localBoundVariables.Add("this")
		localBoundVariables.Add("value")
		v.usesFreeVariables = v.usesFreeVariables || HasFreeVars(tn.In, localBoundVariables)
		return node
	case *ast.VariableRef:
		v.usesFreeVariables = v.usesFreeVariables || !v.boundVariables.Contains(tn.Name)
		return node
	}
	return node.Walk(v)
}
