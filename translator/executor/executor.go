package executor

import (
	"github.com/erh/mongo-sql-temp/translator/planner"
	"github.com/erh/mongo-sql-temp/translator/types"
	"github.com/mongodb/mongo-tools/common/log"
)

// Execute walks an operator and its children to returns results.
func Execute(ctx *planner.ExecutionCtx, operator planner.Operator) ([]string, [][]interface{}, error) {
	rows := make([][]interface{}, 0)

	log.Logf(log.DebugLow, "Executing plan: %#v", operator)

	row := &types.Row{}

	if err := operator.Open(ctx); err != nil {
		return nil, nil, err
	}
	defer operator.Close()

	for operator.Next(row) {

		values := getRowValues(operator.OpFields(), row)

		rows = append(rows, values)

		row.Data = []types.TableRow{}
	}

	if err := operator.Err(); err != nil {
		return nil, nil, err
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

func getRowValues(columns []*planner.Column, row *types.Row) []interface{} {
	values := make([]interface{}, 0)

	for _, column := range columns {

		value, _ := row.GetField(column.Table, column.Name)
		values = append(values, value)
	}

	return values
}
