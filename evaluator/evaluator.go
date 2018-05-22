package evaluator

import (
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/variable"
)

const mongoPrimaryKey string = "_id"

// getFastPlanStage returns a FastPlanStage and true if possible,
// otherwise nil and false.
func getFastPlanStage(plan PlanStage) (FastPlanStage, bool) {
	if fastPlan, ok := plan.(*MongoSourceStage); ok {
		return fastPlan, true
	} else if projectPlan, ok := plan.(*ProjectStage); ok {
		if unionPlan, ok := projectPlan.source.(*UnionStage); ok {
			// Only UNION ALL can be FastOpen'd safely.
			if unionPlan.kind != UnionAll {
				return nil, false
			}
			if left, ok := getFastPlanStage(unionPlan.left); ok {
				if right, ok := getFastPlanStage(unionPlan.right); ok {
					// Note that we remove the project stages, which means
					// we need to create a new stage here just in case we
					// ultimately end up not able to generate a complete
					// FastPlanStage. If we modified the plan in place,
					// such a situation would result in an unusable plan.
					return &UnionStage{
						left:  left,
						right: right,
						kind:  UnionAll,
					}, true
				}
			}
		}
	}
	return nil, false
}

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

	if conn.Variables().GetBool(variable.MongosqldFullPushdownExecMode) {
		// Don't attempt query execution if the plan isn't fully pushed down.
		if err = IsFullyPushedDown(plan); err != nil {
			return nil, nil, err
		}
	}

	executionCtx := NewExecutionCtx(conn)

	var fastIter FastIter
	var iter Iter

	columns := plan.Columns()

	// If we can FastOpen the plan, do so.
	if fastPlan, ok := getFastPlanStage(plan); ok {
		fastIter, err = fastPlan.FastOpen(executionCtx)
		if err != nil {
			return nil, nil, err
		}
		conn.Logger(log.EvaluatorComponent).Debugf(log.Admin,
			"executing query plan with fast iterator: \n%v", PrettyPrintPlan(fastPlan))
		return columns, fastIter, nil
	}
	conn.Logger(log.EvaluatorComponent).Debugf(log.Admin,
		"executing query plan: \n%v",
		PrettyPrintPlan(plan))

	iter, err = plan.Open(executionCtx)
	if err != nil {
		return nil, nil, err
	}

	iter = &memoryIter{
		ctx:    executionCtx,
		plan:   plan,
		source: iter,
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

	executionCtx := NewExecutionCtx(conn)

	if executionCtx.Variables().MongoDBInfo.IsSecurityEnabled() {
		// Make sure this command is authorized.
		err = stmt.Authorize(executionCtx)
		// If it is not authorized, report to the user why.
		if err != nil {
			return nil, err
		}
	}

	conn.Logger(log.OptimizerComponent).Debugf(log.Dev,
		"optimizing query plan: \n%v",
		PrettyPrintCommand(stmt))

	command := OptimizeCommand(conn, stmt)

	conn.Logger(log.EvaluatorComponent).Debugf(log.Admin,
		"executing query plan: \n%v",
		PrettyPrintCommand(stmt))

	return command.Execute(executionCtx), nil
}
