package sqlproxy

import (
	"github.com/erh/mongo-sql-temp/evaluator"
	"github.com/mongodb/mongo-tools/common/log"
)

// Execute walks an operator and its children to returns results.
func Execute(ctx *evaluator.ExecutionCtx, operator evaluator.Operator) ([]string, [][]interface{}, error) {
	rows := make([][]interface{}, 0)

	log.Logf(log.DebugLow, "Executing plan: %#v", operator)

	row := &evaluator.Row{}

	if err := operator.Open(ctx); err != nil {
		return nil, nil, err
	}

	for operator.Next(row) {
		values := getRowValues(operator.OpFields(), row)

		rows = append(rows, values)

		row.Data = []evaluator.TableRow{}
	}

	if err := operator.Close(); err != nil {
		return nil, nil, err
	}

	if err := operator.Err(); err != nil {
		return nil, nil, err
	}

	// no headers are returned for empty sets
	if len(rows) == 0 {
		return nil, nil, nil
	}

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
