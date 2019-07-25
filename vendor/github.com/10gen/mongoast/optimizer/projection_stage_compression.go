package optimizer

import (
	"fmt"
	"strings"

	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/stringutil"
)

// ProjectionStageCompressionDown removes redundant projection stages ($project and
// $addFields) by compressing adjacent projection stages into one. It also moves
// projection stages next to each other wherever possible in order to increase the number
// of adjacent projection stages by moving them lower in the pipeline.
func ProjectionStageCompressionDown(pipeline *ast.Pipeline) *ast.Pipeline {
	// First do a top to bottom compression pass.
	var optimizedStages []ast.Stage
	// The outer loop attempts to find projection stages to compress or move.
	for i := 0; i < len(pipeline.Stages); i++ {
		top := pipeline.Stages[i]
		// If this is not a projection stage, nothing can be done, just add to the output
		// and continue.
		if !isProjectionStage(top) {
			optimizedStages = append(optimizedStages, top)
			continue
		}

		// Here, the top stage must be a projection, we will attempt to move top down
		// until we find another projection stage, compressing as we go.
		{
			// This scope is just to make clear that j does not live past here, j is only
			// associated with the inner loop.
			j := i + 1
			// The inner loop finds bottom stages for compression or reordering. Despite the
			// fact that this looks like an n^2 algorithm, it is actually linear, because i
			// is set to j at the end of this loop.
			for ; j < len(pipeline.Stages); j++ {
				bottom := pipeline.Stages[j]
				// If the bottom stage is a projection, we attempt compression.
				if isProjectionStage(bottom) {
					compressed, ok := compressProjectionStages(top, bottom)
					if ok {
						// If compression succeeds, we will set top to be compressed, so that
						// we can continue to move the newly compressed stage down, hopefully
						// finding more stages with which to compress.
						top = compressed
					} else {
						// If compression failed, we will make top bottom, in the hopes that bottom
						// can be compressed with a later projection stage.
						optimizedStages = append(optimizedStages, top)
						top = bottom
					}
					continue
				}
				// If bottom is a non-projection stage, we will instead try to reorder it before
				// top, so that we can continue to try to compress top with future stages.
				// We cannot reorder if either stage is a field killer, as this could cause killing
				// fields that need to live, or if the bottom stage uses value defined by the top
				// stage.
				if analyzer.IsFieldKiller(bottom) || analyzer.IsFieldKiller(top) || definesFieldsUsedOrRedefinesFields(top, bottom) {
					// In the case that we cannot reorder, we need to add top to the output pipeline,
					// and set top to bottom so that bottom will be added to the pipeline after the loop.
					optimizedStages = append(optimizedStages, top)
					top = bottom
					break
				}
				// Otherwise, we are free to reorder the stages by immediately appending bottom
				// to our output pipeline.
				optimizedStages = append(optimizedStages, bottom)
			}
			i = j
		}

		// If we get to here, that means there are no more stages to move `top` past or compress with,
		// which means we need to append `top` where we are in the output.
		optimizedStages = append(optimizedStages, top)
	}

	pipeline.Stages = optimizedStages
	return pipeline
}

// ProjectionStageCompressionUp removes redundant projection stages ($project and
// $addFields) by compressing adjacent projection stages into one. It also moves
// projection stages next to each other wherever possible in order to increase the number
// of adjacent projection stages by moving them higher in the pipeline.
func ProjectionStageCompressionUp(pipeline *ast.Pipeline) *ast.Pipeline {
	// First do a bottom to top compression pass.
	var optimizedStages []ast.Stage
	// The outer loop attempts to find projection stages to compress or move.
	for i := len(pipeline.Stages) - 1; i >= 0; i-- {
		bottom := pipeline.Stages[i]
		// If this is not a projection stage, nothing can be done, just add to the output
		// and continue.
		if !isProjectionStage(bottom) {
			optimizedStages = append(optimizedStages, bottom)
			continue
		}

		// Here, the bottom stage must be a projection, we will attempt to move bottom down
		// until we find another projection stage, compressing as we go.
		{
			// This scope is just to make clear that j does not live past here, j is only
			// associated with the inner loop.
			j := i - 1
			// The inner loop finds top stages for compression or reordering. Despite the
			// fact that this looks like an n^2 algorithm, it is actually linear, because i
			// is set to j at the end of this loop.
			for ; j >= 0; j-- {
				top := pipeline.Stages[j]
				// If the top stage is a projection, we attempt compression.
				if isProjectionStage(top) {
					compressed, ok := compressProjectionStages(top, bottom)
					if ok {
						// If compression succeeds, we will set bottom to be compressed, so that
						// we can continue to move the newly compressed stage down, hopefully
						// finding more stages with which to compress.
						bottom = compressed
					} else {
						// If compression failed, we will make bottom top, in the hopes that top
						// can be compressed with a later projection stage.
						optimizedStages = append(optimizedStages, bottom)
						bottom = top
					}
					continue
				}
				// If top is a non-projection stage, we will instead try to reorder it after bottom,
				// so that we can continue to try to compress top with future stages.  If either
				// stage is a field killer, it would be potentially illegal to reorder. If the
				// bottom stage uses fields defined by the top stage, it would also be illegal to
				// reorder.
				if analyzer.IsFieldKiller(top) || analyzer.IsFieldKiller(bottom) || definesFieldsUsedOrRedefinesFields(top, bottom) {
					// In the case that we cannot reorder, we need to add bottom to the output pipeline,
					// and set bottom to top so that top will be added to the pipeline after the loop.
					optimizedStages = append(optimizedStages, bottom)
					bottom = top
					break
				}
				// In the case that we can reorder, we just add top to the output stages.
				optimizedStages = append(optimizedStages, top)
			}
			i = j
		}

		// If we get to here, that means there are no more stages to move `bottom` past or compress with,
		// which means we need to append `bottom` where we are in the output.
		optimizedStages = append(optimizedStages, bottom)
	}

	// Reverse the output because we have been adding the stages in reverse.
	for i, j := 0, len(optimizedStages)-1; i < j; i, j = i+1, j-1 {
		optimizedStages[i], optimizedStages[j] = optimizedStages[j], optimizedStages[i]
	}
	pipeline.Stages = optimizedStages
	return pipeline
}

