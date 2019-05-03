package eval

import (
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
)

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
		remaining := make(map[string]ast.Expr)
		for key, value := range v.variables {
			remaining[key] = value
		}
		for _, letVariable := range tn.Variables {
			delete(remaining, letVariable.Name)
		}
		newExpr := SubstituteVariables(tn.Expr, remaining).(ast.Expr)
		if newExpr != tn.Expr {
			return ast.NewLet(tn.Variables, newExpr)
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
