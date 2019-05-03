package astprint

import (
	"io"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/parser"
)

// ShellPrintMatchExpr turns an expression into a bson.Value suitable for use in a match aggregation stage.
func ShellPrintMatchExpr(w io.Writer, e ast.Expr) {
	ShellPrintConstant(w, parser.DeparseMatchExpr(e))
}

// ShellPrintExpr turns an expression into a bson.Value suitable for use in a non-match aggregation stage.
func ShellPrintExpr(w io.Writer, e ast.Expr, inProject ...bool) {
	ShellPrintConstant(w, parser.DeparseExpr(e))
}
