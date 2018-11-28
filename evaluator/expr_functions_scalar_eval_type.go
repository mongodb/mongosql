package evaluator

func convertEvalType(exprs []SQLExpr) EvalType {
	typ, ok := sqlTypeFromSQLExpr(exprs[1])
	if !ok {
		return EvalString
	}
	return typ
}

func greatestEvalType(exprs []SQLExpr) EvalType {
	return preferentialType(exprs...)
}

func leastEvalType(exprs []SQLExpr) EvalType {
	return preferentialType(exprs...)
}

func nopushdownEvalType(exprs []SQLExpr) EvalType {
	return exprs[0].EvalType()
}
