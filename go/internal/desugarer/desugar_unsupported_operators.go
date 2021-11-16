package desugarer

import (
	"strings"

	"github.com/10gen/mongoast/ast"
)

type functionDesugarer func(*ast.Function) ast.Expr

var functionDesugarers = map[string]functionDesugarer{
	"$sqlBetween": desugarSQLBetween,
	"$sqlConvert": desugarSQLConvert,
	"$sqlDivide":  desugarSQLDivide,
	"$sqlSlice":   desugarSQLSlice,
	"$coalesce":   desugarCoalesce,
	"$like":       desugarLike,
	"$nullIf":     desugarNullIf,
}

const regexCharsToEscape = `.^$*+?()[{\|`

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

func desugarSQLConvert(f *ast.Function) ast.Expr {
	args := f.Arg.(*ast.Document).FieldsMap()

	input := args["input"]
	targetType := args["to"]

	var onNull ast.Expr = nullLiteral
	if onNullArg, hasOnNull := args["onNull"]; hasOnNull {
		onNull = onNullArg
	}

	var onError ast.Expr = nullLiteral
	if onErrorArg, hasOnError := args["onError"]; hasOnError {
		onError = onErrorArg
	}

	inputVarName := "sqlConvert_input"
	inputVarRef := ast.NewVariableRef(inputVarName)

	return ast.NewLet(
		[]*ast.LetVariable{
			ast.NewLetVariable(inputVarName, input),
		},
		ast.NewFunction("$switch", ast.NewDocument(
			ast.NewDocumentElement("branches", ast.NewArray(
				ast.NewDocument(
					ast.NewDocumentElement("case",
						ast.NewBinary(ast.Equals, ast.NewFunction("$type", inputVarRef), targetType)),
					ast.NewDocumentElement("then", inputVarRef),
				),
				ast.NewDocument(
					ast.NewDocumentElement("case",
						ast.NewBinary(ast.LessThanOrEquals, inputVarRef, nullLiteral)),
					ast.NewDocumentElement("then", onNull),
				),
			)),
			ast.NewDocumentElement("default", onError),
		)),
	)
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

func desugarSQLSlice(f *ast.Function) ast.Expr {
	args := f.Arg.(*ast.Array)

	var expr ast.Expr = ast.NewFunction("$slice", args)
	if len(args.Elements) == 3 {
		expr = ast.NewConditional(
			ast.NewBinary(ast.LessThanOrEquals, args.Elements[2], zeroLiteral),
			nullLiteral,
			expr,
		)
	}

	return expr
}

func desugarCoalesce(f *ast.Function) ast.Expr {
	args := f.Arg.(*ast.Array)

	branches := make([]ast.Expr, len(args.Elements))

	for i, arg := range args.Elements {
		branches[i] = ast.NewDocument(
			ast.NewDocumentElement("case",
				ast.NewBinary(ast.GreaterThan,
					arg,
					nullLiteral,
				),
			),
			ast.NewDocumentElement("then", arg),
		)
	}

	return ast.NewFunction("$switch", ast.NewDocument(
		ast.NewDocumentElement("branches", ast.NewArray(branches...)),
		ast.NewDocumentElement("default", nullLiteral),
	))
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

func desugarLike(f *ast.Function) ast.Expr {
	args := f.Arg.(*ast.Document).FieldsMap()

	input := args["input"]
	pattern := args["pattern"]
	escape, hasEscape := args["escape"]

	patternLiteral, isLiteral := pattern.(*ast.Constant)
	if !isLiteral {
		panic("$like pattern must be literal")
	}

	var escapeChar rune
	if hasEscape {
		escapeChar = []rune(escape.(*ast.Constant).Value.StringValue())[0]
	}

	regex := convertSQLPattern(patternLiteral.Value.StringValue(), escapeChar)

	return ast.NewFunction("$regexMatch", ast.NewDocument(
		ast.NewDocumentElement("input", input),
		ast.NewDocumentElement("regex", stringLiteral(regex)),
		ast.NewDocumentElement("options", stringLiteral("is")),
	))
}

// convertSQLPattern converts a SQL-style LIKE pattern string into an MQL-style
// regular expression string.
func convertSQLPattern(pattern string, escapeChar rune) string {
	regex := "^"
	escaped := false
	for _, c := range pattern {
		if !escaped && c == escapeChar {
			escaped = true
			continue
		}

		switch {
		case c == '_':
			if escaped {
				regex += "_"
			} else {
				regex += "."
			}
		case c == '%':
			if escaped {
				regex += "%"
			} else {
				regex += ".*"
			}
		case strings.Contains(regexCharsToEscape, string(c)):
			regex += `\` + string(c)
		default:
			regex += string(c)
		}

		escaped = false
	}

	regex += "$"

	return regex
}
