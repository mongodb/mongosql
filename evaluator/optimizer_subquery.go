package evaluator

import (
	"errors"

	"github.com/10gen/sqlproxy/log"
)

func optimizeSubqueries(ctx ConnectionCtx, logger *log.Logger, n node, execute bool) (node, error) {
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

func (v *subqueryOptimizer) visit(n node) (node, error) {
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
				evalCtx := NewEvalCtx(NewExecutionCtx(v.ctx), v.ctx.Variables().CollationConnection)
				n, err = typedN.Evaluate(evalCtx)
				if err != nil {
					return nil, err
				}
			}

		}
	}

	return n, nil
}
