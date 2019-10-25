package optimizer

import (
	"strings"

	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/ast"
)

// Reorder attempts to move stages into their optimal position.
func Reorder(pipeline *ast.Pipeline) *ast.Pipeline {
	if len(pipeline.Stages) < 2 {
		return pipeline
	}

	newStages := make([]ast.Stage, len(pipeline.Stages))
	copy(newStages, pipeline.Stages)

	referencedFields := make([][][]string, len(newStages))
	for i, stage := range newStages {
		if _, ok := stage.(*ast.MatchStage); ok {
			fields, _ := analyzer.ReferencedFields(stage, false)
			dottedFields := make([][]string, len(fields))
			for j, field := range fields {
				dottedFields[j] = strings.Split(ast.GetDottedFieldName(field), ".")
			}
			referencedFields[i] = dottedFields
		}
	}

	for b := 1; b < len(newStages); b++ {
		for a := b; a > 0; a-- {
			if !shouldFlip(newStages[a-1], newStages[a], referencedFields[a]) {
				break
			}
			newStages[a], newStages[a-1] = newStages[a-1], newStages[a]
			referencedFields[a], referencedFields[a-1] = referencedFields[a-1], referencedFields[a]
		}
	}

	return ast.NewPipeline(newStages...)
}

func shouldFlip(a, b ast.Stage, referencedFields [][]string) bool {
	switch b.(type) {
	case *ast.LimitStage, *ast.SampleStage, *ast.SkipStage:
		switch a.(type) {
		case *ast.AddFieldsStage, *ast.ProjectStage, *ast.ReplaceRootStage:
			return true
		}
	case *ast.MatchStage:
		switch ts := a.(type) {
		case *ast.SortStage, *ast.MatchStage:
			return true
		case *ast.LookupStage:
			dottedAs := strings.Split(ts.As, ".")
			for _, field := range referencedFields {
				// If the as field is smaller than the field referenced in the match,
				// it can change the output by eliminating fields that the match referenced.
				// If the as field is longer than the field referenced in the match it can
				// change the output by creating fields that were not originally there that
				// affect the output of the match.
				// If the fields are the same length, the $lookup can obviously change the
				// actual field the match references
				if likeComponentsMatch(dottedAs, field) {
					return false
				}
			}
			return true
		case *ast.AddFieldsStage:
			definedFields := analyzer.DefinedFieldsFullPath(ts)
			for _, definedField := range definedFields {
				dottedDefinedField := strings.Split(definedField, ".")
				for _, dottedReferencedField := range referencedFields {
					if likeComponentsMatch(dottedDefinedField, dottedReferencedField) {
						return false
					}
				}
			}
			return true
		case *ast.ProjectStage:
			if ts.IsInclusion() {
				for _, dottedReferencedField := range referencedFields {
					if len(dottedReferencedField) == 1 && dottedReferencedField[0] == "_id" {
						for _, item := range ts.Items {
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
					for _, item := range ts.Items {
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
			for _, item := range ts.Items {
				exclude, _ := item.(*ast.ExcludeProjectItem)
				dottedExcludeField := strings.Split(ast.GetDottedFieldName(exclude.FieldRef), ".")
				for _, dottedReferencedField := range referencedFields {
					if likeComponentsMatch(dottedExcludeField, dottedReferencedField) {
						return false
					}
				}
			}
			return true
		}
	}
	return false
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
