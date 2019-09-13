package evaluator

import (
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/log"
)

func optimizeFiltering(cfg *OptimizerConfig, n Node) (Node, error) {

	if !cfg.optimizeFiltering {
		cfg.lg.Warnf(log.Admin, "optimize_filtering is false: skipping filtering optimizer")
		return n, nil
	}

	v := newFilteringOptimizer(cfg)
	n, err := v.visit(n)
	if err != nil {
		return nil, err
	}

	if len(v.predicateParts) != 0 {
		v.cfg.lg.Warnf(log.Admin, "filtering optimizer"+
			" failed to re-add all predicate parts. skipping optimization.")
	}

	return n, nil
}

type filteringOptimizer struct {
	cfg                 *OptimizerConfig
	predicateParts      expressionParts
	qualifiedTableNames []string
	allowPredicate      bool
}

func newFilteringOptimizer(cfg *OptimizerConfig) *filteringOptimizer {
	return &filteringOptimizer{
		cfg:            cfg,
		allowPredicate: true,
	}
}

func (v *filteringOptimizer) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case *FilterStage:
		if v.canMoveFilter(typedN) {
			parts, err := getConjunctiveTerms(typedN.matcher)
			if err != nil {
				return nil, err
			}
			v.predicateParts = append(v.predicateParts, parts...)
			newSource, err := v.visit(typedN.source)
			if err != nil {
				return nil, err
			}
			if len(v.predicateParts) == 0 {
				// If len(v.predicateParts) is 0, we managed to move the entire predicate,
				// which means we do not need the FilterStage anymore.
				return newSource, nil
			}
			// Otherwise, we copy the newSource to the source of this FilterStage, because
			// we cannot remove the FilterStage. This might result in some term duplication
			// between the FilterStage and the source, but that will not add a large amount
			// of overhead. Moving part of the filter, even doubling up, is still useful.
			if newSource != typedN.source {
				typedN.source = newSource.(PlanStage)
			}
			return typedN, nil
		}

	case *MongoSourceStage:
		for _, alias := range typedN.aliasNames {
			fullyQualifiedTableName := fullyQualifiedTableName(typedN.dbName, alias)
			v.qualifiedTableNames = append(v.qualifiedTableNames, fullyQualifiedTableName)
		}

		if v.allowPredicate {
			combined, remaining := v.getMatchingPredicate()
			if combined != nil {
				n = NewFilterStage(typedN, combined)
			}
			v.predicateParts = remaining
		}

		return n, nil
	case *DynamicSourceStage:
		v.qualifiedTableNames = append(v.qualifiedTableNames,
			fullyQualifiedTableName(typedN.dbName,
				typedN.aliasName))

		if v.allowPredicate {
			combined, remaining := v.getMatchingPredicate()
			if combined != nil {
				n = NewFilterStage(typedN, combined)
			}
			v.predicateParts = remaining
		}

		return n, nil
	case *JoinStage:
		if !strutil.StringSliceContains(commutativeJoinKinds, string(typedN.kind)) {
			// If we hit a node level where we're unable to optimize - e.g. if it's a left join or a
			// right join - we can possibly further optimize the subtree rooted at this node.
			// For example, in the plan tree below, we can optimize the subtree rooted in B.
			//
			//				A(CrossJoin)
			//				/	\
			//			B(RightJoin)	 C
			//			/	\
			//		D(CrossJoin)	 E
			//		/	\
			//		F	 G

			newL, err := newFilteringOptimizer(v.cfg).visit(typedN.left)
			if err != nil {
				return nil, err
			}
			newR, err := newFilteringOptimizer(v.cfg).visit(typedN.right)
			if err != nil {
				return nil, err
			}
			if typedN.left != newL.(PlanStage) || typedN.right != newR.(PlanStage) {
				n = NewJoinStage(typedN.kind, newL.(PlanStage), newR.(PlanStage), typedN.matcher)
			}
			return n, nil
		}

		oldAllowPredicate := v.allowPredicate
		v.allowPredicate = true
		v.qualifiedTableNames = nil
		left, err := v.visit(typedN.left)
		if err != nil {
			return nil, err
		}

		v.allowPredicate = false
		qualifiedTableNames := v.qualifiedTableNames
		v.qualifiedTableNames = nil

		right, err := v.visit(typedN.right)
		if err != nil {
			return nil, err
		}
		qualifiedTableNames = append(qualifiedTableNames, v.qualifiedTableNames...)

		v.qualifiedTableNames = qualifiedTableNames
		v.allowPredicate = oldAllowPredicate

		if left != typedN.left || right != typedN.right {
			n = NewJoinStage(typedN.kind, left.(PlanStage), right.(PlanStage), typedN.matcher)
		}

		if v.allowPredicate {
			combined, remaining := v.getMatchingPredicate()
			if combined != nil {
				n = NewFilterStage(n.(PlanStage), combined)
			}
			v.predicateParts = remaining
		}

		return n, nil
	case *SQLSubqueryExpr:

		plan, err := optimizeFiltering(v.cfg, typedN.plan)
		if err != nil {
			return nil, err
		}

		if plan != typedN.plan {
			n = NewSQLSubqueryExpr(typedN.correlated, typedN.allowRows, plan.(PlanStage))
		}

		return n, nil

	case SQLDoubleSubqueryExpr:

		leftPlan, err := optimizeFiltering(v.cfg, typedN.LeftPlan())
		if err != nil {
			return nil, err
		}

		rightPlan, err := optimizeFiltering(v.cfg, typedN.RightPlan())
		if err != nil {
			return nil, err
		}

		if !(leftPlan == typedN.LeftPlan() && rightPlan == typedN.RightPlan()) {
			switch subtypedN := typedN.(type) {
			case *SQLSubqueryCmpExpr:
				n = NewSQLSubqueryCmpExpr(subtypedN.leftCorrelated, subtypedN.rightCorrelated,
					leftPlan.(PlanStage), rightPlan.(PlanStage), subtypedN.operator)
			case *SQLSubqueryAnyExpr:
				n = NewSQLSubqueryAnyExpr(subtypedN.leftCorrelated, subtypedN.rightCorrelated,
					leftPlan.(PlanStage), rightPlan.(PlanStage), subtypedN.operator)
			case *SQLSubqueryAllExpr:
				n = NewSQLSubqueryAllExpr(subtypedN.leftCorrelated, subtypedN.rightCorrelated,
					leftPlan.(PlanStage), rightPlan.(PlanStage), subtypedN.operator)
			}
		}

		return n, nil

	case *SubquerySourceStage:
		dbNames := generateDbSetFromColumns(typedN.Columns())
		for dbName := range dbNames {
			v.qualifiedTableNames = append(v.qualifiedTableNames,
				fullyQualifiedTableName(dbName,
					typedN.aliasName))
		}

		source, err := optimizeFiltering(v.cfg, typedN.source)
		if err != nil {
			return nil, err
		}
		if source != typedN.source {
			n = NewSubquerySourceStage(source.(PlanStage), typedN.selectID,
				typedN.dbName, typedN.aliasName, typedN.fromCTE)
		}

		if v.allowPredicate {
			combined, remaining := v.getMatchingPredicate()
			if combined != nil {
				n = NewFilterStage(n.(PlanStage), combined)
			}
			v.predicateParts = remaining
		}

		return n, nil
	case *UnionStage:
		left, err := optimizeFiltering(v.cfg, typedN.left)
		if err != nil {
			return nil, err
		}

		right, err := optimizeFiltering(v.cfg, typedN.right)
		if err != nil {
			return nil, err
		}

		if typedN.left != left || typedN.right != right {
			n = NewUnionStage(typedN.kind, left.(PlanStage), right.(PlanStage))
		}

		return n, nil
	}

	return walk(v, n)
}

func (v *filteringOptimizer) getMatchingPredicate() (SQLExpr, []expressionPart) {
	var partsToAdd []SQLExpr
	var remainingParts []expressionPart
	for _, part := range v.predicateParts {
		valid := true
		for _, partTableName := range part.qualifiedTableNames {
			if !containsString(v.qualifiedTableNames, partTableName) {
				valid = false
				break
			}
		}

		if valid {
			// all of this part's tableNames are in scope.
			partsToAdd = append(partsToAdd, part.expr)
		} else {
			// need to keep this part around to add back in later
			remainingParts = append(remainingParts, part)
		}
	}

	return combineExpressions(partsToAdd), remainingParts
}

func (v *filteringOptimizer) canMoveFilter(fs *FilterStage) bool {
	// we can't move a filter across a projecting stage... these
	// would be GroupBy and Project.
	source := fs.source
	for {
		switch typedS := source.(type) {
		case *MongoSourceStage, *BSONSourceStage, *DynamicSourceStage,
			*EmptyStage, *JoinStage, *SubquerySourceStage:
			return true
		case *OrderByStage:
			source = typedS.source
		default:
			return false
		}
	}
}
