package catalog

import (
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

const informationSchema = "information_schema"

// NewDynamicTable creates a new DynamicTable.
func NewDynamicTable(name string, tableType string,
	generator func(tableName string) results.RowIter) *DynamicTable {
	columnMap := make(map[string]*results.Column)
	return &DynamicTable{
		currSelectID: 1,
		name:         name,
		tableType:    tableType,
		columnMap:    columnMap,
		generator:    generator,
	}
}

// DynamicTable is a table that returns its data dynamically.
type DynamicTable struct {
	currSelectID int
	name         string
	columns      results.Columns
	columnMap    map[string]*results.Column
	tableType    string
	generator    func(tableName string) results.RowIter
}

// Name returns the name for the DynamicTable, t.
func (t *DynamicTable) Name() string {
	return t.name
}

// Collation returns the collation for the DynamicTable, t.
func (t *DynamicTable) Collation() *collation.Collation {
	return collation.Default
}

// Column returns the column of the specified name.
func (t *DynamicTable) Column(name string) (*results.Column, error) {
	if c, ok := t.columnMap[strings.ToLower(name)]; ok {
		return c, nil
	}

	return nil, mysqlerrors.Defaultf(mysqlerrors.ErBadFieldError, name, t.Name())
}

// Columns returns the columns in the DynamicTable, t.
func (t *DynamicTable) Columns() results.Columns {
	return t.columns
}

// Comments are comments about the DynamicTable, t.
func (t *DynamicTable) Comments() string {
	return ""
}

// PrimaryKeys returns the primary keys for
// the DynamicTable, t.
func (t *DynamicTable) PrimaryKeys() results.Columns {
	return nil
}

// ForeignKeys returns the foreign keys for the DynamicTable, t.
func (t *DynamicTable) ForeignKeys() []ForeignKey {
	return nil
}

// Indexes returns nil for any DynamicTable, t.
func (t *DynamicTable) Indexes() []Index {
	return nil
}

// Type is the type of the DynamicTable, t.
func (t *DynamicTable) Type() string {
	return t.tableType
}

// AddColumn adds a column to the DynamicTable, t.
func (t *DynamicTable) AddColumn(tableName, name string, evalType types.EvalType) (*results.Column, error) {
	lowerName := strings.ToLower(name)
	if _, ok := t.columnMap[lowerName]; ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErDupFieldname, name)
	}

	cb := results.NewColumnBuilder()
	cb.SetColumnType(results.NewColumnType(evalType, schema.MongoNone))
	cb.SetSelectID(t.currSelectID)
	cb.SetTable(tableName)
	cb.SetOriginalTable(tableName)
	cb.SetDatabase(informationSchema)
	cb.SetName(name)
	cb.SetOriginalName(name)
	cb.SetMappingRegistryName("")
	cb.SetMongoName(name)
	cb.SetPrimaryKey(false)
	cb.SetComments("")
	cb.SetIsPolymorphic(false)
	cb.SetHasAlteredType(false)
	cb.SetNullable(true)
	c := cb.Build()
	t.currSelectID++

	t.columns = append(t.columns, c)
	t.columnMap[lowerName] = c

	return c, nil
}

// DynamicColumnDeclaration is a column name with an EvalType that declares a column.
type DynamicColumnDeclaration struct {
	columnName string
	evalType   types.EvalType
}

// NewDynamicColumnDeclaration creates a new DynamicColumnDeclaration.
func NewDynamicColumnDeclaration(columnName string, evalType types.EvalType) DynamicColumnDeclaration {
	return DynamicColumnDeclaration{
		columnName: columnName,
		evalType:   evalType,
	}
}

// AddColumns is a helper function for combining multiple calls to AddColumn.
func (t *DynamicTable) AddColumns(tableName string, args ...DynamicColumnDeclaration) {

	for _, decl := range args {
		_, err := t.AddColumn(
			tableName,
			decl.columnName,
			decl.evalType)
		if err != nil {
			panic(err)
		}
	}
}

// Rows returns the row iterator. The tableName passed is the alias
// for the table. Without that name, it will not be possible
// to join dynamic tables with themselves or alias them.
func (t *DynamicTable) Rows(tableName string) results.RowIter {
	return t.generator(tableName)
}
