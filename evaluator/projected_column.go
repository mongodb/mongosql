package evaluator

import (
	"github.com/10gen/sqlproxy/evaluator/results"
)

// ProjectedColumn is a column projection. It contains the SQLExpr for the column
// as well as the column information that will be projected.
type ProjectedColumn struct {
	// Column holds the projection information.
	*results.Column

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

// NewColumnFromSQLColumnExpr returns a new Column struct created
// using the values from the SQLColumnExpr and isPrimaryKey.
func NewColumnFromSQLColumnExpr(sqlColExpr SQLColumnExpr, isPrimaryKey bool) *results.Column {
	return results.NewColumn(
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

func newProjectedColumnFromColumn(c *results.Column) ProjectedColumn {
	clone := c.Clone()
	clone.Name = c.Name
	return ProjectedColumn{
		Column: clone,
		Expr:   newSQLColumnExprFromColumn(c),
	}
}

func newProjectedColumnFromColumnWithName(c *results.Column, name string) ProjectedColumn {
	clone := c.Clone()
	clone.Name = name
	return ProjectedColumn{
		Column: clone,
		Expr:   newSQLColumnExprFromColumn(c),
	}
}

func newProjectedColumnFromColumnWithExpr(c *results.Column, expr SQLExpr) *ProjectedColumn {
	clone := c.Clone()
	clone.EvalType = expr.EvalType()
	return &ProjectedColumn{
		Column: clone,
		Expr:   expr,
	}
}

// columnsToProjectedColumns converts cs to ProjectedColumns
// using the SQLColumnExpr type - constructed from
// the ProjectedColumn as the wrapped expression.
func columnsToProjectedColumns(cs results.Columns) ProjectedColumns {
	var projectedColumns ProjectedColumns
	for _, c := range cs {
		projectedColumn := ProjectedColumn{
			Expr: NewSQLColumnExpr(c.SelectID, c.Database,
				c.Table, c.Name, c.EvalType, c.MongoType, false),
			Column: c.Clone(),
		}
		projectedColumns = append(projectedColumns, projectedColumn)
	}
	return projectedColumns
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

// String is useful for ProjectedColumn because during pushdown we
// use the result of this method for field names in $project stages.
// The special-cased subquery exprs return the ProjectedColumn.Name
// as opposed to ProjectedColumn.Expr.String() for two main reasons.
// Their String() output contains pretty-printed PlanStages which
// (1) are unnecessarily long, and (2) contain newlines and tabs.
func (pc ProjectedColumn) String() string {
	switch pc.Expr.(type) {
	case *SQLSubqueryAllExpr, *SQLSubqueryAnyExpr, *SQLSubqueryCmpExpr, *SQLSubqueryExpr:
		return pc.Name
	default:
		return pc.Expr.String()
	}
}
