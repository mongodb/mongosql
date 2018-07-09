package evaluator

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/10gen/sqlproxy/collation"
)

const (
	stageID         = "stage_id"
	planStage       = "plan_stage"
	planColumns     = "plan_columns"
	sources         = "sources"
	databases       = "databases"
	tables          = "tables"
	aliases         = "aliases"
	collections     = "collections"
	pipeline        = "pipeline"
	pipelineExplain = "pipeline_explain"
	comment         = "comment"
)

// ExplainPlanStage is a stage containing information on the explain plan table of a query.
type ExplainPlanStage struct {
	plan    PlanStage
	columns []*Column
}

// ExplainPlanIter is an iterator that will iterate through the rows
// of the explain plan table.
type ExplainPlanIter struct {
	ctx        *ExecutionCtx
	rows       []*Row
	currentRow int
}

// NewExplainPlanStage creates a new ExplainPlanStage
// given a PlanStage and the generated columns for the table.
func NewExplainPlanStage(plan PlanStage, conn ConnectionCtx) ExplainPlanStage {
	return ExplainPlanStage{
		plan:    plan,
		columns: generateColumns(conn),
	}
}

// Open creates a visitor that will walk through the explain plan
// and return an iterator with the rows for the table.
func (ep ExplainPlanStage) Open(ctx *ExecutionCtx) (*ExplainPlanIter, error) {

	visitor := explainVisitor{
		ctx:            ctx,
		columns:        ep.columns,
		rows:           []*Row{},
		currentStageID: 0,
		sourceNodes:    []string{},
	}

	_, err := visitor.visit(ep.plan)
	if err != nil {
		return nil, err
	}

	iter := &ExplainPlanIter{
		ctx:        ctx,
		rows:       visitor.rows,
		currentRow: 0,
	}

	iter.sortRows()
	return iter, nil
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (ep ExplainPlanStage) Columns() []*Column {
	return ep.columns
}

// Collation returns the collation to use for comparisons.
func (ep ExplainPlanStage) Collation() *collation.Collation {
	return ep.plan.Collation()
}

// Next will pass the next row's data to the row pointer.
func (ei *ExplainPlanIter) Next(row *Row) bool {
	if ei.currentRow < len(ei.rows) {
		row.Data = ei.rows[ei.currentRow].Data
		ei.currentRow++

		return true
	}
	return false
}

func (ei *ExplainPlanIter) sortRows() {
	sort.Slice(ei.rows, func(i, j int) bool {
		return Int64(ei.rows[i].Data[0].Data) < Int64(ei.rows[j].Data[0].Data)
	})
}

// Close will close the iterator.
func (ei *ExplainPlanIter) Close() error {
	return nil
}

// Err will return the error for the iterator.
func (ei *ExplainPlanIter) Err() error {
	return nil
}

// generateColumns will generate the columns for the explain plan table.
func generateColumns(ctx ConnectionCtx) []*Column {

	var columns []*Column

	tableName := "explain_plan"
	dbName := ctx.DB()
	colNames := []string{stageID, planStage, planColumns, sources, databases,
		tables, aliases, collections, pipeline, pipelineExplain, comment}

	for i := 0; i < len(colNames); i++ {
		selectID := i + 1

		switch colNames[i] {
		case stageID:
			columns = append(columns, NewColumn(
				selectID, tableName, "", dbName, colNames[i], "", "", EvalInt64, "int", true))
		case planStage, planColumns, databases, comment, pipeline, pipelineExplain:
			columns = append(columns, NewColumn(
				selectID, tableName, "", dbName, colNames[i], "", "", EvalString, "string", false))
		case sources, tables, aliases, collections:
			columns = append(columns, NewColumn(
				selectID, tableName, "", dbName, colNames[i], "", "", EvalString, "array", false))
		}
	}

	return columns

}

// getPlanColumns returns a string representation of a stage's columns
// given the stage's columns.
func getPlanColumns(columns []*Column) string {

	b := bytes.NewBufferString("[")
	for i, c := range columns {
		if i != 0 {
			b.WriteString(", ")
		}
		if len(c.Database) != 0 {
			b.WriteString(fmt.Sprintf("{name: %v.%v.'%v', type: '%v'}",
				c.Database, c.Table, c.Name, EvalTypeToSQLType(c.EvalType)))
		} else {
			b.WriteString(fmt.Sprintf("{name: '%v', type: '%v'}",
				c.Name, EvalTypeToSQLType(c.EvalType)))
		}
	}
	b.WriteString("]")
	return b.String()
}
