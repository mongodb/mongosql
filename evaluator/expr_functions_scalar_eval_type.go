package evaluator

import "github.com/10gen/sqlproxy/evaluator/types"

func convertEvalType(exprs []SQLExpr) types.EvalType {
	typ, ok := evalTypeFromSQLTypeExpr(exprs[1])
	if !ok {
		return types.EvalString
	}
	return typ
}

func exprsToEvalTypers(exprs []SQLExpr) []types.EvalTyper {
	ret := make([]types.EvalTyper, len(exprs))
	for i := range exprs {
		ret[i] = exprs[i]
	}
	return ret
}

func greatestEvalType(exprs []SQLExpr) types.EvalType {
	return preferentialType(exprsToEvalTypers(exprs)...)
}

func leastEvalType(exprs []SQLExpr) types.EvalType {
	return preferentialType(exprsToEvalTypers(exprs)...)
}

func nopushdownEvalType(exprs []SQLExpr) types.EvalType {
	return exprs[0].EvalType()
}
