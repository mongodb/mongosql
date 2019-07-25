package parser

import (
	"strings"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// ParsePipeline parses the pipeline into an ast.Pipeline.
func ParsePipeline(arr bsoncore.Document) (*ast.Pipeline, error) {
	values, _ := arr.Values()
	stages := make([]ast.Stage, len(values))
	for i, v := range values {
		doc, ok := v.DocumentOK()
		if !ok {
			return nil, errors.New("stages can only be documents")
		}

		stage, err := ParseStage(doc)
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing stage")
		}

		stages[i] = stage
	}

	return ast.NewPipeline(stages...), nil
}

// ParsePipelineJSON parses an *ast.Pipeline from a string.
func ParsePipelineJSON(input string) (*ast.Pipeline, error) {
	v, err := parseJSON(input)
	if err != nil {
		return nil, err
	}
	var arr bsoncore.Document
	arr, ok := v.ArrayOK()
	if !ok {
		_, ok := v.DocumentOK()
		if !ok {
			return nil, errors.New("Each element of the 'pipeline' array must be an object")
		}
		arr = bsonutil.ArrayFromValues(v).Array()
	}

	p, err := ParsePipeline(arr)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func parseJSON(input string) (bsoncore.Value, error) {
	vr, err := bsonrw.NewExtJSONValueReader(strings.NewReader(input), false)
	if err != nil {
		return bsoncore.Value{}, err
	}

	c := bsonrw.NewCopier()
	t, bytes, err := c.CopyValueToBytes(vr)
	if err != nil {
		return bsoncore.Value{}, err
	}

	return bsoncore.Value{
		Type: t,
		Data: bytes,
	}, nil
}
