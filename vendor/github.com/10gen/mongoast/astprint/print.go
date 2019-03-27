package astprint

import (
	"bytes"
	"fmt"
	"io"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/parser"

	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// String prints the node to a string.
func String(n ast.Node) string {
	var buf bytes.Buffer
	Print(&buf, n)
	return buf.String()
}

// Print prints the node to the writer.
func Print(w io.Writer, n ast.Node) {
	var v bsoncore.Value
	switch tn := n.(type) {
	case *ast.Pipeline:
		v = parser.DeparsePipeline(tn)
	case ast.Stage:
		v = parser.DeparseStage(tn)
	case ast.Expr:
		v = parser.DeparseExpr(tn)
	}

	fmt.Fprint(w, v.String())
}
