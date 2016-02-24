package evaluator

import (
	"bytes"
	"fmt"
	"github.com/10gen/sqlproxy/schema"
	"github.com/deafgoat/mixer/sqlparser"
	"strings"
	"unsafe"
)

var (
	ErrNilCtx = fmt.Errorf("received nil context")
)

// ParseCtx holds information that is used to resolve
// columns and table names in a select query.
type ParseCtx struct {
	// ColumnReferences holds a slice of all columns that reference other
	// more complex expressions. For example, in:
	//
	// select a+b as c from foo;
	//
	// ColumnReferences will contain an entry indicating c references the
	// a+b expression.
	ColumnReferences ColumnReferences

	// Columns holds a slice of all columns in this context
	Columns ColumnInfos

	// Tables holds a slice of all tables in this context
	Tables TableInfos

	// Parent holds the context from which this one is derived
	Parent *ParseCtx

	// Children holds all new contexts derived from this context
	Children []*ParseCtx

	// NonStarAlias holds the column alias for a non-star expression
	NonStarAlias string

	// NonStarAlias holds the table alias for a non-star subquery expression
	DerivedTableName string

	// Expr holds the expression in a select query that is being
	// processed
	Expr sqlparser.Expr

	// Phase describes where in a select clause we're algebrizing.
	// e.g. if we're algebrizing within a where clause, the Phase will
	// hold the value of 'PhaseWhere'
	Phase string

	// State holds information on what kind of select expression is being parsed
	State ParseState

	// Schema hold the schema configuration data for all the databases and tables
	Schema *schema.Schema

	// Database holds the current database referenced in the select expression
	Database string
}

// ParseState holds bookkeeping information for parsing.
type ParseState uint32

// Parse states.
const (
	StateFuncExpr     ParseState = 1 << iota
	StateRefColExpr              = 2 << iota
	StateSubQueryExpr            = 3 << iota
)

// Parse context phases.
const (
	PhaseInit       = "INIT"
	PhaseOrderBy    = "ORDER BY"
	PhaseGroupBy    = "GROUP BY"
	PhaseHaving     = "HAVING"
	PhaseFrom       = "FROM"
	PhaseWhere      = "WHERE"
	PhaseSelectExpr = "SELECT EXPRESSION"
	PhaseLimit      = "LIMIT"
)

func (pCtx *ParseCtx) InFuncExpr() bool {
	return (pCtx.State & StateFuncExpr) > 0
}

func (pCtx *ParseCtx) InRefColumn() bool {
	return (pCtx.State & StateRefColExpr) > 0
}

func (pCtx *ParseCtx) InSubquery() bool {
	return (pCtx.State & StateSubQueryExpr) > 0
}

func (state ParseState) String() string {
	states := []string{}

	if (state & StateFuncExpr) > 0 {
		states = append(states, "StateFuncExpr")
	}

	if (state & StateRefColExpr) > 0 {
		states = append(states, "StateRefColExpr")
	}

	if (state & StateSubQueryExpr) > 0 {
		states = append(states, "StateSubQueryExpr")
	}

	return strings.Join(states, ",")
}

