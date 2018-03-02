package evaluator

import (
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/parser"
)

// EvaluateQuery creates an iterator in order to stream results.
func EvaluateQuery(sql string, ast parser.Statement,
	conn ConnectionCtx) ([]*Column, ErrCloser, error) {

	lgr := conn.Logger(log.AlgebrizerComponent)

	switch ast.(type) {
	case parser.SelectStatement:
		lgr.Infof(log.Admin, `generating query plan for sql: "%v"`, sql)
	case *parser.Show:
		lgr.Infof(log.Admin, `generating query plan for show statement: "%v"`, sql)
	default:
		// Should never happen
		lgr.Warnf(log.Admin, `generating query plan for unknown statement: "%v"`, sql)
	}

	plan, err := AlgebrizeQuery(ast, conn.DB(), conn.Variables(), conn.Catalog())
	if err != nil {
		return nil, nil, err
	}

	conn.Logger(log.OptimizerComponent).Debugf(log.Dev,
		"optimizing query plan: \n%v",
		PrettyPrintPlan(plan))

	plan = OptimizePlan(conn, plan)
	executionCtx := NewExecutionCtx(conn)

	var fastIter FastIter
	var iter Iter

	// In the case of full pushdown (which we know we have achieved because
	// the plan is a MongoSourceStage), we can bypass a lot of in-memory
	// work, and thus optimize data streaming.
	if fastPlan, ok := plan.(*MongoSourceStage); ok {
		fastIter, err = fastPlan.FastOpen(executionCtx)
	} else {
		iter, err = plan.Open(executionCtx)
	}

	if err != nil {
		return nil, nil, err
	}

	conn.Logger(log.EvaluatorComponent).Debugf(log.Admin,
		"executing query plan: \n%v",
		PrettyPrintPlan(plan))

	columns := plan.Columns()

	if fastIter != nil {
		return columns, fastIter, nil
	}
	return columns, iter, nil
}

// EvaluateCommand creates an executor in which to execute the command.
func EvaluateCommand(ast parser.Statement, conn ConnectionCtx) (Executor, error) {

	conn.Logger(log.AlgebrizerComponent).Infof(log.Admin,
		`generating query plan: "%v"`,
		parser.String(ast))

	stmt, err := AlgebrizeCommand(ast, conn.DB(), conn.Variables(), conn.Catalog())
	if err != nil {
		return nil, err
	}

	conn.Logger(log.OptimizerComponent).Debugf(log.Dev,
		"optimizing query plan: \n%v",
		PrettyPrintCommand(stmt))

	command := OptimizeCommand(conn, stmt)
	executionCtx := NewExecutionCtx(conn)

	conn.Logger(log.EvaluatorComponent).Debugf(log.Admin,
		"executing query plan: \n%v",
		PrettyPrintCommand(stmt))

	return command.Execute(executionCtx), nil
}
