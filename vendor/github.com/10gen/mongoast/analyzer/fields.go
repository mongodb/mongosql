package analyzer

import (
	"github.com/10gen/mongoast/ast"
)

// ReferencedFields returns all the fields referenced in
// the provided node. In addition, it returns a boolean indicating
// whether this field list is complete.
//
// For instance, a pipeline that has no projecting stage
// ($project, $group, etc...) cannot only request the subset
// of fields used in the pipeline as other fields may be
// used by a client. Hence, the boolean would be returned
// as false indicating that the field list, which is accurate
// for the fields required to evaluate this node, is incomplete
// given the entire picture.
//
// For all intents and purposes, the boolean will always be
// true for anything other than an *ast.Pipeline, for which it
// can be false based on the stages present.
func ReferencedFields(n ast.Node) ([]*ast.FieldRef, bool) {
	hasUnknown := false
	complete := false
	var result []*ast.FieldRef
	_, _ = ast.Visit(n, func(v ast.Visitor, n ast.Node) ast.Node {
		if complete {
			return n
		}

		switch tn := n.(type) {
		case *ast.ExcludeProjectItem:
			return n
		case *ast.FieldRef:
			result = append(result, tn)
		default:
			_ = n.Walk(v)
		}

		switch n.(type) {
		case *ast.GroupStage, *ast.ProjectStage:
			complete = true
		case *ast.Unknown:
			hasUnknown = true
		}

		return n
	})

	if _, ok := n.(*ast.Pipeline); !ok && !hasUnknown {
		complete = true
	}

	return result, complete
}

// ReferencedFieldNames returns the unique set of field names
// as collected from ReferencedFields.
func ReferencedFieldNames(n ast.Node) ([]string, bool) {
	fields, complete := ReferencedFields(n)

	m := make(map[string]struct{})
	var result []string
	for _, ref := range fields {
		if _, ok := m[ref.Name]; !ok {
			result = append(result, ref.Name)
			m[ref.Name] = struct{}{}
		}
	}

	return result, complete
}
