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
	modifiedStages := make([]ast.Stage, 0, len(pipeline.Stages))
	for _, stage := range pipeline.Stages {

		out, _ := ast.Visit(stage, func(v ast.Visitor, n ast.Node) ast.Node {
			switch tn := n.(type) {
			case *ast.Pipeline:
				// do not walk sub-pipelines
				return n
			case *ast.Function:
				if tn.Name == "$sqlAnd" {
					n = desugarSqlAnd(tn)
				} else if tn.Name == "$sqlOr" {
					n = desugarSqlOr(tn)
				} else if mqlOp, ok := nullSemanticFuncMapping[tn.Name]; ok {
					n = desugarSqlOp(tn, mqlOp)
				}
			}
			return n.Walk(v)
		})

		modifiedStages = append(modifiedStages, out.(ast.Stage))
	}

	return ast.NewPipeline(modifiedStages...)
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
		panic("Unsupported argument when desugaring null semantics")
	}
}

// desugarSQLAnd desugars a $sqlAnd expression to a nested $cond expression. In the first level of the $cond,
// it checks if any of the input variable is False, if so, return False literal directly. Otherwise in the second level
// it checks if any of the input variable is null, if so, return Null directly, otherwise return True literal.
// {$'sqlAnd': [<expr1>, <expr2>]}
//
// will be desugared into
//
//	"$let": {
//	  "vars": {
//	    "desugared_sqlAnd_input0": "$<expr1>",
//	    "desugared_sqlAnd_input1": "$<expr2>"
//	  },
//	  "in": {
//	    "$cond": {
//	      "if": {
//	        "$or": [
//	          // check if any of the vars is False
//	        ]
//	      },
//	      "then": {
//	        "$literal": false
//	      },
//	      "else": {
//	        "$cond": {
//	          "if": {
//	              // check if any of the vars is Null
//	          },
//	          "then": {
//	            "$literal": null
//	          },
//	          "else": {
//	            "$literal": true
//	          }
//	        }
//	      }
//	    }
//	  }
//	}
func desugarSqlAnd(f *ast.Function) ast.Expr {
	args, ok := f.Arg.(*ast.Array)
	if !ok {
		panic("Unsupported argument when desugaring $sqlAnd")
	}
	numChildren := len(args.Elements)

	assignments := make([]*ast.LetVariable, 0, numChildren)
	newArgs := make([]ast.Expr, 0, numChildren)
	nullCheckArgs := make([]ast.Expr, 0, numChildren)
	falseCheckArgs := make([]ast.Expr, 0, numChildren)

	for i, arg := range args.Elements {
		constant, isConstant := arg.(*ast.Constant)
		if isConstant && constant.Value.Type != bsontype.Null {
			// The arg is a non-null literal, we will not create $let binding or check null value for it
			newArgs = append(newArgs, arg)
			falseCheckArgs = append(falseCheckArgs, arg)
		} else {
			binding := fmt.Sprintf("desugared_sqlAnd_input%d", i)
			ref := ast.NewVariableRef(binding)

			assignments = append(assignments, ast.NewLetVariable(binding, arg))
			newArgs = append(newArgs, ref)
			falseCheckArgs = append(falseCheckArgs, ref)
			nullCheckArgs = append(nullCheckArgs, ref)
		}
	}

	nullCheckedCond := wrapInNullCheckedCond(
		// if any operand is null, return null.
		nullLiteral,
		// else return true
		trueLiteral,
		nullCheckArgs...,
	)

	// First check if any of the input param is evaluated to False, if so, evaluate the whole expression to False.
	var condition ast.Expr
	conds := make([]ast.Expr, len(falseCheckArgs))
	for i, cond := range falseCheckArgs {
		conds[i] = ast.NewBinary("$eq", cond, falseLiteral)
	}
	condition = wrapInOp("$or", conds...)

	evaluation := ast.NewConditional(condition, falseLiteral, nullCheckedCond)

	return ast.NewLet(assignments, evaluation)
}

// desugarSqlOr desugars a $sqlOr expression to a nested $cond expression. In the first level of the $cond,
// it checks if any of the input variable is True, if so, return True literal directly. Otherwise in the second level
// it checks if any of the input variable is null, if so, return Null directly, otherwise return False literal.
// {$'sqlOr': [<expr1>, <expr2>]}
//
// will be desugared into
//
//	"$let": {
//	  "vars": {
//	    "desugared_sqlOr_input0": "$<expr1>",
//	    "desugared_sqlOr_input1": "$<expr2>"
//	  },
//	  "in": {
//	    "$cond": {
//	      "if": {
//	          // check if any of the vars is True
//	      },
//	      "then": {
//	        "$literal": true
//	      },
//	      "else": {
//	        "$cond": {
//	          "if": {
//	            // check if any of the vars is Null
//	          },
//	          "then": {
//	            "$literal": null
//	          },
//	          "else": {
//	            "$literal": false
//	          },
//	        }
//	      }
//	    }
//	  }
//	}
func desugarSqlOr(f *ast.Function) ast.Expr {
	args, ok := f.Arg.(*ast.Array)
	if !ok {
		panic("Unsupported argument when desugaring $sqlOr")
	}
	numChildren := len(args.Elements)

	assignments := make([]*ast.LetVariable, 0, numChildren)
	newArgs := make([]ast.Expr, 0, numChildren)
	nullCheckArgs := make([]ast.Expr, 0, numChildren)
	trueCheckArgs := make([]ast.Expr, 0, numChildren)

	for i, arg := range args.Elements {
		constant, isConstant := arg.(*ast.Constant)
		if isConstant && constant.Value.Type != bsontype.Null {
			// The arg is a non-null literal, we will not create $let binding or check null value for it
			newArgs = append(newArgs, arg)
			trueCheckArgs = append(trueCheckArgs, arg)
		} else {
			binding := fmt.Sprintf("desugared_sqlOr_input%d", i)
			ref := ast.NewVariableRef(binding)

			assignments = append(assignments, ast.NewLetVariable(binding, arg))
			newArgs = append(newArgs, ref)
			trueCheckArgs = append(trueCheckArgs, ref)
			nullCheckArgs = append(nullCheckArgs, ref)
		}
	}

	nullCheckedCond := wrapInNullCheckedCond(
		// if any operand is null, return null.
		nullLiteral,
		// else return false
		falseLiteral,
		nullCheckArgs...,
	)

	// First check if any of the input param is evaluated to True, if so, evaluate the whole expression to True.
	var condition ast.Expr
	conds := make([]ast.Expr, len(trueCheckArgs))
	for i, cond := range trueCheckArgs {
		conds[i] = ast.NewBinary("$eq", cond, trueLiteral)
	}
	condition = wrapInOp("$or", conds...)

	evaluation := ast.NewConditional(condition, trueLiteral, nullCheckedCond)

	return ast.NewLet(assignments, evaluation)
}
