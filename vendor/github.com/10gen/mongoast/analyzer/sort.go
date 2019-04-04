package analyzer

import "github.com/10gen/mongoast/ast"

// IsSortStage returns true if this is a type of sort stage.
func IsSortStage(stage ast.Stage) bool {
	switch stage.(type) {
	case *ast.SortStage,
		*ast.SortByCountStage:
		return true
	default:
		return false
	}
}
