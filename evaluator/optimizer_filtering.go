package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/log"
)

func optimizeFiltering(n node, _ *EvalCtx, logger *log.Logger) (node, error) {
	v := &filteringOptimizer{
		allowPredicate: true,
		logger:         logger,
	}
	newN, err := v.visit(n)
	if err != nil {
		return nil, err
	}

	if len(v.predicateParts) != 0 {
		v.logger.Errf(log.Always, "filtering optimizer failed to re-add all predicate parts. skipping optimization.")
		return n, nil
	}

	return newN, nil
}

type filteringOptimizer struct {
	logger         *log.Logger
	predicateParts expressionParts
	tableNames     []string
	allowPredicate bool
	root           PlanStage
}

func (v *filteringOptimizer) visit(n node) (node, error) {
	switch typedN := n.(type) {
	case *FilterStage:
		if v.canMoveFilter(typedN) {
			parts, err := splitExpressionIntoParts(typedN.matcher)
			if err != nil {
				return nil, err
			}
			v.predicateParts = append(v.predicateParts, parts...)
			return v.visit(typedN.source)
		}

	case *MongoSourceStage:
		v.tableNames = append(v.tableNames, typedN.aliasNames...)

		if v.allowPredicate {
			combined, remaining := v.getMatchingPredicate()
			if combined != nil {
				n = NewFilterStage(typedN, combined)
			}
			v.predicateParts = remaining
		}

		return n, nil
	case *DynamicSourceStage:
		v.tableNames = append(v.tableNames, typedN.aliasName)

		if v.allowPredicate {
			combined, remaining := v.getMatchingPredicate()
			if combined != nil {
				n = NewFilterStage(typedN, combined)
			}
			v.predicateParts = remaining
		}

		return n, nil
	case *JoinStage:
		oldAllowPredicate := v.allowPredicate
		v.allowPredicate = true
		v.tableNames = nil
		left, err := v.visit(typedN.left)
		if err != nil {
			return nil, err
		}

		v.allowPredicate = false
		tableNames := v.tableNames
		v.tableNames = nil
		right, err := v.visit(typedN.right)
		if err != nil {
			return nil, err
		}
		v.tableNames = append(tableNames, v.tableNames...)
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

		plan, err := optimizeFiltering(typedN.plan, nil, v.logger)
		if err != nil {
			return nil, err
		}

		if plan != typedN.plan {
			n = &SQLSubqueryExpr{
				correlated: typedN.correlated,
				plan:       plan.(PlanStage),
			}
		}

		return n, nil

	case *SubquerySourceStage:
		v.tableNames = append(v.tableNames, typedN.aliasName)

		source, err := optimizeFiltering(typedN.source, nil, v.logger)
		if err != nil {
			return nil, err
		}
		if source != typedN.source {
			n = NewSubquerySourceStage(source.(PlanStage), typedN.selectID, typedN.aliasName)
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
		left, err := optimizeFiltering(typedN.left, nil, v.logger)
		if err != nil {
			return nil, err
		}

		right, err := optimizeFiltering(typedN.right, nil, v.logger)
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
		for _, partTableName := range part.tableNames {
			if !containsString(v.tableNames, partTableName) {
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
		case *ProjectStage, *GroupByStage, *LimitStage, *UnionStage:
			return false
		default:
			panic(fmt.Sprintf("unsupported PlanStage (%T)", source))
		}
	}
}
