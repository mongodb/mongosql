package evaluator

import (
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/variable"
)

const mongoPrimaryKey string = "_id"

// getFastPlanStage returns a FastPlanStage and true if possible,
// otherwise nil and false. Also, remove any unncessary UnionDistincts,
// which are any UnionDistincts other another UnionDistinct.
// The parameter underDistinct tells us if we are below a UnionDistinct in
// the Plan, in which case all UnionDisticts should be replaced with UnionAll
// in order to improve performance: there is no reason to remove duplicates
// twice.
//
// is32 is true if the server versions is 3.2.x.
func getFastPlanStage(plan PlanStage, is32 bool, underDistinct bool) (FastPlanStage, bool) {
	if fastPlan, ok := plan.(*MongoSourceStage); ok {
		return fastPlan, true
	} else if projectPlan, ok := plan.(*ProjectStage); ok {
		if groupPlan, ok := projectPlan.source.(*GroupByStage); ok {
			if unionPlan, ok := groupPlan.source.(*UnionStage); ok {
				// The presence of a UnionDistinct under a GroupByStage
				// tells us the GroupByStage is just being used for uniqueness.
				// A GroupByStage above a UnionAll could have other uses.
				if unionPlan.kind == UnionDistinct {
					if left, ok := getFastPlanStage(unionPlan.left, is32, true); ok {
						if right, ok := getFastPlanStage(unionPlan.right, is32, true); ok {
							unionType := UnionDistinct
							localIs32 := is32
							if underDistinct {
								localIs32 = false
								unionType = UnionAll
							}
							// Note that we remove the project stages, which means
							// we need to create a new stage here just in case we
							// ultimately end up not able to generate a complete
							// FastPlanStage. If we modified the plan in place,
							// such a situation would result in an unusable plan.
							ret := NewUnionStage(unionType, left, right)
							ret.is32 = localIs32
							return ret, true
						}
					}
				}
			}
		} else if unionPlan, ok := projectPlan.source.(*UnionStage); ok {
			// A UnionDistinct should always be under a GroupByStage under
			// the way we currently generated plan stages, but this check
			// protects us against future changes.
			if unionPlan.kind != UnionAll {
				return nil, false
			}
			if left, ok := getFastPlanStage(unionPlan.left, is32, underDistinct); ok {
				if right, ok := getFastPlanStage(unionPlan.right, is32, underDistinct); ok {
					return NewUnionStage(UnionAll, left, right), true
				}
			}
		}
	}
	return nil, false
}

// EvaluateQuery creates an iterator in order to stream results.
func EvaluateQuery(sql string, ast parser.Statement,
	conn ConnectionCtx) ([]*Column, ErrCloser, error) {

	is32 := !conn.Variables().MongoDBInfo.VersionAtLeast(3, 4, 0)
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

	columns := plan.Columns()

	if fastPlan, ok := getFastPlanStage(plan, is32, false); ok {
		fastIter, err = fastPlan.FastOpen(executionCtx)
		if err != nil {
			return nil, nil, err
		}
		conn.Logger(log.EvaluatorComponent).Debugf(log.Admin,
			"executing query plan with fast iterator: \n%v", PrettyPrintPlan(fastPlan))
		return columns, fastIter, nil
	}

	if conn.Variables().GetBool(variable.MongosqldFullPushdownExecMode) {
		// Don't attempt query execution if the plan isn't fully pushed down
		// or a fast plan.
		if err = IsFullyPushedDown(plan); err != nil {
			return nil, nil, err
		}
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

// EvaluateExplain creates an iterator to stream the explain plan results table.
func EvaluateExplain(sql string, ast *parser.Explain,
	conn ConnectionCtx) ([]*Column, Iter, error) {

	conn.Logger(log.AlgebrizerComponent).Infof(log.Admin,
		`generating query plan for explain statement: "%v"`, sql)

	switch ast.Statement.(type) {
	case parser.SelectStatement:

		plan, err := AlgebrizeQuery(ast.Statement, conn.DB(), conn.Variables(), conn.Catalog())
		if err != nil {
			return nil, nil, err
		}
		conn.Logger(log.EvaluatorComponent).Debugf(log.Admin,
			"query plan: \n%v",
			PrettyPrintPlan(plan))

		plan = OptimizePlan(conn, plan)

		explainPlan := NewExplainStage(plan, conn)

		var iter Iter
		iter, err = explainPlan.Open(NewExecutionCtx(conn))
		if err != nil {
			return nil, nil, err
		}

		return explainPlan.Columns(), iter, err
	}

	return nil, nil, mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet,
		"no support for explain (%s) for now", sql)
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
