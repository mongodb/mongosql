package eval

import (
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
)

func copyVariables(variables map[string]ast.Expr) map[string]ast.Expr {
	copy := make(map[string]ast.Expr)
	for key, value := range variables {
		copy[key] = value
	}
	return copy
}

type variableSubstitutionVisitor struct {
	variables map[string]ast.Expr
}

func (v *variableSubstitutionVisitor) Visit(node ast.Node) ast.Node {
	switch tn := node.(type) {
	case *ast.FieldRef:
		if tn.Parent == nil {
			return node
		}
		substExpr := v.Visit(tn.Parent).(ast.Expr)
		if substExpr == tn.Parent {
			return node
		}
		switch te := substExpr.(type) {
		case *ast.Constant:
			doc, ok := te.Value.DocumentOK()
			if !ok {
				return ast.NewConstant(bsonutil.Null())
			}
			value, err := doc.LookupErr(tn.Name)
			if err != nil {
				return ast.NewConstant(bsonutil.Null())
			}
			return ast.NewConstant(value)
		case *ast.Document:
			for _, e := range te.Elements {
				if e.Name == tn.Name {
					return e.Expr
				}
			}
		}
		return node
	case *ast.Let:
		remaining := copyVariables(v.variables)
		for _, letVariable := range tn.Variables {
			// delete aliases
			delete(remaining, letVariable.Name)

			// substitute outer variables into this Let's variables
			letVariable.Expr = SubstituteVariables(letVariable.Expr, v.variables).(ast.Expr)
		}
		newExpr := SubstituteVariables(tn.Expr, remaining).(ast.Expr)
		if newExpr != tn.Expr {
			return ast.NewLet(tn.Variables, newExpr)
		}
		return node
	case *ast.Map:
		remaining := copyVariables(v.variables)
		if tn.As != "" {
			delete(remaining, tn.As)
		} else {
			delete(remaining, "this")
		}
		newInput := v.Visit(tn.Input).(ast.Expr)
		newIn := SubstituteVariables(tn.In, remaining).(ast.Expr)
		if newInput != tn.Input || newIn != tn.In {
			return ast.NewMap(newInput, tn.As, newIn)
		}
		return node
	case *ast.Filter:
		remaining := copyVariables(v.variables)
		if tn.As != "" {
			delete(remaining, tn.As)
		} else {
			delete(remaining, "this")
		}
		newInput := v.Visit(tn.Input).(ast.Expr)
		newCond := SubstituteVariables(tn.Cond, remaining).(ast.Expr)
		if newInput != tn.Input || newCond != tn.Cond {
			return ast.NewFilter(newInput, tn.As, newCond)
		}
		return node
	case *ast.Reduce:
		remaining := copyVariables(v.variables)
		delete(remaining, "this")
		delete(remaining, "value")
		newInput := v.Visit(tn.Input).(ast.Expr)
		newInitialValue := v.Visit(tn.InitialValue).(ast.Expr)
		newIn := SubstituteVariables(tn.In, remaining).(ast.Expr)
		if newInput != tn.Input || newInitialValue != tn.InitialValue || newIn != tn.In {
			return ast.NewReduce(newInput, newInitialValue, newIn)
		}
		return node
	case *ast.VariableRef:
		value, ok := v.variables[tn.Name]
		if ok {
			return value
		}
	}
	return node.Walk(v)
}

func SubstituteVariables(node ast.Node, variables map[string]ast.Expr) ast.Node {
	return (&variableSubstitutionVisitor{variables}).Visit(node)
}
