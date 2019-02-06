package evaluator

import (
	"github.com/10gen/sqlproxy/log"
)

// OptimizeEvaluations takes a Node and optimizes it by normalizing
// it into a semantically equivalent tree and partially evaluating
// any subtrees that are evaluatable without data.
func OptimizeEvaluations(cfg *OptimizerConfig, n Node) (Node, error) {

	if !cfg.optimizeEvaluations {
		cfg.lg.Warnf(log.Admin, "optimize_evaluations is false: skipping evaluation optimizer")
		return n, nil
	}

	return FoldConstants(cfg, n)
}
