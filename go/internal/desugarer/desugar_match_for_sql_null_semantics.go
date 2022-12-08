package desugarer

import (
	"sort"

	"github.com/10gen/mongoast/ast"
)

// desugarMatch inspects the pipeline for $match stages that
// contain an $expr with a $sql-X operator. It then generates a sql constraint for
// the arguments such that the values to be considered exist (not missing) and
// are not null.
func desugarMatch(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {
	newStages := make([]ast.Stage, 0, len(pipeline.Stages)*2)

	for _, stage := range pipeline.Stages {
		if stage, isMatch := stage.(*ast.MatchStage); isMatch {
			schemaStage, generatedSchema := desugarStage(stage)
			if generatedSchema {
				optimizedMatch := desugarSQLOpsInMatch(stage)
				newStages = append(newStages, schemaStage, optimizedMatch)
				continue
			}
		}
		newStages = append(newStages, stage)
	}

	return ast.NewPipeline(newStages...)
}

func desugarStage(stage *ast.MatchStage) (*ast.MatchStage, bool) {
	var generatedSchema bool
	out, _ := ast.Visit(stage, func(v ast.Visitor, n ast.Node) ast.Node {
		if expr, isExpr := n.(*ast.AggExpr); isExpr {
			n, generatedSchema = generateSchema(expr)
		}
		return n.Walk(v)
	})
	return out.(*ast.MatchStage), generatedSchema
}

// generateSchema collects all schema information for child nodes
// and returns them
func generateSchema(aggExpr *ast.AggExpr) (ast.Node, bool) {
	schemaVars := make(map[string]ast.FieldLikeRef)
	out, _ := ast.Visit(aggExpr, func(v ast.Visitor, n ast.Node) ast.Node {
		switch tn := n.(type) {
		// do not attempt optimizations in reduce
		case *ast.Reduce:
			return n
		case *ast.Function:
			if _, ok := nullSemanticFuncMapping[tn.Name]; ok {
				gatherVarsForNullSemanticConstriants(tn, schemaVars)
			}
		}
		return n.Walk(v)
	})
	schema := make([]ast.Expr, 0, len(schemaVars)*2)
	// we have to sort the keys so that they are in lexicographic order
	// in order for our test to pass. This shouldn't be needed in real-world
	// use-cases.
	refNames := make([]string, 0, len(schema))
	for k := range schemaVars {
		refNames = append(refNames, k)
	}
	sort.Strings(refNames)
	for _, name := range refNames {
		ref := schemaVars[name]
		existsDoc := ast.NewExists(ref, true)
		typeDoc := ast.NewFunction(ast.GetDottedFieldName(ref), ast.NewFunction("$not", ast.NewFunction("$type", nullStringLiteral)))
		schema = append(schema, existsDoc, typeDoc)
	}

	// if the length of the schema is zero, no schema was created
	if len(schema) > 0 {
		return wrapInOp("$and", schema...), true
	} else {
		return out, false
	}
}

func gatherVarsForNullSemanticConstriants(f *ast.Function, s map[string]ast.FieldLikeRef) {
	switch f.Arg.(type) {
	case *ast.Array:
		args := f.Arg.(*ast.Array)
		for _, arg := range args.Elements {
			handleRefs(arg, s)
		}

	case *ast.FieldRef:
		arg := f.Arg
		handleRefs(arg, s)

	default:
		panic("Unsupported argument when desugaring null semantics")
	}
}

func handleRefs(arg ast.Expr, s map[string]ast.FieldLikeRef) {
	ref, isRef := arg.(*ast.FieldRef)

	if isRef {
		// only check pure fieldrefs for existence
		if ast.IsPureFieldRef(ref) {
			s[ast.GetDottedFieldName(ref)] = ref
		}
	}
}

func desugarSQLOpsInMatch(pipeline *ast.MatchStage) *ast.MatchStage {
	out, _ := ast.Visit(pipeline, func(v ast.Visitor, n ast.Node) ast.Node {
		// if it's an AggExpr ($expr), we can simplify the $sqlOp since
		// we have schema constraints in place already
		if expr, isAggExpr := n.(*ast.AggExpr); isAggExpr {
			n = desugarExpr(expr)
		}
		return n.Walk(v)
	})

	return out.(*ast.MatchStage)
}

func desugarExpr(aggExpr *ast.AggExpr) ast.Node {
	out, _ := ast.Visit(aggExpr, func(v ast.Visitor, n ast.Node) ast.Node {
		if fn, isFN := n.(*ast.Function); isFN {
			if mqlOp, ok := nullSemanticFuncMapping[fn.Name]; ok {
				n = ast.NewFunction(mqlOp, fn.Arg)
			}
		}
		return n.Walk(v)
	})
	return out
}
