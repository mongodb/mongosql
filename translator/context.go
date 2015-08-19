package translator

import (
	"github.com/mongodb/mongo-tools/common/log"
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
