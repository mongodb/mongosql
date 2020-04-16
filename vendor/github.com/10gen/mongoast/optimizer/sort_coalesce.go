package optimizer

import (
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/astprint"
)

// SortCoalescing combines adjacent $sort or $sortByExpr stages in the
// pipeline. If one or both of the stages is a $sortByExpr stage, the
// they will combine into one $sortByExpr stage.
func SortCoalescing(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {
	var lastSortStage ast.SortingStage
	allStages := make([]ast.Stage, 0, len(pipeline.Stages))

	for i := range pipeline.Stages {
		switch ts := pipeline.Stages[i].(type) {
		case ast.SortingStage:
			if lastSortStage != nil {
				lastSortStage = mergeSortStages(lastSortStage, ts)
				continue
			}
			lastSortStage = ts
		default:
			if lastSortStage != nil {
				allStages = append(allStages, lastSortStage)
			}
			lastSortStage = nil
			allStages = append(allStages, pipeline.Stages[i])
		}
	}

	// Add sort stage if last stage in pipeline.
	if lastSortStage != nil {
		allStages = append(allStages, lastSortStage)
	}

	pipeline.Stages = allStages
	return pipeline
}

// mergeSortStages prepends the SortItems contained in the later stage 'b'
// to the slice of SortItems contained in the earlier stage 'a'.
func mergeSortStages(a ast.SortingStage, b ast.SortingStage) ast.SortingStage {
	newItems := make([]*ast.SortItem, len(b.SortItems()))
	_ = copy(newItems, b.SortItems())

	// If we're moving up a sort on a field ref that is also sorted on by the
	// earlier stage, we need to not include the earlier sort item.
	for _, itemA := range a.SortItems() {
		found := false
		for _, itemB := range b.SortItems() {
			if astprint.String(itemA.Expr) == astprint.String(itemB.Expr) {
				found = true
			}
		}
		if !found {
			newItems = append(newItems, itemA)
		}

	}

	// Return a SortByExpr if either stage is a SortByExpr.
	_, isByExprA := a.(*ast.SortByExprStage)
	_, isByExprB := b.(*ast.SortByExprStage)
	if isByExprA || isByExprB {
		return ast.NewSortByExprStage(newItems...)
	}

	return ast.NewSortStage(newItems...)
}
