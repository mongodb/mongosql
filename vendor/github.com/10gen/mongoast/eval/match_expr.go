package eval

import (
	"github.com/10gen/mongoast/ast"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// EvaluateMatchExpr applies the expression to the value.
func EvaluateMatchExpr(expr ast.Expr, value bsoncore.Value, memoryLimit uint64) (bool, error) {
	ee := exprEvaluator{
		isMatchLanguage: true,
		memoryLimit:     memoryLimit,
	}
	v, err := ee.evalToConstant(expr, value)
	if err != nil {
		return false, err
	}

	if b, ok := v.BooleanOK(); ok && b {
		return true, nil
	}

	return false, nil
}
