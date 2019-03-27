package astutil

import (
	oldbson "github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/parser"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// RawDeparsePipeline converts an ast.Pipeline into a slice of oldbson.Raw.
// This function is used to convert an ast.Pipeline into a structure that can
// be sent to the go driver.
func RawDeparsePipeline(pipeline *ast.Pipeline) []oldbson.Raw {
	docs := make([]oldbson.Raw, len(pipeline.Stages))
	for i, stage := range pipeline.Stages {
		docs[i] = bsoncoreValueToRaw(parser.DeparseStage(stage))
	}

	return docs
}

// DeparsePipeline converts an ast.Pipeline into a slice of oldbson.D.
func DeparsePipeline(pipeline *ast.Pipeline) ([]oldbson.D, error) {
	docs := make([]oldbson.D, len(pipeline.Stages))
	for i, stage := range pipeline.Stages {
		bv := parser.DeparseStage(stage)
		d := bson.D{}
		err := bson.UnmarshalExtJSON([]byte(bv.String()), true, &d)
		if err != nil {
			return nil, err
		}
		docs[i] = newToOldBSOND(d)
	}

	return docs, nil
}

// ParsePipeline converts a slice of oldbson.D into an ast.Pipeline.
func ParsePipeline(docs []oldbson.D) (*ast.Pipeline, error) {
	stages := make([]ast.Stage, len(docs))
	for i, doc := range docs {
		d := oldToNewBSOND(doc)
		json, err := bson.MarshalExtJSON(&d, false, false)
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

func bsoncoreValueToRaw(v bsoncore.Value) oldbson.Raw {
	return oldbson.Raw{
		Kind: byte(v.Type),
		Data: v.Data,
	}
}

func oldToNewBSONA(old []interface{}) bson.A {
	newA := make(bson.A, len(old))

	for i, v := range old {
		switch t := v.(type) {
		case oldbson.D:
			v = oldToNewBSOND(t)
		case oldbson.M:
			v = oldToNewBSONM(t)
		case []interface{}:
			v = oldToNewBSONA(t)
		}

		newA[i] = v
	}

	return newA
}

func oldToNewBSOND(old oldbson.D) bson.D {
	newD := make(bson.D, len(old))

	for i, e := range old {
		newD[i] = oldToNewBSONE(e)
	}

	return newD
}

func oldToNewBSONE(old oldbson.DocElem) bson.E {
	v := old.Value
	switch t := old.Value.(type) {
	case oldbson.D:
		v = oldToNewBSOND(t)
	case oldbson.M:
		v = oldToNewBSONM(t)
	case []interface{}:
		v = oldToNewBSONA(t)
	}

	return primitive.E{Key: old.Name, Value: v}
}

func oldToNewBSONM(old oldbson.M) bson.M {
	newM := make(bson.M)

	for k, v := range old {
		switch t := v.(type) {
		case oldbson.D:
			v = oldToNewBSOND(t)
		case oldbson.M:
			v = oldToNewBSONM(t)
		case []interface{}:
			v = oldToNewBSONA(t)
		}

		newM[k] = v
	}

	return newM
}

func newToOldBSONA(newA bson.A) []interface{} {
	old := make([]interface{}, len(newA))

	for i, e := range newA {
		switch t := e.(type) {
		case bson.A:
			e = newToOldBSONA(t)
		case bson.D:
			e = newToOldBSOND(t)
		case bson.M:
			e = newToOldBSONM(t)
		}

		old[i] = e
	}

	return old
}

func newToOldBSOND(newD bson.D) oldbson.D {
	old := make(oldbson.D, len(newD))

	for i, e := range newD {
		old[i] = newToOldBSONE(e)
	}

	return old
}

func newToOldBSONE(newE bson.E) oldbson.DocElem {
	v := newE.Value
	switch t := newE.Value.(type) {
	case bson.A:
		v = newToOldBSONA(t)
	case bson.D:
		v = newToOldBSOND(t)
	case bson.M:
		v = newToOldBSONM(t)
	}

	return oldbson.NewDocElem(newE.Key, v)
}

func newToOldBSONM(newM bson.M) oldbson.M {
	old := make(oldbson.M)

	for k, v := range newM {
		switch t := v.(type) {
		case bson.A:
			v = newToOldBSONA(t)
		case bson.D:
			v = newToOldBSOND(t)
		case bson.M:
			v = newToOldBSONM(t)
		}

		old[k] = v
	}

	return old
}
