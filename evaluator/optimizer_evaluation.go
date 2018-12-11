package evaluator

import (
	"github.com/10gen/sqlproxy/log"
)

// OptimizeEvaluations takes a Node and optimizes it by normalizing
// it into a semantically equivalent tree and partially evaluating
// any subtrees that are evaluatable without data.
func OptimizeEvaluations(cfg *OptimizerConfig, n Node) (Node, error) {
	var newN Node = n
	var err error

	if cfg.optimizeEvaluations {
		newN, err = FoldConstants(cfg, n)
		if err != nil {
			return nil, err
		}
	} else {
		cfg.lg.Warnf(log.Admin, "optimize_evaluations is false: skipping evaluation optimizer")
	}

	newN, err = reconcileExprs(cfg, newN)
	if err != nil {
		return nil, err
	}

	if cfg.optimizeEvaluations {
		newN, err = FoldConstants(cfg, newN)
		if err != nil {
			return nil, err
		}
	}

	return newN, nil
}