func isProjectionStage(stage ast.Stage) bool {
	if _, ok := stage.(*ast.ProjectStage); ok {
		return true
	}
	if _, ok := stage.(*ast.AddFieldsStage); ok {
		return true
	}
	return false
}

func definesFieldsUsedOrRedefinesFields(top, bottom ast.Stage) bool {
	topDefs, bottomDefs, bottomUses := stringutil.NewStringSet(), stringutil.NewStringSet(), stringutil.NewStringSet()
	topDefs.AddSlice(analyzer.DefinedFields(top))
	bottomDefs.AddSlice(analyzer.DefinedFields(bottom))
	refFields, _ := analyzer.ReferencedFieldRoots(bottom)
	bottomUses.AddSlice(refFields)
	return topDefs.HasIntersection(bottomUses) || topDefs.HasIntersection(bottomDefs)
}

// compressProjectionStages takes two projection stages (a $project or $addFields) and
// attempts to compress the two stages into one. It returns the compressed stage (if there
// is one) and a boolean indicating whether compression was successful. Note that the
// "top" stage refers to the stage which came first in the pipeline, and "bottom" refers
// to the stage which followed it. For example, in the pipeline
// [
// 		{$project: {a: 1, b: {$literal: 12}}},
//		{$addFields: {c: "hello"}},
// ],
//the $project stage is the "top" stage and the $addFields stage is the "bottom" stage.
func compressProjectionStages(top ast.Stage, bottom ast.Stage) (ast.Stage, bool) {
	// First gather information about whether top and bottom are an inclusion $project.
	// This is used later to figure out whether a field is implicitly excluded by top or
	// bottom, as an inclusion $project implicitly excludes all fields that it does not
	// define.
	var topIsInclusionProjection, bottomIsInclusionProjection bool
	if topProject, ok := top.(*ast.ProjectStage); ok {
		topIsInclusionProjection = topProject.IsInclusion()
	}
	if bottomProject, ok := bottom.(*ast.ProjectStage); ok {
		bottomIsInclusionProjection = bottomProject.IsInclusion()
	}

	// Get a list of all of the items from top and bottom in the order in which they appear
	// in their respective stages.
	orderedItems := append(getOrderedItemNames(top), getOrderedItemNames(bottom)...)

	// Put the top projection items into a map for easy access
	// and manipulation. We will do this for the bottom projection
	// items after we have replaced references in bottom with their definition
	// from top.
	topItems := getItems(top)

	// Iterate through all referenced fields in bottom, checking top for a) anything that
	// would cause us to be unable to compress and b) any new definitions that need to
	// replace the references in bottom. We need to handle references first because we
	// only want the definitions from top to replace references that were originally in
	// the bottom stage.
	var ok bool
	if bottom, ok = handleReferences(topItems, orderedItems, bottom, topIsInclusionProjection); !ok {
		return nil, false
	}

	// Put bottom projection items into a map.
	bottomItems := getItems(bottom)

	// Next iterate through each included/excluded/defined field in bottom, checking top
	// for values that would either cause us to be unable to compress the two stages or to
	// have to have to change the definition of the defined field.
	var seenItems []interface{}
	switch typedBottom := bottom.(type) {
	case *ast.ProjectStage:
		for _, item := range typedBottom.Items {
			// Get all instances of the field, a descendent, or an ancestor from top.
			relatedItems, ok := getRelatedItems(item.GetName(), topItems, orderedItems)
			if !ok {
				return nil, false
			}

			switch typedItem := item.(type) {
			case *ast.IncludeProjectItem:
				if !handleBottomInclude(typedItem, relatedItems, bottomItems, topIsInclusionProjection,
					bottomIsInclusionProjection) {
					return nil, false
				}
			case *ast.ExcludeProjectItem:
				if !handleBottomExclude(typedItem, relatedItems, bottomItems, topIsInclusionProjection) {
					return nil, false
				}
			case *ast.AssignProjectItem:
				if !handleBottomAssign(typedItem, relatedItems, bottomItems, topIsInclusionProjection) {
					return nil, false
				}
			}

			seenItems = append(seenItems, relatedItems...)
		}
	case *ast.AddFieldsStage:
		for _, item := range typedBottom.Items {
			// Get all instances of the field, a descendent, or an ancestor from top.
			relatedItems, ok := getRelatedItems(item.Name, topItems, orderedItems)
			if !ok {
				return nil, false
			}

			if !handleBottomAddFields(item, relatedItems, bottomItems, topIsInclusionProjection) {
				return nil, false
			}

			seenItems = append(seenItems, relatedItems...)
		}
	}

	// At this point we have looked through all defined and referenced fields in bottom
	// and changed their definitions based on related fields we found in top. However,
	// fields included/excluded/defined in top that are NOT related to anything in bottom
	// have not yet been accounted for. If bottom is an inclusion $project, these fields
	// are implicitly excluded, and we can ignore them. Otherwise, we add these fields to
	// bottom.
	if !bottomIsInclusionProjection {
		for _, item := range seenItems {
			switch top.(type) {
			case *ast.ProjectStage:
				delete(topItems, item.(ast.ProjectItem).GetName())
			case *ast.AddFieldsStage:
				delete(topItems, item.(*ast.AddFieldsItem).Name)
			}
		}
		for k, v := range topItems {
			bottomItems[k] = v
		}
	}

	return constructCompressedStage(bottomItems, orderedItems, topIsInclusionProjection, bottomIsInclusionProjection)
}

