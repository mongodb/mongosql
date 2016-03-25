package evaluator

func optimizeOperatorSQLExprs(o Operator) (Operator, error) {
	v := &sqlExprOptimizer{}
	return v.Visit(o)
}

type sqlExprOptimizer struct{}

func (v *sqlExprOptimizer) Visit(o Operator) (Operator, error) {
	o, err := walkOperatorTree(v, o)
	if err != nil {
		return nil, err
	}

	switch typedO := o.(type) {
	case *Filter:
		if typedO.matcher == nil {
			break
		}
		matcher, err := OptimizeSQLExpr(typedO.matcher)
		if err != nil {
			return nil, err
		}

		if matcher != typedO.matcher {
			newO := typedO.clone()
			newO.matcher = matcher
			o = newO
		}
	case *GroupBy:
		hasNew := false
		keyExprs := SelectExpressions{}
		for _, ke := range typedO.keyExprs {
			newExpr, err := v.optimizeSelectExpression(&ke)
			if err != nil {
				return nil, err
			}

			if newExpr != &ke {
				hasNew = true
				ke = *newExpr
			}

			keyExprs = append(keyExprs, ke)
		}

		selectExprs := SelectExpressions{}
		for _, se := range typedO.selectExprs {
			newExpr, err := v.optimizeSelectExpression(&se)
			if err != nil {
				return nil, err
			}

			if newExpr != &se {
				hasNew = true
				se = *newExpr
			}

			selectExprs = append(selectExprs, se)
		}

		if hasNew {
			newO := typedO.clone()
			newO.keyExprs = keyExprs
			newO.selectExprs = selectExprs
			o = newO
		}

	case *Join:
		if typedO.matcher == nil {
			break
		}

		matcher, err := OptimizeSQLExpr(typedO.matcher)
		if err != nil {
			return nil, err
		}

		if matcher != typedO.matcher {
			newO := typedO.clone()
			newO.matcher = matcher
			o = newO
		}
	case *OrderBy:
		hasNewKey := false
		keys := []orderByKey{}
		for _, k := range typedO.keys {
			newExpr, err := v.optimizeSelectExpression(k.expr)
			if err != nil {
				return nil, err
			}
			if newExpr != k.expr {
				hasNewKey = true
				k = k.clone()
				k.expr = newExpr
			}

			keys = append(keys, k)
		}

		if hasNewKey {
			newO := typedO.clone()
			newO.keys = keys
			o = newO
		}
	case *Project:
		hasNew := false
		exprs := SelectExpressions{}
		for _, expr := range typedO.sExprs {
			newExpr, err := v.optimizeSelectExpression(&expr)
			if err != nil {
				return nil, err
			}
			if newExpr != &expr {
				hasNew = true
				expr = *newExpr
			}

			exprs = append(exprs, expr)
		}

		if hasNew {
			newO := typedO.clone()
			newO.sExprs = exprs
			o = newO
		}
	}

	return o, nil
}

func (v *sqlExprOptimizer) optimizeSelectExpression(se *SelectExpression) (*SelectExpression, error) {
	expr, err := OptimizeSQLExpr(se.Expr)
	if err != nil {
		return nil, err
	}
	if expr != se.Expr {
		se = se.clone()
		se.Expr = expr
	}

	return se, nil
}
