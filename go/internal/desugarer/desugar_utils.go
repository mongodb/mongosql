package desugarer

import (
	"math/rand"

	"github.com/10gen/mongoast/ast"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func bsonValueIsNull(value bsoncore.Value) *ast.Constant {
	if value.Type == bsontype.Null {
		return trueLiteral
	}
	return falseLiteral
}

// wrapInNullCheck returns true if v is null, false otherwise.
func wrapInNullCheck(v ast.Expr) ast.Expr {
	// Constants simply wrap bsoncore.Values; if those values
	// are known to be null, there is no need to wrap in a Binary operation.
	switch t := v.(type) {
	case *ast.Constant:
		return bsonValueIsNull(t.Value)
	}

	return ast.NewBinary("$lte", v, nullLiteral)
}

// wrapInNullCheckedCond returns a document that evaluates to truePart
// if any of the element in values is null, and falsePart otherwise.
func wrapInNullCheckedCond(truePart, falsePart ast.Expr, values ...ast.Expr) ast.Expr {
	var condition ast.Expr
	conds := make([]ast.Expr, len(values))
	for i, cond := range values {
		conds[i] = wrapInNullCheck(cond)
	}
	switch len(conds) {
	case 0:
		return falsePart
	case 1:
		condition = conds[0]
	default:
		condition = wrapInOp("$or", conds...)
	}

	return ast.NewConditional(condition, truePart, falsePart)
}

// wrapInOp returns a document which passes all arguments to the op.
func wrapInOp(op string, args ...ast.Expr) *ast.Function {
	return ast.NewFunction(op, ast.NewArray(args...))
}

// generateUniqueNameFunc returns a function that generates a unique string
func generateUniqueNameFunc(length int) (f func() string) {
	randSeed := int64(1)
	f = func() string {
		rand.Seed(randSeed)
		randSeed++
		alphaNumerics := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
		uniqueName := make([]rune, length)
		for i := range uniqueName {
			uniqueName[i] = alphaNumerics[rand.Intn(len(alphaNumerics))]
		}
		return string(uniqueName)
	}
	return
}
