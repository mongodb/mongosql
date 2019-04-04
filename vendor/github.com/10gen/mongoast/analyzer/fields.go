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
	// closedRoots are the roots of fields defined by the pipeline at stage x.
	// It is always set after inspecting the children of the stage to correctly
	// handle self-referential stages, e.g.: {$project: {a: {$add: ["$a", 1]}}}.
	// By root we mean, in the case of {$project: {a.b: {$add: ["$a", 1]}}}.
	closedRoots := make(map[string]struct{})
	_, _ = ast.Visit(n, func(v ast.Visitor, n ast.Node) ast.Node {
		if complete {
			return n
		}
		// If this Node is a pipeline, remove anything after a $replaceRoot,
		// because any of those references must come from under the new root.
		if pipeline, ok := n.(*ast.Pipeline); ok {
			foundStage := -1
			for i := range pipeline.Stages {
				if _, ok := pipeline.Stages[i].(*ast.ReplaceRootStage); ok {
					foundStage = i
					// This should stop at the first $replaceRoot found!
					break
				}
			}
			// We do not need to deep copy because this is
			// a read-only analysis, but we do need to create
			// a new pipeline to avoid modifying the previous node.
			if foundStage != -1 {
				newStages := make([]ast.Stage, foundStage+1)
				for i := 0; i <= foundStage; i++ {
					newStages[i] = pipeline.Stages[i]
				}
				n = ast.NewPipeline(newStages...)
			}
		}

		switch tn := n.(type) {
		case *ast.ArrayIndexRef, *ast.FieldRef, *ast.FieldOrArrayIndexRef:
			rootName, isField := GetPathRootFromRef(tn.(ast.Expr))
			// If the root is a variable, we don't care.
			if !isField {
				return n
			}
			if _, ok := closedRoots[rootName]; !ok {
				// Only add the name if the name does not appear
				// in closedFields, which contains the prefixes of
				// every field defined by the pipeline before this stage.
				result = append(result, tn.(*ast.FieldRef))
			}
		case *ast.Unknown:
			hasUnknown = true
		default:
			_ = n.Walk(v)
		}

		switch tn := n.(type) {
		case ast.Stage:
			for _, rootName := range DefinedFields(tn) {
				closedRoots[rootName] = struct{}{}
			}
			complete = IsFieldKiller(tn)
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

// ReferencedFieldRoots returns the unique set of field roots
// as collected from ReferencedFields.
func ReferencedFieldRoots(n ast.Node) ([]string, bool) {
	fields, complete := ReferencedFields(n)

	m := make(map[string]struct{})
	var result []string
	for _, ref := range fields {
		rootName, _ := GetPathRootFromRef(ref)
		if _, ok := m[rootName]; !ok {
			result = append(result, rootName)
			m[rootName] = struct{}{}
		}
	}

	return result, complete
}