// getItems takes a projection stage and a map and puts the items from the stage into the map.
func getItems(stage ast.Stage) map[string]interface{} {
	itemMap := make(map[string]interface{})

	switch typedStage := stage.(type) {
	case *ast.ProjectStage:
		for _, item := range typedStage.Items {
			itemMap[item.GetName()] = item
		}
	case *ast.AddFieldsStage:
		for _, item := range typedStage.Items {
			itemMap[item.Name] = item
		}
	default:
		panic("Expected $project or $addFields")
	}

	return itemMap
}

// getRelatedItems takes a map of the items from a projection stage and checks the item
// map to see if it includes/excludes/defines the passed in field (fieldName), one of its
// ancestors, or one or more of its descendents. If the field or one of its ancestors
// exists in the items list, we encounter at most one, since a $project or $addFields
// cannot have fields with conflicting paths.  It can however have multiple descendents,
// as they can be siblings (unrelated to each other). The function is primarily used to
// see the top projection stage contains any items related to a field in the bottom
// projection stage.
func getRelatedItems(fieldName string, items map[string]interface{}, orderedItems []string) ([]interface{}, bool) {
	var relatedItems []interface{}

	// Check if items list contains the field. If so, it cannot contain any other related
	// items, since another instance of the field, an ancestor of the field, or a
	// descendent of the field would conflict. Therefore we can return when and if we find
	// one.
	if items[fieldName] != nil {
		return append(relatedItems, items[fieldName]), true
	}

	// Check if items list contains an ancestor of the field. If so, it cannot contain any
	// other related items, since an instance of the field, another ancestor of the field,
	// or a descendent of the field would conflict. Therefore we can return when and if we
	// find one.
	ancestors := analyzer.GetDotPrefixesOfString(fieldName)
	for _, ancestorName := range ancestors {
		if items[ancestorName] != nil {
			return append(relatedItems, items[ancestorName]), true
		}
	}

	for _, itemName := range orderedItems {
		// Check if items list contains one or more descendents of the field. We iterate
		// through the orderedItems list so that we return the descendents in the order in
		// which they appear in the original stages.
		if analyzer.IsDotPrefixOfString(fieldName, itemName) && items[itemName] != nil {
			relatedItems = append(relatedItems, items[itemName])
		}
		// Check if items list contains a sibling of the field. For now we will stop
		// compression here. TODO: BI-2207 will add coverage for these cases in the
		// handlers. This if statement should be removed once that coverage is added.
		if areSiblingFields(fieldName, itemName) && items[itemName] != nil {
			return nil, false
		}
	}
	return relatedItems, true

}

func areSiblingFields(firstField string, secondField string) bool {
	lastDot := strings.LastIndex(firstField, ".")
	if lastDot == -1 {
		return false
	}
	firstParent := firstField[:lastDot]
	lastDot = strings.LastIndex(secondField, ".")
	if lastDot == -1 {
		return false
	}
	secondParent := secondField[:lastDot]
	return (firstParent == secondParent)

}

