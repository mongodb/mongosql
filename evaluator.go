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
func (e *Evaluator) EvalSelect(db, sql string, stmt sqlparser.SelectStatement, conn evaluator.ConnectionCtx) ([]string, [][]interface{}, error) {

	if stmt == nil {
		// we can parse ourselves
		raw, err := sqlparser.Parse(sql)
		if err != nil {
			return nil, nil, err
		}
		var ok bool
		if stmt, ok = raw.(sqlparser.SelectStatement); !ok {
			return nil, nil, fmt.Errorf("got a non-select statement in EvalSelect")
		}
	}

	log.Logf(log.DebugLow, "Preparing select query: %#v", sqlparser.String(stmt))

	if _, ok := stmt.(*sqlparser.Select); ok {

		// create initial parse context
		pCtx, err := evaluator.NewParseCtx(stmt, e.cfg, db)
		if err != nil {
			return nil, nil, fmt.Errorf("error constructing new parse context: %v", err)
		}

		// resolve names
		if err = evaluator.AlgebrizeStatement(stmt, pCtx); err != nil {
			return nil, nil, fmt.Errorf("error algebrizing select statement: %v", err)
		}
	}

	eCtx := &evaluator.ExecutionCtx{
		Config:        e.cfg,
		Db:            db,
		ConnectionCtx: conn,
	}

	// construct select plan
	queryPlan, err := evaluator.PlanQuery(eCtx, stmt)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting query plan: %v", err)
	}

	// execute plan
	columns, results, err := executeQueryPlan(eCtx, queryPlan)
	if err != nil {
		return nil, nil, fmt.Errorf("error executing query: %v", err)
	}

	return columns, results, nil
}

// executeQueryPlan executes the query plan held in the operator by reading from it
// until it exhausts all results.
func executeQueryPlan(ctx *evaluator.ExecutionCtx, operator evaluator.Operator) ([]string, [][]interface{}, error) {
	rows := make([][]interface{}, 0)

	log.Logf(log.DebugLow, "Executing plan: %#v", operator)

	row := &evaluator.Row{}

	if err := operator.Open(ctx); err != nil {
		return nil, nil, fmt.Errorf("operator open: %v", err)
	}

	for operator.Next(row) {
		values := getRowValues(operator.OpFields(), row)

		rows = append(rows, values)

		row.Data = []evaluator.TableRow{}
	}

	if err := operator.Close(); err != nil {
		return nil, nil, fmt.Errorf("operator close: %v", err)
	}

	if err := operator.Err(); err != nil {
		return nil, nil, fmt.Errorf("operator err: %v", err)
	}

	log.Logf(log.DebugLow, "Done executing plan")

	// make sure all rows have same number of values
	for idx, row := range rows {
		for len(row) < len(operator.OpFields()) {
			row = append(row, nil)
		}
		rows[idx] = row
	}

	var headers []string

	for _, field := range operator.OpFields() {
		headers = append(headers, field.View)
	}

	return headers, rows, nil
}

func getRowValues(columns []*evaluator.Column, row *evaluator.Row) []interface{} {
	values := make([]interface{}, 0)

	for _, column := range columns {

		value, _ := row.GetField(column.Table, column.Name)
		values = append(values, value)
	}

	return values
}
