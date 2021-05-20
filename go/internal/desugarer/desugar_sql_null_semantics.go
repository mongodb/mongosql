package desugarer

import (
	"fmt"
	"github.com/10gen/mongoast/ast"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

var nullSemanticFuncMapping = map[string]string{
	"$sqlEq":          "$eq",
	"$sqlIndexOfCP":   "$indexOfCP",
	"$sqlLt":          "$lt",
	"$sqlLte":         "$lte",
	"$sqlGt":          "$gt",
	"$sqlGte":         "$gte",
	"$sqlNe":          "$ne",
	"$sqlNot":         "$not",
	"$sqlSize":        "$size",
	"$sqlStrLenBytes": "$strLenBytes",
	"$sqlStrLenCP":    "$strLenCP",
	"$sqlSubstrCP":    "$substrCP",
	"$sqlToLower":     "$toLower",
	"$sqlToUpper":     "$toUpper",
}

// desugarSQLNullSemantics desugars any $sql-prefixed operators (other than
// $sqlSlice and $sqlBetween) into their corresponding aggregation operator
// wrapped in operations that null-check the arguments.
func desugarSQLNullSemantics(pipeline *ast.Pipeline, _ uint64) *ast.Pipeline {
	out, _ := ast.Visit(pipeline, func(v ast.Visitor, n ast.Node) ast.Node {
		switch tn := n.(type) {
		case *ast.Function:
			if mqlOp, ok := nullSemanticFuncMapping[tn.Name]; ok {
				n = desugarSqlOp(tn, mqlOp)
			}
		}
		return n.Walk(v)
	})

	return out.(*ast.Pipeline)
}

func desugarSqlOp(f *ast.Function, mqlOp string) ast.Expr {
	switch f.Arg.(type) {
	case *ast.Array:
		args := f.Arg.(*ast.Array)
		numChildren := len(args.Elements)

		assignments := make([]*ast.LetVariable, 0, numChildren)
		newArgs := make([]ast.Expr, 0, numChildren)
		nullCheckArgs := make([]ast.Expr, 0, numChildren)

		for i, arg := range args.Elements {
			constant, isConstant := arg.(*ast.Constant)
			if isConstant && constant.Value.Type != bsontype.Null {
				// The arg is a non-null literal, we will not create $let binding or check null value for it
				newArgs = append(newArgs, arg)
			} else {
				binding := fmt.Sprintf("desugared_%s_input%d", f.Name[1:], i)
				ref := ast.NewVariableRef(binding)

				assignments = append(assignments, ast.NewLetVariable(binding, arg))
				newArgs = append(newArgs, ref)
				nullCheckArgs = append(nullCheckArgs, ref)
			}
		}

		mqlFunc := wrapInOp(mqlOp, newArgs...)

		evaluation := wrapInNullCheckedCond(
			// if any operand is null, return null.
			nullLiteral,
			// else return MQL native result
			mqlFunc,
			nullCheckArgs...,
		)

		return ast.NewLet(assignments, evaluation)
	case ast.Expr:
		arg := f.Arg

		constant, isConstant := arg.(*ast.Constant)
		if isConstant && constant.Value.Type != bsontype.Null {
			// The only arg is a non-null literal, we just need to return the MQL Op expression
			return ast.NewFunction(mqlOp, arg)
		} else {
			// Use $let bind the arg and then wrap the MQL Op expression in null checked condition
			binding := fmt.Sprintf("desugared_%s_input", f.Name[1:])
			ref := ast.NewVariableRef(binding)
			assignments := []*ast.LetVariable{ast.NewLetVariable(binding, arg)}

			evaluation := wrapInNullCheckedCond(
				nullLiteral,
				ast.NewFunction(mqlOp, ref),
				ref,
			)
			return ast.NewLet(assignments, evaluation)
		}
	default:
		panic("Unsupported argument when desurgaring null semantics")
	}
}
