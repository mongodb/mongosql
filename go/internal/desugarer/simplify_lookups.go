package desugarer

import "github.com/10gen/mongoast/ast"

// getSimplifyLookups accepts a string representing the current database name and returns a pass
// that will remove the FromDB name from any $lookup when the FromDB name matches the current
// database name because that information is redudant and some implementations of mongodb aggregation
// language, e.g., mongod itself, do not support the FromDB field.
func getSimplifyLookups(currentDB string) func(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {
	return func(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {
		// modify the pipeline in place.
		for i := range pipeline.Stages {
			switch s := pipeline.Stages[i].(type) {
			case *ast.LookupStage:
				if s.FromDB == currentDB {
					s.FromDB = ""
				}
			}
		}

		return pipeline
	}
}
