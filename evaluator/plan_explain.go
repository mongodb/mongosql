package evaluator

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
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

// ExplainStage is a stage containing information on the explain plan table of a query.
type ExplainStage struct {
	plan    PlanStage
	columns []*Column
}

// ExplainIter is an iterator that will iterate through the rows
// of the explain plan table.
type ExplainIter struct {
	ctx        *ExecutionCtx
	rows       []*Row
	currentRow int
}

// NewExplainStage creates a new ExplainStage
// given a PlanStage and the generated columns for the table.
func NewExplainStage(plan PlanStage, conn ConnectionCtx) ExplainStage {
	return ExplainStage{
		plan:    plan,
		columns: generateColumns(conn),
	}
}

// Open creates a visitor that will walk through the explain plan
// and return an iterator with the rows for the table.
func (es ExplainStage) Open(ctx *ExecutionCtx) (*ExplainIter, error) {

	visitor := explainVisitor{
		ctx:            ctx,
		columns:        es.columns,
		rows:           []*Row{},
		currentStageID: 0,
		sourceNodes:    []string{},
	}

	evalCtx := NewEvalCtx(ctx, ctx.Variables().GetCollation(variable.CollationConnection))
	_, e := optimizePushDown(es.plan, evalCtx, ctx.Logger(log.OptimizerComponent))

	pde, ok := e.(*pushDownError)
	if e != nil && ok {
		visitor.pushDownErrors = pde.errors
	}

	_, err := visitor.visit(es.plan)
	if err != nil {
		return nil, err
	}

	iter := &ExplainIter{
		ctx:        ctx,
		rows:       visitor.rows,
		currentRow: 0,
	}

	iter.sortRows()
	return iter, nil
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (es ExplainStage) Columns() []*Column {
	return es.columns
}

// Collation returns the collation to use for comparisons.
func (es ExplainStage) Collation() *collation.Collation {
	return es.plan.Collation()
}

// Next will pass the next row's data to the row pointer.
func (ei *ExplainIter) Next(row *Row) bool {
	if ei.currentRow < len(ei.rows) {
		row.Data = ei.rows[ei.currentRow].Data
		ei.currentRow++

		return true
	}
	return false
}

func (ei *ExplainIter) sortRows() {
	sort.Slice(ei.rows, func(i, j int) bool {
		return Int64(ei.rows[i].Data[0].Data) < Int64(ei.rows[j].Data[0].Data)
	})
}

// Close will close the iterator.
func (ei *ExplainIter) Close() error {
	return nil
}

// Err will return the error for the iterator.
func (ei *ExplainIter) Err() error {
	return nil
}

// generateColumns will generate the columns for the explain plan table.
func generateColumns(ctx ConnectionCtx) []*Column {

	var columns []*Column

	tableName := "explain"
	dbName := ctx.DB()
	colNames := []string{stageID, planStage, planColumns, sources, databases,
		tables, aliases, collections, pipeline, pipelineExplain, comment}

	for i := 0; i < len(colNames); i++ {
		switch colNames[i] {
		case stageID:
			columns = append(columns,
				NewColumn(i, tableName,
					tableName, dbName, colNames[i],
					colNames[i], "", EvalInt64,
					schema.MongoInt64, false))
		default:
			columns = append(columns,
				NewColumn(i, tableName,
					tableName, dbName, colNames[i],
					colNames[i], "", EvalString,
					schema.MongoString, false))
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
		if len(c.Database) == 0 || len(c.Table) == 0 {
			b.WriteString(fmt.Sprintf("{name: '%v', type: '%v'}",
				c.Name, EvalTypeToSQLType(c.EvalType)))
		} else {
			b.WriteString(fmt.Sprintf("{name: %v.%v.'%v', type: '%v'}",
				c.Database, c.Table, c.Name, EvalTypeToSQLType(c.EvalType)))
		}
	}
	b.WriteString("]")
	return b.String()
}
