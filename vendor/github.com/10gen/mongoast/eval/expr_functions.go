package eval

import (
	"github.com/10gen/mongoast/ast"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func (v *exprEvaluator) evalFunction(n *ast.Function, value bsoncore.Value) (ast.Expr, uint64, error) {
	an, amem, err := v.eval(n.Arg, value)
	if err != nil {
		return nil, 0, err
	}
	if an != n.Arg {
		return ast.NewFunction(n.Name, an), uint64(len(n.Name)) + amem, nil
	}
	return n, uint64(len(n.Name)) + amem, nil
}
