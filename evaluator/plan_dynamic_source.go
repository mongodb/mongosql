package evaluator

import (
	"context"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/schema"
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
		aliasName = string(table.Name())
	}

	return &DynamicSourceStage{
		selectID:  selectID,
		table:     table,
		dbName:    string(db.Name()),
		aliasName: aliasName,
	}
}

// Columns gets the columns that will be projected from the stage.
func (s *DynamicSourceStage) Columns() []*Column {
	var columns []*Column
	for _, c := range s.table.Columns() {
		column := NewColumn(s.selectID,
			s.aliasName,
			string(s.table.Name()),
			s.dbName,
			string(c.Name()),
			string(c.Name()),
			"",
			SQLTypeToEvalType(schema.SQLType(c.Type())),
			schema.MongoNone,
			false,
		)
		columns = append(columns, column)
	}

	return columns
}

// Collation gets the collation the stage uses.
func (s *DynamicSourceStage) Collation() *collation.Collation {
	return collation.Default
}

// Open creates an Iter to loop over the results.
func (s *DynamicSourceStage) Open(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) (Iter, error) {
	reader, err := s.table.OpenReader()
	if err != nil {
		return nil, err
	}

	i := &dynamicDataSourceIter{
		cfg:       cfg,
		selectID:  s.selectID,
		dbName:    s.dbName,
		tableName: s.aliasName,
		columns:   s.table.Columns(),
		reader:    reader,
	}
	return i, nil
}

type dynamicDataSourceIter struct {
	cfg       *ExecutionConfig
	selectID  int
	dbName    string
	tableName string
	columns   []catalog.Column
	reader    catalog.DataReader

	dataRow catalog.DataRow
	err     error
}

func (i *dynamicDataSourceIter) Next(ctx context.Context, row *Row) bool {
	if i.err != nil {
		return false
	}

	found, err := i.reader.Next(&i.dataRow)
	if err != nil {
		i.err = err
		return false
	}
	if !found {
		return false
	}

	row.Data = Values{}
	for x := 0; x < len(i.dataRow.Values); x++ {
		sqlValue := GoValueToSQLValue(i.cfg.sqlValueKind, i.dataRow.Values[x])
		converted := ConvertTo(sqlValue, SQLTypeToEvalType(schema.SQLType(i.columns[x].Type())))
		row.Data = append(row.Data, NewValue(
			i.selectID,
			i.dbName,
			i.tableName,
			string(i.columns[x].Name()),
			converted))
	}

	i.err = i.cfg.memoryMonitor.Acquire(row.Data.Size())
	return i.err == nil
}

func (i *dynamicDataSourceIter) Close() error {
	return i.reader.Close()
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
