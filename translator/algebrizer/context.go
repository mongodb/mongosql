package algebrizer

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"strings"
)

var (
	ErrNilCtx              = fmt.Errorf("received nil context")
	ErrColumnAliasNotFound = fmt.Errorf("no column alias found")
	ErrTableAliasNotFound  = fmt.Errorf("no table alias found")
)

// ParseCtx holds information that is used to resolve
// columns and table names in a select query.
type ParseCtx struct {
	Column   []ColumnInfo
	Table    []TableInfo
	Parent   *ParseCtx
	Children []*ParseCtx
}

// TableInfo holds a mapping of aliases (or real names
// if not aliased) to the actual table name.
type TableInfo struct {
	Alias string
	Db    string
	Table string
}

// ColumnInfo holds a mapping of aliases (or real names
// if not aliased) to the actual column name. The actual
// table name for the column is also stored here.
type ColumnInfo struct {
	// using a mapping in case the name is an alias
	// TODO: e.g. SELECT a+b AS x FROM foo WHERE x<10;
	// x should be replaced with 'a+b' expr.
	Field      string
	Alias      string
	Collection string
}

func NewTableInfo(alias string, longName string) TableInfo {
	if strings.Index(longName, ".") > 0 {
		pcs := strings.SplitN(longName, ".", 2)
		return TableInfo{pcs[1], pcs[0], pcs[1]}
	}

	return TableInfo{alias, "", longName}
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
func (pCtx *ParseCtx) TableName(qualifier string, alias string) (string, error) {
	// no guarantee that this table exists without checking the db
	if pCtx == nil {
		return alias, nil
	}

	for _, tableInfo := range pCtx.Table {
		if qualifier == tableInfo.Db && alias == tableInfo.Alias {
			return tableInfo.Table, nil
		}
	}

	return pCtx.Parent.TableName(qualifier, alias)
}

// ColumnAlias searches current context for the given alias
// It searches in the parent context if the alias is not found
// in the current context.
func (pCtx *ParseCtx) ColumnAlias(alias string) (*ColumnInfo, error) {
	if pCtx == nil {
		return nil, ErrColumnAliasNotFound
	}

	for _, columnInfo := range pCtx.Column {
		if columnInfo.Alias == alias {
			return &columnInfo, nil
		}
	}

	return pCtx.Parent.ColumnAlias(alias)
}

// ChildCtx returns a new child context for the current context.
func (pCtx *ParseCtx) ChildCtx(ss sqlparser.SelectStatement) (*ParseCtx, error) {
	if pCtx == nil {
		return nil, ErrNilCtx
	}

	ctx, err := NewParseCtx(ss)
	if err != nil {
		return nil, err
	}

	ctx.Parent = pCtx
	pCtx.Children = append(pCtx.Children, ctx)

	return ctx, nil
}

// GetCurrentTable finds a given table in the current context.
func (pCtx *ParseCtx) GetCurrentTable(tableName string) (string, error) {
	if pCtx == nil {
		return "", ErrNilCtx
	}

	if len(pCtx.Table) == 0 {
		return pCtx.Parent.GetCurrentTable(tableName)
	} else if len(pCtx.Table) == 1 {
		// in this case, we're either referencing a table directly,
		// using an alias or implicitly referring to the current table

		curTable := pCtx.Table[0]

		// TODO: this is broken, ignores db
		if curTable.Table == tableName || curTable.Alias == tableName || tableName == "" {
			// TODO: this is broken, ignores db
			return pCtx.Table[0].Table, nil
		}
	} else {
		// if there are multiple tables in the current context
		// then tableName must not be empty
		if tableName == "" {
			return "", fmt.Errorf("found more than one table in context")
		}

		// the table name can either be actual or aliased
		for _, tableInfo := range pCtx.Table {
			// TODO: this is broken, ignores db
			if tableInfo.Alias == tableName || tableInfo.Table == tableName {
				// TODO: this is broken, ignores db
				return tableInfo.Table, nil
			}
		}
	}

	return pCtx.Parent.GetCurrentTable(tableName)
}
