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
		var newKeys []SQLExpr
		for _, key := range typedP.keys {
			newKey, err := OptimizeSQLExpr(key)
			if err != nil {
				return nil, err
			}

			if newKey != key {
				hasNew = true
				key = newKey
			}

			newKeys = append(newKeys, key)
		}

		var newProjectedColumns ProjectedColumns
		for _, pc := range typedP.projectedColumns {
			newPC, err := v.optimizeProjectedColumn(&pc)
			if err != nil {
				return nil, err
			}

			if newPC != &pc {
				hasNew = true
				pc = *newPC
			}

			newProjectedColumns = append(newProjectedColumns, pc)
		}

		if hasNew {
			newP := typedP.clone()
			newP.keys = newKeys
			newP.projectedColumns = newProjectedColumns
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
		newProjectedColumns := ProjectedColumns{}
		for _, projectedColumn := range typedP.projectedColumns {
			newProjectedColumn, err := v.optimizeProjectedColumn(&projectedColumn)
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

func (v *sqlExprOptimizer) optimizeProjectedColumn(se *ProjectedColumn) (*ProjectedColumn, error) {
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
