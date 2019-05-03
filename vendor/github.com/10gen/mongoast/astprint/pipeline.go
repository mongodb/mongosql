package astprint

import (
	"fmt"
	"io"

	"github.com/10gen/mongoast/ast"
)

// ShellPrintPipeline prints a pipeline for pasting into the MongoShell.
func ShellPrintPipeline(w io.Writer, p *ast.Pipeline, printNewLines ...bool) {
	tab, newline := "", ""
	if len(printNewLines) != 0 && printNewLines[0] {
		tab, newline = "\t", "\n"
	}
	fmt.Fprintln(w, "[")
	l := len(p.Stages) - 1
	for i := 0; i < l; i++ {
		fmt.Fprint(w, tab)
		ShellPrintStage(w, p.Stages[i])
		fmt.Fprint(w, ",", newline)
	}
	fmt.Fprint(w, tab)
	ShellPrintStage(w, p.Stages[l])
	fmt.Fprintln(w, newline, "]")
}
