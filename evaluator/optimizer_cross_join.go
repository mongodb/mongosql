package evaluator

func optimizeCrossJoins(o Operator) (Operator, error) {
	v := &crossJoinOptimizer{}
	return v.Visit(o)
}

type crossJoinOptimizer struct {
	predicateParts []crossJoinPredicatePart
	tableNames     []string
}

func (v *crossJoinOptimizer) Visit(o Operator) (Operator, error) {
	var err error
	switch typedO := o.(type) {
	case *Filter:
		// save the old parts before assigning a new one
		old := v.predicateParts
		v.predicateParts, err = v.getPredicateParts(typedO.matcher)
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
					matcher: v.combinePredicateParts(v.predicateParts),
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
			partsToUse := []crossJoinPredicatePart{}
			savedParts := v.predicateParts
			v.predicateParts = nil
			for _, part := range savedParts {
				if v.canUsePredicatePartInJoinClause(part) {
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
					predicate = v.combinePredicateParts(partsToUse)
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

func (v *crossJoinOptimizer) canUsePredicatePartInJoinClause(part crossJoinPredicatePart) bool {
	contains := func(strs []string, str string) bool {
		for _, n := range strs {
			if n == str {
				return true
			}
		}

		return false
	}

	if len(part.tableNames) > 0 {
		// the right-most table must be present in the part
		if !contains(part.tableNames, v.tableNames[len(v.tableNames)-1]) {
			return false
		}

		// all the names in the part must be in scope
		for _, n := range part.tableNames {
			if !contains(v.tableNames, n) {
				return false
			}
		}

		return true
	}

	return false
}

func (v *crossJoinOptimizer) getPredicateParts(e SQLExpr) ([]crossJoinPredicatePart, error) {
	// this flattens hierarchical SQLAndExprs into a list
	var splitConjunctions func(SQLExpr) []SQLExpr
	splitConjunctions = func(e SQLExpr) []SQLExpr {
		andE, ok := e.(*SQLAndExpr)
		if !ok {
			return []SQLExpr{e}
		}

		left := splitConjunctions(andE.left)
		right := splitConjunctions(andE.right)
		return append(left, right...)
	}

	conjunctions := splitConjunctions(e)
	result := []crossJoinPredicatePart{}
	for _, conjunction := range conjunctions {
		tableNames, err := v.getTableNames(conjunction)
		if err != nil {
			return nil, err
		}
		result = append(result, crossJoinPredicatePart{conjunction, tableNames})
	}
	return result, nil
}

func (v *crossJoinOptimizer) combinePredicateParts(parts []crossJoinPredicatePart) SQLExpr {
	predicate := parts[0].expr
	for _, part := range parts[1:] {
		predicate = &SQLAndExpr{predicate, part.expr}
	}
	return predicate
}

type crossJoinPredicatePart struct {
	expr       SQLExpr
	tableNames []string
}

func (v *crossJoinOptimizer) getTableNames(e SQLExpr) ([]string, error) {
	finder := &crossJoinSQLExprTableNameFinder{}
	_, err := finder.Visit(e)
	if err != nil {
		return nil, err
	}

	return finder.tableNames, nil
}

type crossJoinSQLExprTableNameFinder struct {
	tableNames []string
}

func (v *crossJoinSQLExprTableNameFinder) Visit(e SQLExpr) (SQLExpr, error) {
	switch typedE := e.(type) {
	case SQLColumnExpr:
		v.tableNames = append(v.tableNames, typedE.tableName)
	}
	return walk(v, e)
}
