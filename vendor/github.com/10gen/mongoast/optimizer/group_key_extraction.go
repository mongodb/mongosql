package optimizer

import (
	"github.com/10gen/mongoast/analyzer"
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/parser"
)

// GroupKeyExtraction rewrites complex group keys that only contain a
// single unique field reference into simple field refs, and
// reconstruct the _id field in a subsequent AddFields stage. This
// allows more pipelines to take advantage of the faster DISTINCT_SCAN
// even before SERVER-40090 is fixed.
func GroupKeyExtraction(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {
	allStages := make([]ast.Stage, 0, len(pipeline.Stages))

	for _, stage := range pipeline.Stages {
		if group, ok := stage.(*ast.GroupStage); ok {
			allStages = append(allStages, extractGroupKey(group)...)
		} else {
			allStages = append(allStages, stage)
		}
	}

	pipeline.Stages = allStages
	return pipeline
}

// extractGroupKey attempts to extract a group key as required for the
// GroupKeyExtraction optimization. If the rewrite can be applied, a
// slice of two stages (a GroupStage and an AddFieldsStage) is
// returned. Otherwise, this function returns a slice containing only
// the input GroupStage.
func extractGroupKey(group *ast.GroupStage) []ast.Stage {
	originalGroup := group.DeepCopy().(*ast.GroupStage)

	switch group.By.(type) {
	case ast.FieldLikeRef, *ast.Constant:
		// If the group key is a reference or a constant, this rewrite
		// won't improve anything.
		return []ast.Stage{originalGroup}
	}

	// If the group key has complex expressions (anything other than
	// constants, documents, arrays, and refs), we cannot safely apply
	// this rewrite.
	if hasComplexExprs(group.By) {
		return []ast.Stage{originalGroup}
	}

	// get the set of unique field names referenced in the group key
	names, complete := analyzer.ReferencedFieldNames(group.By)
	if !complete {
		panic("referenced field list should always be complete for an expr")
	}

	// We can only apply this rewrite if there is exactly one unique
	// field reference in the group key.
	if len(names) != 1 {
		return []ast.Stage{originalGroup}
	}

	// parse the field reference name into the appropriate ast.Expr
	ref, err := parser.ParseFieldRef(names[0])
	if err != nil {
		panic("a stringified ast.FieldRef should always parse successfully")
	}

	// save a reference to the original group key, because we are
	// still going to use that expr after we change the group key
	oldGroupKey := group.By

	// change the GroupStage to group by the field ref instead of the
	// original group key
	group.By = ref

	// get all ref exprs in oldGroupKey for in-place modification
	refs, complete := analyzer.ReferencedFields(oldGroupKey, false)
	if !complete {
		panic("referenced field list should always be complete for an expr")
	}

	// modify the field refs in place to reference _id instead of the
	// field referenced in the original group key
	for _, ref := range refs {
		var fieldRef *ast.FieldRef

		var ok bool
		fieldRef, ok = ref.(*ast.FieldRef)
		if !ok {
			return []ast.Stage{originalGroup}
		}

		fieldRef.Name = "_id"
		fieldRef.Parent = nil
	}

	// project _id back to the original group key with the field
	// references fixed up
	addFields := ast.NewAddFieldsStage(
		ast.NewAddFieldsItem("_id", oldGroupKey),
	)

	return []ast.Stage{group, addFields}
}

func hasComplexExprs(expr ast.Expr) bool {
	v := &complexExprFinder{}
	expr.Walk(v)
	return v.hasComplexExpr
}

type complexExprFinder struct {
	hasComplexExpr bool
}

func (v *complexExprFinder) Visit(n ast.Node) ast.Node {
	switch n.(type) {
	case ast.FieldLikeRef, *ast.Constant, *ast.Document, *ast.Array:
		n.Walk(v)
	default:
		v.hasComplexExpr = true
	}
	return n
}