// handleBottomExclude checks the fields in the top projection stage to ensure that the
// ExcludeProjectItem in the bottom is handled correctly.
func handleBottomExclude(bottomExclude *ast.ExcludeProjectItem, relatedItems []interface{},
	bottomItems map[string]interface{}, topIsInclusionProjection bool) bool {

	bottomFieldName := bottomExclude.GetName()
	if len(relatedItems) == 0 {
		if topIsInclusionProjection && bottomFieldName != "_id" {
			// If no relatedItems were found in the top projection stage and top is an
			// inclusion $project, top implicitly excludes the field, unless the field is
			// "_id." Therefore the bottom exclude has no effect on the already-excluded
			// field, so we exclude it from bottom as well. Because the top is an
			// inclusion $project, the two stages will compress to an inclusion $project,
			// which excludes fields implicitly. For this reason we delete the exclude
			// from bottom.
			delete(bottomItems, bottomFieldName)
		}
	}

	for _, item := range relatedItems {
		switch topTypedItem := item.(type) {
		case *ast.IncludeProjectItem:
			if topTypedItem.GetName() == bottomFieldName { // field
				if bottomExclude.FieldRef.Parent != nil {
					// Unlike a top-level field, including and excluding the same nested
					// field does NOT merely exclude the field. When included in top and
					// excluded in bottom, if "a" is a document containing "b", the top
					// include redefines "a" as a document containing JUST "b" (excludes
					// all other fields). The bottom exclude then leaves "a" as an empty
					// document. Therefore the only potential solution is to project "a"
					// as an empty document. But we only want to make "a" an empty
					// document if "a" was already a document, and we don't have access to
					// that information.
					return false
				}
				// Including and excluding the same field evaluates to an exclude. Because
				// the top has an include item, it must be in an inclusion $project, so
				// the compressed stage will be an inclusion $project. An inclusion
				// $project excludes values implicitly, so we exclude the field by
				// removing it from bottomItems. The exception to this is "_id", which can
				// only be excluded explicitly.
				if bottomFieldName != "_id" {
					delete(bottomItems, bottomFieldName)
				}
				break
			} else if analyzer.IsDotPrefixOfString(topTypedItem.GetName(), bottomFieldName) { // ancestor
				// There's no way to implicitly exclude a nested field, so we cannot
				// combine these stages into an inclusion projection. Though we could try
				// to compress into an exclusion project stage, which would implicitly
				// include the ancestor, this would require explicitly excluding all other
				// fields of the document (which right now are excluded by the top stage),
				// and we don't know what those are.
				if topIsInclusionProjection {
					return false
				}
				// This handles the very specific case of the ancestor/include item in top
				// being the inclusion of "_id". This is the only way
				// topIsInclusionProjection could be false. In this case, the exclude item
				// in bottom is the exclusion of some descendent of "_id", eg. "_id.a".
				// The query language only allows "_id" to be excluded in an inclusion
				// projection, so if one of its descendents is excluded in bottom, the
				// bottom stage MUST be an exclusion projection. When both top and bottom
				// are exclusion projections, we compress to an exclusion projection.
				// Therefore we keep the exclude item from bottom.
				break
			}
			// descendent
			// Excluding the field excludes the descendent as well. Therefore in our
			// compressed stage we want to exclude the field. Because the top is an
			// inclusion projection, the compressed stage must be an inclusion
			// projection, so we can only exclude implicitly (unless the field is
			// "_id"). If the field is a top-level field, we exclude it by not
			// defining it. If the field is "_id", we keep the exclude item. If the
			// field is nested, there is no way for us to express this, so we cannot
			// compress.
			if bottomExclude.FieldRef.Parent != nil {
				return false
			} else if bottomFieldName != "_id" {
				delete(bottomItems, bottomFieldName)
			}
		case *ast.ExcludeProjectItem:
			if topTypedItem.GetName() == bottomFieldName { // field
				break
			} else if analyzer.IsDotPrefixOfString(topTypedItem.GetName(), bottomFieldName) { // ancestor
				delete(bottomItems, bottomFieldName)
				bottomItems[topTypedItem.GetName()] = topTypedItem
				break
			}
			// descendent
			// do nothing
		case *ast.AssignProjectItem:
			if topTypedItem.GetName() == bottomFieldName { // field
				// We want to exclude the item. If the field is a top-level field, we
				// exclude it by not defining the item. If the field is "_id", we keep the exclude
				// item.
				if bottomFieldName != "_id" {
					delete(bottomItems, bottomFieldName)
					if bottomExclude.FieldRef.Parent != nil {
						// Defining a nested field redefines all instances of the parent
						// field that were literals as documents. Excluding the nested
						// field in bottom leaves those parent fields as empty documents.
						// However, if the parent was already a document with other
						// fields, and one or more of those fields are also included in
						// top, it merely excludes that field. For the document case, we
						// could express this as an include project stage that contains
						// all the included fields, but this is incorrect for the literal
						// case. Therefore we cannot compress.
						return false
					}
				}
				break
			} else if analyzer.IsDotPrefixOfString(topTypedItem.GetName(), bottomFieldName) { // ancestor
				// Defining the ancestor of a field means the field no longer exists. For
				// example, defining "a" means that "a" is now a literal and "a.b" is a
				// nested field of a non-document. Therefore excluding it has no effect,
				// so we can delete the exclude and keep the definition of the ancestor
				// for our compressed stage.
				delete(bottomItems, bottomFieldName)
				bottomItems[topTypedItem.GetName()] = topTypedItem
				break
			}
			// descendent
			// Excluding a field also excludes all of its descendents. Therefore the
			// descendent's definition in top is removed by the exclude. Because the
			// top is an inclusion projection, the two stages will compress into an
			// inclusion projection, so we exclude the field and its descendents
			// implicitly (unless the field is "_id"), meaning we can remove the
			// explicit exclude. If the field is "_id", we can only exclude it
			// explicitly, so we keep the bottom exclude.
			if bottomFieldName != "_id" {
				delete(bottomItems, bottomFieldName)
			}
		case *ast.AddFieldsItem:
			// A $addFields has no mechanism for excluding fields, and an exclusion
			// projection has no mechanism for defining them. Therefore we can only
			// compress an exclusion projection and a $addFields if either all of the
			// definitions or all of the excludes are "voided" by the other stage. For
			// now, we'll just include whichever "wins" between the two (the exclude or
			// the definition), but we'll check later on while we're constructing the
			// final compressed stage and return an error if we find mixed types in the
			// compressed stage.
			if topTypedItem.Name == bottomFieldName { // field
				if bottomExclude.FieldRef.Parent != nil {
					// Unlike $project, using $addFields to define a nested field does NOT
					// exclude the rest of the fields in the parent document. Let's say we
					// define "a.b" in the top $addFields. If "a" is a document, "a.b"
					// will be added/modified, and excluding it in bottom will leave "a"
					// as a document containing all other fields it originally had. If "a"
					// is not a document or doesn't exist, the two stages will result in
					// "a" being an empty document. We'd like to project "a" as an empty
					// document when "a" is not already a document, and exclude "a.b" when
					// it is, but this obviously is not possible.
					return false
				}
				// If a field is defined with $addFields in the top but excluded in
				// bottom, it should be excluded in the bottom, so we keep the exclude
				// item in our compressed stage.
				break
			} else if analyzer.IsDotPrefixOfString(topTypedItem.Name, bottomFieldName) { // ancestor
				// Defining the ancestor in top means the ancestor is no longer a document
				// and the excluded nested field no longer exists. We hope we can compress
				// this into a $addFields so we grab the definition from top.
				delete(bottomItems, bottomFieldName)
				bottomItems[topTypedItem.Name] = topTypedItem
				break
			}
			// descendent
			// Excluding a field excludes all of its descendents, so the definition
			// from top gets excluded.
		}
	}
	return true
}

// handleBottomAssign checks the fields in the top projection stage to ensure that the
// AssignProjectItem in the bottom is handled correctly.
func handleBottomAssign(bottomAssign *ast.AssignProjectItem, relatedItems []interface{},
	bottomItems map[string]interface{}, topIsInclusionProjection bool) bool {
	// An AssignProjectItem in the bottom projection stage always overwrites any related
	// items from the top.  Regardless of whether the field is included, excluded, or
	// defined in top, the bottom assignment sets the field to a completely new value.
	// Even if the bottom field is a nested field, or the top projection stage modifies an
	// ancestor or descendent of the bottom field, the bottom assignment implicitly
	// excludes all related fields, getting rid of any descendents (by defining it as a
	// literal) and overwriting its ancestors.
	return true
}

