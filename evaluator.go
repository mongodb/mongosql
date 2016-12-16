package sqlproxy

import (
	"fmt"

	"github.com/10gen/sqlproxy/client"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/options"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"gopkg.in/mgo.v2"
)

type Evaluator struct {
	config  *schema.Schema
	session *mgo.Session
	options options.SqldOptions
}

func NewEvaluator(cfg *schema.Schema, opts options.SqldOptions) (*Evaluator, error) {
	sp, err := client.NewSqldSessionProvider(opts)
	if err != nil {
		return nil, err
	}

	session, err := sp.GetSession()
	if err != nil {
		return nil, err
	}

	bi, err := session.BuildInfo()
	if err != nil {
		return nil, fmt.Errorf("can't fetch build information: %v", err)
	}

	if !bi.VersionAtLeast(3, 2, 0) {
		return nil, fmt.Errorf("server version is %v but version >= 3.2.0 required", bi.Version)
	}

	return &Evaluator{cfg, session, opts}, nil
}

// Session returns a copy of the evaluator's session.
func (e *Evaluator) Session() *mgo.Session {
	return e.session.New()
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

	conn.Logger(log.OptimizerComponent).Logf(log.DebugHigh, "optimizing query plan: \n%v", evaluator.PrettyPrintPlan(plan))
	plan, err = evaluator.OptimizePlan(conn, plan)
	if err != nil {
		return nil, nil, err
	}

	executionCtx := evaluator.NewExecutionCtx(conn)

	conn.Logger(log.EvaluatorComponent).Logf(log.DebugHigh, "executing query plan: \n%v", evaluator.PrettyPrintPlan(plan))

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

	conn.Logger(log.OptimizerComponent).Logf(log.DebugHigh, "optimizing query plan: \n%v", evaluator.PrettyPrintCommand(stmt))

	command, err := evaluator.OptimizeCommand(conn, stmt)
	if err != nil {
		return nil, err
	}

	conn.Logger(log.EvaluatorComponent).Logf(log.DebugHigh, "executing query plan: \n%v", evaluator.PrettyPrintCommand(stmt))

	executionCtx := evaluator.NewExecutionCtx(conn)

	return command.Execute(executionCtx), nil
}
