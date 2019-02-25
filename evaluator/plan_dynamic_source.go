package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/results"
)

// DynamicSourceStage handles reading data from a catalog.DynamicTable.
type DynamicSourceStage struct {
	selectID  int
	table     *catalog.DynamicTable
	aliasName string
	dbName    string
}

// Children returns a slice of all the Node children of the Node.
func (DynamicSourceStage) Children() []Node {
	return []Node{}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (DynamicSourceStage) ReplaceChild(i int, n Node) {
	panicWithInvalidIndex("DynamicSourceStage", i, -1)
}

// NewDynamicSourceStage creates a new DynamicSourceStage.
func NewDynamicSourceStage(db catalog.Database, table *catalog.DynamicTable,
	selectID int, aliasName string) *DynamicSourceStage {
	if aliasName == "" {
		aliasName = table.Name()
	}

	return &DynamicSourceStage{
		selectID:  selectID,
		table:     table,
		dbName:    string(db.Name()),
		aliasName: aliasName,
	}
}

// Columns gets the columns that will be projected from the stage.
func (s *DynamicSourceStage) Columns() []*results.Column {
	return s.table.Columns()
}

// Collation gets the collation the stage uses.
func (s *DynamicSourceStage) Collation() *collation.Collation {
	return collation.Default
}

// Open creates an Iter to loop over the results.
func (s *DynamicSourceStage) Open(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (RowIter, error) {
	i := &dynamicDataSourceIter{
		cfg:        cfg,
		currentRow: 0,
		selectID:   s.selectID,
		dbName:     s.dbName,
		tableName:  s.aliasName,
		columns:    s.table.Columns(),
		rows:       s.table.Rows(),
	}
	return i, nil
}

type dynamicDataSourceIter struct {
	cfg        *ExecutionConfig
	currentRow int
	selectID   int
	dbName     string
	tableName  string
	columns    results.Columns
	rows       results.Rows
	err        error
}

func (i *dynamicDataSourceIter) Next(ctx context.Context, row *results.Row) bool {
	if i.err != nil {
		return false
	}
	if i.currentRow == len(i.rows) {
		return false
	}
	*row = i.rows[i.currentRow]
	i.currentRow++

	i.err = i.cfg.memoryMonitor.Acquire(row.Data.Size())
	return i.err == nil
}

func (i *dynamicDataSourceIter) Close() error {
	return nil
}

func (i *dynamicDataSourceIter) Err() error {
	return i.err
}

func (s *DynamicSourceStage) clone() PlanStage {
	return &DynamicSourceStage{
		selectID:  s.selectID,
		table:     s.table,
		aliasName: s.aliasName,
		dbName:    s.dbName,
	}
}
