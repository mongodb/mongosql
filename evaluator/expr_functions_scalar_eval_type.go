package evaluator

import (
	"github.com/10gen/sqlproxy/evaluator/types"
)

func absEvalType(exprs []SQLExpr) types.EvalType {
	return exprs[0].EvalType()
}

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

func modEvalType(exprs []SQLExpr) types.EvalType {
	return preferentialType(exprsToEvalTypers(exprs)...)
}

func powEvalType(exprs []SQLExpr) types.EvalType {
	return preferentialType(exprsToEvalTypers(exprs)...)
}

func nopushdownEvalType(exprs []SQLExpr) types.EvalType {
	return exprs[0].EvalType()
}

// The str_to_date function returns either a date or a datetime
// depending on the format string, so this function looks at
// the format string to determine what type the str_to_date
// function should return. If the format string is a literal string
// and does not contain any time format operators, it returns a
// date. Otherwise, it returns a datetime.
func strToDateEvalType(exprs []SQLExpr) types.EvalType {
	formatValueExpr, ok := exprs[1].(SQLValueExpr)
	if !ok {
		return types.EvalDatetime
	}
	formatValueStr := formatValueExpr.Value.String()

	foundPercent := false
	for _, r := range formatValueStr {
		if foundPercent {
			switch r {
			case 'H', 'i', 'S', 's', 'T':
				return types.EvalDatetime
			case '%':
			default:
				foundPercent = false
			}

		} else if r == '%' {
			foundPercent = true
		}
	}

	return types.EvalDate
}

func roundWithDecimalPlacesEvalType(exprs []SQLExpr) types.EvalType {
	return exprs[0].EvalType()
}

func truncateEvalType(exprs []SQLExpr) types.EvalType {
	return exprs[0].EvalType()
}
