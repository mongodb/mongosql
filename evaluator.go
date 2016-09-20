package sqlproxy

import (
	"fmt"

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
	options options.Options
}

func NewEvaluator(cfg *schema.Schema, opts options.Options) (*Evaluator, error) {
	info, err := options.GetDialInfo(opts)
	if err != nil {
		return nil, err
	}

	session, err := mgo.DialWithInfo(info)
	if err != nil {
		return nil, fmt.Errorf("connecting to mongodb failed: %v", err.Error())
	}

	session.SetSocketTimeout(0)

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

// Evaluate executes the query and returns an iterator
// capable of going over all the generated results.
func (e *Evaluator) Evaluate(ast parser.SelectStatement, conn evaluator.ConnectionCtx) ([]*evaluator.Column, evaluator.Iter, error) {

	conn.Logger(log.AlgebrizerComponent).Logf(log.Info, `generating query plan: "%v"`, parser.String(ast))

	plan, err := evaluator.AlgebrizeSelect(ast, conn.DB(), e.config)
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

func (e *Evaluator) EvaluateCommand(ast parser.Statement, conn evaluator.ConnectionCtx) (evaluator.Executor, error) {

	conn.Logger(log.AlgebrizerComponent).Logf(log.Info, `generating query plan: "%v"`, parser.String(ast))

	stmt, err := evaluator.AlgebrizeCommand(ast, conn.DB(), e.config)
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
