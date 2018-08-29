package evaluator

import (
	"fmt"

	"github.com/10gen/sqlproxy/internal/collation"
	"github.com/10gen/sqlproxy/internal/memory"
)

// RowGeneratorStage generates empty rows based on a counter field from its source PlanStage.
type RowGeneratorStage struct {
	// source is the operator that provides the data to project.
	source PlanStage
	// rowCountColumn is the column that RowGeneratorStage needs to store
	// for lookup in Next function.
	rowCountColumn *Column
}

// NewRowGeneratorStage creates a new RowGeneratorStage.
func NewRowGeneratorStage(source PlanStage, rowCountColumn *Column) *RowGeneratorStage {
	return &RowGeneratorStage{
		source:         source,
		rowCountColumn: rowCountColumn,
	}
}

// Open gets the iterator of its source plan stage and returns a RowGeneratorIter.
func (rg *RowGeneratorStage) Open(ctx *ExecutionCtx) (Iter, error) {
	sourceIter, err := rg.source.Open(ctx)
	if err != nil {
		return nil, err
	}
	return &rowGeneratorIter{
		memoryMonitor:  ctx.MemoryMonitor(),
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
	memoryMonitor  *memory.Monitor
	rowCountColumn *Column
	source         Iter
	err            error
	currentRow     uint64
	totalRows      uint64
}

// Next calls its source's next function to get the number of rows to generate.
func (rgIter *rowGeneratorIter) Next(row *Row) bool {
	for rgIter.currentRow >= rgIter.totalRows {
		if !rgIter.source.Next(row) {
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
		rgIter.totalRows = Uint64(rowCountField)
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
