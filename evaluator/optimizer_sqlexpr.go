package evaluator

// OptimizeSQLExpr takes a SQLExpr and optimizes it by normalizing
// it into a semantically equivalent tree and partially evaluating
// any subtrees that evaluatable without data.
func OptimizeSQLExpr(e SQLExpr) (SQLExpr, error) {

	newE, err := normalize(e)
	if err != nil {
		return nil, err
	}

	newE, err = partiallyEvaluate(newE)
	if err != nil {
		return nil, err
	}

	if e != newE {
		// normalized and partially evaluated trees might allow for further
		// optimization
		return OptimizeSQLExpr(newE)
	}

	return newE, nil
}

// partiallyEvaluate will take an expression tree and partially evaluate any nodes that can
// evaluated without needing data from the database. It functions by using the
// nominateForPartialEvaluation function to gather candidates that are evaluatable. Then
// it walks the tree from top-down and, when it finds a candidate node, replaces the
// candidate node with the result of calling Evaluate on the candidate node.
func partiallyEvaluate(e SQLExpr) (SQLExpr, error) {
	candidates, err := nominateForPartialEvaluation(e)
	if err != nil {
		return nil, err
	}
	v := &partialEvaluator{candidates}
	n, err := v.visit(e)
	if err != nil {
		return nil, err
	}
	return n.(SQLExpr), nil
}

type partialEvaluator struct {
	candidates map[node]bool
}

func (v *partialEvaluator) visit(n node) (node, error) {
	if !v.candidates[n] {
		return walk(v, n)
	}

	// if we need an evaluation context, the partialEvaluatorNominator
	// is returning bad candidates.
	return (n.(SQLExpr)).Evaluate(nil)
}

// nominateForPartialEvaluation walks a SQLExpr tree from bottom up
// identifying nodes that are able to be evaluated without executing
// a query. It returns these identified nodes as candidates.
func nominateForPartialEvaluation(e SQLExpr) (map[node]bool, error) {
	v := &partialEvaluatorNominator{
		candidates: make(map[node]bool),
	}
	_, err := v.visit(e)
	if err != nil {
		return nil, err
	}

	return v.candidates, nil
}

type partialEvaluatorNominator struct {
	blocked    bool
	candidates map[node]bool
}

func (v *partialEvaluatorNominator) visit(n node) (node, error) {
	oldBlocked := v.blocked
	v.blocked = false

	switch typedN := n.(type) {
	case PlanStage:
		v.blocked = true
	case *SQLExistsExpr:
		v.blocked = true
	case SQLColumnExpr:
		v.blocked = true
	case *SQLSubqueryCmpExpr:
		v.blocked = true
	case *SQLSubqueryExpr:
		v.blocked = true
	case *SQLAggFunctionExpr:
		v.blocked = true
	case *SQLScalarFunctionExpr:
		v.blocked = typedN.RequiresEvalCtx()
		if !v.blocked {
			_, err := walk(v, n)
			if err != nil {
				return nil, err
			}
		}
	default:
		_, err := walk(v, n)
		if err != nil {
			return nil, err
		}
	}

	if !v.blocked {
		v.candidates[n] = true
	}

	v.blocked = v.blocked || oldBlocked
	return n, nil
}

// normalize makes semantically equivalent expressions all
// look the same. For instance, it will make "3 > a" look like
// "a < 3".
func normalize(e SQLExpr) (SQLExpr, error) {
	v := &normalizer{}

	n, err := v.visit(e)
	if err != nil {
		return nil, err
	}

	return n.(SQLExpr), nil
}

type normalizer struct{}

func (v *normalizer) visit(n node) (node, error) {

	shouldFlip := func(n sqlBinaryNode) bool {
		if _, ok := n.left.(SQLValue); ok {
			if _, ok := n.right.(SQLValue); !ok {
				return true
			}
		}

		return false
	}

	// walk the children first as they might get normalized
	// on the way up.
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	switch typedN := n.(type) {
	case *SQLAndExpr:
		if left, ok := typedN.left.(SQLValue); ok {
			matches, err := Matches(left, nil)
			if err != nil {
				return nil, err
			}
			if matches {
				return typedN.right, nil
			}
			return SQLFalse, nil
		}
		if right, ok := typedN.right.(SQLValue); ok {
			matches, err := Matches(right, nil)
			if err != nil {
				return nil, err
			}
			if matches {
				return typedN.left, nil
			}
			return SQLFalse, nil
		}
	case *SQLEqualsExpr:
		if shouldFlip(sqlBinaryNode(*typedN)) {
			return &SQLEqualsExpr{typedN.right, typedN.left}, nil
		}
	case *SQLGreaterThanExpr:
		if shouldFlip(sqlBinaryNode(*typedN)) {
			return &SQLLessThanExpr{typedN.right, typedN.left}, nil
		}
	case *SQLGreaterThanOrEqualExpr:
		if shouldFlip(sqlBinaryNode(*typedN)) {
			return &SQLLessThanOrEqualExpr{typedN.right, typedN.left}, nil
		}
	case *SQLLessThanExpr:
		if shouldFlip(sqlBinaryNode(*typedN)) {
			return &SQLGreaterThanExpr{typedN.right, typedN.left}, nil
		}
	case *SQLLessThanOrEqualExpr:
		if shouldFlip(sqlBinaryNode(*typedN)) {
			return &SQLGreaterThanOrEqualExpr{typedN.right, typedN.left}, nil
		}
	case *SQLNotEqualsExpr:
		if shouldFlip(sqlBinaryNode(*typedN)) {
			return &SQLNotEqualsExpr{typedN.right, typedN.left}, nil
		}
	case *SQLOrExpr:
		if left, ok := typedN.left.(SQLValue); ok {
			matches, err := Matches(left, nil)
			if err != nil {
				return nil, err
			}
			if matches {
				return SQLTrue, nil
			}
			return typedN.right, nil
		}
		if right, ok := typedN.right.(SQLValue); ok {
			matches, err := Matches(right, nil)
			if err != nil {
				return nil, err
			}
			if matches {
				return SQLTrue, nil
			}
			return typedN.left, nil
		}
	case *SQLTupleExpr:
		if len(typedN.Exprs) == 1 {
			return typedN.Exprs[0], nil
		}
	case *SQLValues:
		if len(typedN.Values) == 1 {
			return typedN.Values[0], nil
		}
	}

	return n, nil
}

func optimizePlanSQLExprs(o PlanStage) (PlanStage, error) {
	v := &sqlExprOptimizer{}
	n, err := v.visit(o)
	if err != nil {
		return nil, err
	}

	return n.(PlanStage), nil
}

type sqlExprOptimizer struct{}

func (v *sqlExprOptimizer) visit(n node) (node, error) {

	switch typedN := n.(type) {
	case SQLExpr:
		e, err := OptimizeSQLExpr(typedN)
		if err != nil {
			return nil, err
		}

		return e, nil
	default:
		return walk(v, n)
	}
}
