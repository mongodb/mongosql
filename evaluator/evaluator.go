package evaluator

import (
	"context"
	"strings"

	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/parser"
)

const mongoPrimaryKey string = "_id"

// Constant values for QueryOp returned by evaluator
const (
	COMMAND     QueryOp = iota // commands: AlterTable, DropTable, Flush, Kill, RenameTable, Set, Use
	QUERY                      // queries: Select, SimpleSelect, Union
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
	Stmt    parser.Statement // allows statement execution tracking in server
	Columns []*Column
	Iter    ErrCloser
	Stats   *PlanStats
	Op      QueryOp
}

// NewQueryResult is a constructor for a QueryResult.
func NewQueryResult(stmt parser.Statement, columns []*Column, iter ErrCloser, stats *PlanStats, op QueryOp) *QueryResult {
	return &QueryResult{
		Stmt:    stmt,
		Columns: columns,
		Iter:    iter,
		Stats:   stats,
		Op:      op,
	}
}

// PlanStats contains some statistics about a query plan.
type PlanStats struct {
	FullyPushedDown bool
	Explain         []*ExplainRecord
}

// EvaluateCommand runs a command, returning a QueryResult containing Op=COMMAND
// so that server knows to writeOK to client or any error encountered during execution.
func EvaluateCommand(ctx context.Context, rCfg *RewriterConfig, aCfg *AlgebrizerConfig, eCfg *ExecutionConfig, stmt parser.Statement) (*QueryResult, error) {

	parsedStmt := stmt.Copy().(parser.Statement)

	rewritten, err := RewriteQuery(rCfg, stmt)
	if err = util.CheckForContextCancellationAndError(ctx, err); err != nil {
		return nil, err
	}

	cmd, err := AlgebrizeCommand(aCfg, rewritten)
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

	rewritten, err := RewriteQuery(qCfg.rCfg, stmt)
	if err = util.CheckForContextCancellationAndError(ctx, err); err != nil {
		return nil, err
	}

	_, ok := rewritten.(parser.SelectStatement)

	if !ok {
		return nil, mysqlerrors.Newf(
			mysqlerrors.ErNotSupportedYet,
			"no explain plan support for this statement for now",
		)
	}

	var plan PlanStage

	algebrized, err := AlgebrizeQuery(qCfg.aCfg, rewritten)
	if err != nil {
		// We can't create a query plan, so we have to exit.
		return nil, err
	}
	plan = algebrized

	qCfg.aCfg.lg.Debugf(log.Admin,
		"query plan: \n%v",
		PrettyPrintPlan(plan),
	)

	optimized, err := OptimizePlan(ctx, qCfg.oCfg, plan)
	if err != nil {
		return nil, err
	}
	plan = optimized

	pushedDown, err := PushdownPlan(qCfg.pCfg, plan)
	if err != nil && !IsNonFatalPushdownError(err) {
		return nil, err
	}
	plan = pushedDown

	st := NewExecutionState()
	explainPlan := NewExplainStage(plan, qCfg.eCfg)
	iter, err := explainPlan.Open(ctx, qCfg.pCfg, qCfg.eCfg, st)
	if err != nil {
		// couldn't get an iterator, so we have to exit
		return nil, err
	}

	res := NewQueryResult(parsedStmt, explainPlan.Columns(), iter, nil, EXPLAIN)

	return res, nil
}

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

	// Step 1: Perform any syntactic rewrites
	rewritten, err := RewriteQuery(qCfg.rCfg, stmt)
	if err = util.CheckForContextCancellationAndError(ctx, err); err != nil {
		return nil, err
	}

	// Step 2: Algebrize
	algebrized, err := AlgebrizeQuery(qCfg.aCfg, rewritten)

	if err = util.CheckForContextCancellationAndError(ctx, err); err != nil {
		return nil, err
	}

	plan = algebrized

	// Step 3: Optimize
	optimized, err := OptimizePlan(ctx, qCfg.oCfg, plan)
	if err = util.CheckForContextCancellationAndError(ctx, err); err != nil {
		return nil, err
	}

	plan = optimized

	// Step 4: Push Down
	pushedDown, err := PushdownPlan(qCfg.pCfg, plan)
	err = util.CheckForContextCancellationAndError(ctx, err)
	if err != nil && !IsNonFatalPushdownError(err) {
		return nil, err
	}

	plan = pushedDown

	// Step 5: Gather query plan statistics
	stats, err := getPlanStats(plan, qCfg.pCfg)
	if err != nil {
		return nil, err
	}

	// Step 6: Execute
	iter, err := ExecutePlan(ctx, qCfg.eCfg, plan)
	if err = util.CheckForContextCancellationAndError(ctx, err); err != nil {
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

	res := NewQueryResult(parsedStmt, plan.Columns(), iter, stats, op)

	return res, nil
}
