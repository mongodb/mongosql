package evaluator

import (
	"errors"

	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/variable"
)

func OptimizeSubqueries(ctx ConnectionCtx, logger *log.Logger, n node, execute bool) (node, error) {
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

// cachePlanStage executes a PlanStage within the evalCtx and returns the cached results.
func cachePlanStage(ps PlanStage, evalCtx *EvalCtx) (*CacheStage, error) {
	var iter Iter
	var err error
	execCtx := evalCtx.ExecutionCtx
	if iter, err = ps.Open(execCtx); err != nil {
		return nil, err
	}
	// maxCacheSizeBytes is the maximimum size a single cached query can be in bytes
	// It is set to be equal to the max plan stage size
	maxCacheSizeBytes := execCtx.ConnectionCtx.Variables().GetUInt64(variable.MongoDBMaxStageSize)

	size := uint64(0)
	row, allRows := &Row{}, Rows{}
	for iter.Next(row) {
		if maxCacheSizeBytes != 0 && size > maxCacheSizeBytes {
			return nil, newPlanStageMemoryError(maxCacheSizeBytes)
		}
		allRows = append(allRows, *row)
		size += row.Data.Size()
		row = &Row{}
	}

	if err = iter.Close(); err != nil {
		return nil, err
	}
	if err = iter.Err(); err != nil {
		return nil, err
	}

	return NewCacheStage(size, allRows, ps.Columns(), ps.Collation()), nil
}
