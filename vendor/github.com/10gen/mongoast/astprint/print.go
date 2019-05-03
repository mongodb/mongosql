package astprint

import (
	"bytes"
	"fmt"
	"io"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/parser"
)

// Print prints the node to the writer.
func Print(w io.Writer, n ast.Node) {
	v := parser.DeparseNode(n)

	fmt.Fprint(w, v.String())
}

// String prints the node to a string.
func String(n ast.Node) string {
	var buf bytes.Buffer
	Print(&buf, n)
	return buf.String()
}

// ShellPrint prints the node as a string that is Mongo Shell pasteable.
func ShellPrint(w io.Writer, n ast.Node) {
	switch tn := n.(type) {
	case *ast.Pipeline:
		ShellPrintPipeline(w, tn, true)
	case ast.Stage:
		ShellPrintStage(w, tn)
	case ast.Expr:
		ShellPrintExpr(w, tn)
	}
}

// ShellString prints the node as a string that is Mongo Shell pasteable.
func ShellString(n ast.Node) string {
	var buf bytes.Buffer
	ShellPrint(&buf, n)
	return buf.String()
}
