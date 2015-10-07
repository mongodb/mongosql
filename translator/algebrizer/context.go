package algebrizer

import (
	"bytes"
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
	Columns ColumnInfos
	// Tables holds a slice of all tables in this context
	Tables TableInfos
	// Parent holds the context from which this one is dereived
	Parent *ParseCtx
	// Children holds all new contexts derived from this contenxt
	Children []*ParseCtx
	// NonStarAlias holds the alias for a non-star expression
	NonStarAlias string
	// ParseExpr holds what expression in a select query we're
	// currently processing
	Expr     interface{}
	Config   *config.Config
	Database string
}

func (pCtx *ParseCtx) String() string {
	if pCtx == nil {
		return ""
	}

	b := bytes.NewBufferString(fmt.Sprintf("In %#v\n", pCtx.Expr))

	pCtx.string(b, 0)

	return b.String()
}

func printTabs(b *bytes.Buffer, d int) {
	for i := 0; i < d; i++ {
		b.WriteString("\t")
	}
}

func (pCtx *ParseCtx) string(b *bytes.Buffer, d int) {

	if len(pCtx.Tables) != 0 {
		printTabs(b, d)
		b.WriteString(fmt.Sprintf("→ Tables\n"))
	}

	for i, table := range pCtx.Tables {
		printTabs(b, d)
		b.WriteString(fmt.Sprintf("\tT%v: → %#v\n", i, table))
	}

	if len(pCtx.Columns) != 0 {
		printTabs(b, d)
		b.WriteString(fmt.Sprintf("→ Columns\n"))
	}

	for i, column := range pCtx.Columns {
		printTabs(b, d)
		b.WriteString(fmt.Sprintf("\tC%v: → %#v\n", i, column))
	}

	if len(pCtx.Children) != 0 {
		printTabs(b, d)
		b.WriteString(fmt.Sprintf("→ Children\n"))
	}

	for i, ctx := range pCtx.Children {
		printTabs(b, d+1)
		b.WriteString(fmt.Sprintf("Child %v:\n", i))
		ctx.string(b, d+1)
	}
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

type TableInfos []TableInfo

func (tInfos TableInfos) Contains(t TableInfo) bool {
	for _, tInfo := range tInfos {
		if tInfo.Name == t.Name && tInfo.Db == t.Db &&
			tInfo.Alias == t.Alias && tInfo.Derived == t.Derived {
			return true
		}
	}
	return false
}

// ColumnInfo holds a mapping of aliases (or real names
// if not aliased) to the actual column name. The actual
// table name for the column is also stored here.
type ColumnInfo struct {
	Name  string
	Alias string
	Table string
}

type ColumnInfos []ColumnInfo

func (cInfos ColumnInfos) Contains(c *ColumnInfo) bool {
	for _, cInfo := range cInfos {
		if cInfo.Name == c.Name && cInfo.Alias == c.Alias && cInfo.Table == c.Table {
			return true
		}
	}
	return false
}

func (ci *ColumnInfo) String() string {
	return fmt.Sprintf("%v.%v", ci.Table, ci.Name)
}

func NewTableInfo(alias, db, longName string, derived bool) TableInfo {
	return TableInfo{alias, db, longName, derived}
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
func (pCtx *ParseCtx) TableInfo(db, alias string) (*TableInfo, error) {
	// no guarantee that this table exists without checking the db
	if pCtx == nil {
		return nil, fmt.Errorf("Unknown table '%v'", alias)
	}

	for _, tableInfo := range pCtx.Tables {
		if db == tableInfo.Db && alias == tableInfo.Alias {
			return &tableInfo, nil
		}
	}

	return pCtx.Parent.TableInfo(db, alias)
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

// AddColumns adds columns from the schema configuration to the context.
func (pCtx *ParseCtx) AddColumns() {
	for _, table := range pCtx.Tables {
		if table.Derived {
			continue
		}

		tableSchema := pCtx.TableSchema(table.Name)

		if tableSchema == nil {
			continue
		}

		for _, tableColumn := range tableSchema.Columns {
			column := ColumnInfo{
				Name:  tableColumn.Name,
				Alias: tableColumn.Name,
				Table: table.Name,
			}
			pCtx.Columns = append(pCtx.Columns, column)
		}
	}
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

	// for star expression on derived tables, add known columns
	if hasStarExpr(ss) {
		ctx.AddColumns()
	}

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

// checkColumn checks that a referenced column in the context of
// a derived table is valid.
func (pCtx *ParseCtx) checkColumn(table, column string) error {

	for _, ctx := range pCtx.Children {
		err := ctx.checkColumn(table, column)
		if err == nil {
			return nil
		}
	}

	for _, col := range pCtx.Columns {
		if col.Alias == column {
			return nil
		}
	}

	return fmt.Errorf("Unknown column '%v'", column)
}

// TableSchema returns the schema for a given table.
func (pCtx *ParseCtx) TableSchema(table string) *config.TableConfig {
	schema := pCtx.Config.Schemas[pCtx.Database]
	if schema == nil {
		return nil
	}

	return schema.Tables[table]
}

// CheckColumn checks that the column information is valid in the given context.
func (pCtx *ParseCtx) CheckColumn(tName, cName string) error {
	// whitelist all 'virtual' schemas including information_schema
	// TODO: more precise validation needed
	if strings.EqualFold(pCtx.Database, planner.InformationSchema) ||
		strings.EqualFold(tName, planner.InformationSchema) {
		return nil
	}

	table := pCtx.TableSchema(tName)

	if table == nil {
		// check if it's a derived table
		for _, tableInfo := range pCtx.Tables {
			if tableInfo.Alias == tName && tableInfo.Derived {
				return pCtx.checkColumn(tName, cName)
			}
		}
		return fmt.Errorf("Table '%v' doesn't exist", tName)
	}

	for _, c := range table.Columns {
		if c.Name == cName {
			return nil
		}
	}

	return fmt.Errorf("Unknown column '%v' in table %v", cName, tName)
}
