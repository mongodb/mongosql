package evaluator

func optimizeCrossJoins(p PlanStage) (PlanStage, error) {
	v := &crossJoinOptimizer{}
	return v.Visit(p)
}

type crossJoinOptimizer struct {
	predicateParts expressionParts
	tableNames     []string
}

func (v *crossJoinOptimizer) Visit(p PlanStage) (PlanStage, error) {
	var err error
	switch typedP := p.(type) {
	case *FilterStage:
		// save the old parts before assigning a new one
		old := v.predicateParts
		v.predicateParts, err = splitExpressionIntoParts(typedP.matcher)
		if err != nil {
			return nil, err
		}

		// Walk the children and let the joins optimize with relevant predicateParts
		source, err := v.Visit(typedP.source)
		if err != nil {
			return nil, err
		}

		if source != typedP.source {
			if len(v.predicateParts) > 0 {
				// if the parts haven't been fully utilized,
				// add a Filter back into the tree with the remaining
				// parts.
				p = &FilterStage{
					source:  source,
					matcher: v.predicateParts.combine(),
				}
			} else {
				p = source
			}
		}

		// reset the parts back to the way it was
		v.predicateParts = old
		return p, nil
	case *JoinStage:
		matcherOk := typedP.matcher == nil
		if !matcherOk {
			switch typedM := typedP.matcher.(type) {
			case SQLBool:
				matcherOk = typedM.Value().(bool)
			}
		}
		if matcherOk && (typedP.kind == InnerJoin || typedP.kind == CrossJoin) && len(v.predicateParts) > 0 {
			// We have a filter and a join without any criteria
			v.tableNames = nil
			left, err := v.Visit(typedP.left)
			if err != nil {
				return nil, err
			}
			tableNames := v.tableNames
			v.tableNames = nil
			right, err := v.Visit(typedP.right)
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
			if len(partsToUse) > 0 || left != typedP.left || right != typedP.right {
				var predicate SQLExpr
				kind := CrossJoin
				if len(partsToUse) > 0 {
					kind = InnerJoin
					predicate = partsToUse.combine()
				}

				p = &JoinStage{
					kind:    kind,
					left:    left,
					right:   right,
					matcher: predicate,
				}
			}
			return p, nil
		}

	case *MongoSourceStage:
		v.tableNames = append(v.tableNames, typedP.aliasName)
		return p, nil
	case *SubqueryStage:
		v.tableNames = append(v.tableNames, typedP.tableName)

		// We are going to create a whole new visitor and run subqueries in their own context.
		// Ultimately, they end up with a single table name we can use in the current context.
		source, err := optimizeCrossJoins(typedP.source)
		if err != nil {
			return nil, err
		}
		if source != typedP.source {
			p = &SubqueryStage{
				source:    source,
				tableName: typedP.tableName,
			}
		}
		return p, nil
	}

	return walkPlanTree(v, p)
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