// handleBottomAddFields checks the fields in the top projection stage to ensure that the
// AddFieldsItem in the bottom is handled correctly.
func handleBottomAddFields(bottomAddFields *ast.AddFieldsItem, relatedItems []interface{},
	bottomItems map[string]interface{}, topIsInclusionProjection bool) bool {

	bottomFieldName := bottomAddFields.Name
	// If there are no related items in top, the field is either implicitly included or
	// excluded depending on the type of the top stage. If the top is an exclusion
	// projection or a $addFields, we want to compress to a $addFields, so we keep the
	// addFields item as-is. If the top is an inclusion projection, we want to compress to
	// an inclusion projection, so we replace the addFields item with a project assign
	// item.
	if len(relatedItems) == 0 {
		if topIsInclusionProjection {
			bottomItems[bottomFieldName] = ast.NewAssignProjectItem(bottomFieldName, bottomAddFields.Expr)
		}
	}

	for _, item := range relatedItems {
		switch topTypedItem := item.(type) {
		case *ast.IncludeProjectItem:
			if topTypedItem.GetName() == bottomFieldName { // field
				bottomItems[bottomFieldName] = ast.NewAssignProjectItem(bottomFieldName, bottomAddFields.Expr)
				break
			} else if analyzer.IsDotPrefixOfString(topTypedItem.GetName(), bottomFieldName) { // ancestor
				// The $project implicitly excludes all fields not named in the stage.
				// There is no way to express this exclusion in a $addFields. Though we
				// could try to project the definition of the field, $project and
				// $addFields are different in that the $project would exclude all other
				// fields of the ancestor, which we don't want. Therefore we cannot
				// compress this.
				return false
			} else { // descendent
				bottomItems[bottomFieldName] = ast.NewAssignProjectItem(bottomFieldName, bottomAddFields.Expr)
			}
		case *ast.ExcludeProjectItem:
			if topTypedItem.GetName() == bottomFieldName { // field
				// The definition in the $addFields redefines the field, so the exclusion
				// makes no difference. Since the top is an exclusion projection stage,
				// the two stages will compress to a $addFields (if possible), so we leave
				// the bottom addFields item alone.
				break
			} else if analyzer.IsDotPrefixOfString(topTypedItem.GetName(), bottomFieldName) { // ancestor
				// The top excludes all fields of the ancestor, if it's a document. The
				// bottom adds back in a sub-field. This is similar to how $project
				// excludes all fields of the document except the one it defines, but the
				// problem is it also excludes all other fields in the document, which we
				// don't want to do here if the top is an exclusion projection. The top is
				// almost certainly an exclusion projection, unless the ancestor is "_id".
				if !topIsInclusionProjection {
					return false
				}
				// If the top contains {_id: 0} in an inclusion projection, then we can
				// compress this into an inclusion projection.
				bottomItems[bottomFieldName] = ast.NewAssignProjectItem(bottomFieldName, bottomAddFields.Expr)
			} else { // nolint:staticcheck
				// Descendent; similar to the first case, we keep the $addFields item.
			}
		case *ast.AssignProjectItem:
			// In all three cases below, the definition in bottom will overwrite the
			// definition in top. Because the top item is an assign project item, we know
			// the top is an inclusion projection, so we replace the bottom item with an
			// assign project item so that we can compress to an inclusion projection.
			if topTypedItem.GetName() == bottomFieldName { // field
				bottomItems[bottomFieldName] = ast.NewAssignProjectItem(bottomFieldName, bottomAddFields.Expr)
				break
			} else if analyzer.IsDotPrefixOfString(topTypedItem.GetName(), bottomFieldName) { // ancestor
				bottomItems[bottomFieldName] = ast.NewAssignProjectItem(bottomFieldName, bottomAddFields.Expr)
				break
			} else { // descendent
				bottomItems[bottomFieldName] = ast.NewAssignProjectItem(bottomFieldName, bottomAddFields.Expr)
			}
		case *ast.AddFieldsItem:
			if topTypedItem.Name == bottomFieldName { // field
				break
			} else if analyzer.IsDotPrefixOfString(topTypedItem.Name, bottomFieldName) { // ancestor
				// Defining the ancestor in top makes the ancestor a non-document,
				// excluding all of its children. Defining the field in bottom redefines
				// the ancestor as a document with one descendent. We cannot express this
				// using a $addFields defining the nested field, as this will NOT exclude
				// the sibling fields. Though $project would, it would also exclude all
				// other fields in the return document, which we don't want.
				return false
			} else { // nolint:staticcheck
				// descendent
			}
		}
	}
	return true
}

