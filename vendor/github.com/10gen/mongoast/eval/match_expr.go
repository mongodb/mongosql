package eval

import (
	"github.com/10gen/mongoast/ast"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// EvaluateMatchExpr applies the expression to the value.
func EvaluateMatchExpr(expr ast.Expr, value bsoncore.Value) (bool, error) {
	v, err := exprEvaluator{isMatchExpr: true}.evalToConstant(expr, value)
	if err != nil {
		return false, err
	}

	if b, ok := v.BooleanOK(); ok && b {
		return true, nil
	}

	return false, nil
}