func (pCtx *ParseCtx) String() string {
	if pCtx == nil {
		return ""
	}

	addr := unsafe.Pointer(pCtx)

	b := bytes.NewBufferString(fmt.Sprintf("%v Phase (%v); State (%v):", addr, pCtx.Phase, pCtx.State))

	if pCtx.Expr != nil {
		b.WriteString(fmt.Sprintf(" %v", sqlparser.String(pCtx.Expr)))
	}

	b.WriteString("\n")

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
		b.WriteString(fmt.Sprintf("→ Tables (%v)\n", len(pCtx.Tables)))
	}

	for i, table := range pCtx.Tables {
		printTabs(b, d)
		b.WriteString(fmt.Sprintf("\tT%v: → %#v\n", i, table))
	}

	if len(pCtx.Columns) != 0 {
		printTabs(b, d)
		b.WriteString(fmt.Sprintf("→ Columns (%v)\n", len(pCtx.Columns)))
	}

	for i, column := range pCtx.Columns {
		printTabs(b, d)
		b.WriteString(fmt.Sprintf("\tC%v: → %#v\n", i, column))
	}

	if len(pCtx.ColumnReferences) != 0 {
		printTabs(b, d)
		b.WriteString(fmt.Sprintf("→ References (%v)\n", len(pCtx.ColumnReferences)))
	}

	for i, reference := range pCtx.ColumnReferences {
		printTabs(b, d)
		b.WriteString(fmt.Sprintf("\tR%v: → %#v\n", i, reference))
	}

	if len(pCtx.Children) != 0 {
		printTabs(b, d)
		b.WriteString(fmt.Sprintf("→ Children (%v)\n", len(pCtx.Children)))
	}

	for i, ctx := range pCtx.Children {
		printTabs(b, d+1)
		addr := unsafe.Pointer(ctx)
		b.WriteString(fmt.Sprintf("Child %v: %v\n", i, addr))
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

type ColumnReferences []ColumnReference

type ColumnReference struct {
	Name  string
	Table string
	Expr  sqlparser.Expr
	Index int
}

// ColumnInfo holds a mapping of aliases (or real names
// if not aliased) to the actual column name. The actual
// table name for the column is also stored here.
type ColumnInfo struct {
	Name  string
	Alias string
	Table string
	Index int
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

func NewParseCtx(ss sqlparser.SelectStatement, c *schema.Schema, db string) (*ParseCtx, error) {

	switch stmt := ss.(type) {

	case *sqlparser.Select:
		ctx := &ParseCtx{
			Schema:   c,
			Database: db,
			Phase:    PhaseInit,
		}

		tableInfo, err := GetTableInfo(stmt.From, ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting table info: %v", err)
		}

		for _, table := range tableInfo {
			if !ctx.Tables.Contains(table) {
				ctx.Tables = append(ctx.Tables, table)
			}
		}

		return ctx, nil

	default:
		return nil, fmt.Errorf("select statement type %T not yet supported", stmt)
	}

}

// TableInfo returns the table infor for the given alias.
// It searches in the parent context if the alias is not found
// in the current context.
func (pCtx *ParseCtx) TableInfo(db, alias string) (*TableInfo, error) {
	return pCtx.tableInfo(db, alias, 0)
}

func (pCtx *ParseCtx) tableInfo(db, alias string, depth int) (*TableInfo, error) {

	if pCtx != nil && pCtx.Parent != nil && depth == 0 {
		return pCtx.Parent.tableInfo(db, alias, depth)
	}

	for _, tableInfo := range pCtx.Tables {
		if db == tableInfo.Db && alias == tableInfo.Alias {
			return &tableInfo, nil
		}
	}

	for _, ctx := range pCtx.Children {
		tableInfo, err := ctx.tableInfo(db, alias, depth+1)
		if err == nil {
			return tableInfo, nil
		}
	}

	return nil, fmt.Errorf("Unknown table '%v'", alias)
}

// ColumnInfo searches current context for the given alias
// It searches in the parent context if the alias is not found
// in the current context.
func (pCtx *ParseCtx) ColumnInfo(column string) (*ColumnInfo, error) {
	if pCtx == nil {
		return nil, fmt.Errorf("Unknown column alias '%v'", column)
	}

	for _, columnInfo := range pCtx.Columns {
		if columnInfo.Alias == column {
			return &columnInfo, nil
		}
	}

	return pCtx.Parent.ColumnInfo(column)
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

		for index, colDef := range tableSchema.RawColumns {
			column := ColumnInfo{
				Index: index,
				Name:  colDef.SqlName,
				Alias: colDef.SqlName,
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

	ctx, err := NewParseCtx(ss, pCtx.Schema, pCtx.Database)
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
func (pCtx *ParseCtx) GetCurrentTable(dbName, tableName, columnName string) (*TableInfo, error) {

	if pCtx == nil {
		return nil, fmt.Errorf("Current table '%v' doesn't exist", tableName)
	}

	if len(pCtx.Tables) == 1 {
		// in this case, we're either referencing a table directly,
		// using an alias or implicitly referring to the current table

		curTable := pCtx.Tables[0]

		if curTable.Alias == tableName || tableName == "" {
			if curTable.Db == dbName {
				return &curTable, nil
			}
		}
	} else if len(pCtx.Tables) > 1 {

		// if there are multiple tables in the current context
		// then tableName may or may not be qualified
		if tableName == "" {

			if pCtx.Parent == nil {

				var tInfo TableInfo

				// indicates a star expression
				if columnName == "" {
					return nil, fmt.Errorf("Can not find empty table in join context")
				}

				column := &Column{Name: columnName}

				var count int64

				for _, tableInfo := range pCtx.Tables {

					// derived tables must be qualified
					if tableInfo.Derived {
						continue
					}

					column.Table = tableInfo.Alias

					if err := pCtx.CheckColumn(&tableInfo, columnName); err != nil {

						if pCtx.IsSchemaColumn(column) {
							tInfo = tableInfo
							count++
						}
						continue
					}

					tInfo = tableInfo
					count++
				}

				if count == 0 {
					return nil, fmt.Errorf("Column '%v' not found", columnName)
				} else if count > 1 {
					return nil, fmt.Errorf("Column '%v' in field list is ambiguous", columnName)

				}
				return &tInfo, nil
			}

			return pCtx.Parent.GetCurrentTable(dbName, tableName, columnName)
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

	return pCtx.Parent.GetCurrentTable(dbName, tableName, columnName)
}

// checkColumn checks that a referenced column in the context of
// a derived table is valid.
func (pCtx *ParseCtx) checkColumn(table, column string, depth int) error {

	err := fmt.Errorf("Unknown column '%v'", column)

	if pCtx == nil {
		return err
	}

	for _, c := range pCtx.Columns {

		if c.Alias == column {
			return nil
		}

	}

	for _, r := range pCtx.ColumnReferences {

		if r.Name == column {
			return nil
		}
	}

	// only look deeper if columns are referenced from
	// the child context
	if depth == 0 || len(pCtx.Columns) == 0 {

		for _, ctx := range pCtx.Children {

			for _, c := range ctx.Columns {

				if c.Alias == column {
					return nil
				}
			}

			for _, r := range ctx.ColumnReferences {

				if r.Name == column {
					return nil
				}

			}

			if err = ctx.checkColumn(table, column, depth+1); err == nil {
				return nil
			}

		}
	}

	return err
}

// TableSchema returns the named table's configuration.
func (pCtx *ParseCtx) TableSchema(table string) *schema.Table {
	db := pCtx.Schema.Databases[pCtx.Database]

	if db == nil {
		return nil
	}

	return db.Tables[table]
}

// IsColumnReference returns true if the given column is present
// column references for this context.
func (pCtx *ParseCtx) IsColumnReference(column *Column) bool {

	for _, reference := range pCtx.ColumnReferences {

		if reference.Name == column.Name && reference.Table == column.Table {
			return true
		}
	}

	return false

}

// IsSchemaColumn returns true if the given column is present
// schema table's configuration.
func (pCtx *ParseCtx) IsSchemaColumn(column *Column) bool {

	if strings.ToLower(pCtx.Database) == InformationDatabase {
		return true
	}

	tableInfo, err := pCtx.TableInfo(pCtx.Database, column.Table)
	if err != nil {
		panic(fmt.Sprintf("Unknown column table '%v'", column.Table))
	}

	if tableInfo.Derived {
		return true
	}

	db := pCtx.Schema.Databases[pCtx.Database]

	if db == nil {
		return false
	}

	table := db.Tables[tableInfo.Name]

	if table == nil {
		return false
	}

	return table.SQLColumns[column.Name] != nil

}

// CheckColumn checks that the column information is valid in the given context.
func (pCtx *ParseCtx) CheckColumn(table *TableInfo, cName string) error {

	tName := table.Alias

	// whitelist all 'virtual' schemas including information_schema
	// TODO: more precise validation needed
	if strings.EqualFold(pCtx.Database, InformationDatabase) ||
		strings.EqualFold(tName, InformationDatabase) {
		return nil
	}

	if table.Derived {
		// For derived tables, column names must come from children
		for _, tableInfo := range pCtx.Tables {
			if tableInfo.Alias == tName {
				return pCtx.checkColumn(tName, cName, 0)
			}
		}

		return fmt.Errorf("Derived table '%v' doesn't exist", tName)
	}

	// For schema tables, columns must come from schema configuration
	tableSchema := pCtx.TableSchema(table.Name)

	if tableSchema == nil {
		return fmt.Errorf("Table '%v' doesn't exist", tName)
	}

	for _, c := range tableSchema.RawColumns {

		if c.SqlName == cName {
			return nil
		}

	}

	for _, r := range pCtx.ColumnReferences {

		if r.Name == cName {
			return nil
		}

	}

	return fmt.Errorf("Unknown column '%v' in table '%v'", cName, tName)
}

// GetTableColumns returns all the columns for a derived table.
func (pCtx *ParseCtx) GetTableColumns(table *TableInfo) sqlparser.SelectExprs {

	exprs := make(map[int]sqlparser.SelectExpr, 0)

	for _, child := range pCtx.Children {

		for _, column := range child.Columns {
			if column.Table == table.Name {
				expr := &sqlparser.ColName{
					Name:      []byte(column.Alias),
					Qualifier: []byte(column.Table),
				}
				exprs[column.Index] = &sqlparser.NonStarExpr{Expr: expr}
			}
		}

		for _, ref := range child.ColumnReferences {
			if ref.Table == table.Name {
				expr := &sqlparser.ColName{
					Name:      []byte(ref.Name),
					Qualifier: []byte(ref.Table),
				}
				exprs[ref.Index] = &sqlparser.NonStarExpr{Expr: expr}
			}
		}
	}

	selectExprs := sqlparser.SelectExprs{}

	for i := 0; i < len(exprs); i++ {
		selectExprs = append(selectExprs, exprs[i])
	}

	return selectExprs

}
