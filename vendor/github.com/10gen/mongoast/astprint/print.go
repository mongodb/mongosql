package astprint

import (
	"bytes"
	"fmt"
	"io"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/internal/bsonutil"
	"github.com/10gen/mongoast/parser"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// Print prints the node as extended JSON to the writer.
func Print(w io.Writer, n ast.Node) {
	s := ""
	bv := parser.DeparseNode(n)

	// We use the go driver's bson.MarshExtJSON function to turn the deparsed
	// node into an extended JSON string. The bsoncore.Value we get back may
	// not be a document, but marshaling only works for documents. To address
	// this, we wrap the value in a document if necessary and marshal that.
	doc, isWrapped := wrapInDoc("x", bv)
	extJSON, err := bson.MarshalExtJSONWithRegistry(bsonutil.Registry, doc, true, false)

	// If there is an error, we will leave the output string s empty.
	// If there is not error, we strip away the document characters
	// we added above for marshaling purposes and assign the remaining
	// string to s.
	if err == nil {
		if isWrapped {
			// The extended JSON string has the form:
			//   `{"x":<value>}`
			// so we strip away the first 4 characters and the last character.
			s = string(extJSON[5 : len(extJSON)-1])
		} else {
			s = string(extJSON)
		}
	}

	fmt.Fprint(w, s)
}

// String prints the node to an extended JSON string.
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

// wrapInDoc wraps the bsoncore.Value in a document (under the provided key)
// if the value is not already a document. This function returns the document
// value and a boolean indicating if it was wrapped.
func wrapInDoc(key string, bv bsoncore.Value) (bsoncore.Value, bool) {
	if bv.Type == bsontype.EmbeddedDocument {
		return bv, false
	}

	return bsonutil.DocumentFromElements(key, bv), true
}
