package evaluator

func optimizePlanSQLExprs(o PlanStage) (PlanStage, error) {
	v := &sqlExprOptimizer{}
	return v.Visit(o)
}

type sqlExprOptimizer struct{}

func (v *sqlExprOptimizer) Visit(p PlanStage) (PlanStage, error) {
	p, err := walkPlanTree(v, p)
	if err != nil {
		return nil, err
	}

	switch typedP := p.(type) {
	case *FilterStage:
		if typedP.matcher == nil {
			break
		}
		matcher, err := OptimizeSQLExpr(typedP.matcher)
		if err != nil {
			return nil, err
		}

		if matcher != typedP.matcher {
			newP := typedP.clone()
			newP.matcher = matcher
			p = newP
		}
	case *GroupByStage:
		hasNew := false
		keyExprs := SelectExpressions{}
		for _, ke := range typedP.keyExprs {
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
		for _, se := range typedP.selectExprs {
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
			newP := typedP.clone()
			newP.keyExprs = keyExprs
			newP.selectExprs = selectExprs
			p = newP
		}

	case *JoinStage:
		if typedP.matcher == nil {
			break
		}

		matcher, err := OptimizeSQLExpr(typedP.matcher)
		if err != nil {
			return nil, err
		}

		if matcher != typedP.matcher {
			newP := typedP.clone()
			newP.matcher = matcher
			p = newP
		}
	case *OrderByStage:
		hasNewTerm := false
		terms := []*orderByTerm{}
		for _, term := range typedP.terms {
			newExpr, err := OptimizeSQLExpr(term.expr)
			if err != nil {
				return nil, err
			}
			if newExpr != term.expr {
				hasNewTerm = true
				term = term.clone()
				term.expr = newExpr
			}

			terms = append(terms, term)
		}

		if hasNewTerm {
			newP := typedP.clone()
			newP.terms = terms
			p = newP
		}
	case *ProjectStage:
		hasNew := false
		newProjectedColumns := SelectExpressions{}
		for _, projectedColumn := range typedP.projectedColumns {
			newProjectedColumn, err := v.optimizeSelectExpression(&projectedColumn)
			if err != nil {
				return nil, err
			}
			if newProjectedColumn != &projectedColumn {
				hasNew = true
				projectedColumn = *newProjectedColumn
			}

			newProjectedColumns = append(newProjectedColumns, projectedColumn)
		}

		if hasNew {
			newP := typedP.clone()
			newP.projectedColumns = newProjectedColumns
			p = newP
		}
	}

	return p, nil
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
