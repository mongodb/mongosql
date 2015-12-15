package evaluator

// partiallyEvaluate will take an expression tree and partially evaluate any nodes that can
// evaluated without needing data.
func partiallyEvaluate(e SQLExpr) (SQLExpr, error) {
	candidates, err := nominateForPartialEvaluation(e)
	if err != nil {
		return nil, err
	}
	v := &partialEvaluator{candidates}
	return v.Visit(e)
}

type partialEvaluator struct {
	candidates map[SQLExpr]bool
}

func (pe *partialEvaluator) Visit(e SQLExpr) (SQLExpr, error) {
	if !pe.candidates[e] {
		return walk(pe, e)
	}

	// if we need an evaluation context, the partialEvaluatorNominator
	// is returning bad candidates.
	return e.Evaluate(nil)
}

func nominateForPartialEvaluation(e SQLExpr) (map[SQLExpr]bool, error) {
	n := &partialEvaluatorNominator{
		candidates: make(map[SQLExpr]bool),
	}
	_, err := n.Visit(e)
	if err != nil {
		return nil, err
	}

	return n.candidates, nil
}

type partialEvaluatorNominator struct {
	blocked    bool
	candidates map[SQLExpr]bool
}

func (n *partialEvaluatorNominator) Visit(e SQLExpr) (SQLExpr, error) {
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

// normalize makes semantically equivalent expressions all
// look the same. For instance, it will make "3 > a" look like
// "a < 3".
func normalize(e SQLExpr) (SQLExpr, error) {
	v := &normalizer{}
	return v.Visit(e)
}

type normalizer struct{}

func (n *normalizer) Visit(e SQLExpr) (SQLExpr, error) {

	// walk the children first as they might get normalized
	// on the way up.
	e, err := walk(n, e)
	if err != nil {
		return nil, err
	}

	switch typedE := e.(type) {
	case *SQLAndExpr:
		if left, ok := typedE.left.(SQLValue); ok {
			matches, err := Matches(left, nil)
			if err != nil {
				return nil, err
			}
			if matches {
				return typedE.right, nil
			} else {
				return SQLFalse, nil
			}
		}
		if right, ok := typedE.right.(SQLValue); ok {
			matches, err := Matches(right, nil)
			if err != nil {
				return nil, err
			}
			if matches {
				return typedE.left, nil
			} else {
				return SQLFalse, nil
			}
		}
	case *SQLEqualsExpr:
		if shouldFlip(sqlBinaryNode(*typedE)) {
			return &SQLEqualsExpr{typedE.right, typedE.left}, nil
		}
	case *SQLGreaterThanExpr:
		if shouldFlip(sqlBinaryNode(*typedE)) {
			return &SQLLessThanExpr{typedE.right, typedE.left}, nil
		}
	case *SQLGreaterThanOrEqualExpr:
		if shouldFlip(sqlBinaryNode(*typedE)) {
			return &SQLLessThanOrEqualExpr{typedE.right, typedE.left}, nil
		}
	case *SQLLessThanExpr:
		if shouldFlip(sqlBinaryNode(*typedE)) {
			return &SQLGreaterThanExpr{typedE.right, typedE.left}, nil
		}
	case *SQLLessThanOrEqualExpr:
		if shouldFlip(sqlBinaryNode(*typedE)) {
			return &SQLGreaterThanOrEqualExpr{typedE.right, typedE.left}, nil
		}
	case *SQLNotEqualsExpr:
		if shouldFlip(sqlBinaryNode(*typedE)) {
			return &SQLNotEqualsExpr{typedE.right, typedE.left}, nil
		}
	case *SQLOrExpr:
		if left, ok := typedE.left.(SQLValue); ok {
			matches, err := Matches(left, nil)
			if err != nil {
				return nil, err
			}
			if matches {
				return SQLTrue, nil
			} else {
				return typedE.right, nil
			}
		}
		if right, ok := typedE.right.(SQLValue); ok {
			matches, err := Matches(right, nil)
			if err != nil {
				return nil, err
			}
			if matches {
				return SQLTrue, nil
			} else {
				return typedE.left, nil
			}
		}
	}

	return e, nil
}

func shouldFlip(n sqlBinaryNode) bool {
	if _, ok := n.left.(SQLValue); ok {
		if _, ok := n.right.(SQLValue); !ok {
			return true
		}
	}

	return false
}
