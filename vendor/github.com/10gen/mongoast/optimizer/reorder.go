package optimizer

import (
	"strings"

	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/eval"
	"github.com/10gen/mongoast/parser"
)

// Reorder attempts to move stages into their optimal position.
func Reorder(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {
	if len(pipeline.Stages) < 2 {
		return pipeline
	}

	newStages := make([]ast.Stage, len(pipeline.Stages))
	copy(newStages, pipeline.Stages)

	referencedFields := make([][][]string, len(newStages))
	for i, stage := range newStages {
		if match, ok := stage.(*ast.MatchStage); ok {
			referencedFields[i] = getReferencedMatchFields(match)
		}
	}

	for b := 1; b < len(newStages); b++ {
		for a := b; a > 0; a-- {
			x, y, ok := trySwapStages(newStages[a-1], newStages[a], referencedFields[a])
			if !ok {
				break
			}

			// If we substituted into a $match stage, we may have generated an AST
			// that's not valid match language. Therefore, we attempt to deparse
			// the newly generated stages and ensure that there are no errors. If
			// there are, then cancel the flip.
			if _, err := parser.DeparseStageErr(x); err != nil {
				break
			}
			if _, err := parser.DeparseStageErr(y); err != nil {
				break
			}

			// If we substituted into a $match stage, we need to update the
			// referencedFields for that stage.
			if match, ok := y.(*ast.MatchStage); ok {
				referencedFields[a] = getReferencedMatchFields(match)
			}

			newStages[a], newStages[a-1] = x, y
			referencedFields[a], referencedFields[a-1] = referencedFields[a-1], referencedFields[a]
		}
	}

	return ast.NewPipeline(newStages...)
}

func trySwapStages(a, b ast.Stage, referencedFields [][]string) (ast.Stage, ast.Stage, bool) {
	switch ts := b.(type) {
	case *ast.LimitStage, *ast.SampleStage, *ast.SkipStage:
		switch a.(type) {
		case *ast.AddFieldsStage, *ast.ProjectStage, *ast.ReplaceRootStage:
			return a, b, true
		}
	case *ast.MatchStage:
		return trySwapMatch(a, ts, referencedFields)
	}
	return nil, nil, false
}

func trySwapMatch(a ast.Stage, b *ast.MatchStage, referencedMatchFields [][]string) (ast.Stage, ast.Stage, bool) {
	// If the $match stage references a field that is introduced by the
	// stage with which it's being flipped, we need to substitute the
	// definition of the field in place of the reference when it's flipped.
	switch ts := a.(type) {
	case *ast.GroupStage:
		for _, dottedReferencedField := range referencedMatchFields {
			if dottedReferencedField[0] != "_id" {
				return nil, nil, false
			}
		}
		newNode, ok := eval.SubstituteFields(b, map[string]ast.Expr{
			"_id": ts.By,
		})
		if !ok {
			return nil, nil, false
		}
		return a, newNode.(ast.Stage), true
	case *ast.ProjectStage:
		if canSwapProjectAsIs(ts, referencedMatchFields) {
			return a, b, true
		} else if !ts.IsInclusion() {
			// Any exclusion that can be flipped should be covered by canFlipProjectAsIs.
			return nil, nil, false
		}
		fields := make(map[string]ast.Expr)
		idExcluded := false
		for _, pi := range ts.Items {
			switch tpi := pi.(type) {
			case *ast.AssignProjectItem:
				if strings.Contains(tpi.Name, ".") {
					return nil, nil, false
				}
				fields[tpi.Name] = tpi.Expr
			case *ast.IncludeProjectItem:
				if tpi.Ref.ParentExpr() != nil {
					return nil, nil, false
				}
				fields[tpi.GetName()] = tpi.Ref
			case *ast.ExcludeProjectItem:
				if tpi.GetName() == "_id" && tpi.Ref.ParentExpr() == nil {
					idExcluded = true
				}
			}
		}
		if !idExcluded {
			fields["_id"] = ast.NewFieldRef("_id", nil)
		}
		newNode, ok := eval.SubstituteFields(b, fields)
		if !ok {
			return nil, nil, false
		}
		return a, newNode.(ast.Stage), true
	case *ast.AddFieldsStage:
		if canSwapAddFieldsAsIs(ts, referencedMatchFields) {
			return a, b, true
		}
		fields := make(map[string]ast.Expr)
		for _, afi := range ts.Items {
			if strings.Contains(afi.Name, ".") {
				return nil, nil, false
			}
			fields[afi.Name] = afi.Expr
		}
		// Deliberately ignore substitution errors here. If there's a
		// reference to a field not defined by the $addFields stage, just
		// leave it as is.
		newNode, _ := eval.SubstituteFields(b, fields)
		return a, newNode.(ast.Stage), true
	case *ast.ReplaceRootStage:
		var newNode ast.Node
		if doc, ok := ts.NewRoot.(*ast.Document); ok {
			fields := make(map[string]ast.Expr)
			for _, e := range doc.Elements {
				fields[e.Name] = e.Expr
			}
			newNode, ok = eval.SubstituteFields(b, fields)
			if !ok {
				return nil, nil, false
			}
		} else {
			newNode = eval.SubstituteRoot(b, ts.NewRoot)
		}
		return a, newNode.(ast.Stage), true
	case *ast.UnwindStage:
		fieldParts := ast.GetDottedFields(ts.Path)

		for _, field := range referencedMatchFields {
			if likeComponentsMatch(fieldParts, field) || likeComponentsMatch(ast.SplitDottedFieldPath(ts.IncludeArrayIndex), field) {
				return nil, nil, false
			}
		}
		return a, b, true
	case *ast.SortStage, *ast.MatchStage:
		return a, b, true
	case *ast.LookupStage:
		dottedAs := strings.Split(ts.As, ".")
		for _, field := range referencedMatchFields {
			// If the as field is smaller than the field referenced in the match,
			// it can change the output by eliminating fields that the match referenced.
			// If the as field is longer than the field referenced in the match it can
			// change the output by creating fields that were not originally there that
			// affect the output of the match.
			// If the fields are the same length, the $lookup can obviously change the
			// actual field the match references
			if likeComponentsMatch(dottedAs, field) {
				return nil, nil, false
			}
		}
		return a, b, true
	}
	return nil, nil, false
}

func canSwapProjectAsIs(ps *ast.ProjectStage, referencedFields [][]string) bool {
	if ps.IsExclusion() {
		for _, item := range ps.Items {
			exclude, _ := item.(*ast.ExcludeProjectItem)
			dottedExcludeField := strings.Split(ast.GetDottedFieldName(exclude.Ref), ".")
			for _, dottedReferencedField := range referencedFields {
				if likeComponentsMatch(dottedExcludeField, dottedReferencedField) {
					return false
				}
			}
		}
		return true
	}
	for _, dottedReferencedField := range referencedFields {
		if len(dottedReferencedField) == 1 && dottedReferencedField[0] == "_id" {
			for _, item := range ps.Items {
				exclude, ok := item.(*ast.ExcludeProjectItem)
				if ok && exclude.GetName() == "_id" {
					return false
				}
				assign, ok := item.(*ast.AssignProjectItem)
				if ok && assign.GetName() == "_id" {
					return false
				}
			}
			continue
		}
		included := false
		for _, item := range ps.Items {
			include, ok := item.(*ast.IncludeProjectItem)
			if !ok {
				continue
			}
			dottedIncludeField := strings.Split(include.GetName(), ".")
			if len(dottedIncludeField) <= len(dottedReferencedField) &&
				likeComponentsMatch(dottedIncludeField, dottedReferencedField) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}
	return true
}

func canSwapAddFieldsAsIs(afs *ast.AddFieldsStage, referencedFields [][]string) bool {
	definedFields := analyzer.DefinedFieldsFullPath(afs)
	for _, definedField := range definedFields {
		dottedDefinedField := strings.Split(definedField, ".")
		for _, dottedReferencedField := range referencedFields {
			if likeComponentsMatch(dottedDefinedField, dottedReferencedField) {
				return false
			}
		}
	}
	return true
}

func likeComponentsMatch(a []string, b []string) bool {
	minLength := len(a)
	if len(b) < minLength {
		minLength = len(b)
	}
	for i := 0; i < minLength; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func getReferencedMatchFields(stage *ast.MatchStage) [][]string {
	fields, _ := analyzer.ReferencedFields(stage, false)
	dottedFields := make([][]string, len(fields))
	for j, field := range fields {
		dottedFields[j] = strings.Split(ast.GetDottedFieldName(field), ".")
	}

	return dottedFields
}
