package evaluator

func optimizeCrossJoins(o Operator) (Operator, error) {
	v := &crossJoinOptimizer{}
	return v.Visit(o)
}

type crossJoinOptimizer struct {
	predicateParts expressionParts
	tableNames     []string
}

func (v *crossJoinOptimizer) Visit(o Operator) (Operator, error) {
	var err error
	switch typedO := o.(type) {
	case *Filter:
		// save the old parts before assigning a new one
		old := v.predicateParts
		v.predicateParts, err = splitExpressionIntoParts(typedO.matcher)
		if err != nil {
			return nil, err
		}

		// Walk the children and let the joins optimize with relevant predicateParts
		source, err := v.Visit(typedO.source)
		if err != nil {
			return nil, err
		}

		if source != typedO.source {
			if len(v.predicateParts) > 0 {
				// if the parts haven't been fully utilized,
				// add a Filter back into the tree with the remaining
				// parts.
				o = &Filter{
					source:  source,
					matcher: v.predicateParts.combine(),
				}
			} else {
				o = source
			}
		}

		// reset the parts back to the way it was
		v.predicateParts = old
		return o, nil
	case *Join:
		if len(v.predicateParts) > 0 && typedO.matcher == nil && (typedO.kind == InnerJoin || typedO.kind == CrossJoin) {
			// We have a filter and a join without any criteria
			v.tableNames = nil
			left, err := v.Visit(typedO.left)
			if err != nil {
				return nil, err
			}
			tableNames := v.tableNames
			v.tableNames = nil
			right, err := v.Visit(typedO.right)
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
			if len(partsToUse) > 0 || left != typedO.left || right != typedO.right {
				var predicate SQLExpr
				kind := CrossJoin
				if len(partsToUse) > 0 {
					kind = InnerJoin
					predicate = partsToUse.combine()
				}

				o = &Join{
					kind:    kind,
					left:    left,
					right:   right,
					matcher: predicate,
				}
			}
			return o, nil
		}

	case *MongoSource:
		v.tableNames = append(v.tableNames, typedO.aliasName)
		return o, nil
	case *Subquery:
		v.tableNames = append(v.tableNames, typedO.tableName)

		// We are going to create a whole new visitor and run subqueries in their own context.
		// Ultimately, they end up with a single table name we can use in the current context.
		source, err := optimizeCrossJoins(typedO.source)
		if err != nil {
			return nil, err
		}
		if source != typedO.source {
			o = &Subquery{
				source:    source,
				tableName: typedO.tableName,
			}
		}
		return o, nil
	}

	return walkOperatorTree(v, o)
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
