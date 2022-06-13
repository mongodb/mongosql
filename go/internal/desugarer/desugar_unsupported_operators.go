package desugarer

import (
	"strings"

	"github.com/10gen/mongoast/ast"
)

type functionDesugarer func(*ast.Function) ast.Expr

var functionDesugarers = map[string]functionDesugarer{
	"$sqlBetween": desugarSQLBetween,
	"$sqlConvert": desugarSQLConvert,
	"$sqlCos":     desugarSQLCos,
	"$sqlDivide":  desugarSQLDivide,
	"$sqlLog":     desugarSQLLog,
	"$sqlMod":     desugarSQLMod,
	"$sqlRound":   desugarSQLRound,
	"$sqlSlice":   desugarSQLSlice,
	"$sqlSplit":   desugarSQLSplit,
	"$sqlSin":     desugarSQLSin,
	"$sqlSqrt":    desugarSQLSqrt,
	"$sqlTan":     desugarSQLTan,
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

func desugarSQLMod(f *ast.Function) ast.Expr {
	args := f.Arg.(*ast.Array)

	inputNumberVarName := "desugared_sqlMod_input0"
	inputNumberVarRef := ast.NewVariableRef(inputNumberVarName)

	inputDivisorVarName := "desugared_sqlMod_input1"
	inputDivisorVarRef := ast.NewVariableRef(inputDivisorVarName)

	return ast.NewLet(
		[]*ast.LetVariable{
			ast.NewLetVariable(inputNumberVarName, args.Elements[0]),
			ast.NewLetVariable(inputDivisorVarName, args.Elements[1]),
		},
		ast.NewConditional(
			ast.NewBinary(ast.Equals, inputDivisorVarRef, zeroLiteral),
			nullLiteral,
			ast.NewBinary(ast.Mod, inputNumberVarRef, inputDivisorVarRef),
		))
}

func desugarSQLLog(f *ast.Function) ast.Expr {
	args := f.Arg.(*ast.Array)

	firstArgNan := ast.NewBinary(ast.Equals, args.Elements[0], nanLiteral)
	secondArgNan := ast.NewBinary(ast.Equals, args.Elements[1], nanLiteral)

	firstArgNegative := ast.NewBinary(ast.LessThanOrEquals, args.Elements[0], zeroLiteral)
	secondArgOne := ast.NewBinary(ast.Equals, args.Elements[1], oneLiteral)
	secondArgNegative := ast.NewBinary(ast.LessThanOrEquals, args.Elements[1], zeroLiteral)

	nanCheckCondition := wrapInOp("$or", firstArgNan, secondArgNan)
	invalidArgCondition := wrapInOp("$or", firstArgNegative, secondArgOne, secondArgNegative)

	invalidArgConditional := ast.NewConditional(
		invalidArgCondition,
		nullLiteral,
		ast.NewBinary(ast.Log, args.Elements[0], args.Elements[1]),
	)

	return ast.NewConditional(nanCheckCondition, nanLiteral, invalidArgConditional)
}

func desugarSQLRound(f *ast.Function) ast.Expr {
	args := f.Arg.(*ast.Array)

	inputNumberVarName := "desugared_sqlRound_input0"
	inputNumberVarRef := ast.NewVariableRef(inputNumberVarName)

	inputPlaceVarName := "desugared_sqlRound_input1"
	inputPlaceVarRef := ast.NewVariableRef(inputPlaceVarName)

	argIsNan := ast.NewBinary(ast.Equals, inputPlaceVarRef, nanLiteral)

	withinRange := wrapInOp("$and",
		ast.NewBinary(ast.GreaterThanOrEquals,
			inputPlaceVarRef,
			negTwentyLiteral,
		),
		ast.NewBinary(ast.LessThanOrEquals,
			inputPlaceVarRef,
			oneHundredLiteral,
		),
	)

	argCheckConditional := ast.NewConditional(
		withinRange,
		wrapInOp("$round", inputNumberVarRef, inputPlaceVarRef),
		nullLiteral,
	)

	return ast.NewLet(
		[]*ast.LetVariable{
			ast.NewLetVariable(inputNumberVarName, args.Elements[0]),
			ast.NewLetVariable(inputPlaceVarName, args.Elements[1]),
		},
		ast.NewConditional(
			argIsNan,
			nanLiteral,
			argCheckConditional,
		))
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

func desugarSQLCos(f *ast.Function) ast.Expr {
	return desugarSQLTrig(f, "$cos")
}

func desugarSQLSin(f *ast.Function) ast.Expr {
	return desugarSQLTrig(f, "$sin")
}

func desugarSQLTan(f *ast.Function) ast.Expr {
	return desugarSQLTrig(f, "$tan")
}

func desugarSQLTrig(f *ast.Function, name string) ast.Expr {
	args := f.Arg.(*ast.Array)

	return ast.NewConditional(
		wrapInInfinityCheck(args.Elements[0]),
		nullLiteral,
		ast.NewFunction(name, args.Elements[0]),
	)
}

func wrapInInfinityCheck(arg ast.Expr) *ast.Function {
	posInfCheck := ast.NewBinary(ast.Equals, arg, posInfLiteral)
	negInfCheck := ast.NewBinary(ast.Equals, arg, negInfLiteral)
	return wrapInOp("$or", posInfCheck, negInfCheck)
}

func desugarSQLSqrt(f *ast.Function) ast.Expr {
	args := f.Arg.(*ast.Array)

	inputNumberVarName := "desugared_sqlSqrt_input"
	inputNumberVarRef := ast.NewVariableRef(inputNumberVarName)

	argIsNan := ast.NewBinary(ast.Equals, inputNumberVarRef, nanLiteral)
	argNegativeCheck := ast.NewBinary(ast.LessThan, inputNumberVarRef, zeroLiteral)
	argNegativeCheckConditional := ast.NewConditional(argNegativeCheck, nullLiteral, ast.NewFunction("$sqrt", inputNumberVarRef))

	return ast.NewLet(
		[]*ast.LetVariable{
			ast.NewLetVariable(inputNumberVarName, args.Elements[0]),
		},
		ast.NewConditional(
			argIsNan,
			nanLiteral,
			argNegativeCheckConditional,
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

// desugarSQLSplit desugars {"$sqlSplit": [<expr1>, <expr2>, <expr3>]} into
// existing MQL operators. Note that $sqlSplit returns null if <expr1>, <expr2>,
// or <expr3> is null or missing, or if <expr2> is an empty string.
func desugarSQLSplit(f *ast.Function) ast.Expr {
	args := f.Arg.(*ast.Array)

	inputStrVarName := "desugared_sqlSplit_input0"
	inputStrVarRef := ast.NewVariableRef(inputStrVarName)

	inputDelimVarName := "desugared_sqlSplit_input1"
	inputDelimVarRef := ast.NewVariableRef(inputDelimVarName)

	inputTokenNumVarName := "desugared_sqlSplit_input2"
	inputTokenNumVarRef := ast.NewVariableRef(inputTokenNumVarName)

	splitExprVarName := "desugared_sqlSplit_split_expr"
	splitExprVarRef := ast.NewVariableRef(splitExprVarName)

	sliceExprVarName := "desugared_sqlSplit_slice_expr"
	sliceExprVarRef := ast.NewVariableRef(sliceExprVarName)

	var splitExpr ast.Expr = wrapInOp("$split", inputStrVarRef, inputDelimVarRef)

	// If a token number's absolute value is greater than the length of the split
	// array, set the token number to that absolute value. This is done to
	// circumvent MongoDB's default behavior of setting the starting position of
	// $slice to the beginning of the array even when the absolute value of the
	// given position is larger than the array itself.
	var absTokenNumExpr ast.Expr = ast.NewUnary(ast.Abs, inputTokenNumVarRef)
	var sizeExpr ast.Expr = ast.NewFunction("$size", splitExprVarRef)
	var tokenNumCondExpr ast.Expr = ast.NewConditional(ast.NewBinary(ast.GreaterThan, absTokenNumExpr, sizeExpr), absTokenNumExpr, inputTokenNumVarRef)
	var sliceExpr ast.Expr = wrapInOp("$slice", splitExprVarRef, tokenNumCondExpr, oneLiteral)

	// If $slice returns an empty vector, populate it with the empty string.
	// This is done to circumvent MongoDB's default behavior of $arrayElemAt[
	// [], 0] returning MISSING instead of the empty string.
	var sliceCondExpr ast.Expr = ast.NewConditional(ast.NewBinary(ast.Equals, sliceExprVarRef, ast.NewArray()), ast.NewArray(stringLiteral("")), sliceExprVarRef)
	var arrayElemAtExpr ast.Expr = wrapInOp("$arrayElemAt", sliceCondExpr, zeroLiteral)

	return ast.NewLet(
		[]*ast.LetVariable{
			ast.NewLetVariable(inputStrVarName, args.Elements[0]),
			ast.NewLetVariable(inputDelimVarName, args.Elements[1]),
			ast.NewLetVariable(inputTokenNumVarName, args.Elements[2]),
		},
		ast.NewConditional(
			ast.NewBinary(ast.Equals, inputDelimVarRef, stringLiteral("")),
			nullLiteral,
			ast.NewLet(
				[]*ast.LetVariable{
					ast.NewLetVariable(splitExprVarName, splitExpr),
				},
				ast.NewConditional(
					wrapInNullCheck(splitExprVarRef),
					nullLiteral,
					ast.NewLet([]*ast.LetVariable{ast.NewLetVariable(sliceExprVarName, sliceExpr)}, arrayElemAtExpr),
				),
			),
		),
	)
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
