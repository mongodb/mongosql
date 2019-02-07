package evaluator

import (
	"context"
	"fmt"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/memory"
	"github.com/10gen/sqlproxy/evaluator/values"
)

// RowGeneratorStage generates empty rows based on a counter field from its source PlanStage.
type RowGeneratorStage struct {
	// source is the operator that provides the data to project.
	source PlanStage
	// rowCountColumn is the column that RowGeneratorStage needs to store
	// for lookup in Next function.
	rowCountColumn *Column
}

// Children returns a slice of all the Node children of the Node.
func (rg RowGeneratorStage) Children() []Node {
	return []Node{rg.source}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (rg *RowGeneratorStage) ReplaceChild(i int, n Node) {
	switch i {
	case 0:
		rg.source = panicIfNotPlanStage("RowGeneratorStage", n)
	default:
		panicWithInvalidIndex("RowGeneratorStage", i, 0)
	}
}

// NewRowGeneratorStage creates a new RowGeneratorStage.
func NewRowGeneratorStage(source PlanStage, rowCountColumn *Column) *RowGeneratorStage {
	return &RowGeneratorStage{
		source:         source,
		rowCountColumn: rowCountColumn,
	}
}

// Open gets the iterator of its source plan stage and returns a RowGeneratorIter.
func (rg *RowGeneratorStage) Open(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (RowIter, error) {
	sourceIter, err := rg.source.Open(ctx, cfg, st)
	if err != nil {
		return nil, err
	}
	return &rowGeneratorIter{
		memoryMonitor:  cfg.memoryMonitor,
		rowCountColumn: rg.rowCountColumn,
		source:         sourceIter,
		currentRow:     0,
		totalRows:      0,
	}, nil
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (rg *RowGeneratorStage) Columns() (columns []*Column) {
	return []*Column{}
}

// Collation returns the collation to use for comparisons.
func (rg *RowGeneratorStage) Collation() *collation.Collation {
	return rg.source.Collation()
}

// RowGeneratorIter is used to iterate over data that it is getting from its iterator.
type rowGeneratorIter struct {
	memoryMonitor  memory.Monitor
	rowCountColumn *Column
	source         RowIter
	err            error
	currentRow     uint64
	totalRows      uint64
}

// Next calls its source's next function to get the number of rows to generate.
func (rgIter *rowGeneratorIter) Next(ctx context.Context, row *Row) bool {
	for rgIter.currentRow >= rgIter.totalRows {
		if !rgIter.source.Next(ctx, row) {
			return false
		}

		rowCountField := row.Data.Map()[rgIter.rowCountColumn.MappingRegistryName]
		if rowCountField == nil {
			rgIter.err = fmt.Errorf("%v field is not present in source's row",
				rgIter.rowCountColumn.MappingRegistryName)
			return false
		}
		rgIter.err = rgIter.memoryMonitor.Release(row.Data.Size())
		if rgIter.err != nil {
			return false
		}
		rgIter.totalRows = values.Uint64(rowCountField)
		rgIter.currentRow = 0
	}

	row.Data = nil
	rgIter.currentRow++
	return true
}

func (rgIter *rowGeneratorIter) Close() error {
	return rgIter.source.Close()
}

func (rgIter *rowGeneratorIter) Err() error {
	if err := rgIter.source.Err(); err != nil {
		return err
	}
	return rgIter.err
}

func (rg *RowGeneratorStage) clone() PlanStage {
	return NewRowGeneratorStage(rg.source.clone(), rg.rowCountColumn.clone())
}
