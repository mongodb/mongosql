package evaluator

import (
	"errors"

	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/variable"
)

// OptimizeSubqueries optimizes plans containing subqueries by eagerly executing
// non-correlated subqueries and saving their results in a CacheStage.
func OptimizeSubqueries(ctx ConnectionCtx, logger *log.Logger, n Node, execute bool) (Node, error) {
	v := &subqueryOptimizer{
		logger:  logger,
		ctx:     ctx,
		execute: execute,
	}
	n, err := v.visit(n)
	if err != nil {
		return nil, err
	}
	return n, nil
}

type subqueryOptimizer struct {
	logger  *log.Logger
	ctx     ConnectionCtx
	execute bool
}

func (v *subqueryOptimizer) visit(n Node) (Node, error) {
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	if typedN, ok := n.(*SQLSubqueryExpr); ok {
		if !typedN.correlated {

			v.logger.Infof(log.Dev, "optimizing non-correlated subquery: \n%v", PrettyPrintPlan(typedN.plan))
			n = optimize(v.ctx, n, true)
			typedN, ok = n.(*SQLSubqueryExpr)
			if !ok {
				return nil, errors.New("Optimized subquery plan not rooted in a SQLSubqueryExpr")
			}

			if v.execute {
				v.logger.Infof(log.Dev, "executing non-correlated subquery: \n%v", PrettyPrintPlan(typedN.plan))
				evalCtx := NewEvalCtx(NewExecutionCtx(v.ctx), v.ctx.Variables().GetCollation(variable.CollationConnection))
				// Subqueries in SQLSubqueryCmpExpr can return multiple rows. Attempt to evaluate and cache rows
				if typedN.allowRows {
					v.logger.Infof(log.Dev, "attempting to cache non-correlated subquery")
					typedN.plan, err = cachePlanStage(typedN.plan, evalCtx)
					if err != nil {
						return nil, err
					}
					v.logger.Infof(log.Dev, "non-correlated subquery cached successfully")
					return n, nil
				}
				n, err = typedN.Evaluate(evalCtx)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return n, nil
}
