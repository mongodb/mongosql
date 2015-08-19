package translator

import (
	"fmt"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/siddontang/mixer/sqlparser"
)

// ParseCtx holds information that is used to resolve
// columns and table names in a select query.
type ParseCtx struct {
	Column []ColumnInfo
	Table  []TableInfo
	Parent *ParseCtx
}

// TableInfo holds a mapping of aliases (or real names
// if not aliased) to the actual table name.
type TableInfo struct {
	Name map[string]string
}

// ColumnInfo holds a mapping of aliases (or real names
// if not aliased) to the actual column name. The actual
// table name for the column is also stored here.
type ColumnInfo struct {
	// using a mapping in case the name is an alias
	// e.g. SELECT a+b AS x FROM foo WHERE x<10;
	Name  map[string]string
	Table string
}

func NewParseCtx(stmt *sqlparser.Select) (*ParseCtx, error) {

	ctx := &ParseCtx{}

	tableInfo, err := getTableInfo(stmt.From, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting table info: %v", err)
	}

	ctx.Table = tableInfo

	// handle select expressions like as aliasing
	// e.g. select FirstName as f, LastName as l from foo;
	columnInfo, err := getColumnInfo(stmt.SelectExprs, ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting column info: %v", err)
	}

	ctx.Column = columnInfo

	return ctx, nil
}

// TableName returns the unaliased table name given an alias.
// It searches in the parent context if the alias is not found
// in the current context.
func (c *ParseCtx) TableName(alias string) string {
	if c == nil {
		return alias
	}
	for _, a := range c.Table {
		if name, ok := a.Name[alias]; ok {
			return name
		}
	}

	return c.Parent.TableName(alias)
}

// ColumnName returns the unaliased column name given an alias.
// It searches in the parent context if the alias is not found
// in the current context.
func (c *ParseCtx) ColumnName(alias string) string {
	if c == nil {
		return alias
	}

	for _, a := range c.Column {
		if name, ok := a.Name[alias]; ok {
			return name
		}
	}

	return c.Parent.ColumnName(alias)
}

// GetDefaultTable finds a given table in the current context.
// It searches in the parent context if the alias is not found
// in the current context.
func (c *ParseCtx) GetDefaultTable() string {
	if c == nil {
		return ""
	}

	if len(c.Table) != 1 {
		log.Logf(log.DebugLow, "found more than one 'default' table returning ''")
		return ""
	}

	for k, _ := range c.Table[0].Name {
		return k
	}

	return c.Parent.GetDefaultTable()
}
