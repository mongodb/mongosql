package optimizer

import (
	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/stringutil"
)

var idRef = ast.NewFieldRef("_id", nil)

// DeadCodeElimination removes dead code.
func DeadCodeElimination(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {
	// Indices of the stages to keep.
	keepIndices := make([]bool, len(pipeline.Stages))
	keepCount := 0
	liveFields := stringutil.NewStringSet()
	i := len(pipeline.Stages) - 1
	// Find the last field killing stage.
	for ; i >= 0; i-- {
		keepIndices[i] = true
		keepCount++
		if analyzer.IsFieldKiller(pipeline.Stages[i]) {
			keepList, _ := analyzer.ReferencedFieldRoots(pipeline.Stages[i])

			liveFields.AddSlice(keepList)
			i--
			break
		}
	}
	// setKeep sets that we want to keep the current stage, and makes sure
	// that these two pieces of state used to track this fact are always
	// kept in sync.
	setKeep := func(idx int) {
		keepIndices[idx] = true
		keepCount++
	}
	// Remove anything not needed before the last field killing stage, that is,
	// any field definitions not used by the last field killing stage, transitively.
	// The only caveat here is that any cardinality altering stage must be kept, and
	// any field used by a cardinality altering stage is also considered live and must
	// be kept.
	for ; i >= 0; i-- {
		// Any sort stage must be kept, and any field it depends on must be kept.
		if analyzer.IsSortStage(pipeline.Stages[i]) {
			refs, _ := analyzer.ReferencedFieldRoots(pipeline.Stages[i])
			liveFields.AddSlice(refs)
			setKeep(i)
			continue
		}
		// Remove dead definitions from cardinality altering stages, and update liveFields.
		if analyzer.IsCardinalityAlteringStage(pipeline.Stages[i]) {
			removeDeadDefinitionsFromCardinalityAlteringStagesAndUpdateLiveFields(
				pipeline.Stages[i], liveFields)
			setKeep(i)
			continue
		}
		// Remove dead definitions from each stage that isn't a sort or cardinality altering, and
		// update liveFields.
		keep := removeDeadDefinitionsAndUpdateLiveFields(pipeline.Stages[i], liveFields)
		if keep {
			setKeep(i)
		}
	}
	// Now generate the new pipeline keeping only those stages that we need to keep.
	newStages := make([]ast.Stage, 0, keepCount)
	for j, keep := range keepIndices {
		if keep {
			newStages = append(newStages, pipeline.Stages[j])
		}
	}
	pipeline.Stages = newStages
	return pipeline
}

// removeDeadDefinitions removes dead definitions from stages and returns a bool
// value where true means keep the stage, and false means remove the stage.
func removeDeadDefinitionsAndUpdateLiveFields(stage ast.Stage, liveFields *stringutil.StringSet) bool {
	// TODO: It would also be nice if all stages that define values just had
	// "DefinitionItems" instead of AddFields vs GroupItem vs FacetItem, etc.
	definitions := analyzer.DefinedFields(stage)
	// Indices of the definitions to keep.
	keepIndices := make([]bool, len(definitions))
	keepCount := 0
	// I do think that perhaps Definitions and RemoveDefinition should be added
	// to the ast interface, but that can be done in the future.
	for i, def := range definitions {
		if liveFields.Contains(def) {
			keepIndices[i] = true
			keepCount++
		}
	}
	// We can't remove anything, just add to the live set and return.
	if keepCount == len(definitions) {
		refs, _ := analyzer.ReferencedFieldRoots(stage)
		// If this stage is a FieldKiller, it kills previous live values
		// (previous is backwards, recall: we start from the end).
		if analyzer.IsFieldKiller(stage) {
			liveFields.RemoveAll()
		}
		liveFields.AddSlice(refs)
		return true
	}
	switch typedStage := stage.(type) {
	case *ast.AddFieldsStage:
		// We removed everything, just return false. There is no need to update a stage
		// we will just remove.
		if keepCount == 0 {
			// This is the only place where liveFields will not be added, because the
			// entire stage is being removed, and we would be adding unneeded
			// fields if we added refs, because we have no bothered to remove all the
			// definitions from this stage.
			return false
		}
		newItems := make([]*ast.AddFieldsItem, 0, keepCount)
		for i, keep := range keepIndices {
			if keep {
				newItems = append(newItems, typedStage.Items[i])
			}
		}
		typedStage.Items = newItems
	case *ast.LookupStage:
		// We can only get here if we want to remove the $lookup 'as' field,
		// in which case it is safe to remove the entire stage.
		return false
	// of the if cases above.
	case *ast.ProjectStage:
		if keepCount == 0 {
			// If this is an exclusion $project, just keep it as is.
			if typedStage.IsExclusion() {
				return true
			}
			// Since we aren't keeping anything, we need to make a $project stage that
			// kills all the currently live values. UNLESS there are no live values.
			if liveFields.Len() == 0 {
				return false
			}
			excludes := make([]ast.ProjectItem, liveFields.Len())
			for i, fieldName := range liveFields.SortedSlice() {
				excludes[i] = ast.NewExcludeProjectItem(ast.NewFieldRef(fieldName, nil))
			}
			typedStage.Items = excludes
			// $project kills previous live values
			// (previous is backwards, recall: we start from the end).
			// Note, if this actually happens this pipeline generates empty
			// documents, but it's best to be correct.
			liveFields.RemoveAll()
			return true
		}
		var newItems []ast.ProjectItem
		// Technically this has to be 1 or 0 at this point, since we are keeping
		// any inclusions, any other $project would be malformed and caught
		// as invalid by the parser.
		hasIDExclude := len(typedStage.ExcludeItems()) == 1
		if hasIDExclude {
			newItems = make([]ast.ProjectItem, 0, keepCount+1)
		} else {
			newItems = make([]ast.ProjectItem, 0, keepCount)
		}
		nonExcludeItems := typedStage.NonExcludeItems()
		for i, keep := range keepIndices {
			if keep {
				newItems = append(newItems, nonExcludeItems[i])
			}
		}
		if hasIDExclude {
			newItems = append(newItems, ast.NewExcludeProjectItem(idRef))
		}
		typedStage.Items = newItems
	case *ast.ReplaceRootStage:
		if keepCount == 0 {
			// $replaceRoot kills previous live values
			// (previous is backwards, recall: we start from the end).
			// Note, if this actually happens this pipeline generates empty
			// documents, but it's best to be correct.
			liveFields.RemoveAll()
			// If we remove a $replaceRoot stage it can technically cause issues since it
			// should kill everything not included, instead we modify the $replaceRoot to be
			// {$replaceRoot: {newRoot:_{}}}, to at least make the pipeline more efficient.
			typedStage.NewRoot = ast.NewDocument()
			return true
		}
		// If we are here, the root must be a document, and we can prune unnecessary fields.
		newElements := make([]*ast.DocumentElement, 0, keepCount)
		doc := typedStage.NewRoot.(*ast.Document)
		for i, keep := range keepIndices {
			if keep {
				newElements = append(newElements, doc.Elements[i])
			}
		}
		doc.Elements = newElements
		typedStage.NewRoot = doc
	}
	// We always want to add all references unless the stage is going to be removed.
	refs, _ := analyzer.ReferencedFieldRoots(stage)
	if analyzer.IsFieldKiller(stage) {
		// If this stage is a FieldKiller, it kills previous live values
		// (previous is backwards, recall: we start from the end).
		liveFields.RemoveAll()
	}
	liveFields.AddSlice(refs)
	return true
}

// removeDeadDefinitionsFromCardinalityAlteringStages removes dead definitions from cardinality
// altering stages. Since we never want to remove these stages, there is no need to return a bool.
func removeDeadDefinitionsFromCardinalityAlteringStagesAndUpdateLiveFields(
	stage ast.Stage,
	liveFields *stringutil.StringSet) {
	defer func() {
		// We always want to add any remaining references after removing definitions.
		refs, _ := analyzer.ReferencedFieldRoots(stage)
		if analyzer.IsFieldKiller(stage) {
			liveFields.RemoveAll()
		}
		liveFields.AddSlice(refs)
	}()
	definitions := analyzer.DefinedFields(stage)
	// Indices of the definitions to keep.
	keepIndices := make([]bool, len(definitions))
	keepCount := 0
	// I do think that perhaps Definitions and RemoveDefinition should be added
	// to the ast interface, but that can be done in the future.
	for i, def := range definitions {
		if liveFields.Contains(def) {
			keepIndices[i] = true
			keepCount++
		}
	}
	// We can't remove anything, just add to the live set and return.
	if keepCount == len(definitions) {
		return
	}
	switch typedStage := stage.(type) {
	case *ast.BucketStage:
		// This might allocate one extra index, not worried about it.
		newOutput := make([]*ast.GroupItem, 0, keepCount)
		// Slice out $_id, since it is not part of them items.
		// _id will always be the first defined field.
		for i, keep := range keepIndices[1:] {
			if keep {
				newOutput = append(newOutput, typedStage.Output[i])
			}
		}
		typedStage.Output = newOutput
	case *ast.BucketAutoStage:
		// This might allocate one extra index, not worried about it.
		newOutput := make([]*ast.GroupItem, 0, keepCount)
		// Slice out $_id, since it is not part of them items.
		// _id will always be the first defined field.
		for i, keep := range keepIndices[1:] {
			if keep {
				newOutput = append(newOutput, typedStage.Output[i])
			}
		}
		typedStage.Output = newOutput
	case *ast.FacetStage:
		newItems := make([]*ast.FacetItem, 0, keepCount)
		for i, keep := range keepIndices {
			if keep {
				newItems = append(newItems, typedStage.Items[i])
			}
		}
		typedStage.Items = newItems
	case *ast.GroupStage:
		// This might allocate one extra index, not worried about it.
		newItems := make([]*ast.GroupItem, 0, keepCount)
		// Slice out $_id, since it is not part of them items.
		// _id will always be the first defined field.
		for i, keep := range keepIndices[1:] {
			if keep {
				newItems = append(newItems, typedStage.Items[i])
			}
		}
		typedStage.Items = newItems
	// CASE *ast.LimitStage, no definitions in $limit
	// CASE *ast.MatchStage, no definitions in $match
	// CASE *ast.SampleStage, no definitions in $sample
	// CASE *ast.SkipStage, no definitions in $skip
	case *ast.UnwindStage:
		// We will always keep the $unwind path, but we might be able to
		// remove the 'includeArrayIndex'.
		if len(keepIndices) == 2 && !keepIndices[1] {
			typedStage.IncludeArrayIndex = ""
		}
	}
}
