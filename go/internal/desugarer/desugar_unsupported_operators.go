package desugarer

import (
	"github.com/10gen/mongoast/ast"
)

type functionDesugarer func(*ast.Function) ast.Expr

var functionDesugarers = map[string]functionDesugarer{
	"$sqlBetween": desugarSQLBetween,
	"$sqlSlice":   desugarSQLSlice,
	"$sqlDivide":  desugarSQLDivide,
	"$nullIf":     desugarNullIf,
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

func desugarSQLDivide(f *ast.Function) ast.Expr {
	args := f.Arg.(*ast.Document).FieldsMap()

	dividend := args["dividend"]
	divisor := args["divisor"]
	onError := args["onError"]

	return ast.NewConditional(
		ast.NewBinary(ast.Equals, divisor, zeroLiteral),
		onError,
		ast.NewBinary(ast.Divide, dividend, divisor),
	)
}

// desugarNullIf desugars { $nullIf: [<expr1>, <expr2>] } into existing MQL
// operators.
// Note that $nullIf returns null if <expr1> = <expr2> is true, and otherwise
// returns <expr1>. Like all MQL operators, $nullIf should never return missing
// so the desugared version wraps <expr1> in a $ifNull to promote missing to
// null if necessary.
func desugarNullIf(f *ast.Function) ast.Expr {
	args := f.Arg.(*ast.Array)

	expr1VarName := "expr1"
	expr1VarRef := ast.NewVariableRef(expr1VarName)

	return ast.NewLet(
		[]*ast.LetVariable{
			ast.NewLetVariable(expr1VarName,
				wrapInOp("$ifNull", args.Elements[0], nullLiteral),
			),
		},
		ast.NewConditional(
			ast.NewBinary(ast.Equals, expr1VarRef, args.Elements[1]),
			nullLiteral,
			expr1VarRef,
		),
	)
}
