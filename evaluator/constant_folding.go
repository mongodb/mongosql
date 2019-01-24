package evaluator

type constantFolder struct {
	cfg *OptimizerConfig
}

func (v *constantFolder) visit(n Node) (Node, error) {
	// we do not need to check the Node output because
	// the Node can only change if it is a SQLExpr, which
	// is caught below. Any other changes will happen in place.
	// We also do not need to check the error as walk can only
	// return errors from visit, and this visitor never returns
	// errors, but the linter requires us to check.
	_, err := walk(v, n)
	if err != nil {
		return nil, err
	}
	// check if the newNode is a SQLExpr.
	expr, ok := n.(SQLExpr)
	if ok {
		folded := expr.FoldConstants(v.cfg)
		return folded, nil
	}
	return n, nil
}

// FoldConstants simplifies all expressions under the passed Node n where possible,
// based on the presence of embedded constants. For example, x + 0 can be simplified
// to x.
func FoldConstants(cfg *OptimizerConfig, n Node) Node {
	cf := &constantFolder{cfg: cfg}
	// Visit cannot actually return an error for FoldConstants.
	folded, _ := cf.visit(n)
	return folded
}
