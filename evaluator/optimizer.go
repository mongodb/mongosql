package evaluator

import (
	"os"

	"github.com/mongodb/mongo-tools/common/log"
)

// OptimizeOperator applies optimizations to the operator tree to
// aid in performance.
func OptimizeOperator(ctx *ExecutionCtx, o Operator) (Operator, error) {
	if os.Getenv(NoOptimize) != "" {
		return o, nil
	}

	newO, err := optimizeCrossJoins(o)
	if err != nil {
		return o, nil
	}
	o = newO

	log.Logf(log.DebugHigh, "Cross Join Optimization query plan: \n%v\n", PrettyPrintPlan(o))

	newO, err = optimizePushDown(ctx, o)
	if err != nil {
		return o, nil
	}
	o = newO

	log.Logf(log.DebugHigh, "Optimized query plan: \n%v\n", PrettyPrintPlan(o))
	return o, nil
}
