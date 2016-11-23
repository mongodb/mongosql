package evaluator

import (
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

			v.logger.Logf(log.Info, "optimizing non-correlated subquery: \n%v", PrettyPrintPlan(typedN.plan))
			n, err = optimize(v.ctx, n, true)
			if err != nil {
				return nil, err
			}

			if v.execute {
				v.logger.Logf(log.Info, "executing non-correlated subquery: \n%v", PrettyPrintPlan(typedN.plan))
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
