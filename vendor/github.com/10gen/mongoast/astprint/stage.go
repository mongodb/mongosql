package astprint

import (
	"io"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/parser"
)

// ShellPrintStage prints a stage for pasting into the MongoShell.
func ShellPrintStage(w io.Writer, s ast.Stage) {
	ShellPrintConstant(w, parser.DeparseStage(s))
}
