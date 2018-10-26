package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/parser"
)

const mongoPrimaryKey string = "_id"

// QueryResult represents the result of a query. It contains
// the columns in the result set and a result iterator.
type QueryResult struct {
	Columns []*Column
	Iter    ErrCloser
	Stats   *PlanStats
}

// PlanStats contains some statistics about a query plan.
type PlanStats struct {
	FullyPushedDown bool
	Explain         []*ExplainRecord
}

// EvaluateCommand runs a command, returning any error
// encountered during execution.
func EvaluateCommand(ctx context.Context, aCfg *AlgebrizerConfig, eCfg *ExecutionConfig) error {

	cmd, err := AlgebrizeCommand(aCfg)
	if err != nil {
		return err
	}

	eCfg.lg.Debugf(
		log.Admin,
		"executing command query plan: \n%v",
		PrettyPrintCommand(cmd),
	)

	st := NewExecutionState()

	return cmd.Execute(ctx, eCfg, st)
}

// EvaluateExplain algebrizes, optimizes, and translates a query, returning
// metadata about the generated plan instead of executing it.
func EvaluateExplain(ctx context.Context, aCfg *AlgebrizerConfig, oCfg *OptimizerConfig, pCfg *PushdownConfig, eCfg *ExecutionConfig) (*QueryResult, error) {

	_, ok := aCfg.stmt.(parser.SelectStatement)
	if !ok {
		return nil, mysqlerrors.Newf(
			mysqlerrors.ErNotSupportedYet,
			"no support for explain (%s) for now",
			aCfg.sql,
		)
	}

	aCfg.lg.Infof(log.Admin, `generating explain plan for statement: "%v"`, aCfg.sql)

	var plan PlanStage

	algebrized, err := AlgebrizeQuery(aCfg)
	if err != nil {
		// We can't create a query plan, so we have to exit.
		return nil, err
	}
	plan = algebrized

	aCfg.lg.Debugf(log.Admin,
		"query plan: \n%v",
		PrettyPrintPlan(plan),
	)

	optimized, err := OptimizePlan(ctx, oCfg, plan)
	if err != nil {
		return nil, err
	}
	plan = optimized

	pushedDown, err := PushdownPlan(pCfg, plan)
	if err != nil && !IsPushdownError(err) {
		return nil, err
	}
	plan = pushedDown

	st := NewExecutionState()
	explainPlan := NewExplainStage(plan, eCfg)
	iter, err := explainPlan.Open(ctx, pCfg, eCfg, st)
	if err != nil {
		// couldn't get an iterator, so we have to exit
		return nil, err
	}

	res := &QueryResult{
		Columns: explainPlan.Columns(),
		Iter:    iter,
	}

	return res, nil
}

// EvaluateQuery algebrizes, optimizes, translates, and executes a query
// according to the provided configuration structs.
func EvaluateQuery(ctx context.Context, aCfg *AlgebrizerConfig, oCfg *OptimizerConfig, pCfg *PushdownConfig, eCfg *ExecutionConfig) (*QueryResult, error) {

	var plan PlanStage

	// Step 1: Algebrize
	algebrized, err := AlgebrizeQuery(aCfg)
	if err = util.CheckForContextCancellationAndError(ctx, err); err != nil {
		return nil, err
	}

	plan = algebrized

	// Step 2: Optimize
	optimized, err := OptimizePlan(ctx, oCfg, plan)
	if err = util.CheckForContextCancellationAndError(ctx, err); err != nil {
		return nil, err
	}

	plan = optimized

	// Step 3: Push Down
	pushedDown, err := PushdownPlan(pCfg, plan)
	err = util.CheckForContextCancellationAndError(ctx, err)
	if err != nil && !IsPushdownError(err) {
		return nil, err
	}

	plan = pushedDown

	// Step 4: Gather query plan statistics
	stats, err := getPlanStats(plan, pCfg)
	if err != nil {
		return nil, err
	}

	// Step 5: Execute
	iter, err := ExecutePlan(ctx, eCfg, plan)
	if err = util.CheckForContextCancellationAndError(ctx, err); err != nil {
		return nil, err
	}

	res := &QueryResult{
		Columns: plan.Columns(),
		Iter:    iter,
		Stats:   stats,
	}

	return res, nil
}