// handleBottomInclude checks the fields in the top projection stage to ensure that the
// IncludeProjectItem in the bottom is handled correctly.
func handleBottomInclude(bottomInclude *ast.IncludeProjectItem, relatedItems []interface{},
	bottomItems map[string]interface{}, topIsInclusionProjection bool, bottomIsInclusionProjection bool) bool {

	bottomFieldName := bottomInclude.GetName()
	if len(relatedItems) == 0 {
		if topIsInclusionProjection && bottomFieldName != "_id" {
			// If no relatedItems were found in the top projection stage and top is an
			// inclusion $project, top implicitly excludes the field, unless the field is
			// "_id."  Therefore the bottom include has no effect on the already-excluded
			// field, so we exclude it from bottom by removing it.
			delete(bottomItems, bottomFieldName)
		}
	}

	for _, item := range relatedItems {
		switch topTypedItem := item.(type) {
		case *ast.IncludeProjectItem:
			if topTypedItem.GetName() == bottomFieldName { // field
				break
			} else if analyzer.IsDotPrefixOfString(topTypedItem.GetName(), bottomFieldName) { // ancestor
				break
			} else { // descendent
				// Including a field and its descendent, no matter what the order, has the
				// effect of including just the descendent, as including the descendent
				// implicitly excludes all other descendents of the field (unless they are
				// also explicitly included in top).
				delete(bottomItems, bottomFieldName)
				bottomItems[topTypedItem.GetName()] = topTypedItem
			}
		case *ast.ExcludeProjectItem:
			if topTypedItem.GetName() == bottomFieldName { // field
				if bottomInclude.FieldRef.Parent != nil {
					// Unlike a top-level field, including & excluding the same nested
					// field doesn't just cancel the operations. Say we're looking at
					// nested field "a.b". The top exclude gets rid of field "b", but if
					// "a" is a document, including even a nonexistent field of "a"
					// projects it as an empty document. Therefore the only potential
					// solution is to project "a" as an empty document. But we only want
					// to make "a" an empty document if "a" was already a document, and we
					// don't have access to that information.
					return false
				} else if bottomFieldName == "_id" {
					// _id is the only field implicitly included in a $project, so if it's
					// excluded in top, we want to explicitly specify this in bottom.
					bottomItems[topTypedItem.GetName()] = topTypedItem
				} else {
					// If the field is excluded in one stage and included in the other, is
					// should be excluded in the compressed stage. Because the bottom is
					// an inclusion projection, the compressed stage will also be an
					// inclusion projection, so we exclude the field by removing it.
					delete(bottomItems, bottomFieldName)
				}
				break
			} else if analyzer.IsDotPrefixOfString(topTypedItem.GetName(), bottomFieldName) { // ancestor
				// Excluding a field's ancestor also excludes the field.
				delete(bottomItems, bottomFieldName)
				break
			} else { // descendent
				// There's no way to implicitly exclude a nested field, so we cannot
				// combine these stages into an inclusion projection. Though we could try
				// to compress into an exclusion project stage, which could implicitly
				// include the field while excluding its descendent, this only works if
				// the field is a top-level field. Also, this would require explicitly
				// excluding all other fields of the document (which right now are
				// excluded by the bottom stage), and we don't know what those are.
				return false
			}
		case *ast.AssignProjectItem:
			if topTypedItem.GetName() == bottomFieldName { // field
				// The top redefines the field, so including the field in the bottom just
				// projects the new value.
				bottomItems[topTypedItem.GetName()] = topTypedItem
				break
			} else if analyzer.IsDotPrefixOfString(topTypedItem.GetName(), bottomFieldName) { // ancestor
				// Defining the field's ancestor means it is no longer a document. When we
				// try to project the field, which is now a nested field of a
				// non-document, nothing gets projected.

				// This case does not quite work because it's possible do have a case like this:
				// {$project: {"a": {$cond: ["$b", {"c": {literal: 42}}, {"c": {$literal: 43}}]}}}
				// {$project: {"a.c": 1}}
				// It will remove the definition of a in the top $project. We could eventually
				// do a more complex analysis here, but  for now we are just going to keep the
				// top item in the bottom project, while removing the bottom item,
				// because it is conservatively correct.
				// Previously we just issued this delete:
				//delete(bottomItems, bottomFieldName)
				// Now we also add the top item to the bottom:
				//bottomItems[topTypedItem.GetName()] = topTypedItem
				// TODO: we should actually check to see if a dynamic case is possible, which
				// can be determined by seeing if the AssignProjectItem contains and field
				// references. If there is no dynamic element, the original code of
				// deleting the  bottomFieldName from the bottomItems is still good, while
				// not compressing will be required when there is a dynamic element.
				return false
			} else { // descendent
				// Defining a nested field (this field's descendent for eg.) in a $project
				// implicitly excludes all of its siblings. It redefines the field as a
				// document containing just this descendent and the new definition.
				delete(bottomItems, bottomFieldName)
				bottomItems[topTypedItem.GetName()] = topTypedItem
			}
		case *ast.AddFieldsItem:
			if topTypedItem.Name == bottomFieldName { // field
				if bottomFieldName == "_id" && !bottomIsInclusionProjection {
					// We don't know that the bottom is an inclusion $project if the
					// included field is "_id," since $project allows the explicit
					// inclusion of "_id" in an exclusion $project. For example, we could
					// have {$project: {_id: 1, a: 0, b: 0}} - even though we have an
					// IncludeProjectItem in this stage, this is still an exclusion
					// $project. If the bottom is an exclusion $project, we want to
					// replace the include "_id" with the $addFields definition from top.
					// The problem is the other fields in the exclusion $project are
					// excluded, and there's no way to define "_id" while excluding these.
					// Therefore we cannot compress.
					return false
				}
				// Since the bottom stage is a $project and the top is a $addFields, the
				// resulting stage will be a $project, so we take the definition from the
				// top but convert it to an AssignProjectItem.
				bottomItems[topTypedItem.Name] = ast.NewAssignProjectItem(topTypedItem.Name, topTypedItem.Expr)
				break
			} else if analyzer.IsDotPrefixOfString(topTypedItem.Name, bottomFieldName) { // ancestor
				// Same with the $project case; we define the ancestor as a literal and
				// then try to project the missing field.
				delete(bottomItems, bottomFieldName)
				break
			} else { // descendent
				// The bottom $project implicitly excludes all fields not named in the
				// stage. There is no way to express this exclusion in a $addFields, so we
				// must compress to a $project. Though we could try to project the
				// definition of the descendent rather than defining it with $addFields,
				// $project and $addFields are different in that the $project would
				// exclude all other fields of the parent field. For example, if "a" is
				// included in bottom, and "a.b" is defined with $addFields in top, the
				// top $addFields would modify the value of "a.b" without excluding a
				// field like "a.c" (if such a field exists). Defining it in a $project,
				// however, WOULD exclude it.  Therefore we cannot compress this.
				return false
			}
		}
	}
	return true
}

