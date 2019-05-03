package parsertest

import (
	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/jsonutil"
	"github.com/10gen/mongoast/parser"
)

// ParseExpr parses an ast.Expr for a string, and panics if there is an
// error while parsing.
func ParseExpr(input string) ast.Expr {
	e, err := ParseExprErr(input)
	if err != nil {
		panic(err)
	}
	return e
}

// ParseExprErr parses an ast.Expr from a string, but may also return an
// error if there is an error while parsing.
func ParseExprErr(input string) (ast.Expr, error) {
	v := jsonutil.ParseJSON(input)
	return parser.ParseExpr(v)
}

// ParseMatchExpr parses an ast.Expr from a string.
func ParseMatchExpr(input string) ast.Expr {
	v := jsonutil.ParseJSON(input)
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
	v := jsonutil.ParseJSON(input)
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
	v := jsonutil.ParseJSON(input)
	doc, ok := v.DocumentOK()
	if !ok {
		panic("stages must be documents")
	}
	return parser.ParseStage(doc)
}

// ParseStageForError parses an ast.Stage from a string returning the error produced.
func ParseStageForError(input string) error {
	v := jsonutil.ParseJSON(input)
	doc, ok := v.DocumentOK()
	if !ok {
		panic("stages must be documents")
	}
	_, err := parser.ParseStage(doc)
	return err
}
