package parser

import (
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"

	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// DeparsePipeline turns the pipeline into a bson.Array.
func DeparsePipeline(pipeline *ast.Pipeline) bsoncore.Value {
	v, err := DeparsePipelineErr(pipeline)
	if err != nil {
		panic(err)
	}
	return v
}

func DeparsePipelineErr(pipeline *ast.Pipeline) (bsoncore.Value, error) {
	values := make([]bsoncore.Value, len(pipeline.Stages))
	for i, s := range pipeline.Stages {
		v, err := DeparseStageErr(s)
		if err != nil {
			return bsoncore.Value{}, err
		}
		values[i] = v
	}

	return bsonutil.ArrayFromValues(values...), nil
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
