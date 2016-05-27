package sqlproxy

import (
	"fmt"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	"github.com/kr/pretty"
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

	log.Logf(log.Always, "connecting to mongodb at %v.", info.Addrs)

	session, err := mgo.DialWithInfo(info)
	if err != nil {
		log.Logf(log.Always, "connecting to mongodb failed.")
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

// EvaluateRows executes the query and returns the
// generated results as a slice of rows.
func (e *Evaluator) EvaluateRows(db, sql string, ast sqlparser.SelectStatement, conn evaluator.ConnectionCtx) ([]string, [][]interface{}, error) {

	columns, iter, err := e.Evaluate(db, sql, ast, conn)
	if err != nil {
		return nil, nil, err
	}

	rows := make([][]interface{}, 0)

	row := &evaluator.Row{}

	for iter.Next(row) {
		rows = append(rows, row.GetValues())
		row.Data = evaluator.Values{}
	}

	if err := iter.Close(); err != nil {
		return nil, nil, fmt.Errorf("iterator close: %v", err)
	}

	if err := iter.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterator err: %v", err)
	}

	// make sure all rows have same number of values
	for idx, row := range rows {
		for len(row) < len(columns) {
			row = append(row, nil)
		}
		rows[idx] = row
	}

	var headers []string

	for _, column := range columns {
		headers = append(headers, column.Name)
	}

	return headers, rows, nil
}

// Evaluate executes the query and returns an iterator
// capable of going over all the generated results.
func (e *Evaluator) Evaluate(db, sql string, ast sqlparser.SelectStatement, conn evaluator.ConnectionCtx) ([]*evaluator.Column, evaluator.Iter, error) {

	plan, executionCtx, err := e.Plan(db, sql, ast, conn)
	if err != nil {
		return nil, nil, err
	}

	log.Logf(log.DebugLow, "[conn%v] executing plan", conn.ConnectionId())

	iter, err := plan.Open(executionCtx)
	if err != nil {
		return nil, nil, err
	}

	columns := plan.Columns()

	return columns, iter, nil
}

// Plan returns a query plan for the SQL query.
func (e *Evaluator) Plan(db, sql string, ast sqlparser.SelectStatement, conn evaluator.ConnectionCtx) (evaluator.PlanStage, *evaluator.ExecutionCtx, error) {

	var ok bool

	if ast == nil {
		raw, err := sqlparser.Parse(sql)
		if err != nil {
			return nil, nil, err
		}

		ast, ok = raw.(sqlparser.SelectStatement)
		if !ok {
			return nil, nil, fmt.Errorf("got a non-select statement in algebrization")
		}
	}

	log.Logf(log.DebugLow, "Preparing query plan for: %#v", sqlparser.String(ast))

	plan, err := evaluator.Algebrize(ast, db, e.config)
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("\nBEFORE: %# v", pretty.Formatter(plan))

	plan, err = evaluator.OptimizePlan(plan)
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("\nAFTER: %# v", pretty.Formatter(plan))

	executionCtx := evaluator.NewExecutionCtx(conn)

	return plan, executionCtx, nil
}
