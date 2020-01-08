package evaluator

import (
	"context"
	"strings"

	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/parser"
)

const mongoPrimaryKey string = "_id"

// Constant values for QueryOp returned by evaluator
const (
	COMMAND     QueryOp = iota // commands: AlterTable, DropTable, Flush, Kill, RenameTable, Set, Use
	QUERY                      // queries: Select, Union
	SHOW                       // show command
	SHOWNOTIMPL                // show command that is not implemented
	EXPLAIN                    // explain query
	UNKNOWN                    // request didn't match any supported parser.Statement
)

// A QueryOp is returned to the server in a QueryResult to control how the results are handled
type QueryOp byte

// QueryResult represents the result of a query. It contains the parsed statement, the operation
// executed, and optionally the columns in the result set, a result iterator, and PlanStats.
type QueryResult struct {
	Stmt      parser.Statement // allows statement execution tracking in server
	Columns   []*results.Column
	PlanStage PlanStage
	Stats     *PlanStats
	Op        QueryOp
}

// GetDocIter returns the Iter as a DocIter, if it is
// a DocIter, otherwise nil.
func (qr *QueryResult) GetDocIter(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (results.DocIter, error) {
	switch typedPlan := qr.PlanStage.(type) {
	case FastPlanStage:
		return typedPlan.FastOpen(ctx, cfg, st)
	}
	return nil, nil
}

// GetRowIter returns the Iter as a RowIter, if it is
// a RowIter, otherwise nil.
func (qr *QueryResult) GetRowIter(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (results.RowIter, error) {
	return qr.PlanStage.Open(ctx, cfg, st)
}

// NewQueryResult is a constructor for a QueryResult.
func NewQueryResult(stmt parser.Statement, columns []*results.Column, planStage PlanStage, stats *PlanStats, op QueryOp) *QueryResult {
	return &QueryResult{
		Stmt:      stmt,
		Columns:   columns,
		PlanStage: planStage,
		Stats:     stats,
		Op:        op,
	}
}

// EvaluateCommand runs a command, returning a QueryResult containing Op=COMMAND
// so that server knows to writeOK to client or any error encountered during execution.
func EvaluateCommand(ctx context.Context, rCfg *RewriterConfig, aCfg *AlgebrizerConfig, eCfg *ExecutionConfig, stmt parser.Statement) (*QueryResult, error) {

	parsedStmt := stmt.Copy().(parser.Statement)

	cmd, err := AlgebrizeCommand(ctx, aCfg, stmt)
	if err != nil {
		return nil, err
	}

	eCfg.lg.Debugf(
		log.Admin,
		"executing command query plan: \n%v",
		PrettyPrintCommand(cmd),
	)

	st := NewExecutionState()
	err = cmd.Execute(ctx, eCfg, st)
	if err != nil {
		return nil, err
	}

	res := NewQueryResult(parsedStmt, nil, nil, nil, COMMAND)

	return res, nil
}

// EvaluateExplain algebrizes, optimizes, and translates a query, returning
// metadata about the generated plan instead of executing it.
func EvaluateExplain(ctx context.Context, qCfg *QueryConfig, stmt parser.Statement) (*QueryResult, error) {
	qCfg.lg.Infof(log.Admin, "generating explain plan")
	parsedStmt := stmt.Copy().(parser.Statement)

	_, ok := stmt.(parser.SelectStatement)
	if !ok {
		return nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet, "no explain plan support for this statement for now")
	}

	var plan PlanStage

	algebrized, err := AlgebrizeQuery(ctx, qCfg.aCfg, stmt)
	if err != nil {
		// We can't create a query plan, so we have to exit.
		return nil, err
	}
	plan = algebrized

	qCfg.aCfg.lg.Debugf(log.Admin, "explain query plan: \n%v", PrettyPrintPlan(plan))

	optimized, err := OptimizePlan(ctx, qCfg.oCfg, plan)
	if err != nil {
		return nil, err
	}
	plan = optimized

	pushedDown, err := PushdownPlan(ctx, qCfg.pCfg, plan)
	if err != nil && !IsNonFatalPushdownError(err) {
		return nil, err
	}
	plan = pushedDown
	pde, ok := err.(PushdownError)
	var pushdownFailures map[PlanStage][]PushdownFailure
	if err != nil && ok {
		pushdownFailures = pde.Failures()
	}

	explainPlan := NewExplainStage(plan, qCfg.eCfg, pushdownFailures)

	stats, err := getPlanStats(plan, pushdownFailures)
	if err != nil {
		return nil, err
	}

	res := NewQueryResult(parsedStmt, explainPlan.Columns(), explainPlan, stats, EXPLAIN)

	return res, nil
}

// EvaluateShow algebrizes and executes a show statement.
func EvaluateShow(ctx context.Context, qCfg *QueryConfig, stmt *parser.Show) (*QueryResult, error) {
	switch strings.ToLower(stmt.Section) {
	case "charset", "collation", "columns", "create database", "create table",
		"databases", "index", "indexes", "keys", "processlist", "schemas", "status", "tables",
		"variables":
		return EvaluateQuery(ctx, qCfg, stmt)
	default:
		return evaluateShowNotImplemented(stmt)
	}
}

func evaluateShowNotImplemented(stmt *parser.Show) (*QueryResult, error) {
	return NewQueryResult(stmt, nil, nil, nil, SHOWNOTIMPL), nil
}

// EvaluateQuery algebrizes, optimizes, translates, and executes a query
// according to the provided configuration structs.
func EvaluateQuery(ctx context.Context, qCfg *QueryConfig, stmt parser.Statement) (*QueryResult, error) {

	parsedStmt := stmt.Copy().(parser.Statement)
	var plan PlanStage

	// Step 1: Algebrize
	algebrized, err := AlgebrizeQuery(ctx, qCfg.aCfg, stmt)

	if err = procutil.CheckForContextCancellationAndError(ctx, err); err != nil {
		return nil, err
	}

	plan = algebrized

	// Step 2: Optimize
	optimized, err := OptimizePlan(ctx, qCfg.oCfg, plan)
	if err = procutil.CheckForContextCancellationAndError(ctx, err); err != nil {
		return nil, err
	}

	plan = optimized

	// Step 3: Push Down
	pushedDown, err := PushdownPlan(ctx, qCfg.pCfg, plan)
	err = procutil.CheckForContextCancellationAndError(ctx, err)
	if err != nil && !IsNonFatalPushdownError(err) {
		return nil, err
	}

	plan = pushedDown
	pde, ok := err.(PushdownError)
	var pushdownFailures map[PlanStage][]PushdownFailure
	if err != nil && ok {
		pushdownFailures = pde.Failures()
	}

	// Step 4: Gather query plan statistics
	stats, err := getPlanStats(plan, pushdownFailures)
	if err != nil {
		return nil, err
	}

	// Step 5: Check if this query should be executed
	maybeFastPlan, err := CheckPlanExecution(ctx, qCfg.eCfg, plan)
	if err = procutil.CheckForContextCancellationAndError(ctx, err); err != nil {
		return nil, err
	}

	var op QueryOp

	switch stmt.(type) {
	case parser.SelectStatement:
		op = QUERY
	case *parser.Show:
		op = SHOW
	default:
		op = UNKNOWN
	}

	res := NewQueryResult(parsedStmt, plan.Columns(), maybeFastPlan, stats, op)

	return res, nil
}
