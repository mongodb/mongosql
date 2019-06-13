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
func DeparsePipeline(pipeline *ast.Pipeline) ([]bson.D, error) {
	docs := make([]bson.D, len(pipeline.Stages))
	for i, stage := range pipeline.Stages {
		bv := parser.DeparseStage(stage)
		d := bson.D{}
		err := bson.UnmarshalExtJSON([]byte(bv.String()), true, &d)
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
		newA[i] = OldToNewBSON(v)
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
	return primitive.E{Key: old.Name, Value: OldToNewBSON(old.Value)}
}

// OldToNewBSONM converts from the old Go driver's bson.M to the new
// Go driver's bson.M.
func OldToNewBSONM(old oldbson.M) bson.M {
	newM := make(bson.M)

	for k, v := range old {
		newM[k] = OldToNewBSON(v)
	}

	return newM
}

// OldToNewBSON recursively converts from old bson types to new bson types.
func OldToNewBSON(old interface{}) interface{} {
	v := old
	switch t := old.(type) {
	case oldbson.D:
		v = OldToNewBSOND(t)
	case oldbson.M:
		v = OldToNewBSONM(t)
	case []oldbson.D:
		n := make([]bson.D, len(t))
		for i, oldD := range t {
			n[i] = OldToNewBSOND(oldD)
		}
		v = n
	case []oldbson.M:
		n := make([]bson.M, len(t))
		for i, oldM := range t {
			n[i] = OldToNewBSONM(oldM)
		}
		v = n
	case []interface{}:
		v = OldToNewBSONA(t)
	case []uint8:
		v = primitive.Binary{Subtype: 0, Data: t}
	case oldbson.Binary:
		v = primitive.Binary{Subtype: t.Kind, Data: t.Data}
	case oldbson.ObjectId:
		v, _ = primitive.ObjectIDFromHex(t.Hex())
	case oldbson.MongoTimestamp:
		v = primitive.Timestamp{T: uint32(uint64(t) >> 32), I: uint32(t)}
	case nil:
		v = primitive.Null{}
	case oldbson.RegEx:
		v = primitive.Regex{Pattern: t.Pattern, Options: t.Options}
	case oldbson.DBPointer:
		oid, _ := primitive.ObjectIDFromHex(t.Id.Hex())
		v = primitive.DBPointer{DB: t.Namespace, Pointer: oid}
	case oldbson.JavaScript:
		code := primitive.JavaScript(t.Code)
		v = code
		if t.Scope != nil {
			v = primitive.CodeWithScope{Code: code, Scope: OldToNewBSON(t.Scope)}
		}
	case oldbson.Symbol:
		v = primitive.Symbol(t)
	case oldbson.Decimal128:
		v, _ = primitive.ParseDecimal128(t.String())
	}

	switch v {
	case oldbson.Undefined:
		v = primitive.Undefined{}
	case oldbson.MinKey:
		v = primitive.MinKey{}
	case oldbson.MaxKey:
		v = primitive.MaxKey{}
	}

	return v
}

// NewToOldBSONA converts from the new Go driver's bson.A to a
// []interface{}, which is the old Go driver's representation of
// a bson array.
func NewToOldBSONA(newA bson.A) []interface{} {
	old := make([]interface{}, len(newA))

	for i, e := range newA {
		old[i] = NewToOldBSON(e)
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
	return oldbson.NewDocElem(newE.Key, NewToOldBSON(newE.Value))
}

// NewToOldBSONM converts from the new Go driver's bson.M to the old
// Go driver's bson.M.
func NewToOldBSONM(newM bson.M) oldbson.M {
	old := make(oldbson.M)

	for k, v := range newM {
		old[k] = NewToOldBSON(v)
	}

	return old
}

// NewToOldBSON recursively converts from new bson types to old bson types.
func NewToOldBSON(new interface{}) interface{} {
	v := new
	switch t := new.(type) {
	case bson.D:
		v = NewToOldBSOND(t)
	case bson.M:
		v = NewToOldBSONM(t)
	case bson.A:
		v = NewToOldBSONA(t)
	case []bson.D:
		old := make([]oldbson.D, len(t))
		for i, newD := range t {
			old[i] = NewToOldBSOND(newD)
		}
		v = old
	case []bson.M:
		old := make([]oldbson.M, len(t))
		for i, newM := range t {
			old[i] = NewToOldBSONM(newM)
		}
		v = old
	case primitive.Binary:
		if t.Subtype == 0 {
			// nolint: unconvert
			v = []uint8(t.Data)
		} else {
			v = oldbson.Binary{Kind: t.Subtype, Data: t.Data}
		}
	case primitive.ObjectID:
		v = oldbson.ObjectIdHex(t.Hex())
	case primitive.DateTime:
		v = oldbson.MongoTimestamp(t)
	case primitive.Null:
		v = nil
	case primitive.Regex:
		v = oldbson.RegEx{Pattern: t.Pattern, Options: t.Options}
	case primitive.DBPointer:
		oid := oldbson.ObjectIdHex(t.Pointer.Hex())
		v = oldbson.DBPointer{Namespace: t.DB, Id: oid}
	case primitive.JavaScript:
		v = oldbson.JavaScript{Code: string(t)}
	case primitive.CodeWithScope:
		v = oldbson.JavaScript{Code: string(t.Code), Scope: NewToOldBSON(t.Scope)}
	case primitive.Symbol:
		v = oldbson.Symbol(t)
	case primitive.Decimal128:
		v, _ = oldbson.ParseDecimal128(t.String())
	case primitive.Undefined:
		v = oldbson.Undefined
	case primitive.MinKey:
		v = oldbson.MinKey
	case primitive.MaxKey:
		v = oldbson.MaxKey
	}

	return v
}
