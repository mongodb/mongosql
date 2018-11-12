package evaluator

import (
	"context"

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

	newN, err := Normalize(n, cfg.sqlValueKind)
	if err != nil {
		return nil, err
	}

	newN, err = partiallyEvaluate(cfg, newN)
	if err != nil {
		return nil, err
	}

	if n != newN {
		// normalized and partially evaluated trees might allow for further
		// optimization
		return OptimizeEvaluations(cfg, newN)
	}

	return newN, nil
}

// partiallyEvaluate will take a PlanStage and partially evaluate any nodes that can
// evaluated without needing data from the database. It functions by using the
// nominateForPartialEvaluation function to gather candidates that are evaluatable. Then
// it walks the tree from top-down and, when it finds a candidate Node, replaces the
// candidate Node with the result of calling Evaluate on the candidate Node.
func partiallyEvaluate(cfg *OptimizerConfig, n Node) (Node, error) {
	candidates, err := nominateForPartialEvaluation(n)
	if err != nil {
		return nil, err
	}
	v := &partialEvaluator{
		cfg:        cfg,
		candidates: candidates,
	}
	return v.visit(n)
}

type partialEvaluator struct {
	cfg        *OptimizerConfig
	candidates map[Node]bool
}

// visit walks the tree from top-down, utilizing the candidates
// for whether or not to evaluate a particular SQLExpr.
func (v *partialEvaluator) visit(n Node) (Node, error) {
	if !v.candidates[n] {
		return walk(v, n)
	}

	ctx := context.Background()
	cfg := v.cfg.executionCfg
	st := NewExecutionState()
	return (n.(SQLExpr)).Evaluate(ctx, cfg, st)
}

// nominateForPartialEvaluation walks a SQLExpr tree from bottom up
// identifying nodes that are able to be evaluated without executing
// a query. It returns these identified nodes as candidates.
func nominateForPartialEvaluation(n Node) (map[Node]bool, error) {
	v := &partialEvaluatorNominator{
		candidates: make(map[Node]bool),
	}
	_, err := v.visit(n)
	if err != nil {
		return nil, err
	}

	return v.candidates, nil
}

type partialEvaluatorNominator struct {
	blocked    bool
	candidates map[Node]bool
}

func (v *partialEvaluatorNominator) visit(n Node) (Node, error) {
	oldBlocked := v.blocked
	v.blocked = false

	switch typedN := n.(type) {
	case *SQLAssignmentExpr:
		// We can't evaluate the SQLVariableExpr inside a SQLAssignment, so we skip it
		// entirely which means it won't be in the candidates list.
		_, err := walk(v, typedN.expr)
		if err != nil {
			return nil, err
		}
	default:
		_, err := walk(v, n)
		if err != nil {
			return nil, err
		}
	}

	if !v.blocked {
		switch typedN := n.(type) {
		case SkipConstantFolding:
			v.blocked = typedN.SkipConstantFolding()
		case *AlterCommand,
			*FlushCommand,
			*KillCommand,
			*MongoFilterExpr,
			PlanStage,
			*SetCommand,
			*SQLAssignmentExpr,
			SQLColumnExpr,
			*SQLExistsExpr,
			*SQLSubqueryExpr,
			SQLAggFunctionExpr:

			v.blocked = true
		}

		if !v.blocked {
			v.candidates[n] = true
		}
	}

	v.blocked = v.blocked || oldBlocked
	return n, nil
}

// SkipConstantFolding can be implemented by a node to indicate whether
// it should be evaluated during constant folding, overriding the default
// behavior of the partialEvaluatorNominator.
type SkipConstantFolding interface {
	SkipConstantFolding() bool
}

// Normalize descends through the semantic tree
// and calls normalize() on each that supports
// normalization.
func Normalize(n Node, kind SQLValueKind) (Node, error) {
	v := &normalizer{kind: kind}
	return v.visit(n)
}

type normalizer struct {
	kind SQLValueKind
}

func (v *normalizer) visit(n Node) (Node, error) {

	// walk the children first as they might get normalized
	// on the way up.
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	if normalizer, ok := n.(normalizingNode); ok {
		return normalizer.Normalize(v.kind), nil
	}

	return n, nil
}
