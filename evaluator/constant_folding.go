package evaluator

type constantFolder struct {
	cfg *OptimizerConfig
}

func (v *constantFolder) visit(n Node) (Node, error) {
	newN, err := walk(v, n)
	if err != nil {
		return nil, err
	}
	// check if the current node is a SQLExpr.
	expr, ok := newN.(SQLExpr)
	if !ok {
		return newN, nil
	}

	out := expr.FoldConstants(v.cfg)
	return out, nil
}

// FoldConstants simplifies all expressions under the passed Node n where possible,
// based on the presence of embedded constants. For example, x + 0 can be simplified
// to x.
func FoldConstants(cfg *OptimizerConfig, n Node) Node {
	cf := &constantFolder{cfg: cfg}
	// Visit cannot actually return an error for FoldConstants.
	out, _ := cf.visit(n)
	return out
}
