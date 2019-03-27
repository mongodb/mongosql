package parsertest

import (
	"strings"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/parser"

	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// ParseExpr parses an ast.Expr from a string.
func ParseExpr(input string) ast.Expr {
	v := parseJSON(input)
	e, err := parser.ParseExpr(v)
	if err != nil {
		panic(err)
	}
	return e
}

// ParseMatchExpr parses an ast.Expr from a string.
func ParseMatchExpr(input string) ast.Expr {
	v := parseJSON(input)
	doc, ok := v.DocumentOK()
	if !ok {
		panic("match expressions must be documents")
	}
	e, err := parser.ParseMatchExpr(doc)
	if err != nil {
		panic(err)
	}
	return e
}

// ParsePipeline parses an *ast.Pipeline from a string.
func ParsePipeline(input string) *ast.Pipeline {
	v := parseJSON(input)
	arr, ok := v.ArrayOK()
	if !ok {
		panic("pipeline expressions must be arrays")
	}
	p, err := parser.ParsePipeline(arr)
	if err != nil {
		panic(err)
	}
	return p
}

// ParseStage parses an ast.Stage from a string, and panics if there is an
// error when parsing.
func ParseStage(input string) ast.Stage {
	s, err := ParseStageErr(input)
	if err != nil {
		panic(err)
	}
	return s
}

// ParseStageErr parses an ast.Stage from a string, but may also return an
// error if there is an error while parsing.
func ParseStageErr(input string) (ast.Stage, error) {
	v := parseJSON(input)
	doc, ok := v.DocumentOK()
	if !ok {
		panic("stages must be documents")
	}
	return parser.ParseStage(doc)
}

func parseJSON(input string) bsoncore.Value {
	vr, err := bsonrw.NewExtJSONValueReader(strings.NewReader(input), false)
	if err != nil {
		panic(err)
	}
	c := bsonrw.NewCopier()

	t, bytes, err := c.CopyValueToBytes(vr)

	if err != nil {
		panic(err)
	}

	return bsoncore.Value{
		Type: t,
		Data: bytes,
	}
}
