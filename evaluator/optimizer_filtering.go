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
		v.logger.Warnf(log.Admin, "filtering optimizer failed to re-add all predicate parts. skipping optimization.")
		return n, nil
	}

	return newN, nil
}

type filteringOptimizer struct {
	logger              *log.Logger
	predicateParts      expressionParts
	qualifiedTableNames []string
	allowPredicate      bool
	root                PlanStage
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
		v.qualifiedTableNames = append(v.qualifiedTableNames, fullyQualifiedTableName(typedN.dbName, typedN.aliasName))

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

		plan, err := optimizeFiltering(typedN.plan, nil, v.logger)
		if err != nil {
			return nil, err
		}

		if plan != typedN.plan {
			n = &SQLSubqueryExpr{
				correlated: typedN.correlated,
				plan:       plan.(PlanStage),
				allowRows:  typedN.allowRows,
			}
		}

		return n, nil

	case *SubquerySourceStage:
		dbNames := generateDbSetFromColumns(typedN.Columns())
		for dbName := range dbNames {
			v.qualifiedTableNames = append(v.qualifiedTableNames, fullyQualifiedTableName(dbName, typedN.aliasName))
		}

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
		case *ProjectStage, *GroupByStage, *LimitStage, *UnionStage:
			return false
		default:
			panic(fmt.Sprintf("unsupported PlanStage (%T)", source))
		}
	}
}
