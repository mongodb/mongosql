package evaluator

type partialEvaluator struct {
	candidates map[SQLExpr]bool
}

func (pe *partialEvaluator) Visit(e SQLExpr) (SQLExpr, error) {
	if !pe.candidates[e] {
		return walk(pe, e)
	}

	// if we need an evaluation context, or nominator is returning
	// bad candidates.
	return e.Evaluate(nil)
}

func nominateForPartialEvaluation(e SQLExpr) (map[SQLExpr]bool, error) {
	n := &nominator{
		candidates: make(map[SQLExpr]bool),
	}
	_, err := n.Visit(e)
	if err != nil {
		return nil, err
	}

	return n.candidates, nil
}

type nominator struct {
	blocked    bool
	candidates map[SQLExpr]bool
}

func (n *nominator) Visit(e SQLExpr) (SQLExpr, error) {
	oldBlocked := n.blocked
	n.blocked = false

	switch e.(type) {
	case *SQLExistsExpr:
		n.blocked = true
	case SQLFieldExpr:
		n.blocked = true
	case *SQLSubqueryCmpExpr:
		n.blocked = true
	case *SQLSubqueryExpr:
		n.blocked = true
	default:
		_, err := walk(n, e)
		if err != nil {
			return nil, err
		}
	}

	if !n.blocked {
		n.candidates[e] = true
	}

	n.blocked = n.blocked || oldBlocked
	return e, nil
}
