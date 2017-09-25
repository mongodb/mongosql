package evaluator

import (
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/parser"
)

// EvaluateQuery creates an iterator in order to stream results.
func EvaluateQuery(sql string, ast parser.Statement, conn ConnectionCtx) ([]*Column, Iter, error) {
	switch ast.(type) {
	case parser.SelectStatement:
		conn.Logger(log.AlgebrizerComponent).Logf(log.Info, `generating query plan for sql: "%v"`, sql)
	case *parser.Show:
		conn.Logger(log.AlgebrizerComponent).Logf(log.Info, `generating query plan for show statement: "%v"`, sql)
	default:
		// Should never happen
		conn.Logger(log.AlgebrizerComponent).Errf(log.Info, `generating query plan for unknown statement: "%v"`, sql)
	}

	plan, err := AlgebrizeQuery(ast, conn.DB(), conn.Variables(), conn.Catalog())
	if err != nil {
		return nil, nil, err
	}

	conn.Logger(log.OptimizerComponent).Logf(log.DebugLow, "optimizing query plan: \n%v", PrettyPrintPlan(plan))

	plan = OptimizePlan(conn, plan)
	executionCtx := NewExecutionCtx(conn)

	conn.Logger(log.EvaluatorComponent).Logf(log.DebugLow, "executing query plan: \n%v", PrettyPrintPlan(plan))

	iter, err := plan.Open(executionCtx)
	if err != nil {
		return nil, nil, err
	}

	columns := plan.Columns()

	return columns, iter, nil
}

// EvaluateCommand creates an executor in which to execute the command.
func EvaluateCommand(ast parser.Statement, conn ConnectionCtx) (Executor, error) {

	conn.Logger(log.AlgebrizerComponent).Logf(log.Info, `generating query plan: "%v"`, parser.String(ast))

	stmt, err := AlgebrizeCommand(ast, conn.DB(), conn.Variables(), conn.Catalog())
	if err != nil {
		return nil, err
	}

	conn.Logger(log.OptimizerComponent).Logf(log.DebugLow, "optimizing query plan: \n%v", PrettyPrintCommand(stmt))

	command := OptimizeCommand(conn, stmt)
	executionCtx := NewExecutionCtx(conn)

	conn.Logger(log.EvaluatorComponent).Logf(log.DebugLow, "executing query plan: \n%v", PrettyPrintCommand(stmt))

	return command.Execute(executionCtx), nil
}
