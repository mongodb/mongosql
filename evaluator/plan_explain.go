package evaluator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/schema"
)

const (
	stageID         = "stage_id"
	planStage       = "plan_stage"
	planColumns     = "plan_columns"
	sources         = "sources"
	database        = "database"
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
	cfg        *ExecutionConfig
	rows       []*Row
	currentRow int
}

// NewExplainStage creates a new ExplainStage
// given a PlanStage and the generated columns for the table.
func NewExplainStage(plan PlanStage, cfg *ExecutionConfig) *ExplainStage {
	return &ExplainStage{
		plan:    plan,
		columns: generateExplainColumns(cfg.dbName),
	}
}

// Children returns a slice of all the Node children of the Node.
func (es ExplainStage) Children() []Node {
	return []Node{es.plan}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (es *ExplainStage) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		es.plan = panicIfNotPlanStage("ExplainStage", n)
	default:
		panicWithInvalidIndex("ExplainStage", i, 0)
	}
}

// Open creates a visitor that will walk through the explain plan
// and return an iterator with the rows for the table.
func (es *ExplainStage) Open(_ context.Context, pCfg *PushdownConfig, eCfg *ExecutionConfig, _ *ExecutionState) (Iter, error) {
	explainRecords, err := explainQuery(es.plan, pCfg)
	if err != nil {
		return nil, err
	}

	rows := []*Row{}
	for _, exp := range explainRecords {
		row := generateExplainRow(eCfg.sqlValueKind, es.Columns(), exp)
		rows = append(rows, row)
	}

	iter := &ExplainIter{
		cfg:        eCfg,
		rows:       rows,
		currentRow: 0,
	}

	return iter, nil
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (es *ExplainStage) Columns() []*Column {
	return es.columns
}

// Collation returns the collation to use for comparisons.
func (es *ExplainStage) Collation() *collation.Collation {
	return es.plan.Collation()
}

// Next will pass the next row's data to the row pointer.
func (ei *ExplainIter) Next(_ context.Context, row *Row) bool {
	if ei.currentRow < len(ei.rows) {
		row.Data = ei.rows[ei.currentRow].Data
		ei.currentRow++

		return true
	}
	return false
}

// Close will close the iterator.
func (ei *ExplainIter) Close() error {
	return nil
}

// Err will return the error for the iterator.
func (ei *ExplainIter) Err() error {
	return nil
}

// generateExplainColumns will generate the columns for the explain plan table.
func generateExplainColumns(dbName string) []*Column {

	var columns []*Column

	tableName := "explain"
	colNames := []string{stageID, planStage, planColumns, sources, database,
		tables, aliases, collections, pipeline, pipelineExplain, comment}

	for i := 0; i < len(colNames); i++ {
		switch colNames[i] {
		case stageID:
			columns = append(columns,
				NewColumn(i, tableName,
					tableName, dbName, colNames[i],
					colNames[i], "", types.EvalInt64,
					schema.MongoInt64, false))
		default:
			columns = append(columns,
				NewColumn(i, tableName,
					tableName, dbName, colNames[i],
					colNames[i], "", types.EvalString,
					schema.MongoString, false))
		}
	}

	return columns

}

// generateExplainRows generates the rows to be returned by an ExplainStage from
// the provided slice of stage explanations.
func generateExplainRow(kind values.SQLValueKind, cols []*Column, rec *ExplainRecord) *Row {
	var vs []Value
	for i := 0; i < len(cols); i++ {

		selectID := cols[i].SelectID
		dbName := cols[i].Database
		tableName := cols[i].Table
		name := cols[i].Name

		var value values.SQLValue

		switch name {
		case stageID:
			value = values.NewSQLInt64(kind, int64(rec.ID))
		case planStage:
			value = values.NewSQLVarchar(kind, rec.StageType)
		case planColumns:
			value = values.NewSQLVarchar(kind, rec.Columns)
		case sources:
			if len(rec.Sources) > 0 {
				strs := []string{}
				for _, i := range rec.Sources {
					strs = append(strs, strconv.Itoa(i))
				}
				result := fmt.Sprintf("[%v]", strings.Join(strs, ", "))
				value = values.NewSQLVarchar(kind, result)
			} else {
				value = values.NewSQLNull(kind)
			}
		case database:
			value = values.NewSQLVarcharFromOpt(kind, rec.Database)
		case tables:
			value = values.NewSQLVarcharFromOpt(kind, rec.Tables)
		case aliases:
			value = values.NewSQLVarcharFromOpt(kind, rec.Aliases)
		case collections:
			value = values.NewSQLVarcharFromOpt(kind, rec.Collections)
		case pipeline:
			value = values.NewSQLVarcharFromOpt(kind, rec.Pipeline)
		case pipelineExplain:
			value = values.NewSQLVarcharFromOpt(kind, rec.PipelineExplain)
		case comment:
			if len(rec.PushdownFailures) > 0 {
				val := struct {
					PushdownErrors []PushdownFailure `json:"pushdown_errors"`
				}{
					rec.PushdownFailures,
				}

				b, err := json.Marshal(val)
				if err != nil {
					panic(err)
				}

				value = values.NewSQLVarchar(kind, string(b))
			} else {
				value = values.NewSQLNull(kind)
			}
		default:
			panic(fmt.Sprintf("unexpected explain column %q", name))
		}

		vs = append(vs, NewValue(selectID, dbName, tableName, name, value))
	}

	return &Row{Data: vs}
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
				c.Name, types.EvalTypeToSQLType(c.EvalType)))
		} else {
			b.WriteString(fmt.Sprintf("{name: %v.%v.'%v', type: '%v'}",
				c.Database, c.Table, c.Name, types.EvalTypeToSQLType(c.EvalType)))
		}
	}
	b.WriteString("]")
	return b.String()
}
