package evaluator

import (
	"github.com/10gen/sqlproxy/log"
)

type constantFolder struct {
	cfg *OptimizerConfig
}

func (v *constantFolder) visit(n Node) (Node, error) {
	// We do not need to check the error as walk can only return errors from
	// visit, and this visitor never returns errors, but the linter requires us
	// to check.
	n, err := walk(v, n)
	if err != nil {
		panic(err)
	}

	// If the newNode is a SQLExpr, attempt to constant fold it. If
	// constant-folding succeeds, return the folded expr.
	if expr, ok := n.(SQLExpr); ok {
		newN, err := expr.FoldConstants(v.cfg)
		if err == nil {
			return newN, nil
		}
		v.cfg.lg.Warnf(log.Admin, "error running FoldConstants: %v", err)
	}

	return n, nil
}

// FoldConstants simplifies all expressions under the passed Node n where possible,
// based on the presence of embedded constants. For example, x + 0 can be simplified
// to x.
func FoldConstants(cfg *OptimizerConfig, n Node) (Node, error) {
	cf := &constantFolder{cfg: cfg}
	return cf.visit(n)
}
