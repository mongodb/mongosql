package analyzer

import (
	"github.com/10gen/mongoast/ast"
)

// IsFieldKiller returns true if this given pipeline stage kills any values
// not named in its output.
func IsFieldKiller(stage ast.Stage) bool {
	switch typedStage := stage.(type) {
	case *ast.BucketStage,
		*ast.BucketAutoStage,
		*ast.CollStatsStage,
		*ast.CountStage,
		*ast.FacetStage,
		*ast.GroupStage,
		*ast.ReplaceRootStage:
		return true
	case *ast.ProjectStage:
		return len(typedStage.Items) != len(typedStage.ExcludeItems())
	default:
		return false
	}
}
