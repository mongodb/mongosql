package desugarer

import (
	"github.com/10gen/mongoast/ast"
)

type functionDesugarer func(*ast.Function) ast.Expr

var functionDesugarers = map[string]functionDesugarer{
	"$sqlBetween": desugarSQLBetween,
	"$sqlSlice":   desugarSQLSlice,
}

// desugarUnsupportedOperators desugars any new operators ($sqlBetween,
// $divide, $sqlSlice, $like, $assert, $nullIf, $coalesce) into appropriate,
// equivalent aggregation operators.
func desugarUnsupportedOperators(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {
	out, _ := ast.Visit(pipeline, func(v ast.Visitor, n ast.Node) ast.Node {
		switch tn := n.(type) {
		case *ast.Function:
			if desugarerFunc, ok := functionDesugarers[tn.Name]; ok {
				n = desugarerFunc(tn)
			}
		}
		return n.Walk(v)
	})

	return out.(*ast.Pipeline)
}

func desugarSQLBetween(f *ast.Function) ast.Expr {
	args := f.Arg.(*ast.Array)
	inputVarName := "desugared_sqlBetween_input"
	inputVarRef := ast.NewVariableRef(inputVarName)

	return ast.NewLet(
		[]*ast.LetVariable{
			ast.NewLetVariable(inputVarName, args.Elements[0]),
		},
		ast.NewFunction("$sqlAnd", ast.NewArray(
			ast.NewFunction("$sqlGte", ast.NewArray(
				inputVarRef,
				args.Elements[1],
			)),
			ast.NewFunction("$sqlLte", ast.NewArray(
				inputVarRef,
				args.Elements[2],
			)),
		)),
	)
}

func desugarSQLSlice(f *ast.Function) ast.Expr {
	args := f.Arg.(*ast.Array)

	var expr ast.Expr = ast.NewFunction("$slice", args)
	if len(args.Elements) == 3 {
		expr = ast.NewConditional(
			ast.NewBinary(ast.LessThanOrEquals, args.Elements[1], zeroLiteral),
			nullLiteral,
			expr,
		)
	}

	return expr
}
