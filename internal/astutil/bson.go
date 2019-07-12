package astutil

import (
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/parser"

	"go.mongodb.org/mongo-driver/bson"
)

// DeparsePipeline converts an ast.Pipeline into a slice of bson.D.
func DeparsePipeline(pipeline *ast.Pipeline) ([]bson.D, error) {
	docs := make([]bson.D, len(pipeline.Stages))
	for i, stage := range pipeline.Stages {
		bv := parser.DeparseStage(stage)
		d := bson.D{}
		err := bson.Unmarshal(bv.Data, &d)
		if err != nil {
			return nil, err
		}
		docs[i] = d
	}

	return docs, nil
}

// ParsePipeline converts a slice of bson.D into an ast.Pipeline.
func ParsePipeline(docs []bson.D) (*ast.Pipeline, error) {
	stages := make([]ast.Stage, len(docs))
	for i, doc := range docs {
		json, err := bson.MarshalExtJSON(&doc, false, false)
		if err != nil {
			return nil, err
		}

		stages[i], err = parser.ParseStageJSON(string(json))
		if err != nil {
			return nil, err
		}
	}

	return ast.NewPipeline(stages...), nil
}
