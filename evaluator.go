package sqlproxy

//go:generate go run testdata/generate.go

import (
	"context"

	"github.com/10gen/sqlproxy/client"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
)

type Evaluator struct {
	config          *schema.Schema
	sessionProvider *client.SessionProvider
	options         options.SqldOptions
}

func NewEvaluator(cfg *schema.Schema, opts options.SqldOptions) (*Evaluator, error) {
	sp, err := client.NewSqldSessionProvider(opts)
	if err != nil {
		return nil, err
	}

	// on proxy startup attempt to get a session - to validate
	// the server version
	session, err := sp.GetSession(context.Background())
	if err != nil {
		return nil, err
	}

	session.Close()

	return &Evaluator{cfg, sp, opts}, nil
}

// Session returns a new MongoDB session.
func (e *Evaluator) Session(ctx context.Context) (*mongodb.Session, error) {
	return e.sessionProvider.GetSession(ctx)
}

// Schema returns a copy of the evaluator's schema.
func (e *Evaluator) Schema() schema.Schema {
	return *e.config
}

// EvaluateQuery creates an iterator in order to stream results.
func (e *Evaluator) EvaluateQuery(ast parser.Statement, conn evaluator.ConnectionCtx) ([]*evaluator.Column, evaluator.Iter, error) {
	switch ast.(type) {
	case parser.SelectStatement:
		conn.Logger(log.AlgebrizerComponent).Logf(log.Info, `generating query plan for parsed sql: "%v"`, parser.String(ast))
	}

	plan, err := evaluator.AlgebrizeQuery(ast, conn.DB(), conn.Variables(), conn.Catalog())
	if err != nil {
		return nil, nil, err
	}

	conn.Logger(log.OptimizerComponent).Logf(log.DebugLow, "optimizing query plan: \n%v", evaluator.PrettyPrintPlan(plan))

	plan = evaluator.OptimizePlan(conn, plan)
	executionCtx := evaluator.NewExecutionCtx(conn)

	conn.Logger(log.EvaluatorComponent).Logf(log.DebugLow, "executing query plan: \n%v", evaluator.PrettyPrintPlan(plan))

	iter, err := plan.Open(executionCtx)
	if err != nil {
		return nil, nil, err
	}

	columns := plan.Columns()

	return columns, iter, nil
}

// EvaluateCommand creates an executor in which to execute the command.
func (e *Evaluator) EvaluateCommand(ast parser.Statement, conn evaluator.ConnectionCtx) (evaluator.Executor, error) {

	conn.Logger(log.AlgebrizerComponent).Logf(log.Info, `generating query plan: "%v"`, parser.String(ast))

	stmt, err := evaluator.AlgebrizeCommand(ast, conn.DB(), conn.Variables(), conn.Catalog())
	if err != nil {
		return nil, err
	}

	conn.Logger(log.OptimizerComponent).Logf(log.DebugLow, "optimizing query plan: \n%v", evaluator.PrettyPrintCommand(stmt))

	command := evaluator.OptimizeCommand(conn, stmt)
	executionCtx := evaluator.NewExecutionCtx(conn)

	conn.Logger(log.EvaluatorComponent).Logf(log.DebugLow, "executing query plan: \n%v", evaluator.PrettyPrintCommand(stmt))

	return command.Execute(executionCtx), nil
}
