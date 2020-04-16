package optimizer

import (
	"strings"

	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/eval"
	"github.com/10gen/mongoast/internal/stringutil"
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
	case ast.SortingStage:
		return trySwapSort(a, ts)
	}
	return nil, nil, false
}

func trySwapSort(a ast.Stage, b ast.SortingStage) (ast.Stage, ast.Stage, bool) {
	switch ts := a.(type) {
	// Stages that unconditionally preserve sort order.
	// Do not swap with $match stage, since we want $match stages to come
	// before $sort stages and $match stages will move up the pipeline,
	// making room for $sort.
	case *ast.CountStage:
	case *ast.OutStage:
	case *ast.SampleStage:
	// Stages that preserve sort order if they don't alter fields that will
	// be sorted by the later sort stage.
	case *ast.AddFieldsStage:
		changedRefs := stringutil.NewStringSet()
		for _, item := range ts.Items {
			changedRefs.Add(item.Name)
		}
		if doesSortRef(b, changedRefs) {
			return nil, nil, false
		}
	case *ast.LookupStage:
		changedRefs := stringutil.NewStringSet()
		changedRefs.Add(ts.As)
		if doesSortRef(b, changedRefs) {
			return nil, nil, false
		}
	case *ast.ProjectStage:
		excludedRefs := stringutil.NewStringSet()
		includedRefs := stringutil.NewStringSet()
		for _, item := range ts.Items {
			switch ti := item.(type) {
			case *ast.AssignProjectItem:
				return nil, nil, false
			case *ast.ExcludeProjectItem:
				includedRefs.Remove(ti.GetName())
				excludedRefs.Add(ti.GetName())
			case *ast.IncludeProjectItem:
				excludedRefs.Remove(ti.GetName())
				includedRefs.Add(ti.GetName())
			default:
				return nil, nil, false
			}
		}
		// Handle implicitly included _id field when any fields are included.
		if includedRefs.Len() > 0 && !excludedRefs.Contains("_id") {
			includedRefs.Add("_id")
		}
		if doesSortRef(b, excludedRefs) || doesSortOtherRef(b, includedRefs) {
			return nil, nil, false
		}
	case *ast.UnwindStage:
		changedRefs := stringutil.NewStringSet()
		changedRefs.Add(ts.IncludeArrayIndex)
		changedRefs.Add(ast.GetDottedFieldName(ts.Path))
		if doesSortRef(b, changedRefs) {
			return nil, nil, false
		}
	// Assume all other stages will not preserve sort order.
	default:
		return nil, nil, false
	}

	return a, b, true
}

// doesSortRef returns true if any of the fields referenced by the specified
// SortStage are among the field refs in the specified slice of refs.
func doesSortRef(s ast.SortingStage, refs *stringutil.StringSet) bool {
	if refs.Len() == 0 {
		return false
	}
	sortedRefs, ok := getSortedPrefixesByRef(s)
	if !ok {
		// Couldn't determine whether the sort uses one of the refs,
		// so err on the side of saying it does.
		return true
	}
	for _, sortedRef := range sortedRefs {
		for _, sortedPrefix := range sortedRef {
			if refs.Contains(sortedPrefix) {
				return true
			}
		}
	}

	return false
}

// doesSortOtherRef returns true if any of fields referenced by the specified
// SortStage are NOT among the field refs in the specified slice of refs.
func doesSortOtherRef(s ast.SortingStage, refs *stringutil.StringSet) bool {
	if refs.Len() == 0 {
		return false
	}
	sortedRefs, ok := getSortedPrefixesByRef(s)
	if !ok {
		// Couldn't determine whether the sort uses other refs,
		// so err on the side of saying it does.
		return true
	}
	for _, sortedRef := range sortedRefs {
		found := false
		for _, sortedPrefix := range sortedRef {
			if refs.Contains(sortedPrefix) {
				found = true
			}
		}
		if !found {
			return true
		}
	}

	return false
}

func getSortedPrefixesByRef(s ast.SortingStage) ([][]string, bool) {
	prefixesByRef := make([][]string, 0, len(s.SortItems()))
	for _, item := range s.SortItems() {
		refs, ok := analyzer.ReferencedFieldNames(item.Expr)
		if !ok {
			return nil, false
		}
		for _, ref := range refs {
			prefixesByRef = append(prefixesByRef, analyzer.GetDotPrefixesOfString(ref))
		}
	}
	return prefixesByRef, true
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
	case *ast.SortStage, *ast.SortByExprStage, *ast.MatchStage:
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
