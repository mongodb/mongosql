package evaluator

import (
	"fmt"
	"github.com/erh/mongo-sql-temp/translator/planner"
	"github.com/mongodb/mongo-tools/common/log"
)

// Execute walks an operator and its children to returns results.
func Execute(ctx *planner.ExecutionCtx, operator planner.Operator) ([]string, [][]interface{}, error) {
	rows := make([][]interface{}, 0)

	log.Logf(log.DebugLow, "Executing plan: %#v", operator)

	row := &planner.Row{}

	if err := operator.Open(ctx); err != nil {
		return nil, nil, err
	}
	defer operator.Close()

	s, ok := operator.(*planner.Select)
	if !ok {
		return nil, nil, fmt.Errorf("select operator must be root of query tree")
	}

	for operator.Next(row) {

		// TODO: merge eliot's work so we have
		// definite field names before Fields()
		// is called

		values := getRowValues(s.Namespaces, row)
		rows = append(rows, values)

		row.Data = []planner.TableRow{}
	}

	if err := operator.Err(); err != nil {
		return nil, nil, err
	}

	if len(rows) == 0 {
		return nil, nil, nil
	}

	// make sure all rows have same number of values
	for idx, row := range rows {
		for len(row) < len(s.Fields()) {
			row = append(row, nil)
		}
		rows[idx] = row
	}

	return s.Fields(), rows, nil
}

func getRowValues(n []*planner.Namespace, r *planner.Row) []interface{} {
	values := make([]interface{}, 0)

	for _, ns := range n {
		value, _ := r.GetField(ns.Table, ns.Column)
		values = append(values, value)
	}

	return values
}
