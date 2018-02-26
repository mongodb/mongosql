package evaluator

import (
	"strings"

	"github.com/10gen/sqlproxy/schema"
)

// Column contains information used to select data
// from a PlanStage.
type Column struct {
	SelectID            int
	Table               string
	OriginalTable       string
	Database            string
	Name                string
	OriginalName        string
	MappingRegistryName string
	SQLType             schema.SQLType
	MongoType           schema.MongoType
	PrimaryKey          bool
}

// NewColumn is a constructor for the Column struct.
func NewColumn(selectID int, table, originalTable, database, name, originalName, mappingRegistryName string, sqlType schema.SQLType, mongoType schema.MongoType, primaryKey bool) *Column {
	return &Column{selectID, table, originalTable, database, name, originalName, mappingRegistryName,
		sqlType, mongoType, primaryKey}
}

func (c *Column) clone() *Column {
	return NewColumn(c.SelectID, c.Table, c.OriginalTable, c.Database, c.Name, c.OriginalName, c.MappingRegistryName, c.SQLType, c.MongoType, c.PrimaryKey)
}

func (c *Column) expr() SQLColumnExpr {
	return NewSQLColumnExpr(c.SelectID, c.Database, c.Table, c.Name, c.SQLType, c.MongoType)
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
	clone.SQLType = expr.Type()
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

func (se *ProjectedColumn) clone() *ProjectedColumn {
	return &ProjectedColumn{
		Column: se.Column,
		Expr:   se.Expr,
	}
}

// ProjectedColumns is a slice of ProjectedColumn.
type ProjectedColumns []ProjectedColumn

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
