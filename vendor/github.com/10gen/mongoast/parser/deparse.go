package parser

import (
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"

	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// DeparsePipeline turns the pipeline into a bson.Array.
func DeparsePipeline(pipeline *ast.Pipeline) bsoncore.Value {
	values := make([]bsoncore.Value, len(pipeline.Stages))
	for i, s := range pipeline.Stages {
		values[i] = DeparseStage(s)
	}

	return bsonutil.ArrayFromValues(values...)
}

// DeparseNode can deparse any node.
func DeparseNode(n ast.Node) bsoncore.Value {
	switch tn := n.(type) {
	case *ast.Pipeline:
		return DeparsePipeline(tn)
	case ast.Stage:
		return DeparseStage(tn)
	case ast.Expr:
		return DeparseExpr(tn)
	}
	panic("unreachable")
}
