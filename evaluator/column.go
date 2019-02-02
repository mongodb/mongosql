package evaluator

import (
	"strings"

	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/schema"
)

// ColumnType is the type of a column.
type ColumnType struct {
	EvalType    types.EvalType
	MongoType   schema.MongoType
	UUIDSubType types.EvalType
}

// Column contains information used to select data
// from a PlanStage.
type Column struct {
	ColumnType
	SelectID            int
	Table               string
	OriginalTable       string
	Database            string
	Name                string
	OriginalName        string
	MappingRegistryName string
	PrimaryKey          bool
}

// NewColumnType returns a ColumnType with the specified types.EvalType and MongoType.
func NewColumnType(evalType types.EvalType, mongoType schema.MongoType) ColumnType {
	return ColumnType{
		EvalType:  evalType,
		MongoType: mongoType,
		// Because the need to set the UUIDSubType is so rare, we just use
		// the default EvalBinary encoding unless otherwise specified with the
		// other constructor.
		UUIDSubType: types.EvalBinary,
	}
}

// NewColumnTypeWithUUIDSubtype returns a ColumnType with the specified types.EvalType, MongoType, and
// UUIDSubType.
func NewColumnTypeWithUUIDSubtype(evalType types.EvalType,
	mongoType schema.MongoType,
	uuidSubType types.EvalType) ColumnType {
	return ColumnType{
		EvalType:    evalType,
		MongoType:   mongoType,
		UUIDSubType: uuidSubType,
	}
}

// NewColumn is a constructor for the Column struct.
func NewColumn(selectID int, table, originalTable, database, name,
	originalName, mappingRegistryName string, evalType types.EvalType,
	mongoType schema.MongoType, primaryKey bool) *Column {
	uuidSubType := types.EvalBinary
	if mongoType == schema.MongoUUIDJava {
		uuidSubType = types.EvalJavaUUID
	} else if mongoType == schema.MongoUUIDCSharp {
		uuidSubType = types.EvalCSharpUUID
	}
	return &Column{
		ColumnType: ColumnType{
			MongoType:   mongoType,
			EvalType:    evalType,
			UUIDSubType: uuidSubType,
		},
		SelectID:            selectID,
		Table:               table,
		OriginalTable:       originalTable,
		Database:            database,
		Name:                name,
		OriginalName:        originalName,
		MappingRegistryName: mappingRegistryName,
		PrimaryKey:          primaryKey,
	}
}

// NewColumnFromSQLColumnExpr returns a new Column struct created
// using the values from the SQLColumnExpr and isPrimaryKey.
func NewColumnFromSQLColumnExpr(sqlColExpr SQLColumnExpr, isPrimaryKey bool) *Column {
	return NewColumn(
		sqlColExpr.selectID,
		sqlColExpr.tableName,
		"",
		sqlColExpr.databaseName,
		sqlColExpr.columnName,
		"",
		"",
		sqlColExpr.columnType.EvalType,
		sqlColExpr.columnType.MongoType,
		isPrimaryKey,
	)
}

func (c *Column) clone() *Column {
	return NewColumn(c.SelectID,
		c.Table,
		c.OriginalTable,
		c.Database,
		c.Name,
		c.OriginalName,
		c.MappingRegistryName,
		c.EvalType,
		c.MongoType,
		c.PrimaryKey,
	)
}

func (c *Column) expr() SQLColumnExpr {
	return NewSQLColumnExpr(c.SelectID,
		c.Database,
		c.Table,
		c.Name,
		c.EvalType,
		c.MongoType,
	)
}

func (c *Column) projectAs(name string) ProjectedColumn {
	clone := c.clone()
	clone.Name = name
	return ProjectedColumn{
		Column: clone,
		Expr:   c.expr(),
	}
}

func (c *Column) projectWithExpr(expr SQLExpr) *ProjectedColumn {
	clone := c.clone()
	clone.EvalType = expr.EvalType()
	return &ProjectedColumn{
		Column: clone,
		Expr:   expr,
	}
}

// Columns is a slice of Column pointers.
type Columns []*Column

// FindByName searches Columns for a column of a matching name.
func (cs Columns) FindByName(name string) (*Column, bool) {
	for _, c := range cs {
		if strings.EqualFold(name, c.Name) {
			return c, true
		}
	}

	return nil, false
}

// ToProjectedColumns converts cs to ProjectedColumns
// using the SQLColumnExpr type - constructed from
// the ProjectedColumn as the wrapped expression.
func (cs Columns) ToProjectedColumns() ProjectedColumns {
	var projectedColumns ProjectedColumns
	for _, c := range cs {
		projectedColumn := ProjectedColumn{
			Expr: NewSQLColumnExpr(c.SelectID, c.Database,
				c.Table, c.Name, c.EvalType, c.MongoType),
			Column: c.clone(),
		}
		projectedColumns = append(projectedColumns, projectedColumn)
	}
	return projectedColumns
}

// Unique ensures that only unique columns exist in the resulting slice.
func (cs Columns) Unique() Columns {
	var results Columns
	contains := func(column *Column) bool {
		for _, c := range results {
			if c.SelectID == column.SelectID &&
				c.Name == column.Name &&
				c.Table == column.Table &&
				c.Database == column.Database {
				return true
			}
		}

		return false
	}

	for _, c := range cs {
		if !contains(c) {
			results = append(results, c)
		}
	}

	return results
}

// ProjectedColumn is a column projection. It contains the SQLExpr for the column
// as well as the column information that will be projected.
type ProjectedColumn struct {
	// Column holds the projection information.
	*Column

	// Expr holds the expression to be evaluated.
	Expr SQLExpr
}

// ProjectedColumns is a slice of ProjectedColumn.
type ProjectedColumns []ProjectedColumn

// Exprs returns a slice of the expressions within pcs.
func (pcs ProjectedColumns) Exprs() []SQLExpr {
	exprs := []SQLExpr{}
	for _, pc := range pcs {
		exprs = append(exprs, pc.Expr)
	}
	return exprs
}

// Unique ensures that only unique projected columns exist in the resulting slice.
func (pcs ProjectedColumns) Unique() ProjectedColumns {
	var results ProjectedColumns
	contains := func(column *ProjectedColumn) bool {
		for _, expr := range results {
			if expr.Column.SelectID == column.SelectID &&
				expr.Column.Name == column.Name &&
				expr.Column.Table == column.Table &&
				expr.Column.Database == column.Database {
				return true
			}
		}

		return false
	}

	for _, c := range pcs {
		if !contains(&c) {
			results = append(results, c)
		}
	}

	return results
}