// ReplaceRef replaces a reference to a field that matches contents in a passed map,
// theta, according to the reference name.
func ReplaceRef(root ast.Node, theta map[string]ast.Node) ast.Node {
	n, _ := ast.Visit(root, func(v ast.Visitor, n ast.Node) ast.Node {
		switch tn := n.(type) {
		// Replaces references to a field within the definition of another field
		case *ast.FieldRef:
			if repl, ok := theta[ast.GetDottedFieldName(tn)]; ok {
				return repl
			}
		case *ast.IncludeProjectItem:
			return n
		default:
			return n.Walk(v)
		}
		return n
	})
	return n
}

// handleReferences looks at all referenced fields in the bottom projection stage and
// replaces them with their definitions in top if they exist. It halts compression under
// certain circumstances.
func handleReferences(topItems map[string]interface{}, orderedItems []string, bottom ast.Stage,
	topIsInclusionProjection bool) (ast.Stage, bool) {

	topDefinedAssigned := make(map[string]ast.Node)

	references, _ := analyzer.ReferencedFields(bottom, false)
	for _, reference := range references {
		bottomReferenceName := ast.GetDottedFieldName(reference)
		relatedItems, ok := getRelatedItems(bottomReferenceName, topItems, orderedItems)
		if !ok {
			return nil, false
		}
		if len(relatedItems) == 0 {
			if topIsInclusionProjection && bottomReferenceName != "_id" {
				// If the top is an inclusion $project and there are no relatedItems found
				// in top, the field is implicitly excluded in top (unless the field is
				// _id). This means that the referenced value is missing. Depending on how
				// the referenced value is used, it may be treated like a NULL value, but
				// since various MongoDB functions handle NULL and missing values
				// differently, we cannot generalize here, so we return without
				// compressing.
				return nil, false
			}
		}

		for _, item := range relatedItems {
			switch topTypedItem := item.(type) {
			case *ast.IncludeProjectItem:
				if topTypedItem.GetName() == bottomReferenceName { // field
					break
				} else if analyzer.IsDotPrefixOfString(topTypedItem.GetName(), bottomReferenceName) { // ancestor
					// Including a field's ancestor also includes the field itself.
					break
				} else { // descendent
					// Including a nested field implicitly excludes all of its siblings.
					// For example, including "a.b" redefines "a" as a document containing
					// just field "b" (if "a" was a document and "b" existed). We could
					// try compressing by replacing references to "$a" with document {"b":
					// "$a.b"}, but if "a" is a non-document, this will define "a" as an
					// empty document rather than a missing value (which is the behavior
					// when we don't compress). Therefore we cannot compress this.
					return nil, false
				}
			case *ast.ExcludeProjectItem:
				if topTypedItem.GetName() == bottomReferenceName { // field
					return nil, false
				} else if analyzer.IsDotPrefixOfString(topTypedItem.GetName(), bottomReferenceName) { // ancestor
					// Excluding a field's ancestor also excludes the field itself.
					return nil, false
				} else { // descendent
					// Excluding a nested field implicitly includes its siblings,
					// redefining the sub-document as a document containing all fields but
					// the one excluded. For example, if "a" contains fields "b" and "c"
					// and "a.b" is excluded in top, we could replace references to "$a"
					// with document {"c": "$a.c"}. However, we don't know which fields
					// are included (only which field is excluded), so there's no way for
					// us to express this.
					return nil, false
				}
			case *ast.AssignProjectItem:
				if topTypedItem.GetName() == bottomReferenceName { // field
					// If the field is assigned a new value in the top projection stage,
					// we need to replace all references to the field in bottom with the
					// new value from top. We add the new value to a map which will be
					// used in a walker function later.
					topDefinedAssigned[bottomReferenceName] = topTypedItem.Expr
					break
				} else if analyzer.IsDotPrefixOfString(topTypedItem.GetName(), bottomReferenceName) { // ancestor
					// Defining a nested field's ancestor makes the ancestor a literal and
					// the nested field a missing value.  This is similar to excluding the
					// field in top - it is missing in bottom and we cannot know how it
					// will be handled in the reference.
					return nil, false
				} else { // descendent
					// Defining the descendent of a field implicitly excludes all of the
					// descendent's siblings, unless they are defined elsewhere in the top
					// projection stage. For example, defining "a.b" and "a.c" in top
					// defines "a" as a document containing fields "b" and "c". References
					// to "$a" in bottom can be replaced with {"b": <definition of a.b>,
					// "c": <definition of <a.c>}.
					definedField := topTypedItem.GetName()[len(bottomReferenceName)+1:]
					if topDefinedAssigned[bottomReferenceName] != nil { // multiple descendents in top
						existingDoc, ok := topDefinedAssigned[bottomReferenceName].(*ast.Document)
						if !ok {
							panic("Expected field in topDefinedAssigned to be assigned as a document")
						}
						topDefinedAssigned[bottomReferenceName] = ast.NewDocument(
							append(existingDoc.Elements,
								ast.NewDocumentElement(definedField, topTypedItem.Expr))...)
					} else { // only one descendent found (so far) in top
						topDefinedAssigned[bottomReferenceName] = ast.NewDocument(
							ast.NewDocumentElement(definedField, topTypedItem.Expr))
					}
				}
			case *ast.AddFieldsItem:
				if topTypedItem.Name == bottomReferenceName { // field
					topDefinedAssigned[bottomReferenceName] = topTypedItem.Expr
					break
				} else if analyzer.IsDotPrefixOfString(topTypedItem.Name, bottomReferenceName) { // ancestor
					// Similar to the $project assign case, defining a field's ancestor
					// makes the field a missing value.
					return nil, false
				} else { // descendent
					// If the field is already a document, the top $addFields
					// adds/modifies the descendent. Unlike $project, it does this while
					// retaining all other fields. Since we don't know what those are, we
					// can't just replace references to the field with a document
					// containing all of its sub-fields.
					return nil, false
				}
			}
		}
	}

	newBottom, ok := ReplaceRef(bottom, topDefinedAssigned).(ast.Stage)
	if !ok {
		panic(fmt.Sprintf("Expected ReplaceRef to return an ast.Stage, but instead it returned %v of type %T",
			newBottom, newBottom))
	}
	return newBottom, true
}

