package evaluator

import (
	"strconv"
)

//
// SQLExpr is the base type for a SQL expression.
//
type SQLExpr interface {
	Evaluate(*EvalCtx) (SQLValue, error)
}

//
// SQLValue is a comparable SQLExpr.
//
type SQLValue interface {
	SQLExpr
	CompareTo(SQLValue) (int, error)
}

//
// SQLNumeric is a numeric SQLValue.
//
type SQLNumeric interface {
	SQLValue
	Add(o SQLNumeric) SQLNumeric
	Sub(o SQLNumeric) SQLNumeric
	Product(o SQLNumeric) SQLNumeric
	Float64() float64
}

//
// SQLTemporal is a time-based SQLValue.
//
type SQLTemporal interface {
	SQLValue
}

// A base type for a binary node.
type sqlBinaryNode struct {
	left, right SQLExpr
}

//
// EvalCtx holds a slice of rows used to evaluate a SQLValue.
//
type EvalCtx struct {
	Rows    []Row
	ExecCtx *ExecutionCtx
}

// Matches checks if a given SQLExpr is "truthy" by coercing it to a boolean value.
// - booleans: the result is simply that same return value
// - numeric values: the result is true if and only if the value is non-zero.
// - strings, the result is true if and only if that string can be parsed as a number,
//   and that number is non-zero.
func Matches(expr SQLExpr, ctx *EvalCtx) (bool, error) {

	sv, err := expr.Evaluate(ctx)
	if err != nil {
		return false, err
	}

	if asBool, ok := sv.(SQLBool); ok {
		return bool(asBool), nil
	}
	if asNum, ok := sv.(SQLNumeric); ok {
		return asNum.Float64() != float64(0), nil
	}
	if asStr, ok := sv.(SQLString); ok {
		// check if the string should be considered "truthy" by trying to convert it to a number and comparing to 0.
		// more info: http://stackoverflow.com/questions/12221211/how-does-string-truthiness-work-in-mysql
		if parsedFloat, err := strconv.ParseFloat(string(asStr), 64); err == nil {
			return parsedFloat != float64(0), nil
		}
		return false, nil
	}

	// TODO - handle other types with possible values that are "truthy" : dates, etc?
	return false, nil
}
