package analyzer

import (
	"github.com/10gen/mongoast/ast"
)

// IsCardinalityAlteringStage returns true if this given pipeline stage changes the number
// of documents returned from a collection.
func IsCardinalityAlteringStage(stage ast.Stage) bool {
	switch stage.(type) {
	case *ast.AddFieldsStage,
		*ast.LookupStage,
		*ast.ProjectStage,
		*ast.RedactStage,
		*ast.ReplaceRootStage,
		*ast.SortStage,
		*ast.SortByCountStage:
		return false
	default:
		// We want default to be true because we would like to error on the side of being conservative.
		return true
	}
}
