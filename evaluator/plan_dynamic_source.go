package evaluator

import (
	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/schema"
)

// DynamicSourceStage handles reading data from a catalog.DynamicTable.
type DynamicSourceStage struct {
	selectID  int
	table     *catalog.DynamicTable
	aliasName string
	dbName    string
}

// NewDynamicSourceStage creates a new DynamicSourceStage.
func NewDynamicSourceStage(db *catalog.Database, table *catalog.DynamicTable,
	selectID int, aliasName string) *DynamicSourceStage {
	if aliasName == "" {
		aliasName = string(table.Name())
	}

	return &DynamicSourceStage{
		selectID:  selectID,
		table:     table,
		dbName:    string(db.Name),
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
			c.Type(), schema.MongoNone, false)
		columns = append(columns, column)
	}

	return columns
}

// Collation gets the collation the stage uses.
func (s *DynamicSourceStage) Collation() *collation.Collation {
	return collation.Default
}

// Open creates an Iter to loop over the results.
func (s *DynamicSourceStage) Open(ctx *ExecutionCtx) (Iter, error) {
	reader, err := s.table.OpenReader()
	if err != nil {
		return nil, err
	}

	i := &dynamicDataSourceIter{
		selectID:  s.selectID,
		dbName:    s.dbName,
		tableName: s.aliasName,
		columns:   s.table.Columns(),
		reader:    reader,
	}
	return i, nil
}

type dynamicDataSourceIter struct {
	selectID  int
	dbName    string
	tableName string
	columns   []catalog.Column
	reader    catalog.DataReader

	dataRow catalog.DataRow
	err     error
}

func (i *dynamicDataSourceIter) Next(row *Row) bool {
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
		sqlValue, _ := NewSQLValue(i.dataRow.Values[x], i.columns[x].Type(), "")
		row.Data = append(row.Data, NewValue(
			i.selectID,
			i.dbName,
			i.tableName,
			string(i.columns[x].Name()),
			sqlValue))
	}

	return true
}

func (i *dynamicDataSourceIter) Close() error {
	return i.reader.Close()
}

func (i *dynamicDataSourceIter) Err() error {
	return i.err
}
