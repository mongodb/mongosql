package eval

import (
	"strconv"

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

func extractField(expr ast.Expr, fieldName string) ast.Expr {
	// It's possible for a document to be a bsoncore.Document inside of an
	// ast.Constant or an ast.Document. This function handles extracting
	// fields from both of these cases.
	switch te := expr.(type) {
	case *ast.Constant:
		doc, ok := te.Value.DocumentOK()
		if !ok {
			return ast.NewConstant(bsonutil.Null())
		}
		value, err := doc.LookupErr(fieldName)
		if err != nil {
			return ast.NewConstant(bsonutil.Null())
		}
		return ast.NewConstant(value)
	case *ast.Document:
		for _, e := range te.Elements {
			if e.Name == fieldName {
				return e.Expr
			}
		}
	}
	return nil
}

func extractIndex(expr ast.Expr, index int) ast.Expr {
	// It's possible for an array to be a bsoncore.Array inside of an
	// ast.Constant or an ast.Array. This function handles extracting elements
	// from both of these cases.
	switch te := expr.(type) {
	case *ast.Constant:
		arr, ok := te.Value.ArrayOK()
		if !ok {
			return nil
		}
		value, err := arr.IndexErr(uint(index))
		if err != nil {
			return nil
		}
		return ast.NewConstant(value.Value())
	case *ast.Array:
		if index < 0 {
			index += len(te.Elements)
		}
		if index < 0 || index >= len(te.Elements) {
			return nil
		}
		return te.Elements[index]
	}
	return nil
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
		if newExpr := extractField(substExpr, tn.Name); newExpr != nil {
			return newExpr
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

// SubstituteVariables replaces each reference to one of the specified
// variables in the tree with the corresponding AST fragment.
func SubstituteVariables(node ast.Node, variables map[string]ast.Expr) ast.Node {
	return (&variableSubstitutionVisitor{variables}).Visit(node)
}

type fieldSubstitutionVisitor struct {
	fields map[string]ast.Expr
	ok     bool
}

func (v *fieldSubstitutionVisitor) Visit(node ast.Node) ast.Node {
	switch tn := node.(type) {
	case *ast.FieldRef:
		if tn.Parent != nil {
			substExpr := v.Visit(tn.Parent).(ast.Expr)
			if substExpr == tn.Parent {
				return node
			}
			if newExpr := extractField(substExpr, tn.Name); newExpr != nil {
				return newExpr
			}
		} else {
			value, ok := v.fields[tn.Name]
			if ok {
				return value
			}
			v.ok = false
		}
	case *ast.ArrayIndexRef:
		indexConst, ok := v.Visit(tn.Index).(*ast.Constant)
		if ok {
			index, ok := bsonutil.AsInt32OK(indexConst.Value)
			if ok {
				substExpr := v.Visit(tn.Parent).(ast.Expr)
				if newExpr := extractIndex(substExpr, int(index)); newExpr != nil {
					return newExpr
				}
			}
		}
	case *ast.FieldOrArrayIndexRef:
		substExpr := v.Visit(tn.Parent).(ast.Expr)
		if newExpr := extractIndex(substExpr, int(tn.Number)); newExpr != nil {
			return newExpr
		} else if newExpr := extractField(substExpr, strconv.Itoa(int(tn.Number))); newExpr != nil {
			return newExpr
		}
	}
	return node.Walk(v)
}

// SubstituteFields replaces each reference to one of the specified fields
// in the tree with the corresponding AST fragement.
func SubstituteFields(node ast.Node, fields map[string]ast.Expr) (ast.Node, bool) {
	v := &fieldSubstitutionVisitor{
		fields: fields,
		ok:     true,
	}
	newNode := v.Visit(node)
	return newNode, v.ok
}

// SubstituteRoot modifies every field reference in tree with a nil parent to
// have the specified AST fragment as its parent.
func SubstituteRoot(node ast.Node, newRoot ast.Expr) ast.Node {
	out, _ := ast.Visit(node, func(v ast.Visitor, n ast.Node) ast.Node {
		switch tn := n.(type) {
		case *ast.FieldRef:
			if tn.Parent == nil {
				return ast.NewFieldRef(tn.Name, newRoot)
			}
		case *ast.FieldOrArrayIndexRef:
			if tn.Parent == nil {
				return ast.NewFieldOrArrayIndexRef(tn.Number, newRoot)
			}
		}
		return n.Walk(v)
	})
	return out
}
