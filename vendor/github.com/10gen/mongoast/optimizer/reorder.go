package optimizer

import "github.com/10gen/mongoast/ast"

// Reorder attempts to move stages into their optimal position.
func Reorder(pipeline *ast.Pipeline) *ast.Pipeline {
	if len(pipeline.Stages) < 2 {
		return pipeline
	}

	newStages := make([]ast.Stage, len(pipeline.Stages))
	copy(newStages, pipeline.Stages)

	for b := len(newStages) - 1; b >= 0; b-- {
		changed := false
		for a := b; a > 0; a-- {
			if !shouldFlip(newStages[a-1], newStages[a]) {
				break
			}

			newStages[a], newStages[a-1] = newStages[a-1], newStages[a]
			changed = true
		}

		if changed {
			// start over
			b = len(newStages)
		}
	}

	return ast.NewPipeline(newStages...)
}

func shouldFlip(a, b ast.Stage) bool {
	switch a.(type) {
	case *ast.ProjectStage:
		switch b.(type) {
		case *ast.LimitStage, *ast.SkipStage:
			return true
		}
	case *ast.SortStage:
		switch b.(type) {
		case *ast.MatchStage:
			return true
		}
	}

	return false
}
