package analyzer

import (
	"github.com/10gen/mongoast/ast"
)

// IsAccumulatorFunction returns true if this is an accumulator function.
func IsAccumulatorFunction(f *ast.Function) bool {
	switch f.Name {
	case "$addToSet",
		"$avg",
		"$first",
		"$last",
		"$max",
		"$mergeObjects",
		"$min",
		"$push",
		"$stdDevPop",
		"$stdDevSamp",
		"$sum":
		return true
	}
	return false
}
