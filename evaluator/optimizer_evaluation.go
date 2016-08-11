package evaluator

// optimizeEvaluations takes a node and optimizes it by normalizing
// it into a semantically equivalent tree and partially evaluating
// any subtrees that are evaluatable without data.
func optimizeEvaluations(ctx *EvalCtx, n node) (node, error) {

	newN, err := normalize(n)
	if err != nil {
		return nil, err
	}

	newN, err = partiallyEvaluate(ctx, newN)
	if err != nil {
		return nil, err
	}

	if n != newN {
		// normalized and partially evaluated trees might allow for further
		// optimization
		return optimizeEvaluations(ctx, newN)
	}

	return newN, nil
}

// partiallyEvaluate will take a PlanStage and partially evaluate any nodes that can
// evaluated without needing data from the database. It functions by using the
// nominateForPartialEvaluation function to gather candidates that are evaluatable. Then
// it walks the tree from top-down and, when it finds a candidate node, replaces the
// candidate node with the result of calling Evaluate on the candidate node.
func partiallyEvaluate(ctx *EvalCtx, n node) (node, error) {
	candidates, err := nominateForPartialEvaluation(n)
	if err != nil {
		return nil, err
	}
	v := &partialEvaluator{ctx, candidates}
	return v.visit(n)
}

type partialEvaluator struct {
	ctx        *EvalCtx
	candidates map[node]bool
}

// visit walks the tree from top-down, utilizing the candidates
// for whether or not to evaluate a particular SQLExpr.
func (v *partialEvaluator) visit(n node) (node, error) {
	if !v.candidates[n] {
		return walk(v, n)
	}

	return (n.(SQLExpr)).Evaluate(v.ctx)
}

// nominateForPartialEvaluation walks a SQLExpr tree from bottom up
// identifying nodes that are able to be evaluated without executing
// a query. It returns these identified nodes as candidates.
func nominateForPartialEvaluation(n node) (map[node]bool, error) {
	v := &partialEvaluatorNominator{
		candidates: make(map[node]bool),
	}
	_, err := v.visit(n)
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
	case *SQLAssignmentExpr:
		// We can't evaluate the SQLVariableExpr inside a SQLAssignment, so we skip it
		// entirely which means it won't be in the candidates list.
		_, err := walk(v, typedN.expr)
		if err != nil {
			return nil, err
		}
	default:
		_, err := walk(v, n)
		if err != nil {
			return nil, err
		}
	}

	if !v.blocked {
		switch typedN := n.(type) {
		case RequiresEvalCtx:
			v.blocked = typedN.RequiresEvalCtx()
		case *KillCommand, *MongoFilterExpr, PlanStage, *SetCommand, *SQLAssignmentExpr, SQLColumnExpr,
			*SQLExistsExpr, *SQLSubqueryCmpExpr, *SQLSubqueryExpr, *SQLAggFunctionExpr:
			v.blocked = true
		}

		if !v.blocked {
			v.candidates[n] = true
		}
	}

	v.blocked = v.blocked || oldBlocked
	return n, nil
}

// normalize descends through the semantic tree
// and calls normalize() on each that supports
// normalization.
func normalize(n node) (node, error) {
	v := &normalizer{}
	return v.visit(n)
}

type normalizer struct{}

func (v *normalizer) visit(n node) (node, error) {

	// walk the children first as they might get normalized
	// on the way up.
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	if normalizer, ok := n.(normalizingNode); ok {
		return normalizer.normalize(), nil
	}

	return n, nil
}
