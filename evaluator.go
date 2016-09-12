package sqlproxy

import (
	"fmt"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/mongodb/mongo-tools/common/log"
	"gopkg.in/mgo.v2"
)

type Evaluator struct {
	config  *schema.Schema
	session *mgo.Session
	options Options
}

func NewEvaluator(cfg *schema.Schema, opts Options) (*Evaluator, error) {
	info, err := GetDialInfo(opts)
	if err != nil {
		return nil, err
	}

	log.Logf(log.Always, "connecting to mongodb at %v", info.Addrs)

	session, err := mgo.DialWithInfo(info)
	if err != nil {
		log.Logf(log.Always, "connecting to mongodb failed")
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

	log.Logf(log.DebugHigh, "[conn%v] preparing query plan for: %#v", conn.ConnectionId(), parser.String(ast))

	plan, err := evaluator.AlgebrizeSelect(ast, conn.DB(), e.config)
	if err != nil {
		return nil, nil, err
	}

	log.Logf(log.DebugHigh, "[conn%v] optimizing plan: \n%v", conn.ConnectionId(), evaluator.PrettyPrintPlan(plan))
	plan, err = evaluator.OptimizePlan(conn, plan)
	if err != nil {
		return nil, nil, err
	}

	executionCtx := evaluator.NewExecutionCtx(conn)

	log.Logf(log.DebugHigh, "[conn%v] executing plan: \n%v", conn.ConnectionId(), evaluator.PrettyPrintPlan(plan))

	iter, err := plan.Open(executionCtx)
	if err != nil {
		return nil, nil, err
	}

	columns := plan.Columns()

	return columns, iter, nil
}

func (e *Evaluator) EvaluateCommand(ast parser.Statement, conn evaluator.ConnectionCtx) (evaluator.Executor, error) {
	log.Logf(log.DebugLow, "Preparing plan for: %#v", parser.String(ast))

	stmt, err := evaluator.AlgebrizeCommand(ast, conn.DB(), e.config)
	if err != nil {
		return nil, err
	}

	log.Logf(log.DebugHigh, "[conn%v] optimizing plan: \n%v", conn.ConnectionId(), evaluator.PrettyPrintCommand(stmt))

	command, err := evaluator.OptimizeCommand(conn, stmt)
	if err != nil {
		return nil, err
	}

	executionCtx := evaluator.NewExecutionCtx(conn)

	return command.Execute(executionCtx), nil

}
