package catalog

import (
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
)

// NewInMemoryTable creates a new InMemoryTable.
func NewInMemoryTable(name string, columns ...*results.Column) *InMemoryTable {
	var columnMap = make(map[string]*results.Column)
	for _, col := range columns {
		columnMap[strings.ToLower(col.Name)] = col
	}
	return &InMemoryTable{
		name:      name,
		columns:   columns,
		columnMap: columnMap,
	}
}

// InMemoryTable is an in-memory table.
type InMemoryTable struct {
	selectID  int
	name      string
	columns   results.Columns
	columnMap map[string]*results.Column
	Rows      results.Rows
}

// AddColumn adds a columns to the InMemoryTable, t.
func (t *InMemoryTable) AddColumn(name string, evalType types.EvalType) (*results.Column, error) {
	lowerName := strings.ToLower(name)
	if _, ok := t.columnMap[lowerName]; ok {
		return nil, mysqlerrors.Defaultf(mysqlerrors.ErDupFieldname, name)
	}

	cb := results.NewColumnBuilder()
	cb.SetColumnType(results.NewColumnType(evalType, schema.MongoNone))
	cb.SetSelectID(t.selectID)
	cb.SetTable("")
	cb.SetOriginalTable("")
	cb.SetDatabase("")
	cb.SetName(name)
	cb.SetOriginalName("")
	cb.SetMappingRegistryName("")
	cb.SetMongoName("")
	cb.SetPrimaryKey(false)
	cb.SetComments("")
	cb.SetIsPolymorphic(false)
	cb.SetHasAlteredType(false)
	c := cb.Build()

	t.selectID++

	t.columns = append(t.columns, c)
	t.columnMap[lowerName] = c

	return c, nil
}

// Collation returns the collation for the InMemoryTable.
func (*InMemoryTable) Collation() *collation.Collation {
	return collation.Default
}

// Column returns the column of the specified name.
func (t *InMemoryTable) Column(name string) (*results.Column, error) {
	if c, ok := t.columnMap[strings.ToLower(name)]; ok {
		return c, nil
	}
	return nil, mysqlerrors.Defaultf(mysqlerrors.ErBadFieldError, name, t.Name())
}

// Columns returns the columns for the InMemoryTable, t.
func (t *InMemoryTable) Columns() results.Columns {
	var cols results.Columns
	cols = append(cols, t.columns...)
	return cols
}

// Comments returns comments about the InMemoryTable.
func (*InMemoryTable) Comments() string {
	return ""
}

// ForeignKeys returns the foreign keys for the InMemoryTable.
func (*InMemoryTable) ForeignKeys() []ForeignKey {
	return nil
}

// Indexes returns the indexes for the InMemoryTable.
func (*InMemoryTable) Indexes() []Index {
	return nil
}

// Name returns the name for the InMemoryTable, t.
func (t *InMemoryTable) Name() string {
	return t.name
}

// PrimaryKeys returns the primary keys for
// the InMemoryTable.
func (*InMemoryTable) PrimaryKeys() results.Columns {
	return nil
}

// Type returns the type of the InMemoryTable.
func (*InMemoryTable) Type() string {
	return BaseTable
}
