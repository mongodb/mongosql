package sqlproxy

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/evaluator"
	"github.com/mongodb/mongo-tools/common/log"
	"gopkg.in/mgo.v2"
)

type Evaluator struct {
	cfg           *config.Config
	globalSession *mgo.Session
}

func NewEvaluator(cfg *config.Config) (*Evaluator, error) {
	e := new(Evaluator)
	e.cfg = cfg

	session, err := mgo.Dial(cfg.Url)
	if err != nil {
		return nil, err
	}
	e.globalSession = session

	return e, nil
}

func (e *Evaluator) getSession() *mgo.Session {
	if e.globalSession == nil {
		panic("No global session has been set")
	}
	return e.globalSession.Copy()
}

// EvalSelect returns all rows matching the query.
func (e *Evaluator) EvalSelect(db, sql string, stmt *sqlparser.Select, conn evaluator.ConnectionCtx) ([]string, [][]interface{}, error) {
	log.Logf(log.DebugLow, "Evaluating select: %#v", stmt)

	if stmt == nil {
		// we can parse ourselves
		raw, err := sqlparser.Parse(sql)
		if err != nil {
			return nil, nil, err
		}
		var ok bool
		if stmt, ok = raw.(*sqlparser.Select); !ok {
			return nil, nil, fmt.Errorf("got a non-select statement in EvalSelect")
		}
	}

	// create initial parse context
	pCtx, err := evaluator.NewParseCtx(stmt, e.cfg, db)
	if err != nil {
		return nil, nil, fmt.Errorf("error constructing new parse context: %v", err)
	}

	// resolve names
	if err = evaluator.AlgebrizeStatement(stmt, pCtx); err != nil {
		return nil, nil, fmt.Errorf("error algebrizing select statement: %v", err)
	}

	eCtx := &evaluator.ExecutionCtx{
		Config:        e.cfg,
		Db:            db,
		ConnectionCtx: conn,
	}

	// construct plan
	queryPlan, err := evaluator.PlanQuery(eCtx, stmt)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting query plan: %v", err)
	}

	// execute plan
	columns, results, err := Execute(eCtx, queryPlan)
	if err != nil {
		return nil, nil, fmt.Errorf("error executing query: %v", err)
	}

	return columns, results, nil
}
