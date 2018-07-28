package bsonutil

import "github.com/10gen/mongo-go-driver/bson"

// WrapInType wraps the passed expression in an expression
// that returns the type of the expression.
func WrapInType(v interface{}) interface{} {
	return bson.D{{Name: "$type", Value: v}}
}

// WrapInBinOp builds an expression that evaluates a two argument operator
// on the two passed argument expressions.
func WrapInBinOp(op string, v1 interface{}, v2 interface{}) interface{} {
	return bson.D{{Name: op, Value: []interface{}{v1, v2}}}
}

// WrapInCond builds a contial expression, the first expresssion
// is the condition, the second is the true expression, the third
// is the false expression.
func WrapInCond(c, b1, b2 interface{}) interface{} {
	return bson.D{
		{Name: "$cond",
			Value: []interface{}{c, b1, b2},
		},
	}
}
