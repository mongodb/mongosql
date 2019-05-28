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
		docs[i] = NewToOldBSOND(d)
	}

	return docs, nil
}

// ParsePipeline converts a slice of oldbson.D into an ast.Pipeline.
func ParsePipeline(docs []oldbson.D) (*ast.Pipeline, error) {
	stages := make([]ast.Stage, len(docs))
	for i, doc := range docs {
		d := OldToNewBSOND(doc)
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

// OldToNewBSONA converts from the old Go driver's representation of a
// bson array, []interface{}, to the new Go driver's bson.A.
func OldToNewBSONA(old []interface{}) bson.A {
	newA := make(bson.A, len(old))

	for i, v := range old {
		switch t := v.(type) {
		case oldbson.D:
			v = OldToNewBSOND(t)
		case oldbson.M:
			v = OldToNewBSONM(t)
		case []interface{}:
			v = OldToNewBSONA(t)
		}

		newA[i] = v
	}

	return newA
}

// OldToNewBSOND converts from the old Go driver's bson.D to the new
// Go driver's bson.D.
func OldToNewBSOND(old oldbson.D) bson.D {
	newD := make(bson.D, len(old))

	for i, e := range old {
		newD[i] = OldToNewBSONE(e)
	}

	return newD
}

// OldToNewBSONE converts from the old Go driver's bson.DocElem to
// the new Go driver's bson.E.
func OldToNewBSONE(old oldbson.DocElem) bson.E {
	v := old.Value
	switch t := old.Value.(type) {
	case oldbson.D:
		v = OldToNewBSOND(t)
	case oldbson.M:
		v = OldToNewBSONM(t)
	case []interface{}:
		v = OldToNewBSONA(t)
	}

	return primitive.E{Key: old.Name, Value: v}
}

// OldToNewBSONM converts from the old Go driver's bson.M to the new
// Go driver's bson.M.
func OldToNewBSONM(old oldbson.M) bson.M {
	newM := make(bson.M)

	for k, v := range old {
		switch t := v.(type) {
		case oldbson.D:
			v = OldToNewBSOND(t)
		case oldbson.M:
			v = OldToNewBSONM(t)
		case []interface{}:
			v = OldToNewBSONA(t)
		}

		newM[k] = v
	}

	return newM
}

// NewToOldBSONA converts from the new Go driver's bson.A to a
// []interface{}, which is the old Go driver's representation of
// a bson array.
func NewToOldBSONA(newA bson.A) []interface{} {
	old := make([]interface{}, len(newA))

	for i, e := range newA {
		switch t := e.(type) {
		case bson.A:
			e = NewToOldBSONA(t)
		case bson.D:
			e = NewToOldBSOND(t)
		case bson.M:
			e = NewToOldBSONM(t)
		}

		old[i] = e
	}

	return old
}

// NewToOldBSOND converts from the new Go driver's bson.D to the old
// Go driver's bson.D.
func NewToOldBSOND(newD bson.D) oldbson.D {
	old := make(oldbson.D, len(newD))

	for i, e := range newD {
		old[i] = NewToOldBSONE(e)
	}

	return old
}

// NewToOldBSONE converts from the new Go driver's bson.E to the old
// Go driver's bson.DocElem.
func NewToOldBSONE(newE bson.E) oldbson.DocElem {
	v := newE.Value
	switch t := newE.Value.(type) {
	case bson.A:
		v = NewToOldBSONA(t)
	case bson.D:
		v = NewToOldBSOND(t)
	case bson.M:
		v = NewToOldBSONM(t)
	}

	return oldbson.NewDocElem(newE.Key, v)
}

// NewToOldBSONM converts from the new Go driver's bson.M to the old
// Go driver's bson.M.
func NewToOldBSONM(newM bson.M) oldbson.M {
	old := make(oldbson.M)

	for k, v := range newM {
		switch t := v.(type) {
		case bson.A:
			v = NewToOldBSONA(t)
		case bson.D:
			v = NewToOldBSOND(t)
		case bson.M:
			v = NewToOldBSONM(t)
		}

		old[k] = v
	}

	return old
}
