package algebrizer

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
)

var (
	ErrNilCtx              = fmt.Errorf("received nil context")
	ErrColumnAliasNotFound = fmt.Errorf("no column alias found")
	ErrTableAliasNotFound  = fmt.Errorf("no table alias found")
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
	Alias      string
	Collection string
}

// ColumnInfo holds a mapping of aliases (or real names
// if not aliased) to the actual column name. The actual
// table name for the column is also stored here.
type ColumnInfo struct {
	// using a mapping in case the name is an alias
	// e.g. SELECT a+b AS x FROM foo WHERE x<10;
	Field string
	// TODO: parser does not currently support column aliasing :(
	Alias      string
	Collection string
}

func (ci *ColumnInfo) String() string {
	return fmt.Sprintf("%v.%v", ci.Collection, ci.Field)
}

func NewParseCtx(ss sqlparser.SelectStatement) (*ParseCtx, error) {

	switch stmt := ss.(type) {

	case *sqlparser.Select:
		ctx := &ParseCtx{}

		tableInfo, err := GetTableInfo(stmt.From, ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting table info: %v", err)
		}

		ctx.Table = tableInfo

		return ctx, nil

	default:
		return nil, fmt.Errorf("select statement type %T not yet supported", stmt)

	}

}

// TableName returns the unaliased table name given an alias.
// It searches in the parent context if the alias is not found
// in the current context.
func (c *ParseCtx) TableName(alias string) (string, error) {
	// no guarantee that this table exists without checking the db
	if c == nil {
		return alias, nil
	}

	for _, tableInfo := range c.Table {
		if alias == tableInfo.Alias {
			return tableInfo.Collection, nil
		}
	}

	return c.Parent.TableName(alias)
}

// TableExists returns true if an table within this or the parent
// context matches 'tableName' unaliased table name given an alias.
// It searches in the parent context if the alias is not found
// in the current context.
func (c *ParseCtx) TableExists(tableName string) bool {
	if c == nil {
		return false
	}

	for _, tableInfo := range c.Table {
		if tableName == tableInfo.Collection {
			return true
		}
	}

	return c.Parent.TableExists(tableName)
}

// ColumnAlias searches current context for the given alias
// It searches in the parent context if the alias is not found
// in the current context.
func (c *ParseCtx) ColumnAlias(alias string) (*ColumnInfo, error) {
	if c == nil {
		return nil, ErrColumnAliasNotFound
	}

	for _, columnInfo := range c.Column {
		if columnInfo.Alias == alias {
			return &columnInfo, nil
		}
	}

	return c.Parent.ColumnAlias(alias)
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

	return c.Table[0].Collection
}

// GetCurrentTable finds a given table in the current context.
func (c *ParseCtx) GetCurrentTable(tableName string) (string, error) {
	if c == nil {
		return "", ErrNilCtx
	}

	if len(c.Table) == 0 {
		return c.Parent.GetCurrentTable(tableName)
	} else if len(c.Table) == 1 {
		// in this case, we're either referencing a table directly,
		// using an alias or implicitly referring to the current table

		curTable := c.Table[0]

		if curTable.Collection == tableName || curTable.Alias == tableName || tableName == "" {
			return c.Table[0].Collection, nil
		}
	} else {
		// if there are multiple tables in the current context
		// then tableName must not be empty
		if tableName == "" {
			return "", fmt.Errorf("found more than one table in context")
		}

		// the table name can either be actual or aliased
		for _, tableInfo := range c.Table {
			if tableInfo.Alias == tableName || tableInfo.Collection == tableName {
				return tableInfo.Collection, nil
			}
		}
	}

	return c.Parent.GetCurrentTable(tableName)
}
