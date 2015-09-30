package algebrizer

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/translator/planner"
	"strings"
)

var (
	ErrNilCtx = fmt.Errorf("received nil context")
)

// ParseCtx holds information that is used to resolve
// columns and table names in a select query.
type ParseCtx struct {
	// Columns holds a slice of all columns in this context
	Columns []ColumnInfo
	// Tables holds a slice of all tables in this context
	Tables []TableInfo
	// Parent holds the context from which this one is dereived
	Parent *ParseCtx
	// Children holds all new contexts derived from this contenxt
	Children []*ParseCtx
	// NonStarAlias holds the alias for a non-star expression
	NonStarAlias string
	Config       *config.Config
	Database     string
}

// TableInfo holds a mapping of aliases (or real names
// if not aliased) to the actual table name.
type TableInfo struct {
	Alias string
	Db    string
	Name  string
	// Derived indicates if context table is from a subquery
	Derived bool
}

// ColumnInfo holds a mapping of aliases (or real names
// if not aliased) to the actual column name. The actual
// table name for the column is also stored here.
type ColumnInfo struct {
	Name  string
	Alias string
	Table string
}

func NewTableInfo(alias, db, longName string, derived bool) TableInfo {
	return TableInfo{alias, db, longName, derived}
}

func (ci *ColumnInfo) String() string {
	return fmt.Sprintf("%v.%v", ci.Table, ci.Name)
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

// TableInfo returns the table infor for the given alias.
// It searches in the parent context if the alias is not found
// in the current context.
func (pCtx *ParseCtx) TableInfo(qualifier string, alias string) (*TableInfo, error) {
	// no guarantee that this table exists without checking the db
	if pCtx == nil {
		return nil, fmt.Errorf("Unknown table '%v'", alias)
	}

	for _, tableInfo := range pCtx.Tables {
		if qualifier == tableInfo.Db && alias == tableInfo.Alias {
			return &tableInfo, nil
		}
	}

	return pCtx.Parent.TableInfo(qualifier, alias)
}

// ColumnInfo searches current context for the given alias
// It searches in the parent context if the alias is not found
// in the current context.
func (pCtx *ParseCtx) ColumnInfo(alias string) (*ColumnInfo, error) {
	if pCtx == nil {
		return nil, fmt.Errorf("Unknown column '%v'", alias)
	}

	for _, columnInfo := range pCtx.Columns {
		if columnInfo.Alias == alias {
			return &columnInfo, nil
		}
	}

	return pCtx.Parent.ColumnInfo(alias)
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
		return nil, fmt.Errorf("Current table '%v' doesn't exist", tableName)
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
		// check if it's a derived table
		for _, tableInfo := range pCtx.Tables {
			if tableInfo.Alias == table && tableInfo.Derived {
				return nil
			}
		}

		return fmt.Errorf("Table '%v' doesn't exist", table)
	}

	for _, c := range collection.Columns {
		if c.Name == column {
			return nil
		}
	}

	return fmt.Errorf("Unknown column '%v' in table %v", column, table)
}
