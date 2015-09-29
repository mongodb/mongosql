package algebrizer

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/translator/planner"
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
	Tables   []TableInfo
	Parent   *ParseCtx
	Children []*ParseCtx
	Config   *config.Config
	Database string
}

// TableInfo holds a mapping of aliases (or real names
// if not aliased) to the actual table name.
type TableInfo struct {
	Alias string
	Db    string
	Name  string
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

func NewTableInfo(alias, db, longName string) TableInfo {
	return TableInfo{alias, db, longName}
}

func (ci *ColumnInfo) String() string {
	return fmt.Sprintf("%v.%v", ci.Collection, ci.Field)
}

func NewParseCtx(ss sqlparser.SelectStatement, c *config.Config, db string) (*ParseCtx, error) {

	switch stmt := ss.(type) {

	case *sqlparser.Select:
		ctx := &ParseCtx{
			Config:   c,
			Database: db,
		}

		tableInfo, err := GetTableInfo(stmt.From, ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting table info: %v", err)
		}

		ctx.Tables = tableInfo
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

	for _, tableInfo := range pCtx.Tables {
		if qualifier == tableInfo.Db && alias == tableInfo.Alias {
			return tableInfo.Name, nil
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

	ctx, err := NewParseCtx(ss, pCtx.Config, pCtx.Database)
	if err != nil {
		return nil, err
	}

	ctx.Parent = pCtx
	pCtx.Children = append(pCtx.Children, ctx)

	return ctx, nil
}

// GetCurrentTable finds a given table in the current context.
func (pCtx *ParseCtx) GetCurrentTable(dbName, tableName string) (*TableInfo, error) {
	if pCtx == nil {
		return nil, fmt.Errorf("Table '%v' doesn't exist", tableName)
	}

	if len(pCtx.Tables) == 0 {
		return pCtx.Parent.GetCurrentTable(dbName, tableName)
	} else if len(pCtx.Tables) == 1 {
		// in this case, we're either referencing a table directly,
		// using an alias or implicitly referring to the current table

		curTable := pCtx.Tables[0]

		if curTable.Alias == tableName || tableName == "" {
			if curTable.Db == dbName {
				return &curTable, nil
			}
		}
	} else {
		// if there are multiple tables in the current context
		// then tableName must not be empty
		if tableName == "" {
			return nil, fmt.Errorf("ambiguity in column field(s) - qualify with table name")
		}

		// the table name can either be actual or aliased
		for _, tableInfo := range pCtx.Tables {
			if tableInfo.Alias == tableName {
				if tableInfo.Db == dbName {
					return &tableInfo, nil
				}
			}
		}
	}

	return pCtx.Parent.GetCurrentTable(dbName, tableName)
}

// CheckColumn checks that the column information is valid in the given context.
func (pCtx *ParseCtx) CheckColumn(table, column string) error {
	// whitelist all 'virtual' schemas including information_schema
	// TODO: more precise validation needed
	if strings.EqualFold(pCtx.Database, planner.InformationSchema) ||
		strings.EqualFold(table, planner.InformationSchema) {
		return nil
	}

	schema := pCtx.Config.Schemas[pCtx.Database]
	if schema == nil {
		return fmt.Errorf("Schema for db '%v' doesn't exist", pCtx.Database)
	}

	collection := schema.Tables[table]
	if collection == nil {
		return fmt.Errorf("Table '%v' doesn't exist", table)
	}

	for _, c := range collection.Columns {
		if c.Name == column {
			return nil
		}
	}

	return fmt.Errorf("Unknown column '%v' in table %v", column, table)
}