func getOrderedItemNames(stage ast.Stage) []string {
	switch typedStage := stage.(type) {
	case *ast.ProjectStage:
		names := make([]string, len(typedStage.Items))
		for i, item := range typedStage.Items {
			names[i] = item.GetName()
		}
		return names
	case *ast.AddFieldsStage:
		names := make([]string, len(typedStage.Items))
		for i, item := range typedStage.Items {
			names[i] = item.Name
		}
		return names
	default:
		panic("Not a project or addFields stage")
	}
}

// constructCompressedStage takes a map of projection items and constructs a $project or
// $addFields based on its contents.
func constructCompressedStage(compressedItems map[string]interface{}, orderedItemNames []string,
	topIsInclusionProjection bool, bottomIsInclusionProjection bool) (ast.Stage, bool) {

	if len(compressedItems) == 0 {
		return ast.NewProjectStage(ast.NewIncludeProjectItem(ast.NewFieldRef("_id", nil))), true
	} else if len(compressedItems) == 1 && compressedItems["_id"] != nil {
		if _, ok := compressedItems["_id"].(*ast.ExcludeProjectItem); ok &&
			(topIsInclusionProjection || bottomIsInclusionProjection) {
			// In various places in our projection item handlers, we add the exclude item
			// {_id: 0} to our compressed stage in order to exclude "_id". If neither
			// stage is an inclusion projection, this works well. The issue is if {_id:
			// 0} ends up being the only definition in the compressed stage, this is
			// interpreted as an exclusion projection even when the compressed stage
			// should be an inclusion projection. Consider the following query:
			//
			// [{$project: {_id: 0}}, {$project: {_id: 1}}]
			//
			// For this query, we grab the exclude item from top and put it in our
			// compressed stage. The problem is that when our stages are separate, the
			// bottom stage containing {_id: 1} implicitly excludes all other fields in
			// the document, whereas the compressed stage, [{$project: {_id: 0}}], does
			// not. Therefore we cannot compress this. We stop compression here rather
			// than earlier because we do not necessarily know how many items the final
			// compressed stage will have. Consider another example:
			//
			// [{$project: {_id: 0, a: {$literal: 12}}}, {$project: {_id: 1, b: 1, c: 1}}]
			//
			// Even though both the top and bottom stages define fields other than "_id",
			// the other fields are implicitly excluded and get removed as we scan the
			// stages. For this reason we wait until we've gathered all of the items for
			// the compressed stage before checking for this special case.
			return nil, false
		}
	}

	var isAddFields bool
	for _, v := range compressedItems {
		if _, ok := v.(*ast.AddFieldsItem); ok {
			isAddFields = true
		}
	}

	if isAddFields {
		addFieldsItems := make([]*ast.AddFieldsItem, len(compressedItems))
		i := 0
		// Right now we have a list of all the items from our original stages in order
		// (ordered items from top first, followed by ordered items from bottom). For our
		// compressed stage, we want the fields ordered in this way, except if a related
		// field in bottom "replaced" a field in top, that field should follow the top
		// order. For example,
		//
		// top        = {$project: {a: 1, b: 1}}
		// bottom     = {$project: {b: 1, d: 1, a.c: 1, a.b: 1}}
		// compressed = {$project: {a.c: 1, a.b: 1, b: 1, d: 1}}
		//
		// To achieve this ordering, we iterate through our ordered list (which contains
		// all fields from both original stages) and see if there are any related fields
		// in our map of compressedItems. We add the fields to our final list of items in
		// this order.
		for _, itemName := range orderedItemNames {
			relatedItems, _ := getRelatedItems(itemName, compressedItems, orderedItemNames)
			for _, relatedItem := range relatedItems {
				if typedItem, ok := relatedItem.(*ast.AddFieldsItem); ok {
					addFieldsItems[i] = typedItem
					i++
					// If the relatedItem is an ancestor or descendent of the current
					// field, we will encounter that ancestor or descendent later in
					// orderedItemNames, since orderedItemNames contains ALL items from
					// BOTH of the original stages. Because of this, we remove the
					// relatedItem from compressedItems so that we don't add it to
					// addFieldsItems a second time when we encounter the field later on.
					delete(compressedItems, typedItem.Name)
				} else {
					// mixed types should only occur when we're trying to compress a
					// $addFields and an exclusion projection
					return nil, false
				}
			}
		}
		return ast.NewAddFieldsStage(addFieldsItems...), true
	}

	projectItems := make([]ast.ProjectItem, len(compressedItems))
	i := 0
	for _, itemName := range orderedItemNames {
		relatedItems, _ := getRelatedItems(itemName, compressedItems, orderedItemNames)
		for _, relatedItem := range relatedItems {
			if typedItem, ok := relatedItem.(ast.ProjectItem); ok {
				projectItems[i] = typedItem
				i++
				delete(compressedItems, typedItem.GetName())
			} else {
				return nil, false
			}
		}

	}
	return ast.NewProjectStage(projectItems...), true
}
