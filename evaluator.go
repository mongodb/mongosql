package sqlproxy

import (
	"fmt"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
	"gopkg.in/mgo.v2"
)

type Evaluator struct {
	config  *schema.Schema
	session *mgo.Session
}

func NewEvaluator(cfg *schema.Schema) (*Evaluator, error) {

	session, err := mgo.Dial(cfg.Url)
	if err != nil {
		return nil, err
	}

	session.SetSocketTimeout(0)

	return &Evaluator{cfg, session}, nil
}

// Session returns a copy of the evaluator's session.
func (e *Evaluator) Session() *mgo.Session {
	return e.session.Copy()
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

	var pCtx *evaluator.ParseCtx
	var err error

	if _, ok := stmt.(*sqlparser.Select); ok {
		// create initial parse context
		pCtx, err = evaluator.NewParseCtx(stmt, e.config, db)
		if err != nil {
			return nil, nil, fmt.Errorf("error constructing new parse context: %v", err)
		}

		// resolve names
		if err = evaluator.AlgebrizeStatement(stmt, pCtx); err != nil {
			return nil, nil, fmt.Errorf("error algebrizing select statement: %v", err)
		}

		if pCtx != nil {
			log.Logf(log.DebugLow, "pCTX: %v\n", pCtx.String())
		}

	}

	// construct execution context
	eCtx := &evaluator.ExecutionCtx{
		Db:            db,
		ParseCtx:      pCtx,
		ConnectionCtx: conn,
		Schema:        e.config,
		Session:       e.session,
	}

	// construct query plan
	queryPlan, err := evaluator.PlanQuery(eCtx, stmt)
	if err != nil {
		return nil, nil, fmt.Errorf("error constructing query plan: %v", err)
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

	log.Logf(log.DebugLow, "Executing plan: %#v\n", operator)

	row := &evaluator.Row{}

	if err := operator.Open(ctx); err != nil {
		return nil, nil, fmt.Errorf("operator open: %v", err)
	}

	for operator.Next(row) {

		rows = append(rows, row.GetValues(operator.OpFields()))

		row.Data = []evaluator.TableRow{}
	}

	if err := operator.Close(); err != nil {
		return nil, nil, fmt.Errorf("operator close: %v", err)
	}

	if err := operator.Err(); err != nil {
		return nil, nil, fmt.Errorf("operator err: %v", err)
	}

	log.Log(log.DebugLow, "Done executing plan")

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
