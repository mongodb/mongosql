package evaluator

func optimizeCrossJoins(n node) (node, error) {
	v := &crossJoinOptimizer{}
	n, err := v.visit(n)
	if err != nil {
		return nil, err
	}

	return n, nil
}

type crossJoinOptimizer struct {
	predicateParts expressionParts
	tableNames     []string
}

func (v *crossJoinOptimizer) visit(n node) (node, error) {
	var err error
	switch typedN := n.(type) {
	case *FilterStage:
		// save the old parts before assigning a new one
		old := v.predicateParts
		v.predicateParts, err = splitExpressionIntoParts(typedN.matcher)
		if err != nil {
			return nil, err
		}

		// Walk the children and let the joins optimize with relevant predicateParts
		source, err := v.visit(typedN.source)
		if err != nil {
			return nil, err
		}

		if source != typedN.source {
			if len(v.predicateParts) > 0 {
				// if the parts haven't been fully utilized,
				// add a Filter back into the tree with the remaining
				// parts.
				n = NewFilterStage(source.(PlanStage), v.predicateParts.combine())
			} else {
				n = source
			}
		}

		// reset the parts back to the way it was
		v.predicateParts = old
		return n, nil
	case *JoinStage:
		matcherOk := typedN.matcher == nil
		if !matcherOk {
			switch typedM := typedN.matcher.(type) {
			case SQLBool:
				matcherOk = typedM.Value().(bool)
			}
		}
		if matcherOk && (typedN.kind == InnerJoin || typedN.kind == CrossJoin) && len(v.predicateParts) > 0 {
			// We have a filter and a join without any criteria
			v.tableNames = nil
			left, err := v.visit(typedN.left)
			if err != nil {
				return nil, err
			}
			tableNames := v.tableNames
			v.tableNames = nil
			right, err := v.visit(typedN.right)
			if err != nil {
				return nil, err
			}

			// this is now the table names from the left and the right side
			v.tableNames = append(tableNames, v.tableNames...)

			// go through each part of the predicate and figure out which
			// ones are associated to the tables in the current join.
			partsToUse := expressionParts{}
			savedParts := v.predicateParts
			v.predicateParts = nil
			for _, part := range savedParts {
				if v.canUseExpressionPartInJoinClause(part) {
					partsToUse = append(partsToUse, part)
				} else {
					v.predicateParts = append(v.predicateParts, part)
				}
			}

			// if we have parts or the left or right have been changed, we
			// need a new join operator
			if len(partsToUse) > 0 || left != typedN.left || right != typedN.right {
				var predicate SQLExpr
				kind := CrossJoin
				if len(partsToUse) > 0 {
					kind = InnerJoin
					predicate = partsToUse.combine()
				}

				n = NewJoinStage(kind, left.(PlanStage), right.(PlanStage), predicate)
			}
			return n, nil
		}

	case *MongoSourceStage:
		v.tableNames = append(v.tableNames, typedN.aliasNames...)
		return n, nil
	}

	return walk(v, n)
}

func (v *crossJoinOptimizer) canUseExpressionPartInJoinClause(part expressionPart) bool {
	if len(part.tableNames) > 0 {
		// the right-most table must be present in the part
		if !containsString(part.tableNames, v.tableNames[len(v.tableNames)-1]) {
			return false
		}

		// all the names in the part must be in scope
		for _, n := range part.tableNames {
			if !containsString(v.tableNames, n) {
				return false
			}
		}

		return true
	}

	return false
}
