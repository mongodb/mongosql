package eval

import (
	"github.com/10gen/mongoast/ast"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func (v *exprEvaluator) evalFunction(n *ast.Function, value bsoncore.Value) (ast.Expr, error) {
	an, err := v.eval(n.Arg, value)
	if err != nil {
		return nil, err
	}
	if an != n.Arg {
		return ast.NewFunction(n.Name, an), nil
	}
	return n, nil
}
